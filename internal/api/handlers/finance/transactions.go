package finance

// transactions.go — HTTP handlers for journal transaction management.
//
// Routes (registered in routes.go):
//
//	POST   /api/v1/finance/transactions               → CreateTransaction
//	GET    /api/v1/finance/transactions               → ListTransactions
//	GET    /api/v1/finance/transactions/:id           → GetTransaction
//	POST   /api/v1/finance/transactions/:id/submit    → SubmitTransaction
//	POST   /api/v1/finance/transactions/:id/approve   → ApproveTransaction
//	POST   /api/v1/finance/transactions/:id/reject    → RejectTransaction
//
// Approval state machine (enforced by TransactionService):
//
//	DRAFT → submit → PENDING_APPROVAL → approve → POSTED
//	                                  → reject  → REJECTED

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"

	"awo.so/internal/core/finance/domain"
	financeDomain "awo.so/internal/core/finance/domain"
)

// CreateTransaction creates a new journal entry in DRAFT status.
// The entry is not posted until SubmitTransaction + ApproveTransaction are called.
//
// POST /api/v1/finance/transactions
// Permission: finance.transactions.create
func (h *FinanceHandler) CreateTransaction(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.CreateTransaction")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	var req financeDomain.CreateTransactionRequest
	if err := h.bind(c, &req); err != nil {
		return h.fail(c, err)
	}

	// 3. Delegate
	tx, err := h.services.Transaction.CreateTransaction(ctx, req)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	span.SetAttributes(
		attribute.String("transaction.id", tx.ID.String()),
		attribute.String("transaction.number", tx.TransactionNumber),
	)
	return h.ok201(c, tx)
}

// GetTransaction retrieves a single transaction with its line entries.
//
// GET /api/v1/finance/transactions/:id
// Permission: finance.transactions.read
func (h *FinanceHandler) GetTransaction(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.GetTransaction")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	txID, err := parseUUID(c.Params("id"), "transaction ID")
	if err != nil {
		return h.fail(c, err)
	}

	// 3. Delegate
	tx, err := h.services.Transaction.GetTransactionByID(ctx, txID)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, tx)
}

// ListTransactions returns a paginated list of transactions.
// Filter by account_id to see all entries touching a specific account.
//
// GET /api/v1/finance/transactions?offset=0&limit=20&account_id=<uuid>
// Permission: finance.transactions.read
func (h *FinanceHandler) ListTransactions(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.ListTransactions")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate — query params only
	offset, limit := h.pagination(c)
	filters := financeDomain.TransactionFilter{
		Offset: &offset,
		Limit:  &limit,
	}
	// account_id filter is optional; parse separately to avoid shadowing the
	// uuid package (a common mistake when using short := declarations here).
	if raw := c.Query("account_id"); raw != "" {
		accountUUID, err := parseUUID(raw, "account_id")
		if err != nil {
			return h.fail(c, err)
		}
		filters.AccountID = &accountUUID
	}

	// 3. Delegate
	txs, err := h.services.Transaction.ListTransactions(ctx, &filters)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200Meta(c, txs, map[string]any{
		"pagination": map[string]any{
			"offset": offset,
			"limit":  limit,
			"count":  len(txs),
		},
	})
}

// SubmitTransaction moves a DRAFT transaction into PENDING_APPROVAL.
// The submitting user is the one who will be recorded as having submitted.
//
// POST /api/v1/finance/transactions/:id/submit
// Permission: finance.transactions.submit
func (h *FinanceHandler) SubmitTransaction(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.SubmitTransaction")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	txID, err := parseUUID(c.Params("id"), "transaction ID")
	if err != nil {
		return h.fail(c, err)
	}

	// 3. Delegate — PostTransaction transitions DRAFT → PENDING_APPROVAL
	tx, err := h.services.Transaction.PostTransaction(ctx, txID, nil)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, tx)
}

// ApproveTransaction moves a PENDING_APPROVAL transaction to POSTED,
// which triggers double-entry ledger recording in the service layer.
//
// POST /api/v1/finance/transactions/:id/approve
// Permission: finance.transactions.approve
func (h *FinanceHandler) ApproveTransaction(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.ApproveTransaction")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	txID, err := parseUUID(c.Params("id"), "transaction ID")
	if err != nil {
		return h.fail(c, err)
	}

	// Notes are optional on approval
	var req struct {
		Notes string `json:"notes"`
	}
	_ = c.BodyParser(&req) // ignore parse error — all fields optional

	// 3. Delegate
	tx, err := h.services.Transaction.ApproveTransaction(ctx, txID, req.Notes)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, tx)
}

// RejectTransaction moves a PENDING_APPROVAL transaction back to REJECTED.
// A reason code is required; notes provide additional context.
//
// POST /api/v1/finance/transactions/:id/reject
// Permission: finance.transactions.approve
//
// Valid reason codes are defined in domain.ValidReasons.
func (h *FinanceHandler) RejectTransaction(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.RejectTransaction")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	txID, err := parseUUID(c.Params("id"), "transaction ID")
	if err != nil {
		return h.fail(c, err)
	}

	var req struct {
		// Reason must match one of domain.ValidReasons.
		Reason string `json:"reason" validate:"required"`
		// Notes give the approver space for free-form context.
		Notes string `json:"notes"`
	}
	if err := h.bind(c, &req); err != nil {
		return h.fail(c, err)
	}

	reason := domain.RejectionReason(req.Reason)
	if !isValidRejectionReason(reason) {
		return h.fail(c, newInvalidRejectionReasonError(req.Reason))
	}

	// 3. Delegate
	tx, err := h.services.Transaction.RejectTransaction(ctx, txID, reason, req.Notes)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, tx)
}

// ============================================================================
// Private helpers — transactions only
// ============================================================================

// isValidRejectionReason checks whether reason is in the domain allowlist.
func isValidRejectionReason(reason domain.RejectionReason) bool {
	for _, v := range domain.ValidReasons {
		if reason == v {
			return true
		}
	}
	return false
}

// newInvalidRejectionReasonError builds a descriptive error that lists all
// accepted values, so clients know exactly what to send.
func newInvalidRejectionReasonError(received string) error {
	valid := make([]string, len(domain.ValidReasons))
	for i, r := range domain.ValidReasons {
		valid[i] = string(r)
	}
	return newValidationError(
		"INVALID_REJECTION_REASON",
		"rejection reason '"+received+"' is not valid",
		"reason",
		"Valid values are: "+strings.Join(valid, ", "),
	)
}
