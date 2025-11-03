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
- **Test Coverage Improvement: 49.0% → 59.0% (November 2025)**
  - Fixed all failing test cases for stable foundation
  - Added schema utility method tests with proper signatures
  - Comprehensive enterprise feature tests (Security, Tenant, Workflow, I18n, HTMX, Alpine, Meta)
  - Business rules TestRule, ExplainRule, UpdateRule coverage
  - 300+ new test cases with proper error handling

## Phase Structure

Each phase contains specific tasks that must be completed in order. Tasks are marked with checkboxes for progress tracking.

---

## Phase 1: Core Processing Components (High Priority)

**Goal:** Implement the missing core processing components as defined in the architecture.

### Task 1.1: Fix Directory Structure
**Status:** 🟡 In Progress  
**Dependencies:** None  
**Estimate:** 30 minutes

**Description:**
Reorganize the processing components into the correct directory structure as specified in the architecture documentation.

**What to do:**
1. Ensure proper package structure:
   ```
   pkg/schema/
   ├── parse/          # Parser (JSON → Go)
   ├── validate/       # Validator (rules)
   ├── enrich/         # Enricher (permissions)
   └── registry/       # Registry (storage + cache)
   ```
2. Update import statements in all files
3. Verify package declarations match directory names

**Expected Output:**
- [x] `pkg/schema/parse/parser.go` - JSON to Go struct conversion
- [ ] `pkg/schema/validate/validator.go` - Server-side validation
- [ ] `pkg/schema/enrich/enricher.go` - Runtime permissions and context
- [ ] `pkg/schema/registry/` - Storage implementations

**Testing:**
```go
// Test all packages import correctly
import (
    "github.com/niiniyare/erp/pkg/schema/parse"
    "github.com/niiniyare/erp/pkg/schema/validate"
    "github.com/niiniyare/erp/pkg/schema/enrich"
    "github.com/niiniyare/erp/pkg/schema/registry"
)
```

**Success Criteria:**
- All packages compile without import errors
- Package names match directory structure
- No circular dependencies

**Commit Message:**
```
reorganize schema processing components into correct directory structure

- Move parser to pkg/schema/parse/
- Create validate, enrich, registry directories
- Update package declarations and imports
- Follow architecture specification from docs/schema/02-architecture.md
```

---

### Task 1.2: Implement Validator Component
**Status:** 🔴 Not Started  
**Dependencies:** Task 1.1  
**Estimate:** 2 hours

**Description:**
Create the server-side validation component that validates form data against schema rules, including business rules integration with the condition package.

**What to do:**
1. Create `pkg/schema/validate/validator.go`
2. Implement HTML5 validation rules
3. Implement server-side business rules
4. Integrate with condition package for cross-field validation
5. Support for uniqueness checks (database queries)

**Expected Output:**
```go
// pkg/schema/validate/validator.go
type Validator struct {
    db Database // For uniqueness checks
}

func (v *Validator) ValidateData(schema *schema.Schema, data map[string]any) (*ValidationResult, error)
func (v *Validator) ValidateField(field *schema.Field, value any) []ValidationError
func (v *Validator) ValidateBusinessRules(schema *schema.Schema, data map[string]any) []ValidationError
```

**Testing:**
```go
func TestValidator_ValidateData(t *testing.T) {
    // Test required field validation
    // Test field type validation
    // Test min/max length validation
    // Test email format validation
    // Test business rules validation
    // Test uniqueness validation
}
```

**Success Criteria:**
- Validates all field types correctly
- Integrates with business rules engine
- Supports database uniqueness checks
- Returns structured validation errors
- 100% test coverage for validation logic

**Commit Message:**
```
implement server-side validation component

- Add comprehensive field validation (required, type, constraints)
- Integrate business rules with condition package
- Support database uniqueness checks
- Include structured error reporting
- Add complete test suite with 100% coverage
```

---

### Task 1.3: Implement Enricher Component
**Status:** 🔴 Not Started  
**Dependencies:** Task 1.1  
**Estimate:** 1.5 hours

**Description:**
Create the enricher component that adds runtime permissions, applies tenant customization, populates default values, and injects user context into schemas.

**What to do:**
1. Create `pkg/schema/enrich/enricher.go`
2. Implement permission-based field visibility
3. Apply tenant-specific customizations
4. Populate default values based on user context
5. Set runtime properties on fields

