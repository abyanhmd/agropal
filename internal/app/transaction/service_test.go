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
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

/* ==================== Fakes ====================== */

// fakeTransactor runs WithTx's callback directly against a nil Querier
// instead of opening a real transaction. Safe here because the fake
// repositories below never touch the Querier they're given - they stand in
// for the whole repository, not just the SQL layer underneath it.
type fakeTransactor struct{}

func (fakeTransactor) DB() database.Querier {
	return nil
}

func (fakeTransactor) WithTx(_ context.Context, fn func(q database.Querier) error) error {
	return fn(nil)
}

// fakeUserRepository implements dUser.Repository. Only FindUserByID is
// exercised by transaction.Service.
type fakeUserRepository struct {
	findUserByIDFunc func(ctx context.Context, q database.Querier, userID uuid.UUID) (dUser.User, error)
}

func (f *fakeUserRepository) CreateUser(context.Context, database.Querier, dUser.User) error {
	panic("fakeUserRepository: CreateUser not stubbed")
}

func (f *fakeUserRepository) FindUserByID(ctx context.Context, q database.Querier, userID uuid.UUID) (dUser.User, error) {
	if f.findUserByIDFunc == nil {
		panic("fakeUserRepository: findUserByIDFunc not stubbed")
	}
	return f.findUserByIDFunc(ctx, q, userID)
}

func (f *fakeUserRepository) FindUserByEmail(context.Context, database.Querier, string) (dUser.User, error) {
	panic("fakeUserRepository: FindUserByEmail not stubbed")
}

func (f *fakeUserRepository) FindUserByUsername(context.Context, database.Querier, string) (dUser.User, error) {
	panic("fakeUserRepository: FindUserByUsername not stubbed")
}

func (f *fakeUserRepository) ListUsers(context.Context, database.Querier) ([]dUser.User, error) {
	panic("fakeUserRepository: ListUsers not stubbed")
}

func (f *fakeUserRepository) UpdateUser(context.Context, database.Querier, dUser.User) error {
	panic("fakeUserRepository: UpdateUser not stubbed")
}

func (f *fakeUserRepository) UpdateUserPassword(context.Context, database.Querier, uuid.UUID, string) error {
	panic("fakeUserRepository: UpdateUserPassword not stubbed")
}

func (f *fakeUserRepository) VerifyUserEmail(context.Context, database.Querier, uuid.UUID, time.Time) error {
	panic("fakeUserRepository: VerifyUserEmail not stubbed")
}

func (f *fakeUserRepository) SoftDeleteUser(context.Context, database.Querier, uuid.UUID, time.Time) error {
	panic("fakeUserRepository: SoftDeleteUser not stubbed")
}

func (f *fakeUserRepository) DeleteUser(context.Context, database.Querier, uuid.UUID) error {
	panic("fakeUserRepository: DeleteUser not stubbed")
}

func (f *fakeUserRepository) ListUsersByAsset(context.Context, database.Querier, shared.AssetType, uuid.UUID) ([]dUser.User, error) {
	panic("fakeUserRepository: ListUsersByAsset not stubbed")
}

// fakeFarmRepository implements dFarm.Repository. FindFarmByID and
// ListFarmByUserID are exercised by transaction.Service.
type fakeFarmRepository struct {
	findFarmByIDFunc     func(ctx context.Context, q database.Querier, farmID uuid.UUID) (dFarm.Farm, error)
	listFarmByUserIDFunc func(ctx context.Context, q database.Querier, userID uuid.UUID) ([]dFarm.Farm, error)
}

func (f *fakeFarmRepository) CreateFarm(context.Context, database.Querier, dFarm.Farm) error {
	panic("fakeFarmRepository: CreateFarm not stubbed")
}

func (f *fakeFarmRepository) FindFarmByID(ctx context.Context, q database.Querier, farmID uuid.UUID) (dFarm.Farm, error) {
	if f.findFarmByIDFunc == nil {
		panic("fakeFarmRepository: findFarmByIDFunc not stubbed")
	}
	return f.findFarmByIDFunc(ctx, q, farmID)
}

