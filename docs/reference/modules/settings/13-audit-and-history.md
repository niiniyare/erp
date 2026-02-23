# Chapter 13: Audit & History

[← Bulk Operations](./12-bulk-operations.md) | [Next: API Reference →](./14-api-reference.md)

---

## Why a Separate Audit Table?

The `configuration_audit` table is dedicated to configuration change history. It exists separately from the platform `audit_logs` table (which is a security/IAM audit trail) because:

- Configuration audits have different retention requirements
- Configuration audits need indexed querying by module and key
- Configuration audits carry domain-specific context (old value, new value, source)
- Compliance teams query configuration history independently from access logs

---

## Audit Record Structure

```sql
CREATE TABLE configuration_audit (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    entity_id       UUID,            -- NULL = tenant-level change
    module_name     VARCHAR(50) NOT NULL,
    config_key_name VARCHAR(100) NOT NULL,
    old_value       JSONB,           -- NULL for first-time set
    new_value       JSONB NOT NULL,
    source          VARCHAR(20),     -- "tenant" | "entity" | "template"
    operation       VARCHAR(20),     -- "SET" | "DELETE" | "TEMPLATE_APPLY"
    user_id         UUID NOT NULL,
    session_id      UUID,
    correlation_id  UUID,            -- Links related changes (e.g., bulk ops)
    applied_at      TIMESTAMPTZ DEFAULT NOW()
);
```

### Why Split `module_name` + `config_key_name`?

Storing them as separate indexed columns instead of a single `"finance.invoice_prefix"` string enables:

```sql
-- Efficient: uses idx_configuration_audit_module
WHERE module_name = 'finance'

-- Efficient: uses idx_configuration_audit_key
WHERE config_key_name = 'invoice_prefix'

-- Both: both indexes can be combined
WHERE module_name = 'finance' AND config_key_name = 'invoice_prefix'
```

A single string column would require `LIKE 'finance.%'` which is not index-friendly.

---

## Creating an Audit Record

Every configuration write creates an audit record:

```sql
-- name: CreateConfigurationAudit :one
INSERT INTO configuration_audit (
    tenant_id, entity_id, module_name, config_key_name,
    old_value, new_value, source, operation,
    user_id, session_id, correlation_id
) VALUES (
    current_tenant_id(), $1, $2, $3,
    $4, $5, $6, $7, $8, $9, $10
) RETURNING *;
```

In the repository layer, the `module.key` string is split before inserting:

```go
parts := strings.SplitN(record.ConfigKey, ".", 2)
moduleName := parts[0]
configKeyName := ""
if len(parts) == 2 {
    configKeyName = parts[1]
}

_, err = r.store.CreateConfigurationAudit(ctx, db.CreateConfigurationAuditParams{
    EntityID:      record.EntityID,
    ModuleName:    moduleName,
    ConfigKeyName: configKeyName,
    OldValue:      record.OldValue,
    NewValue:      record.NewValue,
    Source:        record.Source,
    Operation:     record.Operation,
    UserID:        record.UserID,
    CorrelationID: record.CorrelationID,
})
```

---

## Querying History

### By Module and Key

```sql
-- name: GetConfigurationHistory :many
SELECT * FROM configuration_audit
WHERE tenant_id = current_tenant_id()
  AND (entity_id_filter IS NULL OR entity_id = entity_id_filter)
  AND (module_filter IS NULL OR module_name = module_filter)
  AND (config_key_filter IS NULL OR config_key_name = config_key_filter)
ORDER BY applied_at DESC
LIMIT $limit OFFSET $offset;
```

### By Correlation ID (Bulk/Template Operations)

```sql
-- name: GetConfigurationAuditByCorrelation :many
SELECT * FROM configuration_audit
WHERE tenant_id = current_tenant_id()
  AND correlation_id = $1
ORDER BY applied_at DESC;
```

This retrieves all changes made in a single bulk operation or template application.

---

## Indexes

```sql
CREATE INDEX idx_configuration_audit_tenant    ON configuration_audit(tenant_id, applied_at DESC);
CREATE INDEX idx_configuration_audit_entity    ON configuration_audit(entity_id) WHERE entity_id IS NOT NULL;
CREATE INDEX idx_configuration_audit_module    ON configuration_audit(module_name);
CREATE INDEX idx_configuration_audit_key       ON configuration_audit(config_key_name);
CREATE INDEX idx_configuration_audit_correlation ON configuration_audit(correlation_id) WHERE correlation_id IS NOT NULL;
```

---

## Template Application History

Template applications are separately tracked in `template_applications`:

```sql
-- name: GetTemplateApplicationHistory :many
SELECT
    ta.*,
    ct.name as template_name,
    ct.version as template_version
FROM template_applications ta
JOIN configuration_templates ct ON ct.id = ta.template_id
WHERE ta.tenant_id = current_tenant_id()
  AND (entity_id_filter IS NULL OR ta.entity_id = entity_id_filter)
ORDER BY ta.applied_at DESC
LIMIT $limit OFFSET $offset;
```

This gives the full picture: when a template was applied, which version, how many configs were set, and how many were skipped.

---

[← Bulk Operations](./12-bulk-operations.md) | [Next: API Reference →](./14-api-reference.md)
