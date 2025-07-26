# ABAC Implementation Plan & Progress Tracker

## 📋 Project Overview

This document tracks the implementation progress of the Attribute-Based Access Control (ABAC) system for the ERP platform. The ABAC system will provide fine-grained, dynamic access control based on user attributes, resource attributes, environmental context, and organizational policies.

## 🎯 Implementation Goals

- [ ] Implement comprehensive ABAC policy engine
- [ ] Create attribute management system
- [ ] Build policy evaluation engine with real-time decisions
- [ ] Establish policy conflict resolution mechanisms
- [ ] Integrate with existing RBAC system
- [ ] Provide audit trail and compliance features
- [ ] Optimize for enterprise-scale performance

## 📅 Implementation Timeline

**Total Estimated Duration:** 8-10 weeks  
**Start Date:** [To be filled]  
**Target Completion:** [To be filled]

---

## Phase 1: Foundation & Data Models (Week 1-2) ✅ **COMPLETED**

### 1.1 Database Schema Implementation ✅ **COMPLETED**
- [x] Integrate with `internal/shared/errors/` for enhanced error handling
- [x] Integrate with `internal/shared/tracing/` for distributed tracing
- [x] Integrate with `internal/shared/logger/` for structured logging  
- [x] Integrate with `internal/shared/metrics/` for performance monitoring
- [x] Follow existing tenant context patterns from TENANT_CONTEXT_LIFECYCLE.md
- [x] Create `attribute_definitions` table with tenant isolation (RLS) - **Schema exists in db/migration/**
- [x] Create `policies` table with JSONB fields and tenant isolation - **Schema exists in db/migration/**
- [x] Create `policy_evaluations` cache table with tenant-aware keys - **Schema exists in db/migration/**
- [x] Create `attribute_sources` table for external attribute providers - **Schema exists in db/migration/**
- [x] Add ABAC-specific indexes for performance with tenant partitioning - **Schema exists in db/migration/**
- [x] Implement Row Level Security (RLS) policies using existing patterns - **Schema exists in db/migration/**
- [x] Create database migration scripts following existing SQLC patterns - **Schema exists in db/migration/**
- [x] Create SQLC queries using `current_tenant_id()` function for automatic tenant filtering

### 1.2 Core Domain Models ✅ **COMPLETED**
- [x] Define `AttributeDefinition` struct with validation using shared/errors
- [x] Define `Policy` struct with target/rule/obligations and tenant context
- [x] Define `PolicyEvaluationRequest` and `Response` structs with tracing context
- [x] Define `AttributeValue` and `AttributeContext` structs
- [x] Define `PolicyDecision` enum (ALLOW/DENY/NOT_APPLICABLE)
- [x] Define `PolicyCombiningAlgorithm` types
- [x] Implement model validation methods with enhanced error reporting
- [x] **NEW: Created comprehensive ABAC types in `internal/shared/types/abac.go`**
- [x] **NEW: Enhanced domain models in `internal/core/abac/models/domain.go`**
- [x] **NEW: Added ABAC-specific errors to `internal/shared/errors/errors.go`**
- [x] **NEW: Full OpenTelemetry tracing integration with policy-specific attributes**
- [x] **NEW: Structured logging with policy evaluation context**
- [x] **NEW: Metrics collection for policy operations and performance**

### 1.3 Repository Layer ✅ **COMPLETED**
- [x] Implement `AttributeDefinitionRepository` interface with tenant RLS
- [x] Implement `PolicyRepository` interface with SQLC integration
- [x] Implement `PolicyEvaluationRepository` interface with metrics
- [x] Create SQLC queries using `current_tenant_id()` for automatic tenant filtering
- [x] Add error handling using shared/errors BusinessError types
- [x] Implement caching layer using internal/platform/cache with tenant isolation
- [x] **NEW: Created comprehensive repository interfaces in `internal/core/abac/repository/interfaces.go`**
- [x] **NEW: Implemented PolicyRepository with full SQLC integration in `internal/core/abac/repository/policy_repository.go`**
- [x] **NEW: Added tenant-aware caching with automatic key isolation**
- [x] **NEW: Integrated OpenTelemetry tracing with policy-specific attributes**
- [x] **NEW: Added performance metrics and structured logging**

**Deliverables:** ✅ **ALL COMPLETED**
- [x] Complete database schema with migrations - **Schema exists in db/migration/**
- [x] Domain models with full validation - **Enhanced models with shared services integration**
- [x] Repository interfaces with SQLC integration - **Full repository layer with caching**
- [ ] Unit tests for all repository operations - **Pending for next phase**

---

## Phase 2: Attribute Management System (Week 2-3)

### 2.1 Attribute Definition Service
- [ ] Implement `AttributeDefinitionService` interface
- [ ] Create attribute validation logic
- [ ] Implement attribute data type handling (STRING, NUMBER, BOOLEAN, DATE, JSON, ARRAY, ENUM)
- [ ] Add attribute category management (USER, RESOURCE, ENVIRONMENT, ACTION, ENTITY, SESSION)
- [ ] Implement default value and allowed values logic
- [ ] Create attribute encryption/decryption for sensitive data

### 2.2 Attribute Collection Engine
- [ ] Implement `AttributeCollector` interface
- [ ] Create user attribute collector (from identity module)
- [ ] Create resource attribute collector
- [ ] Create environment attribute collector (time, location, device)
- [ ] Create session attribute collector
- [ ] Implement attribute caching with TTL
- [ ] Add attribute freshness validation

### 2.3 Attribute Sources Integration
- [ ] Integrate with identity service for user attributes
- [ ] Integrate with tenant/entity service for organizational attributes
- [ ] Integrate with session management for context attributes
- [ ] Create external attribute source adapters (future: LDAP, HR systems)
- [ ] Implement attribute source failover mechanisms

**Deliverables:**
- [ ] Complete attribute management service
- [ ] Attribute collection engine with multiple sources
- [ ] Attribute caching and validation system
- [ ] Integration tests for attribute collection

---

## Phase 3: Policy Engine Core (Week 3-5)

### 3.1 Policy Evaluation Engine
- [ ] Implement `PolicyEvaluationEngine` interface
- [ ] Create policy target matching logic
- [ ] Implement policy rule evaluation engine
- [ ] Create condition evaluation logic (AND, OR, NOT operations)
- [ ] Implement comparison operators (EQ, NE, GT, LT, GTE, LTE, IN, NOT_IN, LIKE, REGEX)
- [ ] Add function evaluation support (time_between, geo_within, etc.)
- [ ] Implement policy obligations and advice handling

### 3.2 Policy Combining Algorithms
- [ ] Implement `DenyOverrides` combining algorithm
- [ ] Implement `PermitOverrides` combining algorithm
- [ ] Implement `FirstApplicable` combining algorithm
- [ ] Implement `OnlyOneApplicable` combining algorithm
- [ ] Create configurable combining algorithm selection
- [ ] Add policy priority-based resolution

### 3.3 Policy Decision Point (PDP)
- [ ] Implement main `PolicyDecisionPoint` service
- [ ] Create policy retrieval and filtering logic
- [ ] Implement decision caching with context-aware keys
- [ ] Add performance monitoring and metrics
- [ ] Implement decision audit logging
- [ ] Create policy evaluation result formatting

### 3.4 Policy Information Point (PIP)
- [ ] Implement `PolicyInformationPoint` interface
- [ ] Create attribute fetching coordination
- [ ] Implement attribute resolution strategies
- [ ] Add missing attribute handling policies
- [ ] Create attribute dependency resolution
- [ ] Implement attribute transformation functions

**Deliverables:**
- [ ] Complete policy evaluation engine
- [ ] All combining algorithms implemented
- [ ] PDP service with caching and monitoring
- [ ] PIP with attribute resolution
- [ ] Performance benchmarks and optimization

---

## Phase 4: Policy Management & Administration (Week 5-6)

### 4.1 Policy CRUD Operations
- [ ] Implement policy creation with validation
- [ ] Implement policy update with version control
- [ ] Implement policy deletion with dependency checks
- [ ] Create policy import/export functionality
- [ ] Add policy template system
- [ ] Implement policy categorization and tagging

### 4.2 Policy Testing & Simulation
- [ ] Create policy testing framework
- [ ] Implement "what-if" policy simulation
- [ ] Create policy impact analysis tools
- [ ] Implement policy conflict detection
- [ ] Add policy coverage analysis
- [ ] Create policy effectiveness metrics

### 4.3 Policy Lifecycle Management
- [ ] Implement policy versioning system
- [ ] Create policy approval workflow
- [ ] Add policy activation/deactivation
- [ ] Implement policy expiration handling
- [ ] Create policy backup and restore
- [ ] Add policy change notifications

**Deliverables:**
- [ ] Complete policy management system
- [ ] Policy testing and simulation tools
- [ ] Policy lifecycle management features
- [ ] Admin UI for policy management

---

## Phase 5: Integration & Optimization (Week 6-7)

### 5.1 RBAC-ABAC Hybrid Integration
- [ ] Implement hybrid evaluation logic
- [ ] Create RBAC-to-ABAC policy migration tools
- [ ] Add role-based attribute inheritance
- [ ] Implement permission elevation through ABAC
- [ ] Create compatibility layer for existing RBAC
- [ ] Add gradual migration support

### 5.2 Performance Optimization
- [ ] Implement policy evaluation result caching
- [ ] Create attribute value caching with invalidation
- [ ] Add policy compilation and optimization
- [ ] Implement batch evaluation for bulk operations
- [ ] Create policy evaluation parallelization
- [ ] Add database query optimization

### 5.3 Monitoring & Observability
- [ ] Implement evaluation time metrics
- [ ] Create policy decision rate monitoring
- [ ] Add attribute resolution performance tracking
- [ ] Implement anomaly detection for access patterns
- [ ] Create policy evaluation dashboards
- [ ] Add alerting for policy failures

**Deliverables:**
- [ ] RBAC-ABAC hybrid system
- [ ] Performance-optimized evaluation engine
- [ ] Comprehensive monitoring and observability
- [ ] Load testing and performance benchmarks

---

## Phase 6: API Layer & Client Integration (Week 7-8)

### 6.1 REST API Implementation
- [ ] Implement policy evaluation endpoints
- [ ] Create attribute management APIs
- [ ] Add policy management APIs
- [ ] Implement bulk evaluation endpoints  
- [ ] Create policy testing APIs
- [ ] Add monitoring and metrics APIs

### 6.2 GraphQL API (Optional)
- [ ] Design GraphQL schema for ABAC operations
- [ ] Implement resolvers for policy evaluation
- [ ] Create subscription support for real-time updates
- [ ] Add query complexity analysis
- [ ] Implement field-level authorization

### 6.3 Client SDK & Libraries
- [ ] Create Go client library
- [ ] Add TypeScript/JavaScript client
- [ ] Create policy DSL and builders
- [ ] Implement client-side caching
- [ ] Add retry and circuit breaker logic
- [ ] Create middleware for common frameworks

**Deliverables:**
- [ ] Complete REST API with documentation
- [ ] Optional GraphQL API
- [ ] Client libraries and SDKs
- [ ] API testing and validation

---

## Phase 7: Security & Compliance (Week 8-9)

### 7.1 Security Hardening
- [ ] Implement policy encryption at rest
- [ ] Add attribute value encryption for PII
- [ ] Create secure policy distribution
- [ ] Implement policy integrity verification
- [ ] Add rate limiting and DDoS protection
- [ ] Create security audit logging

### 7.2 Compliance Features
- [ ] Implement GDPR compliance features
- [ ] Add SOX audit trail requirements
- [ ] Create HIPAA-compliant attribute handling
- [ ] Implement data retention policies
- [ ] Add consent management integration
- [ ] Create compliance reporting tools

### 7.3 Audit & Forensics
- [ ] Create comprehensive audit logging
- [ ] Implement decision replay capability
- [ ] Add forensic analysis tools
- [ ] Create access pattern analytics
- [ ] Implement suspicious activity detection
- [ ] Add investigation workflow support

**Deliverables:**
- [ ] Security-hardened ABAC system
- [ ] Compliance-ready features
- [ ] Comprehensive audit and forensics capabilities

---

## Phase 8: Testing & Documentation (Week 9-10)

### 8.1 Testing Suite
- [ ] Unit tests for all components (>90% coverage)
- [ ] Integration tests for full workflows
- [ ] Performance tests with load scenarios
- [ ] Security penetration testing
- [ ] Compliance validation testing
- [ ] Regression test automation

### 8.2 Documentation
- [ ] API documentation with examples
- [ ] Policy authoring guide
- [ ] Administrative user manual
- [ ] Developer integration guide
- [ ] Troubleshooting documentation
- [ ] Security best practices guide

### 8.3 Deployment & Migration
- [ ] Create deployment automation scripts
- [ ] Implement database migration tools
- [ ] Create configuration management
- [ ] Add health check endpoints
- [ ] Implement graceful shutdown procedures
- [ ] Create disaster recovery procedures

**Deliverables:**
- [ ] Complete test suite with automation
- [ ] Comprehensive documentation
- [ ] Deployment and migration tools
- [ ] Production readiness validation

---

## 🔧 Integration with Shared Services

The ABAC implementation will integrate with the existing shared services infrastructure:

### Error Handling (`internal/shared/errors/`)
- **BusinessError Integration**: Use enhanced error system with specific ABAC error codes
- **Error Categories**: Leverage existing categories (Security, Validation, Business, Tenant)
- **Contextual Errors**: Provide detailed error messages with suggestions

**ABAC-Specific Error Types:**
```go
// ABAC Policy Errors
ErrPolicyNotFound = NewBusinessError("POLICY_NOT_FOUND", "ABAC policy not found")
ErrPolicyInvalid = NewBusinessError("POLICY_INVALID", "ABAC policy validation failed")
ErrPolicyConflict = NewBusinessError("POLICY_CONFLICT", "Conflicting ABAC policies detected")

// Attribute Errors  
ErrAttributeNotFound = NewBusinessError("ATTRIBUTE_NOT_FOUND", "Required attribute not found")
ErrAttributeInvalid = NewBusinessError("ATTRIBUTE_INVALID", "Attribute value validation failed")
ErrAttributeExpired = NewBusinessError("ATTRIBUTE_EXPIRED", "Attribute value has expired")

// Evaluation Errors
ErrEvaluationFailed = NewBusinessError("EVALUATION_FAILED", "Policy evaluation failed")
ErrEvaluationTimeout = NewBusinessError("EVALUATION_TIMEOUT", "Policy evaluation timed out")
```

### Tracing (`internal/shared/tracing/`)
- **Distributed Tracing**: Full OpenTelemetry integration for policy evaluation flows
- **Span Attributes**: Rich metadata for debugging and monitoring
- **Performance Monitoring**: Track evaluation latency and bottlenecks

**ABAC Tracing Integration:**
```go
// Policy evaluation with tracing
func (e *PolicyEvaluationEngine) EvaluatePolicy(ctx context.Context, req *EvaluationRequest) (*EvaluationResult, error) {
    ctx, span := e.tracing.StartSpan(ctx, "abac.policy.evaluate", 
        tracing.WithAttributes(
            attribute.String("policy.id", req.PolicyID),
            attribute.String("user.id", req.UserID),
            attribute.String("resource.type", req.ResourceType),
            attribute.String("action", req.Action),
        ))
    defer span.End()
    
    // Add evaluation context to span
    span.SetAttributes(
        attribute.String("tenant.id", req.TenantID),
        attribute.Int("attributes.count", len(req.Attributes)),
    )
    
    // Evaluation logic with sub-spans
    result, err := e.evaluateInternal(ctx, req)
    if err != nil {
        e.tracing.RecordError(ctx, err, tracing.WithErrorStatus())
        return nil, err
    }
    
    // Record evaluation result
    span.SetAttributes(
        attribute.String("evaluation.decision", result.Decision),
        attribute.Int("evaluation.time_ms", result.EvaluationTimeMS),
        attribute.Bool("evaluation.cache_hit", result.CacheHit),
    )
    
    return result, nil
}
```

### Logging (`internal/shared/logger/`)
- **Structured Logging**: JSON-formatted logs with rich context
- **Security Audit**: Comprehensive logging of policy decisions
- **Debug Information**: Detailed evaluation traces for troubleshooting

**ABAC Logging Patterns:**
```go
// Policy evaluation logging
func (e *PolicyEvaluationEngine) logEvaluation(ctx context.Context, req *EvaluationRequest, result *EvaluationResult) {
    e.logger.InfoContext(ctx, "ABAC policy evaluation completed",
        logger.Fields{
            "policy_id": req.PolicyID,
            "user_id": req.UserID,
            "resource_type": req.ResourceType,
            "action": req.Action,
            "decision": result.Decision,
            "evaluation_time_ms": result.EvaluationTimeMS,
            "cache_hit": result.CacheHit,
            "applicable_policies": result.ApplicablePolicies,
        })
}

// Security audit logging
func (e *PolicyEvaluationEngine) auditSecurityDecision(ctx context.Context, req *EvaluationRequest, result *EvaluationResult) {
    if result.Decision == "DENY" || result.Decision == "ALLOW" {
        e.logger.WarnContext(ctx, "ABAC security decision made",
            logger.Fields{
                "event_type": "security_decision",
                "decision": result.Decision,
                "user_id": req.UserID,
                "resource_id": req.ResourceID,
                "action": req.Action,
                "policy_reasons": result.PolicyDecisions,
                "ip_address": extractIPFromContext(ctx),
                "user_agent": extractUserAgentFromContext(ctx),
            })
    }
}
```

### Metrics (`internal/shared/metrics/`)
- **Policy Evaluation Metrics**: Track evaluation performance and success rates
- **Cache Performance**: Monitor attribute and policy cache hit rates
- **Security Metrics**: Monitor access patterns and anomalies

**ABAC Metrics Integration:**
```go
// Initialize ABAC metrics
func (s *ABACService) initializeMetrics() {
    s.policyEvaluationCounter = s.metrics.Counter(
        "abac_policy_evaluations_total",
        "Total number of ABAC policy evaluations",
        "decision", "policy_type", "tenant_id",
    )
    
    s.evaluationDuration = s.metrics.Histogram(
        "abac_evaluation_duration_seconds",
        "Time spent evaluating ABAC policies",
        metrics.StandardHTTPDurationBuckets(),
        "decision", "cache_hit",
    )
    
    s.attributeResolutionCounter = s.metrics.Counter(
        "abac_attribute_resolutions_total",
        "Total number of attribute resolutions",
        "attribute_type", "source", "success",
    )
    
    s.cacheMissCounter = s.metrics.Counter(
        "abac_cache_misses_total",
        "Total number of cache misses",
        "cache_type", "tenant_id",
    )
}

// Record metrics during evaluation
func (e *PolicyEvaluationEngine) recordMetrics(ctx context.Context, req *EvaluationRequest, result *EvaluationResult, duration time.Duration) {
    labels := metrics.Fields{
        "decision": result.Decision,
        "policy_type": req.PolicyType,
        "tenant_id": req.TenantID,
    }
    
    e.metrics.IncrementCounter("abac_policy_evaluations_total", labels)
    
    evalLabels := metrics.Fields{
        "decision": result.Decision,
        "cache_hit": fmt.Sprintf("%t", result.CacheHit),
    }
    e.metrics.ObserveHistogram("abac_evaluation_duration_seconds", duration.Seconds(), evalLabels)
}
```

## 🏢 Tenant Context Integration

Based on the tenant lifecycle documentation, ABAC will integrate seamlessly with the existing tenant context management:

### Tenant-Aware Policy Evaluation
- **Automatic Tenant Isolation**: Leverage RLS policies for tenant-specific policy storage
- **Tenant Context Propagation**: Use existing tenant context from middleware
- **Cross-Tenant Admin Operations**: Support admin roles for system-wide policy management

**Tenant Integration Pattern:**
```go
// Service layer integration with tenant context
func (s *ABACService) EvaluatePolicy(ctx context.Context, req *PolicyEvaluationRequest) (*PolicyEvaluationResult, error) {
    // Get current tenant from database session context (existing pattern)
    currentTenant, err := s.tenantService.GetCurrentTenant(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get current tenant context: %w", err)
    }
    
    // Add tenant context to evaluation request
    req.TenantID = currentTenant.ID
    req.TenantFeatures = currentTenant.FeatureFlags
    req.TenantSecurityLevel = currentTenant.SecurityLevel
    
    // Repository operations automatically use RLS for tenant isolation
    return s.evaluationEngine.Evaluate(ctx, req)
}

// Repository layer with automatic tenant isolation
func (r *PolicyRepository) GetApplicablePolicies(ctx context.Context, req *PolicyEvaluationRequest) ([]*Policy, error) {
    // SQLC query uses current_tenant_id() function for automatic tenant filtering
    // RLS policies ensure additional security layer
    return r.store.GetPoliciesForEvaluation(ctx, db.GetPoliciesForEvaluationParams{
        ResourceType: req.ResourceType,
        Action:       req.Action,
        UserType:     req.UserType,
        // tenant_id filtering handled by current_tenant_id() in SQL query
    })
}
```

### Database Session Context Usage
- **SetTenant Integration**: Use existing tenant service for context management
- **Transaction Support**: Leverage store's WithTenant method for policy operations
- **Admin Context**: Support admin_role for cross-tenant policy management

**Transaction Integration:**
```go
// Complex policy operations using tenant-aware transactions
func (s *ABACService) CreatePolicyWithAttributes(ctx context.Context, req *CreatePolicyRequest) (*Policy, error) {
    // Get current tenant context
    currentTenant, err := s.tenantService.GetCurrentTenant(ctx)
    if err != nil {
        return nil, err
    }
    
    // Use store's WithTenant method for transaction management
    var policy *Policy
    err = s.store.WithTenant(ctx, currentTenant.ID, func(ctx context.Context, txStore db.Store) error {
        // Create policy - tenant_id populated by current_tenant_id() function
        policyParams := db.CreatePolicyParams{
            Name:        req.Name,
            PolicyType:  req.PolicyType,
            Target:      req.Target,
            Rule:        req.Rule,
            // tenant_id is set automatically by SQLC query using current_tenant_id()
            // ... other fields
        }
        
        createdPolicy, err := txStore.CreatePolicy(ctx, policyParams)
        if err != nil {
            return err
        }
        
        // Create associated attribute definitions
        for _, attrDef := range req.AttributeDefinitions {
            attrParams := db.CreateAttributeDefinitionParams{
                PolicyID:     createdPolicy.ID,
                Name:         attrDef.Name,
                DataType:     attrDef.DataType,
                Category:     attrDef.Category,
                // tenant_id set automatically by current_tenant_id() in SQLC query
                // ... other fields
            }
            
            if err := txStore.CreateAttributeDefinition(ctx, attrParams); err != nil {
                return err
            }
        }
        
        policy = convertFromSQLCPolicy(createdPolicy)
        return nil
    })
    
    return policy, err
}
```

### Caching Integration (`internal/platform/cache/`)
- **Tenant-Specific Cache Keys**: Automatic tenant isolation using existing cache service
- **Cache Compression**: Built-in compression for large policy objects
- **Circuit Breaker**: Protection against cache failures
- **Performance Monitoring**: Cache metrics and statistics

**ABAC Cache Integration:**
```go
// Policy cache using existing cache service with tenant isolation
func (c *PolicyCache) GetPolicy(ctx context.Context, policyID uuid.UUID) (*Policy, error) {
    // Get current tenant for cache context
    currentTenant, err := c.tenantService.GetCurrentTenant(ctx)
    if err != nil {
        return nil, err
    }
    
    // Use existing cache service with tenant context and namespace
    ctx = cache.WithTenantAndNamespace(ctx, currentTenant.ID, currentTenant.Slug, "abac")
    
    // Cache key (tenant isolation is handled automatically by cache service)
    cacheKey := fmt.Sprintf("policy:%s", policyID)
    
    var policy Policy
    if err := c.cache.Get(ctx, cacheKey, &policy); err == nil {
        return &policy, nil // Cache hit
    }
    
    // Cache miss - get from repository (RLS handles tenant isolation)
    dbPolicy, err := c.repo.GetPolicyByID(ctx, policyID)
    if err != nil {
        return nil, err
    }
    
    // Cache with tenant-specific TTL based on tenant features
    ttl := c.getTenantCacheTTL(currentTenant)
    c.cache.Set(ctx, cacheKey, dbPolicy, ttl)
    
    return dbPolicy, nil
}

// Bulk policy operations using MGet/MSet
func (c *PolicyCache) GetPolicies(ctx context.Context, policyIDs []uuid.UUID) ([]*Policy, error) {
    currentTenant, err := c.tenantService.GetCurrentTenant(ctx)
    if err != nil {
        return nil, err
    }
    
    // Set tenant context for automatic isolation
    ctx = cache.WithTenantAndNamespace(ctx, currentTenant.ID, currentTenant.Slug, "abac")
    
    // Build cache keys
    keys := make([]string, len(policyIDs))
    for i, id := range policyIDs {
        keys[i] = fmt.Sprintf("policy:%s", id)
    }
    
    var policies []*Policy
    if err := c.cache.MGet(ctx, keys, &policies); err == nil {
        return policies, nil
    }
    
    // Fallback to repository for cache misses
    return c.repo.GetPoliciesByIDs(ctx, policyIDs)
}

// Attribute cache with compression for large attribute sets
func (c *AttributeCache) CacheAttributes(ctx context.Context, userID uuid.UUID, attributes map[string]interface{}) error {
    currentTenant, err := c.tenantService.GetCurrentTenant(ctx)
    if err != nil {
        return err
    }
    
    // Enable compression for potentially large attribute sets
    ctx = cache.WithTenantAndNamespace(ctx, currentTenant.ID, currentTenant.Slug, "abac-attrs")
    ctx = cache.WithCompression(ctx, true)
    
    cacheKey := fmt.Sprintf("user-attrs:%s", userID)
    ttl := 15 * time.Minute // Shorter TTL for dynamic attributes
    
    return c.cache.Set(ctx, cacheKey, attributes, ttl)
}

// Policy evaluation cache with circuit breaker protection
func (c *EvaluationCache) CacheEvaluation(ctx context.Context, req *EvaluationRequest, result *EvaluationResult) error {
    currentTenant, err := c.tenantService.GetCurrentTenant(ctx)
    if err != nil {
        return err
    }
    
    ctx = cache.WithTenantAndNamespace(ctx, currentTenant.ID, currentTenant.Slug, "abac-eval")
    
    // Create evaluation cache key from request context
    cacheKey := c.buildEvaluationKey(req)
    
    // Short TTL for evaluation results to ensure freshness
    ttl := 5 * time.Minute
    
    return c.cache.Set(ctx, cacheKey, result, ttl)
}
```

**Cache Invalidation Patterns:**
```go
// Invalidate policy cache when policies change
func (s *PolicyService) UpdatePolicy(ctx context.Context, policyID uuid.UUID, req *UpdatePolicyRequest) (*Policy, error) {
    // Update policy in repository
    updatedPolicy, err := s.repo.UpdatePolicy(ctx, policyID, req)
    if err != nil {
        return nil, err
    }
    
    // Invalidate related caches
    currentTenant, _ := s.tenantService.GetCurrentTenant(ctx)
    ctx = cache.WithTenantAndNamespace(ctx, currentTenant.ID, currentTenant.Slug, "abac")
    
    // Delete specific policy cache
    s.cache.Delete(ctx, fmt.Sprintf("policy:%s", policyID))
    
    // Delete evaluation cache pattern for this policy
    s.cache.DeletePattern(ctx, fmt.Sprintf("eval:*:policy:%s:*", policyID))
    
    return updatedPolicy, nil
}
```

## 🗄️ SQLC Query Patterns with Tenant Context

Following the existing tenant lifecycle patterns, all ABAC SQLC queries will use the `current_tenant_id()` function for automatic tenant filtering:

### Example SQLC Queries

**Policy Management:**
```sql
-- name: CreatePolicy :one
INSERT INTO policies (
    id, tenant_id, name, display_name, description, policy_type, 
    effect, priority, category, target, rule, obligations, advice, 
    is_active, created_by
) VALUES (
    gen_random_uuid(), 
    current_tenant_id(), -- Automatic tenant context
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, true, $13
) RETURNING *;

-- name: GetPolicyByID :one
SELECT * FROM policies 
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: GetPoliciesForEvaluation :many
SELECT * FROM policies 
WHERE target->>'resource_type' = $1 
  AND target->>'action' = $2
  AND tenant_id = current_tenant_id()
  AND is_active = true 
  AND deleted_at IS NULL
ORDER BY priority DESC;

-- name: UpdatePolicy :one
UPDATE policies 
SET name = $2, display_name = $3, description = $4, rule = $5, 
    obligations = $6, advice = $7, updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL
RETURNING *;

-- name: DeletePolicy :exec
UPDATE policies 
SET deleted_at = NOW(), is_active = false
WHERE id = $1 AND tenant_id = current_tenant_id();
```

**Attribute Management:**
```sql
-- name: CreateAttributeDefinition :one
INSERT INTO attribute_definitions (
    id, tenant_id, name, display_name, description, data_type, 
    category, is_required, is_sensitive, default_value, 
    allowed_values, validation_rules, is_active
) VALUES (
    gen_random_uuid(),
    current_tenant_id(), -- Automatic tenant context
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, true
) RETURNING *;

-- name: GetAttributeDefinitionsByCategory :many
SELECT * FROM attribute_definitions 
WHERE category = $1 
  AND tenant_id = current_tenant_id()
  AND is_active = true
ORDER BY name;

-- name: GetRequiredAttributes :many
SELECT * FROM attribute_definitions 
WHERE is_required = true 
  AND tenant_id = current_tenant_id()
  AND is_active = true;
```

**Policy Evaluation Cache:**
```sql
-- name: CreatePolicyEvaluation :one
INSERT INTO policy_evaluations (
    id, tenant_id, user_id, resource_id, action_id, context_hash,
    decision, applicable_policies, evaluation_time_ms, expires_at
) VALUES (
    gen_random_uuid(),
    current_tenant_id(), -- Automatic tenant context
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetCachedEvaluation :one
SELECT * FROM policy_evaluations 
WHERE user_id = $1 
  AND resource_id = $2 
  AND action_id = $3
  AND context_hash = $4
  AND tenant_id = current_tenant_id()
  AND expires_at > NOW();

-- name: InvalidateEvaluationCache :exec
DELETE FROM policy_evaluations 
WHERE tenant_id = current_tenant_id()
  AND (expires_at <= NOW() OR $1 = ANY(applicable_policies));
```

**Cross-Policy Operations:**
```sql
-- name: GetPoliciesWithConflicts :many
SELECT p1.id as policy1_id, p1.name as policy1_name,
       p2.id as policy2_id, p2.name as policy2_name,
       p1.priority as priority1, p2.priority as priority2
FROM policies p1
JOIN policies p2 ON p1.target = p2.target 
  AND p1.id != p2.id
  AND p1.effect != p2.effect
WHERE p1.tenant_id = current_tenant_id()
  AND p2.tenant_id = current_tenant_id()
  AND p1.is_active = true 
  AND p2.is_active = true
  AND p1.deleted_at IS NULL 
  AND p2.deleted_at IS NULL;

-- name: GetPolicyImpactAnalysis :many
SELECT p.id, p.name, p.effect, p.priority,
       COUNT(DISTINCT pe.user_id) as affected_users,
       COUNT(pe.id) as total_evaluations,
       AVG(pe.evaluation_time_ms) as avg_eval_time
FROM policies p
LEFT JOIN policy_evaluations pe ON p.id = ANY(pe.applicable_policies)
WHERE p.tenant_id = current_tenant_id()
  AND p.id = $1
  AND p.deleted_at IS NULL
GROUP BY p.id, p.name, p.effect, p.priority;
```

**Admin Queries (Cross-Tenant):**
```sql
-- name: GetAllPoliciesAdmin :many
-- Note: This query bypasses current_tenant_id() for admin operations
-- Should only be used with admin_role database context
SELECT p.*, t.name as tenant_name, t.slug as tenant_slug
FROM policies p
JOIN tenants t ON p.tenant_id = t.id
WHERE p.deleted_at IS NULL
ORDER BY t.name, p.name;

-- name: GetTenantPolicyStats :many
-- Admin query to get policy statistics across all tenants
SELECT t.id as tenant_id, t.name as tenant_name,
       COUNT(p.id) as total_policies,
       COUNT(CASE WHEN p.is_active THEN 1 END) as active_policies,
       COUNT(CASE WHEN p.effect = 'ALLOW' THEN 1 END) as allow_policies,
       COUNT(CASE WHEN p.effect = 'DENY' THEN 1 END) as deny_policies
FROM tenants t
LEFT JOIN policies p ON t.id = p.tenant_id AND p.deleted_at IS NULL
GROUP BY t.id, t.name
ORDER BY t.name;
```

### Repository Implementation Pattern

**Standard Repository Methods:**
```go
// Repository method using SQLC with automatic tenant filtering
func (r *policyRepository) GetPolicyByID(ctx context.Context, id uuid.UUID) (*Policy, error) {
    // SQLC query automatically uses current_tenant_id() for tenant filtering
    sqlcPolicy, err := r.store.GetPolicyByID(ctx, id)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, errors.ErrPolicyNotFound
        }
        return nil, fmt.Errorf("failed to get policy: %w", err)
    }
    
    return r.convertFromSQLCPolicy(sqlcPolicy)
}

// Create operation - tenant_id populated automatically
func (r *policyRepository) CreatePolicy(ctx context.Context, req *CreatePolicyRequest) (*Policy, error) {
    // SQLC CreatePolicy query uses current_tenant_id() internally
    // No need to pass tenant_id explicitly in params
    sqlcPolicy, err := r.store.CreatePolicy(ctx, db.CreatePolicyParams{
        Name:        req.Name,
        DisplayName: req.DisplayName,
        Description: req.Description,
        PolicyType:  req.PolicyType,
        Effect:      req.Effect,
        Priority:    req.Priority,
        Category:    req.Category,
        Target:      req.Target,
        Rule:        req.Rule,
        Obligations: req.Obligations,
        Advice:      req.Advice,
        CreatedBy:   req.CreatedBy,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create policy: %w", err)
    }
    
    return r.convertFromSQLCPolicy(sqlcPolicy)
}
```

### Benefits of Using `current_tenant_id()`

1. **Automatic Tenant Filtering**: No manual tenant_id parameters needed in Go code
2. **Security**: Database-level tenant isolation enforced by PostgreSQL session context
3. **Consistency**: Follows existing tenant lifecycle patterns in the codebase
4. **Performance**: Leverages existing RLS policies and database session management
5. **Maintainability**: Reduces parameter passing and potential tenant context errors

---

## 🔧 Technical Architecture

### Core Components

1. **Policy Decision Point (PDP)**
   - Central evaluation engine
   - Policy combining algorithms
   - Decision caching
   - Performance monitoring

2. **Policy Information Point (PIP)**
   - Attribute collection and resolution
   - Multiple attribute sources
   - Caching and freshness validation

3. **Policy Administration Point (PAP)**
   - Policy CRUD operations
   - Policy testing and simulation
   - Lifecycle management

4. **Policy Enforcement Point (PEP)**
   - Integration points for applications
   - Decision enforcement
   - Obligation execution

### Data Flow

```
Request → PEP → PDP ↔ PIP (attributes)
                  ↓
              PAP (policies)
                  ↓
              Decision → PEP → Response
```

---

## 📊 Success Metrics

### Performance Targets
- [ ] Policy evaluation < 10ms (95th percentile)
- [ ] Attribute resolution < 5ms (95th percentile)
- [ ] Support 10K+ concurrent evaluations
- [ ] 99.9% availability SLA
- [ ] Cache hit rate > 80%

### Quality Targets
- [ ] >90% test coverage
- [ ] Zero critical security vulnerabilities
- [ ] <1% policy evaluation errors
- [ ] Complete API documentation
- [ ] Performance benchmarks established

---

## 🚨 Risk Assessment & Mitigation

### High-Risk Items
- [ ] **Policy Engine Complexity** - Mitigation: Incremental development, extensive testing
- [ ] **Performance at Scale** - Mitigation: Early performance testing, caching strategies
- [ ] **Attribute Source Reliability** - Mitigation: Fallback mechanisms, circuit breakers
- [ ] **Policy Conflicts** - Mitigation: Clear combining algorithms, conflict detection tools

### Dependencies
- [ ] Identity management system completion
- [ ] Database performance optimization
- [ ] Caching infrastructure setup
- [ ] Monitoring and observability platform

---

## 📝 Notes & Decisions

### Architecture Decisions
- **Policy Storage**: PostgreSQL with JSONB for flexibility
- **Caching**: Redis for policy and attribute caching
- **Evaluation Engine**: Custom Go implementation for performance
- **API Design**: REST-first with optional GraphQL

### Open Questions
- [ ] Integration with external identity providers
- [ ] Multi-region policy distribution strategy
- [ ] Real-time policy updates mechanism
- [ ] Policy versioning and rollback procedures

---

## 👥 Team & Responsibilities

| Component | Owner | Status |
|-----------|-------|--------|
| Database Schema | [TBD] | Not Started |
| Policy Engine | [TBD] | Not Started |
| Attribute Management | [TBD] | Not Started |
| API Layer | [TBD] | Not Started |
| Testing & QA | [TBD] | Not Started |
| Documentation | [TBD] | Not Started |

---

**Last Updated:** January 26, 2025  
**Next Review:** February 2, 2025  
**Overall Progress:** 12.5% Complete (Phase 1 of 8 completed)

---

*This document is a living tracker and should be updated regularly as implementation progresses.*