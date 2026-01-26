[<-- Back to Index](README.md)

## Purchase Order Processing

### Purchase Order Creation

```markdown
PURCHASE ORDER CREATION FROM RFQ

Award Decision: Precision Machines Ltd
RFQ: RFQ-2025-0025
Award Date: 2025-04-10
Created By: Jane Muthoni (Procurement Officer)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Purchase Order Header:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PO Number: PO-2025-00234 (auto-generated)
PO Date: 2025-04-11
Status: DRAFT

Supplier Information:
  Supplier: Precision Machines Ltd
  Code: SUP-00234
  Contact: Sales Manager
  Email: sales@precisionmachines.co.ke
  Phone: +254 711 234 567

Delivery Information:
  Ship To: AWO Manufacturing Ltd
           Main Factory
           Industrial Area, Nairobi
           P.O. Box 12345-00100
  
  Warehouse: WH-NBI-01 (Main Warehouse)
  Contact: James Omondi (+254 720 123 456)

Billing Information:
  Bill To: AWO Manufacturing Ltd
           Head Office
           Industrial Area, Nairobi
           P.O. Box 12345-00100
  
  Tax PIN: P051234567X
  VAT No: 0512345678

PO Line Items:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Line 1:
  Item Code: CAP-CNC-001 (created for capital equipment)
  Description: Haas VF-2SS CNC Milling Machine
  Specifications:
    - 3-Axis Vertical Machining Center
    - Table: 1016mm x 508mm
    - Spindle: 8,100 RPM
    - Tool Changer: 24 tools (ATC)
    - Controller: Haas NGC
    - Country of Origin: USA
  
  Quantity: 1 Unit
  Unit Price: 78,000 USD
  Line Total: 78,000 USD

Line 2:
  Description: Installation & Commissioning
  Quantity: 1 Service
  Unit Price: 3,000 USD
  Line Total: 3,000 USD

Line 3:
  Description: Operator Training (5 days)
  Quantity: 1 Service
  Unit Price: 2,000 USD
  Line Total: 2,000 USD

Line 4:
  Description: Standard Tool Kit
  Quantity: 1 Set
  Unit Price: 1,500 USD
  Line Total: 1,500 USD

Line 5:
  Description: Freight to Nairobi
  Quantity: 1 Lot
  Unit Price: 4,500 USD
  Line Total: 4,500 USD

Line 6:
  Description: Import Clearance Services
  Quantity: 1 Service
  Unit Price: 2,000 USD
  Line Total: 2,000 USD

Line 7 (Value Addition):
  Description: Measurement Probe System (FOC)
  Quantity: 1 Unit
  Unit Price: 0 USD (Free - negotiated)
  Line Total: 0 USD
  Note: Valued at 3,500 USD

Line 8 (Value Addition):
  Description: HSK-63 Tool Holders (5 pcs) (FOC)
  Quantity: 5 Units
  Unit Price: 0 USD (Free - negotiated)
  Line Total: 0 USD
  Note: Valued at 600 USD

PO Summary:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Subtotal (USD): 91,000
  Less: Free Items: (4,100)
  Net Amount (USD): 86,900
  
  Exchange Rate: 130 KES/USD (locked)
  Amount (KES): 11,297,000
  
  VAT: Not applicable (capital equipment import)
  
  Total PO Value: 86,900 USD / 11,297,000 KES

Terms & Conditions:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Payment Terms:
  20% Advance: 17,380 USD upon PO (within 7 days)
  60% On Delivery: 52,140 USD upon delivery to site
  20% Final: 17,380 USD after successful commissioning
  
  Total: 86,900 USD

Payment Method: Wire Transfer
Bank Details: As per supplier master

Delivery Schedule:
  Manufacturing Lead Time: 8 weeks from advance payment
  Shipping: 2 weeks
  Clearance & Delivery: 2 weeks
  Total: 12 weeks from advance payment
  
  Expected Delivery: Week of July 8, 2025

Incoterms: DDP Nairobi (Delivered Duty Paid)
  Supplier responsible for:
    ✓ Export clearance
    ✓ International shipping
    ✓ Import duty and taxes
    ✓ Customs clearance
    ✓ Delivery to site

Warranty:
  Period: 18 months from commissioning
  Coverage: Parts and labor
  Response Time: 24 hours
  On-site Service: Included

Acceptance Criteria:
  1. Machine delivered complete and undamaged
  2. Installation completed per specifications
  3. Machine operational and producing parts
  4. Training completed (5 days, 3 operators minimum)
  5. All documentation provided (manuals, certificates)
  6. Performance test passed

Penalties:
  Late Delivery: 0.5% per week (max 5%)
  Performance Failure: Replacement or full refund

Special Instructions:
  ✓ Coordinate delivery 2 weeks in advance
  ✓ Foundation drawings provided
  ✓ Electrical connection ready
  ✓ Lifting equipment arranged by AWO
  ✓ Safety training required for installers

Attachments:
  - Site layout plan
  - Electrical specifications
  - Foundation drawings
  - Access route photos

Budget Information:
  Budget Code: CAPEX-2025-PROD
  Budget Allocated: 10,000,000 KES
  Additional Approved: 1,297,000 KES
  Total Authorized: 11,297,000 KES
  
  Commitment Created: 11,297,000 KES
```

