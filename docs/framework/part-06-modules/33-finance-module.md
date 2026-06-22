---
title: "Chapter 33: Finance Module"
part: "Part VI — Built-In Modules"
chapter: 33
section: "33-finance-module"
related:
  - "[Chapter 29: Saga Pattern](../part-05-workflow/29-saga-pattern.md)"
  - "[Chapter 35: Inventory Module](35-inventory-module.md)"
---

# Chapter 33: Finance Module

The Finance module provides double-entry bookkeeping, period management, financial statements, and multi-currency support. It is the foundation module — all other Awo modules that move value (Inventory, Payroll, Sales) post GL entries through the Finance module.

---

## 33.1. Chart of Accounts

### 33.1.1. Account Types

| Type | Normal balance | Kenya examples |
|---|---|---|
| Asset | Debit | Cash, Bank, Accounts Receivable, Inventory, Equipment |
| Liability | Credit | Accounts Payable, VAT Payable, Loans |
| Equity | Credit | Share Capital, Retained Earnings |
| Income | Credit | Sales Revenue, Service Fees |
| Expense | Debit | Cost of Goods Sold, Salaries, Rent, Fuel |

The account type determines how balances are displayed in financial statements:
- Assets increase with debits, decrease with credits
- Liabilities and Equity increase with credits, decrease with debits

### 33.1.2. Account Hierarchy — Materialised Path

Accounts are organised in a hierarchy: `1000 Assets → 1100 Current Assets → 1110 Bank Accounts → 1111 KCB Current Account`. This maps directly to the materialised path pattern from Chapter 6:

```sql
CREATE TABLE accounts (
    id          uuid PRIMARY KEY,
    code        varchar(20) UNIQUE NOT NULL,
    name        varchar(255) NOT NULL,
    type        varchar(20) NOT NULL,  -- Asset, Liability, Equity, Income, Expense
    parent_id   uuid REFERENCES accounts(id),
    path        varchar(4096),         -- /root-id/parent-id/this-id/
    depth       int NOT NULL DEFAULT 0,
    is_group    bool NOT NULL DEFAULT false,
    currency    varchar(3) NOT NULL DEFAULT 'KES',
    is_active   bool NOT NULL DEFAULT true
);
```

Leaf accounts (non-group) receive GL postings. Group accounts display aggregate balances.

### 33.1.3. Kenya Chart of Accounts — KRA-Aligned Account Groups

The default Kenya chart of accounts follows KRA's prescribed structure for income tax reporting:

```
1000 Assets
  1100 Current Assets
    1110 Bank and Cash
      1111 KCB Current Account
      1112 Equity Bank Account
      1113 M-PESA Business Account
      1114 Cash on Hand
    1120 Accounts Receivable
    1130 Inventory
    1140 Prepaid Expenses
  1200 Fixed Assets
    1210 Equipment
    1220 Furniture
    1230 Motor Vehicles
    1299 Accumulated Depreciation

2000 Liabilities
  2100 Current Liabilities
    2110 Accounts Payable
    2120 VAT Payable (16%)
    2130 Withholding Tax Payable
    2140 PAYE Payable
    2150 NHIF Payable
    2160 NSSF Payable
  2200 Long-term Liabilities
    2210 Bank Loans

3000 Equity
  3100 Share Capital
  3200 Retained Earnings

4000 Income
  4100 Sales Revenue
  4200 Service Revenue
  4300 Other Income

5000 Expenses
  5100 Cost of Goods Sold
  5200 Employee Costs
    5210 Basic Salaries
    5220 NHIF - Employer
    5230 NSSF - Employer
  5300 Operating Expenses
    5310 Rent
    5320 Utilities
    5330 Fuel
    5340 Communication
  5400 Finance Costs
    5410 Bank Charges
    5420 Interest Expense
```

### 33.1.4. Account Numbering Conventions

- First digit: account type (1=Asset, 2=Liability, 3=Equity, 4=Income, 5=Expense)
- Second digit: sub-type group
- Third digit: specific category
- Fourth digit: individual account

