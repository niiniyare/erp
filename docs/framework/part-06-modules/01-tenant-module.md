---
title: "Platform Module: Tenant & Organisation"
part: "Part VI — Platform Entities"
chapter: 42
section: "platform-tenant-module"
related:
  - "[Chapter 41: Platform Module Overview](./platform-module-overview.md)"
  - "[Chapter 43: IAM Module](./platform-iam-module.md)"
  - "[Chapter 11: Database Migrations](../part-02-entity-system/11-migrations.md)"
  - "[Chapter 38: Tenant Lifecycle](../part-07-multitenancy/38-tenant-lifecycle.md)"
---

# Chapter 42 — Tenant & Organisation Module

> **Primary source for:** multi-tenancy isolation model, the `tenants` and `org_nodes` tables, org hierarchy design, tenant lifecycle, and how RLS is applied.
>
> **Audience:** new developers joining the platform team, stakeholders who want to understand how data isolation works, and any developer whose module needs to understand *why* `tenant_id` is on every table.

---

## 42.1 The Big Picture: What Is a Tenant?

Before touching any code, build the right mental model.

A **tenant** is one business running on the Awo ERP platform. When a fuel station chain signs up, they become one tenant. When a hospital signs up, they become another tenant. The platform runs both — at the same time, on the same database server — without either business ever seeing the other's data.

This is **multi-tenancy**: many tenants sharing one platform.

### Why Multi-Tenancy?

The alternative is running a separate database (or even a separate server) for every customer. That approach is called single-tenancy. It sounds safe, but it is operationally catastrophic at scale:

- 500 customers = 500 databases to back up, patch, monitor, and upgrade.
- A security fix must be deployed 500 times.
- Capacity planning is impossible — some databases sit empty while others burst.

Multi-tenancy solves all of this. One deployment, one database, one place to apply updates. Customers share infrastructure but **never share data**.

### The Apartment Building Analogy

Think of the Awo platform as an apartment building:

- The building is the **database**. All tenants share it.
- Each apartment is a **tenant's data space**. Private, locked.
- The front door key is **Row-Level Security (RLS)** — each tenant can only open their own door.
- The building management (platform admins) have a master key but use it only for maintenance.

The lock in PostgreSQL is an RLS policy. Every table that contains tenant data has this policy attached:

```sql
CREATE POLICY tenant_isolation ON some_table
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);
```

Before every query, the application calls:

```sql
SELECT set_tenant_context('the-tenant-uuid');
```

PostgreSQL then silently appends `WHERE tenant_id = 'the-tenant-uuid'` to every query on that table. Even if a bug in Go code forgot to filter by tenant, the database itself enforces it.

> **Security property:** Cross-tenant data leakage is structurally impossible at the database layer. It is not just "we wrote the queries correctly" — it is "the database engine prevents it."

---

## 42.2 What Is an OrgNode?

Within one tenant (one business), there can be many **organisational units**: headquarters, regions, branches, departments, cost centres.

An **OrgNode** (Organisation Node) models this internal hierarchy. It is a tree:

```
Acme Petroleum Ltd (COMPANY)
├── Nairobi Region (DIVISION)
│   ├── Westlands Branch (BRANCH)
│   │   ├── Fuel Sales (DEPARTMENT)
│   │   └── Lubes Store (DEPARTMENT)
│   └── Karen Branch (BRANCH)
└── Mombasa Region (DIVISION)
    └── Mombasa Port Branch (BRANCH)
```

### Why OrgNodes Matter for Access Control

A cashier at Westlands Branch should see Westlands transactions — not Karen Branch, not Mombasa. An OrgNode-scoped entity (like an invoice or a stock transaction) carries an `org_unit_id` column. The framework's `AllowWithinOrgScope` policy checks whether the viewer's org unit is an ancestor-or-equal of the record's org unit.

This means:
- **Cashier at Westlands** → sees Westlands records only.
- **Nairobi Region Manager** → sees all Nairobi branches (Westlands + Karen).
- **CEO (tenant-wide)** → sees everything.
- **Another tenant's CEO** → sees nothing. RLS blocks at the DB level before org-scope logic even runs.

OrgNodes are also used in **Settings** (branch-level config overrides) and **Reporting** (filter by org unit).

---

## 42.3 Directory Layout

