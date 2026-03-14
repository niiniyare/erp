# Security & Architecture Analysis — AWO ERP
**Scope**: Tenant Isolation · Identity · Authorization · Access Control
**Date**: 2026-03-15
**Based on**: Direct source-code and migration-file review

---

## Executive Summary

The AWO ERP codebase implements a **well-engineered multi-layer security model**. The
authorization engine is a fully working Casbin-backed RBAC system with domain
isolation, deny-override semantics, and lazy temporal role expiry — all backed by
integration tests. The tenant context system is extensively documented with explicit
Architecture Decision Records (ADR-012 through ADR-018) justifying every design
choice. Fail-closed RLS, transaction-local session variables, and SECURITY DEFINER
search-path pinning reflect real production awareness.

Several **real gaps** remain and are individually assessed below. The most important
are: authentication post-processing TODOs (brute-force protection, last-login
tracking), an atomicity hazard in role assignment, unreachable ABAC infrastructure,
and an unsafe RLS expression on `policy_evaluations`.

The previous version of this report contained multiple inaccuracies (wrong DB role
names, wrongly claiming the authorization engine was unimplemented). This version is
based solely on direct file reads.

---

## 1. Tenant Model & Isolation

### 1.1 DB Roles

**File**: `db/migration/000051_tenant_extensions_and_roles.up.sql`

Three roles, not two, and **none uses `BYPASSRLS`** attribute:

| Role | Access | Note |
|------|--------|------|
| `application_role` | DML on all public tables, governed by RLS | Runtime API credential |
| `admin_role` | Full via explicit `USING(TRUE)` policies | Ops/migration only |
| `readonly_role` | SELECT on active rows only | Reporting replicas |

**ADR-018** documents the explicit choice of policy-based bypass over `BYPASSRLS`:
> "The explicit policy is visible in pg_policies and therefore auditable. BYPASSRLS is
> invisible in standard role inspection and easier to accidentally grant broadly."

### 1.2 Tenant Context Functions

**File**: `db/migration/000055_tenant_functions.up.sql`

Four functions, all with explicit security decisions:

```sql
-- SECURITY INVOKER — reads only session variable, no table access needed
CREATE OR REPLACE FUNCTION current_tenant_id() RETURNS UUID
  LANGUAGE plpgsql STABLE SECURITY INVOKER
AS $$
BEGIN
  RETURN COALESCE(NULLIF(current_setting('app.current_tenant_id', TRUE), ''), NULL)::UUID;
EXCEPTION
  WHEN invalid_text_representation THEN
    RAISE WARNING '...'; RETURN NULL;  -- fail-closed, not error
END; $$;

-- SECURITY DEFINER with pinned search_path (ADR-014: search-path injection guard)
CREATE OR REPLACE FUNCTION set_tenant_context(p_tenant_id UUID) RETURNS VOID
  LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog, public
AS $$
DECLARE v_status TEXT;
BEGIN
  SELECT "Status" INTO v_status FROM tenants
   WHERE id = p_tenant_id AND deleted_at IS NULL;   -- existence + soft-delete
  IF NOT FOUND THEN RAISE EXCEPTION '...' USING ERRCODE = 'no_data_found'; END IF;
  IF v_status <> 'ACTIVE' THEN RAISE EXCEPTION '...' USING ERRCODE = 'check_violation'; END IF;
  PERFORM set_config('app.current_tenant_id', p_tenant_id::TEXT, TRUE);  -- transaction-local (ADR-013)
END; $$;
```

Key ADRs (all documented in migration header):
- **ADR-012**: Single-argument only; dead `user_role` param removed
- **ADR-013**: Transaction-local only (`set_config(..., TRUE)`) — cleared on COMMIT/ROLLBACK, safe for pooled connections
- **ADR-014**: SECURITY DEFINER functions pin `search_path = pg_catalog, public` to prevent schema-injection attacks
- **ADR-015**: Only `app.current_tenant_id` is used; old `app.tenant_status`, `app.context_set_at` removed

### 1.3 Row-Level Security Policies

**File**: `db/migration/000056_tenant_rls.up.sql`

Three policies on `tenants`, each documented:

