# Awo ERP — Finance Module
## FILE-03: Journal Entry Pipeline & General Ledger

**Document Version:** 2.0.0  
**Series:** FILE-03 of 10  
**Depends On:** FILE-01 (Domain Model), FILE-02 (Chart of Accounts, Account Balances)  
**Depended On By:** FILE-04 (Period Management), FILE-05 (AR/AP), FILE-06 (Budgets), FILE-08 (Reports)

---

## Table of Contents

1. [Why the Journal Pipeline Exists](#1-why-the-journal-pipeline-exists)
2. [What You Lose Without It](#2-what-you-lose-without-it)
3. [How ERPNext, NetSuite and QuickBooks Handle Journal Entries](#3-how-erpnext-netsuite-and-quickbooks-handle-journal-entries)
4. [Entry Lifecycle & Status Machine](#4-entry-lifecycle--status-machine)
5. [Journal Entry Types](#5-journal-entry-types)
6. [The Pipeline: Stages & Execution](#6-the-pipeline-stages--execution)
7. [General Ledger Engine](#7-general-ledger-engine)
8. [Subsidiary Ledger](#8-subsidiary-ledger)
9. [Reversal Engine](#9-reversal-engine)
10. [Recurring Entries](#10-recurring-entries)
11. [Integration Event Consumer Pattern](#11-integration-event-consumer-pattern)
12. [Business Rules & Validation](#12-business-rules--validation)
13. [Performance & Storage](#13-performance--storage)
14. [Database Schema](#14-database-schema)
15. [API Reference](#15-api-reference)
16. [Feature Flags & Configuration](#16-feature-flags--configuration)
17. [v1.0 Rollout Assessment](#17-v10-rollout-assessment)

---

## 1. Why the Journal Pipeline Exists

The journal entry pipeline is the core enforcement mechanism of the Finance module. Every debit and credit — regardless of whether it originates from a fuel sale, a payroll run, a manual adjustment, or a bank fee — passes through the same staged validation and posting process before reaching the General Ledger.

Without a pipeline, every code path that creates a journal entry would need to implement its own validation. Some paths would miss checks, some would implement them differently, and over time the ledger would accumulate entries that violate invariants. The pipeline centralises all rules in one place: validation of structure, balance, accounts, period, references, cost centres, duplicates, budget, approval, and finally the atomic GL write — in that exact order, with no exceptions.

The General Ledger (GL) is the output of the pipeline. It is a read-only view over `journal_lines` and `account_balances`. It never mutates data directly — all mutations flow through the pipeline.

---

## 2. What You Lose Without It

| Scenario | Without Pipeline | With Pipeline |
|---|---|---|
| Unbalanced entry posted | Possible if any code path has a bug | Impossible — DB constraint + pipeline stage both block it |
| Posting to closed period | Depends on which code path is used | Impossible — DB trigger + `ValidatePeriod` stage both block it |
| Integration event posted twice | Duplicate GL entries | Idempotency key checked at `HandleEvent` — second call is a no-op |
| Manual entry to a system account | Depends on application code | Blocked by `ValidateAccounts` stage regardless of code path |
| Concurrent period close + journal post | Race condition possible | Period row locked with `SELECT FOR UPDATE` inside `PostToGL` |
| Missing approval on large entry | Possible | `CheckApproval` stage blocks post if approval not obtained |

---

## 3. How ERPNext, NetSuite and QuickBooks Handle Journal Entries

### ERPNext

ERPNext routes all financial documents through a `submit` action that calls `make_gl_entries`. The validation is implemented as Python methods on each DocType, not as a shared pipeline. This means a `Journal Entry` DocType, a `Sales Invoice` DocType, and a `Payment Entry` DocType each have their own validation logic — some consistent, some not. Over the years this has led to known edge cases where, for example, a Sales Invoice can be submitted against a soft-closed period that a Journal Entry would have been blocked on.

ERPNext does not have a concept of "pending approval" built into the GL posting flow. Workflow is a separate system that can optionally prevent submission, but the default path bypasses it.

**What Awo improves:** A single pipeline with the same stages for all entry types. Approval is stage 9 in the pipeline — it cannot be bypassed regardless of the entry's origin.

### NetSuite

NetSuite uses a concept called "posting periods" with explicit open/close controls per subsidiary. Journal entries go through a standard create → approve → post lifecycle. NetSuite's equivalent of Awo's pipeline is handled internally and is not directly configurable by developers.

One notable NetSuite feature: "reversing journals" can be automatically scheduled for the first day of the next period. Awo implements this as `auto_reverse_accruals` (see §16).

### QuickBooks

QuickBooks Online does not have a formal journal entry approval workflow. Any user with the "Accountant" or "Company Administrator" role can post a journal entry to any date without restriction. There is no period gating, no budget check, and no duplicate detection. For a small single-user business this is acceptable; for an organisation with multiple finance staff it is a significant control gap.

**What Awo provides that QuickBooks does not:** Every one of the 10 pipeline stages.

---

## 4. Entry Lifecycle & Status Machine

```
Draft
  │
  ├──[Cancel]────────────────────────────────► Cancelled (terminal)
  │
  ├──[Post directly if no approval required]──► Approved (implicit)
  │                                                │
  └──[Submit for approval]──► PendingApproval      │
                                    │              │
                               [Approve]           │
                                    │              │
                                    ▼              ▼
                                Approved ──[Post]──► Posted
                                    │                  │
                               [Cancel]           [Reverse]
                                    │                  │
                                    ▼                  ▼
                               Cancelled           Reversed (terminal)
                                                 + new Reversal entry (Draft)
```

**Status definitions:**

| Status | Editable | In GL | What Can Happen Next |
|---|---|---|---|
| `draft` | Yes | No | Submit, Post (if no approval required), Cancel |
| `pending_approval` | No | No | Approve, Reject (returns to Draft), Cancel |
| `approved` | No | No | Post, Cancel |
| `posted` | No | Yes | Reverse only |
| `reversed` | No | Yes | Nothing — terminal |
| `cancelled` | No | No | Nothing — terminal |

---

## 5. Journal Entry Types

The `type` field on a journal entry determines which validation rules apply and how the entry is generated.

| Type | Origin | Manual Entry Restriction | Approval Required |
|---|---|---|---|
| `manual` | Finance team creates directly | Accounts must have `allow_manual_entries = true` | Configurable |
| `system` | Generated by the system (e.g. opening balance propagation) | No restriction | No |
| `recurring` | Generated from a recurring template | No restriction | Configurable |
| `imported` | Uploaded via bulk CSV import | No restriction | Configurable |
| `integration` | Generated by event consumer from another module | No restriction | No (by default) |
| `reversal` | Created by `CreateReversal()` from a posted entry | No restriction | No |

---

## 6. The Pipeline: Stages & Execution

### Pipeline Architecture

The pipeline is a sequence of `Stage` implementations. Each stage receives the entry and either enriches it (e.g. resolves the period ID), validates it (e.g. checks balance), or performs a side effect (e.g. writes to the GL). A stage that returns an error halts the pipeline — no subsequent stages run.

```go
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

All stages before `PostToGL` are pure — they perform no database writes (validation stages may perform reads). Only `PostToGL` writes to the database, and it does so in a single atomic transaction.

### Stage 1 — ValidateStructure

Checks the entry's shape before any database lookups.

- Entry has at least 2 lines
- Each line has a debit amount XOR a credit amount (not both, not neither)
- All amounts are positive (> 0)
- `transaction_date` is not in the future
- `description` is not empty

### Stage 2 — ValidateBalance

```go
func (s *ValidateBalanceStage) Execute(_ context.Context, entry *JournalEntry, _ PostingConfig) error {
    var totalDebit, totalCredit decimal.Decimal
    for _, line := range entry.Lines {
        totalDebit  = totalDebit.Add(line.DebitAmount)
        totalCredit = totalCredit.Add(line.CreditAmount)
    }
    diff := totalDebit.Sub(totalCredit).Abs().Round(2)
    if diff.IsPositive() {
        return ValidationError{
            Code:    "ENTRY_NOT_BALANCED",
            Message: fmt.Sprintf("debits %s ≠ credits %s (diff: %s)", totalDebit, totalCredit, diff),
        }
    }
    return nil
}
```

Amounts are rounded to 2 decimal places before comparison to tolerate floating-point representation issues in upstream systems.

### Stage 3 — ValidateAccounts

For each journal line:
- Account exists and belongs to the same organisation
- Account is active (`is_active = true`)
- Account is a leaf (`is_group = false`)
- If entry type is `manual`, account has `allow_manual_entries = true`
- If account has `locked_currency`, the entry's currency matches

### Stage 4 — ValidatePeriod

Resolves the accounting period for `transaction_date` and checks its status. Calls `PeriodGate.CheckPosting`:

```go
func (g *PeriodGate) CheckPosting(ctx context.Context, orgID uuid.UUID, txDate time.Time, role string) error {
    period, err := g.periodRepo.ForDate(ctx, orgID, txDate)
    if err != nil {
        return ErrNoPeriodForDate{Date: txDate}
    }
    switch period.Status {
    case PeriodOpen:
        return nil
    case PeriodSoftClosed:
        if role == "finance_manager" || role == "cfo" {
            return nil
        }
        return ErrPeriodSoftClosed{Period: period.Name}
    case PeriodHardClosed, PeriodLocked:
        return ErrPeriodClosed{Period: period.Name, Status: string(period.Status)}
    }
    return ErrUnknownPeriodStatus
}
```

If the gate passes, `entry.PeriodID` is set to the resolved period's ID.

### Stage 5 — ValidateReferences

For each journal line whose account has `require_reference = true`, the line must have a non-empty `reference` field.

### Stage 6 — ValidateCostCentres

If the tenant configuration has `require_cost_centre_on_expense = true`, every line posting to an expense account must carry a `cost_centre_id`. This stage is also configurable to only warn (produce a `ValidationWarning`) rather than block.

### Stage 7 — CheckDuplicates

Queries for any posted entry within the past 24 hours (configurable) with the same:
- Organisation ID
- Reference number
- Total amount
- Set of account codes

If a match is found, behaviour depends on configuration:
- `duplicate_action: warn` — appends a `ValidationWarning` to the entry; posting continues
- `duplicate_action: block` — returns a `ValidationError`; posting stops

### Stage 8 — CheckBudget

Only runs if `budget_module_enabled = true` and a `BudgetControlMode` other than `none` is configured. Checks each expense line against the active budget for the entry's period:

```go
func (s *CheckBudgetStage) Execute(ctx context.Context, entry *JournalEntry, cfg PostingConfig) error {
    if cfg.BudgetControlMode == BudgetNone {
        return nil
    }
    results, err := s.checker.Check(ctx, entry.OrganisationID, entry.PeriodID, entry.Lines)
    if err != nil {
        return err
    }
    for _, r := range results {
        if r.IsOverBudget {
            switch cfg.BudgetControlMode {
            case BudgetSoft:
                entry.AddWarning(fmt.Sprintf("account %s will exceed budget by %s", r.AccountCode, r.WouldExceedBy))
            case BudgetHard:
                return ErrBudgetExceeded{AccountCode: r.AccountCode, ExceedBy: r.WouldExceedBy}
            }
        }
    }
    return nil
}
```

### Stage 9 — CheckApproval

If the entry requires approval (based on type, amount, and organisation rules), verifies that `entry.Status == StatusApproved` and that `entry.ApprovedBy != entry.SubmittedBy` (no self-approval). If approval is not yet obtained, returns `ErrApprovalRequired`.

### Stage 10 — PostToGL

The only stage that writes to the database. Runs in a single atomic transaction:

```go
func (s *PostToGLStage) Execute(ctx context.Context, entry *JournalEntry, _ PostingConfig) error {
    return s.db.WithTransaction(ctx, func(tx pgx.Tx) error {
        // 1. Lock the accounting period row to prevent concurrent close racing with this post
        if err := s.lockPeriodRow(ctx, tx, entry.PeriodID); err != nil {
            return err
        }
        // 2. Insert all journal lines
        for _, line := range entry.Lines {
            if err := s.lineRepo.Insert(ctx, tx, line); err != nil {
                return err
            }
        }
        // 3. Upsert account balances (atomic increment)
        if err := s.balanceRepo.ApplyLines(ctx, tx, entry.Lines, entry.PeriodID); err != nil {
            return err
        }
        // 4. Update subsidiary ledgers (AR, AP, inventory, employee receivable)
        if err := s.subLedger.Apply(ctx, tx, entry); err != nil {
            return err
        }
        // 5. Mark entry as posted
        entry.Status   = StatusPosted
        entry.PostedAt = ptr(time.Now())
        if err := s.entryRepo.UpdateStatus(ctx, tx, entry); err != nil {
            return err
        }
        // 6. Write outbox event for other modules
        return s.outbox.Insert(ctx, tx, JournalPostedEvent{
            TenantID:  entry.TenantID,
            EntryID:   entry.ID,
            Reference: entry.Reference,
            PostedAt:  *entry.PostedAt,
        })
    })
}
```

**Why lock the period row?** Without this, a concurrent `HardClosePeriod` call and a `PostJournalEntry` call could both proceed simultaneously — the period close checks "no pending journals" and finds none; the journal insert fires and succeeds after the period closes. The `SELECT FOR UPDATE` on the `accounting_periods` row serialises these operations: whichever arrives second sees the row locked and must wait, then re-reads the (now closed) status and fails cleanly.

---

## 7. General Ledger Engine

### GL as a Read View

The GL is not a separate table. It is a structured read layer over `journal_lines` and `account_balances`. The `GeneralLedger` domain service provides query methods; it never mutates data.

```go
type GeneralLedger struct {
    balanceRepo AccountBalanceRepository
    lineRepo    JournalLineRepository
}

// TrialBalance: all account balances as of the given period
func (gl *GeneralLedger) TrialBalance(ctx context.Context, orgID uuid.UUID, periodID uuid.UUID, showZero bool) ([]TrialBalanceLine, error)

// AccountLedger: transaction-level detail for one account over a date range, with running balance
func (gl *GeneralLedger) AccountLedger(ctx context.Context, accountID uuid.UUID, from, to time.Time) ([]LedgerEntry, error)

// BalanceAt: closing balance for an account as of a specific date
func (gl *GeneralLedger) BalanceAt(ctx context.Context, accountID uuid.UUID, asAt time.Time) (decimal.Decimal, error)
```

### Running Balance in Account Ledger

The GL detail report (account ledger) shows a running balance beside each transaction. This is computed using a PostgreSQL window function — no application-side loop is needed:

```sql
SELECT
    je.transaction_date,
    je.reference,
    je.description,
    jl.debit_amount,
    jl.credit_amount,
    SUM(jl.debit_amount - jl.credit_amount)
        OVER (ORDER BY je.transaction_date, je.id
              ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW)
        + ab.opening_debit - ab.opening_credit  AS running_balance
FROM journal_lines jl
JOIN journal_entries je ON je.id = jl.entry_id
JOIN account_balances ab ON ab.account_id = jl.account_id
    AND ab.period_id = (
        SELECT id FROM accounting_periods
        WHERE organisation_id = $org_id
          AND start_date = (
              SELECT MIN(start_date) FROM accounting_periods
              WHERE organisation_id = $org_id AND start_date >= $from
          )
    )
WHERE jl.account_id = $account_id
  AND je.transaction_date BETWEEN $from AND $to
  AND je.status = 'posted'
ORDER BY je.transaction_date, je.id;
```

### Financial Statement Construction

Financial statements are built from the trial balance using the account hierarchy. The `ReportingService` walks the COA tree, aggregates balances from leaf nodes upward, and formats the output according to the report type.

```go
type BalanceSheet struct {
    AsAt        time.Time
    Currency    string
    Assets      StatementSection
    Liabilities StatementSection
    Equity      StatementSection
    // Invariant: Assets.Total == Liabilities.Total + Equity.Total
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
```

---

## 8. Subsidiary Ledger

The subsidiary ledger keeps customer-level (AR) and supplier-level (AP) detail separate from the GL control accounts. The control account holds the total; the subsidiary holds the breakdown.

The `SubsidiaryLedger.Apply` method is called inside the `PostToGL` transaction and routes to the appropriate subsidiary based on the entry's source module:

```go
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
    case "forecourt":
        return s.applyForecourt(ctx, tx, entry)
    }
    return nil // manual and system entries do not touch subsidiaries directly
}
```

**Subsidiary reconciliation** (verifying that the sum of all subsidiary entries equals the GL control account balance) is a mandatory part of the period close checklist. See FILE-04.

---

## 9. Reversal Engine

### Why Reversals, Not Edits

Posted entries are immutable. When a mistake is found in a posted entry, the correction path is always: reverse the original entry, then post the correct entry. This preserves a complete audit trail — you can see exactly what was believed at the time and when the correction was made.

### How a Reversal Is Created

```go
func (e *JournalEntry) CreateReversal(reversalDate time.Time, reason string, by uuid.UUID) (*JournalEntry, error) {
    if e.Status != StatusPosted {
        return nil, ErrOnlyPostedEntriesCanBeReversed
    }
    reversal := &JournalEntry{
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
    }
    for _, line := range e.Lines {
        reversal.Lines = append(reversal.Lines, JournalLine{
            AccountID:    line.AccountID,
            DebitAmount:  line.CreditAmount, // swap
            CreditAmount: line.DebitAmount,  // swap
            Description:  line.Description,
            CostCentreID: line.CostCentreID,
        })
    }
    e.Status = StatusReversed
    return reversal, nil
}
```

The reversal is created as a `Draft` entry. It must then pass through the full pipeline (including period validation — the reversal date must be in an open period, which may differ from the original entry's date). This ensures the reversal is subject to the same controls as any other entry.

### Auto-Reversal (Accruals)

When a journal entry has `AutoReversalDate` set (typically the first day of the next month for accrual entries), the Temporal `AccrualReversalWorkflow` picks it up and calls `CreateReversal` on that date automatically. This requires the next period to be open at that time.

---

## 10. Recurring Entries

A recurring entry template stores a journal entry definition and a schedule. On each scheduled date, the system generates a new `Draft` entry from the template, which then goes through the normal pipeline.

```go
type RecurringEntry struct {
    ID             uuid.UUID
    TenantID       uuid.UUID
    OrganisationID uuid.UUID
    Name           string
    Template       JournalEntryTemplate  // header + lines without amounts (or with fixed amounts)
    Schedule       RecurringSchedule     // frequency, day of month, end date
    LastRunAt      *time.Time
    NextRunAt      time.Time
    IsActive       bool
}

type RecurringSchedule struct {
    Frequency  string  // monthly | quarterly | weekly
    DayOfMonth int     // 1–28 (28 max to handle Feb)
    EndDate    *time.Time
}
```

**Examples of recurring entries:**
- Monthly rent (fixed amount, same accounts each month)
- Quarterly depreciation catch-up
- Monthly loan repayment with a fixed split between principal and interest accounts

Recurring entries do not automatically post — they create a Draft that a finance user reviews and posts. If automatic posting is required (e.g. for depreciation computed by the Fixed Assets module), the entry is created as an `integration` type by the source module's event.

---

## 11. Integration Event Consumer Pattern

All events from other modules are processed by dedicated consumer functions. The consumer pattern enforces idempotency — processing the same event twice produces the same result as processing it once.

```go
type HRConsumer struct {
    consumer    IdempotentConsumer
    mappingRepo AccountMappingRepository
    journalSvc  JournalService
}

func (c *HRConsumer) HandlePayrollPosted(ctx context.Context, event IntegrationEvent) error {
    var payload PayrollPostedPayload
    if err := json.Unmarshal(event.Payload, &payload); err != nil {
        return fmt.Errorf("unmarshal PayrollPosted: %w", err)
    }
    return c.consumer.Handle(ctx, event, func(ctx context.Context, tx pgx.Tx) error {
        // 1. Resolve semantic codes to actual GL account IDs for this tenant/org
        mapping, err := c.mappingRepo.ResolveAll(ctx, tx, event.TenantID, payload.SemanticCodes())
        if err != nil {
            return err
        }
        // 2. Build the journal entry from the event payload
        entry, err := buildPayrollJournalEntry(payload, mapping)
        if err != nil {
            return err
        }
        // 3. Run through the full pipeline (all 10 stages)
        return c.journalSvc.PostIntegrationEntry(ctx, tx, entry)
    })
}
```

**Idempotency mechanism:** The `IdempotentConsumer` checks the `source_event_id` against `journal_entries` before processing. If a row with that `source_event_id` already exists, the handler is skipped and the existing entry ID is returned. This uses a unique index on `(tenant_id, source_event_id)`.

**Dead-letter queue:** If an event cannot be processed (e.g. a semantic code has no mapping configured, or the period is closed), the event is written to `integration_event_failures` with an error description. A finance manager can resolve the issue (configure the mapping, reopen the period) and retry the event from the UI.

---

## 12. Business Rules & Validation

### Journal Entry Rules

| Rule ID | Rule | Stage | Severity |
|---|---|---|---|
| `FIN-JNL-001` | Total debits must equal total credits (to 2dp in functional currency) | ValidateBalance | Error |
| `FIN-JNL-002` | Minimum 2 lines; at least one debit and one credit | ValidateStructure | Error |
| `FIN-JNL-003` | Each line has debit XOR credit — not both, not neither | ValidateStructure | Error |
| `FIN-JNL-004` | All line amounts must be positive | ValidateStructure | Error |
| `FIN-JNL-005` | Transaction date cannot be in the future | ValidateStructure | Error |
| `FIN-JNL-006` | All accounts must exist, be active, be leaf accounts, and belong to the same organisation | ValidateAccounts | Error |
| `FIN-JNL-007` | Transaction date must fall in an open accounting period | ValidatePeriod | Error |
| `FIN-JNL-008` | Accounts with `allow_manual_entries = false` cannot appear in MANUAL entries | ValidateAccounts | Error |
| `FIN-JNL-009` | Accounts with `require_reference = true` must have a line reference | ValidateReferences | Error |
| `FIN-JNL-010` | Accounts with a locked currency must use that currency | ValidateAccounts | Error |
| `FIN-JNL-011` | A posted entry cannot be edited — only reversed | Application + DB trigger | Error |
| `FIN-JNL-012` | Self-approval is prohibited | CheckApproval | Error |
| `FIN-JNL-013` | Integration entries bypass `allow_manual_entries` restriction | ValidateAccounts | Auto |
| `FIN-JNL-014` | Duplicate detection (same ref + amount + accounts within 24h) | CheckDuplicates | Configurable |
| `FIN-JNL-015` | Budget overrun on expense lines | CheckBudget | Configurable |

### Alternatives Considered

**Alternative: Skip the pipeline for integration entries and write directly to journal_lines.**
Rejected. Integration entries still need period validation, account validation, and balance verification. The difference is that some rules (like `allow_manual_entries`) are relaxed for integration entries — this is handled by the `PostingConfig.EntryType` passed to each stage, not by skipping the pipeline.

**Alternative: Run all stages in a single database transaction.**
Rejected. The read-only stages (1–9) do not need to hold a database transaction open — doing so would waste connection pool resources during approval waits and network round-trips. Only stage 10 (`PostToGL`) runs inside a transaction. The cost is that data checked in stages 1–9 could theoretically change between stage 9 and stage 10 (e.g. an account could be deactivated in that window). The `PostToGL` stage re-validates account status inside the transaction as a final check.

---

## 13. Performance & Storage

### Journal Line Volume Estimates

This is the most important performance surface in the Finance module.

| Scenario | Lines/Day | Lines/Year | Lines/5 Years |
|---|---|---|---|
| Single petrol station (10 pumps, 3 shifts) | ~300 | ~110,000 | ~550,000 |
| 5-site operator | ~1,500 | ~550,000 | ~2,750,000 |
| 20-site operator | ~6,000 | ~2,200,000 | ~11,000,000 |

At the 5-site level, `journal_lines` becomes the largest table in the Finance module by row count. The design choices below keep it fast even at this scale.

### Query Performance Targets

| Query | Mechanism | Target P95 |
|---|---|---|
| Post a journal entry (full pipeline + GL write) | Stages 1–9 are in-memory; stage 10 is a single transaction with indexed upserts | < 500ms |
| Account ledger detail (1 account, 1 year) | Composite index on `(tenant_id, account_id)` + window function | < 300ms |
| Trial balance (200 accounts, current period) | Index scan on `account_balances` by period | < 50ms |
| Journal entry list (paginated, filtered by status) | Composite index on `(tenant_id, status, created_at DESC)` | < 100ms |
| Duplicate check (same ref + amount within 24h) | Index on `(tenant_id, organisation_id, reference)` + date filter | < 30ms |

### Index Strategy for journal_lines

```sql
-- Primary GL detail query: all lines for an account in a date range
CREATE INDEX idx_journal_lines_account_date
    ON journal_lines (tenant_id, account_id, created_at DESC)
    INCLUDE (debit_amount, credit_amount, entry_id);

-- Entry → lines lookup (loading a single entry's lines)
CREATE INDEX idx_journal_lines_entry
    ON journal_lines (tenant_id, entry_id);

-- Cost centre reporting (P&L by cost centre)
CREATE INDEX idx_journal_lines_cost_centre
    ON journal_lines (tenant_id, cost_centre_id)
    WHERE cost_centre_id IS NOT NULL;
```

### Partitioning Strategy (Post-v1.0)

At the 20-site scale, `journal_lines` with 11M rows benefits from range partitioning by `created_at` (yearly partitions). This is not needed at v1.0 but the schema should be designed to accommodate it:
- Use `created_at` as the partitioning key, not `transaction_date`, because `created_at` is monotonically increasing and allows partition pruning on recent queries
- Each partition covers one fiscal year
- Old partitions can be moved to cheaper tablespace (or object storage via FDW) without affecting the current-year partition

### Storage Estimates

| Table | Row Size (avg) | Rows/Year (5-site) | Annual Storage |
|---|---|---|---|
| `journal_entries` | ~500 bytes | ~100,000 | ~50 MB |
| `journal_lines` | ~300 bytes | ~550,000 | ~165 MB |
| `account_balances` | ~120 bytes | 300 accts × 12 = 3,600 | ~0.4 MB |

Total Finance module storage for a 5-site operator: approximately **215 MB per year**, growing linearly. 5 years of history ≈ 1 GB. This is well within a single PostgreSQL instance.

For a 20-site operator: approximately **850 MB per year**, 5 years ≈ 4 GB. Still manageable on a single instance with standard hardware.

---

## 14. Database Schema

```sql
CREATE TABLE journal_entries (
    tenant_id        UUID          NOT NULL REFERENCES tenants(id),
    id               UUID          NOT NULL DEFAULT gen_random_uuid(),
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    created_by       UUID,
    updated_by       UUID,

    organisation_id  UUID          NOT NULL,
    period_id        UUID          NOT NULL REFERENCES accounting_periods(id),
    reference        TEXT          NOT NULL,
    transaction_date DATE          NOT NULL,
    posting_date     DATE,
    description      TEXT          NOT NULL,
    currency         CHAR(3)       NOT NULL,
    type             TEXT          NOT NULL DEFAULT 'manual'
                         CHECK (type IN ('manual','system','recurring','imported','integration','reversal')),
    status           TEXT          NOT NULL DEFAULT 'draft'
                         CHECK (status IN ('draft','pending_approval','approved','posted','reversed','cancelled')),
    source_module    TEXT,
    source_event_id  UUID,
    reversal_of_id   UUID          REFERENCES journal_entries(id),
    submitted_by     UUID,
    submitted_at     TIMESTAMPTZ,
    approved_by      UUID,
    approved_at      TIMESTAMPTZ,
    posted_by        UUID,
    posted_at        TIMESTAMPTZ,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, organisation_id, reference),
    UNIQUE (tenant_id, source_event_id) -- idempotency for integration entries
);

CREATE INDEX idx_je_tenant_status   ON journal_entries (tenant_id, status, created_at DESC);
CREATE INDEX idx_je_tenant_period   ON journal_entries (tenant_id, period_id);
CREATE INDEX idx_je_source_event    ON journal_entries (tenant_id, source_event_id)
    WHERE source_event_id IS NOT NULL;

-- Period gate trigger
CREATE OR REPLACE FUNCTION fn_check_period_gate()
RETURNS TRIGGER LANGUAGE plpgsql SECURITY INVOKER AS $$
DECLARE
    v_status TEXT;
BEGIN
    SELECT status INTO v_status
    FROM accounting_periods
    WHERE tenant_id       = NEW.tenant_id
      AND organisation_id = NEW.organisation_id
      AND start_date      <= NEW.transaction_date
      AND end_date        >= NEW.transaction_date
    FOR SHARE; -- prevents concurrent period close racing with this insert
    IF v_status IN ('hard_closed', 'locked') THEN
        RAISE EXCEPTION 'Period is % — posting not permitted', v_status
            USING ERRCODE = 'invalid_parameter_value';
    END IF;
    RETURN NEW;
END;$$;

CREATE TRIGGER trg_journal_period_gate
    BEFORE INSERT ON journal_entries
    FOR EACH ROW EXECUTE FUNCTION fn_check_period_gate();

-- Immutability trigger
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
CREATE POLICY je_app ON journal_entries FOR ALL TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);
CREATE POLICY je_ro ON journal_entries FOR SELECT TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);

-- ── journal_lines ─────────────────────────────────────────────────────────────

CREATE TABLE journal_lines (
    tenant_id      UUID          NOT NULL REFERENCES tenants(id),
    id             UUID          NOT NULL DEFAULT gen_random_uuid(),
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    entry_id       UUID          NOT NULL REFERENCES journal_entries(id),
    line_number    INT           NOT NULL,
    account_id     UUID          NOT NULL REFERENCES accounts(id),
    debit_amount   NUMERIC(18,4) NOT NULL DEFAULT 0 CHECK (debit_amount  >= 0),
    credit_amount  NUMERIC(18,4) NOT NULL DEFAULT 0 CHECK (credit_amount >= 0),
    description    TEXT,
    cost_centre_id UUID,
    reference      TEXT,
    currency       CHAR(3),
    fx_rate        NUMERIC(18,8),
    base_amount    NUMERIC(18,4),

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, entry_id, line_number),
    CONSTRAINT chk_debit_xor_credit CHECK (
        (debit_amount > 0 AND credit_amount = 0) OR
        (credit_amount > 0 AND debit_amount = 0)
    )
);

CREATE INDEX idx_jl_account_date ON journal_lines (tenant_id, account_id, created_at DESC)
    INCLUDE (debit_amount, credit_amount, entry_id);
CREATE INDEX idx_jl_entry        ON journal_lines (tenant_id, entry_id);
CREATE INDEX idx_jl_cost_centre  ON journal_lines (tenant_id, cost_centre_id)
    WHERE cost_centre_id IS NOT NULL;

-- Lines of a posted entry are append-only
CREATE OR REPLACE FUNCTION fn_protect_posted_line()
RETURNS TRIGGER LANGUAGE plpgsql SECURITY INVOKER AS $$
DECLARE v_status TEXT;
BEGIN
    SELECT status INTO v_status FROM journal_entries WHERE id = OLD.entry_id;
    IF v_status = 'posted' THEN
        RAISE EXCEPTION 'Lines of posted entry % are immutable', OLD.entry_id;
    END IF;
    RETURN OLD;
END;$$;

CREATE TRIGGER trg_jl_immutability
    BEFORE UPDATE OR DELETE ON journal_lines
    FOR EACH ROW EXECUTE FUNCTION fn_protect_posted_line();

ALTER TABLE journal_lines ENABLE ROW LEVEL SECURITY;
ALTER TABLE journal_lines FORCE  ROW LEVEL SECURITY;
CREATE POLICY jl_app ON journal_lines FOR ALL TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);
CREATE POLICY jl_ro ON journal_lines FOR SELECT TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);
```

---

## 15. API Reference

### Journal Entries

```
GET    /finance/journal-entries                       List (filter: status, type, from, to, account_code, reference, q)
POST   /finance/journal-entries                       Create draft entry
GET    /finance/journal-entries/{id}                  Get entry with all lines
POST   /finance/journal-entries/{id}/submit           Draft → PendingApproval
POST   /finance/journal-entries/{id}/approve          PendingApproval → Approved
POST   /finance/journal-entries/{id}/reject           PendingApproval → Draft (with reason)
POST   /finance/journal-entries/{id}/post             Approved (or Draft if no approval required) → Posted
POST   /finance/journal-entries/{id}/reverse          Posted → create reversal draft
POST   /finance/journal-entries/{id}/cancel           Draft | Approved → Cancelled
POST   /finance/journal-entries/import                Bulk import from CSV
```

**Create entry request:**
```json
{
  "organisation_id": "uuid",
  "transaction_date": "2025-01-31",
  "reference": "CHQ-001234",
  "description": "Office rent — January 2025",
  "currency": "KES",
  "lines": [
    { "account_code": "7310", "debit": 50000.00, "description": "January rent", "cost_centre_code": "CC-ADMIN" },
    { "account_code": "1112", "credit": 50000.00 }
  ]
}
```

**Response (201 Created):**
```json
{
  "id": "uuid",
  "reference_number": "JE-2025-0145",
  "status": "draft",
  "total_amount": 50000.00,
  "currency": "KES",
  "lines": [ ... ]
}
```

**Reverse entry request:**
```json
{
  "reversal_date": "2025-02-05",
  "reason": "Incorrectly coded to rent — should be maintenance"
}
```

---

## 16. Feature Flags & Configuration

| Flag | Type | Default | Description |
|---|---|---|---|
| `approval_required_above_amount` | decimal | `null` | All manual entries above this amount require approval. `null` means approval is never automatically required (still configurable per entry type in workflow rules). |
| `approval_required_for_integration` | bool | `false` | Integration entries require approval before posting. Usually false — integration entries are generated by trusted code. |
| `duplicate_check_enabled` | bool | `true` | Enable duplicate detection in stage 7. |
| `duplicate_check_window_hours` | int | `24` | How far back to check for duplicates. |
| `duplicate_action` | enum | `warn` | `warn` · `block` — what to do when a duplicate is detected. |
| `auto_reverse_accruals` | bool | `true` | Entries with a set `auto_reversal_date` are automatically reversed by the Temporal workflow. |
| `integration_dead_letter_enabled` | bool | `true` | Failed integration events go to dead-letter queue for manual resolution. If false, they are silently dropped (not recommended). |
| `journal_reference_prefix` | string | `JE` | Prefix for auto-generated references. |
| `journal_reference_sequence_length` | int | `4` | Digit count in sequence: `JE-2025-0001`. |
| `max_lines_per_entry` | int | `200` | Hard limit on lines per journal entry. Prevents accidental payroll entries with thousands of lines from being created manually. Integration entries are exempt. |

---

## 17. v1.0 Rollout Assessment

### Must Have at v1.0

- All 10 pipeline stages implemented and tested
- `PostToGL` atomicity with period row locking
- DB immutability triggers on `journal_entries` and `journal_lines`
- Idempotency on integration event consumers (`source_event_id` unique index)
- Reversal engine
- Manual journal entry creation and posting
- Integration consumers for HR, Sales, Procurement, Forecourt (the four modules in v1.0 scope)
- Dead-letter queue for failed integration events

### Can Be Deferred

| Feature | Suggested Release |
|---|---|
| Recurring entry templates and scheduler | v1.1 |
| Budget check stage (stage 8) | v1.1 (after Budget module is live) |
| Bulk CSV import for journal entries | v1.1 |
| Auto-reversal via Temporal workflow | v1.1 |
| Duplicate detection stage | v1.1 (but ship as `warn` mode so it doesn't block) |

### Never Defer

- DB-level immutability triggers
- Period gate trigger with `FOR SHARE` locking
- `chk_debit_xor_credit` constraint on `journal_lines`
- `UNIQUE (tenant_id, source_event_id)` for idempotency
- RLS policies on both tables

---

*End of FILE-03. Proceeding to FILE-04: Period Management & Bank Reconciliation.*
