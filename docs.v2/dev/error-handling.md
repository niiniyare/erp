# Error Handling Guide

 error handling strategies for building resilient applications with proper error propagation, classification, and observability.

## 🎯 Error Handling Philosophy

Our error handling follows these principles:
1. **Fail Fast**: Detect and handle errors as early as possible
2. **Contextual Information**: Provide meaningful error messages and context
3. **Error Classification**: Distinguish between different types of errors
4. **Observability**: Log and trace errors for debugging and monitoring
5. **Graceful Degradation**: Handle errors gracefully without breaking the user experience

## 📊 Error Flow Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                 Error Classification                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   Domain    │  │ Validation  │  │Infrastructure│       │
│  │   Errors    │  │   Errors    │  │   Errors     │       │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Error Flow                               │
│                                                             │
│  Database Error → Repository (Wrap) → Service (Handle)     │
│                     → Handler (Format) → Client            │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                 Error Observability                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   Logging   │  │   Tracing   │  │   Metrics   │        │
│  │  (Context)  │  │  (Spans)    │  │ (Counters)  │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
```

## 🏗️ Error Type System

### 1. **Domain Errors**

Domain errors represent business rule violations and expected error conditions.

```go
// internal/shared/errors/domain.go
package errors

import (
    "errors"
    "fmt"
)

// Core domain errors
var (
    ErrTenantNotFound         = errors.New("tenant not found")
    ErrSubdomainAlreadyExists = errors.New("subdomain already exists")
    ErrInvalidTenantStatus    = errors.New("invalid tenant status")
    ErrTenantInactive         = errors.New("tenant is inactive")
    ErrTenantSuspended        = errors.New("tenant is suspended")
    ErrTenantDeleted          = errors.New("tenant is deleted")
    
    // User domain errors
    ErrUserNotFound           = errors.New("user not found")
    ErrUserAlreadyExists      = errors.New("user already exists")
    ErrInvalidCredentials     = errors.New("invalid credentials")
    ErrUserInactive           = errors.New("user is inactive")
    
    // Permission domain errors
    ErrPermissionDenied       = errors.New("permission denied")
    ErrInsufficientRole       = errors.New("insufficient role")
)

// Domain error with context
type DomainError struct {
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}

func (e DomainError) Error() string {
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func NewDomainError(code, message string) *DomainError {
    return &DomainError{
        Code:    code,
        Message: message,
        Details: make(map[string]interface{}),
    }
}

func (e *DomainError) WithDetail(key string, value interface{}) *DomainError {
    e.Details[key] = value
    return e
}

// Predefined domain errors
func ErrTenantNotFoundWithID(id string) error {
    return NewDomainError("TENANT_NOT_FOUND", "Tenant not found").
        WithDetail("tenant_id", id)
}

func ErrSubdomainExistsWithValue(subdomain string) error {
    return NewDomainError("SUBDOMAIN_EXISTS", "Subdomain already exists").
        WithDetail("subdomain", subdomain)
}
```

### 2. **Validation Errors**

Validation errors provide detailed field-level error information.

```go
// internal/shared/errors/validation.go
package errors

import (
    "fmt"
    "strings"
)

// ValidationError represents field-level validation errors
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
    Value   string `json:"value,omitempty"`
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
}

func NewValidationError(field, message string) *ValidationError {
    return &ValidationError{
        Field:   field,
        Message: message,
    }
}

func (e *ValidationError) WithValue(value string) *ValidationError {
    e.Value = value
    return e
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
    Errors []ValidationError `json:"errors"`
}

func (e ValidationErrors) Error() string {
    var messages []string
    for _, err := range e.Errors {
        messages = append(messages, err.Error())
    }
    return strings.Join(messages, "; ")
}

func NewValidationErrors() *ValidationErrors {
    return &ValidationErrors{
        Errors: make([]ValidationError, 0),
    }
}

func (e *ValidationErrors) Add(field, message string) *ValidationErrors {
    e.Errors = append(e.Errors, ValidationError{
        Field:   field,
        Message: message,
    })
    return e
}

func (e *ValidationErrors) AddWithValue(field, message, value string) *ValidationErrors {
    e.Errors = append(e.Errors, ValidationError{
        Field:   field,
        Message: message,
        Value:   value,
    })
    return e
}

func (e *ValidationErrors) HasErrors() bool {
    return len(e.Errors) > 0
}

