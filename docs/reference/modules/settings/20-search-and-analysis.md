# Chapter 20: Search & Configuration Analysis

[← Event-Driven Integration](./19-event-driven-integration.md) | [Next: Common Scenarios →](./21-common-scenarios.md)

---

## Configuration Search

The `SearchConfigurations` query scans across system, tenant, and entity configurations in one operation, supporting full-text search on keys, filtering by module, source, type, and modification date.

### Search Request

```http
POST /api/v1/settings/search

{
  "criteria": {
    "modules":        ["finance", "inventory"],
    "value_contains": "INV",
    "sources":        ["entity", "tenant"],
    "data_types":     ["string"],
    "modified_after": "2025-08-01T00:00:00Z"
  },
  "scope": {
    "tenant_id":  "tenant-uuid",
    "entity_ids": ["entity-1", "entity-2"]
  },
  "sort": [
    { "field": "last_modified", "order": "desc" }
  ],
  "pagination": { "page": 1, "limit": 50 }
}
```

### Search Response

```json
{
  "results": [
    {
      "module":        "finance",
      "config_key":    "invoice_prefix",
      "value":         "BRANCH-INV-",
      "data_type":     "string",
      "source":        "entity",
      "entity_id":     "entity-1",
      "last_modified": "2025-09-01T14:30:00Z",
      "modified_by":   "user-456"
    }
  ],
  "pagination": { "total": 25, "page": 1, "limit": 50 },
  "search_metadata": { "query_time_ms": 45, "total_scanned": 1250 }
}
```

---

## Database Implementation

```sql
-- name: SearchConfigurations :many
WITH all_configs AS (
    -- System configurations
    SELECT cd.module_name, cd.config_key,
           cd.module_name || '.' || cd.config_key as full_key,
           cd.default_value as value, 'system' as source,
           NULL::UUID as entity_id, cd.config_type, cd.updated_at
    FROM config_definitions cd

    UNION ALL

    -- Tenant configurations
    SELECT split_part(key, '.', 1) as module_name,
           split_part(key, '.', 2) as config_key,
           key as full_key, value, 'tenant' as source,
           NULL::UUID, cd.config_type, tc.updated_at
    FROM tenant_configurations tc,
         jsonb_each(tc.settings) as setting(key, value)
    JOIN config_definitions cd ON cd.module_name = split_part(key, '.', 1)
    WHERE tc.tenant_id = current_tenant_id()

    UNION ALL

    -- Entity configurations
    SELECT split_part(key, '.', 1), split_part(key, '.', 2),
           key, value, 'entity', e.uuid, cd.config_type, e.updated_at
    FROM entities e,
         jsonb_each(e.settings) as setting(key, value)
    JOIN config_definitions cd ON cd.module_name = split_part(key, '.', 1)
    WHERE e.tenant_id = current_tenant_id()
      AND e.deleted_at IS NULL
      AND (entity_id_filter IS NULL OR e.uuid = entity_id_filter)
)
SELECT * FROM all_configs
WHERE (modules_filter IS NULL OR module_name = ANY(modules_filter))
  AND (search_term IS NULL OR full_key ILIKE '%' || search_term || '%')
  AND (sources_filter IS NULL OR source = ANY(sources_filter))
  AND (updated_after IS NULL OR updated_at >= updated_after)
ORDER BY
    CASE WHEN sort_by = 'module' THEN module_name END,
    CASE WHEN sort_by = 'updated_at' THEN updated_at END DESC,
    full_key
LIMIT $limit OFFSET $offset;
```

---

## Configuration Usage Analysis

Understand which configurations are most commonly overridden across the tenant:

```http
GET /api/v1/settings/analysis/usage?module=finance&period=30d
```

```json
{
  "analysis_period": "30d",
  "module_summary": [
    {
      "module": "finance",
      "total_configs": 45,
      "entity_overrides": 123,
      "most_customized": [
        {
          "config_key": "invoice_prefix",
          "override_percentage": 85,
          "entity_count": 34
        }
      ]
    }
  ],
  "configuration_patterns": {
    "most_overridden":  ["finance.invoice_prefix", "inventory.reorder_threshold"],
    "never_overridden": ["finance.default_currency", "hr.base_pay_frequency"]
  },
  "template_adoption": {
    "total_applications": 15,
    "most_applied_template": "Manufacturing Standard",
    "recent_applications": 3
  }
}
```

---

## Finding Drift Between Entities

Use search to find entities that have diverged from tenant defaults on a specific key:

```bash
# Find all entity-level overrides for invoice_prefix
POST /api/v1/settings/search
{
  "criteria": {
    "modules":  ["finance"],
    "sources":  ["entity"]
  }
}
```

Then compare with the tenant value to identify which entities have customized it.

---

## Admin Use Cases for Search

| Use Case | Search Criteria |
|---|---|
| Find all entities using non-standard currency | `module=finance`, `key_contains=currency`, `source=entity` |
| Find configs modified in last week | `modified_after=<7d ago>` |
| Find all configurations for a specific entity | `entity_id=<uuid>`, all sources |
| Identify unused system defaults | All `source=system` not overridden at any level |
| Audit template adoption gaps | Keys from template that remain at system default |

---

[← Event-Driven Integration](./19-event-driven-integration.md) | [Next: Common Scenarios →](./21-common-scenarios.md)
