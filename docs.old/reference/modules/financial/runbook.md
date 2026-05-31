# awo ERP — Finance Module Runbook

**Document Type:** Operational Runbook  
**Audience:** Platform Administrators, Tenant Finance Users, Accountants, Cashiers, Approvers, Auditors  
**Module:** Finance  
**Status:** Production  

---

## Preface

### P.1 About This Runbook

This runbook is the authoritative operational guide for the awo ERP Finance Module. It is intended for all users who interact with financial data in any capacity — from daily transaction entry to period-end close and audit review.

This document does not replace formal accounting training or regulatory guidance. It describes how the awo platform implements financial processes and how users should operate within it safely, accurately, and in compliance with internal controls.

### P.2 How to Use This Guide

This runbook is structured for both sequential reading and targeted reference. New users are encouraged to read Chapters 1 through 3 in full before operating the module. Experienced users may navigate directly to the relevant chapter for their task.

Each operational section follows a consistent format:

- **Purpose** — what the section covers and when it applies
- **Prerequisites** — what must be true before you begin
- **Step-by-step procedure** — numbered, actionable instructions
- **Rules and controls** — what is required, permitted, and forbidden
- **Common errors** — known failure modes and how to resolve them

### P.3 Document Conventions

| Convention | Meaning |
|---|---|
| `Menu → Sub-menu → Item` | Navigation path within awo |
| **Bold text** | UI element name (button, field, tab) |
| > Callout | Important note or warning |
| ⚠️ Warning | Action with irreversible or high-impact consequences |
| ✅ Required | Mandatory step or control |
| ❌ Forbidden | Action that must never be performed |

> **Platform vs. Tenant:** Where procedures differ between platform-level administrators and tenant-level users, this is explicitly noted. Assume tenant-level context unless stated otherwise.

### P.4 Version History

| Version | Date | Author | Summary of Changes |
|---|---|---|---|
| 1.0 | — | Finance Operations | Initial release |

---

## Chapter 1 — Introduction to the Finance Module

### 1.1 Module Overview

The awo Finance Module is a fully integrated accounting engine embedded within the awo ERP platform. It provides a complete double-entry bookkeeping system, sub-ledger management, approval workflows, and financial reporting — all within a multi-tenant architecture that allows each organisation (tenant) to maintain independent, isolated financial records on a shared platform infrastructure.

The module is designed around three core principles:

**Accuracy** — Every transaction must balance. The system enforces double-entry accounting at the data layer, preventing imbalanced postings from reaching the ledger.

**Control** — No financial transaction moves from draft to posted without passing through the configured approval workflow. Controls are enforced by the system, not by convention.

**Auditability** — Every action — creation, modification, approval, reversal — is recorded in an immutable audit log with user identity, timestamp, and change detail.

### 1.2 Key Capabilities at a Glance

| Capability | Description |
|---|---|
| General Ledger | Double-entry journals, posting, and ledger inquiry |
| Accounts Receivable | Customer invoicing, receipts, credit notes, collections |
| Accounts Payable | Supplier bills, payment runs, debit notes |
| Banking | Multi-bank account management and reconciliation |
| Tax Management | VAT/GST, withholding tax, tax reporting |
| Multi-Currency | Foreign currency transactions, FX revaluation |
| Approvals | Configurable multi-level approval workflows |
| Period Management | Month-end and year-end close controls |
| Financial Reporting | Standard financial statements and custom reports |
| Audit Trail | Immutable log of all financial activity |

### 1.3 Platform vs. Tenant Context

The awo platform serves multiple tenants — independent organisations — on a single infrastructure. Understanding this distinction is important for all users.

**Platform administrators** operate at the infrastructure level. They can create and configure tenants, manage platform-wide settings, and access cross-tenant reporting where permitted. Platform administrators do not interact with individual tenant financial data in the normal course of operations.

**Tenant users** operate within a single organisation's financial environment. All data they create, view, and manage is scoped to their tenant. A tenant user cannot access another tenant's data under any circumstances.

Most procedures in this runbook apply to tenant users. Platform-specific procedures are clearly marked.

### 1.4 Accounting Standards Alignment

The awo Finance Module is designed to support operations under both **IFRS (International Financial Reporting Standards)** and **GAAP (Generally Accepted Accounting Principles)**. The system does not enforce a specific standard — it provides the tools to comply with either. Responsibility for correct accounting treatment rests with the tenant's finance team and its qualified accountants.

Key structural alignments include:

- Accrual-basis accounting by default
- Period-based ledger management with open/close controls
- Separate recognition of revenue, expenses, assets, liabilities, and equity
- Support for foreign currency transactions per IAS 21 / ASC 830

### 1.5 Integration with Other awo Modules

The Finance Module receives and generates data across the awo platform. Understanding these integrations prevents duplication and confusion.

| Source Module | Finance Integration |
|---|---|
| Procurement | Approved purchase orders generate AP bills |
| Inventory | Goods receipts trigger cost-of-goods postings |
| Sales | Confirmed sales orders generate AR invoices |
| HR & Payroll | Payroll runs post salary and deduction journals |
| Fixed Assets | Depreciation schedules post automatically to GL |
| Projects | Project costs are coded to GL via cost centres |

### 1.6 Data Flow and System Boundaries

Transactions originate in sub-ledgers (AR, AP, Banking) or directly in the General Ledger. All posted transactions are reflected in the GL. Reports aggregate from the GL. The period close locks the GL for a defined date range, protecting the integrity of finalised data.

```
Source Modules → Sub-Ledgers (AR / AP / Banking) → General Ledger → Reports
                                                        ↑
                                              Direct Journal Entries
```

---

## Chapter 2 — Access, Roles and Permissions

### 2.1 User Role Definitions

The Finance Module uses role-based access control (RBAC). Each user is assigned one or more roles that define what they can see and do within the module. Roles should be assigned on the principle of least privilege — users receive only the access they need to perform their job.

#### Finance Administrator

Responsible for configuring the Finance Module for the tenant. This includes chart of accounts setup, tax codes, approval workflow design, and user role assignment. Finance Administrators do not typically perform day-to-day transaction processing.

#### Accountant

Performs the core accounting operations: creating and posting journal entries, processing accruals, managing reconciliations, and producing financial reports. Accountants require a thorough understanding of double-entry accounting and the tenant's chart of accounts.

#### Cashier

Processes incoming receipts and outgoing payments. Cashiers work within the Accounts Receivable and Accounts Payable sub-modules. They do not have access to GL journal entries or period management.

#### Approver

Reviews and approves transactions that are submitted to the approval workflow. Approvers may hold this role in addition to another finance role. Where Segregation of Duties (SoD) is enabled, an approver cannot approve their own transactions.

#### Finance Manager

Oversees the finance function. Responsible for period-end close, exception review, and financial sign-off. The Finance Manager typically holds elevated permissions including the ability to reopen closed periods and override approval queues in emergency situations.

#### Auditor

Read-only access to the full financial history, audit log, and all reports. Auditors cannot create, modify, or approve any transaction. This role is designed for internal audit teams and external auditor access.

### 2.2 Platform-Level vs. Tenant-Level Roles

| Role Type | Scope | Who Assigns |
|---|---|---|
| Platform Administrator | Across all tenants | Platform owner |
| Tenant Finance Administrator | Single tenant | Platform Admin or Tenant Owner |
| All other finance roles | Single tenant | Tenant Finance Administrator |

A user may hold roles in multiple tenants if they work across organisations, but each tenant's data remains isolated.

### 2.3 Permission Matrix

| Action | Finance Admin | Accountant | Cashier | Approver | Finance Manager | Auditor |
|---|---|---|---|---|---|---|
| Configure chart of accounts | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Create journal entries | ❌ | ✅ | ❌ | ❌ | ✅ | ❌ |
| Approve journal entries | ❌ | ❌ | ❌ | ✅ | ✅ | ❌ |
| Process receipts | ❌ | ✅ | ✅ | ❌ | ✅ | ❌ |
| Process supplier payments | ❌ | ✅ | ✅ | ❌ | ✅ | ❌ |
| Approve payments | ❌ | ❌ | ❌ | ✅ | ✅ | ❌ |
| Perform bank reconciliation | ❌ | ✅ | ❌ | ❌ | ✅ | ❌ |
| Close accounting period | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ |
| Reopen closed period | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ |
| View financial reports | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ |
| View audit log | ✅ | ❌ | ❌ | ❌ | ✅ | ✅ |
| Export financial data | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ |
| Manage user roles | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |

### 2.4 Segregation of Duties (SoD)

Segregation of Duties is a fundamental internal control that prevents a single user from having end-to-end control over a financial transaction. When SoD is enabled for a tenant, the system enforces the following restrictions:

- A user who creates a transaction cannot approve it
- A user who initiates a payment cannot also authorise it
- A user who reconciles a bank account cannot be the same user who posted the transactions being reconciled

SoD configuration is managed by the Finance Administrator in `Finance → Settings → Controls → Segregation of Duties`.

> **Note:** If SoD conflicts arise due to staffing constraints (e.g., a small team), the Finance Manager must document and approve exceptions in writing. The Compliance Officer should be notified of any standing SoD exceptions.

### 2.5 Requesting and Revoking Access

**To request access:**

1. The user's line manager submits an access request to the Finance Administrator
2. The Finance Administrator verifies the request is appropriate for the user's role
3. The Finance Administrator creates the user account in `Finance → Settings → Users → New User`
4. The user receives login credentials via secure email and is required to change their password on first login

**To revoke access:**

1. Notify the Finance Administrator immediately when a user leaves or changes roles
2. The Finance Administrator navigates to `Finance → Settings → Users → [User] → Deactivate`
3. Deactivation takes effect immediately — the user's active session is terminated
4. All transactions previously created or approved by the user remain in the audit log

> ⚠️ Access must be revoked on the day of a user's departure. Delayed revocation is a compliance violation and must be reported to the Compliance Officer.

### 2.6 Multi-Tenant User Management

Platform administrators can provision users who work across multiple tenants. Such users log in once and switch between tenant contexts using the tenant selector in the navigation header. Each tenant context is fully isolated — actions performed in Tenant A have no effect on Tenant B.

Finance Administrators should be aware that a user's role may differ between tenants. A user may be an Accountant in one tenant and an Auditor in another.

---

## Chapter 3 — Getting Started and Navigation

### 3.1 Logging In and Workspace Setup

**To log in:**

