package transaction

import (
	"time"

	"github.com/google/uuid"
)

type ModelTransaction struct {
	TransactionID uuid.UUID `db:"transaction_id"`
	Date          time.Time `db:"date"`
	Type          string    `db:"type"`
	Amount        float64   `db:"amount"`
	Description   string    `db:"description"`
	ProductID     uuid.UUID `db:"product_id"`
}
