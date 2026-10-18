package bus

import "github.com/google/uuid"

const QueueMessenger = "messenger"

const (
	SubjectSendMessage = "messenger.message.send"
)

const (
	SubjectSessionRevoked = "gateway.session.revoked"
)

func SubjectGatewayDeliver(gatewayID uuid.UUID) string {
	return "gateway." + gatewayID.String() + ".deliver"
}

const HeaderRecipient = "Recipient"

const (
	CodeInvalidRequest = "400"
	CodeForbidden      = "403"
	CodeInternal       = "500"
)
