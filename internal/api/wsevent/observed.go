package wsevent

import "github.com/google/uuid"

// события по наблюдаемым пользователем компонентам

type UserOnlineStatusChange struct {
	UserID   uuid.UUID `json:"userId"`
	IsOnline bool      `json:"isOnline"`
}
