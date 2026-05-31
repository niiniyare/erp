# Chapter 4: Configuration Hierarchy & Inheritance

[← Architecture Overview](./03-architecture-overview.md) | [Next: Data Model →](./05-data-model.md)

---

## Three-Level Hierarchy

Every configuration value is resolved through three levels. The most specific level wins:

```
Level 0 — System (Platform Default)
│   Stored in: config_definitions.default_value
│   Set by: Platform engineers
│   Scope: All tenants, all entities
│
├── Level 1 — Tenant (Organization Override)
│   Stored in: tenant_configurations.settings JSONB
│   Set by: Tenant administrators
│   Scope: All entities of this tenant
│
│   └── Level 2 — Entity (Branch/Location Override)
│       Stored in: entities.settings JSONB
│       Set by: Entity managers
│       Scope: This specific entity only
```

### Resolution Rule

> **Entity value** if present, else **Tenant value** if present, else **System default**.

This is enforced in the database by the `GetEffectiveConfiguration` query using `ORDER BY priority DESC LIMIT 1`:

```sql
WITH RECURSIVE config_resolution AS (
    -- priority 0: system default
    SELECT cd.default_value as value, 'system' as source, 0 as priority ...
    UNION ALL
    -- priority 1: tenant override
    SELECT tc.settings -> (module || '.' || key) as value, 'tenant', 1 ...
    UNION ALL
    -- priority 2: entity override
    SELECT e.settings -> (module || '.' || key) as value, 'entity', 2 ...
)
SELECT value, source FROM config_resolution
ORDER BY priority DESC
LIMIT 1;
```

---

## Inheritance in Practice

### Example: Invoice Prefix

```
System default:     invoice_prefix = "DOC-"
Tenant A override:  invoice_prefix = "INV-"
  ├── Branch NYC (no override) → resolves "INV-"   (inherits tenant)
  ├── Branch LA  (overrides)   → resolves "LA-INV-" (entity wins)
  └── Branch CHI (no override) → resolves "INV-"   (inherits tenant)

Tenant B (no override):
  └── All entities            → resolves "DOC-"    (inherits system)
```

### Example: Approval Limit

```
System default:  auto_approval_limit = 500
Tenant A:        auto_approval_limit = 2000
  ├── HQ entity: auto_approval_limit = 10000 (executive override)
  └── Branches:  auto_approval_limit = 2000  (inherit tenant)
```

---

## Configuration Key Format

All configuration keys follow the `module_name.config_key` dot notation:

```
finance.invoice_prefix
finance.auto_approval_limit
finance.default_currency
hr.overtime_threshold
hr.default_pay_frequency
inventory.valuation_method
inventory.allow_negative_stock
```

In JSONB storage, the full key is used as the object key:

```json
{
  "finance.invoice_prefix": "INV-",
  "finance.auto_approval_limit": 2000,
  "inventory.valuation_method": "FIFO"
}
```

In the `configuration_audit` table, the key is split into `module_name` and `config_key_name` separate columns for indexed querying.

---

## What Cannot Be Overridden

When `is_overridable = false` on a config definition (planned feature), lower levels cannot override the system default. Currently the system uses `required_feature_flag` to gate configuration access — if the tenant's feature flag is not enabled, the configuration key is not available.

---

## Relationship to Entity Hierarchy

Entities themselves form a hierarchy (Company → Division → Branch → Location) stored in the `hierarchy_paths` closure table. Configuration inheritance, however, does **not** traverse the entity hierarchy — it only uses the three levels (System, Tenant, Entity). A branch does not automatically inherit from its parent division's entity-level settings; it inherits from the tenant level.

---

[← Architecture Overview](./03-architecture-overview.md) | [Next: Data Model →](./05-data-model.md)
