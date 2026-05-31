# Forecourt Management Module — Integration Architecture
## ERPNext / Frappe Code Reference

**Version:** 3.1.0 (addendum to v3.0.0 Implementation Guide)
**Station Reference:** Shell Maanzoni Service Station (Anika Global Limited)
**Stack:** ERPNext 15 / Frappe Framework / Python 3.11 / JavaScript (ES2020)
**App Name (assumed):** `forecourt` (custom Frappe app)

---

## Table of Contents

15. [Integration Architecture Overview](#15-integration-architecture-overview)
16. [Stock / Inventory Module Integration](#16-stock--inventory-module-integration)
17. [Accounts / Finance Module Integration](#17-accounts--finance-module-integration)
18. [HR Module Integration](#18-hr-module-integration)
19. [Buying / Procurement Module Integration](#19-buying--procurement-module-integration)
20. [Selling / POS Module Integration](#20-selling--pos-module-integration)
21. [Python — DocType Validation Hooks](#21-python--doctype-validation-hooks)
22. [Python — Shift Status Transition Enforcement](#22-python--shift-status-transition-enforcement)
23. [Python — Reconciliation Computation Script](#23-python--reconciliation-computation-script)
24. [Python — Pre-Fetch Logic for Shift Close](#24-python--pre-fetch-logic-for-shift-close)
25. [Python — GL Journal Entry Creation](#25-python--gl-journal-entry-creation)
26. [Python — Wetstock Variance Calculation](#26-python--wetstock-variance-calculation)
27. [Python — Meter Validation Checks](#27-python--meter-validation-checks)
28. [Python — POS Invoice Blocking Until Shift Opens](#28-python--pos-invoice-blocking-until-shift-opens)
29. [JavaScript — Client-Side Real-Time Cross-Checks](#29-javascript--client-side-real-time-cross-checks)
30. [Dashboard and Report Queries](#30-dashboard-and-report-queries)
31. [hooks.py — Complete Registration](#31-hookspy--complete-registration)

---

## 15. Integration Architecture Overview

### 15.1 How the Forecourt App Relates to ERPNext Modules

The `forecourt` custom app does not replace ERPNext modules — it sits **on top of** them, using their native APIs for all financial, stock, and HR operations. The custom doctypes own the forecourt-specific data layer; ERPNext modules own the accounting, stock, and payroll ledgers.

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    FORECOURT CUSTOM APP                                  │
│                                                                         │
│  Forecourt Pump   Meter Reading   Dip Reading   Fuel Delivery Dip      │
│  Forecourt Shift  Cashier Session Cash Event    Drive-Off Record        │
│  Meter Validation Result          Shift Reconciliation                  │
│                                                                         │
│  ┌─────────────┐ ┌────────────┐ ┌──────────┐ ┌────────────────────┐   │
│  │  Calls      │ │  Calls     │ │  Calls   │ │  Subscribes to     │   │
│  │  Stock APIs │ │  Account   │ │  HR APIs │ │  POS Invoice       │   │
│  │             │ │  APIs      │ │          │ │  on_submit hook    │   │
│  └──────┬──────┘ └─────┬──────┘ └────┬─────┘ └─────────┬──────────┘   │
└─────────┼──────────────┼─────────────┼───────────────────┼─────────────┘
          │              │             │                   │
┌─────────▼──────────────▼─────────────▼───────────────────▼─────────────┐
│                       ERPNEXT CORE                                       │
│                                                                         │
│  ┌─────────────┐  ┌────────────────┐  ┌──────────┐  ┌───────────────┐  │
│  │ Stock /     │  │ Accounts /     │  │ HR /     │  │ Selling /     │  │
│  │ Inventory   │  │ Finance        │  │ Payroll  │  │ POS           │  │
│  │             │  │                │  │          │  │               │  │
│  │ Warehouse   │  │ Journal Entry  │  │ Employee │  │ POS Invoice   │  │
│  │ Item        │  │ GL Entry       │  │          │  │ Sales Invoice │  │
│  │ Stock Ledger│  │ Chart of Accts │  │          │  │ POS Profile   │  │
│  │ Purchase Rcpt│  │ Payment Entry  │  │          │  │               │  │
│  └─────────────┘  └────────────────┘  └──────────┘  └───────────────┘  │
└─────────────────────────────────────────────────────────────────────────┘
```

### 15.2 Data Ownership Boundaries

| Data Domain | Owned By | ERPNext Module Writes To |
|---|---|---|
| Fuel volumes in tanks | Dip Reading (custom) | Warehouse / Stock Ledger |
| Pump meter readings | Meter Reading (custom) | — (no native equivalent) |
| Shift accountability | Forecourt Shift (custom) | — |
| Cash movements | Cash Event (custom) → Journal Entry | GL Entry |
| Revenue | POS Invoice / Sales Invoice | GL Entry, Accounts Receivable |
| Fuel cost / COGS | Shift Reconciliation → Journal Entry | GL Entry |
| Inventory value | Purchase Receipt (native) | Stock Ledger, GL Entry |
| Employee records | Employee (native) | — |
| Delivery dockets | Fuel Delivery Dip → Purchase Receipt | Stock Ledger, GL Entry |

### 15.3 Module Integration Summary

| ERPNext Module | Integration Type | Custom App Role |
|---|---|---|
| **Stock / Inventory** | Read WAC; write via Purchase Receipt & Stock Reconciliation | Calculates COGS using fetched WAC; raises Purchase Receipts for deliveries |
| **Accounts / Finance** | Write Journal Entries; read CoA | Posts shift GL, wetstock adjustments, cash events |
| **HR** | Read Employee records; validate cashier ≠ supervisor | Links all accountability records to Employee |
| **Buying / Procurement** | Write Purchase Receipt | Creates Purchase Receipt after dip-confirmed delivery |
| **Selling / POS** | Validate on POS submission; read POS Invoice totals | Blocks POS if no open shift; aggregates invoice totals for reconciliation |

---

## 16. Stock / Inventory Module Integration

### 16.1 Reading WAC (Weighted Average Cost) for a Fuel Item

WAC is the `valuation_rate` stored on the Item's stock ledger. Always read it **at reconciliation time** from `tabStock Ledger Entry` — never cache it in the custom doctype.

```python
# forecourt/utils/stock.py

import frappe
from frappe.utils import flt


def get_current_wac(item_code: str, warehouse: str) -> float:
    """
    Fetch the current Weighted Average Cost (WAC) for a fuel item
    from the most recent Stock Ledger Entry for the given warehouse.
    
    Returns 0.0 if no stock ledger entry exists yet (pre-first-delivery).
    
    Args:
        item_code: e.g. "FUEL-PMS-UNL"
        warehouse: e.g. "Tank 1 - Unleaded - SMSS"
    
    Returns:
        float: WAC in KES per litre
    """
    result = frappe.db.sql(
        """
        SELECT valuation_rate
        FROM `tabStock Ledger Entry`
        WHERE item_code = %(item_code)s
          AND warehouse = %(warehouse)s
          AND is_cancelled = 0
        ORDER BY posting_date DESC, posting_time DESC, creation DESC
        LIMIT 1
        """,
        {"item_code": item_code, "warehouse": warehouse},
        as_dict=True,
    )
    return flt(result[0].valuation_rate) if result else 0.0


def get_stock_balance(item_code: str, warehouse: str, posting_date: str = None) -> float:
    """
    Fetch the current stock balance in litres for a fuel item in a warehouse.
    Uses ERPNext's built-in get_stock_balance utility.
    
    Args:
        item_code: e.g. "FUEL-PMS-UNL"
        warehouse: e.g. "Tank 1 - Unleaded - SMSS"
        posting_date: optional, defaults to today
    
    Returns:
        float: balance in litres
    """
    from erpnext.stock.utils import get_stock_balance as erp_stock_balance

    posting_date = posting_date or frappe.utils.today()
    return flt(erp_stock_balance(item_code, warehouse, posting_date))


def get_tank_to_product_map() -> dict:
    """
    Returns a dict mapping warehouse name → (item_code, item_name)
    for all active tank-type warehouses.
    
    Reads from Forecourt Pump nozzle definitions — a nozzle links a 
    tank (Warehouse) to a fuel product (Item). One tank can only store
    one product, so we deduplicate by tank.
    """
    rows = frappe.db.sql(
        """
        SELECT DISTINCT
            pn.tank      AS warehouse,
            pn.fuel_product AS item_code,
            i.item_name
        FROM `tabPump Nozzle` pn
        JOIN `tabItem` i ON i.name = pn.fuel_product
        JOIN `tabForecourt Pump` fp ON fp.name = pn.parent
        WHERE fp.is_active = 1
          AND pn.is_active = 1
        """,
        as_dict=True,
    )
    return {r.warehouse: {"item_code": r.item_code, "item_name": r.item_name} for r in rows}
```

### 16.2 Posting a Stock Reconciliation for Wetstock Adjustment

When a wetstock variance exceeds tolerance and requires a physical stock adjustment (not just an expense journal), use `Stock Reconciliation`.

```python
# forecourt/utils/stock.py  (continued)

def post_wetstock_stock_reconciliation(
    shift_name: str,
    tank: str,
    item_code: str,
    actual_qty: float,
    posting_date: str,
    posting_time: str,
    company: str,
) -> str:
    """
    Create and submit a Stock Reconciliation to align ERPNext inventory
    with the dip-measured actual closing stock.

    Called only when wetstock variance is Critical (> 0.5%) and the
    manager has approved the adjustment.

    Args:
        shift_name: Forecourt Shift reference, stored in remarks
        tank: warehouse name
        item_code: fuel item code
        actual_qty: closing dip volume in litres
        posting_date: shift close date
        posting_time: shift close time
        company: Anika Global Limited

    Returns:
        str: the Stock Reconciliation name
    """
    wac = get_current_wac(item_code, tank)

    sr = frappe.new_doc("Stock Reconciliation")
    sr.purpose = "Stock Reconciliation"
    sr.posting_date = posting_date
    sr.posting_time = posting_time
    sr.company = company
    sr.set_posting_time = 1
    sr.remarks = f"Wetstock adjustment — Shift {shift_name}"

    sr.append(
        "items",
        {
            "item_code": item_code,
            "warehouse": tank,
            "qty": actual_qty,
            "valuation_rate": wac,
        },
    )

    sr.insert(ignore_permissions=True)
    sr.submit()

    frappe.logger().info(
        f"Wetstock Stock Reconciliation {sr.name} submitted for shift {shift_name}, "
        f"tank {tank}, actual_qty={actual_qty} L @ WAC {wac}"
    )
    return sr.name
```

---

## 17. Accounts / Finance Module Integration

### 17.1 Account Lookup Helpers

```python
# forecourt/utils/accounts.py

import frappe
from frappe.utils import flt, nowdate, nowtime


# ---------------------------------------------------------------------------
# Account name helpers — centralise all hardcoded account names here so a
# single-tenant account rename does not break the reconciliation engine.
# ---------------------------------------------------------------------------

def get_account(short_name: str, company: str) -> str:
    """
    Resolve a logical account short name to the full ERPNext account name
    (which includes the company abbreviation suffix).

    Reads from a custom Site Preferences doctype. Falls back to a default
    map so the system works before Site Preferences is configured.

    Args:
        short_name: e.g. "till_active", "safe_main", "fuel_sales_pms_unl"
        company: company name

    Returns:
        str: Full account name, e.g. "Till — Active - SMSS"

    Raises:
        frappe.ValidationError if account is not found in ERPNext
    """
    _DEFAULTS = {
        "till_active":          "Till — Active",
        "safe_main":            "Safe — Main",
        "mpesa_clearing":       "MPesa Clearing",
        "card_clearing":        "Card Payment Clearing",
        "fleet_card_clearing":  "Fleet Card Clearing",
        "fuel_sales_pms_unl":   "Fuel Sales — PMS Unleaded",
        "fuel_sales_pms_vp":    "Fuel Sales — PMS V-Power",
        "fuel_sales_ago":       "Fuel Sales — AGO",
        "fuel_sales_dpk":       "Fuel Sales — DPK",
        "cogs_pms_unl":         "COGS — Fuel PMS Unleaded",
        "cogs_pms_vp":          "COGS — Fuel PMS V-Power",
        "cogs_ago":             "COGS — Fuel AGO",
        "cogs_dpk":             "COGS — Fuel DPK",
        "fuel_inv_pms_unl":     "Fuel Inventory — PMS Unleaded",
        "fuel_inv_pms_vp":      "Fuel Inventory — PMS V-Power",
        "fuel_inv_ago":         "Fuel Inventory — AGO",
        "fuel_inv_dpk":         "Fuel Inventory — DPK",
        "wetstock_var_pms":     "Wetstock Variance — PMS",
        "wetstock_var_ago":     "Wetstock Variance — AGO",
        "wetstock_var_dpk":     "Wetstock Variance — DPK",
        "cash_short_over":      "Cash Short / Over",
        "drive_off_losses":     "Drive-Off Losses",
        "accounts_payable":     "Creditors",
    }

    # Try Site Preferences first (allows per-station override)
    try:
        prefs = frappe.get_doc("Forecourt Site Preferences", company)
        acct_name = prefs.get(short_name) or _DEFAULTS.get(short_name)
    except frappe.DoesNotExistError:
        acct_name = _DEFAULTS.get(short_name)

    if not acct_name:
        frappe.throw(f"No account mapping found for '{short_name}'. "
                     "Configure Forecourt Site Preferences.")

    # Verify account exists in ERPNext
    abbr = frappe.get_cached_value("Company", company, "abbr")
    full_name = f"{acct_name} - {abbr}"
    if not frappe.db.exists("Account", full_name):
        frappe.throw(
            f"Account '{full_name}' does not exist in ERPNext. "
            "Create it in Chart of Accounts or update Site Preferences."
        )
    return full_name


# Map fuel item_code → (sales_account_key, cogs_account_key, inventory_account_key,
#                        wetstock_account_key)
FUEL_ACCOUNT_KEYS = {
    "FUEL-PMS-UNL": ("fuel_sales_pms_unl", "cogs_pms_unl", "fuel_inv_pms_unl", "wetstock_var_pms"),
    "FUEL-PMS-VP":  ("fuel_sales_pms_vp",  "cogs_pms_vp",  "fuel_inv_pms_vp",  "wetstock_var_pms"),
    "FUEL-AGO":     ("fuel_sales_ago",      "cogs_ago",     "fuel_inv_ago",     "wetstock_var_ago"),
    "FUEL-DPK":     ("fuel_sales_dpk",      "cogs_dpk",     "fuel_inv_dpk",     "wetstock_var_dpk"),
}


def get_fuel_accounts(item_code: str, company: str) -> dict:
    """
    Returns resolved account names for a fuel item.

    Returns:
        dict with keys: sales, cogs, inventory, wetstock_variance
    """
    if item_code not in FUEL_ACCOUNT_KEYS:
        frappe.throw(f"Unknown fuel item code: {item_code}. "
                     "Add it to FUEL_ACCOUNT_KEYS in accounts.py")
    sales_k, cogs_k, inv_k, ws_k = FUEL_ACCOUNT_KEYS[item_code]
    return {
        "sales":              get_account(sales_k, company),
        "cogs":               get_account(cogs_k,  company),
        "inventory":          get_account(inv_k,   company),
        "wetstock_variance":  get_account(ws_k,    company),
    }
```

---

## 18. HR Module Integration

### 18.1 Employee Validation Utilities

```python
# forecourt/utils/hr.py

import frappe


def get_employee_name(employee: str) -> str:
    """Return the employee_name for display in error messages."""
    return frappe.get_cached_value("Employee", employee, "employee_name") or employee


def assert_employees_differ(cashier: str, supervisor: str, context: str = ""):
    """
    Hard-block if cashier and supervisor are the same employee.
    Enforces Rule 7: dual control on all accountability events.
    
    Args:
        cashier: Employee link value
        supervisor: Employee link value
        context: human-readable context for the error message
    
    Raises:
        frappe.ValidationError
    """
    if cashier and supervisor and cashier == supervisor:
        frappe.throw(
            f"The same employee ({get_employee_name(cashier)}) cannot be both "
            f"cashier and supervisor{' in ' + context if context else ''}. "
            "Assign a different supervisor.",
            title="Dual Control Violation",
        )


def assert_is_active_employee(employee: str, role_label: str = "Employee"):
    """
    Verify the employee record is active (status = 'Active').
    Prevents assigning terminated staff to a shift.
    """
    status = frappe.get_cached_value("Employee", employee, "status")
    if status != "Active":
        frappe.throw(
            f"{role_label} {get_employee_name(employee)} is not Active "
            f"(current status: {status}). Only active employees can be assigned to shifts.",
            title="Inactive Employee",
        )


def get_shift_cashier(shift_name: str) -> str:
    """
    Return the cashier Employee linked to a Forecourt Shift.
    Used by Cash Event and Cashier Session validators to cross-check
    the authorised_by field.
    """
    return frappe.get_cached_value("Forecourt Shift", shift_name, "cashier")
```

---

## 19. Buying / Procurement Module Integration

### 19.1 Creating a Purchase Receipt After a Dip-Confirmed Delivery

```python
# forecourt/utils/buying.py

import frappe
from frappe.utils import flt, now_datetime


def create_purchase_receipt_for_delivery(
    fuel_delivery_dip_name: str,
    company: str,
    cost_center: str,
) -> str:
    """
    Creates and submits an ERPNext Purchase Receipt for a dip-confirmed
    fuel delivery. Called when a Fuel Delivery Dip is set to 'Accepted'.

    The Purchase Receipt quantity uses dip_measured_l (not docket_volume_l)
    when the variance exceeds 0.5%. This protects against short deliveries.

    Args:
        fuel_delivery_dip_name: Fuel Delivery Dip document name
        company: Anika Global Limited
        cost_center: Main Cost Center - SMSS

    Returns:
        str: Purchase Receipt name

    Raises:
        frappe.ValidationError if the delivery is still Pending or Disputed
    """
    fdd = frappe.get_doc("Fuel Delivery Dip", fuel_delivery_dip_name)

    if fdd.status != "Accepted":
        frappe.throw(
            f"Fuel Delivery Dip {fuel_delivery_dip_name} is {fdd.status}. "
            "Only Accepted deliveries can have a Purchase Receipt created.",
        )

    if fdd.purchase_receipt:
        frappe.throw(
            f"Purchase Receipt {fdd.purchase_receipt} already exists for "
            f"Fuel Delivery Dip {fuel_delivery_dip_name}.",
        )

    # Decide which volume to receipt: dip-measured if variance > 0.5%
    variance_pct = flt(fdd.delivery_variance_pct)
    receipt_qty = (
        flt(fdd.dip_measured_l)
        if abs(variance_pct) > 0.5
        else flt(fdd.docket_volume_l)
    )

    # Get supplier from the shift (supplier field must be on Forecourt Shift
    # or we read from a Site Preferences default)
    supplier = frappe.get_cached_value(
        "Forecourt Shift", fdd.shift, "fuel_supplier"
    ) or frappe.get_cached_value("Forecourt Site Preferences", company, "default_fuel_supplier")

    if not supplier:
        frappe.throw(
            "No fuel supplier found on the Forecourt Shift or Site Preferences. "
            "Set 'fuel_supplier' on the shift before creating a Purchase Receipt."
        )

    # Get item buying rate from most recent Purchase Order / Supplier Quotation
    # or fall back to Item.last_purchase_rate
    buying_rate = flt(
        frappe.get_cached_value("Item", fdd.fuel_product, "last_purchase_rate")
    )

    pr = frappe.new_doc("Purchase Receipt")
    pr.company = company
    pr.supplier = supplier
    pr.posting_date = frappe.utils.today()
    pr.set_posting_time = 1
    pr.remarks = (
        f"Fuel delivery — Shift {fdd.shift} | Docket {fdd.docket_number} | "
        f"Truck {fdd.truck_reg} | Driver {fdd.driver_name or 'N/A'} | "
        f"FDD {fuel_delivery_dip_name}"
    )
    pr.append(
        "items",
        {
            "item_code": fdd.fuel_product,
            "item_name": frappe.get_cached_value("Item", fdd.fuel_product, "item_name"),
            "qty": receipt_qty,
            "uom": "Litre",
            "stock_uom": "Litre",
            "conversion_factor": 1.0,
            "rate": buying_rate,
            "warehouse": fdd.tank,
            "cost_center": cost_center,
        },
    )

    pr.insert(ignore_permissions=True)
    pr.submit()

    # Write back to Fuel Delivery Dip
    frappe.db.set_value(
        "Fuel Delivery Dip",
        fuel_delivery_dip_name,
        "purchase_receipt",
        pr.name,
    )

    frappe.logger().info(
        f"Purchase Receipt {pr.name} submitted for FDD {fuel_delivery_dip_name}, "
        f"qty={receipt_qty} L, supplier={supplier}"
    )
    return pr.name
```

---

## 20. Selling / POS Module Integration

### 20.1 Aggregating POS Invoice Totals for Reconciliation

```python
# forecourt/utils/pos.py

import frappe
from frappe.utils import flt


def get_shift_pos_totals(shift_name: str) -> dict:
    """
    Aggregate submitted POS Invoice totals for a shift, broken down by
    payment method and fuel product.

    A POS Invoice is linked to a shift via a custom 'forecourt_shift' field
    that must be added to POS Invoice (via Custom Field).

    Returns:
        dict with structure:
        {
            "by_payment": {
                "Cash": float,
                "MPesa Clearing": float,
                "Card Payment Clearing": float,
                "Fleet Card Clearing": float,
            },
            "by_product": {
                "FUEL-PMS-UNL": {"qty": float, "amount": float},
                "FUEL-PMS-VP":  {"qty": float, "amount": float},
                ...
            },
            "grand_total": float,
            "invoice_count": int,
            "void_count": int,
        }
    """
    # Payment method totals
    payment_rows = frappe.db.sql(
        """
        SELECT
            mop.mode_of_payment,
            SUM(mop.amount) AS total
        FROM `tabPOS Invoice` pi
        JOIN `tabSales Invoice Payment` mop ON mop.parent = pi.name
        WHERE pi.forecourt_shift = %(shift)s
          AND pi.docstatus = 1
        GROUP BY mop.mode_of_payment
        """,
        {"shift": shift_name},
        as_dict=True,
    )

    by_payment = {r.mode_of_payment: flt(r.total) for r in payment_rows}

    # Product totals
    product_rows = frappe.db.sql(
        """
        SELECT
            pii.item_code,
            SUM(pii.qty)       AS qty,
            SUM(pii.net_amount) AS amount
        FROM `tabPOS Invoice` pi
        JOIN `tabPOS Invoice Item` pii ON pii.parent = pi.name
        WHERE pi.forecourt_shift = %(shift)s
          AND pi.docstatus = 1
        GROUP BY pii.item_code
        """,
        {"shift": shift_name},
        as_dict=True,
    )

    by_product = {
        r.item_code: {"qty": flt(r.qty), "amount": flt(r.amount)}
        for r in product_rows
    }

    # Grand total and counts
    summary = frappe.db.sql(
        """
        SELECT
            COUNT(*)            AS invoice_count,
            SUM(grand_total)    AS grand_total
        FROM `tabPOS Invoice`
        WHERE forecourt_shift = %(shift)s
          AND docstatus = 1
        """,
        {"shift": shift_name},
        as_dict=True,
    )[0]

    void_count = frappe.db.count(
        "POS Invoice",
        {"forecourt_shift": shift_name, "docstatus": 2},
    )

    return {
        "by_payment":    by_payment,
        "by_product":    by_product,
        "grand_total":   flt(summary.grand_total),
        "invoice_count": summary.invoice_count or 0,
        "void_count":    void_count,
    }


def get_shift_sales_invoice_totals(shift_name: str) -> dict:
    """
    Aggregate submitted Sales Invoice totals for fleet/credit customers
    linked to a shift. Sales Invoices generate AR entries — they are
    never included in the cash till calculation.

    Returns:
        dict: {item_code: {"qty": float, "amount": float}}, plus "grand_total"
    """
    rows = frappe.db.sql(
        """
        SELECT
            sii.item_code,
            SUM(sii.qty)        AS qty,
            SUM(sii.net_amount) AS amount
        FROM `tabSales Invoice` si
        JOIN `tabSales Invoice Item` sii ON sii.parent = si.name
        WHERE si.forecourt_shift = %(shift)s
          AND si.docstatus = 1
        GROUP BY sii.item_code
        """,
        {"shift": shift_name},
        as_dict=True,
    )

    by_product = {
        r.item_code: {"qty": flt(r.qty), "amount": flt(r.amount)}
        for r in rows
    }

    grand_total = frappe.db.sql(
        """
        SELECT COALESCE(SUM(grand_total), 0) AS total
        FROM `tabSales Invoice`
        WHERE forecourt_shift = %(shift)s AND docstatus = 1
        """,
        {"shift": shift_name},
    )[0][0]

    return {"by_product": by_product, "grand_total": flt(grand_total)}
```

---

## 21. Python — DocType Validation Hooks

### 21.1 Forecourt Pump Validation

```python
# forecourt/forecourt/doctype/forecourt_pump/forecourt_pump.py

import frappe
from frappe.model.document import Document
from frappe.utils import today, add_days, date_diff


class ForecourtPump(Document):

    def validate(self):
        self._validate_unique_tank_per_nozzle()
        self._warn_calibration_due()
        self._set_nozzle_inactive_on_pump_deactivate()

    def _validate_unique_tank_per_nozzle(self):
        """
        No two active nozzles on the same pump may link to the same tank.
        E.g. Pump 3 cannot have both Nozzle 1 and Nozzle 2 pointing to Tank 1.
        """
        active_nozzles = [n for n in self.nozzles if n.is_active]
        tanks_seen = {}
        for nozzle in active_nozzles:
            if nozzle.tank in tanks_seen:
                frappe.throw(
                    f"Pump {self.pump_number}: Nozzle {nozzle.nozzle_number} and "
                    f"Nozzle {tanks_seen[nozzle.tank]} both point to tank {nozzle.tank}. "
                    "Each active nozzle on a pump must link to a different tank.",
                    title="Duplicate Tank Assignment",
                )
            tanks_seen[nozzle.tank] = nozzle.nozzle_number

    def _warn_calibration_due(self):
        """Show a dashboard warning if calibration is due within 30 days."""
        if self.next_calibration_due:
            days_remaining = date_diff(self.next_calibration_due, today())
            if days_remaining <= 0:
                frappe.msgprint(
                    f"Pump {self.pump_number} calibration is OVERDUE "
                    f"(due {self.next_calibration_due}). Contact KEBS / Weights & Measures.",
                    indicator="red",
                    title="Calibration Overdue",
                )
            elif days_remaining <= 30:
                frappe.msgprint(
                    f"Pump {self.pump_number} calibration is due in {days_remaining} day(s) "
                    f"({self.next_calibration_due}). Schedule a calibration visit.",
                    indicator="orange",
                    title="Calibration Due Soon",
                )

    def _set_nozzle_inactive_on_pump_deactivate(self):
        """
        When a pump is set inactive (is_active = No), mark all its
        nozzles inactive automatically so they are excluded from shift readings.
        """
        if not self.is_active:
            for nozzle in self.nozzles:
                if nozzle.is_active:
                    nozzle.is_active = 0
                    frappe.msgprint(
                        f"Nozzle {nozzle.nozzle_number} on Pump {self.pump_number} "
                        "has been automatically deactivated because the pump is inactive.",
                        indicator="blue",
                    )
```

### 21.2 Meter Reading Validation

```python
# forecourt/forecourt/doctype/meter_reading/meter_reading.py

import frappe
from frappe.model.document import Document
from frappe.utils import flt


class MeterReading(Document):

    def validate(self):
        self._set_unit_from_meter_type()
        self._validate_totalizer_positive()
        self._validate_closing_ge_opening()
        self._validate_amendment_fields()
        self._enforce_immutability_on_edit()

    def _set_unit_from_meter_type(self):
        """Auto-set the unit field based on meter_type."""
        unit_map = {
            "Electronic Volume": "Litres",
            "Manual Mechanical": "Litres",
            "Electronic Cash":   "KES",
        }
        self.unit = unit_map.get(self.meter_type, "")

    def _validate_totalizer_positive(self):
        if flt(self.totalizer_value) <= 0:
            frappe.throw(
                f"Totalizer value must be greater than zero. "
                f"Got: {self.totalizer_value}",
                title="Invalid Meter Reading",
            )

    def _validate_closing_ge_opening(self):
        """
        For Shift Close readings: the totalizer must be ≥ the Shift Open reading
        for the same nozzle, pump, and meter type. Meters only count forward.
        """
        if self.reading_position != "Shift Close":
            return

        opening = frappe.db.get_value(
            "Meter Reading",
            {
                "shift":            self.shift,
                "pump":             self.pump,
                "nozzle_number":    self.nozzle_number,
                "meter_type":       self.meter_type,
                "reading_position": "Shift Open",
                "docstatus":        1,
            },
            "totalizer_value",
        )

        if opening is None:
            frappe.throw(
                f"No submitted Shift Open reading found for Pump {self.pump}, "
                f"Nozzle {self.nozzle_number}, Meter Type {self.meter_type}. "
                "Create and submit the opening reading first.",
                title="Missing Opening Reading",
            )

        if flt(self.totalizer_value) < flt(opening):
            frappe.throw(
                f"Closing totalizer ({self.totalizer_value}) is less than "
                f"opening totalizer ({opening}) for Pump {self.pump}, "
                f"Nozzle {self.nozzle_number}, {self.meter_type}. "
                "Meters only count forward. Check for a transcription error.",
                title="Invalid Closing Reading",
            )

    def _validate_amendment_fields(self):
        """Amendment readings must have a reason and reference the superseded reading."""
        if self.reading_position != "Amendment":
            return
        if not self.amendment_reason:
            frappe.throw(
                "Amendment readings require an Amendment Reason.",
                title="Amendment Reason Required",
            )
        if not self.superseded_by:
            frappe.throw(
                "Amendment readings must reference the original reading in 'Superseded By'.",
                title="Superseded Reading Required",
            )

    def _enforce_immutability_on_edit(self):
        """
        Once submitted (docstatus = 1), totalizer_value must not change.
        If it has changed, throw. This protects the audit trail.
        """
        if self.docstatus != 1:
            return  # Only enforce on submitted docs

        old_value = frappe.db.get_value("Meter Reading", self.name, "totalizer_value")
        if old_value is not None and flt(old_value) != flt(self.totalizer_value):
            frappe.throw(
                f"Submitted Meter Reading {self.name} is immutable. "
                f"Create an Amendment reading instead of editing this one.",
                title="Immutable Reading",
            )
```

### 21.3 Dip Reading Validation

```python
# forecourt/forecourt/doctype/dip_reading/dip_reading.py

import frappe
from frappe.model.document import Document
from frappe.utils import flt, time_diff_in_seconds


class DipReading(Document):

    WATER_ALERT_MM = 20.0
    MIN_SETTLE_MINUTES = 10

    def validate(self):
        self._validate_volume_within_capacity()
        self._check_water_level()
        self._validate_delivery_after_settle_time()
        self._validate_shift_close_has_open()

    def _validate_volume_within_capacity(self):
        """Volume must not exceed the tank's configured capacity."""
        capacity = frappe.get_cached_value("Warehouse", self.tank, "capacity_in_litres")
        if capacity and flt(self.volume_observed_l) > flt(capacity):
            frappe.throw(
                f"Observed volume {self.volume_observed_l} L exceeds tank capacity "
                f"{capacity} L for {self.tank}. "
                "Check dipstick reading and calibration chart.",
                title="Volume Exceeds Capacity",
            )
        if flt(self.volume_observed_l) <= 0:
            frappe.throw(
                "Observed volume must be greater than zero.",
                title="Invalid Dip Volume",
            )

    def _check_water_level(self):
        """Alert if water bottom level exceeds threshold."""
        if flt(self.water_level_mm) > self.WATER_ALERT_MM:
            frappe.msgprint(
                f"HIGH WATER LEVEL: {self.water_level_mm} mm in {self.tank} "
                f"(threshold: {self.WATER_ALERT_MM} mm). "
                "Possible water contamination — alert manager immediately and "
                "suspend dispensing from this tank until tested.",
                indicator="red",
                title="Water Contamination Alert",
                raise_exception=False,
            )

    def _validate_delivery_after_settle_time(self):
        """
        Delivery After dip must be ≥ MIN_SETTLE_MINUTES after delivery end.
        """
        if self.reading_type != "Delivery After":
            return

        # Find the linked Fuel Delivery Dip via the dip_after link
        fdd_name = frappe.db.get_value(
            "Fuel Delivery Dip",
            {"dip_after": self.name},
            "name",
        )
        if not fdd_name:
            return  # Not yet linked — skip (link may be set after creation)

        delivery_end = frappe.db.get_value("Fuel Delivery Dip", fdd_name, "delivery_end")
        if delivery_end and self.observed_at:
            diff_seconds = time_diff_in_seconds(self.observed_at, str(delivery_end))
            if diff_seconds < self.MIN_SETTLE_MINUTES * 60:
                frappe.throw(
                    f"Delivery After dip must be taken at least {self.MIN_SETTLE_MINUTES} "
                    f"minutes after delivery end ({delivery_end}). "
                    f"Current gap: {int(diff_seconds / 60)} min {int(diff_seconds % 60)} sec. "
                    "Wait for the fuel to settle before dipping.",
                    title="Insufficient Settle Time",
                )

    def _validate_shift_close_has_open(self):
        """
        A Shift Close dip requires a submitted Shift Open dip for the same
        tank in the same shift.
        """
        if self.reading_type != "Shift Close":
            return
        if not self.shift:
            return

        exists = frappe.db.exists(
            "Dip Reading",
            {
                "shift":        self.shift,
                "tank":         self.tank,
                "reading_type": "Shift Open",
                "docstatus":    1,
            },
        )
        if not exists:
            frappe.throw(
                f"No submitted Shift Open dip reading found for tank {self.tank} "
                f"in shift {self.shift}. Create and submit the opening dip first.",
                title="Missing Opening Dip",
            )
```

### 21.4 Cash Event Validation

```python
# forecourt/forecourt/doctype/cash_event/cash_event.py

import frappe
from frappe.model.document import Document
from frappe.utils import flt


class CashEvent(Document):

    def validate(self):
        self._validate_amount_positive()
        self._validate_dual_control()
        self._validate_one_float_per_session()
        self._require_reference_for_pickup_and_drop()

    def _validate_amount_positive(self):
        if flt(self.amount) <= 0:
            frappe.throw(
                "Cash Event amount must be positive and non-zero.",
                title="Invalid Amount",
            )

    def _validate_dual_control(self):
        """
        authorised_by must differ from the session cashier.
        Enforces Rule 7: no cashier can authorise their own cash movements.
        """
        from forecourt.utils.hr import assert_employees_differ

        session_cashier = frappe.get_cached_value(
            "Cashier Session", self.cashier_session, "cashier"
        )
        if session_cashier:
            assert_employees_differ(
                session_cashier,
                self.authorised_by,
                context=f"Cash Event {self.event_type}",
            )

    def _validate_one_float_per_session(self):
        """Only one Float Issued event is permitted per Cashier Session."""
        if self.event_type != "Float Issued":
            return
        existing = frappe.db.exists(
            "Cash Event",
            {
                "cashier_session": self.cashier_session,
                "event_type":      "Float Issued",
                "name":            ("!=", self.name),
                "docstatus":       ("!=", 2),
            },
        )
        if existing:
            frappe.throw(
                f"A Float Issued event already exists for session {self.cashier_session}. "
                "Only one float may be issued per cashier session.",
                title="Duplicate Float",
            )

    def _require_reference_for_pickup_and_drop(self):
        """Cash Pickup and Safe Drop require a physical envelope reference."""
        if self.event_type in ("Cash Pickup", "Safe Drop") and not self.reference:
            frappe.throw(
                f"{self.event_type} requires an envelope or reference number. "
                "Enter the physical envelope number in the Reference field.",
                title="Reference Required",
            )
```

---

## 22. Python — Shift Status Transition Enforcement

```python
# forecourt/forecourt/doctype/forecourt_shift/forecourt_shift.py

import frappe
from frappe.model.document import Document
from frappe.utils import now_datetime, flt

from forecourt.utils.hr import (
    assert_employees_differ,
    assert_is_active_employee,
)


# Valid one-directional transitions (from → to).
# Disputed can return to Closing once manager clears.
ALLOWED_TRANSITIONS = {
    "Draft":              ["Open"],
    "Open":               ["Readings Captured", "Disputed"],
    "Readings Captured":  ["Closing",           "Disputed"],
    "Closing":            ["Closed",             "Disputed"],
    "Closed":             [],                    # Terminal state
    "Disputed":           ["Closing"],            # Manager re-opens investigation
}


class ForecourtShift(Document):

    def validate(self):
        self._validate_cashier_ne_supervisor()
        self._validate_only_one_open_shift()
        self._validate_rates_locked_before_open()
        self._validate_closed_at_after_opened_at()

    def before_save(self):
        self._enforce_status_transition()

    # ------------------------------------------------------------------
    # Validators
    # ------------------------------------------------------------------

    def _validate_cashier_ne_supervisor(self):
        assert_employees_differ(
            self.cashier, self.supervisor, context="Forecourt Shift"
        )
        if self.cashier:
            assert_is_active_employee(self.cashier, role_label="Cashier")
        if self.supervisor:
            assert_is_active_employee(self.supervisor, role_label="Supervisor")

    def _validate_only_one_open_shift(self):
        """
        Only one shift per station may be in Open or Closing status at a time.
        Prevents two simultaneous open shifts creating duplicate meter deltas.
        """
        if self.status not in ("Open", "Closing"):
            return

        conflict = frappe.db.get_value(
            "Forecourt Shift",
            {
                "station": self.station,
                "status":  ("in", ["Open", "Closing"]),
                "name":    ("!=", self.name),
            },
            "name",
        )
        if conflict:
            frappe.throw(
                f"Shift {conflict} is already Open or Closing at {self.station}. "
                "Close or dispute that shift before opening a new one.",
                title="Concurrent Shift Conflict",
            )

    def _validate_rates_locked_before_open(self):
        """EPRA rates must be set before status moves to Open."""
        if self.status not in ("Open", "Readings Captured", "Closing", "Closed"):
            return
        # At minimum, the station must sell one product — PMS Unleaded
        if not flt(self.rate_pms_unl):
            frappe.throw(
                "EPRA Rate for PMS Unleaded must be set and locked before "
                "the shift can be opened.",
                title="EPRA Rates Not Set",
            )

    def _validate_closed_at_after_opened_at(self):
        if self.closed_at and self.opened_at:
            if self.closed_at <= self.opened_at:
                frappe.throw(
                    f"Closed At ({self.closed_at}) must be after "
                    f"Opened At ({self.opened_at}).",
                    title="Invalid Shift Timing",
                )

    # ------------------------------------------------------------------
    # Status transition enforcement
    # ------------------------------------------------------------------

    def _enforce_status_transition(self):
        """
        Block invalid status transitions. Read the previous status from the
        database to compare against the new status in self.status.
        """
        if self.is_new():
            # New document — only Draft or Open allowed at creation
            if self.status not in ("Draft", "Open"):
                frappe.throw(
                    f"A new shift must start with status Draft or Open, "
                    f"not '{self.status}'.",
                    title="Invalid Initial Status",
                )
            return

        old_status = frappe.db.get_value("Forecourt Shift", self.name, "status")
        if old_status == self.status:
            return  # No change — nothing to validate

        allowed = ALLOWED_TRANSITIONS.get(old_status, [])
        if self.status not in allowed:
            frappe.throw(
                f"Cannot transition Forecourt Shift from '{old_status}' to "
                f"'{self.status}'. Allowed transitions from {old_status}: "
                f"{allowed or 'none (terminal state)'}.",
                title="Invalid Status Transition",
            )

        # Extra checks for specific transitions
        if self.status == "Readings Captured":
            self._assert_all_opening_readings_present()

        if self.status == "Closing":
            self._assert_closing_readings_present()

        if self.status == "Closed":
            self._assert_gl_posted()

    def _assert_all_opening_readings_present(self):
        """
        Moving to Readings Captured requires all 3 meter types × all active
        nozzles to have submitted Shift Open readings.
        """
        active_nozzles = self._get_active_nozzles()
        meter_types = ["Electronic Volume", "Electronic Cash", "Manual Mechanical"]
        missing = []

        for pump_name, nozzle_number in active_nozzles:
            for mt in meter_types:
                exists = frappe.db.exists(
                    "Meter Reading",
                    {
                        "shift":            self.name,
                        "pump":             pump_name,
                        "nozzle_number":    nozzle_number,
                        "meter_type":       mt,
                        "reading_position": "Shift Open",
                        "docstatus":        1,
                    },
                )
                if not exists:
                    missing.append(
                        f"Pump {pump_name} Nozzle {nozzle_number} — {mt}"
                    )

        if missing:
            frappe.throw(
                "Cannot move to 'Readings Captured'. Missing Shift Open readings:\n"
                + "\n".join(f"  • {m}" for m in missing),
                title="Incomplete Opening Readings",
            )

    def _assert_closing_readings_present(self):
        """
        Moving to Closing requires all 3 meter types × all active nozzles to
        have submitted Shift Close readings, plus one closing dip per active tank.
        """
        active_nozzles = self._get_active_nozzles()
        meter_types = ["Electronic Volume", "Electronic Cash", "Manual Mechanical"]
        missing = []

        for pump_name, nozzle_number in active_nozzles:
            for mt in meter_types:
                exists = frappe.db.exists(
                    "Meter Reading",
                    {
                        "shift":            self.name,
                        "pump":             pump_name,
                        "nozzle_number":    nozzle_number,
                        "meter_type":       mt,
                        "reading_position": "Shift Close",
                        "docstatus":        1,
                    },
                )
                if not exists:
                    missing.append(
                        f"Closing Meter — Pump {pump_name} Nozzle {nozzle_number} — {mt}"
                    )

        active_tanks = self._get_active_tanks()
        for tank in active_tanks:
            exists = frappe.db.exists(
                "Dip Reading",
                {
                    "shift":        self.name,
                    "tank":         tank,
                    "reading_type": "Shift Close",
                    "docstatus":    1,
                },
            )
            if not exists:
                missing.append(f"Closing Dip — {tank}")

        if missing:
            frappe.throw(
                "Cannot move to 'Closing'. Missing closing records:\n"
                + "\n".join(f"  • {m}" for m in missing),
                title="Incomplete Closing Records",
            )

    def _assert_gl_posted(self):
        """GL Journal Entry must be posted before shift can close."""
        if not self.gl_journal:
            frappe.throw(
                "Cannot close shift: GL Journal Entry has not been posted. "
                "Run 'Approve and Post GL' from the Shift Reconciliation.",
                title="GL Not Posted",
            )

    # ------------------------------------------------------------------
    # Helpers
    # ------------------------------------------------------------------

    def _get_active_nozzles(self) -> list:
        """Return list of (pump_name, nozzle_number) for all active nozzles."""
        return frappe.db.sql(
            """
            SELECT fp.name AS pump, pn.nozzle_number
            FROM `tabForecourt Pump` fp
            JOIN `tabPump Nozzle` pn ON pn.parent = fp.name
            WHERE fp.is_active = 1 AND pn.is_active = 1
            ORDER BY fp.pump_number, pn.nozzle_number
            """,
            as_list=True,
        )

    def _get_active_tanks(self) -> list:
        """Return list of warehouse names (tanks) used by active nozzles."""
        rows = frappe.db.sql(
            """
            SELECT DISTINCT pn.tank
            FROM `tabForecourt Pump` fp
            JOIN `tabPump Nozzle` pn ON pn.parent = fp.name
            WHERE fp.is_active = 1 AND pn.is_active = 1
            """,
            as_list=True,
        )
        return [r[0] for r in rows]
```

---

## 23. Python — Reconciliation Computation Script

```python
# forecourt/forecourt/doctype/shift_reconciliation/reconciliation_engine.py

"""
Reconciliation Engine — Computes the Shift Reconciliation record.

Called by the Forecourt Shift "Compute Reconciliation" button action.
Never call directly from a DocType validate/on_submit — this is a 
manager-triggered computation.
"""

import frappe
from frappe.utils import flt, now_datetime

from forecourt.utils.accounts import get_fuel_accounts, FUEL_ACCOUNT_KEYS
from forecourt.utils.stock import get_current_wac
from forecourt.utils.pos import get_shift_pos_totals, get_shift_sales_invoice_totals
from forecourt.utils.meter import compute_meter_validation_results


# Variance thresholds (percent or KES) — override in Forecourt Site Preferences
WETSTOCK_NORMAL_PCT   = 0.30
WETSTOCK_ELEVATED_PCT = 0.50
CASH_NORMAL_KES       = 50.0
CASH_ELEVATED_KES     = 200.0


def compute_reconciliation(shift_name: str) -> str:
    """
    Compute and save the Shift Reconciliation for a given shift.

    Steps:
    1.  Run meter validation for all nozzles → Meter Validation Results
    2.  Compute per-product volume and revenue summaries
    3.  Compute per-tank wetstock summaries
    4.  Compute cash reconciliation
    5.  Aggregate non-cash tenders
    6.  Set status flags and verdict
    7.  Save Shift Reconciliation

    Args:
        shift_name: Forecourt Shift document name

    Returns:
        str: Shift Reconciliation document name

    Raises:
        frappe.ValidationError if any meter validation fails (unfixed)
    """
    shift = frappe.get_doc("Forecourt Shift", shift_name)

    # ------------------------------------------------------------------
    # Step 1: Meter Validation
    # ------------------------------------------------------------------
    meter_results = compute_meter_validation_results(shift_name)
    failed = [r for r in meter_results if r["overall_status"] == "Fail"]
    if failed:
        failed_list = "\n".join(
            f"  • Pump {r['pump']} Nozzle {r['nozzle_number']}: "
            f"Check A={r['check_a_status']}, Check B={r['check_b_status']}"
            for r in failed
        )
        frappe.throw(
            "Meter validation failures must be resolved before computing "
            "the reconciliation:\n" + failed_list,
            title="Meter Validation Failures",
        )

    # ------------------------------------------------------------------
    # Step 2: Per-product volume and revenue
    # ------------------------------------------------------------------
    company = shift.company or frappe.defaults.get_global_default("company")
    product_summaries, nozzle_summaries = _compute_product_summaries(shift, company)

    # ------------------------------------------------------------------
    # Step 3: Wetstock per tank
    # ------------------------------------------------------------------
    tank_summaries = _compute_tank_wetstock(shift, product_summaries, company)

    # ------------------------------------------------------------------
    # Step 4: Cash reconciliation
    # ------------------------------------------------------------------
    cash_summary = _compute_cash_reconciliation(shift)

    # ------------------------------------------------------------------
    # Step 5: Non-cash tenders
    # ------------------------------------------------------------------
    pos_totals = get_shift_pos_totals(shift_name)
    si_totals  = get_shift_sales_invoice_totals(shift_name)

    mpesa_total       = flt(pos_totals["by_payment"].get("MPesa Clearing", 0))
    card_total        = flt(pos_totals["by_payment"].get("Card Payment Clearing", 0))
    fleet_credit_total = flt(si_totals["grand_total"])

    # ------------------------------------------------------------------
    # Step 6: Status flags and verdict
    # ------------------------------------------------------------------
    cash_variance_kes = flt(cash_summary["cash_variance"])
    abs_cv = abs(cash_variance_kes)

    if abs_cv <= CASH_NORMAL_KES:
        cash_status = "Balanced"
        is_cash_balanced = True
    elif abs_cv <= CASH_ELEVATED_KES:
        cash_status = "Elevated"
        is_cash_balanced = False
    else:
        cash_status = "Critical"
        is_cash_balanced = False

    all_wetstock_ok = all(
        t["classification"] in ("Normal",) for t in tank_summaries
    )
    any_critical_wetstock = any(
        t["classification"] == "Critical" for t in tank_summaries
    )

    any_meter_warning = any(
        r["overall_status"] == "Warning" for r in meter_results
    )
    all_meters_ok = not any(
        r["overall_status"] in ("Warning", "Fail") for r in meter_results
    )

    requires_approval = (
        not is_cash_balanced
        or not all_wetstock_ok
        or any_critical_wetstock
    )

    # ------------------------------------------------------------------
    # Step 7: Save Shift Reconciliation
    # ------------------------------------------------------------------
    # Delete existing if re-computing
    existing = frappe.db.get_value("Shift Reconciliation", {"shift": shift_name}, "name")
    if existing:
        frappe.delete_doc("Shift Reconciliation", existing, force=True)

    sr = frappe.new_doc("Shift Reconciliation")
    sr.shift            = shift_name
    sr.computed_at      = now_datetime()

    # Meter
    sr.all_meters_ok    = int(all_meters_ok)
    sr.meter_warnings   = int(any_meter_warning)

    # Volume / Revenue
    total_volume = sum(flt(p["volume_sold_l"]) for p in product_summaries)
    total_revenue = sum(flt(p["gross_revenue"]) for p in product_summaries)
    sr.total_volume_l       = total_volume
    sr.total_gross_revenue  = total_revenue

    for field_map in [
        ("vol_pms_unl", "FUEL-PMS-UNL"),
        ("vol_pms_vp",  "FUEL-PMS-VP"),
        ("vol_ago",     "FUEL-AGO"),
        ("vol_dpk",     "FUEL-DPK"),
    ]:
        field, item_code = field_map
        ps = next((p for p in product_summaries if p["item_code"] == item_code), None)
        setattr(sr, field, flt(ps["volume_sold_l"]) if ps else 0.0)

    # Cash
    sr.expected_cash        = cash_summary["expected_cash"]
    sr.actual_cash          = cash_summary["actual_cash"]
    sr.cash_variance        = cash_variance_kes
    sr.cash_variance_status = cash_status
    sr.mpesa_total          = mpesa_total
    sr.card_total           = card_total
    sr.fleet_credit_total   = fleet_credit_total

    # Wetstock aggregate
    total_ws_l   = sum(flt(t["variance_l"])   for t in tank_summaries)
    total_ws_kes = sum(flt(t["variance_kes"])  for t in tank_summaries)
    sr.wetstock_variance_l   = total_ws_l
    sr.wetstock_variance_kes = total_ws_kes

    # Outcome
    sr.is_cash_balanced      = int(is_cash_balanced)
    sr.is_wetstock_balanced  = int(all_wetstock_ok)
    sr.is_meter_ok           = int(all_meters_ok)
    sr.requires_approval     = int(requires_approval)

    # Child tables
    for p in product_summaries:
        sr.append("product_summaries", p)
    for t in tank_summaries:
        sr.append("tank_wetstock_summaries", t)
    for n in nozzle_summaries:
        sr.append("nozzle_meter_summary", n)

    sr.insert(ignore_permissions=True)

    frappe.logger().info(
        f"Shift Reconciliation {sr.name} computed for shift {shift_name}. "
        f"Cash: {cash_status} ({cash_variance_kes:+.2f} KES), "
        f"Wetstock: {'OK' if all_wetstock_ok else 'ELEVATED'}."
    )
    return sr.name


# ------------------------------------------------------------------
# Internal helpers
# ------------------------------------------------------------------

def _compute_product_summaries(shift, company: str) -> tuple:
    """
    Compute per-product volume sold, revenue, WAC, COGS, and margin.
    Also returns per-nozzle data for the child table.
    """
    # Get all meter deltas (Elec Vol) grouped by nozzle → product
    nozzle_rows = frappe.db.sql(
        """
        SELECT
            pn.fuel_product AS item_code,
            mr_close.pump,
            mr_close.nozzle_number,
            (mr_close.totalizer_value - mr_open.totalizer_value) AS elec_vol_sold,
            (mc_close.totalizer_value - mc_open.totalizer_value) AS elec_cash_sold,
            (mm_close.totalizer_value - mm_open.totalizer_value) AS mech_vol_sold
        FROM `tabMeter Reading` mr_close
        JOIN `tabMeter Reading` mr_open
            ON mr_open.shift = mr_close.shift
           AND mr_open.pump = mr_close.pump
           AND mr_open.nozzle_number = mr_close.nozzle_number
           AND mr_open.meter_type = 'Electronic Volume'
           AND mr_open.reading_position = 'Shift Open'
           AND mr_open.docstatus = 1
        JOIN `tabMeter Reading` mc_close
            ON mc_close.shift = mr_close.shift
           AND mc_close.pump = mr_close.pump
           AND mc_close.nozzle_number = mr_close.nozzle_number
           AND mc_close.meter_type = 'Electronic Cash'
           AND mc_close.reading_position = 'Shift Close'
           AND mc_close.docstatus = 1
        JOIN `tabMeter Reading` mc_open
            ON mc_open.shift = mr_close.shift
           AND mc_open.pump = mr_close.pump
           AND mc_open.nozzle_number = mr_close.nozzle_number
           AND mc_open.meter_type = 'Electronic Cash'
           AND mc_open.reading_position = 'Shift Open'
           AND mc_open.docstatus = 1
        JOIN `tabMeter Reading` mm_close
            ON mm_close.shift = mr_close.shift
           AND mm_close.pump = mr_close.pump
           AND mm_close.nozzle_number = mr_close.nozzle_number
           AND mm_close.meter_type = 'Manual Mechanical'
           AND mm_close.reading_position = 'Shift Close'
           AND mm_close.docstatus = 1
        JOIN `tabMeter Reading` mm_open
            ON mm_open.shift = mr_close.shift
           AND mm_open.pump = mr_close.pump
           AND mm_open.nozzle_number = mr_close.nozzle_number
           AND mm_open.meter_type = 'Manual Mechanical'
           AND mm_open.reading_position = 'Shift Open'
           AND mm_open.docstatus = 1
        JOIN `tabPump Nozzle` pn
            ON pn.parent = mr_close.pump
           AND pn.nozzle_number = mr_close.nozzle_number
        WHERE mr_close.shift = %(shift)s
          AND mr_close.meter_type = 'Electronic Volume'
          AND mr_close.reading_position = 'Shift Close'
          AND mr_close.docstatus = 1
        """,
        {"shift": shift.name},
        as_dict=True,
    )

    # Aggregate by product
    product_map = {}
    nozzle_summaries = []

    for row in nozzle_rows:
        item = row.item_code
        if item not in product_map:
            product_map[item] = {"volume": 0.0, "elec_cash": 0.0, "mech_vol": 0.0}
        product_map[item]["volume"]    += flt(row.elec_vol_sold)
        product_map[item]["elec_cash"] += flt(row.elec_cash_sold)
        product_map[item]["mech_vol"]  += flt(row.mech_vol_sold)

        # Per-nozzle summary row (for child table)
        mvr_name = frappe.db.get_value(
            "Meter Validation Result",
            {"shift": shift.name, "pump": row.pump, "nozzle_number": row.nozzle_number},
            "name",
        )
        nozzle_summaries.append({
            "pump":          row.pump,
            "nozzle":        row.nozzle_number,
            "elec_vol_sold": flt(row.elec_vol_sold),
            "elec_cash_sold": flt(row.elec_cash_sold),
            "mech_vol_sold": flt(row.mech_vol_sold),
            "check_a_status": frappe.db.get_value("Meter Validation Result", mvr_name, "check_a_status") if mvr_name else "",
            "check_b_status": frappe.db.get_value("Meter Validation Result", mvr_name, "check_b_status") if mvr_name else "",
            "linked_mvr":    mvr_name,
        })

    # Build product summary rows
    product_summaries = []
    for item_code, data in product_map.items():
        rate = _get_shift_rate(shift, item_code)
        vol  = flt(data["volume"])
        revenue = vol * rate
        # Get WAC from ERPNext stock ledger
        tank = _get_tank_for_item(item_code)
        wac  = get_current_wac(item_code, tank) if tank else 0.0
        cogs = vol * wac
        margin = revenue - cogs
        margin_pct = (margin / revenue * 100) if revenue else 0.0

        product_summaries.append({
            "fuel_product":     item_code,
            "volume_sold_l":    vol,
            "shift_rate":       rate,
            "gross_revenue":    revenue,
            "wac_unit_cost":    wac,
            "cogs":             cogs,
            "gross_margin":     margin,
            "gross_margin_pct": round(margin_pct, 2),
        })

    return product_summaries, nozzle_summaries


def _get_shift_rate(shift, item_code: str) -> float:
    """Return the EPRA rate locked on the shift for a given item_code."""
    rate_map = {
        "FUEL-PMS-UNL": "rate_pms_unl",
        "FUEL-PMS-VP":  "rate_pms_vp",
        "FUEL-AGO":     "rate_ago",
        "FUEL-DPK":     "rate_dpk",
    }
    field = rate_map.get(item_code)
    return flt(getattr(shift, field, 0)) if field else 0.0


def _get_tank_for_item(item_code: str) -> str:
    """Return the first active tank warehouse that stores the given fuel product."""
    result = frappe.db.sql(
        """
        SELECT pn.tank
        FROM `tabPump Nozzle` pn
        JOIN `tabForecourt Pump` fp ON fp.name = pn.parent
        WHERE pn.fuel_product = %(item_code)s
          AND fp.is_active = 1 AND pn.is_active = 1
        LIMIT 1
        """,
        {"item_code": item_code},
    )
    return result[0][0] if result else None


def _compute_cash_reconciliation(shift) -> dict:
    """Compute expected cash and variance for the shift."""
    from forecourt.utils.pos import get_shift_pos_totals

    # Opening float
    float_amount = flt(
        frappe.db.get_value(
            "Cash Event",
            {"shift": shift.name, "event_type": "Float Issued", "docstatus": 1},
            "amount",
        )
    )

    # Cash POS sales
    pos_totals = get_shift_pos_totals(shift.name)
    cash_sales = flt(pos_totals["by_payment"].get("Cash", 0))

    # Cash pickups and payouts (reduce expected cash in till)
    deductions = flt(frappe.db.sql(
        """
        SELECT COALESCE(SUM(amount), 0)
        FROM `tabCash Event`
        WHERE shift = %(shift)s
          AND event_type IN ('Cash Pickup', 'Payout')
          AND docstatus = 1
        """,
        {"shift": shift.name},
    )[0][0])

    # Drive-offs (fuel dispensed, cash not collected)
    drive_offs = flt(frappe.db.sql(
        """
        SELECT COALESCE(SUM(amount_kes), 0)
        FROM `tabDrive-Off Record`
        WHERE shift = %(shift)s AND docstatus = 1
        """,
        {"shift": shift.name},
    )[0][0])

    expected_cash = float_amount + cash_sales - deductions - drive_offs

    # Actual cash count from the Cashier Session
    actual_cash = flt(frappe.db.get_value(
        "Cashier Session",
        {"shift": shift.name, "is_primary": 1},
        "actual_cash_close",
    ))

    return {
        "float_amount":   float_amount,
        "cash_sales":     cash_sales,
        "deductions":     deductions,
        "drive_offs":     drive_offs,
        "expected_cash":  expected_cash,
        "actual_cash":    actual_cash,
        "cash_variance":  actual_cash - expected_cash,
    }
```

---

## 24. Python — Pre-Fetch Logic for Shift Close

```python
# forecourt/forecourt/doctype/forecourt_shift/pre_fetch.py

"""
Pre-fetch module — called when the supervisor clicks "Begin Shift Close".
Returns a structured dict that the client-side form uses to pre-populate
the closing wizard fields. Nothing is saved to the database here.
"""

import frappe
from frappe.utils import flt

from forecourt.utils.pos import (
    get_shift_pos_totals,
    get_shift_sales_invoice_totals,
)


@frappe.whitelist()
def get_shift_close_prefetch(shift_name: str) -> dict:
    """
    Collect and return all data that can be pre-populated on the shift
    close form without any new physical readings.

    Called by the "Begin Shift Close" button action on Forecourt Shift.
    Requires frappe.whitelist() so it can be called from client scripts.

    Returns:
        dict: structured pre-fetch data
    """
    shift = frappe.get_doc("Forecourt Shift", shift_name)

    return {
        "opening_readings":   _get_opening_readings(shift_name),
        "opening_dips":       _get_opening_dips(shift_name),
        "accepted_deliveries": _get_accepted_deliveries(shift_name),
        "cash_events":        _get_cash_events_summary(shift_name),
        "pos_totals":         get_shift_pos_totals(shift_name),
        "sales_invoice_totals": get_shift_sales_invoice_totals(shift_name),
        "drive_offs":         _get_drive_off_total(shift_name),
        "active_nozzles":     _get_active_nozzles_for_form(shift_name),
        "active_tanks":       _get_active_tanks_for_form(shift_name),
        "shift_rates": {
            "pms_unl": flt(shift.rate_pms_unl),
            "pms_vp":  flt(shift.rate_pms_vp),
            "ago":     flt(shift.rate_ago),
            "dpk":     flt(shift.rate_dpk),
        },
    }


def _get_opening_readings(shift_name: str) -> list:
    """
    Return all submitted Shift Open meter readings for the shift.
    Displayed as reference next to each closing reading input field.
    """
    return frappe.db.sql(
        """
        SELECT
            pump,
            nozzle_number,
            meter_type,
            totalizer_value,
            unit
        FROM `tabMeter Reading`
        WHERE shift = %(shift)s
          AND reading_position = 'Shift Open'
          AND docstatus = 1
        ORDER BY pump, nozzle_number, meter_type
        """,
        {"shift": shift_name},
        as_dict=True,
    )


def _get_opening_dips(shift_name: str) -> list:
    """Return submitted Shift Open dip readings per tank."""
    return frappe.db.sql(
        """
        SELECT
            tank,
            volume_observed_l,
            observed_at
        FROM `tabDip Reading`
        WHERE shift = %(shift)s
          AND reading_type = 'Shift Open'
          AND docstatus = 1
        ORDER BY tank
        """,
        {"shift": shift_name},
        as_dict=True,
    )


def _get_accepted_deliveries(shift_name: str) -> list:
    """Return accepted fuel deliveries for the shift, per tank."""
    return frappe.db.sql(
        """
        SELECT
            tank,
            fuel_product,
            SUM(dip_measured_l) AS total_delivered_l,
            GROUP_CONCAT(docket_number SEPARATOR ', ') AS docket_numbers
        FROM `tabFuel Delivery Dip`
        WHERE shift = %(shift)s
          AND status = 'Accepted'
          AND docstatus = 1
        GROUP BY tank, fuel_product
        """,
        {"shift": shift_name},
        as_dict=True,
    )


def _get_cash_events_summary(shift_name: str) -> dict:
    """
    Return cash event totals that feed the expected cash calculation.
    """
    events = frappe.db.sql(
        """
        SELECT event_type, SUM(amount) AS total
        FROM `tabCash Event`
        WHERE shift = %(shift)s AND docstatus = 1
        GROUP BY event_type
        """,
        {"shift": shift_name},
        as_dict=True,
    )
    summary = {e.event_type: flt(e.total) for e in events}
    return {
        "float_issued":  summary.get("Float Issued", 0.0),
        "cash_pickups":  summary.get("Cash Pickup", 0.0),
        "payouts":       summary.get("Payout", 0.0),
    }


def _get_drive_off_total(shift_name: str) -> float:
    """Return total drive-off amount for the shift."""
    result = frappe.db.sql(
        """
        SELECT COALESCE(SUM(amount_kes), 0)
        FROM `tabDrive-Off Record`
        WHERE shift = %(shift)s AND docstatus = 1
        """,
        {"shift": shift_name},
    )
    return flt(result[0][0]) if result else 0.0


def _get_active_nozzles_for_form(shift_name: str) -> list:
    """
    Return the nozzle list structure needed to render the closing meter
    reading form sections. One entry per active nozzle.
    """
    return frappe.db.sql(
        """
        SELECT
            fp.name      AS pump,
            fp.pump_number,
            fp.pump_name,
            pn.nozzle_number,
            pn.fuel_product,
            pn.tank,
            pn.colour_code
        FROM `tabForecourt Pump` fp
        JOIN `tabPump Nozzle` pn ON pn.parent = fp.name
        WHERE fp.is_active = 1 AND pn.is_active = 1
        ORDER BY fp.pump_number, pn.nozzle_number
        """,
        as_dict=True,
    )


def _get_active_tanks_for_form(shift_name: str) -> list:
    """Return unique active tanks with product mapping for the dip form."""
    return frappe.db.sql(
        """
        SELECT DISTINCT
            pn.tank,
            pn.fuel_product
        FROM `tabForecourt Pump` fp
        JOIN `tabPump Nozzle` pn ON pn.parent = fp.name
        WHERE fp.is_active = 1 AND pn.is_active = 1
        ORDER BY pn.tank
        """,
        as_dict=True,
    )
```

---

## 25. Python — GL Journal Entry Creation

```python
# forecourt/utils/gl.py

import frappe
from frappe.utils import flt, now_datetime

from forecourt.utils.accounts import get_account, get_fuel_accounts, FUEL_ACCOUNT_KEYS


@frappe.whitelist()
def post_shift_journal_entry(shift_reconciliation_name: str) -> str:
    """
    Create and submit the shift GL Journal Entry from a completed
    Shift Reconciliation. Handles mixed tenders, cash variance,
    and wetstock loss/gain.

    Must be idempotent — if called twice, the second call raises
    an error if a GL Journal is already linked.

    Args:
        shift_reconciliation_name: Shift Reconciliation document name

    Returns:
        str: Journal Entry name

    Raises:
        frappe.ValidationError on duplicate or data issues
    """
    sr   = frappe.get_doc("Shift Reconciliation", shift_reconciliation_name)
    shift = frappe.get_doc("Forecourt Shift", sr.shift)

    if sr.gl_journal:
        frappe.throw(
            f"Journal Entry {sr.gl_journal} already posted for this reconciliation. "
            "Reverse the existing entry before re-posting.",
            title="Duplicate GL Post",
        )

    if sr.requires_approval and not sr.approved_by:
        frappe.throw(
            "This reconciliation has Critical variances and requires manager approval "
            "before the GL can be posted.",
            title="Manager Approval Required",
        )

    company = shift.company or frappe.defaults.get_global_default("company")
    cost_center = frappe.get_cached_value("Company", company, "cost_center")

    je = frappe.new_doc("Journal Entry")
    je.voucher_type   = "Journal Entry"
    je.posting_date   = shift.shift_date
    je.company        = company
    je.user_remark    = (
        f"Shift close — {shift.name} | Cashier: {shift.cashier} | "
        f"Supervisor: {shift.supervisor}"
    )

    accounts = []  # List of account rows

    # ------------------------------------------------------------------
    # Revenue lines — one credit per active fuel product
    # ------------------------------------------------------------------
    for ps in sr.product_summaries:
        item_code = ps.fuel_product
        accts = get_fuel_accounts(item_code, company)
        if flt(ps.gross_revenue) == 0:
            continue
        accounts.append({
            "account":      accts["sales"],
            "credit":       flt(ps.gross_revenue),
            "debit":        0,
            "cost_center":  cost_center,
            "user_remark":  f"{item_code} — {flt(ps.volume_sold_l):.3f} L @ {flt(ps.shift_rate):.2f}",
        })

    # ------------------------------------------------------------------
    # Debit lines — tenders received
    # ------------------------------------------------------------------
    total_revenue = flt(sr.total_gross_revenue)

    # Cash tender: use actual cash count (variance is handled separately)
    actual_cash = flt(sr.actual_cash)
    if actual_cash > 0:
        accounts.append({
            "account":     get_account("till_active", company),
            "debit":       actual_cash,
            "credit":      0,
            "cost_center": cost_center,
            "user_remark": "Cash — actual count",
        })

    # MPesa clearing
    if flt(sr.mpesa_total) > 0:
        accounts.append({
            "account":     get_account("mpesa_clearing", company),
            "debit":       flt(sr.mpesa_total),
            "credit":      0,
            "cost_center": cost_center,
            "user_remark": "MPesa collections",
        })

    # Card clearing
    if flt(sr.card_total) > 0:
        accounts.append({
            "account":     get_account("card_clearing", company),
            "debit":       flt(sr.card_total),
            "credit":      0,
            "cost_center": cost_center,
            "user_remark": "Card terminal collections",
        })

    # Fleet / Credit: these are already in AR from Sales Invoices.
    # Post a Fleet Card Clearing debit here only if fleet paid cash at POS,
    # not for credit terms customers (skip if fleet_credit_total is AR-based).
    if flt(sr.fleet_credit_total) > 0:
        accounts.append({
            "account":     get_account("fleet_card_clearing", company),
            "debit":       flt(sr.fleet_credit_total),
            "credit":      0,
            "cost_center": cost_center,
            "user_remark": "Fleet card / credit — see linked Sales Invoices",
        })

    # ------------------------------------------------------------------
    # Cash variance line (Short or Over)
    # ------------------------------------------------------------------
    cash_var = flt(sr.cash_variance)  # actual − expected
    if abs(cash_var) > 0.01:
        cash_short_over_acct = get_account("cash_short_over", company)
        if cash_var < 0:
            # Cashier SHORT: we received less cash than expected → debit Cash Short
            accounts.append({
                "account":     cash_short_over_acct,
                "debit":       abs(cash_var),
                "credit":      0,
                "cost_center": cost_center,
                "user_remark": f"Cash short — {shift.cashier}",
            })
        else:
            # Cashier OVER: we received more cash than expected → credit Cash Short/Over
            accounts.append({
                "account":     cash_short_over_acct,
                "debit":       0,
                "credit":      cash_var,
                "cost_center": cost_center,
                "user_remark": f"Cash over — {shift.cashier}",
            })

    # ------------------------------------------------------------------
    # COGS lines (only if Perpetual Inventory is disabled for fuel items)
    # ------------------------------------------------------------------
    perpetual_inv = frappe.get_cached_value(
        "Company", company, "enable_perpetual_inventory"
    )
    if not perpetual_inv:
        for ps in sr.product_summaries:
            item_code = ps.fuel_product
            accts = get_fuel_accounts(item_code, company)
            if flt(ps.cogs) == 0:
                continue
            accounts.append({
                "account":     accts["cogs"],
                "debit":       flt(ps.cogs),
                "credit":      0,
                "cost_center": cost_center,
                "user_remark": f"COGS {item_code} — {flt(ps.volume_sold_l):.3f} L × WAC {flt(ps.wac_unit_cost):.4f}",
            })
            accounts.append({
                "account":     accts["inventory"],
                "debit":       0,
                "credit":      flt(ps.cogs),
                "cost_center": cost_center,
                "user_remark": f"Fuel Inventory relief — {item_code}",
            })

    # ------------------------------------------------------------------
    # Wetstock variance lines (expense if loss, income-reversal if gain)
    # ------------------------------------------------------------------
    for ts in sr.tank_wetstock_summaries:
        var_kes = flt(ts.variance_kes)
        if abs(var_kes) < 1.0:
            continue  # Sub-KES-1 rounding — skip
        accts = get_fuel_accounts(ts.fuel_product, company)
        if var_kes > 0:
            # Theoretical > Actual → fuel lost → expense
            accounts.append({
                "account":     accts["wetstock_variance"],
                "debit":       var_kes,
                "credit":      0,
                "cost_center": cost_center,
                "user_remark": f"Wetstock loss — {ts.tank} — {flt(ts.variance_l):.3f} L",
            })
            accounts.append({
                "account":     accts["inventory"],
                "debit":       0,
                "credit":      var_kes,
                "cost_center": cost_center,
                "user_remark": f"Inventory reduction — wetstock loss {ts.tank}",
            })
        else:
            # Theoretical < Actual → apparent gain → credit variance account
            accounts.append({
                "account":     accts["wetstock_variance"],
                "debit":       0,
                "credit":      abs(var_kes),
                "cost_center": cost_center,
                "user_remark": f"Wetstock gain — {ts.tank} — investigate",
            })
            accounts.append({
                "account":     accts["inventory"],
                "debit":       abs(var_kes),
                "credit":      0,
                "cost_center": cost_center,
                "user_remark": f"Inventory increase — apparent wetstock gain {ts.tank}",
            })

    # ------------------------------------------------------------------
    # Build Journal Entry accounts child table
    # ------------------------------------------------------------------
    for row in accounts:
        je.append("accounts", {
            "account":        row["account"],
            "debit_in_account_currency":  flt(row.get("debit", 0)),
            "credit_in_account_currency": flt(row.get("credit", 0)),
            "cost_center":    row.get("cost_center", ""),
            "user_remark":    row.get("user_remark", ""),
        })

    # Validate balance
    total_debit  = sum(flt(a.get("debit", 0))  for a in accounts)
    total_credit = sum(flt(a.get("credit", 0)) for a in accounts)
    if abs(total_debit - total_credit) > 0.05:
        frappe.throw(
            f"Journal Entry is out of balance: "
            f"DR {total_debit:.2f} / CR {total_credit:.2f} "
            f"(diff {total_debit - total_credit:.2f}). "
            "Review product summaries and tender totals.",
            title="GL Imbalance",
        )

    je.insert(ignore_permissions=True)
    je.submit()

    # Link back
    frappe.db.set_value("Shift Reconciliation", sr.name, "gl_journal", je.name)
    frappe.db.set_value("Forecourt Shift",      sr.shift, "gl_journal", je.name)

    frappe.logger().info(
        f"Journal Entry {je.name} posted for Shift {sr.shift} "
        f"(DR {total_debit:.2f} / CR {total_credit:.2f})"
    )
    return je.name
```

---

## 26. Python — Wetstock Variance Calculation

```python
# forecourt/utils/wetstock.py

import frappe
from frappe.utils import flt


# Variance classification thresholds (%)
NORMAL_PCT   = 0.30
ELEVATED_PCT = 0.50


def compute_tank_wetstock(shift_name: str, tank: str) -> dict:
    """
    Compute the full wetstock equation for one tank in one shift.

    Returns:
        dict with keys:
            tank, fuel_product,
            opening_stock_l, deliveries_l,
            elec_vol_sales_l, mech_vol_sales_l,
            theoretical_closing_l, actual_closing_l,
            variance_l, variance_pct, classification,
            wac, variance_kes
    """
    # 1. Opening stock
    opening = frappe.db.get_value(
        "Dip Reading",
        {
            "shift":        shift_name,
            "tank":         tank,
            "reading_type": "Shift Open",
            "docstatus":    1,
        },
        "volume_observed_l",
    )
    if opening is None:
        frappe.throw(
            f"No submitted Shift Open dip found for tank {tank} in shift {shift_name}.",
            title="Missing Opening Dip",
        )
    opening_stock_l = flt(opening)

    # 2. Accepted deliveries
    deliveries_result = frappe.db.sql(
        """
        SELECT COALESCE(SUM(dip_measured_l), 0)
        FROM `tabFuel Delivery Dip`
        WHERE shift = %(shift)s
          AND tank = %(tank)s
          AND status = 'Accepted'
          AND docstatus = 1
        """,
        {"shift": shift_name, "tank": tank},
    )
    deliveries_l = flt(deliveries_result[0][0]) if deliveries_result else 0.0

    # 3. Meter sales (Electronic Volume) for all nozzles on this tank
    sales_result = frappe.db.sql(
        """
        SELECT
            COALESCE(SUM(mr_close.totalizer_value - mr_open.totalizer_value), 0) AS elec_vol,
            COALESCE(SUM(mm_close.totalizer_value - mm_open.totalizer_value), 0) AS mech_vol
        FROM `tabMeter Reading` mr_close
        JOIN `tabMeter Reading` mr_open
            ON mr_open.shift = mr_close.shift
           AND mr_open.pump = mr_close.pump
           AND mr_open.nozzle_number = mr_close.nozzle_number
           AND mr_open.meter_type = 'Electronic Volume'
           AND mr_open.reading_position = 'Shift Open'
           AND mr_open.docstatus = 1
        JOIN `tabMeter Reading` mm_close
            ON mm_close.shift = mr_close.shift
           AND mm_close.pump = mr_close.pump
           AND mm_close.nozzle_number = mr_close.nozzle_number
           AND mm_close.meter_type = 'Manual Mechanical'
           AND mm_close.reading_position = 'Shift Close'
           AND mm_close.docstatus = 1
        JOIN `tabMeter Reading` mm_open
            ON mm_open.shift = mr_close.shift
           AND mm_open.pump = mr_close.pump
           AND mm_open.nozzle_number = mr_close.nozzle_number
           AND mm_open.meter_type = 'Manual Mechanical'
           AND mm_open.reading_position = 'Shift Open'
           AND mm_open.docstatus = 1
        JOIN `tabPump Nozzle` pn
            ON pn.parent = mr_close.pump
           AND pn.nozzle_number = mr_close.nozzle_number
           AND pn.tank = %(tank)s
        WHERE mr_close.shift = %(shift)s
          AND mr_close.meter_type = 'Electronic Volume'
          AND mr_close.reading_position = 'Shift Close'
          AND mr_close.docstatus = 1
        """,
        {"shift": shift_name, "tank": tank},
        as_dict=True,
    )
    elec_vol_sales_l = flt(sales_result[0].elec_vol) if sales_result else 0.0
    mech_vol_sales_l = flt(sales_result[0].mech_vol) if sales_result else 0.0

    # 4. Theoretical closing
    theoretical_closing_l = opening_stock_l + deliveries_l - elec_vol_sales_l

    # 5. Actual closing (from closing dip)
    actual = frappe.db.get_value(
        "Dip Reading",
        {
            "shift":        shift_name,
            "tank":         tank,
            "reading_type": "Shift Close",
            "docstatus":    1,
        },
        "volume_observed_l",
    )
    if actual is None:
        frappe.throw(
            f"No submitted Shift Close dip found for tank {tank} in shift {shift_name}.",
            title="Missing Closing Dip",
        )
    actual_closing_l = flt(actual)

    # 6. Variance
    variance_l = theoretical_closing_l - actual_closing_l
    denominator = opening_stock_l + deliveries_l
    variance_pct = (variance_l / denominator * 100) if denominator else 0.0

    # 7. Classification
    abs_pct = abs(variance_pct)
    if variance_l < 0 and abs_pct > ELEVATED_PCT:
        classification = "Critical"
    elif variance_l < 0 and abs_pct > NORMAL_PCT:
        classification = "Elevated"
    elif variance_l > 0:
        classification = "Gain" if abs_pct > NORMAL_PCT else "Normal"
    else:
        classification = "Normal"

    # 8. WAC and KES value
    from forecourt.utils.stock import get_current_wac
    fuel_product = frappe.db.get_value(
        "Dip Reading",
        {"shift": shift_name, "tank": tank, "reading_type": "Shift Open", "docstatus": 1},
        "fuel_product",
    )
    # Fall back: look up from nozzle mapping
    if not fuel_product:
        fuel_product = frappe.db.get_value(
            "Pump Nozzle", {"tank": tank, "is_active": 1}, "fuel_product"
        )
    wac = get_current_wac(fuel_product, tank) if fuel_product else 0.0
    variance_kes = variance_l * wac

    return {
        "tank":                 tank,
        "fuel_product":         fuel_product or "",
        "opening_stock_l":      opening_stock_l,
        "deliveries_l":         deliveries_l,
        "elec_vol_sales_l":     elec_vol_sales_l,
        "mech_vol_sales_l":     mech_vol_sales_l,
        "theoretical_closing_l": theoretical_closing_l,
        "actual_closing_l":     actual_closing_l,
        "variance_l":           variance_l,
        "variance_pct":         round(variance_pct, 4),
        "classification":       classification,
        "wac":                  wac,
        "variance_kes":         variance_kes,
    }


def _compute_tank_wetstock(shift, product_summaries: list, company: str) -> list:
    """
    Compute wetstock for all active tanks in a shift.
    Returns list of dicts for the Tank Wetstock Summaries child table.
    """
    active_tanks = frappe.db.sql(
        """
        SELECT DISTINCT pn.tank
        FROM `tabForecourt Pump` fp
        JOIN `tabPump Nozzle` pn ON pn.parent = fp.name
        WHERE fp.is_active = 1 AND pn.is_active = 1
        """,
        as_list=True,
    )
    results = []
    for (tank,) in active_tanks:
        try:
            result = compute_tank_wetstock(shift.name, tank)
            results.append(result)
        except frappe.ValidationError as e:
            frappe.log_error(
                f"Wetstock computation failed for tank {tank}: {e}",
                "Wetstock Error",
            )
    return results
```

---

## 27. Python — Meter Validation Checks

```python
# forecourt/utils/meter.py

import frappe
from frappe.utils import flt, now_datetime


# Check A tolerances (KES)
CHECK_A_PASS_KES    = 5.0
CHECK_A_WARN_KES    = 20.0

# Check B tolerances (%)
CHECK_B_PASS_PCT    = 0.30
CHECK_B_WARN_PCT    = 0.50
CHECK_B_TAMPER_PCT  = 1.00


def compute_meter_validation_results(shift_name: str) -> list:
    """
    Run Check A (cash vs volume) and Check B (mechanical vs electronic)
    for all active nozzles in a shift.

    Creates / updates Meter Validation Result documents.
    Returns a list of result dicts.

    Raises:
        frappe.ValidationError if closing readings are not yet submitted.
    """
    shift = frappe.get_doc("Forecourt Shift", shift_name)

    # Get all nozzle reading triplets (Elec Vol, Elec Cash, Manual Mech)
    nozzle_data = _get_nozzle_reading_triplets(shift_name)
    if not nozzle_data:
        frappe.throw(
            f"No submitted closing meter readings found for shift {shift_name}. "
            "Submit all closing meter readings before running validation.",
            title="No Readings Found",
        )

    results = []
    for nd in nozzle_data:
        result = _validate_nozzle(nd, shift)
        _save_mvr(result, shift_name)
        results.append(result)

        # Lock pump immediately if tamper threshold exceeded
        if result["check_b_status"] == "Critical":
            _lock_pump(result["pump"])

    return results


def _get_nozzle_reading_triplets(shift_name: str) -> list:
    """
    Return one dict per nozzle with opening and closing values for all
    three meter types.
    """
    return frappe.db.sql(
        """
        SELECT
            ev_c.pump,
            ev_c.nozzle_number,
            pn.fuel_product,
            pn.tank,
            (ev_c.totalizer_value - ev_o.totalizer_value) AS elec_vol_sold,
            (ec_c.totalizer_value - ec_o.totalizer_value) AS elec_cash_sold,
            (mm_c.totalizer_value - mm_o.totalizer_value) AS mech_vol_sold
        FROM `tabMeter Reading` ev_c
        -- Elec Vol open
        JOIN `tabMeter Reading` ev_o
            ON ev_o.shift=ev_c.shift AND ev_o.pump=ev_c.pump
           AND ev_o.nozzle_number=ev_c.nozzle_number
           AND ev_o.meter_type='Electronic Volume'
           AND ev_o.reading_position='Shift Open' AND ev_o.docstatus=1
        -- Elec Cash close
        JOIN `tabMeter Reading` ec_c
            ON ec_c.shift=ev_c.shift AND ec_c.pump=ev_c.pump
           AND ec_c.nozzle_number=ev_c.nozzle_number
           AND ec_c.meter_type='Electronic Cash'
           AND ec_c.reading_position='Shift Close' AND ec_c.docstatus=1
        -- Elec Cash open
        JOIN `tabMeter Reading` ec_o
            ON ec_o.shift=ev_c.shift AND ec_o.pump=ev_c.pump
           AND ec_o.nozzle_number=ev_c.nozzle_number
           AND ec_o.meter_type='Electronic Cash'
           AND ec_o.reading_position='Shift Open' AND ec_o.docstatus=1
        -- Manual Mech close
        JOIN `tabMeter Reading` mm_c
            ON mm_c.shift=ev_c.shift AND mm_c.pump=ev_c.pump
           AND mm_c.nozzle_number=ev_c.nozzle_number
           AND mm_c.meter_type='Manual Mechanical'
           AND mm_c.reading_position='Shift Close' AND mm_c.docstatus=1
        -- Manual Mech open
        JOIN `tabMeter Reading` mm_o
            ON mm_o.shift=ev_c.shift AND mm_o.pump=ev_c.pump
           AND mm_o.nozzle_number=ev_c.nozzle_number
           AND mm_o.meter_type='Manual Mechanical'
           AND mm_o.reading_position='Shift Open' AND mm_o.docstatus=1
        -- Nozzle mapping
        JOIN `tabPump Nozzle` pn
            ON pn.parent=ev_c.pump AND pn.nozzle_number=ev_c.nozzle_number
        WHERE ev_c.shift=%(shift)s
          AND ev_c.meter_type='Electronic Volume'
          AND ev_c.reading_position='Shift Close'
          AND ev_c.docstatus=1
        """,
        {"shift": shift_name},
        as_dict=True,
    )


def _validate_nozzle(nd: dict, shift) -> dict:
    """Run Check A and Check B for one nozzle and return a result dict."""
    elec_vol   = flt(nd.elec_vol_sold)
    elec_cash  = flt(nd.elec_cash_sold)
    mech_vol   = flt(nd.mech_vol_sold)

    # Get shift rate for this nozzle's product
    rate = _get_rate_for_product(shift, nd.fuel_product)

    # Check A: cash vs volume consistency
    expected_cash      = elec_vol * rate
    check_a_discrepancy = abs(elec_cash - expected_cash)

    if check_a_discrepancy <= CHECK_A_PASS_KES:
        check_a_status = "Pass"
    elif check_a_discrepancy <= CHECK_A_WARN_KES:
        check_a_status = "Warning"
    else:
        check_a_status = "Fail"

    # Check B: mechanical vs electronic consistency
    if elec_vol > 0:
        check_b_pct = abs(elec_vol - mech_vol) / elec_vol * 100
    else:
        check_b_pct = 0.0

    if check_b_pct <= CHECK_B_PASS_PCT:
        check_b_status = "Pass"
    elif check_b_pct <= CHECK_B_WARN_PCT:
        check_b_status = "Warning"
    elif check_b_pct <= CHECK_B_TAMPER_PCT:
        check_b_status = "Fail"
    else:
        check_b_status = "Critical"

    # Overall: worst of A and B
    severity = {"Pass": 0, "Warning": 1, "Fail": 2, "Critical": 3}
    worst = max(check_a_status, check_b_status, key=lambda s: severity[s])

    return {
        "pump":                nd.pump,
        "nozzle_number":       nd.nozzle_number,
        "fuel_product":        nd.fuel_product,
        "tank":                nd.tank,
        "shift_rate":          rate,
        "elec_vol_sold":       elec_vol,
        "elec_cash_sold":      elec_cash,
        "mech_vol_sold":       mech_vol,
        "expected_cash":       expected_cash,
        "check_a_discrepancy": check_a_discrepancy,
        "check_a_status":      check_a_status,
        "check_b_divergence_pct": round(check_b_pct, 4),
        "check_b_status":      check_b_status,
        "overall_status":      worst,
        "computed_at":         now_datetime(),
    }


def _get_rate_for_product(shift, fuel_product: str) -> float:
    rate_map = {
        "FUEL-PMS-UNL": "rate_pms_unl",
        "FUEL-PMS-VP":  "rate_pms_vp",
        "FUEL-AGO":     "rate_ago",
        "FUEL-DPK":     "rate_dpk",
    }
    field = rate_map.get(fuel_product)
    return flt(getattr(shift, field, 0.0)) if field else 0.0


def _save_mvr(result: dict, shift_name: str):
    """Upsert a Meter Validation Result document for the nozzle."""
    existing = frappe.db.get_value(
        "Meter Validation Result",
        {"shift": shift_name, "pump": result["pump"], "nozzle_number": result["nozzle_number"]},
        "name",
    )
    if existing:
        mvr = frappe.get_doc("Meter Validation Result", existing)
    else:
        mvr = frappe.new_doc("Meter Validation Result")
        mvr.shift = shift_name

    for key, val in result.items():
        if hasattr(mvr, key):
            setattr(mvr, key, val)

    mvr.save(ignore_permissions=True)


def _lock_pump(pump_name: str):
    """
    Set pump is_active = 0 when tamper threshold exceeded.
    Logs the action for the audit trail.
    """
    frappe.db.set_value("Forecourt Pump", pump_name, "is_active", 0)
    frappe.db.set_value(
        "Forecourt Pump", pump_name, "notes",
        f"AUTO-LOCKED {now_datetime()} — Meter Check B exceeded 1.0% tamper threshold. "
        "Contact KEBS / Weights & Measures before reactivating.",
    )
    frappe.logger().warning(
        f"Pump {pump_name} AUTO-LOCKED due to Check B Critical (>1.0% divergence)."
    )
    frappe.publish_realtime(
        event="pump_locked",
        message={"pump": pump_name, "reason": "Meter Check B Critical — possible tampering"},
        room=frappe.local.site,
    )
```

---

## 28. Python — POS Invoice Blocking Until Shift Opens

```python
# forecourt/overrides/pos_invoice.py

"""
Override for POS Invoice — enforces that no fuel invoice can be submitted
unless a valid Open Forecourt Shift exists (Rule 5).

Register in hooks.py:
    doc_events = {
        "POS Invoice": {
            "before_submit": "forecourt.overrides.pos_invoice.before_submit",
            "on_submit":     "forecourt.overrides.pos_invoice.on_submit",
        }
    }
"""

import frappe
from frappe.utils import flt


FUEL_ITEM_CODES = {"FUEL-PMS-UNL", "FUEL-PMS-VP", "FUEL-AGO", "FUEL-DPK"}


def before_submit(doc, method=None):
    """
    Block POS Invoice submission if:
    1. Any line item is a fuel product AND
    2. No Open Forecourt Shift exists for the station.

    Non-fuel invoices (shop items, lubricants) are not blocked.
    """
    if not _invoice_contains_fuel(doc):
        return

    company = doc.company or frappe.defaults.get_global_default("company")
    open_shift = _get_open_shift(company)

    if not open_shift:
        frappe.throw(
            "No open Forecourt Shift found for this station. "
            "A supervisor must open a shift and capture opening meter readings "
            "before fuel sales can be recorded.",
            title="Shift Not Open — POS Blocked",
        )

    # Link the invoice to the open shift
    doc.forecourt_shift = open_shift

    # Validate attendant is set
    _validate_attendant(doc)

    # Validate payment method is specified on all lines
    _validate_payment_method(doc)


def on_submit(doc, method=None):
    """
    On submit: verify the linked shift is still Open (race condition guard).
    """
    if not doc.forecourt_shift:
        return

    status = frappe.db.get_value("Forecourt Shift", doc.forecourt_shift, "status")
    if status != "Open":
        frappe.throw(
            f"Forecourt Shift {doc.forecourt_shift} is no longer Open "
            f"(current status: {status}). Cannot submit POS Invoice.",
            title="Shift Closed — Cannot Submit",
        )


def _invoice_contains_fuel(doc) -> bool:
    """Return True if any invoice line is a fuel item."""
    return any(item.item_code in FUEL_ITEM_CODES for item in doc.items)


def _get_open_shift(company: str) -> str | None:
    """Return the name of the Open Forecourt Shift for the company/station."""
    return frappe.db.get_value(
        "Forecourt Shift",
        {"company": company, "status": "Open"},
        "name",
    )


def _validate_attendant(doc):
    """
    Every fuel line must have an attendant. Raises if attendant is missing
    or is the placeholder 'N/A'.
    """
    # Attendant is set at header level via a custom field 'pump_attendant'
    attendant = getattr(doc, "pump_attendant", None)
    if not attendant or attendant.lower() in ("n/a", "na", "none", ""):
        frappe.throw(
            "Pump Attendant is required on all fuel POS Invoices. "
            "Select the attendant who dispensed the fuel before submitting.",
            title="Attendant Required",
        )


def _validate_payment_method(doc):
    """
    Payment method must be set and not left blank. ERPNext allows unspecified
    payment methods in some configurations; this catches the gap.
    """
    for payment in doc.payments:
        if flt(payment.amount) > 0 and not payment.mode_of_payment:
            frappe.throw(
                "Payment method must be selected for all non-zero payment lines. "
                "Do not leave a payment method blank.",
                title="Payment Method Required",
            )
```

---

## 29. JavaScript — Client-Side Real-Time Cross-Checks

### 29.1 Meter Reading Form — Instant Delta and Cross-Check

```javascript
// forecourt/public/js/meter_reading.js
// Provides real-time feedback as the supervisor types closing readings.

frappe.ui.form.on("Meter Reading", {
    // Trigger on any totalizer value change
    totalizer_value(frm) {
        if (frm.doc.reading_position === "Shift Close") {
            fetch_opening_and_check(frm);
        }
    },

    reading_position(frm) {
        if (frm.doc.reading_position === "Amendment") {
            frm.set_df_property("amendment_reason", "reqd", 1);
            frm.set_df_property("superseded_by",    "reqd", 1);
        } else {
            frm.set_df_property("amendment_reason", "reqd", 0);
            frm.set_df_property("superseded_by",    "reqd", 0);
        }
    },

    refresh(frm) {
        // Show computed delta if already saved
        if (frm.doc.reading_position === "Shift Close" && frm.doc.shift) {
            fetch_opening_and_check(frm);
        }
    }
});


function fetch_opening_and_check(frm) {
    if (!frm.doc.shift || !frm.doc.pump || !frm.doc.nozzle_number || !frm.doc.meter_type) {
        return;
    }
    if (!frm.doc.totalizer_value || frm.doc.totalizer_value <= 0) {
        clear_delta_indicator(frm);
        return;
    }

    frappe.call({
        method: "frappe.client.get_value",
        args: {
            doctype:  "Meter Reading",
            filters: {
                shift:            frm.doc.shift,
                pump:             frm.doc.pump,
                nozzle_number:    frm.doc.nozzle_number,
                meter_type:       frm.doc.meter_type,
                reading_position: "Shift Open",
                docstatus:        1,
            },
            fieldname: "totalizer_value",
        },
        callback(r) {
            if (!r.message || !r.message.totalizer_value) {
                set_indicator(frm, "orange", "⚠ No opening reading found");
                return;
            }

            const opening = r.message.totalizer_value;
            const closing = frm.doc.totalizer_value;
            const delta   = closing - opening;

            if (delta < 0) {
                set_indicator(frm, "red",
                    `✗ Closing (${closing}) < Opening (${opening}). Meters only count forward.`);
                return;
            }

            const unit = frm.doc.meter_type === "Electronic Cash" ? "KES" : "L";
            set_indicator(frm, "green",
                `✓ Opening: ${opening.toLocaleString()} | Delta: ${delta.toFixed(3)} ${unit}`);

            // For Electronic Volume: if we also have the rate, show quick sanity check
            if (frm.doc.meter_type === "Electronic Volume") {
                check_vs_elec_cash(frm, delta);
            }
        }
    });
}


function check_vs_elec_cash(frm, elec_vol_sold) {
    // Fetch the Elec Cash closing reading for same nozzle to cross-check
    frappe.call({
        method: "forecourt.utils.meter_api.get_nozzle_shift_rate",
        args: { shift: frm.doc.shift, fuel_product_for_pump: frm.doc.pump },
        callback(r) {
            if (!r.message) return;
            const rate = r.message;
            const expected_cash = elec_vol_sold * rate;
            // Show the expected cash in a helper note
            frm.set_df_property("totalizer_value", "description",
                `Delta: ${elec_vol_sold.toFixed(3)} L × KES ${rate.toFixed(2)} = ` +
                `KES ${expected_cash.toFixed(2)} expected in Elec Cash delta`
            );
        }
    });
}


function set_indicator(frm, colour, message) {
    // Use Frappe's built-in indicator or a custom HTML wrapper
    frm.dashboard.set_headline_alert(`
        <span class="indicator ${colour}">
            <span>${message}</span>
        </span>
    `);
}

function clear_delta_indicator(frm) {
    frm.dashboard.clear_headline();
}
```

### 29.2 Forecourt Shift Form — Shift Close Wizard

```javascript
// forecourt/public/js/forecourt_shift.js
// Manages the shift close wizard: pre-fetch, closing entry sections,
// real-time wetstock and cash variance display.

frappe.ui.form.on("Forecourt Shift", {

    refresh(frm) {
        add_custom_buttons(frm);
        display_status_indicator(frm);
    },

    status(frm) {
        display_status_indicator(frm);
    }
});


// -----------------------------------------------------------------------
// Custom buttons
// -----------------------------------------------------------------------

function add_custom_buttons(frm) {
    frm.clear_custom_buttons();

    if (frm.doc.status === "Open") {
        frm.add_custom_button("Begin Shift Close", () => begin_shift_close(frm), "Actions");
    }

    if (frm.doc.status === "Readings Captured") {
        frm.add_custom_button("Run Meter Validation", () => run_meter_validation(frm), "Actions");
    }

    if (frm.doc.status === "Closing") {
        frm.add_custom_button("Compute Reconciliation", () => compute_reconciliation(frm), "Actions");
    }

    if (frm.doc.status === "Closing" && frm.doc.__onload?.reconciliation_ready) {
        frm.add_custom_button("Approve and Post GL", () => approve_and_post_gl(frm), "Actions");
    }
}


// -----------------------------------------------------------------------
// Begin Shift Close — Pre-Fetch
// -----------------------------------------------------------------------

async function begin_shift_close(frm) {
    frappe.dom.freeze("Loading shift data…");

    try {
        const r = await frappe.call({
            method: "forecourt.forecourt.doctype.forecourt_shift.pre_fetch.get_shift_close_prefetch",
            args:   { shift_name: frm.doc.name },
        });

        frappe.dom.unfreeze();

        if (!r.message) {
            frappe.msgprint("Failed to load shift data. Please refresh and try again.");
            return;
        }

        render_shift_close_dialog(frm, r.message);

    } catch (err) {
        frappe.dom.unfreeze();
        frappe.msgprint(`Error loading shift data: ${err.message}`);
    }
}


function render_shift_close_dialog(frm, data) {
    const {
        opening_readings,
        active_nozzles,
        active_tanks,
        cash_events,
        pos_totals,
        sales_invoice_totals,
        drive_offs,
        shift_rates,
    } = data;

    // Build a tabbed dialog: Meters | Dips | Cash | Summary
    const d = new frappe.ui.Dialog({
        title:  `Close Shift — ${frm.doc.name}`,
        size:   "extra-large",
        fields: build_close_dialog_fields(active_nozzles, active_tanks, opening_readings),
        primary_action_label: "Save Closing Readings",
        primary_action(values) {
            save_closing_readings(frm, values, active_nozzles, active_tanks);
            d.hide();
        },
    });

    // Pre-populate cash summary in dialog description
    const expected_cash = calculate_expected_cash(cash_events, pos_totals, drive_offs);
    d.$wrapper.find("[data-fieldname='cash_summary_html']").html(
        render_cash_summary_html(cash_events, pos_totals, drive_offs, expected_cash)
    );

    // Live variance on cash count input
    d.fields_dict.actual_cash_count.df.onchange = function() {
        const actual = parseFloat(d.get_value("actual_cash_count") || 0);
        const variance = actual - expected_cash;
        const colour = Math.abs(variance) <= 50 ? "green" : Math.abs(variance) <= 200 ? "orange" : "red";
        d.$wrapper.find("#cash-variance-display").html(
            `<span class="indicator ${colour}">` +
            `Variance: KES ${variance >= 0 ? "+" : ""}${variance.toFixed(2)} ` +
            `[${classify_cash_variance(variance)}]</span>`
        );
    };

    d.show();
}


function build_close_dialog_fields(active_nozzles, active_tanks, opening_readings) {
    const fields = [];

    // --- Meter Readings section ---
    fields.push({
        fieldtype: "Section Break",
        label:     "Closing Meter Readings",
        collapsible: 0,
    });

    // Group nozzles by pump
    const pumps = {};
    active_nozzles.forEach(n => {
        const key = n.pump;
        if (!pumps[key]) pumps[key] = { pump_number: n.pump_number, pump_name: n.pump_name, nozzles: [] };
        pumps[key].nozzles.push(n);
    });

    Object.values(pumps).forEach(pump => {
        fields.push({
            fieldtype: "HTML",
            options: `<h6 style="margin:8px 0 4px">Pump ${pump.pump_number} — ${pump.pump_name}</h6>`,
        });

        pump.nozzles.forEach(nozzle => {
            // Get opening readings for this nozzle
            const open_ev = opening_readings.find(r =>
                r.pump === nozzle.pump && r.nozzle_number === nozzle.nozzle_number && r.meter_type === "Electronic Volume");
            const open_ec = opening_readings.find(r =>
                r.pump === nozzle.pump && r.nozzle_number === nozzle.nozzle_number && r.meter_type === "Electronic Cash");
            const open_mm = opening_readings.find(r =>
                r.pump === nozzle.pump && r.nozzle_number === nozzle.nozzle_number && r.meter_type === "Manual Mechanical");

            const prefix = `p${pump.pump_number}_n${nozzle.nozzle_number}`;

            fields.push(
                { fieldtype: "Column Break" },
                {
                    fieldtype: "Float",
                    fieldname: `${prefix}_ev`,
                    label:     `P${pump.pump_number}/N${nozzle.nozzle_number} Elec Vol (L)`,
                    description: `Opening: ${open_ev ? open_ev.totalizer_value.toLocaleString() : "—"}`,
                    precision: 3,
                },
                {
                    fieldtype: "Float",
                    fieldname: `${prefix}_ec`,
                    label:     `P${pump.pump_number}/N${nozzle.nozzle_number} Elec Cash (KES)`,
                    description: `Opening: ${open_ec ? open_ec.totalizer_value.toLocaleString() : "—"}`,
                    precision: 2,
                },
                {
                    fieldtype: "Float",
                    fieldname: `${prefix}_mm`,
                    label:     `P${pump.pump_number}/N${nozzle.nozzle_number} Manual Mech (L)`,
                    description: `Opening: ${open_mm ? open_mm.totalizer_value.toLocaleString() : "—"}`,
                    precision: 1,
                },
            );
        });
        fields.push({ fieldtype: "Section Break" });
    });

    // --- Dip Readings section ---
    fields.push({
        fieldtype: "Section Break",
        label:     "Closing Dip Readings",
        collapsible: 0,
    });

    active_tanks.forEach(tank => {
        fields.push({
            fieldtype: "Float",
            fieldname: `dip_${tank.tank.replace(/[^a-z0-9]/gi, "_")}`,
            label:     `${tank.tank} Closing Dip (L)`,
            description: `Product: ${tank.fuel_product}`,
            precision: 1,
        });
    });

    // --- Cash section ---
    fields.push(
        { fieldtype: "Section Break", label: "Cash Reconciliation" },
        { fieldtype: "HTML", fieldname: "cash_summary_html", options: "<div id='cash-summary'></div>" },
        { fieldtype: "Currency", fieldname: "actual_cash_count", label: "Actual Cash Count (KES)", reqd: 1 },
        { fieldtype: "HTML", options: "<div id='cash-variance-display'></div>" },
    );

    return fields;
}


function calculate_expected_cash(cash_events, pos_totals, drive_offs) {
    return (
        (cash_events.float_issued || 0)
        + ((pos_totals.by_payment || {})["Cash"] || 0)
        - (cash_events.cash_pickups || 0)
        - (cash_events.payouts || 0)
        - (drive_offs || 0)
    );
}


function render_cash_summary_html(cash_events, pos_totals, drive_offs, expected_cash) {
    const fmt = v => `KES ${Number(v).toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ",")}`;
    return `
        <table class="table table-condensed" style="font-size:0.9em">
            <tr><td>Opening Float</td><td class="text-right">${fmt(cash_events.float_issued || 0)}</td></tr>
            <tr><td>+ Cash Sales (POS)</td><td class="text-right">+ ${fmt((pos_totals.by_payment || {})["Cash"] || 0)}</td></tr>
            <tr><td>− Cash Pickups</td><td class="text-right">− ${fmt(cash_events.cash_pickups || 0)}</td></tr>
            <tr><td>− Payouts</td><td class="text-right">− ${fmt(cash_events.payouts || 0)}</td></tr>
            <tr><td>− Drive-Offs</td><td class="text-right">− ${fmt(drive_offs || 0)}</td></tr>
            <tr class="active"><td><strong>= Expected Cash</strong></td><td class="text-right"><strong>${fmt(expected_cash)}</strong></td></tr>
        </table>
        <div id="cash-variance-display"></div>
    `;
}


function classify_cash_variance(variance) {
    const abs_v = Math.abs(variance);
    if (abs_v <= 50)  return "Balanced";
    if (abs_v <= 200) return "Elevated — Review";
    return "CRITICAL — Investigate";
}


// -----------------------------------------------------------------------
// Save closing readings to Meter Reading and Dip Reading doctypes
// -----------------------------------------------------------------------

async function save_closing_readings(frm, values, active_nozzles, active_tanks) {
    frappe.dom.freeze("Saving closing readings…");
    const errors = [];

    // Create Meter Reading documents
    for (const nozzle of active_nozzles) {
        const prefix = `p${nozzle.pump_number}_n${nozzle.nozzle_number}`;
        const readings = [
            { meter_type: "Electronic Volume", value: values[`${prefix}_ev`] },
            { meter_type: "Electronic Cash",   value: values[`${prefix}_ec`] },
            { meter_type: "Manual Mechanical", value: values[`${prefix}_mm`] },
        ];

        for (const { meter_type, value } of readings) {
            if (!value || value <= 0) continue;
            try {
                await create_meter_reading(frm.doc.name, nozzle.pump, nozzle.nozzle_number, meter_type, value);
            } catch (e) {
                errors.push(`Pump ${nozzle.pump_number} / N${nozzle.nozzle_number} ${meter_type}: ${e.message}`);
            }
        }
    }

    // Create Dip Reading documents
    for (const tank of active_tanks) {
        const field = `dip_${tank.tank.replace(/[^a-z0-9]/gi, "_")}`;
        const vol = values[field];
        if (!vol || vol <= 0) continue;
        try {
            await create_dip_reading(frm.doc.name, tank.tank, vol);
        } catch (e) {
            errors.push(`Dip ${tank.tank}: ${e.message}`);
        }
    }

    // Save actual cash count to Cashier Session
    if (values.actual_cash_count) {
        await frappe.call({
            method: "forecourt.utils.session_api.save_cash_count",
            args: {
                shift:      frm.doc.name,
                cash_count: values.actual_cash_count,
            },
        });
    }

    frappe.dom.unfreeze();

    if (errors.length) {
        frappe.msgprint({
            title:   "Some readings could not be saved",
            message: errors.join("<br>"),
            indicator: "orange",
        });
    } else {
        frappe.show_alert({ message: "Closing readings saved", indicator: "green" });
        frm.reload_doc();
    }
}


async function create_meter_reading(shift, pump, nozzle_number, meter_type, totalizer_value) {
    return frappe.call({
        method: "frappe.client.insert",
        args: {
            doc: {
                doctype:          "Meter Reading",
                shift,
                pump,
                nozzle_number,
                meter_type,
                totalizer_value,
                reading_position: "Shift Close",
                observed_at:      frappe.datetime.now_datetime(),
                read_by:          frappe.session.user,
            },
        },
    });
}


async function create_dip_reading(shift, tank, volume_observed_l) {
    return frappe.call({
        method: "frappe.client.insert",
        args: {
            doc: {
                doctype:           "Dip Reading",
                shift,
                tank,
                reading_type:      "Shift Close",
                volume_observed_l,
                observed_at:       frappe.datetime.now_datetime(),
                read_by:           frappe.session.user,
                source:            "Manual Dipstick",
            },
        },
    });
}


// -----------------------------------------------------------------------
// Meter Validation
// -----------------------------------------------------------------------

function run_meter_validation(frm) {
    frappe.dom.freeze("Running meter validation…");
    frappe.call({
        method: "forecourt.utils.meter.compute_meter_validation_results",
        args:   { shift_name: frm.doc.name },
        callback(r) {
            frappe.dom.unfreeze();
            if (!r.message) return;
            render_mvr_summary(frm, r.message);
        },
        error() {
            frappe.dom.unfreeze();
        }
    });
}


function render_mvr_summary(frm, results) {
    const rows = results.map(r => `
        <tr class="${r.overall_status === "Fail" ? "danger" : r.overall_status === "Warning" ? "warning" : ""}">
            <td>${r.pump}</td>
            <td>${r.nozzle_number}</td>
            <td>${r.elec_vol_sold.toFixed(3)} L</td>
            <td>KES ${r.elec_cash_sold.toFixed(2)}</td>
            <td>${r.mech_vol_sold.toFixed(1)} L</td>
            <td><span class="indicator ${status_colour(r.check_a_status)}">${r.check_a_status}</span></td>
            <td><span class="indicator ${status_colour(r.check_b_status)}">${r.check_b_status} (${r.check_b_divergence_pct.toFixed(3)}%)</span></td>
            <td><strong>${r.overall_status}</strong></td>
        </tr>
    `).join("");

    frappe.msgprint({
        title: "Meter Validation Results",
        message: `
            <table class="table table-bordered table-condensed" style="font-size:0.85em">
                <thead>
                    <tr><th>Pump</th><th>Nozzle</th><th>Elec Vol</th><th>Elec Cash</th>
                    <th>Mech Vol</th><th>Check A</th><th>Check B</th><th>Status</th></tr>
                </thead>
                <tbody>${rows}</tbody>
            </table>
        `,
        wide: true,
    });
}

function status_colour(status) {
    return { Pass: "green", Warning: "orange", Fail: "red", Critical: "red" }[status] || "grey";
}


// -----------------------------------------------------------------------
// Compute Reconciliation
// -----------------------------------------------------------------------

function compute_reconciliation(frm) {
    frappe.confirm(
        "Run the reconciliation computation? This will generate the Shift Reconciliation record.",
        () => {
            frappe.dom.freeze("Computing reconciliation…");
            frappe.call({
                method: "forecourt.forecourt.doctype.shift_reconciliation.reconciliation_engine.compute_reconciliation",
                args:   { shift_name: frm.doc.name },
                callback(r) {
                    frappe.dom.unfreeze();
                    if (r.message) {
                        frappe.show_alert({ message: `Reconciliation saved: ${r.message}`, indicator: "green" });
                        frm.reload_doc();
                    }
                },
                error() { frappe.dom.unfreeze(); }
            });
        }
    );
}


// -----------------------------------------------------------------------
// Approve and Post GL
// -----------------------------------------------------------------------

function approve_and_post_gl(frm) {
    const sr_name = frm.doc.__onload?.shift_reconciliation_name;
    if (!sr_name) {
        frappe.msgprint("No Shift Reconciliation found. Run Compute Reconciliation first.");
        return;
    }

    frappe.confirm(
        `Post the GL Journal Entry for shift ${frm.doc.name}? This action is irreversible.`,
        () => {
            frappe.dom.freeze("Posting GL Journal Entry…");
            frappe.call({
                method: "forecourt.utils.gl.post_shift_journal_entry",
                args:   { shift_reconciliation_name: sr_name },
                callback(r) {
                    frappe.dom.unfreeze();
                    if (r.message) {
                        frappe.show_alert({ message: `GL Posted: ${r.message}`, indicator: "green" });
                        frm.reload_doc();
                    }
                },
                error() { frappe.dom.unfreeze(); }
            });
        }
    );
}


// -----------------------------------------------------------------------
// Status indicator
// -----------------------------------------------------------------------

function display_status_indicator(frm) {
    const colours = {
        "Draft":             "grey",
        "Open":              "blue",
        "Readings Captured": "yellow",
        "Closing":           "orange",
        "Closed":            "green",
        "Disputed":          "red",
    };
    const colour = colours[frm.doc.status] || "grey";
    frm.dashboard.set_headline_alert(
        `<span class="indicator ${colour}"><span>Status: ${frm.doc.status}</span></span>`
    );
}
```

### 29.3 Dip Reading Form — Instant Wetstock Preview

```javascript
// forecourt/public/js/dip_reading.js
// When a Shift Close dip is entered, instantly preview the wetstock calculation.

frappe.ui.form.on("Dip Reading", {

    volume_observed_l(frm) {
        if (frm.doc.reading_type === "Shift Close" && frm.doc.shift && frm.doc.tank) {
            preview_wetstock(frm);
        }
    },

    reading_type(frm) {
        if (frm.doc.reading_type === "Shift Close") {
            preview_wetstock(frm);
        }
    }
});


function preview_wetstock(frm) {
    if (!frm.doc.volume_observed_l || frm.doc.volume_observed_l <= 0) return;

    frappe.call({
        method: "forecourt.utils.wetstock.compute_tank_wetstock",
        args: {
            shift_name: frm.doc.shift,
            tank:       frm.doc.tank,
            // Pass the actual_closing_l from the form (not yet saved)
            actual_closing_override: frm.doc.volume_observed_l,
        },
        callback(r) {
            if (!r.message) return;
            const ws = r.message;

            const colour = {
                "Normal":   "green",
                "Elevated": "orange",
                "Critical": "red",
                "Gain":     "blue",
            }[ws.classification] || "grey";

            const sign = ws.variance_l >= 0 ? "+" : "";
            frm.dashboard.set_headline_alert(`
                <span class="indicator ${colour}">
                    Wetstock: Opening ${ws.opening_stock_l.toFixed(1)} L 
                    + Deliveries ${ws.deliveries_l.toFixed(1)} L 
                    − Sales ${ws.elec_vol_sales_l.toFixed(3)} L 
                    = Theoretical ${ws.theoretical_closing_l.toFixed(1)} L 
                    | Actual ${ws.actual_closing_l.toFixed(1)} L 
                    | Variance <strong>${sign}${ws.variance_l.toFixed(1)} L 
                    (${sign}${ws.variance_pct.toFixed(3)}%)</strong> 
                    [${ws.classification}]
                </span>
            `);
        }
    });
}
```

---

## 30. Dashboard and Report Queries

### 30.1 Daily Shift Report — Script Report Query

```python
# forecourt/forecourt/report/daily_shift_report/daily_shift_report.py

import frappe
from frappe.utils import flt


def execute(filters=None):
    """
    Script Report: Daily Shift Report
    Filters: shift (Link → Forecourt Shift)
    """
    if not filters or not filters.get("shift"):
        frappe.throw("Please select a shift.")

    shift_name = filters["shift"]
    shift = frappe.get_doc("Forecourt Shift", shift_name)

    sr_name = frappe.db.get_value("Shift Reconciliation", {"shift": shift_name}, "name")
    if not sr_name:
        frappe.throw(f"No Shift Reconciliation found for shift {shift_name}. "
                     "Compute the reconciliation first.")
    sr = frappe.get_doc("Shift Reconciliation", sr_name)

    columns = _get_columns()
    data    = _get_data(shift, sr)
    return columns, data


def _get_columns():
    return [
        {"fieldname": "section",      "label": "Section",      "fieldtype": "Data",     "width": 200},
        {"fieldname": "description",  "label": "Description",  "fieldtype": "Data",     "width": 300},
        {"fieldname": "value",        "label": "Value",        "fieldtype": "Data",     "width": 200},
        {"fieldname": "status",       "label": "Status",       "fieldtype": "Data",     "width": 120},
    ]


def _get_data(shift, sr):
    rows = []

    # Shift Header
    _add_section(rows, "SHIFT INFORMATION")
    _add_row(rows, "Shift Reference",  shift.name)
    _add_row(rows, "Date",             str(shift.shift_date))
    _add_row(rows, "Label",            shift.shift_label or "—")
    _add_row(rows, "Opened At",        str(shift.opened_at))
    _add_row(rows, "Closed At",        str(shift.closed_at or "—"))
    _add_row(rows, "Cashier",          shift.cashier)
    _add_row(rows, "Supervisor",       shift.supervisor)

    # Revenue Summary
    _add_section(rows, "REVENUE SUMMARY")
    for ps in sr.product_summaries:
        _add_row(rows,
            f"  {ps.fuel_product}",
            f"{flt(ps.volume_sold_l):.3f} L × KES {flt(ps.shift_rate):.2f}",
            f"KES {flt(ps.gross_revenue):,.2f}"
        )
    _add_row(rows, "TOTAL GROSS REVENUE", "", f"KES {flt(sr.total_gross_revenue):,.2f}")

    # Tank Wetstock
    _add_section(rows, "WETSTOCK RECONCILIATION")
    for ts in sr.tank_wetstock_summaries:
        var_sign = "+" if flt(ts.variance_l) >= 0 else ""
        _add_row(rows,
            f"  {ts.tank}",
            f"Open {flt(ts.opening_stock_l):.1f} + Del {flt(ts.deliveries_l):.1f} "
            f"− Sales {flt(ts.elec_vol_sales_l):.3f} = Theoretical {flt(ts.theoretical_closing_l):.1f} "
            f"| Actual {flt(ts.actual_closing_l):.1f}",
            f"{var_sign}{flt(ts.variance_l):.1f} L ({var_sign}{flt(ts.variance_pct):.3f}%)",
            ts.classification,
        )

    # Meter Validation
    _add_section(rows, "METER VALIDATION")
    for nm in sr.nozzle_meter_summary:
        _add_row(rows,
            f"  Pump {nm.pump} / Nozzle {nm.nozzle}",
            f"Elec Vol {flt(nm.elec_vol_sold):.3f} L | Elec Cash KES {flt(nm.elec_cash_sold):,.2f} "
            f"| Mech {flt(nm.mech_vol_sold):.1f} L",
            f"A:{nm.check_a_status} B:{nm.check_b_status}",
        )

    # Cash Reconciliation
    _add_section(rows, "CASH RECONCILIATION")
    _add_row(rows, "Expected Cash",   "", f"KES {flt(sr.expected_cash):,.2f}")
    _add_row(rows, "Actual Cash",     "", f"KES {flt(sr.actual_cash):,.2f}")
    cv = flt(sr.cash_variance)
    _add_row(rows, "Variance",        "",
        f"KES {cv:+,.2f}",
        sr.cash_variance_status)

    # Non-Cash Tenders
    _add_section(rows, "NON-CASH TENDERS")
    _add_row(rows, "MPesa",        "", f"KES {flt(sr.mpesa_total):,.2f}")
    _add_row(rows, "Card",         "", f"KES {flt(sr.card_total):,.2f}")
    _add_row(rows, "Fleet/Credit", "", f"KES {flt(sr.fleet_credit_total):,.2f}")

    # Verdict
    _add_section(rows, "VERDICT")
    _add_row(rows, "Meters OK",      "", "", "✓" if sr.is_meter_ok else "⚠")
    _add_row(rows, "Wetstock OK",    "", "", "✓" if sr.is_wetstock_balanced else "⚠")
    _add_row(rows, "Cash Balanced",  "", "", "✓" if sr.is_cash_balanced else "⚠")
    _add_row(rows, "GL Journal",     "", sr.gl_journal or "Not posted")

    return rows


def _add_section(rows, label):
    rows.append({"section": f"── {label}", "description": "", "value": "", "status": ""})

def _add_row(rows, section, description="", value="", status=""):
    rows.append({"section": section, "description": description, "value": value, "status": status})
```

### 30.2 Wetstock Variance Trend — SQL Report

```python
# forecourt/forecourt/report/wetstock_variance_trend/wetstock_variance_trend.py

import frappe


def execute(filters=None):
    filters = filters or {}

    columns = [
        {"fieldname": "shift_date",    "label": "Date",          "fieldtype": "Date",     "width": 100},
        {"fieldname": "shift_label",   "label": "Shift",         "fieldtype": "Data",     "width": 80},
        {"fieldname": "tank",          "label": "Tank",          "fieldtype": "Data",     "width": 200},
        {"fieldname": "fuel_product",  "label": "Product",       "fieldtype": "Data",     "width": 120},
        {"fieldname": "opening_l",     "label": "Opening L",     "fieldtype": "Float",    "width": 100},
        {"fieldname": "deliveries_l",  "label": "Deliveries L",  "fieldtype": "Float",    "width": 100},
        {"fieldname": "sales_l",       "label": "Sales L",       "fieldtype": "Float",    "width": 100},
        {"fieldname": "theoretical_l", "label": "Theoretical L", "fieldtype": "Float",    "width": 110},
        {"fieldname": "actual_l",      "label": "Actual L",      "fieldtype": "Float",    "width": 90},
        {"fieldname": "variance_l",    "label": "Variance L",    "fieldtype": "Float",    "width": 100},
        {"fieldname": "variance_pct",  "label": "Variance %",    "fieldtype": "Percent",  "width": 100},
        {"fieldname": "classification","label": "Class",         "fieldtype": "Data",     "width": 90},
        {"fieldname": "variance_kes",  "label": "Variance KES",  "fieldtype": "Currency", "width": 120},
    ]

    conditions = []
    params = {}

    if filters.get("from_date"):
        conditions.append("fs.shift_date >= %(from_date)s")
        params["from_date"] = filters["from_date"]
    if filters.get("to_date"):
        conditions.append("fs.shift_date <= %(to_date)s")
        params["to_date"] = filters["to_date"]
    if filters.get("tank"):
        conditions.append("tws.tank = %(tank)s")
        params["tank"] = filters["tank"]

    where = "WHERE " + " AND ".join(conditions) if conditions else ""

    data = frappe.db.sql(
        f"""
        SELECT
            fs.shift_date,
            fs.shift_label,
            tws.tank,
            tws.fuel_product,
            tws.opening_stock_l   AS opening_l,
            tws.deliveries_l,
            tws.elec_vol_sales_l  AS sales_l,
            tws.theoretical_closing_l AS theoretical_l,
            tws.actual_closing_l  AS actual_l,
            tws.variance_l,
            tws.variance_pct,
            tws.classification,
            tws.variance_kes
        FROM `tabShift Reconciliation Tank Wetstock` tws
        JOIN `tabShift Reconciliation` sr ON sr.name = tws.parent
        JOIN `tabForecourt Shift` fs ON fs.name = sr.shift
        {where}
        ORDER BY fs.shift_date DESC, fs.shift_label, tws.tank
        """,
        params,
        as_dict=True,
    )

    return columns, data
```

### 30.3 Cashier Performance Report

```python
# forecourt/forecourt/report/cashier_performance/cashier_performance.py

import frappe
from frappe.utils import flt


def execute(filters=None):
    filters = filters or {}

    columns = [
        {"fieldname": "cashier",           "label": "Cashier",          "fieldtype": "Link", "options": "Employee", "width": 160},
        {"fieldname": "shift_count",       "label": "Shifts",           "fieldtype": "Int",      "width": 70},
        {"fieldname": "total_cash_kes",    "label": "Total Cash (KES)",  "fieldtype": "Currency", "width": 150},
        {"fieldname": "total_variance",    "label": "Net Variance",      "fieldtype": "Currency", "width": 130},
        {"fieldname": "avg_variance",      "label": "Avg Variance/Shift","fieldtype": "Currency", "width": 150},
        {"fieldname": "short_count",       "label": "Short Count",       "fieldtype": "Int",      "width": 100},
        {"fieldname": "over_count",        "label": "Over Count",        "fieldtype": "Int",      "width": 100},
        {"fieldname": "largest_variance",  "label": "Largest Variance",  "fieldtype": "Currency", "width": 140},
        {"fieldname": "drive_offs_kes",    "label": "Drive-Offs (KES)",  "fieldtype": "Currency", "width": 130},
    ]

    data = frappe.db.sql(
        """
        SELECT
            fs.cashier,
            COUNT(DISTINCT sr.shift)           AS shift_count,
            SUM(sr.actual_cash)                AS total_cash_kes,
            SUM(sr.cash_variance)              AS total_variance,
            AVG(sr.cash_variance)              AS avg_variance,
            SUM(CASE WHEN sr.cash_variance < 0 THEN 1 ELSE 0 END) AS short_count,
            SUM(CASE WHEN sr.cash_variance > 0 THEN 1 ELSE 0 END) AS over_count,
            MAX(ABS(sr.cash_variance))         AS largest_variance,
            COALESCE((
                SELECT SUM(d.amount_kes)
                FROM `tabDrive-Off Record` d
                WHERE d.shift = fs.name AND d.docstatus = 1
            ), 0)                              AS drive_offs_kes
        FROM `tabShift Reconciliation` sr
        JOIN `tabForecourt Shift` fs ON fs.name = sr.shift
        GROUP BY fs.cashier
        ORDER BY ABS(SUM(sr.cash_variance)) DESC
        """,
        as_dict=True,
    )

    return columns, data
```

---

## 31. hooks.py — Complete Registration

```python
# forecourt/hooks.py

app_name    = "forecourt"
app_title   = "Forecourt Management"
app_publisher = "Anika Global Limited"
app_description = "Petrol station forecourt management for ERPNext"
app_email   = ""
app_license = "Proprietary"

# Jinja templates
jinja = {}

# DocType event hooks
doc_events = {
    "Forecourt Pump": {
        "validate": "forecourt.forecourt.doctype.forecourt_pump.forecourt_pump.validate",
    },
    "Forecourt Shift": {
        "validate":    "forecourt.forecourt.doctype.forecourt_shift.forecourt_shift.validate",
        "before_save": "forecourt.forecourt.doctype.forecourt_shift.forecourt_shift.before_save",
    },
    "Meter Reading": {
        "validate": "forecourt.forecourt.doctype.meter_reading.meter_reading.validate",
    },
    "Dip Reading": {
        "validate": "forecourt.forecourt.doctype.dip_reading.dip_reading.validate",
    },
    "Cash Event": {
        "validate": "forecourt.forecourt.doctype.cash_event.cash_event.validate",
    },
    # POS Invoice — block submission until shift opens; link to shift on submit
    "POS Invoice": {
        "before_submit": "forecourt.overrides.pos_invoice.before_submit",
        "on_submit":     "forecourt.overrides.pos_invoice.on_submit",
    },
    # Sales Invoice — link to open shift on submit (fleet/credit customers)
    "Sales Invoice": {
        "before_submit": "forecourt.overrides.sales_invoice.before_submit",
    },
}

# Whitelisted methods (callable from client scripts and API)
whitelisted_methods = [
    "forecourt.forecourt.doctype.forecourt_shift.pre_fetch.get_shift_close_prefetch",
    "forecourt.utils.meter.compute_meter_validation_results",
    "forecourt.utils.wetstock.compute_tank_wetstock",
    "forecourt.forecourt.doctype.shift_reconciliation.reconciliation_engine.compute_reconciliation",
    "forecourt.utils.gl.post_shift_journal_entry",
    "forecourt.utils.buying.create_purchase_receipt_for_delivery",
    "forecourt.utils.session_api.save_cash_count",
    "forecourt.utils.meter_api.get_nozzle_shift_rate",
]

# JS and CSS assets
app_include_js  = ["/assets/forecourt/js/forecourt.bundle.js"]
app_include_css = ["/assets/forecourt/css/forecourt.css"]

# Doctype-specific JS
doctype_js = {
    "Forecourt Shift":  "public/js/forecourt_shift.js",
    "Meter Reading":    "public/js/meter_reading.js",
    "Dip Reading":      "public/js/dip_reading.js",
    "Cash Event":       "public/js/cash_event.js",
    "Shift Reconciliation": "public/js/shift_reconciliation.js",
}

# Scheduled tasks
scheduler_events = {
    "daily": [
        # Email shift report to management each morning
        "forecourt.tasks.send_daily_shift_summary",
        # Check for pumps with calibration due within 30 days
        "forecourt.tasks.check_calibration_due_dates",
    ],
    "hourly": [
        # Publish realtime alerts if an open shift has been idle > 14 hours
        "forecourt.tasks.check_stale_open_shifts",
    ],
}

# Fixtures — export these with bench export-fixtures
fixtures = [
    {"dt": "Custom Field",   "filters": [["dt", "in", [
        "POS Invoice", "Sales Invoice", "Warehouse"
    ]]]},
    {"dt": "Property Setter", "filters": [["doc_type", "in", [
        "POS Invoice", "Sales Invoice"
    ]]]},
    {"dt": "Role",            "filters": [["name", "in", [
        "Forecourt Manager", "Forecourt Supervisor", "Cashier", "Pump Attendant"
    ]]]},
]

# Custom roles required by this app
roles_to_create = [
    {"role_name": "Forecourt Manager",    "desk_access": 1},
    {"role_name": "Forecourt Supervisor", "desk_access": 1},
    {"role_name": "Cashier",              "desk_access": 1},
    {"role_name": "Pump Attendant",       "desk_access": 0},
]
```

---

*End of Integration Architecture Addendum — Version 3.1.0*
*Prepared for: Shell Maanzoni Service Station (Anika Global Limited)*
*Stack: ERPNext 15 / Frappe Framework / Python 3.11 / JavaScript ES2020*

---

## 32. Python — Sales Invoice Override (Fleet / Credit Customers)

Fleet and credit customers pay on account via Sales Invoice, not POS. This override links every submitted Sales Invoice to the open shift so the reconciliation engine can aggregate fleet/credit totals alongside cash and MPesa.

```python
# forecourt/overrides/sales_invoice.py

"""
Override for Sales Invoice — links fleet/credit invoices to the open
Forecourt Shift so the reconciliation engine can include them in the
non-cash tender totals.

Only applies when the invoice contains fuel items. Shop and lubricant
invoices that do not contain fuel items are not linked to a shift.

Register in hooks.py:
    doc_events = {
        "Sales Invoice": {
            "before_submit": "forecourt.overrides.sales_invoice.before_submit",
        }
    }
"""

import frappe
from forecourt.overrides.pos_invoice import FUEL_ITEM_CODES, _get_open_shift


def before_submit(doc, method=None):
    """
    Link the Sales Invoice to the open Forecourt Shift.
    Block submission if no shift is open and the invoice contains fuel.
    """
    if not _invoice_contains_fuel(doc):
        return  # Non-fuel invoices (shop items) are not controlled by the shift

    company = doc.company or frappe.defaults.get_global_default("company")
    open_shift = _get_open_shift(company)

    if not open_shift:
        frappe.throw(
            "No open Forecourt Shift found. A supervisor must open a shift before "
            "fuel credit invoices can be submitted.",
            title="Shift Not Open",
        )

    doc.forecourt_shift = open_shift

    # Validate that this customer has a credit account set up
    customer_type = frappe.get_cached_value("Customer", doc.customer, "customer_type")
    if customer_type not in ("Company", "Individual"):
        frappe.throw(
            f"Customer {doc.customer} does not have a valid customer type configured. "
            "Set Customer Type before raising fleet credit invoices.",
            title="Invalid Credit Customer",
        )

    # Log for audit trail
    frappe.logger().info(
        f"Sales Invoice {doc.name} linked to shift {open_shift} "
        f"(customer: {doc.customer}, total: {doc.grand_total})"
    )


def _invoice_contains_fuel(doc) -> bool:
    return any(item.item_code in FUEL_ITEM_CODES for item in doc.items)
```

---

## 33. Python — API Helpers Called from Client Scripts

These are the `@frappe.whitelist()` methods invoked from JavaScript. Each is small, focused, and defensively validated.

### 33.1 session_api.py — Save Cash Count

```python
# forecourt/utils/session_api.py

import frappe
from frappe.utils import flt, now_datetime


@frappe.whitelist()
def save_cash_count(shift: str, cash_count: float) -> dict:
    """
    Save the actual cash count to the primary Cashier Session for a shift.
    Also sets counted_by to the current user's linked Employee.

    Called from the shift close dialog after the supervisor physically
    counts the till.

    Args:
        shift:      Forecourt Shift name
        cash_count: Physically counted cash in KES

    Returns:
        dict: {"session": session_name, "cash_count": cash_count}
    """
    cash_count = flt(cash_count)
    if cash_count < 0:
        frappe.throw("Cash count cannot be negative.", title="Invalid Cash Count")

    session_name = frappe.db.get_value(
        "Cashier Session",
        {"shift": shift, "is_primary": 1},
        "name",
    )
    if not session_name:
        frappe.throw(
            f"No primary Cashier Session found for shift {shift}. "
            "Create a Cashier Session before saving the cash count.",
            title="Session Not Found",
        )

    # Look up the Employee linked to the current user
    counted_by = frappe.db.get_value("Employee", {"user_id": frappe.session.user}, "name")

    frappe.db.set_value(
        "Cashier Session",
        session_name,
        {
            "actual_cash_close":  cash_count,
            "session_closed_at":  now_datetime(),
            "counted_by":         counted_by,
        },
    )
    frappe.db.commit()

    frappe.logger().info(
        f"Cash count KES {cash_count:,.2f} saved to session {session_name} "
        f"by {frappe.session.user}"
    )
    return {"session": session_name, "cash_count": cash_count}


@frappe.whitelist()
def open_cashier_session(shift: str, cashier: str, float_amount: float, till_id: str = "TILL-01") -> str:
    """
    Create and save a primary Cashier Session for a shift.
    Also creates the Float Issued Cash Event.

    Args:
        shift:        Forecourt Shift name
        cashier:      Employee name
        float_amount: Opening float in KES
        till_id:      Till identifier (default TILL-01)

    Returns:
        str: Cashier Session name
    """
    from forecourt.utils.hr import assert_is_active_employee, assert_employees_differ

    float_amount = flt(float_amount)
    supervisor = frappe.get_cached_value("Forecourt Shift", shift, "supervisor")

    assert_is_active_employee(cashier, role_label="Cashier")
    assert_employees_differ(cashier, supervisor, context="Cashier Session open")

    if float_amount <= 0:
        frappe.throw("Float amount must be positive.", title="Invalid Float")

    # Prevent duplicate sessions
    existing = frappe.db.exists("Cashier Session", {"shift": shift, "is_primary": 1})
    if existing:
        frappe.throw(
            f"Primary Cashier Session {existing} already exists for shift {shift}.",
            title="Duplicate Session",
        )

    session = frappe.new_doc("Cashier Session")
    session.shift           = shift
    session.cashier         = cashier
    session.till_id         = till_id
    session.is_primary      = 1
    session.float_amount    = float_amount
    session.session_opened_at = now_datetime()
    session.insert(ignore_permissions=True)

    # Create Float Issued Cash Event
    event = frappe.new_doc("Cash Event")
    event.shift           = shift
    event.cashier_session = session.name
    event.event_type      = "Float Issued"
    event.amount          = float_amount
    event.authorised_by   = supervisor
    event.occurred_at     = now_datetime()
    event.reference       = f"Float — {shift}"
    event.insert(ignore_permissions=True)
    event.submit()

    frappe.logger().info(
        f"Cashier Session {session.name} opened for {cashier} on shift {shift}, "
        f"float KES {float_amount:,.2f}"
    )
    return session.name
```

### 33.2 meter_api.py — Client Helpers

```python
# forecourt/utils/meter_api.py

import frappe
from frappe.utils import flt


@frappe.whitelist()
def get_nozzle_shift_rate(shift: str, pump: str) -> float:
    """
    Return the EPRA shift rate for the fuel product dispensed by a given pump.
    Used by the Meter Reading form JS to show the expected cash cross-check.

    Args:
        shift: Forecourt Shift name
        pump:  Forecourt Pump name

    Returns:
        float: KES per litre
    """
    # Find the fuel product for the first active nozzle on this pump
    fuel_product = frappe.db.get_value(
        "Pump Nozzle",
        {"parent": pump, "is_active": 1},
        "fuel_product",
        order_by="nozzle_number asc",
    )
    if not fuel_product:
        return 0.0

    rate_map = {
        "FUEL-PMS-UNL": "rate_pms_unl",
        "FUEL-PMS-VP":  "rate_pms_vp",
        "FUEL-AGO":     "rate_ago",
        "FUEL-DPK":     "rate_dpk",
    }
    field = rate_map.get(fuel_product)
    if not field:
        return 0.0
    return flt(frappe.get_cached_value("Forecourt Shift", shift, field))


@frappe.whitelist()
def get_opening_reading(shift: str, pump: str, nozzle_number: int, meter_type: str) -> dict:
    """
    Fetch the submitted Shift Open reading for a specific nozzle and meter type.
    Used by client scripts to display the opening value next to the closing input.

    Returns:
        dict: {totalizer_value, unit, observed_at} or {} if not found
    """
    result = frappe.db.get_value(
        "Meter Reading",
        {
            "shift":            shift,
            "pump":             pump,
            "nozzle_number":    int(nozzle_number),
            "meter_type":       meter_type,
            "reading_position": "Shift Open",
            "docstatus":        1,
        },
        ["totalizer_value", "unit", "observed_at"],
        as_dict=True,
    )
    return result or {}


@frappe.whitelist()
def get_shift_meter_summary(shift: str) -> list:
    """
    Return a per-nozzle meter summary showing opening readings and any
    submitted closing readings. Used by the shift close form to show
    which readings still need to be entered.

    Returns:
        list of dicts: one per active nozzle, with open/close/delta for all 3 types
    """
    active_nozzles = frappe.db.sql(
        """
        SELECT
            fp.name        AS pump,
            fp.pump_number,
            pn.nozzle_number,
            pn.fuel_product,
            pn.tank
        FROM `tabForecourt Pump` fp
        JOIN `tabPump Nozzle` pn ON pn.parent = fp.name
        WHERE fp.is_active = 1 AND pn.is_active = 1
        ORDER BY fp.pump_number, pn.nozzle_number
        """,
        as_dict=True,
    )

    meter_types = ["Electronic Volume", "Electronic Cash", "Manual Mechanical"]
    summary = []

    for nozzle in active_nozzles:
        nozzle_data = {
            "pump":          nozzle.pump,
            "pump_number":   nozzle.pump_number,
            "nozzle_number": nozzle.nozzle_number,
            "fuel_product":  nozzle.fuel_product,
            "readings":      {},
        }
        for mt in meter_types:
            opening = frappe.db.get_value(
                "Meter Reading",
                {"shift": shift, "pump": nozzle.pump, "nozzle_number": nozzle.nozzle_number,
                 "meter_type": mt, "reading_position": "Shift Open", "docstatus": 1},
                "totalizer_value",
            )
            closing = frappe.db.get_value(
                "Meter Reading",
                {"shift": shift, "pump": nozzle.pump, "nozzle_number": nozzle.nozzle_number,
                 "meter_type": mt, "reading_position": "Shift Close", "docstatus": 1},
                "totalizer_value",
            )
            delta = (flt(closing) - flt(opening)) if (opening is not None and closing is not None) else None
            nozzle_data["readings"][mt] = {
                "opening": flt(opening) if opening is not None else None,
                "closing": flt(closing) if closing is not None else None,
                "delta":   delta,
                "complete": opening is not None and closing is not None,
            }
        summary.append(nozzle_data)

    return summary
```

---

## 34. Python — Scheduled Tasks

```python
# forecourt/tasks.py

"""
Scheduled tasks registered in hooks.py scheduler_events.
These run silently in the background via the Frappe scheduler.
"""

import frappe
from frappe.utils import today, add_days, date_diff, now_datetime, time_diff_in_hours


# ---------------------------------------------------------------------------
# Daily: Email Shift Summary to Management
# ---------------------------------------------------------------------------

def send_daily_shift_summary():
    """
    At midnight: email the closed shift summary from yesterday to the
    management distribution list configured in Forecourt Site Preferences.
    Skips gracefully if no shifts were closed yesterday.
    """
    company = frappe.defaults.get_global_default("company")
    if not company:
        return

    yesterday = add_days(today(), -1)

    closed_shifts = frappe.db.get_all(
        "Forecourt Shift",
        filters={"shift_date": yesterday, "status": "Closed"},
        fields=["name", "shift_label", "cashier", "supervisor"],
    )

    if not closed_shifts:
        frappe.logger().info(f"No closed shifts on {yesterday} — skipping summary email.")
        return

    # Get recipient list from Site Preferences
    try:
        prefs = frappe.get_doc("Forecourt Site Preferences", company)
        recipients = [r.email for r in (prefs.report_recipients or []) if r.email]
    except frappe.DoesNotExistError:
        frappe.logger().warning("Forecourt Site Preferences not found — skipping email.")
        return

    if not recipients:
        return

    for shift in closed_shifts:
        sr_name = frappe.db.get_value("Shift Reconciliation", {"shift": shift.name}, "name")
        if not sr_name:
            continue
        sr = frappe.get_doc("Shift Reconciliation", sr_name)
        subject = f"[Shell Maanzoni] Shift Report — {yesterday} {shift.shift_label or ''} — {sr.cash_variance_status}"
        body = _build_summary_email_body(shift, sr)

        frappe.sendmail(
            recipients=recipients,
            subject=subject,
            message=body,
            delayed=False,
        )
        frappe.logger().info(f"Shift summary email sent for {shift.name} to {recipients}")


def _build_summary_email_body(shift: dict, sr) -> str:
    def fmt(v):
        return f"KES {float(v or 0):,.2f}"

    cv = float(sr.cash_variance or 0)
    cv_str = f"KES {cv:+,.2f}"
    ws_l = float(sr.wetstock_variance_l or 0)
    ws_kes = float(sr.wetstock_variance_kes or 0)

    return f"""
    <h3>Shell Maanzoni — Daily Shift Report</h3>
    <table style="border-collapse:collapse;font-family:monospace;font-size:13px">
      <tr><td><b>Shift</b></td><td>{shift.name}</td></tr>
      <tr><td><b>Label</b></td><td>{shift.shift_label or '—'}</td></tr>
      <tr><td><b>Cashier</b></td><td>{shift.cashier}</td></tr>
      <tr><td><b>Supervisor</b></td><td>{shift.supervisor}</td></tr>
      <tr><td colspan=2><hr></td></tr>
      <tr><td><b>Total Volume</b></td><td>{float(sr.total_volume_l or 0):.3f} L</td></tr>
      <tr><td><b>Gross Revenue</b></td><td>{fmt(sr.total_gross_revenue)}</td></tr>
      <tr><td><b>MPesa</b></td><td>{fmt(sr.mpesa_total)}</td></tr>
      <tr><td><b>Card</b></td><td>{fmt(sr.card_total)}</td></tr>
      <tr><td><b>Fleet/Credit</b></td><td>{fmt(sr.fleet_credit_total)}</td></tr>
      <tr><td colspan=2><hr></td></tr>
      <tr><td><b>Cash Variance</b></td>
          <td style="color:{'green' if abs(cv)<=50 else 'red'}">{cv_str} [{sr.cash_variance_status}]</td></tr>
      <tr><td><b>Wetstock Variance</b></td>
          <td style="color:{'green' if abs(ws_l)<50 else 'red'}">{ws_l:+.1f} L ({fmt(ws_kes)})</td></tr>
      <tr><td><b>Meters OK</b></td><td>{'✓' if sr.is_meter_ok else '⚠ CHECK MVR'}</td></tr>
      <tr><td><b>GL Journal</b></td><td>{sr.gl_journal or 'Not posted'}</td></tr>
    </table>
    """


# ---------------------------------------------------------------------------
# Daily: Calibration Due Date Check
# ---------------------------------------------------------------------------

def check_calibration_due_dates():
    """
    Flag pumps where calibration is due within 30 days or overdue.
    Creates a Frappe Notification for the Forecourt Manager role.
    """
    pumps = frappe.db.get_all(
        "Forecourt Pump",
        filters={"is_active": 1, "next_calibration_due": ("is", "set")},
        fields=["name", "pump_number", "pump_name", "next_calibration_due"],
    )

    for pump in pumps:
        days = date_diff(pump.next_calibration_due, today())
        if days <= 0:
            _notify_calibration(pump, "OVERDUE", "red")
        elif days <= 30:
            _notify_calibration(pump, f"Due in {days} day(s)", "orange")


def _notify_calibration(pump: dict, message: str, colour: str):
    notification_key = f"calibration_{pump.name}_{today()}"
    if frappe.cache().get(notification_key):
        return  # Already notified today

    frappe.publish_realtime(
        event="forecourt_alert",
        message={
            "type":    "calibration",
            "pump":    pump.pump_name,
            "number":  pump.pump_number,
            "due":     str(pump.next_calibration_due),
            "status":  message,
            "colour":  colour,
        },
    )
    frappe.cache().set(notification_key, 1, expires_in_sec=86400)
    frappe.logger().warning(
        f"Calibration alert — Pump {pump.pump_number} ({pump.pump_name}): {message}"
    )


# ---------------------------------------------------------------------------
# Hourly: Stale Open Shift Check
# ---------------------------------------------------------------------------

def check_stale_open_shifts():
    """
    Alert management if any shift has been Open for more than 14 hours
    without a shift close being initiated. This catches forgotten open
    shifts that would corrupt the next shift's meter deltas.
    """
    stale_threshold_hours = 14

    open_shifts = frappe.db.get_all(
        "Forecourt Shift",
        filters={"status": ("in", ["Open", "Closing"])},
        fields=["name", "cashier", "opened_at"],
    )

    for shift in open_shifts:
        if not shift.opened_at:
            continue
        hours_open = time_diff_in_hours(now_datetime(), shift.opened_at)
        if hours_open >= stale_threshold_hours:
            frappe.publish_realtime(
                event="forecourt_alert",
                message={
                    "type":       "stale_shift",
                    "shift":      shift.name,
                    "cashier":    shift.cashier,
                    "hours_open": round(hours_open, 1),
                    "message":    f"Shift {shift.name} has been open for "
                                  f"{round(hours_open, 1)} hours. "
                                  "Close or investigate immediately.",
                },
            )
            frappe.logger().error(
                f"STALE OPEN SHIFT: {shift.name} has been open "
                f"{round(hours_open, 1)} hours."
            )
```

---

## 35. Python — Drive-Off Record DocType

```python
# forecourt/forecourt/doctype/drive_off_record/drive_off_record.py

import frappe
from frappe.model.document import Document
from frappe.utils import flt


POLICE_REPORT_THRESHOLD_KES = 500.0


class DriveOffRecord(Document):

    def validate(self):
        self._validate_shift_is_open()
        self._compute_amount()
        self._warn_police_report()
        self._validate_authorisation()

    def on_submit(self):
        self._post_drive_off_journal()

    def _validate_shift_is_open(self):
        shift_status = frappe.get_cached_value("Forecourt Shift", self.shift, "status")
        if shift_status not in ("Open", "Closing"):
            frappe.throw(
                f"Cannot record a drive-off against shift {self.shift} "
                f"(status: {shift_status}). Only Open or Closing shifts accept drive-off records.",
                title="Invalid Shift Status",
            )

    def _compute_amount(self):
        """
        Auto-compute KES amount from volume × shift EPRA rate if not manually set.
        """
        if flt(self.volume_dispensed_l) <= 0:
            return

        if not flt(self.amount_kes):
            rate_map = {
                "FUEL-PMS-UNL": "rate_pms_unl",
                "FUEL-PMS-VP":  "rate_pms_vp",
                "FUEL-AGO":     "rate_ago",
                "FUEL-DPK":     "rate_dpk",
            }
            rate_field = rate_map.get(self.fuel_product)
            if rate_field:
                rate = flt(frappe.get_cached_value("Forecourt Shift", self.shift, rate_field))
                self.amount_kes = flt(self.volume_dispensed_l) * rate

    def _warn_police_report(self):
        if flt(self.amount_kes) >= POLICE_REPORT_THRESHOLD_KES and not self.police_reference:
            frappe.msgprint(
                f"Drive-off value KES {flt(self.amount_kes):,.2f} exceeds the KES "
                f"{POLICE_REPORT_THRESHOLD_KES:,.0f} threshold. "
                "A police report is recommended. Enter the police reference number once filed.",
                indicator="orange",
                title="Police Report Recommended",
                raise_exception=False,
            )

    def _validate_authorisation(self):
        if not self.authorised_by:
            frappe.throw(
                "Drive-off records must be authorised by a supervisor.",
                title="Authorisation Required",
            )
        # Supervisor cannot be the same as the attendant who witnessed the drive-off
        if self.authorised_by == self.attendant:
            frappe.throw(
                "The authorising supervisor cannot be the same person as the attendant. "
                "A different supervisor must authorise drive-off records.",
                title="Dual Control Violation",
            )

    def _post_drive_off_journal(self):
        """
        Post the accounting entry for the drive-off:
            DR Drive-Off Losses / CR Fuel Sales — [Product]

        Only posts if a linked POS Invoice does not already exist (to avoid
        double-posting if the invoice was already created and needs to be
        written off differently).
        """
        if self.pos_invoice:
            # POS Invoice already exists: the write-off is handled by cancelling
            # the invoice against Drive-Off Losses, not here.
            return

        from forecourt.utils.accounts import get_fuel_accounts, get_account
        company = frappe.get_cached_value("Forecourt Shift", self.shift, "company") \
                  or frappe.defaults.get_global_default("company")
        cost_center = frappe.get_cached_value("Company", company, "cost_center")
        accts = get_fuel_accounts(self.fuel_product, company)
        amount = flt(self.amount_kes)

        je = frappe.new_doc("Journal Entry")
        je.voucher_type  = "Journal Entry"
        je.posting_date  = frappe.utils.today()
        je.company       = company
        je.user_remark   = (
            f"Drive-off — {self.name} | Shift {self.shift} | "
            f"Pump {self.pump} | Vehicle {self.vehicle_description or 'Unknown'} | "
            f"{self.licence_plate or 'No plate'}"
        )
        je.append("accounts", {
            "account":                      get_account("drive_off_losses", company),
            "debit_in_account_currency":    amount,
            "credit_in_account_currency":   0,
            "cost_center":                  cost_center,
        })
        je.append("accounts", {
            "account":                      accts["sales"],
            "debit_in_account_currency":    0,
            "credit_in_account_currency":   amount,
            "cost_center":                  cost_center,
        })
        je.insert(ignore_permissions=True)
        je.submit()

        frappe.db.set_value("Drive-Off Record", self.name, "journal_entry", je.name)
        frappe.logger().info(
            f"Drive-off JE {je.name} posted for {self.name}, amount KES {amount:,.2f}"
        )
```

---

## 36. Python — Forecourt Site Preferences DocType

This Single doctype stores per-station configuration that the reconciliation engine, GL helpers, and scheduled tasks read at runtime. It avoids hardcoding account names, thresholds, and contacts in Python files.

```python
# forecourt/forecourt/doctype/forecourt_site_preferences/forecourt_site_preferences.py

import frappe
from frappe.model.document import Document


class ForecourtSitePreferences(Document):
    """
    Single DocType — one record per company.
    Name field = Company name (set on insert).
    """

    def autoname(self):
        self.name = self.company

    def validate(self):
        self._validate_thresholds()
        self._validate_recipients()

    def _validate_thresholds(self):
        """Ensure variance thresholds are sensible values."""
        checks = [
            ("wetstock_normal_pct",   0.10, 1.0,  "Wetstock Normal %"),
            ("wetstock_elevated_pct", 0.10, 2.0,  "Wetstock Elevated %"),
            ("cash_normal_kes",       0,    500,   "Cash Normal KES"),
            ("cash_elevated_kes",     0,    2000,  "Cash Elevated KES"),
        ]
        for field, min_v, max_v, label in checks:
            val = self.get(field)
            if val is not None and not (min_v <= float(val) <= max_v):
                frappe.throw(
                    f"{label} must be between {min_v} and {max_v}. Got: {val}",
                    title="Invalid Threshold",
                )

    def _validate_recipients(self):
        """Email addresses in report_recipients must be valid."""
        import re
        pattern = re.compile(r"^[^@]+@[^@]+\.[^@]+$")
        for row in (self.report_recipients or []):
            if row.email and not pattern.match(row.email):
                frappe.throw(
                    f"Invalid email address in report recipients: {row.email}",
                    title="Invalid Email",
                )
```

**DocType Field Schema (create via Frappe Developer Mode):**

```
Forecourt Site Preferences
├── company              Link → Company          (Mandatory, Name)
├── default_fuel_supplier Link → Supplier        (default supplier for Purchase Receipts)
│
├── [Section] Variance Thresholds
│   ├── wetstock_normal_pct    Float  default 0.30  (%)
│   ├── wetstock_elevated_pct  Float  default 0.50  (%)
│   ├── cash_normal_kes        Currency default 50
│   ├── cash_elevated_kes      Currency default 200
│   ├── meter_check_a_warn_kes Currency default 5
│   ├── meter_check_a_fail_kes Currency default 20
│   ├── meter_check_b_warn_pct Float   default 0.30
│   ├── meter_check_b_fail_pct Float   default 0.50
│   └── meter_check_b_tamper_pct Float default 1.00
│
├── [Section] GL Account Overrides (all Link → Account)
│   ├── till_active         (override "Till — Active")
│   ├── safe_main           (override "Safe — Main")
│   ├── mpesa_clearing
│   ├── card_clearing
│   ├── fleet_card_clearing
│   ├── cash_short_over
│   ├── drive_off_losses
│   └── accounts_payable
│
├── [Section] Email Reports
│   ├── send_daily_summary  Check  default 1
│   └── report_recipients   Table → Forecourt Report Recipient
│       ├── email  Data (Mandatory)
│       └── name_  Data
│
└── [Section] Operations
    ├── min_settle_minutes   Int  default 10  (delivery settle time)
    ├── cash_pickup_threshold Currency default 30000
    └── police_report_threshold Currency default 500
```

---

## 37. Custom Fields Required on ERPNext Native DocTypes

These Custom Fields must be created (via Fixtures or Developer Mode) to link native ERPNext documents back to the forecourt layer.

### 37.1 Custom Fields — POS Invoice

```python
# Create via: Customise → Custom Field → New (or use fixtures)

custom_fields = {
    "POS Invoice": [
        {
            "fieldname":    "forecourt_shift",
            "label":        "Forecourt Shift",
            "fieldtype":    "Link",
            "options":      "Forecourt Shift",
            "insert_after": "pos_profile",
            "read_only":    1,
            "in_list_view": 1,
            "search_index": 1,
            "description":  "Auto-linked to the Open Forecourt Shift on submit.",
        },
        {
            "fieldname":    "pump_attendant",
            "label":        "Pump Attendant",
            "fieldtype":    "Link",
            "options":      "Employee",
            "insert_after": "forecourt_shift",
            "reqd":         0,   # Made mandatory in POS Invoice override
            "in_list_view": 1,
            "description":  "Required for all fuel POS Invoices.",
        },
        {
            "fieldname":    "pump",
            "label":        "Pump",
            "fieldtype":    "Link",
            "options":      "Forecourt Pump",
            "insert_after": "pump_attendant",
            "description":  "Pump from which fuel was dispensed.",
        },
    ],

    "Sales Invoice": [
        {
            "fieldname":    "forecourt_shift",
            "label":        "Forecourt Shift",
            "fieldtype":    "Link",
            "options":      "Forecourt Shift",
            "insert_after": "debit_to",
            "read_only":    1,
            "search_index": 1,
        },
    ],

    "Warehouse": [
        {
            "fieldname":    "capacity_in_litres",
            "label":        "Capacity (Litres)",
            "fieldtype":    "Float",
            "insert_after": "warehouse_name",
            "description":  "Maximum fuel storage capacity. Used in Dip Reading validation.",
        },
        {
            "fieldname":    "is_fuel_tank",
            "label":        "Is Fuel Tank",
            "fieldtype":    "Check",
            "insert_after": "capacity_in_litres",
            "default":      "0",
            "description":  "Check if this warehouse represents a physical fuel storage tank.",
        },
    ],

    "Item": [
        {
            "fieldname":    "fuel_grade",
            "label":        "Fuel Grade",
            "fieldtype":    "Select",
            "options":      "\nPMS Unleaded\nPMS V-Power\nAGO\nDPK",
            "insert_after": "item_group",
            "description":  "Regulatory fuel grade classification (EPRA).",
        },
    ],
}
```

**Install via bench command (recommended for reproducibility):**

```python
# forecourt/setup.py  — called by: bench --site [site] run-script forecourt.setup.install_custom_fields

import frappe
from frappe.custom.doctype.custom_field.custom_field import create_custom_fields


def install_custom_fields():
    """
    Create all required Custom Fields for the forecourt app.
    Idempotent — safe to run multiple times.
    """
    from forecourt.setup import custom_fields as CF
    create_custom_fields(CF, ignore_validate=True)
    frappe.db.commit()
    print("Forecourt custom fields installed.")
```

---

## 38. JavaScript — Additional Client Scripts

### 38.1 Cash Event Form — Real-Time Dual Control Enforcement

```javascript
// forecourt/public/js/cash_event.js

frappe.ui.form.on("Cash Event", {

    refresh(frm) {
        if (frm.doc.event_type === "Float Issued") {
            frm.set_df_property("reference", "reqd", 0);
        }
        if (frm.doc.event_type === "Cash Pickup" || frm.doc.event_type === "Safe Drop") {
            frm.set_df_property("reference", "reqd", 1);
        }
        display_expected_till_impact(frm);
    },

    event_type(frm) {
        const requires_ref = ["Cash Pickup", "Safe Drop"].includes(frm.doc.event_type);
        frm.set_df_property("reference", "reqd", requires_ref ? 1 : 0);
        frm.refresh_field("reference");
    },

    authorised_by(frm) {
        if (!frm.doc.authorised_by || !frm.doc.cashier_session) return;

        // Fetch the cashier from the linked session and warn if same as authorised_by
        frappe.db.get_value("Cashier Session", frm.doc.cashier_session, "cashier", r => {
            if (r && r.cashier && r.cashier === frm.doc.authorised_by) {
                frappe.msgprint({
                    message: "The authorising supervisor cannot be the same person as the cashier. "
                             "Select a different supervisor.",
                    indicator: "red",
                    title: "Dual Control Violation",
                });
                frm.set_value("authorised_by", "");
            }
        });
    },

    amount(frm) {
        display_expected_till_impact(frm);
    },
});


function display_expected_till_impact(frm) {
    if (!frm.doc.event_type || !frm.doc.amount) {
        frm.dashboard.clear_headline();
        return;
    }

    const impact_map = {
        "Float Issued": { direction: "increase", label: "Opens till with" },
        "Cash Pickup":  { direction: "decrease", label: "Removes from till" },
        "Payout":       { direction: "decrease", label: "Removes from till" },
        "Safe Drop":    { direction: "decrease", label: "Deposits to safe" },
    };

    const info = impact_map[frm.doc.event_type];
    if (!info) return;

    const colour = info.direction === "increase" ? "blue" : "orange";
    const amt    = frappe.format(frm.doc.amount, { fieldtype: "Currency" });

    frm.dashboard.set_headline_alert(
        `<span class="indicator ${colour}">${info.label}: ${amt}</span>`
    );
}
```

### 38.2 Shift Reconciliation Form — Approval and GL Post Button

```javascript
// forecourt/public/js/shift_reconciliation.js

frappe.ui.form.on("Shift Reconciliation", {

    refresh(frm) {
        render_verdict_banner(frm);
        add_approval_button(frm);
        add_gl_post_button(frm);
    },
});


function render_verdict_banner(frm) {
    const cash_ok    = frm.doc.is_cash_balanced;
    const ws_ok      = frm.doc.is_wetstock_balanced;
    const meters_ok  = frm.doc.is_meter_ok;
    const gl_posted  = !!frm.doc.gl_journal;

    const all_ok = cash_ok && ws_ok && meters_ok;
    const colour = gl_posted ? "green" : (all_ok ? "blue" : "red");

    const parts = [
        `Meters: ${meters_ok ? "✓ OK" : "⚠ Review"}`,
        `Wetstock: ${ws_ok ? "✓ OK" : "⚠ Review"}`,
        `Cash: ${cash_ok ? "✓ Balanced" : "⚠ " + frm.doc.cash_variance_status}`,
        `GL: ${gl_posted ? "✓ " + frm.doc.gl_journal : "Not posted"}`,
    ];

    frm.dashboard.set_headline_alert(
        `<span class="indicator ${colour}">${parts.join(" &nbsp;|&nbsp; ")}</span>`
    );
}


function add_approval_button(frm) {
    if (!frm.doc.requires_approval || frm.doc.approved_by) return;
    if (!frappe.user.has_role("Forecourt Manager")) return;

    frm.add_custom_button("Approve Variances", () => {
        frappe.confirm(
            `This reconciliation has variances that exceed normal thresholds. 
             By approving, you confirm you have investigated the variances 
             and accept the results. Proceed?`,
            () => {
                frappe.call({
                    method: "frappe.client.set_value",
                    args: {
                        doctype: "Shift Reconciliation",
                        name:    frm.doc.name,
                        fieldname: {
                            approved_by: frappe.session.user,
                            approved_at: frappe.datetime.now_datetime(),
                        },
                    },
                    callback() {
                        frappe.show_alert({ message: "Variances approved", indicator: "green" });
                        frm.reload_doc();
                    },
                });
            }
        );
    }, "Actions");
}


function add_gl_post_button(frm) {
    if (frm.doc.gl_journal) return;

    // Require either no approval needed, or approval already given
    const can_post = !frm.doc.requires_approval || frm.doc.approved_by;
    if (!can_post) return;
    if (!frappe.user.has_role(["Forecourt Manager", "Forecourt Supervisor"])) return;

    frm.add_custom_button("Post GL Journal Entry", () => {
        frappe.confirm(
            `Post the GL Journal Entry for Shift ${frm.doc.shift}? This cannot be undone.`,
            () => {
                frappe.dom.freeze("Posting GL…");
                frappe.call({
                    method: "forecourt.utils.gl.post_shift_journal_entry",
                    args:   { shift_reconciliation_name: frm.doc.name },
                    callback(r) {
                        frappe.dom.unfreeze();
                        if (r.message) {
                            frappe.show_alert({
                                message:   `GL Journal ${r.message} posted`,
                                indicator: "green",
                            });
                            frm.reload_doc();
                        }
                    },
                    error() { frappe.dom.unfreeze(); },
                });
            }
        );
    }, "Actions");
}
```

### 38.3 Fuel Delivery Dip Form — Accept / Dispute Buttons

```javascript
// forecourt/public/js/fuel_delivery_dip.js

frappe.ui.form.on("Fuel Delivery Dip", {

    refresh(frm) {
        display_variance_badge(frm);
        add_accept_dispute_buttons(frm);
    },

    dip_after(frm) {
        if (frm.doc.dip_after) {
            // Auto-populate dip_after_l from the linked Dip Reading
            frappe.db.get_value("Dip Reading", frm.doc.dip_after, "volume_observed_l", r => {
                if (r && r.volume_observed_l) {
                    frm.set_value("dip_after_l", r.volume_observed_l);
                    compute_variance(frm);
                }
            });
        }
    },

    dip_before(frm) {
        if (frm.doc.dip_before) {
            frappe.db.get_value("Dip Reading", frm.doc.dip_before, "volume_observed_l", r => {
                if (r && r.volume_observed_l) {
                    frm.set_value("dip_before_l", r.volume_observed_l);
                    compute_variance(frm);
                }
            });
        }
    },
});


function compute_variance(frm) {
    const before  = frm.doc.dip_before_l || 0;
    const after   = frm.doc.dip_after_l  || 0;
    const docket  = frm.doc.docket_volume_l || 0;

    if (before >= 0 && after > 0) {
        const measured  = after - before;
        const variance  = measured - docket;
        const pct       = docket > 0 ? (variance / docket * 100) : 0;

        frm.set_value("dip_measured_l",       measured);
        frm.set_value("delivery_variance_l",  variance);
        frm.set_value("delivery_variance_pct", pct);

        display_variance_badge(frm);
    }
}


function display_variance_badge(frm) {
    const pct = frm.doc.delivery_variance_pct || 0;
    const abs_pct = Math.abs(pct);
    let colour = "green", label = "Normal";

    if (abs_pct > 0.5) {
        colour = "red";
        label  = `CRITICAL — ${pct.toFixed(3)}% — Dispute this delivery`;
    } else if (abs_pct > 0.3) {
        colour = "orange";
        label  = `Elevated — ${pct.toFixed(3)}% — Review`;
    } else {
        label = `Normal — ${pct.toFixed(3)}%`;
    }

    frm.dashboard.set_headline_alert(
        `<span class="indicator ${colour}">Delivery Variance: ${label}</span>`
    );
}


function add_accept_dispute_buttons(frm) {
    if (frm.doc.status !== "Pending" || frm.doc.docstatus !== 0) return;

    frm.add_custom_button("Accept Delivery", () => {
        frappe.confirm(
            `Accept delivery of ${(frm.doc.dip_measured_l || frm.doc.docket_volume_l)} L? ` +
            "This will create a Purchase Receipt and add stock to the tank.",
            () => {
                frm.set_value("status", "Accepted");
                frm.save().then(() => {
                    frappe.call({
                        method: "forecourt.utils.buying.create_purchase_receipt_for_delivery",
                        args: {
                            fuel_delivery_dip_name: frm.doc.name,
                            company:      frm.doc.company || frappe.defaults.get_default("company"),
                            cost_center:  frappe.defaults.get_default("cost_center"),
                        },
                        callback(r) {
                            if (r.message) {
                                frappe.show_alert({
                                    message:   `Purchase Receipt ${r.message} created`,
                                    indicator: "green",
                                });
                                frm.reload_doc();
                            }
                        },
                    });
                });
            }
        );
    }, "Actions");

    frm.add_custom_button("Dispute Delivery", () => {
        frappe.prompt(
            [{ fieldtype: "Small Text", fieldname: "dispute_notes", label: "Dispute Reason", reqd: 1 }],
            values => {
                frm.set_value("status",        "Disputed");
                frm.set_value("dispute_notes", values.dispute_notes);
                frm.save().then(() => {
                    frappe.show_alert({ message: "Delivery marked as Disputed", indicator: "orange" });
                });
            },
            "Dispute Delivery",
            "Submit Dispute",
        );
    }, "Actions");
}
```

---

## 39. Role and Permission Setup

### 39.1 Roles Required

| Role | Description | Desk Access |
|---|---|---|
| `Forecourt Manager` | Full access to all forecourt doctypes; can approve variances; can post GL; can dispute/resolve shifts | Yes |
| `Forecourt Supervisor` | Can open/close shifts; can authorise cash events; can accept deliveries; cannot post GL | Yes |
| `Cashier` | Can create POS Invoices; can create Cash Events (own session only); read-only on shift | Yes |
| `Pump Attendant` | No Desk access; may be linked to Employee records only | No |

### 39.2 Permission Matrix

```python
# forecourt/config/docs.py
# Defines DocType permissions — import into ERPNext via Fixtures

PERMISSIONS = {
    "Forecourt Shift": [
        {"role": "Forecourt Manager",    "read": 1, "write": 1, "create": 1, "delete": 0, "submit": 1, "cancel": 0, "amend": 0},
        {"role": "Forecourt Supervisor", "read": 1, "write": 1, "create": 1, "delete": 0, "submit": 1, "cancel": 0, "amend": 0},
        {"role": "Cashier",              "read": 1, "write": 0, "create": 0, "delete": 0, "submit": 0, "cancel": 0, "amend": 0},
    ],
    "Meter Reading": [
        {"role": "Forecourt Manager",    "read": 1, "write": 1, "create": 1, "delete": 0, "submit": 1, "cancel": 0, "amend": 1},
        {"role": "Forecourt Supervisor", "read": 1, "write": 1, "create": 1, "delete": 0, "submit": 1, "cancel": 0, "amend": 0},
        {"role": "Cashier",              "read": 1, "write": 0, "create": 0, "delete": 0, "submit": 0, "cancel": 0, "amend": 0},
    ],
    "Dip Reading": [
        {"role": "Forecourt Manager",    "read": 1, "write": 1, "create": 1, "delete": 0, "submit": 1, "cancel": 0},
        {"role": "Forecourt Supervisor", "read": 1, "write": 1, "create": 1, "delete": 0, "submit": 1, "cancel": 0},
        {"role": "Cashier",              "read": 1, "write": 0, "create": 0, "delete": 0, "submit": 0, "cancel": 0},
    ],
    "Cash Event": [
        {"role": "Forecourt Manager",    "read": 1, "write": 1, "create": 1, "delete": 0, "submit": 1, "cancel": 1},
        {"role": "Forecourt Supervisor", "read": 1, "write": 1, "create": 1, "delete": 0, "submit": 1, "cancel": 0},
        {"role": "Cashier",              "read": 1, "write": 0, "create": 1, "delete": 0, "submit": 0, "cancel": 0},
    ],
    "Shift Reconciliation": [
        {"role": "Forecourt Manager",    "read": 1, "write": 1, "create": 0, "delete": 0, "submit": 0, "cancel": 0},
        {"role": "Forecourt Supervisor", "read": 1, "write": 0, "create": 0, "delete": 0, "submit": 0, "cancel": 0},
        {"role": "Cashier",              "read": 0, "write": 0, "create": 0, "delete": 0, "submit": 0, "cancel": 0},
    ],
    "Drive-Off Record": [
        {"role": "Forecourt Manager",    "read": 1, "write": 1, "create": 1, "delete": 0, "submit": 1, "cancel": 1},
        {"role": "Forecourt Supervisor", "read": 1, "write": 1, "create": 1, "delete": 0, "submit": 1, "cancel": 0},
        {"role": "Cashier",              "read": 1, "write": 0, "create": 0, "delete": 0, "submit": 0, "cancel": 0},
    ],
    "Fuel Delivery Dip": [
        {"role": "Forecourt Manager",    "read": 1, "write": 1, "create": 1, "delete": 0, "submit": 1, "cancel": 1},
        {"role": "Forecourt Supervisor", "read": 1, "write": 1, "create": 1, "delete": 0, "submit": 1, "cancel": 0},
        {"role": "Cashier",              "read": 1, "write": 0, "create": 0, "delete": 0, "submit": 0, "cancel": 0},
    ],
    "Forecourt Site Preferences": [
        {"role": "Forecourt Manager",    "read": 1, "write": 1, "create": 1, "delete": 0},
        {"role": "Forecourt Supervisor", "read": 1, "write": 0, "create": 0, "delete": 0},
        {"role": "Cashier",              "read": 0, "write": 0, "create": 0, "delete": 0},
    ],
}
```

---

## 40. App Directory Structure

Complete file tree for the `forecourt` Frappe app as referenced throughout this document.

```
forecourt/
├── hooks.py
├── setup.py                              # install_custom_fields(), post_install()
├── tasks.py                              # Scheduled: daily summary, calibration, stale shifts
│
├── forecourt/                            # Module (matches app name)
│   │
│   ├── doctype/
│   │   ├── forecourt_pump/
│   │   │   ├── forecourt_pump.json       # DocType schema
│   │   │   └── forecourt_pump.py         # §21.1
│   │   │
│   │   ├── forecourt_shift/
│   │   │   ├── forecourt_shift.json
│   │   │   ├── forecourt_shift.py        # §22
│   │   │   └── pre_fetch.py              # §24
│   │   │
│   │   ├── meter_reading/
│   │   │   ├── meter_reading.json
│   │   │   └── meter_reading.py          # §21.2
│   │   │
│   │   ├── meter_validation_result/
│   │   │   └── meter_validation_result.json
│   │   │
│   │   ├── dip_reading/
│   │   │   ├── dip_reading.json
│   │   │   └── dip_reading.py            # §21.3
│   │   │
│   │   ├── fuel_delivery_dip/
│   │   │   └── fuel_delivery_dip.json
│   │   │
│   │   ├── cashier_session/
│   │   │   └── cashier_session.json
│   │   │
│   │   ├── cash_event/
│   │   │   ├── cash_event.json
│   │   │   └── cash_event.py             # §21.4
│   │   │
│   │   ├── shift_reconciliation/
│   │   │   ├── shift_reconciliation.json
│   │   │   └── reconciliation_engine.py  # §23
│   │   │
│   │   ├── drive_off_record/
│   │   │   ├── drive_off_record.json
│   │   │   └── drive_off_record.py       # §35
│   │   │
│   │   └── forecourt_site_preferences/
│   │       ├── forecourt_site_preferences.json
│   │       └── forecourt_site_preferences.py  # §36
│   │
│   └── report/
│       ├── daily_shift_report/
│       │   ├── daily_shift_report.json   # Script Report definition
│       │   └── daily_shift_report.py     # §30.1
│       ├── wetstock_variance_trend/
│       │   ├── wetstock_variance_trend.json
│       │   └── wetstock_variance_trend.py # §30.2
│       ├── cashier_performance/
│       │   ├── cashier_performance.json
│       │   └── cashier_performance.py    # §30.3
│       ├── meter_reading_discrepancy_log/
│       │   └── meter_reading_discrepancy_log.py
│       ├── attendant_fuel_volume/
│       │   └── attendant_fuel_volume.py
│       └── delivery_reconciliation/
│           └── delivery_reconciliation.py
│
├── overrides/
│   ├── __init__.py
│   ├── pos_invoice.py                    # §28
│   └── sales_invoice.py                 # §32
│
├── utils/
│   ├── __init__.py
│   ├── accounts.py                       # §17.1
│   ├── buying.py                         # §19.1
│   ├── gl.py                             # §25
│   ├── hr.py                             # §18.1
│   ├── meter.py                          # §27
│   ├── meter_api.py                      # §33.2
│   ├── pos.py                            # §20.1
│   ├── session_api.py                    # §33.1
│   ├── stock.py                          # §16.1–16.2
│   └── wetstock.py                       # §26
│
└── public/
    ├── js/
    │   ├── forecourt_shift.js            # §29.2
    │   ├── meter_reading.js              # §29.1
    │   ├── dip_reading.js                # §29.3
    │   ├── cash_event.js                 # §38.1
    │   ├── shift_reconciliation.js       # §38.2
    │   └── fuel_delivery_dip.js          # §38.3
    └── css/
        └── forecourt.css
```

---

## 41. bench Commands — Setup and Maintenance Reference

All commands assume: `cd /home/frappe/frappe-bench`

### 41.1 App Creation and Installation

```bash
# Create the custom app (run once)
bench new-app forecourt

# Install on your site
bench --site [your-site.com] install-app forecourt

# After code changes: rebuild assets and reload
bench --site [your-site.com] build --app forecourt
bench --site [your-site.com] clear-cache
```

### 41.2 Custom Field and Fixture Management

```bash
# Export fixtures after configuring Custom Fields via the UI
bench --site [your-site.com] export-fixtures --app forecourt

# Import fixtures on a new site
bench --site [your-site.com] import-fixtures --app forecourt

# Run the setup script to install Custom Fields programmatically
bench --site [your-site.com] execute forecourt.setup.install_custom_fields
```

### 41.3 Schema Migration

```bash
# After modifying any DocType JSON file, run migrate to apply DDL changes
bench --site [your-site.com] migrate

# To create a new DocType from JSON (after bench new-app):
# Place forecourt_pump.json in forecourt/forecourt/doctype/forecourt_pump/
# Then run migrate
```

### 41.4 Scheduler and Background Jobs

```bash
# Check if scheduler is running
bench --site [your-site.com] scheduler status

# Trigger a specific scheduled task manually (testing)
bench --site [your-site.com] execute forecourt.tasks.send_daily_shift_summary
bench --site [your-site.com] execute forecourt.tasks.check_calibration_due_dates

# View background worker logs
bench --site [your-site.com] worker --queue default
```

### 41.5 Go-Live Baseline

```bash
# Before first live shift — run the baseline data migration console
bench --site [your-site.com] console

# Inside the console:
# import frappe
# frappe.set_user("administrator")
#
# Create the baseline shift
# baseline = frappe.new_doc("Forecourt Shift")
# baseline.shift_date = frappe.utils.today()
# baseline.status = "Draft"
# baseline.station = "Shell Maanzoni"
# ...etc
#
# After baseline readings are entered and first live shift is open:
# from forecourt.utils.stock import get_current_wac
# print(get_current_wac("FUEL-PMS-UNL", "Tank 1 - Unleaded - SMSS"))
# Confirm WAC is seeded from the baseline Purchase Receipt
```

### 41.6 Common Debugging Queries

```sql
-- Check for open shifts (should be max 1 per station)
SELECT name, status, cashier, opened_at
FROM `tabForecourt Shift`
WHERE status IN ('Open', 'Closing')
ORDER BY opened_at DESC;

-- POS Invoices with no shift linked (data quality check)
SELECT name, posting_date, grand_total
FROM `tabPOS Invoice`
WHERE forecourt_shift IS NULL
  AND docstatus = 1
ORDER BY posting_date DESC
LIMIT 20;

-- Meter readings missing for a shift (identify gaps before close)
SELECT fp.pump_number, pn.nozzle_number, mr.meter_type, mr.reading_position
FROM `tabForecourt Pump` fp
JOIN `tabPump Nozzle` pn ON pn.parent = fp.name
LEFT JOIN `tabMeter Reading` mr
    ON mr.pump = fp.name
   AND mr.nozzle_number = pn.nozzle_number
   AND mr.shift = 'SHIFT-2026-0001'   -- replace with target shift
   AND mr.docstatus = 1
WHERE fp.is_active = 1 AND pn.is_active = 1
  AND mr.name IS NULL
ORDER BY fp.pump_number, pn.nozzle_number;

-- Wetstock variance trend (last 30 days)
SELECT
    fs.shift_date,
    tws.tank,
    ROUND(tws.variance_l, 2)   AS variance_l,
    ROUND(tws.variance_pct, 4) AS variance_pct,
    tws.classification
FROM `tabShift Reconciliation Tank Wetstock` tws
JOIN `tabShift Reconciliation` sr ON sr.name = tws.parent
JOIN `tabForecourt Shift` fs ON fs.name = sr.shift
WHERE fs.shift_date >= DATE_SUB(CURDATE(), INTERVAL 30 DAY)
ORDER BY fs.shift_date DESC, tws.tank;

-- Cashier variance summary (all time)
SELECT
    fs.cashier,
    COUNT(*) AS shift_count,
    ROUND(SUM(sr.cash_variance), 2)  AS net_variance_kes,
    ROUND(AVG(sr.cash_variance), 2)  AS avg_variance_kes,
    SUM(CASE WHEN sr.cash_variance < 0 THEN 1 ELSE 0 END) AS short_count,
    MAX(ABS(sr.cash_variance))       AS max_single_variance
FROM `tabShift Reconciliation` sr
JOIN `tabForecourt Shift` fs ON fs.name = sr.shift
GROUP BY fs.cashier
ORDER BY ABS(SUM(sr.cash_variance)) DESC;
```

---

## 42. Testing Reference

### 42.1 Python Unit Tests

```python
# forecourt/tests/test_meter_validation.py

import frappe
import unittest
from forecourt.utils.meter import _validate_nozzle


class TestMeterValidation(unittest.TestCase):

    def _make_shift(self, rate=197.60):
        """Return a mock shift object with PMS Unleaded rate."""
        shift = frappe._dict(rate_pms_unl=rate, rate_pms_vp=212.0, rate_ago=160.0, rate_dpk=140.0)
        return shift

    def _make_nozzle_data(self, elec_vol=100.0, elec_cash=None, mech_vol=None,
                          fuel_product="FUEL-PMS-UNL"):
        rate = 197.60
        if elec_cash is None:
            elec_cash = elec_vol * rate        # perfect agreement
        if mech_vol is None:
            mech_vol = elec_vol                # perfect agreement
        return frappe._dict(
            pump="Pump 1",
            nozzle_number=1,
            fuel_product=fuel_product,
            tank="Tank 1 - Unleaded - SMSS",
            elec_vol_sold=elec_vol,
            elec_cash_sold=elec_cash,
            mech_vol_sold=mech_vol,
        )

    def test_check_a_pass_within_tolerance(self):
        """Elec Cash within KES 5 of expected — should Pass."""
        nd = self._make_nozzle_data(elec_vol=100.0, elec_cash=100.0 * 197.60 + 3.0)
        result = _validate_nozzle(nd, self._make_shift())
        self.assertEqual(result["check_a_status"], "Pass")

    def test_check_a_warning_between_5_and_20(self):
        """Elec Cash KES 10 off — Warning."""
        nd = self._make_nozzle_data(elec_vol=100.0, elec_cash=100.0 * 197.60 + 10.0)
        result = _validate_nozzle(nd, self._make_shift())
        self.assertEqual(result["check_a_status"], "Warning")

    def test_check_a_fail_above_20(self):
        """Elec Cash KES 50 off — Fail."""
        nd = self._make_nozzle_data(elec_vol=100.0, elec_cash=100.0 * 197.60 + 50.0)
        result = _validate_nozzle(nd, self._make_shift())
        self.assertEqual(result["check_a_status"], "Fail")

    def test_check_b_pass_within_0_3_pct(self):
        """Mech 0.1% off Elec Vol — Pass."""
        nd = self._make_nozzle_data(elec_vol=1000.0, mech_vol=999.0)
        result = _validate_nozzle(nd, self._make_shift())
        self.assertEqual(result["check_b_status"], "Pass")

    def test_check_b_warning_between_0_3_and_0_5_pct(self):
        """Mech 0.4% off — Warning."""
        nd = self._make_nozzle_data(elec_vol=1000.0, mech_vol=996.0)
        result = _validate_nozzle(nd, self._make_shift())
        self.assertEqual(result["check_b_status"], "Warning")

    def test_check_b_fail_above_0_5_pct(self):
        """Mech 0.8% off — Fail."""
        nd = self._make_nozzle_data(elec_vol=1000.0, mech_vol=992.0)
        result = _validate_nozzle(nd, self._make_shift())
        self.assertEqual(result["check_b_status"], "Fail")

    def test_check_b_critical_above_1_pct(self):
        """Mech 1.5% off — Critical (tamper threshold)."""
        nd = self._make_nozzle_data(elec_vol=1000.0, mech_vol=985.0)
        result = _validate_nozzle(nd, self._make_shift())
        self.assertEqual(result["check_b_status"], "Critical")

    def test_overall_status_is_worst_of_a_and_b(self):
        """Overall = max severity of Check A and Check B."""
        # A=Pass, B=Warning → overall=Warning
        nd = self._make_nozzle_data(elec_vol=1000.0, mech_vol=996.0,
                                     elec_cash=1000.0 * 197.60 + 2.0)
        result = _validate_nozzle(nd, self._make_shift())
        self.assertEqual(result["overall_status"], "Warning")

    def test_zero_elec_vol_does_not_crash(self):
        """Inactive nozzle with zero sales should produce Pass without ZeroDivision."""
        nd = self._make_nozzle_data(elec_vol=0.0, elec_cash=0.0, mech_vol=0.0)
        result = _validate_nozzle(nd, self._make_shift())
        self.assertEqual(result["check_b_divergence_pct"], 0.0)
        self.assertEqual(result["check_b_status"], "Pass")

    def test_shell_maanzoni_actual_shift_data(self):
        """
        Reference test using actual shift data from Shell Maanzoni (2026-04-24).
        Unleaded: 1,731.95 L @ KES 197.60 = KES 342,193.42 expected Elec Cash.
        """
        nd = self._make_nozzle_data(
            elec_vol=1731.95,
            elec_cash=342_193.42,
            mech_vol=1731.8,
            fuel_product="FUEL-PMS-UNL",
        )
        result = _validate_nozzle(nd, self._make_shift(rate=197.60))
        # KES discrepancy should be essentially zero (< KES 5)
        self.assertEqual(result["check_a_status"], "Pass")
        # Mech divergence: |1731.95 - 1731.8| / 1731.95 = 0.0087% → well under 0.3%
        self.assertEqual(result["check_b_status"], "Pass")
        self.assertEqual(result["overall_status"], "Pass")


# Run with:
# bench --site [site] run-tests --app forecourt --module forecourt.tests.test_meter_validation
```

### 42.2 Wetstock Test

```python
# forecourt/tests/test_wetstock.py

import unittest
from frappe.utils import flt


class TestWetstockFormula(unittest.TestCase):
    """
    Unit tests for the wetstock variance formula using Shell Maanzoni reference data.
    These tests exercise the formula logic without hitting the database.
    """

    def _compute(self, opening, deliveries, elec_vol_sales, actual_closing):
        """Mirror of the core wetstock formula."""
        theoretical = opening + deliveries - elec_vol_sales
        variance_l  = theoretical - actual_closing
        denominator = opening + deliveries
        variance_pct = (variance_l / denominator * 100) if denominator else 0.0

        abs_pct = abs(variance_pct)
        if variance_l < 0 and abs_pct > 0.50:
            classification = "Critical"
        elif variance_l < 0 and abs_pct > 0.30:
            classification = "Elevated"
        elif variance_l > 0 and abs_pct > 0.30:
            classification = "Gain"
        else:
            classification = "Normal"

        return {
            "theoretical_closing_l": theoretical,
            "variance_l":            variance_l,
            "variance_pct":          round(variance_pct, 4),
            "classification":        classification,
        }

    def test_no_delivery_balanced(self):
        """Perfect shift: no delivery, actual = theoretical."""
        r = self._compute(opening=5000, deliveries=0, elec_vol_sales=250, actual_closing=4750)
        self.assertAlmostEqual(r["variance_l"], 0.0)
        self.assertEqual(r["classification"], "Normal")

    def test_small_loss_normal(self):
        """0.15% loss — Normal classification."""
        r = self._compute(opening=5000, deliveries=0, elec_vol_sales=250, actual_closing=4742.5)
        self.assertAlmostEqual(r["variance_pct"], 0.15, places=2)
        self.assertEqual(r["classification"], "Normal")

    def test_elevated_loss(self):
        """0.4% loss — Elevated."""
        r = self._compute(opening=5000, deliveries=0, elec_vol_sales=250, actual_closing=4730)
        self.assertAlmostEqual(abs(r["variance_pct"]), 0.4, places=1)
        self.assertEqual(r["classification"], "Elevated")

    def test_critical_loss(self):
        """0.8% loss — Critical."""
        r = self._compute(opening=5000, deliveries=0, elec_vol_sales=250, actual_closing=4710)
        self.assertGreater(abs(r["variance_pct"]), 0.5)
        self.assertEqual(r["classification"], "Critical")

    def test_gain_flagged(self):
        """Apparent gain > 0.3% — flagged as Gain (suspicious)."""
        r = self._compute(opening=5000, deliveries=0, elec_vol_sales=250, actual_closing=4766)
        self.assertGreater(r["variance_pct"], 0.3)
        self.assertEqual(r["classification"], "Gain")

    def test_delivery_included(self):
        """Delivery of 8000 L on top of opening 5000 L."""
        r = self._compute(
            opening=5000, deliveries=8000,
            elec_vol_sales=1750, actual_closing=11240
        )
        # Theoretical = 5000 + 8000 - 1750 = 11250
        # Variance = 11250 - 11240 = +10 L
        self.assertAlmostEqual(r["theoretical_closing_l"], 11250.0)
        self.assertAlmostEqual(r["variance_l"], 10.0)

    def test_shell_maanzoni_tank1_reference(self):
        """
        Reference: Shell Maanzoni Tank 1 (Unleaded), shift 2026-04-24.
        Opening: ~7,200 L (estimated baseline), no delivery, Elec Vol sold 1731.95 L.
        Actual closing dip: ~5,470 L.  Variance ~ +1.95 L (rounding in dip) → Normal.
        """
        r = self._compute(opening=7200, deliveries=0, elec_vol_sales=1731.95,
                          actual_closing=5470)
        # Theoretical = 7200 - 1731.95 = 5468.05
        # Variance = 5468.05 - 5470 = -1.95 L (slight loss — Normal)
        self.assertAlmostEqual(r["theoretical_closing_l"], 5468.05, places=1)
        self.assertEqual(r["classification"], "Normal")


# Run with:
# bench --site [site] run-tests --app forecourt --module forecourt.tests.test_wetstock
```

---

*End of Integration Architecture Addendum — Sections 32–42*
*Document complete: v3.1.0*
*Next steps: (1) bench new-app forecourt, (2) install_custom_fields, (3) migrate, (4) configure Forecourt Site Preferences, (5) go-live baseline procedure (§11)*
