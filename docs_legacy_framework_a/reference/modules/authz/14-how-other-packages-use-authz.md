> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

[<-- Back to Index](README.md)

## How Other Packages Use authz

### The One Import Pattern

Every consumer of the authz module depends on the `Service` interface only — never on the concrete `*service` struct. This keeps coupling minimal and makes mocking trivial in tests.

```go
// In any module that needs authorization:
import "awo.so/internal/core/authz"

type InvoiceService struct {
    repo  InvoiceRepository
    authz authz.Service    // ← interface only
    log   logger.Logger
}
```

### 0. Session Service — Builds Permissions via authz at Login

`internal/core/identity/session/service.go` is the most important consumer. At login it uses `authz.GetPolicies` + `authz.GetRoles` to compute the permission map stored in `user_sessions.permissions` JSONB. This is the **only time** Casbin is hit on the request path.

```go
// internal/core/identity/session/service.go

func (s *service) buildPermissions(ctx context.Context, user *identity.User) (map[string]bool, error) {
    ctx, span := s.tracer.StartSpan(ctx, "session.buildPermissions")
    defer span.End()

    domain := authz.TenantDomain(user.TenantID.String())
    subject := authz.TenantSubject(user.ID.String())

    // Get all g-rules for this user in this domain (from Casbin in-memory)
    roles, err := s.authz.GetRoles(ctx, subject, domain)
    if err != nil {
        return nil, fmt.Errorf("session: get roles: %w", err)
    }

    // Get all p-rules for this domain
    policies, err := s.authz.GetPolicies(ctx, domain)
    if err != nil {
        return nil, fmt.Errorf("session: get policies: %w", err)
    }

    roleSet := make(map[string]bool, len(roles))
    for _, r := range roles {
        roleSet[r] = true
    }

    // Build {module}.{resource}.{action} permission map
    perms := make(map[string]bool)
    for _, p := range policies {
        if roleSet[p.Subject] && p.Effect == "allow" {
            // object already is "finance.receivables.invoices", action is "read"
            perms[p.Object+"."+p.Action] = true
        }
    }

    s.metrics.RecordCount("session.permissions_computed", float64(len(perms)), nil)
    return perms, nil
}
```

### 1. API Gateway / Route Layer

The API gateway is the outermost consumer. It registers routes and attaches middleware. This is the most common usage pattern.

```go
// cmd/api/routes/finance.go

func RegisterFinanceRoutes(app *fiber.App, svc authz.Service, handler *FinanceHandler) {
    g := app.Group("/api/v1/finance", authnMiddleware)

    // READ operations — anyone with finance-viewer or higher
    g.Get("/invoices",               svc.Middleware("invoice", "read"),   handler.ListInvoices)
    g.Get("/invoices/:id",           svc.Middleware("invoice", "read"),   handler.GetInvoice)
    g.Get("/invoices/:id/pdf",       svc.Middleware("invoice", "export"), handler.ExportInvoicePDF)

    // WRITE operations — finance-manager or higher
    g.Post("/invoices",              svc.Middleware("invoice", "create"), handler.CreateInvoice)
    g.Put("/invoices/:id",           svc.Middleware("invoice", "update"), handler.UpdateInvoice)

    // SENSITIVE operations — cfo or higher
    g.Post("/invoices/:id/approve",  svc.Middleware("invoice", "approve"),handler.ApproveInvoice)
    g.Post("/invoices/:id/post",     svc.Middleware("invoice", "post"),   handler.PostInvoice)
    g.Delete("/invoices/:id",        svc.Middleware("invoice", "delete"), handler.DeleteInvoice)

    // PERIOD closing — restricted
    g.Post("/periods/:id/close",     svc.Middleware("period", "close"),   handler.ClosePeriod)
}
```

### 2. Service Layer (Non-HTTP Operations)

Services that perform cross-module operations call `Enforce` directly, not via middleware.

```go
// internal/core/finance/invoice_service.go

func (s *InvoiceService) BulkApproveInvoices(ctx context.Context, invoiceIDs []string, actor authz.Principal) error {
    // Pre-check all permissions in one batch call
    reqs := make([]authz.Request, len(invoiceIDs))
    for i, id := range invoiceIDs {
        reqs[i] = authz.Request{
            Subject: actor.Subject,
            Domain:  actor.Domain,
            Object:  "invoice/" + id,
            Action:  "approve",
        }
    }

    results, err := s.authz.EnforceBatch(ctx, reqs)
    if err != nil {
        return fmt.Errorf("permission check failed: %w", err)
    }

    var denied []string
    for i, allowed := range results {
        if !allowed {
            denied = append(denied, invoiceIDs[i])
        }
    }
    if len(denied) > 0 {
        return fmt.Errorf("not authorized to approve invoices: %v", denied)
    }

    // All authorized — proceed with bulk approve
    return s.repo.BulkApprove(ctx, invoiceIDs)
}
```

### 3. Report Export Service

Report generation is async (can take minutes). Authorization must be checked at job submission, not at result retrieval.

