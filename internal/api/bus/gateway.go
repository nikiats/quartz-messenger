package bus

import "github.com/google/uuid"

type SessionRevokedEvent struct {
	SessionID uuid.UUID `json:"sessionId"`
}
