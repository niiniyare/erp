package contract

import (
	"github.com/gofiber/fiber/v2"

	"awo.so/internal/core/iam"
)

// InjectSessionContext is a Fiber middleware that bridges the IAM authentication
// middleware and service-layer code.
//
// It reads the [iam.ResolvedSession] that [middleware.Authenticate] stored in
// Fiber Locals and injects a [SessionContext] into the Go request context, so
// that service methods can call [FromContext] without depending on *fiber.Ctx.
//
//	Route chain example:
//	  app.Use(middleware.Authenticate(cfg))
//	  app.Use(contract.InjectSessionContext())
//	  app.Get("/invoices", invoiceHandler.List)
//
//	In invoiceHandler.List (or any downstream service):
//	  sc, ok := contract.FromContext(ctx)
//	  tenantID := sc.TenantID()
//
// Must be placed AFTER [middleware.Authenticate]. Returns 401 if no session
// is present in Fiber Locals.
func InjectSessionContext() fiber.Handler {
	return func(c *fiber.Ctx) error {
		resolved, ok := c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession)
		if !ok || resolved == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
		}
		sc := newSessionContext(resolved)
		ctx := WithContext(c.UserContext(), sc)
		c.SetUserContext(ctx)
		return c.Next()
	}
}
