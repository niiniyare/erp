# Awo ERP — Finance Module
## FILE-04: Period Management & Bank Reconciliation

**Document Version:** 2.0.0  
**Series:** FILE-04 of 10  
**Depends On:** FILE-01 (Domain Model), FILE-02 (COA), FILE-03 (Journal Pipeline)  
**Depended On By:** FILE-05 (AR/AP), FILE-06 (Budgets), FILE-08 (Reports)

---

## Table of Contents

1. [Why Period Management Exists](#1-why-period-management-exists)
2. [Why Bank Reconciliation Exists](#2-why-bank-reconciliation-exists)
3. [What You Lose Without Them](#3-what-you-lose-without-them)
4. [How ERPNext, NetSuite and QuickBooks Handle These](#4-how-erpnext-netsuite-and-quickbooks-handle-these)
5. [Fiscal Year & Accounting Period Model](#5-fiscal-year--accounting-period-model)
6. [Period Status Machine](#6-period-status-machine)
7. [Period Gate (Application + Database)](#7-period-gate-application--database)
8. [Period Close Workflow & Checklist](#8-period-close-workflow--checklist)
9. [Bank Reconciliation Workspace](#9-bank-reconciliation-workspace)
10. [Bank Statement Import](#10-bank-statement-import)
11. [Auto-Matching Engine](#11-auto-matching-engine)
12. [Reconciliation Lifecycle](#12-reconciliation-lifecycle)
13. [Business Rules & Validation](#13-business-rules--validation)
14. [Performance & Storage](#14-performance--storage)
15. [Database Schema](#15-database-schema)
16. [API Reference](#16-api-reference)
17. [Feature Flags & Configuration](#17-feature-flags--configuration)
18. [v1.0 Rollout Assessment](#18-v10-rollout-assessment)

---

## 1. Why Period Management Exists

Accounting operates in time-bounded segments called accounting periods. Without period management, journal entries can be posted to any date — past or future — at any time, which means:

- A report generated today might show different figures tomorrow because someone backdated a correction to last month
- An audit cannot rely on "closing figures" because those figures can always be changed
- There is no concept of "the books are complete for January" — they are never complete
- Tax returns based on the system's figures might differ from what the system shows a week later

Period management solves this by creating explicit time gates. Once a period is closed, its figures are locked in. Auditors, management, regulators, and tax authorities can rely on period figures being stable.

---

## 2. Why Bank Reconciliation Exists

The bank balance per the General Ledger and the balance per the bank statement will almost never match exactly at any given moment. This is normal — it reflects timing differences (deposits in transit, outstanding cheques) and items the company hasn't recorded yet (bank charges, interest). Bank reconciliation is the process of systematically explaining every difference.

Without bank reconciliation:

- Cash figures on the balance sheet cannot be verified against an external source
- Fraud involving bank accounts (fictitious payments, diverted receipts) can go undetected for months
- Period close cannot be certified with confidence
- The system has no way to discover unrecorded bank charges, interest income, or direct debits

The bank reconciliation workspace in Awo is not just a comparison report. It is an active workspace where statement lines are imported, matched to GL entries, unmatched items are investigated and resolved, and adjusting entries are created for items not yet in the books — all within a structured lifecycle that ends with an approval that can't be given by the person who prepared it.

---

## 3. What You Lose Without Them

| Capability | Without Period Management | Without Bank Reconciliation |
|---|---|---|
| Stable closing figures | Impossible | N/A |
| Audit readiness | Reports may change after audit fieldwork | Cash cannot be verified to external source |
| Regulatory compliance | Month-end figures are a moving target | Balance sheet cash is unverified |
| Fraud detection | N/A | Undetected for months |
| Period close integrity | No concept of "closed" | Close checklist cannot complete |
| Historical comparisons | Prior period figures can change | Cash history unreliable |

---

## 4. How ERPNext, NetSuite and QuickBooks Handle These

### Period Management

**ERPNext:** Period close is available but advisory. The `Period Closing Voucher` DocType generates year-end closing entries. For monthly close, ERPNext allows you to mark a period as closed, but the enforcement is at the application level only — there is no database-level gate preventing backdated entries. A determined user (or a background job) can bypass the close.

**NetSuite:** Period management is one of NetSuite's strongest features. Periods have explicit open/close/locked states per subsidiary. The period gate is enforced at the transaction processing level. NetSuite also has a `Manage Accounting Periods` screen with a checklist approach similar to Awo's.

**QuickBooks:** QuickBooks Online has a "Closing Date" with an optional password. Entries before the closing date prompt a warning and require the password. There is no hard enforcement — it is trivially bypassed. QuickBooks does not have a formal period status machine.

**Awo's position:** Enforces at both application level (PeriodGate service) and database level (trigger with `FOR SHARE` locking). The enforcement is not bypassed by any code path.

### Bank Reconciliation

**ERPNext:** Has a Bank Reconciliation tool but it is largely manual — you select uncleared transactions and mark them as cleared. The v14 "Bank Reconciliation" feature adds a statement upload and matching UI, but auto-matching is limited to exact amount matches.

**NetSuite:** Has a robust bank reconciliation with rule-based matching. Supports OFX and CSV imports. The reconciliation is tied to the period and is part of the close checklist.

**QuickBooks:** QuickBooks has the best bank reconciliation UX of the three for non-accountant users. The "Banking" feed auto-matches downloaded transactions. However it lacks the formal approval workflow and period-tie that Awo provides.

**Awo's approach:** Combines the UX clarity of QuickBooks with the control rigour of NetSuite — auto-matcher with confidence levels, formal prepare/review/approve workflow, mandatory for period close.

---

## 5. Fiscal Year & Accounting Period Model

### FiscalYear

```go
type FiscalYear struct {
    ID             uuid.UUID
    TenantID       uuid.UUID
    OrganisationID uuid.UUID
    Name           string       // e.g. "FY2025"
    StartDate      time.Time
    EndDate        time.Time
    IsClosed       bool         // all periods within are hard_closed or locked
    IsLocked       bool         // no transactions whatsoever; requires CFO + CEO to unlock
    Periods        []AccountingPeriod
    CreatedAt      time.Time
}
```

A fiscal year is created with 12 periods generated automatically — one per calendar month, or one per quarter if quarterly periods are configured. The system does not support non-standard period counts (e.g. 13-period retail calendars) in v1.0.

### AccountingPeriod

```go
type AccountingPeriod struct {
    ID             uuid.UUID
    TenantID       uuid.UUID
    OrganisationID uuid.UUID
    FiscalYearID   uuid.UUID
    Name           string       // e.g. "Jan 2025"
    StartDate      time.Time
    EndDate        time.Time
    Status         PeriodStatus
    ClosedAt       *time.Time
    ClosedBy       *uuid.UUID
    LockedAt       *time.Time
    LockedBy       *uuid.UUID
}
```

### Key Rules on Fiscal Years

- A tenant must always have at least one open fiscal year. Attempting to close or lock the only remaining open fiscal year is blocked.
- Fiscal years must not overlap for the same organisation. The system enforces `UNIQUE (tenant_id, organisation_id, start_date)`.
- A fiscal year cannot be marked `IsClosed` until all 12 of its periods are `hard_closed` or `locked`.
- Once `IsLocked`, the fiscal year cannot be reopened without a dual-authorisation event logged to the audit trail.

---

## 6. Period Status Machine

```
Open
  │
  ├──[SoftClose by finance_manager or cfo]──► SoftClosed
  │                                               │
  │                                          [Reopen]
  │                                               │
  │                                               ▼
  │                                             Open
  │
  └──[HardClose (checklist must pass)]────────► HardClosed
                                                    │
                                               [Reopen with reason + CFO]
                                                    │
                                                    ▼
                                                  Open
                                                    │
                                              [Lock — CFO only]
                                                    │
                                                    ▼
                                                 Locked (terminal — no reopen via UI)
```

### Status Definitions

| Status | Who Can Post | Who Can Close/Lock | Notes |
|---|---|---|---|
| `open` | Any authorised user | `finance_manager` can soft-close | Normal operating state |
| `soft_closed` | `finance_manager`, `cfo` only | `finance_manager` can hard-close | Used during month-end close process |
| `hard_closed` | Nobody | `cfo` can lock | All checklist items must pass to reach this state |
| `locked` | Nobody | Nobody (requires out-of-band process) | Permanent archive state |

### Transition Methods on AccountingPeriod

```go
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

func (p *AccountingPeriod) Lock(by uuid.UUID, now time.Time) error {
    if p.Status != PeriodHardClosed {
        return ErrMustHardCloseBeforeLocking
    }
    p.Status   = PeriodLocked
    p.LockedAt = &now
    p.LockedBy = &by
    return nil
}
```

---

## 7. Period Gate (Application + Database)

### Application Layer: PeriodGate Service

Called by the `ValidatePeriod` pipeline stage. Returns a structured error with the period name and status so the UI can show a meaningful message.

```go
type PeriodGate struct {
    periodRepo PeriodRepository
}

func (g *PeriodGate) CheckPosting(
    ctx    context.Context,
    orgID  uuid.UUID,
    txDate time.Time,
    role   string,
) error {
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

### Database Layer: Trigger

The trigger is the safety net. It fires `BEFORE INSERT` on `journal_entries` and uses `FOR SHARE` to prevent a concurrent period-close transaction from finishing before the insert is complete.

```sql
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
    FOR SHARE; -- serialises with concurrent HardClosePeriod
    IF v_status IN ('hard_closed', 'locked') THEN
        RAISE EXCEPTION 'Period is % — posting not permitted', v_status
            USING ERRCODE = 'invalid_parameter_value';
    END IF;
    RETURN NEW;
END;$$;
```

Note: `soft_closed` is intentionally not blocked at the trigger level because the trigger does not have access to the user's role. The role check happens at the application layer only. The database guarantees the hard boundary (`hard_closed`, `locked`); the application enforces the softer boundary (`soft_closed` for non-finance roles).

---

## 8. Period Close Workflow & Checklist

### Mandatory Close Checks

Before a period can transition from `soft_closed` to `hard_closed`, all of the following checks must pass. These are evaluated by the Temporal `PeriodCloseWorkflow` (or synchronously via the API if Temporal is not yet deployed):

```go
var mandatoryCloseChecks = []PeriodCloseCheckDef{
    {Name: "bank_reconciliations_approved",    Blocking: true},
    {Name: "ar_subsidiary_reconciled",         Blocking: true},
    {Name: "ap_subsidiary_reconciled",         Blocking: true},
    {Name: "fx_revaluation_completed",         Blocking: configurable}, // only if FC balances exist
    {Name: "intercompany_balances_matched",    Blocking: configurable}, // only if IC enabled
    {Name: "no_pending_approvals",             Blocking: true},
    {Name: "trial_balance_balanced",           Blocking: true},
    {Name: "no_failed_integration_events",     Blocking: configurable},
}
```

Each check is a query against the read replica. If a check fails, the close is blocked and the API response includes the check name and a human-readable explanation.

**`bank_reconciliations_approved`:** Queries `bank_reconciliations` for all bank accounts linked to the organisation. All must have `status = 'approved'` or `status = 'locked'` for the current period.

**`ar_subsidiary_reconciled`:** Calls `SubsidiaryLedger.ReconcileAR()` — verifies that the sum of all customer open balances equals the AR control account balance. A non-zero difference (above a configurable tolerance of 0.01 KES) blocks the close.

**`trial_balance_balanced`:** Queries `account_balances` for the period and verifies that the sum of all debit closing balances equals the sum of all credit closing balances. This should always pass if the pipeline is working correctly, but it is an explicit sanity check before close.

### Temporal Period Close Workflow

```go
func PeriodCloseWorkflow(ctx workflow.Context, input PeriodCloseInput) error {
    ao := workflow.ActivityOptions{StartToCloseTimeout: 5 * time.Minute}
    ctx = workflow.WithActivityOptions(ctx, ao)

    // 1. Run all mandatory checks
    var checks []PeriodCloseCheck
    if err := workflow.ExecuteActivity(ctx, RunPeriodCloseChecksActivity, input).Get(ctx, &checks); err != nil {
        return err
    }
    for _, c := range checks {
        if c.Blocking && !c.Passed {
            return fmt.Errorf("close blocked: %s — %s", c.Name, c.Detail)
        }
    }

    // 2. Trigger FX revaluation (emits event to Currency module; waits for JournalPosted confirmation)
    if err := workflow.ExecuteActivity(ctx, TriggerFXRevaluationActivity, input.PeriodID).Get(ctx, nil); err != nil {
        return err
    }

    // 3. Run IC matching check (if enabled)
    if input.ICEnabled {
        if err := workflow.ExecuteActivity(ctx, RunICMatchingActivity, input.OrganisationID, input.PeriodEndDate).Get(ctx, nil); err != nil {
            return err
        }
    }

    // 4. Hard close the period
    return workflow.ExecuteActivity(ctx, HardClosePeriodActivity, input.PeriodID, input.ClosedBy).Get(ctx, nil)
}
```

---

## 9. Bank Reconciliation Workspace

### Workspace Model

One `BankReconciliation` is opened per bank account per accounting period. It cannot be re-opened once locked.

```go
type BankReconciliation struct {
    ID                  uuid.UUID
    TenantID            uuid.UUID
    OrganisationID      uuid.UUID
    BankAccountID       uuid.UUID
    PeriodID            uuid.UUID
    StatementOpeningBal decimal.Decimal  // per bank statement
    StatementClosingBal decimal.Decimal  // per bank statement
    GLClosingBalance    decimal.Decimal  // from account_balances at period end
    AdjustedBankBalance decimal.Decimal  // statement balance + deposits-in-transit - outstanding cheques - errors
    AdjustedGLBalance   decimal.Decimal  // GL balance + adjusting entries posted during rec
    Difference          decimal.Decimal  // AdjustedBank - AdjustedGL (generated column in DB)
    Status              ReconciliationStatus
    PreparedBy          uuid.UUID
    ApprovedBy          *uuid.UUID
    LockedAt            *time.Time
    Matches             []ReconciliationMatch
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
    if r.Status != RecUnderReview {
        return ErrInvalidTransition{}
    }
    if r.PreparedBy == by {
        return ErrSelfApprovalProhibited
    }
    r.ApprovedBy = &by
    r.Status     = RecApproved
    return nil
}
```

### Reconciliation Status Machine

```
Open ──► InProgress ──► Balanced ──► UnderReview ──► Approved ──► Locked
             ▲                            │
             └────────────────────────────┘  (sent back by reviewer)
```

| Status | Who Can Act | What They Can Do |
|---|---|---|
| `open` | AP/AR clerk | Import statement, begin matching |
| `in_progress` | AP/AR clerk | Match, create GL adjustments, mark timing differences |
| `balanced` | AP/AR clerk | Submit for review |
| `under_review` | Finance manager | Approve or return with comments |
| `approved` | Finance manager | Lock reconciliation |
| `locked` | Nobody | Read-only, archived |

---

## 10. Bank Statement Import

### Supported Formats

| Format | Description | Typical Source |
|---|---|---|
| OFX/QFX | Open Financial Exchange — XML-based standard | Equity Bank, KCB, Co-op Bank online banking |
| MT940 | SWIFT standard — used for correspondent banking | International wire transfers, Standard Chartered |
| CSV | Configurable column mapping | Any bank with statement export |
| Excel (.xlsx) | Column-mapped spreadsheet | Manual statement extraction |

### CSV Column Mapping Configuration

Each bank account stores a `ColumnMap` that maps CSV column positions or headers to statement fields. This is configured once and reused for every import.

```json
{
  "date": { "column": "A", "format": "DD/MM/YYYY" },
  "description": { "column": "B" },
  "debit": { "column": "C" },
  "credit": { "column": "D" },
  "reference": { "column": "E" },
  "running_balance": { "column": "F" }
}
```

### Import Validation

On import, the system validates:
- Opening balance matches the prior period's reconciliation closing balance (within 0.01)
- If `running_balance` column is present, the computed running balance matches it throughout
- No duplicate lines (same date + amount + reference already imported for this bank account)
- All dates fall within the period's date range

A failed validation returns the specific row and field that caused the failure.

### Statement Importer Interface

```go
type StatementImporter interface {
    Format()  string
    Parse(r io.Reader, mapping ColumnMap) ([]StatementLine, StatementSummary, error)
}

type StatementLine struct {
    Date        time.Time
    Description string
    Amount      decimal.Decimal  // positive = credit to bank account; negative = debit
    Reference   string
    RawData     map[string]string  // original row for audit
}
```

---

## 11. Auto-Matching Engine

### Why Auto-Matching

Manual matching of hundreds of bank statement lines against GL entries is time-consuming and error-prone. The auto-matcher systematically applies a prioritised set of rules to propose or confirm matches, reducing the manual workload to only the genuinely ambiguous items.

### Matching Rules (Priority Order)

```go
type MatchRule interface {
    Name()       string
    Confidence() MatchConfidence  // auto | suggested | manual
    Match(line StatementLine, entries []GLEntry) *MatchCandidate
}
```

| Rule | Criteria | Confidence | Auto-Confirm? |
|---|---|---|---|
| M-01 Exact | Amount + Date (±0 days) + Reference exact match | `auto` | Yes (if threshold = `auto`) |
| M-02 Amount + Date | Amount exact + Date within ±2 business days | `suggested` | No — user confirms |
| M-03 Amount + Partial Ref | Amount exact + Reference contains substring of GL reference | `suggested` | No |
| M-04 Batch (1:N) | Bank total = sum of multiple GL entries (max 10, within tolerance) | `suggested` | No |
| M-05 No Match | No rule fired | — | Manual investigation required |

```go
type AutoMatcher struct {
    rules []MatchRule
    tolerance decimal.Decimal  // configurable rounding tolerance, default 0.01
}

func (m *AutoMatcher) Run(lines []StatementLine, glEntries []GLEntry) []MatchResult {
    results    := []MatchResult{}
    usedGLIDs  := map[uuid.UUID]bool{}

    for _, line := range lines {
        available := filterUnused(glEntries, usedGLIDs)
        var best *MatchCandidate
        for _, rule := range m.rules {
            if candidate := rule.Match(line, available); candidate != nil {
                best = candidate
                break
            }
        }
        result := MatchResult{StatementLine: line}
        if best != nil {
            result.Candidate  = best
            result.Confidence = best.Confidence
            for _, id := range best.GLEntryIDs {
                usedGLIDs[id] = true
            }
        }
        results = append(results, result)
    }
    return results
}
```

### Batch Match Rule (M-04)

The batch match finds subsets of GL entries that sum to a statement line amount. This covers the common case where a bank credits a single batch amount that corresponds to several individual customer receipts.

```go
type BatchMatchRule struct {
    Tolerance decimal.Decimal
}

func (r *BatchMatchRule) Match(line StatementLine, entries []GLEntry) *MatchCandidate {
    // Bounded knapsack: find subsets summing to line.Amount within tolerance
    // Limited to max 10 entries to prevent combinatorial explosion (2^10 = 1024 combinations)
    candidates := findSubsets(entries, line.Amount, r.Tolerance, 10)
    if len(candidates) == 1 {
        return &MatchCandidate{
            GLEntryIDs: candidates[0],
            Confidence: ConfidenceSuggested,
            RuleName:   "M-04 Batch Match",
        }
    }
    // If multiple subsets match, ambiguous — don't suggest
    return nil
}
```

### Match Result Types

Beyond matched/unmatched, the reconciler supports explicit classification of timing differences:

| Match Type | Meaning | GL Effect |
|---|---|---|
| `confirmed` | Statement line = GL entry | Mark both as matched |
| `timing_difference_deposit` | In GL but not yet on statement (deposit in transit) | Mark GL entry as timing difference; carry to next period |
| `timing_difference_cheque` | In GL but not yet presented (outstanding cheque) | Mark GL entry as timing difference; carry to next period |
| `bank_error` | On statement but should not be there (bank mistake) | Note for follow-up; no GL entry |
| `unrecorded` | On statement, no GL entry exists | Create GL entry (adjusting journal) |

---

## 12. Reconciliation Lifecycle

### Reconciliation Statement Output

The system produces a formal reconciliation statement that can be exported as PDF:

```
BANK RECONCILIATION STATEMENT
Account:    1112 – Checking Account Main (KES)
Period:     January 2025
Prepared:   Jane Waweru  |  Approved: John Kamau (Finance Manager)

Balance per Bank Statement:                       1,225,000
  Add: Deposits in Transit
    31 Jan — REC-0299 Customer payment              180,000
  Less: Outstanding Cheques
    15 Jan — CHQ-001290 Supplier payment            (85,000)
    22 Jan — CHQ-001295 Supplier payment            (60,000)
  Less: Bank Errors (reported to bank)
    20 Jan — Erroneous debit (ref ERR-2025-0120)    (20,000)
──────────────────────────────────────────────────────────
Adjusted Bank Balance:                            1,240,000

Balance per GL (Account 1112):                    1,225,000
  Add: GL adjustments posted during reconciliation
    31 Jan — Bank service fee                           500
    28 Jan — Interest income                        (14,500)
──────────────────────────────────────────────────────────
Adjusted GL Balance:                              1,240,000

DIFFERENCE:                                               0 ✓
```

### Outstanding Item Ageing

Outstanding items (deposits in transit and uncleared cheques) that survive reconciliation carry forward to the next period's reconciliation. The system tracks their age and applies flag thresholds:

- Outstanding > `outstanding_cheque_alert_days` (default 60): flagged as `requires_investigation`
- Outstanding > 180 days: flagged as `escalated`, notification sent to Finance Manager

---

## 13. Business Rules & Validation

### Period Rules

| Rule ID | Rule | Enforcement |
|---|---|---|
| `FIN-PRD-001` | Transaction date cannot be in the future | Application + DB trigger (via journal entry insert) |
| `FIN-PRD-002` | Transaction date must fall within a valid fiscal year | `ValidatePeriod` stage |
| `FIN-PRD-003` | Soft-closed period accepts entries only from `finance_manager` or `cfo` | Application layer (PeriodGate) |
| `FIN-PRD-004` | Hard-closed or locked period accepts no entries whatsoever | DB trigger + application |
| `FIN-PRD-005` | Locked period cannot be reopened via UI | Hard block — requires support escalation |
| `FIN-PRD-006` | Period cannot hard-close unless all mandatory checklist items pass | PeriodCloseWorkflow |
| `FIN-PRD-007` | Reopening a period must record the reason and authoriser | Audit log entry created on every reopen |
| `FIN-PRD-008` | A tenant must always have at least one open period | Application block on last-period close attempt |

### Bank Reconciliation Rules

| Rule ID | Rule | Enforcement |
|---|---|---|
| `FIN-REC-001` | Only one reconciliation per bank account per period | DB unique constraint |
| `FIN-REC-002` | Opening balance must equal the prior period's closing balance (within 0.01) | Import validation |
| `FIN-REC-003` | Reconciliation cannot be submitted unless `Difference` ≤ 0.01 | `Submit()` method check |
| `FIN-REC-004` | Preparer cannot be the approver | `Approve()` method check |
| `FIN-REC-005` | Period cannot hard-close without an approved reconciliation for every active bank account | Period close checklist |
| `FIN-REC-006` | Once locked, the reconciliation record is immutable | DB trigger + application |
| `FIN-REC-007` | Outstanding items > 60 days are auto-flagged | Background job on daily cron |
| `FIN-REC-008` | A bank account must have `status = active` to open a reconciliation | Application check |

### Alternatives Considered

**Alternative: Allow period reopen without reason.**
Rejected. Reopening a period after close is a significant control event that auditors specifically look for. Every reopen must be documented with a reason and the authorising user.

**Alternative: Allow the same person to prepare and approve the reconciliation.**
Rejected. This is a fundamental segregation of duties control. A person who reconciles their own work can cover up errors or fraud. The self-approval prohibition is enforced programmatically — it cannot be bypassed by configuration.

**Alternative: Make bank reconciliation optional (not required for period close).**
Available as a configuration option (`bank_rec_required_for_close = false`) for very small operators who manage a single cash account. However, disabling this removes a major control and is not recommended.

---

## 14. Performance & Storage

### Period Management Performance

| Query | Mechanism | Target P95 |
|---|---|---|
| Find period for a date (period gate) | Index on `(tenant_id, organisation_id, start_date, end_date)` | < 2ms |
| List all periods for fiscal year | Full tenant scan (~12 rows) | < 5ms |
| Run period close checklist (all checks) | Parallel queries against read replica | < 10s total |

The period gate query fires on every journal entry insert. At 1,000 inserts per day, this index is read 1,000 times per day — it must be fast. The index on `(tenant_id, organisation_id, start_date, end_date)` supports a range query (`start_date <= $date AND end_date >= $date`) efficiently.

### Bank Reconciliation Performance

| Query | Mechanism | Target P95 |
|---|---|---|
| Load GL entries for matching (1 month, 1 account) | Index on `(tenant_id, account_id, created_at)` | < 200ms |
| Auto-matcher (500 statement lines, 600 GL entries) | In-memory rule evaluation | < 2s |
| Batch match subset search (max 10 entries) | In-memory knapsack (1,024 max combinations) | < 50ms per line |
| Export reconciliation PDF | Template rendering + PDF generation | < 5s |

### Storage Estimates

| Table | Row Size | Rows/Year | Annual Storage |
|---|---|---|---|
| `fiscal_years` | ~200 bytes | 1 per org | Negligible |
| `accounting_periods` | ~300 bytes | 12 per org | ~3.6 KB/year |
| `bank_reconciliations` | ~400 bytes | 1 per bank account per month (e.g. 3 accounts = 36/year) | ~14 KB/year |
| `bank_statement_lines` | ~250 bytes | ~2,000 lines/month/account | ~18 MB/year (3 accounts) |
| `reconciliation_matches` | ~200 bytes | ~2,000/month | ~14 MB/year |

Total reconciliation-related storage: approximately **32 MB per year** for a 3-bank-account operator. This grows linearly with number of bank accounts and transaction volume.

---

## 15. Database Schema

```sql
CREATE TABLE fiscal_years (
    tenant_id       UUID        NOT NULL REFERENCES tenants(id),
    id              UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    organisation_id UUID        NOT NULL,
    name            TEXT        NOT NULL,
    start_date      DATE        NOT NULL,
    end_date        DATE        NOT NULL,
    is_closed       BOOLEAN     NOT NULL DEFAULT FALSE,
    is_locked       BOOLEAN     NOT NULL DEFAULT FALSE,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, organisation_id, name),
    CONSTRAINT chk_fy_dates CHECK (end_date > start_date)
);

ALTER TABLE fiscal_years ENABLE ROW LEVEL SECURITY;
ALTER TABLE fiscal_years FORCE  ROW LEVEL SECURITY;
CREATE POLICY fy_app ON fiscal_years FOR ALL TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

-- ── accounting_periods ────────────────────────────────────────────────────────

CREATE TABLE accounting_periods (
    tenant_id       UUID        NOT NULL REFERENCES tenants(id),
    id              UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    organisation_id UUID        NOT NULL,
    fiscal_year_id  UUID        NOT NULL REFERENCES fiscal_years(id),
    name            TEXT        NOT NULL,
    start_date      DATE        NOT NULL,
    end_date        DATE        NOT NULL,
    status          TEXT        NOT NULL DEFAULT 'open'
                        CHECK (status IN ('open','soft_closed','hard_closed','locked')),
    closed_at       TIMESTAMPTZ,
    closed_by       UUID,
    locked_at       TIMESTAMPTZ,
    locked_by       UUID,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, organisation_id, start_date)
);

CREATE INDEX idx_periods_date ON accounting_periods
    (tenant_id, organisation_id, start_date, end_date);
CREATE INDEX idx_periods_open ON accounting_periods (tenant_id, status)
    WHERE status = 'open';

ALTER TABLE accounting_periods ENABLE ROW LEVEL SECURITY;
ALTER TABLE accounting_periods FORCE  ROW LEVEL SECURITY;
CREATE POLICY ap_app ON accounting_periods FOR ALL TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

-- ── bank_reconciliations ──────────────────────────────────────────────────────

CREATE TABLE bank_reconciliations (
    tenant_id             UUID          NOT NULL REFERENCES tenants(id),
    id                    UUID          NOT NULL DEFAULT gen_random_uuid(),
    created_at            TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    organisation_id       UUID          NOT NULL,
    bank_account_id       UUID          NOT NULL,
    period_id             UUID          NOT NULL REFERENCES accounting_periods(id),
    statement_opening_bal NUMERIC(18,4) NOT NULL,
    statement_closing_bal NUMERIC(18,4) NOT NULL,
    gl_closing_balance    NUMERIC(18,4) NOT NULL DEFAULT 0,
    adjusted_bank_balance NUMERIC(18,4) NOT NULL DEFAULT 0,
    adjusted_gl_balance   NUMERIC(18,4) NOT NULL DEFAULT 0,
    difference            NUMERIC(18,4) GENERATED ALWAYS AS
                              (adjusted_bank_balance - adjusted_gl_balance) STORED,
    status                TEXT          NOT NULL DEFAULT 'open'
                              CHECK (status IN ('open','in_progress','balanced','under_review','approved','locked')),
    prepared_by           UUID          NOT NULL,
    approved_by           UUID,
    locked_at             TIMESTAMPTZ,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, bank_account_id, period_id)
);

CREATE INDEX idx_bank_recs_period ON bank_reconciliations (tenant_id, period_id);

ALTER TABLE bank_reconciliations ENABLE ROW LEVEL SECURITY;
ALTER TABLE bank_reconciliations FORCE  ROW LEVEL SECURITY;
CREATE POLICY br_app ON bank_reconciliations FOR ALL TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

-- ── bank_statement_lines ──────────────────────────────────────────────────────

CREATE TABLE bank_statement_lines (
    tenant_id          UUID          NOT NULL REFERENCES tenants(id),
    id                 UUID          NOT NULL DEFAULT gen_random_uuid(),
    created_at         TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    reconciliation_id  UUID          NOT NULL REFERENCES bank_reconciliations(id),
    statement_date     DATE          NOT NULL,
    description        TEXT,
    amount             NUMERIC(18,4) NOT NULL,  -- positive = credit, negative = debit
    reference          TEXT,
    match_status       TEXT          NOT NULL DEFAULT 'unmatched'
                           CHECK (match_status IN ('unmatched','matched','timing_difference','bank_error','unrecorded')),
    match_note         TEXT,
    raw_data           JSONB,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id)
);

CREATE INDEX idx_bsl_reconciliation ON bank_statement_lines (tenant_id, reconciliation_id);
```

---

## 16. API Reference

### Fiscal Years & Periods

```
GET    /finance/fiscal-years                         List fiscal years
POST   /finance/fiscal-years                         Create fiscal year (auto-generates 12 periods)
GET    /finance/periods                              List periods for current fiscal year (?fiscal_year_id=)
GET    /finance/periods/{id}                         Get period + close checklist status
POST   /finance/periods/{id}/soft-close             Open → SoftClosed
POST   /finance/periods/{id}/hard-close             SoftClosed → HardClosed (runs checklist)
POST   /finance/periods/{id}/reopen                 Any closed state → Open (logged, reason required)
POST   /finance/periods/{id}/lock                   HardClosed → Locked (cfo only)
```

### Bank Reconciliation

```
GET    /finance/reconciliations                      List (?bank_account_id=, ?period_id=, ?status=)
POST   /finance/reconciliations                      Open new reconciliation workspace
POST   /finance/reconciliations/{id}/import          Import bank statement (multipart/form-data)
GET    /finance/reconciliations/{id}/matches         Get auto-match suggestions
POST   /finance/reconciliations/{id}/matches         Confirm matches / classify timing differences / create GL adjustments
DELETE /finance/reconciliations/{id}/matches/{matchId} Unmatch a confirmed match
POST   /finance/reconciliations/{id}/submit          InProgress → UnderReview (checks Difference ≤ 0.01)
POST   /finance/reconciliations/{id}/approve         UnderReview → Approved (different user than preparer)
POST   /finance/reconciliations/{id}/lock            Approved → Locked (linked to period hard-close)
GET    /finance/reconciliations/{id}/statement       Download reconciliation statement as PDF
```

**Create reconciliation:**
```json
{
  "bank_account_id": "uuid",
  "period_id": "uuid",
  "statement_opening_balance": 441350.00,
  "statement_closing_balance": 1225000.00
}
```

**Confirm matches:**
```json
{
  "matches": [
    {
      "statement_line_id": "uuid",
      "gl_entry_ids": ["uuid1"],
      "match_type": "confirmed"
    },
    {
      "statement_line_id": "uuid2",
      "gl_entry_ids": [],
      "match_type": "timing_difference_deposit",
      "note": "Credited by bank on Feb 1"
    },
    {
      "statement_line_id": "uuid3",
      "gl_entry_ids": [],
      "match_type": "unrecorded",
      "gl_entry": {
        "account_code": "7740",
        "debit": 500.00,
        "description": "Bank service fee Jan 2025"
      }
    }
  ]
}
```

---

## 17. Feature Flags & Configuration

| Flag | Type | Default | Description |
|---|---|---|---|
| `bank_rec_required_for_close` | bool | `true` | Period cannot hard-close without approved bank rec for every active bank account |
| `bank_rec_tolerance` | decimal | `0.01` | Maximum acceptable difference (KES) to allow submission of reconciliation |
| `outstanding_cheque_alert_days` | int | `60` | Flag outstanding items older than N days |
| `auto_match_confidence_threshold` | enum | `auto` | `auto` — auto-confirm exact matches; `suggested` — all matches require user confirmation; `manual` — no auto-matching |
| `bank_statement_import_formats` | string[] | `["ofx","mt940","csv"]` | Enabled import formats |
| `batch_match_max_entries` | int | `10` | Maximum GL entries to combine in M-04 batch match |
| `period_close_mode` | enum | `workflow` | `workflow` — use Temporal PeriodCloseWorkflow; `sync` — run checks synchronously in the API request |
| `fx_revaluation_required_for_close` | bool | `true` | Block hard-close if unrevalued FC balances exist |
| `subsidiary_reconcile_required_for_close` | bool | `true` | AR/AP subsidiary must reconcile to GL control |
| `ar_sub_rec_tolerance` | decimal | `0.01` | Acceptable difference between AR control and subsidiary total |
| `fiscal_year_period_type` | enum | `monthly` | `monthly` (12 periods) or `quarterly` (4 periods) per fiscal year |
| `allow_period_reopen_self` | bool | `false` | If false, the person who closed the period cannot be the one who reopens it |

---

## 18. v1.0 Rollout Assessment

### Must Have at v1.0

- Full period status machine with all four states
- Period gate at both application and DB trigger level (with `FOR SHARE` fix)
- Period close checklist (all mandatory checks)
- Bank reconciliation workspace (full lifecycle)
- CSV bank statement import (OFX and MT940 can be v1.1)
- Manual statement matching
- Auto-matcher M-01 (exact match) and M-02 (amount + date)

### Can Be Deferred

| Feature | Suggested Release |
|---|---|
| OFX/MT940 import formats | v1.1 |
| Auto-matcher M-04 (batch match) | v1.1 |
| Temporal period close workflow | v1.1 (synchronous API-based close acceptable at v1.0) |
| Reconciliation PDF export | v1.1 |
| Outstanding item ageing background job | v1.1 |

### Never Defer

- DB unique constraint on `(tenant_id, bank_account_id, period_id)` in `bank_reconciliations`
- Self-approval prohibition on reconciliation
- Period gate trigger with `FOR SHARE`
- Audit log on every period reopen event
- `difference` as a generated column (prevents inconsistent state)

---

*End of FILE-04. Proceeding to FILE-05: Accounts Receivable, Accounts Payable, Payments & Petty Cash.*