func (f *fakeFarmRepository) UpdateFarm(context.Context, database.Querier, dFarm.Farm) error {
	panic("fakeFarmRepository: UpdateFarm not stubbed")
}

func (f *fakeFarmRepository) DeleteFarm(context.Context, database.Querier, uuid.UUID) error {
	panic("fakeFarmRepository: DeleteFarm not stubbed")
}

func (f *fakeFarmRepository) ListFarmByUserID(ctx context.Context, q database.Querier, userID uuid.UUID) ([]dFarm.Farm, error) {
	if f.listFarmByUserIDFunc == nil {
		panic("fakeFarmRepository: listFarmByUserIDFunc not stubbed")
	}
	return f.listFarmByUserIDFunc(ctx, q, userID)
}

// fakeProductRepository implements dProduct.Repository. Only
// FindProductByID is exercised by transaction.Service.
type fakeProductRepository struct {
	findProductByIDFunc func(ctx context.Context, q database.Querier, id uuid.UUID) (dProduct.Product, error)
}

func (f *fakeProductRepository) CreateProduct(context.Context, database.Querier, dProduct.Product) error {
	panic("fakeProductRepository: CreateProduct not stubbed")
}

func (f *fakeProductRepository) UpdateProduct(context.Context, database.Querier, dProduct.Product) error {
	panic("fakeProductRepository: UpdateProduct not stubbed")
}

func (f *fakeProductRepository) DeleteProduct(context.Context, database.Querier, uuid.UUID) error {
	panic("fakeProductRepository: DeleteProduct not stubbed")
}

func (f *fakeProductRepository) FindProductByID(ctx context.Context, q database.Querier, id uuid.UUID) (dProduct.Product, error) {
	if f.findProductByIDFunc == nil {
		panic("fakeProductRepository: findProductByIDFunc not stubbed")
	}
	return f.findProductByIDFunc(ctx, q, id)
}

func (f *fakeProductRepository) FindProductDetailByID(context.Context, database.Querier, uuid.UUID) (dProduct.ProductDetail, error) {
	panic("fakeProductRepository: FindProductDetailByID not stubbed")
}

func (f *fakeProductRepository) ListProductsByFarmID(context.Context, database.Querier, uuid.UUID) ([]dProduct.ProductDetail, error) {
	panic("fakeProductRepository: ListProductsByFarmID not stubbed")
}

func (f *fakeProductRepository) AddProductStock(context.Context, database.Querier, dProduct.ProductStock) error {
	panic("fakeProductRepository: AddProductStock not stubbed")
}

func (f *fakeProductRepository) AddProductPrice(context.Context, database.Querier, dProduct.ProductPrice) error {
	panic("fakeProductRepository: AddProductPrice not stubbed")
}

func (f *fakeProductRepository) DeleteProductStock(context.Context, database.Querier, uuid.UUID) error {
	panic("fakeProductRepository: DeleteProductStock not stubbed")
}

func (f *fakeProductRepository) DeleteProductPrice(context.Context, database.Querier, uuid.UUID) error {
	panic("fakeProductRepository: DeleteProductPrice not stubbed")
}

// fakeTransactionRepository implements dTransaction.Repository.
type fakeTransactionRepository struct {
	createTransactionFunc         func(ctx context.Context, q database.Querier, t []dTransaction.Transaction) error
	findTransactionByIDFunc       func(ctx context.Context, q database.Querier, id uuid.UUID) (dTransaction.Transaction, error)
	listTransactionsByFarmIDFunc  func(ctx context.Context, q database.Querier, farmID uuid.UUID) ([]dTransaction.Transaction, error)
	listTransactionsByProductFunc func(ctx context.Context, q database.Querier, productID uuid.UUID) ([]dTransaction.Transaction, error)
	updateTransactionFunc         func(ctx context.Context, q database.Querier, t dTransaction.Transaction) error
	deleteTransactionsFunc        func(ctx context.Context, q database.Querier, productID uuid.UUID, transactionIDs []uuid.UUID) error
}

