package shared

import "errors"

type Error struct {
	code      string
	message   string
	kind      Kind
	retryable bool
	details   any
	err       error
}

type Kind string

const (
	KindInternal        Kind = "internal"
	KindInvalidArgument Kind = "invalid_argument"
	KindNotReady        Kind = "not_ready"
	KindRequestTimeout  Kind = "request_timeout"
	KindNotFound        Kind = "not_found"
)

type ErrorOption func(*Error)

func WithKind(kind Kind) ErrorOption {
	return func(target *Error) {
		target.kind = kind
	}
}

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

func NewError(code, message string, options ...ErrorOption) error {
	return Wrap(nil, code, message, options...)
}

func Wrap(err error, code, message string, options ...ErrorOption) error {
	target := &Error{
		code:    code,
		message: message,
		err:     err,
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
	if kind := KindOf(err); kind != "" {
		options = append(options, WithKind(kind))
	}
	return Wrap(err, Code(err), Message(err), options...)
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

func KindOf(err error) Kind {
	var target *Error
	if errors.As(err, &target) {
		return target.kind
	}
	return ""
}

func Details(err error) any {
	var target *Error
	if errors.As(err, &target) {
		return target.details
	}
	return nil
}
