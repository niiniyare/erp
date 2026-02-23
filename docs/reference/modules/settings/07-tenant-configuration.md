# Chapter 7: Tenant Configuration

[← Configuration Definitions](./06-configuration-definitions.md) | [Next: Entity Configuration →](./08-entity-configuration.md)

---

## Overview

Tenant-level configuration represents the organization-wide defaults for a tenant. Every entity within the tenant inherits these values unless it explicitly overrides them. This is the primary configuration layer that administrators manage during onboarding and ongoing operations.

---

## Storage Structure

Each tenant has exactly one row in `tenant_configurations`:

```sql
tenant_id        | UUID (PK, FK → tenants)
settings         | JSONB  -- flat map of "module.key" → value
settings_version | INTEGER  -- increments on every write
last_template_applied | UUID
template_applied_at   | TIMESTAMPTZ
```

The `settings` column is a flat JSON object where keys are the full `module.config_key` string:

```json
{
  "finance.invoice_prefix":      "INV-",
  "finance.default_currency":    "USD",
  "finance.auto_approval_limit": 2000,
  "hr.overtime_threshold":       40,
  "inventory.valuation_method":  "FIFO"
}
```

---

## Reading Tenant Configuration

```sql
-- name: GetTenantConfigurations :one
SELECT settings, settings_version
FROM tenant_configurations
WHERE tenant_id = current_tenant_id();
```

RLS ensures `current_tenant_id()` is set from the session context. No explicit `tenant_id` filter is needed — the database enforces it.

---

## Writing a Tenant Configuration

Updates use `jsonb_set` to surgically update a single key without overwriting the entire blob:

```sql
-- name: UpdateTenantConfigurationSettings :exec
UPDATE tenant_configurations
SET
    settings = jsonb_set(
        COALESCE(settings, '{}'),
        $1::text[],         -- config_path, e.g., ARRAY['finance.invoice_prefix']
        $2::jsonb           -- config_value, e.g., '"INV-2025-"'
    ),
    updated_at = NOW(),
    settings_version = settings_version + 1
WHERE tenant_id = current_tenant_id();
```

Go call:

```go
err = store.UpdateTenantConfigurationSettings(ctx, db.UpdateTenantConfigurationSettingsParams{
    ConfigPath:  []string{"finance.invoice_prefix"},
    ConfigValue: []byte(`"INV-2025-"`),
})
```

`settings_version` increments atomically, supporting optimistic locking: callers can check the version before writing and reject stale updates.

---

## Deleting a Tenant Configuration

Deletion removes the key from the JSONB blob using the `#-` operator, causing the configuration to fall back to the system default:

```sql
-- name: DeleteTenantConfigurationSettings :exec
UPDATE tenant_configurations
SET
    settings = settings #- $1::text[],
    updated_at = NOW(),
    settings_version = settings_version + 1
WHERE tenant_id = current_tenant_id();
```

---

## Listing All Tenant Configurations

To see all currently set tenant configurations (all modules):

```sql
-- name: GetTenantConfigurations :one
SELECT settings, settings_version FROM tenant_configurations
WHERE tenant_id = current_tenant_id();
```

Parse the JSONB in the application layer to iterate over all `module.key` pairs.

---

## Template-Assisted Setup

The most common way to populate initial tenant configuration is by applying an industry template (see [Chapter 10: Template Management](./10-template-management.md)). This sets multiple keys at once and records the template application in `tenant_configurations.last_template_applied`.

---

## Versioning & Optimistic Locking

`settings_version` enables safe concurrent updates:

1. Client reads `settings` and `settings_version = 5`
2. Client computes new value
3. Client updates with `WHERE settings_version = 5`
4. If another writer changed it in between, the update affects 0 rows → client retries

---

[← Configuration Definitions](./06-configuration-definitions.md) | [Next: Entity Configuration →](./08-entity-configuration.md)
