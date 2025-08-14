# IAM Services Restructuring Implementation Plan

## 🎯 Overview

This document outlines the comprehensive plan to restructure existing IAM-related services into a unified, domain-driven IAM module following clean architecture principles. This is a **restructuring effort** that consolidates existing functionality while improving organization and maintainability.

## 📋 Current State Analysis

### Existing Services Structure

```
internal/core/
├── identity/           # Complete User/Person/Employee management (40+ files)
├── abac/              # Comprehensive ABAC system (40+ files) 
├── access/            # Access request workflows
├── analytics/         # User analytics (part of IAM)
├── audit/             # Security audit logging
└── iam/               # Empty directory structure
```

### Current Dependencies

- **ABAC Service** → **Identity Service** (as PIP - Policy Information Point)
- **Access Service** → **Identity Service** (for user context)
- **Analytics Service** → Part of IAM domain
- **Audit Service** → Required for all data operations
- **Feature Flag Service** → Existing service to be reused

### Existing Integration Patterns

- **Tenant Context Lifecycle**: PostgreSQL RLS with session variables
- **Database Transactions**: WithTenant, WithTx, BeginTxWithTenant patterns
- **Multi-tenant Cache**: Redis with automatic tenant isolation
- **Shared Services**: Tracing, Logger, Metrics (OpenTelemetry, structured logging, Prometheus)
- **SQLC Integration**: Type-safe database operations

## 🏗️ Target Architecture (UPDATED)

### New IAM Service Structure

```
internal/core/iam/
├── service.go                    # Main IAM Service interface (unified entry point)
├── model/                        # All IAM domain models and types
│   ├── types.go                  # IAM-specific enums and constants
│   └── entities.go               # All IAM domain models (User, Policy, etc.)
├── repo/                         # Repository interfaces
├── authn/                        # Authentication domain
│   └── service.go                # Login, password reset, MFA, user identity
├── authz/                        # Authorization domain
│   └── service.go                # ABAC, RBAC, hybrid access control
└── policy/                       # Policy domain
    └── service.go                # Policy evaluation, caching, lifecycle
```

**Key Improvements:**
- **Centralized Models**: All IAM types and entities in `model/` directory
- **Domain Separation**: Clear boundaries between `authn`, `authz`, and `policy`
- **Flat Structure**: Follows existing service patterns in the codebase
- **Local Types**: IAM-specific types copied from shared types for domain independence

### Unified IAM Service Interface (UPDATED)

```go
// IAM Service consolidates all identity and access management functionality
type Service interface {
    // Authentication operations (includes identity management)
    authn.Service
    
    // Authorization operations (ABAC, RBAC, hybrid)
    authz.Service
    
    // Policy operations (evaluation, caching, lifecycle)
    policy.Service
}

// Domain service interfaces
type authn.Service interface {
    // User Management
    CreateUser(ctx context.Context, req *CreateUserRequest) (*model.User, error)
    GetUser(ctx context.Context, userID uuid.UUID) (*model.User, error)
    UpdateUser(ctx context.Context, req *UpdateUserRequest) (*model.User, error)
    
    // Authentication
    Authenticate(ctx context.Context, req *AuthenticationRequest) (*AuthenticationResult, error)
    
    // Password Management & MFA
    ChangePassword(ctx context.Context, req *ChangePasswordRequest) error
    EnableMFA(ctx context.Context, req *EnableMFARequest) (*MFASetupResult, error)
    
    // Role Management
    AssignRole(ctx context.Context, req *AssignRoleRequest) error
    GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*model.Role, error)
}

type authz.Service interface {
    // Permission Evaluation
    EvaluatePermission(ctx context.Context, req *PermissionEvaluationRequest) (*PermissionEvaluationResult, error)
    BulkEvaluatePermissions(ctx context.Context, req *BulkPermissionEvaluationRequest) (*BulkPermissionEvaluationResult, error)
    
    // Access Requests
    CreateAccessRequest(ctx context.Context, req *CreateAccessRequestRequest) (*model.AccessRequest, error)
    ProcessAccessRequest(ctx context.Context, req *ProcessAccessRequestRequest) error
    
    // Permission Management
    GrantPermission(ctx context.Context, req *GrantPermissionRequest) error
    ListUserPermissions(ctx context.Context, userID uuid.UUID, entityID *uuid.UUID) ([]*model.Permission, error)
}

type policy.Service interface {
    // Policy Management
    CreatePolicy(ctx context.Context, req *CreatePolicyRequest) (*model.Policy, error)
    GetPolicy(ctx context.Context, policyID uuid.UUID) (*model.Policy, error)
    UpdatePolicy(ctx context.Context, req *UpdatePolicyRequest) (*model.Policy, error)
    
    // Policy Evaluation
    EvaluatePolicy(ctx context.Context, req *PolicyEvaluationRequest) (*PolicyEvaluationResult, error)
    TestPolicy(ctx context.Context, req *PolicyTestRequest) (*PolicyTestResult, error)
    
    // Policy Caching
    InvalidatePolicyCache(ctx context.Context, policyIDs []uuid.UUID) error
    GetPolicyCacheStats(ctx context.Context) (*PolicyCacheStats, error)
}
```

