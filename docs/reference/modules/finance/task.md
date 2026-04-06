# Finance Module — Implementation Task List

> Derived from gap analysis between `PRD.md` (business specification) and `internal/core/finance/` (current codebase).
> Each task is self-contained: problem → solution → expected outcome → test plan.
> Tasks are ordered by dependency — complete earlier tasks before later ones.

---

## Legend

- `[ ]` Not started
- `[~]` In progress
- `[x]` Complete
- **P0** — Blocks other work / correctness bug
- **P1** — Core module feature, needed for MVP
- **P2** — Important but not blocking
- **P3** — Enhancement / polish

---

## P0 — Correctness Bugs (Fix First)

---

### TASK-001 `[ ]` Fix `TransactionStatus` editability contradiction

**Priority:** P0
**File:** `internal/core/finance/domain/types.go:181`

**Problem:**
`TransactionStatus.IsEditable()` returns `true` for both `DRAFT` and `PENDING_APPROVAL`. The PRD (section 6, Transaction Lifecycle) explicitly states `PENDING_APPROVAL` is **not editable** — it is frozen awaiting an approver's decision. Allowing edits to a submitted-for-approval entry undermines the approval control entirely.

```go
// Current — WRONG
func (ts TransactionStatus) IsEditable() bool {
    switch ts {
    case TransactionStatusDraft, TransactionStatusPendingApproval: // ← PENDING_APPROVAL should not be here
        return true
    ...
```

**Solution:**
Remove `TransactionStatusPendingApproval` from `IsEditable()`. Only `DRAFT` and `REJECTED` (after rejection, entry can be corrected and resubmitted) should be editable.

```go
func (ts TransactionStatus) IsEditable() bool {
    switch ts {
    case TransactionStatusDraft, TransactionStatusRejected:
        return true
    default:
        return false
    }
}
```

**Expected Outcome:**
- A transaction in `PENDING_APPROVAL` status returns `IsEditable() == false`
- Service layer rejects update attempts on `PENDING_APPROVAL` transactions with `TRANSACTION_NOT_EDITABLE` error
- `REJECTED` transactions can be corrected and resubmitted

**How to Test:**
```go
// Unit test in domain/types_test.go
func TestTransactionStatusEditability(t *testing.T) {
    assert.True(t,  TransactionStatusDraft.IsEditable())
    assert.True(t,  TransactionStatusRejected.IsEditable())
    assert.False(t, TransactionStatusPendingApproval.IsEditable()) // was true — must be false
    assert.False(t, TransactionStatusApproved.IsEditable())
    assert.False(t, TransactionStatusPosted.IsEditable())
    assert.False(t, TransactionStatusCancelled.IsEditable())
    assert.False(t, TransactionStatusReversed.IsEditable())
}
```

---

### TASK-002 `[ ]` Remove payment-processor codes from `RejectionReason`

**Priority:** P0
**File:** `internal/core/finance/domain/types.go:239`

**Problem:**
`RejectionReason` contains payment-gateway codes (`EXPIRED_CARD`, `INVALID_MERCHANT`, `FRAUD_SUSPECTED`, `DAILY_LIMIT_EXCEEDED`, `INSUFFICIENT_FUNDS`) that make no sense for a GL journal entry approval rejection. These codes will appear in the approval rejection UI for accountants rejecting a journal entry, which is misleading and unprofessional. They also suggest the enum was copied from a payments module without cleanup.

**Solution:**
Replace with accounting-specific rejection reasons that reflect why a Finance Manager would reject a journal entry:

```go
const (
    RejectionReasonInsufficientSupportingDoc RejectionReason = "INSUFFICIENT_SUPPORTING_DOCUMENTATION"
    RejectionReasonIncorrectAccount          RejectionReason = "INCORRECT_ACCOUNT_CODE"
    RejectionReasonPeriodClosed              RejectionReason = "ACCOUNTING_PERIOD_CLOSED"
    RejectionReasonAmountMismatch            RejectionReason = "AMOUNT_MISMATCH_WITH_SOURCE"
    RejectionReasonDuplicateEntry            RejectionReason = "DUPLICATE_ENTRY"
    RejectionReasonPolicyViolation           RejectionReason = "POLICY_VIOLATION"
    RejectionReasonBudgetExceeded            RejectionReason = "BUDGET_EXCEEDED"
    RejectionReasonUnauthorisedAccount       RejectionReason = "UNAUTHORISED_ACCOUNT_ACCESS"
    RejectionReasonOther                     RejectionReason = "OTHER"
)
```

**Expected Outcome:**
- Finance staff rejecting a journal entry see contextually relevant rejection reasons
- No payment-gateway terminology in the accounting UI
- Existing `ValidReasons` slice updated to match

**How to Test:**
```go
// Unit test
func TestRejectionReasonValidity(t *testing.T) {
    // Payment codes must NOT be valid
    assert.False(t, RejectionReason("EXPIRED_CARD").IsValid())
    assert.False(t, RejectionReason("INVALID_MERCHANT").IsValid())
    assert.False(t, RejectionReason("FRAUD_SUSPECTED").IsValid())

    // Accounting codes must be valid
    assert.True(t, RejectionReasonIncorrectAccount.IsValid())
    assert.True(t, RejectionReasonDuplicateEntry.IsValid())
    assert.True(t, RejectionReasonBudgetExceeded.IsValid())
}
```

---

### TASK-003 `[ ]` Resolve duplicate `TransactionType` values

**Priority:** P0
**File:** `internal/core/finance/domain/types.go:108`

**Problem:**
`TransactionType` has both `JOURNAL` and `JOURNAL_ENTRY` as distinct constants, but they appear to mean the same thing. This creates ambiguity: which one should be used when creating a manual journal entry? If both exist in the database, reports and filters will silently miss half the data.

```go
TransactionTypeJournal      TransactionType = "JOURNAL"        // ← duplicate intent
TransactionTypeJournalEntry TransactionType = "JOURNAL_ENTRY"  // ← duplicate intent
```

**Solution:**
1. Determine canonical value (prefer `JOURNAL_ENTRY` — more explicit)
2. Write a migration to update existing rows: `UPDATE transactions SET transaction_type = 'JOURNAL_ENTRY' WHERE transaction_type = 'JOURNAL'`
3. Remove `TransactionTypeJournal` from the enum
4. Update all code references

