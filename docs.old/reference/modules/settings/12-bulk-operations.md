# Chapter 12: Bulk Operations

[← Template Application](./11-template-application.md) | [Next: Audit & History →](./13-audit-and-history.md)

---

## Overview

Bulk operations allow updating configuration values across many entities in a single request. This is essential for scenarios like:

- Rolling out a new approval limit to all branches of a region
- Enabling a feature flag configuration for all retail locations
- Standardizing valuation methods across a warehouse network

---

## Bulk Update Request

```http
POST /api/v1/settings/bulk-update
Authorization: Bearer <token>
X-Tenant-ID: <tenant-id>

{
  "targets": {
    "type": "entity",
    "filters": {
      "entity_types": ["BRANCH", "LOCATION"],
      "tags": ["retail"],
      "created_after": "2025-01-01"
    },
    "explicit_ids": ["entity-1", "entity-2"]
  },
  "configurations": [
    { "module": "finance", "config_key": "approval_limit",  "value": 2500 },
    { "module": "inventory", "config_key": "reorder_enabled", "value": true }
  ],
  "options": {
    "preserve_explicit_overrides": true,
    "batch_size": 100,
    "dry_run": false,
    "continue_on_error": true
  }
}
```

---

## Response (Async)

Large bulk operations run asynchronously and return an operation ID:

```json
{
  "operation_id": "bulk-op-789",
  "status": "in_progress",
  "targets": { "total_entities": 156, "batches": 2 },
  "configurations": 2,
  "estimated_duration": "45 seconds",
  "progress_url": "/api/v1/settings/operations/bulk-op-789",
  "started_at": "2025-09-01T17:00:00Z"
}
```

Poll for completion:

```http
GET /api/v1/settings/operations/bulk-op-789
```

```json
{
  "status": "completed",
  "progress": {
    "total_targets": 156, "processed": 156,
    "successful": 154, "failed": 2, "percentage": 100
  },
  "summary": {
    "total_configurations": 312, "updated": 308, "skipped": 2, "errors": 2
  }
}
```

---

## Database Implementation

### Finding Target Entities

```sql
-- name: GetEntitiesForBulkUpdate :many
SELECT uuid, name, type, settings
FROM entities
WHERE tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND (entity_types_filter IS NULL OR type = ANY(entity_types_filter::TEXT[]))
  AND (tags_filter IS NULL OR EXISTS (
      SELECT 1 FROM jsonb_array_elements_text(tags) AS tag
      WHERE tag = ANY(tags_filter::TEXT[])
  ))
  AND (created_after IS NULL OR created_at >= created_after);
```

### Applying Configuration

```sql
-- name: BulkUpdateEntityConfiguration :exec
INSERT INTO entities (uuid, tenant_id, settings, updated_at)
VALUES ($1, current_tenant_id(), $2, NOW())
ON CONFLICT (uuid, tenant_id)
DO UPDATE SET
    settings = entities.settings || EXCLUDED.settings,
    updated_at = NOW();
```

The `||` operator merges JSONB objects: existing keys not in `EXCLUDED.settings` are preserved; keys in both are overwritten with the new value.

---

## Batching

The application layer batches entity updates:

```go
func (s *BulkUpdateService) ProcessBulkUpdate(
    ctx context.Context, req BulkUpdateRequest,
) (*BulkUpdateResult, error) {
    // Find all target entities
    entities, err := s.store.GetEntitiesForBulkUpdate(ctx, req.Filters)

    // Process in batches to avoid locking and memory issues
    batchSize := req.Options.BatchSize
    if batchSize == 0 {
        batchSize = 100
    }

    result := &BulkUpdateResult{}
    for i := 0; i < len(entities); i += batchSize {
        batch := entities[i:min(i+batchSize, len(entities))]
        batchResult, err := s.processBatch(ctx, batch, req.Configurations)
        if err != nil && !req.Options.ContinueOnError {
            return nil, err
        }
        result.Merge(batchResult)
    }

    return result, nil
}
```

---

## `preserve_explicit_overrides` Option

When `true`, entities that have an explicit override for a configuration key will not be overwritten by the bulk update. Only entities inheriting the tenant default (no explicit override) will be updated.

This is useful for rolling out a new tenant default while respecting branches that intentionally diverged.

---

## Audit Trail

Each entity updated in a bulk operation generates an audit record in `configuration_audit`, all sharing the same `correlation_id` — making it easy to query all changes from a single bulk operation:

```sql
SELECT * FROM configuration_audit
WHERE correlation_id = 'bulk-op-789'
ORDER BY applied_at;
```

---

[← Template Application](./11-template-application.md) | [Next: Audit & History →](./13-audit-and-history.md)