## 📊 Implementation Progress

### ✅ **COMPLETED WORK**

#### Phase 1.1: Foundation Structure (COMPLETED)
- ✅ **Improved Directory Structure**: Created domain-driven IAM structure
  ```
  internal/core/iam/
  ├── service.go         # ✅ Main unified IAM service interface
  ├── model/            # ✅ Centralized model directory
  │   ├── types.go      # ✅ IAM-specific enums and constants 
  │   └── entities.go   # ✅ All IAM domain models
  ├── repo/             # ✅ Repository interfaces directory
  ├── authn/            # ✅ Authentication domain
  │   └── service.go    # ✅ Authentication service interface
  ├── authz/            # ✅ Authorization domain  
  │   └── service.go    # ✅ Authorization service interface
  └── policy/           # ✅ Policy domain
      └── service.go    # ✅ Policy service interface
  ```

#### Phase 1.2: Model Definitions (COMPLETED)
- ✅ **Local Type Definitions**: Copied and localized IAM types from `internal/shared/types/`
  - ✅ `PolicyEffect`, `PolicyDecisionType`, `AttributeDataType`
  - ✅ `UserAccountStatus`, `EmploymentStatus`, `MFAMethod`
  - ✅ `RequestType`, `ApprovalStatus`, `SessionStatus`

- ✅ **Comprehensive Domain Models**: Created all IAM entity models
  - ✅ **Identity Models**: `User`, `Person`, `Employee`, `Role`, `UserRole`, `Session`
  - ✅ **Authorization Models**: `Permission`, `Policy`, `PolicyTarget`, `PolicyDecision`, `Attribute`
  - ✅ **Access Models**: `AccessRequest`, `ApprovalWorkflow`, `ConditionalAccessPolicy`
  - ✅ **Analytics Models**: `UserActivity`, `UserAnalytics`
  - ✅ **Policy Models**: `PolicyTemplate`, `PolicyVersion`, `PolicyConflict`

#### Phase 1.3: Service Interface Design (COMPLETED)
- ✅ **Main IAM Service**: Unified service interface using composition pattern
- ✅ **Authentication Service**: Complete interface with 25+ methods covering:
  - User/Person/Employee management
  - Authentication & token management
  - Password management & MFA
  - Session management & role assignment
- ✅ **Authorization Service**: Complete interface with 15+ methods covering:
  - Permission evaluation (single & bulk)
  - Access request workflows
  - Permission management & conditional access
- ✅ **Policy Service**: Complete interface with 20+ methods covering:
  - Policy lifecycle management
  - Policy evaluation & testing
  - Template & versioning support

#### Phase 1.4: Integration Patterns (COMPLETED)
- ✅ **Tenant Context Integration**: Helper methods for tenant management
- ✅ **Shared Services Integration**: Logger, metrics, tracing, audit, feature flags
- ✅ **Model Type Integration**: Updated service interfaces to use local model types
- ✅ **Mock Generation Setup**: Added proper mock generation directives

#### Phase 1.5: Repository Interface Design (COMPLETED)
- ✅ **Repository Interfaces**: Comprehensive interfaces for all IAM entities
- ✅ **Transaction Management**: Proper Store interface integration
- ✅ **Repository Implementation Structure**: Created IAMRepository with proper dependency injection

