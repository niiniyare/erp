---
title: "Chapter 9: Privacy Policies — Row-Level Security"
part: "Part II — The EntityDefinition System"
chapter: 9
section: "09-privacy-policies"
related:
  - "[Chapter 8: The Persistence Interface](08-persistence-interface.md)"
  - "[Chapter 16: RBAC](../part-03-api/16-rbac.md)"
  - "[Chapter 14: Multi-Tenancy Middleware](../part-03-api/14-multitenancy-middleware.md)"
---

# Chapter 9: Privacy Policies — Row-Level Security

RBAC answers "can this user perform this operation on this entity type?" Privacy policies answer "which rows can this user see or modify within that entity type?" The two systems are complementary and independently necessary. Conflating them leads to either over-permissioned queries (users see data they should not) or an unmaintainable mess of WHERE clauses scattered throughout business logic.

---

## 9.1. Why Privacy Policies Are Separate From RBAC

### 9.1.1. RBAC Controls Operations; Privacy Policies Control Rows

RBAC is a coarse-grained gate: a `finance_manager` role can `read` Invoices, but a `sales_rep` role cannot. This is enforced before any query runs.

Privacy policies are a fine-grained filter: even among users who can `read` Invoices, a sales rep in the Nairobi branch should only see Nairobi branch invoices. A collection agent should see only their assigned accounts. A tenant's data should never be visible in another tenant's query results.

These two concerns require different enforcement mechanisms:
- RBAC is checked once per request against a static policy table (Casbin)
- Privacy policies are injected as WHERE predicates into every database query dynamically

### 9.1.2. Privacy Policies Control What Rows a Query Can Return and Modify

A user with `finance_manager` role and `read` permission on `Invoice` calls `GET /api/v1/invoices`. Without privacy policies, they get all invoices across all tenants, all branches, all customers. With privacy policies:
- Tenant isolation scopes the result to their tenant's schema
- Branch scoping limits results to their assigned branch
- Ownership policy may further limit results to invoices they created or are assigned to

The combined effect is that the same SQL query infrastructure produces radically different result sets depending on the caller's context.

### 9.1.3. Why Application-Level WHERE Clauses Are Insufficient

The naive approach is to add a `WHERE tenant_id = ?` clause in every repository method. This fails because:

1. **It must be remembered everywhere**: every developer writing a new query must remember to add the clause. One missed clause = data leak.
2. **It must be maintained across schema changes**: if you add a new query method, you must add the clause again.
3. **It cannot be composed**: combining tenant isolation, ownership filtering, and role-based row filtering requires increasingly complex WHERE conditions inline.
4. **It does not apply to eager-loaded edges**: `invoice.Customer` lazy-loading bypasses repository methods entirely.

Privacy policies are declarative and applied automatically by the repository implementation — you declare the policy once, and it is enforced everywhere.

### 9.1.4. How Privacy Policies Are Enforced at the `EntityRepository` Interface Layer

Every `EntityRepository` implementation reads the entity's `Policy()` declaration and injects the resulting predicates into every read query. This happens transparently:

```
repo.Query(ctx, filter.Eq("status", "submitted"))
        │
        ▼ (inside the repository)
policy := entity.Policy()
policyPredicate := policy.QueryRule(ctx)   // e.g. Eq("tenant_id", tenantFromCtx)
finalFilter := filter.And(userFilter, policyPredicate)
// → SQL: WHERE status = 'submitted' AND tenant_schema IS current_schema()
```

The caller's filter and the policy predicate are combined with AND. The caller cannot bypass the policy — they have no access to the final SQL.

---

## 9.2. Policy Types

### 9.2.1. Query Rules — Applied to All SELECT Operations

A query rule adds a predicate to every SELECT: `Get`, `Query`, `Exists`, `Count`, `Aggregate`. It operates as an additional WHERE clause.

```go
func (Invoice) Policy() ent.Policy {
    return privacy.Policy{
        // Tenant isolation — always applied
        privacy.QueryRuleFunc(func(ctx context.Context, q *ent.InvoiceQuery) error {
            tid := tenant.IDFromContext(ctx)
            if tid == uuid.Nil {
                return privacy.Deny  // no tenant context = deny all reads
            }
            // search_path already set to tenant_{id} by middleware
            // but explicit tenant_id filter adds defence in depth
            return privacy.Allow  // search_path handles isolation
        }),
    }
}
```

### 9.2.2. Mutation Rules — Applied to CREATE, UPDATE, DELETE

A mutation rule intercepts all writes. It can inspect the incoming record, the actor's context, and existing data to allow, deny, or skip the operation.