// Common validation errors
var (
    ErrValidationRequired = "field is required"
    ErrValidationMinLength = "field must be at least %d characters"
    ErrValidationMaxLength = "field must be at most %d characters"
    ErrValidationEmail     = "field must be a valid email address"
    ErrValidationURL       = "field must be a valid URL"
    ErrValidationUUID      = "field must be a valid UUID"
)

// Validation helper functions
func ValidateRequired(field, value string) *ValidationError {
    if strings.TrimSpace(value) == "" {
        return NewValidationError(field, ErrValidationRequired)
    }
    return nil
}

func ValidateMinLength(field, value string, minLength int) *ValidationError {
    if len(value) < minLength {
        return NewValidationError(field, fmt.Sprintf(ErrValidationMinLength, minLength))
    }
    return nil
}

func ValidateMaxLength(field, value string, maxLength int) *ValidationError {
    if len(value) > maxLength {
        return NewValidationError(field, fmt.Sprintf(ErrValidationMaxLength, maxLength))
    }
    return nil
}
```

### 3. **Infrastructure Errors**

Infrastructure errors represent external system failures and technical issues.

```go
// internal/shared/errors/infrastructure.go
package errors

import (
    "fmt"
    "net/http"
)

// InfrastructureError represents external system failures
type InfrastructureError struct {
    Component string `json:"component"`
    Operation string `json:"operation"`
    Message   string `json:"message"`
    Cause     error  `json:"-"`
}

func (e InfrastructureError) Error() string {
    return fmt.Sprintf("%s.%s: %s", e.Component, e.Operation, e.Message)
}

func (e InfrastructureError) Unwrap() error {
    return e.Cause
}

func NewInfrastructureError(component, operation, message string, cause error) *InfrastructureError {
    return &InfrastructureError{
        Component: component,
        Operation: operation,
        Message:   message,
        Cause:     cause,
    }
}

// HTTP Client Error
type HTTPError struct {
    StatusCode int               `json:"status_code"`
    Method     string            `json:"method"`
    URL        string            `json:"url"`
    Message    string            `json:"message"`
    Headers    map[string]string `json:"headers,omitempty"`
}

func (e HTTPError) Error() string {
    return fmt.Sprintf("HTTP %d: %s %s - %s", e.StatusCode, e.Method, e.URL, e.Message)
}

func NewHTTPError(statusCode int, method, url, message string) *HTTPError {
    return &HTTPError{
        StatusCode: statusCode,
        Method:     method,
        URL:        url,
        Message:    message,
        Headers:    make(map[string]string),
    }
}

// Database Error
type DatabaseError struct {
    Operation string `json:"operation"`
    Table     string `json:"table,omitempty"`
    Message   string `json:"message"`
    Cause     error  `json:"-"`
}

func (e DatabaseError) Error() string {
    if e.Table != "" {
        return fmt.Sprintf("database.%s.%s: %s", e.Operation, e.Table, e.Message)
    }
    return fmt.Sprintf("database.%s: %s", e.Operation, e.Message)
}

func (e DatabaseError) Unwrap() error {
    return e.Cause
}

func NewDatabaseError(operation, table, message string, cause error) *DatabaseError {
    return &DatabaseError{
        Operation: operation,
        Table:     table,
        Message:   message,
        Cause:     cause,
    }
}
```

## 🔄 Error Handling by Layer

### 1. **Repository Layer Error Handling**

Repository layer converts database errors to domain errors and wraps infrastructure errors.

```go
// internal/core/tenant/repository.go
package tenant

import (
    "context"
    "fmt"
    "strings"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
    
    "internal/platform/db"
    "internal/shared/errors"
    "internal/shared/logger"
    "internal/shared/tracing"
)

