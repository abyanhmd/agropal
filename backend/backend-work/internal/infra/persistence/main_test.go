package persistence

import (
	"agropal/config"
	"agropal/pkg"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// testDB is shared by every test in this package: standing up a fresh
// Postgres container per test would make the suite painfully slow, and
// each repository test is expected to generate its own randomized IDs so
// rows from different tests never collide.
var testDB *sqlx.DB

// testLogger is shared by every test in this package. Each domain's
// NewXRepository(testLogger) constructor is called directly in that
// domain's own test file - there's no single generic repository type to
// share here (RepositoryImpl/NewRepository in persistence.go isn't wired
// to any of the concrete SQL implementations).
var testLogger *pkg.Logger

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

// runTests is split out from TestMain so `defer`/cleanup actually runs
// before the process exits (TestMain itself must call os.Exit directly,
// which skips deferred calls in that function).
func runTests(m *testing.M) int {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:18-alpine",
		tcpostgres.WithDatabase("agropal_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp"),
		),
	)
	if err != nil {
		// No Docker daemon (or some other environment issue) - integration
		// tests can't run here. Skip the whole package rather than failing
		// the build, so `go test ./...` still succeeds on machines without
		// Docker; run this package directly with Docker running to actually
		// exercise it.
		fmt.Fprintf(os.Stderr, "persistence: skipping integration tests, could not start postgres container: %v\n", err)
		return 0
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "persistence: failed to terminate postgres container: %v\n", err)
		}
	}()

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "persistence: failed to get connection string: %v\n", err)
		return 1
	}

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "persistence: failed to connect to postgres: %v\n", err)
		return 1
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		fmt.Fprintf(os.Stderr, "persistence: failed to set goose dialect: %v\n", err)
		return 1
	}
	// Migrations live at backend/internal/migration; this package is at
	// backend/internal/infra/persistence, two directories down.
	if err := goose.Up(db.DB, "../../migration"); err != nil {
		fmt.Fprintf(os.Stderr, "persistence: failed to run migrations: %v\n", err)
		return 1
	}

	testDB = db
	testLogger = pkg.NewLogger(config.Config{})

	return m.Run()
}
