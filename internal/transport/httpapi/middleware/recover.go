package middleware

import (
	"log/slog"
	"net/http"

	"ai-go-chi-starter/internal/transport/httpapi/httpx"
)

func Recover(base *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger := httpx.RequestLogger(req, base)
					if logger != nil {
						logger.Error("panic recovered", "err", recovered)
					}
					httpx.WriteRequestError(
						w,
						req,
						http.StatusInternalServerError,
						"INTERNAL",
						"internal server error",
						false,
					)
				}
			}()
			next.ServeHTTP(w, req)
		})
	}
}
