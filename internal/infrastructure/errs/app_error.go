package errs

import (
	"errors"
	"fmt"
)

var (
	Internal        = &AppError{Code: CodeInternal}
	InvalidArgument = &AppError{Code: CodeInvalidArgument}
	NotFound        = &AppError{Code: CodeNotFound}
	AlreadyExists   = &AppError{Code: CodeAlreadyExists}
	Conflict        = &AppError{Code: CodeConflict}
	Unauthorized    = &AppError{Code: CodeUnauthorized}
	Forbidden       = &AppError{Code: CodeForbidden}
	Unavailable     = &AppError{Code: CodeUnavailable}
	Timeout         = &AppError{Code: CodeTimeout}
	RateLimited     = &AppError{Code: CodeRateLimited}
)

type AppError struct {
	Code    Code
	Message string
	Op      string
	Err     error
}

func New(code Code, msg string) *AppError {
	return &AppError{Code: code, Message: msg}
}

func (e *AppError) Error() string {
	switch {
	case e.Op != "" && e.Err != nil:
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Message, e.Err)
	case e.Op != "":
		return fmt.Sprintf("%s: %s", e.Op, e.Message)
	case e.Err != nil:
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	default:
		return e.Message
	}
}

func (e *AppError) Unwrap() error { return e.Err }

// Is сравнивает по коду — это даёт errors.Is(err, errs.NotFound)
// независимо от Message и Op.
func (e *AppError) Is(target error) bool {
	var t *AppError
	if !errors.As(target, &t) {
		return false
	}
	return e.Code == t.Code
}

// Retryable — для транспортов очередей (Kafka, RabbitMQ).
func (e *AppError) Retryable() bool {
	switch e.Code {
	case CodeUnavailable, CodeTimeout:
		return true
	default:
		return false
	}
}

func (e *AppError) WithMessage(msg string) *AppError {
	return &AppError{Code: e.Code, Message: msg, Op: e.Op, Err: e.Err}
}

func (e *AppError) WithOp(op string) *AppError {
	return &AppError{Code: e.Code, Message: e.Message, Op: op, Err: e.Err}
}

func (e *AppError) Wrap(err error) *AppError {
	return &AppError{Code: e.Code, Message: e.Message, Op: e.Op, Err: err}
}

// Wrap превращает любую ошибку в AppError.
// Если err уже AppError — возвращает как есть.
func Wrap(err error) *AppError {
	if err == nil {
		return nil
	}
	if appErr, ok := AsAppError(err); ok {
		return appErr
	}
	return &AppError{Code: CodeInternal, Message: "internal error", Err: err}
}

func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
