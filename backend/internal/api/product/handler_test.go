package product

import (
	"agropal/config"
	aProduct "agropal/internal/app/product"
	dFarm "agropal/internal/domain/farm"
	dNote "agropal/internal/domain/note"
	dProduct "agropal/internal/domain/product"
	"agropal/internal/domain/shared"
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

// fakeTransactor runs WithTx's callback directly against a nil Querier; the
// fake repositories below never touch the Querier they're given.
type fakeTransactor struct{}

func (fakeTransactor) DB() database.Querier { return nil }

func (fakeTransactor) WithTx(_ context.Context, fn func(q database.Querier) error) error {
	return fn(nil)
}

// fakeUserRepository implements dUser.Repository. Only the methods the
// handler tests exercise are wired to a *Func field; everything else panics
// if called.
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

// fakeFarmRepository implements dFarm.Repository.
type fakeFarmRepository struct {
	findFarmByIDFunc func(ctx context.Context, q database.Querier, farmID uuid.UUID) (dFarm.Farm, error)
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

func (f *fakeFarmRepository) ListFarmByUserID(context.Context, database.Querier, uuid.UUID) ([]dFarm.Farm, error) {
	panic("fakeFarmRepository: ListFarmByUserID not stubbed")
}

// fakeNoteRepository implements dNote.Repository.
type fakeNoteRepository struct {
	createManyNotesFunc func(ctx context.Context, q database.Querier, notes []dNote.Note) error
}

func (f *fakeNoteRepository) CreateNote(context.Context, database.Querier, dNote.Note) error {
	panic("fakeNoteRepository: CreateNote not stubbed")
}

func (f *fakeNoteRepository) CreateManyNotes(ctx context.Context, q database.Querier, notes []dNote.Note) error {
	if f.createManyNotesFunc == nil {
		panic("fakeNoteRepository: createManyNotesFunc not stubbed")
	}
	return f.createManyNotesFunc(ctx, q, notes)
}

func (f *fakeNoteRepository) ListNotes(context.Context, database.Querier, dNote.NoteType, []uuid.UUID, map[string]string) ([]dNote.Note, error) {
	panic("fakeNoteRepository: ListNotes not stubbed")
}

func (f *fakeNoteRepository) DeleteNotes(context.Context, database.Querier, []uuid.UUID) error {
	panic("fakeNoteRepository: DeleteNotes not stubbed")
}

// fakeProductRepository implements dProduct.Repository. Only the methods the
// handler tests exercise are wired to a *Func field; everything else panics
// if called.
type fakeProductRepository struct {
	createProductFunc         func(ctx context.Context, q database.Querier, p dProduct.Product) error
	updateProductFunc         func(ctx context.Context, q database.Querier, p dProduct.Product) error
	deleteProductFunc         func(ctx context.Context, q database.Querier, id uuid.UUID) error
	findProductByIDFunc       func(ctx context.Context, q database.Querier, id uuid.UUID) (dProduct.Product, error)
	findProductDetailByIDFunc func(ctx context.Context, q database.Querier, id uuid.UUID) (dProduct.ProductDetail, error)
	listProductsByFarmIDFunc  func(ctx context.Context, q database.Querier, farmID uuid.UUID) ([]dProduct.ProductDetail, error)
	addProductStockFunc       func(ctx context.Context, q database.Querier, s dProduct.ProductStock) error
	addProductPriceFunc       func(ctx context.Context, q database.Querier, p dProduct.ProductPrice) error
	deleteProductStockFunc    func(ctx context.Context, q database.Querier, stockID uuid.UUID) error
	deleteProductPriceFunc    func(ctx context.Context, q database.Querier, priceID uuid.UUID) error
}

func (f *fakeProductRepository) CreateProduct(ctx context.Context, q database.Querier, p dProduct.Product) error {
	if f.createProductFunc == nil {
		panic("fakeProductRepository: createProductFunc not stubbed")
	}
	return f.createProductFunc(ctx, q, p)
}

func (f *fakeProductRepository) UpdateProduct(ctx context.Context, q database.Querier, p dProduct.Product) error {
	if f.updateProductFunc == nil {
		panic("fakeProductRepository: updateProductFunc not stubbed")
	}
	return f.updateProductFunc(ctx, q, p)
}

func (f *fakeProductRepository) DeleteProduct(ctx context.Context, q database.Querier, id uuid.UUID) error {
	if f.deleteProductFunc == nil {
		panic("fakeProductRepository: deleteProductFunc not stubbed")
	}
	return f.deleteProductFunc(ctx, q, id)
}

func (f *fakeProductRepository) FindProductByID(ctx context.Context, q database.Querier, id uuid.UUID) (dProduct.Product, error) {
	if f.findProductByIDFunc == nil {
		panic("fakeProductRepository: findProductByIDFunc not stubbed")
	}
	return f.findProductByIDFunc(ctx, q, id)
}

func (f *fakeProductRepository) FindProductDetailByID(ctx context.Context, q database.Querier, id uuid.UUID) (dProduct.ProductDetail, error) {
	if f.findProductDetailByIDFunc == nil {
		panic("fakeProductRepository: findProductDetailByIDFunc not stubbed")
	}
	return f.findProductDetailByIDFunc(ctx, q, id)
}

func (f *fakeProductRepository) ListProductsByFarmID(ctx context.Context, q database.Querier, farmID uuid.UUID) ([]dProduct.ProductDetail, error) {
	if f.listProductsByFarmIDFunc == nil {
		panic("fakeProductRepository: listProductsByFarmIDFunc not stubbed")
	}
	return f.listProductsByFarmIDFunc(ctx, q, farmID)
}

func (f *fakeProductRepository) AddProductStock(ctx context.Context, q database.Querier, s dProduct.ProductStock) error {
	if f.addProductStockFunc == nil {
		panic("fakeProductRepository: addProductStockFunc not stubbed")
	}
	return f.addProductStockFunc(ctx, q, s)
}

func (f *fakeProductRepository) AddProductPrice(ctx context.Context, q database.Querier, p dProduct.ProductPrice) error {
	if f.addProductPriceFunc == nil {
		panic("fakeProductRepository: addProductPriceFunc not stubbed")
	}
	return f.addProductPriceFunc(ctx, q, p)
}

func (f *fakeProductRepository) DeleteProductStock(ctx context.Context, q database.Querier, stockID uuid.UUID) error {
	if f.deleteProductStockFunc == nil {
		panic("fakeProductRepository: deleteProductStockFunc not stubbed")
	}
	return f.deleteProductStockFunc(ctx, q, stockID)
}

func (f *fakeProductRepository) DeleteProductPrice(ctx context.Context, q database.Querier, priceID uuid.UUID) error {
	if f.deleteProductPriceFunc == nil {
		panic("fakeProductRepository: deleteProductPriceFunc not stubbed")
	}
	return f.deleteProductPriceFunc(ctx, q, priceID)
}

/* ==================== Helpers ====================== */

type testRepos struct {
	userRepo    *fakeUserRepository
	farmRepo    *fakeFarmRepository
	productRepo *fakeProductRepository
	noteRepo    *fakeNoteRepository
}

// newTestHandler wires a real *aProduct.Service (backed by fakes) into a
// real Handler, so these tests exercise the handler's actual request
// parsing, status codes, and response bodies rather than a mocked service.
func newTestHandler(t *testing.T, repos testRepos) *Handler {
	t.Helper()

	if repos.userRepo == nil {
		repos.userRepo = &fakeUserRepository{}
	}
	if repos.farmRepo == nil {
		repos.farmRepo = &fakeFarmRepository{}
	}
	if repos.productRepo == nil {
		repos.productRepo = &fakeProductRepository{}
	}
	if repos.noteRepo == nil {
		repos.noteRepo = &fakeNoteRepository{}
	}

	svc := aProduct.NewService(fakeTransactor{}, repos.userRepo, repos.farmRepo, repos.productRepo, repos.noteRepo)
	logger := pkg.NewLogger(config.Config{})
	ioHandler := utils.NewIOHandler(logger, nil, utils.NewValidator())

	h := NewHandler(ioHandler, svc)
	return &h
}

// authedRequest builds a request carrying the metadata the JWT middleware
// would normally attach to the context (see internal/api/middleware), since
// the product service reads the caller's identity via shared.ExtractUserID
// and role checks via shared.GetMetadata(ctx, shared.KeyRole).
func authedRequest(method, path, body string, userID uuid.UUID, role string) *http.Request {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
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

/* ==================== Tests: CreateProduct ====================== */

func TestHandler_CreateProduct(t *testing.T) {
	ownerID := uuid.New()
	farmID := uuid.New()

	newRepos := func() testRepos {
		return testRepos{
			userRepo: &fakeUserRepository{
				findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
					return dUser.User{ID: ownerID}, nil
				},
			},
			farmRepo: &fakeFarmRepository{
				findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
					return dFarm.Farm{ID: farmID, UserID: ownerID}, nil
				},
			},
		}
	}

	t.Run("Success", func(t *testing.T) {
		repos := newRepos()
		repos.productRepo = &fakeProductRepository{
			createProductFunc: func(context.Context, database.Querier, dProduct.Product) error { return nil },
		}
		h := newTestHandler(t, repos)

		req := authedRequest(http.MethodPost, "/farm/x/product", `{"name":"Corn","description":"Sweet corn","image":"corn.png"}`, ownerID, "user")
		req.SetPathValue("farm_id", farmID.String())
		rec := httptest.NewRecorder()

		h.CreateProduct(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusCreated, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dProduct.SuccessCreated.String() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dProduct.SuccessCreated.String())
		}
	})

	t.Run("Invalid Farm ID In Path", func(t *testing.T) {
		h := newTestHandler(t, newRepos())

		req := authedRequest(http.MethodPost, "/farm/x/product", `{"name":"Corn","description":"Sweet corn","image":"corn.png"}`, ownerID, "user")
		req.SetPathValue("farm_id", "not-a-uuid")
		rec := httptest.NewRecorder()

		h.CreateProduct(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dProduct.ErrInvalidFarmID.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dProduct.ErrInvalidFarmID.Error())
		}
	})

	t.Run("Invalid JSON Body", func(t *testing.T) {
		h := newTestHandler(t, newRepos())

		req := authedRequest(http.MethodPost, "/farm/x/product", "not-json", ownerID, "user")
		req.SetPathValue("farm_id", farmID.String())
		rec := httptest.NewRecorder()

		h.CreateProduct(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != shared.ErrInvalidRequest.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, shared.ErrInvalidRequest.Error())
		}
	})

	t.Run("Missing Required Field Fails Struct Validation", func(t *testing.T) {
		repos := newRepos()
		repos.productRepo = &fakeProductRepository{
			createProductFunc: func(context.Context, database.Querier, dProduct.Product) error {
				t.Fatal("service should not be called when validation fails")
				return nil
			},
		}
		h := newTestHandler(t, repos)

		req := authedRequest(http.MethodPost, "/farm/x/product", `{"name":"Corn","description":"Sweet corn"}`, ownerID, "user")
		req.SetPathValue("farm_id", farmID.String())
		rec := httptest.NewRecorder()

		h.CreateProduct(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != shared.ErrValidation.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, shared.ErrValidation.Error())
		}
		if len(envelope.Errors["image"]) == 0 {
			t.Errorf("Errors[image] = %v, want at least one validation message", envelope.Errors["image"])
		}
	})

	t.Run("Non Owner Non Admin Maps To Forbidden", func(t *testing.T) {
		otherUserID := uuid.New()
		repos := testRepos{
			userRepo: &fakeUserRepository{
				findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
					return dUser.User{ID: otherUserID}, nil
				},
			},
			farmRepo: &fakeFarmRepository{
				findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
					return dFarm.Farm{ID: farmID, UserID: ownerID}, nil
				},
			},
		}
		h := newTestHandler(t, repos)

		req := authedRequest(http.MethodPost, "/farm/x/product", `{"name":"Corn","description":"Sweet corn","image":"corn.png"}`, otherUserID, "user")
		req.SetPathValue("farm_id", farmID.String())
		rec := httptest.NewRecorder()

		h.CreateProduct(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusForbidden, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != shared.ErrForbidden.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, shared.ErrForbidden.Error())
		}
	})
}