**Expected Outcome:**
- Single canonical type `JOURNAL_ENTRY` for manually created journal entries
- No ambiguity in filters, reports, or service logic
- Migration converts any legacy `JOURNAL` rows

**How to Test:**
- Grep confirms zero uses of `TransactionTypeJournal` after cleanup
- DB migration runs cleanly on test schema
- `ParseTransactionType("JOURNAL")` returns an error (no longer valid)
- `ParseTransactionType("JOURNAL_ENTRY")` succeeds

---

## P1 — Core Domain Models (Must Build for MVP)

---

### TASK-004 `[ ]` Create `FiscalYear` and `AccountingPeriod` domain models

**Priority:** P1
**File:** `internal/core/finance/domain/period.go` (new file)

**Problem:**
The PRD's "Financial Period Management" section is fully documented but there is zero code for it. `Transaction.TransactionDate` exists but nothing validates it against an open period. Without period management, the system cannot prevent posting to closed months, cannot enforce month-end close, and cannot support year-end closing entries.

**Solution:**
Create `domain/period.go` with:

```go
type FiscalYear struct {
    ID          uuid.UUID
    TenantID    uuid.UUID
    Name        string        // "FY2025", "2025"
    StartDate   time.Time
    EndDate     time.Time
    IsClosed    bool
    ClosedAt    *time.Time
    ClosedBy    *uuid.UUID
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type PeriodStatus string
const (
    PeriodStatusOpen       PeriodStatus = "OPEN"
    PeriodStatusSoftClosed PeriodStatus = "SOFT_CLOSED"
    PeriodStatusHardClosed PeriodStatus = "HARD_CLOSED"
    PeriodStatusLocked     PeriodStatus = "LOCKED"
)

type AccountingPeriod struct {
    ID           uuid.UUID
    TenantID     uuid.UUID
    FiscalYearID uuid.UUID
    Name         string       // "January 2025"
    PeriodNumber int          // 1–12
    StartDate    time.Time
    EndDate      time.Time
    Status       PeriodStatus
    ClosedAt     *time.Time
    ClosedBy     *uuid.UUID
    LockedAt     *time.Time
    LockedBy     *uuid.UUID
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

func (p *AccountingPeriod) CanPost() bool {
    return p.Status == PeriodStatusOpen
}

func (p *AccountingPeriod) CanFinancePost() bool {
    return p.Status == PeriodStatusOpen || p.Status == PeriodStatusSoftClosed
}
```

Add `PeriodRepository` interface and `PeriodService` with: `GetOpenPeriodForDate`, `ClosePeriod`, `ReopenPeriod`, `LockPeriod`.

Wire period validation into `TransactionService.PostTransaction` — reject if period is not open.

**Expected Outcome:**
- Posting a transaction to a date in a closed period returns `PERIOD_CLOSED` error
- `SOFT_CLOSED` allows finance-role users to post but blocks others
- `LOCKED` blocks everyone including CFO
- Fiscal year close generates closing entries (Revenue/Expense → Retained Earnings)

**How to Test:**
```go
func TestPeriodValidation(t *testing.T) {
    // Open period — post succeeds
    open := &AccountingPeriod{Status: PeriodStatusOpen}
    assert.True(t, open.CanPost())

    // Hard closed — post blocked
    closed := &AccountingPeriod{Status: PeriodStatusHardClosed}
    assert.False(t, closed.CanPost())
    assert.False(t, closed.CanFinancePost())

    // Soft closed — blocked for regular users, open for finance
    soft := &AccountingPeriod{Status: PeriodStatusSoftClosed}
    assert.False(t, soft.CanPost())
    assert.True(t, soft.CanFinancePost())
}

// Integration: attempt to post transaction to closed period via service
func TestPostToClosed_ReturnsError(t *testing.T) {
    // seed closed period for Jan 2025
    // create transaction dated Jan 15 2025
    // call PostTransaction → expect PERIOD_CLOSED error
}
```

---

### TASK-005 `[ ]` Create `Currency` and `ExchangeRate` domain models

**Priority:** P1
**File:** `internal/core/finance/domain/currency.go` (new file)

**Problem:**
`Transaction` has `CurrencyCode` and `ExchangeRate` fields, but there is no `Currency` entity, no `ExchangeRate` table, and no service to load/validate rates. Multi-currency transactions cannot be processed correctly without knowing the rate at transaction date. The PRD's entire "Multi-Currency Operations" section is unsupported.

**Solution:**
Create `domain/currency.go`:

```go
type Currency struct {
    Code          string  // ISO 4217: "USD", "KES", "EUR"
    Name          string  // "US Dollar"
    Symbol        string  // "$"
    DecimalPlaces int     // 2
    IsActive      bool
    IsBaseCurrency bool   // only one per tenant
}

type RateType string
const (
    RateTypeSpot      RateType = "SPOT"
    RateTypeAverage   RateType = "AVERAGE"
    RateTypeHistorical RateType = "HISTORICAL"
    RateTypeBudget    RateType = "BUDGET"
)

type ExchangeRate struct {
    ID           uuid.UUID
    TenantID     uuid.UUID
    FromCurrency string
    ToCurrency   string
    Rate         decimal.Decimal  // units of ToCurrency per 1 FromCurrency
    RateType     RateType
    EffectiveDate time.Time
    Source       string           // "MANUAL", "CBK", "API"
    LoadedBy     uuid.UUID
    CreatedAt    time.Time
}
```

Add `CurrencyRepository` and `ExchangeRateRepository` interfaces.
Add `CurrencyService` with: `GetRate(from, to, date, rateType)`, `LoadRates([]ExchangeRate)`, `RevalueOpenBalances(periodID)`.

Wire into `TransactionService.PostTransaction`: if `CurrencyCode != baseCurrency`, require a rate for `TransactionDate` and compute `base_amount` on each entry.

**Expected Outcome:**
- Posting a USD transaction without a loaded rate returns `EXCHANGE_RATE_NOT_FOUND` error
- Each `TransactionEntry` stores both FC amount and KES equivalent
- Month-end revaluation job updates unrealized FX gain/loss entries
- `GetRate` returns correct rate for the exact date or nearest prior date

