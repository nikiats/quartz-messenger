package usecase

import (
	"context"
	"quartz/internal/common/token"

	"github.com/google/uuid"
)

type Authenticator interface {
	AuthenticateToken(ctx context.Context, accessToken string) (token.Claims, error)
	CheckRevoked(ctx context.Context, sessionID uuid.UUID) (bool, error)
}