func (r *repository) Create(ctx context.Context, tenant *Tenant) error {
    ctx, span := r.tracing.StartSpan(ctx, "repository.create_tenant")
    defer span.End()
    
    params, err := ToSQLCCreateParams(tenant)
    if err != nil {
        span.RecordError(err)
        r.logger.ErrorContext(ctx, "Failed to convert tenant to SQLC params", logger.Fields{
            "error": err.Error(),
        })
        return fmt.Errorf("failed to convert tenant: %w", err)
    }
    
    _, err = r.store.CreateTenant(ctx, params)
    if err != nil {
        // Convert database errors to domain errors
        if isDuplicateKeyError(err) {
            if strings.Contains(err.Error(), "tenants_subdomain_key") {
                span.SetAttributes(attribute.String("error.type", "duplicate_subdomain"))
                r.logger.WarnContext(ctx, "Subdomain already exists", logger.Fields{
                    "subdomain": *tenant.Subdomain,
                })
                return errors.ErrSubdomainAlreadyExists
            }
            
            if strings.Contains(err.Error(), "tenants_slug_key") {
                span.SetAttributes(attribute.String("error.type", "duplicate_slug"))
                r.logger.WarnContext(ctx, "Slug already exists", logger.Fields{
                    "slug": tenant.Slug,
                })
                return errors.NewDomainError("SLUG_EXISTS", "Slug already exists").
                    WithDetail("slug", tenant.Slug)
            }
        }
        
        // Wrap infrastructure errors
        span.RecordError(err)
        span.SetAttributes(attribute.String("error.type", "database_error"))
        
        r.logger.ErrorContext(ctx, "Database operation failed", logger.Fields{
            "operation": "create_tenant",
            "error":     err.Error(),
        })
        
        return errors.NewDatabaseError("create", "tenants", "Failed to create tenant", err)
    }
    
    span.SetStatus(codes.Ok, "Tenant created successfully")
    return nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error) {
    ctx, span := r.tracing.StartSpan(ctx, "repository.get_tenant_by_id")
    defer span.End()
    
    span.SetAttributes(attribute.String("tenant.id", id.String()))
    
    sqlcTenant, err := r.store.GetTenantByID(ctx, id)
    if err != nil {
        if isNoRowsError(err) {
            span.SetAttributes(attribute.String("error.type", "not_found"))
            r.logger.DebugContext(ctx, "Tenant not found", logger.Fields{
                "tenant_id": id.String(),
            })
            return nil, errors.ErrTenantNotFound
        }
        
        // Infrastructure error
        span.RecordError(err)
        span.SetAttributes(attribute.String("error.type", "database_error"))
        
        r.logger.ErrorContext(ctx, "Failed to get tenant from database", logger.Fields{
            "tenant_id": id.String(),
            "error":     err.Error(),
        })
        
        return nil, errors.NewDatabaseError("select", "tenants", "Failed to get tenant", err)
    }
    
    tenant, err := FromSQLCTenant(sqlcTenant)
    if err != nil {
        span.RecordError(err)
        r.logger.ErrorContext(ctx, "Failed to convert SQLC tenant", logger.Fields{
            "tenant_id": id.String(),
            "error":     err.Error(),
        })
        return nil, fmt.Errorf("failed to convert tenant: %w", err)
    }
    
    span.SetStatus(codes.Ok, "Tenant retrieved successfully")
    return tenant, nil
}

// Helper functions for error classification
func isDuplicateKeyError(err error) bool {
    return strings.Contains(err.Error(), "duplicate key") ||
           strings.Contains(err.Error(), "UNIQUE constraint")
}

func isNoRowsError(err error) bool {
    return err == pgx.ErrNoRows ||
           strings.Contains(err.Error(), "no rows in result set")
}

func isForeignKeyError(err error) bool {
    return strings.Contains(err.Error(), "foreign key") ||
           strings.Contains(err.Error(), "FOREIGN KEY constraint")
}

func isConnectionError(err error) bool {
    return strings.Contains(err.Error(), "connection") ||
           strings.Contains(err.Error(), "timeout") ||
           strings.Contains(err.Error(), "network")
}
```

### 2. **Service Layer Error Handling**

Service layer handles business logic errors and orchestrates error responses.

```go
// internal/core/tenant/service.go
package tenant

import (
    "context"
    "fmt"

    "github.com/google/uuid"
    
    "internal/shared/errors"
    "internal/shared/logger"
    "internal/shared/tracing"
)

