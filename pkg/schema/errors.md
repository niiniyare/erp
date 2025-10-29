# Error Handling System

## Overview

The schema system uses a sophisticated error handling system with:
- Typed errors with codes and categories
- Error collections for multiple validation errors
- HTTP status code mapping
- API response structures
- Logger integration
- Panic recovery

## Error Interface

```go
type SchemaError interface {
    error
    Code() string                              // Error code (e.g., "field_not_found")
    Type() ErrorType                           // Error category
    Field() string                             // Field that caused error
    Details() map[string]any                   // Additional context
    WithField(field string) SchemaError        // Add field context
    WithDetail(key string, value any) SchemaError  // Add detail
}
```

## Error Types

```go
const (
    ErrorTypeValidation ErrorType = "validation" // 400 Bad Request
    ErrorTypeNotFound   ErrorType = "not_found"  // 404 Not Found
    ErrorTypeConflict   ErrorType = "conflict"   // 409 Conflict
    ErrorTypePermission ErrorType = "permission" // 403 Forbidden
    ErrorTypeDataSource ErrorType = "data_source" // 502 Bad Gateway
    ErrorTypeRender     ErrorType = "render"     // 500 Internal Error
    ErrorTypeWorkflow   ErrorType = "workflow"   // 422 Unprocessable
    ErrorTypeTenant     ErrorType = "tenant"     // 403 Forbidden
    ErrorTypeInternal   ErrorType = "internal"   // 500 Internal Error
)
```

## Pre-defined Errors

### Schema Errors
```go
ErrInvalidSchemaID    // Schema ID required
ErrInvalidSchemaType  // Schema type required
ErrInvalidSchemaTitle // Schema title required
ErrSchemaNotFound     // Schema not found
ErrSchemaDuplicate    // Schema already exists
```

### Field Errors
```go
ErrInvalidFieldName  // Field name required
ErrInvalidFieldType  // Field type required
ErrInvalidFieldLabel // Field label required
ErrFieldNotFound     // Field not found
ErrFieldDuplicate    // Duplicate field name
```

### Validation Errors
```go
ErrValidationFailed     // Generic validation failed
ErrRequiredField        // Required field missing
ErrInvalidValue         // Invalid field value
ErrAsyncValidationFail  // Async validation failed
```

### Permission Errors
```go
ErrPermissionDenied // Permission denied
ErrUnauthorized     // Not authorized
ErrForbidden        // Action forbidden
```

### Other Errors
```go
ErrDataSourceFailed   // Data source failed
ErrDataSourceTimeout  // Data source timeout
ErrRenderingFailed    // Rendering failed
ErrTemplateNotFound   // Template not found
ErrWorkflowFailed     // Workflow failed
ErrApprovalRequired   // Approval required
ErrTenantMismatch     // Tenant mismatch
ErrInternalError      // Internal error
```

## Creating Errors

### Basic Validation Error
```go
err := NewValidationError("invalid_email", "email format is invalid")
```

### With Field Context
```go
err := NewValidationError("required", "field is required").
    WithField("email")
// Output: [required] email: field is required
```

### With Additional Details
```go
err := NewValidationError("out_of_range", "value out of range").
    WithField("age").
    WithDetail("min", 18).
    WithDetail("max", 100).
    WithDetail("provided", 150)
```

### Not Found Error
```go
err := NewNotFoundError("user", "user not found")
// Creates: user_not_found error
```

### Permission Error
```go
err := NewPermissionError("access_denied", "insufficient permissions").
    WithDetail("required", []string{"admin", "editor"}).
    WithDetail("user_has", []string{"viewer"})
```

## Collecting Multiple Errors

### Using ErrorCollector
```go
collector := NewErrorCollector()

// Add errors
if field.Name == "" {
    collector.AddError(ErrInvalidFieldName)
}

if field.Type == "" {
    collector.AddFieldError(field.Name, ErrInvalidFieldType)
}

if age < 18 {
    collector.AddValidationError("age", "too_young", "must be 18 or older")
}

// Check and return
if collector.HasErrors() {
    return collector.Errors() // Returns ValidationErrorCollection
}
```

### ValidationErrorCollection
```go
// Check if has errors
if errorCollection.HasErrors() {
    // Get count
    count := errorCollection.Count()
    
    // Get all errors
    errors := errorCollection.Errors()
    
    // Group by field
    fieldErrors := errorCollection.ErrorsByField()
    // Returns: map[string][]string
    // Example: {
    //   "email": ["email is required", "invalid format"],
    //   "age": ["must be 18 or older"],
    //   "general": ["invalid request"]
    // }
}
```

## Checking Error Types

### Type Checking
```go
if IsValidationError(err) {
    // Handle validation error
}

if IsNotFoundError(err) {
    // Return 404
}

if IsPermissionError(err) {
    // Return 403
}

// Generic type check
if IsErrorType(err, ErrorTypeWorkflow) {
    // Handle workflow error
}
```

### Getting Error Info
```go
// Get error code
code := GetErrorCode(err)

// Get error details
details := GetErrorDetails(err)

// Get HTTP status code
statusCode := GetHTTPStatusCode(err)
```

## API Responses

