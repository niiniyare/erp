# Enterprise Resource Planning API Documentation

## Version 1.0

---

## Overview

The ERP API is a comprehensive multi-tenant enterprise resource planning system providing:
- **Multi-tenancy** with tenant isolation
- **RBAC/ABAC** authorization
- **Financial management** (accounts, transactions, reporting)
- **HR management** (employees, persons)
- **Organizational structure** (entities, hierarchies)
- **Audit & compliance** tracking
- **Feature flags** and configuration management

---

## Base URLs

### Admin Console
```
https://console.domain.so
```
Access for system administrators managing tenants, global configurations, and cross-tenant operations.

### Tenant Applications
```
https://{subdomain}.app.domain.so
OR
https://app.domain.so (with X-Tenant-ID header)
```
Access for tenant-specific operations. Subdomain is optional if `X-Tenant-ID` header is provided.

---

## Authentication

### JWT Bearer Token
All authenticated endpoints require a JWT token in the Authorization header:

```http
Authorization: Bearer <jwt_token>
```

**JWT Scopes:**
- `api:read` - Read access to API resources
- `api:write` - Write access to API resources  
- `admin` - Administrative access

---

## Common Headers

### Required Headers

| Header | Required | Description | Example |
|--------|----------|-------------|---------|
| `X-Tenant-ID` | Yes* | Tenant UUID identifier | `550e8400-e29b-41d4-a716-446655440000` |
| `Authorization` | Yes** | JWT bearer token | `Bearer eyJhbGc...` |
| `Content-Type` | Yes*** | Request content type | `application/json` |

\* Required for all tenant-scoped operations. Optional if using subdomain routing.  
\*\* Required for all endpoints except `/auth/login` and `/auth/refresh`  
\*\*\* Required for POST, PUT, PATCH requests

### Optional Headers

| Header | Description | Example |
|--------|-------------|---------|
| `Accept` | Response format preference | `application/json` |
| `X-Request-ID` | Request tracking ID | `req-123456` |

---

## Common Response Structure

### Success Response
```json
{
  "data": { /* resource or collection */ },
  "metadata": {
    "request_id": "req-123456",
    "timestamp": "2025-01-15T10:30:00Z",
    "version": "1.0"
  }
}
```

### Paginated Response
```json
{
  "data": [ /* array of resources */ ],
  "pagination": {
    "total": 150,
    "page": 1,
    "per_page": 20,
    "total_pages": 8
  },
  "metadata": {
    "request_id": "req-123456",
    "timestamp": "2025-01-15T10:30:00Z",
    "version": "1.0"
  }
}
```

### Error Response
```json
{
  "error": {
    "code": "unauthorized",
    "message": "Invalid or expired token",
    "details": {}
  },
  "metadata": {
    "request_id": "req-123456",
    "timestamp": "2025-01-15T10:30:00Z",
    "version": "1.0"
  }
}
```

---

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `unauthorized` | 401 | Authentication required or invalid token |
| `forbidden` | 403 | Insufficient permissions |
| `not_found` | 404 | Resource not found |
| `invalid_input` | 400 | Invalid request data |
| `validation_error` | 400 | Data validation failed |
| `invalid_credentials` | 401 | Login credentials incorrect |
| `account_locked` | 403 | User account is locked |
| `conflict` | 409 | Resource conflict |
| `rate_limit_exceeded` | 429 | Too many requests |
| `internal_error` | 500 | Internal server error |

---

## Pagination

All list endpoints support pagination using query parameters:

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | integer | 1 | Page number (min: 1) |
| `per_page` | integer | 20 | Items per page (min: 1, max: 100) |

**Example:**
```http
GET /api/users?page=2&per_page=50
```

---

# API Endpoints

## 1. Authentication Service

### POST /auth/login
Authenticate user and obtain access token.

**Request:**
```http
POST /auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "mfa_code": "123456"  // Optional, if MFA enabled
}
```

**Response:** `200 OK`
```json
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "refresh_token_here",
    "token_type": "Bearer",
    "expires_in": 3600,
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "user@example.com",
      "username": "johndoe",
      "user_type": "INTERNAL",
      "is_active": true,
      "roles": ["admin", "finance_manager"]
    }
  },
  "metadata": {
    "request_id": "req-123456",
    "timestamp": "2025-01-15T10:30:00Z",
    "version": "1.0"
  }
}
```

**Errors:**
- `401 invalid_credentials` - Invalid email/password
- `403 account_locked` - Account is locked

---

### POST /auth/refresh
Refresh access token using refresh token.

**Request:**
```http
POST /auth/refresh
Content-Type: application/json

{
  "refresh_token": "refresh_token_here"
}
```

**Response:** `200 OK`
```json
{
  "data": {
    "access_token": "new_access_token",
    "refresh_token": "new_refresh_token",
    "token_type": "Bearer",
    "expires_in": 3600,
    "user": { /* user object */ }
  },
  "metadata": { /* ... */ }
}
```

---

### POST /auth/logout
Logout and invalidate current token.

**Request:**
```http
POST /auth/logout
Authorization: Bearer <token>
X-Tenant-ID: 550e8400-e29b-41d4-a716-446655440000
```

