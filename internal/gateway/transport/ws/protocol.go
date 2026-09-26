package ws

import (
	"time"

	"github.com/google/uuid"

	"quaztz/internal/gateway/connection"
)

const (
	methodPing                = "ping"
	methodPong                = "pong"
	methodAuthorize           = "authorize"
	methodAuthorizeResponse   = "authorizeResponse"
	methodRequireAuthorize    = "requireAuthorize"
	methodSendMessage         = "sendMessage"
	methodSendMessageResponse = "sendMessageResponse"
	methodReceiveMessage      = "receiveMessage"
	methodError               = "error"
)

const (
	codeValidationFailed = "validation_failed"
	codeInternal         = "internal"
)

type inbound struct {
	Method          string    `json:"method"`
	Token           string    `json:"token"`
	ChatID          uuid.UUID `json:"chatId"`
	ClientMessageID uuid.UUID `json:"clientMessageId"`
	Message         struct {
		Content string `json:"content"`
	} `json:"message"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type signal struct {
	Method string `json:"method"`
}

type errorMessage struct {
	Method string    `json:"method"`
	Error  errorBody `json:"error"`
}

type authorizeResponse struct {
	Method  string `json:"method"`
	Success bool   `json:"success"`
}

type sendMessageResponse struct {
	Method          string     `json:"method"`
	ClientMessageID uuid.UUID  `json:"clientMessageId"`
	Success         bool       `json:"success"`
	Error           *errorBody `json:"error,omitempty"`
	Seqnum          int64      `json:"seqnum,omitzero"`
	SentAt          time.Time  `json:"sentAt,omitzero"`
}

type receiveMessage struct {
	Method   string    `json:"method"`
	ChatID   uuid.UUID `json:"chatId"`
	SenderID uuid.UUID `json:"senderId"`
	Seqnum   int64     `json:"seqnum"`
	Message  message   `json:"message"`
}

type message struct {
	Content string    `json:"content"`
	SentAt  time.Time `json:"sentAt"`
}

func newError(code, msg string) errorMessage {
	return errorMessage{
		Method: methodError,
		Error:  errorBody{Code: code, Message: msg},
	}
}

func newAuthorizeResponse(success bool) authorizeResponse {
	return authorizeResponse{
		Method:  methodAuthorizeResponse,
		Success: success,
	}
}

func newSendMessageResponse(clientMessageID uuid.UUID) sendMessageResponse {
	return sendMessageResponse{
		Method:          methodSendMessageResponse,
		ClientMessageID: clientMessageID,
	}
}

func newReceiveMessage(m connection.Message) receiveMessage {
	return receiveMessage{
		Method:   methodReceiveMessage,
		ChatID:   m.ChatID,
		SenderID: m.SenderID,
		Seqnum:   m.Seqnum,
		Message: message{
			Content: m.Content,
			SentAt:  m.SentAt,
		},
	}
}
