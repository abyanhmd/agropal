package farm

import (
	"github.com/google/uuid"
)

type ModelFarm struct {
	FarmID uuid.UUID `db:"farm_id"`
	UserID uuid.UUID `db:"user_id"`
	Name   string    `db:"name"`
}
