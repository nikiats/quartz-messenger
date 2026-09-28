package domain

import (
	"time"

	"github.com/google/uuid"
)

type ChatType string

const (
	ChatPrivate ChatType = "private"
	ChatGroup   ChatType = "group"
)

type Chat struct {
	ID            uuid.UUID
	Type          ChatType
	Title         string
	LastSeqnum    int64
	LastMessageAt time.Time
}

type Role string

const (
	RoleRegular Role = "regular"
	RoleAdmin   Role = "admin"
)

type Member struct {
	UserID uuid.UUID
	Role   Role
}
