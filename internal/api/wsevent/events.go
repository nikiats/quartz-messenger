package wsevent

import (
	"time"

	"github.com/google/uuid"
)

const MethodReceiveMessage = "receiveMessage"

type MessageContent struct {
	Text string `json:"text"`
}

type ReceiveMessage struct {
	Method    string         `json:"method"`
	ChatID    int64          `json:"chatId,string"`
	MessageID int64          `json:"messageId,string"`
	SenderID  uuid.UUID      `json:"senderId"`
	SentAt    time.Time      `json:"sentAt"`
	Content   MessageContent `json:"content"`
}
