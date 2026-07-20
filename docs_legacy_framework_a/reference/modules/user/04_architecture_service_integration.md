> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
**Change Notes (2025-07-28):**
- **Major Revision:** Aligned document with the current ABAC-centric architecture.
- **File Structures:** Updated all file structure diagrams to match the current codebase.
- **Integration Flows:**
    - Removed the inaccurate sequential `Access -> ABAC` flow.
    - Rewrote the integration model to reflect ABAC as the central evaluation engine that queries Identity and Access modules for attributes.
    - Clarified that the Identity module acts as a primary Policy Information Point (PIP) for the ABAC service.
- **Security Model:** Updated the "Key Integration Patterns" section to demonstrate a more realistic, ABAC-driven evaluation process.
---

# Core Modules Integration Guide: Identity, Access & ABAC

##  Overview

This document explains how the three core security modules—Identity, Access, and ABAC—work together. The architecture is **ABAC-centric**, meaning the ABAC module is the primary decision-maker, while the Identity and Access modules primarily serve as **attribute providers** (also known as Policy Information Points or PIPs).

## ️ Architecture Overview

The high-level architecture remains the same, with a clear separation between the API, Core Domain, and Data Access layers.

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
│  │  (Attribute     │  │  (Attribute     │  │  (Policy        │ │
│  │   Provider)     │  │   Provider)     │  │   Engine)       │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
│           ▲                     ▲                              │
│           │                     │                              │
│           └───────(Queries for Attributes)───────┘             │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
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

##  Module Responsibilities

### Identity Module (`/internal/core/identity/`)
**Purpose**: Manages user lifecycle and serves as a primary **source of truth for user attributes**.

**Key Responsibilities:**
- User creation, updates, and deactivation.
- Storing core user information (status, profile, etc.).
- **Providing user attributes** (e.g., `user.status`, `user.roles`) to the ABAC module for policy evaluation.

### Access Module (`/internal/core/access/`)
**Purpose**: Manages access request workflows and serves as a **source of truth for temporary permissions and role assignments**.

**Key Responsibilities:**
- Handling workflows for access requests, approvals, and revocations.
- **Providing contextual attributes** (e.g., `request.status`, `approval.is_granted`) to the ABAC module.
- Executing decisions made by the ABAC engine (e.g., granting a role after an approved request).

### ABAC Module (`/internal/core/abac/`)
**Purpose**: Provides fine-grained, context-aware authorization decisions. It is the central **Policy Decision Point (PDP)**.

**Key Responsibilities:**
- **Collecting attributes** from various sources, including the Identity and Access modules.
- Evaluating policies based on the collected attributes.
- Making the final `Allow` or `Deny` decision.
- Caching policies and attributes for performance.

##  Core Integration Pattern: ABAC-Centric Evaluation

The primary integration flow is not a simple linear chain. Instead, the ABAC service acts as the central orchestrator when a permission check is required.

```
1. Request --------> [ API Endpoint ]
                         |
                         | Calls abacService.EvaluatePermission(...)
                         |
                         ▼
+--------------------[ ABAC Service ]----------------------+
|                                                          |
|  1. Receives evaluation request (User, Resource, Action) |
|                                                          |
|  2. Calls Attribute Collector                            |
|     |                                                    |
|     ├─> [ Identity Service ] --> Gets user status, roles  |
|     |                                                    |
|     ├─> [ Access Service ] ----> Gets request status      |
|     |                                                    |
|     └─> [ Other Sources ] -----> Gets env context         |
|                                                          |
|  3. Gathers all attributes                               |
|                                                          |
|  4. Evaluates policies with combined attributes          |
|                                                          |
|  5. Returns final decision (Allow/Deny)                  |
|                                                          |
+----------------------------------------------------------+
                         |
                         |
                         ▼
[ API Endpoint ] <---- Decision
```

### **Identity ↔ ABAC Integration**
This is the most critical integration. The ABAC service directly calls the Identity service to fetch user-related attributes.

```go
// In internal/core/abac/service.go

// The ABAC service has a direct dependency on the Identity service.
type service struct {
    // ...
    identityService identity.Service
    // ...
}

// When evaluating, it collects user attributes.
func (s *service) CollectUserAttributes(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
    // 1. ABAC calls Identity to get user data
    user, err := s.identityService.GetUserByID(ctx, userID)
    if err != nil {
        return nil, err
    }

    // 2. ABAC calls Identity to get role data
    roles, err := s.identityService.GetUserRoles(ctx, userID)
    // ...

    // 3. Attributes are packaged for the policy engine
    attributes := make(map[string]interface{})
    attributes["user.status"] = string(user.Status)
    // ...
    return attributes, nil
}
```
**Files Involved:**
- `internal/core/abac/service.go`: Shows the direct dependency and calls to `identityService`.
- `internal/core/identity/service.go`: Provides the methods (`GetUserByID`, `GetUserRoles`) that supply the attributes.

