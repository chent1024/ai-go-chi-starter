package middleware

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	rttrace "ai-go-chi-starter/internal/runtime/tracing"
	"ai-go-chi-starter/internal/transport/httpapi/httpx"
)

var requestCounter uint64

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestID := httpx.RequestID(req)
		if requestID == "" {
			requestID = nextRequestID()
			req.Header.Set(httpx.RequestIDHeader, requestID)
		}
		httpx.ReplaceRequestContext(req, rttrace.ContextWithRequestID(req.Context(), requestID))
		w.Header().Set(httpx.RequestIDHeader, requestID)
		next.ServeHTTP(w, req)
	})
}

func nextRequestID() string {
	value := atomic.AddUint64(&requestCounter, 1)
	return fmt.Sprintf("req_%d_%06d", time.Now().UTC().Unix(), value)
}
