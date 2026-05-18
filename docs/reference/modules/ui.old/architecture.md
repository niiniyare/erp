# UI Architecture

AWO ERP uses a **Go-first typed UI compilation system**. Every page schema is
defined in Go, compiled through a typed AST, validated by the pipeline, and
served as AMIS-compatible JSON. No schema is ever written by hand in JSON.

---

## Core Principles

| Principle | Rule |
|-----------|------|
| Typed, not stringly-typed | Schemas are Go structs (`ast.Node`), not `map[string]any` |
| Pipeline-only compilation | PageFn is called exclusively inside `CompileStage` |
| Pre-resolved permissions | `UISessionContext.Can()` reads a pre-computed map — no Casbin at render time |
| Block ownership of permissions | Blocks decide what to render; callers must never gate before calling a block |
| Cache after auth | Cache key is built after `AuthzStage` — never before |
| Immutable nodes | AST nodes use value receivers; no mutation after construction |

---

## System Layers

```
HTTP Request
    │
    ▼
┌─────────────────────────────────────┐
│  SchemaHandler (api/handlers)       │  Parse route, build OperationContext
└────────────────┬────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────┐
│  UIPipeline (web/wire.go)           │  Inject CacheVersions, run pipeline
└────────────────┬────────────────────┘
                 │
    ┌────────────▼────────────┐
    │    Stage DAG             │  Ordered by Priority + DependsOn
    │                          │
    │  Session (10)            │  Authenticate
    │  Authz (20)              │  Bulk-resolve permissions
    │  CacheLookup (30)        │  Check Redis + in-process LRU
    │  Registry (40)           │  Resolve PageFn / ASTPageFn
    │  Compile (50)            │  Call ASTPageFn → ast.CompileTree
    │  Normalize (60)          │  Canonicalize (lowercase, trim) — never errors
    │  Validate (70)           │  Structural + security checks — BusinessError
    │  CacheStore (80)         │  Write compiled schema to cache
    │  Response (90)           │  Serialize to JSON
    └──────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────┐
│  AMIS Frontend                      │  Renders the schema
└─────────────────────────────────────┘
```

---

## Package Map

```
internal/web/
├── ast/          Typed node definitions (PageNode, CRUDNode, FormNode, …)
├── authz/        UIAuthzService — Casbin bulk-enforce + fingerprinting
├── cache/        InvalidationScope, CacheVersions, pattern helpers
├── dsl/
│   ├── blocks/   Reusable UI building blocks (27 files)
│   ├── screens/  Composed screen layouts — under 60 lines each (20 files)
│   └── builders/ Domain-specific node convenience constructors (5 files)
├── metrics/      Metric name constants + RegisterUIMetrics
├── registry/     PageRegistration, RegisterPage, ValidateRegistry
├── stages/       All 9 pipeline stages + InstrumentedStage wrapper
└── ui/           UISessionContext, PageFn, ASTPageFn, error sentinels
```

---

## Error Handling

Every error that crosses a layer boundary must be a `*sharedErrors.BusinessError`.

| Layer | Code prefix | Example |
|-------|-------------|---------|
| AST compilation | `AST_*` | `AST_REQUIRED_FIELD` |
| Pipeline validation | `VALIDATE_*` | `VALIDATE_API_METHOD_PREFIX` |
| Page registry | `REGISTRY_*` | `REGISTRY_MISSING_MODULE` |
| DSL construction | `DSL_*` | `DSL_INVALID_CONFIG` |
| Cache | `CACHE_*` | `CACHE_INVALIDATION_FAILED` |

`ui.ErrSchemaInvalid` is the sentinel used by `SchemaHandler` to detect any
validation failure. All `VALIDATE_*` errors are chained with
`.WithCause(ui.ErrSchemaInvalid)` so `errors.Is(err, ui.ErrSchemaInvalid)`
continues to work in the handler.

---

## Key Invariants

1. **NormalizeStage never returns an error.** It mutates silently (canonicalize
   component types to lowercase, trim API URL whitespace). If a value cannot be
   normalized it is left as-is. Rejection is ValidateStage's job.

2. **ValidateStage never mutates.** It reads the schema and returns
   `*sharedErrors.BusinessError` on violation. Structural checks skip when
   `DataKeyASTCompiled=true` because the typed AST already enforces them.

3. **ASTPageFn path bypasses most NormalizeStage checks.** When a page is
   migrated to `ASTPageFn`, `CompileStage` sets `DataKeyASTCompiled=true`.
   `ValidateStage` still runs security rules on all schemas.

4. **Cache key is built AFTER AuthzStage.** Building it before would mean two
   users with different permissions could share a schema — a privilege
   escalation defect.
