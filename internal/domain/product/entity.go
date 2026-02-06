package product

import (
	"time"

	"github.com/google/uuid"
)

type ProductDetail struct {
	Product Product
	Stocks  []ProductStock
	Prices  []ProductPrice
}

type Product struct {
	ProductID   uuid.UUID
	Name        string
	Description string
	Image       string
	FarmID      uuid.UUID
}

type ProductStock struct {
	StockID   uuid.UUID
	ProductID uuid.UUID
	DateAdded time.Time
	Type      StockType
	Stock     float64
}

type ProductPrice struct {
	PriceID   uuid.UUID
	ProductID uuid.UUID
	DateAdded time.Time
	Price     float64
}

type StockType string

const (
	StockTypeIn  StockType = "in"
	StockTypeOut StockType = "out"
)
