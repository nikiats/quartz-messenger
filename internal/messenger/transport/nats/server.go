package nats

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"

	"quartz/internal/api/bus"
	"quartz/internal/messenger/domain"
)

const maxConcurrentSends = 1024

type Server struct {
	nc       *nats.Conn
	m        Messages
	s        Sessions
	limitter chan struct{}
}

func NewServer(nc *nats.Conn, m Messages, s Sessions) *Server {
	return &Server{
		nc:       nc,
		m:        m,
		s:        s,
		limitter: make(chan struct{}, maxConcurrentSends),
	}
}

func (s *Server) Start() error {
	server, err := micro.AddService(s.nc, micro.Config{
		Name:       "messenger",
		Version:    "0.1.0",
		QueueGroup: bus.QueueMessenger,
	})
	if err != nil {
		return err
	}

	return server.AddEndpoint(
		"send-message",
		micro.HandlerFunc(s.handleSendMessage),
		micro.WithEndpointSubject(bus.SubjectSendMessage),
	)
}

func (s *Server) handleSendMessage(r micro.Request) {
	s.limitter <- struct{}{}
	go func() {
		defer func() {
			<-s.limitter
		}()
		s.sendMessage(r)
	}()
}

func (s *Server) sendMessage(r micro.Request) {
	var req bus.SendMessageRequest
	if err := json.Unmarshal(r.Data(), &req); err != nil {
		_ = r.Error(bus.CodeInvalidRequest, "invalid request", nil)
		return
	}

	ctx := context.Background()

	if err := s.s.CheckRevoked(ctx, req.SessionID); err != nil {
		respondError(r, err)
		return
	}

	message, err := s.m.Send(ctx, req.SenderID, req.ChatID, req.ClientMessageID, domain.MessageContent(req.Content))
	if err != nil {
		respondError(r, err)
		return
	}

	_ = r.RespondJSON(bus.SendMessageResponse{MessageID: message.ID, SentAt: message.SentAt})
}

func respondError(r micro.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrSessionRevoked), errors.Is(err, domain.ErrChatNotFound):
		_ = r.Error(bus.CodeForbidden, err.Error(), nil)
	default:
		_ = r.Error(bus.CodeInternal, "internal error", nil)
	}
}
