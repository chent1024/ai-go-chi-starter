package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"ai-go-chi-starter/internal/config"
	"ai-go-chi-starter/internal/runtime"
)

type JobHandler interface {
	Handle(ctx context.Context) error
}

type application struct {
	logger   *slog.Logger
	shutdown func(context.Context) error
	runLoop  func(context.Context) error
	close    func()
}

func run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
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
		return err
	case <-ctx.Done():
		app.logger.Info("worker shutting down")
		return app.shutdown(context.Background())
	}
}

func newApplication(ctx context.Context, cfg config.Config, handler JobHandler) (*application, error) {
	logger, logCloser := runtime.NewLogger(cfg.Logging, "worker", os.Stdout)
	runtime.StartLogCleanup(ctx, logger.With("component", "logging"), cfg.Logging)
	done := make(chan struct{})

	worker := tickerWorker{
		interval: cfg.Worker.PollInterval,
		handler:  handler,
		logger:   logger.With("component", "worker"),
	}

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
			return worker.Run(ctx)
		},
		shutdown: func(parent context.Context) error {
			waitCtx, cancel := context.WithTimeout(parent, cfg.Worker.ShutdownTimeout)
			defer cancel()
			select {
			case <-done:
				return nil
			case <-waitCtx.Done():
				if waitCtx.Err() == context.DeadlineExceeded {
					return fmt.Errorf("worker shutdown timed out after %s", cfg.Worker.ShutdownTimeout)
				}
				return waitCtx.Err()
			}
		},
		close: func() {
			_ = logCloser.Close()
		},
	}, nil
}

type tickerWorker struct {
	interval time.Duration
	handler  JobHandler
	logger   *slog.Logger
}

func (w tickerWorker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := w.handler.Handle(ctx); err != nil {
				return err
			}
			w.logger.Debug("worker tick completed")
		}
	}
}

type noopJobHandler struct{}

func (noopJobHandler) Handle(ctx context.Context) error {
	_ = ctx
	return nil
}
