# ABAC Implementation Plan - Temporal Workflows Architecture

## 📋 Project Overview

**Goal**: Implement ABAC (Attribute-Based Access Control) system using Temporal workflows while simplifying the identity module to focus only on basic person/employee management.

**Architecture Decision**: Use Temporal workflows to eliminate circular dependencies and provide robust, scalable ABAC functionality.

**System Relationship**: The new ABAC module (`@internal/core/abac/`) complements the existing access control system (`@internal/core/access/`) by providing automated policy-based decisions while the existing system handles human approval workflows.

## 📋 Implementation Guidelines

**Critical Requirements:**
- **Tenant Context Lifecycle**: Follow tenant context management guidelines from `@docs/TENANT_CONTEXT_LIFECYCLE.md` for all ABAC operations
- **Shared Components**: Use existing infrastructure from `@internal/shared/` for errors, logging, tracing, and metrics
- **Database Integration**: Ensure proper RLS (Row Level Security) enforcement and SQLC integration
- **Multi-Tenant Isolation**: All ABAC operations must respect tenant boundaries with proper context propagation

## 🎯 Phase-Based Implementation Plan

### **Phase 1: Module Restructuring** ⏱️ **Est: 3-4 days**

#### ✅ **Completed Items**
- [x] Created `user_activities` table for behavioral analytics
- [x] Created `attribute_definitions` table for standardized ABAC attributes  
- [x] Created `policy_evaluations` table for ABAC performance caching
- [x] Analyzed existing access control modules
- [x] Analyzed current identity service structure

#### 🔄 **Phase 1 Tasks**

##### **1.1 Simplify Identity Module** ✅ **COMPLETED**
- [x] **Task**: Remove ABAC-related methods from identity service
  - [x] Remove `EvaluatePermission()` method
  - [x] Remove `BulkEvaluatePermissions()` method  
  - [x] Remove `GetUserEffectivePermissions()` method
  - [x] Remove `CalculateRoleHierarchy()` method
  - [x] Remove `TestPolicy()` method
  - [x] Keep only: `RegisterNewUser`, `GetUserByID`, `CreatePerson`, `CreateEmployee`, basic CRUD
- [x] **Task**: Update identity service interface to focus on core identity management
- [x] **Task**: Remove ABAC-related imports and dependencies
- [x] **Task**: Update identity service tests to reflect simplified scope

##### **1.2 Create ABAC Module Structure** ✅ **COMPLETED**
- [x] **Task**: Create new module structure:
  ```
  internal/core/abac/
  ├── workflows/
  │   └── permission_evaluation.go     ✅ Core & bulk evaluation workflows
  ├── activities/
  │   ├── attribute_collection.go      ✅ User/resource/environment attributes
  │   ├── policy_evaluation.go         ✅ ABAC policy engine with rule evaluation
  │   └── cache_activities.go          ✅ Performance caching & audit logging
  ├── client/
  │   └── temporal_client.go           ✅ Temporal client wrapper & worker setup
  ├── models/
  │   └── abac_models.go               ✅ Complete workflow & activity models
  └── service.go                       ✅ ABAC service interface
  ```

##### **1.3 Database Schema Completion** ✅ **COMPLETED**
- [x] **Task**: Run migration to create missing tables
  - [x] Execute `user_activities` migration (Fixed partitioned table constraint)
  - [x] Execute `attribute_definitions` migration  
  - [x] Execute `policy_evaluations` migration
- [x] **Task**: Verify all ABAC tables exist and are properly indexed

---

### **Phase 2: Temporal Workflow Foundation** ⏱️ **Est: 5-6 days**

##### **2.1 Temporal Client Setup** ✅ **COMPLETED**
- [x] **Task**: Create Temporal client wrapper
  - [x] Connection management
  - [x] Worker configuration
  - [x] Workflow/activity registration
  - [x] Error handling and retries
- [x] **Task**: Configure Temporal server connection
- [x] **Task**: Set up development/testing Temporal instance

