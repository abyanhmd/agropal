package transaction

import (
	dTransaction "agropal/internal/domain/transaction"
	"agropal/internal/response"
)

func DomainToTransactionResponse(t dTransaction.Transaction) response.TransactionResponse {
	return response.TransactionResponse{
		TransactionID: t.ID,
		ProductID:     t.ProductID,
		Date:          t.Date,
		Type:          string(t.Type),
		Amount:        t.Amount,
		Description:   t.Description,
	}
}
