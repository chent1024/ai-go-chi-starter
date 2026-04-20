package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-go-chi-starter/internal/service/shared"
	apidrain "ai-go-chi-starter/internal/transport/httpapi/drain"
	"ai-go-chi-starter/internal/transport/httpapi/httpx"
	apimetrics "ai-go-chi-starter/internal/transport/httpapi/metrics"
)

func TestRouterHealthz(t *testing.T) {
	handler := NewRouter(RouterOptions{
		AccessLogEnabled: true,
		Logger:           slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestRouterReadyzReportsDrainState(t *testing.T) {
	drainState := &apidrain.State{}
	drainState.BeginDrain()

	handler := NewRouter(RouterOptions{
		AccessLogEnabled: true,
		Logger:           slog.New(slog.NewTextHandler(io.Discard, nil)),
		DrainState:       drainState,
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
		AccessLogEnabled: true,
		Logger:           slog.New(slog.NewTextHandler(io.Discard, nil)),
		ReadyChecker:     readyCheckerFunc(func(ctx context.Context) error { _ = ctx; return errors.New("db down") }),
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	req.Header.Set("X-Request-Id", "req_01")
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestRouterVersion(t *testing.T) {
	handler := NewRouter(RouterOptions{
		AccessLogEnabled: true,
		Logger:           slog.New(slog.NewTextHandler(io.Discard, nil)),
		BuildInfo: apimetrics.BuildInfo{
			Service:   "api",
			Version:   "1.2.3",
			Commit:    "abc123",
			BuildTime: "2026-04-16T12:00:00Z",
		},
	})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/version", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"version":"1.2.3"`) {
		t.Fatalf("version body = %s", rec.Body.String())
	}
}

func TestRouterMetrics(t *testing.T) {
	metrics := apimetrics.New(apimetrics.BuildInfo{Service: "api", Version: "dev", Commit: "unknown", BuildTime: "unknown"})
	handler := NewRouter(RouterOptions{
		AccessLogEnabled: true,
		Logger:           slog.New(slog.NewTextHandler(io.Discard, nil)),
		Metrics:          metrics,
	})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("healthz status = %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("metrics status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `http_requests_total{route="/healthz",method="GET",status="200"} 1`) {
		t.Fatalf("metrics body = %s", rec.Body.String())
	}
	for _, want := range []string{"http_requests_in_flight 1", "process_uptime_seconds ", `http_request_duration_ms_max{route="/healthz",method="GET",status="200"}`} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Fatalf("metrics body missing %q: %s", want, rec.Body.String())
		}
	}
}

func TestRouterAppliesSecurityHeaders(t *testing.T) {
	handler := NewRouter(RouterOptions{
		AccessLogEnabled: true,
		Logger:           slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q", got)
	}
	if got := rec.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Fatalf("X-Frame-Options = %q", got)
	}
	if got := rec.Header().Get("Referrer-Policy"); got != "no-referrer" {
		t.Fatalf("Referrer-Policy = %q", got)
	}
}

type readyCheckerFunc func(context.Context) error

func (f readyCheckerFunc) Ready(ctx context.Context) error {
	return f(ctx)
}