**How to Test:**
```go
func TestExchangeRateRetrieval(t *testing.T) {
    // Load rate: USD/KES 128.00 on Jan 10
    // GetRate("USD", "KES", Jan10, SPOT) → 128.00 ✓
    // GetRate("USD", "KES", Jan11, SPOT) → 128.00 (nearest prior) ✓
    // GetRate("USD", "KES", Jan09, SPOT) → error (no rate before Jan 10)
}

func TestFCTransactionPosting(t *testing.T) {
    // Load rate USD/KES 128.00
    // Post transaction: $10,000 USD
    // Entry base_amount = 10,000 × 128.00 = 1,280,000 KES ✓
    // Without rate loaded → EXCHANGE_RATE_NOT_FOUND error ✓
}
```

---

### TASK-006 `[ ]` Create `CostCenter` domain model and wire into transaction entries

**Priority:** P1
**File:** `internal/core/finance/domain/costcenter.go` (new file)

**Problem:**
The PRD documents full cost center management (hierarchy, distributed allocation, budget tracking). `TransactionEntry` has no `CostCenterID` field in the reviewed code, despite the PRD requiring every expense entry to be tagged with a cost center. Without this, departmental P&L and cost center comparison reports are impossible.

**Solution:**
Create `domain/costcenter.go`:

```go
type CostCenter struct {
    ID              uuid.UUID
    TenantID        uuid.UUID
    Code            string
    Name            string
    ParentID        *uuid.UUID
    IsGroup         bool
    IsDistributed   bool
    IsActive        bool
    // Distributed allocation config (JSON)
    AllocationMethod  *string  // "PERCENTAGE", "HEADCOUNT", "SQFT"
    AllocationTargets []CostCenterAllocation
    CreatedAt         time.Time
    UpdatedAt         time.Time
}

type CostCenterAllocation struct {
    TargetCostCenterID uuid.UUID
    Percentage         decimal.Decimal
}
```

Add `CostCenterID *uuid.UUID` field to `TransactionEntry` struct.
Add `CostCenterRepository`, `CostCenterService` with CRUD + `AllocateDistributed(periodID)`.

**Expected Outcome:**
- Every expense transaction entry can be tagged with a cost center
- Group-level cost centers aggregate children totals
- Distributed cost centers auto-generate allocation journal entries on month-end
- `GetCostCenterPL(costCenterID, from, to)` returns income statement scoped to that center

**How to Test:**
```go
func TestCostCenterAllocation(t *testing.T) {
    // Create IT dept (distributed): Sales 40%, Ops 30%, Admin 20%, R&D 10%
    // Post 500,000 expense to IT dept
    // Run AllocateDistributed(periodID)
    // Verify 4 allocation entries: 200k, 150k, 100k, 50k
    // Verify IT dept balance = 0 after allocation
    // Verify Sales CC balance = 200,000
}
```

---

### TASK-007 `[ ]` Create `Budget` and `BudgetLine` domain models

**Priority:** P1
**File:** `internal/core/finance/domain/budget.go` (new file)

**Problem:**
`Accounts.IsBudgetable` and `Accounts.BudgetVarianceThreshold` exist in the domain, but there is no `Budget` entity, no `BudgetLine`, and no service to check budget availability at transaction posting time. The PRD's budget controls (soft warn, hard block) cannot function without these.

**Solution:**
Create `domain/budget.go`:

```go
type BudgetStatus string
const (
    BudgetStatusDraft    BudgetStatus = "DRAFT"
    BudgetStatusApproved BudgetStatus = "APPROVED"
    BudgetStatusActive   BudgetStatus = "ACTIVE"
    BudgetStatusArchived BudgetStatus = "ARCHIVED"
)

type BudgetControlType string
const (
    BudgetControlNone        BudgetControlType = "NONE"
    BudgetControlSoft        BudgetControlType = "SOFT"   // warn, allow override
    BudgetControlHard        BudgetControlType = "HARD"   // block
    BudgetControlHierarchical BudgetControlType = "HIERARCHICAL"
)

type Budget struct {
    ID             uuid.UUID
    TenantID       uuid.UUID
    FiscalYearID   uuid.UUID
    Name           string
    VersionLabel   string          // "V1", "V2-Revised"
    Status         BudgetStatus
    ControlType    BudgetControlType
    Currency       string
    ApprovedBy     *uuid.UUID
    ApprovedAt     *time.Time
    CreatedAt      time.Time
    UpdatedAt      time.Time
}

type BudgetLine struct {
    ID             uuid.UUID
    BudgetID       uuid.UUID
    AccountCode    string
    CostCenterID   *uuid.UUID
    PeriodAmounts  map[int]decimal.Decimal // period_number → amount
    AnnualTotal    decimal.Decimal
}
```

Add `BudgetService` with: `CheckBudget(accountCode, costCenterID, periodID, amount)` returning `BudgetCheckResult{Available, Requested, WouldExceed, ControlType}`.

Wire `CheckBudget` into `TransactionService.PostTransaction` — soft: add warning to response; hard: return `BUDGET_EXCEEDED` error.

**Expected Outcome:**
- Posting an expense that exceeds soft budget returns `HTTP 200` with `warnings: ["BUDGET_SOFT_EXCEEDED"]`
- Posting an expense that exceeds hard budget returns `HTTP 422` with `error.code: "BUDGET_EXCEEDED"`
- `GetBudgetVariance(fiscalYearID, periodID, costCenterID)` returns budget vs. actual with variance %

**How to Test:**
```go
func TestBudgetSoftControl(t *testing.T) {
    // Budget: 100,000 for Rent/Jan, control=SOFT
    // Post 80,000 → success, no warning
    // Post 25,000 more (total 105,000) → success + warning "BUDGET_SOFT_EXCEEDED"
}

func TestBudgetHardControl(t *testing.T) {
    // Budget: 100,000 for Rent/Jan, control=HARD
    // Post 80,000 → success
    // Post 25,000 more → BUDGET_EXCEEDED error, transaction not posted
}
```

---

### TASK-008 `[ ]` Document and enforce `AccountStatus` state machine

**Priority:** P1
**File:** `internal/core/finance/domain/accounts.go` + PRD update

