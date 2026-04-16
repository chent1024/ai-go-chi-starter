package middleware

import "net/http"

func BodyLimit(maxBodyBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if maxBodyBytes <= 0 {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req != nil && req.Body != nil {
				req.Body = http.MaxBytesReader(w, req.Body, maxBodyBytes)
			}
			next.ServeHTTP(w, req)
		})
	}
}
