package transaction

import (
	dTransaction "agropal/internal/domain/transaction"
	"time"

	"github.com/google/uuid"
)

type TransactionCommand struct {
	Date        time.Time
	Type        dTransaction.TransactionType
	Amount      float64
	Currency    string
	Description string
}

func (c TransactionCommand) ToDomain(productID, transactionID uuid.UUID) dTransaction.Transaction {
	return dTransaction.Transaction{
		ID:          transactionID,
		Date:        c.Date,
		Type:        c.Type,
		Amount:      c.Amount,
		Currency:    c.Currency,
		Description: c.Description,
		ProductID:   productID,
	}
}
