package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	rttrace "ai-go-chi-starter/internal/runtime/tracing"
)

func TestTraceAddsTraceparent(t *testing.T) {
	handler := Trace(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		trace, ok := rttrace.TraceFromContext(r.Context())
		if !ok || !trace.Valid() {
			t.Fatal("trace missing from context")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Header().Get("Traceparent") == "" {
		t.Fatal("Traceparent header missing")
	}
}
