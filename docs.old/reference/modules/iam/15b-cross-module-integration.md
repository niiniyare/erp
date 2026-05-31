[<-- Back to Index](README.md)

## Cross-Module Integration

### The Integration Contract

Every business module uses these mechanisms to interact with IAM and configuration. Modules **never** call `AccessService`, `FlagService`, or `SettingService` directly in handlers.

| Mechanism | Where | Purpose |
|---|---|---|
| `middleware.RequirePermission(resource, action)` | Route | Blocks before handler runs |
| `middleware.RequireFlag(flagKey)` | Route | Feature gate before handler runs |
| `session.CanDo(resource, action)` or `session.Can("resource.action")` | Handler | Conditional UI logic (show/hide buttons) |
| `session.FeatureEnabled(flagKey)` | Handler | Optional form sections |
| `session.SettingDecimal/Int/Bool/String(key, default)` | Handler | Tenant configuration values |

---

### Finance Module Integration

```go
// Route: permission + flag both required
api.Post("/transactions/:id/post",
    middleware.RequireFlag("finance.transactions"),
    middleware.RequirePermission("finance.transactions", "post"),
    handlers.PostTransaction(deps))

// Schema handler: all configuration from session, zero DB hits
func TransactionDetailSchema(deps *app.Deps) fiber.Handler {
    return func(c *fiber.Ctx) error {
        s := middleware.ContextSession(c)

        showApproval := s.FeatureEnabled("finance.transactions.approval_workflow")
        threshold    := s.SettingDecimal("finance.transactions.approval_threshold", decimal.Zero)

        return c.JSON(buildTransactionDetail(TransactionDetailConfig{
            ShowApprovalSection: showApproval,
            ApprovalThreshold:   threshold,
            CanPost:    s.CanDo("finance.transactions", "post"),
            CanApprove: s.CanDo("finance.transactions", "approve"),
            CanVoid:    s.CanDo("finance.transactions", "void"),
            CanReverse: s.CanDo("finance.transactions", "reverse"),
        }))
    }
}
```

---

### Portal Module Integration

```go
api.Get("/portal/invoices",
    middleware.RequireFlag("portal.self.invoices"),
    middleware.RequirePermission("portal.self.invoices", "read"),
    handlers.GetMyInvoices(deps))

func GetMyInvoices(deps *app.Deps) fiber.Handler {
    return func(c *fiber.Ctx) error {
        session := middleware.ContextSession(c)
        // PrincipalID from session — NEVER from request params
        invoices, _ := deps.InvoiceSvc.GetByCustomer(c.Context(),
            domain.GetInvoicesParams{CustomerID: session.PrincipalID})
        return c.JSON(response.OK(invoices))
    }
}
// Two independent protections:
// 1. Handler reads PrincipalID from session, not query params
// 2. DB RLS portal_self_isolation rejects customer_id != session.principal_id
```

---

### Tenant Lifecycle Events

IAM responds to tenant lifecycle changes via domain events:

```go
func (s *IAMService) OnTenantProvisioned(ctx context.Context, e domain.TenantProvisioned) error {
    return s.seedSystemRoles(ctx, e.TenantID)
    // Flag definitions were already auto-seeded by DB triggers when modules were created
    // Tenant starts with all default flag values; admin configures from there
}

func (s *IAMService) OnTenantSuspended(ctx context.Context, e domain.TenantSuspended) error {
    // Immediately invalidate all sessions
    s.sessionRepo.InvalidateByTenant(ctx, e.TenantID)
    // Add blanket deny in Casbin — even if a session survives, it gets blocked
    s.casbin.AddPolicy(ctx, domain.PolicyParams{
        Subject: "tenant:*", Domain: e.TenantID.String(),
        Object: "*", Action: "*", Effect: "deny",
    })
    return nil
}
```

---

Next: [Audit Trail](./16b-audit-trail.md)
