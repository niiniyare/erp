[<-- Back to Index](README.md)

## Purchase Requisition

### Requisition Creation Process

```markdown
PURCHASE REQUISITION WORKFLOW

Requisition Initiation:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Requested By: Peter Kimani (Production Manager)
Department: Production
Cost Center: CC-PROD-01
Date: 2025-03-15
Required By: 2025-03-30

Purchase Requisition No: PR-2025-0125 (auto-generated)
Status: DRAFT

Requisition Header:
  Purpose: Production material requirement
  Project: Q2 Production Run
  Budget Code: PROD-RM-2025
  Priority: Normal
    Options: Normal / Urgent / Emergency

Item Details:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Line 1:
  Item Code: RM-STL-001
  Item Name: Cold Rolled Steel Sheet
  Description: CR1 Grade, 1.0mm thickness, 1000mm width
  Quantity: 5 Tons
  UOM: Ton
  Estimated Price: 85,000 KES/Ton
  Total Estimated: 425,000 KES

  Stock Information:
    Current Stock: 1.5 Tons
    Reorder Level: 2 Tons
    Available Stock: 1.5 Tons
    Committed: 0.5 Tons
    Free Stock: 1.0 Tons
    Status: ⚠ Below Reorder Level

  Suggested Supplier: Ace Steel Suppliers Ltd
  Last Purchase Price: 85,000 KES/Ton
  Last Purchase Date: 2025-02-28

  Justification: Production requirement for Job #2025-045
  Required Date: 2025-03-30
  Delivery Location: WH-NBI-01 (Main Warehouse)

Line 2:
  Item Code: MRO-LUB-015
  Item Name: Industrial Lubricant
  Quantity: 20 Liters
  UOM: Liter
  Estimated Price: 1,500 KES/Liter
  Total Estimated: 30,000 KES

  Stock Information:
    Current Stock: 15 Liters
    Status: ✓ Adequate

  Suggested Supplier: ChemSupply Ltd
  Last Purchase Price: 1,500 KES/Liter

Line 3:
  Item: Safety Gloves (Non-cataloged item)
  Description: Cut-resistant safety gloves, size L
  Quantity: 50 Pairs
  Estimated Price: 800 KES/Pair
  Total Estimated: 40,000 KES
  Category: PPE - Personal Protective Equipment

Requisition Summary:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Subtotal: 495,000 KES
  VAT (16%): 79,200 KES
  ─────────────────────
  Total: 574,200 KES

Budget Check:
  Budget Code: PROD-RM-2025
  Annual Budget: 50,000,000 KES
  Spent to Date: 19,900,000 KES
  This Requisition: 574,200 KES
  New Total: 20,474,200 KES
  Remaining: 29,525,800 KES
  Status: ✓ BUDGET AVAILABLE

Attachments:
  ☑ Production schedule (production_plan_q2.pdf)
  ☑ Material specification sheet (spec_sheet.pdf)

Submit for Approval: [Button]
```

### Requisition Approval Workflow

```markdown
APPROVAL PROCESS

PR-2025-0125: Requisition Submitted
Date/Time: 2025-03-15 10:30 AM
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Approval Level 1: Department Head
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Approver: James Ndungu (Production Director)
Sent: 2025-03-15 10:31 AM
Status: PENDING

Email Notification Sent:
  To: james.ndungu@awo.co.ke
  Subject: Purchase Requisition PR-2025-0125 Awaits Your Approval

  Content:
    Requisition: PR-2025-0125
    Requester: Peter Kimani
    Amount: 574,200 KES
    Items: Cold Rolled Steel (5 Tons), Lubricant, Safety Gloves
    Required Date: 2025-03-30

    [Approve] [Reject] [View Details]

Approval Action:
Date/Time: 2025-03-15 2:45 PM
Action: APPROVED
Comments: "Approved. Aligned with production schedule."

Approval Level 2: Procurement Manager
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Approver: Paul Kariuki (Procurement Manager)
Sent: 2025-03-15 2:46 PM
Status: PENDING

Approval Action:
Date/Time: 2025-03-15 4:15 PM
Action: APPROVED
Comments: "Budget verified. Preferred suppliers available. Approved for PO creation."

Final Status: APPROVED
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Approved Date: 2025-03-15 4:15 PM
Next Step: Create Purchase Order
Assigned To: Procurement Officer - Jane Muthoni

Notifications:
  ✓ Requester notified (Peter Kimani)
  ✓ Procurement team notified
  ✓ Budget commitment recorded

Approval History:
┌─────────────────────┬──────────┬────────────┬──────────┐
│ Approver            │ Level    │ Date/Time  │ Decision │
├─────────────────────┼──────────┼────────────┼──────────┤
│ James Ndungu        │ Dept Head│ Mar 15 2PM │ Approved │
│ Paul Kariuki        │ Proc Mgr │ Mar 15 4PM │ Approved │
└─────────────────────┴──────────┴────────────┴──────────┘
```

