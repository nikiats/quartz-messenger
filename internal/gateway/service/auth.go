package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"quaztz/internal/platform/rediskey"
	"quaztz/internal/platform/token"
)

var (
	ErrInternal         = fmt.Errorf("internal error")
	ErrSessionIncorrect = fmt.Errorf("incorrect session")
)

type Session struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
	ExpiresAt time.Time
}

// валидация сессий на стороне шлюза без обращения к мессенджеру
type Auth struct {
	rdb      *redis.Client
	verifier *token.Verifier
}

func NewAuth(rdb *redis.Client, verifier *token.Verifier) *Auth {
	return &Auth{rdb: rdb, verifier: verifier}
}

func (a *Auth) ValidateSession(ctx context.Context, accessToken string) (Session, error) {
	claims, err := a.verifier.Verify(accessToken, token.Access)
	if err != nil {
		return Session{}, ErrSessionIncorrect
	}

	revoked, err := a.IsRevoked(ctx, claims.SessionID)
	if err != nil {
		return Session{}, err
	}
	if revoked {
		return Session{}, ErrSessionIncorrect
	}

	return Session{
		UserID:    claims.UserID,
		SessionID: claims.SessionID,
		ExpiresAt: claims.ExpiresAt,
	}, nil
}

func (a *Auth) IsRevoked(ctx context.Context, sessionID uuid.UUID) (bool, error) {
	n, err := a.rdb.Exists(ctx, rediskey.Revoked(sessionID)).Result()

	if err != nil {
		return false, ErrInternal
	}

	return n > 0, nil
}
