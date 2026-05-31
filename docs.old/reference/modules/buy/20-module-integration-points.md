[<-- Back to Index](README.md)

## Module Integration Points

### Financial Module Integration

```markdown
GENERAL LEDGER INTEGRATION

Automatic GL Posting Events:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. Purchase Requisition (No GL Impact)
   System: Budget commitment tracking only
   Status: Informational

2. Purchase Order Creation
   ┌─────────────────────────────────────────────────────┐
   │ Dr. PO Commitment (Memo Account)  1,000,000         │
   │     Cr. Budget Reserve                  1,000,000   │
   │                                                      │
   │ Description: PO-2025-00156 commitment                │
   │ Note: Off-balance sheet, budget control only         │
   └─────────────────────────────────────────────────────┘

3. Goods Receipt Note (GRN)
   ┌─────────────────────────────────────────────────────┐
   │ Dr. Inventory - Raw Materials      1,000,000        │
   │     Cr. GRN Clearing Account               1,000,000│
   │                                                      │
   │ Description: GRN-2025-0145, PO-2025-00156           │
   │ Effect: Inventory increased, liability accrued       │
   └─────────────────────────────────────────────────────┘
   
   Parallel Entry (PO Commitment Reversal):
   ┌─────────────────────────────────────────────────────┐
   │ Dr. Budget Reserve                 1,000,000        │
   │     Cr. PO Commitment                      1,000,000│
   │                                                      │
   │ Description: Reverse commitment on receipt           │
   └─────────────────────────────────────────────────────┘

4. Purchase Invoice
   ┌─────────────────────────────────────────────────────┐
   │ Dr. GRN Clearing Account           1,000,000        │
   │ Dr. VAT Input                        160,000        │
   │     Cr. Accounts Payable - Supplier        1,160,000│
   │                                                      │
   │ Description: PINV-2025-00234, invoice posted         │
   │ Effect: GRN cleared, formal liability recorded       │
   └─────────────────────────────────────────────────────┘

5. Payment
   ┌─────────────────────────────────────────────────────┐
   │ Dr. Accounts Payable - Supplier    1,160,000        │
   │     Cr. Bank Account                       1,160,000│
   │                                                      │
   │ Description: Payment PMT-2025-0345                   │
   │ Effect: Liability settled, cash reduced              │
   └─────────────────────────────────────────────────────┘

Account Mapping Configuration:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Purchase Categories → GL Accounts:
┌──────────────────────────┬──────────────┬─────────────────┐
│ Purchase Category        │ GL Account   │ Description     │
├──────────────────────────┼──────────────┼─────────────────┤
│ Raw Materials            │ 1300         │ Inventory-RM    │
│ Finished Goods           │ 1400         │ Inventory-FG    │
│ Packaging Materials      │ 1320         │ Inventory-Pack  │
│ MRO & Consumables        │ 1350         │ Inventory-MRO   │
│                          │              │                 │
│ Capital Equipment        │ 1600         │ Fixed Assets    │
│ IT Equipment             │ 1620         │ FA-IT           │
│                          │              │                 │
│ Utilities                │ 5200         │ Utility Expense │
│ Professional Services    │ 5400         │ Service Expense │
│ Rent                     │ 5100         │ Rent Expense    │
│ Office Supplies          │ 5300         │ Office Expense  │
└──────────────────────────┴──────────────┴─────────────────┘

Tax Account Mapping:
┌──────────────────────────┬──────────────┬─────────────────┐
│ Tax Type                 │ GL Account   │ Description     │
├──────────────────────────┼──────────────┼─────────────────┤
│ VAT Input (Recoverable)  │ 1520         │ VAT Receivable  │
│ WHT on Services (5%)     │ 1540         │ WHT Recoverable │
│ Import Duty              │ 1600/Cost    │ Capitalize      │
└──────────────────────────┴──────────────┴─────────────────┘

Cost Center Allocation:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Production Materials → Cost Center: PROD-01
Maintenance Supplies → Cost Center: MAINT-01
IT Equipment → Cost Center: IT-01
Office Supplies → Cost Center: ADMIN-01

Journal Entry Example with Dimensions:
┌─────────────────────────────────────────────────────────┐
│ Account: 1300 (Raw Materials)                           │
│ Cost Center: PROD-01                                    │
│ Project: (if applicable)                                │
│ Department: Production                                   │
│ Dr: 1,000,000 KES                                       │
└─────────────────────────────────────────────────────────┘
```

### Inventory Module Integration

