package shared

type ErrorKey string

func (e ErrorKey) Error() string {
	return string(e)
}

var (
	ErrInvalidRequest     ErrorKey = "COMMON.INVALID_REQUEST"
	ErrInternalServer     ErrorKey = "COMMON.INTERNAL_SERVER_ERROR"
	ErrUnauthorized       ErrorKey = "COMMON.UNAUTHORIZED"
	ErrForbidden          ErrorKey = "COMMON.FORBIDDEN"
	ErrNotFound           ErrorKey = "COMMON.NOT_FOUND"
	ErrMethodNotAllowed   ErrorKey = "COMMON.METHOD_NOT_ALLOWED"
	ErrConflict           ErrorKey = "COMMON.CONFLICT"
	ErrTooManyRequests    ErrorKey = "COMMON.TOO_MANY_REQUESTS"
	ErrServiceUnavailable ErrorKey = "COMMON.SERVICE_UNAVAILABLE"
	ErrGatewayTimeout     ErrorKey = "COMMON.GATEWAY_TIMEOUT"
	ErrInvalidInput       ErrorKey = "COMMON.INVALID_INPUT"
	ErrDatabaseError      ErrorKey = "COMMON.DATABASE_ERROR"
)
