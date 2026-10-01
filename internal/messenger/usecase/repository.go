package usecase

import (
	"context"
	"quaztz/internal/messenger/domain"
	"time"

	"github.com/google/uuid"
)

type MessengerRepository interface {
	CreateMessage(ctx context.Context, senderID, chatID, clientMessageID uuid.UUID, content domain.MessageContent) (domain.Message, error)
	GetMessages(ctx context.Context, chatID uuid.UUID, query domain.MessageQuery) ([]domain.Message, error)

	IsChatMember(ctx context.Context, chatID, userID uuid.UUID) (bool, error)

	GetChats(ctx context.Context, userID uuid.UUID) ([]domain.Chat, error)
	GetChat(ctx context.Context, chatID uuid.UUID) (domain.Chat, error)
	GetChatMembers(ctx context.Context, chatID uuid.UUID) ([]domain.Member, error)
	CreateChat(ctx context.Context,
		userID uuid.UUID,
		chatType domain.ChatType,
		title string,
		memberIDs []uuid.UUID,
	) (chat domain.Chat, created bool, err error)
	CreateChatMember(ctx context.Context, userID, chatID uuid.UUID, memberID uuid.UUID) error
	DeleteChat(ctx context.Context, userID, chatID uuid.UUID) error
	DeleteChatMember(ctx context.Context, userID, chatID uuid.UUID, memberID uuid.UUID) error
}

type UserRepository interface {
	CreateUser(ctx context.Context, userID uuid.UUID, login, passwordHash string) error
	GetUser(ctx context.Context, userID uuid.UUID) (domain.User, error)
	GetCredentials(ctx context.Context, login string) (domain.Credentials, error)
	SearchUsers(ctx context.Context, username string, limit int) ([]domain.User, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, visibleName, username *string) error
}

type SessionRepository interface {
	CreateSession(ctx context.Context, session domain.Session, refreshTokenHash string) error
	RefreshSession(
		ctx context.Context,
		oldRefreshTokenHash, newRefreshTokenHash string,
		expiresAt time.Time,
	) (domain.Session, error)
	RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error
	GetRevokedSessions(ctx context.Context, since time.Time) (map[uuid.UUID]time.Time, error)
}
