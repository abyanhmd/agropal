package email

import (
	"agropal/internal/infra/queue"
	"fmt"
)

// purpose subjects mirror the verificationPurpose values set in
// internal/app/auth/service.go ("register", "login", "reset_password");
// anything else falls back to a generic subject/body below.
var verificationSubjects = map[string]string{
	"register":       "Verify your Agropal account",
	"login":          "Your Agropal login code",
	"reset_password": "Reset your Agropal password",
}

var verificationIntros = map[string]string{
	"register":       "Use the code below to verify your new Agropal account.",
	"login":          "Use the code below to finish signing in to Agropal.",
	"reset_password": "Use the code below to reset your Agropal password.",
}

// RenderVerificationEmail builds the subject/body pair for a queue.EmailVerificationTask.
func RenderVerificationEmail(task queue.EmailVerificationTask) (subject, body string) {
	subject, ok := verificationSubjects[task.Purpose]
	if !ok {
		subject = "Your Agropal verification code"
	}

	intro, ok := verificationIntros[task.Purpose]
	if !ok {
		intro = "Use the code below to continue."
	}

	body = fmt.Sprintf(
		"%s\n\nCode: %s\nExpires at: %s\n\nIf you didn't request this, you can safely ignore this email.",
		intro,
		task.Code,
		task.ExpiresAt.Local().Format("2006-01-02 15:04:05 MST"),
	)

	return subject, body
}