Tenants can customise account codes and names but cannot change the account type structure.

### 33.1.5. Cost Centres — Departmental Allocation

Cost centres allow GL postings to be attributed to departments:

```go
type CostCentre struct {
    ent.Schema
}

func (CostCentre) Fields() []ent.Field {
    return []ent.Field{
        field.String("code").MaxLen(20).Unique(),
        field.String("name").MaxLen(255),
        field.UUID("parent_id", uuid.UUID{}).Optional(),
        field.String("path").MaxLen(4096),
    }
}
```

GL entries can carry an optional `cost_centre_id`. The P&L report can break down expenses by cost centre.

---

## 33.2. Journal Entry and Double-Entry Enforcement

### 33.2.1. Journal Entry Entity

```go
// JournalEntry — the header
type JournalEntry struct {
    ID          uuid.UUID
    PostingDate time.Time
    Reference   string     // document that triggered this entry
    Description string
    Status      string     // draft | posted | reversed
    CreatedBy   uuid.UUID
    PostedAt    *time.Time
    ReversedAt  *time.Time
    SourceType  string     // Invoice, Payment, StockEntry, etc.
    SourceID    uuid.UUID
}

// GLLine — the debit/credit lines
type GLLine struct {
    ID              uuid.UUID
    JournalEntryID  uuid.UUID
    AccountID       uuid.UUID
    CostCentreID    *uuid.UUID
    Debit           decimal.Decimal  // exactly one of debit/credit is non-zero
    Credit          decimal.Decimal
    Description     string
}
```

### 33.2.2. Double-Entry Validation — Debits Must Equal Credits

Enforced in the `before_save` hook on `JournalEntry`:

```go
func validateDoubleEntry(r *schema.EntityRecord) error {
    lines := r.GetRelated("lines")
    var totalDebit, totalCredit decimal.Decimal
    for _, line := range lines {
        totalDebit = totalDebit.Add(line.GetDecimal("debit"))
        totalCredit = totalCredit.Add(line.GetDecimal("credit"))
    }
    if !totalDebit.Equal(totalCredit) {
        return errs.NewBusinessError("UNBALANCED_ENTRY",
            "journal entry is unbalanced: debits %s, credits %s",
            totalDebit.StringFixed(2), totalCredit.StringFixed(2))
    }
    if len(lines) < 2 {
        return errs.NewBusinessError("MINIMUM_LINES",
            "journal entry must have at least 2 lines")
    }
    return nil
}
```

### 33.2.3. GL Posting from Workflows — The `PostJournalEntry` Activity

```go
func (a *FinanceActivities) PostJournalEntry(ctx context.Context, input PostJournalInput) (uuid.UUID, error) {
    ctx = tenant.WithIDContext(ctx, input.TenantID)

    // Idempotency: don't post twice for the same source document
    existing, err := a.journalEntryRepo.Exists(ctx,
        filter.Eq("source_type", input.SourceType).
        And(filter.Eq("source_id", input.SourceID)).
        And(filter.Eq("status", "posted")),
    )
    if err != nil {
        return uuid.Nil, err
    }
    if existing {
        // Already posted — return the existing entry ID
        entry, _ := a.journalEntryRepo.Query(ctx,
            filter.Eq("source_id", input.SourceID))
        return entry[0].ID, nil
    }

    // Create and post the journal entry
    entry, err := a.journalEntryRepo.Create(ctx, JournalEntryCreate{
        PostingDate: input.PostingDate,
        Reference:   input.Reference,
        Description: input.Description,
        SourceType:  input.SourceType,
        SourceID:    input.SourceID,
        Lines:       input.Lines,
    })
    if err != nil {
        return uuid.Nil, err
    }

    // Mark as posted
    updated, err := a.journalEntryRepo.Update(ctx, entry.ID, JournalEntryUpdate{
        Status:   "posted",
        PostedAt: ptr.Time(time.Now()),
    })
    return updated.ID, err
}
```

