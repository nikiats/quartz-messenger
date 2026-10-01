package postgres

import (
	"context"
	"quaztz/internal/messenger/domain"
	"quaztz/internal/messenger/repository/postgres/db"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Sessions struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewSessions(pool *pgxpool.Pool) *Sessions {
	return &Sessions{
		pool: pool,
		q:    db.New(pool),
	}
}

func (r *Sessions) CreateSession(ctx context.Context, session domain.Session, refreshTokenHash string) error {
	panic("not implemented")
}

func (r *Sessions) RefreshSession(
	ctx context.Context,
	oldRefreshTokenHash, newRefreshTokenHash string,
	expiresAt time.Time,
) (domain.Session, error) {
	panic("not implemented")
}

func (r *Sessions) RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	panic("not implemented")
}

func (r *Sessions) GetRevokedSessions(ctx context.Context, since time.Time) (map[uuid.UUID]time.Time, error) {
	panic("not implemented")
}
