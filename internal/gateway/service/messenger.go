package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

var ErrSendFailed = fmt.Errorf("message send failed")

const requestTimeout = 10 * time.Second

type MessageContent struct {
	Text string
}

type Message struct {
	ChatID          uuid.UUID
	SenderID        uuid.UUID
	SessionID       uuid.UUID
	ClientMessageID uuid.UUID
	Content         MessageContent
}

type Sent struct {
	Seqnum int64
	SentAt time.Time
}

// запросы к мессенджерам через NATS
type Messenger struct {
	nc *nats.Conn
}

func NewMessenger(nc *nats.Conn) *Messenger {
	return &Messenger{nc: nc}
}

// TODO: bus.SendMessageRequest -> запрос на bus.SubjectSendMessage -> bus.SendMessageResponse
func (m *Messenger) SendMessage(ctx context.Context, message Message) (Sent, error) {
	return Sent{}, ErrInternal
}
