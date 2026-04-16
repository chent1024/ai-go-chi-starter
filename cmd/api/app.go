package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"ai-go-chi-starter/internal/config"
	"ai-go-chi-starter/internal/infra/store/postgres"
	"ai-go-chi-starter/internal/runtime"
	"ai-go-chi-starter/internal/service/example"
	"ai-go-chi-starter/internal/transport/httpapi"
	v1 "ai-go-chi-starter/internal/transport/httpapi/v1"
)

func run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	app, err := newApplication(ctx, cfg)
	if err != nil {
		return err
	}
	defer app.close()

	app.logger.Info("api server starting", "addr", cfg.API.ListenAddr)

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		app.logger.Info("api shutdown requested", "reason", ctx.Err())
		return app.shutdown(context.Background())
	}
}

type application struct {
	server     *http.Server
	logger     *slog.Logger
	drainState *runtime.DrainState
	shutdown   func(context.Context) error
	close      func()
}

func newApplication(ctx context.Context, cfg config.Config) (*application, error) {
	logger, logCloser := runtime.NewLogger(cfg.Logging, "api", os.Stdout)
	runtime.StartLogCleanup(ctx, logger.With("component", "logging"), cfg.Logging)
	drainState := &runtime.DrainState{}

	db, err := postgres.Open(ctx, cfg.Database)
	if err != nil {
		_ = logCloser.Close()
		return nil, err
	}

	repo := postgres.NewExampleRepository(db)
	service := example.NewService(repo)
	handler := v1.NewExampleHandler(service).WithLogger(logger.With("component", "example_handler"))
	router := httpapi.NewRouter(httpapi.RouterOptions{
		Logging:        cfg.Logging,
		RequestTimeout: cfg.API.RequestTimeout,
		DrainState:     drainState,
		Logger:         logger,
		ExampleHandler: handler,
		ReadyChecker:   postgres.ReadyChecker{DB: db},
	})

	server := &http.Server{
		Addr:              cfg.API.ListenAddr,
		Handler:           router,
		ReadTimeout:       cfg.API.ReadTimeout,
		ReadHeaderTimeout: cfg.API.ReadTimeout,
		WriteTimeout:      cfg.API.WriteTimeout,
		IdleTimeout:       cfg.API.IdleTimeout,
		MaxHeaderBytes:    cfg.API.MaxHeaderBytes,
	}

	return &application{
		server:     server,
		logger:     logger,
		drainState: drainState,
		shutdown:   newShutdownFunc(server, logger, drainState, cfg.API.ShutdownTimeout),
		close: func() {
			_ = db.Close()
			_ = logCloser.Close()
		},
	}, nil
}

func newShutdownFunc(
	server *http.Server,
	logger *slog.Logger,
	drainState *runtime.DrainState,
	timeout time.Duration,
) func(context.Context) error {
	return func(parent context.Context) error {
		server.SetKeepAlivesEnabled(false)
		if drainState != nil {
			drainState.BeginDrain()
		}
		startedAt := time.Now()
		if logger != nil {
			logger.Info(
				"api graceful shutdown started",
				"shutdown_timeout", timeout.String(),
				"active_requests", drainState.ActiveRequests(),
			)
		}

		ctx, cancel := context.WithTimeout(parent, timeout)
		defer cancel()

		err := server.Shutdown(ctx)
		if logger != nil {
			if err != nil {
				logger.Error(
					"api graceful shutdown failed",
					"err", err,
					"active_requests", drainState.ActiveRequests(),
					"elapsed_ms", time.Since(startedAt).Milliseconds(),
				)
			} else {
				logger.Info(
					"api graceful shutdown completed",
					"active_requests", drainState.ActiveRequests(),
					"elapsed_ms", time.Since(startedAt).Milliseconds(),
				)
			}
		}
		return err
	}
}
