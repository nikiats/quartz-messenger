package postgres

import (
	"context"
	"quaztz/internal/messenger/domain"
	"quaztz/internal/messenger/repository/postgres/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Messenger struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewMessenger(pool *pgxpool.Pool) *Messenger {
	return &Messenger{
		pool: pool,
		q:    db.New(pool),
	}
}

func (r *Messenger) CreateMessage(ctx context.Context, senderID, chatID, clientMessageID uuid.UUID, content domain.MessageContent) (domain.Message, error) {
	panic("not implemented")
}

func (r *Messenger) GetMessages(ctx context.Context, chatID uuid.UUID, query domain.MessageQuery) ([]domain.Message, error) {
	panic("not implemented")
}

func (r *Messenger) IsChatMember(ctx context.Context, chatID, userID uuid.UUID) (bool, error) {
	panic("not implemented")
}

func (r *Messenger) GetChats(ctx context.Context, userID uuid.UUID) ([]domain.Chat, error) {
	panic("not implemented")
}

func (r *Messenger) GetChat(ctx context.Context, chatID uuid.UUID) (domain.Chat, error) {
	panic("not implemented")
}

func (r *Messenger) GetChatMembers(ctx context.Context, chatID uuid.UUID) ([]domain.Member, error) {
	panic("not implemented")
}

func (r *Messenger) CreateChat(
	ctx context.Context,
	userID uuid.UUID,
	chatType domain.ChatType,
	title string,
	memberIDs []uuid.UUID,
) (chat domain.Chat, created bool, err error) {
	panic("not implemented")
}

func (r *Messenger) CreateChatMember(ctx context.Context, userID, chatID uuid.UUID, memberID uuid.UUID) error {
	panic("not implemented")
}

func (r *Messenger) DeleteChat(ctx context.Context, userID, chatID uuid.UUID) error {
	panic("not implemented")
}

func (r *Messenger) DeleteChatMember(ctx context.Context, userID, chatID uuid.UUID, memberID uuid.UUID) error {
	panic("not implemented")
}
