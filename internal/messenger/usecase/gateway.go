package usecase

import (
	"context"
	"quaztz/internal/messenger/domain"

	"github.com/google/uuid"
)

type Gateway interface {
	DeliverMessage(ctx context.Context, recipientIDs []uuid.UUID, message domain.Message) error
	InvalidateSession(ctx context.Context, sessionID uuid.UUID) error
	IsRevoked(ctx context.Context, sessionID uuid.UUID) (bool, error)
}