**Response:** `204 No Content`

---

### GET /auth/me
Get current authenticated user information.

**Request:**
```http
GET /auth/me
Authorization: Bearer <token>
X-Tenant-ID: 550e8400-e29b-41d4-a716-446655440000
```

**Response:** `200 OK`
```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "tenant_id": "660e8400-e29b-41d4-a716-446655440001",
    "entity_id": "770e8400-e29b-41d4-a716-446655440002",
    "person_id": "880e8400-e29b-41d4-a716-446655440003",
    "employee_id": "990e8400-e29b-41d4-a716-446655440004",
    "email": "user@example.com",
    "username": "johndoe",
    "user_type": "INTERNAL",
    "account_status": "ACTIVE",
    "is_active": true,
    "last_login_at": "2025-01-15T09:00:00Z",
    "mfa_enabled": true,
    "session_timeout_minutes": 30,
    "roles": ["admin", "finance_manager"],
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2025-01-15T10:30:00Z"
  },
  "metadata": { /* ... */ }
}
```

---

## 2. Tenant Management Service
**Admin Console Only** (`console.domain.so`)

### GET /tenants
List all tenants with pagination.

**Request:**
```http
GET /tenants?page=1&per_page=20&status=ACTIVE&search=acme
Authorization: Bearer <admin_token>
```

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | integer | Page number |
| `per_page` | integer | Items per page |
| `status` | string | Filter by status: ACTIVE, SUSPENDED, TRIAL, ARCHIVED |
| `search` | string | Search in name/email |

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "slug": "acme-corp",
      "name": "ACME Corporation",
      "email": "admin@acme.com",
      "status": "ACTIVE",
      "timezone": "America/New_York",
      "currency_code": "USD",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2025-01-15T10:30:00Z"
    }
  ],
  "pagination": {
    "total": 150,
    "page": 1,
    "per_page": 20,
    "total_pages": 8
  },
  "metadata": { /* ... */ }
}
```

---

### GET /tenants/{id}
Get tenant details by ID.

**Request:**
```http
GET /tenants/550e8400-e29b-41d4-a716-446655440000
Authorization: Bearer <admin_token>
```

**Response:** `200 OK`
```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "slug": "acme-corp",
    "name": "ACME Corporation",
    "email": "admin@acme.com",
    "subdomain": "acme",
    "status": "ACTIVE",
    "timezone": "America/New_York",
    "currency_code": "USD",
    "industry": "Technology",
    "company_size": "ENTERPRISE",
    "metadata": {
      "crm_id": "12345",
      "sales_rep": "Jane Smith"
    },
    "settings": {
      "allow_api_access": true,
      "sso_enabled": true
    },
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2025-01-15T10:30:00Z"
  },
  "metadata": { /* ... */ }
}
```

**Errors:**
- `404 not_found` - Tenant not found

---

### POST /tenants
Create new tenant.

**Request:**
```http
POST /tenants
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "slug": "new-company",
  "name": "New Company Inc",
  "email": "admin@newcompany.com",
  "timezone": "UTC",
  "currency_code": "USD",
  "industry": "Finance",
  "company_size": "MEDIUM"
}
```

**Response:** `201 Created`
```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "slug": "new-company",
    "name": "New Company Inc",
    "email": "admin@newcompany.com",
    "status": "TRIAL",
    "timezone": "UTC",
    "currency_code": "USD",
    "industry": "Finance",
    "company_size": "MEDIUM",
    "metadata": {},
    "settings": {},
    "created_at": "2025-01-15T10:30:00Z",
    "updated_at": "2025-01-15T10:30:00Z"
  },
  "metadata": { /* ... */ }
}
```

**Errors:**
- `400 invalid_input` - Invalid data provided

---

### PATCH /tenants/{id}
Update tenant information.

**Request:**
```http
PATCH /tenants/550e8400-e29b-41d4-a716-446655440000
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "name": "ACME Corporation Ltd",
  "email": "contact@acme.com",
  "status": "ACTIVE",
  "settings": {
    "sso_enabled": true
  }
}
```

**Response:** `200 OK`
```json
{
  "data": { /* updated tenant object */ },
  "metadata": { /* ... */ }
}
```

---

### GET /tenants/{id}/configuration
Get tenant configuration.

**Request:**
```http
GET /tenants/550e8400-e29b-41d4-a716-446655440000/configuration
Authorization: Bearer <admin_token>
```

**Response:** `200 OK`
```json
{
  "data": {
    "max_users": 100,
    "max_entities": 50,
    "max_transactions_per_month": 10000,
    "storage_quota": 10737418240,
    "accounting_method": "ACCRUAL",
    "fiscal_year_start_month": 1,
    "default_currency": "USD",
    "password_policy": {
      "min_length": 12,
      "require_uppercase": true,
      "require_numbers": true
    },
    "api_rate_limits": {
      "requests_per_minute": 60
    }
  },
  "metadata": { /* ... */ }
}
```

---

### PUT /tenants/{id}/configuration
Update tenant configuration.

**Request:**
```http
PUT /tenants/550e8400-e29b-41d4-a716-446655440000/configuration
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "max_users": 150,
  "max_entities": 75,
  "accounting_method": "CASH",
  "fiscal_year_start_month": 7
}
```

**Response:** `200 OK`
```json
{
  "data": { /* updated configuration */ },
  "metadata": { /* ... */ }
}
```

---

## 3. Entity Management Service

### GET /entities
List organizational entities.

**Request:**
```http
GET /entities?page=1&per_page=20&parent_id=550e8400-e29b-41d4-a716-446655440000&type=DEPARTMENT&is_active=true
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
```

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `page` | integer | No | Page number |
| `per_page` | integer | No | Items per page |
| `parent_id` | UUID | No | Filter by parent entity |
| `type` | string | No | Filter by type: COMPANY, REGIONAL, DEPARTMENT, COST_CENTER, PROJECT, DIVISION |
| `is_active` | boolean | No | Filter active only |

**Response:** `200 OK`
```json
{
  "data": [
    {
      "uuid": "550e8400-e29b-41d4-a716-446655440000",
      "parent_id": null,
      "name": "ACME Corporation",
      "code": "ACME-HQ",
      "type": "COMPANY",
      "is_active": true
    }
  ],
  "pagination": { /* ... */ },
  "metadata": { /* ... */ }
}
```

---

### GET /entities/{id}
Get entity details.

**Request:**
```http
GET /entities/550e8400-e29b-41d4-a716-446655440000
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
```

**Response:** `200 OK`
```json
{
  "data": {
    "uuid": "550e8400-e29b-41d4-a716-446655440000",
    "parent_id": null,
    "name": "ACME Corporation",
    "code": "ACME-HQ",
    "type": "COMPANY",
    "is_active": true,
    "hidden": false,
    "accrual_method": true,
    "fy_start_month": 1,
    "address": {
      "street": "123 Main St",
      "city": "New York",
      "state": "NY",
      "zip": "10001",
      "country": "USA"
    },
    "settings": {},
    "metadata": {},
    "validation_status": "VALIDATED",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2025-01-15T10:30:00Z"
  },
  "metadata": { /* ... */ }
}
```

---

### POST /entities
Create new entity.

**Request:**
```http
POST /entities
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
Content-Type: application/json

