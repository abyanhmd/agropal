package transaction

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	TransactionID uuid.UUID
	Date          time.Time
	Type          TransactionType
	Amount        float64
	Description   string
	ProductID     uuid.UUID
}

type TransactionType string

const (
	TransactionTypeIncome  TransactionType = "income"
	TransactionTypeExpense TransactionType = "expense"
)
