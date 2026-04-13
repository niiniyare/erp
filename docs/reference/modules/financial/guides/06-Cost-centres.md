# Awo ERP — Finance Module
## FILE-06: Cost Centre & Budget Management

**Document Version:** 2.0.0  
**Series:** FILE-06 of 10  
**Depends On:** FILE-01 (Domain Model), FILE-02 (COA), FILE-03 (Journal Pipeline — Stage 8 Budget Check)  
**Depended On By:** FILE-08 (Reports — Budget vs. Actual, Cost Centre P&L)

---

## Table of Contents

1. [Why Cost Centres Exist](#1-why-cost-centres-exist)
2. [Why Budgets Exist](#2-why-budgets-exist)
3. [What You Lose Without Them](#3-what-you-lose-without-them)
4. [How ERPNext, NetSuite and QuickBooks Handle These](#4-how-erpnext-netsuite-and-quickbooks-handle-these)
5. [Cost Centre Model](#5-cost-centre-model)
6. [Cost Centre Allocation (Distributed Centres)](#6-cost-centre-allocation-distributed-centres)
7. [Budget Model](#7-budget-model)
8. [Budget Checker](#8-budget-checker)
9. [Budget Versioning](#9-budget-versioning)
10. [Business Rules & Validation](#10-business-rules--validation)
11. [Performance & Storage](#11-performance--storage)
12. [Database Schema](#12-database-schema)
13. [API Reference](#13-api-reference)
14. [Feature Flags & Configuration](#14-feature-flags--configuration)
15. [v1.0 Rollout Assessment](#15-v10-rollout-assessment)

---

## 1. Why Cost Centres Exist

A Chart of Accounts tells you *what* money was spent on (salaries, rent, utilities). Cost centres tell you *where* it was spent — which department, which site, which project. Without cost centres, you know your total salary expense is 5,000,000 KES but you cannot tell how much of that belongs to the Mombasa branch versus the Nairobi head office.

In the context of Awo's primary market (petroleum retail and general SME operations in East Africa), cost centres address three specific business needs:

1. **Multi-site P&L:** An operator with three fuel stations needs to know which station is profitable and which is breaking even. The GL alone cannot provide this — cost centres do.
2. **Departmental accountability:** In a business with sales, operations, administration, and finance teams, department heads need to see their own expenses versus their budget. They should not need to see or approve entries from other departments.
3. **Project tracking:** A capital expenditure project (e.g. installing a new underground tank) needs a cost centre so all related expenditures can be aggregated and compared to the approved project budget.

---

## 2. Why Budgets Exist

A budget is a financial plan expressed as target amounts per account, per cost centre, per period. The Finance module's budget engine does two things with that plan:

1. **Report:** Compares actual GL postings to the budget, showing variance by account, cost centre, and period
2. **Control:** Optionally blocks or warns when a transaction would cause an account to exceed its budget

Without a budget, financial management is entirely reactive — you review what was spent after the fact. With a budget, the system can alert you at the moment a transaction would breach an approved limit, before the money is spent.

---

## 3. What You Lose Without Them

| Capability | Without Cost Centres | Without Budgets |
|---|---|---|
| Per-site profitability | Impossible | Available (but no variance analysis) |
| Departmental P&L | Impossible | Available |
| Expense accountability | Total company only | Available |
| Budget vs. actual reporting | N/A | Impossible |
| Pre-transaction budget check | N/A | Impossible |
| Capital project tracking | Manual spreadsheet | Can track but no control |
| Multi-dimensional reporting | Single dimension (account) | Two dimensions (account + period) |

---

## 4. How ERPNext, NetSuite and QuickBooks Handle These

### ERPNext

**Cost centres:** First-class feature. ERPNext uses a tree-based cost centre model identical in concept to Awo's. Every GL entry can carry a cost centre. Cost centre-scoped P&L is a standard report. ERPNext also supports "profit centres" as a separate concept for revenue tracking — Awo does not make this distinction; cost centres handle both cost and revenue tracking.

**Budgets:** ERPNext has a Budget DocType that links cost centres, accounts, and monthly amounts. Budget variance reporting is available. Budget control is configurable per account group (warn/stop). The implementation is solid but the UI for budget entry is cumbersome for large organisations.

**What Awo improves:** Budget versioning (ERPNext does not have native version management — a revised budget overwrites the original). Awo keeps all versions and marks exactly one as `active`.

### NetSuite

**Cost centres (Departments):** NetSuite's Department segment is analogous. It supports unlimited hierarchy and distributed allocation (called "Statistical Journal Entries" in NetSuite). The allocation engine is sophisticated but complex to configure.

**Budgets:** NetSuite has one of the most powerful budget engines in ERP. Multi-dimensional (account + department + class + location), versioned, importable from Excel, with hard and soft controls. Awo's budget engine is modelled after NetSuite's approach but simplified for the SME context.

### QuickBooks

**Cost centres:** QuickBooks uses "Classes" and "Locations" as approximate analogues to cost centres. These are simpler — no hierarchy, limited inheritance to child transactions. Class-based P&L is available in QuickBooks Online Plus and above.

**Budgets:** QuickBooks Online has basic budgeting (profit and loss, balance sheet) per class. No version management, no hard controls. No transaction-level budget checking.

---

## 5. Cost Centre Model

### Hierarchy

Cost centres form a tree with materialised paths, identical in structure to the chart of accounts tree. This allows subtree queries (e.g. "all costs under the East Africa region") to use the same efficient `LIKE` pattern.

```go
type CostCentre struct {
    ID               uuid.UUID
    TenantID         uuid.UUID
    OrganisationID   uuid.UUID
    Code             string
    Name             string
    ParentID         *uuid.UUID
    Path             string
    Level            int
    IsGroup          bool
    IsDistributed    bool
    AllocationMethod AllocationMethod   // percentage | headcount | square_footage | usage
    Allocations      []CostAllocation   // only relevant when IsDistributed = true
    IsActive         bool
}

type CostAllocation struct {
    TargetCentreID uuid.UUID
    Percentage     decimal.Decimal
}

func (cc *CostCentre) ValidateAllocations() error {
    if !cc.IsDistributed {
        return nil
    }
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

### Typical Cost Centre Structures for Awo's Target Market

**Petroleum retail operator (by site):**
```
Company
├── Nairobi Station
├── Mombasa Station
├── Kisumu Station
└── Head Office
    ├── Finance & Administration
    ├── Operations Management
    └── IT & Systems
```

**General SME (by function):**
```
Company
├── Sales & Marketing
├── Operations
├── Finance & HR
└── Management
```

**Multi-activity operator (by activity):**
```
Company
├── Fuel Retail
├── Lubricants
├── Car Wash
├── Convenience Store
└── Shared Services (Distributed)
```

### Cost Centre on Journal Lines

Cost centres are attached at the **journal line level**, not at the journal entry level. This allows a single journal entry to distribute costs across multiple centres:

```json
{
  "description": "January payroll",
  "lines": [
    { "account_code": "7110", "debit": 300000, "cost_centre_code": "CC-NAIROBI" },
    { "account_code": "7110", "debit": 200000, "cost_centre_code": "CC-MOMBASA" },
    { "account_code": "7110", "debit": 100000, "cost_centre_code": "CC-HEAD-OFFICE" },
    { "account_code": "2210", "credit": 600000 }
  ]
}
```

A header-level `cost_centre_id` on the journal entry serves as a default that propagates to lines where no line-level cost centre is specified. This avoids requiring every line to specify a cost centre when most lines belong to the same centre.

---

## 6. Cost Centre Allocation (Distributed Centres)

### Purpose

Some cost centres serve multiple other centres without generating their own direct revenue (IT department, HR, shared facilities). Distributed cost centres allow these indirect costs to be allocated to the centres that benefit from them.

### How Allocation Works

A distributed cost centre has `IsDistributed = true` and a set of `CostAllocation` targets. At the end of each period (or on demand), the `CostCentreAllocationService` reads all costs posted to the distributed centre during the period and generates a journal entry that charges those costs to the target centres:

```
IT Department spent 500,000 KES in January:
  → Sales (40%):         200,000 KES
  → Operations (30%):    150,000 KES
  → Administration (20%): 100,000 KES
  → R&D (10%):            50,000 KES

Allocation journal entry:
Dr. IT Allocation Expense — Sales         200,000  [CC: Sales]
Dr. IT Allocation Expense — Ops           150,000  [CC: Operations]
Dr. IT Allocation Expense — Admin         100,000  [CC: Administration]
Dr. IT Allocation Expense — R&D            50,000  [CC: R&D]
    Cr. IT Department — Total Costs            500,000  [CC: IT Dept]
```

After allocation, the IT Department cost centre has a net zero balance for the period. The costs appear where they were consumed.

### Allocation Methods

| Method | Basis | When to Use |
|---|---|---|
| `percentage` | Fixed % per target centre | Stable cost relationships (IT, HR) |
| `headcount` | Proportional to employee count per centre | HR costs, office space |
| `square_footage` | Proportional to office space per centre | Rent, cleaning, utilities |
| `usage` | Proportional to actual usage metric (e.g. support tickets) | Variable consumption |

For `headcount`, `square_footage`, and `usage` methods, the allocation percentages are computed from the driver data at the time the allocation is run, not stored as fixed percentages. The driver data (employee count, floor area, ticket count) is stored in `allocation_drivers` and updated by the HR module (headcount) or manually (floor area, tickets).

---

## 7. Budget Model

### Budget Structure

A budget defines expected amounts per account, per cost centre, per accounting period. It belongs to a fiscal year and has a version label.

```go
type Budget struct {
    ID             uuid.UUID
    TenantID       uuid.UUID
    OrganisationID uuid.UUID
    FiscalYearID   uuid.UUID
    Name           string        // "FY2025 Original Budget"
    VersionLabel   string        // "V1", "V2 Mid-Year Revision"
    Status         BudgetStatus  // draft | approved | active | archived
    Currency       string
    Lines          []BudgetLine
    ApprovedBy     *uuid.UUID
    ApprovedAt     *time.Time
    CreatedAt      time.Time
}

type BudgetLine struct {
    AccountID    uuid.UUID
    CostCentreID *uuid.UUID    // nil = applies to all cost centres for this account
    PeriodID     uuid.UUID
    Amount       decimal.Decimal
}
```

### Budget Status Lifecycle

```
Draft ──► Approved ──► Active ──► Archived
  │                       │
  └──[Delete allowed]     └──[Only one ACTIVE per fiscal year]
```

Only one budget version can have status `active` per organisation per fiscal year. When a new version is activated, the previous one is automatically archived.

### Budget Line Granularity Options

Budget lines can be defined at different levels of granularity. The system matches them in order of specificity when performing a budget check:

1. **Account + Cost Centre + Period** (most specific — used first)
2. **Account + Period** (no cost centre — applies to all centres for this account)
3. **Account + Cost Centre** (no period — applies to all periods for this account/centre combination)

This hierarchy allows "default" budgets at the account level with overrides for specific cost centres or periods.

---

## 8. Budget Checker

The budget checker (pipeline stage 8) runs on every journal entry that contains expense account lines. It computes the remaining budget for each affected account + cost centre + period combination and determines whether the entry would cause a breach.

```go
type BudgetControlMode string
const (
    BudgetNone BudgetControlMode = "none"
    BudgetSoft BudgetControlMode = "soft"  // warn; allow with override comment
    BudgetHard BudgetControlMode = "hard"  // block; require CFO amendment
)

type BudgetCheckResult struct {
    AccountCode   string
    CostCentre    string
    Budgeted      decimal.Decimal
    ActualToDate  decimal.Decimal
    ThisEntry     decimal.Decimal
    Remaining     decimal.Decimal
    WouldExceedBy decimal.Decimal
    IsOverBudget  bool
}

func (c *BudgetChecker) Check(
    ctx      context.Context,
    orgID    uuid.UUID,
    periodID uuid.UUID,
    lines    []JournalLine,
) ([]BudgetCheckResult, error) {
    results := []BudgetCheckResult{}
    for _, line := range lines {
        if !isExpenseLine(line) {
            continue // budget only applies to expense accounts
        }
        result, err := c.checkLine(ctx, orgID, periodID, line)
        if err != nil {
            return nil, err
        }
        results = append(results, result)
    }
    return results, nil
}
```

### Soft Budget Behaviour

When budget control is `soft` and a line would exceed the budget:

1. The journal entry pipeline adds a `BudgetWarning` to the entry's warnings list
2. The entry is **not** blocked — it continues to the next stage
3. If `budget_override_requires_reason = true`, the API response includes a flag requiring the submitter to provide a reason comment before the entry can be approved
4. The override reason is stored on the journal entry header and appears in the budget exception report

### Hard Budget Behaviour

When budget control is `hard` and a line would exceed the budget:

1. The pipeline returns `ErrBudgetExceeded` — the entry is blocked
2. The API response identifies the specific account, cost centre, available budget, and shortfall
3. The finance team must either: post a budget amendment (create a new approved budget version with a higher amount), or request CFO approval for an override

### Budget Amendment Workflow

```
Current Active Budget (V1) ──[new version needed]──► Draft Budget (V2)
                                                           │
                                               Finance Manager approves V2
                                                           │
                                               Activate V2 → V1 archived
```

A budget amendment does not delete the original budget — it creates a new version. Reports can compare V1 vs. V2 vs. actual to show the full picture of how the budget evolved during the year.

---

## 9. Budget Versioning

### Why Versioning Matters

A budget set at the beginning of the year becomes outdated. Staff changes, unexpected costs, revenue shortfalls, and business opportunities all create a need to revise the budget during the year. Without versioning, revising the budget overwrites the original, losing the ability to compare actual results against what was originally planned.

Awo keeps all versions. The `active` version is used for budget checks. Reports can be run against any version:

- **V1 (Original)** — the board-approved annual budget
- **V2 (Q2 Revision)** — updated after a strategic review
- **V3 (Q3 Emergency)** — revised due to a specific event

```
Budget Variance Report — July 2025
Account: 7110 Salaries  |  Cost Centre: Mombasa Station

           V1 Budget    V2 Budget    Actual    V1 Variance    V2 Variance
Jan         100,000      100,000    102,000       (2,000)        (2,000)
Feb         100,000      100,000     98,000        2,000          2,000
Mar         100,000      100,000    103,500       (3,500)        (3,500)
Apr         100,000      115,000    116,000       (16,000)        (1,000)
May         100,000      115,000    114,500       (14,500)           500
Jun         100,000      115,000    113,000       (13,000)         2,000
──────────────────────────────────────────────────────────────────────────
YTD         600,000      645,000    647,000       (47,000)        (2,000)
```

V2 was revised upward in April (new hire). Comparing against V1 shows the total overrun; comparing against V2 shows the remaining budget discipline within the revised plan.

### Budget Import

Large organisations will not enter 300 budget lines manually. The system supports CSV import directly into a draft budget:

```csv
account_code,cost_centre_code,jan,feb,mar,apr,may,jun,jul,aug,sep,oct,nov,dec
7110,CC-NAIROBI,1000000,1000000,1000000,1000000,1000000,1000000,1000000,1000000,1000000,1000000,1000000,1000000
7110,CC-MOMBASA,600000,600000,600000,700000,700000,700000,700000,700000,700000,700000,700000,700000
6300,CC-MARKETING,500000,200000,200000,800000,200000,200000,500000,200000,200000,800000,200000,300000
```

The import converts month columns into individual `BudgetLine` rows, matching period IDs by the fiscal year's period start dates.

---

## 10. Business Rules & Validation

### Cost Centre Rules

| Rule ID | Rule | Enforcement |
|---|---|---|
| `FIN-CC-001` | Cost centre code is unique within an organisation | DB unique constraint |
| `FIN-CC-002` | Cost centre hierarchy must not create cycles | Application check on parent assignment |
| `FIN-CC-003` | A distributed cost centre's allocations must sum to exactly 100% | `ValidateAllocations()` method |
| `FIN-CC-004` | A group cost centre cannot receive journal entries directly — only leaf cost centres can | Application check in `ValidateCostCentres` pipeline stage |
| `FIN-CC-005` | A cost centre with posted entries cannot be deactivated | Application check |
| `FIN-CC-006` | Allocation journal entries are system type and bypass manual entry restrictions | Enforced by entry type |

### Budget Rules

| Rule ID | Rule | Enforcement |
|---|---|---|
| `FIN-BDG-001` | Only one ACTIVE budget version per organisation per fiscal year | DB partial unique index on `(tenant_id, organisation_id, fiscal_year_id) WHERE status = 'active'` |
| `FIN-BDG-002` | Budget checker only applies to expense account lines (not revenue or balance sheet) | `isExpenseLine()` check in BudgetChecker |
| `FIN-BDG-003` | Soft budget override requires a non-empty reason comment if `budget_override_requires_reason = true` | Application check on entry approval |
| `FIN-BDG-004` | Hard budget stop blocks entry regardless of user role | Pipeline stage error — no override mechanism except budget amendment |
| `FIN-BDG-005` | Budget version requires `finance.budgets.approve` permission to activate | Permission check |
| `FIN-BDG-006` | Activating a budget version automatically archives the currently active version | Application transaction in `BudgetService.Activate()` |
| `FIN-BDG-007` | A budget line amount must be ≥ 0 | DB check constraint |

### Alternatives Considered

**Alternative: Budget per account only (no cost centre dimension).**
Available — cost centre is optional on budget lines. An organisation that does not use cost centres can budget purely at the account level. This is the `null` cost centre case in `BudgetLine`.

**Alternative: Rolling budget (continuous 12-month forward view) instead of annual.**
Considered for v1.1. A rolling budget updates the budget for the next month based on actuals from the current month plus a growth assumption. Not in v1.0 scope — the annual versioned budget is the foundation it builds on.

**Alternative: Zero-based budgeting (all lines start from zero each period).**
No system change needed — this is a process choice, not a feature. Zero-based budgeting is fully supported; the organisation simply does not copy lines forward from the prior year when creating a new budget.

---

## 11. Performance & Storage

### Budget Checker Performance

The budget check runs on every expense line in every posted journal entry. It must be fast.

```
Per budget check: 2 DB reads
  1. Resolve active budget ID for the org/fiscal year   → cached in memory after first call
  2. Find budget line for (budget_id, account_id, cost_centre_id, period_id) → indexed lookup

Target: < 5ms per line, < 20ms for a 10-line payroll entry
```

The active budget ID is cached per-request (set in `PostingConfig` at the start of each pipeline run). This avoids re-querying on every line in a multi-line entry.

```sql
-- Covering index for budget line lookups
CREATE INDEX idx_budget_lines_lookup
    ON budget_lines (tenant_id, budget_id, account_id, period_id)
    INCLUDE (amount, cost_centre_id);
```

### Cost Centre P&L Performance

The cost centre P&L report aggregates journal line amounts by cost centre. At scale, this can involve scanning millions of `journal_lines` rows filtered by cost centre. Two approaches:

1. **v1.0:** Direct query with `(tenant_id, cost_centre_id)` index on `journal_lines`. Acceptable at under 500k rows.
2. **v1.1:** Materialised cost centre balances — extend `account_balances` to include a `cost_centre_id` dimension, or create a separate `cost_centre_balances` table updated atomically on posting. This reduces the P&L query to a simple index scan.

### Storage Estimates

| Table | Row Size | Rows/Org/Year | Annual Storage |
|---|---|---|---|
| `cost_centres` | ~300 bytes | ~50 (stable) | ~15 KB (one-time) |
| `budgets` | ~300 bytes | 2–3 versions | ~1 KB/year |
| `budget_lines` | ~150 bytes | 200 accounts × 12 periods = 2,400 lines/version | ~360 KB/version |
| `allocation_runs` | ~200 bytes | 12 per year | ~2.4 KB/year |
| `allocation_run_items` | ~150 bytes | ~100 per run | ~180 KB/year |

Total cost centre and budget storage: well under **1 MB per year** per organisation. Negligible.

---

## 12. Database Schema

```sql
-- ── cost_centres ──────────────────────────────────────────────────────────────

CREATE TABLE cost_centres (
    tenant_id         UUID        NOT NULL REFERENCES tenants(id),
    id                UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    organisation_id   UUID        NOT NULL,
    code              TEXT        NOT NULL,
    name              TEXT        NOT NULL,
    parent_id         UUID        REFERENCES cost_centres(id),
    path              TEXT        NOT NULL,
    level             INT         NOT NULL DEFAULT 1,
    is_group          BOOLEAN     NOT NULL DEFAULT FALSE,
    is_distributed    BOOLEAN     NOT NULL DEFAULT FALSE,
    allocation_method TEXT        CHECK (allocation_method IN ('percentage','headcount','square_footage','usage')),
    allocations       JSONB       NOT NULL DEFAULT '[]',
    is_active         BOOLEAN     NOT NULL DEFAULT TRUE,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, organisation_id, code),
    CONSTRAINT chk_distributed_has_method
        CHECK (NOT is_distributed OR allocation_method IS NOT NULL)
);

CREATE INDEX idx_cc_tenant ON cost_centres (tenant_id);
CREATE INDEX idx_cc_path   ON cost_centres (tenant_id, path text_pattern_ops);

ALTER TABLE cost_centres ENABLE ROW LEVEL SECURITY;
ALTER TABLE cost_centres FORCE  ROW LEVEL SECURITY;
CREATE POLICY cc_app ON cost_centres FOR ALL TO application_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

-- ── budgets ────────────────────────────────────────────────────────────────────

CREATE TABLE budgets (
    tenant_id       UUID        NOT NULL REFERENCES tenants(id),
    id              UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      UUID,

    organisation_id UUID        NOT NULL,
    fiscal_year_id  UUID        NOT NULL REFERENCES fiscal_years(id),
    name            TEXT        NOT NULL,
    version_label   TEXT        NOT NULL DEFAULT 'V1',
    status          TEXT        NOT NULL DEFAULT 'draft'
                        CHECK (status IN ('draft','approved','active','archived')),
    currency        CHAR(3)     NOT NULL,
    approved_by     UUID,
    approved_at     TIMESTAMPTZ,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id)
);

-- Enforce single active budget per org per fiscal year
CREATE UNIQUE INDEX idx_budgets_one_active
    ON budgets (tenant_id, organisation_id, fiscal_year_id)
    WHERE status = 'active';

ALTER TABLE budgets ENABLE ROW LEVEL SECURITY;
ALTER TABLE budgets FORCE  ROW LEVEL SECURITY;
CREATE POLICY bgt_app ON budgets FOR ALL TO application_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

-- ── budget_lines ───────────────────────────────────────────────────────────────

CREATE TABLE budget_lines (
    tenant_id       UUID          NOT NULL REFERENCES tenants(id),
    id              UUID          NOT NULL DEFAULT gen_random_uuid(),
    budget_id       UUID          NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    account_id      UUID          NOT NULL REFERENCES accounts(id),
    cost_centre_id  UUID,         -- NULL = applies to all cost centres
    period_id       UUID          NOT NULL,
    amount          NUMERIC(18,4) NOT NULL CHECK (amount >= 0),

    PRIMARY KEY (id),
    UNIQUE (tenant_id, budget_id, account_id, cost_centre_id, period_id)
);

CREATE INDEX idx_budget_lines_lookup
    ON budget_lines (tenant_id, budget_id, account_id, period_id)
    INCLUDE (amount, cost_centre_id);

ALTER TABLE budget_lines ENABLE ROW LEVEL SECURITY;
ALTER TABLE budget_lines FORCE  ROW LEVEL SECURITY;
CREATE POLICY bl_app ON budget_lines FOR ALL TO application_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

-- ── cost centre balance materialisation (v1.1) ────────────────────────────────
-- This table is created in v1.1 when cost centre P&L performance requires it.

-- CREATE TABLE cost_centre_balances (
--     tenant_id       UUID          NOT NULL,
--     account_id      UUID          NOT NULL REFERENCES accounts(id),
--     cost_centre_id  UUID          NOT NULL REFERENCES cost_centres(id),
--     period_id       UUID          NOT NULL,
--     period_debit    NUMERIC(18,4) NOT NULL DEFAULT 0,
--     period_credit   NUMERIC(18,4) NOT NULL DEFAULT 0,
--     PRIMARY KEY (tenant_id, account_id, cost_centre_id, period_id)
-- );
```

---

## 13. API Reference

### Cost Centres

```
GET    /finance/cost-centres                  List cost centres (tree or flat, ?is_group=, ?is_active=)
POST   /finance/cost-centres                  Create cost centre
GET    /finance/cost-centres/{code}           Get cost centre with current period balances
PATCH  /finance/cost-centres/{code}           Update name, allocation, active status
POST   /finance/cost-centres/allocate         Trigger distributed allocation for a period
GET    /finance/cost-centres/{code}/pl        Cost centre P&L for a date range (?from=&to=)
```

### Budgets

```
GET    /finance/budgets                       List budget versions (?fiscal_year_id=, ?status=)
POST   /finance/budgets                       Create draft budget
GET    /finance/budgets/{id}                  Get budget with all lines
POST   /finance/budgets/{id}/lines            Set or replace budget lines
POST   /finance/budgets/{id}/import           Bulk import lines from CSV
POST   /finance/budgets/{id}/approve          Draft → Approved
POST   /finance/budgets/{id}/activate         Approved → Active (archives current active)
GET    /finance/budgets/{id}/variance         Budget vs. actual variance (?period_id=, ?cost_centre_id=)
POST   /finance/budgets/{id}/copy             Copy lines from another budget version (for creating V2 from V1)
```

**Create budget:**
```json
{
  "fiscal_year_id": "uuid",
  "name": "FY2025 Original Budget",
  "version_label": "V1",
  "currency": "KES"
}
```

**Set budget lines (replaces all lines for the budget):**
```json
{
  "lines": [
    { "account_code": "7110", "cost_centre_code": "CC-NAIROBI", "period_id": "uuid-jan", "amount": 1000000 },
    { "account_code": "7110", "cost_centre_code": "CC-MOMBASA", "period_id": "uuid-jan", "amount": 600000 }
  ]
}
```

**Variance report response:**
```json
{
  "period": "January 2025",
  "lines": [
    {
      "account_code": "7110",
      "account_name": "Salaries",
      "cost_centre": "Nairobi Station",
      "budget": 1000000,
      "actual": 1025000,
      "variance": -25000,
      "variance_pct": -2.5,
      "is_over_budget": true
    }
  ]
}
```

---

## 14. Feature Flags & Configuration

| Flag | Type | Default | Description |
|---|---|---|---|
| `cost_centre_enabled` | bool | `true` | Enable cost centre tracking. If false, cost centre fields are hidden and not required. |
| `require_cost_centre_on_expense` | bool | `false` | Every expense journal line must carry a cost centre. Useful for organisations that need full cost centre P&L accuracy. |
| `budget_module_enabled` | bool | `true` | Enable budget creation and variance reporting. If false, budget check stage is skipped. |
| `budget_control_mode` | enum | `soft` | `none` · `soft` · `hard` — applies to all expense accounts by default |
| `budget_control_mode_per_account` | bool | `false` | Allow individual accounts to override the global budget control mode |
| `budget_override_requires_reason` | bool | `true` | Soft budget override requires a non-empty reason comment |
| `budget_check_includes_pending` | bool | `false` | If true, `ActualToDate` in the budget check includes pending (not yet posted) journal entries. More conservative but slightly slower. |
| `distributed_allocation_auto_run` | bool | `false` | Automatically run cost centre allocations on the last day of each period via Temporal cron |
| `distributed_allocation_day` | int | `0` | Day of month to run automatic allocation (0 = last day of month) |
| `budget_import_allow_overwrite` | bool | `false` | Whether a CSV import replaces existing lines or only adds new ones |
| `cost_centre_p_and_l_dimensions` | string[] | `["cost_centre"]` | Additional dimensions available in cost centre P&L. Reserved for future `["cost_centre", "project"]` support. |

---

## 15. v1.0 Rollout Assessment

### Must Have at v1.0

- Cost centre tree with materialised path
- Cost centre field on journal lines (stored and indexed)
- Budget model with draft/approved/active lifecycle
- Single active budget constraint (DB partial unique index)
- Budget checker in pipeline stage 8 (can be `budget_module_enabled = false` initially, enabled when first budget is created)
- Budget vs. actual variance report
- Cost centre P&L report (direct query — materialised version deferred)

### Can Be Deferred

| Feature | Suggested Release |
|---|---|
| Distributed cost centre allocation engine | v1.1 — most v1.0 operators do not need automated allocation |
| Budget versioning UI (copy V1 → V2) | v1.1 — V1 budget + variance report covers v1.0 |
| Budget CSV import | v1.1 — manual entry acceptable for small budgets |
| Hard budget control | v1.1 — ship soft control only at v1.0 |
| Per-account budget control mode override | v1.2 |
| Cost centre balance materialisation table | v1.1 — direct query adequate at v1.0 volume |
| Rolling budget | v1.2 |

### Never Defer

- DB partial unique index on active budget — without it, two active budgets can coexist and the budget check becomes non-deterministic
- `ValidateAllocations()` check on distributed cost centres — without it, allocations silently lose or double-count costs
- Cost centre `path` index — required for subtree queries used in P&L reports

---

*End of FILE-06. Proceeding to FILE-07: Cash Management, Intercompany & Multi-Currency.*
