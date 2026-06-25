---
title: "Implementing Platform Modules with EntityDefinition"
part: "Part VI — Platform Entities"
chapter: 41b
section: "implementing-platform-modules"
related:
  - "[Chapter 2: The EntityDefinition](../part-01-foundations/02-entity-definition.md)"
  - "[Chapter 34: Platform Entities Reference](./platform-entities.md)"
  - "[Chapter 11: Database Migrations](../part-02-entity-system/11-migrations.md)"
  - "[Chapter 5: Privacy Policies](../part-01-foundations/05-privacy.md)"
---

# Chapter 41b — Implementing Platform Modules

> **Purpose of this chapter:** Platform modules (Tenant, IAM, Feature Flags, Settings, Audit Log, Metadata) are built *exactly* like any business module — using `definition.EntityDefinition`, golang-migrate SQL files, privacy policies, and hooks. There is no special framework path for them. This chapter is the definitive guide for AwoERP core team members building or extending a platform module.

If you are a Finance, CRM, or HR module developer, this chapter doubles as a worked example of the most complex EntityDefinitions in the codebase.

---

## 39.1 Module Layout Convention

Every module — platform or business — follows the same directory layout:

```
internal/platform/<module>/
├── <module>.go          ← package entry point, init() registration
├── definition.go        ← EntityDefinition declarations
├── policy.go            ← PolicyFunc implementations
├── hooks.go             ← HookDef implementations
├── service.go           ← thin service layer on top of EntityStore
├── handler.go           ← HTTP handlers if the module has custom endpoints
└── migrations/
    ├── 20240101000001_create_<entity>.up.sql
    └── 20240101000001_create_<entity>.down.sql
```

Business modules hosted in separate repos follow the same layout under `internal/<module>/`.

---

## 39.2 Registration Pattern

Every module registers its EntityDefinitions in an `init()` function. The framework calls `definition.All()` at startup to enumerate every registered definition for API mounting, migration generation, and SDUI rendering.

```go
// internal/platform/tenant/tenant.go
package tenant

import "awo.so/framework/definition"

func init() {
    definition.Register(&TenantDef)
    definition.Register(&OrgNodeDef)
}
```

`definition.Register` panics at startup if the definition fails `Validate()` — a duplicate name, unknown field type, or empty edge target causes a hard crash, not a silent bug.

---

## 39.3 Tenant Module

The Tenant entity is **global** (no `tenant_id` column). It sits outside the RLS system and is managed exclusively by platform administrators and the provisioning workflow.

### 39.3.1 EntityDefinition

```go
// internal/platform/tenant/definition.go
package tenant

import (
    "awo.so/framework/definition"
    "awo.so/framework/org"
)

var TenantDef = definition.EntityDefinition{
    Name:        "tenant",
    Label:       "Tenant",
    Description: "Top-level isolation boundary. One tenant = one business on the platform.",
    Module:      "Platform",
    Table:       "tenants",
    OrgScope:    org.ScopeLevelGlobal, // no tenant_id column; shared across all tenants

    SoftDelete: false, // Use Status=ARCHIVED instead of deletion
    Audited:    true,

    Fields: []*definition.FieldDef{
        {Name: "slug",           Type: definition.FieldTypeData,     Label: "Slug",            Required: true},
        {Name: "name",           Type: definition.FieldTypeData,     Label: "Name",            Required: true},
        {Name: "domain",         Type: definition.FieldTypeData,     Label: "Custom Domain"},
        {Name: "plan",           Type: definition.FieldTypeSelect,   Label: "Plan",
            Options: []string{"starter", "growth", "enterprise"}, Required: true},
        {Name: "status",         Type: definition.FieldTypeSelect,   Label: "Status",
            Options: []string{"PENDING", "ACTIVE", "SUSPENDED", "ARCHIVED"}, Required: true},

        // Kenya business identity
        {Name: "kra_pin",        Type: definition.FieldTypeData,     Label: "KRA PIN"},
        {Name: "company_reg_no", Type: definition.FieldTypeData,     Label: "Company Reg No"},

        // Locale defaults
        {Name: "timezone",       Type: definition.FieldTypeData,     Label: "Timezone",        Default: "Africa/Nairobi"},
        {Name: "currency",       Type: definition.FieldTypeData,     Label: "Default Currency", Default: "KES"},
        {Name: "fiscal_year_end",Type: definition.FieldTypeData,     Label: "Fiscal Year End", Default: "06-30"},

        // Lifecycle timestamps
        {Name: "trial_ends_at",  Type: definition.FieldTypeDateTime, Label: "Trial Ends"},
        {Name: "suspended_at",   Type: definition.FieldTypeDateTime, Label: "Suspended At"},
        {Name: "suspend_reason", Type: definition.FieldTypeSmallText,Label: "Suspend Reason"},
        {Name: "archived_at",    Type: definition.FieldTypeDateTime, Label: "Archived At"},
        {Name: "provisioned_at", Type: definition.FieldTypeDateTime, Label: "Provisioned At"},
    },

    Policies: []definition.PolicyDef{
        // System callers (provisioning workflow, platform admin API) have full access.
        definition.Policy(definition.OpAll, definition.AllowSystem),
        // All other callers are denied — tenants cannot see or edit each other.
        definition.Policy(definition.OpAll, definition.DenyAll),
    },

    Hooks: []definition.HookDef{
        {
            Name: "validate_status_transition",
            Ops:  definition.OpUpdate,
            When: definition.HookBefore,
            Fn:   validateTenantStatusTransition,
        },
        {
            Name: "invalidate_sessions_on_suspend",
            Ops:  definition.OpUpdate,
            When: definition.HookAfter,
            Fn:   invalidateSessionsOnSuspend,
        },
    },
}

var OrgNodeDef = definition.EntityDefinition{
    Name:     "org_node",
    Label:    "Organisation Node",
    Module:   "Platform",
    Table:    "org_nodes",
    OrgScope: org.ScopeLevelTenant,
    Audited:  true,

    Fields: []*definition.FieldDef{
        {Name: "parent_id",  Type: definition.FieldTypeLink,   Label: "Parent",  TargetEntity: "org_node"},
        {Name: "type",       Type: definition.FieldTypeSelect, Label: "Type",
            Options: []string{"COMPANY", "DIVISION", "DEPARTMENT", "BRANCH", "COST_CENTRE"}, Required: true},
        {Name: "code",       Type: definition.FieldTypeData,   Label: "Code",    Required: true},
        {Name: "name",       Type: definition.FieldTypeData,   Label: "Name",    Required: true},
        {Name: "path",       Type: definition.FieldTypeData,   Label: "Path"},   // materialised path, set by hook
        {Name: "is_active",  Type: definition.FieldTypeBool,   Label: "Active",  Default: "true"},
    },

    Policies: []definition.PolicyDef{
        definition.Policy(definition.OpAll, definition.AllowSystem),
        definition.Policy(definition.OpRead, allowTenantViewer),
        definition.Policy(definition.OpWrite|definition.OpDelete, requireRole("admin")),
    },

    Hooks: []definition.HookDef{
        {
            Name: "materialise_path",
            Ops:  definition.OpCreate | definition.OpUpdate,
            When: definition.HookBefore,
            Fn:   materialisePath,
        },
    },
}
```

