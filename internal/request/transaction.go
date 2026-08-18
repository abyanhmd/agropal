package request

import (
	aTransaction "agropal/internal/app/transaction"
	dTransaction "agropal/internal/domain/transaction"
	"time"
)

type TransactionRequest struct {
	Date        time.Time `json:"date" validate:"required"`
	Type        string    `json:"type" validate:"required,oneof=INCOME EXPENSE"`
	Amount      float64   `json:"amount" validate:"required,gt=0"`
	Description string    `json:"description" validate:"max=255"`
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
	Transactions []TransactionRequest `json:"transactions" validate:"required,min=1,dive"`
}

func (r *CreateTransactionsRequest) ToCommands() []aTransaction.TransactionCommand {
	cmds := make([]aTransaction.TransactionCommand, len(r.Transactions))
	for i, t := range r.Transactions {
		cmds[i] = t.ToCommand()
	}
	return cmds
}

type DeleteTransactionsRequest struct {
	TransactionIDs []string `json:"transaction_ids" validate:"required,min=1,dive,uuid"`
}
