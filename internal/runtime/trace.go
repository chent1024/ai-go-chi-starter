package runtime

import (
	"context"
	"log/slog"

	"ai-go-chi-starter/internal/service/shared"
)

func StartTrace(ctx context.Context) (context.Context, shared.Trace) {
	if parent, ok := shared.TraceFromContext(ctx); ok {
		trace := shared.NewChildTrace(parent)
		return shared.WithTrace(ctx, trace), trace
	}
	trace := shared.NewRootTrace()
	return shared.WithTrace(ctx, trace), trace
}

func ContinueTrace(ctx context.Context, traceparent string) (context.Context, shared.Trace) {
	if parent, ok := shared.ParseTraceparent(traceparent); ok {
		trace := shared.NewChildTrace(parent)
		return shared.WithTrace(ctx, trace), trace
	}
	return StartTrace(ctx)
}

func WithTrace(logger *slog.Logger, trace shared.Trace) *slog.Logger {
	if logger == nil {
		return nil
	}
	fields := shared.TraceLogFields(trace)
	if len(fields) == 0 {
		return logger
	}
	return logger.With(fields...)
}
