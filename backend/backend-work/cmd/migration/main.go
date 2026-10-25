package main

import (
	"agropal/config"
	"agropal/internal/domain/shared"
	"agropal/internal/infra/database"
	"agropal/pkg"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

// runGoose drives goose against the migrations directory configured in
// cfg.Path.Migration. args mirrors the goose CLI's own subcommand syntax,
// e.g. ["up"], ["down"], ["status"], ["redo"], ["up-to", "3"].
func runGoose(db *sqlx.DB, cfg *config.Config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no migration command given")
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	command := args[0]
	commandArgs := args[1:]

	return goose.RunContext(context.Background(), command, db.DB, cfg.Path.Migration, commandArgs...)
}

func main() {
	ctx := context.Background()
	cfg := config.LoadConfig("./config/config.yaml")
	if cfg == nil {
		panic("Failed to load config")
	}

	logger := pkg.NewLogger(*cfg)
	logger.Info(ctx, shared.System, fmt.Sprintf("Running on %s stage", strings.ToUpper(cfg.Stage)))

	db := database.NewPostgres(cfg)
	defer func() {
		if dbErr := db.Close(); dbErr != nil {
			logger.Error(ctx, shared.System, fmt.Sprintf("Failed to close database connection: %s", dbErr.Error()))
		}
	}()
	logger.Info(ctx, shared.System, "Connected to PostgreSQL Database")

	args := os.Args[1:]
	if len(args) == 0 {
		logger.Error(ctx, shared.System, "Usage: migration <up|up-to|down|down-to|redo|reset|status|version> [args]")
		os.Exit(1)
	}

	if err := runGoose(db, cfg, args); err != nil {
		logger.Error(ctx, shared.System, err.Error())
		os.Exit(1)
	}

	logger.Info(ctx, shared.System, fmt.Sprintf("Migrate %s done", strings.Join(args, " ")))
}
