# Chapter 14: API Reference

[← Audit & History](./13-audit-and-history.md) | [Next: Service Integration →](./15-service-integration.md)

---

## Base Information

- **Base URL**: `/api/v1/settings`
- **Authentication**: `Authorization: Bearer <jwt>` + `X-Tenant-ID: <uuid>`
- **Content-Type**: `application/json`
- **Version**: v1

### Required Headers

| Header | Required | Description |
|--------|----------|-------------|
| `Authorization` | Yes | JWT bearer token |
| `X-Tenant-ID` | Yes | Tenant context |
| `Content-Type` | Yes | `application/json` for POST/PUT |

---

## Configuration Endpoints

### GET `/config/{module}/{config_key}`

Resolve a configuration value through the inheritance hierarchy.

**Query Parameters**:
| Param | Type | Description |
|-------|------|-------------|
| `entity_id` | UUID | Entity context for resolution |
| `include_metadata` | boolean | Include inheritance chain |

**Response**:
```json
{
  "module": "finance",
  "config_key": "invoice_prefix",
  "value": "INV-2025-",
  "data_type": "string",
  "source": "entity",
  "is_inherited": false,
  "inheritance_chain": { "system": "DOC-", "tenant": "INV-", "entity": "INV-2025-" }
}
```

---

### PUT `/config/{module}/{config_key}`

Update a configuration value at tenant or entity level.

**Request Body**:
```json
{
  "value": "BRANCH-INV-2025-",
  "target_type": "entity",
  "target_id": "entity-uuid",
  "data_type": "string"
}
```

---

### DELETE `/config/{module}/{config_key}`

Reset a configuration to its inherited value.

**Query Parameters**: `target_type` (tenant|entity), `target_id` (UUID)

---

### GET `/modules/{module}/config`

List all configurations for a module with inheritance resolution.

**Query Parameters**: `entity_id`, `source` (system|tenant|entity), `include_inherited`, `include_defaults`

---

### POST `/resolve`

Batch resolve multiple configurations for a context.

**Request Body**:
```json
{
  "context": { "tenant_id": "uuid", "entity_id": "uuid" },
  "configurations": [
    { "module": "finance", "config_key": "invoice_prefix" }
  ],
  "options": { "include_inheritance_chain": true }
}
```

---

### GET `/effective`

Get the complete effective configuration set for an entity.

**Query Parameters**: `entity_id` (required), `modules[]`, `changed_since`

---

## Template Endpoints

### GET `/templates`

List available templates.

**Query Parameters**: `category`, `scope`, `applicable_to`, `search`

---

### GET `/templates/{template_id}`

Get template details including all configuration values.

---

### POST `/templates/{template_id}/apply`

Apply a template to a tenant or entity.

**Request Body**:
```json
{
  "target_type": "tenant",
  "target_id": "uuid",
  "options": {
    "preserve_existing": true,
    "conflict_resolution": "merge",
    "dry_run": false
  },
  "selective_application": {
    "modules": ["finance"],
    "exclude_keys": ["finance.bank_account"]
  }
}
```

---

## Bulk Operations Endpoints

### POST `/bulk-update`

Update multiple entities. Returns operation ID for async polling.

**Request Body**:
```json
{
  "targets": {
    "type": "entity",
    "filters": { "entity_types": ["BRANCH"], "tags": ["retail"] }
  },
  "configurations": [
    { "module": "finance", "config_key": "approval_limit", "value": 2500 }
  ],
  "options": { "batch_size": 100, "dry_run": false }
}
```

---

### GET `/operations/{operation_id}`

Check bulk operation status and progress.

---

## Search Endpoint

### POST `/search`

Advanced configuration search.

**Request Body**:
```json
{
  "criteria": {
    "modules": ["finance"],
    "value_contains": "INV",
    "sources": ["entity"],
    "modified_after": "2025-08-01T00:00:00Z"
  },
  "pagination": { "page": 1, "limit": 50 }
}
```

---

### GET `/analysis/usage`

Configuration usage analysis across the tenant.

**Query Parameters**: `module`, `period` (1d|7d|30d)

---

## Error Codes

| HTTP | Code | Description |
|------|------|-------------|
| 400 | `validation_failed` | Request validation failed |
| 400 | `invalid_configuration_value` | Value invalid for data type |
| 403 | `configuration_not_overridable` | Cannot override this key |
| 403 | `feature_flag_disabled` | Required feature flag not enabled |
| 404 | `configuration_not_found` | Configuration not found |
| 404 | `template_not_found` | Template not found |
| 409 | `version_conflict` | Optimistic locking conflict |
| 422 | `template_dependency_missing` | Template dependencies not satisfied |

---

## Permissions

| Permission | Description |
|---|---|
| `settings:system:read` | View system-level configurations |
| `settings:tenant:write` | Modify tenant-level configurations |
| `settings:entity:write` | Modify entity-level configurations |
| `settings:templates:apply` | Apply configuration templates |
| `settings:bulk:write` | Perform bulk configuration operations |

---

[← Audit & History](./13-audit-and-history.md) | [Next: Service Integration →](./15-service-integration.md)
