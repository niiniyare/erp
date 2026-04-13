# Awo ERP — Finance Module
## FILE-02: Chart of Accounts Engine

**Document Version:** 2.0.0  
**Series:** FILE-02 of 10  
**Depends On:** FILE-01 (Domain Model, Organisation aggregate)  
**Depended On By:** FILE-03 (Journal Pipeline), FILE-04 (Period Management), FILE-05 (AR/AP), FILE-06 (Budgets), FILE-08 (Reports)

---

## Table of Contents

1. [Why the Chart of Accounts Exists](#1-why-the-chart-of-accounts-exists)
2. [What You Lose Without a Proper COA Engine](#2-what-you-lose-without-a-proper-coa-engine)
3. [How ERPNext, NetSuite and QuickBooks Handle COA](#3-how-erpnext-netsuite-and-quickbooks-handle-coa)
4. [Account Hierarchy & Materialised Path](#4-account-hierarchy--materialised-path)
5. [Account Types Reference](#5-account-types-reference)
6. [Account Balance Materialisation](#6-account-balance-materialisation)
7. [Business Rules & Validation](#7-business-rules--validation)
8. [Account Mapping (Semantic Codes)](#8-account-mapping-semantic-codes)
9. [COA Import & Bulk Operations](#9-coa-import--bulk-operations)
10. [Performance & Storage](#10-performance--storage)
11. [Database Schema](#11-database-schema)
12. [API Reference](#12-api-reference)
13. [Feature Flags & Configuration](#13-feature-flags--configuration)
14. [v1.0 Rollout Assessment](#14-v10-rollout-assessment)

---

## 1. Why the Chart of Accounts Exists

The Chart of Accounts (COA) is the filing system of the entire Finance module. Every journal entry, every balance query, every financial statement, every budget line, every cost centre report — all of them reference an account from this tree. The COA defines what things *are* financially: whether a balance is an asset or an expense, whether it belongs on the balance sheet or the income statement, whether it increases with a debit or a credit.

Without a well-structured COA, you have transactions but no meaning. With one, every monetary event in the system is automatically classified, aggregated, and reported in the right place.

The COA engine in Awo is not just a list of accounts. It is a tree with materialised paths for efficient hierarchy queries, a balance materialisation layer that avoids full-table scans, a type system that drives financial statement construction, and a set of posting rules that control which accounts can receive which kinds of entries.

---

## 2. What You Lose Without a Proper COA Engine

| Scenario | Without COA Engine | With COA Engine |
|---|---|---|
| Financial statement generation | Manual aggregation in a spreadsheet | Automatic, driven by `root_type` and `report_section` |
| Subtree balance queries (e.g. "all Current Assets") | Recursive CTE or application-side loop — slow at scale | Single `LIKE '/1000/1100/%'` index scan |
| Preventing postings to group accounts | Application-only check, easily bypassed | DB constraint via `is_group` + trigger |
| Reactivating a closed account | No record of whether balance is zero | `hasOpenBalance` check on deactivation |
| Adding a new revenue stream | Create account, add to a spreadsheet mapping | Create account under correct parent — reports update automatically |
| Multi-currency account isolation | Manual tracking | `locked_currency` field enforced at posting time |

---

## 3. How ERPNext, NetSuite and QuickBooks Handle COA

### ERPNext

ERPNext uses a tree-based COA stored in the `Account` DocType. Accounts have a `root_type` (Asset, Liability, Equity, Income, Expense) and an `account_type` sub-classification. The hierarchy is unlimited depth.

**Balance computation:** ERPNext does *not* materialise balances. When you run a trial balance or balance sheet, it executes a `SUM` over the `GL Entry` table filtered by account and date range. This works acceptably for small ledgers (under ~200k GL entries) but degrades significantly at scale. A tenant with 3 years of daily fuel sales across 10 pumps can accumulate 2–3 million GL entries, at which point a balance sheet render can take 30–60 seconds.

**COA templates:** ERPNext ships with country-specific COA templates. The Kenya template is available but was last updated in 2021 and does not reflect the current KRA chart of accounts structure.

**What Awo improves:** Awo materialises balances in `account_balances` (one row per account per period), updated atomically inside every posting transaction. Balance queries are a primary-key lookup rather than an aggregate scan.

### NetSuite

NetSuite's COA supports segments (account + subsidiary + department + class + location) which together form the full account string. An "account" in NetSuite is really a combination of these dimensions. This is powerful for large enterprises but adds significant setup complexity.

**Balance computation:** NetSuite materialises balances per period, similar to Awo's approach. Reports are fast.

**What Awo adopts:** The concept of cost centres as a separate dimension from the account code (analogous to NetSuite's Department segment). Awo deliberately keeps cost centres out of the account code itself — the account code stays clean and short; the cost centre is a separate field on the journal line.

### QuickBooks

QuickBooks uses a flat-ish account list with a single level of sub-accounts. The hierarchy is limited to two levels (Account → Sub-Account) in most versions. There is no concept of a group account that aggregates children — the hierarchy is visual only.

**Balance computation:** Materialised per account. Fast for the transaction volumes QuickBooks handles.

**What Awo improves:** Awo supports unlimited depth, true group accounts with aggregated balances, and materialised paths — giving the reporting flexibility of NetSuite's segments without the configuration overhead.

---

## 4. Account Hierarchy & Materialised Path

### Tree Structure

The COA is an unlimited-depth tree. Every account has a `parent_id` pointing to its parent account (or `NULL` for root accounts). The tree has two kinds of nodes:

- **Group accounts** (`is_group = true`): Cannot receive journal entries directly. Their balance is the sum of all descendant leaf accounts. They exist purely for hierarchy and reporting.
- **Leaf accounts** (`is_group = false`): The only accounts that can receive journal entries.

```
1000 Assets                         ← Group (root)
├── 1100 Current Assets             ← Group
│   ├── 1110 Cash & Equivalents     ← Group
│   │   ├── 1111 Petty Cash         ← Leaf (postable)
│   │   ├── 1112 Checking – Main    ← Leaf (postable)
│   │   └── 1113 Savings Account    ← Leaf (postable)
│   └── 1120 Accounts Receivable    ← Group
│       ├── 1121 Trade Receivables  ← Leaf (postable)
│       └── 1122 Staff Advances     ← Leaf (postable)
└── 1200 Non-Current Assets         ← Group
    └── 1210 Equipment              ← Leaf (postable)
```

### Materialised Path

Rather than traversing parent-child links at query time, every account stores its full path from the root as a string. This is the materialised path pattern.

```
Account: 1112 Checking – Main
Path:    /1000/1100/1110/1112
Level:   4
Parent:  1110
```

This single field enables queries that would otherwise require recursive CTEs:

```sql
-- All leaf accounts under Current Assets
SELECT * FROM accounts
WHERE path LIKE '/1000/1100/%'
  AND is_group = FALSE
  AND tenant_id = current_setting('app.current_tenant')::UUID;

-- Subtree balance roll-up for a group account (any depth)
SELECT
    SUM(ab.period_debit) AS total_debit,
    SUM(ab.period_credit) AS total_credit
FROM account_balances ab
JOIN accounts a ON a.id = ab.account_id
WHERE a.path LIKE '/1000/1100/%'
  AND ab.period_id = $1
  AND a.tenant_id = current_setting('app.current_tenant')::UUID;
```

The `path` column is indexed with `text_pattern_ops` to make `LIKE '/prefix/%'` queries use the B-tree index efficiently.

**Path maintenance:** When an account is moved to a new parent (only allowed before any transactions exist), the path of the account and all its descendants must be updated. This is done in a single recursive UPDATE:

```sql
UPDATE accounts
SET path = $new_parent_path || '/' || code || substr(path, length($old_path) + 1)
WHERE path LIKE $old_path || '/%'
  AND tenant_id = $tenant_id;
```

### Account Domain Model

```go
type Account struct {
    ID                  uuid.UUID
    TenantID            uuid.UUID
    OrganisationID      uuid.UUID
    Code                string       // immutable once transactions exist
    Name                string
    RootType            RootType     // asset | liability | equity | revenue | expense
    AccountType         AccountType  // bank | cash | receivable | payable | ... (see §5)
    NormalBalance       NormalBalance // debit | credit
    ParentID            *uuid.UUID
    Path                string       // e.g. /1000/1100/1110/1112
    Level               int
    IsGroup             bool
    IsActive            bool
    IsSystem            bool         // protected from deletion
    AllowManualEntries  bool
    RequireReference    bool
    RequireCostCentre   bool
    LockedCurrency      *string      // e.g. "USD" for a USD-only bank account
    ReportSection       string       // balance_sheet | profit_loss | cash_flow
    CreatedAt           time.Time
    UpdatedAt           time.Time
}

func (a *Account) Deactivate(hasOpenBalance bool) error {
    if a.IsGroup {
        return ErrGroupAccountMustDeactivateChildrenFirst
    }
    if hasOpenBalance {
        return ErrCannotDeactivateWithOpenBalance
    }
    if a.IsSystem {
        return ErrSystemAccountCannotBeDeactivated
    }
    a.IsActive = false
    return nil
}

func (a *Account) CanReceivePosting(entryType EntryType) error {
    if a.IsGroup {
        return ErrGroupAccountCannotPost
    }
    if !a.IsActive {
        return ErrInactiveAccountCannotPost
    }
    if entryType == EntryTypeManual && !a.AllowManualEntries {
        return ErrManualEntriesNotAllowed{Account: a.Code}
    }
    return nil
}
```

---

## 5. Account Types Reference

Account type drives automatic normal balance assignment, financial statement section placement, and sub-ledger linkage.

### Asset Types

| Type | Normal Balance | Used For | Sub-Ledger Link |
|---|---|---|---|
| `bank` | Debit | Bank accounts — linked to BankAccount entity | Bank reconciliation |
| `cash` | Debit | Cash on hand, petty cash | Petty cash fund |
| `receivable` | Debit | Accounts receivable control account | AR sub-ledger |
| `inventory` | Debit | Goods held for sale or production | Inventory sub-ledger |
| `fixed_asset` | Debit | Property, plant and equipment at cost | Fixed assets register |
| `accumulated_depreciation` | Credit | Contra-asset reducing fixed asset book value | Fixed assets register |
| `prepaid` | Debit | Expenses paid in advance | None |
| `tax_receivable` | Debit | VAT input, withholding tax recoverable | None |
| `other_current_asset` | Debit | Any current asset not classified above | None |
| `investment` | Debit | Long-term investments, shares held | None |

### Liability Types

| Type | Normal Balance | Used For | Sub-Ledger Link |
|---|---|---|---|
| `payable` | Credit | Accounts payable control account | AP sub-ledger |
| `tax_payable` | Credit | VAT output, PAYE, NSSF, NHIF payable | None |
| `accrued` | Credit | Accrued expenses not yet invoiced | None |
| `loan` | Credit | Bank loans, director loans | Loan schedule |
| `deferred_revenue` | Credit | Revenue received before delivery | None |
| `other_current_liability` | Credit | Any current liability not classified above | None |

### Equity Types

| Type | Normal Balance | Used For |
|---|---|---|
| `equity_capital` | Credit | Paid-up share capital |
| `retained_earnings` | Credit | Accumulated profit/loss from prior periods |
| `current_year_earnings` | Credit | Net income for the current fiscal year (auto-calculated) |

### Revenue Types

| Type | Normal Balance | Used For |
|---|---|---|
| `revenue_sales` | Credit | Product and fuel sales revenue |
| `revenue_service` | Credit | Service income |
| `other_income` | Credit | Interest income, rental income, forex gain |

### Expense Types

| Type | Normal Balance | Used For |
|---|---|---|
| `cogs` | Debit | Cost of goods sold, fuel cost |
| `expense_operating` | Debit | Salaries, rent, utilities, marketing |
| `expense_depreciation` | Debit | Depreciation and amortisation charges |
| `expense_interest` | Debit | Interest on loans and finance leases |
| `other_expense` | Debit | Forex loss, write-offs, one-off charges |

---

## 6. Account Balance Materialisation

### Why Materialise

Every balance query — balance sheet, trial balance, AR aging total, budget variance — needs the closing balance of one or more accounts. If balances are computed on the fly by summing `journal_lines`, query time grows linearly with transaction history. A busy petroleum retail operator posting 1,000 journal lines per day accumulates 365,000 lines per year. After 3 years that is over 1 million rows — a `SUM` across all of them for a single account balance takes hundreds of milliseconds, and a full trial balance across 200 accounts takes seconds.

Awo instead maintains one row per account per period in `account_balances`, updated atomically inside the same database transaction that inserts the journal lines. Balance queries become a primary-key lookup (microseconds) regardless of how many years of history exist.

### Balance Structure

```go
type AccountBalance struct {
    ID            uuid.UUID
    TenantID      uuid.UUID
    AccountID     uuid.UUID
    PeriodID      uuid.UUID
    OpeningDebit  decimal.Decimal  // carried forward from prior period closing
    OpeningCredit decimal.Decimal
    PeriodDebit   decimal.Decimal  // accumulated during this period
    PeriodCredit  decimal.Decimal
    // ClosingDebit and ClosingCredit are GENERATED ALWAYS AS stored columns in PostgreSQL.
    // NetBalance is also generated: ClosingDebit - ClosingCredit.
}
```

**Note on generated columns:** PostgreSQL generated columns are computed by the database engine and cannot be written to directly. The application writes only `opening_debit`, `opening_credit`, `period_debit`, `period_credit`. The closing and net values are always consistent with those inputs, with no risk of a bug writing an inconsistent pre-computed value.

### Upsert on Posting

When a journal entry is posted, every line triggers an upsert into `account_balances`. This is done inside the same transaction as the `journal_lines` insert:

```sql
INSERT INTO account_balances (
    tenant_id, account_id, period_id, opening_debit, opening_credit,
    period_debit, period_credit
)
VALUES ($1, $2, $3, 0, 0, $4, $5)
ON CONFLICT (tenant_id, account_id, period_id)
DO UPDATE SET
    period_debit  = account_balances.period_debit  + EXCLUDED.period_debit,
    period_credit = account_balances.period_credit + EXCLUDED.period_credit,
    updated_at    = NOW();
```

Because `closing_debit`, `closing_credit`, and `net_balance` are PostgreSQL `GENERATED ALWAYS AS STORED` columns, they are recomputed automatically by the database whenever `opening_*` or `period_*` change. No application code needs to compute or write them.

### Period Opening Balance Propagation

When a new accounting period is opened, the system creates `account_balances` rows for every account that had a non-zero closing balance in the prior period, setting `opening_debit` and `opening_credit` to the prior period's closing values. This is done by the `PeriodService.OpenNewPeriod` application service, not by a trigger — it is an explicit, auditable operation:

```go
func (s *PeriodService) PropagateOpeningBalances(
    ctx      context.Context,
    orgID    uuid.UUID,
    newPeriodID uuid.UUID,
    priorPeriodID uuid.UUID,
) error {
    return s.db.WithTransaction(ctx, func(tx pgx.Tx) error {
        return s.balanceRepo.CopyClosingToOpening(ctx, tx, orgID, priorPeriodID, newPeriodID)
    })
}
```

For revenue and expense accounts, opening balances are always zero — only balance sheet accounts (asset, liability, equity) carry forward. This is enforced in `CopyClosingToOpening` by filtering on `root_type IN ('asset', 'liability', 'equity')`.

---

## 7. Business Rules & Validation

### Account Creation Rules

| Rule | Description | Enforcement |
|---|---|---|
| `FIN-ACT-001` | Account code must be unique within an organisation | DB unique constraint on `(tenant_id, organisation_id, code)` |
| `FIN-ACT-002` | Account code is immutable once any transaction exists against it | Application check + error on update attempt |
| `FIN-ACT-003` | Root type is immutable once any transaction exists | Application check |
| `FIN-ACT-004` | Parent account must belong to the same organisation | FK + application check |
| `FIN-ACT-005` | Account hierarchy must not create cycles | Application check on parent assignment |
| `FIN-ACT-006` | A group account cannot be a leaf account (mutually exclusive) | Application enforced |
| `FIN-ACT-007` | `NormalBalance` must be consistent with `RootType` | Application enforced on creation; asset/expense = debit, liability/equity/revenue = credit |

### Account Deactivation Rules

| Rule | Description | Enforcement |
|---|---|---|
| `FIN-ACT-010` | An account with a non-zero balance in the current or any open period cannot be deactivated | `hasOpenBalance` check in `Deactivate()` |
| `FIN-ACT-011` | A group account cannot be deactivated while it has active child accounts | Application check |
| `FIN-ACT-012` | A system account (`is_system = true`) cannot be deactivated or deleted | Hard block |
| `FIN-ACT-013` | Deactivating an account does not delete its history — all prior postings remain intact | Soft deactivation only |

### Posting Rules

| Rule | Description | Enforcement |
|---|---|---|
| `FIN-ACT-020` | Only leaf accounts (`is_group = false`) can receive journal entries | `ValidateAccounts` pipeline stage |
| `FIN-ACT-021` | Only active accounts can receive journal entries | `ValidateAccounts` pipeline stage |
| `FIN-ACT-022` | Accounts with `allow_manual_entries = false` cannot appear in MANUAL-type entries | `ValidateAccounts` pipeline stage |
| `FIN-ACT-023` | Accounts with `require_reference = true` must have a non-empty reference on the journal line | `ValidateReferences` pipeline stage |
| `FIN-ACT-024` | Accounts with `locked_currency` set must only be used in entries with that currency | `ValidateAccounts` pipeline stage |
| `FIN-ACT-025` | Integration entries (source_module ≠ "manual") bypass `allow_manual_entries` restriction | Enforced by entry type in the pipeline |

### Alternatives Considered

**Alternative: Store balances as a single `net_balance` column.**
Rejected. Storing gross debits and gross credits separately allows the system to produce a proper trial balance (which shows both sides), detect unusual credit balances on asset accounts, and compute turnover ratios without re-scanning journal lines.

**Alternative: Compute balances using a database view over journal_lines.**
Rejected for the same performance reason described in §6. A view is a query — it runs every time it is referenced.

**Alternative: Use a closing entry approach (zero out revenue/expense at year-end).**
Accepted as an option, not a requirement. `auto_year_end_closing` flag (see §13) controls whether the system automatically generates closing entries at fiscal year end. For most SME customers this is not needed — the system calculates retained earnings as `opening_retained_earnings + net_income_for_period`.

---

## 8. Account Mapping (Semantic Codes)

### Why Account Mapping Exists

Other modules (HR, Sales, Inventory, Forecourt) do not know which GL account codes a specific tenant uses. A payroll system should not be hardcoded to account `7110`. Different organisations within the same tenant may use different account codes. The account mapping table solves this by introducing a layer of semantic codes — stable identifiers that mean the same thing regardless of which GL account code the tenant assigned to them.

```
HR module emits:  "BASIC_SALARY"  →  AccountMapping  →  Account 7110 (or 7120, or whatever this tenant uses)
```

### Mapping Structure

```go
type AccountMapping struct {
    ID             uuid.UUID
    TenantID       uuid.UUID
    OrganisationID uuid.UUID
    SemanticCode   string     // stable identifier used by source modules
    AccountID      uuid.UUID  // the actual GL account for this tenant/org
    CostCentreID   *uuid.UUID // optional default cost centre
    EffectiveFrom  time.Time
    EffectiveTo    *time.Time // nil = currently active
}
```

Mappings are date-effective: an organisation can change which GL account a semantic code maps to without affecting historical postings.

### Standard Semantic Codes

The following codes are recognised by the system's event consumers. Organisations must configure a mapping for every code that a module they use will emit. Missing mappings cause integration events to be held in a dead-letter queue until resolved.

**HR / Payroll**

| Semantic Code | Meaning |
|---|---|
| `BASIC_SALARY` | Gross basic salary expense |
| `ALLOWANCES` | Non-basic allowances (transport, housing) |
| `OVERTIME` | Overtime pay |
| `EMPLOYER_NSSF` | Employer NSSF contribution (expense) |
| `EMPLOYEE_NSSF` | Employee NSSF deduction (liability) |
| `EMPLOYER_NHIF` | Employer NHIF contribution (expense) |
| `EMPLOYEE_NHIF` | Employee NHIF deduction (liability) |
| `PAYE_PAYABLE` | PAYE withholding liability |
| `SALARY_PAYABLE` | Net salaries payable to employees |
| `EMPLOYEE_RECEIVABLE` | Employee overpayment or advance |

**Sales**

| Semantic Code | Meaning |
|---|---|
| `FUEL_REVENUE` | Revenue from fuel sales |
| `LPO_REVENUE` | Revenue from local purchase order (fleet/credit) sales |
| `LUBRICANTS_REVENUE` | Revenue from lubricants and accessories |
| `VAT_OUTPUT` | VAT output on sales |
| `AR_CONTROL` | Accounts receivable control account |

**Forecourt**

| Semantic Code | Meaning |
|---|---|
| `FUEL_COGS` | Cost of fuel sold |
| `WET_STOCK_VARIANCE` | Wet stock gain or loss |
| `SHIFT_CASH_SHORTAGE` | Cash shortage on shift reconciliation |
| `SHIFT_CASH_OVERAGE` | Cash overage on shift reconciliation |
| `METER_READING_SUSPENSE` | Temporary suspense for unreconciled meter readings |

**Procurement**

| Semantic Code | Meaning |
|---|---|
| `AP_CONTROL` | Accounts payable control account |
| `VAT_INPUT` | VAT input on purchases |
| `INVENTORY_CONTROL` | Inventory control account |
| `GOODS_RECEIPT_CLEARING` | GR/IR clearing account |

---

## 9. COA Import & Bulk Operations

### Bulk Import Format (CSV)

Organisations can import their existing COA from a CSV file rather than creating accounts one by one. The import validates the entire file before committing any rows.

```csv
code,name,root_type,account_type,parent_code,is_group,allow_manual_entries,require_reference,locked_currency
1000,Assets,asset,root,,true,false,false,
1100,Current Assets,asset,group,1000,true,false,false,
1110,Cash & Cash Equivalents,asset,group,1100,true,false,false,
1111,Petty Cash,asset,cash,1110,false,true,false,KES
1112,Checking Account – Main,asset,bank,1110,false,true,false,KES
1120,Accounts Receivable,asset,receivable,1100,false,false,false,
```

**Import validation rules:**
- All parent codes must either exist in the database already or appear earlier in the same file
- Codes must be unique within the organisation
- `root_type` of a child must match the `root_type` of its parent (or be compatible)
- Circular parent references are detected before any row is committed
- The import is transactional — all rows succeed or none are committed

### Industry-Specific COA Templates

Awo ships with pre-built COA templates suited to its target markets. A template is applied once at organisation setup time and then becomes a normal editable COA.

| Template | Suited For |
|---|---|
| `kenya_petroleum_retail` | Petroleum retail stations (forecourt, wet stock, fleet cards) |
| `kenya_sme_general` | General SME with trading operations |
| `kenya_services` | Service businesses (consulting, logistics) |
| `east_africa_manufacturing` | Manufacturing and processing operations |

The `kenya_petroleum_retail` template includes pre-configured semantic code mappings for all Forecourt module events.

---

## 10. Performance & Storage

### Query Performance Targets

| Query Type | Expected Volume | Target P95 | Mechanism |
|---|---|---|---|
| Single account balance | 1 per UI request | < 2ms | Primary key lookup on `account_balances` |
| Trial balance (all accounts) | ~200 accounts | < 50ms | Index scan on `(tenant_id, period_id)` in `account_balances` |
| Subtree balance (e.g. all Current Assets) | ~50 accounts | < 10ms | `LIKE` on indexed `path` column |
| Account ledger detail (1 year) | ~12,000 lines/account/year | < 200ms | Composite index on `(tenant_id, account_id)` in `journal_lines` |
| COA tree load (full list) | ~300 accounts | < 30ms | Full tenant scan on `accounts` — fits in single page |
| Account search (autocomplete) | Real-time | < 20ms | GIN index on `name` for trigram search |

### Index Strategy

```sql
-- Primary lookup: balance by account and period
CREATE UNIQUE INDEX idx_account_balances_pk
    ON account_balances (tenant_id, account_id, period_id);

-- Subtree queries: all accounts under a path prefix
CREATE INDEX idx_accounts_path
    ON accounts (tenant_id, path text_pattern_ops);

-- Journal line lookup by account (GL detail report)
CREATE INDEX idx_journal_lines_account
    ON journal_lines (tenant_id, account_id, created_at DESC);

-- Account search by name (autocomplete)
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_accounts_name_trgm
    ON accounts USING GIN (name gin_trgm_ops);
```

### Storage Estimates

| Table | Row Size (avg) | Rows per Tenant per Year | Annual Storage |
|---|---|---|---|
| `accounts` | ~400 bytes | 300 (stable, not growing) | ~120 KB (one-time) |
| `account_balances` | ~120 bytes | 300 accounts × 12 periods = 3,600 | ~430 KB/year |
| `account_mappings` | ~200 bytes | ~50 mappings (stable) | ~10 KB (one-time) |

**Key insight:** The `accounts` and `account_balances` tables are small. A tenant with 300 accounts and 5 years of history has under 3 MB of COA-related data. The storage concern in the Finance module is `journal_lines` (covered in FILE-03), not the COA itself.

### Caching Strategy

The full COA tree for a tenant changes rarely (a few times per month at most). It is a strong candidate for application-level caching:

- Cache key: `coa:{tenant_id}:{organisation_id}`
- TTL: 5 minutes
- Invalidation: On any `INSERT`, `UPDATE`, or `DELETE` to `accounts` for the tenant
- Cache store: Redis (or in-process LRU cache for single-node deployments)
- Cache hit rate expected: >99% — most requests read the COA, very few write it

Account balances are **not** cached at the application level because they change on every journal posting. They are kept fast by the materialisation strategy described in §6.

---

## 11. Database Schema

```sql
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
    path                  TEXT        NOT NULL,
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

CREATE INDEX idx_accounts_tenant ON accounts (tenant_id);
CREATE INDEX idx_accounts_path   ON accounts (tenant_id, path text_pattern_ops);
CREATE INDEX idx_accounts_parent ON accounts (tenant_id, parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX idx_accounts_name_trgm ON accounts USING GIN (name gin_trgm_ops);

ALTER TABLE accounts ENABLE ROW LEVEL SECURITY;
ALTER TABLE accounts FORCE  ROW LEVEL SECURITY;

CREATE POLICY accounts_app ON accounts FOR ALL TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY accounts_ro ON accounts FOR SELECT TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);

-- ── account_balances ──────────────────────────────────────────────────────────

CREATE TABLE account_balances (
    tenant_id      UUID          NOT NULL REFERENCES tenants(id),
    id             UUID          NOT NULL DEFAULT gen_random_uuid(),
    updated_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    account_id     UUID          NOT NULL REFERENCES accounts(id),
    period_id      UUID          NOT NULL,
    opening_debit  NUMERIC(18,4) NOT NULL DEFAULT 0,
    opening_credit NUMERIC(18,4) NOT NULL DEFAULT 0,
    period_debit   NUMERIC(18,4) NOT NULL DEFAULT 0,
    period_credit  NUMERIC(18,4) NOT NULL DEFAULT 0,

    -- Generated columns: always consistent, never writable directly
    closing_debit  NUMERIC(18,4) GENERATED ALWAYS AS (opening_debit  + period_debit)  STORED,
    closing_credit NUMERIC(18,4) GENERATED ALWAYS AS (opening_credit + period_credit) STORED,
    net_balance    NUMERIC(18,4) GENERATED ALWAYS AS
                       ((opening_debit + period_debit) - (opening_credit + period_credit)) STORED,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, account_id, period_id)
);

CREATE INDEX idx_account_balances_period  ON account_balances (tenant_id, period_id);
CREATE INDEX idx_account_balances_account ON account_balances (tenant_id, account_id);

ALTER TABLE account_balances ENABLE ROW LEVEL SECURITY;
ALTER TABLE account_balances FORCE  ROW LEVEL SECURITY;

CREATE POLICY ab_app ON account_balances FOR ALL TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY ab_ro ON account_balances FOR SELECT TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);

-- ── account_mappings ──────────────────────────────────────────────────────────

CREATE TABLE account_mappings (
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    id              UUID NOT NULL DEFAULT gen_random_uuid(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    organisation_id UUID NOT NULL,
    semantic_code   TEXT NOT NULL,
    account_id      UUID NOT NULL REFERENCES accounts(id),
    cost_centre_id  UUID,
    effective_from  DATE NOT NULL,
    effective_to    DATE,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, organisation_id, semantic_code, effective_from)
);

CREATE INDEX idx_account_mappings_code
    ON account_mappings (tenant_id, organisation_id, semantic_code);

ALTER TABLE account_mappings ENABLE ROW LEVEL SECURITY;
ALTER TABLE account_mappings FORCE  ROW LEVEL SECURITY;

CREATE POLICY am_app ON account_mappings FOR ALL TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);
```

---

## 12. API Reference

All endpoints require `Authorization: Bearer <token>` and `X-Tenant: <slug>`.

### Accounts

```
GET    /finance/accounts                     List accounts (filterable by root_type, account_type, is_active, is_group, parent_code, q)
POST   /finance/accounts                     Create account
GET    /finance/accounts/{code}              Get account with current balance and recent activity
PATCH  /finance/accounts/{code}              Update mutable fields (name, flags)
POST   /finance/accounts/{code}/deactivate   Deactivate (validates open balance)
GET    /finance/accounts/{code}/ledger       GL detail for date range (?from=&to=&page=&per_page=)
POST   /finance/accounts/import              Bulk COA import from CSV
GET    /finance/accounts/tree                Full COA tree with balances for a period (?period_id=)
```

**Create account request:**
```json
{
  "code": "4310",
  "name": "Consulting Services Revenue",
  "root_type": "revenue",
  "account_type": "revenue_service",
  "parent_code": "4300",
  "allow_manual_entries": true,
  "require_reference": false
}
```

### Account Mappings

```
GET    /finance/account-mappings             List all mappings for the organisation
POST   /finance/account-mappings             Create or update a mapping
DELETE /finance/account-mappings/{id}        Deactivate a mapping (sets effective_to = today)
GET    /finance/account-mappings/validate    Check all required semantic codes are mapped
```

---

## 13. Feature Flags & Configuration

| Flag | Type | Default | Description |
|---|---|---|---|
| `coa_max_depth` | int | `10` | Maximum allowed depth of account hierarchy. Prevents accidental deeply nested trees. |
| `coa_code_format` | string | `null` | Optional regex pattern that account codes must match. e.g. `^\d{4}$` for 4-digit numeric codes only. |
| `allow_account_reparenting` | bool | `false` | Allow moving an account to a different parent after creation (only permitted before transactions exist). |
| `auto_create_balance_sheet_groups` | bool | `true` | When a new organisation is created, auto-generate the 5 root group accounts (Assets, Liabilities, Equity, Revenue, Expenses). |
| `coa_template` | string | `kenya_sme_general` | Which industry COA template to apply on organisation creation. |
| `auto_year_end_closing` | bool | `false` | Whether the system automatically generates closing journal entries to zero revenue/expense accounts at fiscal year end. Most East African SMEs do not use formal closing entries — the system computes retained earnings from net income. |
| `require_account_mapping_for_integration` | bool | `true` | If true, integration events with unmapped semantic codes are held in dead-letter queue rather than failing silently. |
| `coa_import_allow_updates` | bool | `false` | Whether a COA import can update existing accounts (name, flags). If false, only new accounts are created. |

---

## 14. v1.0 Rollout Assessment

### Must Have at v1.0

- Full tree-based COA with materialised path
- `account_balances` materialisation (non-negotiable for performance)
- All posting rules and account validation logic
- Account mapping table and resolution for all v1.0 integration modules
- COA import from CSV (organisations migrating from QuickBooks or ERPNext need this)
- `kenya_petroleum_retail` COA template
- RLS policies on `accounts` and `account_balances`

### Can Be Deferred

| Feature | Suggested Release | Reason |
|---|---|---|
| Trigram name search | v1.1 | Exact match search is sufficient at launch |
| Additional COA templates | v1.1 | Petroleum retail template covers v1.0 target |
| Account re-parenting | v1.1 | Setup-time operation; manual correction is acceptable |
| COA import with update mode | v1.1 | Create-only is sufficient for migration |

### Never Defer

- DB unique constraint on `(tenant_id, organisation_id, code)`
- `is_group` validation (posting to group accounts must be blocked)
- `locked_currency` enforcement
- Generated columns for closing/net balances (prevents inconsistent state)
- RLS policies

---

*End of FILE-02. Proceeding to FILE-03: Journal Entry Pipeline & General Ledger.*
