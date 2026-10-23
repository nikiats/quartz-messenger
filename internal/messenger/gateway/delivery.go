package gateway

import (
	"context"

	"github.com/google/uuid"

	"quartz/internal/api/wsevent"
	"quartz/internal/messenger/domain"
)

func (c *Client) DeliverMessage(ctx context.Context, recipientIDs []uuid.UUID, message domain.Message) error {
	return c.dispatch(
		ctx,
		recipientIDs,
		wsevent.NewReceiveMessage(
			message.ChatID,
			message.ID,
			message.SenderID,
			message.SentAt,
			message.Content.Text,
		),
	)
}
