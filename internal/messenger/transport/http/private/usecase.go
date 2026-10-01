package private

import (
	"context"

	"github.com/google/uuid"

	"quaztz/internal/messenger/domain"
)

type Authenticator interface {
	Authenticate(ctx context.Context, accessToken string) (domain.Session, error)
}

type Sessions interface {
	CheckRevoked(ctx context.Context, sessionID uuid.UUID) error
}

type Messages interface {
	Send(
		ctx context.Context,
		senderID uuid.UUID,
		chatID uuid.UUID,
		clientMessageID uuid.UUID,
		content domain.MessageContent,
	) (domain.Message, error)
}
