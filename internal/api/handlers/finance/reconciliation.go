package finance

// reconciliation.go — HTTP handlers for bank statement import and line matching.
//
// Routes (registered in routes.go):
//
//	POST   /api/v1/finance/reconciliation/statements                            → ImportBankStatement
//	GET    /api/v1/finance/reconciliation/statements                            → ListBankStatements
//	GET    /api/v1/finance/reconciliation/statements/:id                        → GetBankStatement
//	GET    /api/v1/finance/reconciliation/statements/:id/lines                  → ListStatementLines
//	POST   /api/v1/finance/reconciliation/statements/:id/lines/:line_id/match   → MatchStatementLine
//	DELETE /api/v1/finance/reconciliation/statements/:id/lines/:line_id/match   → UnmatchStatementLine
//	POST   /api/v1/finance/reconciliation/statements/:id/complete               → CompleteReconciliation
//
// Reconciliation flow:
//
//  1. Client imports a bank statement via ImportBankStatement.
//  2. Client (or auto-matching job) calls MatchStatementLine to pair each
//     bank line with a journal entry.
//  3. Once all lines are matched, CompleteReconciliation closes the statement.

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	financeDomain "awo.so/internal/core/finance/domain"
)

// ============================================================================
// Bank statement request types
// ============================================================================

// bankStatementLineRequest represents one row in the imported CSV/OFX statement.
// Amounts are strings to preserve decimal precision.
type bankStatementLineRequest struct {
	TransactionDate string `json:"transaction_date" validate:"required"`
	// ValueDate is the date the funds are available; may differ from transaction date.
	ValueDate    *string `json:"value_date"`
	Description  string  `json:"description"      validate:"required,max=500"`
	Reference    *string `json:"reference"`
	DebitAmount  string  `json:"debit_amount"  validate:"required,decimal"`
	CreditAmount string  `json:"credit_amount" validate:"required,decimal"`
	Balance      string  `json:"balance"       validate:"required,decimal"`
}

// importBankStatementRequest is the body for POST /reconciliation/statements.
type importBankStatementRequest struct {
	AccountID          string                     `json:"account_id"           validate:"required,uuid"`
	StatementReference string                     `json:"statement_reference"  validate:"required,max=100"`
	StatementDate      string                     `json:"statement_date"       validate:"required"`
	StartDate          string                     `json:"start_date"           validate:"required"`
	CurrencyCode       string                     `json:"currency_code"        validate:"required,len=3"`
	OpeningBalance     string                     `json:"opening_balance"      validate:"required,decimal"`
	ClosingBalance     string                     `json:"closing_balance"      validate:"required,decimal"`
	Lines              []bankStatementLineRequest `json:"lines" validate:"dive"`
}

// ============================================================================
// Bank statement handlers
// ============================================================================

// ImportBankStatement imports a bank statement and its transaction lines.
// Lines are unmatched on import; use MatchStatementLine to pair them with
// journal entries.
//
// POST /api/v1/finance/reconciliation/statements
// Permission: finance.reconciliation.import
func (h *FinanceHandler) ImportBankStatement(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.ImportBankStatement")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	var req importBankStatementRequest
	if err := h.bind(c, &req); err != nil {
		return h.fail(c, err)
	}

	accountID, err := parseUUID(req.AccountID, "account_id")
	if err != nil {
		return h.fail(c, err)
	}

	stmtDate, err := time.Parse("2006-01-02", req.StatementDate)
	if err != nil {
		return h.fail(c, newDateParseError("statement_date"))
	}
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return h.fail(c, newDateParseError("start_date"))
	}

	openingBalance, err := decimalFromString(req.OpeningBalance)
	if err != nil {
		return h.fail(c, err)
	}
	closingBalance, err := decimalFromString(req.ClosingBalance)
	if err != nil {
		return h.fail(c, err)
	}

	byUserID := extractUserID(c)

	stmt := &financeDomain.BankStatement{
		AccountID:          accountID,
		StatementReference: req.StatementReference,
		StatementDate:      stmtDate,
		StartDate:          startDate,
		CurrencyCode:       req.CurrencyCode,
		OpeningBalance:     openingBalance,
		ClosingBalance:     closingBalance,
		CreatedBy:          byUserID,
	}

	// Parse statement lines — each amount comes in as a string
	lines := make([]*financeDomain.BankStatementLine, 0, len(req.Lines))
	for i, l := range req.Lines {
		txDate, err := time.Parse("2006-01-02", l.TransactionDate)
		if err != nil {
			return h.fail(c, newIndexedDateParseError("transaction_date", i))
		}
		debit, err := decimalFromString(l.DebitAmount)
		if err != nil {
			return h.fail(c, err)
		}
		credit, err := decimalFromString(l.CreditAmount)
		if err != nil {
			return h.fail(c, err)
		}
		balance, err := decimalFromString(l.Balance)
		if err != nil {
			return h.fail(c, err)
		}
		li := &financeDomain.BankStatementLine{
			TransactionDate: txDate,
			Description:     l.Description,
			Reference:       l.Reference,
			DebitAmount:     debit,
			CreditAmount:    credit,
			Balance:         balance,
		}
		if l.ValueDate != nil {
			vd, err := time.Parse("2006-01-02", *l.ValueDate)
			if err != nil {
				return h.fail(c, newIndexedDateParseError("value_date", i))
			}
			li.ValueDate = &vd
		}
		lines = append(lines, li)
	}

	// 3. Delegate
	created, err := h.services.Reconciliation.ImportStatement(ctx, stmt, lines)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok201(c, created)
}

