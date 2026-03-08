package transaction

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID          uuid.UUID
	Date        time.Time
	Type        TransactionType
	Amount      float64
	Currency    string
	Description string
	ProductID   uuid.UUID
}

type TransactionType string

const (
	TypeIncome  TransactionType = "INCOME"
	TypeExpense TransactionType = "EXPENSE"
)