**Problem:**
`AccountStatus` has 14 states in code but the PRD only acknowledges `ACTIVE` and `INACTIVE`. States like `AUDIT_LOCK`, `COMPLIANCE_HOLD`, `YEAR_END_PROCESSING`, `SYSTEM_MAINTENANCE` are meaningful but undocumented. The `AllowedTransitions` array exists in the struct definition but the actual transitions are not populated in the reviewed code. Without enforced transitions, accounts can jump to any state arbitrarily.

**Solution:**
1. Populate `AllowedTransitions` with the full valid state machine:

```go
var AccountStatusTransitions = []StatusTransition{
    {From: AccountStatusDraft,           To: AccountStatusPendingApproval, RequiredPermission: "finance.accounts.create"},
    {From: AccountStatusPendingApproval, To: AccountStatusActive,          RequiredPermission: "finance.accounts.approve"},
    {From: AccountStatusPendingApproval, To: AccountStatusDraft,           RequiredPermission: "finance.accounts.approve"}, // reject back
    {From: AccountStatusActive,          To: AccountStatusInactive,         RequiredPermission: "finance.accounts.deactivate"},
    {From: AccountStatusActive,          To: AccountStatusSuspended,        RequiredPermission: "finance.accounts.update"},
    {From: AccountStatusActive,          To: AccountStatusAuditLock,        RequiredPermission: "finance.accounts.update"},
    {From: AccountStatusActive,          To: AccountStatusComplianceHold,   RequiredPermission: "finance.settings.update"},
    {From: AccountStatusSuspended,       To: AccountStatusActive,           RequiredPermission: "finance.accounts.update"},
    {From: AccountStatusInactive,        To: AccountStatusActive,           RequiredPermission: "finance.accounts.update"},
    {From: AccountStatusAuditLock,       To: AccountStatusActive,           RequiredPermission: "finance.periods.lock"}, // CFO only
    {From: AccountStatusComplianceHold,  To: AccountStatusActive,           RequiredPermission: "finance.settings.update"},
    {From: AccountStatusActive,          To: AccountStatusClosed,           RequiredPermission: "finance.accounts.deactivate"},
    {From: AccountStatusClosed,          To: AccountStatusArchived,         RequiredPermission: "finance.accounts.deactivate"},
}
```

2. Add `CanTransitionTo(newStatus AccountStatus) bool` method that checks the transitions table.
3. Add PRD section documenting all 14 states and their meanings.

**Expected Outcome:**
- Invalid status transitions return `INVALID_STATUS_TRANSITION` error
- `AUDIT_LOCK` and `COMPLIANCE_HOLD` require elevated permissions
- Documented states give developers a clear mental model

**How to Test:**
```go
func TestAccountStatusTransitions(t *testing.T) {
    a := &Accounts{Status: AccountStatusActive}
    assert.True(t,  a.CanTransitionTo(AccountStatusInactive))
    assert.True(t,  a.CanTransitionTo(AccountStatusSuspended))
    assert.False(t, a.CanTransitionTo(AccountStatusDraft))      // can't go back to draft
    assert.False(t, a.CanTransitionTo(AccountStatusArchived))   // must close first
}
```

---

### TASK-009 `[ ]` Implement `GetTransactionWithEntries`

**Priority:** P1
**File:** `internal/core/finance/service/transaction_service.go:36`

**Problem:**
`GetTransactionWithEntries` is commented out as TODO in the `TransactionService` interface. This method is needed for: posting validation (confirm entries balance before GL update), reversal creation (need all entries to mirror them), display of a full journal entry in the UI, and API response for `GET /journal-entries/{id}`. Without it, every caller must make two separate calls and assemble the result themselves.

**Solution:**
1. Uncomment and implement the method signature:
```go
GetTransactionWithEntries(ctx context.Context, id uuid.UUID) (*domain.TransactionWithEntries, error)
```

2. `TransactionWithEntries` already exists in `domain/transaction.go` — implement the repository query:
```sql
SELECT t.*, te.*
FROM transactions t
LEFT JOIN transaction_entries te ON te.transaction_id = t.id
WHERE t.id = $1 AND t.tenant_id = $2
ORDER BY te.line_number ASC
```

3. Wire the method into `PostTransaction` — call `GetTransactionWithEntries` and run `IsBalanced()` check before writing to GL.

**Expected Outcome:**
- `GetTransactionWithEntries` returns the transaction header and all its lines in a single call
- `PostTransaction` validates balance using this method before committing
- API `GET /journal-entries/{id}` returns lines without a second round trip

**How to Test:**
```go
func TestGetTransactionWithEntries(t *testing.T) {
    // Seed transaction with 2 entries
    // Call GetTransactionWithEntries(id)
    // Assert Entries has 2 items
    // Assert IsBalanced() == true
    // Assert TotalDebitAmount == TotalCreditAmount
}
```

---

### TASK-010 `[ ]` Document two-level account grouping (`AccountGroupID` + `AccountHeaderID`)

**Priority:** P1
**File:** `internal/core/finance/domain/accounts.go` + PRD update

**Problem:**
`Accounts` has both `AccountGroupID` and `AccountHeaderID`. The PRD only documents a single parent-child hierarchy via `AccountPath`. Two additional grouping layers are unexplained — developers don't know when to use Groups vs. Headers vs. Parent hierarchy. This will lead to inconsistent data entry.

**Solution:**
Define the three-tier grouping model clearly:

```
AccountHeader (top-level grouping, e.g. "Current Assets")
    └── AccountGroup (mid-level grouping, e.g. "Cash & Cash Equivalents")
            └── Account (leaf, e.g. "1120 - Checking Account - Main")
```

1. Create `domain/account_group.go` if not already complete — ensure `AccountGroup` and `AccountHeader` structs are defined with their roles.
2. Update PRD COA section to document the three-tier structure.
3. Add validation: an account cannot have `AccountGroupID` from a different `RootType` than the account.

**Expected Outcome:**
- New developers understand when to use Header vs. Group vs. Parent
- Financial statements use Headers for top-level sections, Groups for sub-sections
- COA import template supports all three tiers

**How to Test:**
```go
func TestAccountGroupingValidation(t *testing.T) {
    // Account: RootType=ASSET
    // Group: RootType=LIABILITY → validation error
    // Group: RootType=ASSET → valid
}
```

---

## P1 — Missing Service Implementations

---

### TASK-011 `[ ]` Implement `AccountService.CreateAccount` fully

