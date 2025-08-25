package domain

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ValidationError represents a single validation error
// ValidationResult holds the result of a validation operation
type ValidationResult struct {
	IsValid bool              `json:"is_valid"`
	Errors  []ValidationError `json:"errors,omitempty"`
}

// HasErrors returns true if the validation result contains errors
func (vr *ValidationResult) HasErrors() bool {
	return len(vr.Errors) > 0
}

// AddError adds a validation error to the result
func (vr *ValidationResult) AddError(field, message, code string) {
	vr.IsValid = false
	vr.Errors = append(vr.Errors, ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
	})
}

// Merge combines multiple validation results
func (vr *ValidationResult) Merge(other *ValidationResult) {
	if other.HasErrors() {
		vr.IsValid = false
		vr.Errors = append(vr.Errors, other.Errors...)
	}
}

// GetErrorsByField returns errors for a specific field
func (vr *ValidationResult) GetErrorsByField(fieldName string) []ValidationError {
	var errors []ValidationError
	for _, err := range vr.Errors {
		if err.Field == fieldName {
			errors = append(errors, err)
		}
	}
	return errors
}

// Common validation patterns and constants
const (
	// Regular expressions for validation
	AccountCodePattern       = `^[A-Za-z0-9\-\.]+$`
	CurrencyCodePattern      = `^[A-Z]{3}$`
	TransactionNumberPattern = `^[A-Za-z0-9\-]+$`
	TaxCodePattern           = `^[A-Za-z0-9\-]+$`
	EmailPattern             = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	PhonePattern             = `^[\+]?[1-9][\d]{0,15}$`

	// Maximum lengths
	MaxAccountCodeLength       = 20
	MaxAccountNameLength       = 200
	MaxDescriptionLength       = 1000
	MaxTransactionNumberLength = 50
	MaxReferenceLength         = 100
	MaxCostCenterLength        = 20
	MaxDepartmentLength        = 20
	MaxTaxCodeLength           = 20
	MaxEmailLength             = 255
	MaxPhoneLength             = 20
	MaxAddressLength           = 500
	MaxNotesLength             = 2000

	// Minimum values
	MinExchangeRate = 0.000001
	MaxExchangeRate = 1000000.0
	MaxTaxRate      = 100.0
	MinTaxRate      = 0.0

	// Date constraints
	MaxFutureDays = 365
)

// ValidatorFunc defines a validation function signature
type ValidatorFunc func() []ValidationError

// FieldValidator provides validation for individual fields
type FieldValidator struct{}

// ValidateUUID validates that a UUID is not nil/empty
func (v *FieldValidator) ValidateUUID(value uuid.UUID, fieldName string, required bool) []ValidationError {
	var errors []ValidationError

	if required && value == uuid.Nil {
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s is required", fieldName),
			Code:    "REQUIRED_FIELD",
		})
	}

	return errors
}

// ValidateString validates string fields with length and pattern constraints
func (v *FieldValidator) ValidateString(value, fieldName string, required bool, maxLength int, pattern string) []ValidationError {
	var errors []ValidationError

	trimmedValue := strings.TrimSpace(value)

	if required && trimmedValue == "" {
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s is required", fieldName),
			Code:    "REQUIRED_FIELD",
		})
		return errors // Don't continue validation if required field is empty
	}

	if len(value) > maxLength {
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s must be %d characters or less", fieldName, maxLength),
			Code:    "MAX_LENGTH_EXCEEDED",
		})
	}

	if pattern != "" && trimmedValue != "" {
		matched, err := regexp.MatchString(pattern, trimmedValue)
		if err != nil || !matched {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%s has invalid format", fieldName),
				Code:    "INVALID_FORMAT",
			})
		}
	}

	return errors
}

// ValidateOptionalString validates optional string fields
func (v *FieldValidator) ValidateOptionalString(value *string, fieldName string, maxLength int, pattern string) []ValidationError {
	if value == nil {
		return nil
	}
	return v.ValidateString(*value, fieldName, false, maxLength, pattern)
}

// ValidateEmail validates email format
func (v *FieldValidator) ValidateEmail(email, fieldName string, required bool) []ValidationError {
	return v.ValidateString(email, fieldName, required, MaxEmailLength, EmailPattern)
}

// ValidatePhone validates phone number format
func (v *FieldValidator) ValidatePhone(phone, fieldName string, required bool) []ValidationError {
	return v.ValidateString(phone, fieldName, required, MaxPhoneLength, PhonePattern)
}

// ValidateDecimal validates decimal fields with range constraints
func (v *FieldValidator) ValidateDecimal(value decimal.Decimal, fieldName string, minValue, maxValue *decimal.Decimal) []ValidationError {
	var errors []ValidationError

	if minValue != nil && value.LessThan(*minValue) {
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s must be greater than or equal to %s", fieldName, minValue.String()),
			Code:    "VALUE_OUT_OF_RANGE",
		})
	}

	if maxValue != nil && value.GreaterThan(*maxValue) {
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s must be less than or equal to %s", fieldName, maxValue.String()),
			Code:    "VALUE_OUT_OF_RANGE",
		})
	}

	return errors
}

// ValidatePositiveDecimal validates that a decimal is positive
func (v *FieldValidator) ValidatePositiveDecimal(value decimal.Decimal, fieldName string) []ValidationError {
	var errors []ValidationError

	if value.LessThanOrEqual(decimal.Zero) {
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s must be greater than zero", fieldName),
			Code:    "VALUE_MUST_BE_POSITIVE",
		})
	}

	return errors
}

// ValidateNonNegativeDecimal validates that a decimal is non-negative
func (v *FieldValidator) ValidateNonNegativeDecimal(value decimal.Decimal, fieldName string) []ValidationError {
	var errors []ValidationError

	if value.LessThan(decimal.Zero) {
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s must be greater than or equal to zero", fieldName),
			Code:    "VALUE_MUST_BE_NON_NEGATIVE",
		})
	}

	return errors
}

// ValidateDate validates date fields with range constraints
func (v *FieldValidator) ValidateDate(value time.Time, fieldName string, minDate, maxDate *time.Time) []ValidationError {
	var errors []ValidationError

	if value.IsZero() {
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s is required", fieldName),
			Code:    "REQUIRED_FIELD",
		})
		return errors
	}

	if minDate != nil && value.Before(*minDate) {
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s cannot be before %s", fieldName, minDate.Format("2006-01-02")),
			Code:    "DATE_OUT_OF_RANGE",
		})
	}

	if maxDate != nil && value.After(*maxDate) {
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s cannot be after %s", fieldName, maxDate.Format("2006-01-02")),
			Code:    "DATE_OUT_OF_RANGE",
		})
	}

	return errors
}