func (f *fakeTransactionRepository) CreateTransaction(ctx context.Context, q database.Querier, t []dTransaction.Transaction) error {
	if f.createTransactionFunc == nil {
		panic("fakeTransactionRepository: createTransactionFunc not stubbed")
	}
	return f.createTransactionFunc(ctx, q, t)
}

func (f *fakeTransactionRepository) FindTransactionByID(ctx context.Context, q database.Querier, id uuid.UUID) (dTransaction.Transaction, error) {
	if f.findTransactionByIDFunc == nil {
		panic("fakeTransactionRepository: findTransactionByIDFunc not stubbed")
	}
	return f.findTransactionByIDFunc(ctx, q, id)
}

func (f *fakeTransactionRepository) ListTransactionsByFarmID(ctx context.Context, q database.Querier, farmID uuid.UUID) ([]dTransaction.Transaction, error) {
	if f.listTransactionsByFarmIDFunc == nil {
		panic("fakeTransactionRepository: listTransactionsByFarmIDFunc not stubbed")
	}
	return f.listTransactionsByFarmIDFunc(ctx, q, farmID)
}

func (f *fakeTransactionRepository) ListTransactionsByProductID(ctx context.Context, q database.Querier, productID uuid.UUID) ([]dTransaction.Transaction, error) {
	if f.listTransactionsByProductFunc == nil {
		panic("fakeTransactionRepository: listTransactionsByProductFunc not stubbed")
	}
	return f.listTransactionsByProductFunc(ctx, q, productID)
}

func (f *fakeTransactionRepository) UpdateTransaction(ctx context.Context, q database.Querier, t dTransaction.Transaction) error {
	if f.updateTransactionFunc == nil {
		panic("fakeTransactionRepository: updateTransactionFunc not stubbed")
	}
	return f.updateTransactionFunc(ctx, q, t)
}

func (f *fakeTransactionRepository) DeleteTransactions(ctx context.Context, q database.Querier, productID uuid.UUID, transactionIDs []uuid.UUID) error {
	if f.deleteTransactionsFunc == nil {
		panic("fakeTransactionRepository: deleteTransactionsFunc not stubbed")
	}
	return f.deleteTransactionsFunc(ctx, q, productID, transactionIDs)
}

/* ==================== Helpers ====================== */

func ctxWithUser(userID uuid.UUID) context.Context {
	return shared.WithMetadata(context.Background(), shared.KeyUserID, userID.String())
}

func ctxWithUserAndRole(userID uuid.UUID, role config.Role) context.Context {
	ctx := ctxWithUser(userID)
	return shared.WithMetadata(ctx, shared.KeyRole, string(role))
}

func newVerifiedRepos(ownerID, farmID, productID uuid.UUID) (*fakeUserRepository, *fakeFarmRepository, *fakeProductRepository) {
	userRepo := &fakeUserRepository{
		findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
			return dUser.User{ID: ownerID}, nil
		},
	}
	farmRepo := &fakeFarmRepository{
		findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
			return dFarm.Farm{ID: farmID, UserID: ownerID}, nil
		},
	}
	productRepo := &fakeProductRepository{
		findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
			return dProduct.Product{ID: productID, FarmID: farmID}, nil
		},
	}
	return userRepo, farmRepo, productRepo
}

/* ==================== Tests: Create ==================== */