```
internal/platform/tenant/
├── tenant.go          ← init() — registers TenantDef and OrgNodeDef
├── definition.go      ← EntityDefinition declarations
├── policy.go          ← allowTenantViewer, requireRole helpers
├── hooks.go           ← validateTenantStatusTransition, materialisePath, etc.
├── service.go         ← ProvisionService, SuspendService (thin wrappers)
└── migrations/
    ├── 20240101000001_create_tenants.up.sql
    └── 20240101000001_create_tenants.down.sql
```

---

## 42.4 The Tenant EntityDefinition

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

    // ── CRITICAL: ScopeLevelGlobal ────────────────────────────────────────────
    // The tenants table has NO tenant_id column and NO RLS policy.
    //
    // Why? A tenant IS the isolation boundary. It cannot itself be isolated by
    // tenant_id — that would be circular. "This row belongs to tenant X" makes
    // no sense when this row IS tenant X.
    //
    // Access to the tenants table is controlled entirely at the application layer
    // via Policies below (system-only). Not at the database layer.
    OrgScope: org.ScopeLevelGlobal,

    SoftDelete: false, // Terminal state = ARCHIVED. No row deletion needed.
    Audited:    true,  // Every change to a tenant record is audited.

    Fields: []*definition.FieldDef{
        // ── Identity ──────────────────────────────────────────────────────────
        {
            Name:     "slug",
            Type:     definition.FieldTypeData,
            Label:    "Slug",
            Required: true,
            Description: "URL-safe unique identifier: 'acme-petroleum'. " +
                "Used in subdomain routing: acme-petroleum.awoerp.com. " +
                "Set at creation; never changed after provisioning.",
        },
        {
            Name:     "name",
            Type:     definition.FieldTypeData,
            Label:    "Display Name",
            Required: true,
            Description: "Full business name shown in UI: 'Acme Petroleum Ltd'",
        },
        {
            Name:        "domain",
            Type:        definition.FieldTypeData,
            Label:       "Custom Domain",
            Description: "Optional vanity domain: 'erp.acmepetroleum.co.ke'",
        },

        // ── Subscription ──────────────────────────────────────────────────────
        {
            Name:     "plan",
            Type:     definition.FieldTypeSelect,
            Label:    "Subscription Plan",
            Options:  []string{"starter", "growth", "enterprise"},
            Required: true,
            Description: "Controls resource limits: max_users, max_entities, " +
                "storage_quota, api_rate_limit. Checked by the quota service.",
        },
        {
            Name:     "status",
            Type:     definition.FieldTypeSelect,
            Label:    "Lifecycle Status",
            Options:  []string{"PENDING", "ACTIVE", "SUSPENDED", "ARCHIVED"},
            Required: true,
            Description: "See §42.6 for the full status machine and allowed transitions.",
        },

        // ── Kenya Business Identity ────────────────────────────────────────────
        // Kenya-specific fields are first-class, not custom fields.
        // The KRA PIN is required for eTIMS invoice submission (VAT compliance).
        {Name: "kra_pin",        Type: definition.FieldTypeData, Label: "KRA PIN"},
        {Name: "company_reg_no", Type: definition.FieldTypeData, Label: "Companies Registry No"},

        // ── Locale Defaults ───────────────────────────────────────────────────
        // These are DEFAULTS for new users and documents within this tenant.
        // Individual users or org nodes may override via the Settings module.
        {Name: "timezone",        Type: definition.FieldTypeData, Label: "Default Timezone",
            Default: "Africa/Nairobi"},
        {Name: "currency",        Type: definition.FieldTypeData, Label: "Base Currency",
            Default: "KES"},
        {Name: "fiscal_year_end", Type: definition.FieldTypeData, Label: "Fiscal Year End",
            Default:     "06-30",
            Description: "MM-DD format. Kenya default: 30 June (July–June fiscal year)."},

        // ── Lifecycle Timestamps ──────────────────────────────────────────────
        // Stored on the tenant row so queries like "show all tenants suspended
        // in the last 30 days" are cheap (no JOIN needed).
        {Name: "trial_ends_at",  Type: definition.FieldTypeDateTime, Label: "Trial Ends At"},
        {Name: "provisioned_at", Type: definition.FieldTypeDateTime, Label: "Provisioned At"},
        {Name: "suspended_at",   Type: definition.FieldTypeDateTime, Label: "Suspended At"},
        {Name: "suspend_reason", Type: definition.FieldTypeSmallText, Label: "Suspension Reason"},
        {Name: "archived_at",    Type: definition.FieldTypeDateTime, Label: "Archived At"},
    },

    // ── Policies ──────────────────────────────────────────────────────────────
    // The tenants table is not protected by RLS. Policies here are the ONLY
    // access control. They must be correct — there is no database backstop.
    Policies: []definition.PolicyDef{
        // System callers: provisioning Temporal workflows, platform admin API,
        // billing service. They get full access with no further questions.
        definition.Policy(definition.OpAll, definition.AllowSystem),

        // Everyone else — including authenticated tenant users — is denied.
        // A user inside Acme Petroleum cannot read or modify the tenant row.
        // They interact with their own data via modules; they do not need
        // raw access to the tenants table.
        definition.Policy(definition.OpAll, definition.DenyAll),
    },

    Hooks: []definition.HookDef{
        {
            // Runs BEFORE an UPDATE is committed.
            // Validates that the status transition is legal.
            // See §42.6 for the allowed transitions.
            Name: "validate_status_transition",
            Ops:  definition.OpUpdate,
            When: definition.HookBefore,
            Fn:   validateTenantStatusTransition,
        },
        {
            // Runs AFTER a successful UPDATE is committed.
            // Detects status changing to SUSPENDED and revokes all active sessions
            // for this tenant. Users are immediately logged out.
            Name: "invalidate_sessions_on_suspend",
            Ops:  definition.OpUpdate,
            When: definition.HookAfter,
            Fn:   invalidateSessionsOnSuspend,
        },
    },
}
```

---

## 42.5 The OrgNode EntityDefinition

```go
var OrgNodeDef = definition.EntityDefinition{
    Name:     "org_node",
    Label:    "Organisation Node",
    Module:   "Platform",
    Table:    "org_nodes",
    OrgScope: org.ScopeLevelTenant, // each org tree belongs to one tenant
    Audited:  true,

    Fields: []*definition.FieldDef{
        {
            Name:         "parent_id",
            Type:         definition.FieldTypeLink,
            Label:        "Parent Node",
            TargetEntity: "org_node",
            Description:  "Nil for the root COMPANY node. Only one root per tenant.",
        },
        {
            Name:     "type",
            Type:     definition.FieldTypeSelect,
            Label:    "Node Type",
            Options:  []string{"COMPANY", "DIVISION", "DEPARTMENT", "BRANCH", "COST_CENTRE"},
            Required: true,
        },
        {Name: "code", Type: definition.FieldTypeData, Label: "Code", Required: true,
            Description: "Short code, unique within tenant: 'NBI-WEST', 'HQ'"},
        {Name: "name", Type: definition.FieldTypeData, Label: "Name", Required: true},
        {
            // The materialised path stores the full ancestry as a string.
            // Example: "hq/nairobi/westlands"
            // Set automatically by the materialisePath BeforeHook.
            // Never set manually by callers.
            Name:        "path",
            Type:        definition.FieldTypeData,
            Label:       "Materialised Path",
            Description: "Computed. Do not set manually.",
        },
        {Name: "is_active", Type: definition.FieldTypeBool, Label: "Active", Default: "true"},
    },

    Policies: []definition.PolicyDef{
        // System callers can always read/write org nodes.
        definition.Policy(definition.OpAll, definition.AllowSystem),

        // Any authenticated tenant user can READ org nodes.
        // Reason: users need to see their own org hierarchy for:
        //   - Navigation ("which branch am I operating as?")
        //   - Document creation ("which branch does this invoice belong to?")
        //   - Reporting ("show me all records for Nairobi Region")
        definition.Policy(definition.OpRead, allowTenantViewer),

        // Only admins can CREATE, UPDATE, or DELETE org nodes.
        // Reason: restructuring the org hierarchy affects everyone.
        // A cashier cannot accidentally reparent their branch.
        definition.Policy(definition.OpWrite|definition.OpDelete, requireRole("admin")),
    },

    Hooks: []definition.HookDef{
        {
            // Runs BEFORE create or update.
            // Recomputes the "path" field whenever a node is created or its
            // parent changes. See §42.7 for the path algorithm.
            Name: "materialise_path",
            Ops:  definition.OpCreate | definition.OpUpdate,
            When: definition.HookBefore,
            Fn:   materialisePath,
        },
    },
}
```

---

## 42.6 Policy Helpers

```go
// internal/platform/tenant/policy.go
package tenant

