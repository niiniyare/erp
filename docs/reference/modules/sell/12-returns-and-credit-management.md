[<-- Back to Index](README.md)

## 12. Returns & Credit Management

### Sales Return Process

```markdown
SALES RETURN WORKFLOW

Return Authorization:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Customer Request:
  Customer: ABC Manufacturing Ltd
  Original Invoice: INV-2025-001
  Reason: Product defective
  Items: Machine Model B (Qty: 1)
  Value: 1,620,000 KES

Return Policy Check:
  ✓ Within 30-day return window
  ✓ Original packaging intact
  ✓ Proof of purchase (invoice) provided
  ✓ Not a custom/special order
  
Create Return Authorization:
  RMA No: RMA-2025-001
  Date: 2025-03-15
  Customer: ABC Manufacturing Ltd
  Status: PENDING APPROVAL

Return Authorization Form:
┌────────────────────────────────────────────┐
│ RETURN MERCHANDISE AUTHORIZATION           │
│                                            │
│ RMA No: RMA-2025-001                       │
│ Date: March 15, 2025                       │
│ Customer: ABC Manufacturing Ltd            │
│                                            │
│ Original Invoice: INV-2025-001             │
│ Invoice Date: February 10, 2025            │
│                                            │
│ Items to Return:                           │
│ - Machine Model B                          │
│ - Serial No: 12346                         │
│ - Quantity: 1                              │
│ - Value: 1,620,000 KES + VAT              │
│                                            │
│ Reason: Product defective                  │
│ Description: Motor not functioning         │
│                                            │
│ Return Method:                             │
│ □ Customer Drop-off                        │
│ ☑ Company Pickup                           │
│                                            │
│ Approved By: ________________              │
│ Date: ________________                     │
└────────────────────────────────────────────┘

Approval Workflow:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
RMA Request → Sales Manager Review → Approved/Rejected

If Approved:
  1. Schedule pickup from customer
  2. Notify warehouse to expect return
  3. Email customer with RMA number
  4. Provide return instructions

Physical Return Process:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Pickup/Delivery:
  Date: March 18, 2025
  Collected from: Factory Warehouse, Ruiru
  Collected by: Peter Maina (Driver)
  Condition on Pickup: Packed in original crate

Receiving at Warehouse:
  1. Verify RMA number
  2. Inspect item condition
  3. Take photos
  4. Check serial number matches
  5. Complete inspection report

Inspection Report:
┌────────────────────────────────────────────┐
│ Item: Machine Model B                      │
│ Serial No: 12346                           │
│                                            │
│ Physical Condition: Good                   │
│ Packaging: Original, intact                │
│ Accessories: All present                   │
│                                            │
│ Defect Verification: ✓ Confirmed          │
│ Issue: Motor shaft broken                  │
│                                            │
│ Recommendation:                            │
│ ☑ Accept Return (Replace)                 │
│ □ Accept Return (Repair)                   │
│ □ Accept Return (Refund)                   │
│ □ Reject Return                            │
│                                            │
│ Inspected by: John (QC)                    │
│ Date: March 18, 2025                       │
└────────────────────────────────────────────┘

Create Delivery Return:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Delivery Return No: DR-2025-001
Date: March 18, 2025
Against Delivery: DN-2025-001
RMA: RMA-2025-001

Items Returned:
┌────────────────────────┬──────┬────────────┐
│ Item                   │ Qty  │ Rate       │
├────────────────────────┼──────┼────────────┤
│ Machine Model B        │  1   │ 1,620,000  │
│ SN: 12346              │      │            │
└────────────────────────┴──────┴────────────┘

Inventory Impact:
  Returned to Stock: Yes/No
  
  If Yes (Good Condition):
    Cr. COGS                    1,400,000
        Dr. Inventory                   1,400,000
  
  If No (Defective):
    Move to: Defective Stock
    No accounting entry (not saleable)

Issue Credit Note:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Credit Note No: CN-2025-001
Date: March 18, 2025
Against Invoice: INV-2025-001
Reason: Product defective - returned

Items:
┌────────────────────────┬──────┬────────────┐
│ Item                   │ Qty  │ Amount     │
├────────────────────────┼──────┼────────────┤
│ Machine Model B        │  1   │ 1,620,000  │
│ VAT @ 16%              │      │   259,200  │
│                        │      │            │
│ TOTAL CREDIT           │      │ 1,879,200  │
└────────────────────────┴──────┴────────────┘

Accounting Entry:
  Dr. Sales Returns (Contra-Revenue)  1,620,000
  Dr. VAT Payable                       259,200
      Cr. Accounts Receivable - ABC         1,879,200

Customer Account:
  Original Invoice: +1,879,200 (DR)
  Credit Note: -1,879,200 (CR)
  Net Effect: 0 (invoice portion cancelled)

Resolution Options:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Option 1: Replacement (Most Common)
  1. Issue credit note
  2. Create new sales order
  3. Deliver replacement unit
  4. New invoice at same/discounted price
  5. Customer may pay difference if price changed

Option 2: Refund
  1. Issue credit note
  2. Process refund payment
  3. Entry:
     Dr. Accounts Receivable      1,879,200
         Cr. Bank                         1,879,200

Option 3: Credit Balance
  1. Issue credit note
  2. Keep as customer credit balance
  3. Apply to future purchases

Option 4: Repair & Return
  1. Repair defective unit
  2. Return to customer
  3. No credit note (warranty service)

Customer Choice:
  Selected: REPLACEMENT
  New Order: SO-2025-085
  New Machine: Model B (Serial 15789)
  Price: Same as original (goodwill)
  Delivery: March 25, 2025
```

