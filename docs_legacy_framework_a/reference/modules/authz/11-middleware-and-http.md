> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

[<-- Back to Index](README.md)

## Middleware & HTTP Integration

### Overview

The `Middleware` method returns a Fiber handler factory that is the primary way HTTP routes consume the authz module. It requires zero boilerplate — one call per route.

```go
func (s *service) Middleware(object, action string) fiber.Handler
```

**Two-layer approach**: The `authz.Middleware()` factory (Casbin, source-of-truth, used for management operations) works alongside a fast O(1) map check via `ResolvedSession.Can()` for hot-path requests. Use whichever fits the context.

### Basic Usage

```go
// In your Fiber route setup:
app.Get("/invoices",          svc.Middleware("invoice", "read"),   listInvoices)
app.Post("/invoices",         svc.Middleware("invoice", "create"), createInvoice)
app.Put("/invoices/:id",      svc.Middleware("invoice", "update"), updateInvoice)
app.Delete("/invoices/:id",   svc.Middleware("invoice", "delete"), deleteInvoice)
app.Post("/invoices/:id/approve", svc.Middleware("invoice", "approve"), approveInvoice)
```

### What the Middleware Does

```markdown
Middleware("invoice", "read") creates a handler that:

1. Read Principal from c.Locals("authz_principal")
   → Set by Authenticate middleware earlier in the chain
   → If missing → 401 Unauthorized

2. Build object string:
   → No :id param → object = "invoice"
   → With :id param → object = "invoice/inv_123"
   → This enables per-resource wildcard policies:
      "invoice/*"   matches any specific invoice
      "invoice"     matches the collection (list/create)

3. Call service.Enforce(c.Context(), Request{
       Subject: principal.Subject,
       Domain:  principal.Domain,
       Object:  obj,
       Action:  "read",
   })

4. false → fiber.NewError(403, "access denied")
5. true  → c.Next()
```

### Middleware Chain Order (sys-desing.md)

```markdown
REQUEST PIPELINE:

HTTP Request
     │
     ▼
Logger → Recovery → CORS → RateLimiter
     │
     ▼
Authenticate  (internal/api/middleware/jwt_auth.go — REWRITTEN)
  → reads opaque session token from Cookie or Authorization: Bearer header
  → calls identity/session.Service.ValidateSession(ctx, token)
  → ValidateSession: sha256(token) → DB lookup → checks is_active + expires_at
  → builds ResolvedSession{UserID, UserType, TenantID, PrincipalID, Permissions map}
  → c.Locals("session", resolved)                          ← for handlers
  → c.Locals("authz_principal", resolved.ToPrincipal())   ← for authz.Middleware()
  → sets ctx with tenant_id and user_id via context.WithValue
     │
     ▼
ResolveTenant  (internal/api/middleware/tenant.go — existing, unchanged)
  → reads tenantID from c.Locals or request header
  → validates tenant status via tenant.Service
     │
     ▼
SetTenantContext
  → calls store.SetTenantContextFromCtx(ctx)  (sets PostgreSQL RLS context)
     │
     ▼
Authorize  (internal/api/middleware/authorization.go — REWRITTEN, per-route)
  → fast path: session.Can("finance.receivables.invoices.read")  O(1) map lookup
  → or Casbin path: authz.Middleware("invoice", "read") for management ops
     │
     ▼
Route Handler
  → handler can trust: "if we got here, user is authorized"
  → reads session: c.Locals("session").(*session.ResolvedSession)
  → tenant_id and user_id are in ctx via context.Value — no extra params needed
```

### Context Values — No Extra Function Params

`tenant_id` and `user_id` are passed via `context.Value` using the keys defined in `internal/platform/cache`:

```go
// internal/platform/cache/cache.go defines:
const (
    TenantIDKey   contextKey = "tenant_id"
    TenantSlugKey contextKey = "tenant_subdomain"
)

// The Authenticate middleware sets these after ValidateSession:
ctx = context.WithValue(ctx, cache.TenantIDKey, resolved.TenantID.String())
ctx = context.WithValue(ctx, userIDKey, resolved.UserID.String())

// Any service method reads from ctx — no extra params needed:
func (s *service) SomeOperation(ctx context.Context) error {
    tenantID := ctx.Value(cache.TenantIDKey).(string)
    // ...
}
```

### Authenticate Middleware (Rewritten)

`internal/api/middleware/jwt_auth.go` — replaces dead iam imports with identity/session:

```go
// AuthConfig holds dependencies — passed at wire-up time
type AuthConfig struct {
    SessionSvc  session.Service
    CookieName  string // default: "session"
    FeatureFlags condition.Evaluator // for MFA flag, etc.
}

func Authenticate(cfg AuthConfig) fiber.Handler {
    return func(c *fiber.Ctx) error {
        token := extractToken(c, cfg.CookieName)
        if token == "" {
            return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
        }

        resolved, err := cfg.SessionSvc.ValidateSession(c.Context(), token)
        if err != nil {
            return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired session")
        }

        // Set locals for downstream middleware and handlers
        c.Locals(session.LocalsKeySession, resolved)
        c.Locals(authz.LocalsKeyPrincipal, resolved.ToPrincipal())

        // Set context values — downstream services read these, no extra params
        ctx := context.WithValue(c.Context(), cache.TenantIDKey, resolved.TenantID.String())
        c.SetUserContext(ctx)
        return c.Next()
    }
}
```

### Authorize Middleware (Rewritten)

`internal/api/middleware/authorization.go` — replaces dead iam imports:

```go
// Authorize uses the pre-computed permission map — O(1), zero DB
func Authorize(permission string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        sess, ok := c.Locals(session.LocalsKeySession).(*session.ResolvedSession)
        if !ok || sess == nil {
            return fiber.NewError(fiber.StatusUnauthorized, "no session")
        }
        if !sess.Can(permission) {
            return fiber.NewError(fiber.StatusForbidden, "insufficient permissions")
        }
        return c.Next()
    }
}

// AuthorizeCasbin uses the Casbin engine — for management ops where the
// pre-computed map may be stale (after role changes without re-login)
func AuthorizeCasbin(svc authz.Service, object, action string) fiber.Handler {
    return svc.Middleware(object, action)
}
```

### Route Groups with Shared Middleware

```go
api := app.Group("/api")

// Public — no auth
api.Post("/auth/login",  loginHandler)

// All authenticated routes
auth := api.Group("",
    middleware.Authenticate(authCfg),
    middleware.ResolveTenant(tenantCfg),
)

// Finance routes — permission string follows {module}.{resource}.{action}
finance := auth.Group("/finance")
finance.Get("/invoices",
    middleware.Authorize("finance.receivables.invoices.read"),
    handler.ListInvoices)
finance.Post("/invoices",
    middleware.Authorize("finance.receivables.invoices.create"),
    handler.CreateInvoice)
finance.Post("/invoices/:id/approve",
    middleware.Authorize("finance.receivables.invoices.approve"),
    handler.ApproveInvoice)

auth.Post("/auth/logout", logoutHandler)
```

### Route Groups with Shared Middleware

For a group of routes requiring the same base permission:

```go
// Finance routes — all require finance-viewer minimum
finance := app.Group("/finance",
    authnMiddleware,       // JWT validation (sets Principal)
)

// List/read operations
finance.Get("/invoices",       svc.Middleware("invoice", "read"),   listInvoices)
finance.Get("/invoices/:id",   svc.Middleware("invoice", "read"),   getInvoice)
finance.Get("/payments",       svc.Middleware("payment", "read"),   listPayments)

// Mutating operations — stricter permission
finance.Post("/invoices",      svc.Middleware("invoice", "create"), createInvoice)
finance.Put("/invoices/:id",   svc.Middleware("invoice", "update"), updateInvoice)
finance.Delete("/invoices/:id",svc.Middleware("invoice", "delete"), deleteInvoice)

// High-sensitivity operations
finance.Post("/invoices/:id/approve",
    svc.Middleware("invoice", "approve"), approveInvoice)
finance.Post("/periods/:id/close",
    svc.Middleware("period", "close"), closePeriod)
```

### Custom Object Expansion

The default middleware expands `object + "/" + id` when an `:id` Fiber param exists. For more complex resource paths, build the middleware manually:

```go
// Route: /tenants/:tenantID/invoices/:id
// Object should be: "tenant/{tenantID}/invoice/{id}"
app.Get("/tenants/:tenantID/invoices/:id", func(c *fiber.Ctx) error {
    p := c.Locals(authz.LocalsKeyPrincipal).(authz.Principal)
    obj := "tenant/" + c.Params("tenantID") + "/invoice/" + c.Params("id")

    ok, err := svc.Enforce(c.Context(), authz.Request{
        Subject: p.Subject,
        Domain:  p.Domain,
        Object:  obj,
        Action:  "read",
    })
    if err != nil { return err }
    if !ok { return fiber.NewError(fiber.StatusForbidden) }
    return c.Next()
}, getInvoiceHandler)
```

### Batch Permission Pre-Check (EnforceBatch)

For UIs that need to show/hide multiple action buttons at once:

```go
// Before rendering the invoice detail page:
// Check which actions the user can perform
results, err := svc.EnforceBatch(ctx, []authz.Request{
    {Subject: sub, Domain: dom, Object: "invoice/" + id, Action: "read"},
    {Subject: sub, Domain: dom, Object: "invoice/" + id, Action: "update"},
    {Subject: sub, Domain: dom, Object: "invoice/" + id, Action: "delete"},
    {Subject: sub, Domain: dom, Object: "invoice/" + id, Action: "approve"},
    {Subject: sub, Domain: dom, Object: "invoice/" + id, Action: "export"},
})
// results[0] = canRead
// results[1] = canUpdate
// results[2] = canDelete
// results[3] = canApprove
// results[4] = canExport

// Return to frontend as permissions object:
return c.JSON(fiber.Map{
    "permissions": fiber.Map{
        "read":    results[0],
        "update":  results[1],
        "delete":  results[2],
        "approve": results[3],
        "export":  results[4],
    },
})
```

### Error Responses

The middleware returns standard Fiber errors which map to HTTP status codes:

| Condition | HTTP Status | Error Code |
|-----------|-------------|------------|
| No Principal in Locals | 401 | `AUTHZ_UNAUTHORIZED` |
| Enforce returns false | 403 | `AUTHZ_FORBIDDEN` |
| Enforce returns error | 500 | (wrapped error) |
| Invalid request fields | 400 | `AUTHZ_INVALID` |

```json
// 403 response body (Fiber default error format):
{
    "error": "[authz] AUTHZ_FORBIDDEN: access denied"
}

// 401 response body:
{
    "error": "[authz] AUTHZ_UNAUTHORIZED: authentication required"
}
```

### Integration with OpenTelemetry (Optional)

When `cfg.Tracer` is non-nil, each `Enforce` call creates a span:

```markdown
Span: "authz.Enforce"
Attributes:
  authz.subject   = "tenant:usr_001"
  authz.domain    = "a1b2c3d4-uuid"
  authz.object    = "invoice/inv_123"
  authz.action    = "read"
  authz.allowed   = true | false
  authz.latency   = 0.4ms

This makes authorization decisions visible in Jaeger/Tempo traces.
Denied requests appear as error spans in the trace.
```

---

Next: [System Module Integration](./12-system-module-integration.md)
