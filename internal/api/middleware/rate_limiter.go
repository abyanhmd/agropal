package middleware

import (
	"agropal/internal/domain/shared"
	"net"
	"net/http"
	"strings"
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
			ip := m.clientIP(r)

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

// clientIP returns the address to key rate limiting (and logging) on.
//
// The connection peer address (r.RemoteAddr, minus its ephemeral port) is
// used directly, unless that peer is a configured trusted proxy: in that
// case X-Forwarded-For is walked right-to-left and the first entry that is
// not itself a trusted proxy is returned, i.e. the address the outermost
// trusted proxy saw the request come from. Without this, every request
// behind a reverse proxy keys to the proxy's IP and shares one budget.
func (m *Middleware) clientIP(r *http.Request) string {
	peer := hostOnly(r.RemoteAddr)

	if !m.isTrustedProxy(peer) {
		return peer
	}

	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded == "" {
		return peer
	}

	parts := strings.Split(forwarded, ",")
	for i := len(parts) - 1; i >= 0; i-- {
		candidate := strings.TrimSpace(parts[i])
		if candidate == "" {
			continue
		}
		if !m.isTrustedProxy(candidate) {
			return candidate
		}
	}

	return peer
}

// hostOnly strips the ephemeral port from a host:port address so repeat
// connections from the same client key to the same limiter. It returns the
// raw value if it isn't in host:port form.
func hostOnly(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
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
