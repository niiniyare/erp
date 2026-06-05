# Chapter 08 — The Compilation Pipeline

> **Volume:** II — DSL and AST
> **Audience:** Platform Engineers, Backend Engineers
> **Prerequisites:** Chapter 07 — AST Design

---

## Table of Contents

- [8.1 Pipeline Overview](#81-pipeline-overview)
- [8.2 Stage 10 — SessionStage](#82-stage-10--sessionstage)
- [8.3 Stage 20 — AuthzStage](#83-stage-20--authzstage)
- [8.4 Stage 30 — CacheStage](#84-stage-30--cachestage)
- [8.5 Stage 40 — RegistryStage](#85-stage-40--registrystage)
- [8.6 Stage 50 — CompileStage](#86-stage-50--compilestage)
- [8.7 Stage 60 — NormalizeStage](#87-stage-60--normalizestage)
- [8.8 Stage 70 — ValidateStage](#88-stage-70--validatestage)
- [8.9 Stage 80 — CacheStoreStage](#89-stage-80--cachestorestage)
- [8.10 Stage 90 — ResponseStage](#810-stage-90--responsestage)
- [8.11 Data Flow Between Stages](#811-data-flow-between-stages)
- [8.12 Error Handling](#812-error-handling)
- [8.13 Cache Key Design](#813-cache-key-design)
- [8.14 Performance Characteristics](#814-performance-characteristics)

---

## 8.1 Pipeline Overview

The UI pipeline is a sequential, ordered pipeline of 9 stages. Each stage has a priority number that determines execution order. Stages communicate via `opCtx.Data`, a `map[string]any` keyed by `DataKey*` constants defined in `internal/web/ui/pipeline.go`.

The pipeline is **read-only and non-transactional**. It performs no database writes, no compensating actions, and no TxHooks. The only write is to Redis (CacheStoreStage).

```
Priority  Stage              Always Runs?
   10     SessionStage       Yes
   20     AuthzStage         Yes
   30     CacheStage         Yes
   40     RegistryStage      No — skipped on cache hit
   50     CompileStage       No — skipped on cache hit
   60     NormalizeStage     No — skipped on cache hit
   70     ValidateStage      No — skipped on cache hit
   80     CacheStoreStage    No — skipped on cache hit
   90     ResponseStage      Yes — always runs, cache hit or miss
```

Two operation keys:
- `ui.OperationKey = "ui.schema.compile"` — serves a single page schema
- `ui.AppOperationKey = "ui.app.compile"` — serves the app shell (nav tree)

---

## 8.2 Stage 10 — SessionStage

**Priority:** 10
**Always runs:** Yes

Extracts the IAM contract session from `opCtx.Ctx` using `contract.FromContext()`. Verifies that the session is present and non-zero.

**On success:** Stores session in stage-local state; passes to AuthzStage via the shared context.
**On failure:** Returns `ui.ErrUnauthenticated`. The handler returns the 401 AMIS envelope.

This stage is why `SchemaHandler.Handle` must be preceded by `contract.InjectSessionContext` middleware. Without that middleware, `contract.FromContext` returns a zero value and this stage fails.

---

## 8.3 Stage 20 — AuthzStage

**Priority:** 20
**Always runs:** Yes

This is the most important stage for correctness and cache efficiency.

**What it does:**

1. Calls `UIAuthzService.BulkEnforce(userID, tenantID, allUIPermissions)`. This is a single Casbin batch call that resolves all permissions the UI layer might need, returning a `map[string]bool`.

2. Constructs `UISessionContext` via `ui.NewUISessionContext(sc, permissions)`. This copies feature flags from the contract session into the struct.

3. Computes `perm_fingerprint`: SHA-256 hash of the sorted permission map keys and values. Two users with identical permissions produce identical fingerprints.

4. Computes `flag_fingerprint`: SHA-256 hash of the sorted feature flag map. Two users with identical active flags produce identical fingerprints.

5. Stores `UISessionContext` in `opCtx.Data[DataKeySessionCtx]`.
6. Stores `perm_fingerprint` in `opCtx.Data[DataKeyPermFingerprint]`.
7. Stores `flag_fingerprint` in `opCtx.Data[DataKeyFlagFingerprint]`.

**Why this stage runs before CacheStage:** The cache key requires both fingerprints. Without them, the cache cannot be keyed on authorization state. Running CacheStage before AuthzStage would be a security defect — it could serve one user's schema to a different user with different permissions.

**On failure:** Returns `ui.ErrPermissionResolution`. The handler returns HTTP 503.

---

## 8.4 Stage 30 — CacheStage

**Priority:** 30
**Always runs:** Yes

Computes the cache key and performs a Redis lookup.

**Cache key format:**

```
<route>:<tenantID>:<perm_fingerprint>:<flag_fingerprint>
```

Example:
```
/finance/dashboard:ten_abc123:sha256:a3f9b2:sha256:c8d1e4
```

If a cache version token is injected at startup (`DataKeyCacheVersions`), it is also included in the key, enabling generation-based invalidation without iterating all keys.

**On cache hit:**
1. Stores the cached schema in `opCtx.Data[DataKeySchema]`.
2. Sets `opCtx.Data[DataKeyCacheHit] = true`.
3. Stages 40–80 check this flag and short-circuit (skip their main logic).
4. ResponseStage still runs and assembles the final envelope.

**On cache miss:** Sets nothing. Stages 40–80 proceed normally.

**On Redis error:** Logs the error and treats as a cache miss. The pipeline continues without caching. This is intentional — a Redis outage degrades performance (every request recompiles) but does not break the user experience.

---

## 8.5 Stage 40 — RegistryStage

**Priority:** 40
**Skipped on cache hit.**

Calls `registry.Match(route)` to find the `PageRegistration` for the requested route.

`registry.Match` tries exact match (O(1) map lookup) first, then parameterized pattern matching (O(n) over registered param patterns).

**On match:**
- Stores `ASTFn` in `opCtx.Data[DataKeyASTPageFn]` (if set)
- Stores `Fn` in `opCtx.Data[DataKeyPageFn]` (if set; legacy)
- Stores extracted URL params in `opCtx.Data[DataKeyRouteParams]`

**On no match:** Returns `ui.ErrPageNotFound`. The handler returns `{"status": 404, "msg": "schema not found: <route>"}`.

---

## 8.6 Stage 50 — CompileStage

**Priority:** 50
**Skipped on cache hit.**

Retrieves `UISessionContext` from `DataKeySessionCtx`. Injects `Params` from `DataKeyRouteParams` into the session context copy.

**ASTPageFn path (preferred):**
1. Retrieves `ASTPageFn` from `DataKeyASTPageFn`.
2. Calls `ASTPageFn(sess)` — returns `any` (actual type: `ast.Node`).
3. Asserts to `ast.Node`.
4. Calls `ast.CompileTree(node)` — validates all nodes, then emits JSON.
5. Stores schema in `DataKeySchema`.
6. Sets `DataKeyASTCompiled = true`.

**PageFn path (legacy fallback):**
1. If `DataKeyASTPageFn` is absent, retrieves `PageFn` from `DataKeyPageFn`.
2. Calls `PageFn(sess)` — returns `ui.Schema` (raw `map[string]any`).
3. Stores schema in `DataKeySchema`.
4. `DataKeyASTCompiled` remains false.

**On `CompileTree` error:** Returns `ui.ErrSchemaInvalid`. Handler returns HTTP 500.

**Page functions must be pure.** Calling a database, making an HTTP request, or reading a file inside a page function is a violation. It will not be caught by the pipeline — it is the author's responsibility to keep page functions side-effect-free.

---

## 8.7 Stage 60 — NormalizeStage

**Priority:** 60
**Skipped on cache hit.**
**Never returns an error.**

Canonicalizes the compiled schema to ensure AMIS compliance. Modifies `DataKeySchema` in place.

**Always applied:**
- Converts all `"type"` field values to lowercase.
- Trims leading/trailing whitespace from API URL strings.

**Skipped for AST-compiled schemas** (when `DataKeyASTCompiled` is true):
- Structural rules for CRUD `syncLocation` and chart transparent backgrounds.
- These are already guaranteed by the typed node implementations.

NormalizeStage never returns an error. If a transformation cannot be applied (unexpected data type), it logs a warning and skips that field. The worst outcome is a non-normalized schema, which ValidateStage will catch if it violates rules.

---

## 8.8 Stage 70 — ValidateStage

**Priority:** 70
**Skipped on cache hit.**

Applies structural and security validation rules to the compiled schema.

**Structural rules (applied to both PageFn and ASTPageFn schemas):**
- `crud` components must have `syncLocation: false`. AMIS CRUD syncs its state to the URL by default; this causes unexpected browser history behavior. All CRUD nodes must opt out.
- `chart` components must have `"style": {"background": "transparent"}` AND `"config": {"backgroundColor": "transparent"}`. Both are required — AMIS applies background from both locations.

**Security rules:**
- No IAM expression strings in the schema. The pipeline scans for patterns that would leak IAM-internal data to the browser (e.g., role names, user ID references in unexpected fields).

**All errors are `BusinessError` with `VALIDATE_*` codes.** They are programmer errors. The pipeline returns `ui.ErrSchemaInvalid` and the handler returns HTTP 500.

For AST-compiled schemas, structural rules that are guaranteed by the typed node implementations are skipped (based on `DataKeyASTCompiled`). Security rules always run.

---

## 8.9 Stage 80 — CacheStoreStage

**Priority:** 80
**Skipped on cache hit.**

Writes the compiled, normalized, validated schema to Redis under the cache key computed in CacheStage.

The TTL is configurable at startup. Default: 5 minutes. This balances cache efficiency against staleness from permission changes (which produce new fingerprints and bypass the cache anyway).

**On Redis write error:** Logs the error and continues. The response is still returned to the caller. The next request will recompile.

---

## 8.10 Stage 90 — ResponseStage

**Priority:** 90
**Always runs — cache hit or miss.**

Reads `DataKeySchema` and assembles the final `UISchemaOutput`:

```go
type UISchemaOutput struct {
    Schema   Schema // the compiled map[string]any
    CacheHit bool   // true if served from cache
    Route    string
}
```

Stores it in `DataKeyResponse`. The handler reads this and returns:

```json
{"status": 0, "data": <schema>}
```

ResponseStage always runs so that the handler always reads `DataKeyResponse` via the same code path, regardless of whether the request was a cache hit or miss.

---

## 8.11 Data Flow Between Stages

All inter-stage communication uses `opCtx.Data`. Key constants are defined in `internal/web/ui/pipeline.go`:

```go
const (
    DataKeyRoute            = "ui.request.route"
    DataKeyPermissions      = "ui.authz.permissions"
    DataKeyPermFingerprint  = "ui.authz.perm_fingerprint"
    DataKeyFlagFingerprint  = "ui.authz.flag_fingerprint"
    DataKeyCacheKey         = "ui.cache.key"
    DataKeyCacheHit         = "ui.cache.hit"
    DataKeyPageFn           = "ui.registry.page_fn"
    DataKeyASTPageFn        = "ui.registry.ast_page_fn"
    DataKeyRouteParams      = "ui.registry.route_params"
    DataKeyASTCompiled      = "ui.compile.ast_compiled"
    DataKeyCacheVersions    = "ui.cache.versions"
    DataKeySessionCtx       = "ui.authz.session_ctx"
    DataKeySchema           = "ui.compile.schema"
    DataKeyResponse         = "ui.response"
)
```

Naming convention: `"ui.<stage_name>.<key>"`. A stage only reads keys written by earlier stages (lower priority). No stage reads a key written by a later stage.

---

## 8.12 Error Handling

| Error | Stage | HTTP Status | Response Body |
|-------|-------|-------------|---------------|
| `ErrUnauthenticated` | SessionStage | 401 | AMIS page with login button |
| `ErrPermissionResolution` | AuthzStage | 503 | `{"status": 503, "msg": "service temporarily unavailable"}` |
| `ErrPageNotFound` | RegistryStage | 404 | `{"status": 404, "msg": "schema not found: <route>"}` |
| `ErrSchemaInvalid` | CompileStage, ValidateStage | 500 | `{"status": 500, "msg": "An internal error occurred."}` |
| Unclassified error | Any | 500 | Same as above |

All errors are logged with the route and tenant ID for observability.

---

## 8.13 Cache Key Design

The cache key design is fundamental to the platform's security and performance model.

**Security:** The key includes the permission fingerprint. If a user's permissions change (role added, policy updated), the fingerprint changes and the new request computes a fresh key, missing the cache. The stale entry remains in Redis but is never served — it has a different key.

**Performance:** Two users with identical permission sets within the same tenant share a cache entry. For a tenant with 100 users in the same role, the first request compiles the schema; the next 99 serve from cache.

**Tenant isolation:** The tenant ID is part of the key. A schema for tenant A is never served to tenant B, even if their permission fingerprints happen to collide.

**Invalidation:** There is no explicit cache invalidation mechanism today. Entries expire after TTL. A code deployment that changes a page function will start producing different schemas, but cached schemas from before the deployment continue to be served until they expire.

> **Status: Planned — Not Yet Implemented**
> Webhook-driven cache invalidation from business events (e.g., invalidate all `/finance/invoices` schemas when an invoice is approved) is on the roadmap.

---

## 8.14 Performance Characteristics

| Path | Typical Latency | Bottleneck |
|------|----------------|------------|
| Cache hit | 2–5ms | Redis round-trip |
| Cache miss, simple page | 20–30ms | Casbin BulkEnforce |
| Cache miss, complex page | 30–50ms | Casbin BulkEnforce + CompileTree |

**Casbin BulkEnforce** is the dominant cost on cache miss. It is a single batch call but queries the Casbin policy store (PostgreSQL). The result is not cached between requests — each cache miss pays the full Casbin cost.

**CompileTree** is O(n) in the number of AST nodes. For typical pages (20–50 nodes), this is sub-millisecond.

**Redis** is the primary performance lever. A 90%+ cache hit rate reduces average latency from ~25ms to ~3ms. Cache hit rate depends on the diversity of permission fingerprints within a tenant.

---

*End of Chapter 08*

**Previous:** [Chapter 07 — AST Design](./07-ast-design.md)
**Next:** [Chapter 09 — Component System](../vol-03-component-system/09-component-system.md)
