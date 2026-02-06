package main

import (
	"agropal/config"
	"agropal/internal/database"
	"agropal/internal/domain/shared"
	"agropal/internal/utils"
	"context"
	"fmt"
	"strings"
)

func main() {
	/* Config */
	cfg := config.LoadConfig("./config/config.yaml")

	/* Logger */
	logger := utils.NewLogger(*cfg)
	ctx := context.Background()
	logger.Info(ctx, shared.System, fmt.Sprintf("Running on %s stage", strings.ToUpper(cfg.Stage)))

	/* Database */
	db := database.NewPostgres(cfg)
	defer func() {
		if errClose := db.Close(); errClose != nil {
			logger.Error(ctx, shared.System, fmt.Sprintf("Error closing DB: %v", errClose))
		}
	}()
	logger.Info(ctx, shared.System, "Connected to PostgreSQL Database")
}