### 39.3.2 Policy Helpers

```go
// internal/platform/tenant/policy.go
package tenant

import (
    "context"
    "awo.so/framework/definition"
)

// allowTenantViewer grants read access to any authenticated, non-anonymous viewer
// within the same tenant. Used for org_nodes, which all tenant users need.
func allowTenantViewer(_ context.Context, v definition.ViewerContext, _ definition.Op, _ definition.Record) error {
    if v.TenantID() == "" {
        return definition.ErrDeny
    }
    return definition.ErrAllow
}

// requireRole returns a PolicyFunc that allows only viewers holding the named role.
func requireRole(role string) definition.PolicyFunc {
    return func(_ context.Context, v definition.ViewerContext, _ definition.Op, _ definition.Record) error {
        if v.IsSystem() || v.HasRole(role) {
            return definition.ErrAllow
        }
        return definition.ErrDeny
    }
}
```

### 39.3.3 Migration

```sql
-- internal/platform/tenant/migrations/20240101000001_create_tenants.up.sql

CREATE TABLE tenants (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    slug            text        NOT NULL UNIQUE,
    name            text        NOT NULL,
    domain          text,
    plan            text        NOT NULL DEFAULT 'starter',
    status          text        NOT NULL DEFAULT 'PENDING',

    kra_pin         text,
    company_reg_no  text,

    timezone        text        NOT NULL DEFAULT 'Africa/Nairobi',
    currency        text        NOT NULL DEFAULT 'KES',
    fiscal_year_end text        NOT NULL DEFAULT '06-30',

    trial_ends_at   timestamptz,
    suspended_at    timestamptz,
    suspend_reason  text,
    archived_at     timestamptz,
    provisioned_at  timestamptz,

    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

-- No RLS on tenants — it is a global table managed only by platform callers.
-- The application layer enforces access via TenantDef.Policies.

CREATE TRIGGER set_updated_at
    BEFORE UPDATE ON tenants
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE org_nodes (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid        NOT NULL REFERENCES tenants(id),
    parent_id   uuid        REFERENCES org_nodes(id),
    type        text        NOT NULL,
    code        text        NOT NULL,
    name        text        NOT NULL,
    path        text        NOT NULL DEFAULT '',
    is_active   boolean     NOT NULL DEFAULT true,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),

    UNIQUE (tenant_id, code)
);

ALTER TABLE org_nodes ENABLE ROW LEVEL SECURITY;

CREATE POLICY org_nodes_tenant_isolation ON org_nodes
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

CREATE TRIGGER set_updated_at
    BEFORE UPDATE ON org_nodes
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

---

## 39.4 IAM Module

IAM owns Users, Roles, Permissions, UserRoles, and Sessions. The complexity here is in the policies — different operations have different access rules within the same entity.

### 39.4.1 User EntityDefinition

```go
// internal/platform/iam/definition.go
package iam

import (
    "awo.so/framework/definition"
    "awo.so/framework/org"
)

var UserDef = definition.EntityDefinition{
    Name:       "user",
    Label:      "User",
    Module:     "Platform",
    Table:      "users",
    OrgScope:   org.ScopeLevelTenant,
    SoftDelete: false, // status="DELETED" is used instead
    Audited:    true,

    Fields: []*definition.FieldDef{
        {Name: "email",         Type: definition.FieldTypeData,     Label: "Email",     Required: true},
        {Name: "name",          Type: definition.FieldTypeData,     Label: "Name",      Required: true},
        {Name: "password_hash", Type: definition.FieldTypeData,     Label: "Password",  Sensitive: true},
        {Name: "status",        Type: definition.FieldTypeSelect,   Label: "Status",
            Options: []string{"INVITED", "ACTIVE", "SUSPENDED", "DELETED"}, Required: true},
        {Name: "mfa_secret",    Type: definition.FieldTypeData,     Label: "MFA Secret", Sensitive: true},
        {Name: "mfa_enabled",   Type: definition.FieldTypeBool,     Label: "MFA Enabled", Default: "false"},
        {Name: "last_login_at", Type: definition.FieldTypeDateTime, Label: "Last Login"},
        {Name: "org_node_id",   Type: definition.FieldTypeLink,     Label: "Org Node",  TargetEntity: "org_node"},
    },

    Policies: []definition.PolicyDef{
        // System always passes (provisioning, workflow, service tokens).
        definition.Policy(definition.OpAll, definition.AllowSystem),

        // Tenant admins can do everything with users.
        definition.Policy(definition.OpAll, requireRole("admin")),

        // Any authenticated user can read their own record.
        definition.Policy(definition.OpRead, allowSelf),

        // Any authenticated user can update their own non-sensitive fields.
        // Sensitive field writes (password_hash, mfa_secret) are handled
        // by dedicated service endpoints, not the generic CRUD API.
        definition.Policy(definition.OpUpdate, allowSelf),

        // Deny all other access.
        definition.Policy(definition.OpAll, definition.DenyAll),
    },

    Hooks: []definition.HookDef{
        {
            Name: "hash_password_on_write",
            Ops:  definition.OpCreate | definition.OpUpdate,
            When: definition.HookBefore,
            Fn:   hashPasswordIfChanged,
        },
        {
            Name: "sync_casbin_on_status_change",
            Ops:  definition.OpUpdate,
            When: definition.HookAfter,
            Fn:   syncCasbinOnStatusChange,
        },
    },
}

var RoleDef = definition.EntityDefinition{
    Name:     "role",
    Label:    "Role",
    Module:   "Platform",
    Table:    "roles",
    OrgScope: org.ScopeLevelTenant,
    Audited:  true,

    Fields: []*definition.FieldDef{
        {Name: "name",        Type: definition.FieldTypeData,   Label: "Name",        Required: true},
        {Name: "description", Type: definition.FieldTypeSmallText, Label: "Description"},
        {Name: "is_system",   Type: definition.FieldTypeBool,   Label: "System Role", Default: "false"},
    },

    Policies: []definition.PolicyDef{
        definition.Policy(definition.OpAll,  definition.AllowSystem),
        definition.Policy(definition.OpRead, allowTenantViewer),
        // Only admins may create/modify/delete roles.
        // System roles (is_system=true) are additionally protected by a hook.
        definition.Policy(definition.OpWrite|definition.OpDelete, requireRole("admin")),
    },

    Hooks: []definition.HookDef{
        {
            Name: "protect_system_roles",
            Ops:  definition.OpUpdate | definition.OpDelete,
            When: definition.HookBefore,
            Fn:   protectSystemRoles,
        },
    },
}

