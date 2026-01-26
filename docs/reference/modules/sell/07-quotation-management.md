[<-- Back to Index](README.md)

## Quotation Management

### Quotation Creation Process

```markdown
QUOTATION WORKFLOW

Step 1: Create Quotation
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Source:
  - New quotation (manual)
  - From Opportunity (auto-populated)
  - From previous quotation (revision)
  - From customer inquiry

Quotation Header:
  Quotation No: QTN-2025-001 (auto-generated)
  Date: 2025-01-20
  Valid Until: 2025-02-19 (30 days default)
  Customer: ABC Manufacturing Ltd
  Contact Person: John Kamau
  Opportunity: OPP-2025-001 (if linked)
  
Customer Details (auto-populated):
  Billing Address: [From customer master]
  Shipping Address: [Selectable if multiple]
  Price List: Corporate Pricing (from customer)
  Payment Terms: Net 30 Days
  Currency: KES
  
Sales Team:
  Sales Person: Sarah Johnson
  Territory: Nairobi Corporate
  Sales Manager: James Ndungu (for approval)

Step 2: Add Items
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Item Selection Process:
  1. Search/Select product from catalog
  2. System shows:
     - Current stock level
     - Standard price
     - Customer-specific price (if any)
     - Available quantity
     
Item Line Entry:
┌─────────────────────────────────────────────────┐
│ Item: Machine Model A                           │
│ Description: Industrial Machine - Model A       │
│ Quantity: 1                                     │
│ UOM: Unit                                       │
│                                                 │
│ Stock Available: 2 units ✓                     │
│                                                 │
│ Price List Rate: 2,200,000 KES                 │
│ Customer Discount: -10% (200,000)              │
│ Rate: 2,000,000 KES                            │
│                                                 │
│ Additional Discount: -5% (100,000)             │
│ Final Rate: 1,900,000 KES                      │
│                                                 │
│ Amount: 1,900,000 KES                          │
│                                                 │
│ Delivery Date: 2025-02-10                      │
│ Lead Time: 3 weeks                             │
└─────────────────────────────────────────────────┘

Multiple Items:
┌────────┬─────────────┬─────┬────────────┬────────────┐
│ Item   │ Description │ Qty │ Rate       │ Amount     │
├────────┼─────────────┼─────┼────────────┼────────────┤
│ MDL-A  │ Machine A   │  1  │ 1,900,000  │ 1,900,000  │
│ MDL-B  │ Machine B   │  1  │ 1,620,000  │ 1,620,000  │
│ MDL-C  │ Machine C   │  1  │ 1,080,000  │ 1,080,000  │
│ SVC-01 │ Installation│  1  │   150,000  │   150,000  │
│ TRN-01 │ Training    │  3  │    50,000  │   150,000  │
└────────┴─────────────┴─────┴────────────┴────────────┘

Step 3: Pricing & Discounts
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Automatic Price Calculation:
  1. Base Price from Price List
  2. Customer-specific discount applied
  3. Volume discount (if applicable)
  4. Promotional discount (if active)
  5. Manual override (with authorization)

Discount Approval Matrix:
  0-10%:   Sales Person (no approval)
  10-15%:  Sales Manager approval
  15-20%:  Sales Director approval
  >20%:    CFO approval

Example Calculation:
  Subtotal (Items):         4,900,000 KES
  Additional Discount (2%):   (98,000) KES
  ─────────────────────────────────────
  Net Amount:               4,802,000 KES
  VAT (16%):                  768,320 KES
  ─────────────────────────────────────
  Grand Total:              5,570,320 KES

Step 4: Terms & Conditions
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Select Templates:
  □ Payment Terms (Net 30 Days)
  □ Delivery Terms (3-4 weeks)
  □ Warranty Terms (12 months)
  □ Return Policy (30 days)
  
Custom Terms:
  - 50% deposit required for order confirmation
  - Balance payable on delivery
  - Installation within 1 week of delivery
  - Training provided within 2 weeks

Step 5: Additional Information
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Shipping Details:
  Shipping Rule: Customer pickup / Delivery
  Incoterms: EXW / FOB / CIF (if export)
  Expected Delivery: 2025-02-10
  
Notes (Internal):
  "Customer comparing with Competitor X.
   Price match required. Emphasize faster
   delivery and local support."
  
Notes (Customer Visible):
  "This quotation includes installation and
   training as discussed. Equipment will be
   delivered fully tested and certified."

Step 6: Review & Submit
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Pre-submission Checklist:
  ✓ Customer details verified
  ✓ Items and quantities confirmed
  ✓ Pricing approved (if discount >10%)
  ✓ Stock availability checked
  ✓ Delivery dates realistic
  ✓ Terms & conditions attached
  ✓ Print format reviewed

Submit Actions:
  1. Status changes: DRAFT → SUBMITTED
  2. Quotation PDF generated
  3. Email sent to customer
  4. Notification to sales manager
  5. Follow-up reminder scheduled (7 days)
```

### Quotation Document Structure