/* ==================== Tests: UpdateProduct ====================== */

func TestHandler_UpdateProduct(t *testing.T) {
	ownerID := uuid.New()
	farmID := uuid.New()
	productID := uuid.New()

	baseRepos := func() testRepos {
		return testRepos{
			userRepo: &fakeUserRepository{
				findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
					return dUser.User{ID: ownerID}, nil
				},
			},
			farmRepo: &fakeFarmRepository{
				findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
					return dFarm.Farm{ID: farmID, UserID: ownerID}, nil
				},
			},
			noteRepo: &fakeNoteRepository{
				createManyNotesFunc: func(context.Context, database.Querier, []dNote.Note) error { return nil },
			},
		}
	}

	t.Run("Success", func(t *testing.T) {
		repos := baseRepos()
		repos.productRepo = &fakeProductRepository{
			findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
				return dProduct.Product{ID: productID, FarmID: farmID, Name: "Corn"}, nil
			},
			updateProductFunc: func(context.Context, database.Querier, dProduct.Product) error { return nil },
		}
		h := newTestHandler(t, repos)

		req := authedRequest(http.MethodPatch, "/farm/x/product/y", `{"name":"Sweet Corn"}`, ownerID, "user")
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("product_id", productID.String())
		rec := httptest.NewRecorder()

		h.UpdateProduct(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dProduct.SuccessUpdated.String() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dProduct.SuccessUpdated.String())
		}
	})

	t.Run("Invalid Product ID In Path", func(t *testing.T) {
		h := newTestHandler(t, baseRepos())

		req := authedRequest(http.MethodPatch, "/farm/x/product/y", `{"name":"Sweet Corn"}`, ownerID, "user")
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("product_id", "not-a-uuid")
		rec := httptest.NewRecorder()

		h.UpdateProduct(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dProduct.ErrInvalidID.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dProduct.ErrInvalidID.Error())
		}
	})

	t.Run("Product Belongs To A Different Farm Maps To Forbidden", func(t *testing.T) {
		repos := baseRepos()
		repos.productRepo = &fakeProductRepository{
			findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
				return dProduct.Product{ID: productID, FarmID: uuid.New(), Name: "Corn"}, nil
			},
		}
		h := newTestHandler(t, repos)

		req := authedRequest(http.MethodPatch, "/farm/x/product/y", `{"name":"Sweet Corn"}`, ownerID, "user")
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("product_id", productID.String())
		rec := httptest.NewRecorder()

		h.UpdateProduct(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusForbidden, rec.Body.String())
		}
	})
}

