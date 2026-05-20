package finance

// budgets.go — HTTP handlers for budget management.
//
// Routes (registered in routes.go):
//
//	POST  /api/v1/finance/budgets                → CreateBudget
//	GET   /api/v1/finance/budgets                → ListBudgets
//	GET   /api/v1/finance/budgets/:id            → GetBudget
//	GET   /api/v1/finance/budgets/:id/lines      → GetBudgetLines
//	POST  /api/v1/finance/budgets/:id/submit     → SubmitBudget
//	POST  /api/v1/finance/budgets/:id/approve    → ApproveBudget
//	POST  /api/v1/finance/budgets/:id/reject     → RejectBudget
//	POST  /api/v1/finance/budgets/:id/close      → CloseBudget
//
// Budget state machine (enforced by BudgetService):
//
//	DRAFT → submit → PENDING_APPROVAL → approve → APPROVED → close → CLOSED
//	                                  → reject  → REJECTED → revise → DRAFT

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	financeDomain "awo.so/internal/core/finance/domain"
)

// ============================================================================
// Budget request types
// ============================================================================

// budgetLineItemRequest represents a single budget line in CreateBudget.
// Amounts are strings to avoid float64 precision loss in JSON.
type budgetLineItemRequest struct {
	// AccountID is the chart-of-accounts account this line applies to.
	AccountID string `json:"account_id" validate:"required,uuid"`
	// CostCenterID optionally ties this line to a cost centre for departmental reporting.
	CostCenterID string `json:"cost_center_id" validate:"omitempty,uuid"`
	// PeriodID optionally ties this line to a specific accounting period.
	PeriodID string `json:"period_id" validate:"omitempty,uuid"`
	// BudgetedAmount is the planned spend/revenue as a decimal string, e.g. "50000.00".
	BudgetedAmount string `json:"budgeted_amount" validate:"required,decimal"`
	Notes          string `json:"notes"`
}

// createBudgetRequest is the body for POST /budgets.
type createBudgetRequest struct {
	FiscalYearID string                  `json:"fiscal_year_id" validate:"required,uuid"`
	Name         string                  `json:"name"           validate:"required,max=200"`
	Description  string                  `json:"description"`
	BudgetType   string                  `json:"budget_type"    validate:"required"`
	CurrencyCode string                  `json:"currency_code"  validate:"required,len=3"`
	CostCenterID string                  `json:"cost_center_id" validate:"omitempty,uuid"`
	Lines        []budgetLineItemRequest `json:"lines"          validate:"dive"`
}

// ============================================================================
// Budget CRUD handlers
// ============================================================================

// CreateBudget creates a new budget in DRAFT status with its line items.
// The budget must be submitted and approved before it constrains transactions.
//
// POST /api/v1/finance/budgets
// Permission: finance.budgets.create
func (h *FinanceHandler) CreateBudget(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.CreateBudget")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	var req createBudgetRequest
	if err := h.bind(c, &req); err != nil {
		return h.fail(c, err)
	}

	fyID, err := parseUUID(req.FiscalYearID, "fiscal_year_id")
	if err != nil {
		return h.fail(c, err)
	}

	budget := &financeDomain.Budget{
		FiscalYearID: fyID,
		Name:         req.Name,
		BudgetType:   financeDomain.BudgetType(req.BudgetType),
		CurrencyCode: req.CurrencyCode,
		Status:       financeDomain.BudgetStatusDraft,
		Version:      1,
	}
	if req.Description != "" {
		budget.Description = &req.Description
	}
	if req.CostCenterID != "" {
		ccID, err := parseUUID(req.CostCenterID, "cost_center_id")
		if err != nil {
			return h.fail(c, err)
		}
		budget.CostCenterID = &ccID
	}

	// Build line items — each amount is parsed from string to preserve precision
	lines := make([]*financeDomain.BudgetLineItem, 0, len(req.Lines))
	for i, l := range req.Lines {
		acctID, err := parseUUID(l.AccountID, "lines["+itoa(i)+"].account_id")
		if err != nil {
			return h.fail(c, err)
		}
		amount, err := decimalFromString(l.BudgetedAmount)
		if err != nil {
			return h.fail(c, err)
		}
		li := &financeDomain.BudgetLineItem{
			AccountID:      acctID,
			BudgetedAmount: amount,
		}
		if l.Notes != "" {
			li.Notes = &l.Notes
		}
		if l.CostCenterID != "" {
			ccID, err := parseUUID(l.CostCenterID, "lines["+itoa(i)+"].cost_center_id")
			if err != nil {
				return h.fail(c, err)
			}
			li.CostCenterID = &ccID
		}
		if l.PeriodID != "" {
			pID, err := parseUUID(l.PeriodID, "lines["+itoa(i)+"].period_id")
			if err != nil {
				return h.fail(c, err)
			}
			li.PeriodID = &pID
		}
		lines = append(lines, li)
	}

	// 3. Delegate
	created, err := h.services.Budget.CreateBudget(ctx, budget, lines)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok201(c, created)
}

