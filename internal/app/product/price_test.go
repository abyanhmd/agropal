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

func TestService_CreateProductPrice(t *testing.T) {
	farmID := uuid.New()
	productID := uuid.New()
	product := dProduct.Product{ID: productID, FarmID: farmID}
	requesterID := uuid.New()

	t.Run("adds a price for a matching farm/product pair", func(t *testing.T) {
		var added dProduct.ProductPrice
		productRepo := &fakeProductRepository{
			findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
				return product, nil
			},
			addProductPriceFunc: func(_ context.Context, _ database.Querier, p dProduct.ProductPrice) error {
				added = p
				return nil
			},
		}
		svc := NewService(fakeTransactor{}, &fakeUserRepository{}, &fakeFarmRepository{}, productRepo, &fakeNoteRepository{})

		cmd := ProductPriceCommand{DateAdded: shared.Ptr(time.Now().UTC().Add(-time.Hour)), Price: shared.Ptr(19.99)}
		if err := svc.CreateProductPrice(ctxWithUser(requesterID), farmID, productID, cmd); err != nil {
			t.Fatalf("CreateProductPrice() error = %v, want nil", err)
		}
		if added.ProductID != productID || added.Price != 19.99 {
			t.Errorf("added = %+v, want ProductID=%v Price=19.99", added, productID)
		}
	})

	// CreateProductPrice only checks that the product belongs to the given
	// farmID - it never calls verifyFarm, so nothing here confirms the
	// *requester* actually owns that farm. Unlike CreateProduct/
	// UpdateProduct/GetOneProduct/DeleteProduct, which all call verifyFarm,
	// this method (and CreateProductStock/DeleteProductPrice/
	// DeleteProductStock) currently has no such check. This test locks in
	// that existing behavior rather than the more restrictive one you'd
	// expect from the sibling methods.
	t.Run("does not check the requester owns the farm", func(t *testing.T) {
		productRepo := &fakeProductRepository{
			findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
				return product, nil
			},
			addProductPriceFunc: func(context.Context, database.Querier, dProduct.ProductPrice) error {
				return nil
			},
		}
		svc := NewService(fakeTransactor{}, &fakeUserRepository{}, &fakeFarmRepository{}, productRepo, &fakeNoteRepository{})

		strangerID := uuid.New()
		cmd := ProductPriceCommand{DateAdded: shared.Ptr(time.Now().UTC()), Price: shared.Ptr(1.0)}
		if err := svc.CreateProductPrice(ctxWithUser(strangerID), farmID, productID, cmd); err != nil {
			t.Fatalf("CreateProductPrice() error = %v, want nil (no ownership check exists)", err)
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

		cmd := ProductPriceCommand{DateAdded: shared.Ptr(time.Now().UTC()), Price: shared.Ptr(1.0)}
		err := svc.CreateProductPrice(ctxWithUser(requesterID), otherFarmID, productID, cmd)
		if err == nil {
			t.Fatal("CreateProductPrice() error = nil, want a farm/product mismatch error")
		}
	})

	t.Run("rejects a non-positive price", func(t *testing.T) {
		productRepo := &fakeProductRepository{
			findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
				return product, nil
			},
		}
		svc := NewService(fakeTransactor{}, &fakeUserRepository{}, &fakeFarmRepository{}, productRepo, &fakeNoteRepository{})

		cmd := ProductPriceCommand{DateAdded: shared.Ptr(time.Now().UTC()), Price: shared.Ptr(0.0)}
		if err := svc.CreateProductPrice(ctxWithUser(requesterID), farmID, productID, cmd); err == nil {
			t.Fatal("CreateProductPrice() error = nil, want a non-positive-price error")
		}
	})

	t.Run("rejects a date in the future", func(t *testing.T) {
		productRepo := &fakeProductRepository{
			findProductByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dProduct.Product, error) {
				return product, nil
			},
		}
		svc := NewService(fakeTransactor{}, &fakeUserRepository{}, &fakeFarmRepository{}, productRepo, &fakeNoteRepository{})

		cmd := ProductPriceCommand{DateAdded: shared.Ptr(time.Now().UTC().Add(time.Hour)), Price: shared.Ptr(1.0)}
		if err := svc.CreateProductPrice(ctxWithUser(requesterID), farmID, productID, cmd); err == nil {
			t.Fatal("CreateProductPrice() error = nil, want a future-date error")
		}
	})
}

func TestService_DeleteProductPrice(t *testing.T) {
	priceID := uuid.New()
	var deletedID uuid.UUID

	productRepo := &fakeProductRepository{
		deleteProductPriceFunc: func(_ context.Context, _ database.Querier, id uuid.UUID) error {
			deletedID = id
			return nil
		},
	}
	svc := NewService(fakeTransactor{}, &fakeUserRepository{}, &fakeFarmRepository{}, productRepo, &fakeNoteRepository{})

	if err := svc.DeleteProductPrice(context.Background(), priceID); err != nil {
		t.Fatalf("DeleteProductPrice() error = %v, want nil", err)
	}
	if deletedID != priceID {
		t.Errorf("deleted price ID = %v, want %v", deletedID, priceID)
	}
}
