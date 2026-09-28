package connection

import (
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ChatID   uuid.UUID
	SenderID uuid.UUID
	Seqnum   int64
	Content  string
	SentAt   time.Time
}

type Conn interface {
	UserID() uuid.UUID
	SessionID() uuid.UUID
	Deliver(Message)
	Close(reason string)
}

type connSet map[Conn]struct{}

type Registry struct {
	mu        sync.RWMutex
	byUser    map[uuid.UUID]connSet
	bySession map[uuid.UUID]connSet
}

func NewRegistry() *Registry {
	return &Registry{
		byUser:    map[uuid.UUID]connSet{},
		bySession: map[uuid.UUID]connSet{},
	}
}

func (r *Registry) Add(c Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()

	add(r.byUser, c.UserID(), c)
	add(r.bySession, c.SessionID(), c)
}

func (r *Registry) Remove(c Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()

	removeConnection(r.byUser, c.UserID(), c)
	removeConnection(r.bySession, c.SessionID(), c)
}

func (r *Registry) Deliver(userID uuid.UUID, m Message) {
	for _, conn := range r.getConnections(r.byUser, userID) {
		conn.Deliver(m)
	}
}

func (r *Registry) CloseSession(sessionID uuid.UUID) {
	for _, conn := range r.getConnections(r.bySession, sessionID) {
		conn.Close("session revoked")
	}
}

func (r *Registry) getConnections(mapping map[uuid.UUID]connSet, id uuid.UUID) []Conn {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return slices.Collect(maps.Keys(mapping[id]))
}

func add(mapping map[uuid.UUID]connSet, id uuid.UUID, conn Conn) {
	if mapping[id] == nil {
		mapping[id] = connSet{}
	}

	mapping[id][conn] = struct{}{}
}

func removeConnection(mapping map[uuid.UUID]connSet, id uuid.UUID, conn Conn) {
	delete(mapping[id], conn)

	if len(mapping[id]) == 0 {
		delete(mapping, id)
	}
}
