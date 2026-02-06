package middleware

import (
	"agropal/internal/utils"
	"net/http"
)

type HandlerMiddleware func(handler http.Handler) http.Handler

type Middleware struct {
	JWT    *utils.JWT
	Logger *utils.Logger
}

func NewMiddleware(
	jwt *utils.JWT,
	logger *utils.Logger,
) *Middleware {
	return &Middleware{
		JWT:    jwt,
		Logger: logger,
	}
}

func Chain(handler http.Handler, xs ...HandlerMiddleware) http.Handler {
	for _, m := range xs {
		handler = m(handler)
	}
	return handler
}