### 33.2.4. Reversal Entries

When a posted journal entry must be reversed (invoice cancelled, payment voided):

```go
func (a *FinanceActivities) ReverseJournalEntry(ctx context.Context, entryID uuid.UUID, reason string) error {
    original, err := a.journalEntryRepo.Get(ctx, entryID)
    if err != nil {
        return err
    }

    if original.Status != "posted" {
        return errs.NewBusinessError("NOT_POSTED", "only posted entries can be reversed")
    }
    if original.ReversedAt != nil {
        return nil  // already reversed — idempotent
    }

    // Create reversal entry with flipped debits/credits
    reversalLines := flipDebitsCredits(original.Lines)
    reversal, err := a.journalEntryRepo.Create(ctx, JournalEntryCreate{
        PostingDate: time.Now(),
        Reference:   "REV:" + original.Reference,
        Description: fmt.Sprintf("Reversal of %s: %s", original.Reference, reason),
        SourceType:  "Reversal",
        SourceID:    original.ID,
        Lines:       reversalLines,
    })
    if err != nil {
        return err
    }

    // Mark original as reversed
    _, err = a.journalEntryRepo.Update(ctx, entryID, JournalEntryUpdate{
        ReversedAt: ptr.Time(time.Now()),
        ReversalID: &reversal.ID,
    })
    return err
}
```

---

## 33.3. Period Management

### 33.3.1. Fiscal Year and Period Entities

```go
// FiscalYear: e.g. "2025-26" (July 2025 - June 2026 for Kenya)
// Period: e.g. "2025-07" (July 2025)
type AccountingPeriod struct {
    FiscalYear string
    Name       string    // "July 2025"
    StartDate  time.Time
    EndDate    time.Time
    Status     string    // open | closing | closed
    ClosedAt   *time.Time
    ClosedBy   *uuid.UUID
}
```

Kenya's fiscal year runs July to June, aligned with the government budget cycle.

### 33.3.2. Period Closing Workflow

```go
func PeriodCloseWorkflow(ctx workflow.Context, params PeriodCloseParams) error {
    // 1. Validate no unposted entries
    workflow.ExecuteActivity(ctx, activities.ValidateNoPendingEntries, params.PeriodID)

    // 2. Compute period-end balances
    workflow.ExecuteActivity(ctx, activities.ComputePeriodBalances, params.PeriodID)

    // 3. Post depreciation entries (if configured)
    workflow.ExecuteActivity(ctx, activities.PostDepreciationEntries, params.PeriodID)

    // 4. Lock the period
    workflow.ExecuteActivity(ctx, activities.LockAccountingPeriod, params.PeriodID)

    // 5. Carry forward balances to next period
    workflow.ExecuteActivity(ctx, activities.CarryForwardBalances, params.PeriodID)

    return nil
}
```

### 33.3.3. Posting Into a Closed Period

Blocked by default in `before_save` on `JournalEntry`:

```go
func validatePeriodOpen(ctx context.Context, postingDate time.Time) error {
    period, err := periodRepo.FindForDate(ctx, postingDate)
    if err != nil {
        return err
    }
    if period.Status == "closed" {
        if !actor.HasPermission(ctx, "finance:post_to_closed_period") {
            return errs.NewBusinessError("PERIOD_CLOSED",
                "accounting period %s is closed; cannot post new entries",
                period.Name)
        }
        // Log the exception for audit
        auditLog.Record(ctx, AuditEntry{
            Action: "POST_TO_CLOSED_PERIOD",
            Note:   fmt.Sprintf("backdated posting to closed period %s", period.Name),
        })
    }
    return nil
}
```

---

## 33.4. Financial Statements

### 33.4.1. Trial Balance

```go
func (r *GLReportService) TrialBalance(ctx context.Context, input TrialBalanceInput) (TrialBalanceReport, error) {
    rows, err := r.glRepo.Aggregate(ctx,
        filter.Lte("posting_date", input.AsOfDate).
        And(filter.Eq("status", "posted")),
        aggregate.GroupBy("account_id").
            Sum("debit").
            Sum("credit"),
    )
    // Compute net balance per account
    // Return sorted by account code
}
```

