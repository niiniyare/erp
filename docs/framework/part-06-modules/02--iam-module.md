---
title: "Platform Module: IAM — Identity & Access Management"
part: "Part VI — Platform Entities"
chapter: 43
section: "platform-iam-module"
related:
  - "[Chapter 42: Tenant & Organisation](./platform-tenant-module.md)"
  - "[Chapter 41: Platform Module Overview](./platform-module-overview.md)"
  - "[Chapter 44: Feature Flags](./platform-feature-flags-module.md)"
---

# Chapter 43 — IAM: Identity & Access Management

> **Primary source for:** the `users`, `roles`, `user_roles`, and `sessions` tables, authentication flow, RBAC, the Casbin integration, and the difference between EntityDefinition policies (data-layer access) and Casbin policies (business-operation access).
>
> **Audience:** any developer who touches login, permissions, user management, or needs to understand how "is this user allowed to do this?" is answered.

---

## 43.1 What IAM Means

**Identity** — proving who you are. (Authentication / AuthN)
**Access** — deciding what you can do. (Authorization / AuthZ)
**Management** — administering both over time.

In plain language: IAM makes sure that:
- The cashier can record sales but cannot edit the chart of accounts.
- The HR manager can run payroll but cannot approve their own expense claim.
- A suspended user cannot log in even with the correct password.
- A user at Tenant A cannot see anything belonging to Tenant B.

The IAM module owns four entities: **User**, **Role**, **UserRole**, and **Session**. It also owns the service layer that wraps them with login/logout logic and integrates with Casbin for fine-grained RBAC.

---

## 43.2 Three Questions Every Request Answers

Before any handler runs, three questions are resolved in middleware:

```
Request arrives
    │
    ▼
1. AUTHENTICATION: Who are you?
   ─ Extract Bearer token
   ─ Look up session (Redis → PostgreSQL)
   ─ Load user + tenant + org unit
   ─ Build ViewerContext
    │
    ▼
2. AUTHORISATION: What can you do?
   ─ ViewerContext.HasRole() → Casbin RBAC
   ─ EntityDefinition.Policies → data-access control
    │
    ▼
3. CONFIGURATION: How should this behave?
   ─ Feature flags pre-loaded in session snapshot
   ─ Settings cached in Redis per tenant
    │
    ▼
Handler runs with full context
```

IAM owns steps 1 and 2. Feature Flags (Chapter 44) and Settings (Chapter 45) own step 3.

---

## 43.3 Directory Layout

```
framework/platform/iam/
├── iam.go          ← init() — registers all four EntityDefinitions
├── definition.go   ← UserDef, RoleDef, UserRoleDef, SessionDef
├── policy.go       ← allowSelf, allowTenantViewer, requireRole
├── hooks.go        ← hashPasswordIfChanged, syncCasbinOnStatusChange,
│                      protectSystemRoles, setAssignedBy, reloadCasbinPolicy
├── service.go      ← Service{}: Login, Logout, RevokeAllForTenant, ChangePassword
└── migrations/
    └── 20240101000010_create_iam.up.sql
```

---

## 43.4 User EntityDefinition

The User entity is the most policy-rich in the platform. Different operations have different rules:

