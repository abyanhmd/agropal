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
	Description string
	ProductID   uuid.UUID
}

func (c TransactionCommand) ToDomain(transactionID uuid.UUID) dTransaction.Transaction {
	return dTransaction.Transaction{
		TransactionID: transactionID,
		Date:          c.Date,
		Type:          c.Type,
		Amount:        c.Amount,
		Description:   c.Description,
		ProductID:     c.ProductID,
	}
}
