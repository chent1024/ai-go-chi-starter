package runtime

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"ai-go-chi-starter/internal/config"
)

type dailyLogWriter struct {
	service     string
	dir         string
	location    *time.Location
	now         func() time.Time
	mu          sync.Mutex
	currentDate string
	file        *os.File
}

func newDailyLogWriter(service, dir string, location *time.Location) io.WriteCloser {
	if strings.TrimSpace(dir) == "" {
		return nopWriteCloser{Writer: io.Discard}
	}
	return &dailyLogWriter{
		service:  service,
		dir:      dir,
		location: logLocation(location),
		now:      time.Now,
	}
}

func (w *dailyLogWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	file, err := w.ensureFileLocked(w.now())
	if err != nil {
		return 0, err
	}
	return file.Write(p)
}

func (w *dailyLogWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	w.currentDate = ""
	return err
}

func (w *dailyLogWriter) ensureFileLocked(now time.Time) (*os.File, error) {
	datedNow := now.In(w.location)
	dateKey := datedNow.Format(time.DateOnly)
	if w.file != nil && w.currentDate == dateKey {
		return w.file, nil
	}
	if err := os.MkdirAll(w.dir, 0o755); err != nil {
		return nil, err
	}
	if w.file != nil {
		_ = w.file.Close()
	}
	filePath := filepath.Join(w.dir, buildLogFilename(w.service, datedNow))
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	w.file = file
	w.currentDate = dateKey
	return w.file, nil
}

func buildLogFilename(service string, now time.Time) string {
	return fmt.Sprintf("%s-%s.log", service, now.Format(time.DateOnly))
}

func CleanupLogFiles(dir string, retentionDays int, location *time.Location) error {
	return cleanupLogFiles(dir, retentionDays, location, time.Now)
}

func cleanupLogFiles(
	dir string,
	retentionDays int,
	location *time.Location,
	now func() time.Time,
) error {
	if retentionDays <= 0 {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	current := now().In(logLocation(location))
	today := time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, logLocation(location))
	cutoff := today.AddDate(0, 0, -retentionDays+1)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		_, fileDate, ok := parseLogFilename(entry.Name(), logLocation(location))
		if !ok {
			continue
		}
		if fileDate.Before(cutoff) {
			if err := os.Remove(filepath.Join(dir, entry.Name())); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}
	return nil
}

func StartLogCleanup(ctx context.Context, logger *slog.Logger, cfg config.LoggingConfig) {
	if logger == nil || cfg.Output == "stdout" {
		return
	}
	runLogCleanup(logger, cfg)
	ticker := time.NewTicker(cfg.CleanupInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				runLogCleanup(logger, cfg)
			}
		}
	}()
}

func runLogCleanup(logger *slog.Logger, cfg config.LoggingConfig) {
	if err := CleanupLogFiles(cfg.Dir, cfg.RetentionDays, cfg.Location); err != nil {
		logger.Error(
			"log cleanup failed",
			LogFieldComponent, "logging",
			LogFieldErrorCode, "LOG_CLEANUP_FAILED",
			"err", err,
			"log_dir", cfg.Dir,
		)
	}
}

func parseLogFilename(filename string, location *time.Location) (string, time.Time, bool) {
	if !strings.HasSuffix(filename, ".log") {
		return "", time.Time{}, false
	}
	base := strings.TrimSuffix(filename, ".log")
	if len(base) <= len("-2006-01-02") {
		return "", time.Time{}, false
	}
	rawDate := base[len(base)-len(time.DateOnly):]
	service := strings.TrimSuffix(base[:len(base)-len(time.DateOnly)], "-")
	if service == "" {
		return "", time.Time{}, false
	}
	fileDate, err := time.ParseInLocation(time.DateOnly, rawDate, logLocation(location))
	if err != nil {
		return "", time.Time{}, false
	}
	return service, fileDate, true
}

type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error {
	return nil
}
