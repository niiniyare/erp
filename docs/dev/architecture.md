# Architecture Overview

Our ERP system follows a **Clean Architecture** pattern with clear separation of concerns and dependency inversion.

## 🏗️ System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    API Layer                                │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   Handlers  │  │   Routers   │  │ Middleware  │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                  Core Business Layer                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   Services  │  │   Models    │  │ Interfaces  │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                 Repository Layer                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │ Repositories│  │ Converters  │  │ SQLC Store  │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                Infrastructure Layer                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │  Database   │  │    Cache    │  │  External   │        │
│  │   (pgx)     │  │   (Redis)   │  │  Services   │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
```

## 📁 Directory Structure

```
internal/
├── api/                    # API Layer
│   ├── handlers/          # HTTP request handlers
│   ├── middleware/        # HTTP middleware
│   └── routes/           # Route definitions
├── core/                 # Core Business Layer
│   ├── tenant/           # Domain module
│   │   ├── model.go      # Domain models
│   │   ├── service.go    # Business logic
│   │   └── repository.go # Repository interface
│   └── shared/           # Shared domain logic
├── platform/             # Infrastructure Layer
│   ├── database/         # Database connections
│   ├── cache/           # Cache implementations
│   └── config/          # Configuration
└── shared/              # Cross-cutting concerns
    ├── logger/          # Logging
    ├── tracing/         # Distributed tracing
    └── metrics/         # Metrics collection
