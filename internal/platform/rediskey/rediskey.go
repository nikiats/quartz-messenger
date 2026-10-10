package rediskey

import (
	"strconv"

	"github.com/google/uuid"
)

func Revoked(sessionID uuid.UUID) string {
	return "revoked:" + sessionID.String()
}

func Conn(userID uuid.UUID) string {
	return "conn:" + userID.String()
}

func Gateway(gatewayID uuid.UUID) string {
	return "gateway:" + gatewayID.String()
}

func SnowflakeNode(nodeNum int64) string {
	return "sfnode:" + strconv.FormatInt(nodeNum, 10)
}