```sql
-- Fail-closed (ADR-016): zero rows when no context, never cross-tenant leakage
CREATE POLICY tenant_isolation_policy ON tenants FOR ALL TO application_role
  USING (id = current_tenant_id() AND deleted_at IS NULL)
  WITH CHECK (id = current_tenant_id() AND deleted_at IS NULL);

-- Admin: full access including soft-deleted rows (ADR-018)
CREATE POLICY admin_full_access_policy ON tenants FOR ALL TO admin_role
  USING (TRUE) WITH CHECK (TRUE);

-- Reporting: SELECT only on live rows
CREATE POLICY readonly_access_policy ON tenants FOR SELECT TO readonly_role
  USING (deleted_at IS NULL);
```

All data tables (users, roles, permissions, role_assignments, policies, etc.) follow
the same pattern with `current_tenant_id() IS NOT NULL AND tenant_id =
current_tenant_id()`. Soft-deleted rows are excluded at policy level as defense-in-
depth (ADR-017).

### 1.4 Cross-Tenant Trigger

**File**: `db/migration/000107_enforce_tenant_isolation.up.sql`

A `BEFORE INSERT/UPDATE` trigger `enforce_tenant_isolation()` validates that foreign
key references (`entity_id`, etc.) belong to the same tenant. Currently only wired for
the `persons` table — the `ELSE RAISE NOTICE` branch suggests other tables are planned
but not yet connected.

**Finding T1 — [LOW] Cross-tenant trigger incomplete**
Only `persons` is validated. Tables like `employees`, `users` that reference `entity_id`
do not yet have the trigger attached. The RLS policies still prevent cross-tenant data
access, so this is defense-in-depth only.

### 1.5 Status Inconsistency Between Context Functions

**File**: `db/migration/000106_validate_set_tenant_context.up.sql`

The older `validate_and_set_tenant_context()` function (migration 106) accepts both
`'active'` and `'pending'` tenant statuses, and uses lowercase comparison:

```sql
IF v_tenant_record.status NOT IN ('active', 'pending') THEN
  RAISE EXCEPTION 'Tenant is not active...';
END IF;
```

The canonical `set_tenant_context()` (migration 055) uses uppercase `'ACTIVE'` only:

```sql
IF v_status <> 'ACTIVE' THEN RAISE EXCEPTION '...'; END IF;
```

**Finding T2 — [MEDIUM] Inconsistent tenant status validation**
If application code calls `validate_and_set_tenant_context()` instead of
`set_tenant_context()`, a `PENDING` tenant can be granted context — bypassing the
intended lifecycle gate. Recommendation: deprecate `validate_and_set_tenant_context()`
and standardise all callers on `set_tenant_context()`.

---

## 2. Identity & Authentication

### 2.1 User Schema

**File**: `db/migration/000303_identity_create_users.up.sql`

```sql
CREATE TABLE users (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id           UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id           UUID NOT NULL REFERENCES entities(uuid) ON DELETE RESTRICT,
  person_id           UUID REFERENCES persons(id) ON DELETE SET NULL,
  employee_id         UUID REFERENCES employees(id) ON DELETE SET NULL,
  email               VARCHAR(255) NOT NULL,
  username            VARCHAR(100) NOT NULL,
  password_hash       VARCHAR(255),               -- bcrypt
  user_type           VARCHAR(20) NOT NULL DEFAULT 'INTERNAL'
                      CHECK (user_type IN ('INTERNAL','CUSTOMER','VENDOR','PARTNER','SYSADMIN')),
  account_status      VARCHAR(20) DEFAULT 'ACTIVE'
                      CHECK (account_status IN ('ACTIVE','INACTIVE','LOCKED','SUSPENDED','PENDING_VERIFICATION')),
  is_active           BOOLEAN NOT NULL DEFAULT TRUE,
  last_login_at       TIMESTAMPTZ,                -- ⚠ never updated — see SEC-1
  failed_login_attempts INTEGER DEFAULT 0,        -- ⚠ never incremented — see SEC-1
  lockout_until       TIMESTAMPTZ,                -- ⚠ never set — see SEC-1
  session_timeout_minutes INTEGER DEFAULT 480,    -- 8 hours
  mfa_enabled         BOOLEAN DEFAULT false,
  mfa_secret          VARCHAR(255),               -- ⚠ not enforced — see SEC-2
  password_strength   INT DEFAULT 0,
  compromised         BOOLEAN DEFAULT false,
  rotation_required   BOOLEAN DEFAULT false,
  ...
);

-- Unique email/username per live tenant row (correct use of EXCLUDE)
CONSTRAINT users_email_unique_active EXCLUDE (tenant_id WITH =, email WITH =) WHERE (deleted_at IS NULL)
CONSTRAINT valid_lockout_time CHECK (lockout_until IS NULL OR lockout_until > NOW())
```