import (
    "context"
    "awo.so/framework/definition"
)

// allowTenantViewer grants read access to any viewer with a valid tenant context.
// Returns ErrDeny for anonymous (unauthenticated) requests.
// Used for entities that all tenant users legitimately need to see.
func allowTenantViewer(_ context.Context, v definition.ViewerContext, _ definition.Op, _ definition.Record) error {
    if v.TenantID() == "" {
        return definition.ErrDeny // unauthenticated
    }
    return definition.ErrAllow
}

// requireRole returns a PolicyFunc that allows access only when the viewer
// holds the named role (or is a system caller).
//
// Usage:
//   definition.Policy(definition.OpWrite, requireRole("admin"))
//   definition.Policy(definition.OpDelete, requireRole("finance_manager"))
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

## 42.7 The Materialised Path — How OrgNode Trees Work Fast

### The Problem with Recursive Queries

A naïve tree stored as `(id, parent_id)` requires a recursive Common Table Expression (CTE) to find all descendants of a node:

```sql
-- This works but is slow on deep trees with many nodes.
WITH RECURSIVE subtree AS (
    SELECT id FROM org_nodes WHERE id = $1
    UNION ALL
    SELECT n.id FROM org_nodes n
    JOIN subtree s ON n.parent_id = s.id
)
SELECT * FROM some_entity WHERE org_unit_id IN (SELECT id FROM subtree);
```

