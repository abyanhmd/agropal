package farm

import dFarm "agropal/internal/domain/farm"

func ToFarmModel(domain dFarm.Farm) FarmModel {
	return FarmModel{
		ID:     domain.ID,
		Name:   domain.Name,
		UserID: domain.UserID,
	}
}

func (m FarmModel) ToDomain() dFarm.Farm {
	return dFarm.Farm{
		ID:     m.ID,
		Name:   m.Name,
		UserID: m.UserID,
	}
}
