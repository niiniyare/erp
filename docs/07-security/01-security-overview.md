---
title: Security Overview
portal: 7 — Security
section: 07-security
audience: [architect, backend-engineer, devops, tech-lead]
related:
  - "[Authorization Model](../03-platform-architecture/02-iam/03-authorization-model.md)"
  - "[RLS Enforcement](../03-platform-architecture/01-multi-tenancy/02-rls-enforcement.md)"
  - "[Audit Logging](../04-backend-engineering/00-module-development-guide/15-audit-logging/01-audit-overview.md)"
---

# Security Overview

## Defence-in-Depth Layers

```
Layer 1: Network        — TLS everywhere, VPC isolation, no direct DB access from internet
Layer 2: Authentication — Session token validation on every request
Layer 3: Authorization  — Casbin RBAC, route-level + service-level checks
Layer 4: Data isolation — PostgreSQL RLS, tenant_id on every table
Layer 5: Input validation — Request body validation, parameterized SQL only
Layer 6: Audit          — Immutable audit trail for every write
```

No single layer is the only protection. A bug in one layer should not compromise the entire system.

## Authentication Security

- Session tokens are opaque UUIDs stored in Redis — not JWTs
- No client-side decoding possible
- Tokens expire after 8 hours of inactivity
- Revocation is immediate (DELETE from Redis)
- Brute-force protection: 10 req/minute on `/auth/*` endpoints

## Authorization Security

- **Fail closed**: authz errors deny access (never default permit)
- **Two levels**: route-level (coarse) + service-level (per-instance)
- **Explicit deny**: Casbin supports `deny` effect to override any `allow` rule
- **No role check in business code**: `session.HasRole()` does not exist — use authz service

## Data Security

- **RLS enforced at DB layer**: even if application code has a bug, PostgreSQL prevents cross-tenant reads
- **`missing_ok` prohibited**: `current_setting('app.tenant_id')` must never use the optional flag
- **Parameterized queries only**: SQLC generates parameterized queries — SQL injection is structurally prevented
- **No secrets in code**: all credentials via environment variables

## Input Validation

- All request bodies validated with `go-playground/validator`
- UUID parameters parsed with `uuid.Parse` — rejects non-UUID values with 400
- File uploads: size limit (10 MB), content-type validation
- Monetary values validated as decimals — never parsed as float64

## Sensitive Data Handling

| Data type | Storage | Transmission |
|-----------|---------|-------------|
| Passwords | bcrypt hash | Never logged |
| Session tokens | Redis (TTL) | HTTPS only |
| API keys | HMAC-SHA256 hash | HTTPS only |
| PII (names, emails) | PostgreSQL encrypted at rest | TLS in transit |
| Audit logs | Append-only PostgreSQL | Internal only |

## Audit Trail

Every write operation (create, update, delete, status transition) generates an audit event with:
- Who (user ID + tenant ID)
- What (resource type + resource ID)
- When (timestamp)
- Before/after state (JSON snapshots)

Audit records are immutable — the `audit_log` table has no UPDATE or DELETE policies.

## Common Vulnerability Prevention

| Vulnerability | Prevention |
|--------------|-----------|
| SQL injection | SQLC parameterized queries |
| XSS | AMIS output encoding + CSP headers |
| CSRF | Session token in header (not cookie) |
| Mass assignment | Explicit DTO mapping — never bind req body directly to domain struct |
| Broken access control | Two-level authorization + RLS |
| Excessive data exposure | Response DTOs — never return raw DB rows |
| Rate limiting | Per-IP + per-tenant limits |
| Insecure direct object reference | Tenant-scoped queries (RLS) |
