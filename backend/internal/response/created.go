package response

import "github.com/google/uuid"

// What a create endpoint answers in the envelope's data, so a client can go straight to the new
// record instead of listing and searching for it.

type FarmCreatedResponse struct {
	FarmID uuid.UUID `json:"farm_id"`
}

type LocationCreatedResponse struct {
	LocationID uuid.UUID `json:"location_id"`
}

type ProductCreatedResponse struct {
	ProductID uuid.UUID `json:"product_id"`
}

type PriceCreatedResponse struct {
	PriceID uuid.UUID `json:"price_id"`
}

type StockCreatedResponse struct {
	StockID uuid.UUID `json:"stock_id"`
}

// TransactionsCreatedResponse lists the new transactions in the order they were sent
type TransactionsCreatedResponse struct {
	TransactionIDs []uuid.UUID `json:"transaction_ids"`
}

type PostCreatedResponse struct {
	PostID uuid.UUID `json:"post_id"`
	Slug   string    `json:"slug"`
}

type CommentCreatedResponse struct {
	CommentID uuid.UUID `json:"comment_id"`
}