##### **2.2 Core ABAC Models** ⏳ **IN PROGRESS**
- [x] **Task**: Define workflow request/response models
  ```go
  type PermissionEvaluationWorkflowRequest
  type AccessRequestWorkflowRequest  
  type PolicyTestWorkflowRequest
  ```
- [x] **Task**: Define activity input/output models
- [x] **Task**: Define ABAC context models for Temporal
- [ ] **Task**: Create model validation and serialization

##### **2.3 Basic Repository Layer** ✅ **COMPLETED**
- [x] **Task**: Implement policy repository with SQLC (SQL queries and sqlc generation complete)
  - [x] CRUD operations for policies table
  - [x] Query policies by resource/action
  - [x] Policy filtering and sorting
- [x] **Task**: Implement attribute definition repository (SQL queries and sqlc generation complete)
- [x] **Task**: Implement policy evaluation cache repository (SQL queries and sqlc generation complete)
- [x] **Task**: Create repository interfaces and mocks

---

### **Phase 3: Core ABAC Activities** ⏱️ **Est: 7-8 days**

##### **3.1 Attribute Collection Activities**
- [ ] **Task**: Implement `CollectUserAttributesActivity`
  - [ ] Get user data from identity service
  - [ ] Get person attributes (if linked)
  - [ ] Get employee attributes (if linked)
  - [ ] Combine into unified attribute map
- [ ] **Task**: Implement `CollectResourceAttributesActivity`
  - [ ] Get resource metadata from database
  - [ ] Apply resource-specific attribute rules
  - [ ] Handle dynamic resource attributes
- [ ] **Task**: Implement `BuildEnvironmentContextActivity`
  - [ ] Time context (current time, business hours)
  - [ ] Location context (IP geolocation)
  - [ ] Device context (user agent parsing)
  - [ ] Risk context (behavioral analysis)

##### **3.2 Policy Evaluation Activities**
- [ ] **Task**: Implement `EvaluatePoliciesActivity`
  - [ ] Load applicable policies from database
  - [ ] Parse JSONB policy rules
  - [ ] Evaluate AND/OR conditions
  - [ ] Handle policy priorities and conflicts
  - [ ] Return detailed evaluation results
- [ ] **Task**: Implement policy rule engine
  - [ ] Support for comparison operators (eq, gt, lt, in, contains)
  - [ ] Support for complex nested conditions
  - [ ] Support for time-based rules
  - [ ] Support for location-based rules
- [ ] **Task**: Implement `CachePolicyResultActivity`
  - [ ] Generate cache keys
  - [ ] Store evaluation results with TTL
  - [ ] Handle cache invalidation

##### **3.3 Context Enrichment Activities**
- [ ] **Task**: Implement `EnrichDeviceContextActivity`
  - [ ] Parse user agent strings
  - [ ] Device fingerprinting
  - [ ] Trusted device validation
- [ ] **Task**: Implement `EnrichLocationContextActivity`
  - [ ] IP geolocation lookup
  - [ ] Trusted location validation
  - [ ] Country/region restrictions
- [ ] **Task**: Implement `CalculateRiskScoreActivity`
  - [ ] Behavioral anomaly detection
  - [ ] Risk factor aggregation
  - [ ] Risk-based access decisions

---

### **Phase 4: Permission Evaluation Workflow** ⏱️ **Est: 4-5 days**

##### **4.1 Core Permission Workflow**
- [ ] **Task**: Implement `PermissionEvaluationWorkflow`
  - [ ] Orchestrate attribute collection
  - [ ] Coordinate policy evaluation
  - [ ] Handle caching strategy
  - [ ] Return formatted results
- [ ] **Task**: Add workflow error handling and retries
- [ ] **Task**: Add workflow timeouts and cancellation
- [ ] **Task**: Add workflow logging and metrics

