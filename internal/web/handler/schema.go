// Package handler provides the HTTP handler that serves AMIS page schemas.
// One handler, zero boilerplate per new page. Register your schema function
// in the registry package and it's automatically served.
package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"awo.so/internal/shared"
	"awo.so/internal/web/amis"
	"awo.so/internal/web/registry"
)

// SchemaHandler serves AMIS page schemas from the registry.
// Mount it at /schema/* — any sub-path is looked up in the registry.
//
//	app.Get("/schema/*", schemaHandler.Handle)
type SchemaHandler struct{}

// NewSchemaHandler returns a new SchemaHandler.
func NewSchemaHandler() *SchemaHandler {
	return &SchemaHandler{}
}

// Handle serves the schema for the requested path.
// The path is everything after /schema, e.g. /schema/finance/invoices → /finance/invoices.
func (h *SchemaHandler) Handle(c *fiber.Ctx) error {
	path := "/" + c.Params("*")

	fn := registry.Get(path)
	if fn == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status": 404,
			"msg":    "schema not found: " + path,
		})
	}

	ctx := buildCtx(c)
	schema := fn(ctx)

	// Return AMIS envelope: { status: 0, data: <schema> }
	return c.JSON(fiber.Map{
		"status": 0,
		"data":   schema,
	})
}

// buildCtx extracts user/tenant/permission data from the Fiber request context
// and packages it into an amis.Ctx for schema functions.
func buildCtx(c *fiber.Ctx) amis.Ctx {
	// UserContext() returns the standard context.Context set by middleware.
	// Falls back to background context when running without auth middleware (dev UI server).
	goCtx := c.UserContext()
	if goCtx == nil {
		goCtx = context.Background()
	}

	userID, _ := shared.GetUserID(goCtx)
	tenantID, _ := shared.GetTenantID(goCtx)

	// CapabilityContext carries non-auth identity metadata (role label, scope).
	cap, _ := shared.GetCapabilityContext(goCtx)

	ctx := amis.Ctx{
		// Feature flags are evaluated per-feature via contract.SessionContext.FeatureEnabled;
		// a bulk flags map is not exposed here to prevent schema functions from making
		// authorization decisions.
		Flags: nil,
		User: amis.CtxUser{
			ID:       userID.String(),
			TenantID: tenantID.String(),
			Role:     cap.Role,
		},
		// Can always returns false — authorization is enforced by middleware.Authorize
		// at route registration, not inside AMIS schema functions.
		Can: func(action, resource string) bool {
			return false
		},
	}

	// Tenant name may come from a request local set by the tenant middleware.
	if name, ok := c.Locals("tenant_name").(string); ok {
		ctx.User.TenantName = name
	}

	return ctx
}
