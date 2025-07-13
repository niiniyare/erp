# Best Practices Guide

Essential development guidelines and patterns for building maintainable, scalable features in our ERP system.

## 🏗️ Architecture Best Practices

### 1. Dependency Direction Rules

#### ✅ **Correct Dependency Flow**
```go
// Higher layers can import lower layers
package handlers

import (
    "internal/core/tenant"        // ✅ API → Core
    "internal/shared/logger"      // ✅ API → Shared
)

// Core layer depends on interfaces, not implementations
package tenant

type Repository interface {
    Create(ctx context.Context, tenant *Tenant) error
}

type Service struct {
    repo Repository  // ✅ Depends on interface
}
```

#### ❌ **Wrong Dependency Direction**
```go
// DON'T: Lower layers importing higher layers
package tenant

import (
    "internal/api/handlers"  // ❌ Core → API (WRONG!)
)

// DON'T: Core depending on infrastructure implementations
package tenant

import (
    "internal/platform/db"  // ❌ Core → Infrastructure (WRONG!)
)

type Service struct {
    store db.Store  // ❌ Direct dependency on implementation
}
```

### 2. Layer Isolation

#### **API Layer Responsibilities**
```go
// ✅ API layer handles HTTP concerns only
func (h *TenantHandler) CreateTenant(c *gin.Context) {
    // ✅ Parse request
    var req CreateTenantRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    // ✅ Call service layer
    tenant, err := h.service.CreateTenant(c.Request.Context(), req)
    
    // ✅ Handle errors and format response
    if err != nil {
        handleError(c, err)
        return
    }
    
    c.JSON(201, tenant)
}

// ❌ DON'T put business logic in handlers
func (h *TenantHandler) CreateTenant(c *gin.Context) {
    // ❌ Business validation in API layer
    if len(req.Name) < 3 {
        c.JSON(400, gin.H{"error": "Name too short"})
        return
    }
    
    // ❌ Direct database access from API layer
    _, err := h.db.Exec("INSERT INTO tenants...")
}
```

#### **Service Layer Responsibilities**
```go
// ✅ Service layer handles business logic
func (s *service) CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error) {
    // ✅ Business validation
    if err := s.validateCreateRequest(req); err != nil {
        return nil, err
    }
    
    // ✅ Business rules
    tenant := &Tenant{
        ID:     uuid.New(),
        Status: StatusActive,  // Business rule
        // ... other business logic
    }
    
    // ✅ Use repository interface
    return s.repo.Create(ctx, tenant)
}

// ❌ DON'T put HTTP concerns in service
func (s *service) CreateTenant(ctx context.Context, req *http.Request) (*gin.Context, error) {
    // ❌ HTTP request/response handling in service
}

// ❌ DON'T put database queries in service
func (s *service) CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error) {
    // ❌ Direct SQL in service layer
    _, err := s.db.Exec("INSERT INTO tenants...")
}
```

## 🔄 Model Conversion Patterns

### 1. **Conversion at Boundaries**

#### Repository Layer Conversions
```go
// ✅ Convert domain models to database parameters
func (r *repository) Create(ctx context.Context, tenant *Tenant) error {
    // Convert domain → SQLC params
    params := db.CreateTenantParams{
        ID:           tenant.ID,
        Name:         tenant.Name,
        Status:       string(tenant.Status),  // Enum conversion
        Metadata:     marshalJSON(tenant.Metadata),  // JSON conversion
        CreatedAt:    tenant.CreatedAt,
    }
    
    _, err := r.store.CreateTenant(ctx, params)
    return err
}

// ✅ Convert database models to domain models
func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error) {
    sqlcTenant, err := r.store.GetTenantByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Convert SQLC → domain
    return FromSQLCTenant(sqlcTenant)
}
```

