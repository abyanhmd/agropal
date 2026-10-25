package location

import (
	dLocation "agropal/internal/domain/location"

	"github.com/google/uuid"
)

/* ==================== Struct ====================== */

type LocationCommand struct {
	Longitude float64 `json:"longitude" validate:"required"`
	Latitude  float64 `json:"latitude" validate:"required"`
}

/* ==================== Mapping ====================== */

func (c LocationCommand) ToDomain(locationID, farmID uuid.UUID) dLocation.Location {
	return dLocation.Location{
		LocationID: locationID,
		Longitude:  c.Longitude,
		Latitude:   c.Latitude,
		FarmID:     farmID,
	}
}
