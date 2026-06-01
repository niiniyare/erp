---
title: API Design Overview
portal: 4 — Backend Engineering
section: 00-module-development-guide/14-api-design
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-url-conventions.md
    title: URL Conventions
  - path: ./03-request-response.md
    title: Request / Response Design
  - path: ../15-fiber-handlers/01-handler-overview.md
    title: Handler Overview
---

# API Design Overview

AwoERP APIs are REST-over-HTTP, JSON-encoded, tenant-scoped, and version-prefixed. This section defines the conventions that every module must follow.

## Base URL

```
/api/v1/{resource}
```

No tenant prefix in the URL — the tenant is always derived from the authenticated session. This prevents URL enumeration attacks and simplifies client code.

## Design Principles

1. **Resources, not actions** — URLs name resources; HTTP verbs express actions
2. **Stable contracts** — once published, never break request/response shape without versioning
3. **Fail fast with detail** — validation errors return field-level 422 messages
4. **Consistent error envelope** — every error response uses the same JSON structure
5. **Pagination always** — list endpoints always paginate; never return unbounded sets

## HTTP Status Code Usage

| Status | When |
|--------|------|
| 200 OK | Successful read or update |
| 201 Created | Successful create (with Location header) |
| 204 No Content | Successful delete |
| 400 Bad Request | Malformed JSON or invalid URL parameter |
| 401 Unauthorized | Missing or expired session |
| 403 Forbidden | Valid session but insufficient permission |
| 404 Not Found | Resource doesn't exist or RLS hides it |
| 409 Conflict | Optimistic lock version mismatch |
| 422 Unprocessable Entity | Valid JSON but failed domain validation |
| 429 Too Many Requests | Rate limit exceeded |
| 500 Internal Server Error | Unexpected failure (never expose internals) |

## Error Envelope

All error responses use a single JSON structure:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      { "field": "title", "message": "title is required" },
      { "field": "start_date", "message": "must be before end_date" }
    ]
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `code` | string | Machine-readable error code (uppercase snake) |
| `message` | string | Human-readable summary |
| `details` | array | Per-field validation errors (optional) |

## Request ID

Every response includes a `X-Request-ID` header for tracing:

```
X-Request-ID: 550e8400-e29b-41d4-a716-446655440000
```

Set by the request ID middleware before the handler runs.