func (s *service) CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error) {
    ctx, span := s.tracing.StartSpan(ctx, "service.create_tenant")
    defer span.End()
    
    // 1. Validate request
    if err := s.validateCreateRequest(ctx, req); err != nil {
        span.SetAttributes(attribute.String("error.type", "validation_error"))
        span.RecordError(err)
        
        s.logger.WarnContext(ctx, "Tenant creation validation failed", logger.Fields{
            "error": err.Error(),
        })
        
        return nil, err  // Pass validation errors directly
    }
    
    // 2. Business logic validation
    if req.Subdomain != nil {
        exists, err := s.repo.SubdomainExists(ctx, *req.Subdomain)
        if err != nil {
            // Infrastructure error from repository
            span.RecordError(err)
            s.logger.ErrorContext(ctx, "Failed to check subdomain existence", logger.Fields{
                "subdomain": *req.Subdomain,
                "error":     err.Error(),
            })
            
            return nil, fmt.Errorf("failed to validate subdomain: %w", err)
        }
        
        if exists {
            span.SetAttributes(attribute.String("error.type", "business_error"))
            
            s.logger.WarnContext(ctx, "Subdomain already exists", logger.Fields{
                "subdomain": *req.Subdomain,
            })
            
            return nil, errors.ErrSubdomainExistsWithValue(*req.Subdomain)
        }
    }
    
    // 3. Create tenant entity
    tenant := &Tenant{
        ID:           uuid.New(),
        Name:         req.Name,
        Slug:         req.Slug,
        Email:        req.Email,
        Subdomain:    req.Subdomain,
        Status:       StatusActive,
        // ... other fields
    }
    
    // 4. Persist tenant
    if err := s.repo.Create(ctx, tenant); err != nil {
        // Handle different types of repository errors
        switch {
        case errors.Is(err, errors.ErrSubdomainAlreadyExists):
            // Domain error - pass through
            span.SetAttributes(attribute.String("error.type", "domain_error"))
            return nil, err
            
        case isInfrastructureError(err):
            // Infrastructure error - wrap with service context
            span.RecordError(err)
            span.SetAttributes(attribute.String("error.type", "infrastructure_error"))
            
            s.logger.ErrorContext(ctx, "Infrastructure error during tenant creation", logger.Fields{
                "error": err.Error(),
            })
            
            return nil, fmt.Errorf("service: failed to create tenant: %w", err)
            
        default:
            // Unknown error - wrap and log
            span.RecordError(err)
            span.SetAttributes(attribute.String("error.type", "unknown_error"))
            
            s.logger.ErrorContext(ctx, "Unknown error during tenant creation", logger.Fields{
                "error": err.Error(),
            })
            
            return nil, fmt.Errorf("service: unexpected error creating tenant: %w", err)
        }
    }
    
    // 5. Success
    span.SetStatus(codes.Ok, "Tenant created successfully")
    s.logger.InfoContext(ctx, "Tenant created successfully", logger.Fields{
        "tenant_id": tenant.ID.String(),
    })
    
    return tenant, nil
}

func (s *service) GetTenant(ctx context.Context, id uuid.UUID) (*Tenant, error) {
    ctx, span := s.tracing.StartSpan(ctx, "service.get_tenant")
    defer span.End()
    
    span.SetAttributes(attribute.String("tenant.id", id.String()))
    
    tenant, err := s.repo.GetByID(ctx, id)
    if err != nil {
        // Let domain errors (like ErrTenantNotFound) pass through
        if errors.Is(err, errors.ErrTenantNotFound) {
            span.SetAttributes(attribute.String("error.type", "domain_error"))
            s.logger.InfoContext(ctx, "Tenant not found", logger.Fields{
                "tenant_id": id.String(),
            })
            return nil, err
        }
        
        // Wrap infrastructure errors
        if isInfrastructureError(err) {
            span.RecordError(err)
            span.SetAttributes(attribute.String("error.type", "infrastructure_error"))
            
            s.logger.ErrorContext(ctx, "Infrastructure error getting tenant", logger.Fields{
                "tenant_id": id.String(),
                "error":     err.Error(),
            })
            
            return nil, fmt.Errorf("service: failed to get tenant: %w", err)
        }
        
        // Unknown error
        span.RecordError(err)
        s.logger.ErrorContext(ctx, "Unknown error getting tenant", logger.Fields{
            "tenant_id": id.String(),
            "error":     err.Error(),
        })
        
        return nil, fmt.Errorf("service: unexpected error getting tenant: %w", err)
    }
    
    span.SetStatus(codes.Ok, "Tenant retrieved successfully")
    return tenant, nil
}

