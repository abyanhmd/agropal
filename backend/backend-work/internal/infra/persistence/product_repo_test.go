package persistence

import (
	dProduct "agropal/internal/domain/product"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestProductRepository_CreateAndFindByID(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewProductRepository(testLogger)
	farmID := insertTestFarm(t)

	product := dProduct.Product{ID: uuid.New(), FarmID: farmID, Name: "Tomato", Description: "Fresh", Image: "t.png", Unit: "kg"}
	if err := repo.CreateProduct(ctx, testDB, product); err != nil {
		t.Fatalf("CreateProduct() error = %v, want nil", err)
	}

	got, err := repo.FindProductByID(ctx, testDB, product.ID)
	if err != nil {
		t.Fatalf("FindProductByID() error = %v, want nil", err)
	}
	if got != product {
		t.Errorf("FindProductByID() = %+v, want %+v", got, product)
	}
}

func TestProductRepository_UpdateProduct(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewProductRepository(testLogger)
	farmID := insertTestFarm(t)

	product := dProduct.Product{ID: uuid.New(), FarmID: farmID, Name: "Old Name", Description: "Old Desc", Image: "old.png"}
	if err := repo.CreateProduct(ctx, testDB, product); err != nil {
		t.Fatalf("CreateProduct() error = %v, want nil", err)
	}

	product.Name = "New Name"
	product.Description = "New Desc"
	if err := repo.UpdateProduct(ctx, testDB, product); err != nil {
		t.Fatalf("UpdateProduct() error = %v, want nil", err)
	}

	got, err := repo.FindProductByID(ctx, testDB, product.ID)
	if err != nil {
		t.Fatalf("FindProductByID() error = %v, want nil", err)
	}
	if got.Name != "New Name" || got.Description != "New Desc" {
		t.Errorf("got = %+v, want Name=New Name Description=New Desc", got)
	}
}

func TestProductRepository_DeleteProduct(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewProductRepository(testLogger)
	farmID := insertTestFarm(t)

	product := dProduct.Product{ID: uuid.New(), FarmID: farmID, Name: "To Delete"}
	if err := repo.CreateProduct(ctx, testDB, product); err != nil {
		t.Fatalf("CreateProduct() error = %v, want nil", err)
	}

	if err := repo.DeleteProduct(ctx, testDB, product.ID); err != nil {
		t.Fatalf("DeleteProduct() error = %v, want nil", err)
	}
	if _, err := repo.FindProductByID(ctx, testDB, product.ID); err == nil {
		t.Error("FindProductByID() error = nil after DeleteProduct, want a not-found error")
	}
}

// TestProductRepository_FindProductDetailByID is a regression test for a
// bug where FindProductDetailByID indexed into nil prices/stocks slices
// without allocating them first, guaranteeing an index-out-of-range panic
// whenever a product actually had price or stock history.
func TestProductRepository_FindProductDetailByID(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewProductRepository(testLogger)
	farmID := insertTestFarm(t)

	product := dProduct.Product{ID: uuid.New(), FarmID: farmID, Name: "Tomato"}
	if err := repo.CreateProduct(ctx, testDB, product); err != nil {
		t.Fatalf("CreateProduct() error = %v, want nil", err)
	}

	now := time.Now().UTC().Truncate(time.Microsecond)
	stock := dProduct.ProductStock{ID: uuid.New(), ProductID: product.ID, DateAdded: now, Type: dProduct.MovementIncome, Stock: 10}
	if err := repo.AddProductStock(ctx, testDB, stock); err != nil {
		t.Fatalf("AddProductStock() error = %v, want nil", err)
	}
	price := dProduct.ProductPrice{ID: uuid.New(), ProductID: product.ID, DateAdded: now, Price: 5.5, Currency: "EUR"}
	if err := repo.AddProductPrice(ctx, testDB, price); err != nil {
		t.Fatalf("AddProductPrice() error = %v, want nil", err)
	}

	got, err := repo.FindProductDetailByID(ctx, testDB, product.ID)
	if err != nil {
		t.Fatalf("FindProductDetailByID() error = %v, want nil", err)
	}
	if got.Product != product {
		t.Errorf("Product = %+v, want %+v", got.Product, product)
	}
	if len(got.Stocks) != 1 || got.Stocks[0].Stock != 10 || got.Stocks[0].Type != dProduct.MovementIncome {
		t.Errorf("Stocks = %+v, want a single income entry of 10", got.Stocks)
	}
	if len(got.Prices) != 1 || got.Prices[0].Price != 5.5 || got.Prices[0].Currency != "EUR" {
		t.Errorf("Prices = %+v, want a single entry of 5.5 EUR", got.Prices)
	}
}

func TestProductRepository_FindProductDetailByID_NoHistory(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewProductRepository(testLogger)
	farmID := insertTestFarm(t)

	product := dProduct.Product{ID: uuid.New(), FarmID: farmID, Name: "No History"}
	if err := repo.CreateProduct(ctx, testDB, product); err != nil {
		t.Fatalf("CreateProduct() error = %v, want nil", err)
	}

	got, err := repo.FindProductDetailByID(ctx, testDB, product.ID)
	if err != nil {
		t.Fatalf("FindProductDetailByID() error = %v, want nil", err)
	}
	if len(got.Stocks) != 0 || len(got.Prices) != 0 {
		t.Errorf("got = %+v, want empty Stocks/Prices, not nil-slice panics", got)
	}
}

func TestProductRepository_ListProductsByFarmID(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewProductRepository(testLogger)
	farmID := insertTestFarm(t)
	otherFarmID := insertTestFarm(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	productA := dProduct.Product{ID: uuid.New(), FarmID: farmID, Name: "A"}
	productB := dProduct.Product{ID: uuid.New(), FarmID: farmID, Name: "B"}
	productOther := dProduct.Product{ID: uuid.New(), FarmID: otherFarmID, Name: "Other"}
	for _, p := range []dProduct.Product{productA, productB, productOther} {
		if err := repo.CreateProduct(ctx, testDB, p); err != nil {
			t.Fatalf("CreateProduct(%q) error = %v, want nil", p.Name, err)
		}
	}

	stockA := dProduct.ProductStock{ID: uuid.New(), ProductID: productA.ID, DateAdded: now, Type: dProduct.MovementIncome, Stock: 3}
	if err := repo.AddProductStock(ctx, testDB, stockA); err != nil {
		t.Fatalf("AddProductStock() error = %v, want nil", err)
	}
	priceB := dProduct.ProductPrice{ID: uuid.New(), ProductID: productB.ID, DateAdded: now, Price: 1.25}
	if err := repo.AddProductPrice(ctx, testDB, priceB); err != nil {
		t.Fatalf("AddProductPrice() error = %v, want nil", err)
	}

	got, err := repo.ListProductsByFarmID(ctx, testDB, farmID)
	if err != nil {
		t.Fatalf("ListProductsByFarmID() error = %v, want nil", err)
	}
	if len(got) != 2 {
		t.Fatalf("ListProductsByFarmID() returned %d products, want 2 (not the other farm's)", len(got))
	}

	byID := make(map[uuid.UUID]dProduct.ProductDetail, len(got))
	for _, d := range got {
		byID[d.Product.ID] = d
	}
	if detail, ok := byID[productA.ID]; !ok || len(detail.Stocks) != 1 || detail.Stocks[0].Stock != 3 {
		t.Errorf("product A detail = %+v, want a single stock entry of 3", detail)
	}
	if detail, ok := byID[productB.ID]; !ok || len(detail.Prices) != 1 || detail.Prices[0].Price != 1.25 {
		t.Errorf("product B detail = %+v, want a single price entry of 1.25", detail)
	}
}

func TestProductRepository_DeleteProductStockAndPrice(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewProductRepository(testLogger)
	farmID := insertTestFarm(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	product := dProduct.Product{ID: uuid.New(), FarmID: farmID, Name: "Tomato"}
	if err := repo.CreateProduct(ctx, testDB, product); err != nil {
		t.Fatalf("CreateProduct() error = %v, want nil", err)
	}

	stock := dProduct.ProductStock{ID: uuid.New(), ProductID: product.ID, DateAdded: now, Type: dProduct.MovementExpense, Stock: 2}
	if err := repo.AddProductStock(ctx, testDB, stock); err != nil {
		t.Fatalf("AddProductStock() error = %v, want nil", err)
	}
	price := dProduct.ProductPrice{ID: uuid.New(), ProductID: product.ID, DateAdded: now, Price: 9.99}
	if err := repo.AddProductPrice(ctx, testDB, price); err != nil {
		t.Fatalf("AddProductPrice() error = %v, want nil", err)
	}

	if err := repo.DeleteProductStock(ctx, testDB, stock.ID); err != nil {
		t.Fatalf("DeleteProductStock() error = %v, want nil", err)
	}
	if err := repo.DeleteProductPrice(ctx, testDB, price.ID); err != nil {
		t.Fatalf("DeleteProductPrice() error = %v, want nil", err)
	}

	got, err := repo.FindProductDetailByID(ctx, testDB, product.ID)
	if err != nil {
		t.Fatalf("FindProductDetailByID() error = %v, want nil", err)
	}
	if len(got.Stocks) != 0 || len(got.Prices) != 0 {
		t.Errorf("got = %+v, want both Stocks and Prices empty after deletion", got)
	}
}