{
  "parent_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Engineering Department",
  "code": "ENG-001",
  "type": "DEPARTMENT",
  "fy_start_month": 1,
  "address": {
    "street": "456 Tech Blvd",
    "city": "San Francisco",
    "state": "CA",
    "zip": "94105",
    "country": "USA"
  }
}
```

**Response:** `201 Created`
```json
{
  "data": { /* created entity */ },
  "metadata": { /* ... */ }
}
```

---

### PATCH /entities/{id}
Update entity.

**Request:**
```http
PATCH /entities/550e8400-e29b-41d4-a716-446655440000
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
Content-Type: application/json

{
  "name": "Engineering & Product Department",
  "code": "ENG-PROD-001",
  "is_active": true,
  "settings": {
    "budget_alerts": true
  }
}
```

**Response:** `200 OK`

---

### GET /entities/{id}/hierarchy
Get entity hierarchy (ancestors/descendants).

**Request:**
```http
GET /entities/550e8400-e29b-41d4-a716-446655440000/hierarchy?direction=descendants
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
```

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `direction` | string | descendants | Options: ancestors, descendants, both |

**Response:** `200 OK`
```json
{
  "data": [
    {
      "entity_id": "550e8400-e29b-41d4-a716-446655440000",
      "ancestor_id": "550e8400-e29b-41d4-a716-446655440000",
      "descendant_id": "660e8400-e29b-41d4-a716-446655440001",
      "depth": 1
    }
  ],
  "metadata": { /* ... */ }
}
```

---

### DELETE /entities/{id}
Soft delete entity.

**Request:**
```http
DELETE /entities/550e8400-e29b-41d4-a716-446655440000
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
```

**Response:** `204 No Content`

---

## 4. User Management Service

### GET /users
List users.

**Request:**
```http
GET /users?page=1&per_page=20&entity_id=550e8400-e29b-41d4-a716-446655440000&user_type=INTERNAL&is_active=true&search=john
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
```

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `entity_id` | UUID | Filter by entity |
| `user_type` | string | Filter by type: INTERNAL, CUSTOMER, VENDOR, PARTNER, API, SERVICE, ADMIN |
| `is_active` | boolean | Filter active users |
| `search` | string | Search in email/username |

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "john.doe@acme.com",
      "username": "johndoe",
      "user_type": "INTERNAL",
      "is_active": true,
      "roles": ["finance_user", "hr_viewer"]
    }
  ],
  "pagination": { /* ... */ },
  "metadata": { /* ... */ }
}
```

---

### GET /users/{id}
Get user details.

**Response:** `200 OK`
```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "entity_id": "660e8400-e29b-41d4-a716-446655440001",
    "person_id": "770e8400-e29b-41d4-a716-446655440002",
    "employee_id": "880e8400-e29b-41d4-a716-446655440003",
    "email": "john.doe@acme.com",
    "username": "johndoe",
    "user_type": "INTERNAL",
    "account_status": "ACTIVE",
    "is_active": true,
    "last_login_at": "2025-01-15T09:00:00Z",
    "mfa_enabled": true,
    "session_timeout_minutes": 30,
    "roles": ["finance_user", "hr_viewer"],
    "created_at": "2024-06-01T00:00:00Z",
    "updated_at": "2025-01-15T10:30:00Z"
  },
  "metadata": { /* ... */ }
}
```