// GetBudget retrieves a single budget with its header fields.
// Use GetBudgetLines to fetch the associated line items.
//
// GET /api/v1/finance/budgets/:id
// Permission: finance.budgets.read
func (h *FinanceHandler) GetBudget(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.GetBudget")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	budgetID, err := parseUUID(c.Params("id"), "budget ID")
	if err != nil {
		return h.fail(c, err)
	}

	// 3. Delegate
	budget, err := h.services.Budget.GetBudget(ctx, budgetID)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, budget)
}

// ListBudgets returns budgets for the current tenant, optionally filtered by
// fiscal year.
//
// GET /api/v1/finance/budgets?fiscal_year_id=<uuid>
// Permission: finance.budgets.read
func (h *FinanceHandler) ListBudgets(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.ListBudgets")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse — optional fiscal year filter
	var fiscalYearID *uuid.UUID
	if s := c.Query("fiscal_year_id"); s != "" {
		id, err := parseUUID(s, "fiscal_year_id")
		if err != nil {
			return h.fail(c, err)
		}
		fiscalYearID = &id
	}

	// 3. Delegate
	budgets, err := h.services.Budget.ListBudgets(ctx, fiscalYearID)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, budgets)
}

// GetBudgetLines returns the line items for a budget.
// Kept as a separate endpoint to avoid large payloads on list/get operations.
//
// GET /api/v1/finance/budgets/:id/lines
// Permission: finance.budgets.read
func (h *FinanceHandler) GetBudgetLines(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.GetBudgetLines")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	budgetID, err := parseUUID(c.Params("id"), "budget ID")
	if err != nil {
		return h.fail(c, err)
	}

	// 3. Delegate
	lines, err := h.services.Budget.GetLineItems(ctx, budgetID)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, lines)
}

// ============================================================================
// Budget workflow handlers
// ============================================================================

// SubmitBudget moves a DRAFT budget into PENDING_APPROVAL.
// The submitting user's ID is recorded for audit purposes.
//
// POST /api/v1/finance/budgets/:id/submit
// Permission: finance.budgets.submit
func (h *FinanceHandler) SubmitBudget(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.SubmitBudget")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	budgetID, err := parseUUID(c.Params("id"), "budget ID")
	if err != nil {
		return h.fail(c, err)
	}

	// extractUserID reads user_id as string from Locals and parses to UUID.
	// Auth middleware stores it as string; asserting to uuid.UUID directly
	// would silently yield uuid.Nil.
	byUserID := extractUserID(c)

	// 3. Delegate
	budget, err := h.services.Budget.SubmitBudget(ctx, budgetID, byUserID)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, budget)
}

// ApproveBudget moves a PENDING_APPROVAL budget to APPROVED.
// Only users with the finance.budgets.approve permission may call this.
//
// POST /api/v1/finance/budgets/:id/approve
// Permission: finance.budgets.approve
func (h *FinanceHandler) ApproveBudget(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.ApproveBudget")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	budgetID, err := parseUUID(c.Params("id"), "budget ID")
	if err != nil {
		return h.fail(c, err)
	}
	byUserID := extractUserID(c)

	// 3. Delegate
	budget, err := h.services.Budget.ApproveBudget(ctx, budgetID, byUserID)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, budget)
}

// RejectBudget returns a PENDING_APPROVAL budget to REJECTED.
// The approver may provide a note explaining the rejection.
//
// POST /api/v1/finance/budgets/:id/reject
// Permission: finance.budgets.approve
func (h *FinanceHandler) RejectBudget(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.RejectBudget")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	budgetID, err := parseUUID(c.Params("id"), "budget ID")
	if err != nil {
		return h.fail(c, err)
	}

	var req struct {
		Note string `json:"note"`
	}
	_ = c.BodyParser(&req) // note is optional; ignore parse errors

	byUserID := extractUserID(c)

	// 3. Delegate
	budget, err := h.services.Budget.RejectBudget(ctx, budgetID, byUserID, req.Note)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, budget)
}

// CloseBudget marks an APPROVED budget as CLOSED at the end of the fiscal year.
// Closed budgets are read-only for reporting purposes.
//
// POST /api/v1/finance/budgets/:id/close
// Permission: finance.budgets.close
func (h *FinanceHandler) CloseBudget(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.CloseBudget")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	budgetID, err := parseUUID(c.Params("id"), "budget ID")
	if err != nil {
		return h.fail(c, err)
	}

	// 3. Delegate
	budget, err := h.services.Budget.CloseBudget(ctx, budgetID)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, budget)
}

// ============================================================================
// Private helpers — budgets only
// ============================================================================

// itoa converts an int to string for readable error field names like "lines[2].account_id".
func itoa(i int) string { return strconv.Itoa(i) }
