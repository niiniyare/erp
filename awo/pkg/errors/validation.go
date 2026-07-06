package errors

import "fmt"

// FieldError represents a single field-level validation failure — e.g.
// "email: must be a valid email address (invalid_email)". This is a
// different concern from AwoError/Code: a whole HTTP request typically
// fails validation with SEVERAL FieldErrors at once (see FieldErrors
// below), each pointing at one form field.
type FieldError struct {
	Field   string `json:"field"`           // e.g. "email", "line_items[2].quantity"
	Message string `json:"message"`         // human-readable explanation
	Code    string `json:"code,omitempty"`  // one of the Field* constants in codes.go
	Value   any    `json:"value,omitempty"` // the offending value, sanitized
}

// Error implements the standard error interface for a single FieldError,
// so a lone FieldError can be returned/wrapped like any other error.
func (e FieldError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Field, e.Message, e.Code)
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// FieldErrors is a collection of FieldError, typically representing every
// validation failure found across a single request payload so the caller
// can fix them all at once instead of one-at-a-time round trips.
type FieldErrors []FieldError

// Error implements the standard error interface. For a single error it
// reads naturally; for several, it reports a count (the full detail lives
// in ToProblem(err).Errors for API responses, or in each FieldError for
// programmatic access via e.g. range).
func (fe FieldErrors) Error() string {
	switch len(fe) {
	case 0:
		return "no validation errors"
	case 1:
		return fe[0].Error()
	default:
		return fmt.Sprintf("validation failed with %d field errors", len(fe))
	}
}

// ErrorCode implements Coded so FieldErrors integrates with ToProblem,
// Is, IsCategory, etc. exactly like an AwoError would.
func (fe FieldErrors) ErrorCode() Code { return CodeValidationFailed }

// ErrorCategory implements Categorized.
func (fe FieldErrors) ErrorCategory() Category { return CategoryValidation }

// HTTPStatus implements HTTPStatuser.
func (fe FieldErrors) HTTPStatus() int { return 400 }

// Add appends a plain field error with no code or value.
func (fe *FieldErrors) Add(field, message string) {
	*fe = append(*fe, FieldError{Field: field, Message: message})
}

// AddWithCode appends a field error tagged with one of the Field*
// constants from codes.go (e.g. FieldRequired, FieldInvalidEmail).
func (fe *FieldErrors) AddWithCode(field, message, code string) {
	*fe = append(*fe, FieldError{Field: field, Message: message, Code: code})
}

// AddWithValue appends a field error that also records the offending
// value, passed through sanitizeValue first so secrets never end up in a
// validation error response by accident.
func (fe *FieldErrors) AddWithValue(field, message, code string, value any) {
	*fe = append(*fe, FieldError{
		Field:   field,
		Message: message,
		Code:    code,
		// sanitizeDetail also checks the field NAME (e.g. a field literally
		// called "password"), not just the value's content — see errors.go.
		Value: sanitizeDetail(field, value),
	})
}

// HasErrors reports whether any field errors have been recorded. Typical
// use at the end of a validation function:
//
//	var fe errors.FieldErrors
//	if req.Email == "" {
//	    fe.AddWithCode("email", "email is required", errors.FieldRequired)
//	}
//	if fe.HasErrors() {
//	    return fe
//	}
func (fe FieldErrors) HasErrors() bool {
	return len(fe) > 0
}

// ByField returns only the errors recorded against a specific field —
// useful for rendering per-field inline errors in a form-driven SDUI
// client.
func (fe FieldErrors) ByField(field string) []FieldError {
	var out []FieldError
	for _, e := range fe {
		if e.Field == field {
			out = append(out, e)
		}
	}
	return out
}

// ToMap groups messages by field name, e.g. for simple key->[]string
// JSON shapes some frontend form libraries expect.
func (fe FieldErrors) ToMap() map[string][]string {
	out := make(map[string][]string, len(fe))
	for _, e := range fe {
		out[e.Field] = append(out[e.Field], e.Message)
	}
	return out
}
