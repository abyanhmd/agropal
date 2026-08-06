package auth

import "agropal/internal/domain/shared"

var (
	ErrAlreadyRegistered    shared.ErrorKey = "AUTH.ALREADY_REGISTERED"
	ErrInvalidEmail         shared.ErrorKey = "AUTH.INVALID_EMAIL"
	ErrInvalidPassword      shared.ErrorKey = "AUTH.INVALID_PASSWORD"
	ErrInvalidUsername      shared.ErrorKey = "AUTH.INVALID_USERNAME"
	ErrInvalidName          shared.ErrorKey = "AUTH.INVALID_NAME"
	ErrInvalidAvatar        shared.ErrorKey = "AUTH.INVALID_AVATAR"
	ErrWeakPassword         shared.ErrorKey = "AUTH.WEAK_PASSWORD"
	ErrPasswordMismatch     shared.ErrorKey = "AUTH.PASSWORD_MISMATCH"
	ErrPasswordTooShort     shared.ErrorKey = "AUTH.PASSWORD_TOO_SHORT"
	ErrPasswordTooLong      shared.ErrorKey = "AUTH.PASSWORD_TOO_LONG"
	ErrPasswordHasNoLower   shared.ErrorKey = "AUTH.PASSWORD_HAS_NO_LOWER"
	ErrPasswordHasNoUpper   shared.ErrorKey = "AUTH.PASSWORD_HAS_NO_UPPER"
	ErrPasswordHasNoNumber  shared.ErrorKey = "AUTH.PASSWORD_HAS_NO_NUMBER"
	ErrPasswordHasNoSpecial shared.ErrorKey = "AUTH.PASSWORD_HAS_NO_SPECIAL"
	ErrInvalidInviteCode    shared.ErrorKey = "AUTH.INVALID_INVITE_CODE"

	ErrInvalidCredentials shared.ErrorKey = "AUTH.INVALID_CREDENTIALS"
	ErrAccountDisabled    shared.ErrorKey = "AUTH.ACCOUNT_DISABLED"
	ErrAccountLocked      shared.ErrorKey = "AUTH.ACCOUNT_LOCKED"
	ErrAccountNotFound    shared.ErrorKey = "AUTH.ACCOUNT_NOT_FOUND"
	ErrLoginRequired      shared.ErrorKey = "AUTH.LOGIN_REQUIRED"

	ErrTokenExpired   shared.ErrorKey = "AUTH.TOKEN_EXPIRED"
	ErrTokenInvalid   shared.ErrorKey = "AUTH.TOKEN_INVALID"
	ErrTokenRevoked   shared.ErrorKey = "AUTH.TOKEN_REVOKED"
	ErrSessionExpired shared.ErrorKey = "AUTH.SESSION_EXPIRED"
	ErrSessionInvalid shared.ErrorKey = "AUTH.SESSION_INVALID"

	ErrUnauthorized      shared.ErrorKey = "AUTH.UNAUTHORIZED"
	ErrForbidden         shared.ErrorKey = "AUTH.FORBIDDEN"
	ErrInsufficientScope shared.ErrorKey = "AUTH.INSUFFICIENT_SCOPE"
	ErrRoleRequired      shared.ErrorKey = "AUTH.ROLE_REQUIRED"

	ErrPasswordResetRequired shared.ErrorKey = "AUTH.PASSWORD_RESET_REQUIRED"
	ErrPasswordResetExpired  shared.ErrorKey = "AUTH.PASSWORD_RESET_EXPIRED"
	ErrPasswordResetInvalid  shared.ErrorKey = "AUTH.PASSWORD_RESET_INVALID"
	ErrOldPasswordMismatch   shared.ErrorKey = "AUTH.OLD_PASSWORD_MISMATCH"

	ErrTooManyAttempts shared.ErrorKey = "AUTH.TOO_MANY_ATTEMPTS"
	ErrRateLimited     shared.ErrorKey = "AUTH.RATE_LIMITED"
	ErrSuspiciousLogin shared.ErrorKey = "AUTH.SUSPICIOUS_LOGIN"
)
