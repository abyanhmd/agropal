package database

import (
	"agropal/config"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// testDB is a direct sqlx connection to the test Postgres container, used by
// transactor_test.go. testConfig carries the same container's connection
// details so postgres_test.go can exercise NewPostgres itself end-to-end.
var (
	testDB     *sqlx.DB
	testConfig *config.Config
)

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
		// No Docker daemon (or some other environment issue) - Postgres
		// integration tests can't run here, but leave testDB/testConfig nil
		// and still run the suite: subtests that don't need Postgres (e.g.
		// the miniredis-backed Redis tests, and the connection-failure
		// panic tests) work fine, and Postgres-dependent subtests skip
		// themselves via `if testDB == nil`. Run this package directly with
		// Docker running to exercise the Postgres-backed tests too.
		fmt.Fprintf(os.Stderr, "database: Postgres-backed tests will be skipped, could not start postgres container: %v\n", err)
		return m.Run()
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "database: failed to terminate postgres container: %v\n", err)
		}
	}()

	host, err := container.Host(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "database: failed to get container host: %v\n", err)
		return 1
	}
	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		fmt.Fprintf(os.Stderr, "database: failed to get mapped port: %v\n", err)
		return 1
	}

	testConfig = &config.Config{}
	testConfig.Database.Host = host
	testConfig.Database.Port = port.Port()
	testConfig.Database.User = "test"
	testConfig.Database.Password = "test"
	testConfig.Database.Dbname = "agropal_test"

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "database: failed to get connection string: %v\n", err)
		return 1
	}

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "database: failed to connect to postgres: %v\n", err)
		return 1
	}
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE transactor_test (id TEXT PRIMARY KEY)`); err != nil {
		fmt.Fprintf(os.Stderr, "database: failed to create test table: %v\n", err)
		return 1
	}

	testDB = db

	return m.Run()
}
