package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

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

// TODO: verifier.Verify(accessToken, token.Access), затем проверка token.RevokedKey(sessionId) в Redis (операция 21)
func (a *Auth) ValidateSession(ctx context.Context, accessToken string) (Session, error) {

	return Session{}, ErrInternal
}
