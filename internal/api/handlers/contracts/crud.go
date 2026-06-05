package contracts

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"awo.so/internal/core/contracts"
)

// Create handles POST /api/v1/contracts.
// Body: CreateContractRequest (JSON).
// Returns 201 Created with the new contract.
func (h *Handler) Create(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.UserContext(), "contracts.handler.Create")
	defer span.End()

	var req contracts.CreateContractRequest
	if err := h.validateRequest(c, &req); err != nil {
		return h.handleError(c, err)
	}

	contract, err := h.svc.Create(ctx, req)
	if err != nil {
		return h.handleError(c, err)
	}

	return h.created(c, contract)
}

// GetByID handles GET /api/v1/contracts/:id.
// Returns 200 OK with the contract, or 404 if not found.
func (h *Handler) GetByID(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.UserContext(), "contracts.handler.GetByID")
	defer span.End()

	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return h.handleError(c, err)
	}

	contract, err := h.svc.GetByID(ctx, id)
	if err != nil {
		return h.handleError(c, err)
	}

	return h.success(c, contract)
}

// GetByNumber handles GET /api/v1/contracts/number/:number.
// Returns 200 OK with the contract, or 404 if not found.
func (h *Handler) GetByNumber(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.UserContext(), "contracts.handler.GetByNumber")
	defer span.End()

	number := c.Params("number")
	if number == "" {
		return h.handleError(c, newBadRequest("contract number is required"))
	}

	contract, err := h.svc.GetByNumber(ctx, number)
	if err != nil {
		return h.handleError(c, err)
	}

	return h.success(c, contract)
}

// Update handles PATCH /api/v1/contracts/:id.
// Body: UpdateContractRequest (JSON). Only non-nil fields are applied.
// Returns 200 OK with the updated contract.
func (h *Handler) Update(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.UserContext(), "contracts.handler.Update")
	defer span.End()

	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return h.handleError(c, err)
	}

	var req contracts.UpdateContractRequest
	if err := h.validateRequest(c, &req); err != nil {
		return h.handleError(c, err)
	}

	contract, err := h.svc.Update(ctx, id, req)
	if err != nil {
		return h.handleError(c, err)
	}

	return h.success(c, contract)
}

// Delete handles DELETE /api/v1/contracts/:id.
// Soft-deletes the contract. Returns 204 No Content.
func (h *Handler) Delete(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.UserContext(), "contracts.handler.Delete")
	defer span.End()

	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return h.handleError(c, err)
	}

	if err := h.svc.Delete(ctx, id); err != nil {
		return h.handleError(c, err)
	}

	return h.noContent(c)
}

// List handles GET /api/v1/contracts.
// Supports query params: entity_id, status, contract_type, search, limit, offset.
// Returns 200 OK with paginated results.
func (h *Handler) List(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.UserContext(), "contracts.handler.List")
	defer span.End()

	filter, err := parseContractFilter(c)
	if err != nil {
		return h.handleError(c, err)
	}

	items, total, err := h.svc.List(ctx, filter)
	if err != nil {
		return h.handleError(c, err)
	}

	return h.successWithMeta(c, items, fiber.Map{
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// ── private helpers ────────────────────────────────────────────────────────────

// parseUUIDParam extracts and parses a named UUID path parameter.
func parseUUIDParam(c *fiber.Ctx, param string) (uuid.UUID, error) {
	raw := c.Params(param)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, newBadRequest("invalid " + param + ": must be a valid UUID")
	}
	return id, nil
}

// parseContractFilter builds a ContractFilter from request query parameters.
func parseContractFilter(c *fiber.Ctx) (contracts.ContractFilter, error) {
	filter := contracts.ContractFilter{
		Limit:  20,
		Offset: 0,
	}

	// entity_id — optional UUID
	if raw := c.Query("entity_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return filter, newBadRequest("invalid entity_id: must be a valid UUID")
		}
		filter.EntityID = &id
	}

	// status — optional string
	if s := c.Query("status"); s != "" {
		filter.Status = &s
	}

	// contract_type — optional string
	if ct := c.Query("contract_type"); ct != "" {
		filter.ContractType = &ct
	}

	// search — optional free-text
	if q := c.Query("search"); q != "" {
		filter.Search = &q
	}

	// limit — default 20, max 100
	if raw := c.Query("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			return filter, newBadRequest("limit must be a positive integer")
		}
		if n > 100 {
			n = 100
		}
		filter.Limit = int32(n)
	}

	// offset — default 0
	if raw := c.Query("offset"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 {
			return filter, newBadRequest("offset must be a non-negative integer")
		}
		filter.Offset = int32(n)
	}

	return filter, nil
}

// newBadRequest returns a minimal HTTP 400 error for quick parameter validation.
func newBadRequest(msg string) error {
	return fiber.NewError(http.StatusBadRequest, msg)
}
