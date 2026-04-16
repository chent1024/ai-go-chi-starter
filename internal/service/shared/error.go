package shared

import "errors"

type Error struct {
	code       string
	message    string
	retryable  bool
	httpStatus int
	details    any
	err        error
}

type ErrorOption func(*Error)

func WithRetryable(retryable bool) ErrorOption {
	return func(target *Error) {
		target.retryable = retryable
	}
}

func WithDetails(details any) ErrorOption {
	return func(target *Error) {
		target.details = details
	}
}

func NewError(code, message string, httpStatus int, options ...ErrorOption) error {
	return Wrap(nil, code, message, httpStatus, options...)
}

func Wrap(err error, code, message string, httpStatus int, options ...ErrorOption) error {
	target := &Error{
		code:       code,
		message:    message,
		httpStatus: httpStatus,
		err:        err,
	}
	for _, option := range options {
		if option != nil {
			option(target)
		}
	}
	return target
}

func MarkRetryable(err error) error {
	if err == nil {
		return nil
	}
	options := []ErrorOption{WithRetryable(true)}
	if details := Details(err); details != nil {
		options = append(options, WithDetails(details))
	}
	return Wrap(err, Code(err), Message(err), HTTPStatus(err), options...)
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.message != "" {
		return e.message
	}
	if e.err != nil {
		return e.err.Error()
	}
	if e.code != "" {
		return e.code
	}
	return "unknown error"
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

func Code(err error) string {
	var target *Error
	if errors.As(err, &target) && target.code != "" {
		return target.code
	}
	return ""
}

func Message(err error) string {
	var target *Error
	if errors.As(err, &target) && target.message != "" {
		return target.message
	}
	if err != nil {
		return err.Error()
	}
	return ""
}

func Retryable(err error) bool {
	var target *Error
	return errors.As(err, &target) && target.retryable
}

func HTTPStatus(err error) int {
	var target *Error
	if errors.As(err, &target) && target.httpStatus != 0 {
		return target.httpStatus
	}
	return 0
}

func Details(err error) any {
	var target *Error
	if errors.As(err, &target) {
		return target.details
	}
	return nil
}
