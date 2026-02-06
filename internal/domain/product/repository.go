package product

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreateProduct(ctx context.Context, product Product) error
	CreateProductStock(ctx context.Context, stock ProductStock) error
	CreateProductPrice(ctx context.Context, price ProductPrice) error

	UpdateProduct(ctx context.Context, product Product) error
	UpdateProductStock(ctx context.Context, stock ProductStock) error
	UpdateProductPrice(ctx context.Context, price ProductPrice) error

	GetOnlyProduct(ctx context.Context, productID uuid.UUID) (Product, error)
	GetOneProduct(ctx context.Context, productID uuid.UUID) (ProductDetail, error)
	GetManyProducts(ctx context.Context, farmID uuid.UUID) ([]ProductDetail, error)

	DeleteProduct(ctx context.Context, productID uuid.UUID) error
	DeleteProductStock(ctx context.Context, stockID uuid.UUID) error
	DeleteProductPrice(ctx context.Context, priceID uuid.UUID) error
}
