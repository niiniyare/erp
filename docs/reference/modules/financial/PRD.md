# Financial Module - Complete Business Domain Guide

> **Comprehensive guide covering business concepts, setup, operations, and technical specifications for the AWO ERP Financial Module.**

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Why Financial Management in ERP](#why-financial-management-in-erp)
3. [Core Accounting Principles](#core-accounting-principles)
4. [Initial Setup & Configuration](#initial-setup--configuration)
5. [Chart of Accounts Management](#chart-of-accounts-management)
6. [Transaction Processing](#transaction-processing)
7. [Financial Period Management](#financial-period-management)
8. [Multi-Currency Operations](#multi-currency-operations)
9. [Cost Center & Budget Management](#cost-center--budget-management)
10. [Module Integration Points](#module-integration-points)
11. [Financial Reporting Framework](#financial-reporting-framework)
12. [Approval Workflows](#approval-workflows)
13. [Compliance & Regulatory Features](#compliance--regulatory-features)
14. [Common Business Scenarios](#common-business-scenarios)
15. [Troubleshooting Guide](#troubleshooting-guide)
16. [Business Rules & Validation](#business-rules--validation)

---

## Executive Summary

### Module Purpose

The Financial Module is the backbone of the AWO ERP system, providing comprehensive accounting and financial management capabilities that ensure:

- **Accurate Record-Keeping**: Double-entry bookkeeping enforced at all levels
- **Real-Time Visibility**: Instant access to financial position and performance
- **Regulatory Compliance**: Built-in controls for GAAP, IFRS, and SOX compliance
- **Multi-Entity Support**: Manage multiple legal entities with consolidated reporting
- **Currency Flexibility**: Full multi-currency transaction and reporting capabilities

### Why Use Integrated ERP Accounting?

Accounting is mandatory for every business, but it can be time-consuming and error-prone when managed through disparate systems. An integrated ERP financial module delivers:

**Automatic Integration Benefits:**
- Sales transactions automatically create receivables and revenue entries
- Purchase orders flow into payables and expense recognition
- Inventory movements update cost of goods sold in real-time
- Payroll transactions post to expense accounts and liabilities
- Fixed asset acquisitions trigger depreciation schedules

**Efficiency Gains:**
- Eliminate manual data entry between systems
- Reduce reconciliation time by 70-80%
- No more searching across spreadsheets and applications
- Single source of truth for all financial data
- Everyone works with the same up-to-date information

**Error Reduction:**
- Built-in validation prevents unbalanced entries
- Automated calculations eliminate math errors
- Controlled workflows ensure proper approvals
- Complete audit trail of all changes
- Segregation of duties enforced systemically

**Enhanced Reporting:**
- Real-time financial statements available 24/7
- Drill-down capability from summary to transaction detail
- Multi-dimensional analysis (by department, project, product)
- Comparative reporting across periods and entities
- Customizable dashboards for different user roles

### Key Stakeholders

**Finance Team:**
- Daily transaction processing and data entry
- Bank reconciliation and cash management
- Accounts receivable and payable management
- Month-end close procedures

**Controllers:**
- Financial statement preparation
- Period close oversight and validation
- Intercompany reconciliation
- Variance analysis and reporting

**CFO/Finance Leadership:**
- Strategic financial analysis
- Budget planning and monitoring
- Cash flow forecasting
- Board and stakeholder reporting

**Business Unit Leaders:**
- Departmental P&L review
- Budget vs. actual analysis
- Cost management and optimization
- Resource allocation decisions

**Auditors:**
- Compliance verification
- Control testing
- Audit trail review
- Financial statement validation

**Operations Teams:**
- Expense submission and approval
- Project cost tracking
- Budget availability checking
- Purchase requisition processing

### Module Capabilities

**What the Financial Module CAN Do:**
1. Enforce double-entry bookkeeping automatically
2. Process transactions in multiple currencies
3. Generate real-time financial reports
4. Integrate seamlessly with other ERP modules
5. Maintain complete audit trails
6. Support multi-entity operations
7. Automate recurring transactions
8. Calculate depreciation and amortization
9. Manage approval workflows
10. Ensure regulatory compliance

**What the Financial Module CANNOT Do:**
1. Automatically enter transaction data without user input
2. Make business decisions for you
3. Interpret financial results (requires human analysis)
4. Guarantee financial success
5. Replace the need for skilled accounting professionals
6. Automatically reconcile unreconcilable differences

### Success Metrics

Organizations successfully implementing the Financial Module typically achieve:
- **80% reduction** in manual data entry
- **60% faster** month-end close process
- **90% fewer** accounting errors
- **50% reduction** in audit preparation time
- **Real-time** financial visibility (vs. weeks of delay)

---

## Why Financial Management in ERP

### The Integration Advantage

The fundamental value of ERP financial management lies in seamless integration with all business operations. When you operate in a unified system:

**Single Data Entry, Multiple Updates:**
When a sales order is created, the system automatically:
1. Reserves inventory for shipment
2. Prepares for revenue recognition upon delivery
3. Creates accounts receivable upon invoicing
4. Updates sales forecasts and pipeline
5. Triggers commission calculations
6. Reflects in customer credit limit tracking

**Elimination of Reconciliation Headaches:**
- No more comparing sales reports to accounting reports
- Inventory values automatically match general ledger
- Payroll totals reconcile to expense accounts instantly
- Bank transactions match to accounting entries in real-time

**Real-Time Business Intelligence:**
Because all modules feed the same database, you get:
- Current profitability by product, customer, or region
- Real-time cash position and forecasts
- Immediate impact analysis of business decisions
- Comprehensive dashboards without manual data gathering

### Pain Points Solved

**Before Integrated ERP:**
- ❌ Manual entry of the same data in multiple systems
- ❌ Days or weeks to close the books each month
- ❌ Frequent errors from rekeying data
- ❌ Hours spent reconciling different data sources
- ❌ Delayed financial reports (data already outdated)
- ❌ Difficulty tracking down transaction details
- ❌ Different versions of "the truth" across departments
- ❌ Limited ability to drill down into details

**After Integrated ERP:**
- ✅ Single entry point for all transaction data
- ✅ Close the books in hours instead of days
- ✅ Validation rules prevent most errors at entry
- ✅ Reconciliation happens automatically in real-time
- ✅ Financial reports available instantly, always current
- ✅ Complete audit trail from summary to source
- ✅ One version of truth shared across organization
- ✅ Drill-down from any report to transaction detail

### Cross-Module Data Flow

**Sales to Finance:**
```
Quotation → Sales Order → Delivery Note → Sales Invoice → Payment Receipt
    ↓            ↓              ↓              ↓              ↓
  (No GL)    (No GL)      (No GL)       A/R & Revenue    Cash & A/R
```

**Purchasing to Finance:**
```
Purchase Req → Purchase Order → Goods Receipt → Purchase Invoice → Payment
     ↓               ↓                ↓                ↓              ↓
  (No GL)        (No GL)        Inventory (†)    Expense & A/P   Cash & A/P

(†) Or direct expense if non-inventory item
```

**HR/Payroll to Finance:**
```
Timesheet → Payroll Run → Salary Slip → Payment
    ↓            ↓             ↓            ↓
 (No GL)   Salary Accrual  Detailed GL   Cash & Clear
```

This automatic flow eliminates manual journal entries and ensures consistency across all business processes.

---

## Core Accounting Principles

### Double-Entry Bookkeeping Fundamentals

The Financial Module is built on the foundation of double-entry bookkeeping, a system that has stood the test of time for over 500 years. Understanding this principle is essential to using the system effectively.

#### The Fundamental Accounting Equation

Every transaction in the system maintains this eternal balance:

**Assets = Liabilities + Equity**

This equation must always be true. The system enforces this by requiring:
- Every transaction has at least two entries (double-entry)
- Total debits must equal total credits
- No transaction can be posted unless perfectly balanced

#### Account Types & Their Behavior

Think of accounts as buckets where money flows in and out. Each type of bucket has specific rules:

| Account Type | Normal Balance | Increases With | Decreases With | Examples |
|--------------|---------------|----------------|----------------|----------|
| **Assets** | Debit | Debit | Credit | Cash, Receivables, Inventory, Equipment |
| **Liabilities** | Credit | Credit | Debit | Payables, Loans, Accrued Expenses |
| **Equity** | Credit | Credit | Debit | Capital, Retained Earnings, Draws |
| **Revenue** | Credit | Credit | Debit | Sales, Service Income, Interest Income |
| **Expenses** | Debit | Debit | Credit | Salaries, Rent, Utilities, Supplies |

#### Understanding Debits and Credits

**The Golden Rules:**
1. **Assets & Expenses**: Increase with debits, decrease with credits
2. **Liabilities, Equity & Revenue**: Increase with credits, decrease with debits
3. **Every transaction**: Debits must equal credits

**Think of it as Newton's Third Law for Money:**
*"For every financial action, there is an equal and opposite reaction"*

When you purchase equipment for $10,000 cash:
- Equipment (Asset) goes up by $10,000 → **Debit** Equipment
- Cash (Asset) goes down by $10,000 → **Credit** Cash

Both accounts are assets, but one increases while the other decreases, maintaining balance.

#### Practical Double-Entry Examples

**Example 1: Cash Sale**
```
Transaction: Sell goods for $5,000 cash

Dr. Cash (Asset)                    $5,000
    Cr. Sales Revenue (Revenue)             $5,000

Impact: Cash increases, Revenue increases
Balance Check: $5,000 debit = $5,000 credit ✓
```

**Example 2: Purchase on Credit**
```
Transaction: Buy office supplies for $1,200 on account

Dr. Office Supplies Expense (Expense) $1,200
    Cr. Accounts Payable (Liability)          $1,200

Impact: Expense increases, Liability increases
Balance Check: $1,200 debit = $1,200 credit ✓
```

**Example 3: Payment of Liability**
```
Transaction: Pay $3,000 on loan

Dr. Loan Payable (Liability)        $3,000
    Cr. Cash (Asset)                        $3,000

Impact: Liability decreases, Cash decreases
Balance Check: $3,000 debit = $3,000 credit ✓
```

**Example 4: Multi-Line Entry**
```
Transaction: Purchase equipment for $15,000: $5,000 cash, $10,000 loan

Dr. Equipment (Asset)              $15,000
    Cr. Cash (Asset)                         $5,000
    Cr. Loan Payable (Liability)            $10,000

Impact: Equipment up $15k, Cash down $5k, Loan up $10k
Balance Check: $15,000 debit = $15,000 credit ✓
```

### The Accounting Cycle

Financial operations follow a continuous cycle:

```
1. TRANSACTION OCCURS
   ↓
2. DOCUMENT CREATED (Invoice, Receipt, etc.)
   ↓
3. JOURNAL ENTRY RECORDED
   ↓
4. POST TO GENERAL LEDGER
   ↓
5. GENERATE TRIAL BALANCE
   ↓
6. MAKE ADJUSTING ENTRIES
   ↓
7. PREPARE FINANCIAL STATEMENTS
   ↓
8. CLOSE THE PERIOD
   ↓
9. NEXT PERIOD BEGINS
   ↓
(Return to Step 1)
```

The Financial Module automates most of these steps while maintaining proper controls and audit trails.

### Financial Statement Relationships

The three primary financial statements are interconnected:

**Balance Sheet** (Snapshot at a point in time)
- Shows what you own (Assets)
- Shows what you owe (Liabilities)
- Shows owner's stake (Equity)

**Income Statement** (Performance over a period)
- Shows what you earned (Revenue)
- Shows what you spent (Expenses)
- Shows your profit/loss (Net Income)

**Cash Flow Statement** (Cash movements over a period)
- Operating Activities (core business)
- Investing Activities (assets bought/sold)
- Financing Activities (loans, equity)

**The Connection:**
```
Beginning Equity (Balance Sheet)
+ Net Income (Income Statement)
- Owner Draws
= Ending Equity (Balance Sheet)

Beginning Cash (Balance Sheet)
+ Net Cash Flow (Cash Flow Statement)
= Ending Cash (Balance Sheet)
```

---

## Initial Setup & Configuration

### Prerequisites Checklist

Before implementing the Financial Module, gather the following information:

**Company Information:**
- [ ] Legal entity name and registration details
- [ ] Tax identification numbers (VAT, EIN, etc.)
- [ ] Physical and mailing addresses
- [ ] Contact information (phone, email, website)
- [ ] Fiscal year start and end dates
- [ ] Legal structure (corporation, partnership, etc.)

**Banking Information:**
- [ ] All bank account details (account numbers, banks)
- [ ] Credit card accounts
- [ ] Payment gateway information
- [ ] Bank contact information

**Chart of Accounts:**
- [ ] Current chart of accounts (if migrating)
- [ ] Account numbering scheme preference
- [ ] Industry-specific account requirements
- [ ] Cost center/department structure

**Opening Balances:**
- [ ] Trial balance as of go-live date
- [ ] Detailed receivables by customer
- [ ] Detailed payables by supplier
- [ ] Inventory values by location
- [ ] Fixed asset schedules with depreciation
- [ ] Loan schedules and balances

**Currency & Tax:**
- [ ] Base (functional) currency
- [ ] Foreign currencies needed
- [ ] Tax rates and jurisdictions
- [ ] Tax registration numbers

**Organizational Structure:**
- [ ] Department/cost center hierarchy
- [ ] Project list (if applicable)
- [ ] Location/branch information
- [ ] Profit center designations

**Workflow Requirements:**
- [ ] Approval hierarchies
- [ ] Authority limits by role
- [ ] Segregation of duties matrix
- [ ] Document numbering sequences

### Setup Sequence

The order matters! Follow this sequence for a smooth implementation:

```
1. Company Master Setup
   ↓
2. Fiscal Year Definition
   ↓
3. Currency Configuration
   ↓
4. Chart of Accounts
   ↓
5. Cost Centers/Departments
   ↓
6. Opening Balances
   ↓
7. Tax Configuration
   ↓
8. Approval Workflows
   ↓
9. User Roles & Permissions
   ↓
10. Testing & Validation
```

### Step 1: Company Master Setup

The Company master is the foundation of your financial system. Every transaction will reference this data.

**Essential Company Details:**

```markdown
**Basic Information:**
- Company Name: [Legal name exactly as registered]
- Short Name: [Abbreviation for reports]
- Company Abbreviation: [Code for transaction prefixes]
- Default Currency: [Base currency, e.g., KES, USD]
- Country: [Jurisdiction]

**Registration Details:**
- Tax ID / EIN: [Primary tax identification]
- VAT Registration Number: [If applicable]
- Company Registration Number: [Business registration]
- Date of Incorporation: [Legal establishment date]

**Contact Information:**
- Address: [Complete physical address]
- Phone: [Primary contact number]
- Email: [Official email]
- Website: [Company website URL]

**Financial Defaults:**
- Default Finance Book: [If using multiple books]
- Default Bank Account: [Primary operating account]
- Default Cash Account: [Petty cash account]
- Default Receivable Account: [AR control account]
- Default Payable Account: [AP control account]
- Default Income Account: [Default revenue account]
- Default Expense Account: [Default expense account]
- Cost Center: [Default cost center]

**Operational Settings:**
- Standard Working Hours: [e.g., 40 hours/week]
- Holiday List: [Annual holiday calendar]
- Default Letterhead: [For printed documents]
- Default Terms & Conditions: [Standard T&C]
```

**⚠️ Critical Setup Rules:**
1. **Get it right the first time**: Changing company basics after go-live is difficult
2. **Use legal names**: Must match government registration exactly
3. **Validate tax IDs**: Incorrect tax numbers cause compliance issues
4. **Default accounts must exist**: Create chart of accounts before setting defaults

### Step 2: Fiscal Year Configuration

**What is a Fiscal Year?**
A fiscal year is the 12-month period your company uses for accounting and tax purposes. It doesn't have to match the calendar year.

**Common Fiscal Year Examples:**
- **Calendar Year**: January 1 - December 31
- **UK Tax Year**: April 6 - April 5
- **US Federal Fiscal Year**: October 1 - September 30
- **July-June**: July 1 - June 30 (common in many countries)

**Configuration Fields:**
```markdown
- Fiscal Year Name: [e.g., "2025", "FY2024-25"]
- Start Date: [First day of fiscal year]
- End Date: [Last day of fiscal year]
- Is Closed: [No - until year-end close complete]
```

**Accounting Periods Within Fiscal Year:**

While the fiscal year is typically 12 months, you can define shorter accounting periods for reporting:

```markdown
**Monthly Periods:** (Most common)
- Period 1: January 2025
- Period 2: February 2025
- ... 
- Period 12: December 2025

**Quarterly Periods:**
- Q1 2025: Jan-Mar 2025
- Q2 2025: Apr-Jun 2025
- Q3 2025: Jul-Sep 2025
- Q4 2025: Oct-Dec 2025
```

**Period Controls:**
- **Open Period**: Allows transaction posting
- **Closed Period**: Prevents new transactions (except reversals with permissions)
- **Locked Period**: Completely locked, no changes allowed

**Best Practices:**
1. Set up the current fiscal year during implementation
2. Add next fiscal year before current year ends
3. Close periods monthly after reconciliation
4. Keep minimum 2 periods open (current + prior for corrections)
5. Document who has authority to reopen periods

### Step 3: Currency Configuration

**Base (Functional) Currency:**
Your company's primary operating currency. All financial statements will ultimately be presented in this currency.

```markdown
**Example for Kenyan Company:**
- Base Currency: KES (Kenyan Shilling)
- Symbol: KSh
- Decimal Places: 2
```

**Foreign Currency Setup:**

If you transact in multiple currencies, configure each one:

```markdown
**Currency Master:**
- Currency Code: USD
- Currency Name: US Dollar
- Symbol: $
- Fraction: Cents
- Fraction Units: 100
- Smallest Unit: 0.01
```

**Exchange Rate Configuration:**

```markdown
**Exchange Rate Types:**
1. **Spot Rate**: Current market rate
2. **Average Rate**: Period average (for P&L items)
3. **Historical Rate**: Rate at transaction date
4. **Budget Rate**: Planning rate for forecasting

**Rate Entry Methods:**
- Manual entry by authorized users
- Automatic import from financial data providers
- API integration with forex services
- Scheduled batch updates

**Example Rate Entry:**
Date: 2025-01-13
From Currency: USD
To Currency: KES
Rate: 129.50
Rate Type: Spot
```

**Multi-Currency Best Practices:**
1. Load exchange rates daily for active currencies
2. Use average rates for income statement translation
3. Use spot rates for balance sheet items
4. Maintain rate history for audit purposes
5. Set up automatic rate updates where possible

### Step 4: Chart of Accounts Setup

**What is a Chart of Accounts?**

The Chart of Accounts (COA) is the complete list of all accounts used to record transactions. Think of it as the filing system for your company's financial data.

**Account Numbering Schemes:**

Choose a consistent numbering structure before you start:

**Example 1: Traditional Numbering (5-digit)**
```
1000-1999: Assets
  1000-1499: Current Assets
    1100-1199: Cash & Equivalents
    1200-1299: Accounts Receivable
    1300-1399: Inventory
  1500-1999: Non-Current Assets
    1500-1599: Property & Equipment
    1600-1699: Intangible Assets

2000-2999: Liabilities
  2000-2499: Current Liabilities
    2100-2199: Accounts Payable
    2200-2299: Accrued Expenses
  2500-2999: Long-Term Liabilities
    2500-2599: Long-Term Debt

3000-3999: Equity
  3100-3199: Capital
  3900-3999: Retained Earnings

4000-4999: Revenue
  4100-4199: Product Sales
  4200-4299: Service Revenue

5000-5999: Cost of Sales

6000-8999: Operating Expenses
  6000-6999: Selling Expenses
  7000-7999: Administrative Expenses
  8000-8999: Other Expenses

9000-9999: Other Income/Expense
```

**Example 2: Segment-Based Numbering**
```
AAAA-BBB-CC

AAAA = Account Type
  1000 = Assets
  2000 = Liabilities
  etc.

BBB = Sub-Category
  100 = Cash
  200 = AR
  etc.

CC = Detail
  01, 02, 03...
```

**Sample Starter Chart of Accounts:**

```markdown
ASSETS (1000-1999)

Current Assets (1000-1499)
├── Cash & Cash Equivalents (1100-1199)
│   ├── 1110 - Petty Cash
│   ├── 1120 - Checking Account - Main
│   ├── 1121 - Checking Account - Payroll
│   ├── 1130 - Savings Account
│   └── 1140 - Money Market Account
│
├── Accounts Receivable (1200-1299)
│   ├── 1210 - Accounts Receivable - Trade
│   ├── 1220 - Accounts Receivable - Other
│   ├── 1250 - Allowance for Doubtful Accounts (contra-asset)
│   └── 1260 - Employee Advances
│
├── Inventory (1300-1399)
│   ├── 1310 - Raw Materials
│   ├── 1320 - Work in Progress
│   ├── 1330 - Finished Goods
│   └── 1340 - Goods in Transit
│
└── Other Current Assets (1400-1499)
    ├── 1410 - Prepaid Insurance
    ├── 1420 - Prepaid Rent
    └── 1430 - VAT Receivable

Non-Current Assets (1500-1999)
├── Property, Plant & Equipment (1500-1699)
│   ├── 1510 - Land
│   ├── 1520 - Buildings
│   ├── 1525 - Accumulated Depreciation - Buildings (contra)
│   ├── 1530 - Machinery & Equipment
│   ├── 1535 - Accumulated Depreciation - Equipment (contra)
│   ├── 1540 - Furniture & Fixtures
│   ├── 1545 - Accumulated Depreciation - Furniture (contra)
│   ├── 1550 - Vehicles
│   └── 1555 - Accumulated Depreciation - Vehicles (contra)
│
├── Intangible Assets (1700-1799)
│   ├── 1710 - Software
│   ├── 1715 - Accumulated Amortization - Software (contra)
│   ├── 1720 - Patents & Trademarks
│   └── 1730 - Goodwill
│
└── Investments (1800-1899)
    ├── 1810 - Long-Term Investments
    └── 1820 - Investment in Subsidiaries

LIABILITIES (2000-2999)

Current Liabilities (2000-2499)
├── Accounts Payable (2100-2199)
│   ├── 2110 - Accounts Payable - Trade
│   └── 2120 - Accounts Payable - Other
│
├── Accrued Expenses (2200-2299)
│   ├── 2210 - Accrued Salaries
│   ├── 2220 - Accrued Interest
│   ├── 2230 - Accrued Utilities
│   └── 2240 - Accrued Rent
│
├── Taxes Payable (2300-2399)
│   ├── 2310 - VAT Payable
│   ├── 2320 - Income Tax Payable
│   ├── 2330 - Payroll Tax Payable
│   └── 2340 - Withholding Tax Payable
│
└── Short-Term Debt (2400-2499)
    ├── 2410 - Short-Term Loans
    ├── 2420 - Current Portion of Long-Term Debt
    └── 2430 - Credit Card Payable

Long-Term Liabilities (2500-2999)
├── 2510 - Long-Term Loans
├── 2520 - Bonds Payable
├── 2530 - Deferred Tax Liability
└── 2540 - Lease Obligations

EQUITY (3000-3999)
├── 3100 - Owner's Capital
├── 3200 - Additional Paid-In Capital
├── 3300 - Retained Earnings
├── 3400 - Current Year Earnings
├── 3500 - Owner's Draws
└── 3600 - Treasury Stock (contra-equity)

REVENUE (4000-4999)
├── Product Revenue (4100-4299)
│   ├── 4110 - Product Sales - Category A
│   ├── 4120 - Product Sales - Category B
│   └── 4190 - Sales Returns & Allowances (contra-revenue)
│
├── Service Revenue (4300-4499)
│   ├── 4310 - Consulting Services
│   ├── 4320 - Maintenance Services
│   └── 4330 - Subscription Revenue
│
└── Other Revenue (4500-4999)
    ├── 4510 - Interest Income
    ├── 4520 - Dividend Income
    ├── 4530 - Rental Income
    └── 4540 - Foreign Exchange Gain

COST OF SALES (5000-5999)
├── 5100 - Cost of Goods Sold - Products
├── 5200 - Cost of Services
├── 5300 - Direct Labor
├── 5400 - Manufacturing Overhead
└── 5500 - Freight In

OPERATING EXPENSES (6000-8999)

Selling Expenses (6000-6999)
├── 6100 - Salaries - Sales
├── 6200 - Commissions
├── 6300 - Advertising & Marketing
├── 6400 - Travel & Entertainment - Sales
└── 6500 - Shipping & Delivery

Administrative Expenses (7000-7999)
├── Payroll (7100-7299)
│   ├── 7110 - Salaries - Administrative
│   ├── 7120 - Employee Benefits
│   ├── 7130 - Payroll Taxes
│   └── 7140 - Workers Compensation
│
├── Facility Costs (7300-7499)
│   ├── 7310 - Rent
│   ├── 7320 - Utilities
│   ├── 7330 - Property Insurance
│   ├── 7340 - Repairs & Maintenance
│   └── 7350 - Security
│
├── Office Expenses (7500-7699)
│   ├── 7510 - Office Supplies
│   ├── 7520 - Postage & Shipping
│   ├── 7530 - Telephone & Internet
│   ├── 7540 - Printing & Copying
│   └── 7550 - Subscriptions & Dues
│
└── Professional Fees (7700-7899)
    ├── 7710 - Legal Fees
    ├── 7720 - Accounting Fees
    ├── 7730 - Consulting Fees
    └── 7740 - Bank Charges

Depreciation & Amortization (8000-8999)
├── 8100 - Depreciation Expense
└── 8200 - Amortization Expense

OTHER INCOME & EXPENSES (9000-9999)
├── 9100 - Interest Expense
├── 9200 - Loss on Asset Disposal
├── 9300 - Foreign Exchange Loss
└── 9900 - Income Tax Expense
```

**Account Properties Configuration:**

For each account, configure:

```markdown
**Required Fields:**
- Account Code: [Unique identifier]
- Account Name: [Descriptive name]
- Root Type: [Asset, Liability, Equity, Revenue, Expense]
- Account Type: [More specific classification]

**Control Flags:**
- Is Group: [Yes = has children, No = leaf account]
- Parent Account: [If child account]
- Is Active: [Controls whether account accepts transactions]
- Is System Account: [Protected system accounts]
- Allow Manual Entries: [Can users post directly?]
- Require Reference: [Must have reference number?]

**Financial Statement Mapping:**
- Report Type: [Balance Sheet, P&L, Cash Flow]
- Report Group: [Where it appears on statements]
```

**Setup Process:**

**Option 1: Import from Template**
1. Download industry-specific COA template
2. Customize account names and numbers
3. Validate parent-child relationships
4. Import via CSV or Excel
5. Review and activate accounts

**Option 2: Manual Creation**
1. Create top-level groups (Assets, Liabilities, etc.)
2. Add second-level groups (Current Assets, Fixed Assets, etc.)
3. Add leaf accounts under each group
4. Set account properties
5. Validate hierarchy

**⚠️ Critical COA Rules:**

1. **Account codes are immutable**: Once transactions exist, codes cannot be changed
2. **Only leaf accounts accept transactions**: Parent/group accounts cannot have direct entries
3. **Maintain hierarchy integrity**: Accounts inherit properties from parents
4. **Reserve number ranges**: Leave gaps for future accounts
5. **Document your scheme**: Future staff need to understand the logic

### Step 5: Cost Center Structure

**What is a Cost Center?**

A cost center is an organizational unit where costs and revenues can be tracked separately. Think of cost centers as "buckets" within your organization that help answer: "Where did we spend money?" and "Which part of the business generated this income?"

**Common Cost Center Structures:**

**By Department:**
```
Company
├── Sales & Marketing
│   ├── Inside Sales
│   └── Field Sales
├── Operations
│   ├── Manufacturing
│   ├── Quality Control
│   └── Logistics
├── Administration
│   ├── Finance
│   ├── HR
│   └── IT
└── R&D
```

**By Location:**
```
Company
├── Nairobi Branch
├── Mombasa Branch
├── Kisumu Branch
└── Head Office
```

**By Product Line:**
```
Company
├── Product A Division
├── Product B Division
├── Service Division
└── Corporate Services
```

**Distributed Cost Centers:**

Sometimes costs from one center need to be allocated to others. This is called a distributed cost center.

**Example:**
```
IT Department spends 100,000 KES/month

Allocation:
- Sales Department: 30% → 30,000 KES
- Operations: 50% → 50,000 KES
- Administration: 20% → 20,000 KES

Journal Entry (Automatic):
Dr. IT Expense Allocation - Sales        30,000
Dr. IT Expense Allocation - Operations   50,000
Dr. IT Expense Allocation - Admin        20,000
    Cr. IT Department Costs                     100,000
```

**Cost Center Configuration:**

```markdown
**Cost Center Master:**
- Cost Center Code: [e.g., CC-SALES-001]
- Cost Center Name: [Sales Department]
- Parent Cost Center: [If hierarchical]
- Is Group: [Yes/No - can have sub-centers?]
- Is Distributed: [Yes/No - allocate costs?]

**If Distributed:**
- Allocation Method: [Percentage, Headcount, Square Footage]
- Target Cost Centers: [List with percentages]
  - CC-SALES: 30%
  - CC-OPS: 50%
  - CC-ADMIN: 20%
```

**Using Cost Centers:**

Every transaction can be tagged with a cost center:

```markdown
**Sales Invoice:**
- Customer: ABC Corp
- Amount: 100,000 KES
- Cost Center: CC-SALES-NAIROBI

**Result:** Revenue credited to Sales - Nairobi cost center

**Purchase Invoice:**
- Supplier: Office Supplies Ltd
- Amount: 15,000 KES
- Cost Center: CC-ADMIN-HR

**Result:** Expense charged to HR department
```

**Cost Center Reporting:**

Generate reports showing:
- Income statement by cost center
- Cost center budget vs. actual
- Trend analysis by center
- Profitability by division

### Step 6: Opening Balance Entry

**What are Opening Balances?**

When you switch to the new ERP system, you need to bring forward all existing balances from your previous system. This is called entering opening balances.

**Critical Rule:** Opening balances must be taken from a balanced trial balance as of the go-live date.

**Preparation Checklist:**

```markdown
**Balance Sheet Accounts (as of go-live date):**
- [ ] All bank account balances
- [ ] Customer receivables (detail by customer)
- [ ] Supplier payables (detail by supplier)
- [ ] Inventory values (by item and location)
- [ ] Fixed assets with accumulated depreciation
- [ ] All loan balances
- [ ] Prepaid expenses
- [ ] Accrued expenses
- [ ] Equity accounts

**Verification:**
- [ ] Trial balance from old system
- [ ] Bank reconciliation completed
- [ ] AR aging report matches
- [ ] AP aging report matches
- [ ] Physical inventory count done
- [ ] All accounts reconciled
```

**Opening Balance Entry Methods:**

**Method 1: Single Opening Entry** (Recommended for clean cutover)

```
Transaction Date: Go-Live Date (e.g., 2025-01-01)
Reference: "Opening Balances - Migration from [Old System]"

ASSETS:
Dr. Cash - Main Account              125,000
Dr. Petty Cash                         5,000
Dr. Accounts Receivable              450,000
Dr. Inventory                        320,000
Dr. Equipment                        850,000
Dr. Accumulated Depreciation - Equipment    (340,000) [Credit]

LIABILITIES:
Cr. Accounts Payable                         280,000
Cr. Loan Payable                             400,000
Cr. Accrued Expenses                          35,000

EQUITY:
Cr. Retained Earnings                        695,000
                                    =========  =========
TOTALS:                            1,750,000  1,750,000 ✓
```

**Method 2: Detailed Opening Entries** (Better audit trail)

**a) Bank Accounts:**
```
Dr. Cash - Checking Account - Main    100,000
Dr. Cash - Checking Account - Payroll  25,000
Dr. Petty Cash                          5,000
    Cr. Opening Balance Clearing              130,000
```

**b) Accounts Receivable (by customer):**
```
Dr. AR - Customer A                    50,000
Dr. AR - Customer B                   125,000
Dr. AR - Customer C                    75,000
Dr. AR - Customer D                   200,000
    Cr. Opening Balance Clearing              450,000
```

**c) Inventory (by item):**
```
Dr. Inventory - Product X              80,000
Dr. Inventory - Product Y             150,000
Dr. Inventory - Raw Material Z         90,000
    Cr. Opening Balance Clearing              320,000
```

**d) Fixed Assets:**
```
Dr. Equipment                         850,000
    Cr. Accumulated Depreciation - Equipment  340,000
    Cr. Opening Balance Clearing              510,000
```

**e) Accounts Payable (by supplier):**
```
Dr. Opening Balance Clearing          280,000
    Cr. AP - Supplier 1                        80,000
    Cr. AP - Supplier 2                       120,000
    Cr. AP - Supplier 3                        80,000
```

**f) Other Liabilities:**
```
Dr. Opening Balance Clearing          435,000
    Cr. Loan Payable - Bank                   400,000
    Cr. Accrued Salaries                       25,000
    Cr. Accrued Utilities                      10,000
```

**g) Clear Opening Balance Account:**
```
Dr. Opening Balance Clearing          695,000
    Cr. Retained Earnings                     695,000
```

**Validation Steps:**

After entering opening balances:

```markdown
1. **Run Trial Balance**
   - Verify total debits = total credits
   - Compare to trial balance from old system

2. **Verify Balance Sheet**
   - Assets = Liabilities + Equity
   - Individual account balances match old system

3. **Subsidiary Ledgers**
   - AR subsidiary total = AR control account
   - AP subsidiary total = AP control account
   - Inventory subsidiary = Inventory control

4. **Reconciliations**
   - Bank balances match bank statements
   - AR aging matches customer balances
   - AP aging matches supplier balances

5. **Management Review**
   - CFO/Controller approval
   - Document validation date
   - Archive source documents
```

**⚠️ Common Mistakes to Avoid:**

1. **Entering P&L balances**: Only enter balance sheet accounts (no revenue/expense opening balances)
2. **Imbalanced entries**: Always verify debits = credits
3. **Wrong date**: Use exact go-live date, not approximations
4. **Missing detail**: AR and AP need customer/supplier detail, not just totals
5. **No documentation**: Keep trial balance and reconciliations for audit

### Step 7: Tax Configuration

Configure tax rates and rules applicable to your jurisdiction:

```markdown
**Tax Master Setup:**
- Tax Name: [e.g., "VAT - Standard Rate"]
- Tax Type: [Sales Tax, Value Added Tax, Withholding Tax]
- Tax Rate: [e.g., 16%]
- Account Head: [Tax Payable account]

**Tax Rules:**
- Applicable From/To Dates
- Minimum Taxable Amount
- Tax Category (Goods, Services, etc.)
- Geographic Applicability

**Example - Kenya VAT:**
Name: VAT Standard
Rate: 16%
Account: 2310 - VAT Payable
Category: Output VAT on Sales

Name: VAT Zero Rated
Rate: 0%
Account: 2310 - VAT Payable
Category: Zero-rated supplies

Name: VAT Exempt
Rate: 0% (Not reclaimable)
Category: Exempt supplies
```

### Step 8: Approval Workflow Configuration

Set up approval requirements based on your internal controls:

```markdown
**Approval Rules Examples:**

**Rule 1: Journal Entry Approvals**
- Transaction Type: Journal Entry
- Amount Threshold: > 100,000 KES
- Approver: Finance Manager
- Sequential Approval: Yes

**Rule 2: Sensitive Account Approvals**
- Account: All Cash Accounts
- Approver: CFO
- Override: CEO only

**Rule 3: Payment Approvals**
- Amount < 50,000: Department Manager
- Amount 50,000 - 500,000: Finance Manager
- Amount > 500,000: CFO
- Amount > 1,000,000: CEO
```

### Step 9: User Roles & Permissions

Define who can do what:

```markdown
**Sample Roles:**

**Accounts Payable Clerk:**
- Create purchase invoices
- Record supplier payments
- View AP reports
- Cannot: Delete posted entries, approve own transactions

**Accounts Receivable Clerk:**
- Create sales invoices
- Record customer payments
- Manage customer accounts
- Cannot: Write off bad debts without approval

**Finance Manager:**
- Approve journal entries
- Close accounting periods
- Run all financial reports
- Post adjusting entries

**CFO:**
- All finance manager permissions
- Open closed periods (emergency)
- Configure accounting settings
- Approve high-value transactions
```

### Step 10: Testing & Go-Live Validation

Before going live, conduct thorough testing:

```markdown
**Test Scenarios:**

1. **Transaction Processing:**
   - [ ] Create and post sales invoice
   - [ ] Record customer payment
   - [ ] Create and post purchase invoice
   - [ ] Record supplier payment
   - [ ] Create manual journal entry

2. **Integration Testing:**
   - [ ] Sales order → invoice → payment
   - [ ] Purchase order → receipt → invoice → payment
   - [ ] Payroll run → GL posting

3. **Reporting:**
   - [ ] Trial balance
   - [ ] Balance sheet
   - [ ] Income statement
   - [ ] AR aging
   - [ ] AP aging

4. **Controls:**
   - [ ] Approval workflows trigger correctly
   - [ ] Users cannot exceed permissions
   - [ ] Period close prevents posting
   - [ ] Imbalanced entries rejected

5. **Validation:**
   - [ ] Numbers match old system
   - [ ] All required accounts present
   - [ ] Tax calculations correct
   - [ ] Multi-currency working properly
```

**Go-Live Checklist:**

```markdown
- [ ] All setup completed and verified
- [ ] Opening balances entered and balanced
- [ ] Users trained on new system
- [ ] Old system data archived
- [ ] Support team available
- [ ] Rollback plan documented
- [ ] First month reconciliation plan ready
- [ ] Management sign-off obtained
```

---

## Chart of Accounts Management

### Hierarchical Structure

The chart of accounts supports unlimited nesting with materialized path optimization for performance:

**Account Hierarchy Example:**
```
1000 - ASSETS (Root)
├── 1100 - Current Assets (Group)
│   ├── 1110 - Cash & Cash Equivalents (Group)
│   │   ├── 1111 - Petty Cash (Leaf - can post)
│   │   ├── 1112 - Checking - Main (Leaf - can post)
│   │   └── 1113 - Savings Account (Leaf - can post)
│   └── 1120 - Accounts Receivable (Group)
│       ├── 1121 - Trade Receivables (Leaf - can post)
│       └── 1122 - Other Receivables (Leaf - can post)
└── 1200 - Non-Current Assets (Group)
    └── 1210 - Property, Plant & Equipment (Group)
        ├── 1211 - Land (Leaf - can post)
        └── 1212 - Buildings (Leaf - can post)
```

**Materialized Path Example:**
```
Account: 1113 - Savings Account
Path: /1000/1100/1110/1113
Level: 4
Parent: 1110

This allows efficient queries like:
- "Find all current asset accounts" → WHERE path LIKE '/1000/1100/%'
- "Get account ancestry" → Parse path segments
- "Calculate group totals" → SUM WHERE path LIKE '/1000/1100/%'
```

### Account Properties & Rules

**Essential Properties:**
```markdown
- Code: Unique, immutable identifier
- Name: Human-readable description
- Root Type: Asset, Liability, Equity, Revenue, Expense
- Account Type: More specific classification
- Normal Balance: Debit or Credit
- Parent Account: Hierarchical parent
- Is Group: Has children (cannot post transactions)
- Is Active: Accepts new transactions
- Is System Account: Protected from user modification
```

**Posting Rules:**
1. Only **leaf accounts** (no children) accept transactions
2. **Group accounts** aggregate child balances
3. Balances **roll up** to parents automatically
4. **System accounts** cannot be deleted or renamed
5. **Inactive accounts** hide from user selection but retain history

### Account Types Reference

**Asset Types:**
- Bank
- Cash
- Accounts Receivable
- Stock (Inventory)
- Fixed Asset
- Accumulated Depreciation (contra-asset)
- Prepaid Expense
- Other Current Asset
- Long-Term Investment

**Liability Types:**
- Accounts Payable
- Credit Card
- Tax Payable
- Accrued Expense
- Short-Term Loan
- Long-Term Liability
- Deferred Revenue

**Equity Types:**
- Capital Stock
- Retained Earnings
- Owner's Draw (contra-equity)
- Treasury Stock (contra-equity)
- Other Comprehensive Income

**Revenue Types:**
- Sales Revenue
- Service Revenue
- Other Income
- Sales Discounts (contra-revenue)
- Sales Returns (contra-revenue)

**Expense Types:**
- Cost of Goods Sold
- Operating Expense
- Depreciation
- Interest Expense
- Tax Expense

### Managing the Chart of Accounts

**Adding New Accounts:**
```markdown
1. Determine correct parent account
2. Choose next available code in range
3. Set account type and properties
4. Activate account
5. Update user permissions if needed
```

**Modifying Accounts:**
```markdown
**Can Change:**
- Account name
- Active/inactive status
- Control flags (allow manual entry, require reference)

**Cannot Change:**
- Account code (immutable)
- Root type (would break financial statements)
- Parent if transactions exist (orphaned data)
```

**Merging/Consolidating Accounts:**
```markdown
If you need to combine accounts:
1. Create journal entry to transfer balance
2. Inactivate old account
3. Update recurring transactions to new account
4. Document the change for audit purposes
```

**Industry-Specific Templates:**

The system provides starter COA templates:
- Retail/Wholesale
- Manufacturing
- Service Business
- Non-Profit Organization
- Construction
- Healthcare
- Hospitality
- Technology/SaaS

Each template includes industry-standard accounts configured with appropriate types and structure.

---

## Transaction Processing

### Transaction Lifecycle

Every financial transaction follows a controlled lifecycle:

```mermaid
graph TD
    A[DRAFT] -->|Submit| B[PENDING_APPROVAL]
    A -->|Cancel| C[CANCELLED]
    B -->|Approve| D[APPROVED]
    B -->|Reject| E[REJECTED]
    D -->|Post| F[POSTED]
    D -->|Cancel| C
    F -->|Reverse| G[REVERSED]
    E -->|Edit| A
```

**State Definitions:**

| State | Editable | In GL | Can Approve | Can Post | Can Cancel |
|-------|----------|-------|-------------|----------|------------|
| DRAFT | Yes | No | No | Yes* | Yes |
| PENDING_APPROVAL | No | No | Yes | No | Yes** |
| APPROVED | No | No | No | Yes | Yes |
| POSTED | No | Yes | No | No | No*** |
| REJECTED | Yes | No | No | No | Yes |
| CANCELLED | No | No | No | No | No |
| REVERSED | No | Yes | No | No | No |

\* Direct post only if approval not required  
\** Requires approval permission  
\*** Can only reverse, cannot cancel

### Transaction Types

**1. MANUAL Transactions**
User-created entries for recording business events:
- General journal entries
- Adjusting entries
- Correcting entries
- Reclassifications

**2. SYSTEM Transactions**
Automatically generated by the system:
- Depreciation calculations
- Accrual reversals
- Period-end adjustments
- Automatic allocations

**3. IMPORTED Transactions**
Data from external sources:
- Bank statement imports
- Third-party integrations
- Bulk uploads via CSV/Excel

**4. RECURRING Transactions**
Template-based repeating entries:
- Monthly rent
- Loan payments
- Depreciation schedules
- Subscription expenses

**5. INTEGRATION Transactions**
Generated by other modules:
- Sales invoices (from Sales module)
- Purchase invoices (from Purchasing)
- Payroll entries (from HR)
- Inventory movements (from Inventory)

### Creating Manual Journal Entries

**Standard Journal Entry Structure:**

```markdown
**Header Information:**
- Transaction Date: Date event occurred
- Posting Date: Date to post to GL (usually same as transaction date)
- Reference Number: External reference (invoice #, check #)
- Description: Purpose of entry
- Transaction Type: MANUAL
- Cost Center: (optional) Department/division
- Project: (optional) Project code

**Entry Lines:** (Minimum 2)
- Account: From chart of accounts
- Debit Amount: If debit entry
- Credit Amount: If credit entry
- Description: Line-specific notes
- Cost Center: (optional) Can override header
- Reference: (optional) Supporting document

**Validation:**
- Total Debits must equal Total Credits
- All accounts must be active
- All accounts must allow manual entries
- At least one debit and one credit required
```

**Example: Record Rent Payment**

```
Date: 2025-01-05
Reference: Check #1234
Description: Office rent payment for January 2025

Account                          Debit      Credit
-----------------------------------------------------
7310 - Rent Expense            50,000
1120 - Checking Account - Main            50,000
-----------------------------------------------------
TOTALS:                        50,000     50,000 ✓
```

### Common Journal Entry Scenarios

**Scenario 1: Prepaid Expense**

When you pay for insurance upfront for 12 months:

```
**Initial Payment (Jan 1):**
Dr. 1410 - Prepaid Insurance       120,000
    Cr. 1120 - Cash                        120,000
Description: Annual insurance premium paid

**Monthly Expense Recognition (each month):**
Dr. 7330 - Insurance Expense        10,000
    Cr. 1410 - Prepaid Insurance           10,000
Description: Insurance expense for [Month]
```

**Scenario 2: Accrued Expense**

When you receive utility service but haven't been billed:

```
**Month-End Accrual:**
Dr. 7320 - Utilities Expense        15,000
    Cr. 2230 - Accrued Utilities           15,000
Description: Utility expense for December (not yet billed)

**When Bill Arrives (next month):**
Dr. 2230 - Accrued Utilities        15,000
    Cr. 2110 - Accounts Payable           15,000
Description: Utility bill received - clear accrual
```

**Scenario 3: Depreciation**

Recording monthly depreciation on equipment:

```
Dr. 8100 - Depreciation Expense     8,500
    Cr. 1535 - Accumulated Depreciation - Equipment  8,500
Description: Monthly depreciation on manufacturing equipment
```

**Scenario 4: Correction Entry**

Fixing an error where expense was coded to wrong account:

```
**Original Entry (WRONG):**
Dr. 7510 - Office Supplies          5,000
    Cr. Cash                                5,000

**Correction Entry:**
Dr. 6300 - Advertising Expense      5,000
    Cr. 7510 - Office Supplies             5,000
Description: Reclassify - expense was for marketing materials
```

**Scenario 5: Bank Reconciliation Adjustments**

Recording bank fees not yet in books:

```
Dr. 7740 - Bank Charges                25
    Cr. 1120 - Checking Account - Main     25
Description: Monthly bank service fee per statement
```

### Transaction Validation Rules

**Financial Validation:**
1. ✅ Debits must equal credits
2. ✅ Minimum 2 transaction lines required
3. ✅ Each line must have debit OR credit (not both or neither)
4. ✅ Amounts must be positive numbers
5. ✅ Currency must be consistent

**Date Validation:**
1. ✅ Transaction date cannot be in future
2. ✅ Transaction date must be in open period
3. ✅ Posting date required when status = POSTED
4. ✅ Due date (if applicable) must be >= transaction date

**Account Validation:**
1. ✅ All accounts must exist and be active
2. ✅ Accounts must allow manual entries (if MANUAL transaction)
3. ✅ Accounts must be leaf accounts (no children)
4. ✅ Reference required if account has "Require Reference" flag

**Business Logic Validation:**
1. ✅ User has permission to post to selected accounts
2. ✅ Cost center is valid (if specified)
3. ✅ Project is active (if specified)
4. ✅ Approval obtained if required

### Posting Process

**What Happens When You Post:**

```markdown
1. **Validation:**
   - All validation rules checked
   - Approval status verified
   - Period open status confirmed

2. **General Ledger Update:**
   - Entries added to GL with posting date
   - Account balances updated
   - Parent account balances rolled up

3. **Subsidiary Ledgers:**
   - Customer balances updated (AR entries)
   - Supplier balances updated (AP entries)
   - Cost center balances updated

4. **Status Changes:**
   - Transaction status → POSTED
   - Posting date timestamp recorded
   - Posted by user captured

5. **Triggers:**
   - Budget checking (if enabled)
   - Alert notifications
   - Report cache invalidation
   - Audit log entry

6. **Immutability:**
   - Entry becomes read-only
   - Can only be reversed, not edited
   - Audit trail locked
```

**Post vs. Submit vs. Approve:**

```markdown
**SUBMIT:**
- Moves DRAFT → PENDING_APPROVAL
- Required if approval workflows active
- Entry frozen, awaiting approval

**APPROVE:**
- Moves PENDING_APPROVAL → APPROVED
- Requires appropriate permissions
- Entry ready for posting

**POST:**
- Moves to POSTED status
- Updates general ledger
- Entry becomes immutable
- Affects financial reports
```

### Reversing Entries

**When to Reverse (vs. Delete):**
- ✅ Entry is already POSTED
- ✅ Need audit trail of correction
- ✅ Original entry was correct at the time
- ✅ Situation has changed since posting

**Reversal Process:**

```markdown
1. Original Entry (Posted 2025-01-15):
   Dr. Equipment                  100,000
       Cr. Cash                           100,000

2. Reversal Entry (Created 2025-01-20):
   Transaction Date: 2025-01-20
   Reference: REV-12345 (links to original)
   
   Dr. Cash                       100,000
       Cr. Equipment                      100,000
   
   Description: "Reversal of entry #12345"

3. Results:
   - Both entries remain in GL
   - Net effect is zero
   - Complete audit trail maintained
   - Original entry marked as REVERSED
```

**Automatic Reversal:**

System can auto-reverse entries on specified date:

```markdown
**Example: Month-End Accrual**

Entry Date: 2025-01-31
Dr. Utilities Expense              10,000
    Cr. Accrued Utilities                  10,000

Reversal Date: 2025-02-01 (automatic)
Dr. Accrued Utilities              10,000
    Cr. Utilities Expense                  10,000

Result: Expense recognized in January, automatically cleared in February
```

---

## Financial Period Management

### Fiscal Year vs. Accounting Period

**Fiscal Year:**
- 12-month period for annual reporting
- Defined by government or company policy
- Examples: Jan-Dec, Apr-Mar, Jul-Jun
- Usually cannot have transactions span multiple fiscal years

**Accounting Period:**
- Shorter periods within fiscal year
- Usually monthly or quarterly
- Used for interim reporting
- Controls when transactions can be posted

### Period Control States

**OPEN Period:**
- Accepts all transaction types
- Users can create and post entries
- Default state for current period

**SOFT CLOSE Period:**
- Prevents regular users from posting
- Finance team can still make entries
- Used during month-end close process
- Allows corrections while preventing new activity

**HARD CLOSE Period:**
- No new transactions allowed
- Only reversals permitted (with special permission)
- Used after month-end close complete
- Can be reopened only by authorized users

**LOCKED Period:**
- Completely frozen
- No transactions whatsoever
- Used after audit or year-end
- Typically requires CEO/CFO to unlock

### Period Close Process

**Standard Month-End Close Workflow:**

```markdown
**Week Before Close:**
- [ ] Notify all departments of cutoff date
- [ ] Request accrual information
- [ ] Prepare recurring entry schedules
- [ ] Review outstanding items list

**Last Day of Month:**
- [ ] Final transaction cutoff (e.g., 5 PM)
- [ ] Process all pending invoices and payments
- [ ] Soft close period to prevent new entries

**Days 1-3 of New Month:**
- [ ] Bank reconciliations for all accounts
- [ ] AR aging review and reconciliation
- [ ] AP aging review and reconciliation
- [ ] Inventory reconciliation (if applicable)

**Days 3-5:**
- [ ] Post accrual entries
- [ ] Record depreciation
- [ ] Process recurring entries
- [ ] Inter-company reconciliations

**Days 5-7:**
- [ ] Run preliminary financial statements
- [ ] Review for unusual items
- [ ] Investigate variances from budget
- [ ] Post correcting entries

**Days 7-10:**
- [ ] Final review with management
- [ ] Prepare financial package
- [ ] Hard close period
- [ ] Document close exceptions

**By Day 10:**
- [ ] Distribute financial reports
- [ ] Management meeting
- [ ] File supporting documentation
- [ ] Archive month-end package
```

### Year-End Close Process

Year-end close is more comprehensive than month-end:

```markdown
**Preparation (Month 12):**
- [ ] Review all accounts for proper classification
- [ ] Reconcile all balance sheet accounts
- [ ] Confirm all prepaid/accrued items
- [ ] Physical inventory count
- [ ] Fixed asset verification
- [ ] Review long-term liabilities
- [ ] Confirm equity transactions

**Year-End Adjustments:**
- [ ] Final depreciation calculations
- [ ] Inventory adjustments
- [ ] Bad debt provisions
- [ ] Tax provision calculations
- [ ] Bonus accruals
- [ ] Warranty reserves

**Closing Entries:**
- [ ] Close all revenue accounts to income summary
- [ ] Close all expense accounts to income summary
- [ ] Close income summary to retained earnings
- [ ] Reset YTD balances

**Post-Close:**
- [ ] Prepare annual financial statements
- [ ] Tax return preparation begins
- [ ] Audit preparation (if applicable)
- [ ] Lock prior year permanently
```

**Example Closing Entries:**

```markdown
**Step 1: Close Revenue Accounts**
Dr. Sales Revenue                 5,000,000
Dr. Interest Income                  50,000
Dr. Other Income                     25,000
    Cr. Income Summary                     5,075,000

**Step 2: Close Expense Accounts**
Dr. Income Summary               4,200,000
    Cr. Cost of Goods Sold                3,000,000
    Cr. Operating Expenses                1,100,000
    Cr. Interest Expense                     50,000
    Cr. Tax Expense                          50,000

**Step 3: Close Income Summary to Retained Earnings**
Dr. Income Summary                 875,000
    Cr. Retained Earnings                    875,000

Net Income for Year: 875,000
```

### Period Reopening

**When to Reopen a Period:**
- Discovery of material error
- Late receipt of vendor invoice
- Audit adjustments required
- Bank reconciliation correction

**Reopening Process:**
```markdown
1. Document reason for reopening
2. Get appropriate authorization (CFO/Controller)
3. Reopen period in system
4. Post correcting entries
5. Review impact on prior reports
6. Re-close period
7. Communicate changes to stakeholders
8. Update any filed reports if necessary
```

**Controls Around Reopening:**
- Requires special user permission
- Audit log of who reopened and why
- Notification to management
- Comparison report (before/after)

---

## Multi-Currency Operations

### Currency Configuration

**Base Currency (Functional Currency):**
The primary currency for your business operations.

```markdown
**Example: Kenyan Company**
Base Currency: KES
- All financial statements in KES
- Default for all transactions
- Equity and retained earnings always in KES
```

**Foreign Currencies:**
Additional currencies for international transactions.

```markdown
**Common Foreign Currencies:**
- USD (US Dollar)
- EUR (Euro)
- GBP (British Pound)
- CNY (Chinese Yuan)
- AED (UAE Dirham)
```

**Currency Master Data:**
```markdown
- Currency Code: USD
- Currency Name: US Dollar
- Symbol: $
- Fraction: Cents
- Fraction Units: 100
- Smallest Currency Unit: 0.01
- Number Format: #,###.##
```

### Exchange Rate Management

**Exchange Rate Types:**

**1. Spot Rate (Current Market Rate):**
- Used for: Most transactions
- Source: Current market exchange rate
- Updated: Daily
- Use: Transaction date rate

**2. Average Rate:**
- Used for: Income statement translation
- Calculation: Period average
- Updated: Monthly
- Use: Revenue and expense translation

**3. Historical Rate:**
- Used for: Equity transactions
- Fixed: Original transaction date
- Updated: Never
- Use: Share capital, investments

**4. Budget/Forecast Rate:**
- Used for: Planning
- Fixed: Set at budget time
- Updated: Annually
- Use: Budgets and forecasts

**Exchange Rate Entry:**

```markdown
**Rate Master:**
Date: 2025-01-13
From Currency: USD
To Currency: KES
Exchange Rate: 129.50
Rate Type: Spot

**Interpretation:**
1 USD = 129.50 KES
OR
1 KES = 0.007722 USD (inverse)
```

**Rate Entry Methods:**

**Manual Entry:**
```markdown
- Finance team enters rates daily
- Sourced from Central Bank or forex provider
- Requires authorization
- Audit trail maintained
```

**Automatic Import:**
```markdown
- API integration with rate provider
- Scheduled import (e.g., 9 AM daily)
- Automatic validation checks
- Alert if rate change > threshold
```

**Multiple Rate Sources:**
```markdown
- Official Central Bank rate
- Commercial bank buying/selling rates
- Market spot rates
- Custom formula (e.g., average of sources)
```

### Foreign Currency Transactions

**Recording FC Transactions:**

```markdown
**Example: Purchase from US Supplier**

Purchase Order Date: 2025-01-10
Amount: $10,000 USD
Exchange Rate: 128.00 KES/USD
Equivalent: 1,280,000 KES

**Journal Entry (Invoice Date):**
Dr. Inventory                    1,280,000 KES
    Cr. Accounts Payable - USD              1,280,000 KES

**System Records Both:**
- Original Currency: $10,000 USD
- Functional Currency: 1,280,000 KES
- Exchange Rate: 128.00
- Transaction Date: 2025-01-10
```

**Payment of FC Liability:**

```markdown
**Payment Date: 2025-02-15**
Amount: $10,000 USD
Exchange Rate: 131.00 KES/USD
KES Required: 1,310,000 KES

**Journal Entry:**
Dr. Accounts Payable - USD       1,280,000 KES (original)
Dr. Foreign Exchange Loss           30,000 KES (difference)
    Cr. Cash - KES                          1,310,000 KES (actual)

**Calculation:**
Original: 10,000 × 128.00 = 1,280,000
Payment: 10,000 × 131.00 = 1,310,000
Loss: 1,310,000 - 1,280,000 = 30,000 KES
```

### Realized vs. Unrealized Gains/Losses

**Unrealized Gain/Loss:**
Paper gain/loss due to rate changes before settlement.

```markdown
**Example:**

**Invoice Date (Jan 10):**
Amount: $5,000 USD
Rate: 128.00
AR Recorded: 640,000 KES

**Month-End (Jan 31):**
Rate: 130.00
Current Value: 5,000 × 130.00 = 650,000 KES

**Unrealized Gain:**
650,000 - 640,000 = 10,000 KES

**Revaluation Entry:**
Dr. Accounts Receivable - USD    10,000
    Cr. Unrealized FX Gain              10,000
Description: Month-end revaluation per current rate

**Important:** 
- Reverses at start of next month
- Not realized until payment received
- Affects balance sheet, not P&L (IFRS)
- Or affects P&L based on accounting policy (GAAP)
```

**Realized Gain/Loss:**
Actual gain/loss when transaction settles.

```markdown
**Continuing Example:**

**Payment Received (Feb 5):**
Amount: $5,000 USD
Rate: 129.00
Cash Received: 5,000 × 129.00 = 645,000 KES

**Reverse January Revaluation:**
Dr. Unrealized FX Gain          10,000
    Cr. Accounts Receivable - USD       10,000

**Record Payment:**
Dr. Cash - KES                   645,000
    Cr. Accounts Receivable - USD       640,000
    Cr. Realized FX Gain                  5,000

**Net Result:**
Original AR: 640,000 KES
Cash Received: 645,000 KES
Realized Gain: 5,000 KES ✓
```

### Foreign Currency Bank Accounts

**Maintaining FC Bank Accounts:**

```markdown
**Bank Account: USD Checking**
- Account Currency: USD
- Tracked Balance: $50,000 USD
- KES Equivalent: Varies daily

**Daily Revaluation:**
Jan 1:  $50,000 × 128.00 = 6,400,000 KES
Jan 31: $50,000 × 130.00 = 6,500,000 KES
Unrealized Gain: 100,000 KES
```

**Bank Reconciliation in FC:**

```markdown
1. Reconcile in USD (original currency)
2. Convert balance to KES using period-end rate
3. Compare to GL balance in KES
4. Post revaluation entry for differences
```

### Multi-Currency Reporting

**Financial Statements in Foreign Currency:**

```markdown
**Translation Methods:**

**Balance Sheet:**
- Assets & Liabilities: Current rate (period-end)
- Equity: Historical rates (original transaction)

**Income Statement:**
- Revenue & Expenses: Average rate for period
- Or actual rates if material

**Cash Flow Statement:**
- Operating Activities: Average rate
- Investing/Financing: Actual transaction rates
```

**Translation Example:**

```markdown
**USD Subsidiary → KES Parent Consolidation**

**Balance Sheet (Dec 31, 2024):**
Item                    USD         Rate        KES
Assets                  100,000     130.00      13,000,000
Liabilities             (40,000)    130.00      (5,200,000)
Equity                   60,000     125.00*     (7,500,000)
Translation Adj                                   (300,000)
                        =======                 ===========

* Historical rate when equity invested

**Income Statement (Year 2024):**
Revenue                 500,000     128.50**    64,250,000
Expenses               (450,000)    128.50**   (57,825,000)
Net Income              50,000                   6,425,000

** Average rate for the year
```

### Multi-Currency Best Practices

**1. Rate Management:**
- Load rates before daily processing begins
- Use consistent rate sources
- Document rate sources for audit
- Set rate change alerts for unusual movements
- Archive historical rates indefinitely

**2. Transaction Processing:**
- Always enter original currency amount
- Let system calculate functional currency
- Review FC transaction reports daily
- Reconcile FC bank accounts weekly

**3. Period Close:**
- Revalue all FC balances at month-end
- Verify significant rate changes
- Review unrealized gain/loss reasonableness
- Document material currency impacts

**4. Reporting:**
- Provide both FC and base currency views
- Show exchange rates used
- Highlight FX impact on results
- Trend analysis by currency

**5. Controls:**
- Limit who can modify exchange rates
- Require approval for manual rate entry
- Alert on rates outside normal ranges
- Segregate duties: rate entry vs. transaction processing

---

## Cost Center & Budget Management

### Cost Center Structure

**What Problem Do Cost Centers Solve?**

Without cost centers, you only know:
- ✓ Total company revenue
- ✓ Total company expenses
- ✓ Overall profit/loss

With cost centers, you know:
- ✓ Revenue by department
- ✓ Expenses by project
- ✓ Profitability by location
- ✓ Cost allocation across the organization
- ✓ Performance by business unit

**Types of Cost Centers:**

**1. Functional Cost Centers:**
```
By Department/Function:
├── Sales & Marketing
├── Research & Development
├── Manufacturing
├── Quality Control
├── Customer Support
├── Finance & Accounting
├── Human Resources
└── Information Technology
```

**2. Geographic Cost Centers:**
```
By Location:
├── Head Office - Nairobi
├── Branch - Mombasa
├── Branch - Kisumu
├── Branch - Eldoret
└── Warehouse - Athi River
```

**3. Product/Service Cost Centers:**
```
By Product Line:
├── Product A Division
├── Product B Division
├── Service Division
└── Corporate/Shared Services
```

**4. Project Cost Centers:**
```
By Project:
├── Project Alpha
├── Project Beta
├── Project Gamma
└── General Operations
```

**Multi-Dimensional Cost Tracking:**

Many businesses need multiple views:

```markdown
**Transaction Tagged With:**
- Primary Cost Center: Sales - Mombasa Branch
- Secondary Dimension: Product A
- Project: Customer Acquisition Campaign Q1

**Enables Analysis:**
- Mombasa branch profitability
- Product A performance across all locations
- Q1 campaign ROI
- Comparison: Mombasa vs. other branches
```

### Distributed Cost Centers

**The Problem:**

Some costs benefit multiple departments. How do you allocate them fairly?

**Example: IT Department Costs**

```markdown
**Monthly IT Costs: 500,000 KES**

IT supports:
- Sales (20 employees, 40% of IT time)
- Operations (15 employees, 30%)
- Administration (10 employees, 20%)
- R&D (5 employees, 10%)

**Allocation:**
- Sales:          500,000 × 40% = 200,000 KES
- Operations:     500,000 × 30% = 150,000 KES
- Administration: 500,000 × 20% = 100,000 KES
- R&D:            500,000 × 10% =  50,000 KES
                                  =========
                                  500,000 KES ✓
```

**Allocation Methods:**

**1. Fixed Percentage:**
```markdown
Allocation Base: Predetermined percentages
Best For: Stable relationships
Example: HR costs by headcount percentage

Configuration:
- IT Department → Sales: 40%
- IT Department → Operations: 30%
- IT Department → Admin: 20%
- IT Department → R&D: 10%
```

**2. Headcount:**
```markdown
Allocation Base: Number of employees
Best For: HR, facilities, general overhead

Example:
Total Employees: 100
- Sales: 30 employees (30%)
- Operations: 40 employees (40%)
- Admin: 20 employees (20%)
- R&D: 10 employees (10%)

Total Cost: 1,000,000
- Sales: 300,000
- Operations: 400,000
- Admin: 200,000
- R&D: 100,000
```

**3. Square Footage:**
```markdown
Allocation Base: Office space occupied
Best For: Rent, utilities, cleaning

Example:
Total Space: 10,000 sq ft
Rent: 2,000,000 KES/month

Department Allocation:
- Sales: 3,000 sq ft → 600,000 KES
- Operations: 4,000 sq ft → 800,000 KES
- Admin: 2,000 sq ft → 400,000 KES
- R&D: 1,000 sq ft → 200,000 KES
```

**4. Usage/Activity-Based:**
```markdown
Allocation Base: Actual consumption
Best For: IT support hours, printing costs

Example: IT Support Tickets
Total Tickets Month: 400
Cost: 800,000 KES

Department Tickets:
- Sales: 160 tickets (40%) → 320,000 KES
- Operations: 120 tickets (30%) → 240,000 KES
- Admin: 80 tickets (20%) → 160,000 KES
- R&D: 40 tickets (10%) → 80,000 KES
```

**Distributed Cost Center Configuration:**

```markdown
**Cost Center Master:**
Name: IT Department
Type: Distributed Cost Center
Status: Active

**Allocation Configuration:**
Allocation Method: Percentage
Frequency: Monthly (automatic)

Target Allocations:
┌─────────────────┬────────────┬────────┐
│ Cost Center     │ Percentage │ Driver │
├─────────────────┼────────────┼────────┤
│ Sales           │ 40%        │ Fixed  │
│ Operations      │ 30%        │ Fixed  │
│ Administration  │ 20%        │ Fixed  │
│ R&D             │ 10%        │ Fixed  │
└─────────────────┴────────────┴────────┘

**Automatic Journal Entry (Month-End):**
Dr. IT Allocation - Sales         200,000
Dr. IT Allocation - Operations    150,000
Dr. IT Allocation - Admin         100,000
Dr. IT Allocation - R&D            50,000
    Cr. IT Department Expense             500,000
```

### Budget Management

**Budget Structure:**

```markdown
**Budget Master:**
- Budget Year: 2025
- Budget Version: V1 (Original), V2 (Revised), etc.
- Status: Draft / Approved / Active
- Currency: KES
- Approval Date: 2024-12-15
- Approved By: CFO

**Budget Dimensions:**
- By Account (required)
- By Cost Center (optional)
- By Project (optional)
- By Month or Quarter
```

**Budget Entry Methods:**

**1. Annual Budget:**
```markdown
Account: 7110 - Salaries
Cost Center: Sales
Annual Budget: 12,000,000 KES

System allocates:
Monthly: 1,000,000 KES
Quarterly: 3,000,000 KES
```

**2. Monthly Budget:**
```markdown
Account: 6300 - Advertising
Cost Center: Marketing

Jan: 500,000  (New Year campaign)
Feb: 200,000
Mar: 200,000
Apr: 800,000  (Product launch)
...
Dec: 300,000
```

**3. Formula-Based Budget:**
```markdown
Account: 6200 - Sales Commissions
Formula: 5% of budgeted revenue

If Revenue Budget = 100,000,000
Then Commission Budget = 5,000,000
```

**Budget vs. Actual Reporting:**

```markdown
**Income Statement - Budget vs. Actual**
Period: January 2025
Cost Center: Sales Department

Account              Budget    Actual    Variance    Var %
──────────────────────────────────────────────────────────
REVENUE
Sales Revenue      5,000,000  5,250,000   250,000   +5.0%

EXPENSES
Salaries           1,000,000  1,050,000   (50,000)  -5.0%
Commissions          250,000    262,500   (12,500)  -5.0%
Travel               150,000    175,000   (25,000)  -16.7% ⚠️
Marketing            500,000    480,000    20,000   +4.0%
Utilities             50,000     52,000    (2,000)  -4.0%
──────────────────────────────────────────────────────────
Total Expenses     1,950,000  2,019,500   (69,500)  -3.6%

NET INCOME         3,050,000  3,230,500   180,500   +5.9% ✓

Favorable Variance: Revenue higher, some expenses lower
Unfavorable: Travel over budget by 16.7% - investigate
```

**Budget Controls:**

**1. Soft Budget Control:**
```markdown
- Warns user when exceeding budget
- Allows override with comment
- Tracks budget exceptions
- Reports to management

Example:
"Warning: This expense will exceed monthly budget by 15,000 KES.
Proceed? [Yes] [No]
Reason: _______________"
```

**2. Hard Budget Control:**
```markdown
- Prevents transactions exceeding budget
- Requires budget increase approval
- Enforced at transaction entry
- No override except by CFO

Example:
"Error: Insufficient budget. 
Available: 25,000 KES
Required: 40,000 KES
Contact Finance to request budget increase."
```

**3. Hierarchical Budget Control:**
```markdown
- Department budget has sub-budgets
- Can exceed sub-budget if total OK
- Flexibility within department
- Control at department level

Example:
Sales Department Budget: 2,000,000
- Salaries: 1,000,000 (limit)
- Travel: 300,000 (guideline)
- Marketing: 500,000 (guideline)
- Other: 200,000 (guideline)

Can move budget between line items as long as total ≤ 2M
```

**Budget Revision Process:**

```markdown
**Budget Amendments:**

Original Budget (V1): Approved Dec 2024
├── Q1 Revision (V2): New product launch
├── Q2 Revision (V3): Staffing changes
└── Q3 Revision (V4): Market conditions

**Revision Workflow:**
1. Department submits revision request
2. Finance reviews and consolidates
3. Management approves changes
4. New budget version activated
5. Reports show V1 vs. V4 comparison
6. Variance analysis: original vs. current budget
```

### Cost Center Reporting

**Standard Cost Center Reports:**

**1. Cost Center P&L:**
```markdown
Sales Department - P&L
Month: January 2025

REVENUE                          Amount        % of Total
─────────────────────────────────────────────────────────
Product Sales                  4,500,000         85.7%
Service Revenue                  750,000         14.3%
─────────────────────────────────────────────────────────
Total Revenue                  5,250,000        100.0%

EXPENSES
Salaries & Benefits            1,050,000         20.0%
Commissions                      262,500          5.0%
Travel & Entertainment           175,000          3.3%
Marketing                        480,000          9.1%
Other Expenses                    52,000          1.0%
─────────────────────────────────────────────────────────
Total Expenses                 2,019,500         38.5%

NET INCOME                     3,230,500         61.5%
```

**2. Cost Center Comparison:**
```markdown
All Branches - Comparison
Month: January 2025

Branch      Revenue    Expenses  Net Income  Margin
────────────────────────────────────────────────────
Nairobi     8,500,000  4,250,000  4,250,000  50.0%
Mombasa     5,250,000  2,019,500  3,230,500  61.5% ✓
Kisumu      3,200,000  1,920,000  1,280,000  40.0%
Eldoret     1,500,000  1,200,000    300,000  20.0%
────────────────────────────────────────────────────
TOTAL      18,450,000  9,389,500  9,060,500  49.1%

Analysis: Mombasa highest margin; investigate Eldoret
```

**3. Trend Analysis:**
```markdown
Sales Department - 6-Month Trend

Month      Revenue     Expenses   Net Income  Margin
──────────────────────────────────────────────────────
Aug 2024   4,800,000  2,100,000  2,700,000   56.3%
Sep 2024   5,100,000  2,050,000  3,050,000   59.8%
Oct 2024   4,950,000  1,980,000  2,970,000   60.0%
Nov 2024   5,200,000  2,150,000  3,050,000   58.7%
Dec 2024   6,500,000  2,600,000  3,900,000   60.0%
Jan 2025   5,250,000  2,019,500  3,230,500   61.5% ✓

Trend: Revenue stable, margins improving
```

---

## Module Integration Points

### Sales Module Integration

**Order-to-Cash Cycle:**

```mermaid
graph LR
    A[Quotation] --> B[Sales Order]
    B --> C[Delivery Note]
    C --> D[Sales Invoice]
    D --> E[Payment Receipt]
    D -.-> F[General Ledger]
    E -.-> F
```

**Financial Impact at Each Stage:**

**1. Quotation Created:**
```markdown
**Financial Impact:** None
**Purpose:** Customer price proposal
**Status:** Not binding
```

**2. Sales Order Confirmed:**
```markdown
**Financial Impact:** None (or Optional commitment accounting)
**Purpose:** Customer commitment to purchase
**Inventory:** Reserved for customer
**Optional Entry (if commitment accounting):**
Dr. Unbilled Receivables
    Cr. Deferred Revenue
```

**3. Delivery Note (Goods Shipped):**
```markdown
**Financial Impact:** Inventory reduction

**Journal Entry:**
Dr. Cost of Goods Sold           300,000
    Cr. Inventory                        300,000

**Result:** 
- Inventory reduced
- COGS recognized
- No revenue yet (not invoiced)
```

**4. Sales Invoice Created:**
```markdown
**Financial Impact:** Revenue & Receivable recognition

**Journal Entry:**
Dr. Accounts Receivable          500,000
    Cr. Sales Revenue                    500,000

**If VAT applicable:**
Dr. Accounts Receivable          580,000
    Cr. Sales Revenue                    500,000
    Cr. VAT Payable (16%)                 80,000

**Result:**
- Revenue recognized
- Accounts receivable created
- Customer balance updated
- VAT liability recorded
```

**5. Payment Received:**
```markdown
**Financial Impact:** Cash collection

**Journal Entry:**
Dr. Cash/Bank Account            580,000
    Cr. Accounts Receivable              580,000

**Result:**
- Cash increased
- Receivable cleared
- Customer balance reduced
- Available credit restored
```

**Complete Transaction Example:**

```markdown
**Scenario: Sale of 100 units @ 5,000 each**

Step 1: Sales Order
- Quantity: 100 units
- Price: 5,000 KES/unit
- Total: 500,000 KES
- VAT (16%): 80,000 KES
- Invoice Total: 580,000 KES
GL Impact: None (reservation only)

Step 2: Delivery (Cost: 3,000 per unit)
Dr. Cost of Goods Sold           300,000
    Cr. Inventory - Product X            300,000

Step 3: Invoice
Dr. Accounts Receivable - ABC Co 580,000
    Cr. Sales Revenue - Product X        500,000
    Cr. VAT Payable                       80,000

Step 4: Payment (30 days later)
Dr. Bank Account - Main          580,000
    Cr. Accounts Receivable - ABC Co     580,000

**Net Impact on Financial Statements:**
Assets: +280,000 (Cash +580k, Inventory -300k)
Liabilities: +80,000 (VAT Payable)
Equity: +200,000 (Retained Earnings via Net Income)

Income Statement:
Revenue: +500,000
COGS: -300,000
Gross Profit: +200,000
```

### Purchasing Module Integration

**Procure-to-Pay Cycle:**

```mermaid
graph LR
    A[Purchase Requisition] --> B[Purchase Order]
    B --> C[Goods Receipt]
    C --> D[Purchase Invoice]
    D --> E[Payment]
    C -.-> F[General Ledger]
    D -.-> F
    E -.-> F
```

**Financial Impact at Each Stage:**

**1. Purchase Requisition:**
```markdown
**Financial Impact:** None (or optional budget reservation)
**Purpose:** Internal request for purchase
**Approval:** Department manager
```

**2. Purchase Order:**
```markdown
**Financial Impact:** None (or commitment if accrual accounting)
**Purpose:** Legally binding order to supplier
**Optional Entry:**
Dr. Inventory in Transit
    Cr. Purchase Order Commitment
```

**3. Goods Receipt:**
```markdown
**Financial Impact:** Inventory increase (if inventory item)

**For Inventory Items:**
Dr. Inventory - Raw Materials    200,000
    Cr. GR/IR Clearing Account*          200,000

* Goods Receipt/Invoice Receipt clearing

**For Non-Inventory (Direct Expense):**
Dr. Office Supplies Expense       15,000
    Cr. Accrued Expenses                  15,000
```

**4. Purchase Invoice Received:**
```markdown
**Financial Impact:** Payable recognition

**Journal Entry:**
Dr. GR/IR Clearing Account       200,000
Dr. VAT Receivable (16%)          32,000
    Cr. Accounts Payable - Supplier      232,000

**If Direct Expense:**
Dr. Office Supplies Expense       15,000
Dr. VAT Receivable                 2,400
    Cr. Accounts Payable                  17,400

**Result:**
- Payable to supplier created
- VAT recoverable recorded
- Supplier balance updated
- GR/IR account cleared
```

**5. Payment Made:**
```markdown
**Journal Entry:**
Dr. Accounts Payable - Supplier  232,000
    Cr. Bank Account - Main              232,000

**Result:**
- Cash decreased
- Payable cleared
- Supplier balance reduced
```

**Complete Transaction Example:**

```markdown
**Scenario: Purchase 500 units @ 400 KES each**

Step 1: Purchase Order
- Quantity: 500 units
- Cost: 400 KES/unit
- Subtotal: 200,000 KES
- VAT (16%): 32,000 KES
- Total: 232,000 KES
GL Impact: None (or commitment entry)

Step 2: Goods Receipt
Dr. Inventory - Raw Material A   200,000
    Cr. GR/IR Clearing                   200,000

Step 3: Invoice Received
Dr. GR/IR Clearing               200,000
Dr. VAT Receivable                32,000
    Cr. Accounts Payable - XYZ Ltd       232,000

Step 4: Payment (60 days later)
Dr. Accounts Payable - XYZ Ltd   232,000
    Cr. Bank Account - Main              232,000

**Net Impact on Financial Statements:**
Assets: -32,000 (Cash -232k, Inventory +200k, VAT +32k)
Liabilities: 0 (AP created then paid)
Equity: 0 (no P&L impact yet - inventory not sold)
```

### Inventory Module Integration

**Inventory Movement Impacts:**

**1. Purchase Receipt:**
```markdown
Dr. Inventory                    200,000
    Cr. GR/IR Clearing or AP            200,000
```

**2. Production Completion:**
```markdown
Dr. Inventory - Finished Goods   500,000
    Cr. Inventory - Raw Materials       200,000
    Cr. Inventory - WIP                  300,000
```

**3. Sales Delivery:**
```markdown
Dr. Cost of Goods Sold           300,000
    Cr. Inventory - Finished Goods      300,000
```

**4. Inventory Write-Off:**
```markdown
Dr. Inventory Loss Expense        25,000
    Cr. Inventory                         25,000
Description: Obsolete/damaged stock write-off
```

**5. Inventory Adjustment:**
```markdown
**Physical Count Reveals Shortage:**
Dr. Inventory Shrinkage Expense   10,000
    Cr. Inventory                         10,000
Description: Cycle count adjustment - shortage
```

**6. Inter-Location Transfer:**
```markdown
**No GL Impact - Just Location Change:**
Inventory - Warehouse A: -50,000
Inventory - Warehouse B: +50,000
(Same account, different locations)
```

### HR/Payroll Module Integration

**Payroll Processing:**

```mermaid
graph TD
    A[Timesheet Entry] --> B[Payroll Run]
    B --> C[Salary Calculation]
    C --> D[Deductions Calculated]
    D --> E[Payroll Journal Entry]
    E --> F[Payment Processing]
```

**Payroll Journal Entry Example:**

```markdown
**Monthly Payroll: January 2025**

GROSS SALARIES BY DEPARTMENT:
Dr. Salary Expense - Sales         800,000
Dr. Salary Expense - Operations    600,000
Dr. Salary Expense - Administration 400,000
Dr. Salary Expense - R&D           200,000
                                 ---------
Total Gross Salaries             2,000,000

EMPLOYER CONTRIBUTIONS:
Dr. NSSF Expense                    120,000
Dr. NHIF Expense                     80,000
Dr. Pension Expense                 100,000
                                   --------
Total Employer Costs                300,000

EMPLOYEE DEDUCTIONS:
    Cr. PAYE Payable                     400,000
    Cr. NSSF Payable                     120,000
    Cr. NHIF Payable                      80,000
    Cr. Pension Payable                  100,000
    Cr. Loan Deduction Payable            50,000
    Cr. Salary Advance Recovery           50,000
    Cr. Garnishments Payable              20,000
                                        --------
Total Deductions                         820,000

NET PAY:
    Cr. Salary Payable                 1,480,000

BALANCE CHECK:
Debits: 2,300,000 = Credits: 2,300,000 ✓

**Payment Processing:**
Dr. Salary Payable               1,480,000
    Cr. Bank Account - Payroll          1,480,000

Dr. PAYE Payable                   400,000
Dr. NSSF Payable                   240,000 (Employee + Employer)
Dr. NHIF Payable                   160,000 (Employee + Employer)
Dr. Pension Payable                200,000 (Employee + Employer)
Dr. Loan Deduction Payable          50,000
Dr. Garnishments Payable            20,000
    Cr. Bank Account - Tax Payments    1,070,000
```

**Employee Advances:**

```markdown
**When Advance Given:**
Dr. Employee Advances - [Name]     50,000
    Cr. Cash                              50,000

**Monthly Recovery (from payroll):**
Dr. Salary Payable                 10,000
    Cr. Employee Advances - [Name]        10,000
(Reduces net pay, clears advance over time)
```

**Employee Benefits:**

```markdown
**Medical Insurance (Employer Paid):**
Dr. Employee Benefits Expense      45,000
    Cr. Medical Insurance Payable         45,000

**When Premium Paid:**
Dr. Medical Insurance Payable      45,000
    Cr. Bank Account                      45,000
```

### Fixed Assets Module Integration

**Asset Acquisition:**

```markdown
**Purchase of Equipment:**
Dr. Equipment - Asset              850,000
Dr. VAT Receivable                 136,000
    Cr. Accounts Payable                 986,000

OR (if cash purchase):
    Cr. Cash                             986,000
```

**Depreciation:**

```markdown
**Monthly Depreciation Entry:**
Dr. Depreciation Expense            8,500
    Cr. Accumulated Depreciation - Equipment  8,500

**Allocation by Department (if asset used by specific dept):**
Dr. Depreciation - Sales            3,400
Dr. Depreciation - Operations       5,100
    Cr. Accumulated Depreciation - Equipment  8,500
```

**Asset Disposal:**

```markdown
**Scenario: Sell equipment**
Original Cost: 100,000
Accumulated Depreciation: 60,000
Book Value: 40,000
Sale Price: 35,000

**Journal Entry:**
Dr. Cash                            35,000
Dr. Accumulated Depreciation        60,000
Dr. Loss on Asset Disposal           5,000
    Cr. Equipment                        100,000

**If Sold for Profit (e.g., 45,000):**
Dr. Cash                            45,000
Dr. Accumulated Depreciation        60,000
    Cr. Equipment                        100,000
    Cr. Gain on Asset Disposal             5,000
```

---

## Financial Reporting Framework

### Report Categories

The Financial Module produces four categories of reports, each serving a distinct audience and purpose.

**1. Statutory Financial Statements**
Mandatory reports for regulatory compliance and external stakeholders:
- Balance Sheet (Statement of Financial Position)
- Income Statement (Profit & Loss)
- Cash Flow Statement
- Statement of Changes in Equity
- Notes to Financial Statements

**2. Management Reports**
Internal reports for decision-making:
- Budget vs. Actual (by cost center, period, account)
- Departmental P&L
- Rolling forecasts
- KPI dashboards

**3. Subsidiary Ledger Reports**
Detailed supporting reports:
- Accounts Receivable Aging
- Accounts Payable Aging
- General Ledger Detail
- Trial Balance

**4. Compliance Reports**
Tax and regulatory filings:
- VAT Return (input/output tax summary)
- Withholding Tax Register
- PAYE Summary
- Statutory deduction schedules

---

### Core Financial Statements

#### Balance Sheet

Snapshot of financial position at a specific date. Follows the fundamental equation:
**Assets = Liabilities + Equity**

```
BALANCE SHEET
As at 31 January 2025

ASSETS                              KES
─────────────────────────────────────────
Current Assets
  Cash & Cash Equivalents       1,250,000
  Accounts Receivable           3,450,000
  Inventory                     2,100,000
  Prepaid Expenses                180,000
  VAT Receivable                  320,000
─────────────────────────────────────────
Total Current Assets            7,300,000

Non-Current Assets
  Property, Plant & Equipment   8,500,000
  Less: Accumulated Depreciation(2,040,000)
  Intangible Assets               500,000
  Long-Term Investments           750,000
─────────────────────────────────────────
Total Non-Current Assets        7,710,000

TOTAL ASSETS                   15,010,000

LIABILITIES
─────────────────────────────────────────
Current Liabilities
  Accounts Payable              1,850,000
  VAT Payable                     480,000
  Accrued Expenses                220,000
  Short-Term Loan                 500,000
─────────────────────────────────────────
Total Current Liabilities       3,050,000

Non-Current Liabilities
  Long-Term Loan                3,000,000
  Deferred Tax Liability          150,000
─────────────────────────────────────────
Total Non-Current Liabilities   3,150,000

TOTAL LIABILITIES               6,200,000

EQUITY
─────────────────────────────────────────
  Share Capital                 5,000,000
  Retained Earnings             2,935,000
  Current Year Earnings           875,000
─────────────────────────────────────────
TOTAL EQUITY                    8,810,000

TOTAL LIABILITIES + EQUITY     15,010,000 ✓
```

**Balance Sheet Rules:**
- Assets must always equal Liabilities + Equity
- Comparative columns (prior year) required for statutory filings
- Current vs. non-current classification based on 12-month rule
- Contra-asset accounts (Accumulated Depreciation, Allowance for Doubtful Accounts) reduce gross values

---

#### Income Statement (Profit & Loss)

Performance over a period — revenue earned minus expenses incurred.

```
INCOME STATEMENT
For the Month of January 2025

                                   Jan 2025    YTD 2025
─────────────────────────────────────────────────────────
REVENUE
  Product Sales                  5,250,000    5,250,000
  Service Revenue                  750,000      750,000
  Other Income                      25,000       25,000
─────────────────────────────────────────────────────────
Total Revenue                    6,025,000    6,025,000

COST OF SALES
  Cost of Goods Sold             2,100,000    2,100,000
  Direct Labor                     400,000      400,000
  Manufacturing Overhead            200,000      200,000
─────────────────────────────────────────────────────────
Total Cost of Sales              2,700,000    2,700,000

GROSS PROFIT                     3,325,000    3,325,000
Gross Margin %                       55.2%        55.2%

OPERATING EXPENSES
  Salaries & Benefits            1,050,000    1,050,000
  Rent                              50,000       50,000
  Utilities                         52,000       52,000
  Depreciation                       8,500        8,500
  Marketing                        480,000      480,000
  Travel & Entertainment           175,000      175,000
  Other Operating                  150,000      150,000
─────────────────────────────────────────────────────────
Total Operating Expenses         1,965,500    1,965,500

OPERATING INCOME (EBIT)          1,359,500    1,359,500

  Interest Expense                 (75,000)     (75,000)
  Foreign Exchange Gain/(Loss)      30,000       30,000
─────────────────────────────────────────────────────────
INCOME BEFORE TAX                1,314,500    1,314,500

  Income Tax Expense (30%)        (394,350)    (394,350)
─────────────────────────────────────────────────────────
NET INCOME                         920,150      920,150
Net Margin %                         15.3%        15.3%
```

---

#### Cash Flow Statement

Cash movements classified by activity type. Critical for liquidity management.

```
CASH FLOW STATEMENT
For the Month of January 2025

OPERATING ACTIVITIES                         KES
─────────────────────────────────────────────────
Net Income                                920,150

Adjustments for non-cash items:
  Add: Depreciation                         8,500
  Add: Amortization                         5,000

Changes in Working Capital:
  (Increase)/Decrease in Receivables     (450,000)
  (Increase)/Decrease in Inventory        120,000
  (Increase)/Decrease in Prepaid           (30,000)
  Increase/(Decrease) in Payables         280,000
  Increase/(Decrease) in Accruals          45,000
─────────────────────────────────────────────────
Net Cash from Operating Activities        898,650

INVESTING ACTIVITIES
─────────────────────────────────────────────────
  Purchase of Equipment                  (250,000)
  Proceeds from Asset Disposal             35,000
  Purchase of Investments                (100,000)
─────────────────────────────────────────────────
Net Cash used in Investing Activities    (315,000)

FINANCING ACTIVITIES
─────────────────────────────────────────────────
  Loan Repayment                         (100,000)
  New Loan Drawdown                       500,000
  Owner Drawings                         (200,000)
─────────────────────────────────────────────────
Net Cash from Financing Activities        200,000

NET INCREASE IN CASH                      783,650
Opening Cash Balance                      466,350
─────────────────────────────────────────────────
CLOSING CASH BALANCE                    1,250,000 ✓
```

**Indirect Method (above) vs. Direct Method:**
- Indirect: Starts from net income, adjusts for non-cash items (most common)
- Direct: Lists actual cash receipts and payments (preferred by IFRS, harder to prepare)

---

#### Trial Balance

Internal listing of all account balances — primary tool for verifying double-entry integrity.

```
TRIAL BALANCE
As at 31 January 2025

Account                           Debit        Credit
──────────────────────────────────────────────────────
1120 Cash - Main Account       1,200,000
1121 Cash - Payroll               50,000
1210 Accounts Receivable       3,450,000
1310 Inventory                 2,100,000
1410 Prepaid Expenses            180,000
1430 VAT Receivable              320,000
1530 Equipment                 8,500,000
1535 Accum Depr - Equipment                  2,040,000
2110 Accounts Payable                        1,850,000
2310 VAT Payable                               480,000
2230 Accrued Expenses                          220,000
2410 Short-Term Loan                           500,000
2510 Long-Term Loan                          3,000,000
3100 Share Capital                           5,000,000
3300 Retained Earnings                       2,935,000
4110 Product Sales Revenue                   5,250,000
4310 Service Revenue                           750,000
5100 Cost of Goods Sold        2,100,000
7110 Salaries Expense          1,050,000
7310 Rent Expense                 50,000
...
──────────────────────────────────────────────────────
TOTALS                        22,025,000   22,025,000 ✓
```

---

### AR & AP Aging Reports

**Accounts Receivable Aging:**

```
AR AGING REPORT
As at 31 January 2025

Customer        Total     Current   31-60d    61-90d    >90d
──────────────────────────────────────────────────────────────
ABC Corp        450,000   450,000        -         -         -
XYZ Ltd         325,000   125,000   200,000        -         -
MNO Co          180,000        -    80,000    100,000        -
PQR Inc          95,000        -         -     45,000    50,000
──────────────────────────────────────────────────────────────
TOTAL         1,050,000   575,000   280,000   145,000    50,000
%                           54.8%     26.7%     13.8%     4.8%

⚠️ PQR Inc 50,000 over 90 days — consider provision
```

**Bad Debt Provision:**
```
Dr. Bad Debt Expense           10,000
    Cr. Allowance for Doubtful Accounts  10,000
Description: 20% provision on >90 day balance PQR Inc
```

---

### General Ledger Detail Report

Drill-down from any account balance to individual transactions:

```
GL DETAIL — Account: 1210 Accounts Receivable
Period: January 2025

Date        Ref         Description              Debit      Credit    Balance
────────────────────────────────────────────────────────────────────────────
01/01/25    OB-001      Opening Balance                             3,000,000
05/01/25    INV-0451    Sale - ABC Corp          580,000             3,580,000
12/01/25    REC-0122    Payment - XYZ Ltd                 325,000   3,255,000
18/01/25    INV-0472    Sale - MNO Co            195,000             3,450,000
────────────────────────────────────────────────────────────────────────────
CLOSING BALANCE                                                     3,450,000
```

---

### Report Configuration & Customization

**Report Parameters:**
- Date range (as-at date, period, custom)
- Comparative period (prior month, prior year, budget)
- Currency (base, foreign, both)
- Cost center filter or grouping
- Account group filter
- Detail level (summary vs. transaction-level)

**Output Formats:**
- On-screen (with drill-down)
- PDF (for distribution/archiving)
- Excel (for analysis)
- CSV (for import to other systems)

**Scheduling:**
- On-demand
- Scheduled delivery (daily, weekly, monthly)
- Triggered (on period close)
- Email distribution to stakeholders

---

## Approval Workflows

### Why Approval Workflows Matter

Without controlled approvals:
- ❌ Anyone can post any amount to any account
- ❌ No check on errors or fraud
- ❌ No audit trail of authorization
- ❌ Regulators and auditors find control gaps

With structured approvals:
- ✅ Segregation of duties enforced systemically
- ✅ Authority limits aligned to policy
- ✅ Complete approval audit trail
- ✅ Escalation when approver unavailable
- ✅ SOX and IFRS compliance supported

---

### Approval Rule Design

**Trigger Dimensions:**

Each approval rule is triggered by a combination of:

```markdown
**Rule Definition Fields:**
- Transaction Type: [Journal Entry, Payment, Invoice, etc.]
- Amount Threshold: [Min and Max range]
- Account Group: [e.g., "All Cash Accounts", "Equity Accounts"]
- Cost Center: [Specific department or "All"]
- Currency: [Specific or "All"]
- Approver Role: [Finance Manager, CFO, etc.]
- Approval Mode: [Sequential, Parallel, Any-one-of]
```

---

### Approval Hierarchy Examples

**Example 1: Journal Entry Approvals by Amount**

```markdown
Rule Set: Manual Journal Entries

Tier 1 — Up to 50,000 KES:
  Approver: Senior Accountant
  SLA: 4 hours

Tier 2 — 50,001 to 500,000 KES:
  Approver: Finance Manager
  SLA: 8 hours

Tier 3 — 500,001 to 2,000,000 KES:
  Approver: CFO
  SLA: 24 hours

Tier 4 — Above 2,000,000 KES:
  Sequential: Finance Manager → CFO → CEO
  SLA: 48 hours
```

**Example 2: Payment Approvals**

```markdown
Rule Set: Supplier Payments

< 25,000 KES:        AP Supervisor (single approval)
25,001–250,000 KES:  Finance Manager (single approval)
250,001–1,000,000:   Finance Manager + CFO (sequential)
> 1,000,000 KES:     Finance Manager + CFO + CEO (sequential)

Special Rule — Cash payments:
  Any amount: Requires CFO approval regardless of amount
```

**Example 3: Account-Based Rules**

```markdown
Sensitive Accounts:
- All accounts 1110–1140 (Cash): CFO approval for any debit
- Account 3500 (Owner Draws): CEO approval always
- Account 9900 (Tax Expense): CFO + Tax Advisor sign-off
- Retained Earnings adjustments: Board resolution required
```

---

### Approval States & Actions

```
DRAFT
  ↓ [Submit]
PENDING_APPROVAL
  ↓ [Approve]          ↓ [Reject]          ↓ [Request Info]
APPROVED              REJECTED            INFO_REQUESTED
  ↓ [Post]              ↓ [Edit]              ↓ [Respond]
POSTED                DRAFT               PENDING_APPROVAL
```

**Approver Actions:**

| Action | Effect | Notification |
|--------|--------|-------------|
| Approve | Moves to next tier or APPROVED | Submitter + next approver |
| Reject | Returns to REJECTED (editable) | Submitter with reason |
| Request Information | Pauses workflow | Submitter must respond |
| Delegate | Transfers to alternate approver | Delegate + submitter |
| Escalate | Moves to higher authority | Escalation target |

---

### Delegation & Escalation

**Delegation (Planned Absence):**
```markdown
Finance Manager sets delegation:
- Delegate To: Senior Accountant
- Valid From: 2025-02-10
- Valid To: 2025-02-17
- Scope: All approvals up to 200,000 KES
- Excludes: Cash account approvals

System automatically routes approvals to delegate during period.
```

**Escalation (SLA Breach):**
```markdown
SLA Configuration:
- Finance Manager approval SLA: 8 hours
- If not actioned in 8 hours:
  → Reminder notification sent to Finance Manager
- If not actioned in 12 hours:
  → Auto-escalate to CFO
  → Alert to Finance Director
  → Log SLA breach in audit trail
```

---

### Segregation of Duties Matrix

The system enforces these incompatible roles — no single user can hold conflicting permissions:

```markdown
INCOMPATIBLE COMBINATIONS (must be different users):

1. Create Transaction + Approve Transaction
   → Prevents self-approval of own entries

2. Create Supplier + Approve Payment to Supplier
   → Prevents fictitious vendor fraud

3. Manage Exchange Rates + Process FC Transactions
   → Prevents rate manipulation

4. Administer Users + Approve Transactions
   → Prevents creating ghost approvers

5. Process Payroll + Approve Payroll
   → Prevents unauthorized salary changes

6. Enter Opening Balances + Approve Opening Balances
   → Protects migration integrity
```

---

### Approval Audit Trail

Every approval action is permanently logged:

```markdown
Audit Entry:
  Transaction: JE-2025-0145
  Amount: 750,000 KES
  Action: APPROVED
  Approver: Jane Waweru (Finance Manager)
  Timestamp: 2025-01-15 14:32:07 EAT
  IP Address: 192.168.1.45
  Device: Desktop Chrome
  Comment: "Verified against PO #2345 and delivery receipt"
  Previous Approver: John Kamau (Senior Accountant) @ 11:15:22
```

This log is:
- Immutable (cannot be edited or deleted)
- Exportable for auditors
- Searchable by transaction, user, date, action

---

## Compliance & Regulatory Features

### GAAP & IFRS Compliance

The module is designed to support both GAAP (US) and IFRS (international) reporting standards. Key controls:

**Revenue Recognition (IFRS 15 / ASC 606):**
```markdown
Revenue is recognized when (or as) performance obligations are satisfied:

1. Contract identified with customer
2. Performance obligations identified
3. Transaction price determined
4. Price allocated to obligations
5. Revenue recognized when obligation satisfied

System Control:
- Revenue cannot be posted before delivery confirmation
- Deferred revenue account used for advance payments
- Automatic recognition schedule for subscriptions/contracts
```

**Lease Accounting (IFRS 16):**
```markdown
Operating leases now on balance sheet:

Right-of-Use Asset recognition:
Dr. Right-of-Use Asset           500,000
    Cr. Lease Liability                  500,000

Monthly Recognition:
Dr. Depreciation - ROU Asset      8,333
Dr. Interest Expense               2,500
    Cr. Lease Liability                    2,500
    Cr. Accumulated Depreciation          8,333

Dr. Lease Liability               10,000
    Cr. Cash (monthly payment)            10,000
```

**Inventory Costing Methods:**
```markdown
Supported methods (consistent application required):
- FIFO (First In, First Out)
- Weighted Average Cost
- Specific Identification

Note: LIFO not permitted under IFRS
```

---

### VAT / Tax Compliance

**VAT Calculation and Posting:**

```markdown
Standard Sales Transaction with VAT:

Gross Sale: 100,000 KES
VAT (16%):   16,000 KES
Total:       116,000 KES

Journal Entry:
Dr. Accounts Receivable          116,000
    Cr. Sales Revenue                    100,000
    Cr. VAT Payable (Output)              16,000
```

**Input VAT (Purchases):**

```markdown
Purchase with Claimable VAT:

Invoice: 58,000 KES (50,000 + 8,000 VAT)

Dr. Inventory / Expense           50,000
Dr. VAT Receivable (Input)         8,000
    Cr. Accounts Payable                  58,000
```

**VAT Return Calculation:**

```markdown
Monthly VAT Return:

Output VAT (charged to customers):        480,000
Input VAT (paid to suppliers):           (320,000)
─────────────────────────────────────────────────
NET VAT PAYABLE to KRA:                   160,000

Journal Entry on Filing:
Dr. VAT Payable                  480,000
    Cr. VAT Receivable                    320,000
    Cr. Cash / Bank                       160,000
```

**Withholding Tax:**

```markdown
When paying a consultant (5% WHT applies):

Invoice Amount:    100,000 KES
WHT (5%):           (5,000) KES
Net Payment:        95,000 KES

Dr. Consulting Expense           100,000
    Cr. Accounts Payable                 100,000

On Payment:
Dr. Accounts Payable             100,000
    Cr. Cash                              95,000
    Cr. WHT Payable                        5,000

WHT remitted to tax authority monthly:
Dr. WHT Payable                    5,000
    Cr. Cash                               5,000
```

---

### SOX Compliance Controls

For companies subject to Sarbanes-Oxley (or equivalent corporate governance):

**Control 1: Access Controls**
```markdown
- Role-based permissions with least-privilege principle
- No shared user accounts
- Password policy enforced at system level
- Session timeout after inactivity
- Multi-factor authentication for sensitive operations
```

**Control 2: Change Management**
```markdown
- All configuration changes logged with before/after values
- COA changes require Finance Manager approval
- Period open/close logged with authorizer identity
- User role changes require IT + Finance co-approval
```

**Control 3: Audit Trail Completeness**
```markdown
Every record maintains:
- Created by (user ID + timestamp)
- Last modified by (user ID + timestamp)
- Approved by (user ID + timestamp)
- Posted by (user ID + timestamp)
- Reversed by (user ID + timestamp, if applicable)

Cannot be overridden or deleted.
```

**Control 4: Reconciliation Controls**
```markdown
Automated reconciliation checks run at period close:
- GL control account = subsidiary ledger total (AR, AP, Inventory)
- Bank balance per GL = bank statement balance
- Intercompany accounts net to zero
- Budget-to-actual variance flagged above threshold
```

---

### Multi-Entity & Intercompany

**Intercompany Transaction Example:**

```markdown
Scenario: Parent loans 1,000,000 KES to Subsidiary

PARENT BOOKS:
Dr. Intercompany Receivable - Sub    1,000,000
    Cr. Cash                                 1,000,000

SUBSIDIARY BOOKS:
Dr. Cash                             1,000,000
    Cr. Intercompany Payable - Parent        1,000,000

CONSOLIDATION ELIMINATION:
Dr. Intercompany Payable             1,000,000
    Cr. Intercompany Receivable              1,000,000
(Eliminates intra-group balances on consolidated statements)
```

**Intercompany Rules:**
1. Corresponding entries must be posted in both entities
2. System validates intercompany accounts net to zero at consolidation
3. Intercompany profit on transferred inventory eliminated on consolidation
4. Currency differences on intercompany balances posted to translation reserve

---

### Data Retention & Archiving

```markdown
Retention Policy:
- Transaction records: 7 years minimum (statutory requirement)
- Audit logs: 7 years minimum
- Tax-related records: 10 years
- Annual financial statements: Permanent

Archiving Process:
- Records older than 5 years moved to archive storage
- Archived records remain queryable (read-only)
- Archive export available for auditor access
- Deletion requires CFO + Legal approval and is logged
```

---

## Common Business Scenarios

### Scenario 1: Month-End Accruals

**Problem:** Received electricity service in January but bill arrives February 5.

```markdown
Step 1 — January 31 (Accrual):
Dr. Utilities Expense             22,000
    Cr. Accrued Utilities                 22,000
Description: Jan electricity estimate (meter reading)
Auto-Reversal Date: 2025-02-01

Step 2 — February 1 (Automatic Reversal):
Dr. Accrued Utilities             22,000
    Cr. Utilities Expense                 22,000
Description: Auto-reversal of Jan accrual

Step 3 — February 5 (Actual Bill):
Dr. Utilities Expense             21,450
Dr. VAT Receivable                 3,432
    Cr. Accounts Payable                  24,882
Description: Jan electricity bill KEN-2025-0145

Result:
- January P&L includes 22,000 estimate ✓
- February P&L shows actual 21,450 ✓
- Accrual cleared, no double-count ✓
```

---

### Scenario 2: Customer Advance Payment

**Problem:** Customer pays 500,000 KES upfront before delivery.

```markdown
Step 1 — Receipt of Advance:
Dr. Bank Account                 500,000
    Cr. Customer Deposits (Liability)    500,000
Description: Advance from ABC Corp - Order #5001

Step 2 — Goods Delivered (Invoice):
Dr. Accounts Receivable          580,000
    Cr. Sales Revenue                    500,000
    Cr. VAT Payable                       80,000
Description: Invoice INV-2025-0301 for Order #5001

Step 3 — Apply Advance to Invoice:
Dr. Customer Deposits            500,000
    Cr. Accounts Receivable              500,000
Description: Apply advance to INV-2025-0301

Step 4 — Collect Balance:
Dr. Bank Account                  80,000
    Cr. Accounts Receivable               80,000
Description: Balance payment - VAT only

Net Result:
- Revenue recognized on delivery ✓
- No premature P&L recognition ✓
- Liability cleared ✓
```

---

### Scenario 3: Loan Drawdown and Repayment

**Problem:** Company draws a new 5,000,000 KES term loan.

```markdown
Loan Drawdown:
Dr. Bank Account               5,000,000
    Cr. Long-Term Loan                 5,000,000

Monthly Repayment (Capital + Interest):
Loan: 5,000,000 | Rate: 12% p.a. | Term: 60 months
Monthly Payment: ~111,222 KES

Month 1 Interest: 5,000,000 × 1% = 50,000
Month 1 Capital:  111,222 - 50,000 = 61,222

Dr. Long-Term Loan (Capital)      61,222
Dr. Interest Expense              50,000
    Cr. Bank Account                     111,222

Reclassify current portion at year-end:
Dr. Long-Term Loan               734,664  (12 months capital)
    Cr. Current Portion of LT Loan       734,664
```

---

### Scenario 4: Bad Debt Write-Off

**Problem:** Customer PQR Inc has gone insolvent. Balance: 50,000 KES.

```markdown
Step 1 — Provision (when doubt arises):
Dr. Bad Debt Expense              50,000
    Cr. Allowance for Doubtful Accounts  50,000
Description: Provision — PQR Inc insolvency proceedings

Step 2 — Write-Off (when confirmed uncollectable):
Dr. Allowance for Doubtful Accounts 50,000
    Cr. Accounts Receivable - PQR Inc    50,000
Description: Write-off confirmed — liquidator notice ref LIQ-456

Step 3 — Recovery (if later recovered):
Dr. Accounts Receivable - PQR Inc 50,000
    Cr. Bad Debt Recovery Income         50,000

Dr. Bank Account                  50,000
    Cr. Accounts Receivable - PQR Inc    50,000

Note: P&L hit is in Step 1 (provision), not Step 2.
Write-off is balance sheet only when provision already exists.
```

---

### Scenario 5: Foreign Currency Supplier Payment

**Problem:** Pay US supplier $20,000. Invoice rate was 128.00 KES/USD; payment rate is 132.50 KES/USD.

```markdown
Invoice Date (Rate 128.00):
Dr. Inventory                  2,560,000  (20,000 × 128)
    Cr. AP - USD Supplier              2,560,000

Payment Date (Rate 132.50):
Cash needed: 20,000 × 132.50 = 2,650,000 KES
FX Loss:  2,650,000 - 2,560,000 = 90,000 KES

Dr. AP - USD Supplier          2,560,000
Dr. Foreign Exchange Loss         90,000
    Cr. Bank Account - KES         2,650,000

P&L Impact:
- Inventory cost remains at invoice rate (2,560,000) ✓
- FX loss (90,000) separately recognized in Other Expenses ✓
```

---

### Scenario 6: Prepaid Insurance Schedule

**Problem:** Pay annual insurance premium 480,000 KES on January 1.

```markdown
January 1 — Full Payment:
Dr. Prepaid Insurance            480,000
    Cr. Bank Account                     480,000

Each Month (40,000 KES/month recognition):
Dr. Insurance Expense             40,000
    Cr. Prepaid Insurance                 40,000

Recurring entry set up for 12 months automatically.

Balance Sheet Impact:
- Jan 1:  Prepaid = 480,000
- Jan 31: Prepaid = 440,000 (11 months remaining)
- Jun 30: Prepaid = 240,000 (6 months remaining)
- Dec 31: Prepaid = 0 (fully expensed) ✓
```

---

### Scenario 7: Intercompany Recharge

**Problem:** Head office provides shared services to branch. Recharge 150,000 KES.

```markdown
HEAD OFFICE BOOKS:
Dr. Intercompany Receivable - Branch  150,000
    Cr. Shared Services Revenue               150,000

BRANCH BOOKS:
Dr. Shared Services Expense      150,000
    Cr. Intercompany Payable - HO            150,000

Settlement (monthly):
HEAD OFFICE:
Dr. Bank Account                 150,000
    Cr. Intercompany Receivable          150,000

BRANCH:
Dr. Intercompany Payable         150,000
    Cr. Bank Account                     150,000

Consolidation: Intercompany revenue/expense eliminated.
Each entity's standalone P&L shows full cost/income.
```

---

## Troubleshooting Guide

### Problem 1: Trial Balance Does Not Balance

**Symptom:** Total debits ≠ total credits on trial balance.

**Causes & Resolutions:**

```markdown
Check 1 — Incomplete Transactions
- Run "Incomplete Transactions" report
- Look for entries with missing lines
- Resolution: Complete or cancel the entry

Check 2 — Opening Balance Error
- Run trial balance as at go-live date only
- Compare to source system trial balance
- Resolution: Post correcting opening entry

Check 3 — Currency Conversion Rounding
- Check transactions with very small amounts
- Multi-currency rounding can create 1-cent differences
- Resolution: Post rounding adjustment to Suspense account

Check 4 — System/Integration Error
- Check system error logs for failed postings
- Look for integration transactions stuck in PENDING
- Resolution: Repost failed entries; escalate to IT if persistent
```

---

### Problem 2: Bank Reconciliation Difference

**Symptom:** GL bank balance differs from bank statement.

**Standard Investigation Checklist:**

```markdown
Step 1 — List Outstanding Items
  Items in GL but not bank statement:
  - Checks issued but not yet presented
  - Deposits in transit

  Items on bank statement but not GL:
  - Bank charges not yet recorded
  - Interest income not yet recorded
  - Direct debits not yet entered

Step 2 — Prepare Reconciliation
  GL Balance per Books:         1,250,000
  Add: Deposits in Transit         80,000
  Less: Outstanding Checks       (120,000)
  ─────────────────────────────────────────
  Adjusted GL Balance:          1,210,000

  Bank Statement Balance:       1,225,000
  Less: Bank Charges not in GL    (15,000)
  ─────────────────────────────────────────
  Adjusted Bank Balance:        1,210,000 ✓

Step 3 — Post Adjustments
  Dr. Bank Charges                 15,000
      Cr. Bank Account                     15,000
```

---

### Problem 3: Accounts Receivable Control Account Mismatch

**Symptom:** AR control account balance ≠ sum of customer balances.

```markdown
Cause 1 — Direct GL Posting to AR Account
- Someone posted directly to AR control instead of via invoice
- Detection: AR detail report vs GL detail report side by side
- Resolution: Reverse direct entry; re-enter via proper invoice

Cause 2 — Multi-Currency Revaluation Not Applied
- FC invoices not revalued at period end
- Resolution: Run period-end revaluation for all FC accounts

Cause 3 — Deleted Customer Record with Open Balance
- Customer deleted while balance remained
- Resolution: Restore customer or post write-off

Cause 4 — Failed Integration Transaction
- Sales module posting partially failed
- Detection: Check integration error log
- Resolution: Retry or manually enter missing entry
```

---

### Problem 4: Posted Transaction with Wrong Account

**Symptom:** Expense posted to wrong account (already posted, cannot edit).

```markdown
Option 1 — Reversal + Re-entry (Preferred)

Step 1: Reverse original entry:
  Dr. Correct Account (Advertising)     5,000
      Cr. Wrong Account (Office Supplies)     5,000
  Reference: "Reversal of JE-2025-0089 - wrong account"

Step 2: Post correct entry:
  Dr. Advertising Expense             5,000
      Cr. Cash                              5,000
  Reference: "Repost of JE-2025-0089 corrected account"

Option 2 — Reclassification Entry (simpler, acceptable for small amounts)
  Dr. Advertising Expense             5,000
      Cr. Office Supplies Expense          5,000
  Description: "Reclassify marketing materials to correct account"

Both options are auditable. Use Option 1 for material amounts.
```

---

### Problem 5: Period Accidentally Closed Too Early

**Symptom:** Transactions need to be posted in a period that is now closed.

```markdown
Procedure:
1. Assess materiality — is this worth reopening?
   - Immaterial: Post in current open period with prior-period flag
   - Material: Reopen the period

2. To reopen (requires CFO permission):
   - Navigate to Period Management
   - Select closed period
   - Click "Reopen Period" — enter reason
   - System logs: who, when, why

3. Post required transactions

4. Re-close the period immediately

5. Re-run and distribute affected reports

6. Notify stakeholders of restatement if reports already distributed

7. Document the exception in the monthly close checklist
```

---

### Problem 6: Duplicate Invoice Posted

**Symptom:** Same supplier invoice posted twice.

```markdown
Detection:
- AP aging shows two entries for same invoice number
- Supplier statement reconciliation reveals duplicate
- Duplicate invoice report (filter by supplier + amount + date range)

Resolution:
Step 1: Identify the duplicate entry number

Step 2: If payment not yet made:
  - Reverse the duplicate invoice entry
  - Dr. AP - Supplier        [Amount]
      Cr. Expense Account          [Amount]

Step 3: If payment already made:
  - Contact supplier for credit note or refund
  - On receipt of credit note:
    Dr. AP - Supplier        [Amount]
        Cr. Expense Account        [Amount]
  - Apply credit to next invoice or request refund

Prevention:
- Enable "Duplicate Invoice Check" system setting
- System warns/blocks if same supplier + invoice number + amount exists
```

---

### Problem 7: Foreign Exchange Rate Missing

**Symptom:** Cannot post foreign currency transaction — no rate found for date.

```markdown
Resolution:
1. Check exchange rate master for the transaction date
   - If missing: Enter rate manually (source from Central Bank)
   - If rate entry blocked: Contact Finance Manager to authorize

2. If transaction is backdated and rate no longer available:
   - Use the nearest available date rate
   - Document the substitution and reason
   - Add note to transaction description

3. Prevent recurrence:
   - Set up daily rate import automation
   - Configure alert: "No rate loaded by 9 AM" → notify Finance team
   - Keep at least 30-day historical rates always loaded
```

---

## Business Rules & Validation

### Core Financial Rules

These rules are enforced by the system and cannot be bypassed:

```markdown
RULE F-01: Double-Entry Integrity
  Condition: Total debits MUST equal total credits
  Enforcement: Hard block — transaction cannot be saved
  Error: "Transaction is not balanced. Debit total: X, Credit total: Y"

RULE F-02: Minimum Lines
  Condition: Every transaction requires minimum 2 lines
  Enforcement: Hard block
  Error: "Transaction must have at least one debit and one credit line"

RULE F-03: Positive Amounts Only
  Condition: Line amounts must be > 0
  Enforcement: Hard block
  Error: "Amount must be greater than zero"

RULE F-04: No Dual-Sided Lines
  Condition: Each line must have debit OR credit, not both
  Enforcement: Hard block
  Error: "A transaction line cannot have both debit and credit amounts"

RULE F-05: Leaf Account Only
  Condition: Transactions can only post to leaf (non-group) accounts
  Enforcement: Hard block
  Error: "Account [X] is a group account and cannot accept transactions"

RULE F-06: Active Account Only
  Condition: Transactions can only post to active accounts
  Enforcement: Hard block
  Error: "Account [X] is inactive"
```

---

### Date & Period Rules

```markdown
RULE P-01: Open Period Required
  Condition: Transaction date must fall in an open accounting period
  Enforcement: Hard block
  Error: "Period [Month Year] is closed. Contact Finance to reopen."

RULE P-02: No Future Dates
  Condition: Transaction date cannot be in the future
  Enforcement: Hard block (configurable to warn for integration transactions)
  Error: "Transaction date cannot be in the future"

RULE P-03: Fiscal Year Boundary
  Condition: Transactions cannot span two fiscal years
  Enforcement: Hard block
  Error: "Transaction date is outside the current fiscal year"

RULE P-04: Posting Date Alignment
  Condition: Posting date must be >= transaction date
  Enforcement: Hard block
  Error: "Posting date cannot be before transaction date"
```

---

### Account-Specific Rules

```markdown
RULE A-01: Manual Entry Permission
  Condition: Account must have "Allow Manual Entries" enabled
  Applies to: MANUAL transaction type only
  Enforcement: Hard block
  Error: "Account [X] does not allow manual journal entries"
  Note: System/Integration entries bypass this rule

RULE A-02: Reference Required
  Condition: If account has "Require Reference" flag, line must have reference
  Enforcement: Hard block
  Error: "Reference number is required for account [X]"

RULE A-03: Cost Center Required
  Condition: If account type = Expense and cost center tracking is enabled
  Enforcement: Configurable (warn or block)
  Error: "Cost center is required for expense account [X]"

RULE A-04: Currency Consistency
  Condition: If account has a locked currency (e.g., USD bank account),
             all transactions must use that currency
  Enforcement: Hard block
  Error: "Account [X] is locked to currency USD"
```

---

### Approval & Authorization Rules

```markdown
RULE W-01: Self-Approval Prohibited
  Condition: Approver cannot approve their own submissions
  Enforcement: Hard block
  Error: "You cannot approve your own transaction"

RULE W-02: Authority Limit
  Condition: Approver cannot approve above their configured limit
  Enforcement: Hard block — auto-escalated to next tier
  Error: "Amount exceeds your approval authority. Escalated to [Role]."

RULE W-03: Sequential Order
  Condition: For sequential approvals, tier N+1 cannot act before tier N
  Enforcement: Hard block
  Error: "Awaiting approval from [User] before this step can proceed"

RULE W-04: Post Requires Approval
  Condition: If approval workflow is active, APPROVED status required before POST
  Enforcement: Hard block
  Error: "Transaction must be approved before posting"
```

---

### Budget Rules

```markdown
RULE B-01: Budget Availability Check (Soft Control)
  Condition: Posting expense > remaining budget triggers warning
  Enforcement: Warning with override option
  Warning: "This entry will exceed budget by [Amount]. Reason required to proceed."

RULE B-02: Hard Budget Stop
  Condition: When hard budget control enabled, posting above budget blocked
  Enforcement: Hard block
  Error: "Insufficient budget. Available: [X]. Required: [Y].
          Request a budget amendment to proceed."

RULE B-03: Budget Period Alignment
  Condition: Budget check uses transaction date to find correct budget period
  Enforcement: Automatic — no user action needed

RULE B-04: Budget Version
  Condition: System uses the ACTIVE budget version for checks
  Rule: Only one budget version can be ACTIVE at a time
  Enforcement: System prevents activating two versions simultaneously
```

---

### Currency Rules

```markdown
RULE C-01: Exchange Rate Required
  Condition: Foreign currency transaction requires rate for transaction date
  Enforcement: Hard block
  Error: "No exchange rate found for [Currency] on [Date]. Load rate and retry."

RULE C-02: Rate Reasonableness Check
  Condition: If new rate deviates > configured threshold from previous rate,
             manual confirmation required
  Default Threshold: 5% deviation triggers alert
  Enforcement: Warning — authorized user must confirm

RULE C-03: Functional Currency Calculation
  Condition: System calculates functional currency = FC Amount × Exchange Rate
  Rule: User cannot manually override the functional currency amount
  Enforcement: Calculated field (read-only)

RULE C-04: Revaluation at Period End
  Condition: Open FC balances must be revalued at period-end rate
  Enforcement: Checklist reminder in period-close workflow
  Blocking: Period cannot hard-close without confirming revaluation run
```

---

### Data Integrity Rules

```markdown
RULE D-01: Posted Entry Immutability
  Condition: Once POSTED, a transaction cannot be edited
  Enforcement: All fields locked; Edit button hidden
  Correction path: Create a reversal entry

RULE D-02: Account Code Immutability
  Condition: Once a transaction exists against an account code, the code
             cannot be changed
  Enforcement: Hard block on COA edit screen
  Error: "Account code cannot be changed after transactions are posted"

RULE D-03: Fiscal Year Lock
  Condition: Once a fiscal year is LOCKED, no transactions permitted
  Enforcement: Hard block across all entry points
  Unlock: Requires CFO + CEO dual authorization; logs to audit trail

RULE D-04: Audit Trail Preservation
  Condition: Audit trail records cannot be deleted, edited, or suppressed
  Enforcement: Database-level (not user-configurable)
  Applies to: All transactions, approvals, configuration changes, logins
```

---

### Validation Execution Order

When a transaction is submitted or posted, validations execute in this order:

```
1. Structural Validation
   - Line count (≥ 2)
   - Balance (debits = credits)
   - Amount signs (positive only)

2. Date Validation
   - Not future
   - Period open
   - Fiscal year boundary

3. Account Validation
   - Exists and active
   - Is leaf account
   - Allows manual entries (if applicable)
   - Currency locked (if applicable)
   - Reference required (if applicable)
   - Cost center required (if applicable)

4. Authorization Validation
   - User has permission for accounts
   - Approval obtained (if required)
   - Authority limit not exceeded

5. Business Rule Validation
   - Budget availability (if enabled)
   - Exchange rate available (if FC)
   - Duplicate check (if enabled)

6. Integration Validation
   - Source document exists (for integration transactions)
   - Source document not already posted to GL

7. Post to General Ledger
   - Only if all above pass
```

All validation failures return specific, actionable error messages — no generic "An error occurred."

---

## Bank Reconciliation

### Overview

Bank reconciliation is the process of matching the company's General Ledger cash account balances against the bank's official statement. It ensures every transaction is accounted for and identifies discrepancies (bank errors, unrecorded charges, timing differences) before the period is closed.

The Financial Module provides a structured reconciliation workspace — not just a report — where finance staff can import statements, match entries, investigate differences, and lock the reconciliation for the period.

---

### Reconciliation Workspace

Each bank account has its own reconciliation workspace. One reconciliation is created per statement period (usually monthly).

**Reconciliation Header:**
```markdown
Bank Account:       1120 - Checking Account - Main (KES)
Bank:               Equity Bank Kenya
Statement Period:   1 January 2025 – 31 January 2025
Statement Balance:  1,225,000 KES
GL Balance (1120):  1,210,000 KES
Status:             IN PROGRESS

Opened by:          Jane Waweru (AP Clerk)
Opened at:          2025-02-03 09:14
```

---

### Bank Statement Import

**Supported Import Formats:**

| Format | Description | Common Source |
|--------|-------------|---------------|
| OFX/QFX | Open Financial Exchange | Most Kenyan banks |
| MT940 | SWIFT bank statement | International/correspondent banks |
| CSV | Comma-separated (configurable mapping) | Any bank with online banking export |
| Excel | .xlsx with column mapping | Manual extraction |
| PDF | Parsed via OCR (last resort) | Legacy statements |

**CSV Field Mapping Configuration:**
```markdown
Column Map (configured once per bank, reused):
  Date:           Column A (format: DD/MM/YYYY)
  Description:    Column B
  Debit Amount:   Column C
  Credit Amount:  Column D
  Reference:      Column E
  Running Balance: Column F (optional, for validation)

Validation on Import:
  - Opening balance matches previous period closing ✓
  - Running balance column reconciles to closing balance ✓
  - No duplicate statement lines (same date + amount + ref) ✓
```

**Import Result Summary:**
```markdown
Bank Statement Import — January 2025
  Total lines imported:     87
  Duplicate lines rejected:  0
  Opening balance:    441,350 KES ✓ (matches prior period close)
  Closing balance:  1,225,000 KES
  Total debits:     2,340,000 KES
  Total credits:    3,123,650 KES
```

---

### Auto-Matching Engine

After import, the system attempts to automatically match statement lines to GL entries.

**Matching Rules (applied in priority order):**

```markdown
RULE M-01: Exact Match (Highest confidence)
  Criteria: Amount + Date (±0 days) + Reference exact match
  Auto-match: Yes — requires no user action
  Example: Bank shows "CHQ 001234 50,000" ↔ GL entry ref "CHQ-001234" 50,000

RULE M-02: Amount + Date Match
  Criteria: Amount exact + Date within ±2 business days
  Auto-match: Suggested (user confirms)
  Example: Bank credit 580,000 on Jan 5 ↔ GL receipt 580,000 on Jan 5

RULE M-03: Amount + Partial Reference Match
  Criteria: Amount exact + Reference contains substring
  Auto-match: Suggested (user confirms)
  Example: Bank "MPESA REF ABC123 25,000" ↔ GL ref "ABC123" 25,000

RULE M-04: Batch Match (One-to-Many)
  Criteria: Bank total = sum of multiple GL entries
  Auto-match: Suggested (user confirms group)
  Example: Bank deposit 150,000 ↔ GL receipts 80,000 + 70,000

RULE M-05: No Match
  Outcome: Listed as unmatched — requires manual investigation
```

**Auto-Match Summary:**
```markdown
Matching Results:
  Total statement lines:        87
  Auto-matched (exact):         61   (70.1%)
  Suggested (needs confirm):    12   (13.8%)
  Unmatched — bank only:         9   (10.3%)
  Unmatched — GL only:           5    (5.7%)
```

---

### Manual Matching

For unmatched items, the finance clerk works through each one:

**Case 1: Bank charge not in GL**
```markdown
Bank Statement Line:
  Date: 31 Jan 2025 | Description: Monthly Service Fee | Amount: -500 KES

Action: Create GL entry
Dr. 7740 - Bank Charges          500
    Cr. 1120 - Checking Account       500
Description: Jan 2025 bank service fee per statement

→ Line now matched ✓
```

**Case 2: Deposit in transit (GL has it, bank doesn't yet)**
```markdown
GL Entry:
  Date: 31 Jan 2025 | Ref: REC-0299 | Customer payment | +180,000 KES

Bank Statement:
  Not on January statement (received Jan 31, bank credited Feb 1)

Action: Mark as "Deposit in Transit"
  → Does NOT create a GL entry
  → Carries forward to February reconciliation automatically
  → Listed in reconciliation notes as timing difference
```

**Case 3: Outstanding cheque**
```markdown
GL Entry:
  Date: 15 Jan 2025 | Ref: CHQ-001290 | Supplier payment | -85,000 KES

Bank Statement:
  Not presented in January

Action: Mark as "Outstanding Cheque"
  → Carries forward to February
  → If outstanding > 6 months → flag for stale cheque investigation
```

**Case 4: Bank error**
```markdown
Bank Statement:
  Date: 20 Jan 2025 | Description: Transfer | Amount: -20,000 KES
  (This transaction belongs to another account — bank error)

Action: Mark as "Bank Error"
  → Add note: "Reported to bank ref: ERR-2025-0120"
  → Does not affect GL
  → Resolves when bank reverses
```

---

### Reconciliation Statement

The reconciliation produces a formal statement:

```
BANK RECONCILIATION STATEMENT
Account: 1120 - Checking Account - Main
Period:  January 2025
Prepared by: Jane Waweru | Reviewed by: John Kamau (Finance Manager)

BALANCE PER BANK STATEMENT:                          1,225,000
─────────────────────────────────────────────────────────────
ADD: Deposits in Transit
  31 Jan — REC-0299 Customer payment                   180,000
─────────────────────────────────────────────────────────────
LESS: Outstanding Cheques
  15 Jan — CHQ-001290 Supplier payment                 (85,000)
  22 Jan — CHQ-001295 Supplier payment                 (60,000)
─────────────────────────────────────────────────────────────
LESS: Bank Errors (reported)
  20 Jan — Erroneous debit (ref ERR-2025-0120)         (20,000)
─────────────────────────────────────────────────────────────
ADJUSTED BANK BALANCE:                                1,240,000

BALANCE PER GL (Account 1120):                        1,225,000
─────────────────────────────────────────────────────────────
ADD: GL adjustments posted during reconciliation
  31 Jan — Bank service fee                               500
  28 Jan — Interest income                             (14,500)
─────────────────────────────────────────────────────────────
ADJUSTED GL BALANCE:                                  1,240,000

DIFFERENCE:                                                   0 ✓

Status: BALANCED — ready for approval
```

---

### Reconciliation Lifecycle

```
OPEN → IN PROGRESS → BALANCED → UNDER REVIEW → APPROVED → LOCKED
         ↑                           ↓
    (reopen if                  (send back if
     needed)                     issues found)
```

| State | Who Can Act | Actions Available |
|-------|------------|------------------|
| OPEN | AP Clerk | Import statement, begin matching |
| IN PROGRESS | AP Clerk | Match, create GL adjustments, mark timing differences |
| BALANCED | AP Clerk | Submit for review |
| UNDER REVIEW | Finance Manager | Approve or return with comments |
| APPROVED | Finance Manager | Lock reconciliation |
| LOCKED | Nobody | Read-only; archived |

**Approval required before period can hard-close.**

---

### Reconciliation Controls

```markdown
CONTROL R-01: Mandatory before period close
  No accounting period can transition to HARD CLOSE
  until all active bank accounts have an APPROVED reconciliation.

CONTROL R-02: GL adjustment audit trail
  Every GL entry created during reconciliation is tagged
  with source: "BANK_REC" and reconciliation ID.

CONTROL R-03: Statement import immutability
  Once a statement is imported, lines cannot be deleted.
  Corrections require an explanatory note.

CONTROL R-04: Outstanding item aging
  Items outstanding > 60 days auto-flag for investigation.
  Items outstanding > 180 days escalate to Finance Manager.

CONTROL R-05: Segregation of duties
  Reconciliation preparer ≠ reconciliation approver.
  System enforces this — the submitter cannot approve their own rec.
```

---

## Cash Management & Forecasting

### Petty Cash (Imprest System)

The imprest system maintains a fixed petty cash float. Disbursements reduce the float; replenishment restores it to the fixed amount.

**Setup:**
```markdown
Petty Cash Fund: 50,000 KES
Account: 1111 - Petty Cash
Custodian: Office Administrator
Replenishment Threshold: When balance < 10,000 KES
```

**Petty Cash Disbursement:**
```markdown
Petty Cash Voucher #PC-2025-0045
  Date: 15 Jan 2025
  Payee: Courier Service
  Amount: 1,500 KES
  Expense Type: Postage & Shipping
  Authorized by: Office Manager
  Receipt attached: Yes

No GL entry at disbursement — petty cash is tracked in custodian's log.
GL entry only on replenishment.
```

**Petty Cash Replenishment:**
```markdown
At replenishment, the full expense detail is recorded:

Replenishment Cheque: CHQ-001310 for 38,500 KES
(Restores 50,000 fund — 38,500 was spent since last top-up)

Dr. 7520 - Postage & Shipping       8,500
Dr. 7510 - Office Supplies          6,000
Dr. 7320 - Utilities                4,000
Dr. 6400 - Travel Expense          15,000
Dr. 7740 - Miscellaneous            5,000
    Cr. 1120 - Checking Account            38,500

Physical cash count after: 50,000 KES ✓
```

**Petty Cash Shortage/Overage:**
```markdown
If count shows 49,200 (short by 800):

Dr. 7510 - Petty Cash Shortage         800
    Cr. 1111 - Petty Cash                    800

Investigate: missing receipts, math errors, pilferage
```

---

### Payment Run (Bulk Supplier Payments)

Instead of paying each invoice individually, the system batches due payments into a single payment run.

**Payment Run Process:**

```markdown
Step 1 — Generate Due Payments Report
  Filter: AP invoices where due_date <= 2025-02-07
  Include: Invoices not yet on a payment run

  Result:
  Supplier          Invoice      Due Date    Amount
  ────────────────────────────────────────────────
  XYZ Ltd           INV-2024-440  03 Feb    232,000
  ABC Supplies      INV-2024-451  05 Feb     58,000
  Office Pro        INV-2025-001  07 Feb     17,400
  ────────────────────────────────────────────────
  Total:                                    307,400

Step 2 — Review & Approve Payment List
  Finance Manager reviews; may exclude or add items
  CFO approves total run (> 250,000 KES threshold)

Step 3 — Generate Payment File
  Output: Bank-compatible payment file
  Format: SWIFT MT101 | Kenya EFT | Custom CSV for Equity Bank

  File includes: Beneficiary name, account, bank code, amount, reference

Step 4 — Upload to Bank
  Finance Manager uploads file to internet banking
  Bank processes batch (same-day or next-day value)

Step 5 — Confirm and Post
  On bank confirmation receipt:

  Dr. Accounts Payable - XYZ Ltd     232,000
  Dr. Accounts Payable - ABC Supplies  58,000
  Dr. Accounts Payable - Office Pro    17,400
      Cr. Bank Account - Main               307,400

  Payment Run: PR-2025-0012 | Posted by: Jane Waweru
```

---

### Cash Position Dashboard

Real-time view of available cash across all accounts:

```
CASH POSITION — As at 06 Feb 2025 09:00 EAT

Account                    Currency    Balance         KES Equivalent
──────────────────────────────────────────────────────────────────────
1120 - Checking Main       KES         1,240,000        1,240,000
1121 - Checking Payroll    KES           350,000          350,000
1111 - Petty Cash          KES            50,000           50,000
1130 - USD Account         USD            15,000        1,987,500 *
1140 - GBP Account         GBP             2,000          329,000 *
──────────────────────────────────────────────────────────────────────
TOTAL CASH (KES equiv.)                                  3,956,500

* Converted at 06 Feb 2025 rates: USD/KES 132.50, GBP/KES 164.50

COMMITTED (next 7 days):
  Payment Run PR-2025-0012 (pending):  (307,400)
  Payroll run (est. 10 Feb):         (2,000,000)
──────────────────────────────────────────────────────────────────────
AVAILABLE POSITION:                                      1,649,100
```

---

### 13-Week Rolling Cash Flow Forecast

Strategic cash visibility 3 months ahead. Updated weekly.

```
13-WEEK CASH FLOW FORECAST
As at: 06 Feb 2025 | Currency: KES

Week    Dates           Receipts    Payments    Net         Closing Balance
─────────────────────────────────────────────────────────────────────────────
W1      10–14 Feb      2,500,000   2,500,000           0     3,956,500
W2      17–21 Feb      1,800,000     900,000     900,000     4,856,500
W3      24–28 Feb      3,200,000   1,200,000   2,000,000     6,856,500
W4      03–07 Mar      1,500,000   2,800,000  (1,300,000)    5,556,500
W5      10–14 Mar      2,800,000   2,100,000     700,000     6,256,500
...
W13     30 Apr–02 May  3,100,000   2,400,000     700,000    10,256,500

MINIMUM BALANCE PROJECTION:  3,956,500 (Week 1)
MAXIMUM BALANCE PROJECTION: 10,256,500 (Week 13)
TARGET MINIMUM BALANCE:       2,000,000 (Policy: 1 month operating expenses)
STATUS: ✓ No cash deficit projected in 13-week horizon
```

**Forecast Data Sources:**
- **Receipts**: AR aging (expected collection dates) + confirmed sales orders + scheduled customer payments
- **Payments**: AP aging (due dates) + approved payment runs + payroll schedule + loan repayment schedule + tax payment dates

**Forecast Accuracy Tracking:**
```markdown
Week W1 (forecast vs. actual):
  Receipts forecast: 2,500,000 | Actual: 2,380,000 | Variance: (120,000) -4.8%
  Payments forecast: 2,500,000 | Actual: 2,495,000 | Variance:     5,000  +0.2%

Variance tracked weekly to improve future forecast accuracy.
```

---

### Cash Pooling (Multi-Entity)

For multi-entity tenants, excess cash in one entity can be swept to reduce borrowing in another.

```markdown
CASH POOL — AWO Group
  Pool Leader: AWO Holdings (holds master account)
  Pool Members: AWO Nairobi, AWO Mombasa, AWO Kisumu

Daily Sweep (automated, 17:00 EAT):
  AWO Nairobi surplus:    +500,000  → swept to pool master
  AWO Mombasa deficit:   (200,000) → funded from pool master
  AWO Kisumu surplus:    +100,000  → swept to pool master
  Net pool contribution: +400,000

Intercompany accounting (automatic on sweep):
  AWO Nairobi:
    Dr. Intercompany Receivable - AWO Holdings  500,000
        Cr. Bank Account - Main                       500,000

  AWO Holdings:
    Dr. Bank Account - Master                   500,000
        Cr. Intercompany Payable - AWO Nairobi        500,000
```

---

## User Roles & Permissions Matrix

### IAM Foundation

The Finance module follows the AWO ERP Module/Resource/Action (MRA) framework. Permission keys are derived as:

```
finance.{resource}.{action}
```

Four actor domains exist:

| Actor | Subject Format | Domain | Use Case |
|-------|---------------|--------|----------|
| Tenant (internal users) | `tenant:{userID}` | `{tenantID}` | All ERP staff |
| Portal (external users) | `portal:{userID}` | `{tenantID}:portal` | Customers, suppliers |
| API (machine clients) | `api:{clientID}` | `{tenantID}:api` | Integrations |
| Platform (ops) | `platform:{userID}` | `_platform_` | AWO support staff |

---

### Finance Module MRA Registry

**Module:** `finance` | Label: Finance | Icon: `fa fa-calculator`

#### Resources & Actions

| Resource Slug | Label | Actions |
|---------------|-------|---------|
| `accounts` | Chart of Accounts | `read` `create` `update` `deactivate` |
| `transactions` | Journal Entries | `read` `create` `submit` `approve` `post` `reverse` `void` |
| `periods` | Accounting Periods | `read` `close` `reopen` `lock` |
| `fiscal-years` | Fiscal Years | `read` `create` `close` `lock` |
| `currencies` | Currencies & Rates | `read` `create` `update` `load-rates` |
| `cost-centers` | Cost Centers | `read` `create` `update` `deactivate` |
| `budgets` | Budgets | `read` `create` `update` `approve` `activate` |
| `bank-accounts` | Bank Accounts | `read` `create` `update` `deactivate` |
| `reconciliation` | Bank Reconciliation | `read` `create` `match` `submit` `approve` `lock` |
| `payments` | Payment Runs | `read` `create` `approve` `post` `cancel` |
| `petty-cash` | Petty Cash | `read` `disburse` `replenish` `count` |
| `tax-config` | Tax Configuration | `read` `create` `update` |
| `intercompany` | Intercompany | `read` `create` `match` `eliminate` |
| `reports` | Financial Reports | `read` `run` `export` `schedule` `share` |
| `settings` | Finance Settings | `read` `update` |

---

### Role Definitions

#### Role Hierarchy (Finance Domain)

```
role:report-viewer
    └── role:finance-viewer
            └── role:finance-manager
                    └── role:cfo
                            └── role:tenant-admin
```

Each role inherits all permissions of the role below it in the chain, plus its own additional grants.

---

#### role:report-viewer

**Description:** Can run and view financial reports. No data mutation. Assigned to non-finance staff who need visibility (e.g., department heads viewing their cost center P&L).

| Resource | Actions Allowed |
|----------|----------------|
| `reports` | `read` `run` `export` |
| `accounts` | `read` |
| `cost-centers` | `read` |
| `periods` | `read` |

---

#### role:finance-viewer

**Description:** Read-only access to all financial data. Inherits `report-viewer`. Suitable for internal auditors during fieldwork or management needing visibility without edit access.

| Resource | Actions Allowed | In addition to report-viewer |
|----------|----------------|------------------------------|
| `transactions` | `read` | ✓ new |
| `budgets` | `read` | ✓ new |
| `bank-accounts` | `read` | ✓ new |
| `reconciliation` | `read` | ✓ new |
| `currencies` | `read` | ✓ new |
| `tax-config` | `read` | ✓ new |
| `payments` | `read` | ✓ new |
| `petty-cash` | `read` | ✓ new |

---

#### role:ap-clerk

**Description:** Accounts payable operations. Creates purchase invoices and processes supplier payments. Cannot approve own transactions or exceed authority limits.

| Resource | Actions Allowed |
|----------|----------------|
| `transactions` | `read` `create` `submit` |
| `payments` | `read` `create` |
| `petty-cash` | `read` `disburse` `replenish` |
| `bank-accounts` | `read` |
| `accounts` | `read` |
| `currencies` | `read` |
| `reports` | `read` `run` `export` |

**Cannot:** Approve, post, or reverse any transaction. Cannot approve own payment run.

---

#### role:ar-clerk

**Description:** Accounts receivable operations. Creates sales invoices and records customer receipts.

| Resource | Actions Allowed |
|----------|----------------|
| `transactions` | `read` `create` `submit` |
| `bank-accounts` | `read` |
| `reconciliation` | `read` `create` `match` |
| `accounts` | `read` |
| `currencies` | `read` |
| `reports` | `read` `run` `export` |

**Cannot:** Approve own transactions. Cannot write off bad debts without Finance Manager approval.

---

#### role:finance-manager

**Description:** Full finance operations. Approves transactions, closes periods, manages reconciliations. Inherits `finance-viewer`.

| Resource | Actions Allowed | Authority Limit |
|----------|----------------|-----------------|
| `transactions` | all: `read` `create` `submit` `approve` `post` `reverse` | Up to 500,000 KES |
| `periods` | `read` `close` `reopen` | Current + prior period only |
| `budgets` | `read` `create` `update` `approve` | — |
| `bank-accounts` | `read` `create` `update` | — |
| `reconciliation` | all actions | — |
| `payments` | `read` `create` `approve` `post` | Up to 500,000 KES |
| `cost-centers` | `read` `create` `update` | — |
| `currencies` | `read` `create` `update` `load-rates` | — |
| `accounts` | `read` `create` `update` `deactivate` | — |
| `reports` | all: `read` `run` `export` `schedule` `share` | — |

**Cannot:** Lock fiscal years. Approve own transactions (self-approval blocked systemically).

---

#### role:cfo

**Description:** Full financial authority. Inherits all of `finance-manager` plus high-value approvals, fiscal year management, and tax configuration.

| Resource | Additional Actions | Authority Limit |
|----------|-------------------|-----------------|
| `transactions` | all — inherited + no amount limit | Unlimited |
| `periods` | `lock` | Any period |
| `fiscal-years` | `read` `close` `lock` | — |
| `tax-config` | `read` `create` `update` | — |
| `settings` | `read` `update` | — |
| `payments` | all — no amount limit | Unlimited |
| `intercompany` | all: `read` `create` `match` `eliminate` | — |

**Approval authority:** All financial transactions without upper limit.

---

#### role:auditor

**Description:** Time-limited read-only access across all finance resources. Typically assigned with `WithExpiry()` for the audit engagement period. Cannot be self-assigned — requires CFO or higher.

| Resource | Actions Allowed |
|----------|----------------|
| All finance resources | `read` only |
| `reports` | `read` `run` `export` |
| `reconciliation` | `read` only |
| `transactions` | `read` only — including reversed/cancelled |

**Temporal:** Always assigned with expiry. Role automatically revoked at expiry (Casbin lazy revocation).

---

#### role:budget-owner

**Description:** Departmental budget responsibility. Can view and propose budget amendments for their cost center(s). Cannot approve own budgets.

| Resource | Actions Allowed |
|----------|----------------|
| `budgets` | `read` `create` (draft only) |
| `cost-centers` | `read` |
| `reports` | `read` `run` (cost-center-scoped) |
| `transactions` | `read` (cost-center-scoped) |

---

#### Portal Roles

| Role | Actor Type | Allowed Actions |
|------|-----------|----------------|
| `role:portal-customer` | `portal:{id}` | View own invoices, download statements, check payment status |
| `role:portal-supplier` | `portal:{id}` | View purchase orders addressed to them, submit invoices, track payment |

Portal actors operate in `{tenantID}:portal` domain — completely isolated from internal ERP data.

---

#### API Roles

| Role | Typical Use | Allowed |
|------|------------|---------|
| `role:api-readonly` | Reporting integrations | `read` on all finance resources |
| `role:api-invoice-submit` | Billing systems | `create` invoices, `read` customers/accounts |
| `role:api-payment-notify` | Bank webhooks | `post` payment receipts, `read` bank-accounts |
| `role:api-full-access` | Trusted service accounts | All actions |

API actors operate in `{tenantID}:api` domain. A compromised API key never escalates to tenant-user rights.

---

### Permissions Quick Reference

| Permission Key | Description |
|---------------|-------------|
| `finance.transactions.read` | View journal entries and GL |
| `finance.transactions.create` | Create draft entries |
| `finance.transactions.submit` | Submit for approval |
| `finance.transactions.approve` | Approve pending entries |
| `finance.transactions.post` | Post to General Ledger |
| `finance.transactions.reverse` | Create reversal entry |
| `finance.periods.close` | Soft/hard close a period |
| `finance.periods.reopen` | Reopen a closed period |
| `finance.periods.lock` | Permanently lock a period |
| `finance.fiscal-years.lock` | Lock a fiscal year |
| `finance.budgets.approve` | Activate a budget version |
| `finance.reconciliation.approve` | Sign off bank reconciliation |
| `finance.reconciliation.lock` | Lock reconciliation permanently |
| `finance.payments.approve` | Approve a payment run |
| `finance.settings.update` | Modify finance module settings |
| `finance.tax-config.update` | Modify tax rates and rules |

---

### Incompatible Role Combinations (SOD)

The system enforces segregation of duties — these pairs cannot be held by the same user:

| Combination | Reason |
|-------------|--------|
| `ap-clerk` + `finance-manager` | Cannot approve own AP invoices |
| `ar-clerk` + `finance-manager` | Cannot approve own AR write-offs |
| Reconciliation preparer + approver | Cannot self-approve bank rec |
| Payment creator + payment approver | Cannot self-approve payment run |
| Budget creator + budget approver | Cannot self-approve budget |
| User admin + transaction approver | Cannot create ghost approvers |

---

## API & Integration Specification

### Authentication

All API requests require a valid Bearer token in the `Authorization` header.

```http
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5...
X-Tenant-ID: a1b2c3d4-0000-0000-0000-000000000001
Content-Type: application/json
```

**Token Types:**

| Type | Subject Format | Domain | Obtain Via |
|------|---------------|--------|-----------|
| Tenant JWT | `tenant:{userID}` | `{tenantID}` | `POST /auth/login` |
| API Key | `api:{clientID}` | `{tenantID}:api` | Tenant admin panel |
| Portal JWT | `portal:{userID}` | `{tenantID}:portal` | `POST /portal/auth/login` |

**Rate Limits:**

| Token Type | Requests/minute | Burst |
|-----------|----------------|-------|
| Tenant JWT | 300 | 500 |
| API Key | 120 | 200 |
| Portal JWT | 60 | 100 |

---

### Base URL & Versioning

```
https://api.awo-erp.com/v1/finance
```

All endpoints are prefixed with `/v1/finance`. Breaking changes increment the version.

---

### Standard Response Envelope

**Success:**
```json
{
  "success": true,
  "data": { ... },
  "meta": {
    "page": 1,
    "per_page": 50,
    "total": 1240,
    "total_pages": 25
  }
}
```

**Error:**
```json
{
  "success": false,
  "error": {
    "code": "PERIOD_CLOSED",
    "message": "Accounting period January 2025 is closed.",
    "field": "transaction_date",
    "docs": "https://docs.awo-erp.com/errors/PERIOD_CLOSED"
  }
}
```

---

### Chart of Accounts

#### `GET /accounts`
List accounts with optional filtering.

**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| `root_type` | string | `asset`, `liability`, `equity`, `revenue`, `expense` |
| `account_type` | string | e.g. `bank`, `receivable`, `payable` |
| `is_active` | bool | Filter active/inactive |
| `is_group` | bool | `false` = leaf accounts only |
| `parent_code` | string | Subtree filter by parent code |
| `q` | string | Search by code or name |
| `page` | int | Page number (default 1) |
| `per_page` | int | Results per page (max 200, default 50) |

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "code": "1120",
      "name": "Checking Account - Main",
      "root_type": "asset",
      "account_type": "bank",
      "normal_balance": "debit",
      "currency": "KES",
      "is_group": false,
      "is_active": true,
      "parent_code": "1110",
      "path": "/1000/1100/1110/1120",
      "level": 4,
      "balance": 1240000.00
    }
  ]
}
```

#### `POST /accounts`
Create a new account. Requires `finance.accounts.create`.

```json
{
  "code": "4310",
  "name": "Consulting Services Revenue",
  "root_type": "revenue",
  "account_type": "income",
  "parent_code": "4300",
  "allow_manual_entries": true,
  "require_reference": false,
  "currency": null
}
```

#### `GET /accounts/{code}`
Get single account with current balance and recent activity.

#### `PATCH /accounts/{code}`
Update mutable fields (name, active status, flags). Code and root_type are immutable after transactions exist.

#### `GET /accounts/{code}/ledger`
GL detail for an account over a date range.

**Query Parameters:** `from`, `to`, `page`, `per_page`

---

### Journal Entries (Transactions)

#### `GET /journal-entries`
List journal entries.

**Query Parameters:**
| Param | Description |
|-------|-------------|
| `status` | `draft`, `pending_approval`, `approved`, `posted`, `reversed`, `cancelled` |
| `type` | `manual`, `system`, `recurring`, `imported`, `integration` |
| `from` | Transaction date from (ISO 8601) |
| `to` | Transaction date to |
| `account_code` | Filter entries touching this account |
| `cost_center_id` | Filter by cost center |
| `reference` | Search by reference number |
| `q` | Full-text search on description |

#### `POST /journal-entries`
Create a new journal entry (status: `draft`).

```json
{
  "transaction_date": "2025-01-31",
  "reference": "CHQ-001234",
  "description": "Office rent payment January 2025",
  "currency": "KES",
  "cost_center_id": "uuid-admin-cc",
  "lines": [
    {
      "account_code": "7310",
      "debit": 50000.00,
      "credit": 0,
      "description": "January 2025 rent",
      "cost_center_id": null
    },
    {
      "account_code": "1120",
      "debit": 0,
      "credit": 50000.00,
      "description": "Payment via cheque"
    }
  ]
}
```

**Response:** `201 Created` with full entry including `id` and `reference_number`.

#### `GET /journal-entries/{id}`
Get a single journal entry with all lines.

#### `POST /journal-entries/{id}/submit`
Submit draft for approval. Requires `finance.transactions.submit`.

#### `POST /journal-entries/{id}/approve`
Approve a submitted entry. Requires `finance.transactions.approve`. Returns `409` if user is the submitter (self-approval blocked).

#### `POST /journal-entries/{id}/post`
Post an approved entry to GL. Requires `finance.transactions.post`.

**Response includes:** `posted_at`, `posting_reference`, updated account balances.

#### `POST /journal-entries/{id}/reverse`
Create a reversal of a posted entry. Requires `finance.transactions.reverse`.

```json
{
  "reversal_date": "2025-02-05",
  "reason": "Incorrectly coded — see JE-2025-0152 for correction"
}
```

**Response:** `201 Created` with the new reversal entry ID.

---

### Accounting Periods

#### `GET /periods`
List all periods for the current fiscal year.

**Response fields:** `id`, `name`, `start_date`, `end_date`, `status` (`open`, `soft_closed`, `hard_closed`, `locked`), `fiscal_year_id`.

#### `POST /periods/{id}/close`
Soft or hard close a period. Requires `finance.periods.close`.

```json
{
  "close_type": "hard",
  "reason": "Month-end close complete — all reconciliations approved"
}
```

#### `POST /periods/{id}/reopen`
Reopen a closed period. Requires `finance.periods.reopen`. Logged to audit trail with mandatory reason.

```json
{
  "reason": "Late vendor invoice received — PO-2025-0089"
}
```

---

### Currencies & Exchange Rates

#### `GET /currencies`
List configured currencies.

#### `POST /rates`
Load exchange rates for a date.

```json
{
  "date": "2025-02-06",
  "rates": [
    { "from": "USD", "to": "KES", "rate": 132.50, "rate_type": "spot" },
    { "from": "GBP", "to": "KES", "rate": 164.50, "rate_type": "spot" },
    { "from": "EUR", "to": "KES", "rate": 143.20, "rate_type": "spot" }
  ]
}
```

#### `GET /rates`
Query historical rates.

**Query Parameters:** `from_currency`, `to_currency`, `date`, `rate_type`

---

### Bank Reconciliation

#### `GET /reconciliations`
List reconciliations by account and period.

#### `POST /reconciliations`
Open a new reconciliation.

```json
{
  "bank_account_id": "uuid",
  "period_id": "uuid",
  "statement_opening_balance": 441350.00,
  "statement_closing_balance": 1225000.00
}
```

#### `POST /reconciliations/{id}/import-statement`
Upload bank statement. `multipart/form-data` with file + format hint.

```
POST /reconciliations/{id}/import-statement
Content-Type: multipart/form-data

file: <statement.csv>
format: csv
column_map: {"date":0,"description":1,"debit":2,"credit":3,"reference":4}
```

#### `GET /reconciliations/{id}/matches`
Get auto-match suggestions. Returns both matched and unmatched items.

#### `POST /reconciliations/{id}/matches`
Confirm or create matches manually.

```json
{
  "matches": [
    {
      "statement_line_id": "uuid",
      "gl_entry_ids": ["uuid1"],
      "match_type": "confirmed"
    },
    {
      "statement_line_id": "uuid2",
      "gl_entry_ids": [],
      "match_type": "timing_difference",
      "note": "Deposit in transit — credited Feb 1"
    }
  ]
}
```

#### `POST /reconciliations/{id}/submit`
Submit balanced reconciliation for approval.

#### `POST /reconciliations/{id}/approve`
Approve and lock. Requires `finance.reconciliation.approve`.

---

### Budgets

#### `GET /budgets`
List budget versions by year.

#### `POST /budgets`
Create a new budget (draft).

```json
{
  "fiscal_year_id": "uuid",
  "name": "FY2025 Original Budget",
  "version_label": "V1",
  "currency": "KES"
}
```

#### `POST /budgets/{id}/lines`
Upload or set budget lines (by account + cost center + period).

#### `POST /budgets/{id}/activate`
Activate a budget version. Requires `finance.budgets.approve`. Deactivates any currently active version.

#### `GET /budgets/{id}/variance`
Budget vs. actual variance report.

**Query Parameters:** `period_id`, `cost_center_id`, `account_code`

---

### Payment Runs

#### `GET /payment-runs`
List payment runs.

#### `POST /payment-runs`
Create a new payment run from due invoices.

```json
{
  "due_on_or_before": "2025-02-07",
  "bank_account_id": "uuid",
  "currency": "KES",
  "invoice_ids": ["uuid1", "uuid2"]
}
```

#### `POST /payment-runs/{id}/approve`
Approve the payment run. Requires `finance.payments.approve`.

#### `POST /payment-runs/{id}/export`
Download payment file for bank upload.

```json
{ "format": "equity_eft" }
```

**Supported formats:** `equity_eft`, `kcb_rtgs`, `swift_mt101`, `generic_csv`

#### `POST /payment-runs/{id}/confirm`
Confirm payments after bank processing and post GL entries.

---

### Financial Reports (Execution)

#### `POST /reports/trial-balance/run`
Run trial balance.

```json
{
  "as_at": "2025-01-31",
  "show_zero_balances": false,
  "cost_center_id": null
}
```

#### `POST /reports/balance-sheet/run`
Run balance sheet with optional comparative period.

```json
{
  "as_at": "2025-01-31",
  "comparative_as_at": "2024-12-31"
}
```

#### `POST /reports/profit-loss/run`
Run income statement.

```json
{
  "from": "2025-01-01",
  "to": "2025-01-31",
  "comparative_from": "2024-01-01",
  "comparative_to": "2024-01-31",
  "cost_center_id": null,
  "show_budget": true
}
```

#### `POST /reports/cash-flow/run`
Run cash flow statement (indirect method).

```json
{
  "from": "2025-01-01",
  "to": "2025-01-31",
  "method": "indirect"
}
```

#### `POST /reports/ar-aging/run`
Accounts receivable aging.

```json
{
  "as_at": "2025-01-31",
  "buckets": [30, 60, 90],
  "currency": "KES"
}
```

#### `POST /reports/ap-aging/run`
Accounts payable aging (same structure as AR).

All report endpoints return:
```json
{
  "success": true,
  "data": {
    "report_id": "uuid",
    "execution_id": "uuid",
    "status": "completed",
    "generated_at": "2025-02-06T09:15:42Z",
    "rows": [ ... ],
    "totals": { ... },
    "download_urls": {
      "pdf":   "https://storage.../report.pdf",
      "excel": "https://storage.../report.xlsx",
      "csv":   "https://storage.../report.csv"
    }
  }
}
```

---

### Webhooks (Outbound Events)

The Finance module emits events that external systems can subscribe to.

#### Event Catalog

| Event | Trigger |
|-------|---------|
| `finance.transaction.posted` | Journal entry reaches POSTED status |
| `finance.transaction.reversed` | Reversal entry created |
| `finance.period.closed` | Period transitions to HARD_CLOSED |
| `finance.period.locked` | Period locked permanently |
| `finance.reconciliation.approved` | Bank rec approved and locked |
| `finance.payment_run.posted` | Payment run GL entries posted |
| `finance.budget.activated` | New budget version activated |

#### Webhook Payload Format

```json
{
  "event": "finance.transaction.posted",
  "tenant_id": "a1b2c3d4-...",
  "timestamp": "2025-02-06T09:15:42Z",
  "data": {
    "id": "uuid",
    "reference_number": "JE-2025-0145",
    "amount": 50000.00,
    "currency": "KES",
    "posted_by": "tenant:usr_001",
    "posted_at": "2025-02-06T09:15:42Z"
  },
  "signature": "sha256=abc123..."
}
```

Webhooks are signed with HMAC-SHA256 using the tenant's webhook secret. Retry on non-2xx: 3 attempts with exponential backoff (5m, 30m, 2h).

---

## Non-Functional Requirements

### Performance

| Operation | Target P95 | Target P99 | Max Acceptable |
|-----------|-----------|-----------|----------------|
| Journal entry create (validation + save draft) | < 200ms | < 500ms | 1s |
| Journal entry post (GL update) | < 500ms | < 1s | 2s |
| Trial balance (up to 500 accounts) | < 2s | < 5s | 10s |
| Balance sheet render | < 3s | < 7s | 15s |
| AR/AP aging (up to 10,000 open items) | < 5s | < 10s | 20s |
| GL detail report (1 account, 1 year) | < 3s | < 5s | 10s |
| Bank statement import (500 lines) | < 5s | < 10s | 30s |
| Auto-match engine (500 lines) | < 10s | < 20s | 60s |
| Exchange rate load (batch of 20 currencies) | < 1s | < 2s | 5s |

**Report Generation:**
- Reports run against a dedicated **read replica** to avoid impacting transactional DB
- Long-running reports (> 30s) are executed asynchronously: `POST /reports/.../run` returns `execution_id`; client polls `GET /report-executions/{id}` for status
- Query timeout hard cap: 120 seconds; returns `timeout` status if exceeded

---

### Scalability

```markdown
TRANSACTION VOLUME TARGETS:

Per Tenant:
  - Up to 500,000 GL entries per fiscal year
  - Up to 10,000 accounts in chart of accounts
  - Up to 500 cost centers
  - Up to 100 concurrent users per tenant

Platform Wide:
  - Up to 1,000 active tenants
  - Up to 50,000 journal entries per minute (platform total)
  - Up to 10,000 report executions per hour

Data Retention at Scale:
  - 7 years of transaction history queryable without archiving
  - GL detail queries must remain < 5s with 3.5M entries (500k × 7 years)
  - Achieved via: account-level partitioning, indexed materialized paths,
    and read-replica routing for all report queries
```

---

### Availability & Reliability

```markdown
UPTIME TARGETS:

Overall API: 99.9% monthly uptime (≤ 43.8 minutes downtime/month)
Report Engine: 99.5% (reports may queue during degraded state)
Background Jobs (reconciliation, rate import): Best-effort

GRACEFUL DEGRADATION:
  - If read replica is unavailable, reports queue (do not fail hard)
  - If exchange rate service is unavailable, last known rate used
    (with warning to user; transaction not blocked)
  - If Temporal scheduler is down, scheduled reports queue until recovery

RECOVERY TARGETS:
  RTO (Recovery Time Objective):  1 hour (DB failover + app restart)
  RPO (Recovery Point Objective): 5 minutes (DB replication lag)

BACKUP:
  - Full DB backup: Daily at 02:00 UTC
  - WAL archiving: Continuous (point-in-time recovery)
  - Backup retention: 30 days full, 7 years WORM archive for compliance
```

---

### Security

```markdown
DATA PROTECTION:
  - All data encrypted at rest (AES-256)
  - All API traffic encrypted in transit (TLS 1.3 minimum)
  - Sensitive fields (bank account numbers) encrypted at column level
  - PII fields (employee name in payroll) encrypted at column level

ROW-LEVEL SECURITY:
  - PostgreSQL RLS enforced on all financial tables
  - Tenant isolation: no query can return rows from another tenant
  - RLS policies verified in automated tests on every migration

SESSION SECURITY:
  - JWT expiry: 1 hour (access token), 7 days (refresh token)
  - Concurrent session limit: 5 sessions per user
  - MFA required for roles with approve/post/lock permissions
  - IP allowlist available for API key clients

AUDIT:
  - Every API request logged: user, endpoint, timestamp, tenant, status
  - Every financial state change logged with before/after values
  - Audit log is append-only (no deletes, no updates)
  - Audit log shipped to SIEM within 60 seconds
```

---

### Data Integrity

```markdown
DATABASE:
  - All financial tables use transactions (ACID compliance)
  - Foreign key constraints enforced on all relationships
  - CHECK constraints on critical fields (status enums, positive amounts)
  - No application-level soft-deletes on posted financial data

IDEMPOTENCY:
  - All POST endpoints accept X-Idempotency-Key header
  - Duplicate requests with same key return original response (no double-post)
  - Idempotency keys stored for 24 hours

CONCURRENCY:
  - Optimistic locking on journal entries (version field)
  - Period close uses advisory locks to prevent concurrent closes
  - Account balance updates use SELECT FOR UPDATE on posting path
```

---

### Compliance & Auditability

```markdown
REGULATORY:
  - Financial data retained for minimum 7 years (configurable to 10)
  - All data deletions require dual authorization + legal hold check
  - GDPR: PII can be anonymized on request (but financial records preserved)
  - SOX: Immutable audit trail, segregation of duties enforced at DB level

EXPORT FOR AUDIT:
  - Full tenant data export available within 24 hours of request
  - Formats: JSON (machine-readable), PDF (human-readable audit package)
  - Includes: all transactions, approvals, configuration changes, user access log
```

---

## Data Import & Export

### Import Overview

All imports follow the same lifecycle:

```
UPLOAD → VALIDATE → PREVIEW → CONFIRM → PROCESS → RESULT
           ↓                               ↓
        REJECTED                        PARTIALLY FAILED
        (schema errors)               (row-level errors reported)
```

Imports are **transactional**: either all rows succeed or none are committed (unless `partial_commit: true` is specified).

---

### Journal Entry Bulk Import

**Use case:** Migrate historical entries, upload payroll data from HR system, import from legacy system.

**Template (CSV):**
```csv
transaction_date,reference,description,account_code,debit,credit,cost_center_code,currency,notes
2025-01-15,CHQ-001234,Office rent January,7310,50000,,CC-ADMIN,,
2025-01-15,CHQ-001234,Office rent January,1120,,50000,CC-ADMIN,,
2025-01-20,INV-0451,Sale to ABC Corp,1210,580000,,CC-SALES,KES,
2025-01-20,INV-0451,Sale to ABC Corp,4110,,500000,CC-SALES,KES,
2025-01-20,INV-0451,Sale to ABC Corp,2310,,80000,CC-SALES,KES,VAT 16%
```

**Rules:**
- Rows with the same `reference` are grouped into one journal entry
- Each group must balance (total debit = total credit)
- `account_code` must exist and be active leaf account
- `transaction_date` must be in an open period
- Missing `currency` defaults to tenant base currency

**API Endpoint:**
```
POST /import/journal-entries
Content-Type: multipart/form-data

file: <entries.csv>
options: {
  "validate_only": false,
  "partial_commit": false,
  "post_immediately": false
}
```

**Response:**
```json
{
  "import_id": "uuid",
  "total_rows": 5,
  "total_entries": 2,
  "valid": 2,
  "invalid": 0,
  "warnings": [],
  "errors": [],
  "entries_created": ["JE-2025-0145", "JE-2025-0146"],
  "status": "completed"
}
```

---

### Chart of Accounts Import

**Template (CSV):**
```csv
code,name,root_type,account_type,parent_code,is_group,allow_manual_entries,require_reference,currency
1000,Assets,asset,root,,,false,false,
1100,Current Assets,asset,group,1000,true,false,false,
1110,Cash & Cash Equivalents,asset,group,1100,true,false,false,
1111,Petty Cash,asset,cash,1110,false,true,false,KES
1120,Checking Account - Main,asset,bank,1110,false,true,false,KES
```

**Validation:**
- Codes must be unique
- Parent must exist before children (import is ordered by hierarchy depth automatically)
- Root type must be consistent with parent's root type
- `is_group: true` accounts may not have `account_type` set to a leaf type

**API Endpoint:**
```
POST /import/accounts
```

---

### Opening Balance Import

**Template (CSV):**
```csv
account_code,debit,credit,currency,notes
1120,1250000,,KES,Opening cash - main account
1210,3450000,,KES,Opening AR per aged debtors list
1310,2100000,,KES,Opening inventory per count sheet
1530,8500000,,KES,Equipment at cost
1535,,2040000,KES,Accumulated depreciation
2110,,1850000,KES,Opening AP per aged creditors list
2510,,3000000,KES,Long-term loan - Equity Bank
3300,,695000,KES,Retained earnings (balancing figure)
```

**Rules:**
- Total debits must equal total credits (hard validation)
- Only balance sheet accounts (asset, liability, equity) — no P&L accounts
- All entries posted to a single opening balance journal entry
- Opening balance period must be open

---

### Budget Import

**Template (CSV):**
```csv
account_code,cost_center_code,jan,feb,mar,apr,may,jun,jul,aug,sep,oct,nov,dec
7110,CC-SALES,1000000,1000000,1000000,1000000,1000000,1000000,1000000,1000000,1000000,1000000,1000000,1000000
7110,CC-OPS,600000,600000,600000,600000,600000,600000,600000,600000,600000,600000,600000,600000
6300,CC-MKT,500000,200000,200000,800000,200000,200000,500000,200000,200000,800000,200000,300000
```

**API Endpoint:**
```
POST /budgets/{id}/import-lines
```

---

### Bank Statement Import

**OFX Format (recommended — most banks support):**
```xml
<OFX>
  <BANKMSGSRSV1>
    <STMTTRNRS>
      <STMTRS>
        <CURDEF>KES</CURDEF>
        <BANKACCTFROM>
          <ACCTID>0123456789</ACCTID>
        </BANKACCTFROM>
        <BANKTRANLIST>
          <DTSTART>20250101</DTSTART>
          <DTEND>20250131</DTEND>
          <STMTTRN>
            <TRNTYPE>DEBIT</TRNTYPE>
            <DTPOSTED>20250105</DTPOSTED>
            <TRNAMT>-50000.00</TRNAMT>
            <FITID>TXN20250105001</FITID>
            <NAME>Office Rent - January</NAME>
            <MEMO>CHQ-001234</MEMO>
          </STMTTRN>
        </BANKTRANLIST>
        <LEDGERBAL>
          <BALAMT>1225000.00</BALAMT>
          <DTASOF>20250131</DTASOF>
        </LEDGERBAL>
      </STMTRS>
    </STMTTRNRS>
  </BANKMSGSRSV1>
</OFX>
```

**MT940 Format (SWIFT):**
```
:20:STMT-20250131
:25:KE0123456789/KES
:28C:00001/001
:60F:C250101KES441350,00
:61:2501050105D50000,00NCHQ001234
:86:Office rent January 2025
:62F:C250131KES1225000,00
```

**CSV Format (configurable column mapping):**
```
POST /reconciliations/{id}/import-statement
Content-Type: multipart/form-data

file: <statement.csv>
format: csv
column_map: {"date":"A","description":"B","debit":"C","credit":"D","reference":"E","balance":"F"}
date_format: DD/MM/YYYY
```

---

### Export Formats

#### GL Export

```
GET /export/general-ledger?from=2025-01-01&to=2025-01-31&format=csv
```

**CSV Output Columns:**
```
entry_date, posting_date, reference, description, account_code, account_name,
debit, credit, balance, cost_center, currency, fx_rate, base_amount,
transaction_type, status, created_by, posted_by, posted_at
```

#### Financial Statements Export

```
POST /reports/balance-sheet/run
→ response includes download_urls.pdf, download_urls.excel, download_urls.csv
```

All statutory report exports include:
- Cover page (company name, period, prepared by, date)
- Comparative column (prior year or prior period)
- Notes section (configurable)
- Digital signature block for CFO sign-off

#### Audit Package Export

Full audit-ready data package for a fiscal year:

```
GET /export/audit-package?fiscal_year_id=uuid&format=zip
```

**Package Contents:**
```
audit-package-FY2025/
  ├── trial-balance.xlsx
  ├── balance-sheet.xlsx
  ├── income-statement.xlsx
  ├── cash-flow.xlsx
  ├── general-ledger-detail.xlsx      (all accounts, all entries)
  ├── ar-aging.xlsx
  ├── ap-aging.xlsx
  ├── journal-entries-detail.xlsx     (all entries with approval chain)
  ├── user-access-log.xlsx            (who had what access, when)
  ├── configuration-changes.xlsx      (COA changes, settings changes)
  └── bank-reconciliations/
      ├── rec-jan-2025.pdf
      ├── rec-feb-2025.pdf
      └── ...
```

---

### Exchange Rate Export

```
GET /export/exchange-rates?from=2025-01-01&to=2025-01-31
```

Returns all loaded rates in date range — required for audit trail of FC transaction valuations.

---

## Intercompany Module

### Overview

The Intercompany module manages financial transactions **between legal entities within the same group**. It ensures:
- Both sides of every intercompany transaction are recorded correctly
- Period-end eliminations are generated automatically
- Intercompany balances net to zero on consolidation
- Intercompany profit on transferred inventory is eliminated

**When it applies:**
- Management fee recharges from head office to subsidiaries
- Loans and cash advances between entities
- Goods or services sold between group companies
- Shared staff costs allocated across entities
- Cash pooling sweeps (see Cash Management section)

---

### Entity Setup

```markdown
Group Structure:
  AWO Holdings Ltd (Parent)
    ├── AWO Nairobi Ltd (Subsidiary 1)
    ├── AWO Mombasa Ltd (Subsidiary 2)
    └── AWO International Ltd (Subsidiary 3, USD functional currency)

Intercompany Account Pairs:
  Each entity pair has dedicated IC accounts:

  AWO Holdings:
    1810 - IC Receivable - AWO Nairobi
    1811 - IC Receivable - AWO Mombasa
    1812 - IC Receivable - AWO International
    2610 - IC Payable - AWO Nairobi
    2611 - IC Payable - AWO Mombasa
    2612 - IC Payable - AWO International

  AWO Nairobi:
    1820 - IC Receivable - AWO Holdings
    2620 - IC Payable - AWO Holdings
```

---

### Intercompany Transaction Types

#### Type 1: Management Fee Recharge

AWO Holdings charges a monthly management fee to subsidiaries for shared services.

```markdown
Monthly Management Fee: 500,000 KES per subsidiary

AWO HOLDINGS BOOKS:
Dr. 1810 - IC Receivable - AWO Nairobi    500,000
    Cr. 4400 - Management Fee Revenue          500,000

AWO NAIROBI BOOKS:
Dr. 7800 - Management Fee Expense         500,000
    Cr. 2620 - IC Payable - AWO Holdings       500,000

System validates:
  IC Receivable (Holdings) = IC Payable (Nairobi) ✓
```

#### Type 2: Intercompany Loan

```markdown
AWO Holdings lends 5,000,000 KES to AWO Mombasa at 10% p.a.

DRAWDOWN:
AWO HOLDINGS:
Dr. 1820 - IC Loan Receivable - Mombasa  5,000,000
    Cr. 1120 - Cash                           5,000,000

AWO MOMBASA:
Dr. 1120 - Cash                          5,000,000
    Cr. 2630 - IC Loan Payable - Holdings     5,000,000

MONTHLY INTEREST:
Rate: 10% / 12 = 41,667 KES/month

AWO HOLDINGS:
Dr. 1820 - IC Loan Receivable - Mombasa     41,667
    Cr. 4510 - Interest Income - IC              41,667

AWO MOMBASA:
Dr. 9100 - Interest Expense - IC            41,667
    Cr. 2630 - IC Loan Payable - Holdings        41,667
```

#### Type 3: Intercompany Goods Transfer

AWO Holdings sells goods to AWO Nairobi (transfer pricing applies).

```markdown
Transfer: 1,000 units at transfer price 400 KES/unit
Cost to Holdings: 300 KES/unit
IC Profit: 100 KES/unit × 1,000 = 100,000 KES

AWO HOLDINGS:
Dr. 1810 - IC Receivable - Nairobi        400,000
    Cr. 5100 - COGS                            300,000
    Cr. 4110 - IC Sales Revenue                100,000

AWO NAIROBI:
Dr. 1310 - Inventory                      400,000
    Cr. 2620 - IC Payable - Holdings           400,000

CONSOLIDATION ELIMINATION:
  Step 1: Eliminate IC sale and purchase
  Dr. 4110 - IC Sales Revenue              400,000
      Cr. 5100 - COGS (in Holdings)            400,000

  Step 2: Eliminate unrealized IC profit in inventory
  (Only if Nairobi has not yet sold the goods to external customers)
  Dr. 3400 - Retained Earnings             100,000
      Cr. 1310 - Inventory (in Nairobi)        100,000
```

---

### Period-End IC Matching

At each period end, the system runs an intercompany matching process:

**Step 1: IC Balance Report**
```
INTERCOMPANY BALANCE REPORT — January 2025

Entity Pair               Receivable (Holdings)  Payable (Nairobi)  Difference
──────────────────────────────────────────────────────────────────────────────
Holdings ↔ Nairobi         5,500,000              5,500,000               0 ✓
Holdings ↔ Mombasa         5,041,667              5,041,667               0 ✓
Holdings ↔ International   2,000,000 KES          $15,503 USD        12,500 ⚠️
──────────────────────────────────────────────────────────────────────────────

⚠️ Holdings ↔ International: KES 12,500 difference
   Cause: Exchange rate movement between invoice date (128.00) and month-end (130.00)
   Action: Post FX revaluation or agree to settlement rate
```

**Step 2: IC Reconciliation Statement**

Each entity pair produces a reconciliation confirming balances agree before the period is closed. This is mandatory — the consolidation cannot proceed with open IC differences.

---

### Consolidation Eliminations

At year-end (or whenever consolidated statements are prepared):

```markdown
ELIMINATION ENTRIES (in consolidation workbook):

1. IC Revenue / IC Expense Elimination
   Dr. IC Sales Revenue (Holdings)      12,000,000
       Cr. IC Purchases/COGS (Nairobi)      12,000,000

2. IC Loan Balance Elimination
   Dr. IC Loan Payable (Mombasa)         5,500,000
       Cr. IC Loan Receivable (Holdings)    5,500,000

3. IC Interest Income / Expense Elimination
   Dr. IC Interest Income (Holdings)       500,000
       Cr. IC Interest Expense (Mombasa)    500,000

4. IC Dividend Elimination
   Dr. Dividend Income (Holdings)        2,000,000
       Cr. Dividends Paid (Subsidiary)      2,000,000

5. Unrealized IC Profit in Inventory
   Dr. Opening Retained Earnings          100,000
       Cr. Inventory                         100,000
   (If goods still in inventory at year-end)

Result: Consolidated statements show only transactions
        with parties OUTSIDE the group.
```

---

### Multi-Currency Intercompany

When entities have different functional currencies (e.g., Holdings in KES, International in USD):

```markdown
SCENARIO: Holdings invoices International for $15,000 USD

Invoice Date (USD/KES: 128.00):
Holdings records: IC Receivable 1,920,000 KES (15,000 × 128)
International records: IC Payable $15,000 USD

Month-End (USD/KES: 132.50):
Holdings revalues: IC Receivable → 1,987,500 KES (15,000 × 132.50)
Unrealized FX Gain: 67,500 KES

CONSOLIDATION:
Holdings IC Receivable: 1,987,500 KES
International IC Payable: $15,000 → 1,987,500 KES at current rate

Difference: 0 ✓ (both revalued at same closing rate)
Translation adjustment posted to OCI (Other Comprehensive Income)
```

---

### IC Netting (Cash Settlement)

Instead of each entity paying separately, netting consolidates all intercompany settlements into a single payment:

```markdown
MONTHLY NETTING — January 2025

AWO Nairobi owes AWO Holdings:       5,500,000 (management fees + goods)
AWO Holdings owes AWO Nairobi:       2,000,000 (services rendered by Nairobi)

NET: AWO Nairobi pays AWO Holdings:  3,500,000 KES

Single payment replaces multiple transactions.

After settlement:
Dr. IC Payable - Holdings        5,500,000
    Cr. IC Receivable - Nairobi      2,000,000
    Cr. Cash                         3,500,000

Both IC accounts cleared to zero ✓
```

---

## Report Catalog

### Report Engine Integration

All financial reports are built on the Dynamic Report Generation System. Each report has a `data_source` pointing to a purpose-built PostgreSQL view on the read replica. Reports can be:
- Run on-demand via UI or API
- Scheduled (daily, weekly, monthly, cron) with email/webhook delivery
- Customized (column selection, filters, grouping) via the Report Builder
- Exported in table, chart, pivot, PDF, Excel, or CSV format

Permission required to run any financial report: `finance.reports.run`

---

### Statutory Financial Statements

| Report Name | Data Source View | Key Parameters | Output Formats |
|-------------|-----------------|---------------|----------------|
| Balance Sheet | `v_balance_sheet` | `as_at`, `comparative_as_at`, `cost_center_id` | PDF, Excel, table |
| Income Statement (P&L) | `v_profit_loss` | `from`, `to`, `comparative_from`, `comparative_to`, `cost_center_id` | PDF, Excel, table |
| Cash Flow Statement | `v_cash_flow` | `from`, `to`, `method` (indirect/direct) | PDF, Excel, table |
| Statement of Changes in Equity | `v_equity_changes` | `fiscal_year_id` | PDF, Excel, table |
| Trial Balance | `v_trial_balance` | `as_at`, `show_zero_balances`, `cost_center_id` | Excel, CSV, table |

**Balance Sheet `v_balance_sheet` columns:**
`section`, `root_type`, `account_group`, `account_code`, `account_name`, `current_balance`, `comparative_balance`, `movement`, `level`, `is_group`

---

### Management & Analysis Reports

| Report Name | Data Source View | Key Parameters | Output Formats |
|-------------|-----------------|---------------|----------------|
| Budget vs. Actual | `v_budget_vs_actual` | `fiscal_year_id`, `period_id`, `cost_center_id`, `budget_version_id` | Excel, table, chart |
| Departmental P&L | `v_cost_center_pl` | `from`, `to`, `cost_center_id` | Excel, table |
| Cost Center Comparison | `v_cost_center_comparison` | `from`, `to`, `group_by_level` | Table, chart |
| Cost Center Trend | `v_cost_center_trend` | `periods` (list), `cost_center_id`, `account_code` | Chart, table |
| Profitability by Division | `v_division_profitability` | `from`, `to` | Table, chart |
| FX Gain/Loss Summary | `v_fx_gainloss` | `from`, `to`, `currency` | Excel, table |
| Intercompany Summary | `v_intercompany_balances` | `as_at`, `entity_pair` | Table, Excel |

---

### Subsidiary Ledger Reports

| Report Name | Data Source View | Key Parameters | Output Formats |
|-------------|-----------------|---------------|----------------|
| GL Detail | `v_gl_detail` | `account_code`, `from`, `to`, `cost_center_id` | Excel, CSV, table |
| Account Statement | `v_account_statement` | `account_code`, `from`, `to` | PDF, Excel |
| AR Aging | `v_ar_aging` | `as_at`, `buckets` [30,60,90], `customer_id` | Excel, PDF, table |
| AP Aging | `v_ap_aging` | `as_at`, `buckets` [30,60,90], `supplier_id` | Excel, PDF, table |
| Customer Statement | `v_customer_statement` | `customer_id`, `from`, `to` | PDF, Excel |
| Supplier Statement | `v_supplier_statement` | `supplier_id`, `from`, `to` | PDF, Excel |
| Outstanding Payments | `v_outstanding_payments` | `as_at`, `currency`, `overdue_only` | Table, Excel |
| Cash Book | `v_cash_book` | `bank_account_id`, `from`, `to` | PDF, Excel |
| Bank Reconciliation Report | `v_bank_rec_summary` | `bank_account_id`, `period_id` | PDF |

---

### Tax & Compliance Reports

| Report Name | Data Source View | Key Parameters | Output Formats |
|-------------|-----------------|---------------|----------------|
| VAT Return Summary | `v_vat_return` | `from`, `to` | PDF, Excel |
| VAT Detail (Input) | `v_vat_input_detail` | `from`, `to`, `supplier_id` | Excel, CSV |
| VAT Detail (Output) | `v_vat_output_detail` | `from`, `to`, `customer_id` | Excel, CSV |
| Withholding Tax Register | `v_wht_register` | `from`, `to`, `payee_id` | Excel, PDF |
| Audit Trail | `v_audit_trail` | `from`, `to`, `user_id`, `action_type` | Excel, CSV |
| Period Close Checklist | `v_period_close_status` | `period_id` | PDF, table |
| Fixed Asset Register | `v_fixed_asset_register` | `as_at`, `asset_class` | Excel, PDF |
| Depreciation Schedule | `v_depreciation_schedule` | `fiscal_year_id`, `asset_class` | Excel |

---

### Cash & Treasury Reports

| Report Name | Data Source View | Key Parameters | Output Formats |
|-------------|-----------------|---------------|----------------|
| Cash Position | `v_cash_position` | `as_at`, `currency` | Table, dashboard |
| 13-Week Cash Forecast | `v_cash_forecast` | `from`, `weeks` (1–26) | Chart, Excel |
| Payment Run History | `v_payment_runs` | `from`, `to`, `bank_account_id` | Table, Excel |
| Outstanding Cheques | `v_outstanding_cheques` | `as_at`, `bank_account_id` | Table, Excel |
| Petty Cash Summary | `v_petty_cash` | `from`, `to`, `custodian_id` | PDF, table |

---

### Exchange Rate Reports

| Report Name | Data Source View | Key Parameters | Output Formats |
|-------------|-----------------|---------------|----------------|
| Exchange Rate History | `v_exchange_rates` | `currency`, `from`, `to`, `rate_type` | Chart, Excel |
| Unrealized FX Positions | `v_unrealized_fx` | `as_at`, `currency` | Table, Excel |
| Realized FX Gain/Loss | `v_realized_fx` | `from`, `to`, `currency` | Table, Excel |
| FC Bank Account Revaluation | `v_fc_revaluation` | `as_at` | Table, Excel |

---

### System Reports (Pre-Built, Non-Customizable)

These reports are `is_system: true` in `report_definitions` — they cannot be deleted or structurally modified by tenants, only filtered.

| Report Name | Refresh | Description |
|-------------|---------|-------------|
| Daily GL Summary | Daily 06:00 | Previous day's posting activity by account |
| Outstanding Approvals | Real-time | All transactions pending approval, by approver |
| Period Close Status | Real-time | Open items blocking period close for each period |
| Budget Exception Report | Daily | Accounts where actuals exceed budget by > 10% |
| IC Balance Discrepancy | Daily | Intercompany pairs with unmatched balances |
| Stale Outstanding Cheques | Weekly | Cheques outstanding > 60 days |
| Audit Log Digest | Daily | All high-risk actions from previous day |

---

### Report Scheduling Examples

```markdown
EXAMPLE 1: Monthly P&L to Management
Report: Income Statement
Schedule: Monthly, Day 10 of each month at 08:00 EAT
Parameters: Prior month (dynamic), comparative prior year
Output: PDF
Delivery: Email → CFO, Finance Manager, CEO
Subject: "[AWO ERP] Income Statement — {month} {year}"

EXAMPLE 2: Daily Cash Position to Treasury
Report: Cash Position
Schedule: Daily, 08:30 EAT (after overnight processing)
Parameters: As at today
Output: Table (email body)
Delivery: Email → CFO, Treasury team

EXAMPLE 3: Weekly AP Aging to Finance
Report: AP Aging
Schedule: Every Monday 07:00 EAT
Parameters: As at previous Friday, buckets 30/60/90
Output: Excel attachment
Delivery: Email → AP Team, Finance Manager

EXAMPLE 4: Audit Log to SIEM via Webhook
Report: Audit Trail
Schedule: Daily 00:05 (prior day's events)
Output: JSON (structured)
Delivery: Webhook → SIEM endpoint
Headers: { "X-Source": "awo-erp-finance", "Authorization": "Bearer ..." }
```

---

