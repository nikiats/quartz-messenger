package ws

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"

	"quartz/internal/gateway/connection"
	"quartz/internal/gateway/service"
)

const (
	idleTimeout     = 60 * time.Second
	reauthGap       = 30 * time.Second
	writeTimeout    = 5 * time.Second
	readLimit       = 16 << 10
	maxContentRunes = 1000
	deliveryQueue   = 256
)

type clientSession struct {
	wsConn    *websocket.Conn
	auth      *service.Auth
	messenger *service.Messenger
	presence  *service.Presence
	registry  *connection.Registry

	userID    uuid.UUID
	sessionID uuid.UUID

	askReauthTimer *time.Timer
	expireTimer    *time.Timer

	deliveries chan connection.Message
	overflowed atomic.Bool
}

func (s *clientSession) UserID() uuid.UUID {
	return s.userID
}

func (s *clientSession) SessionID() uuid.UUID {
	return s.sessionID
}

func (s *clientSession) isAuthorized() bool {
	return s.userID != uuid.Nil
}

func (s *clientSession) hasSession(sess service.Session) bool {
	return sess.UserID == s.userID && sess.SessionID == s.sessionID
}

func (s *clientSession) stopTimers() {
	if s.expireTimer != nil {
		s.expireTimer.Stop()
	}
	if s.askReauthTimer != nil {
		s.askReauthTimer.Stop()
	}
}

func (s *clientSession) setupSessionExpiry(expiresAt time.Time) {
	s.stopTimers()

	ttl := time.Until(expiresAt)

	s.askReauthTimer = time.AfterFunc(ttl-reauthGap, func() {
		_ = s.send(context.Background(), requireAuthorize{
			base: base{methodRequireAuthorize},
		})
	})

	s.expireTimer = time.AfterFunc(ttl, func() {
		s.Close("session expired")
	})
}

func (s *clientSession) send(ctx context.Context, responseStruct any) error {
	data, err := json.Marshal(responseStruct)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()

	return s.wsConn.Write(ctx, websocket.MessageText, data)
}

func (s *clientSession) Deliver(message connection.Message) {
	select {
	case s.deliveries <- message:
	default:
		if !s.overflowed.Swap(true) {
			go s.Close("delivery queue overflow")
		}
	}
}

func (s *clientSession) writeDeliveries(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case message := <-s.deliveries:
			if err := s.send(ctx, newReceiveMessage(message)); err != nil {
				return
			}
		}
	}
}

func (s *clientSession) Close(reason string) {
	_ = s.wsConn.Close(websocket.StatusPolicyViolation, reason)
}

func (s *clientSession) unregister() {
	if s.registry.Remove(s) {
		_ = s.presence.Unregister(context.Background(), s.userID)
	}
}

func (s *clientSession) run(ctx context.Context) {
	defer s.unregister()
	defer s.stopTimers()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go s.writeDeliveries(ctx)

	idleTimer := time.AfterFunc(idleTimeout, func() {
		s.Close("idle time exceeded")
	})
	defer idleTimer.Stop()

	for {
		idleTimer.Reset(idleTimeout)
		msgType, data, err := s.wsConn.Read(ctx)

		if err != nil {
			return
		}

		idleTimer.Stop()

		if msgType != websocket.MessageText {
			s.Close("expected text message")
			return
		}
		if err := s.handle(ctx, data); err != nil {
			s.Close("error during message handling")
			return
		}

		idleTimer.Reset(idleTimeout)
	}
}

func (s *clientSession) handle(ctx context.Context, data []byte) error {
	req, err := parseRequest(data)

	_, isAuth := req.(authorizeRequest)
	if !s.isAuthorized() && !isAuth {
		return errors.New("authorization required")
	}

	if err != nil {
		resp := newErrorResponse(codeInvalidPayload, err.Error())
		return s.send(ctx, resp)
	}

	switch req := req.(type) {
	case pingRequest:
		return s.handlePing(ctx, req)
	case authorizeRequest:
		return s.handleAuthorize(ctx, req)
	case sendMessageRequest:
		return s.handleSendMessage(ctx, req)
	default:
		return errors.New("unknown method")
	}
}

func (s *clientSession) handlePing(ctx context.Context, _ pingRequest) error {
	return s.send(ctx, newPongResponse())
}

func (s *clientSession) handleAuthorize(ctx context.Context, req authorizeRequest) error {
	resp := newAuthorizeResponse(false)

	session, err := s.auth.ValidateSession(ctx, req.Token)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSessionIncorrect):
			resp.Error = &errorStruct{
				Message: "incorrect credentials",
				Code:    codeSessionInvalid,
			}
		default:
			resp.Error = ErrorStructInternal
		}
		defer s.Close("authorization error")
		return s.send(ctx, resp)
	}

	if s.isAuthorized() && !s.hasSession(session) {
		resp.Error = &errorStruct{
			Message: "session mismatch",
			Code:    codeSessionMismatch,
		}
		defer s.Close("authorization error")
		return s.send(ctx, resp)
	}

	s.sessionID = session.SessionID
	s.userID = session.UserID
	s.registry.Add(s)

	revoked, err := s.auth.IsRevoked(ctx, s.sessionID)
	if err != nil {
		resp.Error = ErrorStructInternal
		defer s.Close("authorization error")
		return s.send(ctx, resp)
	}
	if revoked {
		s.registry.CloseSession(s.sessionID)

		resp.Error = &errorStruct{
			Message: "incorrect credentials",
			Code:    codeSessionInvalid,
		}
		defer s.Close("authorization error")
		return s.send(ctx, resp)
	}

	if err := s.presence.Register(ctx, s.userID); err != nil {
		resp.Error = ErrorStructInternal
		defer s.Close("authorization error")
		return s.send(ctx, resp)
	}

	s.setupSessionExpiry(session.ExpiresAt)
	resp.Success = true
	return s.send(ctx, resp)
}

func (s *clientSession) handleSendMessage(ctx context.Context, req sendMessageRequest) error {
	resp := newSendMessageResponse(req.ClientMessageID)

	sent, err := s.messenger.SendMessage(ctx, service.Message{
		ChatID:          req.ChatID,
		SenderID:        s.userID,
		SessionID:       s.sessionID,
		ClientMessageID: req.ClientMessageID,
		Content: service.MessageContent{
			Text: req.Content.Text,
		},
	})

	if err != nil {
		resp.Success = false

		switch {
		case errors.Is(err, service.ErrSendFailed):
			resp.Error = &errorStruct{
				Code:    codeSendFailed,
				Message: "failed to send message",
			}
		default:
			resp.Error = ErrorStructInternal
		}
		return s.send(ctx, resp)
	}

	resp.Success = true
	resp.MessageID = sent.MessageID
	resp.SentAt = sent.SentAt
	return s.send(ctx, resp)
}