// ValidateEnum validates that a value is within allowed enum values
func (v *FieldValidator) ValidateEnum(value string, fieldName string, allowedValues []string, required bool) []ValidationError {
	var errors []ValidationError

	if required && strings.TrimSpace(value) == "" {
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s is required", fieldName),
			Code:    "REQUIRED_FIELD",
		})
		return errors
	}

	if value != "" {
		if !slices.Contains(allowedValues, value) {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%s must be one of: %s", fieldName, strings.Join(allowedValues, ", ")),
				Code:    "INVALID_ENUM_VALUE",
			})
		}
	}

	return errors

}

// BusinessRuleValidator provides business-specific validation
type BusinessRuleValidator struct{}

func NewBusinessRuleValidator() *BusinessRuleValidator {
	return &BusinessRuleValidator{}

}

// ValidateAccountCode validates account code format and uniqueness constraints
func (v *BusinessRuleValidator) ValidateAccountCode(code string, tenantID uuid.UUID, entityID *uuid.UUID) []ValidationError {
	var errors []ValidationError

	fieldValidator := &FieldValidator{}
	errors = append(errors, fieldValidator.ValidateString(code, "account_code", true, MaxAccountCodeLength, AccountCodePattern)...)

	// Business rule: Account codes should not start with special characters
	if len(code) > 0 && (code[0] == '-' || code[0] == '.') {
		errors = append(errors, ValidationError{
			Field:   "account_code",
			Message: "Account code cannot start with '-' or '.'",
			Code:    "INVALID_ACCOUNT_CODE_FORMAT",
		})
	}

	// TODO: Check uniqueness in repository
	// if accountCodeExists(code, tenantID, entityID) {
	//     errors = append(errors, ValidationError{
	//         Field:   "account_code",
	//         Message: "Account code already exists",
	//         Code:    "DUPLICATE_ACCOUNT_CODE",
	//     })
	// }

	return errors
}

// ValidateTransactionNumber validates transaction number format and uniqueness
func (v *BusinessRuleValidator) ValidateTransactionNumber(number string, tenantID uuid.UUID, entityID *uuid.UUID) []ValidationError {
	var errors []ValidationError

	fieldValidator := &FieldValidator{}
	errors = append(errors, fieldValidator.ValidateString(number, "transaction_number", true, MaxTransactionNumberLength, TransactionNumberPattern)...)

	// TODO: Check uniqueness in repository
	// if transactionNumberExists(number, tenantID, entityID) {
	//     errors = append(errors, ValidationError{
	//         Field:   "transaction_number",
	//         Message: "Transaction number already exists",
	//         Code:    "DUPLICATE_TRANSACTION_NUMBER",
	//     })
	// }

	return errors
}

// ValidateExchangeRate validates exchange rate values
func (v *BusinessRuleValidator) ValidateExchangeRate(rate decimal.Decimal, fromCurrency, toCurrency string) []ValidationError {
	var errors []ValidationError

	minRate := decimal.NewFromFloat(MinExchangeRate)
	maxRate := decimal.NewFromFloat(MaxExchangeRate)

	fieldValidator := &FieldValidator{}
	errors = append(errors, fieldValidator.ValidateDecimal(rate, "exchange_rate", &minRate, &maxRate)...)

	// Business rule: Same currency should have rate of 1.0
	if fromCurrency == toCurrency && !rate.Equal(decimal.NewFromInt(1)) {
		errors = append(errors, ValidationError{
			Field:   "exchange_rate",
			Message: "Exchange rate must be 1.0 for same currency conversion",
			Code:    "INVALID_SAME_CURRENCY_RATE",
		})
	}

	return errors
}

// ValidateTaxRate validates tax rate values
func (v *BusinessRuleValidator) ValidateTaxRate(rate decimal.Decimal, taxCode *string) []ValidationError {
	var errors []ValidationError

	minRate := decimal.NewFromFloat(MinTaxRate)
	maxRate := decimal.NewFromFloat(MaxTaxRate)

	fieldValidator := &FieldValidator{}
	errors = append(errors, fieldValidator.ValidateDecimal(rate, "tax_rate", &minRate, &maxRate)...)

	// Business rule: Tax rate should be zero if no tax code provided
	if taxCode == nil && !rate.IsZero() {
		errors = append(errors, ValidationError{
			Field:   "tax_rate",
			Message: "Tax rate should be zero when no tax code is provided",
			Code:    "INCONSISTENT_TAX_DATA",
		})
	}

	// Business rule: Tax code should be provided if tax rate is not zero
	if !rate.IsZero() && (taxCode == nil || strings.TrimSpace(*taxCode) == "") {
		errors = append(errors, ValidationError{
			Field:   "tax_code",
			Message: "Tax code is required when tax rate is not zero",
			Code:    "MISSING_TAX_CODE",
		})
	}

	return errors
}

// ValidateTransactionBalance validates that transaction debits equal credits
func (v *BusinessRuleValidator) ValidateTransactionBalance(entries []TransactionEntry) []ValidationError {
	var errors []ValidationError

	if len(entries) == 0 {
		errors = append(errors, ValidationError{
			Field:   "entries",
			Message: "Transaction must have at least one entry",
			Code:    "NO_TRANSACTION_ENTRIES",
		})
		return errors
	}

	totalDebits := decimal.Zero
	totalCredits := decimal.Zero

	for i, entry := range entries {
		totalDebits = totalDebits.Add(entry.DebitAmount)
		totalCredits = totalCredits.Add(entry.CreditAmount)

		// Validate that entry has either debit or credit, but not both
		hasDebit := !entry.DebitAmount.IsZero()
		hasCredit := !entry.CreditAmount.IsZero()

		if hasDebit && hasCredit {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("entries[%d]", i),
				Message: "Entry cannot have both debit and credit amounts",
				Code:    "ENTRY_BOTH_DEBIT_CREDIT",
			})
		}

		if !hasDebit && !hasCredit {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("entries[%d]", i),
				Message: "Entry must have either debit or credit amount",
				Code:    "ENTRY_NO_AMOUNT",
			})
		}
	}

	if !totalDebits.Equal(totalCredits) {
		errors = append(errors, ValidationError{
			Field:   "entries",
			Message: fmt.Sprintf("Transaction is not balanced: debits=%s, credits=%s", totalDebits.String(), totalCredits.String()),
			Code:    "TRANSACTION_UNBALANCED",
		})
	}

	return errors
}

