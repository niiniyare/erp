// Package services provides financial services and utilities //for enterprise-grade accounting and financial management systems. // //It includes modules for: //- Accounts Receivable (AR) management: invoice processing, payment tracking, and aging reports //- Accounts Payable (AP) management: vendor payments, expense tracking, and payment scheduling //- Foreign exchange operations: real-time currency conversion, rate management, and gain/loss calculations //- Financial reporting: balance sheets, income statements, cash flow statements, and regulatory compliance //- General ledger maintenance: journal entries, chart of accounts, and trial balance //- Audit trail management: transaction logging and compliance documentation // //The package is designed to meet professional accounting standards (GAAP/IFRS) //and provides robust error handling, data validation, and security features //suitable for production financial systems. package finance

package service

import (
	"github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/core/finance/domain"
	"github.com/niiniyare/erp/internal/core/iam"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// Services aggregates all finance-related services
type Services struct {
	Account          AccountService
	Transaction      TransactionService
	TransactionEntry TransactionEntryService
}

// Dependencies contains the required dependencies to create finance services
type Dependencies struct {
	AccountRepo      domain.AccountsRepository
	AccountGroupRepo domain.AccountGroupRepository
	TransactionRepo  domain.TransactionRepository
	// TransactionEntryRepo domain.TransactionRepository // TODO: Create separate entry repository
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

	return &Services{
		Account:          accountService,
		Transaction:      transactionService,
		TransactionEntry: transactionEntryService,
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
