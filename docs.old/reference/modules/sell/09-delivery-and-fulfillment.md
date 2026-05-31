[<-- Back to Index](README.md)

## Delivery & Fulfillment

### Delivery Note Creation

```markdown
DELIVERY PROCESS WORKFLOW

Creation Trigger:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
From Sales Order:
  - Manual: Click "Create Delivery Note"
  - Automatic: Based on settings
  - Scheduled: Based on delivery date

Delivery Note Header:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Delivery Note No: DN-2025-001 (auto)
Date: 2025-02-10
Sales Order: SO-2025-001
Customer: ABC Manufacturing Ltd
Customer PO: PO/ABC/2025/045

Delivery Address:
  Factory Warehouse
  Thika Road, Exit 14, KM 25
  Ruiru, Kiambu County
  Contact: David Omondi
  Phone: +254-700-345-678

Shipping Details:
  Shipping Method: Company Truck
  Vehicle: KBZ 123X
  Driver: Peter Maina
  Phone: +254-722-555-888
  Expected Delivery Time: 10:00 AM

Items to Deliver:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌────────┬─────────────┬─────────┬──────────┐
│ Item   │ Description │ Ordered │ Deliver  │
├────────┼─────────────┼─────────┼──────────┤
│ MDL-A  │ Machine A   │    1    │    1 ✓   │
│ MDL-B  │ Machine B   │    1    │    1 ✓   │
│ MDL-C  │ Machine C   │    1    │    0 ⚠   │
└────────┴─────────────┴─────────┴──────────┘

⚠ Item MDL-C not yet in stock (backorder)

Delivery Type:
  □ Full Delivery (all items)
  ☑ Partial Delivery (some items)
  □ Multiple Deliveries planned

Partial Delivery Note DN-2025-001:
  Delivering: MDL-A, MDL-B
  Remaining: MDL-C (to follow)

Warehouse Operations:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Picking Process:
  1. Generate Pick List
  2. Warehouse staff locates items
  3. Scan barcodes for verification
  4. Move to staging area
  5. Quality check
  6. Pack items

Pick List:
┌────────┬──────────────┬──────┬──────────┐
│ Item   │ Location     │ Qty  │ Status   │
├────────┼──────────────┼──────┼──────────┤
│ MDL-A  │ A-12-03      │  1   │ ☑ Picked │
│ MDL-B  │ A-15-02      │  1   │ ☑ Picked │
└────────┴──────────────┴──────┴──────────┘

Packing:
  - Crate/Package items securely
  - Add packing materials
  - Label packages
  - Create packing list
  - Attach delivery documents

Quality Check:
  □ Items match order
  □ Quantities correct
  □ Items undamaged
  □ All accessories included
  □ Documentation complete
  □ Approved by: [QC Inspector]

Loading:
  - Load onto delivery vehicle
  - Secure items
  - Verify load against delivery note
  - Driver signs acknowledgment

Delivery Execution:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Dispatch:
  Time: 08:30 AM
  Vehicle: KBZ 123X
  Driver: Peter Maina
  Route: Mombasa Rd → Thika Rd → Exit 14

Delivery:
  Arrival Time: 10:15 AM
  Delivered To: Factory Warehouse Gate
  Received By: David Omondi (Technical Manager)
  Condition: Good ✓

Customer Verification:
  □ Items received as per delivery note
  □ Quantities correct
  □ Items undamaged
  □ Quality acceptable
  
Customer Signature:
  Name: David Omondi
  Signature: [Signed]
  Date/Time: 2025-02-10 10:30 AM
  Company Stamp: [Stamped]

Delivery Confirmation:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
On Delivery Confirmation:
  Status: TO DELIVER → DELIVERED
  
Automatic Actions:
  1. Inventory Update:
     Dr. Cost of Goods Sold      3,200,000
         Cr. Inventory - MDL-A         1,600,000
         Cr. Inventory - MDL-B         1,600,000
  
  2. Sales Order Update:
     Status: PARTIALLY DELIVERED
     Delivered: 2/3 items
     Remaining: 1 item (MDL-C)
  
  3. Trigger Invoice Creation:
     Create invoice for delivered items
     Amount: For MDL-A and MDL-B only
  
  4. Notifications:
     → Customer: Delivery confirmation email
     → Sales Person: Delivery completed
     → Finance: Invoice can be generated
     → Warehouse: Update stock levels
  
  5. Customer Portal:
     Delivery note uploaded
     Proof of delivery attached
     Invoice expected notification
```

### Delivery Scheduling