// ValidateAccountHierarchy validates account hierarchy constraints
func (v *BusinessRuleValidator) ValidateAccountHierarchy(account *ChartOfAccounts, parentAccount *ChartOfAccounts) []ValidationError {
	var errors []ValidationError

	if parentAccount == nil {
		return errors // Root account, no hierarchy validation needed
	}

	// Business rule: Child account must have same root type as parent
	if account.RootType != parentAccount.RootType {
		errors = append(errors, ValidationError{
			Field:   "root_type",
			Message: "Child account must have the same root type as parent account",
			Code:    "INCONSISTENT_ROOT_TYPE",
		})
	}

	// Business rule: Account level must be parent level + 1
	if account.AccountLevel != parentAccount.AccountLevel+1 {
		errors = append(errors, ValidationError{
			Field:   "account_level",
			Message: "Account level must be exactly one more than parent level",
			Code:    "INVALID_HIERARCHY_LEVEL",
		})
	}

	// Business rule: Cannot create circular references
	if parentAccount.IsDescendantOf(account.ID) {
		errors = append(errors, ValidationError{
			Field:   "parent_account_id",
			Message: "Cannot create circular reference in account hierarchy",
			Code:    "CIRCULAR_REFERENCE",
		})
	}

	// Business rule: Control account cannot be child of another control account
	if account.IsControlAccount && parentAccount.IsControlAccount {
		errors = append(errors, ValidationError{
			Field:   "is_control_account",
			Message: "Control account cannot be a child of another control account",
			Code:    "INVALID_CONTROL_ACCOUNT_HIERARCHY",
		})
	}

	return errors
}

// ValidateTransactionDates validates date relationships in transactions
func (v *BusinessRuleValidator) ValidateTransactionDates(transactionDate time.Time, postingDate *time.Time, dueDate *time.Time) []ValidationError {
	var errors []ValidationError

	// Business rule: Posting date cannot be before transaction date
	if postingDate != nil && postingDate.Before(transactionDate) {
		errors = append(errors, ValidationError{
			Field:   "posting_date",
			Message: "Posting date cannot be before transaction date",
			Code:    "INVALID_DATE_SEQUENCE",
		})
	}

	// Business rule: Due date validation for payable transactions
	if dueDate != nil && dueDate.Before(transactionDate) {
		errors = append(errors, ValidationError{
			Field:   "due_date",
			Message: "Due date cannot be before transaction date",
			Code:    "INVALID_DATE_SEQUENCE",
		})
	}

	// Business rule: Transaction date should not be too far in the future
	maxFutureDate := time.Now().AddDate(1, 0, 0) // 1 year from now
	if transactionDate.After(maxFutureDate) {
		errors = append(errors, ValidationError{
			Field:   "transaction_date",
			Message: "Transaction date cannot be more than one year in the future",
			Code:    "FUTURE_DATE_TOO_FAR",
		})
	}

	// Business rule: Transaction date should not be too far in the past (reasonable business constraint)
	minDate := time.Date(MinAccountingYear, 1, 1, 0, 0, 0, 0, time.UTC)
	if transactionDate.Before(minDate) {
		errors = append(errors, ValidationError{
			Field:   "transaction_date",
			Message: fmt.Sprintf("Transaction date cannot be before %d", MinAccountingYear),
			Code:    "DATE_TOO_OLD",
		})
	}

	return errors
}

// ValidateApprovalWorkflow validates approval workflow rules
func (v *BusinessRuleValidator) ValidateApprovalWorkflow(transaction *Transaction) []ValidationError {
	var errors []ValidationError

	// Business rule: Approval required consistency
	if transaction.ApprovalRequired && transaction.ApprovalStatus == ApprovalStatusNotRequired {
		errors = append(errors, ValidationError{
			Field:   "approval_status",
			Message: "Approval status cannot be 'NOT_REQUIRED' when approval is required",
			Code:    "INCONSISTENT_APPROVAL_SETTINGS",
		})
	}

	if !transaction.ApprovalRequired && transaction.ApprovalStatus != ApprovalStatusNotRequired {
		errors = append(errors, ValidationError{
			Field:   "approval_status",
			Message: "Approval status must be 'NOT_REQUIRED' when approval is not required",
			Code:    "INCONSISTENT_APPROVAL_SETTINGS",
		})
	}

	// Business rule: Approved transactions must have approver information
	if transaction.ApprovalStatus == ApprovalStatusApproved {
		if transaction.ApprovedBy == nil {
			errors = append(errors, ValidationError{
				Field:   "approved_by",
				Message: "Approved by is required for approved transactions",
				Code:    "MISSING_APPROVER",
			})
		}

		if transaction.ApprovedAt == nil {
			errors = append(errors, ValidationError{
				Field:   "approved_at",
				Message: "Approved at timestamp is required for approved transactions",
				Code:    "MISSING_APPROVAL_TIMESTAMP",
			})
		}
	}

	// Business rule: Rejected transactions must have rejection reason
	if transaction.ApprovalStatus == ApprovalStatusRejected {
		if transaction.RejectionReason == nil ||
			(transaction.RejectionReason != nil &&
				strings.TrimSpace(string(*transaction.RejectionReason)) == "") {
			errors = append(errors, ValidationError{
				Field:   "rejection_reason",
				Message: "Rejection reason is required for rejected transactions",
				Code:    "MISSING_REJECTION_REASON",
			})
		}
	}

	// Business rule: Certain transaction types require approval
	if transaction.TransactionType.RequiresApproval() && !transaction.ApprovalRequired {
		errors = append(errors, ValidationError{
			Field:   "approval_required",
			Message: fmt.Sprintf("Transaction type %s requires approval", transaction.TransactionType),
			Code:    "APPROVAL_REQUIRED_FOR_TYPE",
		})
	}

	return errors
}

// ValidatePeriodClosure validates period closure constraints
func (v *BusinessRuleValidator) ValidatePeriodClosure(transactionDate time.Time, periodEndDate *time.Time, isPeriodClosed bool) []ValidationError {
	var errors []ValidationError

	if isPeriodClosed && periodEndDate != nil {
		if transactionDate.Before(*periodEndDate) || transactionDate.Equal(*periodEndDate) {
			errors = append(errors, ValidationError{
				Field:   "transaction_date",
				Message: "Cannot create transactions in a closed accounting period",
				Code:    "PERIOD_CLOSED",
			})
		}
	}

	return errors
}

