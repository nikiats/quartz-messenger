package ws

import (
	"net/http"

	"github.com/coder/websocket"

	"quaztz/internal/gateway/connection"
	"quaztz/internal/gateway/service"
)

type Handler struct {
	auth           *service.Auth
	messenger      *service.Messenger
	registry       *connection.Registry
	originPatterns []string
}

func NewHandler(a *service.Auth, m *service.Messenger, r *connection.Registry, originPattern []string) *Handler {
	return &Handler{
		auth:           a,
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
		auth:      h.auth,
		messenger: h.messenger,
		registry:  h.registry,
	}

	connSession.run(r.Context())
}
