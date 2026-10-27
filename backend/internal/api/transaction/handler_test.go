package transaction

import (
	"agropal/config"
	aTransaction "agropal/internal/app/transaction"
	dFarm "agropal/internal/domain/farm"
	dProduct "agropal/internal/domain/product"
	"agropal/internal/domain/shared"
	dTransaction "agropal/internal/domain/transaction"
	dUser "agropal/internal/domain/user"
	"agropal/internal/infra/database"
	"agropal/internal/response"
	"agropal/internal/utils"
	"agropal/pkg"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

/* ==================== Fakes ====================== */

type fakeTransactor struct{}

func (fakeTransactor) DB() database.Querier { return nil }

func (fakeTransactor) WithTx(_ context.Context, fn func(q database.Querier) error) error {
	return fn(nil)
}

// fakeUserRepository implements dUser.Repository. Only FindUserByID is
// exercised by the transaction service; everything else panics if called.
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

// fakeFarmRepository implements dFarm.Repository.
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

// fakeProductRepository implements dProduct.Repository.
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
	createTransactionFunc           func(ctx context.Context, q database.Querier, t []dTransaction.Transaction) error
	findTransactionByIDFunc         func(ctx context.Context, q database.Querier, id uuid.UUID) (dTransaction.Transaction, error)
	listTransactionsByFarmIDFunc    func(ctx context.Context, q database.Querier, farmID uuid.UUID) ([]dTransaction.Transaction, error)
	listTransactionsByProductIDFunc func(ctx context.Context, q database.Querier, productID uuid.UUID) ([]dTransaction.Transaction, error)
	updateTransactionFunc           func(ctx context.Context, q database.Querier, t dTransaction.Transaction) error
	deleteTransactionsFunc          func(ctx context.Context, q database.Querier, productID uuid.UUID, transactionIDs []uuid.UUID) error
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
	if f.listTransactionsByProductIDFunc == nil {
		panic("fakeTransactionRepository: listTransactionsByProductIDFunc not stubbed")
	}
	return f.listTransactionsByProductIDFunc(ctx, q, productID)
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

func newTestHandler(userRepo dUser.Repository, farmRepo dFarm.Repository, productRepo dProduct.Repository, transactionRepo dTransaction.Repository) *Handler {
	logger := pkg.NewLogger(config.Config{})
	svc := aTransaction.NewService(fakeTransactor{}, userRepo, farmRepo, productRepo, transactionRepo)
	ioHandler := utils.NewIOHandler(logger, nil, utils.NewValidator())
	h := NewHandler(ioHandler, svc)
	return &h
}

// authedRequest builds a request carrying the metadata the auth middleware
// would normally set post-JWT-verification (see internal/api/middleware),
// since these handler tests call handlers directly without going through
// real routing/middleware.
func authedRequest(method, path, body string, userID uuid.UUID, role string) *http.Request {
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, bytes.NewBufferString(body))
	}

	ctx := shared.WithMetadata(req.Context(), shared.KeyUserID, userID.String())
	ctx = shared.WithMetadata(ctx, shared.KeyRole, role)
	return req.WithContext(ctx)
}

func decodeEnvelope(t *testing.T, rec *httptest.ResponseRecorder) utils.StandardEnvelope {
	t.Helper()
	var envelope utils.StandardEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to decode response body %q: %v", rec.Body.String(), err)
	}
	return envelope
}

/* ==================== Tests: Create ====================== */

