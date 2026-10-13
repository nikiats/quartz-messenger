package bus

import "github.com/google/uuid"

const QueueMessenger = "messenger"

const (
	SubjectSendMessage = "messenger.message.send"
)

const (
	SubjectSessionRevoked = "gateway.session.revoked"
)

func SubjectGatewayMessage(gatewayID uuid.UUID) string {
	return "gateway." + gatewayID.String() + ".message"
}

const (
	CodeInvalidRequest = "400"
	CodeForbidden      = "403"
	CodeInternal       = "500"
)
