package weather

import "agropal/internal/response"

var (
	ErrNotFound response.ErrorKey = "WEATHER.NOT_FOUND"

	ErrInvalidID        response.ErrorKey = "WEATHER.INVALID_ID"
	ErrInvalidFarmID    response.ErrorKey = "WEATHER.INVALID_FARM_ID"
	ErrInvalidLongitude response.ErrorKey = "WEATHER.INVALID_LONGITUDE"
	ErrInvalidLatitude  response.ErrorKey = "WEATHER.INVALID_LATITUDE"

	ErrCreateFailed response.ErrorKey = "WEATHER.CREATE_FAILED"
	ErrUpdateFailed response.ErrorKey = "WEATHER.UPDATE_FAILED"
	ErrDeleteFailed response.ErrorKey = "WEATHER.DELETE_FAILED"

	ErrServiceUnavailable response.ErrorKey = "WEATHER.SERVICE_UNAVAILABLE"
	ErrDataFetchFailed    response.ErrorKey = "WEATHER.DATA_FETCH_FAILED"
)
