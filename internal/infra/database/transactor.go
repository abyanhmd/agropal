package database

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type Querier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row

	GetContext(ctx context.Context, dest interface{}, query string, args ...any) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...any) error
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)

	Rebind(query string) string
}

// Transactor gives app-layer services a Querier for one-off calls (DB) or a
// transactional scope (WithTx), without depending on *sqlx.DB directly.
// Interface rather than concrete type so tests can inject a fake instead of
// standing up a real database.
type Transactor interface {
	DB() Querier
	WithTx(ctx context.Context, fn func(q Querier) error) error
}

// SQLTransactor is Transactor backed by a real *sqlx.DB/Postgres connection.
type SQLTransactor struct {
	db *sqlx.DB
}

func NewTransactor(db *sqlx.DB) Transactor {
	return &SQLTransactor{db: db}
}

func (t *SQLTransactor) WithTx(ctx context.Context, fn func(q Querier) error) error {
	tx, err := t.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
		}
	}()

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (t *SQLTransactor) DB() Querier {
	return t.db
}
