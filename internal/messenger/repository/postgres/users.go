package postgres

import (
	"context"
	"errors"
	"quaztz/internal/messenger/domain"
	"quaztz/internal/messenger/repository/postgres/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	loginUniqueConstraint    = "users_login_key"
	usernameUniqueConstraint = "users_username_key"
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
	err := r.q.CreateUser(ctx, db.CreateUserParams{
		ID:           userID,
		Login:        login,
		PasswordHash: passwordHash,
	})
	if isUniqueViolation(err, loginUniqueConstraint) {
		return domain.ErrLoginTaken
	}
	return err
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
	err := r.q.UpdateProfile(ctx, db.UpdateProfileParams{
		ID:          userID,
		VisibleName: visibleName,
		Username:    username,
	})
	if isUniqueViolation(err, usernameUniqueConstraint) {
		return domain.ErrUsernameTaken
	}
	return err
}
