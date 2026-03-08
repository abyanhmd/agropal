package transaction

import "agropal/internal/response"

var (
	ErrNotFound response.ErrorKey = "TRANSACTION.NOT_FOUND"

	ErrInvalidID          response.ErrorKey = "TRANSACTION.INVALID_ID"
	ErrInvalidType        response.ErrorKey = "TRANSACTION.INVALID_TYPE"
	ErrInvalidAmount      response.ErrorKey = "TRANSACTION.INVALID_AMOUNT"
	ErrInvalidCurrency    response.ErrorKey = "TRANSACTION.INVALID_CURRENCY"
	ErrInvalidDescription response.ErrorKey = "TRANSACTION.INVALID_DESCRIPTION"
	ErrInvalidProductID   response.ErrorKey = "TRANSACTION.INVALID_PRODUCT_ID"

	ErrCreateFailed response.ErrorKey = "TRANSACTION.CREATE_FAILED"
	ErrUpdateFailed response.ErrorKey = "TRANSACTION.UPDATE_FAILED"
	ErrDeleteFailed response.ErrorKey = "TRANSACTION.DELETE_FAILED"
)