RLS policy is fail-closed:
```sql
CREATE POLICY users_tenant_isolation ON users FOR ALL TO application_role
  USING (current_tenant_id() IS NOT NULL AND deleted_at IS NULL AND tenant_id = current_tenant_id())
  WITH CHECK (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());
```

### 2.2 Session Schema

**File**: `db/migration/000304_identity_add_user_sessions.up.sql`

```sql
CREATE TABLE user_sessions (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  session_token   VARCHAR(255) UNIQUE NOT NULL,
  refresh_token   VARCHAR(255),
  ip_address      INET,
  user_agent      TEXT,
  device_info     JSONB DEFAULT '{}',   -- device fingerprinting for ABAC
  location_info   JSONB DEFAULT '{}',   -- geo data for location-based policies
  expires_at      TIMESTAMPTZ NOT NULL,
  risk_score      INT DEFAULT 0,
  last_accessed_at TIMESTAMPTZ DEFAULT NOW(),
  is_active       BOOLEAN DEFAULT TRUE
);
CONSTRAINT valid_expiration_time CHECK (expires_at > created_at)
```

The schema is well-designed for session-based auth with ABAC context data. RLS is
tenant-scoped. Session token and refresh token are plain `VARCHAR(255)` — this means
they are **stored in plaintext**. If the DB is compromised, all live sessions are
immediately hijackable.

**Finding A1 — [MEDIUM] Session tokens stored in plaintext**
`session_token` and `refresh_token` are stored as-is. Best practice is to store only a
hash (SHA-256) of the token and compare the hash on lookup. This limits DB-compromise
blast radius.
**Location**: `db/migration/000304_identity_add_user_sessions.up.sql:17-18`

### 2.3 Authentication Service (Known TODOs)

**File**: `internal/core/identity/service.go` lines ~204–210

```go
if err := bcrypt.CompareHashAndPassword(...); err != nil {
    // TODO: Increment failed login attempts     ← brute-force protection absent
    return nil, ErrAuthenticationFailed
}
// TODO: Update last login time                  ← audit trail incomplete
```

**Finding SEC-1 — [CRITICAL] Brute-force protection non-functional**

- `failed_login_attempts` is never incremented
- `lockout_until` is never set
- `account_status` is never set to `LOCKED` programmatically
- `last_login_at` is never updated

The schema and DB constraint (`valid_lockout_time`) are correctly designed — the
application service just hasn't wired them up. A dictionary attack can proceed
indefinitely.

**Fix**: Implement the TODO items:
1. On auth failure: `UPDATE users SET failed_login_attempts = failed_login_attempts + 1 WHERE id = $1`
2. After N failures (e.g. 5): `UPDATE users SET lockout_until = NOW() + INTERVAL '15 minutes', account_status = 'LOCKED'`
3. On auth success: `UPDATE users SET failed_login_attempts = 0, last_login_at = NOW()`
4. At the start of Authenticate: check `lockout_until > NOW()` and return `ErrAccountLocked`

**Finding SEC-2 — [HIGH] MFA not enforced in authentication flow**

`mfa_enabled` and `mfa_secret` fields exist and are properly indexed
(`idx_users_mfa`). The service model has `MfaEnabled bool`. However, the
`Authenticate()` flow does not branch on MFA. Users with `mfa_enabled = true` receive
a session without completing a TOTP challenge.

**Fix**: After successful password check, if `user.MfaEnabled`, return an intermediate
state requiring TOTP verification before session creation.

---

## 3. Authorization Architecture

### 3.1 Engine: Casbin RBAC with Domain Isolation

**Files**: `internal/core/authz/model.go`, `service.go`, `adapter.go`

The authorization engine is **Casbin v2** with a PostgreSQL-backed adapter and a
domain-scoped RBAC model:

