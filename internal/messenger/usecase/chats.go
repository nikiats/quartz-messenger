package usecase

import (
	"bytes"
	"context"
	"errors"
	"math/rand/v2"
	"quaztz/internal/messenger/domain"
	"quaztz/internal/platform/snowflake"
	"slices"
	"time"

	"github.com/google/uuid"
)

const maxConflictRetries = 5

type Chats struct {
	r           MessengerRepository
	users       UserRepository
	idempotency IdempotencyRepository
	ids         IDGenerator
}

func NewChats(
	r MessengerRepository,
	users UserRepository,
	idempotency IdempotencyRepository,
	ids IDGenerator,
) *Chats {
	return &Chats{
		r:           r,
		users:       users,
		idempotency: idempotency,
		ids:         ids,
	}
}

func (c *Chats) List(ctx context.Context, userID uuid.UUID) ([]domain.Chat, error) {
	chats, err := c.r.GetChats(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := c.setPrivateTitles(ctx, userID, chats); err != nil {
		return nil, err
	}
	return chats, nil
}

func (c *Chats) Get(ctx context.Context, userID uuid.UUID, chatID int64) (domain.Chat, error) {
	chat, err := activeChat(ctx, c.r, chatID, userID)
	if err != nil {
		return domain.Chat{}, err
	}

	chat.Activity, err = c.r.GetChatActivity(ctx, chatID)
	if err != nil {
		return domain.Chat{}, err
	}

	chats := []domain.Chat{chat}
	if err := c.setPrivateTitles(ctx, userID, chats); err != nil {
		return domain.Chat{}, err
	}
	return chats[0], nil
}

func (c *Chats) Members(ctx context.Context, userID uuid.UUID, chatID int64) ([]domain.Member, error) {
	if _, err := activeChat(ctx, c.r, chatID, userID); err != nil {
		return nil, err
	}

	return c.r.GetChatMembers(ctx, chatID)
}

func (c *Chats) Create(
	ctx context.Context,
	userID uuid.UUID,
	chatType domain.ChatType,
	title string,
	memberIDs []uuid.UUID,
	clientRequestID uuid.UUID,
) (domain.Chat, bool, error) {
	var chatID int64
	var created bool
	var err error

	switch chatType {
	case domain.ChatPrivate:
		chatID, created, err = c.createPrivate(ctx, userID, memberIDs[0])
	case domain.ChatGroup:
		chatID, created, err = c.createGroup(ctx, userID, title, memberIDs, clientRequestID)
	default:
		return domain.Chat{}, false, errors.New("unknown chat type")
	}
	if err != nil {
		return domain.Chat{}, false, err
	}

	chat, err := c.Get(ctx, userID, chatID)
	return chat, created, err
}

func (c *Chats) AddMember(ctx context.Context, userID uuid.UUID, chatID int64, memberID uuid.UUID) error {
	return retryOnConflict(func() error {
		chat, err := activeChat(ctx, c.r, chatID, userID)
		if err != nil {
			return err
		}
		if chat.Type != domain.ChatGroup {
			return domain.ErrAddToPrivateChat
		}
		if chat.AdminID != userID {
			return domain.ErrRequiresAdmin
		}

		active, err := c.users.IsActiveUser(ctx, memberID)
		if err != nil {
			return err
		}
		if !active {
			return domain.ErrUserNotFound
		}

		member := domain.Member{UserID: memberID, JoinedAt: time.Now()}

		_, err = c.r.GetChat(ctx, chatID, memberID)
		if err == nil {
			if err := c.r.CreateUserChats(ctx, chatID, []domain.Member{member}); err != nil {
				return err
			}
			return domain.ErrMemberExists
		}
		if !errors.Is(err, domain.ErrChatNotFound) {
			return err
		}

		if err := c.r.CreateChatMember(ctx, chat, member); err != nil {
			return err
		}

		return c.r.CreateUserChats(ctx, chatID, []domain.Member{member})
	})
}

func (c *Chats) RemoveMember(ctx context.Context, userID uuid.UUID, chatID int64, memberID uuid.UUID) error {
	return retryOnConflict(func() error {
		chat, err := activeChat(ctx, c.r, chatID, userID)

		if errors.Is(err, domain.ErrChatNotFound) && memberID == userID {
			if err := c.r.DeleteUserChats(ctx, chatID, []uuid.UUID{userID}); err != nil {
				return err
			}
			return domain.ErrChatNotFound
		}
		if err != nil {
			return err
		}

		if chat.Type != domain.ChatGroup {
			return domain.ErrRemoveFromPrivateChat
		}
		if memberID != userID && chat.AdminID != userID {
			return domain.ErrRequiresAdmin
		}

		if memberID == chat.AdminID {
			members, err := c.r.GetChatMembers(ctx, chatID)
			if err != nil {
				return err
			}

			members = slices.DeleteFunc(members, func(member domain.Member) bool {
				return member.UserID == memberID
			})
			if len(members) == 0 {
				chat.Deleted = true
			} else {
				chat.AdminID = members[rand.IntN(len(members))].UserID
			}
		} else {
			_, err := c.r.GetChat(ctx, chatID, memberID)
			if errors.Is(err, domain.ErrChatNotFound) {
				return domain.ErrMemberNotFound
			}
			if err != nil {
				return err
			}
		}

		if err := c.r.DeleteChatMember(ctx, chat, memberID); err != nil {
			return err
		}

		return c.r.DeleteUserChats(ctx, chatID, []uuid.UUID{memberID})
	})
}

func (c *Chats) Delete(ctx context.Context, userID uuid.UUID, chatID int64) error {
	var chat domain.Chat

	err := retryOnConflict(func() error {
		var err error
		chat, err = c.r.GetChat(ctx, chatID, userID)
		if err != nil {
			return err
		}
		if chat.Type == domain.ChatGroup && chat.AdminID != userID {
			return domain.ErrRequiresAdmin
		}

		if chat.Deleted {
			return nil
		}
		return c.r.DeleteChat(ctx, chat)
	})
	if err != nil {
		return err
	}

	if chat.Type == domain.ChatPrivate {
		if err := c.r.DeletePrivateChat(ctx, chat); err != nil {
			return err
		}
	}

	members, err := c.r.GetChatMembers(ctx, chatID)
	if err != nil {
		return err
	}

	userIDs := make([]uuid.UUID, 0, len(members))
	for _, member := range members {
		userIDs = append(userIDs, member.UserID)
	}
	return c.r.DeleteUserChats(ctx, chatID, userIDs)
}

func (c *Chats) createPrivate(ctx context.Context, userID, peerID uuid.UUID) (int64, bool, error) {
	active, err := c.users.IsActiveUser(ctx, peerID)
	if err != nil {
		return 0, false, err
	}
	if !active {
		return 0, false, domain.ErrUserNotFound
	}

	userA, userB := userID, peerID
	if bytes.Compare(userA[:], userB[:]) > 0 {
		userA, userB = userB, userA
	}

	for {
		chat := domain.Chat{
			ID:      c.ids.Generate(),
			Type:    domain.ChatPrivate,
			UserA:   userA,
			UserB:   userB,
			Version: 1,
		}

		chatID, err := c.r.CreatePrivateChat(ctx, chat)
		if err != nil {
			return 0, false, err
		}

		if chatID != chat.ID {
			existing, err := c.r.GetChat(ctx, chatID, userID)

			if err == nil && existing.Deleted {
				if err := c.r.DeletePrivateChat(ctx, existing); err != nil {
					return 0, false, err
				}
				continue
			}

			if err == nil {
				return chatID, false, c.r.CreateUserChats(ctx, chatID, privateMembers(existing))
			}
			if !errors.Is(err, domain.ErrChatNotFound) {
				return 0, false, err
			}

			chat.ID = chatID
		}

		members := privateMembers(chat)
		err = c.r.CreateChat(ctx, chat, members)
		if err != nil && !errors.Is(err, domain.ErrChatExists) {
			return 0, false, err
		}

		return chat.ID, true, c.r.CreateUserChats(ctx, chat.ID, members)
	}
}

func (c *Chats) createGroup(
	ctx context.Context,
	userID uuid.UUID,
	title string,
	memberIDs []uuid.UUID,
	clientRequestID uuid.UUID,
) (int64, bool, error) {
	count, err := c.users.CountActiveUsers(ctx, memberIDs)
	if err != nil {
		return 0, false, err
	}
	if count != len(memberIDs) {
		return 0, false, domain.ErrUserNotFound
	}

	chatID, err := c.idempotency.ChatID(ctx, userID, clientRequestID, c.ids.Generate())
	if err != nil {
		return 0, false, err
	}

	joinedAt := snowflake.Time(chatID)
	members := []domain.Member{{UserID: userID, JoinedAt: joinedAt}}
	for _, memberID := range memberIDs {
		members = append(members, domain.Member{UserID: memberID, JoinedAt: joinedAt})
	}

	chat := domain.Chat{
		ID:      chatID,
		Type:    domain.ChatGroup,
		Title:   title,
		AdminID: userID,
		Version: 1,
	}
	err = c.r.CreateChat(ctx, chat, members)
	created := err == nil

	if errors.Is(err, domain.ErrChatExists) {
		members, err = c.r.GetChatMembers(ctx, chatID)
	}
	if err != nil {
		return 0, false, err
	}

	return chatID, created, c.r.CreateUserChats(ctx, chatID, members)
}

func (c *Chats) setPrivateTitles(ctx context.Context, userID uuid.UUID, chats []domain.Chat) error {
	var peerIDs []uuid.UUID
	for _, chat := range chats {
		if chat.Type == domain.ChatPrivate {
			peerIDs = append(peerIDs, peerOf(chat, userID))
		}
	}
	if len(peerIDs) == 0 {
		return nil
	}

	names, err := c.users.GetVisibleNames(ctx, peerIDs)
	if err != nil {
		return err
	}

	for i, chat := range chats {
		if chat.Type == domain.ChatPrivate {
			chats[i].Title = names[peerOf(chat, userID)]
		}
	}
	return nil
}

func activeChat(ctx context.Context, r MessengerRepository, chatID int64, userID uuid.UUID) (domain.Chat, error) {
	chat, err := r.GetChat(ctx, chatID, userID)
	if err != nil {
		return domain.Chat{}, err
	}
	if chat.Deleted {
		return domain.Chat{}, domain.ErrChatNotFound
	}
	return chat, nil
}

func retryOnConflict(fn func() error) error {
	var err error
	for range maxConflictRetries {
		err = fn()
		if !errors.Is(err, domain.ErrChatConflict) {
			return err
		}
	}
	return err
}

func peerOf(chat domain.Chat, userID uuid.UUID) uuid.UUID {
	if chat.UserA == userID {
		return chat.UserB
	}
	return chat.UserA
}

func privateMembers(chat domain.Chat) []domain.Member {
	joinedAt := snowflake.Time(chat.ID)
	return []domain.Member{
		{UserID: chat.UserA, JoinedAt: joinedAt},
		{UserID: chat.UserB, JoinedAt: joinedAt},
	}
}
