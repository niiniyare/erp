---
title: Middleware Testing
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Middleware Overview](01-middleware-overview.md)"
  - "[Testing Overview](../18-testing/01-testing-overview.md)"
---

# Middleware Testing

## Testing Authenticate Middleware

```go
// internal/platform/middleware/authenticate_test.go
package middleware_test

import (
    "net/http/httptest"
    "testing"

    "github.com/gofiber/fiber/v2"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "awo.so/internal/platform/middleware"
    "awo.so/internal/testutil"
)

func TestAuthenticate_MissingHeader_Returns401(t *testing.T) {
    app := fiber.New()
    app.Get("/test", middleware.Authenticate(&stubSessionSvc{}), func(c *fiber.Ctx) error {
        return c.SendStatus(fiber.StatusOK)
    })

    req := httptest.NewRequest("GET", "/test", nil)
    // No Authorization header

    resp, err := app.Test(req)
    require.NoError(t, err)
    assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAuthenticate_ValidToken_SetsSession(t *testing.T) {
    expectedSession := testutil.TestSession()
    sessionSvc := &stubSessionSvc{session: expectedSession}

    app := fiber.New()
    app.Get("/test",
        middleware.Authenticate(sessionSvc),
        func(c *fiber.Ctx) error {
            session := middleware.SessionFrom(c)
            assert.Equal(t, expectedSession.TenantID, session.TenantID)
            return c.SendStatus(fiber.StatusOK)
        },
    )

    req := httptest.NewRequest("GET", "/test", nil)
    req.Header.Set("Authorization", "Bearer valid-token")

    resp, err := app.Test(req)
    require.NoError(t, err)
    assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestAuthenticate_InvalidToken_Returns401(t *testing.T) {
    sessionSvc := &stubSessionSvc{err: iam.ErrSessionInvalid}

    app := fiber.New()
    app.Get("/test", middleware.Authenticate(sessionSvc), func(c *fiber.Ctx) error {
        return c.SendStatus(fiber.StatusOK)
    })

    req := httptest.NewRequest("GET", "/test", nil)
    req.Header.Set("Authorization", "Bearer invalid-token")

    resp, err := app.Test(req)
    require.NoError(t, err)
    assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
```

## Testing Authorize Middleware

```go
func TestAuthorize_Allowed_PassesThrough(t *testing.T) {
    authzSvc := &stubAuthzSvc{allowed: true}
    sessionSvc := &stubSessionSvc{session: testutil.TestSession()}

    app := fiber.New()
    app.Get("/test",
        middleware.Authenticate(sessionSvc),
        middleware.Authorize(authzSvc, "contracts.contract.read"),
        func(c *fiber.Ctx) error {
            return c.SendStatus(fiber.StatusOK)
        },
    )

    req := httptest.NewRequest("GET", "/test", nil)
    req.Header.Set("Authorization", "Bearer valid-token")

    resp, err := app.Test(req)
    require.NoError(t, err)
    assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestAuthorize_Denied_Returns403(t *testing.T) {
    authzSvc := &stubAuthzSvc{allowed: false}
    // ...

    resp, err := app.Test(req)
    require.NoError(t, err)
    assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
```

## Stubs

```go
type stubSessionSvc struct {
    session *domain.ResolvedSession
    err     error
}

func (s *stubSessionSvc) Resolve(_ context.Context, _ string) (*domain.ResolvedSession, error) {
    return s.session, s.err
}

type stubAuthzSvc struct{ allowed bool }

func (s *stubAuthzSvc) Enforce(_ context.Context, _ iam.Request) (bool, error) {
    return s.allowed, nil
}
```
