package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"ai-go-chi-starter/internal/config"
	"ai-go-chi-starter/internal/transport/httpapi/httpx"
	"ai-go-chi-starter/internal/transport/httpapi/middleware"
	v1 "ai-go-chi-starter/internal/transport/httpapi/v1"
)

type ReadyChecker interface {
	Ready(ctx context.Context) error
}

type RouterOptions struct {
	Logging        config.LoggingConfig
	Logger         *slog.Logger
	ExampleHandler *v1.ExampleHandler
	ReadyChecker   ReadyChecker
}

func NewRouter(options RouterOptions) http.Handler {
	logger := options.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Trace)
	r.Use(middleware.AccessLog(logger, options.Logging.AccessEnabled))
	r.Use(middleware.Recover(logger))

	r.Get("/healthz", func(w http.ResponseWriter, req *http.Request) {
		httpx.WriteEnvelope(w, http.StatusOK, httpx.RequestID(req), map[string]string{"status": "ok"})
	})
	r.Get("/readyz", func(w http.ResponseWriter, req *http.Request) {
		if options.ReadyChecker != nil {
			if err := options.ReadyChecker.Ready(req.Context()); err != nil {
				httpx.WriteRequestError(
					w,
					req,
					http.StatusServiceUnavailable,
					"NOT_READY",
					"service is not ready",
					true,
				)
				return
			}
		}
		httpx.WriteEnvelope(w, http.StatusOK, httpx.RequestID(req), map[string]string{"status": "ready"})
	})
	if options.ExampleHandler != nil {
		r.Post("/v1/examples", options.ExampleHandler.Create)
		r.Get("/v1/examples", options.ExampleHandler.List)
		r.Get("/v1/examples/{id}", options.ExampleHandler.Get)
	}
	return r
}
