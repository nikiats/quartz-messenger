package ws

import (
	"context"
	"errors"
	"log/slog"

	"quartz/internal/api/bus"
	"quartz/internal/common/auth"
	"quartz/internal/common/token"
	"quartz/internal/gateway/messenger"
)

var (
	errAuthorizationRequired = errors.New("authorization required")
	errAuthorizationRejected = errors.New("authorization rejected")
	errSessionMismatch       = errors.New("session mismatch")
	errUnknownRequest        = errors.New("unknown request")
)

func (s *clientSession) handle(ctx context.Context, data []byte) error {
	request, err := parseRequest(data)
	if err != nil {
		if !s.isAuthorized() {
			return errAuthorizationRequired
		}
		return s.send(ctx, errorResponse{
			base:  base{methodError},
			Error: errorStruct{Code: codeInvalidPayload, Message: err.Error()},
		})
	}

	if _, isAuthorize := request.(authorizeRequest); !isAuthorize && !s.isAuthorized() {
		return errAuthorizationRequired
	}

	switch request := request.(type) {
	case authorizeRequest:
		return s.handleAuthorize(ctx, request)
	case pingRequest:
		return s.send(ctx, pongResponse{base{methodPong}})
	case sendMessageRequest:
		return s.handleSendMessage(ctx, request)
	default:
		return errUnknownRequest
	}
}

func (s *clientSession) handleAuthorize(ctx context.Context, request authorizeRequest) error {
	claims, err := s.auth.AuthenticateToken(ctx, request.Token)
	if err != nil {
		return s.rejectAuthorize(ctx, err)
	}

	if s.isAuthorized() && (claims.UserID != s.userID || claims.SessionID != s.sessionID) {
		return s.rejectAuthorize(ctx, errSessionMismatch)
	}

	if !s.isAuthorized() {
		if err := s.connect(ctx, claims); err != nil {
			return s.rejectAuthorize(ctx, err)
		}
	}

	s.setupSessionExpiry(claims.ExpiresAt)

	return s.send(ctx, authorizeResponse{
		base:    base{methodAuthorizeResponse},
		Success: true,
	})
}

func (s *clientSession) connect(ctx context.Context, claims token.Claims) error {
	s.userID = claims.UserID
	s.sessionID = claims.SessionID

	if err := s.hub.Connect(ctx, s); err != nil {
		return err
	}

	revoked, err := s.auth.CheckRevoked(ctx, s.sessionID)
	if err != nil {
		return err
	}
	if revoked {
		return auth.ErrRevoked
	}

	return nil
}

func (s *clientSession) rejectAuthorize(ctx context.Context, err error) error {
	response := authorizeResponse{
		base:  base{methodAuthorizeResponse},
		Error: authorizeError(err),
	}
	if err := s.send(ctx, response); err != nil {
		return err
	}

	return errAuthorizationRejected
}

func (s *clientSession) handleSendMessage(ctx context.Context, request sendMessageRequest) error {
	sent, err := s.messenger.SendMessage(ctx, bus.SendMessageRequest{
		ChatID:          request.ChatID,
		SenderID:        s.userID,
		SessionID:       s.sessionID,
		ClientMessageID: request.ClientMessageID,
		Content:         bus.MessageContent{Text: request.Content.Text},
	})

	response := sendMessageResponse{
		base:            base{methodSendMessageResponse},
		ClientMessageID: request.ClientMessageID,
	}

	switch {
	case errors.Is(err, messenger.ErrSendFailed):
		response.Error = &errorStruct{Code: codeSendFailed, Message: "failed to send message"}
	case err != nil:
		response.Error = internalError(err)
	default:
		response.Success = true
		response.MessageID = sent.MessageID
		response.SentAt = sent.SentAt
	}

	return s.send(ctx, response)
}

func authorizeError(err error) *errorStruct {
	switch {
	case errors.Is(err, token.ErrInvalid), errors.Is(err, token.ErrExpired), errors.Is(err, auth.ErrRevoked):
		return &errorStruct{Code: codeSessionInvalid, Message: "incorrect credentials"}
	case errors.Is(err, errSessionMismatch):
		return &errorStruct{Code: codeSessionMismatch, Message: "session mismatch"}
	default:
		return internalError(err)
	}
}

func internalError(err error) *errorStruct {
	slog.Error("websocket request failed", "error", err)
	return &errorStruct{Code: codeInternal, Message: "internal error"}
}
