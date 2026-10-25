package main

import (
	"agropal/config"
	"agropal/internal/domain/shared"
	"agropal/internal/infra/email"
	"agropal/internal/infra/queue"
	"agropal/pkg"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	cfg := config.LoadConfig("./config/config.yaml")

	logger := pkg.NewLogger(*cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Info(ctx, shared.System, fmt.Sprintf("Email worker running on %s stage", strings.ToUpper(cfg.Stage)))

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

	var sender email.Sender = email.NewSMTPSender(logger, cfg)
	if cfg.Smtp.Host == "" {
		// Nowhere to deliver to: print the emails so a code can still be read in development
		sender = email.NewLogSender(logger)
		logger.Warn(ctx, shared.System, "SMTP_HOST is not set; emails will be printed to this log instead of sent")
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-quit
		logger.Info(ctx, shared.System, "Shutting down email worker")
		cancel()
	}()

	handler := func(hctx context.Context, body []byte) error {
		var task queue.EmailVerificationTask
		if err := json.Unmarshal(body, &task); err != nil {
			// Malformed payload will never parse on retry either; log and
			// drop it instead of requeuing forever.
			logger.Error(hctx, shared.System, fmt.Sprintf("Failed to decode email task: %s", err.Error()))
			return nil
		}

		subject, message := email.RenderVerificationEmail(task)
		if err := sender.Send(hctx, task.To, subject, message); err != nil {
			logger.Error(hctx, shared.System, fmt.Sprintf("Failed to send verification email to %s: %s", task.To, err.Error()))
			return err
		}

		logger.Info(hctx, shared.System, fmt.Sprintf("Sent %s verification email to %s", task.Purpose, task.To))
		return nil
	}

	if err := mq.ConsumeMessages(ctx, queue.QueueNameEmail, handler); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error(ctx, shared.System, fmt.Sprintf("Email worker stopped: %v", err))
	}
}