#### Conversion Helper Functions
```go
// ✅ Centralized conversion logic
func FromSQLCTenant(sqlcTenant db.Tenant) (*Tenant, error) {
    // Handle JSONB fields safely
    var metadata map[string]interface{}
    if len(sqlcTenant.Metadata) > 0 {
        if err := json.Unmarshal(sqlcTenant.Metadata, &metadata); err != nil {
            return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
        }
    }
    
    // Handle nullable fields
    var deletedAt *time.Time
    if sqlcTenant.DeletedAt.Valid {
        deletedAt = &sqlcTenant.DeletedAt.Time
    }
    
    return &Tenant{
        ID:        sqlcTenant.ID,
        Name:      sqlcTenant.Name,
        Status:    Status(sqlcTenant.Status),  // Type conversion
        Metadata:  metadata,
        DeletedAt: deletedAt,
        // ... other fields
    }, nil
}

func ToSQLCCreateParams(tenant *Tenant) (db.CreateTenantParams, error) {
    metadataBytes, err := json.Marshal(tenant.Metadata)
    if err != nil {
        return db.CreateTenantParams{}, fmt.Errorf("failed to marshal metadata: %w", err)
    }
    
    return db.CreateTenantParams{
        ID:       tenant.ID,
        Name:     tenant.Name,
        Status:   string(tenant.Status),
        Metadata: metadataBytes,
        // ... other fields
    }, nil
}
```

#### ❌ **Wrong Model Usage**
```go
// ❌ DON'T use database models in service layer
func (s *service) CreateTenant(req CreateTenantRequest) (*db.Tenant, error) {
    // ❌ Returning database model from service
}

// ❌ DON'T use domain models in SQLC calls
func (r *repository) Create(ctx context.Context, tenant *Tenant) error {
    // ❌ Passing domain model directly to SQLC
    _, err := r.store.CreateTenant(ctx, tenant)  // Won't compile!
}
```

## 🔒 Error Handling Patterns

### 1. **Domain Error Types**

#### Define Domain-Specific Errors
```go
// internal/shared/errors/errors.go
package errors

import "errors"

// Domain errors
var (
    ErrTenantNotFound          = errors.New("tenant not found")
    ErrSubdomainAlreadyExists  = errors.New("subdomain already exists")
    ErrInvalidTenantStatus     = errors.New("invalid tenant status")
    ErrTenantInactive          = errors.New("tenant is inactive")
)

// Validation errors with details
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
}

func NewValidationError(field, message string) error {
    return ValidationError{Field: field, Message: message}
}
```

### 2. **Error Wrapping Strategy**

#### Repository Layer Error Handling
```go
// ✅ Convert database errors to domain errors
func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error) {
    sqlcTenant, err := r.store.GetTenantByID(ctx, id)
    if err != nil {
        // Convert specific database errors to domain errors
        if err.Error() == "no rows in result set" {
            return nil, errors.ErrTenantNotFound  // Domain error
        }
        
        // Wrap infrastructure errors with context
        return nil, fmt.Errorf("failed to get tenant from database: %w", err)
    }
    
    return FromSQLCTenant(sqlcTenant)
}
```

#### Service Layer Error Handling
```go
// ✅ Let domain errors bubble up, wrap infrastructure errors
func (s *service) GetTenant(ctx context.Context, id uuid.UUID) (*Tenant, error) {
    tenant, err := s.repo.GetByID(ctx, id)
    if err != nil {
        // Let domain errors pass through
        if errors.Is(err, errors.ErrTenantNotFound) {
            return nil, err
        }
        
        // Wrap infrastructure errors with service context
        return nil, fmt.Errorf("service: failed to get tenant: %w", err)
    }
    
    return tenant, nil
}
```