```

## 🎯 Layer Responsibilities

### 1. API Layer (`/internal/api/`)

**Purpose**: Handle HTTP requests and responses

#### Components:
- **Handlers**: Process HTTP requests, validate input, format responses
- **Middleware**: Cross-cutting concerns (auth, logging, rate limiting)
- **Routers**: Define routes and route grouping

#### Responsibilities:
- ✅ HTTP request/response handling
- ✅ Input validation and sanitization
- ✅ Response formatting (JSON, status codes)
- ✅ Authentication and authorization
- ✅ Request context extraction (tenant, user)
- ❌ Business logic implementation
- ❌ Direct database access

#### Example Structure:
```go
// Handlers focus on HTTP concerns only
func (h *TenantHandler) CreateTenant(c *gin.Context) {
    var req CreateTenantRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    tenant, err := h.service.CreateTenant(c.Request.Context(), req)
    // Handle business errors and format response
}
```

### 2. Core Business Layer (`/internal/core/`)

**Purpose**: Implement business logic and domain rules

#### Components:
- **Services**: Orchestrate business operations, handle complex logic
- **Models**: Domain entities, value objects, business data structures
- **Interfaces**: Define contracts for external dependencies

#### Responsibilities:
- ✅ Business rule implementation
- ✅ Domain entity management
- ✅ Service orchestration
- ✅ Caching strategies
- ✅ External service integration
- ❌ HTTP concerns
- ❌ Database-specific logic

#### Key Principles:
- **Domain-Driven Design**: Entities reflect business concepts
- **Dependency Inversion**: Depend on interfaces, not implementations
- **Single Responsibility**: Each service handles one business domain

#### Example Structure:
```go
// Services implement business logic
func (s *TenantService) CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error) {
    // 1. Business validation
    if req.Subdomain != nil {
        exists, err := s.repo.Exists(ctx, *req.Subdomain)
        if err != nil {
            return nil, fmt.Errorf("validation failed: %w", err)
        }
        if exists {
            return nil, ErrSubdomainAlreadyExists
        }
    }
    
    // 2. Apply business rules
    tenant := &Tenant{
        ID:     uuid.New(),
        Name:   req.Name,
        Status: StatusActive,
        // ... business logic
    }
    
    // 3. Persist via repository
    return s.repo.Create(ctx, tenant)
}
```

### 3. Repository Layer (`/internal/core/*/repository.go`)

**Purpose**: Abstract data access and persistence

#### Components:
- **Repositories**: Implement data access interfaces
- **Converters**: Transform between domain and database models
- **Store Integration**: Use SQLC-generated code for database operations

#### Responsibilities:
- ✅ Data access abstraction
- ✅ Domain ↔ Database model conversion
- ✅ Database error handling
- ✅ Transaction management
- ✅ Query optimization
- ❌ Business logic
- ❌ HTTP concerns

#### Key Patterns:
- **Interface Segregation**: Small, focused repository interfaces
- **Model Conversion**: Always convert at boundaries
- **Error Wrapping**: Convert database errors to domain errors

#### Example Structure:
```go
// Repository implements data access interface
func (r *repository) Create(ctx context.Context, tenant *Tenant) error {
    // 1. Convert domain model to database parameters
    params := db.CreateTenantParams{
        ID:        tenant.ID,
        Name:      tenant.Name,
        Slug:      tenant.Slug,
        Status:    string(tenant.Status),
    }
    
    // 2. Use SQLC-generated method
    _, err := r.store.CreateTenant(ctx, params)
    if err != nil {
        return fmt.Errorf("failed to create tenant: %w", err)
    }
    
    return nil
}
```

### 4. Infrastructure Layer (`/internal/platform/`)

**Purpose**: Provide technical infrastructure and external integrations

#### Components:
- **Database**: Connection management, pooling, migration
- **Cache**: Redis client, caching strategies
- **Configuration**: Environment-based configuration
- **External Services**: Third-party API integrations

#### Responsibilities:
- ✅ Database connection management
- ✅ Cache implementation
- ✅ Configuration management
- ✅ External service clients
- ✅ Infrastructure monitoring
- ❌ Business logic
- ❌ Domain models

## 🔄 Dependency Flow

### Correct Dependency Direction
```
API Layer ──depends on──▶ Core Layer ──depends on──▶ Repository Interfaces
                                                           ▲
Infrastructure Layer ──implements──▶ Repository Layer ────┘
```

### Key Rules:
1. **Higher layers can import lower layers**
2. **Lower layers cannot import higher layers**
3. **Core layer depends on interfaces, not implementations**
4. **Infrastructure implements core interfaces**

### Example Dependency Injection:
```go
// main.go - Wire dependencies
func main() {
    // Infrastructure
    db := database.Connect(cfg.Database)
    cache := redis.NewClient(cfg.Redis)
    store := sqlc.NewStore(db)
    
    // Repository (implements core interfaces)
    tenantRepo := tenant.NewRepository(store)
    
    // Service (depends on interfaces)
    tenantService := tenant.NewService(tenantRepo, cache)
    
    // Handler (depends on service)
    tenantHandler := handlers.NewTenantHandler(tenantService)
    
    // Router setup
    router.Setup(tenantHandler)
}
```

## 🏛️ Clean Architecture Benefits

### 1. **Testability**
- Each layer can be tested independently
- Easy to mock dependencies
- Clear separation of concerns

### 2. **Maintainability**
- Changes in one layer don't affect others
- Clear boundaries and responsibilities
- Easy to locate and fix issues

### 3. **Flexibility**
- Can swap implementations (database, cache, etc.)
- Independent deployment of layers
- Technology-agnostic core business logic

### 4. **Scalability**
- Each layer can be optimized independently
- Clear performance bottleneck identification
- Easy to add new features

## 🚨 Common Anti-Patterns to Avoid

### ❌ **Wrong Dependency Direction**
```go
// DON'T: Core layer importing API layer
import "internal/api/handlers" // ❌ Wrong!
```

### ❌ **Business Logic in Handlers**
```go
// DON'T: Business logic in API layer
func (h *Handler) CreateTenant(c *gin.Context) {
    // ❌ Business validation in handler
    if len(req.Name) < 3 {
        c.JSON(400, gin.H{"error": "Name too short"})
        return
    }
}
```

### ❌ **Database Models in Service Layer**
```go
// DON'T: Use database models in service
func (s *Service) CreateTenant(req CreateRequest) (*db.Tenant, error) {
    // ❌ Returning database model from service
}
```

### ❌ **Direct Database Access from Service**
```go
// DON'T: Database access in service
func (s *Service) GetTenant(id uuid.UUID) (*Tenant, error) {
    // ❌ Direct database query in service
    return s.db.Query("SELECT * FROM tenants WHERE id = ?", id)
}
```

## ✅ Best Practices

### 1. **Keep Layers Pure**
- Each layer should only handle its specific concerns
- Avoid mixing responsibilities

### 2. **Use Interfaces for Dependencies**
- Core layer depends on interfaces
- Infrastructure implements interfaces

### 3. **Convert Models at Boundaries**
- Repository layer converts between domain and database models
- API layer converts between HTTP and domain models

### 4. **Handle Errors Appropriately**
- Domain errors at service layer
- Infrastructure errors wrapped at repository layer
- HTTP errors formatted at handler layer

### 5. **Maintain Clear Contracts**
- Define clear interfaces between layers
- Document expected behaviors
- Use meaningful error types

---

📚 **Next Steps**: 
- [Data Flow Patterns](./data-flow.md) - See how requests flow through layers
- [Code Examples](./code-examples.md) - Practical implementation examples
- [Best Practices](./best-practices.md) - Development guidelines
