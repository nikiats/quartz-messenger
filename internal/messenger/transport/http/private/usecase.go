package private

import (
	"context"

	"github.com/google/uuid"

	"quaztz/internal/messenger/domain"
)

type Authenticator interface {
	Authenticate(ctx context.Context, accessToken string) (domain.Session, error)
}

type Messages interface {
	Send(
		ctx context.Context,
		session domain.Session,
		chatID, clientMessageID uuid.UUID,
		content domain.MessageContent,
	) (domain.Message, error)
}