##### **4.2 Bulk Evaluation Workflow**
- [ ] **Task**: Implement `BulkPermissionEvaluationWorkflow`
  - [ ] Parallel activity execution
  - [ ] Batch optimization
  - [ ] Result aggregation
- [ ] **Task**: Add batch size limits and throttling
- [ ] **Task**: Add progress tracking for large batches

##### **4.3 Performance Optimization**
- [ ] **Task**: Implement cache-first evaluation strategy
- [ ] **Task**: Add attribute pre-loading for bulk operations
- [ ] **Task**: Add policy compilation and caching
- [ ] **Task**: Add evaluation time tracking and optimization

---

### **Phase 5: Access Request Workflow** ⏱️ **Est: 6-7 days**

##### **5.1 Access Request Workflow**
- [ ] **Task**: Implement `AccessRequestWorkflow`
  - [ ] Request validation
  - [ ] Approval requirement checking
  - [ ] Approver identification
  - [ ] Approval waiting with timeout
  - [ ] Access granting/rejection
- [ ] **Task**: Add human task integration (approval signals)
- [ ] **Task**: Add automatic timeout handling
- [ ] **Task**: Add notification integration

##### **5.2 Access Request Activities**
- [ ] **Task**: Implement `ValidateAccessRequestActivity`
  - [ ] Request format validation
  - [ ] Business rule validation
  - [ ] Duplicate request checking
- [ ] **Task**: Implement `FindApproversActivity`
  - [ ] Role-based approver lookup
  - [ ] Hierarchical approver lookup
  - [ ] Approver availability checking
- [ ] **Task**: Implement `GrantAccessActivity`
  - [ ] Temporary role assignment
  - [ ] Permission granting
  - [ ] Access expiration scheduling
- [ ] **Task**: Implement `RevokeAccessActivity`
  - [ ] Access removal
  - [ ] Cleanup and auditing

##### **5.3 Approval Management**
- [ ] **Task**: Implement approval signal handling
- [ ] **Task**: Add approval escalation (timeout → next approver)
- [ ] **Task**: Add approval delegation support
- [ ] **Task**: Add approval audit trail

---

### **Phase 6: API Integration** ⏱️ **Est: 3-4 days**

##### **6.1 HTTP API Endpoints**
- [ ] **Task**: Create ABAC API handlers
  - [ ] `POST /api/v1/abac/evaluate-permission`
  - [ ] `POST /api/v1/abac/bulk-evaluate`
  - [ ] `POST /api/v1/abac/request-access`
  - [ ] `GET /api/v1/abac/user/{id}/effective-permissions`
  - [ ] `POST /api/v1/abac/test-policy`
- [ ] **Task**: Add request validation and error handling
- [ ] **Task**: Add response formatting and HTTP status codes
- [ ] **Task**: Add API documentation (OpenAPI/Swagger)

##### **6.2 Temporal Client Integration**
- [ ] **Task**: Create HTTP → Temporal client bridge
- [ ] **Task**: Add workflow execution tracking
- [ ] **Task**: Add async response handling (for long-running workflows)
- [ ] **Task**: Add workflow cancellation endpoints

##### **6.3 Legacy API Migration**
- [ ] **Task**: Update existing identity API endpoints
- [ ] **Task**: Remove ABAC methods from identity handlers
- [ ] **Task**: Add deprecation notices for removed endpoints
- [ ] **Task**: Create migration guide for API consumers

---

### **Phase 7: Policy Management** ⏱️ **Est: 4-5 days**

##### **7.1 Policy CRUD Operations**
- [ ] **Task**: Implement `PolicyManagementWorkflow`
  - [ ] Policy creation with validation
  - [ ] Policy updates with impact analysis
  - [ ] Policy deletion with dependency checking
- [ ] **Task**: Add policy versioning support
- [ ] **Task**: Add policy testing and simulation
- [ ] **Task**: Add policy impact analysis

