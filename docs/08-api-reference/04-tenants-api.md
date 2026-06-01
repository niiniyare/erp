---
title: Tenants API
portal: 8 — API Reference
section: 08-api-reference
audience: [backend-engineer, integrator]
related:
  - "[API Reference Overview](01-api-reference-overview.md)"
  - "[Authentication API](02-authentication.md)"
  - "[Tenancy Model](../03-platform-architecture/01-multi-tenancy/01-tenancy-model.md)"
---

# Tenants API

Tenant management endpoints. Requires `super_admin` role — tenant-scoped users cannot access these endpoints.

## POST /api/v1/admin/tenants

Create a new tenant.

### Request

```http
POST /api/v1/admin/tenants
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Acme Corporation",
  "slug": "acme",
  "domain": "acme.example.com",
  "plan_tier": "enterprise",
  "company_size": "LARGE",
  "country": "US",
  "timezone": "America/New_York",
  "currency": "USD"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | Display name |
| `slug` | string | yes | URL-safe identifier, unique |
| `domain` | string | no | Custom domain for SSO |
| `plan_tier` | string | yes | `starter`, `professional`, `enterprise` |
| `company_size` | string | yes | `SMALL`, `MEDIUM`, `LARGE`, `ENTERPRISE` |
| `country` | string | yes | ISO 3166-1 alpha-2 |
| `timezone` | string | yes | IANA timezone |
| `currency` | string | yes | ISO 4217 |

### Response `201 Created`

```json
{
  "data": {
    "id": "018e1b2c-3d4e-7f8a-9b0c-1d2e3f4a5b6c",
    "name": "Acme Corporation",
    "slug": "acme",
    "status": "PENDING",
    "plan_tier": "enterprise",
    "company_size": "LARGE",
    "country": "US",
    "timezone": "America/New_York",
    "currency": "USD",
    "created_at": "2026-01-15T10:00:00Z"
  }
}
```

### Error Responses

| Status | Code | Condition |
|--------|------|-----------|
| 409 | `CONFLICT` | `slug` already taken |
| 422 | `VALIDATION_ERROR` | Invalid field values |

---

## GET /api/v1/admin/tenants

List all tenants with pagination.

### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `status` | string | — | Filter: `PENDING`, `ACTIVE`, `SUSPENDED`, `ARCHIVED` |
| `plan_tier` | string | — | Filter by plan |
| `limit` | int | 20 | Max 100 |
| `offset` | int | 0 | Pagination offset |

### Response `200 OK`

```json
{
  "data": [
    {
      "id": "018e1b2c-...",
      "name": "Acme Corporation",
      "slug": "acme",
      "status": "ACTIVE",
      "plan_tier": "enterprise",
      "created_at": "2026-01-15T10:00:00Z"
    }
  ],
  "meta": {
    "total": 142,
    "limit": 20,
    "offset": 0
  }
}
```

---

## GET /api/v1/admin/tenants/:id

Get tenant by ID.

### Response `200 OK`

Full tenant object (same shape as POST response).

### Error Responses

| Status | Code | Condition |
|--------|------|-----------|
| 404 | `NOT_FOUND` | Tenant not found |

---

## PUT /api/v1/admin/tenants/:id

Update tenant fields.

### Request

```json
{
  "name": "Acme Corp (Updated)",
  "plan_tier": "enterprise",
  "company_size": "ENTERPRISE"
}
```

All fields optional — only provided fields are updated.

### Response `200 OK`

Updated tenant object.

---

## POST /api/v1/admin/tenants/:id/activate

Transition tenant from `PENDING` to `ACTIVE`.

### Response `200 OK`

```json
{
  "data": {
    "id": "018e1b2c-...",
    "status": "ACTIVE"
  }
}
```

### Error Responses

| Status | Code | Condition |
|--------|------|-----------|
| 409 | `CONFLICT` | Tenant not in `PENDING` status |

---

## POST /api/v1/admin/tenants/:id/suspend

Suspend an active tenant. Blocks all logins for tenant users.

### Request

```json
{
  "reason": "Payment overdue"
}
```

### Response `200 OK`

```json
{
  "data": {
    "id": "018e1b2c-...",
    "status": "SUSPENDED"
  }
}
```

---

## POST /api/v1/admin/tenants/:id/unsuspend

Restore suspended tenant to `ACTIVE`.

---

## POST /api/v1/admin/tenants/:id/archive

Archive a tenant. Irreversible — data is retained but all access is blocked.

### Request

```json
{
  "reason": "Customer churned"
}
```

## Tenant Status Lifecycle

```
PENDING → ACTIVE → SUSPENDED → ACTIVE
                ↓
           ARCHIVED
```

- `PENDING` → `ACTIVE`: via `/activate`
- `ACTIVE` → `SUSPENDED`: via `/suspend`
- `SUSPENDED` → `ACTIVE`: via `/unsuspend`
- `ACTIVE` → `ARCHIVED`: via `/archive`
- Cannot transition from `ARCHIVED`
- Cannot go `PENDING` → `SUSPENDED` directly