```
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act, eft

[role_definition]
g = _, _, _                        ← domain-scoped g(user, role, domain)

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))   ← deny-override

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom
    && keyMatch2(r.obj, p.obj)     ← wildcard: "invoice/*"
    && keyMatch2(r.act, p.act)     ← wildcard: "*"
```

**Properties:**
- **Deny-override**: one explicit `deny` beats any number of `allow` rules
- **Domain scoping**: `g(user, role, domain)` means roles are per-tenant — a role in
  `tenantA` has no effect in `tenantB`
- **Wildcard object/action**: `invoice/*` or `*` patterns via `keyMatch2`
- **Default deny**: if no matching policy, access is denied

### 3.2 Subject and Domain Naming

**File**: `internal/core/authz/types.go`

```go
// Subject prefixes — namespaced to prevent collisions
func PlatformSubject(userID string) string { return "platform:" + userID }
func TenantSubject(userID string) string   { return "tenant:" + userID }
func PortalSubject(userID string) string   { return "portal:" + userID }
func APISubject(clientID string) string    { return "api:" + clientID }

// Domain construction
func TenantDomain(id string) string { return id }             // tenant UUID
func PortalDomain(id string) string { return id + ":portal" }
func APIDomain(id string) string    { return id + ":api" }

const DomainPlatform = "_platform_"  // reserved for platform-level policies
```

Cross-tenant policies are structurally impossible in the RBAC layer: a user's subject
is `tenant:uuid` and policies are scoped to a domain (tenant UUID). A `tenant:userA`
subject in domain `tenantA` cannot match policies in domain `tenantB`.

**Tests confirm this** (`service_test.go:215`):
```go
func (s *ServiceSuite) TestEnforce_PolicyDoesNotCrossDomains() {
    // Policy exists in dom1, user evaluated in dom2 → deny
    ok, _ := s.svc.Enforce(s.ctx, Request{Subject: testSubject, Domain: dom2, ...})
    s.False(ok, "policy in dom1 must not bleed into dom2")
}
```

### 3.3 Role Assignment (roles.go)

```go
func (s *service) AssignRole(ctx, tenantID, subject, role, domain string, opts ...AssignOpt) error {
    // 1. Write metadata to role_assignments table (with UPSERT)
    tx, _ := s.pool.Begin(ctx)
    tx.Exec(ctx, `INSERT INTO role_assignments (...) ON CONFLICT ... DO UPDATE SET is_active=TRUE...`)
    tx.Commit(ctx)
    // 2. Add Casbin g-rule (in-memory + DB via adapter)
    s.enforcer.AddRoleForUserInDomain(subject, role, domain)
}
```

**Finding Z1 — [MEDIUM] Atomicity hazard in role assignment**

`AssignRole` first commits the DB transaction, then calls
`s.enforcer.AddRoleForUserInDomain()`. If the Casbin call fails (e.g., network timeout
to DB adapter) after the transaction commit, the `role_assignments` table shows
`is_active=TRUE` but Casbin's in-memory model does not have the g-rule. The user
cannot use the role until `InvalidateCache()` is called.

**Location**: `internal/core/authz/roles.go:42–49`

```go
if err := tx.Commit(ctx); err != nil { return ... }
// ↑ committed — no rollback possible below
if _, err := s.enforcer.AddRoleForUserInDomain(subject, role, domain); err != nil {
    return fmt.Errorf("authz AssignRole casbin: %w", err)  // DB and Casbin are now diverged
}
```

**Fix**: Wrap the entire operation including the Casbin call in a single DB transaction, or
reverse the order (Casbin first, then DB commit), or add explicit reconciliation on startup.

### 3.4 Lazy Role Expiry

**File**: `internal/core/authz/roles.go:117–151`

```go
func (s *service) revokeExpiredRoles(ctx, subject, domain string) error {
    rows, _ := s.pool.Query(ctx, `
        SELECT role_name FROM role_assignments
        WHERE subject=$1 AND domain=$2
          AND is_active = TRUE
          AND expires_at IS NOT NULL
          AND expires_at < NOW()`, subject, domain)
    // ... for each expired: s.RevokeRole(ctx, subject, role, domain)
}
```