var UserRoleDef = definition.EntityDefinition{
    Name:     "user_role",
    Label:    "User Role Assignment",
    Module:   "Platform",
    Table:    "user_roles",
    OrgScope: org.ScopeLevelTenant,
    Audited:  true,

    Fields: []*definition.FieldDef{
        {Name: "user_id",     Type: definition.FieldTypeLink, Label: "User",        TargetEntity: "user",    Required: true},
        {Name: "role_id",     Type: definition.FieldTypeLink, Label: "Role",        TargetEntity: "role",    Required: true},
        {Name: "assigned_by", Type: definition.FieldTypeLink, Label: "Assigned By", TargetEntity: "user"},
        {Name: "assigned_at", Type: definition.FieldTypeDateTime, Label: "Assigned At"},
        {Name: "expires_at",  Type: definition.FieldTypeDateTime, Label: "Expires At"}, // temporal roles
    },

    Policies: []definition.PolicyDef{
        definition.Policy(definition.OpAll,  definition.AllowSystem),
        definition.Policy(definition.OpRead, requireRole("admin")),
        definition.Policy(definition.OpWrite|definition.OpDelete, requireRole("admin")),
    },

    Hooks: []definition.HookDef{
        {
            Name: "set_assigned_by",
            Ops:  definition.OpCreate,
            When: definition.HookBefore,
            Fn:   setAssignedBy, // sets assigned_by = viewer.ActorID(), assigned_at = now()
        },
        {
            Name: "reload_casbin_after_assignment",
            Ops:  definition.OpCreate | definition.OpDelete,
            When: definition.HookAfter,
            Fn:   reloadCasbinPolicy,
        },
    },
}

var SessionDef = definition.EntityDefinition{
    Name:     "session",
    Label:    "Session",
    Module:   "Platform",
    Table:    "sessions",
    OrgScope: org.ScopeLevelTenant,
    Audited:  false, // sessions are not audited — they are the audit mechanism
    SoftDelete: false,

    Fields: []*definition.FieldDef{
        {Name: "user_id",    Type: definition.FieldTypeLink,     Label: "User",      TargetEntity: "user", Required: true},
        {Name: "token",      Type: definition.FieldTypeData,     Label: "Token",     Sensitive: true},
        {Name: "ip",         Type: definition.FieldTypeData,     Label: "IP Address"},
        {Name: "user_agent", Type: definition.FieldTypeSmallText,Label: "User Agent"},
        {Name: "expires_at", Type: definition.FieldTypeDateTime, Label: "Expires At", Required: true},
        {Name: "last_seen_at",Type: definition.FieldTypeDateTime,Label: "Last Seen"},
        {Name: "revoked_at", Type: definition.FieldTypeDateTime, Label: "Revoked At"},
        {Name: "revoked_by", Type: definition.FieldTypeLink,     Label: "Revoked By", TargetEntity: "user"},
    },

    Policies: []definition.PolicyDef{
        // Sessions are managed only by the IAM service layer. No direct user CRUD.
        definition.Policy(definition.OpAll, definition.AllowSystem),
        // A user may read their own sessions (for "active sessions" UI).
        definition.Policy(definition.OpRead, allowSelf),
        definition.Policy(definition.OpAll, definition.DenyAll),
    },
}
```

### 39.4.2 Session Service Pattern

The IAM service wraps EntityStore with domain-specific logic. The generic CRUD endpoints from EntityDefinition handle admin operations; the IAM service provides the specialised login/logout surface.

```go
// internal/platform/iam/service.go
package iam

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "time"

    "github.com/google/uuid"
    "awo.so/framework/persistence/pgstore"
)

type Service struct {
    sessions pgstore.EntityStore // bound to SessionDef
    users    pgstore.EntityStore // bound to UserDef
    redis    RedisClient
}

// Login authenticates a user and creates a session row + Redis cache entry.
// This is a custom service method, not wired through EntityDefinition's generic API.
func (s *Service) Login(ctx context.Context, req LoginRequest) (*Session, error) {
    // 1. Load user, verify password via bcrypt (UserDef.Policies allows system calls)
    // 2. Optionally verify TOTP
    // 3. Create session row via s.sessions.Create(ctx, sessionRecord)
    //    — The EntityDefinition's AfterHook populates Redis automatically
    // 4. Return opaque token to caller
    _ = req
    return nil, nil // illustrative stub
}

// RevokeAllForTenant is called on tenant suspension.
// Uses BulkUpdate from EntityStore — no loop, single SQL UPDATE.
func (s *Service) RevokeAllForTenant(ctx context.Context, tenantID uuid.UUID) error {
    return s.sessions.BulkUpdate(ctx,
        map[string]any{"tenant_id": tenantID, "revoked_at": nil},
        map[string]any{"revoked_at": time.Now()},
    )
}
```

### 39.4.3 IAM Migration

```sql
-- internal/platform/iam/migrations/20240101000010_create_iam.up.sql

CREATE TABLE users (
    id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     uuid        NOT NULL REFERENCES tenants(id),
    email         text        NOT NULL,
    name          text        NOT NULL,
    password_hash text,                   -- bcrypt; nullable for SSO-only users
    status        text        NOT NULL DEFAULT 'INVITED',
    mfa_secret    text,                   -- encrypted at rest via pgcrypto
    mfa_enabled   boolean     NOT NULL DEFAULT false,
    last_login_at timestamptz,
    org_node_id   uuid        REFERENCES org_nodes(id),
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),

    UNIQUE (tenant_id, email)
);

ALTER TABLE users ENABLE ROW LEVEL SECURITY;
CREATE POLICY users_tenant_isolation ON users
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

CREATE TABLE roles (
    id          uuid    PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid    REFERENCES tenants(id),  -- NULL = system role
    name        text    NOT NULL,
    description text,
    is_system   boolean NOT NULL DEFAULT false,
    created_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, name)
);

ALTER TABLE roles ENABLE ROW LEVEL SECURITY;
CREATE POLICY roles_tenant_isolation ON roles
    USING (tenant_id IS NULL OR tenant_id = current_setting('app.current_tenant_id')::uuid);

