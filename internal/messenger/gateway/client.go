package gateway

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"quaztz/internal/api/gatewayprivate"
	"quaztz/internal/messenger/domain"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInternal = errors.New("internal error")
)

const requestTimeout = 10 * time.Second

type Client struct {
	api *gatewayprivate.ClientWithResponses
}

func NewClient(baseURL, token string) (*Client, error) {
	authorize := func(_ context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	}

	api, err := gatewayprivate.NewClientWithResponses(
		baseURL,
		gatewayprivate.WithHTTPClient(&http.Client{Timeout: requestTimeout}),
		gatewayprivate.WithRequestEditorFn(authorize),
	)
	if err != nil {
		return nil, fmt.Errorf("gateway client: %w", err)
	}

	return &Client{api: api}, nil
}

func (c Client) DeliverMessage(ctx context.Context, recipientID uuid.UUID, message domain.Message) error {
	body := gatewayprivate.ReceivedMessage{
		ChatId:      message.ChatID,
		SenderId:    message.SenderID,
		Seqnum:      message.Seqnum,
		RecipientId: recipientID,
		SentAt:      message.SentAt,
		Content: gatewayprivate.MessageContent{
			Text: message.Content.Text,
		},
	}

	resp, err := c.api.DeliverMessageWithResponse(ctx, body)
	if err != nil {
		slog.Warn("gatewayprivate DeliverMessage error", "error", err)
		return ErrInternal
	}

	if resp.StatusCode() != 202 {
		slog.Warn("gatewayprivate DeliverMessage error", "code", resp.StatusCode(), "body", string(resp.GetBody()))
		return ErrInternal
	}

	return nil
}

func (c Client) InvalidateSession(ctx context.Context, sessionID uuid.UUID) error {
	resp, err := c.api.InvalidateSessionWithResponse(ctx, sessionID)
	if err != nil {
		slog.Warn("gatewayprivate InvalidateSession error", "error", err)
		return ErrInternal
	}

	if resp.StatusCode() != 204 {
		slog.Warn("gatewayprivate InvalidateSession error", "code", resp.StatusCode(), "body", string(resp.GetBody()))
		return ErrInternal
	}

	return nil
}
