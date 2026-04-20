package shared

func ErrInternal(message string, options ...ErrorOption) error {
	options = append([]ErrorOption{WithKind(KindInternal)}, options...)
	return NewError(CodeInternal, message, options...)
}

func ErrInvalidArgument(message string, options ...ErrorOption) error {
	options = append([]ErrorOption{WithKind(KindInvalidArgument)}, options...)
	return NewError(CodeInvalidArgument, message, options...)
}

func ErrNotReady(message string, options ...ErrorOption) error {
	options = append([]ErrorOption{WithKind(KindNotReady)}, options...)
	return NewError(CodeNotReady, message, options...)
}

func ErrRequestTimeout(message string, options ...ErrorOption) error {
	options = append([]ErrorOption{WithKind(KindRequestTimeout)}, options...)
	return NewError(CodeRequestTimeout, message, options...)
}

func ErrNotFound(code, message string, options ...ErrorOption) error {
	options = append([]ErrorOption{WithKind(KindNotFound)}, options...)
	return NewError(code, message, options...)
}
