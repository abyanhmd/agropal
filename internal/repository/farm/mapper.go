package farm

import dFarm "agropal/internal/domain/farm"

func (m ModelFarm) ToDomain() dFarm.Farm {
	return dFarm.Farm{
		FarmID: m.FarmID,
		UserID: m.UserID,
		Name:   m.Name,
	}
}

func DomainToModelFarm(domain dFarm.Farm) ModelFarm {
	return ModelFarm{
		FarmID: domain.FarmID,
		UserID: domain.UserID,
		Name:   domain.Name,
	}
}
