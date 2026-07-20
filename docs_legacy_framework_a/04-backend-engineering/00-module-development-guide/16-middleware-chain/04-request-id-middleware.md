> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Request ID and Logging Middleware
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Middleware Overview](01-middleware-overview.md)"
  - "[Structured Logging Guide](../19-observability/02-structured-logging-guide.md)"
---

# Request ID and Logging Middleware

## Request ID Middleware

Assigns a UUID to every request. If the client sends `X-Request-ID`, use it (for end-to-end tracing). Otherwise generate one.

```go
// internal/platform/middleware/requestid.go
package middleware

import (
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
)

const LocalsKeyRequestID = "request_id"

func RequestID() fiber.Handler {
    return func(c *fiber.Ctx) error {
        requestID := c.Get("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }
        c.Locals(LocalsKeyRequestID, requestID)
        c.Set("X-Request-ID", requestID)
        return c.Next()
    }
}
```

## Logger Middleware

Structured request log with zerolog:

```go
// internal/platform/middleware/logger.go
package middleware

import (
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/rs/zerolog"
)

func Logger(log zerolog.Logger) fiber.Handler {
    return func(c *fiber.Ctx) error {
        start := time.Now()

        err := c.Next()

        status := c.Response().StatusCode()
        level := zerolog.InfoLevel
        if status >= 500 {
            level = zerolog.ErrorLevel
        } else if status >= 400 {
            level = zerolog.WarnLevel
        }

        log.WithLevel(level).
            Str("method", c.Method()).
            Str("path", c.Path()).
            Int("status", status).
            Dur("latency_ms", time.Since(start)).
            Str("request_id", requestIDFrom(c)).
            Str("remote_ip", c.IP()).
            Msg("request")

        return err
    }
}

func requestIDFrom(c *fiber.Ctx) string {
    if id, ok := c.Locals(LocalsKeyRequestID).(string); ok {
        return id
    }
    return ""
}
```

## Recover Middleware

Catches handler panics and returns a 500 without exposing internal details:

```go
func Recover(log zerolog.Logger) fiber.Handler {
    return func(c *fiber.Ctx) error {
        defer func() {
            if r := recover(); r != nil {
                log.Error().
                    Interface("panic", r).
                    Str("path", c.Path()).
                    Str("request_id", requestIDFrom(c)).
                    Msg("handler panicked")

                _ = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                    "error": fiber.Map{
                        "code":    "INTERNAL_ERROR",
                        "message": "An unexpected error occurred",
                    },
                })
            }
        }()
        return c.Next()
    }
}
```

## Global Setup

```go
// cmd/server/main.go or internal/server/server.go
func setupApp(log zerolog.Logger) *fiber.App {
    app := fiber.New(fiber.Config{
        ErrorHandler: customErrorHandler(log),
    })

    app.Use(middleware.Recover(log))
    app.Use(middleware.RequestID())
    app.Use(middleware.Logger(log))
    app.Use(middleware.RateLimit())
    app.Use(middleware.CORS())

    return app
}
```

## Custom Error Handler

Converts `*fiber.Error` to the standard error envelope:

```go
func customErrorHandler(log zerolog.Logger) fiber.ErrorHandler {
    return func(c *fiber.Ctx, err error) error {
        code := fiber.StatusInternalServerError
        message := "An unexpected error occurred"

        var fiberErr *fiber.Error
        if errors.As(err, &fiberErr) {
            code    = fiberErr.Code
            message = fiberErr.Message
        }

        if code >= 500 {
            log.Error().Err(err).Str("path", c.Path()).Msg("5xx error")
        }

        return c.Status(code).JSON(fiber.Map{
            "error": fiber.Map{
                "code":    httpCodeToErrorCode(code),
                "message": message,
            },
        })
    }
}

func httpCodeToErrorCode(code int) string {
    switch code {
    case 400: return "BAD_REQUEST"
    case 401: return "UNAUTHORIZED"
    case 403: return "FORBIDDEN"
    case 404: return "NOT_FOUND"
    case 409: return "CONFLICT"
    case 422: return "VALIDATION_ERROR"
    case 429: return "RATE_LIMITED"
    default:  return "INTERNAL_ERROR"
    }
}
```
