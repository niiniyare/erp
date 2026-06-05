# AwoERP UI Platform Documentation

> **Last Updated:** 2026-06-05
> **Note:** See `claude-review.md` for a full audit of accuracy issues in the original Vol I–II files.

---

## Quick Reference

The UI platform serves AMIS JSON schemas to the browser via `GET /schema/<route>`.
The browser's AMIS SDK renders them. Authorization is resolved by Casbin before compilation.
Compiled schemas are cached in Redis, keyed by route + tenant + permission fingerprint + feature flag fingerprint.

**Adding a new page:** Write a screen in `internal/web/dsl/screens/`, register it with `registry.RegisterPage()` from an `init()` function.

**Key files:**
- `internal/web/ui/pipeline.go` — stage priorities and data key constants
- `internal/web/ui/types.go` — `UISessionContext`, `PageFn`, `ASTPageFn`
- `internal/web/ast/node.go` — `Node`, `ContainerNode` interfaces
- `internal/web/registry/registry.go` — `RegisterPage`, `Match`
- `internal/web/handler/schema.go` — HTTP handler for `/schema/*`
- `internal/web/dsl/blocks/` — reusable DSL blocks
- `internal/web/dsl/screens/` — complete page functions

---

## Document Index

### Volume I — Vision and Philosophy (Rewritten — Accurate)

| File | Chapter | Accuracy |
|------|---------|----------|
| `vol-01-vision/01-introduction.md` | 01 — Introduction | AMIS-first, Casbin, correct feature list, planned items marked |
| `vol-01-vision/02-design-philosophy.md` | 02 — Design Philosophy | Correct trade-offs, actual code patterns |
| `vol-01-vision/03-architectural-principles.md` | 03 — Architectural Principles | Correct pipeline ordering, correct separation of concerns |
| `vol-01-vision/04-system-overview.md` | 04 — System Overview | Actual request lifecycle, real data keys, real components |

### Volume II — DSL and AST (Rewritten — Accurate)

| File | Chapter | Accuracy |
|------|---------|----------|
| `vol-02-dsl-and-ast/05-sdui-fundamentals.md` | 05 — SDUI Fundamentals | AMIS envelope, UISessionContext, correct expression syntax |
| `vol-02-dsl-and-ast/06-ui-dsl-architecture.md` | 06 — UI DSL Architecture | Blocks, screens, builders, registry init() pattern |
| `vol-02-dsl-and-ast/07-ast-design.md` | 07 — AST Design | Struct literals, all node types, CompileTree, validation errors |
| `vol-02-dsl-and-ast/08-compilation-pipeline.md` | 08 — Compilation Pipeline | All 9 stages with correct priorities, cache key design, error envelopes |

### Volume III — Component System (New)

| File | Chapter | Accuracy |
|------|---------|----------|
| `vol-03-component-system/09-component-system.md` | 09 — Component System | AMIS types, two builder paths, common mistakes |
| `vol-03-component-system/10-layout-system.md` | 10 — Layout System | PageNode, GridNode, TabsNode (MountOnEnter), FlexNode, SectionNode |
| `vol-03-component-system/11-forms-framework.md` | 11 — Forms Framework | FormNode, field types, field modifiers, filter bars, wizards |
| `vol-03-component-system/12-tables-and-data-grids.md` | 12 — Tables and Data Grids | CRUDNode, DataTableBlock, status badges, syncLocation enforcement |

### Audit

| File | Description |
|------|-------------|
| `claude-review.md` | Structured audit of original docs/ui/01–04 files: all factual errors, invented architecture, non-compiling code |

---

## Original Files (Superseded — Contain Errors)

The following files remain at `docs/ui/` root for reference but have been superseded by the `vol-01-vision/` rewrites:

| File | Key Errors |
|------|-----------|
| `01-introduction.md` | iOS/Android claimed as implemented; OpenFGA instead of Casbin; "surface/SurfaceID" vocab |
| `02-design-philosophy.md` | Non-existent constructor examples |
| `03-architectural-principles.md` | Cross-references to invented pipeline stages |
| `04-system-overview.md` | 6-stage pipeline (real: 9); CompilationContext struct does not exist; gRPC for UI; wrong JSON envelope |

---

## Pipeline Quick Reference

| Priority | Stage | Cache hit: skip? |
|----------|-------|-----------------|
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
| Backend / Go engineer | [Ch 06 DSL Architecture](./vol-02-dsl-and-ast/06-ui-dsl-architecture.md) then [Ch 07 AST](./vol-02-dsl-and-ast/07-ast-design.md) |
| Wants to add a new page | [Ch 06 §6.7 Registering a Page](./vol-02-dsl-and-ast/06-ui-dsl-architecture.md#67-registering-a-new-page) |
| Debugging a pipeline error | [Ch 08 Compilation Pipeline](./vol-02-dsl-and-ast/08-compilation-pipeline.md) |
| Building a data table | [Ch 12 Tables and Data Grids](./vol-03-component-system/12-tables-and-data-grids.md) |
| Building a form | [Ch 11 Forms Framework](./vol-03-component-system/11-forms-framework.md) |
| Solution architect | [Ch 03 Architectural Principles](./vol-01-vision/03-architectural-principles.md) |
