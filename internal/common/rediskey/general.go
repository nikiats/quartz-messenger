package rediskey

import (
	"github.com/google/uuid"
)

func RevokedSession(sessionID uuid.UUID) string {
	return "revoked_session:" + sessionID.String()
}

func UserGateways(userID uuid.UUID) string {
	return "user_gateways:" + userID.String()
}
