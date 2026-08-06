package user

import (
	"agropal/internal/domain/shared"
)

var (
	ErrNotFound       shared.ErrorKey = "USER.NOT_FOUND"
	ErrEmailExists    shared.ErrorKey = "USER.EMAIL_EXISTS"
	ErrUsernameExists shared.ErrorKey = "USER.USERNAME_EXISTS"

	ErrInvalidUserID   shared.ErrorKey = "USER.INVALID_USER_ID"
	ErrInvalidName     shared.ErrorKey = "USER.INVALID_NAME"
	ErrInvalidEmail    shared.ErrorKey = "USER.INVALID_EMAIL"
	ErrInvalidUsername shared.ErrorKey = "USER.INVALID_USERNAME"
	ErrInvalidAvatar   shared.ErrorKey = "USER.INVALID_AVATAR"
	ErrInvalidPassword shared.ErrorKey = "USER.INVALID_PASSWORD"
	ErrInvalidRole     shared.ErrorKey = "USER.INVALID_ROLE"

	ErrUpdateFailed shared.ErrorKey = "USER.UPDATE_FAILED"
	ErrDeleteFailed shared.ErrorKey = "USER.DELETE_FAILED"
)
