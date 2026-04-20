package middleware

import (
	"net/http"

	rttrace "ai-go-chi-starter/internal/runtime/tracing"
	"ai-go-chi-starter/internal/transport/httpapi/httpx"
)

func Trace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		trace := inboundTrace(req)
		w.Header().Set(httpx.TraceparentHeader(), trace.Traceparent())
		ctx := rttrace.ContextWithTrace(req.Context(), trace)
		httpx.ReplaceRequestContext(req, ctx)
		next.ServeHTTP(w, req)
	})
}

func inboundTrace(req *http.Request) rttrace.Trace {
	if req != nil {
		if parent, ok := rttrace.ParseTraceparent(req.Header.Get(httpx.TraceparentHeader())); ok {
			return rttrace.NewChildTrace(parent)
		}
	}
	return rttrace.NewRootTrace()
}