**Expected Output:**
```go
// pkg/schema/enrich/enricher.go
type Enricher struct {
    permissionService PermissionService
    tenantService     TenantService
}

func (e *Enricher) Enrich(ctx context.Context, schema *schema.Schema, user *User) (*schema.Schema, error)
func (e *Enricher) ApplyPermissions(schema *schema.Schema, permissions []string) error
func (e *Enricher) ApplyTenantCustomization(schema *schema.Schema, tenantID string) error
func (e *Enricher) PopulateDefaults(schema *schema.Schema, user *User) error
```

**Testing:**
```go
func TestEnricher_Enrich(t *testing.T) {
    // Test permission-based field hiding
    // Test tenant customization application
    // Test default value population
    // Test user context injection
    // Test runtime property setting
}
```

**Success Criteria:**
- Fields are hidden based on user permissions
- Tenant customizations are applied correctly
- Default values are populated from user context
- Runtime properties are set appropriately
- No mutations to original schema (returns copy)

**Commit Message:**
```
implement schema enricher with permissions and tenant support

- Add permission-based field visibility control
- Apply tenant-specific schema customizations
- Populate defaults from user context
- Set runtime properties for field behavior
- Return enriched schema copy without mutations
```

---

### Task 1.4: Implement Storage Backends
**Status:** 🔴 Not Started  
**Dependencies:** Task 1.1  
**Estimate:** 4 hours

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
    // Test error handling
    // Test connection failures
    // Test data persistence
}
```

**Success Criteria:**
- All storage backends implement Storage interface correctly
- PostgreSQL includes multi-tenant row-level security
- Redis includes configurable TTL
- Filesystem includes proper file management
- S3 includes proper bucket operations
- Comprehensive error handling for all failure modes

**Commit Message:**
```
implement storage backends for schema registry

- Add PostgreSQL storage with multi-tenant RLS support
- Add Redis cache storage with configurable TTL
- Add filesystem storage for development environment
- Add S3/Minio storage for cloud deployments
- Include comprehensive error handling and testing
```

---

## Phase 2: Rendering System (Medium Priority)

**Goal:** Implement the templ-based rendering system that converts schemas to HTML.

### Task 2.1: Create Base templ Templates
**Status:** 🔴 Not Started  
**Dependencies:** Phase 1 complete  
**Estimate:** 3 hours

**Description:**
Create the base templ templates for rendering forms and individual field types. These templates should follow the design system guidelines and include proper HTMX/Alpine.js integration.

**What to do:**
1. Create `views/` directory structure:
   ```
   views/
   ├── form.templ          # Main form template
   ├── layout.templ        # Layout wrapper
   └── fields/             # Field-specific templates
       ├── text.templ      # Text input
       ├── select.templ    # Select dropdown
       ├── textarea.templ  # Textarea
       ├── checkbox.templ  # Checkbox
       ├── radio.templ     # Radio buttons
       └── ...             # All 40+ field types
   ```
2. Follow design system tokens and architecture
3. Include proper ARIA attributes for accessibility
4. Add HTMX attributes for form submission
5. Include Alpine.js for client-side interactivity

**Expected Output:**
- Complete set of field templates covering all 40+ field types
- Main form template that orchestrates field rendering
- Layout template with proper semantic HTML structure
- WCAG 2.1 AA compliant markup
- HTMX integration for form submission
- Alpine.js integration for client-side state

**Testing:**
```go
func TestTemplateRendering(t *testing.T) {
    // Test each field type renders correctly
    // Test form template orchestration
    // Test HTMX attributes are present
    // Test Alpine.js attributes are correct
    // Test accessibility compliance
}
```

**Success Criteria:**
- All 40+ field types have corresponding templates
- Templates follow design system guidelines
- WCAG 2.1 AA compliant markup
- HTMX and Alpine.js integration working
- Templates compile without errors

**Commit Message:**
```
implement base templ templates for schema rendering

- Add form and layout templates with semantic HTML
- Create field templates for all 40+ field types
- Include HTMX attributes for progressive enhancement
- Add Alpine.js for client-side interactivity
- Ensure WCAG 2.1 AA accessibility compliance
```

---

### Task 2.2: Implement Renderer Component
**Status:** 🔴 Not Started  
**Dependencies:** Task 2.1  
**Estimate:** 2 hours

**Description:**
Create the renderer component that maps schema fields to templ templates and generates final HTML output.

**What to do:**
1. Create `pkg/schema/render/renderer.go`
2. Implement field type to template mapping
3. Add layout composition logic
4. Include design token injection
5. Support for theme variants

**Expected Output:**
```go
// pkg/schema/render/renderer.go
type Renderer struct {
    templates map[string]*template.Template
    tokens    *DesignTokens
}

