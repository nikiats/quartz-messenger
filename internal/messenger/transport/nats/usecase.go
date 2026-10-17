package nats

import (
	"context"

	"github.com/google/uuid"

	"quartz/internal/messenger/domain"
)

type Sessions interface {
	CheckRevoked(ctx context.Context, sessionID uuid.UUID) error
}

type Messages interface {
	Send(
		ctx context.Context,
		senderID uuid.UUID,
		chatID int64,
		clientMessageID uuid.UUID,
		content domain.MessageContent,
	) (domain.Message, error)
}
