package request

import (
	aTransaction "agropal/internal/app/transaction"
	dTransaction "agropal/internal/domain/transaction"
	"time"
)

type TransactionRequest struct {
	Date        time.Time `json:"date"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
}

func (r *TransactionRequest) ToCommand() aTransaction.TransactionCommand {
	return aTransaction.TransactionCommand{
		Date:        r.Date,
		Type:        dTransaction.TransactionType(r.Type),
		Amount:      r.Amount,
		Description: r.Description,
	}
}

type CreateTransactionsRequest struct {
	Transactions []TransactionRequest `json:"transactions"`
}

func (r *CreateTransactionsRequest) ToCommands() []aTransaction.TransactionCommand {
	cmds := make([]aTransaction.TransactionCommand, len(r.Transactions))
	for i, t := range r.Transactions {
		cmds[i] = t.ToCommand()
	}
	return cmds
}

type DeleteTransactionsRequest struct {
	TransactionIDs []string `json:"transaction_ids"`
}