### 33.4.2. Profit and Loss

The P&L report aggregates Income (4xxx) and Expense (5xxx) accounts for a period:

```
Revenue
  Sales Revenue          KES 1,234,567
  Service Revenue        KES    234,000
  Total Revenue          KES 1,468,567

Cost of Goods Sold       KES   (654,321)
Gross Profit             KES   814,246

Operating Expenses
  Salaries               KES   (180,000)
  Rent                   KES    (45,000)
  Fuel                   KES    (23,400)
  Total OpEx             KES   (248,400)

Operating Profit         KES   565,846
Finance Costs            KES    (12,000)
Net Profit Before Tax    KES   553,846
```

### 33.4.3. Balance Sheet

Follows the Kenya IFRS-aligned format: Assets = Liabilities + Equity. Computed as at a point in time (all transactions up to `as_of_date`).

---

## 33.5. Reconciliation Health Subsystem

### 33.5.1. What Reconciliation Health Tracks

The reconciliation health subsystem detects anomalies before they become problems:
- **Bank vs GL**: GL bank balance vs last bank statement balance
- **Wetstock vs GL**: computed fuel stock from GL vs dip readings
- **Accounts Receivable aging**: overdue invoices and their concentration

### 33.5.2. Anomaly Detection

The nightly GL period check workflow includes statistical anomaly detection:

```go
func detectGLAnomalies(ctx context.Context, periodID uuid.UUID) ([]Anomaly, error) {
    // Get last 90 days of daily totals for each account
    // Compute mean and standard deviation
    // Flag any day where posting total > mean + 3*stddev (3-sigma rule)
    // Return anomalies sorted by severity
}
```

### 33.5.3. Reconciliation Health Dashboard

An amis dashboard page showing:
- Reconciliation status per bank account (green/amber/red)
- Last reconciled date and variance amount
- Number of unreconciled transactions
- Top anomalies requiring review

---

## 33.6. Multi-Currency Support

### 33.6.1. KES as Base Currency

All GL balances are stored in KES. Foreign currency transactions are stored with both the original currency amount and the KES equivalent at the posting date exchange rate.

### 33.6.2. Exchange Rate Entity

```go
type ExchangeRate struct {
    BaseCurrency   string          // "KES"
    QuoteCurrency  string          // "USD", "EUR", "GBP", "UGX"
    Rate           decimal.Decimal // KES per 1 unit of quote currency
    RateDate       time.Time
    Source         string          // "CBK" (Central Bank of Kenya), "manual"
}
```

Exchange rates are loaded daily from the CBK (Central Bank of Kenya) API via a Temporal scheduled workflow.

### 33.6.3. Currency Gain/Loss

When a foreign currency receivable is settled, the difference between the original KES equivalent and the settlement KES amount is posted to `5420 Foreign Exchange Gain/Loss`:

```go
func postCurrencyGainLoss(ctx context.Context, invoice *Invoice, payment *Payment) error {
    originalKES := invoice.ForeignAmount.Mul(invoice.ExchangeRate)
    settlementKES := payment.ForeignAmount.Mul(payment.ExchangeRate)
    gainLoss := settlementKES.Sub(originalKES)

    if gainLoss.IsZero() {
        return nil
    }

    return postJournalEntry(ctx, JournalEntryCreate{
        Description: fmt.Sprintf("FX gain/loss on invoice %s", invoice.Number),
        Lines: []GLLineCreate{
            {AccountID: fxGainLossAccountID,
             Debit:  max(decimal.Zero, gainLoss.Neg()),
             Credit: max(decimal.Zero, gainLoss)},
            {AccountID: bankAccountID,
             Debit:  max(decimal.Zero, gainLoss),
             Credit: max(decimal.Zero, gainLoss.Neg())},
        },
    })
}
```
