package farm

import (
	dFarm "agropal/internal/domain/farm"
	"agropal/internal/response"
)

func DomainToFarmResponse(domain dFarm.Farm) response.FarmResponse {
	return response.FarmResponse{
		FarmID: domain.FarmID,
		UserID: domain.UserID,
		Name:   domain.Name,
	}
}
