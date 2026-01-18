# Selling Module - Complete Business Domain Guide

> **Comprehensive guide covering sales processes, customer management, order-to-cash cycle, and integration with the AWO ERP ecosystem.**

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Why Sales Management in ERP](#why-sales-management-in-erp)
3. [Sales Process Overview](#sales-process-overview)
4. [Initial Setup & Configuration](#initial-setup--configuration)
5. [Master Data Management](#master-data-management)
6. [Lead & Opportunity Management](#lead--opportunity-management)
7. [Quotation Management](#quotation-management)
8. [Sales Order Processing](#sales-order-processing)
9. [Delivery & Fulfillment](#delivery--fulfillment)
10. [Sales Invoicing](#sales-invoicing)
11. [Payment Collection & Allocation](#payment-collection--allocation)
12. [Returns & Credit Management](#returns--credit-management)
13. [Pricing & Discount Management](#pricing--discount-management)
14. [Sales Analytics & Reporting](#sales-analytics--reporting)
15. [Sales Teams & Territories](#sales-teams--territories)
16. [Commission Management](#commission-management)
17. [Customer Credit Management](#customer-credit-management)
18. [Multi-Currency Sales](#multi-currency-sales)
19. [Module Integration Points](#module-integration-points)
20. [Common Business Scenarios](#common-business-scenarios)
21. [Approval Workflows](#approval-workflows)
22. [Troubleshooting Guide](#troubleshooting-guide)
23. [Business Rules & Validation](#business-rules--validation)

---

## Executive Summary

### Module Purpose

The Selling Module orchestrates the complete order-to-cash cycle within the AWO ERP ecosystem, providing:

- **End-to-End Sales Management**: From lead capture through payment collection
- **Customer Lifecycle Tracking**: Complete interaction history and relationship management
- **Integrated Order Fulfillment**: Seamless coordination with Inventory and Finance modules
- **Automated Revenue Recognition**: Real-time financial integration
- **Sales Performance Intelligence**: Comprehensive analytics and forecasting
- **Multi-Tenant Support**: Isolated customer data per tenant with shared configurations

### System Architecture Context

The Selling Module operates within the AWO ERP multi-tenant architecture:

```markdown
CORE SYSTEM CAPABILITIES LEVERAGED:

Tenant Management:
✓ Complete data isolation per tenant
✓ Tenant-specific pricing and configurations
✓ Cross-tenant reporting (where authorized)
✓ Tenant-level feature enablement

ABAC/RBAC Security:
✓ Role-based sales permissions
✓ Attribute-based pricing access
✓ Territory-based data visibility
✓ Customer-level access controls

Workflow Engine:
✓ Quotation approval workflows
✓ Discount authorization flows
✓ Credit limit override approvals
✓ Sales order confirmation routing

Configuration Management:
✓ Tenant-specific sales settings
✓ Document numbering sequences
✓ Email templates and notifications
✓ Custom field definitions

Feature Flags:
✓ Progressive feature rollout
✓ A/B testing of sales features
✓ Tenant-specific feature access
✓ Beta feature opt-in

Entity Structure:
✓ Company hierarchy support
✓ Multi-company sales consolidation
✓ Department-level sales tracking
✓ Cost center allocation

Report Builder:
✓ Custom sales reports
✓ Dashboard creation
✓ KPI visualization
✓ Export capabilities
```

### Integration with Finance Module

The Selling Module integrates deeply with the Financial Module:

```markdown
FINANCIAL INTEGRATION POINTS:

Revenue Recognition:
- Sales Invoice → GL Revenue Entry
- Automatic account posting
- Multi-currency conversion
- Tax calculation and posting

Accounts Receivable:
- Customer balance tracking
- Payment allocation
- Aging analysis
- Credit management

Cash Management:
- Payment receipt processing
- Bank reconciliation support
- Cash flow forecasting
- Collection tracking

Cost Accounting:
- COGS calculation (from Inventory)
- Margin analysis
- Profitability reporting
- Cost center allocation
```

### Key Stakeholders

**Sales Representatives:**
- Lead and opportunity management
- Quotation creation and tracking
- Order entry and monitoring
- Customer communication

**Sales Managers:**
- Team performance oversight
- Pipeline management
- Discount approvals
- Territory assignments

**Customer Service:**
- Order status tracking
- Returns processing
- Customer inquiries
- Issue resolution

**Operations/Fulfillment:**
- Order picking and packing
- Delivery scheduling
- Stock allocation
- Shipment tracking

**Finance Team:**
- Invoice generation and approval
- Payment collection
- Credit control
- Revenue analysis

**Executive Leadership:**
- Sales forecasting
- Performance dashboards
- Strategic planning
- Market analysis

### Success Metrics

Organizations implementing integrated sales management achieve:

- **40% faster** quote-to-order conversion
- **50% reduction** in order errors
- **30% improvement** in sales productivity
- **25% faster** order fulfillment
- **35% reduction** in DSO (Days Sales Outstanding)
- **Real-time** inventory visibility
- **95% accuracy** in commission calculations
- **60% reduction** in manual data entry

---

## Why Sales Management in ERP

### The Problem with Fragmented Systems

**Typical Disconnected Sales Process:**

```markdown
SALES REP (CRM/Spreadsheet):
├─ Creates quote manually in Word/Excel
├─ No real-time inventory visibility
├─ Manual pricing calculations
├─ Email quote to customer
└─ Track follow-ups in personal spreadsheet

CUSTOMER ACCEPTS:
├─ Email order details to operations
├─ Operations re-enters data into their system
├─ Inventory check (phone calls/emails)
├─ Confirm availability manually
└─ Create picking list manually

WAREHOUSE:
├─ Receives printed picking list
├─ Manual stock picking
├─ Creates delivery note (handwritten)
├─ Fax/email to accounting
└─ No real-time inventory update

ACCOUNTING (Separate System):
├─ Manually creates invoice from delivery note
├─ Re-enters all customer details
├─ Re-enters all product details
├─ Calculates tax manually
├─ Prints and mails invoice
└─ Manually tracks payment in spreadsheet

MONTH-END:
├─ Reconcile sales across multiple systems
├─ Calculate commissions manually
├─ Fix errors and disputes
├─ Generate reports from spreadsheets
└─ 5-7 days to close sales books

PROBLEMS:
❌ 5+ disconnected systems
❌ Data entered 3-4 times
❌ High error rate (15-20%)
❌ No real-time visibility
❌ 3-5 day quote-to-invoice cycle
❌ Lost quotes and orders
❌ Inventory overselling
❌ Commission disputes
❌ Manual reporting (days old)
❌ Customer service issues
```

**Integrated ERP Sales Process:**

```markdown
SALES REP (AWO ERP):
├─ Create quote in system
│  ├─ Customer auto-populated from master
│  ├─ Products with real-time stock levels
│  ├─ Pricing auto-applied (rules engine)
│  ├─ Discounts calculated automatically
│  └─ Professional PDF generated
├─ System checks inventory availability
├─ Send quote via email (tracked)
└─ Automatic follow-up reminders

CUSTOMER ACCEPTS:
├─ Convert quote to sales order (one click)
├─ Inventory automatically reserved
├─ Workflow notification to operations
├─ Customer receives order confirmation email
└─ All data already in system

WAREHOUSE (Integrated):
├─ Receives order in system
├─ Pick list auto-generated
├─ Scan items for accuracy
├─ Create delivery note (auto-generated)
├─ System updates inventory real-time
└─ Customer signature captured digitally

FINANCE (Automatic):
├─ Invoice auto-created from delivery
├─ All data pre-filled (zero re-entry)
├─ Tax calculated automatically
├─ Posted to GL automatically:
│  Dr. Accounts Receivable
│  Cr. Sales Revenue
│  Cr. Tax Payable
├─ Email invoice to customer
└─ Payment tracking begins

MONTH-END:
├─ All data already reconciled
├─ Commission calculated automatically
├─ Reports available real-time
├─ Close sales books: 1 day
└─ Financial statements ready

BENEFITS:
✅ Single integrated system
✅ Data entered once
✅ < 1% error rate
✅ Real-time visibility
✅ Same-day quote-to-invoice
✅ Complete audit trail
✅ No inventory overselling
✅ Accurate commissions
✅ Real-time reporting
✅ Excellent customer service
```

### Multi-Tenant Architecture Benefits

**For SaaS/Multi-Company Deployments:**

```markdown
TENANT ISOLATION:

Tenant A (Kenya Operations):
├─ Customers: Kenya-based only
├─ Pricing: KES, VAT 16%
├─ Products: Kenya SKU catalog
├─ Sales team: Nairobi office
└─ Completely isolated from Tenant B

Tenant B (Uganda Operations):
├─ Customers: Uganda-based
├─ Pricing: UGX, VAT 18%
├─ Products: Uganda SKU catalog
├─ Sales team: Kampala office
└─ Completely isolated from Tenant A

SHARED CONFIGURATIONS:
├─ Product catalog template
├─ Workflow definitions
├─ Report templates
├─ Email templates
└─ Best practice processes

CROSS-TENANT REPORTING (for holding company):
├─ Consolidated sales by region
├─ Group-level analytics
├─ Performance benchmarking
├─ Controlled via RBAC
└─ Aggregated dashboards
```

---

## Sales Process Overview

### The Complete Sales Cycle

```markdown
STAGE 1: LEAD GENERATION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Source: Marketing, website, referrals, events
Action: Capture lead in system
Data: Name, company, contact, interest, source
Status: NEW → CONTACTED → QUALIFIED → CONVERTED

Example:
Lead: John Kamau from ABC Manufacturing
Source: Website inquiry form
Interest: Industrial equipment
Score: High (budget confirmed, timeline immediate)
Assigned to: Sarah (Territory: Nairobi)

STAGE 2: OPPORTUNITY CREATION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Trigger: Qualified lead with purchase intent
Action: Create opportunity with deal details
Data: Value, probability, expected close date
Status: PROSPECTING → QUALIFICATION → PROPOSAL

Example:
Opportunity: ABC Manufacturing - Equipment Purchase
Value: 5,000,000 KES
Probability: 60%
Expected Close: 2025-02-15
Stage: Proposal

STAGE 3: QUOTATION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Trigger: Customer requests formal pricing
Action: Generate professional quote
Content: Items, quantities, prices, terms
Valid: 30 days (configurable)
Status: DRAFT → SENT → ACCEPTED/REJECTED/EXPIRED

Example:
Quotation: QTN-2025-001
Customer: ABC Manufacturing
Items: 3 different equipment models
Total: 5,000,000 KES (before VAT)
Payment Terms: 50% upfront, 50% on delivery
Validity: Valid until 2025-02-12

STAGE 4: SALES ORDER
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Trigger: Customer accepts quotation (PO received)
Action: Convert to confirmed order
Effect: 
  - Inventory reserved
  - Production/procurement triggered (if needed)
  - Delivery scheduled
Status: DRAFT → CONFIRMED → TO DELIVER → COMPLETED

Example:
Sales Order: SO-2025-001 (from QTN-2025-001)
Customer PO: PO/ABC/2025/045
Inventory Status: 2 items in stock, 1 on backorder
Expected Delivery: 2025-02-20
Deposit Received: 2,500,000 KES

STAGE 5: DELIVERY/FULFILLMENT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Action: Pick, pack, ship items
Document: Delivery Note with packing list
Signature: Customer representative signs
Integration: Inventory updated, COGS calculated
Status: TO DELIVER → PARTIALLY DELIVERED → DELIVERED

Example:
Delivery Note: DN-2025-001
Items: All 3 equipment units
Delivered to: ABC Manufacturing, Industrial Area
Received by: John Kamau (signed)
Delivery Date: 2025-02-20 14:30
Inventory Impact: Stock reduced, COGS: 3,200,000 KES

STAGE 6: INVOICING
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Trigger: Goods delivered (or service completed)
Action: Generate sales invoice
Integration: 
  Dr. Accounts Receivable    5,800,000
      Cr. Sales Revenue              5,000,000
      Cr. VAT Payable                  800,000
Status: DRAFT → SUBMITTED → PAID/OVERDUE

Example:
Sales Invoice: INV-2025-001
Date: 2025-02-20
Amount: 5,800,000 KES (including VAT)
Due Date: 2025-03-02 (Net 10 days)
Payment Status: Partially Paid (deposit applied)
Balance Due: 3,300,000 KES

STAGE 7: PAYMENT COLLECTION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Action: Receive and allocate payment
Methods: Bank transfer, cash, check, mobile money
Entry:
  Dr. Bank/Cash             3,300,000
      Cr. Accounts Receivable       3,300,000
Status: UNPAID → PARTIALLY PAID → PAID

Example:
Payment Entry: PAY-2025-001
Date: 2025-03-01
Amount: 3,300,000 KES (final payment)
Method: Bank transfer
Reference: TRX/2025/12345
Allocated to: INV-2025-001
Customer Balance: 0 (fully paid)

STAGE 8: POST-SALE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Activities: 
  - Customer satisfaction survey
  - Equipment installation support
  - Training provided
  - Warranty registration
  - Upsell opportunities identified
Analytics:
  - Customer lifetime value calculated
  - Margin analysis: 36% gross margin
  - Commission calculated: 250,000 KES
Status: CLOSED-WON → ACTIVE CUSTOMER
```

### Alternative Sales Flows

**1. Retail/Cash Sale (Simplified):**
```
Walk-in Customer → Sales Invoice (immediate) → Payment (immediate)
Timeline: Same day
Documents: Sales invoice only (acts as order, delivery, invoice)
```

**2. Service Sale:**
```
Lead → Opportunity → Quote → Service Order → Service Delivery → Invoice → Payment
Timeline: Varies
Special: May have milestone-based billing
```

**3. Subscription/Recurring:**
```
Initial Sale → Recurring Invoice (automatic monthly/annual) → Auto-payment
Timeline: Ongoing
Special: Automated renewal and billing
```

**4. Project-Based:**
```
RFP → Proposal → Contract → Multiple Deliveries → Progress Billing → Final Payment
Timeline: Months to years
Special: Complex milestone tracking
```

**5. Drop-Ship:**
```
Sales Order → Purchase Order to Supplier → Supplier Ships Direct → Invoice Customer
Timeline: As per supplier
Special: No inventory movement in your warehouse
```

---

## Initial Setup & Configuration

### Prerequisites Checklist

```markdown
BEFORE CONFIGURING SALES MODULE:

□ Finance Module Setup:
  ✓ Chart of Accounts configured
  ✓ Revenue accounts created
  ✓ Receivables account set up
  ✓ Tax accounts defined
  ✓ Fiscal year active

□ Tenant Configuration:
  ✓ Primary tenant created
  ✓ Company information complete
  ✓ Base currency set
  ✓ Timezone configured

□ Entity Structure:
  ✓ Company hierarchy defined
  ✓ Departments/divisions created
  ✓ Cost centers set up
  ✓ Locations/warehouses identified

□ User & Security:
  ✓ RBAC roles defined
  ✓ Sales team users created
  ✓ Approval hierarchies mapped
  ✓ Territory access rules

□ Inventory Module (if available):
  ✓ Product catalog ready
  ✓ Stock locations defined
  ✓ Pricing structure prepared

□ Business Rules Documentation:
  ✓ Credit policies documented
  ✓ Discount approval matrix
  ✓ Commission structure defined
  ✓ Payment terms standard list
```

### Setup Sequence

**Recommended Implementation Order:**

```markdown
PHASE 1: MASTER DATA (Week 1)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Sales Settings Configuration
2. Document Numbering Series
3. Terms & Conditions Templates
4. Email Templates
5. Print Formats

PHASE 2: CUSTOMER DATA (Week 1-2)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Customer Groups
2. Customer Categories
3. Price Lists
4. Payment Terms
5. Customer Master Data Import

PHASE 3: SALES TEAM (Week 2)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Sales Territories
2. Sales Team Setup
3. Sales Person Records
4. Target/Quota Assignment
5. Commission Structure

PHASE 4: PRICING (Week 2-3)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Pricing Rules
2. Discount Rules
3. Promotional Schemes
4. Volume Discounts
5. Customer-Specific Pricing

PHASE 5: WORKFLOWS (Week 3)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Quotation Approval Workflow
2. Discount Authorization
3. Credit Limit Override
4. Sales Order Confirmation
5. Return Authorization

PHASE 6: INTEGRATION (Week 3-4)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Finance Module GL Accounts
2. Inventory Integration Points
3. Email Server Configuration
4. Payment Gateway (if applicable)
5. External System APIs

PHASE 7: TESTING (Week 4)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. End-to-end process testing
2. Workflow validation
3. Integration testing
4. User acceptance testing
5. Performance testing

PHASE 8: GO-LIVE (Week 5)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Final data migration
2. User training
3. Parallel run (optional)
4. Cutover
5. Post-go-live support
```

### Step 1: Sales Settings Configuration

```markdown
SALES MODULE SETTINGS

General Settings:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
□ Company: Select primary company
□ Default Customer Group: Retail/Wholesale/Corporate
□ Default Price List: Standard Selling Price
□ Default Territory: Head Office/Regional
□ Default Warehouse: Main Warehouse
□ Default Sales Person: (Optional)

Document Behavior:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
□ Auto-create Delivery Note from Sales Order: Yes/No
□ Auto-create Invoice from Delivery: Yes/No
□ Require Customer PO Number: Yes/No
□ Allow Multiple Sales Orders against Quotation: Yes/No
□ Allow Sales Order Creation without Quotation: Yes/No

Pricing & Discount:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
□ Apply Pricing Rule Automatically: Yes
□ Allow User to Edit Rate: Yes (with permissions)
□ Allow Discount on Item Level: Yes
□ Allow Discount on Invoice Level: Yes
□ Maximum Discount % (without approval): 10%
□ Validate Price List on Transactions: Yes

Inventory Integration:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
□ Check Stock Availability on Sales Order: Yes
□ Reserve Stock on Sales Order: Yes
□ Allow Backorders: Yes/No
□ Show Stock Balance in Quotation: Yes
□ Auto Reserve Stock on Quotation: No

Credit Control:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
□ Enable Credit Limit: Yes
□ Check Credit Limit at: Sales Order / Delivery / Invoice
□ Credit Limit Action: Warn / Stop / Ignore
□ Allow Sales Order above Credit Limit with Approval: Yes

Financial Integration:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
□ Default Revenue Account: 4000 - Sales Revenue
□ Default Receivable Account: 1200 - Accounts Receivable
□ Default Tax Account: 2300 - VAT Payable
□ Default Cost Center: Sales Department
□ Post Accounting Entry on: Invoice Submit

Email & Notifications:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
□ Send Email on Quotation Submit: Yes
□ Send Email on Sales Order Confirmation: Yes
□ Send Email on Invoice Submit: Yes
□ Send Email on Payment Receipt: Yes
□ Default Email Template: Standard Sales Template

Commission:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
□ Enable Sales Commission: Yes
□ Calculate Commission on: Sales Order / Invoice / Payment
□ Commission Type: Fixed % / Tiered / Product-based
□ Default Commission Rate: 5%
```

### Step 2: Document Numbering Series

```markdown
SALES DOCUMENT NUMBERING

Quotation:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Series: QTN-{YYYY}-{#####}
Example: QTN-2025-00001
Prefix Options:
  - By Fiscal Year: QTN-FY25-
  - By Company: QTN-KE- / QTN-UG-
  - By Territory: QTN-NAI- / QTN-MBA-

Sales Order:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Series: SO-{YYYY}-{#####}
Example: SO-2025-00001
Auto-increment: Yes
Starting Number: 1

Delivery Note:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Series: DN-{YYYY}-{#####}
Example: DN-2025-00001
Alternative: DEL-{YYYY}-{#####}

Sales Invoice:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Series: INV-{YYYY}-{#####}
Example: INV-2025-00001
Tax Invoice: TAX-INV-{YYYY}-{#####}
Proforma: PRO-{YYYY}-{#####}

Sales Return:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Series: RET-{YYYY}-{#####}
Example: RET-2025-00001
Credit Note: CN-{YYYY}-{#####}

Payment Entry:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Series: PAY-{YYYY}-{#####}
Example: PAY-2025-00001
Receipt: RCP-{YYYY}-{#####}

CONFIGURATION:

For Each Document Type:
┌─────────────────────────────────────────┐
│ Document: Sales Order                   │
│ Series Format: SO-{YYYY}-{#####}       │
│ Current Number: 1                       │
│ Padding: 5 digits                       │
│ Reset Period: Yearly / Never            │
│ Active: Yes                             │
│ Set as Default: Yes                     │
└─────────────────────────────────────────┘

Multiple Series Example:
- Regular Sales: SO-2025-xxxxx
- Export Sales: EXP-2025-xxxxx
- Internal Sales: INT-2025-xxxxx
User selects series when creating document
```

### Step 3: Terms & Conditions Templates

```markdown
STANDARD TERMS & CONDITIONS

1. Payment Terms Template
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PAYMENT TERMS:
Payment is due within [X] days from invoice date.
Late payments will incur interest at [Y]% per month.
Accepted payment methods: Bank transfer, Cash, Check.

Bank Details:
Bank: [Bank Name]
Account: [Account Number]
Branch: [Branch Name]
SWIFT: [SWIFT Code]

2. Delivery Terms Template
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
DELIVERY TERMS:
Delivery within [X] business days from order confirmation.
Delivery to [Address] during business hours (8 AM - 5 PM).
Customer responsible for offloading (unless agreed otherwise).
Risk transfers to customer upon delivery and signature.

3. Warranty Terms Template
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
WARRANTY:
Products covered by [X]-month manufacturer warranty.
Warranty covers manufacturing defects only.
Damage from misuse, accidents, or unauthorized repairs void warranty.
Customer must report defects within [X] days of discovery.

4. Return Policy Template
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
RETURNS:
Returns accepted within [X] days of delivery.
Items must be unused, in original packaging.
Return shipping costs borne by [Customer/Seller].
Restocking fee of [X]% may apply.
Custom orders non-returnable.

5. Quotation Validity Template
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
VALIDITY:
This quotation is valid for [X] days from date of issue.
Prices subject to change after validity period.
Availability of items subject to prior sale.
Quotation does not constitute a binding offer until order confirmation.

CONFIGURATION:

Terms & Conditions Master:
┌─────────────────────────────────────────┐
│ Title: Standard Payment Terms           │
│ Category: Payment                       │
│ Content: [Full text as above]           │
│ Apply to: Quotation, Sales Order,       │
│           Sales Invoice                 │
│ Default: Yes/No                         │
│ Require Acceptance: Yes/No              │
└─────────────────────────────────────────┘

Multiple templates can be maintained and selected per document.
```

### Step 4: Email Templates

```markdown
EMAIL TEMPLATES CONFIGURATION

1. Quotation Email Template
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Subject: Quotation {quotation_number} from {company_name}

Dear {customer_name},

Thank you for your interest in our products/services.

Please find attached quotation {quotation_number} as requested.

Quotation Summary:
- Total Amount: {currency} {total_amount}
- Valid Until: {valid_till_date}
- Payment Terms: {payment_terms}

Should you have any questions or require clarification, please don't 
hesitate to contact me.

We look forward to serving you.

Best regards,
{sales_person_name}
{sales_person_email}
{company_name}

Attachment: Quotation_{quotation_number}.pdf

2. Sales Order Confirmation Email
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Subject: Order Confirmation - {sales_order_number}

Dear {customer_name},

Thank you for your order!

We are pleased to confirm your order {sales_order_number}.

Order Details:
- Order Number: {sales_order_number}
- Order Date: {order_date}
- Total Amount: {currency} {total_amount}
- Expected Delivery: {delivery_date}

Your order is being processed and you will receive a delivery 
notification once shipped.

Track your order: [tracking_link]

Thank you for your business!

Best regards,
{company_name}

3. Delivery Notification Email
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Subject: Your order {sales_order_number} has been delivered

Dear {customer_name},

Your order {sales_order_number} was successfully delivered on 
{delivery_date}.

Delivery Note: {delivery_note_number}
Delivered To: {delivery_address}
Received By: {receiver_name}

Please confirm receipt and inspect items for any damage.
Report any issues within 24 hours.

Thank you for choosing {company_name}!

Best regards,
{company_name}

4. Invoice Email Template
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Subject: Invoice {invoice_number} - {company_name}

Dear {customer_name},

Please find attached invoice {invoice_number} for recent delivery.

Invoice Summary:
- Invoice Number: {invoice_number}
- Invoice Date: {invoice_date}
- Due Date: {due_date}
- Amount Due: {currency} {outstanding_amount}

Payment Instructions:
{payment_instructions}

Pay online: [payment_link]

For any queries regarding this invoice, please contact our 
accounts team at {accounts_email}.

Thank you for your business!

Best regards,
{company_name}

Attachment: Invoice_{invoice_number}.pdf

5. Payment Reminder Email
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Subject: Payment Reminder - Invoice {invoice_number}

Dear {customer_name},

This is a friendly reminder that invoice {invoice_number} is due 
for payment on {due_date}.

Invoice Details:
- Invoice Number: {invoice_number}
- Invoice Date: {invoice_date}
- Due Date: {due_date}
- Amount Due: {currency} {outstanding_amount}
- Days Overdue: {days_overdue}

If you have already made payment, please disregard this reminder.
Otherwise, please arrange payment at your earliest convenience.

For payment arrangements or queries, contact: {accounts_email}

Thank you for your prompt attention.

Best regards,
{company_name}

6. Payment Received Email
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Subject: Payment Received - Thank You!

Dear {customer_name},

We confirm receipt of your payment for invoice {invoice_number}.

Payment Details:
- Amount Received: {currency} {amount_paid}
- Payment Date: {payment_date}
- Payment Method: {payment_method}
- Reference: {payment_reference}

Invoice {invoice_number} is now marked as PAID.

Outstanding Balance: {currency} {outstanding_balance}

Thank you for your payment!

Best regards,
{company_name}

CONFIGURATION:

Email Template Master:
┌─────────────────────────────────────────┐
│ Template Name: Quotation Submission     │
│ Document Type: Quotation                │
│ Subject: [Template with variables]      │
│ Body: [HTML/Plain text template]        │
│ Attachments: PDF, Additional Docs       │
│ Send To: Customer Email                 │
│ CC: Sales Person, Sales Manager         │
│ Trigger: On Submit                      │
│ Active: Yes                             │
└─────────────────────────────────────────┘

Available Variables:
{customer_name}, {customer_email}, {company_name},
{document_number}, {total_amount}, {date}, 
{sales_person_name}, {due_date}, etc.
```

### Step 5: Print Format Configuration

```markdown
PRINT FORMATS & BRANDING

Quotation Print Format:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌────────────────────────────────────────────┐
│ [Company Logo]          QUOTATION          │
│                                            │
│ Company Name                    QTN No:   │
│ Address Line 1                  Date:     │
│ Address Line 2                  Valid:    │
│ Phone: xxx  Email: xxx                    │
├────────────────────────────────────────────┤
│ BILL TO:                                  │
│ Customer Name                              │
│ Customer Address                           │
│ Contact Person: xxx                        │
│ Phone: xxx  Email: xxx                    │
├────────────────────────────────────────────┤
│ Item  Description    Qty  Rate    Amount  │
│ ────  ───────────    ───  ────    ──────  │
│ 1.    Product A      10   1,000   10,000  │
│ 2.    Product B      5    2,000   10,000  │
│                                            │
│                      Subtotal:    20,000  │
│                      Discount:    (1,000) │
│                      VAT (16%):    3,040  │
│                      TOTAL:       22,040  │
├────────────────────────────────────────────┤
│ TERMS & CONDITIONS:                        │
│ [Standard terms]                           │
├────────────────────────────────────────────┤
│ Prepared by: [Sales Person]                │
│ Signature: _______________                │
└────────────────────────────────────────────┘

Sales Invoice Format:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌────────────────────────────────────────────┐
│ [Company Logo]      TAX INVOICE            │
│                                            │
│ Company Name             Invoice No:      │
│ PIN: xxx                 Date:            │
│ Address                  Due Date:        │
│                          Sales Order:     │
├────────────────────────────────────────────┤
│ CUSTOMER DETAILS:                          │
│ Name: xxx              PIN: xxx            │
│ Address: xxx                               │
│                                            │
├────────────────────────────────────────────┤
│ Item  Description    Qty  Rate    Amount  │
│ ────  ───────────    ───  ────    ──────  │
│ [Line items]                               │
│                                            │
│                      Subtotal:    xxxxx   │
│                      VAT (16%):   xxxxx   │
│                      TOTAL:       xxxxx   │
├────────────────────────────────────────────┤
│ PAYMENT DETAILS:                           │
│ Bank: xxx  Account: xxx                    │
│ Branch: xxx  Swift: xxx                    │
├────────────────────────────────────────────┤
│ Amount in Words: [Amount in words]         │
│                                            │
│ This is a computer-generated invoice       │
└────────────────────────────────────────────┘

Delivery Note Format:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌────────────────────────────────────────────┐
│ [Company Logo]     DELIVERY NOTE           │
│                                            │
│ DN No: xxx           Date: xxx             │
│ Sales Order: xxx                           │
│                                            │
│ DELIVER TO:                                │
│ Customer: xxx                              │
│ Address: xxx                               │
│ Contact: xxx                               │
│                                            │
│ Item  Description         Qty  Delivered   │
│ ────  ──────────────      ───  ─────────   │
│ [Line items with checkboxes]               │
│                                            │
│ Special Instructions:                      │
│ [Any delivery notes]                       │
│                                            │
│ Delivered By: _______________             │
│ Signature: _______________                │
│ Date/Time: _______________                │
│                                            │
│ Received By: _______________              │
│ Signature: _______________                │
│ Date/Time: _______________                │
│ Condition: □ Good  □ Damaged              │
└────────────────────────────────────────────┘

CONFIGURATION:

Print Format Settings:
┌─────────────────────────────────────────┐
│ Document Type: Sales Invoice            │
│ Format Name: Standard Tax Invoice       │
│ Page Size: A4                           │
│ Orientation: Portrait                   │
│ Show Company Letterhead: Yes            │
│ Show Watermark: Draft/Paid              │
│ Show Terms & Conditions: Yes            │
│ Show Payment Instructions: Yes          │
│ Language: English / Swahili             │
│ Barcode/QR Code: Yes (for payment)     │
│ Default: Yes                            │
└─────────────────────────────────────────┘

Letterhead Configuration:
- Upload company logo
- Define header content
- Define footer content
- Set margins and spacing
- Configure colors and fonts
```

---

## Master Data Management

### Customer Master

**Customer Record Structure:**

```markdown
CUSTOMER MASTER DATA

Basic Information:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Customer ID: AUTO-GENERATED (CUST-00001)
Customer Name: ABC Manufacturing Ltd
Customer Type: Company / Individual
Customer Group: Corporate / Retail / Wholesale / Distributor
Territory: Nairobi / Mombasa / Kisumu / etc.
Industry: Manufacturing / Retail / Services / etc.
Customer Since: 2023-01-15
Status: Active / Inactive / Suspended

Contact Information:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Primary Contact Person: John Kamau
Title: Procurement Manager
Email: john.kamau@abcmanufacturing.com
Phone: +254-700-123-456
Mobile: +254-722-123-456
Website: www.abcmanufacturing.com

Address Details:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Billing Address:
  Address Line 1: Industrial Area
  Address Line 2: Nairobi
  City: Nairobi
  County: Nairobi
  Postal Code: 00100
  Country: Kenya

Shipping Address: □ Same as Billing
  [If different, separate fields]

Tax Information:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Tax ID (PIN): P000123456A
Tax Category: Taxable / Exempt / Zero-Rated
VAT Registration: Yes/No
Tax Exemption Certificate: [Upload if applicable]

Financial Settings:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Currency: KES (can support multi-currency)
Default Price List: Corporate Pricing
Payment Terms: Net 30 Days
Credit Limit: 5,000,000 KES
Credit Days: 30 days
Payment Method: Bank Transfer / Cash / Check

Credit Control:
  Check Credit Limit: Yes
  Credit Limit Override Allowed: With Approval
  Current Outstanding: 1,250,000 KES
  Available Credit: 3,750,000 KES

Banking Information:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Bank Name: Equity Bank
Branch: Industrial Area
Account Number: 0123456789
SWIFT Code: EQBLKENA

Sales Settings:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Sales Person: Sarah Johnson
Sales Team: Enterprise Sales
Territory: Nairobi Corporate
Customer Category: Key Account / Regular / New

Loyalty & Preferences:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Loyalty Program: Gold Tier
Discount Percentage: 15% (approved)
Preferred Delivery Day: Thursday
Preferred Delivery Time: Morning (8-12)
Special Instructions: Requires delivery note in duplicate

Custom Fields (Configurable):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Industry Segment: Industrial Equipment
Business Size: Medium (50-200 employees)
Annual Revenue: 50-100M KES
Purchase Frequency: Monthly
Key Decision Maker: John Kamau

Attachments:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
□ Business Registration Certificate
□ Tax PIN Certificate
□ Bank Details Letter
□ Credit Application Form
□ Signed Terms & Conditions

Audit Trail:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Created By: admin@company.com
Created On: 2023-01-15 10:30:00
Last Modified By: sarah.johnson@company.com
Last Modified On: 2025-01-10 14:25:00
```

**Customer Categorization:**

```markdown
CUSTOMER GROUP STRUCTURE

By Business Type:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
├─ Retail
│  ├─ Walk-in Customers
│  ├─ Online Shoppers
│  └─ Small Businesses
├─ Wholesale
│  ├─ Distributors
│  ├─ Resellers
│  └─ Agents
├─ Corporate
│  ├─ Large Enterprises
│  ├─ SMEs
│  └─ Government
└─ Export
   ├─ East Africa
   ├─ Rest of Africa
   └─ International

By Territory:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
├─ Kenya
│  ├─ Nairobi
│  ├─ Mombasa
│  ├─ Kisumu
│  ├─ Nakuru
│  └─ Other Counties
├─ Uganda
├─ Tanzania
└─ Rwanda

By Industry:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
├─ Manufacturing
├─ Retail & Distribution
├─ Construction
├─ Healthcare
├─ Education
├─ Hospitality
├─ Agriculture
└─ Services

By Value Tier:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
├─ Platinum (> 10M annually)
├─ Gold (5-10M annually)
├─ Silver (1-5M annually)
└─ Bronze (< 1M annually)

PRICING & DISCOUNT IMPACT:

Customer Group: Corporate
├─ Base Discount: 10%
├─ Volume Discount: Additional 5% (>100 units)
├─ Payment Discount: 2% if paid within 10 days
└─ Special Promotions: Eligible

Customer Group: Retail
├─ Base Discount: 0%
├─ Loyalty Discount: 3% (for repeat customers)
├─ Payment Discount: None
└─ Special Promotions: Limited
```

### Contact Persons (Multi-Contact Support)

```markdown
MULTIPLE CONTACTS PER CUSTOMER

Customer: ABC Manufacturing Ltd

Contact 1 (Primary):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Name: John Kamau
Designation: Procurement Manager
Department: Purchasing
Email: john.kamau@abc.com
Phone: +254-700-123-456
Mobile: +254-722-123-456
Is Primary Contact: Yes
Receives: Quotations, Orders, Invoices
Decision Maker: Yes

Contact 2 (Finance):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Name: Mary Wanjiku
Designation: Finance Manager
Department: Finance
Email: mary.wanjiku@abc.com
Phone: +254-700-234-567
Is Primary Contact: No
Receives: Invoices, Payment Receipts, Statements
Decision Maker: For payments

Contact 3 (Technical):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Name: David Omondi
Designation: Technical Manager
Department: Engineering
Email: david.omondi@abc.com
Phone: +254-700-345-678
Is Primary Contact: No
Receives: Technical Specs, Delivery Notes
Decision Maker: For specifications

USAGE IN DOCUMENTS:

Quotation:
  Attention: John Kamau (Procurement)
  CC: David Omondi (for technical review)

Invoice:
  Attention: Mary Wanjiku (Finance)
  CC: John Kamau (FYI)

Delivery Note:
  Contact: David Omondi (to receive goods)
```

### Customer Addresses (Multi-Address Support)

```markdown
MULTIPLE ADDRESSES PER CUSTOMER

Customer: ABC Manufacturing Ltd

Address 1 (Head Office - Billing):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Address Type: Billing
Address Name: Head Office
Address Line 1: Mombasa Road, Industrial Area
Address Line 2: Building 45, Floor 3
City: Nairobi
County: Nairobi County
Postal Code: 00100
Country: Kenya
Is Primary: Yes
Is Billing: Yes
Is Shipping: No

Address 2 (Factory - Shipping):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Address Type: Shipping
Address Name: Factory Warehouse
Address Line 1: Thika Road, Exit 14
Address Line 2: KM 25
City: Ruiru
County: Kiambu County
Postal Code: 00232
Country: Kenya
Is Primary: No
Is Billing: No
Is Shipping: Yes
Contact Person: David Omondi
Phone: +254-700-345-678
Delivery Instructions: Gate closes at 5 PM

Address 3 (Branch Office):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Address Type: Shipping
Address Name: Mombasa Branch
Address Line 1: Moi Avenue
City: Mombasa
County: Mombasa County
Postal Code: 80100
Country: Kenya
Is Primary: No
Is Billing: No
Is Shipping: Yes

USAGE IN DOCUMENTS:

Sales Order:
  Billing Address: Head Office
  Shipping Address: Factory Warehouse
  (User can select from saved addresses)

Delivery Note:
  Deliver To: Factory Warehouse
  Contact: David Omondi (+254-700-345-678)
  Instructions: "Gate closes at 5 PM"
```

### Customer Groups

```markdown
CUSTOMER GROUP CONFIGURATION

Group: Corporate Accounts
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Description: Large corporate customers
Default Price List: Corporate Pricing
Default Discount: 10%
Payment Terms: Net 30 Days
Credit Limit Range: 1M - 10M KES
Requires Approval For:
  - Discount above 15%
  - Credit limit increase
Special Features:
  - Dedicated account manager
  - Quarterly business reviews
  - Priority support

Group: Wholesale Distributors
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Description: Wholesale buyers and distributors
Default Price List: Wholesale Pricing
Default Discount: 20%
Payment Terms: Net 15 Days
Credit Limit Range: 500K - 5M KES
Special Features:
  - Volume discounts available
  - Flexible delivery schedules
  - Marketing support

Group: Retail Customers
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Description: Walk-in and regular retail customers
Default Price List: Standard Retail Price
Default Discount: 0%
Payment Terms: Cash / Immediate
Credit Limit: 0 (Cash only)
Special Features:
  - Loyalty program eligible
  - Seasonal promotions
  - Member discounts

Group: Export Customers
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Description: International customers
Default Price List: Export Pricing (USD)
Default Discount: 0%
Payment Terms: LC / Advance Payment
Credit Limit: Case by case
Special Features:
  - Multi-currency support
  - Export documentation
  - International shipping
  - Zero-rated VAT
```

---

## Lead & Opportunity Management

### Lead Lifecycle

```markdown
LEAD MANAGEMENT PROCESS

Lead Sources:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Website inquiry form
- Trade shows and events
- Cold calling campaigns
- Referrals from existing customers
- Marketing campaigns (email, social media)
- Inbound calls
- Partner referrals
- Online advertising

Lead Capture:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Lead Name: John Kamau
Company: ABC Manufacturing Ltd
Email: john.kamau@abc.com
Phone: +254-700-123-456
Source: Website Inquiry
Interest: Industrial Equipment
Status: New

Lead Scoring (Automatic):
  Budget: Confirmed (30 points)
  Timeline: Immediate (20 points)
  Authority: Decision Maker (25 points)
  Need: Clear requirement (25 points)
  Total Score: 100/100 (Hot Lead)

Lead Assignment:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Assigned To: Sarah Johnson
Territory: Nairobi Corporate
Assignment Rule: Round-robin / Territory-based / Load-based
Notification: Email + In-app notification
Response SLA: Contact within 2 hours

Lead Qualification:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
BANT Framework:
  Budget: 5M KES confirmed ✓
  Authority: Procurement Manager ✓
  Need: Equipment upgrade required ✓
  Timeline: Within 2 months ✓

Status: QUALIFIED → Convert to Opportunity

Lead Disqualification Reasons:
  □ No budget
  □ Timeline too far in future
  □ Not decision maker
  □ No real need
  □ Competitor already selected
  □ Out of service area
```

### Lead Record Structure

```markdown
LEAD MASTER DATA

Basic Information:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Lead ID: LEAD-2025-001
Lead Name: John Kamau
Company: ABC Manufacturing Ltd
Designation: Procurement Manager
Industry: Manufacturing
Employee Count: 150
Annual Revenue: 80M KES

Contact Details:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Email: john.kamau@abc.com
Phone: +254-700-123-456
Mobile: +254-722-123-456
Website: www.abcmanufacturing.com
Address: Industrial Area, Nairobi

Lead Details:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Source: Website Inquiry
Campaign: Q1 2025 Equipment Campaign
Lead Owner: Sarah Johnson
Territory: Nairobi Corporate
Status: Qualified
Rating: Hot
Lead Score: 100/100

Qualification Details:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Budget: 5,000,000 KES
Timeline: Q1 2025
Requirements: 
  - 3 industrial machines
  - Installation and training
  - Warranty support
Decision Maker: Yes
Approval Process: Board approval required
Competitors: Company X, Company Y

Activities:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌──────────┬──────────┬─────────────────────────┐
│ Date     │ Type     │ Description             │
├──────────┼──────────┼─────────────────────────┤
│ Jan 10   │ Call     │ Initial contact made    │
│ Jan 11   │ Email    │ Sent product brochure   │
│ Jan 13   │ Meeting  │ Site visit scheduled    │
│ Jan 15   │ Call     │ Budget confirmed        │
└──────────┴──────────┴─────────────────────────┘

Next Steps:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
□ Schedule site visit (Jan 20)
□ Prepare technical proposal
□ Arrange equipment demo
□ Create formal quotation

Notes:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
"Customer currently using competitor equipment.
Contract expires March 2025. Looking for better
pricing and service support. Key decision maker
is CEO but John has strong influence."
```

### Opportunity Management

```markdown
OPPORTUNITY STRUCTURE

Basic Information:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Opportunity ID: OPP-2025-001
Opportunity Name: ABC Manufacturing - Equipment Purchase
Customer/Lead: ABC Manufacturing Ltd (Lead converted)
Contact Person: John Kamau
Account Owner: Sarah Johnson
Territory: Nairobi Corporate

Opportunity Details:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Type: New Business / Upsell / Renewal / Cross-sell
Opportunity Amount: 5,000,000 KES
Probability: 60%
Weighted Amount: 3,000,000 KES (Amount × Probability)
Expected Close Date: 2025-02-28
Sales Stage: Proposal/Quote

Sales Stages & Probability:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
├─ Prospecting (10%)
├─ Qualification (20%)
├─ Needs Analysis (30%)
├─ Proposal/Quote (60%) ← Current Stage
├─ Negotiation (80%)
├─ Closed Won (100%)
└─ Closed Lost (0%)

Stage History:
  Prospecting: Jan 10 - Jan 12 (2 days)
  Qualification: Jan 12 - Jan 14 (2 days)
  Needs Analysis: Jan 14 - Jan 18 (4 days)
  Proposal: Jan 18 - Present (ongoing)

Products/Services:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌────────────────┬──────┬────────────┬─────────────┐
│ Product        │ Qty  │ Unit Price │ Total       │
├────────────────┼──────┼────────────┼─────────────┤
│ Machine Model A│  1   │ 2,000,000  │ 2,000,000   │
│ Machine Model B│  1   │ 1,800,000  │ 1,800,000   │
│ Machine Model C│  1   │ 1,200,000  │ 1,200,000   │
│                │      │            │             │
│ TOTAL          │      │            │ 5,000,000   │
└────────────────┴──────┴────────────┴─────────────┘

Competition Analysis:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Competitors: Company X, Company Y
Our Advantages:
  ✓ Better pricing (10% lower)
  ✓ Local support and service
  ✓ Faster delivery (2 weeks vs 6 weeks)
  ✓ Flexible payment terms
  ✓ Existing relationship with parent company

Their Advantages:
  ✗ Established brand name
  ✗ Current installed base
  ✗ Long-term service contract

Win Strategy:
  - Emphasize cost savings
  - Highlight local support advantage
  - Offer equipment trial period
  - Flexible payment terms

Decision Criteria:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Most Important:
  1. Total cost of ownership (40%)
  2. Service and support (30%)
  3. Delivery timeline (20%)
  4. Brand reputation (10%)

Decision Makers:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌────────────────┬──────────────┬──────────────┐
│ Name           │ Role         │ Influence    │
├────────────────┼──────────────┼──────────────┤
│ CEO            │ Final Approver│ High        │
│ John Kamau     │ Recommender  │ High        │
│ Finance Manager│ Budget Holder│ Medium      │
│ Technical Head │ User         │ Medium      │
└────────────────┴──────────────┴──────────────┘

Next Actions:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
□ Present formal quotation (Jan 20)
□ Schedule equipment demo (Jan 22)
□ CEO meeting for final approval (Jan 25)
□ Follow up on quotation (Jan 27)
□ Negotiate final terms

Forecast Category:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Pipeline: Likely to close
Commit: Expected to close this quarter
Best Case: Optimistic forecast
Closed: Already won

Current: COMMIT (60% probability, expected Feb close)
```

### Sales Pipeline Visualization

```markdown
SALES PIPELINE DASHBOARD

By Stage:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Prospecting       (10%)   15 opportunities   15M KES
Qualification     (20%)   12 opportunities   20M KES
Needs Analysis    (30%)    8 opportunities   18M KES
Proposal          (60%)    5 opportunities   25M KES ⭐
Negotiation       (80%)    3 opportunities   20M KES
─────────────────────────────────────────────────────
Total Pipeline:           43 opportunities   98M KES
Weighted Pipeline:                           52M KES

By Sales Person:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Sarah Johnson:    12 opportunities   35M KES
Mike Chen:        10 opportunities   28M KES
Jane Mwangi:       8 opportunities   20M KES
Tom Omondi:        7 opportunities   15M KES

By Territory:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Nairobi:          25 opportunities   55M KES
Mombasa:          10 opportunities   25M KES
Kisumu:            5 opportunities   12M KES
Nakuru:            3 opportunities    6M KES

Expected Close This Quarter:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
January:           8 opportunities   18M KES
February:         12 opportunities   25M KES ⭐
March:            10 opportunities   20M KES
─────────────────────────────────────────────────────
Q1 Total:         30 opportunities   63M KES
Q1 Target:                            50M KES
Achievement:                            126% ✓

Opportunity Aging:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
< 30 days:        20 opportunities   40M KES
30-60 days:       15 opportunities   35M KES
60-90 days:        5 opportunities   15M KES ⚠
> 90 days:         3 opportunities    8M KES ⚠⚠

⚠ Review opportunities older than 60 days
⚠⚠ Opportunities older than 90 days - close or disqualify
```

---

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

## Sales Order Processing

### Sales Order Creation

```markdown
SALES ORDER WORKFLOW

Creation Methods:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Method 1: From Quotation (Most Common)
  1. Open accepted quotation
  2. Click "Create Sales Order"
  3. All data auto-populated
  4. Review and confirm

Method 2: Direct Entry
  - For existing customers
  - Repeat orders
  - Phone/email orders
  - Walk-in sales

Method 3: From Portal
  - Customer self-service portal
  - Online order placement
  - Auto-validated against credit limits

Sales Order Header:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Sales Order No: SO-2025-001 (auto-generated)
Date: 2025-01-25
Customer: ABC Manufacturing Ltd
Contact: John Kamau
Customer PO No: PO/ABC/2025/045 (required)

Reference:
  Quotation: QTN-2025-001
  Opportunity: OPP-2025-001

Addresses:
  Billing: Head Office, Nairobi
  Shipping: Factory Warehouse, Ruiru

Order Details:
  Order Type: Sales / Service / Project
  Price List: Corporate Pricing
  Currency: KES
  Payment Terms: Net 30 Days
  Delivery Date: 2025-02-10
  
Sales Team:
  Sales Person: Sarah Johnson
  Commission: 5% on net amount
  Territory: Nairobi Corporate

Order Items:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌────────┬─────────────┬─────┬───────┬───────────┐
│ Item   │ Description │ Qty │ Rate  │ Amount    │
├────────┼─────────────┼─────┼───────┼───────────┤
│ MDL-A  │ Machine A   │  1  │1,900K │ 1,900,000 │
│        │ Stock: 2    │     │       │           │
│        │ Reserved: 1 │     │       │           │
│        │ Available:1 │     │       │           │
│        │ Warehouse:  │     │       │           │
│        │ Main Store  │     │       │           │
├────────┼─────────────┼─────┼───────┼───────────┤
│ MDL-B  │ Machine B   │  1  │1,620K │ 1,620,000 │
│        │ Stock: 1 ✓  │     │       │           │
├────────┼─────────────┼─────┼───────┼───────────┤
│ MDL-C  │ Machine C   │  1  │1,080K │ 1,080,000 │
│        │ Stock: 0 ⚠  │     │       │           │
│        │ On Order:2  │     │       │           │
│        │ Expected:   │     │       │           │
│        │ Feb 5, 2025 │     │       │           │
└────────┴─────────────┴─────┴───────┴───────────┘

⚠ Item MDL-C on backorder
Expected delivery: Feb 5
Confirm with customer: Partial delivery or wait?

Stock Reservation:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
On Order Confirmation:
  ✓ Reserve MDL-A (Qty: 1) from Main Store
  ✓ Reserve MDL-B (Qty: 1) from Main Store
  ⏳ MDL-C pending stock arrival

Reserved Stock:
  - Cannot be sold to other customers
  - Automatically allocated for this order
  - Released if order cancelled
  - Ages if not delivered (alert after 30 days)

Order Totals:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Item Total:           4,600,000 KES
  Additional Discount:    (92,000) KES  (2%)
  ───────────────────────────────────
  Net Amount:           4,508,000 KES
  VAT (16%):              721,280 KES
  ───────────────────────────────────
  Grand Total:          5,229,280 KES
  
  Deposit Required (50%): 2,614,640 KES

Payment Schedule:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌──────────┬─────────────┬────────────┬──────────┐
│ Due Date │ Description │ Amount     │ Status   │
├──────────┼─────────────┼────────────┼──────────┤
│ Jan 25   │ Deposit(50%)│ 2,614,640  │ Pending  │
│ Feb 10   │ On Delivery │ 2,614,640  │ Pending  │
└──────────┴─────────────┴────────────┴──────────┘

Credit Check:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Customer: ABC Manufacturing Ltd
Credit Limit: 5,000,000 KES
Current Outstanding: 1,250,000 KES
This Order: 5,229,280 KES
─────────────────────────────────
Total Exposure: 6,479,280 KES
Credit Available: (1,479,280) KES ⚠

⚠ CREDIT LIMIT EXCEEDED

Options:
  □ Request credit limit increase
  □ Require deposit payment first
  □ Get management approval to proceed

Action: Deposit payment required before confirmation
```

### Sales Order Confirmation Process

```markdown
ORDER CONFIRMATION WORKFLOW

Step 1: Validation
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
System Checks:
  ✓ Customer details complete
  ✓ Customer PO number provided
  ✓ Items available or on backorder
  ✓ Pricing authorized
  ✓ Credit limit (if applicable)
  ✓ Delivery date feasible
  ✓ Payment terms agreed

If Credit Check Fails:
  → Workflow: Send for approval
  → Approver: Credit Manager
  → Options: Approve / Reject / Require Deposit

Step 2: Approval (if required)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Approval Required For:
  - Orders exceeding credit limit
  - Special discount > approval threshold
  - First order from new customer
  - High-value orders (> 5M KES)
  - Backorder situations

Approval Workflow:
  Draft Order → Pending Approval → Approved → Confirmed

Notification:
  To: Credit Manager / Sales Manager
  Subject: Sales Order Approval Required
  Content:
    - Customer details
    - Order value
    - Credit exposure
    - Reason for approval
  
Approver Actions:
  □ Approve (proceed to confirmation)
  □ Approve with conditions (e.g., deposit required)
  □ Reject (order cancelled, customer notified)

Step 3: Confirmation
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
On Confirmation:
  Status: DRAFT → CONFIRMED
  
Automatic Actions:
  1. Stock Reserved:
     - Items allocated from warehouse
     - Available stock updated
     - Reservation logged
     
  2. Production/Procurement Triggered:
     If Make-to-Order:
       → Work Order created
       → Material requirements calculated
       → Production scheduled
     
     If Buy-to-Order:
       → Purchase Requisition created
       → Supplier notification
       → Expected delivery tracked
     
  3. Customer Notification:
     Email sent:
       Subject: Order Confirmation SO-2025-001
       Attachment: Order confirmation PDF
       Content: 
         - Order summary
         - Expected delivery date
         - Payment instructions
         - Tracking link
  
  4. Team Notifications:
     → Warehouse: Prepare items for delivery
     → Finance: Expect deposit payment
     → Customer Service: Order in system
  
  5. Integration Events:
     → Inventory: Stock reserved
     → Finance: Accounts receivable prepared
     → Delivery: Shipment scheduled

Step 4: Order Processing
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Order Status Tracking:
  CONFIRMED → TO DELIVER → DELIVERED → TO BILL → COMPLETED

Sub-statuses:
  - Awaiting Payment (deposit)
  - Awaiting Stock (backorder)
  - In Production
  - Ready to Deliver
  - Partially Delivered
  - Awaiting Invoice Approval
```

### Sales Order Modifications

```markdown
MODIFYING CONFIRMED ORDERS

Amendment Scenarios:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Scenario 1: Before Delivery
  Customer requests to add items
  
  Process:
    1. Create Sales Order Amendment
    2. Add new items
    3. Recalculate totals
    4. Check stock availability
    5. Get customer confirmation
    6. Update original order
  
  Impact:
    - Stock re-reserved
    - New delivery date if needed
    - Revised invoice amount
    - Customer notified

Scenario 2: Change Delivery Date
  Customer needs earlier/later delivery
  
  Process:
    1. Check warehouse availability
    2. Verify production schedule
    3. Update delivery date
    4. Notify warehouse team
    5. Email customer confirmation
  
  Impact:
    - Delivery schedule updated
    - Warehouse notified
    - No financial impact

Scenario 3: Quantity Reduction
  Customer wants fewer items
  
  Process:
    1. Update quantities
    2. Release excess reserved stock
    3. Recalculate amounts
    4. Update payment schedule
    5. Issue revised order confirmation
  
  Impact:
    - Stock released back to available
    - Lower invoice amount
    - Potential refund if deposit paid

Scenario 4: Item Substitution
  Item unavailable, offer alternative
  
  Process:
    1. Propose alternative item
    2. Get customer approval
    3. Update order with new item
    4. Adjust pricing if different
    5. Reserve new item stock
  
  Impact:
    - Original item stock released
    - New item stock reserved
    - Price adjustment (+ or -)

Restrictions:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Cannot Modify If:
  ✗ Items already delivered
  ✗ Invoice already generated
  ✗ Payment already received

Must Cancel & Re-create If:
  ✗ Major changes (>50% of order)
  ✗ Complete product change
  ✗ Different customer

Modification Audit Trail:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌──────────┬─────────┬────────────────────────┐
│ Date     │ User    │ Change                 │
├──────────┼─────────┼────────────────────────┤
│ Jan 25   │ Sarah   │ Order created          │
│ Jan 25   │ System  │ Status: Confirmed      │
│ Jan 27   │ Sarah   │ Delivery date changed  │
│          │         │ From: Feb 10 → Feb 15  │
│ Jan 28   │ Mike    │ Qty changed: Item B    │
│          │         │ From: 1 → 2 units      │
│ Jan 28   │ System  │ Amount recalculated    │
│          │         │ New total: 7,049,280   │
└──────────┴─────────┴────────────────────────┘
```

### Sales Order Cancellation

```markdown
ORDER CANCELLATION PROCESS

Cancellation Reasons:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Customer request
- Payment not received
- Items not available
- Duplicate order
- Customer credit issues
- Force majeure

Before Delivery:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Can Cancel If:
  ✓ No delivery made
  ✓ No invoice generated
  ✓ Or minor deposit paid (refundable)

Cancellation Process:
  1. Verify cancellation authority
  2. Check payment status
  3. Release reserved stock
  4. Cancel production/purchase orders
  5. Process refund if deposit paid
  6. Update customer record
  7. Notify all stakeholders

Automatic Actions:
  → Stock: Reserved items released
  → Production: Work orders cancelled
  → Purchasing: Purchase orders cancelled
  → Finance: Refund processed (if applicable)
  → Customer: Cancellation email sent

Partial Cancellation:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Cancel Specific Items Only:
  - Remove items from order
  - Release their stock
  - Recalculate order total
  - Update delivery schedule

Example:
  Original Order: Items A, B, C
  Cancel Item C
  Result: Order continues with A, B only

After Delivery:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Cannot Cancel:
  ✗ Items already delivered
  ✗ Must process as RETURN instead

Return Process:
  1. Create Sales Return
  2. Receive items back
  3. Inspect condition
  4. Issue credit note
  5. Process refund

Cancellation Charges:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Policy (configurable):
  - If cancelled within 24 hours: No charge
  - If cancelled before production: 10% charge
  - If production started: 25% charge
  - If custom/special order: 50% charge

Deposit Handling:
  Original Deposit: 2,614,640 KES
  Cancellation Fee: 261,464 KES (10%)
  Refund Amount: 2,353,176 KES

Financial Entry:
  Dr. Sales Deposit Account    2,614,640
      Cr. Cash/Bank                     2,353,176
      Cr. Cancellation Income             261,464
```

---

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

## Sales Invoicing

### Invoice Creation Process

```markdown
SALES INVOICE GENERATION

Creation Methods:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Method 1: From Delivery Note (Most Common)
  Trigger: Goods delivered, customer signed
  Process:
    1. Open delivery note DN-2025-001
    2. Click "Create Sales Invoice"
    3. All data auto-populated
    4. Review and submit

Method 2: From Sales Order (Direct)
  Use Case: Service delivery, advance invoice
  Process:
    1. Select sales order
    2. Create invoice without delivery note
    3. Revenue recognized on invoice submission

Method 3: Manual Invoice
  Use Case: Ad-hoc sales, adjustments
  Process:
    1. Create new invoice
    2. Select customer
    3. Add items manually
    4. Submit

Method 4: Recurring Invoice
  Use Case: Subscriptions, maintenance contracts
  Process:
    1. Set up recurring invoice template
    2. System auto-generates on schedule
    3. Auto-email to customer

Invoice Header:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Invoice No: INV-2025-001 (auto-generated)
Invoice Type: Tax Invoice / Proforma / Commercial
Invoice Date: 2025-02-10
Due Date: 2025-03-12 (Net 30 days)

Reference Documents:
  Sales Order: SO-2025-001
  Delivery Note: DN-2025-001
  Customer PO: PO/ABC/2025/045
  Quotation: QTN-2025-001

Customer Details:
  Customer: ABC Manufacturing Ltd
  PIN/Tax ID: P000123456A
  Billing Address: Head Office, Nairobi
  Attention: Mary Wanjiku (Finance Manager)
  Email: mary.wanjiku@abc.com

Company Details (Auto):
  Company PIN: P987654321Z
  Address: [From company master]
  Bank Details: [For payment]

Invoice Items:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Based on Delivery DN-2025-001:
┌────────┬──────────────┬─────┬────────────┬────────────┐
│ Item   │ Description  │ Qty │ Rate       │ Amount     │
├────────┼──────────────┼─────┼────────────┼────────────┤
│ MDL-A  │ Machine A    │  1  │ 1,900,000  │ 1,900,000  │
│        │ SN: 12345    │     │            │            │
│        │ Delivered:   │     │            │            │
│        │ Feb 10, 2025 │     │            │            │
├────────┼──────────────┼─────┼────────────┼────────────┤
│ MDL-B  │ Machine B    │  1  │ 1,620,000  │ 1,620,000  │
│        │ SN: 12346    │     │            │            │
├────────┼──────────────┼─────┼────────────┼────────────┤
│ SVC-01 │ Installation │  1  │   150,000  │   150,000  │
│        │ Completed:   │     │            │            │
│        │ Feb 12, 2025 │     │            │            │
└────────┴──────────────┴─────┴────────────┴────────────┘

Note: MDL-C not invoiced yet (not delivered)

Invoice Calculations:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Item Total:                3,670,000 KES
Additional Discount (2%):    (73,400) KES
───────────────────────────────────────
Net Amount:                3,596,600 KES
VAT (16%):                   575,456 KES
───────────────────────────────────────
Grand Total:               4,172,056 KES

Less: Deposit Applied:    (2,614,640) KES
───────────────────────────────────────
Balance Due:               1,557,416 KES

Payment Schedule:
  Deposit (Received Jan 25): 2,614,640 KES ✓ PAID
  Balance (Due Mar 12):      1,557,416 KES ⏳ PENDING

Financial Integration:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
On Invoice Submission:

Journal Entry (Automatic):
Dr. Accounts Receivable - ABC     4,172,056
    Cr. Sales Revenue - Equipment       3,596,600
    Cr. VAT Payable                       575,456

Description: Sales Invoice INV-2025-001
Customer: ABC Manufacturing Ltd
Sales Order: SO-2025-001
Cost Center: Sales Department
Sales Person: Sarah Johnson

If Deposit Already Paid:
Dr. Customer Deposit Account      2,614,640
    Cr. Accounts Receivable - ABC       2,614,640

Description: Apply deposit from SO-2025-001

Net Receivable:
Opening AR Balance:         1,250,000
New Invoice:               4,172,056
Less: Deposit Applied:    (2,614,640)
───────────────────────────────────
Current AR Balance:        2,807,416 KES

Cost of Goods Sold (from Delivery):
Dr. Cost of Goods Sold            3,200,000
    Cr. Inventory - Equipment           3,200,000

Description: COGS for DN-2025-001
Items: MDL-A (1,600,000), MDL-B (1,600,000)

Margin Analysis:
Revenue:                   3,596,600 KES
Less: COGS:               (3,200,000) KES
───────────────────────────────────
Gross Profit:               396,600 KES
Gross Margin:                  11.0%

Commission Calculation:
Net Revenue:              3,596,600 KES
Commission Rate:                  5%
Commission Amount:          179,830 KES
Payable to: Sarah Johnson
```

### Invoice Document Format

```markdown
TAX INVOICE LAYOUT

┌──────────────────────────────────────────────┐
│           [COMPANY LOGO]                     │
│                                              │
│          TAX INVOICE                         │
│                                              │
│ Company Name                Invoice No:      │
│ PIN: P987654321Z            INV-2025-001     │
│ Address Line 1              Date: Feb 10, 25 │
│ Address Line 2              Due: Mar 12, 25  │
│ Phone: +254-20-xxx-xxxx                      │
│ Email: sales@company.com                     │
├──────────────────────────────────────────────┤
│ BILL TO:                                     │
│                                              │
│ ABC Manufacturing Ltd                        │
│ PIN: P000123456A                             │
│ Head Office, Industrial Area                 │
│ Nairobi, Kenya                               │
│                                              │
│ Attention: Mary Wanjiku, Finance Manager     │
│ Email: mary.wanjiku@abc.com                  │
│ Phone: +254-700-234-567                      │
├──────────────────────────────────────────────┤
│ REFERENCE:                                   │
│ Sales Order: SO-2025-001                     │
│ Your PO: PO/ABC/2025/045                     │
│ Delivery Note: DN-2025-001                   │
│ Quotation: QTN-2025-001                      │
├──────────────────────────────────────────────┤
│ Item  Description       Qty  Rate     Amount │
│ ────  ──────────────    ─── ──────   ─────── │
│ 1.    Machine Model A    1  1,900,000        │
│       Serial: SN12345          1,900,000     │
│       Delivered: Feb 10, 2025                │
│                                              │
│ 2.    Machine Model B    1  1,620,000        │
│       Serial: SN12346          1,620,000     │
│       Delivered: Feb 10, 2025                │
│                                              │
│ 3.    Installation       1    150,000        │
│       Service                    150,000     │
│       Completed: Feb 12, 2025                │
│                                              │
│                       Subtotal: 3,670,000    │
│                       Discount:   (73,400)   │
│                       Net:      3,596,600    │
│                       VAT(16%):   575,456    │
│                       ─────────────────────  │
│                       TOTAL:    4,172,056    │
│                                              │
│ LESS: DEPOSIT PAID                           │
│ Payment Ref: PAY-2025-001  (2,614,640)      │
│ Date: January 25, 2025                       │
│                       ─────────────────────  │
│                    BALANCE DUE: 1,557,416    │
│                                              │
│ Amount in Words:                             │
│ One Million Five Hundred Fifty Seven         │
│ Thousand Four Hundred Sixteen Shillings Only │
├──────────────────────────────────────────────┤
│ PAYMENT INSTRUCTIONS:                        │
│                                              │
│ Bank: Equity Bank Kenya                      │
│ Account Name: Company Name Ltd               │
│ Account Number: 0123456789                   │
│ Branch: Industrial Area                      │
│ Swift Code: EQBLKENA                         │
│                                              │
│ M-Pesa Till: 123456 (for amounts <100K)     │
│                                              │
│ Payment Reference: INV-2025-001              │
├──────────────────────────────────────────────┤
│ PAYMENT TERMS:                               │
│ Net 30 days from invoice date                │
│ Due Date: March 12, 2025                     │
│ Late payment interest: 2% per month          │
├──────────────────────────────────────────────┤
│ NOTES:                                       │
│ • Warranty: 12 months from delivery date     │
│ • For queries: accounts@company.com          │
│ • This is a computer-generated invoice       │
│                                              │
│ [QR Code for Payment]                        │
│                                              │
│ Thank you for your business!                 │
└──────────────────────────────────────────────┘
```

### Invoice Status Management

```markdown
INVOICE LIFECYCLE

Status Flow:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
DRAFT → SUBMITTED → PAID / PARTIALLY PAID / OVERDUE

Status Details:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

DRAFT:
  - Invoice being prepared
  - Can be edited
  - Not posted to GL
  - Not sent to customer
  - Not affecting AR

SUBMITTED:
  - Invoice finalized and sent
  - Posted to GL
  - AR balance updated
  - Email sent to customer
  - Read-only (cannot edit)
  - Payment tracking active

PAID:
  - Full payment received
  - AR cleared
  - Payment allocated
  - Receipt issued
  - Commission released

PARTIALLY PAID:
  - Part payment received
  - Balance outstanding tracked
  - Aging starts on balance
  - Follow-up for balance

OVERDUE:
  - Past due date
  - Payment not received
  - Collection actions triggered
  - Aging bucket assigned
  - Interest may apply

CANCELLED:
  - Invoice cancelled before payment
  - AR reversed
  - Must have authorization
  - Reason documented

Invoice Actions:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

From SUBMITTED:
  □ Record Payment
  □ Send Reminder
  □ Generate Statement
  □ Create Credit Note
  □ Cancel (with approval)

From PAID:
  □ Generate Receipt
  □ View Payment History
  □ Issue Credit Note (for returns)

From OVERDUE:
  □ Send Reminder (automatic)
  □ Escalate to Collections
  □ Apply Late Fee
  □ Record Payment

Aging Analysis:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Invoice: INV-2025-001
Invoice Date: Feb 10, 2025
Due Date: Mar 12, 2025
Balance Due: 1,557,416 KES

As of Date: Mar 15, 2025
Status: OVERDUE
Days Overdue: 3 days
Aging Bucket: Current (0-30 days)

Reminder Actions:
  Mar 10 (2 days before): Friendly reminder ✓
  Mar 13 (1 day after): First reminder ✓
  Mar 20 (8 days after): Second reminder ⏳
  Mar 27 (15 days after): Final notice ⏳
  Apr 10 (30 days after): Escalate to collections ⏳

Payment Tracking:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌────────────┬────────────┬────────────┬────────────┐
│ Date       │ Reference  │ Amount     │ Balance    │
├────────────┼────────────┼────────────┼────────────┤
│ Jan 25     │ Deposit    │ 2,614,640  │ 1,557,416  │
│ Mar 15     │ Partial    │   500,000  │ 1,057,416  │
│ Mar 20     │ Balance    │ 1,057,416  │         0  │
└────────────┴────────────┴────────────┴────────────┘
```

### Proforma Invoice

```markdown
PROFORMA INVOICE

Purpose:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Advance invoice for customs/import
- Bank LC requirements
- Budget approval documentation
- Not a demand for payment
- Not posted to accounting

Key Differences from Tax Invoice:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Proforma Invoice:
  ✗ Not posted to GL
  ✗ Not creating AR
  ✗ No tax obligation
  ✗ Not for payment demand
  ✓ For information/planning only
  ✓ Can be revised freely
  ✓ No accounting impact

Tax Invoice:
  ✓ Posted to GL
  ✓ Creates AR
  ✓ Tax obligation created
  ✓ Legal demand for payment
  ✗ Cannot be revised (need credit note)
  ✓ Full accounting impact

Proforma Usage:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Export Sales:
   Customer needs proforma for import clearance
   
2. Large Projects:
   Customer needs quotation in invoice format
   for budget approval
   
3. Government/Tender:
   Required for procurement process

Conversion Process:
  Proforma Created → Customer Approves → 
  Sales Order → Delivery → Tax Invoice

Document Marking:
  Header: "PROFORMA INVOICE"
  Footer: "THIS IS NOT A TAX INVOICE"
  Watermark: "PROFORMA - FOR PLANNING ONLY"
```

---

## Payment Collection & Allocation

### Payment Entry Process

```markdown
PAYMENT RECEIPT WORKFLOW

Payment Sources:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Bank transfer
- Cash payment
- Check payment
- Mobile money (M-Pesa, Airtel Money)
- Credit card
- Online payment gateway

Payment Entry:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Payment Entry No: PAY-2025-015 (auto)
Payment Date: 2025-03-15
Customer: ABC Manufacturing Ltd

Payment Details:
  Amount Received: 1,557,416 KES
  Payment Method: Bank Transfer
  Bank: Equity Bank
  Reference: TRX/2025/54321
  Received In: Company Main Account
  
Party Details:
  Paid By: ABC Manufacturing Ltd
  Account: Accounts Receivable - ABC
  
Allocation Method:
  ○ Auto-allocate (oldest first - FIFO)
  ○ Manual allocation
  ● Specific invoice selection

Outstanding Invoices:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌──────────────┬─────────┬───────────┬───────────┐
│ Invoice      │ Date    │ Amount    │ Allocate  │
├──────────────┼─────────┼───────────┼───────────┤
│ INV-2025-001 │ Feb 10  │ 1,557,416 │ 1,557,416 │
│ INV-2024-125 │ Dec 20  │   250,000 │         0 │
│ INV-2025-010 │ Jan 30  │   180,000 │         0 │
└──────────────┴─────────┴───────────┴───────────┘

Selected for Payment:
  INV-2025-001: 1,557,416 KES (FULL PAYMENT)
  
Payment Allocation:
  Total Received: 1,557,416 KES
  Total Allocated: 1,557,416 KES
  Unallocated: 0 KES ✓

Financial Entry:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Journal Entry (Automatic):

Dr. Bank - Main Account            1,557,416
    Cr. Accounts Receivable - ABC        1,557,416

Description: Payment for INV-2025-001
Reference: TRX/2025/54321
Payment Entry: PAY-2025-015

Invoice Status Update:
  INV-2025-001:
    Status: PARTIALLY PAID → PAID ✓
    Outstanding: 1,557,416 → 0
    Payment Date: 2025-03-15
    Days to Payment: 33 days (from invoice date)
    
Customer Account Summary:
  Previous Balance: 2,807,416 KES
  Payment Received: (1,557,416) KES
  ──────────────────────────────
  Current Balance: 1,250,000 KES

Automatic Actions:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Email Receipt to Customer:
   Subject: Payment Receipt - PAY-2025-015
   Attachment: Official receipt PDF
   
2. Update Customer Credit:
   Available Credit Increased by 1,557,416 KES
   
3. Release Commission:
   If payment-based commission:
     Invoice: INV-2025-001
     Commission: 179,830 KES
     Payable to: Sarah Johnson
     Status: EARNED (payment received)
     
4. Cancel Payment Reminders:
   Stop reminder emails for INV-2025-001
   
5. Update Reporting:
   DSO calculation
   Collection metrics
   Cash flow forecast
```

### Payment Allocation Scenarios

```markdown
ALLOCATION SCENARIOS

Scenario 1: Exact Match
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Payment: 1,557,416 KES
Invoice: INV-2025-001 = 1,557,416 KES

Result: Perfect match, fully allocate ✓

Scenario 2: Partial Payment
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Payment: 500,000 KES
Invoice: INV-2025-001 = 1,557,416 KES

Allocation:
  Paid: 500,000 KES
  Balance: 1,057,416 KES
  Status: PARTIALLY PAID
  
Follow-up:
  - Send acknowledgment
  - Request balance payment
  - Track remaining amount

Scenario 3: Overpayment
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Payment: 1,600,000 KES
Invoice: INV-2025-001 = 1,557,416 KES

Allocation:
  To Invoice: 1,557,416 KES
  Excess: 42,584 KES
  
Excess Handling Options:
  ○ Credit to customer account (advance payment)
  ○ Refund to customer
  ○ Allocate to other outstanding invoices
  
Selected: Credit to account
Entry:
  Dr. Bank                      1,600,000
      Cr. Accounts Receivable         1,557,416
      Cr. Customer Advance               42,584

Scenario 4: Multiple Invoice Payment
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Payment: 2,000,000 KES

Outstanding Invoices:
  INV-2025-001: 1,557,416 KES
  INV-2024-125:   250,000 KES
  INV-2025-010:   180,000 KES
  Total: 1,987,416 KES

Allocation (Auto - Oldest First):
  INV-2024-125: 250,000 KES (PAID) ✓
  INV-2025-001: 1,557,416 KES (PAID) ✓
  INV-2025-010: 180,000 KES (PAID) ✓
  Excess: 12,584 KES (advance credit)

All three invoices marked PAID ✓

Scenario 5: Payment Without Reference
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Bank Statement: 500,000 KES from "ABC Mfg"
No invoice reference

Process:
  1. Create unallocated payment entry
  2. Contact customer for invoice reference
  3. Manual allocation once confirmed
  
Temporary Entry:
  Dr. Bank                       500,000
      Cr. Unallocated Payments        500,000
  
After Confirmation:
  Dr. Unallocated Payments       500,000
      Cr. Accounts Receivable - ABC   500,000

Scenario 6: Early Payment Discount
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Invoice Terms: 2/10 Net 30
  (2% discount if paid within 10 days)

Invoice Amount: 1,000,000 KES
Invoice Date: Mar 1
Due Date: Mar 31
Discount Valid Until: Mar 11

Payment Date: Mar 8 ✓ (within discount period)
Payment Amount: 980,000 KES (2% discount taken)

Entry:
  Dr. Bank                       980,000
  Dr. Sales Discount              20,000
      Cr. Accounts Receivable         1,000,000

Invoice Status: PAID ✓
Discount Given: 20,000 KES
Effective Discount: 2%
```

### Payment Reconciliation

```markdown
BANK RECONCILIATION INTEGRATION

Daily Bank Statement Import:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Date: Mar 15, 2025

Bank Statement Lines:
┌──────────┬─────────────┬──────────┬───────────┐
│ Date     │ Description │ Debit    │ Credit    │
├──────────┼─────────────┼──────────┼───────────┤
│ Mar 15   │ TRX/54321   │          │ 1,557,416 │
│          │ ABC MFG     │          │           │
├──────────┼─────────────┼──────────┼───────────┤
│ Mar 15   │ TRX/54322   │          │   850,000 │
│          │ XYZ LTD     │          │           │
├──────────┼─────────────┼──────────┼───────────┤
│ Mar 15   │ Bank Fees   │    2,500 │           │
└──────────┴─────────────┴──────────┴───────────┘

Automatic Matching:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Statement Line 1: TRX/54321, 1,557,416
  ↓ Match Rules:
    - Amount matches
    - Customer name contains "ABC"
    - Reference in description
  ↓ Matched to:
    Payment Entry: PAY-2025-015 ✓
    Status: RECONCILED

Statement Line 2: TRX/54322, 850,000
  ↓ Search for matching payment
    - Amount: 850,000
    - Customer: XYZ
  ↓ Matched to:
    Payment Entry: PAY-2025-016 ✓
    Status: RECONCILED

Statement Line 3: Bank Fees, 2,500
  ↓ No matching payment entry
  ↓ Create GL Entry:
    Dr. Bank Charges Expense    2,500
        Cr. Bank Account              2,500
  Status: RECONCILED (expense entry)

Unreconciled Items:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Payment Entries Not in Bank:
  PAY-2025-017: 450,000 KES (Check not cleared)
  Reason: Outstanding check
  Action: Wait for clearance

Bank Entries Not Matched:
  None ✓ All reconciled

Reconciliation Summary:
  Opening Balance: 5,250,000 KES
  Total Receipts: 2,407,416 KES
  Total Payments: 2,500 KES
  Closing Balance: 7,654,916 KES ✓
  
  GL Balance: 7,654,916 KES ✓
  Difference: 0 ✓ RECONCILED
```

### Payment Receipt Document

```markdown
OFFICIAL RECEIPT FORMAT

┌────────────────────────────────────────────┐
│         [COMPANY LOGO]                     │
│                                            │
│        OFFICIAL RECEIPT                    │
│                                            │
│ Receipt No: RCP-2025-015                   │
│ Date: March 15, 2025                       │
│                                            │
│ Company PIN: P987654321Z                   │
├────────────────────────────────────────────┤
│ RECEIVED FROM:                             │
│                                            │
│ ABC Manufacturing Ltd                      │
│ PIN: P000123456A                           │
│ Head Office, Industrial Area               │
│ Nairobi, Kenya                             │
├────────────────────────────────────────────┤
│ THE SUM OF:                                │
│                                            │
│ KES 1,557,416.00                           │
│                                            │
│ (One Million Five Hundred Fifty Seven      │
│  Thousand Four Hundred Sixteen Shillings)  │
├────────────────────────────────────────────┤
│ BEING PAYMENT FOR:                         │
│                                            │
│ Invoice No: INV-2025-001                   │
│ Invoice Date: February 10, 2025            │
│ Invoice Amount: 4,172,056.00               │
│ Previous Payments: 2,614,640.00            │
│ This Payment: 1,557,416.00                 │
│ Balance: 0.00                              │
│                                            │
│ Status: PAID IN FULL ✓                     │
├────────────────────────────────────────────┤
│ PAYMENT DETAILS:                           │
│                                            │
│ Payment Method: Bank Transfer              │
│ Bank: Equity Bank                          │
│ Reference: TRX/2025/54321                  │
│ Date: March 15, 2025                       │
├────────────────────────────────────────────┤
│ ACCOUNT SUMMARY:                           │
│                                            │
│ Previous Balance: 2,807,416.00             │
│ Payment Received: (1,557,416.00)           │
│ Current Balance: 1,250,000.00              │
├────────────────────────────────────────────┤
│ This is a computer-generated receipt       │
│                                            │
│ For: ABC Manufacturing Ltd                 │
│                                            │
│ Received by: _______________               │
│ Signature: _______________                │
│ Date: _______________                      │
│                                            │
│ [Company Stamp]                            │
│                                            │
│ Thank you for your payment!                │
└────────────────────────────────────────────┘
```

---

## Returns & Credit Management

### Sales Return Process

```markdown
RETURN AUTHORIZATION WORKFLOW

Return Request:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Customer: ABC Manufacturing Ltd
Contact: John Kamau
Date: 2025-02-25

Return Request Details:
  Original Invoice: INV-2025-001
  Invoice Date: 2025-02-10
  Items to Return: Machine Model B (MDL-B)
  Quantity: 1 unit
  Reason: Technical issue / damaged
  Serial Number: SN12346

Return Reason Categories:
  ○ Defective product
  ● Technical issue
  ○ Wrong item shipped
  ○ Damaged in transit
  ○ Customer changed mind
  ○ Product not as described

Return Policy Check:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Policy: 30-day return window
Invoice Date: Feb 10
Return Request: Feb 25 (15 days) ✓ WITHIN POLICY

Conditions Check:
  □ Return within 30 days ✓
  □ Product in original condition (TBD on inspection)
  □ All accessories included
  □ Proof of purchase ✓ (invoice)
  
Return Authorization:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Return Authorization No: RMA-2025-001
Date Issued: 2025-02-25
Valid Until: 2025-03-10 (14 days)
Authorized By: Sales Manager

Instructions to Customer:
  1. Pack item securely in original packaging
  2. Include all accessories and documentation
  3. Attach RMA number to package
  4. Deliver to: Main Warehouse, Gate B
  5. Operating hours: Mon-Fri, 8 AM - 5 PM

Return Receipt:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Date Received: 2025-02-28
Received By: Warehouse Supervisor
RMA No: RMA-2025-001

Initial Inspection:
  ☑ Item received
  ☑ RMA number verified
  ☑ Serial number matches
  ☑ Packaging intact
  ☑ All accessories present

Detailed Inspection:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Inspected By: Quality Control Team
Date: 2025-02-29

Inspection Results:
  Item: Machine Model B (SN12346)
  Condition Assessment:
    Physical Damage: None ✓
    Operational Test: Failed ⚠
    Issue Found: Motor malfunction
    Root Cause: Manufacturing defect
    
  Verdict: DEFECTIVE - ACCEPT RETURN ✓
  
  Action:
    ☑ Accept for full refund/replacement
    ○ Reject return (customer fault)
    ○ Partial refund (usage/damage)

Return Approval:
  Approved By: Quality Manager
  Approved Date: 2025-02-29
  Disposition: Replace with new unit

Sales Return Document:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Sales Return No: SR-2025-001
Date: 2025-02-29
Customer: ABC Manufacturing Ltd
RMA: RMA-2025-001

Return Items:
┌────────┬──────────────┬─────┬────────────┬───────────┐
│ Item   │ Description  │ Qty │ Rate       │ Amount    │
├────────┼──────────────┼─────┼────────────┼───────────┤
│ MDL-B  │ Machine B    │  1  │ 1,620,000  │ 1,620,000 │
│        │ Serial:12346 │     │            │           │
│        │ Reason:      │     │            │           │
│        │ Defective    │     │            │           │
└────────┴──────────────┴─────┴────────────┴───────────┘

Return Total:               1,620,000 KES
VAT (16%):                    259,200 KES
───────────────────────────────────────
Total Credit:               1,879,200 KES

Resolution Options:
  ● Replacement with new item
  ○ Store credit
  ○ Refund

Selected: Replacement

Inventory Impact:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Return Entry:
  Dr. Inventory - MDL-B (Defective)   1,600,000
      Cr. Cost of Goods Sold                1,600,000

Description: Return of defective unit SN12346
Warehouse: Quarantine Area
Status: For manufacturer warranty claim

Replacement Shipment:
  Item: Machine Model B (new)
  Serial: SN12890
  Delivery Date: 2025-03-05
  Delivery Note: DN-2025-025

Replacement Entry:
  Dr. Cost of Goods Sold            1,600,000
      Cr. Inventory - MDL-B               1,600,000

Credit Note:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Since replacement (not refund), no credit note issued.

If refund was chosen:
  Credit Note: CN-2025-001
  Amount: 1,879,200 KES
  
  Entry:
  Dr. Sales Returns               1,620,000
  Dr. VAT Payable                   259,200
      Cr. Accounts Receivable           1,879,200
  
  Refund Processing:
  Dr. Accounts Receivable         1,879,200
      Cr. Bank/Cash                     1,879,200
```

### Credit Note Management

```markdown
CREDIT NOTE CREATION

Credit Note Scenarios:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Product Return (Full/Partial)
2. Price Adjustment/Correction
3. Billing Error
4. Damaged Goods
5. Promotional Discount (post-invoice)
6. Service Complaint Resolution
7. Early Payment Discount
8. Volume Rebate

Example: Price Adjustment
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Situation:
  Customer received better price from competitor
  Sales manager approves price match
  Invoice already issued and paid

Original Invoice: INV-2025-020
  Item: MDL-X
  Quantity: 5 units
  Original Price: 500,000 per unit
  Total: 2,500,000 KES
  VAT: 400,000 KES
  Grand Total: 2,900,000 KES
  Status: PAID ✓

Price Adjustment:
  Approved New Price: 450,000 per unit
  Difference: 50,000 per unit
  Total Adjustment: 250,000 KES
  VAT Adjustment: 40,000 KES
  Total Credit: 290,000 KES

Credit Note:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Credit Note No: CN-2025-005
Date: 2025-03-10
Against Invoice: INV-2025-020
Customer: XYZ Corporation
Reason: Price adjustment - competitive match

┌────────────────────────────────────────────┐
│         [COMPANY LOGO]                     │
│                                            │
│        CREDIT NOTE                         │
│                                            │
│ Credit Note No: CN-2025-005                │
│ Date: March 10, 2025                       │
│ Against Invoice: INV-2025-020              │
│ Invoice Date: February 15, 2025            │
├────────────────────────────────────────────┤
│ CUSTOMER:                                  │
│ XYZ Corporation                            │
│ PIN: P999888777B                           │
│ Westlands, Nairobi                         │
├────────────────────────────────────────────┤
│ CREDIT FOR:                                │
│                                            │
│ Item: MDL-X Industrial Machine             │
│ Original Price: 500,000 × 5 = 2,500,000   │
│ Revised Price: 450,000 × 5 = 2,250,000    │
│                                            │
│ Price Adjustment:           (250,000)      │
│ VAT Adjustment (16%):        (40,000)      │
│                         ───────────────    │
│ TOTAL CREDIT:               (290,000)      │
│                                            │
│ Reason: Price match - approved by Sales   │
│         Manager as per customer request    │
├────────────────────────────────────────────┤
│ CREDIT USAGE:                              │
│ ☑ Credit to customer account               │
│ ○ Refund to customer                       │
│ ○ Apply to future invoices                 │
├────────────────────────────────────────────┤
│ Authorized by: James Ndungu               │
│ Sales Manager                              │
│ Date: March 10, 2025                       │
└────────────────────────────────────────────┘

Financial Entry:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Dr. Sales Returns & Allowances    250,000
Dr. VAT Payable                    40,000
    Cr. Accounts Receivable - XYZ      290,000

Description: Credit note CN-2025-005
Reason: Price adjustment
Invoice: INV-2025-020 (already paid)

Customer Account Impact:
  Since invoice was paid:
  Customer now has credit balance: 290,000 KES
  
  Usage Options:
    1. Apply to next invoice
    2. Refund to customer
    3. Keep as advance payment

Credit Note Allocation:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Applied Against:
  Next invoice: INV-2025-045
  Invoice Amount: 1,200,000 KES
  Less: Credit Applied: (290,000) KES
  ─────────────────────────────────
  Balance Due: 910,000 KES

Entry:
  Dr. Accounts Receivable - XYZ    290,000
      Cr. Accounts Receivable - XYZ      290,000
  Description: Credit CN-2025-005 applied to INV-2025-045
```

### Return & Credit Reporting

```markdown
RETURNS & CREDITS ANALYTICS

Return Metrics:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Period: Q1 2025

Total Sales:              50,000,000 KES
Total Returns:             1,250,000 KES
Return Rate:                    2.5%

Returns by Reason:
┌───────────────────────┬──────────┬───────────┐
│ Reason                │ Count    │ Amount    │
├───────────────────────┼──────────┼───────────┤
│ Defective             │    15    │  650,000  │
│ Damaged in transit    │     8    │  320,000  │
│ Wrong item shipped    │     5    │  180,000  │
│ Customer changed mind │     3    │  100,000  │
└───────────────────────┴──────────┴───────────┘

Return Rate by Product:
  MDL-A: 1.2% (acceptable)
  MDL-B: 4.5% (investigate) ⚠
  MDL-C: 0.8% (excellent)

Action Items:
  ⚠ Investigate MDL-B quality issues
  → Contact manufacturer
  → Review incoming inspection process

Credit Notes Issued:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Total Credit Notes: 28
Total Value: 2,100,000 KES

By Category:
  Returns: 1,250,000 KES (59.5%)
  Price Adjustments: 580,000 KES (27.6%)
  Billing Errors: 180,000 KES (8.6%)
  Discounts: 90,000 KES (4.3%)

Credit Recovery:
  Applied to future sales: 1,680,000 KES (80%)
  Refunded: 420,000 KES (20%)
```

---

## Pricing & Discount Management

### Price List Structure

```markdown
PRICE LIST CONFIGURATION

Price List Types:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Standard Selling Price (default)
2. Corporate Pricing (B2B customers)
3. Wholesale Pricing (distributors)
4. Retail Pricing (walk-in customers)
5. Export Pricing (international, USD)
6. Special Project Pricing
7. Promotional Pricing

Price List Master:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Price List Name: Corporate Pricing
Currency: KES
Enabled: Yes
Valid From: 2025-01-01
Valid To: 2025-12-31

Applicable To:
  ☑ Customer Group: Corporate
  ☑ Territory: All
  ☑ Specific Customers: [Select if needed]

Price List Items:
┌──────────┬─────────────┬────────────┬────────────┐
│ Item     │ Description │ Std Price  │ Corp Price │
├──────────┼─────────────┼────────────┼────────────┤
│ MDL-A    │ Machine A   │ 2,200,000  │ 1,900,000  │
│ MDL-B    │ Machine B   │ 1,800,000  │ 1,620,000  │
│ MDL-C    │ Machine C   │ 1,200,000  │ 1,080,000  │
│ SVC-01   │ Installation│   180,000  │   150,000  │
│ TRN-01   │ Training/day│    60,000  │    50,000  │
└──────────┴─────────────┴────────────┴────────────┘

Discount: 10-15% below standard pricing

Price List Priority:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
When multiple price lists apply:

Priority Order:
  1. Customer-specific pricing (highest)
  2. Customer group pricing
  3. Territory pricing
  4. Standard selling price (default)

Example:
  Customer: ABC Manufacturing Ltd
  Customer Group: Corporate
  Territory: Nairobi

  Price Resolution:
    Check: Customer-specific price? NO
    Check: Corporate pricing? YES ✓
    Use: Corporate Pricing (1,900,000 for MDL-A)

Customer-Specific Pricing:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
For VIP/Key accounts with negotiated prices:

Customer: ABC Manufacturing Ltd
Special Agreement: Dated 2024-12-15
Valid: 2025-01-01 to 2025-12-31

┌──────────┬───────────────┬─────────────┬────────────┐
│ Item     │ Std Price     │ Corp Price  │ ABC Price  │
├──────────┼───────────────┼─────────────┼────────────┤
│ MDL-A    │ 2,200,000     │ 1,900,000   │ 1,850,000  │
│ MDL-B    │ 1,800,000     │ 1,620,000   │ 1,580,000  │
└──────────┴───────────────┴─────────────┴────────────┘

When ABC places order:
  System automatically applies: 1,850,000 (best price)

Price List Import:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Bulk price updates via CSV import:

CSV Format:
item_code, price_list, rate, valid_from, valid_to
MDL-A, Corporate Pricing, 1900000, 2025-01-01, 2025-12-31
MDL-B, Corporate Pricing, 1620000, 2025-01-01, 2025-12-31
...

Import Process:
  1. Upload CSV file
  2. Validate format and data
  3. Preview changes
  4. Confirm import
  5. Prices updated
  6. Audit log entry created
```

### Discount Rules & Management

```markdown
DISCOUNT CONFIGURATION

Discount Types:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Customer Group Discount
2. Volume/Quantity Discount
3. Promotional Discount
4. Early Payment Discount
5. Seasonal Discount
6. Bundle Discount
7. Loyalty Discount
8. Manual/Discretionary Discount

Discount Rule: Volume Discount
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Rule Name: Volume Discount - MDL-A
Item: Machine Model A (MDL-A)
Enabled: Yes
Valid From: 2025-01-01
Valid To: 2025-12-31

Discount Tiers:
┌──────────────┬───────────────┬──────────────┐
│ Quantity     │ Discount %    │ Final Price  │
├──────────────┼───────────────┼──────────────┤
│ 1-4 units    │ 0%            │ 1,900,000    │
│ 5-9 units    │ 5%            │ 1,805,000    │
│ 10-19 units  │ 10%           │ 1,710,000    │
│ 20+ units    │ 15%           │ 1,615,000    │
└──────────────┴───────────────┴──────────────┘

Example Application:
  Order: 7 units of MDL-A
  Base Price: 1,900,000 × 7 = 13,300,000
  Volume Discount (5%): (665,000)
  Net Amount: 12,635,000 KES

Discount Rule: Promotional Campaign
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Campaign: Q1 2025 Equipment Sale
Valid: 2025-01-15 to 2025-03-31
Apply to: All Machine Models

Discount: 8% on all orders above 5M KES
Additional: Free installation (worth 150K)

Conditions:
  ☑ Minimum order: 5,000,000 KES
  ☑ Valid for new orders only
  ☑ Cannot combine with other promotions
  ☑ Territory: Kenya only

Discount Rule: Early Payment
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Terms: 2/10 Net 30
  = 2% discount if paid within 10 days
  = Full amount due in 30 days

Application:
  Invoice Amount: 1,000,000 KES
  Invoice Date: Mar 1
  Due Date: Mar 31
  
  If paid by Mar 11:
    Discount (2%): 20,000 KES
    Payment Required: 980,000 KES
    
  If paid after Mar 11:
    Full Amount: 1,000,000 KES

Discount Approval Matrix:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌────────────────┬──────────────┬─────────────┐
│ Discount %     │ Approver     │ Max Value   │
├────────────────┼──────────────┼─────────────┤
│ 0-10%          │ Sales Person │ Any         │
│ 10-15%         │ Sales Manager│ < 5M        │
│ 15-20%         │ Sales Director│ < 10M      │
│ 20-25%         │ CFO          │ < 20M       │
│ >25%           │ CEO + CFO    │ Any         │
└────────────────┴──────────────┴─────────────┘

Approval Workflow:
  Sales Rep creates quote with 18% discount
  → Triggers approval workflow
  → Routed to Sales Director
  → Email notification sent
  → Sales Director reviews and approves
  → Quote status: APPROVED
  → Sales Rep can proceed

Discount Tracking:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Period: Q1 2025

Total Sales Revenue: 50,000,000 KES
Total Discounts Given: 3,500,000 KES
Average Discount %: 7.0%

Discount Breakdown:
  Volume Discounts: 1,800,000 KES (51.4%)
  Customer Group: 1,200,000 KES (34.3%)
  Promotional: 400,000 KES (11.4%)
  Manual/Other: 100,000 KES (2.9%)

Discount by Sales Person:
┌────────────────┬───────────┬────────────┬────────┐
│ Sales Person   │ Sales     │ Discount   │ Avg %  │
├────────────────┼───────────┼────────────┼────────┤
│ Sarah Johnson  │ 15M       │ 900K       │ 6.0%   │
│ Mike Chen      │ 12M       │ 1,080K     │ 9.0% ⚠│
│ Jane Mwangi    │ 10M       │ 650K       │ 6.5%   │
│ Tom Omondi     │ 8M        │ 560K       │ 7.0%   │
└────────────────┴───────────┴────────────┴────────┘

⚠ Mike Chen's discount rate high - review required

Bundle Pricing:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Bundle: Complete Production Line Package

Items in Bundle:
  - Machine Model A × 2
  - Machine Model B × 1
  - Installation Service
  - 1 Year Maintenance Contract
  - Operator Training (5 days)

Individual Prices:
  MDL-A: 1,900,000 × 2 = 3,800,000
  MDL-B: 1,620,000 × 1 = 1,620,000
  Installation: 150,000
  Maintenance: 500,000
  Training: 50,000 × 5 = 250,000
  ─────────────────────────────
  Total Individual: 6,320,000 KES

Bundle Price: 5,500,000 KES
Savings: 820,000 KES (13%)

Bundle Rule:
  - All items must be purchased together
  - Cannot substitute items
  - Single invoice for all items
  - All items delivered together
```

---

---


### Invoice Items (Continued)

```markdown
Invoice Items:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Based on Delivery DN-2025-001:
┌────────┬──────────────┬─────┬────────────┬────────────┐
│ Item   │ Description  │ Qty │ Rate       │ Amount     │
├────────┼──────────────┼─────┼────────────┼────────────┤
│ MDL-A  │ Machine A    │  1  │ 1,900,000  │ 1,900,000  │
│        │ SN: 12345    │     │            │            │
│        │ Delivered:   │     │            │            │
│        │ Feb 10, 2025 │     │            │            │
├────────┼──────────────┼─────┼────────────┼────────────┤
│ MDL-B  │ Machine B    │  1  │ 1,620,000  │ 1,620,000  │
│        │ SN: 12346    │     │            │            │
├────────┼──────────────┼─────┼────────────┼────────────┤
│ SVC-01 │ Installation │  1  │   150,000  │   150,000  │
│        │ Feb 12, 2025 │     │            │            │
└────────┴──────────────┴─────┴────────────┴────────────┘

Invoice Calculation:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Item Total:           3,670,000 KES
  Discount (2%):          (73,400) KES
  ───────────────────────────────────
  Taxable Amount:       3,596,600 KES
  VAT @ 16%:              575,456 KES
  ───────────────────────────────────
  GRAND TOTAL:          4,172,056 KES

Payment Applied:
  Deposit (Jan 25):     2,614,640 KES (Advance Payment)
  ───────────────────────────────────
  BALANCE DUE:          1,557,416 KES

Payment Terms:
  Due Date: 2025-03-12 (Net 30 Days)
  Early Payment Discount: 2% if paid by Feb 20 (10 days)
  Late Payment: 2% per month after due date

Amount in Words:
  Four Million One Hundred Seventy Two Thousand
  Fifty Six Shillings Only
```

### Financial Posting

```markdown
ACCOUNTING INTEGRATION

On Invoice Submission:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Automatic Journal Entry Created:

Dr. Accounts Receivable - ABC Mfg    4,172,056
    Cr. Sales Revenue - Equipment             3,596,600
    Cr. VAT Payable (16%)                       575,456

Account Details:
  Accounts Receivable: 1200 (Balance Sheet)
  Sales Revenue: 4100 (Income Statement)
  VAT Payable: 2300 (Balance Sheet - Liability)

Cost of Goods Sold:
  (Already posted on Delivery)
  Dr. COGS                         2,800,000
      Cr. Inventory                        2,800,000

Deposit Allocation:
  Original deposit entry (Jan 25):
  Dr. Bank                         2,614,640
      Cr. Customer Deposits                2,614,640

  On Invoice (Feb 10):
  Dr. Customer Deposits            2,614,640
      Cr. Accounts Receivable             2,614,640

Customer Balance:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Before Invoice:
  Outstanding: 1,250,000 KES (other invoices)
  Deposit: -2,614,640 KES (advance)
  Net Balance: -1,364,640 KES (in credit)

After Invoice:
  New Invoice: 4,172,056 KES
  Less Deposit: -2,614,640 KES
  Net Invoice: 1,557,416 KES
  Previous Balance: 1,250,000 KES
  ───────────────────────────────────
  Total Outstanding: 2,807,416 KES
```

### Invoice Types

```markdown
TYPES OF SALES INVOICES

1. Standard Tax Invoice
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Use: Regular B2B sales
Features:
  ✓ VAT chargeable
  ✓ Customer PIN required
  ✓ Full tax compliance
  ✓ Posted to GL immediately

2. Proforma Invoice
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Use: Quotation with more detail, advance payment request
Features:
  - Not a tax invoice
  - Not posted to accounts
  - Used for customs, banking
  - Can convert to tax invoice later
  
Status: Not Paid (informational only)
Note: "This is not a tax invoice"

3. Debit Note
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Use: Additional charges after original invoice
Examples:
  - Freight charges discovered later
  - Additional services rendered
  - Price adjustment (increase)
  
Entry:
  Dr. Accounts Receivable
      Cr. Sales Revenue / Other Income

4. Recurring Invoice
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Use: Subscription, maintenance, rent
Features:
  ✓ Auto-generated on schedule
  ✓ Monthly/Quarterly/Annual
  ✓ Auto-emailed to customer
  ✓ Template-based

Example:
  Maintenance Contract: 50,000 KES/month
  Start: Jan 2025, End: Dec 2025
  Invoice: 1st of each month
  Total: 12 invoices @ 50,000 KES each

5. Advance Invoice (Prepayment)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Use: Payment before delivery
Process:
  1. Customer pays in advance
  2. Create advance payment entry
  3. On delivery, create invoice
  4. Allocate advance against invoice

Entry:
  On Advance Receipt:
  Dr. Bank                    1,000,000
      Cr. Advance from Customer        1,000,000

  On Final Invoice (2,000,000):
  Dr. Accounts Receivable     2,000,000
      Cr. Sales Revenue               2,000,000
  
  Allocate Advance:
  Dr. Advance from Customer   1,000,000
      Cr. Accounts Receivable         1,000,000
  
  Balance Due: 1,000,000 KES

6. Zero-Rated Invoice (Export)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Use: Export sales
Features:
  ✓ VAT @ 0%
  ✓ Export documentation required
  ✓ Foreign currency (usually USD/EUR)
  ✓ Incoterms specified

Tax Rate: 0% (Zero-Rated, not Exempt)
Documents: Invoice, Export Declaration, Bill of Lading

7. Credit Note (Returns/Adjustments)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Use: Sales returns, price adjustments (decrease)
See: Returns & Credit Management section
```

### Invoice Status Lifecycle

```markdown
INVOICE STATUS FLOW

Status Progression:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
DRAFT → SUBMITTED → PAID / OVERDUE / CANCELLED

DRAFT:
  - Invoice being prepared
  - Can be edited freely
  - Not posted to accounts
  - Not sent to customer

SUBMITTED:
  - Invoice finalized and approved
  - Posted to General Ledger
  - Customer balance updated
  - Email sent to customer
  - Payment tracking begins
  - Cannot be edited (must credit note to change)

PAID:
  - Full payment received
  - Payment allocated to invoice
  - Customer balance reduced
  - Accounts cleared

PARTIALLY PAID:
  - Partial payment received
  - Remaining balance tracked
  - Aging continues for balance
  - Further payments expected

OVERDUE:
  - Due date passed
  - No/insufficient payment
  - Aging buckets apply
  - Collection process triggered
  - Late fees may apply

CANCELLED:
  - Invoice voided (before payment)
  - Entries reversed
  - Credit note issued if needed
  - Customer notified

WRITTEN OFF:
  - Deemed uncollectible
  - Bad debt expense recognized
  - Removed from active AR
  - Legal action may follow

Invoice Aging:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌──────────────┬────────────┬────────────┬──────────┐
│ Aging Bucket │ Days       │ Amount     │ Action   │
├──────────────┼────────────┼────────────┼──────────┤
│ Current      │ 0-30       │ 2,500,000  │ Monitor  │
│ 30 Days      │ 31-60      │ 1,200,000  │ Reminder │
│ 60 Days      │ 61-90      │   800,000  │ Call     │
│ 90+ Days     │ >90        │   500,000  │ Escalate │
│              │            │            │          │
│ TOTAL AR     │            │ 5,000,000  │          │
└──────────────┴────────────┴────────────┴──────────┘

Collection Actions by Age:
  0-30 days: Regular monitoring
  31-60 days: Friendly reminder email
  61-90 days: Follow-up call + email
  91-120 days: Management escalation
  121+ days: Legal notice / debt collection
```

### Invoice Modifications & Corrections

```markdown
HANDLING INVOICE CHANGES

Scenario 1: Invoice Not Yet Sent
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Status: DRAFT
Action: Edit directly
Process:
  - Make changes as needed
  - Resubmit when ready
  - No additional documentation needed

Scenario 2: Submitted but Customer Not Paid
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Status: SUBMITTED (no payment)
Cannot Edit: Invoice is locked

Options:
Option A: Cancel & Re-issue
  1. Cancel current invoice
  2. Accounting entries reversed
  3. Create new correct invoice
  4. Notify customer

Option B: Issue Credit Note + New Invoice
  1. Issue credit note for error
  2. Create new correct invoice
  3. Net effect = correct amount

Scenario 3: Partially/Fully Paid
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Status: PAID/PARTIALLY PAID
Cannot Cancel: Payment already applied

Must Use Credit Note:
  - Issue credit note for difference
  - Refund if overpaid
  - Additional invoice if underpaid

Example - Overcharge Correction:
  Original Invoice: 5,000,000 KES (ERROR)
  Correct Amount: 4,500,000 KES
  Paid: 5,000,000 KES

  Action:
  1. Issue Credit Note: 500,000 KES
  2. Process refund to customer OR
  3. Apply credit to future invoices

Common Corrections:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Wrong Price:
  → Credit note for difference

Wrong Quantity:
  → Credit note for excess/shortage

Wrong Customer:
  → Cancel original (if no payment)
  → Issue to correct customer

Missing Discount:
  → Credit note for discount amount

Wrong Tax Rate:
  → Credit note + reissue

Duplicate Invoice:
  → Cancel duplicate
  → Notify customer to ignore
```

---

## 11. Payment Collection & Allocation

### Payment Receipt Process

```markdown
PAYMENT COLLECTION WORKFLOW

Payment Received:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Customer: ABC Manufacturing Ltd
Amount: 1,557,416 KES
Date: 2025-03-01
Method: Bank Transfer
Reference: TRX/ABC/2025/0301

Payment Entry Creation:
  Payment Entry No: PAY-2025-001
  Date: 2025-03-01
  Party: ABC Manufacturing Ltd
  Paid To Account: Company Bank - Equity
  Payment Type: Receive
  Mode of Payment: Bank Transfer
  Reference No: TRX/ABC/2025/0301
  Amount: 1,557,416 KES

Payment Allocation:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Customer Outstanding Invoices:
┌──────────────┬────────────┬───────────┬──────────┐
│ Invoice No   │ Date       │ Amount    │ Due      │
├──────────────┼────────────┼───────────┼──────────┤
│ INV-2024-089 │ Dec 15, 24 │   500,000 │ Overdue  │
│ INV-2025-001 │ Feb 10, 25 │ 1,557,416 │ Current  │
│ INV-2025-015 │ Feb 28, 25 │   750,000 │ Current  │
│              │            │           │          │
│ TOTAL        │            │ 2,807,416 │          │
└──────────────┴────────────┴───────────┴──────────┘

Allocation Strategy (Configurable):
1. FIFO (First-In, First-Out) - Default
   Allocate to oldest invoices first
   
2. Specific Invoice
   Customer specifies which invoice
   
3. Proportional
   Distribute across all invoices proportionally

FIFO Allocation of 1,557,416 KES:
  INV-2024-089: 500,000 KES (PAID ✓)
  INV-2025-001: 1,057,416 KES (PARTIALLY PAID)
  
  Remaining on INV-2025-001: 500,000 KES

Updated Balances:
┌──────────────┬────────────┬──────────┬──────────┐
│ Invoice No   │ Original   │ Paid     │ Balance  │
├──────────────┼────────────┼──────────┼──────────┤
│ INV-2024-089 │   500,000  │ 500,000  │        0 │
│ INV-2025-001 │ 1,557,416  │1,057,416 │  500,000 │
│ INV-2025-015 │   750,000  │       0  │  750,000 │
│              │            │          │          │
│ TOTAL        │ 2,807,416  │1,557,416 │1,250,000 │
└──────────────┴────────────┴──────────┴──────────┘

Accounting Entry:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Dr. Bank - Equity               1,557,416
    Cr. Accounts Receivable - ABC Mfg   1,557,416

Customer Notifications:
  ✓ Payment receipt email sent
  ✓ Outstanding balance updated: 1,250,000 KES
  ✓ Receipt PDF attached
```

### Payment Methods

```markdown
PAYMENT MODE CONFIGURATIONS

Bank Transfer / EFT:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Most Common Method:
  - Customer initiates transfer
  - Provide bank details on invoice
  - Reference: Invoice number
  - Reconcile daily with bank statement
  
Entry:
  Dr. Bank Account
      Cr. Accounts Receivable

Cash Payment:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Use: Retail, small amounts
Process:
  1. Receive cash
  2. Issue receipt
  3. Deposit to bank daily
  
Entry:
  Dr. Cash Account
      Cr. Accounts Receivable
  
  On Bank Deposit:
  Dr. Bank Account
      Cr. Cash Account

Check/Cheque:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Process:
  1. Receive cheque
  2. Record payment (pending clearance)
  3. Deposit to bank
  4. Confirm clearance (3-5 days)
  5. If bounced: Reverse entry, charge fee
  
Entry (On Receipt):
  Dr. Uncleared Cheques
      Cr. Accounts Receivable
  
  On Clearance:
  Dr. Bank Account
      Cr. Uncleared Cheques
  
  If Bounced:
  Dr. Accounts Receivable
  Dr. Bank Charges
      Cr. Uncleared Cheques

Mobile Money (M-Pesa, Airtel Money):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Common in East Africa:
  - Instant confirmation
  - Transaction ID provided
  - Lower transaction costs
  - API integration possible
  
Entry:
  Dr. Mobile Money Account
      Cr. Accounts Receivable
  
  Transfer to Bank:
  Dr. Bank Account
  Dr. Transfer Charges
      Cr. Mobile Money Account

Credit Card / Debit Card:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Requires Payment Gateway:
  - Online payments
  - POS terminal payments
  - Processing fees apply (2-3%)
  
Entry:
  Dr. Card Payment Account      100,000
      Cr. Accounts Receivable          100,000
  
  Dr. Bank Account (net)         97,000
  Dr. Payment Gateway Fees        3,000
      Cr. Card Payment Account        100,000

Letter of Credit (LC):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Use: International/large transactions
Process:
  1. Customer's bank issues LC
  2. Deliver goods
  3. Submit documents to bank
  4. Bank releases payment
  
More complex, typically handled in Finance module

Payment Links / Online Portals:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Customer Self-Service:
  - Login to customer portal
  - View outstanding invoices
  - Pay online (card/bank)
  - Auto-receipt generated
  - Auto-allocation to invoice
```

### Unallocated Payments

```markdown
HANDLING ADVANCE/UNALLOCATED PAYMENTS

Scenario: Payment Without Invoice Reference
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Customer: XYZ Corporation
Amount Received: 2,000,000 KES
Reference: "Advance payment for upcoming order"
No specific invoice mentioned

Action: Record as Unallocated Payment
  Payment Entry: PAY-2025-025
  Amount: 2,000,000 KES
  Status: Unallocated
  
Entry:
  Dr. Bank Account                2,000,000
      Cr. Unallocated Payments - XYZ     2,000,000

When Invoice Created:
  Invoice: INV-2025-050 for 2,500,000 KES
  
Allocate Unallocated Payment:
  1. Open payment PAY-2025-025
  2. Allocate to INV-2025-050
  3. Remaining: 500,000 KES still due
  
Entry:
  Dr. Unallocated Payments - XYZ  2,000,000
      Cr. Accounts Receivable - XYZ      2,000,000

Customer View:
  Invoice: 2,500,000 KES
  Paid: 2,000,000 KES
  Balance: 500,000 KES

Overpayment Scenario:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Invoice: INV-2025-030 for 1,000,000 KES
Payment Received: 1,200,000 KES

Options:
1. Apply Full Amount to Invoice
   Invoice: PAID
   Excess: 200,000 KES as credit balance
   
   Entry:
   Dr. Bank                        1,200,000
       Cr. Accounts Receivable             1,000,000
       Cr. Customer Advance Payments         200,000
   
   Future: Apply 200K to next invoice

2. Reject Overpayment
   Accept only 1,000,000 KES
   Return excess 200,000 KES to customer

3. Partial Refund
   Ask customer preference:
   - Keep as advance?
   - Refund excess?

Customer Credit Balance:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Customer: ABC Manufacturing
Current Status: 500,000 KES in CREDIT

Means:
  - Customer overpaid
  - Or deposit for future orders
  
Next Invoice:
  New Invoice: 800,000 KES
  Apply Credit: -500,000 KES
  Net Due: 300,000 KES
```

### Payment Terms & Discounts

```markdown
PAYMENT TERMS MANAGEMENT

Standard Payment Terms:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Net 7 Days:
  Payment due 7 days from invoice date
  
Net 10 Days:
  Payment due 10 days from invoice date
  
Net 30 Days:
  Payment due 30 days from invoice date
  Most common for B2B
  
Net 60 Days:
  Extended terms for key customers
  
Net 90 Days:
  Long payment cycle (government, large corps)

COD (Cash On Delivery):
  Payment required upon delivery
  Common for new customers
  
CIA (Cash In Advance):
  Full payment before delivery
  High-risk customers
  Custom orders

Early Payment Discounts:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
2/10 Net 30:
  Meaning:
    - 2% discount if paid within 10 days
    - Full amount due in 30 days
  
Example:
  Invoice Amount: 1,000,000 KES
  Invoice Date: Feb 1
  
  If paid by Feb 11 (10 days):
    Discount: 20,000 KES (2%)
    Pay: 980,000 KES
  
  If paid Feb 12-Mar 3:
    No discount
    Pay: 1,000,000 KES

Entry (with discount):
  Dr. Bank                          980,000
  Dr. Sales Discount (Expense)       20,000
      Cr. Accounts Receivable             1,000,000

1/7 Net 30:
  1% discount if paid within 7 days
  Encourages very quick payment

5/10 Net 60:
  5% discount if paid within 10 days
  60 days net
  Aggressive discount for cash flow

Late Payment Penalties:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Penalty Terms:
  "2% per month on overdue amounts"
  
Calculation:
  Invoice: 1,000,000 KES
  Due: Mar 1
  Paid: Apr 15 (45 days late)
  
  Late Fee:
    1 month = 2% = 20,000 KES
    1.5 months = 3% = 30,000 KES
  
  Total Due: 1,030,000 KES

Entry:
  Dr. Bank                        1,030,000
      Cr. Accounts Receivable            1,000,000
      Cr. Interest Income                   30,000

Implementation:
  - Automatic calculation
  - Grace period (usually 5-7 days)
  - Separate invoice or debit note
  - Customer agreement required

Split Payment Terms:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Milestone-Based:
  Large projects with milestones
  
Example - Construction Equipment:
  Total: 10,000,000 KES
  
  Milestone 1: Order Confirmation (30%)
    Due: Immediately
    Amount: 3,000,000 KES
  
  Milestone 2: Delivery (40%)
    Due: On delivery
    Amount: 4,000,000 KES
  
  Milestone 3: Installation Complete (30%)
    Due: 7 days after installation
    Amount: 3,000,000 KES

System tracks each milestone separately.
```

### Payment Reconciliation

```markdown
BANK RECONCILIATION PROCESS

Bank Statement Matching:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Daily Process:
  1. Download bank statement
  2. Import to system (CSV/API)
  3. Match transactions with payments
  4. Identify unmatched items
  5. Investigate discrepancies

Bank Statement:
┌────────────┬─────────────┬────────────┬───────────┐
│ Date       │ Reference   │ Debit      │ Credit    │
├────────────┼─────────────┼────────────┼───────────┤
│ Mar 01     │ TRX/ABC/001 │            │ 1,557,416 │
│ Mar 01     │ CHQ-456     │            │   500,000 │
│ Mar 02     │ MPESA-789   │            │   250,000 │
│ Mar 02     │ Bank Charge │     2,500  │           │
│ Mar 03     │ TRX/DEF/002 │            │ 3,200,000 │
└────────────┴─────────────┴────────────┴───────────┘

System Payment Records:
┌────────────┬──────────────┬───────────┬───────────┐
│ Date       │ Payment No   │ Customer  │ Amount    │
├────────────┼──────────────┼───────────┼───────────┤
│ Mar 01     │ PAY-2025-001 │ ABC Mfg   │ 1,557,416 │
│ Mar 01     │ PAY-2025-002 │ XYZ Ltd   │   500,000 │
│ Mar 02     │ PAY-2025-003 │ DEF Corp  │   250,000 │
│ Mar 03     │ PAY-2025-004 │ GHI Ent   │ 3,200,000 │
└────────────┴──────────────┴───────────┴───────────┘

Matching:
  ✓ TRX/ABC/001 ↔ PAY-2025-001 (Matched)
  ✓ CHQ-456 ↔ PAY-2025-002 (Matched)
  ✓ MPESA-789 ↔ PAY-2025-003 (Matched)
  ✗ Bank Charge (No match - create expense entry)
  ✓ TRX/DEF/002 ↔ PAY-2025-004 (Matched)

Unmatched Items:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Type 1: Payment in System, Not in Bank
  Reason: Pending clearance (cheques)
  Action: Wait for clearance
  
Type 2: Payment in Bank, Not in System
  Reason: Missing payment entry
  Action: Create payment entry
  
Type 3: Amount Mismatch
  Bank: 1,000,000 KES
  System: 1,002,000 KES
  Reason: Bank charges deducted
  Action: Record bank charge (2,000)

Type 4: Unidentified Deposit
  Bank shows deposit, customer unknown
  Action: Hold as suspense, contact bank

Bank Charges:
  Dr. Bank Charges (Expense)        2,500
      Cr. Bank Account                     2,500
```

### Payment Reporting

```markdown
PAYMENT ANALYTICS & REPORTS

Cash Collection Report:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Period: March 2025
┌─────────────────┬───────────┬──────────┬──────────┐
│ Payment Method  │ Count     │ Amount   │ %        │
├─────────────────┼───────────┼──────────┼──────────┤
│ Bank Transfer   │    45     │ 25.5M    │   65%    │
│ M-Pesa          │    120    │  5.2M    │   13%    │
│ Cheque          │    15     │  6.8M    │   17%    │
│ Cash            │    35     │  1.5M    │    4%    │
│ Credit Card     │    10     │  0.4M    │    1%    │
│                 │           │          │          │
│ TOTAL           │    225    │ 39.4M    │  100%    │
└─────────────────┴───────────┴──────────┴──────────┘

Payment Performance:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Average Days to Pay: 28 days
On-Time Payment Rate: 72%
Early Payment (with discount): 18%
Late Payments: 28%

By Customer Segment:
┌─────────────────┬──────────┬──────────┬──────────┐
│ Segment         │ Avg Days │ On-Time  │ Late Fee │
├─────────────────┼──────────┼──────────┼──────────┤
│ Corporate       │    25    │   85%    │   45K    │
│ Wholesale       │    32    │   68%    │  120K    │
│ Retail          │    15    │   90%    │   10K    │
│ Government      │    55    │   45%    │  250K    │
└─────────────────┴──────────┴──────────┴──────────┘

Collection Efficiency:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Opening AR: 45,000,000 KES
Sales: 55,000,000 KES
Collections: 52,000,000 KES
Closing AR: 48,000,000 KES

Collection Rate: 94.5%
DSO (Days Sales Outstanding): 32 days

Target DSO: 30 days
Variance: +2 days ⚠
```

---

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

## 13. Pricing & Discount Management

### Price List Management

```markdown
PRICE LIST STRUCTURE

Price List Hierarchy:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Standard Retail Price (Base)
├─ Wholesale Price (15% discount)
├─ Corporate Price (10% discount)
├─ Distributor Price (25% discount)
├─ Export Price (USD, Euro based)
└─ Promotional Price (temporary)

Price List Master:
┌────────────────────────────────────────────┐
│ Price List Name: Standard Retail Price     │
│ Currency: KES                              │
│ Valid From: 2025-01-01                     │
│ Valid To: 2025-12-31                       │
│ Enabled: Yes                               │
│ Default: Yes (for new customers)           │
│                                            │
│ Applicable To:                             │
│ ☑ All Customers                            │
│ □ Specific Customer Group                  │
│ □ Specific Territory                       │
│ □ Specific Customers                       │
└────────────────────────────────────────────┘

Item Price Definition:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Product: Machine Model A

Price List: Standard Retail Price
  Rate: 2,200,000 KES
  Currency: KES
  UOM: Unit

Price List: Wholesale Price
  Rate: 1,870,000 KES (15% off retail)
  Minimum Quantity: 1
  
Price List: Corporate Price
  Rate: 1,980,000 KES (10% off retail)
  Valid From: 2025-01-01
  Valid To: 2025-06-30

Price List: Volume Discount
  Quantity 1-5: 2,200,000 KES
  Quantity 6-10: 2,090,000 KES (5% off)
  Quantity 11+: 1,980,000 KES (10% off)

Customer-Specific Pricing:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Customer: ABC Manufacturing (VIP Customer)
Special Price: 1,900,000 KES
Reason: Long-term contract pricing
Valid: 2025-01-01 to 2025-12-31
Approved By: Sales Director

Priority Hierarchy:
  1. Customer-specific price (highest priority)
  2. Customer group price list
  3. Territory price list
  4. Standard price list (lowest priority)

Example Resolution:
  Customer: ABC Manufacturing
  Product: Machine Model A
  
  Available Prices:
    - Standard Retail: 2,200,000 KES
    - Corporate Price List: 1,980,000 KES (customer group)
    - Customer-Specific: 1,900,000 KES
  
  System selects: 1,900,000 KES (highest priority)
```

### Dynamic Pricing Rules

```markdown
PRICING RULE ENGINE

Rule Types:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Volume Discount
2. Product Bundle Discount
3. Promotional Discount
4. Early Payment Discount
5. Customer Loyalty Discount
6. Seasonal Discount
7. Clearance/Close-out Pricing

Volume Discount Rule:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Rule Name: Bulk Purchase Discount
Item: Machine Model A
Rule:
  ├─ Quantity 1-5 units: No discount
  ├─ Quantity 6-10 units: 5% discount
  ├─ Quantity 11-20 units: 10% discount
  └─ Quantity 21+ units: 15% discount

Priority: 10 (higher number = higher priority)
Valid From: 2025-01-01
Valid To: 2025-12-31
Applies To: All customers
Can Combine With Other Rules: Yes

Example:
  Base Price: 2,000,000 KES
  Quantity: 12 units
  
  Calculation:
    Rate: 2,000,000 KES
    Volume Discount (10%): -200,000 KES
    Net Rate: 1,800,000 KES
    Total: 1,800,000 × 12 = 21,600,000 KES

Product Bundle Rule:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Rule Name: Machine + Service Bundle
Condition: If buying Machine + Installation + Training
Discount: 10% on total

Products in Bundle:
  - Machine Model A: 2,000,000 KES
  - Installation: 150,000 KES
  - Training: 100,000 KES
  
Without Bundle:
  Total: 2,250,000 KES + VAT = 2,610,000 KES

With Bundle (10% off):
  Subtotal: 2,250,000 KES
  Bundle Discount (10%): -225,000 KES
  Net: 2,025,000 KES + VAT = 2,349,000 KES
  Savings: 261,000 KES

Promotional Discount Rule:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Rule Name: January 2025 Promo
Type: Promotional
Discount: 15% on all products
Valid: 2025-01-15 to 2025-01-31
Applies To: Corporate Customers only
Max Discount: 500,000 KES per order

Promo Code: JAN2025
Customer must enter code at checkout

Mixed Discount Scenario:
  Base Price: 2,000,000 KES
  Customer Group Discount (10%): -200,000 = 1,800,000
  Promotional Discount (15%): -270,000
  
System calculates:
  Option A: Sequential (10% then 15%)
    After 10%: 1,800,000
    After 15%: 1,530,000 KES
  
  Option B: Best discount only (15%)
    2,000,000 - 15% = 1,700,000 KES
  
  Option C: Additive (25% total)
    2,000,000 - 25% = 1,500,000 KES

Configuration setting determines which method to use.

Customer Loyalty Discount:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Rule Name: Platinum Customer Discount
Condition: 
  Customer Tier = Platinum
  (Based on annual purchase > 10M KES)

Discount: 5% on all orders
Can Combine: Yes (with volume discounts)
Cannot Combine: Promotional discounts

Example:
  Customer: ABC Manufacturing (Platinum)
  Item: Machine Model A × 8 units
  Base Price: 2,000,000 KES
  
  Calculation:
    Base: 2,000,000 × 8 = 16,000,000
    Volume Discount (5% for 6-10 units): -800,000
    Subtotal: 15,200,000
    Loyalty Discount (5%): -760,000
    Net: 14,440,000 KES
    Total Savings: 1,560,000 KES (9.75%)

Seasonal Pricing:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Rule Name: Q4 Year-End Sale
Products: Selected items only
Discount: 20%
Valid: 2025-10-01 to 2025-12-31
Reason: Clear old stock before new models

Items on Sale:
  - Machine Model A (2024): 20% off
  - Machine Model B (2024): 25% off
  
Regular price maintained for 2025 models
```

### Discount Approval Matrix

```markdown
DISCOUNT AUTHORIZATION LEVELS

Approval Hierarchy:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Discount Level → Required Approval
┌──────────────┬────────────────────────────┐
│ 0-10%        │ Sales Person (no approval) │
│ 10-15%       │ Sales Manager              │
│ 15-20%       │ Sales Director             │
│ 20-25%       │ CFO                        │
│ 25%+         │ CEO                        │
└──────────────┴────────────────────────────┘

Order Value Consideration:
  Orders < 1M KES: Standard approval matrix
  Orders 1M-5M KES: +1 level approval
  Orders > 5M KES: +2 levels approval

Example Scenarios:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Scenario 1:
  Order Value: 800,000 KES
  Discount: 12%
  Required Approval: Sales Manager ✓

Scenario 2:
  Order Value: 3,500,000 KES
  Discount: 12%
  Required Approval: Sales Director
  (Order > 1M, so +1 level)

Scenario 3:
  Order Value: 8,000,000 KES
  Discount: 18%
  Required Approval: CFO
  (Order > 5M, normally Sales Director, but +2 levels)

Approval Process:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Sales person creates quotation with 12% discount:
  
1. System detects discount > 10%
2. Status: PENDING APPROVAL
3. Notification sent to Sales Manager
4. Sales Manager reviews:
   - Customer history
   - Margin impact
   - Competitive situation
   - Order value
5. Decision:
   ☑ Approve (quotation can proceed)
   □ Reject (sales person notified)
   □ Approve with conditions (e.g., min quantity)

Approval Comments:
  "Approved. Customer is high-volume account
   with good payment history. Margin still
   acceptable at 12% discount."

Audit Trail:
  Requested by: Sarah Johnson (Sales)
  Requested on: 2025-01-20 10:30
  Discount: 12% (439,440 KES)
  Approved by: James Ndungu (Sales Manager)
  Approved on: 2025-01-20 14:15
  Comments: [As above]

Special Approval Rules:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
New Customer Rule:
  Discount > 5% for first order
  Requires: Sales Director approval
  Reason: Establish proper pricing expectations

Loss Leader Rule:
  Discount resulting in negative margin
  Requires: CFO + Sales Director approval
  Justification: Strategic reasons only

Competitor Match:
  Discount > standard but matching competitor
  Attach: Competitor quote
  Approval: Sales Director
  Valid: One-time match only

Blanket Approval:
  Annual contracts with pre-approved discount
  Set: 15% for ABC Manufacturing (all orders)
  No per-order approval needed
  Review: Quarterly
```

### Margin Management

```markdown
GROSS MARGIN TRACKING

Margin Calculation:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Selling Price: 2,000,000 KES
Cost (COGS): 1,400,000 KES

Gross Profit: 600,000 KES
Gross Margin %: 30%

Formula:
  Margin % = (Selling Price - Cost) / Selling Price × 100

With Discount:
  Original Price: 2,000,000 KES
  Discount (10%): -200,000 KES
  Net Price: 1,800,000 KES
  Cost: 1,400,000 KES
  
  Gross Profit: 400,000 KES
  Gross Margin %: 22.2%

Margin Alerts:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Company Policy:
  Minimum Margin: 20%
  Target Margin: 35%
  
Alert Levels:
  🟢 Margin > 35%: Excellent
  🟡 Margin 25-35%: Good
  🟠 Margin 20-25%: Acceptable (warning)
  🔴 Margin < 20%: Requires approval

Real-time Calculation:
  Sales person enters discount
  System immediately shows:
    - New margin %
    - Margin impact in KES
    - Alert if below threshold
    - Approval required if < 20%

Quotation Screen:
┌────────────────────────────────────────────┐
│ Item: Machine Model A                      │
│ Quantity: 1                                │
│                                            │
│ Base Price: 2,000,000 KES                  │
│ Cost: 1,400,000 KES                        │
│                                            │
│ Customer Discount: 10%                     │
│ Additional Discount: 5%  [WARNING! 🟠]     │
│                                            │
│ Net Price: 1,710,000 KES                   │
│                                            │
│ Gross Profit: 310,000 KES                  │
│ Gross Margin: 18.1% [BELOW MINIMUM!]      │
│                                            │
│ ⚠ This margin requires Sales Director      │
│   approval before submission.              │
│                                            │
│ Justification: _____________________      │
└────────────────────────────────────────────┘

Margin Reporting:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Sales Performance by Margin:
┌──────────────┬───────────┬──────────┬────────┐
│ Sales Person │ Sales     │ Margin   │ Avg %  │
├──────────────┼───────────┼──────────┼────────┤
│ Sarah        │ 45.2M     │ 15.8M    │  35%   │
│ Mike         │ 38.5M     │ 11.6M    │  30%   │
│ Jane         │ 52.1M     │ 13.0M    │  25%   │
│ Tom          │ 29.8M     │  5.4M    │  18% ⚠ │
└──────────────┴───────────┴──────────┴────────┘

⚠ Tom consistently below target margin
   Action: Review discount practices

Product Profitability:
┌──────────────┬─────────┬────────┬─────────┐
│ Product      │ Volume  │ Margin │ Avg %   │
├──────────────┼─────────┼────────┼─────────┤
│ Machine A    │ 120     │ 45M    │  35%    │
│ Machine B    │  85     │ 28M    │  32%    │
│ Machine C    │  50     │  8M    │  18% ⚠  │
│ Installation │ 180     │ 12M    │  40%    │
└──────────────┴─────────┴────────┴─────────┘

⚠ Machine C low margin - review pricing or cost
```

---

## 14. Sales Analytics & Reporting

### Key Sales Metrics

```markdown
SALES KPIs & DASHBOARDS

Executive Sales Dashboard:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Current Month (January 2025):
┌──────────────────────────────────────────┐
│ MTD SALES                                │
│ 45.2M KES                                │
│ Target: 50M | 90.4% achieved  🟡         │
│ YoY Growth: +18%                         │
└──────────────────────────────────────────┘

┌──────────────────────────────────────────┐
│ PIPELINE VALUE                           │
│ 98M KES                                  │
│ Weighted: 52M | Deals: 43                │
│ Expected Close (Q1): 63M                 │
└──────────────────────────────────────────┘

┌──────────────────────────────────────────┐
│ CONVERSION RATES                         │
│ Lead → Opportunity: 45%                  │
│ Quote → Order: 50%                       │
│ Overall Win Rate: 22.5%                  │
└──────────────────────────────────────────┘

┌──────────────────────────────────────────┐
│ AVERAGE METRICS                          │
│ Deal Size: 2.3M KES                      │
│ Sales Cycle: 28 days                     │
│ Gross Margin: 32%                        │
└──────────────────────────────────────────┘

Sales Trend Analysis:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Monthly Sales (Last 12 Months):
┌────────┬──────────┬──────────┬──────────┐
│ Month  │ 2024     │ 2025     │ Growth   │
├────────┼──────────┼──────────┼──────────┤
│ Jan    │ 38.5M    │ 45.2M    │  +17%    │
│ Feb    │ 42.1M    │ 48.5M *  │  +15%    │
│ Mar    │ 51.2M    │ 55.0M *  │  +7%     │
│ Q1     │ 131.8M   │ 148.7M * │  +13%    │
└────────┴──────────┴──────────┴──────────┘
* Forecast

Quarterly Comparison:
  Q1 2024: 131.8M
  Q2 2024: 145.2M
  Q3 2024: 138.5M
  Q4 2024: 165.3M
  Total 2024: 580.8M
  
  Q1 2025 (Forecast): 148.7M
  Annual Target 2025: 650M
  On Track: Yes ✓

Sales by Product Category:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌────────────────┬──────────┬──────┬──────────┐
│ Category       │ Revenue  │ %    │ Units    │
├────────────────┼──────────┼──────┼──────────┤
│ Machinery      │ 285M     │  49% │   145    │
│ Equipment      │ 175M     │  30% │   380    │
│ Spare Parts    │  82M     │  14% │  2,450   │
│ Services       │  39M     │   7% │   520    │
│                │          │      │          │
│ TOTAL          │ 581M     │ 100% │  3,495   │
└────────────────┴──────────┴──────┴──────────┘

Top 10 Products:
┌─────────────────┬──────────┬──────┬──────────┐
│ Product         │ Revenue  │ Units│ Margin % │
├─────────────────┼──────────┼──────┼──────────┤
│ Machine Model A │ 125M     │  58  │   35%    │
│ Machine Model B │  95M     │  52  │   32%    │
│ Equipment X     │  68M     │ 120  │   28%    │
│ Machine Model C │  55M     │  35  │   30%    │
│ Installation    │  32M     │ 180  │   40%    │
└─────────────────┴──────────┴──────┴──────────┘

Sales by Territory:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌────────────┬──────────┬────────┬────────────┐
│ Territory  │ Revenue  │ Target │ Achievement│
├────────────┼──────────┼────────┼────────────┤
│ Nairobi    │ 285M     │ 280M   │  102%  ✓   │
│ Mombasa    │ 145M     │ 160M   │   91%  🟡  │
│ Kisumu     │  82M     │  70M   │  117%  ✓   │
│ Nakuru     │  45M     │  40M   │  113%  ✓   │
│ Other      │  24M     │  30M   │   80%  🔴  │
│            │          │        │            │
│ TOTAL      │ 581M     │ 580M   │  100%  ✓   │
└────────────┴──────────┴────────┴────────────┘

Sales by Customer Type:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌────────────────┬──────────┬──────┬─────────┐
│ Customer Type  │ Revenue  │ %    │ Margin  │
├────────────────┼──────────┼──────┼─────────┤
│ Corporate      │ 320M     │  55% │   32%   │
│ Wholesale      │ 180M     │  31% │   25%   │
│ Retail         │  65M     │  11% │   38%   │
│ Government     │  16M     │   3% │   20%   │
│                │          │      │         │
│ TOTAL          │ 581M     │ 100% │   30%   │
└────────────────┴──────────┴──────┴─────────┘
```

### Sales Team Performance

```markdown
SALES PERSON METRICS

Individual Performance:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Sarah Johnson (Nairobi Territory):
┌────────────────────────────────────────────┐
│ QUOTA ATTAINMENT                           │
│ YTD Sales: 45.2M                           │
│ Quota: 50M                                 │
│ Achievement: 90.4%  🟡                     │
│ Rank: 2 of 15                              │
└────────────────────────────────────────────┘

┌────────────────────────────────────────────┐
│ ACTIVITY METRICS                           │
│ Leads Contacted: 85                        │
│ Opportunities Created: 38                  │
│ Quotes Sent: 52                            │
│ Orders Won: 28                             │
│ Average Deal Size: 1.6M                    │
└────────────────────────────────────────────┘

┌────────────────────────────────────────────┐
│ CONVERSION RATES                           │
│ Lead → Opportunity: 45% (vs 40% avg)  ✓    │
│ Opportunity → Quote: 137% (vs 125% avg) ✓  │
│ Quote → Order: 54% (vs 50% avg)  ✓         │
│ Win Rate: 74% (vs 68% avg)  ✓              │
└────────────────────────────────────────────┘

┌────────────────────────────────────────────┐
│ QUALITY METRICS                            │
│ Average Margin: 35% (vs 30% target)  ✓     │
│ Average Discount: 8% (vs 10% avg)  ✓       │
│ Deal Velocity: 25 days (vs 28 avg)  ✓      │
│ Customer Satisfaction: 4.6/5  ✓            │
└────────────────────────────────────────────┘

Team Leaderboard:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌──────┬──────────────┬──────────┬────────────┐
│ Rank │ Sales Person │ Revenue  │ vs Quota   │
├──────┼──────────────┼──────────┼────────────┤
│  1   │ Mike Chen    │ 52.1M    │  104%  🥇  │
│  2   │ Sarah J.     │ 45.2M    │   90%  🥈  │
│  3   │ Jane Mwangi  │ 38.5M    │   77%  🥉  │
│  4   │ Tom Omondi   │ 29.8M    │   60%      │
│  5   │ Lucy Wanjiru │ 25.4M    │   51%      │
└──────┴──────────────┴──────────┴────────────┘

Pipeline Health by Sales Person:
┌──────────────┬──────────┬──────────┬─────────┐
│ Sales Person │ Pipeline │ Weighted │ Quality │
├──────────────┼──────────┼──────────┼─────────┤
│ Sarah J.     │ 35M      │ 18M      │  Good   │
│ Mike Chen    │ 28M      │ 15M      │  Good   │
│ Jane Mwangi  │ 20M      │ 10M      │  Fair   │
│ Tom Omondi   │ 15M      │  6M      │  Poor ⚠ │
└──────────────┴──────────┴──────────┴─────────┘

⚠ Tom's pipeline is thin - needs more lead generation
```

### Customer Analytics

```markdown
CUSTOMER INSIGHTS

Customer Segmentation:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
RFM Analysis (Recency, Frequency, Monetary):

Top Tier (Platinum):
  - Purchased in last 30 days
  - 10+ orders per year
  - Annual value > 10M KES
  - Count: 15 customers
  - Revenue: 180M (31%)

Mid Tier (Gold):
  - Purchased in last 60 days
  - 5-10 orders per year
  - Annual value 5-10M KES
  - Count: 45 customers
  - Revenue: 285M (49%)

Regular (Silver):
  - Purchased in last 90 days
  - 2-5 orders per year
  - Annual value 1-5M KES
  - Count: 120 customers
  - Revenue: 95M (16%)

At Risk:
  - Last purchase > 90 days
  - Declining order frequency
  - Count: 35 customers
  - Revenue: 21M (4%)
  Action: Re-engagement campaign

Top 20 Customers:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌────┬───────────────┬──────────┬────────┬─────┐
│ #  │ Customer      │ Revenue  │ Orders │ LTV │
├────┼───────────────┼──────────┼────────┼─────┤
│ 1  │ ABC Mfg       │ 25.5M    │  18    │ 45M │
│ 2  │ XYZ Corp      │ 22.1M    │  15    │ 38M │
│ 3  │ DEF Ent       │ 18.8M    │  22    │ 35M │
│ 4  │ GHI Ltd       │ 16.2M    │  12    │ 28M │
│ 5  │ JKL Inc       │ 14.5M    │  14    │ 25M │
└────┴───────────────┴──────────┴────────┴─────┘

Top 20 = 215M revenue (37% of total)
Strategy: Dedicated account management

Customer Lifetime Value (LTV):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Average Customer:
  First Purchase: 1.5M KES
  Average Order: 1.8M KES
  Orders per Year: 4.5
  Customer Lifespan: 3.5 years
  
  LTV = 1.8M × 4.5 × 3.5 = 28.4M KES
  
Customer Acquisition Cost (CAC):
  Marketing Spend: 15M KES/year
  Sales Team Cost: 45M KES/year
  New Customers: 85/year
  
  CAC = 60M / 85 = 706K KES
  
LTV:CAC Ratio = 28.4M / 706K = 40:1 (Excellent!)

Customer Churn Analysis:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Customers Lost (2024):
  Total: 25 customers
  Churn Rate: 8.5%
  Lost Revenue: 18M KES

Churn Reasons:
  - Price (45%): Moved to cheaper competitor
  - Service (25%): Poor support/delivery issues
  - Business Closure (15%): Customer went out of business
  - No Longer Need (10%): Changed business model
  - Other (5%)

At-Risk Customers (Early Warning):
  - No purchase in 90+ days: 35 customers
  - Declining order frequency: 22 customers
  - Increased complaints: 8 customers
  - Payment delays: 12 customers

Retention Actions:
  ✓ Personal outreach by account manager
  ✓ Special retention offers
  ✓ Service quality review
  ✓ Payment plan offers

Customer Satisfaction:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
NPS (Net Promoter Score): +45
  Promoters (9-10): 58%
  Passives (7-8): 32%
  Detractors (0-6): 10%
  
Satisfaction by Touchpoint:
  Sales Process: 4.5/5
  Product Quality: 4.3/5
  Delivery: 4.1/5 ⚠
  Customer Service: 4.6/5
  Value for Money: 4.0/5 ⚠

⚠ Focus areas: Delivery timeliness, Pricing perception
```

### Standard Sales Reports

```markdown
COMMON SALES REPORTS

1. Sales Register
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Period: January 2025
┌──────────┬────────────┬──────────┬──────────┐
│ Date     │ Invoice    │ Customer │ Amount   │
├──────────┼────────────┼──────────┼──────────┤
│ Jan 10   │ INV-2025-1 │ ABC Mfg  │ 4,255K   │
│ Jan 15   │ INV-2025-2 │ XYZ Corp │ 2,180K   │
│ Jan 20   │ INV-2025-3 │ DEF Ent  │ 3,650K   │
│ ...      │ ...        │ ...      │ ...      │
│          │            │ TOTAL    │ 45,200K  │
└──────────┴────────────┴──────────┴──────────┘

2. Sales by Item Report
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Top Selling Items - January 2025:
┌───────────────┬──────┬────────┬──────────┐
│ Item          │ Qty  │ Rate   │ Revenue  │
├───────────────┼──────┼────────┼──────────┤
│ Machine A     │  12  │ 1.9M   │ 22.8M    │
│ Machine B     │   8  │ 1.6M   │ 12.8M    │
│ Equipment X   │  25  │ 450K   │ 11.3M    │
│ Installation  │  32  │ 150K   │  4.8M    │
└───────────────┴──────┴────────┴──────────┘

3. Sales by Customer Report
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Top 10 Customers - January 2025:
┌────────────────┬────────┬──────────┬────────┐
│ Customer       │ Orders │ Revenue  │ Margin │
├────────────────┼────────┼──────────┼────────┤
│ ABC Mfg        │   3    │ 8.5M     │  32%   │
│ XYZ Corp       │   2    │ 6.2M     │  30%   │
│ DEF Ent        │   4    │ 5.8M     │  28%   │
└────────────────┴────────┴──────────┴────────┘

4. Sales Person Performance
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
January 2025 Performance:
┌──────────────┬──────────┬──────┬────────┐
│ Sales Person │ Revenue  │ Deals│ Margin │
├──────────────┼──────────┼──────┼────────┤
│ Sarah J.     │ 12.5M    │  8   │  35%   │
│ Mike Chen    │ 10.2M    │  7   │  32%   │
│ Jane Mwangi  │  8.8M    │  6   │  28%   │
└──────────────┴──────────┴──────┴────────┘

5. Pending Orders Report
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Orders Awaiting Delivery:
┌────────────┬───────────┬──────────┬────────┐
│ Order No   │ Customer  │ Value    │ Days   │
├────────────┼───────────┼──────────┼────────┤
│ SO-2025-08 │ ABC Mfg   │ 5.5M     │  12    │
│ SO-2025-15 │ XYZ Corp  │ 2.8M     │   8    │
│ SO-2025-22 │ DEF Ent   │ 4.2M     │  25 ⚠  │
└────────────┴───────────┴──────────┴────────┘
⚠ Order delayed - follow up required

6. Sales vs Target
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Monthly Comparison:
┌──────────┬──────────┬──────────┬───────────┐
│ Month    │ Target   │ Actual   │ Variance  │
├──────────┼──────────┼──────────┼───────────┤
│ Jan 2025 │ 50.0M    │ 45.2M    │ -4.8M (90%)│
│ Dec 2024 │ 55.0M    │ 58.2M    │ +3.2M (106%)│
│ Nov 2024 │ 48.0M    │ 52.5M    │ +4.5M (109%)│
└──────────┴──────────┴──────────┴───────────┘

7. Accounts Receivable Aging
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
As of: January 31, 2025
┌─────────────┬──────────┬──────┬────────────┐
│ Age Bucket  │ Amount   │ %    │ # Invoices │
├─────────────┼──────────┼──────┼────────────┤
│ Current     │ 25.0M    │  50% │     45     │
│ 1-30 days   │ 12.5M    │  25% │     22     │
│ 31-60 days  │  8.0M    │  16% │     15     │
│ 61-90 days  │  3.0M    │   6% │      8     │
│ 90+ days    │  1.5M    │   3% │      5 ⚠   │
│             │          │      │            │
│ TOTAL       │ 50.0M    │ 100% │     95     │
└─────────────┴──────────┴──────┴────────────┘

8. Sales Forecast
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Q1 2025 Forecast:
┌────────────────────┬──────────────────────┐
│ Category           │ Amount               │
├────────────────────┼──────────────────────┤
│ Closed Won         │ 45.2M (actual)       │
│ Commit (80%+ prob) │ 25.5M (expected)     │
│ Best Case (60-80%) │ 18.2M (possible)     │
│ Pipeline (40-60%)  │ 12.8M (potential)    │
│                    │                      │
│ Conservative       │ 70.7M                │
│ Most Likely        │ 88.9M                │
│ Optimistic         │ 101.7M               │
│                    │                      │
│ Q1 Target          │ 150M                 │
│ Gap to Target      │ -61.1M (most likely) │
└────────────────────┴──────────────────────┘

Action: Accelerate pipeline conversion
```

---

## 15. Sales Teams & Territories

### Territory Definition

```markdown
TERRITORY MANAGEMENT

Territory Structure:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Geographic Territories:

Kenya
├─ Nairobi Region
│  ├─ Nairobi CBD
│  ├─ Westlands
│  ├─ Industrial Area
│  └─ Satellite Towns (Ruiru, Thika, Kiambu)
│
├─ Coast Region
│  ├─ Mombasa
│  ├─ Malindi
│  └─ Kilifi
│
├─ Western Region
│  ├─ Kisumu
│  ├─ Eldoret
│  └─ Kakamega
│
├─ Central Region
│  ├─ Nakuru
│  ├─ Nyeri
│  └─ Nanyuki
│
└─ Other Counties

Regional East Africa:
├─ Uganda (Kampala, Entebbe)
├─ Tanzania (Dar es Salaam, Arusha)
├─ Rwanda (Kigali)
└─ South Sudan (Juba)

Territory Master Record:
┌────────────────────────────────────────────┐
│ Territory Name: Nairobi Corporate          │
│ Parent Territory: Nairobi Region           │
│ Territory Manager: Sarah Johnson           │
│                                            │
│ Coverage:                                  │
│ - Nairobi CBD                              │
│ - Westlands                                │
│ - Industrial Area                          │
│                                            │
│ Customer Segments:                         │
│ ☑ Corporate                                │
│ ☑ Large Enterprise                         │
│ □ SME                                      │
│ □ Retail                                   │
│                                            │
│ Annual Target: 200M KES                    │
│ YTD Achievement: 180M (90%)                │
│                                            │
│ Team Size: 4 sales persons                │
│ Active Customers: 85                       │
│ Pipeline Value: 120M KES                   │
└────────────────────────────────────────────┘

Territory Assignment Rules:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Automatic Assignment:
  When new customer created:
    → System checks customer address
    → Matches to territory
    → Assigns default sales person
    → Notifies sales team

Manual Override:
  Sales manager can:
    - Reassign customer to different territory
    - Transfer customers between sales persons
    - Split large accounts (team selling)

Territory Overlap:
  Some customers span multiple territories
  Solution: Primary & Secondary assignment
  
  Example:
    Customer: Nationwide Retail Chain
    Primary: Nairobi (HQ location)
    Secondary: All regions (store locations)
    Revenue split: 60% primary, 40% split

Territory Performance:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Comparative Analysis:
┌──────────────┬──────────┬──────────┬───────────┐
│ Territory    │ Target   │ Actual   │ Achievement│
├──────────────┼──────────┼──────────┼───────────┤
│ Nairobi Corp │ 200M     │ 180M     │   90%     │
│ Nairobi SME  │ 120M     │ 135M     │  113% ✓   │
│ Mombasa      │ 150M     │ 142M     │   95%     │
│ Kisumu       │  80M     │  92M     │  115% ✓   │
│ Nakuru       │  50M     │  48M     │   96%     │
└──────────────┴──────────┴──────────┴───────────┘

Market Penetration:
┌──────────────┬─────────────┬────────┬─────────┐
│ Territory    │ Total Market│ Our    │ Share   │
├──────────────┼─────────────┼────────┼─────────┤
│ Nairobi Corp │ 2,000M      │ 180M   │   9%    │
│ Mombasa      │   800M      │ 142M   │  18% ✓  │
│ Kisumu       │   400M      │  92M   │  23% ✓  │
└──────────────┴─────────────┴────────┴─────────┘

Growth Opportunity:
  Nairobi has low penetration but high potential
  Focus: Increase market share from 9% to 12%
```

### Sales Team Structure

```markdown
SALES TEAM ORGANIZATION

Team Hierarchy:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Chief Sales Officer (CSO)
│
├─ National Sales Director
│  │
│  ├─ Regional Sales Manager - Nairobi
│  │  ├─ Team Lead - Corporate
│  │  │  ├─ Senior Sales Executive
│  │  │  ├─ Sales Executive
│  │  │  └─ Junior Sales Executive
│  │  │
│  │  └─ Team Lead - SME
│  │     ├─ Sales Executive
│  │     └─ Sales Executive
│  │
│  ├─ Regional Sales Manager - Coast
│  │  ├─ Sales Executive (Mombasa)
│  │  └─ Sales Executive (Malindi)
│  │
│  └─ Regional Sales Manager - Western
│     ├─ Sales Executive (Kisumu)
│     └─ Sales Executive (Eldoret)
│
├─ Key Account Manager (Large Accounts)
│  ├─ Strategic Account Exec (Top 10 customers)
│  └─ Strategic Account Exec (Next 20 customers)
│
└─ Inside Sales Manager (Telesales/Online)
   ├─ Inside Sales Rep
   ├─ Inside Sales Rep
   └─ Inside Sales Rep

Team Specialization Models:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Model 1: Geographic (Current)
  Each sales person owns a territory
  Pros: Clear ownership, local expertise
  Cons: May lack product expertise

Model 2: Industry Vertical
  Sales teams by industry:
    - Manufacturing Team
    - Retail/Distribution Team
    - Construction Team
    - Healthcare Team
  Pros: Deep industry knowledge
  Cons: Territory conflicts

Model 3: Product Line
  Sales teams by product:
    - Machinery Team
    - Equipment Team
    - Services Team
  Pros: Product expertise
  Cons: Customer confusion (multiple reps)

Model 4: Customer Size
  Segmented by customer value:
    - Enterprise Team (>10M annual)
    - Mid-Market Team (1-10M)
    - SMB Team (<1M)
  Pros: Appropriate resource allocation
  Cons: Customers may graduate between teams

Hybrid Model (Recommended):
  - Geographic territories (primary)
  - Industry specialists (overlay)
  - Key account managers (strategic)
  - Inside sales (small customers)

Sales Team Master:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌────────────────────────────────────────────┐
│ Team Name: Nairobi Corporate Team          │
│ Team Manager: James Ndungu                 │
│                                            │
│ Members:                                   │
│ 1. Sarah Johnson (Team Lead)               │
│ 2. Mike Chen (Senior SE)                   │
│ 3. Jane Mwangi (SE)                        │
│ 4. Tom Omondi (Junior SE)                  │
│                                            │
│ Territory: Nairobi Corporate               │
│ Customer Segment: Enterprise (Corporate)   │
│                                            │
│ Team Target: 200M KES (2025)               │
│ Individual Targets:                        │
│   Sarah: 60M                               │
│   Mike: 55M                                │
│   Jane: 50M                                │
│   Tom: 35M                                 │
│                                            │
│ Commission Structure: Team-based           │
│ Split: 60% individual, 40% team pool       │
└────────────────────────────────────────────┘

Team Collaboration:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Deal Registration:
  Sales person registers opportunity
  Prevents multiple reps approaching same customer
  First to register gets the deal

Team Selling:
  Large complex deals require team:
    - Account Executive (relationship)
    - Technical Sales Engineer (solution)
    - Sales Manager (negotiations)
  
  Revenue Split:
    Account Exec: 50%
    Technical SE: 30%
    Manager: 20%

Lead Distribution:
  Inbound leads distributed via:
    - Round Robin (equal distribution)
    - Territory Match (geographic)
    - Skill Match (product/industry)
    - Load Balancing (current pipeline)

Handoff Process:
  Lead → SDR qualifies → AE closes
  Inbound → Inside Sales → Field Sales (large deal)
  Trial → Success Team → Renewal Team
```

### Sales Meetings & Cadence

```markdown
SALES RHYTHM & MEETINGS

Daily Huddle (15 minutes):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Time: 8:30 AM
Attendees: Entire sales team
Format: Stand-up

Agenda:
  □ Yesterday's wins (deals closed)
  □ Today's priorities
  □ Blockers/support needed
  □ Quick updates

Example:
  Sarah: "Closed ABC Mfg deal - 5M. Today meeting 
         XYZ Corp for final negotiation. Need pricing 
         approval for 12% discount."
  
  Manager: "Great work! I'll fast-track the approval. 
           Mike, can you join the XYZ meeting for 
           technical support?"

Weekly Sales Meeting (1 hour):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Time: Monday 9:00 AM
Attendees: Regional team + manager

Agenda:
  1. Week in Review (20 min)
     - Revenue vs target
     - Key wins and losses
     - Pipeline review
  
  2. Week Ahead (15 min)
     - Major opportunities closing
     - Customer meetings
     - Priorities
  
  3. Coaching Corner (15 min)
     - Deal strategy discussion
     - Role play scenarios
     - Best practices sharing
  
  4. Announcements (10 min)
     - New products/promotions
     - Policy updates
     - Recognition

Pipeline Review (Bi-weekly, 90 min):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Deep dive on each rep's pipeline:

Review per Sales Person:
  Opportunities > 1M KES
  For each deal:
    - Customer background
    - Opportunity size
    - Stage and probability
    - Competition
    - Next steps
    - Close date
    - Support needed

Manager Actions:
  - Challenge assumptions
  - Provide coaching
  - Commit forecasts
  - Assign resources

Example Review:
  Opportunity: DEF Corporation - 8M
  Stage: Proposal
  Probability: 60%
  Competition: Competitor X
  
  Manager: "What's their main objection?"
  Rep: "Price. We're 10% higher."
  Manager: "Have you quantified TCO benefits?"
  Rep: "Not yet."
  Manager: "Work with product team on ROI 
           analysis. Present that next week."

Monthly Business Review (2 hours):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Attendees: Sales leadership

Agenda:
  1. Monthly Performance
     - Actual vs Target
     - YTD performance
     - Trend analysis
  
  2. Pipeline Health
     - Weighted pipeline
     - Coverage ratio
     - Stage distribution
  
  3. Team Performance
     - Individual rankings
     - Activity metrics
     - Win/loss analysis
  
  4. Customer Analysis
     - Top customers
     - Churn risks
     - Expansion opportunities
  
  5. Next Month Plan
     - Target allocation
     - Focus areas
     - Initiatives

Quarterly Planning (Half-day):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Strategic planning session:
  - Review previous quarter
  - Set next quarter goals
  - Territory realignment (if needed)
  - Training needs
  - Process improvements
  - Market opportunities
```

---

## 16. Commission Management

### Commission Structure

```markdown
SALES COMMISSION FRAMEWORK

Commission Models:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Model 1: Flat Percentage
  Simple: X% of revenue
  
  Example:
    Rate: 5% of sales
    Sales: 10M KES
    Commission: 500K KES

Model 2: Tiered (Progressive)
  Rate increases with achievement
  
  Example:
    0-80% of quota: 3%
    80-100% of quota: 5%
    100-120% of quota: 7%
    120%+ of quota: 10%
  
  Quota: 10M KES
  Actual: 12M KES (120%)
  
  Calculation:
    First 8M (80%): 8M × 3% = 240K
    Next 2M (100%): 2M × 5% = 100K
    Last 2M (120%): 2M × 7% = 140K
    Total: 480K KES

Model 3: Profit-Based
  Commission on margin, not revenue
  Encourages profitable sales
  
  Example:
    Rate: 15% of gross profit
    Sale: 10M KES
    Cost: 7M KES
    Profit: 3M KES
    Commission: 3M × 15% = 450K KES

Model 4: Hybrid (Revenue + Profit)
  Balanced approach
  
  Example:
    Base: 2% of revenue = 200K
    Bonus: 10% of margin = 300K
    Total: 500K KES

Model 5: Team + Individual
  Split between personal and team performance
  
  Example:
    Individual Component (60%):
      Personal sales: 12M
      Rate: 5%
      Amount: 600K × 60% = 360K
    
    Team Component (40%):
      Team sales: 50M
      Rate: 3%
      Share: (12M/50M) of 1.5M
      Amount: 360K × 40% = 144K
    
    Total: 504K KES

Commission Rules:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Eligibility:
  ✓ Invoice submitted (not draft)
  ✓ Payment received (configurable: invoice or payment)
  ✓ No return/cancellation
  ✓ Within employment period

Calculation Basis:
  Option A: On Invoice Submission
    Commission earned when invoice submitted
    Risk: Customer may not pay
  
  Option B: On Payment Receipt (Recommended)
    Commission earned when payment received
    Fair: Rep gets paid when company gets paid

Split Scenarios:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Scenario 1: Mid-Period Rep Change
  Original Rep left midway through deal
  New rep closed the deal
  
  Solution:
    Original Rep: 50% (for initial work)
    New Rep: 50% (for closing)

Scenario 2: Team Selling
  Multiple people involved
  
  Solution:
    Account Executive: 60%
    Sales Engineer: 25%
    Sales Manager: 15%

Scenario 3: Referral
  Existing rep refers customer in another territory
  
  Solution:
    Territory Rep (closes): 80%
    Referring Rep: 20%

Commission Caps & Floors:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Caps (Maximum):
  Why: Prevent windfall from single large deal
  Example: Maximum 2M KES per quarter
  
  If exceeded:
    Option A: Hard cap at 2M
    Option B: Reduced rate above cap
    Option C: Defer excess to next period

Floors (Minimum):
  Guaranteed minimum (Draw)
  Even if no sales, rep gets base
  
  Example:
    Draw: 50K KES/month
    Commission: 45K earned
    Payment: 50K (company covers shortfall)
  
  Recovery:
    Future commissions offset past draws
    Or: Non-recoverable (pure guarantee)

Special Situations:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Product Launch Bonus:
  Extra incentive for new products
  New Product: 8% (vs standard 5%)
  Duration: First 6 months

Strategic Account Bonus:
  Extra for winning key accounts
  Target Account List: 20 companies
  Bonus: Additional 2% if won

Retention Bonus:
  Commission on renewals
  New Sale: 5%
  Renewal: 2%

Deal Size Accelerators:
  Larger deals = higher rate
  <1M: 4%
  1-5M: 5%
  5-10M: 6%
  10M+: 7%
```

### Commission Calculation Process

```markdown
COMMISSION PROCESSING

Monthly Commission Cycle:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Timeline:
  Day 1-31: Sales activity (Jan 2025)
  Day 32-35: Finance validates invoices & payments
  Day 36-38: Commission calculation runs
  Day 39-40: Manager reviews & approves
  Day 41-42: Payroll processing
  Day 43: Commission paid (Feb 12)

Calculation Steps:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Step 1: Gather Eligible Transactions
  Query:
    - Sales Invoices submitted in Jan 2025
    - With payments received (if policy = on payment)
    - Not cancelled or returned
    - Assigned to sales person

Step 2: Calculate Commission per Transaction
  For each invoice:
    Invoice: INV-2025-001
    Amount: 4,172,056 KES (net of discounts)
    Sales Person: Sarah Johnson
    Commission Rate: 5%
    Commission: 208,603 KES

Step 3: Apply Rules
  - Check if team-selling (split commission)
  - Apply tiered rates if applicable
  - Check for special bonuses
  - Apply any caps or clawbacks

Step 4: Aggregate
  Sarah Johnson - January 2025:
  ┌────────────┬───────────┬──────┬───────────┐
  │ Invoice    │ Amount    │ Rate │ Commission│
  ├────────────┼───────────┼──────┼───────────┤
  │ INV-2025-1 │ 4,172,056 │  5%  │  208,603  │
  │ INV-2025-5 │ 2,180,000 │  5%  │  109,000  │
  │ INV-2025-8 │ 3,650,000 │  5%  │  182,500  │
  │ ...        │ ...       │ ...  │  ...      │
  │            │           │      │           │
  │ TOTAL      │12,500,000 │      │  625,000  │
  └────────────┴───────────┴──────┴───────────┘

Step 5: Adjustments
  Gross Commission: 625,000 KES
  Less: Previous advance (draw): 0
  Less: Clawbacks (returns): -25,000
  Add: Bonuses (new product): +50,000
  ───────────────────────────────
  Net Commission: 650,000 KES

Step 6: Approval
  Manager reviews for accuracy
  Checks for disputes
  Approves payment

Step 7: Payment
  Added to payroll
  Paid with monthly salary
  Statement sent to sales person

Commission Statement:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌────────────────────────────────────────────┐
│ COMMISSION STATEMENT                       │
│ Period: January 2025                       │
│ Sales Person: Sarah Johnson                │
│                                            │
│ SALES SUMMARY                              │
│ Total Sales: 12,500,000 KES                │
│ Quota: 10,000,000 KES                      │
│ Achievement: 125% ⭐                        │
│                                            │
│ COMMISSION CALCULATION                     │
│ Base Commission (5%):      625,000         │
│ Tier Bonus (>120%):         62,500         │
│ New Product Bonus:          50,000         │
│ ──────────────────────────────────         │
│ Gross Commission:          737,500         │
│                                            │
│ ADJUSTMENTS                                │
│ Returns (INV-2024-089):    -25,000         │
│ Q4 Recovery:                    0          │
│ ──────────────────────────────────         │
│ Net Commission:            712,500         │
│                                            │
│ PAYMENT                                    │
│ Pay Date: February 12, 2025                │
│ Method: Bank Transfer                      │
│                                            │
│ YEAR-TO-DATE                               │
│ YTD Sales: 12,500,000                      │
│ YTD Commission: 712,500                    │
│ Avg Rate: 5.7%                             │
└────────────────────────────────────────────┘

Dispute Resolution:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Sales Person Reviews Statement:
  - Checks all transactions included
  - Verifies rates applied
  - Confirms adjustments

If Dispute:
  1. Sales person submits dispute form
  2. Manager & Finance review
  3. Investigation (2-3 days)
  4. Resolution:
     - Adjust current payment OR
     - Correct next month
  5. Updated statement issued

Common Disputes:
  - Missing transaction
  - Wrong rate applied
  - Team-selling split incorrect
  - Unrecognized adjustment
```

### Commission Clawbacks

```markdown
COMMISSION RECOVERY SCENARIOS

When Clawback Applies:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Customer Non-Payment
   Invoice: INV-2025-001
   Amount: 5M KES
   Commission Paid: 250K KES
   Status: Customer defaulted (unpaid after 90 days)
   
   Action:
     If policy = "on payment": No clawback (not paid yet)
     If policy = "on invoice": Clawback 250K

2. Sales Return
   Invoice: INV-2025-005
   Amount: 3M KES
   Commission Paid: 150K KES
   Return: Full return approved
   
   Action: Clawback 150K from next commission

3. Partial Return
   Original: 5M KES, Commission: 250K
   Return: 2M KES worth of goods
   Clawback: 2M × 5% = 100K

4. Credit Note Issued
   Invoice: 4M KES, Commission: 200K
   Credit Note: 500K (price adjustment)
   Clawback: 500K × 5% = 25K

5. Deal Cancelled Pre-Delivery
   Order taken, commission paid
   Customer cancels before delivery
   Full clawback of commission

Clawback Processing:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Method 1: Next Commission Offset
  Most common approach
  Deduct from next month's commission
  
  Example:
    Feb Commission Earned: 500K
    Jan Clawback: -100K
    Net Payment: 400K

Method 2: Salary Deduction
  If no future commissions expected
  (e.g., rep resigned)
  Deduct from final salary

Method 3: Direct Repayment
  Rep no longer employed
  Invoice rep for amount
  Legal action if not paid

Time Limits:
  Clawback Period: 6 months
  After 6 months: Company absorbs loss
  Exception: Fraud (no time limit)

Rep Protection:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Maximum Clawback: 50% of monthly commission
  Spreads impact over multiple months
  
  Example:
    Clawback Due: 300K
    Monthly Commission: 400K
    Max Deduction: 200K (50%)
    
    Recovery Schedule:
      Month 1: -200K
      Month 2: -100K (balance)

No Clawback Situations:
  - Customer bankruptcy (not rep's fault)
  - Force majeure events
  - Company-caused delays/issues
  - Returns due to product defects
```

---

## 17. Customer Credit Management

### Credit Policy Framework

```markdown
CREDIT MANAGEMENT SYSTEM

Credit Application Process:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
New Customer - Credit Request:

Step 1: Application
  New customer applies for credit terms
  
  Required Documents:
    □ Business registration certificate
    □ Tax PIN certificate
    □ Bank reference letter
    □ Trade references (3)
    □ Financial statements (2 years)
    □ Ownership structure
    □ Signed credit application form

Step 2: Credit Assessment
  Credit Manager reviews:
    - Business legitimacy
    - Financial health
    - Industry reputation
    - Payment history (credit bureau)
    - Bank statements
    - Trade references

Step 3: Credit Scoring
  ┌────────────────────────────────────┐
  │ CREDIT SCORING MODEL               │
  ├────────────────────────────────────┤
  │ Financial Strength:        30 pts  │
  │ - Revenue size                     │
  │ - Profitability                    │
  │ - Assets                           │
  │                                    │
  │ Payment History:           25 pts  │
  │ - Credit bureau report             │
  │ - Trade references                 │
  │                                    │
  │ Business Stability:        20 pts  │
  │ - Years in business                │
  │ - Industry position                │
  │                                    │
  │ Management Quality:        15 pts  │
  │ - Experience                       │
  │ - Reputation                       │
  │                                    │
  │ Relationship Strength:     10 pts  │
  │ - Referrals                        │
  │ - Network                          │
  │                                    │
  │ TOTAL SCORE:             /100 pts  │
  └────────────────────────────────────┘

  Score Interpretation:
    80-100: Excellent (A)
    65-79: Good (B)
    50-64: Fair (C)
    35-49: Poor (D)
    <35: Decline (F)

Step 4: Credit Limit Determination
  Based on score and financial strength:
  
  Grade A (80-100):
    Max Credit: 3x monthly purchases
    Terms: Net 60 Days
    
  Grade B (65-79):
    Max Credit: 2x monthly purchases
    Terms: Net 30 Days
    
  Grade C (50-64):
    Max Credit: 1x monthly purchases
    Terms: Net 15 Days
    Require: Personal guarantee
    
  Grade D (35-49):
    Cash upfront or 50% deposit
    Terms: Net 7 Days
    
  Grade F (<35):
    Decline credit
    Cash only

Step 5: Approval
  ┌────────────────────────────────────┐
  │ Credit Limit Approval Matrix       │
  ├────────────────────────────────────┤
  │ 0 - 1M KES:    Credit Manager      │
  │ 1M - 5M KES:   Finance Manager     │
  │ 5M - 10M KES:  CFO                 │
  │ 10M+ KES:      CEO + Board         │
  └────────────────────────────────────┘

Step 6: Documentation
  Credit approval letter issued:
    - Approved credit limit
    - Payment terms
    - Conditions
    - Review period (annual)

Example Credit Approval:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌────────────────────────────────────────────┐
│ CREDIT APPROVAL LETTER                     │
│                                            │
│ To: ABC Manufacturing Ltd                  │
│ Date: January 20, 2025                     │
│                                            │
│ We are pleased to approve credit facilities│
│ as follows:                                │
│                                            │
│ Credit Limit: 5,000,000 KES                │
│ Payment Terms: Net 30 Days                 │
│ Credit Grade: B                            │
│                                            │
│ CONDITIONS:                                │
│ - Credit subject to timely payment         │
│ - Limit reviewed annually                  │
│ - Orders exceeding limit require approval  │
│ - Late payment may suspend credit          │
│                                            │
│ Please sign and return acknowledgment.     │
│                                            │
│ Approved by: [Credit Manager]              │
└────────────────────────────────────────────┘
```

### Credit Control

```markdown
ONGOING CREDIT MONITORING

Real-Time Credit Checks:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Every Sales Order:
  System automatically checks:
  
  Customer: ABC Manufacturing Ltd
  Credit Limit: 5,000,000 KES
  
  Current Exposure:
    Outstanding Invoices: 2,500,000
    Pending Orders: 1,000,000
    ──────────────────────────────
    Total Exposure: 3,500,000
  
  Available Credit: 1,500,000 KES
  
  New Order: 2,000,000 KES
  
  ⚠ CREDIT LIMIT EXCEEDED!
  Excess: 500,000 KES
  
  Actions:
    □ Block order (automatic)
    □ Request credit limit increase
    □ Require deposit payment
    □ Get management override

Credit Hold Management:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Reasons for Credit Hold:
  1. Overdue invoices
  2. Credit limit exceeded
  3. Bounced check
  4. Bankruptcy rumor
  5. Payment dispute
  6. Account under review

When on Hold:
  ✗ Cannot create new orders
  ✗ Cannot increase existing orders
  ✓ Can receive payments
  ✓ Can issue credit notes
  ✓ Can cancel orders

Hold Process:
  1. System flags account
  2. All orders blocked
  3. Sales team notified
  4. Customer contacted
  5. Issue resolved
  6. Hold released

Hold Release Conditions:
  - All overdue invoices paid, OR
  - Payment plan agreed, OR
  - Credit limit increased, OR
  - Management override

Credit Limit Increase:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Temporary Increase:
  For specific large order
  
  Request:
    Customer: ABC Manufacturing
    Current Limit: 5M KES
    Request: 8M KES (for Project X)
    Duration: 3 months
  
  Review:
    - Recent payment behavior: Excellent ✓
    - Current utilization: 70% ✓
    - Financial health: Stable ✓
    - Order profitability: Good ✓
  
  Decision: APPROVED (temporary)
    New Limit: 8M KES
    Valid: Feb 1 - Apr 30, 2025
    Auto-revert: May 1 → 5M KES

Permanent Increase:
  Based on relationship growth
  
  Trigger:
    - Consistently hitting limit
    - Perfect payment record (6+ months)
    - Business growth
  
  Review Process:
    1. Annual credit review
    2. Update financial docs
    3. Reassess credit score
    4. Approve new limit
  
  Example:
    Original: 5M KES (Grade B)
    Payment Record: 12 months, never late
    Average Monthly: 4M KES
    
    New Limit: 8M KES (promoted to Grade A)

Aging Analysis:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Customer: ABC Manufacturing Ltd
┌─────────────┬──────────┬──────────────────┐
│ Age Bucket  │ Amount   │ Action           │
├─────────────┼──────────┼──────────────────┤
│ Current     │ 1,500K   │ Monitor          │
│ 1-30 days   │   800K   │ Reminder sent    │
│ 31-60 days  │   200K   │ Follow-up call   │
│ 61-90 days  │        │ -                │
│ 90+ days    │        │ -                │
│             │          │                  │
│ TOTAL       │ 2,500K   │ Status: Good ✓   │
└─────────────┴──────────┴──────────────────┘

Risk Rating:
  ✓ Low Risk: All current, minimal 30-day
  🟡 Medium Risk: Significant 30-60 day aging
  🔴 High Risk: Any 90+ days overdue

Collection Actions by Age:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
0-30 Days (Current):
  - Invoice sent on due date
  - No action needed

31-60 Days (Overdue):
  - Day 31: Automated reminder email
  - Day 38: Follow-up email
  - Day 45: Phone call from accounts

61-90 Days (Seriously Overdue):
  - Day 61: Credit hold applied
  - Day 65: Manager escalation call
  - Day 70: Demand letter sent
  - Day 80: Stop supply

91-120 Days (Critical):
  - Day 91: Final demand
  - Day 100: Legal notice
  - Day 110: Debt collection agency

121+ Days (Default):
  - Legal action
  - Write-off consideration
  - Blacklist customer
```

### Bad Debt Management

```markdown
BAD DEBT PROVISIONING

Provision Calculation:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Aging-Based Provision:
┌─────────────┬──────────┬──────────┬─────────┐
│ Age         │ Balance  │ Provision│ Amount  │
├─────────────┼──────────┼──────────┼─────────┤
│ Current     │ 25.0M    │    1%    │  250K   │
│ 31-60 days  │  8.0M    │    5%    │  400K   │
│ 61-90 days  │  3.0M    │   25%    │  750K   │
│ 91-120 days │  1.5M    │   50%    │  750K   │
│ 120+ days   │  0.8M    │  100%    │  800K   │
│             │          │          │         │
│ TOTAL       │ 38.3M    │  7.8%    │ 2,950K  │
└─────────────┴──────────┴──────────┴─────────┘

Monthly Provision Entry:
  Dr. Bad Debt Expense           2,950K
      Cr. Allowance for Bad Debts      2,950K

Write-Off Process:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
When to Write Off:
  - Customer bankrupt/liquidated
  - Legal action unsuccessful
  - Cost of collection > debt amount
  - Uncollectible after 180+ days

Approval Required:
  <100K: Finance Manager
  100K-500K: CFO
  500K+: Board approval

Write-Off Entry:
  Customer: XYZ Corporation
  Invoice: INV-2024-150
  Amount: 500,000 KES
  Age: 200 days overdue
  
  Entry:
    Dr. Allowance for Bad Debts    500,000
        Cr. Accounts Receivable - XYZ    500,000
  
  Customer account closed
  Blacklisted for future business

Recovery of Written-Off Debt:
  If payment received after write-off:
  
  Dr. Bank                        500,000
      Cr. Bad Debt Recovery (Income)    500,000
```

---

## 18. Multi-Currency Sales

### Foreign Currency Operations

```markdown
MULTI-CURRENCY SALES MANAGEMENT

Currency Setup:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Base Currency: KES (Kenya Shillings)
  - All financial statements in KES
  - Default for local customers

Foreign Currencies:
  - USD (US Dollar)
  - EUR (Euro)
  - GBP (British Pound)
  - CNY (Chinese Yuan)
  - UGX (Uganda Shilling)
  - TZS (Tanzania Shilling)

Exchange Rate Sources:
  1. Central Bank of Kenya (official)
  2. Commercial banks
  3. Forex providers (API)
  4. Manual entry

Example Rate Table:
┌──────────┬─────────────┬──────────┬──────────┐
│ Currency │ Buy Rate    │ Sell Rate│ Date     │
├──────────┼─────────────┼──────────┼──────────┤
│ USD      │ 129.50 KES  │ 130.50   │ Jan 20   │
│ EUR      │ 142.30 KES  │ 143.50   │ Jan 20   │
│ GBP      │ 165.80 KES  │ 167.20   │ Jan 20   │
└──────────┴─────────────┴──────────┴──────────┘

Multi-Currency Price Lists:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Product: Machine Model A

KES Price List (Local Sales):
  Rate: 2,000,000 KES

USD Price List (Export/Expat):
  Rate: 15,000 USD
  
  Conversion Check:
    15,000 USD × 130 = 1,950,000 KES
    (Slightly lower for competitive export pricing)

EUR Price List (European Customers):
  Rate: 14,000 EUR
  
  Conversion Check:
    14,000 EUR × 143 = 2,002,000 KES

Automatic Currency Selection:
  Customer Country → Default Currency
  Kenya → KES
  USA → USD
  Uganda → UGX or USD
  Europe → EUR or USD

Foreign Currency Sales Process:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Export Sale Example:

Quotation:
  Customer: Global Import Co. (USA)
  Currency: USD
  Amount: 50,000 USD
  Exchange Rate (Quote Date): 130.00 KES/USD
  Equivalent: 6,500,000 KES

Sales Order:
  Date: Jan 20, 2025
  Amount: 50,000 USD
  Rate: 130.00 KES/USD (locked)
  
  System Records:
    - Transaction Currency: 50,000 USD
    - Base Currency: 6,500,000 KES
    - Exchange Rate: 130.00
    - Rate Date: Jan 20, 2025

Invoice:
  Date: Feb 10, 2025
  Amount: 50,000 USD
  Rate: 130.00 (from sales order)
  Equivalent: 6,500,000 KES
  
  Accounting Entry:
    Dr. AR - Global Import (USD)     6,500,000 KES
        Cr. Sales Revenue                    6,500,000 KES
  
  System tracks both:
    USD: 50,000 (original currency)
    KES: 6,500,000 (functional currency)

Payment Receipt:
  Date: Mar 5, 2025
  Amount: 50,000 USD received
  Exchange Rate (Payment Date): 132.00 KES/USD
  KES Received: 6,600,000 KES
  
  Accounting Entry:
    Dr. Bank (USD)                   6,600,000 KES
        Cr. AR - Global Import               6,500,000 KES
        Cr. Foreign Exchange Gain               100,000 KES

Foreign Exchange Gain/Loss:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Realized Gain (as above):
  Invoice Rate: 130.00
  Payment Rate: 132.00
  Difference: +2.00 KES per USD
  Gain: 50,000 × 2.00 = 100,000 KES
  
  Income Statement Impact:
    Cr. FX Gain (Other Income): 100,000 KES

Realized Loss Example:
  Invoice Rate: 130.00
  Payment Rate: 128.00
  Difference: -2.00 KES per USD
  Loss: 50,000 × 2.00 = 100,000 KES
  
  Income Statement Impact:
    Dr. FX Loss (Other Expense): 100,000 KES

Unrealized Gain/Loss:
  Month-end revaluation of outstanding AR/AP
  
  Outstanding Invoice: 50,000 USD
  Invoice Rate: 130.00 → 6,500,000 KES
  Month-end Rate: 133.00 → 6,650,000 KES
  Unrealized Gain: 150,000 KES
  
  Entry:
    Dr. AR - Global Import (revaluation)  150,000
        Cr. Unrealized FX Gain                   150,000
  
  Note: Reverses next month or on payment

Currency Risk Management:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Strategy 1: Forward Contracts
  Lock in exchange rate for future payment
  
  Export Invoice: 50,000 USD due in 60 days
  Current Rate: 130.00
  Forward Rate: 130.50 (slight premium)
  
  Action: Buy forward contract
  Benefit: Certainty on KES amount

Strategy 2: Natural Hedging
  Match USD receivables with USD payables
  
  Example:
    USD Receivables: 100,000 USD
    USD Payables (imports): 80,000 USD
    Net Exposure: 20,000 USD only

Strategy 3: Pricing Buffer
  Build in FX buffer in export prices
  
  Cost in KES: 6,000,000
  Target Margin: 20%
  Target KES: 7,200,000
  
  Expected Rate: 130 → 55,385 USD
  Buffer (5%): +2,769 USD
  Quote: 58,000 USD
  
  If rate drops to 125:
    58,000 × 125 = 7,250,000 KES (still profitable)

Strategy 4: Payment Terms
  Reduce exposure period
  
  Standard: Net 60 days
  Export: Payment in advance or LC
  
  Or: Price incentive for early payment
    60 days: 50,000 USD
    Advance: 48,500 USD (3% discount)
```

### Export Sales Specifics

```markdown
EXPORT SALES DOCUMENTATION

Required Documents:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Commercial Invoice
   Standard sales invoice
   Additional fields:
     - Incoterms (FOB/CIF/EXW)
     - Country of origin
     - HS Code (customs classification)
     - Export declaration number

2. Proforma Invoice
   Preliminary invoice for:
     - Customer to arrange payment/LC
     - Customs valuation
     - Import permit application

3. Packing List
   Detailed contents of each package:
     - Box numbers
     - Item descriptions
     - Quantities
     - Weights (gross and net)
     - Dimensions

4. Certificate of Origin
   Certifies goods manufactured in Kenya
   Required for preferential tariffs
   Issued by Kenya Chamber of Commerce

5. Bill of Lading (B/L)
   Shipping document from freight forwarder
   Evidence of shipment
   Required for customs clearance

6. Export Declaration
   Filed with Kenya Revenue Authority
   Required for VAT zero-rating
   EDF (Electronic Declaration Form)

7. Quality Certificate (if required)
   Inspection certificate
   Conformity to standards
   May require pre-shipment inspection

8. Insurance Certificate (if CIF)
   Marine cargo insurance
   Covers goods in transit
   As per Incoterms requirement

Incoterms:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
EXW (Ex Works):
  Customer responsibility: Pick up from our premises
  Our cost: Production only
  Quote: 50,000 USD EXW Nairobi

FOB (Free On Board):
  Our responsibility: Deliver to ship at port
  Includes: Inland transport + export clearance
  Quote: 52,000 USD FOB Mombasa

CIF (Cost, Insurance, Freight):
  Our responsibility: Deliver to destination port
  Includes: FOB + ocean freight + insurance
  Quote: 55,000 USD CIF New York

VAT Treatment:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Export Sales = Zero-Rated (0% VAT)
  Conditions:
    ✓ Goods exported outside Kenya/EAC
    ✓ Evidence of export (B/L, customs docs)
    ✓ Export declaration filed
    ✓ Payment received in forex (usually)

Invoice Format:
  Subtotal: 50,000 USD
  VAT @ 0%: 0 USD (Zero-rated export)
  Total: 50,000 USD

Regional Sales (EAC):
  Sales to Uganda, Tanzania, Rwanda, etc.
  Still zero-rated
  Require: C2 Form (EAC goods movement cert)

Letter of Credit (LC):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Process:
  1. Proforma invoice sent to customer
  2. Customer's bank opens LC
  3. LC sent to our bank (advising bank)
  4. We ship goods
  5. We present documents to bank
  6. Bank verifies documents vs LC terms
  7. Bank pays us (3-5 days)
  8. Customer's bank debits customer

Document Requirements (typical):
  □ Commercial Invoice (3 copies)
  □ Packing List (3 copies)
  □ Bill of Lading (full set, original)
  □ Certificate of Origin (original)
  □ Insurance Certificate (if CIF)
  □ Inspection Certificate (if required)
  □ Any other docs per LC terms

Critical: Documents must match LC terms exactly
  - Customer name spelling
  - Description of goods
  - Quantities
  - Values
  - Shipment dates
  
Any discrepancy = Bank may reject payment
```

---

## 19. Module Integration Points

### Integration with Financial Module

```markdown
FINANCE MODULE INTEGRATION

Real-Time Accounting:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Every transaction in Sales module auto-creates
corresponding entries in Finance module:

Sales Invoice Submission:
  Selling Module Action:
    - Invoice INV-2025-001 submitted
    - Amount: 5,000,000 KES
    - Customer: ABC Manufacturing
  
  Finance Module Auto-Entry:
    Dr. Accounts Receivable - ABC     5,800,000
        Cr. Sales Revenue                     5,000,000
        Cr. VAT Payable                         800,000
  
  GL Accounts (configurable):
    1200 - Accounts Receivable (Asset)
    4100 - Sales Revenue (Income)
    2300 - VAT Payable (Liability)

Payment Receipt:
  Selling Module:
    - Payment PAY-2025-001 received
    - Amount: 5,800,000 KES
    - Method: Bank Transfer
  
  Finance Module:
    Dr. Bank - Equity Bank            5,800,000
        Cr. Accounts Receivable - ABC         5,800,000

Sales Return / Credit Note:
  Selling Module:
    - Credit Note CN-2025-001
    - Return of 1 item worth 1,620,000 + VAT
  
  Finance Module:
    Dr. Sales Returns (Contra-Revenue) 1,620,000
    Dr. VAT Payable                      259,200
        Cr. Accounts Receivable               1,879,200

Delivery of Goods (with Inventory):
  Selling Module:
    - Delivery Note DN-2025-001 submitted
    - Items delivered and signed
  
  Finance Module:
    Dr. Cost of Goods Sold            3,200,000
        Cr. Inventory                         3,200,000

Account Mapping:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Configure GL accounts per:
  - Company
  - Customer group
  - Product category
  - Territory

Example Configuration:
  Product Category: Machinery
    Revenue Account: 4110 - Machinery Sales
    Expense Account: 5110 - Machinery COGS
    
  Product Category: Services
    Revenue Account: 4200 - Service Revenue
    Expense Account: N/A (no COGS)
  
  Customer Group: Export
    Revenue Account: 4300 - Export Sales
    VAT Account: N/A (zero-rated)

Financial Reports Available:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
From Sales Transactions:

1. Revenue Recognition Report
   Sales by period, customer, product
   
2. Accounts Receivable Report
   Outstanding balances, aging
   
3. Cash Collection Report
   Payments received, methods
   
4. VAT Report
   Output VAT summary for filing
   
5. Profit & Loss Impact
   Revenue, returns, COGS, gross profit
```

### Integration with Inventory Module

```markdown
INVENTORY MODULE INTEGRATION

Stock Reservation:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Sales Order Confirmation:
  Selling Module:
    - Order SO-2025-001 confirmed
    - Item: Machine Model A, Qty: 2
  
  Inventory Module:
    Available Stock Before: 5 units
    Reserved for SO-2025-001: 2 units
    ───────────────────────────────
    Available to Sell: 3 units
  
  Stock Status:
    Physical Stock: 5
    Reserved: 2
    Available: 3
    
  Prevents: Overselling

Stock Picking & Delivery:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Delivery Note Creation:
  Selling Module:
    - DN-2025-001 for SO-2025-001
    - Items to deliver: 2 units
  
  Inventory Module:
    Generates Pick List:
      Item: Machine Model A
      Quantity: 2
      Location: Warehouse A, Shelf 12
      Picker: John
  
Delivery Note Submission:
  Selling Module:
    - DN-2025-001 goods delivered
    - Customer signed
  
  Inventory Module:
    Stock Movement Entry:
      Type: Delivery
      From: Warehouse (Stock)
      To: Customer (Out)
      Quantity: -2 units
    
    Physical Stock: 5 → 3 units
    Reserved: -2 (released)
    Available: 3 units
    
    Valuation Entry (FIFO):
      2 units @ cost 1,400,000 each
      COGS: 2,800,000 KES
  
  Finance Module:
    Dr. COGS                         2,800,000
        Cr. Inventory                        2,800,000

Real-Time Stock Visibility:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Quotation Creation:
  Sales Person selects item
  System shows:
    ┌────────────────────────────────────┐
    │ Item: Machine Model A              │
    │ Price: 2,000,000 KES               │
    │                                    │
    │ Stock Status:                      │
    │ Available: 3 units                 │
    │ In Production: 5 units (Feb 25)    │
    │ On Order: 10 units (Mar 5)         │
    │                                    │
    │ Lead Time: 2 weeks                 │
    │ Expected Delivery: Feb 20          │
    └────────────────────────────────────┘

Backorder Management:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Sales Order: 10 units requested
Available: Only 6 units

Options:
  1. Partial Delivery
     Deliver 6 now, 4 later
     
  2. Wait for Full Stock
     Delay delivery until all available
     
  3. Split Order
     Multiple deliveries as stock arrives

System Tracking:
  SO-2025-001:
    Total Ordered: 10 units
    Delivered: 6 units
    Backorder: 4 units
    Expected: When stock arrives
  
  Inventory triggers:
    Auto-purchase if reorder point hit
    Notify sales of expected date

Stock Transfer for Delivery:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Multi-Location Scenario:
  Order placed: Mombasa customer
  Stock location: Nairobi warehouse
  
  Process:
    1. Sales Order: Mombasa customer
    2. Stock Transfer: Nairobi → Mombasa
    3. Delivery: From Mombasa to customer
  
  Inventory tracks:
    - Inter-branch transfer
    - Transit stock
    - Final delivery
```

### Integration with CRM Module

```markdown
CRM MODULE INTEGRATION

Customer Master Sync:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Single Customer Record:
  Shared between CRM & Sales modules
  
  CRM maintains:
    - Contact details
    - Communication history
    - Opportunities
    - Marketing campaigns
  
  Sales maintains:
    - Credit limit
    - Payment terms
    - Price lists
    - Transaction history

Lead-to-Cash Flow:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
CRM → Selling Module:

1. Marketing Campaign (CRM)
   Generate leads from website, events
   
2. Lead Qualification (CRM)
   SDR qualifies, creates opportunity
   
3. Convert to Customer (CRM → Selling)
   Lead converted to customer record
   Customer created in Sales module
   
4. Quotation (Selling)
   Sales person creates quote
   
5. Sales Order (Selling)
   Customer accepts, order created
   
6. Opportunity Closed-Won (CRM)
   Auto-updated from Sales Order

Customer 360° View:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Unified customer profile shows:

From CRM:
  - Contact details
  - Lead source
  - Campaign history
  - Communications (emails, calls)
  - Opportunities
  - Activities

From Sales:
  - Quotations
  - Orders
  - Invoices
  - Payments
  - Outstanding balance
  - Purchase history

From Support (if available):
  - Support tickets
  - Complaints
  - Resolutions
  - Satisfaction scores

Campaign ROI Tracking:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Marketing Campaign: Q1 Industrial Expo
  Cost: 2,000,000 KES
  Leads Generated: 150
  Opportunities: 45
  Closed Won: 12
  
Linked Sales Orders:
  SO-2025-008: 5,000,000 KES
  SO-2025-015: 3,200,000 KES
  SO-2025-022: 4,500,000 KES
  ... (9 more)
  
Total Revenue: 48,000,000 KES
ROI: 2,300% (24x return)

CRM feeds this data back for analysis
```

### Integration with Project Management

```markdown
PROJECT MODULE INTEGRATION

Project-Based Sales:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Complex sales requiring project management:
  - Custom manufacturing
  - Installation projects
  - Implementation services
  - Multi-phase deliveries

Sales Order → Project Creation:
  SO-2025-001: ABC Manufacturing
  Value: 50,000,000 KES
  Scope: Supply and install 10 machines
  
  Auto-creates Project:
    Project: PRJ-2025-001
    Customer: ABC Manufacturing
    Budget: 50,000,000 KES
    
  Project Tasks:
    1. Design & Engineering (2 weeks)
    2. Manufacturing (8 weeks)
    3. Delivery (1 week)
    4. Installation (3 weeks)
    5. Testing & Commissioning (2 weeks)
    6. Training (1 week)

Milestone Billing:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Contract Structure:
  Total: 50,000,000 KES
  
  Milestone 1: Design Complete (15%)
    Amount: 7,500,000 KES
    Due: On approval
    
  Milestone 2: Manufacturing (40%)
    Amount: 20,000,000 KES
    Due: Equipment ready
    
  Milestone 3: Installation (30%)
    Amount: 15,000,000 KES
    Due: Installation complete
    
  Milestone 4: Commissioning (15%)
    Amount: 7,500,000 KES
    Due: System operational

Invoicing Linked to Milestones:
  When Milestone 1 complete:
    Project Manager marks complete
    → Auto-triggers invoice creation
    
  Sales Invoice:
    Against: SO-2025-001
    Description: Milestone 1 - Design
    Amount: 7,500,000 KES
    Terms: Net 15 Days

Project Cost Tracking:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Project tracks all costs:
  - Material costs
  - Labor hours
  - Subcontractor costs
  - Equipment rental
  - Travel & expenses

Real-time profitability:
  Project Budget: 50,000,000 KES
  Costs to Date: 28,000,000 KES
  Remaining Budget: 22,000,000 KES
  Invoiced: 42,500,000 KES
  Profit Margin: 29% (to date)
```

---

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

## 21. Approval Workflows

### Workflow Engine Integration

```markdown
SALES APPROVAL WORKFLOWS

Workflow Types:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Discount Approval
2. Credit Limit Override
3. Pricing Exception
4. Sales Order Confirmation
5. Return Authorization
6. Credit Note Approval
7. Write-Off Approval

Workflow Definition Example:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Workflow: Discount Approval

Trigger Conditions:
  IF discount > 10%
  THEN initiate approval workflow

Approval Levels:
  Level 1 (10-15%): Sales Manager
  Level 2 (15-20%): Sales Director
  Level 3 (20%+): CFO

Routing Rules:
  Document Type: Quotation
  Field: Additional Discount %
  Condition: > 10
  
  Route to:
    Role: Sales Manager
    User: Current territory manager
    Timeout: 24 hours
    Escalate to: Sales Director

Approval Workflow States:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
DRAFT:
  Sales person creating quotation
  Can edit freely

PENDING APPROVAL:
  Discount > threshold
  Submitted for approval
  Sales person cannot edit
  Cannot send to customer

APPROVED:
  Manager approved
  Can proceed to send quote
  Sales person can submit

REJECTED:
  Manager rejected
  Sales person notified
  Must revise or cancel

CANCELLED:
  Sales person cancelled request
  Workflow terminated

Approval Actions:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Approver Options:
  ☑ Approve
    Proceed with request
    
  ☑ Approve with Comments
    Conditional approval
    Example: "Approved for this customer only,
             not to set precedent"
    
  ☑ Reject
    Request denied
    Reason required
    
  ☑ Request More Info
    Need clarification
    Returns to requester
    
  ☑ Delegate
    Forward to another approver
    (When on leave, etc.)

Approval Form:
┌────────────────────────────────────────────┐
│ APPROVAL REQUEST                           │
│                                            │
│ Document: QTN-2025-150                     │
│ Requester: Sarah Johnson                   │
│ Date: Jan 25, 2025                         │
│                                            │
│ Request Type: Discount Approval            │
│                                            │
│ Details:                                   │
│ Customer: ABC Manufacturing Ltd            │
│ Order Value: 5,000,000 KES                 │
│ Standard Discount: 10%                     │
│ Requested Discount: 15%                    │
│ Additional Discount: 5%                    │
│ Amount Impact: 250,000 KES                 │
│                                            │
│ Justification:                             │
│ "Customer is comparing with Competitor X   │
│ who quoted 10% lower. This discount        │
│ maintains our margin at 25% and wins a     │
│ high-value account with good repeat        │
│ potential."                                │
│                                            │
│ Supporting Documents:                      │
│ □ Competitor quote (attached)              │
│ □ Customer email                           │
│                                            │
│ Approver: James Ndungu (Sales Manager)     │
│ Action: ☐ Approve  ☐ Reject  ☐ More Info  │
│ Comments: _______________________________  │
│ __________________________________________ │
└────────────────────────────────────────────┘

Multi-Level Approval:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Scenario: 22% Discount Request
  Value: 10,000,000 KES order

Approval Chain:
  Level 1: Sales Manager (approve <20%)
    Insufficient authority
    Auto-routes to next level
  
  Level 2: Sales Director (approve <25%)
    Has authority
    Reviews and approves
  
  Level 3: CFO
    Not reached (Director approved)

Parallel Approval:
  Some workflows need multiple approvers
  
  Example: Large Custom Order
    Technical Approval: Engineering Manager
    Commercial Approval: Sales Director
    Financial Approval: Finance Manager
    
  All three must approve before proceeding

Serial Approval:
  Approvals happen in sequence
  
  Example: Credit Limit Increase
    Step 1: Sales Manager recommends
    Step 2: Credit Manager reviews risk
    Step 3: CFO approves amount
```

### Specific Workflow Examples

```markdown
DETAILED WORKFLOW SCENARIOS

1. Quotation Discount Workflow:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Event: Sales person adds 12% discount

Auto-Actions:
  1. Status → PENDING APPROVAL
  2. Notification → Sales Manager
  3. Email sent with details
  4. Dashboard alert

Manager Reviews:
  - Customer value & history
  - Competitor situation
  - Margin impact
  - Strategic importance

Decision: APPROVED
  Comments: "Approved. Customer has been
            paying on time for 2 years.
            Good strategic account."

Result:
  - Status → APPROVED
  - Sales person notified
  - Can submit to customer
  - Audit trail logged

2. Credit Limit Override:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Event: Sales order exceeds credit limit

System Check:
  Customer: ABC Manufacturing
  Credit Limit: 5,000,000 KES
  Outstanding: 3,500,000 KES
  New Order: 2,500,000 KES
  Total Exposure: 6,000,000 KES
  
  ⚠ EXCEEDS LIMIT BY: 1,000,000 KES

Workflow Triggered:
  1. Order on hold
  2. Routed to Credit Manager
  3. Sales Manager notified

Credit Manager Review:
  Checks:
    - Payment history: Perfect ✓
    - Current aging: All current ✓
    - Financial health: Stable ✓
    - Order profitability: Good ✓
    - Relationship: 3-year customer ✓

Decision: APPROVED (Temporary Increase)
  New Limit: 6,000,000 KES
  Duration: This order only
  Condition: Next order reverts to 5M

Result:
  - Order proceeds
  - Invoice created
  - Limit tracked
  - Review in 90 days

3. Return Authorization:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Event: Customer requests return

Workflow:
  Step 1: Support Team Assessment
    - Troubleshoot issue
    - Confirm defect/reason
    - Recommend: Approve/Reject

  Step 2: Sales Manager Review
    - Order value: 2,000,000 KES
    - Return reason: Defective
    - Support confirms: Yes
    - Decision: APPROVED

  Step 3: Logistics Coordination
    - Schedule pickup
    - Create RMA
    - Issue credit note authority

  Step 4: Finance Approval (if >1M)
    - Credit note amount: 2,000,000
    - Finance Manager approves
    - GL entries authorized

Automated Notifications:
  → Customer: RMA number & instructions
  → Warehouse: Expect return
  → Finance: Credit note pending

4. Sales Order Confirmation:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
For High-Value Orders (>5M KES):

Event: Sales person creates SO-2025-250
  Value: 8,000,000 KES
  Customer: New customer (first order)

Auto-Workflow Trigger:
  Condition: Value > 5M AND First Order
  
Approval Required:
  Level 1: Sales Manager
    Checks: Customer verified, pricing OK
    Action: APPROVED
  
  Level 2: Credit Manager
    Checks: Credit application complete
    New customer assessment
    Action: APPROVED with conditions
    Condition: 50% deposit required
  
  Level 3: Finance Manager
    Final check on financials
    Action: APPROVED

Result:
  Order confirmed with conditions:
    - 50% deposit before delivery
    - Balance on delivery
  
  Sales person notified
  Customer contacted for deposit

Workflow Tracking:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Audit Trail Example:
┌──────────┬──────────┬────────────────────┐
│ Date     │ User     │ Action             │
├──────────┼──────────┼────────────────────┤
│ Jan 25   │ Sarah J. │ Created QTN-2025-1 │
│ 09:00    │          │ Discount: 15%      │
├──────────┼──────────┼────────────────────┤
│ Jan 25   │ System   │ Routed to Manager  │
│ 09:01    │          │ (Approval required)│
├──────────┼──────────┼────────────────────┤
│ Jan 25   │ James N. │ Reviewed request   │
│ 14:30    │          │ Added comments     │
├──────────┼──────────┼────────────────────┤
│ Jan 25   │ James N. │ APPROVED           │
│ 14:32    │          │ Conditional: This  │
│          │          │ customer only      │
├──────────┼──────────┼────────────────────┤
│ Jan 25   │ System   │ Notified Sarah J.  │
│ 14:33    │          │ Status: Approved   │
├──────────┼──────────┼────────────────────┤
│ Jan 25   │ Sarah J. │ Submitted to       │
│ 15:00    │          │ customer           │
└──────────┴──────────┴────────────────────┘

Complete trail for audit/compliance
```

---

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
