> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Finance API
portal: 8 — API Reference
section: 08-api-reference
audience: [backend-engineer, frontend-engineer, integrator]
related:
  - "[API Reference Overview](01-api-reference-overview.md)"
  - "[Authentication](02-authentication.md)"
  - "[Tenants API](04-tenants-api.md)"
---

# Finance API

Manage the chart of accounts, post journal entries, and query transaction history.

## Chart of Accounts

### List Accounts

```
GET /api/v1/tenants/{tenant_id}/finance/accounts
```

**Query parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `type` | string | `asset`, `liability`, `equity`, `revenue`, `expense` |
| `group_id` | uuid | Filter by account group |
| `active` | bool | `true` to exclude archived accounts (default: true) |
| `search` | string | Search by code or name |
| `page` | int | Page number (default: 1) |
| `page_size` | int | Results per page (default: 50, max: 200) |

**Response 200:**

```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "code": "1100",
      "name": "Accounts Receivable",
      "type": "asset",
      "normal_balance": "debit",
      "group_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
      "group_name": "Current Assets",
      "currency": "USD",
      "is_active": true,
      "balance": "125000.000000",
      "created_at": "2025-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 50,
    "total": 87,
    "total_pages": 2
  }
}
```

### Get Account

```
GET /api/v1/tenants/{tenant_id}/finance/accounts/{account_id}
```

Returns same shape as list item, plus `recent_transactions` (last 5).

### Create Account

```
POST /api/v1/tenants/{tenant_id}/finance/accounts
```

**Request body:**

```json
{
  "code": "6200",
  "name": "Travel Expenses",
  "type": "expense",
  "group_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "currency": "USD",
  "description": "Employee travel and accommodation costs"
}
```

Account `type` determines `normal_balance`: `asset`/`expense` → `debit`; `liability`/`equity`/`revenue` → `credit`.

**Response 201:** Account object.

### Update Account

```
PATCH /api/v1/tenants/{tenant_id}/finance/accounts/{account_id}
```

**Request body** (all optional):

```json
{
  "name": "Travel & Entertainment Expenses",
  "description": "Employee travel, accommodation, and entertainment",
  "group_id": "new-group-uuid"
}
```

Cannot change `code`, `type`, or `currency` after an account has transactions.

### Archive Account

```
POST /api/v1/tenants/{tenant_id}/finance/accounts/{account_id}/archive
```

Sets `is_active: false`. Account becomes read-only — no new entries can be posted to it. Fails with `409` if account has a non-zero balance.

---

## Account Groups

### List Groups

```
GET /api/v1/tenants/{tenant_id}/finance/account-groups
```

**Response 200:**

```json
{
  "data": [
    {
      "id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
      "name": "Current Assets",
      "parent_group_id": null,
      "display_order": 1,
      "accounts_count": 5
    },
    {
      "id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
      "name": "Cash and Equivalents",
      "parent_group_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
      "display_order": 1,
      "accounts_count": 2
    }
  ]
}
```

### Create Group

```
POST /api/v1/tenants/{tenant_id}/finance/account-groups
```

```json
{
  "name": "Operating Expenses",
  "parent_group_id": null,
  "display_order": 5
}
```

---

## Transactions

### List Transactions

```
GET /api/v1/tenants/{tenant_id}/finance/transactions
```

**Query parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `account_id` | uuid | Filter by account |
| `status` | string | `draft`, `posted`, `voided` |
| `date_from` | date | Inclusive start date (`YYYY-MM-DD`) |
| `date_to` | date | Inclusive end date |
| `reference` | string | Filter by reference number |
| `page` | int | Default: 1 |
| `page_size` | int | Default: 20, max: 100 |

**Response 200:**

```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "transaction_number": "JE-2025-0001",
      "reference": "INV-2025-1234",
      "date": "2025-05-15",
      "description": "Invoice payment received",
      "status": "posted",
      "currency": "USD",
      "total_debit": "50000.000000",
      "total_credit": "50000.000000",
      "posted_by": "jane.smith@example.com",
      "posted_at": "2025-05-15T14:30:00Z",
      "created_at": "2025-05-15T14:25:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 312,
    "total_pages": 16
  }
}
```

### Get Transaction

```
GET /api/v1/tenants/{tenant_id}/finance/transactions/{transaction_id}
```

