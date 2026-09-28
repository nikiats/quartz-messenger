package domain

import (
	"time"

	"github.com/google/uuid"
)

type MessageContent struct {
	Content string
}

type Message struct {
	ChatID   uuid.UUID
	SenderID uuid.UUID
	Seqnum   int64
	SentAt   time.Time
	Message  MessageContent
}
