package weather

import (
	"github.com/google/uuid"
)

type WeatherModel struct {
	ID        uuid.UUID `db:"weather_id"`
	Longitude float64   `db:"longitude"`
	Latitude  float64   `db:"latitude"`
	FarmID    uuid.UUID `db:"farm_id"`
}
