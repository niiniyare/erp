[<-- Back to Index](README.md)

## Architecture Overview

> **Implementation status**: This document describes the v1.0 runtime architecture as implemented.
> Features labeled **[PLANNED - NOT IN v1.0]** are not active at runtime.

---

### Module Placement

The IAM bounded context lives in `internal/core/iam/`. External callers import only the facade package `iam`; internal sub-packages (`service/`, `domain/`, `repository/`) are implementation details.

```
┌───────────────────────────────────────────────────────────────┐
│                    HTTP / Fiber Layer                          │
│  Routes: Authenticate → SetTenantContext → Enforce → Handler  │
└──────────────────────────┬────────────────────────────────────┘
                           │  domain.Request
┌──────────────────────────▼────────────────────────────────────┐
│              AuthzService (interface)                          │
│  Enforce / EnforceBatch / AssignRole / AddPolicy / ...         │
│  Package: internal/core/iam/service/authz.go                   │
└──────────┬───────────────────────────────────┬────────────────┘
           │ casbin.SyncedEnforcer             │ AuthzRepository
┌──────────▼───────────┐       ┌──────────────▼────────────────┐
│   Casbin Engine       │       │   PostgreSQL                   │
│  (in-memory model)   │◄─────►│   casbin_rule                  │
│  SyncedEnforcer       │       │   role_assignments             │
│  AutoLoad: 30s        │       └───────────────────────────────┘
└──────────────────────┘
```

---

### Actual Module File Structure

```
internal/core/iam/
├── iam.go              — public facade
├── service.go          — empty
├── types.go            — package helpers
├── adapter.go          — Casbin pgx adapter helpers
├── seed.go             — seeding utilities
├── domain/
│   ├── authz.go        — ActorType, Principal, Request, Policy, RoleAssignment, CasbinModel constant
│   ├── session.go      — Session, ResolvedSession, EntityScope, Configuration
│   ├── sso.go          — OAuthProvider, SSOProvider
│   ├── apikey.go       — APIKey
│   └── errors.go       — IAM error type
├── service/
│   ├── authz.go        — AuthzService (SyncedEnforcer, enforce, role/policy management, bootstrap)
│   ├── identity.go     — UserService (CRUD, authenticate, MFA, password reset)
│   ├── session.go      — SessionService (login, MFA flow, SSO login, logout, validate)
│   ├── sso.go          — SSOService (OAuth PKCE-less flow, JIT provisioning)
│   └── apikey.go       — APIKeyService (create, validate, revoke)
└── repository/
    ├── session.go      — SessionRepository (DB + Redis cache-aside, MFA pending state)
    └── ...

db/migration/
├── 000063_authz_casbin_rule.up.sql       — casbin_rule table + indexes + RLS
├── 000064_authz_role_assignments.up.sql  — role metadata + RLS
├── 000304_identity_add_user_sessions.up.sql
├── 000305_identity_sessions_add_permissions.up.sql
├── 000309_iam_sso_providers.up.sql
├── 000310_iam_api_keys.up.sql
└── 000405-000407 — V2.0 RESERVED (roles table, ABAC) — NOT ACTIVE IN v1.0
```

---

### Request Flow [IMPLEMENTED]

Every protected HTTP request:

```
HTTP Request arrives
       │
       ▼
Fiber Router matches route
       │
       ▼
Authenticate middleware
  → reads opaque token from Cookie ("session") or Authorization: Bearer header
  → sha256hex(token)
  → SessionRepository.ValidateToken():
      1. Redis cache-aside (cache key: "session:{hash}")
      2. DB fallback: TouchAndGetSession (atomic UPDATE last_seen + SELECT)
  → builds ResolvedSession (UserID, TenantID, UserType, EntityScope, Configuration)
  → c.Locals("resolved_session") = ResolvedSession
  → c.Locals("authz_principal")  = ResolvedSession.ToPrincipal()
  → ctx.Value(cache.TenantIDKey) = TenantID.String()
       │
       ▼
SetTenantContext middleware
  → executes: SET LOCAL awo.tenant_id = '<TenantID>'  (PostgreSQL RLS boundary)
       │
       ▼
Authorization middleware (per-route)
  → reads Principal from c.Locals("authz_principal")
  → calls authzService.Enforce(ctx, domain.Request{Subject, Domain, Object, Action})
       │
       ├── revokeExpiredRoles(subject, domain)  [lazy expiry cleanup]
       ├── enforcer.Enforce(sub, dom, obj, act) [in-memory: no DB hit]
       │     ├── g-rule lookup (role membership)
       │     ├── p-rule match (keyMatch2 on obj, keyMatch on act)
       │     └── effect: some(allow) && !some(deny)
       └── false → 403 Forbidden / true → c.Next()
       │
       ▼
Route Handler executes
  → reads session: c.Locals("resolved_session").(*domain.ResolvedSession)
  → reads flags:   session.Configuration.Flags["feature.name"]
  → reads settings: session.Configuration.Settings["setting.key"]
  → authorization is already decided — handler does NOT re-check
```

