package farm

import "github.com/google/uuid"

type Farm struct {
	FarmID uuid.UUID
	UserID uuid.UUID
	Name   string
}
