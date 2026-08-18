package middleware

import (
	"agropal/internal/domain/shared"
	"net/http"

	"golang.org/x/time/rate"
)

func (m *Middleware) RateLimiter(rps int, burst int) func(http.Handler) http.Handler {
	limiter := rate.NewLimiter(rate.Limit(rps), burst)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if !limiter.Allow() {
				m.ioHandler.WriteError(r.Context(), w, shared.ErrTooManyRequests, nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