#### API Layer Error Handling
```go
// ✅ Convert domain errors to appropriate HTTP responses
func (h *TenantHandler) GetTenant(c *gin.Context) {
    tenant, err := h.service.GetTenant(c.Request.Context(), id)
    if err != nil {
        switch {
        case errors.Is(err, errors.ErrTenantNotFound):
            c.JSON(http.StatusNotFound, gin.H{
                "error": "Tenant not found",
                "code":  "TENANT_NOT_FOUND",
            })
        case errors.Is(err, errors.ErrValidation):
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "Validation failed",
                "code":  "VALIDATION_ERROR",
            })
        default:
            h.logger.ErrorContext(ctx, "Internal server error", logger.Fields{
                "error": err.Error(),
            })
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "Internal server error",
                "code":  "INTERNAL_ERROR",
            })
        }
        return
    }
    
    c.JSON(http.StatusOK, tenant)
}
```

## 🚀 Caching Patterns

### 1. **Cache-Aside Pattern**

#### Service Layer Caching
```go
// ✅ Implement cache-aside pattern in service layer
func (s *service) GetTenantBySubdomain(ctx context.Context, subdomain string) (*Tenant, error) {
    // 1. Check cache first
    cacheKey := fmt.Sprintf("tenant:subdomain:%s", subdomain)
    var tenant Tenant
    
    if err := s.cache.Get(ctx, cacheKey, &tenant); err == nil {
        // Cache hit - log and return
        s.logger.DebugContext(ctx, "Cache hit for tenant subdomain", logger.Fields{
            "subdomain": subdomain,
        })
        return &tenant, nil
    }
    
    // 2. Cache miss - get from repository
    s.logger.DebugContext(ctx, "Cache miss for tenant subdomain", logger.Fields{
        "subdomain": subdomain,
    })
    
    dbTenant, err := s.repo.GetBySubdomain(ctx, subdomain)
    if err != nil {
        return nil, err
    }
    
    // 3. Store in cache for future requests
    if err := s.cache.Set(ctx, cacheKey, dbTenant, 30*time.Minute); err != nil {
        // Log cache error but don't fail the operation
        s.logger.WarnContext(ctx, "Failed to cache tenant", logger.Fields{
            "subdomain": subdomain,
            "error":     err.Error(),
        })
    }
    
    return dbTenant, nil
}
```

### 2. **Cache Invalidation**

#### Write-Through Cache Invalidation
```go
// ✅ Invalidate cache on updates
func (s *service) UpdateTenant(ctx context.Context, id uuid.UUID, req UpdateTenantRequest) (*Tenant, error) {
    // 1. Get existing tenant for cache invalidation
    existing, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // 2. Update in repository
    updated, err := s.repo.Update(ctx, id, req)
    if err != nil {
        return nil, err
    }
    
    // 3. Invalidate relevant cache entries
    s.invalidateTenantCache(ctx, existing)
    
    return updated, nil
}

func (s *service) invalidateTenantCache(ctx context.Context, tenant *Tenant) {
    // Invalidate ID-based cache
    idCacheKey := fmt.Sprintf("tenant:id:%s", tenant.ID.String())
    if err := s.cache.Delete(ctx, idCacheKey); err != nil {
        s.logger.WarnContext(ctx, "Failed to invalidate tenant ID cache", logger.Fields{
            "cache_key": idCacheKey,
            "error":     err.Error(),
        })
    }
    
    // Invalidate subdomain-based cache
    if tenant.Subdomain != nil {
        subdomainCacheKey := fmt.Sprintf("tenant:subdomain:%s", *tenant.Subdomain)
        if err := s.cache.Delete(ctx, subdomainCacheKey); err != nil {
            s.logger.WarnContext(ctx, "Failed to invalidate tenant subdomain cache", logger.Fields{
                "cache_key": subdomainCacheKey,
                "error":     err.Error(),
            })
        }
    }
}
```

### 3. **Cache Keys Strategy**

