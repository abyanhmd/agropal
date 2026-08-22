package product

import (
	"agropal/config"
	dFarm "agropal/internal/domain/farm"
	dNote "agropal/internal/domain/note"
	dProduct "agropal/internal/domain/product"
	"agropal/internal/domain/shared"
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
// exercised by product.Service, so every other method panics if called - a
// test relying on an un-stubbed call fails loudly instead of silently
// zero-valuing.
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

// fakeFarmRepository implements dFarm.Repository. Only FindFarmByID is
// exercised by product.Service.
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

// fakeProductRepository implements dProduct.Repository. Shared by
// product_test.go, price_test.go and stock_test.go (same package).
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

// fakeNoteRepository implements dNote.Repository. Only CreateManyNotes is
// exercised by product.Service today.
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

/* ==================== Helpers ====================== */

func ctxWithUser(userID uuid.UUID) context.Context {
	return shared.WithMetadata(context.Background(), shared.KeyUserID, userID.String())
}

func ctxWithUserAndRole(userID uuid.UUID, role config.Role) context.Context {
	ctx := ctxWithUser(userID)
	return shared.WithMetadata(ctx, shared.KeyRole, string(role))
}

/* ==================== Tests: CreateProduct / GetOne / GetMany / Delete ==================== */

func TestService_CreateProduct(t *testing.T) {
	ownerID := uuid.New()
	farmID := uuid.New()
	farm := dFarm.Farm{ID: farmID, UserID: ownerID}

	t.Run("creates a product for the farm's owner", func(t *testing.T) {
		var created dProduct.Product

		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: ownerID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return farm, nil
			},
		}
		productRepo := &fakeProductRepository{
			createProductFunc: func(_ context.Context, _ database.Querier, p dProduct.Product) error {
				created = p
				return nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, farmRepo, productRepo, &fakeNoteRepository{})

		cmd := ProductCommand{Name: shared.Ptr("Tomato"), Description: shared.Ptr("Fresh"), Image: shared.Ptr("t.png")}
		if err := svc.CreateProduct(ctxWithUser(ownerID), farmID, cmd); err != nil {
			t.Fatalf("CreateProduct() error = %v, want nil", err)
		}
		if created.FarmID != farmID || created.Name != "Tomato" {
			t.Errorf("created = %+v, want FarmID=%v Name=Tomato", created, farmID)
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
				return farm, nil
			},
		}
		productRepo := &fakeProductRepository{
			createProductFunc: func(context.Context, database.Querier, dProduct.Product) error {
				t.Fatal("CreateProduct should not be called when the requester is forbidden")
				return nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, farmRepo, productRepo, &fakeNoteRepository{})

		cmd := ProductCommand{Name: shared.Ptr("Tomato"), Description: shared.Ptr("Fresh"), Image: shared.Ptr("t.png")}
		err := svc.CreateProduct(ctxWithUser(otherUserID), farmID, cmd)
		if !errors.Is(err, shared.ErrForbidden) {
			t.Errorf("CreateProduct() error = %v, want %v", err, shared.ErrForbidden)
		}
	})
}

func TestService_GetOneProduct(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()
	farmID := uuid.New()
	productID := uuid.New()
	farm := dFarm.Farm{ID: farmID, UserID: ownerID}
	detail := dProduct.ProductDetail{Product: dProduct.Product{ID: productID, FarmID: farmID, Name: "Tomato"}}

	t.Run("returns the product detail to its owner", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: ownerID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return farm, nil
			},
		}
		productRepo := &fakeProductRepository{
			findProductDetailByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.ProductDetail, error) {
				return detail, nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, farmRepo, productRepo, &fakeNoteRepository{})

		got, err := svc.GetOneProduct(ctxWithUser(ownerID), farmID, productID)
		if err != nil {
			t.Fatalf("GetOneProduct() error = %v, want nil", err)
		}
		if got.Product != detail.Product {
			t.Errorf("GetOneProduct().Product = %+v, want %+v", got.Product, detail.Product)
		}
	})

	t.Run("forbids a non-owner", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: otherUserID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return farm, nil
			},
		}
		productRepo := &fakeProductRepository{
			findProductDetailByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.ProductDetail, error) {
				return detail, nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, farmRepo, productRepo, &fakeNoteRepository{})

		_, err := svc.GetOneProduct(ctxWithUser(otherUserID), farmID, productID)
		if !errors.Is(err, shared.ErrForbidden) {
			t.Errorf("GetOneProduct() error = %v, want %v", err, shared.ErrForbidden)
		}
	})
}

