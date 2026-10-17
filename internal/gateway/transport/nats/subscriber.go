package nats

import (
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"

	"quartz/internal/api/bus"
	"quartz/internal/gateway/usecase"
)

type Subscriber struct {
	nc  *nats.Conn
	hub *usecase.Hub
}

func NewSubscriber(nc *nats.Conn, hub *usecase.Hub) *Subscriber {
	return &Subscriber{nc: nc, hub: hub}
}

func (s *Subscriber) Start(gatewayID uuid.UUID) error {
	if _, err := s.nc.Subscribe(bus.SubjectSessionRevoked, s.handleSessionRevoked); err != nil {
		return err
	}

	_, err := s.nc.Subscribe(bus.SubjectGatewayUsers(gatewayID), s.handleUserEvent)
	return err
}

func (s *Subscriber) handleSessionRevoked(msg *nats.Msg) {
	var event bus.SessionRevokedEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		slog.Warn("invalid session revoked event", "error", err)
		return
	}

	s.hub.CloseSession(event.SessionID)
}

func (s *Subscriber) handleUserEvent(msg *nats.Msg) {
	userID, err := bus.UserFromGatewaySubject(msg.Subject)
	if err != nil {
		slog.Warn("invalid user event subject", "subject", msg.Subject, "error", err)
		return
	}

	s.hub.Deliver(userID, msg.Data)
}
