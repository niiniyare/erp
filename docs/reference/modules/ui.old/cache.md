# UI Schema Cache

Compiled schemas are cached per (tenant × route × permission set × feature flags
× compiler versions). The cache dramatically reduces Casbin + PageFn invocations
at scale. Two requests with identical context always produce identical schemas —
caching is safe.

---

## Cache Key

The canonical key is 8 components joined with `:`:

```
ui:schema:{tenantID}:{route}:{permFP}:{flagFP}:{compilerVer}:{astVer}:{policyGen}:{schemaGen}
```

| Component | Source | Purpose |
|-----------|--------|---------|
| `tenantID` | `opCtx.TenantID` | Tenant isolation |
| `route` | `UISchemaInput.Route` | Page identity |
| `permFP` | SHA256 of sorted permission keys | Permission set identity |
| `flagFP` | SHA256 of sorted flag keys | Feature flag identity |
| `compilerVer` | `CacheVersions.CompilerVersion` | Invalidate on pipeline code change |
| `astVer` | `CacheVersions.ASTVersion` | Invalidate on AST type change |
| `policyGen` | `CacheVersions.PolicyGeneration` | Soft-invalidate on IAM policy change |
| `schemaGen` | `CacheVersions.SchemaGeneration` | Soft-invalidate on page DSL change |

`CacheVersions` are injected by `UIPipeline.Run()` before the pipeline starts.
In production, inject real values from config/env. In tests, use
`uicache.DefaultVersions()`.

### Why build the key AFTER AuthzStage?

If the key were built before permission resolution, two users with different
roles on the same route could share a cache entry — serving one user's
permission-filtered schema to another. The cache key is only safe to compute
after `DataKeyPermFingerprint` and `DataKeyFlagFingerprint` are set.

---

## Invalidation Scopes

Call `cache.InvalidateSchemaCache(ctx, svc, req)` to evict cached schemas.
Never call `cache.Service.DeletePattern` directly.

| Scope | What gets evicted | When to use |
|-------|-------------------|-------------|
| `InvalidateTenant` | All schemas for a tenant | Tenant config change, subscription change |
| `InvalidateModule` | All schemas for one module prefix | Module DSL change for specific tenant |
| `InvalidatePolicy` | Nothing (soft) — increments `PolicyGeneration` | IAM policy change (role assignment, permission update) |
| `InvalidateHard` | All schemas for a tenant (same as tenant, intent: urgent) | Stale data that cannot wait for TTL |

### Soft vs Hard Invalidation

**Soft invalidation** (`InvalidatePolicy`): increments the `PolicyGeneration`
counter in Redis. No keys are deleted. On the next request, the key contains
the new generation value and misses the old cached entry. Old entries expire by
TTL (5 minutes). Use this for IAM policy changes — it avoids a cache stampede
when many tenants are affected simultaneously.

**Hard invalidation** (`InvalidateTenant`, `InvalidateHard`): calls
`cache.Service.DeletePattern(ctx, "ui:schema:{tenantID}:*")`. All schemas for
the tenant are immediately evicted. Use only when stale data would cause a
correctness or security issue.

---

## Cache Patterns

```go
// All schemas for a tenant
"ui:schema:{tenantID}:*"

// All schemas for one module under a tenant
"ui:schema:{tenantID}:{modulePrefix}/*"
```

Patterns are defined in `internal/web/cache/` — never inline Redis glob strings.

---

## TTL

The default schema cache TTL is **5 minutes**. This is intentionally short:

- IAM policy changes are soft-invalidated immediately (no stale window)
- DSL/page changes are caught by incrementing `SchemaGeneration`
- The 5-minute TTL is a safety net for any cache-key miss scenarios

---

## Metrics

| Metric | Labels | Meaning |
|--------|--------|---------|
| `ui_cache_generation_mismatch_total` | `route`, `tenant_id`, `reason` | Cache miss caused by generation increment |
| `ui_invalidation_events_total` | `tenant_id`, `scope` | Invalidation calls dispatched |

`reason` values: `"policy_gen"` (PolicyGeneration changed), `"schema_gen"`
(SchemaGeneration changed).
