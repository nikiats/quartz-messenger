package domain

import (
	"time"

	"github.com/google/uuid"
)

type MessageContent struct {
	Text string
}

type Message struct {
	ChatID   uuid.UUID
	SenderID uuid.UUID
	Seqnum   int64
	SentAt   time.Time
	Content  MessageContent
}

type MessageQuery struct {
	Before *int64
	After  *int64
	Limit  int
	Search string
}
