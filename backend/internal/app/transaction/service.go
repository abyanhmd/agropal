package transaction

import (
	"agropal/config"
	dFarm "agropal/internal/domain/farm"
	dProduct "agropal/internal/domain/product"
	"agropal/internal/domain/shared"
	dTransaction "agropal/internal/domain/transaction"
	dUser "agropal/internal/domain/user"
	"agropal/internal/infra/database"
	"context"

	"github.com/google/uuid"
)

type Service struct {
	transactor      database.Transactor
	userRepo        dUser.Repository
	farmRepo        dFarm.Repository
	productRepo     dProduct.Repository
	transactionRepo dTransaction.Repository
}

func NewService(
	t database.Transactor,
	ur dUser.Repository,
	fr dFarm.Repository,
	pr dProduct.Repository,
	tr dTransaction.Repository,
) *Service {
	return &Service{
		transactor:      t,
		userRepo:        ur,
		farmRepo:        fr,
		productRepo:     pr,
		transactionRepo: tr,
	}
}

func (s *Service) Create(ctx context.Context, farmID, productID uuid.UUID, cmd []TransactionCommand) error {
	if err := s.verifyProduct(ctx, farmID, productID); err != nil {
		return err
	}

	transactions := make([]dTransaction.Transaction, len(cmd))
	for i, c := range cmd {
		transactionID, err := uuid.NewV7()
		if err != nil {
			return err
		}
		transactions[i] = c.ToDomain(productID, transactionID)
	}

	return s.transactionRepo.CreateTransaction(ctx, s.transactor.DB(), transactions)
}

func (s *Service) GetOne(ctx context.Context, farmID, productID, transactionID uuid.UUID) (dTransaction.Transaction, error) {
	if err := s.verifyProduct(ctx, farmID, productID); err != nil {
		return dTransaction.Transaction{}, err
	}
	transaction, err := s.transactionRepo.FindTransactionByID(ctx, s.transactor.DB(), transactionID)
	if err != nil {
		return dTransaction.Transaction{}, err
	}
	if transaction == (dTransaction.Transaction{}) || transaction.ProductID != productID {
		return dTransaction.Transaction{}, dTransaction.ErrInvalidID
	}
	return transaction, nil
}

func (s *Service) ListByFarmID(ctx context.Context, farmID uuid.UUID) ([]dTransaction.Transaction, error) {
	if err := s.verifyProduct(ctx, farmID, uuid.Nil); err != nil {
		return []dTransaction.Transaction{}, err
	}
	return s.transactionRepo.ListTransactionsByFarmID(ctx, s.transactor.DB(), farmID)
}

func (s *Service) ListByProductID(ctx context.Context, farmID, productID uuid.UUID) ([]dTransaction.Transaction, error) {
	if err := s.verifyProduct(ctx, farmID, productID); err != nil {
		return []dTransaction.Transaction{}, err
	}
	return s.transactionRepo.ListTransactionsByProductID(ctx, s.transactor.DB(), productID)
}

func (s *Service) Update(ctx context.Context, farmID, productID, transactionID uuid.UUID, cmd TransactionCommand) error {
	if err := s.verifyProduct(ctx, farmID, productID); err != nil {
		return err
	}
	exists, err := s.transactionRepo.FindTransactionByID(ctx, s.transactor.DB(), transactionID)
	if err != nil {
		return err
	}
	if exists == (dTransaction.Transaction{}) || exists.ProductID != productID {
		return dTransaction.ErrInvalidID
	}
	transaction := cmd.ToDomain(productID, transactionID)

	return s.transactionRepo.UpdateTransaction(ctx, s.transactor.DB(), transaction)
}

func (s *Service) Delete(ctx context.Context, farmID, productID uuid.UUID, transactionIDs []uuid.UUID) error {
	if err := s.verifyProduct(ctx, farmID, productID); err != nil {
		return err
	}

	return s.transactionRepo.DeleteTransactions(ctx, s.transactor.DB(), productID, transactionIDs)
}

// DeleteMany deletes an arbitrary set of transactions, verifying that the
// requester owns (or administers) the farm behind each transaction's product
// before removing it. Deletes are grouped per product to match the
// repository's per-product delete contract.
func (s *Service) DeleteMany(ctx context.Context, transactionIDs []uuid.UUID) error {
	db := s.transactor.DB()
	byProduct := make(map[uuid.UUID][]uuid.UUID)
	verifiedProducts := make(map[uuid.UUID]bool)

	for _, transactionID := range transactionIDs {
		transaction, err := s.transactionRepo.FindTransactionByID(ctx, db, transactionID)
		if err != nil {
			return err
		}
		if transaction == (dTransaction.Transaction{}) {
			return dTransaction.ErrInvalidID
		}

		if !verifiedProducts[transaction.ProductID] {
			product, err := s.productRepo.FindProductByID(ctx, db, transaction.ProductID)
			if err != nil {
				return err
			}
			if err = s.verifyProduct(ctx, product.FarmID, product.ID); err != nil {
				return err
			}
			verifiedProducts[transaction.ProductID] = true
		}

		byProduct[transaction.ProductID] = append(byProduct[transaction.ProductID], transactionID)
	}

	for productID, ids := range byProduct {
		if err := s.transactionRepo.DeleteTransactions(ctx, db, productID, ids); err != nil {
			return err
		}
	}

	return nil
}

// DeleteAll deletes every transaction across every farm owned by the
// requesting user.
func (s *Service) DeleteAll(ctx context.Context) error {
	userID, err := shared.ExtractUserID(ctx)
	if err != nil {
		return err
	}

	db := s.transactor.DB()

	farms, err := s.farmRepo.ListFarmByUserID(ctx, db, userID)
	if err != nil {
		return err
	}

	for _, farm := range farms {
		transactions, err := s.transactionRepo.ListTransactionsByFarmID(ctx, db, farm.ID)
		if err != nil {
			return err
		}

		byProduct := make(map[uuid.UUID][]uuid.UUID)
		for _, transaction := range transactions {
			byProduct[transaction.ProductID] = append(byProduct[transaction.ProductID], transaction.ID)
		}

		for productID, ids := range byProduct {
			if err = s.transactionRepo.DeleteTransactions(ctx, db, productID, ids); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Service) verifyProduct(ctx context.Context, farmID, productID uuid.UUID) error {
	return s.transactor.WithTx(ctx, func(q database.Querier) error {
		userID, err := shared.ExtractUserID(ctx)
		if err != nil {
			return err
		}
		user, err := s.userRepo.FindUserByID(ctx, q, userID)
		if err != nil {
			return err
		}
		if user == (dUser.User{}) {
			return dUser.ErrInvalidUserID
		}
		farm, err := s.farmRepo.FindFarmByID(ctx, q, farmID)
		if err != nil {
			return err
		}
		requesterRole := shared.GetMetadata(ctx, shared.KeyRole)
		if farm.UserID != user.ID && requesterRole != string(config.RoleAdmin) {
			return shared.ErrForbidden
		}
		if productID == uuid.Nil {
			return nil
		}
		product, err := s.productRepo.FindProductByID(ctx, q, productID)
		if err != nil {
			return err
		}
		if product == (dProduct.Product{}) {
			return dProduct.ErrInvalidID
		}
		if product.FarmID != farmID {
			return shared.ErrForbidden
		}
		return nil
	})
}
