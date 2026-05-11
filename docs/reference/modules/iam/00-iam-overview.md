[<-- Back to Index](README.md)

## IAM Module — Philosophy & Scope

> **Implementation status**: This document describes the v1.0 production architecture as implemented.
> Features not yet implemented are explicitly labeled **[PLANNED - NOT IN v1.0]**.

---

### The Three Questions Every Request Answers

Every HTTP request that enters Awo ERP answers three questions before touching business data:

1. **Authentication (AuthN): Who are you?** Verify the claimed identity. Produce a trusted session.
2. **Authorization (AuthZ): What can you do?** Given your identity, ask Casbin. Casbin is the sole authorization authority.
3. **Configuration: How should this behave?** Given your tenant's flags and settings, which features are on?

These are distinct concerns with distinct mechanisms:

- **AuthN** resolves at login time and is encoded in the session row and Redis cache.
- **AuthZ** resolves at request time via the Casbin in-memory enforcer.
- **Configuration** (flags, settings, preferences) is pre-computed at login and embedded in the session.

---

### The Authorization Model [IMPLEMENTED]

**RBAC-only in v1.0. Casbin is the single enforcement authority.**

- All permission decisions go through `authzService.Enforce(ctx, domain.Request{...})`.
- The session object carries identity and context only: `UserID`, `TenantID`, `UserType`, `EntityScope`, `Configuration`. It has no `Permissions` map, no `Can()` method, no `CanDo()` method.
- Feature flags and tenant settings are read from `session.Configuration` (O(1) map lookup). These are context reads, not authorization decisions.
- Role revocations take effect on the next request without requiring session invalidation, because enforcement is always live against the in-memory Casbin model.

---

### What the Session Carries [IMPLEMENTED]

The session is **context only**:

```
ResolvedSession:
  UserID        — who the user is
  TenantID      — which tenant (RLS key)
  UserType      — actor type string (mapped to ActorType for Casbin domain construction)
  PrincipalID   — for portal users only: the represented party record
  DisplayName   — for display purposes
  EntityScope   — which entities within the tenant this user can see (application-layer)
  Configuration — pre-computed flags, tenant settings, user preferences
```

The session does **not** contain:
- A permissions map
- A risk score
- Any pre-computed authorization decision

---

### The Trust Chain [IMPLEMENTED]

```
Browser / API Client
      │
      ▼
Fiber HTTP Server
      ├── Recovery, Logger, RateLimit, CORS, SecurityHeaders
      │
      ├── Authenticate()
      │     → reads token from Cookie or Authorization: Bearer header
      │     → sha256(token) → Redis cache-aside → DB fallback
      │     → loads ResolvedSession into c.Locals("resolved_session")
      │     → builds Principal → c.Locals("authz_principal")
      │     → sets ctx tenant_id via cache.TenantIDKey
      │
      ├── SetTenantContext()
      │     → SET LOCAL awo.tenant_id = '<TenantID>'  (PostgreSQL RLS)
      │
      ├── RequirePermission() / authzService.Enforce()
      │     → Casbin in-memory check — no DB hit on hot path
      │     → revokeExpiredRoles() — lazy cleanup, indexed query
      │     → false → 403 Forbidden
      │
      └── Handler
            → reads session from c.Locals("resolved_session")
            → reads feature flags from session.Configuration.Flags
            → reads tenant settings from session.Configuration.Settings
```

---

### What This Module Owns [IMPLEMENTED]

- User lifecycle (create, update, delete, authenticate)
- Authentication: login, logout, TOTP MFA, OAuth/OIDC (Google, Microsoft), passwords, API keys
- Session management: DB-persisted, Redis cache-aside, token hashing, MFA pending tokens
- RBAC via Casbin: roles, policies, temporal assignments, deny-override, domain isolation
- Entity hierarchy resolution (EntityScope) — stored in session as context
- Feature flag and tenant settings snapshots — stored in session Configuration
- User preferences — stored in session Configuration

**Does not own:** HTTP routing, DB connection pools, notification delivery, audit log persistence, tenant provisioning business logic, business module schemas.

---

### Module Structure [IMPLEMENTED]

```
internal/core/iam/
├── iam.go              — public facade: re-exports all types and constructors
├── service.go          — empty stub
├── types.go            — package-level helpers
├── adapter.go          — Casbin pgx adapter helpers
├── seed.go             — role/policy seeding utilities
├── domain/
│   ├── authz.go        — ActorType, Principal, Request, Policy, RoleAssignment, CasbinModel
│   ├── session.go      — Session, ResolvedSession, EntityScope, Configuration
│   ├── sso.go          — OAuthProvider, SSOProvider, SSOUserInfo
│   ├── apikey.go       — APIKey, CreateAPIKeyRequest
│   └── errors.go       — domain error types
├── service/
│   ├── authz.go        — AuthzService (Casbin enforcement, role/policy management)
│   ├── identity.go     — UserService (auth, MFA, password, CRUD)
│   ├── session.go      — SessionService (login, logout, MFA flow, SSO login)
│   ├── sso.go          — SSOService (OAuth flow, JIT provisioning)
│   └── apikey.go       — APIKeyService (create, validate, revoke)
└── repository/
    ├── session.go      — SessionRepository (DB + Redis cache-aside)
    └── ...
```

---

### Operational Guides

For day-to-day operation and administration of the IAM module, see:

- [Tenant Administration](./23-tenant-administration.md) — bootstrap, user lifecycle, custom roles, separation of duties
- [Platform Administration](./24-platform-administration.md) — platform users, super admin bootstrap, platform authority model
- [User Entity Scope](./25-user-entity-scope.md) — ALL / SUBTREE / ENTITY_ONLY scope types with worked examples
- [API Keys and Service Accounts](./26-api-keys-and-service-accounts.md) — key lifecycle, service account model, security and rotation guidance
- [Resource/Action Ownership](./27-resource-action-ownership.md) — module-owned resources, naming conventions, feature flags vs settings vs preferences

---

### What Is Not Implemented

**[PLANNED - NOT IN v1.0]** — see `deferred-features.md`:

- ABAC (attribute-based access control) — `internal/core/access/` is gated with `//go:build ignore`
- Access requests and approval workflows
- Conditional access (time, location, device conditions)
- SAML 2.0 (only OAuth/OIDC is implemented)
- SMS/email OTP MFA (only TOTP is implemented)
- Redis pub/sub for real-time cross-instance policy propagation (auto-reload every 30s is active)

---

Next: [Code Architecture](./00b-code-architecture.md)
