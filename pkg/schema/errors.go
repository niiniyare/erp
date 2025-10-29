package schema

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/niiniyare/erp/internal/shared/logger"
)

// Core error types for schema validation and processing
var (
	// Schema validation errors
	ErrInvalidSchemaID    = NewValidationError("schema_id", "schema ID is required and must be valid")
	ErrInvalidSchemaType  = NewValidationError("schema_type", "schema type is required and must be valid")
	ErrInvalidSchemaTitle = NewValidationError("schema_title", "schema title is required")
	ErrSchemaNotFound     = NewNotFoundError("schema", "schema not found")
	ErrSchemaDuplicate    = NewConflictError("schema", "schema already exists")

	// Field validation errors
	ErrInvalidFieldName  = NewValidationError("field_name", "field name is required and must be valid")
	ErrInvalidFieldType  = NewValidationError("field_type", "field type is required and must be valid")
	ErrInvalidFieldLabel = NewValidationError("field_label", "field label is required")
	ErrFieldNotFound     = NewNotFoundError("field", "field not found")
	ErrFieldDuplicate    = NewConflictError("field", "field name already exists")

	// Validation specific errors
	ErrValidationFailed    = NewValidationError("validation", "validation failed")
	ErrRequiredField       = NewValidationError("required", "field is required")
	ErrInvalidValue        = NewValidationError("invalid_value", "field value is invalid")
	ErrAsyncValidationFail = NewValidationError("async_validation", "async validation failed")

	// Permission and security errors
	ErrPermissionDenied = NewPermissionError("access_denied", "permission denied")
	ErrUnauthorized     = NewPermissionError("unauthorized", "user not authorized")
	ErrForbidden        = NewPermissionError("forbidden", "action forbidden")

	// Data source errors
	ErrDataSourceFailed   = NewDataSourceError("data_source_failed", "data source request failed")
	ErrDataSourceTimeout  = NewDataSourceError("data_source_timeout", "data source request timeout")
	ErrDataSourceNotFound = NewNotFoundError("data_source", "data source not found")

	// Rendering errors
	ErrRenderingFailed  = NewRenderError("rendering_failed", "template rendering failed")
	ErrTemplateNotFound = NewNotFoundError("template", "template not found")

	// Workflow errors
	ErrWorkflowFailed   = NewWorkflowError("workflow_failed", "workflow execution failed")
	ErrApprovalRequired = NewWorkflowError("approval_required", "approval required")
	ErrWorkflowNotFound = NewNotFoundError("workflow", "workflow not found")

	// Multi-tenancy errors
	ErrTenantMismatch = NewTenantError("tenant_mismatch", "tenant mismatch")
	ErrTenantNotFound = NewNotFoundError("tenant", "tenant not found")

	// General errors
	ErrInternalError  = NewInternalError("internal_error", "internal server error")
	ErrInvalidRequest = NewValidationError("invalid_request", "invalid request")
)

// SchemaError is the base interface for all schema-related errors
type SchemaError interface {
	error
	Code() string
	Type() ErrorType
	Field() string
	Details() map[string]any
	WithField(field string) SchemaError
	WithDetail(key string, value any) SchemaError
}

// ErrorType categorizes errors for better handling
type ErrorType string

const (
	ErrorTypeValidation ErrorType = "validation"
	ErrorTypeNotFound   ErrorType = "not_found"
	ErrorTypeConflict   ErrorType = "conflict"
	ErrorTypePermission ErrorType = "permission"
	ErrorTypeDataSource ErrorType = "data_source"
	ErrorTypeRender     ErrorType = "render"
	ErrorTypeWorkflow   ErrorType = "workflow"
	ErrorTypeTenant     ErrorType = "tenant"
	ErrorTypeInternal   ErrorType = "internal"
)

// BaseError implements the SchemaError interface
type BaseError struct {
	code    string
	message string
	errType ErrorType
	field   string
	details map[string]any
}

func (e *BaseError) Error() string {
	if e.field != "" {
		return fmt.Sprintf("[%s] %s: %s", e.code, e.field, e.message)
	}
	return fmt.Sprintf("[%s] %s", e.code, e.message)
}

func (e *BaseError) Code() string {
	return e.code
}

func (e *BaseError) Type() ErrorType {
	return e.errType
}

func (e *BaseError) Field() string {
	return e.field
}

func (e *BaseError) Details() map[string]any {
	if e.details == nil {
		e.details = make(map[string]any)
	}
	return e.details
}

func (e *BaseError) WithField(field string) SchemaError {
	newErr := *e
	newErr.field = field
	return &newErr
}

func (e *BaseError) WithDetail(key string, value any) SchemaError {
	newErr := *e
	if newErr.details == nil {
		newErr.details = make(map[string]any)
	}
	newErr.details[key] = value
	return &newErr
}

