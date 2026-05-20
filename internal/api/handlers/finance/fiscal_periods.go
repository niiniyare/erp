package finance

// fiscal_periods.go — HTTP handlers for fiscal year and accounting period management.
//
// Routes (registered in routes.go):
//
//	POST  /api/v1/finance/fiscal-years                       → CreateFiscalYear
//	GET   /api/v1/finance/fiscal-years                       → ListFiscalYears
//	GET   /api/v1/finance/fiscal-years/:id                   → GetFiscalYear
//	POST  /api/v1/finance/fiscal-years/:id/periods           → CreatePeriod
//	GET   /api/v1/finance/fiscal-years/:id/periods           → ListPeriods
//	GET   /api/v1/finance/periods/:id                        → GetPeriod
//	GET   /api/v1/finance/periods/current                    → GetCurrentPeriod
//	PATCH /api/v1/finance/periods/:id/status                 → ChangePeriodStatus
//
// Period status machine (enforced by PeriodService):
//
//	OPEN → CLOSED → ARCHIVED
//	     ↑ re-open (requires checks_passed=true)

import (
	"time"

	"github.com/gofiber/fiber/v2"

	financeDomain "awo.so/internal/core/finance/domain"
	sharedErrors "awo.so/internal/shared/errors"
)

// ============================================================================
// Fiscal year request types
//
// Defined here rather than inlined as anonymous structs so that:
//   - The validator can apply struct-level rules
//   - Tests can construct values without importing fiber
//   - Generated API docs pick up the type names
// ============================================================================

// createFiscalYearRequest is the body for POST /fiscal-years.
type createFiscalYearRequest struct {
	Name      string    `json:"name"       validate:"required,max=100"`
	StartDate time.Time `json:"start_date" validate:"required"`
	EndDate   time.Time `json:"end_date"   validate:"required,gtfield=StartDate"`
}

// createPeriodRequest is the body for POST /fiscal-years/:id/periods.
type createPeriodRequest struct {
	PeriodNumber int       `json:"period_number" validate:"required,min=1,max=13"`
	Name         string    `json:"name"          validate:"required,max=100"`
	StartDate    time.Time `json:"start_date"    validate:"required"`
	EndDate      time.Time `json:"end_date"      validate:"required,gtfield=StartDate"`
}

// changePeriodStatusRequest is the body for PATCH /periods/:id/status.
type changePeriodStatusRequest struct {
	// Status must match one of financeDomain.PeriodStatus values.
	Status string `json:"status" validate:"required"`
	// ChecksPassed must be true when re-opening a closed period.
	// The service layer validates this against the current status transition.
	ChecksPassed bool `json:"checks_passed"`
}

// ============================================================================
// Fiscal year handlers
// ============================================================================

// CreateFiscalYear creates a new fiscal year for the current tenant.
// Each tenant may have multiple fiscal years; the service enforces that date
// ranges do not overlap.
//
// POST /api/v1/finance/fiscal-years
// Permission: finance.fiscal_years.create
func (h *FinanceHandler) CreateFiscalYear(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.CreateFiscalYear")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	var req createFiscalYearRequest
	if err := h.bind(c, &req); err != nil {
		return h.fail(c, err)
	}

	// 3. Delegate
	fy := &financeDomain.FiscalYear{
		Name:      req.Name,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	}
	created, err := h.services.Period.CreateFiscalYear(ctx, fy)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok201(c, created)
}

// ListFiscalYears returns all fiscal years for the current tenant, ordered
// by start date descending (most recent first).
//
// GET /api/v1/finance/fiscal-years
// Permission: finance.fiscal_years.read
func (h *FinanceHandler) ListFiscalYears(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.ListFiscalYears")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. No request body or query params for this endpoint

	// 3. Delegate
	fys, err := h.services.Period.ListFiscalYears(ctx)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, fys)
}