### ⚠️ **IN PROGRESS**

#### Phase 2.1: Repository Implementation (IN PROGRESS)
- ⚠️ **User Repository**: Partially implemented with Store transaction patterns
- ❌ **SQLC Integration**: Type mismatches and field mapping issues need resolution
- ❌ **Repository Constructors**: Missing implementations for all repositories except User

#### Phase 2.2: Service Logic Implementation (IN PROGRESS)
- ⚠️ **Authentication Service**: Core logic implemented but blocked by repository issues
- ❌ **Authorization Service**: Interface ready, implementation pending
- ❌ **Policy Service**: Interface ready, implementation pending

### 🔄 **CURRENT STATUS**
- **Phase 1**: ✅ **COMPLETED** - Foundation structure and interfaces ready
- **Phase 2**: ⚠️ **IN PROGRESS** - Repository and service implementation
- **Blocker**: SQLC type integration and repository completion needed

### 🧪 **TEST-DRIVEN COMPLETION CRITERIA**

Features are only marked as completed when ALL corresponding test cases pass. Based on `@docs/module/user/test_cases.md`:

#### Core Domain Model Tests (IAM-CORE-001 to IAM-CORE-007)
- ❌ Person model creation and validation
- ❌ Employee model creation and linking
- ❌ User model creation and security defaults
- ❌ Role hierarchy and permission models

#### Repository Layer Tests (IAM-REPO-001 to IAM-REPO-004)
- ❌ CRUD operations for all entities
- ❌ Tenant isolation (RLS) verification
- ❌ Complex queries (effective permissions, role hierarchy)

#### Service Layer Tests (IAM-SVC-001 to IAM-SVC-009)
- ❌ Authentication service (login, lockout, MFA)
- ❌ Authorization service (ABAC evaluation)
- ❌ Policy service (creation, caching, validation)

#### Integration Tests (IAM-API-001 to IAM-API-003)
- ❌ JWT token validation
- ❌ Tenant isolation via headers
- ❌ API endpoint functionality

#### Advanced Tests
- ❌ Hybrid Access Control (IAM-HYBRID-001 to IAM-HYBRID-004)
- ❌ Multi-tenant Isolation (IAM-MULTI-TENANT-001 to IAM-MULTI-TENANT-002)
- ❌ Security Tests (IAM-SEC-001 to IAM-SEC-003)
- ❌ Performance Tests (IAM-PERF-001 to IAM-PERF-002)
- ❌ End-to-End Workflows (IAM-E2E-001 to IAM-E2E-003)

---

## 📊 Implementation Phases

### Phase 1: Foundation Setup (Week 1) ✅ COMPLETED

#### 1.1 Create IAM Module Structure ✅ COMPLETED
- [x] Create main IAM service interface at `internal/core/iam/service.go`
- [x] Set up domain directories (authn, authz, policy)  
- [x] Create shared types and error definitions

#### 1.2 Service Interface Design ✅ COMPLETED
```go
// internal/core/iam/service.go
package iam

import (
    "context"
    "github.com/google/uuid"
    
    "github.com/niiniyare/erp/internal/core/iam/identity"
    "github.com/niiniyare/erp/internal/core/iam/authorization" 
    "github.com/niiniyare/erp/internal/core/iam/access"
    "github.com/niiniyare/erp/internal/core/iam/analytics"
    "github.com/niiniyare/erp/internal/shared/logger"
    "github.com/niiniyare/erp/internal/shared/metrics"
    "github.com/niiniyare/erp/internal/shared/tracing"
)

type Service interface {
    // Identity operations
    identity.Service
    
    // Authorization operations  
    authorization.Service
    
    // Access management operations
    access.Service
    
    // Analytics operations
    analytics.Service
}

type service struct {
    identityService     identity.Service
    authorizationService authorization.Service
    accessService       access.Service
    analyticsService    analytics.Service
    
    // Shared services
    logger  logger.Logger
    metrics metrics.MetricsProvider
    tracer  tracing.TracingService
    
    // Audit logging for all data operations
    auditService audit.Service
    
    // Feature flags
    featureFlag featureflag.Service
}
```

