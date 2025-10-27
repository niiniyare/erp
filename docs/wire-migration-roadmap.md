# Google Wire Migration Roadmap for Awo ERP System

This document outlines the comprehensive plan for integrating Google Wire dependency injection into the Awo ERP system while maintaining Clean Architecture principles and multi-tenant isolation.

## Overview

### Current Architecture
- **Manual DI**: Services manually constructed with constructor injection
- **Interface-based**: Heavy use of interfaces for testability
- **Multi-tenant**: Context-aware services with tenant isolation
- **Clean Architecture**: Clear layer separation (API → Core → Platform)

### Wire Integration Goals
1. **Compile-time Safety**: Eliminate runtime DI errors
2. **Simplified Initialization**: Reduce boilerplate in main.go
3. **Maintain Architecture**: Preserve Clean Architecture boundaries
4. **Multi-tenant Support**: Handle tenant-scoped dependencies
5. **Testing Support**: Easy mock injection for tests

## Phase 1: Foundation (Database & Platform Dependencies)

### 1.1 Platform Layer Setup

**Timeline**: Week 1

**Tasks**:
- [x] Create `internal/platform/wire/` package structure
- [x] Implement platform providers (database, cache, config, observability)
- [x] Set up Wire provider sets for platform dependencies
- [x] Create database connection providers with proper pooling

**Files Created**:
- `internal/platform/wire/providers.go` - Main provider set definitions
- `internal/platform/wire/platform.go` - Platform-specific providers
- `internal/platform/wire/repositories.go` - Repository layer providers

**Key Providers**:
```go
// Platform Provider Set
var PlatformProviderSet = wire.NewSet(
    config.Load,                    // Configuration loading
    NewDatabaseConnection,          // PostgreSQL connection pool
    NewDBStore,                     // SQLC store
    NewCacheService,               // Redis cache
    NewLogger,                     // Structured logger
    NewMetricsProvider,            // Prometheus metrics
    NewTracingService,             // OpenTelemetry tracing
    NewTemporalClient,             // Temporal workflow client
)
```

### 1.2 Testing Platform Dependencies

**Tasks**:
- [ ] Create unit tests for each provider
- [ ] Test database connection pooling
- [ ] Verify cache connectivity
- [ ] Test configuration loading with various environments

### 1.3 Validation Criteria
- [ ] All platform dependencies initialize correctly
- [ ] Database migrations work with Wire-created connections
- [ ] Cache operations function properly
- [ ] Observability (logging, metrics, tracing) works end-to-end

## Phase 2: Core Business Services

### 2.1 Repository Layer

**Timeline**: Week 2

**Tasks**:
- [x] Create repository providers for all domains
- [x] Implement tenant-aware repository injection
- [x] Handle database store and cache dependencies

**Key Repositories**:
- Tenant, User, Person, Employee repositories
- Finance (Account, Transaction) repositories  
- Feature Flag, Audit, Settings repositories

### 2.2 Core Service Layer

**Timeline**: Week 2-3

**Tasks**:
- [x] Create service providers respecting dependency order
- [x] Implement service dependency injection
- [x] Handle circular dependency prevention

**Dependency Order**:
```
1. TenantService (no service dependencies)
2. SettingsService (depends on tenant)
3. AuditService (depends on tenant)
4. FeatureFlagService (depends on tenant, audit)
5. IAM Services (depends on tenant, settings, feature flags)
6. FinanceServices (depends on IAM, feature flags)
```

### 2.3 Multi-Tenant Service Scoping

**Tasks**:
- [x] Create tenant-scoped provider sets
- [x] Implement context-aware service creation
- [x] Handle Row Level Security (RLS) integration

**Tenant-Scoped Architecture**:
```go
// Tenant-scoped providers for per-request operations
var TenantScopedProviderSet = wire.NewSet(
    NewTenantScopedDBStore,         // Tenant-aware database operations
    NewTenantScopedCache,           // Tenant-specific cache keys
    NewTenantAwareRepositories,     // Tenant-scoped repository wrappers
)
```

