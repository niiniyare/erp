---
title: CORS and Security Headers
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, devops]
related:
  - "[Middleware Overview](01-middleware-overview.md)"
  - "[Auth Middleware](02-auth-middleware.md)"
  - "[Security Overview](../../../07-security/01-security-overview.md)"
---

# CORS and Security Headers

## CORS Configuration

AwoERP uses subdomain-based multi-tenancy. CORS must allow requests from tenant subdomains:

```go
// internal/server/fiber.go
app.Use(cors.New(cors.Config{
    AllowOriginsFunc: func(origin string) bool {
        // Allow any subdomain of awoerp.com
        return strings.HasSuffix(origin, ".awoerp.com") ||
               origin == "http://localhost:3000" ||           // local frontend dev
               origin == "http://localhost:8080"
    },
    AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
    AllowHeaders:     "Origin,Content-Type,Authorization,X-Request-ID",
    ExposeHeaders:    "Location,X-Request-ID",
    AllowCredentials: false,    // no cookies — we use Authorization header
    MaxAge:           86400,    // 24h preflight cache
}))
```

**`AllowCredentials: false`** — AwoERP sends auth via `Authorization` header, not cookies. Setting this to true is unnecessary and loosens security.

## Security Headers

```go
app.Use(func(c *fiber.Ctx) error {
    c.Set("X-Content-Type-Options", "nosniff")
    c.Set("X-Frame-Options", "DENY")
    c.Set("X-XSS-Protection", "0")                          // modern browsers: rely on CSP not this header
    c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
    c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
    c.Set("Content-Security-Policy",
        "default-src 'none'; "+
        "script-src 'self'; "+
        "style-src 'self' 'unsafe-inline'; "+   // AMIS requires inline styles
        "img-src 'self' data:; "+
        "font-src 'self'; "+
        "connect-src 'self'; "+
        "frame-ancestors 'none'")
    return c.Next()
})
```

### CSP Notes

AMIS SDK requires `unsafe-inline` for styles — it injects inline styles for dynamic theming. This is a known trade-off. Do not remove `unsafe-inline` from `style-src` or the UI will break.

`script-src 'self'` only — no CDN scripts. All JS is served from the same origin.

## Request ID Header

Every response includes a request ID for correlation:

```go
app.Use(requestid.New(requestid.Config{
    Header: "X-Request-ID",
    Generator: func() string {
        return uuid.New().String()
    },
}))
```

Clients can pass their own `X-Request-ID` header; if absent, one is generated. The ID is propagated into structured logs as `request_id` field.

## Content-Type Enforcement

```go
// Only accept JSON for write endpoints
app.Use("/api/v1", func(c *fiber.Ctx) error {
    if c.Method() != fiber.MethodGet && c.Method() != fiber.MethodDelete {
        ct := c.Get("Content-Type")
        if ct != "" && !strings.HasPrefix(ct, "application/json") &&
           !strings.HasPrefix(ct, "multipart/form-data") {   // for file uploads
            return fiber.NewError(415, "unsupported media type — use application/json")
        }
    }
    return c.Next()
})
```

## Middleware Order

Security and utility middleware runs before auth:

```
1. RequestID      ← generate/propagate X-Request-ID
2. CORS           ← preflight + headers
3. SecurityHeaders ← HSTS, CSP, X-Frame-Options
4. Recover        ← catch panics
5. Logger         ← log requests (includes request_id from step 1)
6. RateLimit      ← IP-based rate limiting
7. Authenticate   ← validate session (per-route)
8. Authorize      ← check permission (per-route)
9. Handler        ← business logic
```

Security headers and CORS must run before auth to ensure preflight OPTIONS requests get correct headers without requiring a session token.