##### **7.2 Policy Management API**
- [ ] **Task**: Create policy management endpoints
  - [ ] `POST /api/v1/policies`
  - [ ] `PUT /api/v1/policies/{id}`
  - [ ] `DELETE /api/v1/policies/{id}`
  - [ ] `GET /api/v1/policies`
  - [ ] `POST /api/v1/policies/{id}/test`
- [ ] **Task**: Add policy validation rules
- [ ] **Task**: Add policy conflict detection
- [ ] **Task**: Add policy performance analysis

##### **7.3 Attribute Management**
- [ ] **Task**: Implement attribute definition management
- [ ] **Task**: Add attribute validation and type checking
- [ ] **Task**: Add attribute usage tracking
- [ ] **Task**: Add attribute migration tools

---

### **Phase 8: Testing & Quality Assurance** ⏱️ **Est: 5-6 days**

##### **8.1 Unit Testing**
- [ ] **Task**: Write activity unit tests
  - [ ] Attribute collection activities
  - [ ] Policy evaluation activities
  - [ ] Context enrichment activities
- [ ] **Task**: Write workflow unit tests
- [ ] **Task**: Write repository unit tests
- [ ] **Task**: Add test coverage reporting

##### **8.2 Integration Testing**
- [ ] **Task**: Write workflow integration tests
- [ ] **Task**: Write API endpoint integration tests
- [ ] **Task**: Write database integration tests
- [ ] **Task**: Add Temporal test server setup

##### **8.3 Performance Testing**
- [ ] **Task**: Add policy evaluation performance tests
- [ ] **Task**: Add bulk operation performance tests
- [ ] **Task**: Add cache performance validation
- [ ] **Task**: Add load testing for workflows

##### **8.4 End-to-End Testing**
- [ ] **Task**: Write complete ABAC scenario tests
- [ ] **Task**: Write access request workflow tests
- [ ] **Task**: Write approval process tests
- [ ] **Task**: Add chaos engineering tests

---

### **Phase 9: Monitoring & Observability** ⏱️ **Est: 3-4 days**

##### **9.1 Metrics and Monitoring**
- [ ] **Task**: Add workflow execution metrics
- [ ] **Task**: Add policy evaluation performance metrics
- [ ] **Task**: Add cache hit/miss metrics
- [ ] **Task**: Add error rate and latency metrics

##### **9.2 Logging and Tracing**
- [ ] **Task**: Add structured logging to activities
- [ ] **Task**: Add distributed tracing integration
- [ ] **Task**: Add audit logging for all ABAC decisions
- [ ] **Task**: Add compliance logging (GDPR, SOX, etc.)

##### **9.3 Alerting and Health Checks**
- [ ] **Task**: Add workflow health checks
- [ ] **Task**: Add policy evaluation SLA alerts
- [ ] **Task**: Add error rate alerts
- [ ] **Task**: Add Temporal cluster health monitoring

---

### **Phase 10: Documentation & Deployment** ⏱️ **Est: 3-4 days**

##### **10.1 Documentation**
- [ ] **Task**: Write ABAC system architecture documentation
- [ ] **Task**: Write workflow documentation
- [ ] **Task**: Write API documentation
- [ ] **Task**: Write deployment guide
- [ ] **Task**: Write troubleshooting guide

##### **10.2 Deployment Preparation**
- [ ] **Task**: Create Temporal cluster deployment configs
- [ ] **Task**: Create database migration scripts
- [ ] **Task**: Create monitoring and alerting configs
- [ ] **Task**: Create rollback procedures

##### **10.3 Migration Planning**
- [ ] **Task**: Create migration plan from old system
- [ ] **Task**: Create data migration scripts
- [ ] **Task**: Create feature flag configurations
- [ ] **Task**: Create rollback strategy

---

## 📊 Progress Tracking

### **Overall Progress: 22/143 tasks completed (15.38%)**

