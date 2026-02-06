package transaction

import dTransaction "agropal/internal/domain/transaction"

func (m ModelTransaction) ToDomain() dTransaction.Transaction {
	return dTransaction.Transaction{
		TransactionID: m.TransactionID,
		Date:          m.Date,
		Type:          dTransaction.TransactionType(m.Type),
		Amount:        m.Amount,
		Description:   m.Description,
		ProductID:     m.ProductID,
	}
}

func DomainToModelTransaction(domain dTransaction.Transaction) ModelTransaction {
	return ModelTransaction{
		TransactionID: domain.TransactionID,
		Date:          domain.Date,
		Type:          string(domain.Type),
		Amount:        domain.Amount,
		Description:   domain.Description,
		ProductID:     domain.ProductID,
	}
}
