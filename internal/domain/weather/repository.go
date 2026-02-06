package weather

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, weather Weather) error
	GetByFarmID(ctx context.Context, farmID uuid.UUID) ([]Weather, error)
	Update(ctx context.Context, weather Weather) error
	Delete(ctx context.Context, weatherID uuid.UUID) error
}
