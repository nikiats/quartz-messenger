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
	ID       int64
	Type     ChatType
	Title    string
	UserA    uuid.UUID
	UserB    uuid.UUID
	AdminID  uuid.UUID
	Deleted  bool
	Version  int32
	Activity LastMessage
}

// разница состоит в том, что контент может быть урезан до превью
type LastMessage Message

type Role string

const (
	RoleRegular Role = "regular"
	RoleAdmin   Role = "admin"
)

type Member struct {
	UserID   uuid.UUID
	Role     Role
	JoinedAt time.Time
}
