package cassandra

import (
	"cmp"
	"context"
	"errors"
	"quaztz/internal/messenger/domain"
	"quaztz/internal/messenger/repository/cassandra/db"
	"slices"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/scylladb/gocqlx/v3/qb"
	"golang.org/x/sync/errgroup"
)

func (r *Messenger) GetChat(ctx context.Context, chatID int64, userID uuid.UUID) (domain.Chat, error) {
	var row db.ChatMembersStruct
	err := getChat.QueryContext(ctx, r.session).BindMap(qb.M{
		"chat_id": chatID,
		"user_id": userID,
	}).GetRelease(&row)
	if errors.Is(err, gocql.ErrNotFound) {
		return domain.Chat{}, domain.ErrChatNotFound
	}
	if err != nil {
		return domain.Chat{}, err
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

func (r *Messenger) GetChats(ctx context.Context, userID uuid.UUID) ([]domain.Chat, error) {
	var chatIDs []int64
	err := getUserChats.QueryContext(ctx, r.session).BindMap(qb.M{
		"user_id": userID,
	}).SelectRelease(&chatIDs)
	if err != nil {
		return nil, err
	}

	chats := make([]domain.Chat, len(chatIDs))
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(chatsConcurrency)
	for i, chatID := range chatIDs {
		g.Go(func() error {
			chat, err := r.GetChat(ctx, chatID, userID)
			if errors.Is(err, domain.ErrChatNotFound) {
				return nil
			}
			if err != nil {
				return err
			}

			chat.Activity, err = r.GetChatActivity(ctx, chatID)
			chats[i] = chat
			return err
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	chats = slices.DeleteFunc(chats, func(chat domain.Chat) bool {
		return chat.ID == 0 || chat.Deleted
	})
	slices.SortFunc(chats, func(a, b domain.Chat) int {
		return cmp.Compare(max(b.ID, b.Activity.LastMessageID), max(a.ID, a.Activity.LastMessageID))
	})
	return chats, nil
}

func (r *Messenger) GetChatActivity(ctx context.Context, chatID int64) (domain.ChatActivity, error) {
	var row db.ChatActivityStruct
	err := getChatActivity.QueryContext(ctx, r.session).BindMap(qb.M{
		"chat_id": chatID,
	}).GetRelease(&row)
	if errors.Is(err, gocql.ErrNotFound) {
		return domain.ChatActivity{}, nil
	}
	if err != nil {
		return domain.ChatActivity{}, err
	}

	return domain.ChatActivity{
		LastMessageID:       row.LastMessageId,
		LastMessageSenderID: row.LastMessageSenderId,
		LastMessagePreview:  row.LastMessagePreview,
	}, nil
}

func (r *Messenger) GetChatMembers(ctx context.Context, chatID int64) ([]domain.Member, error) {
	var rows []db.ChatMembersStruct
	err := getChatMembers.QueryContext(ctx, r.session).BindMap(qb.M{
		"chat_id": chatID,
	}).SelectRelease(&rows)
	if err != nil {
		return nil, err
	}

	members := make([]domain.Member, 0, len(rows))
	for _, row := range rows {
		if row.UserId == uuid.Nil {
			continue
		}

		role := domain.RoleRegular
		if row.UserId == row.AdminId {
			role = domain.RoleAdmin
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
	batch := r.session.ContextBatch(ctx, gocql.LoggedBatch)
	query := insertChatMember.Query(r.session)
	for _, member := range members {
		err := batch.BindStruct(query, db.ChatMembersStruct{
			ChatId:   chat.ID,
			UserId:   member.UserID,
			JoinedAt: member.JoinedAt,
			Type:     string(chat.Type),
			Title:    chat.Title,
			UserA:    chat.UserA,
			UserB:    chat.UserB,
			AdminId:  chat.AdminID,
			Version:  chat.Version,
		})
		if err != nil {
			return err
		}
	}

	applied, err := r.execBatchCAS(batch)
	if err != nil {
		return err
	}
	if !applied {
		return domain.ErrChatExists
	}
	return nil
}

func (r *Messenger) DeleteChat(ctx context.Context, chat domain.Chat) error {
	chat.Deleted = true
	return checkVersion(updateChat.QueryContext(ctx, r.session).BindMap(updateChatArgs(chat)).ExecCASRelease())
}

func (r *Messenger) CreateChatMember(ctx context.Context, chat domain.Chat, member domain.Member) error {
	return checkVersion(addChatMember.QueryContext(ctx, r.session).BindMap(qb.M{
		"chat_id":          chat.ID,
		"user_id":          member.UserID,
		"joined_at":        member.JoinedAt,
		"version":          chat.Version + 1,
		"expected_version": chat.Version,
	}).ExecCASRelease())
}

// admin_id и deleted записываются из chat: так выполняется выход администратора
func (r *Messenger) DeleteChatMember(ctx context.Context, chat domain.Chat, userID uuid.UUID) error {
	batch := r.session.ContextBatch(ctx, gocql.LoggedBatch)

	err := batch.BindMap(deleteChatMember.Query(r.session), qb.M{
		"chat_id": chat.ID,
		"user_id": userID,
	})
	if err != nil {
		return err
	}

	err = batch.BindMap(updateChat.Query(r.session), updateChatArgs(chat))
	if err != nil {
		return err
	}

	return checkVersion(r.execBatchCAS(batch))
}

// возвращает ID чата 2х человек, идемпотентно
func (r *Messenger) CreatePrivateChat(ctx context.Context, chat domain.Chat) (int64, error) {
	row := db.PrivateChatsStruct{
		UserA:  chat.UserA,
		UserB:  chat.UserB,
		ChatId: chat.ID,
	}
	_, err := insertPrivateChat.QueryContext(ctx, r.session).BindStruct(row).GetCASRelease(&row)
	return row.ChatId, err
}

func (r *Messenger) DeletePrivateChat(ctx context.Context, chat domain.Chat) error {
	_, err := deletePrivateChat.QueryContext(ctx, r.session).BindStruct(db.PrivateChatsStruct{
		UserA:  chat.UserA,
		UserB:  chat.UserB,
		ChatId: chat.ID,
	}).ExecCASRelease()
	return err
}

func (r *Messenger) CreateUserChats(ctx context.Context, chatID int64, members []domain.Member) error {
	batch := r.session.ContextBatch(ctx, gocql.UnloggedBatch)
	query := insertUserChat.Query(r.session)
	for _, member := range members {
		err := batch.BindStruct(query, db.UserChatsStruct{
			UserId:   member.UserID,
			ChatId:   chatID,
			JoinedAt: member.JoinedAt,
		})
		if err != nil {
			return err
		}
	}
	return r.session.ExecuteBatch(batch)
}

func (r *Messenger) DeleteUserChats(ctx context.Context, chatID int64, userIDs []uuid.UUID) error {
	batch := r.session.ContextBatch(ctx, gocql.UnloggedBatch)
	query := deleteUserChat.Query(r.session)
	for _, userID := range userIDs {
		err := batch.BindMap(query, qb.M{
			"user_id": userID,
			"chat_id": chatID,
		})
		if err != nil {
			return err
		}
	}
	return r.session.ExecuteBatch(batch)
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
