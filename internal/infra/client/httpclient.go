package client

import (
	"log/slog"
	"net/http"
	"time"

	"ai-go-chi-starter/internal/config"
	"ai-go-chi-starter/internal/runtime"
	"ai-go-chi-starter/internal/service/shared"
)

const traceparentHeader = "Traceparent"

func NewHTTPClient(
	base *http.Client,
	logger *slog.Logger,
	logCfg config.LoggingConfig,
	outboundCfg config.OutboundConfig,
	component string,
	target string,
) *http.Client {
	if base == nil {
		base = &http.Client{}
	}
	clone := *base
	clone.Timeout = outboundCfg.Timeout
	clone.Transport = newLoggingRoundTripper(
		configureTransport(clone.Transport, outboundCfg),
		logger,
		logCfg,
		component,
		target,
	)
	return &clone
}

func configureTransport(base http.RoundTripper, cfg config.OutboundConfig) http.RoundTripper {
	transport, ok := cloneTransport(base)
	if !ok {
		return base
	}
	transport.MaxIdleConns = cfg.MaxIdleConns
	transport.MaxIdleConnsPerHost = cfg.MaxIdleConnsPerHost
	transport.IdleConnTimeout = cfg.IdleConnTimeout
	transport.TLSHandshakeTimeout = cfg.TLSHandshakeTimeout
	transport.ResponseHeaderTimeout = cfg.ResponseHeaderTimeout
	transport.ExpectContinueTimeout = cfg.ExpectContinueTimeout
	return transport
}

func cloneTransport(base http.RoundTripper) (*http.Transport, bool) {
	if base == nil {
		defaultTransport, ok := http.DefaultTransport.(*http.Transport)
		if !ok {
			return nil, false
		}
		return defaultTransport.Clone(), true
	}
	transport, ok := base.(*http.Transport)
	if !ok {
		return nil, false
	}
	return transport.Clone(), true
}

func newLoggingRoundTripper(
	base http.RoundTripper,
	logger *slog.Logger,
	cfg config.LoggingConfig,
	component string,
	target string,
) http.RoundTripper {
	return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		req = withTraceparent(req)
		startedAt := time.Now()
		resp, err := base.RoundTrip(req)
		event := runtime.OutboundLogEvent{
			Component: component,
			Target:    target,
			Method:    req.Method,
			URL:       req.URL.String(),
			Latency:   time.Since(startedAt),
		}
		if resp != nil {
			event.Status = resp.StatusCode
			event.BytesIn = resp.ContentLength
		}
		if err != nil {
			event.Err = err
		}
		if err != nil {
			runtime.LogOutboundFailure(logger, cfg, event)
		} else {
			runtime.LogOutboundSuccess(logger, cfg, event)
		}
		return resp, err
	})
}

func withTraceparent(req *http.Request) *http.Request {
	if req == nil || req.Header.Get(traceparentHeader) != "" {
		return req
	}
	trace, ok := shared.TraceFromContext(req.Context())
	if !ok || !trace.Valid() {
		return req
	}
	cloned := req.Clone(req.Context())
	cloned.Header = req.Header.Clone()
	cloned.Header.Set(traceparentHeader, trace.Traceparent())
	return cloned
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