#### Consistent Cache Key Patterns
```go
// ✅ Use consistent, hierarchical cache key patterns
const (
    CacheKeyTenantByID        = "tenant:id:%s"
    CacheKeyTenantBySubdomain = "tenant:subdomain:%s"
    CacheKeyUsersByTenant     = "users:tenant:%s"
    CacheKeyTenantSettings    = "tenant:settings:%s"
)

// ✅ Cache key builder functions
func buildTenantCacheKey(id uuid.UUID) string {
    return fmt.Sprintf(CacheKeyTenantByID, id.String())
}

func buildSubdomainCacheKey(subdomain string) string {
    return fmt.Sprintf(CacheKeyTenantBySubdomain, subdomain)
}

// ✅ Cache TTL constants
const (
    CacheTTLTenantShort  = 15 * time.Minute  // Frequently changing data
    CacheTTLTenantMedium = 1 * time.Hour     // Moderately stable data
    CacheTTLTenantLong   = 24 * time.Hour    // Stable data
)
```

## 🔄 Transaction Patterns

### 1. **Service-Level Transactions**

#### Using Store.WithTx for Complex Operations
```go
// ✅ Service layer orchestrates transactions
func (s *service) CreateTenantWithUser(ctx context.Context, req CreateTenantWithUserRequest) (*TenantWithUser, error) {
    var result *TenantWithUser
    
    // Use repository's transaction method
    err := s.tenantRepo.WithTransaction(ctx, func(ctx context.Context, repos TransactionalRepositories) error {
        // 1. Create tenant
        tenant := &Tenant{
            ID:   uuid.New(),
            Name: req.TenantName,
            // ... other fields
        }
        
        if err := repos.TenantRepo.Create(ctx, tenant); err != nil {
            return fmt.Errorf("failed to create tenant: %w", err)
        }
        
        // 2. Create admin user
        user := &User{
            ID:       uuid.New(),
            TenantID: tenant.ID,
            Email:    req.AdminEmail,
            Role:     RoleAdmin,
            // ... other fields
        }
        
        if err := repos.UserRepo.Create(ctx, user); err != nil {
            return fmt.Errorf("failed to create admin user: %w", err)
        }
        
        // 3. Create default settings
        settings := &TenantSettings{
            TenantID: tenant.ID,
            Theme:    "default",
            // ... other settings
        }
        
        if err := repos.SettingsRepo.Create(ctx, settings); err != nil {
            return fmt.Errorf("failed to create tenant settings: %w", err)
        }
        
        result = &TenantWithUser{
            Tenant: tenant,
            User:   user,
        }
        
        return nil  // Commit transaction
    })
    
    if err != nil {
        return nil, fmt.Errorf("failed to create tenant with user: %w", err)
    }
    
    return result, nil
}
```

#### Repository Transaction Support
```go
// ✅ Repository supports transactions
type TransactionalRepositories struct {
    TenantRepo   Repository
    UserRepo     user.Repository
    SettingsRepo settings.Repository
}

func (r *repository) WithTransaction(ctx context.Context, fn func(context.Context, TransactionalRepositories) error) error {
    return r.store.WithTx(ctx, func(ctx context.Context, tx db.Store) error {
        repos := TransactionalRepositories{
            TenantRepo:   NewRepository(tx, r.logger, r.tracing, r.metrics),
            UserRepo:     user.NewRepository(tx, r.logger, r.tracing, r.metrics),
            SettingsRepo: settings.NewRepository(tx, r.logger, r.tracing, r.metrics),
        }
        
        return fn(ctx, repos)
    })
}
```

## 📊 Observability Best Practices

### 1. **Structured Logging**

#### Consistent Field Names
```go
// ✅ Use standardized field names across the application
const (
    LogFieldTenantID   = "tenant_id"
    LogFieldUserID     = "user_id"
    LogFieldOperation  = "operation"
    LogFieldDuration   = "duration_ms"
    LogFieldErrorType  = "error_type"
    LogFieldRequestID  = "request_id"
)

// ✅ Consistent logging patterns
func (s *service) CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error) {
    logger.InfoContext(ctx, "Starting tenant creation", logger.Fields{
        LogFieldOperation:  "create_tenant",
        LogFieldTenantName: req.Name,
    })
    
    // ... business logic
    
    logger.InfoContext(ctx, "Tenant created successfully", logger.Fields{
        LogFieldOperation: "create_tenant",
        LogFieldTenantID:  tenant.ID.String(),
        LogFieldDuration:  duration.Milliseconds(),
    })
    
    return tenant, nil
}
```

