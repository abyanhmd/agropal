package request

import (
	aProduct "agropal/internal/app/product"
	dProduct "agropal/internal/domain/product"
	"time"
)

// ProductRequest fields are pointers so the same struct can back both the
// create endpoint (all fields required, validated by the handler) and the
// partial-update endpoint (only the supplied fields are considered).
type ProductRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Image       *string `json:"image"`
}

func (p *ProductRequest) ToCommand() aProduct.ProductCommand {
	return aProduct.ProductCommand{
		Name:        p.Name,
		Description: p.Description,
		Image:       p.Image,
	}
}

type ProductStockRequest struct {
	DateAdded *time.Time `json:"date_added"`
	Type      *string    `json:"type"`
	Stock     *float64   `json:"stock"`
}

func (p *ProductStockRequest) ToCommand() aProduct.ProductStockCommand {
	var movementType *dProduct.StockMovementType
	if p.Type != nil {
		mt := dProduct.StockMovementType(*p.Type)
		movementType = &mt
	}

	return aProduct.ProductStockCommand{
		DateAdded: p.DateAdded,
		Type:      movementType,
		Stock:     p.Stock,
	}
}

type ProductPriceRequest struct {
	DateAdded *time.Time `json:"date_added"`
	Price     *float64   `json:"price"`
}

func (p *ProductPriceRequest) ToCommand() aProduct.ProductPriceCommand {
	return aProduct.ProductPriceCommand{
		DateAdded: p.DateAdded,
		Price:     p.Price,
	}
}
