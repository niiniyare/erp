package errors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ─── CONTEXT KEYS ─────────────────────────────────────────────

type contextKey string

const (
	TenantIDKey  contextKey = "tenant_id"
	UserIDKey    contextKey = "user_id"
	RequestIDKey contextKey = "request_id"
	OperationKey contextKey = "operation"
)

// ─── ERROR SEVERITY LEVELS ─────────────────────────────────────────────

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

// ─── ERROR CATEGORIES ─────────────────────────────────────────────

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

// ─── ENHANCED VALIDATION ERRORS ─────────────────────────────────────────────

// ValidationError represents a validation error with field details
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`  // e.g., "REQUIRED", "INVALID_FORMAT"
	Value   any    `json:"value,omitempty"` // The invalid value (sanitized)
}

func (e ValidationError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Field, e.Message, e.Code)
	}
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

// Add adds a validation error to the collection
func (ve *ValidationErrors) Add(field, message string) {
	*ve = append(*ve, ValidationError{Field: field, Message: message})
}

// AddWithCode adds a validation error with a code
func (ve *ValidationErrors) AddWithCode(field, message, code string) {
	*ve = append(*ve, ValidationError{Field: field, Message: message, Code: code})
}

// AddWithValue adds a validation error with the invalid value
func (ve *ValidationErrors) AddWithValue(field, message, code string, value any) {
	*ve = append(*ve, ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
		Value:   sanitizeValue(value),
	})
}

// HasErrors returns true if there are validation errors
func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

// GetByField returns validation errors for a specific field
func (ve ValidationErrors) GetByField(field string) []ValidationError {
	var fieldErrors []ValidationError
	for _, err := range ve {
		if err.Field == field {
			fieldErrors = append(fieldErrors, err)
		}
	}
	return fieldErrors
}

// ToMap converts validation errors to a map by field
func (ve ValidationErrors) ToMap() map[string][]string {
	result := make(map[string][]string)
	for _, err := range ve {
		result[err.Field] = append(result[err.Field], err.Message)
	}
	return result
}

// ─── ENHANCED REPOSITORY ERRORS ─────────────────────────────────────────────

// RepositoryError represents a failure in the repository layer
type RepositoryError struct {
	Code      string                 `json:"code"`    // e.g., "REJECT_FAILED"
	Message   string                 `json:"message"` // e.g., "Failed to reject access request"
	Err       error                  `json:"-"`       // Underlying cause (not serialized)
	Details   map[string]interface{} `json:"details,omitempty"`
	Operation string                 `json:"operation,omitempty"` // Database operation
	Table     string                 `json:"table,omitempty"`     // Affected table
	TenantID  string                 `json:"tenant_id,omitempty"`
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

// Unwrap allows errors.Is / errors.As support
func (e *RepositoryError) Unwrap() error {
	return e.Err
}

// WithDetail adds details to the error
func (e *RepositoryError) WithDetail(key string, value interface{}) *RepositoryError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithOperation sets the database operation
func (e *RepositoryError) WithOperation(operation string) *RepositoryError {
	e.Operation = operation
	return e
}

// WithTable sets the affected table
func (e *RepositoryError) WithTable(table string) *RepositoryError {
	e.Table = table
	return e
}

// WithTenant sets the tenant ID
func (e *RepositoryError) WithTenant(tenantID string) *RepositoryError {
	e.TenantID = tenantID
	return e
}

// NewRepositoryError creates a new RepositoryError
func NewRepositoryError(code, message string, err error) *RepositoryError {
	return &RepositoryError{
		Code:    code,
		Message: message,
		Err:     err,
		Details: make(map[string]interface{}),
	}
}

// NewRepositoryErrorWithContext creates a repository error with context information
func NewRepositoryErrorWithContext(ctx context.Context, code, message string, err error) *RepositoryError {
	repoErr := NewRepositoryError(code, message, err)

	if tenantID := getTenantIDFromContext(ctx); tenantID != "" {
		repoErr.WithTenant(tenantID)
	}

	if operation := getOperationFromContext(ctx); operation != "" {
		repoErr.WithOperation(operation)
	}

	return repoErr
}

// IsRepositoryErrorCode checks if the error is a RepositoryError with a specific code
func IsRepositoryErrorCode(err error, code string) bool {
	var repoErr *RepositoryError
	if errors.As(err, &repoErr) {
		return repoErr.Code == code
	}
	return false
}

// ─── BUSINESS LOGIC ERRORS ─────────────────────────────────────────────

// BusinessError represents domain-specific business logic errors
type BusinessError struct {
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Suggestions []string               `json:"suggestions,omitempty"` // User-friendly suggestions
	HTTPStatus  int                    `json:"-"`                     // HTTP status code mapping
	Severity    Severity               `json:"severity"`
	Category    Category               `json:"category"`
	TenantID    string                 `json:"tenant_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	Retryable   bool                   `json:"retryable"`
	Err         error                  `json:"-"`
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

