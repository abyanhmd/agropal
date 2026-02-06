package transaction

import (
	"agropal/internal/domain/shared"
	dTransaction "agropal/internal/domain/transaction"
	"agropal/internal/utils"
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type RepositoryImpl struct {
	DB     *sqlx.DB
	Logger *utils.Logger
}

func (r *RepositoryImpl) Create(ctx context.Context, transaction dTransaction.Transaction) error {
	insertQuery := `INSERT INTO transactions (
		transaction_id, date, type,
		amount, description, product_id
	) VALUES (
		:transaction_id, :date, :type,
		:amount, :description, :product_id
	);`

	modelTransaction := DomainToModelTransaction(transaction)
	if _, errInsert := r.DB.NamedExecContext(ctx, insertQuery, modelTransaction); errInsert != nil {
		r.Logger.Error(ctx, shared.AssetTransaction, errInsert.Error())
		return errInsert
	}

	return nil
}

func (r *RepositoryImpl) GetOne(ctx context.Context, transactionID uuid.UUID) (dTransaction.Transaction, error) {
	getQuery := `SELECT * FROM transactions WHERE transaction_id = $1;`

	var modelTransaction ModelTransaction
	if errGet := r.DB.GetContext(ctx, &modelTransaction, getQuery, transactionID); errGet != nil {
		r.Logger.Error(ctx, shared.AssetTransaction, errGet.Error())
		return dTransaction.Transaction{}, errGet
	}

	return modelTransaction.ToDomain(), nil
}

func (r *RepositoryImpl) GetMany(ctx context.Context, farmID uuid.UUID) ([]dTransaction.Transaction, error) {
	getQuery := `SELECT * FROM transactions WHERE farm_id = $1;`

	var modelTransactions []ModelTransaction
	if errSelect := r.DB.SelectContext(ctx, &modelTransactions, getQuery, farmID); errSelect != nil {
		r.Logger.Error(ctx, shared.AssetTransaction, errSelect.Error())
		return nil, errSelect
	}

	transactions := make([]dTransaction.Transaction, len(modelTransactions))
	for i, m := range modelTransactions {
		transactions[i] = m.ToDomain()
	}

	return transactions, nil
}

func (r *RepositoryImpl) Update(ctx context.Context, transaction dTransaction.Transaction) error {
	updateQuery := `UPDATE transactions 
		SET date = :date, 
		    type = :type, 
		    amount = :amount, 
		    description = :description, 
		    product_id = :product_id 
		WHERE transaction_id = :transaction_id;`

	modelTransaction := DomainToModelTransaction(transaction)
	if _, errUpdate := r.DB.NamedExecContext(ctx, updateQuery, modelTransaction); errUpdate != nil {
		r.Logger.Error(ctx, shared.AssetTransaction, errUpdate.Error())
		return errUpdate
	}

	return nil
}

func (r *RepositoryImpl) Delete(ctx context.Context, transactionID uuid.UUID) error {
	deleteQuery := `DELETE FROM transactions WHERE transaction_id = $1;`

	if _, errDelete := r.DB.ExecContext(ctx, deleteQuery, transactionID); errDelete != nil {
		r.Logger.Error(ctx, shared.AssetTransaction, errDelete.Error())
		return errDelete
	}

	return nil
}

func NewRepository(db *sqlx.DB, logger *utils.Logger) dTransaction.Repository {
	return &RepositoryImpl{
		DB:     db,
		Logger: logger,
	}
}
