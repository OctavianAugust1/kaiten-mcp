package domain

import (
	"fmt"
	"time"
)

// ErrorKind is a stable application-facing failure category.
type ErrorKind string

const (
	ErrorInvalidArgument   ErrorKind = "invalid_argument"
	ErrorUnauthorized      ErrorKind = "unauthorized"
	ErrorForbidden         ErrorKind = "forbidden"
	ErrorNotFound          ErrorKind = "not_found"
	ErrorRateLimited       ErrorKind = "rate_limited"
	ErrorTimeout           ErrorKind = "timeout"
	ErrorCancelled         ErrorKind = "cancelled"
	ErrorMalformedResponse ErrorKind = "malformed_response"
	ErrorTemporaryUpstream ErrorKind = "temporary_upstream"
	ErrorUpstream          ErrorKind = "upstream"
)

// Error carries only bounded, non-secret metadata across application layers.
type Error struct {
	Kind       ErrorKind
	Operation  string
	Entity     string
	StatusCode int
	RetryAfter time.Duration
	Cause      error
}

// Error returns a stable safe description and intentionally omits Cause.
func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.StatusCode != 0 {
		return fmt.Sprintf("%s: %s (status %d)", e.Operation, e.Kind, e.StatusCode)
	}
	return fmt.Sprintf("%s: %s", e.Operation, e.Kind)
}

// Unwrap preserves cancellation and deadline checks without printing the
// underlying technical error.
func (e *Error) Unwrap() error {
	return e.Cause
}
