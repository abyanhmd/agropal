package weather

import (
	dWeather "agropal/internal/domain/weather"
	"testing"

	"github.com/google/uuid"
)

func TestWeatherCommand_ToDomain(t *testing.T) {
	weatherID := uuid.New()
	farmID := uuid.New()
	cmd := WeatherCommand{Longitude: 12.34, Latitude: -56.78}

	got := cmd.ToDomain(weatherID, farmID)

	want := dWeather.Weather{
		LocationID: weatherID,
		Longitude:  12.34,
		Latitude:   -56.78,
		FarmID:     farmID,
	}

	if got != want {
		t.Errorf("ToDomain() = %+v, want %+v", got, want)
	}
}