#### 1.3 Dependency Injection Setup ✅ COMPLETED
- [x] Create IAM service constructor with dependency injection
- [x] Integrate with existing shared services (logger, metrics, tracing)
- [x] Set up audit logging integration
- [x] Configure feature flag service integration

### Phase 2: Identity Domain Migration (Week 2)

#### 2.1 Move Identity Service
- [ ] Copy existing identity service to `internal/core/iam/identity/`
- [ ] Update import paths and package declarations
- [ ] Maintain existing interfaces and functionality
- [ ] Preserve User/Person/Employee model separation

#### 2.2 Database Integration
```go
// internal/core/iam/identity/service.go
func (s *service) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    // Use existing tenant context patterns
    return s.store.WithTenant(ctx, s.getCurrentTenantID(ctx), func(ctx context.Context, store Store) error {
        // Audit logging for data operations
        defer s.auditService.LogDataOperation(ctx, &audit.DataOperationLog{
            Operation: "CREATE_USER",
            EntityType: "user",
            // ... audit details
        })
        
        // Use existing SQLC operations
        return s.repository.CreateUser(ctx, req)
    })
}
```

#### 2.3 Cache Integration
- [ ] Integrate with existing multi-tenant Redis cache
- [ ] Use tenant-aware cache keys with automatic isolation
- [ ] Implement cache invalidation patterns

### Phase 3: Authorization Domain Migration (Week 3)

#### 3.1 Move ABAC Service
- [ ] Migrate comprehensive ABAC system to `internal/core/iam/authorization/`
- [ ] Preserve existing 40+ file structure and functionality
- [ ] Maintain PIP (Policy Information Point) integration with identity service

#### 3.2 Service Integration
```go
// internal/core/iam/authorization/service.go  
func NewService(
    identityService identity.Service, // PIP integration
    policyRepo repository.PolicyRepository,
    // ... other dependencies
) Service {
    // Preserve existing ABAC service construction
    return &service{
        identityService: identityService, // Continue using identity as PIP
        // ... existing setup
    }
}
```

#### 3.3 Preserve Existing Features
- [ ] Maintain all existing ABAC activities and workflows
- [ ] Keep policy evaluation caching mechanisms
- [ ] Preserve attribute collection from identity service

### Phase 4: Access Management Migration (Week 4)

#### 4.1 Move Access Service
- [ ] Migrate access request workflows to `internal/core/iam/access/`
- [ ] Preserve existing workflow integrations (Temporal)
- [ ] Maintain approval process functionality

#### 4.2 Workflow Integration
```go
// internal/core/iam/access/service.go
func (s *service) CreateAccessRequest(ctx context.Context, req *CreateAccessRequest) (*AccessRequest, error) {
    // Use existing transaction patterns
    return s.store.WithTenant(ctx, s.getCurrentTenantID(ctx), func(ctx context.Context, store Store) error {
        // Feature flag check using existing service
        if enabled, err := s.featureFlag.IsEnabled(ctx, "access_request_workflow"); err != nil || !enabled {
            return errors.New("access request workflow is disabled")
        }
        
        // Audit logging
        defer s.auditService.LogDataOperation(ctx, &audit.DataOperationLog{
            Operation: "CREATE_ACCESS_REQUEST",
            // ... audit details
        })
        
        return s.repository.CreateAccessRequest(ctx, req)
    })
}
```

### Phase 5: Analytics Integration (Week 5)

#### 5.1 Move Analytics Service
- [ ] Migrate user analytics to `internal/core/iam/analytics/`
- [ ] Integrate with existing tenant context management
- [ ] Preserve analytics data collection patterns

#### 5.2 Multi-tenant Analytics
```go
// internal/core/iam/analytics/service.go
func (s *service) TrackUserActivity(ctx context.Context, req *UserActivityRequest) error {
    // Use tenant-aware cache for analytics data
    cacheKey := fmt.Sprintf("user_activity:%s", req.UserID)
    tenantCtx := s.cache.WithTenant(ctx, s.getCurrentTenantID(ctx), s.getCurrentTenantSlug(ctx))
    
    // Store analytics data with tenant isolation
    return s.cache.Set(tenantCtx, cacheKey, req, time.Hour*24)
}
```

