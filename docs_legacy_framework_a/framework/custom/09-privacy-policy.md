> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

### Chapter 9 — Privacy Policies — Row-Level Security

Privacy policies are the mechanism by which Awo enforces row-level and field-level access control at the persistence layer. They are distinct from RBAC: a user's role determines whether they may perform an operation at all; privacy policies determine which records they may see or modify when they do. A `sales_rep` role may have `read` permission on `SalesOrder` records, but a privacy policy restricts the visible records to only those orders assigned to that representative. Both checks must pass. This chapter documents the policy type system, the built-in policy primitives, the composition operators, and the testing patterns.

---

#### 9.1. Why Privacy Policies Are Separate From RBAC

##### 9.1.1. RBAC controls what operations a role can perform

Role-based access control answers the question: is this user allowed to perform this operation on this entity type? A user with the `cashier` role may `read` and `create` `ShiftTransaction` records but may not `delete` them. This check is binary: the role either has the permission or does not. It is evaluated before any record is loaded, using only the user's role and the entity name — no database query is needed.

RBAC is implemented as a middleware gate (§16.3.1) that runs before the route handler. If the check fails, a 403 response is returned immediately and no `EntityRepository` method is called. Full RBAC documentation is in §16.

##### 9.1.2. Privacy policies control what rows a query can return and modify

Privacy policies answer the question: given that this user is allowed to read `SalesOrder` records, which specific records may they see? A `regional_manager` may see all orders for their region. A `sales_rep` may see only orders they are assigned to. A `finance_reviewer` may see all submitted orders but not draft ones. These distinctions cannot be expressed as RBAC role permissions; they require runtime evaluation against record attributes.

Privacy policies inject additional WHERE clause predicates into every `Query` call and mutation target validation check on the `EntityRepository`. They run at the interface layer, not in the route handler, so they apply regardless of whether the call comes from an API route handler, a hook, a Temporal activity, or an internal framework component.

##### 9.1.3. Why application-level WHERE clauses are insufficient

Adding `WHERE assigned_to = ?` in the route handler is insufficient for three reasons. First, it only protects the specific route where it is added; a hook, a report query, or an activity that directly calls `repo.Query` without the WHERE clause creates a bypass. Second, it requires every developer who writes a query to remember to add the restriction, creating a fragile, manually-maintained security boundary. Third, it is invisible to code reviewers who do not know which entities have access restrictions — there is no authoritative declaration of the restriction to audit.

Privacy policies declared on the `EntityDefinition` are applied unconditionally at the interface layer. There is no path through which a query can skip them, because the `EntityRepository` implementation applies them before executing any SQL.

##### 9.1.4. How privacy policies are enforced at the `EntityRepository` interface layer

When a `Query`, `Exists`, `Count`, `Aggregate`, `Get`, `Update`, or `Delete` call is made, the `EntityRepository` implementation calls the privacy policy evaluation engine with the current user context and tenant context. The engine evaluates the policies registered on the entity's `EntityDefinition` and collects any additional `Filter` predicates they return. These predicates are merged with the caller's `Filter` argument using AND logic before the SQL query is constructed. The caller cannot observe or override this merge.

For `Get` by ID, the policy predicate is applied as an additional WHERE clause. If the record exists but the policy predicate filters it out, `Get` returns `entity.ErrNotFound` — not a permission error. This is intentional: returning a distinct "you can see this record exists but cannot read it" error leaks information. The caller cannot distinguish between "record does not exist" and "record exists but is restricted".

---

#### 9.2. Policy Types

##### 9.2.1. Query rules — applied to all SELECT operations

Query rules inject additional filter predicates into all SELECT operations: `Get`, `Query`, `Exists`, `Count`, and `Aggregate`. They are the most common policy type and the primary mechanism for implementing per-row visibility restrictions.

A query rule returns either a `Filter` to be ANDed into the query, or a deny decision that causes the query to return an empty result set (for `Query`, `Exists`, `Count`) or `entity.ErrNotFound` (for `Get`). A query rule that returns an empty filter (no additional predicate) allows all records to be seen.

##### 9.2.2. Mutation rules — applied to CREATE, UPDATE, DELETE

Mutation rules are evaluated before any `Create`, `Update`, `Delete`, `BulkCreate`, or `BulkUpdate` call executes. For `Create`, the rule receives the proposed record data and can allow or deny the creation. For `Update` and `Delete`, the rule is evaluated against the current record (loaded via a pre-flight `Get`), and can allow or deny the mutation.

