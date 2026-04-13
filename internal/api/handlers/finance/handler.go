package finance

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	financeDomain "awo.so/internal/core/finance/domain"
	financeService "awo.so/internal/core/finance/service"
	"awo.so/internal/shared/errors"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// FinanceHandler handles finance-related HTTP requests with centralized error handling
type FinanceHandler struct {
	services  *financeService.Services
	logger    logger.Logger
	metrics   metrics.MetricsProvider
	tracer    tracing.Service
	validator *validator.Validate
}

// NewFinanceHandler creates a new finance handler with dependencies
func NewFinanceHandler(
	services *financeService.Services,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.Service,
) *FinanceHandler {
	validator := validator.New()

	// Register custom validators
	registerCustomValidators(validator)

	return &FinanceHandler{
		services:  services,
		logger:    logger,
		metrics:   metrics,
		tracer:    tracer,
		validator: validator,
	}
}

// ============================================================================
// ACCOUNT MANAGEMENT ENDPOINTS
// ============================================================================

// CreateAccount handles account creation (POST /api/v1/finance/accounts)
// @Summary Create account
// @Description Creates a new account in the chart of accounts
// @Tags finance
// @Accept json
// @Produce json
// @Param account body financeDomain.CreateAccountRequest true "Account data"
// @Success 201 {object} financeDomain.Accounts
// @Router /api/v1/finance/accounts [post]
func (h *FinanceHandler) CreateAccount(c *fiber.Ctx) error {
	// Step 1: Start tracing
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.CreateAccount")
	defer span.End()
	c.SetUserContext(ctx)

	// Step 2: Parse and validate request
	var req financeDomain.CreateAccountRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		return h.HandleError(c, err)
	}

	// Step 3: Delegate to service
	createdAccount, err := h.services.Account.Create(c.Context(), &req)
	if err != nil {
		span.RecordError(err)
		return h.HandleError(c, err)
	}

	// Step 4: Set trace attributes
	span.SetAttributes(
		attribute.String("account.id", createdAccount.ID.String()),
		attribute.String("account.code", createdAccount.AccountCode),
		attribute.String("account.name", createdAccount.AccountName),
		attribute.String("account.type", string(createdAccount.RootType)),
	)

	// Step 5: Return successful response
	return h.Created(c, createdAccount)
}

// GetAccount handles account retrieval by ID (GET /api/v1/finance/accounts/:id)
func (h *FinanceHandler) GetAccount(c *fiber.Ctx) error {
	// Step 1: Observability
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.GetAccount")
	defer span.End()

	accountID := c.Params("id")
	h.logger.InfoContext(ctx, "Retrieving account", logger.Fields{
		"account_id": accountID,
		"method":     c.Method(),
		"path":       c.Path(),
	})

	// Step 2: Parse and Validate
	if accountID == "" {
		return h.HandleError(c, errors.NewBusinessError("MISSING_ACCOUNT_ID", "Account ID is required").
			WithCategory(errors.CategoryValidation).
			WithSeverity(errors.SeverityError))
	}

	// Parse account ID to UUID
	accountUUID, err := uuid.Parse(accountID)
	if err != nil {
		return h.HandleError(c, errors.NewBusinessError("INVALID_ACCOUNT_ID", "Invalid account ID format").
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation))
	}

	// Step 3: Delegate to Service
	account, err := h.services.Account.GetByID(ctx, accountUUID)
	if err != nil {
		return h.HandleError(c, err)
	}

	// Step 4: Return Response
	h.metrics.IncrementCounter("account_retrieved_total", metrics.Fields{
		"status": "success",
	})

	return h.Success(c, account)
}