CREATE TABLE user_roles (
    user_id     uuid        NOT NULL REFERENCES users(id),
    role_id     uuid        NOT NULL REFERENCES roles(id),
    tenant_id   uuid        NOT NULL REFERENCES tenants(id),
    assigned_by uuid        REFERENCES users(id),
    assigned_at timestamptz NOT NULL DEFAULT now(),
    expires_at  timestamptz,
    PRIMARY KEY (user_id, role_id)
);

ALTER TABLE user_roles ENABLE ROW LEVEL SECURITY;
CREATE POLICY user_roles_tenant_isolation ON user_roles
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

CREATE TABLE sessions (
    id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    uuid        NOT NULL REFERENCES tenants(id),
    user_id      uuid        NOT NULL REFERENCES users(id),
    token        text        NOT NULL UNIQUE,
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

-- Partial index for active session lookups (token lookup on hot path)
CREATE UNIQUE INDEX sessions_token_active_idx ON sessions (token)
    WHERE revoked_at IS NULL AND expires_at > now();
```

---

## 39.5 Feature Flags Module

Feature flags are tenant-scoped but evaluated at multiple granularities: system default → tenant override → user override.

### 39.5.1 EntityDefinitions

```go
// internal/platform/featureflag/definition.go
package featureflag

import (
    "awo.so/framework/definition"
    "awo.so/framework/org"
)

// FeatureFlagDef defines the flag catalogue — what flags exist and their defaults.
// Flags themselves are global (platform-wide catalogue); overrides are per-tenant.
var FeatureFlagDef = definition.EntityDefinition{
    Name:     "feature_flag",
    Label:    "Feature Flag",
    Module:   "Platform",
    Table:    "feature_flags",
    OrgScope: org.ScopeLevelGlobal, // flag catalogue is global
    Audited:  true,

    Fields: []*definition.FieldDef{
        {Name: "key",           Type: definition.FieldTypeData,      Label: "Key",
            Required: true, Description: "Stable dot-namespaced key: 'forecourt.wetstock_alerts'"},
        {Name: "name",          Type: definition.FieldTypeData,      Label: "Display Name", Required: true},
        {Name: "description",   Type: definition.FieldTypeSmallText, Label: "Description"},
        {Name: "type",          Type: definition.FieldTypeSelect,    Label: "Value Type",
            Options: []string{"boolean", "string", "percentage"}, Required: true},
        {Name: "default_value", Type: definition.FieldTypeData,      Label: "Default Value", Required: true},
        {Name: "status",        Type: definition.FieldTypeSelect,    Label: "Status",
            Options: []string{"draft", "active", "deprecated", "removed"}, Required: true},
        {Name: "module",        Type: definition.FieldTypeData,      Label: "Module",
            Description: "Module that owns this flag: 'forecourt', 'hr', 'finance'"},
    },

    Policies: []definition.PolicyDef{
        // Platform admins manage the catalogue.
        definition.Policy(definition.OpAll,  definition.AllowSystem),
        definition.Policy(definition.OpAll,  requireRole("platform_admin")),
        // All authenticated users may read the catalogue (to check flag status).
        definition.Policy(definition.OpRead, allowTenantViewer),
        definition.Policy(definition.OpAll,  definition.DenyAll),
    },
}

// FeatureFlagOverrideDef stores per-tenant and per-user overrides.
var FeatureFlagOverrideDef = definition.EntityDefinition{
    Name:     "feature_flag_override",
    Label:    "Feature Flag Override",
    Module:   "Platform",
    Table:    "feature_flag_overrides",
    OrgScope: org.ScopeLevelTenant, // each override row belongs to a tenant
    Audited:  true,

    Fields: []*definition.FieldDef{
        {Name: "flag_key",   Type: definition.FieldTypeLink,     Label: "Flag",    TargetEntity: "feature_flag", Required: true},
        {Name: "user_id",    Type: definition.FieldTypeLink,     Label: "User",    TargetEntity: "user",
            Description: "nil = tenant-level override; non-nil = user-level override"},
        {Name: "value",      Type: definition.FieldTypeData,     Label: "Value",   Required: true,
            Description: "JSON-encoded value matching the flag's type"},
        {Name: "enabled_at", Type: definition.FieldTypeDateTime, Label: "Enabled At"},
        {Name: "created_by", Type: definition.FieldTypeLink,     Label: "Set By",  TargetEntity: "user"},
    },

    Policies: []definition.PolicyDef{
        definition.Policy(definition.OpAll, definition.AllowSystem),
        // Tenant admins may set tenant-level overrides.
        definition.Policy(definition.OpAll, requireRole("admin")),
        definition.Policy(definition.OpAll, definition.DenyAll),
    },

    Hooks: []definition.HookDef{
        {
            Name: "invalidate_flag_cache",
            Ops:  definition.OpCreate | definition.OpUpdate | definition.OpDelete,
            When: definition.HookAfter,
            Fn:   invalidateFlagCache, // clears Redis key for this tenant's flag snapshot
        },
    },
}
```

### 39.5.2 Evaluation Service

The evaluation service wraps EntityStore but adds Redis caching so every HTTP request does not hit Postgres:

```go
// internal/platform/featureflag/service.go
package featureflag

import (
    "context"
    "encoding/json"

    "awo.so/framework/contextutil"
)

type EvalService struct {
    flags     EntityStore
    overrides EntityStore
    redis     RedisClient
}

// Bool evaluates a boolean feature flag for the current viewer context.
// Resolution order: user override → tenant override → system default.
func (s *EvalService) Bool(ctx context.Context, key string) bool {
    tenantID := contextutil.TenantID(ctx)
    userID   := contextutil.UserID(ctx)

    // Fast path: Redis snapshot built at login time.
    // session.Configuration.Flags is a map[string]string set by SessionService.
    if flags, ok := contextutil.FlagsFromCtx(ctx); ok {
        if val, found := flags[key]; found {
            return val == "true"
        }
    }

    // Slow path: direct DB lookup (used during bootstrap and by admin tooling).
    val := s.evaluate(ctx, key, tenantID, userID)
    return val == "true"
}

// evaluate resolves the flag value through the three-tier cascade.
func (s *EvalService) evaluate(ctx context.Context, key, tenantID, userID string) string {
    // 1. Get system default from flag catalogue
    // 2. Apply tenant-level override if present
    // 3. Apply user-level override if present
    // Implementation omitted — uses s.overrides.List with filter
    _, _, _ = ctx, tenantID, userID
    return "false"
}
```

### 39.5.3 Feature Flag Migration

```sql
-- internal/platform/featureflag/migrations/20240101000020_create_feature_flags.up.sql

CREATE TABLE feature_flags (
    id            uuid  PRIMARY KEY DEFAULT gen_random_uuid(),
    key           text  NOT NULL UNIQUE,
    name          text  NOT NULL,
    description   text,
    type          text  NOT NULL DEFAULT 'boolean',
    default_value text  NOT NULL DEFAULT 'false',
    status        text  NOT NULL DEFAULT 'draft',
    module        text,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

-- No RLS: global table, read by all tenants. Access controlled at app layer.

CREATE TABLE feature_flag_overrides (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid        NOT NULL REFERENCES tenants(id),
    flag_key    text        NOT NULL REFERENCES feature_flags(key),
    user_id     uuid        REFERENCES users(id),  -- NULL = tenant-level
    value       text        NOT NULL,
    enabled_at  timestamptz NOT NULL DEFAULT now(),
    created_by  uuid        REFERENCES users(id),
    created_at  timestamptz NOT NULL DEFAULT now(),

    -- One override per flag per tenant (user_id NULL = tenant level)
    UNIQUE NULLS NOT DISTINCT (tenant_id, flag_key, user_id)
);

ALTER TABLE feature_flag_overrides ENABLE ROW LEVEL SECURITY;
CREATE POLICY ff_overrides_tenant_isolation ON feature_flag_overrides
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);
```

---

## 39.6 Settings Module

Settings store structured configuration at three levels: system → tenant → entity (org node). The EntityDefinition drives the storage and API; the service layer provides the hierarchical resolution.

### 39.6.1 EntityDefinitions

```go
// internal/platform/settings/definition.go
package settings

import (
    "awo.so/framework/definition"
    "awo.so/framework/org"
)

// ConfigKeyDef is the catalogue of known configuration keys.
// Defines type, default, validation, and human description per key.
var ConfigKeyDef = definition.EntityDefinition{
    Name:     "config_key",
    Label:    "Configuration Key",
    Module:   "Platform",
    Table:    "config_keys",
    OrgScope: org.ScopeLevelGlobal, // key catalogue is global
    Audited:  false,

    Fields: []*definition.FieldDef{
        {Name: "module",       Type: definition.FieldTypeData,      Label: "Module",    Required: true},
        {Name: "key",          Type: definition.FieldTypeData,      Label: "Key",       Required: true,
            Description: "Dot-namespaced: 'finance.invoice_prefix'"},
        {Name: "label",        Type: definition.FieldTypeData,      Label: "Label",     Required: true},
        {Name: "description",  Type: definition.FieldTypeSmallText, Label: "Description"},
        {Name: "value_type",   Type: definition.FieldTypeSelect,    Label: "Value Type",
            Options: []string{"string", "integer", "boolean", "json", "date"}},
        {Name: "default_value",Type: definition.FieldTypeData,      Label: "Default Value"},
        {Name: "is_sensitive", Type: definition.FieldTypeBool,      Label: "Sensitive", Default: "false"},
        {Name: "is_tenant_editable", Type: definition.FieldTypeBool, Label: "Tenant Editable", Default: "true"},
    },

    Policies: []definition.PolicyDef{
        definition.Policy(definition.OpAll,  definition.AllowSystem),
        definition.Policy(definition.OpRead, allowTenantViewer),
        definition.Policy(definition.OpAll,  definition.DenyAll),
    },
}

// ConfigValueDef stores the actual values at tenant and org-node levels.
var ConfigValueDef = definition.EntityDefinition{
    Name:     "config_value",
    Label:    "Configuration Value",
    Module:   "Platform",
    Table:    "config_values",
    OrgScope: org.ScopeLevelTenant,
    Audited:  true,
    SoftDelete: false,

    Fields: []*definition.FieldDef{
        {Name: "config_key",  Type: definition.FieldTypeLink,     Label: "Key",       TargetEntity: "config_key", Required: true},
        {Name: "scope",       Type: definition.FieldTypeSelect,   Label: "Scope",
            Options: []string{"system", "tenant", "entity"}, Required: true},
        {Name: "entity_id",   Type: definition.FieldTypeLink,     Label: "Org Node",  TargetEntity: "org_node",
            Description: "Set only when scope=entity"},
        {Name: "value",       Type: definition.FieldTypeData,     Label: "Value",     Required: true},
        {Name: "set_by",      Type: definition.FieldTypeLink,     Label: "Set By",    TargetEntity: "user"},
        {Name: "set_at",      Type: definition.FieldTypeDateTime, Label: "Set At"},
    },

    Policies: []definition.PolicyDef{
        definition.Policy(definition.OpAll,  definition.AllowSystem),
        // Admins manage tenant and entity config.
        definition.Policy(definition.OpAll,  requireRole("admin")),
        // All tenant users may read (needed to drive UI behaviour).
        definition.Policy(definition.OpRead, allowTenantViewer),
        definition.Policy(definition.OpAll,  definition.DenyAll),
    },

    Hooks: []definition.HookDef{
        {
            Name: "validate_value_type",
            Ops:  definition.OpCreate | definition.OpUpdate,
            When: definition.HookBefore,
            Fn:   validateValueMatchesKeyType,
        },
        {
            Name: "check_tenant_editable",
            Ops:  definition.OpCreate | definition.OpUpdate,
            When: definition.HookBefore,
            Fn:   assertKeyIsTenantEditable,
        },
        {
            Name: "invalidate_settings_cache",
            Ops:  definition.OpCreate | definition.OpUpdate | definition.OpDelete,
            When: definition.HookAfter,
            Fn:   invalidateSettingsCache,
        },
    },
}
```

### 39.6.2 Resolution Service

```go
// internal/platform/settings/service.go
package settings

import (
    "context"

    "github.com/google/uuid"
    "awo.so/framework/contextutil"
)

type Service struct {
    values EntityStore
    redis  RedisClient
}

// Get resolves a configuration key with the three-tier cascade.
// entity beats tenant, tenant beats system default.
func (s *Service) Get(ctx context.Context, key string, entityID *uuid.UUID) (string, error) {
    tenantID := contextutil.TenantID(ctx)

    // 1. Try entity-level override
    if entityID != nil {
        if v, err := s.lookup(ctx, key, "entity", tenantID, entityID); err == nil {
            return v, nil
        }
    }

    // 2. Try tenant-level override
    if v, err := s.lookup(ctx, key, "tenant", tenantID, nil); err == nil {
        return v, nil
    }

    // 3. System default from config_keys catalogue
    return s.systemDefault(ctx, key)
}

func (s *Service) lookup(ctx context.Context, key, scope, tenantID string, entityID *uuid.UUID) (string, error) {
    filter := map[string]any{"config_key": key, "scope": scope}
    if entityID != nil {
        filter["entity_id"] = *entityID
    }
    // Uses s.values.List(ctx, ListOptions{Filter: filter, Limit: 1})
    _ = ctx
    return "", nil
}

func (s *Service) systemDefault(ctx context.Context, key string) (string, error) {
    // Reads from config_keys table (ConfigKeyDef.default_value)
    _ = ctx
    return "", nil
}
```

### 39.6.3 Settings Migration

```sql
-- internal/platform/settings/migrations/20240101000030_create_settings.up.sql

CREATE TABLE config_keys (
    id                  uuid    PRIMARY KEY DEFAULT gen_random_uuid(),
    module              text    NOT NULL,
    key                 text    NOT NULL UNIQUE,
    label               text    NOT NULL,
    description         text,
    value_type          text    NOT NULL DEFAULT 'string',
    default_value       text,
    is_sensitive        boolean NOT NULL DEFAULT false,
    is_tenant_editable  boolean NOT NULL DEFAULT true,
    created_at          timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE config_values (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid        NOT NULL REFERENCES tenants(id),
    config_key  text        NOT NULL REFERENCES config_keys(key),
    scope       text        NOT NULL DEFAULT 'tenant',
    entity_id   uuid        REFERENCES org_nodes(id),
    value       text        NOT NULL,
    set_by      uuid        REFERENCES users(id),
    set_at      timestamptz NOT NULL DEFAULT now(),
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),

    -- One value per key per scope per entity within a tenant.
    UNIQUE NULLS NOT DISTINCT (tenant_id, config_key, scope, entity_id)
);

ALTER TABLE config_values ENABLE ROW LEVEL SECURITY;
CREATE POLICY config_values_tenant_isolation ON config_values
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

CREATE TRIGGER set_updated_at
    BEFORE UPDATE ON config_values
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

---

## 39.7 Audit Log Module

The Audit Log is append-only. The EntityDefinition declares no `OpCreate` write policy for users — all writes come from the framework's built-in audit hook and the PostgreSQL trigger.

### 39.7.1 EntityDefinition

```go
// internal/platform/audit/definition.go
package audit

import (
    "awo.so/framework/definition"
    "awo.so/framework/org"
)

var AuditLogDef = definition.EntityDefinition{
    Name:     "audit_log",
    Label:    "Audit Log",
    Module:   "Platform",
    Table:    "audit_log",
    OrgScope: org.ScopeLevelTenant,
    Audited:  false, // audit log does not audit itself
    SoftDelete: false,

    Fields: []*definition.FieldDef{
        {Name: "entity_name",    Type: definition.FieldTypeData,      Label: "Entity"},
        {Name: "record_id",      Type: definition.FieldTypeUUID,      Label: "Record ID"},
        {Name: "operation",      Type: definition.FieldTypeSelect,    Label: "Operation",
            Options: []string{"CREATE", "UPDATE", "DELETE"}},
        {Name: "user_id",        Type: definition.FieldTypeLink,      Label: "Actor",      TargetEntity: "user"},
        {Name: "workflow_id",    Type: definition.FieldTypeData,      Label: "Workflow ID"},
        {Name: "request_id",     Type: definition.FieldTypeData,      Label: "Request ID"},
        {Name: "ip",             Type: definition.FieldTypeData,      Label: "IP Address"},
        {Name: "timestamp",      Type: definition.FieldTypeDateTime,  Label: "Timestamp"},
        {Name: "previous_value", Type: definition.FieldTypeJSON,      Label: "Before"},
        {Name: "new_value",      Type: definition.FieldTypeJSON,      Label: "After"},
    },

    Policies: []definition.PolicyDef{
        // System may write (the framework audit hook is a system caller).
        definition.Policy(definition.OpCreate, definition.AllowSystem),
        // No user can create, update, or delete audit log entries.
        definition.Policy(definition.OpWrite|definition.OpDelete, definition.DenyAll),
        // Auditors and admins may read.
        definition.Policy(definition.OpRead, requireRole("admin")),
        definition.Policy(definition.OpRead, requireRole("auditor")),
        definition.Policy(definition.OpAll, definition.DenyAll),
    },
}
```

### 39.7.2 Audit Log Migration

```sql
-- internal/platform/audit/migrations/20240101000040_create_audit_log.up.sql

-- Monthly-partitioned audit log for 7-year retention (Kenya Companies Act 2015)
CREATE TABLE audit_log (
    id             uuid        NOT NULL DEFAULT gen_random_uuid(),
    tenant_id      uuid        NOT NULL REFERENCES tenants(id),
    entity_name    text        NOT NULL,
    record_id      uuid        NOT NULL,
    operation      text        NOT NULL,
    user_id        uuid,       -- NULL for system/workflow operations
    workflow_id    text,
    request_id     text,
    ip             text,
    timestamp      timestamptz NOT NULL DEFAULT now(),
    previous_value jsonb,
    new_value      jsonb,

    PRIMARY KEY (id, timestamp)
) PARTITION BY RANGE (timestamp);

-- Bootstrap partition; scheduled job creates future partitions monthly.
CREATE TABLE audit_log_2024_01 PARTITION OF audit_log
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

ALTER TABLE audit_log ENABLE ROW LEVEL SECURITY;
CREATE POLICY audit_log_tenant_isolation ON audit_log
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

-- Indexes: most queries filter by entity+record or by user+time range.
CREATE INDEX audit_log_entity_record_idx ON audit_log (tenant_id, entity_name, record_id, timestamp DESC);
CREATE INDEX audit_log_user_time_idx     ON audit_log (tenant_id, user_id, timestamp DESC);

-- Trigger that captures every INSERT/UPDATE/DELETE on any auditable table.
-- The framework wires this to each EntityDefinition where Audited=true.
CREATE OR REPLACE FUNCTION audit_log_trigger_fn()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO audit_log (
        id, tenant_id, entity_name, record_id, operation,
        timestamp, previous_value, new_value
    ) VALUES (
        gen_random_uuid(),
        COALESCE(NEW.tenant_id, OLD.tenant_id),
        TG_TABLE_NAME,
        COALESCE(NEW.id, OLD.id),
        TG_OP,
        now(),
        CASE WHEN TG_OP != 'INSERT' THEN to_jsonb(OLD) ELSE NULL END,
        CASE WHEN TG_OP != 'DELETE' THEN to_jsonb(NEW) ELSE NULL END
    );
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

---

## 39.8 Metadata / Custom Fields Module

Custom fields let tenant administrators extend any entity at runtime without schema migrations. `CustomFieldDef` records are tenant-scoped metadata; the actual values are stored as JSONB on each entity row in a `custom_fields` column.

### 39.8.1 EntityDefinitions

```go
// internal/platform/metadata/definition.go
package metadata

import (
    "awo.so/framework/definition"
    "awo.so/framework/org"
)

// CustomFieldDefDef defines a custom field that a tenant admin has added to an entity.
var CustomFieldDefDef = definition.EntityDefinition{
    Name:     "custom_field_def",
    Label:    "Custom Field",
    Module:   "Platform",
    Table:    "custom_field_defs",
    OrgScope: org.ScopeLevelTenant,
    Audited:  true,
    SoftDelete: true, // soft-delete preserves historical data stored in the field

    Fields: []*definition.FieldDef{
        {Name: "entity_name",  Type: definition.FieldTypeData,   Label: "Entity",     Required: true,
            Description: "EntityDefinition.Name: 'customer', 'sales_order'"},
        {Name: "field_name",   Type: definition.FieldTypeData,   Label: "Field Name", Required: true,
            Description: "snake_case key used in JSONB: 'vat_number', 'nssf_number'"},
        {Name: "label",        Type: definition.FieldTypeData,   Label: "Label",      Required: true},
        {Name: "field_type",   Type: definition.FieldTypeSelect, Label: "Field Type",
            Options: []string{"text", "integer", "boolean", "date", "select", "link"}, Required: true},
        {Name: "options",      Type: definition.FieldTypeJSON,   Label: "Options",
            Description: "For select fields: [{value, label}, ...]"},
        {Name: "target_entity",Type: definition.FieldTypeData,   Label: "Target Entity",
            Description: "For link fields: EntityDefinition.Name of the linked entity"},
        {Name: "required",     Type: definition.FieldTypeBool,   Label: "Required",   Default: "false"},
        {Name: "sort_order",   Type: definition.FieldTypeInt,    Label: "Sort Order", Default: "0"},
        {Name: "section",      Type: definition.FieldTypeData,   Label: "Form Section",
            Description: "Groups related custom fields in SDUI forms"},
    },

    Policies: []definition.PolicyDef{
        definition.Policy(definition.OpAll,  definition.AllowSystem),
        // Only tenant admins may define custom fields.
        definition.Policy(definition.OpAll,  requireRole("admin")),
        // All authenticated users can read the schema (needed to render forms).
        definition.Policy(definition.OpRead, allowTenantViewer),
        definition.Policy(definition.OpAll,  definition.DenyAll),
    },

    Hooks: []definition.HookDef{
        {
            Name: "validate_field_name",
            Ops:  definition.OpCreate,
            When: definition.HookBefore,
            Fn:   validateFieldName, // rejects reserved names, enforces snake_case
        },
        {
            Name: "validate_target_entity_exists",
            Ops:  definition.OpCreate,
            When: definition.HookBefore,
            Fn:   validateLinkTargetExists, // calls definition.Lookup(targetEntity)
        },
        {
            Name: "invalidate_sdui_schema_cache",
            Ops:  definition.OpCreate | definition.OpUpdate | definition.OpDelete,
            When: definition.HookAfter,
            Fn:   invalidateSDUISchemaCache, // forces SDUI to regenerate form schemas
        },
    },
}
```

### 39.8.2 How Business Modules Receive Custom Fields

Entity rows store custom field values in a `custom_fields jsonb` column. The SDUI layer merges framework-defined fields with tenant-defined custom fields at render time:

```go
// Called by SDUI handler when building form schemas for any entity.
func (s *SDUIService) formFields(ctx context.Context, def *definition.EntityDefinition) ([]*FieldSchema, error) {
    // 1. Start with the entity's declared Fields from its EntityDefinition.
    fields := toFieldSchemas(def.Fields)

    // 2. Load tenant's CustomFieldDef rows for this entity.
    customDefs, err := s.customFields.List(ctx, ListOptions{
        Filter: map[string]any{
            "entity_name": def.Name,
            "deleted_at":  nil,
        },
        OrderBy: "sort_order ASC",
    })
    if err != nil {
        return nil, err
    }

    // 3. Append custom fields as additional AMIS controls,
    //    sourced from the record's custom_fields JSONB column.
    for _, cd := range customDefs.Items {
        fields = append(fields, customFieldToSchema(cd))
    }
    return fields, nil
}
```

### 39.8.3 Metadata Migration

```sql
-- internal/platform/metadata/migrations/20240101000050_create_custom_fields.up.sql

CREATE TABLE custom_field_defs (
    id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    uuid        NOT NULL REFERENCES tenants(id),
    entity_name  text        NOT NULL,
    field_name   text        NOT NULL,
    label        text        NOT NULL,
    field_type   text        NOT NULL,
    options      jsonb,
    target_entity text,
    required     boolean     NOT NULL DEFAULT false,
    sort_order   integer     NOT NULL DEFAULT 0,
    section      text,
    deleted_at   timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),

    UNIQUE (tenant_id, entity_name, field_name)
);

