package snowflake

import (
	"context"
	"errors"
	"log/slog"
	"quartz/internal/common/rediskey"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const nodeTTL = 30 * time.Second

var epochStart = time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)

func init() {
	snowflake.Epoch = epochStart.UnixMilli()
}

var renewScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("PEXPIRE", KEYS[1], ARGV[2])
end
return 0
`)

var releaseScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`)

var ErrNodeLost = errors.New("snowflake node ID lost (probably failed to prolong)")

type Snowflake struct {
	node  *snowflake.Node
	rdb   *redis.Client
	key   string
	owner string
	stop  chan struct{}
	errs  chan error
}

func New(ctx context.Context, rdb *redis.Client) (*Snowflake, error) {
	owner := uuid.NewString()

	for nodeID := range int64(1) << snowflake.NodeBits {
		instanceKey := rediskey.SnowflakeInstance(nodeID)

		taken, err := rdb.SetNX(ctx, instanceKey, owner, nodeTTL).Result()
		if err != nil {
			return nil, err
		}
		if !taken {
			continue
		}

		node, err := snowflake.NewNode(nodeID)
		if err != nil {
			return nil, err
		}

		s := &Snowflake{
			node:  node,
			rdb:   rdb,
			key:   instanceKey,
			owner: owner,
			stop:  make(chan struct{}),
			errs:  make(chan error, 1),
		}
		go s.renew()

		slog.Info("snowflake node acquired", "node", nodeID)
		return s, nil
	}

	return nil, errors.New("could not get nodeID: all node IDs are taken")
}

func (s *Snowflake) Generate() int64 {
	return s.node.Generate().Int64()
}

func (s *Snowflake) Err() <-chan error {
	return s.errs
}

func (s *Snowflake) Close() error {
	close(s.stop)
	return releaseScript.Run(context.Background(), s.rdb, []string{s.key}, s.owner).Err()
}

func (s *Snowflake) renew() {
	interval := nodeTTL / 3
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
		}

		ctx, cancel := context.WithTimeout(context.Background(), interval)
		renewed, err := renewScript.Run(ctx, s.rdb, []string{s.key}, s.owner, nodeTTL.Milliseconds()).Int()
		cancel()

		switch {
		case err == nil && renewed == 1:
		case err == nil:
			s.errs <- ErrNodeLost
			return
		default:
			s.errs <- errors.Join(ErrNodeLost, err)
			slog.Warn("snowflake node renew failed", "error", err)
			return
		}
	}
}

func Time(id int64) time.Time {
	return time.UnixMilli(snowflake.ParseInt64(id).Time())
}
