package usecase

import (
	"context"
	"errors"
	"log/slog"
	"quaztz/internal/common/snowflake"
	"quaztz/internal/messenger/domain"

	"github.com/google/uuid"
)

type Messages struct {
	r           MessengerRepository
	gateway     Gateway
	idempotency IdempotencyRepository
	ids         IDGenerator
}

func NewMessages(
	r MessengerRepository,
	gateway Gateway,
	idempotency IdempotencyRepository,
	ids IDGenerator,
) *Messages {
	return &Messages{
		r:           r,
		gateway:     gateway,
		idempotency: idempotency,
		ids:         ids,
	}
}

func (m *Messages) History(
	ctx context.Context,
	userID uuid.UUID,
	chatID int64,
	query domain.MessageQuery,
) ([]domain.Message, bool, error) {
	if _, err := activeChat(ctx, m.r, chatID, userID); err != nil {
		return nil, false, err
	}

	if query.Search != "" {
		return nil, false, nil
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
	senderID uuid.UUID,
	chatID int64,
	clientMessageID uuid.UUID,
	content domain.MessageContent,
) (domain.Message, error) {
	if _, err := activeChat(ctx, m.r, chatID, senderID); err != nil {
		return domain.Message{}, err
	}

	newID := m.ids.Generate()
	messageID, err := m.idempotency.MessageID(ctx, chatID, senderID, clientMessageID, newID)
	if err != nil {
		return domain.Message{}, err
	}

	message := domain.Message{
		ID:       messageID,
		ChatID:   chatID,
		SenderID: senderID,
		SentAt:   snowflake.Time(messageID),
		Content:  content,
	}

	saved := false
	if messageID != newID {
		_, err := m.r.GetMessage(ctx, chatID, messageID)
		switch {
		case err == nil:
			saved = true
		case !errors.Is(err, domain.ErrMessageNotFound):
			return domain.Message{}, err
		}
	}

	if !saved {
		if err := m.r.CreateMessage(ctx, message); err != nil {
			return domain.Message{}, err
		}
	}

	if err := m.r.UpdateChatActivity(ctx, message); err != nil {
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

	recipientIDs := make([]uuid.UUID, 0, len(members))
	for _, member := range members {
		if member.UserID != message.SenderID {
			recipientIDs = append(recipientIDs, member.UserID)
		}
	}

	if err := m.gateway.DeliverMessage(ctx, recipientIDs, message); err != nil {
		slog.Warn("message delivery failed", "chat", message.ChatID, "error", err)
	}
}
