package product

import "agropal/internal/response"

var (
	ErrNotFound response.ErrorKey = "PRODUCT.NOT_FOUND"

	ErrInvalidID          response.ErrorKey = "PRODUCT.INVALID_ID"
	ErrInvalidName        response.ErrorKey = "PRODUCT.INVALID_NAME"
	ErrInvalidDescription response.ErrorKey = "PRODUCT.INVALID_DESCRIPTION"
	ErrInvalidImage       response.ErrorKey = "PRODUCT.INVALID_IMAGE"
	ErrInvalidFarmID      response.ErrorKey = "PRODUCT.INVALID_FARM_ID"

	ErrCreateFailed response.ErrorKey = "PRODUCT.CREATE_FAILED"
	ErrUpdateFailed response.ErrorKey = "PRODUCT.UPDATE_FAILED"
	ErrDeleteFailed response.ErrorKey = "PRODUCT.DELETE_FAILED"

	ErrStockNotFound      response.ErrorKey = "PRODUCT.STOCK_NOT_FOUND"
	ErrInvalidStockID     response.ErrorKey = "PRODUCT.INVALID_STOCK_ID"
	ErrInvalidStockType   response.ErrorKey = "PRODUCT.INVALID_STOCK_TYPE"
	ErrInvalidStockAmount response.ErrorKey = "PRODUCT.INVALID_STOCK_AMOUNT"

	ErrPriceNotFound       response.ErrorKey = "PRODUCT.PRICE_NOT_FOUND"
	ErrInvalidPriceID      response.ErrorKey = "PRODUCT.INVALID_PRICE_ID"
	ErrInvalidPriceAmount  response.ErrorKey = "PRODUCT.INVALID_PRICE_AMOUNT"
	ErrInsufficientStock   response.ErrorKey = "PRODUCT.INSUFFICIENT_STOCK"
	ErrInvalidMovementType response.ErrorKey = "PRODUCT.INVALID_MOVEMENT_TYPE"
)
