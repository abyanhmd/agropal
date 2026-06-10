package persistence

import (
	dFarm "agropal/internal/domain/farm"

	"github.com/google/uuid"
)

type FarmModel struct {
	ID     uuid.UUID `db:"farm_id"`
	Name   string    `db:"name"`
	UserID uuid.UUID `db:"user_id"`
}

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
