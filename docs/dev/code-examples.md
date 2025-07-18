# Code Examples Guide

Practical implementation examples showing how to build features following our Clean Architecture patterns with full observability integration.

## 🎯 Complete Feature Implementation

Let's walk through implementing a complete tenant management feature from API to database.

## 📱 API Layer Implementation

### Handler Example
```go
// internal/api/handlers/tenant.go
package handlers

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    
    "internal/core/tenant"
    "internal/shared/logger"
    "internal/shared/tracing"
    "internal/shared/metrics"
    "internal/shared/errors"
)

type TenantHandler struct {
    service tenant.Service
    logger  logger.Logger
    tracing tracing.Tracer
    metrics metrics.MetricsProvider
}

func NewTenantHandler(
    service tenant.Service,
    logger logger.Logger,
    tracing tracing.Tracer,
    metrics metrics.MetricsProvider,
) *TenantHandler {
    return &TenantHandler{
        service: service,
        logger:  logger,
        tracing: tracing,
        metrics: metrics,
    }
}

// CreateTenant handles POST /api/v1/tenants
func (h *TenantHandler) CreateTenant(c *gin.Context) {
    // Extract tracing context from HTTP headers
    ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)
    
    // Start HTTP span with comprehensive attributes
    ctx, span := h.tracing.StartSpan(ctx, "http.create_tenant", 
        tracing.WithSpanKind(tracing.SpanKindServer),
        tracing.WithAttributes(
            attribute.String("http.method", c.Request.Method),
            attribute.String("http.url", c.Request.URL.String()),
            attribute.String("http.route", "/api/v1/tenants"),
            attribute.String("http.user_agent", c.Request.UserAgent()),
            attribute.String("http.remote_addr", c.ClientIP()),
        ))
    defer span.End()
    
    // Start metrics timer and track active requests
    timer := h.metrics.Timer("http_request_duration", metrics.Fields{
        "method":   c.Request.Method,
        "endpoint": "/api/v1/tenants",
    })
    defer timer.Stop()
    
    h.metrics.IncrementGauge("http_active_requests", metrics.Fields{
        "method":   c.Request.Method,
        "endpoint": "/api/v1/tenants",
    })
    defer h.metrics.DecrementGauge("http_active_requests", metrics.Fields{
        "method":   c.Request.Method,
        "endpoint": "/api/v1/tenants",
    })
    
    // Log request start
    h.logger.InfoContext(ctx, "Processing create tenant request", logger.Fields{
        "method":     c.Request.Method,
        "path":       c.Request.URL.Path,
        "user_agent": c.Request.UserAgent(),
        "remote_ip":  c.ClientIP(),
    })
    
    // 1. Parse and validate request
    var req tenant.CreateTenantRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // Record validation error in span and metrics
        h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())
        span.SetAttributes(attribute.String("error.type", "validation_error"))
        
        h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
            "method":     c.Request.Method,
            "endpoint":   "/api/v1/tenants",
            "error_type": "validation_error",
            "status":     "400",
        })
        
        h.logger.WarnContext(ctx, "Invalid request format", logger.Fields{
            "error": err.Error(),
            "body":  c.Request.Body,
        })
        
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Invalid request format",
            "details": err.Error(),
        })
        return
    }
    
    // Add request data to span
    span.SetAttributes(
        attribute.String("tenant.name", req.Name),
        attribute.String("tenant.slug", req.Slug),
        attribute.String("tenant.email", req.Email),
    )
    if req.Subdomain != nil {
        span.SetAttributes(attribute.String("tenant.subdomain", *req.Subdomain))
    }
    
    // 2. Call service layer
    newTenant, err := h.service.CreateTenant(ctx, req)
    if err != nil {
        // Handle different types of business errors
        switch {
        case errors.Is(err, errors.ErrSubdomainAlreadyExists):
            span.SetAttributes(attribute.String("error.type", "business_error"))
            span.SetStatus(codes.Error, "Subdomain already exists")
            
            h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
                "method":     c.Request.Method,
                "endpoint":   "/api/v1/tenants",
                "error_type": "conflict",
                "status":     "409",
            })
            
            h.logger.WarnContext(ctx, "Subdomain already exists", logger.Fields{
                "subdomain": req.Subdomain,
                "error":     err.Error(),
            })
            
            c.JSON(http.StatusConflict, gin.H{
                "error": "Subdomain already exists",
                "code":  "SUBDOMAIN_EXISTS",
            })
            
        case errors.Is(err, errors.ErrValidation):
            span.SetAttributes(attribute.String("error.type", "validation_error"))
            span.SetStatus(codes.Error, "Validation failed")
            
            h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
                "method":     c.Request.Method,
                "endpoint":   "/api/v1/tenants",
                "error_type": "validation",
                "status":     "400",
            })
            
            h.logger.WarnContext(ctx, "Validation failed", logger.Fields{
                "error": err.Error(),
            })
            
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "Validation failed",
                "code":  "VALIDATION_ERROR",
            })
            
        default:
            // Internal server errors
            span.SetAttributes(attribute.String("error.type", "internal_error"))
            span.RecordError(err)
            span.SetStatus(codes.Error, "Internal server error")
            
            h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
                "method":     c.Request.Method,
                "endpoint":   "/api/v1/tenants",
                "error_type": "internal",
                "status":     "500",
            })
            
            h.logger.ErrorContext(ctx, "Failed to create tenant", logger.Fields{
                "error": err.Error(),
            })
            
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "Internal server error",
                "code":  "INTERNAL_ERROR",
            })
        }
        return
    }
    
    // 3. Success response
    span.SetAttributes(
        attribute.String("tenant.id", newTenant.ID.String()),
        attribute.String("response.status", "created"),
    )
    span.SetStatus(codes.Ok, "Tenant created successfully")
    
    h.metrics.IncrementCounter("http_requests_total", metrics.Fields{
        "method":   c.Request.Method,
        "endpoint": "/api/v1/tenants",
        "status":   "success",
    })
    
    h.metrics.IncrementCounter("tenants_created_total", metrics.Fields{})
    
    h.logger.InfoContext(ctx, "Tenant created successfully", logger.Fields{
        "tenant_id":   newTenant.ID.String(),
        "tenant_name": newTenant.Name,
        "status_code": http.StatusCreated,
    })
    
    c.JSON(http.StatusCreated, newTenant)
}

// GetTenant handles GET /api/v1/tenants/{id}
func (h *TenantHandler) GetTenant(c *gin.Context) {
    ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)
    
    ctx, span := h.tracing.StartSpan(ctx, "http.get_tenant",
        tracing.WithSpanKind(tracing.SpanKindServer))
    defer span.End()
    
    timer := h.metrics.Timer("http_request_duration", metrics.Fields{
        "method":   c.Request.Method,
        "endpoint": "/api/v1/tenants/{id}",
    })
    defer timer.Stop()
    
    // Parse and validate ID parameter
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        span.SetAttributes(attribute.String("error.type", "validation_error"))
        span.SetStatus(codes.Error, "Invalid UUID format")
        
        h.logger.WarnContext(ctx, "Invalid tenant ID format", logger.Fields{
            "tenant_id": idStr,
            "error":     err.Error(),
        })
        
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid tenant ID format",
            "code":  "INVALID_UUID",
        })
        return
    }
    
    span.SetAttributes(attribute.String("tenant.id", id.String()))
    
    // Get tenant from service
    tenant, err := h.service.GetTenant(ctx, id)
    if err != nil {
        if errors.Is(err, errors.ErrTenantNotFound) {
            span.SetAttributes(attribute.String("error.type", "not_found"))
            span.SetStatus(codes.Error, "Tenant not found")
            
            h.logger.InfoContext(ctx, "Tenant not found", logger.Fields{
                "tenant_id": id.String(),
            })
            
            c.JSON(http.StatusNotFound, gin.H{
                "error": "Tenant not found",
                "code":  "TENANT_NOT_FOUND",
            })
        } else {
            span.RecordError(err)
            span.SetStatus(codes.Error, "Internal server error")
            
            h.logger.ErrorContext(ctx, "Failed to get tenant", logger.Fields{
                "tenant_id": id.String(),
                "error":     err.Error(),
            })
            
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "Internal server error",
                "code":  "INTERNAL_ERROR",
            })
        }
        return
    }
    
    span.SetStatus(codes.Ok, "Tenant retrieved successfully")
    c.JSON(http.StatusOK, tenant)
}

// ListTenants handles GET /api/v1/tenants
func (h *TenantHandler) ListTenants(c *gin.Context) {
    ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)
    
    ctx, span := h.tracing.StartSpan(ctx, "http.list_tenants",
        tracing.WithSpanKind(tracing.SpanKindServer))
    defer span.End()
    
    // Parse query parameters
    filter := h.parseListFilter(c)
    
    span.SetAttributes(
        attribute.Int("list.page", filter.Page),
        attribute.Int("list.limit", filter.Limit),
        attribute.String("list.search", filter.Search),
    )
    
    result, err := h.service.ListTenants(ctx, filter)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "Failed to list tenants")
        
        h.logger.ErrorContext(ctx, "Failed to list tenants", logger.Fields{
            "filter": filter,
            "error":  err.Error(),
        })
        
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Internal server error",
        })
        return
    }
    
    // Set pagination headers
    c.Header("X-Total-Count", strconv.Itoa(result.Total))
    c.Header("X-Page", strconv.Itoa(filter.Page))
    c.Header("X-Per-Page", strconv.Itoa(filter.Limit))
    
    span.SetAttributes(
        attribute.Int("response.total", result.Total),
        attribute.Int("response.count", len(result.Items)),
    )
    span.SetStatus(codes.Ok, "Tenants listed successfully")
    
    c.JSON(http.StatusOK, result.Items)
}

// Helper method to parse list filters
func (h *TenantHandler) parseListFilter(c *gin.Context) tenant.ListFilter {
    filter := tenant.ListFilter{
        Page:  1,
        Limit: 20,
    }
    
    if pageStr := c.Query("page"); pageStr != "" {
        if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
            filter.Page = page
        }
    }
    
    if limitStr := c.Query("limit"); limitStr != "" {
        if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
            filter.Limit = limit
        }
    }
    
    filter.Search = c.Query("search")
    filter.Status = c.Query("status")
    
    return filter
}
```

