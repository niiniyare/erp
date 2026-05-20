package finance

// cost_centers.go — HTTP handlers for cost centre hierarchy management.
//
// Routes (registered in routes.go):
//
//	POST   /api/v1/finance/cost-centers       → CreateCostCenter
//	GET    /api/v1/finance/cost-centers        → ListCostCenters
//	GET    /api/v1/finance/cost-centers/:id    → GetCostCenter
//	PUT    /api/v1/finance/cost-centers/:id    → UpdateCostCenter
//	DELETE /api/v1/finance/cost-centers/:id    → DeleteCostCenter
//
// Cost centres form a tree via parent_id. Group nodes (is_group=true) cannot
// receive transactions directly — only leaf nodes can.

import (
	"github.com/gofiber/fiber/v2"

	financeDomain "awo.so/internal/core/finance/domain"
)

// ============================================================================
// Cost centre request type
//
// Shared between CreateCostCenter and UpdateCostCenter — the fields are
// identical; the difference is the HTTP verb and whether an ID exists.
// ============================================================================

// costCenterRequest is the body for POST and PUT cost-center endpoints.
type costCenterRequest struct {
	Code        string  `json:"code"         validate:"required,max=50"`
	Name        string  `json:"name"         validate:"required,max=200"`
	Description *string `json:"description"`
	// ParentID links this node to a parent in the cost centre tree.
	// Omit for root-level cost centres.
	ParentID         *string `json:"parent_id"          validate:"omitempty,uuid"`
	IsGroup          bool    `json:"is_group"`
	IsDistributed    bool    `json:"is_distributed"`
	AllocationMethod *string `json:"allocation_method"`
	IsActive         bool    `json:"is_active"`
}

// toDomain converts the request into a financeDomain.CostCenter.
// The caller sets ID separately (empty for creates, path param for updates).
func (r *costCenterRequest) toDomain() (*financeDomain.CostCenter, error) {
	cc := &financeDomain.CostCenter{
		Code:          r.Code,
		Name:          r.Name,
		Description:   r.Description,
		IsGroup:       r.IsGroup,
		IsDistributed: r.IsDistributed,
		IsActive:      r.IsActive,
	}
	if r.ParentID != nil {
		id, err := parseUUID(*r.ParentID, "parent_id")
		if err != nil {
			return nil, err
		}
		cc.ParentID = &id
	}
	if r.AllocationMethod != nil {
		m := financeDomain.AllocationMethod(*r.AllocationMethod)
		cc.AllocationMethod = &m
	}
	return cc, nil
}

// ============================================================================
// Cost centre handlers
// ============================================================================

// CreateCostCenter creates a new cost centre node.
// If parent_id is provided the node is attached to that parent in the tree.
//
// POST /api/v1/finance/cost-centers
// Permission: finance.cost_centers.create
func (h *FinanceHandler) CreateCostCenter(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.CreateCostCenter")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	var req costCenterRequest
	if err := h.bind(c, &req); err != nil {
		return h.fail(c, err)
	}
	cc, err := req.toDomain()
	if err != nil {
		return h.fail(c, err)
	}

	// 3. Delegate
	created, err := h.services.CostCenter.Create(ctx, cc)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok201(c, created)
}

// GetCostCenter retrieves a single cost centre by ID.
//
// GET /api/v1/finance/cost-centers/:id
// Permission: finance.cost_centers.read
func (h *FinanceHandler) GetCostCenter(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.GetCostCenter")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	ccID, err := parseUUID(c.Params("id"), "cost centre ID")
	if err != nil {
		return h.fail(c, err)
	}

	// 3. Delegate
	cc, err := h.services.CostCenter.GetByID(ctx, ccID)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, cc)
}

// ListCostCenters returns cost centres for the current tenant.
//
// GET /api/v1/finance/cost-centers?active_only=true
// Permission: finance.cost_centers.read
func (h *FinanceHandler) ListCostCenters(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.ListCostCenters")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse — single optional bool query param
	activeOnly := c.QueryBool("active_only", false)

	// 3. Delegate
	ccs, err := h.services.CostCenter.List(ctx, activeOnly)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, ccs)
}

// UpdateCostCenter replaces a cost centre's mutable fields.
// The cost centre code is immutable after creation.
//
// PUT /api/v1/finance/cost-centers/:id
// Permission: finance.cost_centers.update
func (h *FinanceHandler) UpdateCostCenter(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.UpdateCostCenter")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	ccID, err := parseUUID(c.Params("id"), "cost centre ID")
	if err != nil {
		return h.fail(c, err)
	}

	var req costCenterRequest
	if err := h.bind(c, &req); err != nil {
		return h.fail(c, err)
	}
	cc, err := req.toDomain()
	if err != nil {
		return h.fail(c, err)
	}
	// Path param ID is authoritative
	cc.ID = ccID

	// 3. Delegate
	updated, err := h.services.CostCenter.Update(ctx, cc)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, updated)
}

// DeleteCostCenter removes a cost centre.
// Cost centres that have child nodes or posted transactions cannot be deleted
// (enforced by service layer).
//
// DELETE /api/v1/finance/cost-centers/:id
// Permission: finance.cost_centers.delete
func (h *FinanceHandler) DeleteCostCenter(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.DeleteCostCenter")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	ccID, err := parseUUID(c.Params("id"), "cost centre ID")
	if err != nil {
		return h.fail(c, err)
	}

	// 3. Delegate
	if err := h.services.CostCenter.Delete(ctx, ccID); err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. 204 — no body on delete (consistent with DeleteAccount)
	return h.ok204(c)
}
