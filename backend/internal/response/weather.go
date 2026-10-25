package response

import "github.com/google/uuid"

type WeatherResponse struct {
	WeatherID uuid.UUID `json:"weather_id"`
	Longitude float64   `json:"longitude"`
	Latitude  float64   `json:"latitude"`
	FarmID    uuid.UUID `json:"-"`
}
