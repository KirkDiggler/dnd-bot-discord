package internal

import "fmt"

type BaseError string

func (e BaseError) Error() string {
	return string(e)
}

const (
	ErrMissingParam BaseError = "missing parameter"
	ErrInvalidParam BaseError = "invalid parameter"
	ErrNotFound     BaseError = "not found"
)

type ErrorWrapper struct {
	Err     error
	Message string
}

func (e *ErrorWrapper) Error() string {
	return fmt.Sprintf("%v: %s", e.Err, e.Message)
}

func (e *ErrorWrapper) Unwrap() error {
	return e.Err
}

func NewMissingParamError(param string) error {
	return &ErrorWrapper{
		Err:     ErrMissingParam,
		Message: param,
	}
}

func NewInvalidParamError(msg string) error {
	return &ErrorWrapper{
		Err:     ErrInvalidParam,
		Message: msg,
	}
}
