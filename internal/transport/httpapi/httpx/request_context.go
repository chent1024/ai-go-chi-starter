package httpx

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"ai-go-chi-starter/internal/service/shared"
)

const (
	RequestIDHeader   = "X-Request-Id"
	traceparentHeader = "Traceparent"
)

func TraceparentHeader() string {
	return traceparentHeader
}

type loggerContextKey struct{}

func RequestID(req *http.Request) string {
	if req == nil {
		return ""
	}
	if value := strings.TrimSpace(req.Header.Get(RequestIDHeader)); value != "" {
		return value
	}
	if value := strings.TrimSpace(req.Header.Get("X-Request-ID")); value != "" {
		return value
	}
	return ""
}

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	if logger == nil {
		return ctx
	}
	return context.WithValue(ctx, loggerContextKey{}, logger)
}

func LoggerFromContext(ctx context.Context) (*slog.Logger, bool) {
	logger, ok := ctx.Value(loggerContextKey{}).(*slog.Logger)
	return logger, ok
}

func RequestLogger(req *http.Request, base *slog.Logger) *slog.Logger {
	if req == nil {
		return base
	}
	logger := base
	if ctxLogger, ok := LoggerFromContext(req.Context()); ok && ctxLogger != nil {
		return ctxLogger
	}
	if logger == nil {
		return nil
	}
	if requestID := RequestID(req); requestID != "" {
		logger = logger.With(shared.LogFieldRequestID, requestID)
	}
	if trace, ok := shared.TraceFromContext(req.Context()); ok {
		logger = logger.With(shared.TraceLogFields(trace)...)
	}
	return logger
}

func ReplaceRequestContext(req *http.Request, ctx context.Context) {
	if req == nil {
		return
	}
	*req = *req.WithContext(ctx)
}
