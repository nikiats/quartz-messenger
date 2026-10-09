package gateway

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"

	"quaztz/internal/messenger/domain"
)

var (
	ErrInternal = errors.New("internal error")
)

// события для шлюзов через NATS
type Client struct {
	nc *nats.Conn
}

func NewClient(nc *nats.Conn) *Client {
	return &Client{nc: nc}
}

// TODO: bus.MessageEvent -> публикация на bus.SubjectUserMessage(recipientID)
func (c Client) DeliverMessage(ctx context.Context, recipientID uuid.UUID, message domain.Message) error {
	return ErrInternal
}

// TODO: bus.SessionRevokedEvent -> публикация на bus.SubjectSessionRevoked
func (c Client) InvalidateSession(ctx context.Context, sessionID uuid.UUID) error {
	return ErrInternal
}