```go
// framework/platform/iam/definition.go
package iam

import (
    "awo.so/framework/definition"
    "awo.so/framework/org"
)

var UserDef = definition.EntityDefinition{
    Name:     "user",
    Label:    "User",
    Module:   "Platform",
    Table:    "users",
    OrgScope: org.ScopeLevelTenant,

    // Status-based soft-delete: we use status="DELETED" instead of
    // a deleted_at column. Why? If we used SoftDelete=true, deleted
    // users would still occupy the UNIQUE(tenant_id, email) slot —
    // re-inviting the same email would fail. Status approach avoids this.
    SoftDelete: false,
    Audited:    true,

    Fields: []*definition.FieldDef{
        {Name: "email", Type: definition.FieldTypeData, Label: "Email", Required: true},
        {Name: "name",  Type: definition.FieldTypeData, Label: "Full Name", Required: true},

        // Sensitive: true means this field is:
        //   - Excluded from audit log diffs (you never log a password hash)
        //   - Excluded from SDUI form renders and API list responses
        //   - Excluded from SDUI search columns
        // The field is only read/written by explicit service code.
        {Name: "password_hash", Type: definition.FieldTypeData, Label: "Password Hash",
            Sensitive: true,
            Description: "bcrypt hash. Never the raw password. " +
                "Nullable for SSO-only users who authenticate via OAuth."},
        {Name: "mfa_secret", Type: definition.FieldTypeData, Label: "MFA Secret",
            Sensitive: true,
            Description: "TOTP shared secret. Encrypted at rest via pgcrypto. " +
                "Null until user enables MFA."},
        {Name: "mfa_enabled", Type: definition.FieldTypeBool, Label: "MFA Enabled",
            Default: "false"},

        {Name: "status", Type: definition.FieldTypeSelect, Label: "Status",
            Options:  []string{"INVITED", "ACTIVE", "SUSPENDED", "DELETED"},
            Required: true,
            Description: "INVITED: email sent, not yet logged in. " +
                "ACTIVE: normal operation. " +
                "SUSPENDED: admin action; cannot log in. " +
                "DELETED: soft-deleted; email slot freed after grace period."},

        {Name: "last_login_at", Type: definition.FieldTypeDateTime, Label: "Last Login"},

        // Users are assigned to an org node (branch/department).
        // This is their "home" unit for default filtering and org-scoped permissions.
        {Name: "org_node_id", Type: definition.FieldTypeLink, Label: "Org Unit",
            TargetEntity: "org_node"},
    },

    Policies: []definition.PolicyDef{
        // Rule 1: system callers always pass.
        // Reason: provisioning workflows, the IAM service itself (e.g., login
        // checks the password), background jobs — all need unobstructed access.
        definition.Policy(definition.OpAll, definition.AllowSystem),

        // Rule 2: tenant admins can do everything with users.
        // Reason: admins manage the team — invite, suspend, change roles.
        definition.Policy(definition.OpAll, requireRole("admin")),

        // Rule 3: a user can READ their own record.
        // Reason: profile page, "my account" — users need to see their own data.
        definition.Policy(definition.OpRead, allowSelf),

        // Rule 4: a user can UPDATE their own record (non-sensitive fields only).
        // Reason: users update their display name, timezone preference, etc.
        // Sensitive field changes (password, MFA) go through dedicated service
        // endpoints with extra validation (current password required, etc.).
        definition.Policy(definition.OpUpdate, allowSelf),

        // Rule 5: deny everything else.
        // A regular user cannot read other users' records.
        // A regular user cannot delete their own account (admin action).
        definition.Policy(definition.OpAll, definition.DenyAll),
    },

    Hooks: []definition.HookDef{
        {
            // Detects if a raw password was submitted (not yet hashed).
            // Hashes it with bcrypt before the row is written.
            // Idempotent: if the value already looks like a bcrypt hash, skip.
            Name: "hash_password_on_write",
            Ops:  definition.OpCreate | definition.OpUpdate,
            When: definition.HookBefore,
            Fn:   hashPasswordIfChanged,
        },
        {
            // After a user's status changes to SUSPENDED or DELETED:
            // invalidates their Casbin policy entries so no cached permission
            // check can still grant them access.
            Name: "sync_casbin_on_status_change",
            Ops:  definition.OpUpdate,
            When: definition.HookAfter,
            Fn:   syncCasbinOnStatusChange,
        },
    },
}
```

---

## 43.5 Role EntityDefinition

```go
var RoleDef = definition.EntityDefinition{
    Name:     "role",
    Label:    "Role",
    Module:   "Platform",
    Table:    "roles",
    OrgScope: org.ScopeLevelTenant,
    Audited:  true,

    Fields: []*definition.FieldDef{
        {Name: "name",        Type: definition.FieldTypeData,      Label: "Role Name", Required: true},
        {Name: "description", Type: definition.FieldTypeSmallText, Label: "Description"},
        {
            Name:    "is_system",
            Type:    definition.FieldTypeBool,
            Label:   "System Role",
            Default: "false",
            Description: "System roles are seeded at provisioning and cannot be " +
                "deleted or renamed. They define the built-in permission set. " +
                "Tenant admins can create additional custom roles alongside them.",
        },
    },

    Policies: []definition.PolicyDef{
        definition.Policy(definition.OpAll,  definition.AllowSystem),

        // All authenticated users can READ roles.
        // Reason: "show me my own roles", role picker in user management UI.
        definition.Policy(definition.OpRead, allowTenantViewer),

        // Only admins can create, modify, or delete roles.
        // System roles get additional protection from the protectSystemRoles hook.
        definition.Policy(definition.OpWrite|definition.OpDelete, requireRole("admin")),
    },

    Hooks: []definition.HookDef{
        {
            // Prevents modification or deletion of system roles.
            // Even an admin cannot delete the "admin" role — if they could,
            // the tenant would be permanently locked out.
            Name: "protect_system_roles",
            Ops:  definition.OpUpdate | definition.OpDelete,
            When: definition.HookBefore,
            Fn:   protectSystemRoles,
        },
    },
}
```

