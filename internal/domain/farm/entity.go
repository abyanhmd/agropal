package farm

import "github.com/google/uuid"

type Farm struct {
	ID     uuid.UUID
	Name   string
	UserID uuid.UUID
}
