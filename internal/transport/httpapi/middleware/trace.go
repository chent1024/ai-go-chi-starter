package middleware

import (
	"net/http"

	"ai-go-chi-starter/internal/service/shared"
	"ai-go-chi-starter/internal/transport/httpapi/httpx"
)

func Trace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		trace := inboundTrace(req)
		w.Header().Set(httpx.TraceparentHeader(), trace.Traceparent())
		ctx := shared.WithTrace(req.Context(), trace)
		httpx.ReplaceRequestContext(req, ctx)
		next.ServeHTTP(w, req)
	})
}

func inboundTrace(req *http.Request) shared.Trace {
	if req != nil {
		if parent, ok := shared.ParseTraceparent(req.Header.Get(httpx.TraceparentHeader())); ok {
			return shared.NewChildTrace(parent)
		}
	}
	return shared.NewRootTrace()
}
