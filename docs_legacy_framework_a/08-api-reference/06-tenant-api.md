> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Tenant API
portal: 8 — API Reference
section: 08-api-reference
audience: [backend-engineer, integrator]
related:
  - "[API Reference Overview](01-api-reference-overview.md)"
  - "[Tenancy Model](../03-platform-architecture/01-multi-tenancy/01-tenancy-model.md)"
---

# Tenant API

These endpoints are platform-level (not tenant-scoped). Used by platform admins and provisioning systems. Most require elevated permissions not available to regular tenant users.

## Tenants

### GET /api/v1/tenants

List all tenants.

**Permission required**: `platform.tenant.list` (platform admin only)

#### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `status` | string | `pending`, `active`, `suspended`, `archived` |
| `page` | int | — |
| `page_size` | int | — |

#### Response 200

```json
{
  "data": [
    {
      "id": "aaaaaaaa-0000-0000-0000-000000000001",
      "name": "Acme Corp",
      "subdomain": "acme",
      "status": "active",
      "plan": "enterprise",
      "company_size": "LARGE",
      "country": "US",
      "created_at": "2025-01-01T00:00:00Z",
      "activated_at": "2025-01-02T00:00:00Z"
    }
  ],
  "pagination": { "page": 1, "page_size": 20, "total": 5, "total_pages": 1 }
}
```

---

### POST /api/v1/tenants

Create and provision a new tenant.

**Permission required**: `platform.tenant.create`

```json
{
  "name": "NewCo Ltd",
  "subdomain": "newco",
  "admin_email": "admin@newco.com",
  "admin_name": "Admin User",
  "plan": "starter",
  "company_size": "small",
  "country": "GB"
}
```

`company_size` values: `small`, `medium`, `large`, `enterprise` (stored uppercase internally).

**Response 201**

```json
{
  "id": "...",
  "name": "NewCo Ltd",
  "subdomain": "newco",
  "status": "pending",
  "admin_user_id": "...",
  "created_at": "2025-01-15T10:00:00Z"
}
```

**Response 409** — subdomain already taken.

---

### GET /api/v1/tenants/:id

Get tenant details.

---

### PUT /api/v1/tenants/:id

Update tenant. Include `version`.

```json
{
  "name": "NewCo International Ltd",
  "plan": "professional",
  "version": 1
}
```

---

## Tenant Lifecycle

### POST /api/v1/tenants/:id/activate

Activate a `pending` tenant. Allowed transitions: `PENDING → ACTIVE`.

```json
{ "version": 1 }
```

**Response 422** — tenant not in `pending` status.

---

### POST /api/v1/tenants/:id/suspend

Suspend an `active` tenant. Prevents login for all tenant users.

```json
{ "version": 2, "reason": "Payment overdue" }
```

Allowed transitions: `ACTIVE → SUSPENDED`.

---

### POST /api/v1/tenants/:id/reactivate

Reactivate a `suspended` tenant.

```json
{ "version": 3 }
```

Allowed transitions: `SUSPENDED → ACTIVE`.

---

### POST /api/v1/tenants/:id/archive

Archive tenant. Soft-delete — data retained but tenant inaccessible.

```json
{ "version": 4, "reason": "Contract ended" }
```

Allowed transitions: `PENDING → ARCHIVED` or `ACTIVE → ARCHIVED` (not from SUSPENDED directly).

---

## Tenant Configuration

### GET /api/v1/tenants/:id/config

Get all configuration values for a tenant.

#### Response 200

```json
{
  "configurations": [
    {
      "key": "finance.default_currency",
      "value": "USD",
      "description": "Default currency for financial transactions"
    },
    {
      "key": "contracts.approval_required",
      "value": "true",
      "description": "Require approval before contract activation"
    }
  ]
}
```

---

### PUT /api/v1/tenants/:id/config/:key

Set a configuration value.

```json
{ "value": "EUR" }
```

**Response 422** — unknown config key, or value fails validation for that key.

---

## Feature Flags

### GET /api/v1/tenants/:id/features

Get feature flags for tenant.

#### Response 200

```json
{
  "features": [
    { "name": "finance.advanced_reporting", "enabled": true },
    { "name": "contracts.bulk_import", "enabled": false }
  ]
}
```

---

### PUT /api/v1/tenants/:id/features/:name

Enable or disable a feature flag.

```json
{ "enabled": true }
```

**Response 422** — unknown feature flag name.

---

## Usage Stats

### GET /api/v1/tenants/:id/usage

Get tenant usage statistics.

#### Response 200

```json
{
  "tenant_id": "...",
  "period": "2025-01",
  "stats": {
    "active_users": 12,
    "contracts_created": 45,
    "transactions_posted": 230,
    "storage_mb": 128
  }
}
```