func (s *service) validateCreateRequest(ctx context.Context, req CreateTenantRequest) error {
    validationErrors := errors.NewValidationErrors()
    
    // Name validation
    if err := errors.ValidateRequired("name", req.Name); err != nil {
        validationErrors.Add(err.Field, err.Message)
    } else if err := errors.ValidateMinLength("name", req.Name, 3); err != nil {
        validationErrors.Add(err.Field, err.Message)
    } else if err := errors.ValidateMaxLength("name", req.Name, 100); err != nil {
        validationErrors.Add(err.Field, err.Message)
    }
    
    // Slug validation
    if err := errors.ValidateRequired("slug", req.Slug); err != nil {
        validationErrors.Add(err.Field, err.Message)
    } else if !isValidSlug(req.Slug) {
        validationErrors.Add("slug", "Slug must contain only lowercase letters, numbers, and hyphens")
    }
    
    // Email validation
    if err := errors.ValidateRequired("email", req.Email); err != nil {
        validationErrors.Add(err.Field, err.Message)
    } else if !isValidEmail(req.Email) {
        validationErrors.Add("email", "Invalid email format")
    }
    
    // Subdomain validation
    if req.Subdomain != nil {
        if err := errors.ValidateMinLength("subdomain", *req.Subdomain, 3); err != nil {
            validationErrors.Add(err.Field, err.Message)
        } else if !isValidSubdomain(*req.Subdomain) {
            validationErrors.Add("subdomain", "Invalid subdomain format")
        } else if isReservedSubdomain(*req.Subdomain) {
            validationErrors.Add("subdomain", "Subdomain is reserved")
        }
    }
    
    if validationErrors.HasErrors() {
        return validationErrors
    }
    
    return nil
}

// Helper function to identify infrastructure errors
func isInfrastructureError(err error) bool {
    var dbErr *errors.DatabaseError
    var httpErr *errors.HTTPError
    var infraErr *errors.InfrastructureError
    
    return errors.As(err, &dbErr) ||
           errors.As(err, &httpErr) ||
           errors.As(err, &infraErr)
}
```

### 3. **API Layer Error Handling**

API layer formats errors for HTTP responses and provides appropriate status codes.

```go
// internal/api/handlers/tenant.go
package handlers

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    
    "internal/core/tenant"
    "internal/shared/errors"
    "internal/shared/logger"
    "internal/shared/tracing"
)

func (h *TenantHandler) CreateTenant(c *gin.Context) {
    ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)
    
    ctx, span := h.tracing.StartSpan(ctx, "http.create_tenant")
    defer span.End()
    
    var req tenant.CreateTenantRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // JSON binding error
        span.SetAttributes(attribute.String("error.type", "binding_error"))
        span.RecordError(err)
        
        h.logger.WarnContext(ctx, "Invalid request format", logger.Fields{
            "error": err.Error(),
        })
        
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Error: "Invalid request format",
            Code:  "INVALID_REQUEST",
            Details: map[string]interface{}{
                "message": err.Error(),
            },
        })
        return
    }
    
    newTenant, err := h.service.CreateTenant(ctx, req)
    if err != nil {
        h.handleCreateTenantError(c, ctx, err)
        return
    }
    
    span.SetStatus(codes.Ok, "Tenant created successfully")
    c.JSON(http.StatusCreated, newTenant)
}

func (h *TenantHandler) GetTenant(c *gin.Context) {
    ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)
    
    ctx, span := h.tracing.StartSpan(ctx, "http.get_tenant")
    defer span.End()
    
    // Parse and validate UUID
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        span.SetAttributes(attribute.String("error.type", "validation_error"))
        span.RecordError(err)
        
        h.logger.WarnContext(ctx, "Invalid tenant ID format", logger.Fields{
            "tenant_id": idStr,
        })
        
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Error: "Invalid tenant ID format",
            Code:  "INVALID_UUID",
            Details: map[string]interface{}{
                "tenant_id": idStr,
            },
        })
        return
    }
    
    tenant, err := h.service.GetTenant(ctx, id)
    if err != nil {
        h.handleGetTenantError(c, ctx, err, id)
        return
    }
    
    span.SetStatus(codes.Ok, "Tenant retrieved successfully")
    c.JSON(http.StatusOK, tenant)
}

