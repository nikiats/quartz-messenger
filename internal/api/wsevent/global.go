package wsevent

import (
	"time"

	"github.com/google/uuid"
)

// события, приходящие пользователю независимо от его требований

type ReceiveMessage struct {
	ChatID    int64     `json:"chatId,string"`
	MessageID int64     `json:"messageId,string"`
	SenderID  uuid.UUID `json:"senderId"`
	SentAt    time.Time `json:"sentAt"`
	Content   struct {
		Text string `json:"text"`
	} `json:"content"`
}

func NewReceiveMessage(chatID, messageID int64, senderID uuid.UUID, sentAt time.Time, text string) WSEvent {
	return WSEvent{
		Method: MethodReceiveMessage,
		Data: ReceiveMessage{
			ChatID:    chatID,
			MessageID: messageID,
			SenderID:  senderID,
			SentAt:    sentAt,
			Content: struct {
				Text string `json:"text"`
			}{
				Text: text,
			},
		},
	}
}
