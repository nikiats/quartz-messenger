package nats

import (
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

// подписка на события, общие для всех инстансов шлюза
func (s *Subscriber) Start() error {
	_, err := s.nc.Subscribe(bus.SubjectSessionRevoked, s.handleSessionRevoked)
	return err
}

// TODO: подписка на bus.SubjectUserMessage при открытии первого соединения пользователя
func (s *Subscriber) SubscribeUser(userID uuid.UUID) error {
	return nil
}

// TODO: отписка при закрытии последнего соединения пользователя
func (s *Subscriber) UnsubscribeUser(userID uuid.UUID) {
}

// TODO: bus.SessionRevokedEvent -> registry.CloseSession
func (s *Subscriber) handleSessionRevoked(msg *nats.Msg) {
}

// TODO: bus.MessageEvent -> registry.Deliver
func (s *Subscriber) handleUserMessage(msg *nats.Msg) {
}
