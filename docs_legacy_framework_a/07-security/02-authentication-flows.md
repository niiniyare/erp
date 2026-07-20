> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Authentication Flows
portal: 7 — Security
section: 07-security
audience: [backend-engineer, security, architect]
related:
  - "[Security Overview](01-security-overview.md)"
  - "[Session Architecture](../03-platform-architecture/02-iam/02-session-architecture.md)"
  - "[Auth Middleware](../04-backend-engineering/00-module-development-guide/16-middleware-chain/02-auth-middleware.md)"
---

# Authentication Flows

## Login Flow

```
Client                    API Server               Redis            Postgres
  │                            │                     │                 │
  │── POST /auth/login ────────▶                     │                 │
  │   { email, password }      │                     │                 │
  │                            │── SELECT user ──────────────────────▶│
  │                            │◀─ user + hash ───────────────────────│
  │                            │                     │                 │
  │                            │ bcrypt.Compare()    │                 │
  │                            │                     │                 │
  │                            │── SELECT roles, ────────────────────▶│
  │                            │   entity_scope,     │                 │
  │                            │   feature_flags     │                 │
  │                            │◀─ ResolvedSession ──────────────────│
  │                            │                     │                 │
  │                            │── SETEX session ────▶                │
  │                            │   {uuid} 86400      │                 │
  │                            │   {json}            │                 │
  │                            │◀─ OK ───────────────│                 │
  │                            │                     │                 │
  │◀── 200 { token, user } ────│                     │                 │
```

**Rate limit**: 10 requests/minute/IP on this endpoint.

**Password hashing**: bcrypt with cost factor 12. Comparison is constant-time to prevent timing attacks.

## Authenticated Request Flow

```
Client                    API Server               Redis
  │                            │                     │
  │── GET /api/v1/contracts ───▶                     │
  │   Authorization: Bearer    │                     │
  │   {token}                  │                     │
  │                            │── GET session:{t} ──▶
  │                            │◀─ ResolvedSession ──│
  │                            │                     │
  │                            │ Check ExpiresAt     │
  │                            │                     │
  │                            │── EXPIRE 86400 ─────▶  (slide TTL)
  │                            │                     │
  │                            │ Inject session      │
  │                            │ into Fiber locals   │
  │                            │                     │
  │                            │ [Handler executes]  │
  │◀── 200 { data } ───────────│                     │
```

Redis hit on every request. Pipeline: `GET` + `EXPIRE` in one round trip using `TxPipelined`.

## Logout Flow

```
Client                    API Server               Redis
  │── POST /auth/logout ───────▶                     │
  │   Authorization: Bearer    │── DEL session:{t} ──▶
  │   {token}                  │                     │
  │◀── 204 ────────────────────│                     │
```

Session deleted immediately. Subsequent requests with the same token get 401.

## Force Revoke (Security Incident)

```bash
# All sessions for a user (on role change, compromise)
awoctl sessions revoke-all --user-id {uuid}

# All sessions for a tenant (on suspension)
awoctl sessions revoke-tenant --tenant-id {uuid}
```

Implementation: `SCAN session:* WHERE tenant_id = X` then `DEL` each. O(n) but infrequent.

## Token Format

Tokens are UUIDs (v4), stored and transmitted as lowercase hyphenated strings:

```
550e8400-e29b-41d4-a716-446655440000
```

Not JWTs. Benefits:
- Instant revocation (delete from Redis)
- No signature verification overhead
- No token data exposure (UUIDs have no payload)
- No clock skew issues

Trade-off: requires Redis lookup on every request.

## Password Policy

| Rule | Value |
|------|-------|
| Minimum length | 12 characters |
| Must contain | Uppercase, lowercase, digit, symbol |
| Bcrypt cost | 12 |
| Max length | 128 characters (prevents bcrypt DoS) |
| Breach check | Checked against HaveIBeenPwned on registration |

## Session Security

| Property | Value |
|----------|-------|
| TTL | 24 hours, sliding |
| Token entropy | 128-bit (UUID v4) |
| Transmission | HTTPS only; token in `Authorization` header |
| Storage (client) | `localStorage` or `sessionStorage` — document in frontend guide |
| Redis key space | `session:{uuid}` — no tenant prefix in key |

## Tenant Isolation in Auth

The `ResolvedSession` contains `TenantID`. Every DB query runs inside `WithTenant(tenantID)` which sets `SET LOCAL app.tenant_id = $tenantID`. PostgreSQL RLS enforces the boundary.

A token from tenant A cannot access tenant B's data — not through session logic, and not through RLS even if logic is bypassed.

## Common Attack Mitigations

| Attack | Mitigation |
|--------|-----------|
| Brute force login | Rate limit 10/min/IP; bcrypt cost 12 makes each attempt slow |
| Session fixation | Token generated server-side on login, never client-provided |
| Token theft via XSS | Use `HttpOnly` cookie option if switching from header to cookie |
| Token theft in transit | HTTPS required; `Strict-Transport-Security` header |
| Session after account disable | `InvalidateUserSessions` called on user deactivation |
| Replay attack | Each token is one-time until TTL expires; immediate revoke on logout |
