package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimiter(t *testing.T) {
	t.Run("Allows Requests Within Burst", func(t *testing.T) {
		mw := newTestMiddleware()

		called := false
		handler := mw.RateLimiter(1, 3)(okHandler(&called))

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		if !called {
			t.Error("expected next handler to be called")
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("Rejects Requests Beyond Burst", func(t *testing.T) {
		mw := newTestMiddleware()

		callCount := 0
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.WriteHeader(http.StatusOK)
		})

		// rps close to 0 with a burst of 2: the first 2 requests consume the
		// full burst, and the 3rd arrives before the bucket can refill.
		handler := mw.RateLimiter(1, 2)(next)

		var lastCode int
		for i := 0; i < 3; i++ {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
			lastCode = rec.Code
		}

		if callCount != 2 {
			t.Errorf("expected next handler to be called exactly 2 times (the burst), got %d", callCount)
		}
		if lastCode != http.StatusTooManyRequests {
			t.Errorf("expected the request beyond the burst to get status 429, got %d", lastCode)
		}
	})

	t.Run("Each Middleware Instance Has Its Own Limiter", func(t *testing.T) {
		mw := newTestMiddleware()

		called := false
		next := okHandler(&called)

		// Two independently constructed middlewares must not share rate
		// limiter state.
		first := mw.RateLimiter(1, 1)(next)
		second := mw.RateLimiter(1, 1)(next)

		rec1 := httptest.NewRecorder()
		first.ServeHTTP(rec1, httptest.NewRequest(http.MethodGet, "/", nil))

		rec2 := httptest.NewRecorder()
		second.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/", nil))

		if rec1.Code != http.StatusOK {
			t.Errorf("expected first middleware's first request to succeed, got %d", rec1.Code)
		}
		if rec2.Code != http.StatusOK {
			t.Errorf("expected second middleware's first request to succeed on its own limiter, got %d", rec2.Code)
		}
	})

	t.Run("Different Client IPs Get Independent Budgets", func(t *testing.T) {
		mw := newTestMiddleware()

		callCount := 0
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.WriteHeader(http.StatusOK)
		})

		// Burst of 1: a second request from the SAME IP would be rejected,
		// but a request from a DIFFERENT IP must still get its own budget.
		handler := mw.RateLimiter(1, 1)(next)

		reqA1 := httptest.NewRequest(http.MethodGet, "/", nil)
		reqA1.RemoteAddr = "203.0.113.10:1111"
		recA1 := httptest.NewRecorder()
		handler.ServeHTTP(recA1, reqA1)

		reqA2 := httptest.NewRequest(http.MethodGet, "/", nil)
		reqA2.RemoteAddr = "203.0.113.10:2222" // same IP, different port
		recA2 := httptest.NewRecorder()
		handler.ServeHTTP(recA2, reqA2)

		reqB1 := httptest.NewRequest(http.MethodGet, "/", nil)
		reqB1.RemoteAddr = "203.0.113.20:1111"
		recB1 := httptest.NewRecorder()
		handler.ServeHTTP(recB1, reqB1)

		if recA1.Code != http.StatusOK {
			t.Errorf("expected first request from IP A to succeed, got %d", recA1.Code)
		}
		if recA2.Code != http.StatusTooManyRequests {
			t.Errorf("expected second request from IP A (same IP, different port) to be rate-limited, got %d", recA2.Code)
		}
		if recB1.Code != http.StatusOK {
			t.Errorf("expected first request from IP B to succeed on its own budget, got %d", recB1.Code)
		}
		if callCount != 2 {
			t.Errorf("expected next handler to be called exactly 2 times (A's first request + B's first request), got %d", callCount)
		}
	})
}

func TestClientIP(t *testing.T) {
	t.Run("Strips The Port From A Host Port RemoteAddr", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.RemoteAddr = "203.0.113.10:54321"

		if got := clientIP(r); got != "203.0.113.10" {
			t.Errorf("clientIP() = %q, want %q", got, "203.0.113.10")
		}
	})

	t.Run("Strips The Port From An IPv6 RemoteAddr", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.RemoteAddr = "[2001:db8::1]:54321"

		if got := clientIP(r); got != "2001:db8::1" {
			t.Errorf("clientIP() = %q, want %q", got, "2001:db8::1")
		}
	})

	t.Run("Falls Back To The Raw RemoteAddr When Not Host Port Form", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.RemoteAddr = "not-a-host-port"

		if got := clientIP(r); got != "not-a-host-port" {
			t.Errorf("clientIP() = %q, want %q", got, "not-a-host-port")
		}
	})
}