// ListAccounts handles account listing with pagination (GET /api/v1/finance/accounts)
// @Summary List accounts
// @Description List accounts with pagination and filtering
// @Tags finance
// @Accept json
// @Produce json
// @Param offset query int false "Offset for pagination" default(0)
// @Param limit query int false "Limit for pagination" default(20)
// @Param root_type query string false "Filter by root type"
// @Param active query bool false "Filter by active status"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/finance/accounts [get]
func (h *FinanceHandler) ListAccounts(c *fiber.Ctx) error {
	// Step 1: Start tracing
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.ListAccounts")
	defer span.End()
	c.SetUserContext(ctx)

	// Step 2: Extract pagination and filter parameters
	offset, limit := h.ExtractPaginationParams(c)

	// Extract optional filters
	filters := financeDomain.AccountFilter{
		IsActive: h.extractBoolQuery(c, "active"),
	}

	if rootType := c.Query("root_type"); rootType != "" {
		rt, err := financeDomain.ParseRootType(rootType)
		if err == nil {
			filters.RootType = &rt
		}
	}

	// Step 3: Delegate to service
	filters.Offset = &offset
	filters.Limit = &limit
	accounts, err := h.services.Account.List(c.Context(), &filters)
	if err != nil {
		span.RecordError(err)
		return h.HandleError(c, err)
	}

	// Step 4: Return response with metadata
	meta := map[string]interface{}{
		"pagination": map[string]interface{}{
			"offset": offset,
			"limit":  limit,
			"count":  len(accounts),
		},
		"filters": filters,
	}

	span.SetAttributes(
		attribute.Int("pagination.offset", offset),
		attribute.Int("pagination.limit", limit),
		attribute.Int("results.count", len(accounts)),
	)

	return h.SuccessWithMeta(c, accounts, meta)
}

// UpdateAccount handles account updates (PUT /api/v1/finance/accounts/:id)
func (h *FinanceHandler) UpdateAccount(c *fiber.Ctx) error {
	// Step 1: Observability
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.UpdateAccount")
	defer span.End()

	accountID := c.Params("id")
	h.logger.InfoContext(ctx, "Updating account", logger.Fields{
		"account_id": accountID,
		"method":     c.Method(),
		"path":       c.Path(),
	})

	// Step 2: Parse and Validate
	var req financeDomain.UpdateAccountRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		return h.HandleError(c, err)
	}

	// Parse account ID to UUID
	accountUUID, err := uuid.Parse(accountID)
	if err != nil {
		return h.HandleError(c, errors.NewBusinessError("INVALID_ACCOUNT_ID", "Invalid account ID format").
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation))
	}

	// Step 3: Delegate to Service
	updatedAccount, err := h.services.Account.Update(ctx, accountUUID, &req)
	if err != nil {
		return h.HandleError(c, err)
	}

	// Step 4: Return Response
	return h.Success(c, updatedAccount)
}

// DeleteAccount handles account deletion (DELETE /api/v1/finance/accounts/:id)
func (h *FinanceHandler) DeleteAccount(c *fiber.Ctx) error {
	// Step 1: Observability
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.DeleteAccount")
	defer span.End()

	accountID := c.Params("id")
	h.logger.InfoContext(ctx, "Deleting account", logger.Fields{
		"account_id": accountID,
		"method":     c.Method(),
		"path":       c.Path(),
	})

	// Step 2: Parse and Validate
	accountUUID, err := uuid.Parse(accountID)
	if err != nil {
		return h.HandleError(c, errors.NewBusinessError("INVALID_ACCOUNT_ID", "Invalid account ID format").
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation))
	}

	// Step 3: Delegate to Service
	if err := h.services.Account.Delete(ctx, accountUUID); err != nil {
		return h.HandleError(c, err)
	}

	// Step 4: Return Response (204 No Content)
	return c.SendStatus(fiber.StatusNoContent)
}

// ============================================================================
// TRANSACTION MANAGEMENT ENDPOINTS
// ============================================================================

// CreateTransaction handles transaction creation (POST /api/v1/finance/transactions)
func (h *FinanceHandler) CreateTransaction(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.CreateTransaction")
	defer span.End()
	c.SetUserContext(ctx)

	var req financeDomain.CreateTransactionRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		return h.HandleError(c, err)
	}

	transaction, err := h.services.Transaction.CreateTransaction(c.Context(), req)
	if err != nil {
		span.RecordError(err)
		return h.HandleError(c, err)
	}

	span.SetAttributes(
		attribute.String("transaction.id", transaction.ID.String()),
		attribute.String("transaction.number", transaction.TransactionNumber),
	)

	return h.Created(c, transaction)
}

