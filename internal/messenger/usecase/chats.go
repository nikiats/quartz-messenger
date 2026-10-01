package usecase

import (
	"context"
	"quaztz/internal/messenger/domain"

	"github.com/google/uuid"
)

type Chats struct {
	r MessengerRepository
}

func NewChats(r MessengerRepository) *Chats {
	return &Chats{
		r: r,
	}
}

func (c *Chats) List(ctx context.Context, userID uuid.UUID) ([]domain.Chat, error) {
	return c.r.GetChats(ctx, userID)
}

func (c *Chats) Create(
	ctx context.Context,
	userID uuid.UUID,
	chatType domain.ChatType,
	title string,
	memberIDs []uuid.UUID,
) (domain.Chat, bool, error) {
	return c.r.CreateChat(ctx, userID, chatType, title, memberIDs)
}

func (c *Chats) Get(ctx context.Context, userID, chatID uuid.UUID) (domain.Chat, error) {
	return c.r.GetChat(ctx, userID, chatID)
}

func (c *Chats) Delete(ctx context.Context, userID, chatID uuid.UUID) error {
	return c.r.DeleteChat(ctx, userID, chatID)
}

func (c *Chats) Members(ctx context.Context, userID, chatID uuid.UUID) ([]domain.Member, error) {
	member, err := c.r.IsChatMember(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}
	if !member {
		return nil, domain.ErrChatNotFound
	}

	return c.r.GetChatMembers(ctx, chatID)
}

func (c *Chats) AddMember(ctx context.Context, userID, chatID, memberID uuid.UUID) error {
	return c.r.CreateChatMember(ctx, userID, chatID, memberID)
}

func (c *Chats) RemoveMember(ctx context.Context, userID, chatID, memberID uuid.UUID) error {
	return c.r.DeleteChatMember(ctx, userID, chatID, memberID)
}