// GetFiscalYear retrieves a single fiscal year by ID.
//
// GET /api/v1/finance/fiscal-years/:id
// Permission: finance.fiscal_years.read
func (h *FinanceHandler) GetFiscalYear(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.GetFiscalYear")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	fyID, err := parseUUID(c.Params("id"), "fiscal year ID")
	if err != nil {
		return h.fail(c, err)
	}

	// 3. Delegate
	fy, err := h.services.Period.GetFiscalYearByID(ctx, fyID)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, fy)
}

// ============================================================================
// Accounting period handlers
// ============================================================================

// CreatePeriod creates a new accounting period within a fiscal year.
// Period numbers within a fiscal year must be unique (enforced by service).
//
// POST /api/v1/finance/fiscal-years/:id/periods
// Permission: finance.periods.create
func (h *FinanceHandler) CreatePeriod(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.CreatePeriod")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	fyID, err := parseUUID(c.Params("id"), "fiscal year ID")
	if err != nil {
		return h.fail(c, err)
	}

	var req createPeriodRequest
	if err := h.bind(c, &req); err != nil {
		return h.fail(c, err)
	}

	// 3. Delegate — new periods always start as OPEN
	period := &financeDomain.AccountingPeriod{
		FiscalYearID: fyID,
		PeriodNumber: req.PeriodNumber,
		Name:         req.Name,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		Status:       financeDomain.PeriodStatusOpen,
	}
	created, err := h.services.Period.CreatePeriod(ctx, period)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok201(c, created)
}

// ListPeriods returns all accounting periods for a fiscal year.
//
// GET /api/v1/finance/fiscal-years/:id/periods
// Permission: finance.periods.read
func (h *FinanceHandler) ListPeriods(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.ListPeriods")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	fyID, err := parseUUID(c.Params("id"), "fiscal year ID")
	if err != nil {
		return h.fail(c, err)
	}

	// 3. Delegate
	periods, err := h.services.Period.ListPeriods(ctx, fyID)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, periods)
}

// GetPeriod retrieves a single accounting period by its own ID.
//
// GET /api/v1/finance/periods/:id
// Permission: finance.periods.read
func (h *FinanceHandler) GetPeriod(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.GetPeriod")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	periodID, err := parseUUID(c.Params("id"), "period ID")
	if err != nil {
		return h.fail(c, err)
	}

	// 3. Delegate
	period, err := h.services.Period.GetPeriodByID(ctx, periodID)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, period)
}

// GetCurrentPeriod returns the accounting period that contains today's date.
// Returns 404 when no open period covers today (e.g. between fiscal years).
//
// GET /api/v1/finance/periods/current
// Permission: finance.periods.read
func (h *FinanceHandler) GetCurrentPeriod(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.GetCurrentPeriod")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. No request params

	// 3. Delegate
	period, err := h.services.Period.GetCurrentPeriod(ctx)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, period)
}

// ChangePeriodStatus transitions a period between OPEN, CLOSED, and ARCHIVED.
// The user ID is recorded in the audit trail for compliance.
//
// PATCH /api/v1/finance/periods/:id/status
// Permission: finance.periods.close  (or .reopen depending on direction)
func (h *FinanceHandler) ChangePeriodStatus(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.ChangePeriodStatus")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	periodID, err := parseUUID(c.Params("id"), "period ID")
	if err != nil {
		return h.fail(c, err)
	}

	var req changePeriodStatusRequest
	if err := h.bind(c, &req); err != nil {
		return h.fail(c, err)
	}

	newStatus := financeDomain.PeriodStatus(req.Status)
	if !newStatus.IsValid() {
		return h.fail(c, sharedErrors.NewBusinessError("INVALID_PERIOD_STATUS", "period status '"+req.Status+"' is not valid").
			WithHTTPStatus(fiber.StatusBadRequest).
			WithCategory(sharedErrors.CategoryValidation).
			WithSuggestion("Valid values: open, closed, archived"))
	}

	// extractUserID reads user_id from Locals as string then parses to UUID.
	// This is safe — auth middleware stores user_id as a string, not uuid.UUID.
	byUserID := extractUserID(c)

	// 3. Delegate
	period, err := h.services.Period.ChangePeriodStatus(ctx, periodID, newStatus, byUserID, req.ChecksPassed)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, period)
}
