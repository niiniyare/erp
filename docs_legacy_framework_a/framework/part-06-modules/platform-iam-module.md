> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: "Platform Module — IAM (Identity & Access Management)"
part: "Part VI — Platform Entities"
chapter: 43
section: "platform-iam-module"
related:
  - "[Chapter 41: Platform Modules Overview](./platform-module-overview.md)"
  - "[Chapter 42: Tenant Module](./platform-tenant-module.md)"
  - "[Chapter 44: Feature Flags Module](./platform-feature-flags-module.md)"
---

# Chapter 43 — IAM: Identity & Access Management

> **Who should read this?** Anyone implementing login flows, writing access-control policies, debugging "403 Forbidden" errors, or integrating Casbin roles into a business module.

---

## 43.1 Three Questions Every Request Must Answer

Every HTTP request arriving at Awo must answer three questions before any business logic runs:

1. **Who are you?** (Authentication — "AuthN") — Verify the session token and extract a user identity.
2. **Are you allowed?** (Authorisation — "AuthZ") — Check whether this identity may perform the requested operation on this entity.
3. **Which data can you see?** (Data scoping) — Restrict results to the tenant and org unit the user belongs to.

The IAM module answers question 1. The `definition.PolicyFunc` chain (per EntityDefinition) answers question 2. The `OrgScope` + `org.Tree` mechanism answers question 3.

This separation is deliberate. Authentication happens once per request in middleware. Authorisation is evaluated per operation by the framework. Scoping is enforced at the database level by RLS + the `OrgUnitIDs` filter.

---

## 43.2 The Four IAM Entities

| Entity | Table | Scope | Purpose |
|---|---|---|---|
| **User** | `users` | Tenant | A human (or service account) that can log in |
| **Role** | `roles` | Tenant | A named permission group |
| **UserRole** | `user_roles` | Tenant | Join: which user holds which role |
| **Session** | `sessions` | Tenant | An active login session |

---

## 43.3 EntityDefinition: User

```go
// internal/platform/iam/definition.go
package iam

import "awo.so/framework/definition"

var UserDef = definition.EntityDefinition{
    Name:        "user",
    Label:       "User",
    Description: "A human or service account that authenticates to Awo.",
    Module:      "Platform",
    Table:       "users",
    OrgScope:    org.ScopeLevelTenant,

    Fields: []*definition.FieldDef{
        definition.Field("email").
            OfType(definition.FieldTypeData).
            WithLabel("Email").
            RequiredField().
            UniqueField().
            SearchableField().
            ImmutableField(), // email is the primary identifier; change = new account

        definition.Field("name").
            OfType(definition.FieldTypeData).
            WithLabel("Full Name").
            RequiredField().
            SearchableField(),

        definition.Field("password_hash").
            OfType(definition.FieldTypeData).
            WithLabel("Password").
            SensitiveField(), // NEVER returned in API responses or logs

        definition.Field("org_unit_id").
            OfType(definition.FieldTypeLink).
            WithLabel("Primary Org Unit").
            LinksTo("org_node"),

        definition.Field("active").
            OfType(definition.FieldTypeBool).
            WithLabel("Active").
            WithDefault(true),

        definition.Field("last_login_at").
            OfType(definition.FieldTypeDateTime).
            WithLabel("Last Login").
            ReadOnlyField(),
    },

    Hooks: []definition.HookDef{
        {Name: "hashPassword", Before: hashPasswordHook},
    },

    Policies: []definition.PolicyDef{
        {Op: definition.OpCreate, Fn: requireRole("tenant_admin")},
        {Op: definition.OpRead,   Fn: allowSelf},          // user can read their own record
        {Op: definition.OpRead,   Fn: requireRole("hr_manager")},
        {Op: definition.OpUpdate, Fn: allowSelfUpdate},    // user can update own name/password
        {Op: definition.OpUpdate, Fn: requireRole("tenant_admin")},
        {Op: definition.OpDelete, Fn: requireRole("tenant_admin")},
    },

    SoftDelete: false, // deactivate via active=false, never delete
    Audited:    true,
}
```

### Why `password_hash` is `SensitiveField`

