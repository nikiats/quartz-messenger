package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"quaztz/internal/messenger/domain"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	tokenAccess  = "access"
	tokenRefresh = "refresh"
)

type tokenClaims struct {
	SessionID uuid.UUID `json:"sid"`
	Type      string    `json:"typ"`
	jwt.RegisteredClaims
}

type Sessions struct {
	r          SessionRepository
	users      UserRepository
	gateway    Gateway
	revoked    *revokedSessions
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewSessions(
	r SessionRepository,
	users UserRepository,
	gateway Gateway,
	secret string,
	accessTTL, refreshTTL time.Duration,
) *Sessions {
	return &Sessions{
		r:          r,
		users:      users,
		gateway:    gateway,
		revoked:    newRevokedSessions(accessTTL),
		secret:     []byte(secret),
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
	session, err := s.parse(refreshToken, tokenRefresh)
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
	session, err := s.parse(accessToken, tokenAccess)
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
	access, err := s.sign(session, tokenAccess, now.Add(s.accessTTL))
	if err != nil {
		return domain.Tokens{}, err
	}

	refresh, err := s.sign(session, tokenRefresh, now.Add(s.refreshTTL))
	if err != nil {
		return domain.Tokens{}, err
	}

	return domain.Tokens{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

func (s *Sessions) sign(session domain.Session, tokenType string, expiresAt time.Time) (string, error) {
	claims := tokenClaims{
		SessionID: session.SessionID,
		Type:      tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Subject:   session.UserID.String(),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}

func (s *Sessions) parse(token, tokenType string) (domain.Session, error) {
	var claims tokenClaims

	_, err := jwt.ParseWithClaims(
		token,
		&claims,
		func(*jwt.Token) (any, error) { return s.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)

	switch {
	case errors.Is(err, jwt.ErrTokenExpired):
		return domain.Session{}, domain.ErrTokenExpired
	case err != nil:
		return domain.Session{}, domain.ErrTokenInvalid
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil || claims.Type != tokenType {
		return domain.Session{}, domain.ErrTokenInvalid
	}

	return domain.Session{
		UserID:    userID,
		SessionID: claims.SessionID,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
