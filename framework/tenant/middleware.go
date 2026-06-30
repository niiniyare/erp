package tenant

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// globalPrefixes are path prefixes that do not require a tenant context.
// Requests to these paths bypass tenant resolution.
var globalPrefixes = []string{
	"/health",
	"/metrics",
	"/api/v1/platform/",
}

// ResolutionMiddleware resolves the tenant for every incoming request.
// Resolution priority (first non-empty value wins):
//
//  1. X-Awo-Tenant header (preferred — UUID or slug)
//  2. X-Tenant-ID header (deprecated — logs a warning)
//  3. tenant query parameter (webhooks / legacy)
//  4. Subdomain (browser access: {slug}.example.com)
//
// On success, stores the raw tenant value in c.Locals("tenant_slug") and,
// if the value is a valid UUID, also in c.Locals("tenant_id") as uuid.UUID.
// Actual UUID resolution from slug happens in the handler via TenantResolver.
//
// Returns 400 when a tenant is required but not provided.
// Global routes (health, metrics, platform admin) skip enforcement.
func ResolutionMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if isGlobalPath(c.Path()) {
			return c.Next()
		}

		tenant := c.Get("X-Awo-Tenant")

		if tenant == "" {
			if v := c.Get("X-Tenant-ID"); v != "" {
				tenant = v
				log.Warn().
					Str("request_id", safeRequestID(c)).
					Str("client_ip", c.IP()).
					Msg("deprecated X-Tenant-ID header used; migrate to X-Awo-Tenant")
			}
		}

		if tenant == "" {
			tenant = c.Query("tenant")
		}

		if tenant == "" {
			host := c.Hostname()
			// Strip port if present.
			if idx := strings.LastIndexByte(host, ':'); idx > 0 {
				host = host[:idx]
			}
			// Require 3+ parts (subdomain.domain.tld) to avoid treating bare
			// two-part hostnames like "example.com" as tenant subdomains.
			parts := strings.Split(host, ".")
			if len(parts) >= 3 {
				sub := parts[0]
				if sub != "www" && sub != "api" && sub != "app" && sub != "bo" && sub != "portal" {
					tenant = sub
				}
			}
		}

		if tenant == "" {
			return c.Status(fiber.StatusBadRequest).JSON(
				newStatusErr("tenant_required", "X-Awo-Tenant header is required"))
		}

		// Store raw value so downstream (extractContext / TenantResolver) can use it.
		c.Locals("tenant_slug", tenant)

		// If the raw value is already a UUID, pre-populate tenant_id.
		if id, err := uuid.Parse(tenant); err == nil {
			c.Locals("tenant_id", id)
		}

		return c.Next()
	}
}

func isGlobalPath(path string) bool {
	for _, prefix := range globalPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func safeRequestID(c *fiber.Ctx) string {
	if v, ok := c.Locals("request_id").(string); ok {
		return v
	}
	return ""
}
