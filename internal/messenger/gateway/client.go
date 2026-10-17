package gateway

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"

	"quartz/internal/api/bus"
	"quartz/internal/common/rediskey"
	"quartz/internal/messenger/domain"
)

type Client struct {
	nc         *nats.Conn
	rdb        *redis.Client
	revokedTTL time.Duration
}

func NewClient(nc *nats.Conn, rdb *redis.Client, revokedTTL time.Duration) *Client {
	return &Client{nc: nc, rdb: rdb, revokedTTL: revokedTTL}
}

func (c *Client) DeliverMessage(ctx context.Context, recipientIDs []uuid.UUID, message domain.Message) error {
	pipe := c.rdb.Pipeline()

	cmds := make([]*redis.StringSliceCmd, len(recipientIDs))
	for i, userID := range recipientIDs {
		cmds[i] = pipe.SMembers(ctx, rediskey.Conn(userID))
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	byGateway := map[string][]uuid.UUID{}
	for i, cmd := range cmds {
		for _, gatewayID := range cmd.Val() {
			byGateway[gatewayID] = append(byGateway[gatewayID], recipientIDs[i])
		}
	}

	for gatewayID, userIDs := range byGateway {
		id, err := uuid.Parse(gatewayID)
		if err != nil {
			continue
		}

		data, err := json.Marshal(bus.MessageEvent{
			ChatID:       message.ChatID,
			MessageID:    message.ID,
			SenderID:     message.SenderID,
			Content:      bus.MessageContent(message.Content),
			SentAt:       message.SentAt,
			RecipientIDs: userIDs,
		})
		if err != nil {
			return err
		}

		if err := c.nc.Publish(bus.SubjectGatewayMessage(id), data); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) InvalidateSession(ctx context.Context, sessionID uuid.UUID) error {
	if err := c.rdb.Set(ctx, rediskey.Revoked(sessionID), "1", c.revokedTTL).Err(); err != nil {
		return err
	}

	data, err := json.Marshal(bus.SessionRevokedEvent{SessionID: sessionID})
	if err != nil {
		return err
	}

	return c.nc.Publish(bus.SubjectSessionRevoked, data)
}

func (c *Client) IsRevoked(ctx context.Context, sessionID uuid.UUID) (bool, error) {
	n, err := c.rdb.Exists(ctx, rediskey.Revoked(sessionID)).Result()
	if err != nil {
		return false, err
	}

	return n > 0, nil
}
