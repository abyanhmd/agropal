package request

import (
	aWeather "agropal/internal/application/weather"

	"github.com/google/uuid"
)

type WeatherRequest struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

func (r *WeatherRequest) ToCommand(farmID uuid.UUID) aWeather.WeatherCommand {
	return aWeather.WeatherCommand{
		Longitude: r.Longitude,
		Latitude:  r.Latitude,
		FarmID:    farmID,
	}
}
