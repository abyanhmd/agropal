package middleware

import (
	"agropal/internal/domain/shared"
	"net/http"

	"github.com/google/uuid"
)

// RequestIDHeader is the response header the generated correlation id is
// echoed on.
const RequestIDHeader = "X-Request-Id"

// RequestID gives every request a server-generated correlation id (a UUID).
// The backend is the sole source of truth: any inbound X-Request-Id is
// ignored. The id is stored in the request context under
// shared.KeyRequestID (read by the response envelopes and the access log)
// and echoed on the response header so a client can tie a call to a
// server-side log line. It must run outermost so the id is available to
// every layer below it, including Recover.
func (m *Middleware) RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.NewString()

		w.Header().Set(RequestIDHeader, id)
		ctx := shared.WithMetadata(r.Context(), shared.KeyRequestID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
