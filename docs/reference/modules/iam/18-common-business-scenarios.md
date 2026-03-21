[<-- Back to Index](README.md)

## Common Business Scenarios

### Scenario 1: New Company Onboards — Bootstrapping Authorization

```markdown
SITUATION:
  Savannah Electronics Ltd (Kenya) signs up for AWO ERP.
  Platform admin provisions the tenant and needs to set up authorization.

STEP 1: Tenant provisioning (Tenant Module)
  tenantID := "a1b2c3d4-5678-..."
  tenantDomain := authz.TenantDomain(tenantID)

STEP 2: Bootstrap default role policies (one-time setup)
  // Finance roles
  svc.AddPolicy(ctx, authz.Policy{Subject:"role:cfo",              Domain:tenantDomain, Object:"invoice/*",  Action:"*",      Effect:"allow"})
  svc.AddPolicy(ctx, authz.Policy{Subject:"role:finance-manager",  Domain:tenantDomain, Object:"invoice/*",  Action:"*",      Effect:"allow"})
  svc.AddPolicy(ctx, authz.Policy{Subject:"role:finance-viewer",   Domain:tenantDomain, Object:"invoice/*",  Action:"read",   Effect:"allow"})
  svc.AddPolicy(ctx, authz.Policy{Subject:"role:finance-manager",  Domain:tenantDomain, Object:"payment/*",  Action:"*",      Effect:"allow"})
  svc.AddPolicy(ctx, authz.Policy{Subject:"role:finance-viewer",   Domain:tenantDomain, Object:"payment/*",  Action:"read",   Effect:"allow"})

  // Sales roles
  svc.AddPolicy(ctx, authz.Policy{Subject:"role:sales-manager",    Domain:tenantDomain, Object:"order/*",    Action:"*",      Effect:"allow"})
  svc.AddPolicy(ctx, authz.Policy{Subject:"role:sales-rep",        Domain:tenantDomain, Object:"order/*",    Action:"create", Effect:"allow"})
  svc.AddPolicy(ctx, authz.Policy{Subject:"role:sales-rep",        Domain:tenantDomain, Object:"order/*",    Action:"read",   Effect:"allow"})

  // Full admin
  svc.AddPolicy(ctx, authz.Policy{Subject:"role:tenant-admin",     Domain:tenantDomain, Object:"*",          Action:"*",      Effect:"allow"})

STEP 3: Assign first admin user
  adminUserID := "usr_ceo_001"
  svc.AssignRole(ctx, tenantID,
      authz.TenantSubject(adminUserID),
      "role:tenant-admin",
      tenantDomain,
      authz.WithAssignedBy("platform:provisioning-svc"),
  )

RESULT:
  CEO (usr_ceo_001) can do everything.
  CEO then uses the admin UI to assign roles to their team.
```

---

### Scenario 2: Hiring and Onboarding a New Employee

```markdown
SITUATION:
  Savannah Electronics hires Amina Hassan as Finance Manager.
  IT/HR assigns her ERP role.

HR system triggers role assignment:
  svc.AssignRole(ctx, tenantID,
      "tenant:usr_amina_hassan",
      "role:finance-manager",
      tenantDomain,
      authz.WithAssignedBy("tenant:usr_ceo_001"),
  )

From next login: Amina has full finance access.
  Enforce(Request{Subject:"tenant:usr_amina_hassan", Domain:tenantDomain,
      Object:"invoice/123", Action:"read"}) → ALLOW
  Enforce(..., Object:"invoice/123", Action:"approve") → ALLOW
  Enforce(..., Object:"hr/payroll/123", Action:"read") → DENY (no HR policy)
```

---

### Scenario 3: Employee Termination — Zero Tolerance for Stale Access

```markdown
SITUATION:
  James Ochieng (Sales Manager) is terminated immediately.
  Access must be revoked within minutes.

IMMEDIATE REVOCATION:
  // Step 1: Revoke all active roles
  assignments, _ := svc.GetAssignments(ctx, "tenant:usr_james", tenantDomain)
  for _, a := range assignments {
      if a.IsActive {
          svc.RevokeRole(ctx, "tenant:usr_james", a.Role, tenantDomain)
      }
  }

  // Step 2: Add blanket deny (defense in depth)
  svc.AddPolicy(ctx, authz.Policy{
      Subject: "tenant:usr_james",
      Domain:  tenantDomain,
      Object:  "*",
      Action:  "*",
      Effect:  "deny",
  })

  // Step 3: Invalidate JWT (handled by authn module, not authz)
  // JWT blacklist or session revocation

VERIFICATION:
  Enforce(Request{Subject:"tenant:usr_james", Domain:tenantDomain,
      Object:"order/123", Action:"read"}) → DENY ✓
  Enforce(Request{Subject:"tenant:usr_james", Domain:tenantDomain,
      Object:"*", Action:"*"}) → DENY ✓

AUDIT TRAIL:
  SELECT * FROM role_assignments
  WHERE subject = 'tenant:usr_james' ORDER BY created_at DESC;
  → Shows: role:sales-manager was_active=FALSE (revoked), when, by whom
```