// Error handling helper methods
func (h *TenantHandler) handleCreateTenantError(c *gin.Context, ctx context.Context, err error) {
    // Validation errors
    var validationErr *errors.ValidationErrors
    if errors.As(err, &validationErr) {
        h.logger.WarnContext(ctx, "Tenant creation validation failed", logger.Fields{
            "validation_errors": validationErr.Errors,
        })
        
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Error:   "Validation failed",
            Code:    "VALIDATION_ERROR",
            Details: map[string]interface{}{
                "validation_errors": validationErr.Errors,
            },
        })
        return
    }
    
    // Domain errors
    var domainErr *errors.DomainError
    if errors.As(err, &domainErr) {
        status := h.mapDomainErrorToHTTPStatus(domainErr.Code)
        
        h.logger.WarnContext(ctx, "Domain error during tenant creation", logger.Fields{
            "error_code": domainErr.Code,
            "error":      domainErr.Message,
            "details":    domainErr.Details,
        })
        
        c.JSON(status, ErrorResponse{
            Error:   domainErr.Message,
            Code:    domainErr.Code,
            Details: domainErr.Details,
        })
        return
    }
    
    // Check for specific domain errors
    switch {
    case errors.Is(err, errors.ErrSubdomainAlreadyExists):
        h.logger.WarnContext(ctx, "Subdomain already exists", logger.Fields{
            "error": err.Error(),
        })
        
        c.JSON(http.StatusConflict, ErrorResponse{
            Error: "Subdomain already exists",
            Code:  "SUBDOMAIN_EXISTS",
        })
        
    case isInfrastructureError(err):
        h.logger.ErrorContext(ctx, "Infrastructure error during tenant creation", logger.Fields{
            "error": err.Error(),
        })
        
        c.JSON(http.StatusServiceUnavailable, ErrorResponse{
            Error: "Service temporarily unavailable",
            Code:  "SERVICE_UNAVAILABLE",
        })
        
    default:
        h.logger.ErrorContext(ctx, "Internal error during tenant creation", logger.Fields{
            "error": err.Error(),
        })
        
        c.JSON(http.StatusInternalServerError, ErrorResponse{
            Error: "Internal server error",
            Code:  "INTERNAL_ERROR",
        })
    }
}

func (h *TenantHandler) handleGetTenantError(c *gin.Context, ctx context.Context, err error, tenantID uuid.UUID) {
    switch {
    case errors.Is(err, errors.ErrTenantNotFound):
        h.logger.InfoContext(ctx, "Tenant not found", logger.Fields{
            "tenant_id": tenantID.String(),
        })
        
        c.JSON(http.StatusNotFound, ErrorResponse{
            Error: "Tenant not found",
            Code:  "TENANT_NOT_FOUND",
            Details: map[string]interface{}{
                "tenant_id": tenantID.String(),
            },
        })
        
    case isInfrastructureError(err):
        h.logger.ErrorContext(ctx, "Infrastructure error getting tenant", logger.Fields{
            "tenant_id": tenantID.String(),
            "error":     err.Error(),
        })
        
        c.JSON(http.StatusServiceUnavailable, ErrorResponse{
            Error: "Service temporarily unavailable",
            Code:  "SERVICE_UNAVAILABLE",
        })
        
    default:
        h.logger.ErrorContext(ctx, "Internal error getting tenant", logger.Fields{
            "tenant_id": tenantID.String(),
            "error":     err.Error(),
        })
        
        c.JSON(http.StatusInternalServerError, ErrorResponse{
            Error: "Internal server error",
            Code:  "INTERNAL_ERROR",
        })
    }
}

// Error response structure
type ErrorResponse struct {
    Error   string                 `json:"error"`
    Code    string                 `json:"code"`
    Details map[string]interface{} `json:"details,omitempty"`
}

// Map domain error codes to HTTP status codes
func (h *TenantHandler) mapDomainErrorToHTTPStatus(code string) int {
    switch code {
    case "TENANT_NOT_FOUND":
        return http.StatusNotFound
    case "SUBDOMAIN_EXISTS", "SLUG_EXISTS":
        return http.StatusConflict
    case "VALIDATION_ERROR":
        return http.StatusBadRequest
    case "PERMISSION_DENIED":
        return http.StatusForbidden
    case "TENANT_INACTIVE", "TENANT_SUSPENDED":
        return http.StatusForbidden
    default:
        return http.StatusBadRequest
    }
}
```

## 📊 Error Observability

### 1. **Error Logging Patterns**

#### Structured Error Logging
```go
// Consistent error logging across layers
func (s *service) CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error) {
    tenant, err := s.repo.Create(ctx, tenant)
    if err != nil {
        // Log with context and classification
        s.logger.ErrorContext(ctx, "Failed to create tenant", logger.Fields{
            "operation":    "create_tenant",
            "tenant_name":  req.Name,
            "error_type":   classifyError(err),
            "error":        err.Error(),
            "tenant_id":    tenant.ID.String(),
        })
        
        return nil, fmt.Errorf("service: failed to create tenant: %w", err)
    }
    
    return tenant, nil
}