ALTER TABLE custom_field_defs ENABLE ROW LEVEL SECURITY;
CREATE POLICY custom_field_defs_tenant_isolation ON custom_field_defs
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

CREATE TRIGGER set_updated_at
    BEFORE UPDATE ON custom_field_defs
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Each entity table that supports custom fields adds this column via a separate migration.
-- Example for a future Customer entity:
--
-- ALTER TABLE customers ADD COLUMN IF NOT EXISTS custom_fields jsonb NOT NULL DEFAULT '{}';
-- CREATE INDEX customers_custom_fields_gin ON customers USING gin(custom_fields);
```

---

## 39.9 Plugin / Module Registry

A module registry lets operators discover which business modules are installed on a given deployment and which modules a tenant has activated. This is the runtime counterpart to Go's import graph.

### 39.9.1 EntityDefinitions

```go
// internal/platform/registry/definition.go
package registry

import (
    "awo.so/framework/definition"
    "awo.so/framework/org"
)

// ModuleDef describes one installed business module (finance, hr, forecourt, etc.).
// Populated at startup by each module's init() function calling registry.Register().
var ModuleDef = definition.EntityDefinition{
    Name:     "module",
    Label:    "Module",
    Module:   "Platform",
    Table:    "modules",
    OrgScope: org.ScopeLevelGlobal, // module catalogue is global
    Audited:  false,

    Fields: []*definition.FieldDef{
        {Name: "name",        Type: definition.FieldTypeData,      Label: "Name",        Required: true},
        {Name: "label",       Type: definition.FieldTypeData,      Label: "Display Name", Required: true},
        {Name: "version",     Type: definition.FieldTypeData,      Label: "Version",      Required: true},
        {Name: "description", Type: definition.FieldTypeSmallText, Label: "Description"},
        {Name: "repo_url",    Type: definition.FieldTypeData,      Label: "Repository URL"},
        {Name: "status",      Type: definition.FieldTypeSelect,    Label: "Status",
            Options: []string{"stable", "beta", "deprecated"}},
    },

    Policies: []definition.PolicyDef{
        definition.Policy(definition.OpAll,  definition.AllowSystem),
        // Module catalogue is read-only for non-platform callers.
        definition.Policy(definition.OpRead, allowTenantViewer),
        definition.Policy(definition.OpAll,  definition.DenyAll),
    },
}

