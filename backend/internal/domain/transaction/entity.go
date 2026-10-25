package transaction

import (
	"agropal/internal/domain/shared"
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

var (
	SuccessCreated shared.SuccessKey = "TRANSACTION.CREATE_SUCCESS"
	SuccessUpdated shared.SuccessKey = "TRANSACTION.UPDATE_SUCCESS"
	SuccessDeleted shared.SuccessKey = "TRANSACTION.DELETE_SUCCESS"
)