### Phase 6: Service Composition (Week 6)

#### 6.1 Unified IAM Service Implementation
```go
// internal/core/iam/service.go
func NewService(
    identityService identity.Service,
    authorizationService authorization.Service, 
    accessService access.Service,
    analyticsService analytics.Service,
    auditService audit.Service,
    featureFlagService featureflag.Service,
    logger logger.Logger,
    metrics metrics.MetricsProvider,
    tracer tracing.TracingService,
) Service {
    return &service{
        identityService:     identityService,
        authorizationService: authorizationService,
        accessService:       accessService,
        analyticsService:    analyticsService,
        auditService:        auditService,
        featureFlag:         featureFlagService,
        logger:              logger,
        metrics:             metrics,
        tracer:              tracer,
    }
}

// Delegate methods to domain services
func (s *service) GetUser(ctx context.Context, userID uuid.UUID) (*identity.User, error) {
    ctx, span := s.tracer.StartSpan(ctx, "iam.service.GetUser")
    defer span.End()
    
    s.metrics.IncrementCounter("iam_get_user_requests", nil)
    
    return s.identityService.GetUser(ctx, userID)
}

func (s *service) EvaluatePermission(ctx context.Context, req *authorization.PermissionRequest) (*authorization.PermissionResult, error) {
    ctx, span := s.tracer.StartSpan(ctx, "iam.service.EvaluatePermission")
    defer span.End()
    
    s.metrics.IncrementCounter("iam_permission_evaluations", nil)
    
    return s.authorizationService.EvaluatePermission(ctx, req)
}
```

#### 6.2 Update Dependency Injection
- [ ] Update main application to use unified IAM service
- [ ] Update API handlers to use single IAM service interface
- [ ] Maintain backward compatibility during transition

## 🔄 Migration Strategy

### Zero-Downtime Migration

1. **Parallel Implementation**: New IAM structure runs alongside existing services
2. **Gradual Migration**: Update consumers one by one to use unified IAM service
3. **Feature Flags**: Use existing feature flag service to control migration
4. **Rollback Plan**: Keep existing services until migration is complete

### Data Migration

No database schema changes required:
- All existing tables and relationships remain unchanged
- Tenant context and RLS policies continue working
- SQLC queries and models are preserved

### Testing Strategy

```go
// Test unified IAM service with existing patterns
func TestIAMService_Integration(t *testing.T) {
    // Use existing test patterns with multiple tenants
    tenantIDs := []uuid.UUID{
        uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"),
        uuid.MustParse("123e4567-e89b-12d3-a456-426614174002"),
    }
    
    for i, tenantID := range tenantIDs {
        t.Run(fmt.Sprintf("Tenant%d", i+1), func(t *testing.T) {
            ctx := WithTestTenantContext(context.Background(), tenantID)
            
            // Test identity operations
            user, err := iamService.CreateUser(ctx, &identity.CreateUserRequest{
                Name: "Test User",
                Email: "test@example.com",
            })
            require.NoError(t, err)
            
            // Test authorization
            result, err := iamService.EvaluatePermission(ctx, &authorization.PermissionRequest{
                UserID: user.ID,
                Resource: "users",
                Action: "read",
            })
            require.NoError(t, err)
            assert.Equal(t, authorization.Allow, result.Decision)
            
            // Verify tenant isolation
            verifyTenantIsolation(t, tenantID, user.ID)
        })
    }
}
```

## 🔧 Implementation Details

### Tenant Context Integration

All IAM services will use existing tenant management patterns:

```go
// Consistent tenant context usage across IAM services
func (s *service) getCurrentTenantID(ctx context.Context) uuid.UUID {
    return s.tenantService.GetCurrentTenant(ctx).ID
}

func (s *service) getCurrentTenantSlug(ctx context.Context) string {
    return s.tenantService.GetCurrentTenant(ctx).Slug
}
```

### Database Transaction Patterns

```go
// Use existing store patterns for tenant-aware transactions
func (s *service) performTenantAwareOperation(ctx context.Context, operation func(context.Context, Store) error) error {
    tenantID := s.getCurrentTenantID(ctx)
    return s.store.WithTenant(ctx, tenantID, operation)
}
```