---

### POST /users
Create new user.

**Request:**
```http
POST /users
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
Content-Type: application/json

{
  "entity_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "jane.smith@acme.com",
  "username": "janesmith",
  "password": "SecurePass123!",
  "user_type": "INTERNAL",
  "role_ids": [
    "990e8400-e29b-41d4-a716-446655440004",
    "aa0e8400-e29b-41d4-a716-446655440005"
  ]
}
```

**Response:** `201 Created`

---

### PATCH /users/{id}
Update user.

**Request:**
```http
PATCH /users/550e8400-e29b-41d4-a716-446655440000
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
Content-Type: application/json

{
  "email": "john.doe.new@acme.com",
  "account_status": "ACTIVE",
  "is_active": true,
  "mfa_enabled": true
}
```

**Response:** `200 OK`

---

### DELETE /users/{id}
Delete user.

**Request:**
```http
DELETE /users/550e8400-e29b-41d4-a716-446655440000
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
```

**Response:** `204 No Content`

---

## 5. Role Management Service

### GET /roles
List roles.

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `entity_id` | UUID | Filter by entity |
| `role_type` | string | Filter: SYSTEM, TENANT, ENTITY, CUSTOM, FUNCTIONAL |
| `is_active` | boolean | Filter active roles |

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "finance_manager",
      "display_name": "Finance Manager",
      "role_type": "FUNCTIONAL",
      "is_active": true
    }
  ],
  "pagination": { /* ... */ },
  "metadata": { /* ... */ }
}
```

---

### GET /roles/{id}
Get role details.

**Response:** `200 OK`
```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "finance_manager",
    "display_name": "Finance Manager",
    "description": "Manages financial operations",
    "role_type": "FUNCTIONAL",
    "parent_role_id": null,
    "level": 1,
    "is_system_role": false,
    "is_active": true,
    "permission_count": 45,
    "created_at": "2024-01-01T00:00:00Z"
  },
  "metadata": { /* ... */ }
}
```

---

### POST /roles
Create new role.

**Request:**
```http
POST /roles
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
Content-Type: application/json

{
  "entity_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "department_head",
  "display_name": "Department Head",
  "description": "Manages department operations",
  "role_type": "CUSTOM",
  "parent_role_id": null
}
```

**Response:** `201 Created`

---

### POST /roles/{role_id}/permissions
Assign permissions to role.

**Request:**
```http
POST /roles/550e8400-e29b-41d4-a716-446655440000/permissions
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
Content-Type: application/json

{
  "permission_ids": [
    "660e8400-e29b-41d4-a716-446655440001",
    "770e8400-e29b-41d4-a716-446655440002",
    "880e8400-e29b-41d4-a716-446655440003"
  ]
}
```

**Response:** `200 OK`

---

### GET /roles/{role_id}/permissions
Get role permissions.

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "resource_id": "770e8400-e29b-41d4-a716-446655440002",
      "action_id": "880e8400-e29b-41d4-a716-446655440003",
      "name": "finance_accounts:read",
      "display_name": "Read Finance Accounts",
      "description": "Permission to view financial accounts",
      "effect": "ALLOW",
      "conditions": {},
      "is_active": true
    }
  ],
  "metadata": { /* ... */ }
}
```

---

## 6. Finance - Accounts Service

### GET /finance/accounts
List chart of accounts.

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `entity_id` | UUID | Filter by entity |
| `root_type` | string | Filter: ASSET, LIABILITY, EQUITY, INCOME, EXPENSE |
| `account_type` | string | Filter by account type |
| `is_active` | boolean | Filter active accounts |
| `search` | string | Search in code/name |

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "account_code": "1000",
      "account_name": "Cash",
      "root_type": "ASSET",
      "account_type": "Current Asset",
      "current_balance": "125000.00",
      "is_active": true
    }
  ],
  "pagination": { /* ... */ },
  "metadata": { /* ... */ }
}
```

---

### GET /finance/accounts/{id}
Get account details.

**Response:** `200 OK`
```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "account_code": "1000",
    "account_name": "Cash",
    "account_description": "Cash on hand and in banks",
    "root_type": "ASSET",
    "account_type": "Current Asset",
    "account_category": "Liquid Assets",
    "normal_balance": "DEBIT",
    "current_balance": "125000.00",
    "is_active": true,
    "is_control_account": false,
    "parent_account_id": null,
    "account_level": 1,
    "has_children": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2025-01-15T10:30:00Z"
  },
  "metadata": { /* ... */ }
}
```

---

### POST /finance/accounts
Create new account.

**Request:**
```http
POST /finance/accounts
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
Content-Type: application/json

{
  "entity_id": "550e8400-e29b-41d4-a716-446655440000",
  "account_code": "1100",
  "account_name": "Accounts Receivable",
  "account_description": "Money owed by customers",
  "root_type": "ASSET",
  "account_type": "Current Asset",
  "normal_balance": "DEBIT",
  "parent_account_id": null
}
```

**Response:** `201 Created`

---

### PATCH /finance/accounts/{id}
Update account.

**Request:**
```http
PATCH /finance/accounts/550e8400-e29b-41d4-a716-446655440000
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
Content-Type: application/json

