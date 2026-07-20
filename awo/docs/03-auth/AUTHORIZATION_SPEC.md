# Authorization Specification

**Classification:** Specification — Tier 0
**Owner:** `03-auth/AUTHORIZATION_SPEC.md`
**Status:** Frozen at v1.0 (ADR-001, ADR-011)
**Package:** `awo.so/awo/auth`, `awo.so/awo/compiler`

---

## Purpose

This document specifies the complete authorization model: how module authors declare permissions, how the compiler compiles them to `CapabilityGrant` values, and how the `PolicyEvaluator` enforces them at runtime.

## Scope

- `PermissionSet` declaration on EntityDefinition
- `CapabilityGrant` struct (compiler output)
- `PolicyEvaluator` interface and its Casbin default implementation
- Authorization flow for every request
- Platform admin bypass
- Policy function (row-level authorization filter)

## Dependencies

- [`03-auth/VIEWER_CONTEXT.md`](VIEWER_CONTEXT.md) — ViewerContext is the authorization subject
- [`01-entity/ENTITY_DEFINITION_SPEC.md`](../01-entity/ENTITY_DEFINITION_SPEC.md) — PermissionSet on EntityDefinition
- [`05-compiler/COMPILE_SPEC.md`](../05-compiler/COMPILE_SPEC.md) — Compilation of PermissionSet to CapabilityGrant

---

## 1. Two-Layer Authorization Architecture (ADR-001)

Authorization in Awo has two distinct layers:

**Layer 1 — Declaration (in `awo/def`, frozen):**
`PermissionSet` on `EntityDefinition`. Module authors declare which role subjects may perform which operations. This is a Go struct — statically declared, validated at compile time.

**Layer 2 — Enforcement (in `awo/auth`, replaceable):**
`PolicyEvaluator` interface. The runtime consults it on every request. The default implementation is Casbin-backed. It can be replaced with OPA, ReBAC, or a custom engine without touching any EntityDefinition.

**The separation is intentional.** The declaration is stable and engine-agnostic. The enforcement is replaceable. Frappe (ERPNext) conflated these two layers, producing complex role inheritance resolution at request time. Awo keeps them separate.

---

## 2. PermissionSet Declaration

```go
// Package: awo.so/awo/def

type PermissionSet struct {
    // Create contains role subjects that may create records of this entity.
    Create []string

    // Read contains role subjects that may read records of this entity.
    // Applies to Get, Query, Count, Exists, and related operations.
    Read []string

    // Write contains role subjects that may update records of this entity.
    Write []string

    // Delete contains role subjects that may delete records of this entity.
    Delete []string

    // Actions maps custom action names to their permitted role subjects.
    // Keys are action names as declared in ActionDef.Name.
    Actions map[string][]string
}
```

Role subject format: `"role:{domain}.{name}"` or `"role:{name}"`.

**Example:**
```go
Permissions: def.PermissionSet{
    Create: []string{"role:finance.accounts_payable", "role:tenant.admin"},
    Read:   []string{"role:finance.viewer", "role:tenant.admin"},
    Write:  []string{"role:finance.accounts_payable", "role:tenant.admin"},
    Delete: []string{"role:tenant.admin"},
    Actions: map[string][]string{
        "submit":   {"role:finance.accounts_payable", "role:tenant.admin"},
        "approve":  {"role:finance.approver", "role:tenant.admin"},
        "cancel":   {"role:tenant.admin"},
    },
},
```

**Empty PermissionSet:** An entity with no subjects declared for an operation is inaccessible for that operation (deny by default). Only platform admins can access operations with no declared subjects.

---

## 3. CapabilityGrant (Compiler Output, ADR-011)

The compiler reads `PermissionSet` declarations and produces `CapabilityGrant` slices:

```go
// Package: awo.so/awo/compiler

// CapabilityGrant is a compiled authorization assertion.
// Renamed from CasbinPolicy to decouple the compiler from the Casbin engine.
type CapabilityGrant struct {
    Subject string  // "role:finance.accounts_payable"
    Object  string  // "finance_invoice" (qualified entity name)
    Action  string  // "create", "read", "write", "delete", or custom action name
}
```

