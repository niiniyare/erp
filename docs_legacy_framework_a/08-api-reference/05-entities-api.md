> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Entities API
portal: 8 — API Reference
section: 08-api-reference
audience: [backend-engineer, frontend-engineer, integrator]
related:
  - "[API Reference Overview](01-api-reference-overview.md)"
  - "[Tenancy Model](../03-platform-architecture/01-multi-tenancy/01-tenancy-model.md)"
---

# Entities API

Entities represent organizational units: companies, subsidiaries, departments, cost centers, branches. They form a hierarchy. All entity endpoints are tenant-scoped — `tenant_id` is always sourced from the session.

## Entity Types

| Type | Description |
|------|-------------|
| `COMPANY` | Top-level legal entity |
| `SUBSIDIARY` | Child of a company |
| `DEPARTMENT` | Functional unit within any entity |
| `COST_CENTER` | Budget center, may span departments |
| `BRANCH` | Geographic location |

## POST /api/v1/entities

Create a new entity.

### Request

```http
POST /api/v1/entities
Authorization: Bearer <token>
Content-Type: application/json

{
  "code": "HQ",
  "name": "Headquarters",
  "entity_type": "COMPANY",
  "parent_id": null,
  "currency": "USD",
  "country": "US",
  "timezone": "America/New_York",
  "description": "Main holding company"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `code` | string | yes | Short unique code within tenant |
| `name` | string | yes | Display name |
| `entity_type` | string | yes | See Entity Types table |
| `parent_id` | uuid | no | Parent entity — null for root |
| `currency` | string | yes | ISO 4217 |
| `country` | string | yes | ISO 3166-1 alpha-2 |
| `timezone` | string | yes | IANA timezone |
| `description` | string | no | Free-text |

### Response `201 Created`

```json
{
  "data": {
    "id": "018e1b2c-3d4e-7f8a-9b0c-1d2e3f4a5b6c",
    "code": "HQ",
    "name": "Headquarters",
    "entity_type": "COMPANY",
    "parent_id": null,
    "depth": 0,
    "path": "/018e1b2c-3d4e-7f8a-9b0c-1d2e3f4a5b6c",
    "status": "active",
    "currency": "USD",
    "country": "US",
    "timezone": "America/New_York",
    "created_at": "2026-01-15T10:00:00Z"
  }
}
```

### Error Responses

| Status | Code | Condition |
|--------|------|-----------|
| 409 | `CONFLICT` | `code` already exists in tenant |
| 404 | `NOT_FOUND` | `parent_id` not found |
| 422 | `VALIDATION_ERROR` | Invalid field values |

---

## GET /api/v1/entities

List entities with optional hierarchy filter.

### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `entity_type` | string | — | Filter by type |
| `parent_id` | uuid | — | Direct children only |
| `root` | bool | — | Root entities only (no parent) |
| `status` | string | `active` | `active`, `inactive`, `archived` |
| `limit` | int | 20 | Max 100 |
| `offset` | int | 0 | — |

### Response `200 OK`

```json
{
  "data": [
    {
      "id": "018e1b2c-...",
      "code": "HQ",
      "name": "Headquarters",
      "entity_type": "COMPANY",
      "parent_id": null,
      "depth": 0,
      "child_count": 3
    },
    {
      "id": "028e2c3d-...",
      "code": "EMEA",
      "name": "EMEA Region",
      "entity_type": "SUBSIDIARY",
      "parent_id": "018e1b2c-...",
      "depth": 1,
      "child_count": 8
    }
  ],
  "meta": {
    "total": 47,
    "limit": 20,
    "offset": 0
  }
}
```

---

## GET /api/v1/entities/:id

Get entity by ID.

### Response `200 OK`

Full entity object including `path` (materialized path), `depth`, and `child_count`.

---

## GET /api/v1/entities/:id/subtree

Get full subtree rooted at entity — all descendants at all depths.

### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `max_depth` | int | — | Limit tree depth (1 = children only) |
| `status` | string | `active` | Include only entities with this status |

### Response `200 OK`

```json
{
  "data": {
    "id": "018e1b2c-...",
    "name": "Headquarters",
    "children": [
      {
        "id": "028e2c3d-...",
        "name": "EMEA Region",
        "children": [
          { "id": "038f3d4e-...", "name": "UK Office", "children": [] }
        ]
      }
    ]
  }
}
```

---

## GET /api/v1/entities/:id/ancestors

Get ordered list of ancestors from root to entity (breadcrumb path).

### Response `200 OK`

```json
{
  "data": [
    { "id": "018e1b2c-...", "code": "HQ", "name": "Headquarters", "depth": 0 },
    { "id": "028e2c3d-...", "code": "EMEA", "name": "EMEA Region", "depth": 1 },
    { "id": "038f3d4e-...", "code": "UK", "name": "UK Office", "depth": 2 }
  ]
}
```

---

## PUT /api/v1/entities/:id

Update entity fields. Cannot change `entity_type` or `parent_id` (hierarchy changes use a separate endpoint).

### Request

```json
{
  "name": "Global Headquarters",
  "description": "Updated description"
}
```

---

## POST /api/v1/entities/:id/move

Move entity to a new parent. Updates materialized path for entire subtree.

### Request

```json
{
  "new_parent_id": "028e2c3d-..."
}
```

### Response `200 OK`

```json
{
  "data": {
    "id": "038f3d4e-...",
    "parent_id": "028e2c3d-...",
    "depth": 2,
    "path": "/018e1b2c.../028e2c3d-.../038f3d4e-..."
  }
}
```

### Error Responses

| Status | Code | Condition |
|--------|------|-----------|
| 409 | `CONFLICT` | `new_parent_id` is a descendant of this entity (cycle) |

---

## DELETE /api/v1/entities/:id

Soft-delete entity. Only allowed if entity has no active children and no active contracts.

### Error Responses

| Status | Code | Condition |
|--------|------|-----------|
| 409 | `CONFLICT` | Entity has active children or contracts |