### Built-In System Roles

These roles are seeded during tenant provisioning. They cannot be deleted.

| Role | What they can do |
|---|---|
| `admin` | Full access within the tenant. Manage users, roles, settings, all modules. |
| `accountant` | All Finance operations: GL, AP/AR, journal entries, reconciliation. |
| `cashier` | POS sales, cash receipts. Cannot edit historical records. |
| `salesperson` | Create and manage sales orders, view customers. No financial edit access. |
| `hr_manager` | Employee records, leave, payroll processing. |
| `store_keeper` | Inventory: receive stock, issue stock, run stocktake. Cannot purchase. |
| `viewer` | Read-only access to all modules the tenant has activated. |
| `auditor` | Read-only access to audit logs and financial reports. Time-limited. |

---

## 43.6 UserRole EntityDefinition

UserRole is the junction between a user and a role. It is an entity in its own right — not just a foreign key — because it carries metadata about the assignment.

```go
var UserRoleDef = definition.EntityDefinition{
    Name:     "user_role",
    Label:    "User Role Assignment",
    Module:   "Platform",
    Table:    "user_roles",
    OrgScope: org.ScopeLevelTenant,
    Audited:  true, // every role change is audited

    Fields: []*definition.FieldDef{
        {Name: "user_id", Type: definition.FieldTypeLink, Label: "User",
            TargetEntity: "user", Required: true},
        {Name: "role_id", Type: definition.FieldTypeLink, Label: "Role",
            TargetEntity: "role", Required: true},
        {Name: "assigned_by", Type: definition.FieldTypeLink, Label: "Assigned By",
            TargetEntity: "user",
            Description:  "Set automatically by the setAssignedBy hook."},
        {Name: "assigned_at", Type: definition.FieldTypeDateTime, Label: "Assigned At"},
        {
            Name:        "expires_at",
            Type:        definition.FieldTypeDateTime,
            Label:       "Expires At",
            Description: "Optional. Enables time-limited role grants. " +
                "An external auditor might receive the 'auditor' role for 30 days. " +
                "A daily job revokes expired grants. Nil = permanent.",
        },
    },

    Policies: []definition.PolicyDef{
        definition.Policy(definition.OpAll,  definition.AllowSystem),
        // Only admins can view and manage role assignments.
        // Regular users do not need to see who has what role.
        definition.Policy(definition.OpRead,              requireRole("admin")),
        definition.Policy(definition.OpWrite|definition.OpDelete, requireRole("admin")),
    },

    Hooks: []definition.HookDef{
        {
            // On Create: automatically records who made the assignment
            // (viewer.ActorID()) and when (now()). The caller does not set these.
            Name: "set_assigned_by",
            Ops:  definition.OpCreate,
            When: definition.HookBefore,
            Fn:   setAssignedBy,
        },
        {
            // After Create or Delete: immediately reloads this tenant's Casbin
            // policy in memory. The role change takes effect on the very next
            // request from that user — no cache TTL wait.
            Name: "reload_casbin_after_assignment",
            Ops:  definition.OpCreate | definition.OpDelete,
            When: definition.HookAfter,
            Fn:   reloadCasbinPolicy,
        },
    },
}
```

---

## 43.7 Session EntityDefinition