/* ==================== Tests: GetOneProduct / GetManyProducts ====================== */

func TestHandler_GetOneProduct(t *testing.T) {
	ownerID := uuid.New()
	farmID := uuid.New()
	productID := uuid.New()

	baseRepos := func() testRepos {
		return testRepos{
			userRepo: &fakeUserRepository{
				findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
					return dUser.User{ID: ownerID}, nil
				},
			},
			farmRepo: &fakeFarmRepository{
				findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
					return dFarm.Farm{ID: farmID, UserID: ownerID}, nil
				},
			},
		}
	}

	t.Run("Success", func(t *testing.T) {
		repos := baseRepos()
		repos.productRepo = &fakeProductRepository{
			findProductDetailByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.ProductDetail, error) {
				return dProduct.ProductDetail{
					Product: dProduct.Product{ID: productID, FarmID: farmID, Name: "Corn"},
				}, nil
			},
		}
		h := newTestHandler(t, repos)

		req := authedRequest(http.MethodGet, "/farm/x/product/y", "", ownerID, "user")
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("product_id", productID.String())
		rec := httptest.NewRecorder()

		h.GetOneProduct(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		var got response.ProductDetailResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to decode response body %q: %v", rec.Body.String(), err)
		}
		if got.Product.ProductID != productID {
			t.Errorf("Product.ProductID = %v, want %v", got.Product.ProductID, productID)
		}
		if got.Product.Name != "Corn" {
			t.Errorf("Product.Name = %q, want %q", got.Product.Name, "Corn")
		}
	})

	t.Run("Product Detail From A Different Farm Maps To Forbidden", func(t *testing.T) {
		repos := baseRepos()
		repos.productRepo = &fakeProductRepository{
			findProductDetailByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.ProductDetail, error) {
				return dProduct.ProductDetail{Product: dProduct.Product{ID: productID, FarmID: uuid.New()}}, nil
			},
		}
		h := newTestHandler(t, repos)

		req := authedRequest(http.MethodGet, "/farm/x/product/y", "", ownerID, "user")
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("product_id", productID.String())
		rec := httptest.NewRecorder()

		h.GetOneProduct(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusForbidden, rec.Body.String())
		}
	})
}

