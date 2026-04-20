package client

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	rtlog "ai-go-chi-starter/internal/runtime/logging"
	rttrace "ai-go-chi-starter/internal/runtime/tracing"
)

func TestConfigureTransportAppliesOutboundSettings(t *testing.T) {
	options := Options{
		Timeout:               45 * time.Second,
		MaxIdleConns:          120,
		MaxIdleConnsPerHost:   24,
		IdleConnTimeout:       95 * time.Second,
		TLSHandshakeTimeout:   7 * time.Second,
		ResponseHeaderTimeout: 18 * time.Second,
		ExpectContinueTimeout: 2 * time.Second,
	}

	base := &http.Transport{}
	client := NewHTTPClient(&http.Client{Transport: base}, discardLogger(), options, "svc", "dep")

	if client.Timeout != options.Timeout {
		t.Fatalf("client timeout = %v, want %v", client.Timeout, options.Timeout)
	}

	transport, ok := configureTransport(base, options).(*http.Transport)
	if !ok {
		t.Fatal("configureTransport() did not return *http.Transport")
	}
	if transport.MaxIdleConns != options.MaxIdleConns {
		t.Fatalf("MaxIdleConns = %d, want %d", transport.MaxIdleConns, options.MaxIdleConns)
	}
	if transport.MaxIdleConnsPerHost != options.MaxIdleConnsPerHost {
		t.Fatalf("MaxIdleConnsPerHost = %d, want %d", transport.MaxIdleConnsPerHost, options.MaxIdleConnsPerHost)
	}
	if transport.IdleConnTimeout != options.IdleConnTimeout {
		t.Fatalf("IdleConnTimeout = %v, want %v", transport.IdleConnTimeout, options.IdleConnTimeout)
	}
	if transport.TLSHandshakeTimeout != options.TLSHandshakeTimeout {
		t.Fatalf("TLSHandshakeTimeout = %v, want %v", transport.TLSHandshakeTimeout, options.TLSHandshakeTimeout)
	}
	if transport.ResponseHeaderTimeout != options.ResponseHeaderTimeout {
		t.Fatalf("ResponseHeaderTimeout = %v, want %v", transport.ResponseHeaderTimeout, options.ResponseHeaderTimeout)
	}
	if transport.ExpectContinueTimeout != options.ExpectContinueTimeout {
		t.Fatalf("ExpectContinueTimeout = %v, want %v", transport.ExpectContinueTimeout, options.ExpectContinueTimeout)
	}
}

func TestNewHTTPClientInjectsTraceparent(t *testing.T) {
	trace := rttrace.NewRootTrace()
	var captured *http.Request
	base := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		captured = req
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Body:       http.NoBody,
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})

	client := NewHTTPClient(
		&http.Client{Transport: base},
		discardLogger(),
		defaultOutboundConfig(),
		"svc",
		"dep",
	)

	req, err := http.NewRequestWithContext(
		rttrace.ContextWithTrace(context.Background(), trace),
		http.MethodGet,
		"https://example.com/ping",
		nil,
	)
	if err != nil {
		t.Fatalf("NewRequestWithContext() error = %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	_ = resp.Body.Close()

	if captured == nil {
		t.Fatal("round tripper did not receive request")
	}
	childTrace, ok := rttrace.ParseTraceparent(captured.Header.Get(traceparentHeader))
	if !ok {
		t.Fatalf("Traceparent = %q, want valid traceparent", captured.Header.Get(traceparentHeader))
	}
	if childTrace.TraceID != trace.TraceID {
		t.Fatalf("trace id = %q, want %q", childTrace.TraceID, trace.TraceID)
	}
	if childTrace.ParentSpanID != "" {
		t.Fatalf("parsed parent span id should be empty, got %q", childTrace.ParentSpanID)
	}
	if childTrace.SpanID == trace.SpanID {
		t.Fatalf("child span id = %q, want new child span", childTrace.SpanID)
	}
}

func TestConfigureTransportKeepsCustomRoundTripper(t *testing.T) {
	base := staticRoundTripper{}

	configured := configureTransport(base, defaultOutboundConfig())
	if _, ok := configured.(staticRoundTripper); !ok {
		t.Fatal("configureTransport() should leave non-transport round trippers unchanged")
	}
}

func TestNewHTTPClientLogsOutboundWithRequestContext(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
	trace := rttrace.NewRootTrace()
	base := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Body:       http.NoBody,
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})

	client := NewHTTPClient(
		&http.Client{Transport: base},
		logger,
		defaultOutboundConfig(),
		"svc",
		"dep",
	)

	ctx := rttrace.ContextWithRequestID(rttrace.ContextWithTrace(context.Background(), trace), "req_01")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.com/ping", nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext() error = %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	_ = resp.Body.Close()

	output := logs.String()
	for _, want := range []string{`"kind":"outbound"`, `"request_id":"req_01"`, `"trace_id":"`, `"target":"dep"`} {
		if !strings.Contains(output, want) {
			t.Fatalf("log output missing %q: %s", want, output)
		}
	}
}

func TestNewHTTPClientSkipsOutboundSuccessLogWhenDisabled(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelInfo}))
	base := staticRoundTripper{}

	client := NewHTTPClient(
		&http.Client{Transport: base},
		logger,
		disabledOutboundConfig(),
		"svc",
		"dep",
	)

	req, err := http.NewRequest(http.MethodGet, "https://example.com/ping", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	_ = resp.Body.Close()

	if strings.Contains(logs.String(), "outbound request completed") {
		t.Fatalf("unexpected success log: %s", logs.String())
	}
}

func TestNewHTTPClientLogsOutboundFailureWhenDisabled(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelInfo}))
	base := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("upstream unavailable")
	})

	client := NewHTTPClient(
		&http.Client{Transport: base},
		logger,
		disabledOutboundConfig(),
		"svc",
		"dep",
	)

	req, err := http.NewRequest(http.MethodGet, "https://example.com/ping", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}

	_, err = client.Do(req)
	if err == nil {
		t.Fatal("Do() error = nil")
	}

	output := logs.String()
	if !strings.Contains(output, "outbound request failed") {
		t.Fatalf("missing failure log: %s", output)
	}
}

func TestNewHTTPClientPreservesCanceledContext(t *testing.T) {
	base := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return nil, req.Context().Err()
	})

	client := NewHTTPClient(
		&http.Client{Transport: base},
		discardLogger(),
		defaultOutboundConfig(),
		"svc",
		"dep",
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.com/ping", nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext() error = %v", err)
	}

	_, err = client.Do(req)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Do() error = %v, want context.Canceled", err)
	}
}

func defaultOutboundConfig() Options {
	return Options{
		Timeout:               30 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
		ExpectContinueTimeout: time.Second,
		OutboundLogging: rtlog.OutboundOptions{
			Enabled: true,
			Level:   "info",
		},
	}
}

func disabledOutboundConfig() Options {
	options := defaultOutboundConfig()
	options.OutboundLogging = rtlog.OutboundOptions{
		Enabled: false,
		Level:   "info",
	}
	return options
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type staticRoundTripper struct{}

func (staticRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       http.NoBody,
		Header:     make(http.Header),
		Request:    req,
	}, nil
}
