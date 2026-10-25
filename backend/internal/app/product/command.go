package product

import (
	dProduct "agropal/internal/domain/product"
	"time"

	"github.com/google/uuid"
)

type ProductCommand struct {
	Name        *string
	Description *string
	Image       *string
}

type ProductStockCommand struct {
	DateAdded *time.Time
	Type      *dProduct.StockMovementType
	Stock     *float64
}

type ProductPriceCommand struct {
	DateAdded *time.Time
	Price     *float64
}

func (c ProductCommand) ToDomain(farmID, productID uuid.UUID) dProduct.Product {
	return dProduct.Product{
		ID:          productID,
		Name:        *c.Name,
		Description: *c.Description,
		Image:       *c.Image,
		FarmID:      farmID,
	}
}

func (c ProductStockCommand) ToDomain(productID, stockID uuid.UUID) dProduct.ProductStock {
	return dProduct.ProductStock{
		ID:        stockID,
		ProductID: productID,
		DateAdded: *c.DateAdded,
		Type:      *c.Type,
		Stock:     *c.Stock,
	}
}

func (c ProductPriceCommand) ToDomain(productID, priceID uuid.UUID) dProduct.ProductPrice {
	return dProduct.ProductPrice{
		ID:        priceID,
		ProductID: productID,
		DateAdded: *c.DateAdded,
		Price:     *c.Price,
	}
}
