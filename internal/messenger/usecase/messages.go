package usecase

import (
	"context"
	"quaztz/internal/messenger/domain"

	"github.com/google/uuid"
)

type Messages struct {
	r MessengerRepository
}

func NewMessages(r MessengerRepository) *Messages {
	return &Messages{
		r: r,
	}
}

func (m *Messages) Send(
	ctx context.Context,
	senderID, chatID, clientMessageID uuid.UUID,
	content domain.MessageContent,
) (domain.Message, error) {
	return m.r.CreateMessage(ctx, senderID, chatID, clientMessageID, content)
}
