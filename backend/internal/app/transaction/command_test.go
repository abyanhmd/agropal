package transaction

import (
	dTransaction "agropal/internal/domain/transaction"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestTransactionCommand_ToDomain(t *testing.T) {
	productID := uuid.New()
	transactionID := uuid.New()
	date := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	cmd := TransactionCommand{
		Date:        date,
		Type:        dTransaction.TypeExpense,
		Amount:      150.75,
		Description: "Fertilizer purchase",
	}

	got := cmd.ToDomain(productID, transactionID)

	want := dTransaction.Transaction{
		ID:          transactionID,
		Date:        date,
		Type:        dTransaction.TypeExpense,
		Amount:      150.75,
		Description: "Fertilizer purchase",
		ProductID:   productID,
	}

	if got != want {
		t.Errorf("ToDomain() = %+v, want %+v", got, want)
	}
}