```go
var SessionDef = definition.EntityDefinition{
    Name:     "session",
    Label:    "Session",
    Module:   "Platform",
    Table:    "sessions",
    OrgScope: org.ScopeLevelTenant,

    // Sessions are NOT audited.
    // Why: sessions ARE the audit mechanism. Auditing sessions would create
    // infinite regress (an audit entry for every login → audit the audit → ...).
    // Instead, suspicious session activity is detected via the session rows themselves.
    Audited:    false,
    SoftDelete: false,

    Fields: []*definition.FieldDef{
        {Name: "user_id",     Type: definition.FieldTypeLink,      Label: "User",
            TargetEntity: "user", Required: true},
        {Name: "token",       Type: definition.FieldTypeData,      Label: "Session Token",
            Sensitive:   true,
            Description: "Opaque random token (32 bytes hex). Never store or log raw."},
        {Name: "ip",          Type: definition.FieldTypeData,      Label: "IP Address"},
        {Name: "user_agent",  Type: definition.FieldTypeSmallText, Label: "User Agent"},
        {Name: "expires_at",  Type: definition.FieldTypeDateTime,  Label: "Expires At", Required: true},
        {Name: "last_seen_at",Type: definition.FieldTypeDateTime,  Label: "Last Active"},
        {Name: "revoked_at",  Type: definition.FieldTypeDateTime,  Label: "Revoked At",
            Description: "Set on logout or forced revocation. Null = active."},
        {Name: "revoked_by",  Type: definition.FieldTypeLink,      Label: "Revoked By",
            TargetEntity: "user",
            Description:  "The user or admin who revoked the session."},
    },

    Policies: []definition.PolicyDef{
        // The IAM service creates, updates, and deletes sessions.
        // No direct user CRUD through the generic API.
        definition.Policy(definition.OpAll, definition.AllowSystem),

        // A user can read their own sessions.
        // Reason: the "Active Sessions" screen in account settings shows
        // the user which devices are logged in. They can then revoke suspicious ones.
        definition.Policy(definition.OpRead, allowSelf),

        definition.Policy(definition.OpAll, definition.DenyAll),
    },
}
```

---

## 43.8 Policy Helper Implementations

```go
// framework/platform/iam/policy.go
package iam

import (
    "context"
    "awo.so/framework/definition"
)

// allowTenantViewer: any authenticated viewer (non-empty TenantID) passes.
func allowTenantViewer(_ context.Context, v definition.ViewerContext, _ definition.Op, _ definition.Record) error {
    if v.TenantID() == "" {
        return definition.ErrDeny
    }
    return definition.ErrAllow
}

// allowSelf: passes only when the record being accessed belongs to the viewer.
// The record must expose an "id" field that matches viewer.ActorID().
func allowSelf(_ context.Context, v definition.ViewerContext, _ definition.Op, rec definition.Record) error {
    if rec == nil {
        // List-level check: we cannot determine "self" without a record.
        // Abstain; let other policies handle list access.
        return definition.ErrSkip
    }
    recordID, _ := rec.Get("id")
    if recordID != nil && recordID.(string) == v.ActorID() {
        return definition.ErrAllow
    }
    return definition.ErrSkip // not my record; try next policy
}

// requireRole: factory that builds a PolicyFunc for a specific role name.
func requireRole(role string) definition.PolicyFunc {
    return func(_ context.Context, v definition.ViewerContext, _ definition.Op, _ definition.Record) error {
        if v.IsSystem() || v.HasRole(role) {
            return definition.ErrAllow
        }
        return definition.ErrDeny
    }
}
```

---

## 43.9 The IAM Service Layer

The generic CRUD API (auto-generated from EntityDefinitions) handles admin operations like listing users or updating a user's name. But login, logout, password change, and MFA setup are stateful operations that require specific logic. These live in the IAM service.

