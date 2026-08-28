package middleware

import (
	"agropal/internal/domain/shared"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// visitorTTL is how long a client's limiter is kept around after its last
// request. Long enough that an active client never loses its accumulated
// burst between requests, short enough that the map doesn't grow unbounded
// under a churn of distinct IPs.
const visitorTTL = 3 * time.Minute

// visitor pairs a client's limiter with the last time it was used, so the
// cleanup loop can evict entries nobody is using anymore.
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter returns a middleware that enforces rps/burst per client IP,
// rather than one shared budget for every caller. Each call to RateLimiter
// owns an independent set of per-IP limiters (and a background goroutine
// that evicts idle ones), so distinct stacks - e.g. public vs. protected in
// cmd/api/main.go - get their own limits.
func (m *Middleware) RateLimiter(rps int, burst int) func(http.Handler) http.Handler {
	var mu sync.Mutex
	visitors := make(map[string]*visitor)

	go cleanupVisitors(&mu, visitors)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)

			mu.Lock()
			v, ok := visitors[ip]
			if !ok {
				v = &visitor{limiter: rate.NewLimiter(rate.Limit(rps), burst)}
				visitors[ip] = v
			}
			v.lastSeen = time.Now()
			limiter := v.limiter
			mu.Unlock()

			if !limiter.Allow() {
				m.ioHandler.WriteError(r.Context(), w, shared.ErrTooManyRequests, nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// clientIP extracts the request's IP address, stripping the ephemeral port
// on r.RemoteAddr so repeat connections from the same client key to the
// same limiter. Falls back to the raw value if it isn't in host:port form.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// cleanupVisitors periodically evicts limiters that haven't been used
// within visitorTTL, so memory doesn't grow without bound as distinct IPs
// come and go over the life of the process.
func cleanupVisitors(mu *sync.Mutex, visitors map[string]*visitor) {
	ticker := time.NewTicker(visitorTTL)
	defer ticker.Stop()

	for range ticker.C {
		cutoff := time.Now().Add(-visitorTTL)

		mu.Lock()
		for ip, v := range visitors {
			if v.lastSeen.Before(cutoff) {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}
