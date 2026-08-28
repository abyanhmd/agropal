package middleware

import (
	"agropal/config"
	"agropal/internal/domain/shared"
	"agropal/internal/utils"
	"agropal/pkg"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

// newTestMiddlewareWithLogger builds a Middleware backed by a real
// (non-production, console-writing) *pkg.Logger, since Logger() logs
// unconditionally and would panic against the zero-valued *pkg.Logger.
func newTestMiddlewareWithLogger(jwt *fakeTokenProvider) *Middleware {
	logger := pkg.NewLogger(config.Config{Stage: "test"})
	return NewMiddleware(
		utils.NewIOHandler(logger, nil, utils.NewValidator()),
		jwt,
	)
}

func TestLogger(t *testing.T) {
	t.Run("PassesRequestThroughAndPreservesStatusCode", func(t *testing.T) {
		mw := newTestMiddlewareWithLogger(nil)

		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusTeapot)
		})

		handler := mw.Logger()(next)

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/anything", nil))

		if !called {
			t.Error("expected next handler to be called")
		}
		if rec.Code != http.StatusTeapot {
			t.Errorf("expected status %d, got %d", http.StatusTeapot, rec.Code)
		}
	})

	t.Run("DefaultsToStatusOKWhenHandlerNeverWritesHeader", func(t *testing.T) {
		mw := newTestMiddlewareWithLogger(nil)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		handler := mw.Logger()(next)

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		if rec.Code != http.StatusOK {
			t.Errorf("expected default status 200, got %d", rec.Code)
		}
	})

	t.Run("InjectsClaimsIntoContextForValidBearerToken", func(t *testing.T) {
		userID := uuid.New()
		claims := shared.ClaimInfo{UserID: userID, Email: "user@example.com", Role: string(config.RoleUser)}
		mw := newTestMiddlewareWithLogger(&fakeTokenProvider{claims: claims})

		var gotCtx context.Context
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotCtx = r.Context()
			w.WriteHeader(http.StatusOK)
		})

		handler := mw.Logger()(next)

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "Bearer valid-token")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, r)

		if got := shared.GetMetadata(gotCtx, shared.KeyUserID); got != userID.String() {
			t.Errorf("expected user id %q injected into context, got %q", userID, got)
		}
	})

	t.Run("DoesNotRejectRequestWhenTokenIsInvalid", func(t *testing.T) {
		mw := newTestMiddlewareWithLogger(&fakeTokenProvider{err: context.DeadlineExceeded})

		called := false
		next := okHandler(&called)
		handler := mw.Logger()(next)

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "Bearer bad-token")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, r)

		if !called {
			t.Error("expected next handler to still be called since Logger never rejects requests")
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("DoesNotAttemptTokenValidationWhenHeaderMissing", func(t *testing.T) {
		// jwt is nil: if Logger tried to validate a token here, it would
		// panic on a nil TokenProvider.
		mw := newTestMiddlewareWithLogger(nil)

		called := false
		handler := mw.Logger()(okHandler(&called))

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		if !called {
			t.Error("expected next handler to be called")
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})
}
