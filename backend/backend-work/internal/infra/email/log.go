package email

import (
	"agropal/internal/domain/shared"
	"agropal/pkg"
	"context"
	"fmt"
)

// LogSender writes the email to the log instead of delivering it. It stands in for SMTP in
// development when no mail server is configured, so a verification code can be read from the
// worker's output.
type LogSender struct {
	logger *pkg.Logger
}

func NewLogSender(logger *pkg.Logger) *LogSender {
	return &LogSender{logger: logger}
}

func (s *LogSender) Send(ctx context.Context, to, subject, body string) error {
	s.logger.Info(ctx, shared.System, fmt.Sprintf("SMTP_HOST is not set, so this email was not sent. To: %s | Subject: %s | Body: %s", to, subject, body))
	return nil
}
