package persistence

import (
	dLocation "agropal/internal/domain/location"

	"github.com/google/uuid"
)

type LocationModel struct {
	ID        uuid.UUID `db:"location_id"`
	Longitude float64   `db:"longitude"`
	Latitude  float64   `db:"latitude"`
	FarmID    uuid.UUID `db:"farm_id"`
}

func ToLocationModel(domain dLocation.Location) LocationModel {
	return LocationModel{
		ID:        domain.LocationID,
		Longitude: domain.Longitude,
		Latitude:  domain.Latitude,
		FarmID:    domain.FarmID,
	}
}

func (m LocationModel) ToDomain() dLocation.Location {
	return dLocation.Location{
		LocationID: m.ID,
		Longitude:  m.Longitude,
		Latitude:   m.Latitude,
		FarmID:     m.FarmID,
	}
}
