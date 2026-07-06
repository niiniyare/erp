# Organization Model

## The Two Isolation Layers

Awo has two independent isolation layers. They must never be conflated.

| Layer | Mechanism | Enforced By | Scope |
|---|---|---|---|
| **Tenant isolation** | PostgreSQL RLS | Framework | Cross-tenant data never leaks |
| **Organization scope** | Application code | OrganizationService | Within a tenant, who sees which org |

**Tenant isolation** is a hard security boundary enforced at the database level via `current_tenant_id()`. The framework guarantees `WHERE tenant_id = current_tenant_id()` on every RLS-enabled table and nothing more.

**Organization scope** is an authorization and visibility model evaluated entirely in Go. `OrganizationService.ResolveScope()` converts a `ViewerContext` into a `[]uuid.UUID` of visible org IDs. Application services pass that set as an explicit `IN` predicate to their repositories. Organization tables carry no RLS policies.

---

## Tenant vs Organization

| Concept | Entity | Responsibility |
|---|---|---|
| **Tenant** | `platform_tenant` | Multi-tenancy, billing, licensing, module installation, feature flags, RLS boundary |
| **Organization** | `platform_organization` | Business hierarchy, team structure, scope-based data visibility |

A **tenant** is created by Awo during provisioning. A **tenant** has exactly one `platform_tenant` record and zero or more `platform_organization` nodes. One tenant — one account. Tenant identity is immutable.

An **organization** is created by the tenant admin. It models the internal structure: how the business organizes people, data, and operations. Organization nodes are fully tenant-managed.

---

## Organization Tree

Organizations form an arbitrary-depth tree within a tenant. The framework imposes no fixed shape.

```
Holding Company  (type=holding, depth=0)
├── Kenya        (type=country, depth=1)
│   ├── Nairobi  (type=branch,  depth=2)
│   │   ├── Retail Team   (type=team, depth=3)
│   │   └── Credit Team   (type=team, depth=3)
│   ├── Kisumu   (type=branch,  depth=2)
│   └── Eldoret  (type=branch,  depth=2)
├── Tanzania     (type=country, depth=1)
└── Uganda       (type=country, depth=1)
```

### Node Types

Types are **metadata-driven**. Tenant admins register valid types via `platform_org_type`. The framework never validates against a predefined list. Common registrations: `holding`, `company`, `subsidiary`, `division`, `region`, `territory`, `branch`, `department`, `team`, `store`, `warehouse`, `cost_centre`.

The type field on `platform_organization` is a denormalized varchar — no FK constraint to `platform_org_type`. This allows type names to be deactivated without cascading breaks to existing nodes.

### Materialized Path

The `path` column stores the full ancestor chain: `/root-id/parent-id/self-id/`. This enables O(1) ancestor and descendant queries:

```sql
-- All descendants of Nairobi:
SELECT * FROM platform_organization
WHERE path LIKE '/holding-id/kenya-id/nairobi-id/%'
  AND tenant_id = $1;  -- explicit tenant filter; no RLS on this table

-- Ancestor IDs from path string (no recursive CTE needed):
SELECT id FROM platform_organization
WHERE id = ANY(
    string_to_array(trim(both '/' from '/holding-id/kenya-id/nairobi-id/'), '/')::uuid[]
)
AND tenant_id = $1;
```

---

## User Organization Assignment

Users may belong to multiple organizations simultaneously. Assignment is modeled in `platform_org_assignment`.

```
User: Jane (tenant.admin + Finance Manager)
├── Assignment: Holding Company  — role=manager, is_primary=true
├── Assignment: Kenya            — role=member
└── Assignment: Tanzania         — role=viewer
```

### Assignment Fields

| Field | Notes |
|---|---|
| `user_id` | FK to `iam_user` |
| `organization_id` | FK to `platform_organization` |
| `role` | Org-level role (not an IAM role): `manager`, `member`, `viewer`, … |
| `is_primary` | Home org — at most one per user per tenant (partial unique index) |

### IAM Roles vs Organization Roles

These are **distinct** and must not be conflated:

| Type | Table | Purpose |
|---|---|---|
| IAM role | `iam_user_role` | Platform-wide capabilities: `tenant.admin`, `finance.viewer`, … |
| Org role | `platform_org_assignment.role` | Position within a specific org node: manager, member, viewer |

A user may be `tenant.admin` (IAM) and `viewer` in their assigned org (org role). Both apply independently.

---

## ViewerContext

`ViewerContext` is populated by auth middleware after session validation and carries the full identity context of the acting user:

```go
type ViewerContext struct {
    TenantID              uuid.UUID
    UserID                uuid.UUID
    ActiveOrganizationID  uuid.UUID   // current session org (may differ from primary)
    PrimaryOrganizationID uuid.UUID   // home org (is_primary = true)
    OrganizationAssignments []OrgMembership
    VisibilityMode        VisibilityMode
    ExplicitOrganizationIDs []uuid.UUID // VisibilityExplicit only
    ScopeResolver         ScopeResolver // VisibilityCustom only
    Roles                 []string
    IsPlatformAdmin       bool
    IsTenantAdmin         bool
}
```

