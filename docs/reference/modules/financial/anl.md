# ERP Deterministic Algorithms — Phase 1, 2 & 3

> Documentation covering use cases, how industry systems implement them, data requirements, performance considerations, and SQL/Go reference snippets for all deterministic, statistical, and rule-engine layers of an ERP system.

---

## Table of Contents

- [Phase 1 — Pure Formulas & Deterministic Logic](#phase-1--pure-formulas--deterministic-logic)
  - [1.1 Double-Entry Validation](#11-double-entry-validation)
  - [1.2 Payroll Calculations](#12-payroll-calculations)
  - [1.3 VAT Calculation & Return](#13-vat-calculation--return)
  - [1.4 Inventory Valuation](#14-inventory-valuation)
  - [1.5 Dip Variance (Fuel)](#15-dip-variance-fuel)
  - [1.6 Accounts Receivable Aging & Provision](#16-accounts-receivable-aging--provision)
- [Phase 2 — Statistical Thresholds](#phase-2--statistical-thresholds)
  - [2.1 Z-Score Anomaly Detection](#21-z-score-anomaly-detection)
  - [2.2 IQR Anomaly Detection](#22-iqr-anomaly-detection)
  - [2.3 Benford's Law Fraud Detection](#23-benfords-law-fraud-detection)
  - [2.4 Reconciliation Matching Score](#24-reconciliation-matching-score)
- [Phase 3 — Rule Engines](#phase-3--rule-engines)
  - [3.1 Approval Routing](#31-approval-routing)
  - [3.2 Reorder Point Alerts](#32-reorder-point-alerts)
  - [3.3 Credit Limit Enforcement](#33-credit-limit-enforcement)

---

# Phase 1 — Pure Formulas & Deterministic Logic

Phase 1 algorithms are pure functions: given the same inputs they always produce the same outputs. They require no training data, no external services, and no probabilistic reasoning. They are the financial foundation that everything else is built on.

**Design principle:** Every Phase 1 function must be deterministic, auditable, and reversible. If a journal entry is reversed, all computed values must return to their prior state exactly.

---

## 1.1 Double-Entry Validation

### Use Case

The oldest rule in accounting: every financial transaction must have equal debits and credits. This is the primary integrity constraint on all financial data — the ERP equivalent of a foreign key constraint.

Without this check, the general ledger can silently become unbalanced. Unbalanced ledgers produce incorrect trial balances, incorrect financial statements, and invalidate all downstream calculations.

### How Other Systems Use It

**SAP FI** enforces double-entry at the document level via `BSEG` table posting. A document cannot be saved unless the `DMBTR` (debit amount) sum equals the `HWBTR` (credit amount) sum. Posting returns `F5263` error if unbalanced.

**QuickBooks** enforces this silently — the UI hides it by auto-generating the offsetting entry, but the underlying journal validates balance before writing to disk.

**Xero** exposes it via the API: a `POST /api.xro/2.0/Journals` with unequal debits and credits returns `ValidationException: "Journal must balance"`.

**Oracle Financials** validates at subledger level before transferring to GL, so imbalance is caught at the source (AP, AR, Inventory) rather than at the GL.

### Data Required

```
journal_entries table:
  id           UUID
  journal_id   UUID         -- groups lines belonging to one transaction
  account_id   UUID
  debit        NUMERIC(19,4)
  credit       NUMERIC(19,4)
  tenant_id    UUID
  posted_at    TIMESTAMPTZ
```

### Performance & Optimisation

This check runs on every transaction — it must be sub-millisecond. A simple SUM aggregation over the lines of a pending document (never persisted lines) is sufficient. For documents with many lines (e.g. a payroll run posting 200 employee entries), ensure the aggregation is done in memory on the already-fetched line slice, not via a round-trip database query.

**Index:** Partial index on `journal_id WHERE posted_at IS NULL` for fast pre-post validation queries.

### SQL

```sql
-- Validate a journal before posting
SELECT
    journal_id,
    SUM(debit)  AS total_debit,
    SUM(credit) AS total_credit,
    ABS(SUM(debit) - SUM(credit)) AS imbalance
FROM journal_entry_lines
WHERE journal_id = $1
  AND tenant_id  = $2
GROUP BY journal_id
HAVING ABS(SUM(debit) - SUM(credit)) > 0.005; -- tolerance for float precision
```

If this query returns a row, reject the posting.

```sql
-- Trial balance check (run at period end)
SELECT
    ABS(SUM(debit) - SUM(credit)) AS ledger_imbalance
FROM journal_entry_lines jel
JOIN journals j ON j.id = jel.journal_id
WHERE j.tenant_id  = $1
  AND j.period_id  = $2
  AND j.status     = 'POSTED';
```

### Go

```go
// ValidateBalance checks that a set of lines sums to zero net.
func ValidateBalance(lines []JournalLine) error {
    var net decimal.Decimal
    for _, l := range lines {
        net = net.Add(l.Debit).Sub(l.Credit)
    }
    if net.Abs().GreaterThan(decimal.NewFromFloat(0.005)) {
        return fmt.Errorf("journal imbalance: net %s (must be zero)", net)
    }
    return nil
}
```

Use `shopspring/decimal` — never `float64` for financial arithmetic. Float64 accumulates rounding errors that compound across thousands of transactions.

---

## 1.2 Payroll Calculations

### Use Case

Computing gross-to-net pay for every employee on every payroll run, including all statutory deductions. This is arguably the most legally consequential calculation in the system: errors result in underpayment to employees, under-remittance to tax authorities, and statutory penalties.

The calculation must be deterministic and auditable — for any payslip, you must be able to reproduce the exact figures from the inputs alone.

### How Other Systems Use It

**Sage Payroll (Kenya)** maintains a rate table for each statutory deduction, updated on each regulatory change. The calculation engine is a pure function: `compute_net_pay(employee, period, rate_table) → payslip`. Sage stores a snapshot of the rate table used for each payroll run so historical payslips are reproducible even after rate changes.

**QuickBooks Payroll** uses a similar approach but ties the rates to a subscription service that pushes updated tax tables. This is operationally convenient but creates a dependency — rate updates are applied automatically without explicit operator awareness.

**Workday** treats every deduction type as a configurable calculation rule (their "Calculated Fields" engine). The rules are versioned and period-effective: a rule effective from 1 January uses the new rates; older payroll runs use the rates that were effective at the time.

**The critical pattern across all serious payroll systems:** rate tables are versioned and period-effective. The rate that was in effect when the payroll ran is stored alongside the payslip — not just the current rate.

### Data Required

```
employees:
  id, name, employee_number, department_id,
  basic_salary, hire_date, tax_pin, nssf_number

payroll_periods:
  id, start_date, end_date, status, tenant_id

payroll_lines:
  id, period_id, employee_id,
  basic_salary, allowances, overtime_hours,
  gross_pay, taxable_pay,
  paye, nssf_employee, shif, housing_levy_employee,
  total_deductions, net_pay,
  rate_snapshot_id   -- FK to the rate table snapshot used

statutory_rates:     -- versioned, period-effective
  id, effective_from, effective_to,
  nssf_tier1_rate, nssf_tier1_ceiling,
  nssf_tier2_rate, nssf_tier2_ceiling,
  shif_rate,
  housing_levy_rate,
  paye_bands          -- JSONB array of {from, to, rate}
  personal_relief,
  tenant_id           -- NULL for system-wide rates
```

### Performance & Optimisation

A payroll run for 500 employees must complete in under 30 seconds to feel responsive. Each employee's calculation is independent — this is embarrassingly parallel.

**Parallelise with goroutines:** Process employees in a worker pool. Each worker fetches the employee record, computes the full payslip in memory, and returns the result. Batch-insert all results at the end in a single transaction.

**Pre-load shared data once:** The rate table, tax bands, and department data are the same for all employees. Load them once before the pool starts — never inside the per-employee goroutine.

**Single transaction for the entire run:** All payslip inserts happen in one database transaction. If any employee fails, the entire run rolls back. Partial payroll runs are never written to disk.

### SQL

```sql
-- Fetch the rate snapshot effective for a given payroll date
SELECT *
FROM statutory_rates
WHERE (tenant_id = $1 OR tenant_id IS NULL)
  AND effective_from <= $2
  AND (effective_to IS NULL OR effective_to >= $2)
ORDER BY tenant_id NULLS LAST  -- tenant-specific overrides system-wide
LIMIT 1;
```

```sql
-- Payroll summary by department (for management report)
SELECT
    d.name                        AS department,
    COUNT(pl.id)                  AS headcount,
    SUM(pl.gross_pay)             AS total_gross,
    SUM(pl.paye)                  AS total_paye,
    SUM(pl.nssf_employee)         AS total_nssf,
    SUM(pl.shif)                  AS total_shif,
    SUM(pl.housing_levy_employee) AS total_housing_levy,
    SUM(pl.net_pay)               AS total_net
FROM payroll_lines pl
JOIN employees e  ON e.id = pl.employee_id
JOIN departments d ON d.id = e.department_id
WHERE pl.period_id = $1
  AND pl.tenant_id = $2
GROUP BY d.name
ORDER BY d.name;
```

### Go

```go
// ComputePAYE calculates Kenya PAYE from taxable monthly pay.
// bands must be sorted ascending by From.
func ComputePAYE(taxablePay decimal.Decimal, bands []TaxBand, personalRelief decimal.Decimal) decimal.Decimal {
    tax := decimal.Zero
    remaining := taxablePay

    for _, band := range bands {
        if remaining.LessThanOrEqual(decimal.Zero) {
            break
        }
        bandWidth := band.To.Sub(band.From)
        taxable := decimal.Min(remaining, bandWidth)
        tax = tax.Add(taxable.Mul(band.Rate))
        remaining = remaining.Sub(taxable)
    }

    // Any amount above the last band ceiling taxed at top rate
    if remaining.GreaterThan(decimal.Zero) {
        tax = tax.Add(remaining.Mul(bands[len(bands)-1].Rate))
    }

    net := tax.Sub(personalRelief)
    if net.LessThan(decimal.Zero) {
        return decimal.Zero
    }
    return net
}

// ComputeNSSF calculates NSSF under the 2023 Act (Tier I + Tier II).
func ComputeNSSF(grossPay decimal.Decimal, rates NSSFRates) decimal.Decimal {
    tierI := decimal.Min(grossPay, rates.Tier1Ceiling).Mul(rates.Tier1Rate)

    tier2Base := grossPay.Sub(rates.Tier1Ceiling)
    if tier2Base.LessThan(decimal.Zero) {
        tier2Base = decimal.Zero
    }
    tier2Capped := decimal.Min(tier2Base, rates.Tier2Ceiling.Sub(rates.Tier1Ceiling))
    tierII := tier2Capped.Mul(rates.Tier2Rate)

    return tierI.Add(tierII)
}
```

---

## 1.3 VAT Calculation & Return

### Use Case

Computing output VAT on sales, input VAT on purchases, and the net VAT payable (or refundable) for each VAT period. The system must handle multiple VAT categories (standard rate, zero rate, exempt), track VAT by invoice, and produce a return that matches KRA's eTIMS/iTax filing format.

### How Other Systems Use It

**Sage 50 (Kenya)** maintains a tax code table where each code has a rate, a direction (sales/purchase), and an account mapping. Every transaction line is assigned a tax code; the system aggregates by code for the return.

**Xero** handles VAT through "tax rates" applied at line-item level. The VAT return is generated by querying all transactions in the period, grouping by tax rate, and computing box values. Xero also supports VAT on a cash basis (tax point = payment date) vs. invoice basis (tax point = invoice date) — the choice affects when output VAT is recognised.

**QuickBooks** in Kenya uses the same approach with their "VAT Centre" module. A significant difference from Xero: QuickBooks locks the VAT period once the return is filed, preventing amendment of underlying transactions.

**The key design decision for any ERP:** VAT must be stored at the line level (not document level) because a single invoice can contain lines at different VAT rates (e.g. a hospitality invoice with food at 0% and alcoholic beverages at 16%).

### Data Required

```
vat_rates:
  id, code, name, rate, rate_type (STANDARD/ZERO/EXEMPT),
  direction (SALES/PURCHASE/BOTH), gl_account_id, tenant_id

transaction_lines:
  id, transaction_id, account_id,
  net_amount, vat_rate_id, vat_amount, gross_amount,
  tax_point_date, tenant_id

vat_returns:
  id, period_start, period_end, status,
  output_vat, input_vat, net_vat_payable,
  filed_at, tenant_id
```

### Performance & Optimisation

VAT return generation queries can be expensive on large transaction volumes. Key optimisations:

**Materialised view per period:** Pre-aggregate VAT line totals into a materialised view refreshed nightly. The return generation query reads from the view rather than scanning all transaction lines.

**Partial index on tax_point_date:** VAT return queries always filter by date range. A partial index on `tax_point_date` covering only unreconciled transactions significantly reduces scan cost.

**Idempotent return generation:** Generate the return as a read-only computation. Never store computed values that can be re-derived — only store the filed return (which is immutable once filed).

### SQL

```sql
-- Generate VAT return figures for a period
SELECT
    vr.code,
    vr.name,
    vr.rate,
    vr.direction,
    SUM(tl.net_amount)  AS net_of_vat,
    SUM(tl.vat_amount)  AS vat_amount,
    SUM(tl.gross_amount) AS gross_amount,
    COUNT(tl.id)         AS line_count
FROM transaction_lines tl
JOIN vat_rates vr ON vr.id = tl.vat_rate_id
JOIN transactions t ON t.id = tl.transaction_id
WHERE t.tenant_id     = $1
  AND tl.tax_point_date BETWEEN $2 AND $3
  AND t.status        = 'POSTED'
  AND vr.rate_type   != 'EXEMPT'
GROUP BY vr.code, vr.name, vr.rate, vr.direction
ORDER BY vr.direction, vr.rate DESC;
```

```sql
-- VAT payable summary
SELECT
    SUM(CASE WHEN vr.direction = 'SALES'    THEN tl.vat_amount ELSE 0 END) AS output_vat,
    SUM(CASE WHEN vr.direction = 'PURCHASE' THEN tl.vat_amount ELSE 0 END) AS input_vat,
    SUM(CASE WHEN vr.direction = 'SALES'    THEN tl.vat_amount ELSE 0 END) -
    SUM(CASE WHEN vr.direction = 'PURCHASE' THEN tl.vat_amount ELSE 0 END) AS net_payable
FROM transaction_lines tl
JOIN vat_rates vr   ON vr.id = tl.vat_rate_id
JOIN transactions t ON t.id  = tl.transaction_id
WHERE t.tenant_id        = $1
  AND tl.tax_point_date  BETWEEN $2 AND $3
  AND t.status           = 'POSTED'
  AND vr.rate_type      != 'EXEMPT';
```

### Go

```go
// ComputeVAT computes VAT amount from net or gross depending on whether
// the given amount is VAT-inclusive or exclusive.
func ComputeVAT(amount decimal.Decimal, rate decimal.Decimal, inclusive bool) VATResult {
    if inclusive {
        // Extract VAT from gross
        divisor := decimal.NewFromInt(1).Add(rate)
        net := amount.Div(divisor).Round(2)
        vat := amount.Sub(net)
        return VATResult{Net: net, VAT: vat, Gross: amount}
    }
    // Add VAT to net
    vat := amount.Mul(rate).Round(2)
    return VATResult{Net: amount, VAT: vat, Gross: amount.Add(vat)}
}
```

---

## 1.4 Inventory Valuation

### Use Case

Computing the cost of inventory on hand and the cost of goods sold (COGS) as items are sold or consumed. The valuation method determines how cost flows from goods received to goods sold — affecting both the balance sheet (inventory asset value) and the income statement (COGS and therefore gross margin).

The three standard methods are FIFO (First In, First Out), Weighted Average Cost (WAC), and LIFO (Last In, First Out — not permitted under IFRS but common in US GAAP). For fuel retail, Weighted Average is standard because fuel from different deliveries mixes physically in the same tank.

### How Other Systems Use It

**SAP MM/FI** supports all three methods, configured per material. FIFO is implemented via the "Batch" concept: each goods receipt creates a batch with its own cost, and FIFO logic consumes batches in date order. WAC (called "Moving Average Price" in SAP) is the default for raw materials and trading goods.

**Microsoft Dynamics 365** implements FIFO through "inventory layers" — each receipt creates a layer with quantity and cost. Issues consume layers FIFO. A periodic "inventory close" reconciles actual costs against interim standard costs.

**Odoo** stores WAC in the `standard_price` field on the product, updated on every goods receipt using the WAC formula. FIFO is handled via "stock valuation layers" with explicit lot/serial tracking.

**QuickBooks** uses FIFO only (desktop version), tracking each "inventory item" as a set of quantity-cost layers. Average cost is available in QuickBooks Online but calculated differently.

**The key insight from production systems:** WAC must be recalculated on every goods receipt, not periodically. A receipt that arrives between two sales must update the cost used for the second sale. For FIFO, the layer stack must be updated transactionally — no orphaned layers, no negative layers.

### Data Required

```
inventory_items:
  id, sku, name, unit_of_measure,
  valuation_method (FIFO/WAC/LIFO),
  current_quantity, current_avg_cost,  -- WAC cache
  tenant_id

inventory_layers:              -- For FIFO
  id, item_id, receipt_id,
  received_quantity, remaining_quantity,
  unit_cost, received_at,
  tenant_id

inventory_movements:
  id, item_id, movement_type (RECEIPT/ISSUE/ADJUSTMENT),
  quantity, unit_cost, total_cost,
  reference_id, movement_date, tenant_id
```

### Performance & Optimisation

FIFO is more expensive than WAC because each issue requires reading and potentially updating multiple layers. On high-volume items (a fuel station pumping 10,000 transactions per day), layer management can become a bottleneck.

**For FIFO at high volume:** Process layer consumption in the background (deferred FIFO costing). Record the issue with a preliminary cost, then a background job resolves the exact layer cost. The preliminary cost can be the current WAC — the correction is posted as a variance when the FIFO cost is resolved.

**For WAC:** The `current_avg_cost` column is a cache — always derivable from `inventory_movements`. Keep it updated on every receipt. If it ever diverges (due to a bug), the authoritative value is always the movement history.

**Index:** `inventory_layers(item_id, received_at) WHERE remaining_quantity > 0` — partial index only on layers with remaining stock, sorted by date for efficient FIFO consumption.

### SQL

```sql
-- Weighted Average Cost: update on receipt
WITH receipt AS (
    SELECT $1::uuid AS item_id,
           $2::numeric AS receipt_qty,
           $3::numeric AS receipt_unit_cost
)
UPDATE inventory_items i
SET
    current_avg_cost = (
        (i.current_quantity * i.current_avg_cost) +
        (r.receipt_qty * r.receipt_unit_cost)
    ) / NULLIF(i.current_quantity + r.receipt_qty, 0),
    current_quantity = i.current_quantity + r.receipt_qty
FROM receipt r
WHERE i.id = r.item_id
  AND i.tenant_id = $4
RETURNING current_avg_cost, current_quantity;
```

```sql
-- FIFO: consume layers in order for an issue
SELECT
    id,
    remaining_quantity,
    unit_cost
FROM inventory_layers
WHERE item_id        = $1
  AND tenant_id      = $2
  AND remaining_quantity > 0
ORDER BY received_at ASC  -- oldest first
FOR UPDATE;               -- lock for concurrent safety
```

### Go

```go
// ApplyFIFO consumes layers for an issue and returns the total cost and updated layers.
func ApplyFIFO(layers []InventoryLayer, issueQty decimal.Decimal) (decimal.Decimal, []InventoryLayer, error) {
    remaining := issueQty
    totalCost := decimal.Zero

    for i := range layers {
        if remaining.IsZero() {
            break
        }
        consume := decimal.Min(remaining, layers[i].RemainingQty)
        totalCost = totalCost.Add(consume.Mul(layers[i].UnitCost))
        layers[i].RemainingQty = layers[i].RemainingQty.Sub(consume)
        remaining = remaining.Sub(consume)
    }

    if remaining.GreaterThan(decimal.Zero) {
        return decimal.Zero, nil, fmt.Errorf("insufficient stock: %s units short", remaining)
    }
    return totalCost, layers, nil
}
```

---

## 1.5 Dip Variance (Fuel)

### Use Case

Reconciling the theoretical quantity of fuel in each tank (calculated from opening dip + deliveries − meter sales) against the physical closing dip measurement. The variance indicates measurement error, temperature expansion/contraction, equipment calibration issues, or fuel loss and theft.

This is one of the most financially significant calculations for a fuel retailer. A station pumping 20,000 litres of diesel per day with a 0.5% undetected variance loses 100 litres/day (~KES 15,000/day at current prices).

### How Other Systems Use It

**Orpak (Israel-based fuel station management)** performs dip reconciliation per shift. Each shift has an opening dip, all deliveries, all pump meter readings, and a closing dip. Variance is computed per product per tank per shift and automatically posted as a shrinkage or gain entry in the GL.

**Implant (South Africa)** uses the same calculation but adds temperature correction: raw dip volumes are corrected to a standard temperature (15°C / 60°F) before comparison, eliminating thermal expansion as a variance driver.

**Shell Retail Management Systems** set tolerance bands by product: diesel ±0.3%, petrol ±0.5%, kerosene ±0.5%. Variance within tolerance is auto-posted to a "normal shrinkage" account. Variance outside tolerance is held pending investigation before posting.

**The key operational pattern:** dip variance is always computed per product per tank per measurement period — not for the station as a whole. A station variance of zero can mask one tank gaining while another tank loses.

### Data Required

```
tanks:
  id, name, product_id, capacity_litres, tenant_id

dip_readings:
  id, tank_id, read_at, read_by,
  dip_mm, volume_litres,    -- from calibration chart
  shift_id, tenant_id

deliveries:
  id, tank_id, delivered_at,
  delivered_litres, delivery_note_no, tenant_id

pump_meter_readings:
  id, pump_id, tank_id,
  opening_meter, closing_meter,
  volume_dispensed,         -- closing − opening
  shift_id, tenant_id

tank_tolerance_rules:
  id, tank_id, tolerance_pct, tenant_id
```

### Performance & Optimisation

Dip variance computation is lightweight — it is a simple arithmetic aggregation over a small number of records per tank per shift. Performance is not a concern for the calculation itself.

**The operational challenge is data freshness:** the calculation is only as accurate as the meter readings and dip measurements fed into it. Design the input flow to make data capture easy and fast — mobile-friendly shift entry, validation of physically impossible readings (volume cannot exceed tank capacity; meter cannot decrease).

**Automated meter data ingestion:** Where pumps support electronic meter reading (pulse-output or RS-232 meters), integrate direct meter data rather than manual entry. This eliminates the most common source of dip variance data errors.

### SQL

```sql
-- Dip variance per tank for a shift
WITH
  opening AS (
      SELECT volume_litres
      FROM dip_readings
      WHERE tank_id = $1 AND shift_id = $2 AND tenant_id = $3
      ORDER BY read_at ASC LIMIT 1
  ),
  closing AS (
      SELECT volume_litres
      FROM dip_readings
      WHERE tank_id = $1 AND shift_id = $2 AND tenant_id = $3
      ORDER BY read_at DESC LIMIT 1
  ),
  deliveries AS (
      SELECT COALESCE(SUM(delivered_litres), 0) AS total_delivered
      FROM deliveries
      WHERE tank_id = $1 AND shift_id = $2 AND tenant_id = $3
  ),
  sales AS (
      SELECT COALESCE(SUM(volume_dispensed), 0) AS total_sold
      FROM pump_meter_readings
      WHERE tank_id = $1 AND shift_id = $2 AND tenant_id = $3
  )
SELECT
    o.volume_litres                                               AS opening_dip,
    d.total_delivered                                             AS deliveries,
    s.total_sold                                                  AS meter_sales,
    (o.volume_litres + d.total_delivered - s.total_sold)          AS theoretical_closing,
    c.volume_litres                                               AS actual_closing_dip,
    (o.volume_litres + d.total_delivered - s.total_sold)
        - c.volume_litres                                         AS variance_litres,
    ROUND(
        ((o.volume_litres + d.total_delivered - s.total_sold)
            - c.volume_litres)
        / NULLIF(o.volume_litres + d.total_delivered - s.total_sold, 0) * 100,
    3)                                                            AS variance_pct
FROM opening o, closing c, deliveries d, sales s;
```

### Go

```go
type DipVarianceResult struct {
    OpeningDip         decimal.Decimal
    Deliveries         decimal.Decimal
    MeterSales         decimal.Decimal
    TheoreticalClosing decimal.Decimal
    ActualClosing      decimal.Decimal
    VarianceLitres     decimal.Decimal
    VariancePct        decimal.Decimal
    WithinTolerance    bool
}

func ComputeDipVariance(opening, deliveries, meterSales, closingDip, tolerancePct decimal.Decimal) DipVarianceResult {
    theoretical := opening.Add(deliveries).Sub(meterSales)
    variance    := theoretical.Sub(closingDip)
    
    var variancePct decimal.Decimal
    if theoretical.IsPositive() {
        variancePct = variance.Div(theoretical).Mul(decimal.NewFromInt(100))
    }

    return DipVarianceResult{
        OpeningDip:         opening,
        Deliveries:         deliveries,
        MeterSales:         meterSales,
        TheoreticalClosing: theoretical,
        ActualClosing:      closingDip,
        VarianceLitres:     variance,
        VariancePct:        variancePct,
        WithinTolerance:    variancePct.Abs().LessThanOrEqual(tolerancePct),
    }
}
```

---

## 1.6 Accounts Receivable Aging & Provision

### Use Case

Classifying outstanding customer invoices by how long they have been overdue, and computing a provision for expected bad debts. Aging is the primary tool for managing credit risk and collections — it tells the business who owes what and for how long.

The provision (allowance for doubtful debts) is an accounting estimate required by IFRS 9 (Expected Credit Loss model). It reduces the gross receivables balance to net realisable value on the balance sheet.

### How Other Systems Use It

**Sage Business Cloud** runs aging automatically on demand. The user can age by invoice date or due date (the latter being the operationally meaningful choice for collections). Provision rates are configured per aging bucket by the user and applied to produce the allowance journal.

**SAP AR** implements aging via the `S_ALR_87012178` standard report. SAP ages in arrears: a 30-day bucket means 1–30 days past due date, not 1–30 days old. This is the correct approach — age relative to due date, not invoice date.

**Xero** does not compute provisions automatically — it shows aging but leaves the provision calculation to the accountant. Many small business users do not provision at all, which results in overstated receivables.

**Best practice from Tier 1 ERP:** Age relative to **due date**, not invoice date. Include a "not yet due" bucket for current invoices. Run the aging report at the end of each business day (not just at month end) so collections can act immediately on newly overdue accounts.

### Data Required

```
invoices:
  id, customer_id, invoice_date, due_date,
  total_amount, amount_paid, amount_outstanding,
  status (OPEN/PARTIAL/PAID/WRITTEN_OFF),
  tenant_id

customers:
  id, name, credit_limit, credit_terms_days, tenant_id

provision_rates:
  id, bucket_label, days_from, days_to, provision_rate, tenant_id
```

### Performance & Optimisation

Aging is typically run at month end and on demand — it does not need to be real-time. For a business with 10,000 open invoices, a well-indexed query returns results in under a second.

**Critical index:** `invoices(tenant_id, status, due_date)` — every aging query filters on all three. A composite index on these columns eliminates full table scans.

**Materialised view for dashboard:** Pre-compute the aging summary (bucket totals) in a materialised view, refreshed nightly. The detail query (invoice-level aging) runs on demand against the live table.

### SQL

```sql
-- Invoice-level aging as of today
SELECT
    c.name                    AS customer,
    i.id                      AS invoice_id,
    i.invoice_date,
    i.due_date,
    i.total_amount,
    i.amount_outstanding,
    CURRENT_DATE - i.due_date AS days_overdue,
    CASE
        WHEN i.due_date >= CURRENT_DATE              THEN 'CURRENT'
        WHEN CURRENT_DATE - i.due_date BETWEEN 1  AND 30  THEN '1-30'
        WHEN CURRENT_DATE - i.due_date BETWEEN 31 AND 60  THEN '31-60'
        WHEN CURRENT_DATE - i.due_date BETWEEN 61 AND 90  THEN '61-90'
        WHEN CURRENT_DATE - i.due_date BETWEEN 91 AND 120 THEN '91-120'
        ELSE '120+'
    END                       AS aging_bucket
FROM invoices i
JOIN customers c ON c.id = i.customer_id
WHERE i.tenant_id = $1
  AND i.status IN ('OPEN', 'PARTIAL')
ORDER BY days_overdue DESC;
```

```sql
-- Provision calculation using configured rates
SELECT
    aging.bucket,
    SUM(aging.outstanding)                              AS bucket_total,
    pr.provision_rate,
    SUM(aging.outstanding) * pr.provision_rate          AS provision_amount
FROM (
    SELECT
        CASE
            WHEN CURRENT_DATE - due_date <= 0             THEN 'CURRENT'
            WHEN CURRENT_DATE - due_date BETWEEN 1 AND 30  THEN '1-30'
            WHEN CURRENT_DATE - due_date BETWEEN 31 AND 60 THEN '31-60'
            WHEN CURRENT_DATE - due_date BETWEEN 61 AND 90 THEN '61-90'
            ELSE '90+'
        END       AS bucket,
        SUM(amount_outstanding) AS outstanding
    FROM invoices
    WHERE tenant_id = $1 AND status IN ('OPEN', 'PARTIAL')
    GROUP BY 1
) aging
JOIN provision_rates pr
  ON pr.bucket_label = aging.bucket AND pr.tenant_id = $1
GROUP BY aging.bucket, pr.provision_rate;
```

### Go

```go
// AgeBucket returns the aging bucket label for a given due date.
func AgeBucket(dueDate time.Time, asOf time.Time) string {
    days := int(asOf.Sub(dueDate).Hours() / 24)
    switch {
    case days <= 0:
        return "CURRENT"
    case days <= 30:
        return "1-30"
    case days <= 60:
        return "31-60"
    case days <= 90:
        return "61-90"
    case days <= 120:
        return "91-120"
    default:
        return "120+"
    }
}

// ComputeProvision applies provision rates to aging buckets.
func ComputeProvision(buckets map[string]decimal.Decimal, rates map[string]decimal.Decimal) decimal.Decimal {
    total := decimal.Zero
    for bucket, balance := range buckets {
        if rate, ok := rates[bucket]; ok {
            total = total.Add(balance.Mul(rate))
        }
    }
    return total.Round(2)
}
```

---

# Phase 2 — Statistical Thresholds

Phase 2 algorithms require no ML training but do require data history. They build a statistical understanding of "normal" from observed data, then flag deviations. They are the correct approach for anomaly detection before there is enough labeled data to train a model — and often remain the right approach permanently because they are transparent and auditable.

**Design principle:** Every Phase 2 flag must be explainable in one sentence: "This transaction is X standard deviations above the mean for this account and user role over the past 90 days."

---

## 2.1 Z-Score Anomaly Detection

### Use Case

Flagging individual transaction amounts that deviate significantly from the historical norm for that transaction type, account, and user combination. The Z-score measures how many standard deviations a value is from the mean of its reference population.

Applied in ERP to: unusual payment amounts, atypical expense claims, abnormal fuel sales volumes per shift, large journal entries from low-authority users.

### How Other Systems Use It

**Concur Expense** uses a simplified Z-score variant on expense submissions. It computes a per-employee, per-expense-category baseline and flags submissions that exceed 2.5 standard deviations. The flag generates a "policy alert" that routes the expense to additional review.

**HSBC's payment fraud detection** (retail banking, not ERP) uses a Z-score as the first-pass filter before more complex models. Transactions within 3σ of the customer's normal payment distribution are auto-cleared; those outside proceed to ML scoring. The Z-score filter handles ~85% of volume, keeping the ML scorer focused on genuine edge cases.

**NetSuite** exposes transaction limits (hard limits, not statistical) but does not compute Z-scores natively. Most NetSuite implementations add statistical anomaly detection via a BI tool (Tableau, Power BI) sitting on the NetSuite data export.

**The production pattern:** Z-scores are always computed within a reference population — not globally. "This payment is 4σ above normal for this vendor" is meaningful. "This payment is 4σ above the mean of all payments" is not, because the population is too heterogeneous.

### Data Required

```
-- Z-score baselines (pre-computed, refreshed nightly)
transaction_baselines:
  id, tenant_id,
  dimension_type (ACCOUNT/USER_ROLE/VENDOR/ACCOUNT+USER_ROLE),
  dimension_key  (the grouping key value),
  period_days    (rolling window: 30/60/90),
  sample_count,
  mean_amount,
  stddev_amount,
  computed_at
```

### Performance & Optimisation

Z-score computation at transaction time must be sub-100ms. Pre-computing the baselines nightly and storing them in `transaction_baselines` means the runtime check is a single indexed lookup + one arithmetic operation — not an aggregation query.

**Refresh strategy:** Run baseline recomputation as a nightly background job. Each baseline covers a configurable rolling window (default 90 days). For tenants with fewer than 30 transactions in the window, mark the baseline as insufficient and skip Z-score flagging (not enough data for a reliable estimate).

**Minimum sample size:** Never compute a Z-score on fewer than 30 samples. Below this, the mean and standard deviation estimates are unreliable and will generate excessive false positives.

### SQL

```sql
-- Compute baselines per account per tenant (run nightly)
INSERT INTO transaction_baselines
    (tenant_id, dimension_type, dimension_key, period_days,
     sample_count, mean_amount, stddev_amount, computed_at)
SELECT
    jel.tenant_id,
    'ACCOUNT',
    jel.account_id::text,
    90,
    COUNT(*),
    AVG(jel.debit + jel.credit),
    STDDEV_POP(jel.debit + jel.credit),
    NOW()
FROM journal_entry_lines jel
JOIN journals j ON j.id = jel.journal_id
WHERE j.posted_at >= NOW() - INTERVAL '90 days'
  AND j.status = 'POSTED'
GROUP BY jel.tenant_id, jel.account_id
HAVING COUNT(*) >= 30
ON CONFLICT (tenant_id, dimension_type, dimension_key, period_days)
DO UPDATE SET
    sample_count  = EXCLUDED.sample_count,
    mean_amount   = EXCLUDED.mean_amount,
    stddev_amount = EXCLUDED.stddev_amount,
    computed_at   = EXCLUDED.computed_at;
```

```sql
-- Runtime Z-score check for an incoming transaction
SELECT
    ABS($1 - mean_amount) / NULLIF(stddev_amount, 0) AS z_score,
    mean_amount,
    stddev_amount,
    sample_count
FROM transaction_baselines
WHERE tenant_id      = $2
  AND dimension_type = 'ACCOUNT'
  AND dimension_key  = $3   -- account_id
  AND period_days    = 90;
```

### Go

```go
// ZScore computes the Z-score of a value against a pre-loaded baseline.
// Returns the score and whether it exceeds the threshold.
func ZScore(value, mean, stddev decimal.Decimal, threshold float64) (float64, bool) {
    if stddev.IsZero() {
        return 0, false
    }
    z, _ := value.Sub(mean).Abs().Div(stddev).Float64()
    return z, z > threshold
}

// Example usage in a pre-insert hook:
func (s *AnomalyService) CheckTransaction(ctx context.Context, tenantID, accountID string, amount decimal.Decimal) *AnomalyFlag {
    baseline, err := s.repo.GetBaseline(ctx, tenantID, "ACCOUNT", accountID)
    if err != nil || baseline.SampleCount < 30 {
        return nil // insufficient data, skip
    }
    z, flagged := ZScore(amount, baseline.Mean, baseline.Stddev, 3.0)
    if !flagged {
        return nil
    }
    return &AnomalyFlag{
        ZScore:      z,
        Explanation: fmt.Sprintf("Amount %.2f is %.1fσ above the %.0f-day mean of %.2f for this account", amount, z, 90.0, baseline.Mean),
        Severity:    severityFromZScore(z),
    }
}
```

---

## 2.2 IQR Anomaly Detection

### Use Case

An alternative to Z-score that is more robust when the data contains legitimate extreme values (outliers). Financial data almost always has heavy tails — occasional legitimate very large transactions — which inflates the standard deviation and makes Z-score overly permissive.

IQR (Interquartile Range) uses the middle 50% of data to define the "normal" range, so extreme legitimate values do not corrupt the baseline.

**Use Z-score when:** The distribution is reasonably bell-shaped and outliers are rare.
**Use IQR when:** The data has heavy tails, or you want a flagging system that is robust to the presence of occasional large-but-legitimate transactions.

### How Other Systems Use It

**Palantir Foundry** (used by enterprise finance teams) uses Tukey's fences (IQR-based) as the default outlier detection method for financial dashboards, citing robustness as the primary reason over Z-score.

**Tableau** includes IQR as a built-in option in its "Mark Outliers" analytics feature, specifically noting that it is preferred for right-skewed distributions — which describes most financial data (many small transactions, few large ones).

**Internal audit software (ACL/Arbutus)** uses IQR-based outlier detection as a standard test in AP fraud reviews. The test flags expense claims and payment amounts outside the Tukey fences for the vendor or expense category.

### Data Required

Same as Z-score baselines, but storing Q1 and Q3 instead of mean and stddev:

```
transaction_baselines:
  ...existing fields...
  q1_amount     NUMERIC(19,4),
  q3_amount     NUMERIC(19,4),
  iqr_amount    NUMERIC(19,4),  -- Q3 - Q1
  lower_fence   NUMERIC(19,4),  -- Q1 - 1.5*IQR
  upper_fence   NUMERIC(19,4)   -- Q3 + 1.5*IQR
```

### Performance & Optimisation

PostgreSQL's `percentile_cont` function computes quartiles efficiently using a single pass over the data. Include IQR computation in the same nightly job as Z-score baselines — they query the same underlying data.

For very high volume tables (millions of rows), consider computing baselines on a random sample (e.g. 10,000 rows) rather than the full population. Quartile estimates from a 10,000-row sample are statistically identical to the full population for this purpose.

### SQL

```sql
-- Compute IQR baselines (include in nightly baseline job)
SELECT
    tenant_id,
    account_id,
    percentile_cont(0.25) WITHIN GROUP (ORDER BY amount) AS q1,
    percentile_cont(0.75) WITHIN GROUP (ORDER BY amount) AS q3,
    percentile_cont(0.75) WITHIN GROUP (ORDER BY amount) -
    percentile_cont(0.25) WITHIN GROUP (ORDER BY amount) AS iqr
FROM (
    SELECT jel.tenant_id, jel.account_id, (jel.debit + jel.credit) AS amount
    FROM journal_entry_lines jel
    JOIN journals j ON j.id = jel.journal_id
    WHERE j.posted_at >= NOW() - INTERVAL '90 days'
      AND j.status = 'POSTED'
) t
GROUP BY tenant_id, account_id
HAVING COUNT(*) >= 30;

-- Fences derived as:
-- lower_fence = q1 - 1.5 * iqr
-- upper_fence = q3 + 1.5 * iqr
```

### Go

```go
type IQRBaseline struct {
    Q1, Q3, IQR, LowerFence, UpperFence decimal.Decimal
}

func CheckIQR(amount decimal.Decimal, b IQRBaseline) (bool, string) {
    if amount.LessThan(b.LowerFence) {
        return true, fmt.Sprintf("Amount %.2f is below the lower fence (%.2f)", amount, b.LowerFence)
    }
    if amount.GreaterThan(b.UpperFence) {
        return true, fmt.Sprintf("Amount %.2f exceeds the upper fence (%.2f)", amount, b.UpperFence)
    }
    return false, ""
}
```

---

## 2.3 Benford's Law Fraud Detection

### Use Case

Benford's Law states that in naturally occurring numerical datasets, the leading digit follows a predictable logarithmic distribution: 1 appears as the leading digit ~30% of the time, 2 appears ~18%, and so on down to 9 at ~5%. When humans fabricate numbers, they tend to distribute leading digits more uniformly — which produces a statistically detectable signature.

Benford's analysis is most powerful on: expense claims, supplier invoices, purchase orders, payroll figures, and journal entries — any dataset where a human might be inventing or inflating numbers.

### How Other Systems Use It

**ACL Analytics (now Galvanize/Diligent)** includes Benford analysis as a standard audit test. The tool computes the chi-squared statistic for the first digit, first two digits, and last two digits, and flags datasets where the distribution significantly deviates from expected.

**KPMG's audit data analytics platform** applies Benford's Law as a first-pass filter on every GL dataset in an audit engagement. Deviating datasets are escalated to detailed transaction testing.

**The IRS (US)** uses Benford's Law in automated screening of tax returns and reported income figures. Research published by the IRS confirms that Benford deviations correlate with audit adjustments.

**Important limitation:** Benford's Law does not work on: constrained datasets (items where prices have fixed ranges), sequences (invoice numbers, employee IDs), or datasets with fewer than ~500 records. Always check dataset suitability before applying.

### Data Required

```
-- Any amount column with 500+ rows is sufficient
-- Apply to: invoices.total_amount, expense_claims.amount,
--           journal_entries.amount, payroll_lines.gross_pay
```

No additional schema needed — Benford analysis queries existing transaction tables.

### SQL

```sql
-- Benford first-digit analysis for expense claims
WITH digit_counts AS (
    SELECT
        LEFT(CAST(ABS(amount) AS TEXT), 1)::int AS first_digit,
        COUNT(*) AS actual_count,
        COUNT(*) * 100.0 / SUM(COUNT(*)) OVER () AS actual_pct
    FROM expense_claims
    WHERE tenant_id = $1
      AND submitted_at >= NOW() - INTERVAL '12 months'
      AND amount > 0
    GROUP BY 1
),
expected AS (
    SELECT
        generate_series AS first_digit,
        LOG(1 + 1.0/generate_series) / LOG(10) * 100 AS expected_pct
    FROM generate_series(1, 9)
)
SELECT
    d.first_digit,
    d.actual_count,
    ROUND(d.actual_pct, 2)    AS actual_pct,
    ROUND(e.expected_pct, 2)  AS expected_pct,
    ROUND(d.actual_pct - e.expected_pct, 2) AS deviation_pct
FROM digit_counts d
JOIN expected e USING (first_digit)
ORDER BY first_digit;
```

### Go

```go
// BenfordExpected returns the expected Benford frequency for digit d (1-9).
func BenfordExpected(d int) float64 {
    return math.Log10(1 + 1.0/float64(d))
}

// ChiSquaredBenford computes the chi-squared statistic for a digit frequency map.
// A result > 15.51 (8 degrees of freedom, p=0.05) suggests significant deviation.
func ChiSquaredBenford(observed map[int]int, total int) float64 {
    chi2 := 0.0
    for d := 1; d <= 9; d++ {
        expected := BenfordExpected(d) * float64(total)
        actual   := float64(observed[d])
        diff     := actual - expected
        chi2    += (diff * diff) / expected
    }
    return chi2
}
```

---

## 2.4 Reconciliation Matching Score

### Use Case

Automatically matching bank statement lines to general ledger entries during bank reconciliation. Instead of requiring an accountant to manually match each line, the system scores every potential GL-to-bank-statement pair and proposes matches above a confidence threshold.

This is deterministic (no training required) because the matching criteria are explicit: amount match, date proximity, and reference similarity. The score is a weighted combination of these criteria.

### How Other Systems Use It

**Xero's bank reconciliation** uses a matching algorithm that considers amount (required exact match or near-match), date (within configurable window), and payee/description similarity. Xero auto-matches when it finds a single GL entry matching all criteria; otherwise it presents the top 3 candidates for the accountant to choose from.

**QuickBooks** uses a similar approach but also learns from the accountant's historical matching decisions — if the accountant repeatedly matches "SHELL KE" bank descriptions to the "Fuel Purchases" account, QuickBooks begins suggesting this match automatically. This is the boundary between Phase 2 (deterministic scoring) and Phase 4 (ML categorization).

**SAP Bank Communication Management** uses a scoring engine with configurable weights. SAP implementations typically tune the weights for their specific bank format — a bank that includes exact invoice numbers in the reference field gets a much higher weight on reference matching.

### Data Required

```
bank_statement_lines:
  id, statement_id, line_date, description,
  reference, debit_amount, credit_amount, tenant_id

gl_entries_unreconciled:
  id, account_id, entry_date, description,
  reference, amount, tenant_id

bank_reconciliation_matches:
  id, bank_line_id, gl_entry_id,
  match_score, matched_by (USER/AUTO),
  matched_at, tenant_id
```

### Performance & Optimisation

The naive approach (score every bank line against every GL entry) is O(n×m) — expensive when both sets are large. For a business with 500 bank lines and 2,000 unreconciled GL entries per month, that is 1,000,000 pair comparisons.

**Optimise with candidate filtering:** Before scoring, filter the GL entry candidates for each bank line to a small set using amount range and date window filters. Then score only within this candidate set.

**Amount window:** Only compare GL entries where `|gl_amount - bank_amount| ≤ max_tolerance`.
**Date window:** Only compare GL entries within ±7 days of the bank line date.

This reduces the comparison set from thousands to typically 1–5 candidates per bank line, making the scoring trivially fast.

### SQL

```sql
-- Find candidate GL matches for a bank statement line
SELECT
    gl.id,
    gl.entry_date,
    gl.description,
    gl.reference,
    gl.amount,
    -- Amount score: 1.0 if exact match, decays with difference
    CASE
        WHEN ABS(gl.amount - $2) = 0                   THEN 1.0
        WHEN ABS(gl.amount - $2) <= $3                 THEN
            1.0 - (ABS(gl.amount - $2) / $3::numeric)
        ELSE 0.0
    END AS amount_score,
    -- Date score: 1.0 if same day, decays over window
    GREATEST(0, 1.0 - ABS(gl.entry_date - $4::date)::numeric / 7.0) AS date_score
FROM gl_entries_unreconciled gl
WHERE gl.tenant_id = $1
  AND ABS(gl.amount - $2) <= $3      -- amount tolerance
  AND gl.entry_date BETWEEN $4::date - 7 AND $4::date + 7  -- date window
  AND gl.reconciled_at IS NULL
ORDER BY amount_score DESC, date_score DESC
LIMIT 5;
```

### Go

```go
// MatchScore computes a composite matching score [0,1] for a GL/bank pair.
func MatchScore(bank BankLine, gl GLEntry, toleranceAmt decimal.Decimal, weights MatchWeights) float64 {
    // Amount score
    diff := bank.Amount.Sub(gl.Amount).Abs()
    amountScore := 0.0
    if diff.IsZero() {
        amountScore = 1.0
    } else if diff.LessThanOrEqual(toleranceAmt) {
        ratio, _ := diff.Div(toleranceAmt).Float64()
        amountScore = 1.0 - ratio
    }

    // Date score: 1.0 if same day, 0 at 7 days
    daysDiff := math.Abs(bank.Date.Sub(gl.Date).Hours() / 24)
    dateScore := math.Max(0, 1.0-daysDiff/7.0)

    // Reference similarity (simple character overlap)
    refScore := stringSimilarity(bank.Reference, gl.Reference)

    return weights.Amount*amountScore +
           weights.Date*dateScore +
           weights.Reference*refScore
}
```

---

# Phase 3 — Rule Engines

Phase 3 algorithms encode business policy as explicit, configurable rules. Unlike Phase 1 (hardcoded formulas) and Phase 2 (statistical), Phase 3 rules are maintained by business administrators — not developers. They must be readable, configurable, and auditable.

**Design principle:** Every rule must be inspectable by a non-developer. "Purchases over KES 50,000 require Finance Director approval" must be stored as data — not buried in application code.

---

## 3.1 Approval Routing

### Use Case

Determining who must approve a transaction (purchase requisition, payment, journal entry, expense claim, leave request) before it is posted or executed. The routing logic is driven by configurable rules: amount thresholds, transaction types, cost centre, department, and user role.

Approval workflows are how the ERP enforces internal controls. A system without approval routing has no segregation of duties and no purchase authorization framework — critical weaknesses for any audited business.

### How Other Systems Use It

**SAP Workflow** implements approval routing through "Workflow Tasks" linked to business objects. Each task specifies an agent (person or role) determined at runtime by a "rule" that evaluates transaction attributes. SAP's rule engine is powerful but complex — configuration requires a workflow consultant.

**Microsoft Power Automate** (commonly used with Dynamics 365) implements approval flows as visual diagrams. Conditions are evaluated sequentially; each branch specifies the approver. Multi-level approval is modeled as sequential stages.

**Coupa (procurement)** uses an "Approval Chain" model where each step in a chain has an approver type (specific user, role, group) and a condition (amount band, supplier type, commodity). Chains are configured by business administrators via a UI, not code.

**Kissflow / Zoho Approval** take the no-code approach: approval rules are configured in a table-driven UI. This makes the rules maximally accessible but limits flexibility for complex conditions.

**The pattern that scales:** Store approval rules as data in a configurable rule table. Evaluate rules at runtime against the transaction's attributes. Never hardcode approval logic — policy changes must be configuration changes, not deployments.

### Data Required

```
approval_policies:
  id, name, entity_type (PURCHASE/PAYMENT/JOURNAL/EXPENSE),
  is_active, tenant_id

approval_rules:
  id, policy_id, sequence_number,
  condition_field  (e.g. 'amount', 'department_id', 'vendor_category'),
  condition_op     (GT/LT/EQ/IN/BETWEEN),
  condition_value  TEXT,   -- JSON-encoded value
  approver_type    (USER/ROLE/DEPARTMENT_HEAD/COST_CENTRE_OWNER),
  approver_ref     TEXT,   -- user_id or role name
  escalation_days  INT,    -- auto-escalate after N days
  tenant_id

approval_instances:
  id, policy_id, entity_type, entity_id,
  current_step, status (PENDING/APPROVED/REJECTED/ESCALATED),
  initiated_at, tenant_id

approval_steps:
  id, instance_id, step_number,
  approver_id, status, decided_at, comment, tenant_id
```

### Performance & Optimisation

Approval routing runs once per transaction submission — performance is not critical. The expensive part is rule evaluation at scale (e.g. generating an approval chain for a payroll run with 300 expense claims simultaneously).

**Batch routing:** For batch submissions (period-end accruals, payroll), generate all approval chains in a single pass. Load the policy and rules once; evaluate each transaction against the in-memory rule set.

**Rule evaluation order matters:** Sort rules by `sequence_number`. Evaluate in order and stop at the first matching rule. This gives administrators explicit control over precedence.

**Cache active policies:** Approval policies change infrequently. Cache the active policy and its rules in Redis with a TTL of 5 minutes. Cache invalidation on any policy update.

### SQL

```sql
-- Fetch the active approval policy and its rules for an entity type
SELECT
    ap.id   AS policy_id,
    ar.id   AS rule_id,
    ar.sequence_number,
    ar.condition_field,
    ar.condition_op,
    ar.condition_value,
    ar.approver_type,
    ar.approver_ref,
    ar.escalation_days
FROM approval_policies ap
JOIN approval_rules ar ON ar.policy_id = ap.id
WHERE ap.tenant_id   = $1
  AND ap.entity_type = $2
  AND ap.is_active   = TRUE
ORDER BY ar.sequence_number ASC;
```

```sql
-- Pending approvals for a user (their action queue)
SELECT
    ai.id           AS instance_id,
    ai.entity_type,
    ai.entity_id,
    ai.current_step,
    ast.step_number,
    ast.initiated_at,
    CURRENT_DATE - ast.initiated_at::date AS days_waiting
FROM approval_instances ai
JOIN approval_steps ast ON ast.instance_id = ai.id
    AND ast.step_number = ai.current_step
WHERE ast.approver_id = $1
  AND ast.status      = 'PENDING'
  AND ai.tenant_id    = $2
ORDER BY days_waiting DESC;
```

### Go

```go
// EvaluateRule checks a single approval rule against a transaction's attributes.
func EvaluateRule(rule ApprovalRule, attrs map[string]any) bool {
    value, ok := attrs[rule.ConditionField]
    if !ok {
        return false
    }
    switch rule.ConditionOp {
    case "GT":
        return toDecimal(value).GreaterThan(toDecimal(rule.ConditionValue))
    case "LT":
        return toDecimal(value).LessThan(toDecimal(rule.ConditionValue))
    case "EQ":
        return fmt.Sprintf("%v", value) == rule.ConditionValue
    case "IN":
        var allowed []string
        json.Unmarshal([]byte(rule.ConditionValue), &allowed)
        return slices.Contains(allowed, fmt.Sprintf("%v", value))
    }
    return false
}

// BuildApprovalChain evaluates all rules and returns the ordered approver list.
func BuildApprovalChain(rules []ApprovalRule, attrs map[string]any) []ApprovalStep {
    var chain []ApprovalStep
    for _, rule := range rules { // rules pre-sorted by sequence_number
        if EvaluateRule(rule, attrs) {
            chain = append(chain, ApprovalStep{
                ApproverType:   rule.ApproverType,
                ApproverRef:    rule.ApproverRef,
                EscalationDays: rule.EscalationDays,
            })
        }
    }
    return chain
}
```

---

## 3.2 Reorder Point Alerts

### Use Case

Automatically detecting when an inventory item's stock level has fallen to or below its reorder point, and generating a purchase requisition or alert to prompt replenishment. This prevents stockouts — one of the most operationally costly failures in inventory management.

The reorder point itself is calculated from a formula (Phase 1). The alert engine is Phase 3: a configurable rule that fires when the stock level crosses the threshold, routes to the appropriate buyer, and prevents duplicate alerts.

### How Other Systems Use It

**SAP MRP (Materials Requirements Planning)** runs a nightly batch that compares on-hand stock + confirmed receipts − confirmed issues against the reorder point for every material at every plant. When stock falls below the reorder point, SAP generates a "planned order" or "purchase requisition" automatically.

**Cin7** (popular in East Africa for inventory management) uses a "reorder level" and "reorder quantity" per item. When stock falls below the reorder level during any stock movement, Cin7 generates a "purchase order suggestion" visible to the procurement team.

**Odoo Inventory** implements this as a "reordering rule" that can be evaluated on-demand or automatically. The rule specifies minimum quantity, maximum quantity (target after reorder), and the preferred supplier. When triggered, it auto-creates a purchase order draft.

**The key design patterns from production systems:**
1. Check after every stock movement — not just in nightly batches. A sale that pushes stock below reorder should alert immediately.
2. Consider "available stock" (on-hand + confirmed pending receipts − confirmed pending issues), not just on-hand stock.
3. One alert per item per crossing — do not re-alert on every subsequent movement while stock remains below reorder.

### Data Required

```
reorder_rules:
  id, item_id, location_id,
  reorder_point, reorder_quantity,
  preferred_supplier_id,
  lead_time_days,
  alert_recipient_id,    -- user or role to notify
  auto_create_pr,        -- boolean: auto-create purchase requisition
  is_active,
  tenant_id

reorder_alerts:
  id, rule_id, item_id,
  triggered_at,
  stock_at_trigger,
  pr_created_id,         -- FK to purchase_requisitions if auto-created
  acknowledged_at,
  tenant_id

-- View: available stock
-- = on_hand + pending_receipts - pending_issues
```

### Performance & Optimisation

The reorder check runs on every stock movement. For a high-volume business (thousands of movements per day), this check must be extremely fast.

**Post-movement trigger:** Implement the reorder check as a PostgreSQL trigger on the `inventory_movements` table. The trigger evaluates the reorder rule inline with the movement, eliminating a round-trip from the application layer.

**Deduplication:** Use a partial index on `reorder_alerts(item_id, acknowledged_at IS NULL)` to efficiently check whether an open alert already exists. Do not create a new alert if one is already open for this item.

**Async PR creation:** Creating a purchase requisition is a heavier operation (multiple inserts, approval chain creation). Do this asynchronously via a Temporal workflow triggered by the alert, rather than inline with the stock movement.

### SQL

```sql
-- Reorder check trigger function (runs after each stock movement)
CREATE OR REPLACE FUNCTION check_reorder_point()
RETURNS TRIGGER AS $$
DECLARE
    v_available_stock NUMERIC;
    v_rule            reorder_rules%ROWTYPE;
    v_open_alert      UUID;
BEGIN
    -- Only check for outbound movements
    IF NEW.movement_type NOT IN ('ISSUE', 'SALE') THEN
        RETURN NEW;
    END IF;

    SELECT * INTO v_rule
    FROM reorder_rules
    WHERE item_id = NEW.item_id AND is_active = TRUE
    LIMIT 1;

    IF NOT FOUND THEN RETURN NEW; END IF;

    -- Compute available stock
    SELECT
        COALESCE(SUM(CASE WHEN movement_type IN ('RECEIPT','ADJUSTMENT_IN') THEN quantity ELSE -quantity END), 0)
    INTO v_available_stock
    FROM inventory_movements
    WHERE item_id = NEW.item_id AND tenant_id = NEW.tenant_id;

    IF v_available_stock > v_rule.reorder_point THEN
        RETURN NEW;
    END IF;

    -- Check for existing open alert
    SELECT id INTO v_open_alert
    FROM reorder_alerts
    WHERE item_id = NEW.item_id AND acknowledged_at IS NULL
    LIMIT 1;

    IF FOUND THEN RETURN NEW; END IF;

    -- Insert new alert
    INSERT INTO reorder_alerts (id, rule_id, item_id, triggered_at, stock_at_trigger, tenant_id)
    VALUES (gen_random_uuid(), v_rule.id, NEW.item_id, NOW(), v_available_stock, NEW.tenant_id);

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_reorder_check
AFTER INSERT ON inventory_movements
FOR EACH ROW EXECUTE FUNCTION check_reorder_point();
```

### Go

```go
// ProcessReorderAlert handles a new reorder alert from the database trigger.
// Called by a Temporal activity triggered by a pg_notify event.
func (s *ReorderService) ProcessReorderAlert(ctx context.Context, alertID string) error {
    alert, err := s.repo.GetAlert(ctx, alertID)
    if err != nil {
        return err
    }
    rule, err := s.repo.GetRule(ctx, alert.RuleID)
    if err != nil {
        return err
    }

    // Notify buyer
    s.notifier.Send(ctx, Notification{
        RecipientID: rule.AlertRecipientID,
        Title:       "Reorder Required",
        Body:        fmt.Sprintf("%s is below reorder point (%.0f units available, reorder at %.0f)", alert.ItemName, alert.StockAtTrigger, rule.ReorderPoint),
        Link:        fmt.Sprintf("/inventory/items/%s", alert.ItemID),
    })

    // Auto-create purchase requisition if configured
    if rule.AutoCreatePR {
        prID, err := s.prService.CreateFromReorderRule(ctx, rule, alert)
        if err != nil {
            return err
        }
        return s.repo.LinkAlertToPR(ctx, alertID, prID)
    }
    return nil
}
```

---

## 3.3 Credit Limit Enforcement

### Use Case

Preventing the creation of sales orders, invoices, or fuel deliveries on credit when a customer's outstanding balance exceeds their authorised credit limit. This is the primary financial control protecting the business from excessive credit exposure.

The rule engine determines: what the limit is, what counts against it (outstanding invoices only, or including unposted orders), what happens when the limit is exceeded (hard block, soft warning, or route for override approval), and who can override.

### How Other Systems Use It

**SAP SD (Sales & Distribution)** implements credit checks via the "Credit Management" module. SAP supports three credit check types: static (compare against credit limit), dynamic (add open orders to the exposure), and maximum document value (limit per single order). When a credit check fails, the order is blocked and routed to the credit manager's worklist.

**Oracle Order Management** has a similar "Credit Check" step in the order flow. Oracle distinguishes between "exposure" (what's committed but not yet invoiced) and "outstanding" (invoiced but not paid). The total exposure against the credit limit determines whether the order is held.

**Cin7 / Dear Inventory** implements a simpler hard-block: if a new order would take the customer's outstanding balance above their credit limit, the order cannot be saved until either the limit is increased or an existing invoice is paid.

**Fuel station context (relevant for AWO ERP):** Fleet customers draw fuel against a credit account. The credit check must happen at the pump (or at order entry for bulk orders) — not at invoice time. Invoicing happens after the fuel has already been dispensed.

**The critical design choice:** Hard block vs. soft warning vs. approval routing. Hard blocks prevent the sale even if the customer is generally reliable and just had a one-off large draw. Approval routing is more flexible: a pump supervisor can override for a known good customer, with the override logged and visible to the credit manager. Hard blocks are appropriate for new or high-risk customers; approval routing for established accounts.

### Data Required

```
customer_credit_profiles:
  id, customer_id,
  credit_limit,
  credit_terms_days,
  credit_check_type (HARD_BLOCK / SOFT_WARN / APPROVAL_REQUIRED),
  include_open_orders,   -- count unposted orders against limit
  override_role,         -- who can approve override
  is_on_hold,            -- manual hold flag
  tenant_id

-- Real-time credit exposure (view or computed)
-- = outstanding_invoices + (if include_open_orders: undelivered_order_value)
```

### Performance & Optimisation

Credit checks run on every sale/order entry — must be sub-200ms. The exposure calculation (sum of outstanding invoices) is the potentially expensive part.

**Pre-compute exposure:** Maintain a `customer_credit_exposure` table updated incrementally on every invoice post and every payment. The credit check reads from this table (one row lookup) rather than summing invoices at runtime.

**Optimistic UI, strict backend:** Show the credit status in the order entry UI (fetched when the customer is selected, not on every keystroke). Re-validate strictly at submission time. The UI check is for UX; the backend check is the control.

**pg_notify for real-time exposure updates:** When a large payment is received, pg_notify updates the customer's credit status in the UI immediately — so a sales rep doesn't get blocked on a customer who just paid.

### SQL

```sql
-- Credit exposure for a customer (real-time calculation)
SELECT
    ccp.credit_limit,
    COALESCE(outstanding.total, 0)                              AS outstanding_invoices,
    COALESCE(CASE WHEN ccp.include_open_orders
                  THEN open_orders.total ELSE 0 END, 0)         AS open_orders_value,
    COALESCE(outstanding.total, 0) +
    COALESCE(CASE WHEN ccp.include_open_orders
                  THEN open_orders.total ELSE 0 END, 0)         AS total_exposure,
    ccp.credit_limit -
    (COALESCE(outstanding.total, 0) +
     COALESCE(CASE WHEN ccp.include_open_orders
                   THEN open_orders.total ELSE 0 END, 0))       AS available_credit
FROM customer_credit_profiles ccp
LEFT JOIN LATERAL (
    SELECT SUM(amount_outstanding) AS total
    FROM invoices
    WHERE customer_id = ccp.customer_id
      AND status IN ('OPEN', 'PARTIAL')
      AND tenant_id = $2
) outstanding ON TRUE
LEFT JOIN LATERAL (
    SELECT SUM(total_amount) AS total
    FROM sales_orders
    WHERE customer_id = ccp.customer_id
      AND status IN ('DRAFT', 'CONFIRMED')
      AND tenant_id = $2
) open_orders ON TRUE
WHERE ccp.customer_id = $1
  AND ccp.tenant_id   = $2;
```

### Go

```go
type CreditCheckResult struct {
    Approved       bool
    Action         string // APPROVED / SOFT_WARNING / REQUIRES_APPROVAL / HARD_BLOCKED
    CreditLimit    decimal.Decimal
    CurrentExposure decimal.Decimal
    NewExposure    decimal.Decimal
    AvailableAfter decimal.Decimal
    Message        string
}

func (s *CreditService) CheckCredit(ctx context.Context, customerID string, orderAmount decimal.Decimal) CreditCheckResult {
    profile, exposure, err := s.repo.GetCreditStatus(ctx, customerID)
    if err != nil {
        return CreditCheckResult{Approved: false, Action: "ERROR"}
    }

    if profile.IsOnHold {
        return CreditCheckResult{Approved: false, Action: "HARD_BLOCKED",
            Message: "Account is on manual credit hold"}
    }

    newExposure := exposure.Add(orderAmount)
    available   := profile.CreditLimit.Sub(newExposure)

    if newExposure.LessThanOrEqual(profile.CreditLimit) {
        return CreditCheckResult{Approved: true, Action: "APPROVED",
            NewExposure: newExposure, AvailableAfter: available}
    }

    switch profile.CreditCheckType {
    case "HARD_BLOCK":
        return CreditCheckResult{Approved: false, Action: "HARD_BLOCKED",
            Message: fmt.Sprintf("Order would exceed credit limit by %s", available.Abs())}
    case "SOFT_WARN":
        return CreditCheckResult{Approved: true, Action: "SOFT_WARNING",
            Message: fmt.Sprintf("Credit limit exceeded by %s — proceed with caution", available.Abs())}
    case "APPROVAL_REQUIRED":
        return CreditCheckResult{Approved: false, Action: "REQUIRES_APPROVAL",
            Message: fmt.Sprintf("Credit limit exceeded. Requires %s approval", profile.OverrideRole)}
    }
    return CreditCheckResult{Approved: false, Action: "HARD_BLOCKED"}
}
```

---

## Summary: Phase Comparison

| Dimension | Phase 1 — Formulas | Phase 2 — Statistical | Phase 3 — Rule Engine |
|---|---|---|---|
| **Training data needed** | None | 30–500+ samples | None |
| **Explainability** | Perfect — pure math | High — Z-score, IQR are interpretable | Perfect — rules are human-readable |
| **Configurability** | Low — code changes for formula changes | Medium — threshold is configurable | High — rules managed by admins |
| **Failure mode** | Wrong formula = systematic error | Insufficient data = no flag | Wrong rule = missed or excessive flags |
| **Latency** | Sub-millisecond | Sub-millisecond (with pre-computed baselines) | Sub-100ms (with cached rules) |
| **Dependencies** | None | Historical data in DB | Rule store + evaluation engine |
| **When to use** | Always, for financial calculations | When "normal" must be learned from history | When business policy is complex and changeable |
| **AWO ERP priority** | Build first — non-negotiable | Build with Phase 1 — immediate value | Build incrementally — start with approval + credit |

---

*Document version 1.0 — ERP Deterministic Algorithms: Phase 1, 2 & 3*  
*For AWO ERP internal engineering reference*
