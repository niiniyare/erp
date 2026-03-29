# Awo ERP: Complete System Documentation

**Version:** 1.0  
**Last Updated:** 2025  
**Architecture:** Schema-Driven Multi-Tenant ERP System  
**Technology Stack:** Go, PostgreSQL, Temporal, HTMX, Alpine.js, Tailwind CSS

---

## Table of Contents

1. [Business Module Catalog](#section-1-business-module-catalog)
   - Understanding Module Architecture
   - Module Relationships and Integration
   - Core ERP Modules
   - Industry-Specific Modules
   - Technical Foundation Modules
   - Implementation Approach
2. [Technical Architecture Guide](#section-2-technical-architecture-guide)
   - System Philosophy
   - Technology Foundation
   - Multi-Tenancy Architecture
   - Schema-Driven UI System
   - Backend Architecture
   - Frontend Architecture
   - Data Layer Design
   - Temporal Workflow Integration
   - Key Architectural Decisions
   - Testing Strategy
3. [Implementation Guide](#section-3-implementation-guide)
   - Module Implementation Pattern
   - Complete Example: Accounts Payable Module
   - Common Patterns Library
   - Temporal Workflow Patterns
   - JSON Schema Examples
   - Module Integration Patterns
   - Extension Architecture

---

## Section 1: Business Module Catalog

### Understanding the Module Architecture

Awo ERP is built on a hierarchical structure that organizes business functionality into logical, interconnected units.

**The Hierarchy:**
```
Module
  └─ Service (one or more)
      └─ Resource (one or more)
          └─ Action (one or more)
              └─ Attributes (define behavior, state, outcomes)
```

**What Each Level Means:**

- **Module**: A major business domain (e.g., Financial Management, Supply Chain Management). Modules represent complete areas of business operations and often contain multiple related services.

- **Service**: A specific business capability within a module (e.g., within Financial Management, you have "General Ledger" and "Accounts Payable" as separate services). Services focus on distinct business processes.

- **Resource**: A business entity that the service manages (e.g., within the Accounts Payable service, "Invoice" is a resource). Resources are the nouns—the things you create, read, update, and delete.

- **Action**: An operation you can perform on a resource (e.g., "Process Supplier Invoice," "Schedule Payment"). Actions are the verbs—what you actually do with resources.
  - ** Temporal Activity**: Simple actions that operate within a single service
  - **⚙️ Temporal Workflow**: Complex actions that orchestrate multiple services or require long-running processes with compensation logic

- **Attributes**: Properties that control how an action behaves (e.g., payment terms, approval requirements, validation rules, status transitions). Attributes define the configuration and business rules for each action.

**Example to Illustrate:**

When a purchase manager needs to pay a supplier:
1. They work in the **Financial Management Module**
2. Using the **Accounts Payable Service**
3. They select an **Invoice Resource** (the supplier's bill)
4. They execute the **"Schedule Payment" Action** ⚙️ (Temporal Workflow - coordinates AP, Cash Management, and GL)
5. The action has **Attributes** like: payment method (bank transfer, check), payment date, approval status, and whether to send a remittance advice

### Module Relationships and Integration

Modules don't operate in isolation. They share data and trigger actions in other modules through well-defined integration points.

**Integration Patterns:**

1. **Data Flow Integration**: One module creates or updates data that another module reads
   - Example: When Procurement creates a Purchase Order, it flows to Inventory Management (to expect goods) and Financial Management (to anticipate payment)

2. **Event-Driven Integration**: Actions in one module trigger automatic actions in other modules
   - Example: When a Goods Received Note is processed in Inventory , it automatically triggers invoice matching in Accounts Payable 

3. **Workflow Orchestration**: Complex business processes that span multiple modules are managed as Temporal workflows
   - Example: Order-to-Cash workflow ⚙️ coordinates CRM, Inventory, Shipping, and Financial Management

4. **Master Data Sharing**: Multiple modules reference common master data
   - Example: The "Employee" entity is mastered in HCM but referenced by Project Management (for time tracking), Financial Management (for expense claims), and Manufacturing (for labor costs)

**Key Dependency Concepts:**

- **Upstream Dependencies**: Modules that provide data this module needs
- **Downstream Dependencies**: Modules that consume data this module produces
- **Bidirectional Dependencies**: Modules that both send and receive data
- **Saga Patterns**: Long-running transactions using Temporal workflows with compensation

---

### Core ERP Modules

#### Financial Management

**Purpose**: Manages all financial transactions, reporting, and compliance. This is the central hub of your ERP—all financial data from other modules ultimately flows here.

**Module Dependencies**:
- **Receives data from**: All modules (invoices from Procurement, payroll from HCM, costs from Manufacturing, revenue from CRM)
- **Provides data to**: Business Intelligence (for financial reporting), Project Management (for budgeting and cost tracking)

##### Service: General Ledger

The general ledger is the master financial record, tracking all debits and credits across all accounts.

**Resource: Chart of Accounts**

*Actions:*

1. **Create Account**  (Temporal Activity)
   - Attributes: Account code, account name, account type (asset, liability, equity, revenue, expense), parent account (for hierarchical organization), currency, active/inactive status
   - Process: Validates account code uniqueness, ensures proper hierarchy, assigns default tax treatment
   - Temporal: Simple validation and persistence within General Ledger service

2. **Post Journal Entry** ⚙️ (Temporal Workflow)
   - Attributes: Entry date, journal type (manual, automatic, adjustment), debit accounts with amounts, credit accounts with amounts, description, supporting document references, posting status (draft, posted, reversed)
   - Process: Validates double-entry rules (debits = credits), checks account validity, enforces posting period controls, creates audit trail
   - Outcomes: Updates account balances, creates transaction history, affects trial balance
   - Temporal: Workflow ensures atomicity across multiple account updates, supports compensation if posting fails

3. **Generate Trial Balance**  (Temporal Activity)
   - Attributes: As-of date, account filter, currency, include zero-balance accounts
   - Process: Aggregates all posted entries up to specified date, calculates running balances, validates balance equality
   - Outcomes: Produces report showing all account balances

4. **Close Fiscal Period** ⚙️ (Temporal Workflow)
   - Attributes: Period end date, transfer profit/loss to retained earnings, closing journal entry template
   - Process: Validates all entries are posted, transfers temporary accounts to permanent accounts, locks period from further posting, generates closing entries
   - Outcomes: Period marked as closed, financial statements become final, opening balances set for next period
   - Temporal: Long-running workflow that coordinates validation, transfers, and locking with rollback capabilities

##### Service: Accounts Payable & Receivable

**Resource: Invoice**

*Actions:*

1. **Process Supplier Invoice** ⚙️ (Temporal Workflow)
   - Attributes: Supplier identifier, invoice number, invoice date, due date, line items, payment terms, currency, three-way match requirement
   - Process: Validates supplier exists, matches against purchase order if applicable, verifies goods received, calculates taxes, determines due date, routes for approval if amount exceeds threshold
   - Outcomes: Invoice recorded in AP ledger, liability created in general ledger, payment scheduled
   - Temporal: Coordinates validation, three-way matching, approval workflow, and GL posting with compensation

2. **Schedule Payment** ⚙️ (Temporal Workflow)
   - Attributes: Payment date, payment method, bank account, payment batch identifier, early payment discount, approval required
   - Process: Checks available credit, validates bank account details, calculates discount if paying early, creates payment instruction
   - Outcomes: Payment queued in payment run, cash flow forecast updated
   - Temporal: Orchestrates cash availability check, bank validation, discount calculation, and payment queue management

3. **Send Customer Invoice**  (Temporal Activity)
   - Attributes: Customer identifier, invoice date, payment terms, line items, tax calculations, delivery method, invoice template
   - Process: Generates invoice document, calculates totals and taxes, creates receivable in general ledger, sends via specified channel
   - Outcomes: Customer invoiced, revenue recognized, receivable recorded

4. **Record Payment** ⚙️ (Temporal Workflow)
   - Attributes: Payment amount, payment date, payment reference, payment method, bank account, currency, exchange rate, allocation to invoice(s)
   - Process: Validates payment amount, matches to outstanding invoices, handles partial payments, posts to cash account, clears or reduces receivable/payable
   - Outcomes: Invoice marked as paid, cash position updated, bank reconciliation item created
   - Temporal: Coordinates payment validation, invoice allocation, GL posting, and bank reconciliation

5. **Manage Recurring Invoices** ⚙️ (Temporal Workflow - Scheduled)
   - Attributes: Recurrence pattern, start date, end date, invoice template, automatic send
   - Process: Evaluates schedule, generates invoices automatically on due dates, handles variations, manages subscription renewals
   - Outcomes: Invoices generated automatically
   - Temporal: Long-running scheduled workflow with cron-like triggers

##### Service: Asset Management

**Resource: Fixed Asset**

*Actions:*

1. **Record Asset Acquisition**  (Temporal Activity)
   - Attributes: Asset name, asset category, acquisition date, purchase cost, supplier, expected useful life, depreciation method, salvage value, location
   - Process: Creates asset record, assigns unique asset tag, posts acquisition to asset account, initiates depreciation schedule

2. **Calculate Depreciation** ⚙️ (Temporal Workflow - Scheduled)
   - Attributes: Calculation date, depreciation method, frequency, prorate first year
   - Process: Retrieves asset details, applies depreciation formula, calculates period expense, posts depreciation journal entry
   - Outcomes: Asset book value reduced, depreciation expense recorded
   - Temporal: Scheduled monthly/quarterly workflow that processes all assets in batches

3. **Schedule Maintenance**  (Temporal Activity)
   - Attributes: Maintenance type, frequency, next due date, estimated cost, service provider

4. **Process Asset Disposal** ⚙️ (Temporal Workflow)
   - Attributes: Disposal date, disposal method, proceeds received, removal reason, authorization
   - Process: Calculates final depreciation, determines net book value, records gain or loss, removes asset from register, posts disposal entries
   - Temporal: Coordinates depreciation calculation, GL postings, and asset deactivation with compensation

---

#### Supply Chain Management (SCM)

**Purpose**: Oversees the complete flow of goods, information, and finances from suppliers through warehouses to customers.

##### Service: Procurement

**Resource: Purchase Order (PO)**

*Actions:*

1. **Create PO**  (Temporal Activity)
   - Attributes: Supplier identifier, PO date, delivery date, delivery address, line items, payment terms, currency, priority level

2. **Route for Approval** ⚙️ (Temporal Workflow)
   - Attributes: Approval workflow identifier, current approver, approval threshold amounts, escalation timeframe
   - Process: Determines required approvers based on PO value, sends notification to first approver, tracks approval status, handles rejections
   - Temporal: Durable approval workflow with timeout handling and escalation logic

3. **Send to Supplier**  (Temporal Activity)
   - Attributes: Transmission method, supplier contact, PO document template, acknowledgment required

4. **Track PO Status**  (Temporal Activity)
   - Attributes: Current status, status history, expected vs. actual delivery dates, exceptions

5. **Process GRN (Goods Received Note)** ⚙️ (Temporal Workflow)
   - Attributes: PO reference, receipt date, line items received, quantity received, condition, location, discrepancies, quality inspection status
   - Process: Validates against PO quantities, records actual receipt, flags over/under deliveries, updates inventory, triggers invoice matching
   - Temporal: Coordinates inventory update, three-way match initiation, and quality inspection

##### Service: Inventory Management

**Resource: Stock Item**

*Actions:*

1. **Check Real-time Levels**  (Temporal Activity)
   - Attributes: Item identifier, warehouse/location filter, include reserved quantities, include in-transit quantities

2. **Set Reorder Points**  (Temporal Activity)
   - Attributes: Item identifier, warehouse, reorder level, reorder quantity, lead time, safety stock level

3. **Perform Cycle Count** ⚙️ (Temporal Workflow)
   - Attributes: Count date, warehouse/location, count method, items selected, counter assigned, variance threshold
   - Process: Generates count sheets, assigns to counters, records counted quantities, compares to system quantities, identifies variances, requires investigation for significant discrepancies
   - Temporal: Long-running workflow that coordinates count execution, variance investigation, and adjustment approval

4. **Adjust Stock**  (Temporal Activity - with Saga)
   - Attributes: Item identifier, warehouse, adjustment quantity, adjustment reason, reference document, approval required
   - Process: Validates authorization, updates inventory balance, posts value adjustment to general ledger, creates audit trail
   - Temporal: Activity within a Saga pattern to ensure inventory and GL stay in sync

5. **Track Lot/Serial Numbers**  (Temporal Activity)
   - Attributes: Item identifier, lot number or serial number, expiry date, manufacturing date, supplier lot reference, traceability link

##### Service: Warehouse Management

**Resource: Storage Bin**

*Actions:*

1. **Optimize Putaway**  (Temporal Activity)
   - Attributes: Item characteristics, bin characteristics, putaway strategy

2. **Generate Pick List**  (Temporal Activity)
   - Attributes: Order references, pick strategy, pick priority, picker assigned, pick path optimization

3. **Plan Shipment** ⚙️ (Temporal Workflow)
   - Attributes: Orders to ship, carrier, ship date, vehicle capacity, loading sequence, route optimization
   - Process: Groups orders for efficient shipping, optimizes truck loading, generates shipping documents
   - Temporal: Coordinates order consolidation, route optimization, and carrier integration

4. **Schedule Dock**  (Temporal Activity)
   - Attributes: Dock door identifier, appointment date/time, carrier, shipment type, expected duration

---

#### Human Capital Management (HCM)

##### Service: Payroll

**Resource: Employee Record**

*Actions:*

1. **Calculate Gross Pay**  (Temporal Activity)
   - Attributes: Pay period, base salary or hourly rate, regular hours, overtime hours, overtime multiplier, shift differentials, bonuses, commissions

2. **Process Deductions**  (Temporal Activity)
   - Attributes: Statutory deductions, voluntary deductions, deduction priority order, pre-tax vs. post-tax classification

3. **Generate Payslips**  (Temporal Activity)
   - Attributes: Pay period, employee filter, payslip template, delivery method

4. **File Payroll Taxes** ⚙️ (Temporal Workflow - Scheduled)
   - Attributes: Tax jurisdiction, filing period, tax type, filing deadline, payment method
   - Process: Aggregates employee tax withholdings, calculates employer contributions, prepares regulatory filings, submits to tax authorities
   - Temporal: Scheduled workflow with retry logic and compliance tracking

##### Service: Time & Attendance

**Resource: Timesheet**

*Actions:*

1. **Record Clock-in/out**  (Temporal Activity)
   - Attributes: Employee identifier, timestamp, location, device identifier, clock type

2. **Track Leave**  (Temporal Activity)
   - Attributes: Leave type, start date, end date, partial days, approval status, leave balance impact

3. **Approve Overtime** ⚙️ (Temporal Workflow)
   - Attributes: Employee identifier, date, overtime hours, reason, budget code, authorization level
   - Process: Validates overtime is authorized by policy, checks budget availability, routes to appropriate approver
   - Temporal: Approval workflow with budget validation

4. **Export to Payroll**  (Temporal Activity)
   - Attributes: Pay period, employee filter, export format, integration method, validation rules

##### Service: Recruitment

**Resource: Job Opening**

*Actions:*

1. **Post Vacancy**  (Temporal Activity)
   - Attributes: Job title, department, job description, salary range, posting channels, application deadline

2. **Track Applicants**  (Temporal Activity)
   - Attributes: Applicant name, application date, resume attachment, stage in hiring pipeline, rejection reason, source channel

3. **Schedule Interview** ⚙️ (Temporal Workflow)
   - Attributes: Candidate identifier, interview type, interviewer(s), date/time, location or video link
   - Process: Checks interviewer availability, sends calendar invitations, prepares interview materials, collects feedback
   - Temporal: Coordinates calendar availability, notifications, and feedback collection

4. **Generate Offer Letter**  (Temporal Activity)
   - Attributes: Candidate name, position title, start date, salary, benefits summary, employment terms, approval chain

---

#### Customer Relationship Management (CRM)

##### Service: Sales Automation

**Resource: Lead & Opportunity**

*Actions:*

1. **Qualify Lead**  (Temporal Activity)
   - Attributes: Lead source, contact information, company size, industry, budget authority, need identified, timeline, lead score

2. **Log Customer Interaction**  (Temporal Activity)
   - Attributes: Interaction date/time, interaction type, participants, discussion topics, customer sentiment, next steps

3. **Update Sales Stage**  (Temporal Activity)
   - Attributes: Current stage, new stage, stage entry date, probability of close, expected close date, value adjustment

4. **Forecast Revenue**  (Temporal Activity)
   - Attributes: Forecast period, opportunity stage filters, probability weighting, sales representative filter, confidence level

##### Service: Marketing Automation

**Resource: Campaign**

*Actions:*

1. **Create Email Campaign**  (Temporal Activity)
   - Attributes: Campaign name, objective, target segment, email template, subject line variants, send schedule

2. **Segment Contacts**  (Temporal Activity)
   - Attributes: Segmentation criteria, segment name, dynamic vs. static segment, inclusion/exclusion rules

3. **Track Performance**  (Temporal Activity)
   - Attributes: Campaign identifier, metrics tracked, revenue attributed, cost per lead, ROI

4. **Identify Cross-sell Opportunities**  (Temporal Activity)
   - Attributes: Customer purchase history, product affinity analysis, customer lifecycle stage, recommended products

##### Service: Customer Service

**Resource: Support Case**

*Actions:*

1. **Log Customer Issue**  (Temporal Activity)
   - Attributes: Customer identifier, contact channel, issue category, priority, description, case number

2. **Assign to Agent**  (Temporal Activity)
   - Attributes: Assignment method, agent identifier, agent workload, agent skill set, case priority, SLA deadline

3. **Track Resolution** ⚙️ (Temporal Workflow)
   - Attributes: Case status, resolution notes, time spent, resolution category, knowledge article created
   - Process: Updates case as work progresses, documents steps, validates resolution with customer, closes case
   - Temporal: Tracks SLA timers and escalates if deadlines are missed

4. **Send Customer Feedback Request**  (Temporal Activity)
   - Attributes: Trigger event, survey template, delivery timing, delivery channel, response deadline

---

#### Manufacturing & Production

##### Service: Production Planning

**Resource: Bill of Materials (BOM)**

*Actions:*

1. **Create BOM**  (Temporal Activity)
   - Attributes: Parent item, revision number, effective dates, component line items, component type, substitutions allowed

2. **Schedule Production Order** ⚙️ (Temporal Workflow)
   - Attributes: Product to manufacture, quantity required, priority, requested completion date, routing, work centers assigned
   - Process: Evaluates capacity and material availability, performs backward scheduling, reserves materials, assigns to work centers
   - Temporal: Complex workflow that coordinates capacity planning, material reservation, and work center scheduling

3. **Allocate Raw Materials** ⚙️ (Temporal Workflow)
   - Attributes: Production order reference, component requirements, warehouse/bin locations, lot/serial tracking, backflush method
   - Process: Checks material availability, reserves materials, generates pick lists, handles shortages, issues materials to production
   - Temporal: Coordinates inventory reservation and allocation across multiple warehouses

4. **Track Job Status**  (Temporal Activity)
   - Attributes: Production order identifier, current operation, status by operation, quantity completed, quantity rejected, labor/machine hours

##### Service: Quality Assurance

**Resource: Quality Check**

*Actions:*

1. **Define Quality Standard**  (Temporal Activity)
   - Attributes: Product/process identifier, inspection type, specification parameters, acceptance criteria, test methods

2. **Perform In-process Inspection**  (Temporal Activity)
   - Attributes: Production order reference, operation/work center, inspection time, measurements taken, conformance status

3. **Record Defects**  (Temporal Activity)
   - Attributes: Defect type, defect cause, severity, location found, quantity affected, disposition, cost impact

4. **Generate Certificate of Analysis**  (Temporal Activity)
   - Attributes: Product identifier, lot/batch number, production date, test results summary, compliance statement

##### Service: Maintenance Management

**Resource: Work Order**

*Actions:*

1. **Log Equipment Issue**  (Temporal Activity)
   - Attributes: Equipment identifier, issue date/time, reported by, symptom description, severity, impact on production

2. **Schedule Preventive Maintenance** ⚙️ (Temporal Workflow - Scheduled)
   - Attributes: Equipment identifier, maintenance task, frequency, next due date, estimated duration, parts required
   - Process: Calculates schedule based on last service, creates work order in advance, coordinates with production schedule
   - Temporal: Scheduled workflow that manages preventive maintenance calendar

3. **Assign Technician**  (Temporal Activity)
   - Attributes: Work order identifier, technician identifier, technician skill set, certifications required, availability

4. **Track Repair History**  (Temporal Activity)
   - Attributes: Equipment identifier, maintenance history, failure patterns, MTBF, total cost of ownership

---

### Industry-Specific Modules

#### Restaurant & Hotel Management (Hospitality)

##### Service: Property Management System (PMS) - Hotels

**Resource: Reservation**

*Actions:*

1. **Create and Modify Bookings** ⚙️ (Temporal Workflow)
   - Attributes: Guest name, contact info, arrival/departure dates, number of guests, room type, rate plan, special requests, guarantee method
   - Process: Checks room availability, applies rate rules, confirms rate and terms, creates reservation, sends confirmation, blocks room inventory
   - Temporal: Coordinates availability check, rate calculation, payment authorization, and confirmation with compensation for cancellations

2. **Manage Room Rates and Availability**  (Temporal Activity)
   - Attributes: Room type, rate plan, occupancy date, base price, dynamic pricing rules, restrictions, inventory allocation by channel

3. **Assign Rooms**  (Temporal Activity)
   - Attributes: Reservation reference, room number assigned, assignment logic, pre-arrival assignment vs. check-in assignment, upgrade applied

4. **Process Check-ins/outs** ⚙️ (Temporal Workflow)
   - Attributes: Reservation reference, actual arrival/departure time, ID verification, payment method capture, room key issued
   - Process: Verifies guest identity, confirms reservation details, captures payment guarantee, generates room key. At checkout: reconciles charges, processes final payment
   - Temporal: Coordinates reservation lookup, payment processing, room status updates, and housekeeping notification

5. **Schedule Housekeeping**  (Temporal Activity)
   - Attributes: Room number, cleaning type, priority, room status, assigned housekeeper, estimated completion time

##### Service: Point of Sale (POS) & Table Management - Restaurants

**Resource: Dining Table / Order**

*Actions:*

1. **Manage Floor Plan**  (Temporal Activity)
   - Attributes: Table identifier, table capacity, location, table status, server assigned to section, table combination rules

2. **Take Reservation**  (Temporal Activity)
   - Attributes: Guest name, contact info, party size, date and time, special requests, reservation status, no-show policy

3. **Assign Party**  (Temporal Activity)
   - Attributes: Party size, table assignment, server assigned, special needs, guest preferences, estimated duration

4. **Take Orders**  (Temporal Activity)
   - Attributes: Table identifier, order items, course sequence, guest seat position, dietary flags, order time

5. **Route to Kitchen/Prep Stations**  (Temporal Activity)
   - Attributes: Order identifier, menu item, quantity, preparation instructions, priority, course timing, station

6. **Process Payments**  (Temporal Activity)
   - Attributes: Table identifier, check total, payment method, tip amount, split payment, receipt preferences

7. **Handle Order Modifications**  (Temporal Activity)
   - Attributes: Original order item, modification type, reason, timing, refire instructions, discount applied

8. **Track Table Turn Time**  (Temporal Activity)
   - Attributes: Table identifier, party seated time, party departed time, turn time calculated, target turn time

##### Service: Menu & Recipe Management

**Resource: Menu Item**

*Actions:*

1. **Update Menu & Prices**  (Temporal Activity)
   - Attributes: Item name, description, category, price by meal period, availability, active/inactive status

2. **Set Seasonal Items**  (Temporal Activity)
   - Attributes: Menu item, seasonality pattern, active date range, promotion/markup, featured item flag

3. **Manage Recipes & Ingredients**  (Temporal Activity)
   - Attributes: Menu item, ingredient list, preparation steps, portion size, plate presentation, cook time, yield

4. **Track Item Popularity**  (Temporal Activity)
   - Attributes: Menu item, sales volume by period, revenue contribution, food cost percentage, customer ratings

---

#### Retail Management

##### Service: Point of Sale (POS) & Checkout

**Resource: Sales Transaction**

*Actions:*

1. **Scan Item**  (Temporal Activity)
   - Attributes: Product identifier, quantity, unit price, discounts applicable, tax category, inventory location

2. **Process Payment**  (Temporal Activity)
   - Attributes: Payment method, amount tendered, card details, transaction fee, receipt preferences, authorization code

3. **Apply Promotions**  (Temporal Activity)
   - Attributes: Promotion identifier, promotion type, eligibility criteria, stack-ability

4. **Issue Receipt**  (Temporal Activity)
   - Attributes: Transaction number, date/time, store location, itemized list, discounts, tax, total, payment method

5. **Handle Returns** ⚙️ (Temporal Workflow)
   - Attributes: Original transaction reference, items to return, return reason, condition, refund method, restocking fee
   - Process: Locates original transaction, verifies return policy compliance, assesses item condition, processes refund or exchange
   - Temporal: Coordinates return validation, refund processing, and inventory adjustment

##### Service: Omnichannel Sales

**Resource: Product Catalog**

*Actions:*

1. **Sync Stock Levels Across Channels** ⚙️ (Temporal Workflow - Continuous)
   - Attributes: Product identifier, channel-specific inventory, reserved quantities, available-to-promise calculation, refresh frequency
   - Process: Aggregates inventory from all locations, subtracts reservations, calculates available quantity, publishes to all sales channels
   - Temporal: Continuous workflow that maintains real-time inventory sync with conflict resolution

2. **Fulfill Online Orders In-Store** ⚙️ (Temporal Workflow)
   - Attributes: Online order identifier, fulfillment location, inventory source selection logic, picking instructions, shipping method
   - Process: Evaluates optimal fulfillment location, routes order, generates pick list, reserves inventory, packs and ships
   - Temporal: Orchestrates order routing, picking, packing, and shipping across multiple locations

3. **Provide Buy-Online-Pickup-In-Store (BOPIS)** ⚙️ (Temporal Workflow)
   - Attributes: Online order identifier, pickup location, ready-for-pickup notification, pickup time window, customer identification method
   - Process: Processes online order, reserves inventory at pickup location, prepares order, notifies customer, holds at designated area
   - Temporal: Coordinates order preparation, inventory allocation, and customer notification with timeout handling

##### Service: Merchandising & Assortment Planning

**Resource: Product Category**

*Actions:*

1. **Plan Assortment**  (Temporal Activity)
   - Attributes: Category identifier, target breadth and depth, price point distribution, seasonal considerations, space allocation

2. **Set Category-level Pricing**  (Temporal Activity)
   - Attributes: Category identifier, pricing strategy, competitive positioning, margin targets, promotion frequency

3. **Analyze Sales Performance**  (Temporal Activity)
   - Attributes: Category identifier, time period, metrics, comparison periods, top/bottom performers, trend analysis

4. **Manage Supplier Styles**  (Temporal Activity)
   - Attributes: Supplier identifier, product styles offered, lead times, minimum order quantities, quality metrics

---

#### Forecourt/Petrol Station Management

##### Service: Fuel Management

**Resource: Fuel Pump**

*Actions:*

1. **Monitor Tank Levels**  (Temporal Activity - Continuous)
   - Attributes: Tank identifier, fuel grade, current volume, capacity, reorder level, temperature, water detection, variance
   - Process: Continuously monitors tank levels via ATG, tracks dispensing, calculates remaining volume, alerts on variances or low levels
   - Temporal: Continuous activity with alerting on threshold breaches

2. **Set Fuel Prices**  (Temporal Activity)
   - Attributes: Fuel grade, retail price per unit, effective date/time, cost basis, margin target, competitive positioning

3. **Control Pump Authorization**  (Temporal Activity)
   - Attributes: Pump number, authorization mode, authorized amount, payment method captured, customer identifier

4. **Generate Wet Stock Reports**  (Temporal Activity)
   - Attributes: Report date, tank readings, calculated variance, variance tolerance, leak detection status, compliance status

##### Service: Forecourt Controller

**Resource: Transaction**

*Actions:*

1. **Link Pump to POS Sale**  (Temporal Activity)
   - Attributes: Pump number, fuel grade dispensed, volume, price per unit, total fuel amount, transaction timestamp

2. **Authorize Pre-pay/Post-pay** ⚙️ (Temporal Workflow)
   - Attributes: Authorization mode, customer payment method, pre-authorization amount, pump number, authorization status
   - Process: For pre-pay: captures payment, authorizes pump, releases unused authorization. For post-pay: validates payment method, authorizes pump, captures actual amount
   - Temporal: Workflow handles payment authorization, pump control, and settlement with compensation for cancellations

3. **Process Car Wash Activation**  (Temporal Activity)
   - Attributes: Car wash type, price, activation code, validity period, pump transaction link, usage status

4. **Manage Loyalty Points at Pump**  (Temporal Activity)
   - Attributes: Customer loyalty identifier, points balance, points earned, points redeemed, redemption value

##### Service: Convenience Store Retail

**Resource: Non-Fuel Product**

*Actions:*

1. **Manage Store Inventory**  (Temporal Activity)
   - Attributes: Product identifier, category, quantity on hand, reorder point, supplier, shelf location, perishability

2. **Run Promotions**  (Temporal Activity)
   - Attributes: Promotion type, eligible products, promotion period, discount amount, coordination with fuel promotions

3. **Track Sales Performance**  (Temporal Activity)
   - Attributes: Time period, metrics by category, fuel correlation analysis, top selling items, slow movers

4. **Replenish Stock**  (Temporal Activity)
   - Attributes: Supplier identifier, delivery frequency, order lead time, order minimum, product list with quantities

---

### Technical Foundation Modules

#### Security & Access Control

**Key Capabilities**:
- User Authentication (password, MFA, SSO, biometric)
- Role-Based Access Control (RBAC)
- Attribute-Based Access Control (ABAC)
- Data Encryption (at-rest, in-transit, field-level)
- Audit Logging

#### Analytics & Business Intelligence (BI)

**Key Capabilities**:
- Data Warehouse
- ETL Processes
- Report Builder
- Interactive Dashboards
- Predictive Analytics
- Ad Hoc Analysis

#### Integration & APIs

**Key Capabilities**:
- RESTful APIs
- Event-Driven Architecture (Temporal-based)
- Web Services (SOAP for legacy)
- File-Based Integration
- API Gateway
- Webhooks

#### Data Management

**Key Capabilities**:
- Master Data Management (MDM)
- Data Quality Tools
- Data Governance
- Backup & Recovery
- Data Migration
- Data Archiving

---

### Implementation Approach: Dependencies and Sequencing

#### Phase 1: Foundation (Months 1-6)

**Core Modules**:
1. Financial Management (General Ledger, Chart of Accounts)
2. Security & Access Control
3. Data Management

#### Phase 2: Operations (Months 6-12)

**Core Modules**:
1. Inventory Management
2. Accounts Payable & Receivable
3. Basic CRM
4. Basic HCM

#### Phase 3: Extended Operations (Months 12-24)

**Industry-Specific Modules** (choose based on business):
- Manufacturing
- Retail POS
- Restaurant/Hotel PMS
- Forecourt

**Advanced Core Modules**:
- Procurement
- Warehouse Management
- Advanced CRM
- Advanced HCM
- Project Management

#### Phase 4: Optimization (Months 24+)

**Technical Modules**:
- Business Intelligence
- Advanced Integration
- Predictive Analytics

---

## Section 2: Technical Architecture Guide

### System Philosophy

Awo ERP is built on a **schema-driven architecture** that fundamentally separates concerns between backend logic and UI presentation. The core innovation is allowing backend developers to define complete user interfaces through JSON configuration.

**Key Architectural Principles**:

1. **Declarative UI Definition**: User interfaces described as data structures (schemas), not imperative code
2. **Backend-First Development**: Business logic and UI definition coexist in the backend
3. **Progressive Enhancement**: Server-rendered HTML enhanced with minimal JavaScript
4. **Security by Default**: Multi-tenant isolation enforced at database level
5. **Workflow Orchestration**: Complex processes managed by Temporal workflows
6. **Convention Over Configuration**: Sensible defaults reduce boilerplate

---

### Technology Foundation

#### Backend Layer

**Go** serves as the primary language for:
- Interface-Based Design for testability
- Composition Over Inheritance
- Context Propagation (tenant ID, user ID)
- Error Wrapping with context

#### Data Layer

**PostgreSQL** with:
- Row-Level Security (RLS) for tenant isolation
- JSONB for flexible storage
- Advanced Indexing
- Transactional DDL
- Full-Text Search
- Triggers and Functions

**SQLC** generates type-safe Go code from SQL queries.

#### Workflow Orchestration Layer

**Temporal** provides:
- Durable workflow execution
- Activity task distribution
- Long-running process management
- Automatic retry and compensation
- Saga pattern implementation
- Event-driven coordination
- Scheduled/cron workflows

**Temporal Architecture in Awo ERP**:
```
┌─────────────────────────────────────────────────────────┐
│                   Frontend (HTMX/Alpine)                │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│              HTTP Handlers (Go/Fiber)                   │
└─────────────────────┬───────────────────────────────────┘
                      │
         ┌────────────┴────────────┐
         │                         │
┌────────▼──────────┐    ┌────────▼────────────────────┐
│  Service Layer    │    │   Temporal Workflows        │
│  (Simple Logic)   │    │   (Complex Orchestration)   │
└────────┬──────────┘    └────────┬────────────────────┘
         │                        │
         │              ┌─────────┴─────────┐
         │              │                   │
         │    ┌─────────▼─────────┐  ┌─────▼──────────┐
         │    │ Temporal Workers  │  │ Temporal Server│
         │    │  (Activities)     │  │ (Orchestration)│
         │    └─────────┬─────────┘  └────────────────┘
         │              │
┌────────▼──────────────▼──────────────────────────────┐
│           Repository Layer (SQLC)                     │
└────────┬──────────────────────────────────────────────┘
         │
┌────────▼──────────────────────────────────────────────┐
│        PostgreSQL with RLS (Multi-tenant DB)          │
└───────────────────────────────────────────────────────┘
```

#### Frontend Layer

**Templ**: Type-safe HTML templating compiled to Go code

**HTMX**: Dynamic interactions through HTML attributes
- `hx-get`: Load content on demand
- `hx-post`: Submit forms asynchronously
- `hx-swap`: Control response HTML replacement
- `hx-trigger`: Define when requests fire

**Alpine.js**: Lightweight client-side state for:
- Dropdown menus
- Modal dialogs
- Form field visibility
- Client-side validation

**Tailwind CSS**: Utility-first CSS with dynamic theme system

---

### Multi-Tenancy Architecture

#### Shared Database, Shared Schema with Row-Level Isolation

**Session-Based Tenant Isolation**:
```sql
-- When request arrives, middleware sets tenant context:
SET LOCAL app.current_tenant_id = '123e4567-e89b-12d3-a456-426614174000';

-- RLS policy on every tenant-scoped table:
CREATE POLICY tenant_isolation_policy ON invoices
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

-- All queries automatically filtered by tenant
```

**Advantages**:
1. Security enforced at database level
2. Resource efficiency through sharing
3. Operational simplicity (single database)
4. Cost efficiency
5. Cross-tenant analytics capability

**Tenant Resolution Flow**:
1. HTTP request arrives
2. Authentication middleware validates JWT
3. Token contains tenant identifier
4. Middleware sets tenant context in `context.Context`
5. Database transaction starts with session variable
6. All queries respect RLS policies
7. Transaction commits/rolls back
8. Connection returned to pool

---

### Schema-Driven UI System

#### Core Concepts

**Schema as Contract**: JSON document describing UI component—type, properties, behavior, data bindings.

**Component Hierarchy**:
- **Atoms**: Buttons, inputs, labels, badges
- **Molecules**: Form fields, cards
- **Organisms**: Forms, data tables, navigation
- **Templates**: Page layouts
- **Pages**: Specific template instances with data

#### Schema Definitions (Go Structs)
```go
// Page schema
type Page struct {
    Type    string      `json:"type"`
    Title   string      `json:"title"`
    Body    []Component `json:"body"`
    Aside   []Component `json:"aside,omitempty"`
    Actions []Action    `json:"actions,omitempty"`
    Theme   string      `json:"theme,omitempty"`
}

// Form schema
type Form struct {
    Type       string          `json:"type"`
    Title      string          `json:"title,omitempty"`
    Fields     []Field         `json:"fields"`
    Actions    []Action        `json:"actions"`
    InitAPI    string          `json:"initApi,omitempty"`
    SubmitAPI  string          `json:"api"`
    Validation ValidationRules `json:"validation,omitempty"`
}

// DataTable schema
type DataTable struct {
    Type        string   `json:"type"`
    API         string   `json:"api"`
    Columns     []Column `json:"columns"`
    Filter      *Form    `json:"filter,omitempty"`
    BulkActions []Action `json:"bulkActions,omitempty"`
    QuickEdit   *Form    `json:"quickEdit,omitempty"`
    ItemActions []Action `json:"itemActions,omitempty"`
    Pagination  bool     `json:"pagination"`
}

// Field schema
type Field struct {
    Type        string       `json:"type"`
    Name        string       `json:"name"`
    Label       string       `json:"label"`
    Placeholder string       `json:"placeholder,omitempty"`
    Required    bool         `json:"required,omitempty"`
    Validations []Validation `json:"validations,omitempty"`
    Visible     string       `json:"visible,omitempty"`
    Options     []Option     `json:"options,omitempty"`
    SourceAPI   string       `json:"source,omitempty"`
}
```

#### JSON Schema Examples

**Simple Form Schema (JSON)**:
```json
{
  "type": "form",
  "title": "Create Customer",
  "fields": [
    {
      "type": "text",
      "name": "company_name",
      "label": "Company Name",
      "required": true,
      "placeholder": "Acme Corporation",
      "validations": [
        {
          "type": "required",
          "message": "Company name is required"
        },
        {
          "type": "min",
          "value": 2,
          "message": "Company name must be at least 2 characters"
        }
      ]
    },
    {
      "type": "email",
      "name": "email",
      "label": "Email Address",
      "required": true,
      "validations": [
        {
          "type": "email",
          "message": "Invalid email format"
        }
      ]
    },
    {
      "type": "phone",
      "name": "phone",
      "label": "Phone Number",
      "placeholder": "+1 (555) 123-4567"
    },
    {
      "type": "select",
      "name": "industry",
      "label": "Industry",
      "required": true,
      "options": [
        {"label": "Technology", "value": "tech"},
        {"label": "Manufacturing", "value": "mfg"},
        {"label": "Retail", "value": "retail"},
        {"label": "Services", "value": "services"}
      ]
    },
    {
      "type": "textarea",
      "name": "notes",
      "label": "Notes",
      "rows": 3,
      "placeholder": "Additional information..."
    }
  ],
  "actions": [
    {
      "type": "submit",
      "label": "Create Customer",
      "level": "primary"
    },
    {
      "type": "reset",
      "label": "Cancel",
      "level": "default"
    }
  ],
  "api": "/api/customers"
}
```

**DataTable Schema (JSON)**:
```json
{
  "type": "crud",
  "title": "Customer List",
  "api": "/api/customers",
  "filter": {
    "fields": [
      {
        "type": "text",
        "name": "search",
        "label": "Search",
        "placeholder": "Search by name or email..."
      },
      {
        "type": "select",
        "name": "industry",
        "label": "Industry",
        "options": [
          {"label": "All Industries", "value": ""},
          {"label": "Technology", "value": "tech"},
          {"label": "Manufacturing", "value": "mfg"}
        ]
      },
      {
        "type": "select",
        "name": "status",
        "label": "Status",
        "options": [
          {"label": "All", "value": ""},
          {"label": "Active", "value": "active"},
          {"label": "Inactive", "value": "inactive"}
        ]
      }
    ]
  },
  "columns": [
    {
      "name": "company_name",
      "label": "Company Name",
      "sortable": true,
      "type": "text"
    },
    {
      "name": "email",
      "label": "Email",
      "sortable": true,
      "type": "text"
    },
    {
      "name": "industry",
      "label": "Industry",
      "sortable": true,
      "type": "text"
    },
    {
      "name": "total_revenue",
      "label": "Total Revenue",
      "sortable": true,
      "type": "money"
    },
    {
      "name": "status",
      "label": "Status",
      "type": "status",
      "mapping": {
        "active": {
          "label": "Active",
          "variant": "success"
        },
        "inactive": {
          "label": "Inactive",
          "variant": "default"
        }
      }
    }
  ],
  "itemActions": [
    {
      "type": "link",
      "label": "View",
      "icon": "eye",
      "link": "/customers/${id}"
    },
    {
      "type": "link",
      "label": "Edit",
      "icon": "edit",
      "link": "/customers/${id}/edit"
    },
    {
      "type": "ajax",
      "label": "Deactivate",
      "icon": "ban",
      "api": "/api/customers/${id}/deactivate",
      "method": "POST",
      "confirmText": "Deactivate this customer?",
      "visible": "${status === 'active'}"
    }
  ],
  "bulkActions": [
    {
      "type": "ajax",
      "label": "Export Selected",
      "api": "/api/customers/export",
      "method": "POST"
    },
    {
      "type": "ajax",
      "label": "Send Email Campaign",
      "api": "/api/customers/bulk/email-campaign",
      "method": "POST"
    }
  ],
  "pagination": true
}
```

**Complex Form with Conditional Fields (JSON)**:
```json
{
  "type": "form",
  "title": "Supplier Invoice",
  "fields": [
    {
      "type": "select",
      "name": "supplier_id",
      "label": "Supplier",
      "required": true,
      "source": "/api/suppliers/search",
      "searchApi": "/api/suppliers/search?q=${term}"
    },
    {
      "type": "text",
      "name": "invoice_number",
      "label": "Invoice Number",
      "required": true,
      "placeholder": "INV-2025-001"
    },
    {
      "type": "date",
      "name": "invoice_date",
      "label": "Invoice Date",
      "required": true
    },
    {
      "type": "date",
      "name": "due_date",
      "label": "Due Date",
      "description": "Leave blank to calculate from supplier payment terms"
    },
    {
      "type": "select",
      "name": "purchase_order_id",
      "label": "Purchase Order (Optional)",
      "source": "/api/purchase-orders/open",
      "description": "Link to PO for three-way matching"
    },
    {
      "type": "checkbox",
      "name": "requires_approval",
      "label": "Requires Approval",
      "value": false
    },
    {
      "type": "select",
      "name": "approver_id",
      "label": "Approver",
      "required": true,
      "source": "/api/users/approvers",
      "visible": "${requires_approval === true}"
    },
    {
      "type": "table",
      "name": "line_items",
      "label": "Invoice Lines",
      "columns": [
        {
          "name": "description",
          "label": "Description",
          "type": "text",
          "required": true
        },
        {
          "name": "quantity",
          "label": "Qty",
          "type": "number",
          "required": true,
          "validations": [
            {
              "type": "min",
              "value": 0.01,
              "message": "Quantity must be greater than 0"
            }
          ]
        },
        {
          "name": "unit_price",
          "label": "Unit Price",
          "type": "money",
          "required": true
        },
        {
          "name": "line_total",
          "label": "Total",
          "type": "money",
          "readonly": true,
          "formula": "${quantity} * ${unit_price}"
        },
        {
          "name": "account_code",
          "label": "GL Account",
          "type": "select",
          "source": "/api/accounts/expense"
        }
      ]
    },
    {
      "type": "textarea",
      "name": "notes",
      "label": "Notes",
      "rows": 3
    }
  ],
  "actions": [
    {
      "type": "submit",
      "label": "Save Invoice",
      "level": "primary"
    },
    {
      "type": "reset",
      "label": "Cancel",
      "level": "default"
    }
  ],
  "api": "/api/supplier-invoices"
}
```

**Page Layout Schema (JSON)**:
```json
{
  "type": "page",
  "title": "Dashboard",
  "body": [
    {
      "type": "grid",
      "columns": 3,
      "gap": 4,
      "items": [
        {
          "type": "card",
          "title": "Total Revenue",
          "body": {
            "type": "stat",
            "value": "$125,430",
            "trend": "+12.5%",
            "trendDirection": "up",
            "comparison": "vs last month"
          }
        },
        {
          "type": "card",
          "title": "Active Customers",
          "body": {
            "type": "stat",
            "value": "1,234",
            "trend": "+5.2%",
            "trendDirection": "up"
          }
        },
        {
          "type": "card",
          "title": "Pending Orders",
          "body": {
            "type": "stat",
            "value": "56",
            "trend": "-8.1%",
            "trendDirection": "down"
          }
        }
      ]
    },
    {
      "type": "card",
      "title": "Recent Orders",
      "body": {
        "type": "table",
        "api": "/api/orders/recent",
        "columns": [
          {"name": "order_number", "label": "Order #"},
          {"name": "customer", "label": "Customer"},
          {"name": "date", "label": "Date", "type": "date"},
          {"name": "total", "label": "Amount", "type": "money"},
          {"name": "status", "label": "Status", "type": "status"}
        ]
      }
    }
  ],
  "aside": [
    {
      "type": "card",
      "title": "Quick Actions",
      "body": {
        "type": "list",
        "items": [
          {"type": "link", "label": "Create Order", "href": "/orders/new"},
          {"type": "link", "label": "Add Customer", "href": "/customers/new"},
          {"type": "link", "label": "View Reports", "href": "/reports"}
        ]
      }
    }
  ]
}
```

#### Rendering Pipeline

1. **Schema Generation**: Backend service creates schema struct
2. **JSON Serialization**: Schema serialized to JSON or kept as Go struct
3. **Theme Application**: Theme processor injects design tokens
4. **Component Resolution**: Rendering engine identifies component type
5. **HTML Generation**: Templ template generates HTML with HTMX attributes
6. **Client Hydration**: Alpine.js initializes client-side state

#### Theme System
```go
type Theme struct {
    ID         string           `json:"id"`
    Name       string           `json:"name"`
    Colors     ColorTokens      `json:"colors"`
    Typography TypographyTokens `json:"typography"`
    Spacing    SpacingTokens    `json:"spacing"`
    Borders    BorderTokens     `json:"borders"`
    Shadows    ShadowTokens     `json:"shadows"`
}

type ColorTokens struct {
    Primary   ColorScale `json:"primary"`
    Secondary ColorScale `json:"secondary"`
    Success   ColorScale `json:"success"`
    Warning   ColorScale `json:"warning"`
    Error     ColorScale `json:"error"`
    Neutral   ColorScale `json:"neutral"`
}
```

---

### Backend Architecture

#### Architectural Layers
```
HTTP Layer (handlers, middleware)
        ↓
Service Layer (business logic)
        ↓
Temporal Workflows & Activities (complex orchestration)
        ↓
Repository Layer (data access)
        ↓
Database (PostgreSQL with RLS)
```

#### HTTP Layer

**Middleware Chain**:
```go
app.Use(
    Logger(),
    Recovery(),
    CORS(),
    RateLimiter(),
    Authenticate(),
    ResolveTenant(),
    SetTenantContext(),
    Authorize(),
)
```

#### Service Layer

Simple business logic that doesn't require workflow orchestration:
```go
type InvoiceService struct {
    invoiceRepo  InvoiceRepository
    customerRepo CustomerRepository
    temporalClient client.Client
    logger       Logger
}

// Simple operation - no workflow needed
func (s *InvoiceService) ValidateInvoice(
    ctx context.Context,
    invoice *Invoice,
) error {
    // Simple validation logic
    // No external service calls
    // No long-running operations
    return nil
}

// Complex operation - use Temporal workflow
func (s *InvoiceService) ProcessInvoice(
    ctx context.Context,
    invoice *Invoice,
) error {
    workflowOptions := client.StartWorkflowOptions{
        ID:        fmt.Sprintf("process-invoice-%s", invoice.ID),
        TaskQueue: "invoices",
    }
    
    we, err := s.temporalClient.ExecuteWorkflow(
        ctx,
        workflowOptions,
        workflows.ProcessInvoiceWorkflow,
        invoice,
    )
    if err != nil {
        return err
    }
    
    return we.Get(ctx, nil)
}
```

---

### Temporal Workflow Integration

#### When to Use Temporal

**Use Temporal Workflows for**:
- Operations spanning multiple services
- Long-running processes (>30 seconds)
- Operations requiring compensation/rollback
- Operations with complex retry logic
- Scheduled/cron operations
- Event-driven coordination

**Use Temporal Activities for**:
- Individual service operations within workflows
- External API calls
- Database operations that need retry logic
- Operations that might fail and need automatic retry

**Use Simple Service Methods for**:
- Pure business logic calculations
- Single database operations
- Simple validations
- Operations completing in <1 second

#### Temporal Architecture Components
```go
// 1. Workflow Definition
func ProcessInvoiceWorkflow(ctx workflow.Context, invoice *Invoice) error {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 10 * time.Second,
        RetryPolicy: &temporal.RetryPolicy{
            MaximumAttempts: 3,
        },
    }
    ctx = workflow.WithActivityOptions(ctx, ao)
    
    // Step 1: Validate supplier
    var supplierValid bool
    err := workflow.ExecuteActivity(ctx, ValidateSupplier, invoice.SupplierID).Get(ctx, &supplierValid)
    if err != nil || !supplierValid {
        return err
    }
    
    // Step 2: Three-way match (if PO linked)
    if invoice.PurchaseOrderID != nil {
        err = workflow.ExecuteActivity(ctx, PerformThreeWayMatch, invoice).Get(ctx, nil)
        if err != nil {
            return err
        }
    }
    
    // Step 3: Route for approval
    err = workflow.ExecuteActivity(ctx, RouteForApproval, invoice).Get(ctx, nil)
    if err != nil {
        return err
    }
    
    // Step 4: Wait for approval (signal)
    var approved bool
    selector := workflow.NewSelector(ctx)
    
    approvalSignal := workflow.GetSignalChannel(ctx, "invoice_approved")
    rejectionSignal := workflow.GetSignalChannel(ctx, "invoice_rejected")
    
    selector.AddReceive(approvalSignal, func(c workflow.ReceiveChannel, more bool) {
        c.Receive(ctx, &approved)
        approved = true
    })
    
    selector.AddReceive(rejectionSignal, func(c workflow.ReceiveChannel, more bool) {
        approved = false
    })
    
    selector.Select(ctx)
    
    if !approved {
        // Compensate - reverse previous steps
        workflow.ExecuteActivity(ctx, ReverseInvoiceCreation, invoice).Get(ctx, nil)
        return errors.New("invoice rejected")
    }
    
    // Step 5: Post to General Ledger
    err = workflow.ExecuteActivity(ctx, PostToGeneralLedger, invoice).Get(ctx, nil)
    if err != nil {
        // Compensate
        workflow.ExecuteActivity(ctx, ReverseApproval, invoice).Get(ctx, nil)
        return err
    }
    
    // Step 6: Update invoice status
    err = workflow.ExecuteActivity(ctx, UpdateInvoiceStatus, invoice.ID, "approved").Get(ctx, nil)
    
    return err
}

// 2. Activity Definitions
func ValidateSupplier(ctx context.Context, supplierID string) (bool, error) {
    // Call supplier repository
    // Return validation result
}

func PerformThreeWayMatch(ctx context.Context, invoice *Invoice) error {
    // Match invoice to PO and GRN
    // Return error if mismatch
}

func PostToGeneralLedger(ctx context.Context, invoice *Invoice) error {
    // Create journal entries
    // Post to GL
}

// 3. Worker Registration
func main() {
    c, err := client.Dial(client.Options{})
    if err != nil {
        log.Fatalln("Unable to create Temporal client", err)
    }
    defer c.Close()
    
    w := worker.New(c, "invoices", worker.Options{})
    
    // Register workflows
    w.RegisterWorkflow(ProcessInvoiceWorkflow)
    w.RegisterWorkflow(SchedulePaymentWorkflow)
    
    // Register activities
    w.RegisterActivity(ValidateSupplier)
    w.RegisterActivity(PerformThreeWayMatch)
    w.RegisterActivity(RouteForApproval)
    w.RegisterActivity(PostToGeneralLedger)
    w.RegisterActivity(UpdateInvoiceStatus)
    w.RegisterActivity(ReverseInvoiceCreation)
    w.RegisterActivity(ReverseApproval)
    
    err = w.Run(worker.InterruptCh())
    if err != nil {
        log.Fatalln("Unable to start worker", err)
    }
}
```

#### Saga Pattern with Temporal

For operations requiring compensation:
```go
func OrderFulfillmentWorkflow(ctx workflow.Context, order *Order) error {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 30 * time.Second,
    }
    ctx = workflow.WithActivityOptions(ctx, ao)
    
    // Saga step 1: Reserve inventory
    var inventoryReserved bool
    err := workflow.ExecuteActivity(ctx, ReserveInventory, order).Get(ctx, &inventoryReserved)
    if err != nil {
        return err
    }
    
    // Saga step 2: Authorize payment
    var paymentAuthorized bool
    err = workflow.ExecuteActivity(ctx, AuthorizePayment, order).Get(ctx, &paymentAuthorized)
    if err != nil {
        // Compensate: Release inventory
        workflow.ExecuteActivity(ctx, ReleaseInventory, order).Get(ctx, nil)
        return err
    }
    
    // Saga step 3: Create shipment
    var shipmentID string
    err = workflow.ExecuteActivity(ctx, CreateShipment, order).Get(ctx, &shipmentID)
    if err != nil {
        // Compensate: Void payment and release inventory
        workflow.ExecuteActivity(ctx, VoidPayment, order).Get(ctx, nil)
        workflow.ExecuteActivity(ctx, ReleaseInventory, order).Get(ctx, nil)
        return err
    }
    
    // Saga step 4: Capture payment
    err = workflow.ExecuteActivity(ctx, CapturePayment, order).Get(ctx, nil)
    if err != nil {
        // Compensate: Cancel shipment, void payment, release inventory
        workflow.ExecuteActivity(ctx, CancelShipment, shipmentID).Get(ctx, nil)
        workflow.ExecuteActivity(ctx, VoidPayment, order).Get(ctx, nil)
        workflow.ExecuteActivity(ctx, ReleaseInventory, order).Get(ctx, nil)
        return err
    }
    
    return nil
}
```

#### Scheduled Workflows

For recurring operations:
```go
// Cron workflow for recurring invoices
func RecurringInvoiceWorkflow(ctx workflow.Context) error {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 5 * time.Minute,
    }
    ctx = workflow.WithActivityOptions(ctx, ao)
    
    // Get all recurring invoice configurations
    var configs []RecurringInvoiceConfig
    err := workflow.ExecuteActivity(ctx, GetRecurringInvoiceConfigs).Get(ctx, &configs)
    if err != nil {
        return err
    }
    
    // Process each configuration
    for _, config := range configs {
        // Check if invoice is due today
        if config.IsDueToday() {
            // Generate and send invoice
            err = workflow.ExecuteActivity(ctx, GenerateRecurringInvoice, config).Get(ctx, nil)
            if err != nil {
                workflow.GetLogger(ctx).Error("Failed to generate recurring invoice", "config_id", config.ID, "error", err)
                // Continue to next config instead of failing entire workflow
                continue
            }
        }
    }
    
    return nil
}

// Schedule the cron workflow
workflowOptions := client.StartWorkflowOptions{
    ID:           "recurring-invoices-cron",
    TaskQueue:    "invoices",
    CronSchedule: "0 0 * * *", // Daily at midnight
}

we, err := temporalClient.ExecuteWorkflow(
    context.Background(),
    workflowOptions,
    RecurringInvoiceWorkflow,
)
```

---

### Data Layer Design

#### Core Schema Patterns
```sql
-- Tenant-scoped table
CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Business fields
    invoice_number VARCHAR(50) NOT NULL,
    customer_id UUID NOT NULL,
    invoice_date DATE NOT NULL,
    due_date DATE NOT NULL,
    subtotal DECIMAL(15,2) NOT NULL,
    tax_amount DECIMAL(15,2) NOT NULL,
    total DECIMAL(15,2) NOT NULL,
    status VARCHAR(20) NOT NULL,
    
    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    
    -- Constraints
    CONSTRAINT unique_invoice_number_per_tenant UNIQUE(tenant_id, invoice_number),
    CONSTRAINT positive_amounts CHECK (subtotal >= 0 AND tax_amount >= 0 AND total >= 0),
    CONSTRAINT valid_total CHECK (total = subtotal + tax_amount)
);

-- Indexes
CREATE INDEX idx_invoices_tenant ON invoices(tenant_id);
CREATE INDEX idx_invoices_tenant_customer ON invoices(tenant_id, customer_id);
CREATE INDEX idx_invoices_tenant_status ON invoices(tenant_id, status);
CREATE INDEX idx_invoices_tenant_date ON invoices(tenant_id, invoice_date);

-- RLS Policy
ALTER TABLE invoices ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_policy ON invoices
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);
```

#### Audit Trail with JSONB
```sql
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    
    entity_type VARCHAR(100) NOT NULL,
    entity_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    
    user_id UUID NOT NULL REFERENCES users(id),
    username VARCHAR(255) NOT NULL,
    
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    
    old_values JSONB,
    new_values JSONB,
    changes JSONB,
    
    ip_address INET,
    user_agent TEXT,
    request_id UUID
);

CREATE INDEX idx_audit_tenant_entity ON audit_log(tenant_id, entity_type, entity_id);
CREATE INDEX idx_audit_timestamp ON audit_log(timestamp DESC);
```

#### Financial Accounting Schema
```sql
-- Chart of Accounts
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    code VARCHAR(20) NOT NULL,
    name VARCHAR(255) NOT NULL,
    account_type VARCHAR(20) NOT NULL CHECK (account_type IN ('asset', 'liability', 'equity', 'revenue', 'expense')),
    parent_account_id UUID REFERENCES accounts(id),
    level INT NOT NULL,
    active BOOLEAN DEFAULT true,
    CONSTRAINT unique_account_code_per_tenant UNIQUE(tenant_id, code)
);

-- Journal Entries
CREATE TABLE journal_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    entry_date DATE NOT NULL,
    entry_type VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    reference_type VARCHAR(100),
    reference_id UUID,
    posted BOOLEAN DEFAULT false,
    posted_at TIMESTAMPTZ,
    posted_by UUID,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id)
);

-- Journal Entry Lines
CREATE TABLE journal_entry_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_entry_id UUID NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id),
    debit DECIMAL(15,2) DEFAULT 0 CHECK (debit >= 0),
    credit DECIMAL(15,2) DEFAULT 0 CHECK (credit >= 0),
    description TEXT,
    CONSTRAINT debit_or_credit_not_both CHECK (
        (debit > 0 AND credit = 0) OR (credit > 0 AND debit = 0)
    )
);

-- Enforce double-entry bookkeeping
CREATE OR REPLACE FUNCTION check_journal_entry_balance()
RETURNS TRIGGER AS $$
DECLARE
    total_debits DECIMAL(15,2);
    total_credits DECIMAL(15,2);
BEGIN
    SELECT 
        COALESCE(SUM(debit), 0),
        COALESCE(SUM(credit), 0)
    INTO total_debits, total_credits
    FROM journal_entry_lines
    WHERE journal_entry_id = NEW.journal_entry_id;
    
    IF total_debits != total_credits THEN
        RAISE EXCEPTION 'Journal entry debits (%) must equal credits (%)', 
            total_debits, total_credits;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE CONSTRAINT TRIGGER journal_entry_balance_check
    AFTER INSERT OR UPDATE ON journal_entry_lines
    DEFERRABLE INITIALLY DEFERRED
    FOR EACH ROW EXECUTE FUNCTION check_journal_entry_balance();
```

---

### Testing Strategy

#### Unit Testing
```go
func TestCreateInvoice_Success(t *testing.T) {
    mockRepo := new(MockInvoiceRepository)
    mockSupplierRepo := new(MockSupplierRepository)
    service := services.NewInvoiceService(mockRepo, mockSupplierRepo, nil, nil)
    
    supplier := &Supplier{ID: "supplier-123", Active: true}
    mockSupplierRepo.On("GetByID", mock.Anything, "supplier-123").Return(supplier, nil)
    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*Invoice")).Return(nil)
    
    invoice, err := service.CreateInvoice(context.Background(), req, "user-123")
    
    assert.NoError(t, err)
    assert.NotNil(t, invoice)
}
```

#### Integration Testing
```go
func TestInvoiceRepository_Integration(t *testing.T) {
    db := setupTestDatabase(t)
    defer db.Close()
    
    repo := repository.NewPostgresInvoiceRepository(db)
    ctx := setTenantContext(context.Background(), db, tenantID)
    
    invoice := &Invoice{
        ID:            uuid.New().String(),
        InvoiceNumber: "INV-001",
        Total:         decimal.NewFromFloat(1000),
    }
    
    err := repo.Create(ctx, invoice)
    require.NoError(t, err)
    
    retrieved, err := repo.GetByID(ctx, invoice.ID)
    require.NoError(t, err)
    assert.Equal(t, invoice.ID, retrieved.ID)
}
```

#### Temporal Workflow Testing
```go
func TestProcessInvoiceWorkflow(t *testing.T) {
    testSuite := &testsuite.WorkflowTestSuite{}
    env := testSuite.NewTestWorkflowEnvironment()
    
    // Mock activities
    env.OnActivity(ValidateSupplier, mock.Anything, "supplier-123").Return(true, nil)
    env.OnActivity(PerformThreeWayMatch, mock.Anything, mock.Anything).Return(nil)
    env.OnActivity(RouteForApproval, mock.Anything, mock.Anything).Return(nil)
    env.OnActivity(PostToGeneralLedger, mock.Anything, mock.Anything).Return(nil)
    env.OnActivity(UpdateInvoiceStatus, mock.Anything, mock.Anything, "approved").Return(nil)
    
    // Execute workflow
    env.ExecuteWorkflow(ProcessInvoiceWorkflow, &Invoice{SupplierID: "supplier-123"})
    
    // Send approval signal
    env.SignalWorkflow("invoice_approved", true)
    
    require.True(t, env.IsWorkflowCompleted())
    require.NoError(t, env.GetWorkflowError())
}
```

---

## Section 3: Implementation Guide

### Module Implementation Pattern

Every module follows this pattern:

1. Define Database Schema (PostgreSQL with RLS)
2. Generate Data Access Layer (SQLC)
3. Implement Repository Interface
4. Create Temporal Activities for simple operations
5. Create Temporal Workflows for complex orchestration
6. Implement Service Layer (wraps Temporal client)
7. Define UI Schemas (JSON)
8. Create HTTP Handlers
9. Write Tests (unit, integration, workflow)

---

### Complete Example: Accounts Payable Module

#### Step 1: Database Schema
```sql
CREATE TABLE supplier_invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    supplier_id UUID NOT NULL REFERENCES suppliers(id),
    purchase_order_id UUID REFERENCES purchase_orders(id),
    invoice_number VARCHAR(100) NOT NULL,
    invoice_date DATE NOT NULL,
    due_date DATE NOT NULL,
    subtotal DECIMAL(15,2) NOT NULL CHECK (subtotal >= 0),
    tax_amount DECIMAL(15,2) NOT NULL CHECK (tax_amount >= 0),
    total DECIMAL(15,2) NOT NULL CHECK (total >= 0),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    CONSTRAINT valid_total CHECK (total = subtotal + tax_amount),
    CONSTRAINT unique_invoice_number_per_tenant UNIQUE(tenant_id, invoice_number)
);

CREATE INDEX idx_supplier_invoices_tenant ON supplier_invoices(tenant_id);
CREATE INDEX idx_supplier_invoices_supplier ON supplier_invoices(tenant_id, supplier_id);
CREATE INDEX idx_supplier_invoices_status ON supplier_invoices(tenant_id, status);

ALTER TABLE supplier_invoices ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_policy ON supplier_invoices
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);
```

#### Step 2: SQLC Queries
```sql
-- queries/supplier_invoices.sql

-- name: CreateSupplierInvoice :exec
INSERT INTO supplier_invoices (
    id, tenant_id, supplier_id, invoice_number,
    invoice_date, due_date, subtotal, tax_amount, total, status, created_by
) VALUES (
    $1, current_setting('app.current_tenant_id')::uuid, $2, $3,
    $4, $5, $6, $7, $8, $9, $10
);

-- name: GetSupplierInvoiceByID :one
SELECT * FROM supplier_invoices WHERE id = $1;

-- name: UpdateSupplierInvoiceStatus :exec
UPDATE supplier_invoices 
SET status = $2, updated_at = NOW()
WHERE id = $1;
```

#### Step 3: Repository Implementation
```go
type SupplierInvoiceRepository interface {
    Create(ctx context.Context, invoice *SupplierInvoice) error
    GetByID(ctx context.Context, id string) (*SupplierInvoice, error)
    UpdateStatus(ctx context.Context, id string, status InvoiceStatus) error
}

type PostgresSupplierInvoiceRepository struct {
    queries *sqlc.Queries
    db      *sql.DB
}

func (r *PostgresSupplierInvoiceRepository) Create(ctx context.Context, invoice *SupplierInvoice) error {
    return r.queries.CreateSupplierInvoice(ctx, sqlc.CreateSupplierInvoiceParams{
        ID:            invoice.ID,
        SupplierID:    invoice.SupplierID,
        InvoiceNumber: invoice.InvoiceNumber,
        InvoiceDate:   invoice.InvoiceDate,
        DueDate:       invoice.DueDate,
        Subtotal:      invoice.Subtotal,
        TaxAmount:     invoice.TaxAmount,
        Total:         invoice.Total,
        Status:        string(invoice.Status),
        CreatedBy:     invoice.CreatedBy,
    })
}
```

#### Step 4: Temporal Activities
```go
// activities/supplier_invoice_activities.go

type SupplierInvoiceActivities struct {
    invoiceRepo  repository.SupplierInvoiceRepository
    supplierRepo repository.SupplierRepository
    poRepo       repository.PurchaseOrderRepository
}

func (a *SupplierInvoiceActivities) ValidateSupplier(ctx context.Context, supplierID string) (bool, error) {
    supplier, err := a.supplierRepo.GetByID(ctx, supplierID)
    if err != nil {
        return false, err
    }
    return supplier.Active, nil
}

func (a *SupplierInvoiceActivities) PerformThreeWayMatch(ctx context.Context, invoice *SupplierInvoice) error {
    if invoice.PurchaseOrderID == nil {
        return nil // No PO linked, skip three-way match
    }
    
    po, err := a.poRepo.GetByID(ctx, *invoice.PurchaseOrderID)
    if err != nil {
        return fmt.Errorf("PO not found: %w", err)
    }
    
    if po.SupplierID != invoice.SupplierID {
        return errors.New("supplier mismatch between invoice and PO")
    }
    
    if !po.GoodsReceived {
        return errors.New("goods not yet received for this PO")
    }
    
    // Additional matching logic...
    
    return nil
}

func (a *SupplierInvoiceActivities) PostToGeneralLedger(ctx context.Context, invoice *SupplierInvoice) error {
    // Build journal entry
    // Credit: Accounts Payable
    // Debit: Expense accounts
    // Post to GL service
    return nil
}

func (a *SupplierInvoiceActivities) UpdateInvoiceStatus(ctx context.Context, invoiceID string, status string) error {
    return a.invoiceRepo.UpdateStatus(ctx, invoiceID, InvoiceStatus(status))
}
```

#### Step 5: Temporal Workflow
```go
// workflows/supplier_invoice_workflows.go

func ProcessSupplierInvoiceWorkflow(ctx workflow.Context, invoice *SupplierInvoice) error {
    logger := workflow.GetLogger(ctx)
    
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 30 * time.Second,
        RetryPolicy: &temporal.RetryPolicy{
            InitialInterval:    time.Second,
            BackoffCoefficient: 2.0,
            MaximumInterval:    time.Minute,
            MaximumAttempts:    3,
        },
    }
    ctx = workflow.WithActivityOptions(ctx, ao)
    
    // Step 1: Validate supplier
    logger.Info("Validating supplier", "supplier_id", invoice.SupplierID)
    var supplierValid bool
    err := workflow.ExecuteActivity(ctx, "ValidateSupplier", invoice.SupplierID).Get(ctx, &supplierValid)
    if err != nil {
        return fmt.Errorf("supplier validation failed: %w", err)
    }
    if !supplierValid {
        return errors.New("supplier is inactive")
    }
    
    // Step 2: Perform three-way match if PO linked
    if invoice.PurchaseOrderID != nil {
        logger.Info("Performing three-way match", "po_id", *invoice.PurchaseOrderID)
        err = workflow.ExecuteActivity(ctx, "PerformThreeWayMatch", invoice).Get(ctx, nil)
        if err != nil {
            return fmt.Errorf("three-way match failed: %w", err)
        }
    }
    
    // Step 3: Route for approval
    logger.Info("Routing for approval")
    err = workflow.ExecuteActivity(ctx, "RouteForApproval", invoice).Get(ctx, nil)
    if err != nil {
        return fmt.Errorf("approval routing failed: %w", err)
    }
    
    // Step 4: Wait for approval decision (via signal)
    logger.Info("Waiting for approval decision")
    var approved bool
    var rejectionReason string
    
    selector := workflow.NewSelector(ctx)
    approvalSignal := workflow.GetSignalChannel(ctx, "invoice_approved")
    rejectionSignal := workflow.GetSignalChannel(ctx, "invoice_rejected")
    
    selector.AddReceive(approvalSignal, func(c workflow.ReceiveChannel, more bool) {
        c.Receive(ctx, nil)
        approved = true
    })
    
    selector.AddReceive(rejectionSignal, func(c workflow.ReceiveChannel, more bool) {
        c.Receive(ctx, &rejectionReason)
        approved = false
    })
    
    selector.Select(ctx)
    
    if !approved {
        logger.Info("Invoice rejected", "reason", rejectionReason)
        // Compensate - update status to rejected
        workflow.ExecuteActivity(ctx, "UpdateInvoiceStatus", invoice.ID, "rejected").Get(ctx, nil)
        return fmt.Errorf("invoice rejected: %s", rejectionReason)
    }
    
    // Step 5: Post to General Ledger
    logger.Info("Posting to general ledger")
    err = workflow.ExecuteActivity(ctx, "PostToGeneralLedger", invoice).Get(ctx, nil)
    if err != nil {
        logger.Error("Failed to post to GL", "error", err)
        // Compensate - reverse approval
        workflow.ExecuteActivity(ctx, "UpdateInvoiceStatus", invoice.ID, "pending").Get(ctx, nil)
        return fmt.Errorf("GL posting failed: %w", err)
    }
    
    // Step 6: Update invoice status to approved
    logger.Info("Updating invoice status to approved")
    err = workflow.ExecuteActivity(ctx, "UpdateInvoiceStatus", invoice.ID, "approved").Get(ctx, nil)
    if err != nil {
        return fmt.Errorf("status update failed: %w", err)
    }
    
    logger.Info("Invoice processing completed successfully")
    return nil
}
```

#### Step 6: Service Layer
```go
type SupplierInvoiceService struct {
    invoiceRepo    repository.SupplierInvoiceRepository
    temporalClient client.Client
    logger         Logger
}

func (s *SupplierInvoiceService) CreateInvoice(
    ctx context.Context,
    req CreateInvoiceRequest,
    userID string,
) (*SupplierInvoice, error) {
    // Calculate totals, validate data
    invoice := &SupplierInvoice{
        ID:            uuid.New().String(),
        SupplierID:    req.SupplierID,
        InvoiceNumber: req.InvoiceNumber,
        // ... other fields
        Status:    InvoiceStatusPending,
        CreatedBy: userID,
    }
    
    // Persist invoice
    if err := s.invoiceRepo.Create(ctx, invoice); err != nil {
        return nil, err
    }
    
    // Start Temporal workflow for processing
    workflowOptions := client.StartWorkflowOptions{
        ID:        fmt.Sprintf("process-invoice-%s", invoice.ID),
        TaskQueue: "invoices",
    }
    
    _, err := s.temporalClient.ExecuteWorkflow(
        ctx,
        workflowOptions,
        workflows.ProcessSupplierInvoiceWorkflow,
        invoice,
    )
    if err != nil {
        s.logger.Error("Failed to start workflow", "error", err)
        // Workflow will be retried automatically
    }
    
    return invoice, nil
}

func (s *SupplierInvoiceService) ApproveInvoice(
    ctx context.Context,
    invoiceID string,
    userID string,
) error {
    // Send approval signal to workflow
    workflowID := fmt.Sprintf("process-invoice-%s", invoiceID)
    
    return s.temporalClient.SignalWorkflow(
        ctx,
        workflowID,
        "",
        "invoice_approved",
        nil,
    )
}
```

#### Step 7: JSON Schema Definition
```json
{
  "type": "form",
  "title": "Supplier Invoice",
  "fields": [
    {
      "type": "select",
      "name": "supplier_id",
      "label": "Supplier",
      "required": true,
      "source": "/api/suppliers/search"
    },
    {
      "type": "text",
      "name": "invoice_number",
      "label": "Invoice Number",
      "required": true
    },
    {
      "type": "date",
      "name": "invoice_date",
      "label": "Invoice Date",
      "required": true
    },
    {
      "type": "date",
      "name": "due_date",
      "label": "Due Date"
    },
    {
      "type": "select",
      "name": "purchase_order_id",
      "label": "Purchase Order (Optional)",
      "source": "/api/purchase-orders/open"
    },
    {
      "type": "table",
      "name": "line_items",
      "label": "Invoice Lines",
      "columns": [
        {"name": "description", "label": "Description", "type": "text", "required": true},
        {"name": "quantity", "label": "Qty", "type": "number", "required": true},
        {"name": "unit_price", "label": "Unit Price", "type": "money", "required": true},
        {"name": "line_total", "label": "Total", "type": "money", "readonly": true, "formula": "${quantity} * ${unit_price}"}
      ]
    }
  ],
  "actions": [
    {"type": "submit", "label": "Save Invoice", "level": "primary"},
    {"type": "reset", "label": "Cancel", "level": "default"}
  ],
  "api": "/api/supplier-invoices"
}
```

#### Step 8: HTTP Handler
```go
type SupplierInvoiceHandler struct {
    service *SupplierInvoiceService
}

func (h *SupplierInvoiceHandler) Create(c *fiber.Ctx) error {
    var req CreateInvoiceRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(ErrorResponse{Error: "Invalid request"})
    }
    
    userID := c.Locals("user_id").(string)
    invoice, err := h.service.CreateInvoice(c.Context(), req, userID)
    if err != nil {
        return c.Status(500).JSON(ErrorResponse{Error: "Failed to create invoice"})
    }
    
    return c.Status(201).JSON(SuccessResponse{Data: invoice})
}

func (h *SupplierInvoiceHandler) Approve(c *fiber.Ctx) error {
    invoiceID := c.Params("id")
    userID := c.Locals("user_id").(string)
    
    if err := h.service.ApproveInvoice(c.Context(), invoiceID, userID); err != nil {
        return c.Status(500).JSON(ErrorResponse{Error: "Failed to approve"})
    }
    
    return c.JSON(SuccessResponse{Message: "Invoice approved"})
}
```

---

### Common Patterns Library

#### Approval Workflow Pattern (Temporal)
```go
func ApprovalWorkflow(ctx workflow.Context, req ApprovalRequest) error {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 5 * time.Minute,
    }
    ctx = workflow.WithActivityOptions(ctx, ao)
    
    // Determine required approvers
    var approvers []string
    err := workflow.ExecuteActivity(ctx, "DetermineApprovers", req).Get(ctx, &approvers)
    if err != nil {
        return err
    }
    
    // Request approval from each approver sequentially
    for _, approverID := range approvers {
        err = workflow.ExecuteActivity(ctx, "NotifyApprover", approverID, req).Get(ctx, nil)
        if err != nil {
            return err
        }
        
        // Wait for approval with timeout
        var approved bool
        selector := workflow.NewSelector(ctx)
        
        approvalChan := workflow.GetSignalChannel(ctx, fmt.Sprintf("approve-%s", approverID))
        rejectionChan := workflow.GetSignalChannel(ctx, fmt.Sprintf("reject-%s", approverID))
        
        timer := workflow.NewTimer(ctx, 24*time.Hour) // Timeout after 24 hours
        
        selector.AddReceive(approvalChan, func(c workflow.ReceiveChannel, more bool) {
            c.Receive(ctx, nil)
            approved = true
        })
        
        selector.AddReceive(rejectionChan, func(c workflow.ReceiveChannel, more bool) {
            approved = false
        })
        
        selector.AddFuture(timer, func(f workflow.Future) {
            // Timeout - escalate
            workflow.ExecuteActivity(ctx, "EscalateApproval", approverID, req).Get(ctx, nil)
        })
        
        selector.Select(ctx)
        
        if !approved {
            return errors.New("approval rejected")
        }
    }
    
    return nil
}
```

#### Saga Pattern (Temporal)
```go
func OrderSagaWorkflow(ctx workflow.Context, order *Order) error {
    ao := workflow.ActivityOptions{StartToCloseTimeout: 30 * time.Second}
    ctx = workflow.WithActivityOptions(ctx, ao)
    
    compensations := []func() error{}
    
    // Step 1: Reserve inventory
    err := workflow.ExecuteActivity(ctx, "ReserveInventory", order).Get(ctx, nil)
    if err != nil {
        return err
    }
    compensations = append(compensations, func() error {
        return workflow.ExecuteActivity(ctx, "ReleaseInventory", order).Get(ctx, nil)
    })
    
    // Step 2: Authorize payment
    err = workflow.ExecuteActivity(ctx, "AuthorizePayment", order).Get(ctx, nil)
    if err != nil {
        executeCompensations(ctx, compensations)
        return err
    }
    compensations = append(compensations, func() error {
        return workflow.ExecuteActivity(ctx, "VoidPayment", order).Get(ctx, nil)
    })
    
    // Step 3: Create shipment
    var shipmentID string
    err = workflow.ExecuteActivity(ctx, "CreateShipment", order).Get(ctx, &shipmentID)
    if err != nil {
        executeCompensations(ctx, compensations)
        return err
    }
    
    return nil
}

func executeCompensations(ctx workflow.Context, compensations []func() error) {
    for i := len(compensations) - 1; i >= 0; i-- {
        compensations[i]()
    }
}
```

---

### Extension Architecture

#### Custom Fields
```sql
-- Custom fields stored in JSONB
ALTER TABLE supplier_invoices ADD COLUMN custom_fields JSONB;

-- Schema supports rendering custom fields
-- Service layer handles custom field validation
```

#### Custom Workflows
```go
// Extension hooks
type ExtensionRegistry struct {
    hooks map[string][]ExtensionHook
}

func (r *ExtensionRegistry) RegisterHook(event string, hook ExtensionHook) {
    r.hooks[event] = append(r.hooks[event], hook)
}

func (r *ExtensionRegistry) CallHooks(ctx context.Context, event string, data interface{}) error {
    for _, hook := range r.hooks[event] {
        if err := hook.Execute(ctx, data); err != nil {
            return err
        }
    }
    return nil
}

// Usage in workflow
func ProcessInvoiceWithExtensions(ctx workflow.Context, invoice *Invoice) error {
    // Call extension hooks
    err := workflow.ExecuteActivity(ctx, "CallExtensionHooks", "invoice.before_process", invoice).Get(ctx, nil)
    if err != nil {
        return err
    }
    
    // Standard processing
    // ...
    
    return nil
}
```

---

## Conclusion

This documentation provides a complete view of Awo ERP:

1. **Business modules** define what the system does
2. **Technical architecture** explains how it's built
3. **Implementation guide** shows how to build it
4. **Temporal integration** manages complex workflows
5. **JSON schemas** enable rapid UI development

The system is designed for:
- **Scalability**: Multi-tenant architecture with Temporal workflows
- **Maintainability**: Clean separation of concerns
- **Security**: Database-level tenant isolation
- **Productivity**: Schema-driven UI reduces development time
- **Reliability**: Temporal ensures durable execution of critical processes

Use this as your reference for understanding, building, and extending Awo ERP.
