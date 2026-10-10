package postgres

import (
	"context"
	"errors"
	"quaztz/internal/messenger/domain"
	"quaztz/internal/messenger/repository/postgres/db"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	loginUniqueConstraint    = "user_logins_pkey"
	usernameUniqueConstraint = "usernames_pkey"
)

type Users struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewUsers(pool *pgxpool.Pool) *Users {
	return &Users{
		pool: pool,
		q:    db.New(pool),
	}
}

func (r *Users) CreateUser(ctx context.Context, userID uuid.UUID, login, passwordHash string) error {
	return inTx(ctx, r.pool, r.q, func(q *db.Queries) error {
		err := q.CreateLogin(ctx, db.CreateLoginParams{
			Login:  login,
			UserID: userID,
		})
		if isUniqueViolation(err, loginUniqueConstraint) {
			return domain.ErrLoginTaken
		}
		if err != nil {
			return err
		}

		return q.CreateUser(ctx, db.CreateUserParams{
			ID:           userID,
			Login:        login,
			PasswordHash: passwordHash,
		})
	})
}

func (r *Users) GetUser(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	row, err := r.q.GetUser(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, err
	}

	user := domain.User{
		ID:          row.ID,
		VisibleName: row.VisibleName,
	}
	if row.Username != nil {
		user.Username = *row.Username
	}
	return user, nil
}

func (r *Users) GetCredentials(ctx context.Context, login string) (domain.Credentials, error) {
	row, err := r.q.GetCredentials(ctx, login)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Credentials{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.Credentials{}, err
	}

	return domain.Credentials{
		UserID:       row.ID,
		PasswordHash: row.PasswordHash,
		Suspended:    row.Suspended,
	}, nil
}

func (r *Users) SearchUsers(ctx context.Context, username string, limit int) ([]domain.User, error) {
	rows, err := r.q.SearchUsers(ctx, db.SearchUsersParams{
		Query:    username,
		RowLimit: int32(limit),
	})
	if err != nil {
		return nil, err
	}

	users := make([]domain.User, 0, len(rows))
	for _, row := range rows {
		user := domain.User{
			ID:          row.ID,
			VisibleName: row.VisibleName,
		}
		if row.Username != nil {
			user.Username = *row.Username
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *Users) UpdateProfile(ctx context.Context, userID uuid.UUID, visibleName, username *string) error {
	return inTx(ctx, r.pool, r.q, func(q *db.Queries) error {
		if username != nil {
			if err := changeUsername(ctx, q, userID, *username); err != nil {
				return err
			}
		}

		return q.UpdateProfile(ctx, db.UpdateProfileParams{
			ID:          userID,
			VisibleName: visibleName,
			Username:    username,
		})
	})
}

func (r *Users) IsActiveUser(ctx context.Context, userID uuid.UUID) (bool, error) {
	return r.q.IsActiveUser(ctx, userID)
}

func (r *Users) CountActiveUsers(ctx context.Context, userIDs []uuid.UUID) (int, error) {
	count, err := r.q.CountActiveUsers(ctx, userIDs)
	return int(count), err
}

func changeUsername(ctx context.Context, q *db.Queries, userID uuid.UUID, username string) error {
	oldUsername, err := q.LockUsername(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrUserNotFound
	}
	if err != nil {
		return err
	}
	if oldUsername != nil && strings.EqualFold(*oldUsername, username) {
		return nil
	}

	err = q.CreateUsername(ctx, db.CreateUsernameParams{
		Username: username,
		UserID:   userID,
	})
	if isUniqueViolation(err, usernameUniqueConstraint) {
		return domain.ErrUsernameTaken
	}
	if err != nil || oldUsername == nil {
		return err
	}

	return q.DeleteUsername(ctx, *oldUsername)
}
