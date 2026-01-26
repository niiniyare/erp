[<-- Back to Index](README.md)

## 23. Business Rules & Validation

### Sales Transaction Rules

```markdown
SALES BUSINESS RULES

Order Creation Rules:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Customer Required
   Rule: Must select customer
   Error: "Customer is mandatory"

2. Items Required
   Rule: At least 1 item
   Error: "Add at least one item"

3. Pricing Rules
   Rule: Rate must be > 0
   Rule: Rate must be ≥ cost (margin check)
   Warning: "Selling below cost"

4. Credit Limit
   Rule: Order + Outstanding ≤ Credit Limit
   Action: Block or require approval

5. Customer Status
   Rule: Customer must be Active
   Rule: Customer not on credit hold
   Error: "Customer on hold"

6. Payment Terms
   Rule: Payment terms must be set
   Default: Net 30 Days (if not set)

7. Delivery Date
   Rule: Must be ≥ Order Date
   Rule: Must consider lead time
   Warning: "Delivery date too soon"

8. Stock Availability
   Rule: If "Check Stock" enabled
   Action: Warn if insufficient
   Option: Allow backorder

Pricing Validation:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Price List Required
   Rule: Customer must have price list
   Default: Standard price list

2. Price Override
   Rule: Need permission to edit rate
   Check: User has "Override Price" role

3. Discount Limits
   Rule: Max discount by user level
   Level 1: 10%
   Level 2: 15% (with approval)
   Level 3: 20% (senior approval)

4. Negative Pricing
   Rule: Cannot have negative amount
   Exception: Credit notes (returns)

5. Zero Pricing
   Rule: Warning if item rate = 0
   Allow: For free samples/warranty

6. Currency Consistency
   Rule: All items same currency as order
   Error: "Mixed currencies not allowed"

Invoice Rules:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Delivery Required
   Rule: Must have delivery note (configurable)
   Exception: Service invoices

2. Invoice Date
   Rule: ≥ Delivery date
   Rule: Within open fiscal period
   Error: "Cannot invoice in closed period"

3. Tax Calculation
   Rule: Must apply correct VAT rate
   Validation: VAT = Taxable × Rate

4. Customer PO
   Rule: Required for corporate customers
   Warning: "PO number missing"

5. Terms Matching
   Rule: Invoice terms match order terms
   Allow: Override with approval

6. Amount Limits
   Rule: Invoice ≤ Sales Order amount
   Exception: Additional charges

Payment Rules:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Payment Date
   Rule: ≥ Invoice date
   Error: "Payment date before invoice"

2. Amount Limits
   Rule: Payment ≤ Outstanding balance
   Allow: Overpayment (creates credit)

3. Payment Allocation
   Rule: Must allocate to invoice(s)
   OR: Unallocated payment allowed

4. Payment Method
   Rule: Must select payment method
   Validation: Method-specific fields

5. Reference Number
   Rule: Required for bank transfers
   Validation: Unique per payment

6. Currency Matching
   Rule: Payment currency = Invoice currency
   OR: Exchange rate required

Return Rules:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Return Period
   Rule: Within X days of delivery
   Default: 30 days
   Override: Manager approval

2. Original Invoice
   Rule: Must reference original invoice
   Validation: Invoice must exist

3. Quantity Limits
   Rule: Return qty ≤ Invoiced qty
   Error: "Cannot return more than purchased"

4. Item Condition
   Rule: Must pass inspection
   Status: Good/Damaged/Defective

5. Return Reason
   Rule: Reason required
   List: Defective, Wrong item, Changed mind

6. Credit Note
   Rule: Must issue credit note
   Approval: Required if > threshold

7. Restocking Fee
   Rule: Apply if applicable
   Rate: Per policy (0-15%)
```

### Data Validation Rules

