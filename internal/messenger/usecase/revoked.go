package usecase

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type revokedSessions struct {
	mu    sync.Mutex
	ttl   time.Duration
	until map[uuid.UUID]time.Time
}

func newRevokedSessions(ttl time.Duration) *revokedSessions {
	return &revokedSessions{
		ttl:   ttl,
		until: map[uuid.UUID]time.Time{},
	}
}

func (r *revokedSessions) add(sessionID uuid.UUID, revokedAt time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for id, until := range r.until {
		if until.Before(now) {
			delete(r.until, id)
		}
	}
	r.until[sessionID] = revokedAt.Add(r.ttl)
}

func (r *revokedSessions) contains(sessionID uuid.UUID) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	until, ok := r.until[sessionID]
	return ok && until.After(time.Now())
}
