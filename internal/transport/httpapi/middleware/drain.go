package middleware

import (
	"net/http"

	"ai-go-chi-starter/internal/runtime"
	"ai-go-chi-starter/internal/service/shared"
	"ai-go-chi-starter/internal/transport/httpapi/httpx"
)

func Drain(state *runtime.DrainState) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if state == nil {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if bypassDrain(req) {
				next.ServeHTTP(w, req)
				return
			}
			if !state.StartRequest() {
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
			defer state.FinishRequest()
			next.ServeHTTP(w, req)
		})
	}
}

func bypassDrain(req *http.Request) bool {
	if req == nil {
		return false
	}
	return req.URL.Path == "/healthz" || req.URL.Path == "/readyz"
}
