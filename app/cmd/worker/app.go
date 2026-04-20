package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync/atomic"

	"ai-go-chi-starter/internal/config"
	rtlog "ai-go-chi-starter/internal/runtime/logging"
	"ai-go-chi-starter/internal/worker"
)

type application struct {
	logger     *slog.Logger
	shutdown   func(context.Context) error
	runLoop    func(context.Context) error
	close      func()
	activeJobs *atomic.Int64
}

func run(ctx context.Context) error {
	bootstrapLogger := rtlog.NewBootstrapLogger("worker", os.Stderr)
	cfg, err := config.Load()
	if err != nil {
		bootstrapLogger.Error("worker bootstrap failed", "kind", "fatal", "stage", "config", "err", err)
		return fmt.Errorf("load config: %w", err)
	}

	app, err := newApplication(ctx, cfg, noopJobHandler{})
	if err != nil {
		return err
	}
	defer app.close()

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.runLoop(ctx)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			app.logger.Error("worker exited unexpectedly", "err", err, "inflight_jobs", app.activeJobs.Load())
		}
		return err
	case <-ctx.Done():
		app.logger.Info("worker shutdown requested", "reason", ctx.Err(), "inflight_jobs", app.activeJobs.Load())
		return app.shutdown(context.Background())
	}
}

func newApplication(ctx context.Context, cfg config.Config, handler worker.JobHandler) (*application, error) {
	logOptions := rtlog.Options{
		Level:           cfg.Logging.Level,
		Format:          cfg.Logging.Format,
		SourceEnabled:   cfg.Logging.SourceEnabled,
		Output:          cfg.Logging.Output,
		Dir:             cfg.Logging.Dir,
		RetentionDays:   cfg.Logging.RetentionDays,
		CleanupInterval: cfg.Logging.CleanupInterval,
		Location:        cfg.Logging.Location,
	}
	logger, logCloser := rtlog.NewLogger(logOptions, "worker", os.Stdout)
	rtlog.StartCleanup(ctx, logger.With("component", "logging"), logOptions)
	done := make(chan struct{})
	activeJobs := &atomic.Int64{}

	tickerWorker := worker.NewTicker(worker.TickerOptions{
		Interval:   cfg.Worker.PollInterval,
		Handler:    handler,
		Logger:     logger.With("component", "worker"),
		ActiveJobs: activeJobs,
	})

	return &application{
		logger: logger,
		runLoop: func(ctx context.Context) error {
			defer close(done)
			if !cfg.Worker.Enabled {
				logger.Info("worker disabled by config")
				<-ctx.Done()
				return nil
			}
			logger.Info("worker starting", "poll_interval", cfg.Worker.PollInterval.String())
			return tickerWorker.Run(ctx)
		},
		shutdown: func(parent context.Context) error {
			return worker.Shutdown(parent, done, logger, activeJobs, cfg.Worker.ShutdownTimeout)
		},
		activeJobs: activeJobs,
		close: func() {
			_ = logCloser.Close()
		},
	}, nil
}

type noopJobHandler struct{}

func (noopJobHandler) Handle(ctx context.Context) error {
	_ = ctx
	return nil
}
