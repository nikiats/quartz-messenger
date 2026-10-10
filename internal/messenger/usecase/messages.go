package usecase

import (
	"context"
	"errors"
	"log/slog"
	"quaztz/internal/messenger/domain"
	"quaztz/internal/platform/snowflake"

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
	panic("TODO: операции 12, 13")
}

func (m *Messages) Send(
	ctx context.Context,
	senderID uuid.UUID,
	chatID int64,
	clientMessageID uuid.UUID,
	content domain.MessageContent,
) (domain.Message, error) {
	// проверка наличия и неудаленности чата
	chat, err := m.r.GetChat(ctx, chatID, senderID)
	if err != nil {
		return domain.Message{}, err
	}
	if chat.Deleted {
		return domain.Message{}, domain.ErrChatNotFound
	}

	// создаем новый ID, но меняем его на старый, если это повторная отправка
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

	// при повторе сообщение могло сохраниться в прошлой попытке
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

	// бакет и последнее сообщение чата
	if err := m.r.UpdateChatActivity(ctx, message); err != nil {
		return domain.Message{}, err
	}

	// TODO: индексация в OpenSearch

	go m.deliver(context.WithoutCancel(ctx), message)

	return message, nil
}

// рассылка сообщения всем участникам чата кроме отправителя
func (m *Messages) deliver(ctx context.Context, message domain.Message) {
	members, err := m.r.GetChatMembers(ctx, message.ChatID)
	if err != nil {
		slog.Warn("message delivery failed", "chat", message.ChatID, "error", err)
		return
	}

	// список получателей
	recipientIDs := make([]uuid.UUID, 0, len(members))
	for _, member := range members {
		if member.UserID != message.SenderID {
			recipientIDs = append(recipientIDs, member.UserID)
		}
	}

	// доставка сообщения через шлюз
	if err := m.gateway.DeliverMessage(ctx, recipientIDs, message); err != nil {
		slog.Warn("message delivery failed", "chat", message.ChatID, "error", err)
	}
}
