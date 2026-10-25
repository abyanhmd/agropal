package middleware

import (
	"agropal/internal/domain/shared"
	"agropal/internal/utils"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecover(t *testing.T) {
	t.Run("PanicBecomesA500Envelope", func(t *testing.T) {
		mw := newTestMiddlewareWithLogger(nil)

		next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			panic("boom")
		})
		handler := mw.Recover(next)

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
		}
		var envelope utils.StandardEnvelope
		if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
			t.Fatalf("failed to decode body %q: %v", rec.Body.String(), err)
		}
		if envelope.Info.Key != shared.ErrInternalServer.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, shared.ErrInternalServer.Error())
		}
	})

	t.Run("NoPanicPassesThroughUntouched", func(t *testing.T) {
		mw := newTestMiddlewareWithLogger(nil)

		called := false
		handler := mw.Recover(okHandler(&called))

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		if !called {
			t.Error("expected next handler to be called")
		}
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("ErrAbortHandlerIsRePanicked", func(t *testing.T) {
		mw := newTestMiddlewareWithLogger(nil)

		next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			panic(http.ErrAbortHandler)
		})
		handler := mw.Recover(next)

		defer func() {
			if rec := recover(); rec != http.ErrAbortHandler {
				t.Errorf("recovered %v, want http.ErrAbortHandler to propagate", rec)
			}
		}()

		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
		t.Fatal("expected ServeHTTP to re-panic with http.ErrAbortHandler")
	})
}
