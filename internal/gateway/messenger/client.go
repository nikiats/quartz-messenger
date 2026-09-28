package messenger

import (
	"context"
	"fmt"
	"net/http"
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

type Message struct {
	ChatID          uuid.UUID
	SenderID        uuid.UUID
	ClientMessageID uuid.UUID
	Content         string
}

type Sent struct {
	Seqnum int64
	SentAt time.Time
}

type Client struct {
	api *messageprivate.ClientWithResponses
}

func NewClient(baseURL, token string) (*Client, error) {
	authorize := func(_ context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	}

	api, err := messageprivate.NewClientWithResponses(
		baseURL,
		messageprivate.WithHTTPClient(&http.Client{Timeout: requestTimeout}),
		messageprivate.WithRequestEditorFn(authorize),
	)
	if err != nil {
		return nil, fmt.Errorf("messenger client: %w", err)
	}

	return &Client{api: api}, nil
}

func (c *Client) ValidateSession(ctx context.Context, token string) (Session, error) {
	body := messageprivate.TokenValidate{Token: token}

	resp, err := c.api.ValidateSessionWithResponse(ctx, body)
	if err != nil {
		return Session{}, ErrInternal
	}
	if resp.JSON200 == nil {
		return Session{}, ErrInternal
	}

	session := resp.JSON200.Session

	if !resp.JSON200.Success || session == nil {
		return Session{}, ErrSessionIncorrect
	}

	return Session{
		UserID:    session.UserId,
		SessionID: session.SessionId,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

func (c *Client) SendMessage(ctx context.Context, message Message) (Sent, error) {
	body := messageprivate.SendingMessage{
		ChatId:          message.ChatID,
		SenderId:        message.SenderID,
		ClientMessageId: message.ClientMessageID,
		Content:         message.Content,
	}

	resp, err := c.api.SendMessageWithResponse(ctx, body)
	if err != nil {
		return Sent{}, ErrInternal
	}
	if resp.JSON200 == nil {
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
