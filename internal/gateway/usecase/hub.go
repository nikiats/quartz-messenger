package usecase

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"quartz/internal/gateway/registry"
)

const unregisterTimeout = 5 * time.Second

var ErrHubClosed = errors.New("hub is closed")

type Presence interface {
	Register(ctx context.Context, userID uuid.UUID) error
	Unregister(ctx context.Context, userID uuid.UUID) error
}

type Hub struct {
	mu       sync.Mutex
	closed   bool
	registry *registry.Registry
	presence Presence
}

func NewHub(registry *registry.Registry, presence Presence) *Hub {
	return &Hub{
		registry: registry,
		presence: presence,
	}
}

func (h *Hub) Connect(ctx context.Context, conn registry.Connection) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return ErrHubClosed
	}

	h.registry.Add(conn)

	if err := h.presence.Register(ctx, conn.UserID()); err != nil {
		h.registry.Remove(conn)
		return err
	}

	return nil
}

func (h *Hub) Disconnect(conn registry.Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), unregisterTimeout)
	defer cancel()

	h.remove(ctx, conn)
}

func (h *Hub) Deliver(userID uuid.UUID, payload []byte) {
	for _, conn := range h.registry.ByUser(userID) {
		conn.Deliver(payload)
	}
}

func (h *Hub) CloseSession(sessionID uuid.UUID) {
	for _, conn := range h.registry.BySession(sessionID) {
		go conn.Close("session revoked")
	}
}

func (h *Hub) CloseAll(ctx context.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.closed = true

	for _, conn := range h.registry.All() {
		h.remove(ctx, conn)
		go conn.Close("server shutdown")
	}
}

func (h *Hub) remove(ctx context.Context, conn registry.Connection) {
	if !h.registry.Remove(conn) {
		return
	}

	if err := h.presence.Unregister(ctx, conn.UserID()); err != nil {
		slog.Error("presence unregister failed", "error", err)
	}
}
