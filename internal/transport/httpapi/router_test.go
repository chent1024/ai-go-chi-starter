package httpapi

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"ai-go-chi-starter/internal/config"
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
