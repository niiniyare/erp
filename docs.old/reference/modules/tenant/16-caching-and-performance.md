[<-- Back to Index](README.md)

## Caching & Performance

### Overview

The tenant service implements a caching layer to minimize database lookups for frequently accessed tenant data. Every API request requires tenant validation, making cache performance critical.

### Cache Architecture

```markdown
CACHING LAYERS:

  API Request
       │
       ▼
  [In-Memory Cache]     ← Go map (future)
       │ miss
       ▼
  [Redis Cache]          ← Primary cache layer
       │ miss
       ▼
  [PostgreSQL]           ← Source of truth
       │
       ▼
  [Cache Population]     ← Store result in Redis
```

### Cache Keys

```markdown
CACHE KEY PATTERNS:

Key                                  Value           TTL
──────────────────────────────────────────────────────────
tenant:{uuid}                        Tenant JSON     5 min
tenant:subdomain:{subdomain}         Tenant UUID     5 min
tenant:slug:{slug}                   Tenant UUID     5 min
tenant:config:{uuid}                 Config JSON     5 min
tenant:limits:{uuid}                 Limits JSON     5 min

EXAMPLES:

tenant:f7a8b9c0-1234-5678-9abc-def012345678
  → { "id": "f7a8b9c0-...", "name": "Coastal Coffee",
      "status": "ACTIVE", "plan_type": "professional", ... }

tenant:subdomain:coastalcoffee
  → "f7a8b9c0-1234-5678-9abc-def012345678"

tenant:slug:coastal-coffee-co
  → "f7a8b9c0-1234-5678-9abc-def012345678"
```

### Cache Invalidation

```markdown
INVALIDATION TRIGGERS:

On Tenant Update:
├── DELETE tenant:{uuid}
├── DELETE tenant:subdomain:{old_subdomain}
├── DELETE tenant:slug:{old_slug}
└── Re-cache with new values on next access

On Status Change (suspend/reactivate/archive):
├── DELETE tenant:{uuid}
└── Critical: Ensures next request sees new status

On Configuration Change:
├── DELETE tenant:config:{uuid}
└── DELETE tenant:limits:{uuid}

On Tenant Delete (soft):
├── DELETE tenant:{uuid}
├── DELETE tenant:subdomain:{subdomain}
└── DELETE tenant:slug:{slug}

Manual Invalidation:
└── Service.InvalidateTenantCache(tenantID)
    Removes all cached entries for a tenant
```

### Performance Characteristics

```markdown
OPERATION LATENCY:

Operation                    Cache Hit    Cache Miss
─────────────────────────────────────────────────────
Tenant Lookup by ID          < 1ms        5-15ms
Subdomain Resolution         < 1ms        5-15ms
Tenant Validation            < 1ms        10-20ms
Set Tenant Context           N/A          2-5ms
Configuration Lookup         < 1ms        5-15ms

TARGET METRICS:

Cache Hit Rate:              > 95%
P99 Tenant Validation:       < 5ms (cached)
P99 Context Switch:          < 10ms
Concurrent Tenants:          10,000+
```

### Go Service Caching Implementation

```markdown
SERVICE METHODS WITH CACHING:

GetTenant(ctx, tenantID):
├── Check: cache.Get("tenant:" + tenantID)
├── Hit:   Return cached tenant, log cache_hit span
├── Miss:  Query DB, cache result, log cache_miss span
└── Error: Return DB error (don't cache errors)

ValidateTenantAccess(tenantID):
├── Call:  GetTenant (uses cache)
├── Check: status == ACTIVE || status == PENDING
├── Check: deleted_at IS NULL
└── Return: tenant or validation error

ResolveTenantID(subdomain):
├── Check: cache.Get("tenant:subdomain:" + subdomain)
├── Hit:   Return cached UUID
├── Miss:  Query DB by subdomain, cache UUID
└── Return: UUID or not-found error

InvalidateTenantCache(tenantID):
├── Fetch current tenant (for subdomain/slug)
├── Delete: tenant:{uuid}
├── Delete: tenant:subdomain:{subdomain}
├── Delete: tenant:slug:{slug}
└── Delete: tenant:config:{uuid}
```

### OpenTelemetry Tracing

```markdown
TRACE SPANS FOR CACHE OPERATIONS:

Span: TenantService.GetTenant
├── Attribute: tenant.id = "f7a8b9c0-..."
├── Event: cache_hit (or cache_miss)
└── Duration: 0.5ms (hit) / 12ms (miss)

Span: TenantService.ValidateTenantAccess
├── Attribute: tenant.id, tenant.status
├── Event: validation_passed / validation_failed
└── Child spans: GetTenant

Span: TenantRepository.SetTenant
├── Attribute: tenant.id
├── Event: context_set
└── Duration: 3ms (DB call)
```

---

Next: [Module Integration Points](./17-module-integration-points.md)
