package product

import (
	"time"

	"github.com/google/uuid"
)

type ModelProduct struct {
	ProductID   uuid.UUID `db:"product_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	Image       string    `db:"image"`
	FarmID      uuid.UUID `db:"farm_id"`
}

type ModelProductStock struct {
	StockID   uuid.UUID `db:"stock_id"`
	ProductID uuid.UUID `db:"product_id"`
	DateAdded time.Time `db:"date_added"`
	Type      string    `db:"type"`
	Stock     float64   `db:"stock"`
}

type ModelProductPrice struct {
	PriceID   uuid.UUID `db:"price_id"`
	ProductID uuid.UUID `db:"product_id"`
	DateAdded time.Time `db:"date_added"`
	Price     float64   `db:"price"`
}
