package logging

import (
	"io"
	"log/slog"
	"time"
)

func NewLogger(options Options, service string, stdout io.Writer) (*slog.Logger, io.Closer) {
	level := new(slog.LevelVar)
	level.Set(parseLevel(options.Level))

	handlerOptions := &slog.HandlerOptions{
		Level:     level,
		AddSource: options.SourceEnabled,
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.TimeKey && attr.Value.Kind() == slog.KindTime {
				return slog.Time(slog.TimeKey, attr.Value.Time().In(logLocation(options.Location)))
			}
			return redactAttr(attr)
		},
	}

	writer, closer := buildLogWriter(options, service, stdout)
	var handler slog.Handler
	if options.Format == "json" {
		handler = slog.NewJSONHandler(writer, handlerOptions)
	} else {
		handler = slog.NewTextHandler(writer, handlerOptions)
	}
	return slog.New(handler).With(LogFieldService, service), closer
}

func NewNoopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func WithRequestID(logger *slog.Logger, requestID string) *slog.Logger {
	if logger == nil || requestID == "" {
		return logger
	}
	return logger.With(LogFieldRequestID, requestID)
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

func logLocation(location *time.Location) *time.Location {
	if location != nil {
		return location
	}
	return time.UTC
}

func buildLogWriter(options Options, service string, stdout io.Writer) (io.Writer, io.Closer) {
	switch options.Output {
	case "file":
		fileWriter := newDailyLogWriter(service, options.Dir, options.Location)
		return fileWriter, fileWriter
	case "both":
		fileWriter := newDailyLogWriter(service, options.Dir, options.Location)
		return io.MultiWriter(stdout, fileWriter), fileWriter
	default:
		return stdout, nopWriteCloser{Writer: io.Discard}
	}
}
