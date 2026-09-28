package domain

import "github.com/google/uuid"

type User struct {
	ID          uuid.UUID
	Username    string
	VisibleName string
}
