package errors

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

//  HTTP ERROR HANDLING

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

// ToHTTPError converts any error to an HTTPError.
// It uses errors.As so that errors wrapped with fmt.Errorf("%w", ...) are
// correctly unwrapped — e.g. a *BusinessError buried inside a service-layer
// fmt.Errorf still gets the right HTTP status and code.
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

	var be *BusinessError
	var re *RepositoryError
	var ec *ErrorCollection

	switch {
	case errors.As(err, &be):
		httpErr.Status = be.HTTPStatus
		httpErr.Code = be.Code
		httpErr.Message = be.Message
		if be.Details != nil {
			httpErr.Details = be.Details
		}

	case errors.As(err, &re):
		httpErr.Status = http.StatusInternalServerError
		httpErr.Code = re.Code
		httpErr.Message = "A database error occurred"
		// Don't expose internal database details

	case errors.As(err, &ec):
		if len(ec.GetValidationErrors()) > 0 {
			httpErr.Status = http.StatusBadRequest
			httpErr.Code = "VALIDATION_FAILED"
			httpErr.Message = "Validation failed"
			httpErr.Details["validation_errors"] = ec.GetValidationErrors().ToMap()
		} else {
			httpErr.Status = http.StatusInternalServerError
			httpErr.Code = "MULTIPLE_ERRORS"
			httpErr.Message = "Multiple errors occurred"
		}

	default:
		// Check for ValidationErrors (slice type — errors.As doesn't work on non-pointer types)
		if ve, ok := err.(ValidationErrors); ok {
			httpErr.Status = http.StatusBadRequest
			httpErr.Code = "VALIDATION_FAILED"
			httpErr.Message = "Validation failed"
			httpErr.Details["validation_errors"] = ve.ToMap()
		} else {
			httpErr.Details["original_error"] = err.Error()
		}
	}

	return httpErr
}
