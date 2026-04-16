package shared

import "net/http"

func ErrInternal(message string, options ...ErrorOption) error {
	return NewError(CodeInternal, message, http.StatusInternalServerError, options...)
}

func ErrInvalidArgument(message string, options ...ErrorOption) error {
	return NewError(CodeInvalidArgument, message, http.StatusBadRequest, options...)
}

func ErrNotReady(message string, options ...ErrorOption) error {
	return NewError(CodeNotReady, message, http.StatusServiceUnavailable, options...)
}

func ErrRequestTimeout(message string, options ...ErrorOption) error {
	return NewError(CodeRequestTimeout, message, http.StatusGatewayTimeout, options...)
}

func ErrNotFound(code, message string, options ...ErrorOption) error {
	return NewError(code, message, http.StatusNotFound, options...)
}
