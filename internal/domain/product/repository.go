package product

import (
	"agropal/internal/database"
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, q database.Querier, p Product) error
	Update(ctx context.Context, q database.Querier, p Product) error
	Delete(ctx context.Context, q database.Querier, id uuid.UUID) error

	FindByID(ctx context.Context, q database.Querier, id uuid.UUID) (Product, error)
	FindDetailByID(ctx context.Context, q database.Querier, id uuid.UUID) (ProductDetail, error)
	ListByFarmID(ctx context.Context, q database.Querier, farmID uuid.UUID) ([]ProductDetail, error)

	AddStock(ctx context.Context, q database.Querier, s ProductStock) error
	AddPrice(ctx context.Context, q database.Querier, p ProductPrice) error

	DeleteStock(ctx context.Context, q database.Querier, stockID uuid.UUID) error
	DeletePrice(ctx context.Context, q database.Querier, priceID uuid.UUID) error
}
