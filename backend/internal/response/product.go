package response

import (
	"time"

	"github.com/google/uuid"
)

type ProductResponse struct {
	ProductID   uuid.UUID `json:"product_id"`
	FarmID      uuid.UUID `json:"farm_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Image       string    `json:"image"`
	Unit        string    `json:"unit"`
}

type ProductStockResponse struct {
	StockID   uuid.UUID `json:"stock_id"`
	ProductID uuid.UUID `json:"product_id"`
	DateAdded time.Time `json:"date_added"`
	Type      string    `json:"type"`
	Stock     float64   `json:"stock"`
}

type ProductPriceResponse struct {
	PriceID   uuid.UUID `json:"price_id"`
	ProductID uuid.UUID `json:"product_id"`
	DateAdded time.Time `json:"date_added"`
	Price     float64   `json:"price"`
	Currency  string    `json:"currency"`
}

type ProductDetailResponse struct {
	Product ProductResponse        `json:"product"`
	Stocks  []ProductStockResponse `json:"stocks"`
	Prices  []ProductPriceResponse `json:"prices"`
}