##  File Structure Guide for New Developers

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
internal/core/abac/service.go           # Shows integration with Identity service.
internal/core/abac/attribute_collector.go # Core logic for gathering attributes from multiple sources.
internal/core/access/request/access_request_service.go # Shows usage of a UserService interface to validate users.
```

### Module-Specific File Organization (Corrected)

#### Identity Module Structure
```
internal/core/identity/
├── model.go
├── model_test.go
├── repository.go
├── repository_test.go
├── service.go
├── service_test.go
└── mock.go
```

#### Access Module Structure
```
internal/core/access/
├── approval/
│   ├── approver_service.go
│   ├── model.go
│   └── mock.go
├── conditional/
│   ├── conditional_access_service.go
│   ├── interface.go
│   ├── model.go
│   └── mock.go
├── execution/
│   ├── access_execution_service.go
│   ├── interface.go
│   └── mock.go
├── permission/
│   ├── model.go
│   ├── permission_cache.go
│   └── permission_cache_test.go
└── request/
    ├── access_request_model.go
    ├── access_request_repository.go
    ├── access_request_service.go
    ├── interface.go
    ├── mock.go
    └── service_suite_test.go
```

#### ABAC Module Structure
```
internal/core/abac/
├── activities/
│   ├── attribute_collection.go
│   ├── cache_activities.go
│   └── policy_evaluation.go
├── models/
│   └── domain.go
├── repository/
│   ├── attribute_definition_repository.go
│   ├── attribute_repository_impl.go
│   ├── interfaces.go
│   ├── policy_evaluation_repository_impl.go
│   ├── policy_repository.go
│   └── policy_repository_impl.go
├── services/
│   ├── attribute_cache_service.go
│   ├── attribute_cache_service_fixed.go
│   ├── attribute_collection_service.go
│   ├── attribute_validation_service.go
│   ├── policy_combining_service.go
│   ├── policy_decision_service.go
│   └── policy_evaluation_service.go
├── worker/
│   └── ...
├── workflows/
│   ├── cache_management_workflow.go
│   └── policy_evaluation_workflow.go
├── api_layer.go
├── attribute_collector.go
├── attribute_resolver.go
├── attribute_service.go
├── attribute_service_test.go
├── compatibility_layer.go
├── external_sources.go
├── hybrid_evaluator.go
├── integration_test.go
├── migration_engine.go
├── monitoring_service.go
├── performance_optimizer.go
├── policy_evaluation_engine.go
├── policy_evaluation_engine_test.go
├── policy_lifecycle.go
├── policy_manager.go
├── policy_templates.go
├── policy_testing.go
├── security_compliance.go
├── security_compliance_test.go
├── service.go
└── temporal_service.go
```

##  Key Integration Patterns

### 1. **ABAC-Centric Security Model**
The previous "Layered Security Model" is misleading. The actual pattern is a single,  evaluation orchestrated by the ABAC service.

```go
// Example: Document access check
func CheckDocumentAccess(ctx context.Context, userID uuid.UUID, docID uuid.UUID) (bool, error) {
    // A single call to the ABAC service is the entry point.
    // The ABAC service is responsible for gathering all necessary attributes
    // from Identity, Access, and other modules internally.
    decision, err := abacService.EvaluatePermission(ctx, &abac.PermissionEvaluationRequest{
        UserID:       userID,
        ResourceType: "document",
        ResourceID:   &docID,
        Action:       "read",
        Context:      collectContextualData(ctx), // e.g., IP address, time of day
    })

    if err != nil {
        // Handle errors during evaluation
        return false, err
    }
    
    // The final decision is returned.
    return decision.Decision == types.PolicyDecisionAllow, nil
}
```

### 2. **Event-Driven Updates**
This pattern remains crucial for cache invalidation. When data changes in a source module (like Identity), it must notify the ABAC module to prevent stale decisions.

```go
// When user status changes in Identity, notify other modules
func (s *identityService) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status AccountStatus) error {
    // ... update logic ...
    
    // Notify ABAC to invalidate its attribute cache for this user.
    // This is a critical step to ensure future decisions use fresh data.
    err := s.abacService.InvalidateUserCache(ctx, userID)
    if err != nil {
        // Log or handle the error
    }
    
    return nil
}
```

### 3. **Shared Context Propagation**
This pattern is essential and remains unchanged. The `context.Context` is the vehicle for carrying tenant information, tracing IDs, and other cross-cutting concerns through all service calls.

##  Development Guidelines

### For New Developers

1.  **Start with ABAC:** Since it's the central engine, begin by understanding its core concepts.
    - `abac/service.go`: The main entry point for evaluations.
    - `abac/attribute_collector.go`: The logic for gathering data.
    - `abac/models/domain.go`: The core data structures for policies.

2.  **Understand the Attribute Providers (PIPs):**
    - **Identity:** Study `identity/service.go` to see what user attributes are available.
    - **Access:** Study `access/request/access_request_model.go` to see what request-related attributes are available.

3.  **Trace an Evaluation:** Follow a call to `abacService.EvaluatePermission` and trace how it calls out to the `attributeCollectionActivities` and then to the `identityService` to build its context before making a decision.