package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"quaztz/internal/messenger/domain"
	"quaztz/internal/platform/token"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
)

type Sessions struct {
	r          SessionRepository
	users      UserRepository
	gateway    Gateway
	revoked    *revokedSessions
	issuer     *token.Issuer
	verifier   *token.Verifier
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewSessions(
	r SessionRepository,
	users UserRepository,
	gateway Gateway,
	issuer *token.Issuer,
	verifier *token.Verifier,
	accessTTL, refreshTTL time.Duration,
) *Sessions {
	return &Sessions{
		r:          r,
		users:      users,
		gateway:    gateway,
		revoked:    newRevokedSessions(accessTTL),
		issuer:     issuer,
		verifier:   verifier,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (s *Sessions) LoadRevoked(ctx context.Context) error {
	revoked, err := s.r.GetRevokedSessions(ctx, time.Now().Add(-s.accessTTL))
	if err != nil {
		return err
	}

	for sessionID, revokedAt := range revoked {
		s.revoked.add(sessionID, revokedAt)
	}

	return nil
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

	_, err = s.r.RefreshSession(
		ctx,
		hashToken(refreshToken),
		hashToken(tokens.RefreshToken),
		now.Add(s.refreshTTL),
	)
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

	s.revoked.add(sessionID, time.Now())
	_ = s.gateway.InvalidateSession(ctx, sessionID)

	return nil
}

func (s *Sessions) Authenticate(ctx context.Context, accessToken string) (domain.Session, error) {
	session, err := s.parse(accessToken, token.Access)
	if err != nil {
		return domain.Session{}, err
	}

	err = s.CheckRevoked(ctx, session.SessionID)
	if err != nil {
		return domain.Session{}, err
	}

	return session, nil
}

func (s *Sessions) CheckRevoked(_ context.Context, sessionID uuid.UUID) error {
	if s.revoked.contains(sessionID) {
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

	return domain.Session{
		UserID:    claims.UserID,
		SessionID: claims.SessionID,
		ExpiresAt: claims.ExpiresAt,
	}, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
