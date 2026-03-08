package transaction

import dTransaction "agropal/internal/domain/transaction"

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
