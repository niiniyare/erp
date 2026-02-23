# Chapter 8: Entity Configuration

[← Tenant Configuration](./07-tenant-configuration.md) | [Next: Effective Resolution →](./09-effective-resolution.md)

---

## Overview

Entity-level configuration provides the most granular level of customization. It allows individual branches, locations, warehouses, or any entity to override tenant-level defaults with values specific to their operation.

Entity configuration is stored in the `settings` JSONB column on the `entities` table — shared infrastructure with the Entity module, managed by the Settings Module.

---

## When to Use Entity Configuration

| Use Case | Example |
|---|---|
| Branch-specific document prefixes | `finance.invoice_prefix = "NYC-INV-"` |
| Location-specific approval limits | `finance.auto_approval_limit = 5000` for HQ |
| Warehouse-specific inventory rules | `inventory.allow_negative_stock = true` for transit warehouse |
| Department-specific HR rules | `hr.overtime_threshold = 45` for manufacturing floor |

---

## Reading Entity Configuration

```sql
-- name: GetEntityConfigurations :one
SELECT settings FROM entities
WHERE tenant_id = current_tenant_id()
  AND uuid = $1       -- entity_id
  AND deleted_at IS NULL;
```

Returns the raw JSONB settings blob for the entity. The effective resolution query (see Chapter 9) is used when you need the resolved value with inheritance applied.

---

## Writing Entity Configuration

```sql
-- name: UpdateEntityConfiguration :exec
UPDATE entities
SET
    settings = jsonb_set(
        COALESCE(settings, '{}'),
        $1::text[],   -- config_path
        $2::jsonb     -- config_value
    ),
    updated_at = NOW()
WHERE tenant_id = current_tenant_id()
  AND uuid = $3;      -- entity_id
```

Go call:

```go
err = store.UpdateEntityConfiguration(ctx, db.UpdateEntityConfigurationParams{
    ConfigPath:  []string{"finance.invoice_prefix"},
    ConfigValue: []byte(`"NYC-INV-"`),
    EntityID:    entityUUID,
})
```

---

## Deleting Entity Configuration

Removing an entity-level override causes the value to fall back to the tenant or system level:

```sql
-- name: DeleteEntityConfiguration :exec
UPDATE entities
SET
    settings = settings #- $1::text[],
    updated_at = NOW()
WHERE tenant_id = current_tenant_id()
  AND uuid = $2;
```

---

## Bulk Entity Configuration

For updating multiple entities at once, the `BulkUpdateEntityConfiguration` query uses `ON CONFLICT` upsert to merge settings:

```sql
-- name: BulkUpdateEntityConfiguration :exec
INSERT INTO entities (uuid, tenant_id, settings, updated_at)
VALUES ($1, current_tenant_id(), $2, NOW())
ON CONFLICT (uuid, tenant_id)
DO UPDATE SET
    settings = entities.settings || EXCLUDED.settings,
    updated_at = NOW();
```

The `||` operator merges the JSONB objects — new keys are added, existing keys are overwritten, unrelated keys are untouched.

See [Chapter 12: Bulk Operations](./12-bulk-operations.md) for the full bulk update workflow.

---

## Filtering Entities for Bulk Updates

The `GetEntitiesForBulkUpdate` query finds target entities using flexible filters:

```sql
-- name: GetEntitiesForBulkUpdate :many
SELECT uuid, name, type, settings
FROM entities
WHERE tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND (entity_types_filter IS NULL OR type = ANY(entity_types_filter))
  AND (tags_filter IS NULL OR EXISTS (
      SELECT 1 FROM jsonb_array_elements_text(tags) AS tag
      WHERE tag = ANY(tags_filter)
  ))
  AND (created_after IS NULL OR created_at >= created_after);
```

---

## Entity Configuration vs. Entity State

Do not confuse `entities.settings` (configuration) with `entitystate.config` (document sequence format). They serve different purposes:

| Column | Purpose |
|---|---|
| `entities.settings` | Business configuration overrides (approval limits, prefixes, methods) |
| `entitystate.config` | Document sequence formatting (pad length, reset frequency, format template) |

---

[← Tenant Configuration](./07-tenant-configuration.md) | [Next: Effective Resolution →](./09-effective-resolution.md)