### 2. **Distributed Tracing**

#### Span Naming Conventions
```go
// ✅ Use consistent span naming patterns
const (
    SpanHTTPPrefix       = "http."
    SpanServicePrefix    = "service."
    SpanRepositoryPrefix = "repository."
    SpanDatabasePrefix   = "db."
)

// ✅ Comprehensive span attributes
func (h *TenantHandler) CreateTenant(c *gin.Context) {
    ctx, span := h.tracing.StartSpan(ctx, "http.create_tenant",
        tracing.WithSpanKind(tracing.SpanKindServer),
        tracing.WithAttributes(
            attribute.String("http.method", c.Request.Method),
            attribute.String("http.route", "/api/v1/tenants"),
            attribute.String("http.url", c.Request.URL.String()),
            attribute.String("user_agent", c.Request.UserAgent()),
        ))
    defer span.End()
    
    // Add business context as operation progresses
    span.SetAttributes(
        attribute.String("tenant.name", req.Name),
        attribute.String("tenant.slug", req.Slug),
    )
    
    // ... handle request
}
```

### 3. **Metrics Collection**

#### Business and Technical Metrics
```go
// ✅ Collect both business and technical metrics
func (s *service) CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error) {
    // Technical metrics
    timer := s.metrics.Timer("tenant_operation_duration", metrics.Fields{
        "operation": "create",
    })
    defer timer.Stop()
    
    // Business logic...
    
    if err != nil {
        // Error metrics with classification
        s.metrics.IncrementCounter("tenant_operations_total", metrics.Fields{
            "operation":  "create",
            "status":     "error",
            "error_type": classifyError(err),
        })
        return nil, err
    }
    
    // Success metrics
    s.metrics.IncrementCounter("tenant_operations_total", metrics.Fields{
        "operation": "create",
        "status":    "success",
    })
    
    // Business metrics
    s.metrics.IncrementCounter("tenants_created_total", metrics.Fields{})
    s.metrics.SetGauge("active_tenants_count", float64(activeCount), metrics.Fields{})
    
    return tenant, nil
}
```

## 🧪 Testing Patterns

### 1. **Repository Testing with Test Database**

#### Repository Integration Tests
```go
// ✅ Test repository with real database
func TestRepository_Create(t *testing.T) {
    // Setup test database
    testDB := setupTestDB(t)
    defer testDB.Close()
    
    store := db.NewStore(testDB)
    repo := NewRepository(store, logger.NewTestLogger(), tracing.NewNoopTracer(), metrics.NewNoopMetrics())
    
    // Test data
    tenant := &Tenant{
        ID:           uuid.New(),
        Name:         "Test Tenant",
        Slug:         "test-tenant",
        Email:        "test@example.com",
        Status:       StatusActive,
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
    }
    
    // Execute and verify
    err := repo.Create(context.Background(), tenant)
    require.NoError(t, err)
    
    // Verify persistence
    found, err := repo.GetByID(context.Background(), tenant.ID)
    require.NoError(t, err)
    assert.Equal(t, tenant.Name, found.Name)
    assert.Equal(t, tenant.Status, found.Status)
}
```

### 2. **Service Testing with Mocks**