```go
// framework/platform/iam/service.go
package iam

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "errors"
    "time"

    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"
    "awo.so/framework/persistence/pgstore"
)

type Service struct {
    users    pgstore.EntityStore // bound to UserDef
    sessions pgstore.EntityStore // bound to SessionDef
    redis    RedisClient
}

// Login authenticates a user credential pair and returns an opaque session token.
//
// Flow:
//   1. Load user by (tenant_id, email)
//   2. Verify bcrypt password
//   3. Check user status (ACTIVE only)
//   4. Optionally verify TOTP code if mfa_enabled
//   5. Create session row in PostgreSQL
//   6. Write session snapshot to Redis (token → session JSON)
//   7. Return token
func (s *Service) Login(ctx context.Context, req LoginRequest) (string, error) {
    // Step 1: find user (system caller bypasses policies)
    users, err := s.users.List(ctx, pgstore.ListOptions{
        Filter: map[string]any{"email": req.Email},
        Limit:  1,
    })
    if err != nil || len(users.Items) == 0 {
        return "", ErrInvalidCredentials // same error for "not found" and "wrong password"
    }
    user := users.Items[0]

    // Step 2: verify password
    hash, _ := user.Get("password_hash")
    if hash == nil {
        return "", ErrSSOUser // SSO-only account; use OAuth flow
    }
    if err := bcrypt.CompareHashAndPassword([]byte(hash.(string)), []byte(req.Password)); err != nil {
        return "", ErrInvalidCredentials
    }

    // Step 3: check status
    status, _ := user.Get("status")
    if status.(string) != "ACTIVE" {
        return "", ErrUserNotActive
    }

    // Step 4: TOTP (if enabled)
    mfaEnabled, _ := user.Get("mfa_enabled")
    if mfaEnabled.(bool) {
        if err := verifyTOTP(user, req.TOTPCode); err != nil {
            return "", ErrInvalidMFA
        }
    }

    // Step 5: create session row
    token := generateToken() // 32 random bytes as hex
    userID, _ := user.Get("id")
    tenantID, _ := user.Get("tenant_id")

    sessionRec := pgstore.NewMutableRecord(SessionDef)
    sessionRec.Set("user_id",     userID)
    sessionRec.Set("token",       token)
    sessionRec.Set("ip",          req.IP)
    sessionRec.Set("user_agent",  req.UserAgent)
    sessionRec.Set("expires_at",  time.Now().Add(24*time.Hour))
    sessionRec.Set("last_seen_at",time.Now())

    if err := s.sessions.Create(ctx, sessionRec); err != nil {
        return "", err
    }

    // Step 6: write Redis cache (hot path for every subsequent request)
    s.cacheSession(ctx, tenantID.(string), token, sessionRec)

    return token, nil
}

// RevokeAllForTenant bulk-revokes every active session for a tenant.
// Called when a tenant is suspended. Single SQL UPDATE — no loop.
func (s *Service) RevokeAllForTenant(ctx context.Context, tenantID uuid.UUID) error {
    err := s.sessions.BulkUpdate(ctx,
        map[string]any{"tenant_id": tenantID, "revoked_at": nil}, // filter: active sessions
        map[string]any{"revoked_at": time.Now()},                 // patch: revoke now
    )
    if err != nil {
        return err
    }
    // Also clear Redis entries for this tenant's sessions.
    return s.redis.Del(ctx, "sessions:"+tenantID.String()+":*")
}

// ChangePassword verifies the current password, then sets a new bcrypt hash.
// This is a service method — not the generic CRUD Update — because it requires
// the current password to be verified before writing the new one.
func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, current, newPwd string) error {
    user, err := s.users.FindByID(ctx, userID)
    if err != nil {
        return err
    }
    hash, _ := user.Get("password_hash")
    if err := bcrypt.CompareHashAndPassword([]byte(hash.(string)), []byte(current)); err != nil {
        return ErrInvalidCredentials
    }
    newHash, _ := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
    rec := pgstore.NewMutableRecord(UserDef)
    rec.Set("id", userID)
    rec.Set("password_hash", string(newHash))
    return s.users.Update(ctx, rec)
}

func generateToken() string {
    b := make([]byte, 32)
    rand.Read(b)
    return hex.EncodeToString(b)
}
```

---

## 43.10 Session Storage — Dual Persistence

Sessions are stored in two places simultaneously:

### Redis (Hot Path)

Key: `session:{tenant_id}:{token}`
Value: JSON blob containing user_id, tenant_id, roles, feature flags, org_unit_id.

Every HTTP request looks up the token in Redis first. If found, the ViewerContext is built from the cached JSON — zero database queries for authentication. Latency: sub-millisecond.

### PostgreSQL (Audit Trail)

The `sessions` table persists every session row. If Redis is flushed or restarted, sessions are rebuilt from PostgreSQL on next lookup. PostgreSQL is the source of truth; Redis is the cache.