// Specific error constructors
func NewValidationError(code, message string) SchemaError {
	return &BaseError{
		code:    code,
		message: message,
		errType: ErrorTypeValidation,
		details: make(map[string]any),
	}
}

func NewNotFoundError(resource, message string) SchemaError {
	return &BaseError{
		code:    fmt.Sprintf("%s_not_found", resource),
		message: message,
		errType: ErrorTypeNotFound,
		details: map[string]any{
			"resource": resource,
		},
	}
}

func NewConflictError(resource, message string) SchemaError {
	return &BaseError{
		code:    fmt.Sprintf("%s_conflict", resource),
		message: message,
		errType: ErrorTypeConflict,
		details: map[string]any{
			"resource": resource,
		},
	}
}

func NewPermissionError(code, message string) SchemaError {
	return &BaseError{
		code:    code,
		message: message,
		errType: ErrorTypePermission,
		details: make(map[string]any),
	}
}

func NewDataSourceError(code, message string) SchemaError {
	return &BaseError{
		code:    code,
		message: message,
		errType: ErrorTypeDataSource,
		details: make(map[string]any),
	}
}

func NewRenderError(code, message string) SchemaError {
	return &BaseError{
		code:    code,
		message: message,
		errType: ErrorTypeRender,
		details: make(map[string]any),
	}
}

func NewWorkflowError(code, message string) SchemaError {
	return &BaseError{
		code:    code,
		message: message,
		errType: ErrorTypeWorkflow,
		details: make(map[string]any),
	}
}

func NewTenantError(code, message string) SchemaError {
	return &BaseError{
		code:    code,
		message: message,
		errType: ErrorTypeTenant,
		details: make(map[string]any),
	}
}

func NewInternalError(code, message string) SchemaError {
	return &BaseError{
		code:    code,
		message: message,
		errType: ErrorTypeInternal,
		details: make(map[string]any),
	}
}

// ValidationErrorCollection holds multiple validation errors
type ValidationErrorCollection struct {
	errors []SchemaError
}

func NewValidationErrorCollection() *ValidationErrorCollection {
	return &ValidationErrorCollection{
		errors: make([]SchemaError, 0),
	}
}

func (vec *ValidationErrorCollection) Add(err SchemaError) {
	vec.errors = append(vec.errors, err)
}

func (vec *ValidationErrorCollection) AddWithField(field string, err SchemaError) {
	vec.errors = append(vec.errors, err.WithField(field))
}

func (vec *ValidationErrorCollection) HasErrors() bool {
	return len(vec.errors) > 0
}

func (vec *ValidationErrorCollection) Count() int {
	return len(vec.errors)
}

func (vec *ValidationErrorCollection) Errors() []SchemaError {
	return vec.errors
}

func (vec *ValidationErrorCollection) Error() string {
	if len(vec.errors) == 0 {
		return "no validation errors"
	}

	var messages []string
	for _, err := range vec.errors {
		messages = append(messages, err.Error())
	}

	return fmt.Sprintf("validation failed with %d errors: %s",
		len(vec.errors),
		strings.Join(messages, "; "))
}

func (vec *ValidationErrorCollection) ErrorsByField() map[string][]string {
	fieldErrors := make(map[string][]string)

	for _, err := range vec.errors {
		field := err.Field()
		if field == "" {
			field = "general"
		}

		if _, exists := fieldErrors[field]; !exists {
			fieldErrors[field] = make([]string, 0)
		}

		fieldErrors[field] = append(fieldErrors[field], err.Error())
	}

	return fieldErrors
}

// ErrorCollector helps collect errors during validation
type ErrorCollector struct {
	errors *ValidationErrorCollection
}

func NewErrorCollector() *ErrorCollector {
	return &ErrorCollector{
		errors: NewValidationErrorCollection(),
	}
}

func (ec *ErrorCollector) AddError(err SchemaError) {
	ec.errors.Add(err)
}

func (ec *ErrorCollector) AddFieldError(field string, err SchemaError) {
	ec.errors.AddWithField(field, err)
}

func (ec *ErrorCollector) AddValidationError(field, code, message string) {
	ec.errors.AddWithField(field, NewValidationError(code, message))
}

func (ec *ErrorCollector) HasErrors() bool {
	return ec.errors.HasErrors()
}

func (ec *ErrorCollector) Errors() *ValidationErrorCollection {
	return ec.errors
}

// Helper functions for common error patterns
func WrapError(err error, code, message string) SchemaError {
	if schemaErr, ok := err.(SchemaError); ok {
		return schemaErr
	}

	return &BaseError{
		code:    code,
		message: fmt.Sprintf("%s: %v", message, err),
		errType: ErrorTypeInternal,
		details: map[string]any{
			"original_error": err.Error(),
		},
	}
}

func IsErrorType(err error, errType ErrorType) bool {
	if schemaErr, ok := err.(SchemaError); ok {
		return schemaErr.Type() == errType
	}
	return false
}