// GetTransaction handles transaction retrieval (GET /api/v1/finance/transactions/:id)
func (h *FinanceHandler) GetTransaction(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.GetTransaction")
	defer span.End()
	c.SetUserContext(ctx)

	transactionID := c.Params("id")
	transactionUUID, err := uuid.Parse(transactionID)
	if err != nil {
		return h.HandleError(c, errors.NewBusinessError("INVALID_TRANSACTION_ID", "Invalid transaction ID format").
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation))
	}

	transaction, err := h.services.Transaction.GetTransactionByID(ctx, transactionUUID)
	if err != nil {
		return h.HandleError(c, err)
	}

	return h.Success(c, transaction)
}

// ListTransactions handles transaction listing (GET /api/v1/finance/transactions)
func (h *FinanceHandler) ListTransactions(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.ListTransactions")
	defer span.End()
	c.SetUserContext(ctx)

	offset, limit := h.ExtractPaginationParams(c)

	// Extract optional filters
	filters := financeDomain.TransactionFilter{}
	if accountID := c.Query("account_id"); accountID != "" {
		if uuid, err := uuid.Parse(accountID); err == nil {
			filters.AccountID = &uuid
		}
	}

	filters.Offset = &offset
	filters.Limit = &limit
	transactions, err := h.services.Transaction.ListTransactions(c.Context(), &filters)
	if err != nil {
		span.RecordError(err)
		return h.HandleError(c, err)
	}

	meta := map[string]interface{}{
		"pagination": map[string]interface{}{
			"offset": offset,
			"limit":  limit,
			"count":  len(transactions),
		},
	}

	return h.SuccessWithMeta(c, transactions, meta)
}

// ============================================================================
// REPORTING ENDPOINTS
// ============================================================================

// GetTrialBalance handles trial balance report (GET /api/v1/finance/reports/trial-balance)
func (h *FinanceHandler) GetTrialBalance(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.GetTrialBalance")
	defer span.End()
	c.SetUserContext(ctx)
	_ = ctx // Mark as used

	// Extract date parameter
	asOfDate := time.Now()
	if dateStr := c.Query("as_of_date"); dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			asOfDate = parsed
		}
	}

	// This method doesn't exist in AccountService - removing for now
	// trialBalance, err := h.services.Account.GetTrialBalance(ctx, asOfDate)
	trialBalance := map[string]interface{}{"message": "Trial balance not implemented yet"}
	err := error(nil)
	if err != nil {
		return h.HandleError(c, err)
	}

	meta := map[string]interface{}{
		"as_of_date":   asOfDate.Format("2006-01-02"),
		"generated_at": time.Now(),
	}

	return h.SuccessWithMeta(c, trialBalance, meta)
}

// GetAccountBalance handles account balance inquiry (GET /api/v1/finance/accounts/:id/balance)
func (h *FinanceHandler) GetAccountBalance(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.GetAccountBalance")
	defer span.End()
	c.SetUserContext(ctx)
	_ = ctx // Mark as used

	accountID := c.Params("id")
	_, err := uuid.Parse(accountID)
	if err != nil {
		return h.HandleError(c, errors.NewBusinessError("INVALID_ACCOUNT_ID", "Invalid account ID format").
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation))
	}

	// Extract date parameter
	asOfDate := time.Now()
	if dateStr := c.Query("as_of_date"); dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			asOfDate = parsed
		}
	}

	// This method doesn't exist in AccountService - removing for now
	// balance, err := h.services.Account.GetAccountBalance(ctx, accountUUID, asOfDate)
	balance := map[string]interface{}{"message": "Account balance not implemented yet"}
	err = error(nil)
	if err != nil {
		return h.HandleError(c, err)
	}

	meta := map[string]interface{}{
		"account_id": accountID,
		"as_of_date": asOfDate.Format("2006-01-02"),
	}

	return h.SuccessWithMeta(c, balance, meta)
}

// ============================================================================
// BASE HANDLER METHODS (following same pattern as tenant handler)
// ============================================================================

