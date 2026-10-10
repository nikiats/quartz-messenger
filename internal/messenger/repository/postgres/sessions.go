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
	return &Sessions{
		q: db.New(pool),
	}
}

func (r *Sessions) CreateSession(ctx context.Context, session domain.Session, refreshTokenHash string) error {
	return r.q.CreateSession(ctx, db.CreateSessionParams{
		ID:               session.SessionID,
		UserID:           session.UserID,
		RefreshTokenHash: refreshTokenHash,
		ExpiresAt:        session.ExpiresAt,
	})
}

func (r *Sessions) RefreshSession(
	ctx context.Context,
	session domain.Session,
	oldRefreshTokenHash, newRefreshTokenHash string,
) error {
	affected, err := r.q.RefreshSession(ctx, db.RefreshSessionParams{
		UserID:              session.UserID,
		ID:                  session.SessionID,
		RefreshTokenHash:    oldRefreshTokenHash,
		NewRefreshTokenHash: newRefreshTokenHash,
		ExpiresAt:           session.ExpiresAt,
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrTokenInvalid
	}
	return nil
}

func (r *Sessions) RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	affected, err := r.q.RevokeSession(ctx, db.RevokeSessionParams{
		UserID: userID,
		ID:     sessionID,
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrSessionNotFound
	}
	return nil
}
