package middleware

import (
	"agropal/internal/infra/crypto"
	"agropal/pkg"
	"net/http"
)

type HandlerMiddleware func(handler http.Handler) http.Handler

type Middleware struct {
	logger *pkg.Logger
	jwt    *crypto.JWTProvider
}

func NewMiddleware(
	l *pkg.Logger,
	jwt *crypto.JWTProvider,
) *Middleware {
	return &Middleware{
		logger: l,
		jwt:    jwt,
	}
}

func Chain(handler http.Handler, xs ...HandlerMiddleware) http.Handler {
	for _, m := range xs {
		handler = m(handler)
	}
	return handler
}
