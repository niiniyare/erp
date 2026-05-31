# Chapter 5: Data Model & Storage

[← Configuration Hierarchy](./04-configuration-hierarchy.md) | [Next: Configuration Definitions →](./06-configuration-definitions.md)

---

## Database Schema Overview

```sql
-- Schema: public (multi-tenant with RLS)
config_definitions          -- Schema registry for all configurable keys
tenant_configurations       -- Per-tenant settings blob + version tracking
entities                    -- settings JSONB column on shared entity table
configuration_templates     -- Named versioned config bundles
template_applications       -- History of template apply operations
configuration_audit         -- Immutable change log
```

---

## config_definitions

The master registry of every configuration key in the platform.

```sql
CREATE TABLE config_definitions (
    module_name          VARCHAR(50)  NOT NULL,
    config_key           VARCHAR(100) NOT NULL,
    config_type          VARCHAR(20)  NOT NULL,  -- string|integer|boolean|decimal|json
    default_value        JSONB,                  -- System-level default
    validation_rules     JSONB,                  -- JSON Schema or custom rules
    description          TEXT,
    required_feature_flag VARCHAR(100),          -- Gate key on a feature flag
    created_at           TIMESTAMPTZ  DEFAULT NOW(),
    updated_at           TIMESTAMPTZ  DEFAULT NOW(),
    PRIMARY KEY (module_name, config_key)
);
```

Key points:
- Composite primary key `(module_name, config_key)` — no surrogate ID needed
- `default_value` is JSONB so it can hold any type (string, number, boolean, array, object)
- `required_feature_flag` gates the key: if the tenant doesn't have the flag enabled, the key is unavailable

---

## tenant_configurations

One row per tenant. Settings stored as a flat JSONB map keyed by `module.key`.

```sql
CREATE TABLE tenant_configurations (
    tenant_id            UUID        PRIMARY KEY REFERENCES tenants(id),
    settings             JSONB       DEFAULT '{}'::jsonb,
    settings_version     INTEGER     DEFAULT 0,  -- Optimistic locking
    last_template_applied UUID,                   -- FK to configuration_templates
    template_applied_at  TIMESTAMPTZ,
    created_at           TIMESTAMPTZ DEFAULT NOW(),
    updated_at           TIMESTAMPTZ DEFAULT NOW()
);
```

Settings JSONB example:
```json
{
  "finance.invoice_prefix":      "INV-",
  "finance.auto_approval_limit": 2000,
  "finance.default_currency":    "USD",
  "hr.overtime_threshold":       40,
  "inventory.valuation_method":  "FIFO"
}
```

`settings_version` increments on every write, enabling optimistic locking for concurrent updates.

---

## entities.settings (Entity-Level Config)

The `entities` table (owned by the Entity module) has a `settings` JSONB column that stores entity-specific configuration overrides. The Settings Module reads and writes this column via the `GetEntityConfigurations`, `UpdateEntityConfiguration`, and `DeleteEntityConfiguration` queries.

```sql
-- Relevant column on entities table
settings   JSONB DEFAULT '{}'::jsonb
```

The same flat `module.key` format is used as in `tenant_configurations.settings`.

---

## configuration_templates

Versioned bundles of configuration key-value pairs.

```sql
CREATE TABLE configuration_templates (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scope                 VARCHAR(10) NOT NULL DEFAULT 'TENANT'
                            CHECK (scope IN ('SYSTEM', 'TENANT')),
    tenant_id             UUID REFERENCES tenants(id),   -- NULL for SYSTEM scope
    name                  VARCHAR(255) NOT NULL,
    category              VARCHAR(50),
    description           TEXT,
    version               VARCHAR(20) NOT NULL,
    configurations        JSONB NOT NULL,
    applicable_tenant_types TEXT[],
    required_feature_flags  TEXT[],
    conflict_resolution   VARCHAR(20) DEFAULT 'merge'
                            CHECK (conflict_resolution IN ('merge', 'replace', 'preserve')),
    is_active             BOOLEAN DEFAULT TRUE,
    created_by            UUID,
    created_at            TIMESTAMPTZ DEFAULT NOW(),
    updated_at            TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT templates_scope_tenant_check CHECK (
        (scope = 'SYSTEM' AND tenant_id IS NULL)
        OR (scope = 'TENANT' AND tenant_id IS NOT NULL)
    )
);
```

---

## configuration_audit

Immutable audit log. Never updated after insert.

```sql
CREATE TABLE configuration_audit (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    entity_id       UUID,                -- NULL for tenant-level changes
    module_name     VARCHAR(50)  NOT NULL,
    config_key_name VARCHAR(100) NOT NULL,
    old_value       JSONB,
    new_value       JSONB NOT NULL,
    source          VARCHAR(20),         -- "tenant" | "entity" | "template"
    operation       VARCHAR(20),         -- "SET" | "DELETE" | "TEMPLATE_APPLY"
    user_id         UUID NOT NULL,
    session_id      UUID,
    correlation_id  UUID,
    applied_at      TIMESTAMPTZ DEFAULT NOW()
);
```

`module_name` and `config_key_name` are stored as separate columns (not a single `module.key` string) to enable indexed queries by module or key independently.

---

[← Configuration Hierarchy](./04-configuration-hierarchy.md) | [Next: Configuration Definitions →](./06-configuration-definitions.md)
