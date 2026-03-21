# Identity, Authentication & Authorization — Complete Technical Guide

> **Package:** `internal/platform`  
> **Audience:** Platform engineers, module developers, security reviewers, and finance/ops domain experts.  
> **Scope:** Everything from a user typing a password to a DB query being allowed or blocked — including directory layout, data layer conventions, the Module/Resource/Action registry, feature flags, tenant settings, UI navigation generation, form architecture, resource hierarchy, and audit strategy.

---

## Table of Contents

1. [Module Philosophy & Scope](#1-module-philosophy--scope)
2. [Directory Structure & Code Layout](#2-directory-structure--code-layout)
3. [The Module/Resource/Action Registry](#3-the-moduleresourceaction-registry)
4. [Feature Flags — Module & Resource Toggles](#4-feature-flags--module--resource-toggles)
5. [Tenant Settings — Behavioural Configuration](#5-tenant-settings--behavioural-configuration)
6. [Authentication (AuthN) — Who Are You?](#6-authentication-authn--who-are-you)
7. [Authorization (AuthZ) — What Can You Do?](#7-authorization-authz--what-can-you-do)
8. [Database Architecture](#8-database-architecture)
9. [Resource Hierarchy & Entity Model](#9-resource-hierarchy--entity-model)
10. [The Session — Pre-Computed Everything](#10-the-session--pre-computed-everything)
11. [UI Structure — How Configuration Becomes Interface](#11-ui-structure--how-configuration-becomes-interface)
12. [HTTP Middleware Chain & API Layer](#12-http-middleware-chain--api-layer)
13. [Domain Isolation](#13-domain-isolation)
14. [IAM Services, Resources & Actions](#14-iam-services-resources--actions)
15. [Cross-Module Integration](#15-cross-module-integration)
16. [Audit Trail](#16-audit-trail)
17. [Security Threat Model](#17-security-threat-model)
18. [Performance & Caching](#18-performance--caching)
19. [Common Business Scenarios](#19-common-business-scenarios)
20. [Testing Strategy](#20-testing-strategy)
21. [API Reference](#21-api-reference)
22. [Troubleshooting](#22-troubleshooting)

---

## 1. Module Philosophy & Scope

### 1.1 The Three Questions Every Request Answers

Every HTTP request that enters Awo ERP answers three questions before touching business data:

1. **Authentication (AuthN): Who are you?** Verify the claimed identity. Produce a trusted session.
2. **Authorization (AuthZ): What can you do?** Given your identity, decide allow or deny for this operation.
3. **Configuration: How should this behave?** Given your tenant's flags and settings, which features are on and how do they operate?

These are distinct concerns with distinct mechanisms but they all resolve at the same moment — login — and travel together in the session object. After login, every answer to all three questions is an in-memory lookup with no database hit.

### 1.2 The Unified Key Namespace

The most important architectural decision in this module is that **permissions, feature flags, and settings all share the same dot-notation key hierarchy**:

```
{module}.{resource}.{action}     →  permission key (leaf level)
{module}.{resource}              →  resource-level flag or setting key
{module}                         →  module-level flag key

finance                                      module flag: is Finance on?
finance.transactions                         resource flag: are Transactions enabled?
finance.transactions.approval_workflow       setting: is approval workflow active?
finance.transactions.approval_threshold      setting: approval amount threshold
finance.transactions.approve                 permission: can this user approve?
```

This means the Module/Resource/Action (MRA) tables — which already exist and have no `tenant_id` because they are system-wide — are the single source of truth that anchors all three systems. You do not write new key strings in Go code for flags or settings; you derive them from the MRA slugs. Adding a new module seeds its flag definitions automatically. Adding a new action seeds its permission key automatically.

The distinction between flags and settings is intentional even though they share the namespace. A flag is binary — the feature exists or does not for this tenant. A setting is a value — how the feature behaves when it is on. They have different resolution logic, different UI controls, and different permission requirements to change. The key namespace is shared; the tables are not.

### 1.3 The Trust Chain

```
Browser / API Client
      │
      ▼
Fiber HTTP Server
      ├── Recovery, Logger, RateLimit, CORS, SecurityHeaders
      │
      ├── Authenticate()              AuthN: session cookie or Bearer token
      │     → loads ResolvedSession into context
      │     → ResolvedSession carries: identity + permissions + flags + settings + entity scope
      │
      ├── SetDBPool()                 pool selection from session.UserType
      │     platform user  → admin_role (BYPASSRLS)
      │     all others     → application_role (RLS active, per-connection settings)
      │
      ├── RequireTenantContext()      validates X-Tenant-ID header
      │
      ├── RequirePermission()         O(1) map lookup from session — zero DB hit
      ├── RequireFlag()               O(1) map lookup from session — zero DB hit
      │
      └── Handler                    business logic; reads only from session and request
```

### 1.4 What This Module Owns

**Owns:** User lifecycle, AuthN (login/logout/MFA/OAuth/passwords/API keys), session management, RBAC with Casbin, the MRA registry, feature flag catalogue and tenant flag values, tenant settings catalogue and values, user preferences, UI navigation generation, entity hierarchy resolution.

**Does not own:** HTTP routing, DB connection pools, notification delivery, audit log persistence, tenant provisioning (responds to it via events), business module schemas.

---

## 2. Directory Structure & Code Layout

### 2.1 Repository Layout

```

├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── auth_handler.go        # POST /auth/login, /logout, /forgot-password
│   │   │   ├── user_handler.go        # CRUD /api/v1/users
│   │   │   ├── role_handler.go        # CRUD /api/v1/roles
│   │   │   ├── flag_handler.go        # GET/PATCH /api/v1/settings/flags
│   │   │   ├── settings_handler.go    # GET/PATCH /api/v1/settings/{module}
│   │   │   └── schema_handler.go      # GET /schema/boot, /schema/*
│   │   └── middleware/
│   │       ├── auth.go                # Authenticate — session or Bearer
│   │       ├── tenant.go              # ResolveTenant — X-Tenant-ID or subdomain
│   │       ├── permission.go          # RequirePermission, RequireFlag
│   │       ├── db_pool.go             # SetDBPool
│   │       ├── audit.go               # AuditWrap
│   │       └── context.go             # ContextSession, ContextTenantID, helpers
│   │
│   └── platform/
│       ├── service.go                 # Facade: Platform struct — single import for all modules
│       ├── iam.go                     # IAM service composition
│       ├── flags.go                   # FlagService — flag catalogue + tenant values
│       ├── settings.go                # SettingService — setting catalogue + tenant values
│       ├── boot.go                    # BootService — builds app shell from MRA+flags+perms
│       │
│       ├── domain/
│       │   ├── user.go                # User, UserType, UserCreateParams
│       │   ├── session.go             # ResolvedSession, SessionConfiguration
│       │   ├── role.go                # Role, RoleAssignment, AssignOptions
│       │   ├── permission.go          # Permission, PermissionKey
│       │   ├── policy.go              # Policy, Effect
│       │   ├── module.go              # Module, Resource, Action (MRA domain types)
│       │   ├── flag.go                # FlagDefinition, TenantFlag
│       │   ├── setting.go             # SettingDefinition, TenantSetting, UserPreference
│       │   ├── entity.go              # Entity, EntityScope, EntityType
│       │   └── events.go              # All platform domain events
│       │
│       └── repo/
│           ├── user_repo.go           # interface + impl (sqlc + cache)
│           ├── session_repo.go
│           ├── role_repo.go
│           ├── module_repo.go         # MRA queries — list enabled, nav data
│           ├── flag_repo.go           # flag definitions + tenant values
│           ├── setting_repo.go        # setting definitions + tenant values
│           ├── entity_repo.go
│           └── casbin_adapter.go      # Casbin ↔ PostgreSQL bridge
│
├── db/
│   ├── migration/                     # Sequential migration files
│   ├── queries/                       # SQL for sqlc
│   │   ├── users.sql
│   │   ├── modules.sql                # MRA + nav queries
│   │   ├── flags.sql
│   │   ├── settings.sql
│   │   └── ...
│   └── sqlc/                          # Generated code (package: db)
└── ...
```

Unit tests live next to every file they test. No separate test directories.

### 2.2 The Platform Facade

Every business module receives one `*platform.Platform`. It never imports individual platform files:

```go
// internal/platform/service.go

type Platform struct {
    IAM      *IAMService
    Tenant   *TenantService
    Flags    *FlagService
    Settings *SettingService
    Boot     *BootService
    Audit    *AuditService
    Notify   *NotifyService
}

// Usage in any business module:
type FinanceService struct {
    platform *platform.Platform
    repo     FinanceRepository
}

func (s *FinanceService) PostTransaction(ctx context.Context,
    params PostTransactionParams) (*Transaction, error) {

    session := domain.SessionFromContext(ctx)
    if !session.Can("finance.transactions", "post") {
        return nil, domain.ErrForbidden
    }
    // ...
}
```

### 2.3 Repository Layer Convention

```
Service  →  Repository interface  →  Repository implementation  →  sqlc generated code
                                         (owns cache logic)
```

Services call repository methods. Repositories use sqlc internally and own all cache logic. Services never write SQL or set cache keys.

```go
// internal/platform/repo/user_repo_impl.go

type userRepoImpl struct {
    q     *db.Queries    // sqlc — only place raw queries appear
    cache *redis.Client  // cache: the repo's responsibility, not the service's
}

func (r *userRepoImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
    cacheKey := "user:" + id.String()
    if raw, err := r.cache.Get(ctx, cacheKey).Bytes(); err == nil {
        var u domain.User
        json.Unmarshal(raw, &u)
        return &u, nil
    }
    row, err := r.q.GetUserByID(ctx, id) // sqlc call
    if err == pgx.ErrNoRows { return nil, domain.ErrNotFound }
    user := mapRowToUser(row)
    data, _ := json.Marshal(user)
    r.cache.Set(ctx, cacheKey, data, 5*time.Minute)
    return user, nil
}
```

### 2.4 Function Parameter Convention

All service and repository methods use a single params struct:

```go
// DO — struct parameter; adding a field never breaks existing callers
func (s *IAMService) CreateUser(ctx context.Context,
    params domain.UserCreateParams) (*domain.User, error)

// DON'T — breaks call sites when new fields are needed
func (s *IAMService) CreateUser(ctx context.Context,
    email, displayName string, userType domain.UserType) (*domain.User, error)
```

### 2.5 Context Carries Tenant and User Identity

Middleware injects `tenant_id`, `user_id`, and `entity_id` via `context.WithValue`. Services read them without explicit parameters:

```go
// Set by middleware
ctx = context.WithValue(ctx, contextKeyTenantID, tenantID)
ctx = context.WithValue(ctx, contextKeyUserID,   userID)
ctx = context.WithValue(ctx, contextKeyEntityID, entityID)

// Typed accessors (domain package)
func TenantIDFromContext(ctx context.Context) (uuid.UUID, bool)
func UserIDFromContext(ctx context.Context)   (uuid.UUID, bool)
func EntityIDFromContext(ctx context.Context) (uuid.UUID, bool)

// Service reads without explicit parameter
func (s *IAMService) ListUsers(ctx context.Context,
    params domain.UserListParams) ([]*domain.User, int, error) {
    tenantID, ok := domain.TenantIDFromContext(ctx)
    if !ok { return nil, 0, domain.ErrMissingTenantContext }
    params.TenantID = tenantID
    return s.repo.List(ctx, params)
}
```

| Data | How to pass |
|---|---|
| `tenant_id` for query scoping | Context — set by auth middleware |
| `user_id` of the *acting* user (for audit) | Context — set by auth middleware |
| `user_id` of the *target* user (e.g., assign role to this user) | Explicit in params struct |
| Business parameters | Explicit params struct |

### 2.6 Tenant Identification

Tenants have three identifiers. UUID is canonical. Slug and subdomain resolve to UUID at the edge:

```go
func ResolveTenant(tenantRepo TenantRepository) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Path 1: X-Tenant-ID header (API clients, frontend)
        if headerID := c.Get("X-Tenant-ID"); headerID != "" {
            id, err := uuid.Parse(headerID)
            if err != nil { return c.Status(400).JSON(response.Err("invalid X-Tenant-ID")) }
            return injectAndContinue(c, id)
        }
        // Path 2: Subdomain (browser acme.awoerp.com)
        if t, err := tenantRepo.GetBySubdomain(c.Context(), c.Hostname()); err == nil {
            return injectAndContinue(c, t.ID)
        }
        // Path 3: Platform — no tenant required
        if s := middleware.ContextSession(c); s != nil && s.IsPlatform() {
            return c.Next()
        }
        return c.Status(403).JSON(response.Err("X-Tenant-ID header is required"))
    }
}
```

---

## 3. The Module/Resource/Action Registry

### 3.1 These Tables Are the System's Vocabulary

The MRA tables describe every piece of functionality that exists in Awo ERP. They have no `tenant_id` because they are not tenant data — they are system data, shared across all tenants in the same way that code is shared. Every module, every resource, every action that can possibly be performed is registered here.

```sql
CREATE TABLE modules (
  id         uuid  PRIMARY KEY DEFAULT gen_random_uuid(),
  slug       text  UNIQUE NOT NULL,   -- 'finance', 'selling', 'forecourt'
  label      text  NOT NULL,          -- 'Finance', 'Sales', 'Forecourt'
  icon       text,                    -- 'fa fa-calculator'
  nav_order  int   NOT NULL DEFAULT 999,
  is_active  bool  NOT NULL DEFAULT true
);

CREATE TABLE resources (
  id         uuid  PRIMARY KEY DEFAULT gen_random_uuid(),
  module_id  uuid  NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
  slug       text  NOT NULL,          -- 'transactions', 'accounts', 'pumps'
  label      text  NOT NULL,          -- 'Transactions', 'Accounts', 'Pump Control'
  nav_url    text,                    -- '/finance/transactions'
  nav_order  int   NOT NULL DEFAULT 999,
  UNIQUE (module_id, slug)
);

CREATE TABLE actions (
  id          uuid  PRIMARY KEY DEFAULT gen_random_uuid(),
  resource_id uuid  NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
  slug        text  NOT NULL,         -- 'read', 'create', 'approve', 'post', 'void'
  label       text  NOT NULL,         -- 'View', 'Create', 'Approve'
  http_method text,                   -- 'GET', 'POST', 'PATCH', 'DELETE'
  UNIQUE (resource_id, slug)
);
```

`GRANT SELECT ON modules, resources, actions TO application_role;` — all tenants can read these tables. There is no RLS on them because they contain no tenant data.

### 3.2 The Key Derivation Rule

Every permission key, flag key, and setting key is derived from the MRA slugs. This is a rule, not a convention:

```
permission key  = {module.slug}.{resource.slug}.{action.slug}
resource flag   = {module.slug}.{resource.slug}
module flag     = {module.slug}
setting key     = {module.slug}.{resource.slug?}.{setting_name}
```

No key is written as a free-form string in application code. Keys are computed from the MRA tables. The only strings developers write are the slugs themselves when inserting rows.

### 3.3 Auto-Seeding via Triggers

When a module row is inserted, a trigger automatically seeds its feature flag definition. The same applies to resources and permissions:

```sql
-- Auto-seed module flag definition when a module is created
CREATE OR REPLACE FUNCTION seed_module_flag() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO feature_flag_definitions (module_id, flag_key, label, default_value, is_system)
    VALUES (NEW.id,
            NEW.slug,
            NEW.label || ' module',
            false,
            false)
    ON CONFLICT (flag_key) DO NOTHING;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_seed_module_flag
    AFTER INSERT ON modules
    FOR EACH ROW EXECUTE FUNCTION seed_module_flag();

-- Auto-seed resource flag definition when a resource is created
CREATE OR REPLACE FUNCTION seed_resource_flag() RETURNS TRIGGER AS $$
DECLARE v_module_slug text;
BEGIN
    SELECT slug INTO v_module_slug FROM modules WHERE id = NEW.module_id;
    INSERT INTO feature_flag_definitions (module_id, resource_id, flag_key, label, default_value)
    VALUES (NEW.module_id,
            NEW.id,
            v_module_slug || '.' || NEW.slug,
            NEW.label,
            true)    -- resources default to enabled once their module is enabled
    ON CONFLICT (flag_key) DO NOTHING;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_seed_resource_flag
    AFTER INSERT ON resources
    FOR EACH ROW EXECUTE FUNCTION seed_resource_flag();
```

The `awo-gen` CLI tool inserts into these tables. The triggers fire automatically. By the time a developer runs a migration, both the MRA rows and their flag definitions exist.

### 3.4 Adding a New Module — What Changes

```sql
-- 1. Insert the module (trigger seeds module flag definition automatically)
INSERT INTO modules (slug, label, icon, nav_order)
VALUES ('forecourt', 'Forecourt', 'fa fa-gas-pump', 600);

-- 2. Insert resources (each triggers resource flag seed)
INSERT INTO resources (module_id, slug, label, nav_url, nav_order)
SELECT id, 'pumps', 'Pump Control', '/forecourt/pumps', 10
FROM modules WHERE slug = 'forecourt';

-- 3. Insert actions (used by awo-gen to seed permissions)
INSERT INTO actions (resource_id, slug, label, http_method)
SELECT r.id, 'read',      'View Pumps',     'GET'
FROM resources r JOIN modules m ON r.module_id = m.id
WHERE m.slug = 'forecourt' AND r.slug = 'pumps'
UNION ALL
SELECT r.id, 'authorise', 'Authorise Pump', 'POST'
FROM resources r JOIN modules m ON r.module_id = m.id
WHERE m.slug = 'forecourt' AND r.slug = 'pumps';

-- 4. permissions seed (generated by awo-gen, same as today)
INSERT INTO permissions (module, resource, action, full_key, description)
VALUES ('forecourt', 'pumps', 'read',      'forecourt.pumps.read',      'View pump status'),
       ('forecourt', 'pumps', 'authorise', 'forecourt.pumps.authorise', 'Authorise pump');

-- 5. Optionally add setting definitions
INSERT INTO setting_definitions (module_id, setting_key, label, value_type, default_value)
SELECT id, 'forecourt.pumps.auto_authorise_threshold',
       'Auto-authorise pumps below this amount', 'decimal', '0'
FROM modules WHERE slug = 'forecourt';
```

No Go nav registry changes. No router constant additions. The data-driven nav picks up the new module once its flag is enabled for a tenant.

---

## 4. Feature Flags — Module & Resource Toggles

### 4.1 What Flags Control

A feature flag is a binary decision: does this module or resource exist for this tenant? When a flag is off, the affected nav section is absent, the schema endpoint returns 404, and the API route returns 403. From the user's perspective, the feature simply does not exist.

Flags operate at two levels:
- **Module flag** (`finance`) — the entire Finance module. When off, no Finance nav, no Finance API.
- **Resource flag** (`finance.transactions`) — a specific resource within an enabled module. When off, that resource's nav item is absent even though other Finance resources (accounts, periods) remain visible.

### 4.2 Schema

```sql
-- Flag catalogue — shared across all tenants, no tenant_id
CREATE TABLE feature_flag_definitions (
  id            uuid  PRIMARY KEY DEFAULT gen_random_uuid(),
  module_id     uuid  REFERENCES modules(id),    -- NULL = global/platform flag
  resource_id   uuid  REFERENCES resources(id),  -- NULL = module-level flag
  flag_key      text  UNIQUE NOT NULL,            -- 'finance' | 'finance.transactions'
  label         text  NOT NULL,
  description   text,
  default_value bool  NOT NULL DEFAULT false,     -- what tenants get without any configuration
  is_system     bool  NOT NULL DEFAULT false      -- true = only platform operators can toggle
);

-- Per-tenant flag values — this IS tenant-scoped
CREATE TABLE tenant_feature_flags (
  id         uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id  uuid        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  flag_id    uuid        NOT NULL REFERENCES feature_flag_definitions(id),
  flag_key   text        NOT NULL,    -- denormalised from definition for fast lookup
  enabled    bool        NOT NULL,
  set_by     uuid        REFERENCES users(id),
  set_at     timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, flag_id)
);

CREATE INDEX idx_tenant_flags_tenant ON tenant_feature_flags(tenant_id);
```

`feature_flag_definitions` is the catalogue — seeded by migrations/triggers, readable by all tenants. `tenant_feature_flags` is the configuration — one row per tenant per flag they have explicitly configured. A flag with no row in `tenant_feature_flags` takes its `default_value` from the definition.

### 4.3 Resolution

```go
// internal/platform/repo/flag_repo_impl.go

func (r *flagRepoImpl) ResolveForTenant(ctx context.Context,
    tenantID uuid.UUID) (map[string]bool, error) {

    // Single query: all definitions LEFT JOIN tenant overrides
    // Returns the effective value for every flag
    rows, err := r.q.ResolveAllFlagsForTenant(ctx, db.ResolveAllFlagsForTenantParams{
        TenantID: tenantID,
    })
    if err != nil { return nil, err }

    flags := make(map[string]bool, len(rows))
    for _, row := range rows {
        // COALESCE(tenant_override, default_value)
        flags[row.FlagKey] = row.EffectiveValue
    }
    return flags, nil
}
```

```sql
-- db/queries/flags.sql
-- name: ResolveAllFlagsForTenant :many
SELECT
    ffd.flag_key,
    COALESCE(tff.enabled, ffd.default_value) AS effective_value
FROM feature_flag_definitions ffd
LEFT JOIN tenant_feature_flags tff
    ON tff.flag_id = ffd.id AND tff.tenant_id = @tenant_id
ORDER BY ffd.flag_key;
```

One query at login. Result stored in `sessions.configuration JSONB`. Every subsequent check is an O(1) map lookup.

### 4.4 Hierarchy Rule

When checking whether a feature is accessible, both the module flag and the resource flag must be true:

```go
func (s *ResolvedSession) FeatureEnabled(flagKey string) bool {
    // Check the exact flag
    if !s.Configuration.Flags[flagKey] { return false }

    // If it's a resource key (contains a dot), also check the module flag
    if idx := strings.Index(flagKey, "."); idx > 0 {
        moduleKey := flagKey[:idx]
        if !s.Configuration.Flags[moduleKey] { return false }
    }
    return true
}

// finance.transactions.approval_workflow is enabled only if:
// flags["finance"] = true AND flags["finance.transactions"] = true AND
// flags["finance.transactions.approval_workflow"] = true
```

---

## 5. Tenant Settings — Behavioural Configuration

### 5.1 Settings vs. Flags

Settings are values that control how a feature behaves when it is on. They are never binary. A flag says "is approval workflow available?". A setting says "what is the approval threshold?". They share the same key namespace but live in separate tables with different types, different validation, and different UI controls.

| Dimension | Feature Flag | Tenant Setting |
|---|---|---|
| Type | `bool` | `text`, `int`, `decimal`, `enum`, `bool` |
| UI control | Toggle switch | Input / select / number |
| Question answered | Does this exist? | How does this behave? |
| Changed by | Platform operator or tenant admin | Tenant admin (within platform-defined limits) |
| Effect when changed | Feature appears or disappears | Form layout or business logic adjusts |
| Invalidates session | Yes (feature may appear/disappear) | Not always (depends on setting) |

### 5.2 Schema

```sql
-- Setting catalogue — shared, no tenant_id, seeded by developers
CREATE TABLE setting_definitions (
  id            uuid  PRIMARY KEY DEFAULT gen_random_uuid(),
  module_id     uuid  REFERENCES modules(id),
  resource_id   uuid  REFERENCES resources(id),
  action_id     uuid  REFERENCES actions(id),  -- NULL for module/resource-level settings
  setting_key   text  UNIQUE NOT NULL,         -- 'finance.transactions.approval_threshold'
  label         text  NOT NULL,
  description   text,
  value_type    text  NOT NULL,  -- 'bool' | 'int' | 'decimal' | 'text' | 'enum'
  default_value text,
  enum_options  jsonb,           -- [{"value":"soft","label":"Soft"},{"value":"hard","label":"Hard"}]
  min_value     text,            -- for numeric validation
  max_value     text,
  is_system     bool NOT NULL DEFAULT false  -- true = only platform can change
);

-- Per-tenant setting values
CREATE TABLE tenant_settings (
  id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id   uuid        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  setting_id  uuid        NOT NULL REFERENCES setting_definitions(id),
  setting_key text        NOT NULL,  -- denormalised for fast lookup
  value       text        NOT NULL,
  set_by      uuid        REFERENCES users(id),
  set_at      timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, setting_id)
);

-- Per-user preferences (display preferences, not business logic)
CREATE TABLE user_preferences (
  id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  pref_key    text        NOT NULL,  -- 'finance.entry_mode', 'finance.show_account_codes'
  value       text        NOT NULL,
  set_at      timestamptz NOT NULL DEFAULT now(),
  UNIQUE (user_id, pref_key)
);
```

### 5.3 Setting Definitions Are Developer-Seeded

Unlike flag definitions which are auto-seeded by triggers, setting definitions are inserted by developers because settings require human decisions about types, defaults, and validation ranges:

```sql
-- Seeded in migration files alongside the module migration
INSERT INTO setting_definitions
    (module_id, setting_key, label, value_type, default_value, min_value)
SELECT m.id, 'finance.transactions.approval_threshold',
       'Approval threshold — transactions above this amount require approval',
       'decimal', '100000', '0'
FROM modules m WHERE m.slug = 'finance';

INSERT INTO setting_definitions
    (module_id, setting_key, label, value_type, default_value, enum_options)
SELECT m.id, 'finance.budget_control_mode',
       'Budget control mode',
       'enum', 'soft',
       '[{"value":"none","label":"None"},{"value":"soft","label":"Warn only"},{"value":"hard","label":"Block"}]'
FROM modules m WHERE m.slug = 'finance';

INSERT INTO setting_definitions
    (module_id, setting_key, label, value_type, default_value)
SELECT m.id, 'finance.decimal_places',
       'Number of decimal places for amounts',
       'int', '2'
FROM modules m WHERE m.slug = 'finance';
```

### 5.4 Resolution

```go
// internal/platform/repo/setting_repo_impl.go

func (r *settingRepoImpl) ResolveForTenant(ctx context.Context,
    tenantID uuid.UUID) (map[string]string, error) {

    rows, _ := r.q.ResolveAllSettingsForTenant(ctx, db.ResolveAllSettingsForTenantParams{
        TenantID: tenantID,
    })
    settings := make(map[string]string, len(rows))
    for _, row := range rows {
        settings[row.SettingKey] = row.EffectiveValue // COALESCE(tenant_value, default_value)
    }
    return settings, nil
}
```

Stored in `sessions.configuration` at login. Accessed via typed helpers:

```go
func (s *ResolvedSession) SettingBool(key string, def bool) bool {
    if v, ok := s.Configuration.Settings[key]; ok {
        if b, err := strconv.ParseBool(v); err == nil { return b }
    }
    return def
}
func (s *ResolvedSession) SettingDecimal(key string, def decimal.Decimal) decimal.Decimal {
    if v, ok := s.Configuration.Settings[key]; ok {
        if d, err := decimal.NewFromString(v); err == nil { return d }
    }
    return def
}
func (s *ResolvedSession) SettingInt(key string, def int) int {
    if v, ok := s.Configuration.Settings[key]; ok {
        if i, err := strconv.Atoi(v); err == nil { return i }
    }
    return def
}
func (s *ResolvedSession) SettingString(key string, def string) string {
    if v, ok := s.Configuration.Settings[key]; ok { return v }
    return def
}
```

---

## 6. Authentication (AuthN) — Who Are You?

### 6.1 Unified Identity Model

All user types share one `users` table. No separate identity stores per surface.

```sql
CREATE TABLE users (
  id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  email         text        UNIQUE NOT NULL,
  user_type     text        NOT NULL,     -- 'platform' | 'tenant' | 'portal' | 'third_party'
  tenant_id     uuid        REFERENCES tenants(id) NULL,  -- NULL for platform users
  principal_id  uuid        NULL,         -- portal users: contact_id or employee_id
  entity_id     uuid        NULL,         -- organisational node (see Section 9)
  display_name  text        NOT NULL,
  password_hash text        NOT NULL,     -- bcrypt cost 12
  mfa_secret    text        NULL,         -- TOTP secret, AES-256-GCM encrypted
  mfa_enabled   bool        NOT NULL DEFAULT false,
  status        text        NOT NULL DEFAULT 'active', -- 'active'|'suspended'|'invited'
  failed_login_attempts  int  NOT NULL DEFAULT 0,
  locked_until           timestamptz NULL,
  created_at    timestamptz NOT NULL DEFAULT now(),
  last_login_at timestamptz NULL
);
```

**Scopes:**
- `tenant_id = NULL` → platform scope: global admins, Awo operators. Use `admin_role` pool.
- `tenant_id = <uuid>` → all other types: employees, portal contacts, API integrations. Immutable after creation. All queries RLS-scoped.

**`principal_id`** — portal users' identity is their business record (a contact or employee). Portal handlers always read `principal_id` from session — never from query params. DB RLS enforces it as a second layer.

**`entity_id`** — the organisational node. Determines which subtree of the hierarchy the user can access. Resolved into `EntityScope` at login (see Section 9).

```go
type UserCreateParams struct {
    Email       string
    DisplayName string
    UserType    domain.UserType
    TenantID    *uuid.UUID   // nil for platform; required for all others
    PrincipalID *uuid.UUID   // portal only
    EntityID    *uuid.UUID   // organisational node
    RoleIDs     []uuid.UUID  // assign at creation
    // CreatedBy from context
}
```

### 6.2 Login Flow

```
Step 1: Credential Verification
  Load user by email → check status != 'suspended' → check locked_until
  bcrypt.CompareHashAndPassword (constant-time)
  On failure: increment failed_login_attempts
    ≥5 failures → locked_until = NOW() + escalating backoff (15m→30m→1h→2h)
  Always return generic "Invalid email or password" — never reveal which field failed

Step 2: MFA (if mfa_enabled = true)
  No code in request → return { mfa_required: true }; UI shows MFA field
  Code present → TOTP validate ±1 window → replay check

Step 3: Session Construction (the expensive step, runs once)
  Generate 32-byte random token; store SHA-256 hash
  ComputePermissions()      → map[string]bool (roles → permissions + deny policies)
  ResolveEntityScope()      → EntityScope (entity_id → subtree path)
  FlagService.Resolve()     → map[string]bool (flag definitions + tenant overrides)
  SettingService.Resolve()  → map[string]string (setting definitions + tenant overrides)
  PrefService.GetForUser()  → map[string]string (user preferences)
  INSERT sessions: token_hash, permissions JSONB, entity_scope JSONB,
                   configuration JSONB {flags, settings, prefs},
                   expires_at = NOW() + 24h

Step 4: Response
  Set HttpOnly+Secure+SameSite=Lax cookie: awo_session = plaintext token
  Return { status: 0, data: { reload_schema: true } }
  amis reloads /schema/boot → BootService builds app shell → full UI
```

The session construction runs five queries at login. After that, every auth, flag, setting, and preference check for the next 24 hours is an in-memory lookup.

### 6.3 Multi-Factor Authentication

TOTP (RFC 6238) only. SMS OTP not supported — SIM-swapping is unacceptable for a financial system.

```go
func (s *IAMService) ValidateMFA(ctx context.Context,
    params domain.ValidateMFAParams) (bool, error) {
    user, _ := s.repo.GetUser(ctx, domain.GetUserParams{UserID: params.UserID})
    secret, _ := s.crypto.Decrypt(user.MFASecret)
    if !gotp.NewDefaultTOTP(secret).VerifyWithWindow(params.Code, time.Now().Unix(), 1) {
        return false, nil
    }
    // Replay prevention: cache used code for 3 windows (90s)
    key := fmt.Sprintf("mfa_used:%s:%d", params.UserID, time.Now().Unix()/30)
    if s.cache.Exists(ctx, key).Val() > 0 { return false, domain.ErrMFACodeReplayed }
    s.cache.Set(ctx, key, "1", 90*time.Second)
    return true, nil
}
```

MFA mandatory for all users with `finance.*` or `platform.*` permissions. Configurable for others via `iam.mfa.required` flag.

### 6.4 Session Architecture

DB sessions — not JWT. JWTs cannot be revoked; a session row is deleted instantly on logout or suspension.

```sql
CREATE TABLE sessions (
  id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id       uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  user_type     text        NOT NULL,
  tenant_id     uuid        NULL,
  token_hash    text        UNIQUE NOT NULL,
  permissions   jsonb       NOT NULL DEFAULT '{}',
  entity_scope  jsonb       NOT NULL DEFAULT '{}',
  configuration jsonb       NOT NULL DEFAULT '{}',  -- {flags:{}, settings:{}, prefs:{}}
  created_at    timestamptz NOT NULL DEFAULT now(),
  expires_at    timestamptz NOT NULL,
  last_seen_at  timestamptz NOT NULL DEFAULT now(),
  ip_address    text,
  user_agent    text
);
CREATE UNIQUE INDEX idx_sessions_token ON sessions(token_hash);
CREATE INDEX idx_sessions_expires ON sessions(expires_at) WHERE expires_at < NOW();
```

Session validation atomically validates + touches `last_seen_at` in one query (sqlc generated). No separate read then update.

**Invalidation triggers:**

| Event | Action |
|---|---|
| Logout | DELETE session row |
| User suspended | DELETE all user's sessions |
| Sensitive permission revoked | DELETE all user's sessions |
| Significant tenant flag changed | DELETE all tenant sessions |
| Session TTL (24h) | Cleanup job |

**Permission staleness:** For non-urgent role additions, staleness up to 24h is acceptable. For immediate-effect changes (termination, suspension), call `InvalidateByUser()` which deletes sessions, forcing fresh permission computation at next login.

### 6.5 Password Management

bcrypt cost 12 (~250ms/hash, ~4 guesses/sec). Requirements: 12+ chars, mixed case + digit + special, not in HIBP top-10k, not in last-5 hashes. Password reset: 32-byte token, 1-hour expiry, stored hashed, one-time-use, invalidates all sessions on success.

### 6.6 OAuth & SSO

Protocols: OpenID Connect 1.0, SAML 2.0. JIT provisioning (auto-create user on first SSO login) controlled by `iam.sso.auto_provision` flag.

### 6.7 API Keys

```sql
CREATE TABLE api_keys (
  id         uuid    PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id  uuid    NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name       text    NOT NULL,
  key_hash   text    UNIQUE NOT NULL,  -- SHA-256
  scopes     text[]  NOT NULL,         -- subset of permissions
  expires_at timestamptz NULL,
  created_by uuid    NOT NULL REFERENCES users(id),
  revoked_at timestamptz NULL
);
```

API key sessions are built at request time — not stored in `sessions` to avoid DB bloat from high-frequency API calls. Scopes are a ceiling: even if the owning user has broader permissions, the key can only exercise what is in `scopes`.

---

## 7. Authorization (AuthZ) — What Can You Do?

### 7.1 Why Casbin

Custom RBAC implementations grow from `if user.role == "admin"` to 2,000+ lines of special-case policy logic. Casbin replaces that with a formally verified model and in-memory sub-millisecond evaluation.

### 7.2 The Four Actor Types

| Actor | Subject Prefix | Domain | Trust |
|---|---|---|---|
| Platform | `platform:` | `_platform_` | Awo staff |
| Tenant | `tenant:` | `<tenant-uuid>` | Employee |
| Portal | `portal:` | `<tenant-uuid>:portal` | External contact |
| Third-Party | `api:` | `<tenant-uuid>:api` | Integration |

`_platform_` is reserved and hardcoded. UUIDs cannot produce it.

### 7.3 Casbin Policy Engine

```ini
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act, eft

[role_definition]
g = _, _, _          # user, role, domain

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))
# DENY OVERRIDE: one deny blocks all allows

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom &&
    keyMatch2(r.obj, p.obj) && r.act == p.act
```

`r.dom == p.dom` — the multi-tenancy isolation. In-memory before any DB query.

`deny-override` — enables hard blocks (terminated employee, sanctions) that no permission grant can circumvent.

### 7.4 Role Management — Three-Guard Assignment

```go
func (s *IAMService) AssignRole(ctx context.Context, params domain.AssignRoleParams) error {
    user,    _ := s.repo.GetUser(ctx, domain.GetUserParams{UserID: params.UserID})
    role,    _ := s.repo.GetRole(ctx, domain.GetRoleParams{RoleID: params.RoleID})
    granter    := s.repo.GetUserFromContext(ctx)

    // Guard 1: namespace — platform roles cannot go to tenant users
    if !slices.Contains(role.AssignableTo, string(user.UserType)) {
        return domain.ErrForbidden.WithMessage(
            "role %q is not assignable to %s users", role.Slug, user.UserType)
    }
    // Guard 2: tenant scope — cannot assign another tenant's roles
    if granter.UserType == domain.UserTypeTenant {
        if role.TenantID != nil && *role.TenantID != granter.TenantID {
            return domain.ErrForbidden.WithMessage("cannot assign role from another tenant")
        }
        if s.roleContainsPlatformPermissions(ctx, params.RoleID) {
            return domain.ErrForbidden.WithMessage(
                "tenant admins cannot assign platform permissions")
        }
    }
    // Guard 3: delegation — grant only roles you hold yourself
    if granter.UserType != domain.UserTypePlatform {
        if !s.repo.UserHasRole(ctx, granter.ID, params.RoleID) {
            return domain.ErrForbidden.WithMessage(
                "you can only grant roles you hold yourself")
        }
    }
    // Write atomically: role_assignments (metadata) + casbin_rule (enforcement)
    return s.db.BeginTxFunc(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
        s.repo.CreateAssignment(ctx, params)
        s.casbin.AddRoleForUserInDomain(subjectString(user), "role:"+role.Slug, tenantDomain(user.TenantID))
        return nil
    })
}
```

**System roles (seeded at provisioning):**

```go
var SystemRoles = []RoleSeed{
    {Slug: "platform.superadmin", AssignableTo: []string{"platform"},
     Permissions: []string{"platform.*"}},
    {Slug: "platform.support",    AssignableTo: []string{"platform"},
     Permissions: []string{"platform.tenants.read","platform.users.read","platform.audit.read"}},
    {Slug: "finance.controller",  AssignableTo: []string{"tenant"},
     Permissions: []string{"finance.*"}},
    {Slug: "finance.accountant",  AssignableTo: []string{"tenant"},
     Permissions: []string{
         "finance.accounts.read","finance.accounts.create",
         "finance.transactions.read","finance.transactions.create","finance.transactions.post"}},
    {Slug: "finance.viewer",      AssignableTo: []string{"tenant"},
     Permissions: []string{"finance.accounts.read","finance.transactions.read"}},
    {Slug: "portal.customer",     AssignableTo: []string{"portal"},
     Permissions: []string{"portal.self.invoices.read","portal.self.statements.read"}},
    {Slug: "portal.employee",     AssignableTo: []string{"portal"},
     Permissions: []string{"portal.self.payslips.read","portal.self.leave.create"}},
}
```

### 7.5 Temporal Roles

```sql
CREATE TABLE role_assignments (
  id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id    uuid        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id      uuid        NOT NULL,
  role_slug    text        NOT NULL,
  domain       text        NOT NULL,
  granted_by   uuid        REFERENCES users(id),
  granted_at   timestamptz NOT NULL DEFAULT now(),
  expires_at   timestamptz NULL,      -- NULL = permanent
  revoked_at   timestamptz NULL,
  revoke_reason text
);
-- Partial index: only rows with expiry are indexed — near-free check
CREATE INDEX idx_role_assignments_expiry ON role_assignments(expires_at)
    WHERE expires_at IS NOT NULL;
```

Lazy expiry fires on first request after expiry time. For hard deadlines, supplement with an explicit scheduled `RevokeRole()` call.

### 7.6 Deny Policies

```go
// Blanket deny — terminated employee; overrides all allow rules
svc.AddPolicy(ctx, domain.PolicyParams{
    Subject: "tenant:usr_terminated_042", Domain: tenantDomain,
    Object: "*", Action: "*", Effect: "deny",
})
```

All deny additions are audit-logged. Platform UI shows active deny rules with single-click revoke.

---

## 8. Database Architecture

### 8.1 Table Summary

| Table | Tenant-scoped? | Purpose |
|---|---|---|
| `modules` | No | MRA catalogue — system vocabulary |
| `resources` | No | MRA catalogue |
| `actions` | No | MRA catalogue |
| `permissions` | No | Permission key catalogue |
| `feature_flag_definitions` | No | Flag catalogue, auto-seeded from MRA |
| `setting_definitions` | No | Setting catalogue, developer-seeded |
| `users` | Yes | All user types |
| `sessions` | Yes | Active sessions with pre-computed state |
| `roles` | Hybrid | System roles: no tenant_id; tenant roles: tenant_id |
| `role_assignments` | Yes | Who holds which role + expiry |
| `role_permissions` | No (via roles) | Which permissions a role holds |
| `tenant_feature_flags` | Yes | Per-tenant flag values |
| `tenant_settings` | Yes | Per-tenant setting values |
| `user_preferences` | Yes (per-user) | Per-user display preferences |
| `casbin_rule` | No (via domain field) | Casbin enforcement |
| `api_keys` | Yes | Third-party API keys |
| `entities` | Yes | Organisational hierarchy |

### 8.2 Casbin Tables

```sql
CREATE TABLE casbin_rule (
  id    bigserial PRIMARY KEY,
  ptype text NOT NULL,
  v0    text NOT NULL DEFAULT '',  -- subject
  v1    text NOT NULL DEFAULT '',  -- domain (p) | role (g)
  v2    text NOT NULL DEFAULT '',  -- object (p) | domain (g)
  v3    text NOT NULL DEFAULT '',  -- action (p)
  v4    text NOT NULL DEFAULT '',  -- effect: 'allow'|'deny'
  v5    text NOT NULL DEFAULT ''
);
CREATE UNIQUE INDEX idx_casbin_rule_unique ON casbin_rule(ptype, v0, v1, v2, v3, v4, v5);
CREATE INDEX idx_casbin_rule_domain ON casbin_rule(v1) WHERE ptype = 'p';
```

**Two-table design:** `casbin_rule` = Casbin's enforcement source (never write directly). `role_assignments` = application metadata (UI, audit, expiry). Both updated atomically by `AssignRole()`.

### 8.3 Row-Level Security

```sql
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE role_assignments ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_feature_flags ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_settings ENABLE ROW LEVEL SECURITY;
ALTER TABLE entities ENABLE ROW LEVEL SECURITY;

-- Shared tables: no RLS, just GRANT
GRANT SELECT ON modules, resources, actions, permissions,
                feature_flag_definitions, setting_definitions
    TO application_role;

-- Tenant-scoped tables: RLS by tenant_id
CREATE POLICY tenant_isolation ON tenant_feature_flags FOR ALL TO application_role
    USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

CREATE POLICY tenant_isolation ON tenant_settings FOR ALL TO application_role
    USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

-- casbin_rule: no tenant RLS — Casbin loads all rules at startup
-- Isolation enforced by r.dom==p.dom in Casbin's in-memory model
CREATE POLICY application_full_access ON casbin_rule FOR ALL TO application_role USING (TRUE);
```

### 8.4 The Two PostgreSQL Roles

```
application_role        All tenant/portal requests. Subject to RLS.
               Per-connection: app.tenant_id, app.user_id, app.user_type,
                               app.principal_id, app.entity_id

admin_role   Platform operators only. BYPASSRLS at PostgreSQL engine level.
               Accesses platform_views schema (cross-tenant aggregates).
```

```go
func SetDBPool(pools *db.Pools) fiber.Handler {
    return func(c *fiber.Ctx) error {
        session := ContextSession(c)
        if session.IsPlatform() {
            c.Locals("db", pools.Platform)
        } else {
            conn, _ := pools.App.Acquire(c.Context())
            conn.Exec(c.Context(), `
                SELECT set_config('app.tenant_id',    $1, true),
                       set_config('app.user_id',      $2, true),
                       set_config('app.user_type',    $3, true),
                       set_config('app.principal_id', $4, true),
                       set_config('app.entity_id',    $5, true)
            `, session.TenantID, session.UserID,
               session.UserType, session.PrincipalID, session.EntityID)
            c.Locals("db", conn)
        }
        return c.Next()
    }
}
```

---

## 9. Resource Hierarchy & Entity Model

### 9.1 The Entity Tree

Every tenant's organisation is a tree. Access control follows the tree: access to a node implies access to all its descendants.

```
Company (Root)
└── Sub-Company | Region | Division | Department
    └── Branch | Cost Centre | Project (Leaf)
```

```sql
CREATE TABLE entities (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id    uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name         text NOT NULL,
  entity_type  text NOT NULL,  -- 'company'|'subsidiary'|'department'|'region'|'branch'|...
  parent_id    uuid REFERENCES entities(id) ON DELETE RESTRICT,
  entity_path  text,           -- materialized path: /root_id/parent_id/this_id/
  entity_level int  NOT NULL DEFAULT 1,
  is_active    bool NOT NULL DEFAULT true,
  created_at   timestamptz NOT NULL DEFAULT now()
);
```

### 9.2 Access Rules

```
Root entity     → access everything (EntityScope{Type: "all"})
Mid-level node  → access own entity + all descendants (EntityScope{Type: "subtree"})
Leaf node       → access own entity only (EntityScope{Type: "entity"})
```

Resolved at login, stored in `sessions.entity_scope`. Applied by repos:

```go
func (r *transactionRepoImpl) List(ctx context.Context,
    params domain.TransactionListParams) ([]*domain.Transaction, int, error) {
    scope := domain.EntityScopeFromContext(ctx)
    switch scope.Type {
    case "all":
        return r.q.ListTransactions(ctx, ...)
    case "subtree":
        return r.q.ListTransactionsByEntitySubtree(ctx,
            db.ListTransactionsByEntitySubtreeParams{PathPrefix: scope.PathPrefix + "%"})
    default:
        return r.q.ListTransactionsByEntity(ctx,
            db.ListTransactionsByEntityParams{EntityID: scope.EntityID})
    }
}
```

### 9.3 Ownership Fields

Every business table carries:

```sql
created_by   uuid NOT NULL REFERENCES users(id),
entity_id    uuid NOT NULL REFERENCES entities(id)
```

Set automatically in every service `Create` method from context, never from request params:

```go
func (s *TransactionService) Create(ctx context.Context,
    params domain.TransactionCreateParams) (*domain.Transaction, error) {
    params.CreatedBy, _ = domain.UserIDFromContext(ctx)
    params.EntityID, _  = domain.EntityIDFromContext(ctx)
    return s.repo.Create(ctx, params)
}
```

---

## 10. The Session — Pre-Computed Everything

### 10.1 Why Pre-Compute

A single page load triggers 8–12 API calls. Each call could naively query the DB for permissions, flags, settings, and entity scope. At load, this becomes the dominant source of latency. The solution: compute everything once at login, store it in the session JSONB, read it as in-memory map lookups on every subsequent request.

### 10.2 What Gets Built at Login

```go
// internal/platform/iam.go — called once during session creation

func (s *IAMService) buildSession(ctx context.Context,
    user *domain.User) (*domain.ResolvedSession, error) {

    // Five queries, run in parallel where possible
    var (
        perms    map[string]bool   // from role→permissions join + deny policies
        scope    domain.EntityScope
        flags    map[string]bool
        settings map[string]string
        prefs    map[string]string
    )

    g, gctx := errgroup.WithContext(ctx)
    g.Go(func() error { var err error; perms,    err = s.computePermissions(gctx, user); return err })
    g.Go(func() error { var err error; scope,    err = s.resolveEntityScope(gctx, user.EntityID); return err })
    g.Go(func() error { var err error; flags,    err = s.flagRepo.ResolveForTenant(gctx, user.TenantID); return err })
    g.Go(func() error { var err error; settings, err = s.settingRepo.ResolveForTenant(gctx, user.TenantID); return err })
    g.Go(func() error { var err error; prefs,    err = s.prefRepo.GetForUser(gctx, user.ID); return err })
    if err := g.Wait(); err != nil { return nil, err }

    return &domain.ResolvedSession{
        UserID:      user.ID,
        UserType:    user.UserType,
        TenantID:    user.TenantID,
        PrincipalID: user.PrincipalID,
        EntityID:    user.EntityID,
        DisplayName: user.DisplayName,
        Permissions: perms,
        EntityScope: scope,
        Configuration: domain.SessionConfiguration{
            Flags:    flags,
            Settings: settings,
            Prefs:    prefs,
        },
    }, nil
}
```

### 10.3 The ResolvedSession Type

```go
type ResolvedSession struct {
    UserID        uuid.UUID
    UserType      UserType
    TenantID      uuid.UUID
    PrincipalID   uuid.UUID
    EntityID      uuid.UUID
    DisplayName   string
    Permissions   map[string]bool      // "finance.transactions.approve" → true/false
    EntityScope   EntityScope          // type + path prefix for subtree queries
    Configuration SessionConfiguration
}

type SessionConfiguration struct {
    Flags    map[string]bool   // "finance.transactions.approval_workflow" → true
    Settings map[string]string // "finance.transactions.approval_threshold" → "100000"
    Prefs    map[string]string // "finance.entry_mode" → "spreadsheet"
}

// All checks are O(1) map lookups — no DB
func (s *ResolvedSession) Can(resource, action string) bool {
    v, ok := s.Permissions[resource+"."+action]
    return ok && v
}
func (s *ResolvedSession) FeatureEnabled(key string) bool {
    if !s.Configuration.Flags[key] { return false }
    if idx := strings.Index(key, "."); idx > 0 {
        if !s.Configuration.Flags[key[:idx]] { return false }
    }
    return true
}
func (s *ResolvedSession) IsPlatform() bool { return s.UserType == UserTypePlatform }
func (s *ResolvedSession) IsPortal()   bool { return s.UserType == UserTypePortal }
```

### 10.4 Session Invalidation

When a flag or setting that affects security changes, existing sessions are stale. The `SettingService.Update()` and `FlagService.Set()` methods check if the change is significant:

```go
func (s *FlagService) Set(ctx context.Context,
    params domain.SetFlagParams) error {

    def, _ := s.repo.GetDefinition(ctx, params.FlagKey)
    if err := s.repo.UpsertTenantFlag(ctx, params); err != nil { return err }

    // A flag that makes a feature appear or disappear requires session refresh
    // so users immediately see/lose access to the feature
    if def.IsModuleOrResourceFlag() {
        s.sessionRepo.InvalidateByTenant(ctx, params.TenantID)
    }
    return nil
}
```

---

## 11. UI Structure — How Configuration Becomes Interface

### 11.1 The Data-Driven Navigation

The app shell navigation is not hardcoded in Go. It is built by querying the MRA tables, filtered through flags and permissions. The `BootService` runs this at every `/schema/boot` call for an authenticated user.

```
MRA tables (modules + resources)
     ↓
Filter by tenant flags           → only enabled modules and resources
     ↓
Filter by session permissions    → only resources user can read
     ↓
Build nav sections               → what the user sees in the sidebar
```

```go
// internal/platform/boot.go

func (s *BootService) BuildAppShell(ctx context.Context,
    session *domain.ResolvedSession) (map[string]any, error) {

    // One DB query: modules → resources, filtered by tenant flags
    enabledModules, _ := s.moduleRepo.ListEnabledWithResources(ctx,
        domain.ListEnabledModulesParams{TenantID: session.TenantID})

    nav := s.buildNav(enabledModules, session.Permissions)

    return map[string]any{
        "type":      "app",
        "brandName": "Awo",
        "pages":     nav,
    }, nil
}

func (s *BootService) buildNav(modules []domain.ModuleWithResources,
    perms map[string]bool) []any {

    var sections []any
    for _, mod := range modules {
        // Module already passed the flag filter in the DB query
        var items []any
        for _, res := range mod.Resources {
            readKey := mod.Slug + "." + res.Slug + ".read"
            if !perms[readKey] {
                continue  // no read permission → not visible
            }
            items = append(items, map[string]any{
                "label": res.Label,
                "url":   res.NavURL,
                "schema": map[string]any{
                    "type":      "service",
                    "schemaApi": "/schema" + res.NavURL,
                    "fallback":  errorFallback(res.Label),
                },
            })
        }
        if len(items) > 0 {
            sections = append(sections, map[string]any{
                "label":    mod.Label,
                "icon":     mod.Icon,
                "children": items,
            })
        }
    }
    return sections
}
```

**The query that powers this:**

```sql
-- db/queries/modules.sql
-- name: ListEnabledModulesWithResources :many
SELECT
    m.id, m.slug AS module_slug, m.label AS module_label,
    m.icon, m.nav_order AS module_nav_order,
    r.id AS resource_id, r.slug AS resource_slug,
    r.label AS resource_label, r.nav_url, r.nav_order AS resource_nav_order
FROM modules m
JOIN resources r ON r.module_id = m.id

-- Module must be enabled (flag=true or no tenant override and default=true)
JOIN feature_flag_definitions mfd ON mfd.module_id = m.id AND mfd.resource_id IS NULL
LEFT JOIN tenant_feature_flags mtf ON mtf.flag_id = mfd.id AND mtf.tenant_id = @tenant_id
WHERE COALESCE(mtf.enabled, mfd.default_value) = true
  AND m.is_active = true

-- Resource must also be enabled
LEFT JOIN feature_flag_definitions rfd ON rfd.resource_id = r.id
LEFT JOIN tenant_feature_flags rtf ON rtf.flag_id = rfd.id AND rtf.tenant_id = @tenant_id
WHERE (rfd.id IS NULL OR COALESCE(rtf.enabled, rfd.default_value) = true)

ORDER BY m.nav_order, r.nav_order;
```

One query. Returns only what exists and is enabled. Permissions filter is applied in Go over the returned set. Adding a new module requires only DB rows — no Go nav code changes.

### 11.2 UI Form Sections Adapt to Flags and Settings

Forms are built in sections. Each section is conditionally rendered based on the session's pre-computed flags and settings — zero additional DB queries:

```
┌─────────────────────────────────────────────────────────────┐
│ SECTION 1: Transaction Header (always shown)                │
│ Date | Number | Type | Description                         │
├─────────────────────────────────────────────────────────────┤
│ SECTION 2: Currency                                         │
│ Shown when: session.FeatureEnabled("finance.multi_currency")│
├─────────────────────────────────────────────────────────────┤
│ SECTION 3: Dimensions                                       │
│ Cost Centre: shown always; required if setting = true       │
│ Project: shown if FeatureEnabled("finance.project_tracking")│
├─────────────────────────────────────────────────────────────┤
│ SECTION 4: Entry Lines (always shown)                       │
│ Account | Description | Debit | Credit                      │
├─────────────────────────────────────────────────────────────┤
│ SECTION 5: Approval Info (read-only after submission)       │
│ Shown when: FeatureEnabled("finance.transactions            │
│                             .approval_workflow")            │
│ AND total > settings.SettingDecimal("approval_threshold")   │
└─────────────────────────────────────────────────────────────┘
```

```go
// internal/api/handlers/schema_handler.go

func TransactionFormSchema(deps *app.Deps) fiber.Handler {
    return func(c *fiber.Ctx) error {
        session := middleware.ContextSession(c)

        cfg := TransactionFormConfig{
            // Flags — from session, zero DB hit
            ShowCurrencyField:   session.FeatureEnabled("finance.multi_currency"),
            ShowProjectField:    session.FeatureEnabled("finance.project_tracking"),
            ShowApprovalSection: session.FeatureEnabled("finance.transactions.approval_workflow"),

            // Settings — from session, typed helpers
            RequireCostCenter:   session.SettingBool("finance.transactions.cost_center_required", false),
            ApprovalThreshold:   session.SettingDecimal("finance.transactions.approval_threshold", decimal.Zero),
            DecimalPlaces:       session.SettingInt("finance.decimal_places", 2),

            // Permissions — from session
            CanPost:             session.Can("finance.transactions", "post"),
            CanApprove:          session.Can("finance.transactions", "approve"),
            CanVoid:             session.Can("finance.transactions", "void"),

            // User preferences — from session
            EntryMode:           session.Configuration.Prefs["finance.entry_mode"],  // "form" | "spreadsheet"
            ShowAccountCodes:    session.Configuration.Prefs["finance.show_account_codes"] == "true",
        }

        return c.JSON(buildTransactionForm(cfg))
    }
}
```

### 11.3 The Settings Screen — Data-Driven Form

Tenant admin settings screens are generated from `setting_definitions`. No hardcoded forms:

```go
// internal/api/handlers/schema_handler.go

func FinanceSettingsSchema(deps *app.Deps) fiber.Handler {
    return func(c *fiber.Ctx) error {
        session := middleware.ContextSession(c)
        if !session.Can("settings.finance", "update") {
            return c.Status(403).JSON(response.Err("access denied"))
        }

        // Load definitions for the finance module with current tenant values
        defs, _ := deps.Platform.Settings.ListForModule(c.Context(),
            domain.ListSettingsParams{ModuleSlug: "finance"})

        return c.JSON(map[string]any{
            "type":  "page",
            "title": "Finance Configuration",
            "body": map[string]any{
                "type": "form",
                "api":  "patch:/api/v1/settings/finance",
                "body": buildSettingsFields(defs),
            },
        })
    }
}

func buildSettingsFields(defs []domain.SettingDefinitionWithValue) []any {
    var fields []any
    for _, def := range defs {
        field := map[string]any{
            "name":  def.SettingKey,
            "label": def.Label,
            "value": def.CurrentValue,
        }
        switch def.ValueType {
        case "bool":
            field["type"] = "switch"
        case "decimal":
            field["type"] = "input-number"
            field["precision"] = 2
        case "int":
            field["type"] = "input-number"
            field["precision"] = 0
        case "enum":
            field["type"] = "select"
            field["options"] = def.EnumOptions
        default:
            field["type"] = "input-text"
        }
        if def.IsSystem {
            field["disabled"] = true  // tenant admin cannot change system settings
            field["hint"] = "Managed by Awo platform"
        }
        fields = append(fields, field)
    }
    return fields
}
```

Adding a new setting requires one SQL row in `setting_definitions`. The settings screen renders it automatically. No schema handler changes.

### 11.4 The Flags Screen — Module/Resource Toggles

Feature flag management for tenant admins:

```go
func ModuleFlagsSchema(deps *app.Deps) fiber.Handler {
    return func(c *fiber.Ctx) error {
        session := middleware.ContextSession(c)
        if !session.Can("settings.modules", "update") {
            return c.Status(403).JSON(response.Err("access denied"))
        }

        // Load all non-system flag definitions with current tenant values
        flags, _ := deps.Platform.Flags.ListForTenant(c.Context(),
            domain.ListFlagsParams{
                TenantID:       session.TenantID,
                ExcludeSystem:  true,  // system flags not shown to tenant admins
            })

        // Group by module for a cleaner UI
        return c.JSON(buildFlagsForm(flags))
    }
}
```

The flags form is a series of toggle switches, grouped by module. Each module has a master toggle; when a module is turned off, all its resource toggles are disabled in the UI (and the module flag check in the nav query handles the actual enforcement).

### 11.5 Route-Level Flag Enforcement

Flags gate routes at the middleware level, not just the UI:

```go
// Three independent gates for a feature — all must pass

// Gate 1: nav (boot schema) — section only appears if module+resource flags are on
// Gate 2: schema route — 403 if flag is off
sg.Get("/accounting/transactions",
    middleware.RequireFlag("finance.transactions"),   // resource flag
    middleware.RequirePermission("finance.transactions", "read"),
    handlers.TransactionListSchema(deps))

// Gate 3: data API route — same flag check
api.Get("/transactions",
    middleware.RequireFlag("finance.transactions"),
    middleware.RequirePermission("finance.transactions", "read"),
    handlers.ListTransactions(deps))
```

All three read from the session's pre-computed flag map. Zero DB hits. A user who navigates directly to `/schema/accounting/transactions` when the flag is off gets 403, not 404. The distinction is intentional — 404 implies the route doesn't exist, 403 implies it exists but is gated.

---

## 12. HTTP Middleware Chain & API Layer

### 12.1 Router Structure

```go
func BuildRouter(app *fiber.App, deps *Deps) {
    // Infrastructure — all requests
    app.Use(middleware.Recovery(), middleware.Logger(), middleware.RateLimit(),
            middleware.CORS(), middleware.SecurityHeaders())

    // Public — no auth
    app.Post("/auth/login",                    handlers.Login(deps))
    app.Post("/auth/logout",                   handlers.Logout(deps))
    app.Post("/auth/forgot-password",          handlers.ForgotPassword(deps))
    app.Post("/auth/reset-password",           handlers.ResetPassword(deps))
    app.Get("/auth/oauth/:provider",           handlers.OAuthRedirect(deps))
    app.Get("/auth/oauth/:provider/callback",  handlers.OAuthCallback(deps))
    app.Get("/schema/boot",                    handlers.Boot(deps))
    app.Static("/static", "./web/static")

    // Authenticated — all routes below require valid session
    auth := app.Group("",
        middleware.Authenticate(deps.Platform.IAM),
        middleware.SetDBPool(deps.Pools),
        middleware.ResolveTenant(deps.Platform.Tenant),
        middleware.AuditWrap(deps.Platform.Audit),
    )

    // Schema routes — build UI schema for amis frontend
    sg := auth.Group("/schema")
    sg.Get("/accounting/transactions",
        middleware.RequireFlag("finance.transactions"),
        middleware.RequirePermission("finance.transactions", "read"),
        handlers.TransactionListSchema(deps))
    sg.Get("/settings/modules",
        middleware.RequirePermission("settings.modules", "read"),
        handlers.ModuleFlagsSchema(deps))
    sg.Get("/settings/finance",
        middleware.RequirePermission("settings.finance", "read"),
        handlers.FinanceSettingsSchema(deps))

    // Data API routes
    api := auth.Group("/api/v1")
    api.Get("/transactions",
        middleware.RequireFlag("finance.transactions"),
        middleware.RequirePermission("finance.transactions", "read"),
        handlers.ListTransactions(deps))
    api.Patch("/settings/flags/:key",
        middleware.RequirePermission("settings.modules", "update"),
        handlers.SetFlag(deps))
    api.Patch("/settings/:module",
        middleware.RequirePermission("settings.finance", "update"),
        handlers.UpdateModuleSettings(deps))
}
```

### 12.2 Middleware Implementations

```go
func RequirePermission(resource, action string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        if !ContextSession(c).Can(resource, action) {
            return c.Status(403).JSON(response.Err(
                fmt.Sprintf("permission denied: %s.%s", resource, action)))
        }
        return c.Next()
    }
}

func RequireFlag(flagKey string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        if !ContextSession(c).FeatureEnabled(flagKey) {
            return c.Status(403).JSON(response.Err(
                "this feature is not enabled for your organisation"))
        }
        return c.Next()
    }
}

// Context helpers — used in every handler
func ContextSession(c *fiber.Ctx) *domain.ResolvedSession  { return c.Locals("session").(*domain.ResolvedSession) }
func ContextTenantID(c *fiber.Ctx) uuid.UUID               { return ContextSession(c).TenantID }
func ContextUserID(c *fiber.Ctx) uuid.UUID                 { return ContextSession(c).UserID }
func ContextEntityID(c *fiber.Ctx) uuid.UUID               { return ContextSession(c).EntityID }
func ContextPrincipalID(c *fiber.Ctx) uuid.UUID            { return ContextSession(c).PrincipalID }
```

---

## 13. Domain Isolation

Three independent layers enforce that Tenant A cannot access Tenant B's data:

1. **Casbin `r.dom == p.dom`** — in-memory, before any DB query
2. **DB RLS** — engine-level, `current_setting('app.tenant_id')` on every query
3. **Service-layer `tenantID`** — explicit in every query's WHERE clause

The `_platform_` domain is reserved. Platform operators inspecting a tenant pass `?tenant_id=<uuid>` for display context, but their permissions come from `_platform_` policies and the `admin_role` pool bypasses RLS.

Cross-tenant **writes** are never permitted. Cross-tenant reads are permitted only through the `admin_role` pool for platform operators, using `platform_views` schema queries.

---

## 14. IAM Services, Resources & Actions

### 14.1 AuthService

```go
type AuthService interface {
    Login(ctx, params domain.LoginParams)                    (*domain.ResolvedSession, string, error)
    ValidateMFA(ctx, params domain.ValidateMFAParams)         (bool, error)
    ValidateSession(ctx, params domain.ValidateSessionParams) (*domain.ResolvedSession, error)
    Logout(ctx, params domain.LogoutParams)                  error
    RequestPasswordReset(ctx, params domain.PasswordResetRequestParams) error
    ResetPassword(ctx, params domain.PasswordResetParams)    error
    OAuthCallback(ctx, params domain.OAuthCallbackParams)    (*domain.ResolvedSession, string, error)
    InitiateMFA(ctx, params domain.InitiateMFAParams)        (*domain.MFAChallenge, error)
    DisableMFA(ctx, params domain.DisableMFAParams)          error
    ComputePermissions(ctx, params domain.ComputePermissionsParams) (map[string]bool, error)
}
```

### 14.2 IdentityService

```go
type IdentityService interface {
    CreateUser(ctx, params domain.UserCreateParams)          (*domain.User, error)
    InviteUser(ctx, params domain.UserInviteParams)          (*domain.Invitation, error)
    AcceptInvitation(ctx, params domain.AcceptInviteParams)  (*domain.User, error)
    GetUser(ctx, params domain.GetUserParams)                (*domain.User, error)
    ListUsers(ctx, params domain.UserListParams)             ([]*domain.User, int, error)
    UpdateUser(ctx, params domain.UserUpdateParams)          (*domain.User, error)
    DeactivateUser(ctx, params domain.DeactivateParams)      error
    SuspendUser(ctx, params domain.SuspendParams)            error
    UnsuspendUser(ctx, params domain.UnsuspendParams)        error
}
```

| Action | Permission |
|---|---|
| List | `iam.users.read` |
| Invite | `iam.users.create` |
| Update | `iam.users.update` |
| Deactivate | `iam.users.delete` |
| Suspend | `iam.users.suspend` |

### 14.3 AccessService

```go
type AccessService interface {
    CreateRole(ctx, params domain.RoleCreateParams)           (*domain.Role, error)
    UpdateRole(ctx, params domain.RoleUpdateParams)           (*domain.Role, error)
    DeleteRole(ctx, params domain.RoleDeleteParams)           error
    ListRoles(ctx, params domain.RoleListParams)              ([]*domain.Role, error)
    SyncRolePermissions(ctx, params domain.SyncPermsParams)   error
    AssignRole(ctx, params domain.AssignRoleParams)           error
    RevokeRole(ctx, params domain.RevokeRoleParams)           error
    GetUserRoles(ctx, params domain.GetRolesParams)           ([]*domain.Role, error)
    ListAssignments(ctx, params domain.AssignmentListParams)  ([]*domain.Assignment, int, error)
    Can(ctx, params domain.CanParams)                         (bool, error)
    GetPermissionMap(ctx, params domain.PermMapParams)        (map[string]bool, error)
}
```

### 14.4 FlagService

```go
type FlagService interface {
    // Catalogue (read-only at runtime; write via migrations/admin)
    ListDefinitions(ctx, params domain.ListFlagDefsParams)    ([]*domain.FlagDefinition, error)
    GetDefinition(ctx, flagKey string)                        (*domain.FlagDefinition, error)

    // Tenant configuration
    ResolveForTenant(ctx, tenantID uuid.UUID)                 (map[string]bool, error)
    ListForTenant(ctx, params domain.ListFlagsParams)         ([]*domain.TenantFlagWithDef, error)
    Set(ctx, params domain.SetFlagParams)                     error  // enables/disables
    Reset(ctx, params domain.ResetFlagParams)                 error  // removes override, restores default
}
```

### 14.5 SettingService

```go
type SettingService interface {
    // Catalogue (read-only at runtime)
    ListDefinitions(ctx, params domain.ListSettingDefsParams) ([]*domain.SettingDefinition, error)

    // Tenant configuration
    ResolveForTenant(ctx, tenantID uuid.UUID)                 (map[string]string, error)
    ListForModule(ctx, params domain.ListSettingsParams)      ([]*domain.SettingWithValue, error)
    UpdateModuleSettings(ctx, params domain.UpdateSettingsParams) error
    GetSetting(ctx, params domain.GetSettingParams)           (string, error)
    ResetToDefault(ctx, params domain.ResetSettingParams)     error

    // User preferences
    GetUserPreferences(ctx, userID uuid.UUID)                 (map[string]string, error)
    SetUserPreference(ctx, params domain.SetPrefParams)       error
}
```

### 14.6 BootService

```go
type BootService interface {
    // Called by GET /schema/boot on every page load
    BuildAppShell(ctx, session *domain.ResolvedSession)       (map[string]any, error)

    // Called by GET /schema/boot when no session exists — returns login form schema
    LoginSchema(message string)                               map[string]any
}
```

---

## 15. Cross-Module Integration

### 15.1 The Integration Contract

Every business module uses three mechanisms to interact with IAM and configuration:

1. `middleware.RequirePermission(resource, action)` — route-level gate; blocks before handler runs
2. `middleware.RequireFlag(flagKey)` — route-level feature gate
3. `session.Can(resource, action)` — in-handler conditional logic (show/hide UI elements)
4. `session.FeatureEnabled(flagKey)` — in-handler conditional for optional form sections
5. `session.SettingDecimal/Int/Bool/String(key, default)` — read tenant configuration

Modules never call `AccessService`, `FlagService`, or `SettingService` directly in handlers.

### 15.2 Finance Module Integration

```go
// Route: permission + flag both required
api.Post("/transactions/:id/post",
    middleware.RequireFlag("finance.transactions"),
    middleware.RequirePermission("finance.transactions", "post"),
    handlers.PostTransaction(deps))

// Schema handler: one function, all configuration from session
func TransactionDetailSchema(deps *app.Deps) fiber.Handler {
    return func(c *fiber.Ctx) error {
        s := middleware.ContextSession(c)

        // Approval section — only if flag is on AND amount exceeds threshold
        // This logic runs from the session; zero DB hits
        showApproval := s.FeatureEnabled("finance.transactions.approval_workflow")
        threshold    := s.SettingDecimal("finance.transactions.approval_threshold", decimal.Zero)

        return c.JSON(financeschema.TransactionDetail(finance.TransactionViewConfig{
            ShowApprovalSection: showApproval,
            ApprovalThreshold:   threshold,
            CanPost:    s.Can("finance.transactions", "post"),
            CanApprove: s.Can("finance.transactions", "approve"),
            CanVoid:    s.Can("finance.transactions", "void"),
            CanReverse: s.Can("finance.transactions", "reverse"),
        }))
    }
}
```

### 15.3 Tenant Lifecycle Events

```go
func (s *IAMService) OnTenantProvisioned(ctx context.Context, e domain.TenantProvisioned) error {
    return s.seedSystemRoles(ctx, e.TenantID)
    // Flag definitions were already auto-seeded by DB triggers when modules were created
    // Tenant starts with all default flag values; admin configures from there
}

func (s *IAMService) OnTenantSuspended(ctx context.Context, e domain.TenantSuspended) error {
    s.sessionRepo.InvalidateByTenant(ctx, e.TenantID)
    s.casbin.AddPolicy(ctx, domain.PolicyParams{
        Subject: "tenant:*", Domain: e.TenantID.String(),
        Object: "*", Action: "*", Effect: "deny",
    })
    return nil
}
```

---

## 16. Audit Trail

### 16.1 Event-Driven: The Chosen Approach

| Approach | Verdict |
|---|---|
| DB triggers | Captures even direct DB changes; no app context (IP, user agent); hard to maintain; not used as primary |
| Service-level calls | Full control; coupling; easy to omit; not used |
| **Event-driven (chosen)** | Zero coupling; centralised; async; durable outbox prevents loss |

IAM emits domain events after every successful state change. The Audit module subscribes. IAM never imports Audit.

Audit events for *failures* (failed login, permission denied on sensitive operations) are emitted on the failure path — they are the security record.

Durable delivery: events are written to a `domain_events` table in the same DB transaction as the operation. A background worker processes them. No event is lost even during downtime.

### 16.2 IAM Events

```go
// internal/platform/domain/events.go

UserCreated         { UserID, Email, UserType, TenantID, CreatedBy, OccurredAt }
UserSuspended       { UserID, SuspendedBy, Reason, OccurredAt }
UserDeactivated     { UserID, DeactivatedBy, OccurredAt }
SessionStarted      { UserID, SessionID, IP, UserAgent, OccurredAt }
SessionEnded        { UserID, SessionID, Reason, OccurredAt }
LoginFailed         { Email, IP, Reason, AttemptCount, OccurredAt }
MFAEnabled          { UserID, OccurredAt }
MFADisabled         { UserID, DisabledBy, OccurredAt }
PasswordChanged     { UserID, ChangedBy, OccurredAt }
RoleAssigned        { UserID, RoleID, RoleSlug, Domain, GrantedBy, ExpiresAt, OccurredAt }
RoleRevoked         { UserID, RoleID, RoleSlug, Domain, RevokedBy, Reason, OccurredAt }
PolicyAdded         { Subject, Domain, Object, Action, Effect, AddedBy, OccurredAt }
PolicyRemoved       { Subject, Domain, Object, Action, RemovedBy, OccurredAt }
FlagChanged         { TenantID, FlagKey, OldValue, NewValue, ChangedBy, OccurredAt }
SettingChanged      { TenantID, SettingKey, OldValue, NewValue, ChangedBy, OccurredAt }
```

The `AuditWrap` middleware records an HTTP-level trace (method, path, status, duration, IP) for all mutating requests. This coarser layer runs alongside domain event audit.

---

## 17. Security Threat Model

| Threat | Key Mitigations |
|---|---|
| Password brute-force | bcrypt cost 12 (~4 guesses/sec); lockout after 5 fails (15m→30m→1h→2h); 10/min/IP rate limit; HIBP check; generic errors (no user enumeration) |
| Session token theft | HttpOnly+Secure+SameSite=Lax; 24h absolute TTL; SHA-256 hash stored, not plaintext; instant deletion on logout/suspend |
| Privilege escalation | Guard 1 (namespace) + Guard 2 (tenant scope) + Guard 3 (delegation) — all atomic, all audited |
| Cross-tenant data access | Casbin `r.dom==p.dom` + DB RLS + service-layer tenantID — three independent layers |
| MFA code replay | Used codes cached 90s; one code valid once per 30-second window |
| Flag/setting manipulation | Flags/settings require explicit permissions; system flags require `platform.*` permission; all changes audit-logged + session invalidated |
| Entity scope bypass | `entity_scope` from authenticated session — never from request params; DB WHERE clause uses session's path prefix |
| Casbin rule injection | Validates subject prefix, domain ownership; cannot write `_platform_`; all additions audit-logged |
| Account takeover via reset | 32-byte token (256-bit entropy); 1h expiry; stored hashed; one-time use; all sessions invalidated on reset |

---

## 18. Performance & Caching

### 18.1 Hot Path Latency

```
Session validation:     ~1ms      indexed UPDATE on token_hash
Permission check:       <0.001ms  map lookup — nanoseconds
Flag check:             <0.001ms  map lookup
Setting read:           <0.001ms  map lookup + parse
Entity scope check:     0ms       struct field access
Casbin Enforce():       ~0.1ms    in-memory model evaluation (only for record-level checks)

Total auth+config overhead per request: ~1-2ms
```

### 18.2 Session Build at Login (One-Time Cost)

| Operation | Cost |
|---|---|
| ComputePermissions | ~2ms (role→permission JOIN) |
| ResolveEntityScope | ~1ms (entity path lookup) |
| FlagService.Resolve | ~1ms (LEFT JOIN flag tables) |
| SettingService.Resolve | ~1ms (LEFT JOIN setting tables) |
| UserPreferences | ~0.5ms |
| **Total at login** | **~5–6ms** |

All five queries run concurrently via `errgroup`. No further DB hits for auth, flags, or settings for the 24-hour session lifetime.

### 18.3 Cache Strategy

Repository layer owns all cache. Keys:

```
user:<uuid>                      5-minute TTL
session:<token_hash>             TTL = session.expires_at - now()
roles:<user_id>:<domain>         5-minute TTL; invalidated on role change
permissions:<user_id>            5-minute TTL; invalidated on sync
flags:tenant:<tenant_id>         10-minute TTL; invalidated on flag change
settings:tenant:<tenant_id>      10-minute TTL; invalidated on setting change
prefs:user:<user_id>             30-minute TTL; invalidated on preference update
entity:<uuid>                    10-minute TTL (changes rarely)
nav:tenant:<tenant_id>           5-minute TTL; invalidated on flag change
```

### 18.4 Casbin Performance

- Startup load: ~200ms for 10k rules
- In-memory `Enforce()`: p99 < 1ms under parallel load
- `InvalidateCache()`: ~200ms — call only after bulk imports, never in handlers

---

## 19. Common Business Scenarios

### Scenario 1: New Tenant — Module Configuration

```
1. Tenant is provisioned → OnTenantProvisioned seeds system roles
2. Platform operator enables Finance module:
   FlagService.Set({ TenantID: tenantID, FlagKey: "finance", Enabled: true })
   → Invalidates all tenant sessions (users must re-login to see Finance)
3. Platform operator configures settings:
   SettingService.UpdateModuleSettings({
     TenantID: tenantID,
     Settings: map[string]string{
       "finance.transactions.approval_threshold": "500000",
       "finance.budget_control_mode": "soft",
     }
   })
4. Next tenant admin login → session built with finance=true in flags
   → Finance appears in nav automatically
```

### Scenario 2: New Finance Manager (Full Company Access)

```go
platform.IAM.InviteUser(ctx, domain.UserInviteParams{
    Email:    "sarah@acme.com",
    UserType: domain.UserTypeTenant,
    RoleIDs:  []uuid.UUID{financeControllerRoleID},
    EntityID: companyRootEntityID,  // root → access everything
    // TenantID and CreatedBy from context
})
```

### Scenario 3: Regional Manager (Mombasa Branch Only)

```go
platform.IAM.InviteUser(ctx, domain.UserInviteParams{
    Email:    "ali@acme.com",
    UserType: domain.UserTypeTenant,
    RoleIDs:  []uuid.UUID{financeAccountantRoleID},
    EntityID: mombasaBranchEntityID,  // leaf → branch only
})
// EntityScope{Type: "entity"} → all queries filter to Mombasa Branch only
// No custom code in Finance service — handled at repo layer via scope type
```

### Scenario 4: Temporary Auditor

```go
platform.IAM.AssignRole(ctx, domain.AssignRoleParams{
    UserID:    auditorUserID,
    RoleID:    financeViewerRoleID,
    ExpiresAt: ptr(fiscalYearEnd),
    Reason:    "Annual audit engagement FY2025",
})
// Lazy expiry on first post-expiry request. No scheduler needed.
```

### Scenario 5: Employee Termination

```go
func (s *HRService) TerminateEmployee(ctx context.Context, params domain.TerminateParams) error {
    s.repo.SetTerminated(ctx, params.EmployeeID, params.Reason)
    platform.IAM.SuspendUser(ctx, domain.SuspendParams{
        UserID: params.UserID, Reason: "termination",
    }) // → InvalidateByUser() called inside; all sessions deleted immediately
    platform.IAM.AddDenyAll(ctx, domain.DenyAllParams{
        UserID: params.UserID, Reason: params.Reason,
    }) // → Casbin deny policy; even a surviving session gets blocked
    return nil
}
```

### Scenario 6: Enabling a New Feature Mid-Subscription

```
Tenant admin navigates to Settings → Modules
Sees "Airline Bookings" as a toggle (off)
Toggles it on:
  PATCH /api/v1/settings/flags/airline  { "enabled": true }
  → FlagService.Set fires
  → All tenant sessions invalidated (users must refresh)
  → Next login for any tenant user → session built with airline=true
  → If user has airline.*.read permission → Airline nav section appears
  → If user lacks permission → still no nav item (permission filter)
```

### Scenario 7: Portal Customer Viewing Their Invoices

```go
api.Get("/portal/invoices",
    middleware.RequireFlag("portal.self.invoices"),
    middleware.RequirePermission("portal.self.invoices", "read"),
    handlers.GetMyInvoices(deps))

func GetMyInvoices(deps *app.Deps) fiber.Handler {
    return func(c *fiber.Ctx) error {
        session := middleware.ContextSession(c)
        // PrincipalID from session — NEVER from request params
        invoices, _ := deps.InvoiceSvc.GetByCustomer(c.Context(),
            domain.GetInvoicesParams{CustomerID: session.PrincipalID})
        return c.JSON(response.OK(invoices))
    }
}
// Two independent protections:
// 1. Handler reads PrincipalID from session, not query params
// 2. DB RLS portal_self_isolation rejects customer_id != session.principal_id
```

---

## 20. Testing Strategy

### 20.1 Repository Tests (Next to Implementation)

```go
// internal/platform/repo/flag_repo_impl_test.go

func TestFlagRepo_ResolveForTenant_DefaultValue(t *testing.T) {
    repo := setupTestRepo(t) // testcontainers postgres

    tenantID := uuid.New()
    // No tenant_feature_flags row — should return default_value
    flags, err := repo.ResolveForTenant(ctx, tenantID)
    assert.NoError(t, err)
    // finance module is off by default
    assert.False(t, flags["finance"])
    // finance.transactions.approval_workflow is off by default
    assert.False(t, flags["finance.transactions.approval_workflow"])
}

func TestFlagRepo_ResolveForTenant_TenantOverride(t *testing.T) {
    repo := setupTestRepo(t)
    tenantID := uuid.New()

    // Enable finance for this tenant
    repo.SetFlag(ctx, domain.SetFlagParams{TenantID: tenantID, FlagKey: "finance", Enabled: true})

    flags, _ := repo.ResolveForTenant(ctx, tenantID)
    assert.True(t, flags["finance"])
    // resource flag inherits: finance.transactions default is true
    // but module check in FeatureEnabled would still gate it
    assert.True(t, flags["finance.transactions"])
}
```

### 20.2 Casbin Policy Tests

```go
func TestCasbin_DenyOverride(t *testing.T) {
    svc := setupTestCasbin(t)
    dom, sub := uuid.New().String(), "tenant:usr_test"
    svc.AddPolicy(ctx, domain.PolicyParams{Sub: "role:finance.viewer", Dom: dom,
        Obj: "finance/*", Act: "read", Eft: "allow"})
    svc.AddRoleForUser(ctx, sub, "role:finance.viewer", dom)

    ok, _ := svc.Enforce(ctx, domain.AuthzRequest{Sub: sub, Dom: dom, Obj: "finance/invoice/x", Act: "read"})
    assert.True(t, ok)

    svc.AddPolicy(ctx, domain.PolicyParams{Sub: sub, Dom: dom, Obj: "*", Act: "*", Eft: "deny"})
    ok, _ = svc.Enforce(ctx, domain.AuthzRequest{Sub: sub, Dom: dom, Obj: "finance/invoice/x", Act: "read"})
    assert.False(t, ok, "deny must override allow")
}

func TestCasbin_TenantIsolation(t *testing.T) {
    svc := setupTestCasbin(t)
    tenantA, tenantB := uuid.New(), uuid.New()
    svc.AddPolicy(ctx, domain.PolicyParams{Sub: "tenant:usr_A", Dom: tenantA.String(),
        Obj: "finance/invoice/*", Act: "read", Eft: "allow"})

    ok, _ := svc.Enforce(ctx, domain.AuthzRequest{Sub: "tenant:usr_A", Dom: tenantA.String(),
        Obj: "finance/invoice/inv_001", Act: "read"})
    assert.True(t, ok)

    ok, _ = svc.Enforce(ctx, domain.AuthzRequest{Sub: "tenant:usr_A", Dom: tenantB.String(),
        Obj: "finance/invoice/inv_001", Act: "read"})
    assert.False(t, ok, "domain isolation must prevent cross-tenant access")
}
```

### 20.3 Role Assignment Guard Tests

```go
func TestAssignRole_Guards(t *testing.T) {
    svc := setupTestIAM(t)

    t.Run("guard1_namespace", func(t *testing.T) {
        err := svc.AssignRole(ctx, domain.AssignRoleParams{UserID: tenantUserID, RoleID: platformRoleID})
        assert.ErrorIs(t, err, domain.ErrForbidden)
        assert.Contains(t, err.Error(), "assignable to")
    })
    t.Run("guard2_cross_tenant", func(t *testing.T) {
        err := svc.AssignRole(ctx, domain.AssignRoleParams{UserID: tenantAUserID, RoleID: tenantBRoleID})
        assert.ErrorIs(t, err, domain.ErrForbidden)
    })
    t.Run("guard3_delegation", func(t *testing.T) {
        err := svc.AssignRole(ctx, domain.AssignRoleParams{UserID: newUserID, RoleID: controllerRoleID,
            // GrantedBy = accountantUserID (from context) who doesn't hold controller
        })
        assert.ErrorIs(t, err, domain.ErrForbidden)
        assert.Contains(t, err.Error(), "you can only grant roles you hold")
    })
}
```

### 20.4 Entity Scope Tests

```go
func TestEntityScope_SubtreeAccess(t *testing.T) {
    companyID   := createEntity(t, "company", nil)
    nairobiID   := createEntity(t, "region",  &companyID)
    westlandsID := createEntity(t, "branch",  &nairobiID)

    companyTxn   := createTransaction(t, companyID)
    nairobiTxn   := createTransaction(t, nairobiID)
    westlandsTxn := createTransaction(t, westlandsID)

    ctx := contextWithSession(t, createUserWithEntity(t, nairobiID))
    txns, _ := financeRepo.ListTransactions(ctx, domain.TransactionListParams{})
    ids := transactionIDs(txns)

    assert.Contains(t,    ids, nairobiTxn.ID,   "own entity — accessible")
    assert.Contains(t,    ids, westlandsTxn.ID, "child entity — accessible")
    assert.NotContains(t, ids, companyTxn.ID,   "parent entity — not accessible")
}
```

### 20.5 Nav Generation Tests

```go
func TestBootService_BuildNav_FlagGating(t *testing.T) {
    svc := setupTestBoot(t)
    tenantID := createTenantWithFlags(t, map[string]bool{
        "finance":              true,
        "finance.transactions": true,
        "airline":              false,  // airline off
    })

    session := buildSessionForTenant(t, tenantID, map[string]bool{
        "finance.transactions.read": true,
        "airline.bookings.read":     true,  // has permission but flag is off
    })

    shell, _ := svc.BuildAppShell(ctx, session)
    pages := shell["pages"].([]any)
    labels := extractNavLabels(pages)

    assert.Contains(t,    labels, "Finance")  // flag on + permission = visible
    assert.NotContains(t, labels, "Airline")  // flag off = not visible regardless of permission
}
```

### 20.6 Benchmarks

```go
func BenchmarkSessionCan(b *testing.B) {
    session := buildTestSession(b, 200)
    b.ResetTimer()
    for i := 0; i < b.N; i++ { session.Can("finance.transactions", "post") }
    // Expected: < 100ns
}

func BenchmarkSessionFeatureEnabled(b *testing.B) {
    session := buildTestSession(b, 200)
    b.ResetTimer()
    for i := 0; i < b.N; i++ { session.FeatureEnabled("finance.transactions.approval_workflow") }
    // Expected: < 200ns (two map lookups)
}

func BenchmarkSessionBuildAtLogin(b *testing.B) {
    // Login is the expensive step — benchmark to catch regressions
    for i := 0; i < b.N; i++ { iamSvc.buildSession(ctx, testUser) }
    // Expected: < 10ms (five concurrent queries)
}
```

---

## 21. API Reference

### 21.1 Core Types

```go
// Permission check
func (s *ResolvedSession) Can(resource, action string) bool

// Feature flag check (checks module flag too if key has dot)
func (s *ResolvedSession) FeatureEnabled(key string) bool

// Setting accessors — typed, with default
func (s *ResolvedSession) SettingBool(key string, def bool) bool
func (s *ResolvedSession) SettingDecimal(key string, def decimal.Decimal) decimal.Decimal
func (s *ResolvedSession) SettingInt(key string, def int) int
func (s *ResolvedSession) SettingString(key string, def string) string

// User preference accessor
func (s *ResolvedSession) Pref(key string, def string) string

// Identity helpers
func (s *ResolvedSession) IsPlatform() bool
func (s *ResolvedSession) IsPortal()   bool

// Casbin request
type AuthzRequest struct {
    Subject, Domain, Object, Action string
}

// Domain constructors
func TenantDomain(id uuid.UUID) string   { return id.String() }
func TenantSubject(id uuid.UUID) string  { return "tenant:" + id.String() }
func PlatformSubject(id uuid.UUID) string { return "platform:" + id.String() }
func PortalSubject(id uuid.UUID) string  { return "portal:" + id.String() }

// Entity scope
type EntityScope struct {
    Type       string    // "all" | "subtree" | "entity"
    EntityID   uuid.UUID
    PathPrefix string
}

// Struct param example
type AssignRoleParams struct {
    UserID    uuid.UUID
    RoleID    uuid.UUID
    ExpiresAt *time.Time  // nil = permanent
    Reason    string
    // GrantedBy from context
}
```

### 21.2 Error Types

```go
var (
    ErrUnauthenticated      = domain.NewError("UNAUTHENTICATED",      401)
    ErrForbidden            = domain.NewError("FORBIDDEN",            403)
    ErrFeatureDisabled      = domain.NewError("FEATURE_DISABLED",     403)
    ErrMFARequired          = domain.NewError("MFA_REQUIRED",         200)
    ErrInvalidMFA           = domain.NewError("INVALID_MFA",          422)
    ErrMFACodeReplayed      = domain.NewError("MFA_REPLAYED",         422)
    ErrAccountLocked        = domain.NewError("ACCOUNT_LOCKED",       429)
    ErrAccountSuspended     = domain.NewError("ACCOUNT_SUSPENDED",    403)
    ErrSessionExpired       = domain.NewError("SESSION_EXPIRED",      401)
    ErrInvalidCredentials   = domain.NewError("INVALID_CREDENTIALS",  422)
    ErrTokenExpired         = domain.NewError("TOKEN_EXPIRED",        410)
    ErrMissingTenantContext = domain.NewError("MISSING_TENANT_CONTEXT", 400)
)
```

---

## 22. Troubleshooting

### 22.1 Module/Feature Not Appearing in Nav

```sql
-- 1. Verify module flag is on for this tenant
SELECT ffd.flag_key, COALESCE(tff.enabled, ffd.default_value) AS effective
FROM feature_flag_definitions ffd
LEFT JOIN tenant_feature_flags tff ON tff.flag_id = ffd.id AND tff.tenant_id = '<tenant_uuid>'
WHERE ffd.flag_key = 'finance';

-- 2. Verify user has read permission for the resource
SELECT permissions->>'finance.transactions.read' FROM sessions
WHERE user_id = '<user_uuid>' AND expires_at > NOW()
ORDER BY created_at DESC LIMIT 1;

-- 3. If flag is on but permission is missing — check role
SELECT r.slug, rp_p.full_key
FROM role_assignments ra
JOIN roles r ON ra.role_id = r.id
JOIN role_permissions rp ON rp.role_id = r.id
JOIN permissions rp_p ON rp.permission_id = rp_p.id
WHERE ra.user_id = '<user_uuid>' AND rp_p.full_key LIKE 'finance.%';
```

### 22.2 Setting Change Not Taking Effect

```
Cause: sessions.configuration JSONB was built at login. Settings changes
       require session refresh.

Fix:
  Option A: User logs out and back in
  Option B (for non-sensitive settings): wait for session expiry (24h max)
  Option C (immediate): invalidate tenant sessions
    DELETE FROM sessions WHERE tenant_id = '<tenant_uuid>';
    -- SettingService.UpdateModuleSettings does this automatically for significant settings
```

### 22.3 User Gets 403 on a Route They Should Access

```sql
-- 1. Check session permissions
SELECT permissions->>'finance.transactions.post' FROM sessions
WHERE user_id = '<uuid>' AND expires_at > NOW()
ORDER BY created_at DESC LIMIT 1;
-- NULL or false → stale session or role not assigned

-- 2. Check active deny policies
SELECT * FROM casbin_rule WHERE ptype = 'p' AND v4 = 'deny'
  AND v0 IN ('tenant:<user_uuid>', 'tenant:*') AND v1 = '<tenant_uuid>';

-- 3. Check role assignment
SELECT r.slug, ra.expires_at, ra.revoked_at
FROM role_assignments ra JOIN roles r ON ra.role_id = r.id
WHERE ra.user_id = '<uuid>' AND ra.revoked_at IS NULL
  AND (ra.expires_at IS NULL OR ra.expires_at > NOW());
```

### 22.4 Common Error Reference

| Error | Cause | Fix |
|---|---|---|
| `403 FEATURE_DISABLED` | Flag is off for tenant | Enable flag in Settings → Modules |
| `403 permission denied` | User lacks permission | Assign appropriate role |
| `400 MISSING_TENANT_CONTEXT` | No X-Tenant-ID header | Include header in all non-platform requests |
| `403 role not assignable to tenant users` | Guard 1 | Use role with `assignable_to = ['tenant']` |
| `403 you can only grant roles you hold` | Guard 3 | Granter must hold the role |
| `422 INVALID_MFA` | Wrong TOTP code | Check device clock sync; use backup code |
| `429 ACCOUNT_LOCKED` | 5+ failures | Wait for lockout expiry or admin unlock |
| Module missing from nav | Flag off or no read permission | Check both flag and permission |
| Setting not applying | Stale session | Logout and login again |
| Context missing `tenant_id` | Middleware not on route | Ensure route is in authenticated group |
