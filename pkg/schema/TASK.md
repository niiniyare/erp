# Schema Engine Implementation Tasks

## Overview

This document outlines the remaining tasks to complete the Schema Engine implementation. The Schema Engine is a JSON-driven UI framework that allows backend developers to define complete forms using only JSON schemas, with server-side rendering, enterprise features, and 40+ field types.

## Project Status

✅ **Completed:**
- Core schema structures (schema.go, field.go, action.go)
- Advanced features (mixin.go, business_rules.go, repeatable.go)
- Enterprise components (security, multi-tenancy, workflow)
- Design tokens system (tokens.go)
- Comprehensive validation and error handling
- Basic Registry structure
- Parser implementation (moved to parse/ directory)
- **Enricher Implementation (November 2025)**
  - Complete DefaultEnricher with permission-based field control
  - Tenant customization and dynamic default value injection
  - User context integration with runtime state management
  - 13 comprehensive test methods covering all enrichment scenarios
- **Test Coverage Improvement: 49.0% → 59.0% (November 2025)**
  - Fixed all failing test cases for stable foundation
  - Added schema utility method tests with proper signatures
  - Comprehensive enterprise feature tests (Security, Tenant, Workflow, I18n, HTMX, Alpine, Meta)
  - Business rules TestRule, ExplainRule, UpdateRule coverage
  - 300+ new test cases with proper error handling
- **Documentation Alignment (November 2025)**
  - Updated 11-enricher.md to match actual implementation
  - Revised 15-runtime.md for architecture consistency
  - Rewrote 16-I18n.md to follow schema-driven patterns
  - Fixed field structure inconsistencies (Default vs DefaultValue)
  - Aligned all code examples with pkg/schema implementation

## Phase Structure

Each phase contains specific tasks that must be completed in order. Tasks are marked with checkboxes for progress tracking.

---

## Phase 1: Core Processing Components (High Priority)

**Goal:** Complete the remaining core processing components as defined in the architecture.

### Task 1.1: Validator Component Enhancement
**Status:** 🟡 In Progress  
**Dependencies:** None  
**Estimate:** 1.5 hours

**Description:**
Enhance the existing validator component to integrate with the implemented enricher and support advanced validation scenarios.

**What to do:**
1. Review and update `pkg/schema/validate/validator.go`
2. Integrate with enricher's runtime field state (Visible, Editable)
3. Add support for conditional validation based on field visibility
4. Enhance business rules integration with enriched context
5. Add comprehensive validation for enriched schemas

**Expected Output:**
```go
// pkg/schema/validate/validator.go enhancements
func (v *Validator) ValidateEnrichedSchema(ctx context.Context, schema *schema.Schema, data map[string]any) (*ValidationResult, error)
func (v *Validator) ValidateWithFieldState(field *schema.Field, value any) []ValidationError
func (v *Validator) ValidateConditionalFields(schema *schema.Schema, data map[string]any) []ValidationError
```

**Testing:**
```go
func TestValidator_EnrichedSchemaValidation(t *testing.T) {
    // Test validation of enriched schemas with runtime state
    // Test conditional validation based on field visibility
    // Test integration with enricher permissions
    // Test tenant-specific validation rules
}
```

**Success Criteria:**
- Validates enriched schemas with runtime field state
- Integrates with enricher's permission system
- Supports conditional validation scenarios
- Maintains backward compatibility with existing validation
- 100% test coverage for new validation features

**Commit Message:**
```
enhance validator integration with enricher component

- Add support for enriched schema validation
- Integrate with runtime field state (Visible, Editable)
- Support conditional validation based on permissions
- Maintain backward compatibility with existing validation
- Add comprehensive test coverage
```

---

### Task 1.2: Registry Storage Backend Implementation
**Status:** 🔴 Not Started  
**Dependencies:** Task 1.1  
**Estimate:** 3 hours

**Description:**
Implement the storage backend interfaces for PostgreSQL, Redis, Filesystem, and S3/Minio as specified in the storage interface documentation.

**What to do:**
1. Create storage implementations in `pkg/schema/registry/`:
   - `postgres.go` - PostgreSQL storage with ACID guarantees
   - `redis.go` - Redis cache storage with TTL
   - `filesystem.go` - File-based storage for development
   - `s3.go` - S3/Minio cloud storage
2. Follow the Storage interface exactly as defined
3. Include proper error handling and connection management
4. Add SQL schema for PostgreSQL with multi-tenant support
5. Integrate with enricher for tenant-aware storage

