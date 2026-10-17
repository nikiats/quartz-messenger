package presence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"quartz/internal/common/rediskey"
)

type Presence struct {
	rdb       *redis.Client
	gatewayID uuid.UUID
}

func New(rdb *redis.Client, gatewayID uuid.UUID) *Presence {
	return &Presence{
		rdb:       rdb,
		gatewayID: gatewayID,
	}
}

func (p *Presence) Register(ctx context.Context, userID uuid.UUID) error {
	err := p.rdb.SAdd(ctx, rediskey.Connection(userID), p.gatewayID.String()).Err()
	if err != nil {
		return fmt.Errorf("register presence of user %s: %w", userID, err)
	}

	return nil
}

func (p *Presence) Unregister(ctx context.Context, userID uuid.UUID) error {
	err := p.rdb.SRem(ctx, rediskey.Connection(userID), p.gatewayID.String()).Err()
	if err != nil {
		return fmt.Errorf("unregister presence of user %s: %w", userID, err)
	}

	return nil
}