// ValidateCurrencyConsistency validates currency consistency across transaction entries
func (v *BusinessRuleValidator) ValidateCurrencyConsistency(baseCurrency string, entries []TransactionEntry) []ValidationError {
	var errors []ValidationError

	for i, entry := range entries {
		// If entry has a different currency than base, it must have exchange rate
		if *entry.OriginalCurrency != baseCurrency {
			if entry.ExchangeRate.IsZero() {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("entries[%d].exchange_rate", i),
					Message: "Exchange rate is required for foreign currency transactions",
					Code:    "MISSING_EXCHANGE_RATE",
				})
			} else {
				// Validate the exchange rate
				exchangeErrors := v.ValidateExchangeRate(entry.ExchangeRate, *entry.OriginalCurrency, baseCurrency)
				for _, err := range exchangeErrors {
					err.Field = fmt.Sprintf("entries[%d].%s", i, err.Field)
					errors = append(errors, err)
				}
			}
		} else if !entry.ExchangeRate.IsZero() && !entry.ExchangeRate.Equal(decimal.NewFromInt(1)) {
			// Same currency should have exchange rate of 1.0 or nil
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("entries[%d].exchange_rate", i),
				Message: "Exchange rate should be 1.0 or null for base currency transactions",
				Code:    "UNNECESSARY_EXCHANGE_RATE",
			})
		}
	}

	return errors
}

// CompoundValidator combines multiple validators
type CompoundValidator struct {
	fieldValidator        *FieldValidator
	businessRuleValidator *BusinessRuleValidator
}

// NewCompoundValidator creates a new compound validator
func NewCompoundValidator() *CompoundValidator {
	return &CompoundValidator{
		fieldValidator:        &FieldValidator{},
		businessRuleValidator: &BusinessRuleValidator{},
	}
}

// ValidateAccount performs comprehensive account validation
func (v *CompoundValidator) ValidateAccount(account *ChartOfAccounts, parentAccount *ChartOfAccounts) *ValidationResult {
	result := &ValidationResult{IsValid: true}

	// Field-level validation
	result.Errors = append(result.Errors, account.Validate()...)

	// Business rule validation
	result.Errors = append(result.Errors, v.businessRuleValidator.ValidateAccountHierarchy(account, parentAccount)...)

	// Set result status
	result.IsValid = len(result.Errors) == 0

	return result
}

// ValidateTransaction performs comprehensive transaction validation
func (v *CompoundValidator) ValidateTransaction(transaction *Transaction) *ValidationResult {
	result := &ValidationResult{IsValid: true}

	// Field-level validation
	result.Errors = append(result.Errors, transaction.Validate()...)

	// Business rule validation
	result.Errors = append(result.Errors, v.businessRuleValidator.ValidateTransactionBalance(transaction.Entries)...)
	result.Errors = append(result.Errors, v.businessRuleValidator.ValidateTransactionDates(transaction.TransactionDate, transaction.PostingDate, transaction.DueDate)...)
	result.Errors = append(result.Errors, v.businessRuleValidator.ValidateApprovalWorkflow(transaction)...)

	// Additional validations
	// TODO: for the future get default setttings fron tenant oor Entity y to get the currency if matched account currency
	// if transaction.Entity != nil && transaction.Entity.BaseCurrency != "" {
	// 	result.Errors = append(result.Errors, v.businessRuleValidator.ValidateCurrencyConsistency(transaction.Entity.BaseCurrency, transaction.Entries)...)
	// }
	//
	// Set result status
	result.IsValid = len(result.Errors) == 0

	return result
}

// ValidateTransactionEntry performs validation for individual transaction entries
func (v *CompoundValidator) ValidateTransactionEntry(entry *TransactionEntry, baseCurrency string) *ValidationResult {
	result := &ValidationResult{IsValid: true}

	// Field-level validation
	result.Errors = append(result.Errors, entry.Validate()...)

	// Amount validation
	result.Errors = append(result.Errors, v.fieldValidator.ValidateNonNegativeDecimal(entry.DebitAmount, "debit_amount")...)
	result.Errors = append(result.Errors, v.fieldValidator.ValidateNonNegativeDecimal(entry.CreditAmount, "credit_amount")...)

	// Currency validation
	if *entry.OriginalCurrency != baseCurrency && entry.ExchangeRate.IsZero() {
		result.AddError("exchange_rate", "Exchange rate is required for foreign currency entries", "MISSING_EXCHANGE_RATE")
	}

	// Tax validation
	if !entry.TaxRate.IsZero() && entry.TaxCode != nil {
		result.Errors = append(result.Errors, v.businessRuleValidator.ValidateTaxRate(entry.TaxRate, entry.TaxCode)...)
	}

	// Set result status
	result.IsValid = len(result.Errors) == 0

	return result
}

// ValidateBatch validates a batch of transactions
func (v *CompoundValidator) ValidateBatch(transactions []*Transaction) *ValidationResult {
	result := &ValidationResult{IsValid: true}

	if len(transactions) == 0 {
		result.AddError("transactions", "Batch must contain at least one transaction", "EMPTY_BATCH")
		return result
	}

	// Track transaction numbers for uniqueness
	transactionNumbers := make(map[string]bool)

	for i, transaction := range transactions {
		// Validate individual transaction
		transactionResult := v.ValidateTransaction(transaction)

		// Prefix errors with transaction index
		for _, err := range transactionResult.Errors {
			err.Field = fmt.Sprintf("transactions[%d].%s", i, err.Field)
			result.Errors = append(result.Errors, err)
		}

		// Check for duplicate transaction numbers within batch
		if transaction.TransactionNumber != "" {
			if transactionNumbers[transaction.TransactionNumber] {
				result.AddError(fmt.Sprintf("transactions[%d].transaction_number", i),
					"Duplicate transaction number within batch", "DUPLICATE_TRANSACTION_NUMBER_IN_BATCH")
			}
			transactionNumbers[transaction.TransactionNumber] = true
		}
	}

	// Set result status
	result.IsValid = len(result.Errors) == 0

	return result
}

// ContextValidator provides validation with external context
type ContextValidator struct {
	compoundValidator *CompoundValidator
}

// NewContextValidator creates a new context validator
func NewContextValidator() *ContextValidator {
	return &ContextValidator{
		compoundValidator: NewCompoundValidator(),
	}
}

// ValidationContext holds context information for validation
type ValidationContext struct {
	TenantID       uuid.UUID
	EntityID       *uuid.UUID
	UserID         uuid.UUID
	BaseCurrency   string
	PeriodEndDate  *time.Time
	IsPeriodClosed bool
	// Add more context as needed for repository-based validations
}

