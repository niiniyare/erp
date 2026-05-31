[<-- Back to Index](README.md)

## Purchase Invoicing

### Invoice Receipt & Entry

```markdown
PURCHASE INVOICE PROCESSING

Supplier Invoice Received:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Date: 2025-04-02
From: Ace Steel Suppliers Ltd
Method: Email (PDF attachment)

Invoice Details:
  Invoice No: AS-INV-2025-0456
  Invoice Date: 2025-04-01
  Supplier: Ace Steel Suppliers Ltd
  PIN: P098765432Y
  
Invoice Items:
┌──────────────┬──────────┬─────────┬────────────┬───────────┐
│ Description  │ Quantity │ UOM     │ Rate       │ Amount    │
├──────────────┼──────────┼─────────┼────────────┼───────────┤
│ Cold Rolled  │ 10.25    │ Ton     │ 80,750     │ 827,687.50│
│ Steel CR1    │          │         │            │           │
│ 1.0mm x 1000 │          │         │            │           │
│              │          │         │            │           │
│ Subtotal     │          │         │            │ 827,687.50│
│ VAT (16%)    │          │         │            │ 132,430.00│
│              │          │         │            │           │
│ TOTAL        │          │         │            │ 960,117.50│
└──────────────┴──────────┴─────────┴────────────┴───────────┘

Payment Terms: Net 30 Days
Due Date: 2025-05-01

Invoice Entry in System:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Purchase Invoice No: PINV-2025-00234 (system generated)
Entry Date: 2025-04-02
Entered By: Grace Akinyi (AP Clerk)

Header:
  Supplier: Ace Steel Suppliers Ltd (SUP-00123)
  Supplier Invoice: AS-INV-2025-0456
  Invoice Date: 2025-04-01
  Due Date: 2025-05-01 (30 days)
  Currency: KES
  
Line Items:
  Line 1:
    Item: RM-STL-001
    Description: Cold Rolled Steel
    Quantity: 10.25 Tons
    Rate: 80,750 KES/Ton
    Amount: 827,687.50 KES
    
  Tax:
    Type: VAT (16%)
    Amount: 132,430.00 KES
    
Total: 960,117.50 KES

Links:
  Purchase Order: PO-2025-00156
  Goods Receipt: GRN-2025-00145

Status: DRAFT (pending matching)
```

### Three-Way Matching

```markdown
3-WAY MATCHING PROCESS

Automatic Matching System:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Invoice: PINV-2025-00234
Initiated: 2025-04-02 2:30 PM

Matching Documents:
  1. Purchase Order: PO-2025-00156
  2. Goods Receipt: GRN-2025-00145
  3. Supplier Invoice: AS-INV-2025-0456

Comparison Matrix:
┌─────────────────┬──────────┬──────────┬──────────┬─────────┐
│ Field           │ PO       │ GRN      │ Invoice  │ Match   │
├─────────────────┼──────────┼──────────┼──────────┼─────────┤
│ Supplier        │ Ace Steel│ Ace Steel│ Ace Steel│ ✓ Yes   │
│ Item            │RM-STL-001│RM-STL-001│RM-STL-001│ ✓ Yes   │
│ Quantity        │ 10.00 T  │ 10.25 T  │ 10.25 T  │ ✓ Yes   │
│ Unit Price      │ 80,750   │ -        │ 80,750   │ ✓ Yes   │
│ Line Total      │ 807,500  │ -        │ 827,688  │ ⚠ Var   │
│ VAT (16%)       │ 129,200  │ -        │ 132,430  │ ⚠ Var   │
│ Total           │ 936,700  │ -        │ 960,118  │ ⚠ Var   │
└─────────────────┴──────────┴──────────┴──────────┴─────────┘

Variance Analysis:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Line Total Variance:
  PO Amount: 807,500 KES
  Invoice Amount: 827,688 KES
  Difference: 20,188 KES
  Percentage: 2.5%
  
Reason: Quantity variance (10.25 vs 10.00 Tons)
  Extra 0.25 Tons @ 80,750 = 20,188 KES
  
Analysis: Acceptable over-delivery (within ±3% tolerance)

VAT Variance:
  Expected: 129,200 KES (on 807,500)
  Actual: 132,430 KES (on 827,688)
  Difference: 3,230 KES
  
Reason: Follows the quantity variance (correct VAT calc)

Matching Result:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Status: ✓ MATCHED (with acceptable variance)

Tolerance Settings:
  Price Variance: ±2% (within)
  Quantity Variance: ±3% (within)
  
System Action: AUTO-APPROVED for payment
Approval Date: 2025-04-02 2:31 PM
Next: Schedule payment as per terms

Notification:
  ✓ AP Manager notified
  ✓ Added to payment schedule
  ✓ Supplier invoice accepted
```

### Invoice Variance Handling

