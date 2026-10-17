package http

import (
	"errors"
	"fmt"
	api "quartz/internal/api/messagepublic"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	maxTitleRunes  = 100
	maxChatMembers = 100

	minLoginLength    = 3
	maxLoginLength    = 12
	minPasswordLength = 8
	maxPasswordLength = 32

	minUsernameLength    = 3
	maxUsernameLength    = 24
	minVisibleNameLength = 1
	maxVisibleNameLength = 32

	defaultMessagesLimit = 10
	maxMessagesLimit     = 100
	maxSearchLength      = 100
	maxContentTextLength = 1000
)

func validateChatCreate(userID uuid.UUID, chat *api.ChatCreate) error {
	switch chat.Type {
	case api.ChatCreateTypePrivate:
		if chat.Title != nil {
			return errors.New("no title allowed for private chat")
		}
		if len(chat.MemberIds) != 1 {
			return errors.New("exactly 1 member for private chat needed")
		}
	case api.ChatCreateTypeGroup:
		if chat.ClientRequestId == nil {
			return errors.New("clientRequestId is required")
		}
		if chat.Title == nil {
			return errors.New("title is required")
		}
		n := utf8.RuneCountInString(*chat.Title)
		if n < 1 || n > maxTitleRunes {
			return errors.New("title length must be [1; 100]")
		}
	default:
		return errors.New("incorrect chat type")
	}

	if len(chat.MemberIds) > maxChatMembers {
		return errors.New("maximum 100 members on chat creation")
	}

	seen := map[uuid.UUID]bool{userID: true}
	for _, id := range chat.MemberIds {
		if seen[id] {
			return errors.New("need unique members without current user")
		}
		seen[id] = true
	}

	return nil
}

func validateProfileEdit(profile *api.ProfileEdit) error {
	if profile.VisibleName != nil {
		n := utf8.RuneCountInString(*profile.VisibleName)
		if n < minVisibleNameLength || n > maxVisibleNameLength {
			return fmt.Errorf("visibleName length must be [%d; %d]", minVisibleNameLength, maxVisibleNameLength)
		}

	}
	if profile.Username != nil {
		n := utf8.RuneCountInString(*profile.Username)
		if n < minUsernameLength || n > maxUsernameLength {
			return fmt.Errorf("username length must be [%d; %d]", minUsernameLength, maxUsernameLength)
		}
	}

	return nil
}

func validateMessagesParams(params api.GetMessagesParams) error {
	if (params.Before == nil) == (params.After == nil) {
		return errors.New("allowed once of: before/after")
	}

	if params.Limit != nil && (*params.Limit < 1 || *params.Limit > maxMessagesLimit) {
		return fmt.Errorf("limit must be [1; %d]", maxMessagesLimit)
	}

	if params.Q != nil {
		if params.Before == nil {
			return errors.New("search allows only \"before\" pagination method")
		}
		n := utf8.RuneCountInString(*params.Q)
		if n < 1 || n > maxSearchLength {
			return fmt.Errorf("q length must be [1; %d]", maxSearchLength)
		}
	}

	return nil
}

func validateUserCreate(user *api.UserCreate) error {
	n := utf8.RuneCountInString(user.Login)
	if n < minLoginLength || n > maxLoginLength {
		return fmt.Errorf("login length must be [%d; %d]", minLoginLength, maxLoginLength)
	}

	n = utf8.RuneCountInString(user.Password)
	if n < minPasswordLength || n > maxPasswordLength {
		return fmt.Errorf("password length must be [%d; %d]", minPasswordLength, maxPasswordLength)
	}

	return nil
}

func validateSendMessage(send *api.SendMessage) error {
	if send.ClientMessageId == uuid.Nil {
		return errors.New("clientMessageId is required")
	}

	n := utf8.RuneCountInString(send.Content.Text)
	if n < 1 || n > maxContentTextLength {
		return fmt.Errorf("content length must be [1; %d]", maxContentTextLength)
	}

	return nil
}