func (r *Renderer) RenderForm(schema *schema.Schema, data map[string]any) (string, error)
func (r *Renderer) RenderField(field *schema.Field, value any) (string, error)
func (r *Renderer) RenderLayout(content string, layout *schema.Layout) (string, error)
```

**Testing:**
```go
func TestRenderer_RenderForm(t *testing.T) {
    // Test complete form rendering
    // Test field rendering with different types
    // Test layout composition
    // Test design token injection
    // Test theme variant support
}
```

**Success Criteria:**
- Correctly maps all field types to templates
- Generates valid HTML output
- Includes design tokens in CSS variables
- Supports theme switching
- Handles rendering errors gracefully

**Commit Message:**
```
implement schema renderer with templ template integration

- Add field type to template mapping system
- Include layout composition and design token injection
- Support theme variants and CSS variable generation
- Add comprehensive error handling for render failures
- Ensure type-safe template rendering with templ
```

---

## Phase 3: Integration & Testing (Medium Priority)

**Goal:** Integrate all components and ensure comprehensive testing coverage.

### Task 3.1: Create Integration Tests
**Status:** 🔴 Not Started  
**Dependencies:** Phase 1 & 2 complete  
**Estimate:** 2 hours

**Description:**
Create end-to-end integration tests that verify the complete schema processing pipeline from JSON to HTML.

**What to do:**
1. Create `pkg/schema/integration_test.go`
2. Test complete pipeline: Registry → Parser → Enricher → Validator → Renderer
3. Include real-world schema examples
4. Test error propagation through the pipeline
5. Performance benchmarks for the complete flow

**Expected Output:**
```go
func TestCompleteSchemaProcessing(t *testing.T) {
    // Test: JSON schema → HTML form (complete pipeline)
    // Test: Form submission validation
    // Test: Permission-based field hiding
    // Test: Multi-tenant isolation
    // Test: Error handling at each stage
}

func BenchmarkSchemaProcessing(b *testing.B) {
    // Benchmark complete pipeline performance
}
```

**Testing:**
- End-to-end processing pipeline
- Error handling and recovery
- Performance under load
- Memory usage patterns
- Multi-tenant data isolation

**Success Criteria:**
- Complete pipeline processes schemas correctly
- All error cases are handled gracefully
- Performance meets requirements (<100ms for cached schemas)
- Memory usage is within acceptable limits
- Multi-tenant isolation is verified

**Commit Message:**
```
add comprehensive integration tests for schema processing

- Test complete pipeline from JSON schema to HTML rendering
- Include error handling and performance benchmarks
- Verify multi-tenant isolation and security
- Add real-world schema examples for testing
- Ensure sub-100ms performance for cached schemas
```

---

### Task 3.2: Add Example Schemas and Documentation
**Status:** 🔴 Not Started  
**Dependencies:** Phase 1 & 2 complete  
**Estimate:** 1.5 hours

**Description:**
Create comprehensive example schemas and update documentation to reflect the complete implementation.

**What to do:**
1. Create `examples/` directory with real-world schema examples
2. Update README.md with current implementation status
3. Create API documentation for all components
4. Add usage examples for each component
5. Document deployment and configuration

**Expected Output:**
- Complete set of example schemas (user registration, invoice, product catalog, etc.)
- Updated README with implementation status
- API documentation for all public interfaces
- Usage examples and code snippets
- Deployment and configuration guide

**Testing:**
```go
func TestExampleSchemas(t *testing.T) {
    // Test all example schemas process correctly
    // Test examples match documentation
    // Test code snippets in docs compile
}
```

**Success Criteria:**
- Example schemas cover all major use cases
- Documentation is up-to-date and accurate
- Code examples compile and run
- Deployment guide is complete
- API documentation covers all public interfaces

**Commit Message:**
```
add comprehensive examples and documentation

