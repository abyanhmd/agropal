package email

import (
	"agropal/config"
	"agropal/pkg"
	"context"
	"fmt"
	"net/smtp"
)

// Sender delivers a rendered email to a single recipient.
type Sender interface {
	Send(ctx context.Context, to, subject, body string) error
}

// SMTPSender sends mail through a standard SMTP relay using PLAIN auth.
type SMTPSender struct {
	logger *pkg.Logger
	config *config.Config
}

func NewSMTPSender(logger *pkg.Logger, cfg *config.Config) *SMTPSender {
	return &SMTPSender{
		logger: logger,
		config: cfg,
	}
}

// Send dials the configured SMTP host and delivers a plain-text message.
// net/smtp has no context support, so ctx is accepted for interface symmetry
// with callers but doesn't bound the dial/send itself.
func (s *SMTPSender) Send(_ context.Context, to, subject, body string) error {
	host := s.config.Smtp.Host
	addr := fmt.Sprintf("%s:%s", host, s.config.Smtp.Port)
	from := s.config.Smtp.From

	// A relay that needs no login (a local mail catcher in development) is used without one
	var auth smtp.Auth
	if s.config.Smtp.Username != "" {
		auth = smtp.PlainAuth("", s.config.Smtp.Username, s.config.Secret.Smtp, host)
	}

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n%s\r\n",
		from, to, subject, body,
	)

	if err := smtp.SendMail(addr, auth, from, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("failed to send email to %s: %w", to, err)
	}

	return nil
}
