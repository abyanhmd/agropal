package farm

import (
	"github.com/google/uuid"
)

type FarmModel struct {
	ID     uuid.UUID `db:"farm_id"`
	Name   string    `db:"name"`
	UserID uuid.UUID `db:"user_id"`
}
