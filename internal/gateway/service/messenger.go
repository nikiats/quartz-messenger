package service

import (
	"context"
	"encoding/json"
	"fmt"
	"quaztz/internal/api/bus"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
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
// gateway -> любой инстанс сервиса мессенджера
type Messenger struct {
	nc *nats.Conn
}

func NewMessenger(nc *nats.Conn) *Messenger {
	return &Messenger{nc: nc}
}

func (m *Messenger) SendMessage(ctx context.Context, message Message) (Sent, error) {
	// формирование запроса
	request := bus.SendMessageRequest{
		ChatID:          message.ChatID,
		SenderID:        message.SenderID,
		Content:         bus.MessageContent(message.Content),
		SessionID:       message.SessionID,
		ClientMessageID: message.ClientMessageID,
	}
	data, err := json.Marshal(request)
	if err != nil {
		return Sent{}, ErrInternal
	}

	// отправка в NATS
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	msg, err := m.nc.RequestWithContext(ctx, bus.SubjectSendMessage, data)
	if err != nil {
		return Sent{}, ErrInternal
	}

	// обработка ошибок
	switch msg.Header.Get(micro.ErrorCodeHeader) {
	case "":
	case bus.CodeForbidden:
		return Sent{}, ErrSendFailed
	default:
		return Sent{}, ErrInternal
	}

	// разбор ответа
	var resp bus.SendMessageResponse
	if err = json.Unmarshal(msg.Data, &resp); err != nil {
		return Sent{}, ErrInternal
	}
	return Sent{
		Seqnum: resp.Seqnum,
		SentAt: resp.SentAt,
	}, ErrInternal
}
