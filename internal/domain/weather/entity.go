package weather

import (
	"agropal/internal/domain/shared"

	"github.com/google/uuid"
)

type Weather struct {
	LocationID uuid.UUID
	Longitude  float64
	Latitude   float64
	FarmID     uuid.UUID
}

var (
	SuccessCreated shared.SuccessKey = "WEATHER.CREATE_SUCCESS"
	SuccessUpdated shared.SuccessKey = "WEATHER.UPDATE_SUCCESS"
	SuccessDeleted shared.SuccessKey = "WEATHER.DELETE_SUCCESS"
)
