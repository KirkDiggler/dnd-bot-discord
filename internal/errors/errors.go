package errors

import (
	"errors"
	"fmt"
)

// Code represents an error code for categorizing errors
type Code string

const (
	// CodeUnknown indicates an unknown error
	CodeUnknown Code = "unknown"

	// CodeInvalidArgument indicates client specified an invalid argument
	CodeInvalidArgument Code = "invalid_argument"

	// CodeNotFound indicates a requested resource was not found
	CodeNotFound Code = "not_found"

	// CodeAlreadyExists indicates an attempt to create a resource that already exists
	CodeAlreadyExists Code = "already_exists"

	// CodePermissionDenied indicates the caller does not have permission
	CodePermissionDenied Code = "permission_denied"

	// CodeUnauthenticated indicates the request does not have valid authentication
	CodeUnauthenticated Code = "unauthenticated"

	// CodeInternal indicates internal system error
	CodeInternal Code = "internal"

	// CodeUnavailable indicates the service is currently unavailable
	CodeUnavailable Code = "unavailable"

	// CodeUnimplemented indicates operation is not implemented
	CodeUnimplemented Code = "unimplemented"

	// CodeValidation indicates a validation error
	CodeValidation Code = "validation"
)

// Error represents an application error with code and metadata
type Error struct {
	// Code is the error code
	Code Code

	// Message is the error message
	Message string

	// Cause is the wrapped error
	Cause error

	// Meta contains additional context
	Meta map[string]any
}

// Error returns the error message
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the wrapped error
func (e *Error) Unwrap() error {
	return e.Cause
}

// WithMeta adds metadata to the error (builder pattern)
func (e *Error) WithMeta(key string, value any) *Error {
	if e.Meta == nil {
		e.Meta = make(map[string]any)
	}
	e.Meta[key] = value
	return e
}

// New creates a new error with the given code and message
func New(code Code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Newf creates a new error with formatted message
func Newf(code Code, format string, args ...any) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// Wrap wraps an error with additional context
func Wrap(err error, message string) *Error {
	if err == nil {
		return nil
	}

	// If it's already our error type, preserve the code
	var dndErr *Error
	if errors.As(err, &dndErr) {
		return &Error{
			Code:    dndErr.Code,
			Message: message,
			Cause:   err,
			Meta:    copyMeta(dndErr.Meta),
		}
	}

	// Otherwise, create unknown error
	return &Error{
		Code:    CodeUnknown,
		Message: message,
		Cause:   err,
	}
}

// Wrapf wraps an error with formatted message
func Wrapf(err error, format string, args ...any) *Error {
	if err == nil {
		return nil
	}
	return Wrap(err, fmt.Sprintf(format, args...))
}

// WrapWithCode wraps an error with a specific code
func WrapWithCode(err error, code Code, message string) *Error {
	if err == nil {
		return nil
	}

	wrapped := Wrap(err, message)
	wrapped.Code = code
	return wrapped
}

// Helper functions for common error types

// NotFound creates a not found error
func NotFound(message string) *Error {
	return New(CodeNotFound, message)
}

// NotFoundf creates a formatted not found error
func NotFoundf(format string, args ...any) *Error {
	return Newf(CodeNotFound, format, args...)
}

// InvalidArgument creates an invalid argument error
func InvalidArgument(message string) *Error {
	return New(CodeInvalidArgument, message)
}

// InvalidArgumentf creates a formatted invalid argument error
func InvalidArgumentf(format string, args ...any) *Error {
	return Newf(CodeInvalidArgument, format, args...)
}

// AlreadyExists creates an already exists error
func AlreadyExists(message string) *Error {
	return New(CodeAlreadyExists, message)
}

// AlreadyExistsf creates a formatted already exists error
func AlreadyExistsf(format string, args ...any) *Error {
	return Newf(CodeAlreadyExists, format, args...)
}

// Internal creates an internal error
func Internal(message string) *Error {
	return New(CodeInternal, message)
}

// Internalf creates a formatted internal error
func Internalf(format string, args ...any) *Error {
	return Newf(CodeInternal, format, args...)
}

// Validation creates a validation error
func Validation(message string) *Error {
	return New(CodeValidation, message)
}

// Validationf creates a formatted validation error
func Validationf(format string, args ...any) *Error {
	return Newf(CodeValidation, format, args...)
}

// Error checking functions

// Is checks if the error is of a specific code
func Is(err error, code Code) bool {
	var dndErr *Error
	if errors.As(err, &dndErr) {
		return dndErr.Code == code
	}
	return false
}

// IsNotFound checks if the error is a not found error
func IsNotFound(err error) bool {
	return Is(err, CodeNotFound)
}

// IsInvalidArgument checks if the error is an invalid argument error
func IsInvalidArgument(err error) bool {
	return Is(err, CodeInvalidArgument)
}

// IsAlreadyExists checks if the error is an already exists error
func IsAlreadyExists(err error) bool {
	return Is(err, CodeAlreadyExists)
}

// IsInternal checks if the error is an internal error
func IsInternal(err error) bool {
	return Is(err, CodeInternal)
}

// IsValidation checks if the error is a validation error
func IsValidation(err error) bool {
	return Is(err, CodeValidation)
}

// GetCode returns the error code
func GetCode(err error) Code {
	var dndErr *Error
	if errors.As(err, &dndErr) {
		return dndErr.Code
	}
	return CodeUnknown
}

// GetMeta returns the error metadata
func GetMeta(err error) map[string]any {
	var dndErr *Error
	if errors.As(err, &dndErr) {
		return dndErr.Meta
	}
	return nil
}

// copyMeta creates a copy of the metadata map
func copyMeta(meta map[string]any) map[string]any {
	if meta == nil {
		return nil
	}
	
	copied := make(map[string]any, len(meta))
	for k, v := range meta {
		copied[k] = v
	}
	return copied
}