A mutation rule that denies a `Create` returns `entity.ErrPermissionDenied`. A mutation rule that denies an `Update` or `Delete` on a specific record returns `entity.ErrNotFound` (for the same information-leakage reason as query rules — the caller cannot distinguish between "record does not exist" and "record exists but you cannot mutate it").

##### 9.2.3. Field visibility rules — masking or excluding fields from results

Field visibility rules operate on the `EntityRecord` values returned by `Get` and `Query` after the SQL query executes. They can: mask a field value (replace it with a redacted string or zero value), remove a field entirely from the `EntityRecord`, or limit a multi-value field to a subset of its values. Use field visibility rules for salary fields visible only to HR roles, for tax PIN numbers visible only to finance roles, and for any `Sensitive` field (§5.2.4) that requires role-based visibility beyond the blanket `Sensitive` exclusion.

---

#### 9.3. Built-in Policy Primitives

##### 9.3.1. `TenantIsolation` — every query scoped to the resolved tenant

`privacy.TenantIsolation()` is the baseline policy that must be applied to every entity. It does not inject a WHERE clause predicate — tenant isolation is already enforced structurally by the schema-per-tenant model (§3.4.3). Instead, it verifies at policy evaluation time that the tenant context is present in the request context and that it matches the tenant context under which the repository was resolved. If a tenant context mismatch is detected (which should be impossible in correctly written code but is checked as a defence-in-depth measure), the policy returns a deny decision.

Every `EntityDefinition` in the framework has `TenantIsolation` applied by default via the framework's default policy composition. Module developers do not need to declare it explicitly; attempting to declare it explicitly alongside other policies is harmless — the framework deduplicates it.

##### 9.3.2. `OwnerOnly` — user can only see their own records

`privacy.OwnerOnly(fieldName)` restricts query and mutation access to records where the named field equals the current user's ID. It is typically applied to entities that are created by a user and should remain private to that user unless a higher-privilege role overrides.

```go
// Example: OwnerOnly policy on a personal entity
entity.Policy(
    privacy.And(
        privacy.TenantIsolation(),
        privacy.Or(
            privacy.RoleFilter("manager", nil),
            privacy.OwnerOnly("created_by"),
        ),
    ),
)
```

`OwnerOnly` injects the predicate `Eq(fieldName, currentUserID)` into query and mutation rules. The field named by `fieldName` must be declared as a `Link` to `User` or as a `UUID` field on the entity; the framework validates this at startup.

##### 9.3.3. `RoleFilter` — additional filter predicate applied for a given role

`privacy.RoleFilter(roleName, filter)` applies an additional filter predicate when the current user has the named role. The filter argument can be `nil` to allow the role unrestricted access, or a `*entity.Filter` to further restrict what the role can see.

```go
// Example: RoleFilter granting unrestricted access to managers, restricted to own for others
privacy.RoleFilter("regional_manager", nil) // managers see all
privacy.RoleFilter("sales_rep",
    entity.NewFilter().EqField("assigned_to", "$current_user_id"))
```

The `$current_user_id` sentinel is interpolated at evaluation time with the current user's ID. `RoleFilter` is typically composed with `Or` to express "managers see all, reps see their own":

```go
// Example: Composing RoleFilter with Or for tiered access
entity.Policy(
    privacy.And(
        privacy.TenantIsolation(),
        privacy.Or(
            privacy.RoleFilter("sales_manager", nil),
            privacy.RoleFilter("sales_rep",
                entity.NewFilter().Eq("assigned_rep_id", "$current_user_id")),
        ),
    ),
)
```

##### 9.3.4. `DepartmentScope` — records visible within the user's department subtree

`privacy.DepartmentScope(departmentFieldName)` restricts visibility to records where the named department field is within the current user's department or any of its sub-departments (the subtree in the department hierarchy). This policy is appropriate for entities like `Employee`, `LeaveRequest`, and `PerformanceReview` where managers should see records for everyone in their subtree but not for peers' departments.

`DepartmentScope` uses the materialised path pattern from §6.4.2: it loads the current user's department record, extracts its `path` field, and injects a `StartsWith(departmentFieldName+".path", userDeptPath)` predicate. This is a single indexed lookup regardless of hierarchy depth.

```go
// Example: DepartmentScope policy on LeaveRequest
entity.Policy(
    privacy.And(
        privacy.TenantIsolation(),
        privacy.Or(
            privacy.RoleFilter("hr_manager", nil),
            privacy.DepartmentScope("department_id"),
        ),
    ),
)
```

---

#### 9.4. Writing Custom Privacy Policies