{
  "account_name": "Cash and Cash Equivalents",
  "account_description": "Updated description",
  "is_active": true
}
```

**Response:** `200 OK`

---

### GET /finance/accounts/{id}/balance
Get account balance for period.

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `start_date` | datetime | No | Period start date |
| `end_date` | datetime | No | Period end date |

**Response:** `200 OK`
```json
{
  "data": {
    "account_id": "550e8400-e29b-41d4-a716-446655440000",
    "opening_balance": "100000.00",
    "closing_balance": "125000.00",
    "period_debits": "50000.00",
    "period_credits": "25000.00"
  },
  "metadata": { /* ... */ }
}
```

---

## 7. Finance - Transactions Service

### GET /finance/transactions
List financial transactions.

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `entity_id` | UUID | Filter by entity |
| `transaction_type` | string | Filter: JOURNAL_ENTRY, PAYMENT, RECEIPT, TRANSFER, ADJUSTMENT, ACCRUAL, REVERSAL |
| `transaction_status` | string | Filter: DRAFT, PENDING, APPROVED, POSTED, VOID, REVERSED |
| `start_date` | datetime | Period start |
| `end_date` | datetime | Period end |
| `search` | string | Search term |

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "transaction_number": "JE-2025-0001",
      "transaction_type": "JOURNAL_ENTRY",
      "transaction_status": "POSTED",
      "transaction_date": "2025-01-15T00:00:00Z",
      "description": "Monthly depreciation",
      "total_debit_amount": "5000.00",
      "total_credit_amount": "5000.00"
    }
  ],
  "pagination": { /* ... */ },
  "metadata": { /* ... */ }
}
```

---

### GET /finance/transactions/{id}
Get transaction details with entries.

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `include_entries` | boolean | true | Include transaction entries |

**Response:** `200 OK`
```json
{
  "data": {
    "transaction": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "transaction_number": "JE-2025-0001",
      "transaction_type": "JOURNAL_ENTRY",
      "transaction_status": "POSTED",
      "transaction_date": "2025-01-15T00:00:00Z",
      "posting_date": "2025-01-15T00:00:00Z",
      "description": "Monthly depreciation",
      "reference_number": "DEP-JAN-2025",
      "currency_code": "USD",
      "exchange_rate": "1.00",
      "total_debit_amount": "5000.00",
      "total_credit_amount": "5000.00",
      "approval_required": true,
      "approval_status": "APPROVED",
      "is_reversed": false,
      "validation_status": "VALID",
      "created_by": "660e8400-e29b-41d4-a716-446655440001",
      "created_at": "2025-01-15T10:00:00Z",
      "posted_at": "2025-01-15T10:30:00Z"
    },
    "entries": [
      {
        "id": "770e8400-e29b-41d4-a716-446655440002",
        "entry_number": 1,
        "account_id": "880e8400-e29b-41d4-a716-446655440003",
        "account_code": "6100",
        "account_name": "Depreciation Expense",
        "debit_amount": "5000.00",
        "credit_amount": "0.00",
        "description": "Monthly depreciation - Equipment",
        "reference": "DEP-JAN-2025",
        "cost_center": "CC-100",
        "department": "Operations",
        "project_id": null
      },
      {
        "id": "990e8400-e29b-41d4-a716-446655440004",
        "entry_number": 2,
        "account_id": "aa0e8400-e29b-41d4-a716-446655440005",
        "account_code": "1500-A",
        "account_name": "Accumulated Depreciation - Equipment",
        "debit_amount": "0.00",
        "credit_amount": "5000.00",
        "description": "Monthly depreciation - Equipment",
        "reference": "DEP-JAN-2025",
        "cost_center": "CC-100",
        "department": "Operations",
        "project_id": null
      }
    ]
  },
  "metadata": { /* ... */ }
}
```

---

### POST /finance/transactions
Create new transaction.

**Request:**
```http
POST /finance/transactions
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
Content-Type: application/json

{
  "entity_id": "550e8400-e29b-41d4-a716-446655440000",
  "transaction_type": "JOURNAL_ENTRY",
  "transaction_date": "2025-01-15T00:00:00Z",
  "description": "Office supplies purchase",
  "reference_number": "INV-12345",
  "currency_code": "USD",
  "entries": [
    {
      "account_id": "660e8400-e29b-41d4-a716-446655440001",
      "description": "Office supplies expense",
      "debit_amount": "500.00",
      "credit_amount": "0.00",
      "cost_center": "CC-ADMIN"
    },
    {
      "account_id": "770e8400-e29b-41d4-a716-446655440002",
      "description": "Cash payment",
      "debit_amount": "0.00",
      "credit_amount": "500.00"
    }
  ]
}
```

**Response:** `201 Created`

**Errors:**
- `400 validation_error` - Debits don't equal credits

---

### PATCH /finance/transactions/{id}
Update draft transaction.

