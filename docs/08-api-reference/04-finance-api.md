---
title: Finance API
portal: 8 — API Reference
section: 08-api-reference
audience: [backend-engineer, frontend-engineer, integrator]
related:
  - "[API Reference Overview](01-api-reference-overview.md)"
  - "[Finance Module Overview](../02-product-overview/02-module-overview.md)"
---

# Finance API

## Chart of Accounts

### GET /api/v1/finance/accounts

List accounts for the authenticated tenant.

#### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `account_type` | string | Filter: `asset`, `liability`, `equity`, `revenue`, `expense` |
| `parent_id` | uuid | Filter by parent account |
| `active` | bool | Default: true |
| `page` | int | Page number |
| `page_size` | int | Items per page |

#### Response 200

```json
{
  "data": [
    {
      "id": "...",
      "account_code": "1000",
      "account_name": "Cash and Cash Equivalents",
      "account_type": "asset",
      "parent_id": null,
      "balance": "45230.000000",
      "currency": "USD",
      "is_active": true,
      "created_at": "2025-01-01T00:00:00Z"
    }
  ],
  "pagination": { "page": 1, "page_size": 20, "total": 45, "total_pages": 3 }
}
```

---

### POST /api/v1/finance/accounts

Create a new account.

```json
{
  "account_code": "1010",
  "account_name": "Petty Cash",
  "account_type": "asset",
  "parent_id": "...",
  "currency": "USD"
}
```

**Response 201** — returns created account.

**Response 409** — duplicate `account_code` for tenant.

---

### GET /api/v1/finance/accounts/:id

Get single account with current balance.

---

### PUT /api/v1/finance/accounts/:id

Update account. Include `version` for optimistic lock.

---

## Transactions

### GET /api/v1/finance/transactions

List transactions.

#### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `account_id` | uuid | Filter by account |
| `date_from` | date | Start date (YYYY-MM-DD) |
| `date_to` | date | End date |
| `status` | string | `draft`, `posted`, `void` |
| `page` | int | — |
| `page_size` | int | — |

#### Response 200

```json
{
  "data": [
    {
      "id": "...",
      "transaction_number": "TXN-2025-00001",
      "description": "Contract payment received",
      "transaction_date": "2025-01-15",
      "status": "posted",
      "currency": "USD",
      "total_amount": "5000.000000",
      "reference": "CONT-2025-0001",
      "entries": [
        {
          "id": "...",
          "account_id": "...",
          "account_code": "1000",
          "account_name": "Cash",
          "debit_amount": "5000.000000",
          "credit_amount": "0.000000"
        },
        {
          "id": "...",
          "account_id": "...",
          "account_code": "4000",
          "account_name": "Revenue",
          "debit_amount": "0.000000",
          "credit_amount": "5000.000000"
        }
      ]
    }
  ],
  "pagination": { "page": 1, "page_size": 20, "total": 120, "total_pages": 6 }
}
```

---

### POST /api/v1/finance/transactions

Create a journal entry (status: draft).

```json
{
  "description": "Monthly rent payment",
  "transaction_date": "2025-01-31",
  "currency": "USD",
  "reference": "INV-2025-001",
  "entries": [
    { "account_id": "...", "debit_amount": "3000", "credit_amount": "0" },
    { "account_id": "...", "debit_amount": "0",    "credit_amount": "3000" }
  ]
}
```

**Response 422** — if debits ≠ credits (unbalanced journal entry).

---

### POST /api/v1/finance/transactions/:id/post

Post a draft transaction. Moves status to `posted` and updates account balances.

```json
{ "version": 1 }
```

**Permission required**: `finance.transaction.post`

---

### POST /api/v1/finance/transactions/:id/void

Void a posted transaction. Creates reversing entries.

```json
{ "version": 2, "reason": "Duplicate entry" }
```

**Response 422** — transaction already voided, or void not allowed for this type.

---

## Account Groups

### GET /api/v1/finance/account-groups

List account groups (used for financial statement grouping).

### POST /api/v1/finance/account-groups

```json
{
  "name": "Current Assets",
  "account_type": "asset",
  "display_order": 1
}
```

---

## Reports

### GET /api/v1/finance/reports/trial-balance

Trial balance at a point in time.

#### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `as_of_date` | date | Balance date (default: today) |

#### Response 200

```json
{
  "as_of_date": "2025-01-31",
  "accounts": [
    {
      "account_code": "1000",
      "account_name": "Cash",
      "account_type": "asset",
      "debit_balance": "45230.000000",
      "credit_balance": "0.000000"
    }
  ],
  "totals": {
    "total_debits": "250000.000000",
    "total_credits": "250000.000000",
    "balanced": true
  }
}
```

---

### GET /api/v1/finance/reports/balance-sheet

Balance sheet as of date. Returns assets, liabilities, equity sections.

#### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `as_of_date` | date | Required |

---

### GET /api/v1/finance/reports/income-statement

P&L for a period.

#### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `date_from` | date | Required |
| `date_to` | date | Required |