// HandleError provides centralized error handling
func (h *FinanceHandler) HandleError(c *fiber.Ctx, err error) error {
	requestID := h.getRequestID(c)
	tenantID := h.getTenantID(c)

	// Convert to HTTP error using existing system
	httpErr := errors.ToHTTPError(err)
	httpErr.RequestID = requestID

	// Log the error with context
	h.logger.Error("Finance API error", logger.Fields{
		"error":      err.Error(),
		"request_id": requestID,
		"tenant_id":  tenantID,
		"method":     c.Method(),
		"path":       c.Path(),
		"ip":         c.IP(),
		"status":     httpErr.Status,
		"code":       httpErr.Code,
	})

	// Record metrics
	h.recordErrorMetrics(c, httpErr)

	return c.Status(httpErr.Status).JSON(httpErr)
}

// ValidateRequest validates request data
func (h *FinanceHandler) ValidateRequest(c *fiber.Ctx, req interface{}) error {
	if err := c.BodyParser(req); err != nil {
		return errors.NewBusinessError("INVALID_JSON", "Invalid JSON format").
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation).
			WithSuggestion("Ensure request body contains valid JSON")
	}

	if err := h.validator.Struct(req); err != nil {
		return err
	}

	return nil
}

// Success returns a standardized success response
func (h *FinanceHandler) Success(c *fiber.Ctx, data interface{}) error {
	response := fiber.Map{
		"success":    true,
		"data":       data,
		"request_id": h.getRequestID(c),
		"timestamp":  time.Now(),
	}

	h.recordSuccessMetrics(c)
	return c.JSON(response)
}

// SuccessWithMeta returns success response with metadata
func (h *FinanceHandler) SuccessWithMeta(c *fiber.Ctx, data interface{}, meta map[string]interface{}) error {
	response := fiber.Map{
		"success":    true,
		"data":       data,
		"meta":       meta,
		"request_id": h.getRequestID(c),
		"timestamp":  time.Now(),
	}

	h.recordSuccessMetrics(c)
	return c.JSON(response)
}

// Created returns a 201 response
func (h *FinanceHandler) Created(c *fiber.Ctx, data interface{}) error {
	response := fiber.Map{
		"success":    true,
		"data":       data,
		"request_id": h.getRequestID(c),
		"timestamp":  time.Now(),
	}

	h.recordSuccessMetrics(c)
	return c.Status(fiber.StatusCreated).JSON(response)
}

// ExtractPaginationParams extracts pagination parameters
func (h *FinanceHandler) ExtractPaginationParams(c *fiber.Ctx) (offset, limit int) {
	offset = c.QueryInt("offset", 0)
	limit = c.QueryInt("limit", 20)

	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	return offset, limit
}

// Helper methods
func (h *FinanceHandler) getRequestID(c *fiber.Ctx) string {
	if requestID := c.Locals("requestid"); requestID != nil {
		return requestID.(string)
	}
	return "unknown"
}

func (h *FinanceHandler) getTenantID(c *fiber.Ctx) string {
	if tenantID := c.Locals("tenant_id"); tenantID != nil {
		return tenantID.(string)
	}
	return ""
}

func (h *FinanceHandler) extractBoolQuery(c *fiber.Ctx, key string) *bool {
	if value := c.Query(key); value != "" {
		if value == "true" {
			result := true
			return &result
		}
		if value == "false" {
			result := false
			return &result
		}
	}
	return nil
}

func (h *FinanceHandler) recordErrorMetrics(c *fiber.Ctx, httpErr *errors.HTTPError) {
	h.metrics.IncrementCounter("finance_api_errors_total", metrics.Fields{
		"method":   c.Method(),
		"endpoint": c.Route().Path,
		"code":     httpErr.Code,
		"status":   string(rune(httpErr.Status)),
	})
}

func (h *FinanceHandler) recordSuccessMetrics(c *fiber.Ctx) {
	h.metrics.IncrementCounter("finance_api_requests_total", metrics.Fields{
		"method":   c.Method(),
		"endpoint": c.Route().Path,
		"status":   "success",
	})
}

// ============================================================================
// FISCAL YEAR ENDPOINTS
// ============================================================================

