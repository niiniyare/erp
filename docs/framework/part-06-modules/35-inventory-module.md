---
title: "Chapter 35: Inventory Module"
part: "Part VI — Built-In Modules"
chapter: 35
section: "35-inventory-module"
related:
  - "[Chapter 33: Finance Module](33-finance-module.md)"
  - "[Chapter 29: Saga Pattern](../part-05-workflow/29-saga-pattern.md)"
---

# Chapter 35: Inventory Module

The Inventory module manages stock items, warehouses, stock movements, and valuation. It integrates with Finance for COGS posting and with the Forecourt module for fuel stock management.

---

## 35.1. Item and Warehouse Structure

### 35.1.1. Item Entity

```go
type Item struct {
    Code             string
    Name             string
    UnitOfMeasure    string           // "KG", "L", "PCS", "BOX"
    ValuationMethod  string           // "FIFO" | "WAVG" (weighted average)
    IsStockItem      bool             // trackable vs non-stock items
    ReorderPoint     decimal.Decimal  // trigger reorder below this level
    SafetyStock      decimal.Decimal  // minimum buffer
    ItemGroupID      uuid.UUID
    DefaultUOMConversions []UOMConversion
}
```

### 35.1.2. Item Group Hierarchy

Items are grouped for reporting: `Electronics → Computers → Laptops`. Uses materialised path for efficient tree queries. GL accounts (income and expense) are assigned at the item group level and inherited by items.

### 35.1.3. Warehouse Entity

```go
type Warehouse struct {
    Code      string
    Name      string
    IsActive  bool
    AccountID uuid.UUID  // Stock/Inventory GL account for this warehouse
    Bins      []Bin
}
```

### 35.1.4. Bin Entity — Sub-Locations

For warehouses with organised bin locations (rack A, shelf 3, position 2):

```go
type Bin struct {
    WarehouseID uuid.UUID
    Code        string   // "A-03-02"
    Name        string
    Capacity    decimal.Decimal
    IsActive    bool
}
```

---

## 35.2. Stock Entry Types

### 35.2.1. Receipt — Goods In From Purchase

```go
type StockEntry struct {
    Type        string   // "receipt" | "issue" | "transfer" | "adjustment"
    WarehouseID uuid.UUID
    PostingDate time.Time
    Reference   string   // Purchase Order number, Delivery Note number
    Lines       []StockEntryLine
}

type StockEntryLine struct {
    ItemID       uuid.UUID
    BinID        *uuid.UUID
    Quantity     decimal.Decimal
    UOM          string
    UnitCost     decimal.Decimal  // for receipts
    BatchNumber  *string          // for batch-tracked items
    SerialNumber *string          // for serial-tracked items
    ExpiryDate   *time.Time
}
```

### 35.2.2. Issue — Goods Out

Issues reduce stock. Can be:
- Sales delivery: goods issued from warehouse to customer
- Internal consumption: goods issued for production/maintenance
- Write-off: damaged or expired goods removed from stock

### 35.2.3. Transfer — Between Warehouses

A transfer creates two stock entries simultaneously:
1. Issue from source warehouse (decreases stock)
2. Receipt at destination warehouse (increases stock)

Both must succeed atomically — handled via a Temporal saga.

### 35.2.4. Adjustment — Physical Count Correction

After a physical count reveals a discrepancy:

```
System stock: 500 KG
Physical count: 487 KG
Adjustment: -13 KG

GL posting:
  Debit: Stock Write-off Expense KES X
  Credit: Inventory Asset KES X
```

---

## 35.3. Valuation

### 35.3.1. FIFO — First In First Out

Each receipt creates a "cost layer" with quantity and unit cost. Issues deplete the oldest layers first:

```
Layer 1: 100 KG @ KES 45.00 = KES 4,500 (received 2025-07-01)
Layer 2: 200 KG @ KES 47.50 = KES 9,500 (received 2025-07-10)

Issue 150 KG:
  From Layer 1: 100 KG @ KES 45.00 = KES 4,500
  From Layer 2:  50 KG @ KES 47.50 = KES 2,375
  Total COGS: KES 6,875
  Average cost of issue: KES 45.83/KG
```

### 35.3.2. Weighted Average — Moving Average Cost

The unit cost is recalculated after each receipt:

```
Current: 100 KG @ KES 45.00 avg = KES 4,500
Receipt: 200 KG @ KES 47.50 = KES 9,500
New avg: (4500 + 9500) / (100 + 200) = KES 14,000 / 300 = KES 46.67/KG
```

All subsequent issues use KES 46.67 as the unit cost until the next receipt.

### 35.3.3. GL Impact of Stock Entries

| Entry type | Debit | Credit |
|---|---|---|
| Receipt | Inventory Asset | Accounts Payable (or GR/IR) |
| Issue (sale) | COGS Expense | Inventory Asset |
| Transfer out | GIT (Goods in Transit) | Inventory Asset (source WH) |
| Transfer in | Inventory Asset (dest WH) | GIT |
| Adjustment down | Stock Write-off Expense | Inventory Asset |
| Adjustment up | Inventory Asset | Stock Excess Income |

---

## 35.4. Reorder and Physical Count

### 35.4.1. Reorder Point

The nightly stock check workflow identifies items below their reorder point:

```go
func NightlyStockCheckWorkflow(ctx workflow.Context, params StockCheckParams) error {
    var lowStockItems []LowStockItem
    workflow.ExecuteActivity(ctx, activities.FindLowStockItems, params.TenantID).
        Get(ctx, &lowStockItems)

    for _, item := range lowStockItems {
        // Trigger purchase requisition if auto-reorder is enabled
        if item.AutoReorder {
            workflow.ExecuteChildWorkflow(ctx, AutoPurchaseRequisitionWorkflow,
                AutoPRInput{ItemID: item.ID, SuggestedQty: item.ReorderQty})
        }
        // Notify procurement team
        workflow.ExecuteActivity(ctx, activities.SendLowStockAlert, item)
    }
    return nil
}
```

### 35.4.2. Physical Stock Count Workflow

```
1. Freeze stock: pause all stock movements during count
2. Count sheets generated per warehouse/bin
3. Counters enter physical quantities
4. System computes variances
5. Supervisor reviews and approves adjustments
6. Adjustment stock entries posted
7. Stock movements unfrozen
```

The freeze is implemented via a per-warehouse `counting_active` flag. Any stock entry attempted during a freeze returns `WAREHOUSE_COUNT_IN_PROGRESS`.
