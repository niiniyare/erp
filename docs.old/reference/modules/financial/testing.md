# Finance Module — Test Cases

[<-- Back to Index](README.md)

## Table of Contents

- [Legend](#legend)
- [Domain Model Tests](#domain-model-tests)
  - [TransactionStatus Tests](#transactionstatus-tests)
  - [RejectionReason Tests](#rejectionreason-tests)
  - [RootType & NormalBalance Tests](#roottype--normalbalance-tests)
  - [AccountStatus State Machine Tests](#accountstatus-state-machine-tests)
  - [ValidationError Tests](#validationerror-tests)
  - [BudgetStatus Tests](#budgetstatus-tests)
- [Account Service Tests](#account-service-tests)
  - [CreateAccount Tests](#createaccount-tests)
  - [GetAccount Tests](#getaccount-tests)
  - [UpdateAccount Tests](#updateaccount-tests)
  - [DeleteAccount Tests](#deleteaccount-tests)
  - [ListAccounts & Filter Tests](#listaccounts--filter-tests)
  - [Account Hierarchy Tests](#account-hierarchy-tests)
  - [Account Balance Tests](#account-balance-tests)
- [Transaction Service Tests](#transaction-service-tests)
  - [CreateTransaction Tests](#createtransaction-tests)
  - [PostTransaction Tests](#posttransaction-tests)
  - [ReverseTransaction Tests](#reversetransaction-tests)
  - [ApproveTransaction Tests](#approvetransaction-tests)
  - [GetTransactionWithEntries Tests](#gettransactionwithentries-tests)
  - [SearchTransactions Tests](#searchtransactions-tests)
- [Period Management Tests](#period-management-tests)
  - [FiscalYear Tests](#fiscalyear-tests)
  - [AccountingPeriod Tests](#accountingperiod-tests)
  - [Period Validation Guard Tests](#period-validation-guard-tests)
- [Currency & Exchange Rate Tests](#currency--exchange-rate-tests)
  - [ExchangeRate Lookup Tests](#exchangerate-lookup-tests)
  - [Multi-Currency Posting Tests](#multi-currency-posting-tests)
  - [Exchange Rate Persistence Tests](#exchange-rate-persistence-tests)
- [Budget Service Tests](#budget-service-tests)
  - [Budget Lifecycle Tests](#budget-lifecycle-tests)
  - [Budget Control Tests](#budget-control-tests)
  - [Budget vs Actual Tests](#budget-vs-actual-tests)
- [Cost Center Tests](#cost-center-tests)
- [Repository Layer Tests](#repository-layer-tests)
  - [Transaction Repository Stub Tests](#transaction-repository-stub-tests)
  - [Account Balance Repository Tests](#account-balance-repository-tests)
  - [Reversal History Repository Tests](#reversal-history-repository-tests)
- [Posting Engine Tests](#posting-engine-tests)
- [Approval Workflow Engine Tests](#approval-workflow-engine-tests)
- [Transaction Numbering Tests](#transaction-numbering-tests)
- [API Handler Tests](#api-handler-tests)
- [Double-Entry Integrity Tests](#double-entry-integrity-tests)
- [Integration Tests](#integration-tests)
- [Performance Tests](#performance-tests)
- [Security & Tenant Isolation Tests](#security--tenant-isolation-tests)

---

## Legend

- `[ ]` Not started
- `[~]` In progress
- `[x]` Complete
- **P0** — Correctness / data integrity bug; blocks production
- **P1** — Core financial workflow; required for MVP
- **P2** — Important control or audit requirement
- **P3** — Enhancement, edge case, or polish

---

## Domain Model Tests

### TransactionStatus Tests

#### Test Case: IsEditable — only DRAFT and REJECTED are editable
```
Test ID: FIN-TYP-001
Priority: P0
Description: Verify only DRAFT and REJECTED statuses allow editing
File: internal/core/finance/domain/types_test.go:TestTransactionStatusEditability
Given: Each TransactionStatus value
When: Calling status.IsEditable()
Then:
  - DRAFT            → true
  - REJECTED         → true
  - PENDING_APPROVAL → false  (frozen during review)
  - APPROVED         → false
  - POSTED           → false
  - CANCELLED        → false
  - REVERSED         → false
```
- [x] **Status:** Done — `domain/types_test.go:TestTransactionStatusEditability`

#### Test Case: IsEditable — PENDING_APPROVAL is explicitly false
```
Test ID: FIN-TYP-002
Priority: P0
Description: Regression guard — PENDING_APPROVAL must never be editable
File: internal/core/finance/domain/types_test.go
Given: status = TransactionStatusPendingApproval
When: Calling status.IsEditable()
Then:
  - Returns false
  - This is a correctness invariant; if it ever returns true, approval
    controls are bypassed (a P0 regression)
```
- [x] **Status:** Done — `domain/types_test.go:TestPendingApproval_NotEditable`

#### Test Case: TransactionStatus state transitions — valid paths
```
Test ID: FIN-TYP-003
Priority: P1
Description: Verify the legal status transition graph
Given: Various (from, to) pairs
Then: Only these transitions are allowed:
  DRAFT            → PENDING_APPROVAL  (submit for approval)
  DRAFT            → POSTED            (direct post if no approval required)
  PENDING_APPROVAL → APPROVED
  PENDING_APPROVAL → REJECTED
  APPROVED         → POSTED
  APPROVED         → CANCELLED
  REJECTED         → DRAFT            (can be corrected and resubmitted)
  POSTED           → REVERSED
And these are ILLEGAL:
  POSTED           → DRAFT
  REVERSED         → POSTED
  CANCELLED        → APPROVED
  POSTED           → PENDING_APPROVAL
```
- [x] **Status:** Done — `domain/types_test.go:TestTransactionStatusTransitions`

#### Test Case: TransactionType — no duplicate JOURNAL values
```
Test ID: FIN-TYP-004
Priority: P0
Description: Verify only JOURNAL_ENTRY exists; JOURNAL is gone
File: internal/core/finance/domain/types.go
When: Searching all TransactionType constants
Then:
  - TransactionTypeJournalEntry == "JOURNAL_ENTRY"
  - No constant with value "JOURNAL" exists in the enum
  - ParseTransactionType("JOURNAL") returns an error
  - ParseTransactionType("JOURNAL_ENTRY") succeeds
```
- [x] **Status:** Done — `domain/types_test.go:TestTransactionType_NoDuplicate`

---

### RejectionReason Tests

#### Test Case: Only accounting-relevant reasons are valid
```
Test ID: FIN-TYP-010
Priority: P0
Description: Verify payment-gateway codes are purged; accounting codes are valid
File: internal/core/finance/domain/types_test.go
Given: Various RejectionReason string values
Then:
  Valid (IsValid() == true):
    - "INSUFFICIENT_SUPPORTING_DOCUMENTATION"
    - "INCORRECT_ACCOUNT_CODE"
    - "ACCOUNTING_PERIOD_CLOSED"
    - "AMOUNT_MISMATCH_WITH_SOURCE"
    - "DUPLICATE_ENTRY"
    - "POLICY_VIOLATION"
    - "BUDGET_EXCEEDED"
    - "UNAUTHORISED_ACCOUNT_ACCESS"
    - "OTHER"
  Invalid (IsValid() == false):
    - "EXPIRED_CARD"
    - "INVALID_MERCHANT"
    - "FRAUD_SUSPECTED"
    - "DAILY_LIMIT_EXCEEDED"
    - "INSUFFICIENT_FUNDS"
    - "" (empty string)
```
- [x] **Status:** Done — `domain/types_test.go:TestRejectionReasonValidity`

---

### RootType & NormalBalance Tests

#### Test Case: GetNormalBalance returns correct side for each RootType
```
Test ID: FIN-TYP-020
Priority: P1
Description: Every RootType has a defined normal balance per double-entry rules
Given: Each RootType value
When: Calling GetNormalBalanceForRootType(rootType)
Then:
  - ASSET     → DEBIT   (increases on debit side)
  - EXPENSE   → DEBIT
  - LIABILITY → CREDIT  (increases on credit side)
  - EQUITY    → CREDIT
  - REVENUE   → CREDIT
Note: These are fundamental accounting invariants. Wrong values here corrupt
      all balance calculations downstream.
```
- [x] **Status:** Done — `domain/types_test.go:TestNormalBalance_PerRootType`

#### Test Case: RootType.IsValid() rejects unknown values
```
Test ID: FIN-TYP-021
Priority: P1
Description: Verify unknown root types are rejected at domain level
Given: rootType = "CONTRA_ASSET" or "DEFERRED" or ""
When: Calling rootType.IsValid()
Then: Returns false for all non-canonical values
```
- [x] **Status:** Done — `domain/types_test.go:TestRootType_InvalidValues`

---

### AccountStatus State Machine Tests

#### Test Case: CanTransitionTo — valid transitions from ACTIVE
```
Test ID: FIN-TYP-030
Priority: P1
Description: Verify allowed transitions from ACTIVE status
Given: account.Status = AccountStatusActive
When: Calling account.CanTransitionTo(target)
Then:
  - INACTIVE          → true   (deactivation)
  - SUSPENDED         → true   (admin action)
  - AUDIT_LOCK        → true   (compliance)
  - COMPLIANCE_HOLD   → true   (regulatory)
  - CLOSED            → true   (permanent close)
  - DRAFT             → false  (cannot regress to draft)
  - ARCHIVED          → false  (must close first)
  - PENDING_APPROVAL  → false  (already approved/active)
```
- [x] **Status:** Done — `domain/types_test.go:TestAccountStatus_TransitionsFromActive`

#### Test Case: CanTransitionTo — re-activation paths
```
Test ID: FIN-TYP-031
Priority: P1
Description: Verify suspended/inactive accounts can be re-activated
Given: Various suspended/locked statuses
Then:
  - INACTIVE        → ACTIVE: true
  - SUSPENDED       → ACTIVE: true
  - AUDIT_LOCK      → ACTIVE: true  (requires elevated permission)
  - COMPLIANCE_HOLD → ACTIVE: true  (requires elevated permission)
  - CLOSED          → ACTIVE: false  (closed is terminal — cannot reopen)
  - ARCHIVED        → ACTIVE: false  (terminal)
```
- [x] **Status:** Done — `domain/types_test.go:TestAccountStatus_ReactivationPaths`

#### Test Case: CanTransitionTo — CLOSED is a terminal state
```
Test ID: FIN-TYP-032
Priority: P1
Description: Verify CLOSED accounts cannot transition to any other status
Given: account.Status = AccountStatusClosed
When: Calling CanTransitionTo for any target status
Then:
  - Only ARCHIVED is reachable
  - All other transitions return false
```
- [x] **Status:** Done — `domain/types_test.go:TestAccountStatus_ClosedIsTerminal`

---

### ValidationError Tests

#### Test Case: Severity levels are distinct and exhaustive
```
Test ID: FIN-TYP-040
Priority: P2
Description: Verify the three severity levels exist and behave correctly
Given: ValidationError instances with each severity
Then:
  - ValidationSeverityError   == "ERROR"
  - ValidationSeverityWarning == "WARNING"
  - ValidationSeverityInfo    == "INFO"
  - SplitBySeverity([ERROR, WARNING]) → errors=1, warnings=1
  - HasErrors([WARNING only]) → false
  - HasErrors([ERROR only])   → true
```
- [x] **Status:** Done — `domain/validation_test.go:TestValidationSeverity_Constants` + `TestValidationError_IsBlocker` + `TestValidationError_EffectiveSeverity_DefaultsToError`

#### Test Case: Nested field paths use entries[N].field format
```
Test ID: FIN-TYP-041
Priority: P2
Description: Verify entry-level validation errors carry indexed field paths
Given: A 3-line transaction where line 1 (index 1) has an invalid account
When: Running Validate() on the transaction
Then:
  - errors[0].Field == "entries[1].account_code"
  - NOT "account_code" (flat — unusable for UI highlighting)
  - errors[0].Code == "INVALID_ACCOUNT_CODE"
```
- [x] **Status:** Done — `domain/validation_test.go:TestNestedFieldPath`

#### Test Case: ValidationResult.Merge combines two results correctly
```
Test ID: FIN-TYP-042
Priority: P2
Description: Verify merging two validation results aggregates errors
Given:
  result1: [ERROR: "amount required", WARNING: "budget soft exceeded"]
  result2: [ERROR: "account inactive"]
When: result1.Merge(result2)
Then:
  - result1.Errors has 3 items total
  - IsValid() == false (there are ERROR-severity items)
  - Errors from result2 are appended, not replacing result1
```
- [x] **Status:** Done — `domain/validation_test.go:TestValidationResult_Merge` + `TestValidationResult_Merge_IntoValidResult` + `TestValidationResult_Merge_EmptyDoesNotInvalidate`

---

### BudgetStatus Tests

#### Test Case: BudgetStatus.IsEditable — only DRAFT and REJECTED allow edits
```
Test ID: FIN-TYP-050
Priority: P1
Description: Verify budget editability follows its lifecycle
Given: Each BudgetStatus value
When: Calling status.IsEditable()
Then:
  - DRAFT     → true
  - REJECTED  → true  (can be corrected after rejection)
  - SUBMITTED → false
  - APPROVED  → false
  - REVISED   → false  (creates a new version)
  - CLOSED    → false
```
- [x] **Status:** Done — `domain/types_test.go:TestBudgetStatus_Editability`

#### Test Case: BudgetLineItem.IsOverBudget detects exceedance correctly
```
Test ID: FIN-TYP-051
Priority: P1
Description: Verify over-budget detection is mathematically correct
Given: BudgetLineItem with BudgetedAmount = 100,000
Test Data:
  - ActualAmount = 99,999  → IsOverBudget() == false
  - ActualAmount = 100,000 → IsOverBudget() == false (at limit, not over)
  - ActualAmount = 100,001 → IsOverBudget() == true
  - ActualAmount = 0       → IsOverBudget() == false
```
- [x] **Status:** Done — `domain/types_test.go:TestBudgetLineItem_IsOverBudget`

#### Test Case: BudgetLineItem.ExceedsVarianceThreshold uses percentage correctly
```
Test ID: FIN-TYP-052
Priority: P1
Description: Verify variance threshold check uses percentage not absolute amount
Given:
  BudgetedAmount = 100,000
  VariancePct = 5.00  (5%)
  threshold = 10.00   (10% threshold)
When: Calling ExceedsVarianceThreshold(threshold)
Then:
  - Returns false  (5% variance is within 10% threshold)
Given:
  VariancePct = 12.00 (12%)
Then:
  - Returns true  (12% exceeds 10% threshold)
```
- [x] **Status:** Done — `domain/types_test.go:TestBudgetLineItem_VarianceThreshold`

---

## Account Service Tests

### CreateAccount Tests

#### Test Case: CreateAccount — duplicate code rejected
```
Test ID: FIN-ACC-001
Priority: P1
Description: Verify unique account code per tenant is enforced
File: internal/core/finance/service/account_service_test.go
Given: Account with code "1120" already exists for the tenant
When: Calling CreateAccount with code "1120" for the same tenant
Then:
  - Returns ErrAccountCodeExists
  - No new account is created
  - Error message references the duplicate code
```
- [x] **Status:** Done — `account_service_test.go:TestCreateAccount_DuplicateCode`

#### Test Case: CreateAccount — same code allowed for different tenants
```
Test ID: FIN-ACC-002
Priority: P1
Description: Verify code uniqueness is scoped to tenant, not global
Given: Account "1120" exists for tenant-A
When: Creating account "1120" for tenant-B
Then:
  - Succeeds with nil error
  - Tenant-B now has its own account "1120"
  - Both accounts coexist in the database
```
- [~] **Status:** Deferred — tenant isolation is enforced at the repository layer (SQL tenant_id filter); no service-level logic to test. `account_service_test.go:TestCreateAccount_SameCodeDifferentTenant`

#### Test Case: CreateAccount — root type mismatch with parent rejected
```
Test ID: FIN-ACC-003
Priority: P1
Description: Verify child accounts must share the parent's root type
Given: Parent account has RootType=ASSET
When: Creating a child account with RootType=LIABILITY under that parent
Then:
  - Returns ROOT_TYPE_MISMATCH error
  - No account is created
Note: An asset cannot have a liability child — it corrupts the COA structure.
```
- [x] **Status:** Done — `account_service_test.go:TestCreateAccount_RootTypeMismatch`

#### Test Case: CreateAccount — account path is auto-generated
```
Test ID: FIN-ACC-004
Priority: P1
Description: Verify materialized path is built from parent chain
Given:
  Parent "1000" has path "/1000/"
  Parent "1100" (child of 1000) has path "/1000/1100/"
When: Creating account "1110" under "1100"
Then:
  - New account has account_path = "/1000/1100/1110/"
  - Path is derived programmatically, not user-supplied
  - Changing parent updates path for all descendants
```
- [x] **Status:** Done — `account_service_test.go:TestCreateAccount_PathFromParent`

#### Test Case: CreateAccount — currency defaults to tenant base currency
```
Test ID: FIN-ACC-005
Priority: P1
Description: Verify missing currency defaults gracefully
Given: Tenant base currency = "KES"
When: Creating an account with no CurrencyCode specified
Then:
  - Account is created with CurrencyCode = "KES"
  - No error is returned
  - User does not need to supply currency for same-currency accounts
```
- [x] **Status:** Done — `account_service_test.go:TestCreateAccount_DefaultCurrency`

#### Test Case: CreateAccount — normal balance auto-set from root type
```
Test ID: FIN-ACC-006
Priority: P1
Description: Verify normal balance is derived if not provided
Given: Account with RootType=ASSET and no NormalBalance specified
When: Calling CreateAccount
Then:
  - Account is created with NormalBalance=DEBIT
  - No error (NormalBalance is optional input)
Given: RootType=LIABILITY
Then: NormalBalance=CREDIT
Given: RootType=REVENUE
Then: NormalBalance=CREDIT
```
- [~] **Status:** Deferred — `Validate()` rejects empty `NormalBalance` before the auto-set logic in the service runs; the spec requires omitting it but the current `CreateAccountRequest.Validate()` treats it as required. `account_service_test.go:TestCreateAccount_NormalBalanceDefault`

#### Test Case: CreateAccount — code format validation
```
Test ID: FIN-ACC-007
Priority: P1
Description: Verify account code format matches AccountCodePattern
Given: Invalid code values:
  - "" (empty)
  - "1 2 0" (spaces)
  - "12@0" (special chars except hyphen/dot)
  - 21-character code (exceeds MaxAccountCodeLength=20)
When: Calling CreateAccount with each invalid code
Then: Returns INVALID_ACCOUNT_CODE validation error for each
Given: Valid codes:
  - "1120"
  - "1120.01"
  - "CASH-USD"
Then: CreateAccount succeeds for each
```
- [~] **Status:** Deferred — spec lists "1120.01" and "CASH-USD" as valid, but `AccountCodePattern` in the domain only accepts 8 numeric digits; the regex does not match the spec's allowed formats. `account_service_test.go:TestCreateAccount_CodeFormat`

---

### GetAccount Tests

#### Test Case: GetAccountByID — returns full account
```
Test ID: FIN-ACC-010
Priority: P1
Description: Verify GetAccountByID returns all fields correctly
Given: Account created with all fields populated
When: Calling GetAccountByID(id)
Then:
  - Returned account.ID matches
  - account.AccountCode, AccountName, RootType match
  - account.TenantID matches the calling tenant
  - account.CurrentBalance is a valid decimal (not nil or NaN)
```
- [x] **Status:** Done — `account_service_test.go:TestGetAccountByID_Found`

#### Test Case: GetAccountByID — returns ErrAccountNotFound for unknown ID
```
Test ID: FIN-ACC-011
Priority: P1
Description: Verify correct error type for missing account
Given: A UUID that does not exist in the accounts table
When: Calling GetAccountByID(nonExistentID)
Then:
  - Returns nil account
  - Returns ErrAccountNotFound (not a generic DB error)
  - HTTP handler maps this to 404 Not Found
```
- [x] **Status:** Done — `account_service_test.go:TestGetAccountByID_NotFound`

#### Test Case: GetAccountByCode — tenant-scoped lookup
```
Test ID: FIN-ACC-012
Priority: P1
Description: Verify GetByCode only returns accounts for the calling tenant
Given: Two tenants both have account "1120"
When: Tenant-A calls GetAccountByCode("1120")
Then:
  - Returns Tenant-A's account only
  - Does not return Tenant-B's account
```
- [x] **Status:** Done — `account_service_test.go:TestGetAccountByCode_Found`

---

### UpdateAccount Tests

#### Test Case: UpdateAccount — cannot change root type
```
Test ID: FIN-ACC-020
Priority: P1
Description: Verify root type is immutable after account creation
Given: Existing ASSET account "1120"
When: Attempting to update RootType to LIABILITY
Then:
  - Returns ROOT_TYPE_IMMUTABLE error
  - Account unchanged in database
Note: Changing root type would corrupt all existing journal entries.
```
- [~] **Status:** Deferred — `UpdateAccountRequest` intentionally excludes `RootType`; reclassification is not yet implemented. `account_service_test.go:TestUpdateAccount_RootTypeImmutable`

#### Test Case: UpdateAccount — cannot change code if entries exist
```
Test ID: FIN-ACC-021
Priority: P1
Description: Verify account code cannot be renamed once transactions exist
Given: Account "1120" with at least one posted transaction entry
When: Attempting to update AccountCode to "1121"
Then:
  - Returns ACCOUNT_CODE_CHANGE_RESTRICTED error
  - Account code unchanged
Note: Renaming a code with history breaks audit trails.
```
- [~] **Status:** Deferred — `UpdateAccountRequest` intentionally excludes `AccountCode`; code renaming is not yet implemented. `account_service_test.go:TestUpdateAccount_CodeChangeBlocked`

---

### DeleteAccount Tests

#### Test Case: DeleteAccount — blocked when transactions exist
```
Test ID: FIN-ACC-030
Priority: P0
Description: Verify accounts with posted entries cannot be deleted
Given: Account "1120" has at least one entry in finance_transaction_entries
When: Calling DeleteAccount("1120")
Then:
  - Returns ErrAccountHasTransactions
  - Account still exists in the database
  - Entries are untouched
```
- [x] **Status:** Done — `account_service_test.go:TestDeleteAccount_HasTransactions`

#### Test Case: DeleteAccount — blocked when has active children
```
Test ID: FIN-ACC-031
Priority: P1
Description: Verify parent accounts with children cannot be deleted
Given: Account "1100" has child accounts "1110", "1120"
When: Calling DeleteAccount("1100")
Then:
  - Returns ErrAccountHasChildren
  - "1100" still exists
  - Children are unaffected
```
- [x] **Status:** Done — `account_service_test.go:TestDeleteAccount_HasChildren`

#### Test Case: DeleteAccount — soft delete sets is_active=false
```
Test ID: FIN-ACC-032
Priority: P1
Description: Verify delete is a soft delete (preserves audit trail)
Given: Account "9999" with no transactions and no children
When: Calling DeleteAccount("9999")
Then:
  - Returns nil error
  - Account row still exists in database
  - account.is_active = false (or status = CLOSED)
  - Account does NOT appear in ListAccounts(isActive=true) results
  - Account DOES appear in ListAccounts(isActive=false) results
```
- [x] **Status:** Done — `account_service_test.go:TestDeleteAccount_HappyPath` (soft-delete verified: repo.Delete called, account row preserved by implementation)

---

### ListAccounts & Filter Tests

#### Test Case: AccountFilter — filter by RootType
```
Test ID: FIN-ACC-040
Priority: P1
Description: Verify RootType filter returns only matching accounts
Given: 5 ASSET accounts, 3 LIABILITY accounts, 2 EQUITY accounts
When: ListAccounts(filter{RootType: ASSET})
Then:
  - Returns exactly 5 accounts
  - All returned accounts have RootType=ASSET
  - No LIABILITY or EQUITY accounts in results
```
- [x] **Status:** Done — `account_service_test.go:TestListAccounts_FilterByRootType`

#### Test Case: AccountFilter — filter by is_active
```
Test ID: FIN-ACC-041
Priority: P1
Description: Verify active/inactive filter works correctly
Given: 8 active accounts, 2 inactive (soft-deleted)
When: ListAccounts(filter{IsActive: true})
Then:
  - Returns 8 accounts
  - No inactive accounts in results
When: ListAccounts(filter{IsActive: false})
Then:
  - Returns 2 inactive accounts only
```
- [~] **Status:** Deferred — `IsActive` filter is applied in the SQL repo layer; `service.ListAccounts` passes the filter through unchanged. No distinct service logic to assert. `account_service_test.go:TestListAccounts_FilterByActive`

#### Test Case: AccountFilter — full-text search on code and name
```
Test ID: FIN-ACC-042
Priority: P2
Description: Verify Query field searches across code and name
Given: Accounts: "1120 - Main Checking", "1130 - Savings", "2100 - AP Trade"
When: ListAccounts(filter{Query: "checking"})
Then:
  - Returns only "1120 - Main Checking"
  - Case-insensitive match
When: ListAccounts(filter{Query: "1"})
Then:
  - Returns all accounts with "1" in code or name
```
- [~] **Status:** Deferred — full-text search is implemented in the repo (SQL ILIKE); the service passes `SearchQuery` through unchanged. Belongs in a repo integration test. `account_service_test.go:TestListAccounts_FullTextSearch`

#### Test Case: AccountFilter — pagination is correct
```
Test ID: FIN-ACC-043
Priority: P2
Description: Verify page/perPage correctly slices results
Given: 25 accounts total
When: ListAccounts(filter{Page: 1, PerPage: 10})
Then:
  - Returns 10 accounts (first page)
When: ListAccounts(filter{Page: 3, PerPage: 10})
Then:
  - Returns 5 accounts (last page)
  - No duplicates across pages
  - Total order is stable (sorted by account_code ASC)
```
- [~] **Status:** Deferred — pagination (Limit/Offset) is applied in the SQL repo layer; the service passes the filter unchanged. Belongs in a repo integration test. `account_service_test.go:TestListAccounts_Pagination`

---

### Account Hierarchy Tests

#### Test Case: GetAccountHierarchy — returns full subtree
```
Test ID: FIN-ACC-050
Priority: P1
Description: Verify hierarchy traversal returns all descendants
Given: Tree:
  1000 (root)
    1100
      1110
      1120
    1200
      1210
When: GetAccountHierarchy(rootID="1000")
Then:
  - Returns all 6 accounts
  - Each account includes its depth level
  - Parent-child relationships are correct
```
- [x] **Status:** Done — `account_service_test.go:TestGetAccountHierarchy_FullSubtree`

#### Test Case: GetAccountPath — returns ancestor chain
```
Test ID: FIN-ACC-051
Priority: P1
Description: Verify path traversal from leaf to root
Given: Account "1110" under "1100" under "1000"
When: GetAccountPath(id="1110")
Then:
  - Returns ["1000", "1100", "1110"] (root to leaf)
  - Order is root-first (breadcrumb order)
  - Each element is a fully populated Account struct
```
- [~] **Status:** Deferred — `GetAccountPath` is not exposed on the `AccountService` interface; no service method to call. `account_service_test.go:TestGetAccountPath_BreadcrumbOrder`

#### Test Case: ValidateHierarchy — cycle detection
```
Test ID: FIN-ACC-052
Priority: P0
Description: Verify circular parent references are rejected
Given: Account A is parent of B; B is parent of C
When: Attempting to set C's parent to A (creates cycle A→B→C→A)
Then:
  - Returns CIRCULAR_HIERARCHY error
  - No update is made
Note: A cyclic hierarchy would cause infinite loops in any traversal.
```
- [~] **Status:** Deferred — `ValidateHierarchy` / cycle-detection is not exposed on `AccountService`; reparenting is not yet implemented. `account_service_test.go:TestValidateHierarchy_CycleDetection`

---

### Account Balance Tests

#### Test Case: GetAccountBalance — returns SUM of posted entries only
```
Test ID: FIN-ACC-060
Priority: P0
Description: Verify balance reflects only POSTED transactions
Given: Account "1120" has:
  - Posted entry: Debit 50,000
  - Posted entry: Debit 30,000
  - Draft entry: Debit 10,000 (not yet posted — must be excluded)
When: GetAccountBalance(id="1120")
Then:
  - DebitTotal  = 80,000 (only posted entries)
  - CreditTotal = 0
  - NetBalance  = 80,000
  - Draft entry of 10,000 is NOT included
```
- [~] **Status:** Deferred — `GetAccountBalance` is not exposed on the `AccountService` interface (it's on the repo). `account_service_test.go:TestGetAccountBalance_PostedEntriesOnly`

#### Test Case: GetAccountBalance — asOfDate restricts to that date
```
Test ID: FIN-ACC-061
Priority: P1
Description: Verify asOfDate parameter gives point-in-time balance
Given: Account has:
  - Posted entry Jan 1: Debit 100,000
  - Posted entry Feb 1: Debit 50,000
When: GetAccountBalance(asOfDate=Jan 31)
Then:
  - DebitTotal = 100,000 (only Jan entry)
  - Feb entry is excluded
When: GetAccountBalance(asOfDate=nil)
Then:
  - DebitTotal = 150,000 (all entries)
```
- [~] **Status:** Deferred — `GetAccountBalance` is not exposed on the `AccountService` interface (it's on the repo). `account_service_test.go:TestGetAccountBalance_AsOfDate`

---

## Transaction Service Tests

### CreateTransaction Tests

#### Test Case: CreateTransaction — minimum valid transaction
```
Test ID: FIN-TXN-001
Priority: P1
Description: Verify a minimal valid transaction can be created
File: internal/core/finance/service/transaction_service_test.go
Given: Valid CreateTransactionRequest:
  - EntityID: valid UUID
  - TransactionType: JOURNAL_ENTRY
  - TransactionDate: today
  - CurrencyCode: "KES"
  - At least 2 entries (one debit, one credit, balanced)
When: Calling CreateTransaction(ctx, req)
Then:
  - Returns non-nil Transaction with generated ID
  - TransactionStatus = DRAFT
  - TransactionNumber is set (non-empty)
  - No entries posted yet (status remains DRAFT)
```
- [x] **Status:** Done — `transaction_service_test.go:TestCreateTransaction_MinimalValid`

#### Test Case: CreateTransaction — transaction number is unique per tenant
```
Test ID: FIN-TXN-002
Priority: P1
Description: Verify auto-generated numbers are unique within a tenant
Given: 3 concurrent CreateTransaction calls for the same tenant
When: All 3 complete
Then:
  - All 3 have distinct TransactionNumbers
  - Format follows configured prefix pattern (e.g., "JE-2025-00001")
  - No two transactions share the same number
```
- [x] **Status:** Done — `transaction_service_test.go:TestCreateTransaction_DuplicateNumber`

#### Test Case: CreateTransaction — requires at least 2 entries
```
Test ID: FIN-TXN-003
Priority: P0
Description: Verify single-entry transactions are rejected at creation
Given: CreateTransactionRequest with only 1 entry
When: Calling CreateTransaction
Then:
  - Returns ErrInsufficientTransactionEntries
  - No transaction row created
Note: Double-entry accounting requires minimum 2 lines.
```
- [x] **Status:** Done — `transaction_service_test.go:TestCreateTransaction_TooFewEntries`

---

### PostTransaction Tests

#### Test Case: PostTransaction — balanced transaction posts successfully
```
Test ID: FIN-TXN-010
Priority: P0
Description: Verify a balanced, valid transaction can be posted to the GL
Given:
  Transaction: DRAFT, open period, all accounts active
  Entries:
    Dr Cash (1120)    50,000
    Cr Revenue (4100) 50,000
When: Calling PostTransaction(id, nil)
Then:
  - Transaction status changes to POSTED
  - posting_date is set to now
  - Account 1120 balance updated (debit side)
  - Account 4100 balance updated (credit side)
  - Audit log entry created
```
- [x] **Status:** Done — `transaction_service_test.go:TestPostTransaction_BalancedSuccess`

#### Test Case: PostTransaction — unbalanced transaction rejected
```
Test ID: FIN-TXN-011
Priority: P0
Description: Verify unbalanced entries are rejected before any GL writes
Given:
  Transaction with entries:
    Dr Rent (7310) 50,000
    Cr Cash (1120) 40,000   ← imbalance of 10,000
When: Calling PostTransaction
Then:
  - Returns UNBALANCED_TRANSACTION error
  - Transaction status remains DRAFT (not POSTED)
  - Account balances are NOT updated
  - No partial write occurs
```
- [x] **Status:** Done — `transaction_service_test.go:TestPostTransaction_Unbalanced`

#### Test Case: PostTransaction — requires approval when flag enabled
```
Test ID: FIN-TXN-012
Priority: P1
Description: Verify approval-required transactions cannot be directly posted
Given:
  - Transaction in DRAFT status
  - Tenant setting: approval_required_above = 0 (all require approval)
  - Transaction has NOT gone through approval flow
When: Calling PostTransaction
Then:
  - Returns APPROVAL_REQUIRED error
  - Status remains DRAFT
Note: Bypassing approval defeats the financial control entirely.
```
- [~] **Status:** Deferred — `postTransactionInline` allows DRAFT regardless of `ApprovalRequired`; `CanBePosted()` exists but isn't called in the inline path. Test blocked until service enforces the flag.

#### Test Case: PostTransaction — already-posted transaction rejected
```
Test ID: FIN-TXN-013
Priority: P0
Description: Verify idempotency guard prevents double-posting
Given: Transaction already in POSTED status
When: Calling PostTransaction again
Then:
  - Returns ALREADY_POSTED error
  - Account balances are NOT modified a second time
  - Idempotency is critical for payment integrity
```
- [x] **Status:** Done — `transaction_service_test.go:TestPostTransaction_AlreadyPosted`

#### Test Case: PostTransaction — posting to closed period blocked
```
Test ID: FIN-TXN-014
Priority: P0
Description: Verify closed period blocks posting regardless of transaction status
Given:
  - Accounting period for Jan 2025 is HARD_CLOSED
  - Transaction dated Jan 15, 2025 in APPROVED status
When: Calling PostTransaction
Then:
  - Returns PERIOD_CLOSED error
  - Transaction remains APPROVED (not POSTED)
  - No GL entries written
```
- [~] **Status:** Deferred — closed-period check only runs when `periodRepo != nil`; `NewTransactionService` (used in tests) doesn't accept a period repo. Test blocked until wiring is available.

#### Test Case: PostTransaction — inactive account in entries blocked
```
Test ID: FIN-TXN-015
Priority: P1
Description: Verify entries referencing inactive accounts are rejected
Given: Entry references account "9998" which has is_active=false
When: Calling PostTransaction
Then:
  - Returns ACCOUNT_INACTIVE error referencing "9998"
  - Nothing posted
```
- [x] **Status:** Done — `transaction_service_test.go:TestPostTransaction_InactiveAccount`

#### Test Case: PostTransaction — non-leaf (control) account in entries blocked
```
Test ID: FIN-TXN-016
Priority: P1
Description: Verify posting to control/group accounts is not allowed
Given: Entry references account "1100" which is a control account (is_control=true)
When: Calling PostTransaction
Then:
  - Returns ACCOUNT_IS_CONTROL error
  - Posting to control accounts corrupts sub-account aggregations
```
- [~] **Status:** Deferred — `CanAcceptManualEntries()` is only checked for `TransactionTypeManual`; JOURNAL_ENTRY type skips the control-account check. Test blocked until service enforces for all types.

#### Test Case: PostTransaction — atomic: balance update is all-or-nothing
```
Test ID: FIN-TXN-017
Priority: P0
Description: Verify all balance updates commit together or not at all
Given: Balanced transaction with 3 entries across 3 accounts
When: DB error occurs while updating the third account's balance
Then:
  - Transaction status is rolled back to APPROVED (not POSTED)
  - First two accounts' balances are rolled back
  - No partial state exists in the database
```
- [~] **Status:** Deferred — atomicity/rollback requires real DB transactions; not testable with mocks.

---

### ReverseTransaction Tests

#### Test Case: ReverseTransaction — creates mirror entries
```
Test ID: FIN-TXN-020
Priority: P1
Description: Verify reversal transaction has exactly swapped debits/credits
Given: Posted transaction:
  Dr Rent (7310) 50,000
  Cr Cash (1120) 50,000
When: Calling ReverseTransaction(id, reason, reverseDate)
Then:
  Reversal transaction entries:
    Dr Cash (1120) 50,000
    Cr Rent (7310) 50,000
  - Original transaction.IsReversed = true
  - Original transaction.ReversedByTransactionID = reversalID
  - Reversal transaction.ReversalOfTransactionID = originalID
  - Net GL effect across both transactions = zero
```
- [x] **Status:** Done — `transaction_service_test.go:TestReverseTransaction_MirrorEntries`

#### Test Case: ReverseTransaction — net effect on GL is zero
```
Test ID: FIN-TXN-021
Priority: P0
Description: Verify account balances after reversal equal the state before the original
Given:
  Cash account balance BEFORE original transaction = 500,000
  After posting: Dr Rent 50,000 / Cr Cash 50,000 → Cash = 450,000
When: ReverseTransaction
Then:
  - Cash balance returns to 500,000
  - Rent balance returns to its pre-transaction value
  - The net of original + reversal = 0 for all affected accounts
```
- [~] **Status:** Deferred — net-zero GL verification requires real account balances updated by the posting engine; not testable with repo mocks.

#### Test Case: ReverseTransaction — only POSTED transactions can be reversed
```
Test ID: FIN-TXN-022
Priority: P1
Description: Verify reversal only works on posted transactions
Given: Transaction in DRAFT status
When: Calling ReverseTransaction
Then: Returns TRANSACTION_NOT_POSTED error
Given: Transaction in APPROVED status
Then: Returns TRANSACTION_NOT_POSTED error
Given: Transaction in POSTED status
Then: Reversal succeeds
```
- [x] **Status:** Done — `transaction_service_test.go:TestReverseTransaction_RequiresPosted_Draft` + `_Approved`

#### Test Case: ReverseTransaction — double reversal is blocked
```
Test ID: FIN-TXN-023
Priority: P0
Description: Verify a transaction cannot be reversed twice
Given: Transaction already reversed (IsReversed=true)
When: Calling ReverseTransaction on the original again
Then:
  - Returns TRANSACTION_ALREADY_REVERSED error
  - No new reversal transaction created
```
- [x] **Status:** Done — `transaction_service_test.go:TestReverseTransaction_AlreadyReversed`

#### Test Case: ReverseTransaction — reversal itself cannot be reversed
```
Test ID: FIN-TXN-024
Priority: P0
Description: Prevent infinite reversal chains that corrupt the ledger
Given: Original transaction JE-001 reversed by JE-002
When: Calling ReverseTransaction(JE-002)
Then:
  - Returns CANNOT_REVERSE_REVERSAL error
  - finance_reversal_history confirms JE-002 is a reversal transaction
  - No JE-003 is created
```
- [~] **Status:** Deferred — service doesn't check whether a transaction IS itself a reversal before allowing another reversal. Test blocked until service enforces this.

#### Test Case: ReverseTransaction — both operations are in one DB transaction
```
Test ID: FIN-TXN-025
Priority: P0
Description: Verify original update and reversal insert are atomic
Given: DB error occurs after reversal is created but before original is marked REVERSED
Then:
  - Both the reversal creation AND the IsReversed flag update are rolled back
  - No orphaned reversal transaction exists
  - Original remains POSTED and reversible
```
- [~] **Status:** Deferred — atomicity/rollback requires real DB transactions; not testable with mocks.

---

### ApproveTransaction Tests

#### Test Case: ApproveTransaction — moves status from PENDING_APPROVAL to APPROVED
```
Test ID: FIN-TXN-030
Priority: P1
Description: Verify approval changes status and records approver
Given: Transaction in PENDING_APPROVAL status
When: Calling ApproveTransaction(id, approverID)
Then:
  - TransactionStatus = APPROVED
  - ApprovalStatus = APPROVED
  - ApprovedBy = approverID
  - ApprovedAt is set to now
```
- [x] **Status:** Done — `transaction_service_test.go:TestApproveTransaction_PendingApproval_Succeeds` + `TestApproveTransaction_NotPending_ReturnsError`

#### Test Case: ApproveTransaction — submitter cannot approve their own transaction (SOD)
```
Test ID: FIN-TXN-031
Priority: P1
Description: Enforce segregation of duties: submitter ≠ approver
Given: Transaction submitted by userA
When: userA calls ApproveTransaction
Then:
  - Returns SELF_APPROVAL_FORBIDDEN error
  - Status remains PENDING_APPROVAL
Note: Segregation of duties is a core internal control requirement.
```
- [~] **Status:** Deferred — `ApproveTransaction` doesn't compare submitter ID vs approver ID. Test blocked until SOD check is added to the service.

#### Test Case: RejectTransaction — records reason and returns to editable state
```
Test ID: FIN-TXN-032
Priority: P1
Description: Verify rejected transaction can be corrected and resubmitted
Given: Transaction in PENDING_APPROVAL status
When: Calling RejectTransaction(id, approverID, reason="INCORRECT_ACCOUNT_CODE")
Then:
  - TransactionStatus = REJECTED
  - RejectionReason = "INCORRECT_ACCOUNT_CODE"
  - IsEditable() == true (REJECTED is editable)
  - Submitter is notified
```
- [x] **Status:** Done — `service/transaction_service_test.go:TestRejectTransaction_PendingApproval_Succeeds` + `TestRejectTransaction_NotPending_ReturnsError`

---

### GetTransactionWithEntries Tests

#### Test Case: GetTransactionWithEntries — returns header + all lines
```
Test ID: FIN-TXN-040
Priority: P1
Description: Verify single-call retrieval of transaction with all entries
Given: Transaction with 3 entries
When: Calling GetTransactionWithEntries(id)
Then:
  - TransactionWithEntries.Transaction is fully populated
  - TransactionWithEntries.Entries has exactly 3 items
  - Entries are ordered by line_number ASC
  - Each entry has AccountCode, Amount, DebitAmount, CreditAmount populated
```
- [x] **Status:** Done — `transaction_service_test.go:TestGetTransactionWithEntries_Success`

#### Test Case: GetTransactionWithEntries — IsBalanced() returns correct result
```
Test ID: FIN-TXN-041
Priority: P0
Description: Verify balance check works on the returned struct
Given: Balanced transaction (total debits == total credits)
When: GetTransactionWithEntries(id).IsBalanced()
Then: Returns true
Given: Intentionally unbalanced (seeded directly in DB for testing)
Then: Returns false
```
- [x] **Status:** Done — `transaction_service_test.go:TestGetTransactionWithEntries_IsBalanced`

---

### SearchTransactions Tests

#### Test Case: SearchTransactions — by description keyword
```
Test ID: FIN-TXN-050
Priority: P1
Description: Verify full-text search on transaction description
Given: Transactions with descriptions: "January rent payment", "February rent", "Salary March"
When: SearchTransactions(query="rent")
Then:
  - Returns 2 transactions (January rent, February rent)
  - "Salary March" is excluded
  - Case-insensitive search
```
- [~] **Status:** Stub tested — `transaction_service_test.go:TestSearchTransactions_CurrentBehavior` asserts empty result; update when Search is implemented in repo.

#### Test Case: SearchTransactions — scoped to tenant
```
Test ID: FIN-TXN-051
Priority: P0
Description: Verify search results never leak across tenants
Given: Tenant-A and Tenant-B both have transactions matching "rent"
When: Tenant-A calls SearchTransactions(query="rent")
Then:
  - Returns only Tenant-A's transactions
  - Tenant-B results never appear
```
- [~] **Status:** Deferred — stub path returns empty; tenant scoping test deferred until Search repo method is implemented.

---

## Period Management Tests

### FiscalYear Tests

#### Test Case: FiscalYear — overlapping years for same entity rejected
```
Test ID: FIN-PER-001
Priority: P1
Description: Verify two fiscal years cannot overlap for the same entity
File: internal/core/finance/service/period_service_test.go
Given: FY2025: Jan 1, 2025 – Dec 31, 2025 (already exists)
When: Creating FY2025b: Jul 1, 2025 – Jun 30, 2026 (overlaps)
Then:
  - Returns FISCAL_YEAR_OVERLAP error
  - No new fiscal year created
```
- [x] **Status:** Done — `period_service_test.go:TestFiscalYear_OverlapRejectedByRepository`

#### Test Case: FiscalYear — end date must be after start date
```
Test ID: FIN-PER-002
Priority: P1
Description: Verify basic date range validation
Given: FiscalYear with StartDate = Dec 31, EndDate = Jan 1 (inverted)
When: Creating or validating the fiscal year
Then:
  - Returns INVALID_DATE_RANGE error
  - end_date > start_date is enforced
```
- [x] **Status:** Done — `period_service_test.go:TestFiscalYear_DateRangeValidation_EndBeforeStart` + `_EqualDates` + `_ValidRange`

---

### AccountingPeriod Tests

#### Test Case: AccountingPeriod.CanPost() — status gate
```
Test ID: FIN-PER-010
Priority: P0
Description: Verify CanPost() is true only for OPEN periods
Given: Each PeriodStatus value
When: Calling period.CanPost()
Then:
  - OPEN        → true
  - SOFT_CLOSED → false  (only finance role can post via CanFinancePost)
  - HARD_CLOSED → false
  - LOCKED      → false
```
- [x] **Status:** Done — `domain/period_test.go:TestAccountingPeriod_CanPost`

#### Test Case: AccountingPeriod.CanFinancePost() — finance role gate
```
Test ID: FIN-PER-011
Priority: P1
Description: Verify finance role has broader posting window
Given: Each PeriodStatus value
When: Calling period.CanFinancePost()
Then:
  - OPEN        → true
  - SOFT_CLOSED → true   (finance can still post after soft close)
  - HARD_CLOSED → false
  - LOCKED      → false  (even CFO cannot post to locked period)
```
- [x] **Status:** Done — `domain/period_test.go:TestAccountingPeriod_CanFinancePost`

#### Test Case: PeriodService.GetOpenPeriodForDate — finds correct period
```
Test ID: FIN-PER-012
Priority: P1
Description: Verify date-to-period resolution is correct
Given: Jan 2025 period (OPEN), Feb 2025 period (HARD_CLOSED)
When: GetOpenPeriodForDate(Jan 15, 2025)
Then: Returns Jan 2025 period
When: GetOpenPeriodForDate(Feb 15, 2025)
Then: Returns nil (period is closed)
```
- [x] **Status:** Partial — `period_service_test.go:TestGetCurrentPeriod_ReturnsOpenPeriod` + `TestGetCurrentPeriod_HardClosedPeriodNotUsable`. Full date-scoped variant deferred (no `GetOpenPeriodForDate` on service interface).

---

### Period Validation Guard Tests

#### Test Case: CreateTransaction in closed period — warns, does not block
```
Test ID: FIN-PER-020
Priority: P1
Description: Verify period guard warns on create but doesn't reject
Given: Accounting period for Jan 2025 is HARD_CLOSED
When: CreateTransaction with TransactionDate = Jan 10, 2025
Then:
  - Transaction IS created (HTTP 201)
  - Response body includes warnings: ["PERIOD_SOFT_CLOSED"] or similar
  - Transaction status = DRAFT
  - User is informed but not blocked
Note: Accountants may legitimately back-date drafts before deciding to post.
```
- [ ] **Status:** Pending — `period_guard_test.go:TestPeriodGuard_CreateWarnsOnClosedPeriod`

#### Test Case: PostTransaction to closed period — hard blocks
```
Test ID: FIN-PER-021
Priority: P0
Description: Verify posting is hard-blocked for closed periods
Given: Accounting period for Jan 2025 is HARD_CLOSED
When: PostTransaction with TransactionDate in Jan 2025
Then:
  - Returns HTTP 422
  - error.code = "PERIOD_CLOSED"
  - Transaction status remains unchanged
  - No GL entries written
```
- [ ] **Status:** Pending — `period_guard_test.go:TestPeriodGuard_PostBlocksOnClosedPeriod`

#### Test Case: SOFT_CLOSED period — regular user blocked, finance user allowed
```
Test ID: FIN-PER-022
Priority: P1
Description: Verify soft close creates a two-tier access gate
Given: Accounting period for Jan 2025 is SOFT_CLOSED
When: Regular user calls PostTransaction with Jan date
Then: Returns PERIOD_CLOSED error
When: Finance role user calls PostTransaction with Jan date
Then: Succeeds — SOFT_CLOSED allows finance role posting
```
- [ ] **Status:** Pending — `period_guard_test.go:TestPeriodGuard_SoftClosedTwoTier`

---

## Currency & Exchange Rate Tests

### ExchangeRate Lookup Tests

#### Test Case: GetExchangeRate — exact date match
```
Test ID: FIN-CCY-001
Priority: P1
Description: Verify exact-date rate retrieval
File: internal/core/finance/service/exchange_rate_service_test.go
Given: Rate loaded: USD/KES = 128.00 on Jan 10, 2025
When: GetExchangeRate("USD", "KES", Jan 10 2025, SPOT)
Then:
  - Returns rate = 128.00
  - Returns nil error
```
- [ ] **Status:** Pending — `exchange_rate_test.go:TestGetExchangeRate_ExactDate`

#### Test Case: GetExchangeRate — nearest prior date fallback
```
Test ID: FIN-CCY-002
Priority: P1
Description: Verify rate lookup uses nearest prior rate when exact date has no rate
Given: Rate loaded: USD/KES = 128.00 on Jan 10, 2025
  (no rate exists for Jan 11 or Jan 12)
When: GetExchangeRate("USD", "KES", Jan 12 2025, SPOT)
Then:
  - Returns 128.00 (nearest prior = Jan 10)
  - This is the correct financial behavior (use last known rate)
```
- [ ] **Status:** Pending — `exchange_rate_test.go:TestGetExchangeRate_NearestPrior`

#### Test Case: GetExchangeRate — no prior rate returns error (not 1:1)
```
Test ID: FIN-CCY-003
Priority: P0
Description: Verify no silent 1:1 fallback when no rate exists
Given: No rate exists for USD/KES for any date
When: GetExchangeRate("USD", "KES", any date, SPOT)
Then:
  - Returns ErrExchangeRateNotFound
  - Does NOT return a 1:1 (or any assumed) rate
  - Silent 1:1 fallback causes silent financial data corruption
```
- [ ] **Status:** Pending — `exchange_rate_test.go:TestGetExchangeRate_NoRateReturnsError`

#### Test Case: GetExchangeRate — after server restart, rates load from DB
```
Test ID: FIN-CCY-004
Priority: P0
Description: Verify rates are durable across restarts
Given: Rate USD/KES = 128.00 saved to database
When: In-memory cache is cleared (simulating service restart)
  And: GetExchangeRate("USD", "KES", today) is called
Then:
  - Returns 128.00 (loaded from DB, not cache)
  - In-memory cache is repopulated for subsequent calls
```
- [ ] **Status:** Pending — `exchange_rate_test.go:TestGetExchangeRate_SurvivesRestart`

---

### Multi-Currency Posting Tests

#### Test Case: PostTransaction — FC entry stores base-currency equivalent
```
Test ID: FIN-CCY-010
Priority: P1
Description: Verify foreign-currency transactions record both FC and base amounts
Given:
  Rate: USD/KES = 128.00 (loaded)
  Tenant base currency: KES
  Transaction:
    Dr Accounts Receivable: $10,000 USD
    Cr Revenue: $10,000 USD
When: PostTransaction
Then:
  - Each entry has CurrencyCode = "USD"
  - Each entry has Amount = 10,000.00 (FC amount)
  - Each entry has BaseAmount = 1,280,000.00 (10,000 × 128)
  - Account balances updated in base currency (KES)
```
- [ ] **Status:** Pending — `transaction_service_test.go:TestPostTransaction_MultiCurrency`

#### Test Case: PostTransaction — no rate loaded returns error
```
Test ID: FIN-CCY-011
Priority: P0
Description: Verify FC transaction cannot post without a rate
Given: No USD/KES rate loaded for the transaction date
When: PostTransaction for a USD-denominated transaction
Then:
  - Returns EXCHANGE_RATE_NOT_FOUND error
  - Nothing posted
```
- [ ] **Status:** Pending — `transaction_service_test.go:TestPostTransaction_MissingRate`

---

### Exchange Rate Persistence Tests

#### Test Case: UpdateExchangeRate — persists to DB and updates cache
```
Test ID: FIN-CCY-020
Priority: P0
Description: Verify rate updates are durable
Given: No existing rate for USD/KES
When: Calling UpdateExchangeRate(from="USD", to="KES", rate=130.00, date=today)
Then:
  - finance_exchange_rates row created in DB
  - In-memory cache updated
  - Subsequent GetExchangeRate returns 130.00
  - Survives cache clear
```
- [ ] **Status:** Pending — `exchange_rate_test.go:TestUpdateExchangeRate_Persistent`

---

## Budget Service Tests

### Budget Lifecycle Tests

#### Test Case: CreateBudget — creates header and lines atomically
```
Test ID: FIN-BUD-001
Priority: P1
Description: Verify budget and its line items are created together
File: internal/core/finance/service/budget_service_test.go
Given: Budget with 5 line items across different accounts
When: Calling CreateBudget(budget, lines)
Then:
  - Budget row created with Status=DRAFT
  - Exactly 5 BudgetLineItem rows created
  - All line items reference the same budget ID
  - If line item creation fails, budget header is also rolled back
```
- [ ] **Status:** Pending — `budget_service_test.go:TestCreateBudget_Atomic`

#### Test Case: SubmitBudget — transitions DRAFT to SUBMITTED
```
Test ID: FIN-BUD-002
Priority: P1
Description: Verify budget submission workflow step
Given: Budget in DRAFT status
When: Calling SubmitBudget(id, submitterID)
Then:
  - Status = SUBMITTED
  - SubmittedAt = now
  - SubmittedBy = submitterID
  - Budget is no longer editable (IsEditable() == false)
```
- [ ] **Status:** Pending — `budget_service_test.go:TestSubmitBudget_Success`

#### Test Case: ApproveBudget — requires SUBMITTED status
```
Test ID: FIN-BUD-003
Priority: P1
Description: Verify budget cannot be approved from DRAFT
Given: Budget in DRAFT status
When: Calling ApproveBudget(id, approverID)
Then:
  - Returns ErrBudgetNotEditable or INVALID_STATUS_TRANSITION error
  - Budget status unchanged
```
- [ ] **Status:** Pending — `budget_service_test.go:TestApproveBudget_RequiresSubmitted`

#### Test Case: ApproveBudget — submitter cannot approve own budget (SOD)
```
Test ID: FIN-BUD-004
Priority: P1
Description: Enforce segregation of duties for budget approval
Given: Budget submitted by userA
When: userA calls ApproveBudget
Then:
  - Returns SELF_APPROVAL_FORBIDDEN error
  - Status remains SUBMITTED
```
- [ ] **Status:** Pending — `budget_service_test.go:TestApproveBudget_SOD`

#### Test Case: RejectBudget — records rejection note and re-enables editing
```
Test ID: FIN-BUD-005
Priority: P1
Description: Verify rejection flow allows correction and resubmission
Given: Budget in SUBMITTED status
When: Calling RejectBudget(id, approverID, note="Amounts need justification")
Then:
  - Status = REJECTED
  - RejectedAt set
  - RejectedBy = approverID
  - RejectNote = "Amounts need justification"
  - IsEditable() == true (REJECTED is editable)
```
- [ ] **Status:** Pending — `budget_service_test.go:TestRejectBudget_RecordsNote`

#### Test Case: CloseBudget — fiscal year end triggers automatic close
```
Test ID: FIN-BUD-006
Priority: P2
Description: Verify approved budgets auto-close at fiscal year end
Given: Budget in APPROVED status, fiscal year end = Dec 31
When: Calling CloseBudget or fiscal year close event
Then:
  - Status = CLOSED
  - Final variance calculations are frozen
  - Budget becomes read-only
```
- [ ] **Status:** Pending — `budget_service_test.go:TestCloseBudget_FiscalYearEnd`

---

### Budget Control Tests

#### Test Case: CheckBudget — SOFT control warns but allows posting
```
Test ID: FIN-BUD-010
Priority: P1
Description: Verify soft budget control produces warning, not error
Given:
  Budget: Rent account / Jan 2025 = 100,000 KES, control=SOFT
  Already spent: 80,000
When: PostTransaction with Rent expense entry of 25,000 (would exceed by 5,000)
Then:
  - Transaction IS posted (HTTP 200)
  - Response includes: warnings: ["BUDGET_SOFT_EXCEEDED"]
  - BudgetCheckResult: Available=20,000, Requested=25,000, WouldExceed=true
```
- [ ] **Status:** Pending — `budget_service_test.go:TestCheckBudget_SoftControl`

#### Test Case: CheckBudget — HARD control blocks posting
```
Test ID: FIN-BUD-011
Priority: P1
Description: Verify hard budget control prevents over-budget posting
Given:
  Budget: Rent / Jan 2025 = 100,000 KES, control=HARD
  Already spent: 80,000
When: PostTransaction with Rent expense entry of 25,000
Then:
  - Returns BUDGET_EXCEEDED error
  - HTTP 422
  - Transaction NOT posted
  - Account balance unchanged
```
- [ ] **Status:** Pending — `budget_service_test.go:TestCheckBudget_HardControl`

#### Test Case: CheckBudget — NONE control ignores budget
```
Test ID: FIN-BUD-012
Priority: P2
Description: Verify NONE control skips budget validation entirely
Given: Budget control = NONE, budget already 200% over-spent
When: PostTransaction with additional expense
Then:
  - Transaction posts without budget warnings
  - No BUDGET_EXCEEDED error
  - Budget control = NONE is a valid opt-out
```
- [ ] **Status:** Pending — `budget_service_test.go:TestCheckBudget_NoneControl`

#### Test Case: CheckBudget — within budget posts cleanly
```
Test ID: FIN-BUD-013
Priority: P1
Description: Verify no warnings when transaction is within budget
Given: Budget = 100,000, already spent = 60,000
When: PostTransaction with 30,000 expense (total = 90,000, under budget)
Then:
  - Transaction posts successfully
  - No budget warnings in response
  - BudgetCheckResult: WouldExceed=false
```
- [ ] **Status:** Pending — `budget_service_test.go:TestCheckBudget_WithinBudget`

---

### Budget vs Actual Tests

#### Test Case: GetBudgetVariance — returns correct variance per period
```
Test ID: FIN-BUD-020
Priority: P1
Description: Verify variance report matches budget lines against actual entries
Given:
  Budget line: Rent / Jan 2025 = 100,000
  Posted entries: 3 Rent transactions totaling 95,000 in Jan 2025
When: GetBudgetVsActual(accountCode="RENT", periodID=Jan2025)
Then:
  - BudgetedAmount = 100,000
  - ActualAmount   = 95,000
  - VarianceAmount = 5,000 (favorable)
  - VariancePct    = 5.00%
```
- [ ] **Status:** Pending — `budget_service_test.go:TestGetBudgetVariance_Correct`

---

## Cost Center Tests

#### Test Case: AllocateDistributed — splits cost per defined percentages
```
Test ID: FIN-CST-001
Priority: P1
Description: Verify distributed cost center allocation generates correct journal entries
File: internal/core/finance/service/cost_center_service_test.go
Given:
  IT dept is distributed: Sales 40%, Ops 30%, Admin 20%, R&D 10%
  Posted IT expenses for Jan 2025 = 500,000
When: AllocateDistributed(itCenterID, Jan2025Period)
Then:
  - 4 allocation journal entries created:
    Sales CC: 200,000 (40%)
    Ops CC:   150,000 (30%)
    Admin CC: 100,000 (20%)
    R&D CC:    50,000 (10%)
  - IT dept closing balance = 0 (fully allocated)
  - All 4 entries sum to exactly 500,000 (no rounding leakage)
```
- [ ] **Status:** Pending — `cost_center_service_test.go:TestAllocateDistributed_CorrectSplit`

#### Test Case: AllocateDistributed — allocation percentages must sum to 100%
```
Test ID: FIN-CST-002
Priority: P1
Description: Verify misconfigured allocation rules are caught
Given: Cost center allocations: Sales 40%, Ops 30% (total = 70% — missing 30%)
When: Calling AllocateDistributed
Then:
  - Returns ALLOCATION_PERCENTAGES_INVALID error
  - No allocation entries created
```
- [ ] **Status:** Pending — `cost_center_service_test.go:TestAllocateDistributed_PercentageMustSum100`

#### Test Case: CostCenterID required on expense entries when setting enabled
```
Test ID: FIN-CST-003
Priority: P1
Description: Verify tenant setting enforces cost center tagging
Given:
  - Tenant setting: cost_center_required_for_expenses = true
  - Expense entry (account RootType=EXPENSE) with no CostCenterID
When: PostTransaction
Then:
  - Returns COST_CENTER_REQUIRED error
  - Nothing posted
Given: Same expense entry WITH CostCenterID set
Then: Posts successfully
```
- [ ] **Status:** Pending — `cost_center_service_test.go:TestCostCenterRequired_ExpenseValidation`

---

## Repository Layer Tests

### Transaction Repository Stub Tests

#### Test Case: CreateEntry and GetEntriesByTransaction — round-trip
```
Test ID: FIN-REP-001
Priority: P0
Description: Verify entries can be written and retrieved (stubs wired to SQLC)
File: internal/core/finance/repository/transaction_test.go
Given: Transaction with 2 balanced entries
When:
  1. CreateEntry(debitEntry)
  2. CreateEntry(creditEntry)
  3. GetEntriesByTransaction(transactionID)
Then:
  - Returns exactly 2 entries
  - Amounts match what was inserted
  - AccountIDs match
  - LineNumbers are preserved in order
```
- [ ] **Status:** Pending — `transaction_repo_test.go:TestCreateAndRetrieveEntries`

#### Test Case: GetNextTransactionNumber — increments and is tenant-scoped
```
Test ID: FIN-REP-002
Priority: P1
Description: Verify sequence generation is tenant-isolated
Given: Tenant-A and Tenant-B
When:
  Tenant-A calls GetNextTransactionNumber("JE") → "JE-2025-00001"
  Tenant-A calls again → "JE-2025-00002"
  Tenant-B calls → "JE-2025-00001" (own sequence, not affected by A)
Then:
  - Each tenant has its own incrementing sequence
  - Numbers never collide across tenants
  - Within a tenant, numbers are monotonically increasing
```
- [ ] **Status:** Pending — `transaction_repo_test.go:TestGetNextTransactionNumber`

#### Test Case: IsTransactionNumberUnique — correctly detects duplicates
```
Test ID: FIN-REP-003
Priority: P1
Description: Verify uniqueness check is accurate
Given: Transaction "JE-001" exists for tenant
When: IsTransactionNumberUnique("JE-001") for same tenant
Then: Returns false (not unique)
When: IsTransactionNumberUnique("JE-002") for same tenant
Then: Returns true (unique)
When: IsTransactionNumberUnique("JE-001") for different tenant
Then: Returns true (scoped to calling tenant)
```
- [ ] **Status:** Pending — `transaction_repo_test.go:TestIsTransactionNumberUnique`

#### Test Case: GetByStatus — returns only matching status transactions
```
Test ID: FIN-REP-004
Priority: P1
Description: Verify status filter works in repository
Given: 3 DRAFT, 2 PENDING_APPROVAL, 1 POSTED transactions
When: GetByStatus(PENDING_APPROVAL, limit=10)
Then: Returns exactly 2 transactions, all in PENDING_APPROVAL status
```
- [ ] **Status:** Pending — `transaction_repo_test.go:TestGetByStatus`

#### Test Case: MarkAsReversed — updates IsReversed flag
```
Test ID: FIN-REP-005
Priority: P0
Description: Verify reversal flag is written to DB (prevents double-reversal)
Given: Posted transaction
When: MarkAsReversed(id, reversalID, "CORRECTION")
Then:
  - transaction.is_reversed = true in DB
  - transaction.reversed_by_transaction_id = reversalID
  - Subsequent MarkAsReversed call on same transaction returns error or is idempotent
```
- [ ] **Status:** Pending — `transaction_repo_test.go:TestMarkAsReversed`

---

### Account Balance Repository Tests

#### Test Case: CalculateAccountBalance — SUM from posted entries only
```
Test ID: FIN-REP-010
Priority: P0
Description: Verify balance calculation queries the entries table correctly
File: internal/core/finance/repository/accounts_test.go
Given: Account has:
  - Posted debit entry: 100,000
  - Posted debit entry: 50,000
  - Posted credit entry: 30,000
  - Draft debit entry: 20,000 (must be excluded)
When: CalculateAccountBalance(accountID)
Then:
  - DebitTotal  = 150,000
  - CreditTotal =  30,000
  - NetBalance  = 120,000 (debit normal balance)
  - Draft 20,000 is NOT included
```
- [ ] **Status:** Pending — `accounts_repo_test.go:TestCalculateAccountBalance`

#### Test Case: GetTrialBalance — all accounts with their net balances
```
Test ID: FIN-REP-011
Priority: P0
Description: Verify trial balance returns real data from entries table
Given: After posting:
  Dr Cash (1120)    100,000
  Cr Revenue (4100) 100,000
When: GetTrialBalance(entityID, asOfDate=nil)
Then:
  - Account 1120 (Cash): DebitTotal=100,000, CreditTotal=0
  - Account 4100 (Revenue): DebitTotal=0, CreditTotal=100,000
  - Total debits == total credits (trial balance balances)
  - No accounts with zero activity appear if not included by design
```
- [ ] **Status:** Pending — `accounts_repo_test.go:TestGetTrialBalance_MatchesEntries`

#### Test Case: HasTransactions — correctly detects presence of entries
```
Test ID: FIN-REP-012
Priority: P0
Description: Verify HasTransactions reflects actual entry existence
Given: Account with 1 posted entry
When: HasTransactions(accountID)
Then: Returns true
Given: Account with 0 entries
Then: Returns false
Note: If HasTransactions always returns false, DeleteAccount never blocks
      and accounts with history can be deleted, destroying the audit trail.
```
- [ ] **Status:** Pending — `accounts_repo_test.go:TestHasTransactions`

---

### Reversal History Repository Tests

#### Test Case: IsReversal — returns true for reversal transactions
```
Test ID: FIN-REP-020
Priority: P0
Description: Verify reversal history correctly identifies reversal transactions
File: internal/core/finance/repository/reversal_history_test.go
Given: Original=JE-001 reversed by JE-002 (row in finance_reversal_history)
When: IsReversal(JE-002)
Then: Returns true
When: IsReversal(JE-001)
Then: Returns false (JE-001 is the original, not a reversal)
When: IsReversal(JE-003) — unrelated transaction
Then: Returns false
```
- [ ] **Status:** Pending — `reversal_history_test.go:TestIsReversal`

#### Test Case: GetByOriginal — returns all reversals of an original
```
Test ID: FIN-REP-021
Priority: P0
Description: Verify reversal history is accurately queryable
Given: JE-001 was reversed once by JE-002
When: GetByOriginal(JE-001)
Then:
  - Returns 1 record
  - Record.OriginalTransactionID = JE-001
  - Record.ReversalTransactionID = JE-002
  - Record.InitiatedBy = correct user ID
  - Record.CreatedAt is a real timestamp (not a random UUID)
```
- [ ] **Status:** Pending — `reversal_history_test.go:TestGetByOriginal`

---

## Posting Engine Tests

#### Test Case: RecalculateBalances — computes from entries, not cache
```
Test ID: FIN-ENG-001
Priority: P0
Description: Verify RecalculateBalances reads from entries table, not stale cache
File: internal/core/finance/service/transaction_posting_engine_test.go
Given:
  Cached balance for account "1120" = 0 (stale)
  Actual sum of posted entries for "1120" = 150,000 (debit)
When: RecalculateBalances(ctx, accountID)
Then:
  - Reads SUM from finance_transaction_entries (not cache)
  - Writes 150,000 as the new balance
  - account.CurrentBalance = 150,000 in DB
```
- [ ] **Status:** Pending — `posting_engine_test.go:TestRecalculateBalances_FromEntries`

#### Test Case: BatchPostTransactions — aggregates balance updates
```
Test ID: FIN-ENG-002
Priority: P0
Description: Verify batch posting applies all balance movements, not just the last
Given:
  Transaction 1: Dr Cash 100,000 / Cr Revenue 100,000
  Transaction 2: Dr Cash 50,000 / Cr Revenue 50,000
  Both targeting the same Cash account (1120)
When: BatchPostTransactions([txn1, txn2])
Then:
  - Cash account balance increases by 150,000 total (not just 50,000)
  - Revenue account balance increases by 150,000 total
  - Both transactions status = POSTED
  - Balance deltas are aggregated before write (1 UpdateBalance call per account)
```
- [ ] **Status:** Pending — `posting_engine_test.go:TestBatchPostTransactions_AggregatesBalances`

#### Test Case: BatchPostTransactions — one failure rolls back entire batch
```
Test ID: FIN-ENG-003
Priority: P0
Description: Verify batch posting is all-or-nothing
Given: Batch of 3 transactions; transaction 2 is invalid (unbalanced)
When: BatchPostTransactions([txn1, txn2_invalid, txn3])
Then:
  - All 3 transactions remain in APPROVED status (not POSTED)
  - No balance updates are applied
  - Error indicates which transaction failed and why
```
- [ ] **Status:** Pending — `posting_engine_test.go:TestBatchPostTransactions_AtomicRollback`

---

## Approval Workflow Engine Tests

#### Test Case: ValidateApprovalAuthority — requires correct IAM permission
```
Test ID: FIN-WRK-001
Priority: P1
Description: Verify only authorized approvers can approve transactions
File: internal/core/finance/service/transaction_workflow_engine_test.go
Given: User without "finance.transactions.approve" permission
When: User calls ApproveTransaction via the workflow engine
Then:
  - ValidateApprovalAuthority returns false
  - Returns PERMISSION_DENIED error
  - Transaction not approved
```
- [ ] **Status:** Pending — `workflow_engine_test.go:TestValidateApprovalAuthority_RequiresPermission`

#### Test Case: Approval history is persisted after approval decision
```
Test ID: FIN-WRK-002
Priority: P2
Description: Verify approval decisions are recorded for audit trail
Given: Transaction in PENDING_APPROVAL
When: Finance manager calls ApproveTransaction
Then:
  - Row inserted in finance_approval_history
  - Row contains: transaction_id, tier, action=APPROVED, performed_by, created_at
  - GetApprovalHistory(transactionID) returns this record
```
- [ ] **Status:** Pending — `workflow_engine_test.go:TestApprovalHistory_Persisted`

#### Test Case: Approval workflow state survives service restart
```
Test ID: FIN-WRK-003
Priority: P2
Description: Verify in-flight approval state is durable
Given: Transaction submitted for approval; service restarts
When: Service restarts and GetPendingApprovals is called
Then:
  - Transaction still appears in pending list
  - Approval state loaded from DB (not lost on restart)
```
- [ ] **Status:** Pending — `workflow_engine_test.go:TestApprovalWorkflow_SurvivesRestart`

#### Test Case: GetPendingApproval — returns only transactions awaiting action
```
Test ID: FIN-WRK-004
Priority: P1
Description: Verify pending approval query excludes already-decided transactions
Given:
  - JE-001: PENDING_APPROVAL (not yet actioned)
  - JE-002: APPROVED (already actioned)
  - JE-003: REJECTED (already actioned)
When: GetPendingApproval(entityID, limit=10)
Then:
  - Returns only JE-001
  - JE-002 and JE-003 are excluded
```
- [ ] **Status:** Pending — `workflow_engine_test.go:TestGetPendingApproval_FilterCorrect`

---

## Transaction Numbering Tests

#### Test Case: ReserveTransactionNumber — uses authenticated user ID
```
Test ID: FIN-NUM-001
Priority: P1
Description: Verify reservation records the actual user, not a random UUID
File: internal/core/finance/service/transaction_numbering_service_test.go
Given:
  ctx = context with userID = "tenant:usr_amina_hassan"
When: ReserveTransactionNumber(ctx, entityID, "JE")
Then:
  - reservation.ReservedBy = "tenant:usr_amina_hassan"
  - NOT a random UUID
  - Audit trail correctly traces who reserved the number
```
- [ ] **Status:** Pending — `numbering_service_test.go:TestReserveNumber_CorrectUser`

#### Test Case: ReserveTransactionNumber — system user used when no user in context
```
Test ID: FIN-NUM-002
Priority: P1
Description: Verify system-generated reservations use a designated system user
Given: Context has no user ID (background job generating recurring transaction)
When: ReserveTransactionNumber(ctx, entityID, "JE")
Then:
  - reservation.ReservedBy = SystemUserID (from settings constants)
  - Not a random UUID
  - Not nil / empty UUID
```
- [ ] **Status:** Pending — `numbering_service_test.go:TestReserveNumber_SystemUser`

---

## API Handler Tests

#### Test Case: POST /accounts — validates required fields
```
Test ID: FIN-API-001
Priority: P1
Description: Verify handler returns 400 for missing required fields
File: internal/api/handlers/finance/handler_test.go
Given: POST /accounts body missing AccountCode
When: Request is sent
Then:
  - HTTP 400 Bad Request
  - Body: {error: {code: "VALIDATION_ERROR", fields: [{field: "account_code", message: "required"}]}}
  - No account created
```
- [ ] **Status:** Pending — `handler_test.go:TestCreateAccount_MissingFields`

#### Test Case: POST /transactions/:id/post — returns 422 for PERIOD_CLOSED
```
Test ID: FIN-API-002
Priority: P0
Description: Verify period closed error maps to correct HTTP status
Given: Transaction dated in a closed period
When: POST /transactions/{id}/post
Then:
  - HTTP 422 Unprocessable Entity
  - Body: {error: {code: "PERIOD_CLOSED", message: "..."}}
  - NOT a 500 Internal Server Error
```
- [ ] **Status:** Pending — `handler_test.go:TestPostTransaction_PeriodClosed_422`

#### Test Case: POST /transactions/:id/post — returns 200 with warnings for soft budget
```
Test ID: FIN-API-003
Priority: P1
Description: Verify soft budget control produces 200 with warnings, not an error response
Given: Transaction exceeds soft budget control
When: POST /transactions/{id}/post
Then:
  - HTTP 200 OK
  - Body: {data: {...}, warnings: ["BUDGET_SOFT_EXCEEDED"]}
  - Transaction IS posted
  - "warnings" field is present alongside "data"
```
- [ ] **Status:** Pending — `handler_test.go:TestPostTransaction_SoftBudget_200WithWarning`

#### Test Case: GET /accounts/:id — returns 404 for unknown account
```
Test ID: FIN-API-004
Priority: P1
Description: Verify not-found accounts return 404, not 500
Given: Non-existent account UUID
When: GET /accounts/{uuid}
Then:
  - HTTP 404 Not Found
  - Body: {error: {code: "ACCOUNT_NOT_FOUND"}}
```
- [ ] **Status:** Pending — `handler_test.go:TestGetAccount_NotFound`

#### Test Case: GET /accounts — pagination params respected
```
Test ID: FIN-API-005
Priority: P2
Description: Verify list endpoint respects pagination query params
Given: 30 accounts exist
When: GET /accounts?page=2&per_page=10
Then:
  - HTTP 200
  - Body contains 10 accounts (second page)
  - Body includes meta: {total: 30, page: 2, per_page: 10, total_pages: 3}
```
- [ ] **Status:** Pending — `handler_test.go:TestListAccounts_Pagination`

#### Test Case: POST /transactions — errors and warnings are clearly separated
```
Test ID: FIN-API-006
Priority: P2
Description: Verify API response structure separates hard errors from soft warnings
Given: Response with mixed validation results
Then:
  - Hard errors (ERROR severity) appear in response.errors[]
  - Soft warnings (WARNING severity) appear in response.warnings[]
  - If errors[] is non-empty, HTTP status is 4xx
  - If only warnings[], HTTP status is 200
```
- [ ] **Status:** Pending — `handler_test.go:TestApiResponse_ErrorWarningStructure`

---

## Double-Entry Integrity Tests

These tests verify the fundamental accounting invariant: total debits always equal total credits.

#### Test Case: Trial balance always nets to zero
```
Test ID: FIN-DEI-001
Priority: P0
Description: Verify the trial balance sum of debits equals sum of credits at all times
File: internal/core/finance/repository/accounts_test.go
Given: Any combination of posted transactions
When: GetTrialBalance
Then:
  - SUM(all account DebitTotals) == SUM(all account CreditTotals)
  - If they do not balance, there is a data integrity bug in the posting engine
Note: This is the fundamental double-entry invariant. It must hold for every
      state of the ledger — no exceptions.
```
- [ ] **Status:** Pending — `accounts_repo_test.go:TestTrialBalance_AlwaysNets`

#### Test Case: Each transaction's entries are balanced before persisting
```
Test ID: FIN-DEI-002
Priority: P0
Description: Verify IsBalanced() is checked before any transaction is committed
Given: Attempt to persist an unbalanced transaction (total debits ≠ total credits)
When: PostTransaction is called
Then:
  - Returns UNBALANCED_TRANSACTION before any DB write
  - finance_transaction_entries table is NOT modified
  - finance_transactions status is NOT changed
```
- [x] **Status:** Done — covered by `service/transaction_service_test.go:TestPostTransaction_Unbalanced`

#### Test Case: Decimal precision — no rounding errors accumulate
```
Test ID: FIN-DEI-003
Priority: P0
Description: Verify financial amounts use decimal arithmetic, not floating point
Given: Transaction with amount 1,000,000.01 KES (fractional cent)
When: Processing the transaction
Then:
  - Amount is stored exactly (no floating-point precision loss)
  - 100 × 0.10 == 10.00 exactly (not 9.999999...)
  - All calculations use decimal.Decimal, not float64
Note: Using float64 for financial calculations is a critical bug that causes
      cents to accumulate into material discrepancies.
```
- [x] **Status:** Done — `domain/validation_test.go:TestDecimalPrecision_NoFloatError`

#### Test Case: Foreign currency balance = sum of base-currency equivalents
```
Test ID: FIN-DEI-004
Priority: P1
Description: Verify FC account balances track base-currency equivalent
Given: 3 USD entries with rates 128, 130, 129 applied
When: GetAccountBalance for the USD account
Then:
  - FCBalance = sum of original USD amounts
  - BaseBalance = sum of each (USD × rate_at_posting_date)
  - BaseBalance ≠ FCBalance × current_rate (historical rates, not spot)
```
- [ ] **Status:** Pending — `account_service_test.go:TestFCAccountBalance_HistoricalRates`

---

## Integration Tests

### Full Financial Lifecycle Tests

#### Test Case: Complete journal entry lifecycle
```
Test ID: FIN-INT-001
Priority: P1
Description: End-to-end: create → approve → post → verify GL
File: internal/core/finance/integration_test.go
Given: Fresh tenant with open period and loaded chart of accounts
When:
  1. CreateAccount: "1120 Cash", "4100 Revenue" (if not seeded)
  2. CreateTransaction: Dr Cash 50,000 / Cr Revenue 50,000
  3. SubmitForApproval: status → PENDING_APPROVAL
  4. ApproveTransaction (different user): status → APPROVED
  5. PostTransaction: status → POSTED
Then:
  - Steps 1-5 all succeed
  - Cash account DebitTotal = 50,000
  - Revenue account CreditTotal = 50,000
  - GetTrialBalance balances (total debits == total credits)
  - Audit log has entries for each state transition
```
- [ ] **Status:** Pending — `integration_test.go:TestFullJournalEntryLifecycle`

#### Test Case: Month-end close sequence
```
Test ID: FIN-INT-002
Priority: P1
Description: End-to-end: complete month close workflow
Given: Open period with posted transactions
When:
  1. All transactions reviewed and posted
  2. Bank reconciliation approved
  3. SoftClosePeriod(periodID)
  4. Finance manager posts adjusting entries
  5. HardClosePeriod(periodID)
Then:
  - After soft close: regular users cannot post; finance role can
  - After hard close: no one can post to this period
  - Attempting to post after hard close → PERIOD_CLOSED error
```
- [ ] **Status:** Pending — `integration_test.go:TestMonthEndClose_Sequence`

#### Test Case: Bank reconciliation blocks hard close
```
Test ID: FIN-INT-003
Priority: P2
Description: Verify hard close requires approved bank reconciliation
Given: Cash account has unreconciled transactions
When: Attempting HardClosePeriod
Then:
  - Returns BANK_ACCOUNT_NOT_RECONCILED error
  - Period remains SOFT_CLOSED
  - Error message identifies which bank account is unreconciled
```
- [ ] **Status:** Pending — `integration_test.go:TestHardClose_BlockedByUnreconciledAccount`

#### Test Case: Expiring auditor access is revoked at next GL operation
```
Test ID: FIN-INT-004
Priority: P2
Description: Verify time-limited roles are enforced at next authorization check
Given:
  - Auditor role assigned with expiry = 5ms from now
  - Auditor has "finance.transactions.read" permission
When:
  1. Call ListTransactions immediately → succeeds (role not yet expired)
  2. Wait for expiry to pass (> 5ms)
  3. Call ListTransactions again
Then:
  - Step 3: role is lazily revoked → returns PERMISSION_DENIED
  - role_assignments.is_active = false in DB
```
- [ ] **Status:** Pending — `integration_test.go:TestAuditorAccess_ExpiresAtNextCheck`

#### Test Case: Budget revision creates a new version
```
Test ID: FIN-INT-005
Priority: P2
Description: Verify budget revision workflow creates versioned budget
Given: Approved budget v1
When: RevokeBudget / create revised budget
Then:
  - New budget created with OriginalBudgetID pointing to v1
  - New budget has version = 2
  - Both v1 and v2 are queryable
  - Only the latest approved version is used for budget controls
```
- [ ] **Status:** Pending — `integration_test.go:TestBudgetRevision_NewVersion`

#### Test Case: Multi-instance cache sync via external invalidation
```
Test ID: FIN-INT-010
Priority: P2
Description: Verify two service instances share consistent state after cache invalidation
Given: Two FinanceHandler instances sharing the same DB
When: Instance A creates a new account
And: Instance B calls GetAccountByCode before its cache is refreshed
Then:
  - Before invalidation: Instance B may get cache miss (account not found)
  - After InvalidateCache on Instance B: account is visible
  - Both instances converge to consistent state
```
- [ ] **Status:** Pending — `integration_test.go:TestMultiInstance_CacheSync`

---

## Performance Tests

#### Test Case: PostTransaction — single transaction < 50ms p99
```
Test ID: FIN-PERF-001
Priority: P2
Description: Verify single transaction posting meets latency SLA
File: internal/core/finance/bench_test.go
Given: Connected DB, open period, pre-loaded accounts
When: Calling PostTransaction 1,000 times in sequence
Then:
  - p50 < 10ms
  - p99 < 50ms
  - No goroutine leaks
  - No DB connection exhaustion
```
- [ ] **Status:** Pending — `bench_test.go:BenchmarkPostTransaction`

#### Test Case: BatchPostTransactions — 100 transactions < 500ms
```
Test ID: FIN-PERF-002
Priority: P2
Description: Verify batch posting is significantly faster than 100 sequential posts
Given: 100 balanced transactions pre-created in APPROVED status
When: BatchPostTransactions([100 transactions])
Then:
  - Total time < 500ms
  - Faster than 100 × p99_single (proves batching benefit)
  - All 100 transactions POSTED
  - Balance updates aggregated (not 100 × N UpdateBalance calls)
```
- [ ] **Status:** Pending — `bench_test.go:BenchmarkBatchPostTransactions`

#### Test Case: GetTrialBalance — 10,000 entries in < 200ms
```
Test ID: FIN-PERF-003
Priority: P2
Description: Verify trial balance query scales to real-world data volumes
Given: 10,000 posted transaction entries across 200 accounts
When: GetTrialBalance(entityID)
Then:
  - Returns in < 200ms
  - All accounts with balances are included
  - Uses GROUP BY query with index scan (not sequential scan)
```
- [ ] **Status:** Pending — `bench_test.go:TestTrialBalance_10kEntries`

#### Test Case: ListAccounts with filter — < 10ms for 1,000 accounts
```
Test ID: FIN-PERF-004
Priority: P2
Description: Verify account listing is fast even with large COA
Given: 1,000 accounts for a single tenant
When: ListAccounts(filter{RootType: ASSET, IsActive: true, Page: 1, PerPage: 50})
Then:
  - Returns in < 10ms
  - Uses index on (tenant_id, root_type, is_active)
  - No full table scan
```
- [ ] **Status:** Pending — `bench_test.go:BenchmarkListAccounts_1kAccounts`

#### Test Case: Exchange rate lookup — in-memory hit < 1ms
```
Test ID: FIN-PERF-005
Priority: P3
Description: Verify cached rate lookup is near-instant
Given: USD/KES rate loaded in cache
When: GetExchangeRate("USD", "KES", today, SPOT) called 10,000 times
Then:
  - p99 < 1ms per call (served from cache, no DB round trip)
  - Cache hit rate = 100% for the same currency pair + date
```
- [ ] **Status:** Pending — `bench_test.go:BenchmarkExchangeRateLookup_CacheHit`

#### Test Case: Budget variance query — < 50ms for 12-period budget
```
Test ID: FIN-PERF-006
Priority: P2
Description: Verify budget vs actual calculation is responsive
Given: Annual budget with 12 periods and 50 line items, all periods have actuals
When: GetBudgetVsActual(fiscalYearID)
Then:
  - Returns all 12 × 50 = 600 variance rows in < 50ms
  - Uses appropriate join and aggregation indexes
```
- [ ] **Status:** Pending — `bench_test.go:BenchmarkBudgetVariance`

---

## Security & Tenant Isolation Tests

#### Test Case: Account query never returns other tenant's accounts
```
Test ID: FIN-SEC-001
Priority: P0
Description: Verify strict tenant isolation on all account queries
File: internal/core/finance/security_test.go
Given: Tenant-A has 10 accounts; Tenant-B has 5 accounts
When: Tenant-A calls ListAccounts (any filter)
Then:
  - Returns at most 10 accounts
  - None belong to Tenant-B
  - Even with filter{IsActive: false}, no Tenant-B accounts appear
```
- [ ] **Status:** Pending — `security_test.go:TestTenantIsolation_ListAccounts`

#### Test Case: GetAccountByID — cannot fetch another tenant's account by UUID
```
Test ID: FIN-SEC-002
Priority: P0
Description: Verify direct ID lookup does not bypass tenant isolation
Given: Tenant-A's account UUID is known to a Tenant-B user (e.g., from a leak)
When: Tenant-B calls GetAccountByID(tenantA_accountID)
Then:
  - Returns ErrAccountNotFound (not the account)
  - tenant_id check prevents cross-tenant read
  - HTTP 404 (not 403, which would confirm existence)
```
- [ ] **Status:** Pending — `security_test.go:TestTenantIsolation_GetByID`

#### Test Case: PostTransaction — cannot reference another tenant's account
```
Test ID: FIN-SEC-003
Priority: P0
Description: Verify transaction entries cannot reference cross-tenant accounts
Given: Entry referencing an account UUID that belongs to Tenant-B
When: Tenant-A calls PostTransaction
Then:
  - Returns ACCOUNT_NOT_FOUND error (account appears not to exist)
  - Transaction is not posted
  - No data leakage about Tenant-B's account structure
```
- [ ] **Status:** Pending — `security_test.go:TestTenantIsolation_CrossTenantAccountInEntry`

#### Test Case: Trial balance only includes calling tenant's entries
```
Test ID: FIN-SEC-004
Priority: P0
Description: Verify trial balance is tenant-scoped
Given: Tenant-A and Tenant-B both have posted entries
When: Tenant-A calls GetTrialBalance
Then:
  - Only Tenant-A's entries contribute to the balance
  - Tenant-B's debit/credit totals are zero from Tenant-A's perspective
```
- [ ] **Status:** Pending — `security_test.go:TestTenantIsolation_TrialBalance`

#### Test Case: Fiscal year operations are entity/tenant scoped
```
Test ID: FIN-SEC-005
Priority: P1
Description: Verify period management is isolated per tenant
Given: Tenant-A closes Jan 2025 period
When: Tenant-B attempts to post a Jan 2025 transaction
Then:
  - Succeeds — Tenant-A's period close does NOT affect Tenant-B
  - Each tenant manages its own independent period lifecycle
```
- [ ] **Status:** Pending — `security_test.go:TestTenantIsolation_PeriodIndependence`

#### Test Case: Budget controls apply per-tenant configuration
```
Test ID: FIN-SEC-006
Priority: P1
Description: Verify budget control type is tenant-specific
Given:
  - Tenant-A has HARD budget control on Rent
  - Tenant-B has NONE budget control on same account code
When: Both tenants try to post over-budget Rent expense
Then:
  - Tenant-A: BUDGET_EXCEEDED error (hard control)
  - Tenant-B: Posts successfully (no control)
  - Controls are configuration, not schema — no cross-tenant bleed
```
- [ ] **Status:** Pending — `security_test.go:TestBudgetControl_TenantIsolated`

#### Test Case: SQL injection in account code field does not affect queries
```
Test ID: FIN-SEC-010
Priority: P0
Description: Verify parameterized queries prevent injection via account code
Given: AccountCode = "1120'; DROP TABLE finance_transactions; --"
When: CreateAccount or GetAccountByCode is called with this code
Then:
  - finance_transactions table is NOT affected
  - The literal string is stored/queried as a data value
  - All queries use parameterized statements, not string interpolation
```
- [ ] **Status:** Pending — `security_test.go:TestSQLInjection_AccountCode`

#### Test Case: Closed period cannot be reopened without elevated permission
```
Test ID: FIN-SEC-011
Priority: P1
Description: Verify period reopening requires CFO-level permission
Given: Accounting period hard-closed
When: Regular finance user calls ReopenPeriod
Then:
  - Returns PERMISSION_DENIED error
  - Period remains HARD_CLOSED
When: CFO user (with finance.periods.lock permission) calls ReopenPeriod
Then:
  - Succeeds (with audit log entry recording the reopen action)
```
- [ ] **Status:** Pending — `security_test.go:TestPeriodReopen_RequiresElevatedPermission`

#### Test Case: Exchange rate manipulation — rate below zero rejected
```
Test ID: FIN-SEC-012
Priority: P1
Description: Verify malformed rates cannot corrupt transaction amounts
Given: Attempt to create ExchangeRate with rate = -1.00 or rate = 0
When: Calling CreateExchangeRate or UpdateExchangeRate
Then:
  - Returns validation error: INVALID_EXCHANGE_RATE
  - Rate must be > 0 (check constraint + service validation)
  - Negative rates would flip debit/credit semantics
```
- [ ] **Status:** Pending — `security_test.go:TestExchangeRate_NegativeRateRejected`

---

## Test Infrastructure Notes

### Test Database Setup
```go
// Recommended pattern for integration tests
func setupTestDB(t *testing.T) *pgxpool.Pool {
    // Use testcontainers-go with the project's postgres image
    // Run migrations before each test suite
    // Seed tenant + entity with uuid.New() for isolation
    // Defer cleanup to t.Cleanup()
}
```

### Decimal Arithmetic Requirement
```go
// ALL financial amounts must use shopspring/decimal, not float64
// Wrong:
amount := 0.1 + 0.2  // = 0.30000000000000004

// Correct:
amount := decimal.NewFromFloat(0.1).Add(decimal.NewFromFloat(0.2))
// = 0.3 exactly
```

### Tenant Context Injection
```go
// Tests that exercise tenant isolation must inject tenant ID via context:
ctx := shared.WithTenantID(context.Background(), tenantUUID)
ctx  = shared.WithUserID(ctx, userUUID)
// All service and repository calls use this pattern
```

### Test ID Namespace Reference

| Prefix | Section |
|--------|---------|
| FIN-TYP | Domain types & enums |
| FIN-ACC | Account service |
| FIN-TXN | Transaction service |
| FIN-PER | Period management |
| FIN-CCY | Currency & exchange rates |
| FIN-BUD | Budget service |
| FIN-CST | Cost center |
| FIN-REP | Repository layer |
| FIN-ENG | Posting engine |
| FIN-WRK | Workflow / approval engine |
| FIN-NUM | Transaction numbering |
| FIN-API | API handler |
| FIN-DEI | Double-entry integrity |
| FIN-INT | Integration tests |
| FIN-PERF | Performance benchmarks |
| FIN-SEC | Security & tenant isolation |

---

[Back to Index](README.md)
