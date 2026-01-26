[<-- Back to Index](README.md)

## 20. Common Business Scenarios

### Scenario 1: Walk-In Cash Sale

```markdown
SIMPLE RETAIL TRANSACTION

Customer walks in, buys, pays, leaves.

Process:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Step 1: Customer Inquiry
  Customer: "How much for this laptop?"
  Staff: Checks system for price

Step 2: Create Sales Invoice (Direct)
  No quotation, no sales order
  Direct invoice creation
  
  Invoice: INV-2025-100
  Customer: Walk-in Customer (generic)
  Item: Laptop Model X
  Quantity: 1
  Price: 80,000 KES
  VAT: 12,800 KES
  Total: 92,800 KES
  
  Payment: Cash / M-Pesa

Step 3: Receive Payment
  Customer pays immediately
  Payment Entry: PAY-2025-100
  Amount: 92,800 KES
  Method: M-Pesa
  
  Allocated to: INV-2025-100

Step 4: Delivery
  Hand over laptop
  Customer signs delivery note
  Inventory updated (stock reduced)

Step 5: Receipt
  Print receipt for customer
  Transaction complete

Timeline: 10-15 minutes

Accounting Impact:
  Dr. M-Pesa Account              92,800
      Cr. Sales Revenue                  80,000
      Cr. VAT Payable                    12,800
  
  Dr. COGS                        55,000
      Cr. Inventory                      55,000

Optional: Capture customer details for future marketing
```

### Scenario 2: Corporate Credit Sale

```markdown
STANDARD B2B TRANSACTION

Established corporate customer, credit terms.

Process:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Step 1: Customer Inquiry
  Email: "Need 50 units of Product X"

Step 2: Check Stock & Price
  Available: 60 units ✓
  Customer Price: 45,000 KES/unit
  (Corporate discount already applied)

Step 3: Create Quotation
  Quote: QTN-2025-150
  Items: Product X × 50
  Total: 2,250,000 + VAT = 2,610,000 KES
  Valid: 14 days
  Terms: Net 30 Days

Step 4: Send Quotation
  Email to customer
  Customer reviews

Step 5: Customer Accepts
  Emails Purchase Order: PO/CORP/2025/088
  Or: Emails acceptance

Step 6: Create Sales Order
  SO-2025-150
  From QTN-2025-150
  Customer PO: PO/CORP/2025/088
  
  System checks:
    ✓ Stock available
    ✓ Credit limit OK
    ✓ No overdue invoices
  
  Status: CONFIRMED
  Stock reserved: 50 units

Step 7: Schedule Delivery
  Delivery date: 3 days
  Coordinate with warehouse

Step 8: Pick & Pack
  Warehouse receives order
  Picks 50 units
  Packs in 10 boxes
  Labels each box

Step 9: Deliver
  Delivery Note: DN-2025-150
  Truck delivers to customer
  Customer signs
  Delivery confirmed in system

Step 10: Create Invoice
  Auto-created from delivery
  Invoice: INV-2025-150
  Amount: 2,610,000 KES
  Due: 30 days (Feb 20)
  
  Accounting:
    Dr. AR - Corporate Ltd        2,610,000
        Cr. Sales Revenue                 2,250,000
        Cr. VAT Payable                     360,000

Step 11: Payment Collection
  Due date: Feb 20
  Reminder: Feb 18 (auto-email)
  Payment received: Feb 22
  
  Payment: PAY-2025-150
  Amount: 2,610,000 KES
  Method: Bank Transfer
  
  Accounting:
    Dr. Bank                      2,610,000
        Cr. AR - Corporate Ltd            2,610,000

Transaction complete.
Timeline: 7-10 days (from quote to payment)
```

### Scenario 3: Custom Order with Deposit

