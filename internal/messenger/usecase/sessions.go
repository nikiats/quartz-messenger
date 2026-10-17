package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"quartz/internal/common/auth"
	"quartz/internal/common/token"
	"quartz/internal/messenger/domain"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
)

type Sessions struct {
	r             SessionRepository
	users         UserRepository
	gateway       Gateway
	authenticator Authenticator
	issuer        *token.Issuer
	verifier      *token.Verifier
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewSessions(
	r SessionRepository,
	users UserRepository,
	gateway Gateway,
	authenticator Authenticator,
	issuer *token.Issuer,
	verifier *token.Verifier,
	accessTTL, refreshTTL time.Duration,
) *Sessions {
	return &Sessions{
		r:             r,
		users:         users,
		gateway:       gateway,
		authenticator: authenticator,
		issuer:        issuer,
		verifier:      verifier,
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

func (s *Sessions) Login(ctx context.Context, login, password string) (domain.Session, domain.Tokens, error) {
	credentials, err := s.users.GetCredentials(ctx, login)

	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return domain.Session{}, domain.Tokens{}, domain.ErrInvalidCredentials
	case err != nil:
		return domain.Session{}, domain.Tokens{}, err
	}

	match, err := argon2id.ComparePasswordAndHash(password, credentials.PasswordHash)

	switch {
	case err != nil:
		return domain.Session{}, domain.Tokens{}, err
	case !match:
		return domain.Session{}, domain.Tokens{}, domain.ErrInvalidCredentials
	case credentials.Suspended:
		return domain.Session{}, domain.Tokens{}, domain.ErrSuspendedUser
	}

	now := time.Now()
	session := domain.Session{
		UserID:    credentials.UserID,
		SessionID: uuid.New(),
		ExpiresAt: now.Add(s.refreshTTL),
	}

	tokens, err := s.createTokens(session, now)
	if err != nil {
		return domain.Session{}, domain.Tokens{}, err
	}

	err = s.r.CreateSession(ctx, session, hashToken(tokens.RefreshToken))
	if err != nil {
		return domain.Session{}, domain.Tokens{}, err
	}

	return session, tokens, nil
}

func (s *Sessions) Refresh(ctx context.Context, refreshToken string) (domain.Tokens, error) {
	session, err := s.parse(refreshToken, token.Refresh)
	if err != nil {
		return domain.Tokens{}, err
	}

	now := time.Now()
	tokens, err := s.createTokens(session, now)
	if err != nil {
		return domain.Tokens{}, err
	}

	session.ExpiresAt = now.Add(s.refreshTTL)
	err = s.r.RefreshSession(ctx, session, hashToken(refreshToken), hashToken(tokens.RefreshToken))
	if err != nil {
		return domain.Tokens{}, err
	}

	return tokens, nil
}

func (s *Sessions) Revoke(ctx context.Context, userID, sessionID uuid.UUID) error {
	err := s.r.RevokeSession(ctx, userID, sessionID)
	if err != nil {
		return err
	}

	return s.gateway.InvalidateSession(ctx, sessionID)
}

func (s *Sessions) Authenticate(ctx context.Context, accessToken string) (domain.Session, error) {
	claims, err := s.authenticator.AuthenticateToken(ctx, accessToken)

	switch {
	case errors.Is(err, token.ErrExpired):
		return domain.Session{}, domain.ErrTokenExpired
	case errors.Is(err, token.ErrInvalid):
		return domain.Session{}, domain.ErrTokenInvalid
	case errors.Is(err, auth.ErrRevoked):
		return domain.Session{}, domain.ErrSessionRevoked
	case err != nil:
		return domain.Session{}, err
	}

	return sessionFromClaims(claims), nil
}

func (s *Sessions) CheckRevoked(ctx context.Context, sessionID uuid.UUID) error {
	revoked, err := s.authenticator.CheckRevoked(ctx, sessionID)
	if err != nil {
		return err
	}
	if revoked {
		return domain.ErrSessionRevoked
	}

	return nil
}

func (s *Sessions) createTokens(session domain.Session, now time.Time) (domain.Tokens, error) {
	access, err := s.sign(session, token.Access, now.Add(s.accessTTL))
	if err != nil {
		return domain.Tokens{}, err
	}

	refresh, err := s.sign(session, token.Refresh, now.Add(s.refreshTTL))
	if err != nil {
		return domain.Tokens{}, err
	}

	return domain.Tokens{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

func (s *Sessions) sign(session domain.Session, tokenType string, expiresAt time.Time) (string, error) {
	return s.issuer.Issue(tokenType, token.Claims{
		UserID:    session.UserID,
		SessionID: session.SessionID,
		ExpiresAt: expiresAt,
	})
}

func (s *Sessions) parse(raw, tokenType string) (domain.Session, error) {
	claims, err := s.verifier.Verify(raw, tokenType)

	switch {
	case errors.Is(err, token.ErrExpired):
		return domain.Session{}, domain.ErrTokenExpired
	case err != nil:
		return domain.Session{}, domain.ErrTokenInvalid
	}

	return sessionFromClaims(claims), nil
}

func sessionFromClaims(claims token.Claims) domain.Session {
	return domain.Session{
		UserID:    claims.UserID,
		SessionID: claims.SessionID,
		ExpiresAt: claims.ExpiresAt,
	}
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
