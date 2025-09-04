package errors

import (
	"fmt"
	"net/http"
	"time"
)

// ─── HTTP ERROR HANDLING ─────────────────────────────────────────────

// HTTPError represents an error that can be directly returned as HTTP response
type HTTPError struct {
	Status    int            `json:"status"`
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Details   map[string]any `json:"details,omitempty"`
	Errors    []error        `json:"errors,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
	RequestID string         `json:"request_id,omitempty"`
	TraceID   string         `json:"trace_id,omitempty"`
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s - %s", e.Status, e.Code, e.Message)
}

// ToHTTPError converts any error to an HTTPError
func ToHTTPError(err error) *HTTPError {
	if err == nil {
		return nil
	}

	httpErr := &HTTPError{
		Status:    http.StatusInternalServerError,
		Code:      "INTERNAL_ERROR",
		Message:   "An internal error occurred",
		Details:   make(map[string]any),
		Timestamp: time.Now(),
	}

	// Handle different error types
	switch e := err.(type) {
	case *BusinessError:
		httpErr.Status = e.HTTPStatus
		httpErr.Code = e.Code
		httpErr.Message = e.Message
		httpErr.Details = e.Details

	case *RepositoryError:
		httpErr.Status = http.StatusInternalServerError
		httpErr.Code = e.Code
		httpErr.Message = "A database error occurred"
		// Don't expose internal database details

	case ValidationErrors:
		httpErr.Status = http.StatusBadRequest
		httpErr.Code = "VALIDATION_FAILED"
		httpErr.Message = "Validation failed"
		httpErr.Details["validation_errors"] = e.ToMap()

	case *ErrorCollection:
		if len(e.GetValidationErrors()) > 0 {
			httpErr.Status = http.StatusBadRequest
			httpErr.Code = "VALIDATION_FAILED"
			httpErr.Message = "Validation failed"
			httpErr.Details["validation_errors"] = e.GetValidationErrors().ToMap()
		} else {
			httpErr.Status = http.StatusInternalServerError
			httpErr.Code = "MULTIPLE_ERRORS"
			httpErr.Message = "Multiple errors occurred"
		}

	default:
		// Keep default internal server error
		httpErr.Details["original_error"] = err.Error()
	}

	return httpErr
}
