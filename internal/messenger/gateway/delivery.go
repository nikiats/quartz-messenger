package gateway

import (
	"context"

	"github.com/google/uuid"

	"quartz/internal/api/wsevent"
	"quartz/internal/messenger/domain"
)

func (c *Client) DeliverMessage(ctx context.Context, recipientIDs []uuid.UUID, message domain.Message) error {
	return c.dispatch(ctx, recipientIDs, wsevent.ReceiveMessage{
		Method:    wsevent.MethodReceiveMessage,
		ChatID:    message.ChatID,
		MessageID: message.ID,
		SenderID:  message.SenderID,
		SentAt:    message.SentAt,
		Content: wsevent.MessageContent{
			Text: message.Content.Text,
		},
	})
}