Called at the beginning of every `Enforce()` call. Non-fatal: if cleanup fails, the
check still proceeds (logged as warning).

**Test confirms** (`roles_test.go:320`):
```go
func (s *RolesSuite) TestEnforce_ExpiredRole_IsRevoked() {
    past := time.Now().UTC().Add(-1 * time.Hour)
    svc.AssignRole(..., WithExpiry(past))
    ok, _ := svc.Enforce(...)
    s.False(ok, "expired role must cause deny")
    has, _ := svc.HasRole(...)
    s.False(has, "role must be gone from Casbin after lazy revoke")
}
```

This is correctly implemented. The revocation is also confirmed in the DB
(`is_active=FALSE`). Audit trail is preserved.

### 3.5 Middleware

**File**: `internal/core/authz/middleware.go`

```go
func (s *service) Middleware(object, action string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        p, ok := c.Locals(LocalsKeyPrincipal).(Principal)
        if !ok || p.Subject == "" {
            return fiber.NewError(fiber.StatusUnauthorized, ErrUnauthorized.Error())
        }
        obj := object
        if id := c.Params("id"); id != "" {
            obj = object + "/" + id   // per-resource: "invoice/123"
        }
        allowed, err := s.Enforce(c.Context(), Request{
            Subject: p.Subject, Domain: p.Domain, Object: obj, Action: action,
        })
        if !allowed { return fiber.NewError(fiber.StatusForbidden, ...) }
        return c.Next()
    }
}
```

**Usage pattern** (from doc comments):
```go
app.Get("/invoices",        svc.Middleware("invoice", "read"),   listInvoices)
app.Post("/invoices",       svc.Middleware("invoice", "create"), createInvoice)
app.Delete("/invoices/:id", svc.Middleware("invoice", "delete"), deleteInvoice)
```

The middleware depends on an upstream authentication middleware setting
`c.Locals(authz.LocalsKeyPrincipal)` (`"authz_principal"`). The authentication
middleware is **not in this package** — it must be composed correctly at the router
level.

**Finding AC1 — [HIGH] Authentication middleware coupling is implicit**
There is no compile-time guarantee that `authz_principal` is set before
`authz.Middleware()` runs. If routes are wired incorrectly (authz middleware before
authn middleware), `c.Locals(LocalsKeyPrincipal)` returns nil and every request gets
401 — which is at least fail-safe, but indicates a fragile coupling.

**Fix**: Consider a type-safe accessor or a compile-time check. At minimum, add an
integration test that verifies correct middleware ordering.

---

## 4. Database Schema for Authorization

### 4.1 casbin_rule Table

**File**: `db/migration/000063_authz_casbin_rule.up.sql`

```sql
CREATE TABLE casbin_rule (
  id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  ptype VARCHAR(10) NOT NULL,
  v0    VARCHAR(256) NOT NULL DEFAULT '',
  ...v5 VARCHAR(256) NOT NULL DEFAULT ''
);
CREATE UNIQUE INDEX idx_casbin_rule_unique ON casbin_rule(ptype, v0, v1, v2, v3, v4, v5);

-- RLS enabled but policies allow FULL access to application_role
ALTER TABLE casbin_rule ENABLE ROW LEVEL SECURITY;
CREATE POLICY casbin_rule_app   ON casbin_rule FOR ALL TO application_role USING (TRUE) WITH CHECK (TRUE);
CREATE POLICY casbin_rule_admin ON casbin_rule FOR ALL TO admin_role        USING (TRUE) WITH CHECK (TRUE);
```

The migration comment explicitly states:
> "No tenant-level RLS: domain value in v1 enforces isolation at app level"

**Finding SEC-3 — [MEDIUM] casbin_rule has no row-level tenant isolation**

Any `application_role` connection can read and write ALL Casbin rules for ALL tenants.
Tenant isolation relies entirely on the application correctly constructing domain-scoped
requests. An application-layer bug that:
- writes a policy with the wrong domain
- reads policies for the wrong domain
- manipulates g-rules across domain boundaries

...would result in cross-tenant privilege escalation with no DB-level safety net.

The choice is documented and intentional. The mitigation is rigorous application-layer
testing and ensuring all policy mutations go through `authz.Service` methods (not raw
SQL).

