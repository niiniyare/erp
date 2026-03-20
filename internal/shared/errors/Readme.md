# Shared Errors Package

A foundational error handling system for the Awo ERP platform. This package provides core error types, utilities, and patterns that other packages can use to create their own domain-specific errors.

## Package Overview

This is a **shared foundation package** - it provides the building blocks for error handling across the entire ERP system. Individual modules should create their own `errors.go` files that use this package's functionality for domain-specific errors.

### Architecture Philosophy

```
┌─────────────────────────────────────────────────┐
│ internal/shared/errors (Foundation Package)     │
│ - Core error types (BusinessError, etc.)       │
│ - Utility functions (IsTemporary, etc.)        │ 
│ - Common error codes and patterns               │
│ - HTTP integration helpers                      │
└─────────────────────────────────────────────────┘
                         ↑ uses
┌─────────────────────────────────────────────────┐
│ Module-Specific errors.go Files                 │
│ internal/core/finance/errors.go                 │
│ internal/core/tenant/errors.go                  │
│ internal/core/iam/errors.go                     │
│ internal/api/handlers/errors.go                 │
└─────────────────────────────────────────────────┘
```

## File Organization

```
internal/shared/errors/
├── types.go         # Core error types and interfaces
├── business.go      # Common predefined business errors
├── codes.go         # Error code constants
├── context.go       # Context helper functions
├── http.go          # HTTP error handling
├── tenant.go        # Tenant-specific errors
├── utils.go         # Utility and checking functions
├── errors.go        # Main export file
└── README.md        # This documentation
```

## Table of Contents

