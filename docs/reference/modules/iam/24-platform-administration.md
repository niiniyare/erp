[<-- Back to Index](README.md)

## Platform Administration — Operational Guide

> **Implementation status**: This document describes the v1.0 platform authority model.
> Platform roles are seeded in code and DB migrations. The platform super admin is seeded at application boot.

---

### Who Is This For?

This guide covers the **platform level** of Awo ERP — the layer above individual tenants. Platform users are Awo staff (engineers, support, billing, compliance) who operate on the infrastructure itself. This document is for platform engineers, operations teams, and security auditors.

---

### 1. Platform vs Tenant Users

#### The Four Actor Types

Awo ERP recognises four actor types (see `internal/core/iam/domain/authz.go`):

| Actor Type | UserType Values | Casbin Domain | Purpose |
|---|---|---|---|
| `ActorPlatform` | `SYSADMIN`, `PLATFORM` | `_platform_` | Awo staff operating cross-tenant |
| `ActorTenant` | `INTERNAL`, `EMPLOYEE` | `<tenantID>` | Tenant employees and admins |
| `ActorPortal` | `PORTAL`, `CUSTOMER` | `<tenantID>:portal` | External portal users (customers, suppliers) |
| `ActorAPI` | `API`, `SERVICE` | `<tenantID>:api` | Machine principals (API keys, integrations) |

The Casbin domain is derived from the actor type at session construction time (`ToPrincipal()`). The domain **always** comes from the authenticated session — never from a request parameter.

#### Platform Users

Platform users operate in the `_platform_` Casbin domain. This is a reserved constant:
```go
const DomainPlatform = "_platform_"
```

Their subjects are prefixed `platform:<userID>`. A platform policy such as `(platform:uuid, _platform_, tenants/*, read, allow)` grants cross-tenant read access to tenant data — something no tenant user can have.

#### Why Mixing Is Dangerous

Assigning a platform-domain role to a tenant-domain user (or vice versa) is architecturally incorrect:

- A `platform:*` subject in a tenant domain would need explicit tenant-domain policies to do anything — it would have no extra access.
- A `tenant:*` subject added to the platform domain would gain cross-tenant visibility — a privilege escalation.

The Casbin matcher `r.dom == p.dom` provides a hard boundary: a subject in one domain cannot match policies in another domain. But defense-in-depth requires never creating g-rules that cross domain types. The service-layer guard (AUTHZ-4, currently open) is intended to enforce this programmatically. See [Security Considerations](./17-security-considerations.md).

#### Service / API Users

Service accounts use `UserType = "API"` or `"SERVICE"`. They authenticate via API keys (not passwords) and operate in the `<tenantID>:api` domain. See [API Keys and Service Accounts](./26-api-keys-and-service-accounts.md).

---

### 2. Platform Roles

The following platform roles are defined by convention and seeded during bootstrap. These operate exclusively in the `_platform_` Casbin domain.

| Role Name | Purpose | Access |
|---|---|---|
| `platform_super_admin` | Full platform authority | Unrestricted: all platform operations, cross-tenant read/write for support |
| `platform_support` | Tenant support operations | Read tenant data, cannot modify business data |
| `platform_operator` | Operational tasks | Tenant provisioning, module enablement, feature flag management |
| `platform_billing` | Billing and subscription | Subscription data, payment records, plan management |
| `platform_compliance` | Audit and compliance | Read-only access to audit logs, configuration history, compliance reports |

**Note on seeding**: Platform roles are seeded via `SeedDefaultRoles` (for tenant roles) and equivalent platform bootstrap code. Verify the current seed state by querying:
```sql
SELECT v0, v1, v2 FROM casbin_rule
WHERE ptype = 'p' AND v1 = '_platform_'
ORDER BY v0, v2;
```

For the current list of seeded platform policies, see `internal/core/iam/seed.go` and any `BootstrapPlatform` or similar functions in the codebase.

---

### 3. Platform Super User Bootstrap

#### How the Initial Platform Admin Is Created

The platform super admin is created at application boot. Based on the code architecture:

1. The `BootstrapTenantAdmin` function in `service/authz.go` handles per-tenant bootstrapping.
2. For the platform, an equivalent bootstrap must be triggered for the `_platform_` domain.
3. See `internal/core/iam/seed.go` and application boot code for the exact mechanism.

The initial platform user is typically created via:
- A dedicated migration or seed script that runs once
- An environment variable (`PLATFORM_ADMIN_EMAIL`, etc.) consumed at first boot
- A CLI command available only in deployment infrastructure

Because platform admin access is cross-tenant and unrestricted, this bootstrap is intentionally restricted to infrastructure-level operations — it cannot be triggered via the API.

#### Why This Is Intentionally Restricted

If any authenticated user could elevate themselves to platform admin, the entire tenant isolation model would be compromised. The bootstrap mechanism is therefore:
- Executable only by infrastructure operators with server access
- Not exposed via any HTTP endpoint
- Protected by DB admin_role access for direct SQL fallback

#### Emergency Recovery

If no platform super admin exists (e.g., credentials lost, bootstrap failed):

