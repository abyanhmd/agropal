package database

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func rowExists(t *testing.T, id string) bool {
	t.Helper()

	var count int
	if err := testDB.Get(&count, `SELECT COUNT(*) FROM transactor_test WHERE id = $1`, id); err != nil {
		t.Fatalf("failed to check row existence: %v", err)
	}
	return count == 1
}

func TestSQLTransactor(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}

	t.Run("DB Returns A Usable Querier For One Off Calls", func(t *testing.T) {
		transactor := NewTransactor(testDB)
		q := transactor.DB()

		var result int
		if err := q.GetContext(context.Background(), &result, `SELECT 1`); err != nil {
			t.Fatalf("GetContext() error = %v, want nil", err)
		}
		if result != 1 {
			t.Errorf("GetContext() result = %d, want 1", result)
		}
	})

	t.Run("With Tx Commits When Fn Succeeds", func(t *testing.T) {
		transactor := NewTransactor(testDB)
		id := uuid.NewString()

		err := transactor.WithTx(context.Background(), func(q Querier) error {
			_, err := q.ExecContext(context.Background(), `INSERT INTO transactor_test (id) VALUES ($1)`, id)
			return err
		})
		if err != nil {
			t.Fatalf("WithTx() error = %v, want nil", err)
		}
		if !rowExists(t, id) {
			t.Error("expected row to be committed and visible outside the transaction")
		}
	})

	t.Run("With Tx Rolls Back When Fn Returns An Error", func(t *testing.T) {
		transactor := NewTransactor(testDB)
		id := uuid.NewString()
		wantErr := errors.New("boom")

		err := transactor.WithTx(context.Background(), func(q Querier) error {
			if _, err := q.ExecContext(context.Background(), `INSERT INTO transactor_test (id) VALUES ($1)`, id); err != nil {
				return err
			}
			return wantErr
		})
		if !errors.Is(err, wantErr) {
			t.Fatalf("WithTx() error = %v, want %v", err, wantErr)
		}
		if rowExists(t, id) {
			t.Error("expected the insert to be rolled back after fn returned an error")
		}
	})

	t.Run("With Tx Rolls Back And Repanics When Fn Panics", func(t *testing.T) {
		transactor := NewTransactor(testDB)
		id := uuid.NewString()

		func() {
			defer func() {
				r := recover()
				if r == nil {
					t.Fatal("expected WithTx to repanic, it did not")
				}
				if r != "boom" {
					t.Errorf("recovered panic = %v, want %q", r, "boom")
				}
			}()

			_ = transactor.WithTx(context.Background(), func(q Querier) error {
				if _, err := q.ExecContext(context.Background(), `INSERT INTO transactor_test (id) VALUES ($1)`, id); err != nil {
					t.Fatalf("insert before panic failed: %v", err)
				}
				panic("boom")
			})
		}()

		if rowExists(t, id) {
			t.Error("expected the insert to be rolled back after fn panicked")
		}
	})

	t.Run("With Tx Propagates Begin Errors", func(t *testing.T) {
		transactor := NewTransactor(testDB)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := transactor.WithTx(ctx, func(q Querier) error {
			t.Fatal("fn should not run when BeginTxx fails")
			return nil
		})
		if err == nil {
			t.Fatal("WithTx() error = nil, want an error from BeginTxx on an already-canceled context")
		}
	})

	t.Run("Querier Rebind Adapts Placeholders For The Driver", func(t *testing.T) {
		transactor := NewTransactor(testDB)
		q := transactor.DB()

		got := q.Rebind(`SELECT * FROM transactor_test WHERE id = ?`)
		want := `SELECT * FROM transactor_test WHERE id = $1`
		if got != want {
			t.Errorf("Rebind() = %q, want %q", got, want)
		}
	})
}

// compile-time assurance that *sqlx.DB and *sqlx.Tx both satisfy Querier, so
// the interface stays a faithful narrowing of what SQLTransactor hands out
// (DB() returns the former, WithTx's fn receives the latter).
var (
	_ Querier = (*sqlx.DB)(nil)
	_ Querier = (*sqlx.Tx)(nil)
)
