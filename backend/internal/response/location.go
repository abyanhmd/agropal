package response

import "github.com/google/uuid"

type LocationResponse struct {
	LocationID uuid.UUID `json:"location_id"`
	Longitude  float64   `json:"longitude"`
	Latitude   float64   `json:"latitude"`
	FarmID     uuid.UUID `json:"-"`
}