func TestService_GetManyProducts(t *testing.T) {
	ownerID := uuid.New()
	farmID := uuid.New()
	want := []dProduct.ProductDetail{
		{Product: dProduct.Product{ID: uuid.New(), FarmID: farmID, Name: "Tomato"}},
		{Product: dProduct.Product{ID: uuid.New(), FarmID: farmID, Name: "Potato"}},
	}

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
		listProductsByFarmIDFunc: func(_ context.Context, _ database.Querier, fid uuid.UUID) ([]dProduct.ProductDetail, error) {
			if fid != farmID {
				t.Fatalf("ListProductsByFarmID called with %v, want %v", fid, farmID)
			}
			return want, nil
		},
	}
	svc := NewService(fakeTransactor{}, userRepo, farmRepo, productRepo, &fakeNoteRepository{})

	got, err := svc.GetManyProducts(ctxWithUser(ownerID), farmID)
	if err != nil {
		t.Fatalf("GetManyProducts() error = %v, want nil", err)
	}
	if len(got) != len(want) {
		t.Fatalf("GetManyProducts() returned %d products, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].Product != want[i].Product {
			t.Errorf("GetManyProducts()[%d].Product = %+v, want %+v", i, got[i].Product, want[i].Product)
		}
	}
}

func TestService_DeleteProduct(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()
	farmID := uuid.New()
	productID := uuid.New()
	farm := dFarm.Farm{ID: farmID, UserID: ownerID}

	t.Run("owner can delete", func(t *testing.T) {
		var deletedID uuid.UUID
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: ownerID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return farm, nil
			},
		}
		productRepo := &fakeProductRepository{
			deleteProductFunc: func(_ context.Context, _ database.Querier, id uuid.UUID) error {
				deletedID = id
				return nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, farmRepo, productRepo, &fakeNoteRepository{})

		if err := svc.DeleteProduct(ctxWithUser(ownerID), farmID, productID); err != nil {
			t.Fatalf("DeleteProduct() error = %v, want nil", err)
		}
		if deletedID != productID {
			t.Errorf("deleted product ID = %v, want %v", deletedID, productID)
		}
	})

	t.Run("forbids a non-owner", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: otherUserID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return farm, nil
			},
		}
		productRepo := &fakeProductRepository{
			deleteProductFunc: func(context.Context, database.Querier, uuid.UUID) error {
				t.Fatal("DeleteProduct should not be called when the requester is forbidden")
				return nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, farmRepo, productRepo, &fakeNoteRepository{})

		err := svc.DeleteProduct(ctxWithUser(otherUserID), farmID, productID)
		if !errors.Is(err, shared.ErrForbidden) {
			t.Errorf("DeleteProduct() error = %v, want %v", err, shared.ErrForbidden)
		}
	})
}

/* ==================== Tests: UpdateProduct ==================== */

