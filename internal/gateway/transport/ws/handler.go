package ws

import (
	"net/http"

	"github.com/coder/websocket"

	"quaztz/internal/gateway/connection"
	"quaztz/internal/gateway/messenger"
)

type Handler struct {
	messenger *messenger.Client
	registry  *connection.Registry
}

func NewHandler(m *messenger.Client, r *connection.Registry) *Handler {
	return &Handler{messenger: m, registry: r}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return
	}

	conn.SetReadLimit(readLimit)

	s := &session{
		conn:      conn,
		messenger: h.messenger,
		registry:  h.registry,
	}

	s.run(r.Context())
}
