package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"ai-go-chi-starter/internal/service/shared"
	"ai-go-chi-starter/internal/transport/httpapi/httpx"
	apimetrics "ai-go-chi-starter/internal/transport/httpapi/metrics"
)

func RequestTimeout(timeout time.Duration, logger *slog.Logger, metrics *apimetrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if timeout <= 0 {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx, cancel := context.WithTimeout(req.Context(), timeout)
			defer cancel()

			req = req.WithContext(ctx)
			recorder := httpx.NewDeadlineAwareResponseRecorder(w, ctx)
			next.ServeHTTP(recorder, req)

			if lateWrites := recorder.LateWriteCount(); ctx.Err() == context.DeadlineExceeded && lateWrites > 0 {
				if metrics != nil {
					metrics.ObserveTimeoutLateWrite(routePattern(req), lateWrites)
				}
				requestLogger := httpx.RequestLogger(req, logger)
				if requestLogger != nil {
					requestLogger.Warn(
						"request timed out after handler continued writing",
						"kind", "timeout",
						"method", req.Method,
						"route", routePattern(req),
						"path", req.URL.Path,
						"late_write_count", lateWrites,
					)
				}
			}

			if ctx.Err() == context.DeadlineExceeded && !recorder.Written() {
				httpx.WriteRequestError(
					w,
					req,
					http.StatusGatewayTimeout,
					shared.CodeRequestTimeout,
					"request timed out",
					true,
				)
			}
		})
	}
}
