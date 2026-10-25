package location

import (
	"agropal/internal/domain/shared"
)

var (
	ErrNotFound shared.ErrorKey = "LOCATION.NOT_FOUND"

	ErrInvalidID        shared.ErrorKey = "LOCATION.INVALID_ID"
	ErrInvalidFarmID    shared.ErrorKey = "LOCATION.INVALID_FARM_ID"
	ErrInvalidLongitude shared.ErrorKey = "LOCATION.INVALID_LONGITUDE"
	ErrInvalidLatitude  shared.ErrorKey = "LOCATION.INVALID_LATITUDE"

	ErrCreateFailed shared.ErrorKey = "LOCATION.CREATE_FAILED"
	ErrUpdateFailed shared.ErrorKey = "LOCATION.UPDATE_FAILED"
	ErrDeleteFailed shared.ErrorKey = "LOCATION.DELETE_FAILED"

	ErrServiceUnavailable shared.ErrorKey = "LOCATION.SERVICE_UNAVAILABLE"
	ErrDataFetchFailed    shared.ErrorKey = "LOCATION.DATA_FETCH_FAILED"
)
