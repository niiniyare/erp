[<-- Back to Index](README.md)

## Middleware & HTTP Integration

### Overview

The `Middleware` method returns a Fiber handler factory that is the primary way HTTP routes consume the authz module. It requires zero boilerplate — one call per route.

```go
func (s *service) Middleware(object, action string) fiber.Handler
```

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
   → Set by authn middleware earlier in the chain
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

### Middleware Chain Order

```markdown
REQUEST PIPELINE:

HTTP Request
     │
     ▼
Rate Limiter middleware
     │
     ▼
authn middleware (JWT validation)
  → validates token signature
  → extracts: userID, tenantID, actorType
  → builds Principal{Subject: "tenant:usr_001", Domain: "tenant-abc"}
  → c.Locals("authz_principal", principal)
     │
     ▼
authz.Middleware("invoice", "read")  ← PER ROUTE
  → reads Principal from c.Locals
  → calls Enforce
  → 403 or next
     │
     ▼
Route Handler
  → handler can trust: "if we got here, user is authorized"
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
