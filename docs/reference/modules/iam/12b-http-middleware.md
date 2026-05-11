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
        middleware.Authorize(deps.Platform.IAM, "finance.transactions.read"),
        handlers.TransactionListSchema(deps))
    sg.Get("/settings/modules",
        middleware.Authorize(deps.Platform.IAM, "settings.modules.read"),
        handlers.ModuleFlagsSchema(deps))
    sg.Get("/settings/finance",
        middleware.Authorize(deps.Platform.IAM, "settings.finance.read"),
        handlers.FinanceSettingsSchema(deps))

    // Data API routes
    api := auth.Group("/api/v1")
    api.Get("/transactions",
        middleware.RequireFlag("finance.transactions"),
        middleware.Authorize(deps.Platform.IAM, "finance.transactions.read"),
        handlers.ListTransactions(deps))
    api.Patch("/settings/flags/:key",
        middleware.Authorize(deps.Platform.IAM, "settings.modules.update"),
        handlers.SetFlag(deps))
    api.Patch("/settings/:module",
        middleware.Authorize(deps.Platform.IAM, "settings.finance.update"),
        handlers.UpdateModuleSettings(deps))
}
```

---

### Middleware Implementations

All authorization routes through Casbin — no permission snapshot in the session.

```go
// Authorize enforces a single "object.action" permission via Casbin.
// Must run after Authenticate (requires LocalsKeySession and LocalsKeyPrincipal).
//
// Responses:
//   401 — no authenticated session in Locals
//   500 — Enforce returned a non-nil error (enforcer unavailable)
//   403 — Enforce returned false (permission denied)
//   next — Enforce returned true (allowed)
func Authorize(cfg AuthConfig, permission string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        sess, ok := c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession)
        if !ok || sess == nil {
            return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
        }
        if cfg.AuthzService == nil {
            return fiber.NewError(fiber.StatusForbidden, "permission denied")
        }
        principal, ok := c.Locals(iam.LocalsKeyPrincipal).(iam.Principal)
        if !ok {
            return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
        }
        obj, act := splitPermission(permission)
        allowed, err := cfg.AuthzService.Enforce(c.Context(), iam.Request{
            Subject: principal.Subject,
            Domain:  principal.Domain,
            Object:  obj,
            Action:  act,
        })
        if err != nil {
            return fiber.NewError(fiber.StatusInternalServerError, "authorization check failed")
        }
        if !allowed {
            return fiber.NewError(fiber.StatusForbidden, "permission denied")
        }
        return c.Next()
    }
}

// RequireFlag checks whether the session's pre-computed Configuration has the
// named feature flag enabled. No Casbin call — flags are context, not authz.
// Returns 403 if the flag is absent or false.
func RequireFlag(flagKey string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        sess, ok := c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession)
        if !ok || sess == nil {
            return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
        }
        if !sess.FeatureEnabled(flagKey) {
            return fiber.NewError(fiber.StatusForbidden, "feature not enabled")
        }
        return c.Next()
    }
}
```

**Key invariants:**
- `Authorize` calls `authzService.Enforce()` on every request — no cached permission map
- `RequireFlag` reads `session.Configuration.Flags` — O(1), no Casbin, no DB
- Subject and Domain always come from the authenticated session, never from the request body

---

### Single-Path Authorization

There is one enforcement path: every permission check calls `Casbin.Enforce()`.

```
Request
  → Authenticate        (validates token, populates LocalsKeySession + LocalsKeyPrincipal)
  → RequireFlag(...)    (optional; feature gate from session.Configuration — not authz)
  → Authorize(cfg, ...) (Casbin.Enforce → 401 | 500 | 403 | next)
  → Handler
```

```go
// All routes use the same Casbin path:
finance.Get("/invoices",
    middleware.RequireFlag("finance.transactions"),           // feature gate
    middleware.Authorize(cfg, "finance.receivables.invoices.read"), // Casbin enforce
    handler.ListInvoices)

admin.Post("/roles/:id/assign",
    middleware.Authorize(cfg, "role.assign"),                // Casbin enforce
    handler.AssignRole)
```

> **Removed (v1.0)**: `session.CanDo()` / `session.Can()` fast path. The session no
> longer carries a permission snapshot. All authorization is Casbin-only. This ensures
> role revocations take effect on the next request without requiring session invalidation.

---

### Context Helpers

Used in every handler — never pass tenant_id or user_id as function parameters:

```go
func ContextSession(c *fiber.Ctx) *iam.ResolvedSession {
    return c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession)
}
func ContextTenantID(c *fiber.Ctx) uuid.UUID  { return ContextSession(c).TenantID }
func ContextUserID(c *fiber.Ctx) uuid.UUID    { return ContextSession(c).UserID }
func ContextEntityID(c *fiber.Ctx) uuid.UUID  { return ContextSession(c).EntityScope.EntityID }
func ContextPrincipalID(c *fiber.Ctx) *uuid.UUID { return ContextSession(c).PrincipalID }
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

Next: [IAM Services Reference](./14b-iam-services.md)