**Expected Output:**
```go
// Storage interface implementation for each backend
type PostgresStorage struct { db *sql.DB }
type RedisStorage struct { client *redis.Client; ttl time.Duration }
type FilesystemStorage struct { basePath string }
type S3Storage struct { client *minio.Client; bucket string }

// Each implements:
func (s *Storage) Get(ctx context.Context, id string) ([]byte, error)
func (s *Storage) Set(ctx context.Context, id string, data []byte) error
func (s *Storage) Delete(ctx context.Context, id string) error
func (s *Storage) List(ctx context.Context) ([]string, error)
func (s *Storage) Exists(ctx context.Context, id string) (bool, error)
```

**Testing:**
```go
func TestStorageBackends(t *testing.T) {
    // Test each storage backend independently
    // Test storage interface compliance
    // Test error handling and connection failures
    // Test tenant isolation in storage
    // Test integration with enricher
}
```

**Success Criteria:**
- All storage backends implement Storage interface correctly
- PostgreSQL includes multi-tenant row-level security
- Redis includes configurable TTL and caching
- Filesystem includes proper file management
- S3 includes proper bucket operations
- Integration with enricher for tenant-aware operations

**Commit Message:**
```
implement storage backends for schema registry

- Add PostgreSQL storage with multi-tenant RLS support
- Add Redis cache storage with configurable TTL
- Add filesystem storage for development environment
- Add S3/Minio storage for cloud deployments
- Integrate with enricher for tenant-aware operations
```

---

### Task 1.3: Runtime Component Implementation
**Status:** 🔴 Not Started  
**Dependencies:** Task 1.1, Task 1.2  
**Estimate:** 2 hours

**Description:**
Implement the runtime component that manages schema execution state and integrates with the enricher for dynamic behavior.

**What to do:**
1. Create `pkg/schema/runtime/runtime.go`
2. Implement state management for form interactions
3. Integrate with enricher for permission-aware operations
4. Add event handling for field changes and validation
5. Support for conditional field behavior based on enriched state

**Expected Output:**
```go
// pkg/schema/runtime/runtime.go
type Runtime struct {
    schema    *schema.Schema       // Enriched schema
    state     *State              // Current form state
    validator *validate.Validator  // Validator integration
    events    *EventHandler       // Event handling
}

func NewRuntime(enrichedSchema *schema.Schema) *Runtime
func (r *Runtime) Initialize(ctx context.Context, data map[string]interface{}) error
func (r *Runtime) HandleFieldChange(ctx context.Context, field string, value interface{}) error
func (r *Runtime) ValidateCurrentState(ctx context.Context) map[string][]string
```

**Testing:**
```go
func TestRuntime_Integration(t *testing.T) {
    // Test runtime with enriched schemas
    // Test state management with permission controls
    // Test event handling for field changes
    // Test validation integration
    // Test conditional field behavior
}
```

**Success Criteria:**
- Integrates seamlessly with enriched schemas
- Manages form state with permission awareness
- Handles events with proper validation
- Supports conditional field behavior
- Provides comprehensive error handling

**Commit Message:**
```
implement runtime component with enricher integration

- Add form state management with permission awareness
- Integrate with enriched schemas for dynamic behavior
- Support event handling and real-time validation
- Include conditional field behavior based on enriched state
- Add comprehensive error handling and testing
```

---

## Phase 2: Rendering System (Medium Priority)

**Goal:** Implement the templ-based rendering system that converts enriched schemas to HTML.

### Task 2.1: Enhanced templ Templates
**Status:** 🔴 Not Started  
**Dependencies:** Phase 1 complete  
**Estimate:** 2.5 hours

**Description:**
Create templ templates that work with enriched schemas and support all runtime features including permissions and tenant customizations.

**What to do:**
1. Create `views/` directory structure with enricher-aware templates
2. Support for runtime field state (Visible, Editable, Reason)
3. Include tenant customization rendering
4. Add permission-based conditional rendering
5. Integrate with I18n structure from enriched schemas

**Expected Output:**
- Templates that respect field.Runtime.Visible and field.Runtime.Editable
- Support for tenant-specific customizations
- I18n-aware rendering using field.I18n structure
- HTMX integration for dynamic field behavior
- Alpine.js for client-side permission handling

**Testing:**
```go
func TestEnrichedTemplateRendering(t *testing.T) {
    // Test templates with enriched schemas
    // Test permission-based field hiding
    // Test tenant customization rendering
    // Test I18n integration
    // Test runtime state handling
}
```

**Success Criteria:**
- Templates work with enriched schema structure
- Runtime permissions are properly enforced
- Tenant customizations render correctly
- I18n support is fully functional
- HTMX and Alpine.js integration works

**Commit Message:**
```
implement enricher-aware templ templates

- Support runtime field state (Visible, Editable, Reason)
- Include tenant customization and I18n rendering
- Add permission-based conditional rendering
- Integrate HTMX for dynamic behavior
- Add Alpine.js for client-side permission handling
```

---