#### Service Unit Tests
```go
// ✅ Test service layer with mocked dependencies
func TestService_CreateTenant(t *testing.T) {
    mockRepo := new(mocks.MockRepository)
    mockCache := new(mocks.MockCache)
    service := NewService(mockRepo, mockCache, logger.NewTestLogger(), tracing.NewNoopTracer(), metrics.NewNoopMetrics())
    
    // Setup expectations
    mockRepo.On("Exists", mock.Anything, "test-subdomain").Return(false, nil)
    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*tenant.Tenant")).Return(nil)
    mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
    
    // Test request
    req := CreateTenantRequest{
        Name:      "Test Tenant",
        Slug:      "test-tenant",
        Email:     "test@example.com",
        Subdomain: stringPtr("test-subdomain"),
    }
    
    // Execute
    tenant, err := service.CreateTenant(context.Background(), req)
    
    // Verify
    require.NoError(t, err)
    assert.Equal(t, req.Name, tenant.Name)
    assert.Equal(t, StatusActive, tenant.Status)
    
    mockRepo.AssertExpectations(t)
    mockCache.AssertExpectations(t)
}

func stringPtr(s string) *string {
    return &s
}
```

### 3. **Handler Testing**

#### HTTP Handler Tests
```go
// ✅ Test handlers with mock service
func TestTenantHandler_CreateTenant(t *testing.T) {
    mockService := new(mocks.MockService)
    handler := NewTenantHandler(mockService, logger.NewTestLogger(), tracing.NewNoopTracer(), metrics.NewNoopMetrics())
    
    // Setup router
    router := gin.New()
    router.POST("/tenants", handler.CreateTenant)
    
    // Expected response
    expectedTenant := &Tenant{
        ID:   uuid.New(),
        Name: "Test Tenant",
        Slug: "test-tenant",
    }
    
    mockService.On("CreateTenant", mock.Anything, mock.AnythingOfType("tenant.CreateTenantRequest")).
        Return(expectedTenant, nil)
    
    // Test request
    requestBody := `{
        "name": "Test Tenant",
        "slug": "test-tenant",
        "email": "test@example.com"
    }`
    
    req, _ := http.NewRequest("POST", "/tenants", strings.NewReader(requestBody))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    // Verify response
    assert.Equal(t, http.StatusCreated, w.Code)
    
    var response Tenant
    err := json.Unmarshal(w.Body.Bytes(), &response)
    require.NoError(t, err)
    assert.Equal(t, expectedTenant.Name, response.Name)
    
    mockService.AssertExpectations(t)
}
```

## 🔐 Security Best Practices

### 1. **Input Validation**

#### API Layer Validation
```go
// ✅ Validate and sanitize input at API boundary
type CreateTenantRequest struct {
    Name      string                 `json:"name" binding:"required,min=3,max=100"`
    Slug      string                 `json:"slug" binding:"required,min=3,max=50,alphanum"`
    Email     string                 `json:"email" binding:"required,email"`
    Subdomain *string                `json:"subdomain,omitempty" binding:"omitempty,min=3,max=63,alphanum"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ✅ Additional validation in service layer
func (s *service) validateCreateRequest(req CreateTenantRequest) error {
    // Business-specific validation
    if req.Subdomain != nil {
        if !isValidSubdomain(*req.Subdomain) {
            return errors.NewValidationError("subdomain", "Invalid subdomain format")
        }
        
        if isReservedSubdomain(*req.Subdomain) {
            return errors.NewValidationError("subdomain", "Subdomain is reserved")
        }
    }
    
    return nil
}
```

### 2. **SQL Injection Prevention**

#### Always Use Parameterized Queries
```go
// ✅ SQLC generates safe parameterized queries
-- name: GetTenantBySubdomain :one
SELECT * FROM tenants 
WHERE subdomain = $1 AND deleted_at IS NULL;

// Generated safe code:
func (q *Queries) GetTenantBySubdomain(ctx context.Context, subdomain string) (Tenant, error) {
    // Parameterized query - safe from SQL injection
}

// ❌ DON'T build dynamic SQL strings
func (r *repository) GetBySubdomain(ctx context.Context, subdomain string) (*Tenant, error) {
    // ❌ Vulnerable to SQL injection
    query := fmt.Sprintf("SELECT * FROM tenants WHERE subdomain = '%s'", subdomain)
    // ... execute query
}
```

## 📈 Performance Best Practices

### 1. **Database Query Optimization**

#### Efficient Query Patterns
```go
// ✅ Use appropriate indexes
CREATE INDEX CONCURRENTLY idx_tenants_subdomain 
ON tenants(subdomain) 
WHERE deleted_at IS NULL;

