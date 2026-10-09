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
)

// события мессенджера для пользователя, на них подписан шлюз с его соединениями
func SubjectUserMessage(userID uuid.UUID) string {
	return "gateway.user." + userID.String() + ".message"
}

// коды ошибок ответа в заголовке Nats-Service-Error-Code
const (
	CodeInvalidRequest = "400"
	CodeForbidden      = "403"
	CodeInternal       = "500"
)