---

### Scenario 4: External Auditors — Time-Limited Read Access

```markdown
SITUATION:
  Nairobi branch of PwC will audit Savannah Electronics' FY2025 accounts.
  Access needed: Jan 15 – Feb 28, 2026. Read-only. No export.

SETUP (Jan 15):
  expiry := time.Date(2026, 2, 28, 23, 59, 59, 0, time.UTC)

  // Create auditor accounts via normal user provisioning
  // Assign time-limited role
  for _, auditorID := range []string{"usr_pwc_1", "usr_pwc_2", "usr_pwc_3"} {
      svc.AssignRole(ctx, tenantID,
          authz.TenantSubject(auditorID),
          "role:auditor",
          tenantDomain,
          authz.WithExpiry(expiry),
          authz.WithAssignedBy("tenant:usr_cfo"),
          authz.WithDelegatedBy("tenant:usr_ceo_001"),
      )
  }

  // Block export (auditors should not export data)
  svc.AddPolicy(ctx, authz.Policy{
      Subject: "role:auditor",
      Domain:  tenantDomain,
      Object:  "*/export",
      Action:  "*",
      Effect:  "deny",
  })

DURING AUDIT:
  Enforce(Request{Subject:"tenant:usr_pwc_1", Domain:tenantDomain,
      Object:"invoice/inv_2025_001", Action:"read"}) → ALLOW ✓
  Enforce(Request{Subject:"tenant:usr_pwc_1", Domain:tenantDomain,
      Object:"invoice/inv_2025_001/export", Action:"*"}) → DENY ✓ (deny rule)

AFTER EXPIRY (Mar 1):
  Next request from usr_pwc_1:
  → revokeExpiredRoles() fires → role:auditor revoked
  Enforce(...) → DENY ✓ (automatic, no manual cleanup needed)
```

---

### Scenario 5: API Integration with Accounting Software (QuickBooks Sync)

```markdown
SITUATION:
  Savannah Electronics integrates with QuickBooks.
  The integration needs to read invoices and post payments.
  It must NEVER delete or modify existing data.

SETUP (by tenant admin):
  clientID := "cli_quickbooks_sync"
  apiDomain := authz.APIDomain(tenantID)

  // Grant read access to invoices
  svc.AddPolicy(ctx, authz.Policy{
      Subject: "api:" + clientID,
      Domain:  apiDomain,
      Object:  "invoice/*",
      Action:  "read",
      Effect:  "allow",
  })

  // Grant create access to payments only
  svc.AddPolicy(ctx, authz.Policy{
      Subject: "api:" + clientID,
      Domain:  apiDomain,
      Object:  "payment/*",
      Action:  "create",
      Effect:  "allow",
  })

  // Explicit deny on all destructive operations
  svc.AddPolicy(ctx, authz.Policy{Subject:"api:"+clientID, Domain:apiDomain, Object:"*", Action:"delete", Effect:"deny"})
  svc.AddPolicy(ctx, authz.Policy{Subject:"api:"+clientID, Domain:apiDomain, Object:"*", Action:"update", Effect:"deny"})

  // Assign to API role
  svc.AssignRole(ctx, tenantID,
      authz.APISubject(clientID),
      "role:api-accounting-sync",
      apiDomain,
  )

VERIFICATION:
  Enforce(Request{Subject:"api:cli_quickbooks_sync", Domain:apiDomain,
      Object:"invoice/inv_2026_001", Action:"read"}) → ALLOW ✓
  Enforce(Request{Subject:"api:cli_quickbooks_sync", Domain:apiDomain,
      Object:"invoice/inv_2026_001", Action:"delete"}) → DENY ✓
```

---

### Scenario 6: Customer Portal — Self-Service Invoice Access