func TestHandler_GetManyProducts(t *testing.T) {
	ownerID := uuid.New()
	farmID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		repos := testRepos{
			userRepo: &fakeUserRepository{
				findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
					return dUser.User{ID: ownerID}, nil
				},
			},
			farmRepo: &fakeFarmRepository{
				findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
					return dFarm.Farm{ID: farmID, UserID: ownerID}, nil
				},
			},
			productRepo: &fakeProductRepository{
				listProductsByFarmIDFunc: func(context.Context, database.Querier, uuid.UUID) ([]dProduct.ProductDetail, error) {
					return []dProduct.ProductDetail{
						{Product: dProduct.Product{ID: uuid.New(), FarmID: farmID, Name: "Corn"}},
						{Product: dProduct.Product{ID: uuid.New(), FarmID: farmID, Name: "Wheat"}},
					}, nil
				},
			},
		}
		h := newTestHandler(t, repos)

		req := authedRequest(http.MethodGet, "/farm/x/product", "", ownerID, "user")
		req.SetPathValue("farm_id", farmID.String())
		rec := httptest.NewRecorder()

		h.GetManyProducts(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		var got []response.ProductDetailResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to decode response body %q: %v", rec.Body.String(), err)
		}
		if len(got) != 2 {
			t.Fatalf("len(got) = %d, want %d", len(got), 2)
		}
	})
}

