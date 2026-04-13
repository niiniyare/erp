# Awo ERP — Finance Module
## FILE-01: Overview, Design Philosophy & Domain Model

**Document Version:** 2.0.0  
**Status:** Authoritative Specification  
**Audience:** Engineering · Finance · Operations · Compliance  
**Stack:** Go · PostgreSQL · Temporal  
**Series:** FILE-01 of 10

---

## Table of Contents

1. [What Is the Finance Module](#1-what-is-the-finance-module)
2. [Why It Exists in This ERP](#2-why-it-exists-in-this-erp)
3. [What You Lose Without It](#3-what-you-lose-without-it)
4. [How ERPNext, NetSuite and QuickBooks Approach This](#4-how-erpnext-netsuite-and-quickbooks-approach-this)
5. [Design Principles](#5-design-principles)
6. [Scope Boundary](#6-scope-boundary)
7. [Domain Model](#7-domain-model)
8. [Module Dependencies](#8-module-dependencies)
9. [Feature Flags & Configuration](#9-feature-flags--configuration)
10. [v1.0 Rollout Assessment](#10-v10-rollout-assessment)

---

## 1. What Is the Finance Module

The Finance module is the authoritative ledger of Awo ERP. It is the single destination for every monetary event that occurs anywhere in the system — a fuel sale at a forecourt, a payroll disbursement, a supplier payment, a foreign currency revaluation, an inventory write-off. All of these ultimately produce a journal entry that lands in this module's General Ledger.

No other module writes journal entries directly to the ledger. Other modules produce business events; the Finance module consumes those events, applies account mapping rules, and creates the corresponding double-entry records. This separation is deliberate and non-negotiable — it is what keeps the books clean regardless of how many operational modules are added over time.

In practical terms, the Finance module owns:

- The chart of accounts and account hierarchy
- Fiscal years and accounting periods
- Every journal entry ever posted
- All account balances (materialised per period)
- Bank accounts and bank reconciliation workspaces
- Accounts receivable and payable sub-ledgers
- Cost centres and budget control
- Cash position and rolling forecasts
- Intercompany transactions and eliminations
- All financial statement generation

---

## 2. Why It Exists in This ERP

### The Integration Advantage

The fundamental value of an embedded finance module — rather than a separate accounting package — is that business operations and financial records share the same database, the same entities, and the same transaction boundary.

When a fuel sale is confirmed in the Forecourt module, the Finance module receives the event in the same atomic transaction. There is no file export, no nightly sync, no manual re-entry. The journal entry exists the moment the sale exists. This is the core promise of ERP finance.

**Without an integrated finance module:**

- A fuel sale would need to be exported from the Forecourt module and imported into a separate accounting system (e.g., QuickBooks or a local Sage installation)
- Inventory values in the operational system would diverge from the values in the accounting system within hours
- Reconciling operational data to financial data becomes a monthly exercise measured in days, not minutes
- Every audit requires pulling data from two systems and explaining why they differ

**With an integrated finance module:**

- Every operational event that has a financial consequence produces a journal entry immediately
- The balance sheet and income statement are always current — not last-month's figures
- Drill-down from any financial report goes directly to the source transaction
- There is one version of truth across operations and finance

### Why Build It Rather Than Integrate

Awo ERP targets organisations (initially petroleum retail, then broader East African commercial operations) that currently manage finances through a combination of spreadsheets, QuickBooks, or ERPNext alongside separate operational tools. The gap between those tools and the operational reality of a multi-pump, multi-shift fuel station is large.

A purpose-built finance module allows Awo to:

1. Define account mappings that match petroleum retail semantics (`FUEL_REVENUE`, `WET_STOCK_VARIANCE`, `SHIFT_CASH_SHORTAGE`) rather than generic ones
2. Enforce posting rules that reflect local regulatory requirements (KRA, KEBS, NEMA compliance journals)
3. Build period-close workflows that align with the forecourt shift cycle, not just a calendar month
4. Support the specific foreign currency exposure profile of East African fuel importers (USD-denominated crude, KES functional currency)

---

## 3. What You Lose Without It

This section is written for decision-makers evaluating whether a full finance module is necessary at launch versus using a third-party integration.

| Capability | Without Finance Module | With Finance Module |
|---|---|---|
| Real-time P&L | Unavailable — requires nightly export/import | Available instantly |
| Drill-down from report to transaction | Requires jumping between two systems | Single click |
| Multi-currency revaluation | Manual spreadsheet at month-end | Automated on event |
| Bank reconciliation | External tool, manual matching | Built-in workspace |
| Budget vs. actual on expense | Not possible in real-time | Per-posting check |
| Period close integrity | No enforcement | Gate enforced at DB trigger level |
| Intercompany eliminations | Manual consolidation spreadsheet | Automated matching |
| Audit trail | Fragmented across two systems | Single append-only ledger |
| Regulatory compliance (VAT, WHT) | Separate tax tool or manual | Embedded in posting rules |

**The honest trade-off:** Integrating a third-party accounting package (ERPNext, QuickBooks via API) is faster to launch but creates a synchronisation liability that grows with transaction volume. At 500 fuel transactions a day across a multi-site operator, that liability becomes a significant operational burden within three months.

---

## 4. How ERPNext, NetSuite and QuickBooks Approach This

Understanding how established systems handle general ledger finance helps justify Awo's specific design choices.

### ERPNext

ERPNext uses a DocType-based general ledger where every financial document (Sales Invoice, Purchase Invoice, Payment Entry, Journal Entry) posts GL entries via a centralised `make_gl_entries` function. The chart of accounts is a tree, accounts have a `root_type` and an `account_type`, and balances are re-computed from `GL Entry` rows on every report run rather than being materialised.

**What Awo adopts from ERPNext:**
- The tree-based chart of accounts with `root_type` classification
- The semantic account mapping pattern (ERPNext calls these "default accounts" on Company and Item masters)
- The concept of `Cost Center` as a separate dimension from the account hierarchy

**What Awo improves on:**
- ERPNext recomputes balances by summing `GL Entry` rows every time a report runs. At scale this is slow. Awo maintains `account_balances` as a materialised table updated atomically inside the posting transaction — balance queries are O(1) regardless of history depth.
- ERPNext's period close is advisory; it can be bypassed. Awo enforces period gating at the PostgreSQL trigger level — an insert into `journal_lines` against a closed period is rejected by the database, not just the application.
- ERPNext does not have a built-in bank reconciliation engine. It relies on the Bank Reconciliation Statement tool which is largely manual. Awo includes an auto-matcher with rule-based confidence scoring.

### NetSuite

NetSuite is the reference implementation for multi-entity, multi-currency ERP finance. Its OneWorld module handles consolidation, intercompany eliminations, and currency translation at enterprise scale.

**What Awo adopts from NetSuite:**
- Multi-entity design from day one (each `Organisation` is a legal entity with its own COA, periods, and GL)
- The subsidiary ledger pattern — AR, AP, inventory roll up to GL control accounts and are reconciled separately
- The `Period Close Checklist` concept — a structured set of verifiable gates before a period can close

**What Awo does differently:**
- NetSuite requires significant configuration to achieve period-gate enforcement. Awo enforces it structurally via DB triggers.
- NetSuite's intercompany matching is powerful but complex to configure. Awo's IC matching is opinionated and automatic for common patterns.
- NetSuite pricing is prohibitive for the East African SME market Awo targets.

### QuickBooks

QuickBooks (Online and Desktop) is the most widely used accounting package in Awo's target market. Many of the businesses Awo serves are currently on QuickBooks.

**What QuickBooks does well:**
- Ease of use — a non-accountant can be productive in hours
- Bank feed integration and auto-categorisation
- Simple P&L and balance sheet for single-entity, single-currency businesses

**Where QuickBooks falls short for Awo's target customers:**
- No multi-entity support in standard tiers
- No approval workflows on journal entries
- No budget control at the transaction level
- No period gating — you can post to any date at any time
- No cost centre P&L (limited class tracking only)
- No programmatic double-entry API — integrating an operational system requires workarounds
- No intercompany accounting

**What Awo takes from QuickBooks:**
- The emphasis on understandable bank reconciliation UX — QuickBooks' reconciliation interface is genuinely well-designed for non-accountants
- The idea of a "smart suggestion" during bank matching (Awo's auto-matcher confidence levels map to this)

---

## 5. Design Principles

These principles are not preferences — they are architectural constraints that shape every implementation decision in the Finance module.

### 5.1 Double-Entry at the Database Layer

The GL enforces double-entry integrity at the constraint level, not just the application layer.

```sql
-- journal_lines: each line is debit XOR credit, never both, never neither
CONSTRAINT chk_debit_or_credit CHECK (
    (debit_amount > 0 AND credit_amount = 0) OR
    (credit_amount > 0 AND debit_amount = 0)
)
```

Balance validation (total debits = total credits) is enforced in the `ValidateBalance` pipeline stage before any line reaches the database. An unbalanced entry cannot exist in the database even if the application has a bug.

**Why this matters:** Application bugs, concurrent requests, and direct database access by support tooling are all real scenarios. A ledger that can only be corrupted by violating a database constraint is far more trustworthy than one that relies on application code alone.

### 5.2 Append-Only Ledger

Posted journal entries are immutable. The ledger grows in one direction only — forward. Corrections are always a reversal of the original entry followed by a new correct entry. This means the ledger is a complete record of everything that was ever believed to be true, including mistakes and their corrections.

```sql
-- Trigger prevents editing posted entries (except marking as reversed)
CREATE TRIGGER trg_journal_immutability
    BEFORE UPDATE ON journal_entries
    FOR EACH ROW EXECUTE FUNCTION fn_protect_posted_journal();
```

**Why this matters:** Any accounting system that allows in-place editing of posted entries cannot be audited with confidence. The append-only constraint is also what makes point-in-time reporting reliable — you can always reconstruct what the books looked like on any historical date.

### 5.3 Period Gating

All posting goes through a period gate enforced at two layers: the application's `ValidatePeriod` pipeline stage, and a PostgreSQL trigger on `journal_entries`. An open period accepts entries from any authorised user. A soft-closed period accepts entries only from `finance_manager` or `cfo` roles. A hard-closed or locked period accepts no entries whatsoever.

**Why two layers?** The application gate provides a good error message. The database trigger is the safety net for any code path — API, background job, data migration script — that bypasses the application layer.

### 5.4 Event Consumer, Not Event Producer

The Finance module consumes events from HR, Sales, Inventory, Pricing, and Currency modules. It does not read from their tables. When it has posted a journal entry, it emits a confirmation event back.

```
HR module           →  PayrollPosted event         →  Finance module
Sales module        →  SaleInvoicePosted event      →  Finance module
Finance module      →  JournalPosted event          →  HR, Sales (confirmation)
```

**Why this matters:** If the Finance module queried HR's payroll tables directly, a schema change in HR would break Finance. The event boundary decouples the modules — Finance only needs to understand the event payload contract, not the internal structure of any other module.

### 5.5 Subsidiary Ledger Pattern

The GL operates at control account level. AR, AP, inventory, and employee receivables are subsidiary ledgers that roll up to GL control accounts. The sub-ledger holds customer-level or supplier-level detail; the GL holds the total.

**Practical implication:** You can see that Accounts Receivable is 3,450,000 KES on the balance sheet, and you can also see exactly which customers owe what and how old each balance is. Reconciling the two is a first-class, explicitly supported operation — not a manual check.

### 5.6 Multi-Entity by Design

From the first line of code, the module supports multiple legal entities (Organisations) within a single tenant. Each Organisation has its own chart of accounts, periods, and GL. Consolidation across entities is a native operation, not a bolt-on.

**Why from day one?** Adding multi-entity support to a single-entity ledger requires rewriting the core data model. Doing it from the start, even if only one entity uses it initially, costs a small amount of schema complexity and saves an eventual full rewrite.

---

## 6. Scope Boundary

The Finance module owns the ledger. It does not own the business processes that generate ledger entries. This boundary is strict.

| In Scope | Out of Scope | Owned By |
|---|---|---|
| Chart of accounts management | Currency master and exchange rate feeds | Currency module |
| Journal entry lifecycle | Payroll computation | HR module |
| General Ledger and subsidiary ledgers | Invoice generation and sales order management | Sales module |
| Period and fiscal year management | Purchase order management | Procurement module |
| Bank reconciliation engine | Asset depreciation computation | Fixed Assets module |
| AR and AP aging | Tax computation and tax return filing | Tax module |
| Cost centre and budget management | Subscription billing | Billing module |
| Cash position and 13-week forecasting | FX hedge contract management | Currency module |
| Intercompany accounting and eliminations | Forecourt wetstock reconciliation | Forecourt module |
| Financial statement generation | Payroll slip generation | HR module |
| Payment run management | — | — |
| Petty cash management | — | — |

**Note on the Forecourt module:** The Forecourt module will emit events (`ShiftClosed`, `WetStockVarianceConfirmed`, `DipReadingPosted`) that the Finance module consumes and translates into journal entries using account mappings. The specific business logic of how a dip reading becomes a stock variance lives in the Forecourt module. The Finance module only cares about the resulting debit and credit instruction.

---

## 7. Domain Model

### 7.1 Aggregate Roots

These are the top-level entities that the Finance module owns and manages. Each has its own lifecycle, identity, and invariants.

| Aggregate | Description |
|---|---|
| `Organisation` | A legal entity within a tenant. Owns its own COA, periods, and GL. |
| `Account` | A node in the chart of accounts hierarchy. Has a root type, account type, and normal balance. |
| `FiscalYear` | A 12-month reporting period with sub-periods. Once locked, immutable. |
| `AccountingPeriod` | A sub-period within a fiscal year (monthly or quarterly). Status-gated. |
| `JournalEntry` | An immutable (once posted) collection of balanced debit/credit lines. |
| `BankReconciliation` | A workspace for matching GL entries to bank statement lines for one period. |
| `Budget` | A versioned collection of account + cost centre + period target amounts. |
| `PaymentRun` | A batch of supplier payments processed and exported together. |
| `PettyCashFund` | An imprest fund with a fixed float amount managed by a custodian. |

### 7.2 Core Entities

These entities exist within aggregates and have meaning only in that context.

| Entity | Belongs To | Description |
|---|---|---|
| `AccountBalance` | Account + Period | Materialised running balance. Updated atomically on every posting. |
| `JournalLine` | JournalEntry | A single debit or credit line. Immutable once the entry is posted. |
| `BankAccount` | Organisation | A bank account linked to a GL cash/bank account. |
| `BankStatement` | BankReconciliation | An imported bank statement for a period. |
| `BankStatementLine` | BankStatement | A single transaction line from a bank statement. |
| `ReconciliationMatch` | BankReconciliation | A confirmed match between statement lines and GL entries. |
| `CostCentre` | Organisation | An organisational unit for cost and revenue tracking. Tree structure. |
| `BudgetLine` | Budget | A single account + cost centre + period budget amount. |
| `SubsidiaryEntry` | AR/AP/Inventory | Customer-, supplier-, or item-level detail that rolls up to a GL control account. |
| `IntercompanyLink` | Organisation pair | A matching pair of IC transactions across two legal entities. |
| `AccountMapping` | Organisation | Maps a semantic code (e.g. `FUEL_REVENUE`) to a specific GL account ID. |

### 7.3 Value Objects

These have no independent identity — they are defined entirely by their attributes.

| Value Object | Description |
|---|---|
| `Money` | An amount combined with an ISO 4217 currency code. Never just a number. |
| `AccountPath` | Materialised path string enabling efficient subtree queries e.g. `/1000/1100/1110` |
| `NormalBalance` | `debit` or `credit` — the side that increases the account |
| `JournalReference` | Auto-generated sequential reference per organisation e.g. `JE-2025-0145` |
| `PeriodStatus` | `open` · `soft_closed` · `hard_closed` · `locked` |
| `EntryStatus` | `draft` · `pending_approval` · `approved` · `posted` · `reversed` · `cancelled` |
| `MatchConfidence` | `auto` · `suggested` · `manual` — confidence level of a reconciliation match |

### 7.4 Key Relationships

```
Tenant
  └── Organisation (N — one per legal entity)
        ├── Account (tree, unlimited depth)
        │     └── AccountBalance (one row per account per period)
        ├── FiscalYear (N)
        │     └── AccountingPeriod (12 per year, monthly default)
        ├── BankAccount (N)
        │     └── BankReconciliation (one per bank account per period)
        │           └── ReconciliationMatch (N per reconciliation)
        ├── CostCentre (tree, unlimited depth)
        ├── Budget (N, versioned — only one ACTIVE per fiscal year)
        │     └── BudgetLine (one per account + cost centre + period)
        └── JournalEntry (N — the core of everything)
              └── JournalLine (minimum 2, always balanced)

JournalEntry
  ├── source_module: "hr" | "sales" | "inventory" | "forecourt" | "pricing" | "currency" | "manual"
  └── source_event_id: UUID (the integration event that triggered this entry, for idempotency)
```

### 7.5 Account Classification

Every account in the system belongs to one of five root types. This classification drives financial statement placement, normal balance determination, and period-close behaviour.

```
Root Type     Normal Balance    Financial Statement
──────────────────────────────────────────────────
asset         debit             Balance Sheet — Assets
liability     credit            Balance Sheet — Liabilities
equity        credit            Balance Sheet — Equity
revenue       credit            Profit & Loss — Revenue
expense       debit             Profit & Loss — Expenses
```

Account types are more granular sub-classifications within each root type. Examples:

| Root Type | Account Types |
|---|---|
| asset | `bank`, `cash`, `receivable`, `inventory`, `fixed_asset`, `accumulated_depreciation`, `prepaid` |
| liability | `payable`, `tax_payable`, `accrued`, `loan` |
| equity | `equity_capital`, `retained_earnings` |
| revenue | `revenue_sales`, `revenue_service`, `other_income` |
| expense | `cogs`, `expense_operating`, `expense_depreciation`, `expense_interest`, `other_expense` |

---

## 8. Module Dependencies

### 8.1 What the Finance Module Depends On

| Dependency | Why | What Happens If Unavailable |
|---|---|---|
| **Currency module** | Exchange rates for FC transactions; FX revaluation events | FC postings queue; revaluation deferred until rates available |
| **PostgreSQL** | Primary data store; trigger-enforced constraints | Module non-functional — no alternative |
| **Temporal** | Period close workflows; scheduled reconciliation; IC matching cron | Workflows must be triggered manually via API; no functional data loss |
| **Outbox table** | Event emission to other modules | Events queue; eventually consistent delivery guaranteed on reconnect |

### 8.2 What Depends on the Finance Module

| Dependent Module | What It Needs From Finance | Impact If Finance Is Down |
|---|---|---|
| **HR / Payroll** | Confirmation that payroll journal was posted (`JournalPosted` event) | Payroll slips can be issued; GL confirmation delayed |
| **Sales** | AR balance per customer (for credit limit checks); invoice posting confirmation | Credit checks fall back to cached balance; invoicing continues |
| **Procurement** | AP balance per supplier; payment run confirmation | PO approval continues; payment scheduling delayed |
| **Forecourt** | Shift reconciliation journal confirmation | Shift can close operationally; GL posting queued |
| **Reporting / BI** | All financial views on the read replica | Report execution fails or returns stale data |
| **Audit** | Append-only audit trail of all journal entries | No direct impact on operations |

### 8.3 Integration Event Contracts

These are the inbound events the Finance module consumes. The contract is the event payload schema — Finance does not care about the internal data model of the emitting module.

| Source Module | Event | Finance Action |
|---|---|---|
| HR | `PayrollPosted` | Build and post salary/statutory journal via account mapping |
| HR | `PayrollReversed` | Post mirror reversal journal |
| HR | `DiscrepancyConfirmed` | Post DR Employee Receivable, CR Cash Shortage Expense |
| Sales | `SaleInvoicePosted` | Post DR AR, CR Revenue, CR VAT Output |
| Sales | `PaymentReceived` | Post DR Bank, CR AR |
| Sales | `CreditNotePosted` | Post reversal of original sale journal |
| Procurement | `PurchaseInvoicePosted` | Post DR Inventory/Expense + DR VAT Input, CR AP |
| Procurement | `SupplierPaymentPosted` | Post DR AP, CR Bank |
| Inventory | `StockAdjustmentPosted` | Post DR/CR Inventory, CR/DR Stock Variance |
| Forecourt | `ShiftClosed` | Post fuel sales, cash handling, and variance journals |
| Forecourt | `WetStockVarianceConfirmed` | Post inventory shrinkage or gain journal |
| Currency | `FXRevaluationPosted` | Post unrealised FX gain/loss journal |
| Currency | `FXGainLossRealised` | Post realised FX gain/loss on settlement |

---

## 9. Feature Flags & Configuration

The Finance module is designed to be deployable in a minimal configuration for a single-entity, single-currency, KES-only business — and progressively enabled for more complex operations.

All flags are stored per-tenant in `tenant_config.finance`. They can be changed at runtime without a deployment. Some flags are irreversible after data exists (marked below).

### 9.1 Core Flags

| Flag | Type | Default | Description | Irreversible After Data? |
|---|---|---|---|---|
| `default_currency` | string | `KES` | Tenant functional currency | **Yes** — changing after first posting is not supported |
| `fiscal_year_start_month` | int (1–12) | `1` | Month the fiscal year begins | Yes — affects all period calculations |
| `journal_reference_prefix` | string | `JE` | Prefix for auto-generated journal references e.g. `JE-2025-0145` | No |
| `journal_reference_sequence_length` | int | `4` | Zero-padded digits: `JE-2025-0001` vs `JE-2025-00001` | No |

### 9.2 Control Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `budget_control_mode` | enum | `soft` | `none` · `soft` (warn and allow override) · `hard` (block) |
| `budget_override_requires_reason` | bool | `true` | Soft budget override must include a comment |
| `require_cost_centre_on_expense` | bool | `false` | All expense account lines must carry a cost centre |
| `require_approval_above_amount` | decimal | `null` | Journal entries above this amount require approval regardless of type |
| `duplicate_invoice_check` | bool | `true` | Warn on same supplier + invoice number + amount within 30 days |
| `auto_reverse_accruals` | bool | `true` | Month-end accrual entries automatically reversed on the first day of the next period |
| `outstanding_cheque_alert_days` | int | `60` | Flag outstanding cheques older than N days in reconciliation |

### 9.3 Feature Enablement Flags

These turn entire sub-features on or off. A flag set to `false` hides the feature from the UI and disables its API endpoints.

| Flag | Type | Default | Description | v1.0? |
|---|---|---|---|---|
| `multi_entity_enabled` | bool | `false` | Enable multiple Organisations per tenant | Optional |
| `intercompany_enabled` | bool | `false` | Enable IC transactions and eliminations. Requires `multi_entity_enabled` | Optional |
| `petty_cash_enabled` | bool | `true` | Enable petty cash imprest sub-module | Yes |
| `bank_reconciliation_enabled` | bool | `true` | Enable bank reconciliation workspace | Yes |
| `budget_module_enabled` | bool | `true` | Enable budget creation and budget vs. actual reporting | Yes |
| `cash_forecast_enabled` | bool | `false` | Enable 13-week rolling cash forecast engine | Optional |
| `payment_run_enabled` | bool | `true` | Enable batch supplier payment runs | Yes |
| `cost_centre_enabled` | bool | `true` | Enable cost centre tracking and P&L | Yes |

### 9.4 Bank Statement Import Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `bank_statement_import_formats` | string[] | `["ofx","mt940","csv"]` | Enabled import formats |
| `auto_match_confidence_threshold` | enum | `auto` | Minimum confidence level to auto-confirm a match without user review |
| `bank_rec_required_for_close` | bool | `true` | Period cannot hard-close without an APPROVED reconciliation for every active bank account |

### 9.5 Compliance Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `subsidiary_reconcile_required_for_close` | bool | `true` | AR/AP subsidiary must reconcile to GL control before hard close |
| `ic_matching_required_for_close` | bool | `true` | Period cannot hard-close if IC balances are unmatched. Requires `intercompany_enabled` |
| `payment_file_format` | enum | `generic_csv` | Default bank payment export format: `equity_eft` · `kcb_rtgs` · `swift_mt101` · `generic_csv` |
| `fx_revaluation_required_for_close` | bool | `true` | Period cannot hard-close if unrevalued FC balances exist |

### 9.6 Configuration Change Governance

Changing any flag that affects financial integrity (e.g. `default_currency`, `fiscal_year_start_month`, `budget_control_mode` from `hard` to `none`) is recorded in the audit log with the before value, after value, changed-by user, and timestamp. These changes are not reversible via the UI for flags marked irreversible above — they require a support escalation and a data migration assessment.

---

## 10. v1.0 Rollout Assessment

This section identifies which capabilities are essential for the first production rollout versus which can be deferred to later releases without blocking go-live.

### Must Have at v1.0

These are non-negotiable because without them the system cannot produce reliable books.

| Capability | Reason |
|---|---|
| Chart of accounts (tree, materialised path) | Every other feature depends on it |
| Journal entry pipeline (all 10 stages) | Core of the ledger |
| Account balances materialisation | Required for any report to be performant |
| Period management (open/soft close/hard close) | Required for month-end integrity |
| Period gate DB trigger | Required for data integrity |
| AR sub-ledger and aging | Required for cash collection management |
| AP sub-ledger and aging | Required for supplier payment management |
| Bank reconciliation | Required for any credible audit trail |
| Payment runs | Required for efficient supplier payments |
| Trial balance, balance sheet, P&L reports | Required for basic financial visibility |
| GL detail (account ledger) report | Required for audit and investigation |
| Role-based permissions (all roles) | Required for SOD compliance |
| Account mapping (semantic code → GL account) | Required for integration events from other modules |
| Single-entity, single-currency baseline | Core deployment model |

### Can Be Deferred Post-v1.0

These add significant value but do not block go-live for a single-entity operator.

| Capability | Suggested Release | Reason It Can Wait |
|---|---|---|
| Multi-entity / intercompany | v1.2 | Single-entity covers the initial target market |
| 13-week cash flow forecast | v1.1 | Cash position report (v1.0) covers immediate need |
| Auto-match bank reconciliation | v1.1 | Manual matching works at low volume |
| Budget hard-stop control | v1.1 | Soft control (warn only) is sufficient initially |
| MT940 and OFX import | v1.1 | CSV import covers most Kenyan banks initially |
| Petty cash imprest module | v1.1 | Journal entries cover petty cash manually until then |
| Cost centre distributed allocation | v1.1 | Direct cost centre tagging is sufficient at launch |
| Temporal workflow orchestration | v1.1 | Period close via API calls is acceptable initially |
| Report scheduling and email delivery | v1.1 | On-demand report execution covers v1.0 |

### Explicit Non-Starter Deferrals

The following must **not** be deferred, even under schedule pressure:

- DB-level immutability triggers on `journal_entries` and `journal_lines`
- RLS policies on all finance tables
- Period gate trigger
- Double-entry constraint on `journal_lines`
- Idempotency on integration event consumers

Deferring any of these creates a ledger that cannot be trusted, and rebuilding trust in a corrupted ledger is far more expensive than the time saved by skipping the constraint.

---

*End of FILE-01. Continue with FILE-02: Chart of Accounts Engine.*