For a query on every invoice in a region's subtree, this recursive CTE runs for every request. Expensive.

### The Materialised Path Solution

Each node stores its full ancestry as a slash-separated string in the `path` column:

| Node | path |
|---|---|
| Acme Petroleum Ltd (root) | `hq` |
| Nairobi Region | `hq/nairobi` |
| Westlands Branch | `hq/nairobi/westlands` |
| Karen Branch | `hq/nairobi/karen` |
| Mombasa Region | `hq/mombasa` |

Finding all nodes under Nairobi Region becomes a single `LIKE` query:

```sql
SELECT * FROM org_nodes WHERE path LIKE 'hq/nairobi/%' OR path = 'hq/nairobi';
```

No recursion. No CTE. One index scan.

### The materialisePath Hook

When an org node is created or its `parent_id` changes, the `materialisePath` BeforeHook computes the new path:

```go
// internal/platform/tenant/hooks.go
func materialisePath(ctx context.Context, v definition.ViewerContext, op definition.Op, rec definition.MutableRecord) error {
    parentID, _ := rec.Get("parent_id")

    if parentID == nil {
        // Root node — path is just the node's own code.
        code, _ := rec.Get("code")
        rec.Set("path", slugify(code.(string)))
        return nil
    }

    // Load parent's path from DB.
    parent, err := orgNodeStore.FindByID(ctx, parentID.(uuid.UUID))
    if err != nil {
        return fmt.Errorf("materialisePath: parent not found: %w", err)
    }

    parentPath, _ := parent.Get("path")
    code, _ := rec.Get("code")
    rec.Set("path", parentPath.(string)+"/"+slugify(code.(string)))
    return nil
}
```

> **Important:** If a node's parent changes, the paths of ALL its descendants also become stale. The hook must recursively update descendant paths. This is done in a single SQL UPDATE:
>
> ```sql
> UPDATE org_nodes
>    SET path = $new_prefix || substring(path FROM length($old_prefix)+1)
>  WHERE path LIKE $old_prefix || '/%';
> ```

---

## 42.8 Tenant Status Machine

The `status` field follows strict allowed transitions. The `validateTenantStatusTransition` BeforeHook enforces them:

```
PENDING ──────────────► ACTIVE ◄──────── SUSPENDED
                           │                   ▲
                           │                   │
                           └──────────────► SUSPENDED
                           │
                           └──────────────► ARCHIVED
```

| From | To | Allowed? | Reason |
|---|---|---|---|
| PENDING | ACTIVE | ✅ | Normal provisioning completion |
| PENDING | SUSPENDED | ❌ | Can't suspend before activation |
| PENDING | ARCHIVED | ✅ | Cancel before activation |
| ACTIVE | SUSPENDED | ✅ | Billing failure, ToS violation |
| ACTIVE | ARCHIVED | ✅ | Offboarding |
| SUSPENDED | ACTIVE | ✅ | Payment resolved, reactivation |
| SUSPENDED | ARCHIVED | ✅ | Permanent closure |
| ARCHIVED | anything | ❌ | Terminal state, no way back |

