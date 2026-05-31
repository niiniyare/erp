[<-- Back to Index](README.md)

## Feature Flags — Module & Resource Toggles

### What Flags Control

A feature flag is a binary decision: does this module or resource exist for this tenant? When a flag is off:
- The affected nav section is absent from the UI
- The schema endpoint returns 403
- The API route returns 403

From the user's perspective, the feature simply does not exist.

Flags operate at two levels:
- **Module flag** (`finance`) — the entire Finance module. When off: no Finance nav, no Finance API.
- **Resource flag** (`finance.transactions`) — a specific resource within an enabled module. When off: that resource's nav item is absent even though other Finance resources remain visible.

---

### Schema

```sql
-- Flag catalogue — shared across all tenants, no tenant_id
CREATE TABLE feature_flag_definitions (
  id            uuid  PRIMARY KEY DEFAULT gen_random_uuid(),
  module_id     uuid  REFERENCES modules(id),    -- NULL = global/platform flag
  resource_id   uuid  REFERENCES resources(id),  -- NULL = module-level flag
  flag_key      text  UNIQUE NOT NULL,            -- 'finance' | 'finance.transactions'
  label         text  NOT NULL,
  description   text,
  default_value bool  NOT NULL DEFAULT false,     -- what tenants get without any configuration
  is_system     bool  NOT NULL DEFAULT false      -- true = only platform operators can toggle
);

-- Per-tenant flag values — tenant-scoped
CREATE TABLE tenant_feature_flags (
  id         uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id  uuid        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  flag_id    uuid        NOT NULL REFERENCES feature_flag_definitions(id),
  flag_key   text        NOT NULL,    -- denormalised from definition for fast lookup
  enabled    bool        NOT NULL,
  set_by     uuid        REFERENCES users(id),
  set_at     timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, flag_id)
);

CREATE INDEX idx_tenant_flags_tenant ON tenant_feature_flags(tenant_id);
```

`feature_flag_definitions` is the catalogue — seeded by migrations/triggers, readable by all tenants. `tenant_feature_flags` is the configuration — one row per tenant per flag they have explicitly configured. A flag with no row in `tenant_feature_flags` takes its `default_value` from the definition.

---

### Resolution

```go
// internal/platform/repo/flag_repo_impl.go

func (r *flagRepoImpl) ResolveForTenant(ctx context.Context,
    tenantID uuid.UUID) (map[string]bool, error) {

    // Single query: all definitions LEFT JOIN tenant overrides
    rows, err := r.q.ResolveAllFlagsForTenant(ctx, db.ResolveAllFlagsForTenantParams{
        TenantID: tenantID,
    })
    if err != nil { return nil, err }

    flags := make(map[string]bool, len(rows))
    for _, row := range rows {
        // COALESCE(tenant_override, default_value)
        flags[row.FlagKey] = row.EffectiveValue
    }
    return flags, nil
}
```

```sql
-- db/queries/flags.sql
-- name: ResolveAllFlagsForTenant :many
SELECT
    ffd.flag_key,
    COALESCE(tff.enabled, ffd.default_value) AS effective_value
FROM feature_flag_definitions ffd
LEFT JOIN tenant_feature_flags tff
    ON tff.flag_id = ffd.id AND tff.tenant_id = @tenant_id
ORDER BY ffd.flag_key;
```

One query at login. Result stored in `sessions.configuration JSONB`. Every subsequent check is an O(1) map lookup.

---

### Hierarchy Rule

When checking whether a feature is accessible, both the module flag and the resource flag must be true:

```go
func (s *ResolvedSession) FeatureEnabled(flagKey string) bool {
    // Check the exact flag
    if !s.Configuration.Flags[flagKey] { return false }

    // If it's a resource key (contains a dot), also check the module flag
    if idx := strings.Index(flagKey, "."); idx > 0 {
        moduleKey := flagKey[:idx]
        if !s.Configuration.Flags[moduleKey] { return false }
    }
    return true
}

// finance.transactions.approval_workflow is enabled only if:
// flags["finance"] = true AND flags["finance.transactions"] = true AND
// flags["finance.transactions.approval_workflow"] = true
```

---

### FlagService Interface

```go
type FlagService interface {
    // Catalogue (read-only at runtime; written via migrations/admin)
    ListDefinitions(ctx, params domain.ListFlagDefsParams)    ([]*domain.FlagDefinition, error)
    GetDefinition(ctx, flagKey string)                        (*domain.FlagDefinition, error)

    // Tenant configuration
    ResolveForTenant(ctx, tenantID uuid.UUID)                 (map[string]bool, error)
    ListForTenant(ctx, params domain.ListFlagsParams)         ([]*domain.TenantFlagWithDef, error)
    Set(ctx, params domain.SetFlagParams)                     error  // enables/disables
    Reset(ctx, params domain.ResetFlagParams)                 error  // removes override, restores default
}
```

When `Set()` changes a module or resource flag, it calls `sessionRepo.InvalidateByTenant()` — all tenant sessions are deleted, forcing users to re-login to get the updated flag state in their session.

---

### Route-Level Flag Enforcement

Flags gate routes at the middleware level, not just the UI:

```go
// Three independent gates for a feature — all must pass

// Gate 1: nav (boot schema) — section only appears if module+resource flags are on
// Gate 2: schema route — 403 if flag is off
sg.Get("/accounting/transactions",
    middleware.RequireFlag("finance.transactions"),
    middleware.RequirePermission("finance.transactions", "read"),
    handlers.TransactionListSchema(deps))

// Gate 3: data API route — same flag check
api.Get("/transactions",
    middleware.RequireFlag("finance.transactions"),
    middleware.RequirePermission("finance.transactions", "read"),
    handlers.ListTransactions(deps))
```

403 (not 404) when a flag is off — the distinction is intentional. 404 implies the route doesn't exist; 403 implies it exists but is gated.

---

Next: [Tenant Settings](./05b-tenant-settings.md)
