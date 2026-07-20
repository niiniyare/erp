> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: "Platform Module — Tenant & Organisation"
part: "Part VI — Platform Entities"
chapter: 42
section: "platform-tenant-module"
related:
  - "[Chapter 41: Platform Modules Overview](./platform-module-overview.md)"
  - "[Chapter 43: IAM Module](./platform-iam-module.md)"
  - "[Chapter 2: EntityDefinition](../part-01-foundations/02-entity-definition.md)"
  - "[Chapter 11: Database Migrations](../part-02-entity-system/11-migrations.md)"
---

# Chapter 42 — Tenant & Organisation Module

> **Who should read this?** Developers building multi-tenant features, anyone debugging tenant isolation, and stakeholders who need to understand how "one customer" maps to data in Awo.

---

## 42.1 The Mental Model: Apartments in a Building

Think of Awo as an apartment building.

- The **building** is one running Awo server — one PostgreSQL cluster, one set of application processes.
- Each **apartment** is a tenant — a business that has subscribed to Awo. Tenant A cannot see Tenant B's data, ever.
- Each apartment has **rooms** — org units (branches, departments, cost centres). Employees work in specific rooms and can only see records from their room and the rooms below theirs.

The Tenant module manages the building registry (which apartments exist) and the floor plan (which rooms each apartment has). Every other module in the system hangs off this foundation.

---

## 42.2 Why Two Separate Entities?

The module exposes two EntityDefinitions:

| Entity | Purpose |
|---|---|
| **Tenant** | One business customer. Billing, status, configuration root. |
| **OrgNode** | One node in that business's internal hierarchy (branch, department, team). |

They are separate because they have different lifecycles. A tenant is created once and rarely changes. Org nodes are created, rearranged, and disabled constantly as the business reorganises. Separating them means org tree operations never touch the tenant row, and tenant billing logic never needs to know about branch structures.

---

## 42.3 EntityDefinition: Tenant

```go
// internal/platform/tenant/definition.go
package tenant

import "awo.so/framework/definition"

var TenantDef = definition.EntityDefinition{
    Name:        "tenant",
    Label:       "Tenant",
    Description: "A business customer with isolated data and configuration.",
    Module:      "Platform",
    Table:       "tenants",
    OrgScope:    org.ScopeLevelGlobal, // tenants are cross-tenant by definition

    Fields: []*definition.FieldDef{
        definition.Field("name").
            OfType(definition.FieldTypeData).
            WithLabel("Business Name").
            RequiredField().
            SearchableField(),

        definition.Field("slug").
            OfType(definition.FieldTypeData).
            WithLabel("URL Slug").
            RequiredField().
            UniqueField().
            ImmutableField(), // set on creation, never changed — used in URLs

        definition.Field("status").
            OfType(definition.FieldTypeSelect).
            WithLabel("Status").
            WithOptions("PENDING", "ACTIVE", "SUSPENDED", "ARCHIVED").
            WithDefault("PENDING").
            RequiredField(),

        definition.Field("plan").
            OfType(definition.FieldTypeData).
            WithLabel("Subscription Plan"),

        definition.Field("country_code").
            OfType(definition.FieldTypeData).
            WithLabel("Country").
            WithMaxLen(2).
            RequiredField(),

        definition.Field("currency_code").
            OfType(definition.FieldTypeData).
            WithLabel("Default Currency").
            WithMaxLen(3).
            RequiredField(),

        definition.Field("timezone").
            OfType(definition.FieldTypeData).
            WithLabel("Timezone").
            WithDefault("Africa/Nairobi"),

        definition.Field("locale").
            OfType(definition.FieldTypeData).
            WithLabel("Locale").
            WithDefault("en-KE"),

        definition.Field("db_schema").
            OfType(definition.FieldTypeData).
            WithLabel("DB Schema").
            ImmutableField().   // assigned at provisioning, never changed
            SensitiveField(),   // internal infra detail — excluded from API responses
    },

    Hooks: []definition.HookDef{
        {Name: "validateStatusTransition", Before: validateTenantStatusTransition},
        {Name: "provisionSchema",          After:  provisionTenantSchema},
    },

    Policies: []definition.PolicyDef{
        {Op: definition.OpCreate, Fn: requirePlatformAdmin},
        {Op: definition.OpRead,   Fn: allowSelfTenant},
        {Op: definition.OpUpdate, Fn: requirePlatformAdmin},
        {Op: definition.OpDelete, Fn: definition.DenyAll},
    },

    SoftDelete: false, // tenants are ARCHIVED, never deleted
    Audited:    true,
}
```

