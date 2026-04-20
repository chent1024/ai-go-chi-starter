package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ai-go-chi-starter/internal/service/shared"
	"ai-go-chi-starter/internal/transport/httpapi/httpx"
	apimetrics "ai-go-chi-starter/internal/transport/httpapi/metrics"
)

func TestRequestTimeoutWritesGatewayTimeout(t *testing.T) {
	handler := RequestTimeout(10*time.Millisecond, nil, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	req.Header.Set(httpx.RequestIDHeader, "req_timeout")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf("status = %d", rec.Code)
	}
	var env httpx.Envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if env.Code != shared.CodeRequestTimeout || env.RequestID != "req_timeout" {
		t.Fatalf("unexpected envelope = %+v", env)
	}
}

func TestRequestTimeoutAllowsFastHandler(t *testing.T) {
	handler := RequestTimeout(2*time.Second, nil, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			t.Fatal("context unexpectedly canceled")
		default:
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/fast", nil))

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestRequestTimeoutPreservesCanceledContext(t *testing.T) {
	handler := RequestTimeout(50*time.Millisecond, nil, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodGet, "/canceled", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestRequestTimeoutDropsLateWritesAndReturnsGatewayTimeout(t *testing.T) {
	handler := RequestTimeout(10*time.Millisecond, nil, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
		w.WriteHeader(http.StatusNoContent)
		_, _ = w.Write([]byte("late"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/slow-write", nil)
	req.Header.Set(httpx.RequestIDHeader, "req_late")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"code":"REQUEST_TIMEOUT"`) {
		t.Fatalf("response body should contain timeout envelope: %s", rec.Body.String())
	}
}

func TestRequestTimeoutRecordsLateWriteMetricAndLog(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelInfo}))
	metrics := apimetrics.New(apimetrics.BuildInfo{Service: "api", Version: "dev", Commit: "unknown", BuildTime: "unknown"})

	handler := RequestTimeout(10*time.Millisecond, logger, metrics)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
		w.WriteHeader(http.StatusNoContent)
		_, _ = w.Write([]byte("late"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	req.Header.Set(httpx.RequestIDHeader, "req_late_log")
	req = req.WithContext(httpx.WithLogger(req.Context(), logger.With("request_id", "req_late_log")))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	metricsRec := httptest.NewRecorder()
	metrics.ServeHTTP(metricsRec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if !strings.Contains(metricsRec.Body.String(), `http_request_timeout_late_write_total{route="unmatched"} 2`) {
		t.Fatalf("metrics body = %s", metricsRec.Body.String())
	}
	for _, want := range []string{`"request_id":"req_late_log"`, `"late_write_count":2`} {
		if !strings.Contains(logs.String(), want) {
			t.Fatalf("log output missing %q: %s", want, logs.String())
		}
	}
}
