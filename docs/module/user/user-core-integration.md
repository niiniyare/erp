# Core Modules Integration Guide: Identity, Access & ABAC

## 📋 Overview

This document explains how the three core security modules work together to provide a comprehensive, layered security architecture for the Awo ERP system. Understanding this integration is essential for new developers working on security-related features.

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        API Layer                                │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │   Identity  │  │   Access    │  │    ABAC     │            │
│  │  Endpoints  │  │ Endpoints   │  │ Endpoints   │            │
│  └─────────────┘  └─────────────┘  └─────────────┘            │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Core Domain Layer                            │
│                                                                 │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │     Identity    │  │     Access      │  │      ABAC       │ │
│  │ ┌─────────────┐ │  │ ┌─────────────┐ │  │ ┌─────────────┐ │ │
│  │ │   Users     │ │  │ │  Requests   │ │  │ │ Policies    │ │ │
│  │ │ Management  │ │  │ │  Approval   │ │  │ │ Evaluation  │ │ │
│  │ │             │ │  │ │ Execution   │ │  │ │ Attributes  │ │ │
│  │ └─────────────┘ │  │ └─────────────┘ │  │ └─────────────┘ │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
│           │                     │                     │        │
│           └─────────────────────┼─────────────────────┘        │
│                                 │                              │
└─────────────────────────────────┼──────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────┐
│                  Data Access Layer                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │  Identity   │  │   Access    │  │    ABAC     │            │
│  │ Repository  │  │ Repository  │  │ Repository  │            │
│  └─────────────┘  └─────────────┘  └─────────────┘            │
└─────────────────────────────────────────────────────────────────┘
```

## 🔐 Module Responsibilities

### Identity Module (`/internal/core/identity/`)
**Purpose**: User lifecycle management and authentication foundation

**Key Responsibilities:**
- User creation, updates, and deactivation
- User profile management
- Account status management (Active, Inactive, Locked, Suspended)
- Basic user information storage
- Foundation for all other security modules

**Core Files:**
- `model.go` - User domain model and account status definitions
- `service.go` - User management business logic
- `repository.go` - Data access patterns for users

### Access Module (`/internal/core/access/`)
**Purpose**: Role-based access control (RBAC) and access request workflows

**Key Responsibilities:**
- Role assignments and permissions
- Access request workflows
- Approval processes
- Conditional access controls
- Permission execution

**Core Subdirectories:**
- `request/` - Access request lifecycle management
- `approval/` - Approval workflow engine
- `execution/` - Access permission execution
- `conditional/` - Conditional access rules
- `permission/` - Permission management and caching

### ABAC Module (`/internal/core/abac/`)
**Purpose**: Attribute-based access control for fine-grained, context-aware security

**Key Responsibilities:**
- Policy-based authorization decisions
- Attribute collection and evaluation
- Context-aware security
- Real-time policy evaluation
- Advanced compliance and audit

**Core Subdirectories:**
- `models/` - ABAC domain models (policies, attributes, decisions)
- `repository/` - Data access for ABAC entities
- `services/` - Core ABAC business services
- `activities/` - Temporal workflow activities
- `workflows/` - Long-running processes

## 🔄 Integration Flow

### 1. **Identity → Access Flow**
When a user needs role assignments or permissions:

```go
// 1. Identity provides user context
user, err := identityService.GetUser(ctx, userID)

// 2. Access module creates request
accessRequest := &access.Request{
    UserID: user.ID,
    RequestType: access.RequestTypeRoleAssignment,
    // ... other fields
}

// 3. Access approval workflow
approved, err := accessService.ProcessRequest(ctx, accessRequest)
```

**Files Involved:**
- `internal/core/identity/service.go` - User retrieval
- `internal/core/access/request/access_request_service.go` - Request processing
- `internal/core/access/approval/approver_service.go` - Approval logic

### 2. **Access → ABAC Flow**
When evaluating complex permissions with context:

```go
// 1. Access module provides base permissions
permissions, err := accessService.GetUserPermissions(ctx, userID)

// 2. ABAC enhances with context-aware evaluation
abacRequest := &abac.PermissionEvaluationRequest{
    UserID: userID,
    Resource: "financial_report",
    Action: "read",
    Context: map[string]interface{}{
        "time": time.Now(),
        "location": "office",
        "permissions": permissions, // From access module
    },
}

result, err := abacService.EvaluatePermission(ctx, abacRequest)
```

**Files Involved:**
- `internal/core/access/permission/model.go` - Permission definitions
- `internal/core/abac/service.go` - ABAC evaluation orchestration
- `internal/core/abac/activities/policy_evaluation.go` - Policy evaluation logic

### 3. **Identity → ABAC Direct Flow**
For attribute-based decisions requiring user context:

```go
// 1. ABAC collects user attributes from Identity
userAttributes, err := abacService.CollectUserAttributes(ctx, userID)

