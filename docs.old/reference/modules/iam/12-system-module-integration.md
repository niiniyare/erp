[<-- Back to Index](README.md)

## System Module Integration

The authz module does not import any system module — but every system module either informs authz behavior or is informed by it. This chapter documents those relationships.

---

### Integration with Tenant Module

The Tenant Module is the most fundamental dependency. Tenants define the domains that authz policies live in.

```markdown
RELATIONSHIP:
  Tenant ID → authz Domain

  Every active tenant has:
    - A primary domain:    "{tenantID}"
    - A portal sub-domain: "{tenantID}:portal"
    - An api sub-domain:   "{tenantID}:api"

WHAT TENANT MODULE DOES FOR AUTHZ:
  1. Provisions the tenant → caller bootstraps default authz policies
  2. Activates the tenant  → authz policies become live
  3. Suspends the tenant   → caller should add blanket deny or remove policies
  4. Archives the tenant   → caller should call CleanDomain(tenantID)
  5. Hard-deletes tenant   → role_assignments CASCADE-DELETE automatically

AUTHZ BEHAVIOR DURING TENANT LIFECYCLE:
  PENDING:
    → Policies can be written but enforced normally
    → Typically: only tenant-admin role assigned during setup

  ACTIVE:
    → Full policy enforcement
    → All user role assignments active

  SUSPENDED:
    Option A: Add blanket deny at engine level
      AddPolicy(Policy{Subject:"*", Domain:tenantDomain, Object:"*", Action:"*", Effect:"deny"})
      → Fast: one rule blocks everyone
    Option B: Remove all policies + add deny (cleaner)
    Option C: Block at authn layer (JWT validation rejects suspended tenant tokens)
      → authz module doesn't know tenant status — that's the authn middleware's job
    RECOMMENDED: Block at authn middleware, not at authz policy level.
                 Keep authz policies intact for when tenant is reactivated.

  ARCHIVED:
    → Call service.InvalidateCache after cleanup
    → role_assignments already gone (CASCADE DELETE)
    → casbin_rule rows must be cleaned via RemoveFilteredPolicy

ROLE ASSIGNMENTS FOR TENANT PROVISIONING:

When platform admin provisions a new tenant, the bootstrap code assigns:
  AssignRole(ctx, tenantID, "tenant:{newAdminUserID}", "role:tenant-admin", dom,
      WithAssignedBy("platform:provisioning-svc"),
  )

This gives the first tenant admin full access to set up their own roles/policies.
```

---

### Integration with Settings Module

The Settings Module controls feature flags and module enablement. Authorization decisions respect these settings at the application layer — not inside the authz engine itself.

```markdown
RELATIONSHIP:
  Settings → what objects/actions are AVAILABLE to authorize

  If Finance module is disabled for a tenant:
    → Route handlers never register for that tenant
    → authz middleware is never reached
    → No finance policies needed

  If a feature flag "invoice_approval_workflow" is OFF:
    → Route /invoices/:id/approve is not registered
    → authz policy for "invoice/*/approve" can exist but is unreachable
    → No conflict — policy engine doesn't know routes

PRACTICAL PATTERN:
  // In route setup, check feature flag before registering protected route:
  if settings.IsEnabled(tenantID, "invoice_approval") {
      finance.Post("/invoices/:id/approve",
          svc.Middleware("invoice", "approve"),
          approveInvoice,
      )
  }

  // Policy can be pre-defined (it's harmless if route doesn't exist)
  // OR policy can be added/removed when feature flag changes

SETTINGS EVENTS THAT AUTHZ CARES ABOUT:
  Feature flag disabled → optionally remove policies for that feature
                          or leave them (no route = no enforcement)
  Module disabled       → remove all policies for that module's objects
  Module enabled        → bootstrap default policies for the module
```

---

### Integration with Feature Flags

Feature flags in the platform can gate not just UI features but also authorization scopes.