`CompiledSchema.CapabilityGrants` contains all grants from all registered entities. At startup, the runtime loads these into the `PolicyEvaluator` implementation (e.g., into Casbin's policy store).

**Engine agnosticism:** The compiler produces `CapabilityGrant` values regardless of which authorization engine is used. The Casbin implementation reads `CapabilityGrant` slices. An OPA implementation would read the same slices. The compiler knows nothing about Casbin.

---

## 4. PolicyEvaluator Interface

```go
// Package: awo.so/awo/auth

// PolicyEvaluator is the single enforcement point for all capability decisions.
// The framework consults it at the AUTHORIZE pipeline stage for every mutation
// and before every route handler.
//
// The default implementation is Casbin-backed. Replace with OPA, ReBAC,
// or a custom engine by providing an alternative implementation at startup.
type PolicyEvaluator interface {
    // CanPerform returns true if the viewer may perform action on object.
    //
    // object is the qualified entity name (e.g., "finance_invoice").
    // action is one of: "create", "read", "write", "delete", or a custom action name.
    //
    // Implementations MUST be goroutine-safe.
    // Implementations MUST be fast (called on every request) — avoid I/O.
    CanPerform(ctx context.Context, viewer ViewerContext, object string, action string) (bool, error)
}
```

---

## 5. Authorization Flow

For every authenticated request:

```
1. Middleware extracts Session → constructs ViewerContext

2. viewer.IsPlatformAdmin() == true?
   → YES: AUTHORIZE stage is skipped entirely
   → NO: continue

3. PolicyEvaluator.CanPerform(ctx, viewer, qualifiedEntityName, action)
   → (true, nil): proceed to pipeline
   → (false, nil): HTTP 403 PermissionError
   → (_, error): HTTP 500 internal error

4. Policy filter (privacy policy) applied to query predicates:
   → PolicyFunc returns a *filter.Filter
   → filter is AND'd with the request predicate
   → Data that does not match the filter is invisible to this viewer
```

---

## 6. Casbin Default Implementation

The default `PolicyEvaluator` uses Casbin with the following configuration:

**Model:**
```
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
```

**Policy loading:** At startup, `CompiledSchema.CapabilityGrants` are loaded as Casbin `p` assertions. Casbin role inheritance (`g` assertions) is loaded from the IAM module's role hierarchy.

**Thread safety:** Casbin enforcer is safe for concurrent reads after policy loading. Policy reload (for dynamic role changes) uses write-lock swap.

---

## 7. Privacy Policies (Row-Level Filters)

RBAC controls operation-level access. Privacy policies control which rows are visible within an operation.

```go
// Declared on EntityDefinition — a PolicyFunc for row filtering:
Policy: def.PolicyFunc(func(ctx context.Context) def.Filter {
    viewer := auth.ViewerFromContext(ctx)
    return filter.Eq("assigned_to", viewer.UserID())
})
```

The runtime calls `PolicyFunc` on every query and AND's its result with the caller's filter. Records that don't match are invisible — not forbidden, just absent.

**Built-in policies:**

| Policy | Filter applied |
|--------|--------------|
| `TenantIsolation` | Applied automatically by RLS — not a PolicyFunc |
| `OwnerOnly` | `filter.Eq("created_by", viewer.UserID())` |
| `BranchScoped` | `filter.Eq("branch_id", viewer.BranchID())` (v1.1) |

**Difference from RBAC:** RBAC determines if you can access the entity at all. PolicyFunc determines which records you can see within that access. Both are required and complementary.

---

## 8. Built-in Roles

See [`03-auth/RBAC_ROLES_REFERENCE.md`](RBAC_ROLES_REFERENCE.md) for the complete role list.

System roles (seeded at tenant bootstrap, cannot be deleted):

| Role | Scope |
|------|-------|
| `role:platform-admin` | Platform-wide; bypasses all Casbin checks |
| `role:tenant.admin` | Full access within one tenant |
| `role:tenant.user` | Standard user; customized per tenant |
| `role:api-client` | Machine-to-machine; limited scopes |

---

## 9. Normative Requirements

- The AUTHORIZE pipeline stage MUST consult `PolicyEvaluator.CanPerform()` for every mutation.
- Platform admin actors (IsPlatformAdmin() == true) MUST bypass the AUTHORIZE stage.
- `PolicyEvaluator` implementations MUST be goroutine-safe.
- `PolicyEvaluator.CanPerform()` MUST NOT perform database I/O on the hot path.
- The `CapabilityGrant` type MUST be the exclusive output format from the compiler — `CasbinPolicy` is permanently retired.
- Privacy policies (`PolicyFunc`) MUST be AND'd with every query predicate — never OR'd.
- Deny-by-default: operations with no declared subjects in `PermissionSet` MUST be inaccessible to non-platform-admin actors.

---

## References

- `awo/auth/evaluator.go` — PolicyEvaluator interface
- `awo/compiler/schema.go` — CapabilityGrant struct, emitPolicies()
- [`03-auth/VIEWER_CONTEXT.md`](VIEWER_CONTEXT.md) — ViewerContext
- [`03-auth/RBAC_ROLES_REFERENCE.md`](RBAC_ROLES_REFERENCE.md) — Built-in roles
- [`06-filter/FILTER_POLICY_PATTERNS.md`](../06-filter/FILTER_POLICY_PATTERNS.md) — PolicyFunc patterns
- ADR-001 and ADR-011 in [`00-overview/DECISION_REGISTER.md`](../00-overview/DECISION_REGISTER.md)
