# Code Examples Guide

Practical implementation examples showing how to build features following our Clean Architecture patterns with full observability integration.

##  Complete Feature Implementation

Let's walk through implementing a complete tenant management feature from API to database.

##  API Layer Implementation

### Goa Handler Example
```go
// internal/api/handlers/tenant_goa.go
package handlers

import (
    "context"
    "strconv"

    "github.com/google/uuid"
    goaTenant "awo.so/internal/api/gen/tenant"
    
    "awo.so/internal/core/tenant"
    "awo.so/internal/shared/logger"
    "awo.so/internal/shared/tracing"
    "awo.so/internal/shared/metrics"
    "awo.so/internal/shared/errors"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
)

// TenantGoaHandler implements the Goa tenant service interface
type TenantGoaHandler struct {
    service tenant.Service
    tracing tracing.TracingService
    metrics *metrics.MetricsService
    logger  logger.Logger
}

// NewTenantGoaHandler creates a new Goa tenant handler
func NewTenantGoaHandler(
    service tenant.Service,
    tracing tracing.TracingService,
    metrics *metrics.MetricsService,
) goaTenant.Service {
    return &TenantGoaHandler{
        service: service,
        tracing: tracing,
        metrics: metrics,
        logger:  logger.WithFields(logger.Fields{"handler": "tenant"}),
    }
}

// CreateTenant implements the Goa service interface
func (h *TenantGoaHandler) CreateTenant(ctx context.Context, p *goaTenant.CreateTenantPayload) (*goaTenant.TenantResult, error) {
    // Start tracing span with comprehensive attributes
    ctx, span := h.tracing.StartSpan(ctx, "tenant_handler.create_tenant",
        tracing.WithSpanKind(tracing.SpanKindServer),
        tracing.WithAttributes(
            attribute.String("tenant.name", p.Name),
            attribute.String("tenant.slug", p.Slug),
            attribute.String("tenant.email", p.Email),
        ))
    defer span.End()

    // Track metrics for handler operations
    timer := h.metrics.Timer("tenant_handler_duration", metrics.Fields{
        "operation": "create_tenant",
        "handler":   "goa",
    })
    defer timer.Stop()

    h.metrics.IncrementGauge("active_tenant_requests", metrics.Fields{
        "operation": "create",
    })
    defer h.metrics.DecrementGauge("active_tenant_requests", metrics.Fields{
        "operation": "create",
    })

    // Log request processing
    h.logger.InfoContext(ctx, "Processing create tenant request", logger.Fields{
        "tenant_name":  p.Name,
        "tenant_slug":  p.Slug,
        "tenant_email": p.Email,
        "operation":    "create_tenant",
    })

    // Add subdomain to span if provided
    if p.Subdomain != nil {
        span.SetAttributes(attribute.String("tenant.subdomain", *p.Subdomain))
        h.logger.DebugContext(ctx, "Subdomain provided", logger.Fields{
            "subdomain": *p.Subdomain,
        })
    }

    // Convert Goa payload to domain request
    domainReq := tenant.CreateTenantRequest{
        Name:         p.Name,
        Slug:         p.Slug,
        Email:        p.Email,
        Subdomain:    p.Subdomain,
        Timezone:     p.Timezone,
        CurrencyCode: p.CurrencyCode,
        Metadata:     p.Metadata,
    }

    // Call service layer with proper error handling
    newTenant, err := h.service.CreateTenant(ctx, domainReq)
    if err != nil {
        return nil, h.handleError(ctx, err, span)
    }

    // Convert domain model to Goa result
    result := &goaTenant.TenantResult{
        ID:           newTenant.ID.String(),
        Name:         newTenant.Name,
        Slug:         newTenant.Slug,
        Email:        newTenant.Email,
        Status:       string(newTenant.Status),
        Timezone:     newTenant.Timezone,
        CurrencyCode: newTenant.CurrencyCode,
        CreatedAt:    newTenant.CreatedAt.Format(time.RFC3339),
        UpdatedAt:    newTenant.UpdatedAt.Format(time.RFC3339),
    }

    if newTenant.Subdomain != nil {
        result.Subdomain = newTenant.Subdomain
    }

    // Success metrics and logging
    span.SetAttributes(
        attribute.String("tenant.id", newTenant.ID.String()),
        attribute.String("response.status", "created"),
    )
    span.SetStatus(codes.Ok, "Tenant created successfully")

    h.metrics.IncrementCounter("tenant_operations_total", metrics.Fields{
        "operation": "create",
        "status":    "success",
        "handler":   "goa",
    })

    h.logger.InfoContext(ctx, "Tenant created successfully", logger.Fields{
        "tenant_id":   newTenant.ID.String(),
        "tenant_name": newTenant.Name,
        "operation":   "create_tenant",
    })

    return result, nil
}

// GetTenant implements the Goa service interface  
func (h *TenantGoaHandler) GetTenant(ctx context.Context, p *goaTenant.GetTenantPayload) (*goaTenant.TenantResult, error) {
    ctx, span := h.tracing.StartSpan(ctx, "tenant_handler.get_tenant",
        tracing.WithSpanKind(tracing.SpanKindServer))
    defer span.End()

    timer := h.metrics.Timer("tenant_handler_duration", metrics.Fields{
        "operation": "get_tenant",
        "handler":   "goa",
    })
    defer timer.Stop()

    // Parse and validate ID
    id, err := uuid.Parse(p.ID)
    if err != nil {
        span.SetAttributes(attribute.String("error.type", "validation_error"))
        span.SetStatus(codes.Error, "Invalid UUID format")
        
        h.logger.WarnContext(ctx, "Invalid tenant ID format", logger.Fields{
            "tenant_id": p.ID,
            "error":     err.Error(),
        })
        
        return nil, goaTenant.MakeBadRequest(errors.NewValidationError("id", "Invalid UUID format"))
    }

    span.SetAttributes(attribute.String("tenant.id", id.String()))

    h.logger.DebugContext(ctx, "Getting tenant by ID", logger.Fields{
        "tenant_id": id.String(),
        "operation": "get_tenant",
    })

    // Get from service layer
    domainTenant, err := h.service.GetTenant(ctx, id)
    if err != nil {
        return nil, h.handleError(ctx, err, span)
    }

    // Convert to Goa result
    result := &goaTenant.TenantResult{
        ID:           domainTenant.ID.String(),
        Name:         domainTenant.Name,
        Slug:         domainTenant.Slug,
        Email:        domainTenant.Email,
        Status:       string(domainTenant.Status),
        Timezone:     domainTenant.Timezone,
        CurrencyCode: domainTenant.CurrencyCode,
        CreatedAt:    domainTenant.CreatedAt.Format(time.RFC3339),
        UpdatedAt:    domainTenant.UpdatedAt.Format(time.RFC3339),
    }

    if domainTenant.Subdomain != nil {
        result.Subdomain = domainTenant.Subdomain
    }

    span.SetStatus(codes.Ok, "Tenant retrieved successfully")
    
    h.metrics.IncrementCounter("tenant_operations_total", metrics.Fields{
        "operation": "get",
        "status":    "success",
        "handler":   "goa",
    })

    return result, nil
}

// ListTenants implements the Goa service interface
func (h *TenantGoaHandler) ListTenants(ctx context.Context, p *goaTenant.ListTenantsPayload) (*goaTenant.TenantListResult, error) {
    ctx, span := h.tracing.StartSpan(ctx, "tenant_handler.list_tenants",
        tracing.WithSpanKind(tracing.SpanKindServer))
    defer span.End()

    timer := h.metrics.Timer("tenant_handler_duration", metrics.Fields{
        "operation": "list_tenants", 
        "handler":   "goa",
    })
    defer timer.Stop()

    // Parse pagination parameters with defaults
    page := 1
    limit := 20
    
    if p.Page != nil && *p.Page > 0 {
        page = *p.Page
    }
    
    if p.Limit != nil && *p.Limit > 0 && *p.Limit <= 100 {
        limit = *p.Limit
    }

    // Build filter
    filter := tenant.ListFilter{
        Page:  page,
        Limit: limit,
    }
    
    if p.Search != nil {
        filter.Search = *p.Search
        span.SetAttributes(attribute.String("filter.search", *p.Search))
    }
    
    if p.Status != nil {
        filter.Status = *p.Status
        span.SetAttributes(attribute.String("filter.status", *p.Status))
    }

    span.SetAttributes(
        attribute.Int("filter.page", page),
        attribute.Int("filter.limit", limit),
    )

    h.logger.DebugContext(ctx, "Listing tenants", logger.Fields{
        "page":      page,
        "limit":     limit,
        "search":    filter.Search,
        "status":    filter.Status,
        "operation": "list_tenants",
    })

    // Get from service layer
    tenants, total, err := h.service.ListTenants(ctx, filter)
    if err != nil {
        return nil, h.handleError(ctx, err, span)
    }

    // Convert to Goa results
    results := make([]*goaTenant.TenantResult, len(tenants))
    for i, domainTenant := range tenants {
        results[i] = &goaTenant.TenantResult{
            ID:           domainTenant.ID.String(),
            Name:         domainTenant.Name,
            Slug:         domainTenant.Slug,
            Email:        domainTenant.Email,
            Status:       string(domainTenant.Status),
            Timezone:     domainTenant.Timezone,
            CurrencyCode: domainTenant.CurrencyCode,
            CreatedAt:    domainTenant.CreatedAt.Format(time.RFC3339),
            UpdatedAt:    domainTenant.UpdatedAt.Format(time.RFC3339),
        }
        
        if domainTenant.Subdomain != nil {
            results[i].Subdomain = domainTenant.Subdomain
        }
    }

    result := &goaTenant.TenantListResult{
        Tenants:     results,
        Total:       total,
        Page:        page,
        Limit:       limit,
        HasNext:     (page * limit) < total,
        HasPrevious: page > 1,
    }

    span.SetAttributes(
        attribute.Int("response.total", total),
        attribute.Int("response.count", len(results)),
        attribute.Bool("response.has_next", result.HasNext),
        attribute.Bool("response.has_previous", result.HasPrevious),
    )
    span.SetStatus(codes.Ok, "Tenants listed successfully")

    h.metrics.IncrementCounter("tenant_operations_total", metrics.Fields{
        "operation": "list",
        "status":    "success", 
        "handler":   "goa",
    })

    return result, nil
}

// handleError converts domain errors to appropriate Goa errors
func (h *TenantGoaHandler) handleError(ctx context.Context, err error, span tracing.Span) error {
    var businessErr *errors.BusinessError
    if errors.As(err, &businessErr) {
        // Handle specific business error codes
        switch businessErr.Code {
        case "TENANT_NOT_FOUND":
            span.SetAttributes(attribute.String("error.type", "not_found"))
            span.SetStatus(codes.Error, "Tenant not found")
            
            h.metrics.IncrementCounter("tenant_operations_total", metrics.Fields{
                "operation":  "unknown",
                "status":     "error",
                "error_type": "not_found",
                "handler":    "goa",
            })
            
            h.logger.InfoContext(ctx, "Tenant not found", logger.Fields{
                "error": businessErr.Message,
            })
            
            return goaTenant.MakeNotFound(businessErr)

        case "SUBDOMAIN_EXISTS", "TENANT_EXISTS":
            span.SetAttributes(attribute.String("error.type", "conflict"))
            span.SetStatus(codes.Error, "Tenant conflict")
            
            h.metrics.IncrementCounter("tenant_operations_total", metrics.Fields{
                "operation":  "unknown",
                "status":     "error", 
                "error_type": "conflict",
                "handler":    "goa",
            })
            
            h.logger.WarnContext(ctx, "Tenant conflict", logger.Fields{
                "error": businessErr.Message,
                "code":  businessErr.Code,
            })
            
            return goaTenant.MakeConflict(businessErr)

        case "VALIDATION_ERROR", "INVALID_INPUT":
            span.SetAttributes(attribute.String("error.type", "validation"))
            span.SetStatus(codes.Error, "Validation failed")
            
            h.metrics.IncrementCounter("tenant_operations_total", metrics.Fields{
                "operation":  "unknown",
                "status":     "error",
                "error_type": "validation", 
                "handler":    "goa",
            })
            
            h.logger.WarnContext(ctx, "Validation failed", logger.Fields{
                "error": businessErr.Message,
                "code":  businessErr.Code,
            })
            
            return goaTenant.MakeBadRequest(businessErr)

        default:
            // Fallback to HTTP status mapping
            return h.mapHTTPStatusToGoaError(businessErr, span)
        }
    }

    // Handle non-business errors
    span.RecordError(err)
    span.SetStatus(codes.Error, "Internal server error")
    
    h.metrics.IncrementCounter("tenant_operations_total", metrics.Fields{
        "operation":  "unknown",
        "status":     "error",
        "error_type": "internal",
        "handler":    "goa",
    })
    
    h.logger.ErrorContext(ctx, "Internal server error", logger.Fields{
        "error": err.Error(),
    })
    
    return goaTenant.MakeInternalServerError(err)
}

// mapHTTPStatusToGoaError maps BusinessError HTTP status codes to Goa errors
func (h *TenantGoaHandler) mapHTTPStatusToGoaError(businessErr *errors.BusinessError, span tracing.Span) error {
    switch businessErr.HTTPStatus {
    case 400:
        span.SetAttributes(attribute.String("error.type", "bad_request"))
        return goaTenant.MakeBadRequest(businessErr)
    case 401:
        span.SetAttributes(attribute.String("error.type", "unauthorized"))  
        return goaTenant.MakeUnauthorized(businessErr)
    case 403:
        span.SetAttributes(attribute.String("error.type", "forbidden"))
        return goaTenant.MakeForbidden(businessErr)
    case 404:
        span.SetAttributes(attribute.String("error.type", "not_found"))
        return goaTenant.MakeNotFound(businessErr)
    case 409:
        span.SetAttributes(attribute.String("error.type", "conflict"))
        return goaTenant.MakeConflict(businessErr)
    case 422:
        span.SetAttributes(attribute.String("error.type", "unprocessable_entity"))
        return goaTenant.MakeUnprocessableEntity(businessErr)
    case 500:
        span.SetAttributes(attribute.String("error.type", "internal_server_error"))
        return goaTenant.MakeInternalServerError(businessErr)
    default:
        span.SetAttributes(attribute.String("error.type", "unknown"))
        return goaTenant.MakeBadRequest(businessErr)
    }
}
```

##  Service Layer Implementation

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

## ️ Repository Layer Implementation

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

 **Next Steps**:
- [Best Practices](./01-best-practices.md) - Development guidelines and patterns
- [Error Handling](./error-handling.md) - Complete error handling strategies
- [Observability](./observability.md) - Deep dive into observability patterns
