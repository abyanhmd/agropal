package farm

import (
	"agropal/internal/domain/shared"
)

var (
	ErrNotFound shared.ErrorKey = "FARM.NOT_FOUND"

	ErrInvalidID     shared.ErrorKey = "FARM.INVALID_ID"
	ErrInvalidName   shared.ErrorKey = "FARM.INVALID_NAME"
	ErrInvalidUserID shared.ErrorKey = "FARM.INVALID_USER_ID"

	ErrCreateFailed shared.ErrorKey = "FARM.CREATE_FAILED"
	ErrUpdateFailed shared.ErrorKey = "FARM.UPDATE_FAILED"
	ErrDeleteFailed shared.ErrorKey = "FARM.DELETE_FAILED"
)
