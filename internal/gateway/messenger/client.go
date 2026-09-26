package messenger

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"quaztz/internal/api/messageprivate"
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
		return Session{}, fmt.Errorf("validate session: %w", err)
	}

	if resp.JSON200 == nil {
		return Session{}, unexpected("validate session", resp.StatusCode(), resp.Body)
	}

	return Session{
		UserID:    resp.JSON200.UserId,
		SessionID: resp.JSON200.SessionId,
		ExpiresAt: resp.JSON200.ExpiresAt,
	}, nil
}

func (c *Client) SendMessage(ctx context.Context, m Message) (Sent, error) {
	body := messageprivate.SendingMessage{
		ChatId:          m.ChatID,
		SenderId:        m.SenderID,
		ClientMessageId: m.ClientMessageID,
		Content:         m.Content,
	}

	resp, err := c.api.SendMessageWithResponse(ctx, body)
	if err != nil {
		return Sent{}, fmt.Errorf("send message: %w", err)
	}

	if resp.JSON201 == nil {
		return Sent{}, unexpected("send message", resp.StatusCode(), resp.Body)
	}

	return Sent{
		Seqnum: resp.JSON201.Seqnum,
		SentAt: resp.JSON201.SentAt,
	}, nil
}

func unexpected(op string, status int, body []byte) error {
	return fmt.Errorf("%s: unexpected status %d: %s", op, status, body)
}
