package registry

import (
	"maps"
	"slices"
	"sync"

	"github.com/google/uuid"
)

type Connection interface {
	UserID() uuid.UUID
	SessionID() uuid.UUID

	Deliver(payload []byte)
	Close(reason string)
}

type connections map[Connection]struct{}

type Registry struct {
	mu        sync.RWMutex
	byUser    map[uuid.UUID]connections
	bySession map[uuid.UUID]connections
}

func New() *Registry {
	return &Registry{
		byUser:    map[uuid.UUID]connections{},
		bySession: map[uuid.UUID]connections{},
	}
}

func (r *Registry) Add(conn Connection) {
	r.mu.Lock()
	defer r.mu.Unlock()

	add(r.byUser, conn.UserID(), conn)
	add(r.bySession, conn.SessionID(), conn)
}

func (r *Registry) Remove(conn Connection) (lastUserConn bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.contains(conn) {
		return false
	}

	remove(r.byUser, conn.UserID(), conn)
	remove(r.bySession, conn.SessionID(), conn)

	return len(r.byUser[conn.UserID()]) == 0
}

func (r *Registry) ByUser(userID uuid.UUID) []Connection {
	return r.list(r.byUser, userID)
}

func (r *Registry) BySession(sessionID uuid.UUID) []Connection {
	return r.list(r.bySession, sessionID)
}

func (r *Registry) All() []Connection {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var all []Connection
	for _, conns := range r.byUser {
		all = slices.AppendSeq(all, maps.Keys(conns))
	}
	return all
}

func (r *Registry) contains(conn Connection) bool {
	_, found := r.byUser[conn.UserID()][conn]
	return found
}

func (r *Registry) list(index map[uuid.UUID]connections, id uuid.UUID) []Connection {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return slices.Collect(maps.Keys(index[id]))
}

func add(index map[uuid.UUID]connections, id uuid.UUID, conn Connection) {
	if index[id] == nil {
		index[id] = connections{}
	}
	index[id][conn] = struct{}{}
}

func remove(index map[uuid.UUID]connections, id uuid.UUID, conn Connection) {
	delete(index[id], conn)
	if len(index[id]) == 0 {
		delete(index, id)
	}
}
