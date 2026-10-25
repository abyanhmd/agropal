package product

import (
	"agropal/internal/domain/shared"
	"time"

	"github.com/google/uuid"
)

type ProductDetail struct {
	Product Product
	Stocks  []ProductStock
	Prices  []ProductPrice
}

type Product struct {
	ID          uuid.UUID
	Name        string
	Description string
	Image       string
	FarmID      uuid.UUID
}

type ProductStock struct {
	ID        uuid.UUID
	ProductID uuid.UUID
	DateAdded time.Time
	Type      StockMovementType
	Stock     float64
}

type ProductPrice struct {
	ID        uuid.UUID
	ProductID uuid.UUID
	DateAdded time.Time
	Price     float64
}

type StockMovementType string

const (
	MovementIn  StockMovementType = "IN"
	MovementOut StockMovementType = "OUT"
)

var (
	SuccessCreated shared.SuccessKey = "PRODUCT.CREATE_SUCCESS"
	SuccessUpdated shared.SuccessKey = "PRODUCT.UPDATE_SUCCESS"
	SuccessDeleted shared.SuccessKey = "PRODUCT.DELETE_SUCCESS"

	SuccessStockAdded   shared.SuccessKey = "PRODUCT.STOCK_ADD_SUCCESS"
	SuccessStockDeleted shared.SuccessKey = "PRODUCT.STOCK_DELETE_SUCCESS"

	SuccessPriceAdded   shared.SuccessKey = "PRODUCT.PRICE_ADD_SUCCESS"
	SuccessPriceDeleted shared.SuccessKey = "PRODUCT.PRICE_DELETE_SUCCESS"
)
