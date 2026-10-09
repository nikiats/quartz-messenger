package token

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// объект для проверки токенов
type Verifier struct {
	secret []byte
}

func NewVerifier(secret string) *Verifier {
	return &Verifier{secret: []byte(secret)}
}

func (v *Verifier) Verify(token, tokenType string) (Claims, error) {
	var claims jwtClaims

	_, err := jwt.ParseWithClaims(
		token,
		&claims,
		func(*jwt.Token) (any, error) { return v.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)

	switch {
	case errors.Is(err, jwt.ErrTokenExpired):
		return Claims{}, ErrExpired
	case err != nil:
		return Claims{}, ErrInvalid
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil || claims.Type != tokenType {
		return Claims{}, ErrInvalid
	}

	return Claims{
		UserID:    userID,
		SessionID: claims.SessionID,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}
