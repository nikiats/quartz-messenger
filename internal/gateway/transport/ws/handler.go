package ws

import (
	"net/http"

	"github.com/coder/websocket"

	"quartz/internal/gateway/connection"
	"quartz/internal/gateway/service"
)

type Handler struct {
	auth           *service.Auth
	messenger      *service.Messenger
	presence       *service.Presence
	registry       *connection.Registry
	originPatterns []string
}

func NewHandler(
	a *service.Auth,
	m *service.Messenger,
	p *service.Presence,
	r *connection.Registry,
	originPattern []string,
) *Handler {
	return &Handler{
		auth:           a,
		messenger:      m,
		presence:       p,
		registry:       r,
		originPatterns: originPattern,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: h.originPatterns,
	})
	if err != nil {
		return
	}

	conn.SetReadLimit(readLimit)

	connSession := &clientSession{
		wsConn:     conn,
		auth:       h.auth,
		messenger:  h.messenger,
		presence:   h.presence,
		registry:   h.registry,
		deliveries: make(chan connection.Message, deliveryQueue),
	}

	connSession.run(r.Context())
}
