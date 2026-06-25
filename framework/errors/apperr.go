// Package apperr provides a unified, lightweight error system for the Awo Framework.
//
// # Design
//
// One concrete type ([Error]) covers every layer — business logic, repository,
// validation, HTTP transport — via fields rather than four separate structs.
// [ValidationErrors] is the only separate type; it has genuine collection
// semantics (multi-field, per-field lookup, map projection) and plugs into
// Error via [Error.Wrap].
//
// # Typical usage
//
//	// Business error with context
//	return apperr.New("ORG_UNIT_NOT_FOUND", "org unit not found").
//	    AsNotFound().
//	    WithDetail("org_unit_id", id).
//	    FromContext(ctx)
//
//	// Validation
//	var ve apperr.ValidationErrors
//	ve.Add("name", "required", "REQUIRED")
//	if ve.HasErrors() {
//	    return ve.AsError()
//	}
//
//	// Repository
//	return apperr.Repo("DB_QUERY_FAILED", "failed to fetch org unit", err).
//	    WithTable("org_units").
//	    WithOperation("SELECT")
//
//	// HTTP response
//	w.WriteHeader(e.HTTPStatus())
//	json.NewEncoder(w).Encode(e.HTTPBody(requestID))
//
// # Wrapping and unwrapping
//
// All [Error] values support [errors.As] and [errors.Is] through the standard
// [Error.Unwrap] chain. Use the package-level helpers ([IsCode], [IsCategory],
// [HTTPStatus], [AsHTTPBody]) instead of direct type switches to stay
// insulated from chain depth.
package apperr

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Enumerations
// ---------------------------------------------------------------------------

// Severity indicates the operational impact of an error.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

// Category classifies the layer where an error originated.
type Category string

const (
	CategoryValidation  Category = "validation"
	CategoryRepository  Category = "repository"
	CategoryBusiness    Category = "business"
	CategorySecurity    Category = "security"
	CategoryIntegration Category = "integration"
	CategorySystem      Category = "system"
	CategoryTenant      Category = "tenant"
)

// ---------------------------------------------------------------------------
// Core error type
// ---------------------------------------------------------------------------

// Error is the single concrete error type for the entire framework.
// Every layer (business, repository, validation, HTTP) uses this type;
// the Category field identifies the origin.
//
// Build errors through the constructors ([New], [Repo], [Security]) and the
// fluent builder methods — never construct the struct literal directly in
// application code, as zero values for HTTPStatus and Severity are invalid.
type Error struct {
	// Code is a machine-readable, UPPER_SNAKE_CASE identifier stable across
	// releases (e.g. "ORG_UNIT_NOT_FOUND"). Clients should switch on Code,
	// never on Message.
	Code string `json:"code"`

	// Message is a human-readable description safe to show in logs and API
	// responses. Must not contain PII or internal system details.
	Message string `json:"message"`

	// Category classifies the layer where the error originated.
	Category Category `json:"category"`

	// Severity is the operational impact level.
	Severity Severity `json:"severity"`

	// Details holds structured context (IDs, field values) that aids
	// debugging. Omitted from JSON when empty.
	Details map[string]any `json:"details,omitempty"`

	// Suggestions are optional human-readable hints for the caller on how to
	// resolve the error. Omitted from JSON when empty.
	Suggestions []string `json:"suggestions,omitempty"`

	// Retryable signals whether the operation is safe to retry as-is.
	Retryable bool `json:"retryable"`

	// TenantID and UserID are populated from context by [Error.FromContext].
	TenantID string `json:"tenant_id,omitempty"`
	UserID   string `json:"user_id,omitempty"`

	// httpStatus is intentionally unexported so callers use HTTPStatus(),
	// which guards against an uninitialised zero value.
	httpStatus int

	// cause is the underlying error this one wraps, accessible via Unwrap.
	cause error
}

// Error implements the error interface.
// Format: [category] <CODE> message (tenant: id): cause
func (e *Error) Error() string {
	var b strings.Builder
	if e.Category != "" {
		fmt.Fprintf(&b, "[%s] ", e.Category)
	}
	if e.Code != "" {
		fmt.Fprintf(&b, "<%s> ", e.Code)
	}
	b.WriteString(e.Message)
	if e.TenantID != "" {
		fmt.Fprintf(&b, " (tenant: %s)", e.TenantID)
	}
	if e.cause != nil {
		fmt.Fprintf(&b, ": %v", e.cause)
	}
	return b.String()
}

