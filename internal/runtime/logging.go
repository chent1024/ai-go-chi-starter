package runtime

import (
	"io"
	"log/slog"
	"time"

	"ai-go-chi-starter/internal/config"
)

func NewLogger(cfg config.LoggingConfig, service string, stdout io.Writer) (*slog.Logger, io.Closer) {
	level := new(slog.LevelVar)
	level.Set(parseLevel(cfg.Level))

	options := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.SourceEnabled,
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.TimeKey && attr.Value.Kind() == slog.KindTime {
				return slog.Time(slog.TimeKey, attr.Value.Time().In(logLocation(cfg.Location)))
			}
			return attr
		},
	}

	writer, closer := buildLogWriter(cfg, service, stdout)
	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(writer, options)
	} else {
		handler = slog.NewTextHandler(writer, options)
	}
	return slog.New(handler).With(LogFieldService, service), closer
}

func parseLevel(raw string) slog.Level {
	switch raw {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func levelEnabled(cfg config.LoggingConfig, level slog.Level) bool {
	return level >= parseLevel(cfg.Level)
}

func logLocation(location *time.Location) *time.Location {
	if location != nil {
		return location
	}
	return time.UTC
}

func buildLogWriter(cfg config.LoggingConfig, service string, stdout io.Writer) (io.Writer, io.Closer) {
	switch cfg.Output {
	case "file":
		fileWriter := newDailyLogWriter(service, cfg.Dir, cfg.Location)
		return fileWriter, fileWriter
	case "both":
		fileWriter := newDailyLogWriter(service, cfg.Dir, cfg.Location)
		return io.MultiWriter(stdout, fileWriter), fileWriter
	default:
		return stdout, nopWriteCloser{Writer: io.Discard}
	}
}