### Why `OrgScope: ScopeLevelGlobal`?

Tenants are the isolation boundary, not a thing inside one. The `tenants` table has no `tenant_id` column — it is a global catalogue. Only platform administrators can see and modify it. Individual tenants can read their own row via the `allowSelfTenant` policy.

### Why `slug` is `ImmutableField`?

The slug appears in API URLs (`/t/{slug}/...`), email templates, and audit log entries. Changing it would break every external link and cached reference. If a business rebrands, they get a new tenant. The old one is archived.

### Why `db_schema` is `SensitiveField`?

This is the PostgreSQL schema name for this tenant's data. Exposing it in API responses would leak infrastructure topology to end users. With `Sensitive: true`, the framework excludes it from all `GET` responses automatically.

---

## 42.4 Tenant Status Machine

Tenants move through a strict set of states. The `validateTenantStatusTransition` hook enforces the allowed transitions:

```
PENDING ──► ACTIVE ──► SUSPENDED ──► ACTIVE
                └──────────────────► ARCHIVED
PENDING ──────────────────────────► ARCHIVED
```

| From | To | Meaning |
|---|---|---|
| PENDING | ACTIVE | Provisioning complete; tenant can log in |
| PENDING | ARCHIVED | Trial expired before activation |
| ACTIVE | SUSPENDED | Non-payment or policy violation |
| SUSPENDED | ACTIVE | Payment received; access restored |
| ACTIVE | ARCHIVED | Voluntary churn or forced termination |
| SUSPENDED | ARCHIVED | Extended non-payment |

**ARCHIVED is terminal.** No transition out of ARCHIVED exists. Data is retained for the legal retention period and then purged by a scheduled job.

```go
// internal/platform/tenant/hooks.go
func validateTenantStatusTransition(ctx context.Context, mut *definition.Mutation) error {
    if mut.Op != definition.OpUpdate {
        return nil
    }
    from := mut.Before.Get("status").(string)
    to   := mut.After.Get("status").(string)
    if from == to {
        return nil
    }

    allowed := map[string][]string{
        "PENDING":   {"ACTIVE", "ARCHIVED"},
        "ACTIVE":    {"SUSPENDED", "ARCHIVED"},
        "SUSPENDED": {"ACTIVE", "ARCHIVED"},
        "ARCHIVED":  {},
    }
    for _, s := range allowed[from] {
        if s == to {
            return nil
        }
    }
    return fmt.Errorf("tenant: invalid status transition %s → %s", from, to)
}
```

---

## 42.5 EntityDefinition: OrgNode

```go
var OrgNodeDef = definition.EntityDefinition{
    Name:        "org_node",
    Label:       "Organisation Unit",
    Description: "A node in the tenant's organisational hierarchy.",
    Module:      "Platform",
    Table:       "org_nodes",
    OrgScope:    org.ScopeLevelTenant, // each node belongs to one tenant

    Fields: []*definition.FieldDef{
        definition.Field("name").
            OfType(definition.FieldTypeData).
            WithLabel("Unit Name").
            RequiredField().
            SearchableField(),

        definition.Field("code").
            OfType(definition.FieldTypeData).
            WithLabel("Unit Code").
            UniqueField(),

        definition.Field("parent_id").
            OfType(definition.FieldTypeLink).
            WithLabel("Parent Unit").
            LinksTo("org_node"), // self-referential

        definition.Field("path").
            OfType(definition.FieldTypeData).
            WithLabel("Materialised Path").
            ImmutableField(). // maintained by hook, never set directly
            HiddenInList(),   // implementation detail; shown in detail view only

        definition.Field("depth").
            OfType(definition.FieldTypeInt).
            WithLabel("Depth").
            ImmutableField().
            HiddenInList(),

        definition.Field("active").
            OfType(definition.FieldTypeBool).
            WithLabel("Active").
            WithDefault(true),

        definition.Field("type").
            OfType(definition.FieldTypeSelect).
            WithLabel("Unit Type").
            WithOptions("company", "branch", "department", "team", "project"),
    },

    Hooks: []definition.HookDef{
        {Name: "materialisePath", Before: materialisePath},
    },

    Policies: []definition.PolicyDef{
        {Op: definition.OpCreate, Fn: requireRole("tenant_admin")},
        {Op: definition.OpRead,   Fn: allowWithinTenant},
        {Op: definition.OpUpdate, Fn: requireRole("tenant_admin")},
        {Op: definition.OpDelete, Fn: requireRole("tenant_admin")},
    },

    SoftDelete: false,
    Audited:    true,
}
```