### Return Policies

```markdown
RETURN POLICY CONFIGURATION

Standard Return Policy:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Eligibility:
  ✓ Within 30 days of delivery
  ✓ Original packaging and labels
  ✓ Unused and resaleable condition
  ✓ With original invoice/receipt
  ✓ Not damaged by customer

Restocking Fee:
  - 0% if defective (manufacturer fault)
  - 10% if customer error (wrong item ordered)
  - 15% if opened/used but functional
  - 100% (no return) if custom order

Refund Method:
  - Original payment method
  - Store credit (if preferred)
  - Replacement (most common)

Processing Time:
  - Return authorization: 1-2 business days
  - Refund processing: 5-7 business days after receipt
  - Replacement: Within lead time

Non-Returnable Items:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  ✗ Custom-manufactured items
  ✗ Perishable goods
  ✗ Software licenses (once activated)
  ✗ Installed equipment
  ✗ Clearance/final sale items
  ✗ Items marked "non-returnable"

Warranty vs. Return:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Return (within 30 days):
  - Customer decides they don't want it
  - Or item defective
  - Full refund/replacement
  - Customer bears return shipping (unless defect)

Warranty (after 30 days, within warranty period):
  - Manufacturer defect
  - Covered by warranty terms
  - Repair or replace only
  - No refund
  - Company handles return shipping

Policy by Customer Type:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Retail Customers:
  Return Period: 14 days
  Restocking: 15%
  Refund: Original payment method

Corporate Customers:
  Return Period: 30 days
  Restocking: 10% (negotiable)
  Refund: Net 15 days from credit note

Wholesale/Distributors:
  Return Period: 60 days (for unsold stock)
  Restocking: 5%
  Replacement preferred
  No refund on promotional items
```

### Credit Note Management

