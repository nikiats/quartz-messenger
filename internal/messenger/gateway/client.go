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
)

type Client struct {
	nc         *nats.Conn
	r          *redis.Client
	revokedTTL time.Duration
}

func NewClient(natsConn *nats.Conn, redis *redis.Client, revokedTTL time.Duration) *Client {
	return &Client{
		nc:         natsConn,
		r:          redis,
		revokedTTL: revokedTTL,
	}
}

func (c *Client) dispatch(ctx context.Context, recipientIDs []uuid.UUID, event wsevent.WSEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	pipe := c.r.Pipeline()
	userGateways := make([]*redis.StringSliceCmd, len(recipientIDs))
	for i, userID := range recipientIDs {
		userGateways[i] = pipe.SMembers(ctx, rediskey.Connection(userID))
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	gatewayUsers := make(map[string][]uuid.UUID)
	for i, userID := range recipientIDs {
		for _, gateway := range userGateways[i].Val() {
			gatewayUsers[gateway] = append(gatewayUsers[gateway], userID)
		}
	}

	for gateway, userIDs := range gatewayUsers {
		gatewayID, err := uuid.Parse(gateway)
		if err != nil {
			continue
		}

		subject := bus.SubjectGatewayDeliver(gatewayID)
		msg := nats.NewMsg(subject)
		msg.Data = payload
		for _, userID := range userIDs {
			msg.Header.Add(bus.HeaderRecipient, userID.String())
		}

		if err := c.nc.PublishMsg(msg); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) InvalidateSession(ctx context.Context, sessionID uuid.UUID) error {
	if err := c.r.Set(ctx, rediskey.RevokedSession(sessionID), "1", c.revokedTTL).Err(); err != nil {
		return err
	}

	data, err := json.Marshal(bus.SessionRevokedEvent{SessionID: sessionID})
	if err != nil {
		return err
	}

	return c.nc.Publish(bus.SubjectSessionRevoked, data)
}
