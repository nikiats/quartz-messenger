package ws

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"

	"quaztz/internal/gateway/connection"
	"quaztz/internal/gateway/messenger"
)

const (
	idleTimeout     = 60 * time.Second
	reauthGap       = 30 * time.Second
	writeTimeout    = 5 * time.Second
	readLimit       = 16 << 10
	maxContentRunes = 1000
)

// объект для управления жизненным циклом соединения пользователя со шлюзом
type clientSession struct {
	wsConn    *websocket.Conn
	messenger *messenger.Client
	registry  *connection.Registry

	userID    uuid.UUID
	sessionID uuid.UUID

	askReauthTimer *time.Timer
	expireTimer    *time.Timer
}

// getters и helpers
func (s *clientSession) UserID() uuid.UUID {
	return s.userID
}

func (s *clientSession) SessionID() uuid.UUID {
	return s.sessionID
}

func (s *clientSession) isAuthorized() bool {
	return s.userID != uuid.Nil
}

func (s *clientSession) hasSession(sess messenger.Session) bool {
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

// установка таймера сессии и запроса на повторную авторизацию
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

// отправка произвольной структуры в сериализации JSON
func (s *clientSession) send(ctx context.Context, responseStruct any) error {
	data, err := json.Marshal(responseStruct)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()

	return s.wsConn.Write(ctx, websocket.MessageText, data)
}

// функции для вызова через private HTTP API
func (s *clientSession) Deliver(message connection.Message) {
	payload := newReceiveMessage(message)
	_ = s.send(context.Background(), payload)
}

func (s *clientSession) Close(reason string) {
	_ = s.wsConn.Close(websocket.StatusPolicyViolation, reason)
}

// полный цикл работы с WebSocket соединением пользователя
func (s *clientSession) run(ctx context.Context) {
	defer s.registry.Remove(s)
	defer s.stopTimers()

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

// обработчики разобранных и валидированных запросов
func (s *clientSession) handlePing(ctx context.Context, _ pingRequest) error {
	return s.send(ctx, pongResponse{
		base: base{methodPong},
	})
}

func (s *clientSession) handleAuthorize(ctx context.Context, req authorizeRequest) error {
	resp := newAuthorizeResponse(false)

	session, err := s.messenger.ValidateSession(ctx, req.Token)

	if err != nil {
		switch {
		case errors.Is(err, messenger.ErrSessionIncorrect):
			resp.Error = &errorStruct{
				Message: "invalid session token",
				Code:    codeSessionInvalid,
			}
		default:
			resp.Error = ErrorStructInternal
		}
		defer s.Close("auth error")
		return s.send(ctx, resp)
	}

	if s.isAuthorized() && !s.hasSession(session) {
		resp.Error = &errorStruct{
			Message: "session mismatch",
			Code:    codeSessionMismatch,
		}
		defer s.Close("auth error")
		return s.send(ctx, resp)
	}

	s.userID = session.UserID
	s.sessionID = session.SessionID

	if !s.registry.Add(s) {
		resp.Error = &errorStruct{
			Message: "session was revoked",
			Code:    codeSessionInvalid,
		}
		defer s.Close("auth error")
		return s.send(ctx, resp)
	}

	s.setupSessionExpiry(session.ExpiresAt)

	resp.Success = true
	return s.send(ctx, resp)
}

func (s *clientSession) handleSendMessage(ctx context.Context, req sendMessageRequest) error {
	resp := newSendMessageResponse(req.ClientMessageID)

	sent, err := s.messenger.SendMessage(ctx, messenger.Message{
		ChatID:          req.ChatID,
		SenderID:        s.userID,
		SessionID:       s.sessionID,
		ClientMessageID: req.ClientMessageID,
		Content: messenger.MessageContent{
			Text: req.Content.Text,
		},
	})

	if err != nil {
		resp.Success = false

		switch {
		case errors.Is(err, messenger.ErrSendFailed):
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
	resp.Seqnum = sent.Seqnum
	resp.SentAt = sent.SentAt
	return s.send(ctx, resp)
}
