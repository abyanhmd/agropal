package transaction

import (
	"agropal/internal/database"
	"agropal/internal/domain/shared"
	dTransaction "agropal/internal/domain/transaction"
	"agropal/internal/utils"
	"context"

	"github.com/google/uuid"
)

type RepositoryImpl struct {
	Logger *utils.Logger
}

func NewRepository(logger *utils.Logger) dTransaction.Repository {
	return &RepositoryImpl{Logger: logger}
}

func (r *RepositoryImpl) Create(ctx context.Context, q database.Querier, t dTransaction.Transaction) error {
	query := `
		INSERT INTO transactions (
			transaction_id, date, type,
			amount, currency, description,
			product_id
		) VALUES (
			:transaction_id, :date, :type,
			:amount, :currency, :description,
			:product_id
		);`

	transactionModel := ToTransactionModel(t)
	if _, err := q.NamedExecContext(ctx, query, transactionModel); err != nil {
		r.Logger.Error(ctx, shared.AssetTransaction, err.Error())
		return err
	}

	return nil
}

func (r *RepositoryImpl) FindByID(ctx context.Context, q database.Querier, id uuid.UUID) (dTransaction.Transaction, error) {
	getQuery := `SELECT * FROM transactions WHERE transaction_id = $1;`

	var transactionModel TransactionModel
	if err := q.GetContext(ctx, &transactionModel, getQuery, id); err != nil {
		r.Logger.Error(ctx, shared.AssetTransaction, err.Error())
		return dTransaction.Transaction{}, err
	}

	return transactionModel.ToDomain(), nil
}

func (r *RepositoryImpl) ListByFarmID(ctx context.Context, q database.Querier, farmID uuid.UUID) ([]dTransaction.Transaction, error) {
	query := `SELECT * FROM transactions WHERE farm_id = $1;`

	var transactionModels []TransactionModel
	if err := q.SelectContext(ctx, &transactionModels, query, farmID); err != nil {
		r.Logger.Error(ctx, shared.AssetTransaction, err.Error())
		return nil, err
	}

	transactions := make([]dTransaction.Transaction, len(transactionModels))
	for i, m := range transactionModels {
		transactions[i] = m.ToDomain()
	}

	return transactions, nil
}

func (r *RepositoryImpl) Update(ctx context.Context, q database.Querier, t dTransaction.Transaction) error {
	query := `UPDATE transactions 
		SET date = :date, 
		    type = :type, 
		    amount = :amount, 
		    currency = :currency,
		    description = :description, 
		    product_id = :product_id 
		WHERE transaction_id = :transaction_id;`

	transactionModel := ToTransactionModel(t)
	if _, err := q.NamedExecContext(ctx, query, transactionModel); err != nil {
		r.Logger.Error(ctx, shared.AssetTransaction, err.Error())
		return err
	}

	return nil
}

func (r *RepositoryImpl) Delete(ctx context.Context, q database.Querier, id uuid.UUID) error {
	query := `DELETE FROM transactions WHERE transaction_id = $1;`

	if _, err := q.ExecContext(ctx, query, id); err != nil {
		r.Logger.Error(ctx, shared.AssetTransaction, err.Error())
		return err
	}

	return nil
}
