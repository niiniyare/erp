# Forecourt Management Module — Frappe/ERPNext
## Comprehensive Implementation Guide for Petrol Station Operators

**Version:** 2.0.0  
**Station Reference:** Shell Maanzoni Service Station (Anika Global Limited)  
**Stack:** ERPNext 15 / Frappe Framework  
**Currency:** KES (Kenya Shillings)  
**Fuel Grades in Scope:** Unleaded (PMS), V-Power (Premium PMS), AGO (Diesel), DPK (Kerosene)

---

## Table of Contents

1. [Business Context — What Can Go Wrong](#1-business-context)
2. [The Seven Golden Rules](#2-the-seven-golden-rules)
3. [The Three Types of Meter Readings — Core Concept](#3-the-three-types-of-meter-readings)
4. [Cross-Validation — How the Three Meters and the Dip Check Each Other](#4-cross-validation)
5. [Phase 1 — ERPNext Core Setup](#5-phase-1-erpnext-core-setup)
6. [Phase 2 — Custom Doctypes](#6-phase-2-custom-doctypes)
7. [Phase 3 — The Financial Flow](#7-phase-3-the-financial-flow)
8. [Phase 4 — Wetstock Reconciliation](#8-phase-4-wetstock-reconciliation)
9. [Phase 5 — Reporting](#9-phase-5-reporting)
10. [How Everything Relates](#10-how-everything-relates)
11. [Daily Operating Checklist](#11-daily-operating-checklist)
12. [Business Rules Reference](#12-business-rules-reference)
13. [Appendix — Variance Tolerance Quick Reference](#13-appendix)

---

## 1. Business Context

### 1.1 The Three Silent Destroyers

A petrol station appears straightforward: buy fuel, sell fuel. In practice three silent problems can destroy a station financially before management even notices:

**Silent Destroyer 1: Cash Leakage**
Your cashier collects KES 342,310 in a shift (as happened on 2026-04-24 at Shell Maanzoni — 1,731.95 litres of Unleaded at KES 197.6). But the meter says 1,756 litres were dispensed. Where is the cash for 24 litres? Without a system that reconciles cash collected against fuel dispensed per nozzle per attendant, you will never know, and it will happen again tomorrow.

**Silent Destroyer 2: Fuel Losses**
Across 8 active pumps and 3 tanks running 175+ transactions per shift, even a 0.3% wetstock loss is material. On a 50,000-litre/day station that is 150 litres — approximately KES 29,640 at Unleaded rates — lost every single day. Annualised: over KES 10.8 million. The loss can be theft, meter drift, short deliveries, or a leaking tank. The system's job is to tell you which.

**Silent Destroyer 3: Delivery Fraud**
Your supplier's docket says 8,000 litres arrived. Your tank dip says 7,796 litres are actually in the tank. You just paid for 204 litres you never received: 204 × KES 150 = KES 30,600 in a single delivery. Without before-and-after dip readings, you have no evidence to dispute the docket.

### 1.2 What the System Provides

By the end of this guide you will have a Frappe system that:

- Records all three meter reading types per nozzle (Electronic Volume, Electronic Cash, Manual Mechanical) and validates them against each other
- Tracks every litre entering the site (delivery with before/after dips) and leaving it (meter-confirmed sales)
- Reconciles every shilling a cashier is accountable for, separated from non-cash tenders
- Computes wetstock variance automatically: theoretical stock vs actual stock per tank
- Flags inconsistencies between the three meter types as data entry errors before they become financial losses
- Produces a daily shift report and GL journal entries that map directly to your chart of accounts

### 1.3 Real Data Context — Shell Maanzoni 2026-04-24

The data from the daily shift report for 2026-04-24 06:45 → 2026-04-25 06:45 illustrates the scale this system must handle:

| Metric | Value |
|---|---|
| Active Products | Unleaded (PMS), V-Power (Premium PMS) |
| Total Volume Sold | 1,756.45 litres |
| Total Revenue | KES 347,510.85 |
| Active Pumps | 8 (Pumps 1–8) |
| Tanks | 3 (Tank 1: Unleaded, Tank 2: AGO, Tank 3: V-Power) |
| Nozzle Configurations | Tank 1 → Nozzle 2 across Pumps 1–8; Tank 2 → Nozzles 1 and 3; Tank 3 → Nozzles 1 and 3 |
| Attendants Active | James Kitiapi, Shadrack Kimulu, Swedi Masanja, + unassigned (N/A) |
| Unleaded Transactions | 175 transactions, 1,731.95 L @ KES 197.6/L |
| V-Power Transactions | 5 transactions, 24.50 L @ KES 212.00/L |
| Diesel Transactions | 0 (AGO pumps idle this shift) |

The system must handle this volume of transactions while maintaining audit-quality records at every step.

---

## 2. The Seven Golden Rules

Every doctype, every field, every validation in this guide exists to enforce one of these rules.

### Rule 1 — The Shift is the Unit of Everything

A shift is not just a time window. It is a **bounded accountability contract** between the station and a named cashier. Every litre dispensed, every shilling collected, every delivery, every dip, every meter reading — all of it belongs to a shift and is linked to it.

> If it didn't happen in a shift, it didn't happen.

### Rule 2 — Three Meters, One Truth

Every modern dispensing pump has three ways to read how much fuel has been dispensed:
1. The **Electronic Volume meter** (cumulative litres on the digital display)
2. The **Electronic Cash meter** (cumulative KES on the digital display — litres × price)
3. The **Manual Mechanical counter** (physical number wheels, tamper-evident)

All three must be read at shift open and shift close. All three must be stored. The difference (close − open) must be computed for each. The three results must agree within tolerance. If they don't, there is an error — either in data entry or in the pump hardware — and it must be resolved before the shift can be approved.

### Rule 3 — Dip Readings Are the Tank's Bank Statement

A dip reading measures physical fuel volume in the tank independently of the pump meters. It is the external auditor that catches meter drift, unrecorded sales, theft, and leaks. At shift close for every tank:

```
Theoretical Closing Stock = Opening Dip + Deliveries − Σ(Electronic Volume sold, all linked nozzles)
Actual Closing Stock      = Closing Dip
Variance                  = Theoretical − Actual
```

If the variance exceeds tolerance, the shift cannot auto-approve.

### Rule 4 — Cash and Fuel Reconcile Separately

**Cash reconciliation** asks: Did the cashier collect the correct amount of money?
**Wetstock reconciliation** asks: Did the correct amount of fuel leave the tanks?

They are computed independently using different data sources. A discrepancy in one often explains a discrepancy in the other, but they are never merged into a single calculation.

### Rule 5 — No Opening Reading, No Sales

A cashier must not begin dispensing before all three meter readings (Electronic Volume, Electronic Cash, Manual) are captured for every active nozzle. The oldest forecourt fraud is to begin dispensing, pocket the cash, then record the opening reading after the fact. The system enforces this at the application layer.

### Rule 6 — Deliveries Require Before and After Dips

Never accept a fuel delivery on docket volume alone. The sequence is:
1. Dip the tank before the truck pumps (Delivery Before dip)
2. Wait 10–15 minutes after pumping for fuel to settle
3. Dip the tank again (Delivery After dip)
4. Raise the Purchase Receipt using the **dip-measured volume** if it differs from the docket by more than 0.5%

### Rule 7 — Dual Control on All Cash Movements

Any cash movement — float issuance, mid-shift pickup, safe drop — requires two people: the cashier and a supervisor with a different user account. Both are recorded in the system. No exceptions.

---

## 3. The Three Types of Meter Readings

This is the most important section of the entire document. Understanding the three meter types is the foundation of all wetstock integrity and fraud detection.

### 3.1 Overview

Every dispensing pump installed at a petrol station physically contains (or displays) three independent measurement systems. Each measures the same economic reality — fuel dispensed — from a different angle and using different mechanisms.

```
┌─────────────────────────────────────────────────────────┐
│                    DISPENSING PUMP                      │
│                                                         │
│  ┌─────────────────┐  ┌──────────────────┐             │
│  │  DIGITAL DISPLAY │  │  MECHANICAL DRUM  │             │
│  │                  │  │  (Number Wheels)  │             │
│  │  Vol: 148,730.45 │  │  ┌──────────────┐│             │
│  │  Cash: 29,387,  │  │  │ 0 1 4 8 7 3 0││             │
│  │        277.20    │  │  └──────────────┘│             │
│  └─────────────────┘  └──────────────────┘             │
│         │                      │                        │
│  Electronic Volume    Manual Mechanical                 │
│  Electronic Cash      (Totalizer)                      │
└─────────────────────────────────────────────────────────┘
```

### 3.2 Type 1 — Electronic Volume Meter (Elec Vol)

**What it is:** A cumulative, electronic digital counter displaying total litres dispensed through this nozzle since the pump was installed or last calibrated. This is the primary legal measurement for fuel sold.

**Physical location:** Digital display panel on the pump face. Always visible.

**Characteristics:**
- Counts forward only — it never resets, never goes back
- Displays to 2–3 decimal places (e.g. 148,730.450 litres)
- Sealed by Weights & Measures — tampering is a criminal offence
- Is the number used in all wetstock and sales calculations

**How to read it:** Write down exactly what the digital display shows, including all decimal places. Do not round.

**Example from Shell Maanzoni (2026-04-24, end of shift):**

| Pump | Nozzle | Tank | Electronic Volume Reading (L) |
|------|--------|------|-------------------------------|
| 1 | 2 | 1 (Unleaded) | 462,357.45 |
| 2 | 2 | 1 (Unleaded) | 1,468,477.52 |
| 7 | 2 | 1 (Unleaded) | 696,268.15 |
| 8 | 2 | 1 (Unleaded) | 1,199,050.25 |

These are the raw closing totalizer values. Volume sold = Closing − Opening for each nozzle.

**Volume sold formula:**
```
Elec Vol Sold (L) = Closing Electronic Volume − Opening Electronic Volume
```

### 3.3 Type 2 — Electronic Cash Meter (Elec Cash)

**What it is:** A cumulative, electronic digital counter displaying the total monetary value (in KES) of all fuel dispensed through this nozzle since installation or last calibration. It is mathematically linked to the Electronic Volume meter via the programmed price per litre.

**Physical location:** Same digital display panel as the Volume meter, usually on the line directly above or below the volume reading.

**Characteristics:**
- Represents: `Cumulative Litres Dispensed × Price Per Litre at Time of Each Transaction`
- When the EPRA pump price changes, new transactions are calculated at the new price — the historical cash total is not retroactively adjusted
- Can diverge from Volume × current price when a price change occurred mid-history (expected)
- Within a single shift at a fixed price: `Elec Cash Sold = Elec Vol Sold × Rate` should hold exactly

**How to read it:** Write down the full KES figure from the display. The number will be very large (millions) for an established pump. You are interested only in the delta: `Closing Elec Cash − Opening Elec Cash`.

**Cross-check rule (same-rate shift):**
```
Elec Cash Sold = (Closing Elec Cash) − (Opening Elec Cash)

This MUST equal:
Elec Vol Sold × Rate per litre

Example: 197.6 L × KES 197.60/L = ?
If Elec Cash Sold ≠ Elec Vol Sold × Rate → data entry error or rate mismatch
Tolerance: ≤ KES 1.00 (rounding)
```

**Why store the Electronic Cash reading if we can compute it?**
Because it provides an **independent cross-check** that does not rely on your data entry of volume being correct. If someone miskeys the Electronic Volume reading, the Electronic Cash reading will immediately expose the discrepancy. The two meters are physically independent inside the pump.

### 3.4 Type 3 — Manual Mechanical Counter (Manual Mech)

**What it is:** A physical, mechanical odometer-style counter built into the pump hardware. Consists of rotating numbered drums (number wheels) driven by the fuel flow measurement mechanism. It is completely independent of the electronic systems — it has no power supply requirement and cannot be zeroed electronically.

**Physical location:** A small window or panel on the pump body, sometimes requiring the attendant to open a panel to read. It displays a series of digit wheels.

**Characteristics:**
- Purely mechanical — immune to electronic tampering, power surges, and software manipulation
- Cannot be reset without a physical key held by the calibration authority (Weights & Measures)
- Less precise than the electronic meter (typically 1 decimal place or whole numbers)
- Is the last line of defence when electronic tampering is suspected
- Older pumps may have only the mechanical counter; newer pumps have all three

**How to read it:** Read each digit wheel in sequence from left to right. Write the complete number. If the window is hard to read, use a torch. Never estimate.

**Cross-check rule:**
```
Manual Mech Sold = (Closing Manual Mech) − (Opening Manual Mech)

This MUST be close to:
Elec Vol Sold (within 0.5% tolerance)

If Manual Mech Sold diverges from Elec Vol Sold by > 0.5%:
→ Possible electronic meter tampering or meter drift
→ Flag for calibration authority inspection
```

### 3.5 Summary Table — Three Meter Types

| Attribute | Electronic Volume | Electronic Cash | Manual Mechanical |
|---|---|---|---|
| Display type | Digital screen | Digital screen | Physical number wheels |
| Unit | Litres (L) | KES | Litres (L) |
| Precision | 0.001 L | KES 0.01 | 0.1 L or 1 L |
| Power required | Yes | Yes | No |
| Can be electronically zeroed | No (sealed) | No (sealed) | No (mechanical key only) |
| Primary use | Sales volume calculation | Cash cross-check | Tampering detection |
| Calculation | Close − Open = Vol Sold | Close − Open = Cash Sold | Close − Open = Vol Sold |
| Cross-validates with | Elec Cash, Manual Mech | Elec Vol × Rate | Elec Vol |
| Also validates against | Dip Reading | POS Invoice totals | Dip Reading |

---

## 4. Cross-Validation — How the Three Meters and the Dip Check Each Other

### 4.1 The Validation Pyramid

At shift close, four independent data sources must tell a consistent story about how much fuel was sold from a nozzle/tank. Any material inconsistency between them is evidence of an error, a malfunction, or fraud.

```
                   ┌─────────────────────┐
                   │   DIP READING (L)   │  ← Physical reality in tank
                   │  (Tank-level check) │
                   └──────────┬──────────┘
                              │ must agree with
              ┌───────────────┼───────────────┐
              │               │               │
   ┌──────────▼──┐  ┌─────────▼────┐  ┌──────▼──────────┐
   │ Elec Vol    │  │ Elec Cash    │  │ Manual Mech     │
   │ (L)         │  │ ÷ Rate = (L) │  │ (L)             │
   │ Nozzle-level│  │ Nozzle-level │  │ Nozzle-level    │
   └─────────────┘  └──────────────┘  └─────────────────┘
         │                │                    │
         └────────────────┴────────────────────┘
                   Must all agree within tolerance
```

### 4.2 Validation Rules at Shift Close

For each nozzle, compute the following at shift close. All checks must pass before the shift can be approved.

**Check A — Volume vs Cash consistency (nozzle level)**
```
Elec Vol Sold (L)    = Closing Elec Vol  − Opening Elec Vol
Elec Cash Sold (KES) = Closing Elec Cash − Opening Elec Cash
Expected Cash (KES)  = Elec Vol Sold × Shift Rate per Litre

Discrepancy = |Elec Cash Sold − Expected Cash|

If Discrepancy > KES 5.00 → FLAG: "Electronic Cash vs Volume mismatch on Pump X Nozzle Y"
```

**Check B — Volume vs Mechanical consistency (nozzle level)**
```
Mech Vol Sold (L) = Closing Manual Mech − Opening Manual Mech

Volume Divergence % = |Elec Vol Sold − Mech Vol Sold| / Elec Vol Sold × 100

If Divergence > 0.5% → FLAG: "Electronic vs Mechanical meter divergence on Pump X Nozzle Y — inspect pump"
If Divergence > 1.0% → CRITICAL: "Possible meter tampering — lock pump, notify calibration authority"
```

**Check C — Nozzle volumes aggregate to tank (tank level)**
```
For each Tank T:
  Tank Elec Vol Sold (L) = Σ (Elec Vol Sold for all nozzles whose Tank = T)

This is the definitive volume sold from Tank T this shift.
```

**Check D — Tank meter sales vs Dip reconciliation (tank level)**
```
Theoretical Closing Stock (L) = Opening Dip + Deliveries − Tank Elec Vol Sold
Actual Closing Stock (L)      = Closing Dip

Wetstock Variance (L) = Theoretical − Actual
Wetstock Variance %   = Variance / (Opening Dip + Deliveries) × 100
```

**Check E — Electronic Cash total vs POS invoice totals (shift level)**
```
For each nozzle:
  Elec Cash Sold = Σ of all POS Invoice line amounts for that nozzle this shift

If POS total < Elec Cash Sold → missing invoices or unrecorded sales
If POS total > Elec Cash Sold → duplicate invoices or data entry error
Tolerance: ≤ KES 10.00 per nozzle (rounding across many transactions)
```

### 4.3 Outcome Matrix — What Each Discrepancy Means

| Discrepancy Pattern | Most Likely Cause | Action |
|---|---|---|
| Elec Vol ≠ Elec Cash (large) | Rate not updated in system after EPRA price change | Update rate, recheck |
| Elec Vol ≠ Manual Mech (small, consistent) | Electronic meter drift | Schedule calibration |
| Elec Vol ≠ Manual Mech (sudden, large) | Electronic meter tampering | Lock pump, inspect |
| Elec Cash < POS Totals | Unrecorded sales / cashier pocketing cash | Investigate cashier |
| Elec Cash > POS Totals | Missing invoices, POS voids not approved | Reconstruct POS records |
| Wetstock Variance > 0.5% (single tank, consistent) | Tank leak or specific nozzle fault | Inspect tank and nozzle |
| Wetstock Gain > 0.3% | Delivery over-measured or meter under-reading | Verify delivery dip, inspect meter |
| All three meters agree, dip disagrees | Incorrect dip reading technique | Retake dip, use dipstick chart |

---

## 5. Phase 1 — ERPNext Core Setup

### 5.1 Company Setup

Navigate to **Setup → Company → New**:

| Field | Value |
|---|---|
| Company Name | Anika Global Limited T/A Shell Maanzoni Service Station |
| Abbreviation | SMSS |
| Default Currency | KES |
| Country | Kenya |
| Fiscal Year | January – December |

### 5.2 Chart of Accounts

Extend ERPNext's default chart with the following accounts. Navigate to **Accounts → Chart of Accounts**.

**Income Accounts (parent: Sales)**

| Account Name | Account Type | Purpose |
|---|---|---|
| Fuel Sales — PMS (Unleaded) | Income Account | Unleaded petrol revenue |
| Fuel Sales — PMS (V-Power) | Income Account | Premium petrol revenue |
| Fuel Sales — AGO | Income Account | Diesel revenue |
| Fuel Sales — DPK | Income Account | Kerosene revenue |
| Lubricant & Accessories Sales | Income Account | Oils, additives, shop |

**COGS Accounts (parent: Cost of Goods Sold)**

| Account Name | Account Type | Purpose |
|---|---|---|
| COGS — Fuel PMS Unleaded | Expense Account | Cost of Unleaded sold |
| COGS — Fuel PMS V-Power | Expense Account | Cost of V-Power sold |
| COGS — Fuel AGO | Expense Account | Cost of Diesel sold |
| COGS — Fuel DPK | Expense Account | Cost of Kerosene sold |
| Wetstock Variance — PMS | Expense Account | Fuel losses, petrol |
| Wetstock Variance — AGO | Expense Account | Fuel losses, diesel |
| Cash Short / Over | Expense Account | Cashier variances |
| Meter Discrepancy — Adjustment | Expense Account | Resolved meter discrepancies |

**Asset Accounts (parent: Current Assets)**

| Account Name | Account Type | Purpose |
|---|---|---|
| Fuel Inventory — PMS Unleaded | Stock Account | Physical Unleaded in Tank 1 |
| Fuel Inventory — PMS V-Power | Stock Account | Physical V-Power in Tank 3 |
| Fuel Inventory — AGO | Stock Account | Physical AGO in Tank 2 |
| MPesa Clearing | Bank Account | MPesa float, settled daily |
| Card Payment Clearing | Bank Account | Card terminal batch |
| Fleet Card Clearing | Bank Account | Shell/fleet card settlements |
| Safe — Main | Cash Account | Site safe balance |
| Till — Cashier (Active) | Cash Account | Active cashier till |

### 5.3 Fuel Items

Navigate to **Stock → Items → New Item**. Create one item per product-grade combination.

**Shell V-Power (Premium Unleaded) — Example:**

| Field | Value |
|---|---|
| Item Name | Shell V-Power (PMS Premium) |
| Item Code | FUEL-PMS-VP |
| Item Group | Fuel |
| Stock UOM | Litre |
| Is Stock Item | ✓ Yes |
| Maintain Stock | ✓ Yes |
| Default Warehouse | Tank 3 — V-Power |
| Standard Selling Rate | 212.00 (current EPRA price) |
| Standard Buying Rate | 175.00 (depot purchase cost) |
| Valuation Method | Moving Average (WAC) |

> **Why WAC (Moving Average)?** Each delivery arrives at a slightly different cost per litre depending on global crude prices, transport, and depot margins. WAC blends the cost of all fuel in the tank automatically. When 1 litre is sold, profit is calculated against the blended cost — not the first-in cost (FIFO) or an arbitrary standard cost. This is the only correct valuation method for bulk liquid fuel.

Repeat for: `FUEL-PMS-UNL` (Unleaded), `FUEL-AGO` (Diesel), `FUEL-DPK` (Kerosene).

**Important: Keep V-Power and Unleaded as separate items even though both are PMS.** They sell at different prices, have different EPRA price schedules, and may come from different depot batches. Merging them would make WAC meaningless and revenue reporting inaccurate.

### 5.4 Tanks as Warehouses

In Frappe, each physical storage tank is a **Warehouse**. Navigate to **Stock → Warehouses → New Warehouse**.

Based on Shell Maanzoni's actual configuration:

| Warehouse Name | Fuel Product | Pumps Fed | Nozzle on Pump |
|---|---|---|---|
| Tank 1 — Unleaded | FUEL-PMS-UNL | Pumps 1–8 | Nozzle 2 |
| Tank 2 — AGO | FUEL-AGO | Pumps 1–8 | Nozzles 1 and 3 |
| Tank 3 — V-Power | FUEL-PMS-VP | Pumps 1–8 | Nozzles 1 and 3 |

> Create one warehouse per physical tank. Never combine two tanks into one warehouse, even if they hold the same product. You need separate dip readings per physical tank, and two tanks of the same product may have different WAC costs if they were filled from different deliveries.

### 5.5 Employees

Navigate to **HR → Employee → New Employee** for:
- Each named cashier (James Kitiapi, Swedi Masanja, Shadrack Kimulu, etc.)
- The station manager / supervisor
- Each pump attendant

Minimum fields: Employee Name, Date of Joining, Department (Forecourt Operations), Designation (Cashier / Attendant / Supervisor / Manager).

> **Important about "N/A" attendant records:** The shift data shows some transactions with attendant = "N/A" (unassigned). This is a process failure — every transaction must be linked to a named employee. Create a system validation that prevents POS submission without a valid attendant selection. "N/A" or blank is not acceptable in a system designed for accountability.

### 5.6 Suppliers and Customers

**Suppliers:** Navigate to **Buying → Suppliers → New Supplier** for each fuel depot (VIVO Energy / Shell Kenya depot, any secondary depot). Deliveries are raised against these.

**Credit customers:** Navigate to **Selling → Customer → New Customer** for fleet companies and credit account holders. Set Payment Terms to the agreed credit period. These generate Accounts Receivable entries rather than cash.

---

## 6. Phase 2 — Custom Doctypes

These are the doctypes that ERPNext does not provide out of the box. Create each in **Frappe → Customise → DocType**.

---

### 6.1 Forecourt Pump

**Purpose:** Master record for each physical dispensing unit. This is the anchor for all nozzle-level meter readings and establishes the nozzle → tank → product mapping that the entire wetstock calculation depends on.

**DocType Name:** `Forecourt Pump`

| Field Label | Field Name | Type | Req | Notes |
|---|---|---|---|---|
| Pump Number | pump_number | Int | ✓ | Physical number on the island (1–8 for Shell Maanzoni) |
| Pump Name | pump_name | Data | ✓ | e.g. "Island 2 — Pump B" |
| Make | make | Data | | Tokheim, Wayne, Gilbarco, Censtar, etc. |
| Model | model | Data | | |
| Serial Number | serial_number | Data | | For maintenance records |
| Is Active | is_active | Check | ✓ | Default: Yes. Inactive pumps excluded from shift validation |
| Installation Date | installation_date | Date | | |
| Weights & Measures Seal No | seal_number | Data | | Legal calibration seal — record after each recalibration |
| Last Calibration Date | last_calibration_date | Date | | |
| Next Calibration Due | next_calibration_due | Date | | System should alert 30 days before |
| Calibration Certificate | calibration_cert | Attach | | Scanned copy |
| Notes | notes | Text | | |

**Child Table: Pump Nozzles**

Each pump has one or more nozzles. A nozzle is the physical hose and gun that dispenses fuel. Each nozzle connects to exactly one tank and dispenses exactly one product. This mapping is the key to aggregating nozzle-level meter readings to tank-level wetstock figures.

| Field Label | Field Name | Type | Req | Notes |
|---|---|---|---|---|
| Nozzle Number | nozzle_number | Int | ✓ | 1-based. Per Shell Maanzoni: 1=V-Power/AGO, 2=Unleaded, 3=AGO/V-Power |
| Tank | tank | Link → Warehouse | ✓ | Must be a tank-type warehouse |
| Fuel Product | fuel_product | Link → Item | ✓ | Must be a fuel item |
| Is Active | is_active | Check | ✓ | Inactive nozzles excluded from shift readings |
| Colour Code | colour_code | Data | | Green=Unleaded, Black=AGO, Red=V-Power (Shell standard) |

> **Design rationale for the child table:** Without this nozzle-level mapping, the system cannot know which tank's stock was depleted by a given meter reading. From the Shell Maanzoni data: all pumps have Nozzle 2 → Tank 1 (Unleaded). Some pumps have Nozzle 1 or 3 → Tank 2 (AGO) or Tank 3 (V-Power). This mapping must be explicit in the database.

**Validations:**
- No two nozzles on the same pump may link to the same tank
- If a pump is set to inactive, all its nozzles are excluded from shift readings automatically
- Calibration due date alert: when `next_calibration_due` is within 30 days, show a dashboard warning

---

### 6.2 Forecourt Shift

**Purpose:** The master record for a shift. Every other operational record in the system — meter readings, dip readings, cash events, deliveries — hangs off a Shift record. The Shift is the unit of accountability.

**DocType Name:** `Forecourt Shift`

| Field Label | Field Name | Type | Req | Notes |
|---|---|---|---|---|
| Shift Number | shift_number | Data | ✓ | Auto-generated: SHIFT-2026-0001 |
| Station | station | Link → Branch | ✓ | Shell Maanzoni Service Station |
| Shift Date | shift_date | Date | ✓ | Date the shift started |
| Shift Label | shift_label | Select | | Day / Evening / Night (for multi-shift stations) |
| Opened At | opened_at | Datetime | ✓ | Actual start time |
| Closed At | closed_at | Datetime | | Set on close |
| Primary Cashier | cashier | Link → Employee | ✓ | Financially accountable employee |
| Supervisor | supervisor | Link → Employee | ✓ | Must be different from Cashier |
| Status | status | Select | ✓ | Open / Readings Captured / Closing / Closed / Disputed |
| Opening Float (KES) | float_amount | Currency | ✓ | Cash issued to cashier at shift start |
| Shift Notes | notes | Text | | |
| Meter Validation Passed | meter_validation_ok | Check | | System-set after Check A and Check B pass |
| Wetstock Balanced | wetstock_ok | Check | | System-set after Check D passes |
| GL Journal | gl_journal | Link → Journal Entry | | Set after GL posting |

**Child Table: Shift Pump Assignments**

| Field | Type | Notes |
|---|---|---|
| Employee | Link → Employee | Pump attendant |
| Pump | Link → Forecourt Pump | Which pump |
| Assigned At | Datetime | |
| Relieved At | Datetime | For mid-shift handover |
| Notes | Text | |

**Status Transition Rules (enforce in code):**
```
Open → Readings Captured  (all 3 meter types entered for all active nozzles)
Readings Captured → Closing  (all dip readings entered, cash declared)
Closing → Closed  (reconciliation computed and approved, GL posted)
Any status → Disputed  (manager intervention required)
Disputed → Closing  (after investigation, manager clears)
```

**Validations:**
- `Cashier` and `Supervisor` must be different employees
- Only one shift may have status `Open` or `Closing` at any time per station
- `closed_at` must be after `opened_at`
- Status transitions are one-directional (except Disputed → Closing for re-investigation)

---

### 6.3 Meter Reading

**Purpose:** A single snapshot of one reading type on one nozzle at a specific moment in time. This is the most granular operational record in the system and the source of all volume-sold calculations.

**DocType Name:** `Meter Reading`

**Critical design principle: Meter Reading records are immutable once submitted.** If a reading was entered incorrectly, create a new `Amendment` reading linked to the original. The original reading must remain in the database as evidence. Never delete or edit a submitted meter reading.

| Field Label | Field Name | Type | Req | Notes |
|---|---|---|---|---|
| Reading Reference | name | Data | Auto | MR-2026-0001 |
| Shift | shift | Link → Forecourt Shift | ✓ | Parent shift |
| Pump | pump | Link → Forecourt Pump | ✓ | Physical pump |
| Nozzle Number | nozzle_number | Int | ✓ | Which nozzle on this pump |
| **Meter Type** | **meter_type** | **Select** | **✓** | **Electronic Volume / Electronic Cash / Manual Mechanical** |
| Reading Position | reading_position | Select | ✓ | Shift Open / Shift Close / Spot Check / Amendment |
| Observed At | observed_at | Datetime | ✓ | When the reading was physically taken |
| Totalizer Value | totalizer_value | Float | ✓ | The raw number from the display or wheels |
| Unit | unit | Select | Auto | Set from meter_type: Litres (Elec Vol / Manual Mech) or KES (Elec Cash) |
| Read By | read_by | Link → Employee | ✓ | Who physically read the meter |
| Witnessed By | witnessed_by | Link → Employee | | Supervisor or second attendant |
| Verified In System By | recorded_by | Link → User | ✓ | Auto-filled from session |
| Notes | notes | Text | | Unusual observations (e.g. display flickering) |
| Superseded By | superseded_by | Link → Meter Reading | | Set when this reading is corrected by an amendment |
| Amendment Reason | amendment_reason | Text | | Required if reading_position = Amendment |

**The volume-sold calculation (computed at reconciliation time, never stored in this doctype):**
```python
# For a given nozzle and shift:
close_elec_vol  = MeterReading.get(shift, nozzle, "Shift Close",  "Electronic Volume")
open_elec_vol   = MeterReading.get(shift, nozzle, "Shift Open",   "Electronic Volume")
close_elec_cash = MeterReading.get(shift, nozzle, "Shift Close",  "Electronic Cash")
open_elec_cash  = MeterReading.get(shift, nozzle, "Shift Open",   "Electronic Cash")
close_mech      = MeterReading.get(shift, nozzle, "Shift Close",  "Manual Mechanical")
open_mech       = MeterReading.get(shift, nozzle, "Shift Open",   "Manual Mechanical")

elec_vol_sold  = close_elec_vol  - open_elec_vol    # Primary volume figure (Litres)
elec_cash_sold = close_elec_cash - open_elec_cash   # Primary cash figure (KES)
mech_vol_sold  = close_mech      - open_mech        # Secondary volume figure (Litres)

# Cross-checks (see Section 4):
expected_cash = elec_vol_sold * shift_rate
cash_vs_vol_discrepancy = abs(elec_cash_sold - expected_cash)
mech_vs_elec_divergence_pct = abs(elec_vol_sold - mech_vol_sold) / elec_vol_sold * 100
```

**Validations:**
- For `Electronic Volume` and `Manual Mechanical`: totalizer_value must be > 0
- For `Shift Close` readings: totalizer_value must be ≥ the corresponding Shift Open totalizer for the same nozzle and meter type
- All three meter types must have both a Shift Open and Shift Close reading before the shift can move to `Closing` status
- Once a reading reaches `Submitted` docstatus, its `totalizer_value` is read-only (enforced in code)

---

### 6.4 Meter Validation Result

**Purpose:** Stores the output of the four cross-validation checks (A through E from Section 4) per nozzle per shift. This is a computed record, not a data-entry form.

**DocType Name:** `Meter Validation Result`

| Field Label | Field Name | Type | Notes |
|---|---|---|---|
| Reference | name | Data | MVR-2026-0001 |
| Shift | shift | Link → Forecourt Shift | |
| Pump | pump | Link → Forecourt Pump | |
| Nozzle Number | nozzle_number | Int | |
| Shift Rate (KES/L) | shift_rate | Currency | EPRA price in effect this shift |
| Elec Vol Sold (L) | elec_vol_sold | Float | Primary volume figure |
| Elec Cash Sold (KES) | elec_cash_sold | Currency | From Electronic Cash delta |
| Mech Vol Sold (L) | mech_vol_sold | Float | From Mechanical delta |
| Expected Cash (KES) | expected_cash | Currency | Elec Vol × Shift Rate |
| Check A — Cash vs Vol | check_a_discrepancy | Currency | |Elec Cash − Expected Cash| |
| Check A — Status | check_a_status | Select | Pass / Warning / Fail |
| Check B — Mech vs Elec | check_b_divergence_pct | Float | % divergence |
| Check B — Status | check_b_status | Select | Pass / Warning / Fail / Critical |
| Overall Status | overall_status | Select | Pass / Warning / Fail / Critical |
| Computed At | computed_at | Datetime | |
| Notes | notes | Text | |

This record is auto-generated when the supervisor clicks "Run Meter Validation" on the Shift. It cannot be manually edited. To resolve a failure, correct the source Meter Reading (via Amendment) and re-run validation.

---

### 6.5 Dip Reading

**Purpose:** A physical measurement of fuel volume inside a tank at a specific moment. The independent auditor that validates the meter-based calculations.

**DocType Name:** `Dip Reading`

| Field Label | Field Name | Type | Req | Notes |
|---|---|---|---|---|
| Reading Reference | name | Data | Auto | DIP-2026-0001 |
| Shift | shift | Link → Forecourt Shift | | Can be null for routine/spot readings |
| Tank | tank | Link → Warehouse | ✓ | Must be a tank-type warehouse |
| Reading Type | reading_type | Select | ✓ | Shift Open / Shift Close / Delivery Before / Delivery After / Routine / Spot |
| Observed At | observed_at | Datetime | ✓ | |
| Dipstick Level (mm) | level_mm | Float | | Raw dipstick measurement |
| Volume Observed (L) | volume_observed_l | Float | ✓ | Converted from mm using tank calibration chart |
| Calibration Chart Used | calibration_chart | Data | | Reference to the tank strapping/calibration chart version |
| Temperature (°C) | temperature_c | Float | | If measured; for density correction |
| Water Bottom Level (mm) | water_level_mm | Float | | Alert threshold: 20 mm |
| Water Volume (L) | water_volume_l | Float | | Computed from water_level_mm via chart |
| Source | source | Select | ✓ | Manual Dipstick / ATG (Automatic Tank Gauge) |
| Read By | read_by | Link → Employee | ✓ | |
| Witnessed By | witnessed_by | Link → Employee | | Supervisor |
| Notes | notes | Text | | |

**Validations:**
- `volume_observed_l` must be > 0 and ≤ tank maximum capacity
- `water_level_mm > 20` → system warning: "High water level detected in [Tank Name] — possible water contamination. Alert manager immediately."
- For `Delivery After` readings: `observed_at` must be at least 10 minutes after the linked `Fuel Delivery Dip` delivery end time (settling period)
- For `Shift Close` readings: a corresponding `Shift Open` dip must exist for the same tank in the same shift

---

### 6.6 Cashier Session

**Purpose:** A bounded accountability period for one cashier on one till within a shift. One session = one cashier = one set of accountability numbers. If two cashiers share a till without a formal session handover, accountability is lost.

**DocType Name:** `Cashier Session`

| Field Label | Field Name | Type | Req | Notes |
|---|---|---|---|---|
| Session Reference | name | Data | Auto | CS-2026-0001 |
| Shift | shift | Link → Forecourt Shift | ✓ | |
| Cashier | cashier | Link → Employee | ✓ | |
| Till ID | till_id | Data | ✓ | e.g. "TILL-01" |
| Is Primary Session | is_primary | Check | ✓ | First cashier of the shift — only one per shift |
| Opening Float (KES) | float_amount | Currency | ✓ | Cash given to this cashier for this session |
| Session Opened At | session_opened_at | Datetime | ✓ | |
| Session Closed At | session_closed_at | Datetime | | |
| Actual Cash Count at Close | actual_cash_close | Currency | | Physically counted by cashier |
| Counted By | counted_by | Link → Employee | | The cashier |
| Verified By | verified_by | Link → Employee | | Supervisor (different from cashier) |
| Notes | notes | Text | | |

**Float vs Shift Float note:** The `Forecourt Shift` also has a `float_amount` field. For a single-cashier shift, these should be equal. For a mid-shift handover (Cashier A → Cashier B), Cashier A's session closes and Cashier B opens a new session. Cashier B's float is the cash remaining after Cashier A's session close — it is not a new injection from the safe unless management explicitly adds one.

---

### 6.7 Cash Event

**Purpose:** Records every movement of cash that affects the till balance during a shift. The till's running balance at any point = Opening Float + Cash Sales − Pickups − Payouts.

**DocType Name:** `Cash Event`

| Field Label | Field Name | Type | Req | Notes |
|---|---|---|---|---|
| Event Reference | name | Data | Auto | CE-2026-0001 |
| Shift | shift | Link → Forecourt Shift | ✓ | |
| Cashier Session | cashier_session | Link → Cashier Session | ✓ | The active session when this event occurred |
| Event Type | event_type | Select | ✓ | Float Issued / Cash Pickup / Payout / Safe Drop |
| Amount (KES) | amount | Currency | ✓ | |
| Authorised By | authorised_by | Link → Employee | ✓ | Supervisor — must be different from session cashier |
| Occurred At | occurred_at | Datetime | ✓ | |
| Envelope or Reference | reference | Data | | Safe envelope number, voucher number |
| Notes | notes | Text | | |

**Validations:**
- `authorised_by` must not be the same employee as the `cashier` on the linked `Cashier Session` (session-scoped check — a supervisor who is also a cashier on a different session cannot authorise their own session's events)
- `amount` must be positive and non-zero
- Only one `Float Issued` event is permitted per Cashier Session
- For `Cash Pickup` and `Safe Drop`: a physical envelope number or reference is required

---

### 6.8 Fuel Delivery Dip

**Purpose:** Links the procurement document (ERPNext Purchase Receipt) to the before/after dip readings that verify the delivery volume. This is the bridge between Procurement and Forecourt Operations.

**DocType Name:** `Fuel Delivery Dip`

| Field Label | Field Name | Type | Req | Notes |
|---|---|---|---|---|
| Reference | name | Data | Auto | FDD-2026-0001 |
| Shift | shift | Link → Forecourt Shift | ✓ | Shift during which delivery arrived |
| Tank | tank | Link → Warehouse | ✓ | Receiving tank |
| Truck Registration | truck_reg | Data | ✓ | |
| Driver Name | driver_name | Data | | |
| Docket Number | docket_number | Data | ✓ | Supplier's delivery note |
| Fuel Product | fuel_product | Link → Item | ✓ | |
| Docket Volume (L) | docket_volume_l | Float | ✓ | What the supplier claims |
| Delivery Start Time | delivery_start | Datetime | ✓ | When pumping started |
| Delivery End Time | delivery_end | Datetime | ✓ | When pumping stopped |
| Dip Reading Before | dip_before | Link → Dip Reading | ✓ | The Delivery Before dip record |
| Dip Reading After | dip_after | Link → Dip Reading | ✓ | The Delivery After dip record |
| Dip Before Volume (L) | dip_before_l | Float | Computed | Pulled from linked Dip Reading |
| Dip After Volume (L) | dip_after_l | Float | Computed | Pulled from linked Dip Reading |
| Dip-Measured Volume (L) | dip_measured_l | Float | Computed | After − Before |
| Variance vs Docket (L) | delivery_variance_l | Float | Computed | Dip-Measured − Docket |
| Variance % | delivery_variance_pct | Float | Computed | |
| Purchase Receipt | purchase_receipt | Link → Purchase Receipt | | Created after dip confirms volume |
| Status | status | Select | ✓ | Pending / Accepted / Disputed |
| Dispute Notes | dispute_notes | Text | | Required if Status = Disputed |

> **Implementation note:** `dip_before_l` and `dip_after_l` are computed by pulling `volume_observed_l` from the linked Dip Reading records. They are stored here as computed copies for report convenience, but the linked Dip Reading is the source of truth. If the Dip Reading is corrected, this record must be recomputed.

---

### 6.9 Shift Reconciliation

**Purpose:** The computed summary verdict of a closed shift. This is not a data-entry form. It is a calculated output that reads from all other doctypes and produces the financial and wetstock verdict. It must never be manually edited. Fix the source data, then recompute.

**DocType Name:** `Shift Reconciliation`

| Field Label | Field Name | Type | Notes |
|---|---|---|---|
| Reconciliation Reference | name | Data | SR-2026-0001 |
| Shift | shift | Link → Forecourt Shift | 1-to-1 |
| Computed At | computed_at | Datetime | |
| **Meter Validation** | | | |
| All Meters Validated | all_meters_ok | Check | True only if all MVR records = Pass |
| Meter Warnings Present | meter_warnings | Check | Warnings exist but were reviewed |
| **Volume Summary** | | | |
| Total Volume Sold — PMS Unleaded (L) | vol_pms_unl | Float | |
| Total Volume Sold — V-Power (L) | vol_pms_vp | Float | |
| Total Volume Sold — AGO (L) | vol_ago | Float | |
| Total Volume Sold — DPK (L) | vol_dpk | Float | |
| Total Volume Sold — All Products (L) | total_volume_l | Float | |
| **Revenue Summary** | | | |
| Gross Revenue — PMS Unleaded | rev_pms_unl | Currency | |
| Gross Revenue — V-Power | rev_pms_vp | Currency | |
| Gross Revenue — AGO | rev_ago | Currency | |
| Total Gross Revenue | total_gross_revenue | Currency | |
| **Cash Reconciliation** | | | |
| Expected Cash at Close (KES) | expected_cash | Currency | Formula: see §8.2 |
| Actual Cash Declared (KES) | actual_cash | Currency | From Cashier Session |
| Cash Variance (KES) | cash_variance | Currency | Actual − Expected |
| Cash Variance Status | cash_variance_status | Select | Balanced / Short / Over |
| **Non-Cash Tenders** | | | |
| MPesa Total (KES) | mpesa_total | Currency | |
| Card Total (KES) | card_total | Currency | |
| Fleet / Credit Total (KES) | fleet_credit_total | Currency | |
| **Wetstock Summary** | | | |
| Total Wetstock Variance (L) | wetstock_variance_l | Float | Summed across all tanks |
| Wetstock Variance (KES) | wetstock_variance_kes | Currency | Variance L × WAC |
| **Outcome** | | | |
| Cash Balanced | is_cash_balanced | Check | System-set |
| Wetstock Balanced | is_wetstock_balanced | Check | System-set |
| Meter Checks Passed | is_meter_ok | Check | System-set |
| Requires Manager Approval | requires_approval | Check | System-set |
| Approved By | approved_by | Link → User | |
| Approved At | approved_at | Datetime | |
| GL Journal | gl_journal | Link → Journal Entry | Set after GL posting |

**Child Table 1: Product Summaries**

| Field | Type | Notes |
|---|---|---|
| Fuel Product | Link → Item | One row per active product this shift |
| Volume Sold (L) | Float | Sum of Elec Vol Sold across all nozzles for this product |
| Shift Rate (KES/L) | Currency | EPRA price in effect |
| Gross Revenue (KES) | Currency | Volume × Rate |
| WAC Unit Cost (KES/L) | Currency | From ERPNext stock valuation |
| COGS (KES) | Currency | Volume × WAC |
| Gross Margin (KES) | Currency | Revenue − COGS |
| Gross Margin % | Float | |

**Child Table 2: Tank Wetstock Summaries**

| Field | Type | Notes |
|---|---|---|
| Tank | Link → Warehouse | One row per active tank this shift |
| Fuel Product | Link → Item | |
| Opening Stock (L) | Float | From Shift Open dip reading |
| Deliveries (L) | Float | From Fuel Delivery Dip records this shift |
| Nozzle Sales — Elec Vol (L) | Float | Σ Elec Vol Sold for all nozzles on this tank |
| Nozzle Sales — Mech Vol (L) | Float | Σ Mech Vol Sold for all nozzles on this tank (cross-check) |
| Theoretical Closing (L) | Float | Opening + Deliveries − Elec Vol Sales |
| Actual Closing (L) | Float | From Shift Close dip reading |
| Variance (L) | Float | Theoretical − Actual |
| Variance % | Float | |
| Classification | Select | Normal / Elevated / Critical / Gain |
| WAC (KES/L) | Currency | For KES conversion |
| Variance (KES) | Currency | Variance L × WAC |

**Child Table 3: Nozzle Meter Validation Summary**

| Field | Type | Notes |
|---|---|---|
| Pump | Link → Forecourt Pump | |
| Nozzle | Int | |
| Elec Vol Sold (L) | Float | |
| Elec Cash Sold (KES) | Currency | |
| Mech Vol Sold (L) | Float | |
| Check A Status | Select | Pass / Warning / Fail |
| Check B Status | Select | Pass / Warning / Fail / Critical |
| Linked MVR | Link → Meter Validation Result | |

---

## 7. Phase 3 — The Financial Flow

### 7.1 Business Event to ERPNext Transaction Mapping

| Business Event | ERPNext Transaction | Initiated By |
|---|---|---|
| Fuel delivery arrives | Purchase Receipt → Stock Entry (WAC update) | Supervisor / Manager |
| Shift opens | Forecourt Shift created, Cashier Session created | Supervisor |
| Opening float issued | Cash Event: Float Issued + JE (DR Till / CR Safe) | Supervisor |
| Customer fills up — cash | POS Invoice (payment: Cash) | Cashier |
| Customer fills up — MPesa | POS Invoice (payment: MPesa Clearing) | Cashier |
| Customer fills up — card | POS Invoice (payment: Card Clearing) | Cashier |
| Fleet customer fills up | Sales Invoice → AR balance | Cashier |
| Cash pickup mid-shift | Cash Event: Cash Pickup + JE | Supervisor |
| Shift closes — balanced | Journal Entry (shift GL) | System / Manager |
| Wetstock variance posted | Stock Reconciliation or Adjustment JE | Manager |
| Cash short | Journal Entry (DR Cash Short/Over) | System |
| Cash safe drop | Cash Event: Safe Drop + JE (DR Safe / CR Till) | Supervisor |

### 7.2 Recording a Fuel Delivery (Full Workflow)

**Step 1 — Before the truck arrives:**
- Create a `Dip Reading` for the receiving tank: Reading Type = `Delivery Before`
- Record the observed volume in litres (or dipstick mm converted via calibration chart)
- Note who took the reading and the time

**Step 2 — Truck arrives, connect hose, pump:**
- Create a `Fuel Delivery Dip` record
- Fill: Truck Registration, Driver Name, Docket Number, Fuel Product, Docket Volume, Delivery Start Time
- Link to the Delivery Before dip reading

**Step 3 — Pumping complete — wait 10–15 minutes:**
- The fuel must settle before an accurate after-dip can be taken
- Set Delivery End Time on the Fuel Delivery Dip record

**Step 4 — After-dip:**
- Create a `Dip Reading` for the same tank: Reading Type = `Delivery After`
- Link this dip to the Fuel Delivery Dip record
- System computes: `Dip-Measured Volume = After − Before`
- System computes: `Variance = Dip-Measured − Docket`

**Step 5 — Purchase Receipt:**
- If variance ≤ 0.5%: raise Purchase Receipt at docket volume; mark Fuel Delivery Dip as `Accepted`
- If variance > 0.5%: mark Fuel Delivery Dip as `Disputed`; do not post Purchase Receipt until resolved with supplier
- On Purchase Receipt submit: `DR Fuel Inventory − [Product] / CR Accounts Payable`

**Step 6 — If disputed:**
- Call the supplier depot with your dip records as evidence
- Request a credit note or corrected delivery note
- Post Purchase Receipt only at the dip-confirmed volume
- Document all communication in the `dispute_notes` field

### 7.3 Recording Sales During a Shift

**POS Setup Prerequisites:**
- Create a POS Profile for the station
- Payment methods: Cash, MPesa Clearing, Card Clearing, Fleet/Credit AR
- Link the POS to the station warehouse (product-specific warehouses for correct WAC depletion)
- Set `Allow Rate Edit: No` for fuel items — the EPRA price is fixed
- Set UOM to Litres

**Per transaction at the pump:**
1. Cashier selects the fuel product (Unleaded, V-Power, etc.)
2. Enters volume in litres (as shown on the pump display)
3. Price auto-calculates at the configured EPRA rate
4. Cashier selects payment method
5. Cashier selects the attendant's name (mandatory — see §5.5 note on N/A)
6. Submits invoice

> The transaction data from Shell Maanzoni shows non-round-litre sales (e.g. 20.24 L, 46.46 L, 55.47 L) mixed with round-KES amounts (KES 4,000, KES 9,180.50). This is normal: some customers ask for "fill it up" (non-round litres) and others ask for "KES 2,000 worth" (non-round litres). ERPNext POS handles both — ensure rounding is set to 2 decimal places on litres and 2 decimal places on KES.

**For fleet / credit customers:**
- Use Sales Invoice (not POS)
- Customer = the fleet company's customer record
- These do not produce cash in the till
- Include the transaction in the Shift Reconciliation's `fleet_credit_total` field

### 7.4 Mid-Shift Cash Pickup

When the till exceeds the agreed threshold (e.g. KES 30,000):

1. Supervisor counts the excess cash with the cashier present
2. Create `Cash Event`: Type = Cash Pickup, Amount = excess, Authorised By = supervisor
3. Assign a sealed envelope number, physically seal the cash
4. Create Journal Entry:
   ```
   DR  Safe — Main         [amount]
       CR  Till — Cashier (Active)    [amount]
   ```
5. Both cashier and supervisor sign the physical envelope

---

## 8. Phase 4 — Wetstock Reconciliation

### 8.1 The Complete Reconciliation Formula

For each tank at shift close:

```
Step 1 — Establish Opening Stock
Opening Stock (L) = volume_observed_l from Shift Open Dip Reading for this tank

Step 2 — Add Deliveries
Deliveries (L) = Σ dip_measured_l from all Fuel Delivery Dip records
                  where tank = this tank AND shift = this shift
                  AND status = Accepted

Step 3 — Compute Sales (from Meter Readings — Primary Method)
Elec Vol Sales (L) = Σ (Close Elec Vol − Open Elec Vol)
                      for all active nozzles where tank = this tank

Step 4 — Compute Theoretical Closing Stock
Theoretical Closing (L) = Opening Stock + Deliveries − Elec Vol Sales

Step 5 — Measure Actual Closing Stock
Actual Closing (L) = volume_observed_l from Shift Close Dip Reading for this tank

Step 6 — Compute Variance
Variance (L)   = Theoretical Closing − Actual Closing
Variance %     = Variance / (Opening Stock + Deliveries) × 100

Step 7 — Cross-Check with Mechanical Meters
Mech Vol Sales (L) = Σ (Close Manual Mech − Open Manual Mech)
                      for all active nozzles where tank = this tank

Mech vs Dip Check: If |Mech Vol Sales − Elec Vol Sales| / Elec Vol Sales > 0.5%
→ Flag: mechanical and electronic volumes disagree — inspect before approving
```

### 8.2 Cash Reconciliation Formula

```
Expected Cash at Session Close
  = Opening Float (from Cashier Session.float_amount)
  + Σ Cash Sales (POS Invoices where payment_method = Cash, this session)
  − Σ Cash Pickups (Cash Events where event_type = Cash Pickup, this session)
  − Σ Authorised Payouts (Cash Events where event_type = Payout, this session)

Cash Variance = Actual Cash Declared − Expected Cash

Positive variance = Cashier Over (more cash than expected — also suspicious)
Negative variance = Cashier Short (cashier owes the difference)
```

**MPesa, Card, Fleet, and Credit sales are NEVER included in the cash till calculation.** They have their own reconciliation paths:
- MPesa: match against MPESA Business App transaction report
- Card: match against terminal batch settlement report
- Fleet: match against Sales Invoice AR report

### 8.3 Variance Classification and Required Action

| Variance | Classification | Auto-Approve? | Required Action |
|---|---|---|---|
| Wetstock Loss ≤ 0.3% | Normal | ✓ Yes | None |
| Wetstock Loss 0.3%–0.5% | Elevated | No — flag | Supervisor reviews and approves |
| Wetstock Loss > 0.5% | Critical | No — disputed | Manager investigates, must document before approving |
| Any Wetstock Gain | Suspicious | No — flag | Must explain (over-delivery credit? meter drift?) |
| Cash Variance ≤ KES 50 | Balanced | ✓ Yes | None |
| Cash Variance KES 50–200 | Elevated | No — flag | Supervisor reviews |
| Cash Variance > KES 200 | Critical | No — disputed | Manager investigates |
| Meter Check A Failed | Data Error | No | Correct meter reading via Amendment, recompute |
| Meter Check B Warning | Possible Drift | No | Schedule calibration |
| Meter Check B Critical | Possible Tamper | No | Lock pump, notify Weights & Measures |

### 8.4 Investigation Paths

**Cash variance → investigation steps:**
1. Check `Shift Reconciliation` for the cash variance amount
2. Check `Cashier Session` — is the opening float correct?
3. Check `Cash Events` — are all pickups recorded with correct amounts?
4. Check POS Invoices — any voids? Any unrecorded cash sales?
5. Compare `Elec Cash Sold` across all nozzles with total POS invoice amounts — if meters show more cash than invoices, unrecorded sales are indicated

**Wetstock variance → investigation steps:**
1. Check `Shift Reconciliation → Tank Summaries` for which tank is losing
2. Check the closing dip reading — correct technique? Correct dipstick? Same person each shift for comparison?
3. Check meter readings — are Elec Vol and Mech Vol within tolerance on all nozzles feeding that tank?
4. Check `Fuel Delivery Dip` — was a delivery received? Is dip-measured volume confirmed?
5. Review trend: pull the Wetstock Variance Log for this tank over the past 30 shifts. Is this a one-off or a pattern?

### 8.5 Posting the Shift GL Journal

After reconciliation is approved, create a Journal Entry. Below are the standard templates.

**Template A — Balanced shift (Unleaded 1,731.95 L @ KES 197.60, all cash):**
```
DR  Till — Cashier (Active)       342,313.52    (1,731.95 × 197.60, rounded)
    CR  Fuel Sales — PMS Unleaded               342,313.52

DR  COGS — Fuel PMS Unleaded      207,834.00    (1,731.95 L × WAC KES 120.00)
    CR  Fuel Inventory — PMS Unleaded           207,834.00
```

**Template B — Mixed cash + MPesa (Unleaded + V-Power):**
```
DR  Till — Cashier (Active)       280,000.00    (cash portion)
DR  MPesa Clearing                 62,310.85    (MPesa portion)
DR  Fleet Card Clearing             5,200.00    (V-Power fleet)
    CR  Fuel Sales — PMS Unleaded               342,310.85
    CR  Fuel Sales — PMS V-Power                  5,200.00

DR  COGS — Fuel PMS Unleaded      207,834.00
DR  COGS — Fuel PMS V-Power         3,500.00    (24.5 L × WAC KES 142.86)
    CR  Fuel Inventory — PMS Unleaded           207,834.00
    CR  Fuel Inventory — PMS V-Power              3,500.00
```

**Template C — Cash short KES 300:**
```
DR  Till — Cashier (Active)       342,010.85    (actual cash)
DR  Cash Short / Over                 300.00    (shortage amount)
    CR  Fuel Sales — PMS Unleaded               342,310.85
```

**Template D — Wetstock loss 25 L Unleaded (WAC KES 120.00):**
```
DR  Wetstock Variance — PMS         3,000.00    (25 L × KES 120.00)
    CR  Fuel Inventory — PMS Unleaded            3,000.00
```

> **Important on COGS double-posting:** When a POS Invoice is submitted, ERPNext automatically generates a Stock Ledger Entry that depletes the warehouse at WAC. Do not manually post a COGS Journal Entry for the same transaction — this would double the COGS. The GL template above is correct only if COGS is **not** being posted automatically. Confirm your ERPNext configuration before go-live.

---

## 9. Phase 5 — Reporting

### 9.1 Daily Shift Report

The primary operational document. One per shift, produced at close.

**Contents:**
- Shift reference, date, station, cashier, supervisor, open/close times
- Per-nozzle meter readings: Opening Elec Vol / Cash / Mech, Closing Elec Vol / Cash / Mech, computed sales for each type, cross-check results (Check A, Check B)
- Per-tank dip readings: opening, deliveries, theoretical close, actual close, variance, classification
- Sales by product: volume, rate, revenue, WAC cost, margin
- Cash summary: float, cash sales, pickups, expected close, actual close, variance, verdict
- Non-cash tenders: MPesa, card, fleet/credit — totals and individual counts
- Attendant performance: transactions per attendant, volume per attendant (useful for spotting N/A patterns)
- Overall verdict: Balanced / Elevated / Disputed

**How to build in Frappe:** Create a **Script Report** under `Forecourt Module → Reports → Daily Shift Report`. The report queries `Shift Reconciliation`, `Meter Validation Result`, `Dip Reading`, `Cash Event`, and `Cashier Session` for the selected shift.

### 9.2 Meter Reading Discrepancy Log

A report showing every nozzle where Check A or Check B failed or produced a warning, across all shifts in a date range.

**Purpose:** Identify pumps whose electronic and mechanical systems are drifting apart. A pump that consistently shows Check B warnings (Elec Vol vs Mech Vol diverging by 0.3–0.5%) needs calibration. A pump that suddenly shows a Critical Check B failure needs immediate inspection.

**Columns:** Date | Shift | Pump | Nozzle | Elec Vol Sold | Elec Cash Sold | Mech Vol Sold | Check A Status | Check B Status | Check B % | Action Taken

### 9.3 Wetstock Variance Trend Report

Shows wetstock variance per tank per shift over a selected period.

**Purpose:** A single elevated variance can be noise. A consistent elevated variance on the same tank, at the same time of day, correlating with the same attendant, is a pattern — patterns point to systematic problems.

**Columns:** Date | Shift | Tank | Product | Opening L | Deliveries L | Elec Vol Sales L | Theoretical Closing L | Actual Closing L | Variance L | Variance % | Classification

### 9.4 Cashier Performance Report

Shows each cashier's variance history across all their shifts.

**Columns:** Cashier | Shift Count | Total Volume Handled (L) | Total Cash Handled (KES) | Total Variance (KES) | Avg Variance per Shift | Short Count | Over Count | Largest Single Variance

**Purpose:** A cashier consistently short by KES 50–150 every shift is systematically skimming at levels below the investigation threshold. The trend report surfaces this pattern within 10 shifts.

### 9.5 Attendant Fuel Volume Report

Based on Shell Maanzoni's transaction data, each transaction is linked to a named attendant. This report aggregates volume and transaction count per attendant per shift.

**Columns:** Shift | Attendant | Product | Transaction Count | Volume Dispensed (L) | Revenue (KES) | Avg Litres per Transaction

**Purpose:** Cross-check against nozzle meter deltas. If Attendant X is listed on the POS as having dispensed 500 L from Pump 2, the Pump 2 meter delta should be ≥ 500 L (accounting for other attendants using the same pump). Systematic under-reporting by a specific attendant against meter evidence is a fraud signal.

**Also use to resolve N/A attendant records:** Any transaction where attendant = N/A or blank represents a gap in accountability. This report makes the gap visible so management can investigate and close it operationally.

### 9.6 Delivery Reconciliation Report

Shows every fuel delivery, docket volume, dip-measured volume, and variance.

**Columns:** Date | Supplier | Truck Reg | Docket No | Product | Docket Volume L | Dip Volume L | Variance L | Variance % | Status | Action

**Purpose:** If a specific driver or depot consistently delivers short, and the variance is always just below your dispute threshold, you are being systematically defrauded.

---

## 10. How Everything Relates

```
FORECOURT SHIFT (master record)
│
├─ has many ────────────► METER READINGS (3 types × 2 positions × N nozzles)
│                         │
│                         ├── Electronic Volume (L) ─────────────────────────┐
│                         ├── Electronic Cash (KES) ── ÷ Rate = Vol Check ───┤
│                         └── Manual Mechanical (L) ──────────────────────── ┤
│                                                                             │
│                         computed into ─────────────► METER VALIDATION      │
│                                                      RESULT (per nozzle)   │
│                                                             │               │
│                                                        All pass?            │
│                                                             │               │
│                                                       aggregated to ────────┘
│                                                       nozzle → tank level
│
├─ has many ────────────► DIP READINGS (per tank × Shift Open + Close)
│                         │
│                         └─ provides Opening Stock and Actual Closing Stock
│
├─ is enriched by ──────► FUEL DELIVERY DIPS
│                         │
│                         ├─ links to ─────────────► DIP READINGS (Before + After)
│                         └─ links to ─────────────► PURCHASE RECEIPT (ERPNext native)
│
├─ has one ─────────────► CASHIER SESSION
│                         │
│                         └─ has many ─────────────► CASH EVENTS
│                                                    (Float / Pickup / Payout / Safe Drop)
│
├─ sales posted to ─────► POS INVOICES / SALES INVOICES (ERPNext native)
│
└─ produces one ────────► SHIFT RECONCILIATION
                          │
                          ├─ PRODUCT SUMMARIES (child table)
                          ├─ TANK WETSTOCK SUMMARIES (child table)
                          ├─ NOZZLE METER VALIDATION SUMMARY (child table)
                          │
                          └─ triggers ─────────────► JOURNAL ENTRY (ERPNext native)
```

### 10.1 Single Source of Truth for Each Data Type

| Data | Source of Truth | Notes |
|---|---|---|
| Fuel sold per nozzle (volume) | Electronic Volume Meter delta | Corroborated by Elec Cash and Manual Mech |
| Fuel sold per tank (volume) | Σ Elec Vol Sold for all nozzles → tank | Never aggregate from dip alone |
| Physical fuel in tank | Dip Reading | The auditor; must agree with meter-based theoretical |
| Revenue | POS Invoice / Sales Invoice | Must agree with Elec Cash Sold per nozzle |
| Fuel stock (inventory) | ERPNext Stock Balance (Warehouse) | SR Tank Summaries are a snapshot, not the live figure |
| Delivery volume received | Fuel Delivery Dip (dip-measured) | Overrides docket if variance > 0.5% |
| Cash position | ERPNext Cash/Bank Accounts | Cash Events are the audit trail |
| GL accounting | ERPNext Journal Entry | SR stores the link, not the entry |

---

## 11. Daily Operating Checklist

### Shift Opening (Incoming Cashier + Supervisor)

```
□  Previous shift status = Closed (or Disputed with explanation)
□  Create new Forecourt Shift record
   □  Set Shift Date, Opened At
   □  Assign Primary Cashier and Supervisor (must be different employees)
□  Issue opening float → Cash Event: Float Issued
   □  Supervisor physically counts and hands float to cashier
   □  Both sign the float slip
□  Record opening dip for EVERY active tank → Dip Reading: Shift Open
   □  Tank 1 (Unleaded)
   □  Tank 2 (AGO)
   □  Tank 3 (V-Power)
   □  Check water bottom levels — alert if > 20mm
□  Record opening meter readings for EVERY active nozzle:
   □  All 8 pumps × all active nozzles
   □  For EACH nozzle record THREE readings:
      □  Electronic Volume (L) — from digital display
      □  Electronic Cash (KES) — from digital display
      □  Manual Mechanical (L) — from number wheels
□  Assign pump attendants to pumps in Shift Pump Assignments child table
□  Status → Open (system blocks sales until opening readings are complete)
□  Brief all attendants: do not start dispensing until system confirms Open
```

### During the Shift

```
□  All sales posted in POS with correct payment method
   □  Cash / MPesa / Card / Fleet — never leave as unspecified
   □  Attendant name always selected — never leave as N/A
□  Any fleet/credit fill-up → Sales Invoice (not POS)
□  When till > KES 30,000 (or agreed threshold):
   □  Cash Event: Cash Pickup
   □  Physical envelope with reference number
   □  Journal Entry: DR Safe / CR Till
□  Any fuel delivery:
   □  Dip Reading: Delivery Before → link to Fuel Delivery Dip
   □  Wait 10–15 minutes after pumping
   □  Dip Reading: Delivery After → link to Fuel Delivery Dip
   □  Raise Purchase Receipt (at dip volume if variance > 0.5%)
   □  Mark Fuel Delivery Dip: Accepted or Disputed
□  Spot check: supervisor may take spot dip or spot meter readings
   at any point — record as Reading Type = Spot Check
```

### Shift Closing (Cashier + Supervisor)

```
□  Cashier physically counts till cash
□  Record closing meter readings for EVERY active nozzle:
   □  For EACH nozzle:
      □  Electronic Volume (L) — from digital display
      □  Electronic Cash (KES) — from digital display
      □  Manual Mechanical (L) — from number wheels
□  Record closing dip for EVERY active tank → Dip Reading: Shift Close
□  Update Cashier Session: Actual Cash Count, Counted By, Verified By
□  Enter non-cash tender totals:
   □  MPesa total (from MPESA Business Till Report)
   □  Card total (from terminal batch report)
   □  Fleet/Credit total (from Sales Invoice list this shift)
□  Status → Closing
□  Supervisor clicks "Run Meter Validation" on Shift record
   □  Review Meter Validation Results for all nozzles
   □  Resolve any failures via Amendment readings before proceeding
□  Supervisor clicks "Compute Reconciliation" on Shift Reconciliation
   □  Review Tank Wetstock Summaries
   □  Review Cash Reconciliation
   □  Review Nozzle Meter Validation Summary
□  If all balanced:
   □  Approve Shift Reconciliation
   □  Post GL Journal Entry
   □  Cash Event: Safe Drop
   □  Journal Entry: DR Safe / CR Till
   □  Status → Closed
□  If any Critical variance:
   □  Status → Disputed
   □  Document investigation in Shift Notes
   □  Do not post GL until manager resolves
□  Brief incoming shift cashier and supervisor
```

---

## 12. Business Rules Reference

### Meter Reading Rules

| Rule | Enforcement Method |
|---|---|
| All 3 meter types required at Shift Open and Shift Close | Validate on status transition: Readings Captured requires all 6 records per active nozzle |
| Close totalizer ≥ Open totalizer (for the same type) | Validate on Meter Reading save |
| Submitted meter readings are immutable | Remove Edit on submitted docs; corrections via Amendment type only |
| Amendment requires reason | amendment_reason field is mandatory when reading_position = Amendment |
| Meter Check A tolerance: ≤ KES 5.00 | Enforced in Meter Validation computation |
| Meter Check B warning: 0.3%–0.5% divergence | Flag, allow approval with supervisor confirmation |
| Meter Check B critical: > 0.5% divergence | Block shift close, require manager investigation |
| Meter Check B critical > 1.0%: tamper suspicion | Lock pump (set is_active = No), notify Weights & Measures |

### Shift Rules

| Rule | Enforcement Method |
|---|---|
| Only one Open/Closing shift per station | Validate on Shift save |
| Cashier ≠ Supervisor | Validate on Shift save |
| Status transitions are forward-only (except Disputed → Closing) | Validate on status change |
| GL Journal must be posted before status = Closed | Validate on status Closed |

### Dip Reading Rules

| Rule | Enforcement Method |
|---|---|
| Volume ≤ tank maximum capacity | Validate against Warehouse capacity field |
| Delivery After dip ≥ 10 min after delivery end | Validate against Fuel Delivery Dip.delivery_end |
| Water level > 20 mm → warning | Show alert on save |
| Shift Close dip requires corresponding Shift Open dip | Validate before computing reconciliation |

### Cash Event Rules

| Rule | Enforcement Method |
|---|---|
| Pickup and Safe Drop require authorised_by | Required field |
| authorised_by ≠ session cashier (session-scoped check) | Validate on save using linked Cashier Session |
| One Float Issued event per Cashier Session | Validate on save |
| Physical envelope reference required for Pickup and Safe Drop | Required field |

### Reconciliation Rules

| Rule | Enforcement Method |
|---|---|
| Shift Reconciliation is read-only after computation | Remove Edit button after first save; corrections require re-run |
| Critical variance requires manager approval | Validate: approved_by must be a manager-role user |
| GL Journal must be posted = shift truly closed | Shift status Closed is blocked until gl_journal is set |
| All meter validations must be Pass or reviewed Warning | Block reconciliation compute if any MVR = Fail |

---

## 13. Appendix

### 13.1 Variance Tolerance Quick Reference

| Metric | Normal (Auto-Approve) | Elevated (Review Required) | Critical (Block — Investigate) |
|---|---|---|---|
| Wetstock loss (% of opening + deliveries) | ≤ 0.3% | 0.3% – 0.5% | > 0.5% |
| Wetstock gain | — | Any gain ≤ 0.3% | Any gain > 0.3% |
| Delivery: dip vs docket | ≤ 0.3% | 0.3% – 0.5% | > 0.5% |
| Cash variance (KES) | ≤ KES 50 | KES 50 – 200 | > KES 200 |
| Meter Check A: Cash vs Vol (KES) | ≤ KES 5.00 | KES 5 – 20 | > KES 20 |
| Meter Check B: Mech vs Elec (%) | ≤ 0.3% | 0.3% – 0.5% | > 0.5% |
| Meter Check B: Tamper Threshold | — | — | > 1.0% → Lock pump |

### 13.2 Meter Reading Entry Quick Reference Card

*Print and laminate for each pump island.*

```
┌──────────────────────────────────────────────────────────┐
│  SHIFT OPEN / SHIFT CLOSE — METER READING PROCEDURE     │
│                                                          │
│  For EACH active nozzle on this pump:                    │
│                                                          │
│  1. Electronic Volume (L)                                │
│     Read the LITRES figure from the digital display      │
│     Write ALL digits and decimal places exactly          │
│     Example: 148,730.450                                 │
│                                                          │
│  2. Electronic Cash (KES)                                │
│     Read the KES figure from the digital display         │
│     This number will be millions — write it all          │
│     Example: 29,387,277.20                               │
│                                                          │
│  3. Manual Mechanical (L)                                │
│     Read the number wheels in the panel window           │
│     Use a torch if needed — never estimate               │
│     Example: 148,729.3                                   │
│                                                          │
│  Cross-check before submitting:                          │
│  (Elec Cash ÷ Elec Vol) should ≈ current price          │
│  (Manual Mech) should be close to Elec Vol              │
│  If they don't match — call your supervisor              │
└──────────────────────────────────────────────────────────┘
```

### 13.3 EPRA Price Update Procedure

When EPRA announces a new pump price (typically mid-month):

1. Update the `Standard Selling Rate` on each fuel Item in ERPNext
2. Update the `Shift Rate` field in any active POS Profile
3. **Do not recalculate historical Electronic Cash readings** — they were correct at the time they were taken
4. After the price change, Check A (Elec Cash vs Elec Vol × Rate) will use the new rate for all new shifts
5. For any shift that spans a price change midnight, record the rate in `Shift Notes` and validate manually

### 13.4 Calibration Seal Inspection Procedure

When a pump's Meter Check B enters Critical or Tamper territory:

1. Immediately set the pump's `is_active = No` — no further sales through this pump
2. Record the incident in `Pump Notes` with date, time, and Check B % divergence
3. Call the Kenya Bureau of Standards (KEBS) / Weights & Measures calibration officer
4. Do not allow the pump to be reactivated until the officer inspects and issues a new calibration certificate
5. Record the new seal number and certificate in the `Forecourt Pump` record
6. Reset the nozzle's mechanical reference readings after calibration

---

*End of Document — Version 2.0.0*
*Next step: Phase 2 implementation — creating these doctypes in Frappe Developer Mode*
*Reference data: Shell Maanzoni DAILY_SHIFTTransactions 2026-04-24*
