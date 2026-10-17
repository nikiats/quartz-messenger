package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"

	"quartz/internal/common/auth"
	"quartz/internal/gateway/messenger"
	"quartz/internal/gateway/usecase"
)

const (
	idleTimeout  = 60 * time.Second
	reauthGap    = 30 * time.Second
	writeTimeout = 5 * time.Second
	readLimit    = 16 << 10
	outboxSize   = 256
)

type clientSession struct {
	conn      *websocket.Conn
	auth      *auth.Authentication
	messenger *messenger.Client
	hub       *usecase.Hub

	userID    uuid.UUID
	sessionID uuid.UUID

	askReauthTimer *time.Timer
	expireTimer    *time.Timer

	outbox     chan []byte
	overflowed atomic.Bool
}

func (s *clientSession) UserID() uuid.UUID {
	return s.userID
}

func (s *clientSession) SessionID() uuid.UUID {
	return s.sessionID
}

func (s *clientSession) Deliver(payload []byte) {
	select {
	case s.outbox <- payload:
	default:
		if !s.overflowed.Swap(true) {
			go s.close(websocket.StatusTryAgainLater, "delivery queue overflow")
		}
	}
}

func (s *clientSession) Close(reason string) {
	s.close(websocket.StatusPolicyViolation, reason)
}

func (s *clientSession) run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer s.hub.Disconnect(s)
	defer s.stopTimers()

	go s.writeOutbox(ctx)

	idleTimer := time.AfterFunc(idleTimeout, func() {
		s.close(websocket.StatusPolicyViolation, "idle timeout")
	})
	defer idleTimer.Stop()

	for {
		messageType, data, err := s.conn.Read(ctx)
		if err != nil {
			return
		}
		idleTimer.Reset(idleTimeout)

		if messageType != websocket.MessageText {
			s.close(websocket.StatusUnsupportedData, "expected text message")
			return
		}

		if err := s.handle(ctx, data); err != nil {
			slog.Debug("closing websocket session", "user", s.userID, "error", err)
			s.close(websocket.StatusPolicyViolation, "request rejected")
			return
		}
	}
}

func (s *clientSession) writeOutbox(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case payload := <-s.outbox:
			if err := s.write(ctx, payload); err != nil {
				return
			}
		}
	}
}

func (s *clientSession) send(ctx context.Context, response any) error {
	payload, err := json.Marshal(response)
	if err != nil {
		return err
	}

	return s.write(ctx, payload)
}

func (s *clientSession) write(ctx context.Context, payload []byte) error {
	ctx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()

	return s.conn.Write(ctx, websocket.MessageText, payload)
}

func (s *clientSession) close(status websocket.StatusCode, reason string) {
	_ = s.conn.Close(status, reason)
}

func (s *clientSession) isAuthorized() bool {
	return s.userID != uuid.Nil
}

func (s *clientSession) setupSessionExpiry(expiresAt time.Time) {
	s.stopTimers()

	ttl := time.Until(expiresAt)

	s.askReauthTimer = time.AfterFunc(ttl-reauthGap, func() {
		_ = s.send(context.Background(), requireAuthorize{base{methodRequireAuthorize}})
	})

	s.expireTimer = time.AfterFunc(ttl, func() {
		s.close(websocket.StatusNormalClosure, "session expired")
	})
}

func (s *clientSession) stopTimers() {
	if s.askReauthTimer != nil {
		s.askReauthTimer.Stop()
	}
	if s.expireTimer != nil {
		s.expireTimer.Stop()
	}
}