```markdown
VARIANCE SCENARIOS

Scenario 1: Price Variance (Outside Tolerance)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Invoice: PINV-2025-00245
PO: PO-2025-00220

Issue:
  PO Price: 1,500 KES/Unit
  Invoice Price: 1,650 KES/Unit
  Variance: +150 KES/Unit (+10%)
  
  Tolerance: ±2%
  Status: ⚠ OUTSIDE TOLERANCE

System Action:
  Status: HOLD - Pending Review
  Routed To: Procurement Officer
  
Investigation:
  Date: 2025-04-05
  By: Jane Muthoni
  
  Findings:
    - Supplier price increased
    - No PO amendment issued
    - Price increase not authorized
    
Action:
  1. Contact Supplier:
     "Invoice price differs from PO. Please issue
      credit note for variance or confirm pricing error."
      
  2. Supplier Response:
     "Apologies. System error. Credit note being issued
      for the variance."
      
  3. Resolution:
     - Credit Note: AS-CN-2025-012 (for overcharge)
     - Invoice adjusted: System updated
     - Matching completed
     - Approved for payment

Scenario 2: Quantity Variance
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Invoice: PINV-2025-00256

Issue:
  PO Quantity: 100 Units
  GRN Quantity: 95 Units (5 rejected)
  Invoice Quantity: 100 Units ⚠
  
System Flag: Quantity Mismatch

Investigation:
  Reason: Supplier not informed of rejection
  
Action:
  1. Notify Supplier of rejection
  2. Request revised invoice for 95 units
  3. Hold payment until corrected
  
Resolution:
  - Revised Invoice received: AS-INV-2025-0467-R1
  - Quantity: 95 Units ✓
  - Matching completed
  - Payment approved

Scenario 3: No PO Reference
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Invoice: PINV-2025-00267
Supplier: Utility Company (Electricity)

Issue:
  No PO for utility bills (recurring service)
  Cannot perform 3-way matching
  
Process:
  1. Non-PO Invoice Entry:
     - Enter as non-PO invoice
     - Route to Department Head for approval
     - Expense account coding required
     
  2. Approval:
     Department: Administration
     Approver: Admin Manager
     Validation: Check against meter reading
     
  3. Accounting:
     Dr. Electricity Expense    250,000
     Dr. VAT Input               40,000
         Cr. Accounts Payable          290,000
         
  4. Payment: Scheduled as per terms
```

### Invoice Approval Workflow

```markdown
INVOICE APPROVAL PROCESS

Matched Invoice - Auto Approval:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Invoice: PINV-2025-00234
Amount: 960,118 KES

Matching Result: ✓ Perfect Match
Tolerance: Within limits

Workflow:
  1. 3-Way Match: ✓ Passed (Auto)
  2. System Approval: ✓ Approved (Auto)
  3. Payment Queue: ✓ Added
  4. No manual approval needed
  
Status: APPROVED FOR PAYMENT
Payment Due: 2025-05-01

Unmatched/Variance Invoice - Manual Approval:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Invoice: PINV-2025-00278
Amount: 1,250,000 KES
Issue: Price variance 5% (outside tolerance)

Workflow:
  1. 3-Way Match: ⚠ Variance detected
  2. Status: HOLD
  3. Route To: Procurement Manager
  
Approval Level 1: Procurement Manager
  Reviewer: Paul Kariuki
  Date: 2025-04-08
  
  Review:
    - Variance investigated
    - Supplier confirms pricing error
    - Credit note requested
    - Resolution documented
    
  Decision: APPROVED (pending credit note)
  
Approval Level 2: Finance Manager (if > 1M)
  Reviewer: Finance Manager
  Date: 2025-04-08
  
  Review:
    - Budget impact acceptable
    - Credit note tracking in place
    - Cash flow impact minimal
    
  Decision: APPROVED
  
Status: APPROVED FOR PAYMENT (with credit note tracking)

Non-PO Invoice - Department Approval:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Invoice: PINV-2025-00285
Supplier: Consultancy Firm
Amount: 500,000 KES
Type: Professional Services (No PO)

Workflow:
  1. Invoice Entry: By AP Clerk
  2. Department Coding: IT Department
  3. Route To: IT Director
  
Approval Level 1: Department Head (IT Director)
  Review:
    - Service rendered: ✓ Confirmed
    - Quality satisfactory: ✓ Yes
    - Amount as agreed: ✓ Correct
    - Budget available: ✓ Yes
    
  Decision: APPROVED
  Comments: "System implementation completed successfully"
  
Approval Level 2: Finance (if > 200K)
  Reviewer: Finance Manager
  Review:
    - Budget code verified
    - Amount within department budget
    
  Decision: APPROVED
  
Status: APPROVED FOR PAYMENT
```

### Invoice Posting

```markdown
FINANCIAL POSTING

Matched Invoice Posting:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Invoice: PINV-2025-00234
Date Posted: 2025-04-02
Status: APPROVED

Journal Entry (Automatic):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Date: 2025-04-02
Reference: PINV-2025-00234

Dr. GRN Clearing Account         827,687.50
Dr. VAT Input                    132,430.00
    Cr. Accounts Payable - Ace Steel     960,117.50

Description: Purchase Invoice AS-INV-2025-0456
PO: PO-2025-00156
GRN: GRN-2025-00145

Impact:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
GRN Clearing:
  Previous Balance: 827,687.50 Dr (from GRN)
  This Invoice: 827,687.50 Cr
  New Balance: 0 ✓ Cleared

Inventory:
  No change (already booked at GRN)
  
VAT Input:
  Increased: 132,430 KES (recoverable)
  
Accounts Payable:
  Supplier: Ace Steel Suppliers Ltd
  Previous Balance: 500,000 KES
  This Invoice: +960,117.50 KES
  New Balance: 1,460,117.50 KES

Payment Schedule:
  Invoice: PINV-2025-00234
  Amount: 960,117.50 KES
  Terms: Net 30
  Due Date: 2025-05-01
  Status: Scheduled

Service Invoice Posting:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Invoice: PINV-2025-00285 (Consultancy)
Amount: 500,000 KES + VAT

Journal Entry:
Dr. Professional Fees (Expense)  500,000
Dr. VAT Input                     80,000
    Cr. Accounts Payable - Consultant    580,000

Withholding Tax (5% on professional fees):
Dr. Accounts Payable             25,000
    Cr. WHT Payable                      25,000

Net Payment to Supplier: 555,000 KES
WHT Payable to KRA: 25,000 KES
```

---

**Next:** [Payment Processing](./13-payment-processing.md)