```go
privacy.MutationRuleFunc(func(ctx context.Context, m ent.Mutation) error {
    // Only the record owner or a manager can update
    actorID := actor.IDFromContext(ctx)
    if m.Op() == ent.OpUpdate {
        ownerID, _ := m.(*ent.InvoiceMutation).OwnerID()
        if ownerID != actorID && !actor.HasRole(ctx, "finance_manager") {
            return privacy.Deny
        }
    }
    return privacy.Skip  // let subsequent rules decide
})
```

`privacy.Skip` means "I don't have an opinion; let the next rule decide." `privacy.Allow` means "permit this, stop processing rules." `privacy.Deny` means "reject this, stop processing rules."

### 9.2.3. Field Visibility Rules — Masking or Excluding Fields from Results

Field visibility rules apply post-query, masking or removing fields before the entity is returned to the caller. This is how `Sensitive` fields work:

```go
privacy.FieldVisibilityRuleFunc(func(ctx context.Context, r *ent.Invoice) error {
    if !actor.HasPermission(ctx, "invoice:view_sensitive") {
        r.TaxPin = ""             // mask the field
        r.BankAccountNumber = ""  // mask the field
    }
    return nil
})
```

Field visibility rules run after the DB query, before the response is serialised. They are less efficient than query rules (data is fetched then discarded) but necessary for fields that cannot be hidden at the SQL level (e.g. derived fields, JSONB paths).

---

## 9.3. Built-in Policy Primitives

### 9.3.1. `TenantIsolation` — Every Query Scoped to the Resolved Tenant

`TenantIsolation` is the most fundamental policy. It is applied by default to every entity. It denies all reads and writes when there is no resolved tenant in the context, and restricts rows to the current tenant's PostgreSQL schema.

```go
func (Invoice) Policy() ent.Policy {
    return privacy.Policy{
        awo.TenantIsolation(),  // always include this first
        // ... additional policies
    }
}
```

Under schema-per-tenant, `TenantIsolation` works by verifying the `search_path` is set correctly. It does not add a `tenant_id` WHERE clause (the schema itself provides isolation). For entities in the shared platform schema, `TenantIsolation` adds an explicit `tenant_id = ?` predicate.

### 9.3.2. `OwnerOnly` — User Can Only See Their Own Records

```go
awo.OwnerOnly(field: "created_by_id")
// Generates: WHERE created_by_id = current_user_id
```

`OwnerOnly` can be combined with role overrides to let managers see all records:

```go
awo.OwnerOnlyUnless(field: "created_by_id", roles: []string{"finance_manager", "admin"})
// Finance managers see all; others see only their own
```

### 9.3.3. `RoleFilter` — Additional Filter Predicate Applied for a Given Role

Applies a specific WHERE predicate only when the actor holds a specified role:

```go
awo.RoleFilter("branch_manager",
    func(ctx context.Context) filter.Filter {
        branchID := actor.BranchIDFromContext(ctx)
        return filter.Eq("branch_id", branchID)
    },
)
// Branch managers only see invoices for their branch
```

Multiple `RoleFilter` instances can be combined. If the actor holds `branch_manager` AND `collection_agent` roles, both predicates are applied with AND.

### 9.3.4. `DepartmentScope` — Records Visible Within the User's Department Subtree

For organisations with hierarchical departments, `DepartmentScope` restricts a user to records belonging to their department and any sub-departments:

```go
awo.DepartmentScope(field: "department_id")
// Generates: WHERE department_id IN (
//     SELECT id FROM departments WHERE path LIKE '/user-dept-path/%'
// )
```

This uses the materialised path pattern from Chapter 6 for efficient tree traversal.

---

## 9.4. Writing Custom Privacy Policies

### 9.4.1. The `Policy` Interface

```go
// QueryRule intercepts SELECT operations
type QueryRule interface {
    EvalQuery(ctx context.Context, q ent.Query) error
}

// MutationRule intercepts CREATE, UPDATE, DELETE operations
type MutationRule interface {
    EvalMutation(ctx context.Context, m ent.Mutation) error
}

// Policy is a collection of rules evaluated in order
type Policy interface {
    EvalQuery(ctx context.Context, q ent.Query) error
    EvalMutation(ctx context.Context, m ent.Mutation) error
}
```

Custom policies implement one or both interfaces and are included in the entity's `Policy()` declaration.

### 9.4.2. Accessing Tenant and User Context Inside a Policy

```go
type CollectionAgentPolicy struct{}

func (CollectionAgentPolicy) EvalQuery(ctx context.Context, q ent.Query) error {
    // This policy only applies to collection agents
    if !actor.HasRole(ctx, "collection_agent") {
        return privacy.Skip
    }

    agentID := actor.IDFromContext(ctx)
    if agentID == uuid.Nil {
        return privacy.Deny
    }

    // Restrict to invoices assigned to this agent
    q.(*ent.InvoiceQuery).Where(
        invoice.AssignedAgentID(agentID),
    )
    return privacy.Allow
}
```

