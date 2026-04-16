package runtime

import (
	"context"
	"log/slog"
	"net/url"
	"time"

	"ai-go-chi-starter/internal/config"
)

type OutboundLogEvent struct {
	Component string
	Target    string
	Method    string
	URL       string
	Status    int
	Latency   time.Duration
	BytesIn   int64
	BytesOut  int64
	Err       error
}

func LogOutboundSuccess(logger *slog.Logger, cfg config.LoggingConfig, event OutboundLogEvent) {
	if !cfg.OutboundEnabled || logger == nil {
		return
	}
	level := parseLevel(cfg.OutboundLevel)
	if !levelEnabled(cfg, level) {
		return
	}
	logAtLevel(logger, level, "outbound request completed", outboundAttrs(event)...)
}

func LogOutboundFailure(logger *slog.Logger, _ config.LoggingConfig, event OutboundLogEvent) {
	if logger == nil {
		return
	}
	attrs := outboundAttrs(event)
	if event.Err != nil {
		attrs = append(attrs, "err", event.Err)
	}
	level := slog.LevelWarn
	if event.Status == 0 || event.Status >= 500 {
		level = slog.LevelError
	}
	logAtLevel(logger, level, "outbound request failed", attrs...)
}

func logAtLevel(logger *slog.Logger, level slog.Level, message string, attrs ...any) {
	if logger == nil || !logger.Enabled(context.Background(), level) {
		return
	}
	logger.Log(context.Background(), level, message, attrs...)
}

func outboundAttrs(event OutboundLogEvent) []any {
	attrs := []any{
		"kind", "outbound",
		"target", event.Target,
		"method", event.Method,
		"url", sanitizeOutboundURL(event.URL),
		"latency_ms", event.Latency.Milliseconds(),
	}
	if event.Component != "" {
		attrs = append(attrs, LogFieldComponent, event.Component)
	}
	if event.Status > 0 {
		attrs = append(attrs, LogFieldStatus, event.Status)
	}
	if event.BytesIn > 0 {
		attrs = append(attrs, "bytes_in", event.BytesIn)
	}
	if event.BytesOut > 0 {
		attrs = append(attrs, "bytes_out", event.BytesOut)
	}
	return attrs
}

func sanitizeOutboundURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	parsed.RawQuery = ""
	parsed.User = nil
	return parsed.String()
}
