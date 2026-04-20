package middleware

import (
	"net/http"
	"time"

	"ai-go-chi-starter/internal/transport/httpapi/httpx"
	apimetrics "ai-go-chi-starter/internal/transport/httpapi/metrics"
)

func Metrics(metrics *apimetrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if metrics == nil {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			startedAt := time.Now()
			metrics.IncInFlight()
			defer metrics.DecInFlight()
			recorder := httpx.NewResponseRecorder(w)
			next.ServeHTTP(recorder, req)
			metrics.ObserveHTTPRequest(routePattern(req), req.Method, recorder.StatusCode(), time.Since(startedAt))
		})
	}
}
