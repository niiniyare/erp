package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// RequestID injects a request-scoped UUID into X-Request-ID response header
// and c.Locals("request_id"). Uses the incoming X-Request-ID value when present
// so clients can correlate requests across service boundaries.
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Get("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}
		c.Set("X-Request-ID", id)
		c.Locals("request_id", id)
		return c.Next()
	}
}

// RequestLogger logs one structured line per request using zerolog.
// Emits at Info level for 2xx/3xx, Warn for 4xx, Error for 5xx.
// Must be placed AFTER RequestID() in the middleware chain.
//
// Log fields:
//
//	request_id, method, path, status, duration_ms, tenant_id (when set)
func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		status := c.Response().StatusCode()
		ms := time.Since(start).Milliseconds()
		reqID, _ := c.Locals("request_id").(string)

		event := selectLevel(status)
		ev := event.
			Str("request_id", reqID).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", status).
			Int64("duration_ms", ms)

		// Attach tenant_id when available (set by downstream middleware or viewer).
		if tid, ok := c.Locals("tenant_id").(string); ok && tid != "" {
			ev = ev.Str("tenant_id", tid)
		}
		if uid, ok := c.Locals("user_id").(string); ok && uid != "" {
			ev = ev.Str("user_id", uid)
		}

		if err != nil {
			ev.Err(err).Msg("request")
		} else {
			ev.Msg("request")
		}

		return err
	}
}

// TenantHeader extracts the X-Awo-Tenant header and stores it in
// c.Locals("tenant_slug") for downstream middleware and handlers.
// Does NOT validate or resolve the tenant — that happens in extractContext.
func TenantHeader() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if v := c.Get("X-Awo-Tenant"); v != "" {
			c.Locals("tenant_slug", v)
		}
		return c.Next()
	}
}

// selectLevel returns a zerolog event at the appropriate level for status.
func selectLevel(status int) *zerolog.Event {
	switch {
	case status >= 500:
		return log.Error()
	case status >= 400:
		return log.Warn()
	default:
		return log.Info()
	}
}
