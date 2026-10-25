package farm

import (
	"agropal/internal/infra/database"
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreateFarm(ctx context.Context, q database.Querier, farm Farm) error
	FindFarmByID(ctx context.Context, q database.Querier, farmID uuid.UUID) (Farm, error)
	UpdateFarm(ctx context.Context, q database.Querier, farm Farm) error
	DeleteFarm(ctx context.Context, q database.Querier, farmID uuid.UUID) error

	ListFarmByUserID(ctx context.Context, q database.Querier, userID uuid.UUID) ([]Farm, error)
}
