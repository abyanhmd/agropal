package location

import (
	"agropal/internal/domain/shared"

	"github.com/google/uuid"
)

type Location struct {
	LocationID uuid.UUID
	Longitude  float64
	Latitude   float64
	FarmID     uuid.UUID
}

var (
	SuccessCreated shared.SuccessKey = "LOCATION.CREATE_SUCCESS"
	SuccessUpdated shared.SuccessKey = "LOCATION.UPDATE_SUCCESS"
	SuccessDeleted shared.SuccessKey = "LOCATION.DELETE_SUCCESS"
)