- Create real-world schema examples for common use cases
- Update README with current implementation status
- Add complete API documentation for all components
- Include usage examples and deployment guide
- Ensure all code examples compile and run correctly
```

---

## Phase 4: Optimization & Production Readiness (Low Priority)

**Goal:** Optimize performance and prepare for production deployment.

### Task 4.1: Performance Optimization
**Status:** 🔴 Not Started  
**Dependencies:** Phase 1-3 complete  
**Estimate:** 2 hours

**Description:**
Optimize the schema processing pipeline for production performance and memory usage.

**What to do:**
1. Implement schema compilation (pre-parse and cache compiled schemas)
2. Add connection pooling for storage backends
3. Optimize memory allocations in hot paths
4. Add metrics and monitoring hooks
5. Implement graceful degradation

**Expected Output:**
- Schema compilation system for faster runtime processing
- Connection pooling for database backends
- Memory optimization in critical paths
- Prometheus metrics integration
- Circuit breaker patterns for external dependencies

**Testing:**
```go
func BenchmarkOptimizedProcessing(b *testing.B) {
    // Benchmark optimized vs. unoptimized performance
    // Memory allocation benchmarks
    // Concurrency stress tests
}
```

**Success Criteria:**
- 50% performance improvement in schema processing
- Reduced memory allocations in hot paths
- Connection pooling reduces database load
- Metrics provide visibility into system health
- Graceful degradation under high load

**Commit Message:**
```
optimize schema processing performance for production

- Add schema compilation for faster runtime processing
- Implement connection pooling for storage backends
- Optimize memory allocations in critical paths
- Add Prometheus metrics and monitoring hooks
- Include circuit breaker patterns for resilience
```

---

### Task 4.2: Security Hardening
**Status:** 🔴 Not Started  
**Dependencies:** Phase 1-3 complete  
**Estimate:** 1.5 hours

**Description:**
Implement additional security measures and conduct security review of the implementation.

**What to do:**
1. Add input sanitization for all user inputs
2. Implement rate limiting at the component level
3. Add audit logging for sensitive operations
4. Security review of SQL queries (injection prevention)
5. Add CSRF protection for form submissions

**Expected Output:**
- Comprehensive input sanitization
- Component-level rate limiting
- Audit logging for schema operations
- SQL injection prevention verification
- CSRF token integration

**Testing:**
```go
func TestSecurityMeasures(t *testing.T) {
    // Test input sanitization
    // Test SQL injection prevention
    // Test rate limiting
    // Test audit logging
    // Test CSRF protection
}
```

**Success Criteria:**
- All user inputs are properly sanitized
- SQL injection attacks are prevented
- Rate limiting protects against abuse
- Audit logs capture sensitive operations
- CSRF protection is properly implemented

**Commit Message:**
```
implement security hardening for production deployment

- Add comprehensive input sanitization
- Implement component-level rate limiting
- Add audit logging for sensitive operations
- Verify SQL injection prevention measures
- Include CSRF protection for form submissions
```

---

## Summary

### Total Estimated Time: 20 hours

**Phase 1 (High Priority):** 8 hours
- Directory structure: 0.5h
- Validator component: 2h
- Enricher component: 1.5h
- Storage backends: 4h

**Phase 2 (Medium Priority):** 5 hours
- Base templates: 3h
- Renderer component: 2h

**Phase 3 (Medium Priority):** 3.5 hours
- Integration tests: 2h
- Examples & docs: 1.5h

**Phase 4 (Low Priority):** 3.5 hours
- Performance optimization: 2h
- Security hardening: 1.5h

### Success Metrics

1. **Functionality:** All 40+ field types render correctly
2. **Performance:** <100ms for cached schema processing
3. **Security:** WCAG 2.1 AA compliance + security hardening
4. **Reliability:** 100% test coverage for core components
5. **Documentation:** Complete API docs and examples

### Next Steps

1. **Immediate:** Complete Phase 1 tasks in order
2. **Week 1:** Complete Phase 1 & start Phase 2
3. **Week 2:** Complete Phase 2 & Phase 3
4. **Week 3:** Complete Phase 4 & production readiness

### Dependencies

- **External:** condition package for business rules
- **Internal:** Proper Go module structure
- **Infrastructure:** PostgreSQL, Redis for storage backends
- **Frontend:** HTMX 1.9+, Alpine.js 3.x for progressive enhancement

---

**Document Version:** 1.0  
**Last Updated:** 2025-11-02  
**Status:** Ready for implementation