func TestHandler_Create(t *testing.T) {
	userID := uuid.New()
	farmID := uuid.New()
	productID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: userID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return dFarm.Farm{ID: farmID, UserID: userID}, nil
			},
		}
		productRepo := &fakeProductRepository{
			findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
				return dProduct.Product{ID: productID, FarmID: farmID}, nil
			},
		}
		transactionRepo := &fakeTransactionRepository{
			createTransactionFunc: func(context.Context, database.Querier, []dTransaction.Transaction) error { return nil },
		}
		h := newTestHandler(userRepo, farmRepo, productRepo, transactionRepo)

		req := authedRequest(http.MethodPost, "/transactions", `{"transactions":[{"date":"2024-01-01T00:00:00Z","type":"INCOME","amount":100,"currency":"USD","description":"sale"}]}`, userID, "user")
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("product_id", productID.String())
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusCreated, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dTransaction.SuccessCreated.String() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dTransaction.SuccessCreated.String())
		}
	})

	t.Run("Invalid Farm ID Path Value", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakeFarmRepository{}, &fakeProductRepository{}, &fakeTransactionRepository{})

		req := authedRequest(http.MethodPost, "/transactions", `{}`, userID, "user")
		req.SetPathValue("farm_id", "not-a-uuid")
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("Validation Failure Does Not Hit The Service", func(t *testing.T) {
		transactionRepo := &fakeTransactionRepository{
			createTransactionFunc: func(context.Context, database.Querier, []dTransaction.Transaction) error {
				t.Fatal("service should not be called when validation fails")
				return nil
			},
		}
		h := newTestHandler(&fakeUserRepository{}, &fakeFarmRepository{}, &fakeProductRepository{}, transactionRepo)

		req := authedRequest(http.MethodPost, "/transactions", `{"transactions":[]}`, userID, "user")
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("product_id", productID.String())
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	})

	t.Run("Forbidden When Requester Does Not Own The Farm", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: userID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return dFarm.Farm{ID: farmID, UserID: uuid.New()}, nil
			},
		}
		h := newTestHandler(userRepo, farmRepo, &fakeProductRepository{}, &fakeTransactionRepository{})

		req := authedRequest(http.MethodPost, "/transactions", `{"transactions":[{"date":"2024-01-01T00:00:00Z","type":"INCOME","amount":100,"currency":"USD","description":"sale"}]}`, userID, "user")
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("product_id", productID.String())
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusForbidden, rec.Body.String())
		}
	})
}

/* ==================== Tests: GetOne / GetMany ====================== */

func TestHandler_GetOne(t *testing.T) {
	userID := uuid.New()
	farmID := uuid.New()
	productID := uuid.New()
	transactionID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: userID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return dFarm.Farm{ID: farmID, UserID: userID}, nil
			},
		}
		productRepo := &fakeProductRepository{
			findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
				return dProduct.Product{ID: productID, FarmID: farmID}, nil
			},
		}
		transactionRepo := &fakeTransactionRepository{
			findTransactionByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dTransaction.Transaction, error) {
				return dTransaction.Transaction{ID: transactionID, ProductID: productID, Amount: 42}, nil
			},
		}
		h := newTestHandler(userRepo, farmRepo, productRepo, transactionRepo)

		req := authedRequest(http.MethodGet, "/transactions", "", userID, "user")
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("product_id", productID.String())
		req.SetPathValue("transaction_id", transactionID.String())
		rec := httptest.NewRecorder()

		h.GetOne(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		var out response.TransactionResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if out.TransactionID != transactionID {
			t.Errorf("TransactionID = %v, want %v", out.TransactionID, transactionID)
		}
		if out.Amount != 42 {
			t.Errorf("Amount = %v, want %v", out.Amount, 42.0)
		}
	})

	t.Run("Invalid Transaction ID Path Value", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakeFarmRepository{}, &fakeProductRepository{}, &fakeTransactionRepository{})

		req := authedRequest(http.MethodGet, "/transactions", "", userID, "user")
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("product_id", productID.String())
		req.SetPathValue("transaction_id", "not-a-uuid")
		rec := httptest.NewRecorder()

		h.GetOne(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("Transaction Belonging To Different Product Maps To 400", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: userID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return dFarm.Farm{ID: farmID, UserID: userID}, nil
			},
		}
		productRepo := &fakeProductRepository{
			findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
				return dProduct.Product{ID: productID, FarmID: farmID}, nil
			},
		}
		transactionRepo := &fakeTransactionRepository{
			findTransactionByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dTransaction.Transaction, error) {
				return dTransaction.Transaction{ID: transactionID, ProductID: uuid.New()}, nil
			},
		}
		h := newTestHandler(userRepo, farmRepo, productRepo, transactionRepo)

		req := authedRequest(http.MethodGet, "/transactions", "", userID, "user")
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("product_id", productID.String())
		req.SetPathValue("transaction_id", transactionID.String())
		rec := httptest.NewRecorder()

		h.GetOne(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	})
}

