package domain

import (
	"time"

	"github.com/google/uuid"
)

type MessageContent struct {
	Text string
}

type Message struct {
	ID       int64
	ChatID   int64
	SenderID uuid.UUID
	SentAt   time.Time
	Content  MessageContent
}

type MessageQuery struct {
	Before *int64
	After  *int64
	Limit  int
	Search string
}
