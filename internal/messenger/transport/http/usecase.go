package http

import (
	"context"

	"github.com/google/uuid"

	"quartz/internal/messenger/domain"
)

type Authenticator interface {
	Authenticate(ctx context.Context, accessToken string) (domain.Session, error)
}

type Sessions interface {
	Login(ctx context.Context, login, password string) (domain.Session, domain.Tokens, error)
	Refresh(ctx context.Context, refreshToken string) (domain.Tokens, error)
	Revoke(ctx context.Context, userID, sessionID uuid.UUID) error
}

type Users interface {
	Register(ctx context.Context, login, password string) error
	Get(ctx context.Context, userID uuid.UUID) (domain.User, error)
	Search(ctx context.Context, query string) ([]domain.User, error)
	EditProfile(ctx context.Context, userID uuid.UUID, visibleName, username *string) error
}

type Chats interface {
	List(ctx context.Context, userID uuid.UUID) ([]domain.Chat, error)
	Create(
		ctx context.Context,
		userID uuid.UUID,
		chatType domain.ChatType,
		title string,
		memberIDs []uuid.UUID,
		clientRequestID uuid.UUID,
	) (chat domain.Chat, created bool, err error)
	Get(ctx context.Context, userID uuid.UUID, chatID int64) (domain.Chat, error)
	Delete(ctx context.Context, userID uuid.UUID, chatID int64) error
	Members(ctx context.Context, userID uuid.UUID, chatID int64) ([]domain.Member, error)
	AddMember(ctx context.Context, userID uuid.UUID, chatID int64, memberID uuid.UUID) error
	RemoveMember(ctx context.Context, userID uuid.UUID, chatID int64, memberID uuid.UUID) error
}

type Messages interface {
	History(
		ctx context.Context,
		userID uuid.UUID,
		chatID int64,
		query domain.MessageQuery,
	) (messages []domain.Message, hasMore bool, err error)
	Send(
		ctx context.Context,
		senderID uuid.UUID,
		chatID int64,
		clientMessageID uuid.UUID,
		content domain.MessageContent,
	) (domain.Message, error)
}
