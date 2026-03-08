package user

import "agropal/internal/response"

var (
	ErrNotFound       response.ErrorKey = "USER.NOT_FOUND"
	ErrEmailExists    response.ErrorKey = "USER.EMAIL_EXISTS"
	ErrUsernameExists response.ErrorKey = "USER.USERNAME_EXISTS"

	ErrInvalidUserID   response.ErrorKey = "USER.INVALID_USER_ID"
	ErrInvalidName     response.ErrorKey = "USER.INVALID_NAME"
	ErrInvalidEmail    response.ErrorKey = "USER.INVALID_EMAIL"
	ErrInvalidUsername response.ErrorKey = "USER.INVALID_USERNAME"
	ErrInvalidAvatar   response.ErrorKey = "USER.INVALID_AVATAR"
	ErrInvalidPassword response.ErrorKey = "USER.INVALID_PASSWORD"
	ErrInvalidRole     response.ErrorKey = "USER.INVALID_ROLE"

	ErrUpdateFailed response.ErrorKey = "USER.UPDATE_FAILED"
	ErrDeleteFailed response.ErrorKey = "USER.DELETE_FAILED"
)