// CreateFiscalYear handles POST /api/v1/finance/fiscal-years
func (h *FinanceHandler) CreateFiscalYear(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.CreateFiscalYear")
	defer span.End()
	c.SetUserContext(ctx)

	var req struct {
		Name      string    `json:"name" validate:"required"`
		StartDate time.Time `json:"start_date" validate:"required"`
		EndDate   time.Time `json:"end_date" validate:"required"`
	}
	if err := h.ValidateRequest(c, &req); err != nil {
		return h.HandleError(c, err)
	}

	fy := &financeDomain.FiscalYear{
		Name:      req.Name,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	}
	created, err := h.services.Period.CreateFiscalYear(ctx, fy)
	if err != nil {
		span.RecordError(err)
		return h.HandleError(c, err)
	}
	return h.Created(c, created)
}

// ListFiscalYears handles GET /api/v1/finance/fiscal-years
func (h *FinanceHandler) ListFiscalYears(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.ListFiscalYears")
	defer span.End()
	c.SetUserContext(ctx)

	fys, err := h.services.Period.ListFiscalYears(ctx)
	if err != nil {
		span.RecordError(err)
		return h.HandleError(c, err)
	}
	return h.Success(c, fys)
}

// GetFiscalYear handles GET /api/v1/finance/fiscal-years/:id
func (h *FinanceHandler) GetFiscalYear(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.GetFiscalYear")
	defer span.End()
	c.SetUserContext(ctx)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return h.HandleError(c, errors.NewBusinessError("INVALID_ID", "invalid fiscal year ID").WithHTTPStatus(400))
	}
	fy, err := h.services.Period.GetFiscalYearByID(ctx, id)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.Success(c, fy)
}

// ============================================================================
// ACCOUNTING PERIOD ENDPOINTS
// ============================================================================

// CreatePeriod handles POST /api/v1/finance/fiscal-years/:id/periods
func (h *FinanceHandler) CreatePeriod(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.CreatePeriod")
	defer span.End()
	c.SetUserContext(ctx)

	fyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return h.HandleError(c, errors.NewBusinessError("INVALID_ID", "invalid fiscal year ID").WithHTTPStatus(400))
	}

	var req struct {
		PeriodNumber int       `json:"period_number" validate:"required,min=1"`
		Name         string    `json:"name" validate:"required"`
		StartDate    time.Time `json:"start_date" validate:"required"`
		EndDate      time.Time `json:"end_date" validate:"required"`
	}
	if err := h.ValidateRequest(c, &req); err != nil {
		return h.HandleError(c, err)
	}

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
		return h.HandleError(c, err)
	}
	return h.Created(c, created)
}

// ListPeriods handles GET /api/v1/finance/fiscal-years/:id/periods
func (h *FinanceHandler) ListPeriods(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.ListPeriods")
	defer span.End()
	c.SetUserContext(ctx)

	fyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return h.HandleError(c, errors.NewBusinessError("INVALID_ID", "invalid fiscal year ID").WithHTTPStatus(400))
	}
	periods, err := h.services.Period.ListPeriods(ctx, fyID)
	if err != nil {
		span.RecordError(err)
		return h.HandleError(c, err)
	}
	return h.Success(c, periods)
}

// GetPeriod handles GET /api/v1/finance/periods/:id
func (h *FinanceHandler) GetPeriod(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.GetPeriod")
	defer span.End()
	c.SetUserContext(ctx)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return h.HandleError(c, errors.NewBusinessError("INVALID_ID", "invalid period ID").WithHTTPStatus(400))
	}
	period, err := h.services.Period.GetPeriodByID(ctx, id)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.Success(c, period)
}

// GetCurrentPeriod handles GET /api/v1/finance/periods/current
func (h *FinanceHandler) GetCurrentPeriod(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.GetCurrentPeriod")
	defer span.End()
	c.SetUserContext(ctx)

	period, err := h.services.Period.GetCurrentPeriod(ctx)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.Success(c, period)
}

