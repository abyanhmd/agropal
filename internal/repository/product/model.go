package product

import (
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
	Type      string    `db:"type"`
	Quantity  float64   `db:"quantity"`
}

type ProductPriceModel struct {
	ID        uuid.UUID `db:"price_id"`
	ProductID uuid.UUID `db:"product_id"`
	CreatedAt time.Time `db:"created_at"`
	Price     float64   `db:"price"`
}
