package request

// VerificationRequest backs the register/verify and login/verify endpoints:
// confirming a one-time code sent to Email.
type VerificationRequest struct {
	Email *string `json:"email"`
	Code  *string `json:"code"`
}

// PasswordResetRequest backs the reset-password/request endpoint.
type PasswordResetRequest struct {
	Email *string `json:"email"`
}

// ResendVerificationRequest backs the {purpose}/resend-code endpoint.
type ResendVerificationRequest struct {
	Email *string `json:"email"`
}

// PasswordResetConfirmRequest backs the reset-password/confirm endpoint.
type PasswordResetConfirmRequest struct {
	Email       *string `json:"email"`
	Code        *string `json:"code"`
	NewPassword *string `json:"new_password"`
}
