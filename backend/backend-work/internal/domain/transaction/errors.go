package transaction

import (
	"agropal/internal/domain/shared"
)

var (
	ErrNotFound shared.ErrorKey = "TRANSACTION.NOT_FOUND"

	ErrInvalidID          shared.ErrorKey = "TRANSACTION.INVALID_ID"
	ErrInvalidType        shared.ErrorKey = "TRANSACTION.INVALID_TYPE"
	ErrInvalidAmount      shared.ErrorKey = "TRANSACTION.INVALID_AMOUNT"
	ErrInvalidCurrency    shared.ErrorKey = "TRANSACTION.INVALID_CURRENCY"
	ErrInvalidDescription shared.ErrorKey = "TRANSACTION.INVALID_DESCRIPTION"
	ErrInvalidProductID   shared.ErrorKey = "TRANSACTION.INVALID_PRODUCT_ID"

	ErrCreateFailed shared.ErrorKey = "TRANSACTION.CREATE_FAILED"
	ErrUpdateFailed shared.ErrorKey = "TRANSACTION.UPDATE_FAILED"
	ErrDeleteFailed shared.ErrorKey = "TRANSACTION.DELETE_FAILED"
)
