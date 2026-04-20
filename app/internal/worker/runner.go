package worker

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	rttrace "ai-go-chi-starter/internal/runtime/tracing"
)

type JobHandler interface {
	Handle(context.Context) error
}

type Ticker struct {
	interval   time.Duration
	handler    JobHandler
	logger     *slog.Logger
	activeJobs *atomic.Int64
}

type TickerOptions struct {
	Interval   time.Duration
	Handler    JobHandler
	Logger     *slog.Logger
	ActiveJobs *atomic.Int64
}

func NewTicker(options TickerOptions) Ticker {
	return Ticker{
		interval:   options.Interval,
		handler:    options.Handler,
		logger:     options.Logger,
		activeJobs: options.ActiveJobs,
	}
}

func (w Ticker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if w.activeJobs != nil {
				w.activeJobs.Add(1)
			}
			jobCtx, span := rttrace.StartSpan(ctx, w.logger, "worker.job.handle")
			err := w.handler.Handle(jobCtx)
			span.End(err)
			if w.activeJobs != nil {
				w.activeJobs.Add(-1)
			}
			if err != nil {
				return err
			}
			if w.logger != nil {
				w.logger.Debug("worker tick completed")
			}
		}
	}
}

func Shutdown(parent context.Context, done <-chan struct{}, logger *slog.Logger, activeJobs *atomic.Int64, timeout time.Duration) error {
	if logger != nil {
		logger.Info(
			"worker drain started",
			"shutdown_timeout", timeout.String(),
			"inflight_jobs", loadActiveJobs(activeJobs),
		)
	}

	waitCtx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	select {
	case <-done:
		if logger != nil {
			logger.Info("worker drain completed", "inflight_jobs", loadActiveJobs(activeJobs))
		}
		return nil
	case <-waitCtx.Done():
		if waitCtx.Err() == context.DeadlineExceeded {
			err := fmt.Errorf("worker shutdown timed out after %s", timeout)
			if logger != nil {
				logger.Error("worker drain failed", "err", err, "inflight_jobs", loadActiveJobs(activeJobs))
			}
			return err
		}
		return waitCtx.Err()
	}
}

func loadActiveJobs(activeJobs *atomic.Int64) int64 {
	if activeJobs == nil {
		return 0
	}
	return activeJobs.Load()
}
