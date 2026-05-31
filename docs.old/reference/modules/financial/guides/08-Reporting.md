# Awo ERP — Finance Module
## FILE-08: Financial Reporting Engine & Report Catalogue

**Document Version:** 2.0.0  
**Series:** FILE-08 of 10  
**Depends On:** FILE-02 (COA & Balances), FILE-03 (GL), FILE-05 (AR/AP), FILE-06 (Cost Centres & Budgets), FILE-07 (Cash, IC, FX)  
**Depended On By:** FILE-09 (Authorization — report permissions), FILE-10 (API — report endpoints)

---

## Table of Contents

1. [Why a Dedicated Reporting Engine](#1-why-a-dedicated-reporting-engine)
2. [What You Lose Without It](#2-what-you-lose-without-it)
3. [How ERPNext, NetSuite and QuickBooks Handle Reporting](#3-how-erpnext-netsuite-and-quickbooks-handle-reporting)
4. [Architecture: Read Replica & View Layer](#4-architecture-read-replica--view-layer)
5. [Report Definition Model](#5-report-definition-model)
6. [Statutory Financial Statements](#6-statutory-financial-statements)
7. [Management & Operational Reports](#7-management--operational-reports)
8. [Subsidiary Ledger Reports](#8-subsidiary-ledger-reports)
9. [Tax & Compliance Reports](#9-tax--compliance-reports)
10. [Cash & Treasury Reports](#10-cash--treasury-reports)
11. [Report Execution Pipeline](#11-report-execution-pipeline)
12. [Async Report Execution](#12-async-report-execution)
13. [Report Scheduling & Delivery](#13-report-scheduling--delivery)
14. [Export Formats](#14-export-formats)
15. [Audit Package Export](#15-audit-package-export)
16. [Performance & Storage](#16-performance--storage)
17. [Database Schema](#17-database-schema)
18. [API Reference](#18-api-reference)
19. [Feature Flags & Configuration](#19-feature-flags--configuration)
20. [v1.0 Rollout Assessment](#20-v10-rollout-assessment)

---

## 1. Why a Dedicated Reporting Engine

Financial reports are read-heavy workloads. A balance sheet aggregates all account balances. An AR aging scans every open customer invoice. A GL detail report for a year may traverse hundreds of thousands of journal lines. Running these queries against the same PostgreSQL instance that processes journal entries risks starving the transactional workload — a slow balance sheet query should not delay a payroll posting.

The reporting engine solves this by routing all report queries to a read replica, exposing a catalogue of purpose-built SQL views, and providing a definition-driven framework where new reports can be created by configuring a `ReportDefinition` record rather than writing new application code. Reports that take longer than a few seconds are executed asynchronously and the result is stored for retrieval — so the user does not wait for the response.

Beyond infrastructure, the reporting engine provides: consistent output formatting, multi-format export (PDF, Excel, CSV, JSON), scheduled delivery by email or webhook, comparative periods, drill-down links from summary to transaction detail, and a permission model that controls which users see which reports.

---

## 2. What You Lose Without It

| Capability | Without Dedicated Reporting Engine |
|---|---|
| Real-time financial statements | Every report re-scans the transactional database — slow at scale |
| Report scheduling | No automated delivery — users must remember to run reports manually |
| Non-technical report customisation | Requires code changes for every new report variation |
| Consistent export formats | Each report needs its own PDF/Excel generator |
| Drill-down from summary to transaction | Not available without explicit implementation per report |
| Comparative periods in one view | Requires running the same query twice and merging in application code |
| Protection of transaction DB from report load | None — reporting queries compete with posting queries |

---

## 3. How ERPNext, NetSuite and QuickBooks Handle Reporting

### ERPNext

ERPNext generates reports using a Python-based report builder. Each report is a Python script that executes a query (or multiple queries) and returns columns and data. Reports are defined as DocTypes and can be customised by developers. There is no concept of a read replica — all report queries hit the primary MariaDB instance.

**Strengths:** Very flexible — developers can write any query. Large library of standard reports.  
**Weaknesses:** No async execution for slow reports. No scheduled delivery in the core product (available via third-party). No read replica separation. Report performance degrades linearly with data volume since there is no balance materialisation.

### NetSuite

NetSuite's SuiteAnalytics provides a report builder for non-technical users, a saved search framework for developers, and a data warehouse connector for BI tools. Reports run against NetSuite's internal reporting database. Saved searches support scheduling and email delivery.

**Strengths:** Excellent for non-technical report customisation. Real-time data. Multi-entity consolidated reports as standard.  
**Weaknesses:** Complex to build custom reports beyond the standard framework. Pricing is per-user for analytics features.

### QuickBooks

QuickBooks Online has a fixed set of reports with limited customisation (filter by date, class, or customer). Custom reports beyond those available are not possible without exporting to Excel.

**Strengths:** Clean, understandable report output. Easy to understand for non-accountants.  
**Weaknesses:** No custom reports. No scheduling. No drill-down beyond what QuickBooks provides.

**What Awo does:** Combines the ease-of-use of QuickBooks' output with the customisability of ERPNext's report framework and the performance of NetSuite's read-replica separation.

---

## 4. Architecture: Read Replica & View Layer

### Read Replica Routing

All report queries are routed to a PostgreSQL read replica. The application maintains two database connection pools:

```go
type DBPools struct {
    Primary  *pgxpool.Pool  // journal posting, period close, all writes
    Replica  *pgxpool.Pool  // all report queries, AR/AP aging, cash position
}
```

The `ReportingService` always uses `Replica`. The `JournalService` always uses `Primary`. There is no shared pool.

**Replication lag:** PostgreSQL streaming replication typically achieves < 100ms lag on the same availability zone. For financial reporting, this is acceptable — a report run immediately after posting may miss the last second of entries, but this is not a meaningful limitation in practice.

### View Layer

All reports are backed by purpose-built PostgreSQL views on the read replica. Views carry `security_invoker = true` — they execute with the permissions of the calling role, not the view definer. This means RLS policies are applied consistently whether a user queries a view directly or through the reporting API.

```sql
-- All reporting views follow this pattern:
CREATE OR REPLACE VIEW v_trial_balance AS
SELECT
    ab.tenant_id,
    ab.period_id,
    a.code           AS account_code,
    a.name           AS account_name,
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
```

### View Catalogue

| View Name | Purpose | Primary Key Inputs |
|---|---|---|
| `v_trial_balance` | All accounts with debit/credit/net for a period | `tenant_id`, `period_id` |
| `v_balance_sheet` | Hierarchical balance sheet with section grouping | `tenant_id`, `period_id` |
| `v_profit_loss` | P&L with gross profit, EBIT, net income | `tenant_id`, `period_id(s)` |
| `v_cash_flow_indirect` | Cash flow statement (indirect method) | `tenant_id`, `from_period_id`, `to_period_id` |
| `v_gl_detail` | Transaction-level ledger for one account over a date range | `tenant_id`, `account_id`, `from`, `to` |
| `v_ar_aging` | AR aging by customer with configurable day buckets | `tenant_id`, `as_at` |
| `v_ap_aging` | AP aging by supplier | `tenant_id`, `as_at` |
| `v_budget_vs_actual` | Budget lines joined to actual account balances | `tenant_id`, `budget_id`, `period_id` |
| `v_cost_centre_pl` | P&L scoped to a cost centre subtree | `tenant_id`, `cost_centre_path`, `period_id(s)` |
| `v_bank_rec_summary` | Bank reconciliation status per account per period | `tenant_id`, `period_id` |
| `v_cash_position` | Real-time cash across all bank/cash accounts | `tenant_id` |
| `v_cash_forecast` | 13-week rolling forecast | `tenant_id`, `as_at` |
| `v_ic_balances` | Intercompany balances per entity pair | `tenant_id`, `as_at` |
| `v_period_close_status` | Checklist status for a given period | `tenant_id`, `period_id` |
| `v_vat_return` | VAT output and input summary for a date range | `tenant_id`, `from`, `to` |
| `v_wht_register` | Withholding tax register | `tenant_id`, `from`, `to` |

---

## 5. Report Definition Model

The `ReportDefinition` allows new report variants to be configured without writing Go code. It stores the data source view, column definitions, filter parameters, default grouping, and available output formats.

```go
type ReportDefinition struct {
    ID            uuid.UUID
    TenantID      uuid.UUID    // NULL = system report (available to all tenants)
    Code          string       // "balance_sheet", "profit_loss", "ar_aging"
    Name          string
    IsSystem      bool         // system reports cannot be deleted or structurally modified
    DataSource    string       // PostgreSQL view name on the read replica
    Parameters    []ReportParam
    Columns       []ReportColumn
    Grouping      []string
    Sorting       []SortSpec
    DrillDown     *DrillDownConfig  // nil = no drill-down
    OutputFormats []string          // "pdf" | "excel" | "csv" | "json"
}

type ReportParam struct {
    Name       string
    Label      string
    Type       string       // "date" | "period" | "account" | "cost_centre" | "bool"
    Required   bool
    Default    interface{}
}

type ReportColumn struct {
    Field      string
    Label      string
    Type       string       // "text" | "decimal" | "date" | "percent"
    Align      string       // "left" | "right" | "center"
    Subtotal   bool         // whether to show subtotals for this column
    Highlight  string       // "negative_red" | "variance_color" | ""
}

type DrillDownConfig struct {
    FromField    string    // column to click for drill-down
    ToReport     string    // report code to drill into
    ParamMapping map[string]string  // how to pass values from parent to child report
}
```

**Drill-down example:** Balance sheet → account balance → GL detail for that account and period. Clicking the "Accounts Receivable" line on the balance sheet passes `account_id` and `period_id` to `v_gl_detail`.

---

## 6. Statutory Financial Statements

These are the reports required for regulatory compliance, tax filing, and external stakeholder reporting.

### Balance Sheet

```sql
CREATE OR REPLACE VIEW v_balance_sheet AS
WITH account_tree AS (
    SELECT
        a.tenant_id,
        a.id             AS account_id,
        a.code,
        a.name,
        a.root_type,
        a.account_type,
        a.path,
        a.level,
        a.is_group,
        CASE a.root_type
            WHEN 'asset'     THEN 'Assets'
            WHEN 'liability' THEN 'Liabilities'
            WHEN 'equity'    THEN 'Equity'
        END AS bs_section,
        ab.period_id,
        ab.net_balance
    FROM accounts a
    LEFT JOIN account_balances ab ON ab.account_id = a.id
    WHERE a.report_section = 'balance_sheet'
)
SELECT
    t.tenant_id,
    t.period_id,
    t.bs_section,
    t.path,
    t.code,
    t.name,
    t.level,
    t.is_group,
    COALESCE(t.net_balance, 0)  AS balance
FROM account_tree t;

ALTER VIEW v_balance_sheet SET (security_invoker = true);
```

The application layer walks the COA tree, aggregates leaf balances up to group accounts, and formats the hierarchical output. Group account balances are always computed from their children — they are never stored directly.

**Balance sheet invariant check:** After assembly, the reporting service verifies:

```go
func (r *ReportService) verifyBalanceSheet(bs *BalanceSheet) error {
    if !bs.Assets.Total.Equal(bs.Liabilities.Total.Add(bs.Equity.Total)) {
        return ErrBalanceSheetDoesNotBalance{
            Assets:      bs.Assets.Total,
            LiabEquity:  bs.Liabilities.Total.Add(bs.Equity.Total),
            Difference:  bs.Assets.Total.Sub(bs.Liabilities.Total.Add(bs.Equity.Total)),
        }
    }
    return nil
}
```

If the balance sheet does not balance, the report returns a `balance_sheet_unbalanced` error rather than presenting misleading figures. This is a critical guardrail — an unbalanced balance sheet indicates a data integrity issue that must be investigated before the report is shared.

### Income Statement (Profit & Loss)

The P&L aggregates all revenue and expense account movements for a date range. It supports:
- Current period vs. prior period comparison
- Current period vs. same period last year
- Current period vs. budget (if budget module is active)
- Year-to-date columns alongside the current period

```go
type ProfitLoss struct {
    From, To       time.Time
    Currency       string
    Revenue        StatementSection
    CostOfSales    StatementSection
    GrossProfit    decimal.Decimal
    GrossMarginPct decimal.Decimal
    OpExpenses     StatementSection
    EBIT           decimal.Decimal
    EBITMarginPct  decimal.Decimal
    OtherItems     StatementSection
    PBT            decimal.Decimal
    TaxExpense     decimal.Decimal
    NetIncome      decimal.Decimal
    NetMarginPct   decimal.Decimal
    // Comparison columns (nil if not requested)
    Comparative    *ProfitLoss
    Budget         *ProfitLoss
}
```

### Cash Flow Statement (Indirect Method)

```go
type CashFlowStatement struct {
    From, To              time.Time
    Currency              string
    OperatingActivities   CashFlowSection
    InvestingActivities   CashFlowSection
    FinancingActivities   CashFlowSection
    NetCashMovement       decimal.Decimal
    OpeningCashBalance    decimal.Decimal
    ClosingCashBalance    decimal.Decimal
    // Verification: ClosingCash must equal OpeningCash + NetCashMovement
}
```

The indirect method starts from net income and adjusts for:
- Non-cash items (depreciation, amortisation)
- Changes in working capital (AR, inventory, AP, prepayments, accruals)
- Investing activities (asset purchases, disposals)
- Financing activities (loan drawdowns, repayments, equity movements)

### Trial Balance

The trial balance is the foundational verification report. Every other financial statement is derived from it. It lists every account with its opening balance, period movements, and closing balance.

```sql
-- Trial balance with period filter
SELECT *
FROM v_trial_balance
WHERE tenant_id = $1
  AND period_id = $2
  AND (NOT $show_zero OR net_balance <> 0)
ORDER BY path;
```

The trial balance shows debits and credits separately — not just a net balance — because this is the format required to verify double-entry integrity (total debits = total credits).

---

## 7. Management & Operational Reports

### Budget vs. Actual Variance

The budget variance report is the most important management report after the P&L. It shows every expense account alongside its budget, actual, and variance — by period, by cost centre, and by budget version.

```sql
CREATE OR REPLACE VIEW v_budget_vs_actual AS
SELECT
    b.tenant_id,
    b.id          AS budget_id,
    b.version_label,
    bl.account_id,
    bl.cost_centre_id,
    bl.period_id,
    bl.amount     AS budgeted,
    COALESCE(ab.period_debit - ab.period_credit, 0) AS actual,
    bl.amount - COALESCE(ab.period_debit - ab.period_credit, 0) AS variance,
    CASE
        WHEN bl.amount = 0 THEN NULL
        ELSE ROUND(
            (bl.amount - COALESCE(ab.period_debit - ab.period_credit, 0))
            / bl.amount * 100, 2
        )
    END AS variance_pct
FROM budget_lines bl
JOIN budgets b ON b.id = bl.budget_id
LEFT JOIN account_balances ab ON
    ab.account_id = bl.account_id
    AND ab.period_id = bl.period_id
    AND ab.tenant_id = b.tenant_id;

ALTER VIEW v_budget_vs_actual SET (security_invoker = true);
```

### Cost Centre P&L

```sql
CREATE OR REPLACE VIEW v_cost_centre_pl AS
SELECT
    jl.tenant_id,
    jl.cost_centre_id,
    cc.name      AS cost_centre_name,
    cc.path      AS cost_centre_path,
    a.code       AS account_code,
    a.name       AS account_name,
    a.root_type,
    je.period_id,
    SUM(jl.debit_amount)  AS total_debit,
    SUM(jl.credit_amount) AS total_credit,
    SUM(jl.debit_amount - jl.credit_amount) AS net
FROM journal_lines jl
JOIN journal_entries je ON je.id = jl.entry_id AND je.status = 'posted'
JOIN accounts a          ON a.id  = jl.account_id
JOIN cost_centres cc     ON cc.id = jl.cost_centre_id
WHERE jl.cost_centre_id IS NOT NULL
  AND a.root_type IN ('revenue', 'expense')
GROUP BY
    jl.tenant_id, jl.cost_centre_id, cc.name, cc.path,
    a.code, a.name, a.root_type, je.period_id;

ALTER VIEW v_cost_centre_pl SET (security_invoker = true);
```

This view powers cost centre P&L, multi-site comparison, and the budget owner's scoped view.

### Period Close Status

This view is the real-time dashboard for the period close checklist. Finance managers use it to track which items are blocking the close.

```sql
CREATE OR REPLACE VIEW v_period_close_status AS
SELECT
    p.tenant_id,
    p.id                  AS period_id,
    p.name                AS period_name,
    p.status,
    (SELECT COUNT(*)
     FROM journal_entries je
     WHERE je.tenant_id = p.tenant_id
       AND je.period_id = p.id
       AND je.status = 'pending_approval')     AS pending_approvals,
    (SELECT COUNT(*)
     FROM bank_reconciliations br
     WHERE br.tenant_id = p.tenant_id
       AND br.period_id = p.id
       AND br.status NOT IN ('approved','locked')) AS unapproved_bank_recs,
    (SELECT COUNT(*)
     FROM integration_event_failures ief
     WHERE ief.tenant_id = p.tenant_id
       AND ief.period_id = p.id
       AND ief.resolved = FALSE)               AS failed_integration_events
FROM accounting_periods p;

ALTER VIEW v_period_close_status SET (security_invoker = true);
```

---

## 8. Subsidiary Ledger Reports

### AR Aging Report

Described in FILE-05. The `v_ar_aging` view supports configurable bucket boundaries by computing `days_overdue` as a generated column on `ar_subsidiary_entries`. The bucket filter is applied at query time:

```sql
-- Configurable buckets (default: 30, 60, 90 days)
SELECT
    customer_id,
    customer_name,
    SUM(functional_amount) FILTER (WHERE days_overdue <= 0)    AS current_amount,
    SUM(functional_amount) FILTER (WHERE days_overdue BETWEEN 1 AND $b1)  AS bucket_1,
    SUM(functional_amount) FILTER (WHERE days_overdue BETWEEN $b1+1 AND $b2) AS bucket_2,
    SUM(functional_amount) FILTER (WHERE days_overdue BETWEEN $b2+1 AND $b3) AS bucket_3,
    SUM(functional_amount) FILTER (WHERE days_overdue > $b3)   AS over_bucket_3,
    SUM(functional_amount)                                     AS total
FROM ar_subsidiary_entries
WHERE tenant_id = $tenant_id
  AND is_settled = FALSE
GROUP BY customer_id, customer_name;
```

### GL Detail (Account Ledger)

The GL detail report shows every posted journal line for a specific account over a date range, with a running balance computed as a window function.

```sql
SELECT
    je.transaction_date,
    je.reference,
    je.description,
    jl.description        AS line_description,
    jl.debit_amount,
    jl.credit_amount,
    SUM(jl.debit_amount - jl.credit_amount)
        OVER (
            ORDER BY je.transaction_date, je.id
            ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
        ) + (ab.opening_debit - ab.opening_credit) AS running_balance,
    cc.name               AS cost_centre,
    jl.reference          AS line_reference
FROM journal_lines jl
JOIN journal_entries je      ON je.id = jl.entry_id
LEFT JOIN cost_centres cc   ON cc.id = jl.cost_centre_id
JOIN account_balances ab    ON ab.account_id = jl.account_id
    AND ab.period_id = (
        SELECT id FROM accounting_periods
        WHERE tenant_id       = $tenant_id
          AND organisation_id = $org_id
          AND start_date = (
              SELECT MIN(start_date)
              FROM   accounting_periods
              WHERE  tenant_id       = $tenant_id
                AND  organisation_id = $org_id
                AND  start_date >= $from
          )
    )
WHERE jl.account_id        = $account_id
  AND je.transaction_date  BETWEEN $from AND $to
  AND je.status            = 'posted'
  AND jl.tenant_id         = $tenant_id
ORDER BY je.transaction_date, je.id;
```

---

## 9. Tax & Compliance Reports

### VAT Return Summary

```sql
CREATE OR REPLACE VIEW v_vat_return AS
SELECT
    jl.tenant_id,
    je.transaction_date,
    CASE
        WHEN a.account_type = 'tax_payable' AND jl.credit_amount > 0 THEN 'output'
        WHEN a.account_type = 'tax_receivable' AND jl.debit_amount > 0 THEN 'input'
    END AS vat_type,
    SUM(jl.credit_amount - jl.debit_amount) AS vat_amount,
    -- The underlying taxable amount from the same journal entry
    SUM(taxable.net_amount) AS taxable_amount
FROM journal_lines jl
JOIN journal_entries je ON je.id = jl.entry_id AND je.status = 'posted'
JOIN accounts a          ON a.id  = jl.account_id
    AND a.account_type IN ('tax_payable', 'tax_receivable')
-- Self-join to get the taxable base from the same entry
LEFT JOIN LATERAL (
    SELECT SUM(jl2.debit_amount - jl2.credit_amount) AS net_amount
    FROM journal_lines jl2
    JOIN accounts a2 ON a2.id = jl2.account_id
    WHERE jl2.entry_id = jl.entry_id
      AND a2.root_type IN ('revenue', 'expense')
) taxable ON TRUE
GROUP BY jl.tenant_id, je.transaction_date, vat_type;

ALTER VIEW v_vat_return SET (security_invoker = true);
```

### Withholding Tax Register

The WHT register lists all payments subject to withholding tax — supplier payments, professional fees, rent — showing the gross amount, WHT rate, WHT deducted, and net payment. This is used to prepare monthly WHT remittance to KRA.

---

## 10. Cash & Treasury Reports

### Bank Reconciliation Summary

```sql
CREATE OR REPLACE VIEW v_bank_rec_summary AS
SELECT
    br.tenant_id,
    br.period_id,
    ba.account_code,
    ba.account_name,
    br.statement_closing_bal,
    br.gl_closing_balance,
    br.adjusted_bank_balance,
    br.adjusted_gl_balance,
    br.difference,
    br.status,
    br.prepared_by,
    br.approved_by,
    br.locked_at,
    (SELECT COUNT(*) FROM bank_statement_lines bsl
     WHERE bsl.reconciliation_id = br.id
       AND bsl.match_status = 'unmatched') AS unmatched_items
FROM bank_reconciliations br
JOIN bank_accounts ba ON ba.id = br.bank_account_id;

ALTER VIEW v_bank_rec_summary SET (security_invoker = true);
```

---

## 11. Report Execution Pipeline

When a report is requested, the execution pipeline:

1. **Validate parameters** — all required params present, date ranges valid, user has `finance.reports.run` permission
2. **Resolve scope** — apply tenant, organisation, and (if applicable) cost centre scope from the user's permissions
3. **Build query** — substitute parameters into the view query; apply additional filters
4. **Route to replica** — execute against the read replica connection pool
5. **Apply timeout** — hard cap at 120 seconds; returns `timeout` status if exceeded
6. **Format output** — apply column formatting, compute subtotals, build hierarchy for tree reports
7. **Verify invariants** — balance sheet balance check, trial balance debit/credit equality
8. **Generate download URLs** — render PDF and Excel versions asynchronously; return JSON inline
9. **Cache result** — store result with TTL for repeated access (configurable per report type)

```go
func (s *ReportService) Execute(ctx context.Context, req ReportRequest) (*ReportResult, error) {
    def, err := s.defRepo.GetByCode(ctx, req.TenantID, req.ReportCode)
    if err != nil {
        return nil, err
    }

    if err := s.validateParams(def, req.Params); err != nil {
        return nil, err
    }

    if cached := s.cache.Get(req.CacheKey()); cached != nil {
        return cached, nil
    }

    ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
    defer cancel()

    rows, err := s.replica.Query(ctx, buildQuery(def, req))
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            return nil, ErrReportTimeout{ReportCode: req.ReportCode}
        }
        return nil, err
    }

    result := buildResult(def, rows)

    if err := s.verifyInvariants(def.Code, result); err != nil {
        return nil, err
    }

    s.cache.Set(req.CacheKey(), result, def.CacheTTL)

    // Trigger async PDF/Excel generation
    s.asyncRenderer.Enqueue(result)

    return result, nil
}
```

---

## 12. Async Report Execution

Reports that are expected to take more than 5 seconds (configurable via `async_threshold_ms`) are automatically promoted to async execution. The API returns an `execution_id` immediately; the client polls for completion.

```
POST /finance/reports/gl-detail/run
{ "account_code": "1210", "from": "2024-01-01", "to": "2024-12-31" }

→ 202 Accepted
{
  "execution_id": "uuid",
  "status": "queued",
  "estimated_duration_seconds": 45
}

GET /finance/report-executions/uuid
→ 200 OK
{
  "status": "completed",
  "completed_at": "2025-02-06T09:17:32Z",
  "row_count": 12450,
  "download_urls": {
    "json":  "https://storage.../report.json",
    "excel": "https://storage.../report.xlsx",
    "csv":   "https://storage.../report.csv",
    "pdf":   "https://storage.../report.pdf"
  },
  "expires_at": "2025-02-07T09:17:32Z"
}
```

Report results are stored in object storage (S3-compatible) with a 24-hour TTL. After expiry, the report must be re-run.

### Execution States

```
queued → running → completed
                ↘ failed
                ↘ timeout
```

---

## 13. Report Scheduling & Delivery

Report schedules are configured per report per organisation. A schedule defines: which report to run, which parameters to use (with dynamic date resolution e.g. `"prior_month"`), at what time, and where to deliver the output.

```go
type ReportSchedule struct {
    ID            uuid.UUID
    TenantID      uuid.UUID
    ReportCode    string
    Name          string
    CronExpr      string          // standard cron: "0 8 10 * *" = 8am on the 10th of every month
    Params        map[string]interface{} // supports "prior_month", "current_month", "ytd"
    OutputFormat  string
    Delivery      ScheduleDelivery
    IsActive      bool
    LastRunAt     *time.Time
    NextRunAt     time.Time
}

type ScheduleDelivery struct {
    Type       string   // "email" | "webhook" | "storage"
    Recipients []string // email addresses or webhook URL
    Subject    string   // email subject template
}
```

**Dynamic date parameters:**

| Token | Resolves To |
|---|---|
| `current_period` | The currently open accounting period |
| `prior_period` | The most recently closed period |
| `prior_month_same_year` | Same period from 12 months ago |
| `ytd` | From fiscal year start to today |
| `current_fiscal_year` | Full current fiscal year |

---

## 14. Export Formats

### PDF

Financial statement PDFs include:
- Company letterhead (from tenant configuration)
- Report title, period, and parameters
- Page numbers and generation timestamp
- "Prepared by" and (for approved reports) "Approved by" signature blocks
- Hierarchical indentation for account tree reports

PDF generation uses a Go HTML→PDF pipeline with a fixed template per report type.

### Excel (.xlsx)

Excel exports include:
- Separate sheet per report section (e.g. Balance Sheet, Income Statement, Cash Flow in one workbook)
- Formatted numbers with thousands separators and 2 decimal places
- Subtotal rows with bold formatting
- Frozen header row
- Data validation to prevent inadvertent editing
- A metadata sheet with report parameters and generation details

### CSV

CSV exports are unformatted flat files, one row per data line (no hierarchy). Suitable for import into other systems or BI tools. Character encoding is UTF-8 with BOM for Excel compatibility.

### JSON

The JSON format returns the full report result including hierarchy, metadata, and computed subtotals. This is used for downstream processing (BI tools, consolidation systems) and for the in-app report rendering.

---

## 15. Audit Package Export

The audit package is a ZIP archive of all financial data for a fiscal year in a format suitable for external auditors. It is generated on request (typically annually) and takes several minutes to produce.

```
audit-package-FY2025/
├── 01-trial-balance.xlsx
├── 02-balance-sheet.xlsx
├── 03-income-statement.xlsx
├── 04-cash-flow.xlsx
├── 05-gl-detail-all-accounts.xlsx       (one sheet per account)
├── 06-ar-aging.xlsx
├── 07-ap-aging.xlsx
├── 08-journal-entries-all.xlsx          (all entries with approval chain)
├── 09-bank-reconciliations/
│   ├── jan-2025-account-1112.pdf
│   ├── feb-2025-account-1112.pdf
│   └── ...
├── 10-user-access-log.xlsx
├── 11-configuration-changes.xlsx
└── README.txt                           (guide for auditors)
```

The package is stored in object storage and a secure download link is generated with a 7-day expiry. Access to the audit package export is restricted to the `cfo` role.

---

## 16. Performance & Storage

### Report Query Performance Targets

| Report | Data Volume | Read Replica P95 |
|---|---|---|
| Trial balance (200 accounts) | `account_balances` — 2,400 rows | < 50ms |
| Balance sheet | `account_balances` filtered + tree walk | < 100ms |
| P&L (monthly) | `account_balances` + hierarchy walk | < 100ms |
| AR aging (2,000 open entries) | `ar_subsidiary_entries` GROUP BY | < 300ms |
| GL detail (1 year, high-volume account) | Up to 12,000 `journal_lines` rows | < 500ms |
| Budget vs. actual | `budget_lines` JOIN `account_balances` | < 200ms |
| Cost centre P&L (1 centre, 1 month) | `journal_lines` filtered by `cost_centre_id` | < 300ms |
| Period close checklist | 3 subquery counts | < 200ms |
| Cash flow statement | `journal_lines` + classification logic | < 1s |

Reports exceeding 5 seconds are promoted to async execution and the user receives an `execution_id`.

### Caching Strategy

| Report | TTL | Invalidation Trigger |
|---|---|---|
| Trial balance (closed period) | Permanent | None — closed period is immutable |
| Trial balance (open period) | 5 minutes | Any journal posting in that period |
| Balance sheet (closed period) | Permanent | None |
| Balance sheet (open period) | 5 minutes | Any journal posting |
| AR aging | 10 minutes | Any AR subsidiary entry update |
| Cash position | No cache | Always live (< 10ms query) |
| Cash forecast | 30 minutes | Configurable |
| Budget vs. actual | 5 minutes | Any journal posting or budget line change |

Cache keys always include `tenant_id` and `organisation_id` to ensure strict tenant isolation.

### Report Result Storage

Async report results are stored in object storage (S3-compatible), not in PostgreSQL. Storage estimate:

| Report | Typical Size | TTL |
|---|---|---|
| Trial balance (Excel) | ~200 KB | 24 hours |
| Full GL detail (1 year, Excel) | ~5–15 MB | 24 hours |
| Audit package (ZIP) | ~50–200 MB | 7 days |

For a 5-site operator generating one audit package per year plus weekly reports, total object storage usage is under **5 GB per year**.

### Read Replica Sizing

The read replica runs the same PostgreSQL version as the primary but with a read-optimised configuration:
- `shared_buffers`: 40% of RAM (primary uses 25% — replica can afford more since it does no writes)
- `work_mem`: 64 MB per connection (primary uses 4 MB — report queries may need large sort buffers)
- `max_parallel_workers_per_gather`: 4 (primary uses 0 — reports benefit from parallel query)
- `effective_cache_size`: 75% of RAM

For a 5-site operator: a 4 vCPU / 16 GB RAM instance is sufficient for all reporting workloads.

---

## 17. Database Schema

```sql
-- ── report_definitions ────────────────────────────────────────────────────────
-- System report definitions are tenant_id = NULL and available to all tenants.
-- Tenant-specific custom reports have a specific tenant_id.

CREATE TABLE report_definitions (
    tenant_id    UUID,   -- NULL = system report
    id           UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    code         TEXT        NOT NULL,
    name         TEXT        NOT NULL,
    is_system    BOOLEAN     NOT NULL DEFAULT FALSE,
    data_source  TEXT        NOT NULL,  -- view name on read replica
    parameters   JSONB       NOT NULL DEFAULT '[]',
    columns      JSONB       NOT NULL DEFAULT '[]',
    grouping     TEXT[]      NOT NULL DEFAULT '{}',
    sorting      JSONB       NOT NULL DEFAULT '[]',
    drill_down   JSONB,
    output_formats TEXT[]    NOT NULL DEFAULT '{pdf,excel,csv,json}',
    cache_ttl_seconds INT    NOT NULL DEFAULT 300,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, code)
);

-- ── report_executions ─────────────────────────────────────────────────────────

CREATE TABLE report_executions (
    tenant_id    UUID        NOT NULL REFERENCES tenants(id),
    id           UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    report_code  TEXT        NOT NULL,
    params       JSONB       NOT NULL,
    status       TEXT        NOT NULL DEFAULT 'queued'
                     CHECK (status IN ('queued','running','completed','failed','timeout')),
    row_count    INT,
    error_message TEXT,
    storage_path TEXT,       -- path to JSON result in object storage
    pdf_path     TEXT,
    excel_path   TEXT,
    csv_path     TEXT,
    completed_at TIMESTAMPTZ,
    expires_at   TIMESTAMPTZ,
    requested_by UUID        NOT NULL,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id)
);

CREATE INDEX idx_re_tenant_status ON report_executions (tenant_id, status, created_at DESC);
CREATE INDEX idx_re_expires       ON report_executions (expires_at)
    WHERE status = 'completed';

ALTER TABLE report_executions ENABLE ROW LEVEL SECURITY;
ALTER TABLE report_executions FORCE  ROW LEVEL SECURITY;
CREATE POLICY re_app ON report_executions FOR ALL TO application_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

-- ── report_schedules ──────────────────────────────────────────────────────────

CREATE TABLE report_schedules (
    tenant_id       UUID        NOT NULL REFERENCES tenants(id),
    id              UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    report_code     TEXT        NOT NULL,
    name            TEXT        NOT NULL,
    cron_expr       TEXT        NOT NULL,
    params          JSONB       NOT NULL DEFAULT '{}',
    output_format   TEXT        NOT NULL DEFAULT 'excel',
    delivery        JSONB       NOT NULL,
    is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
    last_run_at     TIMESTAMPTZ,
    next_run_at     TIMESTAMPTZ NOT NULL,
    created_by      UUID        NOT NULL,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id)
);

ALTER TABLE report_schedules ENABLE ROW LEVEL SECURITY;
ALTER TABLE report_schedules FORCE  ROW LEVEL SECURITY;
CREATE POLICY rs_app ON report_schedules FOR ALL TO application_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);
```

---

## 18. API Reference

### Report Execution

```
POST   /finance/reports/{code}/run              Run a report (sync or async)
GET    /finance/report-executions/{id}          Poll async execution status
GET    /finance/report-executions/{id}/download Download result in specified format (?format=pdf|excel|csv|json)
```

**Standard report request body:**
```json
{
  "organisation_id": "uuid",
  "params": {
    "as_at": "2025-01-31",
    "comparative_as_at": "2024-12-31",
    "show_zero_balances": false
  },
  "async": false,
  "output_format": "json"
}
```

**Standard system reports:**
```
POST /finance/reports/trial-balance/run
POST /finance/reports/balance-sheet/run
POST /finance/reports/profit-loss/run
POST /finance/reports/cash-flow/run
POST /finance/reports/ar-aging/run
POST /finance/reports/ap-aging/run
POST /finance/reports/gl-detail/run
POST /finance/reports/budget-variance/run
POST /finance/reports/cost-centre-pl/run
POST /finance/reports/cash-position/run
POST /finance/reports/cash-forecast/run
POST /finance/reports/ic-balances/run
POST /finance/reports/vat-return/run
POST /finance/reports/wht-register/run
GET  /finance/reports/period-close-status        (always sync — lightweight checklist view)
```

### Report Definitions & Schedules

```
GET    /finance/report-definitions              List available reports (system + custom)
POST   /finance/report-definitions              Create custom report definition
PATCH  /finance/report-definitions/{id}         Update custom report (system reports are read-only)
GET    /finance/report-schedules                List schedules
POST   /finance/report-schedules                Create schedule
PATCH  /finance/report-schedules/{id}           Update schedule
DELETE /finance/report-schedules/{id}           Delete schedule
POST   /finance/report-schedules/{id}/run-now   Trigger immediate execution of a scheduled report
```

### Audit Package

```
POST   /finance/audit-package/request           Request audit package generation for a fiscal year
GET    /finance/audit-package/status/{id}       Check generation status
GET    /finance/audit-package/download/{id}     Download completed package (ZIP, 7-day link)
```

---

## 19. Feature Flags & Configuration

| Flag | Type | Default | Description |
|---|---|---|---|
| `report_async_threshold_ms` | int | `5000` | Reports expected to take longer than this are automatically run async |
| `report_cache_enabled` | bool | `true` | Enable report result caching (Redis or in-process) |
| `report_default_cache_ttl` | int | `300` | Default cache TTL in seconds for open-period reports |
| `report_storage_ttl_hours` | int | `24` | Hours before async report result files expire in object storage |
| `report_audit_package_ttl_days` | int | `7` | Days before audit package download link expires |
| `report_max_rows_sync` | int | `50000` | Maximum row count for a synchronous report. Reports expected to exceed this are auto-promoted to async. |
| `report_pdf_enabled` | bool | `true` | Enable PDF report generation |
| `report_excel_enabled` | bool | `true` | Enable Excel (.xlsx) report generation |
| `report_scheduling_enabled` | bool | `false` | Enable report scheduling and automated delivery |
| `report_email_delivery_enabled` | bool | `false` | Enable email delivery for scheduled reports. Requires SMTP configuration. |
| `report_webhook_delivery_enabled` | bool | `false` | Enable webhook delivery for scheduled reports |
| `report_replica_lag_tolerance_ms` | int | `500` | If read replica lag exceeds this, report execution is paused and retried with backoff |
| `balance_sheet_invariant_check` | bool | `true` | Verify Assets = Liabilities + Equity before returning balance sheet. Errors are logged even if check is set to `warn` mode. |
| `balance_sheet_invariant_action` | enum | `error` | `error` — return error if balance sheet doesn't balance; `warn` — return result with a warning |

---

## 20. v1.0 Rollout Assessment

### Must Have at v1.0

- Trial balance report
- Balance sheet report (with invariant check)
- Income statement / P&L report
- AR aging report
- AP aging report
- GL detail (account ledger) report
- Period close status report (checklist view)
- JSON and Excel export formats
- Read replica routing (even a local replica is acceptable at v1.0)
- All reporting views with `security_invoker = true`
- Report permission model (`finance.reports.run`)

### Can Be Deferred

| Feature | Suggested Release |
|---|---|
| Cash flow statement (indirect method) | v1.1 — P&L + cash position covers immediate need |
| Budget variance report | v1.1 (requires budget module, which can be v1.1) |
| Cost centre P&L report | v1.1 (requires cost centre module to be in use) |
| PDF export | v1.1 — Excel and JSON cover v1.0 |
| Async report execution engine | v1.1 — at v1.0 data volumes, sync is acceptable |
| Report scheduling and delivery | v1.1 |
| Custom report definitions | v1.2 |
| Audit package export | v1.2 |
| IC balances report | v1.2 (requires IC module) |

### Never Defer

- Read replica connection pool separation — without it, large report queries will block journal postings
- `security_invoker = true` on all views — without it, RLS is bypassed for report queries
- Balance sheet invariant check — returning an unbalanced balance sheet is a data integrity failure that must be surfaced, not hidden

---

*End of FILE-08. Proceeding to FILE-09: Authorization, Roles & Segregation of Duties.*
