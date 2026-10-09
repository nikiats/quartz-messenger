package token

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// объект для выпуска токенов
type Issuer struct {
	secret []byte
}

func NewIssuer(secret string) *Issuer {
	return &Issuer{secret: []byte(secret)}
}

func (i *Issuer) Issue(tokenType string, c Claims) (string, error) {
	claims := jwtClaims{
		SessionID: c.SessionID,
		Type:      tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Subject:   c.UserID.String(),
			ExpiresAt: jwt.NewNumericDate(c.ExpiresAt),
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(i.secret)
}