```markdown
STOCK MANAGEMENT INTEGRATION

GRN to Stock Movement:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Event: Goods Receipt
Trigger: GRN approved

Stock Transaction:
  Transaction Type: Goods Receipt
  Source Document: GRN-2025-0145
  Reference: PO-2025-00156
  
  Item: RM-STL-001 (Cold Rolled Steel)
  Quantity: 10.25 Tons
  Rate: 80,750 KES/Ton
  Value: 827,688 KES
  
  Warehouse: WH-NBI-01
  Location: Zone A, Bin A-12-15
  Batch No: BATCH-RM-STL-2025-0329

Stock Ledger Update:
┌──────┬──────────┬──────────┬──────────┬──────────┬──────────┐
│ Date │ Type     │ Qty In   │ Qty Out  │ Balance  │ Value    │
├──────┼──────────┼──────────┼──────────┼──────────┼──────────┤
│ 03-28│ Opening  │    -     │    -     │ 15.00 T  │1,211,250 │
│ 03-29│ GRN-0145 │ 10.25 T  │    -     │ 25.25 T  │2,038,938 │
│ 03-30│ Issue    │    -     │  8.00 T  │ 17.25 T  │1,392,938 │
└──────┴──────────┴──────────┴──────────┴──────────┴──────────┘

Valuation Method: Moving Average
  Previous: 15 T @ 80,750 = 1,211,250 KES
  Received: 10.25 T @ 80,750 = 827,688 KES
  Total: 25.25 T = 2,038,938 KES
  New Avg: 2,038,938 / 25.25 = 80,750 KES/T (unchanged - same price)

Reorder Level Monitoring:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Item: RM-STL-001
Reorder Level: 10 Tons
Reorder Quantity: 10 Tons
Lead Time: 14 days

Stock Status Check (Daily):
  Current Stock: 17.25 Tons
  Status: Above reorder level ✓
  Action: None

When stock falls below 10 Tons:
  System Action: Auto-generate Purchase Requisition
  Notification: To procurement buyer
  Suggested Supplier: Ace Steel (preferred supplier)

Purchase Returns Impact:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Event: Purchase Return
Document: PR-RET-2025-005

Stock Transaction:
  Type: Purchase Return
  Quantity: 5 Units (returned)
  Value: 40,000 KES
  
  Stock Ledger:
    Dr. (Reduce) Inventory    -5 Units
    Cr. Accounts Payable       40,000 KES (via debit note)

Quality Rejection Handling:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
At GRN:
  Received: 100 Units
  Accepted: 95 Units → Stock
  Rejected: 5 Units → Quarantine (no stock entry)

Rejected items:
  Location: Quarantine Area
  Status: Held (not available for use)
  Accounting: Not booked to inventory
  Next Step: Return to supplier
```

### Manufacturing Module Integration

```markdown
PRODUCTION PLANNING INTEGRATION

MRP to Purchase Requisition:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Scenario: Production Order for 1,000 Units Product A

BOM (Bill of Materials):
  Product A requires:
    - 2 kg Raw Material X per unit
    - 1 pc Component Y per unit

Production Order: PRO-2025-0045
Quantity: 1,000 Units
Start Date: 2025-05-01

Material Requirement:
┌──────────────────┬──────────┬──────────┬──────────┬──────────┐
│ Material         │ Required │ On Hand  │ Allocated│ To Order │
├──────────────────┼──────────┼──────────┼──────────┼──────────┤
│ Raw Material X   │ 2,000 kg │  800 kg  │  500 kg  │ 1,700 kg │
│ Component Y      │ 1,000 pc │  200 pc  │  100 pc  │   900 pc │
└──────────────────┴──────────┴──────────┴──────────┴──────────┘

MRP Run (Automatic):
  1. Calculate net requirement
  2. Check reorder levels
  3. Consider lead times
  4. Generate purchase requisitions

Auto-Generated PRs:
  PR-2025-00450:
    Item: Raw Material X
    Quantity: 2,000 kg (rounded up for safety stock)
    Required By: 2025-04-25 (lead time: 5 days before production)
    Requested For: Production Department
    Priority: High
    
  PR-2025-00451:
    Item: Component Y
    Quantity: 1,000 pcs
    Required By: 2025-04-22 (lead time: 8 days)
    Requested For: Production Department
    Priority: High

Material Issue to Production:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Production Order: PRO-2025-0045
Issue Date: 2025-05-01

Material Requisition Slip (MRS):
  MRS-2025-0234
  Against: PRO-2025-0045
  
  Items Issued:
    Raw Material X: 2,000 kg @ 150 KES/kg = 300,000 KES
    Component Y: 1,000 pcs @ 50 KES/pc = 50,000 KES
    Total Material Cost: 350,000 KES

Accounting Entry:
  Dr. Work in Process (WIP)          350,000
      Cr. Inventory - Raw Materials          300,000
      Cr. Inventory - Components              50,000

Description: Material issue MRS-2025-0234 for PRO-2025-0045

Production Completion Impact:
  When production complete:
    Dr. Finished Goods Inventory
        Cr. WIP
    
  Cost flows: Purchased Materials → WIP → Finished Goods

Subcontracting Integration:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Scenario: Send materials to subcontractor for processing

Subcontract PO: PO-2025-00380
Subcontractor: Machining Services Ltd
Service: CNC Machining

Materials to Send:
  Raw Material: 500 kg steel
  Value: 75,000 KES

Process:
  1. Create Subcontract PO (service + materials)
  2. Issue materials from stock:
     Dr. Subcontractor Materials   75,000
         Cr. Inventory                     75,000
         
  3. Receive processed goods:
     Dr. Inventory (finished)      125,000
         Cr. Subcontractor Materials        75,000
         Cr. GRN Clearing (service)         50,000
         
  4. Invoice for service:
     Dr. GRN Clearing              50,000
     Dr. VAT Input                  8,000
         Cr. Accounts Payable              58,000
```

