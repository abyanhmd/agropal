package farm

import "agropal/internal/response"

var (
	ErrNotFound response.ErrorKey = "FARM.NOT_FOUND"

	ErrInvalidID     response.ErrorKey = "FARM.INVALID_ID"
	ErrInvalidName   response.ErrorKey = "FARM.INVALID_NAME"
	ErrInvalidUserID response.ErrorKey = "FARM.INVALID_USER_ID"

	ErrCreateFailed response.ErrorKey = "FARM.CREATE_FAILED"
	ErrUpdateFailed response.ErrorKey = "FARM.UPDATE_FAILED"
	ErrDeleteFailed response.ErrorKey = "FARM.DELETE_FAILED"
)
