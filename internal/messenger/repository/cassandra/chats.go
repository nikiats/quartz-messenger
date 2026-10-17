package cassandra

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"quartz/internal/common/snowflake"
	"quartz/internal/messenger/domain"
	"quartz/internal/messenger/repository/cassandra/db"
	"slices"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/scylladb/gocqlx/v3/qb"
	"golang.org/x/sync/errgroup"
)

func (r *Messenger) GetChat(ctx context.Context, chatID int64, userID uuid.UUID) (domain.Chat, error) {
	var row db.ChatMembersStruct
	err := getChat.query(ctx, r.session).BindMap(
		qb.M{
			"chat_id": chatID,
			"user_id": userID,
		},
	).GetRelease(&row)

	if errors.Is(err, gocql.ErrNotFound) {
		return domain.Chat{}, domain.ErrChatNotFound
	}
	if err != nil {
		return domain.Chat{}, fmt.Errorf("get chat %d: %w", chatID, err)
	}

	return domain.Chat{
		ID:      row.ChatId,
		Type:    domain.ChatType(row.Type),
		Title:   row.Title,
		UserA:   row.UserA,
		UserB:   row.UserB,
		AdminID: row.AdminId,
		Deleted: row.Deleted,
		Version: row.Version,
	}, nil
}

// TODO: пагинация, сейчас загружаются все чаты пользователя
func (r *Messenger) GetChats(ctx context.Context, userID uuid.UUID) ([]domain.Chat, error) {
	var chatIDs []int64
	err := getUserChats.query(ctx, r.session).BindMap(
		qb.M{
			"user_id": userID,
		},
	).SelectRelease(&chatIDs)

	if err != nil {
		return nil, fmt.Errorf("get user chats: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(queryConcurrency)

	chats := make([]domain.Chat, len(chatIDs))
	for i, chatID := range chatIDs {
		g.Go(func() error {
			chat, err := r.GetChat(ctx, chatID, userID)
			if errors.Is(err, domain.ErrChatNotFound) {
				return nil
			}
			if err != nil {
				return err
			}
			if chat.Deleted {
				return nil
			}

			chat.Activity, err = r.GetLastMessage(ctx, chatID)
			if err != nil {
				return err
			}

			chats[i] = chat

			return nil
		})
	}
	if err = g.Wait(); err != nil {
		return nil, err
	}

	chats = slices.DeleteFunc(chats, func(chat domain.Chat) bool {
		return chat.ID == 0
	})

	slices.SortFunc(chats, func(a, b domain.Chat) int {
		activityA := a.ID
		if a.Activity.ID != 0 {
			activityA = a.Activity.ID
		}

		activityB := b.ID
		if b.Activity.ID != 0 {
			activityB = b.Activity.ID
		}

		return cmp.Compare(activityB, activityA)
	})
	return chats, nil
}

func (r *Messenger) GetLastMessage(ctx context.Context, chatID int64) (domain.LastMessage, error) {
	var row db.ChatActivityStruct
	err := getChatActivity.query(ctx, r.session).BindMap(
		qb.M{
			"chat_id": chatID,
		},
	).GetRelease(&row)

	if errors.Is(err, gocql.ErrNotFound) {
		return domain.LastMessage{}, nil
	}
	if err != nil {
		return domain.LastMessage{}, fmt.Errorf("get last message %d: %w", chatID, err)
	}

	return domain.LastMessage{
		ID:       row.LastMessageId,
		SenderID: row.LastMessageSenderId,
		Content: domain.MessageContent{
			Text: row.LastMessagePreview,
		},
		SentAt: snowflake.Time(row.LastMessageId),
	}, nil
}

func (r *Messenger) GetChatMembers(ctx context.Context, chatID int64) ([]domain.Member, error) {
	var rows []db.ChatMembersStruct
	err := getChatMembers.query(ctx, r.session).BindMap(
		qb.M{
			"chat_id": chatID,
		},
	).SelectRelease(&rows)

	if err != nil {
		return nil, fmt.Errorf("get chat members %d: %w", chatID, err)
	}

	members := make([]domain.Member, 0, len(rows))
	for _, row := range rows {
		if row.UserId == uuid.Nil {
			continue
		}

		var role domain.Role
		switch {
		case row.UserId == row.AdminId:
			role = domain.RoleAdmin
		default:
			role = domain.RoleRegular
		}

		members = append(members, domain.Member{
			UserID:   row.UserId,
			Role:     role,
			JoinedAt: row.JoinedAt,
		})
	}
	return members, nil
}

func (r *Messenger) CreateChat(ctx context.Context, chat domain.Chat, members []domain.Member) error {
	batch := r.session.ContextBatch(ctx, gocql.UnloggedBatch)

	err := batch.BindMap(
		insertChat.query(ctx, r.session),
		qb.M{
			"chat_id":  chat.ID,
			"type":     string(chat.Type),
			"title":    chat.Title,
			"user_a":   chat.UserA,
			"user_b":   chat.UserB,
			"admin_id": chat.AdminID,
			"deleted":  false,
			"version":  chat.Version,
		},
	)

	if err != nil {
		return err
	}

	query := insertChatMember.query(ctx, r.session)
	for _, member := range members {
		err := batch.BindMap(
			query,
			qb.M{
				"chat_id":   chat.ID,
				"user_id":   member.UserID,
				"joined_at": member.JoinedAt,
			},
		)

		if err != nil {
			return err
		}
	}

	applied, err := r.executeBatch(batch)
	if err != nil {
		return fmt.Errorf("create chat %d: %w", chat.ID, err)
	}
	if !applied {
		return domain.ErrChatExists
	}
	return nil
}

func (r *Messenger) DeleteChat(ctx context.Context, chat domain.Chat) error {
	args := updateChatArgs(chat)
	args["deleted"] = true

	applied, err := updateChat.query(ctx, r.session).BindMap(
		args,
	).ExecCASRelease()

	if err != nil {
		return fmt.Errorf("delete chat %d: %w", chat.ID, err)
	}
	return ensureVersion(applied)
}

func (r *Messenger) CreateChatMember(ctx context.Context, chat domain.Chat, member domain.Member) error {
	applied, err := addChatMember.query(ctx, r.session).BindMap(
		qb.M{
			"chat_id":          chat.ID,
			"user_id":          member.UserID,
			"joined_at":        member.JoinedAt,
			"version":          chat.Version + 1,
			"expected_version": chat.Version,
		},
	).ExecCASRelease()

	if err != nil {
		return fmt.Errorf("create chat member %d: %w", chat.ID, err)
	}
	return ensureVersion(applied)
}

func (r *Messenger) DeleteChatMember(ctx context.Context, chat domain.Chat, userID uuid.UUID) error {
	batch := r.session.ContextBatch(ctx, gocql.UnloggedBatch)

	err := batch.BindMap(
		deleteChatMember.query(ctx, r.session),
		qb.M{
			"chat_id": chat.ID,
			"user_id": userID,
		},
	)

	if err != nil {
		return err
	}

	err = batch.BindMap(
		updateChat.query(ctx, r.session),
		updateChatArgs(chat),
	)

	if err != nil {
		return err
	}

	applied, err := r.executeBatch(batch)
	if err != nil {
		return fmt.Errorf("delete chat member %d: %w", chat.ID, err)
	}
	return ensureVersion(applied)
}

func (r *Messenger) CreatePrivateChat(ctx context.Context, chat domain.Chat) (int64, bool, error) {
	row := db.PrivateChatsStruct{
		UserA:  chat.UserA,
		UserB:  chat.UserB,
		ChatId: chat.ID,
	}

	applied, err := insertPrivateChat.query(ctx, r.session).BindStruct(
		row,
	).GetCASRelease(&row)

	if err != nil {
		return 0, false, fmt.Errorf("create private chat %d: %w", chat.ID, err)
	}
	return row.ChatId, applied, nil
}

func (r *Messenger) DeletePrivateChat(ctx context.Context, chat domain.Chat) error {
	_, err := deletePrivateChat.query(ctx, r.session).BindStruct(
		db.PrivateChatsStruct{
			UserA:  chat.UserA,
			UserB:  chat.UserB,
			ChatId: chat.ID,
		},
	).ExecCASRelease()

	if err != nil {
		return fmt.Errorf("delete private chat %d: %w", chat.ID, err)
	}
	return nil
}

func (r *Messenger) CreateUserChats(ctx context.Context, chatID int64, members []domain.Member) error {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(queryConcurrency)

	for _, member := range members {
		g.Go(func() error {
			return insertUserChat.query(ctx, r.session).BindStruct(
				db.UserChatsStruct{
					UserId:   member.UserID,
					ChatId:   chatID,
					JoinedAt: member.JoinedAt,
				},
			).ExecRelease()
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("create user chats %d: %w", chatID, err)
	}
	return nil
}

func (r *Messenger) DeleteUserChats(ctx context.Context, chatID int64, userIDs []uuid.UUID) error {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(queryConcurrency)

	for _, userID := range userIDs {
		g.Go(func() error {
			return deleteUserChat.query(ctx, r.session).BindMap(
				qb.M{
					"user_id": userID,
					"chat_id": chatID,
				},
			).ExecRelease()
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("delete user chats %d: %w", chatID, err)
	}
	return nil
}

func updateChatArgs(chat domain.Chat) qb.M {
	return qb.M{
		"chat_id":          chat.ID,
		"admin_id":         chat.AdminID,
		"deleted":          chat.Deleted,
		"version":          chat.Version + 1,
		"expected_version": chat.Version,
	}
}
