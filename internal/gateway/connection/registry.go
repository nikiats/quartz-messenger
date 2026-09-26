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

type set map[Conn]struct{}

type Registry struct {
	mu        sync.RWMutex
	byUser    map[uuid.UUID]set
	bySession map[uuid.UUID]set
}

func NewRegistry() *Registry {
	return &Registry{
		byUser:    map[uuid.UUID]set{},
		bySession: map[uuid.UUID]set{},
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

	remove(r.byUser, c.UserID(), c)
	remove(r.bySession, c.SessionID(), c)
}

func (r *Registry) Deliver(userID uuid.UUID, m Message) {
	for _, c := range r.lookup(r.byUser, userID) {
		c.Deliver(m)
	}
}

func (r *Registry) CloseSession(sessionID uuid.UUID) {
	for _, c := range r.lookup(r.bySession, sessionID) {
		c.Close("session revoked")
	}
}

func (r *Registry) lookup(index map[uuid.UUID]set, id uuid.UUID) []Conn {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return slices.Collect(maps.Keys(index[id]))
}

func add(index map[uuid.UUID]set, id uuid.UUID, c Conn) {
	if index[id] == nil {
		index[id] = set{}
	}

	index[id][c] = struct{}{}
}

func remove(index map[uuid.UUID]set, id uuid.UUID, c Conn) {
	delete(index[id], c)

	if len(index[id]) == 0 {
		delete(index, id)
	}
}