```markdown
CREDIT NOTE ISSUANCE

Reasons for Credit Notes:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Sales Return
   Customer returns goods
   
2. Price Adjustment
   Overcharge on invoice
   Discount given post-invoice
   
3. Damaged/Defective Goods
   Quality issues discovered after delivery
   
4. Invoice Error
   Wrong quantity billed
   Wrong price charged
   Duplicate billing
   
5. Promotional Credit
   Marketing promotion (e.g., "Buy 10, get 1 free")
   
6. Goodwill Gesture
   Compensation for poor service
   Late delivery compensation

Credit Note Types:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Type 1: Against Specific Invoice
  Links to original invoice
  Reduces invoice balance
  Most common

Type 2: Standalone Credit
  Not linked to specific invoice
  Creates credit balance
  Applied to future invoices

Type 3: Refund Credit Note
  Requires cash refund
  Immediate payment to customer

Credit Note Process:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Step 1: Identify Need
  Reason documented
  Supporting evidence (photos, delivery note)
  
Step 2: Get Approval
  Sales manager approval (< 100K)
  Finance manager (100K - 500K)
  CFO (> 500K)

Step 3: Create Credit Note
  Credit Note No: CN-2025-001
  Date: March 18, 2025
  Against: INV-2025-001
  
Step 4: Post to Accounts
  Accounting entry auto-created
  Customer balance updated
  
Step 5: Notify Customer
  Email with credit note PDF
  Statement showing updated balance

Credit Note Example:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌────────────────────────────────────────────┐
│         [COMPANY LOGO]                     │
│                                            │
│        CREDIT NOTE                         │
│                                            │
│ Credit Note No: CN-2025-001                │
│ Date: March 18, 2025                       │
│                                            │
│ Original Invoice: INV-2025-001             │
│ Invoice Date: February 10, 2025            │
├────────────────────────────────────────────┤
│ CUSTOMER:                                  │
│ ABC Manufacturing Ltd                      │
│ Industrial Area, Nairobi                   │
│ PIN: P000123456A                           │
├────────────────────────────────────────────┤
│ REASON: Product return - defective        │
│ RMA No: RMA-2025-001                       │
├────────────────────────────────────────────┤
│ Item          Qty   Rate        Amount     │
│ ────────────  ───   ─────────   ────────   │
│ Machine B      1    1,620,000   1,620,000  │
│                                            │
│                     Subtotal:   1,620,000  │
│                     VAT(16%):     259,200  │
│                     ───────────────────    │
│                     TOTAL:      1,879,200  │
│                                            │
│ Amount in Words:                           │
│ One Million Eight Hundred Seventy Nine     │
│ Thousand Two Hundred Shillings Only        │
├────────────────────────────────────────────┤
│ This amount has been credited to your      │
│ account and will be reflected in your      │
│ next statement.                            │
│                                            │
│ Authorized by: _______________            │
│ Date: _______________                      │
└────────────────────────────────────────────┘

Impact on Customer Account:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Before Credit Note:
  Outstanding Balance: 2,807,416 KES

After Credit Note (CN-2025-001):
  Previous Balance: 2,807,416 KES
  Less Credit: -1,879,200 KES
  ───────────────────────────────
  New Balance: 928,216 KES
```

### Handling Partial Returns

```markdown
PARTIAL RETURN SCENARIOS

Scenario: Multi-Item Invoice, Partial Return
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Original Invoice: INV-2025-001
Items:
  - Machine A: 1,900,000 KES
  - Machine B: 1,620,000 KES
  - Installation: 150,000 KES
  Total: 3,670,000 + VAT = 4,255,200 KES

Customer Returns:
  - Machine B only (defective)
  - Keeps Machine A and Installation

Process:
1. Create Return for Machine B only
   
2. Credit Note: CN-2025-001
   Item: Machine B
   Amount: 1,620,000 KES
   VAT: 259,200 KES
   Total Credit: 1,879,200 KES

3. Updated Invoice Status:
   Original: 4,255,200 KES
   Credit: -1,879,200 KES
   Net: 2,376,000 KES (for items kept)

4. Customer chooses:
   Option A: Replacement Machine B
     → New sales order
     → Deliver replacement
     → New invoice OR apply credit

   Option B: Keep credit
     → Balance: 2,376,000 DR (owed to us)
     → Credit: 1,879,200 CR (owed to them)
     → Net Balance: 496,800 DR

Quantity Returns:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Invoice:
  Item: Widget X
  Quantity: 100 units
  Rate: 10,000 KES/unit
  Total: 1,000,000 KES

Customer Returns: 20 units (defective)

Credit Note:
  Item: Widget X
  Quantity: 20 units
  Rate: 10,000 KES/unit
  Credit: 200,000 + VAT = 232,000 KES

Net Purchase:
  Original: 100 units = 1,160,000 KES
  Return: 20 units = -232,000 KES
  Net: 80 units = 928,000 KES
```

---