### Sales Module Integration

```markdown
INTER-MODULE WORKFLOWS

Purchase for Sales Order:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Scenario: Trading company - buy to sell

Sales Order: SO-2025-00789
Customer: XYZ Ltd
Item: Product B (not in stock)
Quantity: 100 Units

System Flow:
  1. Sales Order created (SO-2025-00789)
  2. Stock check: Item not available
  3. System suggests: Create Purchase Requisition
  4. PR auto-generated: PR-2025-00455
     Linked to: SO-2025-00789
     Item: Product B
     Quantity: 100 Units
     Required By: (SO delivery date - lead time)
     
  5. Procurement process:
     PR → RFQ → PO → GRN
     
  6. Once received:
     System notification to sales: Stock available
     
  7. Sales can now:
     Create Delivery Note → Invoice Customer

Costing Linkage:
  Purchase Price: 1,000 KES/Unit
  Sales Price: 1,500 KES/Unit
  Margin: 500 KES/Unit (33%)
  
  System tracks:
    Landed cost from purchase
    Allocation to sales order
    Profitability analysis

Drop Shipment:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Process: Supplier ships directly to customer

Sales Order: SO-2025-00800
Customer: ABC Corp
Item: Special Equipment
Quantity: 1 Unit

Purchase Order: PO-2025-00390
Supplier: Overseas Supplier
Delivery Address: ABC Corp (customer address)
Ship Via: Direct

Flow:
  1. SO created with "Drop Ship" flag
  2. PO generated automatically
     Delivery to: Customer address
     
  3. Supplier ships to customer:
     Document: Delivery note to customer
     
  4. Customer confirms receipt:
     Updates system
     
  5. Procurement records GRN:
     Virtual GRN (no warehouse entry)
     
  6. Sales creates invoice to customer
  
  7. Procurement processes supplier invoice

Accounting:
  Dr. Cost of Goods Sold         X
      Cr. Accounts Payable           X
  (No inventory movement)
  
  Dr. Accounts Receivable        Y
      Cr. Sales Revenue              Y
```

### Budget Module Integration

```markdown
BUDGET CONTROL INTEGRATION

Budget Check Points:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. Purchase Requisition Creation
   Trigger: PR submit
   Check:
     Department: Production
     Budget Code: PROD-RM-2025
     Annual Budget: 50,000,000 KES
     Spent YTD: 18,500,000 KES
     Committed (Open POs): 5,200,000 KES
     Available: 26,300,000 KES
     
     PR Amount: 1,500,000 KES
     
     Check Result: ✓ Within budget
     Action: Approve PR (auto if within budget)

2. Over-Budget Scenario
   PR Amount: 30,000,000 KES
   Available: 26,300,000 KES
   Variance: -3,700,000 KES (over budget)
   
   System Action:
     Status: HOLD - Budget Exceeded
     Notification: To Budget Controller
     Options:
       a) Request budget supplement
       b) Defer purchase
       c) Override (with CFO approval)

Budget Supplement Request:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Request: BSR-2025-008
Department: Production
Budget Code: PROD-RM-2025
Current Budget: 50,000,000 KES
Requested Addition: 5,000,000 KES
Revised Budget: 55,000,000 KES

Justification:
  - Increased sales orders
  - Additional production required
  - Material price increases

Approval Chain:
  Department Head → Finance Manager → CFO
  
If Approved:
  Budget updated in system
  Held PR released for processing

Budget Reporting:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Budget Utilization Report:

┌──────────────┬──────────┬──────────┬──────────┬──────────┬─────┐
│ Category     │ Budget   │ Spent    │ Committed│ Available│ %   │
├──────────────┼──────────┼──────────┼──────────┼──────────┼─────┤
│ Raw Material │50,000,000│18,500,000│ 5,200,000│26,300,000│ 47% │
│ Packaging    │25,000,000│ 6,200,000│ 2,100,000│16,700,000│ 33% │
│ MRO          │18,000,000│ 4,800,000│ 1,500,000│11,700,000│ 35% │
│ Services     │15,000,000│ 3,900,000│   800,000│10,300,000│ 31% │
└──────────────┴──────────┴──────────┴──────────┴──────────┴─────┘

Alerts:
  • No categories over 80% utilization ✓
  • Healthy budget position
```

---

**Next:** [Common Business Scenarios](./21-common-business-scenarios.md)
