# Forecourt Management on ERPNext
## A Complete Setup & Implementation Guide for Petrol Station Operators

**Version:** 1.0  
**Audience:** Station Owner / Manager + Developer  
**Stack:** ERPNext 15 / Frappe Framework  
**Currency:** KES (Kenya Shillings)

---

## Table of Contents

1. [Before You Start — Understanding the Business Logic](#1-before-you-start)
2. [The Golden Rules of Forecourt Accounting](#2-the-golden-rules)
3. [Phase 1 — ERPNext Core Setup](#3-phase-1-erpnext-core-setup)
4. [Phase 2 — Custom Doctypes](#4-phase-2-custom-doctypes)
5. [Phase 3 — The Financial Flow](#5-phase-3-the-financial-flow)
6. [Phase 4 — Wetstock Reconciliation](#6-phase-4-wetstock-reconciliation)
7. [Phase 5 — Reporting](#7-phase-5-reporting)
8. [How Everything Relates — The Full Picture](#8-how-everything-relates)
9. [Daily Operating Checklist](#9-daily-operating-checklist)
10. [Business Rules Reference](#10-business-rules-reference)

---

## 1. Before You Start

### 1.1 What Problem Are We Actually Solving?

A petrol station is deceptively simple on the surface — you buy fuel, you sell fuel. But there are **three things that can silently destroy you** if you don't control them:

**Problem 1: Cash leakage**  
Your cashier collects KES 200,000 in a shift. But fuel worth KES 205,000 was dispensed. Where did KES 5,000 go? Without a system that reconciles cash collected against fuel dispensed, you will never know.

**Problem 2: Fuel losses**  
At 50,000 litres/day, even a 0.5% loss is 250 litres — roughly KES 37,500 — every single day. Over a year: KES 13.7 million. This can be theft, meter drift, short deliveries, or a leaking tank. You need to know which.

**Problem 3: Delivery fraud**  
Your supplier's truck docket says 8,000 litres. Your tank dip says 7,796 litres arrived. You just lost 204 litres × KES 150 = KES 30,600. Without before/after dip readings, you will pay for fuel you never received.

### 1.2 What This System Will Do

By the end of this guide you will have a system that:

- Tracks every litre that enters the site (delivery) and leaves it (sales)
- Reconciles every shilling a cashier is responsible for
- Calculates wetstock variance automatically: *theoretical stock vs actual stock*
- Flags losses that exceed normal tolerance
- Produces a daily shift report your accountant can work with
- Creates GL journal entries that map directly to your chart of accounts

### 1.3 What This System Will NOT Do (Yet)

- Connect to a pump controller automatically (manual data entry first)
- Process fleet card payments electronically (manual recording first)
- Integration with mobile money APIs (manual MPesa reconciliation first)

You will add automation later. Getting the data model right first is the entire point.

---

## 2. The Golden Rules

These are the business rules that govern everything. Every doctype, every field, every report in this guide exists to enforce one of these rules.

### Rule 1 — The Shift is the Unit of Everything

A shift is not just a time period. It is a **bounded accountability contract** between the station and one specific cashier. Every litre dispensed, every shilling collected, every delivery received, every dip reading taken — all of it belongs to a shift.

> If it didn't happen in a shift, it didn't happen.

### Rule 2 — Meters Never Lie (Cashiers Can)

A pump meter is a cumulative, sealed, legally calibrated totalizer. It records every litre ever dispensed through that nozzle since installation. It never resets. It only goes forward.

Volume sold = Closing meter − Opening meter

This is the ground truth for fuel sold. If the meter says 500 litres were dispensed and the cashier collected cash for 480 litres, the 20-litre shortfall is a cash problem — not a fuel problem.

### Rule 3 — Dip Readings Are the Tank's Bank Statement

A dip reading (manual or ATG) measures physical fuel in the tank. It is the auditor's check on the meter. At shift close:

```
Theoretical closing stock = Opening dip + Deliveries − Meter sales
Actual closing stock      = Closing dip
Variance                  = Theoretical − Actual
```

If variance exceeds tolerance, investigate before approving the shift.

### Rule 4 — Cash and Fuel Reconcile Separately

Cash reconciliation answers: *Did the cashier collect the right amount of cash?*  
Wetstock reconciliation answers: *Did the right amount of fuel leave the tanks?*

They are linked — a discrepancy in one often explains a discrepancy in the other. But they are computed independently. Never mix them.

### Rule 5 — No Opening Reading, No Sales

A cashier must not be allowed to start selling before opening meter readings are captured. This prevents the oldest forecourt fraud: begin dispensing before the opening reading, pocket the cash for early sales, then record the opening reading after the fact.

### Rule 6 — Deliveries Need Before and After Dips

Never accept a delivery on docket volume alone. Always measure:
- **Before dip** — establishes what was in the tank before the truck arrived
- **After dip** — establishes what is in the tank after pumping

If `(After dip − Before dip)` differs from the docket by more than 0.5%, dispute the delivery with the supplier immediately.

### Rule 7 — Dual Control on Cash Movements

Any time cash leaves the till mid-shift (a pickup to the safe) or at shift close (a safe drop), two people must be present: the cashier and a supervisor. Both must sign. This is recorded in the system as an authorised cash event.

---

## 3. Phase 1 — ERPNext Core Setup

### 3.1 Company Setup

Go to **Setup → Company**. Create your company:

| Field | Value |
|---|---|
| Company Name | Your station trading name (e.g. Shell Maanzoni Service Station) |
| Default Currency | KES |
| Country | Kenya |
| Fiscal Year | January–December (or your trading year) |

### 3.2 Chart of Accounts

ERPNext creates a default chart of accounts. You need to extend it with accounts specific to a fuel station. Go to **Accounts → Chart of Accounts** and add these accounts under the appropriate parent:

**Income Accounts (under Sales)**

| Account Name | Account Type | Purpose |
|---|---|---|
| Fuel Sales — PMS | Income Account | Revenue from petrol sales |
| Fuel Sales — AGO | Income Account | Revenue from diesel sales |
| Fuel Sales — DPK | Income Account | Revenue from kerosene sales |
| Lubricant Sales | Income Account | Oils and lubricants |
| Shop Sales | Income Account | Convenience store |

**Expense Accounts (under Cost of Goods Sold)**

| Account Name | Account Type | Purpose |
|---|---|---|
| COGS — Fuel PMS | Expense Account | Cost of petrol sold |
| COGS — Fuel AGO | Expense Account | Cost of diesel sold |
| COGS — Fuel DPK | Expense Account | Cost of kerosene sold |
| Wetstock Variance — PMS | Expense Account | Fuel losses PMS |
| Wetstock Variance — AGO | Expense Account | Fuel losses AGO |
| Cash Short / Over | Expense Account | Cashier variances |

**Asset Accounts (under Current Assets)**

| Account Name | Account Type | Purpose |
|---|---|---|
| Fuel Inventory — PMS | Stock Account | Physical PMS in tanks |
| Fuel Inventory — AGO | Stock Account | Physical AGO in tanks |
| Fuel Inventory — DPK | Stock Account | Physical DPK in tanks |
| MPesa Clearing | Bank Account | MPesa collections (settled daily) |
| Card Payment Clearing | Bank Account | Card terminal settlements |
| Safe — Main | Cash Account | Site safe balance |
| Till — Cashier 1 | Cash Account | Active till |

### 3.3 Fuel Items (Products)

Go to **Stock → Items → New Item** for each fuel grade.

**Example: AGO (Diesel)**

| Field | Value |
|---|---|
| Item Name | Automotive Gas Oil (AGO) |
| Item Code | FUEL-AGO |
| Item Group | Fuel |
| Stock UOM | Litre |
| Is Stock Item | ✓ Yes |
| Maintain Stock | ✓ Yes |
| Default Warehouse | AGO Tank 1 |
| Standard Selling Rate | 150.00 (current EPRA pump price) |
| Standard Buying Rate | 118.00 (depot purchase price) |
| Valuation Method | Moving Average (WAC) |

Repeat for PMS (petrol) and DPK (kerosene).

> **Why Moving Average (WAC)?** Each fuel delivery arrives at a slightly different cost. Moving average blends the cost automatically. When you sell 1 litre, the system calculates profit against the blended cost — not the last delivery cost or the first delivery cost. This is the industry-standard valuation for fuel.

### 3.4 Tanks as Warehouses

In ERPNext, a tank is modelled as a **Warehouse**. This gives it full stock tracking capability.

Go to **Stock → Warehouses → New Warehouse**

| Field | Value |
|---|---|
| Warehouse Name | AGO Tank 1 |
| Warehouse Type | (create type: Fuel Tank) |
| Parent Warehouse | Your Station |
| Is Group | No |

Create one warehouse per physical tank. If you have two AGO tanks, create "AGO Tank 1" and "AGO Tank 2". Do not combine them — you need separate dip readings per tank.

### 3.5 Employees

Go to **HR → Employee → New Employee** for:
- Each cashier
- Each pump attendant
- The station manager

Minimum required fields: Employee Name, Date of Joining, Department (Forecourt Operations), Designation (Cashier / Attendant / Manager).

### 3.6 Suppliers

Go to **Buying → Suppliers → New Supplier** for your fuel depot(s). This is where fuel purchase orders and delivery receipts will be raised.

### 3.7 Customers (Credit Accounts)

For any customer who fills up on account (fleet companies, credit customers), go to **Selling → Customer → New Customer**. Set Payment Terms to 30 days (or whatever you agreed). These become Accounts Receivable entries when sales are posted.

---

## 4. Phase 2 — Custom Doctypes

These are the doctypes that ERPNext does not have out of the box. You will create them in **Frappe → Customise → DocType**. Each section below describes the fields, the validations to apply, and the business reason for each design decision.

### 4.1 Forecourt Shift

**What it is:** The master record for a shift. Everything else — readings, cash events, reconciliation — hangs off this record.

**Business reason:** A shift is not just a start and end time. It is a chain of evidence. The opening meter reading, every sale, every pickup, the closing dip — all are linked here. If you ever need to investigate a variance or dispute, this is where you start.

**DocType Name:** `Forecourt Shift`

| Field Label | Field Name | Type | Required | Notes |
|---|---|---|---|---|
| Shift Number | shift_number | Data | ✓ | Auto-generated: SHIFT-0001, SHIFT-0002... |
| Station | station | Link → Branch | ✓ | Which physical station |
| Shift Date | shift_date | Date | ✓ | |
| Opening Time | opened_at | Datetime | ✓ | |
| Closing Time | closed_at | Datetime | | Set on close |
| Cashier | cashier | Link → Employee | ✓ | Primary cashier for this shift |
| Supervisor | supervisor | Link → Employee | ✓ | Must be different from Cashier |
| Status | status | Select | ✓ | Open / Closing / Closed / Disputed |
| Float Amount | float_amount | Currency | ✓ | Cash given to cashier at start |
| Notes | notes | Text | | |

**Validations to configure:**
- `Cashier` and `Supervisor` must be different employees
- `closed_at` must be after `opened_at`
- Status can only transition forward: Open → Closing → Closed (or Disputed)
- Only one shift can have status `Open` per station at any time

**Child Table — Shift Pump Attendants:**

| Field | Type | Notes |
|---|---|---|
| Employee | Link → Employee | |
| Pump | Link → Forecourt Pump | |
| Assigned At | Datetime | |
| Relieved At | Datetime | Optional — for mid-shift handover |

---

### 4.2 Forecourt Pump

**What it is:** A physical dispensing unit on the forecourt.

**Business reason:** You need to know which pump each nozzle is on, and which nozzle connects to which tank. This mapping is the foundation of the entire meter reading system.

**DocType Name:** `Forecourt Pump`

| Field Label | Field Name | Type | Required | Notes |
|---|---|---|---|---|
| Pump Number | pump_number | Int | ✓ | Physical number on the island |
| Pump Name | pump_name | Data | ✓ | e.g. "Island 1 — Pump A" |
| Make | make | Data | | Tokheim, Wayne, Gilbarco, etc. |
| Model | model | Data | | |
| Serial Number | serial_number | Data | | For maintenance records |
| Is Active | is_active | Check | ✓ | Default: Yes |
| Installation Date | installation_date | Date | | |
| Last Calibration | last_calibration_date | Date | | |
| Next Calibration Due | next_calibration_due | Date | | |
| Seal Number | weights_measures_seal | Data | | Legal calibration seal |

**Child Table — Nozzles:**

Each pump has one or more nozzles. A nozzle connects a pump to a specific tank and product.

| Field | Type | Notes |
|---|---|---|
| Nozzle Number | Int | 1-based within the pump |
| Tank | Link → Warehouse | Must be a fuel tank warehouse |
| Fuel Product | Link → Item | Must be a fuel item |
| Is Active | Check | Default: Yes |

> **Why model nozzles as a child table?** Because you need to know exactly which physical hose dispensed fuel. Nozzle 1A might dispense AGO; Nozzle 1B might dispense PMS on the same pump unit. Each has its own meter. Without this mapping, you cannot link a meter reading to the correct tank.

---

### 4.3 Meter Reading

**What it is:** A single observation of a nozzle's cumulative totalizer value.

**Business reason:** The meter never resets. You are always reading a number like 148,730.450 litres. Volume sold is always computed as `close − open`. The reading itself is just a snapshot in time.

**Critical design principle: NEVER EDIT A METER READING.** If a reading was entered incorrectly, create a correction reading. The original must remain as evidence.

**DocType Name:** `Meter Reading`

| Field Label | Field Name | Type | Required | Notes |
|---|---|---|---|---|
| Reading Reference | name | Data | Auto | e.g. MR-2024-0001 |
| Shift | shift | Link → Forecourt Shift | ✓ | |
| Nozzle (Pump) | pump | Link → Forecourt Pump | ✓ | |
| Nozzle Number | nozzle_number | Int | ✓ | Which nozzle on the pump |
| Reading Type | reading_type | Select | ✓ | Shift Open / Shift Close / Spot / Amendment |
| Observed At | observed_at | Datetime | ✓ | When the reading was physically taken |
| Totalizer (Litres) | totalizer_l | Float | ✓ | Raw meter display value |
| Recorded By | recorded_by | Link → User | ✓ | Auto-filled from session |
| Notes | notes | Text | | |
| Superseded By | superseded_by | Link → Meter Reading | | Set when this reading is corrected |

**Validations:**
- `Totalizer (Litres)` must be greater than zero
- For `reading_type = Shift Close`: totalizer must be ≥ the corresponding Shift Open reading for the same nozzle
- Once saved, the totalizer value cannot be edited (enforce via code)

**The Volume Sold Calculation** (computed, never stored):
```
Volume Sold = Close Totalizer − Open Totalizer
```
This is calculated at shift close time, not stored in the Meter Reading record itself. Storing it would allow manipulation.

---

### 4.4 Dip Reading

**What it is:** A physical measurement of fuel volume inside a tank at a specific moment.

**Business reason:** A dip reading is the only way to verify that the theoretical stock (calculated from meters and deliveries) matches the physical reality in the tank. It is the auditor that catches meter drift, theft, and leaks.

**DocType Name:** `Dip Reading`

| Field Label | Field Name | Type | Required | Notes |
|---|---|---|---|---|
| Reading Reference | name | Data | Auto | e.g. DIP-2024-0001 |
| Shift | shift | Link → Forecourt Shift | | Can be null for routine readings |
| Tank | tank | Link → Warehouse | ✓ | Must be a Fuel Tank warehouse |
| Reading Type | reading_type | Select | ✓ | Shift Open / Shift Close / Delivery Before / Delivery After / Routine |
| Observed At | observed_at | Datetime | ✓ | |
| Level (mm) | level_mm | Float | | From dipstick measurement |
| Volume Observed (L) | volume_observed_l | Float | ✓ | Converted from mm using calibration chart |
| Temperature (°C) | temperature_c | Float | | If measured |
| Water Level (mm) | water_level_mm | Float | | Alert if > 20mm |
| Source | source | Select | ✓ | Manual / ATG |
| Recorded By | recorded_by | Link → User | ✓ | |
| Notes | notes | Text | | |

**Validations:**
- `Volume Observed` must be > 0 and ≤ tank capacity
- If `Water Level > 20`, show a warning: *"High water level detected — possible contamination. Alert manager."*
- For `Delivery After` type: the `observed_at` time must be at least 10 minutes after the delivery start time (settling period)

---

### 4.5 Cash Event

**What it is:** Any movement of cash related to a shift — the opening float, a mid-shift pickup, or the final safe drop.

**Business reason:** The cash equation is only meaningful if every cash movement is recorded. A pickup that is not in the system means the cashier appears short by that amount. A float that is not recorded means the system cannot calculate expected closing cash.

**DocType Name:** `Cash Event`

| Field Label | Field Name | Type | Required | Notes |
|---|---|---|---|---|
| Event Reference | name | Data | Auto | CE-2024-0001 |
| Shift | shift | Link → Forecourt Shift | ✓ | |
| Cashier Session | cashier_session | Link → Cashier Session | ✓ | |
| Event Type | event_type | Select | ✓ | Float Issued / Pickup / Payout / Safe Drop |
| Amount (KES) | amount | Currency | ✓ | |
| Authorised By | authorised_by | Link → User | ✓ | Must differ from cashier |
| Occurred At | occurred_at | Datetime | ✓ | |
| Envelope / Reference | reference | Data | | Safe envelope number, voucher ref |
| Notes | notes | Text | | |

**Validations:**
- For `Pickup` and `Safe Drop`: `Authorised By` must not be the same user as the cashier on the session. Enforce this in code.
- `Amount` must be positive and non-zero
- Only one `Float Issued` event per cashier session is permitted

---

### 4.6 Cashier Session

**What it is:** A bounded accountability period for a specific cashier on a specific till within a shift.

**Business reason:** If two cashiers share a till without a formal session handover, you cannot determine which cashier caused a variance. One session = one cashier = one accountability unit.

**DocType Name:** `Cashier Session`

| Field Label | Field Name | Type | Required | Notes |
|---|---|---|---|---|
| Session Reference | name | Data | Auto | CS-2024-0001 |
| Shift | shift | Link → Forecourt Shift | ✓ | |
| Cashier | cashier | Link → Employee | ✓ | |
| Till ID | till_id | Data | ✓ | e.g. "TILL-01" |
| Float Amount | float_amount | Currency | ✓ | Opening float for this session |
| Session Opened At | session_opened_at | Datetime | ✓ | |
| Session Closed At | session_closed_at | Datetime | | |
| Actual Cash at Close | actual_cash_close | Currency | | Entered by cashier at session end |
| Is Primary | is_primary | Check | ✓ | First cashier of the shift |
| Notes | notes | Text | | |

---

### 4.7 Shift Reconciliation

**What it is:** The computed summary of a closed shift. This is the system's verdict on whether the shift balanced.

**Business reason:** The reconciliation is not a data entry form. It is a calculated output. It reads from meter readings, dip readings, sales records, and cash events, then computes the verdict. It must never be manually edited. If something is wrong, fix the source data and re-run.

**DocType Name:** `Shift Reconciliation`

| Field Label | Field Name | Type | Notes |
|---|---|---|---|
| Reconciliation Ref | name | Data | SR-2024-0001 |
| Shift | shift | Link → Forecourt Shift | 1-to-1 relationship |
| Reconciliation Date | reconciliation_date | Date | |
| **Volume Summary** | | | |
| Total Volume Sold (L) | total_volume_l | Float | Sum of all nozzle meter deltas |
| **Revenue Summary** | | | |
| Total Gross Revenue | total_gross_revenue | Currency | Volume × unit price |
| **Cash Reconciliation** | | | |
| Expected Cash | expected_cash | Currency | Computed — see formula below |
| Actual Cash Declared | actual_cash | Currency | From cashier session |
| Cash Variance | cash_variance | Currency | Actual − Expected |
| **Non-Cash Tenders** | | | |
| MPesa Total | mpesa_total | Currency | |
| Card Total | card_total | Currency | |
| Fleet / Credit Total | fleet_credit_total | Currency | |
| **Wetstock** | | | |
| Total Wetstock Variance (L) | wetstock_variance_l | Float | Theoretical − Actual |
| **Outcome** | | | |
| Cash Balanced | is_cash_balanced | Check | System-set |
| Wetstock Balanced | is_wetstock_balanced | Check | System-set |
| Requires Approval | requires_approval | Check | System-set |
| Approved By | approved_by | Link → User | |
| Approved At | approved_at | Datetime | |
| GL Journal | gl_journal | Link → Journal Entry | Set after GL posting |

**Child Table 1 — Product Summaries:**

| Field | Type | Notes |
|---|---|---|
| Fuel Product | Link → Item | |
| Volume Sold (L) | Float | |
| Unit Price | Currency | |
| Gross Revenue | Currency | |
| WAC Unit Cost | Currency | |
| COGS | Currency | |

**Child Table 2 — Tank Wetstock Summaries:**

| Field | Type | Notes |
|---|---|---|
| Tank | Link → Warehouse | |
| Fuel Product | Link → Item | |
| Opening Stock (L) | Float | From opening dip |
| Deliveries (L) | Float | From purchase receipts |
| Sales (L) | Float | From nozzle meter delta |
| Theoretical Closing (L) | Float | Opening + Deliveries − Sales |
| Actual Closing (L) | Float | From closing dip |
| Variance (L) | Float | Theoretical − Actual |
| Variance % | Float | |
| Classification | Select | Normal / Elevated / Critical / Gain |

---

### 4.8 Fuel Delivery Dip Record

**What it is:** Links a procurement goods receipt (fuel delivery) to the before/after dip readings.

**Business reason:** A delivery is received by Procurement (Purchase Receipt in ERPNext). But the forecourt needs to verify the volume. This doctype is the bridge between the two.

**DocType Name:** `Fuel Delivery Dip`

| Field Label | Field Name | Type | Notes |
|---|---|---|---|
| Reference | name | Data | FDD-2024-0001 |
| Shift | shift | Link → Forecourt Shift | |
| Purchase Receipt | purchase_receipt | Link → Purchase Receipt | The procurement record |
| Tank | tank | Link → Warehouse | |
| Truck Registration | truck_reg | Data | |
| Docket Number | docket_number | Data | Supplier's delivery note number |
| Docket Volume (L) | docket_volume_l | Float | What supplier claims |
| Delivery Start Time | delivery_start | Datetime | |
| Delivery End Time | delivery_end | Datetime | |
| Dip Before (L) | dip_before_l | Float | From Dip Reading linked below |
| Dip After (L) | dip_after_l | Float | |
| Dip-Measured Volume (L) | dip_measured_l | Float | Computed: After − Before |
| Variance vs Docket (L) | delivery_variance_l | Float | Dip-Measured − Docket |
| Variance % | delivery_variance_pct | Float | |
| Status | status | Select | Accepted / Disputed / Pending |

---

## 5. Phase 3 — The Financial Flow

### 5.1 Overview

Every operational event at the station must eventually produce an accounting entry. The table below maps each business event to the correct ERPNext transaction.

| Business Event | ERPNext Transaction | Posted By |
|---|---|---|
| Fuel delivery arrives | Purchase Receipt → Stock Entry | Procurement / Manager |
| Cashier starts shift | Cashier Session created | Cashier |
| Customer buys fuel (cash) | POS Invoice — payment: Cash | Cashier |
| Customer buys fuel (MPesa) | POS Invoice — payment: MPesa Clearing | Cashier |
| Fleet customer fills up | Sales Invoice → AR | Cashier |
| Cash pickup to safe | Cash Event (PICKUP) + Journal Entry | Supervisor |
| Shift closes — balanced | Journal Entry (shift GL) | System / Manager |
| Wetstock loss | Stock Reconciliation | System / Manager |
| Cashier short/over | Journal Entry (Cash Short/Over) | System |

### 5.2 Recording a Fuel Delivery

**Step 1: Before the truck arrives**
- Open a new `Dip Reading` for the tank: Type = `Delivery Before`
- Record the observed volume (litres or from dipstick in mm)
- This is your pre-delivery stock

**Step 2: Truck arrives and pumps**
- Record truck registration and docket number in `Fuel Delivery Dip`
- Record delivery start time

**Step 3: After settling (wait 10–15 minutes)**
- Take the after-delivery dip: new `Dip Reading`, Type = `Delivery After`
- The system computes: `Dip-Measured Volume = After − Before`

**Step 4: Raise Purchase Receipt in ERPNext**
- Go to **Buying → Purchase Receipt → New**
- Supplier: your fuel depot
- Item: the fuel grade (AGO, PMS, etc.)
- Qty: **use the dip-measured volume, not the docket volume** (unless they match within 0.5%)
- Rate: from the supplier invoice
- Warehouse: the specific tank warehouse

**Step 5: If docket vs dip varies by > 0.5%**
- Set `Fuel Delivery Dip` status to `Disputed`
- Call the supplier immediately with your dip records
- Do not post the Purchase Receipt until resolved

**Accounting effect of Purchase Receipt:**
```
DR  Fuel Inventory — AGO       944,000.00
    CR  Accounts Payable                    944,000.00
```

### 5.3 Recording Sales During a Shift

Use ERPNext's POS module for each sale.

**Setup required first:**
- Create a POS Profile for the station
- Set default payment methods: Cash, MPesa, Card, Fleet/Credit
- Link the POS to the correct company and warehouse

**For each customer transaction:**
1. Select the fuel product (AGO, PMS, DPK)
2. Enter the volume in litres
3. The price auto-fills from the item's selling rate
4. Select payment method
5. Submit

> **Note on volume vs amount:** ERPNext POS works naturally with litres as the unit of measure. Price per litre × litres = total. This is correct for fuel.

**For fleet/credit customers:**
- Use **Sales Invoice** (not POS) — this creates an AR balance
- Set customer to the fleet company
- Payment terms: 30 days (or per agreement)
- The cashier notes the sale in the shift but no cash enters the till

### 5.4 Cash Pickups During the Shift

When the till accumulates more than the agreed threshold (e.g. KES 20,000), a supervisor must remove the excess:

1. Create a new `Cash Event`: Type = `Pickup`
2. Amount = the cash being removed
3. Authorised By = supervisor (different from cashier)
4. Assign an envelope number for the physical cash bag
5. Also create a manual Journal Entry:

```
DR  Safe — Main          20,000.00
    CR  Till — Cashier 1             20,000.00
```

---

## 6. Phase 4 — Wetstock Reconciliation

### 6.1 The Reconciliation Formula

At the end of every shift, for every tank:

```
Opening Stock (L)
  + Deliveries Received (L)       ← from Purchase Receipts this shift
  − Volume Sold (L)               ← from meter readings (close − open)
  = Theoretical Closing Stock (L)

Actual Closing Stock (L)          ← from closing dip reading
= physical fuel measured in tank

Variance (L) = Theoretical − Actual

Variance % = Variance ÷ (Opening + Deliveries) × 100
```

### 6.2 Running the Reconciliation

At shift close, the manager or supervisor:

1. Confirms all closing meter readings are entered (one per active nozzle)
2. Confirms closing dip readings are entered (one per active tank)
3. Confirms the cashier has declared their closing cash count
4. Opens the `Shift Reconciliation` for this shift
5. Clicks "Compute Reconciliation" (a custom button that runs the calculation)
6. Reviews the output
7. Approves or disputes

### 6.3 Variance Classification and Action

| Variance % | Classification | System Action | Manager Action |
|---|---|---|---|
| Loss ≤ 0.3% | **Normal** | Auto-approve | None required |
| Loss 0.3%–0.5% | **Elevated** | Flag for review | Must review, can approve |
| Loss > 0.5% | **Critical** | Shift → Disputed | Must investigate, must approve |
| Any gain | **Gain** | Flag as suspicious | Must explain before approving |

**Why are gains suspicious?** A gain means you appear to have more fuel than you should. This can happen when a delivery was dip-measured as more than the docket (you received a bonus), or when meters are under-reading (dispensing more than recorded — meaning customers are getting more fuel than they paid for). Gains are always investigated.

### 6.4 Cash Reconciliation Formula

```
Expected Cash at Session Close
  = Opening Float
  + Cash Sales (payment method = Cash only)
  − Cash Pickups (sum of PICKUP events this session)
  − Cash Payouts (any authorised PAYOUT events)

Cash Variance = Actual Cash Declared − Expected Cash

Positive = Cashier Over (surplus — unusual, investigate)
Negative = Cashier Short (shortage — cashier owes)
```

**Critical:** MPesa, card, fleet, and credit sales are EXCLUDED from this calculation. They have their own reconciliation. Only cash changes the till balance.

### 6.5 Posting the Shift GL Journal

After the reconciliation is approved, create a Journal Entry in ERPNext. This is the accounting record of the shift's financial activity.

**Template: Balanced shift — AGO 2,000 L @ KES 150/L, all cash:**

```
DR  Till — Cashier 1         300,000.00
    CR  Fuel Sales — AGO                  300,000.00

DR  COGS — Fuel AGO          240,000.00   (2,000 L × WAC KES 120)
    CR  Fuel Inventory — AGO              240,000.00
```

**Template: Cash short KES 500:**

```
DR  Till — Cashier 1         199,500.00
DR  Cash Short / Over            500.00
    CR  Fuel Sales — AGO                  200,000.00
```

**Template: Wetstock loss — 50 L AGO (WAC KES 120):**

```
DR  Wetstock Variance — AGO    6,000.00
    CR  Fuel Inventory — AGO               6,000.00
```

**Template: MPesa sales included:**

```
DR  Till — Cashier 1          150,000.00   (cash portion)
DR  MPesa Clearing             50,000.00   (MPesa portion)
    CR  Fuel Sales — AGO                   200,000.00
```

---

## 7. Phase 5 — Reporting

### 7.1 Daily Shift Report

This is the primary operational document. One per shift, produced at close.

**Contents:**
- Shift number, date, cashier, supervisor
- Opening and closing times
- Per-nozzle meter readings and volume sold
- Per-tank dip readings and wetstock variance
- Sales by product: volume, price, revenue
- Cash summary: expected, actual, variance
- Non-cash tenders: MPesa, card, fleet, credit
- Overall verdict: Balanced / Disputed

**How to build in ERPNext:** Create a **Script Report** under `Forecourt → Reports → Daily Shift Report`. The report queries `Shift Reconciliation`, `Meter Reading`, `Dip Reading`, and `Cash Event` for the selected shift.

### 7.2 Wetstock Variance Log

A trend report showing wetstock variance per tank per shift over a selected date range.

**Why this matters:** A single elevated variance can be noise. A consistent elevated variance on the same tank, same nozzle, same shift pattern is a pattern — and patterns indicate systematic problems (meter drift, specific attendant, specific time of day).

**Columns:** Date | Shift | Tank | Opening L | Deliveries L | Sales L | Theoretical L | Actual L | Variance L | Variance % | Classification

**Build:** Script Report querying `Shift Reconciliation → Tank Summaries` child table filtered by date range.

### 7.3 Cashier Performance Report

Shows each cashier's variance history across all their shifts.

**Columns:** Cashier | Shift Count | Total Cash Handled | Total Variance | Average Variance | Short Count | Over Count

**Why this matters:** A cashier who is consistently short by small amounts (KES 50–200) every shift is either careless or systematically skimming at low enough levels to avoid detection. The trend is the signal.

### 7.4 Delivery Reconciliation Report

Shows every delivery, docket volume, dip-measured volume, and variance.

**Columns:** Date | Supplier | Truck Reg | Docket No | Product | Docket Volume L | Dip Volume L | Variance L | Variance % | Status

**Why this matters:** If a specific supplier's driver consistently delivers short, and the variance is always just below your dispute threshold, you are being systematically defrauded.

### 7.5 Stock Valuation Report

ERPNext's built-in **Stock Balance** report will show current litres in each tank warehouse and the WAC cost. No customisation needed. Run it anytime to see your current fuel inventory value.

---

## 8. How Everything Relates — The Full Picture

This section explains how all the custom doctypes connect to each other and to ERPNext's native documents. Understanding these relationships is what separates a system that produces numbers from a system you can trust.

```
SHIFT (the master record)
│
├─ has one ──────────────► CASHIER SESSION
│                          │
│                          ├─ has many ──► CASH EVENTS
│                          │              (Float / Pickup / Safe Drop)
│                          │
│                          └─ produces ──► CASH RECONCILIATION
│                                         (part of Shift Reconciliation)
│
├─ has many ─────────────► METER READINGS
│                          (one per nozzle × Open + Close)
│                          │
│                          └─ provides ──► Volume Sold per Nozzle
│                                         └─ rolled up to ──► Volume per Tank
│
├─ has many ─────────────► DIP READINGS
│                          (one per tank × Open + Close)
│                          │
│                          └─ provides ──► Opening and Closing Stock
│
├─ is enriched by ───────► FUEL DELIVERY DIPS
│                          │
│                          ├─ links to ──► PURCHASE RECEIPT (ERPNext native)
│                          │              (the official stock-in record)
│                          │
│                          └─ provides ──► Deliveries volume for wetstock
│
├─ sales posted to ──────► POS INVOICES / SALES INVOICES (ERPNext native)
│                          (the official revenue record)
│                          │
│                          └─ cash sales ──► feed Cash Reconciliation
│
└─ produces one ─────────► SHIFT RECONCILIATION
                           │
                           ├─ PRODUCT SUMMARIES (child table)
                           │  Volume sold, revenue, COGS per product
                           │
                           ├─ TANK SUMMARIES (child table)
                           │  Wetstock calculation per tank
                           │
                           └─ triggers ──► JOURNAL ENTRY (ERPNext native)
                                          (the official accounting record)
```

### 8.1 The Chain of Evidence

When you investigate any problem, follow this chain:

**Cash variance → investigation path:**
1. `Shift Reconciliation` shows cash variance amount
2. Check `Cashier Session` — is the float correctly recorded?
3. Check `Cash Events` — are all pickups properly recorded?
4. Check `POS Invoices` — are all cash sales recorded? Any voids?
5. If meter readings suggest more volume was sold than invoiced, meter vs invoice fraud is indicated

**Wetstock variance → investigation path:**
1. `Shift Reconciliation → Tank Summaries` shows variance per tank
2. Check the closing dip reading — was it taken correctly? By whom?
3. Check meter readings — are they plausible? Any anomalies?
4. Check `Fuel Delivery Dips` — was a delivery received? Is there a dispute outstanding?
5. Compare with previous shifts — is this tank consistently losing?

### 8.2 The Single Source of Truth for Each Piece of Data

| Data | Source of Truth | Never Duplicate To |
|---|---|---|
| Fuel in tanks (inventory) | ERPNext Stock Balance (Warehouse) | Shift Reconciliation is a summary only |
| Fuel sold (revenue) | ERPNext POS / Sales Invoice | Meter readings confirm volume, not override |
| Supplier invoices | ERPNext Purchase Invoice | Delivery Dip enriches, not replaces |
| Cash on hand | ERPNext Cash/Bank accounts | Cash Events are accountability trail |
| GL journal | ERPNext Journal Entry | Shift Reconciliation stores the link only |

---

## 9. Daily Operating Checklist

### Shift Opening (incoming cashier + supervisor)

```
□  Previous shift fully closed in system (status = Closed or approved if Disputed)
□  Create new Forecourt Shift
□  Record float issued to cashier → Cash Event: Float Issued
□  Record opening dip for every active tank → Dip Reading: Shift Open
□  Record opening meter for every active nozzle → Meter Reading: Shift Open
□  Assign pump attendants to pumps
□  Confirm: no sales can begin until all opening readings are entered
□  Set shift status → Open
```

### During the Shift

```
□  All sales posted in POS with correct payment method
□  Any fleet/credit sales posted as Sales Invoice (not POS)
□  Cash pickup performed when till exceeds KES 20,000
     → Cash Event: Pickup
     → Journal Entry: DR Safe / CR Till
□  Any fuel delivery:
     → Dip Reading: Delivery Before
     → Wait 10–15 minutes after pumping
     → Dip Reading: Delivery After
     → Raise Purchase Receipt
     → Create Fuel Delivery Dip record
     → If dip vs docket varies > 0.5%: mark as Disputed, call supplier
```

### Shift Closing (cashier + supervisor)

```
□  Cashier physically counts till cash
□  Enter closing meter readings for every active nozzle → Meter Reading: Shift Close
□  Enter closing dip for every active tank → Dip Reading: Shift Close
□  Cashier declares closing cash count → update Cashier Session: Actual Cash at Close
□  Enter MPesa total from M-PESA app or till report
□  Enter card total from terminal batch report
□  Enter fleet/credit total (matches Sales Invoices raised this shift)
□  Supervisor clicks "Compute Reconciliation" on Shift Reconciliation
□  Review variance:
     □  Cash variance within KES 50? → Approve
     □  Wetstock loss ≤ 0.3%? → Approve
     □  Any critical variance? → Dispute and investigate
□  Post GL Journal Entry
□  Cash to safe → Cash Event: Safe Drop → Journal Entry: DR Safe / CR Till
□  Set shift status → Closed
□  Brief next shift's cashier
```

---

## 10. Business Rules Reference

A quick reference for validations to enforce — either in Frappe's DocType validation scripts or through user training.

### Shift Rules

| Rule | Enforcement |
|---|---|
| Only one Open shift per station | Validate on Shift save |
| Supervisor ≠ Cashier | Validate on Shift save |
| Cannot close a shift with missing readings | Validate before computing reconciliation |
| Status can only move forward | Validate on status change |

### Meter Reading Rules

| Rule | Enforcement |
|---|---|
| Close totalizer ≥ Open totalizer | Validate on save |
| Cannot edit a saved meter reading | Remove standard Edit button; corrections are new rows |
| Opening reading required before first sale | Enforce in POS profile setup or training |

### Dip Reading Rules

| Rule | Enforcement |
|---|---|
| Volume cannot exceed tank capacity | Validate against tank capacity field |
| Delivery After dip must be ≥ 10 min after delivery start | Validate against Fuel Delivery Dip delivery_end |
| Water level > 20mm triggers alert | Show warning on save |

### Cash Event Rules

| Rule | Enforcement |
|---|---|
| Pickup and Safe Drop require authorised_by | Required field |
| authorised_by ≠ cashier | Validate on save |
| One Float Issued per session | Validate on save |

### Reconciliation Rules

| Rule | Enforcement |
|---|---|
| Reconciliation cannot be manually edited | Read-only after computation |
| Critical variance requires manager approval | Validate before posting GL |
| GL journal not posted = shift not truly closed | Validate shift status logic |

---

## Appendix — Variance Tolerance Quick Reference

| Metric | Normal | Elevated | Critical |
|---|---|---|---|
| Cash variance | ≤ KES 50 | KES 50–200 | > KES 200 |
| Wetstock loss | ≤ 0.3% | 0.3%–0.5% | > 0.5% |
| Wetstock gain | — | — | Any gain > 0.3% |
| Delivery dip vs docket | ≤ 0.3% | 0.3%–0.5% | > 0.5% |

These are defaults. Adjust in ERPNext's **System Settings** or a custom **Site Preferences** doctype once you have baseline data from your station.

---

*End of Guide v1.0*  
*Next: Phase 2 implementation — creating the custom doctypes in Frappe*