**Risk**: Medium — requires an application bug, not just a DB access.

### 4.2 role_assignments Table

**File**: `db/migration/000064_authz_role_assignments.up.sql`

```sql
CREATE TABLE role_assignments (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id    UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  subject      VARCHAR(256) NOT NULL,
  role_name    VARCHAR(100) NOT NULL,
  domain       VARCHAR(256) NOT NULL,
  assigned_by  VARCHAR(256),
  delegated_by VARCHAR(256),
  expires_at   TIMESTAMPTZ,
  is_active    BOOLEAN DEFAULT TRUE,
  CONSTRAINT role_assignments_unique UNIQUE (subject, role_name, domain)
);

-- RLS: tenant-scoped for application_role
CREATE POLICY ra_tenant ON role_assignments FOR ALL TO application_role
  USING (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id())
  WITH CHECK (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());
```

Role assignments are tenant-isolated at the DB level (unlike `casbin_rule`). The audit
trail is preserved: revocation sets `is_active=FALSE` and never DELETEs rows.

### 4.3 ABAC Infrastructure (DB Only)

Several migrations create ABAC infrastructure:

| Table | Migration | Purpose |
|-------|-----------|---------|
| `policies` | 000408 | ABAC/RBAC/HYBRID/TIME_BASED/LOCATION_BASED policies |
| `permissions` | 000404 | Resource+action combos with ABAC conditions, data_filters, field_restrictions |
| `policy_evaluations` | 000703 | Cache of ABAC evaluation results |
| `policy_decisions` | 000705 | Per-policy decisions in an evaluation |

These tables are fully defined in the DB. However, **no application-layer Go code was
found that evaluates these ABAC policies**. The `internal/core/authz/` package uses
only Casbin (`casbin_rule`) and does not read `policies` or `permissions`.

**Finding SEC-4 — [HIGH] ABAC policy infrastructure exists in DB but has no evaluator**

The `policies` table supports `conditions JSONB`, `data_filters JSONB`,
`field_restrictions JSONB`. The `policy_evaluations` table caches results with 1-hour
TTL. None of this is evaluated by Go application code. Any ABAC policies created
through the admin interface are silently ignored.

Additionally, **`policy_evaluations` has an unsafe RLS expression**:

```sql
-- FROM migration 000703 — BUG
CREATE POLICY policy_evaluations_tenant_isolation ON policy_evaluations FOR ALL TO public
  USING (tenant_id = current_setting('app.current_tenant_id')::UUID);  -- ← unsafe
```

This uses `current_setting()` directly without the safe `current_tenant_id()` wrapper
function. If `app.current_tenant_id` is not set, PostgreSQL **raises an exception**
(not NULL), crashing the query. The safe version `current_setting('app.current_tenant_id', TRUE)`
(with `true` = missing OK) returns NULL instead. Also, `TO public` grants to all roles
including unauthenticated connections.

**Fix**:
```sql
-- Replace with safe pattern
USING (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id())
-- And change TO public → TO application_role
```

---

## 5. Roles Table vs Casbin

**File**: `db/migration/000405_auth_create_roles.up.sql`

```sql
CREATE TABLE roles (
  id             UUID PRIMARY KEY,
  tenant_id      UUID NOT NULL,
  entity_id      UUID NOT NULL,
  name           VARCHAR(50) NOT NULL,
  parent_role_id UUID REFERENCES roles(id),   -- role hierarchy
  level          INTEGER DEFAULT 0,
  permissions    JSONB NOT NULL DEFAULT '{}', -- cached permissions (perf)
  entity_scope   JSONB,
  conditions     JSONB,                       -- time/location/device conditions
  is_system_role BOOLEAN DEFAULT false,
  ...
);
```

This table models roles with hierarchy (`parent_role_id`, `level`) and ABAC conditions.
The Casbin model (`model.go`) uses **only a single `g` relation** (no `g2` for role
hierarchy). Casbin is not aware of `parent_role_id`.

**Finding SEC-5 — [MEDIUM] Role hierarchy in DB not reflected in Casbin**

