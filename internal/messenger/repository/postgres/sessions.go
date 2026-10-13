package postgres

import (
	"context"
	"quaztz/internal/messenger/domain"
	"quaztz/internal/messenger/repository/postgres/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Sessions struct {
	q *db.Queries
}

func NewSessions(pool *pgxpool.Pool) *Sessions {
	return &Sessions{q: db.New(pool)}
}

func (r *Sessions) CreateSession(ctx context.Context, session domain.Session, refreshTokenHash string) error {
	params := db.CreateSessionParams{
		ID:               session.SessionID,
		UserID:           session.UserID,
		RefreshTokenHash: refreshTokenHash,
		ExpiresAt:        session.ExpiresAt,
	}

	return r.q.CreateSession(ctx, params)
}

func (r *Sessions) RefreshSession(ctx context.Context, session domain.Session, oldHash, newHash string) error {
	params := db.RefreshSessionParams{
		UserID:              session.UserID,
		ID:                  session.SessionID,
		RefreshTokenHash:    oldHash,
		NewRefreshTokenHash: newHash,
		ExpiresAt:           session.ExpiresAt,
	}

	changed, err := r.q.RefreshSession(ctx, params)
	if err != nil {
		return err
	}
	if changed == 0 {
		return domain.ErrTokenInvalid
	}

	return nil
}

func (r *Sessions) RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	params := db.RevokeSessionParams{
		UserID: userID,
		ID:     sessionID,
	}

	changed, err := r.q.RevokeSession(ctx, params)
	if err != nil {
		return err
	}
	if changed == 0 {
		return domain.ErrSessionNotFound
	}

	return nil
}