func TestService_Create(t *testing.T) {
	ownerID := uuid.New()
	farmID := uuid.New()
	productID := uuid.New()

	t.Run("generates a distinct ID per transaction in the batch", func(t *testing.T) {
		userRepo, farmRepo, productRepo := newVerifiedRepos(ownerID, farmID, productID)
		var created []dTransaction.Transaction
		transactionRepo := &fakeTransactionRepository{
			createTransactionFunc: func(_ context.Context, _ database.Querier, ts []dTransaction.Transaction) error {
				created = ts
				return nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, farmRepo, productRepo, transactionRepo)

		cmds := []TransactionCommand{
			{Type: dTransaction.TypeIncome, Amount: 10},
			{Type: dTransaction.TypeExpense, Amount: 20},
			{Type: dTransaction.TypeIncome, Amount: 30},
		}
		if err := svc.Create(ctxWithUser(ownerID), farmID, productID, cmds); err != nil {
			t.Fatalf("Create() error = %v, want nil", err)
		}

		if len(created) != 3 {
			t.Fatalf("got %d created transactions, want 3", len(created))
		}
		seen := make(map[uuid.UUID]bool, 3)
		for _, tx := range created {
			if tx.ID == uuid.Nil {
				t.Error("transaction has a nil ID")
			}
			if seen[tx.ID] {
				t.Errorf("transaction ID %v reused across the batch, want every transaction to get its own ID", tx.ID)
			}
			seen[tx.ID] = true
			if tx.ProductID != productID {
				t.Errorf("transaction.ProductID = %v, want %v", tx.ProductID, productID)
			}
		}
	})

	t.Run("forbids a non-owner", func(t *testing.T) {
		otherUserID := uuid.New()
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: otherUserID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return dFarm.Farm{ID: farmID, UserID: ownerID}, nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, farmRepo, &fakeProductRepository{}, &fakeTransactionRepository{})

		err := svc.Create(ctxWithUser(otherUserID), farmID, productID, []TransactionCommand{{Amount: 1}})
		if !errors.Is(err, shared.ErrForbidden) {
			t.Errorf("Create() error = %v, want %v", err, shared.ErrForbidden)
		}
	})
}

/* ==================== Tests: GetOne / ListByFarmID / ListByProductID ==================== */

func TestService_GetOne(t *testing.T) {
	ownerID := uuid.New()
	farmID := uuid.New()
	productID := uuid.New()
	transactionID := uuid.New()

	t.Run("returns a matching transaction", func(t *testing.T) {
		userRepo, farmRepo, productRepo := newVerifiedRepos(ownerID, farmID, productID)
		want := dTransaction.Transaction{ID: transactionID, ProductID: productID, Amount: 42}
		transactionRepo := &fakeTransactionRepository{
			findTransactionByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dTransaction.Transaction, error) {
				return want, nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, farmRepo, productRepo, transactionRepo)

		got, err := svc.GetOne(ctxWithUser(ownerID), farmID, productID, transactionID)
		if err != nil {
			t.Fatalf("GetOne() error = %v, want nil", err)
		}
		if got != want {
			t.Errorf("GetOne() = %+v, want %+v", got, want)
		}
	})

	t.Run("rejects a transaction belonging to a different product", func(t *testing.T) {
		userRepo, farmRepo, productRepo := newVerifiedRepos(ownerID, farmID, productID)
		otherProductID := uuid.New()
		transactionRepo := &fakeTransactionRepository{
			findTransactionByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dTransaction.Transaction, error) {
				return dTransaction.Transaction{ID: transactionID, ProductID: otherProductID}, nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, farmRepo, productRepo, transactionRepo)

		_, err := svc.GetOne(ctxWithUser(ownerID), farmID, productID, transactionID)
		if !errors.Is(err, dTransaction.ErrInvalidID) {
			t.Errorf("GetOne() error = %v, want %v", err, dTransaction.ErrInvalidID)
		}
	})
}

func TestService_ListByFarmID(t *testing.T) {
	ownerID := uuid.New()
	farmID := uuid.New()
	want := []dTransaction.Transaction{{ID: uuid.New()}, {ID: uuid.New()}}

	userRepo := &fakeUserRepository{
		findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
			return dUser.User{ID: ownerID}, nil
		},
	}
	farmRepo := &fakeFarmRepository{
		findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
			return dFarm.Farm{ID: farmID, UserID: ownerID}, nil
		},
	}
	// verifyProduct is called with productID=uuid.Nil here, so it must skip
	// the FindProductByID lookup entirely - a call would panic since
	// findProductByIDFunc is left unset.
	productRepo := &fakeProductRepository{}
	transactionRepo := &fakeTransactionRepository{
		listTransactionsByFarmIDFunc: func(_ context.Context, _ database.Querier, fid uuid.UUID) ([]dTransaction.Transaction, error) {
			if fid != farmID {
				t.Fatalf("ListTransactionsByFarmID called with %v, want %v", fid, farmID)
			}
			return want, nil
		},
	}
	svc := NewService(fakeTransactor{}, userRepo, farmRepo, productRepo, transactionRepo)

	got, err := svc.ListByFarmID(ctxWithUser(ownerID), farmID)
	if err != nil {
		t.Fatalf("ListByFarmID() error = %v, want nil", err)
	}
	if len(got) != len(want) {
		t.Fatalf("ListByFarmID() returned %d transactions, want %d", len(got), len(want))
	}
}

