package persistence

import (
	dTransaction "agropal/internal/domain/transaction"
	"time"

	"github.com/google/uuid"
)

type TransactionModel struct {
	ID          uuid.UUID `db:"transaction_id"`
	Date        time.Time `db:"transaction_date"`
	Type        string    `db:"transaction_type"`
	Amount      float64   `db:"amount"`
	Currency    string    `db:"currency"`
	Description string    `db:"description"`
	ProductID   uuid.UUID `db:"product_id"`
}

func ToTransactionModel(domain dTransaction.Transaction) TransactionModel {
	return TransactionModel{
		ID:          domain.ID,
		Date:        domain.Date,
		Type:        string(domain.Type),
		Amount:      domain.Amount,
		Currency:    domain.Currency,
		Description: domain.Description,
		ProductID:   domain.ProductID,
	}
}

func (m TransactionModel) ToDomain() dTransaction.Transaction {
	return dTransaction.Transaction{
		ID:          m.ID,
		Date:        m.Date,
		Type:        dTransaction.TransactionType(m.Type),
		Amount:      m.Amount,
		Currency:    m.Currency,
		Description: m.Description,
		ProductID:   m.ProductID,
	}
}
