package bus

import "github.com/google/uuid"

// queue group инстансов мессенджера
const QueueMessenger = "messenger"

// запросы шлюза к мессенджеру (request-reply)
const (
	SubjectSendMessage = "messenger.message.send"
)

// события мессенджера для всех инстансов шлюза (broadcast)
const (
	SubjectSessionRevoked = "gateway.session.revoked"
	SubjectReceiveMessage = "gateway.message.received"
)

// события мессенджера для шлюза
func SubjectGatewayMessage(gatewayID uuid.UUID) string {
	return "gateway." + gatewayID.String() + ".message"
}

// коды ошибок ответа в заголовке Nats-Service-Error-Code
const (
	CodeInvalidRequest = "400"
	CodeForbidden      = "403"
	CodeInternal       = "500"
)
