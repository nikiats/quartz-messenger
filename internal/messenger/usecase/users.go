package usecase

import (
	"context"
	"quaztz/internal/messenger/domain"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
)

const searchLimit = 20

type Users struct {
	r              UserRepository
	passwordParams *argon2id.Params
}

func NewUsers(r UserRepository, passwordParams *argon2id.Params) *Users {
	return &Users{
		r:              r,
		passwordParams: passwordParams,
	}
}

func (u *Users) Register(ctx context.Context, login, password string) error {
	hash, err := argon2id.CreateHash(password, u.passwordParams)
	if err != nil {
		return err
	}

	return u.r.CreateUser(ctx, uuid.New(), login, hash)
}

func (u *Users) Get(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	return u.r.GetUser(ctx, userID)
}

func (u *Users) Search(ctx context.Context, query string) ([]domain.User, error) {
	if query == "" {
		return nil, nil
	}

	return u.r.SearchUsers(ctx, query, searchLimit)
}

func (u *Users) EditProfile(ctx context.Context, userID uuid.UUID, visibleName, username *string) error {
	return u.r.UpdateProfile(ctx, userID, visibleName, username)
}
