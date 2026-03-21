package authz

import (
	"github.com/gofiber/fiber/v2"
)

// Middleware returns a Fiber handler factory that enforces object+action for
// the authenticated Principal stored at c.Locals(LocalsKeyPrincipal).
//
// Usage in routes:
//
//	app.Get("/invoices",        svc.Middleware("invoice", "read"),   listInvoices)
//	app.Post("/invoices",       svc.Middleware("invoice", "create"), createInvoice)
//	app.Delete("/invoices/:id", svc.Middleware("invoice", "delete"), deleteInvoice)
//
// If a route has an ":id" param, the object becomes "invoice/{id}" automatically,
// enabling per-resource wildcard policies like "invoice/*".
func (s *service) Middleware(object, action string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		p, ok := c.Locals(LocalsKeyPrincipal).(Principal)
		if !ok || p.Subject == "" {
			return fiber.NewError(fiber.StatusUnauthorized, ErrUnauthorized.Error())
		}

		obj := object
		if id := c.Params("id"); id != "" {
			obj = object + "/" + id
		}

		allowed, err := s.Enforce(c.Context(), Request{
			Subject: p.Subject,
			Domain:  p.Domain,
			Object:  obj,
			Action:  action,
		})
		if err != nil {
			return err
		}
		if !allowed {
			return fiber.NewError(fiber.StatusForbidden, ErrForbidden.Error())
		}
		return c.Next()
	}
}
