package request

import (
	aFarm "agropal/internal/app/farm"
)

type FarmRequest struct {
	Name string `json:"name"`
}

func (r *FarmRequest) ToCommand() aFarm.FarmCommand {
	return aFarm.FarmCommand{
		Name: r.Name,
	}
}