```go
// internal/core/reporting/export_service.go

func (s *ExportService) SubmitExportJob(ctx context.Context, actor authz.Principal, reportType string) (string, error) {
    // Check authorization at job submission time
    ok, err := s.authz.Enforce(ctx, authz.Request{
        Subject: actor.Subject,
        Domain:  actor.Domain,
        Object:  "report/finance/" + reportType,
        Action:  "export",
    })
    if err != nil {
        return "", err
    }
    if !ok {
        return "", authz.ErrForbidden
    }

    // Submit to Temporal workflow
    jobID := uuid.New().String()
    return jobID, s.temporal.ExecuteWorkflow(ctx, reportExportWorkflow, reportType, actor)
}
```

### 4. Platform Administration API

Platform-level operations use `_platform_` domain. The platform admin routes have their own authn middleware that sets `DomainPlatform` in the Principal.

```go
// cmd/api/routes/platform.go

func RegisterPlatformRoutes(app *fiber.App, svc authz.Service, handler *PlatformHandler) {
    g := app.Group("/platform", platformAuthnMiddleware)
    // platformAuthnMiddleware sets:
    //   Principal{Subject: "platform:admin", Domain: authz.DomainPlatform}

    g.Get("/tenants",          svc.Middleware("tenant", "read"),   handler.ListTenants)
    g.Post("/tenants/:id/suspend", svc.Middleware("tenant", "suspend"), handler.SuspendTenant)
    g.Post("/policies/import", svc.Middleware("policy", "import"), handler.BulkImportPolicies)
}
```

### 5. Tenant Admin UI API

Tenant admins manage their own users and roles through a dedicated API:

```go
// cmd/api/routes/tenant_admin.go

func RegisterTenantAdminRoutes(app *fiber.App, svc authz.Service, handler *TenantAdminHandler) {
    g := app.Group("/api/v1/admin", tenantAuthnMiddleware)

    // Role assignment management
    g.Get("/users/:id/roles",   svc.Middleware("user/role", "read"),   handler.GetUserRoles)
    g.Post("/users/:id/roles",  svc.Middleware("user/role", "assign"), handler.AssignRole)
    g.Delete("/users/:id/roles/:role", svc.Middleware("user/role", "revoke"), handler.RevokeRole)

    // Policy management (only tenant-admin can do this)
    g.Get("/policies",   svc.Middleware("policy", "read"),   handler.GetPolicies)
    g.Post("/policies",  svc.Middleware("policy", "create"), handler.AddPolicy)
    g.Delete("/policies",svc.Middleware("policy", "delete"), handler.RemovePolicy)

    // Assignment list (audit view)
    g.Get("/assignments", svc.Middleware("assignment", "read"), handler.GetAssignments)
}

// Handler for assigning roles via API:
func (h *TenantAdminHandler) AssignRole(c *fiber.Ctx) error {
    principal := c.Locals(authz.LocalsKeyPrincipal).(authz.Principal)
    tenantID   := c.Locals("tenantID").(string)

    var body struct {
        Subject string     `json:"subject"`
        Role    string     `json:"role"`
        Expiry  *time.Time `json:"expires_at"`
    }
    if err := c.BodyParser(&body); err != nil {
        return fiber.NewError(400, "invalid body")
    }

    opts := []authz.AssignOpt{authz.WithAssignedBy(principal.Subject)}
    if body.Expiry != nil {
        opts = append(opts, authz.WithExpiry(*body.Expiry))
    }

    return h.authz.AssignRole(c.Context(), tenantID, body.Subject, body.Role,
        authz.TenantDomain(tenantID), opts...)
}
```

### 6. Dependency Injection (Wire)

The authz service is constructed once at startup and injected everywhere:

```go
// internal/di/providers.go

func ProvideAuthzService(pool *pgxpool.Pool, log logger.Logger, metrics metrics.MetricsProvider) (authz.Service, error) {
    return authz.New(authz.Config{
        Pool:    pool,
        Logger:  log,
        Metrics: metrics,
    })
}

// Wire binding — authz.Service is a singleton shared by all modules
var AuthzSet = wire.NewSet(
    ProvideAuthzService,
    wire.Bind(new(authz.Service), new(*authz.service)), // NOT needed — New returns interface
)
```

### 7. Testing (Mocking the Service)

Because consumers depend on the `authz.Service` interface, tests can use any mock:

```go
// Using a simple allow-all mock for unit tests:
type alwaysAllowAuthz struct{}
func (a *alwaysAllowAuthz) Enforce(_ context.Context, _ authz.Request) (bool, error) { return true, nil }
func (a *alwaysAllowAuthz) EnforceBatch(_ context.Context, reqs []authz.Request) ([]bool, error) {
    results := make([]bool, len(reqs)); for i := range results { results[i] = true }; return results, nil
}
// ... implement remaining interface methods as stubs

// In test:
svc := &InvoiceService{repo: mockRepo, authz: &alwaysAllowAuthz{}}

// Using go.uber.org/mock:
ctrl := gomock.NewController(t)
mockAuthz := mocks.NewMockService(ctrl)
mockAuthz.EXPECT().Enforce(gomock.Any(), authz.Request{
    Subject: "tenant:usr_001", Domain: "tenant-abc",
    Object: "invoice/inv_123", Action: "read",
}).Return(true, nil)
```

---

Next: [Workflow Integration](./15-workflow-integration.md)
