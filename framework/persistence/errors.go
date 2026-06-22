package persistence

import "errors"

// Sentinel errors returned by EntityStore implementations.
var (
	// ErrNotFound is returned when a record does not exist (or is soft-deleted).
	// Handlers should map this to HTTP 404.
	ErrNotFound = errors.New("record not found")

	// ErrConflict is returned on unique-constraint violations.
	// Handlers should map this to HTTP 409.
	ErrConflict = errors.New("record conflict")

	// ErrInvalidTenant is returned when the tenant ID is zero or not active.
	ErrInvalidTenant = errors.New("invalid tenant")

	// ErrUnknownEntity is returned when ForEntity is called with an unregistered name.
	ErrUnknownEntity = errors.New("unknown entity")
)

// IsNotFound reports whether err wraps ErrNotFound.
func IsNotFound(err error) bool { return errors.Is(err, ErrNotFound) }

// IsConflict reports whether err wraps ErrConflict.
func IsConflict(err error) bool { return errors.Is(err, ErrConflict) }
