// Package errors provides generic error types for the Awo Framework.
// Applications build domain-specific errors on top of these types.
package errors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// CONTEXT KEYS

type contextKey string

const (
	TenantIDKey  contextKey = "tenant_id"
	UserIDKey    contextKey = "user_id"
	RequestIDKey contextKey = "request_id"
	OperationKey contextKey = "operation"
)

// ERROR SEVERITY LEVELS

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

// ERROR CATEGORIES

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

// BusinessError represents domain-specific business logic errors.
type BusinessError struct {
	Code        string         `json:"code"`
	Message     string         `json:"message"`
	Details     map[string]any `json:"details,omitempty"`
	Suggestions []string       `json:"suggestions,omitempty"`
	HTTPStatus  int            `json:"-"`
	Severity    Severity       `json:"severity"`
	Category    Category       `json:"category"`
	TenantID    string         `json:"tenant_id,omitempty"`
	UserID      string         `json:"user_id,omitempty"`
	Retryable   bool           `json:"retryable"`
	Err         error          `json:"-"`
}

func (e *BusinessError) Error() string {
	var parts []string
	if e.Category != "" {
		parts = append(parts, fmt.Sprintf("[%s]", e.Category))
	}
	if e.Code != "" {
		parts = append(parts, fmt.Sprintf("<%s>", e.Code))
	}
	if e.Message != "" {
		parts = append(parts, e.Message)
	}
	if e.TenantID != "" {
		parts = append(parts, fmt.Sprintf("(tenant: %s)", e.TenantID))
	}
	result := strings.Join(parts, " ")
	if e.Err != nil {
		result += fmt.Sprintf(": %v", e.Err)
	}
	return result
}

func (e *BusinessError) Unwrap() error { return e.Err }

func (e *BusinessError) WithDetail(key string, value any) *BusinessError {
	if e.Details == nil {
		e.Details = make(map[string]any)
	}
	e.Details[key] = value
	return e
}

func (e *BusinessError) WithCause(cause error) *BusinessError   { e.Err = cause; return e }
func (e *BusinessError) WithCategory(c Category) *BusinessError { e.Category = c; return e }
func (e *BusinessError) WithSeverity(s Severity) *BusinessError { e.Severity = s; return e }
func (e *BusinessError) WithHTTPStatus(status int) *BusinessError {
	e.HTTPStatus = status
	return e
}
func (e *BusinessError) WithSuggestion(s string) *BusinessError {
	e.Suggestions = append(e.Suggestions, s)
	return e
}
func (e *BusinessError) IsRetryable() bool { return e.Retryable }

func (e *BusinessError) MarshalJSON() ([]byte, error) {
	type alias BusinessError
	return json.Marshal(&struct {
		*alias
		ErrorMessage string `json:"error_message"`
	}{alias: (*alias)(e), ErrorMessage: e.Error()})
}

// RepositoryError represents a failure in the repository layer.
type RepositoryError struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Err       error          `json:"-"`
	Details   map[string]any `json:"details,omitempty"`
	Operation string         `json:"operation,omitempty"`
	Table     string         `json:"table,omitempty"`
	TenantID  string         `json:"tenant_id,omitempty"`
}

func (e *RepositoryError) Error() string {
	var parts []string
	if e.Code != "" {
		parts = append(parts, fmt.Sprintf("[%s]", e.Code))
	}
	if e.Message != "" {
		parts = append(parts, e.Message)
	}
	if e.Operation != "" && e.Table != "" {
		parts = append(parts, fmt.Sprintf("(operation: %s, table: %s)", e.Operation, e.Table))
	}
	if e.TenantID != "" {
		parts = append(parts, fmt.Sprintf("tenant: %s", e.TenantID))
	}
	result := strings.Join(parts, " ")
	if e.Err != nil {
		result += fmt.Sprintf(": %v", e.Err)
	}
	return result
}

func (e *RepositoryError) Unwrap() error { return e.Err }

func (e *RepositoryError) WithDetail(key string, value any) *RepositoryError {
	if e.Details == nil {
		e.Details = make(map[string]any)
	}
	e.Details[key] = value
	return e
}
func (e *RepositoryError) WithOperation(op string) *RepositoryError { e.Operation = op; return e }
func (e *RepositoryError) WithTable(t string) *RepositoryError      { e.Table = t; return e }
func (e *RepositoryError) WithTenant(id string) *RepositoryError    { e.TenantID = id; return e }

func (e *RepositoryError) MarshalJSON() ([]byte, error) {
	type alias RepositoryError
	return json.Marshal(&struct {
		*alias
		ErrorMessage string `json:"error_message"`
	}{alias: (*alias)(e), ErrorMessage: e.Error()})
}

// ValidationError represents a single field validation failure.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
	Value   any    `json:"value,omitempty"`
}

func (e ValidationError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Field, e.Message, e.Code)
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation failures.
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