**Request:**
```http
PATCH /finance/transactions/550e8400-e29b-41d4-a716-446655440000
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
Content-Type: application/json

{
  "description": "Updated description",
  "reference_number": "INV-12345-UPDATED",
  "entries": [ /* updated entries */ ]
}
```

**Response:** `200 OK`

---

### POST /finance/transactions/{id}/post
Post transaction to general ledger.

**Request:**
```http
POST /finance/transactions/550e8400-e29b-41d4-a716-446655440000/post
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
Content-Type: application/json

{
  "posting_date": "2025-01-15T00:00:00Z"
}
```

**Response:** `200 OK`

**Errors:**
- `400 validation_error` - Transaction invalid or already posted

---

### POST /finance/transactions/{id}/reverse
Reverse posted transaction.

**Request:**
```http
POST /finance/transactions/550e8400-e29b-41d4-a716-446655440000/reverse
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
Content-Type: application/json

{
  "reversal_date": "2025-01-16T00:00:00Z",
  "reversal_reason": "Error in original entry"
}
```

**Response:** `200 OK`
```json
{
  "data": {
    "original": { /* original transaction */ },
    "reversal": { /* reversal transaction */ }
  },
  "metadata": { /* ... */ }
}
```

---

### DELETE /finance/transactions/{id}
Delete draft transaction.

**Request:**
```http
DELETE /finance/transactions/550e8400-e29b-41d4-a716-446655440000
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
```

**Response:** `204 No Content`

---

## 8. Person Management Service

### GET /persons
List persons.

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `entity_id` | UUID | Filter by entity |
| `person_type` | string | Filter: INDIVIDUAL, EMPLOYEE, CONTACT, CUSTOMER, VENDOR, CONTRACTOR |
| `is_active` | boolean | Filter active |
| `search` | string | Search term |

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "first_name": "John",
      "last_name": "Doe",
      "email": "john.doe@example.com",
      "person_type": "EMPLOYEE",
      "is_active": true
    }
  ],
  "pagination": { /* ... */ },
  "metadata": { /* ... */ }
}
```

---

### POST /persons
Create new person.

**Request:**
```http
POST /persons
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
Content-Type: application/json

{
  "entity_id": "550e8400-e29b-41d4-a716-446655440000",
  "person_type": "EMPLOYEE",
  "first_name": "Jane",
  "last_name": "Smith",
  "middle_name": "Marie",
  "email": "jane.smith@acme.com",
  "phone": "+1-555-0123",
  "birth_date": "1990-05-15T00:00:00Z",
  "address": {
    "street": "789 Oak Ave",
    "city": "Boston",
    "state": "MA",
    "zip": "02101",
    "country": "USA"
  }
}
```

**Response:** `201 Created`

---

## 9. Employee Management Service

### GET /employees
List employees.

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `entity_id` | UUID | Filter by entity |
| `department_id` | UUID | Filter by department |
| `employment_status` | string | Filter: ACTIVE, INACTIVE, TERMINATED, ON_LEAVE, SUSPENDED |
| `search` | string | Search term |

**Response:** `200 OK`

---

### POST /employees
Create new employee.

**Request:**
```http
POST /employees
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
Content-Type: application/json

{
  "person_id": "550e8400-e29b-41d4-a716-446655440000",
  "employee_number": "EMP-001",
  "entity_id": "660e8400-e29b-41d4-a716-446655440001",
  "position_title": "Software Engineer",
  "department_id": "770e8400-e29b-41d4-a716-446655440002",
  "manager_id": "880e8400-e29b-41d4-a716-446655440003",
  "hire_date": "2025-01-15T00:00:00Z",
  "security_level": 2
}
```

**Response:** `201 Created`

---

### POST /employees/{id}/terminate
Terminate employment.

**Request:**
```http
POST /employees/550e8400-e29b-41d4-a716-446655440000/terminate
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
Content-Type: application/json

{
  "termination_date": "2025-01-31T00:00:00Z",
  "reason": "Resigned"
}
```

**Response:** `200 OK`

---

## 10. Audit Log Service

### GET /audit/logs
Query audit logs.

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `user_id` | UUID | Filter by user |
| `event_category` | string | Filter: ACCESS, ADMIN, DATA, AUTH, SYSTEM, COMPLIANCE |
| `severity` | string | Filter: LOW, INFO, WARN, HIGH, CRITICAL |
| `decision` | string | Filter: ALLOW, DENY, ERROR |
| `start_date` | datetime | Period start |
| `end_date` | datetime | Period end |
| `min_risk_score` | integer | Minimum risk score (0-100) |

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "event_type": "USER_LOGIN",
      "event_category": "AUTH",
      "severity": "INFO",
      "user_id": "660e8400-e29b-41d4-a716-446655440001",
      "entity_id": "770e8400-e29b-41d4-a716-446655440002",
      "resource_id": null,
      "action_id": null,
      "decision": "ALLOW",
      "reason": "Valid credentials",
      "risk_score": 10,
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0...",
      "context": {},
      "created_at": "2025-01-15T10:30:00Z"
    }
  ],
  "pagination": { /* ... */ },
  "metadata": { /* ... */ }
}
```

---

### GET /audit/summary
Get audit summary statistics.

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `start_date` | datetime | No | Period start |
| `end_date` | datetime | No | Period end |