- [Core Error Types](#core-error-types)
- [Creating Module-Specific Errors](#creating-module-specific-errors)
- [Common Patterns](#common-patterns)
- [HTTP Integration](#http-integration)
- [Testing Patterns](#testing-patterns)
- [Migration Guide](#migration-guide)

## Core Error Types

### BusinessError

The primary error type for business logic errors with rich context:

```go
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
}
```

### RepositoryError

For database/storage layer errors:

```go
type RepositoryError struct {
    Code      string         `json:"code"`
    Message   string         `json:"message"`
    Operation string         `json:"operation,omitempty"`
    Table     string         `json:"table,omitempty"`
    TenantID  string         `json:"tenant_id,omitempty"`
}
```

### ValidationErrors

For input validation errors:

```go
type ValidationErrors []ValidationError

type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
    Code    string `json:"code,omitempty"`
    Value   any    `json:"value,omitempty"`
}
```

## Creating Module-Specific Errors

Each module should create its own `errors.go` file that imports this package and defines domain-specific errors:

### Example: Finance Module Errors

```go
// internal/core/finance/errors.go
package finance

import (
    "net/http"
    "awo/internal/shared/errors"
)

// Finance-specific error codes
const (
    CodeAccountNotFound     = "ACCOUNT_NOT_FOUND"
    CodeInvalidTransaction  = "INVALID_TRANSACTION"
    CodeInsufficientFunds   = "INSUFFICIENT_FUNDS"
)

// Predefined finance errors
var (
    ErrAccountNotFound = errors.NewBusinessError(CodeAccountNotFound, "Account not found").
        WithHTTPStatus(http.StatusNotFound).
        WithCategory(errors.CategoryBusiness).
        WithSuggestion("Verify the account ID").
        WithSuggestion("Check if the account exists in your chart of accounts")
    
    ErrInsufficientFunds = errors.NewBusinessError(CodeInsufficientFunds, "Insufficient funds").
        WithHTTPStatus(http.StatusConflict).
        WithCategory(errors.CategoryBusiness).
        WithSuggestion("Check account balance before transaction")
)

// Contextual constructors
func NewAccountNotFoundError(accountID string) *errors.BusinessError {
    return errors.NewBusinessError(CodeAccountNotFound, "Account not found").
        WithHTTPStatus(http.StatusNotFound).
        WithCategory(errors.CategoryBusiness).
        WithDetail("account_id", accountID).
        WithSuggestion("Verify the account ID is correct")
}

func NewInsufficientFundsError(accountID string, required, available float64) *errors.BusinessError {
    return errors.NewBusinessError(CodeInsufficientFunds, "Insufficient funds").
        WithHTTPStatus(http.StatusConflict).
        WithCategory(errors.CategoryBusiness).
        WithDetail("account_id", accountID).
        WithDetail("required_amount", required).
        WithDetail("available_amount", available).
        WithSuggestion("Check account balance before proceeding")
}

// Checking functions
func IsAccountNotFound(err error) bool {
    return errors.IsBusinessErrorCode(err, CodeAccountNotFound)
}

func IsInsufficientFunds(err error) bool {
    return errors.IsBusinessErrorCode(err, CodeInsufficientFunds)
}
```

### Example: IAM Module Errors

```go
// internal/core/iam/errors.go
package iam

import (
    "net/http"
    "awo/internal/shared/errors"
)

// IAM-specific error codes
const (
    CodeSessionExpired     = "SESSION_EXPIRED"
    CodeInvalidToken       = "INVALID_TOKEN"
    CodePermissionDenied   = "PERMISSION_DENIED"
)

// Predefined IAM errors
var (
    ErrSessionExpired = errors.NewBusinessError(CodeSessionExpired, "Session has expired").
        WithHTTPStatus(http.StatusUnauthorized).
        WithCategory(errors.CategorySecurity).
        WithSuggestion("Please log in again")
    
    ErrPermissionDenied = errors.NewBusinessError(CodePermissionDenied, "Permission denied").
        WithHTTPStatus(http.StatusForbidden).
        WithCategory(errors.CategorySecurity).
        WithSuggestion("Contact your administrator for access")
)

// Contextual constructors
func NewInvalidTokenError(tokenType string) *errors.BusinessError {
    return errors.NewBusinessError(CodeInvalidToken, "Invalid token").
        WithHTTPStatus(http.StatusUnauthorized).
        WithCategory(errors.CategorySecurity).
        WithDetail("token_type", tokenType).
        WithSuggestion("Obtain a new token and try again")
}
```

## Common Patterns

### Basic Error Creation

```go
// Simple business error
err := errors.NewBusinessError("CUSTOM_ERROR", "Something went wrong").
    WithHTTPStatus(http.StatusBadRequest).
    WithCategory(errors.CategoryBusiness)

// With context and details
err := errors.NewBusinessErrorWithContext(ctx, "USER_ACTION_FAILED", "User action failed").
    WithDetail("action", "update_profile").
    WithDetail("user_id", userID).
    WithSuggestion("Try again later")

// Repository error
err := errors.NewRepositoryError("QUERY_FAILED", "Database query failed", originalErr).
    WithOperation("SELECT").
    WithTable("users")
```

### Error Checking

```go
// Check specific error codes
if errors.IsBusinessErrorCode(err, "ACCOUNT_NOT_FOUND") {
    // Handle account not found
}

// Check error categories
if be, ok := err.(*errors.BusinessError); ok {
    switch be.Category {
    case errors.CategorySecurity:
        // Handle security errors
    case errors.CategoryValidation:
        // Handle validation errors
    }
}

// Use utility functions
if errors.IsTemporary(err) {
    // Retry the operation
}

if errors.IsConflict(err) {
    // Handle conflict errors
}
```

### Validation Errors

```go
var validationErrs errors.ValidationErrors

// Add validation errors
validationErrs.Add("email", "Invalid email format")
validationErrs.AddWithCode("password", "Password too short", "TOO_SHORT")
validationErrs.AddWithValue("age", "Must be 18 or older", "INVALID_RANGE", 15)

// Check and return
if validationErrs.HasErrors() {
    return validationErrs
}
```

## HTTP Integration

### Automatic HTTP Response

```go
func HandleError(w http.ResponseWriter, err error) {
    httpErr := errors.ToHTTPError(err)
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(httpErr.Status)
    json.NewEncoder(w).Encode(httpErr)
}
```

### Middleware Integration

```go
func ErrorMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if recovered := recover(); recovered != nil {
                err := errors.NewBusinessError("INTERNAL_PANIC", "Unexpected error").
                    WithHTTPStatus(http.StatusInternalServerError).
                    WithCategory(errors.CategorySystem).
                    WithSeverity(errors.SeverityCritical)
                
                HandleError(w, err)
            }
        }()
        
        next.ServeHTTP(w, r)
    })
}
```

## Testing Patterns

### Testing Error Types

```go
func TestFinanceErrors(t *testing.T) {
    // Test error creation
    err := NewAccountNotFoundError("acc_123")
    assert.True(t, IsAccountNotFound(err))
    
    // Test error details
    be, ok := err.(*errors.BusinessError)
    require.True(t, ok)
    assert.Equal(t, "ACCOUNT_NOT_FOUND", be.Code)
    assert.Equal(t, http.StatusNotFound, be.HTTPStatus)
    assert.Equal(t, "acc_123", be.Details["account_id"])
    
    // Test HTTP conversion
    httpErr := errors.ToHTTPError(err)
    assert.Equal(t, http.StatusNotFound, httpErr.Status)
    assert.Equal(t, "ACCOUNT_NOT_FOUND", httpErr.Code)
}
```

### Testing Error Checking

```go
func TestErrorChecking(t *testing.T) {
    err := NewInsufficientFundsError("acc_123", 100.0, 50.0)
    
    // Test specific checks
    assert.True(t, IsInsufficientFunds(err))
    assert.False(t, IsAccountNotFound(err))
    
    // Test utility checks
    assert.True(t, errors.IsConflict(err))
    assert.False(t, errors.IsTemporary(err))
}
```

## Migration Guide

### From Simple Errors

```go
// Before: Simple errors
var ErrUserNotFound = errors.New("user not found")

// After: Rich business errors (in module-specific errors.go)
var ErrUserNotFound = errors.NewBusinessError("USER_NOT_FOUND", "User not found").
    WithHTTPStatus(http.StatusNotFound).
    WithCategory(errors.CategorySecurity)
```

### Gradual Migration Strategy

1. **Phase 1**: Continue using existing errors (100% backward compatible)
2. **Phase 2**: Create module-specific `errors.go` files using this package
3. **Phase 3**: Gradually replace simple errors with rich business errors
4. **Phase 4**: Add contextual constructors and enhanced checking

### Best Practices

#### ✅ Do

- Create module-specific `errors.go` files
- Use contextual error constructors
- Include helpful suggestions for users
- Add relevant details for debugging
- Use appropriate HTTP status codes
- Test error handling thoroughly

#### ❌ Don't

- Put all errors in this shared package
- Create errors without context
- Expose internal implementation details in error messages
- Use generic error messages
- Ignore error categories and severity

## Available Utilities

### Error Checking Functions

```go
errors.IsBusinessErrorCode(err, "CODE")     // Check specific error code
errors.IsTemporary(err)                     // Check if retryable
errors.IsConflict(err)                      // Check for conflict errors
errors.IsValidationError(err)               // Check for validation errors
errors.IsNotFoundError(err)                 // Check for any "not found" error
errors.IsAuthenticationError(err)           // Check for auth errors
errors.IsSecurityError(err)                 // Check for security category
errors.IsTenantError(err)                   // Check for tenant category
```

### Error Utilities

```go
errors.GetErrorCode(err)                    // Extract error code
errors.GetHTTPStatus(err)                   // Extract HTTP status
errors.GetErrorSummary(err)                 // Get debug summary
errors.GetErrorChain(err)                   // Get wrapped error chain
errors.FormatErrorCodes(errs)               // Format multiple error codes
errors.FilterRetryableErrors(errs)          // Filter retryable errors
```

### Error Constants

```go
// Categories
errors.CategoryValidation, errors.CategorySecurity, 
errors.CategoryBusiness, errors.CategoryTenant, etc.

// Severity Levels  
errors.SeverityInfo, errors.SeverityWarning,
errors.SeverityError, errors.SeverityCritical

// Common Error Codes
errors.CodeUserNotFound, errors.CodeInvalidInput,
errors.CodeUnauthorized, errors.CodeForbidden, etc.
```

This shared errors package provides a solid foundation for building consistent, rich error handling across the entire Awo ERP system while allowing each module to define its own domain-specific errors.

## Backward Compatibility

### Your Existing Code Still Works

```go
// ✅ All existing error checking continues to work
func ExistingUserService(userID string) error {
    user, err := userRepo.GetUser(userID)
    if err != nil {
        // Your existing error checking still works
        if errors.Is(err, ErrUserNotFound) {
            return fmt.Errorf("user %s not found", userID)
        }
        return err
    }
    return nil
}

// ✅ Existing error returns still work
func ExistingAuthService(email, password string) error {
    if !validateCredentials(email, password) {
        return ErrInvalidCredentials // Still works exactly the same
    }
    return nil
}
```

### Simple Error Comparisons Continue Working

```go
// ✅ All these patterns continue to work unchanged
func ErrorHandlingPatterns(err error) {
    // Direct comparison
    if err == ErrUserNotFound {
        // Handle user not found
    }
    
    // errors.Is checking
    if errors.Is(err, ErrInvalidCredentials) {
        // Handle invalid credentials
    }
    
    // Switch statements
    switch err {
    case ErrTenantNotFound:
        // Handle tenant not found
    case ErrEntityNotFound:
        // Handle entity not found
    }
}
```

##  Error Features

### Rich Error Information

```go
func UserService(ctx context.Context, userID string) error {
    user, err := userRepo.GetUser(ctx, userID)
    if err != nil {
        if errors.Is(err, ErrUserNotFound) {
            // 🆕  error now includes HTTP status, suggestions, etc.
            return err // Returns rich BusinessError with HTTP 404, suggestions
        }
        return err
    }
    return nil
}

func HandleError(w http.ResponseWriter, err error) {
    if errors.Is(err, ErrUserNotFound) {
        // 🆕 Extract rich information from error
        if be, ok := err.(*BusinessError); ok {
            log.Printf("Error code: %s", be.Code)                    // "USER_NOT_FOUND"
            log.Printf("HTTP status: %d", be.HTTPStatus)             // 404
            log.Printf("Category: %s", be.Category)                  // "security"
            log.Printf("Suggestions: %v", be.Suggestions)            // ["Verify the user ID is correct", ...]
            
            // Automatic HTTP response
            httpErr := ToHTTPError(err)
            w.WriteHeader(httpErr.Status)
            json.NewEncoder(w).Encode(httpErr)
            return
        }
    }
}
```

## Usage by Domain

### Tenant Errors

```go
func TenantService(ctx context.Context, tenantSlug string) error {
    tenant, err := tenantRepo.GetBySlug(ctx, tenantSlug)
    if err != nil {
        if errors.Is(err, ErrTenantNotFound) {
            // 🆕 Rich tenant error with suggestions
            return err // HTTP 404 with helpful suggestions
        }
        return err
    }
    
    // Check tenant status
    if tenant.Status == "suspended" {
        // 🆕 Use contextual error constructor
        return ErrTenantSuspended(tenant.ID).
            WithDetail("suspended_at", tenant.SuspendedAt).
            WithDetail("reason", tenant.SuspensionReason)
    }
    
    // Check tenant limits
    if tenant.UserCount >= tenant.MaxUsers {
        return ErrTenantLimitExceeded(tenant.ID, "users", tenant.UserCount, tenant.MaxUsers)
    }
    
    return nil
}

func CreateTenantHandler(w http.ResponseWriter, r *http.Request) {
    var req CreateTenantRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    tenant, err := tenantService.CreateTenant(r.Context(), req)
    if err != nil {
        switch {
        case errors.Is(err, ErrTenantExists):
            // 🆕 Automatic HTTP 409 Conflict with suggestions
            httpErr := ToHTTPError(err)
            w.WriteHeader(httpErr.Status)
            json.NewEncoder(w).Encode(httpErr)
            return
        case errors.Is(err, ErrSubdomainAlreadyExists):
            // 🆕 Automatic HTTP 409 with subdomain-specific suggestions
            httpErr := ToHTTPError(err)
            w.WriteHeader(httpErr.Status)
            json.NewEncoder(w).Encode(httpErr)
            return
        default:
            http.Error(w, "Internal error", http.StatusInternalServerError)
            return
        }
    }
    
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(tenant)
}
```

### User Errors

```go
func UserAuthenticationService(ctx context.Context, email, password string) (*User, error) {
    user, err := userRepo.GetByEmail(ctx, email)
    if err != nil {
        if errors.Is(err, ErrUserNotFound) {
            // 🆕 Use contextual constructor for better error info
            return nil, NewUserNotFoundByEmailError(email)
        }
        return nil, err
    }
    
    // Check account status
    if user.Status == "locked" {
        return nil, ErrAccountLocked.
            WithDetail("locked_at", user.LockedAt).
            WithDetail("lock_reason", user.LockReason).
            WithSuggestion("Contact support with your user ID: " + user.ID)
    }
    
    // Validate credentials
    attemptCount := getLoginAttemptCount(ctx, email)
    if !validatePassword(password, user.PasswordHash) {
        // 🆕  error with attempt tracking
        return nil, NewInvalidCredentialsError(email, attemptCount+1)
    }
    
    return user, nil
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    user, err := userService.CreateUser(r.Context(), req)
    if err != nil {
        // 🆕 Multiple specific error types with automatic HTTP mapping
        switch {
        case errors.Is(err, ErrEmailAlreadyExists):
            // HTTP 409 with email-specific suggestions
        case errors.Is(err, ErrUsernameAlreadyExists):
            // HTTP 409 with username-specific suggestions
        case errors.Is(err, ErrInvalidUserType):
            // HTTP 400 with valid user types
        case errors.Is(err, ErrInvalidAccountStatus):
            // HTTP 400 with valid statuses
        }
        
        httpErr := ToHTTPError(err)
        w.WriteHeader(httpErr.Status)
        json.NewEncoder(w).Encode(httpErr)
        return
    }
    
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}
```

### Entity Errors

```go
func EntityService(ctx context.Context, entityID string) error {
    entity, err := entityRepo.GetByID(ctx, entityID)
    if err != nil {
        if errors.Is(err, ErrEntityNotFound) {
            // 🆕 Use contextual constructor
            return NewEntityNotFoundError(entityID)
        }
        return err
    }
    
    return nil
}

func CreateEntityHandler(w http.ResponseWriter, r *http.Request) {
    var req CreateEntityRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    entity, err := entityService.CreateEntity(r.Context(), req)
    if err != nil {
        switch {
        case errors.Is(err, ErrEntityNameExists):
            // 🆕 HTTP 409 with name conflict details
        case errors.Is(err, ErrEntityCodeExists):
            // 🆕 HTTP 409 with code conflict details
        case errors.Is(err, ErrInvalidParentEntity):
            // 🆕 HTTP 400 with parent validation details
        case errors.Is(err, ErrCircularReference):
            // 🆕 HTTP 400 with hierarchy explanation
        case errors.Is(err, ErrInvalidEntityType):
            // 🆕 HTTP 400 with valid entity types
        }
        
        httpErr := ToHTTPError(err)
        w.WriteHeader(httpErr.Status)
        json.NewEncoder(w).Encode(httpErr)
        return
    }
    
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(entity)
}

func DeleteEntityHandler(w http.ResponseWriter, r *http.Request) {
    entityID := mux.Vars(r)["id"]
    
    err := entityService.DeleteEntity(r.Context(), entityID)
    if err != nil {
        switch {
        case errors.Is(err, ErrEntityNotFound):
            // 🆕 HTTP 404 with entity-specific suggestions
        case errors.Is(err, ErrEntityHasChildren):
            // 🆕 HTTP 409 with child entity management suggestions
        }
        
        httpErr := ToHTTPError(err)
        w.WriteHeader(httpErr.Status)
        json.NewEncoder(w).Encode(httpErr)
        return
    }
    
    w.WriteHeader(http.StatusNoContent)
}
```

## Error Checking Patterns

###  Error Checking

```go
// ✅ Traditional checking still works
func TraditionalChecking(err error) {
    if errors.Is(err, ErrUserNotFound) {
        // Handle user not found
    }
}

// 🆕  checking with error details
func Checking(err error) {
    // Check error type and extract details
    if IsUserNotFound(err) {
        if be, ok := err.(*BusinessError); ok {
            userID := be.Details["user_id"]
            log.Printf("User %v not found", userID)
        }
    }
    
    // Check error categories
    if be, ok := err.(*BusinessError); ok {
        switch be.Category {
        case CategorySecurity:
            // Handle security-related errors
            logSecurityEvent(be)
        case CategoryValidation:
            // Handle validation errors
            return ValidationResponse(be)
        case CategoryTenant:
            // Handle tenant-specific errors
            notifyTenantAdmin(be)
        }
    }
    
    // Check if error is retryable
    if IsTemporaryError(err) {
        // Implement retry logic
        return retryOperation()
    }
}

// 🆕 Batch error checking
func BatchErrorChecking(err error) {
    // Check for conflict errors
    if IsConflict(err) {
        // Handle any type of conflict (user exists, entity exists, etc.)
        return handleConflict(err)
    }
    
    // Check for validation errors
    if IsValidationError(err) {
        // Handle any type of validation error
        return handleValidation(err)
    }
    
    // Check for authorization errors
    if IsUnauthorized(err) || IsForbidden(err) {
        // Handle authorization issues
        return handleAuthError(err)
    }
}
```

### Pattern Matching with Error Codes

```go
func ErrorCodePatterns(err error) {
    errorCode := GetErrorCode(err)
    
    switch errorCode {
    case "USER_NOT_FOUND", "ENTITY_NOT_FOUND", "ROLE_NOT_FOUND":
        // Handle all "not found" errors
        return handleNotFound(err)
        
    case "USER_EXISTS", "ENTITY_NAME_EXISTS", "EMAIL_EXISTS":
        // Handle all "already exists" errors
        return handleConflict(err)
        
    case "INVALID_CREDENTIALS", "UNAUTHORIZED", "FORBIDDEN":
        // Handle all auth-related errors
        return handleAuthError(err)
        
    case "TENANT_LIMIT_EXCEEDED", "FEATURE_NOT_ENABLED":
        // Handle tenant limitation errors
        return handleTenantLimits(err)
    }
}
```

## HTTP Integration

### Automatic HTTP Response Generation

```go
func APIErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
    // 🆕 Automatic HTTP error conversion with rich context
    httpErr := ToHTTPError(err)
    
    // Add request context
    httpErr.RequestID = getRequestID(r.Context())
    httpErr.TraceID = getTraceID(r.Context())
    
    // Set headers
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("X-Request-ID", httpErr.RequestID)
    
    // Write response
    w.WriteHeader(httpErr.Status)
    json.NewEncoder(w).Encode(httpErr)
    
    // Log error with context
    LogErrorWithTenant(r.Context(), err)
}

// Example API responses for different error types:

// ErrUserNotFound response:
// {
//   "status": 404,
//   "code": "USER_NOT_FOUND", 
//   "message": "User not found",
//   "details": {
//     "user_id": "123"
//   },
//   "suggestions": [
//     "Verify the user ID is correct",
//     "Check if the user exists in your tenant"
//   ],
//   "timestamp": "2024-01-15T10:30:00Z",
//   "request_id": "req_abc123"
// }

// ErrTenantLimitExceeded response:
// {
//   "status": 402,
//   "code": "TENANT_LIMIT_EXCEEDED",
//   "message": "Tenant users limit exceeded", 
//   "details": {
//     "tenant_id": "tenant_456",
//     "limit_type": "users",
//     "current": 100,
//     "maximum": 100
//   },
//   "suggestions": [
//     "Upgrade your plan to increase limits",
//     "Contact sales for enterprise options"
//   ]
// }
```

### Middleware Integration

```go
func ErrorMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if recovered := recover(); recovered != nil {
                // Convert panic to error
                err := fmt.Errorf("panic: %v", recovered)
                
                // 🆕 Use error for panics
                enhancedErr := NewBusinessError("INTERNAL_PANIC", "An unexpected error occurred").
                    WithHTTPStatus(http.StatusInternalServerError).
                    WithCategory(CategorySystem).
                    WithSeverity(SeverityCritical).
                    WithDetail("panic_value", recovered)
                
                APIErrorHandler(w, r, enhancedErr)
            }
        }()
        
        next.ServeHTTP(w, r)
    })
}
```

## Contextual Error Constructors

### Using  Constructors

```go
func ErrorConstructors(ctx context.Context) {
    // 🆕 Contextual user errors
    userErr := NewUserNotFoundError("user_123")
    emailErr := NewUserNotFoundByEmailError("user@example.com")
    
    // 🆕 Contextual entity errors  
    entityErr := NewEntityNotFoundError("entity_456")
    nameErr := NewEntityNameExistsError("Accounting Department")
    codeErr := NewEntityCodeExistsError("ACC001")
    
    // 🆕 Hierarchy errors with context
    circularErr := NewCircularReferenceError("entity_1", "entity_2")
    
    // 🆕 Authentication errors with attempt tracking
    credErr := NewInvalidCredentialsError("user@example.com", 3) // 3rd attempt
    
    // 🆕 Feature errors with plan requirements
    featureErr := NewFeatureNotEnabledError("advanced_reporting", "Professional")
    
    // 🆕 Invitation errors with timing
    inviteErr := NewInvitationExpiredError("inv_123", "2024-01-15T10:00:00Z")
}

// 🆕 Custom error variations
func CustomErrorVariations(ctx context.Context) error {
    // Add custom details to predefined errors
    return ErrUserNotFound.
        WithDetail("lookup_method", "email").
        WithDetail("search_value", "user@example.com").
        WithSuggestion("Try searching by username instead")
    
    // Chain multiple suggestions
    return ErrEntityHasChildren.
        WithDetail("child_count", 5).
        WithSuggestion("Move child entities to another parent").
        WithSuggestion("Delete child entities first").
        WithSuggestion("Archive the entity instead of deleting")
}
```

## Migration Examples

### Gradual Migration Pattern

```go
// Phase 1: Drop-in replacement (no changes needed)
func Phase1_ExistingCode() error {
    // Your existing code works unchanged
    if userNotFound {
        return ErrUserNotFound // Now but same interface
    }
    return nil
}

// Phase 2: Start using features
func Phase2_Usage(ctx context.Context) error {
    if userNotFound {
        // 🆕 Add context and details
        return NewUserNotFoundError(userID).
            WithDetail("search_method", "database")
    }
    return nil
}

// Phase 3: Full enhancement
func Phase3_Fully(ctx context.Context) error {
    if userNotFound {
        // 🆝 Full context-aware error
        return NewUserNotFoundByEmailError(email).
            WithDetail("tenant_id", getTenantID(ctx)).
            WithDetail("search_timestamp", time.Now()).
            WithSuggestion("Contact support with this error ID")
    }
    return nil
}
```

### Legacy Error Upgrade

```go
// 🆕 Upgrade simple errors to errors
func UpgradeLegacyErrors(err error) error {
    // Automatic upgrade for common errors
    upgradedErr := UpgradeError(err)
    
    // Manual upgrade for specific cases
    if err.Error() == "user not found" {
        return ErrUserNotFound
    }
    
    return upgradedErr
}

// 🆕 Wrap legacy service errors
func WrapLegacyService(legacyService LegacyService) Service {
    return &legacyServiceWrapper{legacy: legacyService}
}

type legacyServiceWrapper struct {
    legacy LegacyService
}

func (w *legacyServiceWrapper) GetUser(ctx context.Context, userID string) (*User, error) {
    user, err := w.legacy.GetUser(userID)
    if err != nil {
        // 🆕 Convert legacy errors to errors
        if err.Error() == "user not found" {
            return nil, NewUserNotFoundError(userID)
        }
        return nil, UpgradeError(err)
    }
    return user, nil
}
```

## Best Practices

### Error Construction Patterns

```go
// ✅ Good: Use contextual constructors
func GoodErrorUsage(ctx context.Context, userID string) error {
    return NewUserNotFoundError(userID).
        WithDetail("lookup_method", "id").
        WithSuggestion("Verify the user ID format")
}

// ❌ Bad: Generic error without context
func BadErrorUsage() error {
    return ErrUserNotFound // Missing specific context
}

// ✅ Good: Chain error details
func ChainErrorDetails(ctx context.Context, operation string) error {
    return ErrFeatureNotEnabled.
        WithDetail("operation", operation).
        WithDetail("tenant_plan", getTenantPlan(ctx)).
        WithSuggestion("Contact sales for plan upgrade information")
}
```

### Error Logging Patterns

```go
// ✅ Good: Structured error logging
func StructuredErrorLogging(ctx context.Context, err error) {
    if be, ok := err.(*BusinessError); ok {
        log.WithFields(map[string]interface{}{
            "error_code":   be.Code,
            "error_category": be.Category,
            "tenant_id":    be.TenantID,
            "user_id":      be.UserID,
            "suggestions":  be.Suggestions,
            "details":      be.Details,
            "http_status":  be.HTTPStatus,
            "retryable":    be.Retryable,
        }).Error(be.Message)
    }
}

// ✅ Good: Error metrics collection
func CollectErrorMetrics(err error) {
    errorCode := GetErrorCode(err)
    httpStatus := GetHTTPStatus(err)
    isRetryable := IsTemporaryError(err)
    
    metrics.Counter("errors_total").
        WithLabelValues(errorCode, fmt.Sprintf("%d", httpStatus)).
        Inc()
    
    if isRetryable {
        metrics.Counter("retryable_errors_total").
            WithLabelValues(errorCode).
            Inc()
    }
}
```

### Testing Patterns

```go
func TestErrors(t *testing.T) {
    // ✅ Test error types
    err := NewUserNotFoundError("user_123")
    assert.True(t, IsUserNotFound(err))
    assert.True(t, errors.Is(err, ErrUserNotFound))
    
    // ✅ Test error details
    be, ok := err.(*BusinessError)
    assert.True(t, ok)
    assert.Equal(t, "USER_NOT_FOUND", be.Code)
    assert.Equal(t, http.StatusNotFound, be.HTTPStatus)
    assert.Equal(t, "user_123", be.Details["user_id"])
    
    // ✅ Test HTTP conversion
    httpErr := ToHTTPError(err)
    assert.Equal(t, http.StatusNotFound, httpErr.Status)
    assert.Equal(t, "USER_NOT_FOUND", httpErr.Code)
}
```

This predefined errors system provides rich, contextual error handling while maintaining complete backward compatibility with your existing error checking patterns.