#### **Phase Completion Status:**
- **Phase 1**: 12/12 tasks (100%) - ✅ **COMPLETE**
- **Phase 2**: 10/11 tasks (90.9%) - ⏳ **In Progress**
- **Phase 3**: 0/20 tasks (0%) - ⏳ **Pending**
- **Phase 4**: 0/12 tasks (0%) - ⏳ **Pending**
- **Phase 5**: 0/18 tasks (0%) - ⏳ **Pending**
- **Phase 6**: 0/12 tasks (0%) - ⏳ **Pending**
- **Phase 7**: 0/15 tasks (0%) - ⏳ **Pending**
- **Phase 8**: 0/16 tasks (0%) - ⏳ **Pending**
- **Phase 9**: 0/12 tasks (0%) - ⏳ **Pending**
- **Phase 10**: 0/13 tasks (0%) - ⏳ **Pending**

### **Timeline Estimate:**
- **Total Estimated Duration**: 43-51 days
- **Target Completion**: 8-10 weeks (accounting for testing and refinement)

### **Current Sprint Focus:**
**Phase 2: Temporal Workflow Foundation** - Core ABAC Models (Validation and Serialization)

### **Phase 1 Status: ✅ COMPLETE (100%)**
**Major Achievements:**
- ✅ Identity service fully simplified and cleaned
- ✅ Complete ABAC module structure created
- ✅ Core workflows implemented (Permission evaluation, bulk evaluation)
- ✅ All ABAC activities implemented (Attribute collection, policy evaluation, caching)
- ✅ Temporal client wrapper with worker setup complete
- ✅ Comprehensive ABAC models defined
- ✅ All ABAC database tables successfully created and indexed

### **Phase 2.1 Status: ✅ COMPLETE (100%)**
**Major Achievements:**
- ✅ Temporal client wrapper created (`internal/core/abac/client/temporal_client.go`).
- ✅ Temporal client configured for connection, workflow/activity registration, and error handling.
- ✅ Integration with `tenant.Service` for database session tenant context management (SetTenant/ResetTenant) implemented.
- ✅ Integration with shared `errors`, `logger`, `metrics`, and `tracing` for robust observability.

### **Phase 2.3 Status: ✅ COMPLETE (100%)**
**Major Achievements:**
- ✅ SQL queries for Policy, Attribute Definition, and Policy Evaluation repositories created.
- ✅ SQLC generation for these repositories completed.
- ✅ Go interfaces for ABAC repositories defined.
- ✅ Concrete Go implementations for ABAC repositories created.
- ✅ Mocks for ABAC repository interfaces generated.



### **Phase 2.3 Status: ✅ COMPLETE (100%)**
**Major Achievements:**
- ✅ SQL queries for Policy, Attribute Definition, and Policy Evaluation repositories created.
- ✅ SQLC generation for these repositories completed.
- ✅ Go interfaces for ABAC repositories defined.
- ✅ Concrete Go implementations for ABAC repositories created.
- ✅ Mocks for ABAC repository interfaces generated.



---

## 🎯 Success Criteria

- [ ] **Functional**: All ABAC features from documentation are implemented
- [ ] **Performance**: Policy evaluation < 100ms for simple policies, < 500ms for complex
- [ ] **Scalability**: Handle 1000+ concurrent permission evaluations
- [ ] **Reliability**: 99.9% uptime for permission evaluation workflows
- [ ] **Security**: Complete audit trail for all access decisions
- [ ] **Maintainability**: Clean separation of concerns, comprehensive documentation

---

## 🔧 Development Environment Setup

### **Prerequisites:**
- [ ] Temporal server running locally/dev environment
- [ ] PostgreSQL with ABAC tables created
- [ ] Go development environment
- [ ] Docker for containerized services

### **Initial Setup Commands:**
```bash
# Run ABAC table migrations
make migrate-up

# Start Temporal server
temporal server start-dev

# Run ABAC workers
go run cmd/worker/main.go

# Run API server with ABAC endpoints
go run cmd/server/main.go
```

---

**Next Action**: Start Phase 3.1 - Implement Attribute Collection Activities.