`Sensitive: true` tells the framework to exclude this field from ALL API responses — `FindByID`, `List`, everything. The hash never leaves the server. Even platform admins cannot retrieve it through the API. If an admin needs to reset a password, they call the `POST /iam/password-reset` custom endpoint which generates a token without exposing the hash.

### Why `email` is `ImmutableField`

Email is the primary login identifier, appears in audit logs, and is used in email communications. Changing it requires identity re-verification (confirm the new email address responds) — a multi-step flow that cannot be done safely through a plain `PUT /users/:id`. The framework enforces this at the SQL level: `mutableFieldNames` excludes `ImmutableField` from UPDATE column lists.

---

## 43.4 EntityDefinition: Role

```go
var RoleDef = definition.EntityDefinition{
    Name:        "role",
    Label:       "Role",
    Description: "A named permission group. Users are assigned roles.",
    Module:      "Platform",
    Table:       "roles",
    OrgScope:    org.ScopeLevelTenant,

    Fields: []*definition.FieldDef{
        definition.Field("name").
            OfType(definition.FieldTypeData).
            WithLabel("Role Name").
            RequiredField().
            UniqueField().
            SearchableField(),

        definition.Field("description").
            OfType(definition.FieldTypeSmallText).
            WithLabel("Description"),

        definition.Field("is_system").
            OfType(definition.FieldTypeBool).
            WithLabel("System Role").
            WithDefault(false).
            ImmutableField(), // set at creation by framework; admins cannot flip this
    },

    Policies: []definition.PolicyDef{
        {Op: definition.OpCreate, Fn: requireRole("tenant_admin")},
        {Op: definition.OpRead,   Fn: allowWithinTenant},
        {Op: definition.OpUpdate, Fn: requireRoleAndNotSystem("tenant_admin")},
        {Op: definition.OpDelete, Fn: requireRoleAndNotSystem("tenant_admin")},
    },

    Audited: true,
}
```

**System roles** (e.g. `tenant_admin`, `platform_admin`) are seeded at tenant provisioning. The `is_system = true` flag prevents them from being renamed or deleted — `requireRoleAndNotSystem` denies the operation if `rec.Get("is_system") == true`.

---

## 43.5 EntityDefinition: UserRole

```go
var UserRoleDef = definition.EntityDefinition{
    Name:        "user_role",
    Label:       "User Role",
    Description: "Assignment of a role to a user.",
    Module:      "Platform",
    Table:       "user_roles",
    OrgScope:    org.ScopeLevelTenant,

    Fields: []*definition.FieldDef{
        definition.Field("user_id").
            OfType(definition.FieldTypeLink).
            WithLabel("User").
            LinksTo("user").
            RequiredField().
            ImmutableField(),

        definition.Field("role_id").
            OfType(definition.FieldTypeLink).
            WithLabel("Role").
            LinksTo("role").
            RequiredField().
            ImmutableField(),
    },

    Hooks: []definition.HookDef{
        {Name: "reloadCasbinPolicy", After: reloadCasbinPolicy},
    },

    Policies: []definition.PolicyDef{
        {Op: definition.OpCreate, Fn: requireRole("tenant_admin")},
        {Op: definition.OpRead,   Fn: allowWithinTenant},
        {Op: definition.OpDelete, Fn: requireRole("tenant_admin")},
    },

    Audited: true,
}
```

### The `reloadCasbinPolicy` Hook

When a `UserRole` is created or deleted, the Casbin enforcer's in-memory policy cache becomes stale. The `reloadCasbinPolicy` AfterHook triggers a synchronous reload from the database:

```go
func reloadCasbinPolicy(ctx context.Context, mut *definition.Mutation) error {
    if mut.Op != definition.OpCreate && mut.Op != definition.OpDelete {
        return nil
    }
    return casbinEnforcer.LoadPolicy()
}
```

This runs **inside the same transaction** as the role assignment (because AfterHooks run inside `WithTx`). If the Casbin reload fails, the entire transaction rolls back — the user_role row is never committed and the policy stays consistent.

---

## 43.6 EntityDefinition: Session

