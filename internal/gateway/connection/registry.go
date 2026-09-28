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
	mu         sync.RWMutex
	byUser     map[uuid.UUID]connSet
	bySession  map[uuid.UUID]connSet
	revoked    map[uuid.UUID]time.Time
	sessionTTL time.Duration
}

func NewRegistry(sessionTTL time.Duration) *Registry {
	return &Registry{
		byUser:     map[uuid.UUID]connSet{},
		bySession:  map[uuid.UUID]connSet{},
		revoked:    map[uuid.UUID]time.Time{},
		sessionTTL: sessionTTL,
	}
}

func (r *Registry) Add(conn Conn) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.revoked[conn.SessionID()]; ok {
		return false
	}

	add(r.byUser, conn.UserID(), conn)
	add(r.bySession, conn.SessionID(), conn)
	return true
}

func (r *Registry) Remove(conn Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()

	removeConnection(r.byUser, conn.UserID(), conn)
	removeConnection(r.bySession, conn.SessionID(), conn)
}

func (r *Registry) Deliver(userID uuid.UUID, msg Message) {
	for _, conn := range r.getConnections(r.byUser, userID) {
		conn.Deliver(msg)
	}
}

func (r *Registry) CloseSession(sessionID uuid.UUID) {
	r.mu.Lock()

	// очистка истекших сессий из списка отозванных
	for id, expiry := range r.revoked {
		if expiry.Before(time.Now()) {
			delete(r.revoked, id)
		}
	}
	r.revoked[sessionID] = time.Now().Add(r.sessionTTL)

	// удаление соединений из регистра соединений
	conns := slices.Collect(maps.Keys(r.bySession[sessionID]))
	for _, conn := range conns {
		removeConnection(r.byUser, conn.UserID(), conn)
		removeConnection(r.bySession, sessionID, conn)
	}

	r.mu.Unlock()

	for _, conn := range conns {
		go conn.Close("session revoked")
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