// ChangePeriodStatus handles PATCH /api/v1/finance/periods/:id/status
func (h *FinanceHandler) ChangePeriodStatus(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.ChangePeriodStatus")
	defer span.End()
	c.SetUserContext(ctx)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return h.HandleError(c, errors.NewBusinessError("INVALID_ID", "invalid period ID").WithHTTPStatus(400))
	}

	var req struct {
		Status       string `json:"status" validate:"required"`
		ChecksPassed bool   `json:"checks_passed"`
	}
	if err := h.ValidateRequest(c, &req); err != nil {
		return h.HandleError(c, err)
	}

	newStatus := financeDomain.PeriodStatus(req.Status)
	if !newStatus.IsValid() {
		return h.HandleError(c, errors.NewBusinessError("INVALID_STATUS", "invalid period status").WithHTTPStatus(400))
	}

	// Extract user ID from context locals (set by auth middleware)
	var byUserID uuid.UUID
	if userIDStr, ok := c.Locals("user_id").(string); ok {
		byUserID, _ = uuid.Parse(userIDStr)
	}

	period, err := h.services.Period.ChangePeriodStatus(ctx, id, newStatus, byUserID, req.ChecksPassed)
	if err != nil {
		span.RecordError(err)
		return h.HandleError(c, err)
	}
	return h.Success(c, period)
}

// ============================================================================
// EXCHANGE RATE ENDPOINTS
// ============================================================================

// UpsertExchangeRate handles POST /api/v1/finance/exchange-rates
func (h *FinanceHandler) UpsertExchangeRate(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.UpsertExchangeRate")
	defer span.End()
	c.SetUserContext(ctx)

	var rate financeDomain.ExchangeRate
	if err := h.ValidateRequest(c, &rate); err != nil {
		return h.HandleError(c, err)
	}

	created, err := h.services.ExchangeRate.UpsertRate(ctx, &rate)
	if err != nil {
		span.RecordError(err)
		return h.HandleError(c, err)
	}
	return h.Created(c, created)
}

// GetExchangeRate handles GET /api/v1/finance/exchange-rates
// Query params: from, to, rate_type, date
func (h *FinanceHandler) GetExchangeRate(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.GetExchangeRate")
	defer span.End()
	c.SetUserContext(ctx)

	fromCurrency := c.Query("from")
	toCurrency := c.Query("to")
	if fromCurrency == "" || toCurrency == "" {
		return h.HandleError(c, errors.NewBusinessError("MISSING_PARAMS", "from and to query parameters are required").WithHTTPStatus(400))
	}

	rateTypeStr := c.Query("rate_type", "SPOT")
	rateType := financeDomain.RateType(rateTypeStr)
	if !rateType.IsValid() {
		rateType = financeDomain.RateTypeSpot
	}

	asOfDate := time.Now()
	if dateStr := c.Query("date"); dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			asOfDate = parsed
		}
	}

	rate, err := h.services.ExchangeRate.GetRate(ctx, fromCurrency, toCurrency, rateType, asOfDate)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.Success(c, rate)
}

// ListExchangeRates handles GET /api/v1/finance/exchange-rates/history
// Query params: from, to, from_date, to_date, limit
func (h *FinanceHandler) ListExchangeRates(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.ListExchangeRates")
	defer span.End()
	c.SetUserContext(ctx)

	fromCurrency := c.Query("from")
	toCurrency := c.Query("to")
	if fromCurrency == "" || toCurrency == "" {
		return h.HandleError(c, errors.NewBusinessError("MISSING_PARAMS", "from and to query parameters are required").WithHTTPStatus(400))
	}

	limit := c.QueryInt("limit", 100)
	var fromDate, toDate *time.Time
	if s := c.Query("from_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			fromDate = &t
		}
	}
	if s := c.Query("to_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			toDate = &t
		}
	}

	rates, err := h.services.ExchangeRate.ListRates(ctx, fromCurrency, toCurrency, fromDate, toDate, limit)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.Success(c, rates)
}

// ============================================================================
// CURRENCY ENDPOINTS
// ============================================================================

// CreateCurrency handles POST /api/v1/finance/currencies
func (h *FinanceHandler) CreateCurrency(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.CreateCurrency")
	defer span.End()
	c.SetUserContext(ctx)

	var currency financeDomain.Currency
	if err := h.ValidateRequest(c, &currency); err != nil {
		return h.HandleError(c, err)
	}

	created, err := h.services.Currency.CreateCurrency(ctx, &currency)
	if err != nil {
		span.RecordError(err)
		return h.HandleError(c, err)
	}
	return h.Created(c, created)
}

