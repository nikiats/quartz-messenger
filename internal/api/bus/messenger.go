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
	ChatID          uuid.UUID      `json:"chatId"`
	SenderID        uuid.UUID      `json:"senderId"`
	SessionID       uuid.UUID      `json:"sessionId"`
	ClientMessageID uuid.UUID      `json:"clientMessageId"`
	Content         MessageContent `json:"content"`
}

type SendMessageResponse struct {
	Seqnum int64     `json:"seqnum"`
	SentAt time.Time `json:"sentAt"`
}
