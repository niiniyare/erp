---
chapter: 32
title: "Performance Optimization"
volume: "vol-07-reliability"
section: "Reliability"
description: "Cache hit path latency, stage priorities, BulkEnforce single-call design, permission fingerprinting, static block memory, and InstrumentedStage thresholds."
status: implemented
---

# Chapter 32 — Performance Optimization

## Table of Contents

- [32.1 Cache Hit Path vs Cache Miss Path](#321-cache-hit-path-vs-cache-miss-path)
- [32.2 Stage Priorities and Skip-on-Cache-Hit](#322-stage-priorities-and-skip-on-cache-hit)
- [32.3 BulkEnforce — One Casbin Call Per Request](#323-bulkenforce)
- [32.4 perm_fingerprint — Sharing Cached Schemas](#324-perm_fingerprint)
- [32.5 Static Blocks and SetGlobalMemory](#325-static-blocks-and-setglobalmemory)
- [32.6 InstrumentedStage — 50ms Warning Threshold](#326-instrumentedstage)

---

## 32.1 Cache Hit Path vs Cache Miss Path

The schema pipeline is designed so that the common case (cache hit) is extremely fast.

### Cache Hit Path (~2–5ms)

```
Request arrives
      ↓
Authenticate (JWT validation — ~1ms)
      ↓
InjectSessionContext (~0.5ms, reads from session store)
      ↓
AuthzStage — BulkEnforce (~1-2ms, resolves perm + flag fingerprints)
      ↓
CacheStage — key lookup in Redis (~0.5ms)
      ↓  CACHE HIT → return cached schema
Response (~2-5ms total)
```

All stages after CacheStage (CompileStage, ValidateStage, CacheStoreStage) are skipped on a cache hit. The response is the raw cached bytes deserialized and written directly.

### Cache Miss Path (~20–100ms)

```
Request arrives
      ↓
Authenticate + InjectSessionContext (~1.5ms)
      ↓
AuthzStage — BulkEnforce (~5-20ms, IAM network call)
      ↓
CacheStage — key lookup → MISS
      ↓
CompileStage — execute PageFn + all block functions (~5-30ms)
      ↓
ValidateStage — schema assertion checks (~1-5ms)
      ↓
CacheStoreStage — write to Redis (~1-2ms)
      ↓
Response (~15-60ms total)
```

The dominant cost in the miss path is `BulkEnforce` (IAM network latency) and `CompileStage` (schema complexity). Complex pages with many blocks and conditional paths take longer to compile.

### Warm-Up

At server startup, frequently accessed schemas (dashboard, nav tree) can be pre-warmed by calling their `PageFn` for known tenant/permission combinations and populating the cache before the first user request.

---

## 32.2 Stage Priorities and Skip-on-Cache-Hit

Each stage in the pipeline has a numeric priority. Lower priority number = runs earlier.

| Stage              | Priority | Skipped on cache hit? |
|--------------------|----------|-----------------------|
| AuthzStage         | 20       | No — always runs (produces fingerprints) |
| CacheStage         | 30       | No — this is the cache check itself       |
| CompileStage       | 50       | Yes                                       |
| ValidateStage      | 60       | Yes                                       |
| CacheStoreStage    | 80       | Yes (nothing to store)                   |

AuthzStage always runs because the fingerprints it produces are required by CacheStage to look up the cache key. This is the minimum work required on every request — JWT validation, session hydration, and one BulkEnforce call.

CacheStage always runs because it performs the lookup. If it finds a hit, it signals the pipeline to skip remaining stages and return the cached response.

Stages at priority > 30 are skipped on a cache hit. This means permission pruning, AMIS schema construction, and ValidateStage security assertions only run when compiling a new schema variant — not on every request.

---

## 32.3 BulkEnforce

`UIAuthzService.BulkEnforce` is called exactly once per request by `AuthzStage`. No other code in the pipeline calls it.

```go
// Called once in AuthzStage.Execute()
perms, err := svc.BulkEnforce(ctx, userID, tenantID, allUIPermissions)
```

`allUIPermissions` contains every permission string in the platform. A typical deployment has 50–200 permission strings. Casbin evaluates all of them in a single batch operation, which is far cheaper than 50–200 individual policy checks.

### Why One Call Matters

Without batch resolution, a page with 50 permission-gated components would make 50 Casbin calls, each involving a policy evaluation and potentially a database read. At 5ms per call, that is 250ms of IAM overhead per page load — making the schema API slower than a traditional server-rendered page.

With `BulkEnforce`, the overhead is constant regardless of schema complexity. Adding a new permission-gated button to a page costs 0ms at request time.

### BulkEnforce and Cache Interaction

Even on a cache hit, `BulkEnforce` runs — AuthzStage is not skipped. However, the BulkEnforce result on a cache hit is used only to compute the `perm_fingerprint` for the cache key lookup. The resulting permission map is not used to compile a schema (CompileStage is skipped).

If IAM response times are a bottleneck, the `UIAuthzService` implementation can be backed by a short-lived local cache (e.g., 30-second TTL) that avoids the IAM network call on warm requests.

---

## 32.4 perm_fingerprint

Users who have identical permissions produce the same `perm_fingerprint`. This allows them to share a single cached schema.

```go
// perm_fingerprint is a deterministic hash of the permission map values
// Users A and B both have: finance.invoices.read=true, finance.invoices.create=false
// → perm_fingerprint = sha256(sorted keys + values) = "abc123"

// Cache key for both users is identical (same tenant, same route)
key = "finance/invoices:tenant-x:v2:v1:pg3:sg1:abc123:flag-xyz"
// → Cache HIT for User B if User A's request already compiled and stored
```

In a typical deployment, most users share a small number of role archetypes. This means the number of distinct cached schema variants is proportional to the number of distinct permission combinations — which is much smaller than the number of users.

### Fingerprint Stability

The fingerprint is computed from the sorted, canonical representation of the permission map. Adding new permissions to `allUIPermissions` changes the input to the hash, producing a new fingerprint and triggering cache misses until schemas are recompiled for each permission combination.

To avoid cascading cache misses when adding new permissions, bump `PolicyGeneration` in `CacheVersions` (§34.3). This invalidates all existing cached schemas before the new permission strings take effect.

---

## 32.5 Static Blocks and SetGlobalMemory

Some blocks are completely static — they do not depend on permissions, feature flags, tenant, or user. For example: a list of country codes, a static help panel, a fixed navigation section.

Static blocks are registered at startup using `SetGlobalMemory`:

```go
// At startup — never expires, no tenant scoping
cache.SetGlobalMemory("blocks:country-select", countrySelectSchema)
cache.SetGlobalMemory("blocks:timezone-select", timezoneSelectSchema)
```

When a `PageFn` requests a static block, the pipeline returns the pre-compiled schema bytes directly from in-process memory without a Redis call or a BulkEnforce call.

Static blocks do not appear in the per-request cache key. They are composed into the page schema at compile time, and the composed result is cached normally.

### When to Use SetGlobalMemory

Use `SetGlobalMemory` for blocks that:
- Have no conditional logic whatsoever.
- Do not reference `sess.*` in any way.
- Do not call external services.
- Are stable across all tenants and users.

If a block has even one `sess.Can()` or `sess.Flag()` call, it is not static and must not use `SetGlobalMemory`.

---

## 32.6 InstrumentedStage

`InstrumentedStage` is a wrapper that adds observability to any stage. Every stage in the production pipeline is wrapped with `InstrumentedStage`.

```go
// Every stage is wrapped at registration time
pipeline.Register(InstrumentedStage{
    Inner:    &AuthzStage{authzSvc: svc},
    StageName: "authz",
})
```

### What It Does

For each `Execute` call on the inner stage, `InstrumentedStage`:

1. **Opens an OTel span** with the stage name and route as span attributes.
2. **Records stage duration** in the `ui_stage_execution_duration_ms` histogram.
3. **Emits a warning log** if the stage takes longer than 50ms.
4. **Propagates trace context** into the inner stage's context.

### 50ms Warning Threshold

The 50ms threshold is a heuristic for "a stage is doing something unexpected." In normal operation:
- `AuthzStage` should complete in <20ms (IAM call + fingerprint computation).
- `CacheStage` should complete in <5ms (Redis round-trip).
- `CompileStage` should complete in <30ms for typical pages.
- `ValidateStage` should complete in <5ms (in-process traversal).

If `CompileStage` is consistently above 50ms, the `PageFn` may be doing I/O (e.g., calling an external service during schema construction). Page functions must not make I/O calls — all data must be available through the session context or baked into the schema as API endpoints that AMIS will call from the browser.

### Available Metrics

| Metric                                  | Type      | Description                                      |
|-----------------------------------------|-----------|--------------------------------------------------|
| `ui_compile_duration_ms`                | Histogram | Total compilation time for cache miss path       |
| `ui_stage_execution_duration_ms`        | Histogram | Per-stage execution time, labeled by stage name  |
| `ui_schema_validation_failures_total`   | Counter   | ValidateStage assertion failures                 |
| `ui_cache_generation_mismatch_total`    | Counter   | Cache hits rejected due to version mismatch      |
| `ui_invalidation_events_total`          | Counter   | Cache invalidation events, labeled by scope      |
| `ui_registry_resolution_failures_total` | Counter   | Route key not found in registry                  |

These metrics are available in Prometheus format at the standard `/metrics` endpoint.

### UIStageAttributes

Each OTel span carries these attributes:

```
stage             = "authz" | "cache" | "compile" | "validate" | "cache_store"
route             = "finance/invoices"
tenant_id         = "tenant-x"
operation_key     = "finance/invoices:tenant-x:v2:..."  (full cache key)
cache_hit         = true | false
ast_compiled      = true | false
```

These attributes make it possible to filter traces by stage, route, tenant, or cache hit status in any OTel-compatible backend (Jaeger, Tempo, etc.).