Permissions inherited through `parent_role_id` in the `roles` table are not
automatically propagated in Casbin. An admin who defines a role hierarchy expecting
child roles to inherit parent permissions will find this does not work at enforcement
time. This can lead to unexpected access denials (weaker than expected) or — if
someone manually grants permissions they believe are inherited — over-permissioning
when the parent role is later restricted.

---

## 6. Risk Assessment Matrix

| ID | Finding | Location | Impact | Likelihood | Priority |
|----|---------|----------|--------|------------|----------|
| SEC-1 | Brute-force protection not wired | `identity/service.go:204` | Critical | High | **P0** |
| SEC-2 | MFA stored but not enforced | `identity/service.go`, `users:mfa_*` | High | High | **P0** |
| SEC-4 | ABAC policies silently ignored; `policy_evaluations` unsafe RLS | `000703_policy_evaluation.up.sql:32` | High | High | **P1** |
| AC1 | authn/authz middleware coupling is implicit | `authz/middleware.go:20` | High | Low | **P1** |
| Z1 | Role assignment atomicity hazard | `authz/roles.go:42` | Medium | Low | **P2** |
| T2 | Inconsistent tenant status validation | `000106_validate_set_tenant_context.up.sql:34` | Medium | Low | **P2** |
| A1 | Session tokens stored in plaintext | `000304_identity_add_user_sessions.up.sql:17` | Medium | Low | **P2** |
| SEC-3 | casbin_rule no row-level isolation | `000063_authz_casbin_rule.up.sql:17` | Medium | Low | **P2** |
| SEC-5 | Role hierarchy not propagated to Casbin | `000405_auth_create_roles.up.sql:13` | Medium | Medium | **P2** |
| T1 | Cross-tenant trigger only wired to `persons` | `000107_enforce_tenant_isolation.up.sql` | Low | Low | **P3** |

---

## 7. Strengths

| Strength | Evidence |
|----------|----------|
| Casbin deny-override RBAC with domain isolation | `authz/model.go` — `e = some(allow) && !some(deny)` |
| Casbin in-memory policy tested against domain bleed | `service_test.go:215` — `TestEnforce_PolicyDoesNotCrossDomains` |
| Lazy temporal role expiry with DB audit trail | `authz/roles.go:117`, `roles_test.go:320` |
| Fail-closed RLS (ADR-016) | `000056_tenant_rls.up.sql:6` |
| Transaction-local tenant context (ADR-013) | `000055_tenant_functions.up.sql:13` |
| SECURITY DEFINER search_path pinned (ADR-014) | `000055_tenant_functions.up.sql:97` |
| admin_role uses explicit USING(TRUE) policy, not BYPASSRLS (ADR-018) | `000056_tenant_rls.up.sql:18` |
| Soft-delete in RLS policy itself (ADR-017) | `000056_tenant_rls.up.sql:12` |
| Role revocation preserves audit row (`is_active=FALSE`) | `authz/roles.go:56`, `roles_test.go:165` |
| Session risk_score + device/location JSONB for future ABAC | `000304_identity_add_user_sessions.up.sql:24` |
| EXCLUDE constraint prevents email/username collisions per tenant | `000303_identity_create_users.up.sql:64` |
| DB-level lockout constraint (`valid_lockout_time`) in place | `000303_identity_create_users.up.sql:207` |
| Integration test suite with DB-backed enforcer tests | `service_test.go`, `roles_test.go` |

---

## 8. Remediation Roadmap

### Immediate (P0 — fix before any user-facing deployment)

**1. Wire brute-force protection** (`internal/core/identity/service.go`)

Three DB operations needed — all fields and constraints already exist:
- On failure: increment `failed_login_attempts`
- After 5 failures: set `lockout_until = NOW() + '15 minutes'`, `account_status = 'LOCKED'`
- On success: reset `failed_login_attempts = 0`, set `last_login_at = NOW()`
- At login start: check `lockout_until > NOW()` → return `ErrAccountLocked`

**2. Enforce MFA when `mfa_enabled = true`**

After password verification, check `user.MfaEnabled`. If true, issue an intermediate
`MFA_REQUIRED` token (not a full session token), require TOTP validation against
`mfa_secret`, then issue the full session.

### Short-term (P1 — within 2 weeks)

**3. Fix `policy_evaluations` RLS expression**