// GetBankStatement retrieves a single bank statement by ID.
//
// GET /api/v1/finance/reconciliation/statements/:id
// Permission: finance.reconciliation.read
func (h *FinanceHandler) GetBankStatement(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.GetBankStatement")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	stmtID, err := parseUUID(c.Params("id"), "statement ID")
	if err != nil {
		return h.fail(c, err)
	}

	// 3. Delegate
	stmt, err := h.services.Reconciliation.GetStatement(ctx, stmtID)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, stmt)
}

// ListBankStatements returns bank statements for the current tenant, optionally
// filtered to a single account.
//
// GET /api/v1/finance/reconciliation/statements?account_id=<uuid>
// Permission: finance.reconciliation.read
func (h *FinanceHandler) ListBankStatements(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.ListBankStatements")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse — optional account filter
	var accountID *uuid.UUID
	if raw := c.Query("account_id"); raw != "" {
		id, err := parseUUID(raw, "account_id")
		if err != nil {
			return h.fail(c, err)
		}
		accountID = &id
	}

	// 3. Delegate
	list, err := h.services.Reconciliation.ListStatements(ctx, accountID)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, list)
}

// ListStatementLines returns the transaction lines for a bank statement.
// Pass unmatched_only=true to see only lines that still need a journal match.
//
// GET /api/v1/finance/reconciliation/statements/:id/lines?unmatched_only=true
// Permission: finance.reconciliation.read
func (h *FinanceHandler) ListStatementLines(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.ListStatementLines")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	stmtID, err := parseUUID(c.Params("id"), "statement ID")
	if err != nil {
		return h.fail(c, err)
	}
	unmatchedOnly := c.QueryBool("unmatched_only", false)

	// 3. Delegate
	lines, err := h.services.Reconciliation.ListLines(ctx, stmtID, unmatchedOnly)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, lines)
}

// ============================================================================
// Matching handlers
// ============================================================================

// MatchStatementLine pairs a bank statement line with a journal entry.
// The match is recorded against the statement; the statement status is
// updated automatically by the service when all lines are matched.
//
// POST /api/v1/finance/reconciliation/statements/:id/lines/:line_id/match
// Permission: finance.reconciliation.match
func (h *FinanceHandler) MatchStatementLine(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.MatchStatementLine")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate — two path params + body
	stmtID, err := parseUUID(c.Params("id"), "statement ID")
	if err != nil {
		return h.fail(c, err)
	}
	lineID, err := parseUUID(c.Params("line_id"), "line ID")
	if err != nil {
		return h.fail(c, err)
	}

	var req struct {
		// EntryID is the journal entry UUID to match against this bank line.
		EntryID string `json:"entry_id" validate:"required,uuid"`
	}
	if err := h.bind(c, &req); err != nil {
		return h.fail(c, err)
	}
	entryID, err := parseUUID(req.EntryID, "entry_id")
	if err != nil {
		return h.fail(c, err)
	}

	byUserID := extractUserID(c)

	// 3. Delegate
	stmt, err := h.services.Reconciliation.MatchLine(ctx, stmtID, lineID, entryID, byUserID)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, stmt)
}

// UnmatchStatementLine removes the match between a bank line and a journal entry,
// returning the line to unmatched status.
//
// DELETE /api/v1/finance/reconciliation/statements/:id/lines/:line_id/match
// Permission: finance.reconciliation.match
func (h *FinanceHandler) UnmatchStatementLine(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.UnmatchStatementLine")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate — two path params, no body
	stmtID, err := parseUUID(c.Params("id"), "statement ID")
	if err != nil {
		return h.fail(c, err)
	}
	lineID, err := parseUUID(c.Params("line_id"), "line ID")
	if err != nil {
		return h.fail(c, err)
	}

	// 3. Delegate
	stmt, err := h.services.Reconciliation.UnmatchLine(ctx, stmtID, lineID)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, stmt)
}

// CompleteReconciliation marks the bank statement as fully reconciled.
// The service enforces that all lines are matched before allowing completion.
//
// POST /api/v1/finance/reconciliation/statements/:id/complete
// Permission: finance.reconciliation.complete
func (h *FinanceHandler) CompleteReconciliation(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.CompleteReconciliation")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	stmtID, err := parseUUID(c.Params("id"), "statement ID")
	if err != nil {
		return h.fail(c, err)
	}
	byUserID := extractUserID(c)

	// 3. Delegate
	stmt, err := h.services.Reconciliation.CompleteReconciliation(ctx, stmtID, byUserID)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, stmt)
}

// ============================================================================
// Private helpers — reconciliation only
// ============================================================================

// newIndexedDateParseError returns a date parse error that includes the line
// index, making it easier to locate the offending row in a large import.
func newIndexedDateParseError(field string, lineIndex int) error {
	return newValidationError(
		"INVALID_DATE_FORMAT",
		"lines["+itoa(lineIndex)+"]."+field+" must be in YYYY-MM-DD format",
		"lines["+itoa(lineIndex)+"]."+field,
		"Example: 2025-01-31",
	)
}
