[<-- Back to Index](README.md)

## Purchase Returns & Debit Notes

### Purchase Return Process

```markdown
RETURN AUTHORIZATION

Return Initiation:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Date: 2025-04-10
Initiated By: Quality Control Department
Reason: Defective goods

Original Receipt:
  GRN: GRN-2025-0150
  PO: PO-2025-00198
  Supplier: ElectroTech Components
  Date: 2025-04-03
  Items: 100 Units electrical components

Issue Identified:
  Defect: 5 Units with wrong voltage specification
  Serial Numbers: #045, #067, #078, #089, #091
  Discovery: During installation

Return Details:
  Quantity: 5 Units
  Reason: Non-conformance to specification
  Original Value: 40,000 KES (5 × 8,000)

Return Authorization Process:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Step 1: Internal Approval
  Quality Manager: ✓ Approved return
  Warehouse Manager: ✓ Segregated items
  Procurement Officer: Notified

Step 2: Supplier Notification
  Date: 2025-04-10
  Method: Email with photos/test reports
  
  Content:
    "Dear ElectroTech,
    
    We need to return 5 units from GRN-2025-0150 due to
    specification mismatch (wrong voltage rating).
    
    Details: [Attached inspection report]
    
    Please arrange collection at your earliest convenience.
    
    Regards, AWO Manufacturing"

Step 3: Supplier Acknowledgment
  Date: 2025-04-11
  Response: "Return accepted. Collection scheduled Apr 13.
             Replacement units will be shipped immediately."

Step 4: Create Return Document
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Purchase Return No: PR-RET-2025-005
Date: 2025-04-13
Against: GRN-2025-0150, PO-2025-00198
Supplier: ElectroTech Components

Return Items:
┌────────┬──────────────────┬─────┬────────┬───────────┐
│ Item   │ Description      │ Qty │ Rate   │ Amount    │
├────────┼──────────────────┼─────┼────────┼───────────┤
│ EC-100 │ Control Module   │  5  │ 8,000  │  40,000   │
│        │ Serial: #045,etc │     │        │           │
│        │ Reason: Wrong    │     │        │           │
│        │ specification    │     │        │           │
└────────┴──────────────────┴─────┴────────┴───────────┘

Return Value: 40,000 KES
VAT: 6,400 KES
Total: 46,400 KES

Physical Return:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Collection Date: 2025-04-13
Collected By: ElectroTech driver
Warehouse: Issued from quarantine area
Gate Pass: GP-2025-0145

Checklist:
  ☑ Items counted: 5 Units ✓
  ☑ Serial numbers verified
  ☑ Packed securely
  ☑ Return note signed by driver
  ☑ Photos taken

Inventory Impact:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Stock Adjustment:
  Dr. Accounts Payable - ElectroTech   46,400
      Cr. Inventory - Components              40,000
      Cr. VAT Input                             6,400

Description: Return of defective items, PR-RET-2025-005

Inventory Balance:
  Opening: 95 Units (after initial 5 were segregated)
  Return: -5 Units (removed from stock)
  Closing: 95 Units (no change - already segregated)

Supplier Account Impact:
  Original Invoice: 800,000 KES (paid)
  Return Value: -46,400 KES
  Net Due from Supplier: 46,400 KES (credit balance)
```

### Debit Note Creation

```markdown
DEBIT NOTE ISSUANCE

Scenario 1: Return of Goods
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Following the return above:

Debit Note No: DN-2025-008
Date: 2025-04-13
Supplier: ElectroTech Components
Reason: Return of defective goods

Against Invoice: ET-INV-2025-234 (already paid)

┌────────────────────────────────────────────────────────┐
│               AWO MANUFACTURING LTD                    │
│               Industrial Area, Nairobi                 │
│                                                        │
│               DEBIT NOTE                               │
│                                                        │
│ Debit Note No: DN-2025-008                            │
│ Date: April 13, 2025                                  │
│                                                        │
│ To:                                                    │
│ ElectroTech Components Ltd                            │
│ Mombasa Road, Nairobi                                 │
│ PIN: P123456789Z                                      │
├────────────────────────────────────────────────────────┤
│ REASON: Return of defective goods                     │
│ Return Reference: PR-RET-2025-005                     │
│ Original Invoice: ET-INV-2025-234                     │
│                                                        │
│ Item: Control Module EC-100                           │
│ Quantity: 5 Units                                      │
│ Serial Nos: #045, #067, #078, #089, #091             │
│                                                        │
│ Defect: Wrong voltage specification                   │
│ GRN: GRN-2025-0150                                    │
│ PO: PO-2025-00198                                     │
│                                                        │
│                     Amount:      40,000 KES            │
│                     VAT (16%):    6,400 KES            │
│                     ─────────────────────              │
│                     TOTAL:       46,400 KES            │
│                                                        │
│ This amount will be:                                   │
│ ☑ Offset against replacement supply                   │
│ □ Refunded                                            │
│ □ Applied to future purchases                         │
│                                                        │
│ Authorized By: _______________                        │
│ Procurement Manager                                    │
│ Date: April 13, 2025                                  │
└────────────────────────────────────────────────────────┘

Supplier Action:
  Response: "Debit note acknowledged. Replacement units
             shipped today. Invoice will reflect credit."

Scenario 2: Price Correction (Overcharge)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Invoice: PINV-2025-00310
Supplier: Office Supplies Co
Issue: Invoiced at wrong (higher) price

Details:
  PO Price: 1,000 KES/Unit
  Invoice Price: 1,200 KES/Unit
  Quantity: 50 Units
  Overcharge: 200 × 50 = 10,000 KES

Debit Note: DN-2025-009
Reason: Price correction - invoiced above PO price

Amount:
  Price Difference: 10,000 KES
  VAT: 1,600 KES
  Total: 11,600 KES

Resolution:
  Supplier issues credit note: OSC-CN-2025-045
  Offset against debit note
  Net effect: Corrected pricing

Scenario 3: Damaged Goods (Partial Credit)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PO: PO-2025-00325
Item: Packaging Material
Issue: 10% of goods damaged in transit

Agreed Resolution:
  Full delivery accepted
  20% price reduction on damaged portion
  
  Original: 100 Units @ 500 = 50,000 KES
  Damaged: 10 Units (kept with discount)
  
  Discount on damaged: 10 × 500 × 20% = 1,000 KES

Debit Note: DN-2025-010
Amount: 1,000 KES + VAT = 1,160 KES
Reason: Damaged goods - partial credit
```