### Task 2.2: Enhanced Renderer Component
**Status:** 🔴 Not Started  
**Dependencies:** Task 2.1  
**Estimate:** 1.5 hours

**Description:**
Update the renderer component to work with enriched schemas and support all runtime features.

**What to do:**
1. Update `pkg/schema/render/renderer.go` for enriched schema support
2. Add permission-aware field rendering
3. Include tenant customization in rendering logic
4. Support for I18n field rendering
5. Integration with runtime component

**Expected Output:**
```go
// pkg/schema/render/renderer.go enhancements
func (r *Renderer) RenderEnrichedForm(schema *schema.Schema, runtime *runtime.Runtime) (string, error)
func (r *Renderer) RenderFieldWithPermissions(field *schema.Field, value any) (string, error)
func (r *Renderer) RenderWithTenantCustomization(schema *schema.Schema, tenantID string) (string, error)
```

**Testing:**
```go
func TestRenderer_EnrichedSchema(t *testing.T) {
    // Test rendering with enriched schemas
    // Test permission-based field exclusion
    // Test tenant customization rendering
    // Test I18n field rendering
    // Test runtime integration
}
```

**Success Criteria:**
- Renders enriched schemas correctly
- Respects runtime permission states
- Applies tenant customizations
- Supports I18n field rendering
- Integrates with runtime component

**Commit Message:**
```
enhance renderer for enriched schema support

- Add permission-aware field rendering
- Support tenant customization in rendering
- Include I18n field rendering capabilities
- Integrate with runtime component
- Ensure backward compatibility
```

---

## Phase 3: Integration & Testing (Medium Priority)

**Goal:** Integrate all components and ensure comprehensive testing coverage.

### Task 3.1: Complete Integration Tests
**Status:** 🔴 Not Started  
**Dependencies:** Phase 1 & 2 complete  
**Estimate:** 2 hours

**Description:**
Create end-to-end integration tests that verify the complete enriched schema processing pipeline.

**What to do:**
1. Create comprehensive integration tests for the enriched pipeline
2. Test Registry → Parser → Enricher → Validator → Runtime → Renderer
3. Include real-world scenarios with permissions and tenants
4. Test error propagation through the enriched pipeline
5. Performance benchmarks for enriched schema processing

**Expected Output:**
```go
func TestCompleteEnrichedSchemaProcessing(t *testing.T) {
    // Test: JSON schema → Enriched schema → HTML form
    // Test: Permission-based field control
    // Test: Tenant customization application
    // Test: Multi-user scenarios
    // Test: Error handling with enriched context
}

func BenchmarkEnrichedSchemaProcessing(b *testing.B) {
    // Benchmark complete enriched pipeline performance
}
```

**Testing:**
- End-to-end enriched processing pipeline
- Permission and tenant isolation
- Error handling with enriched context
- Performance under load with enrichment
- Memory usage with enriched schemas

**Success Criteria:**
- Complete enriched pipeline processes correctly
- Permission and tenant isolation verified
- Performance meets requirements with enrichment
- Error handling works throughout enriched pipeline
- Memory usage remains acceptable

**Commit Message:**
```
add comprehensive integration tests for enriched schema processing

- Test complete pipeline with enricher integration
- Verify permission-based field control
- Test tenant customization and isolation
- Include performance benchmarks for enriched processing
- Ensure error handling throughout enriched pipeline
```

---

### Task 3.2: Documentation and Examples Update
**Status:** 🔴 Not Started  
**Dependencies:** Phase 1 & 2 complete  
**Estimate:** 1 hour

**Description:**
Update documentation and examples to reflect the complete enriched schema implementation.

**What to do:**
1. Update example schemas to show enricher features
2. Add enricher usage examples to documentation
3. Document runtime component integration
4. Update API documentation for all components
5. Add deployment guide for enriched schema system

**Expected Output:**
- Example schemas showing permission controls and tenant customization
- Enricher usage examples and best practices
- Runtime component documentation
- Complete API documentation
- Production deployment guide

**Testing:**
```go
func TestDocumentationExamples(t *testing.T) {
    // Test all documentation examples compile and run
    // Test enricher examples work correctly
    // Test runtime examples function properly
}
```

**Success Criteria:**
- All examples demonstrate enricher features
- Documentation covers complete implementation
- Code examples compile and run correctly
- Deployment guide is comprehensive
- API documentation is complete

**Commit Message:**
```
update documentation and examples for enriched schema system

- Add enricher usage examples and best practices
- Document runtime component integration
- Update API documentation for all components
- Include production deployment guide
- Ensure all code examples work correctly
```

---

## Phase 4: Optimization & Production Readiness (Low Priority)

**Goal:** Optimize performance and prepare for production deployment of the enriched system.

