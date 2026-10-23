package rediskey

import (
	"strconv"

	"github.com/google/uuid"
)

func SnowflakeInstance(nodeNum int64) string {
	return "snowflake_instance:" + strconv.FormatInt(nodeNum, 10)
}

func Message(chatID int64, senderID, clientMessageID uuid.UUID) string {
	return "message:" + strconv.FormatInt(chatID, 10) + ":" + senderID.String() + ":" + clientMessageID.String()
}

func Chat(userID, clientRequestID uuid.UUID) string {
	return "chat:" + userID.String() + ":" + clientRequestID.String()
}
