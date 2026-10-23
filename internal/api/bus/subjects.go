package bus

import "github.com/google/uuid"

// запрос на отправку сообщения
const SubjectSendMessage = "messenger.message.send"
const QueueMessenger = "messenger"

// отзыв сессии
const SubjectSessionRevoked = "gateway.all.sessionrevoke"

// сообщения, передаваемые в доставку данному шлюзу
func SubjectGatewayDeliver(gatewayID uuid.UUID) string {
	return "gateway." + gatewayID.String() + ".deliver"
}

// события, связанные с пользователем
func SubjectUserEvent(userID uuid.UUID) string {
	return "user." + userID.String() + ".event"
}

// события, связанные с чатом
func SubjectChatEvent(chatID uuid.UUID) string {
	return "chat." + chatID.String() + ".event"
}

const HeaderRecipient = "Recipient"

const (
	CodeInvalidRequest = "400"
	CodeForbidden      = "403"
	CodeInternal       = "500"
)
