# Chapter 10: Template Management

[← Effective Resolution](./09-effective-resolution.md) | [Next: Template Application →](./11-template-application.md)

---

## What Are Configuration Templates?

Configuration templates are named, versioned bundles of configuration key-value pairs. They enable fast, consistent, and repeatable configuration deployment — especially useful during tenant onboarding or rolling out new module configurations across many entities.

---

## Template Scopes

Templates use the same `SYSTEM` / `TENANT` scope pattern as modules and actions:

| Scope | Owned By | Visible To | Use Case |
|---|---|---|---|
| `SYSTEM` | Platform (tenant_id = NULL) | All tenants | Industry templates, standard configs |
| `TENANT` | A specific tenant | That tenant only | Custom internal templates |

The database enforces this with a constraint:
```sql
CONSTRAINT templates_scope_tenant_check CHECK (
    (scope = 'SYSTEM' AND tenant_id IS NULL)
    OR (scope = 'TENANT' AND tenant_id IS NOT NULL)
)
```

RLS policies allow tenants to read `SYSTEM` templates and their own `TENANT` templates, but never other tenants' custom templates.

---

## Template Structure

```json
{
  "id": "template-uuid",
  "scope": "SYSTEM",
  "tenant_id": null,
  "name": "Manufacturing Standard",
  "category": "industry",
  "version": "1.2",
  "configurations": {
    "finance.invoice_prefix":       "MFG-INV-",
    "finance.auto_approval_limit":  5000,
    "inventory.valuation_method":   "FIFO",
    "inventory.lot_tracking_required": true,
    "hr.overtime_threshold":        40
  },
  "applicable_tenant_types": ["manufacturing", "industrial"],
  "required_feature_flags":  ["inventory.lot_tracking"],
  "conflict_resolution":     "merge",
  "is_active": true
}
```

---

## Creating a Template

### SYSTEM Template (Platform Engineers)

```sql
-- name: CreateConfigurationTemplate :one
INSERT INTO configuration_templates (
    tenant_id, scope, name, category, description, version, configurations,
    applicable_tenant_types, required_feature_flags, conflict_resolution, created_by
) VALUES (
    NULL, 'SYSTEM', 'Manufacturing Standard', 'industry',
    'Standard configuration for manufacturing companies',
    '1.2', $configurations, ARRAY['manufacturing'], ARRAY['inventory.lot_tracking'],
    'merge', $created_by
) RETURNING *;
```

### TENANT Template (Custom)

```go
result, err := store.CreateConfigurationTemplate(ctx, db.CreateConfigurationTemplateParams{
    TenantID:              &currentTenantID,    // non-nil for TENANT scope
    Scope:                 "TENANT",
    Name:                  "Retail Branch Standard",
    Category:              "operational",
    Configurations:        branchConfigJSON,
    ConflictResolution:    "preserve",
    CreatedBy:             userID,
})
```

---

## Listing Templates

The `ListConfigurationTemplates` query returns SYSTEM templates plus the current tenant's TENANT templates in one query:

```sql
-- name: ListConfigurationTemplates :many
SELECT * FROM configuration_templates
WHERE is_active = true
  AND (scope_filter IS NULL    OR scope    = scope_filter)
  AND (category_filter IS NULL OR category = category_filter)
  AND (tenant_types_filter IS NULL OR applicable_tenant_types && tenant_types_filter)
ORDER BY scope DESC, name;  -- TENANT before SYSTEM
```

RLS automatically filters out other tenants' TENANT-scope templates. The `ORDER BY scope DESC` puts TENANT templates first (allowing tenant customizations to appear above platform defaults in UIs).

---

## Conflict Resolution Strategies

When applying a template, existing configuration values may conflict:

| Strategy | Behavior |
|---|---|
| `merge` | Apply all template values; overwrite existing values |
| `preserve` | Skip keys that already have a value; only set unset keys |
| `replace` | Wipe all existing config in scope, then apply template |

The strategy can be set at the template level (`conflict_resolution` column) or overridden per-application request.

---

## Template Versioning

Templates use semantic versioning. When a template needs to be updated, create a new version rather than modifying in place:

```sql
-- name: UpdateConfigurationTemplate :one
UPDATE configuration_templates
SET
    name          = COALESCE($name, name),
    configurations = COALESCE($configurations, configurations),
    updated_at    = NOW()
WHERE id = $template_id AND is_active = true
RETURNING *;
```

Old versions can be deactivated with `DeactivateConfigurationTemplate` while keeping them in the audit history.

---

[← Effective Resolution](./09-effective-resolution.md) | [Next: Template Application →](./11-template-application.md)