### PO Approval Workflow

```markdown
PURCHASE ORDER APPROVAL

PO-2025-00234: Submitted for Approval
Date: 2025-04-11 10:00 AM
Value: 11,297,000 KES (> 5M threshold)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Approval Matrix (Capital Equipment > 5M):
  Level 1: Procurement Manager
  Level 2: Procurement Director
  Level 3: CFO

Level 1: Procurement Manager
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Approver: Paul Kariuki
Sent: 2025-04-11 10:01 AM
Reviewed: 2025-04-11 11:30 AM

Checklist:
  ☑ RFQ process completed (3 quotes received)
  ☑ Evaluation documented
  ☑ Budget approved (including variance)
  ☑ Supplier verified (approved supplier)
  ☑ Terms negotiated and documented
  ☑ Payment terms acceptable
  ☑ Delivery schedule reasonable

Decision: APPROVED
Comments: "PO aligns with awarded quotation. Terms negotiated 
           successfully. Budget variance pre-approved. Recommend approval."

Date: 2025-04-11 11:30 AM

Level 2: Procurement Director
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Approver: Michael Ndungu
Sent: 2025-04-11 11:31 AM
Reviewed: 2025-04-11 2:15 PM

Review Points:
  ☑ Strategic procurement aligned with production needs
  ☑ Competitive process followed
  ☑ Value for money demonstrated
  ☑ Risk mitigation (warranty, penalties) adequate
  ☑ Local support available

Decision: APPROVED
Comments: "Sound procurement process. Good negotiation results.
           Strategic supplier selected. Approved for CFO review."

Date: 2025-04-11 2:15 PM

Level 3: CFO
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Approver: David Ochieng
Sent: 2025-04-11 2:16 PM
Reviewed: 2025-04-12 9:00 AM

Financial Review:
  ☑ Budget allocation verified (CAPEX-2025-PROD)
  ☑ Cash flow impact assessed
  ☑ Payment schedule manageable
  ☑ Exchange rate locked appropriately
  ☑ ROI case reviewed
  ☑ Depreciation impact calculated

Decision: APPROVED
Comments: "Capital expenditure justified. Payment schedule aligned
           with cash flow projections. Exchange rate locked at
           favorable rate. FINAL APPROVAL GRANTED."

Date: 2025-04-12 9:00 AM

Final Status: FULLY APPROVED
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Approval Date: 2025-04-12 9:00 AM
PO Status: APPROVED → Ready to Send

Approval History:
┌─────────────────────┬──────────────┬──────────────┬─────────┐
│ Approver            │ Level        │ Date/Time    │Decision │
├─────────────────────┼──────────────┼──────────────┼─────────┤
│ Paul Kariuki        │ Proc Manager │ Apr 11 11:30 │Approved │
│ Michael Ndungu      │ Proc Director│ Apr 11 14:15 │Approved │
│ David Ochieng       │ CFO          │ Apr 12 09:00 │Approved │
└─────────────────────┴──────────────┴──────────────┴─────────┘

Next Actions:
  1. Issue PO to supplier
  2. Schedule advance payment
  3. Create receipt schedule
  4. Notify stakeholders
```

### PO Transmission to Supplier

