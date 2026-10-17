package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"quartz/internal/common/rediskey"
	"quartz/internal/common/token"
)

var ErrRevoked = errors.New("session revoked")

type Authentication struct {
	rdb      *redis.Client
	verifier *token.Verifier
}

func NewAuthentication(rdb *redis.Client, verifier *token.Verifier) *Authentication {
	return &Authentication{
		rdb:      rdb,
		verifier: verifier,
	}
}

func (a *Authentication) AuthenticateToken(ctx context.Context, accessToken string) (token.Claims, error) {
	claims, err := a.verifier.Verify(accessToken, token.Access)
	if err != nil {
		return token.Claims{}, err
	}

	revoked, err := a.CheckRevoked(ctx, claims.SessionID)
	if err != nil {
		return token.Claims{}, err
	}
	if revoked {
		return token.Claims{}, ErrRevoked
	}

	return claims, nil
}

func (a *Authentication) CheckRevoked(ctx context.Context, sessionID uuid.UUID) (bool, error) {
	n, err := a.rdb.Exists(ctx, rediskey.RevokedSession(sessionID)).Result()
	if err != nil {
		return false, fmt.Errorf("check session %s revoked: %w", sessionID, err)
	}

	return n > 0, nil
}
