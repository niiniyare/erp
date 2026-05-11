[<-- Back to Index](README.md)

## User Entity Scope

> **Implementation status**: Entity scope types are [IMPLEMENTED] in `domain/session.go`.
> Application of entity scope in repository queries is an [APPLICATION-LAYER CONCERN] — service methods must implement the WHERE clause logic. The scope type is stored and available in every session.

---

### What Is Entity Scope?

Within an organisation (tenant), data is often partitioned by **entities** — branches, departments, subsidiaries, cost centres, warehouses, or any other node in the organisational hierarchy. A branch manager should see data for their branch and its sub-branches. A cashier should only see their own till's transactions. The CFO should see everything.

Entity scope answers the question: **"Within this tenant, how broadly can this user see entity-partitioned data?"**

Entity scope is **additive with Casbin RBAC**:
- Casbin decides **what actions** the user can perform (create invoice, approve transaction, etc.)
- Entity scope decides **which entities' data** falls within the user's visible range

Both must allow access for a request to succeed. A user with `finance:transaction read` permission but `EntityScopeEntity` at branch X cannot read transactions from branch Y — even though their Casbin policy technically permits `finance:transaction read`.

Entity scope is an **application-layer concern** (Layer 2). It is enforced by service methods adding WHERE clauses or ltree path predicates before issuing queries. PostgreSQL RLS (Layer 1) handles the tenant boundary independently and at a lower level. See `internal/core/iam/domain/authz.go` for the isolation model comment.

---

### 2. Scope Types Explained

The three scope types are defined in `internal/core/iam/domain/session.go`:

#### `EntityScopeAll` (`"all"`)

The user can see data across all entities within the tenant. No entity filter is applied.

```go
// In a repository query:
case domain.EntityScopeAll:
    // no extra WHERE clause; RLS already enforces tenant
```

Appropriate for: CFOs, global HR directors, tenant admins, platform operators (cross-tenant context), consolidation report users.

#### `EntityScopeSubtree` (`"subtree"`)

The user can see their home entity and all of its descendants in the entity ltree hierarchy.

```go
case domain.EntityScopeSubtree:
    q = q.Where("entity_path <@ ?", sess.EntityScope.PathPrefix)
```

The `PathPrefix` is the ltree path of the user's home entity (e.g., `/company/region-north/branch-a/`). The `<@` operator matches all nodes at or below that path.

Appropriate for: branch managers, regional supervisors, area managers, department heads, regional HR managers.

#### `EntityScopeEntity` (`"entity"`)

The user can only see data belonging to their exact, directly assigned entity.

```go
case domain.EntityScopeEntity:
    q = q.Where("entity_id = ?", sess.EntityScope.EntityID)
```

Appropriate for: cashiers, front-line staff, station attendants, individual contributors with no management scope.

---

### 3. Practical Examples

#### Multi-Branch Retail

```
Tenant: RetailCo
Entity hierarchy:
  RetailCo HQ
  ├── Region North
  │   ├── Branch A (city centre)
  │   │   ├── POS Terminal 1
  │   │   └── POS Terminal 2
  │   └── Branch B (suburb)
  └── Region South
      └── Branch C (mall)
```

| User | Role | Entity | Scope | Can See |
|---|---|---|---|---|
| CFO | finance_director | HQ | `all` | All branches, all transactions |
| Region North Manager | branch_manager | Region North | `subtree` | Branches A, B and their terminals |
| Branch A Manager | branch_manager | Branch A | `subtree` | Branch A + POS 1 + POS 2 |
| Cashier at POS 1 | cashier | POS Terminal 1 | `entity` | Only POS Terminal 1 data |
| Internal Auditor | auditor (read-only role) | HQ | `all` | All data, read-only |

#### Finance / Accounting

| User | Role | Scope | Can See |
|---|---|---|---|
| Group CFO | finance_director | `all` | All cost centres, all GL |
| Cost Centre Owner | cost_centre_manager | `subtree` | Own cost centre + sub-centres |
| External Auditor | auditor (read-only) | `all` | All GL entries (read-only Casbin policy) |
| AP Clerk | accounts_payable | `entity` | Own department's payables only |

The key insight: the external auditor has `EntityScopeAll` but their Casbin role grants only `read` actions — they cannot create, update, or approve anything regardless of scope.

#### HR Management

