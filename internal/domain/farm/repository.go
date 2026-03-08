package farm

import (
	"agropal/internal/database"
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, q database.Querier, farm Farm) error
	FindByID(ctx context.Context, q database.Querier, farmID uuid.UUID) (Farm, error)
	Update(ctx context.Context, q database.Querier, farm Farm) error
	Delete(ctx context.Context, q database.Querier, farmID uuid.UUID) error

	ListByUserID(ctx context.Context, q database.Querier, userID uuid.UUID) ([]Farm, error)
}
