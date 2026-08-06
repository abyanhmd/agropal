package main

import (
	"agropal/config"
	interfaceAuth "agropal/internal/api/auth"
	apiFarm "agropal/internal/api/farm"
	"agropal/internal/api/middleware"
	apiPost "agropal/internal/api/post"
	apiProduct "agropal/internal/api/product"
	apiTransaction "agropal/internal/api/transaction"
	apiUser "agropal/internal/api/user"
	apiWeather "agropal/internal/api/weather"
	"agropal/internal/app"
	"agropal/internal/domain/shared"
	"agropal/internal/infra/crypto"
	"agropal/internal/infra/database"
	"agropal/internal/utils"
	"agropal/pkg"
	"context"
	"fmt"
	"net/http"
	"strings"
)

func main() {
	/* Config */
	cfg := config.LoadConfig("./config/config.yaml")

	/* Logger */
	logger := pkg.NewLogger(*cfg)
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

	/* Migration */

	/* Toolkit */
	enc := crypto.NewEncryption(logger, cfg)
	toolkit := app.NewToolkit(logger, cfg, enc, db)
	container := app.NewContainer(toolkit)

	/* HTTP */
	ioHandler := utils.NewIOHandler(logger, nil)
	mw := middleware.NewMiddleware(ioHandler, container.JWT)

	mux := http.NewServeMux()

	// Auth endpoints are unauthenticated by definition.
	authHandler := interfaceAuth.NewHandler(ioHandler, container.Auth.Service)
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", authHandler.Logout)

	apiFarm.NewRouter(mux, mw.RequireAuth, apiFarm.NewHandler(ioHandler, container.Farm.Service))
	apiWeather.NewRouter(mux, mw.RequireAuth, apiWeather.NewHandler(ioHandler, container.Weather.Service))
	apiProduct.NewRouter(mux, mw.RequireAuth, apiProduct.NewHandler(ioHandler, container.Product.Service))
	apiTransaction.NewRouter(mux, mw.RequireAuth, apiTransaction.NewHandler(ioHandler, container.Transaction.Service))
	apiUser.NewRouter(mux, mw.RequireAuth, apiUser.NewHandler(ioHandler, container.User.Service))
	apiPost.NewRouter(mux, mw.RequireAuth, apiPost.NewHandler(ioHandler, container.Post.Service))

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Info(ctx, shared.System, fmt.Sprintf("Listening on %s", addr))
	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Error(ctx, shared.System, fmt.Sprintf("Server stopped: %v", err))
	}
}
