package nats

import (
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"

	"quaztz/internal/api/bus"
	"quaztz/internal/gateway/connection"
)

// прием событий мессенджера из NATS
type Subscriber struct {
	nc       *nats.Conn
	registry *connection.Registry
}

func NewSubscriber(nc *nats.Conn, registry *connection.Registry) *Subscriber {
	return &Subscriber{nc: nc, registry: registry}
}

// подписка на рассылку для всех шлюзов и на события этого шлюза
func (s *Subscriber) Start(gatewayID uuid.UUID) error {
	_, err := s.nc.Subscribe(bus.SubjectSessionRevoked, s.handleSessionRevoked)
	if err != nil {
		return err
	}
	_, err = s.nc.Subscribe(bus.SubjectGatewayMessage(gatewayID), s.handleGatewayMessage)
	return err
}

func (s *Subscriber) handleSessionRevoked(msg *nats.Msg) {
	var event bus.SessionRevokedEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		slog.Warn("invalid session revoked event", "error", err)
		return
	}

	s.registry.CloseSession(event.SessionID)
}

func (s *Subscriber) handleGatewayMessage(msg *nats.Msg) {
	var event bus.MessageEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		slog.Warn("invalid message event", "error", err)
		return
	}

	message := connection.Message{
		ChatID:   event.ChatID,
		SenderID: event.SenderID,
		Seqnum:   event.Seqnum,
		Content:  event.Content.Text,
		SentAt:   event.SentAt,
	}

	for _, userID := range event.RecipientIDs {
		s.registry.Deliver(userID, message)
	}
}
