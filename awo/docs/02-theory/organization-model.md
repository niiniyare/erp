# Organization Model

## Overview

Awo separates two concepts that are often conflated in ERP systems:

| Concept | Package | Entity | Scope |
|---|---|---|---|
| **Tenant** | `platform/tenant` | `platform_tenant` | Infrastructure isolation — one per customer account |
| **Organization** | `platform/organization` | `platform_organization` | Business hierarchy — zero or more per tenant |

A **tenant** is an infrastructure boundary. It controls data isolation, billing, and lifecycle (PENDING → ACTIVE → SUSPENDED → ARCHIVED). Every record in every table belongs to exactly one tenant, enforced by PostgreSQL Row-Level Security via `current_tenant_id()`.

An **organization** is a business node. It models the internal structure a tenant uses to organize people, data, and operations: divisions, departments, branches, teams, regions, cost centres, or any other concept the business uses.

---

## Organization Tree

Organization nodes form an arbitrary-depth tree within a tenant. The tree shape is unconstrained — tenant admins define valid node types via settings.

```
ACME Ltd (type=company, depth=0)
├── East Africa Region (type=region, depth=1)
│   ├── Nairobi Branch (type=branch, depth=2)
│   │   ├── Retail Team (type=team, depth=3)
│   │   └── Credit Team (type=team, depth=3)
│   └── Mombasa Branch (type=branch, depth=2)
└── West Africa Region (type=region, depth=1)
    └── Lagos Branch (type=branch, depth=2)
```

### Key Fields

| Field | Type | Notes |
|---|---|---|
| `id` | UUID | Primary key |
| `tenant_id` | UUID FK | Infrastructure boundary — immutable |
| `parent_id` | UUID FK (self-ref) | `NULL` for root nodes |
| `name` | varchar(255) | Display name |
| `code` | varchar(50) | Immutable — embedded in workflow IDs and audit records |
| `type` | varchar(50) | Metadata-driven; not a fixed enum |
| `path` | varchar(4096) | Materialized path: `/root-id/parent-id/self-id/` |
| `depth` | int | 0 for root, 1 for direct children, etc. |
| `active` | bool | Disable cascades to descendants via service layer |

### Materialized Path

The `path` column stores the full ancestor chain as a slash-delimited string of UUIDs: `/root-id/child-id/leaf-id/`. This enables O(1) ancestor and descendant queries without recursive CTEs:

```sql
-- All descendants of node X:
SELECT * FROM platform_organization
WHERE path LIKE '/x-uuid/%'
  AND tenant_id = current_tenant_id();

-- All ancestors of node X (if X.path = '/a/b/x/'):
SELECT * FROM platform_organization
WHERE id = ANY(string_to_array(trim(both '/' from '/a/b/x/'), '/'))
  AND tenant_id = current_tenant_id();
```

The `idx_platform_organization_path` index uses `text_pattern_ops` to make `LIKE '/prefix/%'` queries index-scannable.

### Node Types

Types are **metadata-driven** — the type field is a free-form `varchar(50)`, not a SQL `CHECK` constraint. Tenant admins register valid types via platform settings. The framework does not enforce a fixed list. Common values:

- `company` — top-level legal entity
- `division` — major business division
- `department` — functional department
- `region` — geographic region
- `branch` — physical location
- `team` — operational team
- `cost_centre` — financial rollup node

---

## Visibility Model (OrganizationScope)

`OrganizationScope` controls which org nodes are visible to a query. It is **orthogonal** to `TenantContext`: TenantContext prevents cross-tenant data leaks; OrganizationScope restricts data within a tenant's org tree.

| Scope | Description | Use Case |
|---|---|---|
| `ScopeCurrent` | Caller's own node only | Self-service data entry |
| `ScopeDescendants` | Caller's node + all below | Branch manager view |
| `ScopeAncestors` | Caller's node + all above | Escalation / breadcrumb |
| `ScopeExplicit` | Caller-supplied UUID set | Ad-hoc cross-team access |
| `ScopeEntireTenant` | All nodes in tenant | Tenant admin view |
| `ScopeCustom` | Delegates to `OrganizationResolver` | Matrix orgs, project teams |

Scope is propagated via `context.Context`:

```go
// Middleware sets scope based on the authenticated user's org membership.
ctx = organization.WithScope(ctx, organization.ScopeContext{
    Scope:          organization.ScopeDescendants,
    OrganizationID: user.OrgID,
})

// Service reads it.
sc := organization.ScopeFromContext(ctx)
```

When no scope context is present, `ScopeFromContext` defaults to `ScopeEntireTenant` — callers without an explicit scope see all org nodes, subject to RBAC.

---

## Permission Model

Organization management has two distinct permission domains:

| Permission | Who Holds It | What They Can Do |
|---|---|---|
| `role:platform-admin` | Awo platform operators | Create/modify orgs across any tenant |
| `role:tenant.admin` | Customer's admin user | Create/modify/delete orgs within their tenant |
| `role:tenant.user` | Standard users | Read org nodes (to resolve their own scope) |

Tenant admins cannot see or modify another tenant's org tree — enforced by RLS (`tenant_id = current_tenant_id()`).

---

## API

### Tree Operations (OrganizationService)

```go
svc := organization.NewOrganizationService()

// Create root node.
root, err := svc.Create(ctx, organization.CreateInput{
    TenantID: tenantID,
    Name:     "ACME Ltd",
    Code:     "ACME",
    Type:     "company",
})

// Create child.
nairobi, err := svc.Create(ctx, organization.CreateInput{
    TenantID: tenantID,
    Name:     "Nairobi Branch",
    Code:     "NBI",
    Type:     "branch",
    ParentID: root.ID,
})

// List descendants of root.
nodes, err := svc.Descendants(ctx, root.ID, 0) // depth=0 = unlimited

// Move a node.
err = svc.Move(ctx, nairobi.ID, newParentID)

// Get ancestors (breadcrumb).
breadcrumb, err := svc.Ancestors(ctx, nairobi.ID)
```

### Cycle Prevention

`PathComputeHook.BeforeCreate` and `OrganizationService.Move` both check that the target parent's path does not already contain the node being moved. If it does, the operation is rejected with an error:

```
organization.PathComputeHook: cycle detected — self ID already present in parent path
```

---

## Relationship to Other Platform Modules

| Module | Relationship |
|---|---|
| `platform/tenant` | Organization has FK to `platform_tenant.id`. Tenant is read-only from organization's perspective. |
| `platform/iam` | Users carry an `org_id` FK to `platform_organization`. IAM middleware resolves the user's scope. |
| `platform/audit` | Audit log records include `org_id` for organizational context. |
| `platform/settings` | Org-scoped settings override tenant-scoped settings (scope hierarchy: system → tenant → org → user). |

---

## Migration Notes

Prior to v1.1, organization hierarchy lived in `platform_tenant` as two entities:

- `platform_org_unit` — renamed to `platform_organization`
- `platform_branch` — becomes `platform_organization` rows with `type = 'branch'`

A data migration script (separate from schema migrations) transforms existing rows. The schema migration (`20260101000080`) creates `platform_organization`. The data migration is out-of-band and must be reviewed before execution.
