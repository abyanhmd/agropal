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
	Name        *string `json:"name" validate:"required"`
	Description *string `json:"description" validate:"required"`
	Image       *string `json:"image" validate:"required"`
}

func (p *ProductRequest) ToCommand() aProduct.ProductCommand {
	return aProduct.ProductCommand{
		Name:        p.Name,
		Description: p.Description,
		Image:       p.Image,
	}
}

type ProductStockRequest struct {
	DateAdded *time.Time `json:"date_added" validate:"required"`
	Type      *string    `json:"type" validate:"required,oneof=IN OUT"`
	Stock     *float64   `json:"stock" validate:"required,gt=0"`
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
	DateAdded *time.Time `json:"date_added" validate:"required"`
	Price     *float64   `json:"price" validate:"required,gt=0"`
}

func (p *ProductPriceRequest) ToCommand() aProduct.ProductPriceCommand {
	return aProduct.ProductPriceCommand{
		DateAdded: p.DateAdded,
		Price:     p.Price,
	}
}