**Response:** `200 OK`
```json
{
  "data": {
    "total_events": 15420,
    "by_category": {
      "ACCESS": 8500,
      "AUTH": 4200,
      "DATA": 2100,
      "ADMIN": 420,
      "SYSTEM": 150,
      "COMPLIANCE": 50
    },
    "by_severity": {
      "LOW": 12000,
      "INFO": 2800,
      "WARN": 500,
      "HIGH": 100,
      "CRITICAL": 20
    },
    "unique_users": 245,
    "denied_attempts": 87,
    "high_risk_events": 45
  },
  "metadata": { /* ... */ }
}
```

---

## 11. Access Request Service

### GET /access-requests
List access requests.

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `requester_id` | UUID | Filter by requester |
| `approval_status` | string | Filter: PENDING, APPROVED, REJECTED, EXPIRED, REVOKED |
| `request_type` | string | Filter: ROLE_ASSIGNMENT, PERMISSION_GRANT, RESOURCE_ACCESS, ELEVATION |

**Response:** `200 OK`

---

### POST /access-requests
Create access request.

**Request:**
```http
POST /access-requests
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
Content-Type: application/json

{
  "entity_id": "550e8400-e29b-41d4-a716-446655440000",
  "target_user_id": "660e8400-e29b-41d4-a716-446655440001",
  "request_type": "ROLE_ASSIGNMENT",
  "role_id": "770e8400-e29b-41d4-a716-446655440002",
  "justification": "User needs finance access for Q1 audit",
  "business_reason": "Annual audit support",
  "duration_hours": 168,
  "auto_revoke": true
}
```

**Response:** `201 Created`

---

### POST /access-requests/{id}/approve
Approve access request.

**Request:**
```http
POST /access-requests/550e8400-e29b-41d4-a716-446655440000/approve
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
Content-Type: application/json

{
  "comments": "Approved for audit support"
}
```

**Response:** `200 OK`

---

### POST /access-requests/{id}/reject
Reject access request.

**Request:**
```http
POST /access-requests/550e8400-e29b-41d4-a716-446655440000/reject
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
Content-Type: application/json

{
  "comments": "Insufficient business justification"
}
```

**Response:** `200 OK`

---

### POST /access-requests/{id}/revoke
Revoke approved access.

**Request:**
```http
POST /access-requests/550e8400-e29b-41d4-a716-446655440000/revoke
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
Content-Type: application/json

{
  "reason": "Audit completed, access no longer needed"
}
```

**Response:** `200 OK`

---

## 12. Policy Management Service (ABAC)

### GET /policies
List policies.

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `entity_id` | UUID | Filter by entity |
| `policy_type` | string | Filter: ABAC, RBAC, HYBRID, TIME_BASED, LOCATION_BASED |
| `category` | string | Filter: ACCESS, DATA_FILTER, FIELD_MASK, AUDIT, COMPLIANCE |
| `is_active` | boolean | Filter active |

**Response:** `200 OK`

---

### POST /policies/evaluate
Evaluate policies for access decision.

**Request:**
```http
POST /policies/evaluate
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
Content-Type: application/json

{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "resource_type": "finance_transaction",
  "resource_id": "660e8400-e29b-41d4-a716-446655440001",
  "action": "read",
  "context": {
    "ip_address": "192.168.1.100",
    "time_of_day": "business_hours",
    "transaction_amount": 50000
  }
}
```

**Response:** `200 OK`
```json
{
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "resource_type": "finance_transaction",
    "resource_id": "660e8400-e29b-41d4-a716-446655440001",
    "action": "read",
    "decision": "ALLOW",
    "applicable_policies": [
      "policy-id-1",
      "policy-id-2"
    ],
    "evaluation_time_ms": 45,
    "evaluated_at": "2025-01-15T10:30:00Z"
  },
  "metadata": { /* ... */ }
}
```

---

## 13. Feature Flags Service

### GET /feature-flags
List feature flags.

**Response:** `200 OK`

---

### POST /feature-flags/evaluate
Evaluate feature flag for current context.

**Request:**
```http
POST /feature-flags/evaluate
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
Content-Type: application/json

{
  "flag_name": "new_dashboard_ui",
  "context": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "entity_id": "660e8400-e29b-41d4-a716-446655440001",
    "environment": "production"
  }
}
```

**Response:** `200 OK`
```json
{
  "data": {
    "flag_name": "new_dashboard_ui",
    "enabled": true,
    "value": true,
    "source": "ROLLOUT"
  },
  "metadata": { /* ... */ }
}
```

---

### POST /feature-flags/{id}/toggle
Toggle feature flag on/off.

**Request:**
```http
POST /feature-flags/550e8400-e29b-41d4-a716-446655440000/toggle
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
Content-Type: application/json

{
  "enabled": true
}
```

**Response:** `200 OK`

---

## 14. Configuration Management Service

### GET /configurations
List system configurations.

**Response:** `200 OK`

---

### GET /configurations/{config_key}
Get specific configuration.

**Request:**
```http
GET /configurations/session_timeout_minutes
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
```

**Response:** `200 OK`

---

### PUT /configurations/{config_key}
Update configuration value.

