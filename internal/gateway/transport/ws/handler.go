package ws

import (
	"net/http"

	"github.com/coder/websocket"

	"quaztz/internal/gateway/connection"
	"quaztz/internal/gateway/messenger"
)

type Handler struct {
	messenger      *messenger.Client
	registry       *connection.Registry
	originPatterns []string
}

func NewHandler(m *messenger.Client, r *connection.Registry, originPattern []string) *Handler {
	return &Handler{
		messenger:      m,
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
		wsConn:    conn,
		messenger: h.messenger,
		registry:  h.registry,
	}

	connSession.run(r.Context())
}
