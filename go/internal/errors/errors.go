package errors

import (
	"errors"
	"fmt"
)

var (
	// ErrNotFound is returned when a resource is not found
	ErrNotFound = errors.New("resource not found")
	// ErrUnauthorized is returned when authentication fails
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden is returned when access is denied
	ErrForbidden = errors.New("forbidden")
	// ErrConflict is returned when there's a conflict (e.g., duplicate key)
	ErrConflict = errors.New("conflict")
	// ErrValidation is returned when input validation fails
	ErrValidation = errors.New("validation error")
	// ErrInternal is returned for internal server errors
	ErrInternal = errors.New("internal error")
	// ErrTimeout is returned when an operation times out
	ErrTimeout = errors.New("timeout")
	// ErrUnavailable is returned when a service is unavailable
	ErrUnavailable = errors.New("service unavailable")
)

// Wrap adds context to an error while preserving the original error for unwrapping
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// Wrapf adds formatted context to an error while preserving the original error
func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(format+": %w", append(args, err)...)
}

// IsNotFound checks if the error is a not found error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsUnauthorized checks if the error is an unauthorized error
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsForbidden checks if the error is a forbidden error
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// IsConflict checks if the error is a conflict error
func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}

// IsValidation checks if the error is a validation error
func IsValidation(err error) bool {
	return errors.Is(err, ErrValidation)
}

// IsTimeout checks if the error is a timeout error
func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}

// IsUnavailable checks if the error is an unavailable error
func IsUnavailable(err error) bool {
	return errors.Is(err, ErrUnavailable)
}
