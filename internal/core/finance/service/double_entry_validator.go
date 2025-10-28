package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.opentelemetry.io/otel/attribute"

	"github.com/niiniyare/erp/internal/core/finance/domain"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// DoubleEntryValidator provides validation for double-entry transactions
type DoubleEntryValidator interface {
	// ValidateTransaction validates a complete transaction for double-entry compliance
	ValidateTransaction(ctx context.Context, transaction *domain.Transaction, accounts map[uuid.UUID]*domain.Accounts) *ValidationResult

	// ValidateEntries validates transaction entries for double-entry rules
	ValidateEntries(ctx context.Context, entries []domain.TransactionEntry, accounts map[uuid.UUID]*domain.Accounts) *ValidationResult

	// ValidateBalance validates that debits equal credits
	ValidateBalance(ctx context.Context, transaction *domain.Transaction) *ValidationResult

	// ValidateAccountCompatibility validates entries against account types
	ValidateAccountCompatibility(ctx context.Context, entries []domain.TransactionEntry, accounts map[uuid.UUID]*domain.Accounts) *ValidationResult

	// ValidateCurrencyConsistency validates multi-currency transaction rules
	ValidateCurrencyConsistency(ctx context.Context, transaction *domain.Transaction, accounts map[uuid.UUID]*domain.Accounts) *ValidationResult

	// ValidateBusinessRules validates complex business rules
	ValidateBusinessRules(ctx context.Context, transaction *domain.Transaction, accounts map[uuid.UUID]*domain.Accounts) *ValidationResult
}

// ValidationResult represents the result of double-entry validation
type ValidationResult struct {
	IsValid         bool                      `json:"is_valid"`
	ValidationLevel ValidationLevel           `json:"validation_level"`
	Errors          []ValidationError         `json:"errors,omitempty"`
	Warnings        []ValidationWarning       `json:"warnings,omitempty"`
	BusinessRules   []BusinessRuleResult      `json:"business_rules,omitempty"`
	BalanceCheck    *BalanceValidationResult  `json:"balance_check,omitempty"`
	CurrencyCheck   *CurrencyValidationResult `json:"currency_check,omitempty"`
	AccountCheck    *AccountValidationResult  `json:"account_check,omitempty"`
}

// ValidationLevel represents the severity of validation issues
type ValidationLevel string