func (e *BusinessError) Unwrap() error {
	return e.Err
}

// WithDetail adds details to the business error
func (e *BusinessError) WithDetail(key string, value interface{}) *BusinessError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}
func (e *BusinessError) WithCategory(category Category) *BusinessError {
	e.Category = category
	return e
}
func (e *BusinessError) WithSeverity(severity Severity) *BusinessError {
	e.Severity = severity
	return e
}

// WithSuggestion adds a user-friendly suggestion
func (e *BusinessError) WithSuggestion(suggestion string) *BusinessError {
	e.Suggestions = append(e.Suggestions, suggestion)
	return e
}

// WithHTTPStatus sets the HTTP status code
func (e *BusinessError) WithHTTPStatus(status int) *BusinessError {
	e.HTTPStatus = status
	return e
}

// IsRetryable returns whether this error is retryable
func (e *BusinessError) IsRetryable() bool {
	return e.Retryable
}

// NewBusinessError creates a new business error
func NewBusinessError(code, message string) *BusinessError {
	return &BusinessError{
		Code:       code,
		Message:    message,
		Details:    make(map[string]interface{}),
		HTTPStatus: http.StatusBadRequest,
		Severity:   SeverityError,
		Category:   CategoryBusiness,
		Retryable:  false,
	}
}

// NewBusinessErrorWithContext creates a business error with context
func NewBusinessErrorWithContext(ctx context.Context, code, message string) *BusinessError {
	err := NewBusinessError(code, message)

	if tenantID := getTenantIDFromContext(ctx); tenantID != "" {
		err.TenantID = tenantID
	}

	if userID := getUserIDFromContext(ctx); userID != "" {
		err.UserID = userID
	}

	return err
}

// ─── DOMAIN-SPECIFIC ERROR CONSTRUCTORS ─────────────────────────────────────────────

// Tenant-related errors
// func ErrTenantNotFound(tenantID string) *BusinessError {
// 	return NewBusinessError("TENANT_NOT_FOUND", "Tenant not found").
// 		WithDetail("tenant_id", tenantID).
// 		WithHTTPStatus(http.StatusNotFound).
// 		WithSuggestion("Verify the tenant ID is correct")
// }

func ErrTenantSuspended(tenantID string) *BusinessError {
	return NewBusinessError("TENANT_SUSPENDED", "Tenant account is suspended").
		WithDetail("tenant_id", tenantID).
		WithHTTPStatus(http.StatusForbidden).
		WithSuggestion("Contact support to reactivate your account")
}

func ErrTenantLimitExceeded(tenantID, limitType string, current, max int64) *BusinessError {
	return NewBusinessError("TENANT_LIMIT_EXCEEDED", fmt.Sprintf("Tenant %s limit exceeded", limitType)).
		WithDetail("tenant_id", tenantID).
		WithDetail("limit_type", limitType).
		WithDetail("current", current).
		WithDetail("maximum", max).
		WithHTTPStatus(http.StatusPaymentRequired).
		WithSuggestion("Upgrade your plan to increase limits")
}

//
// // Authentication/Authorization errors
// func ErrUnauthorized(reason string) *BusinessError {
// 	return NewBusinessError("UNAUTHORIZED", "Unauthorized access").
// 		WithDetail("reason", reason).
// 		WithHTTPStatus(http.StatusUnauthorized).
// 		WithCategory(CategorySecurity)
// }
//
// func ErrForbidden(resource string) *BusinessError {
// 	return NewBusinessError("FORBIDDEN", "Access forbidden").
// 		WithDetail("resource", resource).
// 		WithHTTPStatus(http.StatusForbidden).
// 		WithCategory(CategorySecurity).
// 		WithSuggestion("Contact your administrator for access")
// }

