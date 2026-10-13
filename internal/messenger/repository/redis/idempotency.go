package redis

import (
	"context"
	"errors"
	"quaztz/internal/common/rediskey"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Idempotency struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewIdempotency(rdb *redis.Client, ttl time.Duration) *Idempotency {
	return &Idempotency{
		rdb: rdb,
		ttl: ttl,
	}
}

func (r *Idempotency) MessageID(ctx context.Context, chatID int64, senderID, clientMessageID uuid.UUID, messageID int64) (int64, error) {
	return r.reserve(ctx, rediskey.Message(chatID, senderID, clientMessageID), messageID)
}

func (r *Idempotency) ChatID(ctx context.Context, userID, clientRequestID uuid.UUID, chatID int64) (int64, error) {
	return r.reserve(ctx, rediskey.Chat(userID, clientRequestID), chatID)
}

func (r *Idempotency) reserve(ctx context.Context, key string, id int64) (int64, error) {
	old, err := r.rdb.SetArgs(ctx, key, id, redis.SetArgs{
		Mode: "NX",
		Get:  true,
		TTL:  r.ttl,
	}).Result()
	if errors.Is(err, redis.Nil) {
		return id, nil
	}
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(old, 10, 64)
}
