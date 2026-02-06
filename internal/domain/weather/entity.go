package weather

import "github.com/google/uuid"

type Weather struct {
	WeatherID uuid.UUID
	Longitude float64
	Latitude  float64
	FarmID    uuid.UUID
}
