package postgres

import (
	"context"
	"errors"
	"quaztz/internal/messenger/domain"
	"quaztz/internal/messenger/repository/postgres/db"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func inTx(ctx context.Context, pool *pgxpool.Pool, q *db.Queries, fn func(q *db.Queries) error) error {
	return pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		return fn(q.WithTx(tx))
	})
}

func lockChat(ctx context.Context, q *db.Queries, chatID, userID uuid.UUID) (db.LockChatRow, error) {
	chat, err := q.LockChat(ctx, chatID)
	if errors.Is(err, pgx.ErrNoRows) {
		return db.LockChatRow{}, domain.ErrChatNotFound
	}
	if err != nil {
		return db.LockChatRow{}, err
	}

	member, err := q.IsChatMember(ctx, db.IsChatMemberParams{ChatID: chatID, UserID: userID})
	if err != nil {
		return db.LockChatRow{}, err
	}
	if !member {
		return db.LockChatRow{}, domain.ErrChatNotFound
	}

	return chat, nil
}

func isAdmin(chat db.LockChatRow, userID uuid.UUID) bool {
	return chat.Admin != nil && *chat.Admin == userID
}

func isUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) &&
		pgErr.Code == pgerrcode.UniqueViolation &&
		pgErr.ConstraintName == constraint
}
