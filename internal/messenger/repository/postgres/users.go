package postgres

import (
	"context"
	"quaztz/internal/messenger/domain"
	"quaztz/internal/messenger/repository/postgres/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
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
	panic("not implemented")
}

func (r *Users) GetUser(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	panic("not implemented")
}

func (r *Users) GetCredentials(ctx context.Context, login string) (domain.Credentials, error) {
	panic("not implemented")
}

func (r *Users) SearchUsers(ctx context.Context, username string, limit int) ([]domain.User, error) {
	panic("not implemented")
}

func (r *Users) UpdateProfile(ctx context.Context, userID uuid.UUID, visibleName, username *string) error {
	panic("not implemented")
}
