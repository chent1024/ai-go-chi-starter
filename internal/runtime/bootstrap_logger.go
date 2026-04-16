package runtime

import (
	"io"
	"log/slog"
)

func NewBootstrapLogger(service string, stderr io.Writer) *slog.Logger {
	handler := slog.NewJSONHandler(stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			return redactAttr(attr)
		},
	})
	return slog.New(handler).With(LogFieldService, service, LogFieldComponent, "bootstrap")
}