```markdown
PURCHASE ORDER ISSUANCE

PO Status: APPROVED → ISSUED
Issue Date: 2025-04-12
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Transmission Method:
  ☑ Email (primary): sales@precisionmachines.co.ke
  ☑ Email (accounts): accounts@precisionmachines.co.ke
  ☑ Supplier Portal: Notification sent
  ☑ PDF Attachment: PO-2025-00234.pdf (digitally signed)

Email Content:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
To: sales@precisionmachines.co.ke
CC: accounts@precisionmachines.co.ke, procurement@awo.co.ke
Subject: Purchase Order PO-2025-00234 - CNC Machine

Dear Precision Machines Ltd,

Please find attached Purchase Order PO-2025-00234 for the supply
and installation of Haas VF-2SS CNC Milling Machine.

PO Summary:
  PO Number: PO-2025-00234
  Date: April 12, 2025
  Total Value: 86,900 USD (11,297,000 KES)
  
  Payment Schedule:
    Advance (20%): 17,380 USD - Due within 7 days of PO
    On Delivery (60%): 52,140 USD
    Final (20%): 17,380 USD - After commissioning

Please acknowledge receipt of this PO and confirm:
  1. Acceptance of all terms and conditions
  2. Delivery schedule (12 weeks from advance payment)
  3. Bank details for advance payment

Our payment reference: PO-2025-00234
Expected delivery: Week of July 8, 2025

For any queries, contact:
  Jane Muthoni - jane.muthoni@awo.co.ke - +254 720 345 678

Regards,
Paul Kariuki
Procurement Manager
AWO Manufacturing Ltd

Attachment: PO-2025-00234.pdf (digitally signed)

PO Status Updates:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
2025-04-12 10:00 AM: PO Sent to supplier
2025-04-12 10:05 AM: Email delivered successfully
2025-04-12 2:30 PM: Supplier acknowledged receipt (via portal)
2025-04-12 4:00 PM: Formal PO acknowledgment received

Supplier Acknowledgment:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
From: Precision Machines Ltd
Date: 2025-04-12

We acknowledge receipt of PO-2025-00234 and confirm:

  ✓ All terms accepted
  ✓ Delivery: 12 weeks from advance payment receipt
  ✓ Bank details confirmed (as per supplier master)
  ✓ Our reference: PM-ORDER-2025-089

Manufacturing will commence upon receipt of advance payment.
We will provide weekly progress updates.

Best regards,
Sales Manager
Precision Machines Ltd

PO Status: ACKNOWLEDGED
Next: Process advance payment
```

### PO Amendment Process

```markdown
PURCHASE ORDER AMENDMENT

Scenario: Specification Change Request
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Original PO: PO-2025-00156
Supplier: Ace Steel Suppliers Ltd
Original Items: Cold Rolled Steel - 7 Tons
Issue Date: 2025-03-17
Status: IN PRODUCTION

Amendment Request:
Date: 2025-03-25
Requested By: Peter Kimani (Production Manager)
Reason: Production plan revised, quantity adjustment needed

Change Details:
  Line 1 - Original:
    Item: RM-STL-001
    Quantity: 7 Tons
    Price: 85,000 KES/Ton
    Total: 595,000 KES

  Line 1 - Revised:
    Item: RM-STL-001
    Quantity: 10 Tons (increased)
    Price: 80,750 KES/Ton (volume discount applied)
    Total: 807,500 KES

  Impact:
    Original PO Value: 771,400 KES
    Revised PO Value: 983,900 KES
    Increase: 212,500 KES (27.5%)

Amendment Process:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Step 1: Internal Approval
  Budget Check: ✓ Available
  Department Head: ✓ Approved (Peter Kimani)
  Procurement Manager: ✓ Approved (Paul Kariuki)
  
  Reason: Value increase > 20% requires re-approval

Step 2: Supplier Negotiation
  Contact: Ace Steel Suppliers
  Discussion: Quantity increase to 10 Tons
  Result: Volume discount applied (5%)
  New Price: 80,750 KES/Ton (vs 85,000)
  Savings: 42,500 KES on new total

  Supplier Acceptance:
    Date: 2025-03-25
    Lead Time: Unchanged (can accommodate)
    Confirmation: Written acceptance received

Step 3: PO Amendment Creation
  Amendment No: PO-2025-00156-AMD-001
  Date: 2025-03-26
  Type: Quantity and Price Change

  Amendment Document:
  ┌────────────────────────────────────────────┐
  │ PURCHASE ORDER AMENDMENT                   │
  │                                            │
  │ Original PO: PO-2025-00156                │
  │ Amendment: PO-2025-00156-AMD-001          │
  │ Date: March 26, 2025                      │
  │                                            │
  │ This amendment modifies the original PO:  │
  │                                            │
  │ CHANGES:                                   │
  │ Line 1 - RM-STL-001                       │
  │   Quantity: 7 Tons → 10 Tons              │
  │   Unit Price: 85,000 → 80,750 KES         │
  │   Line Total: 595,000 → 807,500 KES       │
  │                                            │
  │ PO Total: 771,400 → 983,900 KES           │
  │                                            │
  │ All other terms remain unchanged.         │
  │                                            │
  │ Approved By: ________________             │
  │ Date: ________________                     │
  └────────────────────────────────────────────┘

Step 4: Transmission
  Sent To: Ace Steel Suppliers Ltd
  Date: 2025-03-26
  Method: Email + Portal

Step 5: Acknowledgment
  Supplier: Confirmed amendment
  Date: 2025-03-26
  Updated Delivery: Still on schedule

PO Version Control:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌─────────┬────────────┬──────────┬──────────┬───────────┐
│ Version │ Date       │ Quantity │ Value    │ Status    │
├─────────┼────────────┼──────────┼──────────┼───────────┤
│ 1.0     │ Mar 17     │ 7 Tons   │ 771,400  │Superseded │
│ 1.1     │ Mar 26     │ 10 Tons  │ 983,900  │ Current   │
└─────────┴────────────┴──────────┴──────────┴───────────┘

Amendment Types:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Quantity Change:
   Requires: Budget check, supplier acceptance
   Approval: If value change > 20%

2. Price Change:
   Requires: Justification, supplier agreement
   Approval: Procurement Manager minimum

3. Delivery Date Change:
   Requires: Department approval, supplier confirmation
   Approval: Procurement Officer

4. Specification Change:
   Requires: Technical review, supplier capability
   Approval: Department Head + Procurement

5. Terms Change:
   Requires: Financial review, legal review (if major)
   Approval: Procurement Director

6. Cancellation:
   Requires: Justification, supplier negotiation
   Approval: As per original PO approval level
   Note: May incur cancellation charges
```

