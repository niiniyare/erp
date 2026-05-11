// Package schema provides the GET /schema/boot handler.
//
// The boot endpoint returns an AMIS "app" shell schema — the sidebar navigation
// tree filtered in real time by:
//
//  1. Feature flag: sess.Configuration.Flags["{module}.enabled"] must be true
//     (or the key must be absent, which means "not explicitly disabled").
//  2. Permission: authzSvc.Enforce(principal, "{module}.{resource}", "read") must return true.
//
// The feature flag check is fast (in-memory). The permission check calls Casbin
// in-process (no extra DB round-trips per resource beyond the two module/resource queries).
package schema

import (
	"github.com/gofiber/fiber/v2"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/iam"
)

// BootHandler returns the AMIS app shell schema for the authenticated user.
//
// Route: GET /schema/boot  (requires Authenticate middleware)
func BootHandler(store db.Store, authzSvc iam.AuthzService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, ok := c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession)
		if !ok || sess == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
		}
		principal, _ := c.Locals(iam.LocalsKeyPrincipal).(iam.Principal)

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
				allowed, err := authzSvc.Enforce(c.Context(), iam.Request{
					Subject: principal.Subject,
					Domain:  principal.Domain,
					Object:  mod.Slug + "." + res.Slug,
					Action:  "read",
				})
				if err != nil || !allowed {
					continue
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
