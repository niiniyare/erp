package errors

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// Problem is the shape every API error response is serialized as. It is
// loosely modelled on RFC 7807 (Problem Details for HTTP APIs) — "status"
// and "title"/"detail" map to that spec's fields — while adding the
// framework-specific Code/RequestID/TraceID fields the Awo API contract
// already relies on.
type Problem struct {
	Status    int            `json:"status"`               // HTTP status code
	Code      string         `json:"code"`                 // stable machine-readable Code, e.g. "iam.user.not_found"
	Title     string         `json:"title"`                // short human-readable summary (safe for end users)
	Detail    map[string]any `json:"detail,omitempty"`      // structured, sanitized context
	Errors    []FieldError   `json:"errors,omitempty"`      // field-level validation errors, if any
	Timestamp time.Time      `json:"timestamp"`
	RequestID string         `json:"request_id,omitempty"`
	TraceID   string         `json:"trace_id,omitempty"`
}

// ToProblem converts ANY error into a Problem suitable for an HTTP JSON
// response.
//
// Unlike the previous version of this package (which type-switched on
// concrete *BusinessError/*RepositoryError/*ErrorCollection types),
// ToProblem dispatches purely on the small interfaces defined in
// errors.go (Coded, HTTPStatuser, Detailer). This means it correctly
// handles:
//
//   - *AwoError values produced by this package,
//   - any module-defined error type that implements those interfaces
//     (e.g. a hand-written PumpOfflineError in the Forecourt module),
//   - errors wrapped with fmt.Errorf("...: %w", cause) at any depth,
//     because errors.As walks the full Unwrap() chain,
//   - and falls back to a safe generic 500 for anything else (a raw
//     error from a third-party library that was never translated).
//
// Repository/infrastructure-category errors are deliberately given a
// generic Title regardless of their real message, so a raw SQL error
// message can never leak to an API client — the real message still goes
// to your logs via err.Error(), just not into the Problem response.
func ToProblem(err error) *Problem {
	p := &Problem{
		Status:    http.StatusInternalServerError,
		Code:      string(CodeInternal),
		Title:     "an internal error occurred",
		Timestamp: time.Now(),
	}
	if err == nil {
		return p
	}

	// Field-level validation errors get their own shape (a list of
	// per-field problems) rather than a single Code/Title, since the
	// caller usually needs to fix multiple fields at once.
	var fieldErrs FieldErrors
	if errors.As(err, &fieldErrs) {
		p.Status = http.StatusBadRequest
		p.Code = string(CodeValidationFailed)
		p.Title = "validation failed"
		p.Errors = []FieldError(fieldErrs)
		return p
	}

	var coded Coded
	hasCoded := errors.As(err, &coded)
	if hasCoded {
		p.Code = string(coded.ErrorCode())
	}

	var statuser HTTPStatuser
	if errors.As(err, &statuser) {
		p.Status = statuser.HTTPStatus()
	}

	var categorized Categorized
	isRepositoryError := errors.As(err, &categorized) && categorized.ErrorCategory() == CategoryRepository

	switch {
	case isRepositoryError:
		// Never leak raw database/driver error text to API clients.
		p.Title = "a database error occurred"
	case hasCoded:
		p.Title = err.Error()
	default:
		// Unrecognised error type: keep the generic message set above,
		// don't guess at exposing err.Error() since we don't know its
		// shape (could be a raw driver error, a panic recovery value, ...).
	}

	var detailer Detailer
	if errors.As(err, &detailer) && !isRepositoryError {
		p.Detail = detailer.ErrorDetails()
	}

	return p
}

// WriteProblem is a small convenience helper for handlers that don't go
// through a shared error-handling middleware: it converts err to a
// Problem and writes it as a JSON response with the correct status code.
// Most Fiber/Gin/chi-based handlers in Awo ERP should prefer a central
// error-handling middleware built on ToProblem instead of calling this
// directly per-handler, but it's provided for standalone/simple cases
// (scripts, CLI-adjacent HTTP tools, tests).
func WriteProblem(w http.ResponseWriter, err error) {
	p := ToProblem(err)
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(p.Status)
	_ = json.NewEncoder(w).Encode(p)
}