// Error classification for consistent logging
func classifyError(err error) string {
    switch {
    case errors.Is(err, errors.ErrTenantNotFound):
        return "not_found"
    case errors.Is(err, errors.ErrSubdomainAlreadyExists):
        return "conflict"
    case isValidationError(err):
        return "validation"
    case isInfrastructureError(err):
        return "infrastructure"
    default:
        return "unknown"
    }
}
```

### 2. **Error Tracing**

####  Error Spans
```go
func (r *repository) Create(ctx context.Context, tenant *Tenant) error {
    ctx, span := r.tracing.StartSpan(ctx, "repository.create_tenant")
    defer span.End()
    
    _, err := r.store.CreateTenant(ctx, params)
    if err != nil {
        // Record error with detailed attributes
        span.RecordError(err)
        span.SetStatus(codes.Error, "Database operation failed")
        span.SetAttributes(
            attribute.String("error.type", classifyDBError(err)),
            attribute.String("db.operation", "insert"),
            attribute.String("db.table", "tenants"),
            attribute.String("tenant.id", tenant.ID.String()),
        )
        
        return fmt.Errorf("failed to create tenant: %w", err)
    }
    
    span.SetStatus(codes.Ok, "Tenant created successfully")
    return nil
}
```

### 3. **Error Metrics**

#### Error Rate and Classification Metrics
```go
func (s *service) CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error) {
    timer := s.metrics.Timer("tenant_operation_duration", metrics.Fields{
        "operation": "create",
    })
    defer timer.Stop()
    
    tenant, err := s.processCreation(ctx, req)
    if err != nil {
        // Increment error metrics with classification
        s.metrics.IncrementCounter("tenant_operations_total", metrics.Fields{
            "operation":  "create",
            "status":     "error",
            "error_type": classifyError(err),
        })
        
        // Specific error metrics
        if errors.Is(err, errors.ErrSubdomainAlreadyExists) {
            s.metrics.IncrementCounter("tenant_subdomain_conflicts_total", metrics.Fields{})
        }
        
        return nil, err
    }
    
    // Success metrics
    s.metrics.IncrementCounter("tenant_operations_total", metrics.Fields{
        "operation": "create",
        "status":    "success",
    })
    
    return tenant, nil
}
```

## 🚨 Error Recovery Patterns

### 1. **Retry with Exponential Backoff**

#### Infrastructure Error Recovery
```go
// internal/shared/retry/retry.go
package retry

import (
    "context"
    "fmt"
    "math"
    "time"
)

type RetryConfig struct {
    MaxAttempts int
    BaseDelay   time.Duration
    MaxDelay    time.Duration
    Multiplier  float64
}

func WithRetry(ctx context.Context, config RetryConfig, operation func() error) error {
    var lastErr error
    
    for attempt := 0; attempt < config.MaxAttempts; attempt++ {
        if attempt > 0 {
            delay := calculateDelay(config, attempt)
            
            select {
            case <-time.After(delay):
                // Continue with retry
            case <-ctx.Done():
                return fmt.Errorf("operation cancelled: %w", ctx.Err())
            }
        }
        
        err := operation()
        if err == nil {
            return nil // Success
        }
        
        lastErr = err
        
        // Check if error is retryable
        if !isRetryableError(err) {
            return err
        }
    }
    
    return fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

func calculateDelay(config RetryConfig, attempt int) time.Duration {
    delay := float64(config.BaseDelay) * math.Pow(config.Multiplier, float64(attempt-1))
    if delay > float64(config.MaxDelay) {
        delay = float64(config.MaxDelay)
    }
    return time.Duration(delay)
}

func isRetryableError(err error) bool {
    // Check for retryable infrastructure errors
    var infraErr *errors.InfrastructureError
    if errors.As(err, &infraErr) {
        return infraErr.Component == "database" || infraErr.Component == "cache"
    }
    
    var httpErr *errors.HTTPError
    if errors.As(err, &httpErr) {
        // Retry on server errors, not client errors
        return httpErr.StatusCode >= 500
    }
    
    // Check for specific error types
    return isConnectionError(err) || isTimeoutError(err)
}
```

### 2. **Circuit Breaker Pattern**

#### Service Protection
```go
// internal/shared/circuitbreaker/circuitbreaker.go
package circuitbreaker

import (
    "context"
    "errors"
    "sync"
    "time"
)

type State int

const (
    StateClosed State = iota
    StateOpen
    StateHalfOpen
)

type CircuitBreaker struct {
    maxFailures  int
    resetTimeout time.Duration
    
    mu           sync.Mutex
    state        State
    failures     int
    lastFailTime time.Time
}

func New(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        maxFailures:  maxFailures,
        resetTimeout: resetTimeout,
        state:        StateClosed,
    }
}

