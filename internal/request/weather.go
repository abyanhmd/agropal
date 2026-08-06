package request

import (
	aWeather "agropal/internal/app/weather"
)

type WeatherRequest struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

func (r *WeatherRequest) ToCommand() aWeather.WeatherCommand {
	return aWeather.WeatherCommand{
		Longitude: r.Longitude,
		Latitude:  r.Latitude,
	}
}