// ValidateTransactionWithContext validates a transaction with additional context
func (v *ContextValidator) ValidateTransactionWithContext(transaction *Transaction, ctx *ValidationContext) *ValidationResult {
	result := v.compoundValidator.ValidateTransaction(transaction)

	// Additional context-based validations
	if ctx.PeriodEndDate != nil {
		periodErrors := v.compoundValidator.businessRuleValidator.ValidatePeriodClosure(
			transaction.TransactionDate, ctx.PeriodEndDate, ctx.IsPeriodClosed)
		result.Errors = append(result.Errors, periodErrors...)
	}

	// Currency consistency with entity
	if ctx.BaseCurrency != "" {
		currencyErrors := v.compoundValidator.businessRuleValidator.ValidateCurrencyConsistency(
			ctx.BaseCurrency, transaction.Entries)
		result.Errors = append(result.Errors, currencyErrors...)
	}

	// Set result status
	result.IsValid = len(result.Errors) == 0

	return result
}

// ValidateAccountWithContext validates an account with additional context
func (v *ContextValidator) ValidateAccountWithContext(account *ChartOfAccounts, parentAccount *ChartOfAccounts, ctx *ValidationContext) *ValidationResult {
	result := v.compoundValidator.ValidateAccount(account, parentAccount)

	// Additional context-based validations can be added here
	// For example, checking account code uniqueness against repository

	return result
}

// Utility functions for common validation scenarios

// IsBusinessDay checks if a date falls on a business day (Monday-Friday)
func IsBusinessDay(date time.Time) bool {
	weekday := date.Weekday()
	return weekday >= time.Monday && weekday <= time.Friday
}

// IsWithinBusinessHours checks if a timestamp falls within business hours
func IsWithinBusinessHours(timestamp time.Time) bool {
	hour := timestamp.Hour()
	return hour >= 9 && hour <= 17 // 9 AM to 5 PM
}

// ValidateBusinessDay validates that a date is a business day
func ValidateBusinessDay(date time.Time, fieldName string) []ValidationError {
	var errors []ValidationError

	if !IsBusinessDay(date) {
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: "Date must be a business day (Monday-Friday)",
			Code:    "NON_BUSINESS_DAY",
		})
	}

	return errors
}

// ValidateQuarter validates that a date falls within a specific quarter
func ValidateQuarter(date time.Time, year int, quarter int, fieldName string) []ValidationError {
	var errors []ValidationError

	var startMonth, endMonth time.Month
	switch quarter {
	case 1:
		startMonth, endMonth = time.January, time.March
	case 2:
		startMonth, endMonth = time.April, time.June
	case 3:
		startMonth, endMonth = time.July, time.September
	case 4:
		startMonth, endMonth = time.October, time.December
	default:
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: "Invalid quarter specified",
			Code:    "INVALID_QUARTER",
		})
		return errors
	}

	startDate := time.Date(year, startMonth, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, endMonth+1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1)

	if date.Before(startDate) || date.After(endDate) {
		errors = append(errors, ValidationError{
			Field: fieldName,
			Message: fmt.Sprintf("Date must be within Q%d %d (%s to %s)",
				quarter, year, startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
			Code: "DATE_NOT_IN_QUARTER",
		})
	}

	return errors
}

// Batch validation utilities

// BatchValidationResult holds results for batch operations
type BatchValidationResult struct {
	OverallValid   bool                `json:"overall_valid"`
	TotalRecords   int                 `json:"total_records"`
	ValidRecords   int                 `json:"valid_records"`
	InvalidRecords int                 `json:"invalid_records"`
	RecordResults  []*ValidationResult `json:"record_results"`
	GlobalErrors   []ValidationError   `json:"global_errors,omitempty"`
}

// HasGlobalErrors returns true if there are batch-level validation errors
func (bvr *BatchValidationResult) HasGlobalErrors() bool {
	return len(bvr.GlobalErrors) > 0
}

// AddGlobalError adds a batch-level validation error
func (bvr *BatchValidationResult) AddGlobalError(message, code string) {
	bvr.OverallValid = false
	bvr.GlobalErrors = append(bvr.GlobalErrors, ValidationError{
		Field:   "batch",
		Message: message,
		Code:    code,
	})
}

// GetInvalidRecordIndices returns indices of records that failed validation
func (bvr *BatchValidationResult) GetInvalidRecordIndices() []int {
	var indices []int
	for i, result := range bvr.RecordResults {
		if result.HasErrors() {
			indices = append(indices, i)
		}
	}
	return indices
}

// BatchValidator provides batch validation capabilities
type BatchValidator struct {
	contextValidator *ContextValidator
	maxBatchSize     int
}

// NewBatchValidator creates a new batch validator
func NewBatchValidator(maxBatchSize int) *BatchValidator {
	return &BatchValidator{
		contextValidator: NewContextValidator(),
		maxBatchSize:     maxBatchSize,
	}
}

// ValidateTransactionBatch validates a batch of transactions with comprehensive checks
func (v *BatchValidator) ValidateTransactionBatch(transactions []*Transaction, ctx *ValidationContext) *BatchValidationResult {
	result := &BatchValidationResult{
		OverallValid:  true,
		TotalRecords:  len(transactions),
		RecordResults: make([]*ValidationResult, len(transactions)),
	}

	// Check batch size limits
	if len(transactions) > v.maxBatchSize {
		result.AddGlobalError(fmt.Sprintf("Batch size %d exceeds maximum allowed size of %d",
			len(transactions), v.maxBatchSize), "BATCH_SIZE_EXCEEDED")
		return result
	}

	if len(transactions) == 0 {
		result.AddGlobalError("Batch cannot be empty", "EMPTY_BATCH")
		return result
	}

	// Track for batch-level uniqueness checks
	transactionNumbers := make(map[string]int)

	// Validate each transaction
	for i, transaction := range transactions {
		transactionResult := v.contextValidator.ValidateTransactionWithContext(transaction, ctx)
		result.RecordResults[i] = transactionResult

		if transactionResult.HasErrors() {
			result.InvalidRecords++
			result.OverallValid = false
		} else {
			result.ValidRecords++
		}

		// Check for duplicate transaction numbers within batch
		if transaction.TransactionNumber != "" {
			if existingIndex, exists := transactionNumbers[transaction.TransactionNumber]; exists {
				// Mark both records as having duplicate transaction numbers
				result.RecordResults[i].AddError("transaction_number",
					fmt.Sprintf("Duplicate transaction number with record at index %d", existingIndex),
					"DUPLICATE_TRANSACTION_NUMBER_IN_BATCH")
				result.RecordResults[existingIndex].AddError("transaction_number",
					fmt.Sprintf("Duplicate transaction number with record at index %d", i),
					"DUPLICATE_TRANSACTION_NUMBER_IN_BATCH")
				result.InvalidRecords++
				if result.ValidRecords > 0 {
					result.ValidRecords--
				}
			}
			transactionNumbers[transaction.TransactionNumber] = i
		}
	}

	return result
}

