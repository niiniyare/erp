[<-- Back to Index](README.md)

> **⚠️ PARTIALLY OUTDATED.** The schema route section below shows the **old approach** (per-route
> handler functions). The current system uses a **single `SchemaHandler`** that dispatches via the
> page registry and runs a 9-stage pipeline. Key differences:
>
> | Aspect | Old (in this doc) | Current (in code) |
> |--------|------------------|-------------------|
> | Schema routes | One handler per page | Single handler + registry lookup |
> | Permission check (schema) | `RequirePermission` middleware | `AuthzStage` (priority 20) inside pipeline |
> | Page registration | Per-route in `RegisterRoutes()` | `registry.RegisterPage()` in `init()` |
> | Handler signature | `func(c *fiber.Ctx) error` per page | `func(sess UISessionContext) any` |
>
> `RequirePermission` middleware is **correct** for `/api/v1/` data routes — do not remove it there.
> For schema routes, permission resolution happens automatically in the pipeline.
>
> See [Pipeline Deep Dive](../02-architecture/02-pipeline-deep-dive.md) and
> [Page Registration Pattern](../03-implementation/02-page-registration-pattern.md) for current approach.

## Routes & Auth

### Route Groups in Fiber

```go
func RegisterRoutes(app *fiber.App, s *Services) {
    // Static files (shell + AMIS SDK)
    app.Static("/sdk",    "./web/sdk")
    app.Static("/utils",  "./web/utils")
    app.Static("/public", "./web/public")

    // Schema files (static JSON, no auth — for now)
    app.Static("/schemas", "./web/schemas")

    // Shell — catch-all for SPA
    app.Get("/*", func(c *fiber.Ctx) error {
        if strings.HasPrefix(c.Path(), "/api/") ||
           strings.HasPrefix(c.Path(), "/schema/") {
            return c.Next()
        }
        return c.SendFile("./web/index.html")
    })

    // Data API routes (authenticated)
    api := app.Group("/api/v1",
        middleware.ResolveTenant,
        middleware.Authenticate,
        middleware.InjectFlags,
        middleware.AuditWrap,
    )

    // Purchase Orders
    po := api.Group("/purchase-orders",
        middleware.RequirePermission("read", "purchase_orders", s.Access))
    po.Get("/",              s.POHandlers.List)
    po.Post("/",             middleware.RequirePermission("create", "purchase_orders", s.Access),
                             s.POHandlers.Create)
    po.Get("/:id",           s.POHandlers.Get)
    po.Put("/:id",           middleware.RequirePermission("update", "purchase_orders", s.Access),
                             s.POHandlers.Update)
    po.Post("/:id/submit",   s.POHandlers.Submit)
    po.Post("/:id/cancel",   s.POHandlers.Cancel)
    po.Get("/:id/lines",     s.POHandlers.ListLines)
    po.Get("/:id/approvals", s.POHandlers.ListApprovals)
    po.Post("/export",       s.POHandlers.Export)
    // ... other resources

    // Dynamic Schema routes (authenticated, future)
    schemaRoutes := app.Group("/schema",
        middleware.ResolveTenant,
        middleware.Authenticate,
        middleware.InjectFlags,
    )
    schemaRoutes.Get("/app",                    s.SchemaHandlers.AppShell)
    schemaRoutes.Get("/purchasing/orders",      s.SchemaHandlers.PurchasingOrders)
    schemaRoutes.Get("/purchasing/orders/new",  s.SchemaHandlers.PurchasingOrdersNew)
    schemaRoutes.Get("/purchasing/orders/:id",  s.SchemaHandlers.PurchasingOrderDetail)
    schemaRoutes.Get("/settings/rules/:module", s.SchemaHandlers.SettingsRules)
    // ...
}
```

### Authentication Flow

The AMIS fetcher sends cookies automatically (same-origin requests). The `Authenticate` middleware reads the session cookie on every `/api/v1/` and `/schema/` request:

```go
func Authenticate(c *fiber.Ctx) error {
    sessionToken := c.Cookies("awo_session")
    if sessionToken == "" {
        return c.Status(401).JSON(response.Err("Session expired. Please log in."))
    }
    session, err := authSvc.ValidateSession(c.Context(), sessionToken)
    if err != nil {
        return c.Status(401).JSON(response.Err("Invalid session"))
    }
    setContextSession(c, session)
    return c.Next()
}
```

When AMIS receives `status: 1` from any fetch, it shows `msg` as an error toast. Handle 401 in the fetcher to redirect to login:

```javascript
// In index.html amisEnv.fetcher
return fetch(url, opts).then(function(res) {
  if (res.status === 401) {
    window.location.href = '/login';
    return;
  }
  return res.json();
});
```

### Permission Gating in Schemas

Inject permissions as page-level data, then use `visibleOn` / `disabledOn`:

```go
// Schema handler
func (h *SchemaHandlers) PurchasingOrders(c *fiber.Ctx) error {
    fl   := middleware.ContextFlags(c)
    user := middleware.ContextUser(c)

    page := schema.Page{
        Type: "page",
        Data: map[string]any{
            "can_create":  h.access.Can(c.Context(), user.ID, "create",  "purchase_orders"),
            "can_approve": h.access.Can(c.Context(), user.ID, "approve", "purchase_orders"),
        },
        Toolbar: []any{
            schema.Button{
                Label: "New Order", Level: "primary",
                ActionType: "link", Link: "/purchasing/orders/new",
                VisibleOn: "${can_create}",
            },
        },
        // ...
    }
    return c.JSON(page)
}
```

### Feature-Flag Gating

If a user navigates directly to a disabled module's URL, return an informational page (defence in depth — the nav item is already hidden):

```go
func (h *SchemaHandlers) PeoplePayroll(c *fiber.Ctx) error {
    fl := middleware.ContextFlags(c)
    if !fl["payroll.module"] {
        return c.JSON(map[string]any{
            "type":  "page",
            "title": "Payroll",
            "body": map[string]any{
                "type":  "alert",
                "level": "info",
                "body":  "The Payroll module is not enabled. Contact your administrator.",
            },
        })
    }
    return c.JSON(payrollschema.RunsPage(fl))
}
```

### Middleware Stack Order

```markdown
Request
  │
  ├─ ResolveTenant     → reads X-Tenant or subdomain → sets ctx.TenantID
  ├─ Authenticate      → validates session → sets ctx.UserID, ctx.User
  ├─ InjectFlags       → loads feature flags for tenant → sets ctx.Flags
  ├─ RequirePermission → checks specific permission → 403 if denied
  └─ AuditWrap         → logs mutation requests (POST/PUT/DELETE) to audit log
```

---
