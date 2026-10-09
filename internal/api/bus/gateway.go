package bus

import (
	"time"

	"github.com/google/uuid"
)

// SubjectUserMessage
type MessageEvent struct {
	ChatID       uuid.UUID      `json:"chatId"`
	SenderID     uuid.UUID      `json:"senderId"`
	Seqnum       int64          `json:"seqnum"`
	Content      MessageContent `json:"content"`
	SentAt       time.Time      `json:"sentAt"`
	RecepientIDs []uuid.UUID    `json:"recepientIds"`
}

// SubjectSessionRevoked
type SessionRevokedEvent struct {
	SessionID uuid.UUID `json:"sessionId"`
}
