package product

import (
	dProduct "agropal/internal/domain/product"
	"agropal/internal/domain/shared"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestProductCommand_ToDomain(t *testing.T) {
	farmID := uuid.New()
	productID := uuid.New()

	cmd := ProductCommand{
		Name:        shared.Ptr("Tomato"),
		Description: shared.Ptr("Fresh red tomatoes"),
		Image:       shared.Ptr("tomato.png"),
	}

	got := cmd.ToDomain(farmID, productID)

	want := dProduct.Product{
		ID:          productID,
		Name:        "Tomato",
		Description: "Fresh red tomatoes",
		Image:       "tomato.png",
		FarmID:      farmID,
	}

	if got != want {
		t.Errorf("ToDomain() = %+v, want %+v", got, want)
	}
}

func TestProductStockCommand_ToDomain(t *testing.T) {
	productID := uuid.New()
	stockID := uuid.New()
	dateAdded := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	movementIn := dProduct.MovementIn

	cmd := ProductStockCommand{
		DateAdded: &dateAdded,
		Type:      &movementIn,
		Stock:     shared.Ptr(42.5),
	}

	got := cmd.ToDomain(productID, stockID)

	want := dProduct.ProductStock{
		ID:        stockID,
		ProductID: productID,
		DateAdded: dateAdded,
		Type:      dProduct.MovementIn,
		Stock:     42.5,
	}

	if got != want {
		t.Errorf("ToDomain() = %+v, want %+v", got, want)
	}
}

func TestProductPriceCommand_ToDomain(t *testing.T) {
	productID := uuid.New()
	priceID := uuid.New()
	dateAdded := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	cmd := ProductPriceCommand{
		DateAdded: &dateAdded,
		Price:     shared.Ptr(19.99),
	}

	got := cmd.ToDomain(productID, priceID)

	want := dProduct.ProductPrice{
		ID:        priceID,
		ProductID: productID,
		DateAdded: dateAdded,
		Price:     19.99,
	}

	if got != want {
		t.Errorf("ToDomain() = %+v, want %+v", got, want)
	}
}
