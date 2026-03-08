package transaction

import (
	"agropal/internal/database"
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, q database.Querier, t Transaction) error
	FindByID(ctx context.Context, q database.Querier, id uuid.UUID) (Transaction, error)
	ListByFarmID(ctx context.Context, q database.Querier, farmID uuid.UUID) ([]Transaction, error)
	Update(ctx context.Context, q database.Querier, t Transaction) error
	Delete(ctx context.Context, q database.Querier, id uuid.UUID) error
}
