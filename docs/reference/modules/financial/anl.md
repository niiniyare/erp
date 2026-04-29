# ERP Analytics Module — Complete Engineering Documentation

> Comprehensive reference for every algorithm, formula, reporting pattern, and infrastructure component in the AWO ERP analytics layer. Covers deterministic computation, statistical detection, rule engines, financial statements, KPI frameworks, and the reporting platform.

---

## Table of Contents

- [Part 1 — Phase 1: Deterministic Formulas](#part-1--phase-1-deterministic-formulas)
  - [1.1 Double-Entry Validation](#11-double-entry-validation)
  - [1.2 Payroll Calculations](#12-payroll-calculations)
  - [1.3 VAT Calculation & Return](#13-vat-calculation--return)
  - [1.4 Inventory Valuation](#14-inventory-valuation)
  - [1.5 Dip Variance](#15-dip-variance)
  - [1.6 AR Aging & Provision](#16-ar-aging--provision)
- [Part 2 — Phase 2: Statistical Thresholds](#part-2--phase-2-statistical-thresholds)
  - [2.1 Z-Score Anomaly Detection](#21-z-score-anomaly-detection)
  - [2.2 IQR Anomaly Detection](#22-iqr-anomaly-detection)
  - [2.3 Benford's Law Fraud Detection](#23-benfords-law-fraud-detection)
  - [2.4 Reconciliation Matching Score](#24-reconciliation-matching-score)
- [Part 3 — Phase 3: Rule Engines](#part-3--phase-3-rule-engines)
  - [3.1 Approval Routing](#31-approval-routing)
  - [3.2 Reorder Point Alerts](#32-reorder-point-alerts)
  - [3.3 Credit Limit Enforcement](#33-credit-limit-enforcement)
- [Part 4 — Financial Reporting Layer](#part-4--financial-reporting-layer)
  - [4.1 Period & Fiscal Year Management](#41-period--fiscal-year-management)
  - [4.2 Profit & Loss Statement](#42-profit--loss-statement)
  - [4.3 Balance Sheet](#43-balance-sheet)
  - [4.4 Cash Flow Statement](#44-cash-flow-statement)
  - [4.5 Budget vs. Actual](#45-budget-vs-actual)
  - [4.6 Comparative Reporting](#46-comparative-reporting)
- [Part 5 — Analytics Dimensions](#part-5--analytics-dimensions)
  - [5.1 KPI Framework](#51-kpi-framework)
  - [5.2 Cost Centre & Department Analytics](#52-cost-centre--department-analytics)
  - [5.3 Overhead Allocation](#53-overhead-allocation)
  - [5.4 Inventory Analytics](#54-inventory-analytics)
  - [5.5 Receivables & Payables Analytics](#55-receivables--payables-analytics)
  - [5.6 Profitability Analytics](#56-profitability-analytics)
- [Part 6 — Reporting Platform Infrastructure](#part-6--reporting-platform-infrastructure)
  - [6.1 Report Definition Schema](#61-report-definition-schema)
  - [6.2 Pre-Aggregation Strategy](#62-pre-aggregation-strategy)
  - [6.3 Drill-Down Architecture](#63-drill-down-architecture)
  - [6.4 Multi-Dimensional Filtering](#64-multi-dimensional-filtering)
  - [6.5 Scheduled Reports & Distribution](#65-scheduled-reports--distribution)
  - [6.6 Export Engine](#66-export-engine)
- [Part 7 — Audit & Activity Analytics](#part-7--audit--activity-analytics)
  - [7.1 GL Change Log](#71-gl-change-log)
  - [7.2 Reconciliation History](#72-reconciliation-history)
  - [7.3 User Activity Analytics](#73-user-activity-analytics)
- [Part 8 — Summary & Build Priority](#part-8--summary--build-priority)

---

# Part 1 — Phase 1: Deterministic Formulas

Phase 1 algorithms are pure functions. Given the same inputs they always produce the same outputs. No training data, no external services, no probabilistic reasoning. They are the financial foundation every other layer depends on.

**Core principle:** Use `shopspring/decimal` throughout — never `float64` for financial arithmetic. Float accumulation errors compound across thousands of transactions and produce incorrect financial statements.

---

## 1.1 Double-Entry Validation

### Use Case

Every financial transaction must carry equal debits and credits. This is the primary integrity constraint on all financial data — the equivalent of a database foreign key at the accounting level. A violated balance means the general ledger is corrupt and every financial statement derived from it is wrong.

### How Other Systems Use It

**SAP FI** validates at document level before persistence. A document cannot be saved unless debit sum equals credit sum — error `F5263` is returned otherwise. **QuickBooks** enforces this silently, auto-generating the offsetting entry and hiding the mechanics from the user. **Xero** exposes it via API: `POST /Journals` with unequal amounts returns `ValidationException: Journal must balance`. **Oracle Financials** validates at subledger level (AP, AR, Inventory) before transferring to the GL — imbalance is caught at source, not at the ledger.

### Data Required

```sql
journal_entry_lines:
  id UUID, journal_id UUID, account_id UUID,
  debit NUMERIC(19,4), credit NUMERIC(19,4),
  tenant_id UUID, posted_at TIMESTAMPTZ
```

### Performance & Optimisation

Runs on every transaction — must be sub-millisecond. Aggregate over the in-memory line slice already fetched for the document; never a separate database round-trip. Partial index on `journal_id WHERE posted_at IS NULL` for pre-post validation queries.

### SQL

```sql
-- Reject if this returns a row
SELECT journal_id,
       ABS(SUM(debit) - SUM(credit)) AS imbalance
FROM   journal_entry_lines
WHERE  journal_id = $1 AND tenant_id = $2
GROUP  BY journal_id
HAVING ABS(SUM(debit) - SUM(credit)) > 0.005;
```

### Go

```go
func ValidateBalance(lines []JournalLine) error {
    var net decimal.Decimal
    for _, l := range lines {
        net = net.Add(l.Debit).Sub(l.Credit)
    }
    if net.Abs().GreaterThan(decimal.NewFromFloat(0.005)) {
        return fmt.Errorf("journal imbalance: net %s", net)
    }
    return nil
}
```

---

## 1.2 Payroll Calculations

### Use Case

Computing gross-to-net pay including all Kenya statutory deductions: PAYE, NSSF (2023 Act Tier I + II), SHIF, and Housing Levy. Every formula is legally mandated — errors result in statutory penalties. Rate tables must be versioned and period-effective so historical payslips remain reproducible after rate changes.

### How Other Systems Use It

**Sage Payroll Kenya** maintains a versioned rate table updated on every regulatory change. The engine is a pure function: `compute_net(employee, period, rate_snapshot) → payslip`. The snapshot used for each run is stored alongside the payslip. **Workday** models each deduction as a versioned "Calculated Field" with effective dates — older runs use the rates that were effective at the time, not current rates.

### Data Required

```sql
statutory_rates:
  id UUID, effective_from DATE, effective_to DATE,
  nssf_tier1_rate NUMERIC, nssf_tier1_ceiling NUMERIC,
  nssf_tier2_rate NUMERIC, nssf_tier2_ceiling NUMERIC,
  shif_rate NUMERIC, housing_levy_rate NUMERIC,
  paye_bands JSONB,   -- [{from, to, rate}]
  personal_relief NUMERIC,
  tenant_id UUID      -- NULL = system-wide

payroll_lines:
  id, period_id, employee_id, gross_pay, taxable_pay,
  paye, nssf_employee, shif, housing_levy_employee,
  net_pay, rate_snapshot_id, tenant_id
```

### Performance & Optimisation

500 employees must complete in under 30 seconds. Each employee's calculation is independent — use a goroutine worker pool. Pre-load the rate snapshot and all department data once before the pool starts. Write all results in a single database transaction — partial payroll runs are never committed.

### SQL

```sql
-- Rate snapshot for a given payroll date
SELECT * FROM statutory_rates
WHERE  (tenant_id = $1 OR tenant_id IS NULL)
  AND  effective_from <= $2
  AND  (effective_to IS NULL OR effective_to >= $2)
ORDER  BY tenant_id NULLS LAST
LIMIT  1;
```

### Go

```go
func ComputePAYE(taxable decimal.Decimal, bands []TaxBand, relief decimal.Decimal) decimal.Decimal {
    tax, remaining := decimal.Zero, taxable
    for _, b := range bands {
        if remaining.IsZero() { break }
        taxable := decimal.Min(remaining, b.To.Sub(b.From))
        tax = tax.Add(taxable.Mul(b.Rate))
        remaining = remaining.Sub(taxable)
    }
    if remaining.IsPositive() {
        tax = tax.Add(remaining.Mul(bands[len(bands)-1].Rate))
    }
    return decimal.Max(decimal.Zero, tax.Sub(relief))
}

func ComputeNSSF(gross decimal.Decimal, r NSSFRates) decimal.Decimal {
    t1 := decimal.Min(gross, r.Tier1Ceiling).Mul(r.Tier1Rate)
    base := decimal.Max(decimal.Zero, gross.Sub(r.Tier1Ceiling))
    t2 := decimal.Min(base, r.Tier2Ceiling.Sub(r.Tier1Ceiling)).Mul(r.Tier2Rate)
    return t1.Add(t2)
}
```

---

## 1.3 VAT Calculation & Return

### Use Case

Computing output VAT on sales, input VAT on purchases, and net VAT payable per period. Must handle multiple rate categories (16% standard, 0% zero-rated, exempt), store VAT at line level (a single invoice can span categories), and produce a KRA eTIMS-compatible return.

### How Other Systems Use It

**Sage 50 Kenya** maintains a tax code table where each code has a rate, direction, and GL account mapping. Every transaction line is assigned a tax code. **Xero** computes the VAT return by querying all transactions in the period, grouping by tax rate, and computing box values. Xero supports cash basis (tax point = payment date) vs. invoice basis (tax point = invoice date) — a critical design decision. **SAP** validates the tax code configuration against the tax authority's published code list, rejecting unknown codes at entry time.

### Data Required

```sql
vat_rates:
  id, code, name, rate, rate_type (STANDARD/ZERO/EXEMPT),
  direction (SALES/PURCHASE), gl_account_id, tenant_id

transaction_lines:
  id, transaction_id, net_amount, vat_rate_id,
  vat_amount, gross_amount, tax_point_date, tenant_id
```

### Performance & Optimisation

Use a materialised view refreshed nightly for the VAT return dashboard. The on-demand return generation queries live data. Partial index on `tax_point_date` for all non-exempt lines. Never store computed VAT return figures as the source of truth — always rederive from transaction lines.

### SQL

```sql
-- VAT return summary
SELECT
    SUM(CASE WHEN vr.direction='SALES'    THEN tl.vat_amount ELSE 0 END) AS output_vat,
    SUM(CASE WHEN vr.direction='PURCHASE' THEN tl.vat_amount ELSE 0 END) AS input_vat,
    SUM(CASE WHEN vr.direction='SALES'    THEN tl.vat_amount ELSE 0 END) -
    SUM(CASE WHEN vr.direction='PURCHASE' THEN tl.vat_amount ELSE 0 END) AS net_payable
FROM   transaction_lines tl
JOIN   vat_rates vr   ON vr.id = tl.vat_rate_id
JOIN   transactions t ON t.id  = tl.transaction_id
WHERE  t.tenant_id = $1 AND tl.tax_point_date BETWEEN $2 AND $3
  AND  t.status = 'POSTED' AND vr.rate_type != 'EXEMPT';
```

### Go

```go
func ComputeVAT(amount, rate decimal.Decimal, inclusive bool) VATResult {
    if inclusive {
        net := amount.Div(decimal.NewFromInt(1).Add(rate)).Round(2)
        return VATResult{Net: net, VAT: amount.Sub(net), Gross: amount}
    }
    vat := amount.Mul(rate).Round(2)
    return VATResult{Net: amount, VAT: vat, Gross: amount.Add(vat)}
}
```

---

## 1.4 Inventory Valuation

### Use Case

Computing cost of inventory on hand and COGS as items are received and issued. FIFO tracks discrete cost layers consumed in date order. Weighted Average Cost (WAC) recalculates a blended unit cost on every receipt. For fuel retail, WAC is standard because fuel from multiple deliveries physically mixes in the tank.

### How Other Systems Use It

**SAP MM** implements WAC as "Moving Average Price" — the default for trading goods. FIFO is managed via batch management with explicit lot-date tracking. **Odoo** stores WAC in `standard_price` on the product, updated on every goods receipt. FIFO uses "stock valuation layers" with explicit lot tracking. **Microsoft Dynamics 365** implements FIFO through "inventory layers" reconciled during a periodic "inventory close" that adjusts interim standard costs to actual FIFO costs.

### Data Required

```sql
inventory_items:
  id, sku, valuation_method (FIFO/WAC),
  current_quantity, current_avg_cost, tenant_id

inventory_layers:   -- FIFO only
  id, item_id, receipt_id, received_quantity,
  remaining_quantity, unit_cost, received_at, tenant_id
```

### Performance & Optimisation

FIFO layer consumption requires a `FOR UPDATE` lock on layers — critical for concurrent sales at the same pump. For high-volume items, use deferred FIFO costing: record the issue at current WAC as a preliminary cost, resolve the true FIFO cost in a background job, post the variance. Index: `inventory_layers(item_id, received_at) WHERE remaining_quantity > 0`.

### SQL

```sql
-- WAC update on receipt
UPDATE inventory_items SET
    current_avg_cost = (current_quantity * current_avg_cost + $2 * $3)
                        / NULLIF(current_quantity + $2, 0),
    current_quantity = current_quantity + $2
WHERE id = $1 AND tenant_id = $4
RETURNING current_avg_cost;
```

### Go

```go
func ApplyFIFO(layers []InventoryLayer, qty decimal.Decimal) (decimal.Decimal, []InventoryLayer, error) {
    cost, remaining := decimal.Zero, qty
    for i := range layers {
        if remaining.IsZero() { break }
        consume := decimal.Min(remaining, layers[i].RemainingQty)
        cost = cost.Add(consume.Mul(layers[i].UnitCost))
        layers[i].RemainingQty = layers[i].RemainingQty.Sub(consume)
        remaining = remaining.Sub(consume)
    }
    if remaining.IsPositive() {
        return decimal.Zero, nil, fmt.Errorf("insufficient stock: %s units short", remaining)
    }
    return cost, layers, nil
}
```

---

## 1.5 Dip Variance

### Use Case

Reconciling theoretical fuel stock (opening dip + deliveries − meter sales) against the physical closing dip reading per tank per shift. Normal variance arises from temperature expansion and measurement imprecision. Abnormal variance indicates leakage, meter calibration errors, or fraud.

### How Other Systems Use It

**Orpak** performs dip reconciliation per shift and automatically posts variance as a shrinkage or gain entry in the GL. **Implant** adds temperature correction — raw volumes are corrected to 15°C before comparison, eliminating thermal expansion as a driver. **Shell Retail Management** uses tolerance bands by product (diesel ±0.3%, petrol ±0.5%) and distinguishes auto-postable normal shrinkage from out-of-tolerance variances requiring investigation.

### Data Required

```sql
dip_readings: id, tank_id, read_at, volume_litres, shift_id, tenant_id
deliveries:   id, tank_id, delivered_litres, shift_id, tenant_id
pump_meter_readings: id, tank_id, volume_dispensed, shift_id, tenant_id
tank_tolerance_rules: id, tank_id, tolerance_pct, tenant_id
```

### SQL

```sql
WITH
  o AS (SELECT volume_litres FROM dip_readings WHERE tank_id=$1 AND shift_id=$2 ORDER BY read_at ASC  LIMIT 1),
  c AS (SELECT volume_litres FROM dip_readings WHERE tank_id=$1 AND shift_id=$2 ORDER BY read_at DESC LIMIT 1),
  d AS (SELECT COALESCE(SUM(delivered_litres),0) AS v FROM deliveries        WHERE tank_id=$1 AND shift_id=$2),
  s AS (SELECT COALESCE(SUM(volume_dispensed),0) AS v FROM pump_meter_readings WHERE tank_id=$1 AND shift_id=$2)
SELECT
    o.volume_litres                                   AS opening,
    d.v                                               AS deliveries,
    s.v                                               AS sales,
    (o.volume_litres + d.v - s.v)                     AS theoretical,
    c.volume_litres                                   AS actual,
    (o.volume_litres + d.v - s.v) - c.volume_litres   AS variance_litres,
    ROUND(((o.volume_litres + d.v - s.v) - c.volume_litres)
          / NULLIF(o.volume_litres + d.v - s.v, 0) * 100, 3) AS variance_pct
FROM o, c, d, s;
```

### Go

```go
func ComputeDipVariance(opening, deliveries, sales, closing, tolerance decimal.Decimal) DipVarianceResult {
    theoretical := opening.Add(deliveries).Sub(sales)
    variance    := theoretical.Sub(closing)
    pct := decimal.Zero
    if theoretical.IsPositive() {
        pct = variance.Div(theoretical).Mul(decimal.NewFromInt(100))
    }
    return DipVarianceResult{
        Theoretical: theoretical, Actual: closing,
        VarianceLitres: variance, VariancePct: pct,
        WithinTolerance: pct.Abs().LessThanOrEqual(tolerance),
    }
}
```

---

## 1.6 AR Aging & Provision

### Use Case

Classifying outstanding customer invoices by days overdue relative to due date and computing a bad debt provision using configurable rates per aging bucket. Required by IFRS 9 (Expected Credit Loss model) for balance sheet presentation of net realisable AR.

### How Other Systems Use It

**Sage Business Cloud** runs aging on demand and applies configurable provision rates per bucket. **SAP AR** ages relative to due date (not invoice date) and distinguishes "not yet due" from "0–30 days overdue". **Oracle Receivables** supports three aging methods: by due date, by invoice date, and by GL date — due date is the operationally correct choice for collections. **Xero** shows aging but does not compute provisions — leaving a compliance gap most SME users do not fill.

### Data Required

```sql
invoices: id, customer_id, due_date, amount_outstanding, status, tenant_id
provision_rates: id, bucket_label, days_from, days_to, provision_rate, tenant_id
```

### Performance & Optimisation

Critical composite index: `invoices(tenant_id, status, due_date)`. Materialise the aging bucket summary nightly as a view; run the invoice-level detail query on demand. Never age by invoice date — always by due date.

### SQL

```sql
SELECT
    c.name,
    i.amount_outstanding,
    CURRENT_DATE - i.due_date AS days_overdue,
    CASE
        WHEN CURRENT_DATE <= i.due_date                    THEN 'CURRENT'
        WHEN CURRENT_DATE - i.due_date BETWEEN 1   AND 30  THEN '1-30'
        WHEN CURRENT_DATE - i.due_date BETWEEN 31  AND 60  THEN '31-60'
        WHEN CURRENT_DATE - i.due_date BETWEEN 61  AND 90  THEN '61-90'
        WHEN CURRENT_DATE - i.due_date BETWEEN 91  AND 120 THEN '91-120'
        ELSE '120+'
    END AS bucket
FROM invoices i JOIN customers c ON c.id = i.customer_id
WHERE i.tenant_id = $1 AND i.status IN ('OPEN','PARTIAL')
ORDER BY days_overdue DESC;
```

### Go

```go
func AgeBucket(due, asOf time.Time) string {
    d := int(asOf.Sub(due).Hours() / 24)
    switch { case d<=0: return "CURRENT"; case d<=30: return "1-30"
             case d<=60: return "31-60";  case d<=90: return "61-90"
             case d<=120: return "91-120"; default: return "120+" }
}
```

---

# Part 2 — Phase 2: Statistical Thresholds

Phase 2 builds a statistical understanding of "normal" from historical data and flags deviations. No ML training — just math on accumulated history. Every flag is explainable in one sentence referencing the specific statistical basis.

---

## 2.1 Z-Score Anomaly Detection

### Use Case

Flagging transaction amounts that deviate significantly from the historical norm for that account, user role, and transaction type combination. Z-score measures standard deviations from the mean of the reference population.

### How Other Systems Use It

**Concur Expense** computes per-employee, per-category baselines and flags submissions exceeding 2.5σ as "policy alerts" requiring additional review. **HSBC fraud detection** uses Z-score as a first-pass filter before ML — transactions within 3σ of the customer's normal distribution auto-clear, flagged ones proceed to ML scoring. The Z-score filter handles ~85% of volume, keeping ML focused on genuine edge cases. **NetSuite** does not compute Z-scores natively — most implementations add this via a BI layer.

### Data Required

```sql
transaction_baselines:
  id, tenant_id, dimension_type, dimension_key,
  period_days, sample_count,
  mean_amount, stddev_amount, computed_at
  -- Refreshed nightly; never read at transaction time from raw data
```

### Performance & Optimisation

Pre-compute baselines nightly. Runtime check = one indexed lookup + one arithmetic operation. Minimum 30 samples before computing — below this the estimate is unreliable and generates false positives.

### SQL

```sql
-- Nightly baseline computation
INSERT INTO transaction_baselines
    (tenant_id, dimension_type, dimension_key, period_days,
     sample_count, mean_amount, stddev_amount, computed_at)
SELECT tenant_id, 'ACCOUNT', account_id::text, 90,
       COUNT(*), AVG(debit+credit), STDDEV_POP(debit+credit), NOW()
FROM   journal_entry_lines jel
JOIN   journals j ON j.id = jel.journal_id
WHERE  j.posted_at >= NOW() - INTERVAL '90 days' AND j.status = 'POSTED'
GROUP  BY tenant_id, account_id
HAVING COUNT(*) >= 30
ON CONFLICT (tenant_id, dimension_type, dimension_key, period_days)
DO UPDATE SET mean_amount=EXCLUDED.mean_amount,
              stddev_amount=EXCLUDED.stddev_amount,
              sample_count=EXCLUDED.sample_count,
              computed_at=EXCLUDED.computed_at;
```

### Go

```go
func ZScore(value, mean, stddev decimal.Decimal, threshold float64) (float64, bool) {
    if stddev.IsZero() { return 0, false }
    z, _ := value.Sub(mean).Abs().Div(stddev).Float64()
    return z, z > threshold
}
```

---

## 2.2 IQR Anomaly Detection

### Use Case

An alternative to Z-score robust against heavy-tailed distributions. Financial data almost always has heavy tails — occasional legitimate very large transactions inflate the standard deviation and make Z-score permissive. IQR uses only the middle 50% of data to define "normal", so extreme-but-legitimate transactions do not corrupt the baseline.

**Use Z-score when:** distribution is reasonably bell-shaped.
**Use IQR when:** data has heavy tails or you want robustness to large-but-legitimate outliers.

### How Other Systems Use It

**Palantir Foundry** uses Tukey's fences (IQR-based) as the default outlier method for financial dashboards, citing robustness over Z-score. **Tableau** includes IQR in its "Mark Outliers" analytics feature, specifically noting it is preferred for right-skewed distributions — which describes most financial data. **ACL/Arbutus** uses IQR-based outlier detection as a standard AP fraud test.

### Data Required

Same baseline table as Z-score, adding quartile columns:

```sql
-- Additional columns on transaction_baselines
q1_amount NUMERIC(19,4), q3_amount NUMERIC(19,4),
lower_fence NUMERIC(19,4), upper_fence NUMERIC(19,4)
-- lower_fence = q1 - 1.5*IQR,  upper_fence = q3 + 1.5*IQR
```

### SQL

```sql
SELECT
    percentile_cont(0.25) WITHIN GROUP (ORDER BY amount) AS q1,
    percentile_cont(0.75) WITHIN GROUP (ORDER BY amount) AS q3
FROM (
    SELECT (jel.debit + jel.credit) AS amount
    FROM   journal_entry_lines jel
    JOIN   journals j ON j.id = jel.journal_id
    WHERE  j.posted_at >= NOW() - INTERVAL '90 days'
      AND  jel.account_id = $1 AND jel.tenant_id = $2
) t HAVING COUNT(*) >= 30;
```

### Go

```go
func CheckIQR(amount decimal.Decimal, b IQRBaseline) (bool, string) {
    if amount.LessThan(b.LowerFence) {
        return true, fmt.Sprintf("%.2f is below lower fence (%.2f)", amount, b.LowerFence)
    }
    if amount.GreaterThan(b.UpperFence) {
        return true, fmt.Sprintf("%.2f exceeds upper fence (%.2f)", amount, b.UpperFence)
    }
    return false, ""
}
```

---

## 2.3 Benford's Law Fraud Detection

### Use Case

Detecting fabricated or manipulated numbers in financial datasets by comparing the actual leading-digit distribution against Benford's expected logarithmic distribution. Humans inventing numbers tend to distribute leading digits too uniformly — statistically detectable via chi-squared test. Most powerful on expense claims, supplier invoices, purchase orders, and payroll figures.

**Limitation:** Does not work on constrained datasets (items with fixed price ranges), sequential numbers (invoice IDs), or datasets with fewer than ~500 records.

### How Other Systems Use It

**ACL Analytics / Galvanize** includes Benford analysis as a standard audit test — first-digit, first-two-digits, and last-two-digits distributions. **KPMG's audit data analytics platform** applies Benford as a first-pass filter on every GL dataset. **The IRS** uses Benford's Law in automated screening of tax returns — research confirms correlation between Benford deviations and audit adjustments.

### SQL

```sql
WITH counts AS (
    SELECT LEFT(CAST(ABS(amount) AS TEXT),1)::int AS d,
           COUNT(*)::float AS n
    FROM   expense_claims
    WHERE  tenant_id = $1 AND amount > 0
      AND  submitted_at >= NOW() - INTERVAL '12 months'
    GROUP  BY 1
),
expected AS (
    SELECT g AS d, LOG(1 + 1.0/g)/LOG(10)*100 AS exp_pct
    FROM   generate_series(1,9) g
)
SELECT c.d,
       ROUND(c.n/(SELECT SUM(n) FROM counts)*100,2) AS actual_pct,
       ROUND(e.exp_pct,2)                            AS expected_pct,
       ROUND(c.n/(SELECT SUM(n) FROM counts)*100 - e.exp_pct, 2) AS deviation
FROM counts c JOIN expected e USING(d)
ORDER BY d;
```

### Go

```go
// Chi-squared > 15.51 (df=8, p=0.05) indicates significant deviation
func ChiSquaredBenford(observed map[int]int, total int) float64 {
    chi2 := 0.0
    for d := 1; d <= 9; d++ {
        exp := math.Log10(1+1.0/float64(d)) * float64(total)
        diff := float64(observed[d]) - exp
        chi2 += (diff * diff) / exp
    }
    return chi2
}
```

---

## 2.4 Reconciliation Matching Score

### Use Case

Automatically matching bank statement lines to GL entries during bank reconciliation using a weighted composite score across amount match, date proximity, and reference similarity. Auto-match above the high-confidence threshold; present top candidates for manual selection below it.

### How Other Systems Use It

**Xero** requires exact amount match but uses fuzzy date and description matching. It auto-matches single unambiguous candidates and presents the top 3 for ambiguous cases. **SAP Bank Communication Management** uses a configurable-weight scoring engine — implementations tune weights for their bank's reference format. **QuickBooks** learns from historical accountant matching decisions and begins suggesting mappings — the boundary between Phase 2 deterministic scoring and Phase 4 ML categorization.

### Data Required

```sql
bank_statement_lines: id, line_date, description, reference, amount, tenant_id
gl_entries_unreconciled: id, entry_date, description, reference, amount, tenant_id
```

### Performance & Optimisation

Filter GL candidates first using amount tolerance and date window — reduces candidate set from thousands to 1–5 per bank line before scoring. Scoring itself is then trivially fast.

### SQL

```sql
SELECT gl.id, gl.amount, gl.entry_date,
    CASE WHEN ABS(gl.amount-$2)=0 THEN 1.0
         WHEN ABS(gl.amount-$2)<=$3 THEN 1.0-(ABS(gl.amount-$2)/$3::numeric)
         ELSE 0.0 END AS amount_score,
    GREATEST(0, 1.0 - ABS(gl.entry_date-$4::date)::numeric/7.0) AS date_score
FROM gl_entries_unreconciled gl
WHERE gl.tenant_id=$1 AND ABS(gl.amount-$2)<=$3
  AND gl.entry_date BETWEEN $4::date-7 AND $4::date+7
  AND gl.reconciled_at IS NULL
ORDER BY amount_score DESC, date_score DESC LIMIT 5;
```

### Go

```go
func MatchScore(bank BankLine, gl GLEntry, tol decimal.Decimal, w Weights) float64 {
    diff := bank.Amount.Sub(gl.Amount).Abs()
    amtScore := 0.0
    if diff.IsZero() { amtScore = 1.0 } else if diff.LessThanOrEqual(tol) {
        r, _ := diff.Div(tol).Float64(); amtScore = 1.0 - r
    }
    daysDiff := math.Abs(bank.Date.Sub(gl.Date).Hours() / 24)
    dateScore := math.Max(0, 1.0-daysDiff/7.0)
    refScore  := stringSimilarity(bank.Reference, gl.Reference)
    return w.Amount*amtScore + w.Date*dateScore + w.Reference*refScore
}
```

---

# Part 3 — Phase 3: Rule Engines

Phase 3 encodes business policy as explicit, configurable rules maintained by business administrators — not developers. Every rule must be readable by a non-developer and auditable by an external reviewer.

---

## 3.1 Approval Routing

### Use Case

Determining the ordered approval chain for any transaction (purchase requisition, payment, journal entry, expense claim) based on configurable rules: amount thresholds, transaction type, cost centre, vendor category, and user role. Approval workflows enforce segregation of duties and purchase authorisation framework.

### How Other Systems Use It

**SAP Workflow** uses "Workflow Tasks" linked to business objects. Each task specifies an agent determined at runtime by a rule evaluating transaction attributes. **Coupa** uses an "Approval Chain" model — each step has an approver type and condition configured by business administrators via UI. **Microsoft Power Automate** models approval flows as visual diagrams with sequential stages evaluated at runtime.

### Data Required

```sql
approval_policies: id, name, entity_type, is_active, tenant_id
approval_rules:
  id, policy_id, sequence_number,
  condition_field, condition_op (GT/LT/EQ/IN),
  condition_value TEXT, approver_type, approver_ref,
  escalation_days INT, tenant_id
approval_instances: id, policy_id, entity_id, current_step, status, tenant_id
approval_steps: id, instance_id, step_number, approver_id, status, decided_at, tenant_id
```

### Performance & Optimisation

Cache active policy + rules in Redis (TTL 5 minutes, invalidate on any policy update). For batch submissions (period-end accruals), load the policy once and evaluate all transactions against the in-memory rule set. Rules evaluated in `sequence_number` order — first matching rule wins.

### SQL

```sql
SELECT ar.sequence_number, ar.condition_field, ar.condition_op,
       ar.condition_value, ar.approver_type, ar.approver_ref, ar.escalation_days
FROM   approval_policies ap
JOIN   approval_rules ar ON ar.policy_id = ap.id
WHERE  ap.tenant_id = $1 AND ap.entity_type = $2 AND ap.is_active = TRUE
ORDER  BY ar.sequence_number;
```

### Go

```go
func BuildApprovalChain(rules []ApprovalRule, attrs map[string]any) []ApprovalStep {
    var chain []ApprovalStep
    for _, r := range rules {
        if EvaluateRule(r, attrs) {
            chain = append(chain, ApprovalStep{
                ApproverType: r.ApproverType,
                ApproverRef:  r.ApproverRef,
                EscalationDays: r.EscalationDays,
            })
        }
    }
    return chain
}
```

---

## 3.2 Reorder Point Alerts

### Use Case

Detecting when available stock falls to or below the reorder point after every stock movement and generating an alert or purchase requisition. Available stock = on-hand + confirmed pending receipts − confirmed pending issues (not just on-hand). One alert per crossing — do not re-alert on subsequent movements while stock remains below threshold.

### How Other Systems Use It

**SAP MRP** runs nightly batch comparing available stock against reorder point per material per plant and generates planned orders automatically. **Cin7** checks after every stock movement and creates "purchase order suggestions" visible to procurement. **Odoo** implements this as a "reordering rule" that auto-creates purchase order drafts when triggered.

### Data Required

```sql
reorder_rules: id, item_id, reorder_point, reorder_quantity,
               preferred_supplier_id, auto_create_pr, alert_recipient_id,
               is_active, tenant_id
reorder_alerts: id, rule_id, item_id, triggered_at, stock_at_trigger,
                acknowledged_at, tenant_id
```

### Performance & Optimisation

Implement as a PostgreSQL trigger on `inventory_movements`. The trigger checks for an existing open alert before inserting a new one — preventing duplicate alerts. PR creation is async via Temporal workflow triggered by `pg_notify`.

### SQL

```sql
-- Trigger: fires AFTER INSERT on inventory_movements
CREATE OR REPLACE FUNCTION check_reorder_point() RETURNS TRIGGER AS $$
DECLARE v_stock NUMERIC; v_rule reorder_rules%ROWTYPE;
BEGIN
    IF NEW.movement_type NOT IN ('ISSUE','SALE') THEN RETURN NEW; END IF;
    SELECT * INTO v_rule FROM reorder_rules
    WHERE item_id=NEW.item_id AND is_active=TRUE LIMIT 1;
    IF NOT FOUND THEN RETURN NEW; END IF;
    SELECT COALESCE(SUM(CASE WHEN movement_type IN('RECEIPT','ADJUSTMENT_IN')
                             THEN quantity ELSE -quantity END),0)
    INTO v_stock FROM inventory_movements WHERE item_id=NEW.item_id AND tenant_id=NEW.tenant_id;
    IF v_stock > v_rule.reorder_point THEN RETURN NEW; END IF;
    IF EXISTS(SELECT 1 FROM reorder_alerts WHERE item_id=NEW.item_id AND acknowledged_at IS NULL) THEN
        RETURN NEW;
    END IF;
    INSERT INTO reorder_alerts(id,rule_id,item_id,triggered_at,stock_at_trigger,tenant_id)
    VALUES(gen_random_uuid(),v_rule.id,NEW.item_id,NOW(),v_stock,NEW.tenant_id);
    PERFORM pg_notify('reorder_alert', NEW.item_id::text);
    RETURN NEW;
END; $$ LANGUAGE plpgsql;
```

---

## 3.3 Credit Limit Enforcement

### Use Case

Preventing sales, fuel draws, or orders from proceeding when a customer's total credit exposure (outstanding invoices + optionally open orders) exceeds their authorised limit. Three enforcement modes: HARD_BLOCK (reject outright), SOFT_WARN (proceed with warning logged), APPROVAL_REQUIRED (route to credit manager).

### How Other Systems Use It

**SAP SD Credit Management** supports static check (balance vs. limit), dynamic check (balance + open orders), and per-document max value. When a check fails the order is blocked and routed to the credit manager's worklist. **Oracle Order Management** distinguishes "exposure" (committed, uninvoiced) from "outstanding" (invoiced, unpaid). **Cin7** implements a hard-block only — simpler but inflexible.

### Data Required

```sql
customer_credit_profiles:
  id, customer_id, credit_limit, credit_terms_days,
  credit_check_type (HARD_BLOCK/SOFT_WARN/APPROVAL_REQUIRED),
  include_open_orders BOOLEAN, override_role TEXT,
  is_on_hold BOOLEAN, tenant_id
```

### Performance & Optimisation

Maintain a `customer_credit_exposure` table updated incrementally on every invoice post and payment. Credit check = one row lookup, not a SUM aggregation at runtime.

### SQL

```sql
SELECT ccp.credit_limit,
       COALESCE(inv.total,0) AS outstanding,
       COALESCE(CASE WHEN ccp.include_open_orders THEN ord.total ELSE 0 END,0) AS open_orders,
       ccp.credit_limit - COALESCE(inv.total,0)
           - COALESCE(CASE WHEN ccp.include_open_orders THEN ord.total ELSE 0 END,0) AS available
FROM   customer_credit_profiles ccp
LEFT JOIN LATERAL (SELECT SUM(amount_outstanding) total FROM invoices
    WHERE customer_id=ccp.customer_id AND status IN('OPEN','PARTIAL') AND tenant_id=$2) inv ON TRUE
LEFT JOIN LATERAL (SELECT SUM(total_amount) total FROM sales_orders
    WHERE customer_id=ccp.customer_id AND status IN('DRAFT','CONFIRMED') AND tenant_id=$2) ord ON TRUE
WHERE  ccp.customer_id=$1 AND ccp.tenant_id=$2;
```

### Go

```go
func CheckCredit(profile CreditProfile, exposure, orderAmt decimal.Decimal) CreditCheckResult {
    if profile.IsOnHold {
        return CreditCheckResult{Action: "HARD_BLOCKED", Message: "Account on manual hold"}
    }
    newExposure := exposure.Add(orderAmt)
    overage     := newExposure.Sub(profile.CreditLimit)
    if !overage.IsPositive() {
        return CreditCheckResult{Approved: true, Action: "APPROVED"}
    }
    switch profile.CheckType {
    case "HARD_BLOCK":
        return CreditCheckResult{Action: "HARD_BLOCKED",
            Message: fmt.Sprintf("Exceeds limit by %s", overage)}
    case "SOFT_WARN":
        return CreditCheckResult{Approved: true, Action: "SOFT_WARNING",
            Message: fmt.Sprintf("Limit exceeded by %s — logged", overage)}
    case "APPROVAL_REQUIRED":
        return CreditCheckResult{Action: "REQUIRES_APPROVAL",
            Message: fmt.Sprintf("Requires %s approval", profile.OverrideRole)}
    }
    return CreditCheckResult{Action: "HARD_BLOCKED"}
}
```

---

# Part 4 — Financial Reporting Layer

The financial reporting layer transforms the raw GL into the three statutory financial statements plus comparative and budget analysis. These are the outputs users actually sign, file with regulators, and present to investors.

---

## 4.1 Period & Fiscal Year Management

### Use Case

Every financial transaction belongs to a period. Periods define the boundaries for all reporting, budgeting, and statutory filing. The system must support custom fiscal years (Kenya businesses often use April–March or July–June), period locking to prevent backdated entries after close, and controlled re-opening with full audit trail.

### How Other Systems Use It

**SAP FI** uses "Posting Periods" — each period has an open/closed status per account type (A = assets, D = customers, K = vendors, S = GL). This allows, for example, the AR team to close their period while the GL accountant still has one day to post adjustments. **Oracle Financials** uses "Accounting Periods" with four states: Open, Closed, Permanently Closed, and Future-Enterable (entries accepted but not immediately reportable). **QuickBooks** implements a simpler "Closing Date" — transactions before the date require a password to modify.

### Data Required

```sql
fiscal_years:
  id, name, start_date, end_date, status (OPEN/CLOSED), tenant_id

accounting_periods:
  id, fiscal_year_id, period_number, name,
  start_date, end_date,
  status (FUTURE/OPEN/SOFT_CLOSED/HARD_CLOSED),
  closed_by, closed_at, tenant_id

period_open_log:
  id, period_id, action (OPENED/CLOSED/REOPENED),
  performed_by, performed_at, reason, tenant_id
```

### Period States

| State | Allows New Entries | Allows Edits | Requires Override |
|---|---|---|---|
| FUTURE | No | No | — |
| OPEN | Yes | Yes | No |
| SOFT_CLOSED | Warning only | Warning only | Password/role |
| HARD_CLOSED | No | No | Board approval |

### Performance & Optimisation

Period status is checked on every transaction posting. Cache the current period map in Redis (tenant_id → period lookup). Invalidate on any period status change. The lookup is: given a posting date, return the period_id and its status.

### SQL

```sql
-- Get period for a given posting date
SELECT id, status, period_number
FROM   accounting_periods
WHERE  tenant_id = $1
  AND  start_date <= $2 AND end_date >= $2
LIMIT  1;

-- Period close — lock and log
UPDATE accounting_periods
SET    status='HARD_CLOSED', closed_by=$2, closed_at=NOW()
WHERE  id=$1 AND tenant_id=$3
RETURNING id;

INSERT INTO period_open_log(id, period_id, action, performed_by, performed_at, reason, tenant_id)
VALUES(gen_random_uuid(), $1, 'CLOSED', $2, NOW(), $4, $3);
```

### Go

```go
func (s *PeriodService) ValidatePostingDate(ctx context.Context, tenantID string, date time.Time) error {
    period, err := s.repo.GetPeriodForDate(ctx, tenantID, date)
    if err != nil { return fmt.Errorf("no open period for %s", date.Format("2006-01-02")) }
    switch period.Status {
    case "HARD_CLOSED": return fmt.Errorf("period %s is hard-closed", period.Name)
    case "SOFT_CLOSED": return ErrSoftClosedPeriod{PeriodName: period.Name}
    case "FUTURE":      return fmt.Errorf("period %s is not yet open", period.Name)
    }
    return nil
}
```

---

## 4.2 Profit & Loss Statement

### Use Case

The income statement showing revenue, cost of sales, gross profit, operating expenses, operating profit, and net profit for a given period. The most frequently reviewed financial report in any business. Must support period filtering, cost centre drill-down, comparative columns (current vs. prior period vs. prior year), and budget variance.

### How Other Systems Use It

**SAP S/4HANA** generates the P&L from the Universal Journal (`ACDOCA`) by summing all postings to revenue and expense accounts within the selected period and organisational unit. The account hierarchy (profit centre hierarchy) determines the subtotal structure. **Xero** builds the P&L by traversing the account type hierarchy configured during setup. **QuickBooks** uses a fixed P&L template with hard-coded subtotal groupings (Ordinary Income, Other Income, Cost of Goods Sold, Expenses) that map to account types. **Odoo** generates P&L from a configurable "financial report" definition stored as rows with formulas referencing account codes or account groups.

### Data Required

```sql
accounts:
  id, code, name, type (REVENUE/EXPENSE/COGS),
  parent_id, sort_order, tenant_id

account_hierarchies:   -- for subtotal groupings
  id, name, level, parent_hierarchy_id,
  account_ids UUID[],  -- accounts rolled up into this node
  tenant_id

journal_entry_lines:
  id, journal_id, account_id, debit, credit,
  cost_centre_id, period_id, tenant_id
```

### Performance & Optimisation

P&L queries against months of transaction data are expensive without pre-aggregation. Materialise period-level account balances in `period_account_balances` nightly — the P&L query reads from this, not the raw lines table. For the current open period, union the materialised balances with a real-time aggregation of the current period's lines.

### SQL

```sql
-- P&L for a period using pre-aggregated balances
SELECT
    ah.name                                    AS section,
    ah.level,
    a.code,
    a.name                                     AS account_name,
    SUM(CASE WHEN a.type='REVENUE'
             THEN pab.credit_total - pab.debit_total
             ELSE pab.debit_total - pab.credit_total END) AS amount
FROM   period_account_balances pab
JOIN   accounts a              ON a.id = pab.account_id
JOIN   account_hierarchy_nodes ah ON a.id = ANY(ah.account_ids)
WHERE  pab.tenant_id = $1
  AND  pab.period_id = $2
  AND  a.type IN ('REVENUE','COGS','EXPENSE')
GROUP  BY ah.name, ah.level, ah.sort_order, a.code, a.name, a.type
ORDER  BY ah.sort_order, a.sort_order;
```

### Go

```go
type PLLine struct {
    Section string; Level int
    Code, Name string; Amount decimal.Decimal
}

func BuildPL(lines []PLLine) PLStatement {
    var revenue, cogs, expenses decimal.Decimal
    for _, l := range lines {
        switch l.Section {
        case "Revenue":  revenue = revenue.Add(l.Amount)
        case "COGS":     cogs = cogs.Add(l.Amount)
        case "Expenses": expenses = expenses.Add(l.Amount)
        }
    }
    grossProfit := revenue.Sub(cogs)
    return PLStatement{
        Revenue: revenue, COGS: cogs,
        GrossProfit: grossProfit, GrossMarginPct: safeDivide(grossProfit, revenue),
        OperatingExpenses: expenses,
        OperatingProfit: grossProfit.Sub(expenses),
    }
}
```

---

## 4.3 Balance Sheet

### Use Case

The statement of financial position as at a point in time: Assets = Liabilities + Equity. Balance sheet accounts accumulate from inception — unlike P&L accounts which reset each period. The closing balance of one period is the opening balance of the next. Equity includes retained earnings from all prior periods' net profits.

### How Other Systems Use It

**SAP** computes the balance sheet from the Universal Journal by summing all postings from inception for each balance sheet account, filtered by the organisational unit and company code. The "financial statement version" configuration maps account codes to balance sheet line items. **Xero** accumulates balance sheet balances from the first transaction date. Retained earnings are automatically computed as the cumulative P&L from all prior periods. **Oracle** uses "retained earnings account" to carry forward prior period net income — a specific GL account that receives the net profit when a year is closed.

### Data Required

Same as P&L plus:

```sql
opening_balances:
  id, account_id, period_id, amount, tenant_id
  -- Populated at fiscal year start for brought-forward balances

-- Balance sheet balance = opening_balance + net_movements_from_inception
-- Alternatively, store cumulative balances and reset only for new fiscal year
```

### Algorithm: Retained Earnings

Retained earnings = Sum of all posted P&L from inception, which equals:
```
Retained Earnings =
    Sum of all Revenue postings (credit balances)
  - Sum of all Expense and COGS postings (debit balances)
  + Opening retained earnings brought forward from prior years
```

This must equal the retained earnings account balance in the equity section.

### SQL

```sql
-- Cumulative balance for a balance sheet account as at a date
SELECT
    a.code, a.name, a.type,
    SUM(CASE WHEN a.type IN('ASSET','EXPENSE')
             THEN jel.debit - jel.credit
             ELSE jel.credit - jel.debit END) AS balance
FROM   journal_entry_lines jel
JOIN   journals j  ON j.id  = jel.journal_id
JOIN   accounts a  ON a.id  = jel.account_id
WHERE  jel.tenant_id = $1
  AND  j.posted_at  <= $2   -- as at date
  AND  j.status      = 'POSTED'
  AND  a.type IN ('ASSET','LIABILITY','EQUITY')
GROUP  BY a.code, a.name, a.type, a.sort_order
ORDER  BY a.sort_order;
```

### Go

```go
func ValidateBalanceSheet(assets, liabilities, equity decimal.Decimal) error {
    // Assets = Liabilities + Equity (accounting equation)
    rhs := liabilities.Add(equity)
    if assets.Sub(rhs).Abs().GreaterThan(decimal.NewFromFloat(0.01)) {
        return fmt.Errorf("balance sheet out of balance: assets=%s, L+E=%s", assets, rhs)
    }
    return nil
}
```

---

## 4.4 Cash Flow Statement (Indirect Method)

### Use Case

The statement showing how net profit translates to actual cash movement, covering operating, investing, and financing activities. Under the indirect method (required by IAS 7 for most entities), it starts from net profit and adjusts for non-cash items and working capital changes. Algorithmically the most complex of the three financial statements.

### How Other Systems Use It

**SAP** generates the cash flow statement from the Universal Journal by classifying every balance sheet movement into operating/investing/financing and applying the indirect adjustment logic. It requires careful configuration of account-to-cash-flow-line mappings. **Xero** does not generate a cash flow statement automatically — it provides a raw "Statement of Cash Flows" that requires manual review. **Oracle Cash Management** computes the indirect cash flow by reconciling opening to closing cash position using the movement in every balance sheet account.

### Indirect Method Algorithm

```
Net Cash from Operating Activities:
  Net Profit (from P&L)
  + Depreciation & Amortisation (non-cash expense add-back)
  + Impairment losses (non-cash)
  ± Changes in Working Capital:
      + Decrease in Trade Receivables (or − Increase)
      + Decrease in Inventories (or − Increase)
      + Decrease in Prepayments (or − Increase)
      − Decrease in Trade Payables (or + Increase)
      − Decrease in Accrued Liabilities (or + Increase)
  = Operating Cash Flow

Net Cash from Investing Activities:
  − Purchase of PPE / Intangibles
  + Proceeds from disposal of PPE
  − Loans extended to third parties
  + Loan repayments received

Net Cash from Financing Activities:
  + Proceeds from borrowings
  − Loan repayments made
  + Capital injected by owners
  − Dividends paid

Net Cash Movement = Operating + Investing + Financing
Opening Cash Balance + Net Movement = Closing Cash Balance
```

Closing Cash Balance must equal the cash and bank account balances on the balance sheet — if not, there is an error in classification.

### Data Required

```sql
cashflow_account_mappings:
  id, account_id, cashflow_line (OPERATING_WC/INVESTING/FINANCING/NON_CASH),
  cashflow_label TEXT, tenant_id
  -- Administrator-configured mapping of each account to its cash flow line
```

### SQL

```sql
-- Working capital movements (balance change between two dates)
SELECT
    cam.cashflow_line,
    cam.cashflow_label,
    SUM(CASE WHEN a.type='ASSET'
             THEN -(curr.balance - prior.balance)   -- increase in asset = use of cash
             ELSE   curr.balance - prior.balance     -- increase in liability = source
        END) AS cash_impact
FROM   cashflow_account_mappings cam
JOIN   accounts a ON a.id = cam.account_id
JOIN   LATERAL (
    SELECT SUM(CASE WHEN a2.type='ASSET' THEN jel.debit-jel.credit
                    ELSE jel.credit-jel.debit END) AS balance
    FROM journal_entry_lines jel JOIN journals j ON j.id=jel.journal_id
    JOIN accounts a2 ON a2.id=jel.account_id
    WHERE jel.account_id=cam.account_id AND jel.tenant_id=$1 AND j.posted_at<=$2
) curr ON TRUE
JOIN   LATERAL (
    SELECT SUM(CASE WHEN a2.type='ASSET' THEN jel.debit-jel.credit
                    ELSE jel.credit-jel.debit END) AS balance
    FROM journal_entry_lines jel JOIN journals j ON j.id=jel.journal_id
    JOIN accounts a2 ON a2.id=jel.account_id
    WHERE jel.account_id=cam.account_id AND jel.tenant_id=$1 AND j.posted_at<$3
) prior ON TRUE
WHERE  cam.tenant_id=$1 AND cam.cashflow_line='OPERATING_WC'
GROUP  BY cam.cashflow_line, cam.cashflow_label, cam.sort_order
ORDER  BY cam.sort_order;
```

---

## 4.5 Budget vs. Actual

### Use Case

Comparing actual financial performance against the plan (budget) to identify where the business is tracking above or below expectations. Budget variance is one of the primary management tools for controlling costs and revenue. Supports original budget, revised budget, and rolling forecast comparisons.

### How Other Systems Use It

**SAP CO (Controlling)** stores budgets against cost centres and profit centres with version control (original plan, latest estimate, actuals). Variance reports compare versions across any organisational dimension. **Adaptive Insights (Workday Planning)** models budgets as structured spreadsheets linked to the GL account hierarchy, with automatic synchronisation of actuals from the ERP. **QuickBooks** supports a simple budget entered per account per month, with a built-in Budget vs. Actuals report comparing YTD figures.

### Data Required

```sql
budgets:
  id, name, fiscal_year_id, version (ORIGINAL/REVISED/FORECAST),
  status, created_by, approved_by, tenant_id

budget_lines:
  id, budget_id, account_id, cost_centre_id, period_id,
  budgeted_amount, tenant_id
```

### Budget Variance Formulas

```
Variance Amount  = Actual - Budget
Variance %       = (Actual - Budget) / ABS(Budget) × 100

Favourable variance:
  Revenue:  Actual > Budget (more revenue than planned)
  Expense:  Actual < Budget (less spent than planned)

Adverse variance:
  Revenue:  Actual < Budget
  Expense:  Actual > Budget
```

### SQL

```sql
SELECT
    a.code, a.name, a.type,
    COALESCE(bl.budgeted_amount, 0)                              AS budget,
    COALESCE(pab.net_amount, 0)                                  AS actual,
    COALESCE(pab.net_amount,0) - COALESCE(bl.budgeted_amount,0)  AS variance,
    CASE WHEN bl.budgeted_amount != 0
         THEN ROUND((COALESCE(pab.net_amount,0) - bl.budgeted_amount)
                    / ABS(bl.budgeted_amount) * 100, 2)
         ELSE NULL END                                           AS variance_pct,
    CASE
        WHEN a.type='REVENUE' AND COALESCE(pab.net_amount,0) >= COALESCE(bl.budgeted_amount,0) THEN 'FAV'
        WHEN a.type='EXPENSE' AND COALESCE(pab.net_amount,0) <= COALESCE(bl.budgeted_amount,0) THEN 'FAV'
        ELSE 'ADV'
    END                                                          AS variance_status
FROM   accounts a
LEFT   JOIN budget_lines bl
       ON bl.account_id=a.id AND bl.period_id=$2 AND bl.tenant_id=$1
LEFT   JOIN period_account_balances pab
       ON pab.account_id=a.id AND pab.period_id=$2 AND pab.tenant_id=$1
WHERE  a.tenant_id=$1 AND a.type IN('REVENUE','COGS','EXPENSE')
ORDER  BY a.sort_order;
```

---

## 4.6 Comparative Reporting

### Use Case

Adding prior period and prior year columns to any report so management can see trends, not just point-in-time snapshots. Comparative columns are standard in IFRS-compliant financial statements (IAS 1 requires at least one comparative period).

### How Other Systems Use It

**SAP** adds comparative columns through "comparison ledgers" — logical views of historical period data. **Xero** supports comparison by adding prior period and prior year columns to any report via the report settings. **QuickBooks** calls this "Previous Period" and "Previous Year" and allows any combination of prior periods as additional columns.

### Period Resolution Logic

```
Given: current period = Month M of Year Y

Prior period   = Month (M-1) of Year Y
                 (or Month 12 of Year Y-1 if M=1)

Prior year     = Month M of Year (Y-1)

YTD            = Sum of all periods from fiscal year start through M
Prior year YTD = Sum of all periods in Y-1 from fiscal year start through M
```

### SQL

```sql
-- Multi-column comparative P&L (current, prior period, prior year)
SELECT
    a.code, a.name,
    COALESCE(curr.amount, 0)  AS current_period,
    COALESCE(prior.amount, 0) AS prior_period,
    COALESCE(pyr.amount, 0)   AS prior_year,
    COALESCE(curr.amount,0) - COALESCE(prior.amount,0) AS period_change,
    COALESCE(curr.amount,0) - COALESCE(pyr.amount,0)   AS yoy_change
FROM accounts a
LEFT JOIN period_account_balances curr  ON curr.account_id=a.id AND curr.period_id=$2
LEFT JOIN period_account_balances prior ON prior.account_id=a.id AND prior.period_id=$3
LEFT JOIN period_account_balances pyr   ON pyr.account_id=a.id   AND pyr.period_id=$4
WHERE a.tenant_id=$1 AND a.type IN('REVENUE','COGS','EXPENSE')
ORDER BY a.sort_order;
-- $2=current period_id, $3=prior period_id, $4=prior year same period_id
```

---

# Part 5 — Analytics Dimensions

Analytics dimensions answer the question "how is the business performing?" across multiple lenses: KPIs, departments, inventory health, customer quality, and product profitability.

---

## 5.1 KPI Framework

### Use Case

Defining, computing, and tracking key performance indicators across the business. A KPI is a named metric with a formula, a frequency, a unit, threshold boundaries (Red/Amber/Green), and a trend. The framework must support industry-specific KPI libraries loaded during onboarding and custom KPIs defined by the tenant.

### How Other Systems Use It

**SAP Analytics Cloud** implements KPIs as "Smart Insights" — computed metrics with predictive thresholds determined by historical variance. **Tableau** uses "calculated fields" as KPI formulas, with threshold-based colour coding configured per dashboard. **Power BI** uses "measures" (DAX formulas) as KPIs, with conditional formatting for RAG status. **Klipfolio** specialises in KPI dashboards and stores KPI definitions as data sources + aggregation formulas, decoupled from the underlying data system.

### Data Required

```sql
kpi_definitions:
  id, name, formula_type (RATIO/SUM/COUNT/CUSTOM),
  numerator_account_ids UUID[], denominator_account_ids UUID[],
  custom_sql TEXT,          -- for complex KPIs not expressible as ratios
  unit (CURRENCY/PERCENT/DAYS/RATIO/LITRES),
  frequency (DAILY/WEEKLY/MONTHLY),
  red_threshold, amber_threshold,  -- direction-aware
  threshold_direction (HIGHER_BETTER/LOWER_BETTER),
  industry_tag TEXT,        -- e.g. 'FUEL_RETAIL', 'HOSPITALITY'
  tenant_id UUID            -- NULL = system library

kpi_values:
  id, kpi_id, period_id, cost_centre_id,
  value NUMERIC(19,4), status (RED/AMBER/GREEN),
  computed_at, tenant_id
```

### Key KPIs by Industry

**Fuel Retail**

| KPI | Formula | Unit | Threshold Direction |
|---|---|---|---|
| Litres Sold / Day | Total litres dispensed / days in period | Litres | Higher better |
| Fuel Gross Margin / Litre | (Revenue − Landed Cost) / Litres | KES/L | Higher better |
| Dip Variance % | (Theoretical − Actual) / Theoretical | % | Lower better |
| Fleet DSO | Fleet AR / Daily fleet revenue × 365 | Days | Lower better |
| Cash-to-Fuel Ratio | Cash sales / Total sales | % | Configurable |
| Tank Fill % | Avg closing dip / Tank capacity | % | Configurable |

**Hospitality**

| KPI | Formula | Unit | Threshold Direction |
|---|---|---|---|
| RevPAR | Room revenue / Available rooms | KES | Higher better |
| F&B Cost % | F&B COGS / F&B Revenue | % | Lower better |
| Covers / Day | Count of meals served | Count | Higher better |
| Avg Check Value | F&B Revenue / Covers | KES | Higher better |

### SQL

```sql
-- Compute all due KPIs for a tenant for a period
SELECT kd.id, kd.name, kd.formula_type,
       kd.numerator_account_ids, kd.denominator_account_ids,
       kd.red_threshold, kd.amber_threshold, kd.threshold_direction
FROM   kpi_definitions kd
WHERE  (kd.tenant_id = $1 OR kd.tenant_id IS NULL)
  AND  kd.frequency  = $2
  AND  NOT EXISTS (
           SELECT 1 FROM kpi_values kv
           WHERE kv.kpi_id=kd.id AND kv.period_id=$3 AND kv.tenant_id=$1
       );
```

### Go

```go
func ComputeKPIStatus(value, red, amber decimal.Decimal, dir string) string {
    if dir == "HIGHER_BETTER" {
        if value.GreaterThanOrEqual(amber) { return "GREEN" }
        if value.GreaterThanOrEqual(red)   { return "AMBER" }
        return "RED"
    }
    // LOWER_BETTER
    if value.LessThanOrEqual(amber) { return "GREEN" }
    if value.LessThanOrEqual(red)   { return "AMBER" }
    return "RED"
}
```

---

## 5.2 Cost Centre & Department Analytics

### Use Case

Producing P&L, budget vs. actual, and KPI views filtered to a specific cost centre or department. Enables managers to be accountable for their department's performance without seeing the full business P&L. Requires every transaction line to carry a cost_centre_id.

### How Other Systems Use It

**SAP CO** is entirely built around cost centres — every cost posting carries a cost centre assignment. The CO module maintains parallel "management accounts" (internal reporting) alongside the FI module's statutory accounts. **Xero** uses "Tracking Categories" as a lightweight cost centre equivalent — up to two categories per transaction line. **QuickBooks** uses "Classes" for cost centre equivalent; one class per transaction line, not per line item. **Oracle** uses "Segments" in the Account Code Combination to capture organisation, department, and project simultaneously.

### Data Required

```sql
cost_centres: id, name, code, parent_id, manager_id, tenant_id
-- journal_entry_lines.cost_centre_id is already required
```

### SQL

```sql
-- Department P&L
SELECT
    cc.name AS department,
    a.type, a.code, a.name AS account,
    SUM(CASE WHEN a.type='REVENUE' THEN jel.credit-jel.debit
             ELSE jel.debit-jel.credit END) AS amount
FROM   journal_entry_lines jel
JOIN   journals j   ON j.id  = jel.journal_id
JOIN   accounts a   ON a.id  = jel.account_id
JOIN   cost_centres cc ON cc.id = jel.cost_centre_id
WHERE  jel.tenant_id = $1
  AND  j.period_id   = $2
  AND  cc.id         = $3
  AND  j.status      = 'POSTED'
  AND  a.type IN ('REVENUE','COGS','EXPENSE')
GROUP  BY cc.name, a.type, a.sort_order, a.code, a.name
ORDER  BY a.sort_order;
```

---

## 5.3 Overhead Allocation

### Use Case

Apportioning shared costs (rent, utilities, admin salaries, insurance) across departments using a defined allocation key. Without allocation, department P&Ls are incomplete and not comparable — the Station may look highly profitable simply because it carries no overhead.

### Allocation Key Types

| Key Type | Use When | Example |
|---|---|---|
| Revenue % | Overhead driven by business activity | Rent apportioned by department revenue share |
| Headcount | Overhead driven by people | HR admin cost by employee count per dept |
| Floor Area | Overhead driven by space | Electricity by square metres per dept |
| Equal Share | Overhead not correlated to any driver | Insurance equally across 3 departments |
| Custom % | Manually set by management | Fixed percentages set at budget time |

### Data Required

```sql
overhead_allocation_rules:
  id, source_account_id,  -- the shared cost account
  allocation_key (REVENUE_PCT/HEADCOUNT/FLOOR_AREA/EQUAL/CUSTOM),
  period_id,
  allocations JSONB,  -- [{cost_centre_id, pct}] for CUSTOM; computed for others
  tenant_id
```

### SQL

```sql
-- Compute revenue-based allocation keys for a period
SELECT
    cc.id AS cost_centre_id,
    cc.name,
    ROUND(dept_rev.amount / total_rev.amount * 100, 4) AS allocation_pct
FROM cost_centres cc
CROSS JOIN LATERAL (
    SELECT SUM(jel.credit - jel.debit) AS amount
    FROM journal_entry_lines jel JOIN journals j ON j.id=jel.journal_id
    JOIN accounts a ON a.id=jel.account_id
    WHERE jel.cost_centre_id=cc.id AND jel.tenant_id=$1
      AND j.period_id=$2 AND j.status='POSTED' AND a.type='REVENUE'
) dept_rev
CROSS JOIN LATERAL (
    SELECT SUM(jel.credit - jel.debit) AS amount
    FROM journal_entry_lines jel JOIN journals j ON j.id=jel.journal_id
    JOIN accounts a ON a.id=jel.account_id
    WHERE jel.tenant_id=$1 AND j.period_id=$2
      AND j.status='POSTED' AND a.type='REVENUE'
) total_rev
WHERE cc.tenant_id=$1 AND dept_rev.amount > 0;
```

---

## 5.4 Inventory Analytics

### Use Case

Going beyond valuation (Phase 1) to understand inventory health: which items are moving fast or slow, which carry the most value, which are approaching obsolescence, and how many days of stock remain at current consumption rates.

### ABC Analysis

Ranks items by revenue or usage contribution. The Pareto principle consistently holds: ~20% of items drive ~80% of value.

```
For each item: contribution = item_revenue / total_revenue
Sort descending by contribution
Cumulative A cutoff: top items summing to 80% → Class A (tight control)
Cumulative B cutoff: next items summing to 15% → Class B (moderate control)
Remainder                                      → Class C (minimal control)
```

### Slow-Moving & Dead Stock

```
Slow-Moving: no issue movement for > N days (configurable, e.g. 60 days)
Dead Stock:  no issue movement for > M days (configurable, e.g. 180 days)
Stock Coverage Days = Current Stock Qty / Avg Daily Consumption (30-day)
```

### SQL

```sql
-- ABC Analysis
WITH item_revenue AS (
    SELECT tl.item_id,
           SUM(tl.net_amount) AS revenue,
           SUM(tl.net_amount) / SUM(SUM(tl.net_amount)) OVER () AS revenue_share
    FROM   transaction_lines tl JOIN transactions t ON t.id=tl.transaction_id
    WHERE  t.tenant_id=$1 AND t.type='SALE'
      AND  t.transaction_date >= NOW()-INTERVAL '12 months'
    GROUP  BY tl.item_id
),
ranked AS (
    SELECT *, SUM(revenue_share) OVER (ORDER BY revenue DESC) AS cumulative_share
    FROM item_revenue
)
SELECT item_id, revenue, revenue_share,
    CASE WHEN cumulative_share <= 0.80 THEN 'A'
         WHEN cumulative_share <= 0.95 THEN 'B'
         ELSE 'C' END AS abc_class
FROM ranked ORDER BY revenue DESC;
```

```sql
-- Slow-moving stock (no movement in last 60 days)
SELECT i.id, i.sku, i.name,
       i.current_quantity, i.current_avg_cost,
       i.current_quantity * i.current_avg_cost AS stock_value,
       MAX(m.movement_date) AS last_movement_date,
       CURRENT_DATE - MAX(m.movement_date) AS days_since_movement
FROM   inventory_items i
LEFT JOIN inventory_movements m ON m.item_id=i.id AND m.movement_type IN('ISSUE','SALE')
WHERE  i.tenant_id=$1 AND i.current_quantity > 0
GROUP  BY i.id, i.sku, i.name, i.current_quantity, i.current_avg_cost
HAVING MAX(m.movement_date) < CURRENT_DATE - $2  -- $2 = days threshold
    OR MAX(m.movement_date) IS NULL
ORDER  BY days_since_movement DESC NULLS FIRST;
```

---

## 5.5 Receivables & Payables Analytics

### Use Case

Tracking the health and trend of the receivables and payables ledger over time, going beyond the point-in-time aging snapshot to reveal whether the business is improving or deteriorating in its credit management.

### Key Metrics & Formulas

**Days Sales Outstanding (DSO) Trend**
```
DSO = (AR Balance / Credit Sales in Period) × Days in Period
Compute monthly. Plot 12-month trend. Rising DSO = deteriorating collections.
```

**Collections Effectiveness Index (CEI)**
```
CEI = (Opening AR + Credit Sales − Closing AR) / (Opening AR + Credit Sales − Current Invoices) × 100
CEI of 100% = perfect collection. Below 80% = collections issue.
```

**Customer Payment Behaviour Profile**
```
For each customer: avg_days_to_pay = AVG(payment_date - invoice_date) per invoice
Customers with avg_days_to_pay > credit_terms → flag as chronic late payers
```

### SQL

```sql
-- Monthly DSO trend (last 12 months)
SELECT
    p.name AS period,
    ROUND(AVG(i.amount_outstanding) /
          NULLIF(SUM(CASE WHEN t.type='SALE' THEN t.total_amount ELSE 0 END),0)
          * p.days_in_period, 1) AS dso_days
FROM   accounting_periods p
JOIN   invoices i      ON i.due_date BETWEEN p.start_date AND p.end_date
JOIN   transactions t  ON t.period_id=p.id AND t.tenant_id=$1
WHERE  p.tenant_id=$1
  AND  p.start_date >= NOW()-INTERVAL '12 months'
GROUP  BY p.id, p.name, p.days_in_period
ORDER  BY p.start_date;
```

```sql
-- Customer payment behaviour (avg days to pay)
SELECT
    c.name,
    COUNT(p.id) AS invoices_paid,
    ROUND(AVG(p.payment_date - i.invoice_date), 1) AS avg_days_to_pay,
    i.credit_terms_days,
    ROUND(AVG(p.payment_date - i.invoice_date),1)
        - i.credit_terms_days AS avg_days_late
FROM   invoice_payments p
JOIN   invoices i   ON i.id = p.invoice_id
JOIN   customers c  ON c.id = i.customer_id
WHERE  i.tenant_id  = $1
  AND  p.payment_date >= NOW()-INTERVAL '12 months'
GROUP  BY c.id, c.name, i.credit_terms_days
ORDER  BY avg_days_late DESC;
```

---

## 5.6 Profitability Analytics

### Use Case

Decomposing gross margin to understand which products, customers, and channels drive profitability — and which destroy it. Essential for pricing decisions, customer portfolio management, and product mix optimisation.

### Key Analyses

**Gross Margin by Product**
```
Margin per Unit   = Selling Price − Unit Cost (WAC or FIFO cost at time of sale)
Margin % per Unit = Margin per Unit / Selling Price × 100
```

**Gross Margin by Customer**
```
Customer Revenue         = Σ invoiced amounts to customer
Customer COGS            = Σ cost of goods delivered to customer
Customer Gross Margin    = Revenue − COGS
Customer Gross Margin %  = Gross Margin / Revenue × 100
Sort descending — the Pareto principle applies: top 20% of customers
often generate 80%+ of gross profit.
```

**Margin Trend**
```
Monthly gross margin % plotted over 12 months.
Sustained compression signals: pricing pressure, rising input costs,
product mix shift toward lower-margin items.
```

### SQL

```sql
-- Gross margin by product over a period
SELECT
    p.sku, p.name,
    SUM(tl.quantity)                              AS units_sold,
    SUM(tl.net_amount)                            AS revenue,
    SUM(tl.quantity * im.unit_cost_at_sale)        AS cogs,
    SUM(tl.net_amount) - SUM(tl.quantity * im.unit_cost_at_sale) AS gross_margin,
    ROUND((SUM(tl.net_amount) - SUM(tl.quantity * im.unit_cost_at_sale))
          / NULLIF(SUM(tl.net_amount),0) * 100, 2) AS margin_pct
FROM   transaction_lines tl
JOIN   transactions t   ON t.id  = tl.transaction_id
JOIN   products p       ON p.id  = tl.product_id
JOIN   inventory_movements im ON im.reference_id = tl.id AND im.movement_type='SALE'
WHERE  t.tenant_id = $1 AND t.period_id = $2 AND t.type='SALE' AND t.status='POSTED'
GROUP  BY p.sku, p.name
ORDER  BY gross_margin DESC;
```

---

# Part 6 — Reporting Platform Infrastructure

The reporting platform is the engine that makes all the above performant, queryable, exportable, and distributable at scale.

---

## 6.1 Report Definition Schema

### Use Case

Storing report definitions as data — enabling administrators to create, modify, and share reports without code deployments. A report definition describes what data to fetch, how to aggregate it, how to display it, and who can see it.

### How Other Systems Use It

**SAP Report Painter/Writer** stores report definitions as configuration tables (rows, columns, characteristic values). Reports can be copied, versioned, and assigned to user groups. **Odoo Financial Reports** stores report definitions as records with rows that reference account groups or formulas. Rows can have child rows for hierarchical subtotaling. **Microsoft SSRS** stores report definitions as `.rdl` XML files — portable, versionable, but requiring developer skills to create.

### Data Required

```sql
report_definitions:
  id, name, description, report_type (PL/BS/CF/KPI/CUSTOM),
  base_query TEXT,         -- parameterised SQL or query reference
  column_definitions JSONB, -- [{key, label, format, formula}]
  filter_schema JSONB,      -- what filters this report accepts
  default_filters JSONB,    -- default parameter values
  is_system BOOLEAN,        -- system reports cannot be deleted
  owner_id UUID, tenant_id UUID

report_permissions:
  id, report_id, role TEXT, can_view BOOLEAN,
  can_export BOOLEAN, tenant_id
```

### Go

```go
type ReportDefinition struct {
    ID              uuid.UUID
    Name            string
    ReportType      string
    BaseQuery       string
    ColumnDefs      []ColumnDef
    FilterSchema    []FilterField
    DefaultFilters  map[string]any
}

type ColumnDef struct {
    Key     string; Label   string
    Format  string // CURRENCY / PERCENT / NUMBER / DATE / TEXT
    Formula string // optional: "col_a - col_b"
}
```

---

## 6.2 Pre-Aggregation Strategy

### Use Case

Financial reports running against months or years of raw transaction data are expensive. Pre-aggregating balances at the period level makes report generation near-instantaneous at the cost of a nightly maintenance window.

### Aggregation Hierarchy

```
Level 1 (raw):            journal_entry_lines        — every posting
Level 2 (period+account): period_account_balances    — nightly roll-up
Level 3 (period+dept):    period_dept_balances       — nightly roll-up with cost centre
Level 4 (KPI):            kpi_values                 — computed from L2/L3

Reports read from:
  Current open period  → L2 (materialised) UNION real-time L1 for current period
  Closed periods       → L2 only (fully materialised)
  KPI dashboard        → L4 only
  Drill-down to detail → L1 directly with tight filters
```

### SQL

```sql
-- Nightly period_account_balances roll-up (run after period close or at 2am)
INSERT INTO period_account_balances
    (id, tenant_id, period_id, account_id, cost_centre_id,
     debit_total, credit_total, net_amount, row_count)
SELECT
    gen_random_uuid(),
    jel.tenant_id,
    j.period_id,
    jel.account_id,
    jel.cost_centre_id,
    SUM(jel.debit),
    SUM(jel.credit),
    SUM(CASE WHEN a.type IN('ASSET','EXPENSE') THEN jel.debit-jel.credit
             ELSE jel.credit-jel.debit END),
    COUNT(*)
FROM   journal_entry_lines jel
JOIN   journals j  ON j.id  = jel.journal_id
JOIN   accounts a  ON a.id  = jel.account_id
WHERE  j.status = 'POSTED'
  AND  j.period_id = $1         -- the period being rolled up
GROUP  BY jel.tenant_id, j.period_id, jel.account_id, jel.cost_centre_id
ON CONFLICT (tenant_id, period_id, account_id, cost_centre_id)
DO UPDATE SET
    debit_total  = EXCLUDED.debit_total,
    credit_total = EXCLUDED.credit_total,
    net_amount   = EXCLUDED.net_amount,
    row_count    = EXCLUDED.row_count,
    updated_at   = NOW();
```

---

## 6.3 Drill-Down Architecture

### Use Case

Every summary figure in a report must be traceable to the underlying transactions. A manager who sees an expense line of KES 340,000 must be able to click through to the individual journal entries, invoices, and source documents that make up that figure. This is both a usability requirement and an audit requirement.

### How Other Systems Use It

**SAP** implements drill-down through "line item reports" linked to every account balance. Each balance is tagged with a document number; clicking the balance opens the document display. **Xero** implements drill-down via hyperlinked totals on every report — clicking a total opens the transaction list filtered to the accounts and period that comprise it. **QuickBooks** calls this "QuickZoom" — every report figure is clickable and opens the underlying transaction register.

### Design Pattern

```
Each report cell carries metadata alongside its value:
  - tenant_id
  - period_id (or date range)
  - account_ids[]
  - cost_centre_id (if applicable)

Drill-down request = these metadata parameters
Drill-down query   = SELECT * FROM journal_entry_lines WHERE
                     account_id = ANY($1) AND period_id = $2 ...

This pattern means drill-down is generic — the same endpoint
serves drill-down for any report cell.
```

### Go

```go
type DrillDownContext struct {
    TenantID      uuid.UUID
    PeriodID      uuid.UUID
    AccountIDs    []uuid.UUID
    CostCentreID  *uuid.UUID
    DateFrom, DateTo time.Time
}

// DrillDown returns the raw journal lines behind any aggregated figure
func (r *ReportService) DrillDown(ctx context.Context, d DrillDownContext) ([]JournalLineDetail, error) {
    return r.repo.GetJournalLines(ctx, GetJournalLinesParams{
        TenantID:     d.TenantID,
        AccountIDs:   d.AccountIDs,
        PeriodID:     d.PeriodID,
        CostCentreID: d.CostCentreID,
    })
}
```

---

## 6.4 Multi-Dimensional Filtering

### Use Case

Every report must be filterable by any combination of: period, cost centre, account, vendor/customer, product, tag, and custom attributes. The filter state must be expressible as a URL-safe parameter set so reports can be shared, bookmarked, and scheduled with specific filters applied.

### Filter Schema Design

Filters are defined in the report definition's `filter_schema` JSONB field. The query engine reads the active filters and builds the WHERE clause dynamically. The critical constraint: **filter application must always enforce RLS** — tenant_id is never a user-supplied filter, it is always injected from the session.

```go
type ReportFilter struct {
    PeriodID      *uuid.UUID   `json:"period_id,omitempty"`
    CostCentreIDs []uuid.UUID  `json:"cost_centre_ids,omitempty"`
    AccountTypes  []string     `json:"account_types,omitempty"`
    DateFrom      *time.Time   `json:"date_from,omitempty"`
    DateTo        *time.Time   `json:"date_to,omitempty"`
    Tags          []string     `json:"tags,omitempty"`
}

func BuildFilterClause(f ReportFilter) (string, []any) {
    var clauses []string
    var args []any
    n := 2 // $1 is always tenant_id
    if f.PeriodID != nil {
        args = append(args, *f.PeriodID)
        clauses = append(clauses, fmt.Sprintf("j.period_id = $%d", n)); n++
    }
    if len(f.CostCentreIDs) > 0 {
        args = append(args, f.CostCentreIDs)
        clauses = append(clauses, fmt.Sprintf("jel.cost_centre_id = ANY($%d)", n)); n++
    }
    return strings.Join(clauses, " AND "), args
}
```

---

## 6.5 Scheduled Reports & Distribution

### Use Case

Automatically generating reports on a schedule (daily, weekly, monthly) and distributing them to configured recipients via email or in-app notification. The daily station report sent to the director every morning at 7am is a scheduled report.

### How Other Systems Use It

**SAP Report Scheduling** uses the Background Processing (batch job) framework. Reports are scheduled as background jobs with a job definition, a schedule (immediate, periodic, event-based), and an output destination (spool, email, archive). **QuickBooks** supports scheduled reports via email — define the report, the schedule (daily/weekly/monthly), and the recipient list. **Xero** does not support scheduled reports natively; third-party tools like Fathom or Spotlight Reporting fill this gap.

### Data Required

```sql
scheduled_reports:
  id, report_id, schedule_cron TEXT,   -- "0 7 * * *" = daily at 7am
  filter_params JSONB,                 -- filters to apply at run time
  output_format (PDF/EXCEL/CSV),
  recipients JSONB,                    -- [{type: EMAIL/USER, address/id}]
  last_run_at, next_run_at,
  is_active, tenant_id
```

### Go (Temporal Workflow)

```go
// ScheduledReportWorkflow runs on a cron schedule per Temporal
func ScheduledReportWorkflow(ctx workflow.Context, scheduleID string) error {
    // Run forever on the configured cron
    for {
        _ = workflow.Sleep(ctx, durationUntilNextRun(scheduleID))

        var result ReportResult
        err := workflow.ExecuteActivity(ctx,
            GenerateReportActivity, scheduleID).Get(ctx, &result)
        if err != nil { continue } // log and continue

        workflow.ExecuteActivity(ctx, DistributeReportActivity, result)
    }
}
```

---

## 6.6 Export Engine

### Use Case

Generating downloadable versions of any report in PDF (for sharing/filing), Excel (for further analysis), and CSV (for data exchange). The export must faithfully represent the report including subtotals, grouping, formatting, and column headers.

### How Other Systems Use It

**SAP** exports via the ALV (ABAP List Viewer) grid — built-in support for spreadsheet, PDF, and CSV export from any list report. **Xero** exports every report as PDF or Google Sheets/Excel. The Google Sheets export maintains the live connection — the spreadsheet updates when Xero data changes. **QuickBooks** exports to Excel and PDF. The Excel export uses the XLSX format with formatted cells matching the on-screen report.

### Implementation Pattern

Use separate rendering pipelines per format. Never try to produce HTML that converts cleanly to both PDF and Excel — they have fundamentally different layout models.

```go
type ExportRequest struct {
    ReportID  uuid.UUID
    Filters   ReportFilter
    Format    string // PDF / EXCEL / CSV
    TenantID  uuid.UUID
}

func (e *ExportEngine) Export(ctx context.Context, req ExportRequest) ([]byte, string, error) {
    // 1. Execute report query
    data, err := e.reportSvc.Execute(ctx, req.ReportID, req.Filters, req.TenantID)
    if err != nil { return nil, "", err }

    // 2. Render in requested format
    switch req.Format {
    case "PDF":
        buf, err := e.pdfRenderer.Render(data)
        return buf, "application/pdf", err
    case "EXCEL":
        buf, err := e.xlsxRenderer.Render(data)
        return buf, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", err
    case "CSV":
        buf, err := e.csvRenderer.Render(data)
        return buf, "text/csv", err
    }
    return nil, "", fmt.Errorf("unsupported format: %s", req.Format)
}
```

**PDF:** Use `chromedp` (headless Chrome) to render an HTML report template to PDF — this produces the highest-fidelity output matching the on-screen view.

**Excel:** Use `excelize` library. Generate one sheet per report section. Apply cell number formats (`#,##0.00` for currency, `0.00%` for percentage). Freeze the header row. Auto-fit column widths.

**CSV:** Flatten all rows to a single header + data format. Include all visible columns. Use UTF-8 with BOM for Excel compatibility on Windows.

---

# Part 7 — Audit & Activity Analytics

---

## 7.1 GL Change Log

### Use Case

Recording every modification to a posted journal entry — who changed what field, from what value, to what value, and when. Posted entries should never be silently modified; every change must be visible, attributed, and reversible only through an explicit reversal entry (not a direct edit).

### How Other Systems Use It

**SAP** maintains the "Change Documents" table (`CDHDR`/`CDPOS`) for every changed object. Posted FI documents cannot be directly edited — they can only be reversed via a reversing document. **Xero** implements an immutable ledger — posted transactions cannot be edited; they must be voided and re-entered. The void creates an offsetting entry, preserving the audit trail. **QuickBooks** logs all changes in the "Audit Log" (`Reports → Audit Log`) — every transaction change shows old value, new value, user, and timestamp.

### Data Required

```sql
gl_audit_log:
  id UUID, entity_type TEXT, entity_id UUID,
  action (CREATE/UPDATE/SOFT_DELETE/REVERSAL),
  changed_by UUID, changed_at TIMESTAMPTZ,
  old_values JSONB, new_values JSONB,
  ip_address INET, session_id TEXT,
  tenant_id UUID
```

### Implementation via PostgreSQL Trigger

```sql
CREATE OR REPLACE FUNCTION audit_journal_changes() RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'UPDATE' THEN
        INSERT INTO gl_audit_log(id, entity_type, entity_id, action,
            changed_by, changed_at, old_values, new_values, tenant_id)
        VALUES(gen_random_uuid(), 'journal_entry_line', NEW.id, 'UPDATE',
            current_setting('app.current_user')::uuid, NOW(),
            to_jsonb(OLD), to_jsonb(NEW), NEW.tenant_id);
    END IF;
    RETURN NEW;
END; $$ LANGUAGE plpgsql;

CREATE TRIGGER trg_journal_audit
AFTER UPDATE ON journal_entry_lines
FOR EACH ROW EXECUTE FUNCTION audit_journal_changes();
```

### SQL (Audit Query)

```sql
-- All changes to a specific journal
SELECT
    al.action, al.changed_at,
    u.name AS changed_by,
    al.old_values->>'debit'  AS old_debit,
    al.new_values->>'debit'  AS new_debit,
    al.old_values->>'credit' AS old_credit,
    al.new_values->>'credit' AS new_credit
FROM   gl_audit_log al
JOIN   users u ON u.id = al.changed_by::uuid
WHERE  al.entity_id  = $1
  AND  al.tenant_id  = $2
ORDER  BY al.changed_at;
```

---

## 7.2 Reconciliation History

### Use Case

Recording which periods were reconciled, by whom, on what date, and what the reconciliation outcome was (balance matched, items outstanding, manually cleared). This history is critical for audit — an auditor will ask "who reconciled this account and when?" for every material account in the balance sheet.

### Data Required

```sql
reconciliation_records:
  id, account_id, period_id,
  reconciled_by UUID, reconciled_at TIMESTAMPTZ,
  gl_closing_balance, statement_closing_balance,
  outstanding_items_count, outstanding_items_value,
  status (BALANCED/OUTSTANDING_ITEMS/MANUAL_OVERRIDE),
  notes TEXT, tenant_id
```

### SQL

```sql
-- Reconciliation status per account per period (audit view)
SELECT
    a.code, a.name,
    p.name          AS period,
    u.name          AS reconciled_by,
    rr.reconciled_at,
    rr.gl_closing_balance,
    rr.statement_closing_balance,
    rr.gl_closing_balance - rr.statement_closing_balance AS difference,
    rr.outstanding_items_count,
    rr.status
FROM   reconciliation_records rr
JOIN   accounts a           ON a.id = rr.account_id
JOIN   accounting_periods p ON p.id = rr.period_id
JOIN   users u              ON u.id = rr.reconciled_by
WHERE  rr.tenant_id = $1
  AND  p.fiscal_year_id = $2
ORDER  BY a.code, p.start_date;
```

---

## 7.3 User Activity Analytics

### Use Case

Tracking how users interact with the ERP: which features they use, which reports they run, login patterns, and data access events. This serves both operational insight (adoption of new features) and security (detecting unusual access patterns).

### Data Required

```sql
user_activity_events:
  id, user_id, tenant_id,
  event_type (LOGIN/LOGOUT/REPORT_VIEW/REPORT_EXPORT/
              TRANSACTION_CREATE/TRANSACTION_APPROVE/
              ADMIN_ACTION),
  entity_type TEXT, entity_id UUID,
  ip_address INET, user_agent TEXT,
  occurred_at TIMESTAMPTZ,
  metadata JSONB   -- event-specific additional data
```

### Key Analytics Queries

```sql
-- Feature adoption: most-used reports this month
SELECT
    rd.name AS report_name,
    COUNT(*)                   AS views,
    COUNT(DISTINCT uae.user_id) AS unique_users
FROM   user_activity_events uae
JOIN   report_definitions rd ON rd.id::text = uae.metadata->>'report_id'
WHERE  uae.tenant_id = $1
  AND  uae.event_type = 'REPORT_VIEW'
  AND  uae.occurred_at >= date_trunc('month', NOW())
GROUP  BY rd.name
ORDER  BY views DESC;
```

```sql
-- Login anomalies: users logging in outside normal hours
SELECT
    u.name, uae.ip_address,
    uae.occurred_at,
    EXTRACT(HOUR FROM uae.occurred_at) AS hour_of_day
FROM   user_activity_events uae
JOIN   users u ON u.id = uae.user_id
WHERE  uae.tenant_id = $1
  AND  uae.event_type = 'LOGIN'
  AND  EXTRACT(HOUR FROM uae.occurred_at) NOT BETWEEN 6 AND 21
  AND  uae.occurred_at >= NOW() - INTERVAL '30 days'
ORDER  BY uae.occurred_at DESC;
```

---

# Part 8 — Summary & Build Priority

## Complete Algorithm Inventory

### Phase 1 — Deterministic Formulas (Build First)

| Algorithm | Latency Target | Key Dependency | Risk if Skipped |
|---|---|---|---|
| Double-entry validation | < 1ms | None | Corrupt GL |
| Payroll calculations | < 500ms/employee | Rate table | Statutory penalties |
| VAT calculation | < 1ms | VAT rate table | Incorrect tax filings |
| Inventory valuation (WAC) | < 5ms | Inventory movements | Wrong COGS, wrong margin |
| Inventory valuation (FIFO) | < 20ms | Layer table | Wrong COGS |
| Dip variance | < 10ms | Dip + meter readings | Undetected fuel loss |
| AR aging | < 1s (batch) | Invoices + due dates | Understated bad debt |
| Provision calculation | < 1s (batch) | Aging + rates | Overstated AR |

### Phase 2 — Statistical Thresholds (Build with Phase 1)

| Algorithm | Data Required | False Positive Risk | Key Benefit |
|---|---|---|---|
| Z-score anomaly | 30+ samples/dimension | Medium if data is skewed | Fast, explainable |
| IQR anomaly | 30+ samples | Lower than Z-score | Robust to heavy tails |
| Benford's Law | 500+ records | Low (batch, not real-time) | Detects fabrication |
| Reconciliation matching | Unreconciled GL + bank | Low (threshold-based) | Reduces manual matching |

### Phase 3 — Rule Engines (Build Incrementally)

| Rule Engine | Config Complexity | Business Impact | Start With |
|---|---|---|---|
| Approval routing | Medium — needs policy UI | High — internal controls | Amount-based simple routing |
| Reorder alerts | Low — per-item config | High — prevents stockouts | Trigger + notification |
| Credit enforcement | Low — per-customer config | High — prevents bad debt | Hard block first, refine later |

### Phase 4 — Financial Reporting Layer

| Component | Dependency | Priority |
|---|---|---|
| Period management | Phase 1 | Critical — build before any reports |
| P&L Statement | Period + pre-aggregation | Build first of the three statements |
| Balance Sheet | Period + P&L | After P&L is solid |
| Cash Flow Statement | Balance Sheet + P&L | Last — most complex |
| Budget vs. Actual | P&L + budget schema | After P&L |
| Comparative reporting | Pre-aggregation | Add to P&L as column option |

### Phase 5 — Analytics Dimensions

| Component | Dependency | Priority |
|---|---|---|
| KPI Framework | P&L + Phase 1 | High — immediately visible value |
| Cost Centre Analytics | P&L + cost_centre_id | High for multi-dept businesses |
| Overhead Allocation | Cost Centre + allocation rules | Medium |
| Inventory Analytics | Phase 1 valuation | High for physical stock businesses |
| AR/AP Analytics | Phase 1 aging | Medium |
| Profitability Analytics | P&L + COGS at line level | High for margin-sensitive businesses |

### Phase 6 — Reporting Platform

| Component | Dependency | Priority |
|---|---|---|
| Pre-aggregation | All Phase 1 | Critical for performance |
| Report definition schema | Pre-aggregation | Build before custom reports |
| Drill-down | Report definitions | High — required for usability |
| Multi-dimensional filtering | Report definitions | Medium |
| Scheduled reports | Report engine | Medium — high ROI for operators |
| Export engine | Report engine | High — CSV/Excel expected by users |

### Phase 7 — Audit & Activity

| Component | Dependency | Priority |
|---|---|---|
| GL audit trigger | Phase 1 | Critical — build from day one |
| Reconciliation history | Phase 2 reconciliation | Medium |
| User activity analytics | Auth system | Low initially, high for compliance |

---

## Technology Decisions Summary

| Concern | Decision | Rationale |
|---|---|---|
| Numeric arithmetic | `shopspring/decimal` | No float rounding errors |
| Pre-aggregation store | PostgreSQL materialised views + `period_account_balances` table | Same database, no sync issues |
| Audit logging | PostgreSQL triggers → append-only table | Atomic with the operation being audited |
| Anomaly baselines | Nightly batch → `transaction_baselines` table | Runtime check = one row lookup |
| Report scheduling | Temporal cron workflow | Durable, observable, retryable |
| Export rendering | chromedp (PDF), excelize (XLSX), stdlib (CSV) | Best-in-class per format |
| Period lock enforcement | Application layer + DB trigger | Belt-and-suspenders for financial integrity |
| KPI computation | Nightly Temporal workflow | Decoupled from OLTP path |

---

*Document version 2.0 — ERP Analytics Module: Complete Engineering Reference*
*Phase 1 through Phase 7 — ~12,000 words*
*For AWO ERP internal engineering reference*
