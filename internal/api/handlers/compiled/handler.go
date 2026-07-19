// Package compiled provides a metadata-driven HTTP handler that serves CRUD
// and action routes for any entity registered via awo EntityDefinition.
//
// One Handler is created per EntitySchema at startup. All routing decisions —
// permission namespaces, operation labels, field validation — derive from the
// compiler output, so adding a new EntityDefinition never requires a new
// handler file.
package compiled

import (
	"github.com/gofiber/fiber/v2"

	"awo.so/awo/compiler"
)

// Handler is a generic entity CRUD handler bound to a single EntitySchema.
// It is stateless beyond the schema reference and safe for concurrent use.
type Handler struct {
	schema *compiler.EntitySchema
}

// New returns a Handler bound to es.
func New(es *compiler.EntitySchema) *Handler {
	return &Handler{schema: es}
}

// List handles GET /api/v1/{module}/{resource}
func (h *Handler) List(c *fiber.Ctx) error {
	// TODO: delegate to EntityRepository.Query scoped to h.schema.QualifiedName.
	// Apply Filter DSL from query params; honour h.schema.SearchableFields for ?q=.
	// Respect privacy policy injected into context by the auth middleware.
	return fiber.NewError(fiber.StatusNotImplemented,
		h.schema.QualifiedName+".list: auto-generated handler not yet wired to repository")
}

// Get handles GET /api/v1/{module}/{resource}/:id
func (h *Handler) Get(c *fiber.Ctx) error {
	// TODO: delegate to EntityRepository.Get(ctx, id).
	// Strip h.schema.SensitiveFields from the response.
	return fiber.NewError(fiber.StatusNotImplemented,
		h.schema.QualifiedName+".get: auto-generated handler not yet wired to repository")
}

// Create handles POST /api/v1/{module}/{resource}
func (h *Handler) Create(c *fiber.Ctx) error {
	// TODO: parse body; enforce h.schema.RequiredFields; reject h.schema.ImmutableFields
	// on create; delegate to EntityRepository.Create.
	return fiber.NewError(fiber.StatusNotImplemented,
		h.schema.QualifiedName+".create: auto-generated handler not yet wired to repository")
}

// Update handles PATCH /api/v1/{module}/{resource}/:id
func (h *Handler) Update(c *fiber.Ctx) error {
	// TODO: parse body; reject mutations to h.schema.ImmutableFields;
	// delegate to EntityRepository.Update.
	return fiber.NewError(fiber.StatusNotImplemented,
		h.schema.QualifiedName+".update: auto-generated handler not yet wired to repository")
}

// Delete handles DELETE /api/v1/{module}/{resource}/:id
func (h *Handler) Delete(c *fiber.Ctx) error {
	// TODO: delegate to EntityRepository.Delete(ctx, id).
	return fiber.NewError(fiber.StatusNotImplemented,
		h.schema.QualifiedName+".delete: auto-generated handler not yet wired to repository")
}

// Action handles POST /api/v1/{module}/{resource}/:id/{action}
// actionName must match one of h.schema.ActionsByName.
func (h *Handler) Action(actionName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: look up actionDef in h.schema.ActionsByName[actionName];
		// call actionDef.HandlerFunc with a pre-resolved ActionContext.
		return fiber.NewError(fiber.StatusNotImplemented,
			h.schema.QualifiedName+".action."+actionName+": auto-generated handler not yet wired")
	}
}