Always check `privacy.Skip` conditions first. A policy that evaluates to `Allow` for every caller (because it forgets to check the role) disables all subsequent more-restrictive policies.

### 9.4.3. Returning Additional Filter Predicates

For complex multi-condition policies, build the predicate incrementally:

```go
func (p RegionalManagerPolicy) EvalQuery(ctx context.Context, q ent.Query) error {
    if !actor.HasRole(ctx, "regional_manager") {
        return privacy.Skip
    }

    a := actor.FromContext(ctx)
    regionIDs := a.ManagedRegionIDs()

    if len(regionIDs) == 0 {
        return privacy.Deny  // manager with no regions = no access
    }

    // Get all branches in their regions
    branchIDs, err := fetchBranchIDsForRegions(ctx, regionIDs)
    if err != nil {
        return err
    }

    q.(*ent.InvoiceQuery).Where(
        invoice.BranchIDIn(branchIDs...),
    )
    return privacy.Allow
}
```

### 9.4.4. Returning Field Masks

```go
type SensitiveFieldPolicy struct{}

func (SensitiveFieldPolicy) EvalQuery(ctx context.Context, q ent.Query) error {
    // No predicate modification needed — runs post-query as a field mask
    return privacy.Skip
}

// Implement post-load field masking via the Interceptor pattern
func (SensitiveFieldPolicy) ApplyFieldMask(ctx context.Context, r *ent.Invoice) {
    if !actor.HasPermission(ctx, "invoice:view_bank_details") {
        r.BankAccountNumber = "****" + r.BankAccountNumber[len(r.BankAccountNumber)-4:]
    }
    if !actor.HasPermission(ctx, "invoice:view_tax_details") {
        r.KraPin = "[REDACTED]"
    }
}
```

---

## 9.5. Composing Policies

### 9.5.1. `privacy.And` — All Policies Must Pass

```go
func (Invoice) Policy() ent.Policy {
    return privacy.Policy{
        // Both must pass — AND semantics by position
        awo.TenantIsolation(),
        CollectionAgentPolicy{},
    }
}
```

Rules in a `privacy.Policy` slice are evaluated in order with the following semantics:
- First `Allow` → short-circuits, allows
- First `Deny` → short-circuits, denies
- `Skip` → continues to the next rule
- All rules return `Skip` → deny by default

### 9.5.2. `privacy.Or` — At Least One Policy Must Pass

```go
func (Invoice) Policy() ent.Policy {
    return privacy.Policy{
        awo.TenantIsolation(),
        privacy.Or(
            awo.OwnerOnly("created_by_id"),
            awo.RoleFilter("finance_manager", nil),  // nil = no additional filter
        ),
    }
}
// Access if: tenant matches AND (owner OR finance_manager)
```

### 9.5.3. `privacy.Not` — Inversion

```go
// Allow access to anyone EXCEPT collection agents
// (collection agents have a separate, more restricted policy)
privacy.MutationRuleFunc(func(ctx context.Context, m ent.Mutation) error {
    if actor.HasRole(ctx, "collection_agent") {
        return privacy.Deny
    }
    return privacy.Skip
})
```

### 9.5.4. Execution Order and Short-Circuit Behaviour

```
Rule 1 → Skip  →  Rule 2 → Skip  →  Rule 3 → Allow  →  ALLOW
Rule 1 → Skip  →  Rule 2 → Deny              →  DENY
Rule 1 → Allow                               →  ALLOW (skip rules 2, 3)
Rule 1 → Skip  →  Rule 2 → Skip  →  (end of rules)  →  DENY (default)
```

Design policies so that `TenantIsolation` is always first and returns `Skip` (not `Allow`) — allowing subsequent rules to further restrict the result set.

---

## 9.6. Testing Privacy Policies

### 9.6.1. Unit Testing a Policy With a Mock Context

```go
func TestCollectionAgentPolicy(t *testing.T) {
    agentID := uuid.New()
    otherAgentID := uuid.New()

    tests := []struct {
        name         string
        ctx          context.Context
        assignedTo   uuid.UUID
        expectPolicy string
    }{
        {
            name:         "agent sees own invoice",
            ctx:          actortest.WithCollectionAgent(agentID),
            assignedTo:   agentID,
            expectPolicy: "allow",
        },
        {
            name:         "agent cannot see another agent's invoice",
            ctx:          actortest.WithCollectionAgent(agentID),
            assignedTo:   otherAgentID,
            expectPolicy: "deny",
        },
        {
            name:         "finance manager sees all",
            ctx:          actortest.WithRole("finance_manager"),
            assignedTo:   otherAgentID,
            expectPolicy: "skip",  // skip = pass to next rule
        },
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            q := &mockInvoiceQuery{assignedAgentID: tc.assignedTo}
            policy := CollectionAgentPolicy{}
            err := policy.EvalQuery(tc.ctx, q)
            assert.Equal(t, tc.expectPolicy, policyResultName(err))
        })
    }
}
```

