[<-- Back to Index](README.md)

## Domain Model — 4 Actor Types

### Overview

The authz module defines four actor types, each with its own subject prefix and domain. This namespacing is what makes cross-domain leakage impossible at the policy level.

```markdown
Actor       Subject prefix       Domain                    Trust Level
─────────────────────────────────────────────────────────────────────
Platform    "platform:{userID}"  "_platform_"              Highest
Tenant      "tenant:{userID}"    "{tenantID}"              High
Portal      "portal:{userID}"    "{tenantID}:portal"       Medium
API         "api:{clientID}"     "{tenantID}:api"          Configurable
```

### Actor 1: Platform

**Who they are**: AWO ERP platform operators — engineers, support staff, and automated platform services.

**Subject format**: `platform:usr_00000001`

**Domain**: `_platform_` (the reserved constant `DomainPlatform`)

**What they can access**:
```markdown
PLATFORM DOMAIN CAPABILITIES:
  ├── Manage all tenants (provision, suspend, archive)
  ├── Read any tenant's data via cross-tenant wildcard objects
  │   Example: policy object "tenant/*/invoice" covers all tenants
  ├── Modify platform-wide settings (modules, feature flags)
  ├── Grant and revoke roles within any tenant domain
  └── View audit logs across all tenants

WHAT PLATFORM ACTORS CANNOT DO IN TENANT DOMAIN:
  → Platform policies are in "_platform_" domain
  → They do NOT automatically apply in "tenant-abc" domain
  → Platform admins must be explicitly granted in tenant domain
    if they need to impersonate a tenant user
  → This separation is deliberate — platform ≠ super-tenant
```

**Go helpers**:
```go
sub := authz.PlatformSubject("usr_00000001")  // "platform:usr_00000001"
dom := authz.DomainPlatform                    // "_platform_"
```

---

### Actor 2: Tenant

**Who they are**: The primary business users — employees of the company that subscribed to AWO ERP.

**Subject format**: `tenant:usr_a1b2c3d4`

**Domain**: `{tenantID}` — the tenant's UUID, e.g. `"a1b2c3d4-..."`

**What they can access**:
```markdown
TENANT DOMAIN CAPABILITIES:
  ├── All business operations within their tenant
  │   invoice/*, customer/*, order/*, payment/*
  ├── Roles: finance-manager, sales-rep, hr-admin, etc.
  ├── Can be granted portal admin rights
  │   → Which lets them manage portal user policies
  └── Cannot access _platform_ or other tenants' domains

TYPICAL TENANT ROLES:
  role:tenant-admin          → full access within tenant domain
  role:finance-manager       → all finance resources
  role:finance-viewer        → read-only finance
  role:sales-manager         → all sales + discount approval
  role:sales-rep             → create orders, view customers
  role:hr-admin              → payroll, employee records
  role:auditor               → read-only, all modules (time-limited)
```

**Go helpers**:
```go
sub := authz.TenantSubject("usr_a1b2c3d4")       // "tenant:usr_a1b2c3d4"
dom := authz.TenantDomain("a1b2c3d4-tenant-uuid") // "a1b2c3d4-tenant-uuid"
```

---

### Actor 3: Portal

**Who they are**: External users — customers, suppliers, or partners who access a limited self-service portal provided by the tenant.

**Subject format**: `portal:usr_c4d5e6f7`

**Domain**: `{tenantID}:portal` — e.g. `"a1b2c3d4-...:portal"`

**What they can access**:
```markdown
PORTAL DOMAIN CAPABILITIES:
  ├── View their own invoices and statements
  ├── Submit payment requests
  ├── Track their own orders
  ├── Access shared documents uploaded for them
  └── CANNOT access any tenant-internal data

KEY ISOLATION PROPERTY:
  Portal domain is "{tenantID}:portal"
  Tenant domain is "{tenantID}"

  A portal user CANNOT call Enforce() in the tenant domain.
  A tenant user CANNOT call Enforce() in the portal domain
  (unless explicitly cross-granted).

  This prevents a compromised portal user from accessing
  internal ERP operations.

TYPICAL PORTAL ROLES:
  role:portal-customer       → view own invoices, download statements
  role:portal-supplier       → view POs, submit invoices, track payments
  role:portal-auditor        → read-only access granted by tenant
```