// Accounting-specific errors
func ErrInvalidAccountingPeriod(periodStart, periodEnd time.Time) *BusinessError {
	return NewBusinessError("INVALID_ACCOUNTING_PERIOD", "Invalid accounting period").
		WithDetail("period_start", periodStart).
		WithDetail("period_end", periodEnd).
		WithHTTPStatus(http.StatusBadRequest).
		WithSuggestion("Ensure the period start date is before the end date")
}

func ErrAccountingPeriodClosed(period string) *BusinessError {
	return NewBusinessError("ACCOUNTING_PERIOD_CLOSED", "Accounting period is closed").
		WithDetail("period", period).
		WithHTTPStatus(http.StatusConflict).
		WithSuggestion("Contact your accountant to reopen the period if necessary")
}

func ErrInsufficientBalance(accountID string, required, available float64) *BusinessError {
	return NewBusinessError("INSUFFICIENT_BALANCE", "Insufficient account balance").
		WithDetail("account_id", accountID).
		WithDetail("required", required).
		WithDetail("available", available).
		WithHTTPStatus(http.StatusConflict).
		WithSuggestion("Ensure sufficient funds are available before proceeding")
}

func ErrDuplicateTransaction(transactionID string) *BusinessError {
	return NewBusinessError("DUPLICATE_TRANSACTION", "Transaction already exists").
		WithDetail("transaction_id", transactionID).
		WithHTTPStatus(http.StatusConflict).
		WithSuggestion("Check if this transaction was already recorded")
}

// Integration errors
func ErrThirdPartyAPIFailure(service string, statusCode int) *BusinessError {
	return NewBusinessError("THIRD_PARTY_API_FAILURE", fmt.Sprintf("External service %s failed", service)).
		WithDetail("service", service).
		WithDetail("status_code", statusCode).
		WithHTTPStatus(http.StatusBadGateway).
		WithCategory(CategoryIntegration).
		WithSuggestion("Please try again later").
		WithDetail("retryable", true)
}

// ─── ERROR AGGREGATION ─────────────────────────────────────────────

// ErrorCollection aggregates multiple errors with context
type ErrorCollection struct {
	Errors    []error                `json:"errors"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Operation string                 `json:"operation,omitempty"`
	TenantID  string                 `json:"tenant_id,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Severity  Severity               `json:"severity"`
	Category  Category               `json:"category"`
}

func (ec *ErrorCollection) Error() string {
	if len(ec.Errors) == 0 {
		return "no errors"
	}

	if len(ec.Errors) == 1 {
		return ec.Errors[0].Error()
	}

	var messages []string
	for _, err := range ec.Errors {
		messages = append(messages, err.Error())
	}

	return fmt.Sprintf("multiple errors occurred: [%s]", strings.Join(messages, "; "))
}

// Add adds an error to the collection
func (ec *ErrorCollection) Add(err error) {
	if err != nil {
		ec.Errors = append(ec.Errors, err)
		ec.updateSeverity(err)
	}
}

// HasErrors returns true if there are errors in the collection
func (ec *ErrorCollection) HasErrors() bool {
	return len(ec.Errors) > 0
}

// Count returns the number of errors
func (ec *ErrorCollection) Count() int {
	return len(ec.Errors)
}

// GetByCategory returns errors of a specific category
func (ec *ErrorCollection) GetByCategory(category Category) []error {
	var categoryErrors []error
	for _, err := range ec.Errors {
		if be, ok := err.(*BusinessError); ok && be.Category == category {
			categoryErrors = append(categoryErrors, err)
		}
	}
	return categoryErrors
}

// GetBusinessErrors returns only business errors
func (ec *ErrorCollection) GetBusinessErrors() []*BusinessError {
	var businessErrors []*BusinessError
	for _, err := range ec.Errors {
		if be, ok := err.(*BusinessError); ok {
			businessErrors = append(businessErrors, be)
		}
	}
	return businessErrors
}

// GetValidationErrors returns only validation errors
func (ec *ErrorCollection) GetValidationErrors() ValidationErrors {
	var validationErrors ValidationErrors
	for _, err := range ec.Errors {
		if ve, ok := err.(ValidationError); ok {
			validationErrors = append(validationErrors, ve)
		} else if ves, ok := err.(ValidationErrors); ok {
			validationErrors = append(validationErrors, ves...)
		}
	}
	return validationErrors
}