### Task 4.1: Performance Optimization for Enriched Pipeline
**Status:** 🔴 Not Started  
**Dependencies:** Phase 1-3 complete  
**Estimate:** 2 hours

**Description:**
Optimize the enriched schema processing pipeline for production performance.

**What to do:**
1. Optimize enricher performance with caching
2. Add connection pooling for storage backends
3. Implement enriched schema compilation
4. Add metrics for enricher operations
5. Optimize memory usage in enriched processing

**Expected Output:**
- Enricher caching for improved performance
- Connection pooling for storage operations
- Compiled enriched schema support
- Prometheus metrics for enricher operations
- Memory optimization in enriched pipeline

**Testing:**
```go
func BenchmarkOptimizedEnrichedProcessing(b *testing.B) {
    // Benchmark optimized enriched pipeline
    // Memory allocation benchmarks
    // Concurrency stress tests with enrichment
}
```

**Success Criteria:**
- 50% performance improvement in enriched processing
- Reduced memory allocations in enricher
- Connection pooling reduces storage load
- Metrics provide visibility into enricher performance
- Graceful degradation under high load

**Commit Message:**
```
optimize enriched schema processing for production

- Add enricher caching for improved performance
- Implement connection pooling for storage backends
- Add compiled enriched schema support
- Include Prometheus metrics for enricher operations
- Optimize memory usage in enriched processing pipeline
```

---

### Task 4.2: Security Hardening for Enriched System
**Status:** 🔴 Not Started  
**Dependencies:** Phase 1-3 complete  
**Estimate:** 1.5 hours

**Description:**
Implement additional security measures for the enriched schema system.

**What to do:**
1. Security review of enricher permission logic
2. Add audit logging for enricher operations
3. Implement rate limiting for enriched processing
4. Add input sanitization for enriched data
5. Security review of tenant isolation in enricher

**Expected Output:**
- Security-hardened enricher permission system
- Audit logging for enricher operations
- Rate limiting for enriched processing
- Input sanitization for enriched data
- Verified tenant isolation in enricher

**Testing:**
```go
func TestEnrichedSystemSecurity(t *testing.T) {
    // Test enricher permission security
    // Test tenant isolation security
    // Test audit logging completeness
    // Test rate limiting effectiveness
}
```

**Success Criteria:**
- Enricher permission logic is security-hardened
- Audit logs capture all enricher operations
- Rate limiting protects enriched processing
- Input sanitization prevents attacks
- Tenant isolation is cryptographically verified

**Commit Message:**
```
implement security hardening for enriched schema system

- Security-harden enricher permission logic
- Add comprehensive audit logging for enricher operations
- Implement rate limiting for enriched processing
- Add input sanitization for enriched data
- Verify tenant isolation security in enricher
```

---

## Summary

### Total Estimated Time: 16 hours

**Phase 1 (High Priority):** 6.5 hours
- Validator enhancement: 1.5h
- Storage backends: 3h
- Runtime component: 2h

**Phase 2 (Medium Priority):** 4 hours
- Enhanced templates: 2.5h
- Enhanced renderer: 1.5h

**Phase 3 (Medium Priority):** 3 hours
- Integration tests: 2h
- Documentation update: 1h

**Phase 4 (Low Priority):** 3.5 hours
- Performance optimization: 2h
- Security hardening: 1.5h

### Success Metrics

1. **Functionality:** All enricher features working correctly with 40+ field types
2. **Performance:** <100ms for enriched schema processing (cached)
3. **Security:** Complete permission system + tenant isolation + security hardening
4. **Reliability:** 100% test coverage for enriched pipeline
5. **Documentation:** Complete API docs and enricher examples

### Next Steps

1. **Immediate:** Complete Phase 1 validator enhancement
2. **Week 1:** Complete Phase 1 & start Phase 2
3. **Week 2:** Complete Phase 2 & Phase 3
4. **Week 3:** Complete Phase 4 & production readiness

### Dependencies

- **External:** condition package for business rules
- **Internal:** Completed enricher implementation (✅ Done)
- **Infrastructure:** PostgreSQL, Redis for storage backends
- **Frontend:** HTMX 1.9+, Alpine.js 3.x for progressive enhancement

### Recent Completions (November 2025)

✅ **Enricher Implementation Complete**
- DefaultEnricher with permission-based field control
- Tenant customization and dynamic defaults
- User context integration with runtime state
- 13 comprehensive test methods

✅ **Documentation Alignment Complete**  
- Updated all architecture docs to match implementation
- Fixed field structure inconsistencies
- Aligned code examples with actual implementation
- Comprehensive enricher documentation

---

**Document Version:** 2.0  
**Last Updated:** 2025-11-03  
**Status:** Updated with enricher completion and documentation alignment