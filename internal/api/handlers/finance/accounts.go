package finance

// accounts.go — HTTP handlers for chart-of-accounts management.
//
// Routes (registered in routes.go):
//
//	POST   /api/v1/finance/accounts          → CreateAccount
//	GET    /api/v1/finance/accounts          → ListAccounts
//	GET    /api/v1/finance/accounts/:id      → GetAccount
//	PUT    /api/v1/finance/accounts/:id      → UpdateAccount
//	DELETE /api/v1/finance/accounts/:id      → DeleteAccount
//	GET    /api/v1/finance/accounts/:id/balance → GetAccountBalance  (stub)

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"

	financeDomain "awo.so/internal/core/finance/domain"
)

// CreateAccount creates a new account in the chart of accounts.
//
// POST /api/v1/finance/accounts
// Permission: finance.accounts.create
func (h *FinanceHandler) CreateAccount(c *fiber.Ctx) error {
	// 1. Observability — span inherits tenant/user values from c.UserContext()
	//    which was enriched by TenantMiddleware and AuthMiddleware.
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.CreateAccount")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	var req financeDomain.CreateAccountRequest
	if err := h.bind(c, &req); err != nil {
		return h.fail(c, err)
	}

	// 3. Delegate to service — pass ctx, never c.Context()
	account, err := h.services.Account.Create(ctx, &req)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Trace attributes + response
	span.SetAttributes(
		attribute.String("account.id", account.ID.String()),
		attribute.String("account.code", account.AccountCode),
		attribute.String("account.type", string(account.RootType)),
	)
	return h.ok201(c, account)
}

// GetAccount retrieves a single account by ID.
//
// GET /api/v1/finance/accounts/:id
// Permission: finance.accounts.read
func (h *FinanceHandler) GetAccount(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.GetAccount")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate — path param only, no body
	accountID, err := parseUUID(c.Params("id"), "account ID")
	if err != nil {
		return h.fail(c, err)
	}

	// 3. Delegate
	account, err := h.services.Account.GetByID(ctx, accountID)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, account)
}

// ListAccounts returns a paginated list of accounts with optional filters.
//
// GET /api/v1/finance/accounts?offset=0&limit=20&root_type=asset&active=true
// Permission: finance.accounts.read
func (h *FinanceHandler) ListAccounts(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.ListAccounts")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate — query params only
	offset, limit := h.pagination(c)

	filters := financeDomain.AccountFilter{
		IsActive: boolQuery(c, "active"),
		Offset:   &offset,
		Limit:    &limit,
	}
	// root_type is optional; skip silently if the value is not a known type
	if rt := c.Query("root_type"); rt != "" {
		parsed, err := financeDomain.ParseRootType(rt)
		if err == nil {
			filters.RootType = &parsed
		}
	}

	// 3. Delegate
	accounts, err := h.services.Account.List(ctx, &filters)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response with pagination metadata
	span.SetAttributes(
		attribute.Int("pagination.offset", offset),
		attribute.Int("pagination.limit", limit),
		attribute.Int("results.count", len(accounts)),
	)
	return h.ok200Meta(c, accounts, map[string]any{
		"pagination": map[string]any{
			"offset": offset,
			"limit":  limit,
			"count":  len(accounts),
		},
	})
}

// UpdateAccount replaces a single account's mutable fields.
//
// PUT /api/v1/finance/accounts/:id
// Permission: finance.accounts.update
func (h *FinanceHandler) UpdateAccount(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.UpdateAccount")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	accountID, err := parseUUID(c.Params("id"), "account ID")
	if err != nil {
		return h.fail(c, err)
	}

	var req financeDomain.UpdateAccountRequest
	if err := h.bind(c, &req); err != nil {
		return h.fail(c, err)
	}

	// 3. Delegate
	updated, err := h.services.Account.Update(ctx, accountID, &req)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, updated)
}

// DeleteAccount soft-deletes an account.
// Accounts with posted transactions cannot be deleted (enforced by service layer).
//
// DELETE /api/v1/finance/accounts/:id
// Permission: finance.accounts.delete
func (h *FinanceHandler) DeleteAccount(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.DeleteAccount")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	accountID, err := parseUUID(c.Params("id"), "account ID")
	if err != nil {
		return h.fail(c, err)
	}

	// 3. Delegate
	if err := h.services.Account.Delete(ctx, accountID); err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. 204 — no body on delete
	return h.ok204(c)
}

// GetAccountBalance returns the running balance for an account as of a given date.
//
// GET /api/v1/finance/accounts/:id/balance?as_of_date=2025-12-31
// Permission: finance.accounts.read
//
// NOTE: AccountService.GetAccountBalance is not yet implemented.
// Returns 501 until the service method is ready.
func (h *FinanceHandler) GetAccountBalance(c *fiber.Ctx) error {
	return h.notImplemented(c, "GetAccountBalance")
}