### Single Error Response
```go
response := ToErrorResponse(err)
// Returns:
// {
//   "error": "[required] email: field is required",
//   "code": "required",
//   "type": "validation",
//   "field": "email",
//   "details": {}
// }
```

### Multiple Errors Response
```go
response := ToMultiErrorResponse(validationErrors)
// Returns:
// {
//   "errors": [...],
//   "count": 3,
//   "fieldErrors": {
//     "email": ["field is required", "invalid format"],
//     "age": ["must be 18 or older"]
//   },
//   "summary": "validation failed with 3 errors"
// }
```

## Error Wrapping

### Wrap Standard Errors
```go
err := someOperation()
if err != nil {
    return WrapError(err, "operation_failed", "failed to perform operation")
}
```

### Preserve Schema Errors
```go
err := validateField()
if err != nil {
    // If already SchemaError, returns as-is
    // If standard error, wraps it
    return WrapError(err, "field_validation", "field validation failed")
}
```

## Panic Recovery

### Recover from Panics
```go
func SafeValidate(schema *Schema) (err error) {
    defer func() {
        if recovered := RecoverSchemaError(); recovered != nil {
            err = recovered
        }
    }()
    
    return models.Validate()
}
```

## Error Handler with Logging

### Using Error Handler
```go
handler := NewErrorHandler(logger)

err := someOperation()
if err != nil {
    // Logs error and converts to SchemaError
    schemaErr := handler.HandleError(err)
    return schemaErr
}
```

### Handles Context Errors
```go
// Automatically detects and converts context errors
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := longRunningOperation(ctx)
schemaErr := handler.HandleError(err)
// If timeout: returns DataSourceError with "timeout" code
// If canceled: returns InternalError with "canceled" code
```

## Best Practices

### 1. Use Pre-defined Errors
```go
// Good
if field.Name == "" {
    return ErrInvalidFieldName
}

// Less good
if field.Name == "" {
    return NewValidationError("field_name", "field name is required")
}
```

### 2. Add Context with WithField()
```go
// Good
return ErrRequiredField.WithField(fieldName)

// Less informative
return ErrRequiredField
```

### 3. Collect Multiple Errors
```go
// Good - collect all errors before returning
collector := NewErrorCollector()
for _, field := range fields {
    if err := field.Validate(); err != nil {
        collector.AddFieldError(field.Name, err.(SchemaError))
    }
}
if collector.HasErrors() {
    return collector.Errors()
}

// Less good - return on first error
for _, field := range fields {
    if err := field.Validate(); err != nil {
        return err // User only sees first error
    }
}
```

### 4. Add Details for Debugging
```go
err := NewValidationError("out_of_range", "value out of acceptable range").
    WithField("price").
    WithDetail("min", minPrice).
    WithDetail("max", maxPrice).
    WithDetail("provided", actualPrice)
```

### 5. Use Appropriate Error Types
```go
// Not found - use NotFoundError
if user == nil {
    return NewNotFoundError("user", "user not found")
}

// Permission - use PermissionError
if !hasPermission {
    return NewPermissionError("access_denied", "insufficient permissions")
}

// Validation - use ValidationError
if !isValid {
    return NewValidationError("invalid_format", "invalid email format")
}
```

## HTTP Handler Example

```go
func HandleSchemaValidation(w http.ResponseWriter, r *http.Request) {
    var schema Schema
    if err := json.NewDecoder(r.Body).Decode(&schema); err != nil {
        respondError(w, WrapError(err, "invalid_json", "invalid JSON"))
        return
    }
    
    if err := models.Validate(); err != nil {
        respondError(w, err)
        return
    }
    
    // Success
    w.WriteHeader(http.StatusOK)
}

func respondError(w http.ResponseWriter, err error) {
    statusCode := GetHTTPStatusCode(err)
    w.WriteHeader(statusCode)
    w.Header().Set("Content-Type", "application/json")
    
    // Single error
    if schemaErr, ok := err.(SchemaError); ok {
        json.NewEncoder(w).Encode(ToErrorResponse(schemaErr))
        return
    }
    
    // Multiple errors
    if validationErrors, ok := err.(*ValidationErrorCollection); ok {
        json.NewEncoder(w).Encode(ToMultiErrorResponse(validationErrors))
        return
    }
    
    // Generic error
    json.NewEncoder(w).Encode(ToErrorResponse(err))
}
```

## Testing Error Handling

```go
func TestFieldValidation(t *testing.T) {
    field := Field{
        Type: FieldText,
        // Missing Name and Label
    }
    
    err := field.Validate(context.Background())
    
    // Check error type
    assert.True(t, IsValidationError(err))
    
    // Check error code
    assert.Equal(t, "field_name", GetErrorCode(err))
    
    // Check SchemaError interface
    schemaErr := err.(SchemaError)
    assert.Equal(t, ErrorTypeValidation, schemaErr.Type())
}
```

## Summary

The error system provides:
- ✅ Type-safe error handling
- ✅ Rich context and details
- ✅ Multiple error collection
- ✅ HTTP status mapping
- ✅ API response structures
- ✅ Logger integration
- ✅ Panic recovery
- ✅ Pre-defined common errors

Use it consistently across the schema system for robust error handling!