// TenantModuleDef records which modules a specific tenant has activated.
var TenantModuleDef = definition.EntityDefinition{
    Name:     "tenant_module",
    Label:    "Tenant Module Activation",
    Module:   "Platform",
    Table:    "tenant_modules",
    OrgScope: org.ScopeLevelTenant,
    Audited:  true,

    Fields: []*definition.FieldDef{
        {Name: "module_name",  Type: definition.FieldTypeLink,     Label: "Module",   TargetEntity: "module", Required: true},
        {Name: "status",       Type: definition.FieldTypeSelect,   Label: "Status",
            Options: []string{"active", "suspended", "trial"}, Required: true},
        {Name: "activated_at", Type: definition.FieldTypeDateTime, Label: "Activated At"},
        {Name: "trial_ends_at",Type: definition.FieldTypeDateTime, Label: "Trial Ends"},
        {Name: "config",       Type: definition.FieldTypeJSON,     Label: "Module Config",
            Description: "Module-specific configuration JSONB blob"},
    },

    Policies: []definition.PolicyDef{
        definition.Policy(definition.OpAll,  definition.AllowSystem),
        // Platform admins activate modules for tenants.
        definition.Policy(definition.OpAll,  requireRole("platform_admin")),
        // Tenant admins and all users read active modules (needed for UI navigation).
        definition.Policy(definition.OpRead, allowTenantViewer),
        definition.Policy(definition.OpAll,  definition.DenyAll),
    },

    Hooks: []definition.HookDef{
        {
            Name: "reload_module_nav_on_activation",
            Ops:  definition.OpCreate | definition.OpUpdate,
            When: definition.HookAfter,
            Fn:   reloadModuleNavigation, // invalidates SDUI navigation cache for tenant
        },
    },
}
```

### 39.9.2 Module Self-Registration

Business modules register into the framework catalogue in their own `init()`:

```go
// In a business module repo — e.g. awo.so/module/finance
package finance

