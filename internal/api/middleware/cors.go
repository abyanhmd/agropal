package middleware

import (
	"net/http"
	"strings"
)

// corsAllowMethods is the method list advertised on a preflight response.
const corsAllowMethods = "GET, POST, PATCH, PUT, DELETE, OPTIONS"

// corsDefaultAllowHeaders is advertised on a preflight only when the browser
// sends no Access-Control-Request-Headers to reflect.
const corsDefaultAllowHeaders = "Authorization, Content-Type"

// corsMaxAge caps how long (seconds) a browser may cache a preflight result.
const corsMaxAge = "600"

// CORS returns a wrapper that applies cross-origin resource sharing rules
// for browser clients. It is meant to wrap the whole mux in cmd/api/main.go,
// so it also answers the preflight OPTIONS for the method-pattern auth
// routes that ServeMux would otherwise reject with a bare 405.
//
// allowedOrigins is a list of exact scheme+host+port origins permitted to
// call the API with credentials (the auth flow's rid cookie). The single
// entry "*" reflects whatever Origin the request carries — handy in
// development, but it still enables credentials, so it should not ship. An
// empty list blocks every browser cross-origin call.
//
// Native/mobile apps, curl and server-to-server callers send no Origin
// header and pass straight through regardless of this list.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	allowAny := false
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		o = strings.TrimSpace(o)
		switch {
		case o == "":
		case o == "*":
			allowAny = true
		default:
			allowed[o] = struct{}{}
		}
	}

	originAllowed := func(origin string) bool {
		if allowAny {
			return true
		}
		_, ok := allowed[origin]
		return ok
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// No Origin: not a browser CORS request (native app, curl,
			// server-to-server). Nothing to negotiate.
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			// The response now depends on Origin, so shared caches must key
			// on it.
			w.Header().Add("Vary", "Origin")

			isPreflight := r.Method == http.MethodOptions &&
				r.Header.Get("Access-Control-Request-Method") != ""

			if !originAllowed(origin) {
				// Send no Access-Control-* headers; the browser blocks the
				// caller. A preflight still needs an HTTP response.
				if isPreflight {
					w.WriteHeader(http.StatusNoContent)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if isPreflight {
				w.Header().Add("Vary", "Access-Control-Request-Method")
				w.Header().Add("Vary", "Access-Control-Request-Headers")
				w.Header().Set("Access-Control-Allow-Methods", corsAllowMethods)

				reqHeaders := r.Header.Get("Access-Control-Request-Headers")
				if reqHeaders == "" {
					reqHeaders = corsDefaultAllowHeaders
				}
				w.Header().Set("Access-Control-Allow-Headers", reqHeaders)
				w.Header().Set("Access-Control-Max-Age", corsMaxAge)

				w.WriteHeader(http.StatusNoContent)
				return
			}

			// Let the browser's JS read the correlation id off the response
			// (it is not a CORS-safelisted response header).
			w.Header().Set("Access-Control-Expose-Headers", RequestIDHeader)

			next.ServeHTTP(w, r)
		})
	}
}
