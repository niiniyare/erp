# Awo ERP — Finance Module
## FILE-07: Cash Management, Intercompany & Multi-Currency

**Document Version:** 2.0.0  
**Series:** FILE-07 of 10  
**Depends On:** FILE-01 (Domain Model), FILE-02 (COA), FILE-03 (Journal Pipeline), FILE-04 (Period Management)  
**Depended On By:** FILE-08 (Reports — Cash Position, Cash Forecast, IC Balances)

---

## Table of Contents

1. [Why Cash Management Exists](#1-why-cash-management-exists)
2. [Why Intercompany Accounting Exists](#2-why-intercompany-accounting-exists)
3. [Why Multi-Currency Exists as a First-Class Feature](#3-why-multi-currency-exists-as-a-first-class-feature)
4. [What You Lose Without Each](#4-what-you-lose-without-each)
5. [How ERPNext, NetSuite and QuickBooks Handle These](#5-how-erpnext-netsuite-and-quickbooks-handle-these)
6. [Cash Position Engine](#6-cash-position-engine)
7. [13-Week Rolling Forecast](#7-13-week-rolling-forecast)
8. [Multi-Currency Architecture](#8-multi-currency-architecture)
9. [Exchange Rate Management](#9-exchange-rate-management)
10. [FX Revaluation](#10-fx-revaluation)
11. [Realised and Unrealised FX Gains/Losses](#11-realised-and-unrealised-fx-gainslosses)
12. [Intercompany Module](#12-intercompany-module)
13. [IC Period-End Matching](#13-ic-period-end-matching)
14. [Consolidation Eliminations](#14-consolidation-eliminations)
15. [Business Rules & Validation](#15-business-rules--validation)
16. [Performance & Storage](#16-performance--storage)
17. [Database Schema](#17-database-schema)
18. [API Reference](#18-api-reference)
19. [Feature Flags & Configuration](#19-feature-flags--configuration)
20. [v1.0 Rollout Assessment](#20-v10-rollout-assessment)

---

## 1. Why Cash Management Exists

Profit and cash are not the same thing. A business can be profitable on paper while running out of cash — if customers are slow to pay, if inventory builds up, if debt repayments cluster in one week. The cash position and forecast modules answer the question that P&L cannot: "How much cash do we actually have right now, and will we have enough to meet our obligations in the next 13 weeks?"

For a petroleum retail operator this is particularly acute. Fuel suppliers require payment in USD (or at minimum in hard currency), payable on short credit terms (7–14 days). If the operator's cash position is unclear, they risk failing to fund a fuel delivery — which means empty pumps and lost revenue.

---

## 2. Why Intercompany Accounting Exists

As an operator grows from one entity to a group structure (a holding company, subsidiaries per region, a trading entity separate from an operating entity), transactions between those entities must be recorded in both entities' books. A management fee from a holding company to a subsidiary is income in the holding company and an expense in the subsidiary. Without intercompany accounting, the group's consolidated financials will show this as external income — overstating group revenue.

The intercompany module ensures that both sides of every intra-group transaction are recorded, that they net to zero on consolidation, and that any unrealised profit in transferred goods is eliminated before the consolidated balance sheet is presented to external stakeholders.

---

## 3. Why Multi-Currency Exists as a First-Class Feature

East African businesses routinely operate across multiple currencies. Fuel importers settle USD invoices. Multinationals report in GBP or EUR to their parent. Export businesses hold USD or EUR bank accounts to manage FX risk. If the system only supports KES, these businesses cannot use it for their financial reporting.

Multi-currency support in Awo is not an add-on — it is built into the core data model. Every journal line stores both the transaction currency amount and the functional currency equivalent. Exchange rates are stored historically. FX revaluation is a first-class operation that runs at period end. This approach means a KES-only operator incurs zero overhead (they simply never use non-KES accounts), while a multi-currency operator has everything they need.

---

## 4. What You Lose Without Each

| Capability | Without Cash Management | Without IC Module | Without Multi-Currency |
|---|---|---|---|
| Real-time cash position | Manual spreadsheet | N/A | Only KES accounts |
| 13-week forecast | Not available | N/A | KES-only forecast |
| Deficit week identification | Manual | N/A | N/A |
| Consolidated financials | N/A | Cannot produce them | KES only |
| IC profit elimination | N/A | Not possible | N/A |
| FX gain/loss recording | KES only | N/A | Incorrect P&L for FC transactions |
| Foreign bank account management | N/A | N/A | Not possible |
| USD supplier invoice settlement | Manual FX calculation | N/A | No systematic tracking |

---

## 5. How ERPNext, NetSuite and QuickBooks Handle These

### Cash Management

**ERPNext:** Has a "Cash Flow Statement" report and a basic cash position view. No rolling forecast engine. Treasury management is not a strong point of ERPNext.

**NetSuite:** Strong cash management with real-time visibility across all bank accounts, FX conversion, and a cash flow forecast report. NetSuite's forecast is period-based rather than week-based — Awo's 13-week view is more granular.

**QuickBooks:** Basic cash flow projections available in QuickBooks Online Advanced. Not a true rolling forecast — more of an estimated P&L projection.

### Intercompany

**ERPNext:** Has intercompany journal entries and intercompany customer/supplier invoices. The matching is manual. No automated elimination generation.

**NetSuite OneWorld:** The reference implementation for intercompany accounting. Automated IC matching, elimination journals, and consolidated reports. Awo's IC module is modelled after NetSuite OneWorld's approach, simplified for the SME context.

**QuickBooks:** No intercompany support. Each company is a completely separate file.

### Multi-Currency

**ERPNext:** Full multi-currency support. Exchange rates managed manually or via a configured feed. FX revaluation runs as a scheduled job. Functional currency is the Company's default currency.

**NetSuite:** Excellent multi-currency. Automatic rate loading from multiple providers. Rate types (spot, average, historical) per transaction type. FX gain/loss recognition configurable per account.

**QuickBooks Online:** Multi-currency available in higher tiers. Rate loading from automatic feed. Less granular control over rate types than NetSuite or Awo.

---

## 6. Cash Position Engine

### Real-Time Cash Position

The cash position is computed in real time from account balances for all accounts with `account_type IN ('bank', 'cash')`. It consolidates balances across currencies by converting to the functional currency at the current exchange rate.

```go
type CashPosition struct {
    AsAt          time.Time
    Currency      string           // functional currency
    Accounts      []CashAccount
    TotalCash     decimal.Decimal  // sum of all accounts in functional currency
    Committed     []CommittedItem  // approved payment runs not yet posted
    AvailableCash decimal.Decimal  // TotalCash - sum(Committed)
}

type CashAccount struct {
    AccountCode       string
    AccountName       string
    Currency          string
    Balance           decimal.Decimal  // in account's native currency
    FunctionalBalance decimal.Decimal  // converted to functional currency
    ExchangeRate      decimal.Decimal
    RateDate          time.Time
}

type CommittedItem struct {
    Type        string   // "payment_run" | "payroll" | "tax_payment"
    Reference   string
    Amount      decimal.Decimal
    ExpectedDate time.Time
}
```

**Sample cash position output:**

```
CASH POSITION — 06 Feb 2025 09:00 EAT

Account                  Currency   Balance       KES Equivalent
─────────────────────────────────────────────────────────────────
1112 Checking – Main     KES       1,240,000       1,240,000
1113 Savings Account     KES         350,000         350,000
1111 Petty Cash          KES          50,000          50,000
1130 USD Account         USD          15,000       1,987,500 *
─────────────────────────────────────────────────────────────────
TOTAL CASH                                         3,627,500

COMMITTED:
  Payment Run PR-2025-0012 (approved, pending):  (307,400)
  Payroll (est. 10 Feb):                       (2,000,000)
─────────────────────────────────────────────────────────────────
AVAILABLE POSITION:                              1,320,100

* Converted at 06 Feb 2025 rate: USD/KES 132.50
```

---

## 7. 13-Week Rolling Forecast

### Why 13 Weeks

Thirteen weeks is the standard treasury management horizon. It is long enough to identify upcoming cash shortfalls with enough lead time to arrange financing, but short enough that the forecast data (AR due dates, AP due dates, payroll schedule) is reasonably reliable.

```go
type WeeklyForecast struct {
    WeekNumber   int
    DateFrom     time.Time
    DateTo       time.Time
    Receipts     decimal.Decimal  // expected inflows
    Payments     decimal.Decimal  // expected outflows
    Net          decimal.Decimal
    ClosingBal   decimal.Decimal  // cumulative
    BelowMinimum bool
}

type CashFlowForecast struct {
    AsAt          time.Time
    OpeningBal    decimal.Decimal
    TargetMinBal  decimal.Decimal  // from config: minimum operating cash policy
    Weeks         []WeeklyForecast
    MinBalance    decimal.Decimal
    MaxBalance    decimal.Decimal
    DeficitWeeks  []int            // week numbers where closing < target minimum
}
```

### Forecast Data Sources

| Item | Source | Confidence |
|---|---|---|
| Customer receipts | AR aging expected collection dates (due date + avg collection days) | Medium |
| Confirmed customer payments | Sales module payment commitments | High |
| Supplier payments | AP aging due dates | High |
| Approved payment runs | `payment_runs` table, `status = approved` | High |
| Payroll | HR module payroll schedule | High |
| Tax payments (PAYE, VAT, NSSF) | Tax calendar configured in settings | High |
| Loan repayments | Loan schedule from account mappings | High |
| Capital expenditure | Approved procurement orders (v1.1+) | Medium |

### Forecast Generation

```go
func (f *ForecastEngine) Generate(ctx context.Context, orgID uuid.UUID, asAt time.Time) (*CashFlowForecast, error) {
    openingBal, err := f.getCashPosition(ctx, orgID, asAt)
    if err != nil {
        return nil, err
    }

    weeks := make([]WeeklyForecast, 13)
    runningBal := openingBal.TotalCash

    for i := 0; i < 13; i++ {
        from := asAt.AddDate(0, 0, i*7)
        to   := from.AddDate(0, 0, 6)

        receipts, err := f.getExpectedReceipts(ctx, orgID, from, to)
        if err != nil {
            return nil, err
        }
        payments, err := f.getExpectedPayments(ctx, orgID, from, to)
        if err != nil {
            return nil, err
        }

        net        := receipts.Sub(payments)
        runningBal  = runningBal.Add(net)
        weeks[i]    = WeeklyForecast{
            WeekNumber:   i + 1,
            DateFrom:     from,
            DateTo:       to,
            Receipts:     receipts,
            Payments:     payments,
            Net:          net,
            ClosingBal:   runningBal,
            BelowMinimum: runningBal.LessThan(f.targetMinBalance),
        }
    }

    return &CashFlowForecast{
        AsAt:       asAt,
        OpeningBal: openingBal.TotalCash,
        Weeks:      weeks,
    }, nil
}
```

---

## 8. Multi-Currency Architecture

### Core Design Principle

Every monetary amount in the system is stored in two forms:
1. **Transaction currency** (`currency`, `original_amount`, `fx_rate`) — the currency in which the transaction was conducted
2. **Functional currency** (`base_amount`) — the tenant's reporting currency (default KES)

The functional currency amount is always computed as `original_amount × fx_rate` at the time of posting. It is never recomputed retroactively — historical entries retain the rate that applied when they were posted.

This two-amount design means:
- Reports in the functional currency are always available regardless of transaction currency
- The original transaction currency is preserved for audit and reconciliation
- FX gain/loss arises from the difference between the original functional amount and the settlement amount

### Functional Currency

The functional currency is set once per tenant at setup time (`default_currency` configuration flag). It cannot be changed after the first transaction is posted. Changing functional currency mid-stream would require revaluing all historical transactions — a complex operation that requires regulatory approval in most jurisdictions and is outside the scope of this module.

---

## 9. Exchange Rate Management

### Rate Structure

```go
type ExchangeRate struct {
    ID           uuid.UUID
    TenantID     uuid.UUID
    FromCurrency string
    ToCurrency   string
    Rate         decimal.Decimal
    RateType     RateType    // spot | average | historical | budget
    RateDate     time.Time
    Source       string      // "manual" | "central_bank" | "api_feed"
    LoadedBy     uuid.UUID
    CreatedAt    time.Time
}
```

### Rate Types

| Type | Used For | Updated |
|---|---|---|
| `spot` | Daily transactions — most common | Daily |
| `average` | Period-average translation (income statement in consolidated reports) | Monthly, computed from daily spot rates |
| `historical` | Equity transactions — rate at investment date | Once, at time of investment |
| `budget` | Budget planning rates — set at the start of the fiscal year | Annually |

### Rate Loading

```go
// POST /finance/rates — batch rate load
{
  "date": "2025-02-06",
  "rates": [
    { "from": "USD", "to": "KES", "rate": 132.50, "rate_type": "spot" },
    { "from": "GBP", "to": "KES", "rate": 164.50, "rate_type": "spot" },
    { "from": "EUR", "to": "KES", "rate": 143.20, "rate_type": "spot" }
  ]
}
```

**Rate availability check on transaction creation:** If a rate is not found for the transaction date, the journal entry cannot be posted. The error message specifies which currency pair is missing and for which date.

**Rate reasonableness check:** If a new rate deviates more than `rate_deviation_alert_threshold` (default 5%) from the previous rate for the same currency pair, the system flags it for confirmation before allowing transactions to proceed. This prevents transposition errors (e.g. entering 13.25 instead of 132.50) from flowing silently into the books.

---

## 10. FX Revaluation

### Why Revaluation Is Needed

An AR entry for a USD invoice was recorded at 128.00 KES/USD. At month end, the rate is 132.50 KES/USD. The balance sheet should show the current KES equivalent, not the historical one — otherwise the balance sheet does not reflect current economic reality.

Revaluation adjusts the KES carrying amount of all open foreign currency balances (bank accounts, AR, AP, loans) to the month-end exchange rate. The adjustment goes to the unrealised FX gain/loss account.

### Revaluation Process

The revaluation is triggered as part of the period close workflow. The Currency module emits a `FXRevaluationRequired` event; the Finance module's consumer runs the revaluation and posts the entries:

```go
func (s *FXRevaluationService) Revalue(ctx context.Context, orgID uuid.UUID, periodEnd time.Time) error {
    // 1. Find all FC balances (bank accounts, AR, AP, loans in non-functional currency)
    positions, err := s.getOpenFCPositions(ctx, orgID, periodEnd)
    if err != nil {
        return err
    }

    var lines []JournalLine
    for _, pos := range positions {
        rate, err := s.rateRepo.GetSpotRate(ctx, pos.Currency, periodEnd)
        if err != nil {
            return fmt.Errorf("no rate for %s on %s", pos.Currency, periodEnd)
        }

        newKESValue    := pos.FCAmount.Mul(rate.Rate)
        currentKESBook := pos.KESBookValue
        diff           := newKESValue.Sub(currentKESBook)

        if diff.IsZero() {
            continue
        }

        if diff.IsPositive() {
            // Unrealised gain: DR Account, CR Unrealised FX Gain
            lines = append(lines,
                JournalLine{AccountID: pos.AccountID, DebitAmount: diff},
                JournalLine{AccountID: s.unrealisedGainAccountID, CreditAmount: diff},
            )
        } else {
            // Unrealised loss: DR Unrealised FX Loss, CR Account
            lines = append(lines,
                JournalLine{AccountID: s.unrealisedLossAccountID, DebitAmount: diff.Abs()},
                JournalLine{AccountID: pos.AccountID, CreditAmount: diff.Abs()},
            )
        }
    }

    if len(lines) == 0 {
        return nil // nothing to revalue
    }

    entry := &JournalEntry{
        Description: fmt.Sprintf("FX revaluation at %s", periodEnd.Format("02 Jan 2006")),
        Type:        TypeSystem,
        AutoReversalDate: ptr(periodEnd.AddDate(0, 0, 1)), // reverse on first day of next period
        Lines:       lines,
    }
    return s.journalSvc.PostIntegrationEntry(ctx, entry)
}
```

**Auto-reversal:** The revaluation entry is automatically reversed on the first day of the next period. This means the unrealised gain/loss only appears in the closed period's balance sheet. If the position is still open in the next period, it will be revalued again at the new month-end rate.

---

## 11. Realised and Unrealised FX Gains/Losses

### The Distinction

| Type | When | P&L? | Balance Sheet? |
|---|---|---|---|
| Unrealised | At period end, on open FC positions | Configurable (IFRS: OCI; GAAP: P&L) | Yes — FC positions at current rate |
| Realised | When FC transaction settles | Yes — always hits P&L | No — position is closed |

### Realised FX Example

Payable recorded at 128.00 KES/USD; settled when rate is 132.50 KES/USD:

```
Invoice (10,000 USD × 128.00 = 1,280,000 KES):
Dr. Inventory                  1,280,000
    Cr. AP – USD Supplier           1,280,000

Payment (10,000 USD × 132.50 = 1,325,000 KES):
Dr. AP – USD Supplier          1,280,000
Dr. Realised FX Loss              45,000    ← difference goes to P&L
    Cr. Bank Account – KES          1,325,000
```

The realised loss of 45,000 KES is permanent — it reflects the actual extra KES cost of settling the USD obligation at a less favourable rate.

---

## 12. Intercompany Module

### When IC Module Is Needed

The IC module is only relevant when `multi_entity_enabled = true` and at least two organisations within the same tenant transact with each other. It is disabled by default and has zero overhead when disabled.

### Entity Pair Configuration

```go
type ICEntityPair struct {
    TenantID       uuid.UUID
    SourceOrgID    uuid.UUID
    TargetOrgID    uuid.UUID
    // Account pairs for automatic cross-posting
    SourceICReceivableAccountID uuid.UUID
    TargetICPayableAccountID    uuid.UUID
    // Currency pair
    SettlementCurrency string
}
```

When an IC transaction is created in one entity, the system can optionally auto-create the corresponding entry in the counterparty entity (if the IC pair is configured with auto-create enabled).

### IC Transaction Model

```go
type IntercompanyTransaction struct {
    ID                uuid.UUID
    TenantID          uuid.UUID
    SourceOrgID       uuid.UUID
    TargetOrgID       uuid.UUID
    Type              ICTransactionType  // management_fee | loan | goods_transfer | services | dividend
    Amount            decimal.Decimal
    Currency          string
    TransactionDate   time.Time
    SourceJournalID   uuid.UUID    // journal in source org
    TargetJournalID   *uuid.UUID   // journal in target org (nil until recorded)
    ICProfit          *decimal.Decimal  // unrealised profit on goods transfers
    IsEliminated      bool
    Status            ICStatus     // pending | matched | eliminated
}
```

### IC Transaction Types and Their Journals

**Management Fee (Holding → Subsidiary):**

```
HOLDING COMPANY:
Dr. IC Receivable – Sub         500,000
    Cr. Management Fee Revenue      500,000

SUBSIDIARY:
Dr. Management Fee Expense      500,000
    Cr. IC Payable – Holding        500,000
```

**IC Loan:**

```
DRAWDOWN — Holding lends 5,000,000 KES to Subsidiary:

HOLDING:
Dr. IC Loan Receivable – Sub    5,000,000
    Cr. Cash                        5,000,000

SUBSIDIARY:
Dr. Cash                        5,000,000
    Cr. IC Loan Payable – Holding   5,000,000

MONTHLY INTEREST (10% p.a. = 41,667/month):

HOLDING:
Dr. IC Loan Receivable – Sub       41,667
    Cr. Interest Income – IC           41,667

SUBSIDIARY:
Dr. Interest Expense – IC          41,667
    Cr. IC Loan Payable – Holding      41,667
```

**Goods Transfer:**

```
Holding transfers 1,000 units at 400 KES each; cost to Holding was 300 KES/unit:

HOLDING:
Dr. IC Receivable – Sub         400,000
    Cr. COGS                        300,000
    Cr. IC Sales Revenue            100,000  ← IC profit = 100,000

SUBSIDIARY:
Dr. Inventory                   400,000
    Cr. IC Payable – Holding        400,000
```

On consolidation, the IC profit of 100,000 KES is eliminated if the goods remain in the subsidiary's inventory unsold to external customers.

---

## 13. IC Period-End Matching

At each period end, every IC entity pair must confirm that their respective IC balances agree. The IC matching check is part of the period close checklist.

```go
type ICMatchResult struct {
    EntityPair      string
    SourceBalance   decimal.Decimal
    TargetBalance   decimal.Decimal
    Difference      decimal.Decimal
    Currency        string
    IsMatched       bool
    DifferenceType  string  // "fx_movement" | "timing" | "error" | ""
}
```

### Common IC Differences and Resolutions

| Difference Type | Cause | Resolution |
|---|---|---|
| `fx_movement` | Two entities with different functional currencies; IC balance held in different rates | Post FX revaluation in one or both entities so both convert at the same closing rate |
| `timing` | One entity posted the IC transaction in the current period; counterparty in the next | Reopen the lagging period and post the late entry |
| `error` | Amounts were entered differently; coding error | Reverse and re-enter the incorrect entry |

---

## 14. Consolidation Eliminations

When consolidated financial statements are prepared, all intercompany transactions must be eliminated so the consolidated statements show only transactions with external parties.

```go
type EliminationEntry struct {
    ICTransactionID uuid.UUID
    Description     string
    Lines           []JournalLine
}

func (e *Eliminator) GenerateEliminations(ctx context.Context, tenantID uuid.UUID, asAt time.Time) ([]EliminationEntry, error) {
    // 1. Eliminate IC revenue and corresponding expense
    // 2. Eliminate IC loan balances (receivable in one entity, payable in another)
    // 3. Eliminate IC interest income and expense
    // 4. Eliminate IC dividends
    // 5. Eliminate unrealised profit in inventory (goods transfers not yet sold externally)
    ...
}
```

The elimination entries are posted into a "Consolidation" journal type that exists only in the consolidated entity — they do not affect any individual entity's standalone financial statements.

---

## 15. Business Rules & Validation

### Multi-Currency Rules

| Rule ID | Rule | Enforcement |
|---|---|---|
| `FIN-FX-001` | Exchange rate must exist for the transaction date and currency pair | `ValidateAccounts` stage + rate lookup |
| `FIN-FX-002` | An account with `locked_currency` can only be used in entries with that currency | `ValidateAccounts` stage |
| `FIN-FX-003` | Functional currency cannot be changed after the first transaction | Application hard block on `default_currency` update |
| `FIN-FX-004` | Rate deviating more than threshold from prior rate requires confirmation | Rate loading validation |
| `FIN-FX-005` | FX revaluation must be completed before period hard close (if FC balances exist) | Period close checklist |

### Intercompany Rules

| Rule ID | Rule | Enforcement |
|---|---|---|
| `FIN-IC-001` | IC module requires `multi_entity_enabled = true` | Feature flag guard |
| `FIN-IC-002` | An IC transaction must have both a source and a target journal entry before period close | IC matching checklist item |
| `FIN-IC-003` | IC pairs must net to zero on consolidation | Elimination verification |
| `FIN-IC-004` | Unrealised IC profit in transferred goods must be eliminated on consolidation | Elimination engine check |
| `FIN-IC-005` | IC settling payment clears both IC receivable and IC payable | Application check in IC netting |

### Cash Management Rules

| Rule ID | Rule | Enforcement |
|---|---|---|
| `FIN-CASH-001` | Cash position includes only `bank` and `cash` account types | CashPositionService account type filter |
| `FIN-CASH-002` | FC bank accounts are converted at the most recent `spot` rate | Rate lookup in CashAccount hydration |
| `FIN-CASH-003` | Committed items are deducted at face value regardless of exchange rate | CommittedItem amount stored in functional currency |

---

## 16. Performance & Storage

### Cash Position Performance

The cash position query must be fast because it runs frequently (dashboard refresh, every payment decision). It is a primary-key lookup on `account_balances` for all bank/cash accounts in the current period:

```sql
SELECT a.code, a.name, a.locked_currency, ab.net_balance
FROM accounts a
JOIN account_balances ab ON ab.account_id = a.id
WHERE a.tenant_id       = $tenant_id
  AND a.account_type    IN ('bank', 'cash')
  AND a.is_active       = TRUE
  AND ab.period_id      = $current_period_id;
```

This query touches at most ~20 rows for even the largest operator. Target P95: < 10ms.

### Forecast Generation Performance

The 13-week forecast requires:
- 1 cash position query (< 10ms)
- 13 weeks × 2 queries (receipts + payments) = 26 queries against AR/AP aging

Total forecast generation target P95: < 3 seconds. The forecast is computed on demand and the result is cached for up to 30 minutes (configurable via `forecast_cache_ttl`). This is acceptable because the forecast horizon is 13 weeks — minute-level precision is not required.

### Exchange Rate Storage

| Table | Row Size | Rows/Year | Annual Storage |
|---|---|---|---|
| `exchange_rates` | ~200 bytes | 365 days × 5 currencies = 1,825 rows | ~365 KB/year |

Exchange rates are small. 10 years of history for 10 currency pairs is under 10 MB.

### IC Transaction Storage

IC transactions are only relevant for multi-entity tenants. For a typical group with 3–4 entities and 50 IC transactions per month:

| Table | Row Size | Rows/Year | Annual Storage |
|---|---|---|---|
| `ic_transactions` | ~400 bytes | 600 | ~240 KB/year |
| `ic_match_results` | ~300 bytes | 12 (one per month) × entity pairs | ~36 KB/year |

Negligible.

---

## 17. Database Schema

```sql
-- ── exchange_rates ────────────────────────────────────────────────────────────

CREATE TABLE exchange_rates (
    tenant_id      UUID          NOT NULL REFERENCES tenants(id),
    id             UUID          NOT NULL DEFAULT gen_random_uuid(),
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    from_currency  CHAR(3)       NOT NULL,
    to_currency    CHAR(3)       NOT NULL,
    rate           NUMERIC(18,8) NOT NULL CHECK (rate > 0),
    rate_type      TEXT          NOT NULL DEFAULT 'spot'
                       CHECK (rate_type IN ('spot','average','historical','budget')),
    rate_date      DATE          NOT NULL,
    source         TEXT          NOT NULL DEFAULT 'manual',
    loaded_by      UUID,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, from_currency, to_currency, rate_type, rate_date)
);

CREATE INDEX idx_rates_lookup
    ON exchange_rates (tenant_id, from_currency, to_currency, rate_type, rate_date DESC);

ALTER TABLE exchange_rates ENABLE ROW LEVEL SECURITY;
ALTER TABLE exchange_rates FORCE  ROW LEVEL SECURITY;
CREATE POLICY er_app ON exchange_rates FOR ALL TO application_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

-- ── ic_transactions ───────────────────────────────────────────────────────────

CREATE TABLE ic_transactions (
    tenant_id          UUID          NOT NULL REFERENCES tenants(id),
    id                 UUID          NOT NULL DEFAULT gen_random_uuid(),
    created_at         TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    source_org_id      UUID          NOT NULL,
    target_org_id      UUID          NOT NULL,
    type               TEXT          NOT NULL
                           CHECK (type IN ('management_fee','loan','goods_transfer','services','dividend')),
    amount             NUMERIC(18,4) NOT NULL CHECK (amount > 0),
    currency           CHAR(3)       NOT NULL,
    transaction_date   DATE          NOT NULL,
    source_journal_id  UUID          NOT NULL REFERENCES journal_entries(id),
    target_journal_id  UUID          REFERENCES journal_entries(id),
    ic_profit          NUMERIC(18,4),
    is_eliminated      BOOLEAN       NOT NULL DEFAULT FALSE,
    status             TEXT          NOT NULL DEFAULT 'pending'
                           CHECK (status IN ('pending','matched','eliminated')),

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id),
    CONSTRAINT chk_different_orgs CHECK (source_org_id <> target_org_id)
);

ALTER TABLE ic_transactions ENABLE ROW LEVEL SECURITY;
ALTER TABLE ic_transactions FORCE  ROW LEVEL SECURITY;
CREATE POLICY ic_app ON ic_transactions FOR ALL TO application_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

-- ── cash_forecast_cache ───────────────────────────────────────────────────────
-- Caches computed 13-week forecasts to avoid re-computing on every request.

CREATE TABLE cash_forecast_cache (
    tenant_id       UUID        NOT NULL,
    organisation_id UUID        NOT NULL,
    computed_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    valid_until     TIMESTAMPTZ NOT NULL,
    forecast_data   JSONB       NOT NULL,

    PRIMARY KEY (tenant_id, organisation_id)
);
```

---

## 18. API Reference

### Cash Management

```
GET    /finance/cash/position                Real-time cash position across all bank accounts
GET    /finance/cash/forecast                13-week rolling forecast (?force_refresh=true to bypass cache)
GET    /finance/cash/committed               List committed items (approved payment runs, upcoming payroll)
```

**Cash position response:**
```json
{
  "as_at": "2025-02-06T09:00:00Z",
  "total_cash_kes": 3627500,
  "available_cash_kes": 1320100,
  "accounts": [
    {
      "code": "1112",
      "name": "Checking – Main",
      "currency": "KES",
      "balance": 1240000,
      "functional_balance": 1240000,
      "exchange_rate": 1.0
    },
    {
      "code": "1130",
      "name": "USD Account",
      "currency": "USD",
      "balance": 15000,
      "functional_balance": 1987500,
      "exchange_rate": 132.5,
      "rate_date": "2025-02-06"
    }
  ]
}
```

### Exchange Rates

```
GET    /finance/rates                        List rates (?from_currency=, ?to_currency=, ?date=, ?rate_type=)
POST   /finance/rates                        Load rates (batch)
GET    /finance/rates/latest                 Most recent rate for each currency pair
```

### Intercompany

```
GET    /finance/ic/transactions              List IC transactions
POST   /finance/ic/transactions              Create IC transaction
GET    /finance/ic/balances                  IC balance status by entity pair (?as_at=)
POST   /finance/ic/match                     Run IC matching for a period (?period_id=)
POST   /finance/ic/eliminate                 Generate elimination entries for consolidation (?as_at=)
GET    /finance/ic/eliminations              List generated elimination entries
```

---

## 19. Feature Flags & Configuration

| Flag | Type | Default | Description |
|---|---|---|---|
| `multi_currency_enabled` | bool | `true` | Enable multi-currency transaction support. Even for KES-only operators, this should remain true — it has zero overhead unless non-KES accounts are created. |
| `fx_revaluation_required_for_close` | bool | `true` | Block period hard-close if unrevalued FC balances exist |
| `rate_deviation_alert_threshold` | decimal | `0.05` | Alert if new rate deviates > 5% from prior rate for same pair |
| `rate_auto_load_enabled` | bool | `false` | Enable automatic rate loading from a configured API feed |
| `rate_auto_load_source` | string | `null` | URL of rate API feed (e.g. Central Bank of Kenya open data endpoint) |
| `rate_auto_load_cron` | string | `0 8 * * *` | Cron expression for automatic rate loading (default: 8am daily) |
| `unrealised_fx_treatment` | enum | `pnl` | `pnl` — unrealised FX goes to P&L; `oci` — goes to Other Comprehensive Income (IFRS preferred for certain instruments) |
| `cash_forecast_enabled` | bool | `false` | Enable 13-week rolling forecast engine |
| `cash_forecast_cache_ttl` | int | `30` | Minutes to cache forecast result before recomputing |
| `cash_target_minimum_balance` | decimal | `0` | Minimum operating cash policy amount (KES). Forecast highlights weeks below this. |
| `multi_entity_enabled` | bool | `false` | Enable multiple organisations per tenant |
| `intercompany_enabled` | bool | `false` | Enable IC transactions. Requires `multi_entity_enabled = true`. |
| `ic_matching_required_for_close` | bool | `true` | IC balances must be matched before period close. Only relevant if `intercompany_enabled = true`. |
| `ic_auto_create_counterparty_entry` | bool | `false` | Automatically create the counterparty IC entry when an IC transaction is created in the source entity |

---

## 20. v1.0 Rollout Assessment

### Must Have at v1.0

- Multi-currency data model (transaction + functional currency on all journal lines) — required even for KES-only operators because removing it later is a schema migration
- Exchange rate table and rate lookup
- Rate validation on transaction creation (block if no rate found)
- Cash position query (simple, essential for any financial user)
- FX revaluation service (can run manually; Temporal cron deferred)

### Can Be Deferred

| Feature | Suggested Release |
|---|---|
| 13-week rolling forecast | v1.1 — cash position covers immediate need |
| Automatic rate loading from API feed | v1.1 — manual rate entry is acceptable at v1.0 |
| Intercompany module | v1.2 — only relevant for multi-entity operators |
| Realised vs. unrealised FX segregation (OCI treatment) | v1.1 — P&L treatment is acceptable default |
| Cash forecast caching | v1.1 (prerequisite: forecast engine) |
| Rate deviation alert with confirmation UI | v1.1 — soft warning logged is sufficient at v1.0 |

### Never Defer

- Dual-amount storage on `journal_lines` (`currency`, `fx_rate`, `base_amount`) — retrofitting this after v1.0 requires a schema migration on the largest table in the system
- Rate uniqueness constraint — prevents duplicate rates for the same currency pair and date
- `locked_currency` enforcement in the pipeline — without it, a KES entry can be posted to a USD bank account

---

*End of FILE-07. Proceeding to FILE-08: Financial Reporting Engine & Report Catalogue.*