```
Request arrives with token
    │
    ▼
Redis HGET session:{tenant_id}:{token}
    │
    ├── HIT → build ViewerContext from JSON (0 DB queries)
    │
    └── MISS → query sessions table by token
              → if found and not expired/revoked: cache in Redis → continue
              → if not found or revoked: return 401
```

### Redis Key Design

Including `{tenant_id}` in the key enables fast bulk invalidation:

```go
// Revoking all sessions for a suspended tenant:
redis.Del(ctx, "session:"+tenantID.String()+":*") // one SCAN+DEL, not N individual deletes
```

---

## 43.11 Two-Layer Access Control

Developers new to Awo are sometimes confused about when to use **EntityDefinition.Policies** versus **Casbin**. Here is the clear distinction:

| Layer | Mechanism | Answers | Example |
|---|---|---|---|
| **Data Access** | `EntityDefinition.Policies` | "Can this caller touch the `users` table at all?" | Admin reads users: ✅. Regular user reads other user: ❌. |
| **Business Operation** | Casbin RBAC | "Can this user perform this specific business action on this record?" | "Can this finance manager *submit* (not just read) this invoice?" |

### Why Two Layers?

EntityDefinition policies are **coarse-grained** (table-level). Casbin is **fine-grained** (operation × resource × record state).

Example: a `salesperson` can read invoices (EntityDefinition policy allows it). But they can only *submit* invoices they created, not invoices in `APPROVED` status (Casbin rule). The EntityDefinition policy doesn't need to know about invoice status. Casbin handles it.

### Casbin Sync

Casbin's in-memory model is loaded at startup from a PostgreSQL table (`casbin_rules`). When a UserRole is created or deleted, the `reloadCasbinPolicy` AfterHook immediately updates the in-memory model:

```go
// framework/platform/iam/hooks.go
func reloadCasbinPolicy(ctx context.Context, v definition.ViewerContext, op definition.Op, rec definition.Record) error {
    tenantID, _ := rec.Get("tenant_id")

    // Reload only this tenant's Casbin policies (not the entire model).
    return casbinEnforcer.LoadFilteredPolicy(context.Background(),
        casbinmodel.FilteredAdapter{V0: tenantID.(string)})
}
```

---

## 43.12 Migration