**Priority:** P1
**File:** `internal/core/finance/service/account_service.go`

**Problem:**
The `AccountService` interface defines `CreateAccount` but the implementation beyond line 120 is unverified. Critical business rules from the PRD must be enforced in the service layer: unique code validation, parent root type consistency, leaf-account enforcement, materialized path generation, feature flag checks.

**Solution:**
Ensure `CreateAccount` enforces:
1. `account_code` is unique per tenant (call `ValidateAccountCode`)
2. `root_type` matches parent's `root_type` if `parent_account_id` is set
3. New account starts as `AccountStatusDraft` (requires approval workflow) or `AccountStatusActive` if `AllowManualAccountCreation` feature flag is on
4. `account_path` is computed from parent's path + `/` + new code
5. `normal_balance` is auto-set from `GetNormalBalanceForRootType` if not provided
6. `currency_code` defaults to tenant base currency from settings

**Expected Outcome:**
- Duplicate account code → `DUPLICATE_ACCOUNT_CODE` error
- Mismatched root type with parent → `ROOT_TYPE_MISMATCH` error
- Missing currency → defaults to tenant base currency (no error)
- `account_path` is correctly set for deeply nested accounts

**How to Test:**
```go
func TestCreateAccount_DuplicateCode(t *testing.T) {
    // Create account code "1120"
    // Create again with same code → DUPLICATE_ACCOUNT_CODE error
}

func TestCreateAccount_RootTypeMismatch(t *testing.T) {
    // Parent: RootType=ASSET
    // Child: RootType=LIABILITY → ROOT_TYPE_MISMATCH error
}

func TestCreateAccount_PathGeneration(t *testing.T) {
    // Parent path: "/1000/1100/"
    // New account code: "1110"
    // Expected path: "/1000/1100/1110/"
}
```

---

### TASK-012 `[ ]` Implement `TransactionService.PostTransaction` fully

**Priority:** P1
**File:** `internal/core/finance/service/transaction_service.go`

**Problem:**
`PostTransaction` is defined in the interface but correctness requires it to execute a specific sequence of validations in order (as documented in PRD Rule Validation Execution Order). The risk of a partial implementation is silent data corruption — unbalanced entries reaching the GL, or postings to closed periods.

**Solution:**
Implement the 7-step validation sequence:

```go
func (s *transactionService) PostTransaction(ctx context.Context, id uuid.UUID, postingDate *time.Time) (*domain.Transaction, error) {
    // Step 1: Load transaction with entries
    txn, err := s.GetTransactionWithEntries(ctx, id)

    // Step 2: Structural validation
    if !txn.IsBalanced() → UNBALANCED_TRANSACTION
    if len(txn.Entries) < 2 → INSUFFICIENT_ENTRIES

    // Step 3: Status check
    if txn.TransactionStatus != APPROVED && approvalRequired → APPROVAL_REQUIRED
    if txn.TransactionStatus == POSTED → ALREADY_POSTED

    // Step 4: Period check (requires TASK-004)
    period := s.periodRepo.GetOpenPeriodForDate(txn.TransactionDate)
    if period == nil || !period.CanPost() → PERIOD_CLOSED

    // Step 5: Account validation per entry
    for each entry: account must be active, leaf, allow entries

    // Step 6: Budget check (requires TASK-007)
    for each expense entry: s.budgetService.CheckBudget(...)

    // Step 7: Write to GL (transactional)
    BEGIN TRANSACTION
      UPDATE transaction SET status=POSTED, posting_date=now
      UPDATE account_balances for each entry
      INSERT audit_log
    COMMIT
}
```

**Expected Outcome:**
- Unbalanced transaction → `UNBALANCED_TRANSACTION` error, nothing posted
- Closed period → `PERIOD_CLOSED` error, nothing posted
- All validations pass → status becomes `POSTED`, account balances updated atomically

**How to Test:**
```go
func TestPostTransaction_Unbalanced(t *testing.T) {
    // Transaction: Dr 50,000 / Cr 40,000 (unbalanced)
    // PostTransaction → UNBALANCED_TRANSACTION
    // Transaction status remains APPROVED (not POSTED)
}

func TestPostTransaction_Success(t *testing.T) {
    // Balanced transaction in open period
    // PostTransaction → status == POSTED
    // Account 1120 balance decremented
    // Account 7310 balance incremented
}
```

---

### TASK-013 `[ ]` Implement `TransactionService.ReverseTransaction` fully

**Priority:** P1
**File:** `internal/core/finance/service/transaction_service.go`

**Problem:**
`ReverseTransaction` is defined but a correct reversal must: create a mirror entry with all debits/credits swapped, link back to the original, post immediately (or require approval per policy), mark the original as `REVERSED`. An incomplete implementation risks double-counting or orphaned reversals.

**Solution:**
Use `domain.Transaction.CreateReversalTransaction()` (already exists) then:
1. Validate original is `POSTED` (only posted entries can be reversed)
2. Call `CreateReversalTransaction(reverseDate, reason)` from domain
3. Save the reversal transaction
4. Post it (using `PostTransaction` — applies same validations including period check)
5. Update original: `IsReversed = true`, `ReversedByTransactionID = reversalID`

Both operations must be in a single DB transaction.

**Expected Outcome:**
- Original transaction marked `REVERSED` with link to reversal
- Reversal transaction is `POSTED` with reference back to original
- Net GL effect = zero
- Cannot reverse a transaction twice (original already `REVERSED`)

**How to Test:**
```go
func TestReverseTransaction(t *testing.T) {
    // Post JE: Dr Rent 50,000 / Cr Cash 50,000
    // ReverseTransaction(id, reason)
    // Original status → REVERSED
    // New reversal: Dr Cash 50,000 / Cr Rent 50,000
    // Cash account net change = 0
    // Rent account net change = 0
}

func TestReverseAlreadyReversed(t *testing.T) {
    // Reverse transaction → succeeds
    // Reverse same transaction again → TRANSACTION_ALREADY_REVERSED error
}
```

---

## P1 — Missing Infrastructure

---

### TASK-014 `[ ]` Add period validation middleware to transaction routes

**Priority:** P1
**File:** `internal/core/finance/` (handler/middleware layer)

