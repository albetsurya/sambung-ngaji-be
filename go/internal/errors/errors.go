package errors

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound = errors.New("resource not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden = errors.New("forbidden")
	ErrConflict = errors.New("conflict")
	ErrValidation = errors.New("validation error")
	ErrInternal = errors.New("internal error")
	ErrTimeout = errors.New("timeout")
	ErrUnavailable = errors.New("service unavailable")
)

func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(format+": %w", append(args, err)...)
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}

func IsValidation(err error) bool {
	return errors.Is(err, ErrValidation)
}

func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}

func IsUnavailable(err error) bool {
	return errors.Is(err, ErrUnavailable)
}
