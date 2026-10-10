package cassandra

import (
	"context"
	"errors"
	"quaztz/internal/messenger/domain"
	"quaztz/internal/messenger/repository/cassandra/db"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v3/qb"
)

func (r *Messenger) GetMessage(ctx context.Context, chatID, messageID int64) (domain.Message, error) {
	var row db.MessagesStruct
	err := getMessage.QueryContext(ctx, r.session).BindMap(qb.M{
		"chat_id":    chatID,
		"bucket":     bucket(messageID),
		"message_id": messageID,
	}).GetRelease(&row)
	if errors.Is(err, gocql.ErrNotFound) {
		return domain.Message{}, domain.ErrMessageNotFound
	}
	if err != nil {
		return domain.Message{}, err
	}

	return toMessage(row), nil
}

func (r *Messenger) GetMessages(ctx context.Context, chatID int64, query domain.MessageQuery) ([]domain.Message, error) {
	var anchor int64
	var getBuckets, getMessages *qb.SelectBuilder

	switch {
	case query.Before != nil:
		anchor, getBuckets, getMessages = *query.Before, getBucketsBefore, getMessagesBefore
	case query.After != nil:
		anchor, getBuckets, getMessages = *query.After, getBucketsAfter, getMessagesAfter
	default:
		return nil, errors.New("messages query requires before or after")
	}

	var buckets []int32
	err := getBuckets.QueryContext(ctx, r.session).BindMap(qb.M{
		"chat_id": chatID,
		"bucket":  bucket(anchor),
	}).SelectRelease(&buckets)
	if err != nil {
		return nil, err
	}

	messages := make([]domain.Message, 0, query.Limit)
	for _, b := range buckets {
		if len(messages) == query.Limit {
			break
		}

		var rows []db.MessagesStruct
		err := getMessages.QueryContext(ctx, r.session).BindMap(qb.M{
			"chat_id":    chatID,
			"bucket":     b,
			"message_id": anchor,
			"limit":      query.Limit - len(messages),
		}).SelectRelease(&rows)
		if err != nil {
			return nil, err
		}

		for _, row := range rows {
			messages = append(messages, toMessage(row))
		}
	}
	return messages, nil
}

func (r *Messenger) CreateMessage(ctx context.Context, message domain.Message) error {
	return insertMessage.QueryContext(ctx, r.session).BindStruct(db.MessagesStruct{
		ChatId:    message.ChatID,
		Bucket:    bucket(message.ID),
		MessageId: message.ID,
		SenderId:  message.SenderID,
		Content:   message.Content.Text,
	}).ExecRelease()
}

func (r *Messenger) UpdateChatActivity(ctx context.Context, message domain.Message) error {
	err := insertBucket.QueryContext(ctx, r.session).BindStruct(db.MessageBucketsStruct{
		ChatId: message.ChatID,
		Bucket: bucket(message.ID),
	}).ExecRelease()
	if err != nil {
		return err
	}

	// метка времени записи - время сообщения: более старое не перезапишет новое
	return upsertChatActivity.QueryContext(ctx, r.session).BindStruct(db.ChatActivityStruct{
		ChatId:              message.ChatID,
		LastMessageId:       message.ID,
		LastMessageSenderId: message.SenderID,
		LastMessagePreview:  preview(message.Content.Text),
	}).WithTimestamp(idTime(message.ID).UnixMicro()).ExecRelease()
}

func toMessage(row db.MessagesStruct) domain.Message {
	return domain.Message{
		ID:       row.MessageId,
		ChatID:   row.ChatId,
		SenderID: row.SenderId,
		SentAt:   idTime(row.MessageId),
		Content:  domain.MessageContent{Text: row.Content},
	}
}

func preview(text string) string {
	runes := []rune(text)
	if len(runes) <= previewLength {
		return text
	}
	return string(runes[:previewLength])
}