```sql
-- db/migration (new up migration)
ALTER POLICY policy_evaluations_tenant_isolation
  ON policy_evaluations TO application_role
  USING (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id())
  WITH CHECK (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());
```

**4. Verify and test middleware ordering**

Write an integration test that:
- Calls an authn-gated route without setting `authz_principal` → assert 401
- Calls an authz-gated route with a principal but no matching policy → assert 403
- Calls with correct credentials and policy → assert 200

**5. Decide on `validate_and_set_tenant_context()` deprecation** (`000106`)

Either: add a migration to `DROP FUNCTION validate_and_set_tenant_context`, or align
its status check with the canonical `set_tenant_context`.

### Medium-term (P2 — within 1 month)

**6. Fix role assignment atomicity** (`authz/roles.go:42`)

Options in order of preference:
- Call `s.enforcer.AddRoleForUserInDomain()` *before* committing, roll back if it fails
- Or: detect in-memory/DB divergence on `Enforce()` using the existing `InvalidateCache` pattern

**7. Hash session tokens before storage**

Store `sha256(session_token)` in the DB. On lookup:
```go
hashedToken := fmt.Sprintf("%x", sha256.Sum256([]byte(rawToken)))
// SELECT * FROM user_sessions WHERE session_token = $1 AND is_active = TRUE AND expires_at > NOW()
```

**8. Document and test casbin_rule isolation boundary**

Add a test that verifies cross-domain policy writes via the `authz.Service` API are
impossible (i.e., `AddPolicy` with a different tenant's domain is blocked at the
application layer). This makes the absence of DB-level isolation an explicit, tested
contract.

### Long-term (P3 — within 3 months)

**9. Implement ABAC policy evaluator or remove the tables**

The `policies`, `permissions`, `policy_evaluations` tables are fully built out but
have no Go evaluator. Either:
- Implement an evaluation engine that reads from `policies` and uses `user_attributes`
  and `user_sessions.device_info/location_info` for ABAC decisions
- Or explicitly mark these tables as "future roadmap" and ensure no admin UI implies
  they are active

**10. Wire role hierarchy into Casbin**

If `parent_role_id` inheritance is desired, add a second `g2 = _, _` relation to the
Casbin model and populate it from the DB `roles` table on startup. Without this, role
hierarchy is a data model feature only.

**11. Extend `enforce_tenant_isolation()` trigger to all tables**

The trigger body currently handles only `persons`. Add cases for `employees`, `users`,
and any table that cross-references `entity_id`.

---

## Appendix — File Reference Map

| Component | File |
|-----------|------|
| DB roles & extensions | `db/migration/000051_tenant_extensions_and_roles.up.sql` |
| Tenant RLS policies | `db/migration/000056_tenant_rls.up.sql` |
| Tenant context functions | `db/migration/000055_tenant_functions.up.sql` |
| Cross-tenant trigger | `db/migration/000107_enforce_tenant_isolation.up.sql` |
| Users table + RLS | `db/migration/000303_identity_create_users.up.sql` |
| Sessions table + RLS | `db/migration/000304_identity_add_user_sessions.up.sql` |
| Casbin rule table | `db/migration/000063_authz_casbin_rule.up.sql` |
| Role assignments table | `db/migration/000064_authz_role_assignments.up.sql` |
| Roles table (ABAC) | `db/migration/000405_auth_create_roles.up.sql` |
| ABAC policies table | `db/migration/000408_policy_define_rules.up.sql` |
| Permissions table | `db/migration/000404_auth_define_permissions.up.sql` |
| Policy evaluations cache | `db/migration/000703_policy_evaluation.up.sql` |
| Authz service interface | `internal/core/authz/authz.go` |
| Casbin model definition | `internal/core/authz/model.go` |
| Service constructor + Enforce | `internal/core/authz/service.go` |
| Role operations | `internal/core/authz/roles.go` |
| Policy CRUD | `internal/core/authz/policies.go` |
| pgx adapter | `internal/core/authz/adapter.go` |
| Fiber middleware | `internal/core/authz/middleware.go` |
| Type definitions & helpers | `internal/core/authz/types.go` |
| Sentinel errors | `internal/core/authz/errors.go` |
| Service integration tests | `internal/core/authz/service_test.go` |
| Role integration tests | `internal/core/authz/roles_test.go` |
