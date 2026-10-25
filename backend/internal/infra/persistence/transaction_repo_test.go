package persistence

import (
	dTransaction "agropal/internal/domain/transaction"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

// insertTestProduct inserts a product directly via SQL (bypassing
// productRepo, which is itself under test elsewhere) so FK-constrained
// tables like transactions have a valid product_id to reference.
func insertTestProduct(t *testing.T, farmID uuid.UUID) uuid.UUID {
	t.Helper()
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}

	productID := uuid.New()
	_, err := testDB.Exec(
		`INSERT INTO products (product_id, farm_id, name, description, image) VALUES ($1, $2, $3, $4, $5)`,
		productID, farmID, "Test Product", "desc", "img.png",
	)
	if err != nil {
		t.Fatalf("insertTestProduct: %v", err)
	}
	return productID
}

func TestTransactionRepository_CreateAndFindByID(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewTransactionRepository(testLogger)
	farmID := insertTestFarm(t)
	productID := insertTestProduct(t, farmID)
	now := time.Now().UTC().Truncate(time.Microsecond)

	txns := []dTransaction.Transaction{
		{ID: uuid.New(), Date: now, Type: dTransaction.TypeIncome, Amount: 100, Currency: "USD", Description: "Sale", ProductID: productID},
		{ID: uuid.New(), Date: now, Type: dTransaction.TypeExpense, Amount: 40, Currency: "USD", Description: "Fertilizer", ProductID: productID},
	}
	if err := repo.CreateTransaction(ctx, testDB, txns); err != nil {
		t.Fatalf("CreateTransaction() error = %v, want nil", err)
	}

	got, err := repo.FindTransactionByID(ctx, testDB, txns[0].ID)
	if err != nil {
		t.Fatalf("FindTransactionByID() error = %v, want nil", err)
	}
	if got != txns[0] {
		t.Errorf("FindTransactionByID() = %+v, want %+v", got, txns[0])
	}
}

func TestTransactionRepository_ListTransactionsByFarmID(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewTransactionRepository(testLogger)
	farmID := insertTestFarm(t)
	otherFarmID := insertTestFarm(t)
	productA := insertTestProduct(t, farmID)
	productB := insertTestProduct(t, farmID)
	productOther := insertTestProduct(t, otherFarmID)
	now := time.Now().UTC().Truncate(time.Microsecond)

	all := []dTransaction.Transaction{
		{ID: uuid.New(), Date: now, Type: dTransaction.TypeIncome, Amount: 1, Currency: "USD", ProductID: productA},
		{ID: uuid.New(), Date: now, Type: dTransaction.TypeIncome, Amount: 2, Currency: "USD", ProductID: productB},
		{ID: uuid.New(), Date: now, Type: dTransaction.TypeIncome, Amount: 3, Currency: "USD", ProductID: productOther},
	}
	if err := repo.CreateTransaction(ctx, testDB, all); err != nil {
		t.Fatalf("CreateTransaction() error = %v, want nil", err)
	}

	// Regression test: transactions has no farm_id column, so this must
	// join through products rather than filtering on a nonexistent column.
	got, err := repo.ListTransactionsByFarmID(ctx, testDB, farmID)
	if err != nil {
		t.Fatalf("ListTransactionsByFarmID() error = %v, want nil", err)
	}
	if len(got) != 2 {
		t.Fatalf("ListTransactionsByFarmID() returned %d transactions, want 2 (not the other farm's)", len(got))
	}
	for _, tx := range got {
		if tx.ProductID != productA && tx.ProductID != productB {
			t.Errorf("ListTransactionsByFarmID() returned a transaction for product %v, want only productA/productB", tx.ProductID)
		}
	}
}

func TestTransactionRepository_ListTransactionsByProductID(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewTransactionRepository(testLogger)
	farmID := insertTestFarm(t)
	productID := insertTestProduct(t, farmID)
	otherProductID := insertTestProduct(t, farmID)
	now := time.Now().UTC().Truncate(time.Microsecond)

	all := []dTransaction.Transaction{
		{ID: uuid.New(), Date: now, Type: dTransaction.TypeIncome, Amount: 1, Currency: "USD", ProductID: productID},
		{ID: uuid.New(), Date: now, Type: dTransaction.TypeIncome, Amount: 2, Currency: "USD", ProductID: otherProductID},
	}
	if err := repo.CreateTransaction(ctx, testDB, all); err != nil {
		t.Fatalf("CreateTransaction() error = %v, want nil", err)
	}

	got, err := repo.ListTransactionsByProductID(ctx, testDB, productID)
	if err != nil {
		t.Fatalf("ListTransactionsByProductID() error = %v, want nil", err)
	}
	if len(got) != 1 || got[0].ProductID != productID {
		t.Errorf("ListTransactionsByProductID() = %+v, want a single transaction for product %v", got, productID)
	}
}

func TestTransactionRepository_UpdateTransaction(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewTransactionRepository(testLogger)
	farmID := insertTestFarm(t)
	productID := insertTestProduct(t, farmID)
	now := time.Now().UTC().Truncate(time.Microsecond)

	txn := dTransaction.Transaction{ID: uuid.New(), Date: now, Type: dTransaction.TypeIncome, Amount: 10, Currency: "USD", Description: "Old", ProductID: productID}
	if err := repo.CreateTransaction(ctx, testDB, []dTransaction.Transaction{txn}); err != nil {
		t.Fatalf("CreateTransaction() error = %v, want nil", err)
	}

	txn.Amount = 99
	txn.Type = dTransaction.TypeExpense
	txn.Description = "New"
	if err := repo.UpdateTransaction(ctx, testDB, txn); err != nil {
		t.Fatalf("UpdateTransaction() error = %v, want nil", err)
	}

	got, err := repo.FindTransactionByID(ctx, testDB, txn.ID)
	if err != nil {
		t.Fatalf("FindTransactionByID() error = %v, want nil", err)
	}
	if got.Amount != 99 || got.Type != dTransaction.TypeExpense || got.Description != "New" {
		t.Errorf("got = %+v, want Amount=99 Type=EXPENSE Description=New", got)
	}
}

func TestTransactionRepository_DeleteTransactions(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewTransactionRepository(testLogger)
	farmID := insertTestFarm(t)
	productID := insertTestProduct(t, farmID)
	now := time.Now().UTC().Truncate(time.Microsecond)

	txns := []dTransaction.Transaction{
		{ID: uuid.New(), Date: now, Type: dTransaction.TypeIncome, Amount: 1, Currency: "USD", ProductID: productID},
		{ID: uuid.New(), Date: now, Type: dTransaction.TypeIncome, Amount: 2, Currency: "USD", ProductID: productID},
		{ID: uuid.New(), Date: now, Type: dTransaction.TypeIncome, Amount: 3, Currency: "USD", ProductID: productID},
	}
	if err := repo.CreateTransaction(ctx, testDB, txns); err != nil {
		t.Fatalf("CreateTransaction() error = %v, want nil", err)
	}

	// Delete only the first two, leave the third untouched.
	if err := repo.DeleteTransactions(ctx, testDB, productID, []uuid.UUID{txns[0].ID, txns[1].ID}); err != nil {
		t.Fatalf("DeleteTransactions() error = %v, want nil", err)
	}

	remaining, err := repo.ListTransactionsByProductID(ctx, testDB, productID)
	if err != nil {
		t.Fatalf("ListTransactionsByProductID() error = %v, want nil", err)
	}
	if len(remaining) != 1 || remaining[0].ID != txns[2].ID {
		t.Errorf("remaining = %+v, want only transaction %v", remaining, txns[2].ID)
	}
}
