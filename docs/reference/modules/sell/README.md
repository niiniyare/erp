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

