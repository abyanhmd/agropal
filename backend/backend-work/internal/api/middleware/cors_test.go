package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS(t *testing.T) {
	const allowed = "https://app.example.com"

	newHandler := func(origins []string, nextCalled *bool) http.Handler {
		return CORS(origins)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if nextCalled != nil {
				*nextCalled = true
			}
			w.WriteHeader(http.StatusOK)
		}))
	}

	t.Run("RequestWithNoOriginPassesThroughUntouched", func(t *testing.T) {
		called := false
		rec := httptest.NewRecorder()
		newHandler([]string{allowed}, &called).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		if !called {
			t.Fatal("next handler was not called")
		}
		if rec.Header().Get("Access-Control-Allow-Origin") != "" || rec.Header().Get("Vary") != "" {
			t.Errorf("CORS headers set for a non-browser request: %v", rec.Header())
		}
	})

	t.Run("AllowedOriginGetsHeadersOnASimpleRequest", func(t *testing.T) {
		called := false
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", allowed)
		newHandler([]string{allowed}, &called).ServeHTTP(rec, req)

		if !called {
			t.Fatal("next handler was not called")
		}
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != allowed {
			t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, allowed)
		}
		if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
			t.Errorf("Access-Control-Allow-Credentials = %q, want %q", got, "true")
		}
		if got := rec.Header().Get("Access-Control-Expose-Headers"); got != RequestIDHeader+", "+TotalCountHeader {
			t.Errorf("Access-Control-Expose-Headers = %q, want %q", got, RequestIDHeader+", "+TotalCountHeader)
		}
		if got := rec.Header().Get("Vary"); got != "Origin" {
			t.Errorf("Vary = %q, want %q", got, "Origin")
		}
	})

	t.Run("DisallowedOriginGetsNoHeadersButRequestStillRuns", func(t *testing.T) {
		called := false
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "https://evil.example.com")
		newHandler([]string{allowed}, &called).ServeHTTP(rec, req)

		if !called {
			t.Fatal("next handler should still run; the browser enforces the block")
		}
		if rec.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Errorf("Access-Control-Allow-Origin set for a disallowed origin: %q", rec.Header().Get("Access-Control-Allow-Origin"))
		}
		if got := rec.Header().Get("Vary"); got != "Origin" {
			t.Errorf("Vary = %q, want %q", got, "Origin")
		}
	})

	t.Run("PreflightFromAllowedOriginIsAnswered", func(t *testing.T) {
		called := false
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodOptions, "/api/v1/auth/login", nil)
		req.Header.Set("Origin", allowed)
		req.Header.Set("Access-Control-Request-Method", http.MethodPost)
		req.Header.Set("Access-Control-Request-Headers", "authorization, content-type")
		newHandler([]string{allowed}, &called).ServeHTTP(rec, req)

		if called {
			t.Fatal("preflight must not reach the next handler")
		}
		if rec.Code != http.StatusNoContent {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != allowed {
			t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, allowed)
		}
		if got := rec.Header().Get("Access-Control-Allow-Methods"); got != corsAllowMethods {
			t.Errorf("Access-Control-Allow-Methods = %q, want %q", got, corsAllowMethods)
		}
		if got := rec.Header().Get("Access-Control-Allow-Headers"); got != "authorization, content-type" {
			t.Errorf("Access-Control-Allow-Headers = %q, want the reflected request headers", got)
		}
		if got := rec.Header().Get("Access-Control-Max-Age"); got != corsMaxAge {
			t.Errorf("Access-Control-Max-Age = %q, want %q", got, corsMaxAge)
		}
	})

	t.Run("PreflightWithNoRequestHeadersGetsTheDefaultAllowList", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		req.Header.Set("Origin", allowed)
		req.Header.Set("Access-Control-Request-Method", http.MethodPost)
		newHandler([]string{allowed}, nil).ServeHTTP(rec, req)

		if got := rec.Header().Get("Access-Control-Allow-Headers"); got != corsDefaultAllowHeaders {
			t.Errorf("Access-Control-Allow-Headers = %q, want %q", got, corsDefaultAllowHeaders)
		}
	})

	t.Run("PreflightFromDisallowedOriginGetsNoCORSHeaders", func(t *testing.T) {
		called := false
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		req.Header.Set("Origin", "https://evil.example.com")
		req.Header.Set("Access-Control-Request-Method", http.MethodPost)
		newHandler([]string{allowed}, &called).ServeHTTP(rec, req)

		if called {
			t.Fatal("preflight must not reach the next handler")
		}
		if rec.Code != http.StatusNoContent {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
		if rec.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Errorf("Access-Control-Allow-Origin set for a disallowed origin")
		}
	})

	t.Run("WildcardReflectsAnyOrigin", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "https://anything.example.net")
		newHandler([]string{"*"}, nil).ServeHTTP(rec, req)

		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://anything.example.net" {
			t.Errorf("Access-Control-Allow-Origin = %q, want the reflected origin", got)
		}
		if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
			t.Errorf("Access-Control-Allow-Credentials = %q, want %q", got, "true")
		}
	})

	t.Run("EmptyAllowlistBlocksEveryBrowserOrigin", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", allowed)
		newHandler(nil, nil).ServeHTTP(rec, req)

		if rec.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Errorf("Access-Control-Allow-Origin set with an empty allowlist")
		}
	})
}