// updateSeverity updates collection severity based on added error
func (ec *ErrorCollection) updateSeverity(err error) {
	if be, ok := err.(*BusinessError); ok {
		if be.Severity == SeverityCritical ||
			(be.Severity == SeverityError && ec.Severity != SeverityCritical) ||
			(be.Severity == SeverityWarning && ec.Severity == SeverityInfo) {
			ec.Severity = be.Severity
		}
	}
}

// NewErrorCollection creates a new error collection
func NewErrorCollection(operation string) *ErrorCollection {
	return &ErrorCollection{
		Errors:    make([]error, 0),
		Context:   make(map[string]interface{}),
		Operation: operation,
		Timestamp: time.Now(),
		Severity:  SeverityInfo,
		Category:  CategorySystem,
	}
}

// NewErrorCollectionWithContext creates an error collection with context
func NewErrorCollectionWithContext(ctx context.Context, operation string) *ErrorCollection {
	ec := NewErrorCollection(operation)

	if tenantID := getTenantIDFromContext(ctx); tenantID != "" {
		ec.TenantID = tenantID
	}

	if requestID := getRequestIDFromContext(ctx); requestID != "" {
		ec.RequestID = requestID
	}

	return ec
}

// ─── HTTP ERROR HANDLING ─────────────────────────────────────────────

// HTTPError represents an error that can be directly returned as HTTP response
type HTTPError struct {
	Status    int                    `json:"status"`
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Errors    []error                `json:"errors,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	RequestID string                 `json:"request_id,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
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
		Details:   make(map[string]interface{}),
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

// ─── CONTEXT HELPERS ─────────────────────────────────────────────

func getTenantIDFromContext(ctx context.Context) string {
	if tenantID, ok := ctx.Value(TenantIDKey).(uuid.UUID); ok {
		return tenantID.String()
	}
	if tenantID, ok := ctx.Value(TenantIDKey).(string); ok {
		return tenantID
	}
	return ""
}

func getUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(uuid.UUID); ok {
		return userID.String()
	}
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

func getRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

func getOperationFromContext(ctx context.Context) string {
	if operation, ok := ctx.Value(OperationKey).(string); ok {
		return operation
	}
	return ""
}

// ─── UTILITY FUNCTIONS ─────────────────────────────────────────────

// sanitizeValue removes sensitive information from values for logging
func sanitizeValue(value any) any {
	switch v := value.(type) {
	case string:
		// Sanitize potential sensitive strings
		if len(v) > 100 {
			return v[:100] + "..."
		}
		// Remove common sensitive patterns
		sensitivePatterns := []string{"password", "token", "secret", "key"}
		lowerV := strings.ToLower(v)
		for _, pattern := range sensitivePatterns {
			if strings.Contains(lowerV, pattern) {
				return "[REDACTED]"
			}
		}
		return v
	case map[string]interface{}:
		// Recursively sanitize maps
		sanitized := make(map[string]interface{})
		for k, val := range v {
			sanitized[k] = sanitizeValue(val)
		}
		return sanitized
	default:
		return v
	}
}

// IsTemporary checks if an error is temporary/retryable
func IsTemporary(err error) bool {
	if be, ok := err.(*BusinessError); ok {
		return be.Retryable
	}

	// Check for common temporary error patterns
	errStr := strings.ToLower(err.Error())
	temporaryPatterns := []string{
		"timeout", "connection reset", "connection refused",
		"temporary failure", "service unavailable", "rate limit",
	}

	for _, pattern := range temporaryPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// GetErrorCode extracts error code from various error types
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

// GetHTTPStatus extracts HTTP status from error
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

// ─── JSON MARSHALING SUPPORT ─────────────────────────────────────────────

// MarshalJSON implements json.Marshaler for safe error serialization
func (e *BusinessError) MarshalJSON() ([]byte, error) {
	type alias BusinessError
	return json.Marshal(&struct {
		*alias
		ErrorMessage string `json:"error_message"`
	}{
		alias:        (*alias)(e),
		ErrorMessage: e.Error(),
	})
}

func (e *RepositoryError) MarshalJSON() ([]byte, error) {
	type alias RepositoryError
	return json.Marshal(&struct {
		*alias
		ErrorMessage string `json:"error_message"`
	}{
		alias:        (*alias)(e),
		ErrorMessage: e.Error(),
	})
}