##### 9.4.1. The `Policy` interface

A custom policy implements the `privacy.Policy` interface:

```go
// Example: The privacy.Policy interface
package privacy

type Policy interface {
    // EvalQuery is called for all SELECT operations.
    // Return a Filter to add additional restrictions, or ErrDeny to block entirely.
    EvalQuery(ctx context.Context) (*entity.Filter, error)

    // EvalMutation is called for all write operations.
    // Return nil to allow, or ErrDeny to block.
    EvalMutation(ctx context.Context, op MutationOp, record entity.EntityRecord) error
}
```

A policy that only needs to restrict reads can implement `EvalQuery` and return `nil` from `EvalMutation`. A policy that only needs to restrict writes can implement `EvalMutation` and return an empty filter from `EvalQuery`.

##### 9.4.2. Accessing tenant and user context inside a policy

The `ctx context.Context` argument to both `EvalQuery` and `EvalMutation` carries the full tenant and user context:

```go
// Example: Accessing tenant and user context inside a custom policy
func (p *RegionPolicy) EvalQuery(ctx context.Context) (*entity.Filter, error) {
    tc, err := tenant.FromContext(ctx)
    if err != nil {
        return nil, privacy.ErrDeny
    }
    uc, err := auth.UserFromContext(ctx)
    if err != nil {
        return nil, privacy.ErrDeny
    }
    _ = tc
    if uc.HasRole("national_manager") {
        return nil, nil // no additional filter; see all records
    }
    regionCode := uc.Attribute("region_code")
    if regionCode == "" {
        return nil, privacy.ErrDeny // no region assigned, deny all
    }
    return entity.NewFilter().Eq("region_code", regionCode), nil
}
```