```markdown
QUOTATION PRINT FORMAT

┌────────────────────────────────────────────────┐
│              [COMPANY LOGO]                    │
│                                                │
│           QUOTATION                            │
│                                                │
│ Company Name               Quotation No: QTN-  │
│ Address Line 1             2025-001            │
│ Address Line 2             Date: Jan 20, 2025  │
│ Phone: xxx                 Valid: Feb 19, 2025 │
│ Email: xxx                                     │
│ Website: xxx                                   │
├────────────────────────────────────────────────┤
│ QUOTATION TO:                                  │
│                                                │
│ ABC Manufacturing Ltd                          │
│ Industrial Area, Nairobi                       │
│ Attention: John Kamau, Procurement Manager     │
│ Email: john.kamau@abc.com                      │
│ Phone: +254-700-123-456                        │
├────────────────────────────────────────────────┤
│ Dear Mr. Kamau,                                │
│                                                │
│ Thank you for your inquiry. We are pleased     │
│ to quote as follows:                           │
├────────────────────────────────────────────────┤
│ Item Description         Qty  Rate     Amount  │
│ ────────────────────     ──── ──────   ──────  │
│ 1. Machine Model A        1   1,900,000        │
│    Industrial Machine                1,900,000 │
│    • Power: 10HP                              │
│    • Capacity: 1000 units/hr                  │
│    • Warranty: 12 months                      │
│                                                │
│ 2. Machine Model B        1   1,620,000        │
│    Industrial Machine                1,620,000 │
│                                                │
│ 3. Machine Model C        1   1,080,000        │
│    Industrial Machine                1,080,000 │
│                                                │
│ 4. Installation Service   1     150,000        │
│    Professional Setup               150,000    │
│                                                │
│ 5. Training               3 days  50,000       │
│    Operator Training                150,000    │
│                                                │
│                          Subtotal: 4,900,000   │
│                          Discount:   (98,000)  │
│                          Net:      4,802,000   │
│                          VAT(16%):   768,320   │
│                          ───────────────────   │
│                          TOTAL:    5,570,320   │
│                                                │
│ Amount in Words:                               │
│ Five Million Five Hundred Seventy Thousand     │
│ Three Hundred Twenty Shillings Only            │
├────────────────────────────────────────────────┤
│ PAYMENT TERMS:                                 │
│ • 50% deposit upon order confirmation          │
│ • Balance payable upon delivery                │
│ • Net 30 days from invoice date               │
│                                                │
│ DELIVERY:                                      │
│ • 3-4 weeks from order confirmation            │
│ • Delivery to your factory warehouse           │
│ • Installation within 1 week of delivery       │
│                                                │
│ WARRANTY:                                      │
│ • 12 months comprehensive warranty             │
│ • Free technical support during warranty       │
│ • Parts replacement as needed                  │
│                                                │
│ VALIDITY:                                      │
│ This quotation is valid for 30 days            │
│                                                │
│ TERMS & CONDITIONS:                            │
│ [Standard terms attached]                      │
├────────────────────────────────────────────────┤
│ We trust our quotation meets your requirements │
│ and look forward to serving you.               │
│                                                │
│ Please contact us for any clarifications.      │
│                                                │
│ Best regards,                                  │
│                                                │
│ Sarah Johnson                                  │
│ Sales Executive                                │
│ sarah.johnson@company.com                      │
│ +254-700-555-001                               │
└────────────────────────────────────────────────┘
```

### Quotation Management

```markdown
QUOTATION STATUS TRACKING

Status Flow:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
DRAFT → SUBMITTED → ACCEPTED/REJECTED/EXPIRED

DRAFT:
  - Quotation being prepared
  - Can be edited freely
  - Not visible to customer
  - Not counted in pipeline

SUBMITTED:
  - Sent to customer
  - Email notification sent
  - Follow-up scheduled
  - Read-only (no edits without revision)

ACCEPTED:
  - Customer accepts quotation
  - Ready to convert to Sales Order
  - Success metric tracked

REJECTED:
  - Customer declines
  - Loss reason recorded
  - Analysis for improvement

EXPIRED:
  - Valid until date passed
  - Can be extended or revised
  - Automatic status update

Quotation Actions:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
From DRAFT:
  □ Submit to Customer
  □ Delete
  □ Create Revision

From SUBMITTED:
  □ Create Revision (new version)
  □ Convert to Sales Order (if accepted)
  □ Mark as Lost (if rejected)
  □ Extend Validity
  □ Send Reminder Email

From ACCEPTED:
  □ Convert to Sales Order ⭐
  □ Create Proforma Invoice

Quotation Revisions:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Original: QTN-2025-001 v1.0
  Submitted: Jan 20
  Amount: 5,570,320 KES
  Status: Customer requested revision

Revision 1: QTN-2025-001 v1.1
  Date: Jan 22
  Changes: 
    - Removed Item C
    - Added express delivery
  Amount: 4,890,320 KES
  Status: SUBMITTED

Revision 2: QTN-2025-001 v1.2
  Date: Jan 25
  Changes:
    - Additional 2% discount
  Amount: 4,792,513 KES
  Status: ACCEPTED ✓

All versions maintained for audit trail.

Quotation Analytics:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Total Quotations: 156
  Submitted: 156 (100%)
  Accepted: 78 (50%)
  Rejected: 45 (29%)
  Expired: 33 (21%)

Conversion Rate: 50%
Average Days to Close: 12 days
Average Quotation Value: 3.5M KES
Win Rate by Territory:
  Nairobi: 55%
  Mombasa: 48%
  Kisumu: 42%

Top Loss Reasons:
  1. Price too high (40%)
  2. Chose competitor (30%)
  3. Project cancelled (20%)
  4. Timeline too long (10%)
```

---