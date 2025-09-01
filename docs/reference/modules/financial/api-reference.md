# Financial Module - API Reference

**Version**: 2.0  
**Date**: August 31, 2025  
**Status**: Production  
**OpenAPI Version**: 3.0.3

---

## Overview

### API Description
The Financial Module API provides comprehensive double-entry bookkeeping, transaction processing, and financial reporting capabilities within the AWO ERP system. This production-ready API supports multi-currency operations, advanced validation frameworks, and enterprise-grade compliance requirements.

### Base Information
- **Base URL**: `https://api.awo-erp.com/api/v1/finance`
- **Authentication**: Bearer Token (JWT)
- **Content Type**: `application/json`
- **API Version**: `v1`

### Quick Links
- [Interactive API Explorer](../../../reference/api/swagger-ui.md) - Test endpoints directly
- [Authentication Guide](../../../reference/api/auth/index.md) - Get started with API authentication
- [SDK Examples](examples/) - Code examples in multiple languages

---

## Table of Contents

1. [Overview](#overview)
2. [Authentication](#authentication)
3. [Core Endpoints](#core-endpoints)
   - [Account Management](#account-management)
   - [Transaction Processing](#transaction-processing)
   - [Financial Reporting](#financial-reporting)
4. [Search Endpoints](#search-endpoints)
5. [Data Models](#data-models)
6. [Error Handling](#error-handling)
7. [Rate Limiting](#rate-limiting)
8. [Code Examples](#code-examples)
9. [Testing](#testing)
10. [Changelog](#changelog)

---

## Authentication

### Bearer Token Authentication
All API endpoints require authentication using JWT bearer tokens.

```bash
# Include in request headers
Authorization: Bearer <your-jwt-token>
X-Tenant-ID: <tenant-id>  # Required for multi-tenant operations
```

### Required Headers
| Header | Required | Description |
|--------|----------|-------------|
| `Authorization` | Yes | Bearer JWT token |
| `X-Tenant-ID` | Yes | Tenant context identifier |
| `Content-Type` | Yes | `application/json` for POST/PUT requests |
| `Accept` | No | `application/json` (default) |

### JWT Token Structure
The JWT token contains essential identity and permission information, including the `tenant_id` to enforce data isolation.
```json
{
  "header": { "alg": "RS256", "typ": "JWT", "kid": "key-id" },
  "payload": {
    "sub": "user-uuid",
    "tenant_id": "tenant-uuid",
    "iss": "awo-erp",
    "aud": "financial-api",
    "exp": 1640995200,
    "iat": 1640908800,
    "permissions": [
      "finance:accounts:read",
      "finance:transactions:create"
    ]
  }
}
```

### Attribute-Based Access Control (ABAC)
In addition to JWT scopes, the system uses ABAC for fine-grained authorization. Policies are evaluated based on user attributes, resource properties, and request context.

**ABAC Context Headers:**
```http
X-Tenant-ID: tenant-uuid
X-Department-ID: dept-uuid (optional)
X-Cost-Center: cost-center-code (optional)
```

**Business Rule Example (Segregation of Duties):**
A user who creates a transaction cannot approve the same transaction. This is enforced by the ABAC policy engine.

---

## Core Endpoints

### Account Management

#### Create Account
Create a new financial account within the chart of accounts.

**Endpoint**: `POST /api/v1/finance/accounts`

**Request Body**:
```json
{
  "entity_id": "uuid-optional",
  "account_code": "1100",
  "account_name": "Cash - Operating Account", 
  "account_description": "Primary operating cash account",
  "parent_account_id": "uuid-optional",
  "root_type": "ASSET",
  "account_type": "BANK",
  "account_subtype": "CHECKING",
  "normal_balance": "DEBIT",
  "currency_code": "USD",
  "is_active": true
}
```

**Response** (201 Created):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "tenantId": "123e4567-e89b-12d3-a456-426614174000",
  "code": "1100",
  "name": "Cash - Operating Account",
  "description": "Primary operating cash account",
  "accountType": "BANK",
  "rootType": "ASSET",
  "normalBalance": "DEBIT",
  "currencyCode": "USD",
  "currentBalance": "0.00",
  "isActive": true,
  "createdAt": "2025-08-31T10:30:00Z",
  "updatedAt": "2025-08-31T10:30:00Z",
  "version": 1
}
```

**Business Rules**:
- Account code must be unique within tenant
- Required fields: `account_code`, `account_name`, `account_type`, `root_type`
- Code format: Alphanumeric with hyphens, max 50 characters
- Name max length: 255 characters
- Root type must align with account type (Assets=ASSET, etc.)

**Error Responses**:
```json
// 400 Bad Request - Validation Error
{
  "error": "validation_failed",
  "message": "Request validation failed",
  "details": [
    {
      "field": "account_code",
      "code": "required",
      "message": "Account code is required"
    }
  ],
  "timestamp": "2025-08-31T10:30:00Z",
  "correlationId": "abc123-def456"
}

// 409 Conflict - Duplicate Code
{
  "error": "duplicate_account_code",
  "message": "Account code already exists in tenant",
  "timestamp": "2025-08-31T10:30:00Z",
  "correlationId": "abc123-def456"
}
```

#### Get Account by ID
Retrieve a specific account by its ID.

**Endpoint**: `GET /api/v1/finance/accounts/{id}`

**Path Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | UUID | Yes | Account identifier |

**Response** (200 OK):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "tenantId": "123e4567-e89b-12d3-a456-426614174000",
  "code": "1100",
  "name": "Cash - Operating Account",
  "description": "Primary operating cash account",
  "accountType": "BANK",
  "rootType": "ASSET",
  "normalBalance": "DEBIT",
  "currencyCode": "USD",
  "currentBalance": "5000.00",
  "isActive": true,
  "parentAccountId": null,
  "depth": 0,
  "createdAt": "2025-08-31T10:30:00Z",
  "updatedAt": "2025-08-31T10:30:00Z",
  "version": 1,
  "_links": {
    "self": {
      "href": "/api/v1/finance/accounts/550e8400-e29b-41d4-a716-446655440000"
    },
    "balance": {
      "href": "/api/v1/finance/accounts/550e8400-e29b-41d4-a716-446655440000/balance"
    },
    "transactions": {
      "href": "/api/v1/finance/transactions?account_id=550e8400-e29b-41d4-a716-446655440000"
    }
  }
}
```

**Error Responses**:
```json
// 404 Not Found
{
  "error": "account_not_found",
  "message": "Account not found or access denied",
  "timestamp": "2025-08-31T10:30:00Z",
  "correlationId": "abc123-def456"
}
```

#### List Accounts
Retrieve a paginated list of accounts with optional filtering.

**Endpoint**: `GET /api/v1/finance/accounts`

**Query Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `page` | integer | No | Page number (default: 1) |
| `limit` | integer | No | Items per page (default: 20, max: 100) |
| `root_type` | string | No | Filter by root type (`ASSET`, `LIABILITY`, `EQUITY`, `INCOME`, `EXPENSE`) |
| `account_type` | string | No | Filter by account type |
| `is_active` | boolean | No | Filter by active status |
| `parent_id` | UUID | No | Filter by parent account |
| `search` | string | No | Search in code and name fields |
| `sort` | string | No | Sort field (`code`, `name`, `createdAt`, `updatedAt`) |
| `order` | string | No | Sort order (`asc`, `desc`) |

**Response** (200 OK):
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "code": "1100",
      "name": "Cash - Operating Account",
      "accountType": "BANK",
      "rootType": "ASSET",
      "currentBalance": "5000.00",
      "isActive": true,
      "createdAt": "2025-08-31T10:30:00Z",
      "updatedAt": "2025-08-31T10:30:00Z"
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "code": "2000",
      "name": "Accounts Payable",
      "accountType": "PAYABLE",
      "rootType": "LIABILITY",
      "currentBalance": "2500.00",
      "isActive": true,
      "createdAt": "2025-08-31T11:00:00Z",
      "updatedAt": "2025-08-31T11:00:00Z"
    }
  ],
  "pagination": {
    "currentPage": 1,
    "totalPages": 5,
    "totalItems": 87,
    "itemsPerPage": 20,
    "hasNextPage": true,
    "hasPreviousPage": false
  },
  "_links": {
    "self": {
      "href": "/api/v1/finance/accounts?page=1&limit=20"
    },
    "next": {
      "href": "/api/v1/finance/accounts?page=2&limit=20"
    },
    "first": {
      "href": "/api/v1/finance/accounts?page=1&limit=20"
    },
    "last": {
      "href": "/api/v1/finance/accounts?page=5&limit=20"
    }
  }
}
```

#### Update Account
Update an existing account.

**Endpoint**: `PUT /api/v1/finance/accounts/{id}`

**Path Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | UUID | Yes | Account identifier |

**Request Body**:
```json
{
  "account_name": "Updated Account Name",
  "account_description": "Updated description",
  "is_active": true,
  "allow_manual_entries": true,
  "version": 1  // For optimistic locking
}
```

**Response** (200 OK):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "tenantId": "123e4567-e89b-12d3-a456-426614174000",
  "code": "1100",
  "name": "Updated Account Name",
  "description": "Updated description",
  "accountType": "BANK",
  "rootType": "ASSET",
  "isActive": true,
  "createdAt": "2025-08-31T10:30:00Z",
  "updatedAt": "2025-08-31T12:15:00Z",
  "version": 2
}
```

#### Delete Account
Soft delete an account (sets isActive to false).

**Endpoint**: `DELETE /api/v1/finance/accounts/{id}`

**Path Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | UUID | Yes | Account identifier |

**Response** (204 No Content)

**Error Responses**:
```json
// 400 Bad Request - Business Rule Violation
{
  "error": "cannot_delete_account",
  "message": "Account cannot be deleted due to active transactions",
  "details": {
    "dependencies": [
      {
        "type": "transactions",
        "count": 25
      }
    ]
  },
  "timestamp": "2025-08-31T10:30:00Z",
  "correlationId": "abc123-def456"
}
```

#### Get Account Balance
Retrieve account balance information for a specific date.

**Endpoint**: `GET /api/v1/finance/accounts/{account_id}/balance`

**Path Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `account_id` | UUID | Yes | Account identifier |

**Query Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `as_of_date` | date | No | Balance as of date (default: current date) |

**Response** (200 OK):
```json
{
  "accountId": "550e8400-e29b-41d4-a716-446655440000",
  "accountCode": "1100",
  "accountName": "Cash - Operating Account",
  "asOfDate": "2025-08-31",
  "currentBalance": "5000.00",
  "totalDebits": "15000.00",
  "totalCredits": "10000.00",
  "lastTransactionDate": "2025-08-30",
  "currencyCode": "USD"
}
```

### Transaction Processing

#### Create Transaction
Creates a new financial transaction in `DRAFT` state with double-entry validation.

**Endpoint**: `POST /api/v1/finance/transactions`

**Request Body**:
```json
{
  "entity_id": "uuid-optional",
  "transaction_number": "TXN-2025-001",
  "transaction_type": "JOURNAL_ENTRY",
  "transaction_date": "2025-08-31",
  "description": "Monthly rent payment",
  "reference_number": "REF-001",
  "currency_code": "USD",
  "entries": [
    {
      "account_id": "cash-account-uuid",
      "debit_amount": "0.00",
      "credit_amount": "1500.00",
      "description": "Rent payment - cash account",
      "cost_center": "OPS-001",
      "department": "Administration"
    },
    {
      "account_id": "rent-expense-uuid",
      "debit_amount": "1500.00", 
      "credit_amount": "0.00",
      "description": "Rent expense",
      "cost_center": "OPS-001",
      "department": "Administration"
    }
  ]
}
```

**Response** (201 Created):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "tenantId": "123e4567-e89b-12d3-a456-426614174000",
  "transactionNumber": "TXN-2025-001",
  "transactionType": "JOURNAL_ENTRY",
  "transactionStatus": "DRAFT",
  "transactionDate": "2025-08-31",
  "description": "Monthly rent payment",
  "referenceNumber": "REF-001",
  "totalAmount": "1500.00",
  "currencyCode": "USD",
  "isBalanced": true,
  "createdAt": "2025-08-31T10:30:00Z",
  "version": 1,
  "entries": [
    {
      "id": "entry-uuid-1",
      "accountId": "cash-account-uuid",
      "lineNumber": 1,
      "debitAmount": "0.00",
      "creditAmount": "1500.00",
      "description": "Rent payment - cash account"
    },
    {
      "id": "entry-uuid-2",
      "accountId": "rent-expense-uuid",
      "lineNumber": 2,
      "debitAmount": "1500.00",
      "creditAmount": "0.00",
      "description": "Rent expense"
    }
  ]
}
```

**Business Rules**:
- All transactions must balance (total debits = total credits)
- Required fields: `transaction_type`, `transaction_date`, `entries`
- Minimum 2 entries required for double-entry validation
- Each entry must have either debit OR credit amount (not both)
- All entries must use the same currency

#### Get Transaction by ID
Retrieve a specific transaction by its ID.

**Endpoint**: `GET /api/v1/finance/transactions/{id}`

**Path Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | UUID | Yes | Transaction identifier |

**Response** (200 OK):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "tenantId": "123e4567-e89b-12d3-a456-426614174000",
  "transactionNumber": "TXN-2025-001",
  "transactionType": "JOURNAL_ENTRY",
  "transactionStatus": "POSTED",
  "transactionDate": "2025-08-31",
  "description": "Monthly rent payment",
  "totalAmount": "1500.00",
  "currencyCode": "USD",
  "isBalanced": true,
  "postedAt": "2025-08-31T14:30:00Z",
  "postedBy": "user-uuid",
  "createdAt": "2025-08-31T10:30:00Z",
  "updatedAt": "2025-08-31T14:30:00Z",
  "version": 3,
  "entries": [
    {
      "id": "entry-uuid-1",
      "accountId": "cash-account-uuid",
      "accountCode": "1100",
      "accountName": "Cash - Operating Account",
      "lineNumber": 1,
      "debitAmount": "0.00",
      "creditAmount": "1500.00",
      "description": "Rent payment - cash account"
    },
    {
      "id": "entry-uuid-2",
      "accountId": "rent-expense-uuid",
      "accountCode": "5100",
      "accountName": "Rent Expense",
      "lineNumber": 2,
      "debitAmount": "1500.00",
      "creditAmount": "0.00",
      "description": "Rent expense"
    }
  ],
  "_links": {
    "self": {
      "href": "/api/v1/finance/transactions/550e8400-e29b-41d4-a716-446655440000"
    },
    "reverse": {
      "href": "/api/v1/finance/transactions/550e8400-e29b-41d4-a716-446655440000/reverse",
      "method": "POST"
    }
  }
}
```

#### List Transactions
Retrieve a paginated list of transactions with optional filtering.

**Endpoint**: `GET /api/v1/finance/transactions`

**Query Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `page` | integer | No | Page number (default: 1) |
| `limit` | integer | No | Items per page (default: 20, max: 100) |
| `status` | string | No | Filter by status (`DRAFT`, `SUBMITTED`, `APPROVED`, `POSTED`) |
| `type` | string | No | Filter by transaction type |
| `date_from` | date | No | Filter transactions from date |
| `date_to` | date | No | Filter transactions to date |
| `account_id` | UUID | No | Filter by specific account involvement |
| `search` | string | No | Search in number and description |
| `sort` | string | No | Sort field (`transactionDate`, `transactionNumber`, `createdAt`) |
| `order` | string | No | Sort order (`asc`, `desc`) |

**Response** (200 OK):
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "transactionNumber": "TXN-2025-001",
      "transactionType": "JOURNAL_ENTRY",
      "transactionStatus": "POSTED",
      "transactionDate": "2025-08-31",
      "description": "Monthly rent payment",
      "totalAmount": "1500.00",
      "currencyCode": "USD",
      "createdAt": "2025-08-31T10:30:00Z"
    }
  ],
  "pagination": {
    "currentPage": 1,
    "totalPages": 10,
    "totalItems": 200,
    "itemsPerPage": 20,
    "hasNextPage": true,
    "hasPreviousPage": false
  }
}
```

#### Post Transaction
Makes a transaction permanent and updates account balances.

**Endpoint**: `POST /api/v1/finance/transactions/{id}/post`

**Path Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | UUID | Yes | Transaction identifier |

**Request Body**:
```json
{
  "posting_date": "2025-08-31",
  "validate_before_posting": true,
  "force_post": false
}
```

**Response** (200 OK):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "transactionStatus": "POSTED",
  "postedAt": "2025-08-31T14:30:00Z",
  "postedBy": "user-uuid",
  "balanceImpacts": [
    {
      "accountId": "cash-account-uuid",
      "accountCode": "1100",
      "previousBalance": "6500.00",
      "newBalance": "5000.00",
      "impact": "-1500.00"
    },
    {
      "accountId": "rent-expense-uuid",
      "accountCode": "5100",
      "previousBalance": "0.00",
      "newBalance": "1500.00",
      "impact": "+1500.00"
    }
  ]
}
```

**Note:** Subject to ABAC segregation of duties rules.

#### Reverse Transaction
Creates reversal transaction with audit trail.

**Endpoint**: `POST /api/v1/finance/transactions/{id}/reverse`

**Path Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | UUID | Yes | Transaction identifier |

**Request Body**:
```json
{
  "reason": "Incorrect entry - duplicate payment",
  "reversal_date": "2025-08-31"
}
```

**Response** (201 Created):
```json
{
  "originalTransactionId": "550e8400-e29b-41d4-a716-446655440000",
  "reversalTransactionId": "550e8400-e29b-41d4-a716-446655440001",
  "reversalTransaction": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "transactionNumber": "REV-TXN-2025-001",
    "transactionType": "REVERSAL",
    "transactionStatus": "POSTED",
    "description": "Reversal: Incorrect entry - duplicate payment",
    "totalAmount": "1500.00",
    "reversedTransactionId": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

#### Approve Transaction
Approves transaction for posting (workflow requirement).

**Endpoint**: `POST /api/v1/finance/transactions/{id}/approve`

**Path Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | UUID | Yes | Transaction identifier |

**Request Body**:
```json
{
  "notes": "Reviewed and approved for posting"
}
```

**Response** (200 OK):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "transactionStatus": "APPROVED",
  "approvedAt": "2025-08-31T13:15:00Z",
  "approvedBy": "approver-user-uuid",
  "approvalNotes": "Reviewed and approved for posting"
}
```

### Financial Reporting

#### Trial Balance
Generate trial balance report showing account balances.

**Endpoint**: `GET /api/v1/finance/reports/trial-balance`

**Query Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `as_of_date` | date | No | Balance as of date (default: current date) |
| `include_zero_balances` | boolean | No | Include accounts with zero balance (default: false) |
| `account_type` | string | No | Filter by account type |

**Response** (200 OK):
```json
{
  "as_of_date": "2025-08-31",
  "accounts": [
    {
      "account_id": "uuid",
      "account_code": "1100",
      "account_name": "Cash",
      "root_type": "ASSET",
      "account_type": "BANK",
      "normal_balance": "DEBIT",
      "total_debits": "10000.00",
      "total_credits": "5000.00", 
      "net_balance": "5000.00"
    }
  ],
  "total_debits": "50000.00",
  "total_credits": "50000.00",
  "is_balanced": true,
  "generated_at": "2025-08-31T10:30:00Z"
}
```

---

## Search Endpoints

### Search by Code
Find account by exact code match.

**Endpoint**: `GET /api/v1/finance/accounts/by-code/{code}`

**Path Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `code` | string | Yes | Account code |

**Response**: Same as Get Account by ID

### Search by Transaction Number
Find transaction by exact number match.

**Endpoint**: `GET /api/v1/finance/transactions/by-number/{transaction_number}`

**Path Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `transaction_number` | string | Yes | Transaction number |

**Response**: Same as Get Transaction by ID

### Advanced Search
Perform complex search with multiple criteria.

**Endpoint**: `POST /api/v1/finance/accounts/search`

**Request Body**:
```json
{
  "criteria": {
    "status": ["active", "inactive"],
    "rootType": "ASSET",
    "createdAfter": "2025-01-01T00:00:00Z",
    "balanceGreaterThan": "1000.00"
  },
  "sort": [
    {
      "field": "currentBalance",
      "order": "desc"
    },
    {
      "field": "accountCode",
      "order": "asc"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 50
  }
}
```

---

## Data Models

### Account Model
```json
{
  "type": "object",
  "properties": {
    "id": {
      "type": "string",
      "format": "uuid",
      "description": "Unique account identifier",
      "readOnly": true
    },
    "tenantId": {
      "type": "string", 
      "format": "uuid",
      "description": "Tenant identifier",
      "readOnly": true
    },
    "code": {
      "type": "string",
      "maxLength": 50,
      "pattern": "^[A-Za-z0-9_-]+$",
      "description": "Unique account code within tenant"
    },
    "name": {
      "type": "string",
      "maxLength": 255,
      "description": "Account display name"
    },
    "description": {
      "type": "string",
      "maxLength": 1000,
      "description": "Optional account description"
    },
    "accountType": {
      "type": "string",
      "enum": ["CASH", "BANK", "RECEIVABLE", "PAYABLE", "INVENTORY", "EXPENSE", "INCOME"],
      "description": "Account type classification"
    },
    "rootType": {
      "type": "string",
      "enum": ["ASSET", "LIABILITY", "EQUITY", "INCOME", "EXPENSE"],
      "description": "Root account type for financial statements"
    },
    "currentBalance": {
      "type": "string",
      "format": "decimal",
      "description": "Current account balance",
      "readOnly": true
    },
    "currencyCode": {
      "type": "string",
      "pattern": "^[A-Z]{3}$",
      "description": "ISO 4217 currency code"
    },
    "isActive": {
      "type": "boolean",
      "description": "Active flag for soft delete"
    },
    "createdAt": {
      "type": "string",
      "format": "date-time",
      "description": "Creation timestamp",
      "readOnly": true
    },
    "version": {
      "type": "integer",
      "description": "Version for optimistic locking",
      "readOnly": true
    }
  },
  "required": ["code", "name", "accountType", "rootType"]
}
```

### Transaction Model
```json
{
  "type": "object",
  "properties": {
    "id": {
      "type": "string",
      "format": "uuid",
      "description": "Unique transaction identifier",
      "readOnly": true
    },
    "transactionNumber": {
      "type": "string",
      "maxLength": 50,
      "description": "Unique transaction number within tenant"
    },
    "transactionType": {
      "type": "string",
      "enum": ["JOURNAL_ENTRY", "PAYMENT", "RECEIPT", "REVERSAL"],
      "description": "Transaction type classification"
    },
    "transactionStatus": {
      "type": "string",
      "enum": ["DRAFT", "SUBMITTED", "APPROVED", "POSTED", "REVERSED"],
      "description": "Current transaction status",
      "readOnly": true
    },
    "transactionDate": {
      "type": "string",
      "format": "date",
      "description": "Transaction effective date"
    },
    "description": {
      "type": "string",
      "maxLength": 1000,
      "description": "Transaction description"
    },
    "totalAmount": {
      "type": "string",
      "format": "decimal",
      "description": "Total transaction amount"
    },
    "currencyCode": {
      "type": "string",
      "pattern": "^[A-Z]{3}$",
      "description": "ISO 4217 currency code"
    },
    "isBalanced": {
      "type": "boolean",
      "description": "Whether debits equal credits",
      "readOnly": true
    },
    "entries": {
      "type": "array",
      "description": "Transaction entries (debits and credits)",
      "items": {
        "$ref": "#/components/schemas/TransactionEntry"
      },
      "minItems": 2
    }
  },
  "required": ["transactionType", "transactionDate", "entries"]
}
```

### Error Response Model
```json
{
  "type": "object",
  "properties": {
    "error": {
      "type": "string",
      "description": "Error code identifier"
    },
    "message": {
      "type": "string",
      "description": "Human-readable error message"
    },
    "details": {
      "type": "array",
      "description": "Validation error details",
      "items": {
        "type": "object",
        "properties": {
          "field": {"type": "string"},
          "code": {"type": "string"},
          "message": {"type": "string"}
        }
      }
    },
    "timestamp": {
      "type": "string",
      "format": "date-time",
      "description": "Error occurrence timestamp"
    },
    "correlationId": {
      "type": "string",
      "description": "Request correlation identifier for tracing"
    }
  },
  "required": ["error", "message", "timestamp"]
}
```

---

## Error Handling

### Standard Error Codes

| HTTP Status | Error Code | Description |
|-------------|------------|-------------|
| 400 | `validation_failed` | Request validation failed |
| 400 | `transaction_unbalanced` | Transaction debits do not equal credits |
| 400 | `invalid_request` | Malformed request |
| 401 | `unauthorized` | Authentication required |
| 403 | `forbidden` | Insufficient permissions |
| 403 | `segregation_violation` | Segregation of duties violation |
| 404 | `account_not_found` | Account not found or access denied |
| 404 | `transaction_not_found` | Transaction not found or access denied |
| 409 | `duplicate_account_code` | Account code already exists |
| 409 | `duplicate_transaction_number` | Transaction number already exists |
| 409 | `version_conflict` | Optimistic locking conflict |
| 422 | `business_rule_violation` | Financial business rule violated |
| 422 | `account_inactive` | Cannot use inactive account |
| 429 | `rate_limit_exceeded` | API rate limit exceeded |
| 500 | `internal_server_error` | Unexpected server error |

### Error Response Format
All error responses follow a consistent format:
```json
{
  "error": "validation_failed",
  "message": "Request validation failed",
  "details": [
    {
      "field": "account_code",
      "code": "required",
      "message": "Account code is required"
    }
  ],
  "timestamp": "2025-08-31T10:30:00Z",
  "correlationId": "abc123-def456"
}
```

---

## Rate Limiting

### Rate Limits
- **Standard Tier**: 1,000 requests per hour
- **Premium Tier**: 5,000 requests per hour  
- **Enterprise Tier**: 10,000 requests per hour

### Rate Limit Headers
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1642248000
```

### Rate Limit Exceeded Response
```json
{
  "error": "rate_limit_exceeded",
  "message": "API rate limit exceeded",
  "details": {
    "limit": 1000,
    "windowSeconds": 3600,
    "retryAfter": 1800
  },
  "timestamp": "2025-08-31T10:30:00Z"
}
```

---

## Code Examples

### cURL Examples

#### Create Account
```bash
curl -X POST "https://api.awo-erp.com/api/v1/finance/accounts" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "X-Tenant-ID: YOUR_TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "account_code": "1300",
    "account_name": "Inventory",
    "account_type": "INVENTORY",
    "root_type": "ASSET",
    "currency_code": "USD",
    "is_active": true
  }'
```

#### Create Transaction
```bash
curl -X POST "https://api.awo-erp.com/api/v1/finance/transactions" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "X-Tenant-ID: YOUR_TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "transaction_type": "JOURNAL_ENTRY",
    "transaction_date": "2025-08-31",
    "description": "Test transaction",
    "entries": [
      {
        "account_id": "acc-debit-uuid",
        "debit_amount": "100.00",
        "credit_amount": "0.00",
        "description": "Debit entry"
      },
      {
        "account_id": "acc-credit-uuid", 
        "debit_amount": "0.00",
        "credit_amount": "100.00",
        "description": "Credit entry"
      }
    ]
  }'
```

### JavaScript/Node.js Examples

#### Using Axios
```javascript
const axios = require('axios');

const client = axios.create({
  baseURL: 'https://api.awo-erp.com/api/v1/finance',
  headers: {
    'Authorization': `Bearer ${process.env.JWT_TOKEN}`,
    'X-Tenant-ID': process.env.TENANT_ID,
    'Content-Type': 'application/json'
  }
});

// Create account
async function createAccount(accountData) {
  try {
    const response = await client.post('/accounts', accountData);
    return response.data;
  } catch (error) {
    console.error('Error creating account:', error.response.data);
    throw error;
  }
}

// Get account by ID
async function getAccount(accountId) {
  try {
    const response = await client.get(`/accounts/${accountId}`);
    return response.data;
  } catch (error) {
    if (error.response.status === 404) {
      return null; // Account not found
    }
    throw error;
  }
}

// Create transaction
async function createTransaction(transactionData) {
  try {
    const response = await client.post('/transactions', transactionData);
    return response.data;
  } catch (error) {
    console.error('Error creating transaction:', error.response.data);
    throw error;
  }
}
```

### Go Examples

#### Using HTTP Client
```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
)

type Account struct {
    ID          string `json:"id,omitempty"`
    Code        string `json:"account_code"`
    Name        string `json:"account_name"`
    Type        string `json:"account_type"`
    RootType    string `json:"root_type"`
    IsActive    bool   `json:"is_active"`
    Currency    string `json:"currency_code"`
}

func createAccount(account Account) (*Account, error) {
    jsonData, err := json.Marshal(account)
    if err != nil {
        return nil, err
    }

    req, err := http.NewRequest("POST", 
        "https://api.awo-erp.com/api/v1/finance/accounts", 
        bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }

    req.Header.Set("Authorization", "Bearer "+os.Getenv("JWT_TOKEN"))
    req.Header.Set("X-Tenant-ID", os.Getenv("TENANT_ID"))
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusCreated {
        return nil, fmt.Errorf("API error: %s", resp.Status)
    }

    var result Account
    err = json.NewDecoder(resp.Body).Decode(&result)
    return &result, err
}
```

---

## Testing

### Postman Collection
A comprehensive Postman collection is available with:
- Pre-configured environments (dev, staging, production)
- Authentication setup scripts
- Complete endpoint coverage
- Example requests and responses
- Automated tests for response validation

**Download**: [Postman Collection](postman/finance-api-collection.json)

### API Testing Checklist
- [ ] Authentication works correctly
- [ ] CRUD operations for accounts function properly
- [ ] Transaction creation and processing work correctly
- [ ] Double-entry validation enforced
- [ ] Pagination works as expected
- [ ] Search and filtering work correctly
- [ ] Error responses are properly formatted
- [ ] Rate limiting is enforced
- [ ] Multi-tenancy isolation is verified
- [ ] Performance requirements are met
- [ ] Financial business rules are validated

### Quality Assurance
Our testing strategy is built on the principle that **financial accuracy is non-negotiable**.

- **Correctness**: Every API endpoint is covered by extensive unit and integration tests to ensure financial calculations are mathematically accurate and business rules are enforced. This includes validation of double-entry bookkeeping for all transactions.
- **Security**: All endpoints undergo rigorous security testing, including validation of ABAC policies (e.g., segregation of duties) and audit trail generation for sensitive operations.
- **Performance**: The system is load-tested to meet strict Service Level Agreements (SLAs). For example, the transaction creation endpoint is benchmarked to sustain over **100 transactions per second** with an average latency below **50ms**.
- **Coverage**: We maintain a code coverage target of over 90% for core domain logic and 80%+ for integration points to guarantee reliability.

---

## Changelog

### Version 2.0.0 (2025-08-31)
- Complete API redesign with standardized structure
- Enhanced error handling and validation
- Improved authentication and authorization
- Added comprehensive data models
- Standardized response formats
- Enhanced search and filtering capabilities
- Added rate limiting implementation
- Complete double-entry bookkeeping support
- Multi-currency transaction processing
- Advanced validation framework (30+ business rules)
- ABAC integration for fine-grained access control

### Version 1.0.0 (2025-01-01)
- Initial API release
- Basic CRUD operations for accounts
- Transaction processing capabilities
- Financial reporting endpoints
- Multi-tenant support
- Authentication and authorization
- Basic error handling

---

**Document Control**  
- **Version**: 2.0
- **Last Updated**: August 31, 2025
- **API Status**: Production
- **OpenAPI Spec**: [swagger.yaml](openapi/swagger.yaml)

**Related Documents**
- [Product Requirements Document](PRD.md)
- [Testing Strategy](testing.md)
- [Architecture Guide](architecture-guide.md)
- [Module Overview](README.md)
