[<-- Back to Index](README.md)

## 22. Troubleshooting Guide

### Common Issues & Solutions

```markdown
TROUBLESHOOTING REFERENCE

Issue 1: Cannot Create Sales Order
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Symptom:
  "Error: Unable to create sales order"

Possible Causes & Solutions:

A) Credit Limit Exceeded
   Check: Customer credit exposure
   Solution: 
     - Request credit limit increase, OR
     - Require deposit payment, OR
     - Get management override

B) Customer On Hold
   Check: Customer status
   Solution:
     - Clear overdue invoices
     - Resolve hold reason
     - Contact credit manager

C) Negative Stock
   Check: Item availability
   Solution:
     - Wait for stock
     - Allow backorder
     - Substitute item

D) Incomplete Customer Record
   Check: Required fields
   Solution:
     - Complete billing address
     - Add payment terms
     - Set price list

E) No Permission
   Check: User role
   Solution:
     - Request access from admin
     - Check role permissions

Issue 2: Invoice Not Posting to GL
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Symptom:
  Invoice submitted but no GL entry

Diagnosis Steps:
  1. Check invoice status
     Must be: SUBMITTED
     If DRAFT: Not yet posted
  
  2. Check GL accounts configured
     Navigate to: Accounts Setup
     Verify: Revenue account set
  
  3. Check fiscal year
     Must be: Open
     If closed: Cannot post
  
  4. Check posting settings
     Auto-post enabled?
     Manual post required?

Solution:
  - Ensure invoice submitted (not draft)
  - Configure missing GL accounts
  - Open fiscal period if needed
  - Manually post if auto-post disabled
  - Contact system admin if persistent

Issue 3: Payment Not Allocating
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Symptom:
  Payment recorded but invoice still showing due

Causes:

A) Wrong Customer
   Check: Payment party matches invoice customer
   Solution: Delete and recreate payment

B) Not Allocated
   Check: Payment details
   Solution: Open payment, allocate to invoice

C) Amount Mismatch
   Check: Payment currency vs invoice currency
   Solution: Check exchange rate, reprocess

D) Payment Date Before Invoice
   Check: Dates
   Solution: Payment date must be ≥ invoice date

Issue 4: Duplicate Invoices
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Symptom:
  Customer complains of duplicate billing

Diagnosis:
  Search all invoices for customer
  Check for:
    - Same items
    - Same amounts
    - Close dates

Solution:
  If genuine duplicate:
    1. Cancel duplicate invoice
    2. Reverse GL entries
    3. Notify customer
    4. Update records
    5. Investigate how it happened
  
  If legitimate:
    - Explain to customer
    - Show different orders/deliveries
    - Provide documentation

Issue 5: Stock Not Reserving
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Symptom:
  Sales order confirmed but stock not reserved

Check:

A) Setting Enabled
   Sales Settings → Reserve Stock on SO
   Should be: Enabled

B) Warehouse Set
   Sales Order → Items → Warehouse
   Must have: Default warehouse

C) Stock Available
   Check: Available quantity
   May be: Already reserved

D) Item Status
   Check: Item master
   May be: Inactive or discontinued

Solution:
  - Enable stock reservation
  - Set default warehouse
  - Check stock availability
  - Activate item if needed

Issue 6: Commission Not Calculating
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Symptom:
  Sales made but no commission showing

Check:

A) Sales Person Assigned
   Order/Invoice must have sales person
   
B) Commission Rule Active
   Check commission structure
   Must be active and valid

C) Calculation Basis
   On Invoice vs On Payment
   Payment received if "on payment"

D) Period Closed
   Commission calculated monthly
   May need to wait for period end

Solution:
  - Assign sales person to transactions
  - Activate commission rule
  - Ensure payment received (if required)
  - Run commission calculation job

Issue 7: Exchange Rate Not Applying
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Symptom:
  Foreign currency transaction using wrong rate

Check:

A) Rate Not Defined
   Currency Exchange → Date
   No rate for transaction date

B) Manual Override
   Transaction may have manual rate
   Check if system rate overridden

C) Rate Type
   Spot vs Forward
   Check rate type setting

Solution:
  - Add exchange rate for date
  - Remove manual override if error
  - Update to correct rate type
  - Recalculate transaction
```

### Performance Issues

```markdown
SYSTEM PERFORMANCE TROUBLESHOOTING

Slow Quote/Order Creation:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Symptom:
  Taking too long to load/save

Possible Causes:

A) Large Item Count
   Loading 1000s of items
   Solution: Use search/filter first

B) Complex Pricing Rules
   Multiple overlapping rules
   Solution: Simplify rule structure

C) Network Latency
   Slow connection to server
   Solution: Check network, clear cache

D) Database Performance
   Slow queries
   Solution: Contact IT for optimization

Best Practices:
  - Use search vs browse all items
  - Limit pricing rules
  - Regular cache clearing
  - Database maintenance

Report Generation Slow:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Symptom:
  Reports timing out or very slow

Solutions:
  1. Reduce date range
     Instead of: All time
     Use: This year or quarter
  
  2. Add filters
     Filter by: Territory, Customer, Product
  
  3. Schedule Reports
     Run overnight for large reports
     Email when complete
  
  4. Use Summary Reports
     Instead of: Transaction detail
     Use: Aggregated summary

Data Not Refreshing:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Symptom:
  Changes not reflecting immediately

Cause:
  Browser/system cache

Solution:
  1. Hard refresh (Ctrl+Shift+R)
  2. Clear browser cache
  3. Log out and back in
  4. Check if change actually saved
```

---