## 🏢 Service Layer Implementation

### Service Interface and Implementation
```go
// internal/core/tenant/service.go
package tenant

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
    
    "internal/shared/cache"
    "internal/shared/logger"
    "internal/shared/tracing"
    "internal/shared/metrics"
    "internal/shared/errors"
)

type Service interface {
    CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error)
    GetTenant(ctx context.Context, id uuid.UUID) (*Tenant, error)
    GetTenantBySubdomain(ctx context.Context, subdomain string) (*Tenant, error)
    ListTenants(ctx context.Context, filter ListFilter) (*PaginatedResult, error)
    UpdateTenant(ctx context.Context, id uuid.UUID, req UpdateTenantRequest) (*Tenant, error)
    DeleteTenant(ctx context.Context, id uuid.UUID) error
}

type service struct {
    repo    Repository
    cache   cache.Cache
    logger  logger.Logger
    tracing tracing.Tracer
    metrics metrics.MetricsProvider
}

func NewService(
    repo Repository,
    cache cache.Cache,
    logger logger.Logger,
    tracing tracing.Tracer,
    metrics metrics.MetricsProvider,
) Service {
    return &service{
        repo:    repo,
        cache:   cache,
        logger:  logger,
        tracing: tracing,
        metrics: metrics,
    }
}

func (s *service) CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error) {
    // Start service span
    ctx, span := s.tracing.StartSpan(ctx, "service.create_tenant",
        tracing.WithSpanKind(tracing.SpanKindInternal))
    defer span.End()
    
    // Start business operation timer
    timer := s.metrics.Timer("tenant_operation_duration", metrics.Fields{
        "operation": "create",
    })
    defer timer.Stop()
    
    // Add business context to span
    span.SetAttributes(
        attribute.String("tenant.name", req.Name),
        attribute.String("tenant.slug", req.Slug),
        attribute.String("tenant.email", req.Email),
    )
    if req.Subdomain != nil {
        span.SetAttributes(attribute.String("tenant.subdomain", *req.Subdomain))
    }
    
    s.logger.InfoContext(ctx, "Starting tenant creation", logger.Fields{
        "tenant_name": req.Name,
        "tenant_slug": req.Slug,
        "tenant_email": req.Email,
    })
    
    // 1. Business validation
    if err := s.validateCreateRequest(ctx, req); err != nil {
        span.SetAttributes(attribute.String("error.type", "validation_error"))
        span.RecordError(err)
        span.SetStatus(codes.Error, "Validation failed")
        
        s.metrics.IncrementCounter("tenant_operations_total", metrics.Fields{
            "operation":  "create",
            "status":     "error",
            "error_type": "validation",
        })
        
        s.logger.WarnContext(ctx, "Tenant creation validation failed", logger.Fields{
            "error": err.Error(),
        })
        
        return nil, err
    }
    
    // 2. Check subdomain availability
    if req.Subdomain != nil {
        ctx, subdomainSpan := s.tracing.StartSpan(ctx, "service.validate_subdomain")
        subdomainSpan.SetAttributes(attribute.String("subdomain", *req.Subdomain))
        
        exists, err := s.repo.Exists(ctx, *req.Subdomain)
        if err != nil {
            subdomainSpan.RecordError(err)
            subdomainSpan.SetStatus(codes.Error, "Subdomain validation failed")
            subdomainSpan.End()
            
            s.logger.ErrorContext(ctx, "Failed to validate subdomain", logger.Fields{
                "subdomain": *req.Subdomain,
                "error":     err.Error(),
            })
            
            return nil, fmt.Errorf("failed to validate subdomain: %w", err)
        }
        
        subdomainSpan.SetAttributes(attribute.Bool("subdomain.exists", exists))
        subdomainSpan.End()
        
        if exists {
            span.SetAttributes(attribute.String("error.type", "subdomain_conflict"))
            span.SetStatus(codes.Error, "Subdomain already exists")
            
            s.metrics.IncrementCounter("tenant_creation_errors", metrics.Fields{
                "error_type": "subdomain_exists",
            })
            
            s.logger.WarnContext(ctx, "Subdomain already exists", logger.Fields{
                "subdomain": *req.Subdomain,
            })
            
            return nil, errors.ErrSubdomainAlreadyExists
        }
    }
    
    // 3. Create tenant entity with business rules
    tenant := &Tenant{
        ID:           uuid.New(),
        Name:         req.Name,
        Slug:         req.Slug,
        Email:        req.Email,
        Subdomain:    req.Subdomain,
        Status:       StatusActive,  // Business rule: new tenants start active
        Timezone:     getDefaultTimezone(req.Timezone),
        CurrencyCode: getDefaultCurrency(req.CurrencyCode),
        Metadata:     req.Metadata,
        Settings:     getDefaultSettings(),
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
    }
    
    span.SetAttributes(attribute.String("tenant.id", tenant.ID.String()))
    
    // 4. Persist to repository
    repositoryTimer := s.metrics.Timer("repository_operation_duration", metrics.Fields{
        "operation": "create_tenant",
    })
    
    if err := s.repo.Create(ctx, tenant); err != nil {
        repositoryTimer.Stop()
        
        span.RecordError(err)
        span.SetStatus(codes.Error, "Failed to create tenant")
        
        s.metrics.IncrementCounter("tenant_operations_total", metrics.Fields{
            "operation":  "create",
            "status":     "error",
            "error_type": "repository",
        })
        
        s.logger.ErrorContext(ctx, "Failed to persist tenant", logger.Fields{
            "tenant_id": tenant.ID.String(),
            "error":     err.Error(),
        })
        
        return nil, fmt.Errorf("failed to create tenant: %w", err)
    }
    duration := repositoryTimer.Stop()
    
    // 5. Cache the result if subdomain is provided
    if tenant.Subdomain != nil {
        cacheKey := fmt.Sprintf("tenant:subdomain:%s", *tenant.Subdomain)
        if err := s.cache.Set(ctx, cacheKey, tenant, 30*time.Minute); err != nil {
            // Log cache error but don't fail the operation
            s.logger.WarnContext(ctx, "Failed to cache tenant", logger.Fields{
                "tenant_id": tenant.ID.String(),
                "cache_key": cacheKey,
                "error":     err.Error(),
            })
        }
    }
    
    // 6. Success metrics and logging
    s.metrics.IncrementCounter("tenant_operations_total", metrics.Fields{
        "operation": "create",
        "status":    "success",
    })
    
    s.metrics.IncrementCounter("tenants_created_total", metrics.Fields{})
    
    s.metrics.ObserveHistogram("tenant_creation_duration", 
        duration.Seconds(), metrics.Fields{})
    
    span.SetStatus(codes.Ok, "Tenant created successfully")
    
    s.logger.InfoContext(ctx, "Tenant created successfully", logger.Fields{
        "tenant_id":   tenant.ID.String(),
        "tenant_name": tenant.Name,
        "duration_ms": duration.Milliseconds(),
    })
    
    return tenant, nil
}

func (s *service) GetTenant(ctx context.Context, id uuid.UUID) (*Tenant, error) {
    ctx, span := s.tracing.StartSpan(ctx, "service.get_tenant",
        tracing.WithSpanKind(tracing.SpanKindInternal))
    defer span.End()
    
    span.SetAttributes(attribute.String("tenant.id", id.String()))
    
    timer := s.metrics.Timer("tenant_operation_duration", metrics.Fields{
        "operation": "get",
    })
    defer timer.Stop()
    
    s.logger.DebugContext(ctx, "Getting tenant by ID", logger.Fields{
        "tenant_id": id.String(),
    })
    
    tenant, err := s.repo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, errors.ErrTenantNotFound) {
            span.SetAttributes(attribute.String("error.type", "not_found"))
            span.SetStatus(codes.Error, "Tenant not found")
            
            s.metrics.IncrementCounter("tenant_operations_total", metrics.Fields{
                "operation":  "get",
                "status":     "error",
                "error_type": "not_found",
            })
            
            s.logger.InfoContext(ctx, "Tenant not found", logger.Fields{
                "tenant_id": id.String(),
            })
            
            return nil, err
        }
        
        span.RecordError(err)
        span.SetStatus(codes.Error, "Failed to get tenant")
        
        s.metrics.IncrementCounter("tenant_operations_total", metrics.Fields{
            "operation":  "get",
            "status":     "error",
            "error_type": "repository",
        })
        
        s.logger.ErrorContext(ctx, "Failed to get tenant", logger.Fields{
            "tenant_id": id.String(),
            "error":     err.Error(),
        })
        
        return nil, fmt.Errorf("failed to get tenant: %w", err)
    }
    
    s.metrics.IncrementCounter("tenant_operations_total", metrics.Fields{
        "operation": "get",
        "status":    "success",
    })
    
    span.SetStatus(codes.Ok, "Tenant retrieved successfully")
    
    return tenant, nil
}

func (s *service) GetTenantBySubdomain(ctx context.Context, subdomain string) (*Tenant, error) {
    ctx, span := s.tracing.StartSpan(ctx, "service.get_tenant_by_subdomain",
        tracing.WithSpanKind(tracing.SpanKindInternal))
    defer span.End()
    
    span.SetAttributes(attribute.String("tenant.subdomain", subdomain))
    
    timer := s.metrics.Timer("tenant_operation_duration", metrics.Fields{
        "operation": "get_by_subdomain",
    })
    defer timer.Stop()
    
    // 1. Check cache first
    cacheKey := fmt.Sprintf("tenant:subdomain:%s", subdomain)
    var tenant Tenant
    
    cacheTimer := s.metrics.Timer("cache_operation_duration", metrics.Fields{
        "operation": "get",
        "cache_key": "tenant_subdomain",
    })
    
    if err := s.cache.Get(ctx, cacheKey, &tenant); err == nil {
        cacheTimer.Stop()
        
        s.metrics.IncrementCounter("cache_hits_total", metrics.Fields{
            "cache_key": "tenant_subdomain",
        })
        
        span.SetAttributes(attribute.Bool("cache.hit", true))
        span.SetStatus(codes.Ok, "Tenant retrieved from cache")
        
        s.logger.DebugContext(ctx, "Cache hit for tenant subdomain", logger.Fields{
            "subdomain": subdomain,
        })
        
        return &tenant, nil
    }
    cacheTimer.Stop()
    
    // 2. Cache miss - get from repository
    s.metrics.IncrementCounter("cache_misses_total", metrics.Fields{
        "cache_key": "tenant_subdomain",
    })
    
    span.SetAttributes(attribute.Bool("cache.hit", false))
    
    s.logger.DebugContext(ctx, "Cache miss for tenant subdomain", logger.Fields{
        "subdomain": subdomain,
    })
    
    dbTenant, err := s.repo.GetBySubdomain(ctx, subdomain)
    if err != nil {
        if errors.Is(err, errors.ErrTenantNotFound) {
            span.SetAttributes(attribute.String("error.type", "not_found"))
            span.SetStatus(codes.Error, "Tenant not found")
            
            s.logger.InfoContext(ctx, "Tenant not found by subdomain", logger.Fields{
                "subdomain": subdomain,
            })
            
            return nil, err
        }
        
        span.RecordError(err)
        span.SetStatus(codes.Error, "Failed to get tenant")
        
        s.logger.ErrorContext(ctx, "Failed to get tenant by subdomain", logger.Fields{
            "subdomain": subdomain,
            "error":     err.Error(),
        })
        
        return nil, fmt.Errorf("failed to get tenant by subdomain: %w", err)
    }
    
    // 3. Cache the result
    if err := s.cache.Set(ctx, cacheKey, dbTenant, 30*time.Minute); err != nil {
        s.logger.WarnContext(ctx, "Failed to cache tenant", logger.Fields{
            "subdomain": subdomain,
            "error":     err.Error(),
        })
    }
    
    span.SetStatus(codes.Ok, "Tenant retrieved and cached")
    
    s.logger.InfoContext(ctx, "Tenant retrieved and cached", logger.Fields{
        "tenant_id": dbTenant.ID.String(),
        "subdomain": subdomain,
    })
    
    return dbTenant, nil
}

// Business validation helper
func (s *service) validateCreateRequest(ctx context.Context, req CreateTenantRequest) error {
    if req.Name == "" {
        return errors.NewValidationError("name", "Name is required")
    }
    
    if len(req.Name) < 3 {
        return errors.NewValidationError("name", "Name must be at least 3 characters")
    }
    
    if req.Slug == "" {
        return errors.NewValidationError("slug", "Slug is required")
    }
    
    if req.Email == "" {
        return errors.NewValidationError("email", "Email is required")
    }
    
    // Email format validation
    if !isValidEmail(req.Email) {
        return errors.NewValidationError("email", "Invalid email format")
    }
    
    // Subdomain validation
    if req.Subdomain != nil {
        if len(*req.Subdomain) < 3 {
            return errors.NewValidationError("subdomain", "Subdomain must be at least 3 characters")
        }
        
        if !isValidSubdomain(*req.Subdomain) {
            return errors.NewValidationError("subdomain", "Invalid subdomain format")
        }
    }
    
    return nil
}

// Helper functions
func getDefaultTimezone(requested string) string {
    if requested != "" {
        return requested
    }
    return "UTC"
}

func getDefaultCurrency(requested string) string {
    if requested != "" {
        return requested
    }
    return "USD"
}

func getDefaultSettings() map[string]interface{} {
    return map[string]interface{}{
        "theme":              "default",
        "notifications":      true,
        "email_notifications": true,
        "language":           "en",
    }
}
```