**Problem:**
Even with period domain models built (TASK-004), period validation only helps if it is consistently enforced on every write path — not just in `PostTransaction`. A DRAFT transaction with a date in a locked period should still warn the user on creation, not only at post time.

**Solution:**
Create a `PeriodGuard` middleware/helper that:
- On `CreateTransaction`: warn (not block) if `transaction_date` is in closed period
- On `UpdateTransaction`: warn if changing `transaction_date` to a closed period
- On `PostTransaction`: hard block if period is not open/soft-closed
- On `ApproveTransaction`: soft check only (approval can occur after soft-close)

**Expected Outcome:**
- Users see actionable warnings when entering transactions for closed periods
- Posting is hard-blocked
- Approval is not blocked (approver may be in a different timezone)

**How to Test:**
- Create transaction in closed period → `HTTP 201` with `warnings: ["PERIOD_SOFT_CLOSED"]`
- Post transaction in closed period → `HTTP 422` with `error.code: "PERIOD_CLOSED"`

---

### TASK-015 `[ ]` Enable and implement Temporal integration

**Priority:** P1
**File:** `internal/core/finance/temporal_integration.go`

**Problem:**
`temporal_integration.go` is 265 lines of commented-out code. The finance module needs async workflows for: recurring transaction generation, approval SLA escalation, scheduled report delivery, and period-end automation (depreciation, accrual reversal). Without Temporal, these must be done manually or via fragile cron jobs.

**Solution:**
1. Uncomment and complete `TemporalIntegration` struct
2. Register these activity types:
   - `GenerateRecurringTransactionActivity` — creates transaction from template
   - `PostTransactionActivity` — wraps `TransactionService.PostTransaction`
   - `EscalateApprovalActivity` — sends notification, reassigns approver
   - `ReversalActivity` — auto-reversal for accruals
3. Register these workflow types:
   - `RecurringTransactionWorkflow` — cron-scheduled, calls generate + post
   - `ApprovalEscalationWorkflow` — timer-based, escalates after SLA breach
   - `PeriodEndWorkflow` — triggered on period soft-close, runs depreciation + accrual reversal

**Expected Outcome:**
- `transaction.is_recurring = true` with `recurring_frequency = "MONTHLY"` → Temporal cron runs monthly
- Approval pending > 8 hours → Finance Manager gets escalation notification
- Month-end soft close triggers auto-depreciation entries

**How to Test:**
```go
// Use Temporal test framework (testsuite.WorkflowTestSuite)
func TestRecurringTransactionWorkflow(t *testing.T) {
    suite := &testsuite.WorkflowTestSuite{}
    env := suite.NewTestWorkflowEnvironment()
    env.RegisterActivity(&GenerateRecurringTransactionActivity{})
    env.ExecuteWorkflow(RecurringTransactionWorkflow, templateID)
    env.AssertExpectations(t)
    // Verify new transaction was created with correct date
}
```

---

## P2 — Reconciliation & Payments

---

### TASK-016 `[ ]` Create `BankReconciliation` domain model and service

**Priority:** P2
**File:** `internal/core/finance/domain/reconciliation.go` (new file)

**Problem:**
`TransactionEntry.IsReconciled()` and `MarkReconciled()` methods exist, but there is no `BankReconciliation` entity to hold the workspace state, statement lines, or match results. Bank reconciliation is a mandatory control — the period-close checklist must verify that all bank accounts have an approved reconciliation before hard-close is allowed.

**Solution:**
Create domain model:

```go
type ReconciliationStatus string
const (
    ReconciliationStatusOpen       ReconciliationStatus = "OPEN"
    ReconciliationStatusInProgress ReconciliationStatus = "IN_PROGRESS"
    ReconciliationStatusBalanced   ReconciliationStatus = "BALANCED"
    ReconciliationStatusUnderReview ReconciliationStatus = "UNDER_REVIEW"
    ReconciliationStatusApproved   ReconciliationStatus = "APPROVED"
    ReconciliationStatusLocked     ReconciliationStatus = "LOCKED"
)

type BankReconciliation struct {
    ID                    uuid.UUID
    TenantID              uuid.UUID
    BankAccountCode       string             // Links to COA account
    PeriodID              uuid.UUID
    StatementOpeningBal   decimal.Decimal
    StatementClosingBal   decimal.Decimal
    GLClosingBal          decimal.Decimal
    AdjustedBankBal       decimal.Decimal
    AdjustedGLBal         decimal.Decimal
    Difference            decimal.Decimal
    Status                ReconciliationStatus
    PreparedBy            *uuid.UUID
    ApprovedBy            *uuid.UUID
    ApprovedAt            *time.Time
    CreatedAt             time.Time
    UpdatedAt             time.Time
}

type StatementLine struct {
    ID               uuid.UUID
    ReconciliationID uuid.UUID
    LineDate         time.Time
    Description      string
    Debit            decimal.Decimal
    Credit           decimal.Decimal
    Reference        string
    IsMatched        bool
    MatchType        string  // "AUTO_EXACT", "AUTO_SUGGESTED", "MANUAL", "TIMING_DIFF", "BANK_ERROR"
    MatchedEntryIDs  []uuid.UUID
    Note             *string
}
```

Add `ReconciliationService` with: `ImportStatement`, `RunAutoMatch`, `ConfirmMatch`, `SubmitForReview`, `Approve`.

Wire: `PeriodService.HardClosePeriod` must check all bank accounts have `APPROVED` reconciliation.

**Expected Outcome:**
- Finance clerk can import a bank statement, auto-match entries, resolve exceptions
- Period cannot hard-close without all bank accounts reconciled
- Reconciliation report is exportable as PDF

**How to Test:**
```go
func TestAutoMatching(t *testing.T) {
    // Load 3 statement lines
    // Load 3 matching GL entries (same amount + date + ref)
    // RunAutoMatch → 3 matched, 0 unmatched
}

func TestPeriodCloseBlockedByUnreconciledAccount(t *testing.T) {
    // Open reconciliation exists for checking account
    // Attempt HardClosePeriod → BANK_ACCOUNT_NOT_RECONCILED error
}
```

---

### TASK-017 `[ ]` Create `PaymentRun` domain model and service

**Priority:** P2
**File:** `internal/core/finance/domain/payment_run.go` (new file)

