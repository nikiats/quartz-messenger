package bus

import (
	"time"

	"github.com/google/uuid"
)

type MessageContent struct {
	Text string `json:"text"`
}

// SubjectSendMessage
type SendMessageRequest struct {
	ChatID          int64          `json:"chatId,string"`
	SenderID        uuid.UUID      `json:"senderId"`
	SessionID       uuid.UUID      `json:"sessionId"`
	ClientMessageID uuid.UUID      `json:"clientMessageId"`
	Content         MessageContent `json:"content"`
}

type SendMessageResponse struct {
	MessageID int64     `json:"messageId,string"`
	SentAt    time.Time `json:"sentAt"`
}
