package gateway

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"

	"quartz/internal/api/bus"
	"quartz/internal/api/wsevent"
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
	payload, err := json.Marshal(wsevent.ReceiveMessage{
		Method:    wsevent.MethodReceiveMessage,
		ChatID:    message.ChatID,
		MessageID: message.ID,
		SenderID:  message.SenderID,
		SentAt:    message.SentAt,
		Content:   wsevent.MessageContent{Text: message.Content.Text},
	})
	if err != nil {
		return err
	}

	pipe := c.rdb.Pipeline()
	gateways := make([]*redis.StringSliceCmd, len(recipientIDs))
	for i, userID := range recipientIDs {
		gateways[i] = pipe.SMembers(ctx, rediskey.Connection(userID))
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	for i, userID := range recipientIDs {
		for _, gateway := range gateways[i].Val() {
			gatewayID, err := uuid.Parse(gateway)
			if err != nil {
				continue
			}

			if err := c.nc.Publish(bus.SubjectGatewayUser(gatewayID, userID), payload); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Client) InvalidateSession(ctx context.Context, sessionID uuid.UUID) error {
	if err := c.rdb.Set(ctx, rediskey.RevokedSession(sessionID), "1", c.revokedTTL).Err(); err != nil {
		return err
	}

	data, err := json.Marshal(bus.SessionRevokedEvent{SessionID: sessionID})
	if err != nil {
		return err
	}

	return c.nc.Publish(bus.SubjectSessionRevoked, data)
}