// Unwrap allows errors.As / errors.Is to traverse the cause chain.
func (e *Error) Unwrap() error { return e.cause }

// HTTPStatus returns the HTTP status code for this error.
// Falls back to 500 if the status was never set (guards against zero value).
func (e *Error) HTTPStatus() int {
	if e.httpStatus == 0 {
		return http.StatusInternalServerError
	}
	return e.httpStatus
}

// ---------------------------------------------------------------------------
// Fluent builder — enrichment methods
// ---------------------------------------------------------------------------

// WithDetail attaches a single key/value pair to Details and returns e for
// chaining.
//
//	apperr.New("NOT_FOUND", "org unit not found").WithDetail("id", id)
func (e *Error) WithDetail(key string, value any) *Error {
	if e.Details == nil {
		e.Details = make(map[string]any)
	}
	e.Details[key] = value
	return e
}

// WithSuggestion appends a human-readable resolution hint.
func (e *Error) WithSuggestion(s string) *Error {
	e.Suggestions = append(e.Suggestions, s)
	return e
}

// Wrap attaches a cause error and returns e for chaining.
// Use this instead of fmt.Errorf("%w", ...) when you also want to carry
// the structured fields of *Error.
func (e *Error) Wrap(cause error) *Error { e.cause = cause; return e }

// WithSeverity overrides the severity set by the constructor.
func (e *Error) WithSeverity(s Severity) *Error { e.Severity = s; return e }

// WithHTTPStatus overrides the HTTP status set by the constructor or shorthand
// methods (AsNotFound, AsUnauthorized, etc.).
func (e *Error) WithHTTPStatus(status int) *Error { e.httpStatus = status; return e }

// AsRetryable marks the operation as safe to retry without modification.
func (e *Error) AsRetryable() *Error { e.Retryable = true; return e }

// WithTable annotates a repository error with the table name.
// Stored in Details under the key "table".
func (e *Error) WithTable(t string) *Error { return e.WithDetail("table", t) }

// WithOperation annotates a repository error with the operation name
// (e.g. "SELECT", "INSERT"). Stored in Details under the key "operation".
func (e *Error) WithOperation(op string) *Error { return e.WithDetail("operation", op) }

// FromContext extracts TenantID, UserID, and RequestID from ctx and attaches
// them to the error. Safe to call with a nil or empty context.
func (e *Error) FromContext(ctx context.Context) *Error {
	if ctx == nil {
		return e
	}
	if id := tenantIDFromCtx(ctx); id != "" {
		e.TenantID = id
	}
	if id := userIDFromCtx(ctx); id != "" {
		e.UserID = id
	}
	if id := requestIDFromCtx(ctx); id != "" {
		e.WithDetail("request_id", id)
	}
	if op := operationFromCtx(ctx); op != "" {
		e.WithDetail("operation", op)
	}
	return e
}

// ---------------------------------------------------------------------------
// HTTP status shorthands — set both httpStatus and Severity in one call
// ---------------------------------------------------------------------------

// AsNotFound sets HTTP 404 / SeverityWarning.
func (e *Error) AsNotFound() *Error {
	e.httpStatus = http.StatusNotFound
	e.Severity = SeverityWarning
	return e
}

// AsConflict sets HTTP 409 / SeverityWarning.
func (e *Error) AsConflict() *Error {
	e.httpStatus = http.StatusConflict
	e.Severity = SeverityWarning
	return e
}

// AsUnauthorized sets HTTP 401 / SeverityError.
func (e *Error) AsUnauthorized() *Error {
	e.httpStatus = http.StatusUnauthorized
	e.Severity = SeverityError
	return e
}

// AsForbidden sets HTTP 403 / SeverityError.
func (e *Error) AsForbidden() *Error {
	e.httpStatus = http.StatusForbidden
	e.Severity = SeverityError
	return e
}

// AsBadRequest sets HTTP 400 / SeverityWarning.
func (e *Error) AsBadRequest() *Error {
	e.httpStatus = http.StatusBadRequest
	e.Severity = SeverityWarning
	return e
}

