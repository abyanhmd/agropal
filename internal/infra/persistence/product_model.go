package persistence

import (
	dProduct "agropal/internal/domain/product"
	"time"

	"github.com/google/uuid"
)

type ProductModel struct {
	ID          uuid.UUID `db:"product_id"`
	FarmID      uuid.UUID `db:"farm_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	Image       string    `db:"image"`
}

type ProductStockModel struct {
	ID        uuid.UUID `db:"stock_id"`
	ProductID uuid.UUID `db:"product_id"`
	CreatedAt time.Time `db:"created_at"`
	Type      string    `db:"stock_type"`
	Quantity  float64   `db:"quantity"`
}

type ProductPriceModel struct {
	ID        uuid.UUID `db:"price_id"`
	ProductID uuid.UUID `db:"product_id"`
	CreatedAt time.Time `db:"created_at"`
	Price     float64   `db:"price"`
}

func ToProductModel(domain dProduct.Product) ProductModel {
	return ProductModel{
		ID:          domain.ID,
		FarmID:      domain.FarmID,
		Name:        domain.Name,
		Description: domain.Description,
		Image:       domain.Image,
	}
}

func ToStockModel(domain dProduct.ProductStock) ProductStockModel {
	return ProductStockModel{
		ID:        domain.ID,
		ProductID: domain.ProductID,
		CreatedAt: domain.DateAdded,
		Type:      string(domain.Type),
		Quantity:  domain.Stock,
	}
}

func ToPriceModel(domain dProduct.ProductPrice) ProductPriceModel {
	return ProductPriceModel{
		ID:        domain.ID,
		ProductID: domain.ProductID,
		CreatedAt: domain.DateAdded,
		Price:     domain.Price,
	}
}

func (m ProductModel) ToDomain() dProduct.Product {
	return dProduct.Product{
		ID:          m.ID,
		FarmID:      m.FarmID,
		Name:        m.Name,
		Description: m.Description,
		Image:       m.Image,
	}
}

func (m ProductStockModel) ToDomain() dProduct.ProductStock {
	return dProduct.ProductStock{
		ID:        m.ID,
		ProductID: m.ProductID,
		DateAdded: m.CreatedAt,
		Type:      dProduct.StockMovementType(m.Type),
		Stock:     m.Quantity,
	}
}

func (m ProductPriceModel) ToDomain() dProduct.ProductPrice {
	return dProduct.ProductPrice{
		ID:        m.ID,
		ProductID: m.ProductID,
		DateAdded: m.CreatedAt,
		Price:     m.Price,
	}
}
