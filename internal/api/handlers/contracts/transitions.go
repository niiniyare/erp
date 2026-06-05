package contracts

import (
	"github.com/gofiber/fiber/v2"
)

// versionBody is the minimal JSON body expected by lifecycle action endpoints.
// All lifecycle transitions require the current contract version for optimistic locking.
type versionBody struct {
	Version int32 `json:"version" validate:"required,min=1"`
}

// terminateBody extends versionBody with an optional termination reason.
type terminateBody struct {
	Version int32  `json:"version"          validate:"required,min=1"`
	Reason  string `json:"reason,omitempty"`
}

// Submit handles POST /api/v1/contracts/:id/submit.
// Transitions the contract from DRAFT → PENDING_APPROVAL.
// Body: { "version": <int> }
func (h *Handler) Submit(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.UserContext(), "contracts.handler.Submit")
	defer span.End()

	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return h.handleError(c, err)
	}

	var body versionBody
	if err := h.validateRequest(c, &body); err != nil {
		return h.handleError(c, err)
	}

	if err := h.svc.Submit(ctx, id, body.Version); err != nil {
		return h.handleError(c, err)
	}

	return h.noContent(c)
}

// Approve handles POST /api/v1/contracts/:id/approve.
// Transitions the contract from PENDING_APPROVAL → ACTIVE.
// The signer is taken from the authenticated session — not from the request body.
// Body: { "version": <int> }
func (h *Handler) Approve(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.UserContext(), "contracts.handler.Approve")
	defer span.End()

	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return h.handleError(c, err)
	}

	var body versionBody
	if err := h.validateRequest(c, &body); err != nil {
		return h.handleError(c, err)
	}

	if err := h.svc.Approve(ctx, id, body.Version); err != nil {
		return h.handleError(c, err)
	}

	return h.noContent(c)
}

// Reject handles POST /api/v1/contracts/:id/reject.
// Transitions the contract from PENDING_APPROVAL → DRAFT.
// Body: { "version": <int> }
func (h *Handler) Reject(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.UserContext(), "contracts.handler.Reject")
	defer span.End()

	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return h.handleError(c, err)
	}

	var body versionBody
	if err := h.validateRequest(c, &body); err != nil {
		return h.handleError(c, err)
	}

	if err := h.svc.Reject(ctx, id, body.Version); err != nil {
		return h.handleError(c, err)
	}

	return h.noContent(c)
}

// Terminate handles POST /api/v1/contracts/:id/terminate.
// Transitions the contract from ACTIVE → TERMINATED.
// Body: { "version": <int>, "reason": "<string>" }
func (h *Handler) Terminate(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.UserContext(), "contracts.handler.Terminate")
	defer span.End()

	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return h.handleError(c, err)
	}

	var body terminateBody
	if err := h.validateRequest(c, &body); err != nil {
		return h.handleError(c, err)
	}

	if err := h.svc.Terminate(ctx, id, body.Reason, body.Version); err != nil {
		return h.handleError(c, err)
	}

	return h.noContent(c)
}

// Routes registers all contract HTTP endpoints on the given router group.
// Call this from routes.go to wire up /api/v1/contracts/*.
func (h *Handler) Routes(r fiber.Router) {
	contracts := r.Group("/contracts")

	// CRUD
	contracts.Post("/", h.Create)                     // POST   /api/v1/contracts
	contracts.Get("/", h.List)                        // GET    /api/v1/contracts
	contracts.Get("/number/:number", h.GetByNumber)   // GET    /api/v1/contracts/number/:number
	contracts.Get("/:id", h.GetByID)                  // GET    /api/v1/contracts/:id
	contracts.Patch("/:id", h.Update)                 // PATCH  /api/v1/contracts/:id
	contracts.Delete("/:id", h.Delete)                // DELETE /api/v1/contracts/:id

	// Lifecycle transitions
	contracts.Post("/:id/submit", h.Submit)       // POST /api/v1/contracts/:id/submit
	contracts.Post("/:id/approve", h.Approve)     // POST /api/v1/contracts/:id/approve
	contracts.Post("/:id/reject", h.Reject)       // POST /api/v1/contracts/:id/reject
	contracts.Post("/:id/terminate", h.Terminate) // POST /api/v1/contracts/:id/terminate
}