```markdown
DELIVERY MANAGEMENT

Delivery Planning:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Delivery Calendar View:
┌─────────────────────────────────────────────┐
│ Week: Feb 10-14, 2025                       │
├──────┬──────────────────────────────────────┤
│ Mon  │ SO-2025-001: ABC Manufacturing       │
│      │ Items: 2, Value: 3.5M, Ruiru         │
│      │ Vehicle: KBZ 123X                    │
│      │ ─────────────────────────────────    │
│      │ SO-2025-015: XYZ Ltd                 │
│      │ Items: 5, Value: 1.2M, Westlands    │
│      │ Vehicle: KBZ 456Y                    │
├──────┼──────────────────────────────────────┤
│ Tue  │ SO-2025-003: DEF Corp                │
│      │ Items: 10, Value: 2.8M, Industrial  │
│      │ Area                                 │
├──────┼──────────────────────────────────────┤
│ Wed  │ SO-2025-008: GHI Enterprises        │
│      │ Items: 3, Value: 4.5M, Mombasa     │
│      │ (Requires truck + trailer)          │
└──────┴──────────────────────────────────────┘

Route Optimization:
  Group deliveries by:
    - Geographic area
    - Delivery time windows
    - Vehicle capacity
    - Priority level

Delivery Constraints:
  - Customer receiving hours
  - Traffic patterns
  - Vehicle availability
  - Driver schedules
  - Special handling requirements

Multiple Deliveries:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Sales Order: SO-2025-001
Total Items: 3 (MDL-A, MDL-B, MDL-C)

Delivery 1: DN-2025-001 (Feb 10)
  Items: MDL-A, MDL-B (in stock)
  Status: DELIVERED ✓

Delivery 2: DN-2025-002 (Feb 20)
  Items: MDL-C (arrived Feb 19)
  Status: SCHEDULED

Sales Order Status:
  PARTIALLY DELIVERED → DELIVERED (after final delivery)

Delivery Issues:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Failed Delivery:
  Reasons:
    - Customer not available
    - Gate closed
    - Wrong address
    - Delivery rejected

  Actions:
    - Log failure reason
    - Contact customer
    - Reschedule delivery
    - Return items to warehouse
    - Update order status

Damaged in Transit:
  1. Document damage (photos)
  2. Get customer statement
  3. File insurance claim (if applicable)
  4. Arrange replacement
  5. Credit note for damaged items

Partial Rejection:
  Customer accepts some items, rejects others
  1. Note accepted items (delivered)
  2. Return rejected items
  3. Create sales return for rejected items
  4. Update delivery note
  5. Adjust invoice accordingly
```

### Delivery Documentation

```markdown
DELIVERY NOTE FORMAT

┌────────────────────────────────────────────┐
│         [COMPANY LOGO]                     │
│                                            │
│        DELIVERY NOTE                       │
│                                            │
│ DN No: DN-2025-001                         │
│ Date: February 10, 2025                    │
│ Time: 10:30 AM                             │
│                                            │
│ Sales Order: SO-2025-001                   │
│ Customer PO: PO/ABC/2025/045               │
├────────────────────────────────────────────┤
│ DELIVER TO:                                │
│                                            │
│ ABC Manufacturing Ltd                      │
│ Factory Warehouse                          │
│ Thika Road, Exit 14, KM 25                │
│ Ruiru, Kiambu County                       │
│                                            │
│ Contact Person: David Omondi               │
│ Phone: +254-700-345-678                    │
├────────────────────────────────────────────┤
│ DELIVERY DETAILS:                          │
│                                            │
│ Item   Description    Serial No    Qty    │
│ ─────  ────────────   ─────────    ───    │
│ MDL-A  Machine Model A  SN12345     1     │
│ MDL-B  Machine Model B  SN12346     1     │
│                                            │
│ Total Items: 2                             │
├────────────────────────────────────────────┤
│ SPECIAL INSTRUCTIONS:                      │
│ - Handle with care                         │
│ - Deliver to warehouse bay 3               │
│ - Installation scheduled for Feb 12        │
├────────────────────────────────────────────┤
│ DELIVERED BY:                              │
│                                            │
│ Name: Peter Maina                          │
│ Signature: _______________                │
│ Vehicle: KBZ 123X                          │
│ Date/Time: _______________                │
│                                            │
│ RECEIVED BY:                               │
│                                            │
│ Name: David Omondi                         │
│ Signature: _______________                │
│ Company Stamp: □                           │
│ Date/Time: _______________                │
│                                            │
│ CONDITION ON RECEIPT:                      │
│ □ Good Condition  □ Damaged                │
│                                            │
│ Remarks: _____________________________    │
│ _______________________________________    │
└────────────────────────────────────────────┘

Additional Documents:
  □ Invoice (if applicable)
  □ Warranty card
  □ User manual
  □ Installation guide
  □ Safety certificate
```

---