package finance

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	financeDomain "github.com/niiniyare/erp/internal/core/finance/domain"
	financeService "github.com/niiniyare/erp/internal/core/finance/service"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
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

	transaction, err := h.services.Transaction.Create(c.Context(), &req)
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

	transaction, err := h.services.Transaction.GetByID(ctx, transactionUUID)
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
	transactions, err := h.services.Transaction.List(c.Context(), &filters)
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
