---
title: API Reference Overview
portal: 8 — API Reference
section: 08-api-reference
audience: [backend-engineer, frontend-engineer, integrator]
related:
  - "[API Design](../04-backend-engineering/00-module-development-guide/14-api-design/01-api-design-overview.md)"
  - "[URL Conventions](../04-backend-engineering/00-module-development-guide/14-api-design/02-url-conventions.md)"
  - "[Authentication](02-authentication.md)"
---

# API Reference Overview

## Base URL

```
https://{tenant}.awoerp.com/api/v1
```

Or for on-premise:
```
https://{hostname}/api/v1
```

## Authentication

All API requests require a session token in the `Authorization` header:

```
Authorization: Bearer {token}
```

Obtain a token via `POST /api/v1/auth/login`.

## Request Format

- Content-Type: `application/json` for all write requests
- Dates: `YYYY-MM-DD` in query parameters and request bodies
- Timestamps: RFC3339 (`2025-01-15T10:30:00Z`) in responses
- Monetary values: decimal strings (`"50000.000000"`) — never floats
- UUIDs: lowercase hyphenated string format

## Response Format

### Success

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "contract_number": "CONT-2025-0001",
  ...
}
```

List responses include pagination:
```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 87,
    "total_pages": 5
  }
}
```

### Error

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      { "field": "contract_number", "message": "contract_number is required" }
    ]
  }
}
```

## HTTP Status Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 201 | Created (POST that creates a resource) |
| 204 | No Content (DELETE) |
| 400 | Bad Request (malformed JSON, invalid UUID) |
| 401 | Unauthorized (missing or expired token) |
| 403 | Forbidden (insufficient permissions) |
| 404 | Not Found |
| 409 | Conflict (optimistic lock, duplicate key) |
| 422 | Unprocessable Entity (validation failed) |
| 429 | Too Many Requests |
| 500 | Internal Server Error |

## Pagination

All list endpoints accept:

| Parameter | Default | Max |
|-----------|---------|-----|
| `page` | 1 | — |
| `page_size` | 20 | 100 |

## Modules

- [Authentication API](02-authentication.md)
- [Contracts API](03-contracts-api.md)
- [Tenants API](04-tenants-api.md)
- [Entities API](05-entities-api.md)
