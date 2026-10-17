package bus

import (
	"strings"

	"github.com/google/uuid"
)

const QueueMessenger = "messenger"

const (
	SubjectSendMessage = "messenger.message.send"
)

const (
	SubjectSessionRevoked = "gateway.session.revoked"
)

func SubjectGatewayUser(gatewayID, userID uuid.UUID) string {
	return gatewayUsersPrefix(gatewayID) + userID.String()
}

func SubjectGatewayUsers(gatewayID uuid.UUID) string {
	return gatewayUsersPrefix(gatewayID) + "*"
}

func UserFromGatewaySubject(subject string) (uuid.UUID, error) {
	return uuid.Parse(subject[strings.LastIndexByte(subject, '.')+1:])
}

func gatewayUsersPrefix(gatewayID uuid.UUID) string {
	return "gateway." + gatewayID.String() + ".user."
}

const (
	CodeInvalidRequest = "400"
	CodeForbidden      = "403"
	CodeInternal       = "500"
)
