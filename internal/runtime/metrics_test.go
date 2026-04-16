package runtime

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMetricsServeHTTPIncludesInFlightAndLatencyMax(t *testing.T) {
	metrics := NewMetrics(BuildInfo{Service: "api", Version: "dev", Commit: "unknown", BuildTime: "unknown"})
	metrics.IncInFlight()
	metrics.ObserveHTTPRequest("/healthz", http.MethodGet, http.StatusOK, 125*time.Millisecond)

	rec := httptest.NewRecorder()
	metrics.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	body := rec.Body.String()
	for _, want := range []string{
		"http_requests_in_flight 1",
		"process_uptime_seconds ",
		`http_request_duration_ms_max{route="/healthz",method="GET",status="200"} 125`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("metrics body missing %q: %s", want, body)
		}
	}
}

func TestMetricsServeHTTPIncludesLateWriteSeries(t *testing.T) {
	metrics := NewMetrics(BuildInfo{Service: "api", Version: "dev", Commit: "unknown", BuildTime: "unknown"})
	metrics.ObserveTimeoutLateWrite("/slow", 2)

	rec := httptest.NewRecorder()
	metrics.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if !strings.Contains(rec.Body.String(), `http_request_timeout_late_write_total{route="/slow"} 2`) {
		t.Fatalf("metrics body = %s", rec.Body.String())
	}
}