1. **Preferred**: Use infrastructure access (DB admin_role) to directly insert a Casbin g-rule:
   ```sql
   -- Execute as admin_role
   INSERT INTO casbin_rule (ptype, v0, v1, v2)
   VALUES ('g', 'platform:<userID>', 'platform_super_admin', '_platform_')
   ON CONFLICT DO NOTHING;
   ```
   Then call `InvalidateCache()` on all running instances to reload the policy.

2. **Alternative**: Re-run the bootstrap migration/seed in a maintenance window.

3. **After recovery**: Audit all `_platform_` domain policies and role assignments.

#### Why Tenant Admins Can Never Create Platform Users

Tenant admins operate in their tenant's Casbin domain. The `AddPolicy` and `AssignRole` calls in the `AuthzService` write to whatever domain is specified — but a tenant admin's HTTP session produces a `Principal` with `Domain = <tenantID>`, not `_platform_`. Any policy the tenant admin creates is scoped to their tenant domain.

The service-layer guard (AUTHZ-4, currently open) will add explicit validation: `if domain == DomainPlatform { return ErrForbidden }` unless the caller is themselves a platform actor.

---

### 4. Platform User Operational Boundaries

Platform users are restricted by their role's policies in the `_platform_` domain. Even `platform_super_admin` should follow operational boundaries as governance policy (not just technical enforcement).

#### Support Engineers (`platform_support`)

Capability: Read tenant data for troubleshooting.
- Can read user records, session logs, configuration state
- Cannot modify tenant business data (invoices, transactions, HR records)
- Cannot modify tenant IAM policies or role assignments
- All access is logged via OTel spans (and will be via the future persistent audit log — AUTHZ-7)

Typical use: "A tenant reports invoice #1234 shows wrong amount — support reads the transaction log."

#### Billing Operators (`platform_billing`)

Capability: Subscription and payment data management.
- Can read/update subscription plans and billing records
- Cannot access tenant business data beyond what is necessary for billing
- Cannot read user passwords, session tokens, or security-sensitive IAM data

Typical use: "Apply a promotional discount to tenant X's next billing cycle."

#### Compliance Auditors (`platform_compliance`)

Capability: Audit logs and compliance data — read only.
- Can read configuration audit history (`configuration_audit` table)
- Can read role assignment history (`role_assignments` table)
- Cannot modify any data
- Cannot read personally identifiable data beyond what is in audit records

Typical use: "Generate SOC 2 evidence for tenant Y's access control changes over 90 days."

#### Infrastructure Operators (`platform_operator`)

Capability: Operational platform tasks.
- Can provision new tenants (trigger `SeedDefaultRoles`, `BootstrapTenantAdmin`)
- Can enable/disable modules and feature flags for tenants
- Cannot modify tenant business data
- Cannot escalate to super admin rights

---

### 5. Platform Authority Model

#### How Platform Policies Are Stored

Platform policies are Casbin p-rules with `domain = '_platform_'` stored in the `casbin_rule` table:

```
ptype | v0 (sub)               | v1 (dom)     | v2 (obj)        | v3 (act) | v4 (eft)
------+------------------------+--------------+-----------------+----------+---------
p     | platform_super_admin   | _platform_   | *               | *        | allow
p     | platform_support       | _platform_   | tenants/*       | read     | allow
p     | platform_operator      | _platform_   | tenants/*/flags | *        | allow
```

Platform g-rules (role assignments):
```
ptype | v0 (sub)               | v1 (role)             | v2 (dom)
------+------------------------+-----------------------+----------
g     | platform:<userID>      | platform_super_admin  | _platform_
```

#### Casbin Domain Isolation

The `_platform_` domain is completely isolated from all tenant domains. The Casbin matcher requires `r.dom == p.dom`:
- A `platform:*` subject making a request with `domain = <tenantID>` will find no matching p-rules in that domain.
- A `tenant:*` subject making a request with `domain = _platform_` will find no matching p-rules in that domain (unless explicitly added — AUTHZ-4 gap).

This is the Casbin layer of the three-layer isolation model. The PostgreSQL RLS layer provides the second boundary (tenant_id columns), and the DB admin_role vs application_role separation provides the third.

#### Monitoring Platform Domain Integrity

Regularly audit the `_platform_` domain for unexpected rules:

```sql
-- All platform-domain subjects — should only be known platform:* UUIDs
SELECT DISTINCT v0 FROM casbin_rule
WHERE ptype = 'g' AND v2 = '_platform_';

-- All platform-domain policies — review for unexpected wildcards or objects
SELECT v0, v2, v3, v4 FROM casbin_rule
WHERE ptype = 'p' AND v1 = '_platform_'
ORDER BY v0, v2;
```

Set up monitoring alerts for:
- Any new `INSERT` into `casbin_rule` where `v1 = '_platform_'` or `v2 = '_platform_'`
- Any `platform:*` subject accessing sensitive resources (`payroll/*`, `salary/*`, `users/*/password`)

---

See also:
- [Tenant Administration](./23-tenant-administration.md)
- [Security Considerations](./17-security-considerations.md)
- [RBAC Enforcement](./rbac-enforcement.md)
- [Deferred Features](./deferred-features.md)
