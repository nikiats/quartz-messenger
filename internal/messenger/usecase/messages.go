package usecase

import (
	"context"
	"quaztz/internal/messenger/domain"

	"github.com/google/uuid"
)

type Messages struct {
	r       MessengerRepository
	gateway Gateway
}

func NewMessages(r MessengerRepository, gateway Gateway) *Messages {
	return &Messages{
		r:       r,
		gateway: gateway,
	}
}

func (m *Messages) History(
	ctx context.Context,
	userID uuid.UUID,
	chatID int64,
	query domain.MessageQuery,
) ([]domain.Message, bool, error) {
	panic("TODO: операции 12, 13")
}

func (m *Messages) Send(
	ctx context.Context,
	senderID uuid.UUID,
	chatID int64,
	clientMessageID uuid.UUID,
	content domain.MessageContent,
) (domain.Message, error) {
	panic("TODO: операция 2")
}
