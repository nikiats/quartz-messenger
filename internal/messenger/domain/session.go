package domain

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
	ExpiresAt time.Time
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
}