// 2. ABAC evaluates policies with user context
decision, err := abacService.EvaluatePolicy(ctx, &abac.EvaluationRequest{
    Subject: abac.Subject{
        ID: userID,
        Attributes: userAttributes,
    },
    Resource: "sensitive_document",
    Action: "access",
})
```

**Files Involved:**
- `internal/core/identity/model.go` - User model
- `internal/core/abac/attribute_collector.go` - Attribute collection
- `internal/core/abac/policy_evaluation_engine.go` - Policy evaluation

## 📁 File Structure Guide for New Developers

### Essential Files to Understand First

#### 1. Domain Models (Start Here)
```
internal/core/identity/model.go          # User and account models
internal/core/access/request/access_request_model.go  # Access request models  
internal/core/abac/models/domain.go      # ABAC policies and attributes
```

#### 2. Service Interfaces
```
internal/core/identity/service.go        # User management operations
internal/core/access/request/interface.go # Access request operations
internal/core/abac/service.go           # ABAC evaluation operations
```

#### 3. Integration Points
```
internal/core/abac/service.go           # Shows integration with identity
internal/core/access/request/access_request_service.go # User context usage
internal/core/abac/attribute_collector.go # Identity data collection
```

### Module-Specific File Organization

#### Identity Module Structure
```
internal/core/identity/
├── model.go              # 🔑 User domain model, account status enums
├── service.go            # 🔧 User CRUD operations, validation logic
├── repository.go         # 💾 Data access interface
├── service_test.go       # 🧪 Business logic tests
├── repository_test.go    # 🧪 Data access tests
└── mock.go              # 🎭 Testing mocks
```

#### Access Module Structure
```
internal/core/access/
├── request/
│   ├── access_request_model.go    # 🔑 Request domain models
│   ├── access_request_service.go  # 🔧 Request lifecycle management
│   ├── access_request_repository.go # 💾 Request data access
│   ├── interface.go              # 📋 Service contracts
│   └── mock.go                   # 🎭 Testing utilities
├── approval/
│   ├── model.go                  # 🔑 Approval workflow models
│   ├── approver_service.go       # 🔧 Approval business logic
│   └── mock.go                   # 🎭 Testing mocks
├── execution/
│   ├── access_execution_service.go # 🔧 Permission execution
│   ├── interface.go              # 📋 Execution contracts
│   └── mock.go                   # 🎭 Testing utilities
├── conditional/
│   ├── model.go                  # 🔑 Conditional access rules
│   ├── conditional_access_service.go # 🔧 Conditional logic
│   ├── interface.go              # 📋 Service interface
│   └── mock.go                   # 🎭 Testing mocks
└── permission/
    ├── model.go                  # 🔑 Permission models
    ├── permission_cache.go       # ⚡ Performance optimization
    └── permission_cache_test.go  # 🧪 Cache behavior tests
```

#### ABAC Module Structure
```
internal/core/abac/
├── models/
│   └── domain.go                 # 🔑 Core ABAC domain models
├── repository/
│   ├── interfaces.go            # 📋 Repository contracts
│   ├── policy_repository.go     # 💾 Policy data access
│   ├── attribute_repository_impl.go # 💾 Attribute storage
│   └── policy_evaluation_repository_impl.go # 💾 Decision caching
├── services/
│   ├── policy_evaluation_service.go # 🔧 Core evaluation logic
│   ├── attribute_collection_service.go # 🔧 Attribute gathering
│   ├── policy_decision_service.go # 🔧 Decision making
│   ├── attribute_cache_service.go # ⚡ Attribute caching
│   └── policy_combining_service.go # 🔧 Policy combination rules
├── activities/
│   ├── policy_evaluation.go     # 🔄 Temporal workflow activities
│   ├── attribute_collection.go  # 🔄 Attribute collection workflows
│   └── cache_activities.go      # 🔄 Cache management workflows
├── workflows/
│   ├── policy_evaluation_workflow.go # 🔄 Long-running evaluations
│   └── cache_management_workflow.go # 🔄 Cache lifecycle management
├── service.go                   # 🔧 Main ABAC service orchestrator
├── attribute_collector.go       # 🔧 Multi-source attribute collection
├── policy_evaluation_engine.go  # 🔧 Policy evaluation engine
├── hybrid_evaluator.go         # 🔧 RBAC+ABAC hybrid evaluation
├── security_compliance.go      # 🔧 Compliance checking
└── temporal_service.go         # 🔄 Temporal workflow integration
```

## 🔗 Key Integration Patterns

### 1. **Layered Security Model**
```go
// Example: Document access check
func CheckDocumentAccess(ctx context.Context, userID uuid.UUID, docID uuid.UUID) (bool, error) {
    // Layer 1: Identity - Is user valid and active?
    user, err := identityService.GetUser(ctx, userID)
    if err != nil || user.Status != identity.AccountStatusActive {
        return false, errors.NewBusinessError("INVALID_USER", "User not found or inactive")
    }
    
    // Layer 2: Access - Does user have base permission?
    hasPermission, err := accessService.HasPermission(ctx, userID, "documents", "read")
    if err != nil || !hasPermission {
        return false, errors.NewBusinessError("INSUFFICIENT_PERMISSION", "User lacks base permission")
    }
    
    // Layer 3: ABAC - Context-aware fine-grained check
    decision, err := abacService.EvaluatePermission(ctx, &abac.PermissionEvaluationRequest{
        UserID: userID,
        Resource: fmt.Sprintf("document:%s", docID),
        Action: "read",
        Context: collectContextualData(ctx),
    })
    
    return decision.Allowed, err
}
```

### 2. **Event-Driven Updates**
```go
// When user status changes in Identity, notify other modules
func (s *identityService) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status AccountStatus) error {
    // Update in Identity
    err := s.repository.UpdateUserStatus(ctx, userID, status)
    if err != nil {
        return err
    }
    
    // Notify Access module to update permissions cache
    s.eventBus.Publish("user.status.changed", UserStatusChangedEvent{
        UserID: userID,
        NewStatus: status,
    })
    
    // Notify ABAC to invalidate attribute cache
    s.abacService.InvalidateUserCache(ctx, userID)
    
    return nil
}
```

### 3. **Shared Context Propagation**
```go
// Context flows through all three modules
type SecurityContext struct {
    UserID     uuid.UUID
    TenantID   uuid.UUID
    EntityID   uuid.UUID
    SessionID  string
    RequestID  string
    Timestamp  time.Time
}

