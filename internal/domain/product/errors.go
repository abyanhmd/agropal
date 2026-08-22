package product

import (
	"agropal/internal/domain/shared"
)

var (
	ErrNotFound shared.ErrorKey = "PRODUCT.NOT_FOUND"

	ErrInvalidID          shared.ErrorKey = "PRODUCT.INVALID_ID"
	ErrInvalidName        shared.ErrorKey = "PRODUCT.INVALID_NAME"
	ErrInvalidDescription shared.ErrorKey = "PRODUCT.INVALID_DESCRIPTION"
	ErrInvalidImage       shared.ErrorKey = "PRODUCT.INVALID_IMAGE"
	ErrInvalidFarmID      shared.ErrorKey = "PRODUCT.INVALID_FARM_ID"

	ErrCreateFailed shared.ErrorKey = "PRODUCT.CREATE_FAILED"
	ErrUpdateFailed shared.ErrorKey = "PRODUCT.UPDATE_FAILED"
	ErrDeleteFailed shared.ErrorKey = "PRODUCT.DELETE_FAILED"

	ErrStockNotFound      shared.ErrorKey = "PRODUCT.STOCK_NOT_FOUND"
	ErrInvalidStockID     shared.ErrorKey = "PRODUCT.INVALID_STOCK_ID"
	ErrInvalidStockType   shared.ErrorKey = "PRODUCT.INVALID_STOCK_TYPE"
	ErrInvalidStockAmount shared.ErrorKey = "PRODUCT.INVALID_STOCK_AMOUNT"

	ErrPriceNotFound       shared.ErrorKey = "PRODUCT.PRICE_NOT_FOUND"
	ErrInvalidPriceID      shared.ErrorKey = "PRODUCT.INVALID_PRICE_ID"
	ErrInvalidPriceAmount  shared.ErrorKey = "PRODUCT.INVALID_PRICE_AMOUNT"
	ErrInsufficientStock   shared.ErrorKey = "PRODUCT.INSUFFICIENT_STOCK"
	ErrInvalidMovementType shared.ErrorKey = "PRODUCT.INVALID_MOVEMENT_TYPE"

	ErrInvalidDate shared.ErrorKey = "PRODUCT.INVALID_DATE"
)