### 9.6.2. Integration Testing — Verifying Rows Are Filtered Correctly End-to-End

```go
func TestInvoicePrivacyPolicies(t *testing.T) {
    client := enttest.Open(t, "sqlite3", ":memory:?_fk=1")

    tenantA := uuid.New()
    tenantB := uuid.New()
    agentA := uuid.New()
    agentB := uuid.New()

    // Create invoices for both agents
    ctxA := tenanttest.WithTenant(agenttest.WithCollectionAgent(t, agentA), tenantA)
    ctxB := tenanttest.WithTenant(agenttest.WithCollectionAgent(t, agentB), tenantA)

    invA, _ := repo.Create(ctxA, invoice.Create{
        AssignedAgentID: agentA,
        TotalAmount:     decimal.NewFromInt(5000),
    })
    invB, _ := repo.Create(ctxB, invoice.Create{
        AssignedAgentID: agentB,
        TotalAmount:     decimal.NewFromInt(3000),
    })

    // Agent A can only see their own invoice
    resultA, _, _ := repo.Query(ctxA, filter.None())
    assert.Len(t, resultA, 1)
    assert.Equal(t, invA.ID, resultA[0].ID)

    // Agent B can only see their own invoice
    resultB, _, _ := repo.Query(ctxB, filter.None())
    assert.Len(t, resultB, 1)
    assert.Equal(t, invB.ID, resultB[0].ID)

    // Finance manager sees both
    ctxManager := tenanttest.WithTenant(agenttest.WithRole(t, "finance_manager"), tenantA)
    resultMgr, _, _ := repo.Query(ctxManager, filter.None())
    assert.Len(t, resultMgr, 2)
}
```

### 9.6.3. Common Mistakes — Policies That Silently Pass Everything

The most dangerous privacy policy bug is a policy that returns `privacy.Allow` without adding any WHERE predicates. This happens when:

```go
// BUG: returns Allow without restricting anything
func (p TenantPolicy) EvalQuery(ctx context.Context, q ent.Query) error {
    if tenant.IDFromContext(ctx) != uuid.Nil {
        return privacy.Allow  // WRONG: no WHERE added, returns ALL records
    }
    return privacy.Deny
}
```

The correct pattern:
```go
func (p TenantPolicy) EvalQuery(ctx context.Context, q ent.Query) error {
    tid := tenant.IDFromContext(ctx)
    if tid == uuid.Nil {
        return privacy.Deny
    }
    // search_path handles isolation; return Skip to allow other rules to run
    return privacy.Skip
}
```

**Always return `Skip` from a "pass-through" check**. Reserve `Allow` for policies that have added a restricting predicate and want to prevent less-restrictive subsequent rules from broadening the result set.

Write an explicit test that:
1. Creates records for two different tenants
2. Queries as tenant A
3. Asserts that tenant B's records are not in the result
4. Queries as tenant B
5. Asserts that tenant A's records are not in the result

Run this test on every entity that has cross-tenant data risk. It should be in your CI pipeline.

---

## Chapter Summary

Chapter 9 establishes why RBAC (operation permission) and privacy policies (row and field visibility) are separate concerns (§9.1), explains why application-level WHERE clauses are architecturally insufficient (§9.1.3), and documents the full policy enforcement mechanism at the `EntityRepository` interface layer (§9.1.4).

The three most critical concepts:

- **Fail closed** (§9.6 / common mistakes): a policy that cannot extract user or tenant context from `ctx` must return `ErrDeny`, never `nil, nil`. Silent allow-all is the worst possible bug in a privacy policy.
- **`Get` returns `ErrNotFound` for policy-restricted records** — not a permission error. This prevents information leakage about what records exist.
- **`Or` compositions produce OR WHERE clauses** — test with `EXPLAIN ANALYZE` on large tables. A nil-returning branch (admin full access) in an `Or` correctly produces no filter; verify every nil branch is an intentional "allow all" decision.

**Next chapters to read:**

- [§16 — RBAC](../part-03-api/16-rbac.md) — the role-based access control layer that runs before privacy policies; both systems must be understood together
- [§10 — Custom Fields](10-custom-fields.md) — custom entity fields stored as JSONB use path predicates in policy filter returns; understanding custom field filters requires reading both chapters
- [§3 — Architecture Overview](../part-01-foundations/03-architecture-overview.md) — §3.4.3 documents the RLS enforcement model (shared schema, `set_tenant_context()`) that the `TenantIsolation` policy builds on
