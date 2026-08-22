package persistence

import (
	"agropal/internal/domain/shared"
	dTransaction "agropal/internal/domain/transaction"
	"agropal/internal/infra/database"
	"agropal/pkg"
	"context"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type transactionRepo struct {
	logger *pkg.Logger
}

func NewTransactionRepository(l *pkg.Logger) dTransaction.Repository {
	return &transactionRepo{logger: l}
}

func (r *transactionRepo) CreateTransaction(ctx context.Context, q database.Querier, t []dTransaction.Transaction) error {
	query := `
		INSERT INTO transactions (
			transaction_id, transaction_date, transaction_type,
			amount, currency, description,
			product_id
		) VALUES (
			:transaction_id, :transaction_date, :transaction_type,
			:amount, :currency, :description,
			:product_id
		);`

	transactionModels := make([]TransactionModel, len(t))
	for i, model := range t {
		transactionModels[i] = ToTransactionModel(model)
	}

	if _, err := q.NamedExecContext(ctx, query, transactionModels); err != nil {
		r.logger.Error(ctx, shared.AssetTransaction, err.Error())
		return err
	}

	return nil
}

func (r *transactionRepo) FindTransactionByID(ctx context.Context, q database.Querier, id uuid.UUID) (dTransaction.Transaction, error) {
	getQuery := `SELECT * FROM transactions WHERE transaction_id = $1;`

	var transactionModel TransactionModel
	if err := q.GetContext(ctx, &transactionModel, getQuery, id); err != nil {
		r.logger.Error(ctx, shared.AssetTransaction, err.Error())
		return dTransaction.Transaction{}, err
	}

	return transactionModel.ToDomain(), nil
}

// ListTransactionsByFarmID joins through products since transactions only
// reference product_id directly - there is no farm_id column on
// transactions itself.
func (r *transactionRepo) ListTransactionsByFarmID(ctx context.Context, q database.Querier, farmID uuid.UUID) ([]dTransaction.Transaction, error) {
	query := `
		SELECT t.* FROM transactions t
		JOIN products p ON p.product_id = t.product_id
		WHERE p.farm_id = $1;`

	var transactionModels []TransactionModel
	if err := q.SelectContext(ctx, &transactionModels, query, farmID); err != nil {
		r.logger.Error(ctx, shared.AssetTransaction, err.Error())
		return nil, err
	}

	transactions := make([]dTransaction.Transaction, len(transactionModels))
	for i, m := range transactionModels {
		transactions[i] = m.ToDomain()
	}

	return transactions, nil
}

func (r *transactionRepo) ListTransactionsByProductID(ctx context.Context, q database.Querier, productID uuid.UUID) ([]dTransaction.Transaction, error) {
	query := `SELECT * FROM transactions WHERE product_id = $1;`

	var transactionModels []TransactionModel
	if err := q.SelectContext(ctx, &transactionModels, query, productID); err != nil {
		r.logger.Error(ctx, shared.AssetTransaction, err.Error())
		return nil, err
	}

	transactions := make([]dTransaction.Transaction, len(transactionModels))
	for i, m := range transactionModels {
		transactions[i] = m.ToDomain()
	}

	return transactions, nil
}

func (r *transactionRepo) UpdateTransaction(ctx context.Context, q database.Querier, t dTransaction.Transaction) error {
	query := `UPDATE transactions
		SET transaction_date = :transaction_date,
		    transaction_type = :transaction_type,
		    amount = :amount,
		    currency = :currency,
		    description = :description,
		    product_id = :product_id
		WHERE transaction_id = :transaction_id;`

	transactionModel := ToTransactionModel(t)
	if _, err := q.NamedExecContext(ctx, query, transactionModel); err != nil {
		r.logger.Error(ctx, shared.AssetTransaction, err.Error())
		return err
	}

	return nil
}

func (r *transactionRepo) DeleteTransactions(ctx context.Context, q database.Querier, productID uuid.UUID, transactionIDs []uuid.UUID) error {
	query := `DELETE FROM transactions WHERE product_id = $1 AND transaction_id = ANY($2);`

	if _, err := q.ExecContext(ctx, query, productID, pq.Array(transactionIDs)); err != nil {
		r.logger.Error(ctx, shared.AssetTransaction, err.Error())
		return err
	}

	return nil
}