```go
// internal/platform/tenant/hooks.go

var allowedTransitions = map[string][]string{
    "PENDING":   {"ACTIVE", "ARCHIVED"},
    "ACTIVE":    {"SUSPENDED", "ARCHIVED"},
    "SUSPENDED": {"ACTIVE", "ARCHIVED"},
    "ARCHIVED":  {}, // terminal
}

func validateTenantStatusTransition(ctx context.Context, v definition.ViewerContext, op definition.Op, rec definition.MutableRecord) error {
    newStatus, _ := rec.Get("status")
    if newStatus == nil {
        return nil // not changing status
    }

    current, err := tenantStore.FindByID(ctx, rec.ID())
    if err != nil {
        return err
    }
    currentStatus, _ := current.Get("status")

    allowed := allowedTransitions[currentStatus.(string)]
    for _, s := range allowed {
        if s == newStatus.(string) {
            return nil // valid transition
        }
    }

    return fmt.Errorf("invalid tenant status transition: %s → %s",
        currentStatus, newStatus)
}
```

### What Suspension Does

When status changes to SUSPENDED, the `invalidateSessionsOnSuspend` AfterHook:

1. Calls `iam.RevokeAllForTenant(ctx, tenantID)` — bulk-updates all session rows with `revoked_at = now()`.
2. Clears the Redis session cache keys for this tenant.
3. All in-flight requests with this tenant's sessions receive 401 on their next DB touch (RLS blocks all queries when the tenant context is set but the tenant status is SUSPENDED — see §42.10 for the `set_tenant_context` stored procedure).

From that moment, users are logged out. New login attempts fail because the IAM login service checks tenant status before issuing a session.

---

## 42.9 Migration

```sql
-- internal/platform/tenant/migrations/20240101000001_create_tenants.up.sql

-- ── Tenants ────────────────────────────────────────────────────────────────
-- No RLS. No tenant_id column. This is the root table.
CREATE TABLE tenants (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    slug            text        NOT NULL UNIQUE,   -- URL-safe, immutable after provisioning
    name            text        NOT NULL,
    domain          text        UNIQUE,             -- optional vanity domain
    plan            text        NOT NULL DEFAULT 'starter'
                                CHECK (plan IN ('starter', 'growth', 'enterprise')),
    status          text        NOT NULL DEFAULT 'PENDING'
                                CHECK (status IN ('PENDING', 'ACTIVE', 'SUSPENDED', 'ARCHIVED')),

    kra_pin         text,
    company_reg_no  text,

    timezone        text        NOT NULL DEFAULT 'Africa/Nairobi',
    currency        text        NOT NULL DEFAULT 'KES',
    fiscal_year_end text        NOT NULL DEFAULT '06-30',

    trial_ends_at   timestamptz,
    provisioned_at  timestamptz,
    suspended_at    timestamptz,
    suspend_reason  text,
    archived_at     timestamptz,

    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

-- The set_updated_at trigger function is created in the framework's base migration.
-- It sets updated_at = now() on every UPDATE.
CREATE TRIGGER set_updated_at
    BEFORE UPDATE ON tenants
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ── set_tenant_context stored procedure ────────────────────────────────────
-- Called by the application at the start of every authenticated request.
-- Sets the PostgreSQL session variable used by ALL RLS policies.
-- Also validates that the tenant is ACTIVE — suspended/archived tenants are blocked here.
CREATE OR REPLACE FUNCTION set_tenant_context(p_tenant_id uuid)
RETURNS void AS $$
DECLARE
    v_status text;
BEGIN
    SELECT status INTO v_status FROM tenants WHERE id = p_tenant_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'tenant_not_found' USING ERRCODE = 'P0001';
    END IF;

    IF v_status = 'SUSPENDED' THEN
        RAISE EXCEPTION 'tenant_suspended' USING ERRCODE = 'P0002';
    END IF;

    IF v_status = 'ARCHIVED' THEN
        RAISE EXCEPTION 'tenant_archived' USING ERRCODE = 'P0003';
    END IF;

    -- Set the session variable. All RLS policies read this.
    PERFORM set_config('app.current_tenant_id', p_tenant_id::text, true);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- SECURITY DEFINER: the function runs as its owner (platform role), not the
-- calling application role. This ensures it can always read the tenants table
-- even if the application role has restricted SELECT on tenants.


-- ── OrgNodes ───────────────────────────────────────────────────────────────
CREATE TABLE org_nodes (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid        NOT NULL REFERENCES tenants(id),
    parent_id   uuid        REFERENCES org_nodes(id),
    type        text        NOT NULL
                            CHECK (type IN ('COMPANY','DIVISION','DEPARTMENT','BRANCH','COST_CENTRE')),
    code        text        NOT NULL,
    name        text        NOT NULL,
    -- Materialised path. Format: "code1/code2/code3"
    -- Maintained by the materialisePath BeforeHook.
    path        text        NOT NULL DEFAULT '',
    is_active   boolean     NOT NULL DEFAULT true,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),

    UNIQUE (tenant_id, code)  -- code unique per tenant, not globally
);

ALTER TABLE org_nodes ENABLE ROW LEVEL SECURITY;

-- Standard RLS policy — tenant_id must match the session context.
CREATE POLICY org_nodes_tenant_isolation ON org_nodes
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

-- Index for subtree queries via materialised path.
-- LIKE 'hq/nairobi/%' uses this index (text_pattern_ops required for LIKE).
CREATE INDEX org_nodes_path_idx ON org_nodes (tenant_id, path text_pattern_ops);

CREATE TRIGGER set_updated_at
    BEFORE UPDATE ON org_nodes
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

```sql
-- internal/platform/tenant/migrations/20240101000001_create_tenants.down.sql
DROP TABLE IF EXISTS org_nodes;
DROP FUNCTION IF EXISTS set_tenant_context(uuid);
DROP TABLE IF EXISTS tenants;
```

---

## 42.10 Integration With Every Other Module

Every table that holds tenant-specific data follows the same pattern established here:

```sql
-- Example: a Finance module table
CREATE TABLE invoices (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid        NOT NULL REFERENCES tenants(id),  -- ← always present
    org_unit_id uuid        REFERENCES org_nodes(id),         -- ← for unit-scoped entities
    -- ... other columns
);

