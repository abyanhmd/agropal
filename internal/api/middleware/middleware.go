package middleware

import (
	"agropal/config"
	dAuth "agropal/internal/domain/auth"
	"agropal/internal/domain/shared"
	"agropal/internal/utils"
	"net"
	"net/http"
	"strings"
)

type HandlerMiddleware func(handler http.Handler) http.Handler

type Middleware struct {
	ioHandler      *utils.IOHandler
	jwt            dAuth.TokenProvider
	trustedProxies []*net.IPNet
}

func NewMiddleware(
	ioHandler *utils.IOHandler,
	jwt dAuth.TokenProvider,
	trustedProxies []string,
) *Middleware {
	return &Middleware{
		ioHandler:      ioHandler,
		jwt:            jwt,
		trustedProxies: parseCIDRs(trustedProxies),
	}
}

// parseCIDRs turns a list of CIDR ranges or bare IPs into networks. A bare
// IP becomes a single-address network. Unparseable entries are skipped so a
// typo in config disables that entry rather than crashing the server.
func parseCIDRs(entries []string) []*net.IPNet {
	nets := make([]*net.IPNet, 0, len(entries))
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if _, ipNet, err := net.ParseCIDR(entry); err == nil {
			nets = append(nets, ipNet)
			continue
		}
		if ip := net.ParseIP(entry); ip != nil {
			bits := 32
			if ip.To4() == nil {
				bits = 128
			}
			nets = append(nets, &net.IPNet{IP: ip, Mask: net.CIDRMask(bits, bits)})
		}
	}
	return nets
}

// isTrustedProxy reports whether ip (string form) is inside one of the
// configured trusted-proxy ranges.
func (m *Middleware) isTrustedProxy(ip string) bool {
	parsed := net.ParseIP(strings.TrimSpace(ip))
	if parsed == nil {
		return false
	}
	for _, n := range m.trustedProxies {
		if n.Contains(parsed) {
			return true
		}
	}
	return false
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
		ctx = shared.WithMetadata(ctx, shared.KeyIPAddress, m.clientIP(r))
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
