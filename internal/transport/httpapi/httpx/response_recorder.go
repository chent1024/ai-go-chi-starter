package httpx

import "net/http"

type ResponseRecorder struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	return &ResponseRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (r *ResponseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *ResponseRecorder) Write(p []byte) (int, error) {
	if r.statusCode == 0 {
		r.statusCode = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(p)
	r.bytesWritten += int64(n)
	return n, err
}

func (r *ResponseRecorder) StatusCode() int {
	if r.statusCode == 0 {
		return http.StatusOK
	}
	return r.statusCode
}

func (r *ResponseRecorder) BytesWritten() int64 {
	return r.bytesWritten
}
