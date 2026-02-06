package weather

import dWeather "agropal/internal/domain/weather"

func (m ModelWeather) ToDomain() dWeather.Weather {
	return dWeather.Weather{
		WeatherID: m.WeatherID,
		Longitude: m.Longitude,
		Latitude:  m.Latitude,
		FarmID:    m.FarmID,
	}
}

func DomainToModelWeather(domain dWeather.Weather) ModelWeather {
	return ModelWeather{
		WeatherID: domain.WeatherID,
		Longitude: domain.Longitude,
		Latitude:  domain.Latitude,
		FarmID:    domain.FarmID,
	}
}
