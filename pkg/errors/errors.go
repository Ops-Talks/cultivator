// Package errors defines custom error types and helper functions for error handling.
package errors

import (
	"fmt"
)

// ErrorType represents the type of error that occurred
type ErrorType string

const (
	// ErrTypeValidation represents validation errors.
	ErrTypeValidation ErrorType = "validation_error"

	// ErrTypeConfig represents configuration errors.
	ErrTypeConfig ErrorType = "config_error"

	// ErrTypeFile represents file operation errors.
	ErrTypeFile ErrorType = "file_error"

	// ErrTypeParse represents parse errors.
	ErrTypeParse ErrorType = "parse_error"

	// ErrTypeExternal represents external service errors.
	ErrTypeExternal ErrorType = "external_error"

	// ErrTypeSystem represents system errors.
	ErrTypeSystem ErrorType = "system_error"

	// ErrTypeNotFound represents not found errors.
	ErrTypeNotFound ErrorType = "not_found"

	// ErrTypeLock represents lock errors.
	ErrTypeLock ErrorType = "lock_error"
)

// CultivatorError is a custom error type for better error handling
type CultivatorError struct {
	Type      ErrorType
	Message   string
	OrigError error
	Context   map[string]interface{}
}

// Error implements the error interface
func (e *CultivatorError) Error() string {
	if e.OrigError != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.OrigError)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// WithContext adds context information to the error
func (e *CultivatorError) WithContext(key string, value interface{}) *CultivatorError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// Unwrap returns the underlying error
func (e *CultivatorError) Unwrap() error {
	return e.OrigError
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *CultivatorError {
	return &CultivatorError{
		Type:    ErrTypeValidation,
		Message: message,
	}
}

// NewConfigError creates a new configuration error
func NewConfigError(message string, err error) *CultivatorError {
	return &CultivatorError{
		Type:      ErrTypeConfig,
		Message:   message,
		OrigError: err,
	}
}

// NewParseError creates a new parse error
func NewParseError(context string, err error) *CultivatorError {
	return &CultivatorError{
		Type:      ErrTypeParse,
		Message:   fmt.Sprintf("failed to parse %s", context),
		OrigError: err,
	}
}

// NewFileError creates a new file operation error
func NewFileError(operation string, path string, err error) *CultivatorError {
	e := &CultivatorError{
		Type:      ErrTypeFile,
		Message:   fmt.Sprintf("failed to %s file", operation),
		OrigError: err,
	}
	return e.WithContext("path", path)
}

// NewExternalError creates a new external service error
func NewExternalError(service string, err error) *CultivatorError {
	return &CultivatorError{
		Type:      ErrTypeExternal,
		Message:   fmt.Sprintf("external service error: %s", service),
		OrigError: err,
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resourceType string, identifier string) *CultivatorError {
	e := &CultivatorError{
		Type:    ErrTypeNotFound,
		Message: fmt.Sprintf("%s not found", resourceType),
	}
	return e.WithContext("identifier", identifier)
}

// NewLockError creates a new lock error
func NewLockError(module string, lockedBy string, prNumber int) *CultivatorError {
	e := &CultivatorError{
		Type:    ErrTypeLock,
		Message: fmt.Sprintf("module %s is locked", module),
	}
	return e.
		WithContext("locked_by", lockedBy).
		WithContext("pr_number", prNumber)
}
