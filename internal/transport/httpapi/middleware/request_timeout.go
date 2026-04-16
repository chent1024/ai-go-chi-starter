package middleware

import (
	"context"
	"net/http"
	"time"

	"ai-go-chi-starter/internal/service/shared"
	"ai-go-chi-starter/internal/transport/httpapi/httpx"
)

func RequestTimeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if timeout <= 0 {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx, cancel := context.WithTimeout(req.Context(), timeout)
			defer cancel()

			recorder := httpx.NewResponseRecorder(w)
			next.ServeHTTP(recorder, req.WithContext(ctx))

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