| User | Role | Scope | Description |
|---|---|---|---|
| Global HR Director | hr_admin | `all` | Manages all employees across all locations |
| Regional HR Manager | hr_manager | `subtree` from region | Manages employees in their region only |
| Line Manager | hr_user | `subtree` from department | Views/manages team members in their department |
| Employee (self-service) | employee_portal | `entity` | Sees only their own records |

#### Airline / Travel Operations

| User | Role | Entity | Scope | Description |
|---|---|---|---|---|
| Station Manager (JFK) | station_manager | JFK Airport | `subtree` | All JFK gates, lounges, ground ops |
| Gate Agent | gate_agent | Gate B12 | `entity` | Only Gate B12 operations |
| Operations Director | ops_director | All Stations | `all` | Network-wide operations |
| Regional Manager (Americas) | region_manager | Americas region | `subtree` | All US/Canada/LATAM stations |

#### Forecourt / Gas Station

| User | Role | Entity | Scope | Description |
|---|---|---|---|---|
| Area Manager | area_manager | District 5 | `subtree` | All sites in District 5 |
| Site Manager | site_manager | Site 42 | `subtree` | Site 42 + all pumps/nozzles |
| Forecourt Attendant | attendant | Pump 3 | `entity` | Only Pump 3 data |

The site manager uses `subtree` (not `entity`) because the site contains child entities (pumps, nozzles, POS terminals) and the manager needs visibility into all of them.

#### Portal Customer

A customer accessing a supplier portal:
- `EntityScopeEntity` scoped to their own account/contact record
- Their Casbin role has read access to `portal:invoice/*` and `portal:delivery/*`
- They cannot see any other customer's data (entity filter) nor perform write operations (Casbin role)

---

### 4. How Enforcement Works

#### Storage

Entity scope is stored as JSONB in `user_sessions.entity_scope`:
```json
{
  "type": "subtree",
  "entity_id": "a1b2c3d4-...",
  "path_prefix": "/company/region-north/branch-a/"
}
```

The session service resolves the user's entity membership and highest-privilege role at login time. This is computed once and embedded in the `ResolvedSession` — no additional DB queries per request.

#### Enforcement in Service Methods

Service methods (not Casbin) are responsible for applying entity scope. The pattern from `domain/session.go`:

```go
switch sess.EntityScope.Type {
case domain.EntityScopeAll:
    // no extra WHERE clause; RLS already enforces tenant
case domain.EntityScopeSubtree:
    q = q.Where("entity_path <@ ?", sess.EntityScope.PathPrefix)
case domain.EntityScopeEntity:
    q = q.Where("entity_id = ?", sess.EntityScope.EntityID)
}
```

This WHERE clause is added **before** the DB query. PostgreSQL's ltree `<@` operator is efficient for subtree queries when an index is present on `entity_path`.

#### The Two-Layer Check

For a request to succeed, both layers must permit it:

1. **Casbin (what)**: Does `role:branch_manager` have `finance:transaction read` permission in this tenant domain? → Yes/No
2. **Entity scope (where)**: Does the requested transaction's `entity_id` fall within the user's scope? → Yes/No

Only if both return Yes does the handler return data.

This is not the same as "two separate authorization checks." Casbin checks the permission type; entity scope filters the data set. They are complementary, not redundant.

---

### 5. Assignment and Management

#### Assigning Entity Scope

Entity scope is determined at role assignment time. When `AssignRole` is called, the entity context is encoded in the `role_assignments` metadata. The session service reads this when building the `EntityScope` for a new login.

For the `user_roles` table (v2.0 reserved): the `entity_id` column on `user_roles` (migration `000407`) stores the entity context for each role assignment.

#### Changing Scope

To change a user's entity scope:
1. Revoke the existing role assignment (`RevokeRole`)
2. Create a new assignment with the updated entity context (`AssignRole`)

There is no in-place update of entity scope on an existing assignment. This is intentional — it creates a clean audit trail in `role_assignments`.

#### Default Scope

If no entity is specified when assigning a role, the default depends on the user's role type:
- Platform users: `EntityScopeAll` (they operate cross-tenant)
- Tenant admins: `EntityScopeAll` (they manage the whole tenant)
- Regular users: `EntityScopeEntity` (conservative default — most restrictive)

API key sessions default to `EntityScopeAll` (see `buildAPIKeySession` in `service/apikey.go`). This is a current simplification — entity scope for API keys is not configurable in v1.0.

---

See also:
- [Tenant Administration](./23-tenant-administration.md)
- [Session Pre-Computation](./10b-session-precomputation.md)
- [API Keys and Service Accounts](./26-api-keys-and-service-accounts.md)
