package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"quaztz/internal/gateway/connection"
)

const (
	// запросы
	methodPing        = "ping"
	methodAuthorize   = "authorize"
	methodSendMessage = "sendMessage"

	// ответы
	methodPong                = "pong"
	methodAuthorizeResponse   = "authorizeResponse"
	methodSendMessageResponse = "sendMessageResponse"

	// события
	methodRequireAuthorize = "requireAuthorize"
	methodReceiveMessage   = "receiveMessage"
	methodError            = "error"
)

const (
	codeSessionInvalid  = "session_invalid"
	codeSessionMismatch = "session_mismatch"
	codeSendFailed      = "send_failed"
	codeInternal        = "internal"
	codeInvalidPayload  = "invalid_payload"
)

type base struct {
	Method string `json:"method"`
}

// схемы запросов
type pingRequest struct {
	base
}

type authorizeRequest struct {
	base

	Token string `json:"token"`
}

type sendMessageRequest struct {
	base
	ChatID          uuid.UUID `json:"chatId"`
	ClientMessageID uuid.UUID `json:"clientMessageId"`
	Message         struct {
		Content string `json:"content"`
	} `json:"message"`
}

// схемы ответов
type errorResponse struct {
	base
	Error errorStruct `json:"error"`
}

type errorStruct struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

var ErrorStructInternal = &errorStruct{
	Code:    codeInternal,
	Message: "internal error",
}

type authorizeResponse struct {
	base
	Success bool         `json:"success"`
	Error   *errorStruct `json:"error,omitempty"`
}

type sendMessageResponse struct {
	base
	ClientMessageID uuid.UUID    `json:"clientMessageId"`
	Success         bool         `json:"success"`
	Error           *errorStruct `json:"error,omitempty"`
	Seqnum          int64        `json:"seqnum,omitzero"`
	SentAt          time.Time    `json:"sentAt,omitzero"`
}

type pongResponse struct {
	base
}

// схемы отправляемых сообщений-событий
type receiveMessage struct {
	base
	ChatID   uuid.UUID `json:"chatId"`
	SenderID uuid.UUID `json:"senderId"`
	Seqnum   int64     `json:"seqnum"`
	SentAt   time.Time `json:"sentAt"`
	Message  message   `json:"message"`
}

type message struct {
	Content string `json:"content"`
}

type requireAuthorize struct {
	base
}

// конструкторы сообщений от шлюза
func newPongResponse() pongResponse {
	return pongResponse{
		base: base{Method: methodPong},
	}
}

func newErrorResponse(code, err string) errorResponse {
	return errorResponse{
		base:  base{methodError},
		Error: errorStruct{Code: code, Message: err},
	}
}

func newAuthorizeResponse(success bool) authorizeResponse {
	return authorizeResponse{
		base:    base{methodAuthorizeResponse},
		Success: success,
		Error:   nil,
	}
}

func newSendMessageResponse(clientMessageID uuid.UUID) sendMessageResponse {
	return sendMessageResponse{
		base:            base{methodSendMessageResponse},
		ClientMessageID: clientMessageID,
	}
}

func newReceiveMessage(msg connection.Message) receiveMessage {
	return receiveMessage{
		base:     base{methodReceiveMessage},
		ChatID:   msg.ChatID,
		SenderID: msg.SenderID,
		Seqnum:   msg.Seqnum,
		SentAt:   msg.SentAt,
		Message: message{
			Content: msg.Content,
		},
	}
}

// валидация
func validateSendMessageRequest(req sendMessageRequest) error {
	if req.ClientMessageID == uuid.Nil {
		return errors.New("clientMessageId is required")
	}

	if req.ChatID == uuid.Nil {
		return errors.New("chatId is required")
	}

	n := utf8.RuneCountInString(req.Message.Content)
	if n < 1 || n > maxContentRunes {
		return errors.New("message.content length must be [1; 1000] UTF characters")
	}

	return nil
}

// разбор запроса и его структурная валидация
func parseRequest(data []byte) (any, error) {
	var head base
	if err := json.Unmarshal(data, &head); err != nil {
		return nil, fmt.Errorf("message parse failed: %w", err)
	}

	switch head.Method {
	case methodPing:
		return decode[pingRequest](data)
	case methodAuthorize:
		return decode[authorizeRequest](data)
	case methodSendMessage:
		req, err := decode[sendMessageRequest](data)
		if err != nil {
			return nil, err
		}
		return req, validateSendMessageRequest(req.(sendMessageRequest))
	default:
		return nil, errors.New("method not allowed")
	}
}

func decode[T any](data []byte) (any, error) {
	var req T
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("message parse failed: %w", err)
	}

	return req, nil
}
