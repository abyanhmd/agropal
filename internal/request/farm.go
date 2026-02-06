package request

import (
	aFarm "agropal/internal/application/farm"

	"github.com/google/uuid"
)

type FarmRequest struct {
	Name string `json:"name"`
}

func (r *FarmRequest) ToCommand(userID uuid.UUID) aFarm.FarmCommand {
	return aFarm.FarmCommand{
		UserID: userID,
		Name:   r.Name,
	}
}
