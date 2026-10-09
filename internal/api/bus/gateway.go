package bus

import (
	"time"

	"github.com/google/uuid"
)

// SubjectGatewayMessage
type MessageEvent struct {
	ChatID       uuid.UUID      `json:"chatId"`
	SenderID     uuid.UUID      `json:"senderId"`
	Seqnum       int64          `json:"seqnum"`
	Content      MessageContent `json:"content"`
	SentAt       time.Time      `json:"sentAt"`
	RecipientIDs []uuid.UUID    `json:"recipientIds"`
}

// SubjectSessionRevoked
type SessionRevokedEvent struct {
	SessionID uuid.UUID `json:"sessionId"`
}
