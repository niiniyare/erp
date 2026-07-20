> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Authentication Middleware
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Middleware Overview](01-middleware-overview.md)"
  - "[Authz Middleware](03-authz-middleware.md)"
  - "[Session Architecture](../../../03-platform-architecture/02-iam/02-session-architecture.md)"
---

# Authentication Middleware

## Purpose

The `Authenticate` middleware:

1. Extracts the session token from the `Authorization: Bearer <token>` header
2. Validates and resolves the token via `SessionService`
3. Stores the `ResolvedSession` in Fiber locals under `domain.LocalsKeySession`
4. Returns `401` if the token is missing or invalid

## Implementation

```go
// internal/platform/middleware/authenticate.go
package middleware

import (
    "github.com/gofiber/fiber/v2"

    "awo.so/internal/core/iam"
    "awo.so/internal/core/iam/domain"
)

// Authenticate returns a middleware that validates the session token.
func Authenticate(sessionSvc iam.SessionService) fiber.Handler {
    return func(c *fiber.Ctx) error {
        authHeader := c.Get("Authorization")
        if authHeader == "" {
            return fiber.NewError(fiber.StatusUnauthorized, "authorization header required")
        }

        if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
            return fiber.NewError(fiber.StatusUnauthorized, "invalid authorization header format")
        }

        token := authHeader[7:]
        session, err := sessionSvc.Resolve(c.Context(), token)
        if err != nil {
            if errors.Is(err, iam.ErrSessionExpired) {
                return fiber.NewError(fiber.StatusUnauthorized, "session expired")
            }
            if errors.Is(err, iam.ErrSessionInvalid) {
                return fiber.NewError(fiber.StatusUnauthorized, "invalid session token")
            }
            return fiber.NewError(fiber.StatusUnauthorized, "authentication failed")
        }

        // Store in locals for downstream middleware and handlers
        c.Locals(domain.LocalsKeySession, session)
        return c.Next()
    }
}
```

## SessionFrom Helper

Use this in handlers and downstream middleware to extract the session:

```go
// internal/platform/middleware/session.go
package middleware

import (
    "github.com/gofiber/fiber/v2"
    "awo.so/internal/core/iam/domain"
)

// SessionFrom extracts the ResolvedSession from Fiber locals.
// Panics if Authenticate middleware was not applied — this is intentional;
// a missing session is a programming error, not a runtime error.
func SessionFrom(c *fiber.Ctx) *domain.ResolvedSession {
    session, ok := c.Locals(domain.LocalsKeySession).(*domain.ResolvedSession)
    if !ok || session == nil {
        panic("SessionFrom called without Authenticate middleware")
    }
    return session
}
```

## ResolvedSession Fields

```go
type ResolvedSession struct {
    UserID      uuid.UUID
    TenantID    uuid.UUID
    EntityScope EntityScope
    Roles       []string
    Settings    map[string]string
    Features    map[string]bool
    ExpiresAt   time.Time
}

// FeatureEnabled returns true if the feature flag is enabled for this session's tenant.
func (s *ResolvedSession) FeatureEnabled(key string) bool

// SettingString returns a tenant-level setting string value.
func (s *ResolvedSession) SettingString(key, defaultValue string) string

// ToPrincipal converts the session to an IAM Principal for authz service calls.
func (s *ResolvedSession) ToPrincipal() iam.Principal

// EntityID returns the entity ID from the session scope (when scope type is Entity).
func (s *ResolvedSession) EntityID() uuid.UUID
```

## 401 vs. 403

| Code | Meaning |
|------|---------|
| 401 | No session or session invalid — who are you? |
| 403 | Session valid but permission denied — you can't do this |

Never return 403 from `Authenticate` — it always returns 401 for missing/invalid tokens.
