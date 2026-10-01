package postgres

import (
	"context"
	"errors"
	"quaztz/internal/messenger/domain"
	"quaztz/internal/messenger/repository/postgres/db"
	"slices"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Messenger struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewMessenger(pool *pgxpool.Pool) *Messenger {
	return &Messenger{
		pool: pool,
		q:    db.New(pool),
	}
}

func (r *Messenger) CreateMessage(ctx context.Context, senderID, chatID, clientMessageID uuid.UUID, content domain.MessageContent) (domain.Message, error) {
	var message domain.Message

	err := inTx(ctx, r.pool, r.q, func(q *db.Queries) error {
		if _, err := lockChat(ctx, q, chatID, senderID); err != nil {
			return err
		}

		row, err := q.InsertMessage(ctx, db.InsertMessageParams{
			ChatID:          chatID,
			SenderID:        senderID,
			ClientMessageID: clientMessageID,
			Content:         content.Text,
		})
		if err != nil {
			return err
		}

		message = domain.Message{
			ChatID:   chatID,
			SenderID: senderID,
			Seqnum:   row.Seqnum,
			SentAt:   row.SentAt,
			Content:  content,
		}
		return nil
	})

	return message, err
}

func (r *Messenger) GetMessages(ctx context.Context, chatID uuid.UUID, query domain.MessageQuery) ([]domain.Message, error) {
	limit := int32(query.Limit)

	var rows []db.Message
	var err error

	switch {
	case query.Before != nil && query.Search != "":
		rows, err = r.q.SearchMessagesBefore(ctx, db.SearchMessagesBeforeParams{
			ChatID:   chatID,
			Before:   *query.Before,
			Query:    query.Search,
			RowLimit: limit,
		})
	case query.Before != nil:
		rows, err = r.q.GetMessagesBefore(ctx, db.GetMessagesBeforeParams{
			ChatID:   chatID,
			Before:   *query.Before,
			RowLimit: limit,
		})
	case query.After != nil:
		rows, err = r.q.GetMessagesAfter(ctx, db.GetMessagesAfterParams{
			ChatID:   chatID,
			After:    *query.After,
			RowLimit: limit,
		})
	default:
		return nil, errors.New("messages query requires before or after")
	}
	if err != nil {
		return nil, err
	}

	messages := make([]domain.Message, 0, len(rows))
	for _, row := range rows {
		messages = append(messages, domain.Message{
			ChatID:   row.ChatID,
			SenderID: row.SenderID,
			Seqnum:   row.Seqnum,
			SentAt:   row.SentAt,
			Content:  domain.MessageContent{Text: row.Content},
		})
	}
	return messages, nil
}

func (r *Messenger) IsChatMember(ctx context.Context, chatID, userID uuid.UUID) (bool, error) {
	return r.q.IsChatMember(ctx, db.IsChatMemberParams{
		ChatID: chatID,
		UserID: userID,
	})
}

func (r *Messenger) GetChats(ctx context.Context, userID uuid.UUID) ([]domain.Chat, error) {
	rows, err := r.q.GetChats(ctx, userID)
	if err != nil {
		return nil, err
	}

	chats := make([]domain.Chat, 0, len(rows))
	for _, row := range rows {
		chats = append(chats, domain.Chat{
			ID:            row.ID,
			Type:          domain.ChatType(row.Type),
			Title:         row.Title,
			LastSeqnum:    row.LastSeqnum,
			LastMessageAt: row.LastMessageAt,
		})
	}
	return chats, nil
}

func (r *Messenger) GetChat(ctx context.Context, userID, chatID uuid.UUID) (domain.Chat, error) {
	return getChat(ctx, r.q, userID, chatID)
}

func (r *Messenger) GetChatMembers(ctx context.Context, chatID uuid.UUID) ([]domain.Member, error) {
	rows, err := r.q.GetChatMembers(ctx, chatID)
	if err != nil {
		return nil, err
	}

	members := make([]domain.Member, 0, len(rows))
	for _, row := range rows {
		role := domain.RoleRegular
		if row.IsAdmin {
			role = domain.RoleAdmin
		}
		members = append(members, domain.Member{
			UserID: row.UserID,
			Role:   role,
		})
	}
	return members, nil
}

func (r *Messenger) CreateChat(
	ctx context.Context,
	userID uuid.UUID,
	chatType domain.ChatType,
	title string,
	memberIDs []uuid.UUID,
) (chat domain.Chat, created bool, err error) {
	err = inTx(ctx, r.pool, r.q, func(q *db.Queries) error {
		var chatID uuid.UUID
		var err error

		switch chatType {
		case domain.ChatPrivate:
			chatID, created, err = createPrivateChat(ctx, q, userID, memberIDs[0])
		case domain.ChatGroup:
			chatID, err = createGroupChat(ctx, q, userID, title, memberIDs)
			created = true
		default:
			return errors.New("unknown chat type")
		}
		if err != nil {
			return err
		}

		chat, err = getChat(ctx, q, userID, chatID)
		return err
	})

	return chat, created, err
}

func (r *Messenger) CreateChatMember(ctx context.Context, userID, chatID uuid.UUID, memberID uuid.UUID) error {
	return inTx(ctx, r.pool, r.q, func(q *db.Queries) error {
		chat, err := lockChat(ctx, q, chatID, userID)
		if err != nil {
			return err
		}
		if chat.Type == db.ChatTypePrivate {
			return domain.ErrAddToPrivateChat
		}
		if !isAdmin(chat, userID) {
			return domain.ErrRequiresAdmin
		}

		active, err := q.IsActiveUser(ctx, memberID)
		if err != nil {
			return err
		}
		if !active {
			return domain.ErrUserNotFound
		}

		_, err = q.AddChatMember(ctx, db.AddChatMemberParams{
			ChatID: chatID,
			UserID: memberID,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrMemberExists
		}
		return err
	})
}

func (r *Messenger) DeleteChat(ctx context.Context, userID, chatID uuid.UUID) error {
	return inTx(ctx, r.pool, r.q, func(q *db.Queries) error {
		chat, err := lockChat(ctx, q, chatID, userID)
		if err != nil {
			return err
		}
		if chat.Type == db.ChatTypeGroup && !isAdmin(chat, userID) {
			return domain.ErrRequiresAdmin
		}

		return q.DeleteChat(ctx, chatID)
	})
}

func (r *Messenger) DeleteChatMember(ctx context.Context, userID, chatID uuid.UUID, memberID uuid.UUID) error {
	return inTx(ctx, r.pool, r.q, func(q *db.Queries) error {
		chat, err := lockChat(ctx, q, chatID, userID)
		if err != nil {
			return err
		}
		if chat.Type == db.ChatTypePrivate {
			return domain.ErrRemoveFromPrivateChat
		}

		leaving := memberID == userID
		if !leaving && !isAdmin(chat, userID) {
			return domain.ErrRequiresAdmin
		}

		if leaving && isAdmin(chat, userID) {
			err := q.PassAdminOnLeave(ctx, db.PassAdminOnLeaveParams{
				ChatID:  chatID,
				AdminID: userID,
			})
			if err != nil {
				return err
			}
			_, err = q.DeleteChatMember(ctx, db.DeleteChatMemberParams{
				ChatID: chatID,
				UserID: userID,
			})
			return err
		}

		affected, err := q.DeleteChatMember(ctx, db.DeleteChatMemberParams{
			ChatID: chatID,
			UserID: memberID,
		})
		if err != nil {
			return err
		}
		if affected == 0 {
			return domain.ErrMemberNotFound
		}
		return nil
	})
}

func getChat(ctx context.Context, q *db.Queries, userID, chatID uuid.UUID) (domain.Chat, error) {
	row, err := q.GetChat(ctx, db.GetChatParams{
		ChatID: chatID,
		UserID: userID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Chat{}, domain.ErrChatNotFound
	}
	if err != nil {
		return domain.Chat{}, err
	}

	return domain.Chat{
		ID:            row.ID,
		Type:          domain.ChatType(row.Type),
		Title:         row.Title,
		LastSeqnum:    row.LastSeqnum,
		LastMessageAt: row.LastMessageAt,
	}, nil
}

func createPrivateChat(ctx context.Context, q *db.Queries, userID, peerID uuid.UUID) (chatID uuid.UUID, created bool, err error) {
	active, err := q.IsActiveUser(ctx, peerID)
	if err != nil {
		return uuid.Nil, false, err
	}
	if !active {
		return uuid.Nil, false, domain.ErrUserNotFound
	}

	chatID, err = q.InsertPrivateChat(ctx, db.InsertPrivateChatParams{
		ID:     uuid.New(),
		UserID: userID,
		PeerID: peerID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		chatID, err = q.FindPrivateChat(ctx, db.FindPrivateChatParams{
			UserID: userID,
			PeerID: peerID,
		})
		return chatID, false, err
	}
	if err != nil {
		return uuid.Nil, false, err
	}

	err = q.InsertChatMembers(ctx, db.InsertChatMembersParams{
		ChatID:  chatID,
		UserIds: []uuid.UUID{userID, peerID},
	})
	return chatID, true, err
}

func createGroupChat(ctx context.Context, q *db.Queries, userID uuid.UUID, title string, memberIDs []uuid.UUID) (uuid.UUID, error) {
	found, err := q.CountActiveUsers(ctx, memberIDs)
	if err != nil {
		return uuid.Nil, err
	}
	if found != int64(len(memberIDs)) {
		return uuid.Nil, domain.ErrUserNotFound
	}

	chatID := uuid.New()
	err = q.InsertGroupChat(ctx, db.InsertGroupChatParams{
		ID:    chatID,
		Title: title,
		Admin: userID,
	})
	if err != nil {
		return uuid.Nil, err
	}

	err = q.InsertChatMembers(ctx, db.InsertChatMembersParams{
		ChatID:  chatID,
		UserIds: slices.Concat(memberIDs, []uuid.UUID{userID}),
	})
	return chatID, err
}
