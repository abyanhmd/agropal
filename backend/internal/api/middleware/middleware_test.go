package middleware

import (
	"agropal/config"
	dAuth "agropal/internal/domain/auth"
	"agropal/internal/domain/shared"
	"agropal/internal/utils"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

// fakeTokenProvider is a minimal dAuth.TokenProvider double for tests. When
// err is set, ValidateToken returns it instead of claims.
type fakeTokenProvider struct {
	claims shared.ClaimInfo
	err    error
}

func (f *fakeTokenProvider) GenerateTokenPair(ctx context.Context, info shared.ClaimInfo) (dAuth.TokenPair, error) {
	return dAuth.TokenPair{}, nil
}

func (f *fakeTokenProvider) ValidateToken(ctx context.Context, tokenType shared.TokenType, token string) (shared.ClaimInfo, error) {
	if f.err != nil {
		return shared.ClaimInfo{}, f.err
	}
	return f.claims, nil
}

func newTestMiddleware() *Middleware {
	return newTestMiddlewareWithJWT(nil)
}

func newTestMiddlewareWithJWT(jwt dAuth.TokenProvider) *Middleware {
	return NewMiddleware(utils.NewIOHandler(nil, nil, utils.NewValidator()), jwt, nil)
}

func newTestMiddlewareWithProxies(proxies []string) *Middleware {
	return NewMiddleware(utils.NewIOHandler(nil, nil, utils.NewValidator()), nil, proxies)
}

func requestWithRole(role config.Role) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := shared.WithMetadata(r.Context(), shared.KeyRole, string(role))
	return r.WithContext(ctx)
}

func okHandler(called *bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*called = true
		w.WriteHeader(http.StatusOK)
	})
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
			config.PermDeleteLocation,
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

func TestRequireAuth(t *testing.T) {
	userID := uuid.New()
	validClaims := shared.ClaimInfo{UserID: userID, Email: "user@example.com", Role: string(config.RoleUser)}

	t.Run("AllowsValidBearerTokenAndInjectsClaims", func(t *testing.T) {
		mw := newTestMiddlewareWithJWT(&fakeTokenProvider{claims: validClaims})

		called := false
		var gotCtx context.Context
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			gotCtx = r.Context()
			w.WriteHeader(http.StatusOK)
		})

		handler := mw.RequireAuth(next)

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "Bearer valid-token")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, r)

		if !called {
			t.Fatal("expected next handler to be called")
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
		if got := shared.GetMetadata(gotCtx, shared.KeyRole); got != string(config.RoleUser) {
			t.Errorf("expected role %q injected into context, got %q", config.RoleUser, got)
		}
		if got := shared.GetMetadata(gotCtx, shared.KeyUserID); got != userID.String() {
			t.Errorf("expected user id %q injected into context, got %q", userID, got)
		}
	})

	t.Run("RejectsMissingAuthorizationHeader", func(t *testing.T) {
		mw := newTestMiddlewareWithJWT(&fakeTokenProvider{claims: validClaims})

		called := false
		handler := mw.RequireAuth(okHandler(&called))

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		if called {
			t.Error("expected next handler NOT to be called")
		}
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rec.Code)
		}
	})

	t.Run("RejectsHeaderMissingBearerPrefix", func(t *testing.T) {
		mw := newTestMiddlewareWithJWT(&fakeTokenProvider{claims: validClaims})

		called := false
		handler := mw.RequireAuth(okHandler(&called))

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "Basic somevalue")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, r)

		if called {
			t.Error("expected next handler NOT to be called")
		}
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rec.Code)
		}
	})

	t.Run("RejectsEmptyBearerToken", func(t *testing.T) {
		mw := newTestMiddlewareWithJWT(&fakeTokenProvider{claims: validClaims})

		called := false
		handler := mw.RequireAuth(okHandler(&called))

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "Bearer ")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, r)

		if called {
			t.Error("expected next handler NOT to be called")
		}
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rec.Code)
		}
	})

	t.Run("RejectsInvalidToken", func(t *testing.T) {
		mw := newTestMiddlewareWithJWT(&fakeTokenProvider{err: dAuth.ErrTokenInvalid})

		called := false
		handler := mw.RequireAuth(okHandler(&called))

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "Bearer bad-token")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, r)

		if called {
			t.Error("expected next handler NOT to be called")
		}
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rec.Code)
		}
	})
}

func TestCreateStack(t *testing.T) {
	t.Run("RunsMiddlewaresInDeclaredOrder", func(t *testing.T) {
		var order []string

		track := func(name string) HandlerMiddleware {
			return func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					order = append(order, name)
					next.ServeHTTP(w, r)
				})
			}
		}

		final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "handler")
			w.WriteHeader(http.StatusOK)
		})

		stack := CreateStack(track("a"), track("b"), track("c"))
		handler := stack(final)

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		want := []string{"a", "b", "c", "handler"}
		if len(order) != len(want) {
			t.Fatalf("expected order %v, got %v", want, order)
		}
		for i := range want {
			if order[i] != want[i] {
				t.Fatalf("expected order %v, got %v", want, order)
			}
		}
	})

	t.Run("NoMiddlewaresReturnsHandlerUnchanged", func(t *testing.T) {
		called := false
		final := okHandler(&called)

		stack := CreateStack()
		handler := stack(final)

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		if !called {
			t.Error("expected wrapped handler to be called")
		}
	})
}

func TestChain(t *testing.T) {
	t.Run("AppliesMiddlewaresInDeclaredOrder", func(t *testing.T) {
		var order []string

		track := func(name string) HandlerMiddleware {
			return func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					order = append(order, name)
					next.ServeHTTP(w, r)
				})
			}
		}

		final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "handler")
			w.WriteHeader(http.StatusOK)
		})

		// Chain wraps the handler in list order, so each subsequent
		// middleware becomes the new outermost layer; the last one listed
		// therefore runs first at request time.
		handler := Chain(final, track("a"), track("b"))

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		want := []string{"b", "a", "handler"}
		if len(order) != len(want) {
			t.Fatalf("expected order %v, got %v", want, order)
		}
		for i := range want {
			if order[i] != want[i] {
				t.Fatalf("expected order %v, got %v", want, order)
			}
		}
	})

	t.Run("NoMiddlewaresReturnsHandlerUnchanged", func(t *testing.T) {
		called := false
		final := okHandler(&called)

		handler := Chain(final)

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		if !called {
			t.Error("expected wrapped handler to be called")
		}
	})
}
