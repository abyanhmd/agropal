package weather

import (
	"agropal/internal/domain/shared"
)

var (
	ErrNotFound shared.ErrorKey = "WEATHER.NOT_FOUND"

	ErrInvalidID        shared.ErrorKey = "WEATHER.INVALID_ID"
	ErrInvalidFarmID    shared.ErrorKey = "WEATHER.INVALID_FARM_ID"
	ErrInvalidLongitude shared.ErrorKey = "WEATHER.INVALID_LONGITUDE"
	ErrInvalidLatitude  shared.ErrorKey = "WEATHER.INVALID_LATITUDE"

	ErrCreateFailed shared.ErrorKey = "WEATHER.CREATE_FAILED"
	ErrUpdateFailed shared.ErrorKey = "WEATHER.UPDATE_FAILED"
	ErrDeleteFailed shared.ErrorKey = "WEATHER.DELETE_FAILED"

	ErrServiceUnavailable shared.ErrorKey = "WEATHER.SERVICE_UNAVAILABLE"
	ErrDataFetchFailed    shared.ErrorKey = "WEATHER.DATA_FETCH_FAILED"
)