// ✅ Select only needed fields
-- name: GetTenantBasicInfo :one
SELECT id, name, slug, status FROM tenants
WHERE id = $1 AND deleted_at IS NULL;

// ✅ Use proper pagination
-- name: ListTenants :many
SELECT * FROM tenants
WHERE deleted_at IS NULL
  AND ($1::text IS NULL OR name ILIKE '%' || $1 || '%')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
```

### 2. **Caching Strategy**

#### Multi-Level Caching
```go
// ✅ Cache at appropriate levels
func (s *service) GetTenant(ctx context.Context, id uuid.UUID) (*Tenant, error) {
    // L1: In-memory cache (short TTL)
    if tenant := s.memoryCache.Get(id); tenant != nil {
        return tenant, nil
    }
    
    // L2: Redis cache (longer TTL)
    cacheKey := buildTenantCacheKey(id)
    var tenant Tenant
    if err := s.cache.Get(ctx, cacheKey, &tenant); err == nil {
        s.memoryCache.Set(id, &tenant, 5*time.Minute)
        return &tenant, nil
    }
    
    // L3: Database
    dbTenant, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Cache for future requests
    s.cache.Set(ctx, cacheKey, dbTenant, 30*time.Minute)
    s.memoryCache.Set(id, dbTenant, 5*time.Minute)
    
    return dbTenant, nil
}
```

## 🚨 Common Anti-Patterns to Avoid

### 1. **Architecture Violations**
```go
// ❌ DON'T: Service importing handler
package tenant
import "internal/api/handlers"  // WRONG!

// ❌ DON'T: Handler containing business logic
func (h *Handler) CreateTenant(c *gin.Context) {
    if len(req.Name) < 3 {  // Business logic in handler
        // WRONG!
    }
}

// ❌ DON'T: Repository containing business logic
func (r *Repository) Create(ctx context.Context, tenant *Tenant) error {
    if tenant.Status == StatusInactive {  // Business logic in repository
        return errors.New("cannot create inactive tenant")  // WRONG!
    }
}
```

### 2. **Error Handling Anti-Patterns**
```go
// ❌ DON'T: Swallow errors
func (s *service) CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error) {
    _, err := s.repo.Create(ctx, tenant)
    if err != nil {
        // Silently ignoring error - WRONG!
        return nil, nil
    }
}

// ❌ DON'T: Generic error messages
func (h *Handler) CreateTenant(c *gin.Context) {
    _, err := h.service.CreateTenant(ctx, req)
    if err != nil {
        c.JSON(500, gin.H{"error": "Something went wrong"})  // Not helpful - WRONG!
    }
}
```

### 3. **Performance Anti-Patterns**
```go
// ❌ DON'T: N+1 query problems
func (s *service) GetTenantsWithUsers(ctx context.Context) ([]*TenantWithUsers, error) {
    tenants, err := s.repo.List(ctx, filter)
    if err != nil {
        return nil, err
    }
    
    for _, tenant := range tenants {
        // N+1 problem - one query per tenant
        users, _ := s.userRepo.GetByTenantID(ctx, tenant.ID)  // WRONG!
        tenant.Users = users
    }
}

// ✅ DO: Use proper joins or batch queries
func (s *service) GetTenantsWithUsers(ctx context.Context) ([]*TenantWithUsers, error) {
    return s.repo.GetTenantsWithUsers(ctx, filter)  // Single query with JOIN
}
```

---

📚 **Next Steps**:
- [Error Handling](./error-handling.md) - Comprehensive error handling strategies
- [Code Examples](./code-examples.md) - See these patterns in action
- [Architecture Overview](./architecture.md) - Review architectural principles
