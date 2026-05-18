package ui

import (
	"errors"
	"fmt"
)

// Sentinel errors for the UI pipeline. Use errors.Is to check.
var (
	// ErrPageNotFound is returned by RegistryStage when no PageFn is
	// registered for the requested route. Maps to HTTP 404.
	ErrPageNotFound = errors.New("ui: page not found")

	// ErrUnauthenticated is returned by SessionStage when the Go context
	// does not carry a valid contract.SessionContext. Maps to HTTP 401.
	ErrUnauthenticated = errors.New("ui: unauthenticated")

	// ErrSchemaInvalid is returned by NormalizeStage or ValidateStage when
	// the compiled schema violates an AMIS compliance rule or contains a
	// forbidden expression. Maps to HTTP 500 (programmer error, not user error).
	ErrSchemaInvalid = errors.New("ui: schema invalid")

	// ErrPermissionResolution is returned by AuthzStage when the UIAuthzService
	// fails to resolve permissions. Maps to HTTP 503.
	ErrPermissionResolution = errors.New("ui: permission resolution failed")
)

// PageNotFoundError carries route details for logging.
type PageNotFoundError struct {
	Route string
}

func (e *PageNotFoundError) Error() string {
	return fmt.Sprintf("ui: page not found for route %q", e.Route)
}

func (e *PageNotFoundError) Is(target error) bool {
	return target == ErrPageNotFound
}

// SchemaValidationError carries the rule name and schema path for debugging.
type SchemaValidationError struct {
	Rule    string
	Path    string
	Message string
}

func (e *SchemaValidationError) Error() string {
	return fmt.Sprintf("ui: schema validation [%s] at %s: %s", e.Rule, e.Path, e.Message)
}

func (e *SchemaValidationError) Is(target error) bool {
	return target == ErrSchemaInvalid
}

// IsPageNotFound returns true when err wraps ErrPageNotFound.
func IsPageNotFound(err error) bool {
	return errors.Is(err, ErrPageNotFound)
}

// IsUnauthenticated returns true when err wraps ErrUnauthenticated.
func IsUnauthenticated(err error) bool {
	return errors.Is(err, ErrUnauthenticated)
}

// IsSchemaInvalid returns true when err wraps ErrSchemaInvalid.
func IsSchemaInvalid(err error) bool {
	return errors.Is(err, ErrSchemaInvalid)
}
