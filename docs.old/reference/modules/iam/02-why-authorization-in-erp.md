[<-- Back to Index](README.md)

## Why Authorization in ERP

### The Problem Without Centralized Authorization

Without a unified authorization layer, access control gets scattered across modules as ad-hoc `if user.Role == "admin"` checks. This is how every major ERP data breach starts.

```markdown
WHAT HAPPENS WITHOUT CENTRALIZED AUTHZ:

Module A (Finance):
  if user.IsAdmin { return invoices }
  → No resource-level control
  → Any admin can see any invoice

Module B (Sales):
  if user.Department == "sales" { return orders }
  → Department-based, not role-based
  → No wildcard or hierarchy

Module C (HR):
  if user.EmployeeLevel >= 5 { return salaries }
  → Arbitrary numeric level
  → Impossible to audit
  → Impossible to revoke cleanly

Result:
  → 3 different authorization models in 3 modules
  → No single place to ask "what can user X do?"
  → No audit trail
  → Cannot enforce company-wide policies
  → Compliance certification impossible
```

### The ERP Access Control Problem Is Harder Than Most Applications

ERP systems face authorization challenges that typical web apps don't:

**Multi-tenancy**: 500 tenants × 50 users × 200 resources × 20 actions = millions of policy combinations. Policies must be isolated per tenant.

**Multi-actor types**: A platform operator, a tenant CFO, a customer-portal viewer, and an API integration client all use the same underlying resources but with radically different trust levels and scopes.

**Cross-module operations**: Approving a purchase order requires checking permissions in both Finance (budget) and Buying (procurement). The authorization engine must handle this without each module importing the other.

**Time-limited access**: An external auditor needs read access to last year's ledger for exactly 30 days. A contractor needs write access to project tasks until the contract end date. Both must be automatically revoked.

**Hierarchical permissions**: A Sales Manager should be able to do everything a Sales Representative can, plus approve discounts. Role inheritance must be explicit and auditable.

**Wildcard resources**: A policy saying "can read all invoices" must match `invoice/inv_2026_001`, `invoice/inv_2026_002`, etc., without creating one rule per document.

### How the authz Module Solves All of This

```markdown
SOLUTION MAPPING:

Problem → Solution
─────────────────────────────────────────────────────────────
Multi-tenancy                → Domain namespacing (v1 in Casbin rule)
                               Each tenant's policies in its own domain
                               Platform in reserved "_platform_" domain

Multi-actor types            → 4 subject prefixes:
                               platform: / tenant: / portal: / api:
                               Each resolves to its own domain

Cross-module operations      → Single Service interface imported by all modules
                               Enforce(ctx, Request{...}) returns bool
                               No module-to-module imports needed

Time-limited access          → role_assignments.expires_at
                               Lazy revoke in Enforce() hot path
                               Indexed query before every decision

Hierarchical permissions     → Casbin g-rule: g(user, role, domain)
                               AddRoleForUserInDomain chains roles
                               Role hierarchy defined once, applies everywhere

Wildcard resources           → keyMatch2 in Casbin matcher
                               "invoice/*" matches any invoice
                               "report/finance/*" matches all finance reports
                               No per-document policy records needed

Deny-override                → e = some(allow) && !some(deny)
                               One deny rule beats all allows
                               Sanctions/embargo enforcement trivial
```

### Real-World Business Justification

**"Why not just use database-level RLS for everything?"**

Row-Level Security handles *data isolation between tenants* but cannot express *within-tenant authorization*. It cannot answer: "Can user Alice in Tenant ABC approve invoices over KES 500,000?" RLS has no concept of actions, roles, or wildcards.

**"Why Casbin instead of a custom rule engine?"**

The previous custom engine (`abac/`, `access/`, `iam/`) was ~2,000 lines duplicating what Casbin provides out of the box. Casbin is:
- Battle-tested in production at scale globally
- Formally verified policy semantics
- Supported by a large community
- Extensible with custom functions (e.g., `keyMatch2`, `ipMatch`, `regexMatch`)
- Single dependency, zero transitive surprises

**"Isn't a centralized authz service a single point of failure?"**

The `authz.Service` is in-process — same binary as the application. There is no network hop, no gRPC call, no separate service to go down. It uses PostgreSQL as durable storage but holds a live in-memory copy. If the database is unavailable, the last known policy continues to serve requests until the DB recovers.

### Industry Standards This Module Implements

| Standard | How authz implements it |
|----------|------------------------|
| **RBAC** (NIST 800-63) | g-rules provide named roles with domain scoping |
| **ABAC** | Resource wildcards enable attribute-like matching on object paths |
| **Least Privilege** | Deny-override means default is deny; explicit allows required |
| **Separation of Duties** | Platform domain isolated from all tenant domains |
| **Audit Trail** | `role_assignments` records every grant and revocation |
| **Time-Bound Access** | `expires_at` on every role assignment |

---

Next: [Architecture Overview](./03-architecture-overview.md)