func (ve *ValidationErrors) Add(field, message string) {
	*ve = append(*ve, ValidationError{Field: field, Message: message})
}

func (ve *ValidationErrors) AddWithCode(field, message, code string) {
	*ve = append(*ve, ValidationError{Field: field, Message: message, Code: code})
}

func (ve ValidationErrors) HasErrors() bool { return len(ve) > 0 }

func (ve ValidationErrors) GetByField(field string) []ValidationError {
	var out []ValidationError
	for _, err := range ve {
		if err.Field == field {
			out = append(out, err)
		}
	}
	return out
}

func (ve ValidationErrors) ToMap() map[string][]string {
	result := make(map[string][]string)
	for _, err := range ve {
		result[err.Field] = append(result[err.Field], err.Message)
	}
	return result
}

// ErrorCollection aggregates multiple errors with context.
type ErrorCollection struct {
	Errors    []error        `json:"errors"`
	Context   map[string]any `json:"context,omitempty"`
	Operation string         `json:"operation,omitempty"`
	TenantID  string         `json:"tenant_id,omitempty"`
	RequestID string         `json:"request_id,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
	Severity  Severity       `json:"severity"`
	Category  Category       `json:"category"`
}

func (ec *ErrorCollection) Error() string {
	if len(ec.Errors) == 0 {
		return "no errors"
	}
	if len(ec.Errors) == 1 {
		return ec.Errors[0].Error()
	}
	var msgs []string
	for _, err := range ec.Errors {
		msgs = append(msgs, err.Error())
	}
	return fmt.Sprintf("multiple errors occurred: [%s]", strings.Join(msgs, "; "))
}

func (ec *ErrorCollection) Add(err error) {
	if err != nil {
		ec.Errors = append(ec.Errors, err)
	}
}

func (ec *ErrorCollection) HasErrors() bool { return len(ec.Errors) > 0 }
func (ec *ErrorCollection) Count() int      { return len(ec.Errors) }

func (ec *ErrorCollection) GetValidationErrors() ValidationErrors {
	var out ValidationErrors
	for _, err := range ec.Errors {
		if ve, ok := err.(ValidationError); ok {
			out = append(out, ve)
		} else if ves, ok := err.(ValidationErrors); ok {
			out = append(out, ves...)
		}
	}
	return out
}

// CONSTRUCTORS

func NewBusinessError(code, message string) *BusinessError {
	return &BusinessError{
		Code:       code,
		Message:    message,
		Details:    make(map[string]any),
		HTTPStatus: http.StatusBadRequest,
		Severity:   SeverityError,
		Category:   CategoryBusiness,
	}
}

func NewBusinessErrorWithContext(ctx context.Context, code, message string) *BusinessError {
	err := NewBusinessError(code, message)
	if id := getTenantIDFromContext(ctx); id != "" {
		err.TenantID = id
	}
	if id := getUserIDFromContext(ctx); id != "" {
		err.UserID = id
	}
	return err
}

func NewRepositoryError(code, message string, cause error) *RepositoryError {
	return &RepositoryError{
		Code:    code,
		Message: message,
		Err:     cause,
		Details: make(map[string]any),
	}
}

func NewErrorCollection(operation string) *ErrorCollection {
	return &ErrorCollection{
		Errors:    make([]error, 0),
		Context:   make(map[string]any),
		Operation: operation,
		Timestamp: time.Now(),
		Severity:  SeverityInfo,
		Category:  CategorySystem,
	}
}

// UTILITY FUNCTIONS

func GetErrorCode(err error) string {
	switch e := err.(type) {
	case *BusinessError:
		return e.Code
	case *RepositoryError:
		return e.Code
	case ValidationError:
		return e.Code
	default:
		return "UNKNOWN_ERROR"
	}
}

func GetHTTPStatus(err error) int {
	switch e := err.(type) {
	case *BusinessError:
		return e.HTTPStatus
	case ValidationErrors:
		return http.StatusBadRequest
	case *RepositoryError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func IsBusinessErrorCode(err error, code string) bool {
	var be *BusinessError
	return errors.As(err, &be) && be.Code == code
}

func IsRepositoryErrorCode(err error, code string) bool {
	var re *RepositoryError
	return errors.As(err, &re) && re.Code == code
}

func IsTemporary(err error) bool {
	if be, ok := err.(*BusinessError); ok {
		return be.Retryable
	}
	s := strings.ToLower(err.Error())
	for _, p := range []string{"timeout", "connection reset", "connection refused", "temporary failure", "service unavailable", "rate limit"} {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}

func IsValidationError(err error) bool {
	if _, ok := err.(ValidationErrors); ok {
		return true
	}
	if _, ok := err.(ValidationError); ok {
		return true
	}
	if be, ok := err.(*BusinessError); ok {
		return be.Category == CategoryValidation
	}
	return false
}