**Request:**
```http
PUT /configurations/session_timeout_minutes
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
Content-Type: application/json

{
  "value": 45
}
```

**Response:** `200 OK`

---

### GET /configurations/audit
Get configuration change history.

**Response:** `200 OK`

---

## 15. Reporting Service

### POST /reports/financial-statement
Generate financial statement.

**Request:**
```http
POST /reports/financial-statement
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
Content-Type: application/json

{
  "entity_id": "550e8400-e29b-41d4-a716-446655440000",
  "statement_type": "BALANCE_SHEET",
  "start_date": "2025-01-01T00:00:00Z",
  "end_date": "2025-01-31T23:59:59Z",
  "include_inactive": false,
  "format": "JSON"
}
```

**Statement Types:**
- `BALANCE_SHEET` - Assets, Liabilities, Equity
- `INCOME_STATEMENT` - Revenue and Expenses
- `CASH_FLOW` - Cash flow statement
- `TRIAL_BALANCE` - Trial balance

**Response:** `200 OK`
```json
{
  "data": {
    "statement_type": "BALANCE_SHEET",
    "period": {
      "start_date": "2025-01-01T00:00:00Z",
      "end_date": "2025-01-31T23:59:59Z"
    },
    "entity": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "ACME Corporation"
    },
    "sections": [
      {
        "name": "Assets",
        "total": "500000.00",
        "accounts": [
          {
            "code": "1000",
            "name": "Cash",
            "balance": "125000.00"
          },
          {
            "code": "1100",
            "name": "Accounts Receivable",
            "balance": "75000.00"
          }
        ]
      },
      {
        "name": "Liabilities",
        "total": "200000.00",
        "accounts": [ /* ... */ ]
      },
      {
        "name": "Equity",
        "total": "300000.00",
        "accounts": [ /* ... */ ]
      }
    ],
    "totals": {
      "total_assets": "500000.00",
      "total_liabilities": "200000.00",
      "total_equity": "300000.00"
    },
    "generated_at": "2025-01-15T10:30:00Z"
  },
  "metadata": { /* ... */ }
}
```

---

### POST /reports/user-activity
Generate user activity report.

**Request:**
```http
POST /reports/user-activity
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
Content-Type: application/json

{
  "start_date": "2025-01-01T00:00:00Z",
  "end_date": "2025-01-15T23:59:59Z",
  "user_id": null,
  "group_by": "DAY"
}
```

**Response:** `200 OK`

---

### GET /reports/security-dashboard
Get security monitoring dashboard.

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `time_range` | string | 24h | Options: 1h, 24h, 7d, 30d |

**Response:** `200 OK`

---

### GET /reports/entity-hierarchy
Get entity hierarchy report.

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `root_entity_id` | UUID | No | Root entity |
| `include_inactive` | boolean | No | Include inactive |

**Response:** `200 OK`

---

## 16. Notifications Service

### GET /notifications
List user notifications.

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `user_id` | UUID | Filter by user |
| `type` | string | Filter by type |
| `acknowledged` | boolean | Filter by acknowledged status |

**Response:** `200 OK`
```json
{
  "data": [ /* notifications */ ],
  "pagination": { /* ... */ },
  "unread_count": 5,
  "metadata": { /* ... */ }
}
```

---

### POST /notifications/{id}/acknowledge
Acknowledge notification.

**Response:** `200 OK`

---

### POST /notifications/acknowledge-all
Acknowledge all notifications.

**Request:**
```http
POST /notifications/acknowledge-all
Authorization: Bearer <token>
X-Tenant-ID: 660e8400-e29b-41d4-a716-446655440001
Content-Type: application/json

{
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Response:** `200 OK`
```json
{
  "data": {
    "acknowledged_count": 12
  },
  "metadata": { /* ... */ }
}
```

---

## Rate Limiting

API requests are rate-limited per tenant:

- **Default:** 60 requests per minute
- **Authenticated:** 120 requests per minute
- **Admin:** 300 requests per minute

Rate limit headers are included in responses:
```http
X-RateLimit-Limit: 120
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1642176000
```

When rate limit is exceeded:
```http
HTTP/1.1 429 Too Many Requests
Retry-After: 30

{
  "error": {
    "code": "rate_limit_exceeded",
    "message": "Too many requests. Please retry after 30 seconds."
  }
}
```

---

## Webhooks

Subscribe to events for real-time notifications:

**Available Events:**
- `tenant.created`, `tenant.updated`
- `user.created`, `user.updated`, `user.deleted`
- `transaction.posted`, `transaction.reversed`
- `access_request.created`, `access_request.approved`
- `audit.high_risk_event`

Configure webhooks in tenant settings or via Admin Console.

---

## Support & Resources

- **API Status:** https://status.domain.so
- **Developer Portal:** https://developers.domain.so
- **API Changelog:** https://developers.domain.so/changelog
- **Support:** support@domain.so
- **Slack Community:** https://slack.domain.so

---

## Changelog

### Version 1.0 (Current)
- Initial release
- Multi-tenant architecture
- RBAC/ABAC authorization
- Finance, HR, and organizational management
- Audit and compliance features

---

*Last Updated: January 15, 2025*