### Why `path` and `depth` are `ImmutableField`?

These columns are computed by the `materialisePath` hook before every create and update. If a user could set them directly, the tree would become inconsistent. `ImmutableField` tells the framework to reject changes to these columns from UPDATE requests (`mutableFieldNames` filters them out of UPDATE SQL).

### Why `HiddenInList` on `path`?

`path` is an implementation detail like `1.2.7.` — not meaningful to end users in a table view. `Hidden: true` tells the SDUI layer to exclude this column from auto-generated list views. The API still returns it (useful for programmatic tree traversal).

---

## 42.6 The Materialised Path Hook

The `materialisePath` hook maintains a `path` column like `1.4.12.` — dot-delimited ancestor IDs from root to node. This pattern makes subtree queries a single ILIKE:

```sql
-- All nodes under node 4:
SELECT * FROM org_nodes WHERE path LIKE '1.4.%';

-- All ancestors of node 12:
SELECT * FROM org_nodes WHERE '1.4.12.' LIKE path || '%';
```

No recursive CTE needed. No adjacency-list walk. One index scan.

```go
// internal/platform/tenant/hooks.go
func materialisePath(ctx context.Context, mut *definition.Mutation) error {
    rec := mut.After
    parentID, _ := rec.Get("parent_id").(uuid.UUID)

    if parentID == uuid.Nil {
        // Root node
        rec.Set("path", fmt.Sprintf("%s.", rec.ID().String()))
        rec.Set("depth", 0)
        return nil
    }

    // Look up parent's path in the same transaction.
    // svc is injected via closure at registration time.
    parent, err := svc.FindOrgNode(ctx, mut.TenantID, parentID)
    if err != nil {
        return fmt.Errorf("materialisePath: parent not found: %w", err)
    }
    parentPath := parent.Get("path").(string)
    rec.Set("path",  parentPath + rec.ID().String() + ".")
    rec.Set("depth", strings.Count(parentPath, "."))
    return nil
}
```

**Reparenting** (moving a subtree to a new parent) requires a dedicated service operation — not a plain `PUT` update. The framework's `sqlbuilder` hardcodes `org_unit_id` and by convention `parent_id` changes trigger path recalculation for the entire subtree via a batch update.

---

## 42.7 Row-Level Security: `set_tenant_context`

Every query runs through PostgreSQL Row-Level Security. The framework calls a stored procedure to bind the current tenant before every query:

```sql
-- migrations/001_create_set_tenant_context.up.sql

CREATE OR REPLACE FUNCTION set_tenant_context(p_tenant_id uuid)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER  -- runs as the DB owner, not the app role
AS $$
BEGIN
    PERFORM set_config('app.current_tenant_id', p_tenant_id::text, true);
    -- true = local to current transaction
END;
$$;
```

`SECURITY DEFINER` is required because the app role does not have `SET` privileges on `app.*` GUC variables — only the DB owner does. The procedure is a thin wrapper that does nothing except call `set_config`.

Every tenant-scoped table has a RLS policy:

```sql
ALTER TABLE finance_accounts ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON finance_accounts
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);
```

