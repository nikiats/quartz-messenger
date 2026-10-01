package postgres

import (
	"context"
	"errors"
	"quaztz/internal/messenger/domain"
	"quaztz/internal/messenger/repository/postgres/db"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
	return r.q.CreateSession(ctx, db.CreateSessionParams{
		ID:               session.SessionID,
		UserID:           session.UserID,
		RefreshTokenHash: refreshTokenHash,
		ExpiresAt:        session.ExpiresAt,
	})
}

func (r *Sessions) RefreshSession(
	ctx context.Context,
	oldRefreshTokenHash, newRefreshTokenHash string,
	expiresAt time.Time,
) (domain.Session, error) {
	var session domain.Session

	err := inTx(ctx, r.pool, r.q, func(q *db.Queries) error {
		row, err := q.GetSessionForRefresh(ctx, oldRefreshTokenHash)
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return domain.ErrTokenInvalid
		case err != nil:
			return err
		case row.RevokedAt != nil:
			return domain.ErrSessionRevoked
		case row.ExpiresAt.Before(time.Now()):
			return domain.ErrTokenExpired
		case row.Suspended:
			return domain.ErrSuspendedUser
		}

		err = q.UpdateRefreshToken(ctx, db.UpdateRefreshTokenParams{
			ID:               row.ID,
			RefreshTokenHash: newRefreshTokenHash,
			ExpiresAt:        expiresAt,
		})
		if err != nil {
			return err
		}

		session = domain.Session{
			UserID:    row.UserID,
			SessionID: row.ID,
			ExpiresAt: expiresAt,
		}
		return nil
	})

	return session, err
}

func (r *Sessions) RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	affected, err := r.q.RevokeSession(ctx, db.RevokeSessionParams{
		ID:     sessionID,
		UserID: userID,
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrSessionNotFound
	}
	return nil
}

func (r *Sessions) GetRevokedSessions(ctx context.Context, since time.Time) (map[uuid.UUID]time.Time, error) {
	rows, err := r.q.GetRevokedSessions(ctx, since)
	if err != nil {
		return nil, err
	}

	revoked := make(map[uuid.UUID]time.Time, len(rows))
	for _, row := range rows {
		revoked[row.ID] = *row.RevokedAt
	}
	return revoked, nil
}
