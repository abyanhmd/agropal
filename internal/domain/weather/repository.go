package weather

import (
	"agropal/internal/database"
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, q database.Querier, weather Weather) error
	ListByFarmID(ctx context.Context, q database.Querier, farmID uuid.UUID) ([]Weather, error)
	Update(ctx context.Context, q database.Querier, weather Weather) error
	Delete(ctx context.Context, q database.Querier, weatherID uuid.UUID) error
}
