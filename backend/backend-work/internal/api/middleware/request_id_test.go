package middleware

import (
	"agropal/internal/domain/shared"
	"agropal/internal/utils"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestRequestID(t *testing.T) {
	mw := newTestMiddleware()

	// captureID runs next and records the request id it sees in context.
	captureID := func(seen *string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*seen = shared.GetMetadata(r.Context(), shared.KeyRequestID)
			w.WriteHeader(http.StatusOK)
		})
	}

	t.Run("GeneratesAUUIDAndEchoesItOnTheHeader", func(t *testing.T) {
		var seen string
		handler := mw.RequestID(captureID(&seen))

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		if _, err := uuid.Parse(seen); err != nil {
			t.Fatalf("context request id %q is not a UUID: %v", seen, err)
		}
		if got := rec.Header().Get(RequestIDHeader); got != seen {
			t.Errorf("response header %s = %q, want %q (the context id)", RequestIDHeader, got, seen)
		}
	})

	t.Run("IgnoresAnyInboundRequestIDHeader", func(t *testing.T) {
		var seen string
		handler := mw.RequestID(captureID(&seen))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(RequestIDHeader, "client-supplied-id")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if seen == "client-supplied-id" {
			t.Fatal("adopted a client-supplied request id; backend must be the source of truth")
		}
		if _, err := uuid.Parse(seen); err != nil {
			t.Errorf("generated id %q is not a UUID: %v", seen, err)
		}
	})

	t.Run("EachRequestGetsAFreshID", func(t *testing.T) {
		var first, second string
		mw.RequestID(captureID(&first)).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
		mw.RequestID(captureID(&second)).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

		if first == second {
			t.Errorf("two requests shared request id %q", first)
		}
	})

	t.Run("FlowsIntoARecoveredPanicEnvelope", func(t *testing.T) {
		lmw := newTestMiddlewareWithLogger(nil)
		stack := CreateStack(lmw.RequestID, lmw.Recover)
		handler := stack(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			panic("boom")
		}))

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		var envelope utils.StandardEnvelope
		if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
			t.Fatalf("decode body %q: %v", rec.Body.String(), err)
		}
		if envelope.RequestID == "" {
			t.Error("recovered-panic envelope has an empty request_id")
		}
		if envelope.RequestID != rec.Header().Get(RequestIDHeader) {
			t.Errorf("envelope request_id %q != response header %q", envelope.RequestID, rec.Header().Get(RequestIDHeader))
		}
	})
}
