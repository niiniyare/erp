# Chapter 9: Effective Configuration Resolution

[← Entity Configuration](./08-entity-configuration.md) | [Next: Template Management →](./10-template-management.md)

---

## What Is Resolution?

Resolution is the process of determining the effective value of a configuration key for a given context (tenant + optional entity). It walks the three-level hierarchy and returns the most specific value found.

---

## Single Key Resolution

The `GetEffectiveConfiguration` query resolves one key in a single database round-trip using a CTE union:

```sql
-- name: GetEffectiveConfiguration :one
WITH RECURSIVE config_resolution AS (
    -- Level 0: system default (priority 0)
    SELECT cd.default_value as value, 'system' as source, 0 as priority, cd.config_type
    FROM config_definitions cd
    WHERE cd.module_name = $1 AND cd.config_key = $2

    UNION ALL

    -- Level 1: tenant override (priority 1)
    SELECT
        (tc.settings -> ($1 || '.' || $2)) as value,
        'tenant' as source,
        1 as priority,
        cd.config_type
    FROM tenant_configurations tc
    JOIN config_definitions cd ON cd.module_name = $1 AND cd.config_key = $2
    WHERE tc.tenant_id = current_tenant_id()
      AND tc.settings ? ($1 || '.' || $2)

    UNION ALL

    -- Level 2: entity override (priority 2)
    SELECT
        (e.settings -> ($1 || '.' || $2)) as value,
        'entity' as source,
        2 as priority,
        cd.config_type
    FROM entities e
    JOIN config_definitions cd ON cd.module_name = $1 AND cd.config_key = $2
    WHERE e.tenant_id = current_tenant_id()
      AND e.uuid = $3
      AND e.deleted_at IS NULL
      AND e.settings ? ($1 || '.' || $2)
)
SELECT module_name, config_key, value, source, config_type
FROM config_resolution
ORDER BY priority DESC
LIMIT 1;
```

The `ORDER BY priority DESC LIMIT 1` selects the most specific value found. If only the system default exists, that is returned with `source = 'system'`.

---

## Multi-Key Resolution

For resolving all configurations for an entity across all modules, use `ListTenantEffectiveConfigurations`. This returns a merged view using `DISTINCT ON`:

```sql
-- name: ListTenantEffectiveConfigurations :many
WITH
  entity_configs AS (SELECT ..., 2 as priority FROM entities ...),
  tenant_configs AS (SELECT ..., 1 as priority FROM tenant_configurations ...),
  system_configs AS (SELECT ..., 0 as priority FROM config_definitions ...)
SELECT DISTINCT ON (config_full_key)
    config_full_key, module_name, config_key, value, source
FROM (
    SELECT *, 2 as priority FROM entity_configs
    UNION ALL SELECT *, 1 as priority FROM tenant_configs
    UNION ALL SELECT *, 0 as priority FROM system_configs
) combined
ORDER BY config_full_key, priority DESC;
```

`DISTINCT ON (config_full_key)` combined with `ORDER BY config_full_key, priority DESC` picks the highest-priority value for each key — the PostgreSQL equivalent of "keep only the first row per group."

---

## Resolution Response

```json
{
  "module": "finance",
  "config_key": "invoice_prefix",
  "value": "BRANCH-INV-",
  "data_type": "string",
  "source": "entity",
  "is_inherited": false,
  "can_override": true,
  "inheritance_chain": {
    "system": "DOC-",
    "tenant": "INV-",
    "entity": "BRANCH-INV-"
  }
}
```

The `inheritance_chain` field (computed by the service layer) shows all three levels, not just the winning value. This is useful for admin UIs that want to show "this entity overrides the tenant default."

---

## Batch Resolution

For performance, resolve multiple keys at once using the `/api/v1/settings/resolve` endpoint:

```go
resolved, err := settingsService.ResolveConfigurations(ctx, service.ResolutionRequest{
    TenantID:       tenantID,
    EntityID:       &entityID,
    Configurations: []service.ConfigurationKey{
        {Module: "finance", Key: "invoice_prefix"},
        {Module: "finance", Key: "auto_approval_limit"},
        {Module: "inventory", Key: "valuation_method"},
    },
    IncludeMetadata: true,
})
```

This avoids N individual database queries for N configuration keys.

---

## Caching Strategy

Resolved values are cached in Redis with source-tiered TTLs:

| Source | TTL | Rationale |
|---|---|---|
| `system` | 30 minutes | System defaults rarely change |
| `tenant` | 10 minutes | Tenant config changes occasionally |
| `entity` | 5 minutes | Entity overrides change more frequently |

Cache keys follow the pattern:
```
config:{tenantID}:{entityID}:{module}:{key}   -- entity-scoped
config:{tenantID}:tenant:{module}:{key}        -- tenant-scoped
```

On any configuration write, the relevant cache key is deleted and the next read re-resolves from the database.

---

[← Entity Configuration](./08-entity-configuration.md) | [Next: Template Management →](./10-template-management.md)
