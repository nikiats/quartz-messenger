package usecase

import (
	"context"
	"log/slog"
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
	userID, chatID uuid.UUID,
	query domain.MessageQuery,
) ([]domain.Message, bool, error) {
	member, err := m.r.IsChatMember(ctx, chatID, userID)
	if err != nil {
		return nil, false, err
	}
	if !member {
		return nil, false, domain.ErrChatNotFound
	}

	limit := query.Limit
	query.Limit = limit + 1

	messages, err := m.r.GetMessages(ctx, chatID, query)
	if err != nil {
		return nil, false, err
	}

	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}

	return messages, hasMore, nil
}

func (m *Messages) Send(
	ctx context.Context,
	senderID, chatID, clientMessageID uuid.UUID,
	content domain.MessageContent,
) (domain.Message, error) {
	message, err := m.r.CreateMessage(ctx, senderID, chatID, clientMessageID, content)
	if err != nil {
		return domain.Message{}, err
	}

	go m.deliver(context.WithoutCancel(ctx), message)

	return message, nil
}

func (m *Messages) deliver(ctx context.Context, message domain.Message) {
	members, err := m.r.GetChatMembers(ctx, message.ChatID)
	if err != nil {
		slog.Warn("message delivery failed", "chat", message.ChatID, "error", err)
		return
	}

	for _, member := range members {
		if member.UserID == message.SenderID {
			continue
		}
		_ = m.gateway.DeliverMessage(ctx, member.UserID, message)
	}
}
