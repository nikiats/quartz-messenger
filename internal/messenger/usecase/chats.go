package usecase

import (
	"context"
	"quaztz/internal/messenger/domain"

	"github.com/google/uuid"
)

type Chats struct {
	r           MessengerRepository
	users       UserRepository
	idempotency IdempotencyRepository
	ids         IDGenerator
}

func NewChats(
	r MessengerRepository,
	users UserRepository,
	idempotency IdempotencyRepository,
	ids IDGenerator,
) *Chats {
	return &Chats{
		r:           r,
		users:       users,
		idempotency: idempotency,
		ids:         ids,
	}
}

func (c *Chats) List(ctx context.Context, userID uuid.UUID) ([]domain.Chat, error) {
	panic("TODO: операция 9")
}

func (c *Chats) Create(
	ctx context.Context,
	userID uuid.UUID,
	chatType domain.ChatType,
	title string,
	memberIDs []uuid.UUID,
	clientRequestID uuid.UUID,
) (domain.Chat, bool, error) {
	panic("TODO: операции 3, 4")
}

func (c *Chats) Get(ctx context.Context, userID uuid.UUID, chatID int64) (domain.Chat, error) {
	panic("TODO: операция 10")
}

func (c *Chats) Delete(ctx context.Context, userID uuid.UUID, chatID int64) error {
	panic("TODO: операция 8")
}

func (c *Chats) Members(ctx context.Context, userID uuid.UUID, chatID int64) ([]domain.Member, error) {
	panic("TODO: операция 11")
}

func (c *Chats) AddMember(ctx context.Context, userID uuid.UUID, chatID int64, memberID uuid.UUID) error {
	panic("TODO: операция 5")
}

func (c *Chats) RemoveMember(ctx context.Context, userID uuid.UUID, chatID int64, memberID uuid.UUID) error {
	panic("TODO: операции 6, 7")
}

// чат, в котором пользователь участник и который не удален
func activeChat(ctx context.Context, r MessengerRepository, chatID int64, userID uuid.UUID) (domain.Chat, error) {
	chat, err := r.GetChat(ctx, chatID, userID)
	if err != nil {
		return domain.Chat{}, err
	}
	if chat.Deleted {
		return domain.Chat{}, domain.ErrChatNotFound
	}
	return chat, nil
}
