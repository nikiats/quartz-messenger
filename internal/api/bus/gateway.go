package bus

import (
	"time"

	"github.com/google/uuid"
)

type MessageEvent struct {
	ChatID       int64          `json:"chatId,string"`
	MessageID    int64          `json:"messageId,string"`
	SenderID     uuid.UUID      `json:"senderId"`
	Content      MessageContent `json:"content"`
	SentAt       time.Time      `json:"sentAt"`
	RecipientIDs []uuid.UUID    `json:"recipientIds"`
}

type SessionRevokedEvent struct {
	SessionID uuid.UUID `json:"sessionId"`
}
