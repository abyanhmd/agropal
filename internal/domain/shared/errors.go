package shared

import "agropal/internal/response"

var (
	ErrInvalidRequest     response.ErrorKey = "COMMON.INVALID_REQUEST"
	ErrInternalServer     response.ErrorKey = "COMMON.INTERNAL_SERVER_ERROR"
	ErrUnauthorized       response.ErrorKey = "COMMON.UNAUTHORIZED"
	ErrForbidden          response.ErrorKey = "COMMON.FORBIDDEN"
	ErrNotFound           response.ErrorKey = "COMMON.NOT_FOUND"
	ErrMethodNotAllowed   response.ErrorKey = "COMMON.METHOD_NOT_ALLOWED"
	ErrConflict           response.ErrorKey = "COMMON.CONFLICT"
	ErrTooManyRequests    response.ErrorKey = "COMMON.TOO_MANY_REQUESTS"
	ErrServiceUnavailable response.ErrorKey = "COMMON.SERVICE_UNAVAILABLE"
	ErrGatewayTimeout     response.ErrorKey = "COMMON.GATEWAY_TIMEOUT"
	ErrInvalidInput       response.ErrorKey = "COMMON.INVALID_INPUT"
	ErrDatabaseError      response.ErrorKey = "COMMON.DATABASE_ERROR"
)
