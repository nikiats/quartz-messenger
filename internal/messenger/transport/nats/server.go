package nats

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"

	"quaztz/internal/api/bus"
	"quaztz/internal/messenger/domain"
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

func (s *Server) handleSendMessage(r micro.Request) {
	var req bus.SendMessageRequest
	if err := json.Unmarshal(r.Data(), &req); err != nil {
		_ = r.Error(bus.CodeInvalidRequest, "invalid request", nil)
		return
	}

	resp, err := s.sendMessage(context.Background(), req)

	switch {
	case err == nil:
		_ = r.RespondJSON(resp)
	case errors.Is(err, domain.ErrSessionRevoked), errors.Is(err, domain.ErrChatNotFound):
		_ = r.Error(bus.CodeForbidden, err.Error(), nil)
	default:
		_ = r.Error(bus.CodeInternal, "internal error", nil)
	}
}

func (s *Server) sendMessage(ctx context.Context, req bus.SendMessageRequest) (bus.SendMessageResponse, error) {
	if err := s.s.CheckRevoked(ctx, req.SessionID); err != nil {
		return bus.SendMessageResponse{}, err
	}

	message, err := s.m.Send(ctx, req.SenderID, req.ChatID, req.ClientMessageID, domain.MessageContent(req.Content))
	if err != nil {
		return bus.SendMessageResponse{}, err
	}

	return bus.SendMessageResponse{Seqnum: message.Seqnum, SentAt: message.SentAt}, nil
}
