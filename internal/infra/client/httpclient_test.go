package client

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"ai-go-chi-starter/internal/config"
	"ai-go-chi-starter/internal/service/shared"
)

func TestConfigureTransportAppliesOutboundSettings(t *testing.T) {
	cfg := config.OutboundConfig{
		Timeout:               45 * time.Second,
		MaxIdleConns:          120,
		MaxIdleConnsPerHost:   24,
		IdleConnTimeout:       95 * time.Second,
		TLSHandshakeTimeout:   7 * time.Second,
		ResponseHeaderTimeout: 18 * time.Second,
		ExpectContinueTimeout: 2 * time.Second,
	}

	base := &http.Transport{}
	client := NewHTTPClient(&http.Client{Transport: base}, discardLogger(), config.LoggingConfig{}, cfg, "svc", "dep")

	if client.Timeout != cfg.Timeout {
		t.Fatalf("client timeout = %v, want %v", client.Timeout, cfg.Timeout)
	}

	transport, ok := configureTransport(base, cfg).(*http.Transport)
	if !ok {
		t.Fatal("configureTransport() did not return *http.Transport")
	}
	if transport.MaxIdleConns != cfg.MaxIdleConns {
		t.Fatalf("MaxIdleConns = %d, want %d", transport.MaxIdleConns, cfg.MaxIdleConns)
	}
	if transport.MaxIdleConnsPerHost != cfg.MaxIdleConnsPerHost {
		t.Fatalf("MaxIdleConnsPerHost = %d, want %d", transport.MaxIdleConnsPerHost, cfg.MaxIdleConnsPerHost)
	}
	if transport.IdleConnTimeout != cfg.IdleConnTimeout {
		t.Fatalf("IdleConnTimeout = %v, want %v", transport.IdleConnTimeout, cfg.IdleConnTimeout)
	}
	if transport.TLSHandshakeTimeout != cfg.TLSHandshakeTimeout {
		t.Fatalf("TLSHandshakeTimeout = %v, want %v", transport.TLSHandshakeTimeout, cfg.TLSHandshakeTimeout)
	}
	if transport.ResponseHeaderTimeout != cfg.ResponseHeaderTimeout {
		t.Fatalf("ResponseHeaderTimeout = %v, want %v", transport.ResponseHeaderTimeout, cfg.ResponseHeaderTimeout)
	}
	if transport.ExpectContinueTimeout != cfg.ExpectContinueTimeout {
		t.Fatalf("ExpectContinueTimeout = %v, want %v", transport.ExpectContinueTimeout, cfg.ExpectContinueTimeout)
	}
}

func TestNewHTTPClientInjectsTraceparent(t *testing.T) {
	trace := shared.NewRootTrace()
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
		config.LoggingConfig{},
		defaultOutboundConfig(),
		"svc",
		"dep",
	)

	req, err := http.NewRequestWithContext(
		shared.WithTrace(context.Background(), trace),
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
	if got := captured.Header.Get(traceparentHeader); got != trace.Traceparent() {
		t.Fatalf("Traceparent = %q, want %q", got, trace.Traceparent())
	}
}

func TestConfigureTransportKeepsCustomRoundTripper(t *testing.T) {
	base := staticRoundTripper{}

	configured := configureTransport(base, defaultOutboundConfig())
	if _, ok := configured.(staticRoundTripper); !ok {
		t.Fatal("configureTransport() should leave non-transport round trippers unchanged")
	}
}

func defaultOutboundConfig() config.OutboundConfig {
	return config.OutboundConfig{
		Timeout:               30 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
		ExpectContinueTimeout: time.Second,
	}
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
