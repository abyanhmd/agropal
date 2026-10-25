package transaction

import (
	"agropal/internal/infra/database"
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreateTransaction(ctx context.Context, q database.Querier, t []Transaction) error
	FindTransactionByID(ctx context.Context, q database.Querier, id uuid.UUID) (Transaction, error)
	ListTransactionsByFarmID(ctx context.Context, q database.Querier, farmID uuid.UUID) ([]Transaction, error)
	ListTransactionsByProductID(ctx context.Context, q database.Querier, productID uuid.UUID) ([]Transaction, error)
	UpdateTransaction(ctx context.Context, q database.Querier, t Transaction) error
	DeleteTransactions(ctx context.Context, q database.Querier, productID uuid.UUID, transactionIDs []uuid.UUID) error
}
