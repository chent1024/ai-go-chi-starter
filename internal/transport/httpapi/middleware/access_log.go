package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"ai-go-chi-starter/internal/service/shared"
	"ai-go-chi-starter/internal/transport/httpapi/httpx"
)

func AccessLog(base *slog.Logger, enabled bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if !enabled {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			startedAt := time.Now()
			logger := httpx.RequestLogger(req, base)
			if logger != nil {
				httpx.ReplaceRequestContext(req, httpx.WithLogger(req.Context(), logger))
			}
			recorder := httpx.NewResponseRecorder(w)
			next.ServeHTTP(recorder, req)
			writeAccessLog(req, recorder, startedAt)
		})
	}
}

func writeAccessLog(req *http.Request, recorder *httpx.ResponseRecorder, startedAt time.Time) {
	logger := httpx.RequestLogger(req, nil)
	if logger == nil {
		return
	}
	logger.Info(
		"http request completed",
		"kind", "access",
		"method", req.Method,
		"route", routePattern(req),
		"path", req.URL.Path,
		"status", recorder.StatusCode(),
		"latency_ms", time.Since(startedAt).Milliseconds(),
		"bytes_in", requestBytesIn(req),
		"bytes_out", recorder.BytesWritten(),
	)
}

func logRequestFailure(req *http.Request, status int, code string, retryable bool, err any) {
	logger := httpx.RequestLogger(req, nil)
	if logger == nil {
		return
	}
	level := slog.LevelWarn
	if status >= 500 {
		level = slog.LevelError
	}
	logger.Log(
		req.Context(),
		level,
		"http request failed",
		"kind", "error",
		"method", req.Method,
		"route", routePattern(req),
		"path", req.URL.Path,
		"status", status,
		shared.LogFieldErrorCode, code,
		shared.LogFieldRetryable, retryable,
		"err", err,
	)
}

func routePattern(req *http.Request) string {
	if req == nil {
		return ""
	}
	if routeContext := chi.RouteContext(req.Context()); routeContext != nil {
		return routeContext.RoutePattern()
	}
	return ""
}

func requestBytesIn(req *http.Request) int64 {
	if req == nil || req.ContentLength < 0 {
		return 0
	}
	return req.ContentLength
}