func TestHandler_GetMany(t *testing.T) {
	userID := uuid.New()
	farmID := uuid.New()
	productID := uuid.New()

	userRepo := func() *fakeUserRepository {
		return &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: userID}, nil
			},
		}
	}
	farmRepo := func() *fakeFarmRepository {
		return &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return dFarm.Farm{ID: farmID, UserID: userID}, nil
			},
		}
	}

	t.Run("Lists By Farm ID When No Product In Path", func(t *testing.T) {
		transactionRepo := &fakeTransactionRepository{
			listTransactionsByFarmIDFunc: func(context.Context, database.Querier, uuid.UUID) ([]dTransaction.Transaction, error) {
				return []dTransaction.Transaction{{ID: uuid.New()}, {ID: uuid.New()}}, nil
			},
		}
		h := newTestHandler(userRepo(), farmRepo(), &fakeProductRepository{}, transactionRepo)

		req := authedRequest(http.MethodGet, "/transactions", "", userID, "user")
		req.SetPathValue("farm_id", farmID.String())
		rec := httptest.NewRecorder()

		h.GetMany(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		var out []response.TransactionResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(out) != 2 {
			t.Errorf("len(out) = %d, want %d", len(out), 2)
		}
	})

	t.Run("Lists By Product ID When Present In Path", func(t *testing.T) {
		productRepo := &fakeProductRepository{
			findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
				return dProduct.Product{ID: productID, FarmID: farmID}, nil
			},
		}
		var gotProductID uuid.UUID
		transactionRepo := &fakeTransactionRepository{
			listTransactionsByProductIDFunc: func(_ context.Context, _ database.Querier, pID uuid.UUID) ([]dTransaction.Transaction, error) {
				gotProductID = pID
				return []dTransaction.Transaction{{ID: uuid.New()}}, nil
			},
		}
		h := newTestHandler(userRepo(), farmRepo(), productRepo, transactionRepo)

		req := authedRequest(http.MethodGet, "/transactions", "", userID, "user")
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("product_id", productID.String())
		rec := httptest.NewRecorder()

		h.GetMany(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		if gotProductID != productID {
			t.Errorf("ListTransactionsByProductID called with %v, want %v", gotProductID, productID)
		}
	})
}

/* ==================== Tests: Update ====================== */

func TestHandler_Update(t *testing.T) {
	userID := uuid.New()
	farmID := uuid.New()
	productID := uuid.New()
	transactionID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: userID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return dFarm.Farm{ID: farmID, UserID: userID}, nil
			},
		}
		productRepo := &fakeProductRepository{
			findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
				return dProduct.Product{ID: productID, FarmID: farmID}, nil
			},
		}
		transactionRepo := &fakeTransactionRepository{
			findTransactionByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dTransaction.Transaction, error) {
				return dTransaction.Transaction{ID: transactionID, ProductID: productID}, nil
			},
			updateTransactionFunc: func(context.Context, database.Querier, dTransaction.Transaction) error { return nil },
		}
		h := newTestHandler(userRepo, farmRepo, productRepo, transactionRepo)

		req := authedRequest(http.MethodPut, "/transactions", `{"date":"2024-01-01T00:00:00Z","type":"EXPENSE","amount":50,"currency":"USD","description":"feed"}`, userID, "user")
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("product_id", productID.String())
		req.SetPathValue("transaction_id", transactionID.String())
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dTransaction.SuccessUpdated.String() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dTransaction.SuccessUpdated.String())
		}
	})

	t.Run("Invalid Type Fails Validation", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakeFarmRepository{}, &fakeProductRepository{}, &fakeTransactionRepository{})

		req := authedRequest(http.MethodPut, "/transactions", `{"date":"2024-01-01T00:00:00Z","type":"NOT_A_TYPE","amount":50,"currency":"USD","description":"feed"}`, userID, "user")
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("product_id", productID.String())
		req.SetPathValue("transaction_id", transactionID.String())
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	})
}

