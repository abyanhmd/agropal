package queue

import (
	"agropal/config"
	"agropal/pkg"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// testConfig points at a live RabbitMQ broker for integration tests;
// testLogger is a real (non-production, console-writing) *pkg.Logger since
// RabbitMQ logs unconditionally on the connect/publish/consume paths.
var (
	testConfig *config.Config
	testLogger *pkg.Logger
)

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

// runTests is split out from TestMain so `defer`/cleanup actually runs
// before the process exits (TestMain itself must call os.Exit directly,
// which skips deferred calls in that function).
func runTests(m *testing.M) int {
	ctx := context.Background()

	testLogger = pkg.NewLogger(config.Config{Stage: "test"})

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "rabbitmq:3.13-alpine",
			ExposedPorts: []string{"5672/tcp"},
			WaitingFor:   wait.ForLog("Server startup complete").WithStartupTimeout(60 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		// No Docker daemon (or some other environment issue) - RabbitMQ
		// integration tests can't run here, but leave testConfig nil and
		// still run the suite: subtests that don't need a live broker (pure
		// functions, nil-guard branches) work fine, and broker-dependent
		// subtests skip themselves via `if testConfig == nil`. Run this
		// package directly with Docker running to exercise those too.
		fmt.Fprintf(os.Stderr, "queue: broker-backed tests will be skipped, could not start rabbitmq container: %v\n", err)
		return m.Run()
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "queue: failed to terminate rabbitmq container: %v\n", err)
		}
	}()

	host, err := container.Host(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "queue: failed to get container host: %v\n", err)
		return 1
	}
	port, err := container.MappedPort(ctx, "5672/tcp")
	if err != nil {
		fmt.Fprintf(os.Stderr, "queue: failed to get mapped port: %v\n", err)
		return 1
	}

	testConfig = &config.Config{}
	testConfig.RabbitMQ.Address = fmt.Sprintf("%s:%s", host, port.Port())
	testConfig.RabbitMQ.Username = "guest"
	testConfig.RabbitMQ.Password = "guest"

	return m.Run()
}
