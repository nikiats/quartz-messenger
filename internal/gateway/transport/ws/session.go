package ws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"
	"unicode/utf8"

	"github.com/coder/websocket"
	"github.com/google/uuid"

	"quaztz/internal/gateway/connection"
	"quaztz/internal/gateway/messenger"
)

const (
	idleTimeout     = 60 * time.Second
	reauthLead      = 30 * time.Second
	writeTimeout    = 5 * time.Second
	readLimit       = 16 << 10
	maxContentRunes = 1000
)

type session struct {
	conn      *websocket.Conn
	messenger *messenger.Client
	registry  *connection.Registry

	userID    uuid.UUID
	sessionID uuid.UUID

	reauth *time.Timer
	expire *time.Timer
}

func (s *session) UserID() uuid.UUID {
	return s.userID
}

func (s *session) SessionID() uuid.UUID {
	return s.sessionID
}

func (s *session) Deliver(m connection.Message) {
	_ = s.send(context.Background(), newReceiveMessage(m))
}

func (s *session) Close(reason string) {
	_ = s.conn.Close(websocket.StatusPolicyViolation, reason)
}

func (s *session) run(ctx context.Context) {
	defer s.registry.Remove(s)
	defer s.stopTimers()

	for {
		typ, data, err := s.read(ctx)
		if err != nil {
			return
		}

		if err := s.handle(ctx, typ, data); err != nil {
			s.Close(err.Error())
			return
		}
	}
}

func (s *session) read(ctx context.Context) (websocket.MessageType, []byte, error) {
	ctx, cancel := context.WithTimeout(ctx, idleTimeout)
	defer cancel()

	return s.conn.Read(ctx)
}

func (s *session) handle(
	ctx context.Context,
	typ websocket.MessageType,
	data []byte,
) error {
	in, err := decode(typ, data)

	if !s.authorized() && in.Method != methodAuthorize {
		return errors.New("authorization required")
	}

	if err != nil {
		return s.send(ctx, newError(codeValidationFailed, err.Error()))
	}

	switch in.Method {
	case methodPing:
		return s.send(ctx, signal{Method: methodPong})
	case methodAuthorize:
		return s.authorize(ctx, in.Token)
	case methodSendMessage:
		return s.sendMessage(ctx, in)
	default:
		return s.send(ctx, newError(codeValidationFailed, "unknown method"))
	}
}

func (s *session) authorize(ctx context.Context, token string) error {
	sess, err := s.messenger.ValidateSession(ctx, token)
	if err != nil {
		slog.Warn("authorize", "error", err)
		return s.reject(ctx, "authorization failed")
	}

	if s.authorized() && !s.matches(sess) {
		return s.reject(ctx, "session mismatch")
	}

	if err := s.send(ctx, newAuthorizeResponse(true)); err != nil {
		return err
	}

	if !s.authorized() {
		s.userID = sess.UserID
		s.sessionID = sess.SessionID
		s.registry.Add(s)
	}

	s.scheduleExpiry(sess.ExpiresAt)

	return nil
}

func (s *session) authorized() bool {
	return s.userID != uuid.Nil
}

func (s *session) matches(sess messenger.Session) bool {
	return sess.UserID == s.userID && sess.SessionID == s.sessionID
}

func (s *session) reject(ctx context.Context, reason string) error {
	_ = s.send(ctx, newAuthorizeResponse(false))

	return errors.New(reason)
}

func (s *session) sendMessage(ctx context.Context, in inbound) error {
	if in.ClientMessageID == uuid.Nil {
		return s.send(ctx, newError(codeValidationFailed, "clientMessageId is required"))
	}

	resp := newSendMessageResponse(in.ClientMessageID)

	if !validMessage(in) {
		resp.Error = &errorBody{
			Code:    codeValidationFailed,
			Message: "chatId and content of 1-1000 characters are required",
		}
		return s.send(ctx, resp)
	}

	sent, err := s.messenger.SendMessage(ctx, messenger.Message{
		ChatID:          in.ChatID,
		SenderID:        s.userID,
		ClientMessageID: in.ClientMessageID,
		Content:         in.Message.Content,
	})
	if err != nil {
		slog.Warn("send message", "error", err)

		resp.Error = &errorBody{
			Code:    codeInternal,
			Message: "message was not sent",
		}
		return s.send(ctx, resp)
	}

	resp.Success = true
	resp.Seqnum = sent.Seqnum
	resp.SentAt = sent.SentAt

	return s.send(ctx, resp)
}

func (s *session) scheduleExpiry(expiresAt time.Time) {
	s.stopTimers()

	ttl := time.Until(expiresAt)

	s.reauth = time.AfterFunc(ttl-reauthLead, func() {
		_ = s.send(context.Background(), signal{Method: methodRequireAuthorize})
	})

	s.expire = time.AfterFunc(ttl, func() {
		s.Close("token expired")
	})
}

func (s *session) stopTimers() {
	if s.expire == nil {
		return
	}

	s.reauth.Stop()
	s.expire.Stop()
}

func (s *session) send(ctx context.Context, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()

	return s.conn.Write(ctx, websocket.MessageText, data)
}

func decode(typ websocket.MessageType, data []byte) (inbound, error) {
	var in inbound

	if typ != websocket.MessageText {
		return in, errors.New("text frame expected")
	}

	if err := json.Unmarshal(data, &in); err != nil {
		return in, fmt.Errorf("malformed message: %w", err)
	}

	return in, nil
}

func validMessage(in inbound) bool {
	n := utf8.RuneCountInString(in.Message.Content)

	return in.ChatID != uuid.Nil && n >= 1 && n <= maxContentRunes
}