**Problem:**
Bulk supplier payments (payment runs) are documented in the PRD's Cash Management section and the API spec, but there is no domain model. Without a payment run entity, each AP invoice must be paid individually, which is operationally impractical and doesn't support bank file generation.

**Solution:**
```go
type PaymentRunStatus string
const (
    PaymentRunStatusDraft     PaymentRunStatus = "DRAFT"
    PaymentRunStatusApproved  PaymentRunStatus = "APPROVED"
    PaymentRunStatusExported  PaymentRunStatus = "EXPORTED"   // file generated
    PaymentRunStatusConfirmed PaymentRunStatus = "CONFIRMED"  // bank processed
    PaymentRunStatusPosted    PaymentRunStatus = "POSTED"     // GL entries posted
    PaymentRunStatusCancelled PaymentRunStatus = "CANCELLED"
)

type PaymentRun struct {
    ID              uuid.UUID
    TenantID        uuid.UUID
    BankAccountCode string
    Currency        string
    DueOnOrBefore   time.Time
    TotalAmount     decimal.Decimal
    Status          PaymentRunStatus
    ApprovedBy      *uuid.UUID
    ExportFormat    string            // "equity_eft", "kcb_rtgs", "swift_mt101"
    ExportedAt      *time.Time
    PostedAt        *time.Time
    CreatedBy       uuid.UUID
    CreatedAt       time.Time
    UpdatedAt       time.Time
    Lines           []PaymentRunLine
}

type PaymentRunLine struct {
    ID            uuid.UUID
    PaymentRunID  uuid.UUID
    SupplierID    uuid.UUID
    InvoiceID     uuid.UUID
    Amount        decimal.Decimal
    BankAccount   string
    BankCode      string
    Reference     string
}
```

**Expected Outcome:**
- Batch payment run created from due AP invoices
- Bank file generated in correct format for upload
- On confirmation, single GL entry: Dr AP (multiple) / Cr Bank
- Payment run cannot be approved by its creator (SOD)

**How to Test:**
```go
func TestPaymentRunCreation(t *testing.T) {
    // 3 due invoices: 232k, 58k, 17.4k
    // CreatePaymentRun → total 307,400
    // Approve → status APPROVED
    // ExportFile("equity_eft") → valid file content
    // Confirm → 3 AP lines posted, Cash reduces by 307,400
}
```

---

## P2 — Discrepancies to Clarify / Minor Fixes

---

### TASK-018 `[ ]` Clarify `ApprovalStatus.PARTIALLY_APPROVED` and `EXPIRED` in PRD

**Priority:** P2
**File:** PRD + `internal/core/finance/domain/types.go`

