# Finance Module — Awo ERP

**Version:** 1.0.0  
**Status:** Architecture Specification  
**Audience:** Engineering · Finance · Compliance  
**Stack:** Go · PostgreSQL · Temporal · pkg/condition

---

## Table of Contents

1. [Overview & Design Philosophy](#1-overview--design-philosophy)
2. [Domain Model](#2-domain-model)
3. [Module Architecture](#3-module-architecture)
4. [Chart of Accounts Engine](#4-chart-of-accounts-engine)
5. [Journal Entry Pipeline](#5-journal-entry-pipeline)
6. [General Ledger Engine](#6-general-ledger-engine)
7. [Period Management](#7-period-management)
8. [Bank Reconciliation Engine](#8-bank-reconciliation-engine)
9. [Accounts Receivable & Payable](#9-accounts-receivable--payable)
10. [Cost Centre & Budget Management](#10-cost-centre--budget-management)
11. [Cash Management & Forecasting](#11-cash-management--forecasting)
12. [Intercompany Module](#12-intercompany-module)
13. [Financial Reporting Engine](#13-financial-reporting-engine)
14. [GL Integration Contract](#14-gl-integration-contract)
15. [API Reference](#15-api-reference)
16. [Business Rules & Validation](#16-business-rules--validation)
17. [Authorization Model](#17-authorization-model)
18. [Database Schema & RLS](#18-database-schema--rls)
19. [Workflow Orchestration (Temporal)](#19-workflow-orchestration-temporal)
20. [Configuration Flags](#20-configuration-flags)

---

## 1. Overview & Design Philosophy

The Finance module is the authoritative ledger of Awo ERP. Every monetary event in the system — a payroll run, a fuel sale, an inventory write-off, a foreign currency revaluation — ultimately produces a journal entry that lands in this module's General Ledger. The Finance module owns the accounts, the periods, the journal entries, and the financial statements. No other module writes journal entries directly.

### Design Principles

**Double-entry at the database layer.** The GL enforces double-entry integrity at the constraint level, not just the application layer. An unbalanced entry cannot exist in the database.

**Append-only ledger.** Posted journal entries are immutable. Corrections are always reversals followed by new correct entries. The ledger grows in one direction only — forward.

**Period gating.** All posting goes through a period gate. An open period accepts entries; a closed period does not. This is enforced by a trigger, not by application code alone.

**Event consumer, not event producer (for business data).** The Finance module consumes events from HR, Sales, Inventory, Pricing, and Currency modules. It does not reach into their tables. When it has posted a journal entry, it emits a confirmation event back.

**Subsidiary ledger pattern.** The GL operates at control account level. AR, AP, inventory, and employee receivables are subsidiary ledgers that roll up to GL control accounts. Reconciliation between the subsidiary and the control is a first-class operation.

**Multi-entity by design.** From day one, the module supports multiple legal entities (organisations) within a single tenant. Each organisation has its own chart of accounts, periods, and GL. Consolidation across entities is a native operation.

### Scope Boundary

| In Scope | Out of Scope |
|---|---|
| Chart of accounts management | Currency master and exchange rates (Currency module) |
| Journal entry lifecycle (draft → posted → reversed) | Payroll computation (HR module) |
| General Ledger and subsidiary ledgers | Invoice generation (Sales module) |
| Period and fiscal year management | Purchase order management (Procurement module) |
| Bank reconciliation engine | Asset depreciation computation (Fixed Assets module) |
| Accounts receivable and payable aging | Tax computation (Tax module / Pricing module) |
| Cost centre and budget management | Subscription billing (Billing module) |
| Cash position and 13-week forecasting | FX hedge contract management (Currency module) |
| Intercompany accounting and eliminations | |
| Financial statement generation | |
| Payment run management | |
| Petty cash management | |

---

## 2. Domain Model

### Aggregate Roots

```
Organisation        — a legal entity within a tenant; owns its own COA, periods, and GL
Account             — a node in the chart of accounts hierarchy
FiscalYear          — a 12-month reporting period with sub-periods
AccountingPeriod    — a sub-period within a fiscal year (monthly or quarterly)
JournalEntry        — an immutable (once posted) collection of balanced debit/credit lines
BankReconciliation  — a workspace for matching GL entries to bank statement lines
Budget              — a versioned collection of account + cost centre + period targets
PaymentRun          — a batch of supplier payments processed together
PettyCashFund       — an imprest fund with a fixed float amount
```

### Core Entities

```
AccountBalance      — the running balance of an account in a period (materialised)
JournalLine         — a single debit or credit line within a JournalEntry
BankAccount         — a bank account linked to a GL cash account
BankStatement       — an imported bank statement for a period
BankStatementLine   — a single transaction line from a bank statement
ReconciliationMatch — a confirmed match between statement lines and GL entries
CostCentre          — an organisational unit for cost and revenue tracking
BudgetLine          — a single account + cost centre + period budget amount
SubsidiaryEntry     — an AR, AP, or inventory sub-ledger record
IntercompanyLink    — a matching pair of IC transactions across two entities
```

### Value Objects

```
Money               — amount + ISO 4217 currency (from Currency module)
AccountPath         — materialised path string for hierarchy queries e.g. /1000/1100/1110
NormalBalance       — debit | credit
JournalReference    — auto-generated sequential reference per tenant e.g. JE-2025-0145
PeriodStatus        — open | soft_closed | hard_closed | locked
EntryStatus         — draft | pending_approval | approved | posted | reversed | cancelled
```

### Key Relationships

```
Tenant
  └── Organisation (N, one per legal entity)
        ├── Account (tree, unlimited depth)
        │     └── AccountBalance (one per account per period)
        ├── FiscalYear (N)
        │     └── AccountingPeriod (12 per year)
        ├── BankAccount (N)
        │     └── BankReconciliation (one per period)
        │           └── ReconciliationMatch (N)
        ├── CostCentre (tree)
        ├── Budget (N, versioned)
        │     └── BudgetLine (N per account per period)
        └── JournalEntry (N)
              └── JournalLine (≥2, balanced)

JournalEntry
  └── source_module: "hr" | "sales" | "inventory" | "pricing" | "currency" | "manual"
  └── source_event_id: UUID (the integration event that triggered this entry)
```

---

## 3. Module Architecture

### Package Layout

```
internal/finance/
├── domain/
│   ├── account/
│   │   ├── account.go              # Account aggregate; hierarchy rules
│   │   ├── balance.go              # AccountBalance — materialised running total
│   │   ├── path.go                 # Materialised path helpers
│   │   └── events.go
│   ├── period/
│   │   ├── fiscal_year.go          # FiscalYear aggregate
│   │   ├── accounting_period.go    # AccountingPeriod; status state machine
│   │   └── gate.go                 # PeriodGate — checks posting is allowed
│   ├── journal/
│   │   ├── entry.go                # JournalEntry aggregate
│   │   ├── line.go                 # JournalLine value object
│   │   ├── pipeline.go             # JournalPipeline — staged validation + posting
│   │   ├── stages/
│   │   │   ├── validate_structure.go
│   │   │   ├── validate_balance.go
│   │   │   ├── validate_accounts.go
│   │   │   ├── validate_period.go
│   │   │   ├── validate_budget.go
│   │   │   ├── check_approval.go
│   │   │   └── post_to_gl.go
│   │   └── reversal.go
│   ├── gl/
│   │   ├── ledger.go               # GeneralLedger — balance queries and roll-up
│   │   └── subsidiary.go           # SubsidiaryLedger — AR, AP, inventory sub-ledgers
│   ├── reconciliation/
│   │   ├── workspace.go            # BankReconciliation aggregate
│   │   ├── matcher.go              # AutoMatcher — rule-based matching engine
│   │   ├── statement.go            # BankStatement + BankStatementLine
│   │   └── rules/
│   │       ├── exact_match.go
│   │       ├── date_amount_match.go
│   │       ├── partial_ref_match.go
│   │       └── batch_match.go
│   ├── budget/
│   │   ├── budget.go               # Budget aggregate (versioned)
│   │   ├── line.go                 # BudgetLine
│   │   └── checker.go              # BudgetChecker — soft/hard control
│   ├── cost_centre/
│   │   └── cost_centre.go          # CostCentre aggregate; distributed allocation
│   ├── cashflow/
│   │   ├── position.go             # CashPosition — real-time across all accounts
│   │   └── forecast.go             # 13-week rolling forecast engine
│   ├── intercompany/
│   │   ├── transaction.go          # IntercompanyTransaction aggregate
│   │   ├── matcher.go              # IC balance matching at period end
│   │   └── eliminator.go           # Consolidation elimination generator
│   └── payment/
│       ├── run.go                  # PaymentRun aggregate
│       └── petty_cash.go           # PettyCashFund aggregate
├── application/
│   ├── account_service.go
│   ├── journal_service.go
│   ├── period_service.go
│   ├── reconciliation_service.go
│   ├── budget_service.go
│   ├── cashflow_service.go
│   ├── intercompany_service.go
│   ├── payment_service.go
│   └── event_consumers/
│       ├── hr_consumer.go          # Handles PayrollPosted, DiscrepancyConfirmed, etc.
│       ├── sales_consumer.go       # Handles SaleInvoicePosted, PaymentReceived, etc.
│       ├── inventory_consumer.go
│       ├── pricing_consumer.go
│       └── currency_consumer.go    # Handles FXRevaluationPosted, FXGainLossRealised
├── infrastructure/
│   ├── postgres/queries/
│   ├── importers/
│   │   ├── ofx_importer.go
│   │   ├── mt940_importer.go
│   │   └── csv_importer.go
│   └── temporal/
│       ├── period_close_workflow.go
│       ├── reconciliation_workflow.go
│       ├── budget_allocation_workflow.go
│       └── ic_matching_workflow.go
└── transport/http/
    ├── account_handler.go
    ├── journal_handler.go
    ├── period_handler.go
    ├── reconciliation_handler.go
    ├── budget_handler.go
    ├── payment_handler.go
    └── report_handler.go
```

### Layer Responsibilities

| Layer | Responsibility |
|---|---|
| `domain/journal/pipeline` | Pure validation stages — no I/O |
| `domain/gl` | Balance queries and roll-up — read-only domain logic |
| `domain/reconciliation/matcher` | Matching rule evaluation — pure |
| `application/` | Load data, call domain, persist, emit events, coordinate workflows |
| `application/event_consumers` | Consume events from other modules, build journal entries |
| `infrastructure/importers` | Parse OFX, MT940, CSV into domain objects |
| `infrastructure/temporal` | Period close, scheduled reconciliation, IC matching |
| `transport/` | HTTP handlers |

---

## 4. Chart of Accounts Engine

### Account Hierarchy

The chart of accounts is an unlimited-depth tree. Every account has a materialised path that enables efficient subtree queries without recursive CTEs.

```go
// domain/account/account.go

type RootType string
const (
    RootAsset     RootType = "asset"
    RootLiability RootType = "liability"
    RootEquity    RootType = "equity"
    RootRevenue   RootType = "revenue"
    RootExpense   RootType = "expense"
)

type AccountType string
const (
    TypeBank              AccountType = "bank"
    TypeCash              AccountType = "cash"
    TypeReceivable        AccountType = "receivable"
    TypePayable           AccountType = "payable"
    TypeInventory         AccountType = "inventory"
    TypeFixedAsset        AccountType = "fixed_asset"
    TypeAccumDepr         AccountType = "accumulated_depreciation"
    TypePrepaid           AccountType = "prepaid"
    TypeTaxPayable        AccountType = "tax_payable"
    TypeAccrued           AccountType = "accrued"
    TypeLoan              AccountType = "loan"
    TypeEquityCapital     AccountType = "equity_capital"
    TypeRetainedEarnings  AccountType = "retained_earnings"
    TypeRevenueSales      AccountType = "revenue_sales"
    TypeRevenueService    AccountType = "revenue_service"
    TypeCOGS              AccountType = "cogs"
    TypeExpenseOperating  AccountType = "expense_operating"
    TypeExpenseDepreciation AccountType = "expense_depreciation"
    TypeExpenseInterest   AccountType = "expense_interest"
    TypeOtherIncome       AccountType = "other_income"
    TypeOtherExpense      AccountType = "other_expense"
)

type Account struct {
    ID                   uuid.UUID
    TenantID             TenantID
    OrganisationID       uuid.UUID
    Code                 string          // immutable after transactions exist
    Name                 string
    RootType             RootType
    AccountType          AccountType
    NormalBalance        NormalBalance   // debit | credit
    ParentID             *uuid.UUID
    Path                 AccountPath     // e.g. /1000/1100/1110/1120
    Level                int
    IsGroup              bool            // group accounts cannot receive postings
    IsActive             bool
    IsSystem             bool            // protected from deletion
    AllowManualEntries   bool
    RequireReference     bool
    RequireCostCentre    bool
    LockedCurrency       *string         // e.g. "USD" for a USD bank account
    ReportSection        string          // balance_sheet | profit_loss | cash_flow
    CreatedAt            time.Time
    UpdatedAt            time.Time
}

// Activate transitions a group account's child from setup to postable.
func (a *Account) Activate() error {
    if a.IsGroup {
        return ErrGroupAccountCannotPost
    }
    a.IsActive = true
    return nil
}

// Deactivate prevents future postings but preserves history.
func (a *Account) Deactivate(hasOpenBalance bool) error {
    if hasOpenBalance {
        return ErrCannotDeactivateWithBalance
    }
    a.IsActive = false
    return nil
}
```

### Materialised Path Queries

```sql
-- All leaf accounts under Current Assets (1100)
SELECT * FROM accounts
WHERE path LIKE '/1000/1100/%'
  AND is_group = FALSE
  AND tenant_id = current_tenant_id();

-- Subtree balance roll-up for a group account
SELECT
    SUM(ab.debit_balance) - SUM(ab.credit_balance) AS net_balance
FROM account_balances ab
JOIN accounts a ON a.id = ab.account_id
WHERE a.path LIKE '/1000/1100/%'
  AND ab.period_id = $1
  AND a.tenant_id = current_tenant_id();
```

### Account Balance Materialisation

Running balances are maintained in `account_balances` — one row per account per period. This avoids full-table SUM scans on the journal_lines table for every balance query.

```go
type AccountBalance struct {
    ID           uuid.UUID
    TenantID     TenantID
    AccountID    uuid.UUID
    PeriodID     uuid.UUID
    OpeningDebit decimal.Decimal
    OpeningCredit decimal.Decimal
    PeriodDebit  decimal.Decimal
    PeriodCredit decimal.Decimal
    ClosingDebit decimal.Decimal   // computed: opening + period
    ClosingCredit decimal.Decimal
    NetBalance   decimal.Decimal   // ClosingDebit - ClosingCredit
    UpdatedAt    time.Time
}
```

Balances are updated atomically inside the journal posting transaction:

```sql
-- Called inside the same transaction as INSERT INTO journal_lines
INSERT INTO account_balances (tenant_id, account_id, period_id, period_debit, period_credit)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (tenant_id, account_id, period_id)
DO UPDATE SET
    period_debit  = account_balances.period_debit  + EXCLUDED.period_debit,
    period_credit = account_balances.period_credit + EXCLUDED.period_credit,
    closing_debit = account_balances.opening_debit + account_balances.period_debit + EXCLUDED.period_debit,
    closing_credit = account_balances.opening_credit + account_balances.period_credit + EXCLUDED.period_credit,
    net_balance   = (account_balances.opening_debit + account_balances.period_debit + EXCLUDED.period_debit)
                  - (account_balances.opening_credit + account_balances.period_credit + EXCLUDED.period_credit),
    updated_at    = NOW();
```

---

## 5. Journal Entry Pipeline

The journal entry pipeline is the core of the Finance module. Every debit and credit — regardless of source — passes through this pipeline before reaching the GL. The pipeline is a sequence of stages; each stage receives the entry and either enriches it, validates it, or fails with a structured error.

### Entry Lifecycle

```
Draft ──► PendingApproval ──► Approved ──► Posted ──► Reversed (terminal)
  │              │                │
  │              └──► Rejected    └──► Cancelled
  └──► Cancelled
```

### Pipeline Stages

| # | Stage | Description | Blocking |
|---|---|---|---|
| 1 | `ValidateStructure` | ≥2 lines; each line has debit XOR credit; amounts positive | Yes |
| 2 | `ValidateBalance` | Total debits = total credits (to 2dp in functional currency) | Yes |
| 3 | `ValidateAccounts` | All accounts exist, active, leaf, match org; currency lock check | Yes |
| 4 | `ValidatePeriod` | Transaction date falls in an open period for this org | Yes |
| 5 | `ValidateReferences` | Accounts with `require_reference` have a reference on the line | Yes |
| 6 | `ValidateCostCentres` | Expense accounts with `require_cost_centre` have one assigned | Configurable |
| 7 | `CheckDuplicates` | Same reference + amount + accounts within 24h — warn or block | Configurable |
| 8 | `CheckBudget` | If budget control active, check each expense line against budget | Configurable |
| 9 | `CheckApproval` | If approval workflow required, verify approved status before post | Yes |
| 10 | `PostToGL` | Write journal_lines; update account_balances; update subsidiary ledgers | Yes |

### Stage Interface

```go
// domain/journal/pipeline.go

type Stage interface {
    Name()    string
    Execute(ctx context.Context, entry *JournalEntry, cfg PostingConfig) error
}

type Pipeline struct {
    stages []Stage
    cfg    PostingConfig
}

func (p *Pipeline) Run(ctx context.Context, entry *JournalEntry) error {
    for _, stage := range p.stages {
        if err := stage.Execute(ctx, entry, p.cfg); err != nil {
            return fmt.Errorf("pipeline stage %s: %w", stage.Name(), err)
        }
    }
    return nil
}
```

### ValidateBalance Stage

```go
// domain/journal/stages/validate_balance.go

type ValidateBalanceStage struct{}

func (s *ValidateBalanceStage) Execute(_ context.Context, entry *JournalEntry, _ PostingConfig) error {
    totalDebit  := decimal.Zero
    totalCredit := decimal.Zero

    for _, line := range entry.Lines {
        totalDebit  = totalDebit.Add(line.DebitAmount)
        totalCredit = totalCredit.Add(line.CreditAmount)
    }

    // Round to 2dp in functional currency before comparison
    diff := totalDebit.Sub(totalCredit).Abs().Round(2)
    if diff.IsPositive() {
        return ValidationError{
            Code:    "ENTRY_NOT_BALANCED",
            Message: fmt.Sprintf("total debits %s ≠ total credits %s (difference: %s)",
                totalDebit, totalCredit, diff),
            Field:   "lines",
        }
    }
    return nil
}
```

### PostToGL Stage

```go
// domain/journal/stages/post_to_gl.go

type PostToGLStage struct {
    db          *pgxpool.Pool
    lineRepo    JournalLineRepository
    balanceRepo AccountBalanceRepository
    subLedger   SubsidiaryLedger
    outbox      OutboxRepository
}

func (s *PostToGLStage) Execute(ctx context.Context, entry *JournalEntry, _ PostingConfig) error {
    return s.db.WithTransaction(ctx, func(tx pgx.Tx) error {
        // 1. Lock the accounting period to prevent concurrent close
        if err := s.lockPeriod(ctx, tx, entry.PeriodID); err != nil {
            return err
        }

        // 2. Insert journal lines
        for _, line := range entry.Lines {
            if err := s.lineRepo.Insert(ctx, tx, line); err != nil {
                return err
            }
        }

        // 3. Update account balances (upsert)
        if err := s.balanceRepo.ApplyLines(ctx, tx, entry.Lines); err != nil {
            return err
        }

        // 4. Update subsidiary ledgers (AR, AP, etc.)
        if err := s.subLedger.Apply(ctx, tx, entry); err != nil {
            return err
        }

        // 5. Mark entry as posted
        entry.Status    = StatusPosted
        entry.PostedAt  = ptr(time.Now())
        entry.PostedBy  = entry.ApprovedBy

        // 6. Write outbox event so other modules learn about the posting
        return s.outbox.Insert(ctx, tx, JournalPostedEvent{
            TenantID:   entry.TenantID,
            EntryID:    entry.ID,
            Reference:  entry.Reference,
            PostedAt:   *entry.PostedAt,
        })
    })
}
```

### Reversal

Reversals are mirror entries. The original entry is marked `reversed`; a new entry is created with all debits and credits swapped.

```go
// domain/journal/reversal.go

func (e *JournalEntry) CreateReversal(reversalDate time.Time, reason string, by uuid.UUID) (*JournalEntry, error) {
    if e.Status != StatusPosted {
        return nil, ErrOnlyPostedEntriesCanBeReversed
    }

    reversed := &JournalEntry{
        ID:              uuid.New(),
        TenantID:        e.TenantID,
        OrganisationID:  e.OrganisationID,
        TransactionDate: reversalDate,
        Reference:       "REV-" + e.Reference,
        Description:     fmt.Sprintf("Reversal of %s: %s", e.Reference, reason),
        Type:            TypeReversal,
        SourceModule:    "finance",
        ReversalOfID:    &e.ID,
        Status:          StatusDraft,
        CreatedBy:       by,
        CreatedAt:       time.Now(),
    }

    for _, line := range e.Lines {
        reversed.Lines = append(reversed.Lines, JournalLine{
            AccountID:    line.AccountID,
            DebitAmount:  line.CreditAmount,  // swap
            CreditAmount: line.DebitAmount,   // swap
            Description:  line.Description,
            CostCentreID: line.CostCentreID,
        })
    }

    e.Status = StatusReversed
    e.record(JournalEntryReversed{EntryID: e.ID, ReversalID: reversed.ID, At: time.Now()})

    return reversed, nil
}
```

---

## 6. General Ledger Engine

### GL Query Patterns

The GL never mutates data — it is a read-only view over `journal_lines` and `account_balances`. All mutations go through the journal pipeline.

```go
// domain/gl/ledger.go

type GeneralLedger struct {
    balanceRepo AccountBalanceRepository
    lineRepo    JournalLineRepository
}

// TrialBalance returns all account balances as of a date.
func (gl *GeneralLedger) TrialBalance(
    ctx    context.Context,
    orgID  uuid.UUID,
    asAt   time.Time,
    showZero bool,
) ([]TrialBalanceLine, error) {
    periodID, err := gl.balanceRepo.PeriodIDForDate(ctx, orgID, asAt)
    if err != nil { return nil, err }

    rows, err := gl.balanceRepo.GetAllForPeriod(ctx, orgID, periodID)
    if err != nil { return nil, err }

    lines := []TrialBalanceLine{}
    for _, row := range rows {
        if !showZero && row.NetBalance.IsZero() { continue }
        lines = append(lines, TrialBalanceLine{
            AccountCode:  row.AccountCode,
            AccountName:  row.AccountName,
            RootType:     row.RootType,
            DebitBalance: row.ClosingDebit,
            CreditBalance: row.ClosingCredit,
            NetBalance:   row.NetBalance,
            Level:        row.Level,
            IsGroup:      row.IsGroup,
            Path:         row.Path,
        })
    }
    return lines, nil
}

// AccountLedger returns a running-balance detail for one account over a date range.
func (gl *GeneralLedger) AccountLedger(
    ctx       context.Context,
    accountID uuid.UUID,
    from, to  time.Time,
) ([]LedgerEntry, error) {
    return gl.lineRepo.GetDetailWithRunningBalance(ctx, accountID, from, to)
}
```

### Financial Statement Construction

Financial statements are built from the trial balance using the account hierarchy and report-section mapping.

```go
type BalanceSheet struct {
    AsAt        time.Time
    Currency    string
    Assets      StatementSection
    Liabilities StatementSection
    Equity      StatementSection
    // Assets.Total must equal Liabilities.Total + Equity.Total
}

type ProfitLoss struct {
    From, To    time.Time
    Currency    string
    Revenue     StatementSection
    CostOfSales StatementSection
    GrossProfit decimal.Decimal
    OpExpenses  StatementSection
    EBIT        decimal.Decimal
    OtherItems  StatementSection
    PBT         decimal.Decimal
    TaxExpense  decimal.Decimal
    NetIncome   decimal.Decimal
}

type StatementSection struct {
    Lines []StatementLine
    Total decimal.Decimal
}

type StatementLine struct {
    AccountCode string
    AccountName string
    Amount      decimal.Decimal
    Level       int
    IsGroup     bool
    Children    []StatementLine
}
```

### Subsidiary Ledger

The subsidiary ledger reconciles customer (AR), supplier (AP), and inventory balances against their GL control accounts.

```go
// domain/gl/subsidiary.go

type SubsidiaryLedger struct {
    db *pgxpool.Pool
}

// Apply is called inside the PostToGL transaction.
// It updates the appropriate subsidiary based on the entry's source module and type.
func (s *SubsidiaryLedger) Apply(ctx context.Context, tx pgx.Tx, entry *JournalEntry) error {
    switch entry.SourceModule {
    case "sales":
        return s.applyAR(ctx, tx, entry)
    case "procurement":
        return s.applyAP(ctx, tx, entry)
    case "inventory":
        return s.applyInventory(ctx, tx, entry)
    case "hr":
        return s.applyEmployeeReceivable(ctx, tx, entry)
    }
    return nil // manual entries don't affect subsidiaries directly
}

// ReconcileAR verifies that the sum of all customer open balances
// equals the AR control account balance. Returns discrepancies.
func (s *SubsidiaryLedger) ReconcileAR(ctx context.Context, orgID uuid.UUID, asAt time.Time) (*ReconciliationResult, error) {
    controlBalance, err := s.getControlBalance(ctx, orgID, "receivable", asAt)
    if err != nil { return nil, err }

    subsidiaryTotal, err := s.getARSubsidiaryTotal(ctx, orgID, asAt)
    if err != nil { return nil, err }

    diff := controlBalance.Sub(subsidiaryTotal)
    return &ReconciliationResult{
        ControlBalance:    controlBalance,
        SubsidiaryTotal:   subsidiaryTotal,
        Difference:        diff,
        IsReconciled:      diff.IsZero(),
    }, nil
}
```

---

## 7. Period Management

### FiscalYear and AccountingPeriod

```go
// domain/period/fiscal_year.go

type FiscalYear struct {
    ID             uuid.UUID
    TenantID       TenantID
    OrganisationID uuid.UUID
    Name           string          // e.g. "FY2025"
    StartDate      time.Time
    EndDate        time.Time
    IsClosed       bool
    IsLocked       bool            // no transactions whatsoever
    Periods        []AccountingPeriod
    CreatedAt      time.Time
}

// domain/period/accounting_period.go

type PeriodStatus string
const (
    PeriodOpen       PeriodStatus = "open"
    PeriodSoftClosed PeriodStatus = "soft_closed"
    PeriodHardClosed PeriodStatus = "hard_closed"
    PeriodLocked     PeriodStatus = "locked"
)

type AccountingPeriod struct {
    ID             uuid.UUID
    TenantID       TenantID
    OrganisationID uuid.UUID
    FiscalYearID   uuid.UUID
    Name           string
    StartDate      time.Time
    EndDate        time.Time
    Status         PeriodStatus
    ClosedAt       *time.Time
    ClosedBy       *uuid.UUID
    LockedAt       *time.Time
    LockedBy       *uuid.UUID
}

func (p *AccountingPeriod) SoftClose(by uuid.UUID, now time.Time) error {
    if p.Status != PeriodOpen {
        return ErrInvalidTransition{From: p.Status, To: PeriodSoftClosed}
    }
    p.Status   = PeriodSoftClosed
    p.ClosedAt = &now
    p.ClosedBy = &by
    return nil
}

func (p *AccountingPeriod) HardClose(by uuid.UUID, now time.Time, checksPassed bool) error {
    if p.Status != PeriodSoftClosed {
        return ErrMustSoftCloseFirst
    }
    if !checksPassed {
        return ErrPeriodCloseChecksFailed
    }
    p.Status = PeriodHardClosed
    return nil
}

func (p *AccountingPeriod) Reopen(by uuid.UUID, reason string, now time.Time) error {
    if p.Status == PeriodLocked {
        return ErrLockedPeriodCannotReopen
    }
    p.Status   = PeriodOpen
    p.ClosedAt = nil
    p.ClosedBy = nil
    p.record(PeriodReopened{PeriodID: p.ID, By: by, Reason: reason, At: now})
    return nil
}
```

### Period Gate

The period gate is called at the start of the `ValidatePeriod` pipeline stage. It is also enforced by a PostgreSQL trigger as a second layer.

```go
// domain/period/gate.go

type PeriodGate struct {
    periodRepo PeriodRepository
}

type PostingPermission string
const (
    PermissionAny      PostingPermission = "any"      // open period
    PermissionFinance  PostingPermission = "finance"  // soft_closed — finance team only
    PermissionNone     PostingPermission = "none"     // hard_closed or locked
)

func (g *PeriodGate) CheckPosting(
    ctx    context.Context,
    orgID  uuid.UUID,
    txDate time.Time,
    role   string,
) (PostingPermission, error) {
    period, err := g.periodRepo.ForDate(ctx, orgID, txDate)
    if err != nil { return "", ErrNoPeriodForDate{Date: txDate} }

    switch period.Status {
    case PeriodOpen:
        return PermissionAny, nil
    case PeriodSoftClosed:
        if role == "finance_manager" || role == "cfo" {
            return PermissionFinance, nil
        }
        return PermissionNone, ErrPeriodSoftClosed{Period: period.Name}
    case PeriodHardClosed, PeriodLocked:
        return PermissionNone, ErrPeriodClosed{Period: period.Name, Status: string(period.Status)}
    }
    return PermissionNone, ErrUnknownPeriodStatus
}
```

### Period Close Checklist

Before a period can hard-close, the system verifies a set of mandatory checks. The Temporal `PeriodCloseWorkflow` coordinates this:

```go
type PeriodCloseCheck struct {
    Name      string
    Passed    bool
    Detail    string
    Blocking  bool
}

var mandatoryCloseChecks = []string{
    "bank_reconciliations_approved",      // all bank accounts have APPROVED rec for this period
    "ar_subsidiary_reconciled",           // AR control = AR subsidiary total
    "ap_subsidiary_reconciled",           // AP control = AP subsidiary total
    "fx_revaluation_completed",           // FX revaluation has run (if FC balances exist)
    "intercompany_balances_matched",      // all IC pairs net to zero
    "no_pending_approvals",               // no journal entries stuck in pending_approval
    "trial_balance_balanced",             // total debits = total credits
}
```

---

## 8. Bank Reconciliation Engine

### Workspace

A `BankReconciliation` is opened once per bank account per period. It is the workspace where statement lines are matched to GL entries.

```go
// domain/reconciliation/workspace.go

type ReconciliationStatus string
const (
    RecOpen        ReconciliationStatus = "open"
    RecInProgress  ReconciliationStatus = "in_progress"
    RecBalanced    ReconciliationStatus = "balanced"
    RecUnderReview ReconciliationStatus = "under_review"
    RecApproved    ReconciliationStatus = "approved"
    RecLocked      ReconciliationStatus = "locked"
)

type BankReconciliation struct {
    ID                    uuid.UUID
    TenantID              TenantID
    OrganisationID        uuid.UUID
    BankAccountID         uuid.UUID
    PeriodID              uuid.UUID
    StatementOpeningBal   decimal.Decimal
    StatementClosingBal   decimal.Decimal
    GLClosingBalance      decimal.Decimal    // from account_balances at period end
    AdjustedBankBalance   decimal.Decimal    // bank + deposits-in-transit - outstanding cheques - errors
    AdjustedGLBalance     decimal.Decimal    // GL + adjustments posted during rec
    Difference            decimal.Decimal    // must be zero before approval
    Status                ReconciliationStatus
    PreparedBy            uuid.UUID
    ApprovedBy            *uuid.UUID
    LockedAt              *time.Time
    Matches               []ReconciliationMatch
    CreatedAt             time.Time
    UpdatedAt             time.Time
}

func (r *BankReconciliation) IsBalanced() bool {
    return r.Difference.Abs().LessThanOrEqual(decimal.NewFromFloat(0.01))
}

func (r *BankReconciliation) Submit() error {
    if !r.IsBalanced() {
        return ErrReconciliationNotBalanced{Difference: r.Difference}
    }
    r.Status = RecUnderReview
    return nil
}

func (r *BankReconciliation) Approve(by uuid.UUID) error {
    if r.Status != RecUnderReview { return ErrInvalidTransition{} }
    if r.PreparedBy == by        { return ErrSelfApprovalProhibited }
    r.ApprovedBy = &by
    r.Status     = RecApproved
    return nil
}
```

### Auto-Matcher

The matcher evaluates statement lines against GL entries in priority order. The first matching rule that fires wins.

```go
// domain/reconciliation/matcher.go

type MatchRule interface {
    Name()       string
    Confidence() MatchConfidence    // auto | suggested | manual
    Match(line StatementLine, glEntries []GLEntry) *MatchCandidate
}

type AutoMatcher struct {
    rules []MatchRule    // ordered: highest confidence first
}

func (m *AutoMatcher) Run(
    lines    []StatementLine,
    glEntries []GLEntry,
) []MatchResult {
    results := []MatchResult{}
    usedGLIDs := map[uuid.UUID]bool{}

    for _, line := range lines {
        available := filterUnused(glEntries, usedGLIDs)
        var best *MatchCandidate

        for _, rule := range m.rules {
            if candidate := rule.Match(line, available); candidate != nil {
                best = candidate
                break // first rule that matches wins
            }
        }

        result := MatchResult{StatementLine: line}
        if best != nil {
            result.Candidate   = best
            result.Confidence  = best.Confidence
            for _, id := range best.GLEntryIDs {
                usedGLIDs[id] = true
            }
        }
        results = append(results, result)
    }
    return results
}
```

### Matching Rules

```go
// domain/reconciliation/rules/exact_match.go

type ExactMatchRule struct{}

func (r *ExactMatchRule) Name() string { return "M-01 Exact Match" }
func (r *ExactMatchRule) Confidence() MatchConfidence { return ConfidenceAuto }

func (r *ExactMatchRule) Match(line StatementLine, entries []GLEntry) *MatchCandidate {
    for _, e := range entries {
        if e.Amount.Equal(line.Amount) &&
            normalise(e.Reference) == normalise(line.Reference) &&
            daysApart(e.PostingDate, line.Date) == 0 {
            return &MatchCandidate{
                GLEntryIDs: []uuid.UUID{e.ID},
                Confidence: ConfidenceAuto,
                RuleName:   r.Name(),
            }
        }
    }
    return nil
}

// domain/reconciliation/rules/batch_match.go — M-04: one bank line = sum of multiple GL entries

type BatchMatchRule struct{ Tolerance decimal.Decimal }

func (r *BatchMatchRule) Match(line StatementLine, entries []GLEntry) *MatchCandidate {
    // Find all subsets of entries that sum to line.Amount within tolerance
    // Implemented as a bounded knapsack (max 10 entries to prevent combinatorial explosion)
    candidates := findSubsets(entries, line.Amount, r.Tolerance, 10)
    if len(candidates) == 1 {
        return &MatchCandidate{
            GLEntryIDs: candidates[0],
            Confidence: ConfidenceSuggested,
            RuleName:   "M-04 Batch Match",
        }
    }
    return nil
}
```

### Statement Importers

Each bank statement format has its own importer implementing the `StatementImporter` interface:

```go
type StatementImporter interface {
    Format() string
    Parse(r io.Reader, mapping ColumnMap) ([]StatementLine, StatementSummary, error)
}

// Supported formats: OFX/QFX, MT940 (SWIFT), CSV (configurable column mapping), Excel
```

---

## 9. Accounts Receivable & Payable

### AR Sub-Ledger

The AR sub-ledger is a view over open invoice lines. Each customer has a running balance maintained by the subsidiary ledger when journal entries are posted.

```go
type CustomerBalance struct {
    TenantID       TenantID
    OrganisationID uuid.UUID
    CustomerID     uuid.UUID
    CustomerName   string
    Currency       string
    OpenBalance    decimal.Decimal    // in functional currency
    OpenBalanceFC  decimal.Decimal    // in transaction currency
    OldestDueDate  *time.Time
}

type ARAgingBucket struct {
    CustomerID uuid.UUID
    CustomerName string
    Current    decimal.Decimal   // not yet due
    Days30     decimal.Decimal   // 1–30 days overdue
    Days60     decimal.Decimal   // 31–60 days
    Days90     decimal.Decimal   // 61–90 days
    Over90     decimal.Decimal   // 91+ days
    Total      decimal.Decimal
}
```

### AR Aging Query

```sql
-- v_ar_aging (security_invoker = true)
SELECT
    s.tenant_id,
    s.customer_id,
    c.name                                                  AS customer_name,
    SUM(s.functional_amount) FILTER (WHERE s.days_overdue <= 0)   AS current_amount,
    SUM(s.functional_amount) FILTER (WHERE s.days_overdue BETWEEN 1 AND 30)  AS days_30,
    SUM(s.functional_amount) FILTER (WHERE s.days_overdue BETWEEN 31 AND 60) AS days_60,
    SUM(s.functional_amount) FILTER (WHERE s.days_overdue BETWEEN 61 AND 90) AS days_90,
    SUM(s.functional_amount) FILTER (WHERE s.days_overdue > 90)  AS over_90,
    SUM(s.functional_amount)                                AS total
FROM ar_subsidiary_entries s
JOIN customers c ON c.id = s.customer_id
WHERE s.tenant_id  = current_tenant_id()
  AND s.is_settled = FALSE
  AND s.as_at_date = $1
GROUP BY s.tenant_id, s.customer_id, c.name;
```

### Payment Run

```go
// domain/payment/run.go

type PaymentRunStatus string
const (
    PRDraft    PaymentRunStatus = "draft"
    PRApproved PaymentRunStatus = "approved"
    PRExported PaymentRunStatus = "exported"    // bank file generated
    PRPosted   PaymentRunStatus = "posted"      // GL entries confirmed
    PRCancelled PaymentRunStatus = "cancelled"
)

type PaymentRun struct {
    ID              uuid.UUID
    TenantID        TenantID
    OrganisationID  uuid.UUID
    DueOnOrBefore   time.Time
    BankAccountID   uuid.UUID
    Currency        string
    Status          PaymentRunStatus
    TotalAmount     decimal.Decimal
    ItemCount       int
    Items           []PaymentRunItem
    ExportedAt      *time.Time
    ExportFormat    string              // equity_eft | kcb_rtgs | swift_mt101 | generic_csv
    ApprovedBy      *uuid.UUID
    PostedAt        *time.Time
    CreatedBy       uuid.UUID
    CreatedAt       time.Time
}

type PaymentRunItem struct {
    InvoiceID       uuid.UUID
    SupplierID      uuid.UUID
    Amount          decimal.Decimal
    Currency        string
    BankAccount     SupplierBankAccount
    Reference       string
    Included        bool
}
```

---

## 10. Cost Centre & Budget Management

### Cost Centre Aggregate

```go
// domain/cost_centre/cost_centre.go

type CostCentre struct {
    ID               uuid.UUID
    TenantID         TenantID
    OrganisationID   uuid.UUID
    Code             string
    Name             string
    ParentID         *uuid.UUID
    Path             string
    Level            int
    IsGroup          bool
    IsDistributed    bool
    AllocationMethod AllocationMethod   // percentage | headcount | square_footage | usage
    Allocations      []CostAllocation   // only when IsDistributed = true
    IsActive         bool
    CreatedAt        time.Time
}

type CostAllocation struct {
    TargetCentreID uuid.UUID
    Percentage     decimal.Decimal   // must sum to 100 across all allocations
}

// Validate ensures allocations sum to exactly 100%.
func (cc *CostCentre) ValidateAllocations() error {
    if !cc.IsDistributed { return nil }
    total := decimal.Zero
    for _, a := range cc.Allocations {
        total = total.Add(a.Percentage)
    }
    if !total.Equal(decimal.NewFromInt(100)) {
        return ErrAllocationsMustSumTo100{Got: total}
    }
    return nil
}
```

### Budget Engine

```go
// domain/budget/budget.go

type BudgetStatus string
const (
    BudgetDraft   BudgetStatus = "draft"
    BudgetApproved BudgetStatus = "approved"
    BudgetActive  BudgetStatus = "active"     // exactly one per fiscal year
    BudgetArchived BudgetStatus = "archived"
)

type Budget struct {
    ID             uuid.UUID
    TenantID       TenantID
    OrganisationID uuid.UUID
    FiscalYearID   uuid.UUID
    Name           string
    VersionLabel   string     // "V1", "V2 Q2 Revision"
    Status         BudgetStatus
    Currency       string
    Lines          []BudgetLine
    ApprovedBy     *uuid.UUID
    ApprovedAt     *time.Time
    CreatedAt      time.Time
}

type BudgetLine struct {
    AccountID     uuid.UUID
    CostCentreID  *uuid.UUID
    PeriodID      uuid.UUID
    Amount        decimal.Decimal
}
```

### Budget Checker

```go
// domain/budget/checker.go

type BudgetControlMode string
const (
    BudgetNone BudgetControlMode = "none"
    BudgetSoft BudgetControlMode = "soft"    // warn but allow override
    BudgetHard BudgetControlMode = "hard"    // block
)

type BudgetCheckResult struct {
    AccountID      uuid.UUID
    Budgeted       decimal.Decimal
    ActualToDate   decimal.Decimal
    ThisEntry      decimal.Decimal
    Remaining      decimal.Decimal
    WouldExceedBy  decimal.Decimal
    IsOverBudget   bool
}

func (c *BudgetChecker) Check(
    ctx       context.Context,
    orgID     uuid.UUID,
    periodID  uuid.UUID,
    lines     []JournalLine,
) ([]BudgetCheckResult, error) {
    results := []BudgetCheckResult{}
    for _, line := range lines {
        if !isExpenseLine(line) { continue }
        result, err := c.checkLine(ctx, orgID, periodID, line)
        if err != nil { return nil, err }
        results = append(results, result)
    }
    return results, nil
}
```

---

## 11. Cash Management & Forecasting

### Cash Position

```go
// domain/cashflow/position.go

type CashPosition struct {
    AsAt          time.Time
    Currency      string           // functional currency
    Accounts      []CashAccount
    TotalCash     decimal.Decimal  // sum of all accounts in functional currency
    Committed     []CommittedItem  // payment runs approved but not yet posted
    AvailableCash decimal.Decimal  // TotalCash - sum(Committed)
}

type CashAccount struct {
    AccountCode       string
    AccountName       string
    Currency          string
    Balance           decimal.Decimal  // in account currency
    FunctionalBalance decimal.Decimal  // converted at today's rate
    ExchangeRate      decimal.Decimal
    RateDate          time.Time
}
```

### 13-Week Rolling Forecast

```go
// domain/cashflow/forecast.go

type WeeklyForecast struct {
    WeekNumber  int
    DateFrom    time.Time
    DateTo      time.Time
    Receipts    decimal.Decimal   // from AR aging expected dates + confirmed orders
    Payments    decimal.Decimal   // from AP aging due dates + payroll + tax calendar
    Net         decimal.Decimal
    ClosingBal  decimal.Decimal
}

type CashFlowForecast struct {
    AsAt          time.Time
    Weeks         []WeeklyForecast
    OpeningBal    decimal.Decimal
    MinBalance    decimal.Decimal
    MaxBalance    decimal.Decimal
    TargetMinBal  decimal.Decimal   // from config: minimum operating cash policy
    DeficitWeeks  []int             // weeks where closing < target minimum
}
```

---

## 12. Intercompany Module

### IC Transaction

```go
// domain/intercompany/transaction.go

type ICTransactionType string
const (
    ICManagementFee ICTransactionType = "management_fee"
    ICLoan          ICTransactionType = "loan"
    ICGoods         ICTransactionType = "goods_transfer"
    ICServices      ICTransactionType = "services"
    ICDividend      ICTransactionType = "dividend"
)

type IntercompanyTransaction struct {
    ID                  uuid.UUID
    TenantID            TenantID
    SourceOrgID         uuid.UUID
    TargetOrgID         uuid.UUID
    Type                ICTransactionType
    Amount              decimal.Decimal
    Currency            string
    TransactionDate     time.Time
    SourceJournalID     uuid.UUID    // journal entry in source org
    TargetJournalID     *uuid.UUID   // journal entry in target org (nil until recorded)
    ICProfit            *decimal.Decimal  // for goods transfers
    IsEliminated        bool
    Status              ICStatus     // pending | matched | eliminated
}
```

### Period-End IC Matching

```go
// domain/intercompany/matcher.go

type ICMatchResult struct {
    EntityPair      string          // "ORG-A ↔ ORG-B"
    SourceBalance   decimal.Decimal
    TargetBalance   decimal.Decimal
    Difference      decimal.Decimal
    Currency        string
    IsMatched       bool
    DifferenceType  string         // "fx_movement" | "timing" | "error"
}

func (m *ICMatcher) MatchAll(ctx context.Context, tenantID TenantID, asAt time.Time) ([]ICMatchResult, error) {
    pairs, err := m.getPairs(ctx, tenantID)
    if err != nil { return nil, err }

    results := []ICMatchResult{}
    for _, pair := range pairs {
        result, err := m.matchPair(ctx, pair, asAt)
        if err != nil { return nil, err }
        results = append(results, result)
    }
    return results, nil
}
```

---

## 13. Financial Reporting Engine

### Report Definition

All financial reports are defined as JSONB report definitions stored in `report_definitions`. This is the same 7-layer report builder referenced in earlier architecture sessions.

```go
type ReportDefinition struct {
    ID          uuid.UUID
    TenantID    TenantID
    Code        string         // "balance_sheet", "profit_loss", "trial_balance"
    Name        string
    IsSystem    bool           // system reports cannot be deleted
    DataSource  string         // PostgreSQL view name on the read replica
    Parameters  []ReportParam  // filterable parameters exposed to user
    Columns     []ReportColumn
    Grouping    []string
    Sorting     []SortSpec
    OutputFormats []string     // pdf | excel | csv | json
}
```

### View Catalogue (read replica)

| View | Purpose |
|---|---|
| `v_trial_balance` | All accounts with debit/credit/net for a period |
| `v_balance_sheet` | Hierarchical balance sheet with section grouping |
| `v_profit_loss` | P&L with gross profit, EBIT, net income calculations |
| `v_cash_flow_indirect` | Cash flow statement (indirect method) |
| `v_gl_detail` | Transaction-level ledger for one account over a date range |
| `v_ar_aging` | AR aging by customer with configurable day buckets |
| `v_ap_aging` | AP aging by supplier |
| `v_budget_vs_actual` | Budget lines joined to actual account_balances |
| `v_cost_centre_pl` | P&L scoped to a cost centre subtree |
| `v_bank_rec_summary` | Bank reconciliation status per account per period |
| `v_cash_position` | Real-time cash across all bank accounts |
| `v_cash_forecast` | 13-week rolling forecast |
| `v_ic_balances` | Intercompany balances per entity pair |
| `v_period_close_status` | Checklist status for the current open period |

All views carry `security_invoker = true`. `readonly_role` can query all views; `application_role` can query all views. `admin_role` bypasses RLS via the `BYPASSRLS` attribute.

---

## 14. GL Integration Contract

The Finance module is the consumer of events from all other modules. It owns the account mapping configuration that translates semantic codes (e.g. `BASIC_SALARY`, `FUEL_REVENUE`) to actual GL account IDs.

### Account Mapping

```go
type AccountMapping struct {
    ID               uuid.UUID
    TenantID         TenantID
    OrganisationID   uuid.UUID
    SemanticCode     string      // e.g. "BASIC_SALARY", "VAT_OUTPUT", "EMPLOYER_NSSF"
    AccountID        uuid.UUID
    CostCentreID     *uuid.UUID  // optional default cost centre
    EffectiveFrom    time.Time
    EffectiveTo      *time.Time
}
```

### Events Consumed

| Source Module | Event | Finance Action |
|---|---|---|
| HR | `PayrollPosted` | Resolve account mapping; post salary/statutory journal |
| HR | `PayrollReversed` | Post mirror reversal journal |
| HR | `DiscrepancyConfirmed` | Post DR Employee Receivable, CR Cash Shortage Expense |
| HR | `DiscrepancyWaived` | Post write-off journal |
| HR | `TerminalBenefitsComputed` | Post terminal benefit journal |
| Sales | `SaleInvoicePosted` | Post DR AR, CR Revenue, CR VAT Output |
| Sales | `PaymentReceived` | Post DR Bank, CR AR |
| Sales | `CreditNotePosted` | Post reversal of original sale journal |
| Procurement | `PurchaseInvoicePosted` | Post DR Inventory/Expense + DR VAT Input, CR AP |
| Procurement | `SupplierPaymentPosted` | Post DR AP, CR Bank |
| Inventory | `StockAdjustmentPosted` | Post DR/CR Inventory, CR/DR Shrinkage/Gain |
| Pricing | `RevenueRecognised` | Post DR Deferred Revenue, CR Revenue |
| Pricing | `DiscountPosted` | Post DR Discount Expense |
| Currency | `FXRevaluationPosted` | Post unrealised FX gain/loss journal |
| Currency | `FXGainLossRealised` | Post realised FX gain/loss on settlement |
| Currency | `HedgeMTMUpdated` | Post derivative asset/liability MTM journal |

### Event Consumer Pattern

```go
// application/event_consumers/hr_consumer.go

type HRConsumer struct {
    consumer     IdempotentConsumer
    mappingRepo  AccountMappingRepository
    journalSvc   JournalService
}

func (c *HRConsumer) HandlePayrollPosted(ctx context.Context, event IntegrationEvent) error {
    var payload PayrollPostedPayload
    if err := json.Unmarshal(event.Payload, &payload); err != nil {
        return fmt.Errorf("unmarshal PayrollPosted: %w", err)
    }

    return c.consumer.Handle(ctx, event, func(ctx context.Context, tx pgx.Tx) error {
        // 1. Resolve semantic codes → GL account IDs
        mapping, err := c.mappingRepo.ResolveAll(ctx, tx, event.TenantID, payload.SemanticCodes())
        if err != nil { return err }

        // 2. Build journal entry from the payload's JournalInstructions
        entry, err := buildPayrollJournalEntry(payload, mapping)
        if err != nil { return err }

        // 3. Post through the journal pipeline
        return c.journalSvc.PostIntegrationEntry(ctx, tx, entry)
    })
}
```

### Events Emitted

| Event | Trigger | Consumers |
|---|---|---|
| `finance.journal.posted` | Journal entry reaches POSTED status | HR (confirms GL posting), Sales, Procurement |
| `finance.journal.reversed` | Reversal entry created | All modules that emitted the original |
| `finance.period.hard_closed` | Period hard-closes | Reporting systems, audit |
| `finance.period.locked` | Period locked | Archive systems |
| `finance.reconciliation.approved` | Bank rec approved | Treasury, audit |
| `finance.payment_run.posted` | Payment run GL entries posted | HR (payroll cleared), Procurement |
| `finance.budget.activated` | New budget version activated | Cost centre owners, reporting |

---

## 15. API Reference

All endpoints require `Authorization: Bearer <token>` and `X-Tenant: <slug>`.

### Chart of Accounts

```
GET    /finance/accounts                              List accounts (filterable)
POST   /finance/accounts                              Create account
GET    /finance/accounts/{code}                       Get account + balance + recent activity
PATCH  /finance/accounts/{code}                       Update mutable fields
PATCH  /finance/accounts/{code}/deactivate            Deactivate (checks open balance)
GET    /finance/accounts/{code}/ledger                GL detail for date range
POST   /finance/accounts/import                       Bulk COA import from CSV
```

### Journal Entries

```
GET    /finance/journal-entries                       List (filterable by status, date, account)
POST   /finance/journal-entries                       Create draft entry
GET    /finance/journal-entries/{id}                  Get entry with lines
POST   /finance/journal-entries/{id}/submit           Draft → PendingApproval
POST   /finance/journal-entries/{id}/approve          PendingApproval → Approved
POST   /finance/journal-entries/{id}/post             Approved → Posted
POST   /finance/journal-entries/{id}/reverse          Posted → create reversal draft
POST   /finance/journal-entries/{id}/cancel           Draft/Approved → Cancelled
POST   /finance/journal-entries/import                Bulk import from CSV
```

#### `POST /finance/journal-entries` Request

```json
{
  "organisation_id": "uuid",
  "transaction_date": "2025-01-31",
  "reference": "CHQ-001234",
  "description": "Office rent — January 2025",
  "currency": "KES",
  "cost_centre_id": "uuid",
  "lines": [
    { "account_code": "7310", "debit": 50000.00, "credit": 0, "description": "January rent" },
    { "account_code": "1120", "debit": 0, "credit": 50000.00 }
  ]
}
```

**Response** `201 Created` — full entry with `reference_number` (auto-generated: `JE-2025-0145`)

### Periods

```
GET    /finance/fiscal-years                          List fiscal years
POST   /finance/fiscal-years                          Create fiscal year (generates 12 periods)
GET    /finance/periods                               List periods for current fiscal year
GET    /finance/periods/{id}                          Get period with close-checklist status
POST   /finance/periods/{id}/soft-close               Open → SoftClosed
POST   /finance/periods/{id}/hard-close               SoftClosed → HardClosed (runs checklist)
POST   /finance/periods/{id}/reopen                   Any closed → Open (logged, reason required)
POST   /finance/periods/{id}/lock                     HardClosed → Locked (CFO only)
```

### Bank Reconciliation

```
GET    /finance/reconciliations                       List (by account, period)
POST   /finance/reconciliations                       Open new reconciliation
POST   /finance/reconciliations/{id}/import           Import bank statement (multipart)
GET    /finance/reconciliations/{id}/matches          Auto-match suggestions
POST   /finance/reconciliations/{id}/matches          Confirm matches / create GL adjustments
POST   /finance/reconciliations/{id}/submit           InProgress → UnderReview
POST   /finance/reconciliations/{id}/approve          UnderReview → Approved (+ lock)
```

### Budgets

```
GET    /finance/budgets                               List budget versions
POST   /finance/budgets                               Create draft budget
POST   /finance/budgets/{id}/lines                    Set budget lines
POST   /finance/budgets/{id}/import                   Bulk import lines from CSV
POST   /finance/budgets/{id}/activate                 Draft/Approved → Active
GET    /finance/budgets/{id}/variance                 Budget vs. actual report
```

### Payment Runs

```
GET    /finance/payment-runs                          List
POST   /finance/payment-runs                          Create from due AP invoices
POST   /finance/payment-runs/{id}/approve             Draft → Approved
POST   /finance/payment-runs/{id}/export              Generate bank payment file
POST   /finance/payment-runs/{id}/confirm             Mark payments received; post GL
POST   /finance/payment-runs/{id}/cancel              Cancel before export
```

### Reports

```
POST   /finance/reports/trial-balance/run
POST   /finance/reports/balance-sheet/run
POST   /finance/reports/profit-loss/run
POST   /finance/reports/cash-flow/run
POST   /finance/reports/ar-aging/run
POST   /finance/reports/ap-aging/run
POST   /finance/reports/gl-detail/run
POST   /finance/reports/budget-variance/run
POST   /finance/reports/cost-centre-pl/run
POST   /finance/reports/cash-position/run
POST   /finance/reports/cash-forecast/run
POST   /finance/reports/ic-balances/run
GET    /finance/reports/period-close-status           Real-time checklist for open period
```

All report endpoints accept `{ "async": true }` to run in the background and return an `execution_id` for polling.

---

## 16. Business Rules & Validation

### Journal Entry Rules

| Rule ID | Rule | Severity |
|---|---|---|
| `FIN-JNL-001` | Total debits must equal total credits (to 2 decimal places in functional currency) | Error |
| `FIN-JNL-002` | Minimum 2 lines per entry; at least one debit and one credit | Error |
| `FIN-JNL-003` | Each line must have debit XOR credit — not both, not neither | Error |
| `FIN-JNL-004` | All line amounts must be positive | Error |
| `FIN-JNL-005` | All accounts must exist, be active, be leaf accounts, and belong to the same organisation | Error |
| `FIN-JNL-006` | Transaction date must fall in an open accounting period | Error |
| `FIN-JNL-007` | Accounts with `allow_manual_entries = false` cannot appear in MANUAL-type entries | Error |
| `FIN-JNL-008` | Accounts with `require_reference = true` must have a reference on that line | Error |
| `FIN-JNL-009` | Accounts with a locked currency must use that currency | Error |
| `FIN-JNL-010` | A POSTED entry cannot be edited — only reversed | Error |
| `FIN-JNL-011` | Self-approval is prohibited — the approver cannot be the submitter | Error |
| `FIN-JNL-012` | Duplicate detection: same reference + amount + accounts within 24 hours triggers a warning | Configurable |
| `FIN-JNL-013` | Integration entries (source_module ≠ "manual") bypass `allow_manual_entries` restriction | Auto |

### Period Rules

| Rule ID | Rule | Severity |
|---|---|---|
| `FIN-PRD-001` | Transaction date cannot be in the future | Error |
| `FIN-PRD-002` | Transaction date must fall within a valid fiscal year | Error |
| `FIN-PRD-003` | A SoftClosed period only accepts entries from `finance_manager` or `cfo` roles | Error |
| `FIN-PRD-004` | A HardClosed or Locked period accepts no entries whatsoever | Error |
| `FIN-PRD-005` | A Locked period cannot be reopened (requires CFO + CEO dual authorisation) | Error |
| `FIN-PRD-006` | A period cannot hard-close unless all mandatory checklist items pass | Error |
| `FIN-PRD-007` | Reopening a period must record the reason and the authoriser | Error |

### Bank Reconciliation Rules

| Rule ID | Rule | Severity |
|---|---|---|
| `FIN-REC-001` | Only one reconciliation per bank account per period | Error |
| `FIN-REC-002` | Opening balance must equal the prior period's closing balance | Error |
| `FIN-REC-003` | Reconciliation cannot be submitted unless `Difference` ≤ 0.01 | Error |
| `FIN-REC-004` | Preparer cannot be the approver | Error |
| `FIN-REC-005` | A period cannot hard-close unless all bank accounts have an APPROVED reconciliation | Error |
| `FIN-REC-006` | Once locked, a reconciliation record is immutable | Error |
| `FIN-REC-007` | Outstanding items older than 60 days are auto-flagged | Warning |

### Budget Rules

| Rule ID | Rule | Severity |
|---|---|---|
| `FIN-BDG-001` | Only one ACTIVE budget version per organisation per fiscal year | Error |
| `FIN-BDG-002` | Soft budget: expense exceeding budget triggers warning with mandatory override comment | Configurable |
| `FIN-BDG-003` | Hard budget: expense exceeding budget is blocked; requires CFO to approve amendment | Configurable |
| `FIN-BDG-004` | Budget checker only applies to expense accounts, not revenue or balance sheet | Auto |
| `FIN-BDG-005` | Budget version requires `finance.budgets.approve` permission to activate | Error |

### Account Rules

| Rule ID | Rule | Severity |
|---|---|---|
| `FIN-ACT-001` | Account code is immutable once any transaction exists against it | Error |
| `FIN-ACT-002` | Root type is immutable once any transaction exists | Error |
| `FIN-ACT-003` | A group account (is_group = true) cannot be deactivated while it has active children | Error |
| `FIN-ACT-004` | An account with an open balance cannot be deactivated | Error |
| `FIN-ACT-005` | A system account cannot be deleted | Error |
| `FIN-ACT-006` | Account hierarchy must not create cycles | Error |

---

## 17. Authorization Model

### Permission Keys

```
finance.accounts.read
finance.accounts.create
finance.accounts.update
finance.accounts.deactivate

finance.transactions.read
finance.transactions.create
finance.transactions.submit
finance.transactions.approve
finance.transactions.post
finance.transactions.reverse
finance.transactions.void

finance.periods.read
finance.periods.close
finance.periods.reopen
finance.periods.lock

finance.fiscal-years.read
finance.fiscal-years.create
finance.fiscal-years.close
finance.fiscal-years.lock

finance.currencies.read
finance.currencies.create
finance.currencies.update
finance.currencies.load-rates

finance.cost-centres.read
finance.cost-centres.create
finance.cost-centres.update
finance.cost-centres.deactivate

finance.budgets.read
finance.budgets.create
finance.budgets.update
finance.budgets.approve
finance.budgets.activate

finance.bank-accounts.read
finance.bank-accounts.create
finance.bank-accounts.update

finance.reconciliation.read
finance.reconciliation.create
finance.reconciliation.match
finance.reconciliation.submit
finance.reconciliation.approve
finance.reconciliation.lock

finance.payments.read
finance.payments.create
finance.payments.approve
finance.payments.post
finance.payments.cancel

finance.petty-cash.read
finance.petty-cash.disburse
finance.petty-cash.replenish

finance.tax-config.read
finance.tax-config.create
finance.tax-config.update

finance.intercompany.read
finance.intercompany.create
finance.intercompany.match
finance.intercompany.eliminate

finance.reports.read
finance.reports.run
finance.reports.export
finance.reports.schedule

finance.settings.read
finance.settings.update
```

### Default Roles

| Role | Key Permissions | Amount Limit |
|---|---|---|
| `report_viewer` | `finance.accounts.read`, `finance.reports.*` | — |
| `finance_viewer` | All `*.read` | — |
| `ap_clerk` | `transactions.create/submit`, `payments.create`, `petty-cash.*` | — |
| `ar_clerk` | `transactions.create/submit`, `reconciliation.create/match` | — |
| `finance_manager` | All finance operations except lock and unlimited approval | 500,000 KES |
| `cfo` | All finance operations including lock, fiscal year management, no amount limit | Unlimited |
| `auditor` | All `*.read` — time-limited with `WithExpiry()` | — |
| `budget_owner` | `budgets.read/create (draft only)`, `reports.read/run (cost centre scoped)` | — |

### SOD Enforcement

| Prohibited Combination | Reason |
|---|---|
| `ap_clerk` + `finance_manager` | Cannot approve own AP invoices |
| `ar_clerk` + `finance_manager` | Cannot approve own AR write-offs |
| Reconciliation preparer + reconciliation approver | Cannot self-approve bank rec |
| Payment creator + payment approver | Cannot self-approve payment run |
| Budget creator + budget approver | Cannot self-approve budget |

---

## 18. Database Schema & RLS

```sql
-- ── accounts ──────────────────────────────────────────────────────────────────
CREATE TABLE accounts (
    tenant_id             UUID        NOT NULL REFERENCES tenants(id),
    id                    UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by            UUID,
    updated_by            UUID,

    organisation_id       UUID        NOT NULL,
    code                  TEXT        NOT NULL,
    name                  TEXT        NOT NULL,
    root_type             TEXT        NOT NULL
                              CHECK (root_type IN ('asset','liability','equity','revenue','expense')),
    account_type          TEXT        NOT NULL,
    normal_balance        TEXT        NOT NULL CHECK (normal_balance IN ('debit','credit')),
    parent_id             UUID        REFERENCES accounts(id),
    path                  TEXT        NOT NULL,         -- materialised path
    level                 INT         NOT NULL DEFAULT 1,
    is_group              BOOLEAN     NOT NULL DEFAULT FALSE,
    is_active             BOOLEAN     NOT NULL DEFAULT TRUE,
    is_system             BOOLEAN     NOT NULL DEFAULT FALSE,
    allow_manual_entries  BOOLEAN     NOT NULL DEFAULT TRUE,
    require_reference     BOOLEAN     NOT NULL DEFAULT FALSE,
    require_cost_centre   BOOLEAN     NOT NULL DEFAULT FALSE,
    locked_currency       CHAR(3),
    report_section        TEXT        NOT NULL DEFAULT 'balance_sheet'
                              CHECK (report_section IN ('balance_sheet','profit_loss','cash_flow')),

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, organisation_id, code)
);

CREATE INDEX idx_accounts_tenant      ON accounts (tenant_id);
CREATE INDEX idx_accounts_path        ON accounts (tenant_id, path text_pattern_ops);
CREATE INDEX idx_accounts_parent      ON accounts (tenant_id, parent_id) WHERE parent_id IS NOT NULL;

ALTER TABLE accounts ENABLE ROW LEVEL SECURITY;
ALTER TABLE accounts FORCE  ROW LEVEL SECURITY;

CREATE POLICY accounts_app_all ON accounts FOR ALL TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY accounts_ro_select ON accounts FOR SELECT TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);


-- ── account_balances ──────────────────────────────────────────────────────────
CREATE TABLE account_balances (
    tenant_id             UUID        NOT NULL REFERENCES tenants(id),
    id                    UUID        NOT NULL DEFAULT gen_random_uuid(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    account_id            UUID        NOT NULL REFERENCES accounts(id),
    period_id             UUID        NOT NULL,
    opening_debit         NUMERIC(18,4) NOT NULL DEFAULT 0,
    opening_credit        NUMERIC(18,4) NOT NULL DEFAULT 0,
    period_debit          NUMERIC(18,4) NOT NULL DEFAULT 0,
    period_credit         NUMERIC(18,4) NOT NULL DEFAULT 0,
    closing_debit         NUMERIC(18,4) GENERATED ALWAYS AS (opening_debit + period_debit) STORED,
    closing_credit        NUMERIC(18,4) GENERATED ALWAYS AS (opening_credit + period_credit) STORED,
    net_balance           NUMERIC(18,4) GENERATED ALWAYS AS
                              ((opening_debit + period_debit) - (opening_credit + period_credit)) STORED,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, account_id, period_id)
);

CREATE INDEX idx_account_balances_tenant  ON account_balances (tenant_id);
CREATE INDEX idx_account_balances_period  ON account_balances (tenant_id, period_id);
CREATE INDEX idx_account_balances_account ON account_balances (tenant_id, account_id);

ALTER TABLE account_balances ENABLE ROW LEVEL SECURITY;
ALTER TABLE account_balances FORCE  ROW LEVEL SECURITY;

CREATE POLICY account_balances_app_all ON account_balances FOR ALL TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY account_balances_ro_select ON account_balances FOR SELECT TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);


-- ── fiscal_years ──────────────────────────────────────────────────────────────
CREATE TABLE fiscal_years (
    tenant_id             UUID        NOT NULL REFERENCES tenants(id),
    id                    UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    organisation_id       UUID        NOT NULL,
    name                  TEXT        NOT NULL,
    start_date            DATE        NOT NULL,
    end_date              DATE        NOT NULL,
    is_closed             BOOLEAN     NOT NULL DEFAULT FALSE,
    is_locked             BOOLEAN     NOT NULL DEFAULT FALSE,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, organisation_id, name),
    CONSTRAINT chk_fy_dates CHECK (end_date > start_date)
);

CREATE INDEX idx_fiscal_years_tenant ON fiscal_years (tenant_id);

ALTER TABLE fiscal_years ENABLE ROW LEVEL SECURITY;
ALTER TABLE fiscal_years FORCE  ROW LEVEL SECURITY;

CREATE POLICY fiscal_years_app_all ON fiscal_years FOR ALL TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY fiscal_years_ro_select ON fiscal_years FOR SELECT TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);


-- ── accounting_periods ────────────────────────────────────────────────────────
CREATE TABLE accounting_periods (
    tenant_id             UUID        NOT NULL REFERENCES tenants(id),
    id                    UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    organisation_id       UUID        NOT NULL,
    fiscal_year_id        UUID        NOT NULL REFERENCES fiscal_years(id),
    name                  TEXT        NOT NULL,
    start_date            DATE        NOT NULL,
    end_date              DATE        NOT NULL,
    status                TEXT        NOT NULL DEFAULT 'open'
                              CHECK (status IN ('open','soft_closed','hard_closed','locked')),
    closed_at             TIMESTAMPTZ,
    closed_by             UUID,
    locked_at             TIMESTAMPTZ,
    locked_by             UUID,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, organisation_id, start_date)
);

CREATE INDEX idx_periods_tenant ON accounting_periods (tenant_id);
CREATE INDEX idx_periods_date   ON accounting_periods (tenant_id, organisation_id, start_date, end_date);
CREATE INDEX idx_periods_open   ON accounting_periods (tenant_id, status)
    WHERE status = 'open';

-- Trigger: enforce period gate — prevent posting to closed periods
CREATE OR REPLACE FUNCTION fn_check_period_gate()
RETURNS TRIGGER LANGUAGE plpgsql SECURITY INVOKER AS $$
DECLARE v_status TEXT;
BEGIN
    SELECT status INTO v_status
    FROM accounting_periods
    WHERE tenant_id = NEW.tenant_id
      AND organisation_id = NEW.organisation_id
      AND start_date <= NEW.transaction_date
      AND end_date   >= NEW.transaction_date;

    IF v_status IN ('hard_closed', 'locked') THEN
        RAISE EXCEPTION 'Period is % — posting not permitted', v_status
            USING ERRCODE = 'invalid_parameter_value';
    END IF;
    RETURN NEW;
END;$$;

ALTER TABLE accounting_periods ENABLE ROW LEVEL SECURITY;
ALTER TABLE accounting_periods FORCE  ROW LEVEL SECURITY;

CREATE POLICY periods_app_all ON accounting_periods FOR ALL TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY periods_ro_select ON accounting_periods FOR SELECT TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);


-- ── journal_entries ───────────────────────────────────────────────────────────
CREATE TABLE journal_entries (
    tenant_id             UUID        NOT NULL REFERENCES tenants(id),
    id                    UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by            UUID,
    updated_by            UUID,

    organisation_id       UUID        NOT NULL,
    period_id             UUID        NOT NULL REFERENCES accounting_periods(id),
    reference             TEXT        NOT NULL,          -- auto-generated: JE-2025-0145
    transaction_date      DATE        NOT NULL,
    posting_date          DATE,
    description           TEXT        NOT NULL,
    currency              CHAR(3)     NOT NULL,
    type                  TEXT        NOT NULL DEFAULT 'manual'
                              CHECK (type IN ('manual','system','recurring','imported','integration','reversal')),
    status                TEXT        NOT NULL DEFAULT 'draft'
                              CHECK (status IN ('draft','pending_approval','approved','posted','reversed','cancelled')),
    source_module         TEXT,                          -- 'hr' | 'sales' | 'inventory' | etc.
    source_event_id       UUID,                          -- integration event that triggered this
    reversal_of_id        UUID        REFERENCES journal_entries(id),
    submitted_by          UUID,
    submitted_at          TIMESTAMPTZ,
    approved_by           UUID,
    approved_at           TIMESTAMPTZ,
    posted_by             UUID,
    posted_at             TIMESTAMPTZ,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, organisation_id, reference)
);

CREATE INDEX idx_journal_entries_tenant   ON journal_entries (tenant_id);
CREATE INDEX idx_journal_entries_period   ON journal_entries (tenant_id, period_id);
CREATE INDEX idx_journal_entries_status   ON journal_entries (tenant_id, status);
CREATE INDEX idx_journal_entries_source   ON journal_entries (tenant_id, source_event_id)
    WHERE source_event_id IS NOT NULL;

-- Period gate trigger
CREATE TRIGGER trg_journal_period_gate
    BEFORE INSERT ON journal_entries
    FOR EACH ROW EXECUTE FUNCTION fn_check_period_gate();

-- Immutability: posted entries cannot be updated (except status → reversed)
CREATE OR REPLACE FUNCTION fn_protect_posted_journal()
RETURNS TRIGGER LANGUAGE plpgsql SECURITY INVOKER AS $$
BEGIN
    IF OLD.status = 'posted' AND NEW.status NOT IN ('reversed') THEN
        RAISE EXCEPTION 'Posted journal entry % is immutable', OLD.id
            USING ERRCODE = 'invalid_parameter_value';
    END IF;
    RETURN NEW;
END;$$;

CREATE TRIGGER trg_journal_immutability
    BEFORE UPDATE ON journal_entries
    FOR EACH ROW EXECUTE FUNCTION fn_protect_posted_journal();

ALTER TABLE journal_entries ENABLE ROW LEVEL SECURITY;
ALTER TABLE journal_entries FORCE  ROW LEVEL SECURITY;

CREATE POLICY journal_entries_app_all ON journal_entries FOR ALL TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY journal_entries_ro_select ON journal_entries FOR SELECT TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);


-- ── journal_lines ─────────────────────────────────────────────────────────────
CREATE TABLE journal_lines (
    tenant_id             UUID        NOT NULL REFERENCES tenants(id),
    id                    UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    entry_id              UUID        NOT NULL REFERENCES journal_entries(id),
    line_number           INT         NOT NULL,
    account_id            UUID        NOT NULL REFERENCES accounts(id),
    debit_amount          NUMERIC(18,4) NOT NULL DEFAULT 0 CHECK (debit_amount >= 0),
    credit_amount         NUMERIC(18,4) NOT NULL DEFAULT 0 CHECK (credit_amount >= 0),
    description           TEXT,
    cost_centre_id        UUID,
    reference             TEXT,
    currency              CHAR(3),
    fx_rate               NUMERIC(18,8),
    base_amount           NUMERIC(18,4),          -- amount in functional currency

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, entry_id, line_number),

    CONSTRAINT chk_debit_or_credit CHECK (
        (debit_amount > 0 AND credit_amount = 0) OR
        (credit_amount > 0 AND debit_amount = 0)
    )
);

CREATE INDEX idx_journal_lines_tenant  ON journal_lines (tenant_id);
CREATE INDEX idx_journal_lines_entry   ON journal_lines (tenant_id, entry_id);
CREATE INDEX idx_journal_lines_account ON journal_lines (tenant_id, account_id);

-- journal_lines are append-only once the entry is posted
CREATE OR REPLACE FUNCTION fn_protect_posted_line()
RETURNS TRIGGER LANGUAGE plpgsql SECURITY INVOKER AS $$
DECLARE v_status TEXT;
BEGIN
    SELECT status INTO v_status FROM journal_entries WHERE id = OLD.entry_id;
    IF v_status = 'posted' THEN
        RAISE EXCEPTION 'Lines of posted journal entry % are immutable', OLD.entry_id;
    END IF;
    RETURN OLD;
END;$$;

CREATE TRIGGER trg_journal_lines_immutability
    BEFORE UPDATE OR DELETE ON journal_lines
    FOR EACH ROW EXECUTE FUNCTION fn_protect_posted_line();

ALTER TABLE journal_lines ENABLE ROW LEVEL SECURITY;
ALTER TABLE journal_lines FORCE  ROW LEVEL SECURITY;

CREATE POLICY journal_lines_app_all ON journal_lines FOR ALL TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY journal_lines_ro_select ON journal_lines FOR SELECT TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);


-- ── bank_reconciliations ──────────────────────────────────────────────────────
CREATE TABLE bank_reconciliations (
    tenant_id             UUID        NOT NULL REFERENCES tenants(id),
    id                    UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    organisation_id       UUID        NOT NULL,
    bank_account_id       UUID        NOT NULL,
    period_id             UUID        NOT NULL REFERENCES accounting_periods(id),
    statement_opening_bal NUMERIC(18,4) NOT NULL,
    statement_closing_bal NUMERIC(18,4) NOT NULL,
    gl_closing_balance    NUMERIC(18,4) NOT NULL DEFAULT 0,
    adjusted_bank_balance NUMERIC(18,4) NOT NULL DEFAULT 0,
    adjusted_gl_balance   NUMERIC(18,4) NOT NULL DEFAULT 0,
    difference            NUMERIC(18,4) GENERATED ALWAYS AS
                              (adjusted_bank_balance - adjusted_gl_balance) STORED,
    status                TEXT        NOT NULL DEFAULT 'open'
                              CHECK (status IN ('open','in_progress','balanced','under_review','approved','locked')),
    prepared_by           UUID        NOT NULL,
    approved_by           UUID,
    locked_at             TIMESTAMPTZ,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, bank_account_id, period_id)
);

CREATE INDEX idx_bank_recs_tenant ON bank_reconciliations (tenant_id);
CREATE INDEX idx_bank_recs_period ON bank_reconciliations (tenant_id, period_id);

ALTER TABLE bank_reconciliations ENABLE ROW LEVEL SECURITY;
ALTER TABLE bank_reconciliations FORCE  ROW LEVEL SECURITY;

CREATE POLICY bank_recs_app_all ON bank_reconciliations FOR ALL TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY bank_recs_ro_select ON bank_reconciliations FOR SELECT TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);


-- ── budgets ───────────────────────────────────────────────────────────────────
CREATE TABLE budgets (
    tenant_id             UUID        NOT NULL REFERENCES tenants(id),
    id                    UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by            UUID,

    organisation_id       UUID        NOT NULL,
    fiscal_year_id        UUID        NOT NULL REFERENCES fiscal_years(id),
    name                  TEXT        NOT NULL,
    version_label         TEXT        NOT NULL DEFAULT 'V1',
    status                TEXT        NOT NULL DEFAULT 'draft'
                              CHECK (status IN ('draft','approved','active','archived')),
    currency              CHAR(3)     NOT NULL,
    approved_by           UUID,
    approved_at           TIMESTAMPTZ,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id)
);

-- Enforce single active budget per org per fiscal year
CREATE UNIQUE INDEX idx_budgets_one_active
    ON budgets (tenant_id, organisation_id, fiscal_year_id)
    WHERE status = 'active';

CREATE INDEX idx_budgets_tenant ON budgets (tenant_id);

ALTER TABLE budgets ENABLE ROW LEVEL SECURITY;
ALTER TABLE budgets FORCE  ROW LEVEL SECURITY;

CREATE POLICY budgets_app_all ON budgets FOR ALL TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY budgets_ro_select ON budgets FOR SELECT TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);


-- ── cost_centres ──────────────────────────────────────────────────────────────
CREATE TABLE cost_centres (
    tenant_id             UUID        NOT NULL REFERENCES tenants(id),
    id                    UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    organisation_id       UUID        NOT NULL,
    code                  TEXT        NOT NULL,
    name                  TEXT        NOT NULL,
    parent_id             UUID        REFERENCES cost_centres(id),
    path                  TEXT        NOT NULL,
    level                 INT         NOT NULL DEFAULT 1,
    is_group              BOOLEAN     NOT NULL DEFAULT FALSE,
    is_distributed        BOOLEAN     NOT NULL DEFAULT FALSE,
    allocation_method     TEXT
                              CHECK (allocation_method IN ('percentage','headcount','square_footage','usage')),
    allocations           JSONB       NOT NULL DEFAULT '[]',
    is_active             BOOLEAN     NOT NULL DEFAULT TRUE,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, organisation_id, code)
);

CREATE INDEX idx_cost_centres_tenant ON cost_centres (tenant_id);
CREATE INDEX idx_cost_centres_path   ON cost_centres (tenant_id, path text_pattern_ops);

ALTER TABLE cost_centres ENABLE ROW LEVEL SECURITY;
ALTER TABLE cost_centres FORCE  ROW LEVEL SECURITY;

CREATE POLICY cost_centres_app_all ON cost_centres FOR ALL TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY cost_centres_ro_select ON cost_centres FOR SELECT TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);


-- ── account_mappings (semantic code → GL account) ─────────────────────────────
CREATE TABLE account_mappings (
    tenant_id             UUID        NOT NULL REFERENCES tenants(id),
    id                    UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    organisation_id       UUID        NOT NULL,
    semantic_code         TEXT        NOT NULL,  -- e.g. "BASIC_SALARY", "VAT_OUTPUT"
    account_id            UUID        NOT NULL REFERENCES accounts(id),
    cost_centre_id        UUID,
    effective_from        DATE        NOT NULL,
    effective_to          DATE,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, organisation_id, semantic_code, effective_from)
);

CREATE INDEX idx_account_mappings_tenant ON account_mappings (tenant_id);
CREATE INDEX idx_account_mappings_code   ON account_mappings (tenant_id, organisation_id, semantic_code);

ALTER TABLE account_mappings ENABLE ROW LEVEL SECURITY;
ALTER TABLE account_mappings FORCE  ROW LEVEL SECURITY;

CREATE POLICY account_mappings_app_all ON account_mappings FOR ALL TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY account_mappings_ro_select ON account_mappings FOR SELECT TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);


-- ── Standard views ────────────────────────────────────────────────────────────

CREATE OR REPLACE VIEW v_trial_balance AS
SELECT
    ab.tenant_id,
    ab.period_id,
    a.code                AS account_code,
    a.name                AS account_name,
    a.root_type,
    a.account_type,
    a.normal_balance,
    a.path,
    a.level,
    a.is_group,
    ab.opening_debit,
    ab.opening_credit,
    ab.period_debit,
    ab.period_credit,
    ab.closing_debit,
    ab.closing_credit,
    ab.net_balance
FROM account_balances ab
JOIN accounts a ON a.id = ab.account_id;

ALTER VIEW v_trial_balance SET (security_invoker = true);
GRANT SELECT ON v_trial_balance TO application_role;
GRANT SELECT ON v_trial_balance TO readonly_role;


CREATE OR REPLACE VIEW v_period_close_status AS
SELECT
    p.tenant_id,
    p.id                  AS period_id,
    p.name                AS period_name,
    p.status,
    -- Pending approvals blocking close
    (SELECT COUNT(*) FROM journal_entries je
     WHERE je.tenant_id = p.tenant_id
       AND je.period_id = p.id
       AND je.status = 'pending_approval')   AS pending_approvals,
    -- Unapproved bank reconciliations
    (SELECT COUNT(*) FROM bank_reconciliations br
     WHERE br.tenant_id = p.tenant_id
       AND br.period_id = p.id
       AND br.status NOT IN ('approved','locked')) AS unapproved_recs
FROM accounting_periods p;

ALTER VIEW v_period_close_status SET (security_invoker = true);
GRANT SELECT ON v_period_close_status TO application_role;
GRANT SELECT ON v_period_close_status TO readonly_role;
```

---

## 19. Workflow Orchestration (Temporal)

### Period Close Workflow

```go
// infrastructure/temporal/period_close_workflow.go

func PeriodCloseWorkflow(ctx workflow.Context, input PeriodCloseInput) error {
    ao := workflow.ActivityOptions{StartToCloseTimeout: 5 * time.Minute}
    ctx = workflow.WithActivityOptions(ctx, ao)

    // 1. Run all mandatory checklist items
    var checks []PeriodCloseCheck
    if err := workflow.ExecuteActivity(ctx, RunPeriodCloseChecksActivity, input).Get(ctx, &checks); err != nil {
        return err
    }

    // 2. If any blocking check fails, abort
    for _, c := range checks {
        if c.Blocking && !c.Passed {
            return fmt.Errorf("period close blocked: %s — %s", c.Name, c.Detail)
        }
    }

    // 3. Trigger FX revaluation (via Currency module event)
    if err := workflow.ExecuteActivity(ctx, TriggerFXRevaluationActivity, input.PeriodID).Get(ctx, nil); err != nil {
        return err
    }

    // 4. Run IC matching check
    if err := workflow.ExecuteActivity(ctx, RunICMatchingActivity, input.OrganisationID, input.PeriodEndDate).Get(ctx, nil); err != nil {
        return err
    }

    // 5. Hard close the period
    return workflow.ExecuteActivity(ctx, HardClosePeriodActivity, input.PeriodID, input.ClosedBy).Get(ctx, nil)
}
```

### Bank Statement Auto-Import Workflow

```go
// Triggered by a webhook from the bank integration or on a nightly schedule.
func BankStatementImportWorkflow(ctx workflow.Context, input BankImportInput) error {
    ao := workflow.ActivityOptions{StartToCloseTimeout: 10 * time.Minute}
    ctx = workflow.WithActivityOptions(ctx, ao)

    // 1. Fetch and parse the statement
    var lines []StatementLine
    if err := workflow.ExecuteActivity(ctx, FetchAndParseStatementActivity, input).Get(ctx, &lines); err != nil {
        return err
    }

    // 2. Open or continue reconciliation for this period
    var recID uuid.UUID
    if err := workflow.ExecuteActivity(ctx, OpenOrContinueReconciliationActivity, input).Get(ctx, &recID); err != nil {
        return err
    }

    // 3. Run auto-matcher
    return workflow.ExecuteActivity(ctx, RunAutoMatcherActivity, recID, lines).Get(ctx, nil)
}
```

### Budget Allocation Cron

```go
// Runs on the first day of each month to open the new period's budget lines
// and carry forward the prior period's actuals for comparison.
func BudgetPeriodOpenWorkflow(ctx workflow.Context) error {
    ao := workflow.ActivityOptions{StartToCloseTimeout: 10 * time.Minute}
    ctx = workflow.WithActivityOptions(ctx, ao)
    return workflow.ExecuteActivity(ctx, OpenBudgetForNewPeriodActivity, time.Now()).Get(ctx, nil)
}
```

### Intercompany Matching Cron

```go
// Runs at 23:00 on the last day of each month.
func ICMatchingCronWorkflow(ctx workflow.Context) error {
    ao := workflow.ActivityOptions{StartToCloseTimeout: 20 * time.Minute}
    ctx = workflow.WithActivityOptions(ctx, ao)
    periodEnd := lastDayOfMonth(time.Now())
    return workflow.ExecuteActivity(ctx, RunICMatchingAllPairsActivity, periodEnd).Get(ctx, nil)
}
```

---

## 20. Configuration Flags

Stored per tenant in `tenant_config.finance`.

| Flag | Default | Description |
|---|---|---|
| `default_currency` | `KES` | Tenant functional currency (immutable after first transaction) |
| `fiscal_year_start_month` | `1` | Month the fiscal year begins (1 = January) |
| `budget_control_mode` | `soft` | `none` \| `soft` \| `hard` |
| `budget_override_requires_reason` | `true` | Soft budget override must include a reason comment |
| `duplicate_invoice_check` | `true` | Warn on same supplier + invoice number + amount within 30 days |
| `require_cost_centre_on_expense` | `false` | All expense account lines must have a cost centre |
| `auto_reverse_accruals` | `true` | Month-end accrual entries automatically reversed on first day of next period |
| `bank_statement_import_formats` | `["ofx","mt940","csv"]` | Enabled import formats |
| `auto_match_confidence_threshold` | `auto` | Minimum confidence to auto-confirm a match without user review |
| `outstanding_cheque_alert_days` | `60` | Flag outstanding cheques older than N days |
| `ic_matching_required_for_close` | `true` | Period cannot hard-close if IC balances are unmatched |
| `bank_rec_required_for_close` | `true` | Period cannot hard-close without approved bank recs |
| `subsidiary_reconcile_required_for_close` | `true` | AR/AP subsidiary must reconcile to GL control |
| `payment_file_format` | `generic_csv` | Default bank payment export format |
| `journal_reference_prefix` | `JE` | Prefix for auto-generated journal references |
| `journal_reference_sequence_length` | `4` | Zero-padded digits in reference: JE-2025-0001 |
| `multi_entity_enabled` | `false` | Enable multiple organisations per tenant |
| `intercompany_enabled` | `false` | Enable IC transactions and eliminations |
| `petty_cash_enabled` | `true` | Enable petty cash imprest sub-module |

---

*This document is the authoritative technical specification for the Awo ERP Finance Module. The Finance module is the final destination for all monetary events in the system. Every other module communicates with it through events — never through direct table writes.*