```markdown
DATA QUALITY RULES

Customer Master:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Required Fields:
  ☑ Customer Name
  ☑ Customer Group
  ☑ Territory
  ☑ Currency
  ☑ Payment Terms

Format Validation:
  Email: Must be valid email format
  Phone: Must be valid phone number
  PIN: Must match country format

Uniqueness:
  Customer Name: Warning if duplicate
  Email: Allow duplicate (multiple contacts)
  Tax ID: Unique per customer

Business Logic:
  Credit Limit: Must be ≥ 0
  Payment Terms: Must exist in master
  Price List: Must be active

Product/Item:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Required:
  ☑ Item Code
  ☑ Item Name
  ☑ UOM (Unit of Measure)
  ☑ Item Group

Validation:
  Item Code: Unique
  Standard Rate: Must be > 0
  Valuation Rate (Cost): Must be > 0

Pricing:
  Selling Rate ≥ Cost (recommended)
  Warning if selling below cost

Sales Transaction:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Document Number:
  Auto-generated
  Sequential
  No gaps (by fiscal year)

Dates:
  Order Date ≤ Delivery Date
  Invoice Date ≥ Delivery Date
  Payment Date ≥ Invoice Date

Amounts:
  Quantity: Must be > 0
  Rate: Must be > 0
  Discount: 0 ≤ Discount ≤ 100%
  Total: Auto-calculated, cannot edit

Status Flow:
  Draft → Submitted → Paid/Overdue
  Cannot skip states
  Cannot reverse (except cancel)

Referential Integrity:
  Customer must exist
  Items must exist
  Warehouse must exist
  GL accounts must exist
```

### Security & Access Rules

```markdown
ROLE-BASED ACCESS CONTROL

Sales Person Role:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Permissions:
  ✓ Create quotations
  ✓ Create sales orders
  ✓ View own transactions
  ✓ View assigned customers
  ✗ Edit submitted documents
  ✗ Cancel orders (need approval)
  ✗ Override credit limit
  ✗ Change pricing (within limits)

Data Access:
  Own Territory: Full access
  Other Territories: Read-only
  Own Customers: Full access
  All Reports: Filtered to own data

Sales Manager Role:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Additional Permissions:
  ✓ Approve discounts (up to 15%)
  ✓ Override prices
  ✓ Cancel orders
  ✓ Modify submitted quotes
  ✓ View team performance
  ✓ Access all territory data

Data Access:
  Region: Full access
  All territories in region
  Team reports and dashboards

Credit Manager Role:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Permissions:
  ✓ Set credit limits
  ✓ Put customers on hold
  ✓ Approve credit increases
  ✓ Write off bad debts
  ✓ View all AR data
  ✓ Collection reports

Data Restrictions:
  Focus: Financial data
  Limited: Sales operations

Finance Team Role:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Permissions:
  ✓ Process payments
  ✓ Issue credit notes
  ✓ View all invoices
  ✓ Run financial reports
  ✗ Create sales orders
  ✗ Modify pricing

System Administrator:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Full Access:
  ✓ All modules
  ✓ Configuration
  ✓ User management
  ✓ System settings
  ✓ Data import/export

Territory-Based Security:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Rule: Users see only their territory data

Example:
  User: John (Nairobi)
  Territory: Nairobi Corporate
  
  Can View:
    ✓ Nairobi customers
    ✓ Nairobi orders
    ✓ Nairobi reports
  
  Cannot View:
    ✗ Mombasa customers
    ✗ Mombasa orders
    ✗ Other territories

Exception: Managers see all territories
          in their region

Audit Trail:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
All Actions Logged:
  - Who: User name
  - What: Action taken
  - When: Date/time
  - Where: Document/record
  - Changes: Before/after values

Cannot Delete:
  - Submitted documents
  - Posted transactions
  - Completed deliveries

Can Only:
  - Cancel (with approval)
  - Amend (creates new version)
  - Credit note (reverse)

Retention:
  Transaction data: 7 years
  Audit logs: 3 years
  Archived data: Permanent
```

---

## Summary

This completes the comprehensive Selling Module documentation covering:

✅ Executive Summary & Module Purpose
✅ Why ERP Sales Management
✅ Complete Sales Cycle Process
✅ Initial Setup & Configuration
✅ Master Data Management
✅ Lead & Opportunity Management
✅ Quotation Management
✅ Sales Order Processing
✅ Delivery & Fulfillment
✅ Sales Invoicing
✅ Payment Collection & Allocation
✅ Returns & Credit Management
✅ Pricing & Discount Management
✅ Sales Analytics & Reporting
✅ Sales Teams & Territories
✅ Commission Management
✅ Customer Credit Management
✅ Multi-Currency Sales
✅ Module Integration Points
✅ Common Business Scenarios
✅ Approval Workflows
✅ Troubleshooting Guide
✅ Business Rules & Validation

**Coverage**: End-to-end sales operations
**Audience**: Business users, implementation teams, system administrators
**Use Cases**: Training, reference, troubleshooting, process design

The documentation provides practical, actionable guidance for implementing and using an enterprise-grade sales management system.