```go
var SessionDef = definition.EntityDefinition{
    Name:        "session",
    Label:       "Session",
    Description: "An active authenticated session.",
    Module:      "Platform",
    Table:       "sessions",
    OrgScope:    org.ScopeLevelTenant,

    Fields: []*definition.FieldDef{
        definition.Field("user_id").
            OfType(definition.FieldTypeLink).
            WithLabel("User").
            LinksTo("user").
            RequiredField().
            ImmutableField(),

        definition.Field("token_hash").
            OfType(definition.FieldTypeData).
            WithLabel("Token Hash").
            ImmutableField().
            SensitiveField(), // raw token is never stored; hash only

        definition.Field("expires_at").
            OfType(definition.FieldTypeDateTime).
            WithLabel("Expires At").
            RequiredField().
            ImmutableField(),

        definition.Field("ip_address").
            OfType(definition.FieldTypeData).
            WithLabel("IP Address").
            ImmutableField().
            SensitiveField(), // PII — excluded from API responses

        definition.Field("user_agent").
            OfType(definition.FieldTypeData).
            WithLabel("User Agent").
            ImmutableField().
            SensitiveField(),

        definition.Field("org_unit_id_snapshot").
            OfType(definition.FieldTypeUUID).
            WithLabel("Org Unit at Login").
            ImmutableField(), // captures org unit at login time; does not change mid-session
    },

    Policies: []definition.PolicyDef{
        {Op: definition.OpCreate, Fn: definition.DenyAll}, // only created by Login service method
        {Op: definition.OpRead,   Fn: allowSelfSession},
        {Op: definition.OpDelete, Fn: allowSelfSession},   // "logout" = delete session
    },

    Audited: false, // session churn would pollute the audit log
}
```

### Why Sessions Are Not Created via the CRUD API

Sessions have `OpCreate: DenyAll`. They are created exclusively by `service.Login()` which:
1. Validates the password
2. Generates a cryptographically random token
3. Stores only the hash
4. Returns the raw token to the caller (once, never stored)

This keeps the token-generation logic in one place and prevents any policy bypass.

---

## 43.7 The Login Flow

```go
// internal/platform/iam/service.go
type Service struct {
    store persistence.TenantStore
}

func (s *Service) Login(ctx context.Context, tenantID uuid.UUID, email, password string) (token string, err error) {
    var session definition.MutableRecord

    if err := s.store.WithTx(ctx, tenantID, func(tx persistence.TenantTx) error {
        // 1. Find the user by email.
        userStore := tx.ForEntity("user")
        users, err := userStore.List(ctx, persistence.ListOptions{
            Filter: map[string]any{"email": email},
            Limit:  1,
        })
        if err != nil || len(users.Records) == 0 {
            return ErrInvalidCredentials
        }
        user := users.Records[0]

        if !user.Get("active").(bool) {
            return ErrAccountDisabled
        }

        // 2. Verify password against stored hash.
        hash := user.Get("password_hash").(string)
        if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
            return ErrInvalidCredentials
        }

        // 3. Generate token, store hash.
        rawToken := generateSecureToken(32)
        tokenHash := sha256Hex(rawToken)
        token = rawToken // captured by closure; returned to caller

        sessionStore := tx.ForEntity("session")
        // Direct insert bypassing the CRUD policy (service-layer privilege).
        session = newSession(user, tenantID, tokenHash)
        return sessionStore.Create(ctx, session)
    }); err != nil {
        return "", err
    }
    return token, nil
}
```

**Never log `rawToken`.** The raw token is the credential. Only `tokenHash` is persisted and logged. If the token appears in logs, an attacker can impersonate the user.

---

## 43.8 Two-Layer Access Control

Awo uses two complementary access-control mechanisms. Understanding when each applies prevents confusion:

### Layer 1: EntityDefinition Policies

Every entity has a `Policies` slice of `PolicyFunc` values. These run synchronously in the HTTP handler for every CRUD request. They answer: **"Can this viewer perform this operation on this record?"**

Policies are evaluated in order. The first policy that returns `nil` (allow) wins. If all policies return a non-nil error, the request is denied. This is **fail-closed**.

```go
// Deny everything — use as last policy to make the deny explicit.
var DenyAll PolicyFunc = func(ctx context.Context, def *EntityDefinition, v ViewerContext, op Op, rec Record) error {
    return ErrDeny
}
```

### Layer 2: Casbin RBAC

