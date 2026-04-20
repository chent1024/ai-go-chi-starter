package client

import (
	"log/slog"
	"net/http"
	"time"

	rtlog "ai-go-chi-starter/internal/runtime/logging"
	rttrace "ai-go-chi-starter/internal/runtime/tracing"
)

const traceparentHeader = "Traceparent"

type Options struct {
	Timeout               time.Duration
	MaxIdleConns          int
	MaxIdleConnsPerHost   int
	IdleConnTimeout       time.Duration
	TLSHandshakeTimeout   time.Duration
	ResponseHeaderTimeout time.Duration
	ExpectContinueTimeout time.Duration
	OutboundLogging       rtlog.OutboundOptions
}

func NewHTTPClient(
	base *http.Client,
	logger *slog.Logger,
	options Options,
	component string,
	target string,
) *http.Client {
	if base == nil {
		base = &http.Client{}
	}
	clone := *base
	clone.Timeout = options.Timeout
	clone.Transport = newLoggingRoundTripper(
		configureTransport(clone.Transport, options),
		logger,
		options.OutboundLogging,
		component,
		target,
	)
	return &clone
}

func configureTransport(base http.RoundTripper, options Options) http.RoundTripper {
	transport, ok := cloneTransport(base)
	if !ok {
		return base
	}
	transport.MaxIdleConns = options.MaxIdleConns
	transport.MaxIdleConnsPerHost = options.MaxIdleConnsPerHost
	transport.IdleConnTimeout = options.IdleConnTimeout
	transport.TLSHandshakeTimeout = options.TLSHandshakeTimeout
	transport.ResponseHeaderTimeout = options.ResponseHeaderTimeout
	transport.ExpectContinueTimeout = options.ExpectContinueTimeout
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
	options rtlog.OutboundOptions,
	component string,
	target string,
) http.RoundTripper {
	return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req == nil {
			return base.RoundTrip(req)
		}
		spanCtx, span := rttrace.StartSpan(
			req.Context(),
			logger,
			"outbound.http.roundtrip",
			"target", target,
			"method", req.Method,
		)
		req = req.Clone(spanCtx)
		req = withTraceparent(req)
		startedAt := time.Now()
		resp, err := base.RoundTrip(req)
		event := rtlog.OutboundEvent{
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
		spanLogger := span.Logger()
		if err != nil {
			rtlog.LogOutboundFailure(spanLogger, event)
		} else {
			rtlog.LogOutboundSuccess(spanLogger, options, event)
		}
		span.End(
			err,
			"target", target,
			"method", req.Method,
			rtlog.LogFieldStatus, event.Status,
		)
		return resp, err
	})
}

func withTraceparent(req *http.Request) *http.Request {
	if req == nil || req.Header.Get(traceparentHeader) != "" {
		return req
	}
	trace, ok := rttrace.TraceFromContext(req.Context())
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