### 2.4 Validation Criteria
- [ ] All core services initialize with correct dependencies
- [ ] Tenant isolation works correctly
- [ ] No circular dependencies detected by Wire
- [ ] Service method calls work end-to-end

## Phase 3: API Layer & HTTP Dependencies

### 3.1 Middleware Integration

**Timeline**: Week 3

**Tasks**:
- [x] Create middleware configuration providers
- [x] Implement tenant middleware with Wire dependencies
- [x] Set up authentication/authorization middleware

### 3.2 Handler Layer

**Timeline**: Week 3-4

**Tasks**:
- [x] Create handler dependency providers
- [x] Implement router with Wire-injected handlers
- [x] Set up API endpoint registration

**Handler Dependencies**:
```go
type Dependencies struct {
    Logger        logger.Logger
    Metrics       metrics.MetricsProvider
    Tracer        tracing.TracingService
    TenantService tenant.Service
    IAMService    iam.Service
    FinanceService *service.Services
}
```

### 3.3 Fiber App Configuration

**Tasks**:
- [x] Create Fiber app provider with middleware
- [x] Set up route registration through Wire
- [x] Configure error handling and observability

### 3.4 Validation Criteria
- [ ] All HTTP endpoints respond correctly
- [ ] Middleware chain functions properly
- [ ] Tenant context flows through requests
- [ ] Authentication and authorization work

## Phase 4: Testing & Validation

### 4.1 Unit Testing Integration

**Timeline**: Week 4

**Tasks**:
- [ ] Create test-specific Wire injectors
- [ ] Implement mock provider sets
- [ ] Update existing tests to use Wire

**Test Injectors**:
```go
// Test-specific injector for unit tests
func InitializeForTesting() (*Application, error) {
    wire.Build(
        TestPlatformProviderSet,     // Mock database, cache
        TestCoreServicesProviderSet, // Mock services
        TestAPIProviderSet,          // Test handlers
        NewApplication,
    )
    return &Application{}, nil
}
```

### 4.2 Integration Testing

**Tasks**:
- [ ] Test complete dependency graph
- [ ] Verify multi-tenant isolation
- [ ] Test error scenarios and edge cases
- [ ] Performance testing with Wire-injected services

### 4.3 Migration Validation

**Tasks**:
- [ ] Compare behavior before/after Wire migration
- [ ] Validate all existing functionality works
- [ ] Performance comparison (startup time, memory usage)
- [ ] Security validation (tenant isolation, access control)

## Integration with Existing Patterns

### Clean Architecture Preservation

**Layer Boundaries**:
```
┌─────────────────────────────────────────┐
│ API Layer (Wire: APIProviderSet)        │
├─────────────────────────────────────────┤
│ Core Layer (Wire: CoreServiceProviderSet)│
├─────────────────────────────────────────┤
│ Platform Layer (Wire: PlatformProviderSet)│
└─────────────────────────────────────────┘
```

**Dependency Rules**:
- API layer can depend on Core layer
- Core layer can depend on Platform layer
- No upward dependencies (enforced by Wire)

### Multi-Tenant Integration

**Tenant Context Flow**:
1. **Request Level**: Middleware extracts tenant ID from context
2. **Service Level**: Services receive tenant-aware dependencies
3. **Repository Level**: Queries automatically include tenant isolation
4. **Database Level**: Row Level Security enforces data isolation

**Wire Integration**:
```go
// Tenant context injected at request time
func (s *TenantMiddleware) HandleRequest(c *fiber.Ctx) error {
    tenantID := extractTenantID(c)
    
    // Create tenant-scoped dependencies
    tenantDeps := createTenantScopedDependencies(tenantID)
    
    // Store in context for handlers
    c.Locals("tenantDeps", tenantDeps)
    return c.Next()
}
```

### Rich Error Types

Wire integrates seamlessly with the existing error handling:
```go
// Provider error handling
func NewTenantService(repo Repository) (tenant.Service, error) {
    if repo == nil {
        return nil, errors.NewBusinessError(
            "MISSING_DEPENDENCY",
            "tenant repository is required",
        ).WithCategory(errors.CategorySystem)
    }
    return tenant.NewService(repo), nil
}
```