import (
    "awo.so/framework/definition"
    "awo.so/platform/registry"
)

func init() {
    // Register all finance EntityDefinitions with the framework.
    definition.Register(&AccountDef)
    definition.Register(&JournalEntryDef)
    definition.Register(&InvoiceDef)
    // ... etc.

    // Register the module itself with the platform registry.
    registry.RegisterModule(registry.ModuleInfo{
        Name:        "finance",
        Label:       "Finance",
        Version:     "1.0.0",
        Description: "General ledger, accounts payable/receivable, invoicing",
        RepoURL:     "https://github.com/awoerp/module-finance",
    })
}
```

### 39.9.3 Module Registry Migration

```sql
-- internal/platform/registry/migrations/20240101000060_create_module_registry.up.sql

CREATE TABLE modules (
    id          uuid    PRIMARY KEY DEFAULT gen_random_uuid(),
    name        text    NOT NULL UNIQUE,
    label       text    NOT NULL,
    version     text    NOT NULL,
    description text,
    repo_url    text,
    status      text    NOT NULL DEFAULT 'stable',
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE tenant_modules (
    id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    uuid        NOT NULL REFERENCES tenants(id),
    module_name  text        NOT NULL REFERENCES modules(name),
    status       text        NOT NULL DEFAULT 'active',
    activated_at timestamptz NOT NULL DEFAULT now(),
    trial_ends_at timestamptz,
    config       jsonb       NOT NULL DEFAULT '{}',
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),

    UNIQUE (tenant_id, module_name)
);