### Cache Integration

```go
// Multi-tenant cache with automatic isolation
func (s *service) cacheUserData(ctx context.Context, userID uuid.UUID, user *User) error {
    tenantCtx := s.cache.WithTenantAndNamespace(
        ctx,
        s.getCurrentTenantID(ctx),
        s.getCurrentTenantSlug(ctx),
        "identity",
    )
    
    return s.cache.Set(tenantCtx, fmt.Sprintf("user:%s", userID), user, time.Hour)
}
```

### Audit Logging Integration

```go
// Audit all data operations using existing audit service
func (s *service) auditDataOperation(ctx context.Context, operation string, entityType string, entityID uuid.UUID) {
    auditLog := &audit.DataOperationLog{
        Operation:   operation,
        EntityType:  entityType,
        EntityID:    entityID,
        TenantID:    s.getCurrentTenantID(ctx),
        UserID:      s.getCurrentUserID(ctx),
        Timestamp:   time.Now(),
        RequestID:   s.getRequestID(ctx),
    }
    
    s.auditService.LogDataOperation(ctx, auditLog)
}
```

## 📝 API Integration

### Handler Updates

```go
// Update existing handlers to use unified IAM service
func (h *UserHandler) CreateUser(ctx context.Context, payload *user.CreateUserPayload) (*user.User, error) {
    // Single IAM service call instead of multiple service calls
    user, err := h.iamService.CreateUser(ctx, &identity.CreateUserRequest{
        Name:  payload.Name,
        Email: payload.Email,
    })
    if err != nil {
        return nil, err
    }
    
    // Convert to API model
    return h.convertToAPIModel(user), nil
}
```

### Dependency Injection

```go
// Update main.go to wire unified IAM service
func setupIAMService(
    store db.Store,
    cache cache.Service,
    auditService audit.Service,
    featureFlagService featureflag.Service,
    tenantService tenant.Service,
    logger logger.Logger,
    metrics metrics.MetricsProvider,
    tracer tracing.TracingService,
) iam.Service {
    // Create domain services
    identityService := identity.NewService(store, cache, tenantService, logger, metrics, tracer)
    authorizationService := authorization.NewService(identityService, store, cache, tenantService, logger, metrics, tracer)
    accessService := access.NewService(identityService, store, cache, tenantService, logger, metrics, tracer)
    analyticsService := analytics.NewService(store, cache, tenantService, logger, metrics, tracer)
    
    // Create unified IAM service
    return iam.NewService(
        identityService,
        authorizationService,
        accessService,
        analyticsService,
        auditService,
        featureFlagService,
        logger,
        metrics,
        tracer,
    )
}
```

## 📈 Benefits

### Improved Organization
- Single entry point for all IAM operations
- Clear domain separation within IAM
- Reduced complexity for API consumers

### Enhanced Maintainability
- Domain-driven structure improves code organization
- Centralized IAM service interface simplifies testing
- Clear separation of concerns within IAM domain

### Preserved Functionality
- All existing features and integrations maintained
- No breaking changes to database or API contracts
- Existing tenant isolation and security patterns preserved

### Better Integration
- Unified audit logging across all IAM operations
- Consistent tenant context management
- Shared error handling and validation

## 🎯 Success Criteria

- [ ] Zero downtime during migration
- [ ] All existing functionality preserved
- [ ] No database schema changes required
- [ ] All tests pass with new structure
- [ ] Performance metrics remain stable
- [ ] Tenant isolation verified across all IAM operations
- [ ] Audit logging working for all data operations
- [ ] Feature flag integration functional

## 📋 Rollback Plan

1. **Phase-by-phase rollback**: Each phase can be rolled back independently
2. **Feature flag controls**: Use existing feature flags to switch between old and new implementations
3. **Parallel operation**: Old services remain available during migration
4. **Monitoring**: Continuous monitoring of performance and error rates
5. **Database consistency**: No schema changes mean easy rollback at database level

---

This restructuring plan provides a comprehensive, low-risk approach to organizing IAM services while preserving all existing functionality and integration patterns. The phased approach allows for careful validation at each step and easy rollback if needed.