[<-- Back to Index](README.md)

## Domain Isolation

### Why Domain Isolation Is the Core Security Property

Domain isolation is the guarantee that a policy in one domain **never** affects enforcement in another domain. It is the authz module's equivalent of the Tenant Module's Row-Level Security.

```markdown
ISOLATION GUARANTEE:

Tenant A's finance manager policy:
  p | role:finance-manager | tenant-A-uuid | invoice/* | * | allow

This rule CANNOT be triggered by a request in tenant B:
  Enforce(Request{
      Subject: "tenant:usr_b1",
      Domain:  "tenant-B-uuid",   ← different domain
      Object:  "invoice/123",
      Action:  "read",
  })
  → Result: DENY (no matching p-rule in tenant-B domain)
  → The tenant-A rule is NEVER evaluated

WHY THIS IS GUARANTEED:
  The Casbin matcher requires: r.dom == p.dom
  If request domain ≠ policy domain → rule does not match
  No cross-domain evaluation is possible at the engine level
```

### The Four Domains and Their Boundaries

```markdown
DOMAIN HIERARCHY:

_platform_
│
│  Completely isolated. No relationship to tenant domains.
│  Platform policies ONLY apply when r.dom == "_platform_"
│  A request with r.dom = "tenant-abc" NEVER matches platform policies.
│
├─ tenant-A-uuid
│    Internal business operations for Tenant A
│    Tenant A's employees (tenant:{id}) act here
│    No access to portal or api sub-domains from here (by default)
│
├─ tenant-A-uuid:portal
│    External facing operations for Tenant A's portal
│    Portal users (portal:{id}) act here only
│    Cannot access tenant-A-uuid domain (different domain string)
│    Tenant A admins can add portal policies from here
│
└─ tenant-A-uuid:api
     Machine client operations for Tenant A's API clients
     API clients (api:{id}) act here only
     Cannot access tenant-A-uuid domain (different domain string)
```

### Cross-Domain Policies (Explicitly Granted)

By default, domains are completely isolated. Cross-domain access requires an explicit policy grant. These are rare and must be carefully considered.

```markdown
USE CASE: Tenant admin wants to view portal activity.

Option 1: Duplicate policies in both domains (not recommended)
  → Maintenance burden

Option 2: Explicitly cross-grant the tenant user into the portal domain
  AssignRole(ctx, tenantID,
      "tenant:usr_admin",     ← subject is a tenant actor
      "role:portal-admin",    ← role defined in portal domain
      authz.PortalDomain(tenantID),  ← PORTAL domain, not tenant domain
  )
  → Now usr_admin can Enforce() in the portal domain

USE CASE: Service account (api:) needs to write invoices in tenant domain.
  AssignRole(ctx, tenantID,
      "api:cli_sync",
      "role:api-invoice-submit",
      authz.TenantDomain(tenantID),  ← tenant domain, not api domain
  )
  → api client gets access in tenant domain
  → This is a deliberate elevated grant — audit carefully
```

### Platform Cross-Tenant Policies

Platform actors operate in `_platform_` domain. To read or manage data across all tenants, the convention is to use a path prefix of `tenant/*/`:

```markdown
PLATFORM CROSS-TENANT RESOURCE PATHS:

tenant/*/invoice          All tenants' invoices
tenant/*/payment          All tenants' payments
tenant/*/user             All tenants' users
tenant/{id}/invoice       Specific tenant's invoices
tenant/{id}/config        Specific tenant's configuration

Platform policy example:
  p | platform:support    | _platform_ | tenant/*/invoice | read   | allow
  p | platform:billing    | _platform_ | tenant/*/payment | read   | allow
  p | platform:admin      | _platform_ | tenant/*         | *      | allow
  p | platform:support    | _platform_ | tenant/*/config  | *      | deny
                                                            ↑ support cannot change config

IMPORTANT: The platform domain "_platform_" is ONLY checked when
the request explicitly uses dom = "_platform_".
Routes served to platform operators set the Principal domain to "_platform_".
Routes served to tenant users set the Principal domain to the tenant UUID.
The two NEVER mix in the same request.
```

### Preventing Domain Escalation Attacks

Domain isolation also prevents a class of escalation attacks where a compromised token tries to access resources in a higher-trust domain.

```markdown
ESCALATION ATTEMPT 1: Tenant user tries platform domain

  Attacker has: valid JWT for "tenant:usr_hacker" in "tenant-abc"
  Attempt:      Modify request to use domain = "_platform_"

  Protection:
    The authn middleware extracts domain from the JWT claims.
    JWT domain is signed by the server. Attacker cannot change it.
    Even if they could: Enforce(sub="tenant:usr_hacker", dom="_platform_", ...)
    → No p-rule or g-rule matches this sub in _platform_ domain
    → DENY

ESCALATION ATTEMPT 2: Portal user tries tenant domain

  Attacker has: valid portal JWT for "portal:cust_123"
  Attempt:      Craft request targeting "tenant-abc" domain (not portal domain)

  Protection:
    authn middleware reads actor type from JWT ("portal").
    Portal tokens always resolve to "{tenantID}:portal" domain.
    Even direct API manipulation: Enforce(sub="portal:cust_123", dom="tenant-abc", ...)
    → No g-rule for portal:cust_123 in tenant-abc domain
    → DENY

ESCALATION ATTEMPT 3: API client tries portal domain

  api:{clientID} has policies in {tenantID}:api domain.
  No rules in {tenantID}:portal domain.
  Enforce request targeting portal domain → DENY.
```

### Domain Isolation and Tenant Deletion

When a tenant is deleted (hard delete), the `role_assignments` rows cascade-delete automatically via the FK. The `casbin_rule` rows do **not** have an FK and must be cleaned up separately:

```markdown
TENANT DELETION CLEANUP SEQUENCE:

1. Tenant module calls: authz service before hard delete:
   svc.CleanDomain(ctx, tenantID)
   → Removes all casbin_rule rows where v1 = tenantID (p-rules)
   → Removes all casbin_rule rows where v2 = tenantID (g-rules)
   → Removes portal domain: v1 = tenantID+":portal"
   → Removes api domain:    v1 = tenantID+":api"

   Implementation: enforcer.RemoveFilteredPolicy(1, tenantID)
                   enforcer.RemoveFilteredGroupingPolicy(2, tenantID)

2. role_assignments cascade-deletes via FK automatically.

3. enforcer.LoadPolicy() to sync in-memory state.

Note: CleanDomain is NOT yet in the Service interface (Phase 2 addition).
For now, platform admins clean up via direct SQL + InvalidateCache.
```

---

Next: [Middleware & HTTP Integration](./11-middleware-and-http.md)