### Automatic Requisition Generation

```markdown
SYSTEM-GENERATED REQUISITIONS

Trigger: Reorder Point Reached
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Date: 2025-03-16 12:00 AM (Automated Daily Check)

Item: RM-STL-001 (Cold Rolled Steel Sheet)
Current Stock: 1.8 Tons
Reorder Level: 2.0 Tons
Status: ⚠ Below Reorder Point

System Action:
  ✓ Auto-generate requisition
  ✓ Calculate reorder quantity
  ✓ Assign to procurement
  ✓ Email notification

Auto-Generated Requisition:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PR-2025-0128 (Auto-generated)
Type: Reorder Point Replenishment
Generated: 2025-03-16 12:01 AM
Requester: System (Inventory Module)
Assigned To: Procurement Officer

Item: RM-STL-001
Quantity: 5 Tons (Economic Order Quantity)
Estimated Cost: 425,000 KES
Suggested Supplier: Ace Steel Suppliers Ltd
Required By: 2025-03-30 (Current Date + Lead Time)

Approval Workflow:
  Auto-approved for reorder point items < 500K
  Status: APPROVED
  Ready for PO creation

From Material Requirements Planning (MRP):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Trigger: Production Order Created

Production Order: JOB-2025-050
Product: Machine Model A
Quantity: 10 Units
Start Date: 2025-04-01

Bill of Materials Analysis:
┌──────────────┬─────────────┬──────────┬───────────┬─────────┐
│ Item         │ Required/Pc │ Total Req│ Available │ To Order│
├──────────────┼─────────────┼──────────┼───────────┼─────────┤
│ RM-STL-001   │ 0.5 Ton     │ 5 Tons   │ 1.8 Tons  │ 3.2 Tons│
│ COM-MOT-100  │ 1 Unit      │ 10 Units │ 5 Units   │ 5 Units │
│ COM-BRG-050  │ 4 Units     │ 40 Units │ 50 Units  │ 0 Units │
└──────────────┴─────────────┴──────────┴───────────┴─────────┘

Auto-Generated PRs:
  PR-2025-0129: RM-STL-001 (3.2 Tons → rounded to 5 Tons EOQ)
  PR-2025-0130: COM-MOT-100 (5 Units)

Status: Pending approval
Assigned: Production Planner for review
```

### Requisition Consolidation

```markdown
PR CONSOLIDATION FOR EFFICIENCY

Multiple Requisitions:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PR-2025-0125 (Peter - Production):
  Item: RM-STL-001, 5 Tons
  Required: 2025-03-30

PR-2025-0128 (System - Auto):
  Item: RM-STL-001, 5 Tons
  Required: 2025-03-30

PR-2025-0131 (Sarah - Maintenance):
  Item: RM-STL-001, 2 Tons
  Required: 2025-04-05

Consolidation Opportunity:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
System Alert: "3 requisitions for same item from same supplier"

Consolidation Recommendation:
  Combined Quantity: 12 Tons
  Supplier: Ace Steel Suppliers Ltd
  Unit Price: 85,000 KES/Ton
  Volume Discount: 5% (for > 10 Tons)

  Savings:
    Without Consolidation: 12 × 85,000 = 1,020,000 KES
    With 5% Discount: 12 × 80,750 = 969,000 KES
    Savings: 51,000 KES

Consolidated PR:
  PR-2025-0135 (Consolidated)
  Source PRs: PR-2025-0125, PR-2025-0128, PR-2025-0131
  Total Quantity: 12 Tons
  Total Value: 969,000 KES (after discount)
  Required By: 2025-03-30 (earliest date)

Allocation Plan:
  Delivery 1 (2025-03-30): 10 Tons
    → PR-2025-0125: 5 Tons
    → PR-2025-0128: 5 Tons

  Delivery 2 (2025-04-05): 2 Tons
    → PR-2025-0131: 2 Tons

Status: Approved
Action: Create single PO with split delivery
```

### Requisition Amendment

