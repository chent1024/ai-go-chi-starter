package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"ai-go-chi-starter/internal/config"
	"ai-go-chi-starter/internal/runtime"
	"ai-go-chi-starter/internal/service/shared"
	"ai-go-chi-starter/internal/transport/httpapi/httpx"
	"ai-go-chi-starter/internal/transport/httpapi/middleware"
	v1 "ai-go-chi-starter/internal/transport/httpapi/v1"
)

type ReadyChecker interface {
	Ready(ctx context.Context) error
}

type RouterOptions struct {
	Logging        config.LoggingConfig
	RequestTimeout time.Duration
	DrainState     *runtime.DrainState
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
	r.Use(middleware.Drain(options.DrainState))
	r.Use(middleware.RequestTimeout(options.RequestTimeout))
	r.Use(middleware.Recover(logger))

	r.Get("/healthz", func(w http.ResponseWriter, req *http.Request) {
		httpx.WriteEnvelope(w, http.StatusOK, httpx.RequestID(req), map[string]string{"status": "ok"})
	})
	r.Get("/readyz", func(w http.ResponseWriter, req *http.Request) {
		if options.DrainState != nil && options.DrainState.Draining() {
			httpx.WriteRequestError(
				w,
				req,
				http.StatusServiceUnavailable,
				shared.CodeNotReady,
				"service is shutting down",
				true,
			)
			return
		}
		if options.ReadyChecker != nil {
			if err := options.ReadyChecker.Ready(req.Context()); err != nil {
				httpx.WriteRequestError(
					w,
					req,
					http.StatusServiceUnavailable,
					shared.CodeNotReady,
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
