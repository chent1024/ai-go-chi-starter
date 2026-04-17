package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

func TestAccessLogWritesRoutePattern(t *testing.T) {
	var logs bytes.Buffer
	router := chi.NewRouter()
	router.Use(RequestID)
	router.Use(Trace)
	router.Use(AccessLog(slog.New(slog.NewJSONHandler(&logs, nil)), true))
	router.Get("/v1/examples/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/examples/exm_01", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	output := logs.String()
	if !strings.Contains(output, `"route":"/v1/examples/{id}"`) {
		t.Fatalf("access log missing route pattern: %s", output)
	}
}

func TestAccessLogCapturesTimeoutStatus(t *testing.T) {
	var logs bytes.Buffer
	router := chi.NewRouter()
	router.Use(RequestID)
	router.Use(Trace)
	router.Use(AccessLog(slog.New(slog.NewJSONHandler(&logs, nil)), true))
	router.Use(RequestTimeout(10*time.Millisecond, nil, nil))
	router.Get("/slow", func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(logs.String(), `"status":504`) {
		t.Fatalf("access log missing timeout status: %s", logs.String())
	}
}
