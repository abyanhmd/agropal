package middleware

import (
	"agropal/config"
	"agropal/internal/domain/shared"
	"agropal/internal/utils"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestMiddleware() *Middleware {
	return NewMiddleware(utils.NewIOHandler(nil, nil, utils.NewValidator()), nil)
}

func requestWithRole(role config.Role) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := shared.WithMetadata(r.Context(), shared.KeyRole, string(role))
	return r.WithContext(ctx)
}

func TestRequirePermission(t *testing.T) {
	mw := newTestMiddleware()

	t.Run("AllowsWhenRoleHasPermission", func(t *testing.T) {
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})

		handler := mw.RequirePermission(config.PermCreateFarm)(next)

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, requestWithRole(config.RoleUser))

		if !called {
			t.Error("expected next handler to be called")
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("DeniesWhenRoleLacksPermission", func(t *testing.T) {
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})

		handler := mw.RequirePermission(config.PermCreateFarm)(next)

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, requestWithRole(config.RoleGuest))

		if called {
			t.Error("expected next handler NOT to be called")
		}
		if rec.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", rec.Code)
		}
	})

	t.Run("DeniesWhenRoleMissingFromContext", func(t *testing.T) {
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})

		handler := mw.RequirePermission(config.PermCreateFarm)(next)

		rec := httptest.NewRecorder()
		// No role injected into context, e.g. RequireAuth was skipped.
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		if called {
			t.Error("expected next handler NOT to be called")
		}
		if rec.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", rec.Code)
		}
	})

	t.Run("AdminAllowedEverywhere", func(t *testing.T) {
		perms := []config.Permission{
			config.PermDeleteAllTransactions,
			config.PermGetManyUsers,
			config.PermDeleteWeatherLocation,
		}

		for _, perm := range perms {
			called := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			})

			handler := mw.RequirePermission(perm)(next)

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, requestWithRole(config.RoleAdmin))

			if !called {
				t.Errorf("expected RoleAdmin to be allowed for %q", perm)
			}
			if rec.Code != http.StatusOK {
				t.Errorf("expected status 200 for %q, got %d", perm, rec.Code)
			}
		}
	})
}