func TestService_ListByProductID(t *testing.T) {
	ownerID := uuid.New()
	farmID := uuid.New()
	productID := uuid.New()
	want := []dTransaction.Transaction{{ID: uuid.New(), ProductID: productID}}

	userRepo, farmRepo, productRepo := newVerifiedRepos(ownerID, farmID, productID)
	transactionRepo := &fakeTransactionRepository{
		listTransactionsByProductFunc: func(_ context.Context, _ database.Querier, pid uuid.UUID) ([]dTransaction.Transaction, error) {
			if pid != productID {
				t.Fatalf("ListTransactionsByProductID called with %v, want %v", pid, productID)
			}
			return want, nil
		},
	}
	svc := NewService(fakeTransactor{}, userRepo, farmRepo, productRepo, transactionRepo)

	got, err := svc.ListByProductID(ctxWithUser(ownerID), farmID, productID)
	if err != nil {
		t.Fatalf("ListByProductID() error = %v, want nil", err)
	}
	if len(got) != len(want) {
		t.Fatalf("ListByProductID() returned %d transactions, want %d", len(got), len(want))
	}
}

/* ==================== Tests: Update / Delete / DeleteMany / DeleteAll ==================== */

func TestService_Update(t *testing.T) {
	ownerID := uuid.New()
	farmID := uuid.New()
	productID := uuid.New()
	transactionID := uuid.New()

	t.Run("updates a matching transaction", func(t *testing.T) {
		userRepo, farmRepo, productRepo := newVerifiedRepos(ownerID, farmID, productID)
		var updated dTransaction.Transaction
		transactionRepo := &fakeTransactionRepository{
			findTransactionByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dTransaction.Transaction, error) {
				return dTransaction.Transaction{ID: transactionID, ProductID: productID}, nil
			},
			updateTransactionFunc: func(_ context.Context, _ database.Querier, tx dTransaction.Transaction) error {
				updated = tx
				return nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, farmRepo, productRepo, transactionRepo)

		cmd := TransactionCommand{Type: dTransaction.TypeExpense, Amount: 99}
		if err := svc.Update(ctxWithUser(ownerID), farmID, productID, transactionID, cmd); err != nil {
			t.Fatalf("Update() error = %v, want nil", err)
		}
		if updated.Amount != 99 || updated.ID != transactionID {
			t.Errorf("updated = %+v, want ID=%v Amount=99", updated, transactionID)
		}
	})

	t.Run("rejects a transaction belonging to a different product", func(t *testing.T) {
		userRepo, farmRepo, productRepo := newVerifiedRepos(ownerID, farmID, productID)
		otherProductID := uuid.New()
		transactionRepo := &fakeTransactionRepository{
			findTransactionByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dTransaction.Transaction, error) {
				return dTransaction.Transaction{ID: transactionID, ProductID: otherProductID}, nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, farmRepo, productRepo, transactionRepo)

		err := svc.Update(ctxWithUser(ownerID), farmID, productID, transactionID, TransactionCommand{})
		if !errors.Is(err, dTransaction.ErrInvalidID) {
			t.Errorf("Update() error = %v, want %v", err, dTransaction.ErrInvalidID)
		}
	})
}