// package models
//
// import (
// 	"fmt"
// 	"regexp"
// 	"strings"
// 	"time"
//
// 	"github.com/google/uuid"
// 	"github.com/shopspring/decimal"
// )
//
// // ValidationResult holds the result of a validation operation
// type ValidationResult struct {
// 	IsValid bool              `json:"is_valid"`
// 	Errors  []ValidationError `json:"errors,omitempty"`
// }
//
// // HasErrors returns true if the validation result contains errors
// func (vr *ValidationResult) HasErrors() bool {
// 	return len(vr.Errors) > 0
// }
//
// // AddError adds a validation error to the result
// func (vr *ValidationResult) AddError(field, message, code string) {
// 	vr.IsValid = false
// 	vr.Errors = append(vr.Errors, ValidationError{
// 		Field:   field,
// 		Message: message,
// 		Code:    code,
// 	})
// }
//
// // Merge combines multiple validation results
// func (vr *ValidationResult) Merge(other *ValidationResult) {
// 	if other.HasErrors() {
// 		vr.IsValid = false
// 		vr.Errors = append(vr.Errors, other.Errors...)
// 	}
// }
//
// // Common validation patterns and constants
// const (
// 	// Regular expressions for validation
// 	AccountCodePattern      = `^[A-Za-z0-9\-\.]+$`
// 	CurrencyCodePattern     = `^[A-Z]{3}$`
// 	TransactionNumberPattern = `^[A-Za-z0-9\-]+$`
// 	TaxCodePattern          = `^[A-Za-z0-9\-]+$`
//
// 	// Maximum lengths
// 	MaxAccountCodeLength      = 20
// 	MaxAccountNameLength      = 200
// 	MaxDescriptionLength      = 1000
// 	MaxTransactionNumberLength = 50
// 	MaxReferenceLength        = 100
// 	MaxCostCenterLength       = 20
// 	MaxDepartmentLength       = 20
// 	MaxTaxCodeLength          = 20
//
// 	// Minimum values
// 	MinExchangeRate = 0.000001
// 	MaxExchangeRate = 1000000.0
// 	MaxTaxRate      = 100.0
// )
//
// // ValidatorFunc defines a validation function signature
// type ValidatorFunc func() []ValidationError
//
// // FieldValidator provides validation for individual fields
// type FieldValidator struct{}
//
// // ValidateUUID validates that a UUID is not nil/empty
// func (v *FieldValidator) ValidateUUID(value uuid.UUID, fieldName string, required bool) []ValidationError {
// 	var errors []ValidationError
//
// 	if required && value == uuid.Nil {
// 		errors = append(errors, ValidationError{
// 			Field:   fieldName,
// 			Message: fmt.Sprintf("%s is required", fieldName),
// 			Code:    "REQUIRED_FIELD",
// 		})
// 	}
//
// 	return errors
// }
//
// // ValidateString validates string fields with length and pattern constraints
// func (v *FieldValidator) ValidateString(value, fieldName string, required bool, maxLength int, pattern string) []ValidationError {
// 	var errors []ValidationError
//
// 	trimmedValue := strings.TrimSpace(value)
//
// 	if required && trimmedValue == "" {
// 		errors = append(errors, ValidationError{
// 			Field:   fieldName,
// 			Message: fmt.Sprintf("%s is required", fieldName),
// 			Code:    "REQUIRED_FIELD",
// 		})
// 		return errors // Don't continue validation if required field is empty
// 	}
//
// 	if len(value) > maxLength {
// 		errors = append(errors, ValidationError{
// 			Field:   fieldName,
// 			Message: fmt.Sprintf("%s must be %d characters or less", fieldName, maxLength),
// 			Code:    "MAX_LENGTH_EXCEEDED",
// 		})
// 	}
//
// 	if pattern != "" && trimmedValue != "" {
// 		matched, err := regexp.MatchString(pattern, trimmedValue)
// 		if err != nil || !matched {
// 			errors = append(errors, ValidationError{
// 				Field:   fieldName,
// 				Message: fmt.Sprintf("%s has invalid format", fieldName),
// 				Code:    "INVALID_FORMAT",
// 			})
// 		}
// 	}
//
// 	return errors
// }
//
// // ValidateOptionalString validates optional string fields
// func (v *FieldValidator) ValidateOptionalString(value *string, fieldName string, maxLength int, pattern string) []ValidationError {
// 	if value == nil {
// 		return nil
// 	}
// 	return v.ValidateString(*value, fieldName, false, maxLength, pattern)
// }
//
// // ValidateDecimal validates decimal fields with range constraints
// func (v *FieldValidator) ValidateDecimal(value decimal.Decimal, fieldName string, minValue, maxValue *decimal.Decimal) []ValidationError {
// 	var errors []ValidationError
//
// 	if minValue != nil && value.LessThan(*minValue) {
// 		errors = append(errors, ValidationError{
// 			Field:   fieldName,
// 			Message: fmt.Sprintf("%s must be greater than or equal to %s", fieldName, minValue.String()),
// 			Code:    "VALUE_OUT_OF_RANGE",
// 		})
// 	}
//
// 	if maxValue != nil && value.GreaterThan(*maxValue) {
// 		errors = append(errors, ValidationError{
// 			Field:   fieldName,
// 			Message: fmt.Sprintf("%s must be less than or equal to %s", fieldName, maxValue.String()),
// 			Code:    "VALUE_OUT_OF_RANGE",
// 		})
// 	}
//
// 	return errors
// }
//
// // ValidateDate validates date fields with range constraints
// func (v *FieldValidator) ValidateDate(value time.Time, fieldName string, minDate, maxDate *time.Time) []ValidationError {
// 	var errors []ValidationError
//
// 	if value.IsZero() {
// 		errors = append(errors, ValidationError{
// 			Field:   fieldName,
// 			Message: fmt.Sprintf("%s is required", fieldName),
// 			Code:    "REQUIRED_FIELD",
// 		})
// 		return errors
// 	}
//
// 	if minDate != nil && value.Before(*minDate) {
// 		errors = append(errors, ValidationError{
// 			Field:   fieldName,
// 			Message: fmt.Sprintf("%s cannot be before %s", fieldName, minDate.Format("2006-01-02")),
// 			Code:    "DATE_OUT_OF_RANGE",
// 		})
// 	}
//
// 	if maxDate != nil && value.After(*maxDate) {
// 		errors = append(errors, ValidationError{
// 			Field:   fieldName,
// 			Message: fmt.Sprintf("%s cannot be after %s", fieldName, maxDate.Format("2006-01-02")),
// 			Code:    "DATE_OUT_OF_RANGE",
// 		})
// 	}
//
// 	return errors
// }
//
// // BusinessRuleValidator provides business-specific validation
// type BusinessRuleValidator struct{}
//
// // ValidateAccountCode validates account code format and uniqueness constraints
// func (v *BusinessRuleValidator) ValidateAccountCode(code string, tenantID uuid.UUID, entityID *uuid.UUID) []ValidationError {
// 	var errors []ValidationError
//
// 	fieldValidator := &FieldValidator{}
// 	errors = append(errors, fieldValidator.ValidateString(code, "account_code", true, MaxAccountCodeLength, AccountCodePattern)...)
//
// 	// Business rule: Account codes should not start with special characters
// 	if len(code) > 0 && (code[0] == '-' || code[0] == '.') {
// 		errors = append(errors, ValidationError{
// 			Field:   "account_code",
// 			Message: "Account code cannot start with '-' or '.'",
// 			Code:    "INVALID_ACCOUNT_CODE_FORMAT",
// 		})
// 	}
//
// 	// TODO: Check uniqueness in repository
// 	// if accountCodeExists(code, tenantID, entityID) {
// 	//     errors = append(errors, ValidationError{
// 	//         Field:   "account_code",
// 	//         Message: "Account code already exists",
// 	//         Code:    "DUPLICATE_ACCOUNT_CODE",
// 	//     })
// 	// }
//
// 	return errors
// }
//
// // ValidateTransactionNumber validates transaction number format and uniqueness
// func (v *BusinessRuleValidator) ValidateTransactionNumber(number string, tenantID uuid.UUID, entityID *uuid.UUID) []ValidationError {
// 	var errors []ValidationError
//
// 	fieldValidator := &FieldValidator{}
// 	errors = append(errors, fieldValidator.ValidateString(number, "transaction_number", true, MaxTransactionNumberLength, TransactionNumberPattern)...)
//
// 	// TODO: Check uniqueness in repository
// 	// if transactionNumberExists(number, tenantID, entityID) {
// 	//     errors = append(errors, ValidationError{
// 	//         Field:   "transaction_number",
// 	//         Message: "Transaction number already exists",
// 	//         Code:    "DUPLICATE_TRANSACTION_NUMBER",
// 	//     })
// 	// }
//
// 	return errors
// }
//
// // ValidateExchangeRate validates exchange rate values
// func (v *BusinessRuleValidator) ValidateExchangeRate(rate decimal.Decimal, fromCurrency, toCurrency string) []ValidationError {
// 	var errors []ValidationError
//
// 	minRate := decimal.NewFromFloat(MinExchangeRate)
// 	maxRate := decimal.NewFromFloat(MaxExchangeRate)
//
// 	fieldValidator := &FieldValidator{}
// 	errors = append(errors, fieldValidator.ValidateDecimal(rate, "exchange_rate", &minRate, &maxRate)...)
//
// 	// Business rule: Same currency should have rate of 1.0
// 	if fromCurrency == toCurrency && !rate.Equal(decimal.NewFromInt(1)) {
// 		errors = append(errors, ValidationError{
// 			Field:   "exchange_rate",
// 			Message: "Exchange rate must be 1.0 for same currency conversion",
// 			Code:    "INVALID_SAME_CURRENCY_RATE",
// 		})
// 	}
//
// 	return errors
// }
//
// // ValidateTaxRate validates tax rate values
// func (v *BusinessRuleValidator) ValidateTaxRate(rate decimal.Decimal, taxCode *string) []ValidationError {
// 	var errors []ValidationError
//
// 	minRate := decimal.Zero
// 	maxRate := decimal.NewFromFloat(MaxTaxRate)
//
// 	fieldValidator := &FieldValidator{}
// 	errors = append(errors, fieldValidator.ValidateDecimal(rate, "tax_rate", &minRate, &maxRate)...)
//
// 	// Business rule: Tax rate should be zero if no tax code provided
// 	if taxCode == nil && !rate.IsZero() {
// 		errors = append(errors, ValidationError{
// 			Field:   "tax_rate",
// 			Message: "Tax rate should be zero when no tax code is provided",
// 			Code:    "INCONSISTENT_TAX_DATA",
// 		})
// 	}
//
// 	return errors
// }
//
// // ValidateTransactionBalance validates that transaction debits equal credits
// func (v *BusinessRuleValidator) ValidateTransactionBalance(entries []TransactionEntry) []ValidationError {
// 	var errors []ValidationError
//
// 	totalDebits := decimal.Zero
// 	totalCredits := decimal.Zero
//
// 	for _, entry := range entries {
// 		totalDebits = totalDebits.Add(entry.DebitAmount)
// 		totalCredits = totalCredits.Add(entry.CreditAmount)
// 	}
//
// 	if !totalDebits.Equal(totalCredits) {
// 		errors = append(errors, ValidationError{
// 			Field:   "entries",
// 			Message: fmt.Sprintf("Transaction is not balanced: debits=%s, credits=%s", totalDebits.String(), totalCredits.String()),
// 			Code:    "TRANSACTION_UNBALANCED",
// 		})
// 	}
//
// 	return errors
// }
//
// // ValidateAccountHierarchy validates account hierarchy constraints
// func (v *BusinessRuleValidator) ValidateAccountHierarchy(account *ChartOfAccounts, parentAccount *ChartOfAccounts) []ValidationError {
// 	var errors []ValidationError
//
// 	if parentAccount == nil {
// 		return errors // Root account, no hierarchy validation needed
// 	}
//
// 	// Business rule: Child account must have same root type as parent
// 	if account.RootType != parentAccount.RootType {
// 		errors = append(errors, ValidationError{
// 			Field:   "root_type",
// 			Message: "Child account must have the same root type as parent account",
// 			Code:    "INCONSISTENT_ROOT_TYPE",
// 		})
// 	}
//
// 	// Business rule: Account level must be parent level + 1
// 	if account.AccountLevel != parentAccount.AccountLevel+1 {
// 		errors = append(errors, ValidationError{
// 			Field:   "account_level",
// 			Message: "Account level must be exactly one more than parent level",
// 			Code:    "INVALID_HIERARCHY_LEVEL",
// 		})
// 	}
//
// 	// Business rule: Cannot create circular references
// 	if parentAccount.IsDescendantOf(account.ID) {
// 		errors = append(errors, ValidationError{
// 			Field:   "parent_account_id",
// 			Message: "Cannot create circular reference in account hierarchy",
// 			Code:    "CIRCULAR_REFERENCE",
// 		})
// 	}
//
// 	// Business rule: Control account cannot be child of another control account
// 	if account.IsControlAccount && parentAccount.IsControlAccount {
// 		errors = append(errors, ValidationError{
// 			Field:   "is_control_account",
// 			Message: "Control account cannot be a child of another control account",
// 			Code:    "INVALID_CONTROL_ACCOUNT_HIERARCHY",
// 		})
// 	}
//
// 	return errors
// }
//
// // ValidateTransactionDates validates date relationships in transactions
// func (v *BusinessRuleValidator) ValidateTransactionDates(transactionDate time.Time, postingDate *time.Time, dueDate *time.Time) []ValidationError {
// 	var errors []ValidationError
//
// 	// Business rule: Posting date cannot be before transaction date
// 	if postingDate != nil && postingDate.Before(transactionDate) {
// 		errors = append(errors, ValidationError{
// 			Field:   "posting_date",
// 			Message: "Posting date cannot be before transaction date",
// 			Code:    "INVALID_DATE_SEQUENCE",
// 		})
// 	}
//
// 	// Business rule: Due date validation for payable transactions
// 	if dueDate != nil && dueDate.Before(transactionDate) {
// 		errors = append(errors, ValidationError{
// 			Field:   "due_date",
// 			Message: "Due date cannot be before transaction date",
// 			Code:    "INVALID_DATE_SEQUENCE",
// 		})
// 	}
//
// 	// Business rule: Transaction date should not be too far in the future
// 	maxFutureDate := time.Now().AddDate(1, 0, 0) // 1 year from now
// 	if transactionDate.After(maxFutureDate) {
// 		errors = append(errors, ValidationError{
// 			Field:   "transaction_date",
// 			Message: "Transaction date cannot be more than one year in the future",
// 			Code:    "FUTURE_DATE_TOO_FAR",
// 		})
// 	}
//
// 	return errors
// }
//
// // ValidateApprovalWorkflow validates approval workflow rules
// func (v *BusinessRuleValidator) ValidateApprovalWorkflow(transaction *Transaction) []ValidationError {
// 	var errors []ValidationError
//
// 	// Business rule: Approval required consistency
// 	if transaction.ApprovalRequired && transaction.ApprovalStatus == ApprovalStatusNotRequired {
// 		errors = append(errors, ValidationError{
// 			Field:   "approval_status",
// 			Message: "Approval status cannot be 'NOT_REQUIRED' when approval is required",
// 			Code:    "INCONSISTENT_APPROVAL_SETTINGS",
// 		})
// 	}
//
// 	if !transaction.ApprovalRequired && transaction.ApprovalStatus != ApprovalStatusNotRequired {
// 		errors = append(errors, ValidationError{
// 			Field:   "approval_status",
// 			Message: "Approval status must be 'NOT_REQUIRED' when approval is not required",
// 			Code:    "INCONSISTENT_APPROVAL_SETTINGS",
// 		})
// 	}
//
// 	// Business rule: Approved transactions must have approver information
// 	if transaction.ApprovalStatus == ApprovalStatusApproved {
// 		if transaction.ApprovedBy == nil {
// 			errors = append(errors, ValidationError{
// 				Field:   "approved_by",
// 				Message: "Approved by is required for approved transactions",
// 				Code:    "MISSING_APPROVER",
// 			})
// 		}
//
// 		if transaction.ApprovedAt == nil {
// 			errors = append(errors, ValidationError{
// 				Field:   "approved_at",
// 				Message: "Approved at timestamp is required for approved transactions",
// 				Code:    "MISSING_APPROVAL_TIMESTAMP",
// 			})
// 		}
// 	}
//
// 	// Business rule: Certain transaction types require approval
// 	if transaction.TransactionType.RequiresApproval() && !transaction.ApprovalRequired {
// 		errors = append(errors, ValidationError{
// 			Field:   "approval_required",
// 			Message: fmt.Sprintf("Transaction type %s requires approval", transaction.TransactionType),
// 			Code:    "APPROVAL_REQUIRED_FOR_TYPE",
// 		})
// 	}
//
// 	return errors
// }
//
// // CompoundValidator combines multiple validators
// type CompoundValidator struct {
// 	fieldValidator        *FieldValidator
// 	businessRuleValidator *BusinessRuleValidator
// }
//
// // NewCompoundValidator creates a new compound validator
// func NewCompoundValidator() *CompoundValidator {
// 	return &CompoundValidator{
// 		fieldValidator:        &FieldValidator{},
// 		businessRuleValidator: &BusinessRuleValidator{},
// 	}
// }
//
// // ValidateAccount performs comprehensive account validation
// func (v *CompoundValidator) ValidateAccount(account *ChartOfAccounts, parentAccount *ChartOfAccounts) *ValidationResult {
// 	result := &ValidationResult{IsValid: true}
//
// 	// Field-level validation
// 	result.Errors = append(result.Errors, account.Validate()...)
//
// 	// Business rule validation
// 	result.Errors = append(result.Errors, v.businessRuleValidator.ValidateAccountHierarchy(account, parentAccount)...)
//
// 	// Set result status
// 	result.IsValid = len(result.Errors) == 0
//
// 	return result
// }
//
// // ValidateTransaction performs comprehensive transaction validation
// func (v *CompoundValidator) ValidateTransaction(transaction *Transaction) *ValidationResult {
// 	result := &ValidationResult{IsValid: true}
//
// 	// Field-level validation
// 	result.Errors = append(result.Errors, transaction.Validate()...)
//
// 	// Business rule validation
// 	result.Errors = append(result.Errors, v.businessRuleValidator.ValidateTransactionBalance(transaction.Entries)...)
// 	result.Errors = append(result.Errors, v.businessRuleValidator.ValidateTransactionDates(transaction.TransactionDate, transaction.PostingDate, transaction.DueDate)
