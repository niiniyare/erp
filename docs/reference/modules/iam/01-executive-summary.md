[<-- Back to Index](README.md)

## Executive Summary

### Module Purpose

The Authorization Module (`internal/core/authz`) is the security backbone of the entire AWO ERP platform. Every data-access decision — whether a user can read an invoice, approve a payment, export a report, or call an API endpoint — flows through this single module. It provides:

- **Unified Authorization Engine**: One `Service` interface used by all modules — finance, sales, HR, inventory, workflows, and the API gateway
- **4-Actor Domain Model**: Clean separation of Platform administrators, Tenant users, Portal users, and API clients — each fully isolated
- **RBAC + ABAC via Casbin**: Role-based access control with wildcard object patterns (`invoice/*`, `report/*`) replacing ~2,000 lines of custom policy evaluation
- **Temporal Roles**: Role assignments with optional expiry dates — perfect for contractors, auditors, and time-limited access grants
- **Deny-Override Policy**: One explicit `deny` rule beats all `allow` rules — critical for compliance and sanctions checking
- **Fiber Middleware Factory**: One-line route protection with automatic resource-ID expansion
- **Zero Cross-Module Coupling**: Imports nothing from `abac/`, `access/`, or `iam/` — self-contained and independently testable

### Why This Module Is the Most Critical in the System

```markdown
IMPACT RADIUS — what breaks if authz fails:

Every HTTP route in the platform:
  → API Gateway routes check authz before forwarding
  → Fiber middleware blocks unauthorized requests

Every background workflow:
  → Temporal activities validate actor permissions before DB writes
  → Scheduled jobs authenticate as service principals

Every business operation:
  → Finance: cannot post journals, approve invoices, close periods
  → Sales: cannot create orders, apply discounts, export reports
  → HR: cannot access payroll, employee records, performance data
  → Inventory: cannot adjust stock, create POs, view costs

Data isolation:
  → Without domain isolation, Tenant A can read Tenant B's data
  → Without platform separation, tenant admins can modify system config
  → Without temporal expiry, ex-employees retain access indefinitely

Compliance:
  → SOC 2, ISO 27001, GDPR all require demonstrable access control
  → Audit trails must show every authorization decision
  → Deny rules must be provably enforced
```

### System Architecture Context

The authz module sits at the intersection of every other module in the platform:

```markdown
CORE SYSTEM CAPABILITIES PROVIDED:

Identity & Access:
✓ 4 actor types with distinct trust levels
✓ Domain-scoped RBAC — roles only apply within their domain
✓ Wildcard policy objects — one rule covers all resources of a type
✓ Deny-override effect — explicit deny always wins

Security Guarantees:
✓ Cross-domain isolation enforced in policy engine
✓ Temporal roles automatically revoked on expiry
✓ Policy changes immediately reflected (in-memory cache)
✓ Role assignments audited in role_assignments table

Integration Surface:
✓ Fiber middleware for HTTP routes (zero boilerplate)
✓ Direct Service.Enforce() for non-HTTP contexts (workflows, jobs)
✓ BatchEnforce for permission pre-computation
✓ InvalidateCache for admin-driven policy reloads

Performance:
✓ In-memory policy cache — sub-millisecond enforcement
✓ Lazy expiry check — O(1) indexed query before enforce
✓ BatchEnforce for bulk permission checks
✓ PostgreSQL as durable policy store — survives restarts
```

### Key Stakeholders

| Role | How authz affects them |
|------|----------------------|
| Platform Administrator | Full access to `_platform_` domain; can manage all tenants' policies |
| Tenant Administrator | Controls roles and policies within their tenant's domains |
| Application Developer | Uses `svc.Middleware(obj, act)` to protect routes — one line per route |
| Security / Compliance | Can audit the `casbin_rule` and `role_assignments` tables directly |
| Business Users | See only resources they are authorized for; receive 403 otherwise |
| API Clients | Authenticate as `api:clientID` actor; scoped to `{tenantID}:api` domain |
| Portal Users | Authenticate as `portal:userID`; scoped to `{tenantID}:portal` domain |
| Workflow Engine | Service accounts call `Enforce()` directly before mutating data |

### Success Metrics

| Metric | Target |
|--------|--------|
| Enforcement latency (cached) | < 1 ms p99 |
| Enforcement latency (cold start) | < 5 ms |
| Policy reload time | < 500 ms for 10,000 rules |
| Authorization failures logged | 100% |
| Cross-domain policy leaks | Zero tolerance |
| Temporal role cleanup lag | ≤ 1 request (lazy revoke on next Enforce) |

---

Next: [Why Authorization in ERP](./02-why-authorization-in-erp.md)
