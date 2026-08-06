package weather

import (
	dWeather "agropal/internal/domain/weather"

	"github.com/google/uuid"
)

/* ==================== Struct ====================== */

type WeatherCommand struct {
	Longitude float64 `json:"longitude" validate:"required"`
	Latitude  float64 `json:"latitude" validate:"required"`
}

/* ==================== Mapping ====================== */

func (c WeatherCommand) ToDomain(weatherID, farmID uuid.UUID) dWeather.Weather {
	return dWeather.Weather{
		LocationID: weatherID,
		Longitude:  c.Longitude,
		Latitude:   c.Latitude,
		FarmID:     farmID,
	}
}
