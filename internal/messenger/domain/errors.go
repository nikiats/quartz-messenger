package domain

import "errors"

// ошибки проверки токена
var (
	ErrTokenExpired   = errors.New("token expired")
	ErrTokenInvalid   = errors.New("token is invalid")
	ErrSessionRevoked = errors.New("token is invalid")
)

// ошибки отправки сообщения
var (
	ErrChatNotFound = errors.New("chat not found")
)
