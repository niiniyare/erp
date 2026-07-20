> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

# AwoERP UI Platform Documentation

> **Last Updated:** 2026-06-05
> **Status:** Stable — all 51 chapters written (implemented chapters full, planned chapters stubbed)
> **Audit:** `claude-review.md` — full accuracy audit of original Vol I–II files

---

## Quick Reference

The UI platform serves AMIS JSON schemas to the browser via `GET /schema/<route>`.
The browser's AMIS SDK renders them. Authorization resolved by Casbin before compilation.
Compiled schemas cached in Redis, keyed by route + tenant + permission fingerprint + feature flag fingerprint.

**Add a new page:** Write screen in `internal/web/dsl/screens/`, call `registry.RegisterPage()` from `init()`.

**Key files:**
- `internal/web/ui/pipeline.go` — stage priorities and data key constants
- `internal/web/ui/types.go` — `UISessionContext`, `PageFn`, `ASTPageFn`
- `internal/web/ast/node.go` — `Node`, `ContainerNode` interfaces
- `internal/web/registry/registry.go` — `RegisterPage`, `Match`
- `internal/web/handler/schema.go` — HTTP handler for `/schema/*`
- `internal/web/dsl/blocks/` — reusable DSL blocks
- `internal/web/dsl/screens/` — complete page functions

---

## Pipeline Quick Reference

| Priority | Stage | Skip on cache hit? |
|----------|-------|--------------------|
| 10 | SessionStage — `contract.FromContext` | No |
| 20 | AuthzStage — Casbin BulkEnforce, build UISessionContext, compute fingerprints | No |
| 30 | CacheStage — Redis lookup by route+tenant+perm_fp+flag_fp | No |
| 40 | RegistryStage — `registry.Match(route)` | Yes |
| 50 | CompileStage — `ASTPageFn(sess)` + `ast.CompileTree()` or `PageFn(sess)` | Yes |
| 60 | NormalizeStage — lowercase types, trim whitespace | Yes |
| 70 | ValidateStage — structural + security rules | Yes |
| 80 | CacheStoreStage — write to Redis | Yes |
| 90 | ResponseStage — `{"status": 0, "data": schema}` | No |

---

## Role-Based Reading Guide

| Role | Start Here |
|------|-----------|
| New to the platform | [Ch 01 Introduction](./vol-01-vision/01-introduction.md) |
| Backend / Go engineer | [Ch 06 DSL Architecture](./vol-02-dsl-and-ast/06-ui-dsl-architecture.md) → [Ch 07 AST](./vol-02-dsl-and-ast/07-ast-design.md) |
| Adding a new page | [Ch 06](./vol-02-dsl-and-ast/06-ui-dsl-architecture.md) → [Ch 49 Examples](./vol-10-reference/49-end-to-end-examples.md) |
| Debugging pipeline error | [Ch 08 Pipeline](./vol-02-dsl-and-ast/08-compilation-pipeline.md) |
| Building a data table | [Ch 12 Tables](./vol-03-component-system/12-tables-and-data-grids.md) |
| Building a form | [Ch 11 Forms](./vol-03-component-system/11-forms-framework.md) |
| Understanding security | [Ch 26 Security](./vol-06-platform/26-security-model.md) → [Ch 27 Authz](./vol-06-platform/27-authorization-integration.md) |
| Understanding caching | [Ch 34 Caching](./vol-07-reliability/34-caching-strategies.md) |
| Testing page functions | [Ch 41 Testing](./vol-09-engineering/41-testing-strategy.md) |
| Avoiding mistakes | [Ch 47 Anti-Patterns](./vol-10-reference/47-anti-patterns.md) |
| Solution architect | [Ch 03 Principles](./vol-01-vision/03-architectural-principles.md) → [Ch 48 Reference Arch](./vol-10-reference/48-reference-architecture.md) |

---

## Full Chapter Index

### Volume I — Vision and Philosophy
| Ch | File | Status |
|----|------|--------|
| 01 | [Introduction](./vol-01-vision/01-introduction.md) | ✅ Full |
| 02 | [Design Philosophy](./vol-01-vision/02-design-philosophy.md) | ✅ Full |
| 03 | [Architectural Principles](./vol-01-vision/03-architectural-principles.md) | ✅ Full |
| 04 | [System Overview](./vol-01-vision/04-system-overview.md) | ✅ Full |

### Volume II — DSL and AST
| Ch | File | Status |
|----|------|--------|
| 05 | [SDUI Fundamentals](./vol-02-dsl-and-ast/05-sdui-fundamentals.md) | ✅ Full |
| 06 | [UI DSL Architecture](./vol-02-dsl-and-ast/06-ui-dsl-architecture.md) | ✅ Full |
| 07 | [AST Design](./vol-02-dsl-and-ast/07-ast-design.md) | ✅ Full |
| 08 | [Compilation Pipeline](./vol-02-dsl-and-ast/08-compilation-pipeline.md) | ✅ Full |

### Volume III — Component System
| Ch | File | Status |
|----|------|--------|
| 09 | [Component System](./vol-03-component-system/09-component-system.md) | ✅ Full |
| 10 | [Layout System](./vol-03-component-system/10-layout-system.md) | ✅ Full |
| 11 | [Forms Framework](./vol-03-component-system/11-forms-framework.md) | ✅ Full |
| 12 | [Tables and Data Grids](./vol-03-component-system/12-tables-and-data-grids.md) | ✅ Full |
| 13 | [Dashboard Framework](./vol-03-component-system/13-dashboard-framework.md) | ✅ Full |
| 14 | [Charts and Analytics](./vol-03-component-system/14-charts-and-analytics.md) | ✅ Full |
| 15 | [Workflow and Approval Components](./vol-03-component-system/15-workflow-and-approval-components.md) | ✅ Full |
| 16 | [ERP-Specific Components](./vol-03-component-system/16-erp-specific-components.md) | ✅ Full |

