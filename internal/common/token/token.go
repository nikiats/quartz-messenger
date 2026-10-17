package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	Access  = "access"
	Refresh = "refresh"
)

var (
	ErrInvalid = errors.New("token invalid")
	ErrExpired = errors.New("token expired")
)

type Claims struct {
	UserID    uuid.UUID
	SessionID uuid.UUID

	ExpiresAt time.Time
}

type jwtClaims struct {
	SessionID uuid.UUID `json:"sid"`
	Type      string    `json:"typ"`
	jwt.RegisteredClaims
}
