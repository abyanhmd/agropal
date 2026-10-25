package farm

import (
	dFarm "agropal/internal/domain/farm"

	"github.com/google/uuid"
)

/* ==================== Struct ====================== */

type FarmCommand struct {
	Name string
}

/* ==================== Mapping ====================== */

func (c FarmCommand) ToDomain(userID uuid.UUID) dFarm.Farm {
	return dFarm.Farm{
		UserID: userID,
		Name:   c.Name,
	}
}
