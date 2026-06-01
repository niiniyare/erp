---
title: Rate Limiting
portal: 4 — Backend Engineering
section: 00-module-development-guide/16-middleware-chain
audience: [backend-engineer, tech-lead]
related:
  - path: ./01-middleware-overview.md
    title: Middleware Overview
---

# Rate Limiting

## Rate Limit Tiers

| Tier | Limit | Window | Applied To |
|------|-------|--------|-----------|
| Global per-IP | 1000 req | 1 minute | All routes |
| Auth endpoints | 10 req | 1 minute | `/auth/*` |
| Write endpoints | 100 req | 1 minute | POST/PUT/DELETE |
| Import endpoint | 5 req | 1 minute | `/contracts/import` |
| Export endpoint | 10 req | 1 minute | `/contracts/export` |

## Implementation

```go
// internal/platform/middleware/ratelimit.go
package middleware

import (
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/limiter"
)

// RateLimit returns a global per-IP rate limiter.
func RateLimit() fiber.Handler {
    return limiter.New(limiter.Config{
        Max:        1000,
        Expiration: time.Minute,
        KeyGenerator: func(c *fiber.Ctx) string {
            return c.IP()
        },
        LimitReached: func(c *fiber.Ctx) error {
            return fiber.NewError(fiber.StatusTooManyRequests, "rate limit exceeded")
        },
    })
}

// StrictRateLimit returns a tight rate limiter for sensitive endpoints.
func StrictRateLimit(max int, window time.Duration) fiber.Handler {
    return limiter.New(limiter.Config{
        Max:        max,
        Expiration: window,
        KeyGenerator: func(c *fiber.Ctx) string {
            // Key by IP + path to limit per-endpoint
            return c.IP() + ":" + c.Path()
        },
        LimitReached: func(c *fiber.Ctx) error {
            return fiber.NewError(fiber.StatusTooManyRequests, "rate limit exceeded")
        },
    })
}
```

## Route-Level Application

```go
// Apply strict limit to auth endpoints
auth := app.Group("/api/v1/auth",
    middleware.StrictRateLimit(10, time.Minute),
)

// Apply import limit
contracts.Post("/import",
    middleware.StrictRateLimit(5, time.Minute),
    middleware.Authenticate(sessionSvc),
    middleware.Authorize(authzSvc, "contracts.contract.create"),
    handlers.Import,
)
```

## Rate Limit Headers

Always return rate limit info in response headers:

```
X-RateLimit-Limit:     1000
X-RateLimit-Remaining: 997
X-RateLimit-Reset:     1704067200
```

The Fiber limiter middleware sets these automatically.

## Tenant-Level Rate Limits

For multi-tenant SaaS, apply per-tenant limits based on plan tier:

```go
func TenantRateLimit(sessionSvc iam.SessionService) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Only applies after Authenticate
        session, ok := c.Locals(domain.LocalsKeySession).(*domain.ResolvedSession)
        if !ok {
            return c.Next()
        }

        maxRPM := session.SettingInt("rate_limit.requests_per_minute", 500)
        key    := "tenant:" + session.TenantID.String()

        // Use Redis sliding window counter
        count, err := rateLimitStore.Increment(key, time.Minute)
        if err != nil || count > maxRPM {
            return fiber.NewError(fiber.StatusTooManyRequests, "tenant rate limit exceeded")
        }
        return c.Next()
    }
}
```
