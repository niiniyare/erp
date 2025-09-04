// Package finance provides comprehensive financial services and utilities
// for enterprise-grade accounting and financial management systems.
//
// This package implements a Facade Pattern to provide a unified interface to all
// finance module functionality, including:
// - Chart of Accounts management: hierarchical account structures and validation
// - Transaction processing: double-entry bookkeeping with multi-currency support
// - Financial reporting: balance sheets, income statements, and cash flow reports
// - Exchange rate management: real-time currency conversion and historical rates
// - Period closing: month-end and year-end accounting processes
// - Audit trail: comprehensive transaction logging and compliance tracking
//
// The service facade integrates with essential platform services (IAM, Settings,
// Feature Flags, Audit) and follows Clean Architecture principles with Temporal
// workflow orchestration for complex business processes.
//
// All operations are multi-tenant aware with entity-level data isolation for
// companies within tenants, supporting both GAAP and IFRS accounting standards.
package finance

//
// import (
// 	"context"
// 	"time"
//
// 	"github.com/google/uuid"
// 	"github.com/niiniyare/erp/internal/core/audit"
// 	"github.com/niiniyare/erp/internal/core/featureflag"
// 	"github.com/niiniyare/erp/internal/core/finance/domain"
// 	"github.com/niiniyare/erp/internal/core/finance/service"
// 	"github.com/niiniyare/erp/internal/core/iam"
// 	settingsService "github.com/niiniyare/erp/internal/core/settings/service"
// 	"github.com/niiniyare/erp/internal/platform/cache"
// 	"github.com/niiniyare/erp/internal/shared/errors"
// 	"github.com/niiniyare/erp/internal/shared/logger"
// 	"github.com/niiniyare/erp/internal/shared/metrics"
// 	"github.com/niiniyare/erp/internal/shared/tracing"
//
// 	"go.temporal.io/sdk/client"
// )
//
// FinanceService provides the main facade interface for all finance module operations.
// This is the primary entry point that external modules should use to interact
// with finance functionality. It implements the Facade Pattern to hide the
// complexity of individual feature services and provide a clean, unified API.
// type FinanceService interface {
// 	// Account Management Operations
// 	// Provides comprehensive chart of accounts functionality with hierarchical support
// 	CreateAccount(ctx context.Context, req *service.CreateAccountRequestt) (*domain.Accounts, error)
// 	GetAccount(ctx context.Context, accountID uuid.UUID) (*domain.Accounts, error)
// 	GetAccountByCode(ctx context.Context, accountCode string) (*domain.Accounts, error)
// 	UpdateAccount(ctx context.Context, req *service.UpdateAccountRequest) (*domain.Account, error)
// 	DeactivateAccount(ctx context.Context, accountID string) error
// 	ListAccounts(ctx context.Context, req *service.ListAccountsRequest) (*service.ListAccountsResponse, error)
// 	GetAccountBalance(ctx context.Context, accountID string, asOfDate string) (*domain.AccountBalance, error)
// 	GetAccountHierarchy(ctx context.Context) ([]*domain.Accounts, error)
//
// 	// Transaction Processing Operations
// 	// Handles double-entry bookkeeping with validation and approval workflows
// 	CreateTransaction(ctx context.Context, req *service.CreateTransactionRequest) (*domain.Transaction, error)
// 	GetTransaction(ctx context.Context, transactionID string) (*domain.Transaction, error)
// 	UpdateTransaction(ctx context.Context, req *service.UpdateTransactionRequest) (*domain.Transaction, error)
// 	PostTransaction(ctx context.Context, transactionID string) error
// 	ReverseTransaction(ctx context.Context, transactionID, reason string) (*domain.Transaction, error)
// 	ListTransactions(ctx context.Context, req *service.ListTransactionsRequest) (*service.ListTransactionsResponse, error)
// 	BulkImportTransactions(ctx context.Context, req *service.BulkImportRequest) (*service.BulkImportResponse, error)
//
// 	// Financial Reporting Operations
// 	// Generates standard financial statements and custom reports
// 	GenerateTrialBalance(ctx context.Context, req *service.TrialBalanceRequest) (*service.TrialBalanceResponse, error)
// 	GenerateIncomeStatement(ctx context.Context, req *service.IncomeStatementRequest) (*service.IncomeStatementResponse, error)
// 	GenerateBalanceSheet(ctx context.Context, req *service.BalanceSheetRequest) (*service.BalanceSheetResponse, error)
// 	GenerateCashFlowStatement(ctx context.Context, req *service.CashFlowRequest) (*service.CashFlowResponse, error)
//
// 	// Exchange Rate Operations
// 	// Manages multi-currency support with real-time and historical rates
// 	SetExchangeRate(ctx context.Context, req *service.SetExchangeRateRequest) (*domain.ExchangeRate, error)
// 	GetExchangeRate(ctx context.Context, fromCurrency, toCurrency, date string) (*domain.ExchangeRate, error)
// 	ListExchangeRates(ctx context.Context, req *service.ListExchangeRatesRequest) (*service.ListExchangeRatesResponse, error)
//
// 	// Period Management Operations
// 	// Handles month-end closing and period management processes
// 	ClosePeriod(ctx context.Context, req *service.ClosePeriodRequest) error
// 	ReopenPeriod(ctx context.Context, req *service.ReopenPeriodRequest) error
// 	GetPeriodStatus(ctx context.Context, year int, month int) (*domain.PeriodStatus, error)
//
// 	// Health and Status Operations
// 	// Provides module health checks and status information
// 	HealthCheck(ctx context.Context) error
// 	GetModuleInfo(ctx context.Context) (*service.ModuleInfo, error)
// }
//
// // financeService implements the FinanceService interface using the Facade Pattern.
// // It coordinates between individual feature services and manages cross-cutting
// // concerns like caching, auditing, and workflow orchestration.
// type financeService struct {
// 	// Core feature services - these handle specific domain areas
// 	accountService      service.AccountService
// 	transactionService  service.TransactionService
// 	reportingService    service.ReportingService
// 	exchangeRateService service.ExchangeRateService
// 	periodService       service.PeriodService
//
// 	// Platform service dependencies for cross-cutting concerns
// 	cache        cache.Service       // Multi-tenant caching with automatic key prefixing
// 	iam          iam.Service         // Identity and access management
// 	settings     settings.Service    // Configuration and settings management
// 	featureFlags featureflag.Service // Feature flag management for gradual rollouts
// 	audit        audit.Service       // Comprehensive audit trail logging
// 	metrics      metrics.Provider    // Performance metrics and monitoring
// 	tracing      tracing.Service     // Distributed tracing for observability
// 	logger       logger.Logger       // Structured logging with context
//
// 	// Temporal workflow client for orchestrating complex business processes
// 	temporalClient client.Client // Temporal workflow orchestration
// }
//
// // Dependencies contains all required dependencies for creating the finance service.
// // This follows dependency injection principles and makes testing easier by
// // allowing mock implementations of external services.
// type Dependencies struct {
// 	// Data layer dependencies
// 	AccountRepo      domain.AccountRepository      // Account data persistence
// 	TransactionRepo  domain.TransactionRepository  // Transaction data persistence
// 	ExchangeRateRepo domain.ExchangeRateRepository // Exchange rate data persistence
//
// 	// Platform service dependencies
// 	Cache        cache.Service       // Multi-tenant Redis cache service
// 	IAM          iam.Service         // Identity and access management
// 	Settings     settings.Service    // Configuration management
// 	FeatureFlags featureflag.Service // Feature toggle management
// 	Audit        audit.Service       // Audit trail logging service
// 	Metrics      metrics.Provider    // Metrics collection and reporting
// 	Tracing      tracing.Service     // Distributed tracing service
// 	Logger       logger.Logger       // Structured logging service
//
// 	// Workflow orchestration
// 	TemporalClient client.Client // Temporal workflow client
// }
//
// // NewFinanceService creates a new finance service instance with all dependencies.
// // This function follows the Facade Pattern by creating individual feature services
// // and composing them into a single, unified interface.
// //
// // The service automatically handles:
// // - Multi-tenant data isolation using context-based tenant/entity information
// // - Cross-cutting concerns like audit logging, caching, and metrics collection
// // - Feature flag evaluation for gradual feature rollouts
// // - Integration with external services (IAM, Settings, etc.)
// func NewFinanceService(deps Dependencies) (FinanceService, error) {
// 	// Validate all required dependencies are provided
// 	if err := deps.Validate(); err != nil {
// 		return nil, errors.Wrap(err, "failed to validate finance service dependencies")
// 	}
//
// 	// Create individual feature services with shared dependencies
// 	accountService := service.NewAccountService(service.AccountServiceDeps{
// 		Repo:         deps.AccountRepo,
// 		Cache:        deps.Cache,
// 		IAM:          deps.IAM,
// 		FeatureFlags: deps.FeatureFlags,
// 		Audit:        deps.Audit,
// 		Metrics:      deps.Metrics,
// 		Tracing:      deps.Tracing,
// 		Logger:       deps.Logger,
// 	})
//
// 	transactionService := service.NewTransactionService(service.TransactionServiceDeps{
// 		TransactionRepo: deps.TransactionRepo,
// 		AccountRepo:     deps.AccountRepo,
// 		Cache:           deps.Cache,
// 		IAM:             deps.IAM,
// 		FeatureFlags:    deps.FeatureFlags,
// 		Audit:           deps.Audit,
// 		Metrics:         deps.Metrics,
// 		Tracing:         deps.Tracing,
// 		Logger:          deps.Logger,
// 	})
//
// 	reportingService := service.NewReportingService(service.ReportingServiceDeps{
// 		TransactionRepo: deps.TransactionRepo,
// 		AccountRepo:     deps.AccountRepo,
// 		Cache:           deps.Cache,
// 		Settings:        deps.Settings,
// 		FeatureFlags:    deps.FeatureFlags,
// 		Metrics:         deps.Metrics,
// 		Tracing:         deps.Tracing,
// 		Logger:          deps.Logger,
// 	})
//
// 	exchangeRateService := service.NewExchangeRateService(service.ExchangeRateServiceDeps{
// 		Repo:         deps.ExchangeRateRepo,
// 		Cache:        deps.Cache,
// 		Settings:     deps.Settings,
// 		FeatureFlags: deps.FeatureFlags,
// 		Audit:        deps.Audit,
// 		Metrics:      deps.Metrics,
// 		Tracing:      deps.Tracing,
// 		Logger:       deps.Logger,
// 	})
//
// 	periodService := service.NewPeriodService(service.PeriodServiceDeps{
// 		TransactionRepo: deps.TransactionRepo,
// 		AccountRepo:     deps.AccountRepo,
// 		Cache:           deps.Cache,
// 		Settings:        deps.Settings,
// 		FeatureFlags:    deps.FeatureFlags,
// 		Audit:           deps.Audit,
// 		Metrics:         deps.Metrics,
// 		Tracing:         deps.Tracing,
// 		Logger:          deps.Logger,
// 	})
//
// 	return &financeService{
// 		// Assign feature services
// 		accountService:      accountService,
// 		transactionService:  transactionService,
// 		reportingService:    reportingService,
// 		exchangeRateService: exchangeRateService,
// 		periodService:       periodService,
//
// 		// Assign platform services
// 		cache:        deps.Cache,
// 		iam:          deps.IAM,
// 		settings:     deps.Settings,
// 		featureFlags: deps.FeatureFlags,
// 		audit:        deps.Audit,
// 		metrics:      deps.Metrics,
// 		tracing:      deps.Tracing,
// 		logger:       deps.Logger,
//
// 		// Assign workflow orchestration
// 		temporalClient: deps.TemporalClient,
// 	}, nil
// }
//
// // Validate ensures all required dependencies are provided for the finance service.
// // This helps catch configuration errors at startup rather than runtime.
// func (d Dependencies) Validate() error {
// 	// Validate data layer dependencies
// 	if d.AccountRepo == nil {
// 		return errors.NewValidationError("account repository is required")
// 	}
// 	if d.TransactionRepo == nil {
// 		return errors.NewValidationError("transaction repository is required")
// 	}
// 	if d.ExchangeRateRepo == nil {
// 		return errors.NewValidationError("exchange rate repository is required")
// 	}
//
// 	// Validate platform service dependencies
// 	if d.Cache == nil {
// 		return errors.NewValidationError("cache service is required")
// 	}
// 	if d.IAM == nil {
// 		return errors.NewValidationError("IAM service is required")
// 	}
// 	if d.Settings == nil {
// 		return errors.NewValidationError("settings service is required")
// 	}
// 	if d.FeatureFlags == nil {
// 		return errors.NewValidationError("feature flags service is required")
// 	}
// 	if d.Audit == nil {
// 		return errors.NewValidationError("audit service is required")
// 	}
// 	if d.Metrics == nil {
// 		return errors.NewValidationError("metrics provider is required")
// 	}
// 	if d.Tracing == nil {
// 		return errors.NewValidationError("tracing service is required")
// 	}
// 	if d.Logger == nil {
// 		return errors.NewValidationError("logger is required")
// 	}
//
// 	// Validate workflow orchestration dependency
// 	if d.TemporalClient == nil {
// 		return errors.NewValidationError("Temporal client is required")
// 	}
//
// 	return nil
// }
//
// // ========================================
// // ACCOUNT MANAGEMENT OPERATIONS
// // ========================================
//
// // CreateAccount creates a new account in the chart of accounts with validation and audit trail.
// // This method delegates to the account service while providing cross-cutting concerns like
// // audit logging, metrics collection, and caching.
// func (s *financeService) CreateAccount(ctx context.Context, req *service.CreateAccountRequest) (*domain.Account, error) {
// 	// Start tracing span for observability
// 	span, ctx := s.tracing.StartSpan(ctx, "finance.CreateAccount")
// 	defer span.End()
//
// 	// Record metrics for monitoring
// 	s.metrics.Counter(domain.MetricAccountsCreated).Inc()
//
// 	// Delegate to account service which handles the actual business logic
// 	account, err := s.accountService.CreateAccount(ctx, req)
// 	if err != nil {
// 		s.metrics.Counter(domain.MetricValidationErrors).Inc()
// 		return nil, err
// 	}
//
// 	// Log audit event for compliance tracking
// 	s.audit.LogEvent(ctx, domain.AuditEventAccountCreated, map[string]interface{}{
// 		"account_id":   account.ID,
// 		"account_code": account.Code,
// 		"account_type": account.Type,
// 	})
//
// 	return account, nil
// }
//
// // GetAccount retrieves an account by ID with caching support for performance.
// func (s *financeService) GetAccount(ctx context.Context, accountID string) (*domain.Account, error) {
// 	span, ctx := s.tracing.StartSpan(ctx, "finance.GetAccount")
// 	defer span.End()
//
// 	return s.accountService.GetAccount(ctx, accountID)
// }
//
// // GetAccountByCode retrieves an account by code with caching and validation.
// func (s *financeService) GetAccountByCode(ctx context.Context, accountCode string) (*domain.Account, error) {
// 	span, ctx := s.tracing.StartSpan(ctx, "finance.GetAccountByCode")
// 	defer span.End()
//
// 	return s.accountService.GetAccountByCode(ctx, accountCode)
// }
//
// // UpdateAccount updates an existing account with validation and audit trail.
// func (s *financeService) UpdateAccount(ctx context.Context, req *service.UpdateAccountRequest) (*domain.Account, error) {
// 	span, ctx := s.tracing.StartSpan(ctx, "finance.UpdateAccount")
// 	defer span.End()
//
// 	account, err := s.accountService.UpdateAccount(ctx, req)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	// Log audit event for account updates
// 	s.audit.LogEvent(ctx, domain.AuditEventAccountUpdated, map[string]interface{}{
// 		"account_id": account.ID,
// 		"changes":    req, // Include the update request for audit trail
// 	})
//
// 	return account, nil
// }
//
// // DeactivateAccount deactivates an account with proper validation and audit logging.
// func (s *financeService) DeactivateAccount(ctx context.Context, accountID string) error {
// 	span, ctx := s.tracing.StartSpan(ctx, "finance.DeactivateAccount")
// 	defer span.End()
//
// 	err := s.accountService.DeactivateAccount(ctx, accountID)
// 	if err != nil {
// 		return err
// 	}
//
// 	// Log audit event for account deactivation
// 	s.audit.LogEvent(ctx, domain.AuditEventAccountDeactivated, map[string]interface{}{
// 		"account_id": accountID,
// 	})
//
// 	return nil
// }
//
// // ListAccounts retrieves a list of accounts with filtering and pagination.
// func (s *financeService) ListAccounts(ctx context.Context, req *service.ListAccountsRequest) (*service.ListAccountsResponse, error) {
// 	span, ctx := s.tracing.StartSpan(ctx, "finance.ListAccounts")
// 	defer span.End()
//
// 	return s.accountService.ListAccounts(ctx, req)
// }
//
// // GetAccountBalance retrieves the current or historical balance for an account.
// func (s *financeService) GetAccountBalance(ctx context.Context, accountID string, asOfDate string) (*domain.AccountBalance, error) {
// 	span, ctx := s.tracing.StartSpan(ctx, "finance.GetAccountBalance")
// 	defer span.End()
//
// 	return s.accountService.GetAccountBalance(ctx, accountID, asOfDate)
// }
//
// // GetAccountHierarchy retrieves the complete chart of accounts hierarchy.
// func (s *financeService) GetAccountHierarchy(ctx context.Context) ([]*domain.Account, error) {
// 	span, ctx := s.tracing.StartSpan(ctx, "finance.GetAccountHierarchy")
// 	defer span.End()
//
// 	return s.accountService.GetAccountHierarchy(ctx)
// }
//
// // ========================================
// // TRANSACTION PROCESSING OPERATIONS
// // ========================================
//
// // CreateTransaction creates a new financial transaction with double-entry validation.
// func (s *financeService) CreateTransaction(ctx context.Context, req *service.CreateTransactionRequest) (*domain.Transaction, error) {
// 	span, ctx := s.tracing.StartSpan(ctx, "finance.CreateTransaction")
// 	defer span.End()
//
// 	s.metrics.Counter(domain.MetricTransactionsProcessed).Inc()
//
// 	transaction, err := s.transactionService.CreateTransaction(ctx, req)
// 	if err != nil {
// 		s.metrics.Counter(domain.MetricValidationErrors).Inc()
// 		return nil, err
// 	}
//
// 	// Log audit event for transaction creation
// 	s.audit.LogEvent(ctx, domain.AuditEventTransactionCreated, map[string]interface{}{
// 		"transaction_id":   transaction.ID,
// 		"transaction_type": transaction.Type,
// 		"amount":          transaction.Amount,
// 		"currency":        transaction.Currency,
// 	})
//
// 	return transaction, nil
// }
//
// // GetTransaction retrieves a transaction by ID with entry details.
// func (s *financeService) GetTransaction(ctx context.Context, transactionID string) (*domain.Transaction, error) {
// 	span, ctx := s.tracing.StartSpan(ctx, "finance.GetTransaction")
// 	defer span.End()
//
// 	return s.transactionService.GetTransaction(ctx, transactionID)
// }
//
// // UpdateTransaction updates an existing transaction with validation.
// func (s *financeService) UpdateTransaction(ctx context.Context, req *service.UpdateTransactionRequest) (*domain.Transaction, error) {
// 	span, ctx := s.tracing.StartSpan(ctx, "finance.UpdateTransaction")
// 	defer span.End()
//
// 	transaction, err := s.transactionService.UpdateTransaction(ctx, req)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	s.audit.LogEvent(ctx, domain.AuditEventTransactionUpdated, map[string]interface{}{
// 		"transaction_id": transaction.ID,
// 		"changes":       req,
// 	})
//
// 	return transaction, nil
// }
//
// // PostTransaction posts a transaction to the general ledger.
// func (s *financeService) PostTransaction(ctx context.Context, transactionID string) error {
// 	span, ctx := s.tracing.StartSpan(ctx, "finance.PostTransaction")
// 	defer span.End()
//
// 	err := s.transactionService.PostTransaction(ctx, transactionID)
// 	if err != nil {
// 		return err
// 	}
//
// 	s.audit.LogEvent(ctx, domain.AuditEventTransactionPosted, map[string]interface{}{
// 		"transaction_id": transactionID,
// 	})
//
// 	return nil
// }
//
// // ReverseTransaction creates a reversal transaction for the specified transaction.
// func (s *financeService) ReverseTransaction(ctx context.Context, transactionID, reason string) (*domain.Transaction, error) {
// 	span, ctx := s.tracing.StartSpan(ctx, "finance.ReverseTransaction")
// 	defer span.End()
//
// 	reversal, err := s.transactionService.ReverseTransaction(ctx, transactionID, reason)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	s.audit.LogEvent(ctx, domain.AuditEventTransactionReversed, map[string]interface{}{
// 		"original_transaction_id": transactionID,
// 		"reversal_transaction_id": reversal.ID,
// 		"reason":                 reason,
// 	})
//
// 	return reversal, nil
// }
//
// // ListTransactions retrieves a list of transactions with filtering and pagination.
// func (s *financeService) ListTransactions(ctx context.Context, req *service.ListTransactionsRequest) (*service.ListTransactionsResponse, error) {
// 	span, ctx := s.tracing.StartSpan(ctx, "finance.ListTransactions")
// 	defer span.End()
//
// 	return s.transactionService.ListTransactions(ctx, req)
// }
//
// // BulkImportTransactions imports multiple transactions in batch with validation.
// func (s *financeService) BulkImportTransactions(ctx context.Context, req *service.BulkImportRequest) (*service.BulkImportResponse, error) {
// 	span, ctx := s.tracing.StartSpan(ctx, "finance.BulkImportTransactions")
// 	defer span.End()
//
// 	return s.transactionService.BulkImportTransactions(ctx, req)
// }
//
// // ========================================
// // FINANCIAL REPORTING OPERATIONS
// // ========================================
//
// // GenerateTrialBalance generates a trial balance report for the specified period.
// func (s *financeService) GenerateTrialBalance(ctx context.Context, req *service.TrialBalanceRequest) (*service.TrialBalanceResponse, error) {
// 	span, ctx := s.tracing.StartSpan(ctx, "finance.GenerateTrialBalance")
// 	defer span.End()
//
// 	return s.reportingService.GenerateTrialBalance(ctx, req)
// }
//
// // GenerateIncomeStatement generates an income statement for the specified period.
// func (s *financeService) GenerateIncomeStatement(ctx context.Context, req *service.IncomeStatementRequest) (*service.IncomeStatementResponse, error) {
// 	span, ctx := s.tracing.StartSpan(ctx, "finance.GenerateIncomeStatement")
// 	defer span.End()
//
// 	return s.reportingService.GenerateIncomeStatement(ctx, req)
// }
//
// // GenerateBalanceSheet generates a balance sheet as of the specified date.
// func (s *financeService) GenerateBalanceSheet(ctx context.Context, req *service.BalanceSheetRequest) (*service.BalanceSheetResponse, error) {
// 	span, ctx := s.tracing.StartSpan(ctx, "finance.GenerateBalanceSheet")
// 	defer span.End()
//
// 	return s.reportingService.GenerateBalanceSheet(ctx, req)
// }
//
// // GenerateCashFlowStatement generates a cash flow statement for the specified period.
// func (s *financeService) GenerateCashFlowStatement(ctx context.Context, req *service.CashFlowRequest) (*service.CashFlowResponse, error) {
// 	span, ctx := s.tracing.StartSpan(ctx, "finance.GenerateCashFlowStatement")
// 	defer span.End()
//
// 	return s.reportingService.GenerateCashFlowStatement(ctx, req)
// }
//
// // ========================================
// // EXCHANGE RATE OPERATIONS
// // ========================================
//
// // SetExchangeRate sets or updates an exchange rate for currency conversion.
// func (s *financeService) SetExchangeRate(ctx context.Context, req *service.SetExchangeRateRequest) (*domain.ExchangeRate, error) {
// 	span, ctx := s.tracing.StartSpan(ctx, "finance.SetExchangeRate")
// 	defer span.End()
//
// 	return s.exchangeRateService.SetExchangeRate(ctx, req)
// }
//
// // GetExchangeRate retrieves an exchange rate for currency conversion.
// func (s *financeService) GetExchangeRate(ctx context.Context, fromCurrency, toCurrency, date string) (*domain.ExchangeRate, error) {
// 	span, ctx := s.tracing.StartSpan(ctx, "finance.GetExchangeRate")
// 	defer span.End()
//
// 	return s.exchangeRateService.GetExchangeRate(ctx, fromCurrency, toCurrency, date)
// }
//
// // ListExchangeRates retrieves a list of exchange rates with filtering.
// func (s *financeService) ListExchangeRates(ctx context.Context, req *service.ListExchangeRatesRequest) (*service.ListExchangeRatesResponse, error) {
// 	span, ctx := s.tracing.StartSpan(ctx, "finance.ListExchangeRates")
// 	defer span.End()
//
// 	return s.exchangeRateService.ListExchangeRates(ctx, req)
// }
//
// // ========================================
// // PERIOD MANAGEMENT OPERATIONS
// // ========================================
//
// // ClosePeriod closes a financial period and prevents further transactions.
// func (s *financeService) ClosePeriod(ctx context.Context, req *service.ClosePeriodRequest) error {
// 	span, ctx := s.tracing.StartSpan(ctx, "finance.ClosePeriod")
// 	defer span.End()
//
// 	err := s.periodService.ClosePeriod(ctx, req)
// 	if err != nil {
// 		return err
// 	}
//
// 	s.audit.LogEvent(ctx, domain.AuditEventPeriodClosed, map[string]interface{}{
// 		"year":  req.Year,
// 		"month": req.Month,
// 	})
//
// 	return nil
// }
//
// // ReopenPeriod reopens a closed financial period to allow additional transactions.
// func (s *financeService) ReopenPeriod(ctx context.Context, req *service.ReopenPeriodRequest) error {
// 	span, ctx := s.tracing.StartSpan(ctx, "finance.ReopenPeriod")
// 	defer span.End()
//
// 	err := s.periodService.ReopenPeriod(ctx, req)
// 	if err != nil {
// 		return err
// 	}
//
// 	s.audit.LogEvent(ctx, domain.AuditEventPeriodReopened, map[string]interface{}{
// 		"year":  req.Year,
// 		"month": req.Month,
// 	})
//
// 	return nil
// }
//
// // GetPeriodStatus retrieves the status of a financial period (open/closed).
// func (s *financeService) GetPeriodStatus(ctx context.Context, year int, month int) (*domain.PeriodStatus, error) {
// 	span, ctx := s.tracing.StartSpan(ctx, "finance.GetPeriodStatus")
// 	defer span.End()
//
// 	return s.periodService.GetPeriodStatus(ctx, year, month)
// }
//
// // ========================================
// // HEALTH AND STATUS OPERATIONS
// // ========================================
//
// // HealthCheck performs a health check of the finance module and its dependencies.
// func (s *financeService) HealthCheck(ctx context.Context) error {
// 	span, ctx := s.tracing.StartSpan(ctx, "finance.HealthCheck")
// 	defer span.End()
//
// 	// Check cache service health
// 	if err := s.cache.Ping(ctx); err != nil {
// 		return errors.Wrap(err, "cache service unhealthy")
// 	}
//
// 	// Check Temporal client connectivity
// 	_, err := s.temporalClient.CheckHealth(ctx, nil)
// 	if err != nil {
// 		return errors.Wrap(err, "Temporal client unhealthy")
// 	}
//
// 	return nil
// }
//
// // GetModuleInfo returns information about the finance module including version and capabilities.
// func (s *financeService) GetModuleInfo(ctx context.Context) (*service.ModuleInfo, error) {
// 	span, ctx := s.tracing.StartSpan(ctx, "finance.GetModuleInfo")
// 	defer span.End()
//
// 	return &service.ModuleInfo{
// 		Module:           "finance",
// 		Version:          domain.ModuleVersion,
// 		APIVersion:       domain.APIVersion,
// 		Features:         s.getEnabledFeatures(ctx),
// 		HealthStatus:     "healthy", // Could be enhanced with actual health checks
// 		LastHealthCheck:  time.Now(),
// 		Dependencies:     []string{"cache", "iam", "settings", "feature-flags", "audit", "temporal"},
// 	}, nil
// }
//
// // getEnabledFeatures checks which finance features are enabled via feature flags.
// func (s *financeService) getEnabledFeatures(ctx context.Context) []string {
// 	features := []string{}
//
// 	// Check each feature flag to determine enabled capabilities
// 	if enabled, _ := s.featureFlags.IsEnabled(ctx, domain.FeatureFlagMultiCurrency); enabled {
// 		features = append(features, "multi-currency")
// 	}
// 	if enabled, _ := s.featureFlags.IsEnabled(ctx, domain.FeatureFlagAdvancedReporting); enabled {
// 		features = append(features, "advanced-reporting")
// 	}
// 	if enabled, _ := s.featureFlags.IsEnabled(ctx, domain.FeatureFlagBudgetManagement); enabled {
// 		features = append(features, "budget-management")
// 	}
// 	if enabled, _ := s.featureFlags.IsEnabled(ctx, domain.FeatureFlagRecurringTrans); enabled {
// 		features = append(features, "recurring-transactions")
// 	}
// 	if enabled, _ := s.featureFlags.IsEnabled(ctx, domain.FeatureFlagApprovalWorkflow); enabled {
// 		features = append(features, "approval-workflow")
// 	}
// 	if enabled, _ := s.featureFlags.IsEnabled(ctx, domain.FeatureFlagBankReconciliation); enabled {
// 		features = append(features, "bank-reconciliation")
// 	}
//
// 	return features
// }
