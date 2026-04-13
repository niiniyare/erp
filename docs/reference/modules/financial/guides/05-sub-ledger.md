# Awo ERP — Finance Module
## FILE-05: Accounts Receivable, Accounts Payable, Payments & Petty Cash

**Document Version:** 2.0.0  
**Series:** FILE-05 of 10  
**Depends On:** FILE-01 (Domain Model), FILE-02 (COA), FILE-03 (Journal Pipeline)  
**Depended On By:** FILE-04 (Period Close Checklist), FILE-08 (Reports — AR/AP Aging)

---

## Table of Contents

1. [Why AR, AP, Payments and Petty Cash Exist](#1-why-ar-ap-payments-and-petty-cash-exist)
2. [What You Lose Without Them](#2-what-you-lose-without-them)
3. [How ERPNext, NetSuite and QuickBooks Handle These](#3-how-erpnext-netsuite-and-quickbooks-handle-these)
4. [Accounts Receivable Sub-Ledger](#4-accounts-receivable-sub-ledger)
5. [AR Aging & Collections](#5-ar-aging--collections)
6. [Accounts Payable Sub-Ledger](#6-accounts-payable-sub-ledger)
7. [AP Aging & Payment Scheduling](#7-ap-aging--payment-scheduling)
8. [Payment Runs (Batch Supplier Payments)](#8-payment-runs-batch-supplier-payments)
9. [Petty Cash Imprest System](#9-petty-cash-imprest-system)
10. [Business Rules & Validation](#10-business-rules--validation)
11. [Performance & Storage](#11-performance--storage)
12. [Database Schema](#12-database-schema)
13. [API Reference](#13-api-reference)
14. [Feature Flags & Configuration](#14-feature-flags--configuration)
15. [v1.0 Rollout Assessment](#15-v10-rollout-assessment)

---

## 1. Why AR, AP, Payments and Petty Cash Exist

### Accounts Receivable

The GL shows a single balance in the "Accounts Receivable" control account — for example, 3,450,000 KES. This tells you how much you are owed in total, but not who owes you, how much each customer owes, or how old each balance is. The AR sub-ledger provides this breakdown.

Without the AR sub-ledger, you cannot chase a specific customer for payment because you cannot see their individual balance. You cannot tell which customers are 90 days overdue without manually going through every invoice. You cannot make credit decisions because you don't know a customer's exposure.

### Accounts Payable

The mirror image: the GL shows the total payable balance, but the AP sub-ledger shows exactly which suppliers you owe, on which invoices, due on which dates. This is what drives payment scheduling, supplier relationship management, and early payment discount capture.

### Payment Runs

Paying suppliers one by one — a separate bank transfer for each invoice — is operationally inefficient and error-prone. A payment run batches all payments due on or before a date, generates a bank-compatible file for bulk upload, and posts all the corresponding GL entries in a single confirmed step after the bank confirms processing.

### Petty Cash

Every organisation has small miscellaneous expenditures that are impractical to pay by bank transfer — courier fees, small office supplies, parking charges. The imprest petty cash system provides a controlled float with clear replenishment accounting. Without it, these expenses either go unrecorded (balance sheet error) or are paid through the main bank account with disproportionate administrative overhead.

---

## 2. What You Lose Without Them

| Capability | Without AR Sub-Ledger | Without AP Sub-Ledger | Without Payment Runs | Without Petty Cash |
|---|---|---|---|---|
| Per-customer balance | Manual spreadsheet | — | — | — |
| Overdue tracking | No visibility | No visibility | — | — |
| Credit limit enforcement | Impossible | — | — | — |
| Supplier payment scheduling | — | Manual calendar | — | — |
| Bulk payment file for bank | — | — | Individual transfers per invoice | — |
| Small expense recording | — | — | — | Unrecorded or over-processed |
| GL control reconciliation | Manual | Manual | — | — |
| Period close readiness | No check possible | No check possible | — | — |

---

## 3. How ERPNext, NetSuite and QuickBooks Handle These

### ERPNext

**AR/AP:** ERPNext maintains AR and AP sub-ledgers via the `GL Entry` table itself — every customer/supplier-related GL entry carries a `party_type` (Customer or Supplier) and `party` field. The sub-ledger is effectively a filtered view of GL entries. This works but means the GL Entry table carries sub-ledger data, creating coupling between the two levels.

**Payment runs:** ERPNext has a "Payment Order" DocType in some regional versions that approximates batch payment processing. The core system's "Payment Entry" is one-at-a-time.

**Petty cash:** No dedicated petty cash module. Handled via journal entries against a petty cash account.

### NetSuite

**AR/AP:** NetSuite has formal Customer and Vendor sub-ledgers completely separate from the GL. AR aging is a first-class report. Credit holds can be placed programmatically. Payment matching is sophisticated.

**Payment runs:** NetSuite has a "Bill Payment" batch process and supports NACHA/BACS/SEPA file generation. Very strong but complex to configure.

**Petty cash:** Handled through "Expense Reports" rather than a petty cash imprest. Different conceptual model — suited to large organisations with staff expense claims, less suited to a petrol station with a physical cash tin.

### QuickBooks

**AR/AP:** First-class features — customer and supplier balances are well-presented. Aging reports are clear. Credit limit tracking exists in QuickBooks Online Advanced.

**Payment runs:** QuickBooks has "Pay Bills" which batches AP payments. Supports cheque printing and basic ACH. No East African bank file format support.

**Petty cash:** Supported via a petty cash bank account with manual reconciliation. No formal imprest tracking.

**What Awo does differently from all three:** The AR/AP sub-ledger is explicitly reconciled to the GL control account as part of the period close checklist, and the reconciliation status is surfaced in real time. The payment run supports Kenyan bank file formats (Equity EFT, KCB RTGS) out of the box.

---

## 4. Accounts Receivable Sub-Ledger

### What the AR Sub-Ledger Contains

Each row in `ar_subsidiary_entries` represents one open (or recently settled) invoice line for a specific customer. The sub-ledger is updated by the `SubsidiaryLedger.applyAR()` method inside the `PostToGL` transaction.

```go
type ARSubsidiaryEntry struct {
    ID               uuid.UUID
    TenantID         uuid.UUID
    OrganisationID   uuid.UUID
    CustomerID       uuid.UUID
    JournalLineID    uuid.UUID      // the specific line in journal_lines
    InvoiceRef       string
    TransactionDate  time.Time
    DueDate          time.Time
    Currency         string
    OriginalAmount   decimal.Decimal // in transaction currency
    FunctionalAmount decimal.Decimal // in functional currency (KES)
    SettledAmount    decimal.Decimal
    IsSettled        bool
    AsAtDate         time.Time       // denormalised for aging queries
}
```

When a customer payment arrives, the corresponding AR entry is marked as settled (partially or fully) and `SettledAmount` is updated. When fully settled, `IsSettled = true`.

### Customer Balance

```go
type CustomerBalance struct {
    CustomerID     uuid.UUID
    CustomerName   string
    Currency       string
    OpenBalance    decimal.Decimal  // sum of unsettled functional amounts
    OpenBalanceFC  decimal.Decimal  // sum of unsettled transaction currency amounts
    OldestDueDate  *time.Time
    InvoiceCount   int
}
```

The `CustomerBalance` is computed from `ar_subsidiary_entries WHERE is_settled = FALSE` — a single indexed query per customer, or a GROUP BY for a full list.

### AR Control Reconciliation

This is the check performed during period close. The AR control account balance in `account_balances` must equal the sum of all unsettled AR sub-ledger entries.

```go
func (s *SubsidiaryLedger) ReconcileAR(ctx context.Context, orgID uuid.UUID, asAt time.Time) (*ReconciliationResult, error) {
    controlBalance, err := s.getControlAccountBalance(ctx, orgID, "receivable", asAt)
    if err != nil {
        return nil, err
    }
    subsidiaryTotal, err := s.getARSubsidiaryTotal(ctx, orgID, asAt)
    if err != nil {
        return nil, err
    }
    diff := controlBalance.Sub(subsidiaryTotal)
    return &ReconciliationResult{
        ControlBalance:  controlBalance,
        SubsidiaryTotal: subsidiaryTotal,
        Difference:      diff,
        IsReconciled:    diff.Abs().LessThanOrEqual(decimal.NewFromFloat(0.01)),
    }, nil
}
```

A non-zero difference indicates one of: a direct GL posting to the AR control account (bypassing the sub-ledger), a failed sub-ledger update, or a multi-currency rounding issue. The difference report surfaces the individual customer balances and the control total side by side to help pinpoint the cause.

---

## 5. AR Aging & Collections

### Aging Buckets

The AR aging report groups open customer balances by how far past their due date they are. The buckets are configurable but default to: Current (not yet due), 1–30 days, 31–60 days, 61–90 days, 91+ days.

The aging query runs against a materialised view on the read replica to avoid impacting the transactional database:

```sql
CREATE OR REPLACE VIEW v_ar_aging AS
SELECT
    s.tenant_id,
    s.customer_id,
    c.name                                                              AS customer_name,
    SUM(s.functional_amount) FILTER (WHERE s.days_overdue <= 0)       AS current_amount,
    SUM(s.functional_amount) FILTER (WHERE s.days_overdue BETWEEN 1  AND 30)  AS days_30,
    SUM(s.functional_amount) FILTER (WHERE s.days_overdue BETWEEN 31 AND 60)  AS days_60,
    SUM(s.functional_amount) FILTER (WHERE s.days_overdue BETWEEN 61 AND 90)  AS days_90,
    SUM(s.functional_amount) FILTER (WHERE s.days_overdue > 90)       AS over_90,
    SUM(s.functional_amount)                                           AS total
FROM ar_subsidiary_entries s
JOIN customers c ON c.id = s.customer_id
WHERE s.tenant_id  = current_setting('app.current_tenant')::UUID
  AND s.is_settled = FALSE
GROUP BY s.tenant_id, s.customer_id, c.name;

ALTER VIEW v_ar_aging SET (security_invoker = true);
```

`days_overdue` is a computed column: `CURRENT_DATE - s.due_date`.

### Bad Debt Provision

When a customer balance becomes doubtful, the finance team posts a provision entry:

```
Dr. Bad Debt Expense (expense)                [Amount]
    Cr. Allowance for Doubtful Accounts (contra-asset)  [Amount]
```

This reduces the net AR balance on the balance sheet without touching the customer's sub-ledger balance. The sub-ledger balance remains open until formally written off.

When written off (confirmed uncollectable):

```
Dr. Allowance for Doubtful Accounts   [Amount]
    Cr. Accounts Receivable (control)       [Amount]
```

At this point the sub-ledger entry is also marked as settled with `write_off = true` and the AR control account balance drops. The provision entry absorbed the P&L impact; the write-off is balance-sheet-only.

---

## 6. Accounts Payable Sub-Ledger

### What the AP Sub-Ledger Contains

Mirror image of AR. Each row represents an open supplier invoice.

```go
type APSubsidiaryEntry struct {
    ID               uuid.UUID
    TenantID         uuid.UUID
    OrganisationID   uuid.UUID
    SupplierID       uuid.UUID
    JournalLineID    uuid.UUID
    InvoiceRef       string        // supplier's invoice number
    TransactionDate  time.Time
    DueDate          time.Time
    Currency         string
    OriginalAmount   decimal.Decimal
    FunctionalAmount decimal.Decimal
    SettledAmount    decimal.Decimal
    IsSettled        bool
    DiscountDueDate  *time.Time    // early payment discount deadline
    DiscountAmount   *decimal.Decimal
}
```

### Early Payment Discount

Some suppliers offer a discount for early payment (e.g. 2% if paid within 10 days, net 30 days). The AP sub-ledger tracks the `DiscountDueDate` and `DiscountAmount`. If a payment is processed before the `DiscountDueDate`, the discount is automatically captured:

```
Dr. Accounts Payable (control)    [Invoice amount]
    Cr. Bank Account                    [Invoice amount - discount]
    Cr. Purchase Discounts Received     [Discount amount]
```

This is triggered automatically in `PaymentRun.BuildPaymentLine()` when the payment date precedes `DiscountDueDate`.

---

## 7. AP Aging & Payment Scheduling

The AP aging report shows what is due and when. It is the primary input to the payment run.

```sql
CREATE OR REPLACE VIEW v_ap_aging AS
SELECT
    s.tenant_id,
    s.supplier_id,
    sup.name                                                                AS supplier_name,
    SUM(s.functional_amount) FILTER (WHERE s.days_overdue <= 0)           AS current_amount,
    SUM(s.functional_amount) FILTER (WHERE s.days_overdue BETWEEN 1 AND 30)  AS days_30,
    SUM(s.functional_amount) FILTER (WHERE s.days_overdue BETWEEN 31 AND 60) AS days_60,
    SUM(s.functional_amount) FILTER (WHERE s.days_overdue BETWEEN 61 AND 90) AS days_90,
    SUM(s.functional_amount) FILTER (WHERE s.days_overdue > 90)           AS over_90,
    SUM(s.functional_amount)                                               AS total
FROM ap_subsidiary_entries s
JOIN suppliers sup ON sup.id = s.supplier_id
WHERE s.tenant_id  = current_setting('app.current_tenant')::UUID
  AND s.is_settled = FALSE
GROUP BY s.tenant_id, s.supplier_id, sup.name;
```

---

## 8. Payment Runs (Batch Supplier Payments)

### Why Payment Runs Exist

An organisation with 30 supplier invoices due this week should not need 30 separate bank transfers. A payment run collects all due invoices, groups payments by supplier (one payment per supplier regardless of invoice count), generates a bank file, and posts a single batch GL entry after the bank confirms processing.

### Payment Run Lifecycle

```
Draft ──► Approved ──► Exported (bank file generated) ──► Posted (GL confirmed) ──► terminal
  │            │
  └──[Cancel]──┘
```

| Status | Description | GL Impact |
|---|---|---|
| `draft` | Items selected, amounts calculated | None |
| `approved` | Authorised for payment | None |
| `exported` | Bank file downloaded for upload to internet banking | None |
| `posted` | Bank confirms payment; GL entries posted | DR AP, CR Bank for each supplier |
| `cancelled` | Cancelled before export | None |

### Payment Run Structure

```go
type PaymentRun struct {
    ID             uuid.UUID
    TenantID       uuid.UUID
    OrganisationID uuid.UUID
    DueOnOrBefore  time.Time        // include all invoices due on or before this date
    BankAccountID  uuid.UUID
    Currency       string
    Status         PaymentRunStatus
    TotalAmount    decimal.Decimal
    ItemCount      int
    Items          []PaymentRunItem
    ExportFormat   string           // equity_eft | kcb_rtgs | swift_mt101 | generic_csv
    ExportedAt     *time.Time
    ApprovedBy     *uuid.UUID
    PostedAt       *time.Time
    CreatedBy      uuid.UUID
}

type PaymentRunItem struct {
    InvoiceID      uuid.UUID
    SupplierID     uuid.UUID
    Amount         decimal.Decimal
    Currency       string
    BankAccount    SupplierBankAccount
    Reference      string           // remittance reference
    Included       bool             // finance manager can exclude individual items
    EarlyDiscount  *decimal.Decimal // if paying before discount_due_date
}
```

### Bank File Formats

| Format | Banks | Description |
|---|---|---|
| `equity_eft` | Equity Bank Kenya | Equity Bank EFT bulk payment format (.txt) |
| `kcb_rtgs` | KCB | KCB RTGS batch file format (.csv) |
| `swift_mt101` | International | SWIFT MT101 format for cross-border |
| `generic_csv` | Any | Configurable CSV with bank name, account, amount columns |

The format is defined by a `PaymentFileBuilder` interface, making it straightforward to add new bank formats:

```go
type PaymentFileBuilder interface {
    Format()  string
    Build(run *PaymentRun) ([]byte, string, error)  // returns file bytes, filename, error
}
```

### GL Posting on Confirmation

When the finance team confirms that the bank has processed the payment run, the system posts the GL entries in a single transaction:

```go
func (s *PaymentService) ConfirmAndPost(ctx context.Context, runID uuid.UUID, by uuid.UUID) error {
    run, err := s.runRepo.Get(ctx, runID)
    if err != nil {
        return err
    }
    if run.Status != PRExported {
        return ErrPaymentRunNotExported
    }

    // Build one journal entry with lines per supplier
    entry := buildPaymentJournalEntry(run)
    if err := s.journalSvc.PostIntegrationEntry(ctx, entry); err != nil {
        return err
    }

    run.Status   = PRPosted
    run.PostedAt = ptr(time.Now())
    return s.runRepo.Update(ctx, run)
}
```

The journal entry looks like:

```
Dr. AP – Supplier A         232,000
Dr. AP – Supplier B          58,000
Dr. AP – Supplier C          17,400
    Cr. Bank Account – Main         307,400
```

Each AP line's amount is applied to close the corresponding `ap_subsidiary_entries` as settled.

---

## 9. Petty Cash Imprest System

### How the Imprest System Works

The imprest system maintains a fixed float (e.g. 50,000 KES). Disbursements reduce the physical cash held. When the cash falls to the replenishment threshold, the custodian submits a replenishment request listing all expenditures since the last replenishment. The GL entry is created at replenishment time, not at disbursement time. The replenishment cheque restores the float to the fixed amount.

```go
type PettyCashFund struct {
    ID                 uuid.UUID
    TenantID           uuid.UUID
    OrganisationID     uuid.UUID
    Name               string
    AccountID          uuid.UUID      // the petty cash GL account (account_type = cash)
    FixedFloat         decimal.Decimal
    ReplenishThreshold decimal.Decimal
    CustodianID        uuid.UUID
    IsActive           bool
    LastReplenishedAt  *time.Time
    LastCountAt        *time.Time
}
```

### Disbursement

No GL entry is created at disbursement. The custodian records the disbursement in the petty cash log (stored in `petty_cash_vouchers`), attaches a receipt, and the log balance decreases accordingly. The GL account balance does not change until replenishment.

```go
type PettyCashVoucher struct {
    ID            uuid.UUID
    FundID        uuid.UUID
    VoucherDate   time.Time
    Payee         string
    Amount        decimal.Decimal
    ExpenseType   string
    Description   string
    AccountCode   string         // expense account to charge on replenishment
    CostCentreID  *uuid.UUID
    ReceiptRef    string
    AuthorisedBy  uuid.UUID
}
```

### Replenishment

At replenishment, all vouchers since the last replenishment are aggregated into a journal entry, grouped by expense account and cost centre:

```go
func (s *PettyCashService) CreateReplenishment(ctx context.Context, fundID uuid.UUID, by uuid.UUID) (*JournalEntry, error) {
    fund, err := s.fundRepo.Get(ctx, fundID)
    if err != nil {
        return nil, err
    }
    unreplenished, err := s.voucherRepo.GetUnreplenished(ctx, fundID)
    if err != nil {
        return nil, err
    }

    total := decimal.Zero
    lines := aggregateByAccountAndCostCentre(unreplenished)

    entry := &JournalEntry{
        Description: fmt.Sprintf("Petty cash replenishment — %s", fund.Name),
        Type:        TypeSystem,
    }
    for _, line := range lines {
        entry.Lines = append(entry.Lines, JournalLine{
            AccountID:    line.AccountID,
            DebitAmount:  line.Total,
            CostCentreID: line.CostCentreID,
        })
        total = total.Add(line.Total)
    }
    // Credit the bank account for the replenishment cheque
    entry.Lines = append(entry.Lines, JournalLine{
        AccountID:    fund.BankAccountID,
        CreditAmount: total,
        Description:  fmt.Sprintf("Replenishment cheque — %s", fund.Name),
    })

    return entry, nil
}
```

### Physical Count

A periodic physical count verifies the physical cash matches the log balance. If there is a shortage or overage, an adjusting entry is posted:

```
Shortage:
Dr. Petty Cash Shortage (expense)     [Amount]
    Cr. Petty Cash Account (asset)         [Amount]

Overage:
Dr. Petty Cash Account (asset)        [Amount]
    Cr. Petty Cash Overage (income)        [Amount]
```

---

## 10. Business Rules & Validation

### AR Rules

| Rule ID | Rule | Enforcement |
|---|---|---|
| `FIN-AR-001` | AR sub-ledger total must equal AR control account balance at period end | Period close checklist |
| `FIN-AR-002` | A direct GL posting to the AR control account (manual entry) is blocked unless account allows manual entries | `ValidateAccounts` pipeline stage |
| `FIN-AR-003` | Customer balance cannot go negative on settlement without explicit credit note or write-off | Application check in `applyAR` |
| `FIN-AR-004` | Bad debt write-off requires `finance.transactions.reverse` permission | Permission check |
| `FIN-AR-005` | Foreign currency AR entries must be revalued at period end | Period close checklist |

### AP Rules

| Rule ID | Rule | Enforcement |
|---|---|---|
| `FIN-AP-001` | AP sub-ledger total must equal AP control account balance at period end | Period close checklist |
| `FIN-AP-002` | A supplier invoice cannot be entered with a date more than 90 days in the past without Finance Manager override | Configurable check in event consumer |
| `FIN-AP-003` | Duplicate supplier invoice detection: same supplier + invoice number + amount within 30 days | `CheckDuplicates` stage |
| `FIN-AP-004` | AP entry on a supplier that is on payment hold is flagged for manual review | Integration check |

### Payment Run Rules

| Rule ID | Rule | Enforcement |
|---|---|---|
| `FIN-PAY-001` | Only one active payment run per bank account at a time | Application check |
| `FIN-PAY-002` | Payment run creator cannot be the approver | Permission + application check |
| `FIN-PAY-003` | Payment run cannot be exported without approval | Status check |
| `FIN-PAY-004` | Payment run cannot be cancelled after export | Status check |
| `FIN-PAY-005` | Total of payment run cannot exceed bank account available balance (if cash position is tracked) | Configurable check |

### Petty Cash Rules

| Rule ID | Rule | Enforcement |
|---|---|---|
| `FIN-PC-001` | Individual disbursement cannot exceed the fund's `max_single_disbursement` limit | Application check |
| `FIN-PC-002` | Replenishment must include all unreplenished vouchers — cannot selectively replenish | Application enforced |
| `FIN-PC-003` | Physical count shortages above `shortage_tolerance` require custodian explanation | Configurable check |
| `FIN-PC-004` | Only the fund's custodian can create disbursement vouchers | Permission check |

---

## 11. Performance & Storage

### AR/AP Sub-Ledger Volume

| Scenario | Open Entries at Any Time | Annual Throughput |
|---|---|---|
| Single petrol station | ~200 AR + ~50 AP | ~2,000 AR + ~600 AP |
| 5-site operator with fleet cards | ~2,000 AR + ~200 AP | ~24,000 AR + ~2,400 AP |
| 20-site operator | ~8,000 AR + ~800 AP | ~100,000 AR + ~10,000 AP |

### Query Performance Targets

| Query | Mechanism | Target P95 |
|---|---|---|
| Customer balance (single customer) | Index on `(tenant_id, customer_id, is_settled)` | < 5ms |
| AR aging (all customers, 2,000 open entries) | Materialised view + GROUP BY | < 500ms |
| AP invoices due this week (50 items) | Index on `(tenant_id, due_date)` + status filter | < 50ms |
| AR control reconciliation | Two index scans + subtraction | < 100ms |
| Payment run generation (30 invoices) | Index scan + aggregation | < 200ms |

### Storage Estimates

| Table | Row Size | Rows/Year (5-site) | Annual Storage |
|---|---|---|---|
| `ar_subsidiary_entries` | ~300 bytes | 24,000 | ~7 MB |
| `ap_subsidiary_entries` | ~300 bytes | 2,400 | ~0.7 MB |
| `petty_cash_vouchers` | ~250 bytes | ~1,200 (5 vouchers/week/site) | ~0.3 MB |
| `payment_runs` | ~400 bytes | ~52 (weekly) | ~21 KB |
| `payment_run_items` | ~200 bytes | ~2,400 (weekly runs, avg 50 items) | ~0.5 MB |

Total AR/AP/payment storage: approximately **8.5 MB per year** for a 5-site operator. Negligible compared to journal_lines.

### Archiving Strategy

`ar_subsidiary_entries` where `is_settled = TRUE` and `transaction_date < 2 years ago` can be moved to an archive table without affecting any active operation. Settled entries are only needed for historical reports and audit, both of which can query the archive table.

---

## 12. Database Schema

```sql
-- ── ar_subsidiary_entries ─────────────────────────────────────────────────────

CREATE TABLE ar_subsidiary_entries (
    tenant_id          UUID          NOT NULL REFERENCES tenants(id),
    id                 UUID          NOT NULL DEFAULT gen_random_uuid(),
    created_at         TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    organisation_id    UUID          NOT NULL,
    customer_id        UUID          NOT NULL,
    journal_line_id    UUID          NOT NULL REFERENCES journal_lines(id),
    invoice_ref        TEXT,
    transaction_date   DATE          NOT NULL,
    due_date           DATE          NOT NULL,
    currency           CHAR(3)       NOT NULL,
    original_amount    NUMERIC(18,4) NOT NULL,
    functional_amount  NUMERIC(18,4) NOT NULL,
    settled_amount     NUMERIC(18,4) NOT NULL DEFAULT 0,
    is_settled         BOOLEAN       NOT NULL DEFAULT FALSE,
    write_off          BOOLEAN       NOT NULL DEFAULT FALSE,
    days_overdue       INT GENERATED ALWAYS AS (GREATEST(0, CURRENT_DATE - due_date)) STORED,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, journal_line_id)
);

CREATE INDEX idx_ar_customer  ON ar_subsidiary_entries (tenant_id, customer_id, is_settled);
CREATE INDEX idx_ar_due_date  ON ar_subsidiary_entries (tenant_id, due_date) WHERE is_settled = FALSE;
CREATE INDEX idx_ar_overdue   ON ar_subsidiary_entries (tenant_id, days_overdue DESC) WHERE is_settled = FALSE;

ALTER TABLE ar_subsidiary_entries ENABLE ROW LEVEL SECURITY;
ALTER TABLE ar_subsidiary_entries FORCE  ROW LEVEL SECURITY;
CREATE POLICY ar_app ON ar_subsidiary_entries FOR ALL TO application_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

-- ── ap_subsidiary_entries ─────────────────────────────────────────────────────

CREATE TABLE ap_subsidiary_entries (
    tenant_id           UUID          NOT NULL REFERENCES tenants(id),
    id                  UUID          NOT NULL DEFAULT gen_random_uuid(),
    created_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    organisation_id     UUID          NOT NULL,
    supplier_id         UUID          NOT NULL,
    journal_line_id     UUID          NOT NULL REFERENCES journal_lines(id),
    invoice_ref         TEXT,
    transaction_date    DATE          NOT NULL,
    due_date            DATE          NOT NULL,
    currency            CHAR(3)       NOT NULL,
    original_amount     NUMERIC(18,4) NOT NULL,
    functional_amount   NUMERIC(18,4) NOT NULL,
    settled_amount      NUMERIC(18,4) NOT NULL DEFAULT 0,
    is_settled          BOOLEAN       NOT NULL DEFAULT FALSE,
    discount_due_date   DATE,
    discount_amount     NUMERIC(18,4),
    days_overdue        INT GENERATED ALWAYS AS (GREATEST(0, CURRENT_DATE - due_date)) STORED,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, journal_line_id)
);

CREATE INDEX idx_ap_supplier  ON ap_subsidiary_entries (tenant_id, supplier_id, is_settled);
CREATE INDEX idx_ap_due_date  ON ap_subsidiary_entries (tenant_id, due_date) WHERE is_settled = FALSE;

ALTER TABLE ap_subsidiary_entries ENABLE ROW LEVEL SECURITY;
ALTER TABLE ap_subsidiary_entries FORCE  ROW LEVEL SECURITY;
CREATE POLICY ap_app ON ap_subsidiary_entries FOR ALL TO application_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

-- ── payment_runs ──────────────────────────────────────────────────────────────

CREATE TABLE payment_runs (
    tenant_id       UUID          NOT NULL REFERENCES tenants(id),
    id              UUID          NOT NULL DEFAULT gen_random_uuid(),
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    organisation_id UUID          NOT NULL,
    due_on_before   DATE          NOT NULL,
    bank_account_id UUID          NOT NULL,
    currency        CHAR(3)       NOT NULL,
    status          TEXT          NOT NULL DEFAULT 'draft'
                        CHECK (status IN ('draft','approved','exported','posted','cancelled')),
    total_amount    NUMERIC(18,4) NOT NULL DEFAULT 0,
    item_count      INT           NOT NULL DEFAULT 0,
    export_format   TEXT,
    exported_at     TIMESTAMPTZ,
    approved_by     UUID,
    posted_at       TIMESTAMPTZ,
    created_by      UUID          NOT NULL,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id)
);

CREATE TABLE payment_run_items (
    tenant_id        UUID          NOT NULL REFERENCES tenants(id),
    id               UUID          NOT NULL DEFAULT gen_random_uuid(),
    payment_run_id   UUID          NOT NULL REFERENCES payment_runs(id),
    invoice_id       UUID          NOT NULL,
    supplier_id      UUID          NOT NULL,
    amount           NUMERIC(18,4) NOT NULL,
    currency         CHAR(3)       NOT NULL,
    bank_account_ref TEXT,
    reference        TEXT,
    included         BOOLEAN       NOT NULL DEFAULT TRUE,
    early_discount   NUMERIC(18,4),

    PRIMARY KEY (id),
    UNIQUE (tenant_id, payment_run_id, invoice_id)
);

-- ── petty_cash_funds & vouchers ───────────────────────────────────────────────

CREATE TABLE petty_cash_funds (
    tenant_id           UUID          NOT NULL REFERENCES tenants(id),
    id                  UUID          NOT NULL DEFAULT gen_random_uuid(),
    created_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    organisation_id     UUID          NOT NULL,
    name                TEXT          NOT NULL,
    account_id          UUID          NOT NULL REFERENCES accounts(id),
    bank_account_id     UUID,
    fixed_float         NUMERIC(18,4) NOT NULL,
    replenish_threshold NUMERIC(18,4) NOT NULL,
    custodian_id        UUID          NOT NULL,
    is_active           BOOLEAN       NOT NULL DEFAULT TRUE,
    last_replenished_at TIMESTAMPTZ,
    last_count_at       TIMESTAMPTZ,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id)
);

CREATE TABLE petty_cash_vouchers (
    tenant_id       UUID          NOT NULL REFERENCES tenants(id),
    id              UUID          NOT NULL DEFAULT gen_random_uuid(),
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    fund_id         UUID          NOT NULL REFERENCES petty_cash_funds(id),
    voucher_date    DATE          NOT NULL,
    payee           TEXT          NOT NULL,
    amount          NUMERIC(18,4) NOT NULL CHECK (amount > 0),
    description     TEXT,
    account_code    TEXT          NOT NULL,
    cost_centre_id  UUID,
    receipt_ref     TEXT,
    authorised_by   UUID          NOT NULL,
    is_replenished  BOOLEAN       NOT NULL DEFAULT FALSE,
    replenishment_journal_id UUID,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id)
);

CREATE INDEX idx_pcv_fund_unreplenished ON petty_cash_vouchers (tenant_id, fund_id)
    WHERE is_replenished = FALSE;
```

---

## 13. API Reference

### Accounts Receivable

```
GET    /finance/ar/customers                   AR balances per customer (?as_at=)
GET    /finance/ar/customers/{id}              Single customer AR balance + open invoices
GET    /finance/ar/aging                       AR aging report (?as_at=, ?buckets=30,60,90)
POST   /finance/ar/write-off                   Write off a specific invoice (requires approval)
GET    /finance/ar/reconciliation              AR control vs. subsidiary reconciliation status
```

### Accounts Payable

```
GET    /finance/ap/suppliers                   AP balances per supplier (?as_at=)
GET    /finance/ap/suppliers/{id}              Single supplier AP balance + open invoices
GET    /finance/ap/aging                       AP aging report (?as_at=, ?buckets=30,60,90)
GET    /finance/ap/due-this-week               Invoices due in next 7 days
GET    /finance/ap/reconciliation              AP control vs. subsidiary reconciliation status
```

### Payment Runs

```
GET    /finance/payment-runs                   List payment runs
POST   /finance/payment-runs                   Create from due AP invoices
PATCH  /finance/payment-runs/{id}/items        Update included/excluded items
POST   /finance/payment-runs/{id}/approve      Draft → Approved
POST   /finance/payment-runs/{id}/export       Generate bank payment file (returns binary download)
POST   /finance/payment-runs/{id}/confirm      Mark bank-processed; post GL entries
POST   /finance/payment-runs/{id}/cancel       Cancel (only before export)
```

**Create payment run:**
```json
{
  "due_on_or_before": "2025-02-07",
  "bank_account_id": "uuid",
  "currency": "KES",
  "invoice_ids": ["uuid1", "uuid2", "uuid3"]
}
```

**Export payment file:**
```json
{ "format": "equity_eft" }
```

### Petty Cash

```
GET    /finance/petty-cash/funds               List petty cash funds
GET    /finance/petty-cash/funds/{id}          Fund details + current balance + recent vouchers
POST   /finance/petty-cash/funds/{id}/disburse Create a disbursement voucher
POST   /finance/petty-cash/funds/{id}/replenish Replenish (creates journal entry + marks vouchers as replenished)
POST   /finance/petty-cash/funds/{id}/count    Record physical count (creates adjustment if needed)
```

---

## 14. Feature Flags & Configuration

| Flag | Type | Default | Description |
|---|---|---|---|
| `ar_sub_rec_tolerance` | decimal | `0.01` | Acceptable difference (KES) between AR control and subsidiary total at close |
| `ap_sub_rec_tolerance` | decimal | `0.01` | Same for AP |
| `ar_credit_limit_enabled` | bool | `false` | Enable credit limit tracking per customer. Requires customer master with credit limit field. |
| `ar_credit_limit_action` | enum | `warn` | `warn` or `block` when credit limit would be exceeded on new invoice |
| `ap_duplicate_check_days` | int | `30` | Lookback window for duplicate supplier invoice detection |
| `ap_backdated_invoice_days` | int | `90` | Maximum days in the past for a supplier invoice date without override |
| `payment_run_enabled` | bool | `true` | Enable batch payment run module |
| `payment_run_export_formats` | string[] | `["generic_csv"]` | Enabled bank file export formats |
| `payment_run_require_approval` | bool | `true` | Payment runs require approval before export |
| `payment_run_max_amount_without_cfo` | decimal | `500000` | Runs above this amount require CFO approval |
| `petty_cash_enabled` | bool | `true` | Enable petty cash imprest module |
| `petty_cash_max_single_disbursement` | decimal | `5000` | Maximum single voucher amount without custodian manager override |
| `petty_cash_shortage_tolerance` | decimal | `50` | Maximum count shortage before explanation is required |
| `early_payment_discount_enabled` | bool | `false` | Track and automatically capture early payment discounts |

---

## 15. v1.0 Rollout Assessment

### Must Have at v1.0

- AR sub-ledger updated on every Sales module event
- AP sub-ledger updated on every Procurement module event
- AR and AP aging queries (even without materialised views — direct queries are acceptable at low volume)
- AR/AP control reconciliation (needed for period close checklist)
- Payment runs with CSV export format (Equity EFT and KCB RTGS as high priority given target market)
- Petty cash fund and disbursement voucher (common in East African SME operations)

### Can Be Deferred

| Feature | Suggested Release |
|---|---|
| Bad debt provision workflow | v1.1 — direct journal entry covers the case at v1.0 |
| Credit limit enforcement | v1.1 — most v1.0 customers are cash/LPO based |
| Early payment discount capture | v1.1 |
| AR/AP aging materialised views on read replica | v1.1 — direct queries acceptable at v1.0 volumes |
| Customer statement PDF | v1.1 |
| SWIFT MT101 payment file | v1.1 |
| Petty cash physical count workflow | v1.1 — manual journal entry handles shortage at v1.0 |

### Never Defer

- Unique constraint on `(tenant_id, journal_line_id)` in AR/AP sub-ledger tables — prevents duplicate sub-ledger entries from the idempotent consumer
- Self-approval prohibition on payment runs
- `days_overdue` as a generated column — ensures aging queries are always consistent

---

*End of FILE-05. Proceeding to FILE-06: Cost Centre & Budget Management.*
