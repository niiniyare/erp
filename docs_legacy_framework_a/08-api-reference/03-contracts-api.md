> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Contracts API
portal: 8 — API Reference
section: 08-api-reference
audience: [backend-engineer, frontend-engineer, integrator]
related:
  - "[API Reference Overview](01-api-reference-overview.md)"
  - "[Contracts Module](../02-product-overview/02-module-overview.md)"
---

# Contracts API

## GET /api/v1/contracts

List contracts for the authenticated tenant.

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | int | Page number (default: 1) |
| `page_size` | int | Items per page (default: 20, max: 100) |
| `status` | string | Filter by status |
| `entity_id` | uuid | Filter by entity |

### Response 200

```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "contract_number": "CONT-2025-0001",
      "title": "Software Development Services",
      "status": "active",
      "contract_type": "service",
      "total_value": "50000.000000",
      "currency": "USD",
      "start_date": "2025-01-01",
      "end_date": "2025-12-31",
      "vendor_id": "...",
      "version": 3,
      "created_at": "2025-01-01T09:00:00Z",
      "updated_at": "2025-01-15T10:30:00Z",
      "deleted_at": null
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 87,
    "total_pages": 5
  }
}
```

---

## POST /api/v1/contracts

Create a new contract (status: draft).

### Request

```json
{
  "contract_number": "CONT-2025-0042",
  "title": "Consulting Services Agreement",
  "vendor_id": "...",
  "contract_type": "service",
  "start_date": "2025-03-01",
  "end_date": "2026-02-28",
  "total_value": "120000",
  "currency": "USD",
  "description": "Annual consulting retainer"
}
```

### Response 201

Returns created contract. `Location: /api/v1/contracts/{id}` header included.

### Response 409

Duplicate `contract_number` for the tenant.

### Response 422

Validation error (missing required fields, invalid format).

---

## GET /api/v1/contracts/:id

Get a single contract by ID.

### Response 404

Contract does not exist or belongs to another tenant (RLS).

---

## PUT /api/v1/contracts/:id

Update a contract. Only allowed when `status = "draft"`.

### Request

Includes `version` for optimistic locking:

```json
{
  "title": "Updated Title",
  "total_value": "130000",
  "version": 2
}
```

### Response 409

`version` mismatch — another request modified the contract. Client must re-fetch and retry.

### Response 422

Contract is not in draft status (`ErrContractNotEditable`).

---

## DELETE /api/v1/contracts/:id

Soft-delete a contract.

### Request Body

```json
{ "version": 3 }
```

### Response 204

Deleted successfully.

---

## POST /api/v1/contracts/:id/submit

Submit contract for review.

```json
{ "version": 1 }
```

**Response 422** if not in `draft` status.

---

## POST /api/v1/contracts/:id/approve

Approve contract (moves to `approved` status).

```json
{ "version": 2, "comment": "Reviewed and approved" }
```

**Permission required**: `contracts.contract.approve`

---

## POST /api/v1/contracts/:id/activate

Activate an approved contract.

```json
{ "version": 3 }
```

---

## POST /api/v1/contracts/:id/terminate

Terminate an active or suspended contract.

```json
{ "version": 4, "reason": "Contract fulfilled" }
```

---

## POST /api/v1/contracts/import

Upload contracts via CSV. `multipart/form-data`.

**Field**: `file` — CSV file (max 10 MB)

### Response 200

```json
{
  "total_rows": 50,
  "succeeded": 48,
  "failed": 2,
  "errors": [
    { "row_number": 12, "field": "contract_number", "message": "already exists" },
    { "row_number": 31, "field": "start_date", "message": "must be YYYY-MM-DD" }
  ],
  "contract_ids": ["...", "..."]
}
```

---

## GET /api/v1/contracts/export

Download all contracts as CSV.

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `status` | string | Filter by status |
| `format` | string | `csv` (only supported format) |

### Response 200

`Content-Type: text/csv`
`Content-Disposition: attachment; filename=contracts-20250115.csv`