1. Navigate to your organisation's awo instance URL
2. Enter your registered email address and password
3. Complete multi-factor authentication (MFA) if prompted
4. Select your tenant from the tenant selector if you have access to multiple organisations
5. You will land on the awo main dashboard

**First-time login:**

- You will be prompted to change your temporary password
- Set a strong password of at least 12 characters including uppercase, lowercase, numbers, and symbols
- Enable MFA from `Profile → Security → Enable Authenticator App`

> **Never share your login credentials with anyone — including colleagues, managers, or IT staff.** All actions in the system are attributed to the logged-in user. Sharing credentials creates unresolvable audit trail contamination and is a disciplinary matter.

### 3.2 Finance Dashboard Overview

The Finance Dashboard is the home screen for all finance users. Navigate to it via the main menu: `Finance → Dashboard`.

The dashboard presents a real-time summary of the financial health of the organisation, divided into functional panels:

| Panel | What It Shows |
|---|---|
| Cash Position | Current balances across all bank accounts |
| Pending Approvals | Count of transactions awaiting your review or action |
| Overdue Receivables | Outstanding customer invoices past their due date |
| Upcoming Payables | Supplier bills due within the next 7 and 30 days |
| Recent Activity | Last 10 transactions posted across all sub-ledgers |
| Reconciliation Status | Open/closed status for each bank account this period |
| Period Status | Whether the current accounting period is open or closed |
| Alerts | System-generated warnings requiring attention |

Users see panels relevant to their role. A Cashier, for example, will not see the Reconciliation Status panel by default.

### 3.3 Module Navigation Map

All Finance Module features are accessible from the top navigation menu under **Finance**. The primary navigation structure is:

```
Finance
├── Dashboard
├── General Ledger
│   ├── Journal Entries
│   ├── Chart of Accounts
│   └── Ledger Inquiry
├── Accounts Receivable
│   ├── Invoices
│   ├── Receipts
│   ├── Credit Notes
│   └── Customer Statements
├── Accounts Payable
│   ├── Bills
│   ├── Payments
│   ├── Debit Notes
│   └── Supplier Statements
├── Banking
│   ├── Bank Accounts
│   ├── Reconciliation
│   └── Transfers
├── Tax
│   ├── Tax Rates
│   └── Tax Reports
├── Reports
│   ├── Financial Statements
│   ├── Management Reports
│   └── Custom Reports
├── Period Management
│   ├── Accounting Periods
│   └── Year-End
├── Approvals
│   ├── Pending My Action
│   └── All Approvals
└── Settings (Finance Admin only)
```

### 3.4 Personalising Your Workspace

Users can customise the Finance Dashboard to prioritise the panels most relevant to their role:

1. Click the **Customise** button (top-right of the dashboard)
2. Drag panels to reorder them
3. Toggle panels on or off using the visibility switch
4. Click **Save Layout**

Saved layouts persist across sessions and devices.

### 3.5 Notifications and Alert Management

The Finance Module generates automated notifications for events relevant to your role. Notifications appear in the bell icon (🔔) in the top navigation bar and can also be delivered by email.

| Notification Type | Default Recipients |
|---|---|
| Transaction awaiting your approval | Designated approvers |
| Approval completed | Transaction originator |
| Transaction rejected | Transaction originator |
| Invoice overdue | Accountant, Finance Manager |
| Reconciliation alert | Accountant, Finance Manager |
| Period closing reminder | Finance Manager |
| Failed posting | Transaction originator |

To configure notification preferences: `Profile → Notifications → Finance`.

### 3.6 Quick Actions and Keyboard Shortcuts

The most common actions are available via the **Quick Action** button (+) visible on every Finance Module page. Clicking it presents a shortcut menu to create a new journal entry, record a receipt, or log a payment without navigating through the full menu.

| Shortcut | Action |
|---|---|
| `Alt + J` | New journal entry |
| `Alt + R` | New receipt |
| `Alt + P` | New payment |
| `Alt + A` | View pending approvals |
| `Alt + S` | Global search |
| `Ctrl + S` | Save current form |
| `Esc` | Cancel / close panel |

---

## Chapter 4 — Chart of Accounts and Configuration

### 4.1 Understanding the Chart of Accounts

The Chart of Accounts (CoA) is the structured list of all ledger accounts used by the organisation to record financial transactions. It is the foundation of the entire Finance Module — every journal entry, invoice, payment, and report depends on a correctly configured CoA.

In awo, the CoA is organised hierarchically:

```
Account Group (e.g., Current Assets)
└── Account Category (e.g., Cash and Cash Equivalents)
    └── Ledger Account (e.g., 1001 - Main Operating Account)
```

Each ledger account has a unique code, a descriptive name, and an account type that determines how it behaves in financial statements.

### 4.2 Account Types and Classifications

| Account Type | Normal Balance | Appears In |
|---|---|---|
| Asset | Debit | Balance Sheet |
| Liability | Credit | Balance Sheet |
| Equity | Credit | Balance Sheet |
| Revenue / Income | Credit | Profit & Loss |
| Cost of Sales | Debit | Profit & Loss |
| Expense | Debit | Profit & Loss |
| Tax | Credit (output) / Debit (input) | Balance Sheet |
| Suspense / Control | Either | Balance Sheet |
| Bank | Debit | Balance Sheet + Cash Flow |

> Choosing the wrong account type is one of the most consequential configuration errors. A revenue account misclassified as an asset will produce incorrect financial statements. Account type changes on live accounts require Finance Manager approval and must be reviewed by a qualified accountant.

### 4.3 Creating and Editing Accounts

**To create a new ledger account:**

1. Navigate to `Finance → General Ledger → Chart of Accounts`
2. Click **New Account**
3. Complete the required fields:

| Field | Description |
|---|---|
| Account Code | Unique numeric code (follow your organisation's numbering convention) |
| Account Name | Clear, descriptive name |
| Account Type | Select from the list (see 4.2) |
| Currency | Default transaction currency for this account |
| Tax Treatment | Taxable / Exempt / Out of Scope |
| Cost Centre | Assign if applicable |
| Description | Internal notes for accountants |
| Active | Toggle on to make available for posting |

4. Click **Save**

**To edit an existing account:**

Only inactive accounts can have their account type changed. Active accounts with posted transactions may only have their name and description edited. To make structural changes to an account that has transaction history, consult your Finance Administrator.

❌ Do not delete accounts that have posted transactions. Deactivate them instead using the **Active** toggle.

### 4.4 Cost Centres and Departments

Cost centres allow transactions to be tracked by organisational unit, department, or project. They operate as a secondary dimension alongside the ledger account.

Cost centres are configured at: `Finance → Settings → Cost Centres → New Cost Centre`

When posting a transaction, users select both a ledger account and (optionally) a cost centre. Reports can then be filtered or grouped by cost centre to produce departmental profit and loss statements.

> ✅ Your organisation's Finance Administrator will define which accounts require mandatory cost centre coding. Transactions posted without a required cost centre will be flagged in the validation layer.

### 4.5 Tax Codes and Configuration

Tax codes define how tax is calculated and posted on transactions. Common tax codes include:

| Code | Description | Rate (example) |
|---|---|---|
| STD | Standard VAT / GST | 16% |
| ZERO | Zero-rated | 0% |
| EXEMPT | Exempt from tax | N/A |
| WHT | Withholding tax | Varies |
| OOS | Out of scope | N/A |

Tax codes are configured by the Finance Administrator at `Finance → Tax → Tax Rates → New Rate`. Each code specifies the rate, the input tax account, and the output tax account in the GL.

Accountants and Cashiers select the appropriate tax code when processing invoices and bills. The system automatically calculates and posts the tax amount to the configured tax accounts.

### 4.6 Currency and Exchange Rate Setup

**Base currency** is set during tenant onboarding and cannot be changed after transactions have been posted. All financial statements are reported in the base currency.

**Foreign currencies** are enabled at `Finance → Settings → Currencies → Enable Currency`.

**Exchange rates** are maintained at `Finance → Settings → Currencies → Exchange Rates`. Rates can be:

- **Manual** — entered by a Finance Administrator or Accountant
- **Automatic** — fetched from a configured rate provider on a scheduled basis

> ⚠️ Exchange rate overrides on individual transactions require Finance Manager approval when SoD is enabled. Unauthorised manual rate overrides are an audit finding.

---

## Chapter 5 — General Ledger

### 5.1 Journal Entry Types and When to Use Them

Not every financial event is captured automatically through a sub-ledger. The General Ledger journal entry is the tool for all adjustments, corrections, and transactions that do not originate in AR, AP, or Banking.

| Journal Type | Use Case |
|---|---|
| Standard Journal | One-time adjustments, corrections, reclassifications |
| Accrual Journal | Recording expenses or revenues earned but not yet invoiced |
| Prepayment Journal | Allocating payments made in advance across future periods |
| Depreciation Journal | Monthly depreciation of fixed assets |
| Reversal Journal | Cancelling a previous journal entry |
| Recurring Journal | Repeating entries on a defined schedule (e.g., monthly rent) |
| Opening Balance Journal | Initial balances when migrating to awo |
| Intercompany Journal | Transactions between related entities within a group |

### 5.2 Creating and Drafting Journal Entries

**Prerequisites:**
- You have the Accountant or Finance Manager role
- The relevant accounting period is open
- You have supporting documentation for the entry

**Procedure:**

1. Navigate to `Finance → General Ledger → Journal Entries`
2. Click **New Journal Entry**
3. Complete the journal header:

| Field | Required | Notes |
|---|---|---|
| Journal Type | ✅ | Select from list in 5.1 |
| Posting Date | ✅ | Must fall within an open period |
| Reference Number | ✅ | Must be unique; follow your numbering convention |
| Description | ✅ | Clear business purpose — not just "adjustment" |
| Currency | ✅ | Defaults to base currency |
| Cost Centre | Conditional | Required if account demands it |

4. Add journal lines. Each line requires:

| Field | Required | Notes |
|---|---|---|
| Account Code | ✅ | Select from chart of accounts |
| Description | ✅ | Line-level description |
| Debit Amount | Conditional | Enter debit or credit, not both |
| Credit Amount | Conditional | Enter credit or debit, not both |
| Tax Code | Conditional | If applicable |
| Cost Centre | Conditional | If required by account |

5. Verify the **Totals** panel shows matching debit and credit totals
6. Click **Attach Document** and upload supporting evidence
7. Click **Save as Draft** — the entry is saved but not yet submitted

> ✅ A journal entry must always have at least two lines and total debits must equal total credits before it can be submitted.

**Submitting for approval:**

8. Review the complete entry
9. Click **Submit for Approval**
10. The entry moves to **Pending Approval** status and the designated approver is notified

### 5.3 Supporting Documentation Requirements

Every journal entry must be supported by documentation that justifies the posting. Without documentation, an entry is untraceable in an audit.

| Journal Type | Minimum Required Documentation |
|---|---|
| Adjustment | Written explanation signed by preparer; source document |
| Accrual | Invoice estimate, contract, or management authorisation |
| Prepayment | Supplier invoice or payment receipt |
| Depreciation | Fixed asset register extract |
| Intercompany | Intercompany agreement or matching transaction evidence |
| Correction | Original erroneous entry reference; explanation |

Files are attached directly to the journal entry. Accepted formats: PDF, PNG, JPG, XLSX, DOCX. Maximum file size: 10 MB per attachment.

### 5.4 Recurring and Template Journals

For journals that repeat on a predictable schedule (e.g., monthly rent, depreciation, management fee), use the Recurring Journal feature to avoid manual re-entry.

**To create a recurring journal template:**

1. Create a standard journal entry as described in 5.2
2. Before submitting, click **Save as Recurring Template**
3. Configure the recurrence:
   - Frequency: Daily / Weekly / Monthly / Quarterly / Annually
   - Start date and end date (or number of occurrences)
   - Auto-submit for approval: Yes / No
4. Click **Save Template**

Recurring journals are generated automatically on the configured schedule and placed in Draft status (or submitted automatically if configured). A Finance Manager or Accountant must review auto-generated entries before they are posted.

### 5.5 Accruals and Prepayments

**Accruals** recognise expenses and revenues in the period they are incurred, regardless of when the cash moves. At the end of a period, accountants record accrual journals to capture obligations not yet invoiced.

A standard accrual entry:

```
Dr  Expense Account         [Amount]
    Cr  Accrued Liabilities     [Amount]
```

In the following period, when the actual invoice arrives, the accrual is reversed and the actual bill is processed.

**Prepayments** are the opposite — cash paid in advance for a future benefit. A prepayment is initially recorded as an asset:

```
Dr  Prepaid Expenses        [Amount]
    Cr  Bank Account            [Amount]
```

Each subsequent period, the prepayment is amortised:

```
Dr  Expense Account         [Monthly amount]
    Cr  Prepaid Expenses        [Monthly amount]
```

> ✅ Best practice: Create the amortisation schedule as a recurring journal at the time of the original prepayment posting. This ensures all future periods are provisioned and reduces month-end manual workload.

### 5.6 Depreciation Postings

If the Fixed Assets module is active, depreciation journals are generated automatically based on the depreciation schedule configured per asset. These journals are posted to the GL without manual intervention after Finance Manager approval.

If the Fixed Assets module is not active, depreciation is posted manually:

```
Dr  Depreciation Expense    [Calculated amount]
    Cr  Accumulated Depreciation [Calculated amount]
```

Depreciation calculations must be based on the organisation's stated accounting policy (straight-line, reducing balance, units of production). The supporting document attached to the depreciation journal must include the fixed asset register extract showing the asset, its cost, accumulated depreciation, and the current period charge.

### 5.7 Intercompany and Elimination Entries

Where the tenant is part of a group of related entities, intercompany transactions must be recorded to reflect the economic substance of the transaction and to support group consolidation.

Every intercompany transaction should have a matching and opposite entry in the related entity's books. At group consolidation, these entries are eliminated.

**To post an intercompany journal:**

1. Select journal type: **Intercompany**
2. Specify the related entity in the **Counterparty Entity** field
3. Post the entry as normal — the counterparty entry is not automatically created in the related entity; that entity's accountant must post the corresponding transaction
4. Attach the intercompany agreement or settlement instruction as evidence

Unresolved intercompany imbalances are an audit finding. Finance Managers should run the Intercompany Reconciliation Report at period end: `Finance → Reports → Management Reports → Intercompany Reconciliation`.

### 5.8 Reversing Entries

Reversals are the correct mechanism for correcting a posted journal. Never attempt to edit a posted entry — posted journals are immutable.

**When to use a reversal:**

- An accrual is no longer needed (the actual invoice was received)
- An entry was posted to the wrong account
- A period correction is required

**Reversal procedure:**

1. Navigate to `Finance → General Ledger → Journal Entries`
2. Locate the entry to be reversed
3. Open the entry and click **Reverse Transaction**
4. In the **Reversal Date** field, select the appropriate posting date (typically the first day of the current period)
5. Enter a clear **Reversal Reason** — this is mandatory and appears in the audit log
6. Click **Submit for Approval**

The system creates a new journal entry with all debit and credit lines swapped. The original journal and the reversal are linked in the system and both appear in the audit trail.

❌ Do not create manual reversal journals by hand. Always use the **Reverse Transaction** function. Manual reversals risk errors and break the automatic linkage between original and reversal entries.

---

## Chapter 6 — Approval Workflows

### 6.1 Workflow Architecture and States

Every financial transaction in awo passes through an approval workflow before it is posted to the ledger. The workflow ensures that no single user has unchecked authority over the financial records.

The standard transaction lifecycle is:

```
Draft → Pending Approval → Approved → Posted
                ↓
            Rejected → Returned to Originator (for correction or resubmission)
```

Each state has a defined meaning:

| State | Description |
|---|---|
| Draft | Transaction created but not submitted. Only the creator can see and edit it. |
| Pending Approval | Submitted and awaiting review. Cannot be edited by the originator. |
| Approved | Reviewed and approved. Queued for posting. |
| Posted | Written to the ledger. Immutable — cannot be edited or deleted. |
| Rejected | Returned by the approver with a reason. The originator can correct and resubmit. |
| Cancelled | Withdrawn by the originator before approval. Cannot be reactivated. |

### 6.2 Approval Thresholds and Delegation

Finance Administrators configure approval thresholds that determine how many approval levels a transaction requires based on its amount.

Example configuration (thresholds are tenant-specific):

| Transaction Amount | Approval Levels Required |
|---|---|
| Up to KES 50,000 | Single approver |
| KES 50,001 – 500,000 | Two approvers (Finance Manager as second) |
| Above KES 500,000 | Three approvers (Finance Manager + Director) |

**Delegation:** Approvers who are unavailable (leave, travel) can delegate their approval authority for a defined period. Delegation is configured at `Finance → Approvals → Delegation → New Delegation`. The original approver remains responsible for transactions approved under delegation.

> ⚠️ Permanent or open-ended delegation is not permitted. All delegations must have a defined end date.

### 6.3 Approver Responsibilities Checklist

Before approving any transaction, the approver must verify each of the following:

**For journal entries:**
- [ ] Supporting documentation is attached and legible
- [ ] The posting date falls within an open accounting period
- [ ] Debit and credit totals match
- [ ] Account codes are appropriate for the transaction
- [ ] The description clearly explains the business purpose
- [ ] Cost centre coding is correct (where required)
- [ ] The preparer had authority to create this entry

**For payments:**
- [ ] The supplier exists in the approved vendor master
- [ ] The invoice has not been previously paid (no duplicate)
- [ ] The payment amount matches the approved invoice
- [ ] Bank details match the verified supplier record
- [ ] The payment account has sufficient funds
- [ ] A purchase order or contractual authority exists

**For receipts:**
- [ ] The customer and invoice are correctly identified
- [ ] The amount received matches the remittance advice
- [ ] The payment method and bank account are correctly recorded

> ⚠️ Approving a transaction without completing the verification checklist makes the approver personally accountable for errors or fraud arising from that transaction. Approval is not a rubber stamp.

### 6.4 Rejecting and Returning Transactions

If a transaction fails verification, the approver must reject it and return it to the originator with a clear explanation.

**Rejection procedure:**

1. Open the transaction from `Finance → Approvals → Pending My Action`
2. Review the transaction fully
3. Click **Reject**
4. In the **Rejection Reason** field, write a specific explanation of what is wrong and what correction is needed
5. Click **Confirm Rejection**

The originator is notified immediately. The transaction returns to **Draft** status and can be corrected and resubmitted.

> Rejection reasons must be specific. "Not approved" or "Please check" are not acceptable reasons. Write: "Invoice number not matching purchase order PO-2024-0445. Please verify and resubmit with correct invoice."

### 6.5 Escalation and Timeout Handling

If an approver does not action a pending transaction within the configured timeout period (typically 48 business hours), the system:

1. Sends a reminder notification to the approver
2. After a second timeout period, escalates to the Finance Manager
3. Logs the escalation in the audit trail

Finance Managers can view all overdue approvals at `Finance → Approvals → All Approvals → Overdue`.

### 6.6 Audit Trail of Approvals

Every approval action is recorded in the immutable audit log. The audit record captures:

- User who performed the action
- Action taken (approved, rejected, delegated)
- Timestamp
- Any comments entered
- Transaction details at the time of the action

This log is accessible to Auditors and Finance Managers at `Finance → Approvals → All Approvals → [Transaction] → Audit History`.

---

## Chapter 7 — Accounts Receivable

### 7.1 Customer Master Data Management

Before processing any customer transaction, the customer must exist in the Customer Master. Transacting with customers not in the master creates unallocated receipts and complicates collections.

**To create a customer:**

1. Navigate to `Finance → Accounts Receivable → Customers → New Customer`
2. Complete the customer record:

| Field | Required | Notes |
|---|---|---|
| Customer Name | ✅ | Legal trading name |
| Customer Code | ✅ | Auto-generated or manual |
| Tax ID / PIN | Conditional | Required for tax-registered customers |
| Credit Limit | ✅ | Set to 0 if no credit is extended |
| Payment Terms | ✅ | e.g., Net 30, COD, 14 days |
| Default Currency | ✅ | |
| Contact Details | ✅ | Billing contact name, email, phone |
| Billing Address | ✅ | |
| AR Account | ✅ | Defaults to the AR control account |

3. Click **Save**

Changes to credit limits require Finance Manager approval.

### 7.2 Creating and Issuing Invoices

**Prerequisites:**
- Customer exists in the master
- Goods or services have been delivered or the billing milestone reached
- All pricing and tax codes confirmed

**Procedure:**

1. Navigate to `Finance → Accounts Receivable → Invoices → New Invoice`
2. Select the **Customer**
3. Set the **Invoice Date** and **Due Date** (the system calculates the due date from payment terms)
4. Add invoice lines:

| Field | Required | Notes |
|---|---|---|
| Description | ✅ | Clear description of goods/service |
| Quantity | ✅ | |
| Unit Price | ✅ | |
| Revenue Account | ✅ | |
| Tax Code | ✅ | Select applicable rate |

5. Attach delivery note, contract, or purchase order as supporting evidence
6. Review the total and tax summary
7. Click **Submit for Approval**

Once approved and posted, the invoice is emailed to the customer automatically (if email delivery is configured).

### 7.3 Recording Customer Receipts

**Prerequisites:**
- The customer invoice is posted
- Cash or bank confirmation of payment received

**Procedure:**

1. Navigate to `Finance → Accounts Receivable → Receipts → New Receipt`
2. Select the **Customer**
3. Select the **Bank Account** or **Cash Account** where the payment was received
4. Enter:

| Field | Required | Notes |
|---|---|---|
| Receipt Date | ✅ | Date funds were received, not processed |
| Payment Method | ✅ | Cash / Bank Transfer / Mobile Money / Cheque / Card |
| Reference Number | ✅ | Bank reference or cheque number |
| Amount Received | ✅ | |

5. In the **Invoice Allocation** panel, match the receipt to the outstanding invoice(s)
6. If the amount received is less than the invoice total, allocate the partial amount — the invoice remains outstanding for the balance
7. If the amount exceeds all outstanding invoices, the excess is held as a customer credit or posted to a suspense account (per your organisation's policy)
8. Attach the bank remittance advice or cash receipt slip
9. Click **Submit for Approval**

### 7.4 Credit Notes and Adjustments

Credit notes are issued to reduce the amount owed by a customer — typically for returned goods, billing errors, or agreed discounts after invoicing.

1. Navigate to `Finance → Accounts Receivable → Credit Notes → New Credit Note`
2. Select the **Customer** and, where applicable, the **Original Invoice** being credited
3. Enter the credit note lines (use the same revenue account and tax code as the original invoice)
4. Attach supporting documentation (return note, email approval, etc.)
5. Submit for approval

Once approved and posted, the credit note reduces the customer's outstanding balance. It can be allocated against an existing invoice or held as a credit on the customer account.

### 7.5 Payment Allocation and Matching

All receipts must be fully allocated to specific invoices or approved credits. Unallocated receipts sitting on a customer account distort the AR aging report and must be resolved promptly.

**To allocate an unmatched receipt:**

1. Navigate to `Finance → Accounts Receivable → Receipts`
2. Filter by **Unallocated** status
3. Open the receipt
4. Click **Allocate**
5. Match to the appropriate invoice(s)
6. Save the allocation

> If you cannot identify the correct invoice for a receipt, do not assume. Contact the customer for a remittance advice. If resolution will take more than 24 hours, post the receipt to the designated suspense account with a clear reference, and escalate to the Finance Manager.

### 7.6 AR Aging and Collections Management

The AR Aging Report shows outstanding customer balances categorised by how long they have been overdue.

Navigate to: `Finance → Reports → AR Aging`

| Aging Bucket | Action Required |
|---|---|
| Current (not yet due) | No action — monitor |
| 1–30 days overdue | Send a courteous payment reminder |
| 31–60 days overdue | Follow up by phone; escalate to account manager |
| 61–90 days overdue | Issue a formal demand letter |
| 91+ days overdue | Escalate to Finance Manager; consider legal action or write-off |

Collections activity should be documented in the customer record's notes field for a complete communication history.

### 7.7 Bad Debt Write-Offs

When a customer debt is determined to be irrecoverable after all collection efforts have been exhausted, it must be written off.

> ⚠️ Write-offs require Finance Manager approval and must be supported by documented collection history, a legal opinion or final demand, and in some cases, approval from the Board or executive management.

**Write-off procedure:**

1. Navigate to `Finance → Accounts Receivable → Invoices → [Invoice] → Write Off`
2. Enter the write-off amount and reason
3. The system posts:

```
Dr  Bad Debt Expense        [Amount]
    Cr  Accounts Receivable     [Amount]
```

4. Submit for Finance Manager approval

If the debt is subsequently recovered, post a reversal of the write-off and record the receipt normally.

### 7.8 Customer Statements

Customer statements summarise all transactions (invoices, receipts, credit notes) on a customer account for a given period. They are used for collections and for customer reconciliation queries.

To generate: `Finance → Accounts Receivable → Customer Statements → Generate`

Select the customer(s), statement period, and output format (PDF for email delivery). Statements can be emailed directly to the customer's billing contact from within the module.

---

## Chapter 8 — Accounts Payable

### 8.1 Supplier Master Data Management

All suppliers must be verified and approved in the Supplier Master before any bill or payment is processed. Processing payments to unlisted or unverified suppliers is a significant fraud risk.

**To create a supplier:**

1. Navigate to `Finance → Accounts Payable → Suppliers → New Supplier`
2. Complete the supplier record:

| Field | Required | Notes |
|---|---|---|
| Supplier Name | ✅ | Must match legal name on invoices and bank records |
| Supplier Code | ✅ | Auto-generated or manual |
| Tax ID / PIN | Conditional | Required for tax-registered suppliers |
| Payment Terms | ✅ | Agreed terms from contract |
| Default Currency | ✅ | |
| Bank Account Name | ✅ | Must match verified bank records |
| Bank Account Number | ✅ | Verify directly with supplier — never from an emailed request |
| Bank Name & Branch | ✅ | |
| AP Account | ✅ | Defaults to AP control account |
| Approved By | ✅ | Finance Manager must approve new suppliers |

> ⚠️ **Supplier bank detail changes are a primary vector for payment fraud (Business Email Compromise).** Never update a supplier's bank details based on an email request alone. Call the supplier on their known phone number to verify before making any changes. All bank detail changes require Finance Manager approval and are logged in the audit trail.

### 8.2 Processing Supplier Bills

**Prerequisites:**
- Supplier exists in the approved master
- Original supplier invoice received
- Goods or services confirmed as received

**Procedure:**

1. Navigate to `Finance → Accounts Payable → Bills → New Bill`
2. Select the **Supplier**
3. Enter:

| Field | Required | Notes |
|---|---|---|
| Invoice Date | ✅ | Date on the supplier's invoice |
| Invoice Number | ✅ | As it appears on the supplier's invoice |
| Due Date | ✅ | System calculates from payment terms |
| Bill Lines | ✅ | Expense account, description, quantity, unit cost, tax code |

4. Attach the original supplier invoice (PDF preferred)
5. Run the **Duplicate Check** (the system will warn if the same invoice number exists for this supplier)
6. Submit for approval

### 8.3 Three-Way Matching (PO / GRN / Invoice)

Where the organisation uses purchase orders (POs), all AP bills must be matched against:

1. **Purchase Order** — the authorised commitment to buy
2. **Goods Receipt Note (GRN)** — confirmation that goods or services were received
3. **Supplier Invoice** — the demand for payment

The system performs this matching automatically when the bill is linked to a PO. Discrepancies between the three documents are flagged and must be resolved before the bill can proceed to payment.

| Discrepancy | Action |
|---|---|
| Invoice amount > PO amount | Obtain a PO amendment approval before processing |
| Invoice quantity > GRN quantity | Dispute with supplier; do not pay for goods not received |
| Invoice number mismatch | Confirm with supplier; recheck document |
| GRN missing | Obtain GRN from warehouse/receiving team before proceeding |

### 8.4 Supplier Payments and Payment Runs

**Individual payment:**

1. Navigate to `Finance → Accounts Payable → Payments → New Payment`
2. Select the **Supplier**
3. Select the bill(s) to pay from the outstanding invoice list
4. Confirm the payment amount and due date
5. Select the payment bank account
6. Submit for approval

**Batch payment run** (for processing multiple suppliers at once):

1. Navigate to `Finance → Accounts Payable → Payments → Payment Run`
2. Set the payment date and bank account
3. Filter by due date range (e.g., all invoices due within the next 7 days)
4. Review the generated payment list
5. Deselect any items that should not be included in this run
6. Submit the entire run for Finance Manager approval
7. Once approved, generate the bank payment file for upload to your banking platform

### 8.5 Debit Notes and Supplier Credits

When a supplier owes the organisation money (returned goods, overcharge, agreed discount), record a debit note:

1. Navigate to `Finance → Accounts Payable → Debit Notes → New Debit Note`
2. Select the supplier and original bill reference
3. Enter the credit amount and reason
4. Attach the returns note or supplier credit note
5. Submit for approval

The debit note reduces the outstanding payable and can be offset against future payments to that supplier.

### 8.6 AP Aging and Due Date Management

The AP Aging Report shows what the organisation owes and when it falls due.

Navigate to: `Finance → Reports → AP Aging`

Review this report daily to:
- Identify invoices approaching their due date
- Prioritise the next payment run
- Avoid late payment penalties on supplier contracts
- Identify invoices that have been sitting un-processed and chase outstanding approvals

### 8.7 Duplicate Invoice Detection

The system automatically checks for duplicate invoices when a new bill is created, matching on supplier + invoice number. If a potential duplicate is detected, a warning is displayed.

**Do not override a duplicate warning without investigation.** Open the existing bill and compare. If the duplicate is confirmed:

- Do not post the duplicate
- Note the duplicate reference in the original bill's comments
- Notify the supplier if they have submitted a duplicate

Paying a duplicate invoice is a recoverable error but results in cash flow disruption and reputational damage with the supplier.

### 8.8 Supplier Reconciliation Statements

At period end, reconcile the balance on each significant supplier account against the supplier's own statement.

**Supplier reconciliation procedure:**

1. Obtain the supplier's statement for the period
2. Navigate to `Finance → Accounts Payable → Supplier Statements → [Supplier]`
3. Compare each transaction line with the supplier's record
4. Investigate and resolve any differences:
   - Timing differences (invoice received but not yet processed): post the bill
   - Disputes (goods not received): raise a debit note or contact the supplier
   - Missing credits: chase the supplier for the credit note
5. Document the reconciliation outcome and attach to the period-end working papers

---

## Chapter 9 — Banking and Cash Management

### 9.1 Bank Account Setup and Management

Each physical bank account operated by the organisation must be configured in awo before it can be used for receipts, payments, or reconciliation.

**To add a bank account:**

1. Navigate to `Finance → Banking → Bank Accounts → New Account`
2. Complete:

| Field | Required | Notes |
|---|---|---|
| Account Name | ✅ | Descriptive name (e.g., "KCB Main Operating Account") |
| Bank Name | ✅ | |
| Account Number | ✅ | |
| Sort Code / Branch Code | Conditional | If applicable |
| Currency | ✅ | |
| GL Account | ✅ | Linked bank account in chart of accounts |
| Opening Balance | ✅ | Balance at the date of going live on awo |

3. Finance Manager approval required to activate

### 9.2 Importing Bank Statements

Bank statements can be imported into awo for reconciliation via:

- **Automatic feed** — bank data is fetched directly via an API connection (where supported by your bank)
- **Manual import** — upload a CSV or OFX/BAI2 statement file exported from your online banking platform

**To import a statement manually:**

1. Navigate to `Finance → Banking → Reconciliation → [Account] → Import Statement`
2. Select the statement period
3. Upload the statement file
4. Click **Import**

The system parses the statement and presents each bank line for matching.

### 9.3 Bank Reconciliation Process

Bank reconciliation is the process of matching transactions recorded in awo against the bank's official statement. It is the primary control over cash and detects posting errors, missing transactions, and fraud.

**Reconciliation must be performed at minimum weekly, and daily for high-volume accounts.**

**Procedure:**

1. Navigate to `Finance → Banking → Reconciliation → [Account]`
2. Set the **Statement Date** and confirm the **Closing Balance** per the bank statement
3. Work through the unmatched items:
   - For each bank line, find the matching awo transaction and click **Match**
   - If no match exists, investigate (see 9.4)
4. When all items are matched, the **Difference** field should read zero
5. Click **Approve Reconciliation**
6. The reconciliation is saved and locked. The status changes to **Reconciled**

> ✅ A reconciliation with a non-zero difference must never be approved. Investigate and resolve all differences first.

### 9.4 Handling Unmatched Transactions

| Scenario | Likely Cause | Resolution |
|---|---|---|
| Bank shows a receipt not in awo | Unrecorded receipt or wrong bank account | Post the receipt and match |
| awo shows a receipt not on bank statement | Timing difference (deposit in transit) | Wait for the next statement; if unresolved after 5 days, investigate |
| Bank shows a payment not in awo | Unrecorded payment or bank charge | Post a journal or payment and match |
| awo shows a payment not on bank statement | Timing difference (outstanding cheque) | Wait and monitor |
| Bank shows an amount different from awo | Partial payment or bank fee | Investigate; create adjustment journal for difference |
| Bank charges or interest | These are not typically pre-posted in awo | Create a bank charge journal and match |

### 9.5 Petty Cash Management

Petty cash funds should be established, controlled, and reconciled independently from bank accounts.

**Imprest system:** A fixed float is established (e.g., KES 10,000). When petty cash is spent, receipts are collected. When the float runs low, a reimbursement payment is requested — equal to the total receipts — restoring the float to its original value.

**Petty cash reconciliation (daily):**

1. Count physical cash
2. Total all receipts held
3. Cash + Receipts = Imprest amount (if not, investigate the discrepancy immediately)
4. Post petty cash expenditure journal from the receipts
5. Request reimbursement through the normal AP payment process

Petty cash accounts should be reviewed and formally reconciled by the Finance Manager at minimum monthly.

### 9.6 Inter-Account Transfers

Funds transferred between the organisation's own bank accounts must be recorded in awo to prevent double-counting.

1. Navigate to `Finance → Banking → Transfers → New Transfer`
2. Select the **From Account** and **To Account**
3. Enter the transfer amount and date
4. Add a reference (e.g., bank transfer reference number)
5. Submit for approval

The system creates two entries: a debit on the source account and a credit on the destination account. Both will appear as unmatched items in each account's reconciliation until the bank statement confirms both sides.

### 9.7 Bank Charges and Interest Postings

Bank charges and interest received are typically identified during bank reconciliation. Post them as follows:

**Bank charge:**
```
Dr  Bank Charges Expense    [Amount]
    Cr  Bank Account            [Amount]
```

**Interest received:**
```
Dr  Bank Account            [Amount]
    Cr  Interest Income          [Amount]
```

These can be posted directly from the reconciliation screen using the **Quick Journal** function, which pre-populates the bank account line.

---

## Chapter 10 — Multi-Currency Operations

### 10.1 Functional vs. Presentation Currency

**Functional currency** is the currency of the primary economic environment in which the organisation operates. It is set at tenant configuration and drives all accounting calculations.

**Presentation currency** is the currency in which financial statements are presented, which may differ from the functional currency for group reporting purposes.

All transactions are recorded in the transaction currency and automatically translated to the functional currency at the applicable exchange rate.

### 10.2 Exchange Rate Management

Exchange rates are maintained at `Finance → Settings → Currencies → Exchange Rates`.

Three rate types are supported:

| Rate Type | Used For |
|---|---|
| Spot Rate | Individual transaction translation (applied at transaction date) |
| Average Rate | Monthly P&L translation for period reports |
| Closing Rate | Balance sheet translation at period end |

> ⚠️ Manually overriding an exchange rate on a transaction bypasses the approved rate table. This requires Finance Manager approval and is logged in the audit trail. Rate manipulation is a red flag in financial audits.

### 10.3 Foreign Currency Transactions

When posting a transaction in a foreign currency:

1. Select the transaction currency in the currency field
2. The system displays the current exchange rate from the rate table
3. Enter the amount in the transaction currency
4. The system calculates and displays the functional currency equivalent
5. If the rate needs to be adjusted (e.g., for a hedged transaction), enter the override rate and obtain approval

### 10.4 Unrealised and Realised FX Gains and Losses

**Unrealised FX difference:** Arises when an outstanding foreign currency balance (e.g., an unpaid invoice) is revalued at a different rate than when it was posted. The gain or loss is recorded but has not yet been settled in cash.

**Realised FX difference:** Arises when a foreign currency transaction is settled (payment received or made) at a different rate than the original invoice rate. The difference represents actual cash gain or loss.

The system posts these automatically:

- Unrealised FX: generated during period-end revaluation
- Realised FX: generated at the time of payment or receipt

Both post to the configured FX Gain/Loss accounts in the chart of accounts.

### 10.5 Period-End Revaluation

At month end, all outstanding foreign currency balances (unpaid invoices, outstanding bills, bank balances in foreign currencies) must be revalued at the period closing rate.

1. Navigate to `Finance → General Ledger → FX Revaluation`
2. Select the accounting period
3. Confirm the closing exchange rates are up to date
4. Click **Run Revaluation**
5. Review the generated journals
6. Submit for approval

Revaluation journals are automatically reversed on the first day of the following period, replacing them with the new period's rates.

### 10.6 Multi-Currency Reporting

Financial reports can be presented in:

- **Functional currency** (default)
- **Transaction currency** (for specific accounts or sub-ledger reports)
- **Presentation currency** (for group consolidation reports)

Select the currency presentation option in the report filters before generating.

---

## Chapter 11 — Tax Management

### 11.1 Tax Configuration and Rates

Tax codes are configured by the Finance Administrator. Each code defines:

- Tax name and code
- Rate (percentage)
- Tax type (VAT / GST / Withholding / Customs)
- Input tax account (for purchases)
- Output tax account (for sales)
- Effective date range

Rates are configured at `Finance → Tax → Tax Rates`. Changes to rates (e.g., due to a statutory rate change) must be dated correctly — the system applies the rate in effect on the transaction date.

### 11.2 VAT / GST on Transactions

When a tax code is applied to an invoice or bill line, the system:

1. Calculates the tax amount based on the net amount and the applicable rate
2. Displays the tax breakdown on the document
3. Posts the tax to the configured input or output tax account at the time of posting

The tax amount is not an expense or income to the organisation — it is a liability (output tax collected) or an asset (input tax recoverable). The net amount between output and input tax is periodically remitted to the tax authority.

### 11.3 Withholding Tax Handling

Withholding tax (WHT) is deducted by the paying party at the time of payment and remitted directly to the tax authority on behalf of the recipient.

**When receiving a payment subject to WHT:**

The customer deducts WHT before paying. In awo, record:
```
Dr  Bank Account            [Net amount received]
Dr  WHT Receivable          [WHT amount]
    Cr  AR Invoice              [Gross invoice amount]
```

The WHT receivable represents a tax credit that can be offset against the organisation's own tax liability.

**When making a payment subject to WHT:**

Deduct WHT before paying the supplier:
```
Dr  AP Invoice              [Gross invoice amount]
    Cr  Bank Account            [Net payment]
    Cr  WHT Payable             [WHT amount]
```

The WHT Payable must be remitted to the tax authority by the statutory deadline.

### 11.4 Tax Reports and Returns

Navigate to `Finance → Tax → Tax Reports` to generate:

| Report | Purpose |
|---|---|
| VAT/GST Return | Summary of output and input tax for the period |
| WHT Summary | Total withholding tax deducted and payable |
| Tax Audit Trail | Transaction-level detail of all taxed entries |
| Tax Reconciliation | Comparison of tax account balances to return figures |

Always reconcile the tax account balances in the GL against the tax return figures before filing. Discrepancies indicate posting errors that must be corrected before submission.

### 11.5 Tax Audit Trail

The tax audit trail records every transaction that had a tax code applied, including the tax code used, the rate, and the calculated amount. This report is the primary document produced for a tax authority audit.

Navigate to: `Finance → Tax → Tax Audit Trail`

This report is read-only and cannot be modified. It reflects the posted state of all transactions.

---

## Chapter 12 — Financial Reporting

### 12.1 Standard Financial Statements

The following financial statements are available and generated directly from the posted General Ledger:

| Statement | Description | Navigate To |
|---|---|---|
| Trial Balance | All ledger account balances for a period | `Reports → Financial Statements → Trial Balance` |
| Balance Sheet | Assets, liabilities, and equity at a point in time | `Reports → Financial Statements → Balance Sheet` |
| Profit & Loss | Revenue and expenses for a period | `Reports → Financial Statements → Profit & Loss` |
| Cash Flow Statement | Cash movements by operating, investing, and financing activities | `Reports → Financial Statements → Cash Flow` |
| Statement of Changes in Equity | Movements in equity accounts | `Reports → Financial Statements → Equity Statement` |

**Before generating any statement for external distribution:**

- [ ] Ensure the period is fully reconciled
- [ ] Confirm all journals are posted
- [ ] Verify the trial balance is balanced (total debits = total credits)
- [ ] Obtain Finance Manager sign-off

### 12.2 Management Reports and Dashboards

In addition to statutory statements, the Finance Module provides management-focused reports for internal decision-making:

| Report | Description |
|---|---|
| AR Aging Analysis | Customer balances by aging bucket |
| AP Aging Analysis | Supplier balances and upcoming due dates |
| Cash Position Summary | Current and projected cash by bank account |
| Revenue by Cost Centre | Departmental income breakdown |
| Expense Analysis | Spend by category and cost centre |
| Intercompany Reconciliation | Balances between related entities |
| Budget vs. Actual | Variance analysis against approved budget (if budgets are loaded) |

### 12.3 Custom Report Builder

Users with the Accountant or Finance Manager role can build custom reports using the report builder.

1. Navigate to `Finance → Reports → Custom Reports → New Report`
2. Select the **Data Source** (GL transactions, AR, AP, etc.)
3. Choose **Dimensions** (date, account, cost centre, customer, supplier)
4. Choose **Measures** (debit, credit, net balance, count)
5. Apply **Filters** (date range, account range, cost centre)
6. Set **Grouping and Sorting**
7. Preview and save the report with a descriptive name

Saved custom reports are available to all users with reporting access. Sensitive custom reports (e.g., payroll cost detail) should be restricted using the report-level access control.

### 12.4 Scheduled and Automated Reports

Recurring reports can be scheduled to generate and deliver automatically.

1. Open any standard or custom report
2. Click **Schedule Report**
3. Set the frequency, delivery time, and recipient list
4. Choose the output format (PDF / XLSX / CSV)
5. Save the schedule

Scheduled reports are delivered by email. Recipients do not need to be awo users.

> ⚠️ Scheduled reports that contain sensitive financial data must only be delivered to authorised recipients. Review scheduled report recipient lists quarterly.

### 12.5 Exporting and Distributing Reports

**To export a report:**

1. Generate the report with the correct parameters
2. Click **Export** and select the format (PDF / XLSX / CSV)
3. Save the file securely

**Distribution controls:**

- Financial reports are confidential. Only distribute to authorised parties
- When sharing externally (auditors, investors, regulators), use the finalised, signed version
- Never share draft or unreconciled reports externally
- Use secure file transfer (encrypted email or a document portal) for sensitive reports

### 12.6 Report Governance and Version Control

The Finance Module maintains a record of every report that was generated, exported, or scheduled — including who generated it, when, and with what parameters. This is accessible at `Finance → Reports → Report Activity Log`.

When producing the same report at different points in time (e.g., monthly P&L packs), save a copy of each month's output in the designated document management system. Do not rely solely on the ability to re-run the report, as retrospective changes (corrections, period adjustments) may alter the figures.

---

## Chapter 13 — Period Management and Close Process

### 13.1 Accounting Calendar Setup

The accounting calendar defines the organisation's financial year and the periods within it. It is configured during initial setup at `Finance → Settings → Accounting Calendar`.

Standard configurations:

| Calendar Type | Description |
|---|---|
| Calendar Year | 12 months, January to December |
| Fiscal Year (non-calendar) | 12 months, any start month (e.g., April to March) |
| 52/53 Week Year | Used in some retail organisations |

Periods can be monthly, quarterly, or semi-annual. The Finance Administrator creates the calendar for each financial year before the year begins.

### 13.2 Month-End Close Checklist

The month-end close is a structured sequence of activities that ensures the accounting period is complete and accurate before it is locked. The Finance Manager owns the close process and should use this checklist to track completion.

**Week 1 (First week of the new month) — Transactional Completion:**

- [ ] All sales invoices for the closed month posted to AR
- [ ] All supplier bills received during the month processed in AP
- [ ] All customer receipts and supplier payments posted
- [ ] Bank reconciliations completed for all accounts
- [ ] All inter-account transfers matched and reconciled
- [ ] Petty cash reconciled and reimbursement processed

**Week 1 — Adjustments and Accruals:**

- [ ] Accrual journals prepared for expenses incurred but not yet invoiced
- [ ] Prepayment amortisation journals posted
- [ ] Depreciation journals posted (or auto-posting confirmed)
- [ ] Intercompany transactions confirmed with related entities
- [ ] FX revaluation run for all foreign currency balances
- [ ] WHT payable confirmed and remittance scheduled

**Week 1 — Review and Validation:**

- [ ] Trial balance reviewed — total debits equal total credits
- [ ] All suspense account balances investigated and cleared
- [ ] AR aging reviewed — collections actions initiated for overdue items
- [ ] AP aging reviewed — no overdue invoices unpaid without reason
- [ ] All pending approvals resolved (no transactions left in Pending Approval status)
- [ ] Budget vs. actual report reviewed and significant variances explained

**Final Step — Period Close:**

- [ ] Finance Manager reviews and signs off the checklist
- [ ] Period is closed (see 13.3)
- [ ] Financial statements generated and saved to document management system

### 13.3 Year-End Close Procedures

Year-end close follows the same logic as month-end but includes additional steps:

1. Complete all 12 months of reconciliation and close
2. Post year-end adjustment journals (audit adjustments, provisions, write-offs)
3. Run the **Year-End Trial Balance** and agree to the external auditor's figures
4. Post the **Retained Earnings Transfer**:

```
Dr  Profit & Loss Summary    [Net profit for the year]
    Cr  Retained Earnings        [Net profit for the year]
```

5. Close the final period of the financial year
6. Open the new financial year at `Finance → Period Management → New Year`
7. Roll forward opening balances (the system does this automatically when the new year is opened)

### 13.4 Period Validation and Error Resolution

Before closing a period, the system runs a validation check. Common validation failures:

| Validation Failure | Cause | Resolution |
|---|---|---|
| Unposted transactions | Transactions in Draft or Pending Approval status | Post or cancel all outstanding items |
| Unreconciled bank accounts | One or more accounts not reconciled | Complete reconciliation |
| Suspense account balances | Funds sitting in suspense at period end | Investigate and clear all suspense balances |
| Imbalanced journals | A journal with unequal debits and credits (should not occur but can in edge cases) | Identify and correct the journal |
| Unallocated receipts | Receipts not matched to invoices | Allocate or post to an appropriate account |

All validation failures must be resolved before the period can be closed.

### 13.5 Reopening a Closed Period

Once a period is closed, it can only be reopened by the Finance Manager. Reopening a closed period is an exceptional action that is logged in full in the audit trail.

> ⚠️ Reopening a closed period should be treated as a significant event. It must be accompanied by written justification (e.g., audit adjustment, statutory correction) and Finance Manager sign-off. If the period has been closed for more than 30 days, Compliance Officer notification may be required per your organisation's policy.

**To reopen a period:**

1. Navigate to `Finance → Period Management → [Period] → Reopen Period`
2. Enter the reason for reopening
3. Confirm
4. Make the necessary corrections
5. Close the period again following the standard checklist

### 13.6 Retained Earnings and Year Rollover

When a new financial year is opened, the system:

1. Creates the new year's accounting periods
2. Carries forward all balance sheet account balances as opening balances
3. Resets all P&L account balances to zero (the net of the prior year P&L is transferred to Retained Earnings)

This process is automatic. The Finance Manager should verify the opening trial balance of the new year against the closing trial balance of the prior year before any new transactions are posted.

---

## Chapter 14 — Suspense Accounts and Clearing

### 14.1 Purpose and Correct Use of Suspense Accounts

Suspense accounts are temporary holding accounts used when:

- A transaction must be posted immediately but the correct account is not yet determined
- A receipt has been received but cannot yet be matched to an invoice
- An entry requires further investigation before final coding

Suspense accounts are not permanent resting places. Every balance in a suspense account represents an unresolved transaction that must be investigated and cleared.

> ❌ Do not use suspense accounts to avoid making a decision on a transaction's correct coding. If you are unsure of the correct account, escalate to your Accountant or Finance Manager before posting.

### 14.2 Monitoring and Aging of Suspense Balances

The Finance Module tracks all suspense account postings and their age. Finance Managers and Accountants should review the Suspense Account Report daily.

Navigate to: `Finance → Reports → Management Reports → Suspense Account Analysis`

This report shows:

- Each suspense account
- All uncleared transactions with their posting date
- The age of each balance (days outstanding)
- Responsible user (who posted it)

Balances over 30 days must be escalated. Balances over 90 days are a compliance concern.

### 14.3 Clearing Procedures and Timelines

| Balance Age | Required Action |
|---|---|
| 0–7 days | Normal — investigate and clear within 7 days |
| 8–30 days | Escalate to Finance Manager; document reason for delay |
| 31–90 days | Finance Manager must review and sign off on continued holding |
| 90+ days | Escalate to Compliance Officer; mandatory resolution plan required |

**To clear a suspense balance:**

1. Identify the correct account for the transaction
2. Create a journal entry reclassifying the amount from the suspense account to the correct account
3. Attach documentation supporting the final coding decision
4. Submit for approval

### 14.4 Escalation When Balances Are Unresolved

If a suspense balance cannot be resolved because of a third party (e.g., awaiting customer remittance advice, awaiting supplier invoice):

1. Document the reason and expected resolution date in the transaction's notes field
2. Notify the Finance Manager
3. Set a follow-up task or calendar reminder
4. If the balance remains unresolved at period end, the Finance Manager must decide whether to leave it in suspense (with documented justification) or reclassify it to the most likely account with a note for reversal when the issue is resolved

---

## Chapter 15 — Internal Controls and Audit Compliance

### 15.1 Internal Control Framework Overview

The awo Finance Module is built on a layered internal control framework:

**Preventive controls** stop errors before they occur:
- Mandatory dual-entry accounting enforced at the data layer
- Approval workflows before any transaction is posted
- SoD restrictions preventing self-approval
- Closed period protection preventing retrospective modification

**Detective controls** identify errors after the fact:
- Bank reconciliation
- AR and AP aging review
- Trial balance review
- Suspense account monitoring
- Audit log review

**Corrective controls** fix errors once detected:
- Reversal mechanism for posted entries
- Credit notes and debit notes for sub-ledger corrections
- Period reopening for significant corrections (with approval)

### 15.2 User Responsibilities and Accountability

Every finance user is personally accountable for the transactions they create and approve. The system's audit trail is irrefutable — it records every action with the user's identity and timestamp.

**All finance users must:**

- Use only their own login credentials
- Create transactions only for legitimate business purposes
- Attach accurate and complete supporting documentation
- Apply the correct accounts, tax codes, and cost centres
- Follow the approval workflow — never attempt to bypass it
- Report errors, suspicious activity, or suspected fraud immediately

**Finance users must never:**

- Share login credentials
- Approve transactions they originated (where SoD is enabled)
- Create backdated entries without Finance Manager approval
- Process transactions for personal benefit
- Discuss financial data with unauthorised parties
- Attempt to modify or delete posted transactions

### 15.3 Audit Log and Transaction History

The audit log is immutable. Once written, it cannot be altered by any user — including Finance Administrators and Platform Administrators.

Every record in the audit log contains:

- Entity type (journal, invoice, payment, etc.)
- Entity ID and reference number
- Action performed (created, edited, submitted, approved, rejected, posted, reversed)
- User who performed the action
- Timestamp (UTC)
- IP address
- Before and after values for any field that was changed

Auditors access the audit log at `Finance → [Transaction] → Audit History` or via the full audit log export at `Finance → Settings → Audit Log`.

### 15.4 Supporting Document Standards

Every transaction must have adequate supporting documentation. The following standards apply:

| Document Type | Standard |
|---|---|
| Invoices | Original supplier invoice (not a copy of a copy) in PDF or clear image |
| Receipts | Bank remittance advice, payment confirmation, or cash receipt |
| Journals | Written explanation signed by preparer; source calculation workings |
| Contracts | Signed agreement or email approval from authorised signatory |
| Bank statements | Official bank statement — not a screenshot of online banking |
| Expense claims | Original receipts (not photocopies); completed expense claim form |

Documents must be:
- Legible (not blurred, cropped, or altered)
- Dated and referenced
- Attached at the time of transaction creation, not retrospectively

### 15.5 Fraud Indicators and Red Flags

Finance users should be alert to the following indicators of potential fraud:

**In Accounts Payable:**
- New supplier with similar name to an existing supplier
- Bank detail change request received by email
- Invoice with no PO, round numbers, or sequential invoice numbers
- Supplier address matching an employee's address
- Unusually frequent payments to a new supplier

**In Accounts Receivable:**
- Unusual credit notes or write-offs, especially near period end
- Receipts that cannot be matched to invoices
- Significant overpayments by customers

**In the General Ledger:**
- Journals posted late at night, on weekends, or on public holidays without explanation
- Entries with vague descriptions ("misc", "adjustment", "correction")
- Large round-number entries without supporting documentation
- Reversals immediately before period close

If you observe any of these indicators: **do not confront the suspected individual directly.** Report immediately and confidentially to the Compliance Officer or Finance Manager.

### 15.6 External Auditor Access

During financial audits, external auditors require read-only access to the Finance Module. The Finance Administrator can provision a temporary Auditor role for external auditors:

1. Navigate to `Finance → Settings → Users → New User`
2. Assign the **Auditor** role
3. Set an account expiry date aligned with the audit timeline
4. Provide credentials to the audit firm's lead partner via secure channel
5. Revoke access immediately upon completion of the audit

---

## Chapter 16 — Security and Data Governance

### 16.1 Credential and Password Policy

All finance users must comply with the following password policy:

| Requirement | Standard |
|---|---|
| Minimum length | 12 characters |
| Complexity | At least one uppercase, one lowercase, one number, one symbol |
| Reuse | Cannot reuse the last 12 passwords |
| Maximum age | 90 days (system enforces a change prompt) |
| MFA | Mandatory for all finance roles |
| Failed login lockout | Account locked after 5 consecutive failures |

Locked accounts must be unlocked by the Finance Administrator or System Administrator — users cannot unlock their own accounts.

### 16.2 Data Confidentiality Obligations

Financial data is among the most sensitive information held by any organisation. All finance users are bound by confidentiality obligations that apply during and after their employment.

**Permitted sharing:**
- Within the finance team, on a need-to-know basis
- With external auditors under a formal engagement letter
- With regulators as required by law
- With management as authorised by the Finance Manager

**Not permitted:**
- Sharing financial data with personal email accounts
- Discussing financial performance with non-authorised colleagues
- Publishing or posting any financial information externally
- Removing financial records from the organisation's systems

### 16.3 Session Management and Auto-Logout

The system automatically logs out inactive sessions after 20 minutes. Users working on long-form entries (e.g., a complex journal with many lines) should save drafts regularly to prevent loss of work.

When leaving a workstation, always log out manually: `Profile → Sign Out` or use the keyboard shortcut `Ctrl + Shift + L`.

**Shared workstations** (e.g., cashier terminals) carry elevated risk. Always log out after each transaction. Never allow another person to use the terminal without logging out and back in under their own credentials.

### 16.4 Data Retention and Archiving

Financial records must be retained for the period required by applicable law and regulation. In most jurisdictions, this is a minimum of seven years for transaction records and ten years for certain statutory documents.

awo retains all financial data on its servers for the configured retention period. Archived data (from prior years beyond the active period) is accessible in read-only mode at `Finance → Archives`.

Finance Administrators should confirm the applicable retention period with their legal and compliance team and configure the retention settings accordingly at `Finance → Settings → Data Retention`.

### 16.5 Reporting Security Incidents

If you suspect a security incident — including unauthorised access to finance data, a compromised credential, or suspected data exfiltration — report it immediately:

1. Do not attempt to investigate or resolve the incident yourself
2. Notify your Finance Manager and System Administrator immediately
3. Document what you observed (time, what you saw, which system or account)
4. Preserve any evidence (do not delete emails, do not log out of systems)
5. Follow the organisation's Incident Response Policy

---

## Chapter 17 — Error Handling and Troubleshooting

### 17.1 Common Errors and Root Causes

#### Error: "Debits and Credits Do Not Match"

**Cause:** One or more journal lines have been entered incorrectly, resulting in total debits not equalling total credits.

**Resolution:**
1. Review each line of the journal
2. Verify that no line has both a debit and credit amount entered
3. Add up all debits and all credits independently
4. Identify the discrepancy and correct the relevant line
5. The totals panel at the bottom of the journal form shows running totals — use this as you enter lines

---

#### Error: "Period Closed — Cannot Post"

**Cause:** The posting date on the transaction falls within a closed accounting period.

**Resolution:**
1. Check the posting date on the transaction
2. If the date is in error, correct it to a date within an open period
3. If the transaction genuinely belongs to the closed period, request a period reopen from the Finance Manager with written justification
4. Once the period is reopened, post the transaction, then close the period again

---

#### Error: "Approval Required"

**Cause:** The transaction amount exceeds the threshold requiring approval, or the workflow configuration requires all transactions to be approved.

**Resolution:**
1. This is not an error — it is the system working as designed
2. Submit the transaction for approval: click **Submit for Approval**
3. The designated approver will be notified automatically
4. Monitor the transaction status under `Finance → Approvals`

---

#### Error: "Duplicate Reference Detected"

**Cause:** A transaction with the same reference number already exists in the system (for the same supplier/customer and period).

**Resolution:**
1. Search for the existing transaction with that reference
2. Compare the two transactions — are they truly the same?
3. If it is a duplicate: cancel the new entry and process no further
4. If it is a different transaction with the same reference: update the reference number on the new entry to make it unique (e.g., append "-A")

---

#### Error: "Supplier/Customer Not Found"

**Cause:** The entity does not exist in the master data, or has been deactivated.

**Resolution:**
1. Search with partial name or code in case of a spelling variation
2. Check if the entity has been deactivated: `Finance → [Accounts Payable/Receivable] → [Suppliers/Customers] → Show Inactive`
3. If deactivated and should be active, the Finance Administrator can reactivate
4. If the entity is genuinely new, create it following the master data procedures in Chapters 7 or 8

---

#### Error: "Insufficient Funds in Payment Account"

**Cause:** The bank account selected for the payment has a balance lower than the payment amount in awo.

**Resolution:**
1. Verify the actual bank balance — awo's balance may be unreconciled
2. If funds are available in the actual bank but awo shows insufficient balance, check for unreconciled receipts that should update the awo balance
3. If funds are genuinely insufficient, defer the payment or arrange a transfer from another account (see Chapter 9.6)
4. Never override this warning without Finance Manager approval

---

#### Error: "Exchange Rate Not Found"

**Cause:** No exchange rate has been configured for the transaction currency and date.

**Resolution:**
1. Navigate to `Finance → Settings → Currencies → Exchange Rates`
2. Add the rate for the relevant currency and date
3. Return to the transaction and proceed

---

### 17.2 When to Escalate vs. Self-Resolve

| Scenario | Self-Resolve | Escalate |
|---|---|---|
| Journal imbalance | ✅ | Only if cause is unclear |
| Closed period posting | ❌ | Always — Finance Manager |
| Duplicate invoice detected | ✅ — cancel duplicate | If in doubt about which is correct |
| Missing exchange rate | ✅ — add rate (if authorised) | If not authorised to add rates |
| Suspected fraud | ❌ | Always — Compliance Officer immediately |
| System error / unexpected behaviour | ❌ | Finance Support |
| Permission denied | ❌ | System Administrator |
| Reconciliation difference >1% of account balance | ❌ | Senior Accountant |

### 17.3 Logging and Reporting Issues to Support

When escalating a technical issue:

1. Note the exact error message (screenshot preferred)
2. Record the date, time, and your user ID
3. Describe the steps you took immediately before the error occurred
4. Note the transaction reference number if applicable
5. Submit a support ticket at `Help → Support → New Ticket` or email the Finance Support address configured by your organisation

---

## Chapter 18 — Escalation Matrix and Support

### 18.1 Escalation Tiers and Contacts

| Issue Type | First Contact | If Unresolved |
|---|---|---|
| Transaction posting failure | Finance Support | Finance Manager |
| Approval delays | Finance Manager | Business Unit Head |
| Missing or unallocated transactions | Accounting Team | Finance Manager |
| Permission or access problems | Finance Administrator | System Administrator |
| Bank reconciliation mismatch | Senior Accountant | Finance Manager |
| Suspected fraud or irregularity | Compliance Officer | CEO / Board Audit Committee |
| System error or downtime | IT / System Administrator | Platform Support |
| Regulatory or tax query | Tax Accountant | External Tax Advisor |
| Intercompany discrepancy | Counterpart Entity Accountant | Group Finance Manager |

### 18.2 SLA Expectations by Issue Type

| Priority | Issue Type | Expected Response | Expected Resolution |
|---|---|---|---|
| Critical | System down / fraud / data breach | 30 minutes | 4 hours |
| High | Posting blocked / approval unavailable | 2 hours | Same business day |
| Medium | Reconciliation discrepancy / access issue | 4 hours | 2 business days |
| Low | Report configuration / minor UI issue | 1 business day | 5 business days |

### 18.3 Raising a Support Ticket

1. Navigate to `Help → Support → New Ticket`
2. Select the issue category
3. Set the priority level based on business impact
4. Describe the issue with as much detail as possible (include screenshots and transaction references)
5. Submit — you will receive a ticket reference number by email
6. Monitor the ticket status in `Help → Support → My Tickets`

### 18.4 Platform vs. Tenant Support Boundaries

| Type of Issue | Supported By |
|---|---|
| Application bugs, system errors, downtime | Platform Support (awo) |
| Configuration questions (chart of accounts, workflows) | Tenant Finance Administrator |
| User access and role issues | Tenant Finance Administrator |
| Accounting guidance | Tenant's own finance team or external advisor |
| Data issues within a tenant | Tenant Finance Manager + Platform Support |
| Cross-tenant issues | Platform Support only |

---

## Chapter 19 — Daily, Monthly, and Year-End Checklists

### 19.1 Start-of-Day Checklist

| Task | Role |
|---|---|
| Log in and review Finance Dashboard | All finance users |
| Check and clear pending approvals | Approvers, Finance Manager |
| Review overnight failed postings or alerts | Accountant, Finance Manager |
| Confirm bank feed has updated (if automatic) | Accountant |
| Review overdue AR items for collections action | Accountant |
| Check upcoming AP due dates | Accountant |

### 19.2 End-of-Day Checklist

| Task | Role |
|---|---|
| Ensure all receipts collected today are posted | Cashier, Accountant |
| Submit all draft transactions for approval | All finance users |
| Confirm pending approvals are actioned | Approvers |
| Verify cash balance matches physical count (petty cash) | Cashier |
| Review and clear any alerts on the dashboard | All finance users |
| Log out of the system | All finance users |

### 19.3 Month-End Checklist

*(See Chapter 13.2 for the detailed month-end close checklist)*

| Phase | Responsible |
|---|---|
| Transactional completion (all invoices, bills, receipts, payments posted) | Accountant, Cashier |
| Bank reconciliations complete | Accountant |
| Accruals and prepayments posted | Accountant |
| Depreciation posted | Accountant |
| FX revaluation run | Accountant |
| Suspense accounts cleared | Accountant |
| Trial balance reviewed | Finance Manager |
| Period closed | Finance Manager |
| Financial statements generated and filed | Finance Manager |

### 19.4 Year-End Checklist

| Task | Responsible |
|---|---|
| All 12 month-end checklists completed | Finance Manager |
| Audit adjustments posted and approved | Finance Manager + Accountant |
| Final trial balance agreed to auditor's figures | Finance Manager |
| Retained earnings transfer posted | Finance Manager |
| Final period of financial year closed | Finance Manager |
| New financial year opened | Finance Administrator |
| Opening balances in new year verified | Accountant |
| Statutory financial statements prepared | Finance Manager + Accountant |
| Tax returns filed | Tax Accountant |
| Prior year archived | Finance Administrator |

---

## Appendix A — Glossary of Finance Terms

| Term | Definition |
|---|---|
| Accrual | An expense or revenue recognised before cash moves |
| AP | Accounts Payable — money the organisation owes to suppliers |
| AR | Accounts Receivable — money owed to the organisation by customers |
| Audit Trail | An immutable record of all system actions |
| Chart of Accounts (CoA) | The complete list of ledger accounts used in the organisation |
| Cost Centre | An organisational unit used to track spending or revenue |
| Credit | An accounting entry that increases liabilities, equity, or income |
| Debit | An accounting entry that increases assets or expenses |
| Double-Entry | The accounting principle requiring every transaction to have equal debits and credits |
| FX | Foreign Exchange — the conversion between currencies |
| GRN | Goods Receipt Note — confirmation that goods were received |
| IFRS | International Financial Reporting Standards |
| Imprest | A petty cash system where the float is replenished to a fixed amount |
| Journal Entry | A manual accounting record in the General Ledger |
| Ledger | The record of all transactions for a specific account |
| Period Close | The process of locking an accounting period against further changes |
| Prepayment | A payment made in advance for a future benefit |
| Reconciliation | The process of matching two sets of records to confirm they agree |
| Reversal | A system-generated journal that cancels a prior posting |
| SoD | Segregation of Duties — the control that prevents one person from controlling an entire process |
| Sub-ledger | A detailed ledger (e.g., AR, AP) that feeds into the General Ledger |
| Suspense Account | A temporary holding account for unresolved transactions |
| Trial Balance | A report listing all ledger account balances to verify that debits equal credits |
| VAT/GST | Value Added Tax / Goods and Services Tax |
| WHT | Withholding Tax — tax deducted at source by the payer |

---

## Appendix B — Keyboard Shortcuts Reference

| Shortcut | Action |
|---|---|
| `Alt + J` | New journal entry |
| `Alt + R` | New receipt |
| `Alt + P` | New payment |
| `Alt + A` | View pending approvals |
| `Alt + S` | Global search |
| `Ctrl + S` | Save current form |
| `Esc` | Cancel / close panel |
| `Ctrl + Shift + L` | Log out |
| `Alt + D` | Go to Finance Dashboard |
| `Alt + T` | Go to trial balance |

---

## Appendix C — Full Permission Matrix

*(See Chapter 2.3 for the role-level summary. This appendix provides granular feature-level permissions.)*

| Feature | Finance Admin | Accountant | Cashier | Approver | Finance Manager | Auditor |
|---|---|---|---|---|---|---|
| **Chart of Accounts** | | | | | | |
| View | ✅ | ✅ | ❌ | ❌ | ✅ | ✅ |
| Create/Edit | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Deactivate | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Journal Entries** | | | | | | |
| Create draft | ❌ | ✅ | ❌ | ❌ | ✅ | ❌ |
| Submit for approval | ❌ | ✅ | ❌ | ❌ | ✅ | ❌ |
| Approve | ❌ | ❌ | ❌ | ✅ | ✅ | ❌ |
| Reverse | ❌ | ✅ | ❌ | ❌ | ✅ | ❌ |
| **AR Invoices** | | | | | | |
| Create | ❌ | ✅ | ✅ | ❌ | ✅ | ❌ |
| Approve | ❌ | ❌ | ❌ | ✅ | ✅ | ❌ |
| Write off | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ |
| **AP Bills** | | | | | | |
| Create | ❌ | ✅ | ✅ | ❌ | ✅ | ❌ |
| Approve | ❌ | ❌ | ❌ | ✅ | ✅ | ❌ |
| **Payments** | | | | | | |
| Create | ❌ | ✅ | ✅ | ❌ | ✅ | ❌ |
| Approve | ❌ | ❌ | ❌ | ✅ | ✅ | ❌ |
| Run payment batch | ❌ | ✅ | ❌ | ❌ | ✅ | ❌ |
| **Banking** | | | | | | |
| Configure bank accounts | ✅ | ❌ | ❌ | ❌ | ✅ | ❌ |
| Import statements | ❌ | ✅ | ❌ | ❌ | ✅ | ❌ |
| Reconcile | ❌ | ✅ | ❌ | ❌ | ✅ | ❌ |
| **Period Management** | | | | | | |
| Close period | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ |
| Reopen period | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ |
| **Settings** | | | | | | |
| Manage users | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Configure tax rates | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Configure workflows | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| View audit log | ✅ | ❌ | ❌ | ❌ | ✅ | ✅ |

---

## Appendix D — Document and Evidence Standards

All supporting documents attached to transactions must meet the following standards:

| Standard | Requirement |
|---|---|
| Format | PDF preferred; PNG, JPG, XLSX, DOCX accepted |
| Legibility | Fully readable — no blurred, cropped, or dark images |
| Authenticity | Original documents — not handwritten copies or retyped versions |
| Completeness | Document must show: date, amount, parties involved, description of transaction |
| File size | Maximum 10 MB per file |
| Naming convention | `[DocumentType]_[Reference]_[Date]` e.g., `Invoice_SUP-2024-0012_2024-03-15.pdf` |
| Timing | Must be attached at time of transaction creation |
| Retention | System retains attached documents for the configured retention period |

---

## Appendix E — Accounting Standards Reference

The awo Finance Module supports compliance with the following standards. The organisation's finance team is responsible for applying the correct treatment.

| Standard | Area | Key Reference |
|---|---|---|
| IFRS 15 | Revenue recognition | Recognise revenue when performance obligations are satisfied |
| IFRS 9 | Financial instruments | Classification and measurement of financial assets and liabilities |
| IFRS 16 | Leases | Recognise right-of-use assets and lease liabilities |
| IAS 21 | Foreign currency | Translate foreign currency transactions at spot rate; revalue at closing rate |
| IAS 36 | Impairment | Test assets for impairment at least annually |
| IAS 37 | Provisions | Recognise provisions when a present obligation exists with probable outflow |
| GAAP ASC 606 | Revenue (US GAAP) | Five-step revenue recognition model |
| GAAP ASC 842 | Leases (US GAAP) | Lessee recognises right-of-use asset and lease liability |

This reference is informational. Always consult a qualified accountant for the correct application of accounting standards to specific transactions.

---

## Appendix F — Release Notes and Known Limitations

This section is updated with each awo Finance Module release. Refer to the in-application release notes at `Help → What's New` for the most current version details.

**Known Limitations (as of this runbook version):**

- Bank statement auto-import is not available for all banking institutions. Contact Platform Support for the list of supported banks.
- Recurring journals do not support variable-amount recurrence. For variable amounts, create a recurring template with a zero amount and update the amount manually before each submission.
- The custom report builder does not currently support pivot-table style output. Use the XLSX export for advanced pivot analysis in Excel.
- Multi-level intercompany consolidation (more than two entity levels) requires Platform Administrator involvement. Contact Platform Support for group consolidation requirements.

---

*End of Document*

---

**Document Classification:** Internal — Finance Operations  
**Review Cycle:** Annual or upon significant module update  
**Document Owner:** Finance Administrator  
**Approved By:** Finance Manager
