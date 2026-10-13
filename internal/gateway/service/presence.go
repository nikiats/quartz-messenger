package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"quaztz/internal/common/rediskey"
)

type Presence struct {
	rdb       *redis.Client
	gatewayID uuid.UUID
}

func NewPresence(rdb *redis.Client, gatewayID uuid.UUID) *Presence {
	return &Presence{rdb: rdb, gatewayID: gatewayID}
}

func (p *Presence) Register(ctx context.Context, userID uuid.UUID) error {
	if err := p.rdb.SAdd(ctx, rediskey.Conn(userID), p.gatewayID.String()).Err(); err != nil {
		return ErrInternal
	}

	return nil
}

func (p *Presence) Unregister(ctx context.Context, userID uuid.UUID) error {
	if err := p.rdb.SRem(ctx, rediskey.Conn(userID), p.gatewayID.String()).Err(); err != nil {
		return ErrInternal
	}

	return nil
}
