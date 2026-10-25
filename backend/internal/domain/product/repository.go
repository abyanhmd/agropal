package product

import (
	"agropal/internal/infra/database"
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreateProduct(ctx context.Context, q database.Querier, p Product) error
	UpdateProduct(ctx context.Context, q database.Querier, p Product) error
	DeleteProduct(ctx context.Context, q database.Querier, id uuid.UUID) error

	FindProductByID(ctx context.Context, q database.Querier, id uuid.UUID) (Product, error)
	FindProductDetailByID(ctx context.Context, q database.Querier, id uuid.UUID) (ProductDetail, error)
	ListProductsByFarmID(ctx context.Context, q database.Querier, farmID uuid.UUID) ([]ProductDetail, error)

	AddProductStock(ctx context.Context, q database.Querier, s ProductStock) error
	AddProductPrice(ctx context.Context, q database.Querier, p ProductPrice) error

	DeleteProductStock(ctx context.Context, q database.Querier, stockID uuid.UUID) error
	DeleteProductPrice(ctx context.Context, q database.Querier, priceID uuid.UUID) error
}