### Temporal Workflow Integration

Temporal workflows maintain their existing patterns with Wire-injected dependencies:
```go
// Workflow dependencies injected via Wire
type WorkflowDependencies struct {
    TenantService tenant.Service
    FinanceService *service.Services
    ActivityRegistry *activities.Registry
}

// Wire provider for workflow dependencies
func NewWorkflowDependencies(...) WorkflowDependencies {
    return WorkflowDependencies{...}
}
```

## Configuration Management

### Environment-Based Configuration

Wire providers adapt to different environments:
```go
func NewLogger(cfg *config.Config) (logger.Logger, error) {
    switch cfg.App.Stage {
    case config.DevelopmentStage:
        return logger.NewDevelopmentLogger()
    case config.ProductionStage:
        return logger.NewProductionLogger()
    default:
        return logger.NewDefaultLogger()
    }
}
```

### Feature Flag Integration

Feature flags control service behavior through Wire:
```go
func NewFinanceServices(
    deps service.Dependencies,
    featureFlags featureflag.Service,
) *service.Services {
    services := service.NewServices(deps)
    
    // Configure based on feature flags
    if featureFlags.IsEnabled("advanced_reporting") {
        services.EnableAdvancedReporting()
    }
    
    return services
}
```

## Implementation Commands

### 1. Install Wire
```bash
go install github.com/google/wire/cmd/wire@latest
```

### 2. Add Wire to go.mod
```bash
go get github.com/google/wire
```

### 3. Generate Wire Code
```bash
# Run the generation script
./scripts/generate-wire.sh

# Or manually
cd cmd/server && wire generate
```

### 4. Update Makefile
```makefile
.PHONY: wire
wire: ## Generate Wire dependency injection code
	@echo "$(BLUE)Generating Wire code...$(NC)"
	@./scripts/generate-wire.sh
	@echo "$(GREEN)✅ Wire generation complete$(NC)"

.PHONY: wire-check
wire-check: ## Check if Wire files are up to date
	@echo "$(BLUE)Checking Wire files...$(NC)"
	@wire check ./cmd/server/
	@echo "$(GREEN)✅ Wire files are up to date$(NC)"

# Add wire to generate target
generate: sqlc mock wire
```

### 5. Update CI/CD Pipeline
```yaml
# Add to your CI pipeline
- name: Generate Wire Code
  run: |
    go install github.com/google/wire/cmd/wire@latest
    make wire
    
- name: Verify Wire Code
  run: |
    make wire-check
    git diff --exit-code
```

## Success Metrics

### Immediate Benefits
- [ ] **Compile-time Safety**: Zero runtime DI errors
- [ ] **Reduced Boilerplate**: 50% reduction in main.go complexity
- [ ] **Faster Development**: Quick feedback on dependency issues

### Long-term Benefits
- [ ] **Maintainability**: Clear dependency graphs
- [ ] **Testing**: Easy mock injection
- [ ] **Performance**: Optimized dependency creation
- [ ] **Scalability**: Easy addition of new services

### Monitoring
- [ ] Application startup time
- [ ] Memory usage during initialization
- [ ] Test execution time
- [ ] Developer productivity metrics

## Risk Mitigation

### Potential Issues
1. **Wire Learning Curve**: Team needs to understand Wire concepts
2. **Complex Dependency Graphs**: Large number of dependencies
3. **Multi-tenant Complexity**: Context-aware dependency injection

### Mitigation Strategies
1. **Training**: Wire documentation and examples
2. **Gradual Migration**: Phase-by-phase implementation
3. **Testing**: Comprehensive test coverage for each phase
4. **Documentation**: Clear examples and troubleshooting guides

## Conclusion

This Wire migration plan provides a structured approach to modernizing the Awo ERP dependency injection while maintaining all existing functionality. The phased approach ensures minimal risk and allows for thorough testing at each stage.

The implementation maintains Clean Architecture principles, preserves multi-tenant isolation, and provides significant improvements in compile-time safety and developer experience.