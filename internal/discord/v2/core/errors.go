package core

import "fmt"

// HandlerError represents an error that occurred during handler execution
type HandlerError struct {
	// The underlying error
	Err error

	// User-friendly message to display
	UserMessage string

	// Whether this error should be shown to the user
	ShowToUser bool

	// HTTP-like status code for categorization
	Code int
}

// Error implements the error interface
func (e *HandlerError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.UserMessage
}

// Unwrap returns the underlying error
func (e *HandlerError) Unwrap() error {
	return e.Err
}

// Common error codes
const (
	ErrorCodeBadRequest   = 400
	ErrorCodeUnauthorized = 401
	ErrorCodeForbidden    = 403
	ErrorCodeNotFound     = 404
	ErrorCodeConflict     = 409
	ErrorCodeInternal     = 500
)

// NewHandlerError creates a new handler error
func NewHandlerError(err error, userMessage string, code int) *HandlerError {
	return &HandlerError{
		Err:         err,
		UserMessage: userMessage,
		ShowToUser:  true,
		Code:        code,
	}
}

// NewInternalError creates an internal error that shouldn't be shown to users
func NewInternalError(err error) *HandlerError {
	return &HandlerError{
		Err:         err,
		UserMessage: "An internal error occurred. Please try again later.",
		ShowToUser:  true,
		Code:        ErrorCodeInternal,
	}
}

// NewUserError creates an error with a user-friendly message
func NewUserError(message string, code int) *HandlerError {
	return &HandlerError{
		UserMessage: message,
		ShowToUser:  true,
		Code:        code,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *HandlerError {
	return &HandlerError{
		UserMessage: fmt.Sprintf("%s not found", resource),
		ShowToUser:  true,
		Code:        ErrorCodeNotFound,
	}
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *HandlerError {
	return &HandlerError{
		UserMessage: message,
		ShowToUser:  true,
		Code:        ErrorCodeUnauthorized,
	}
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string) *HandlerError {
	return &HandlerError{
		UserMessage: message,
		ShowToUser:  true,
		Code:        ErrorCodeForbidden,
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string) *HandlerError {
	return &HandlerError{
		UserMessage: message,
		ShowToUser:  true,
		Code:        ErrorCodeBadRequest,
	}
}
