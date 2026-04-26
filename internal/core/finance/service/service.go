// Package services provides financial services and utilities //for enterprise-grade accounting and financial management systems. // //It includes modules for: //- Accounts Receivable (AR) management: invoice processing, payment tracking, and aging reports //- Accounts Payable (AP) management: vendor payments, expense tracking, and payment scheduling //- Foreign exchange operations: real-time currency conversion, rate management, and gain/loss calculations //- Financial reporting: balance sheets, income statements, cash flow statements, and regulatory compliance //- General ledger maintenance: journal entries, chart of accounts, and trial balance //- Audit trail management: transaction logging and compliance documentation // //The package is designed to meet professional accounting standards (GAAP/IFRS) //and provides robust error handling, data validation, and security features //suitable for production financial systems. package finance

package service

import (
	"awo.so/internal/core/featureflag"
	"awo.so/internal/core/finance/domain"
	"awo.so/internal/core/iam"
	"awo.so/internal/shared/errors"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// Services aggregates all finance-related services
type Services struct {
	Account          AccountService
	Transaction      TransactionService
	TransactionEntry TransactionEntryService
	Period           PeriodService
	ExchangeRate     ExchangeRateService
	Currency         CurrencyService
	CostCenter       CostCenterService
	Budget           BudgetService
	Tax              TaxService
	Reconciliation   ReconciliationService
}

// Dependencies contains the required dependencies to create finance services
type Dependencies struct {
	AccountRepo      domain.AccountsRepository
	AccountGroupRepo domain.AccountGroupRepository
	TransactionRepo  domain.TransactionRepository
	// TransactionEntryRepo domain.TransactionRepository // TODO: Create separate entry repository
	PeriodRepo       domain.PeriodRepository
	ExchangeRateRepo domain.ExchangeRateRepository
	CurrencyRepo     CurrencyRepository
	CostCenterRepo      domain.CostCenterRepository
	BudgetRepo          domain.BudgetRepository
	TaxRepo             domain.TaxRepository
	ReconciliationRepo  domain.ReconciliationRepository
	Tracing            tracing.Service
	Metrics            metrics.MetricsProvider
	IAMService         iam.Service
	FeatureFlagService featureflag.Service
}

// NewServices creates a new instance of finance services with all dependencies
func NewServices(deps Dependencies) *Services {
	// Create TransactionEntry service first as Transaction service depends on it
	transactionEntryService := NewTransactionEntryService(
		// deps.TransactionEntryRepo,
		nil,
		deps.AccountRepo,
		deps.Tracing,
		deps.Metrics,
	)

	// Create Transaction service with entry service dependency
	transactionService := NewTransactionService(
		deps.TransactionRepo,
		deps.AccountRepo,
		transactionEntryService,
		deps.Tracing,
		deps.Metrics,
	)

	// Create Account service with account group repository
	accountService := NewAccountService(
		deps.AccountRepo,
		deps.AccountGroupRepo,
		deps.Tracing,
		deps.Metrics,
		deps.IAMService,
		deps.FeatureFlagService,
	)

	periodService := NewPeriodService(deps.PeriodRepo, deps.Tracing, deps.Metrics)
	exchangeRateService := NewExchangeRateService(deps.ExchangeRateRepo, deps.Tracing, deps.Metrics)
	currencyService := NewCurrencyService(deps.CurrencyRepo, deps.Tracing, deps.Metrics)
	costCenterService := NewCostCenterService(deps.CostCenterRepo, deps.Tracing, deps.Metrics)
	budgetService := NewBudgetService(deps.BudgetRepo, deps.Tracing, deps.Metrics)
	taxService := NewTaxService(deps.TaxRepo, deps.Tracing, deps.Metrics)
	reconciliationService := NewReconciliationService(deps.ReconciliationRepo, deps.Tracing, deps.Metrics)

	return &Services{
		Account:          accountService,
		Transaction:      transactionService,
		TransactionEntry: transactionEntryService,
		Period:           periodService,
		ExchangeRate:     exchangeRateService,
		Currency:         currencyService,
		CostCenter:       costCenterService,
		Budget:           budgetService,
		Tax:              taxService,
		Reconciliation:   reconciliationService,
	}
}

// Validate ensures all required dependencies are provided
func (d Dependencies) Validate() error {
	if d.AccountRepo == nil {
		return errors.NewBusinessError("MISSING_DEPENDENCY", "Account repository is required")
	}

	if d.TransactionRepo == nil {
		return errors.NewBusinessError("MISSING_DEPENDENCY", "Transaction repository is required")
	}

	// if d.TransactionEntryRepo == nil {
	// 	return errors.NewBusinessError("MISSING_DEPENDENCY", "Transaction entry repository is required")
	// }

	if d.Tracing == nil {
		return errors.NewBusinessError("MISSING_DEPENDENCY", "Tracing service is required")
	}

	if d.Metrics == nil {
		return errors.NewBusinessError("MISSING_DEPENDENCY", "Metrics provider is required")
	}

	return nil
}