/* ==================== Tests: DeleteProduct ====================== */

func TestHandler_DeleteProduct(t *testing.T) {
	ownerID := uuid.New()
	farmID := uuid.New()
	productID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		repos := testRepos{
			userRepo: &fakeUserRepository{
				findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
					return dUser.User{ID: ownerID}, nil
				},
			},
			farmRepo: &fakeFarmRepository{
				findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
					return dFarm.Farm{ID: farmID, UserID: ownerID}, nil
				},
			},
			productRepo: &fakeProductRepository{
				deleteProductFunc: func(context.Context, database.Querier, uuid.UUID) error { return nil },
			},
		}
		h := newTestHandler(t, repos)

		req := authedRequest(http.MethodDelete, "/farm/x/product/y", "", ownerID, "user")
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("product_id", productID.String())
		rec := httptest.NewRecorder()

		h.DeleteProduct(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dProduct.SuccessDeleted.String() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dProduct.SuccessDeleted.String())
		}
	})

	t.Run("Invalid Product ID In Path", func(t *testing.T) {
		repos := testRepos{
			userRepo: &fakeUserRepository{
				findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
					return dUser.User{ID: ownerID}, nil
				},
			},
			farmRepo: &fakeFarmRepository{
				findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
					return dFarm.Farm{ID: farmID, UserID: ownerID}, nil
				},
			},
		}
		h := newTestHandler(t, repos)

		req := authedRequest(http.MethodDelete, "/farm/x/product/y", "", ownerID, "user")
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("product_id", "not-a-uuid")
		rec := httptest.NewRecorder()

		h.DeleteProduct(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dProduct.ErrInvalidID.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dProduct.ErrInvalidID.Error())
		}
	})
}

/* ==================== Tests: Product Price ====================== */

func TestHandler_CreateProductPrice(t *testing.T) {
	farmID := uuid.New()
	productID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		repos := testRepos{
			productRepo: &fakeProductRepository{
				findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
					return dProduct.Product{ID: productID, FarmID: farmID}, nil
				},
				addProductPriceFunc: func(context.Context, database.Querier, dProduct.ProductPrice) error { return nil },
			},
		}
		h := newTestHandler(t, repos)

		req := httptest.NewRequest(http.MethodPost, "/farm/x/product/y/price", bytes.NewBufferString(`{"date_added":"2025-01-01T00:00:00Z","price":10.5}`))
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("product_id", productID.String())
		rec := httptest.NewRecorder()

		h.CreateProductPrice(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusCreated, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dProduct.SuccessPriceAdded.String() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dProduct.SuccessPriceAdded.String())
		}
	})

	t.Run("Non Positive Price Fails Struct Validation Before Hitting The Service", func(t *testing.T) {
		repos := testRepos{
			productRepo: &fakeProductRepository{
				findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
					t.Fatal("service should not be called when validation fails")
					return dProduct.Product{}, nil
				},
			},
		}
		h := newTestHandler(t, repos)

		req := httptest.NewRequest(http.MethodPost, "/farm/x/product/y/price", bytes.NewBufferString(`{"date_added":"2025-01-01T00:00:00Z","price":0}`))
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("product_id", productID.String())
		rec := httptest.NewRecorder()

		h.CreateProductPrice(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	})

	t.Run("Product/Farm Mismatch Maps To 400", func(t *testing.T) {
		repos := testRepos{
			productRepo: &fakeProductRepository{
				findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
					return dProduct.Product{ID: productID, FarmID: uuid.New()}, nil
				},
			},
		}
		h := newTestHandler(t, repos)

		req := httptest.NewRequest(http.MethodPost, "/farm/x/product/y/price", bytes.NewBufferString(`{"date_added":"2025-01-01T00:00:00Z","price":10.5}`))
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("product_id", productID.String())
		rec := httptest.NewRecorder()

		h.CreateProductPrice(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	})
}

