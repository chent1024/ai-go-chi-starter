package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ai-go-chi-starter/internal/runtime"
	"ai-go-chi-starter/internal/service/shared"
	"ai-go-chi-starter/internal/transport/httpapi/httpx"
)

func TestDrainRejectsNewRequests(t *testing.T) {
	var state runtime.DrainState
	state.BeginDrain()

	handler := RequestID(Drain(&state)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should not reach handler while draining")
	})))

	req := httptest.NewRequest(http.MethodGet, "/v1/examples", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
	var env httpx.Envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if env.Code != shared.CodeNotReady {
		t.Fatalf("code = %q", env.Code)
	}
}

func TestDrainAllowsHealthzDuringShutdown(t *testing.T) {
	var state runtime.DrainState
	state.BeginDrain()

	handler := Drain(&state)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
}