/* ==================== Tests: Delete / DeleteMany / DeleteAll ====================== */

func TestHandler_Delete(t *testing.T) {
	userID := uuid.New()
	farmID := uuid.New()
	productID := uuid.New()
	transactionID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: userID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return dFarm.Farm{ID: farmID, UserID: userID}, nil
			},
		}
		productRepo := &fakeProductRepository{
			findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
				return dProduct.Product{ID: productID, FarmID: farmID}, nil
			},
		}
		transactionRepo := &fakeTransactionRepository{
			deleteTransactionsFunc: func(context.Context, database.Querier, uuid.UUID, []uuid.UUID) error { return nil },
		}
		h := newTestHandler(userRepo, farmRepo, productRepo, transactionRepo)

		req := authedRequest(http.MethodDelete, "/transactions", `{"transaction_ids":["`+transactionID.String()+`"]}`, userID, "user")
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("product_id", productID.String())
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dTransaction.SuccessDeleted.String() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dTransaction.SuccessDeleted.String())
		}
	})

	t.Run("Invalid Transaction ID In Body", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakeFarmRepository{}, &fakeProductRepository{}, &fakeTransactionRepository{})

		req := authedRequest(http.MethodDelete, "/transactions", `{"transaction_ids":["not-a-uuid"]}`, userID, "user")
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("product_id", productID.String())
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	})
}

func TestHandler_DeleteMany(t *testing.T) {
	userID := uuid.New()
	farmID := uuid.New()
	productID := uuid.New()
	transactionID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: userID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return dFarm.Farm{ID: farmID, UserID: userID}, nil
			},
		}
		productRepo := &fakeProductRepository{
			findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
				return dProduct.Product{ID: productID, FarmID: farmID}, nil
			},
		}
		transactionRepo := &fakeTransactionRepository{
			findTransactionByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dTransaction.Transaction, error) {
				return dTransaction.Transaction{ID: transactionID, ProductID: productID}, nil
			},
			deleteTransactionsFunc: func(context.Context, database.Querier, uuid.UUID, []uuid.UUID) error { return nil },
		}
		h := newTestHandler(userRepo, farmRepo, productRepo, transactionRepo)

		req := authedRequest(http.MethodDelete, "/transactions/many", `{"transaction_ids":["`+transactionID.String()+`"]}`, userID, "user")
		rec := httptest.NewRecorder()

		h.DeleteMany(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
	})

	t.Run("Invalid JSON Body", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakeFarmRepository{}, &fakeProductRepository{}, &fakeTransactionRepository{})

		req := authedRequest(http.MethodDelete, "/transactions/many", "not-json", userID, "user")
		rec := httptest.NewRecorder()

		h.DeleteMany(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

func TestHandler_DeleteAll(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		farmID := uuid.New()

		farmRepo := &fakeFarmRepository{
			listFarmByUserIDFunc: func(context.Context, database.Querier, uuid.UUID) ([]dFarm.Farm, error) {
				return []dFarm.Farm{{ID: farmID, UserID: userID}}, nil
			},
		}
		transactionRepo := &fakeTransactionRepository{
			listTransactionsByFarmIDFunc: func(context.Context, database.Querier, uuid.UUID) ([]dTransaction.Transaction, error) {
				return []dTransaction.Transaction{}, nil
			},
		}
		h := newTestHandler(&fakeUserRepository{}, farmRepo, &fakeProductRepository{}, transactionRepo)

		req := authedRequest(http.MethodDelete, "/transactions/all", "", userID, "user")
		rec := httptest.NewRecorder()

		h.DeleteAll(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
	})

	t.Run("Missing Authenticated User Maps To 401", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakeFarmRepository{}, &fakeProductRepository{}, &fakeTransactionRepository{})

		req := httptest.NewRequest(http.MethodDelete, "/transactions/all", nil)
		rec := httptest.NewRecorder()

		h.DeleteAll(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusUnauthorized, rec.Body.String())
		}
	})
}