---

### Data Model [IMPLEMENTED]

**Three active tables for the authorization layer:**

```
casbin_rule — Casbin policy store (source of truth for enforcement)
  id    : UUID
  ptype : "p" (policy rule) or "g" (role assignment/inheritance)
  v0-v5 : positional fields
           p-rule: v0=sub, v1=dom, v2=obj, v3=act, v4=eft
           g-rule: v0=user, v1=role, v2=domain

  NOTE: No RLS tenant isolation on casbin_rule.
        Multi-tenant isolation is via domain value in v1 (application layer).

role_assignments — Metadata for audit/expiry (paired with casbin_rule g-rows)
  id          : UUID
  tenant_id   : FK → tenants(id)  [RLS-enforced]
  subject     : "tenant:uuid" | "portal:uuid" | etc.
  role_name   : "tenant_admin" | "finance-manager" | etc.
  domain      : tenantID UUID | "_platform_" | "tenantID:portal"
  assigned_by : subject string of granter
  granted_by  : UUID FK to users(id)
  delegated_by: delegation chain subject
  expires_at  : TIMESTAMPTZ (NULL = permanent)
  is_active   : BOOLEAN
  created_at  : TIMESTAMPTZ

user_sessions — Active session rows
  id            : UUID
  user_id       : FK → users(id)
  tenant_id     : UUID (RLS boundary)
  user_type     : persisted enum string
  session_token : sha256hex(raw_token)  [column name omits "_hash" for brevity]
  principal_id  : *UUID (portal users only)
  entity_scope  : JSONB  {"type":"subtree","entity_id":"...","path_prefix":"..."}
  configuration : JSONB  {"flags":{...},"settings":{...},"prefs":{...}}
  is_active     : BOOLEAN
  expires_at    : TIMESTAMPTZ
  last_seen_at  : TIMESTAMPTZ
  ip_address    : INET
  user_agent    : TEXT
```

---

### Design Decisions

**Why single Casbin enforcement (not session.Can() fast path)?**
- One enforcement authority. Role revocations take effect immediately (next request after RevokeRole updates in-memory model — no session invalidation needed for revocations).
- Session.Can() is not implemented. The `Permissions` map was deliberately removed.
- Casbin Enforce() is in-memory (~0.1ms). The latency is acceptable.

**Why SyncedEnforcer?**
- `casbin.SyncedEnforcer` is goroutine-safe (vs the non-synced variant).
- `StartAutoLoadPolicy(30 * time.Second)` reloads from DB every 30 seconds, making multi-instance deployments eventually consistent without a separate watcher.

**Why DB sessions (not JWT)?**
- JWT tokens cannot be revoked without a separate blocklist.
- A session row is marked inactive immediately on logout or suspension.
- Redis cache-aside makes validation fast on the hot path.

**Why two tables (casbin_rule + role_assignments)?**
- `casbin_rule` is the source of truth for enforcement (what Casbin reads).
- `role_assignments` provides the audit trail (who assigned, when, with what expiry).

---

### Multi-Instance Policy Propagation [IMPLEMENTED]

Each app instance holds the full policy set in memory. Propagation mechanism:

- `SyncedEnforcer.StartAutoLoadPolicy(30s)` — each instance reloads from PostgreSQL every 30 seconds.
- Maximum drift between instances: 30 seconds.
- `InvalidateCache()` triggers an immediate reload on the calling instance (not broadcast).
- **[PLANNED - NOT IN v1.0]**: Redis pub/sub for sub-second propagation.

---

Next: [Domain Model](./04-domain-model.md)
