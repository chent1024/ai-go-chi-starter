package httpx

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"ai-go-chi-starter/internal/service/shared"
)

func WriteRequestDomainError(w http.ResponseWriter, req *http.Request, err error) {
	if req == nil {
		WriteDomainError(w, "", err)
		return
	}
	code := shared.Code(err)
	status := shared.HTTPStatus(err)
	if code == "" || status == 0 {
		WriteRequestError(
			w,
			req,
			http.StatusInternalServerError,
			"INTERNAL",
			"internal server error",
			shared.Retryable(err),
		)
		return
	}
	WriteRequestError(w, req, status, code, shared.Message(err), shared.Retryable(err))
}

func WriteRequestError(
	w http.ResponseWriter,
	req *http.Request,
	status int,
	code, message string,
	retryable bool,
) {
	requestID := ""
	if req != nil {
		requestID = RequestID(req)
		logRequestFailure(req, status, code, retryable, message)
	}
	WriteError(w, status, requestID, code, message, retryable)
}

func WriteDomainError(w http.ResponseWriter, requestID string, err error) {
	code := shared.Code(err)
	status := shared.HTTPStatus(err)
	if code == "" || status == 0 {
		WriteError(
			w,
			http.StatusInternalServerError,
			requestID,
			"INTERNAL",
			"internal server error",
			shared.Retryable(err),
		)
		return
	}
	WriteError(w, status, requestID, code, shared.Message(err), shared.Retryable(err))
}

func logRequestFailure(req *http.Request, status int, code string, retryable bool, err any) {
	logger := RequestLogger(req, nil)
	if logger == nil {
		return
	}
	level := slog.LevelWarn
	if status >= 500 {
		level = slog.LevelError
	}
	logger.Log(
		req.Context(),
		level,
		"http request failed",
		"kind", "error",
		"method", req.Method,
		"route", routePattern(req),
		"path", req.URL.Path,
		"status", status,
		shared.LogFieldErrorCode, code,
		shared.LogFieldRetryable, retryable,
		"err", err,
	)
}

func routePattern(req *http.Request) string {
	if req == nil {
		return ""
	}
	if routeContext := chi.RouteContext(req.Context()); routeContext != nil {
		return routeContext.RoutePattern()
	}
	return ""
}
