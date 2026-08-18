package middleware

import (
	"agropal/internal/domain/shared"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// statusRecorder wraps http.ResponseWriter to capture the status code
// written by the next handler so it can be included in the access log.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Logger returns a middleware that logs each request's method, path,
// status code and duration. It best-effort decodes the bearer token, if
// present, so authenticated requests are logged with the caller's
// identity, but unlike RequireAuth it never rejects the request, since
// this middleware runs ahead of routes that may be public.
func (m *Middleware) Logger() HandlerMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now().UTC()

			ctx := r.Context()
			var claims *shared.ClaimInfo = nil

			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				tokenParts := strings.SplitN(authHeader, " ", 2)
				if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
					accessClaims, err := m.jwt.ValidateToken(ctx, shared.TypeAccess, tokenParts[1])
					if err != nil {
						m.ioHandler.Logger.Error(ctx, shared.System, err.Error())
					} else {
						claims = &accessClaims
					}
				}
			}
			ctx = shared.SaveContextFromRequest(r, claims)
			r = r.WithContext(ctx)

			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)

			m.ioHandler.Logger.Info(ctx, shared.System, fmt.Sprintf(
				"%s %s -> %d (%s)", r.Method, r.URL.Path, rec.status, time.Since(start),
			))
		})
	}
}
