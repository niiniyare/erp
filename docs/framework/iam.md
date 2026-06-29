# AwoERP — Identity, Access Management & Multi-Tenant Authorization
## Users, Roles, Permissions, Authentication & Authorization — v2.0

**Platform:** AwoERP (`awo.so`) — Go · Fiber v2 · PostgreSQL RLS · Casbin · CEL  
**Scope:** Conceptual entities, data model design, and pseudo-code patterns. No implementation code.  
**Audience:** Platform architects, ERP implementors, module contributors

---

## Table of Contents

1. [The Two Principal Planes](#1-the-two-principal-planes)
2. [Plane 1 — Platform Users](#2-plane-1--platform-users)
   - 2.1 Platform User Tiers
   - 2.2 Platform User Entity
   - 2.3 What Platform Users Can and Cannot Do
   - 2.4 Root Account Bootstrap
3. [Plane 2 — Tenant Users](#3-plane-2--tenant-users)
   - 3.1 Tenant User Entity
   - 3.2 User Status Lifecycle
4. [Tenancy Model](#4-tenancy-model)
   - 4.1 Tenant Entity
   - 4.2 Tenant Status Lifecycle
   - 4.3 Tenant Resolution from HTTP Request
5. [Organisation Units & Data Scoping](#5-organisation-units--data-scoping)
   - 5.1 The OU Tree
   - 5.2 OU Entity
   - 5.3 Real-World Example — Vivo Energy Africa
   - 5.4 OU Visibility Rules
6. [Role & Permission Model](#6-role--permission-model)
   - 6.1 Permission String Taxonomy
   - 6.2 Role Entity
   - 6.3 Permission Entity
   - 6.4 Standard Actions
7. [Role Assignment & OU Scoping](#7-role-assignment--ou-scoping)
   - 7.1 UserRoleAssignment Entity
   - 7.2 Scope Resolution Precedence
   - 7.3 Effective Permissions Computation
   - 7.4 Permission Snapshot (Session Bake)
8. [Authentication](#8-authentication)
   - 8.1 Authentication Methods
   - 8.2 Password Policy
   - 8.3 Login Flow — Platform User
   - 8.4 Login Flow — Tenant User
   - 8.5 SSO Flow
   - 8.6 MFA Enrollment
9. [Session & Token Lifecycle](#9-session--token-lifecycle)
   - 9.1 Session Entity
   - 9.2 JWT Structure — Platform vs Tenant
   - 9.3 Token Refresh & Rotation
   - 9.4 Session Revocation
10. [Authorization Engine](#10-authorization-engine)
    - 10.1 Architecture
    - 10.2 Platform-Plane Authorization
    - 10.3 Tenant-Plane Authorization
    - 10.4 Authorize Middleware Pseudo-Code
    - 10.5 CEL Context & Example Conditions
11. [Row-Level Security](#11-row-level-security)
    - 11.1 Tenant Isolation Layer
    - 11.2 OU Isolation Layer
    - 11.3 Context Injection Pattern
12. [API Keys](#12-api-keys)
13. [Audit Trail](#13-audit-trail)
14. [Cross-Cutting Request Flow](#14-cross-cutting-request-flow)
15. [Entity Relationship Summary](#15-entity-relationship-summary)
16. [Appendix A — Implementation Checklist](#appendix-a--implementation-checklist)
17. [Appendix B — Frappe Concept Mapping](#appendix-b--frappe-concept-mapping)

---

## 1. The Two Principal Planes

AwoERP operates with **two completely separate identity planes**. They share no user table, no session table, and no permission model. They are isolated at every layer — HTTP routing, database access, audit logging, and token signing.

```
┌──────────────────────────────────────────────────────────────────────┐
│                        AwoERP Platform                               │
│                                                                      │
│  ┌─────────────────────────────┐                                     │
│  │     PLANE 1: PLATFORM       │  Anthropic / SaaS operator staff    │
│  │     USERS                   │                                     │
│  │                             │  • Manage tenants (create,          │
│  │  Root Account               │    suspend, configure)              │
│  │  Platform Admins            │  • Read business data for support   │
│  │  Platform Support           │  • Cannot write tenant business     │
│  │  Platform Readonly          │    data                             │
│  │                             │  • Access every tenant              │
│  └─────────────────────────────┘                                     │
│                                                                      │
│  ┌─────────────────────────────┐                                     │
│  │     PLANE 2: TENANT         │  Tenant staff — employees,          │
│  │     USERS                   │  managers, cashiers, etc.           │
│  │                             │                                     │
│  │  Tenant Admin               │  • Access ONLY their own tenant     │
│  │  Tenant Managers            │  • Scoped to their OU subtree       │
│  │  Tenant Staff               │  • Cannot cross tenant boundary     │
│  │  Tenant API Keys            │  • Cannot see platform config       │
│  └─────────────────────────────┘                                     │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

The two planes are distinguished at the **token level** (separate `aud` claim and signing key), at the **middleware level** (two separate auth middleware chains), and at the **database level** (separate schema or clear column separation for platform entities).

---

## 2. Plane 1 — Platform Users

### 2.1 Platform User Tiers

Platform users are internal to the SaaS operator (Anthropic / AwoERP company). There are four tiers, ordered from most to least privileged. Tier membership is determined at startup from configuration — not from a database table modifiable at runtime.

```
Tier 0: ROOT
  The root account. One instance. Credentials set via environment variables
  or a sealed secrets manager at deploy time. Can perform any operation on
  any resource across the entire platform including other platform user
  management. This account should never be used for routine operations.

Tier 1: PLATFORM_ADMIN
  Full platform configuration access. Can create/suspend/archive tenants,
  modify platform-level feature flags, manage platform user accounts,
  and read any tenant's business data. Cannot execute financial transactions
  or write operational records within a tenant.

Tier 2: PLATFORM_SUPPORT
  Scoped read access to tenant business data for customer support purposes.
  Can read any tenant's records, users, sessions, and audit logs. Cannot
  modify tenant configuration, user roles, or business data.

Tier 3: PLATFORM_READONLY
  Observability and monitoring access only. Can read platform metrics,
  health status, tenant status, and aggregate billing data. Cannot read
  individual tenant business records or user PII.
```

### 2.2 Platform User Entity

```
Entity: PlatformUser
─────────────────────────────────────────────────────────────────────
id                  UUID
email               VARCHAR(320)    Unique across platform
full_name           VARCHAR(255)
tier                ENUM            ROOT | PLATFORM_ADMIN | PLATFORM_SUPPORT | PLATFORM_READONLY
status              ENUM            ACTIVE | SUSPENDED | DEACTIVATED
password_hash       TEXT?           Argon2id; NULL if SSO-only
mfa_secret          TEXT?           TOTP (encrypted at rest)
mfa_enabled         BOOL            Default TRUE for ADMIN+; enforced
mfa_required        BOOL            Computed from tier; ROOT and ADMIN always TRUE
last_login_at       TIMESTAMPTZ?
last_login_ip       INET?
failed_login_count  INT             Default 0
locked_until        TIMESTAMPTZ?
created_by          UUID?           FK → PlatformUser (NULL for bootstrapped root)
created_at          TIMESTAMPTZ
updated_at          TIMESTAMPTZ
```

There is **no `tenant_id`** on a platform user. They exist outside the tenancy model entirely.

### 2.3 What Platform Users Can and Cannot Do

The fundamental rule is: **platform users operate on tenant configurations and SaaS management — they never write tenant business data.**

```
┌────────────────────────────────────┬───────┬───────┬─────────┬──────────┐
│ Capability                         │ ROOT  │ ADMIN │ SUPPORT │ READONLY │
├────────────────────────────────────┼───────┼───────┼─────────┼──────────┤
│ Create / archive tenants           │  ✓    │  ✓    │   ✗     │   ✗      │
│ Suspend / reactivate tenants       │  ✓    │  ✓    │   ✗     │   ✗      │
│ Modify tenant feature flags        │  ✓    │  ✓    │   ✗     │   ✗      │
│ Manage tenant billing / plan       │  ✓    │  ✓    │   ✗     │   ✗      │
│ Manage platform user accounts      │  ✓    │  ✓    │   ✗     │   ✗      │
│ Read any tenant's business data    │  ✓    │  ✓    │   ✓     │   ✗      │
│ Read tenant user list & roles      │  ✓    │  ✓    │   ✓     │   ✗      │
│ Read tenant audit logs             │  ✓    │  ✓    │   ✓     │   ✗      │
│ Write tenant business records      │  ✗    │  ✗    │   ✗     │   ✗      │
│ Modify tenant user roles/passwords │  ✗    │  ✗    │   ✗     │   ✗      │
│ Read platform health & metrics     │  ✓    │  ✓    │   ✓     │   ✓      │
│ Read aggregate billing summaries   │  ✓    │  ✓    │   ✗     │   ✓      │
│ Change platform configuration      │  ✓    │  ✗    │   ✗     │   ✗      │
│ Manage signing keys & secrets      │  ✓    │  ✗    │   ✗     │   ✗      │
│ Revoke any active session          │  ✓    │  ✓    │   ✗     │   ✗      │
└────────────────────────────────────┴───────┴───────┴─────────┴──────────┘

✗ on "Write tenant business records" is enforced at the DB layer as a
  read-only connection role, not just in application logic.
```

### 2.4 Root Account Bootstrap

The root account is the only account that is never stored in a modifiable database table. It is seeded from environment configuration at first startup and thereafter cannot be created or deleted through any API endpoint.

```
Environment / Sealed Secret Configuration:
─────────────────────────────────────────────────────────────────────
PLATFORM_ROOT_EMAIL            root@internal.awo.so
PLATFORM_ROOT_PASSWORD_HASH    <argon2id hash set at deploy time>
PLATFORM_ROOT_MFA_SECRET       <TOTP secret, sealed>
PLATFORM_ADMIN_IPS             10.0.0.0/8,172.16.0.0/12   (allowlist)
PLATFORM_JWT_SIGNING_KEY       <ES256 private key, sealed>
PLATFORM_JWT_SIGNING_KID       platform-2025-01
```

The root account:
- Has no `id` row in `platform_users` — it is the identity the process itself runs as.
- Cannot be suspended via the API.
- Its credentials are rotated through the secrets manager, not through any application endpoint.
- All actions taken by root are tagged `actor_type: system_root` in the audit log.
- Login from the root account requires both password AND TOTP and is only permitted from allowlisted IP ranges.

```
PlatformConfig entity (read-only after startup):
─────────────────────────────────────────────────────────────────────
platform_name           STRING      "AwoERP"
platform_domain         STRING      "awo.so"
root_email              STRING      (from env)
admin_ip_allowlist      CIDR[]      (from env)
mfa_required_tiers      ENUM[]      [ROOT, PLATFORM_ADMIN]
tenant_subdomain_tpl    STRING      "{slug}.awo.so"
default_tenant_plan     ENUM        STARTER
jwt_platform_kid        STRING      (from env)
jwt_tenant_kid          STRING      (from env)
```

---

## 3. Plane 2 — Tenant Users

### 3.1 Tenant User Entity

Tenant users belong to exactly one tenant. Their data is stored in RLS-protected tables — the database engine refuses to return rows unless `app.current_tenant_id` matches. There is no application-level bypass.

```
Entity: TenantUser
─────────────────────────────────────────────────────────────────────
id                   UUID
tenant_id            UUID            FK → Tenant (RLS column)
email                VARCHAR(320)    Unique within tenant (not globally)
username             VARCHAR(64)?    Optional; unique within tenant
full_name            VARCHAR(255)
status               ENUM            PENDING_VERIFICATION | ACTIVE | SUSPENDED | DEACTIVATED
actor_type           ENUM            user | service_account
password_hash        TEXT?           Argon2id; NULL if SSO-only
mfa_secret           TEXT?           TOTP secret (encrypted at rest)
mfa_enabled          BOOL            Default FALSE
email_verified       BOOL            Default FALSE
last_login_at        TIMESTAMPTZ?
last_login_ip        INET?
failed_login_count   INT             Default 0
locked_until         TIMESTAMPTZ?
must_change_password BOOL            Default FALSE
locale               VARCHAR(10)     e.g. "en-KE"
timezone             VARCHAR(50)     e.g. "Africa/Nairobi"
metadata             JSONB
created_at           TIMESTAMPTZ
updated_at           TIMESTAMPTZ
```

A tenant user with the email `manager@vivoenergy.co.ke` is entirely separate from a platform user with that same email. They are different principals, authenticated against different tables, issued tokens with different signing keys and different `aud` claims.

### 3.2 User Status Lifecycle

```
PENDING_VERIFICATION ──► ACTIVE ──► SUSPENDED ──► ACTIVE
                              │
                              └──────────────────► DEACTIVATED
```

- `PENDING_VERIFICATION` — email link sent, not yet clicked.
- `SUSPENDED` — admin action; data intact, login rejected.
- `DEACTIVATED` — soft-delete; data retained for audit; not recoverable via API.

---

## 4. Tenancy Model

### 4.1 Tenant Entity

A tenant is the top-level isolation container for all business data. Every row of every business table belongs to exactly one tenant.

```
Entity: Tenant
─────────────────────────────────────────────────────────────────────
id               UUID
slug             VARCHAR(63)     Subdomain label e.g. "vivo-energy-africa"
display_name     VARCHAR(255)    "Vivo Energy Africa"
status           ENUM            PENDING | ACTIVE | SUSPENDED | ARCHIVED
plan             ENUM            FREE | STARTER | GROWTH | ENTERPRISE
country_code     VARCHAR(2)      ISO 3166-1 alpha-2 e.g. "KE"
currency_code    VARCHAR(3)      ISO 4217 e.g. "KES"
timezone         VARCHAR(50)     e.g. "Africa/Nairobi"
locale           VARCHAR(10)     e.g. "en-KE"
feature_flags    JSONB           Per-tenant feature toggles set by platform admin
settings         JSONB           Tenant-managed configuration
created_by       UUID?           FK → PlatformUser (who provisioned this tenant)
created_at       TIMESTAMPTZ
updated_at       TIMESTAMPTZ
```

### 4.2 Tenant Status Lifecycle

```
PENDING ──► ACTIVE ──► SUSPENDED ──► ACTIVE
   │             │
   └─────────────┴──────────────────► ARCHIVED
```

Only a platform admin (Tier 1+) can change tenant status. A tenant user — including the tenant admin — cannot suspend or archive their own tenant.

### 4.3 Tenant Resolution from HTTP Request

Every incoming request is resolved to a tenant before any authentication or handler logic executes.

```
Incoming Request  →  Host: vivo-energy-africa.awo.so
                                │
                                ▼
                     TenantResolutionMiddleware
                      ├── Extract slug from Host header
                      ├── Cache lookup (slug → Tenant, TTL 5 min)
                      ├── On cache miss: DB lookup, populate cache
                      ├── Assert tenant.status == ACTIVE
                      │     else → 403 "Tenant unavailable"
                      └── Inject tenant into request context
                               │
                               ▼ (used by all downstream middleware)
                     ctx.tenant_id = tenant.id
                     ctx.tenant_slug = tenant.slug
```

For platform API routes (subdomain: `platform.awo.so` or `api.awo.so`), the tenant resolution middleware is **skipped** — these routes operate outside the tenant context.

---

## 5. Organisation Units & Data Scoping

An Organisation Unit (OU) is a node in a tenant's internal hierarchy. OUs define **who can see what data** within a tenant. A user's data visibility is bounded by the OU(s) they are assigned to and all child OUs beneath.

### 5.1 The OU Tree

The OU tree is a **rooted, ordered tree** of unlimited depth stored using PostgreSQL `ltree` for efficient ancestor/descendant queries. Every tenant has exactly one root OU (the company itself). All other OUs hang beneath it.

```
Root (company)
  └── Level 1 (region / country / subsidiary)
        └── Level 2 (branch / division / department)
              └── Level 3 (team / cost centre / workstation)
                    └── Level N (unlimited)
```

### 5.2 OU Entity

```
Entity: OrgUnit
─────────────────────────────────────────────────────────────────────
id              UUID
tenant_id       UUID            FK → Tenant (RLS-protected)
parent_id       UUID?           FK → OrgUnit; NULL = root node
code            VARCHAR(32)     Unique within tenant e.g. "VEA-KE-NBI-001"
name            VARCHAR(255)    "Nairobi Branch"
type            ENUM            COMPANY | SUBSIDIARY | REGION | COUNTRY |
                                BRANCH | DIVISION | DEPARTMENT | COST_CENTRE |
                                WORKSTATION
depth           INT             0 = root; auto-computed on write
path            LTREE           e.g. "VEA.VEA_KE.VEA_KE_NBI.VEA_KE_NBI_001"
                                Auto-maintained; used for subtree queries
is_active       BOOL
metadata        JSONB           Custom attributes (EPRA licence, KRA PIN, etc.)
created_at      TIMESTAMPTZ
updated_at      TIMESTAMPTZ
```

### 5.3 Real-World Example — Vivo Energy Africa

This example shows a multinational fuel retailer operating across East Africa as a single AwoERP tenant, with country-level subsidiaries and branch-level stations.

```
Vivo Energy Africa HQ          [type: COMPANY,    depth: 0, path: VEA]
├── Vivo Energy Kenya           [type: SUBSIDIARY, depth: 1, path: VEA.KE]
│   ├── Nairobi Region          [type: REGION,     depth: 2, path: VEA.KE.NBI]
│   │   ├── Shell Maanzoni      [type: BRANCH,     depth: 3, path: VEA.KE.NBI.MZN]
│   │   │   ├── Forecourt       [type: DEPARTMENT, depth: 4, path: VEA.KE.NBI.MZN.FC]
│   │   │   └── Shop            [type: DEPARTMENT, depth: 4, path: VEA.KE.NBI.MZN.SHP]
│   │   └── Shell Westlands     [type: BRANCH,     depth: 3, path: VEA.KE.NBI.WST]
│   └── Mombasa Region          [type: REGION,     depth: 2, path: VEA.KE.MBA]
│       └── Shell Nyali         [type: BRANCH,     depth: 3, path: VEA.KE.MBA.NYL]
├── Vivo Energy Uganda          [type: SUBSIDIARY, depth: 1, path: VEA.UG]
│   └── Kampala Region          [type: REGION,     depth: 2, path: VEA.UG.KLA]
│       └── Shell Kololo        [type: BRANCH,     depth: 3, path: VEA.UG.KLA.KOL]
└── Vivo Energy Tanzania        [type: SUBSIDIARY, depth: 1, path: VEA.TZ]
    └── Dar es Salaam Region    [type: REGION,     depth: 2, path: VEA.TZ.DSM]
```

### 5.4 OU Visibility Rules

Visibility is determined by a single, consistent rule applied at the database query layer via RLS:

> **A user can see a data record if and only if the record's `org_unit_path` is a descendant of (or equal to) at least one of the user's assigned OU paths.**

In `ltree` terms: `record.org_unit_path <@ user.assigned_ou_path`

```
User                      Assigned OU Path     Visible Data
─────────────────────────────────────────────────────────────────────
Group CFO                 VEA                  All records (Kenya + Uganda + Tanzania)
Vivo Kenya CEO            VEA.KE               All Kenya records only
Nairobi Regional Manager  VEA.KE.NBI           Nairobi region records only
Shell Maanzoni Manager    VEA.KE.NBI.MZN       Maanzoni station records only
Forecourt Supervisor      VEA.KE.NBI.MZN.FC    Forecourt department records only
```

The Vivo Kenya CEO assigned at `VEA.KE` **cannot** see records tagged `VEA.UG.*` (Uganda) or `VEA.TZ.*` (Tanzania) even if they have the correct permission string — the RLS policy rejects those rows at the database engine before the application sees them.

A user may be assigned to **multiple OUs** with different roles. Visibility is the **union** of all their assigned subtrees. For example, a regional auditor could be assigned read-only access to both `VEA.KE.NBI` and `VEA.KE.MBA`, seeing both Nairobi and Mombasa data but not Uganda or Tanzania.

**Sibling isolation is absolute** — there is no mechanism to grant a user in branch `VEA.KE.NBI.MZN` access to sibling branch `VEA.KE.NBI.WST` without also granting `VEA.KE.NBI` (the parent), which would expose all Nairobi branches. This is intentional: access is always granted at an ancestor node, never selectively at sibling nodes.

---

## 6. Role & Permission Model

### 6.1 Permission String Taxonomy

All permissions follow the format: `<module>.<resource>.<action>`

The module and resource segments identify what is being accessed. The action identifies what operation is being performed.

```
Standard actions across all resources:
─────────────────────────────────────────────────────────────────────
read        View list and individual record detail
write       Create new records and update existing ones
delete      Soft-delete or hard-delete (distinguished by resource policy)
submit      Advance a document through its workflow (e.g. close a shift)
cancel      Reverse a submitted document
amend       Create an amendment of a cancelled document
export      Download or extract data outside the system
import      Bulk-load data into the system
print       Generate printed output / PDF
report      Access analytics and reporting views
share       Share a record with other users
admin       Shorthand for all actions on this resource

Platform-only action prefix (Plane 1 routes only):
─────────────────────────────────────────────────────────────────────
platform.*  Prefix for all platform-administration permissions
            e.g. platform.tenants.write, platform.users.read
```

**Example permission strings by module:**

```
# IAM (Tenant Plane)
iam.users.read              iam.users.write             iam.users.delete
iam.roles.read              iam.roles.write             iam.roles.admin
iam.sessions.read           iam.org-units.read          iam.org-units.write

# Finance
finance.accounts.read       finance.accounts.write
finance.transactions.read   finance.transactions.write
finance.transactions.submit finance.transactions.cancel
finance.reports.read        finance.reports.export

# Forecourt (FMS)
fms.shifts.read             fms.shifts.write            fms.shifts.submit
fms.pumps.read              fms.pumps.write
fms.wetstock.read           fms.wetstock.write
fms.prices.read             fms.prices.write
fms.cashier-reconciliation.read   fms.cashier-reconciliation.submit
fms.dips.read               fms.dips.write

# Inventory
inventory.items.read        inventory.items.write
inventory.stock.read        inventory.stock.write       inventory.stock.submit
inventory.suppliers.read    inventory.suppliers.write

# Tenant Administration
tenant.settings.read        tenant.settings.write
tenant.org-units.read       tenant.org-units.write
tenant.billing.read

# Platform Administration (Plane 1 only — unavailable to tenant users)
platform.tenants.read       platform.tenants.write
platform.tenants.suspend    platform.tenants.archive
platform.users.read         platform.users.write
platform.config.read        platform.config.write
platform.audit.read         platform.metrics.read
```

### 6.2 Role Entity

```
Entity: Role
─────────────────────────────────────────────────────────────────────
id              UUID
tenant_id       UUID?           NULL = system-defined built-in role
                                NOT NULL = tenant-specific custom role
plane           ENUM            PLATFORM | TENANT
name            VARCHAR(100)    "Country Manager"
slug            VARCHAR(100)    "country-manager"
description     TEXT?
is_system       BOOL            TRUE = immutable; cannot be modified
is_active       BOOL
parent_role_id  UUID?           FK → Role (optional inheritance)
created_at      TIMESTAMPTZ
updated_at      TIMESTAMPTZ
```

**Built-in system roles — Tenant Plane:**

| Slug | Description |
|---|---|
| `tenant-admin` | All tenant permissions; manages users, roles, settings |
| `tenant-manager` | Write access to business modules; no IAM or billing |
| `tenant-staff` | Read-only by default; additional permissions added per role |
| `tenant-api-readonly` | Read-only API key role |

**Built-in system roles — Platform Plane:**

| Slug | Description |
|---|---|
| `platform-root` | All platform permissions; tied to root account only |
| `platform-admin` | All platform management; cannot write tenant business data |
| `platform-support` | Read any tenant business data; no management |
| `platform-readonly` | Platform metrics and health only |

Tenants can create unlimited custom roles extending any tenant-plane system role via `parent_role_id`. Platform roles cannot be extended by tenants.

### 6.3 Permission Entity

```
Entity: Permission
─────────────────────────────────────────────────────────────────────
id              UUID
tenant_id       UUID?           NULL = system-defined
role_id         UUID            FK → Role
permission      VARCHAR(200)    Dotted permission string
scope_type      ENUM?           NULL | org_unit | record
scope_id        UUID?           Specific OU or record ID (when scope-typed)
condition       TEXT?           CEL expression for attribute-based conditions
granted         BOOL            TRUE = grant; FALSE = explicit deny
created_at      TIMESTAMPTZ
```

The `condition` field allows attribute-based rules beyond the permission string alone:

```cel
// Only allow reading records the user personally created
resource.created_by == subject.user_id

// Only allow write operations during business hours (tenant timezone)
now().getHours() >= 8 && now().getHours() < 18

// Only allow high-value transaction approval if specific flag is present
resource.amount_kes <= 500000 ||
  "finance.transactions.approve-high-value" in subject.permissions

// OU subtree: record must fall within user's assigned OU
resource.org_unit_path.startsWith(subject.org_unit_path)
```

---

## 7. Role Assignment & OU Scoping

### 7.1 UserRoleAssignment Entity

This table binds a user to a role within an optional OU scope. It is the single source of truth for "who has what access where."

```
Entity: UserRoleAssignment
─────────────────────────────────────────────────────────────────────
id              UUID
tenant_id       UUID            FK → Tenant (RLS-protected)
user_id         UUID            FK → TenantUser
role_id         UUID            FK → Role
org_unit_id     UUID?           FK → OrgUnit
                                NULL = tenant-wide (no OU restriction on permissions;
                                       RLS still limits data visibility)
org_unit_path   LTREE?          Denormalized from OrgUnit.path (for fast RLS queries)
granted_by      UUID            FK → TenantUser (the admin who made this assignment)
granted_at      TIMESTAMPTZ
expires_at      TIMESTAMPTZ?    Time-limited role grants
is_active       BOOL
reason          TEXT?           Justification for audit trail
```

### 7.2 Scope Resolution Precedence

When a user has multiple role assignments (possibly across multiple OUs), permission conflicts are resolved in this order:

```
Priority 1 (highest): Explicit DENY at any scope         → always wins; no override
Priority 2:           Grant at record scope              → most specific
Priority 3:           Grant at exact OU scope            → exact OU match
Priority 4:           Grant at ancestor OU scope         → inherited from parent
Priority 5:           Grant at tenant scope (no OU)      → broadest permission grant
Priority 6 (default): No matching rule                   → DENY
```

The system is **default-deny**. A user has no access to anything unless a rule explicitly grants it.

### 7.3 Effective Permissions Computation

```
function computeEffectivePermissions(userID, tenantID):

    assignments = loadActiveRoleAssignments(userID, tenantID)

    // Filter expired assignments
    assignments = filter(assignments, a => a.expires_at == nil or now() < a.expires_at)

    rawPermissions = []

    for assignment in assignments:
        role = loadRole(assignment.role_id)
        rolePerms = collectRolePermissions(role)   // walks parent_role chain

        for perm in rolePerms:
            rawPermissions.append({
                permission:   perm.permission,
                granted:      perm.granted,
                scope_type:   perm.scope_type ?? (assignment.org_unit_id != nil ? "org_unit" : nil),
                scope_id:     perm.scope_id   ?? assignment.org_unit_id,
                condition:    perm.condition,
            })

    return resolveConflicts(rawPermissions)


function collectRolePermissions(role):
    perms = loadPermissionsForRole(role.id)

    if role.parent_role_id != nil:
        parentPerms = collectRolePermissions(loadRole(role.parent_role_id))
        // Child permissions override parent on conflict (more specific wins)
        perms = mergeWithChildPriority(parentPerms, perms)

    return perms


function resolveConflicts(permissions):
    byPermissionString = groupBy(permissions, p => p.permission)
    resolved = []

    for permString, group in byPermissionString:
        denies = filter(group, p => p.granted == false)
        if len(denies) > 0:
            // Any explicit deny takes precedence — append deny
            resolved.append({ permission: permString, granted: false })
            continue

        // Among grants, pick the most specific scope
        grants = sortBySpecificity(group)    // record > org_unit > tenant
        resolved.append(grants[0])

    return resolved
```

### 7.4 Permission Snapshot (Session Bake)

At login, the full effective permission set is computed once and embedded in the session. No permission database lookup occurs during ordinary request processing.

```
Session.Permissions = [
    "finance.accounts.read",
    "finance.transactions.read",
    "finance.transactions.write",
    "fms.shifts.read",
    "fms.shifts.submit",
    ...
]

Session.OrgUnitPaths = [
    "VEA.KE.NBI.MZN",    // Maanzoni Branch — full write access
    "VEA.KE.MBA",         // Mombasa Region — read-only (different role)
]
```

The snapshot is **invalidated** (and the user's next token refresh will recompute it) when:
- A role assignment is added, removed, or expires for this user
- Any permission is changed on any of this user's roles
- The tenant admin explicitly forces a session refresh
- The session's TTL expires

---

## 8. Authentication

### 8.1 Authentication Methods

Both planes support the same authentication primitives, but credentials are stored and validated in completely separate tables with separate secrets.

| Method | Platform Users | Tenant Users |
|---|---|---|
| Email + Password | ✓ | ✓ |
| TOTP MFA | ✓ (enforced for ROOT + ADMIN) | ✓ (optional, tenant-configurable) |
| SSO / OIDC | ✓ | ✓ |
| Magic Link | ✗ | ✓ |
| API Key | ✗ | ✓ |

### 8.2 Password Policy

```
Entity: PasswordPolicy
─────────────────────────────────────────────────────────────────────
tenant_id            UUID?           NULL = platform default policy
min_length           INT             Default 10
require_uppercase    BOOL            Default TRUE
require_lowercase    BOOL            Default TRUE
require_digit        BOOL            Default TRUE
require_symbol       BOOL            Default FALSE
max_age_days         INT?            NULL = no forced expiry
prevent_reuse_count  INT             Default 5 (cannot reuse last N passwords)
lockout_attempts     INT             Default 5 failed attempts
lockout_duration_m   INT             Default 15 minutes
session_timeout_m    INT             Default 480 (8 hours)
```

Platform users inherit the platform-level password policy which is stricter than the default tenant policy:
- Minimum length: 16 characters
- Symbol required: TRUE
- MFA enforced for ROOT and ADMIN tiers regardless of policy setting

### 8.3 Login Flow — Platform User

```
function platformLogin(email, password, mfaCode?, requestIP):

    // Platform users must come from allowlisted IPs
    if PLATFORM_ADMIN_IPS is configured:
        assert requestIP in PLATFORM_ADMIN_IPS
            else → 403 "Access from this IP is not permitted"

    user = findPlatformUserByEmail(email)
    assert user != nil and user.status == ACTIVE
        else → 401 "Invalid credentials"   // never reveal which failed

    if user.locked_until != nil and now() < user.locked_until:
        return 401 "Account temporarily locked"

    if not argon2Verify(password, user.password_hash):
        incrementFailedCount(user)
        if user.failed_login_count >= PLATFORM_LOCKOUT_THRESHOLD:
            lockUser(user, duration: PLATFORM_LOCKOUT_DURATION)
        return 401 "Invalid credentials"

    // MFA is mandatory for ROOT and PLATFORM_ADMIN
    if user.mfa_required or user.mfa_enabled:
        if mfaCode == nil:
            return 200, { mfa_required: true }
        assert verifyTOTP(mfaCode, decrypt(user.mfa_secret))
            else → 401 "Invalid MFA code"

    resetFailedCount(user)
    updateLastLogin(user, ip: requestIP)

    // Platform session — no tenant_id, no OU scope
    session = createPlatformSession(user)

    return 200, {
        access_token:  signJWT(session, key: PLATFORM_SIGNING_KEY, ttl: 15m),
        refresh_token: signRefreshToken(session.id, key: PLATFORM_REFRESH_KEY, ttl: 4h),
        user:          platformUserView(user),
        tier:          user.tier,
    }

    // Note: platform refresh tokens have a shorter TTL than tenant tokens
    // because platform sessions should not persist overnight unattended
```

### 8.4 Login Flow — Tenant User

```
function tenantLogin(tenantSlug, email, password, mfaCode?):

    tenant = resolveTenant(tenantSlug)              // from middleware context
    assert tenant.status == ACTIVE

    user = findTenantUserByEmail(tenant.id, email)
    assert user != nil and user.status == ACTIVE
        else → 401 "Invalid credentials"

    if user.locked_until != nil and now() < user.locked_until:
        return 401 "Account temporarily locked"

    policy = loadPasswordPolicy(tenant.id)

    if not argon2Verify(password, user.password_hash):
        incrementFailedCount(user)
        if user.failed_login_count >= policy.lockout_attempts:
            lockUser(user, duration: policy.lockout_duration_m)
        return 401 "Invalid credentials"

    if user.mfa_enabled:
        if mfaCode == nil:
            return 200, { mfa_required: true }
        assert verifyTOTP(mfaCode, decrypt(user.mfa_secret))
            else → 401 "Invalid MFA code"

    if user.must_change_password:
        // Issue a limited token valid only for the password-change endpoint
        changeToken = signJWT({ user_id: user.id, scope: "password_change_only" }, ttl: 10m)
        return 200, { must_change_password: true, change_token: changeToken }

    resetFailedCount(user)
    updateLastLogin(user, ip: request.ip)

    permissions = computeEffectivePermissions(user.id, tenant.id)
    orgUnitPaths = loadUserOrgUnitPaths(user.id, tenant.id)
    session = createTenantSession(user, tenant, permissions, orgUnitPaths)

    return 200, {
        access_token:  signJWT(session, key: TENANT_SIGNING_KEY, ttl: 15m),
        refresh_token: signRefreshToken(session.id, key: TENANT_REFRESH_KEY, ttl: 7d),
        user:          tenantUserView(user),
        permissions:   session.permissions,
        org_unit_paths: session.org_unit_paths,
    }
```

### 8.5 SSO / OIDC Flow

```
Entity: SSOConfig
─────────────────────────────────────────────────────────────────────
id                UUID
tenant_id         UUID?           NULL = platform-level SSO config
provider          ENUM            google | microsoft | okta | custom_oidc
client_id         TEXT
client_secret     TEXT            Encrypted at rest
issuer_url        TEXT            OIDC discovery endpoint
allowed_domains   TEXT[]?         e.g. ["vivoenergy.com"] — restrict by email domain
auto_provision    BOOL            Create tenant user on first SSO login
default_role_id   UUID?           Role assigned on auto-provisioning
attribute_mapping JSONB           Map OIDC claims to TenantUser fields
is_active         BOOL


function ssoCallback(tenantSlug, provider, authCode):

    tenant = resolveTenant(tenantSlug)
    config = loadSSOConfig(tenant.id, provider)

    tokens = exchangeCodeForTokens(authCode, config)
    idToken = verifyAndDecodeIDToken(tokens.id_token, config.issuer_url)

    if config.allowed_domains != nil:
        assert emailDomain(idToken.email) in config.allowed_domains
            else → 403 "Email domain not permitted for this tenant"

    user = findTenantUserByEmail(tenant.id, idToken.email)

    if user == nil:
        if not config.auto_provision:
            return 403 "User not found. Contact your administrator."
        user = autoProvisionUser(tenant, idToken, config)

    assert user.status == ACTIVE

    // Continue identically to password-based login from here
    permissions = computeEffectivePermissions(user.id, tenant.id)
    session = createTenantSession(user, tenant, permissions, ...)

    return redirect(config.post_login_url + "?token=" + signJWT(session, ttl: 15m))
```

### 8.6 MFA Enrollment

```
function initiateMFAEnrollment(session):
    secret = generateTOTPSecret()                   // 20-byte random, base32-encoded
    encryptedSecret = encryptWithTenantKey(secret)
    storeUnconfirmedSecret(session.user_id, encryptedSecret)

    qrURI = buildOTPAuthURI(secret, email: session.user.email, issuer: "AwoERP")
    backupCodes = generateBackupCodes(count: 10)    // 10 × 8-char alphanumeric
    storeHashedBackupCodes(session.user_id, backupCodes)

    return { qr_uri: qrURI, backup_codes: backupCodes }
    // Raw backup codes shown ONCE; hashed copies stored


function confirmMFAEnrollment(session, totpCode):
    secret = loadUnconfirmedSecret(session.user_id)
    assert verifyTOTP(totpCode, decrypt(secret))
        else → 422 "Invalid TOTP code — ensure device clock is accurate"

    user.mfa_secret = secret                        // promote from unconfirmed
    user.mfa_enabled = TRUE
    save(user)
    auditRecord("user.mfa_enabled", session)
```

---

## 9. Session & Token Lifecycle

### 9.1 Session Entity

Platform and tenant sessions are stored in separate tables but share the same schema shape.

```
Entity: PlatformSession  /  TenantSession
─────────────────────────────────────────────────────────────────────
id              UUID            Primary key (also JWT `jti` claim)
user_id         UUID            FK → PlatformUser / TenantUser
tenant_id       UUID?           NULL for platform sessions
actor_type      ENUM            user | service_account | api_key
tier            ENUM?           Platform tier (for platform sessions only)
permissions     TEXT[]          Snapshot at session creation
org_unit_paths  LTREE[]?        OU subtree roots for this session (tenant only)
ip_address      INET
user_agent      TEXT
mfa_verified    BOOL
refresh_token   TEXT            Argon2id hash of raw token (never stored plain)
issued_at       TIMESTAMPTZ
expires_at      TIMESTAMPTZ
last_active_at  TIMESTAMPTZ
revoked_at      TIMESTAMPTZ?    NULL = active
revoke_reason   TEXT?
```

### 9.2 JWT Structure — Platform vs Tenant

Both planes use ES256 (ECDSA P-256) but with separate signing keys and distinguishable `aud` claims so a tenant access token can never be used against a platform API endpoint and vice versa.

```json
// Platform JWT payload
{
  "iss": "https://auth.awo.so",
  "aud": "platform.awo.so",
  "sub": "<platform_user_uuid>",
  "jti": "<platform_session_uuid>",
  "iat": 1735689600,
  "exp": 1735690500,
  "plane": "platform",
  "tier": "PLATFORM_SUPPORT",
  "permissions": ["platform.tenants.read", "platform.audit.read"]
}

// Tenant JWT payload
{
  "iss": "https://auth.awo.so",
  "aud": "vivo-energy-africa.awo.so",
  "sub": "<tenant_user_uuid>",
  "jti": "<tenant_session_uuid>",
  "iat": 1735689600,
  "exp": 1735690500,
  "plane": "tenant",
  "tenant_id": "<tenant_uuid>",
  "org_unit_paths": ["VEA.KE.NBI.MZN"],
  "permissions": ["fms.shifts.read", "fms.shifts.submit", "fms.cashier-reconciliation.submit"]
}
```

The `plane` claim is validated by the authentication middleware on every request. A request to a tenant route carrying a `plane: platform` token is rejected with 401 even if the signature is valid.

### 9.3 Token Refresh & Rotation

```
function refreshTokens(rawRefreshToken, expectedPlane):

    // Identify session by hashing the token
    session = findSessionByHash(argon2id(rawRefreshToken), plane: expectedPlane)
    assert session != nil and session.revoked_at == nil
        else → 401 "Invalid or revoked refresh token"
    assert now() < session.expires_at
        else → 401 "Refresh token expired — please log in again"

    // Rotate: invalidate old refresh token, issue new one
    session.revoked_at = now()
    session.revoke_reason = "rotation"
    save(session)

    newRefreshToken = secureRandom(32)
    newSession = cloneSession(session, newRefreshToken: newRefreshToken)

    // Re-compute permissions (catches role changes since last login)
    if expectedPlane == "tenant":
        newSession.permissions = computeEffectivePermissions(session.user_id, session.tenant_id)
        newSession.org_unit_paths = loadUserOrgUnitPaths(session.user_id, session.tenant_id)

    save(newSession)

    return {
        access_token:  signJWT(newSession, ttl: 15m),
        refresh_token: newRefreshToken,
    }
```

### 9.4 Session Revocation

```
// Logout — single session
function revokeSession(sessionID, reason):
    session = loadSession(sessionID)
    session.revoked_at = now()
    session.revoke_reason = reason
    save(session)
    cache.addToRevocationSet("revoked:" + session.jti, ttl: session.ttl_remaining)


// Force re-login for all of a user's sessions (e.g. password change, compromise)
function revokeAllUserSessions(userID, plane, reason):
    sessions = loadActiveSessionsByUser(userID, plane)
    for session in sessions:
        revokeSession(session.id, reason)
    cache.delete("permissions:" + session.tenant_id + ":" + userID)


// Platform admin: emergency lock of all sessions for a tenant
function revokeAllTenantSessions(tenantID, reason):
    sessions = loadActiveSessionsByTenant(tenantID)
    for session in sessions:
        revokeSession(session.id, reason)
```

The revocation check uses a **Redis revocation set** keyed by `jti`. On every request the middleware checks this set before trusting the JWT signature. This avoids a database lookup on the happy path while still allowing instantaneous revocation.

---

## 10. Authorization Engine

### 10.1 Architecture

AwoERP uses **Casbin v2** with CEL expression evaluation for attribute conditions. Two separate Casbin enforcer instances are maintained — one for the platform plane and one for the tenant plane — with different policy sets loaded from the database at startup and refreshed on change events.

### 10.2 Platform-Plane Authorization

Platform routes are secured by checking the `tier` field of the platform session against a static tier-permission matrix defined in configuration (not in the database). This matrix is intentionally non-modifiable at runtime.

```
TierPermissionMatrix (loaded from config at startup):
─────────────────────────────────────────────────────────────────────
ROOT              → ["*.*"]                            (all)
PLATFORM_ADMIN    → ["platform.*", "tenant_read.*"]   (manage + read)
PLATFORM_SUPPORT  → ["tenant_read.*"]                 (read only)
PLATFORM_READONLY → ["platform.metrics.read",
                      "platform.health.read"]
```

```
function authorizePlatformRequest(session, requiredPermission):

    if session.plane != "platform":
        return DENY, "Wrong authentication plane"

    tier = session.tier
    allowedPatterns = TIER_PERMISSION_MATRIX[tier]

    for pattern in allowedPatterns:
        if wildcardMatch(pattern, requiredPermission):
            return ALLOW

    return DENY, "Insufficient platform tier"
```

### 10.3 Tenant-Plane Authorization

Tenant authorization has two layers working together:

**Layer 1 — Permission check (application layer):** Does the user's permission snapshot contain the required permission string? This is a fast in-memory check against the session payload.

**Layer 2 — Data visibility (database layer):** Even with the correct permission, the RLS policy on the table restricts which rows the query returns based on the session's `org_unit_paths`. This layer is enforced by PostgreSQL itself.

The combination means:
- A user who lacks `fms.shifts.read` gets 403 before the query runs.
- A user who has `fms.shifts.read` but is scoped to `VEA.KE.NBI.MZN` runs the query but only receives Maanzoni shifts — Uganda shifts are invisible to the query even if the query contained no WHERE clause.

### 10.4 Authorize Middleware Pseudo-Code

```
middleware AuthenticateTenant:
    token = extractBearerToken(request)
    assert token != nil
        else → 401

    claims = verifyJWT(token, key: TENANT_SIGNING_KEY)
    assert claims.plane == "tenant"
        else → 401 "Wrong authentication plane"
    assert claims.aud == request.host
        else → 401 "Token audience mismatch"

    // Check revocation set (Redis O(1) lookup)
    assert not cache.isMember("revoked:" + claims.jti)
        else → 401 "Session revoked"

    session = buildSessionFromClaims(claims)
    ctx.session = session
    ctx.tenant_id = session.tenant_id
    ctx.org_unit_paths = session.org_unit_paths

    // Inject RLS context into DB connection
    db.exec("SET LOCAL app.current_tenant_id = $1", session.tenant_id)
    db.exec("SET LOCAL app.current_ou_paths = $1", session.org_unit_paths)

    return next(ctx)


middleware Authorize(requiredPermission string):
    session = ctx.session
    assert session != nil
        else → 401

    // Fast path: exact match in pre-baked snapshot
    if requiredPermission in session.permissions:
        return next(ctx)

    // Wildcard match (e.g. "fms.*.*" covers "fms.shifts.read")
    if anyWildcardMatch(session.permissions, requiredPermission):
        return next(ctx)

    // Slow path: evaluate CEL-conditioned permissions via Casbin
    // (record-scoped or time-windowed conditions not in snapshot)
    cel_ctx = buildCELContext(ctx, session)
    allowed = platformCasbin.Enforce(session.user_id, requiredPermission, cel_ctx)

    if not allowed:
        auditRecord("authorization.denied", { permission: requiredPermission }, session)
        return 403 "Forbidden"

    return next(ctx)
```

### 10.5 CEL Context & Example Conditions

```
CELContext passed to Casbin evaluation:
─────────────────────────────────────────────────────────────────────
subject:
    user_id         UUID
    tenant_id       UUID
    org_unit_paths  []LTREE
    permissions     []STRING
    actor_type      ENUM

resource:
    type            STRING      e.g. "fms.shift"
    id              UUID?
    owner_id        UUID?       (user who created this record)
    org_unit_path   LTREE?
    status          STRING?
    amount          FLOAT?
    attributes      MAP

request:
    ip              STRING
    timestamp       Timestamp
    user_agent      STRING
```

**Example conditions:**

```cel
// User can only submit shifts they personally opened
resource.owner_id == subject.user_id

// Record must fall within the user's OU subtree
resource.org_unit_path.startsWith(subject.org_unit_paths[0])

// High-value approvals require a specific elevated permission
resource.amount_kes <= 100000 ||
  "finance.transactions.approve-high-value" in subject.permissions

// Write access only during business hours in Nairobi time
request.timestamp.inTimeZone("Africa/Nairobi").getHours() >= 8 &&
request.timestamp.inTimeZone("Africa/Nairobi").getHours() < 18

// Country managers can see all records under their country subtree
subject.org_unit_paths.exists(p, resource.org_unit_path.startsWith(p))
```

---

## 11. Row-Level Security

### 11.1 Tenant Isolation Layer

Every table holding tenant business data is protected by a tenant isolation policy. This is the outermost isolation boundary — no application logic bypass is possible.

```sql
-- Applied to every tenant-owned business table
ALTER TABLE <table> ENABLE ROW LEVEL SECURITY;
ALTER TABLE <table> FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON <table>
    USING (
        tenant_id = current_setting('app.current_tenant_id')::uuid
    );
```

This policy uses `FORCE ROW LEVEL SECURITY`, which applies the policy even to the database role that owns the table — closing the superuser bypass. The application DB role is not a superuser.

### 11.2 OU Isolation Layer

For tables where data must be visible only within the user's OU subtree, a second policy layer is applied on top of the tenant policy:

```sql
-- Applied to tables with org_unit_path column
CREATE POLICY ou_scoped_access ON <table>
    USING (
        -- Tenant boundary (always required)
        tenant_id = current_setting('app.current_tenant_id')::uuid
        AND (
            -- No OU scope set means tenant-wide access
            current_setting('app.current_ou_paths', true) IS NULL
            OR current_setting('app.current_ou_paths', true) = ''
            OR
            -- Record's OU path is within any of the user's assigned OU subtrees
            org_unit_path <@ ANY(
                string_to_array(
                    current_setting('app.current_ou_paths'), ','
                )::lquery[]
            )
        )
    );
```

The `<@` operator (ltree "is descendant of or equal to") efficiently evaluates the subtree containment in a single index scan. No application-level filtering is needed.

### 11.3 Context Injection Pattern

```
function withTenantContext(ctx, tenantID, session, queryFn):

    conn = acquireConnection()

    try:
        // Always set tenant boundary
        conn.exec("SET LOCAL app.current_tenant_id = $1", tenantID)

        // Set OU scope if the session has restricted paths
        if session.org_unit_paths != nil and len(session.org_unit_paths) > 0:
            pathsCSV = join(session.org_unit_paths, ",")
            conn.exec("SET LOCAL app.current_ou_paths = $1", pathsCSV)
        else:
            // Tenant-wide access — explicitly clear OU scope
            conn.exec("SET LOCAL app.current_ou_paths = ''")

        return queryFn(ctx, conn)

    finally:
        // SET LOCAL resets automatically at transaction end
        releaseConnection(conn)
```

**Platform read-only access to tenant data** uses a separate mechanism: the platform support connection role has read-only grants on business tables but uses a special platform policy that bypasses the `app.current_tenant_id` check — it can read any tenant. This role is never issued to application code running tenant requests.

```sql
-- Platform read-only role policy (separate policy, not the tenant isolation policy)
CREATE POLICY platform_readonly ON <table>
    AS PERMISSIVE
    FOR SELECT
    USING (
        current_setting('app.actor_plane', true) = 'platform'
        AND current_setting('app.platform_tier', true) IN ('ROOT', 'PLATFORM_ADMIN', 'PLATFORM_SUPPORT')
    );
```

---

## 12. API Keys

API keys are a tenant-plane-only concept. Platform users authenticate via password + MFA exclusively — there are no platform API keys.

### 12.1 APIKey Entity

```
Entity: APIKey
─────────────────────────────────────────────────────────────────────
id              UUID
tenant_id       UUID            FK → Tenant (RLS-protected)
created_by      UUID            FK → TenantUser
name            VARCHAR(100)    "POS Integration — Maanzoni"
key_prefix      VARCHAR(8)      First 8 chars displayed in UI for identification
key_hash        TEXT            Argon2id hash; raw key never stored
permissions     TEXT[]          Explicit permission subset
org_unit_id     UUID?           Optional OU scope — key can only access this subtree
org_unit_path   LTREE?          Denormalized from OrgUnit for RLS injection
ip_allowlist    INET[]?         Restrict to specific IP ranges
expires_at      TIMESTAMPTZ?
last_used_at    TIMESTAMPTZ?
is_active       BOOL
created_at      TIMESTAMPTZ
```

**Key invariants:**
- An API key can never hold more permissions than the user who created it.
- An API key can never be scoped to an OU the creator cannot access.
- API keys carry their own `org_unit_path` injected into RLS — they are not implicitly scoped to the creator's current OU.

### 12.2 API Key Format

```
awo_{tenant_slug}_{32_random_bytes_base62}

Example:
awo_vivo-energy-africa_4xKpR7mNqZwLdTvHjBsY3nGcAuFe9i2X
```

The tenant slug in the key allows the authentication middleware to identify the tenant without a database lookup on the prefix alone.

### 12.3 API Key Authentication

```
function authenticateAPIKey(bearerToken):

    if not bearerToken.startsWith("awo_"):
        return nil, "Not an API key"

    parts = split(bearerToken, "_", maxParts: 3)
    tenantSlug = parts[1]
    rawKey = parts[2]

    tenant = resolveTenantBySlug(tenantSlug)
    assert tenant != nil and tenant.status == ACTIVE

    prefix = bearerToken[0:8]
    candidates = loadAPIKeysByTenantAndPrefix(tenant.id, prefix)

    for candidate in candidates:
        if argon2Verify(bearerToken, candidate.key_hash):
            assert now() < candidate.expires_at (if set)
            assert candidate.is_active
            assert request.ip in candidate.ip_allowlist (if set)

            updateLastUsed(candidate)

            // Synthetic session for downstream middleware
            return Session{
                tenant_id:      tenant.id,
                actor_type:     "api_key",
                permissions:    candidate.permissions,
                org_unit_paths: [candidate.org_unit_path],  // may be nil (tenant-wide)
            }

    return nil, "Invalid API key"
```

---

## 13. Audit Trail

All IAM and authorization events are recorded in an append-only audit log. Neither application code nor database maintenance jobs may UPDATE or DELETE audit rows. This constraint is enforced at the PostgreSQL role level — the application role has INSERT-only permission on audit tables.

### 13.1 AuditEvent Entity

```
Entity: AuditEvent
─────────────────────────────────────────────────────────────────────
id              UUID
plane           ENUM            PLATFORM | TENANT
tenant_id       UUID?           NULL for platform events
actor_id        UUID?           PlatformUser or TenantUser ID
actor_type      ENUM            user | service_account | api_key | system_root
actor_tier      ENUM?           Platform tier (for platform events)
action          VARCHAR(100)    e.g. "role.assigned", "user.login_failed"
resource_type   VARCHAR(100)    e.g. "tenant_user", "role"
resource_id     UUID?
org_unit_id     UUID?           OU context at time of action
changes         JSONB           { field: { before: X, after: Y } }
metadata        JSONB           { ip, user_agent, session_id, request_id }
outcome         ENUM            SUCCESS | FAILURE | DENIED
error_message   TEXT?
created_at      TIMESTAMPTZ     Immutable
```

### 13.2 Events That Must Always Be Audited

**Platform Events:**

| Action | Trigger |
|---|---|
| `platform.tenant.created` | Tenant provisioned |
| `platform.tenant.suspended` | Tenant suspended |
| `platform.tenant.archived` | Tenant archived |
| `platform.user.created` | Platform user added |
| `platform.user.login_success` | Platform login succeeded |
| `platform.user.login_failed` | Platform login failed |
| `platform.tenant_data.read` | Platform support reads tenant data |
| `platform.config.changed` | Platform configuration mutated |

**Tenant IAM Events:**

| Action | Trigger |
|---|---|
| `user.created` | New tenant user provisioned |
| `user.suspended` | User account suspended |
| `user.login_success` | Successful login |
| `user.login_failed` | Failed login attempt |
| `user.account_locked` | Lockout threshold reached |
| `user.password_changed` | Password updated |
| `user.mfa_enabled` | MFA activated |
| `user.mfa_disabled` | MFA deactivated |
| `role.created` | New role created |
| `role.modified` | Role permissions changed |
| `role.assigned` | Role granted to user |
| `role.revoked` | Role removed from user |
| `org_unit.created` | OU node added |
| `org_unit.modified` | OU attributes changed |
| `session.created` | Login session started |
| `session.revoked` | Session ended |
| `api_key.created` | API key issued |
| `api_key.revoked` | API key deactivated |
| `authorization.denied` | Permission check failed |

---

## 14. Cross-Cutting Request Flow

### 14.1 Tenant User Request — Full Trace

```
HTTP Request  →  POST /api/v1/fms/shifts/close
                 Host: vivo-energy-africa.awo.so
                 Authorization: Bearer <tenant_jwt>
                                │
┌──────────────────────────────────────────────────────┐
│ [1] TenantResolutionMiddleware                        │
│     Extract slug "vivo-energy-africa" from Host       │
│     Load tenant from cache                           │
│     Assert status == ACTIVE                          │
│     Set ctx.tenant_id                                │
└──────────────────────────────────────────────────────┘
                                │
┌──────────────────────────────────────────────────────┐
│ [2] AuthenticateTenantMiddleware                      │
│     Extract Bearer JWT                               │
│     Verify signature with TENANT_SIGNING_KEY         │
│     Assert claims.plane == "tenant"                  │
│     Assert claims.aud == request.host                │
│     Check Redis revocation set for claims.jti        │
│     Build session from claims                        │
│     Inject RLS settings on DB connection:            │
│       SET LOCAL app.current_tenant_id = <uuid>       │
│       SET LOCAL app.current_ou_paths = "VEA.KE.NBI.MZN" │
└──────────────────────────────────────────────────────┘
                                │
┌──────────────────────────────────────────────────────┐
│ [3] Authorize("fms.shifts.submit")                    │
│     Check "fms.shifts.submit" in session.permissions │
│     Found → proceed                                  │
│     Not found → 403 + audit record                   │
└──────────────────────────────────────────────────────┘
                                │
┌──────────────────────────────────────────────────────┐
│ [4] Handler → Service → Repository                    │
│     repository.CloseShift(ctx, shiftID)              │
│     Executes:                                        │
│       store.WithTenant(ctx, tenantID, fn)            │
│         UPDATE shifts SET status='CLOSED'            │
│         WHERE id = $1                                │
│     ← PostgreSQL RLS silently filters: shift must    │
│       have org_unit_path <@ "VEA.KE.NBI.MZN"        │
│       If not → query affects 0 rows → 404           │
└──────────────────────────────────────────────────────┘
                                │
┌──────────────────────────────────────────────────────┐
│ [5] AuditMiddleware (post-handler)                    │
│     Record "fms.shift.submitted" to audit log        │
└──────────────────────────────────────────────────────┘
                                │
                           200 OK  { shift: ... }
```

### 14.2 Platform Support Reading Tenant Data

```
HTTP Request  →  GET /platform/api/v1/tenants/vivo-energy-africa/users
                 Host: platform.awo.so
                 Authorization: Bearer <platform_jwt>
                                │
┌──────────────────────────────────────────────────────┐
│ [1] Platform Route Guard                              │
│     Assert Host == "platform.awo.so"                 │
│     (No tenant resolution — platform routes only)    │
└──────────────────────────────────────────────────────┘
                                │
┌──────────────────────────────────────────────────────┐
│ [2] AuthenticatePlatformMiddleware                    │
│     Verify JWT with PLATFORM_SIGNING_KEY             │
│     Assert claims.plane == "platform"                │
│     Assert requestIP in PLATFORM_ADMIN_IPS           │
│     Build platform session                           │
│     Set ctx.platform_tier = PLATFORM_SUPPORT         │
│     Inject RLS settings:                             │
│       SET LOCAL app.actor_plane = 'platform'         │
│       SET LOCAL app.platform_tier = 'PLATFORM_SUPPORT'│
└──────────────────────────────────────────────────────┘
                                │
┌──────────────────────────────────────────────────────┐
│ [3] AuthorizePlatform("platform.tenants.read")        │
│     Check tier PLATFORM_SUPPORT against matrix       │
│     "tenant_read.*" covers "platform.tenants.read"   │
│     → proceed                                        │
└──────────────────────────────────────────────────────┘
                                │
┌──────────────────────────────────────────────────────┐
│ [4] Handler reads tenant user list                    │
│     Uses platform read-only DB connection role       │
│     Platform RLS policy allows cross-tenant read     │
│     Returns user list (PII fields may be masked      │
│     depending on platform_tier)                      │
└──────────────────────────────────────────────────────┘
                                │
┌──────────────────────────────────────────────────────┐
│ [5] Audit: "platform.tenant_data.read"                │
│     actor: platform support user                     │
│     resource: tenant vivo-energy-africa / users      │
└──────────────────────────────────────────────────────┘
```

### 14.3 Role Assignment → Session Invalidation

```
Admin grants "Country Manager" role to User X scoped to VEA.KE
                                │
RoleService.AssignRole(adminSession, userX.id, countryManagerRole.id, orgUnit: VEA.KE)
                                │
├── Assert admin has "iam.roles.write" permission
├── Assert admin's own OU path covers VEA.KE (can't grant beyond own scope)
├── INSERT UserRoleAssignment (userX, countryManagerRole, org_unit: VEA.KE)
├── Record audit event "role.assigned"
└── Invalidate User X's permission snapshot:
        cache.delete("permissions:" + tenantID + ":" + userX.id)
        // User X's current access tokens remain valid until expiry (15 min max)
        // Next token refresh recomputes permissions with new role included
        // For immediate effect, revoke all User X sessions:
        //   SessionService.RevokeAll(userX.id, tenantID, "role_changed")
```

---

## 15. Entity Relationship Summary

```
                        ┌─────────────────────┐
                        │   PlatformConfig     │  (from env/config; read-only)
                        │   (ROOT credentials) │
                        └──────────┬──────────┘
                                   │ creates
                        ┌──────────▼──────────┐
                        │    PlatformUser      │
                        │  tier: ROOT/ADMIN    │
                        │        /SUPPORT      │
                        │        /READONLY     │
                        └──────────┬──────────┘
                                   │ manages
        ┌──────────────────────────▼──────────────────────────┐
        │                       Tenant                         │
        │  slug · status · plan · feature_flags               │
        └──┬────────────────────┬──────────────────────┬──────┘
           │                    │                       │
           │ (1:∞)              │ (1:∞)                 │ (1:∞)
           ▼                    ▼                       ▼
      OrgUnit             TenantUser               SSOConfig
   (ltree tree)        status · actor_type
        │                      │
        │                      │ (∞:∞ via)
        │               UserRoleAssignment ──────► Role
        │                 org_unit_id?                │
        │                 expires_at?            parent_role_id?
        │                                             │
        │                                        Permission[]
        │                                     permission · condition
        │                                        scope_type
        │
        │  (org_unit_path baked into)
        ▼
   TenantSession ─────────────────────────────────────────────
     permissions[]  (snapshot)
     org_unit_paths[]  (snapshot)
        │
        ├── derives access token (JWT, 15 min)
        └── rotates refresh token (7 days)

   APIKey ──► (tenant_id, permissions[], org_unit_path?)
           └── creates synthetic Session on auth

   AuditEvent ◄── all mutations in both planes (append-only)
```

### Key Cardinalities

| Relationship | Cardinality |
|---|---|
| Platform → Tenants | 1 : ∞ |
| Tenant → TenantUsers | 1 : ∞ |
| Tenant → OrgUnits | 1 : ∞ (tree) |
| OrgUnit → OrgUnit (parent) | ∞ : 1 (optional) |
| TenantUser → Roles (via assignments) | ∞ : ∞ |
| UserRoleAssignment → OrgUnit | ∞ : 1 (optional) |
| Role → Role (parent) | ∞ : 1 (optional) |
| Role → Permissions | 1 : ∞ |
| TenantUser → Sessions | 1 : ∞ |
| TenantUser → APIKeys | 1 : ∞ |
| PlatformUser → PlatformSessions | 1 : ∞ |
| All mutations → AuditEvent | 1 : ∞ |

---

## Appendix A — Implementation Checklist

### Platform Plane
- [ ] Root credentials loaded from sealed secrets at startup — never written to DB
- [ ] Separate `platform_users` table outside any tenant schema
- [ ] Separate JWT signing key for platform tokens (`kid: platform-*`)
- [ ] Platform routes behind separate subdomain (`platform.awo.so`)
- [ ] IP allowlist enforced before credential check for platform logins
- [ ] MFA enforced for ROOT and PLATFORM_ADMIN at middleware level (not policy level)
- [ ] Platform read-only DB role with separate RLS policy (cross-tenant read, no write)
- [ ] `platform.tenant_data.read` audit event on every support data access
- [ ] Platform refresh token TTL: 4 hours (shorter than tenant)

### Tenant Plane — Database
- [ ] `ENABLE ROW LEVEL SECURITY` and `FORCE ROW LEVEL SECURITY` on all business tables
- [ ] `ltree` extension enabled for OU path queries
- [ ] `org_unit_path` indexed with GiST index for `<@` operator performance
- [ ] Composite index `(tenant_id, email)` on `tenant_users`
- [ ] Composite index `(tenant_id, slug)` on `roles`
- [ ] Composite index `(tenant_id, user_id)` on `user_role_assignments`
- [ ] Audit tables: INSERT-only permission for application role
- [ ] `mfa_secret` and SSO `client_secret` encrypted at rest with per-tenant key

### Tenant Plane — Service Layer
- [ ] All permission checks via `Authorize(permission_string)` middleware, not ad-hoc
- [ ] `WithTenant(ctx, tenantID, fn)` wrapping all repository calls
- [ ] OU path injected into DB connection alongside tenant ID
- [ ] Permission snapshot cached in Redis: `permissions:{tenantID}:{userID}`
- [ ] Cache invalidated on role assignment change
- [ ] Argon2id parameters: t=3, m=65536, p=4 (minimum)
- [ ] TOTP window tolerance: ±1 period (accepts codes ±30 seconds)
- [ ] Backup codes: 10 single-use, bcrypt-hashed, consumed on use
- [ ] Refresh token rotation on every use (invalidate old, issue new)
- [ ] Rate limiting on `/auth/login` (10 req/min per IP per tenant)
- [ ] Rate limiting on `/auth/mfa/verify` (5 req/min per user)

### Audit
- [ ] Both planes write to audit log before sending HTTP response
- [ ] Audit table `created_at` is set by DB server (`DEFAULT NOW()`), not application
- [ ] High-sensitivity events (root login, tenant suspension, MFA disabled) trigger alerts
- [ ] Audit retention: minimum 7 years for any tenant with financial data

---

## Appendix B — Frappe Concept Mapping

| Frappe Concept | AwoERP Equivalent |
|---|---|
| System Manager | Built-in `tenant-admin` role |
| Administrator user | `platform-root` — platform plane only, never a tenant user |
| DocType permissions | `module.resource.*` permission strings |
| Permission Level 0–9 | `scope_type`: tenant → org_unit → record |
| Role | `Role` entity with optional `parent_role_id` inheritance |
| Role Profile | Pre-composed `Role` with curated child permissions |
| User Type (System / Website) | `actor_type`: user vs service_account |
| User Permission (record filter) | CEL condition on `Permission.condition` field + RLS policy |
| `frappe.only_if_creator` | CEL: `resource.created_by == subject.user_id` |
| Document sharing | `permission.share` action + `ShareRecord` entity |
| Workspace | AMIS page schemas in `web/schemas/pages/` |
| Company (multi-company) | OrgUnit of type SUBSIDIARY within a single Tenant |
| Branch | OrgUnit of type BRANCH |
| `frappe.get_list` with User Permissions | DB query filtered by RLS `org_unit_path <@` policy |
| Session Document | `TenantSession` entity |
| OAuth Client | `SSOConfig` entity |

---

*End of Document — AwoERP IAM Guide v2.0*
