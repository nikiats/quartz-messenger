package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const maxContentRunes = 1000

const (
	methodPing        = "ping"
	methodAuthorize   = "authorize"
	methodSendMessage = "sendMessage"

	methodPong                = "pong"
	methodAuthorizeResponse   = "authorizeResponse"
	methodSendMessageResponse = "sendMessageResponse"

	methodRequireAuthorize = "requireAuthorize"
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

type pingRequest struct {
	base
}

type authorizeRequest struct {
	base
	Token string `json:"token"`
}

type messageContent struct {
	Text string `json:"text"`
}

type sendMessageRequest struct {
	base
	ChatID          int64          `json:"chatId,string"`
	ClientMessageID uuid.UUID      `json:"clientMessageId"`
	Content         messageContent `json:"content"`
}

type errorStruct struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

type errorResponse struct {
	base
	Error errorStruct `json:"error"`
}

type pongResponse struct {
	base
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
	MessageID       int64        `json:"messageId,string,omitzero"`
	SentAt          time.Time    `json:"sentAt,omitzero"`
}

type requireAuthorize struct {
	base
}

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
		request, err := decode[sendMessageRequest](data)
		if err != nil {
			return nil, err
		}
		return request, validateSendMessageRequest(request)
	default:
		return nil, errors.New("method not allowed")
	}
}

func decode[T any](data []byte) (T, error) {
	var request T
	if err := json.Unmarshal(data, &request); err != nil {
		return request, fmt.Errorf("message parse failed: %w", err)
	}

	return request, nil
}

func validateSendMessageRequest(request sendMessageRequest) error {
	if request.ClientMessageID == uuid.Nil {
		return errors.New("clientMessageId is required")
	}

	if request.ChatID == 0 {
		return errors.New("chatId is required")
	}

	n := utf8.RuneCountInString(request.Content.Text)
	if n < 1 || n > maxContentRunes {
		return fmt.Errorf("content.text length must be [1; %d] UTF characters", maxContentRunes)
	}

	return nil
}
