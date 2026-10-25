package middleware

import (
	"agropal/internal/domain/shared"
	"fmt"
	"net/http"
	"runtime/debug"
)

// Recover converts a panic in any downstream handler (or in a middleware
// layered below it) into a logged stack trace plus the standard 500 error
// envelope, instead of letting net/http drop the connection with an empty
// response and no request_id. It should sit at the very top of every stack
// built with CreateStack.
func (m *Middleware) Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			// net/http's sentinel for a handler that deliberately aborts the
			// response; re-panic so the server treats it as intended rather
			// than logging it as a crash.
			if rec == http.ErrAbortHandler {
				panic(rec)
			}

			ctx := r.Context()
			m.ioHandler.Logger.Error(ctx, shared.System, fmt.Sprintf(
				"panic recovered: %v\n%s", rec, debug.Stack(),
			))

			// If the handler already wrote a response before panicking this
			// is a best-effort second write (net/http logs a superfluous
			// WriteHeader warning); there is nothing better we can do.
			m.ioHandler.WriteError(ctx, w, shared.ErrInternalServer, nil)
		}()

		next.ServeHTTP(w, r)
	})
}