### Replacement Processing

```markdown
REPLACEMENT WORKFLOW

Defective Return with Replacement:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Original PO: PO-2025-00198
Return: PR-RET-2025-005 (5 Units defective)
Debit Note: DN-2025-008 (46,400 KES)

Replacement Supply:
  Date: 2025-04-18
  Quantity: 5 Units
  Supplier DN: ET-DN-2025-088
  GRN: GRN-2025-0175

Receipt Entry:
┌────────┬──────────────┬─────┬────────┬───────────┐
│ Item   │ Description  │ Qty │ Rate   │ Amount    │
├────────┼──────────────┼─────┼────────┼───────────┤
│ EC-100 │ Control Mod  │  5  │ 8,000  │  40,000   │
│        │ Replacement  │     │        │           │
└────────┴──────────────┴─────┴────────┴───────────┘

Value: 40,000 KES + VAT = 46,400 KES

Accounting Treatment:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Option A: Offset Against Debit Note (No New Invoice)
  
  Receipt:
  Dr. Inventory                     40,000
  Dr. VAT Input                      6,400
      Cr. Accounts Payable                  46,400
  
  Offset with Debit Note:
  Dr. Accounts Payable              46,400
      Cr. Accounts Payable                  46,400
  
  Net Effect: Zero (replacement at no charge)
  
Option B: New Invoice with Credit Reference
  
  Supplier Invoice: ET-INV-2025-345
  Amount: 46,400 KES
  Less: Credit from DN-2025-008: (46,400 KES)
  Net Due: 0 KES

Final Status:
  Original 100 Units: 95 Good + 5 Replaced
  Total Good Units: 100 ✓
  Cost: As per original PO ✓
  Supplier Account: Settled ✓
```

### Return Reporting

```markdown
PURCHASE RETURNS ANALYTICS

Period: Q1 2025
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Summary:
  Total Purchases: 25,000,000 KES
  Total Returns: 500,000 KES
  Return Rate: 2.0%

Returns by Reason:
┌──────────────────────┬──────────┬────────────┬─────────┐
│ Reason               │ Count    │ Value (KES)│ %       │
├──────────────────────┼──────────┼────────────┼─────────┤
│ Defective/Quality    │    12    │  250,000   │  50%    │
│ Wrong Item Shipped   │     5    │  120,000   │  24%    │
│ Damaged in Transit   │     4    │   80,000   │  16%    │
│ Specification Error  │     3    │   50,000   │  10%    │
│                      │          │            │         │
│ TOTAL                │    24    │  500,000   │ 100%    │
└──────────────────────┴──────────┴────────────┴─────────┘

Returns by Supplier:
┌─────────────────────┬───────────┬────────────┬─────────┐
│ Supplier            │ Purchases │ Returns    │ Rate    │
├─────────────────────┼───────────┼────────────┼─────────┤
│ ElectroTech         │ 2,000,000 │  200,000   │  10% ⚠  │
│ Office Supplies     │ 1,500,000 │   50,000   │   3.3%  │
│ Ace Steel           │ 5,000,000 │        0   │   0% ✓  │
│ ChemSupply          │ 3,000,000 │   90,000   │   3.0%  │
│ Others              │13,500,000 │  160,000   │   1.2%  │
└─────────────────────┴───────────┴────────────┴─────────┘

Action Items:
  ⚠ ElectroTech: High return rate (10%)
     → Quality review meeting scheduled
     → Supplier performance evaluation
     → Consider alternative suppliers

Resolution Tracking:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Returns: 24 cases
Replaced: 18 cases (75%)
Refunded: 4 cases (17%)
Credit Note: 2 cases (8%)

Average Resolution Time: 8 days
Target: < 10 days ✓
```

---

**Next:** [Pricing & Discount Management](./15-pricing-and-discount-management.md)
