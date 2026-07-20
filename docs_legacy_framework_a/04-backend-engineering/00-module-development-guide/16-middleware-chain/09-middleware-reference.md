> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Middleware Quick Reference
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[Middleware Overview](01-middleware-overview.md)"
  - "[Auth Middleware](02-auth-middleware.md)"
  - "[Authz Middleware](03-authz-middleware.md)"
  - "[Request ID Middleware](04-request-id-middleware.md)"
  - "[Rate Limiting](05-rate-limiting.md)"
  - "[Tenant Context Middleware](06-tenant-context-middleware.md)"
---

# Middleware Quick Reference

## Global Middleware Chain

Registered on the Fiber app (all routes):

```go
app.Use(requestid.New(...))        // 1. generate X-Request-ID
app.Use(corsMiddleware)            // 2. CORS headers + preflight
app.Use(securityHeaders)           // 3. HSTS, CSP, X-Frame-Options
app.Use(recover.New(...))          // 4. catch panics → 500
app.Use(logger.New(...))           // 5. structured request log
app.Use(globalRateLimit)           // 6. IP rate limiting (60/min default)
```

## Per-Route Middleware

Applied per route or group:

```go
// On every authenticated route group
api.Use(middleware.Authenticate(sessionSvc))
api.Use(middleware.TenantContext(tenantSvc))

// On specific routes
api.Post("/contracts", middleware.Authorize(authzSvc, "contracts.contract.create"), handler.Create)
api.Post("/auth/login", middleware.StrictRateLimit(10, time.Minute), handler.Login)
```

## Authenticate Middleware

```go
func Authenticate(sessionSvc iam.SessionService) fiber.Handler {
    return func(c *fiber.Ctx) error {
        authHeader := c.Get("Authorization")
        if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
            return fiber.NewError(401, "authorization header required")
        }
        token := strings.TrimPrefix(authHeader, "Bearer ")

        sess, err := sessionSvc.Validate(c.UserContext(), token)
        if err != nil {
            return fiber.NewError(401, "invalid or expired session")
        }

        c.Locals("resolved_session", sess)
        return c.Next()
    }
}
```

## Authorize Middleware

```go
func Authorize(authzSvc authz.Service, permission string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        sess := SessionFrom(c)
        if err := authzSvc.Can(c.UserContext(), sess.ToPrincipal(), permission); err != nil {
            return fiber.NewError(403, "permission denied")
        }
        return c.Next()
    }
}
```

## SessionFrom Helper

```go
func SessionFrom(c *fiber.Ctx) iam.ResolvedSession {
    sess, ok := c.Locals("resolved_session").(iam.ResolvedSession)
    if !ok {
        panic("resolved_session not set — Authenticate middleware missing on this route")
    }
    return sess
}
```

Panics intentionally on missing session — caught by Recover middleware → 500 + logged stack trace. Indicates a wiring bug.

## Tenant Context Middleware

```go
func TenantContext(tenantSvc tenant.Service) fiber.Handler {
    return func(c *fiber.Ctx) error {
        sess := SessionFrom(c)
        t, err := tenantSvc.GetByID(c.UserContext(), sess.TenantID)
        if err != nil || t.Status != tenant.StatusActive {
            return fiber.NewError(403, "tenant account is not active")
        }
        return c.Next()
    }
}
```

Validates tenant is still ACTIVE on each request. Handles suspended tenants transparently.

## StrictRateLimit

```go
func StrictRateLimit(max int, window time.Duration) fiber.Handler {
    return limiter.New(limiter.Config{
        Max:        max,
        Expiration: window,
        KeyGenerator: func(c *fiber.Ctx) string {
            return c.IP()
        },
        LimitReached: func(c *fiber.Ctx) error {
            return fiber.NewError(429, "rate limit exceeded")
        },
    })
}
```

## Recover Middleware

```go
recover.New(recover.Config{
    EnableStackTrace: cfg.Env != "production",   // stack in dev, not prod
    StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
        slog.Error("panic recovered",
            "path", c.Path(),
            "panic", e,
            "request_id", c.GetRespHeader("X-Request-ID"),
        )
    },
})
```

## Custom Error Handler

```go
func customErrorHandler(c *fiber.Ctx, err error) error {
    code := 500
    msg := "internal server error"

    var fiberErr *fiber.Error
    if errors.As(err, &fiberErr) {
        code = fiberErr.Code
        msg = fiberErr.Message
    }

    return c.Status(code).JSON(fiber.Map{
        "error": fiber.Map{
            "code":    httpCodeToErrorCode(code),
            "message": msg,
        },
    })
}

func httpCodeToErrorCode(status int) string {
    switch status {
    case 400: return "BAD_REQUEST"
    case 401: return "UNAUTHORIZED"
    case 403: return "FORBIDDEN"
    case 404: return "NOT_FOUND"
    case 409: return "CONFLICT"
    case 422: return "UNPROCESSABLE_ENTITY"
    case 429: return "TOO_MANY_REQUESTS"
    default:  return "INTERNAL_SERVER_ERROR"
    }
}
```

## Rate Limit Tiers

| Endpoint type | Limit |
|--------------|-------|
| Auth (`/auth/login`, `/auth/refresh`) | 10/min/IP |
| Standard API endpoints | 60/min/IP |
| Bulk import (`/import`) | 5/min/IP |
| Health checks | Unlimited |
| Metrics | Unlimited |
