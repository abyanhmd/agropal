package weather

import dWeather "agropal/internal/domain/weather"

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
