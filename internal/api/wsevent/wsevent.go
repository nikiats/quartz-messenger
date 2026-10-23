package wsevent

const (
	MethodReceiveMessage         = "receiveMessage"
	MethodUserOnlineStatusChange = "userOnlineStatusChange"
)

type WSEvent struct {
	Method string `json:"method"`
	Data   any    `json:"data"`
}