// ListCurrencies handles GET /api/v1/finance/currencies
func (h *FinanceHandler) ListCurrencies(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.ListCurrencies")
	defer span.End()
	c.SetUserContext(ctx)

	activeOnly := c.QueryBool("active_only", false)
	currencies, err := h.services.Currency.ListCurrencies(ctx, activeOnly)
	if err != nil {
		span.RecordError(err)
		return h.HandleError(c, err)
	}
	return h.Success(c, currencies)
}

// UpdateCurrency handles PUT /api/v1/finance/currencies/:id
func (h *FinanceHandler) UpdateCurrency(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.UpdateCurrency")
	defer span.End()
	c.SetUserContext(ctx)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return h.HandleError(c, errors.NewBusinessError("INVALID_ID", "invalid currency ID").WithHTTPStatus(400))
	}

	var currency financeDomain.Currency
	if err := h.ValidateRequest(c, &currency); err != nil {
		return h.HandleError(c, err)
	}
	currency.ID = id

	updated, err := h.services.Currency.UpdateCurrency(ctx, &currency)
	if err != nil {
		span.RecordError(err)
		return h.HandleError(c, err)
	}
	return h.Success(c, updated)
}

// ============================================================================
// APPROVAL WORKFLOW ENDPOINTS
// ============================================================================

// SubmitTransaction handles POST /api/v1/finance/transactions/:id/submit
func (h *FinanceHandler) SubmitTransaction(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.SubmitTransaction")
	defer span.End()
	c.SetUserContext(ctx)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return h.HandleError(c, errors.NewBusinessError("INVALID_ID", "invalid transaction ID").WithHTTPStatus(400))
	}

	transaction, err := h.services.Transaction.PostTransaction(ctx, id, nil)
	if err != nil {
		span.RecordError(err)
		return h.HandleError(c, err)
	}
	return h.Success(c, transaction)
}

// ApproveTransaction handles POST /api/v1/finance/transactions/:id/approve
func (h *FinanceHandler) ApproveTransaction(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.ApproveTransaction")
	defer span.End()
	c.SetUserContext(ctx)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return h.HandleError(c, errors.NewBusinessError("INVALID_ID", "invalid transaction ID").WithHTTPStatus(400))
	}

	var req struct {
		Notes string `json:"notes"`
	}
	_ = c.BodyParser(&req)

	transaction, err := h.services.Transaction.ApproveTransaction(ctx, id, req.Notes)
	if err != nil {
		span.RecordError(err)
		return h.HandleError(c, err)
	}
	return h.Success(c, transaction)
}

// RejectTransaction handles POST /api/v1/finance/transactions/:id/reject
func (h *FinanceHandler) RejectTransaction(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.RejectTransaction")
	defer span.End()
	c.SetUserContext(ctx)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return h.HandleError(c, errors.NewBusinessError("INVALID_ID", "invalid transaction ID").WithHTTPStatus(400))
	}

	var req struct {
		Notes string `json:"notes"`
	}
	_ = c.BodyParser(&req)

	transaction, err := h.services.Transaction.RejectTransaction(ctx, id, req.Notes)
	if err != nil {
		span.RecordError(err)
		return h.HandleError(c, err)
	}
	return h.Success(c, transaction)
}

// ============================================================================
// COST CENTER ENDPOINTS
// ============================================================================

// CreateCostCenter handles POST /api/v1/finance/cost-centers
func (h *FinanceHandler) CreateCostCenter(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.CreateCostCenter")
	defer span.End()
	c.SetUserContext(ctx)

	var req struct {
		Code             string  `json:"code" validate:"required"`
		Name             string  `json:"name" validate:"required"`
		Description      string  `json:"description"`
		ParentID         *string `json:"parent_id"`
		IsGroup          bool    `json:"is_group"`
		IsDistributed    bool    `json:"is_distributed"`
		AllocationMethod *string `json:"allocation_method"`
		IsActive         bool    `json:"is_active"`
	}
	if err := h.ValidateRequest(c, &req); err != nil {
		return h.HandleError(c, err)
	}

	cc := &financeDomain.CostCenter{
		Code:          req.Code,
		Name:          req.Name,
		Description:   req.Description,
		IsGroup:       req.IsGroup,
		IsDistributed: req.IsDistributed,
		IsActive:      req.IsActive,
	}
	if req.ParentID != nil {
		id, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return h.HandleError(c, errors.NewBusinessError("INVALID_ID", "invalid parent_id").WithHTTPStatus(400))
		}
		cc.ParentID = &id
	}
	if req.AllocationMethod != nil {
		m := financeDomain.AllocationMethod(*req.AllocationMethod)
		cc.AllocationMethod = &m
	}

	created, err := h.services.CostCenter.Create(ctx, cc)
	if err != nil {
		span.RecordError(err)
		return h.HandleError(c, err)
	}
	return h.Created(c, created)
}

