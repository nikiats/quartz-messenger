package domain

import "errors"

// ошибки проверки токена
var (
	ErrTokenExpired   = errors.New("token expired")
	ErrTokenInvalid   = errors.New("token is invalid")
	ErrSessionRevoked = errors.New("session is revoked")
	ErrSuspendedUser  = errors.New("account is suspended")
)

// ошибки при работе с сессиями
var (
	ErrSessionNotFound    = errors.New("session not found")
	ErrInvalidCredentials = errors.New("invalid login or password")
)

// ошибки отправки сообщения
var (
	ErrChatNotFound = errors.New("chat not found")
)

// ошибки пользователей
var (
	ErrUserNotFound  = errors.New("user not found")
	ErrLoginTaken    = errors.New("login already taken")
	ErrUsernameTaken = errors.New("username already taken")
)

// ошибки при работе с чатами
var (
	ErrMemberExists          = errors.New("member already exists")
	ErrMemberNotFound        = errors.New("member not found")
	ErrRequiresAdmin         = errors.New("need to be admin")
	ErrAddToPrivateChat      = errors.New("forbidden to add to private chat")
	ErrRemoveFromPrivateChat = errors.New("forbidden to remove from private chat")
)
