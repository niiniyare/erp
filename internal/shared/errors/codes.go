package errors

import (
	"errors"
	"fmt"
)

//
// ─── VALIDATION ERRORS ─────────────────────────────────────────────
//

// ValidationError represents a validation error with field details
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}
	if len(ve) == 1 {
		return ve[0].Error()
	}
	return fmt.Sprintf("validation failed with %d errors", len(ve))
}

//
// ─── REPOSITORY ERRORS ─────────────────────────────────────────────
//

// RepositoryError represents a failure in the repository layer
type RepositoryError struct {
	Code    string // e.g., "REJECT_FAILED"
	Message string // e.g., "Failed to reject access request"
	Err     error  // Underlying cause
}

func (e *RepositoryError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap allows errors.Is / errors.As support
func (e *RepositoryError) Unwrap() error {
	return e.Err
}

// NewRepositoryError creates a new RepositoryError
func NewRepositoryError(code, message string, err error) error {
	return &RepositoryError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// IsRepositoryErrorCode checks if the error is a RepositoryError with a specific code
func IsRepositoryErrorCode(err error, code string) bool {
	var repoErr *RepositoryError
	if errors.As(err, &repoErr) {
		return repoErr.Code == code
	}
	return false
}
