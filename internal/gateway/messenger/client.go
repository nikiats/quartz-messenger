package messenger

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"quaztz/internal/api/messageprivate"
)

var (
	ErrInternal         = fmt.Errorf("internal error")
	ErrSessionIncorrect = fmt.Errorf("incorrect session")
	ErrSendFailed       = fmt.Errorf("message send failed")
)

const requestTimeout = 10 * time.Second

type Session struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
	ExpiresAt time.Time
}

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

type Client struct {
}

func NewClient() (*Client, error) {
	return &Client{}, nil
}

func (c *Client) ValidateSession(ctx context.Context, token string) (Session, error) {
	return Session{}, nil
}

func (c *Client) SendMessage(ctx context.Context, message Message) (Sent, error) {
	body := messageprivate.SendingMessage{
		ChatId:          message.ChatID,
		SenderId:        message.SenderID,
		ClientMessageId: message.ClientMessageID,
		Content: messageprivate.MessageContent{
			Text: message.Content.Text,
		},
		SessionId: message.SessionID,
	}

	resp, err := c.api.SendMessageWithResponse(ctx, body)
	if err != nil {
		slog.Warn("messengerprivate SendMessage error", "error", err)
		return Sent{}, ErrInternal
	}
	if resp.JSON200 == nil {
		slog.Warn("messengerprivate SendMessage error", "code", resp.StatusCode(), "body", string(resp.GetBody()))
		return Sent{}, ErrInternal
	}

	info := resp.JSON200.Message
	if !resp.JSON200.Success || info == nil {
		return Sent{}, ErrSendFailed
	}

	return Sent{
		Seqnum: info.Seqnum,
		SentAt: info.SentAt,
	}, nil
}
