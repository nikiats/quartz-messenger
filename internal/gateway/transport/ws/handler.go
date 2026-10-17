package ws

import (
	"net/http"

	"github.com/coder/websocket"

	"quartz/internal/common/auth"
	"quartz/internal/gateway/messenger"
	"quartz/internal/gateway/usecase"
)

type Handler struct {
	auth           *auth.Authentication
	messenger      *messenger.Client
	hub            *usecase.Hub
	originPatterns []string
}

func NewHandler(
	authentication *auth.Authentication,
	messengerClient *messenger.Client,
	hub *usecase.Hub,
	originPatterns []string,
) *Handler {
	return &Handler{
		auth:           authentication,
		messenger:      messengerClient,
		hub:            hub,
		originPatterns: originPatterns,
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

	session := &clientSession{
		conn:      conn,
		auth:      h.auth,
		messenger: h.messenger,
		hub:       h.hub,
		outbox:    make(chan []byte, outboxSize),
	}

	session.run(r.Context())
}