### Active Organization

Users with multiple org assignments may switch their **active organization** during a session. The active org governs which data context they operate in. If not explicitly set, `EffectiveOrganizationID()` falls back to the primary org.

---

## Visibility Modes

`VisibilityMode` is set by auth middleware based on the viewer's roles and active organization.

| Mode | Who Sees What | Typical User |
|---|---|---|
| `VisibilityCurrent` | Own org only | Branch cashier |
| `VisibilityDescendants` | Own org + all subtree below | Regional manager |
| `VisibilityAncestors` | Own org + all nodes above | Escalation / breadcrumb |
| `VisibilityEntireTenant` | All orgs in tenant | HQ Finance Manager, tenant.admin |
| `VisibilityExplicit` | Fixed UUID set | Internal auditor (non-contiguous) |
| `VisibilityCustom` | Delegated to ScopeResolver | Matrix orgs, project teams |

### Scope Resolution

```
OrganizationService.ResolveScope(ctx, viewer) → []uuid.UUID
```

`nil` return means `VisibilityEntireTenant` — caller omits the org filter entirely.

Application services apply the result as an explicit IN predicate:

```go
// Application service — example
func (s *InvoiceService) List(ctx context.Context, opts ListOptions) ([]*Invoice, error) {
    viewer := organization.MustViewerFromContext(ctx)
    orgIDs, err := s.orgSvc.ResolveScope(ctx, viewer)
    if err != nil {
        return nil, err
    }
    f := buildFilter(opts)
    if len(orgIDs) > 0 {
        f = filter.And(f, filter.In("org_id", toAny(orgIDs)...))
    }
    // orgIDs == nil → no org filter, all orgs visible
    records, _, err := s.repo.Query(ctx, f)
    return records, err
}
```

---

## Repository Contract

Repositories are **org-unaware**. They receive filter criteria and execute queries. They never compute org visibility.

```
HTTP Request
  ↓ auth middleware (session validation + org assignment load)
ViewerContext populated
  ↓ application service
OrganizationService.ResolveScope() → []uuid.UUID
  ↓ application service builds filter
Repository.List(ctx, filter)
  ↓ SQL: WHERE tenant_id = $1 AND org_id = ANY($2)
    (tenant_id enforced by explicit param; org_id enforced by IN predicate)
```

The `tenant_id` in the repository query is passed as an explicit parameter (matching `TenantContext.TenantID`), not via RLS — org tables have no RLS policies.

---

## Security Properties

| Property | How Enforced |
|---|---|
| Cross-tenant data leak | RLS on `platform_tenant` and all business entity tables |
| Cross-org data leak | Explicit org IN predicate from `ResolveScope` in application services |
| Unauthorized org switch | Session carries org assignments; switch validated against `OrganizationAssignments` |
| Org admin scope creep | `VisibilityDescendants` — cannot see parent/sibling orgs |
| HQ full access | `VisibilityEntireTenant` — requires `tenant.admin` or explicit grant |
| Auditor non-contiguous access | `VisibilityExplicit` — fixed UUID list, reviewed and granted explicitly |

---

## API Reference

### OrganizationService

```go
// Tree operations
Create(ctx, CreateInput) (*OrganizationDTO, error)
GetByID(ctx, id) (*OrganizationDTO, error)
GetByCode(ctx, tenantID, code) (*OrganizationDTO, error)
Move(ctx, id, newParentID) error              // recomputes path+depth for subtree
Disable(ctx, id) error                        // cascades to descendants
Enable(ctx, id) error
Tree(ctx, tenantID) ([]*OrganizationDTO, error)
Ancestors(ctx, id) ([]*OrganizationDTO, error)
Descendants(ctx, id, maxDepth) ([]*OrganizationDTO, error)

// Scope resolution
ResolveScope(ctx, ViewerContext) ([]uuid.UUID, error)

// Type registry
RegisterType(ctx, tenantID, OrgTypeInput) (*OrgTypeDTO, error)
ListTypes(ctx, tenantID) ([]*OrgTypeDTO, error)
DeactivateType(ctx, id) error

// Assignments
Assign(ctx, AssignInput) (*AssignmentDTO, error)
Unassign(ctx, tenantID, userID, orgID) error
SetPrimary(ctx, tenantID, userID, orgID) error
ListAssignments(ctx, tenantID, userID) ([]*AssignmentDTO, error)
LoadViewerMemberships(ctx, tenantID, userID) ([]OrgMembership, error)
```

---

## Migration Notes

Prior to v1.1, organization hierarchy lived in `platform/tenant`:

- `platform_org_unit` → renamed to `platform_organization` (type preserved)
- `platform_branch` → `platform_organization` rows with `type = 'branch'`

The old migration files remain for historical correctness. A separate data migration script transforms existing rows. Schema migration `20260101000080` creates `platform_organization`, `platform_org_type`, and `platform_org_assignment`.
