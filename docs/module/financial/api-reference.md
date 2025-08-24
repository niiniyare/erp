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

### **Core Finance Module**
1. [API Overview](#api-overview)
2. [Authentication & Authorization](#authentication--authorization)
3. [Account Management APIs](#account-management-apis)
4. [Transaction Processing APIs](#transaction-processing-apis)
5. [Banking & Cash Management APIs](#banking--cash-management-apis)
6. [Financial Reporting APIs](#financial-reporting-apis)
7. [Currency Management APIs](#currency-management-apis)

### **Extended Modules**
8. [Buy Module APIs](#buy-module-apis) - Vendor, Purchase, AP
9. [Sell Module APIs](#sell-module-apis) - Customer, Sales, AR
10. [Misc Module APIs](#misc-module-apis) - Budget, Tax, Project

### **System APIs**
11. [Error Handling](#error-handling)
12. [Rate Limiting & Quotas](#rate-limiting--quotas)

---

## API Overview

### **Modular Base URL Structure**
```
# Core Finance Module
Production:  https://api.awo-erp.com/v2/finance
Staging:     https://staging-api.awo-erp.com/v2/finance
Development: https://dev-api.awo-erp.com/v2/finance

# Buy Module (Procurement & AP)
Production:  https://api.awo-erp.com/v2/buy
Staging:     https://staging-api.awo-erp.com/v2/buy
Development: https://dev-api.awo-erp.com/v2/buy

# Sell Module (Sales & AR)
Production:  https://api.awo-erp.com/v2/sell
Staging:     https://staging-api.awo-erp.com/v2/sell
Development: https://dev-api.awo-erp.com/v2/sell

# Misc Modules
Production:  https://api.awo-erp.com/v2/{budget|tax|project|currency}
```

### **Supported Protocols**
- **REST API**: Primary interface with JSON payloads
- **gRPC**: High-performance service-to-service communication
- **GraphQL**: Flexible query interface for complex data requirements

### **Content Types**
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
```json
{
  "header": {
    "alg": "RS256",
    "typ": "JWT",
    "kid": "key-id"
  },
  "payload": {
    "sub": "user-uuid",
    "tenant_id": "tenant-uuid",
    "iss": "awo-erp",
    "aud": "financial-api",
    "exp": 1640995200,
    "iat": 1640908800,
    "permissions": [
      "finance:accounts:read",
      "finance:transactions:create",
      "finance:reports:read"
    ]
  }
}
```

### **ABAC Context Headers**
```http
X-Tenant-ID: tenant-uuid
X-Department-ID: dept-uuid (optional)
X-Cost-Center: cost-center-code (optional)
X-IP-Address: client-ip (auto-detected)
X-Risk-Level: low|medium|high (calculated)
```

### **Permission Scopes**
```yaml
Account Management:
  - finance:accounts:read
  - finance:accounts:create
  - finance:accounts:update
  - finance:accounts:delete

Transaction Processing:
  - finance:transactions:read
  - finance:transactions:create
  - finance:transactions:post
  - finance:transactions:reverse

Customer Management:
  - finance:customers:read
  - finance:customers:create
  - finance:customers:update
  - finance:invoices:create

Vendor Management:
  - finance:vendors:read
  - finance:vendors:create
  - finance:vendors:update
  - finance:invoices:approve

Banking Operations:
  - finance:banking:read
  - finance:banking:reconcile
  - finance:payments:create
  - finance:payments:approve

Reporting:
  - finance:reports:read
  - finance:reports:create
  - finance:reports:export
```

### **Currency Management APIs**

#### **Get Exchange Rates**
```http
GET /currency/exchange-rates
```

**Query Parameters:**
```yaml
base_currency: string (required) - Base currency code (e.g., USD)
target_currencies: string[] (optional) - Target currency codes
date: string (optional) - Date for historical rates (YYYY-MM-DD)
provider: string (optional) - Rate provider (ecb, fed, custom)
```

**Response:**
```json
{
  "data": {
    "base_currency": "USD",
    "rate_date": "2025-01-24",
    "provider": "ecb",
    "rates": [
      {
        "currency": "EUR", 
        "rate": "0.85000",
        "inverse_rate": "1.17647",
        "last_updated": "2025-01-24T10:00:00Z"
      },
      {
        "currency": "GBP",
        "rate": "0.75000", 
        "inverse_rate": "1.33333",
        "last_updated": "2025-01-24T10:00:00Z"
      }
    ]
  }
}
```

#### **Create Currency Conversion**
```http
POST /currency/convert
```

**Request Body:**
```json
{
  "amount": "1000.00",
  "from_currency": "USD",
  "to_currency": "EUR",
  "rate_date": "2025-01-24",
  "rate_type": "spot" // spot, forward, budget
}
```

**Response:**
```json
{
  "data": {
    "original_amount": "1000.00",
    "converted_amount": "850.00", 
    "exchange_rate": "0.85000",
    "from_currency": "USD",
    "to_currency": "EUR",
    "conversion_date": "2025-01-24T10:30:00Z"
  }
}
```

---

## Buy Module APIs

### **Vendor Management**

#### **Create Vendor**
```http
POST /buy/vendors
```

**Request Body:**
```json
{
  "vendor_code": "VEND-001",
  "vendor_name": "Acme Supplies Ltd",
  "vendor_type": "supplier", // supplier, contractor, service_provider
  "tax_id": "12-3456789",
  "payment_terms": {
    "term_type": "net_days",
    "days": 30,
    "discount_percentage": 2.0,
    "discount_days": 10
  },
  "addresses": [
    {
      "type": "billing",
      "address_line_1": "123 Business St",
      "city": "New York",
      "state": "NY", 
      "postal_code": "10001",
      "country": "US"
    }
  ],
  "banking_details": {
    "bank_name": "First National Bank",
    "account_number": "1234567890",
    "routing_number": "021000021",
    "swift_code": "FNBKUS33"
  }
}
```

#### **Create Purchase Order**
```http
POST /buy/purchase-orders
```

**Request Body:**
```json
{
  "vendor_id": "vendor-uuid",
  "po_number": "PO-2025-001",
  "po_date": "2025-01-24",
  "delivery_date": "2025-02-15",
  "currency": "USD",
  "line_items": [
    {
      "item_code": "ITEM-001",
      "description": "Office Supplies",
      "quantity": "100",
      "unit_price": "15.00",
      "line_total": "1500.00",
      "delivery_date": "2025-02-15"
    }
  ],
  "terms_conditions": "Standard purchase terms apply",
  "approval_required": true
}
```

### **Accounts Payable**

#### **Process Purchase Invoice**
```http
POST /buy/purchase-invoices
```

**Request Body:**
```json
{
  "vendor_id": "vendor-uuid", 
  "invoice_number": "INV-VENDOR-001",
  "invoice_date": "2025-01-24",
  "due_date": "2025-02-23",
  "purchase_order_id": "po-uuid",
  "currency": "USD",
  "line_items": [
    {
      "description": "Office Supplies Delivered",
      "quantity": "100",
      "unit_price": "15.00", 
      "line_amount": "1500.00",
      "tax_code": "INPUT_VAT",
      "tax_amount": "150.00"
    }
  ],
  "total_amount": "1650.00",
  "three_way_matching_required": true
}
```

#### **Create Vendor Payment**
```http
POST /buy/vendor-payments
```

**Request Body:**
```json
{
  "vendor_id": "vendor-uuid",
  "payment_date": "2025-01-24", 
  "payment_method": "bank_transfer",
  "payment_amount": "1650.00",
  "currency": "USD",
  "bank_account_id": "bank-account-uuid",
  "invoices": [
    {
      "invoice_id": "purchase-invoice-uuid",
      "payment_amount": "1650.00",
      "discount_taken": "33.00"
    }
  ],
  "payment_reference": "PAY-VENDOR-001"
}
```

---

## Sell Module APIs

### **Customer Management**

#### **Create Customer**
```http  
POST /sell/customers
```

**Request Body:**
```json
{
  "customer_code": "CUST-001",
  "customer_name": "Global Corp Inc",
  "customer_type": "corporate", // individual, corporate, government
  "tax_id": "98-7654321",
  "credit_limit": "50000.00",
  "currency": "USD",
  "payment_terms": {
    "term_type": "net_days", 
    "days": 30,
    "discount_percentage": 2.0,
    "discount_days": 10
  },
  "billing_address": {
    "address_line_1": "456 Corporate Ave",
    "city": "Chicago",
    "state": "IL",
    "postal_code": "60601", 
    "country": "US"
  },
  "loyalty_program_id": "loyalty-program-uuid"
}
```

#### **Create Sales Invoice** 
```http
POST /sell/sales-invoices
```

**Request Body:**
```json
{
  "customer_id": "customer-uuid",
  "invoice_date": "2025-01-24",
  "due_date": "2025-02-23", 
  "currency": "USD",
  "sales_order_id": "sales-order-uuid",
  "line_items": [
    {
      "item_code": "PROD-001",
      "description": "Professional Services",
      "quantity": "40",
      "unit_price": "150.00",
      "line_amount": "6000.00",
      "tax_code": "SALES_TAX",
      "tax_amount": "480.00"
    }
  ],
  "subtotal": "6000.00",
  "tax_total": "480.00", 
  "total_amount": "6480.00"
}
```

### **Accounts Receivable**

#### **Process Customer Payment**
```http
POST /sell/customer-payments
```

**Request Body:**
```json
{
  "customer_id": "customer-uuid",
  "payment_date": "2025-01-24",
  "payment_method": "bank_transfer",
  "payment_amount": "6480.00",
  "currency": "USD", 
  "bank_account_id": "bank-account-uuid",
  "invoices": [
    {
      "invoice_id": "sales-invoice-uuid",
      "payment_amount": "6480.00",
      "discount_taken": "0.00"
    }
  ],
  "payment_reference": "PAY-CUST-001"
}
```

#### **Customer Aging Report**
```http
GET /sell/aging-report
```

**Query Parameters:**
```yaml
customer_id: string (optional) - Specific customer
as_of_date: string (default: today) - Report date
aging_buckets: string[] (default: 0-30,31-60,61-90,90+) - Age ranges
include_zero_balances: boolean (default: false)
```

### **Dynamic Pricing & Loyalty**

#### **Calculate Dynamic Price**
```http
POST /sell/pricing/calculate
```

**Request Body:**
```json
{
  "customer_id": "customer-uuid",
  "item_code": "PROD-001",
  "quantity": "10",
  "order_date": "2025-01-24",
  "context": {
    "channel": "online",
    "season": "Q1",
    "promotion_codes": ["WINTER2025"]
  }
}
```

**Response:**
```json
{
  "data": {
    "base_price": "100.00",
    "applied_rules": [
      {
        "rule_name": "Volume Discount",
        "discount_percentage": "5.00",
        "discount_amount": "50.00"
      },
      {
        "rule_name": "Seasonal Promotion", 
        "discount_percentage": "10.00",
        "discount_amount": "95.00"
      }
    ],
    "final_price": "85.50",
    "loyalty_points_earned": 86
  }
}
```

---

## Misc Module APIs

### **Budget Module**

#### **Create Budget**
```http
POST /budget/budgets
```

**Request Body:**
```json
{
  "budget_name": "2025 Operating Budget",
  "fiscal_year": 2025,
  "budget_type": "operating", // operating, capital, project, department
  "currency": "USD", 
  "budget_lines": [
    {
      "account_id": "account-uuid",
      "department_id": "dept-uuid",
      "Q1_amount": "25000.00",
      "Q2_amount": "30000.00", 
      "Q3_amount": "35000.00",
      "Q4_amount": "40000.00"
    }
  ],
  "approval_workflow": true
}
```

#### **Budget vs Actual Report**
```http
GET /budget/variance-report
```

**Query Parameters:**
```yaml
budget_id: string (required) - Budget ID
period_start: string (required) - Start date
period_end: string (required) - End date
variance_threshold: number (default: 10.0) - Variance % threshold
```

---

## Account Management APIs

### **List Accounts**
```http
GET /accounts
```

**Query Parameters:**
```yaml
account_type: string (optional) - Filter by account type
root_type: string (optional) - Filter by root type (asset, liability, etc.)
is_active: boolean (optional) - Filter active/inactive accounts
search: string (optional) - Search by code or name
page: integer (default: 1) - Page number
page_size: integer (default: 50, max: 500) - Records per page
sort_by: string (default: code) - Sort field
sort_order: string (default: asc) - Sort direction
```

**Response:**
```json
{
  "data": [
    {
      "id": "acc-uuid",
      "code": "1000",
      "name": "Cash",
      "account_type": "cash",
      "root_type": "asset",
      "currency": "USD",
      "is_active": true,
      "balance": {
        "amount": "10000.00",
        "currency": "USD",
        "as_of": "2025-01-20T10:30:00Z"
      },
      "parent_account_id": null,
      "children_count": 0,
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-20T10:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 50,
    "total_records": 150,
    "total_pages": 3
  },
  "meta": {
    "request_id": "req-uuid",
    "response_time_ms": 23
  }
}
```

### **Create Account**
```http
POST /accounts
```

**Request Body:**
```json
{
  "code": "1100",
  "name": "Petty Cash",
  "description": "Small cash fund for miscellaneous expenses",
  "account_type": "cash",
  "root_type": "asset",
  "currency": "USD",
  "parent_account_id": "acc-uuid",
  "is_active": true,
  "attributes": {
    "cost_center": "CC001",
    "department": "Finance"
  }
}
```

**Response:**
```json
{
  "data": {
    "id": "acc-new-uuid",
    "code": "1100",
    "name": "Petty Cash",
    "account_type": "cash",
    "root_type": "asset",
    "currency": "USD",
    "is_active": true,
    "balance": {
      "amount": "0.00",
      "currency": "USD",
      "as_of": "2025-01-20T10:35:00Z"
    },
    "created_at": "2025-01-20T10:35:00Z",
    "updated_at": "2025-01-20T10:35:00Z"
  },
  "meta": {
    "request_id": "req-uuid",
    "response_time_ms": 45
  }
}
```

### **Get Account Details**
```http
GET /accounts/{account-id}
```

**Path Parameters:**
- `account-id`: UUID of the account

**Query Parameters:**
```yaml
include_balance: boolean (default: true) - Include current balance
include_children: boolean (default: false) - Include child accounts
include_transactions: boolean (default: false) - Include recent transactions
transaction_limit: integer (default: 10) - Number of transactions to include
```

**Response:**
```json
{
  "data": {
    "id": "acc-uuid",
    "code": "1000",
    "name": "Cash",
    "account_type": "cash",
    "root_type": "asset",
    "currency": "USD",
    "is_active": true,
    "balance": {
      "amount": "10000.00",
      "currency": "USD",
      "as_of": "2025-01-20T10:30:00Z"
    },
    "children": [
      {
        "id": "acc-child-uuid",
        "code": "1001",
        "name": "Cash on Hand",
        "balance": {
          "amount": "500.00",
          "currency": "USD"
        }
      }
    ],
    "recent_transactions": [
      {
        "id": "txn-uuid",
        "number": "TXN-2025-001",
        "date": "2025-01-20",
        "description": "Deposit",
        "amount": "1000.00",
        "transaction_type": "deposit"
      }
    ]
  }
}
```

---

## Transaction Processing APIs

### **Create Transaction**
```http
POST /transactions
```

**Request Body:**
```json
{
  "transaction_type": "manual",
  "reference_number": "TXN-2025-001",
  "date": "2025-01-20",
  "description": "Monthly office rent payment",
  "currency": "USD",
  "entries": [
    {
      "account_id": "acc-rent-expense",
      "debit_amount": "2500.00",
      "credit_amount": "0.00",
      "description": "Office rent expense",
      "cost_center": "CC001",
      "department": "Operations"
    },
    {
      "account_id": "acc-cash",
      "debit_amount": "0.00",
      "credit_amount": "2500.00",
      "description": "Cash payment for rent"
    }
  ],
  "attachments": [
    {
      "filename": "rent_invoice.pdf",
      "content_type": "application/pdf",
      "url": "https://storage.awo-erp.com/attachments/rent_invoice.pdf"
    }
  ]
}
```

**Response:**
```json
{
  "data": {
    "id": "txn-uuid",
    "number": "TXN-2025-001",
    "transaction_type": "manual",
    "status": "draft",
    "date": "2025-01-20",
    "description": "Monthly office rent payment",
    "currency": "USD",
    "total_amount": "2500.00",
    "entries": [
      {
        "id": "entry-uuid-1",
        "account_id": "acc-rent-expense",
        "account_code": "5100",
        "account_name": "Rent Expense",
        "debit_amount": "2500.00",
        "credit_amount": "0.00",
        "description": "Office rent expense"
      },
      {
        "id": "entry-uuid-2",
        "account_id": "acc-cash",
        "account_code": "1000",
        "account_name": "Cash",
        "debit_amount": "0.00",
        "credit_amount": "2500.00",
        "description": "Cash payment for rent"
      }
    ],
    "created_by": "user-uuid",
    "created_at": "2025-01-20T10:35:00Z"
  }
}
```

### **Post Transaction**
```http
POST /transactions/{transaction-id}/post
```

**Request Body:**
```json
{
  "posting_date": "2025-01-20",
  "force_post": false,
  "notes": "Regular monthly posting"
}
```

**Response:**
```json
{
  "data": {
    "id": "txn-uuid",
    "status": "posted",
    "posted_at": "2025-01-20T10:40:00Z",
    "posted_by": "user-uuid",
    "posting_reference": "POST-2025-001",
    "journal_entries": [
      {
        "account_id": "acc-rent-expense",
        "running_balance": "15000.00"
      },
      {
        "account_id": "acc-cash",
        "running_balance": "7500.00"
      }
    ]
  }
}
```

### **Reverse Transaction**
```http
POST /transactions/{transaction-id}/reverse
```

**Request Body:**
```json
{
  "reversal_date": "2025-01-20",
  "reason": "Duplicate entry correction",
  "create_new_entry": true
}
```

**Response:**
```json
{
  "data": {
    "original_transaction_id": "txn-uuid",
    "reversal_transaction_id": "txn-reversal-uuid",
    "reversal_date": "2025-01-20",
    "status": "reversed",
    "reversal_reason": "Duplicate entry correction"
  }
}
```

---

## Customer Management APIs

### **Create Customer**
```http
POST /customers
```

**Request Body:**
```json
{
  "name": "Acme Corporation",
  "customer_code": "CUST-001",
  "email": "billing@acme.com",
  "phone": "+1-555-0123",
  "tax_id": "12-3456789",
  "billing_address": {
    "street": "123 Business Ave",
    "city": "New York",
    "state": "NY",
    "postal_code": "10001",
    "country": "US"
  },
  "shipping_address": {
    "street": "456 Delivery St",
    "city": "Brooklyn",
    "state": "NY",
    "postal_code": "11201",
    "country": "US"
  },
  "payment_terms": {
    "term_type": "net_days",
    "days": 30,
    "discount_percentage": 2.0,
    "discount_days": 10
  },
  "credit_limit": "50000.00",
  "currency": "USD",
  "is_active": true
}
```

### **Create Sales Invoice**
```http
POST /invoices/sales
```

**Request Body:**
```json
{
  "customer_id": "cust-uuid",
  "invoice_date": "2025-01-20",
  "due_date": "2025-02-19",
  "currency": "USD",
  "exchange_rate": 1.0,
  "reference_number": "SO-2025-001",
  "terms": "Net 30 days",
  "line_items": [
    {
      "product_id": "prod-uuid",
      "description": "Professional Services",
      "quantity": "40.00",
      "unit_price": "150.00",
      "line_amount": "6000.00",
      "account_id": "acc-revenue",
      "tax_code": "SALES_TAX",
      "tax_amount": "480.00"
    }
  ],
  "tax_details": [
    {
      "tax_code": "SALES_TAX",
      "rate": "8.00",
      "amount": "480.00"
    }
  ],
  "total_amount": "6480.00"
}
```

---

## Vendor Management APIs

### **Create Vendor**
```http
POST /vendors
```

**Request Body:**
```json
{
  "name": "Office Supplies Inc",
  "vendor_code": "VEND-001",
  "email": "ap@officesupplies.com",
  "phone": "+1-555-0456",
  "tax_id": "98-7654321",
  "vendor_type": "supplier",
  "payment_terms": {
    "term_type": "net_days",
    "days": 30
  },
  "banking_info": {
    "bank_name": "First National Bank",
    "account_number": "1234567890",
    "routing_number": "021000021",
    "account_type": "checking"
  },
  "addresses": [
    {
      "type": "billing",
      "street": "789 Supplier Blvd",
      "city": "Chicago",
      "state": "IL",
      "postal_code": "60601",
      "country": "US"
    }
  ]
}
```

### **Create Purchase Invoice**
```http
POST /invoices/purchase
```

**Request Body:**
```json
{
  "vendor_id": "vend-uuid",
  "invoice_number": "INV-2025-001",
  "invoice_date": "2025-01-20",
  "due_date": "2025-02-19",
  "purchase_order_id": "po-uuid",
  "currency": "USD",
  "line_items": [
    {
      "description": "Office Paper - A4",
      "quantity": "100.00",
      "unit_price": "8.50",
      "line_amount": "850.00",
      "account_id": "acc-office-supplies",
      "cost_center": "CC001"
    }
  ],
  "total_amount": "850.00",
  "approval_required": true
}
```

---

## Banking & Cash Management APIs

### **Create Bank Account**
```http
POST /bank-accounts
```

**Request Body:**
```json
{
  "account_name": "Operating Account",
  "bank_name": "First National Bank",
  "account_number": "1234567890",
  "routing_number": "021000021",
  "account_type": "checking",
  "currency": "USD",
  "is_active": true,
  "reconciliation_account_id": "acc-bank-uuid"
}
```

### **Import Bank Statement**
```http
POST /bank-accounts/{bank-account-id}/statements
```

**Request Body (Multipart Form):**
```
file: bank_statement.csv
format: csv|qif|ofx|mt940
date_format: MM/dd/yyyy
delimiter: comma
```

### **Bank Reconciliation**
```http
POST /bank-accounts/{bank-account-id}/reconciliation
```

**Request Body:**
```json
{
  "statement_date": "2025-01-31",
  "ending_balance": "25000.00",
  "reconciliation_items": [
    {
      "transaction_id": "txn-uuid",
      "statement_amount": "1000.00",
      "book_amount": "1000.00",
      "status": "matched"
    },
    {
      "statement_entry": {
        "date": "2025-01-30",
        "description": "Bank Fee",
        "amount": "-25.00"
      },
      "status": "unmatched",
      "suggested_account_id": "acc-bank-fees"
    }
  ]
}
```

---

## Reporting APIs

### **Trial Balance**
```http
GET /reports/trial-balance
```

**Query Parameters:**
```yaml
as_of_date: string (required) - Date in YYYY-MM-DD format
include_zero_balances: boolean (default: false)
account_type: string (optional) - Filter by account type
format: json|pdf|excel (default: json)
```

### **Balance Sheet**
```http
GET /reports/balance-sheet
```

**Query Parameters:**
```yaml
as_of_date: string (required)
comparative: boolean (default: false)
comparative_date: string (optional)
format: json|pdf|excel (default: json)
```

### **Income Statement**
```http
GET /reports/income-statement
```

**Query Parameters:**
```yaml
start_date: string (required)
end_date: string (required)
comparative: boolean (default: false)
comparative_start_date: string (optional)
comparative_end_date: string (optional)
format: json|pdf|excel (default: json)
```

### **Custom Report**
```http
POST /reports/custom
```

**Request Body:**
```json
{
  "report_name": "Department Expense Report",
  "date_range": {
    "start_date": "2025-01-01",
    "end_date": "2025-01-31"
  },
  "filters": {
    "account_types": ["expense"],
    "departments": ["IT", "HR", "Finance"],
    "cost_centers": ["CC001", "CC002"]
  },
  "grouping": ["department", "cost_center"],
  "columns": [
    "account_code",
    "account_name", 
    "department",
    "amount"
  ],
  "format": "json"
}
```

---

## Error Handling

### **Standard Error Response**
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

### **HTTP Status Codes**
```yaml
200: Success
201: Created
400: Bad Request (validation errors)
401: Unauthorized (authentication failed)
403: Forbidden (authorization failed)
404: Not Found
409: Conflict (business rule violation)
422: Unprocessable Entity (semantic errors)
429: Too Many Requests (rate limit exceeded)
500: Internal Server Error
503: Service Unavailable
```

### **Error Code Categories**
```yaml
Validation Errors:
  - VALIDATION_ERROR: General validation failure
  - REQUIRED_FIELD: Required field missing
  - INVALID_FORMAT: Invalid data format
  - OUT_OF_RANGE: Value outside acceptable range

Business Logic Errors:
  - INSUFFICIENT_BALANCE: Account balance insufficient
  - TRANSACTION_UNBALANCED: Debits don't equal credits
  - ACCOUNT_INACTIVE: Account is not active
  - DUPLICATE_ENTRY: Duplicate transaction detected

Authorization Errors:
  - PERMISSION_DENIED: Insufficient permissions
  - TENANT_MISMATCH: Resource belongs to different tenant
  - ACCESS_RESTRICTED: Time-based access restriction

System Errors:
  - DATABASE_ERROR: Database operation failed
  - EXTERNAL_SERVICE_ERROR: Third-party service failure
  - RATE_LIMIT_EXCEEDED: API rate limit exceeded
```

---

## Rate Limiting & Quotas

### **Rate Limits**
```yaml
Tier Limits (per minute):
  Basic: 100 requests
  Professional: 500 requests
  Enterprise: 2000 requests

Endpoint-Specific Limits:
  POST /transactions: 50/minute
  POST /invoices/*: 100/minute
  GET /reports/*: 200/minute
  All other endpoints: Standard tier limit
```

### **Rate Limit Headers**
```http
X-RateLimit-Limit: 500
X-RateLimit-Remaining: 450
X-RateLimit-Reset: 1640995200
X-RateLimit-Retry-After: 60
```

### **Quota Management**
```yaml
Monthly Quotas:
  API Requests: Based on subscription tier
  Storage: 1GB per tenant (Basic), 10GB (Professional), Unlimited (Enterprise)
  Reports Generated: 100 (Basic), 1000 (Professional), Unlimited (Enterprise)
  
Usage Headers:
  X-Quota-Limit: 10000
  X-Quota-Remaining: 8500
  X-Quota-Reset: 2025-02-01T00:00:00Z
```

---

## SDK & Integration Examples

### **Go SDK Example**
```go
package main

import (
    "context"
    "github.com/awo-erp/financial-client-go"
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
- **Version**: 1.0
- **Last Updated**: January 2025  
- **Next Review**: Monthly during development
- **Approval Required**: API Lead, Security Architect

**Related Documents**
- Financial Implementation Plan (@docs/module/financial/financial-implementation-plan.md)
- Security & Compliance Guide (@docs/module/financial/financial-security-compliance-guide.md)
- Architecture Guide (@docs/module/financial/financial-architecture-guide.md)