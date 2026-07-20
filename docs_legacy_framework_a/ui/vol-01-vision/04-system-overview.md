> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

# Chapter 04 — System Overview

> **Volume:** I — Vision & Philosophy
> **Audience:** All
> **Prerequisites:** Chapters 01–03

---

## Table of Contents

- [4.1 High-Level System Diagram](#41-high-level-system-diagram)
- [4.2 Platform Components Inventory](#42-platform-components-inventory)
- [4.3 Request Lifecycle: A Concrete Example](#43-request-lifecycle-a-concrete-example)
- [4.4 Technology Stack](#44-technology-stack)
- [4.5 Deployment Topology](#45-deployment-topology)
- [4.6 Scalability Characteristics](#46-scalability-characteristics)
- [4.7 Known Limitations](#47-known-limitations)

---

## 4.1 High-Level System Diagram

```
┌──────────────────────────────────────────┐
│            Browser (AMIS SDK)            │
│  web/sdk/sdk.js + sdk.css                │
│  Renders {"status":0,"data":<schema>}    │
└─────────────────┬────────────────────────┘
                  │  GET /schema/<route>
                  │  Authorization: Bearer <jwt>
┌─────────────────▼────────────────────────┐
│    AwoERP API (Fiber HTTP Server)        │
│                                          │
│  Middleware chain (per /schema/* route): │
│  1. Authenticate (validates JWT)         │
│  2. InjectSessionContext                 │
│  3. SchemaHandler.Handle                 │
└─────────────────┬────────────────────────┘
                  │
┌─────────────────▼────────────────────────┐
│         9-Stage UI Pipeline              │
│                                          │
│  10 SessionStage                         │
│     contract.FromContext → SessionCtx    │
│                                          │
│  20 AuthzStage                           │
│     UIAuthzService.BulkEnforce (Casbin)  │
│     → UISessionContext                   │
│     → perm_fingerprint, flag_fingerprint │
│                                          │
│  30 CacheStage ──────────────────────────┼──► Redis
│     key = route+tenant+perm_fp+flag_fp   │    (hit → skip 40–80)
│                                          │
│  40 RegistryStage                        │
│     registry.Match(route)                │
│     → PageFn or ASTPageFn               │
│                                          │
│  50 CompileStage                         │
│     ASTPageFn(sess) → ast.CompileTree()  │
│     or PageFn(sess) → Schema             │
│                                          │
│  60 NormalizeStage                       │
│     lowercase types, trim API whitespace │
│                                          │
│  70 ValidateStage                        │
│     structural + security rules          │
│                                          │
│  80 CacheStoreStage ─────────────────────┼──► Redis
│     write compiled schema                │    (write-through)
│                                          │
│  90 ResponseStage (always runs)          │
│     {"status":0,"data":<schema>}         │
└──────────────────────────────────────────┘
```

---

## 4.2 Platform Components Inventory

### 4.2.1 SchemaHandler

**Location:** `internal/web/handler/schema.go`
**Mounted at:** `/schema/*`
**Responsibility:** Accept HTTP GET requests, extract the route from the wildcard path, build an `OperationContext`, run the pipeline, return the AMIS JSON envelope.

Required middleware chain (in order):
1. `iam/middleware.Authenticate` — validates JWT, sets Fiber Locals
2. `contract.InjectSessionContext` — bridges Fiber Locals to Go context
3. `SchemaHandler.Handle` — runs pipeline, returns response

The handler never reads Fiber Locals directly. All identity data arrives via `contract.FromContext(c.UserContext())`.

### 4.2.2 UI Pipeline

**Location:** `internal/web/ui/pipeline.go` (constants and data keys)
**Responsibility:** 9-stage ordered pipeline. Stages communicate via `opCtx.Data[DataKey*]` constants. The pipeline is read-only and non-transactional — no DB writes, no compensation logic, no TxHooks.

Data key constants:

| Constant | Value | Set By |
|----------|-------|--------|
| `DataKeyRoute` | `"ui.request.route"` | Handler before Run() |
| `DataKeyPermissions` | `"ui.authz.permissions"` | AuthzStage |
| `DataKeyPermFingerprint` | `"ui.authz.perm_fingerprint"` | AuthzStage |
| `DataKeyFlagFingerprint` | `"ui.authz.flag_fingerprint"` | AuthzStage |
| `DataKeyCacheKey` | `"ui.cache.key"` | CacheStage |
| `DataKeyCacheHit` | `"ui.cache.hit"` | CacheStage |
| `DataKeyPageFn` | `"ui.registry.page_fn"` | RegistryStage |
| `DataKeyASTPageFn` | `"ui.registry.ast_page_fn"` | RegistryStage |
| `DataKeyRouteParams` | `"ui.registry.route_params"` | RegistryStage |
| `DataKeyASTCompiled` | `"ui.compile.ast_compiled"` | CompileStage |
| `DataKeyCacheVersions` | `"ui.cache.versions"` | Injected at startup |
| `DataKeySessionCtx` | `"ui.authz.session_ctx"` | AuthzStage |
| `DataKeySchema` | `"ui.compile.schema"` | CompileStage |
| `DataKeyResponse` | `"ui.response"` | ResponseStage |

### 4.2.3 Page Registry

**Location:** `internal/web/registry/registry.go`
**Responsibility:** Maps URL routes to `PageFn` or `ASTPageFn`. Supports both exact routes and parameterized patterns (`/finance/invoices/:id`). Thread-safe via `sync.RWMutex`. Panics on duplicate registration at startup.

Registration happens in `init()` functions:

```go
func init() {
    registry.RegisterPage(registry.PageRegistration{
        Route:       "/finance/dashboard",
        Module:      "finance",
        Title:       "Finance Dashboard",
        Description: "KPI cards, revenue chart, quick actions",
        ASTFn:       screens.FinanceDashboardScreen,
    })
}
```

`ValidateRegistry()` is called at startup after all `init()` functions run. It panics if any registration is missing required fields.

### 4.2.4 AST Package

**Location:** `internal/web/ast/`
**Responsibility:** Typed node definitions and `CompileTree`. Each node implements `Node` (or `ContainerNode`). `CompileTree` validates all nodes depth-first, collects all errors, and only emits JSON if the entire tree is valid.

### 4.2.5 AMIS Builder Package

**Location:** `internal/web/amis/`
**Responsibility:** Fluent Go builders for AMIS JSON schemas. Used by legacy `PageFn` pages. Builders implement `json.Marshaler` — no `.Build()` call needed. Provided builders: `PageBuilder`, `ServiceBuilder`, `GridBuilder`, `PanelBuilder`, `TabsBuilder`, `StatBuilder`, `ChartBuilder`, `AlertBuilder`, `TimelineBuilder`, `DescriptionsBuilder`, `CRUDBuilder`, `ColumnBuilder`, `FormBuilder`, `WizardBuilder`.

### 4.2.6 DSL Blocks and Screens

**Location:** `internal/web/dsl/blocks/`, `internal/web/dsl/screens/`, `internal/web/dsl/builders/`
**Responsibility:** The primary API for page authors. Blocks are reusable fragments; screens are complete page functions; builders are module-domain helpers.

### 4.2.7 Redis Cache

**Responsibility:** Stores compiled AMIS schemas keyed by `route + tenant + perm_fingerprint + flag_fingerprint`. A cache hit skips stages 40–80. Cache entries are invalidated by permission changes (new fingerprint) or explicit flush.

---

## 4.3 Request Lifecycle: A Concrete Example

**Scenario:** A finance manager opens the Finance Dashboard.

### Step 1 — Browser Requests Schema

```
GET /schema/finance/dashboard
Authorization: Bearer <jwt>
```

Fiber matches `/schema/*`. The middleware chain runs: JWT is validated, `SessionContext` is injected into the Go context.

### Step 2 — Handler Builds OperationContext

`SchemaHandler.Handle` extracts the session via `contract.FromContext(c.UserContext())`. It sets:

```go
opCtx.Input = ui.UISchemaInput{Route: "/finance/dashboard"}
opCtx.OperationKey = ui.OperationKey  // "ui.schema.compile"
```

### Step 3 — SessionStage (Priority 10)

Reads `contract.SessionContext` from `opCtx.Ctx`. Verifies it is present and non-zero. If missing, returns `ErrUnauthenticated`.

### Step 4 — AuthzStage (Priority 20)

Calls `UIAuthzService.BulkEnforce(userID, tenantID, allUIPermissions)` via Casbin. Builds `UISessionContext` with the resolved `map[string]bool`. Computes `perm_fingerprint` (SHA-256 of sorted permission keys+values) and `flag_fingerprint`. Stores `UISessionContext` in `opCtx.Data[DataKeySessionCtx]`.

### Step 5 — CacheStage (Priority 30)

Computes `cacheKey = "/finance/dashboard" + tenantID + perm_fingerprint + flag_fingerprint`. Queries Redis. On a hit, reads the stored schema into `opCtx.Data[DataKeySchema]`, sets `DataKeyCacheHit = true`, and the pipeline skips stages 40–80.

On a miss, continues.

### Step 6 — RegistryStage (Priority 40)

Calls `registry.Match("/finance/dashboard")`. Finds the registration with `ASTFn: screens.FinanceDashboardScreen`. Stores `ASTFn` in `opCtx.Data[DataKeyASTPageFn]`. No URL params for this route, so `DataKeyRouteParams` is an empty map.

### Step 7 — CompileStage (Priority 50)

Retrieves `UISessionContext` from `DataKeySessionCtx`. Calls `FinanceDashboardScreen(sess)`, which returns an `ast.PageNode`. Calls `ast.CompileTree(root)`. On success, stores the compiled `map[string]any` in `DataKeySchema`. Sets `DataKeyASTCompiled = true`.

The compiled schema looks like:

```json
{
  "type": "page",
  "title": "Finance Dashboard",
  "initApi": {
    "method": "get",
    "url": "/api/v1/finance/dashboard/summary"
  },
  "body": [
    {
      "type": "cards",
      "items": [
        {"type": "stat", "label": "Total Revenue", "value": "${revenue}"},
        {"type": "stat", "label": "Outstanding AR", "value": "${ar_balance}"}
      ]
    }
  ]
}
```

### Step 8 — NormalizeStage (Priority 60)

Canonicalizes type values to lowercase, trims whitespace from API URLs. Since `DataKeyASTCompiled` is true, skips structural AMIS rules (already guaranteed by the typed AST).

### Step 9 — ValidateStage (Priority 70)

Checks security rules: no IAM expression strings in the schema. Checks structural rules that the typed AST does not guarantee. For this dashboard, all checks pass.

### Step 10 — CacheStoreStage (Priority 80)

Writes the compiled schema to Redis with the computed cache key.

### Step 11 — ResponseStage (Priority 90)

Assembles the AMIS envelope and stores it in `DataKeyResponse`:

```json
{
  "status": 0,
  "data": { <compiled schema> }
}
```

### Step 12 — Handler Returns Response

The handler reads `opCtx.Data[DataKeyResponse]` and returns it with HTTP 200. The browser's AMIS SDK receives the envelope, reads `data`, and renders the Finance Dashboard.

---

## 4.4 Technology Stack

| Layer | Technology | Role |
|-------|-----------|------|
| HTTP server | Go + Fiber | Routes, middleware, handler |
| UI pipeline | Go (custom pipeline framework) | 9-stage compilation |
| Authorization | Casbin | `UIAuthzService.BulkEnforce` |
| Schema cache | Redis | Compiled schema storage, key by fingerprint |
| AST compilation | Go (`internal/web/ast`) | Typed node graph → AMIS JSON |
| AMIS rendering | AMIS SDK (`web/sdk/`) | Browser JSON-to-UI renderer |
| Session identity | `contract.SessionContext` | IAM boundary; no raw JWT parsing in UI layer |

---

## 4.5 Deployment Topology

The UI pipeline runs inside the main AwoERP API server process. There is no separate UI compilation service. The `/schema/*` routes are registered alongside business API routes on the same Fiber server.

Redis is a separate stateful service. The cache client is initialized at startup and injected into `CacheStoreStage` and `CacheStage` via dependency injection.

The AMIS browser SDK (`web/sdk/sdk.js`, `web/sdk/sdk.css`) is served as static files. The main page shell (`web/pages/index.html`) bootstraps the AMIS renderer and makes schema requests for each navigation event.

---

## 4.6 Scalability Characteristics

**Cache efficiency is the primary performance lever.** A cold compilation (cache miss) takes approximately 20–50ms, dominated by the Casbin `BulkEnforce` call. A warm compilation (cache hit) takes approximately 2–5ms (Redis round-trip only).

For a tenant with users sharing the same permission set, the cache hit rate approaches 100% after the first user requests each page. Users with unique per-record permission configurations produce distinct fingerprints and lower hit rates.

**The API server scales horizontally.** The pipeline is stateless. All state lives in Redis and the Casbin policy store (PostgreSQL). Adding API server replicas increases throughput linearly.

---

## 4.7 Known Limitations

**No real-time schema push.** Schema updates reach users on the next page navigation that produces a cache miss. Active sessions do not receive updates.

**No mobile client.** Only the AMIS web SDK is implemented. Flutter and React Native clients are on the roadmap.

> **Status: Planned — Not Yet Implemented**

**No compression in the pipeline.** The pipeline returns uncompressed JSON. Fiber-level gzip middleware may be applied globally but is not part of the UI pipeline's responsibility.

**No parameterized edit screens.** The registry supports `:id` patterns, but no edit screens (e.g., `/finance/invoices/:id`) are currently registered.

**No ETag or conditional request support.** Every schema request hits the pipeline. Browser-level caching via `Cache-Control` headers is not implemented.

**No tenant form customization UI.** Tenant-specific overrides must be encoded in page functions, not configured by an admin. A no-code customization layer is planned.

---

*End of Chapter 04*

**Previous:** [Chapter 03 — Architectural Principles](./03-architectural-principles.md)
**Next:** [Chapter 05 — SDUI Fundamentals](../vol-02-dsl-and-ast/05-sdui-fundamentals.md)
