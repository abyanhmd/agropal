package persistence

import (
	dWeather "agropal/internal/domain/weather"

	"github.com/google/uuid"
)

type WeatherModel struct {
	ID        uuid.UUID `db:"weather_id"`
	Longitude float64   `db:"longitude"`
	Latitude  float64   `db:"latitude"`
	FarmID    uuid.UUID `db:"farm_id"`
}

func ToWeatherModel(domain dWeather.Weather) WeatherModel {
	return WeatherModel{
		ID:        domain.WeatherID,
		Longitude: domain.Longitude,
		Latitude:  domain.Latitude,
		FarmID:    domain.FarmID,
	}
}

func (m WeatherModel) ToDomain() dWeather.Weather {
	return dWeather.Weather{
		WeatherID: m.ID,
		Longitude: m.Longitude,
		Latitude:  m.Latitude,
		FarmID:    m.FarmID,
	}
}
