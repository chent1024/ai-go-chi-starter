package runtime

import (
	"context"
	"log/slog"
	"time"
)

const (
	LogFieldSpanName   = "span_name"
	LogFieldSpanStatus = "span_status"
)

type Span struct {
	ctx       context.Context
	logger    *slog.Logger
	name      string
	startedAt time.Time
}

func StartSpan(ctx context.Context, logger *slog.Logger, name string, attrs ...any) (context.Context, *Span) {
	if ctx == nil {
		ctx = context.Background()
	}
	spanCtx, trace := StartTrace(ctx)
	spanLogger := WithContext(logger, spanCtx)
	span := &Span{
		ctx:       spanCtx,
		logger:    spanLogger,
		name:      name,
		startedAt: time.Now(),
	}
	if spanLogger != nil && spanLogger.Enabled(spanCtx, slog.LevelDebug) {
		spanLogger.Log(spanCtx, slog.LevelDebug, "span started", span.startAttrs(trace, attrs...)...)
	}
	return spanCtx, span
}

func (s *Span) Logger() *slog.Logger {
	if s == nil {
		return nil
	}
	return s.logger
}

func (s *Span) End(err error, attrs ...any) {
	if s == nil || s.logger == nil || !s.logger.Enabled(s.ctx, slog.LevelDebug) {
		return
	}
	status := "ok"
	if err != nil {
		status = "error"
	}
	logAttrs := []any{
		"kind", "span",
		LogFieldSpanName, s.name,
		LogFieldSpanStatus, status,
		"latency_ms", time.Since(s.startedAt).Milliseconds(),
	}
	if err != nil {
		logAttrs = append(logAttrs, "err", err)
	}
	logAttrs = append(logAttrs, attrs...)
	s.logger.Log(s.ctx, slog.LevelDebug, "span completed", logAttrs...)
}

func (s *Span) startAttrs(_ any, attrs ...any) []any {
	logAttrs := []any{
		"kind", "span",
		LogFieldSpanName, s.name,
	}
	return append(logAttrs, attrs...)
}
