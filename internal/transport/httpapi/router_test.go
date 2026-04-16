package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"ai-go-chi-starter/internal/config"
	"ai-go-chi-starter/internal/runtime"
	"ai-go-chi-starter/internal/service/shared"
	"ai-go-chi-starter/internal/transport/httpapi/httpx"
)

func TestRouterHealthz(t *testing.T) {
	handler := NewRouter(RouterOptions{
		Logging: config.LoggingConfig{AccessEnabled: true},
		Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestRouterReadyzReportsDrainState(t *testing.T) {
	drainState := &runtime.DrainState{}
	drainState.BeginDrain()

	handler := NewRouter(RouterOptions{
		Logging:    config.LoggingConfig{AccessEnabled: true},
		Logger:     slog.New(slog.NewTextHandler(io.Discard, nil)),
		DrainState: drainState,
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	req.Header.Set(httpx.RequestIDHeader, "req_02")
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

func TestRouterReadyzFailure(t *testing.T) {
	handler := NewRouter(RouterOptions{
		Logging:      config.LoggingConfig{AccessEnabled: true},
		Logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
		ReadyChecker: readyCheckerFunc(func(ctx context.Context) error { _ = ctx; return errors.New("db down") }),
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	req.Header.Set("X-Request-Id", "req_01")
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
}

type readyCheckerFunc func(context.Context) error

func (f readyCheckerFunc) Ready(ctx context.Context) error {
	return f(ctx)
}