```markdown
SPECIAL ORDER WORKFLOW

Customer needs custom/made-to-order product.

Process:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Step 1: Initial Inquiry
  Customer: "Need custom machine with
            specific modifications"

Step 2: Technical Consultation
  Sales Engineer meets customer
  Specs documented
  Feasibility confirmed

Step 3: Cost Estimation
  Engineering: Design cost
  Manufacturing: Production cost
  Materials: Special components
  Total Cost: 8,000,000 KES
  Target Margin: 25%
  Selling Price: 10,000,000 KES

Step 4: Quotation
  QTN-2025-200
  Custom Machine - Spec Sheet Attached
  Price: 10,000,000 KES + VAT
  Validity: 30 days
  
  Special Terms:
    - 50% deposit on order
    - 30% on delivery
    - 20% after commissioning
  
  Lead Time: 12 weeks

Step 5: Customer Accepts
  Provides PO
  Ready to pay deposit

Step 6: Sales Order
  SO-2025-200
  Value: 10,000,000 KES
  Type: Custom Manufacturing
  
  Payment Schedule:
    ├─ Deposit (50%): 5,000,000 KES
    ├─ Delivery (30%): 3,000,000 KES
    └─ Final (20%): 2,000,000 KES

Step 7: Receive Deposit
  Invoice for Deposit:
    INV-2025-200-DEP
    Description: Advance Payment - 50%
    Amount: 5,800,000 KES (incl VAT)
  
  Payment: PAY-2025-200-DEP
  Amount: 5,800,000 KES
  
  Accounting:
    Dr. Bank                      5,800,000
        Cr. Customer Advances             5,800,000

Step 8: Production
  Work Order created
  12-week manufacturing
  Project tracking
  Weekly updates to customer

Step 9: Delivery
  Machine ready
  Quality inspection
  Delivery to customer site
  
  Delivery Note: DN-2025-200

Step 10: Invoice for Delivery Payment
  INV-2025-200-DEL
  Description: On Delivery - 30%
  Amount: 3,480,000 KES (incl VAT)
  
  Payment: PAY-2025-200-DEL
  Amount: 3,480,000 KES

Step 11: Installation & Commissioning
  2 weeks on-site work
  Testing
  Training
  Customer acceptance

Step 12: Final Invoice
  INV-2025-200-FINAL
  Description: Final Payment - 20%
  Amount: 2,320,000 KES (incl VAT)
  
  Accounting (consolidate):
    Dr. AR - Customer             11,600,000
        Cr. Customer Advances              5,800,000
        Cr. Sales Revenue                 10,000,000
        Cr. VAT Payable                    1,600,000
  
  As payments come:
    Dr. Bank
        Cr. AR - Customer

Transaction complete.
Timeline: 16 weeks (from order to final payment)
```

### Scenario 4: Handling Returns

```markdown
CUSTOMER RETURN SCENARIO

Customer wants to return defective product.

Process:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Step 1: Return Request
  Customer calls: "Product X not working"
  Original Invoice: INV-2025-050
  Purchase Date: Jan 15
  Days since purchase: 10 days ✓ (within policy)

Step 2: Troubleshooting
  Support team tries remote fix
  Unable to resolve
  Confirms defect

Step 3: Return Authorization
  Create RMA: RMA-2025-005
  Reason: Defective - motor failure
  Approved: Yes
  Action: Replace

Step 4: Return Logistics
  Options:
    A) Customer brings to office
    B) Company picks up
  
  Selected: Option B
  Pickup scheduled: Jan 28

Step 5: Receive Return
  Item received at warehouse
  Quality check confirms defect
  Inspection report completed

Step 6: Create Delivery Return
  DR-2025-005
  Against: DN-2025-050
  Item: Product X, Qty: 1
  Reason: Defective
  
  Inventory Impact:
    Received into: Defective Stock
    Not added to saleable inventory

Step 7: Issue Credit Note
  CN-2025-005
  Against: INV-2025-050
  Amount: 116,000 KES (incl VAT)
  
  Accounting:
    Dr. Sales Returns             100,000
    Dr. VAT Payable                16,000
        Cr. AR - Customer                 116,000

Step 8: Replacement
  New Sales Order: SO-2025-205
  Item: Product X (replacement)
  Price: Same as original
  No additional charge
  
  Invoice: INV-2025-205
  Amount: 116,000 KES
  
  Net Effect:
    Original invoice: +116,000
    Credit note: -116,000
    New invoice: +116,000
    Customer owes: 116,000 (for replacement)

Step 9: Deliver Replacement
  DN-2025-205
  Customer receives
  Transaction complete

Step 10: Claim from Supplier
  If covered by warranty:
    Submit claim to manufacturer
    Receive replacement or refund
    Compensates for our cost
```

---