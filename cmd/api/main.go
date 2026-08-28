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
	"agropal/internal/infra/queue"
	"agropal/internal/utils"
	"agropal/pkg"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
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

	redis := database.NewRedis(cfg)

	/* Message Queue */
	mq, err := queue.NewRabbitMQ(cfg, logger)
	if err != nil {
		logger.Fatal(ctx, shared.System, fmt.Sprintf("Failed to connect to RabbitMQ: %v", err))
	}
	defer func() {
		if errClose := mq.Close(); errClose != nil {
			logger.Error(ctx, shared.System, fmt.Sprintf("Error closing RabbitMQ: %v", errClose))
		}
	}()
	logger.Info(ctx, shared.System, "Connected to RabbitMQ")

	/* Migration */

	/* Toolkit */
	enc := crypto.NewEncryption(logger, cfg)
	toolkit := app.NewToolkit(logger, cfg, enc, db, redis, mq)
	container := app.NewContainer(toolkit)

	/* HTTP */
	ioHandler := utils.NewIOHandler(logger, nil, utils.NewValidator())
	mw := middleware.NewMiddleware(ioHandler, container.JWT, cfg.Server.TrustedProxies)

	mux := http.NewServeMux()

	// RequestID runs outermost so the correlation id reaches every layer
	// below (Recover included); Recover then catches panics in Logger/
	// RateLimiter; Logger runs on every request; RequireAuth only guards
	// protected routes.
	publicStack := middleware.CreateStack(mw.RequestID, mw.Recover, mw.Logger(), mw.RateLimiter(5, 10))
	protectedStack := middleware.CreateStack(mw.RequestID, mw.Recover, mw.Logger(), mw.RateLimiter(5, 10), mw.RequireAuth)

	// Auth endpoints are unauthenticated by definition.
	authHandler := interfaceAuth.NewHandler(ioHandler, container.Auth.Service)
	mux.Handle("POST /api/v1/auth/register", publicStack(http.HandlerFunc(authHandler.Register)))
	mux.Handle("POST /api/v1/auth/register/verify", publicStack(http.HandlerFunc(authHandler.VerifyRegister)))
	mux.Handle("POST /api/v1/auth/{purpose}/resend-code", publicStack(http.HandlerFunc(authHandler.ResendVerificationCode)))
	mux.Handle("POST /api/v1/auth/login", publicStack(http.HandlerFunc(authHandler.Login)))
	mux.Handle("POST /api/v1/auth/login/verify", publicStack(http.HandlerFunc(authHandler.VerifyLogin)))
	mux.Handle("POST /api/v1/auth/refresh", publicStack(http.HandlerFunc(authHandler.Refresh)))
	mux.Handle("POST /api/v1/auth/logout", publicStack(http.HandlerFunc(authHandler.Logout)))
	mux.Handle("POST /api/v1/auth/reset-password/request", publicStack(http.HandlerFunc(authHandler.RequestPasswordReset)))
	mux.Handle("POST /api/v1/auth/reset-password/confirm", publicStack(http.HandlerFunc(authHandler.ConfirmPasswordReset)))

	apiFarm.NewRouter(mux, protectedStack, mw, apiFarm.NewHandler(ioHandler, container.Farm.Service))
	apiWeather.NewRouter(mux, protectedStack, mw, apiWeather.NewHandler(ioHandler, container.Weather.Service))
	apiProduct.NewRouter(mux, protectedStack, mw, apiProduct.NewHandler(ioHandler, container.Product.Service))
	apiTransaction.NewRouter(mux, protectedStack, mw, apiTransaction.NewHandler(ioHandler, container.Transaction.Service))
	apiUser.NewRouter(mux, protectedStack, mw, apiUser.NewHandler(ioHandler, container.User.Service))
	apiPost.NewRouter(mux, protectedStack, mw, apiPost.NewHandler(ioHandler, container.Post.Service))

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr: addr,
		// CORS wraps the whole mux so it also covers the preflight OPTIONS
		// for the method-pattern auth routes, which ServeMux would otherwise
		// answer with a bare 405.
		Handler: middleware.CORS(cfg.Cors.AllowedOrigins)(mux),
	}

	go func() {
		logger.Info(ctx, shared.System, fmt.Sprintf("Listening on %s", addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error(ctx, shared.System, fmt.Sprintf("Server error: %v", err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info(ctx, shared.System, "Shutting down server")

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error(ctx, shared.System, fmt.Sprintf("Graceful shutdown failed: %v", err))
	} else {
		logger.Info(ctx, shared.System, "Server stopped gracefully")
	}
}
