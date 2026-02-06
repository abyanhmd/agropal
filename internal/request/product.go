package request

import (
	aProduct "agropal/internal/application/product"
	dProduct "agropal/internal/domain/product"
	"time"

	"github.com/google/uuid"
)

type ProductRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       string `json:"image"`
}

func (p *ProductRequest) ToCommand(farmID uuid.UUID) aProduct.ProductCommand {
	return aProduct.ProductCommand{
		Name:        p.Name,
		Description: p.Description,
		Image:       p.Image,
		FarmID:      farmID,
	}
}

type ProductStockRequest struct {
	DateAdded time.Time `json:"date_added"`
	Type      string    `json:"type"`
	Stock     float64   `json:"stock"`
}

func (p *ProductStockRequest) ToCommand(productID uuid.UUID) aProduct.ProductStockCommand {
	return aProduct.ProductStockCommand{
		ProductID: productID,
		DateAdded: p.DateAdded,
		Type:      dProduct.StockType(p.Type),
		Stock:     p.Stock,
	}
}

type ProductPriceRequest struct {
	DateAdded time.Time `json:"date_added"`
	Price     float64   `json:"price"`
}

func (p *ProductPriceRequest) ToCommand(productID uuid.UUID) aProduct.ProductPriceCommand {
	return aProduct.ProductPriceCommand{
		ProductID: productID,
		DateAdded: p.DateAdded,
		Price:     p.Price,
	}
}