func TestService_Delete(t *testing.T) {
	ownerID := uuid.New()
	farmID := uuid.New()
	productID := uuid.New()
	ids := []uuid.UUID{uuid.New(), uuid.New()}

	userRepo, farmRepo, productRepo := newVerifiedRepos(ownerID, farmID, productID)
	var deletedProductID uuid.UUID
	var deletedIDs []uuid.UUID
	transactionRepo := &fakeTransactionRepository{
		deleteTransactionsFunc: func(_ context.Context, _ database.Querier, pid uuid.UUID, txIDs []uuid.UUID) error {
			deletedProductID = pid
			deletedIDs = txIDs
			return nil
		},
	}
	svc := NewService(fakeTransactor{}, userRepo, farmRepo, productRepo, transactionRepo)

	if err := svc.Delete(ctxWithUser(ownerID), farmID, productID, ids); err != nil {
		t.Fatalf("Delete() error = %v, want nil", err)
	}
	if deletedProductID != productID || len(deletedIDs) != len(ids) {
		t.Errorf("DeleteTransactions called with (%v, %v), want (%v, %v)", deletedProductID, deletedIDs, productID, ids)
	}
}

func TestService_DeleteMany(t *testing.T) {
	ownerID := uuid.New()
	farmID := uuid.New()
	productAID := uuid.New()
	productBID := uuid.New()
	txA1, txA2, txB1 := uuid.New(), uuid.New(), uuid.New()

	userRepo := &fakeUserRepository{
		findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
			return dUser.User{ID: ownerID}, nil
		},
	}
	farmRepo := &fakeFarmRepository{
		findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
			return dFarm.Farm{ID: farmID, UserID: ownerID}, nil
		},
	}
	var productLookups int
	productRepo := &fakeProductRepository{
		findProductByIDFunc: func(_ context.Context, _ database.Querier, id uuid.UUID) (dProduct.Product, error) {
			productLookups++
			return dProduct.Product{ID: id, FarmID: farmID}, nil
		},
	}
	transactions := map[uuid.UUID]dTransaction.Transaction{
		txA1: {ID: txA1, ProductID: productAID},
		txA2: {ID: txA2, ProductID: productAID},
		txB1: {ID: txB1, ProductID: productBID},
	}
	deletedByProduct := make(map[uuid.UUID][]uuid.UUID)
	transactionRepo := &fakeTransactionRepository{
		findTransactionByIDFunc: func(_ context.Context, _ database.Querier, id uuid.UUID) (dTransaction.Transaction, error) {
			return transactions[id], nil
		},
		deleteTransactionsFunc: func(_ context.Context, _ database.Querier, pid uuid.UUID, ids []uuid.UUID) error {
			deletedByProduct[pid] = ids
			return nil
		},
	}
	svc := NewService(fakeTransactor{}, userRepo, farmRepo, productRepo, transactionRepo)

	err := svc.DeleteMany(ctxWithUser(ownerID), []uuid.UUID{txA1, txA2, txB1})
	if err != nil {
		t.Fatalf("DeleteMany() error = %v, want nil", err)
	}

	if len(deletedByProduct[productAID]) != 2 {
		t.Errorf("deleted %d transactions for product A, want 2", len(deletedByProduct[productAID]))
	}
	if len(deletedByProduct[productBID]) != 1 {
		t.Errorf("deleted %d transactions for product B, want 1", len(deletedByProduct[productBID]))
	}
	// Product A appears twice in the input but should only be verified
	// once. Each verification does its own FindProductByID call (one
	// directly in DeleteMany's loop, one more inside verifyProduct itself),
	// so 2 distinct products means 4 lookups total, not 2 - a mildly
	// redundant round trip, but not incorrect.
	if productLookups != 4 {
		t.Errorf("FindProductByID called %d times, want 4 (2 distinct products x 2 lookups each)", productLookups)
	}
}

