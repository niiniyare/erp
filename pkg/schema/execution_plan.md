# Schema Theme & Token System Implementation Plan

**Status:** Implementation Phase  
**Target:** Replace legacy theme/token system with three-tier architecture  
**Timeline:** 8 developer-days across 3 phases  
**Testing Coverage:** ≥95% with comprehensive unit testing strategy

## 🎯 Implementation Overview

Replace `pkg/schema/theme.go` and `pkg/schema/tokens.go` with modern three-tier token architecture:
- **Primitives** → Raw design values (`hsl(220, 50%, 50%)`)
- **Semantic** → Functional assignments (`{colors.primary.base}`)
- **Components** → Component-specific styles (`{semantic.interactive.default}`)

## 📋 Phase Timeline & Testing Strategy

### Phase 1: Core Foundation (3 days)
**Deliverables:**
- [ ] Token structure implementation
- [ ] TokenReference resolution system
- [ ] Basic ThemeManager with registry

**Unit Testing Focus:**
```go
// Critical test coverage areas
✓ TokenReference.IsReference() edge cases
✓ TokenReference.Path() extraction logic
✓ Circular reference detection
✓ Token resolution with missing references
✓ Thread-safe registry operations
✓ Token merging with override precedence
```

### Phase 2: Integration & Extensions (3 days)
**Deliverables:**
- [ ] Multi-tenant theme management
- [ ] Context-aware theme resolution
- [ ] Integration with existing modules

**Unit Testing Focus:**
```go
// Integration & context testing
✓ Context theme resolution accuracy
✓ Tenant isolation verification
✓ Cache invalidation correctness
✓ Theme inheritance behavior
✓ Performance under load (benchmarks)
```

### Phase 3: Migration & Validation (2 days)
**Deliverables:**
- [ ] Legacy compatibility layer
- [ ] Migration utilities
- [ ] Production validation

**Unit Testing Focus:**
```go
// Migration & validation testing
✓ Legacy theme conversion accuracy
✓ Migration rollback safety
✓ Error handling completeness
✓ Memory leak prevention
✓ Concurrent access safety
```

## 🔗 Integration Touchpoints

| Module | Integration Point | Test Requirements |
|--------|------------------|-------------------|
| **builder** | `Theme` field in Schema | Theme application correctness |
| **parser** | Token resolution during parse | Token reference validation |
| **runtime** | Dynamic theme switching | Runtime consistency checks |
| **enricher** | Theme-aware enrichment | Theme inheritance verification |
| **registry** | Theme storage/retrieval | Registry consistency validation |
| **validate** | Theme validation rules | Validation accuracy testing |

## 🧪 Testing Architecture

### Core Test Suites
- **Unit Tests** (95%+ coverage): `*_test.go` files
- **Integration Tests**: Cross-module theme behavior
- **Benchmark Tests**: Performance regression prevention
- **Property Tests**: Edge case discovery via fuzzing

### Critical Test Scenarios
```go
// High-priority test cases
func TestTokenReferenceResolution(t *testing.T)     // Core functionality
func TestCircularReferenceDetection(t *testing.T)   // Prevents infinite loops
func TestConcurrentThemeAccess(t *testing.T)        // Thread safety
func TestThemeInheritancePrecedence(t *testing.T)   // Override behavior
func TestMigrationDataIntegrity(t *testing.T)       // Migration safety
```

## 🚀 Success Criteria

**Functional Requirements:**
- [ ] 100% backward compatibility during migration
- [ ] Context-aware multi-tenant theme resolution
- [ ] Sub-10ms token resolution performance
- [ ] Memory-efficient token caching

**Quality Requirements:**
- [ ] ≥95% unit test coverage
- [ ] Zero memory leaks under load
- [ ] Complete godoc documentation
- [ ] Integration test coverage for all touchpoints

**Migration Requirements:**
- [ ] Zero-downtime deployment capability
- [ ] Rollback safety mechanisms
- [ ] Legacy theme conversion utilities
- [ ] Performance parity with existing system

---

**Implementation Notes:**
- Use table-driven tests for token resolution scenarios
- Implement property-based testing for TokenReference edge cases
- Benchmark critical paths (token resolution, theme switching)
- Mock integration points for isolated unit testing
- Validate WCAG compliance in accessibility configurations