```markdown
FEATURE FLAG → AUTHZ POLICY MAPPING:

Flag: "advanced_discount_approval"
  OFF: sales-manager can approve any discount
       p | role:sales-manager | {dom} | discount/* | approve | allow

  ON:  CFO approval required for discounts > 20%
       p | role:sales-manager | {dom} | discount/standard/* | approve | allow
       p | role:sales-manager | {dom} | discount/high/*     | approve | deny
       p | role:cfo           | {dom} | discount/high/*     | approve | allow

The flag change triggers:
  RemovePolicy(old rule for sales-manager)
  AddPolicy(new scoped rule for sales-manager)
  AddPolicy(new rule for cfo)
  → Effective immediately, no restart needed

IMPORTANT: authz module has NO knowledge of feature flags internally.
Feature flag logic lives in the caller (settings service, route setup code).
authz just evaluates the policies that are currently loaded.
```

---

### Integration with Entities Module

Entities (companies, branches, departments) create sub-scopes within a tenant's domain. authz models entity-level access via resource path conventions.

```markdown
ENTITY-SCOPED RESOURCES:

Resource path convention:
  entity/{entityID}/invoice/*
  entity/{entityID}/payment/*
  entity/{entityID}/report/*

Example: Branch manager at Mombasa branch (entity-id = msa-001)
  p | role:branch-manager | {tenantDom} | entity/msa-001/* | * | allow

  → Mombasa branch manager can do anything in their entity
  → CANNOT access entity/nbo-001/* (Nairobi branch)

Example: Multi-entity viewer (Regional Manager)
  p | role:regional-manager | {tenantDom} | entity/msa-001/* | read | allow
  p | role:regional-manager | {tenantDom} | entity/nbo-001/* | read | allow
  p | role:regional-manager | {tenantDom} | entity/kisumu-001/* | read | allow

REQUEST:
  Enforce(Request{
      Subject: "tenant:regional-mgr",
      Domain:  tenantDomain,
      Object:  "entity/msa-001/invoice/inv_123",
      Action:  "read",
  })
  → ALLOW (matches entity/msa-001/*)

ENTITY DELETION:
  When an entity is deleted → all policies for entity/{id}/* become
  unreachable (no route → no enforcement)
  → Clean up with RemoveFilteredPolicy(2, "entity/{id}")
    (fieldIndex=2 targets the object column)
```

---

### Integration with Audit Log Module

Every authorization decision — especially denials — feeds the audit log. The authz module itself does not write to the audit log (to keep it self-contained), but exposes the information callers need.

```markdown
WHAT CALLERS SHOULD AUDIT:

After Enforce():
  if !allowed {
      auditLog.Record(ctx, audit.Event{
          EventType: "ACCESS_DENIED",
          Subject:   request.Subject,
          Domain:    request.Domain,
          Object:    request.Object,
          Action:    request.Action,
          Reason:    "authz deny",
          Timestamp: time.Now(),
      })
  }

After AssignRole():
  auditLog.Record(ctx, audit.Event{
      EventType:   "ROLE_ASSIGNED",
      Subject:     subject,
      Role:        role,
      Domain:      domain,
      AssignedBy:  opts.assignedBy,
      ExpiresAt:   opts.expiresAt,
      Timestamp:   time.Now(),
  })

After RevokeRole():
  auditLog.Record(ctx, audit.Event{
      EventType: "ROLE_REVOKED",
      Subject:   subject,
      Role:      role,
      Domain:    domain,
  })

NOTE: role_assignments table ALREADY provides a partial audit trail.
The audit log adds context (IP, request ID, correlation ID) that
the role_assignments table doesn't have.

COMPLIANCE: For SOC 2 / ISO 27001, every ACCESS_DENIED must be recorded.
The middleware's 403 response is the right place to emit this event.
```

---

### Integration with Analytics / Usage Module

The tenant usage tracking module can consume authz events for security dashboards.

```markdown
METRICS EMITTED BY authz MODULE (when Metrics configured):

  authz_enforce_total{result="allow",domain="..."} counter
  authz_enforce_total{result="deny",domain="..."}  counter
  authz_enforce_duration_ms histogram
  authz_role_assignments_active{domain="..."} gauge
  authz_expired_roles_revoked_total counter

USEFUL PLATFORM METRICS:
  "Which tenants have the most denied requests?"
  "Which roles are most frequently used?"
  "Are there domains with unusual deny rates? (possible attack)"
  "Which API clients are being rate-limited by policy?"
```

---

Next: [Business Module Integration](./13-business-module-integration.md)