### PO Tracking & Monitoring

```markdown
PURCHASE ORDER STATUS TRACKING

Dashboard View - Open Purchase Orders:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌──────────────┬───────────────┬──────────┬────────────┬─────────┐
│ PO Number    │ Supplier      │ Value    │ Due Date   │ Status  │
├──────────────┼───────────────┼──────────┼────────────┼─────────┤
│ PO-2025-00234│ Precision M.  │11,297,000│ 2025-07-08 │ In Prod │
│ PO-2025-00156│ Ace Steel     │  983,900 │ 2025-03-30 │ Overdue │
│ PO-2025-00189│ ChemSupply    │  245,000 │ 2025-04-15 │ On Track│
│ PO-2025-00201│ Office Supply │   85,000 │ 2025-04-20 │ On Track│
└──────────────┴───────────────┴──────────┴────────────┴─────────┘

Status Categories:
  ⚠ Overdue: Past expected delivery date, no goods received
  🟡 At Risk: Within 3 days of due date
  ✓ On Track: Within schedule
  🔵 In Production: Manufacturing/processing
  ⏸ On Hold: Awaiting action (payment, specs, etc.)

Detailed Tracking: PO-2025-00234
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PO: PO-2025-00234 (CNC Machine)
Status: IN PRODUCTION

Timeline:
┌────────────────────┬──────────────┬────────────┬──────────┐
│ Milestone          │ Planned Date │ Actual Date│ Status   │
├────────────────────┼──────────────┼────────────┼──────────┤
│ PO Issued          │ Apr 12       │ Apr 12     │ ✓ Done   │
│ Advance Paid       │ Apr 19       │ Apr 16     │ ✓ Done   │
│ Production Start   │ Apr 20       │ Apr 18     │ ✓ Done   │
│ Production Complete│ Jun 15       │ -          │ On Track │
│ Shipped            │ Jun 20       │ -          │ Pending  │
│ Arrival Mombasa    │ Jul 05       │ -          │ Pending  │
│ Clearance Complete │ Jul 07       │ -          │ Pending  │
│ Delivered to Site  │ Jul 08       │ -          │ Pending  │
│ Installation Done  │ Jul 15       │ -          │ Pending  │
│ Training Complete  │ Jul 20       │ -          │ Pending  │
│ Final Payment      │ Jul 25       │ -          │ Pending  │
└────────────────────┴──────────────┴────────────┴──────────┘

Progress Updates:
  Apr 12: PO issued and acknowledged
  Apr 16: Advance payment 17,380 USD sent ✓
  Apr 18: Production commenced (supplier confirmed)
  Apr 25: Progress update: 20% complete
  May 09: Progress update: 50% complete
  May 23: Progress update: 80% complete

Latest Update (May 23):
  "Machine assembly 80% complete. On schedule for Jun 15
   completion. Quality inspections passed. Shipping preparations
   underway."

Payment Tracking:
┌──────────────┬────────────┬──────────────┬────────────┐
│ Installment  │ Amount USD │ Planned Date │ Status     │
├──────────────┼────────────┼──────────────┼────────────┤
│ Advance 20%  │  17,380    │ Apr 19       │ ✓ Paid     │
│ Delivery 60% │  52,140    │ Jul 08       │ Pending    │
│ Final 20%    │  17,380    │ Jul 25       │ Pending    │
└──────────────┴────────────┴──────────────┴────────────┘

Risk Alerts:
  🟢 No risks identified
  ✓ On schedule
  ✓ No quality issues reported
  ✓ Supplier communication excellent
```

---

**Next:** [Goods Receipt & Inspection](./11-goods-receipt-and-inspection.md)