## 🗄️ Repository Layer Implementation

### Repository Implementation with SQLC
```go
// internal/core/tenant/repository.go
package tenant

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
    
    "internal/platform/db"
    "internal/shared/logger"
    "internal/shared/tracing"
    "internal/shared/metrics"
    "internal/shared/errors"
)

type Repository interface {
    Create(ctx context.Context, tenant *Tenant) error
    GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error)
    GetBySubdomain(ctx context.Context, subdomain string) (*Tenant, error)
    List(ctx context.Context, filter ListFilter) ([]*Tenant, int, error)
    Update(ctx context.Context, id uuid.UUID, updates UpdateTenantRequest) error
    Delete(ctx context.Context, id uuid.UUID) error
    Exists(ctx context.Context, subdomain string) (bool, error)
}

type repository struct {
    store   db.Store
    logger  logger.Logger
    tracing tracing.Tracer
    metrics metrics.MetricsProvider
}

func NewRepository(
    store db.Store,
    logger logger.Logger,
    tracing tracing.Tracer,
    metrics metrics.MetricsProvider,
) Repository {
    return &repository{
        store:   store,
        logger:  logger,
        tracing: tracing,
        metrics: metrics,
    }
}

func (r *repository) Create(ctx context.Context, tenant *Tenant) error {
    // Start database span
    ctx, span := r.tracing.StartSpan(ctx, "repository.create_tenant",
        tracing.WithSpanKind(tracing.SpanKindClient),
        tracing.WithAttributes(
            attribute.String("db.system", "postgresql"),
            attribute.String("db.operation", "create"),
            attribute.String("db.table", "tenants"),
        ))
    defer span.End()
    
    // Add entity attributes
    span.SetAttributes(
        attribute.String("tenant.id", tenant.ID.String()),
        attribute.String("tenant.name", tenant.Name),
    )
    
    r.logger.DebugContext(ctx, "Creating tenant in database", logger.Fields{
        "tenant_id":   tenant.ID.String(),
        "tenant_name": tenant.Name,
        "operation":   "create",
    })
    
    // Convert domain model to SQLC parameters
    params, err := ToSQLCCreateParams(tenant)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "Failed to convert tenant")
        
        r.logger.ErrorContext(ctx, "Failed to convert tenant to SQLC params", logger.Fields{
            "tenant_id": tenant.ID.String(),
            "error":     err.Error(),
        })
        
        return fmt.Errorf("failed to convert tenant: %w", err)
    }
    
    // Execute database operation with metrics
    timer := r.metrics.Timer("database_operation_duration", metrics.Fields{
        "operation": "create_tenant",
        "table":     "tenants",
    })
    
    _, err = r.store.CreateTenant(ctx, params)
    duration := timer.Stop()
    
    if err != nil {
        // Database error metrics
        r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
            "operation": "create_tenant",
            "table":     "tenants",
            "status":    "error",
        })
        
        r.metrics.IncrementCounter("database_errors_total", metrics.Fields{
            "operation":  "create_tenant",
            "table":      "tenants",
            "error_type": classifyDBError(err),
        })
        
        // Record error in span
        span.RecordError(err)
        span.SetStatus(codes.Error, "Database operation failed")
        
        r.logger.ErrorContext(ctx, "Database operation failed", logger.Fields{
            "error":       err.Error(),
            "operation":   "create_tenant",
            "duration_ms": duration.Milliseconds(),
        })
        
        return fmt.Errorf("failed to create tenant: %w", err)
    }
    
    // Success metrics
    r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
        "operation": "create_tenant",
        "table":     "tenants",
        "status":    "success",
    })
    
    r.metrics.ObserveHistogram("database_operation_duration", 
        duration.Seconds(), metrics.Fields{
            "operation": "create_tenant",
            "table":     "tenants",
        })
    
    span.SetStatus(codes.Ok, "Tenant created successfully")
    
    r.logger.DebugContext(ctx, "Tenant created successfully in database", logger.Fields{
        "tenant_id":   tenant.ID.String(),
        "duration_ms": duration.Milliseconds(),
    })
    
    return nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error) {
    ctx, span := r.tracing.StartSpan(ctx, "repository.get_tenant_by_id",
        tracing.WithSpanKind(tracing.SpanKindClient),
        tracing.WithAttributes(
            attribute.String("db.system", "postgresql"),
            attribute.String("db.operation", "select"),
            attribute.String("db.table", "tenants"),
            attribute.String("tenant.id", id.String()),
        ))
    defer span.End()
    
    timer := r.metrics.Timer("database_operation_duration", metrics.Fields{
        "operation": "get_tenant_by_id",
        "table":     "tenants",
    })
    defer timer.Stop()
    
    r.logger.DebugContext(ctx, "Getting tenant by ID from database", logger.Fields{
        "tenant_id": id.String(),
    })
    
    sqlcTenant, err := r.store.GetTenantByID(ctx, id)
    if err != nil {
        if err.Error() == "no rows in result set" {
            span.SetAttributes(attribute.String("error.type", "not_found"))
            span.SetStatus(codes.Error, "Tenant not found")
            
            r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
                "operation":  "get_tenant_by_id",
                "table":      "tenants",
                "status":     "not_found",
            })
            
            r.logger.DebugContext(ctx, "Tenant not found in database", logger.Fields{
                "tenant_id": id.String(),
            })
            
            return nil, errors.ErrTenantNotFound
        }
        
        r.metrics.IncrementCounter("database_errors_total", metrics.Fields{
            "operation":  "get_tenant_by_id",
            "table":      "tenants",
            "error_type": classifyDBError(err),
        })
        
        span.RecordError(err)
        span.SetStatus(codes.Error, "Database operation failed")
        
        r.logger.ErrorContext(ctx, "Failed to get tenant from database", logger.Fields{
            "tenant_id": id.String(),
            "error":     err.Error(),
        })
        
        return nil, fmt.Errorf("failed to get tenant: %w", err)
    }
    
    // Convert SQLC model to domain model
    tenant, err := FromSQLCTenant(sqlcTenant)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "Failed to convert tenant")
        
        r.logger.ErrorContext(ctx, "Failed to convert SQLC tenant to domain model", logger.Fields{
            "tenant_id": id.String(),
            "error":     err.Error(),
        })
        
        return nil, fmt.Errorf("failed to convert tenant: %w", err)
    }
    
    r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
        "operation": "get_tenant_by_id",
        "table":     "tenants",
        "status":    "success",
    })
    
    span.SetStatus(codes.Ok, "Tenant retrieved successfully")
    
    return tenant, nil
}

func (r *repository) List(ctx context.Context, filter ListFilter) ([]*Tenant, int, error) {
    ctx, span := r.tracing.StartSpan(ctx, "repository.list_tenants",
        tracing.WithSpanKind(tracing.SpanKindClient),
        tracing.WithAttributes(
            attribute.String("db.system", "postgresql"),
            attribute.String("db.operation", "select"),
            attribute.String("db.table", "tenants"),
            attribute.Int("filter.page", filter.Page),
            attribute.Int("filter.limit", filter.Limit),
        ))
    defer span.End()
    
    if filter.Search != "" {
        span.SetAttributes(attribute.String("filter.search", filter.Search))
    }
    
    timer := r.metrics.Timer("database_operation_duration", metrics.Fields{
        "operation": "list_tenants",
        "table":     "tenants",
    })
    defer timer.Stop()
    
    r.logger.DebugContext(ctx, "Listing tenants from database", logger.Fields{
        "page":   filter.Page,
        "limit":  filter.Limit,
        "search": filter.Search,
    })
    
    // Get total count first
    var searchTerm *string
    if filter.Search != "" {
        searchTerm = &filter.Search
    }
    
    countTimer := r.metrics.Timer("database_operation_duration", metrics.Fields{
        "operation": "count_tenants",
        "table":     "tenants",
    })
    
    total, err := r.store.CountTenants(ctx, searchTerm)
    countTimer.Stop()
    
    if err != nil {
        r.metrics.IncrementCounter("database_errors_total", metrics.Fields{
            "operation":  "count_tenants",
            "table":      "tenants",
            "error_type": classifyDBError(err),
        })
        
        span.RecordError(err)
        span.SetStatus(codes.Error, "Failed to count tenants")
        
        return nil, 0, fmt.Errorf("failed to count tenants: %w", err)
    }
    
    span.SetAttributes(attribute.Int("total_count", int(total)))
    
    // Get paginated results
    offset := (filter.Page - 1) * filter.Limit
    sqlcTenants, err := r.store.ListTenants(ctx, db.ListTenantsParams{
        Search: searchTerm,
        Limit:  int32(filter.Limit),
        Offset: int32(offset),
    })
    
    if err != nil {
        r.metrics.IncrementCounter("database_errors_total", metrics.Fields{
            "operation":  "list_tenants",
            "table":      "tenants",
            "error_type": classifyDBError(err),
        })
        
        span.RecordError(err)
        span.SetStatus(codes.Error, "Failed to list tenants")
        
        return nil, 0, fmt.Errorf("failed to list tenants: %w", err)
    }
    
    // Convert to domain models
    tenants := make([]*Tenant, len(sqlcTenants))
    for i, sqlcTenant := range sqlcTenants {
        tenant, err := FromSQLCTenant(sqlcTenant)
        if err != nil {
            span.RecordError(err)
            span.SetStatus(codes.Error, "Failed to convert tenant")
            
            return nil, 0, fmt.Errorf("failed to convert tenant %d: %w", i, err)
        }
        tenants[i] = tenant
    }
    
    r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
        "operation": "list_tenants",
        "table":     "tenants",
        "status":    "success",
    })
    
    span.SetAttributes(attribute.Int("result_count", len(tenants)))
    span.SetStatus(codes.Ok, "Tenants listed successfully")
    
    return tenants, int(total), nil
}

// Helper function to classify database errors for metrics
func classifyDBError(err error) string {
    errStr := err.Error()
    switch {
    case strings.Contains(errStr, "duplicate key"):
        return "duplicate_key"
    case strings.Contains(errStr, "foreign key"):
        return "foreign_key"
    case strings.Contains(errStr, "connection"):
        return "connection"
    case strings.Contains(errStr, "timeout"):
        return "timeout"
    case strings.Contains(errStr, "no rows"):
        return "not_found"
    default:
        return "unknown"
    }
}
```

---

📚 **Next Steps**:
- [Best Practices](./best-practices.md) - Development guidelines and patterns
- [Error Handling](./error-handling.md) - Complete error handling strategies
- [Observability](./observability.md) - Deep dive into observability patterns