func (cb *CircuitBreaker) Execute(ctx context.Context, operation func() error) error {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    // Check if circuit breaker should reset
    if cb.state == StateOpen && time.Since(cb.lastFailTime) > cb.resetTimeout {
        cb.state = StateHalfOpen
        cb.failures = 0
    }
    
    // Reject if circuit is open
    if cb.state == StateOpen {
        return errors.New("circuit breaker is open")
    }
    
    // Execute operation
    err := operation()
    
    if err != nil {
        cb.failures++
        cb.lastFailTime = time.Now()
        
        if cb.failures >= cb.maxFailures {
            cb.state = StateOpen
        }
        
        return err
    }
    
    // Success - reset failure count
    if cb.state == StateHalfOpen {
        cb.state = StateClosed
    }
    cb.failures = 0
    
    return nil
}
```

## ✅ Error Handling Best Practices

### 1. **Error Message Guidelines**

#### User-Friendly vs Technical Messages
```go
// ✅ Provide context-appropriate error messages
func (h *TenantHandler) CreateTenant(c *gin.Context) {
    _, err := h.service.CreateTenant(ctx, req)
    if err != nil {
        switch {
        case errors.Is(err, errors.ErrSubdomainAlreadyExists):
            // User-friendly message
            c.JSON(409, ErrorResponse{
                Error: "The subdomain you chose is already taken. Please try a different one.",
                Code:  "SUBDOMAIN_EXISTS",
            })
            
        case isInfrastructureError(err):
            // Generic message for users, detailed logging for developers
            h.logger.ErrorContext(ctx, "Infrastructure error", logger.Fields{
                "error": err.Error(),  // Technical details in logs
            })
            
            c.JSON(503, ErrorResponse{
                Error: "We're experiencing technical difficulties. Please try again later.",
                Code:  "SERVICE_UNAVAILABLE",
            })
        }
    }
}
```

### 2. **Error Testing**

####  Error Testing
```go
func TestService_CreateTenant_SubdomainExists(t *testing.T) {
    mockRepo := new(mocks.MockRepository)
    service := NewService(mockRepo, nil, logger.NewTestLogger(), tracing.NewNoopTracer(), metrics.NewNoopMetrics())
    
    // Setup: subdomain already exists
    mockRepo.On("SubdomainExists", mock.Anything, "existing-subdomain").Return(true, nil)
    
    req := CreateTenantRequest{
        Name:      "Test Tenant",
        Subdomain: stringPtr("existing-subdomain"),
    }
    
    // Execute
    _, err := service.CreateTenant(context.Background(), req)
    
    // Verify error type and message
    require.Error(t, err)
    assert.True(t, errors.Is(err, errors.ErrSubdomainAlreadyExists))
    
    var domainErr *errors.DomainError
    if errors.As(err, &domainErr) {
        assert.Equal(t, "SUBDOMAIN_EXISTS", domainErr.Code)
        assert.Equal(t, "existing-subdomain", domainErr.Details["subdomain"])
    }
    
    mockRepo.AssertExpectations(t)
}
```

### 3. **Error Documentation**

#### API Error Documentation
```go
// @Summary Create a new tenant
// @Description Create a new tenant with the provided information
// @Tags tenants
// @Accept json
// @Produce json
// @Param tenant body CreateTenantRequest true "Tenant information"
// @Success 201 {object} Tenant
// @Failure 400 {object} ErrorResponse "Validation error"
// @Failure 409 {object} ErrorResponse "Subdomain already exists"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/tenants [post]
func (h *TenantHandler) CreateTenant(c *gin.Context) {
    // Implementation
}
```

---

📚 **Related Documentation**:
- [Architecture Overview](./architecture.md) - Understanding layer responsibilities
- [Code Examples](./code-examples.md) - See error handling in practice  
- [Best Practices](./best-practices.md) - Development guidelines
- [Observability](./observability.md) - Error monitoring and alerting