func TestService_UpdateProduct(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()
	adminID := uuid.New()
	farmID := uuid.New()
	productID := uuid.New()
	farm := dFarm.Farm{ID: farmID, UserID: ownerID}
	existing := dProduct.Product{ID: productID, FarmID: farmID, Name: "Old Name", Description: "Old Desc", Image: "old.png"}

	t.Run("owner update records a note and persists the product row", func(t *testing.T) {
		var notes []dNote.Note
		var persisted dProduct.Product

		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: ownerID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return farm, nil
			},
		}
		productRepo := &fakeProductRepository{
			findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
				return existing, nil
			},
			updateProductFunc: func(_ context.Context, _ database.Querier, p dProduct.Product) error {
				persisted = p
				return nil
			},
		}
		noteRepo := &fakeNoteRepository{
			createManyNotesFunc: func(_ context.Context, _ database.Querier, n []dNote.Note) error {
				notes = n
				return nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, farmRepo, productRepo, noteRepo)

		cmd := ProductCommand{Name: shared.Ptr("New Name")}
		if err := svc.UpdateProduct(ctxWithUser(ownerID), farmID, productID, cmd); err != nil {
			t.Fatalf("UpdateProduct() error = %v, want nil", err)
		}

		if len(notes) != 1 {
			t.Fatalf("got %d audit notes, want 1", len(notes))
		}
		if shared.Str(notes[0].OldValue) != "Old Name" || shared.Str(notes[0].NewValue) != "New Name" {
			t.Errorf("note old/new = %q/%q, want %q/%q", shared.Str(notes[0].OldValue), shared.Str(notes[0].NewValue), "Old Name", "New Name")
		}

		// Locks in a currently-documented limitation (see README's Known
		// Limitations): UpdateProduct records the audit note but saves the
		// unmodified row it fetched, so the persisted name is still the old
		// one even though the note claims it changed.
		if persisted.Name != "Old Name" {
			t.Errorf("persisted.Name = %q, want %q (UpdateProduct does not yet apply cmd fields before persisting)", persisted.Name, "Old Name")
		}
	})

	t.Run("non-owner, non-admin is forbidden", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: otherUserID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return farm, nil
			},
		}
		productRepo := &fakeProductRepository{
			findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
				return existing, nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, farmRepo, productRepo, &fakeNoteRepository{})

		err := svc.UpdateProduct(ctxWithUser(otherUserID), farmID, productID, ProductCommand{Name: shared.Ptr("New Name")})
		if !errors.Is(err, shared.ErrForbidden) {
			t.Errorf("UpdateProduct() error = %v, want %v", err, shared.ErrForbidden)
		}
	})

	t.Run("admin can update any farm's product", func(t *testing.T) {
		var notes []dNote.Note
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: adminID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return farm, nil
			},
		}
		productRepo := &fakeProductRepository{
			findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
				return existing, nil
			},
			updateProductFunc: func(context.Context, database.Querier, dProduct.Product) error {
				return nil
			},
		}
		noteRepo := &fakeNoteRepository{
			createManyNotesFunc: func(_ context.Context, _ database.Querier, n []dNote.Note) error {
				notes = n
				return nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, farmRepo, productRepo, noteRepo)

		err := svc.UpdateProduct(ctxWithUserAndRole(adminID, config.RoleAdmin), farmID, productID, ProductCommand{Description: shared.Ptr("New Desc")})
		if err != nil {
			t.Fatalf("UpdateProduct() error = %v, want nil", err)
		}
		if len(notes) != 1 || shared.Str(notes[0].UpdateField) != "description" {
			t.Errorf("notes = %+v, want a single note for field %q", notes, "description")
		}
	})

	t.Run("no changed fields records no notes", func(t *testing.T) {
		var noteCalls int
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: ownerID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return farm, nil
			},
		}
		productRepo := &fakeProductRepository{
			findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
				return existing, nil
			},
			updateProductFunc: func(context.Context, database.Querier, dProduct.Product) error {
				return nil
			},
		}
		noteRepo := &fakeNoteRepository{
			createManyNotesFunc: func(_ context.Context, _ database.Querier, n []dNote.Note) error {
				noteCalls++
				if len(n) != 0 {
					t.Errorf("CreateManyNotes called with %d notes, want 0", len(n))
				}
				return nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, farmRepo, productRepo, noteRepo)

		if err := svc.UpdateProduct(ctxWithUser(ownerID), farmID, productID, ProductCommand{}); err != nil {
			t.Fatalf("UpdateProduct() error = %v, want nil", err)
		}
		if noteCalls != 1 {
			t.Errorf("CreateManyNotes called %d times, want 1", noteCalls)
		}
	})
}

/* ==================== Tests: verifyFarm ==================== */

func TestService_verifyFarm_InvalidUserID(t *testing.T) {
	userRepo := &fakeUserRepository{
		findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
			return dUser.User{}, nil
		},
	}
	svc := NewService(fakeTransactor{}, userRepo, &fakeFarmRepository{}, &fakeProductRepository{}, &fakeNoteRepository{})

	_, err := svc.GetManyProducts(ctxWithUser(uuid.New()), uuid.New())
	if !errors.Is(err, dUser.ErrInvalidUserID) {
		t.Errorf("GetManyProducts() error = %v, want %v", err, dUser.ErrInvalidUserID)
	}
}