The framework (via `pgstore.TenantStoreAdapter`) calls `SELECT set_tenant_context($1)` at the start of every transaction via `WithTx`, and at the start of every non-transactional `ForEntity` call. There is no path through the persistence layer that skips this.

---

## 42.8 Migration

```sql
-- internal/platform/tenant/migrations/20240101000001_create_tenants.up.sql

CREATE TABLE tenants (
    id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at   timestamptz NOT NULL DEFAULT NOW(),
    updated_at   timestamptz NOT NULL DEFAULT NOW(),
    name         text NOT NULL,
    slug         text NOT NULL UNIQUE,
    status       text NOT NULL DEFAULT 'PENDING'
                     CHECK (status IN ('PENDING','ACTIVE','SUSPENDED','ARCHIVED')),
    plan         text,
    country_code char(2) NOT NULL,
    currency_code char(3) NOT NULL,
    timezone     text NOT NULL DEFAULT 'Africa/Nairobi',
    locale       text NOT NULL DEFAULT 'en-KE',
    db_schema    text NOT NULL UNIQUE
);

-- internal/platform/tenant/migrations/20240101000002_create_org_nodes.up.sql

CREATE TABLE org_nodes (
    id         uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id  uuid NOT NULL REFERENCES tenants(id),
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    name       text NOT NULL,
    code       text,
    parent_id  uuid REFERENCES org_nodes(id),
    path       text NOT NULL,
    depth      int  NOT NULL DEFAULT 0,
    active     boolean NOT NULL DEFAULT true,
    type       text CHECK (type IN ('company','branch','department','team','project')),
    UNIQUE (tenant_id, code)
);

CREATE INDEX org_nodes_tenant_idx ON org_nodes (tenant_id);
CREATE INDEX org_nodes_path_idx   ON org_nodes (path text_pattern_ops);
-- text_pattern_ops enables ILIKE 'prefix%' to use the index

ALTER TABLE org_nodes ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON org_nodes
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);
```

---

## 42.9 The `org.Tree` Interface

The framework defines an `org.Tree` interface used by `api.Handler` for list-scoping:

```go
// framework/org/tree.go
type Tree interface {
    // Descendants returns IDs of node and all its descendants (inclusive).
    // Used to build the OrgUnitIDs filter for ScopeLevelUnit entities.
    Descendants(ctx context.Context, tenantID, nodeID uuid.UUID) ([]uuid.UUID, error)
}
```

The Tenant module provides the concrete implementation using the materialised path:

```sql
SELECT id FROM org_nodes
WHERE tenant_id = $1
  AND path LIKE (SELECT path FROM org_nodes WHERE id = $2 AND tenant_id = $1) || '%';
```

When a `ScopeLevelUnit` entity is listed and the viewer has `OrgUnitID != nil`, the handler calls `Descendants` and passes the result as `ListOptions.OrgUnitIDs`. The SQL builder converts this to `WHERE org_unit_id = ANY($n)`.

---

## 42.10 FAQ

**Q: Can a user belong to multiple org units?**
A: A user has one primary `OrgUnitID` in their viewer context (from their session). Cross-unit access is handled by assigning the user to a parent org node that covers both units. Fine-grained cross-unit access can also be granted via Casbin roles (see Chapter 43 IAM).

**Q: What happens to org_nodes when a tenant is archived?**
A: Org nodes are tenant-scoped. When a tenant is archived, RLS blocks all access immediately. Data retention jobs run separately on an archival schedule defined by the deployment.

**Q: Can the tenant table use RLS like everything else?**
A: No. The `tenants` table is `ScopeLevelGlobal` — it has no `tenant_id` column, so standard RLS based on `app.current_tenant_id` does not apply. Access control for tenants is entirely through the `Policies` chain (platform admin check).

**Q: Why not just use PostgreSQL schemas per tenant (schema-per-tenant isolation)?**
A: Awo uses shared-schema + RLS (row-level security) rather than schema-per-tenant. Schema-per-tenant requires dynamic SQL everywhere and makes cross-tenant analytics impossible. RLS gives the same isolation with standard queries and allows reporting views across tenants when needed by platform administrators.