**Go helpers**:
```go
sub := authz.PortalSubject("usr_c4d5e6f7")        // "portal:usr_c4d5e6f7"
dom := authz.PortalDomain("a1b2c3d4-tenant-uuid")  // "a1b2c3d4-tenant-uuid:portal"
```

---

### Actor 4: API

**Who they are**: Machine clients — third-party integrations, mobile apps, and programmatic access clients using API keys or service account tokens.

**Subject format**: `api:cli_x7y8z9w0`

**Domain**: `{tenantID}:api` — e.g. `"a1b2c3d4-...:api"`

**What they can access**:
```markdown
API DOMAIN CAPABILITIES:
  ├── Defined entirely by the tenant's API policies
  ├── Can be scoped narrowly: "read invoices only"
  ├── Can be scoped broadly: full ERP access (service accounts)
  └── Tenant admin controls all API client policies

KEY ISOLATION PROPERTY:
  API clients are in "{tenantID}:api"
  A compromised API key never gets tenant-user rights
  unless explicitly cross-granted.

TYPICAL API ROLES:
  role:api-readonly          → GET only, all resources
  role:api-invoice-submit    → POST invoice/*, read customer/*
  role:api-full-access       → all actions (trusted service accounts)
  role:api-webhook-consumer  → read event streams only
```

**Go helpers**:
```go
sub := authz.APISubject("cli_x7y8z9w0")          // "api:cli_x7y8z9w0"
dom := authz.APIDomain("a1b2c3d4-tenant-uuid")    // "a1b2c3d4-tenant-uuid:api"
```

---

### Domain Hierarchy Visual

```markdown
DOMAIN STRUCTURE:

_platform_
├── platform:admin-1        (platform admin)
├── platform:support-bot    (automated platform service)
└── platform:migration-svc  (data migration service)

a1b2c3d4-tenant-uuid  (Tenant: Savannah Electronics)
├── tenant:usr-001          (CEO — role:tenant-admin)
├── tenant:usr-002          (CFO — role:finance-manager)
├── tenant:usr-003          (Sales Rep — role:sales-rep)
└── tenant:usr-004          (Auditor — role:auditor, expires 2026-03-31)

a1b2c3d4-tenant-uuid:portal  (Portal)
├── portal:cust-001         (Customer: Nairobi Electronics Ltd)
└── portal:supp-001         (Supplier: Mombasa Parts Co.)

a1b2c3d4-tenant-uuid:api  (API Clients)
├── api:cli-mobile-app      (Tenant's mobile app)
└── api:cli-accounting-sync (Accounting software integration)

c5d6e7f8-tenant-uuid  (Tenant: Kenya Imports Ltd — completely separate)
├── tenant:usr-101          ...
└── ...
```

### Principal in HTTP Context

The authn middleware (JWT validation, not part of authz package) extracts the actor and stores it for authz to consume:

```go
// authn middleware sets this after JWT validation:
c.Locals(authz.LocalsKeyPrincipal, authz.Principal{
    Subject: authz.TenantSubject(claims.UserID),
    Domain:  authz.TenantDomain(claims.TenantID),
})

// authz.Middleware then reads it:
p, ok := c.Locals(authz.LocalsKeyPrincipal).(authz.Principal)
// → p.Subject = "tenant:usr_a1b2c3d4"
// → p.Domain  = "a1b2c3d4-tenant-uuid"
```

### Summary Table

| Actor | Subject | Domain | Typical Use |
|-------|---------|--------|-------------|
| Platform | `platform:{id}` | `_platform_` | Ops, support, automated services |
| Tenant | `tenant:{id}` | `{tenantID}` | All ERP business users |
| Portal | `portal:{id}` | `{tenantID}:portal` | Customers, suppliers, partners |
| API | `api:{id}` | `{tenantID}:api` | Integrations, mobile apps |

---

Next: [Casbin Policy Engine](./05-casbin-policy-engine.md)
