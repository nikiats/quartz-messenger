package nats

import (
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"

	"quaztz/internal/api/bus"
)

// обработка запросов шлюзов через NATS request-reply
type Server struct {
	nc *nats.Conn
	m  Messages
	s  Sessions
}

func NewServer(nc *nats.Conn, m Messages, s Sessions) *Server {
	return &Server{nc: nc, m: m, s: s}
}

// регистрация сервиса в queue group мессенджеров
func (s *Server) Start() error {
	srv, err := micro.AddService(s.nc, micro.Config{
		Name:       "messenger",
		Version:    "0.1.0",
		QueueGroup: bus.QueueMessenger,
	})
	if err != nil {
		return err
	}

	return srv.AddEndpoint(
		"send-message",
		micro.HandlerFunc(s.handleSendMessage),
		micro.WithEndpointSubject(bus.SubjectSendMessage),
	)
}

// TODO: bus.SendMessageRequest -> CheckRevoked, Send -> bus.SendMessageResponse или r.Error
func (s *Server) handleSendMessage(r micro.Request) {
	_ = r.Error(bus.CodeInternal, "not implemented", nil)
}