func TestHandler_DeleteProductPrice(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		priceID := uuid.New()
		repos := testRepos{
			productRepo: &fakeProductRepository{
				deleteProductPriceFunc: func(context.Context, database.Querier, uuid.UUID) error { return nil },
			},
		}
		h := newTestHandler(t, repos)

		req := httptest.NewRequest(http.MethodDelete, "/farm/x/product/y/price/z", nil)
		req.SetPathValue("price_id", priceID.String())
		rec := httptest.NewRecorder()

		h.DeleteProductPrice(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dProduct.SuccessPriceDeleted.String() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dProduct.SuccessPriceDeleted.String())
		}
	})

	t.Run("Invalid Price ID In Path", func(t *testing.T) {
		h := newTestHandler(t, testRepos{})

		req := httptest.NewRequest(http.MethodDelete, "/farm/x/product/y/price/z", nil)
		req.SetPathValue("price_id", "not-a-uuid")
		rec := httptest.NewRecorder()

		h.DeleteProductPrice(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dProduct.ErrInvalidPriceID.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dProduct.ErrInvalidPriceID.Error())
		}
	})
}

/* ==================== Tests: Product Stock ====================== */

func TestHandler_CreateProductStock(t *testing.T) {
	farmID := uuid.New()
	productID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		repos := testRepos{
			productRepo: &fakeProductRepository{
				findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
					return dProduct.Product{ID: productID, FarmID: farmID}, nil
				},
				addProductStockFunc: func(context.Context, database.Querier, dProduct.ProductStock) error { return nil },
			},
		}
		h := newTestHandler(t, repos)

		req := httptest.NewRequest(http.MethodPost, "/farm/x/product/y/stock", bytes.NewBufferString(`{"date_added":"2025-01-01T00:00:00Z","type":"IN","stock":5}`))
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("product_id", productID.String())
		rec := httptest.NewRecorder()

		h.CreateProductStock(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusCreated, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dProduct.SuccessStockAdded.String() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dProduct.SuccessStockAdded.String())
		}
	})

	t.Run("Invalid Movement Type Fails Struct Validation", func(t *testing.T) {
		repos := testRepos{
			productRepo: &fakeProductRepository{
				findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
					t.Fatal("service should not be called when validation fails")
					return dProduct.Product{}, nil
				},
			},
		}
		h := newTestHandler(t, repos)

		req := httptest.NewRequest(http.MethodPost, "/farm/x/product/y/stock", bytes.NewBufferString(`{"date_added":"2025-01-01T00:00:00Z","type":"SIDEWAYS","stock":5}`))
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("product_id", productID.String())
		rec := httptest.NewRecorder()

		h.CreateProductStock(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	})
}

func TestHandler_DeleteProductStock(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		stockID := uuid.New()
		repos := testRepos{
			productRepo: &fakeProductRepository{
				deleteProductStockFunc: func(context.Context, database.Querier, uuid.UUID) error { return nil },
			},
		}
		h := newTestHandler(t, repos)

		req := httptest.NewRequest(http.MethodDelete, "/farm/x/product/y/stock/z", nil)
		req.SetPathValue("stock_id", stockID.String())
		rec := httptest.NewRecorder()

		h.DeleteProductStock(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dProduct.SuccessStockDeleted.String() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dProduct.SuccessStockDeleted.String())
		}
	})

	t.Run("Invalid Stock ID In Path", func(t *testing.T) {
		h := newTestHandler(t, testRepos{})

		req := httptest.NewRequest(http.MethodDelete, "/farm/x/product/y/stock/z", nil)
		req.SetPathValue("stock_id", "not-a-uuid")
		rec := httptest.NewRecorder()

		h.DeleteProductStock(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dProduct.ErrInvalidStockID.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dProduct.ErrInvalidStockID.Error())
		}
	})
}
