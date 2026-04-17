package shared

func ErrInternal(message string, options ...ErrorOption) error {
	return NewError(CodeInternal, message, StatusInternalServerError, options...)
}

func ErrInvalidArgument(message string, options ...ErrorOption) error {
	return NewError(CodeInvalidArgument, message, StatusBadRequest, options...)
}

func ErrNotReady(message string, options ...ErrorOption) error {
	return NewError(CodeNotReady, message, StatusServiceUnavailable, options...)
}

func ErrRequestTimeout(message string, options ...ErrorOption) error {
	return NewError(CodeRequestTimeout, message, StatusGatewayTimeout, options...)
}

func ErrNotFound(code, message string, options ...ErrorOption) error {
	return NewError(code, message, StatusNotFound, options...)
}