```markdown
REQUISITION MODIFICATION

Original Requisition:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PR-2025-0125
Status: APPROVED (Not yet converted to PO)

Original Item 1:
  Item: RM-STL-001
  Quantity: 5 Tons
  Required: 2025-03-30

Amendment Request:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Requested By: Peter Kimani
Date: 2025-03-16
Reason: Production schedule changed

Changes:
  Quantity: 5 Tons → 7 Tons
  Required Date: 2025-03-30 → 2025-03-28

Amendment Impact:
  Original Value: 574,200 KES
  New Value: 744,200 KES
  Difference: +170,000 KES

Budget Check:
  ✓ Budget still available

Re-approval Required:
  Reason: Value change > 100,000 KES
  Status: Pending re-approval
  Approver: James Ndungu (Department Head)

Amendment Approved:
  Date: 2025-03-16 3:00 PM
  PR Status: APPROVED (Amended)
  Next: Convert to PO

Version History:
┌─────────┬────────────┬──────────┬──────────┬──────────┐
│ Version │ Date       │ Quantity │ Value    │ Status   │
├─────────┼────────────┼──────────┼──────────┼──────────┤
│ 1.0     │ Mar 15     │ 5 Tons   │ 574,200  │ Approved │
│ 2.0     │ Mar 16     │ 7 Tons   │ 744,200  │ Approved │
└─────────┴────────────┴──────────┴──────────┴──────────┘
```

### Requisition to Purchase Order Conversion

```markdown
PR TO PO CONVERSION

Approved Requisition Ready for PO:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PR-2025-0125 (Version 2.0)
Approved: 2025-03-16 3:00 PM
Assigned To: Jane Muthoni (Procurement Officer)

Conversion Options:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Option 1: Direct PO Creation
  - Use suggested supplier (Ace Steel)
  - Use last purchase price
  - Fast-track process

Option 2: Request for Quotation
  - Create RFQ from requisition
  - Get competitive quotes
  - Select best supplier

Selected: Option 1 (Direct PO)
Reason:
  - Strategic supplier with contract
  - Price already negotiated
  - Urgent requirement

PO Creation Preview:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Purchase Order: PO-2025-00156 (draft)
Supplier: Ace Steel Suppliers Ltd
Order Date: 2025-03-17
Delivery Date: 2025-03-28

Source PRs:
  PR-2025-0125 (Line 1, 2, 3)

PO Items:
┌──────────────┬─────────┬─────────┬────────────┬───────────┐
│ Item         │ Qty     │ UOM     │ Price      │ Amount    │
├──────────────┼─────────┼─────────┼────────────┼───────────┤
│ RM-STL-001   │ 7       │ Ton     │ 85,000     │   595,000 │
│ MRO-LUB-015  │ 20      │ Liter   │  1,500     │    30,000 │
│ Safety Gloves│ 50      │ Pair    │    800     │    40,000 │
│              │         │         │            │           │
│ Subtotal     │         │         │            │   665,000 │
│ VAT (16%)    │         │         │            │   106,400 │
│ Total        │         │         │            │   771,400 │
└──────────────┴─────────┴─────────┴────────────┴───────────┘

Payment Terms: Net 30 Days
Delivery: WH-NBI-01, Main Warehouse
Incoterms: Ex-works

Action: Submit PO for approval
Next Approver: Paul Kariuki (Procurement Manager)

PR Status Update:
  PR-2025-0125: CONVERTED TO PO
  PO Reference: PO-2025-00156
  Conversion Date: 2025-03-17
```

### Requisition Types

```markdown
REQUISITION CATEGORIES

1. Stock Replenishment:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Purpose: Maintain inventory levels
Trigger: Reorder point, MRP
Approval: Often auto-approved
Example: Raw materials, MRO items

2. Project-Based:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Purpose: Specific project needs
Trigger: Project plan
Budget: Project budget code
Approval: Project manager + procurement
Example: New machine installation

3. Capital Expenditure:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Purpose: Fixed asset acquisition
Value: Usually > 1M KES
Approval: CFO + CEO (if > 5M)
Process: Business case, ROI analysis
Example: Machinery, vehicles

4. Service Requisition:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Purpose: Service contracts
Type: Maintenance, consultancy
Approval: Department head + procurement
Example: Annual maintenance contract

5. Emergency Requisition:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Purpose: Urgent unplanned needs
Priority: CRITICAL
Approval: Fast-track (Director level)
Process: Verbal PO, retroactive documentation
Example: Breakdown spares

6. Blanket Requisition:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Purpose: Recurring needs over period
Duration: Usually annual
Value: Estimated total
Call-off: As needed
Example: Office supplies, cleaning materials
```

---

**Next:** [Request for Quotation (RFQ)](./08-request-for-quotation.md)
