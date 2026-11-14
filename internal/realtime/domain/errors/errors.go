package errors

import (
	"errors"
	"fmt"
)

// Core domain error definitions
var (
	ErrNotFound       = errors.New("resource not found")
	ErrValidation     = errors.New("validation failed")
	ErrInternal       = errors.New("internal server error")
	ErrConflict       = errors.New("resource already exists")
	ErrUnauthorized   = errors.New("unauthorized access")
	ErrNotJoined      = errors.New("client has not joined a page")
	ErrInvalidProject = errors.New("invalid project context")
)

// Error codes for consistent machine-readable identification across WebSocket/API
const (
	CodeNotFound     = "NOT_FOUND"
	CodeValidation   = "VALIDATION_ERROR"
	CodeInternal     = "INTERNAL_ERROR"
	CodeUnauthorized = "UNAUTHORIZED"
	CodeConflict     = "CONFLICT"
	CodeJoinRequired = "JOIN_REQUIRED"
	CodeProcessError = "PROCESS_ERROR"
	CodeAccessDenied = "ACCESS_DENIED"
)

// AppError is a structured error that carries a domain-specific code
type AppError struct {
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new structured AppError
func New(code string, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Wrap helps in wrapping an existing error with a domain code
func Wrap(code string, err error) *AppError {
	if err == nil {
		return nil
	}
	return &AppError{
		Code:    code,
		Message: err.Error(),
		Err:     err,
	}
}

// IsCode checks if an error (or any error in its chain) matches a specific code
func IsCode(err error, code string) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}