ALTER TABLE tenant_modules ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_modules_isolation ON tenant_modules
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);
```

---

## 39.10 Summary — Platform Module vs Business Module

The table below shows that platform and business modules use **identical** framework patterns. The only difference is where the Go source lives and who owns the EntityDefinitions.

| Concern | Platform Module | Business Module |
|---|---|---|
| **Go package** | `internal/platform/<module>/` | Separate repo: `awo.so/module/<name>/` |
| **EntityDefinition** | `definition.EntityDefinition{}` | `definition.EntityDefinition{}` — same struct |
| **Registration** | `definition.Register(&Def)` in `init()` | `definition.Register(&Def)` in `init()` — identical |
| **OrgScope** | Global (catalogue) or Tenant | Usually Tenant or Unit |
| **Policies** | Platform roles + system | Domain roles + system |
| **Hooks** | Cache invalidation, Casbin reload | Business rule validation, workflow triggers |
| **Migrations** | `internal/platform/<module>/migrations/` | `migrations/` in module repo |
| **Service layer** | Thin wrapper over EntityStore | Thin wrapper over EntityStore |
| **HTTP API** | Auto-generated by framework + custom endpoints | Auto-generated by framework + custom endpoints |
| **SDUI** | Auto-generated AMIS pages | Auto-generated AMIS pages |
| **Audit** | `Audited: true` on sensitive entities | `Audited: true` on financial/HR entities |

The framework cannot tell, at runtime, whether an EntityDefinition came from the `internal/platform` package or from `awo.so/module/finance`. Both are identical Go values stored in the same in-memory registry.

---

## 39.11 Startup Wire-Up

Platform modules are wired into the server at startup before business modules, because business modules may reference platform entity types (User, OrgNode, Tenant) in their EdgeDefs.

```go
// cmd/server/main.go
package main

import (
    "awo.so/framework/bootstrap"

    // Platform modules — init() registers their EntityDefinitions.
    _ "awo.so/internal/platform/tenant"
    _ "awo.so/internal/platform/iam"
    _ "awo.so/internal/platform/featureflag"
    _ "awo.so/internal/platform/settings"
    _ "awo.so/internal/platform/audit"
    _ "awo.so/internal/platform/metadata"
    _ "awo.so/internal/platform/registry"

    // Business modules — registered after platform modules.
    _ "awo.so/module/finance"
    _ "awo.so/module/hr"
    _ "awo.so/module/crm"
)

func main() {
    app := fiber.New()
    bootstrap.Mount(app, bootstrap.Options{
        DSN:       os.Getenv("DATABASE_URL"),
        RedisAddr: os.Getenv("REDIS_ADDR"),
    })
    log.Fatal(app.Listen(":8080"))
}
```

`bootstrap.Mount` reads `definition.All()` to get every registered EntityDefinition and mounts CRUD routes, SDUI endpoints, and migration sources for all of them in a single pass. No per-module bootstrap call needed.
