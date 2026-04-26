package domain

import "errors"

// Business logic errors
var (
	// Account errors
	ErrAccountNotFound          = errors.New("account not found")
	ErrAccountCodeExists        = errors.New("account code already exists")
	ErrAccountHasChildren       = errors.New("account has child accounts")
	ErrAccountHasTransactions   = errors.New("account has transactions")
	ErrAccountInactive          = errors.New("account is inactive")
	ErrAccountNoManualEntries   = errors.New("account does not allow manual entries")
	ErrInvalidAccountCode       = errors.New("invalid account code")
	ErrInvalidAccountName       = errors.New("invalid account name")
	ErrInvalidRootType          = errors.New("invalid root type")
	ErrInvalidEntityType        = errors.New("invalid entity type")
	ErrIncompatibleAccountType  = errors.New("incompatible account type with parent")
	ErrCircularAccountReference = errors.New("circular account reference detected")

	// Transaction errors
	ErrCannotReverseReversal          = errors.New("cannot reverse a reversal transaction")
	ErrTransactionNotFound            = errors.New("transaction not found")
	ErrTransactionNumberExists        = errors.New("transaction number already exists")
	ErrTransactionAlreadyPosted       = errors.New("transaction already posted")
	ErrTransactionAlreadyApproved     = errors.New("transaction already approved")
	ErrTransactionNotApproved         = errors.New("transaction not approved")
	ErrTransactionNotPosted           = errors.New("transaction not posted")
	ErrTransactionAlreadyReversed     = errors.New("transaction already reversed")
	ErrTransactionNotRequireApproval  = errors.New("transaction does not require approval")
	ErrUnbalancedTransaction          = errors.New("transaction debits do not equal credits")
	ErrInvalidTransactionEntry        = errors.New("invalid transaction entry")
	ErrInsufficientTransactionEntries = errors.New("transaction must have at least 2 entries")

	// Period errors
	ErrPeriodNotFound = errors.New("accounting period not found for date")
	ErrPeriodClosed   = errors.New("accounting period is closed; no further postings allowed")

	// Exchange rate errors
	ErrExchangeRateNotFound = errors.New("exchange rate not found")

	// Cost center errors
	ErrCostCenterNotFound   = errors.New("cost centre not found")
	ErrCostCenterCodeExists = errors.New("cost centre code already exists")

	// Budget errors
	ErrBudgetNotFound        = errors.New("budget not found")
	ErrBudgetNotEditable     = errors.New("budget is not editable in its current status")
	ErrBudgetAlreadyApproved = errors.New("budget is already approved")

	// Tax errors
	ErrTaxAuthorityNotFound = errors.New("tax authority not found")
	ErrTaxAuthorityExists   = errors.New("tax authority code already exists")
	ErrTaxCodeNotFound      = errors.New("tax code not found")
	ErrTaxCodeExists        = errors.New("tax code already exists within tenant")

	// Bank reconciliation errors
	ErrStatementNotFound      = errors.New("bank statement not found")
	ErrStatementAlreadyClosed = errors.New("bank statement is already completed")
	ErrLineNotFound           = errors.New("bank statement line not found")
	ErrLineAlreadyMatched     = errors.New("statement line is already matched to a journal entry")
	ErrEntryAlreadyReconciled = errors.New("journal entry is already reconciled")

	// Validation errors
	ErrInvalidCurrencyCode = errors.New("invalid currency code")
	ErrInvalidExchangeRate = errors.New("invalid exchange rate")
	ErrValidationFailed    = errors.New("validation failed")
	ErrDuplicateEntry      = errors.New("duplicate entry")
	ErrInvalidDateRange    = errors.New("invalid date range")

	// Authorization errors
	ErrUnauthorizedAccess      = errors.New("unauthorized access")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
)
