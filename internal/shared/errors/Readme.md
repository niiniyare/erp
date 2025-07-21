# Predefined Errors Usage Guide

A comprehensive guide for using the enhanced predefined errors in your ERP/Accounting system while maintaining backward compatibility.

## Table of Contents

- [Backward Compatibility](#backward-compatibility)
- [Enhanced Error Features](#enhanced-error-features)
- [Usage by Domain](#usage-by-domain)
- [Error Checking Patterns](#error-checking-patterns)
- [HTTP Integration](#http-integration)
- [Contextual Error Constructors](#contextual-error-constructors)
- [Migration Examples](#migration-examples)
- [Best Practices](#best-practices)

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

## Enhanced Error Features

### Rich Error Information

```go
func EnhancedUserService(ctx context.Context, userID string) error {
    user, err := userRepo.GetUser(ctx, userID)
    if err != nil {
        if errors.Is(err, ErrUserNotFound) {
            // 🆕 Enhanced error now includes HTTP status, suggestions, etc.
            return err // Returns rich BusinessError with HTTP 404, suggestions
        }
        return err
    }
    return nil
}

func HandleEnhancedError(w http.ResponseWriter, err error) {
    if errors.Is(err, ErrUserNotFound) {
        // 🆕 Extract rich information from enhanced error
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
        // 🆕 Enhanced error with attempt tracking
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

### Enhanced Error Checking

```go
// ✅ Traditional checking still works
func TraditionalChecking(err error) {
    if errors.Is(err, ErrUserNotFound) {
        // Handle user not found
    }
}

// 🆕 Enhanced checking with error details
func EnhancedChecking(err error) {
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
                
                // 🆕 Use enhanced error for panics
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

### Using Enhanced Constructors

```go
func EnhancedErrorConstructors(ctx context.Context) {
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
        return ErrUserNotFound // Now enhanced but same interface
    }
    return nil
}

// Phase 2: Start using enhanced features
func Phase2_EnhancedUsage(ctx context.Context) error {
    if userNotFound {
        // 🆕 Add context and details
        return NewUserNotFoundError(userID).
            WithDetail("search_method", "database")
    }
    return nil
}

// Phase 3: Full enhancement
func Phase3_FullyEnhanced(ctx context.Context) error {
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
// 🆕 Upgrade simple errors to enhanced errors
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
func WrapLegacyService(legacyService LegacyService) EnhancedService {
    return &legacyServiceWrapper{legacy: legacyService}
}

type legacyServiceWrapper struct {
    legacy LegacyService
}

func (w *legacyServiceWrapper) GetUser(ctx context.Context, userID string) (*User, error) {
    user, err := w.legacy.GetUser(userID)
    if err != nil {
        // 🆕 Convert legacy errors to enhanced errors
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
func TestEnhancedErrors(t *testing.T) {
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

This enhanced predefined errors system provides rich, contextual error handling while maintaining complete backward compatibility with your existing error checking patterns.
