package farm

import (
	dFarm "agropal/internal/domain/farm"

	"github.com/google/uuid"
)

/* ==================== Struct ====================== */

type FarmCommand struct {
	UserID uuid.UUID
	Name   string
}

/* ==================== Mapping ====================== */

func (c FarmCommand) ToDomain(userID uuid.UUID) dFarm.Farm {
	return dFarm.Farm{
		UserID: c.UserID,
		Name:   c.Name,
	}
}