// GetCostCenter handles GET /api/v1/finance/cost-centers/:id
func (h *FinanceHandler) GetCostCenter(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.GetCostCenter")
	defer span.End()
	c.SetUserContext(ctx)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return h.HandleError(c, errors.NewBusinessError("INVALID_ID", "invalid cost centre ID").WithHTTPStatus(400))
	}
	cc, err := h.services.CostCenter.GetByID(ctx, id)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.Success(c, cc)
}

// ListCostCenters handles GET /api/v1/finance/cost-centers
func (h *FinanceHandler) ListCostCenters(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.ListCostCenters")
	defer span.End()
	c.SetUserContext(ctx)

	activeOnly := c.QueryBool("active_only", false)
	ccs, err := h.services.CostCenter.List(ctx, activeOnly)
	if err != nil {
		span.RecordError(err)
		return h.HandleError(c, err)
	}
	return h.Success(c, ccs)
}

// UpdateCostCenter handles PUT /api/v1/finance/cost-centers/:id
func (h *FinanceHandler) UpdateCostCenter(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.UpdateCostCenter")
	defer span.End()
	c.SetUserContext(ctx)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return h.HandleError(c, errors.NewBusinessError("INVALID_ID", "invalid cost centre ID").WithHTTPStatus(400))
	}

	var req struct {
		Code             string  `json:"code" validate:"required"`
		Name             string  `json:"name" validate:"required"`
		Description      string  `json:"description"`
		ParentID         *string `json:"parent_id"`
		IsGroup          bool    `json:"is_group"`
		IsDistributed    bool    `json:"is_distributed"`
		AllocationMethod *string `json:"allocation_method"`
		IsActive         bool    `json:"is_active"`
	}
	if err := h.ValidateRequest(c, &req); err != nil {
		return h.HandleError(c, err)
	}

	cc := &financeDomain.CostCenter{
		ID:            id,
		Code:          req.Code,
		Name:          req.Name,
		Description:   req.Description,
		IsGroup:       req.IsGroup,
		IsDistributed: req.IsDistributed,
		IsActive:      req.IsActive,
	}
	if req.ParentID != nil {
		pid, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return h.HandleError(c, errors.NewBusinessError("INVALID_ID", "invalid parent_id").WithHTTPStatus(400))
		}
		cc.ParentID = &pid
	}
	if req.AllocationMethod != nil {
		m := financeDomain.AllocationMethod(*req.AllocationMethod)
		cc.AllocationMethod = &m
	}

	updated, err := h.services.CostCenter.Update(ctx, cc)
	if err != nil {
		span.RecordError(err)
		return h.HandleError(c, err)
	}
	return h.Success(c, updated)
}

// DeleteCostCenter handles DELETE /api/v1/finance/cost-centers/:id
func (h *FinanceHandler) DeleteCostCenter(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "finance.DeleteCostCenter")
	defer span.End()
	c.SetUserContext(ctx)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return h.HandleError(c, errors.NewBusinessError("INVALID_ID", "invalid cost centre ID").WithHTTPStatus(400))
	}
	if err := h.services.CostCenter.Delete(ctx, id); err != nil {
		span.RecordError(err)
		return h.HandleError(c, err)
	}
	return h.Success(c, fiber.Map{"deleted": true})
}

// Custom validator registration
func registerCustomValidators(v *validator.Validate) {
	v.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		if value == "" {
			return true
		}
		return len(value) == 36 && value[8] == '-' && value[13] == '-' && value[18] == '-' && value[23] == '-'
	})
}
