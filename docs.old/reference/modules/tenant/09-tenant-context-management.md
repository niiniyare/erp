[<-- Back to Index](README.md)

## Tenant Context Management

### Overview

Every API request must establish a tenant context before accessing any data. The Go service layer manages this through a combination of middleware, service methods, and repository operations.

### Request Flow

```markdown
API REQUEST LIFECYCLE:

  HTTP Request
  ├── Header: X-Tenant-ID: f7a8b9c0-...
  │   OR
  ├── Subdomain: coastalcoffee.awo-erp.com
       │
       ▼
  [Middleware: Extract Tenant ID]
  ├── From header: Direct UUID lookup
  └── From subdomain: ResolveTenantID("coastalcoffee")
       │
       ▼
  [Service: ValidateTenantAccess(tenantID)]
  ├── Check cache first (Redis)
  ├── Cache miss → Query database
  ├── Validate: status is ACTIVE or PENDING
  ├── Validate: not soft-deleted
  └── Cache result for future requests
       │
       ▼
  [Repository: SetTenant(ctx, tenantID)]
  ├── Calls: validate_and_set_tenant_context(tenantID)
  ├── Updates: last_activity_at
  └── Sets: PostgreSQL session variable
       │
       ▼
  [Handler: Process Business Logic]
  ├── All queries automatically filtered by RLS
  └── Tenant context active for entire transaction
       │
       ▼
  [Cleanup: ResetTenant(ctx)]
  └── Calls: clear_tenant_context()
```

### Subdomain Resolution

```markdown
SUBDOMAIN RESOLUTION:

Input: "coastalcoffee" (from coastalcoffee.awo-erp.com)

Service.ResolveTenantID("coastalcoffee"):
├── Check cache: tenant:subdomain:coastalcoffee
├── Cache hit → Return cached tenant ID
├── Cache miss:
│   ├── Query: SELECT id FROM tenants
│   │          WHERE subdomain = 'coastalcoffee'
│   │          AND deleted_at IS NULL
│   ├── Found → Cache result, return tenant ID
│   └── Not found → Return error
└── Return: f7a8b9c0-1234-5678-9abc-def012345678
```

### Caching Strategy

The service layer caches tenant data to avoid repeated database lookups:

```markdown
CACHE KEYS AND TTL:

Key Pattern                              TTL      Content
────────────────────────────────────────────────────────────
tenant:{uuid}                            5 min    Full tenant record
tenant:subdomain:{subdomain}             5 min    Tenant ID
tenant:slug:{slug}                       5 min    Tenant ID

CACHE INVALIDATION:

On tenant update:
├── Delete: tenant:{uuid}
├── Delete: tenant:subdomain:{old_subdomain}
├── Delete: tenant:slug:{old_slug}
└── If subdomain/slug changed, cache new values

On tenant delete:
└── Delete all keys for this tenant

On status change:
└── Delete: tenant:{uuid} (forces re-validation)
```

### Go Service Interface

```markdown
SERVICE METHODS FOR CONTEXT MANAGEMENT:

SetTenant(ctx, tenantID)
├── Validates tenant access
├── Sets database context
└── Returns error if tenant invalid

ResetTenant(ctx)
├── Clears database context
└── Called in defer after SetTenant

ValidateTenantAccess(tenantID) → (*Tenant, error)
├── Checks cache
├── Falls back to database
├── Validates status and deleted_at
└── Returns tenant or error

ResolveTenantID(subdomain) → (UUID, error)
├── Maps subdomain to tenant UUID
├── Uses cache with DB fallback
└── Returns error if not found

InvalidateTenantCache(tenantID)
├── Removes all cached data for tenant
└── Called after any tenant modification
```

### Tracing Integration

All tenant operations include OpenTelemetry tracing:

```markdown
TRACE SPANS:

TenantService.GetTenant
├── Attributes: tenant.id
└── Events: cache_hit or cache_miss

TenantService.ValidateTenantAccess
├── Attributes: tenant.id, tenant.status
└── Events: validation_passed or validation_failed

TenantRepository.SetTenant
├── Attributes: tenant.id
└── Events: context_set

TenantRepository.GetBySubdomain
├── Attributes: tenant.subdomain
└── Events: tenant_found or tenant_not_found
```

---

Next: [Usage Tracking & Analytics](./10-usage-tracking-and-analytics.md)
