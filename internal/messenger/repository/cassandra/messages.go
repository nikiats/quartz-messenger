package cassandra

import (
	"context"
	"errors"
	"fmt"
	"quaztz/internal/common/snowflake"
	"quaztz/internal/messenger/domain"
	"quaztz/internal/messenger/repository/cassandra/db"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v3/qb"
)

func (r *Messenger) GetMessage(ctx context.Context, chatID, messageID int64) (domain.Message, error) {
	var row db.MessagesStruct
	err := getMessage.query(ctx, r.session).BindMap(
		qb.M{
			"chat_id":    chatID,
			"bucket":     bucketForMessage(messageID),
			"message_id": messageID,
		},
	).GetRelease(&row)

	if errors.Is(err, gocql.ErrNotFound) {
		return domain.Message{}, domain.ErrMessageNotFound
	}
	if err != nil {
		return domain.Message{}, fmt.Errorf("get message %d in chat %d: %w", messageID, chatID, err)
	}

	return rowToMessage(row), nil
}

func (r *Messenger) GetMessages(ctx context.Context, chatID int64, query domain.MessageQuery) ([]domain.Message, error) {
	if query.Limit <= 0 {
		return nil, errors.New("messages query requires positive limit")
	}

	var anchor int64
	var getBuckets, getMessages cqlQuery

	switch {
	case query.Before != nil:
		anchor, getBuckets, getMessages = *query.Before, getBucketsBefore, getMessagesBefore
	case query.After != nil:
		anchor, getBuckets, getMessages = *query.After, getBucketsAfter, getMessagesAfter
	default:
		return nil, errors.New("messages query requires before or after")
	}

	var buckets []int32
	err := getBuckets.query(ctx, r.session).BindMap(
		qb.M{
			"chat_id": chatID,
			"bucket":  bucketForMessage(anchor),
		},
	).SelectRelease(&buckets)

	if err != nil {
		return nil, fmt.Errorf("get message buckets of chat %d: %w", chatID, err)
	}

	messages := make([]domain.Message, 0, query.Limit)
	for _, b := range buckets {
		if len(messages) == query.Limit {
			break
		}

		var rows []db.MessagesStruct
		err := getMessages.query(ctx, r.session).BindMap(
			qb.M{
				"chat_id":    chatID,
				"bucket":     b,
				"message_id": anchor,
				"limit":      query.Limit - len(messages),
			},
		).SelectRelease(&rows)

		if err != nil {
			return nil, fmt.Errorf("get messages of chat %d in bucket %d: %w", chatID, b, err)
		}

		for _, row := range rows {
			messages = append(messages, rowToMessage(row))
		}
	}
	return messages, nil
}

func (r *Messenger) CreateMessage(ctx context.Context, message domain.Message) error {
	err := insertBucket.query(ctx, r.session).BindStruct(
		db.MessageBucketsStruct{
			ChatId: message.ChatID,
			Bucket: bucketForMessage(message.ID),
		},
	).ExecRelease()

	if err != nil {
		return fmt.Errorf("create message bucket of chat %d: %w", message.ChatID, err)
	}

	err = insertMessage.query(ctx, r.session).BindStruct(
		db.MessagesStruct{
			ChatId:    message.ChatID,
			Bucket:    bucketForMessage(message.ID),
			MessageId: message.ID,
			SenderId:  message.SenderID,
			Content:   message.Content.Text,
		},
	).ExecRelease()

	if err != nil {
		return fmt.Errorf("create message %d in chat %d: %w", message.ID, message.ChatID, err)
	}
	return nil
}

func (r *Messenger) UpdateChatActivity(ctx context.Context, message domain.Message) error {
	err := upsertChatActivity.query(ctx, r.session).BindStruct(
		db.ChatActivityStruct{
			ChatId:              message.ChatID,
			LastMessageId:       message.ID,
			LastMessageSenderId: message.SenderID,
			LastMessagePreview:  message.Content.Text,
		},
	).WithTimestamp(snowflake.Time(message.ID).UnixMicro()).ExecRelease()

	if err != nil {
		return fmt.Errorf("update activity of chat %d: %w", message.ChatID, err)
	}
	return nil
}

func rowToMessage(row db.MessagesStruct) domain.Message {
	return domain.Message{
		ID:       row.MessageId,
		ChatID:   row.ChatId,
		SenderID: row.SenderId,
		SentAt:   snowflake.Time(row.MessageId),
		Content:  domain.MessageContent{Text: row.Content},
	}
}
