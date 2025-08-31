# AWO ERP Financial Module - API Reference Guide

**Version**: 2.0  
**Date**: January 2025  
**Status**: Technical Documentation  

---

## **🏛️ Modular Architecture Overview**

The AWO ERP Financial system is organized into focused, domain-specific modules:

- **🏦 FINANCE** - Core accounting engine (accounts, transactions, banking, reporting)
- **🛒 BUY** - Procurement and accounts payable operations  
- **💰 SELL** - Revenue and accounts receivable operations
- **📊 BUDGET** - Planning, forecasting, and variance analysis
- **📈 PROJECT** - Project accounting and cost management
- **💸 TAX** - Multi-jurisdiction tax calculation and compliance
- **🌍 CURRENCY** - Exchange rates and multi-currency operations

---

## Table of Contents

1.  [API Overview](#api-overview)
2.  [Authentication & Authorization](#authentication-authorization)
3.  [Core Finance Module APIs](#core-finance-module-apis)
    *   [Account Management](#account-management-apis)
    *   [Transaction Processing](#transaction-processing-apis)
    *   [Banking & Cash Management](#banking-cash-management-apis)
    *   [Financial Reporting](#financial-reporting-apis)
4.  [Buy Module APIs (Procurement & AP)](#buy-module-apis)
5.  [Sell Module APIs (Sales & AR)](#sell-module-apis)
6.  [Ancillary Module APIs](#ancillary-module-apis)
    *   [Budget Module](#budget-module)
    *   [Currency Module](#currency-management-apis)
7.  [System & Quality](#system-quality)
    *   [Error Handling](#error-handling)
    *   [Rate Limiting & Quotas](#rate-limiting-quotas)
    *   [Testing & Quality Assurance](#testing-quality-assurance)
8.  [SDK & Integration Examples](#sdk-integration-examples)

---

## API Overview

### **Modular Base URL Structure**
All API endpoints are prefixed with a module path.

```
# Core Finance Module
Production:  https://api.awo-erp.com/v1/finance
Staging:     https://staging-api.awo-erp.com/v1/finance

# Buy Module (Procurement & AP)
Production:  https://api.awo-erp.com/v1/buy
Staging:     https://staging-api.awo-erp.com/v1/buy

# Sell Module (Sales & AR)
Production:  https://api.awo-erp.com/v1/sell
Staging:     https://staging-api.awo-erp.com/v1/sell

# Ancillary Modules
Production:  https://api.awo-erp.com/v1/{budget|tax|project|currency}
```

### **Supported Protocols**
- **REST API**: Primary interface with JSON payloads
- **gRPC**: High-performance service-to-service communication
- **GraphQL**: Flexible query interface for complex data requirements

### **Common Headers**
```
Content-Type: application/json
Accept: application/json
Authorization: Bearer <jwt-token>
X-Tenant-ID: <tenant-uuid>
X-Request-ID: <unique-request-id>
```

---

## Authentication & Authorization

### **JWT Token Structure**
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

### **Attribute-Based Access Control (ABAC)**
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

## Core Finance Module APIs

Base Path: `/finance`

### **Account Management APIs**

#### **Create Account**
```http
POST /api/v1/finance/accounts
```

**Request Body:**
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

**Response:** `201 Created` with `AccountResult`

#### **Get Account**
```http
GET /api/v1/finance/{id}
```

**Response:** `AccountResult` with full account details including current balance

#### **Get Account by Code**
```http
GET /api/v1/finance/accounts/by-code/{account_code}
```

**Example:** `GET /api/v1/finance/accounts/by-code/1100`
**Response:** `AccountResult` for account with code "1100"

#### **Get Account by Name**
```http
GET /api/v1/finance/accounts/by-name?account_name=Cash - Operating Account
```

**Response:** `AccountResult` for account with exact name match

#### **List Accounts**
```http
GET /api/v1/finance/accounts
```
**Query Parameters:** `root_type`, `account_type`, `is_active`, `parent_id`, `search`, `limit`, `offset`

**Response:** `AccountListResult` with pagination

#### **Update Account**
```http
PUT /api/v1/finance/{id}
```

**Request Body:**
```json
{
  "id": "account-uuid",
  "account_name": "Updated Account Name",
  "account_description": "Updated description",
  "is_active": true,
  "allow_manual_entries": true,
  "require_reference": false
}
```

#### **Delete Account**
```http
DELETE /api/v1/finance/{id}
```

**Response:** `204 No Content`

#### **Get Account Hierarchy**
```http
GET /api/v1/finance/accounts/hierarchy?root_id=uuid-optional
```

**Response:** Hierarchical account tree structure

#### **Get Account Balance**
```http
GET /api/v1/finance/accounts/{account_id}/balance?as_of_date=2025-08-31
```

**Response:** Current balance, total debits/credits, last transaction date

### **Transaction Processing APIs**

#### **Create Transaction**
Creates a new financial transaction in `DRAFT` state with double-entry validation.
```http
POST /api/v1/finance/transactions
```

**Request Body:**
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

**Response:** `201 Created` with `TransactionResult`

#### **Get Transaction**
```http
GET /api/v1/finance/transactions/{id}
```

**Response:** `TransactionWithEntriesResult` including all entries and validation status

#### **Get Transaction by Number**
```http
GET /api/v1/finance/transactions/by-number/{transaction_number}
```

**Example:** `GET /api/v1/finance/transactions/by-number/TXN-2025-001`
**Response:** `TransactionWithEntriesResult` for transaction with number "TXN-2025-001"

#### **List Transactions**
```http
GET /api/v1/finance/transactions
```
**Query Parameters:** `status`, `type`, `date_from`, `date_to`, `account_id`, `search`, `limit`, `offset`

#### **Post Transaction**
Makes a transaction permanent and updates account balances.
```http
POST /api/v1/finance/transactions/{id}/post
```

**Request Body:**
```json
{
  "id": "transaction-uuid",
  "posting_date": "2025-08-31",
  "validate_before_posting": true,
  "force_post": false
}
```

**Note:** Subject to ABAC segregation of duties rules.

#### **Reverse Transaction**
Creates reversal transaction with audit trail.
```http
POST /api/v1/finance/transactions/{id}/reverse
```

**Request Body:**
```json
{
  "id": "transaction-uuid",
  "reason": "Incorrect entry - duplicate payment",
  "reversal_date": "2025-08-31"
}
```

#### **Approve Transaction**
Approves transaction for posting (workflow requirement).
```http
POST /api/v1/finance/transactions/{id}/approve
```

**Request Body:**
```json
{
  "id": "transaction-uuid",
  "notes": "Reviewed and approved for posting"
}
```

#### **Validate Transaction**
Pre-validates transaction without saving.
```http
POST /api/v1/finance/transactions/validate
```

**Request Body:** Same as create transaction
**Response:** `ValidationResult` with detailed validation status

### **Banking & Cash Management APIs**

#### **Create Bank Account**
```http
POST /finance/bank-accounts
```

#### **Import Bank Statement**
```http
POST /finance/bank-accounts/{bank-account-id}/statements
```

#### **Bank Reconciliation**
```http
POST /finance/bank-accounts/{bank-account-id}/reconciliation
```

### **Financial Reporting APIs**

#### **Trial Balance**
```http
GET /api/v1/finance/reports/trial-balance?as_of_date=2025-08-31&include_zero_balances=false
```

**Response:**
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

#### **Balance Sheet** (Future Implementation)
```http
GET /api/v1/finance/reports/balance-sheet
```

#### **Income Statement** (Future Implementation)
```http
GET /api/v1/finance/reports/income-statement
```

---

## Buy Module APIs

Base Path: `/buy`

### **Vendor Management**

#### **Create Vendor**
```http
POST /buy/vendors
```

#### **Create Purchase Order**
```http
POST /buy/purchase-orders
```

### **Accounts Payable**

#### **Process Purchase Invoice**
```http
POST /buy/purchase-invoices
```

#### **Create Vendor Payment**
```http
POST /buy/vendor-payments
```

---

## Sell Module APIs

Base Path: `/sell`

### **Customer Management**

#### **Create Customer**
```http  
POST /sell/customers
```

#### **Create Sales Invoice** 
```http
POST /sell/sales-invoices
```

### **Accounts Receivable**

#### **Process Customer Payment**
```http
POST /sell/customer-payments
```

#### **Customer Aging Report**
```http
GET /sell/aging-report
```

### **Dynamic Pricing & Loyalty**

#### **Calculate Dynamic Price**
```http
POST /sell/pricing/calculate
```

---

## Ancillary Module APIs

### **Budget Module**

Base Path: `/budget`

#### **Create Budget**
```http
POST /budget/budgets
```

#### **Budget vs Actual Report**
```http
GET /budget/variance-report
```

### **Currency Management APIs**

Base Path: `/currency`

#### **Get Exchange Rates**
```http
GET /currency/exchange-rates
```

#### **Create Currency Conversion**
```http
POST /currency/convert
```

---

## System & Quality

### **Error Handling**

#### **Standard Error Response**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": [
      {
        "field": "amount",
        "message": "Amount must be greater than zero",
        "code": "AMOUNT_INVALID"
      }
    ],
    "request_id": "req-uuid",
    "timestamp": "2025-01-20T10:35:00Z"
  }
}
```

#### **HTTP Status Codes & Error Categories**
- **400/422 (Validation/Semantic Errors):** `VALIDATION_ERROR`, `TRANSACTION_UNBALANCED`, `ACCOUNT_INACTIVE`
- **401/403 (Auth Errors):** `UNAUTHORIZED`, `PERMISSION_DENIED`, `TENANT_MISMATCH`
- **409 (Conflict):** `DUPLICATE_ENTRY`
- **429 (Rate Limit):** `RATE_LIMIT_EXCEEDED`
- **5xx (Server Errors):** `DATABASE_ERROR`, `INTERNAL_SERVER_ERROR`

### **Rate Limiting & Quotas**
The API employs rate limiting and quota management to ensure stability. Standard rate limit headers (`X-RateLimit-Limit`, `X-RateLimit-Remaining`, etc.) are returned with every request.

### **Testing & Quality Assurance**
Our testing strategy is built on the principle that **financial accuracy is non-negotiable**.

- **Correctness**: Every API endpoint is covered by extensive unit and integration tests to ensure financial calculations are mathematically accurate and business rules are enforced. This includes validation of double-entry bookkeeping for all transactions.
- **Security**: All endpoints undergo rigorous security testing, including validation of ABAC policies (e.g., segregation of duties) and audit trail generation for sensitive operations.
- **Performance**: The system is load-tested to meet strict Service Level Agreements (SLAs). For example, the transaction creation endpoint is benchmarked to sustain over **100 transactions per second** with an average latency below **50ms**.
- **Coverage**: We maintain a code coverage target of over 90% for core domain logic and 80%+ for integration points to guarantee reliability.

---

## SDK & Integration Examples

### **Go SDK Example**
```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/niiniyare/financial-client-go"
)

func main() {
    client := financial.NewClient(&financial.Config{
        BaseURL:   "https://api.awo-erp.com/v1/finance",
        Token:     "your-jwt-token",
        TenantID:  "tenant-uuid",
    })
    
    account, err := client.Accounts.Create(context.Background(), &financial.CreateAccountRequest{
        Code:        "1200",
        Name:        "Accounts Receivable",
        AccountType: "receivable",
        RootType:    "asset",
        Currency:    "USD",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Created account: %s\n", account.ID)
}
```

### **cURL Examples**
```bash
# Create Account
curl -X POST https://api.awo-erp.com/v1/finance/accounts \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "1300",
    "name": "Inventory",
    "account_type": "inventory",
    "root_type": "asset",
    "currency": "USD"
  }'

# Create Transaction
curl -X POST https://api.awo-erp.com/v1/finance/transactions \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "transaction_type": "manual",
    "date": "2025-01-20",
    "description": "Test transaction",
    "entries": [
      {
        "account_id": "acc-debit-uuid",
        "debit_amount": "100.00",
        "credit_amount": "0.00"
      },
      {
        "account_id": "acc-credit-uuid", 
        "debit_amount": "0.00",
        "credit_amount": "100.00"
      }
    ]
  }'
```

---

**Document Control**
- **Version**: 2.0
- **Last Updated**: January 2025  
- **Next Review**: Monthly during development
- **Approval Required**: API Lead, Security Architect

**Related Documents**
- Financial Implementation Plan (@docs/module/financial/financial-implementation-plan.md)
- Security & Compliance Guide (@docs/module/financial/financial-security-compliance-guide.md)
- Architecture Guide (@docs/module/financial/financial-architecture-guide.md)
