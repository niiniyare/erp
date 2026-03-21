[<-- Back to Index](README.md)

## Resource Hierarchy & Entity Model

### The Entity Tree

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

---

### Access Rules

```
Root entity     → access everything (EntityScope{Type: "all"})
Mid-level node  → access own entity + all descendants (EntityScope{Type: "subtree"})
Leaf node       → access own entity only (EntityScope{Type: "entity"})
```

Resolved at login, stored in `sessions.entity_scope`. Applied by repositories:

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

---

### Ownership Fields

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

### EntityScope Type

```go
type EntityScope struct {
    Type       string    // "all" | "subtree" | "entity"
    EntityID   uuid.UUID
    PathPrefix string    // used for LIKE queries on entity_path
}
```

Users assigned to the company root entity get `Type: "all"` — they see everything. Users assigned to a branch get `Type: "subtree"` with `PathPrefix` set to the branch's materialized path, restricting them to that branch and its children. Users at a leaf get `Type: "entity"`, restricting to exactly one node.

---

### Practical Examples

**Scenario: Regional Manager (Mombasa Branch Only)**
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

**Scenario: Nairobi Region Manager (all Nairobi branches)**
```go
platform.IAM.InviteUser(ctx, domain.UserInviteParams{
    EntityID: nairobiRegionEntityID,  // mid-level → access own + descendants
})
// EntityScope{Type: "subtree", PathPrefix: "/company/nairobi/"}
// → sees all Nairobi branches via LIKE '/company/nairobi/%'
```

**Scenario: Finance Controller (Full Company Access)**
```go
platform.IAM.InviteUser(ctx, domain.UserInviteParams{
    EntityID: companyRootEntityID,  // root → access everything
})
// EntityScope{Type: "all"} → no entity filter in queries
```

---

### Entity Security Guarantee

`entity_scope` is loaded from the authenticated session — **never from request params**. The DB WHERE clause uses the session's path prefix. A user cannot craft a request to see data outside their entity scope.

---

Next: [UI Navigation](./11b-ui-navigation.md)
