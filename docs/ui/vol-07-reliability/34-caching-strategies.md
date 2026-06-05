---
chapter: 34
title: "Caching Strategies"
volume: "vol-07-reliability"
section: "Reliability"
description: "CacheStage and CacheStoreStage, the 8-component cache key, CacheVersions struct, four invalidation scopes, cache.Service.DeletePattern, TTL strategy, and cache miss metrics."
status: implemented
---

# Chapter 34 — Caching Strategies

## Table of Contents

- [34.1 CacheStage and CacheStoreStage in the Pipeline](#341-cachestage-and-cachestorestage)
- [34.2 The 8-Component Cache Key](#342-the-8-component-cache-key)
- [34.3 CacheVersions — Global Version Bumping](#343-cacheversions)
- [34.4 Four Invalidation Scopes](#344-four-invalidation-scopes)
- [34.5 cache.Service.DeletePattern — The Only Invalidation API](#345-cacheservicedeletepattern)
- [34.6 TTL Strategy](#346-ttl-strategy)
- [34.7 Cache Miss Metrics](#347-cache-miss-metrics)

---

## 34.1 CacheStage and CacheStoreStage in the Pipeline

The cache is split across two stages that run at different points in the pipeline.

### CacheStage (priority 30)

`CacheStage` runs immediately after `AuthzStage`. It has the permission fingerprint and flag fingerprint it needs to construct the full cache key. It attempts to retrieve the compiled schema from the cache backend.

On a **cache hit**: the stage writes the cached schema bytes into the pipeline response and signals all remaining stages (CompileStage, ValidateStage, CacheStoreStage) to skip. The response is served immediately.

On a **cache miss**: the stage does nothing and allows the pipeline to continue. `CompileStage` will run next, build the schema, and eventually `CacheStoreStage` will store the result.

```go
// Conceptual flow inside CacheStage.Execute()
key := uicache.Key(route, tenantID, versions, permFingerprint, flagFingerprint)
schema, err := cache.Get(ctx, key)
if err == nil {
    pipeline.SetResponse(schema)
    pipeline.SkipRemaining()
    return
}
// cache miss — continue
```

### CacheStoreStage (priority 80)

`CacheStoreStage` runs after `ValidateStage`. It stores the compiled and validated schema in the cache backend using the same key computed by `CacheStage`.

If CacheStage signaled a cache hit, `CacheStoreStage` is not reached (the pipeline was short-circuited). `CacheStoreStage` only runs when a new schema variant has been compiled.

```go
// Conceptual flow inside CacheStoreStage.Execute()
key := uicache.Key(route, tenantID, versions, permFingerprint, flagFingerprint)
schema := pipeline.GetCompiledSchema()
cache.Set(ctx, key, schema, ttlForSchema(schema))
```

---

## 34.2 The 8-Component Cache Key

Every cached schema is indexed by an 8-component key constructed by `uicache.Key`:

```
<route>:<tenant_id>:<compiler_version>:<ast_version>:<policy_generation>:<schema_generation>:<perm_fingerprint>:<flag_fingerprint>
```

### Component Breakdown

| Position | Component            | Source                        | Changes when...                              |
|----------|----------------------|-------------------------------|----------------------------------------------|
| 1        | `route`              | Request path                  | Different page is requested                  |
| 2        | `tenant_id`          | JWT claim                     | Different tenant logs in                     |
| 3        | `compiler_version`   | `CacheVersions.CompilerVersion` | Pipeline internals change                  |
| 4        | `ast_version`        | `CacheVersions.ASTVersion`    | DSL node types change                        |
| 5        | `policy_generation`  | `CacheVersions.PolicyGeneration` | Permission model changes                  |
| 6        | `schema_generation`  | `CacheVersions.SchemaGeneration` | Page function code changes                |
| 7        | `perm_fingerprint`   | Hash of user's permission map | User gains/loses permissions                 |
| 8        | `flag_fingerprint`   | Hash of user's flag map       | Feature flags change for user/tenant         |

### Example Keys

```
finance/invoices:tenant-abc:v3:v2:pg5:sg12:perm-aef3c2:flag-b819d4
hr/employees:tenant-abc:v3:v2:pg5:sg12:perm-aef3c2:flag-b819d4
finance/invoices:tenant-xyz:v3:v2:pg5:sg12:perm-aef3c2:flag-c003a1
```

User A and User B in the same tenant with the same permissions share a cache entry (keys 1–6 and 8 match, key 7 matches because same permissions → same fingerprint).

---

## 34.3 CacheVersions

`CacheVersions` is a struct that holds the four version integers baked into every cache key:

```go
type CacheVersions struct {
    CompilerVersion  int  // Increment when the pipeline execution engine changes
    ASTVersion       int  // Increment when DSL node schemas change
    PolicyGeneration int  // Increment when the permission model changes
    SchemaGeneration int  // Increment when page functions change
}
```

These versions are set at server startup (typically from configuration) and are constant for the lifetime of the process. Bumping a version makes all existing cached schemas for that version component obsolete — they will be cache misses on the next request and will be recompiled.

### When to Bump Each Version

| Version            | Bump when...                                                                    |
|--------------------|---------------------------------------------------------------------------------|
| `CompilerVersion`  | The pipeline stage order, stage logic, or key construction logic changes        |
| `ASTVersion`       | A DSL node type is added, removed, or its fields change in a breaking way       |
| `PolicyGeneration` | Permission strings are added/removed from `allUIPermissions`, or Casbin policies are reorganized |
| `SchemaGeneration` | Any `PageFn` or block function is modified, or the schema registry changes      |

### Version Bump Strategy

For development deployments, bump `SchemaGeneration` freely — developers expect stale schemas after code changes.

For production, increment versions as part of the deployment. The version numbers should be tracked in the deployment configuration alongside the binary version.

Bumping `SchemaGeneration` is the most common operation and is safe to do on every deployment that touches schema code. The cost is a cache miss wave on first load after deployment — all schemas are recompiled once, then cached for subsequent requests.

---

## 34.4 Four Invalidation Scopes

Cache invalidation uses glob pattern matching against the cache key string. Four scopes cover all operational needs:

### Scope 1: Route-Specific

Invalidate all cached variants of a single page (all tenants, all permission combinations):

```go
cache.DeletePattern(ctx, "finance/invoices:*")
```

Use this when you fix a bug in the `InvoiceListPage` function or its blocks.

### Scope 2: Module-Level

Invalidate all pages in a module:

```go
cache.DeletePattern(ctx, "finance/*:*")
```

Use this when a shared block used by all finance pages is updated, or when the finance module's permission set changes.

### Scope 3: Tenant-Level

Invalidate all cached schemas for a single tenant (all routes, all permission combinations):

```go
cache.DeletePattern(ctx, "*:tenant-abc:*")
```

Use this when:
- A tenant's feature flags are updated.
- A tenant's role assignments change significantly.
- A tenant's configuration (currency, locale) changes.

Tenant-level invalidation is the most common operational trigger. It is safe to issue because recompilation is fast and the cache will repopulate on first load.

### Scope 4: Global

Invalidate the entire schema cache:

```go
cache.DeletePattern(ctx, "*")
```

Use this during:
- Major version bumps where all schemas must be recompiled.
- Emergency rollbacks where stale data may be cached.
- Platform upgrades that change the AMIS schema format.

Global invalidation should be paired with a `CacheVersions` bump in the new deployment so that old cached entries (from the previous version) are also rejected even if they survive the pattern delete.

---

## 34.5 cache.Service.DeletePattern

`cache.Service.DeletePattern` is the only authorized cache invalidation API. Direct Redis operations (`KEYS`, `DEL`, `FLUSHDB`) are prohibited.

```go
type Service interface {
    Get(ctx context.Context, key string) ([]byte, error)
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    SetNoExpiry(ctx context.Context, key string, value []byte) error
    DeletePattern(ctx context.Context, pattern string) error
}
```

The reasons for this restriction:

1. **Key format is centralized**: Only `uicache.Key` knows the exact format. If the format changes, all code using `DeletePattern` with glob patterns would break — but at least the break is visible. Direct `DEL` calls with hardcoded key strings would become silently incorrect.

2. **Multi-backend safety**: The `cache.Service` interface may be backed by Redis, Memcached, or an in-process store (for testing). Direct Redis calls bypass the interface and break tests.

3. **Instrumentation**: `DeletePattern` is instrumented with the `ui_invalidation_events_total` metric. Direct Redis calls produce no metrics.

4. **Audit trail**: All invalidation events are logged with the pattern, scope, and caller. Direct Redis calls are untracked.

---

## 34.6 TTL Strategy

Different schema types have different expiry policies.

### Static Blocks — No Expiry

Schemas composed entirely of static blocks (no session data, no permissions, no flags) are stored with `SetNoExpiry` (i.e., `SetGlobalMemory` at startup):

```go
cache.SetNoExpiry(ctx, "blocks:country-select", countrySelectBytes)
```

These entries remain until the process restarts or an explicit invalidation is issued. Static block schemas never need to be recomputed unless the block code itself changes (in which case bump `SchemaGeneration` and invalidate).

### Permission-Filtered Schemas — Short TTL

Schemas that include permission-based components are stored with a short TTL (typically 5–15 minutes):

```go
cache.Set(ctx, key, schemaBytes, 10*time.Minute)
```

The short TTL ensures that permission changes propagate within a bounded time window, even if no explicit cache invalidation is triggered. For immediate propagation after a permission change, issue a tenant-level `DeletePattern`.

### Navigation Tree — Medium TTL

The nav tree schema is recompiled less frequently but changes when modules are enabled/disabled or when the user's role changes. A 5-minute TTL balances freshness with compilation cost:

```go
cache.Set(ctx, navTreeKey, navTreeBytes, 5*time.Minute)
```

---

## 34.7 Cache Miss Metrics

The `ui_cache_generation_mismatch_total` counter increments when a cache entry is retrieved but rejected because its embedded version numbers don't match the current `CacheVersions`.

This counter measures a specific failure mode: a cached entry exists (the key matches) but the content is from a different generation and cannot be used. This happens when:
- The cache was not flushed after a version bump.
- A Redis key survived a version bump due to a pattern invalidation failure.
- Two processes with different versions are racing (during a rolling deploy).

```
ui_cache_generation_mismatch_total{route="finance/invoices", tenant="tenant-abc"} 14
```

A sustained high value for this metric after a deploy indicates that old cached entries are still present and the invalidation was incomplete. Issue a global `DeletePattern("*")` to clear them.

Other cache metrics to monitor:

| Metric                                  | What high values mean                                     |
|-----------------------------------------|-----------------------------------------------------------|
| `ui_compile_duration_ms` p99 > 100ms    | Complex pages; investigate slow block functions           |
| `ui_stage_execution_duration_ms` > 50ms | Stage doing I/O — block function calling external service |
| `ui_schema_validation_failures_total`   | Developer error — permission string in expressions        |
| `ui_registry_resolution_failures_total` | Client requesting routes not in registry — 404s           |
| `ui_invalidation_events_total`          | Operational activity — track scope distribution           |