// AsUnprocessable sets HTTP 422 / SeverityWarning.
func (e *Error) AsUnprocessable() *Error {
	e.httpStatus = http.StatusUnprocessableEntity
	e.Severity = SeverityWarning
	return e
}

// AsServiceUnavailable sets HTTP 503 / SeverityCritical and marks the error
// as retryable.
func (e *Error) AsServiceUnavailable() *Error {
	e.httpStatus = http.StatusServiceUnavailable
	e.Severity = SeverityCritical
	e.Retryable = true
	return e
}

// ---------------------------------------------------------------------------
// HTTP transport serialisation
// ---------------------------------------------------------------------------

// HTTPBody is a JSON-serialisable envelope for HTTP error responses.
// Obtain one via [Error.ToHTTPBody] or the package-level [AsHTTPBody].
type HTTPBody struct {
	Status    int            `json:"status"`
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Details   map[string]any `json:"details,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
	RequestID string         `json:"request_id,omitempty"`
	TraceID   string         `json:"trace_id,omitempty"`
}

// ToHTTPBody produces the wire-safe JSON envelope for this error.
// requestID and traceID may be empty strings.
//
//	body := appErr.ToHTTPBody(r.Header.Get("X-Request-ID"), "")
//	w.WriteHeader(appErr.HTTPStatus())
//	json.NewEncoder(w).Encode(body)
func (e *Error) ToHTTPBody(requestID, traceID string) HTTPBody {
	body := HTTPBody{
		Status:    e.HTTPStatus(),
		Code:      e.Code,
		Message:   e.Message,
		Timestamp: time.Now().UTC(),
		RequestID: requestID,
		TraceID:   traceID,
	}
	if len(e.Details) > 0 {
		body.Details = e.Details
	}
	return body
}

// ---------------------------------------------------------------------------
// Constructors
// ---------------------------------------------------------------------------

// New creates a business-layer error with HTTP 400 / SeverityError defaults.
// Use the fluent methods to adjust status, severity, and details.
//
//	apperr.New("ORG_UNIT_NOT_FOUND", "org unit not found").AsNotFound()
func New(code, message string) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		Category:   CategoryBusiness,
		Severity:   SeverityError,
		httpStatus: http.StatusBadRequest,
	}
}

// NewWithContext is [New] + [Error.FromContext] in one call.
//
//	apperr.NewWithContext(ctx, "NOT_FOUND", "org unit not found").AsNotFound()
func NewWithContext(ctx context.Context, code, message string) *Error {
	return New(code, message).FromContext(ctx)
}

// Repo creates a repository-layer error wrapping cause.
// Defaults to HTTP 500 / SeverityCritical.
// The cause error is available via Unwrap and is NOT included in API responses
// (call ToHTTPBody to get a safe wire envelope).
//
//	apperr.Repo("DB_QUERY_FAILED", "failed to list org units", err).
//	    WithTable("org_units").WithOperation("SELECT")
func Repo(code, message string, cause error) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		Category:   CategoryRepository,
		Severity:   SeverityCritical,
		httpStatus: http.StatusInternalServerError,
		cause:      cause,
	}
}

// Security creates a security-layer error.
// Defaults to HTTP 403 / SeverityCritical.
//
//	apperr.Security("TENANT_MISMATCH", "access denied")
func Security(code, message string) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		Category:   CategorySecurity,
		Severity:   SeverityCritical,
		httpStatus: http.StatusForbidden,
	}
}

// System creates a system / infrastructure error.
// Defaults to HTTP 500 / SeverityCritical and is marked retryable.
//
//	apperr.System("CACHE_UNAVAILABLE", "cache is unreachable").Wrap(err)
func System(code, message string) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		Category:   CategorySystem,
		Severity:   SeverityCritical,
		Retryable:  true,
		httpStatus: http.StatusServiceUnavailable,
	}
}

// ---------------------------------------------------------------------------
// ValidationErrors — collection type for multi-field validation
// ---------------------------------------------------------------------------

// ValidationError represents a single field-level validation failure.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
	// Value holds the rejected input value. Set explicitly when the value is
	// safe to expose (no PII, no secrets).
	Value any `json:"value,omitempty"`
}

func (e ValidationError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Field, e.Message, e.Code)
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is an ordered list of field-level validation failures.
// Implements the error interface so it can be returned and passed to
// [AsHTTPBody] directly.
//
//	var ve apperr.ValidationErrors
//	ve.Add("name", "required", "REQUIRED")
//	ve.AddWithValue("email", "invalid format", "INVALID_EMAIL", rawEmail)
//	if ve.HasErrors() {
//	    return ve.AsError()
//	}
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	switch len(ve) {
	case 0:
		return "validation errors (0)" // non-empty so a nil-check distinguishes it
	case 1:
		return ve[0].Error()
	default:
		return fmt.Sprintf("validation failed with %d errors", len(ve))
	}
}

// Add appends a field error with an optional code.
func (ve *ValidationErrors) Add(field, message, code string) {
	*ve = append(*ve, ValidationError{Field: field, Message: message, Code: code})
}

// AddWithValue is like Add but also records the rejected value.
// Only call this when value contains no PII or secrets.
func (ve *ValidationErrors) AddWithValue(field, message, code string, value any) {
	*ve = append(*ve, ValidationError{Field: field, Message: message, Code: code, Value: value})
}

// HasErrors reports whether the collection is non-empty.
func (ve ValidationErrors) HasErrors() bool { return len(ve) > 0 }

// GetByField returns all errors for a specific field.
func (ve ValidationErrors) GetByField(field string) []ValidationError {
	var out []ValidationError
	for _, e := range ve {
		if e.Field == field {
			out = append(out, e)
		}
	}
	return out
}

// ToMap converts the errors to a map of field → []message, suitable for JSON
// API responses.
func (ve ValidationErrors) ToMap() map[string][]string {
	out := make(map[string][]string, len(ve))
	for _, e := range ve {
		out[e.Field] = append(out[e.Field], e.Message)
	}
	return out
}

// AsError wraps the collection in a structured [*Error] ready for return.
// The returned error has HTTP 422 and CategoryValidation.
// Returns nil if the collection is empty.
func (ve ValidationErrors) AsError() *Error {
	if !ve.HasErrors() {
		return nil
	}
	return &Error{
		Code:       "VALIDATION_FAILED",
		Message:    ve.Error(),
		Category:   CategoryValidation,
		Severity:   SeverityWarning,
		httpStatus: http.StatusUnprocessableEntity,
		Details:    map[string]any{"validation_errors": ve.ToMap()},
		cause:      ve, // preserves errors.As traversal
	}
}

// Unwrap implements the error interface so errors.As can reach individual
// ValidationError values inside a wrapped chain.
func (ve ValidationErrors) Unwrap() []error {
	out := make([]error, len(ve))
	for i, e := range ve {
		out[i] = e
	}
	return out
}

// ---------------------------------------------------------------------------
// Package-level helpers — work on any error in the chain
// ---------------------------------------------------------------------------

// IsCode reports whether any error in the chain is an *Error with the given
// code. Prefer this over direct type assertions.
//
//	if apperr.IsCode(err, "ORG_UNIT_NOT_FOUND") { ... }
func IsCode(err error, code string) bool {
	var e *Error
	return errors.As(err, &e) && e.Code == code
}

// IsCategory reports whether any error in the chain belongs to the given
// category.
func IsCategory(err error, cat Category) bool {
	var e *Error
	return errors.As(err, &e) && e.Category == cat
}

// IsRetryable reports whether the first *Error in the chain is marked
// retryable, or whether the error satisfies the net.Error Temporary interface.
func IsRetryable(err error) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.Retryable
	}
	// Honour net.Error and any other type that declares Temporary.
	type temporary interface{ Temporary() bool }
	var t temporary
	if errors.As(err, &t) {
		return t.Temporary()
	}
	return false
}

// IsValidation reports whether any error in the chain is a validation error
// (either *Error with CategoryValidation, or a ValidationErrors value).
func IsValidation(err error) bool {
	if IsCategory(err, CategoryValidation) {
		return true
	}
	var ve ValidationErrors
	return errors.As(err, &ve)
}

// HTTPStatus returns the HTTP status code appropriate for err.
// Returns 500 if err does not contain an *Error.
func HTTPStatus(err error) int {
	var e *Error
	if errors.As(err, &e) {
		return e.HTTPStatus()
	}
	var ve ValidationErrors
	if errors.As(err, &ve) {
		return http.StatusUnprocessableEntity
	}
	return http.StatusInternalServerError
}

// AsHTTPBody converts any error to an [HTTPBody] safe for wire serialisation.
// Internal details (table names, raw DB errors) from repository errors are
// replaced with a generic message so they never leak to clients.
//
// requestID and traceID may be empty strings.
//
//	body := apperr.AsHTTPBody(err, reqID, "")
//	w.WriteHeader(apperr.HTTPStatus(err))
//	json.NewEncoder(w).Encode(body)
func AsHTTPBody(err error, requestID, traceID string) HTTPBody {
	if err == nil {
		return HTTPBody{
			Status:    http.StatusOK,
			Code:      "OK",
			Message:   "OK",
			Timestamp: time.Now().UTC(),
			RequestID: requestID,
			TraceID:   traceID,
		}
	}

	base := HTTPBody{
		Status:    http.StatusInternalServerError,
		Code:      "INTERNAL_ERROR",
		Message:   "An internal error occurred",
		Timestamp: time.Now().UTC(),
		RequestID: requestID,
		TraceID:   traceID,
	}

	// Structured *Error — most specific match.
	var e *Error
	if errors.As(err, &e) {
		base.Status = e.HTTPStatus()
		base.Code = e.Code
		// Repository errors must never expose internal messages to clients.
		if e.Category == CategoryRepository {
			base.Message = "A data access error occurred"
		} else {
			base.Message = e.Message
			if len(e.Details) > 0 {
				base.Details = e.Details
			}
		}
		return base
	}

	// Bare ValidationErrors not yet wrapped in an *Error.
	var ve ValidationErrors
	if errors.As(err, &ve) {
		base.Status = http.StatusUnprocessableEntity
		base.Code = "VALIDATION_FAILED"
		base.Message = "Validation failed"
		base.Details = map[string]any{"validation_errors": ve.ToMap()}
		return base
	}

	// Unknown error — keep the safe default, add nothing that could leak.
	return base
}

// ---------------------------------------------------------------------------
// Context keys and helpers
// ---------------------------------------------------------------------------

// contextKey is an unexported type for context keys in this package.
// The exported constants below are of this unexported type, preventing
// accidental key collisions from other packages.
type contextKey string

const (
	CtxTenantID  contextKey = "tenant_id"
	CtxUserID    contextKey = "user_id"
	CtxRequestID contextKey = "request_id"
	CtxOperation contextKey = "operation"
)

// ContextWithTenantID returns a derived context carrying the tenant ID.
// Accepts both string and uuid.UUID values for flexibility.
//
//	ctx = apperr.ContextWithTenantID(ctx, tenantID.String())
func ContextWithTenantID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, CtxTenantID, id)
}

// ContextWithUserID returns a derived context carrying the user ID.
func ContextWithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, CtxUserID, id)
}

// ContextWithRequestID returns a derived context carrying the request ID.
func ContextWithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, CtxRequestID, id)
}

// ContextWithOperation returns a derived context carrying the operation name
// (e.g. "CreateOrgUnit"). Used to annotate errors emitted inside that operation.
func ContextWithOperation(ctx context.Context, op string) context.Context {
	return context.WithValue(ctx, CtxOperation, op)
}

// tenantIDFromCtx handles both uuid.UUID (stored by infrastructure layer) and
// plain string (stored by ContextWithTenantID). Both storage shapes are
// supported so callers are not forced to convert.
func tenantIDFromCtx(ctx context.Context) string {
	if id, ok := ctx.Value(CtxTenantID).(uuid.UUID); ok {
		return id.String()
	}
	id, _ := ctx.Value(CtxTenantID).(string)
	return id
}

func userIDFromCtx(ctx context.Context) string {
	if id, ok := ctx.Value(CtxUserID).(uuid.UUID); ok {
		return id.String()
	}
	id, _ := ctx.Value(CtxUserID).(string)
	return id
}

func requestIDFromCtx(ctx context.Context) string {
	id, _ := ctx.Value(CtxRequestID).(string)
	return id
}

func operationFromCtx(ctx context.Context) string {
	op, _ := ctx.Value(CtxOperation).(string)
	return op
}
