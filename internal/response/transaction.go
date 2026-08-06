package response

import (
	"time"

	"github.com/google/uuid"
)

type TransactionResponse struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	ProductID     uuid.UUID `json:"product_id"`
	Date          time.Time `json:"date"`
	Type          string    `json:"type"`
	Amount        float64   `json:"amount"`
	Description   string    `json:"description"`
}