### Volume IV — Rendering Architecture
| Ch | File | Status |
|----|------|--------|
| 17 | [Mobile Rendering Architecture](./vol-04-rendering/17-mobile-rendering-architecture.md) | 🔲 Planned |
| 18 | [Web Rendering Architecture](./vol-04-rendering/18-web-rendering-architecture.md) | ✅ Full |
| 19 | [Navigation Framework](./vol-04-rendering/19-navigation-framework.md) | ✅ Partial |
| 20 | [Action System](./vol-04-rendering/20-action-system.md) | ✅ Full |
| 21 | [Event System](./vol-04-rendering/21-event-system.md) | ✅ Partial |

### Volume V — Runtime and State
| Ch | File | Status |
|----|------|--------|
| 22 | [State Management](./vol-05-runtime/22-state-management.md) | ✅ Partial |
| 23 | [Validation Framework](./vol-05-runtime/23-validation-framework.md) | ✅ Full |
| 24 | [Data Sources](./vol-05-runtime/24-data-sources.md) | ✅ Full |

### Volume VI — Platform and API
| Ch | File | Status |
|----|------|--------|
| 25 | [API Contracts](./vol-06-platform/25-api-contracts.md) | ✅ Full |
| 26 | [Security Model](./vol-06-platform/26-security-model.md) | ✅ Full |
| 27 | [Authorization Integration](./vol-06-platform/27-authorization-integration.md) | ✅ Full |
| 28 | [Multi-Tenant Architecture](./vol-06-platform/28-multi-tenant-architecture.md) | ✅ Full |
| 29 | [Feature Flag Architecture](./vol-06-platform/29-feature-flag-architecture.md) | ✅ Full |
| 30 | [Customization Framework](./vol-06-platform/30-customization-framework.md) | ✅ Partial |
| 31 | [Extension and Plugin Architecture](./vol-06-platform/31-extension-and-plugin-architecture.md) | 🔲 Planned |

### Volume VII — Reliability and Performance
| Ch | File | Status |
|----|------|--------|
| 32 | [Performance Optimization](./vol-07-reliability/32-performance-optimization.md) | ✅ Full |
| 33 | [Offline Support](./vol-07-reliability/33-offline-support.md) | 🔲 Planned |
| 34 | [Caching Strategies](./vol-07-reliability/34-caching-strategies.md) | ✅ Full |
| 35 | [Synchronization](./vol-07-reliability/35-synchronization.md) | ✅ Partial |

### Volume VIII — Developer Experience
| Ch | File | Status |
|----|------|--------|
| 36 | [Internationalization](./vol-08-dx/36-internationalization.md) | ✅ Partial |
| 37 | [Accessibility](./vol-08-dx/37-accessibility.md) | 🔲 Planned |
| 38 | [Observability](./vol-08-dx/38-observability.md) | ✅ Full |
| 39 | [Telemetry](./vol-08-dx/39-telemetry.md) | ✅ Full |
| 40 | [Logging](./vol-08-dx/40-logging.md) | ✅ Full |

### Volume IX — Engineering Excellence
| Ch | File | Status |
|----|------|--------|
| 41 | [Testing Strategy](./vol-09-engineering/41-testing-strategy.md) | ✅ Full |
| 42 | [CI/CD Considerations](./vol-09-engineering/42-cicd-considerations.md) | ✅ Full |
| 43 | [Versioning Strategy](./vol-09-engineering/43-versioning-strategy.md) | ✅ Full |
| 44 | [Migration Strategy](./vol-09-engineering/44-migration-strategy.md) | ✅ Full |
| 45 | [Governance Model](./vol-09-engineering/45-governance-model.md) | ✅ Full |

### Volume X — Reference and Guidance
| Ch | File | Status |
|----|------|--------|
| 46 | [Best Practices](./vol-10-reference/46-best-practices.md) | ✅ Full |
| 47 | [Anti-Patterns](./vol-10-reference/47-anti-patterns.md) | ✅ Full |
| 48 | [Reference Architecture](./vol-10-reference/48-reference-architecture.md) | ✅ Full |
| 49 | [End-to-End Examples](./vol-10-reference/49-end-to-end-examples.md) | ✅ Full |
| 50 | [SDK Development](./vol-10-reference/50-sdk-development.md) | 🔲 Planned |
| 51 | [Future Roadmap](./vol-10-reference/51-future-roadmap.md) | ✅ Full |

---

## Status Key

| Symbol | Meaning |
|--------|---------|
| ✅ Full | Implemented — based on actual code, no invented architecture |
| ✅ Partial | Implemented sections full; planned sections stubbed with callouts |
| 🔲 Planned | Feature not yet implemented; chapter describes intent and contracts |

---

## Audit and Superseded Files

| File | Description |
|------|-------------|
| `claude-review.md` | Accuracy audit — all errors in original docs/ui/01–08 files |
| `01-introduction.md` (root) | Superseded — iOS/Android as implemented, OpenFGA, wrong vocab |
| `02-design-philosophy.md` (root) | Superseded — non-existent constructor examples |
| `03-architectural-principles.md` (root) | Superseded — invented pipeline stages |
| `04-system-overview.md` (root) | Superseded — 6-stage pipeline, CompilationContext, gRPC, wrong envelope |
| `05–08.md` (root) | Superseded — see vol-02-dsl-and-ast/ for accurate versions |