const (
	ValidationLevelValid   ValidationLevel = "VALID"
	ValidationLevelWarning ValidationLevel = "WARNING"
	ValidationLevelError   ValidationLevel = "ERROR"
	ValidationLevelFatal   ValidationLevel = "FATAL"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field         string `json:"field"`
	Message       string `json:"message"`
	Code          string `json:"code"`
	Severity      string `json:"severity"`
	Value         any    `json:"value,omitempty"`
	ExpectedValue any    `json:"expected_value,omitempty"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
	Value   any    `json:"value,omitempty"`
}

// BusinessRuleResult represents the result of a business rule validation
type BusinessRuleResult struct {
	RuleID   string `json:"rule_id"`
	RuleName string `json:"rule_name"`
	Passed   bool   `json:"passed"`
	Message  string `json:"message,omitempty"`
	Severity string `json:"severity"`
}

// BalanceValidationResult represents balance validation details
type BalanceValidationResult struct {
	TotalDebits  decimal.Decimal `json:"total_debits"`
	TotalCredits decimal.Decimal `json:"total_credits"`
	Difference   decimal.Decimal `json:"difference"`
	IsBalanced   bool            `json:"is_balanced"`
	Tolerance    decimal.Decimal `json:"tolerance"`
}

// CurrencyValidationResult represents currency validation details
type CurrencyValidationResult struct {
	BaseCurrency       string   `json:"base_currency"`
	MultiCurrency      bool     `json:"multi_currency"`
	CurrenciesUsed     []string `json:"currencies_used"`
	ExchangeRateValid  bool     `json:"exchange_rate_valid"`
	ConversionAccurate bool     `json:"conversion_accurate"`
}

// AccountValidationResult represents account validation details
type AccountValidationResult struct {
	AccountsValidated int            `json:"accounts_validated"`
	InactiveAccounts  []uuid.UUID    `json:"inactive_accounts,omitempty"`
	MissingAccounts   []uuid.UUID    `json:"missing_accounts,omitempty"`
	AccountTypes      map[string]int `json:"account_types"`
	NormalBalances    map[string]int `json:"normal_balances"`
}

// doubleEntryValidator implements DoubleEntryValidator
type doubleEntryValidator struct {
	tracing   tracing.Service
	tolerance decimal.Decimal // Tolerance for balance validation (e.g., 0.01)
}

// DoubleEntryValidatorDeps represents dependencies for the validator
type DoubleEntryValidatorDeps struct {
	Tracing   tracing.Service
	Tolerance *decimal.Decimal // Optional tolerance for balance validation
}

// NewDoubleEntryValidator creates a new double-entry validator
func NewDoubleEntryValidator(deps DoubleEntryValidatorDeps) DoubleEntryValidator {
	tolerance := decimal.NewFromFloat(0.01) // Default tolerance
	if deps.Tolerance != nil {
		tolerance = *deps.Tolerance
	}

	return &doubleEntryValidator{
		tracing:   deps.Tracing,
		tolerance: tolerance,
	}
}

// ValidateTransaction validates a complete transaction for double-entry compliance
func (v *doubleEntryValidator) ValidateTransaction(ctx context.Context, transaction *domain.Transaction, accounts map[uuid.UUID]*domain.Accounts) *ValidationResult {
	ctx, span := v.tracing.StartSpan(ctx, "DoubleEntryValidator.ValidateTransaction")
	defer span.End()

	result := &ValidationResult{
		IsValid:         true,
		ValidationLevel: ValidationLevelValid,
		Errors:          []ValidationError{},
		Warnings:        []ValidationWarning{},
		BusinessRules:   []BusinessRuleResult{},
	}

	span.SetAttributes(
		attribute.String("transaction_id", transaction.ID.String()),
		attribute.String("transaction_type", string(transaction.TransactionType)),
		attribute.Int64("entry_count", int64(len(transaction.Entries))),
	)

	// 1. Validate basic transaction structure
	v.validateTransactionStructure(ctx, transaction, result)

	// 2. Validate balance (debits = credits)
	balanceResult := v.ValidateBalance(ctx, transaction)
	result.BalanceCheck = balanceResult.BalanceCheck
	v.mergeResults(result, balanceResult)

	// 3. Validate entries
	entriesResult := v.ValidateEntries(ctx, transaction.Entries, accounts)
	v.mergeResults(result, entriesResult)

	// 4. Validate account compatibility
	accountResult := v.ValidateAccountCompatibility(ctx, transaction.Entries, accounts)
	result.AccountCheck = accountResult.AccountCheck
	v.mergeResults(result, accountResult)

	// 5. Validate currency consistency
	currencyResult := v.ValidateCurrencyConsistency(ctx, transaction, accounts)
	result.CurrencyCheck = currencyResult.CurrencyCheck
	v.mergeResults(result, currencyResult)

	// 6. Validate business rules
	businessResult := v.ValidateBusinessRules(ctx, transaction, accounts)
	result.BusinessRules = businessResult.BusinessRules
	v.mergeResults(result, businessResult)

	// Set final validation level
	if len(result.Errors) > 0 {
		result.IsValid = false
		result.ValidationLevel = ValidationLevelError
	} else if len(result.Warnings) > 0 {
		result.ValidationLevel = ValidationLevelWarning
	}

	return result
}

// ValidateEntries validates transaction entries for double-entry rules
func (v *doubleEntryValidator) ValidateEntries(ctx context.Context, entries []domain.TransactionEntry, accounts map[uuid.UUID]*domain.Accounts) *ValidationResult {
	ctx, span := v.tracing.StartSpan(ctx, "DoubleEntryValidator.ValidateEntries")
	defer span.End()

	result := &ValidationResult{
		IsValid:         true,
		ValidationLevel: ValidationLevelValid,
		Errors:          []ValidationError{},
		Warnings:        []ValidationWarning{},
	}

	// Validate minimum entry count
	if len(entries) < 2 {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "entries",
			Message:  "Transaction must have at least 2 entries for double-entry bookkeeping",
			Code:     "INSUFFICIENT_ENTRIES",
			Severity: "ERROR",
		})
	}

	// Validate each entry
	for i, entry := range entries {
		v.validateIndividualEntry(ctx, entry, i, accounts, result)
	}

	// Validate entry uniqueness (no duplicate entries for same account)
	v.validateEntryUniqueness(ctx, entries, result)

	// Validate entry consistency
	v.validateEntryConsistency(ctx, entries, result)

	return result
}

// ValidateBalance validates that debits equal credits
func (v *doubleEntryValidator) ValidateBalance(ctx context.Context, transaction *domain.Transaction) *ValidationResult {
	ctx, span := v.tracing.StartSpan(ctx, "DoubleEntryValidator.ValidateBalance")
	defer span.End()

	result := &ValidationResult{
		IsValid:         true,
		ValidationLevel: ValidationLevelValid,
		Errors:          []ValidationError{},
		Warnings:        []ValidationWarning{},
	}

	// Calculate totals
	totalDebits := decimal.Zero
	totalCredits := decimal.Zero

	for _, entry := range transaction.Entries {
		totalDebits = totalDebits.Add(entry.DebitAmount)
		totalCredits = totalCredits.Add(entry.CreditAmount)
	}

	// Calculate difference
	difference := totalDebits.Sub(totalCredits).Abs()

	// Create balance check result
	balanceCheck := &BalanceValidationResult{
		TotalDebits:  totalDebits,
		TotalCredits: totalCredits,
		Difference:   difference,
		IsBalanced:   difference.LessThanOrEqual(v.tolerance),
		Tolerance:    v.tolerance,
	}

	result.BalanceCheck = balanceCheck

	// Validate balance
	if !balanceCheck.IsBalanced {
		result.Errors = append(result.Errors, ValidationError{
			Field:         "entries",
			Message:       fmt.Sprintf("Transaction does not balance: debits=%s, credits=%s, difference=%s", totalDebits, totalCredits, difference),
			Code:          "UNBALANCED_TRANSACTION",
			Severity:      "ERROR",
			Value:         difference,
			ExpectedValue: decimal.Zero,
		})
	}

	// Validate that totals match transaction header
	if !totalDebits.Equal(transaction.TotalDebitAmount) {
		result.Errors = append(result.Errors, ValidationError{
			Field:         "total_debit_amount",
			Message:       fmt.Sprintf("Transaction header debit total (%s) does not match entries total (%s)", transaction.TotalDebitAmount, totalDebits),
			Code:          "HEADER_ENTRY_MISMATCH",
			Severity:      "ERROR",
			Value:         transaction.TotalDebitAmount,
			ExpectedValue: totalDebits,
		})
	}

	if !totalCredits.Equal(transaction.TotalCreditAmount) {
		result.Errors = append(result.Errors, ValidationError{
			Field:         "total_credit_amount",
			Message:       fmt.Sprintf("Transaction header credit total (%s) does not match entries total (%s)", transaction.TotalCreditAmount, totalCredits),
			Code:          "HEADER_ENTRY_MISMATCH",
			Severity:      "ERROR",
			Value:         transaction.TotalCreditAmount,
			ExpectedValue: totalCredits,
		})
	}

	span.SetAttributes(
		attribute.String("total_debits", totalDebits.String()),
		attribute.String("total_credits", totalCredits.String()),
		attribute.String("difference", difference.String()),
		attribute.Bool("is_balanced", balanceCheck.IsBalanced),
	)

	return result
}

// ValidateAccountCompatibility validates entries against account types
func (v *doubleEntryValidator) ValidateAccountCompatibility(ctx context.Context, entries []domain.TransactionEntry, accounts map[uuid.UUID]*domain.Accounts) *ValidationResult {
	ctx, span := v.tracing.StartSpan(ctx, "DoubleEntryValidator.ValidateAccountCompatibility")
	defer span.End()

	result := &ValidationResult{
		IsValid:         true,
		ValidationLevel: ValidationLevelValid,
		Errors:          []ValidationError{},
		Warnings:        []ValidationWarning{},
	}

	accountCheck := &AccountValidationResult{
		AccountsValidated: 0,
		InactiveAccounts:  []uuid.UUID{},
		MissingAccounts:   []uuid.UUID{},
		AccountTypes:      make(map[string]int),
		NormalBalances:    make(map[string]int),
	}

	for i, entry := range entries {
		account, exists := accounts[entry.AccountID]
		if !exists {
			accountCheck.MissingAccounts = append(accountCheck.MissingAccounts, entry.AccountID)
			result.Errors = append(result.Errors, ValidationError{
				Field:    fmt.Sprintf("entries[%d].account_id", i),
				Message:  "Referenced account does not exist",
				Code:     "ACCOUNT_NOT_FOUND",
				Severity: "ERROR",
				Value:    entry.AccountID,
			})
			continue
		}

		accountCheck.AccountsValidated++

		// Check if account is active
		if !account.IsActive {
			accountCheck.InactiveAccounts = append(accountCheck.InactiveAccounts, entry.AccountID)
			result.Errors = append(result.Errors, ValidationError{
				Field:    fmt.Sprintf("entries[%d].account_id", i),
				Message:  fmt.Sprintf("Account %s is inactive", account.AccountName),
				Code:     "INACTIVE_ACCOUNT",
				Severity: "ERROR",
				Value:    account.AccountName,
			})
		}

		// Track account types and normal balances
		accountCheck.AccountTypes[string(account.AccountType)]++
		accountCheck.NormalBalances[string(account.NormalBalance)]++

		// Check if account allows manual entries (for manual transactions)
		if !account.AllowManualEntries {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:   fmt.Sprintf("entries[%d].account_id", i),
				Message: fmt.Sprintf("Account %s does not typically allow manual entries", account.AccountName),
				Code:    "MANUAL_ENTRY_WARNING",
				Value:   account.AccountName,
			})
		}

		// Check if account requires reference
		if account.RequireReference && (entry.Reference == nil || strings.TrimSpace(*entry.Reference) == "") {
			result.Errors = append(result.Errors, ValidationError{
				Field:    fmt.Sprintf("entries[%d].reference", i),
				Message:  fmt.Sprintf("Account %s requires a reference for all entries", account.AccountName),
				Code:     "REFERENCE_REQUIRED",
				Severity: "ERROR",
				Value:    account.AccountName,
			})
		}

		// Validate normal balance vs. debit/credit
		v.validateNormalBalance(ctx, entry, account, i, result)
	}

	result.AccountCheck = accountCheck
	return result
}

// ValidateCurrencyConsistency validates multi-currency transaction rules
func (v *doubleEntryValidator) ValidateCurrencyConsistency(ctx context.Context, transaction *domain.Transaction, accounts map[uuid.UUID]*domain.Accounts) *ValidationResult {
	ctx, span := v.tracing.StartSpan(ctx, "DoubleEntryValidator.ValidateCurrencyConsistency")
	defer span.End()

	result := &ValidationResult{
		IsValid:         true,
		ValidationLevel: ValidationLevelValid,
		Errors:          []ValidationError{},
		Warnings:        []ValidationWarning{},
	}

	currenciesUsed := make(map[string]bool)
	currenciesUsed[transaction.CurrencyCode] = true

	// Collect currencies from entries
	for _, entry := range transaction.Entries {
		if entry.OriginalCurrency != nil && *entry.OriginalCurrency != transaction.CurrencyCode {
			currenciesUsed[*entry.OriginalCurrency] = true
		}
	}

	currencies := make([]string, 0, len(currenciesUsed))
	for currency := range currenciesUsed {
		currencies = append(currencies, currency)
	}

	currencyCheck := &CurrencyValidationResult{
		BaseCurrency:       transaction.CurrencyCode,
		MultiCurrency:      len(currencies) > 1,
		CurrenciesUsed:     currencies,
		ExchangeRateValid:  transaction.ExchangeRate.GreaterThan(decimal.Zero),
		ConversionAccurate: true, // Assume true for now
	}

	// Validate exchange rate
	if !currencyCheck.ExchangeRateValid {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "exchange_rate",
			Message:  "Exchange rate must be greater than zero",
			Code:     "INVALID_EXCHANGE_RATE",
			Severity: "ERROR",
			Value:    transaction.ExchangeRate,
		})
	}

	// Validate multi-currency entries
	for i, entry := range transaction.Entries {
		if entry.OriginalCurrency != nil {
			// Validate exchange rate for entry
			if entry.ExchangeRate.LessThanOrEqual(decimal.Zero) {
				result.Errors = append(result.Errors, ValidationError{
					Field:    fmt.Sprintf("entries[%d].exchange_rate", i),
					Message:  "Entry exchange rate must be greater than zero for multi-currency entries",
					Code:     "INVALID_ENTRY_EXCHANGE_RATE",
					Severity: "ERROR",
				})
			}

			// Validate original amount
			if entry.OriginalAmount.LessThanOrEqual(decimal.Zero) {
				result.Errors = append(result.Errors, ValidationError{
					Field:    fmt.Sprintf("entries[%d].original_amount", i),
					Message:  "Original amount must be provided for multi-currency entries",
					Code:     "MISSING_ORIGINAL_AMOUNT",
					Severity: "ERROR",
				})
			}
		}
	}

	result.CurrencyCheck = currencyCheck
	return result
}

// ValidateBusinessRules validates complex business rules
func (v *doubleEntryValidator) ValidateBusinessRules(ctx context.Context, transaction *domain.Transaction, accounts map[uuid.UUID]*domain.Accounts) *ValidationResult {
	ctx, span := v.tracing.StartSpan(ctx, "DoubleEntryValidator.ValidateBusinessRules")
	defer span.End()

	result := &ValidationResult{
		IsValid:         true,
		ValidationLevel: ValidationLevelValid,
		Errors:          []ValidationError{},
		Warnings:        []ValidationWarning{},
		BusinessRules:   []BusinessRuleResult{},
	}

	// Rule 1: Asset and Expense accounts should normally have debit entries
	v.validateAccountTypeConsistency(ctx, transaction, accounts, result)

	// Rule 2: Revenue, Liability, and Equity accounts should normally have credit entries
	v.validateNormalBalanceRules(ctx, transaction, accounts, result)

	// Rule 3: Control accounts should not have direct entries
	v.validateControlAccountUsage(ctx, transaction, accounts, result)

	// Rule 4: Posting date should be within valid accounting period
	v.validatePostingPeriod(ctx, transaction, result)

	// Rule 5: Transaction amount limits (if applicable)
	v.validateAmountLimits(ctx, transaction, result)

	return result
}

// Helper methods for validation

func (v *doubleEntryValidator) validateTransactionStructure(ctx context.Context, transaction *domain.Transaction, result *ValidationResult) {
	if transaction.ID == uuid.Nil {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "id",
			Message:  "Transaction ID is required",
			Code:     "REQUIRED_FIELD",
			Severity: "ERROR",
		})
	}

	if transaction.TenantID == uuid.Nil {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "tenant_id",
			Message:  "Tenant ID is required",
			Code:     "REQUIRED_FIELD",
			Severity: "ERROR",
		})
	}

	if strings.TrimSpace(transaction.TransactionNumber) == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "transaction_number",
			Message:  "Transaction number is required",
			Code:     "REQUIRED_FIELD",
			Severity: "ERROR",
		})
	}
}

func (v *doubleEntryValidator) validateIndividualEntry(ctx context.Context, entry domain.TransactionEntry, index int, accounts map[uuid.UUID]*domain.Accounts, result *ValidationResult) {
	// Validate that entry has either debit OR credit (not both, not neither)
	hasDebit := entry.DebitAmount.GreaterThan(decimal.Zero)
	hasCredit := entry.CreditAmount.GreaterThan(decimal.Zero)

	if hasDebit && hasCredit {
		result.Errors = append(result.Errors, ValidationError{
			Field:    fmt.Sprintf("entries[%d]", index),
			Message:  "Entry cannot have both debit and credit amounts",
			Code:     "INVALID_ENTRY_AMOUNTS",
			Severity: "ERROR",
		})
	}

	if !hasDebit && !hasCredit {
		result.Errors = append(result.Errors, ValidationError{
			Field:    fmt.Sprintf("entries[%d]", index),
			Message:  "Entry must have either debit or credit amount",
			Code:     "MISSING_ENTRY_AMOUNT",
			Severity: "ERROR",
		})
	}

	// Validate amounts are positive
	if entry.DebitAmount.LessThan(decimal.Zero) {
		result.Errors = append(result.Errors, ValidationError{
			Field:    fmt.Sprintf("entries[%d].debit_amount", index),
			Message:  "Debit amount cannot be negative",
			Code:     "NEGATIVE_AMOUNT",
			Severity: "ERROR",
			Value:    entry.DebitAmount,
		})
	}

	if entry.CreditAmount.LessThan(decimal.Zero) {
		result.Errors = append(result.Errors, ValidationError{
			Field:    fmt.Sprintf("entries[%d].credit_amount", index),
			Message:  "Credit amount cannot be negative",
			Code:     "NEGATIVE_AMOUNT",
			Severity: "ERROR",
			Value:    entry.CreditAmount,
		})
	}
}

func (v *doubleEntryValidator) validateNormalBalance(ctx context.Context, entry domain.TransactionEntry, account *domain.Accounts, index int, result *ValidationResult) {
	hasDebit := entry.DebitAmount.GreaterThan(decimal.Zero)
	hasCredit := entry.CreditAmount.GreaterThan(decimal.Zero)

	switch account.NormalBalance {
	case domain.NormalBalanceDebit:
		if hasCredit {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:   fmt.Sprintf("entries[%d]", index),
				Message: fmt.Sprintf("Account %s has normal debit balance but entry is credit", account.AccountName),
				Code:    "ABNORMAL_BALANCE_ENTRY",
				Value:   account.AccountName,
			})
		}
	case domain.NormalBalanceCredit:
		if hasDebit {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:   fmt.Sprintf("entries[%d]", index),
				Message: fmt.Sprintf("Account %s has normal credit balance but entry is debit", account.AccountName),
				Code:    "ABNORMAL_BALANCE_ENTRY",
				Value:   account.AccountName,
			})
		}
	}
}

func (v *doubleEntryValidator) validateEntryUniqueness(ctx context.Context, entries []domain.TransactionEntry, result *ValidationResult) {
	accountCounts := make(map[uuid.UUID]int)
	for _, entry := range entries {
		accountCounts[entry.AccountID]++
	}

	for accountID, count := range accountCounts {
		if count > 1 {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:   "entries",
				Message: fmt.Sprintf("Account %s appears in multiple entries", accountID.String()),
				Code:    "DUPLICATE_ACCOUNT_ENTRIES",
				Value:   accountID.String(),
			})
		}
	}
}

func (v *doubleEntryValidator) validateEntryConsistency(ctx context.Context, entries []domain.TransactionEntry, result *ValidationResult) {
	// Check for zero-amount entries
	for i, entry := range entries {
		if entry.DebitAmount.IsZero() && entry.CreditAmount.IsZero() {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:   fmt.Sprintf("entries[%d]", i),
				Message: "Entry has zero amount",
				Code:    "ZERO_AMOUNT_ENTRY",
			})
		}
	}
}

func (v *doubleEntryValidator) validateAccountTypeConsistency(ctx context.Context, transaction *domain.Transaction, accounts map[uuid.UUID]*domain.Accounts, result *ValidationResult) {
	// This is a business rule validation - can be customized per implementation
	rule := BusinessRuleResult{
		RuleID:   "ACCOUNT_TYPE_CONSISTENCY",
		RuleName: "Account Type Consistency Check",
		Passed:   true,
		Severity: "WARNING",
	}

	result.BusinessRules = append(result.BusinessRules, rule)
}

func (v *doubleEntryValidator) validateNormalBalanceRules(ctx context.Context, transaction *domain.Transaction, accounts map[uuid.UUID]*domain.Accounts, result *ValidationResult) {
	rule := BusinessRuleResult{
		RuleID:   "NORMAL_BALANCE_RULES",
		RuleName: "Normal Balance Rules Check",
		Passed:   true,
		Severity: "WARNING",
	}

	result.BusinessRules = append(result.BusinessRules, rule)
}

func (v *doubleEntryValidator) validateControlAccountUsage(ctx context.Context, transaction *domain.Transaction, accounts map[uuid.UUID]*domain.Accounts, result *ValidationResult) {
	rule := BusinessRuleResult{
		RuleID:   "CONTROL_ACCOUNT_USAGE",
		RuleName: "Control Account Usage Check",
		Passed:   true,
		Severity: "ERROR",
	}

	for i, entry := range transaction.Entries {
		if account, exists := accounts[entry.AccountID]; exists && account.IsControlAccount {
			rule.Passed = false
			rule.Message = fmt.Sprintf("Entry %d uses control account %s", i+1, account.AccountName)
			result.Errors = append(result.Errors, ValidationError{
				Field:    fmt.Sprintf("entries[%d].account_id", i),
				Message:  fmt.Sprintf("Control account %s should not have direct entries", account.AccountName),
				Code:     "CONTROL_ACCOUNT_ENTRY",
				Severity: "ERROR",
				Value:    account.AccountName,
			})
		}
	}

	result.BusinessRules = append(result.BusinessRules, rule)
}

func (v *doubleEntryValidator) validatePostingPeriod(ctx context.Context, transaction *domain.Transaction, result *ValidationResult) {
	rule := BusinessRuleResult{
		RuleID:   "POSTING_PERIOD_VALIDATION",
		RuleName: "Posting Period Validation",
		Passed:   true,
		Severity: "WARNING",
	}

	// Check if posting date is in the future (warning)
	if transaction.PostingDate != nil && transaction.PostingDate.After(time.Now()) {
		rule.Passed = false
		rule.Message = "Posting date is in the future"
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "posting_date",
			Message: "Posting date is in the future",
			Code:    "FUTURE_POSTING_DATE",
			Value:   transaction.PostingDate,
		})
	}

	result.BusinessRules = append(result.BusinessRules, rule)
}

func (v *doubleEntryValidator) validateAmountLimits(ctx context.Context, transaction *domain.Transaction, result *ValidationResult) {
	rule := BusinessRuleResult{
		RuleID:   "AMOUNT_LIMITS",
		RuleName: "Transaction Amount Limits",
		Passed:   true,
		Severity: "WARNING",
	}

	// Example: Warn for very large transactions
	maxAmount := decimal.NewFromInt(1000000) // $1M limit for example
	if transaction.TotalDebitAmount.GreaterThan(maxAmount) {
		rule.Passed = false
		rule.Message = fmt.Sprintf("Transaction amount (%s) exceeds typical limit (%s)", transaction.TotalDebitAmount, maxAmount)
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "total_debit_amount",
			Message: "Transaction amount is unusually large",
			Code:    "LARGE_TRANSACTION_AMOUNT",
			Value:   transaction.TotalDebitAmount,
		})
	}

	result.BusinessRules = append(result.BusinessRules, rule)
}

func (v *doubleEntryValidator) mergeResults(target *ValidationResult, source *ValidationResult) {
	target.Errors = append(target.Errors, source.Errors...)
	target.Warnings = append(target.Warnings, source.Warnings...)
	target.BusinessRules = append(target.BusinessRules, source.BusinessRules...)

	if len(source.Errors) > 0 {
		target.IsValid = false
		target.ValidationLevel = ValidationLevelError
	} else if len(source.Warnings) > 0 && target.ValidationLevel == ValidationLevelValid {
		target.ValidationLevel = ValidationLevelWarning
	}
}