ALTER TABLE invoices ENABLE ROW LEVEL SECURITY;
CREATE POLICY invoices_tenant_isolation ON invoices
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);
```

The `tenant_id` column and RLS policy are not optional. They are the contract that makes multi-tenancy work. Any table that stores per-tenant data without RLS is a security vulnerability.

### The Request Lifecycle

Here is the exact flow for every authenticated API request:

```
1. Client sends HTTP request with Bearer token
2. Auth middleware extracts tenant_id from JWT claims
3. Middleware calls: SELECT set_tenant_context($tenant_id)
   - PostgreSQL checks tenant status (ACTIVE/SUSPENDED/ARCHIVED)
   - Sets app.current_tenant_id for this connection
4. All subsequent queries on this connection are filtered by RLS
5. Request handler runs — it does NOT need to manually filter by tenant
6. Response returned
7. Connection returned to pool (app.current_tenant_id resets)
```

Step 3 is where suspended tenants are blocked. The stored procedure raises an exception if the tenant is not ACTIVE. The middleware catches this and returns HTTP 402 Payment Required (or 403 for archived).

---

## 42.11 Frequently Asked Questions

**"Can two tenants share a customer record?"**

No. Customer records have `tenant_id`. RLS makes it structurally impossible for Tenant A to read Tenant B's records — even if a bug in Go code skipped the tenant filter.

**"What if the developer forgets to call `set_tenant_context`?"**

`current_setting('app.current_tenant_id')` returns an empty string. The RLS policy evaluates `''::uuid` which is a type error — PostgreSQL raises an exception. Zero rows returned and no data leaked. In development, the error is visible immediately.

**"Can a user belong to two tenants?"**

Yes. They have two separate User rows — one in each tenant. Same email address, but separate passwords, roles, and data. There is no shared identity across tenants.

**"Why not just create a separate schema per tenant (schema-based isolation)?"**

Schema-per-tenant requires creating a new schema on tenant signup (DDL inside an HTTP request — dangerous and slow), running migrations separately for each schema (500 tenants × 50 migrations = 25,000 migration steps), and managing connection pooling across schemas. RLS gives equal isolation with none of these costs.

**"Is RLS really safe? What if a superuser connects?"**

A PostgreSQL superuser bypasses RLS by default. The application role used in production is NOT a superuser — it is a dedicated `application_role` with minimal privileges. Superuser access requires direct server access and is separately audited. Application bugs cannot escalate to superuser.

**"What does `SECURITY DEFINER` on `set_tenant_context` mean?"**

The function runs as its *owner* (the migration user, typically `postgres`), not as the *caller* (the application role). This means even a restricted application role can call `set_tenant_context` — it does not need SELECT on the `tenants` table itself. The function enforces the status check and then sets the session variable. Safe by construction.
