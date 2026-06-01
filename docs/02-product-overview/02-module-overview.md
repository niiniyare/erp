---
title: Module Overview
portal: 2 — Product Overview
section: 02-product-overview
audience: [all]
related:
  - "[Product Overview](01-product-overview.md)"
  - "[Module Map](../03-platform-architecture/00-overview/03-module-map.md)"
---

# Module Overview

## Contracts Module

Manages the full lifecycle of organisational contracts with vendors, clients, and service providers.

**Key features:**
- Draft → Submit → Review → Approve → Activate lifecycle
- Optimistic locking for concurrent edits
- Contract line items with auto-calculated totals
- Bulk import via CSV
- Expiry notifications and renewal reminders
- Multi-approver workflows via Temporal

**Data owned:** `contracts`, `contract_lines`

## Finance Module

Double-entry bookkeeping with chart of accounts, journal entries, and financial statements.

**Key features:**
- Chart of accounts (assets, liabilities, equity, income, expense)
- Journal entry creation with balanced debit/credit validation
- Account balance tracking
- Financial statement generation (balance sheet, P&L)
- Integration with contracts (liability creation on activation)

**Data owned:** `finance_accounts`, `finance_transactions`, `finance_transaction_entries`, `finance_account_balances`

## HR Module

Employee records, organisational assignments, and workforce lifecycle management.

**Key features:**
- Employee profile management
- Org hierarchy assignments
- Onboarding/offboarding workflows
- Integration with contracts (employee-linked service contracts)

**Data owned:** `employees`, `persons`

## Procurement Module

Purchase request to purchase order workflow.

**Key features:**
- Purchase requisition creation and approval
- Purchase order generation
- Supplier management
- Budget validation against finance module

## Inventory Module

Stock management and warehouse operations.

**Key features:**
- Multi-warehouse inventory
- Stock adjustments with reason codes
- Low-stock alerts
- Integration with procurement (PO receipt)

## Entity/Organisation Module

The foundational hierarchy on which all other modules operate.

**Key features:**
- Unlimited depth entity hierarchy (company → division → department → cost centre)
- Closure table for O(1) subtree queries
- Entity type management (custom org unit types)
- Dynamic attributes per entity type

**Data owned:** `entities`, `hierarchy_paths`, `entity_types`, `attribute_definitions`, `attribute_values`
