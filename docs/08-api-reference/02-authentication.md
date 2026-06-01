---
title: Authentication API
portal: 8 — API Reference
section: 08-api-reference
audience: [backend-engineer, frontend-engineer, integrator]
related:
  - "[API Reference Overview](01-api-reference-overview.md)"
  - "[Session Architecture](../03-platform-architecture/02-iam/02-session-architecture.md)"
---

# Authentication API

## POST /api/v1/auth/login

Authenticate with email and password. Returns a session token.

**Rate limit**: 10 requests per minute per IP.

### Request

```json
{
  "email": "user@example.com",
  "password": "secret"
}
```

### Response 200

```json
{
  "token": "550e8400-e29b-41d4-a716-446655440000",
  "expires_at": "2025-01-15T18:30:00Z",
  "user": {
    "id": "bbbbbbbb-0000-0000-0000-000000000001",
    "email": "user@example.com",
    "name": "Jane Smith"
  },
  "tenant": {
    "id": "aaaaaaaa-0000-0000-0000-000000000001",
    "name": "Acme Corp"
  }
}
```

### Response 401

```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "invalid email or password"
  }
}
```

### Response 403

```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "tenant account is not active"
  }
}
```

---

## POST /api/v1/auth/logout

Revoke the current session token.

### Request

No body. Token is read from `Authorization` header.

### Response 204

No content.

---

## POST /api/v1/auth/refresh

Extend the session TTL. Equivalent to any authenticated request — the TTL slides automatically on each request. Provided as an explicit endpoint for clients that want to heartbeat without making business API calls.

### Response 200

```json
{
  "expires_at": "2025-01-15T18:30:00Z"
}
```

---

## GET /api/v1/auth/me

Returns the current user and session details.

### Response 200

```json
{
  "user_id": "bbbbbbbb-0000-0000-0000-000000000001",
  "tenant_id": "aaaaaaaa-0000-0000-0000-000000000001",
  "email": "user@example.com",
  "roles": ["contracts.editor", "finance.viewer"],
  "entity_scope": {
    "type": "entity",
    "entity_id": "cccccccc-0000-0000-0000-000000000001"
  },
  "expires_at": "2025-01-15T18:30:00Z"
}
```
