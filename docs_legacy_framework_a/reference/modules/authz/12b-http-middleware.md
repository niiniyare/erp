> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

[<-- Back to Index](README.md)

## HTTP Middleware Chain & API Layer

### Router Structure

```go
func BuildRouter(app *fiber.App, deps *Deps) {
    // Infrastructure — all requests
    app.Use(middleware.Recovery(), middleware.Logger(), middleware.RateLimit(),
            middleware.CORS(), middleware.SecurityHeaders())

    // Public — no auth
    app.Post("/auth/login",                   handlers.Login(deps))
    app.Post("/auth/logout",                  handlers.Logout(deps))
    app.Post("/auth/forgot-password",         handlers.ForgotPassword(deps))
    app.Post("/auth/reset-password",          handlers.ResetPassword(deps))
    app.Get("/auth/oauth/:provider",          handlers.OAuthRedirect(deps))
    app.Get("/auth/oauth/:provider/callback", handlers.OAuthCallback(deps))
    app.Get("/schema/boot",                   handlers.Boot(deps))
    app.Static("/static", "./web/static")

    // Authenticated — all routes below require valid session
    auth := app.Group("",
        middleware.Authenticate(deps.Platform.IAM),
        middleware.SetDBPool(deps.Pools),
        middleware.ResolveTenant(deps.Platform.Tenant),
        middleware.AuditWrap(deps.Platform.Audit),
    )

    // Schema routes — build UI schema for amis frontend
    sg := auth.Group("/schema")
    sg.Get("/accounting/transactions",
        middleware.RequireFlag("finance.transactions"),
        middleware.RequirePermission("finance.transactions", "read"),
        handlers.TransactionListSchema(deps))
    sg.Get("/settings/modules",
        middleware.RequirePermission("settings.modules", "read"),
        handlers.ModuleFlagsSchema(deps))
    sg.Get("/settings/finance",
        middleware.RequirePermission("settings.finance", "read"),
        handlers.FinanceSettingsSchema(deps))

    // Data API routes
    api := auth.Group("/api/v1")
    api.Get("/transactions",
        middleware.RequireFlag("finance.transactions"),
        middleware.RequirePermission("finance.transactions", "read"),
        handlers.ListTransactions(deps))
    api.Patch("/settings/flags/:key",
        middleware.RequirePermission("settings.modules", "update"),
        handlers.SetFlag(deps))
    api.Patch("/settings/:module",
        middleware.RequirePermission("settings.finance", "update"),
        handlers.UpdateModuleSettings(deps))
}
```

---

### Middleware Implementations

```go
func RequirePermission(resource, action string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        if !ContextSession(c).Can(resource, action) {
            return c.Status(403).JSON(response.Err(
                fmt.Sprintf("permission denied: %s.%s", resource, action)))
        }
        return c.Next()
    }
}

func RequireFlag(flagKey string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        if !ContextSession(c).FeatureEnabled(flagKey) {
            return c.Status(403).JSON(response.Err(
                "this feature is not enabled for your organisation"))
        }
        return c.Next()
    }
}
```

Both read from the pre-computed session — zero DB hits.

---

### Context Helpers

Used in every handler — never pass tenant_id or user_id as function parameters:

```go
func ContextSession(c *fiber.Ctx) *domain.ResolvedSession  { return c.Locals("session").(*domain.ResolvedSession) }
func ContextTenantID(c *fiber.Ctx) uuid.UUID               { return ContextSession(c).TenantID }
func ContextUserID(c *fiber.Ctx) uuid.UUID                 { return ContextSession(c).UserID }
func ContextEntityID(c *fiber.Ctx) uuid.UUID               { return ContextSession(c).EntityID }
func ContextPrincipalID(c *fiber.Ctx) uuid.UUID            { return ContextSession(c).PrincipalID }
```

---

### SetDBPool Middleware

Selects the appropriate DB connection pool based on user type and sets per-connection RLS context:

```go
func SetDBPool(pools *db.Pools) fiber.Handler {
    return func(c *fiber.Ctx) error {
        session := ContextSession(c)
        if session.IsPlatform() {
            // Platform users: admin_role pool, BYPASSRLS
            c.Locals("db", pools.Platform)
        } else {
            // All other users: application_role pool, RLS active
            conn, _ := pools.App.Acquire(c.Context())
            conn.Exec(c.Context(), `
                SELECT set_config('app.current_tenant_id', $1, true),
                       set_config('app.user_id',           $2, true),
                       set_config('app.user_type',         $3, true),
                       set_config('app.principal_id',      $4, true),
                       set_config('app.entity_id',         $5, true)
            `, session.TenantID, session.UserID,
               session.UserType, session.PrincipalID, session.EntityID)
            c.Locals("db", conn)
        }
        return c.Next()
    }
}
```

---

### Two-Path Authorization

| Path | Middleware | When to use |
|---|---|---|
| Fast path | `RequirePermission(resource, action)` — O(1) session map | Hot-path API routes |
| Casbin path | `authz.Middleware(obj, act)` | Management ops where session may be stale after role changes |

```go
// Hot path (most routes)
finance.Get("/invoices",
    middleware.RequirePermission("finance.receivables.invoices", "read"),
    handler.ListInvoices)

// Casbin path (role management operations)
admin.Post("/roles/:id/assign",
    authzSvc.Middleware("role", "assign"),
    handler.AssignRole)
```

---

Next: [IAM Services Reference](./14b-iam-services.md)