**Problem:**
`PARTIALLY_APPROVED` and `EXPIRED` appear in the code's `ApprovalStatus` enum but are not documented anywhere in the PRD. Other developers won't know:
- What triggers `PARTIALLY_APPROVED` (is it for multi-tier approval where tier 1 approved but tier 2 hasn't)?
- What is the approval SLA that triggers `EXPIRED`?
- What happens when a transaction's approval expires — does it go back to `DRAFT`?

**Solution:**
1. Document both statuses in PRD Approval Workflows section:
   - `PARTIALLY_APPROVED`: used in sequential multi-tier approval — tier N has approved, still awaiting tier N+1
   - `EXPIRED`: approval request not acted on within configured SLA (e.g. 48 hours) — Temporal timer triggers this
2. Define the expiry-to-status flow: `EXPIRED` → transaction moves to `DRAFT` with notification to submitter
3. Implement the expiry logic in `ApprovalEscalationWorkflow` (see TASK-015)

**Expected Outcome:**
- Developers understand all 6 `ApprovalStatus` values and when to use them
- Expiry is handled by Temporal workflow, not polling
- PRD and code are consistent

**How to Test:**
- Approval pending 48h → status becomes `EXPIRED` → transaction returns to `DRAFT`
- Tier 1 of 2 approves → `PARTIALLY_APPROVED` → tier 2 notified

---

### TASK-019 `[ ]` Add `CostCenterID` to `TransactionEntry`

**Priority:** P2
**File:** `internal/core/finance/domain/transaction_entry.go`

**Problem:**
After TASK-006 creates the `CostCenter` model, `TransactionEntry` must reference it. Cost centers are assigned at the line level (not just the header) because a single journal entry may span multiple departments (e.g., a shared expense split between Sales and Admin).

**Solution:**
Add field to `TransactionEntry`:
```go
CostCenterID *uuid.UUID `json:"cost_center_id,omitempty"`
```

Update `CreateEntryRequest` to include `CostCenterID`.
Update validation: if tenant has `cost_center_required_for_expenses` setting enabled and `account.root_type == EXPENSE`, `CostCenterID` must not be nil.

**Expected Outcome:**
- Expense entries can be tagged per-line with a cost center
- Cost-center-required setting enforces tagging at posting time
- `GetCostCenterPL` can aggregate by `CostCenterID` on entries

**How to Test:**
```go
func TestCostCenterRequiredValidation(t *testing.T) {
    // Setting: cost_center_required_for_expenses = true
    // Post expense entry without cost_center_id → COST_CENTER_REQUIRED error
    // Post expense entry with cost_center_id → success
}
```

---

### TASK-020 `[ ]` Add `AccountFilter` — confirm `RootType` and `AccountType` filtering

**Priority:** P2
**File:** `internal/core/finance/domain/accounts.go`

**Problem:**
`AccountService.ListAccounts` takes an `*domain.AccountFilter` but the filter struct was not fully reviewed. The API spec documents filtering by `root_type`, `account_type`, `is_active`, `is_group`, `parent_code`. If `AccountFilter` is missing these fields, the API handler cannot implement them.

**Solution:**
Confirm (or add) to `AccountFilter`:
```go
type AccountFilter struct {
    TenantID    uuid.UUID
    EntityID    *uuid.UUID
    RootType    *RootType
    AccountType *string
    IsActive    *bool
    IsGroup     *bool
    ParentCode  *string
    Query       string        // full-text search on code + name
    Page        int
    PerPage     int
}
```

**Expected Outcome:**
- `GET /accounts?root_type=asset&is_group=false` returns only active leaf asset accounts
- Pagination works correctly
- Empty filter returns all accounts for the tenant (scoped by tenant_id from context)

**How to Test:**
```go
func TestAccountFilter_ByRootType(t *testing.T) {
    // Seed 5 asset, 3 liability accounts
    // Filter RootType=ASSET → 5 results
    // Filter RootType=LIABILITY → 3 results
}
```

---

## P3 — Polish & Documentation

---

### TASK-021 `[ ]` Remove backward-compatibility method duplicates from service interfaces

**Priority:** P3
**File:** `internal/core/finance/service/account_service.go`, `transaction_service.go`

**Problem:**
Both services have commented "Handler convenience methods (for backward compatibility)" that duplicate the main methods:
- `Create` → `CreateAccount`
- `GetByID` → `GetAccountByID`
- `List` → `ListAccounts`
- etc.

This doubles the interface surface area and makes it unclear which method to use. The comment says "backward compatibility" but there is no legacy caller to be compatible with — this is a greenfield module.

**Solution:**
Remove the duplicate shim methods. Keep only the descriptive names (`CreateAccount`, `GetAccountByID`, etc.). Update any callers.

**Expected Outcome:**
- `AccountService` interface has ~40 methods instead of ~50
- Clear, consistent naming throughout
- No ambiguity for new developers

**How to Test:**
- Grep: no remaining callers of `accountService.Create`, `accountService.GetByID`, `accountService.List`
- All tests pass after removal

---

### TASK-022 `[ ]` Add severity levels to `ValidationError`

**Priority:** P3
**File:** `internal/core/finance/domain/types.go:338`

**Problem:**
There is a TODO comment in the code:
```go
// TODO: Add severity levels to ValidationError (ERROR, WARNING, INFO)
// TODO: Add support for nested field paths (e.g., "entries[0].amount")
```
The API spec returns `warnings` separate from `errors` — this requires the `ValidationError` struct to carry a severity level. Without it, soft budget warnings cannot be distinguished from hard validation errors in the API response.

**Solution:**
```go
type ValidationSeverity string
const (
    ValidationSeverityInfo    ValidationSeverity = "INFO"
    ValidationSeverityWarning ValidationSeverity = "WARNING"
    ValidationSeverityError   ValidationSeverity = "ERROR"
)

type ValidationError struct {
    Field    string             `json:"field"`
    Message  string             `json:"message"`
    Code     string             `json:"code"`
    Severity ValidationSeverity `json:"severity"`
}
```

Update API response builder to split `ValidationErrors` into `errors` (ERROR severity) and `warnings` (WARNING/INFO severity).

**Expected Outcome:**
- Budget soft-exceed produces `WARNING` severity — request succeeds
- Unbalanced entry produces `ERROR` severity — request fails
- API response clearly separates errors from warnings

**How to Test:**
```go
func TestValidationSeveritySplit(t *testing.T) {
    errors, warnings := splitValidationErrors([]ValidationError{
        {Severity: "ERROR",   Code: "UNBALANCED"},
        {Severity: "WARNING", Code: "BUDGET_SOFT_EXCEEDED"},
    })
    assert.Len(t, errors, 1)
    assert.Len(t, warnings, 1)
}
```

---

### TASK-023 `[ ]` Add `TransactionEntry` nested field path support

**Priority:** P3
**File:** `internal/core/finance/domain/types.go`

**Problem:**
`ValidationError.Field` currently stores flat field names like `"amount"`. When validating a multi-line journal entry, the error for line 2 should reference `"entries[1].amount"` not just `"amount"`. Without nested paths, the UI cannot highlight the specific line that failed.

**Solution:**
Add a helper for building nested paths:
```go
func FieldPath(parts ...string) string {
    // entries[1].account_code → from ("entries", 1, "account_code")
}
```

Update `TransactionEntry.Validate()` to pass the entry index into the validation context so errors carry `entries[N].field_name`.

**Expected Outcome:**
- API error response for a multi-line JE shows: `"field": "entries[1].account_code"` not `"field": "account_code"`
- UI can scroll to the exact line that failed

**How to Test:**
```go
func TestNestedFieldPath(t *testing.T) {
    // 3-line entry, line index 1 has invalid account
    // errors[0].Field == "entries[1].account_code"
}
```

---

## Task Summary

| ID | Task | Priority | Effort | Depends On |
|----|------|----------|--------|------------|
| TASK-001 | Fix `IsEditable()` contradiction | P0 | XS | — |
| TASK-002 | Fix `RejectionReason` enum | P0 | XS | — |
| TASK-003 | Remove duplicate `TransactionType` | P0 | S | — |
| TASK-004 | `FiscalYear` + `AccountingPeriod` domain | P1 | L | — |
| TASK-005 | `Currency` + `ExchangeRate` domain | P1 | L | — |
| TASK-006 | `CostCenter` domain | P1 | M | — |
| TASK-007 | `Budget` + `BudgetLine` domain | P1 | M | TASK-004, TASK-006 |
| TASK-008 | `AccountStatus` state machine | P1 | M | — |
| TASK-009 | Implement `GetTransactionWithEntries` | P1 | S | — |
| TASK-010 | Document account grouping tiers | P1 | S | — |
| TASK-011 | Implement `CreateAccount` fully | P1 | M | TASK-008 |
| TASK-012 | Implement `PostTransaction` fully | P1 | L | TASK-004, TASK-005, TASK-007, TASK-009 |
| TASK-013 | Implement `ReverseTransaction` fully | P1 | M | TASK-012 |
| TASK-014 | Period validation middleware | P1 | S | TASK-004 |
| TASK-015 | Enable Temporal integration | P1 | L | TASK-012, TASK-013 |
| TASK-016 | `BankReconciliation` domain | P2 | L | TASK-004 |
| TASK-017 | `PaymentRun` domain | P2 | M | — |
| TASK-018 | Document `PARTIALLY_APPROVED` + `EXPIRED` | P2 | XS | TASK-015 |
| TASK-019 | Add `CostCenterID` to `TransactionEntry` | P2 | XS | TASK-006 |
| TASK-020 | Confirm `AccountFilter` completeness | P2 | S | — |
| TASK-021 | Remove duplicate service method shims | P3 | S | — |
| TASK-022 | `ValidationError` severity levels | P3 | S | — |
| TASK-023 | Nested field paths in `ValidationError` | P3 | S | TASK-022 |

**Effort key:** XS < 2h · S = half-day · M = 1–2 days · L = 3–5 days
