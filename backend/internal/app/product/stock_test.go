package product

import (
	dProduct "agropal/internal/domain/product"
	"agropal/internal/domain/shared"
	"agropal/internal/infra/database"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

// Fakes and helpers (fakeTransactor, fakeProductRepository, ctxWithUser,
// etc.) are defined in product_test.go, same package.

func TestService_CreateProductStock(t *testing.T) {
	farmID := uuid.New()
	productID := uuid.New()
	product := dProduct.Product{ID: productID, FarmID: farmID}
	requesterID := uuid.New()

	t.Run("adds stock for a matching farm/product pair", func(t *testing.T) {
		var added dProduct.ProductStock
		productRepo := &fakeProductRepository{
			findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
				return product, nil
			},
			addProductStockFunc: func(_ context.Context, _ database.Querier, s dProduct.ProductStock) error {
				added = s
				return nil
			},
		}
		svc := NewService(fakeTransactor{}, &fakeUserRepository{}, &fakeFarmRepository{}, productRepo, &fakeNoteRepository{})

		cmd := ProductStockCommand{
			DateAdded: shared.Ptr(time.Now().UTC().Add(-time.Hour)),
			Type:      shared.Ptr(dProduct.MovementIncome),
			Stock:     shared.Ptr(42.0),
		}
		if _, err := svc.CreateProductStock(ctxWithUser(requesterID), farmID, productID, cmd); err != nil {
			t.Fatalf("CreateProductStock() error = %v, want nil", err)
		}
		if added.ProductID != productID || added.Stock != 42.0 || added.Type != dProduct.MovementIncome {
			t.Errorf("added = %+v, want ProductID=%v Stock=42 Type=IN", added, productID)
		}
	})

	// Same gap as CreateProductPrice (see price_test.go): no verifyFarm
	// call, so the requester's relationship to the farm is never checked.
	t.Run("does not check the requester owns the farm", func(t *testing.T) {
		productRepo := &fakeProductRepository{
			findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
				return product, nil
			},
			addProductStockFunc: func(context.Context, database.Querier, dProduct.ProductStock) error {
				return nil
			},
		}
		svc := NewService(fakeTransactor{}, &fakeUserRepository{}, &fakeFarmRepository{}, productRepo, &fakeNoteRepository{})

		strangerID := uuid.New()
		cmd := ProductStockCommand{
			DateAdded: shared.Ptr(time.Now().UTC()),
			Type:      shared.Ptr(dProduct.MovementExpense),
			Stock:     shared.Ptr(1.0),
		}
		if _, err := svc.CreateProductStock(ctxWithUser(strangerID), farmID, productID, cmd); err != nil {
			t.Fatalf("CreateProductStock() error = %v, want nil (no ownership check exists)", err)
		}
	})

	t.Run("rejects a farm/product mismatch", func(t *testing.T) {
		otherFarmID := uuid.New()
		productRepo := &fakeProductRepository{
			findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
				return product, nil
			},
		}
		svc := NewService(fakeTransactor{}, &fakeUserRepository{}, &fakeFarmRepository{}, productRepo, &fakeNoteRepository{})

		cmd := ProductStockCommand{
			DateAdded: shared.Ptr(time.Now().UTC()),
			Type:      shared.Ptr(dProduct.MovementIncome),
			Stock:     shared.Ptr(1.0),
		}
		_, err := svc.CreateProductStock(ctxWithUser(requesterID), otherFarmID, productID, cmd)
		if err == nil {
			t.Fatal("CreateProductStock() error = nil, want a farm/product mismatch error")
		}
	})

	t.Run("rejects a non-positive stock amount", func(t *testing.T) {
		productRepo := &fakeProductRepository{
			findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
				return product, nil
			},
		}
		svc := NewService(fakeTransactor{}, &fakeUserRepository{}, &fakeFarmRepository{}, productRepo, &fakeNoteRepository{})

		cmd := ProductStockCommand{
			DateAdded: shared.Ptr(time.Now().UTC()),
			Type:      shared.Ptr(dProduct.MovementIncome),
			Stock:     shared.Ptr(0.0),
		}
		if _, err := svc.CreateProductStock(ctxWithUser(requesterID), farmID, productID, cmd); err == nil {
			t.Fatal("CreateProductStock() error = nil, want a non-positive-stock error")
		}
	})

	t.Run("rejects a date in the future", func(t *testing.T) {
		productRepo := &fakeProductRepository{
			findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
				return product, nil
			},
		}
		svc := NewService(fakeTransactor{}, &fakeUserRepository{}, &fakeFarmRepository{}, productRepo, &fakeNoteRepository{})

		cmd := ProductStockCommand{
			DateAdded: shared.Ptr(time.Now().UTC().Add(time.Hour)),
			Type:      shared.Ptr(dProduct.MovementIncome),
			Stock:     shared.Ptr(1.0),
		}
		if _, err := svc.CreateProductStock(ctxWithUser(requesterID), farmID, productID, cmd); err == nil {
			t.Fatal("CreateProductStock() error = nil, want a future-date error")
		}
	})
}

func TestService_DeleteProductStock(t *testing.T) {
	stockID := uuid.New()
	var deletedID uuid.UUID

	productRepo := &fakeProductRepository{
		deleteProductStockFunc: func(_ context.Context, _ database.Querier, id uuid.UUID) error {
			deletedID = id
			return nil
		},
	}
	svc := NewService(fakeTransactor{}, &fakeUserRepository{}, &fakeFarmRepository{}, productRepo, &fakeNoteRepository{})

	if err := svc.DeleteProductStock(context.Background(), stockID); err != nil {
		t.Fatalf("DeleteProductStock() error = %v, want nil", err)
	}
	if deletedID != stockID {
		t.Errorf("deleted stock ID = %v, want %v", deletedID, stockID)
	}
}
