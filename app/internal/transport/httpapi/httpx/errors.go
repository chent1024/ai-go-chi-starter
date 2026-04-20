package httpx

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	rtlog "ai-go-chi-starter/internal/runtime/logging"
	"ai-go-chi-starter/internal/service/shared"
)

func WriteRequestDomainError(w http.ResponseWriter, req *http.Request, err error) {
	if req == nil {
		WriteDomainError(w, "", err)
		return
	}
	code := shared.Code(err)
	status := statusFromError(err)
	details := shared.Details(err)
	if code == "" || status == 0 {
		logRequestFailure(req, http.StatusInternalServerError, shared.CodeInternal, shared.Retryable(err), err, nil)
		WriteError(w, http.StatusInternalServerError, RequestID(req), shared.CodeInternal, "internal server error", shared.Retryable(err))
		return
	}
	WriteRequestError(w, req, status, code, shared.Message(err), shared.Retryable(err), details)
}

func WriteRequestError(
	w http.ResponseWriter,
	req *http.Request,
	status int,
	code, message string,
	retryable bool,
	details ...any,
) {
	requestID := ""
	if req != nil {
		requestID = RequestID(req)
		logRequestFailure(req, status, code, retryable, message, firstDetail(details))
	}
	WriteError(w, status, requestID, code, message, retryable, details...)
}

func WriteDomainError(w http.ResponseWriter, requestID string, err error) {
	code := shared.Code(err)
	status := statusFromError(err)
	details := shared.Details(err)
	if code == "" || status == 0 {
		WriteError(
			w,
			http.StatusInternalServerError,
			requestID,
			shared.CodeInternal,
			"internal server error",
			shared.Retryable(err),
		)
		return
	}
	WriteError(w, status, requestID, code, shared.Message(err), shared.Retryable(err), details)
}

func statusFromError(err error) int {
	switch shared.KindOf(err) {
	case shared.KindInvalidArgument:
		return http.StatusBadRequest
	case shared.KindNotFound:
		return http.StatusNotFound
	case shared.KindNotReady:
		return http.StatusServiceUnavailable
	case shared.KindRequestTimeout:
		return http.StatusGatewayTimeout
	case shared.KindInternal:
		return http.StatusInternalServerError
	default:
		return 0
	}
}

func logRequestFailure(req *http.Request, status int, code string, retryable bool, err any, details any) {
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
		rtlog.LogFieldErrorCode, code,
		rtlog.LogFieldRetryable, retryable,
		"err", err,
		"details", details,
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
