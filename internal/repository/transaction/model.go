package transaction

import (
	"time"

	"github.com/google/uuid"
)

type TransactionModel struct {
	ID          uuid.UUID `db:"transaction_id"`
	Date        time.Time `db:"date"`
	Type        string    `db:"type"`
	Amount      float64   `db:"amount"`
	Currency    string    `db:"currency"`
	Description string    `db:"description"`
	ProductID   uuid.UUID `db:"product_id"`
}
