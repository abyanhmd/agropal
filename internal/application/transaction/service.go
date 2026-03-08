package transaction

import (
	dTransaction "agropal/internal/domain/transaction"
	"context"

	"github.com/google/uuid"
)

type Service struct {
	TransactionRepository dTransaction.Repository
}

func NewService(transactionRepository dTransaction.Repository) Service {
	return Service{
		TransactionRepository: transactionRepository,
	}
}

func (s *Service) Create(ctx context.Context, cmd TransactionCommand) error {
	transactionID, err := uuid.NewV7()
	if err != nil {
		return err
	}
	transaction := cmd.ToDomain(transactionID)

	// TODO: Validate create input

	return s.TransactionRepository.Create(ctx, transaction)
}

func (s *Service) GetOne(ctx context.Context, transactionID uuid.UUID) (dTransaction.Transaction, error) {
	return s.TransactionRepository.FindByID(ctx, transactionID)
}

func (s *Service) GetMany(ctx context.Context, farmID uuid.UUID) ([]dTransaction.Transaction, error) {
	return s.TransactionRepository.ListByFarmID(ctx, farmID)
}

func (s *Service) Update(ctx context.Context, transactionID uuid.UUID, cmd TransactionCommand) error {
	transaction := cmd.ToDomain(transactionID)

	// TODO: Validate update input

	return s.TransactionRepository.Update(ctx, transaction)
}

func (s *Service) Delete(ctx context.Context, transactionID uuid.UUID) error {
	return s.TransactionRepository.Delete(ctx, transactionID)
}
