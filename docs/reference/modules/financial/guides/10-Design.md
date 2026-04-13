# Awo ERP — Finance Module
## FILE-10: Database Schema, Workflow Orchestration, Full API Reference & Configuration Catalogue

[<-- Back to Authorization](./09-Authorization.md)


**Document Version:** 2.0.0  
**Depends On:** [Overview](./01-Overview.md)  through [Authorization](./09-Authorization.md)  (all preceding files)  
**This file consolidates:** cross-cutting schema notes, Temporal workflows, the complete API surface, the master configuration catalogue, non-functional requirements, and deployment guidance.

---

## Table of Contents

1. [Database Design Principles (Cross-Cutting)](#1-database-design-principles-cross-cutting)
2. [Schema Conventions & Standards](#2-schema-conventions--standards)
3. [Cross-Table Indexes & Composite Keys](#3-cross-table-indexes--composite-keys)
4. [Temporal Workflow Orchestration](#4-temporal-workflow-orchestration)
5. [Complete API Surface](#5-complete-api-surface)
6. [Webhook Outbound Events](#6-webhook-outbound-events)
7. [Data Import Contracts](#7-data-import-contracts)
8. [Non-Functional Requirements](#8-non-functional-requirements)
9. [Deployment & Infrastructure Guidance](#9-deployment--infrastructure-guidance)
10. [Master Configuration Catalogue](#10-master-configuration-catalogue)
11. [Module Completion Checklist](#11-module-completion-checklist)

---

## 1. Database Design Principles (Cross-Cutting)

These principles apply to every table in the Finance module. They are not repeated in individual FILE documents — this section is the authoritative reference.

### 1.1 Every Table Has tenant_id as the First Column

```sql
CREATE TABLE {table_name} (
    tenant_id  UUID  NOT NULL REFERENCES tenants(id),
    id         UUID  NOT NULL DEFAULT gen_random_uuid(),
    ...
);
```

`tenant_id` is always the first column. This is a convention enforced by code review. It makes the column order on every table predictable and ensures the RLS policy always has a `tenant_id` to filter on.

### 1.2 Soft Delete Is Not Used for Financial Records

Posted financial records are never deleted — they are reversed or voided. This is a fundamental accounting principle. The Finance module has no `deleted_at` column on any table that holds posted financial data. For tables that represent configuration (accounts, cost centres, bank accounts), deactivation via an `is_active` flag is used instead of deletion.

The only records that can be hard-deleted are draft journal entries that have never been posted. This is implemented as a DELETE with a WHERE clause requiring `status = 'draft'`.

### 1.3 NUMERIC(18,4) for All Monetary Amounts

All monetary amounts are stored as `NUMERIC(18,4)` — 18 significant digits with 4 decimal places. This gives:
- Maximum value: 99,999,999,999,999.9999 (99 trillion)
- Sub-cent precision: supports currencies with 3 decimal places (e.g. KWD)
- No floating-point rounding errors (NUMERIC is exact, unlike FLOAT/DOUBLE)

Exchange rates use `NUMERIC(18,8)` — 8 decimal places for sufficient precision in rate calculations.

### 1.4 All Timestamps Are TIMESTAMPTZ

All timestamp columns use `TIMESTAMPTZ` (timestamp with time zone), stored as UTC. The application layer converts to local time (EAT, UTC+3) for display. This ensures:
- No DST ambiguity
- Correct time ordering across DST transitions
- Consistent timestamps when the server is in a different timezone than the user

### 1.5 UUIDs for All Primary Keys

All primary keys are `UUID` generated with `gen_random_uuid()`. Integer sequences are not used. UUIDs:
- Are safe to expose in API responses (not enumerable)
- Support distributed generation without coordination
- Are unambiguous across tables (no "is this ID from journal_entries or journal_lines?")

### 1.6 RLS Is Enabled AND Forced on Every Table

```sql
ALTER TABLE {table} ENABLE ROW LEVEL SECURITY;
ALTER TABLE {table} FORCE  ROW LEVEL SECURITY;
```

`FORCE ROW LEVEL SECURITY` ensures that the table owner (the application database user) is also subject to RLS. Without `FORCE`, a table owner bypasses RLS silently — a dangerous default.

### 1.7 Generated Columns for Derived Values

Where a column is always a function of other columns in the same row, use a `GENERATED ALWAYS AS STORED` column rather than computing it in application code. This ensures the derived value is always consistent with its inputs, even if data is modified by a support tool or migration script.

Examples already applied:
- `account_balances.closing_debit = opening_debit + period_debit`
- `bank_reconciliations.difference = adjusted_bank_balance - adjusted_gl_balance`
- `ar_subsidiary_entries.days_overdue = GREATEST(0, CURRENT_DATE - due_date)`

### 1.8 Outbox Pattern for Event Emission

Events emitted to other modules are written to an `outbox` table inside the same transaction as the state change. A separate process (or Temporal activity) reads the outbox and delivers events. This guarantees at-least-once delivery — the event is either delivered or the state change is rolled back with it.

```sql
CREATE TABLE outbox_events (
    tenant_id    UUID          NOT NULL,
    id           UUID          NOT NULL DEFAULT gen_random_uuid(),
    created_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    event_type   TEXT          NOT NULL,
    payload      JSONB         NOT NULL,
    delivered_at TIMESTAMPTZ,
    attempts     INT           NOT NULL DEFAULT 0,
    last_error   TEXT,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id)
);

CREATE INDEX idx_outbox_pending ON outbox_events (created_at ASC)
    WHERE delivered_at IS NULL;
```

---

## 2. Schema Conventions & Standards

### Column Naming

| Pattern | Convention | Example |
|---|---|---|
| Boolean flags | `is_` prefix | `is_active`, `is_group`, `is_system` |
| Timestamps (when something happened) | `_at` suffix | `posted_at`, `closed_at`, `approved_at` |
| User references (who did something) | `_by` suffix | `posted_by`, `closed_by`, `approved_by` |
| Foreign keys to other tables | `{table_singular}_id` | `account_id`, `period_id`, `entry_id` |
| Status fields | `status` — TEXT with CHECK constraint | `status TEXT CHECK (status IN ('draft', ...))` |
| Monetary amounts | `_amount` or `_balance` suffix | `debit_amount`, `net_balance`, `opening_debit` |

### Constraint Naming

| Type | Pattern | Example |
|---|---|---|
| Primary key | Implicit (PostgreSQL names it) | — |
| Unique constraint | `uq_{table}_{columns}` | `uq_accounts_org_code` |
| Foreign key | `fk_{table}_{ref_table}` | `fk_journal_lines_account` |
| Check constraint | `chk_{table}_{description}` | `chk_journal_lines_debit_xor_credit` |
| Index | `idx_{table}_{columns}` | `idx_journal_lines_account_date` |

### Policy Naming

All RLS policies follow the pattern: `{table_abbrev}_{role_abbrev}`

Examples: `je_app` (journal_entries, application_role), `ab_ro` (account_balances, readonly_role).

---

## 3. Cross-Table Indexes & Composite Keys

These indexes are not covered in individual FILE documents because they span the join paths between multiple tables.

```sql
-- GL detail query: join journal_lines → journal_entries → accounts
-- Most common report query pattern
CREATE INDEX idx_jl_account_date_covering
    ON journal_lines (tenant_id, account_id, created_at DESC)
    INCLUDE (debit_amount, credit_amount, description, entry_id, cost_centre_id, reference);

-- Budget check: join budget_lines → account_balances in pipeline stage 8
CREATE INDEX idx_bl_period_account
    ON budget_lines (tenant_id, budget_id, period_id, account_id)
    INCLUDE (amount, cost_centre_id);

-- AR/AP sub-ledger reconciliation: join subsidiary entries → journal_lines → accounts
CREATE INDEX idx_ar_unsettled_org
    ON ar_subsidiary_entries (tenant_id, organisation_id, is_settled, due_date)
    WHERE is_settled = FALSE;

-- Period close checklist: join journal_entries → accounting_periods
CREATE INDEX idx_je_period_status
    ON journal_entries (tenant_id, period_id, status);

-- Integration event idempotency: ensure source_event_id is globally unique
CREATE UNIQUE INDEX idx_je_source_event_id
    ON journal_entries (tenant_id, source_event_id)
    WHERE source_event_id IS NOT NULL;

-- Account path for hierarchy queries (already in FILE-02, repeated here for completeness)
CREATE INDEX idx_accounts_path_prefix
    ON accounts (tenant_id, path text_pattern_ops);

-- Cost centre path for hierarchy queries
CREATE INDEX idx_cost_centres_path_prefix
    ON cost_centres (tenant_id, path text_pattern_ops);
```

---

## 4. Temporal Workflow Orchestration

Temporal is used to coordinate multi-step, long-running processes that span multiple database operations and external module interactions. In the Finance module, three workflows are orchestrated by Temporal.

### 4.1 PeriodCloseWorkflow

Coordinates the 5–7 step period close process. This is the most important workflow — if any step fails, the workflow retries that step rather than requiring a full restart.

```go
func PeriodCloseWorkflow(ctx workflow.Context, input PeriodCloseInput) error {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 5 * time.Minute,
        RetryPolicy: &temporal.RetryPolicy{
            MaximumAttempts: 3,
            InitialInterval: 30 * time.Second,
        },
    }
    ctx = workflow.WithActivityOptions(ctx, ao)

    // Step 1: Run all mandatory checklist checks
    var checks []PeriodCloseCheck
    if err := workflow.ExecuteActivity(ctx, RunPeriodCloseChecksActivity, input).Get(ctx, &checks); err != nil {
        return fmt.Errorf("checklist checks failed: %w", err)
    }
    for _, c := range checks {
        if c.Blocking && !c.Passed {
            return fmt.Errorf("close blocked by %s: %s", c.Name, c.Detail)
        }
    }

    // Step 2: Trigger FX revaluation (emits event to Currency module)
    // Wait for JournalPosted confirmation before proceeding
    if input.HasFCBalances {
        if err := workflow.ExecuteActivity(ctx, TriggerAndAwaitFXRevaluationActivity, input.PeriodID).Get(ctx, nil); err != nil {
            return fmt.Errorf("FX revaluation failed: %w", err)
        }
    }

    // Step 3: Run IC matching check (if intercompany is enabled)
    if input.ICEnabled {
        if err := workflow.ExecuteActivity(ctx, RunICMatchingActivity, input.OrganisationID, input.PeriodEndDate).Get(ctx, nil); err != nil {
            return fmt.Errorf("IC matching failed: %w", err)
        }
    }

    // Step 4: Re-run checklist to catch anything that changed during FX/IC steps
    var finalChecks []PeriodCloseCheck
    if err := workflow.ExecuteActivity(ctx, RunPeriodCloseChecksActivity, input).Get(ctx, &finalChecks); err != nil {
        return err
    }
    for _, c := range finalChecks {
        if c.Blocking && !c.Passed {
            return fmt.Errorf("final check failed: %s", c.Name)
        }
    }

    // Step 5: Hard close the period
    return workflow.ExecuteActivity(ctx, HardClosePeriodActivity, input).Get(ctx, nil)
}
```

**Failure handling:** If step 2 (FX revaluation) fails because the Currency module is unavailable, Temporal retries it every 30 seconds for up to 3 attempts. If all retries are exhausted, the workflow fails with a clear error, the period remains `soft_closed`, and the finance manager is notified. No data is corrupted — the period is simply not hard-closed yet.

### 4.2 BankStatementImportWorkflow

Triggered by a bank webhook or on a nightly schedule. Fetches, parses, and auto-matches a bank statement.

```go
func BankStatementImportWorkflow(ctx workflow.Context, input BankImportInput) error {
    ao := workflow.ActivityOptions{StartToCloseTimeout: 10 * time.Minute}
    ctx = workflow.WithActivityOptions(ctx, ao)

    // Step 1: Fetch and parse statement from bank feed or uploaded file
    var lines []StatementLine
    if err := workflow.ExecuteActivity(ctx, FetchAndParseStatementActivity, input).Get(ctx, &lines); err != nil {
        return err
    }

    // Step 2: Open or continue reconciliation workspace for this period
    var recID uuid.UUID
    if err := workflow.ExecuteActivity(ctx, OpenOrContinueReconciliationActivity, input).Get(ctx, &recID); err != nil {
        return err
    }

    // Step 3: Run auto-matcher against unmatched GL entries
    return workflow.ExecuteActivity(ctx, RunAutoMatcherActivity, recID, lines).Get(ctx, nil)
}
```

### 4.3 Scheduled Workflows (Cron)

| Workflow | Schedule | Description |
|---|---|---|
| `AccrualReversalCron` | Daily 00:01 | Reverse any entries with `auto_reversal_date = today` |
| `BudgetPeriodOpenCron` | 1st of each month 00:05 | Carry forward prior period budget actuals for comparison |
| `ICMatchingCron` | Last day of month 23:00 | Run IC balance matching for all entity pairs |
| `OutboxDeliveryCron` | Every 30 seconds | Deliver pending outbox events to consuming modules |
| `ExpiredRoleRevocationCron` | Hourly | Revoke role assignments past their `expires_at` |
| `OutstandingItemAgingCron` | Daily 06:00 | Flag reconciliation outstanding items past age threshold |
| `AuditLogShippingCron` | Every 60 seconds | Ship new audit entries to SIEM (if enabled) |
| `ReportExecutionCleanupCron` | Daily 02:00 | Delete expired report result files from object storage |
| `ForecastCacheRefreshCron` | Daily 07:00 | Pre-warm cash forecast cache before business hours |

```go
func AccrualReversalCronWorkflow(ctx workflow.Context) error {
    ao := workflow.ActivityOptions{StartToCloseTimeout: 30 * time.Minute}
    ctx = workflow.WithActivityOptions(ctx, ao)
    return workflow.ExecuteActivity(ctx, ReverseMaturedAccrualsActivity, time.Now()).Get(ctx, nil)
}
```

---

## 5. Complete API Surface

All endpoints require:
- `Authorization: Bearer <token>`
- `X-Tenant: <tenant-slug>`
- `Content-Type: application/json` (for POST/PATCH)

### Standard Response Envelope

```json
// Success
{
  "success": true,
  "data": { ... },
  "meta": { "page": 1, "per_page": 50, "total": 240, "total_pages": 5 }
}

// Error
{
  "success": false,
  "error": {
    "code": "PERIOD_CLOSED",
    "message": "Accounting period January 2025 is closed.",
    "field": "transaction_date",
    "docs": "https://docs.awo-erp.com/errors/PERIOD_CLOSED"
  }
}
```

### Chart of Accounts

```
GET    /finance/accounts                           List (filter: root_type, account_type, is_active, is_group, q)
POST   /finance/accounts                           Create account
GET    /finance/accounts/{code}                    Get account + balance + recent activity
PATCH  /finance/accounts/{code}                    Update mutable fields
POST   /finance/accounts/{code}/deactivate         Deactivate
GET    /finance/accounts/{code}/ledger             GL detail (?from=, ?to=, ?page=, ?per_page=)
GET    /finance/accounts/tree                      Full COA tree with balances (?period_id=)
POST   /finance/accounts/import                    Bulk import from CSV
GET    /finance/account-mappings                   List semantic code mappings
POST   /finance/account-mappings                   Create or update mapping
DELETE /finance/account-mappings/{id}              Deactivate mapping
GET    /finance/account-mappings/validate          Check all required semantic codes are mapped
```

### Journal Entries

```
GET    /finance/journal-entries                    List (filter: status, type, from, to, account_code, reference, q)
POST   /finance/journal-entries                    Create draft
GET    /finance/journal-entries/{id}               Get entry with all lines
POST   /finance/journal-entries/{id}/submit        Draft → PendingApproval
POST   /finance/journal-entries/{id}/approve       PendingApproval → Approved
POST   /finance/journal-entries/{id}/reject        PendingApproval → Draft (with reason)
POST   /finance/journal-entries/{id}/post          Approved → Posted
POST   /finance/journal-entries/{id}/reverse       Posted → new Reversal Draft
POST   /finance/journal-entries/{id}/cancel        Draft | Approved → Cancelled
POST   /finance/journal-entries/import             Bulk import from CSV
GET    /finance/journal-entries/{id}/audit         Audit trail for this entry
```

### Fiscal Years & Periods

```
GET    /finance/fiscal-years                       List
POST   /finance/fiscal-years                       Create (auto-generates periods)
GET    /finance/periods                            List (?fiscal_year_id=)
GET    /finance/periods/{id}                       Get period + close checklist status
POST   /finance/periods/{id}/soft-close            Open → SoftClosed
POST   /finance/periods/{id}/hard-close            SoftClosed → HardClosed (runs checklist)
POST   /finance/periods/{id}/reopen                Any closed → Open (reason required)
POST   /finance/periods/{id}/lock                  HardClosed → Locked (CFO only)
```

### Currencies & Exchange Rates

```
GET    /finance/currencies                         List configured currencies
POST   /finance/rates                              Load rates (batch)
GET    /finance/rates                              Query rates (?from_currency=, ?to_currency=, ?date=)
GET    /finance/rates/latest                       Most recent rate for each active pair
```

### Bank Reconciliation

```
GET    /finance/reconciliations                    List (?bank_account_id=, ?period_id=, ?status=)
POST   /finance/reconciliations                    Open new workspace
POST   /finance/reconciliations/{id}/import        Import bank statement (multipart/form-data)
GET    /finance/reconciliations/{id}/matches       Auto-match suggestions
POST   /finance/reconciliations/{id}/matches       Confirm matches / classify timing differences
DELETE /finance/reconciliations/{id}/matches/{id}  Unmatch
POST   /finance/reconciliations/{id}/submit        InProgress → UnderReview
POST   /finance/reconciliations/{id}/approve       UnderReview → Approved
POST   /finance/reconciliations/{id}/lock          Approved → Locked
GET    /finance/reconciliations/{id}/statement     Download reconciliation statement (PDF)
```

### AR & AP

```
GET    /finance/ar/customers                       AR balances per customer
GET    /finance/ar/customers/{id}                  Single customer AR balance + open invoices
GET    /finance/ar/aging                           AR aging (?as_at=, ?buckets=30,60,90)
POST   /finance/ar/write-off                       Write off invoice (approval required)
GET    /finance/ar/reconciliation                  AR control vs. subsidiary check
GET    /finance/ap/suppliers                       AP balances per supplier
GET    /finance/ap/suppliers/{id}                  Single supplier AP balance
GET    /finance/ap/aging                           AP aging (?as_at=, ?buckets=30,60,90)
GET    /finance/ap/due-this-week                   Invoices due in next 7 days
GET    /finance/ap/reconciliation                  AP control vs. subsidiary check
```

### Payment Runs

```
GET    /finance/payment-runs                       List
POST   /finance/payment-runs                       Create from due AP invoices
PATCH  /finance/payment-runs/{id}/items            Update included/excluded items
POST   /finance/payment-runs/{id}/approve          Draft → Approved
POST   /finance/payment-runs/{id}/export           Generate bank payment file
POST   /finance/payment-runs/{id}/confirm          Post GL entries after bank processing
POST   /finance/payment-runs/{id}/cancel           Cancel (only before export)
```

### Petty Cash

```
GET    /finance/petty-cash/funds                   List funds
GET    /finance/petty-cash/funds/{id}              Fund details + balance + recent vouchers
POST   /finance/petty-cash/funds/{id}/disburse     Create disbursement voucher
POST   /finance/petty-cash/funds/{id}/replenish    Replenish (creates journal entry)
POST   /finance/petty-cash/funds/{id}/count        Record physical count
```

### Cost Centres & Budgets

```
GET    /finance/cost-centres                       List (?is_group=, ?is_active=)
POST   /finance/cost-centres                       Create
GET    /finance/cost-centres/{code}                Get with period balances
PATCH  /finance/cost-centres/{code}                Update
POST   /finance/cost-centres/allocate              Trigger distributed allocation for period
GET    /finance/cost-centres/{code}/pl             Cost centre P&L (?from=, ?to=)
GET    /finance/budgets                            List (?fiscal_year_id=, ?status=)
POST   /finance/budgets                            Create draft
GET    /finance/budgets/{id}                       Get with all lines
POST   /finance/budgets/{id}/lines                 Set budget lines
POST   /finance/budgets/{id}/import                Bulk import lines from CSV
POST   /finance/budgets/{id}/approve               Draft → Approved
POST   /finance/budgets/{id}/activate              Approved → Active
POST   /finance/budgets/{id}/copy                  Copy lines from another version
GET    /finance/budgets/{id}/variance              Budget vs. actual variance
```

### Cash Management

```
GET    /finance/cash/position                      Real-time cash position
GET    /finance/cash/forecast                      13-week rolling forecast (?force_refresh=)
GET    /finance/cash/committed                     Committed items (approved payment runs etc.)
```

### Intercompany

```
GET    /finance/ic/transactions                    List IC transactions
POST   /finance/ic/transactions                    Create IC transaction
GET    /finance/ic/balances                        Balance by entity pair (?as_at=)
POST   /finance/ic/match                           Run IC matching for period
POST   /finance/ic/eliminate                       Generate elimination entries for consolidation
GET    /finance/ic/eliminations                    List elimination entries
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
POST   /finance/reports/vat-return/run
POST   /finance/reports/wht-register/run
GET    /finance/reports/period-close-status

GET    /finance/report-executions/{id}             Poll async execution
GET    /finance/report-executions/{id}/download    Download result (?format=pdf|excel|csv|json)

GET    /finance/report-definitions                 List available reports
POST   /finance/report-definitions                 Create custom report
PATCH  /finance/report-definitions/{id}            Update custom report

GET    /finance/report-schedules                   List schedules
POST   /finance/report-schedules                   Create schedule
PATCH  /finance/report-schedules/{id}              Update schedule
DELETE /finance/report-schedules/{id}              Delete schedule
POST   /finance/report-schedules/{id}/run-now      Trigger immediate execution

POST   /finance/audit-package/request              Request audit package for fiscal year
GET    /finance/audit-package/status/{id}          Check generation status
GET    /finance/audit-package/download/{id}        Download (7-day link)
```

### Authorization & Audit

```
GET    /auth/users/{id}/roles                      List user roles
POST   /auth/users/{id}/roles                      Assign role
DELETE /auth/users/{id}/roles/{role}               Revoke role
GET    /auth/sod-matrix                            View incompatible combinations
POST   /auth/sod-check                             Check proposed assignment
GET    /finance/delegations                        List delegations
POST   /finance/delegations                        Create delegation
DELETE /finance/delegations/{id}                   Deactivate
GET    /finance/audit-log                          Query audit entries
GET    /finance/audit-log/{id}                     Single entry with full before/after
POST   /finance/audit-log/export                   Export audit log (async)
GET    /finance/settings                           Get finance configuration
PATCH  /finance/settings                           Update configuration (CFO only)
```

---

## 6. Webhook Outbound Events

The Finance module emits these events that other modules (and external integrations) can subscribe to.

### Event Catalogue

| Event | Trigger | Payload |
|---|---|---|
| `finance.journal.posted` | Journal entry reaches POSTED status | `entry_id`, `reference`, `amount`, `currency`, `source_module`, `source_event_id`, `posted_at` |
| `finance.journal.reversed` | Reversal entry created | `reversal_id`, `original_entry_id`, `reference`, `reversed_at` |
| `finance.period.soft_closed` | Period soft-closed | `period_id`, `period_name`, `closed_by`, `closed_at` |
| `finance.period.hard_closed` | Period hard-closed | `period_id`, `period_name`, `closed_by`, `closed_at` |
| `finance.period.locked` | Period locked permanently | `period_id`, `period_name`, `locked_by`, `locked_at` |
| `finance.reconciliation.approved` | Bank rec approved and locked | `reconciliation_id`, `bank_account_id`, `period_id`, `approved_by`, `approved_at` |
| `finance.payment_run.posted` | Payment run GL entries posted | `payment_run_id`, `total_amount`, `item_count`, `posted_at` |
| `finance.budget.activated` | New budget version activated | `budget_id`, `fiscal_year_id`, `version_label`, `activated_by` |

### Webhook Payload Format

```json
{
  "event":      "finance.journal.posted",
  "tenant_id":  "a1b2c3d4-0000-0000-0000-000000000001",
  "timestamp":  "2025-02-06T09:15:42Z",
  "version":    "1",
  "data": {
    "entry_id":        "uuid",
    "reference":       "JE-2025-0145",
    "amount":          50000.00,
    "currency":        "KES",
    "source_module":   "manual",
    "posted_at":       "2025-02-06T09:15:42Z"
  },
  "signature": "sha256=abc123..."
}
```

Webhooks are signed with HMAC-SHA256 using the tenant's webhook secret. The signature covers the full JSON body. Retries: 3 attempts with exponential backoff (5 minutes, 30 minutes, 2 hours). After 3 failures, the event is marked `delivery_failed` and the tenant is notified.

---

## 7. Data Import Contracts

### Journal Entry Import (CSV)

```csv
transaction_date,reference,description,account_code,debit,credit,cost_centre_code,currency,notes
2025-01-15,CHQ-001234,Office rent January,7310,50000,,CC-ADMIN,,
2025-01-15,CHQ-001234,Office rent January,1112,,50000,,,
2025-01-20,INV-0451,Sale to ABC Corp,1121,580000,,,KES,
2025-01-20,INV-0451,Sale to ABC Corp,4110,,500000,,,
2025-01-20,INV-0451,Sale to ABC Corp,2310,,80000,,,VAT 16%
```

- Rows with the same `reference` are grouped into one journal entry
- Each group must balance (total debit = total credit)
- Missing `currency` defaults to tenant functional currency
- Import is transactional — all rows succeed or none commit

### Chart of Accounts Import (CSV)

```csv
code,name,root_type,account_type,parent_code,is_group,allow_manual_entries,require_reference,locked_currency
1000,Assets,asset,root,,true,false,false,
1100,Current Assets,asset,group,1000,true,false,false,
1111,Petty Cash,asset,cash,1110,false,true,false,KES
```

### Budget Import (CSV)

```csv
account_code,cost_centre_code,jan,feb,mar,apr,may,jun,jul,aug,sep,oct,nov,dec
7110,CC-NAIROBI,1000000,1000000,1000000,1000000,1000000,1000000,1000000,1000000,1000000,1000000,1000000,1000000
6300,CC-MARKETING,500000,200000,200000,800000,200000,200000,500000,200000,200000,800000,200000,300000
```

### Opening Balance Import (CSV)

```csv
account_code,debit,credit,currency,notes
1112,1250000,,KES,Opening cash per bank statement
1121,3450000,,KES,Opening AR per aged debtors list
2110,,1850000,KES,Opening AP per aged creditors list
3300,,695000,KES,Retained earnings (balancing figure)
```

Opening balance imports must balance (total debit = total credit) and may only reference balance sheet accounts.

---

## 8. Non-Functional Requirements

### API Performance Targets

| Operation | Target P95 | Target P99 | Max Acceptable |
|---|---|---|---|
| Journal entry create (draft save) | < 200ms | < 500ms | 1s |
| Journal entry post (full pipeline + GL write) | < 500ms | < 1s | 2s |
| Account balance query (single account) | < 5ms | < 10ms | 50ms |
| Trial balance (200 accounts) | < 100ms | < 300ms | 1s |
| Balance sheet render | < 200ms | < 500ms | 2s |
| AR/AP aging (2,000 open items) | < 500ms | < 1s | 5s |
| GL detail (1 account, 1 year) | < 500ms | < 1s | 5s |
| Bank statement import (500 lines) | < 3s | < 8s | 30s |
| Auto-match engine (500 lines) | < 2s | < 5s | 15s |
| Exchange rate load (20 currencies) | < 500ms | < 1s | 5s |
| Cash position | < 10ms | < 30ms | 100ms |
| Period close checklist | < 5s | < 10s | 30s |

Reports exceeding 5s are promoted to async execution. Hard timeout on all queries: 120s.

### Scalability Targets

**Per tenant:**
- Up to 500,000 GL entries per fiscal year
- Up to 10,000 accounts in COA
- Up to 500 cost centres
- Up to 100 concurrent API users

**Platform-wide:**
- Up to 1,000 active tenants on shared infrastructure
- Up to 50,000 journal entries per minute (platform total)
- Up to 10,000 report executions per hour

### Availability

| Component | Target Monthly Uptime |
|---|---|
| API (journal posting, core operations) | 99.9% (≤ 43.8 min downtime/month) |
| Report engine (read replica) | 99.5% (reports queue when replica unavailable) |
| Temporal workflows | 99.5% (workflows retry on restart) |
| Webhook delivery | 99% (retries up to 3 attempts) |

RTO: 1 hour (database failover + application restart).  
RPO: 5 minutes (continuous WAL archiving, streaming replication).

### Security Requirements

| Requirement | Implementation |
|---|---|
| Data at rest | AES-256 encryption on all storage volumes |
| Data in transit | TLS 1.3 minimum for all API traffic |
| Sensitive columns (bank account numbers) | Column-level encryption using `pgcrypto` |
| Multi-tenancy isolation | PostgreSQL RLS + tenant context in every session |
| Audit completeness | 100% of state-changing operations logged |
| Brute force protection | Rate limiting per API key (120 req/min) |
| Token expiry | Access token: 1 hour; Refresh token: 7 days |
| MFA requirement | Required for roles with approve/post/lock permissions |

### Data Retention

| Data Type | Minimum Retention | Archive After | Delete After |
|---|---|---|---|
| Journal entries and lines | 7 years | 5 years | Never (permanent) |
| Audit log | 7 years | 3 years | 10 years |
| Report results (async) | 24 hours | N/A | 24 hours |
| Audit packages | 7 years | — | Never |
| Exchange rates | 10 years | 5 years | Never |
| Outbox events (delivered) | 30 days | — | 30 days |

---

## 9. Deployment & Infrastructure Guidance

### Database Configuration

**Primary (write) database:**
```
PostgreSQL 16+
vCPU: 4–8 (write workload is IO-bound, not CPU-bound)
RAM: 16–32 GB
Storage: NVMe SSD, provisioned IOPS
shared_buffers: 4 GB (25% of RAM)
work_mem: 4 MB (conservative — many concurrent connections)
max_connections: 200
wal_level: replica
archive_mode: on
archive_command: configured for WAL archiving to object storage
```

**Read replica (reporting database):**
```
PostgreSQL 16+ streaming replica
vCPU: 4 (CPU available for parallel queries)
RAM: 16 GB
shared_buffers: 6 GB (40% of RAM — more aggressive than primary)
work_mem: 64 MB (reporting queries may need large sort buffers)
max_parallel_workers_per_gather: 4
effective_cache_size: 12 GB (75% of RAM)
hot_standby: on
max_standby_streaming_delay: 30s (acceptable lag for reporting)
```

### Connection Pooling

PgBouncer is recommended between the application and PostgreSQL:

```
Primary pool:
  pool_mode: transaction  (connections released after each transaction)
  max_client_conn: 500
  default_pool_size: 25   (25 server-side connections for write workload)

Replica pool:
  pool_mode: session      (held for duration of report query)
  max_client_conn: 200
  default_pool_size: 20
```

**Why session mode for the replica?** Report queries may run for several seconds. In transaction mode, the server-side connection is released between statements — this breaks the cursor-based pattern used for large result set streaming. Session mode holds the connection for the full duration of the report query.

### Redis

Redis is used for:
- Permission cache (user role → resolved permissions, TTL 5 minutes)
- COA tree cache (per tenant, TTL 5 minutes)
- Cash forecast cache (TTL configurable, default 30 minutes)
- Idempotency keys for API requests (TTL 24 hours)

Minimum Redis spec: 1 GB RAM, persistence enabled (AOF with `appendfsync everysec`).

### Object Storage

Used for:
- Async report result files (JSON, Excel, PDF, CSV)
- Audit package ZIP archives
- Bank statement file uploads (staging before import)

All objects are stored with server-side encryption. Report results are automatically deleted after their TTL via lifecycle rules.

### Temporal

Temporal is optional at v1.0 but required for v1.1. At v1.0, period close and bank import workflows can be triggered synchronously via API.

Temporal deployment: minimum 1 frontend + 1 history + 1 matching service (can run on a single host for development, separate hosts for production).

---

## 10. Master Configuration Catalogue

This is the consolidated list of all configuration flags across all Finance module sections. All flags are stored per-tenant in `tenant_config.finance`.

### Core

| Flag | Type | Default | Irreversible? |
|---|---|---|---|
| `default_currency` | string | `KES` | Yes — after first transaction |
| `fiscal_year_start_month` | int (1–12) | `1` | Yes |
| `fiscal_year_period_type` | enum | `monthly` | No |
| `journal_reference_prefix` | string | `JE` | No |
| `journal_reference_sequence_length` | int | `4` | No |

### Journal & Posting Control

| Flag | Type | Default |
|---|---|---|
| `approval_required_above_amount` | decimal | `null` |
| `approval_required_for_integration` | bool | `false` |
| `duplicate_check_enabled` | bool | `true` |
| `duplicate_check_window_hours` | int | `24` |
| `duplicate_action` | enum | `warn` |
| `auto_reverse_accruals` | bool | `true` |
| `max_lines_per_entry` | int | `200` |
| `integration_dead_letter_enabled` | bool | `true` |

### Chart of Accounts

| Flag | Type | Default |
|---|---|---|
| `coa_max_depth` | int | `10` |
| `coa_code_format` | string | `null` |
| `coa_template` | string | `kenya_sme_general` |
| `auto_create_balance_sheet_groups` | bool | `true` |
| `auto_year_end_closing` | bool | `false` |
| `require_account_mapping_for_integration` | bool | `true` |

### Period & Reconciliation

| Flag | Type | Default |
|---|---|---|
| `bank_rec_required_for_close` | bool | `true` |
| `bank_rec_tolerance` | decimal | `0.01` |
| `outstanding_cheque_alert_days` | int | `60` |
| `auto_match_confidence_threshold` | enum | `auto` |
| `bank_statement_import_formats` | string[] | `["ofx","mt940","csv"]` |
| `fx_revaluation_required_for_close` | bool | `true` |
| `subsidiary_reconcile_required_for_close` | bool | `true` |
| `ar_sub_rec_tolerance` | decimal | `0.01` |
| `ap_sub_rec_tolerance` | decimal | `0.01` |
| `period_close_mode` | enum | `workflow` |
| `allow_period_reopen_self` | bool | `false` |
| `require_reason_for_period_reopen` | bool | `true` |

### AR & AP

| Flag | Type | Default |
|---|---|---|
| `ar_credit_limit_enabled` | bool | `false` |
| `ar_credit_limit_action` | enum | `warn` |
| `ap_duplicate_check_days` | int | `30` |
| `ap_backdated_invoice_days` | int | `90` |
| `early_payment_discount_enabled` | bool | `false` |

### Payments & Petty Cash

| Flag | Type | Default |
|---|---|---|
| `payment_run_enabled` | bool | `true` |
| `payment_run_export_formats` | string[] | `["generic_csv"]` |
| `payment_run_require_approval` | bool | `true` |
| `payment_run_max_amount_without_cfo` | decimal | `500000` |
| `payment_file_format` | enum | `generic_csv` |
| `petty_cash_enabled` | bool | `true` |
| `petty_cash_max_single_disbursement` | decimal | `5000` |
| `petty_cash_shortage_tolerance` | decimal | `50` |

### Cost Centres & Budgets

| Flag | Type | Default |
|---|---|---|
| `cost_centre_enabled` | bool | `true` |
| `require_cost_centre_on_expense` | bool | `false` |
| `budget_module_enabled` | bool | `true` |
| `budget_control_mode` | enum | `soft` |
| `budget_control_mode_per_account` | bool | `false` |
| `budget_override_requires_reason` | bool | `true` |
| `budget_check_includes_pending` | bool | `false` |
| `distributed_allocation_auto_run` | bool | `false` |

### Multi-Currency & FX

| Flag | Type | Default |
|---|---|---|
| `multi_currency_enabled` | bool | `true` |
| `fx_revaluation_required_for_close` | bool | `true` |
| `rate_deviation_alert_threshold` | decimal | `0.05` |
| `rate_auto_load_enabled` | bool | `false` |
| `rate_auto_load_source` | string | `null` |
| `rate_auto_load_cron` | string | `0 8 * * *` |
| `unrealised_fx_treatment` | enum | `pnl` |

### Cash Management & IC

| Flag | Type | Default |
|---|---|---|
| `cash_forecast_enabled` | bool | `false` |
| `cash_forecast_cache_ttl` | int | `30` |
| `cash_target_minimum_balance` | decimal | `0` |
| `multi_entity_enabled` | bool | `false` |
| `intercompany_enabled` | bool | `false` |
| `ic_matching_required_for_close` | bool | `true` |
| `ic_auto_create_counterparty_entry` | bool | `false` |

### Reporting

| Flag | Type | Default |
|---|---|---|
| `report_async_threshold_ms` | int | `5000` |
| `report_cache_enabled` | bool | `true` |
| `report_default_cache_ttl` | int | `300` |
| `report_storage_ttl_hours` | int | `24` |
| `report_audit_package_ttl_days` | int | `7` |
| `report_max_rows_sync` | int | `50000` |
| `report_pdf_enabled` | bool | `true` |
| `report_excel_enabled` | bool | `true` |
| `report_scheduling_enabled` | bool | `false` |
| `balance_sheet_invariant_check` | bool | `true` |
| `balance_sheet_invariant_action` | enum | `error` |

### Authorization & Audit

| Flag | Type | Default |
|---|---|---|
| `finance_manager_approval_limit` | decimal | `500000` |
| `ceo_approval_required_above` | decimal | `null` |
| `ceo_approval_blocks` | bool | `false` |
| `multi_approver_threshold` | decimal | `5000000` |
| `sod_enforcement_mode` | enum | `block` |
| `auditor_role_max_days` | int | `90` |
| `audit_log_shipping_enabled` | bool | `false` |
| `audit_log_siem_endpoint` | string | `null` |
| `permission_cache_ttl_seconds` | int | `300` |
| `delegation_max_days` | int | `30` |

---

## 11. Module Completion Checklist

This checklist is for the engineering team to track implementation completeness before each release.

### Non-Negotiable Pre-Launch Items (v1.0 Gate)

```
Database Layer
  ✓ RLS ENABLED + FORCED on every Finance table
  ✓ Period gate trigger (fn_check_period_gate) with FOR SHARE locking
  ✓ Journal entry immutability trigger (fn_protect_posted_journal)
  ✓ Journal line immutability trigger (fn_protect_posted_line)
  ✓ chk_debit_xor_credit constraint on journal_lines
  ✓ UNIQUE (tenant_id, source_event_id) on journal_entries
  ✓ Generated columns on account_balances (closing_debit, closing_credit, net_balance)
  ✓ Generated column on bank_reconciliations (difference)
  ✓ Generated column on ar_subsidiary_entries (days_overdue)
  ✓ Partial unique index for single active budget per org per fiscal year
  ✓ Audit trigger on journal_entries and accounting_periods
  ✓ Audit log INSERT-only policy (no UPDATE or DELETE for application_role)
  ✓ chk_auditor_must_expire constraint on role_assignments

Application Layer
  ✓ All 10 journal entry pipeline stages implemented
  ✓ PostToGL stage: period row locking with SELECT FOR UPDATE
  ✓ Idempotency check in all integration event consumers
  ✓ Self-approval prohibition in CheckApproval stage
  ✓ Self-approval prohibition in BankReconciliation.Approve()
  ✓ Self-approval prohibition in PaymentRun approval service
  ✓ SOD enforcement at role assignment (block mode)
  ✓ Authority limits in ApprovalService.Route()
  ✓ Balance sheet invariant check in ReportService
  ✓ Tenant context set in every database connection acquisition
  ✓ Read replica routing for all report queries

Security
  ✓ TLS 1.3 on all external endpoints
  ✓ JWT validation on every API request
  ✓ Rate limiting per API key
  ✓ RLS tenant isolation verified in integration tests

Testing
  ✓ Integration test: unbalanced entry is rejected
  ✓ Integration test: posting to hard-closed period is rejected (via trigger)
  ✓ Integration test: posting same event twice produces one journal entry (idempotency)
  ✓ Integration test: self-approval is rejected
  ✓ Integration test: RLS prevents cross-tenant data access
  ✓ Integration test: balance sheet balances after posting
  ✓ Integration test: AR control reconciles to subsidiary after posting
  ✓ Load test: 100 concurrent journal postings complete in < 2s each
```

### v1.1 Target Features

```
  □ 13-week cash flow forecast engine
  □ Async report execution engine
  □ Report scheduling and email delivery
  □ OFX and MT940 bank statement import
  □ Auto-matcher M-04 (batch match)
  □ AR/AP aging materialised views on read replica
  □ Temporal period close workflow
  □ Temporal accrual reversal cron
  □ Budget hard-stop control
  □ Auditor role with auto-expiry enforcement
  □ Approval delegation
  □ Multi-tier approval (Finance Manager + CFO sequential)
  □ Cost centre balance materialisation table
  □ Audit log partitioning
  □ Petty cash physical count workflow

### v1.2 Target Features

  □ Multi-entity support
  □ Intercompany module (IC transactions, matching, eliminations)
  □ Per-account budget control mode override
  □ Custom report definitions
  □ Audit package export
  □ Portal roles (customer and supplier)
  □ SWIFT MT101 payment file format
  □ Rolling budget engine
  □ Cost centre ABAC scoping
  □ Dual-authorisation for fiscal year lock
```

---

*End of FILE-10 — Final file in the Awo ERP Finance Module documentation series.*

---

## Document Series Index

| File | Title | Key Content |
|---|---|---|
| FILE-01 | Overview, Design Philosophy & Domain Model | Why it exists, ERP comparisons, design principles, domain model, dependencies, v1.0 scope |
| FILE-02 | Chart of Accounts Engine | COA hierarchy, materialised path, account types, balance materialisation, account mapping |
| FILE-03 | Journal Entry Pipeline & General Ledger | 10-stage pipeline, GL engine, subsidiary ledger, reversal, recurring entries, integration consumers |
| FILE-04 | Period Management & Bank Reconciliation | Period status machine, period gate, close workflow, bank rec workspace, auto-matcher |
| FILE-05 | AR, AP, Payments & Petty Cash | Sub-ledgers, aging, payment runs, petty cash imprest |
| FILE-06 | Cost Centre & Budget Management | Cost centre tree, distributed allocation, budget model, budget checker, versioning |
| FILE-07 | Cash Management, Intercompany & Multi-Currency | Cash position, 13-week forecast, FX architecture, exchange rates, revaluation, IC module |
| FILE-08 | Financial Reporting Engine & Report Catalogue | Read replica, view layer, report definition model, all reports, async execution, scheduling |
| FILE-09 | Authorization, Roles & Segregation of Duties | MRA framework, all roles, SOD enforcement, authority limits, RLS, audit trail |
| FILE-10 | Database Schema, Workflows, API & Configuration | Cross-cutting schema, Temporal workflows, complete API surface, master config catalogue |
