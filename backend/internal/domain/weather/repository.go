package weather

import (
	"agropal/internal/infra/database"
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreateWeatherLocation(ctx context.Context, q database.Querier, weather Weather) error
	GetWeatherLocationByID(ctx context.Context, q database.Querier, id uuid.UUID) (Weather, error)
	ListWeatherLocationsByFarmID(ctx context.Context, q database.Querier, farmID uuid.UUID) ([]Weather, error)
	UpdateWeatherLocation(ctx context.Context, q database.Querier, weather Weather) error
	DeleteWeatherLocation(ctx context.Context, q database.Querier, id uuid.UUID) error
}
