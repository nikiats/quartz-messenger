package bus

import (
	"time"

	"github.com/google/uuid"
)

// SubjectUserMessage
type MessageEvent struct {
	ChatID    int64          `json:"chatId"`
	MessageID int64          `json:"messageId"`
	SenderID  uuid.UUID      `json:"senderId"`
	Content   MessageContent `json:"content"`
	SentAt    time.Time      `json:"sentAt"`
}

// SubjectSessionRevoked
type SessionRevokedEvent struct {
	SessionID uuid.UUID `json:"sessionId"`
}
