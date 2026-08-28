package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func handlerWithStatus(status int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}
}

func TestSetupV1Routes(t *testing.T) {
	t.Run("Routes Request To Matching Path And Method", func(t *testing.T) {
		mux := http.NewServeMux()
		routes := []Routes{
			{Path: "/ping", Method: http.MethodGet, Handler: handlerWithStatus(http.StatusOK)},
		}

		SetupV1Routes(routes, CreateStack(), mux)

		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil))

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("Runs Route Specific Middleware Then Global Stack", func(t *testing.T) {
		var order []string

		track := func(name string) HandlerMiddleware {
			return func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					order = append(order, name)
					next.ServeHTTP(w, r)
				})
			}
		}

		mux := http.NewServeMux()
		routes := []Routes{
			{
				Path:        "/ping",
				Method:      http.MethodGet,
				Handler:     handlerWithStatus(http.StatusOK),
				Middlewares: []HandlerMiddleware{track("route")},
			},
		}
		stack := CreateStack(track("global"))

		SetupV1Routes(routes, stack, mux)

		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil))

		want := []string{"global", "route"}
		if len(order) != len(want) || order[0] != want[0] || order[1] != want[1] {
			t.Errorf("expected middleware order %v, got %v", want, order)
		}
	})

	t.Run("Different Methods On Same Path Route Independently", func(t *testing.T) {
		mux := http.NewServeMux()
		routes := []Routes{
			{Path: "/resource", Method: http.MethodGet, Handler: handlerWithStatus(http.StatusOK)},
			{Path: "/resource", Method: http.MethodPost, Handler: handlerWithStatus(http.StatusCreated)},
		}

		SetupV1Routes(routes, CreateStack(), mux)

		getRec := httptest.NewRecorder()
		mux.ServeHTTP(getRec, httptest.NewRequest(http.MethodGet, "/api/v1/resource", nil))
		if getRec.Code != http.StatusOK {
			t.Errorf("expected GET status 200, got %d", getRec.Code)
		}

		postRec := httptest.NewRecorder()
		mux.ServeHTTP(postRec, httptest.NewRequest(http.MethodPost, "/api/v1/resource", nil))
		if postRec.Code != http.StatusCreated {
			t.Errorf("expected POST status 201, got %d", postRec.Code)
		}
	})

	t.Run("Unregistered Method On Known Path Returns 405", func(t *testing.T) {
		mux := http.NewServeMux()
		routes := []Routes{
			{Path: "/resource", Method: http.MethodGet, Handler: handlerWithStatus(http.StatusOK)},
		}

		SetupV1Routes(routes, CreateStack(), mux)

		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/v1/resource", nil))

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405, got %d", rec.Code)
		}
	})

	t.Run("Options Request Returns No Content", func(t *testing.T) {
		mux := http.NewServeMux()
		routes := []Routes{
			{Path: "/resource", Method: http.MethodGet, Handler: handlerWithStatus(http.StatusOK)},
		}

		SetupV1Routes(routes, CreateStack(), mux)

		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodOptions, "/api/v1/resource", nil))

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", rec.Code)
		}
	})
}
