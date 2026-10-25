package location

import (
	"agropal/internal/infra/database"
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreateLocation(ctx context.Context, q database.Querier, location Location) error
	GetLocationByID(ctx context.Context, q database.Querier, id uuid.UUID) (Location, error)
	ListLocationsByFarmID(ctx context.Context, q database.Querier, farmID uuid.UUID) ([]Location, error)
	UpdateLocation(ctx context.Context, q database.Querier, location Location) error
	DeleteLocation(ctx context.Context, q database.Querier, id uuid.UUID) error
}