```markdown
SITUATION:
  Nairobi Electronics Ltd is a customer of Savannah Electronics.
  They need to log in to the portal to view their invoices and pay.

SETUP (when customer account created):
  customerUserID := "cust_nairobi_001"
  portalDomain := authz.PortalDomain(tenantID)

  svc.AssignRole(ctx, tenantID,
      authz.PortalSubject(customerUserID),
      "role:portal-customer",
      portalDomain,
  )

  // Portal customer policy (set up during tenant bootstrap)
  // p | role:portal-customer | {tenantID}:portal | invoice/customer/{customerID}/* | read | allow
  // p | role:portal-customer | {tenantID}:portal | payment/customer/{customerID}/* | create | allow

PORTAL REQUEST:
  Enforce(Request{
      Subject: "portal:cust_nairobi_001",
      Domain:  portalDomain,
      Object:  "invoice/customer/nairobi_001/inv_2026_001",
      Action:  "read",
  }) → ALLOW ✓

  Enforce(Request{
      Subject: "portal:cust_nairobi_001",
      Domain:  portalDomain,
      Object:  "invoice/customer/mombasa_parts_001/inv_2026_999",  // ANOTHER customer
      Action:  "read",
  }) → DENY ✓ (customer ID doesn't match their policy)
```

---

### Scenario 7: Finance Period Closing — Restricted Operation

```markdown
SITUATION:
  Savannah Electronics closes Q1 2026.
  Only CFO can close periods. Once closed, no modifications allowed.

POLICY SETUP:
  svc.AddPolicy(ctx, authz.Policy{
      Subject: "role:cfo",
      Domain:  tenantDomain,
      Object:  "period/*",
      Action:  "close",
      Effect:  "allow",
  })
  // Deny all writes to closed periods (enforced by resource path in handler)
  svc.AddPolicy(ctx, authz.Policy{
      Subject: "*",
      Domain:  tenantDomain,
      Object:  "journal/closed/*",
      Action:  "create",
      Effect:  "deny",
  })
  svc.AddPolicy(ctx, authz.Policy{
      Subject: "*",
      Domain:  tenantDomain,
      Object:  "journal/closed/*",
      Action:  "update",
      Effect:  "deny",
  })

CFO closes period:
  Enforce({Subject:"tenant:usr_cfo", Domain:dom, Object:"period/q1-2026", Action:"close"})
  → ALLOW ✓

Finance manager tries to post to closed period:
  Enforce({Subject:"tenant:usr_fin_mgr", Domain:dom, Object:"journal/closed/q1-2026/je_001", Action:"create"})
  → DENY ✓ (deny rule for journal/closed/*)
```

---

### Scenario 8: New Tenant — Module Configuration

```
1. Tenant is provisioned → OnTenantProvisioned seeds system roles
2. Platform operator enables Finance module:
   FlagService.Set({ TenantID: tenantID, FlagKey: "finance", Enabled: true })
   → Invalidates all tenant sessions (users must re-login to see Finance)
3. Platform operator configures settings:
   SettingService.UpdateModuleSettings({
     TenantID: tenantID,
     Settings: map[string]string{
       "finance.transactions.approval_threshold": "500000",
       "finance.budget_control_mode": "soft",
     }
   })
4. Next tenant admin login → session built with finance=true in flags
   → Finance appears in nav automatically
```

### Scenario 9: Enabling a Feature Mid-Subscription

```
Tenant admin navigates to Settings → Modules
Sees "Airline Bookings" as a toggle (off)
Toggles it on:
  PATCH /api/v1/settings/flags/airline  { "enabled": true }
  → FlagService.Set fires
  → All tenant sessions invalidated (users must refresh)
  → Next login for any tenant user → session built with airline=true
  → If user has airline.*.read permission → Airline nav section appears
  → If user lacks permission → still no nav item (permission filter)
```

### Scenario 10: Employee Termination (Immediate Access Revocation)

```go
func (s *HRService) TerminateEmployee(ctx context.Context, params domain.TerminateParams) error {
    s.repo.SetTerminated(ctx, params.EmployeeID, params.Reason)
    platform.IAM.SuspendUser(ctx, domain.SuspendParams{
        UserID: params.UserID, Reason: "termination",
    }) // → InvalidateByUser() called inside; all sessions deleted immediately
    platform.IAM.AddDenyAll(ctx, domain.DenyAllParams{
        UserID: params.UserID, Reason: params.Reason,
    }) // → Casbin deny policy; even a surviving session gets blocked
    return nil
}
```

### Scenario 11: Regional Manager (Branch-Scoped Access)

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

---

Next: [Troubleshooting Guide](./19-troubleshooting-guide.md)
