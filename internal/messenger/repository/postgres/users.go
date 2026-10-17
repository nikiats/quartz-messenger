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

type Users struct {
	q *db.Queries
}

func NewUsers(pool *pgxpool.Pool) *Users {
	return &Users{q: db.New(pool)}
}

func (r *Users) CreateUser(ctx context.Context, userID uuid.UUID, login, passwordHash string) error {
	err := r.q.CreateUser(ctx, db.CreateUserParams{ID: userID, Login: login, PasswordHash: passwordHash})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return domain.ErrLoginTaken
		}
		return err
	}

	return nil
}

func (r *Users) GetUser(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	row, err := r.q.GetUser(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	}

	return domain.User(row), err
}

func (r *Users) GetVisibleNames(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	rows, err := r.q.GetVisibleNames(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	names := make(map[uuid.UUID]string, len(rows))
	for _, row := range rows {
		names[row.ID] = row.VisibleName
	}
	return names, nil
}

func (r *Users) GetCredentials(ctx context.Context, login string) (domain.Credentials, error) {
	row, err := r.q.GetCredentials(ctx, login)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Credentials{}, domain.ErrUserNotFound
	}

	return domain.Credentials(row), err
}

func (r *Users) SearchUsers(ctx context.Context, username string, limit int) ([]domain.User, error) {
	rows, err := r.q.SearchUsers(ctx, db.SearchUsersParams{Query: username, RowLimit: int32(limit)})
	if err != nil {
		return nil, err
	}

	users := make([]domain.User, len(rows))
	for i, row := range rows {
		users[i] = domain.User(row)
	}
	return users, nil
}

func (r *Users) UpdateProfile(ctx context.Context, userID uuid.UUID, visibleName, username *string) error {
	err := r.q.UpdateProfile(ctx, db.UpdateProfileParams{ID: userID, VisibleName: visibleName, Username: username})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return domain.ErrUsernameTaken
		}
		return err
	}

	return nil
}

func (r *Users) IsActiveUser(ctx context.Context, userID uuid.UUID) (bool, error) {
	return r.q.IsActiveUser(ctx, userID)
}

func (r *Users) CountActiveUsers(ctx context.Context, userIDs []uuid.UUID) (int, error) {
	count, err := r.q.CountActiveUsers(ctx, userIDs)
	return int(count), err
}