func TestService_DeleteAll(t *testing.T) {
	userID := uuid.New()
	farmAID := uuid.New()
	farmBID := uuid.New()
	productID := uuid.New()

	farmRepo := &fakeFarmRepository{
		listFarmByUserIDFunc: func(_ context.Context, _ database.Querier, uid uuid.UUID) ([]dFarm.Farm, error) {
			if uid != userID {
				t.Fatalf("ListFarmByUserID called with %v, want %v", uid, userID)
			}
			return []dFarm.Farm{{ID: farmAID, UserID: userID}, {ID: farmBID, UserID: userID}}, nil
		},
	}
	byFarm := map[uuid.UUID][]dTransaction.Transaction{
		farmAID: {{ID: uuid.New(), ProductID: productID}},
		farmBID: {},
	}
	var deletedCalls int
	transactionRepo := &fakeTransactionRepository{
		listTransactionsByFarmIDFunc: func(_ context.Context, _ database.Querier, fid uuid.UUID) ([]dTransaction.Transaction, error) {
			return byFarm[fid], nil
		},
		deleteTransactionsFunc: func(context.Context, database.Querier, uuid.UUID, []uuid.UUID) error {
			deletedCalls++
			return nil
		},
	}
	svc := NewService(fakeTransactor{}, &fakeUserRepository{}, farmRepo, &fakeProductRepository{}, transactionRepo)

	if err := svc.DeleteAll(ctxWithUser(userID)); err != nil {
		t.Fatalf("DeleteAll() error = %v, want nil", err)
	}
	if deletedCalls != 1 {
		t.Errorf("DeleteTransactions called %d times, want 1 (only farm A has transactions)", deletedCalls)
	}
}

/* ==================== Tests: verifyProduct edge cases ==================== */

func TestService_verifyProduct_InvalidUserID(t *testing.T) {
	userRepo := &fakeUserRepository{
		findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
			return dUser.User{}, nil
		},
	}
	svc := NewService(fakeTransactor{}, userRepo, &fakeFarmRepository{}, &fakeProductRepository{}, &fakeTransactionRepository{})

	_, err := svc.ListByProductID(ctxWithUser(uuid.New()), uuid.New(), uuid.New())
	if !errors.Is(err, dUser.ErrInvalidUserID) {
		t.Errorf("ListByProductID() error = %v, want %v", err, dUser.ErrInvalidUserID)
	}
}

func TestService_verifyProduct_ProductBelongsToDifferentFarm(t *testing.T) {
	ownerID := uuid.New()
	farmID := uuid.New()
	otherFarmID := uuid.New()
	productID := uuid.New()

	userRepo := &fakeUserRepository{
		findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
			return dUser.User{ID: ownerID}, nil
		},
	}
	farmRepo := &fakeFarmRepository{
		findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
			return dFarm.Farm{ID: farmID, UserID: ownerID}, nil
		},
	}
	productRepo := &fakeProductRepository{
		findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
			return dProduct.Product{ID: productID, FarmID: otherFarmID}, nil
		},
	}
	svc := NewService(fakeTransactor{}, userRepo, farmRepo, productRepo, &fakeTransactionRepository{})

	_, err := svc.ListByProductID(ctxWithUser(ownerID), farmID, productID)
	if !errors.Is(err, shared.ErrForbidden) {
		t.Errorf("ListByProductID() error = %v, want %v", err, shared.ErrForbidden)
	}
}

func TestService_verifyProduct_AdminBypassesOwnership(t *testing.T) {
	adminID := uuid.New()
	ownerID := uuid.New()
	farmID := uuid.New()
	productID := uuid.New()

	userRepo := &fakeUserRepository{
		findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
			return dUser.User{ID: adminID}, nil
		},
	}
	farmRepo := &fakeFarmRepository{
		findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
			return dFarm.Farm{ID: farmID, UserID: ownerID}, nil
		},
	}
	productRepo := &fakeProductRepository{
		findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
			return dProduct.Product{ID: productID, FarmID: farmID}, nil
		},
	}
	transactionRepo := &fakeTransactionRepository{
		listTransactionsByProductFunc: func(context.Context, database.Querier, uuid.UUID) ([]dTransaction.Transaction, error) {
			return nil, nil
		},
	}
	svc := NewService(fakeTransactor{}, userRepo, farmRepo, productRepo, transactionRepo)

	_, err := svc.ListByProductID(ctxWithUserAndRole(adminID, config.RoleAdmin), farmID, productID)
	if err != nil {
		t.Errorf("ListByProductID() error = %v, want nil", err)
	}
}
