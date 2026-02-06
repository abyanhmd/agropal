package transaction

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, transaction Transaction) error
	GetOne(ctx context.Context, transactionID uuid.UUID) (Transaction, error)
	GetMany(ctx context.Context, farmID uuid.UUID) ([]Transaction, error)
	Update(ctx context.Context, transaction Transaction) error
	Delete(ctx context.Context, transactionID uuid.UUID) error
}
