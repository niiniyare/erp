package errors

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// HTTPError is an error serialisable as an HTTP response body.
type HTTPError struct {
	Status    int            `json:"status"`
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Details   map[string]any `json:"details,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
	RequestID string         `json:"request_id,omitempty"`
	TraceID   string         `json:"trace_id,omitempty"`
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s - %s", e.Status, e.Code, e.Message)
}

// ToHTTPError converts any error to an HTTPError using errors.As for safe
// unwrapping through fmt.Errorf("%w", ...) chains.
func ToHTTPError(err error) *HTTPError {
	if err == nil {
		return nil
	}

	h := &HTTPError{
		Status:    http.StatusInternalServerError,
		Code:      "INTERNAL_ERROR",
		Message:   "An internal error occurred",
		Details:   make(map[string]any),
		Timestamp: time.Now(),
	}

	var be *BusinessError
	var re *RepositoryError
	var ec *ErrorCollection

	switch {
	case errors.As(err, &be):
		h.Status = be.HTTPStatus
		h.Code = be.Code
		h.Message = be.Message
		if be.Details != nil {
			h.Details = be.Details
		}

	case errors.As(err, &re):
		h.Status = http.StatusInternalServerError
		h.Code = re.Code
		h.Message = "A database error occurred"

	case errors.As(err, &ec):
		if len(ec.GetValidationErrors()) > 0 {
			h.Status = http.StatusBadRequest
			h.Code = "VALIDATION_FAILED"
			h.Message = "Validation failed"
			h.Details["validation_errors"] = ec.GetValidationErrors().ToMap()
		} else {
			h.Status = http.StatusInternalServerError
			h.Code = "MULTIPLE_ERRORS"
			h.Message = "Multiple errors occurred"
		}

	default:
		if ve, ok := err.(ValidationErrors); ok {
			h.Status = http.StatusBadRequest
			h.Code = "VALIDATION_FAILED"
			h.Message = "Validation failed"
			h.Details["validation_errors"] = ve.ToMap()
		} else {
			h.Details["original_error"] = err.Error()
		}
	}

	return h
}