Policies must not perform slow operations. They are evaluated on every query; a policy that makes a database call on every invocation multiplies latency. For policies that require a database lookup (fetching the user's department ID, for example), use a context value: the first evaluation fetches and caches the department ID in the context; subsequent evaluations read from the context value.

##### 9.4.3. Returning additional filter predicates

`EvalQuery` returns a `*entity.Filter` that is ANDed with the caller's filter before the SQL query executes. A nil return means "no additional restriction" (allow everything, subject to other policies in the composition). Returning `privacy.ErrDeny` causes the query to return an empty result set immediately without executing any SQL.

The returned filter is subject to the same type and field validation as any other filter: the field names must exist on the entity, and the value types must match the field declarations. An invalid filter predicate returned from a policy causes the query to fail with `entity.ErrInvalidFilter`, which surfaces as a 500 error — a policy that returns invalid filters is a bug, not expected runtime behaviour.

##### 9.4.4. Returning field masks

Field visibility rules are implemented by registering a `FieldMask` function alongside the `Policy`:

```go
// Example: Policy with field masking for sensitive salary data
type SalaryPolicy struct{}

func (p *SalaryPolicy) EvalQuery(ctx context.Context) (*entity.Filter, error) {
    return nil, nil // no row-level restriction from this policy
}

func (p *SalaryPolicy) EvalMutation(ctx context.Context, op privacy.MutationOp, record entity.EntityRecord) error {
    return nil // no mutation restriction from this policy
}

func (p *SalaryPolicy) MaskFields(ctx context.Context, record entity.EntityRecord) entity.FieldMask {
    uc, err := auth.UserFromContext(ctx)
    if err != nil {
        // Deny field access on context error — fail closed
        return entity.MaskAll("basic_salary", "gross_salary", "net_salary")
    }
    if uc.HasRole("hr_manager") || uc.HasRole("payroll_admin") {
        return entity.MaskNone() // full visibility
    }
    if uc.UserID() == record.Get("employee_id") {
        return entity.MaskNone() // employees see their own salary
    }
    return entity.MaskAll("basic_salary", "gross_salary", "net_salary")
}
```

`entity.MaskAll(fields...)` returns a mask that replaces all named field values with `nil` in the returned `EntityRecord`. `entity.MaskNone()` returns a mask that passes all fields through unchanged.

---

#### 9.5. Composing Policies

Policy composition combines multiple policies into a single effective policy using boolean logic. The composition operators evaluate each constituent policy and combine their results.

##### 9.5.1. `privacy.And` — all policies must pass

`privacy.And(policies...)` creates a policy that passes only when all constituent policies pass. For `EvalQuery`, the filter predicates from all passing constituent policies are ANDed together. If any constituent policy returns `ErrDeny`, the composition denies immediately.

```go
// Example: And composition requiring both tenant isolation and ownership
entity.Policy(
    privacy.And(
        privacy.TenantIsolation(),
        privacy.OwnerOnly("created_by"),
    ),
)
```

`And` is the default composition for most entities: tenant isolation AND some visibility restriction. It is also the outer composition when combining `Or` blocks with baseline requirements.

##### 9.5.2. `privacy.Or` — at least one policy must pass

`privacy.Or(policies...)` creates a policy that passes when at least one constituent policy passes. For `EvalQuery`, it collects the filter predicates from all non-denying constituent policies and ORs them together. If all constituent policies return `ErrDeny`, the composition denies.

```go
// Example: Or composition for tiered role access
privacy.Or(
    privacy.RoleFilter("admin", nil),           // admins see all
    privacy.RoleFilter("manager",               // managers see their department
        entity.NewFilter().Eq("dept_id", "$current_dept_id")),
    privacy.OwnerOnly("created_by"),            // others see their own
)
```

> **Warning:** `Or` with constituent policies that return different filter predicates generates an OR compound WHERE clause. On large tables, ORed predicates often cannot use indexes efficiently. Test query plans with `EXPLAIN ANALYZE` when composing `Or` policies over entities with large data volumes.

##### 9.5.3. `privacy.Not` — inversion

`privacy.Not(policy)` inverts a policy. If the constituent policy would pass (return a filter or nil), `Not` denies. If the constituent policy would deny, `Not` allows with no additional filter. This is used to implement "everyone except" rules: a policy that allows access to all records except those owned by a specific system account.

`Not` is uncommon; most access control patterns are expressible with `And` and `Or`. Use `Not` only when the exclusion pattern is cleaner than the equivalent inclusion pattern.

##### 9.5.4. Execution order and short-circuit behaviour

Policies in an `And` composition are evaluated left to right. If the first policy returns `ErrDeny`, the remaining policies are not evaluated (short-circuit). This means the most selective (most likely to deny) policy should be placed first in an `And` composition to avoid unnecessary evaluations.

Policies in an `Or` composition are evaluated left to right. If the first policy allows (returns a non-deny result), the remaining policies are still evaluated to collect their filter predicates for the OR compound predicate. This is different from `And`; `Or` does not short-circuit on the first pass. Place the widest-access policy first in an `Or` to allow the nil-filter case (full access) to suppress unnecessary filter predicate collection for the other constituents.

---

#### 9.6. Testing Privacy Policies

##### 9.6.1. Unit testing a policy with a mock context

Privacy policies are testable without a database because they receive a `context.Context` and return a `Filter`. Construct a test context with mock tenant and user information using `testkit.NewTestContext`:

```go
// Example: Unit testing a custom privacy policy
func TestRegionPolicyFiltersCorrectly(t *testing.T) {
    policy := &RegionPolicy{}

    // User with a specific region
    ctx := testkit.NewTestContext(t,
        testkit.WithUser(testkit.TestUser{
            ID:         "user-001",
            Roles:      []string{"sales_rep"},
            Attributes: map[string]string{"region_code": "NBI"},
        }),
        testkit.WithTenant(testkit.TestTenant{Slug: "dev"}),
    )

    filter, err := policy.EvalQuery(ctx)
    if err != nil {
        t.Fatalf("unexpected deny: %v", err)
    }
    if filter == nil {
        t.Fatal("expected a filter predicate, got nil")
    }
    // Verify the filter contains the expected region predicate
    predicates := filter.Predicates()
    if len(predicates) != 1 {
        t.Fatalf("expected 1 predicate, got %d", len(predicates))
    }
    if predicates[0].Field != "region_code" || predicates[0].Value != "NBI" {
        t.Errorf("unexpected predicate: %+v", predicates[0])
    }
}

func TestRegionPolicyDeniesWithNoRegion(t *testing.T) {
    policy := &RegionPolicy{}
    ctx := testkit.NewTestContext(t,
        testkit.WithUser(testkit.TestUser{
            ID:    "user-002",
            Roles: []string{"sales_rep"},
            // No region_code attribute
        }),
        testkit.WithTenant(testkit.TestTenant{Slug: "dev"}),
    )

    _, err := policy.EvalQuery(ctx)
    if !errors.Is(err, privacy.ErrDeny) {
        t.Errorf("expected ErrDeny, got: %v", err)
    }
}
```

##### 9.6.2. Integration testing — verifying rows are filtered correctly end-to-end

Unit tests verify that the policy returns the correct filter predicate. Integration tests verify that the filter predicate, once injected into a real query, returns the expected rows. Use `testkit.NewIntegrationDB` with seeded records:

```go
// Example: Integration test verifying policy-filtered query results
func TestRegionPolicyEndToEnd(t *testing.T) {
    db := testkit.NewIntegrationDB(t, SalesOrderDefinition)
    tc := testkit.NewTestTenantContext(t, db)

    // Seed two orders in different regions
    adminCtx := testkit.NewTestContext(t,
        testkit.WithUser(testkit.TestUser{ID: "admin", Roles: []string{"admin"}}),
        testkit.WithTenant(tc),
    )
    repo, _ := entity.Resolve(adminCtx, tc, "SalesOrder")
    repo.Create(adminCtx, map[string]any{"region_code": "NBI", "amount": "1000"})
    repo.Create(adminCtx, map[string]any{"region_code": "MSA", "amount": "2000"})

    // Query as a Nairobi sales rep — should see only NBI order
    repCtx := testkit.NewTestContext(t,
        testkit.WithUser(testkit.TestUser{
            ID:         "rep-001",
            Roles:      []string{"sales_rep"},
            Attributes: map[string]string{"region_code": "NBI"},
        }),
        testkit.WithTenant(tc),
    )
    repRepo, _ := entity.Resolve(repCtx, tc, "SalesOrder")
    orders, _, err := repRepo.Query(repCtx, entity.NewFilter())
    if err != nil {
        t.Fatal(err)
    }
    if len(orders) != 1 {
        t.Errorf("expected 1 order, got %d", len(orders))
    }
    if orders[0].Get("region_code") != "NBI" {
        t.Errorf("expected NBI order, got region: %v", orders[0].Get("region_code"))
    }
}
```

##### 9.6.3. Common mistakes — policies that silently pass everything

The most dangerous mistake in privacy policy code is a policy that returns `nil, nil` (no filter, no deny) when it should be restricting access. This typically happens when a context extraction fails silently and the function falls through to a default return. Always fail closed: if tenant or user context cannot be extracted, return `privacy.ErrDeny`, not `nil, nil`.

```go
// Example: Fail-closed pattern — deny on any context error
func (p *MyPolicy) EvalQuery(ctx context.Context) (*entity.Filter, error) {
    uc, err := auth.UserFromContext(ctx)
    if err != nil {
        // CORRECT: fail closed — deny rather than allow on missing context
        return nil, privacy.ErrDeny
    }
    if uc.HasRole("admin") {
        return nil, nil // admins see all
    }
    ownerID := uc.UserID()
    if ownerID == "" {
        // CORRECT: fail closed — deny if user ID is somehow empty
        return nil, privacy.ErrDeny
    }
    return entity.NewFilter().Eq("owner_id", ownerID), nil
}
```

A second common mistake is composing `Or` policies where one branch has no filter (nil) and the other has a restriction: `Or(RoleFilter("admin", nil), OwnerOnly("created_by"))`. When the user is an admin, `RoleFilter` returns nil (pass all), and `Or` correctly produces no filter. But if the nil-return path is caused by a bug rather than an intentional admin grant, the `Or` composition will silently allow all records. Always audit `Or` compositions to verify that every nil-returning branch is an intentional "allow all" decision.

---

#### Chapter summary

Chapter 9 establishes the distinction between RBAC (operation permission) and privacy policies (row and field visibility), explains why application-level WHERE clauses are architecturally insufficient (§9.1.3), and documents the full policy enforcement mechanism at the `EntityRepository` interface layer (§9.1.4). The built-in policy primitives — `TenantIsolation`, `OwnerOnly`, `RoleFilter`, and `DepartmentScope` — cover the majority of access control patterns in multi-tenant ERP systems (§9.3). The three most critical concepts are the fail-closed pattern for policy implementations (§9.6.3, which prevents silent allow-all bugs), the information-leakage prevention in `Get` returning `ErrNotFound` for policy-restricted records (§9.2.2), and the `And`/`Or` composition short-circuit behaviour (§9.5.4, which determines evaluation order and performance characteristics for composed policies).

**Next chapters to read:**

- §16 — Role-Based Access Control (the RBAC layer that runs before privacy policies, determining which operations a user may perform on an entity type; both systems must be understood together to reason about access control completely)
- §10 — Custom Fields — Runtime Schema Extension (custom entities use JSONB predicates in `EvalQuery` returns; understanding how custom field filters interact with policy predicates requires reading both chapters)
- §51 — Testing Guide (the `testkit` helpers used in §9.6 are documented fully in the testing guide, including all available context builder options and the integration DB lifecycle management)
