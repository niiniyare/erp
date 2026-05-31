[<-- Back to Index](README.md)

## Middleware & HTTP Integration

> **[IMPLEMENTED]** — describes the v1.0 middleware architecture.
>
> **IMPORTANT**: A prior version of this document described a two-path authorization model:
> a `session.Can()` "fast path" and a Casbin "management path". That model is NOT active.
> All authorization in v1.0 goes through `authzService.Enforce()` (Casbin).
> `session.Can()` and `session.CanDo()` do not exist.

---

### Middleware Chain [IMPLEMENTED]

```
HTTP Request
     │
     ▼
Logger → Recovery → CORS → RateLimiter
     │
     ▼
Authenticate middleware
  → reads token from Cookie ("session") or Authorization: Bearer header
  → if token == "" → 401
  → calls sessionService.ValidateSession(ctx, token)
      → sha256hex(token)
      → Redis cache-aside ("session:{hash}") → DB fallback (TouchAndGetSession)
      → returns nil → 401
  → c.Locals("resolved_session") = *domain.ResolvedSession
  → c.Locals("authz_principal")  = resolved.ToPrincipal()
  → ctx = context.WithValue(ctx, cache.TenantIDKey, tenantID.String())
  → c.SetUserContext(ctx)
     │
     ▼
SetTenantContext middleware
  → executes: SET LOCAL awo.tenant_id = '<TenantID>'
  → PostgreSQL RLS now active for this connection
     │
     ▼
Authorization middleware (per-route)
  → reads Principal from c.Locals("authz_principal")
  → calls authzService.Enforce(ctx, domain.Request{Subject, Domain, Object, Action})
  → false → 403 Forbidden
  → true  → c.Next()
     │
     ▼
Route Handler
```

---

### Authorization Middleware [IMPLEMENTED]

Authorization calls `authzService.Enforce()`:

```go
// Pseudo-code for authorization middleware
func Authorize(authzSvc AuthzService, object, action string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        principal, ok := c.Locals("authz_principal").(domain.Principal)
        if !ok {
            return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
        }
        ok, err := authzSvc.Enforce(c.Context(), domain.Request{
            Subject: principal.Subject,
            Domain:  principal.Domain,
            Object:  object,
            Action:  action,
        })
        if err != nil {
            return fiber.NewError(fiber.StatusInternalServerError, "authorization error")
        }
        if !ok {
            return fiber.NewError(fiber.StatusForbidden, "access denied")
        }
        return c.Next()
    }
}
```

This is the only authorization path. There is no `session.Can()` or `session.CanDo()` fast path.

---

### Principal Construction [IMPLEMENTED]

The `Principal` is built from `ResolvedSession.ToPrincipal()`:

```go
// ActorType → Subject/Domain mapping
switch ActorTypeFromUserType(sess.UserType) {
case ActorPlatform:
    Subject: "platform:<userID>", Domain: "_platform_"
case ActorPortal:
    Subject: "portal:<userID>",   Domain: "<tenantID>:portal"
case ActorAPI:
    Subject: "api:<userID>",      Domain: "<tenantID>:api"
default: // ActorTenant
    Subject: "tenant:<userID>",   Domain: "<tenantID>"
}
```

---

### Route Setup Examples [IMPLEMENTED]

```go
// Public routes — no auth
api.Post("/auth/login", loginHandler)
api.Get("/auth/sso/:provider/begin", ssoBeginHandler)
api.Get("/auth/sso/callback", ssoCallbackHandler)

// Authenticated + tenant-scoped routes
auth := api.Group("",
    middleware.Authenticate(sessionSvc),
    middleware.SetTenantContext(store),
)

// Per-route authorization via Casbin
auth.Get("/invoices",
    middleware.Authorize(authzSvc, "invoice", "read"),
    handler.ListInvoices)

auth.Post("/invoices",
    middleware.Authorize(authzSvc, "invoice", "create"),
    handler.CreateInvoice)

auth.Delete("/invoices/:id",
    middleware.Authorize(authzSvc, "invoice", "delete"),
    handler.DeleteInvoice)
```

---

### Reading Session in Handlers [IMPLEMENTED]

```go
func MyHandler(c *fiber.Ctx) error {
    sess, ok := c.Locals("resolved_session").(*domain.ResolvedSession)
    if !ok || sess == nil {
        return fiber.ErrUnauthorized
    }

    // Identity
    userID   := sess.UserID
    tenantID := sess.TenantID

    // Feature flags (O(1) — no DB, no Casbin)
    if sess.FeatureEnabled("finance.multi_currency") {
        // show currency selector
    }

    // Tenant settings (O(1))
    threshold := sess.SettingDecimal("finance.approval_threshold", 0)

    // User preferences (O(1))
    theme := sess.Configuration.Prefs["ui.theme"]

    // Entity scope (for query filtering, not authorization)
    scope := sess.EntityScope
    // Pass scope to repository methods that build WHERE clauses

    // NOTE: session.Can() does NOT exist. If you need a permission check
    // in business logic, use authzSvc.Enforce() directly.
    return c.JSON(...)
}
```

---

### Batch Permission Checks (EnforceBatch) [IMPLEMENTED]

For UIs that need to show/hide multiple action buttons:

```go
results, err := authzSvc.EnforceBatch(ctx, []domain.Request{
    {Subject: p.Subject, Domain: p.Domain, Object: "invoice/" + id, Action: "read"},
    {Subject: p.Subject, Domain: p.Domain, Object: "invoice/" + id, Action: "update"},
    {Subject: p.Subject, Domain: p.Domain, Object: "invoice/" + id, Action: "delete"},
    {Subject: p.Subject, Domain: p.Domain, Object: "invoice/" + id, Action: "approve"},
})
// results[0] = canRead, [1] = canUpdate, [2] = canDelete, [3] = canApprove
```

---

### What Is NOT Implemented

- **[REMOVED]**: `session.Can(permission string) bool` — was the old fast-path; does not exist.
- **[REMOVED]**: `session.CanDo(resource, action string) bool` — same.
- **[REMOVED]**: `Authorize("finance.receivables.invoices.read")` session-map-based middleware.
- **[PARTIAL]**: The actual Fiber middleware wiring (`middleware.Authenticate`, `middleware.Authorize`) exists in `internal/api/middleware/`. Confirm wiring at application startup — this document describes the contract, not the wire-up code.

---

### Error Responses

| Condition | HTTP Status | Body |
|-----------|-------------|------|
| No token in request | 401 | `"authentication required"` |
| Token not found / expired | 401 | `"invalid or expired session"` |
| Enforce returns false | 403 | `"access denied"` |
| Enforce returns error | 500 | wrapped error |

---

Next: [Performance & Caching](./16-performance-and-caching.md)
