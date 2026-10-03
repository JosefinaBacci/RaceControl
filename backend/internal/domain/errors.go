package domain

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrAccountLocked      = errors.New("account temporarily locked after repeated failed logins")
	ErrUnauthenticated    = errors.New("authentication required")
	ErrForbidden          = errors.New("operation not allowed for this role")
	ErrNotFound           = errors.New("resource not found")
	ErrConflict           = errors.New("resource already exists")
	ErrSelfLockout        = errors.New("an administrator cannot deactivate or demote their own account")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{Field: field, Message: message}
}

type ConflictError struct {
	Field   string
	Message string
}

func (e *ConflictError) Error() string {
	return e.Field + ": " + e.Message
}

func (e *ConflictError) Is(target error) bool {
	return target == ErrConflict
}

func NewConflictError(field, message string) *ConflictError {
	return &ConflictError{Field: field, Message: message}
}
