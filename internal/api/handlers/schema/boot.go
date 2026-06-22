// Package schema provides the GET /schema/boot handler.
//
// The boot endpoint returns an AMIS "app" shell schema — the sidebar navigation
// tree filtered in real time by:
//
//  1. Feature flag: sess.Configuration.Flags["{module}.enabled"] must be true
//     (or the key must be absent, which means "not explicitly disabled").
//  2. Permission: checker.Check(principal, "{module}.{resource}", "read") must return true.
//
// The feature flag check is fast (in-memory). The permission check delegates to
// contract.PolicyChecker which wraps AuthzService — direct .Enforce() calls are
// forbidden in handler code by the contract boundary guard.
package schema

import (
	"github.com/gofiber/fiber/v2"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/iam/contract"
)

// BootHandler returns the AMIS app shell schema for the authenticated user.
//
// Route: GET /schema/boot  (requires Authenticate middleware)
func BootHandler(store db.Store, checker contract.PolicyChecker) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, ok := c.Locals(contract.LocalsKeySession).(*contract.ResolvedSession)
		if !ok || sess == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
		}
		principal, _ := c.Locals(contract.LocalsKeyPrincipal).(contract.Principal)

		modules, err := store.ListActiveSystemModules(c.Context())
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to load modules")
		}

		pages := make([]fiber.Map, 0, len(modules))

		for _, mod := range modules {
			// 1. Feature flag gate: if the flag key exists and is false, skip.
			//    An absent flag key means "not explicitly disabled" → allowed.
			if enabled, exists := sess.Configuration.Flags[mod.Slug+".enabled"]; exists && !enabled {
				continue
			}

			resources, err := store.ListActiveResourcesByModule(c.Context(), mod.ID)
			if err != nil {
				// Non-fatal: skip this module rather than failing the whole boot
				continue
			}

			children := make([]fiber.Map, 0, len(resources))
			for _, res := range resources {
				// 2. Permission gate: user must have at least read on this resource.
				//    checker may be nil when authz is not configured (dev/test).
				if checker != nil {
					allowed, err := checker.Check(c.Context(), contract.Request{
						Subject: principal.Subject,
						Domain:  principal.Domain,
						Object:  mod.Slug + "." + res.Slug,
						Action:  "read",
					})
					if err != nil || !allowed {
						continue
					}
				}

				child := fiber.Map{
					"label": strVal(res.DisplayName, res.Name),
					"url":   strVal(res.NavUrl, ""),
					"icon":  strVal(res.Icon, ""),
				}
				children = append(children, child)
			}

			if len(children) == 0 {
				continue
			}

			pages = append(pages, fiber.Map{
				"label":    strVal(mod.DisplayName, mod.Name),
				"icon":     strVal(mod.Icon, ""),
				"children": children,
			})
		}

		return c.JSON(fiber.Map{
			"type":      "app",
			"brandName": "ERP",
			"pages":     pages,
		})
	}
}

// strVal returns *s if non-nil, otherwise fallback.
func strVal(s *string, fallback string) string {
	if s != nil {
		return *s
	}
	return fallback
}
