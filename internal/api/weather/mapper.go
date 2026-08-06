package weather

import (
	dWeather "agropal/internal/domain/weather"
	"agropal/internal/response"
)

func DomainToWeatherResponse(weather dWeather.Weather) response.WeatherResponse {
	return response.WeatherResponse{
		WeatherID: weather.LocationID,
		Longitude: weather.Longitude,
		Latitude:  weather.Latitude,
		FarmID:    weather.FarmID,
	}
}