**Response 200:**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "transaction_number": "JE-2025-0001",
  "reference": "INV-2025-1234",
  "date": "2025-05-15",
  "description": "Invoice payment received",
  "status": "posted",
  "currency": "USD",
  "entries": [
    {
      "id": "a1b2c3d4-...",
      "account_id": "acct-cash-uuid",
      "account_code": "1000",
      "account_name": "Cash",
      "side": "debit",
      "amount": "50000.000000"
    },
    {
      "id": "e5f6a7b8-...",
      "account_id": "acct-ar-uuid",
      "account_code": "1100",
      "account_name": "Accounts Receivable",
      "side": "credit",
      "amount": "50000.000000"
    }
  ],
  "total_debit": "50000.000000",
  "total_credit": "50000.000000",
  "posted_by": "jane.smith@example.com",
  "posted_at": "2025-05-15T14:30:00Z",
  "created_at": "2025-05-15T14:25:00Z"
}
```

### Create Transaction (Draft)

```
POST /api/v1/tenants/{tenant_id}/finance/transactions
```

Creates a draft transaction. Debits and credits need not balance at creation — balance is enforced on post.

**Request body:**

```json
{
  "date": "2025-05-15",
  "reference": "INV-2025-1234",
  "description": "Invoice payment received",
  "currency": "USD",
  "entries": [
    {
      "account_id": "acct-cash-uuid",
      "side": "debit",
      "amount": "50000.000000"
    },
    {
      "account_id": "acct-ar-uuid",
      "side": "credit",
      "amount": "50000.000000"
    }
  ]
}
```

**Constraints:**
- Minimum 2 entries
- All entries must use the same currency as the transaction header
- Account must be active (`is_active: true`)
- Cannot post to archived accounts

**Response 201:** Full transaction object with `status: "draft"`.

### Post Transaction

```
POST /api/v1/tenants/{tenant_id}/finance/transactions/{transaction_id}/post
```

Validates that `total_debit == total_credit`, then sets status to `posted` and updates account balances. Atomic operation.

**Response 200:**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "posted",
  "posted_at": "2025-05-15T14:30:00Z",
  "posted_by": "jane.smith@example.com"
}
```

**Errors:**

| Code | Error | Cause |
|------|-------|-------|
| 422 | `UNBALANCED_TRANSACTION` | `total_debit != total_credit` |
| 409 | `ALREADY_POSTED` | Transaction already in `posted` status |
| 422 | `INACTIVE_ACCOUNT` | Entry references an archived account |

### Void Transaction

```
POST /api/v1/tenants/{tenant_id}/finance/transactions/{transaction_id}/void
```

Creates reversing entries and sets the transaction status to `voided`. Account balances are updated atomically. Only `posted` transactions can be voided.

**Request body:**

```json
{
  "reason": "Duplicate entry"
}
```

**Response 200:**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "voided",
  "void_reason": "Duplicate entry",
  "voided_at": "2025-05-16T09:00:00Z",
  "voided_by": "finance.manager@example.com",
  "reversal_transaction_id": "new-txn-uuid"
}
```

---

## Account Balances

### Get Balance

```
GET /api/v1/tenants/{tenant_id}/finance/accounts/{account_id}/balance
```

**Query parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `as_of` | date | Balance as of date (default: today) |
| `include_draft` | bool | Include unposted drafts (default: false) |

**Response 200:**

```json
{
  "account_id": "acct-uuid",
  "account_code": "1100",
  "account_name": "Accounts Receivable",
  "currency": "USD",
  "balance": "125000.000000",
  "normal_balance": "debit",
  "as_of": "2025-05-31",
  "debit_total": "200000.000000",
  "credit_total": "75000.000000"
}
```

### Trial Balance

```
GET /api/v1/tenants/{tenant_id}/finance/trial-balance
```

**Query parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `as_of` | date | Balance date (default: today) |
| `include_zero_balances` | bool | Include accounts with zero balance (default: false) |

**Response 200:**

```json
{
  "as_of": "2025-05-31",
  "currency": "USD",
  "accounts": [
    {
      "code": "1000",
      "name": "Cash",
      "type": "asset",
      "debit_total": "500000.000000",
      "credit_total": "200000.000000",
      "balance": "300000.000000"
    }
  ],
  "total_debits": "1500000.000000",
  "total_credits": "1500000.000000",
  "is_balanced": true
}
```

---

## Error Responses

| Code | Error | Cause |
|------|-------|-------|
| 404 | `ACCOUNT_NOT_FOUND` | Account ID not found in tenant |
| 404 | `TRANSACTION_NOT_FOUND` | Transaction ID not found |
| 404 | `GROUP_NOT_FOUND` | Account group ID not found |
| 409 | `ACCOUNT_CODE_EXISTS` | Duplicate account code in tenant |
| 409 | `ACCOUNT_HAS_BALANCE` | Archive attempted on account with non-zero balance |
| 409 | `ACCOUNT_HAS_TRANSACTIONS` | Cannot change code/type/currency after transactions exist |
| 422 | `UNBALANCED_TRANSACTION` | Post failed: debits ≠ credits |
| 422 | `CURRENCY_MISMATCH` | Entry currency ≠ transaction currency |
| 422 | `INACTIVE_ACCOUNT` | Entry references archived account |
| 403 | `CANNOT_VOID_DRAFT` | Void called on non-posted transaction |