// Each module enriches the context
func EnrichSecurityContext(ctx context.Context) context.Context {
    secCtx := &SecurityContext{
        RequestID: extractRequestID(ctx),
        Timestamp: time.Now(),
    }
    
    // Identity enriches with user info
    if userID := extractUserID(ctx); userID != uuid.Nil {
        secCtx.UserID = userID
        secCtx.TenantID = getUserTenantID(ctx, userID)
    }
    
    return context.WithValue(ctx, "security_context", secCtx)
}
```

## 🎯 Development Guidelines

### For New Developers

#### 1. **Start with Identity**
- Understand the `User` model in `identity/model.go`
- Learn user lifecycle operations in `identity/service.go`
- Practice with user creation and status management

#### 2. **Move to Access** (RBAC)
- Study request models in `access/request/access_request_model.go`
- Understand approval workflows in `access/approval/`
- Learn permission models in `access/permission/`

#### 3. **Master ABAC**
- Begin with basic models in `abac/models/domain.go`
- Understand policy evaluation in `abac/policy_evaluation_engine.go`
- Learn attribute collection in `abac/attribute_collector.go`

#### 4. **Integration Patterns**
- Study how `abac/service.go` imports and uses `identity` services
- Examine cross-module communication patterns
- Practice implementing security checks that span multiple modules

### Common Development Tasks

#### Adding a New User Attribute
1. **Identity Module**: Update `User` model in `model.go`
2. **Access Module**: Consider if attribute affects permissions
3. **ABAC Module**: Add attribute definition in `models/domain.go`
4. **Integration**: Update attribute collection in `attribute_collector.go`

#### Implementing New Permission Type  
1. **Access Module**: Define in `permission/model.go`
2. **ABAC Module**: Create corresponding policy templates
3. **Integration**: Update evaluation logic to handle new permission

#### Adding Context-Aware Security Rule
1. **ABAC Module**: Define in policy evaluation engine
2. **Access Module**: Integrate with conditional access
3. **Identity Module**: Provide necessary user context

## 📈 Monitoring and Observability

### Key Metrics to Track
```go
// Each module exposes metrics for monitoring integration health
type IntegrationMetrics struct {
    // Cross-module operation counts
    IdentityToAccessCalls    metrics.Counter
    AccessToABACCalls       metrics.Counter
    ABACToIdentityCalls     metrics.Counter
    
    // Integration performance
    CrossModuleLatency      metrics.Histogram
    CacheHitRates          metrics.Gauge
    
    // Error rates
    IntegrationErrors       metrics.Counter
    PermissionDenials      metrics.Counter
}
```

### Tracing Integration Points
- Each cross-module call should have proper tracing spans
- Context propagation should be monitored
- Decision audit trails should capture module interactions

## 🚀 Best Practices

### 1. **Maintain Module Independence**
- Each module should be deployable independently
- Use clear interfaces between modules
- Avoid tight coupling in data models

### 2. **Consistent Error Handling**
- Use shared error types from `internal/shared/errors`
- Maintain error context across module boundaries
- Provide meaningful error messages for integration failures

### 3. **Performance Optimization**
- Cache frequently accessed cross-module data
- Use bulk operations when possible
- Monitor and optimize integration latencies

### 4. **Security First**
- Always validate data crossing module boundaries
- Maintain audit trails for security decisions
- Implement proper authorization for inter-module communications

This integration architecture provides a robust, scalable foundation for enterprise security while maintaining clear separation of concerns and enabling independent development of each security domain.
