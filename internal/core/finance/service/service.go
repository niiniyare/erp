package service

import (
	"github.com/niiniyare/erp/internal/core/finance/domain"
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
	AccountRepo     domain.AccountsRepository
	TransactionRepo domain.TransactionRepository
	// TransactionEntryRepo domain.TransactionRepository // TODO: Create separate entry repository
	Tracing tracing.TracingService
	Metrics metrics.MetricsProvider
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

	// Create Account service
	accountService := NewAccountService(
		deps.AccountRepo,
		deps.Tracing,
		deps.Metrics,
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
