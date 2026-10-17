package usecase

import (
	"context"
	"quaztz/internal/messenger/domain"

	"github.com/google/uuid"
)

type MessengerRepository interface {
	GetMessage(ctx context.Context, chatID, messageID int64) (domain.Message, error)
	GetMessages(ctx context.Context, chatID int64, query domain.MessageQuery) ([]domain.Message, error)
	CreateMessage(ctx context.Context, message domain.Message) error
	UpdateChatActivity(ctx context.Context, message domain.Message) error

	GetChat(ctx context.Context, chatID int64, userID uuid.UUID) (domain.Chat, error)
	GetChats(ctx context.Context, userID uuid.UUID) ([]domain.Chat, error)
	GetLastMessage(ctx context.Context, chatID int64) (domain.LastMessage, error)
	GetChatMembers(ctx context.Context, chatID int64) ([]domain.Member, error)
	CreateChat(ctx context.Context, chat domain.Chat, members []domain.Member) error
	DeleteChat(ctx context.Context, chat domain.Chat) error
	CreateChatMember(ctx context.Context, chat domain.Chat, member domain.Member) error
	DeleteChatMember(ctx context.Context, chat domain.Chat, userID uuid.UUID) error

	CreatePrivateChat(ctx context.Context, chat domain.Chat) (int64, bool, error)
	DeletePrivateChat(ctx context.Context, chat domain.Chat) error

	CreateUserChats(ctx context.Context, chatID int64, members []domain.Member) error
	DeleteUserChats(ctx context.Context, chatID int64, userIDs []uuid.UUID) error
}

type UserRepository interface {
	CreateUser(ctx context.Context, userID uuid.UUID, login, passwordHash string) error
	GetUser(ctx context.Context, userID uuid.UUID) (domain.User, error)
	GetVisibleNames(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]string, error)
	GetCredentials(ctx context.Context, login string) (domain.Credentials, error)
	SearchUsers(ctx context.Context, username string, limit int) ([]domain.User, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, visibleName, username *string) error
	IsActiveUser(ctx context.Context, userID uuid.UUID) (bool, error)
	CountActiveUsers(ctx context.Context, userIDs []uuid.UUID) (int, error)
}

type SessionRepository interface {
	CreateSession(ctx context.Context, session domain.Session, refreshTokenHash string) error
	RefreshSession(ctx context.Context, session domain.Session, oldHash, newHash string) error
	RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error
}

type IdempotencyRepository interface {
	MessageID(ctx context.Context, chatID int64, senderID, clientMessageID uuid.UUID, messageID int64) (int64, error)
	ChatID(ctx context.Context, userID, clientRequestID uuid.UUID, chatID int64) (int64, error)
}