```sql
-- framework/platform/iam/migrations/20240101000010_create_iam.up.sql

-- ── Users ──────────────────────────────────────────────────────────────────
CREATE TABLE users (
    id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     uuid        NOT NULL REFERENCES tenants(id),
    email         text        NOT NULL,
    name          text        NOT NULL,

    -- nullable: SSO-only users authenticate via OAuth and have no password.
    password_hash text,

    status        text        NOT NULL DEFAULT 'INVITED'
                              CHECK (status IN ('INVITED','ACTIVE','SUSPENDED','DELETED')),

    -- TOTP shared secret; encrypted at rest with pgcrypto.
    -- NULL until user enables MFA.
    mfa_secret    text,
    mfa_enabled   boolean     NOT NULL DEFAULT false,

    last_login_at timestamptz,
    org_node_id   uuid        REFERENCES org_nodes(id),

    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),

    -- Email is unique per tenant, not globally.
    -- Same person at two different franchise tenants has two separate accounts.
    UNIQUE (tenant_id, email)
);

ALTER TABLE users ENABLE ROW LEVEL SECURITY;
CREATE POLICY users_tenant_isolation ON users
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

CREATE TRIGGER set_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ── Roles ──────────────────────────────────────────────────────────────────
CREATE TABLE roles (
    id          uuid    PRIMARY KEY DEFAULT gen_random_uuid(),

    -- NULL tenant_id = platform-wide system role (admin, viewer, etc.)
    -- These are seeded in the base migration and shared across all tenants.
    -- Non-NULL tenant_id = custom role defined by a specific tenant.
    tenant_id   uuid    REFERENCES tenants(id),

    name        text    NOT NULL,
    description text,
    is_system   boolean NOT NULL DEFAULT false,
    created_at  timestamptz NOT NULL DEFAULT now(),

    UNIQUE (tenant_id, name)
);

ALTER TABLE roles ENABLE ROW LEVEL SECURITY;
-- System roles (tenant_id IS NULL) are visible to all tenants.
-- Tenant-defined roles (tenant_id IS NOT NULL) are visible only to their tenant.
CREATE POLICY roles_tenant_isolation ON roles
    USING (tenant_id IS NULL OR tenant_id = current_setting('app.current_tenant_id')::uuid);

-- Seed system roles (no tenant_id = available to all tenants)
INSERT INTO roles (name, description, is_system) VALUES
    ('admin',        'Full access within the tenant',                   true),
    ('accountant',   'Finance module: GL, AP/AR, journal entries',      true),
    ('cashier',      'POS sales and cash receipts',                     true),
    ('salesperson',  'Sales orders and customer management',             true),
    ('hr_manager',   'Employee records, leave, payroll',                 true),
    ('store_keeper', 'Inventory: receive, issue, stocktake',             true),
    ('viewer',       'Read-only access to all active modules',           true),
    ('auditor',      'Read-only audit logs and financial reports',       true);


-- ── UserRoles ──────────────────────────────────────────────────────────────
CREATE TABLE user_roles (
    -- Composite PK prevents duplicate assignments.
    -- A user cannot be assigned the same role twice.
    user_id     uuid        NOT NULL REFERENCES users(id),
    role_id     uuid        NOT NULL REFERENCES roles(id),
    tenant_id   uuid        NOT NULL REFERENCES tenants(id),
    assigned_by uuid        REFERENCES users(id),
    assigned_at timestamptz NOT NULL DEFAULT now(),
    -- NULL expires_at = permanent grant.
    -- A daily Temporal schedule revokes expired grants.
    expires_at  timestamptz,

    PRIMARY KEY (user_id, role_id)
);

ALTER TABLE user_roles ENABLE ROW LEVEL SECURITY;
CREATE POLICY user_roles_tenant_isolation ON user_roles
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);


-- ── Sessions ───────────────────────────────────────────────────────────────
CREATE TABLE sessions (
    id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    uuid        NOT NULL REFERENCES tenants(id),
    user_id      uuid        NOT NULL REFERENCES users(id),
    token        text        NOT NULL,
    ip           text,
    user_agent   text,
    expires_at   timestamptz NOT NULL,
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    revoked_at   timestamptz,
    revoked_by   uuid        REFERENCES users(id),
    created_at   timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE sessions ENABLE ROW LEVEL SECURITY;
CREATE POLICY sessions_tenant_isolation ON sessions
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

-- Partial unique index on token for active sessions only.
--
-- Why partial? Revoked tokens are dead. We don't need uniqueness enforced
-- on dead tokens, and excluding them keeps the index small and fast.
-- The WHERE clause also means the index is only consulted for active session lookups
-- (the 99% case) — not for historical queries on the full table.
CREATE UNIQUE INDEX sessions_token_active_idx ON sessions (token)
    WHERE revoked_at IS NULL AND expires_at > now();

-- Index for the "show me all my active sessions" query.
CREATE INDEX sessions_user_active_idx ON sessions (tenant_id, user_id)
    WHERE revoked_at IS NULL AND expires_at > now();
```

---

## 43.13 Common Mistakes to Avoid

**"Should I call Casbin enforce inside a PolicyFunc?"**

No. PolicyFuncs control data-layer access (can you query this table?). Casbin controls business-operation access (can you perform this specific action?). Mixing them creates circular dependency and confusing logic. Keep them separate.

**"Should I check roles inside a hook?"**

Only when the hook implements a business rule (e.g., `protectSystemRoles` — a domain invariant). For access control decisions, use a PolicyFunc. Hooks should not be aware of who is calling them; they enforce *what is allowed*, not *who is allowed*.

**"Can I query the users table directly from the Finance module?"**

Never directly. Import and call the IAM service interface (`iam.Service`). Cross-module direct DB queries break isolation — if the IAM table schema changes, every module that queried it directly breaks. Service interfaces provide a stable contract.

**"What happens when a session expires without explicit logout?"**

The partial index `WHERE expires_at > now()` means expired sessions stop matching the fast-path lookup automatically. The auth middleware gets a cache miss, queries PostgreSQL, finds `expires_at < now()`, and returns 401. No explicit cleanup is needed for correctness — though a periodic job vacuum-deletes old expired rows to keep the table small.
