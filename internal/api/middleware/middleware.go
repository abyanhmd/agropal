package middleware

import (
	"agropal/config"
	dAuth "agropal/internal/domain/auth"
	"agropal/internal/domain/shared"
	"agropal/internal/utils"
	"net/http"
	"strings"
)

type HandlerMiddleware func(handler http.Handler) http.Handler

type Middleware struct {
	ioHandler *utils.IOHandler
	jwt       dAuth.TokenProvider
}

func NewMiddleware(
	ioHandler *utils.IOHandler,
	jwt dAuth.TokenProvider,
) *Middleware {
	return &Middleware{
		ioHandler: ioHandler,
		jwt:       jwt,
	}
}

func CreateStack(xs ...HandlerMiddleware) HandlerMiddleware {
	return func(next http.Handler) http.Handler {
		for i := len(xs) - 1; i >= 0; i-- {
			x := xs[i]
			next = x(next)
		}
		return next
	}
}

func Chain(handler http.Handler, xs ...HandlerMiddleware) http.Handler {
	for _, m := range xs {
		handler = m(handler)
	}
	return handler
}

// RequireAuth validates the Authorization bearer token and, on success,
// injects the caller's claims into the request context via the same
// metadata keys the app layer reads with shared.ExtractUserID /
// shared.GetMetadata.
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		authHeader := r.Header.Get("Authorization")
		token, ok := strings.CutPrefix(authHeader, "Bearer ")
		if authHeader == "" || !ok || token == "" {
			m.ioHandler.WriteError(ctx, w, dAuth.ErrUnauthorized, nil)
			return
		}

		claims, err := m.jwt.ValidateToken(ctx, shared.TypeAccess, token)
		if err != nil {
			m.ioHandler.WriteError(ctx, w, dAuth.ErrTokenInvalid, nil)
			return
		}

		ctx = shared.SaveContextFromRequest(r, &claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequirePermission returns a middleware that denies the request with 403
// unless the caller's role (injected into context by RequireAuth) has been
// granted perm. It must run after RequireAuth so the role is present in
// context.
func (m *Middleware) RequirePermission(perm config.Permission) HandlerMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			role := config.Role(shared.GetMetadata(ctx, shared.KeyRole))
			if !config.HasPermission(role, perm) {
				m.ioHandler.WriteError(ctx, w, shared.ErrForbidden, nil)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
