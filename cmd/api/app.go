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
		app.logger.Info("api server shutting down")
		return app.shutdown(context.Background())
	}
}

type application struct {
	server   *http.Server
	logger   *slog.Logger
	shutdown func(context.Context) error
	close    func()
}

func newApplication(ctx context.Context, cfg config.Config) (*application, error) {
	logger, logCloser := runtime.NewLogger(cfg.Logging, "api", os.Stdout)
	runtime.StartLogCleanup(ctx, logger.With("component", "logging"), cfg.Logging)

	db, err := postgres.Open(ctx, cfg.Database.URL)
	if err != nil {
		_ = logCloser.Close()
		return nil, err
	}

	repo := postgres.NewExampleRepository(db)
	service := example.NewService(repo)
	handler := v1.NewExampleHandler(service).WithLogger(logger.With("component", "example_handler"))
	router := httpapi.NewRouter(httpapi.RouterOptions{
		Logging:        cfg.Logging,
		Logger:         logger,
		ExampleHandler: handler,
		ReadyChecker:   postgres.ReadyChecker{DB: db},
	})

	server := &http.Server{
		Addr:              cfg.API.ListenAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &application{
		server:   server,
		logger:   logger,
		shutdown: newShutdownFunc(server, cfg.API.ShutdownTimeout),
		close: func() {
			_ = db.Close()
			_ = logCloser.Close()
		},
	}, nil
}

func newShutdownFunc(server *http.Server, timeout time.Duration) func(context.Context) error {
	return func(parent context.Context) error {
		ctx, cancel := context.WithTimeout(parent, timeout)
		defer cancel()
		return server.Shutdown(ctx)
	}
}