func IsValidationError(err error) bool {
	return IsErrorType(err, ErrorTypeValidation)
}

func IsNotFoundError(err error) bool {
	return IsErrorType(err, ErrorTypeNotFound)
}

func IsPermissionError(err error) bool {
	return IsErrorType(err, ErrorTypePermission)
}

func GetErrorCode(err error) string {
	if schemaErr, ok := err.(SchemaError); ok {
		return schemaErr.Code()
	}
	return "unknown_error"
}

func GetErrorDetails(err error) map[string]any {
	if schemaErr, ok := err.(SchemaError); ok {
		return schemaErr.Details()
	}
	return map[string]any{
		"error": err.Error(),
	}
}

// HTTP status code mapping for errors
func GetHTTPStatusCode(err error) int {
	if schemaErr, ok := err.(SchemaError); ok {
		switch schemaErr.Type() {
		case ErrorTypeValidation:
			return 400 // Bad Request
		case ErrorTypeNotFound:
			return 404 // Not Found
		case ErrorTypeConflict:
			return 409 // Conflict
		case ErrorTypePermission:
			return 403 // Forbidden
		case ErrorTypeDataSource:
			return 502 // Bad Gateway
		case ErrorTypeRender:
			return 500 // Internal Server Error
		case ErrorTypeWorkflow:
			return 422 // Unprocessable Entity
		case ErrorTypeTenant:
			return 403 // Forbidden
		case ErrorTypeInternal:
			return 500 // Internal Server Error
		}
	}
	return 500 // Default to Internal Server Error
}

// Error response structure for API responses
type ErrorResponse struct {
	Error   string         `json:"error"`
	Code    string         `json:"code"`
	Type    string         `json:"type"`
	Field   string         `json:"field,omitempty"`
	Details map[string]any `json:"details,omitempty"`
}

func ToErrorResponse(err error) *ErrorResponse {
	if schemaErr, ok := err.(SchemaError); ok {
		return &ErrorResponse{
			Error:   schemaErr.Error(),
			Code:    schemaErr.Code(),
			Type:    string(schemaErr.Type()),
			Field:   schemaErr.Field(),
			Details: schemaErr.Details(),
		}
	}

	return &ErrorResponse{
		Error: err.Error(),
		Code:  "unknown_error",
		Type:  string(ErrorTypeInternal),
	}
}

// Multiple error response for validation collections
type MultiErrorResponse struct {
	Errors      []ErrorResponse     `json:"errors"`
	Count       int                 `json:"count"`
	FieldErrors map[string][]string `json:"fieldErrors"`
	Summary     string              `json:"summary"`
}

func ToMultiErrorResponse(err *ValidationErrorCollection) *MultiErrorResponse {
	if err == nil || !err.HasErrors() {
		return &MultiErrorResponse{
			Errors:      []ErrorResponse{},
			Count:       0,
			FieldErrors: make(map[string][]string),
			Summary:     "no errors",
		}
	}

	var errorResponses []ErrorResponse
	for _, schemaErr := range err.Errors() {
		errorResponses = append(errorResponses, *ToErrorResponse(schemaErr))
	}

	return &MultiErrorResponse{
		Errors:      errorResponses,
		Count:       err.Count(),
		FieldErrors: err.ErrorsByField(),
		Summary:     fmt.Sprintf("validation failed with %d errors", err.Count()),
	}
}

// Panic recovery for schema operations
func RecoverSchemaError() SchemaError {
	if r := recover(); r != nil {
		if err, ok := r.(error); ok {
			return WrapError(err, "panic_recovered", "schema operation panicked")
		}
		return NewInternalError("panic_recovered", fmt.Sprintf("schema operation panicked: %v", r))
	}
	return nil
}

// Error middleware for consistent error handling
type ErrorHandler struct {
	logger logger.Logger
}

func NewErrorHandler(logger logger.Logger) *ErrorHandler {
	return &ErrorHandler{logger: logger}
}

func (eh *ErrorHandler) HandleError(err error) SchemaError {
	if err == nil {
		return nil
	}

	// Log the error
	if eh.logger != nil {
		eh.logger.Error("schema error occurred", map[string]any{
			"error": err.Error(),
			"type":  fmt.Sprintf("%T", err),
		})
	}

	// Convert to SchemaError if needed
	if schemaErr, ok := err.(SchemaError); ok {
		return schemaErr
	}

	// Handle specific Go error types
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return NewDataSourceError("timeout", "operation timed out")
	case errors.Is(err, context.Canceled):
		return NewInternalError("canceled", "operation was canceled")
	default:
		return WrapError(err, "unknown_error", "unexpected error occurred")
	}
}

// ErrRendererNotFound represents a renderer not found error
type ErrRendererNotFound struct {
	Type string
}

func (e ErrRendererNotFound) Error() string {
	return fmt.Sprintf("renderer not found for type: %s", e.Type)
}
