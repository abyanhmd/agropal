package auth

import "agropal/internal/response"

var (
	ErrAlreadyRegistered response.ErrorKey = "AUTH.ALREADY_REGISTERED"
	ErrInvalidEmail      response.ErrorKey = "AUTH.INVALID_EMAIL"
	ErrInvalidPassword   response.ErrorKey = "AUTH.INVALID_PASSWORD"
	ErrWeakPassword      response.ErrorKey = "AUTH.WEAK_PASSWORD"
	ErrInvalidInviteCode response.ErrorKey = "AUTH.INVALID_INVITE_CODE"

	ErrInvalidCredentials response.ErrorKey = "AUTH.INVALID_CREDENTIALS"
	ErrAccountDisabled    response.ErrorKey = "AUTH.ACCOUNT_DISABLED"
	ErrAccountLocked      response.ErrorKey = "AUTH.ACCOUNT_LOCKED"
	ErrAccountNotFound    response.ErrorKey = "AUTH.ACCOUNT_NOT_FOUND"
	ErrLoginRequired      response.ErrorKey = "AUTH.LOGIN_REQUIRED"

	ErrTokenExpired   response.ErrorKey = "AUTH.TOKEN_EXPIRED"
	ErrTokenInvalid   response.ErrorKey = "AUTH.TOKEN_INVALID"
	ErrTokenRevoked   response.ErrorKey = "AUTH.TOKEN_REVOKED"
	ErrSessionExpired response.ErrorKey = "AUTH.SESSION_EXPIRED"
	ErrSessionInvalid response.ErrorKey = "AUTH.SESSION_INVALID"

	ErrUnauthorized      response.ErrorKey = "AUTH.UNAUTHORIZED"
	ErrForbidden         response.ErrorKey = "AUTH.FORBIDDEN"
	ErrInsufficientScope response.ErrorKey = "AUTH.INSUFFICIENT_SCOPE"
	ErrRoleRequired      response.ErrorKey = "AUTH.ROLE_REQUIRED"

	ErrPasswordResetRequired response.ErrorKey = "AUTH.PASSWORD_RESET_REQUIRED"
	ErrPasswordResetExpired  response.ErrorKey = "AUTH.PASSWORD_RESET_EXPIRED"
	ErrPasswordResetInvalid  response.ErrorKey = "AUTH.PASSWORD_RESET_INVALID"
	ErrOldPasswordMismatch   response.ErrorKey = "AUTH.OLD_PASSWORD_MISMATCH"

	ErrTooManyAttempts response.ErrorKey = "AUTH.TOO_MANY_ATTEMPTS"
	ErrRateLimited     response.ErrorKey = "AUTH.RATE_LIMITED"
	ErrSuspiciousLogin response.ErrorKey = "AUTH.SUSPICIOUS_LOGIN"
)
