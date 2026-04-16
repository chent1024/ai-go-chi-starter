package runtime

import (
	"context"
	"log/slog"

	"ai-go-chi-starter/internal/service/shared"
)

func WithRequestID(logger *slog.Logger, requestID string) *slog.Logger {
	if logger == nil || requestID == "" {
		return logger
	}
	return logger.With(LogFieldRequestID, requestID)
}

func WithContext(logger *slog.Logger, ctx context.Context) *slog.Logger {
	if logger == nil || ctx == nil {
		return logger
	}
	trace, ok := shared.TraceFromContext(ctx)
	if !ok {
		return logger
	}
	return WithTrace(logger, trace)
}