Casbin is a rule-engine that evaluates role-based permissions from a policy file (or database table). It answers: **"Does this user's role grant access to this resource/action?"**

In Awo, Casbin is used inside `PolicyFunc` implementations:

```go
func requireRole(role string) definition.PolicyFunc {
    return func(ctx context.Context, def *definition.EntityDefinition, v definition.ViewerContext, op definition.Op, rec definition.Record) error {
        ok, err := casbinEnforcer.Enforce(v.ActorID(), def.Name, string(op))
        if err != nil || !ok {
            return definition.ErrDeny
        }
        return nil
    }
}
```

**Layer 1 is fast and code-defined.** It is best for structural rules: "only the record owner can update this", "this entity is read-only", "create requires a specific role."

**Layer 2 is data-driven.** It is best for configurable permissions: tenant admins can grant roles to users at runtime without a code deployment.

---

## 43.9 The `ViewerContext` Interface

Every request carries a `ViewerContext` extracted by the middleware. The framework defines the interface:

```go
type ViewerContext interface {
    ActorID()   string    // user ID or service account ID
    TenantID()  string    // tenant UUID or slug
    OrgUnitID() uuid.UUID // primary org unit; uuid.Nil for tenant-wide viewers
}
```

Middleware extracts the session token from the `Authorization: Bearer <token>` header, hashes it, looks up the session in the database, and constructs a concrete `ViewerContext`. This is host-app code — the framework defines the interface but not the implementation.

---

## 43.10 Migration

```sql
-- internal/platform/iam/migrations/20240101000001_create_users.up.sql

CREATE TABLE users (
    id             uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id      uuid NOT NULL REFERENCES tenants(id),
    created_at     timestamptz NOT NULL DEFAULT NOW(),
    updated_at     timestamptz NOT NULL DEFAULT NOW(),
    email          text NOT NULL,
    name           text NOT NULL,
    password_hash  text NOT NULL,
    org_unit_id    uuid REFERENCES org_nodes(id),
    active         boolean NOT NULL DEFAULT true,
    last_login_at  timestamptz,
    UNIQUE (tenant_id, email)
);

ALTER TABLE users ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON users
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

-- internal/platform/iam/migrations/20240101000002_create_sessions.up.sql

CREATE TABLE sessions (
    id                    uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id             uuid NOT NULL REFERENCES tenants(id),
    created_at            timestamptz NOT NULL DEFAULT NOW(),
    updated_at            timestamptz NOT NULL DEFAULT NOW(),
    user_id               uuid NOT NULL REFERENCES users(id),
    token_hash            text NOT NULL UNIQUE,
    expires_at            timestamptz NOT NULL,
    ip_address            text,
    user_agent            text,
    org_unit_id_snapshot  uuid
);

-- Partial index: only one active (non-expired) session per token hash.
CREATE UNIQUE INDEX sessions_token_active_idx
    ON sessions (token_hash)
    WHERE expires_at > NOW();

ALTER TABLE sessions ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON sessions
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);
```

---

## 43.11 FAQ

**Q: How does session lookup work in middleware?**
A: Middleware hashes the `Authorization: Bearer` token with SHA-256, then queries `SELECT * FROM sessions WHERE token_hash = $1 AND expires_at > NOW()`. One indexed read. The session row contains `user_id`, `tenant_id`, and `org_unit_id_snapshot` — enough to build the `ViewerContext`.

**Q: What about refresh tokens?**
A: The current implementation uses single long-lived tokens with an `expires_at`. Refresh token rotation is a future enhancement. The `sessions` table schema supports it: add a `refresh_token_hash` column and a `refreshed_at` timestamp in a new migration.

**Q: Can a user be in multiple tenants?**
A: No. A `users` row belongs to exactly one tenant (enforced by `tenant_id` FK and RLS). If a person works for two businesses that both use Awo, they have two separate user accounts. There is no cross-tenant user federation in the base platform.

**Q: How are passwords reset?**
A: Via a custom endpoint `POST /iam/password-reset/request` that generates a time-limited reset token stored in Redis (not in the sessions table — it's not an auth session). The user clicks the link, the endpoint validates the token, and calls `bcrypt.GenerateFromPassword` to set a new `password_hash`. This flow is a service-layer operation, not a CRUD update.
