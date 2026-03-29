[<-- Back to Index](README.md)


##  API Design Philosophy

### Business-Centric Naming
```
❌ Avoid: /tenants, /provision, /rls_policies
✅ Use:   /organizations, /workspaces, /companies
         /onboarding, /setup, /activate
         /permissions, /access-control
```

### Core Principles
1. **Predictability**: Consistent patterns across all endpoints
2. **Efficiency**: Minimize round trips, support batching
3. **Clarity**: Self-documenting responses with rich metadata
4. **Resilience**: Graceful degradation, detailed error context
5. **Observability**: Request tracing, performance metrics

---


<!-- toc -->

- [ API Design Philosophy](#-api-design-philosophy)
  - [Business-Centric Naming](#business-centric-naming)
  - [Core Principles](#core-principles)
- [ Request Anatomy](#-request-anatomy)
  - [Request Headers (In-Depth)](#request-headers-in-depth)
- [ Request Body Patterns](#-request-body-patterns)
  - [Pattern 1: Simple Resource Creation](#pattern-1-simple-resource-creation)
  - [Pattern 2: Complex Nested Operations](#pattern-2-complex-nested-operations)
  - [Pattern 3: Partial Updates with Field Masks](#pattern-3-partial-updates-with-field-masks)
  - [Pattern 4: Batch Operations](#pattern-4-batch-operations)
- [ Response Anatomy](#-response-anatomy)
  - [Response Headers (In-Depth)](#response-headers-in-depth)
- [✅ Success Response Structure](#-success-response-structure)
  - [Pattern 1: Simple Resource Response (201 Created)](#pattern-1-simple-resource-response-201-created)
  - [Pattern 2: Complex Operation with Nested Results (202 Accepted)](#pattern-2-complex-operation-with-nested-results-202-accepted)
  - [Pattern 3: List Response with Rich Pagination](#pattern-3-list-response-with-rich-pagination)
  - [Pattern 4: Minimal Response (204 No Content)](#pattern-4-minimal-response-204-no-content)
- [❌ Error Response Structure](#-error-response-structure)
  - [Pattern 1: Validation Errors (400 Bad Request)](#pattern-1-validation-errors-400-bad-request)
  - [Pattern 2: Business Logic Error (422 Unprocessable Entity)](#pattern-2-business-logic-error-422-unprocessable-entity)
  - [Pattern 3: Authorization Error (403 Forbidden)](#pattern-3-authorization-error-403-forbidden)
  - [Pattern 4: Rate Limit Error (429 Too Many Requests)](#pattern-4-rate-limit-error-429-too-many-requests)
  - [Pattern 5: System Error (500 Internal Server Error)](#pattern-5-system-error-500-internal-server-error)
  - [Pattern 6: Field-Level Validation Errors (400)](#pattern-6-field-level-validation-errors-400)
- [ Advanced Response Patterns](#-advanced-response-patterns)
  - [Batch Operation Results](#batch-operation-results)
- [Security & Compliance](#security--compliance)
  - [ Authentication Flows](#-authentication-flows)
    - [1. OAuth 2.0 Token Exchange](#1-oauth-20-token-exchange)
    - [2. API Key Authentication (Server-to-Server)](#2-api-key-authentication-server-to-server)
  - [️ Permission System (RBAC/ABAC)](#-permission-system-rbacabac)
  - [ Data Encryption & Privacy](#-data-encryption--privacy)
  - [ Data Residency & Compliance](#-data-residency--compliance)
- [WebSocket & Real-time Updates](#websocket--real-time-updates)
  - [ WebSocket Connection](#-websocket-connection)
  - [ Server-Sent Events (SSE) Alternative](#-server-sent-events-sse-alternative)
- [Webhooks](#webhooks)
  - [ Webhook Configuration](#-webhook-configuration)
- [Advanced Features](#advanced-features)
  - [ Multi-Currency Support](#-multi-currency-support)
  - [ Localization (i18n)](#-localization-i18n)

<!-- tocstop -->


##  Request Anatomy

### Request Headers (In-Depth)

```http
POST /api/v1/organizations HTTP/1.1
Host: api.awoerp.com
Content-Type: application/json; charset=utf-8
Accept: application/json
Accept-Language: en-KE, sw-KE, en;q=0.9
Accept-Encoding: gzip, deflate, br

# ============================================
# AUTHENTICATION & AUTHORIZATION
# ============================================
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
X-API-Key: live_pk_1234567890abcdef                    # Alternative: API key auth

# ============================================
# REQUEST IDENTIFICATION & TRACING
# ============================================
X-Request-ID: req_2KnD8xY4mP7jR3vN                     # Client-generated UUID
X-Correlation-ID: trace_9Zm5LpQ2wX8hK1sT               # For distributed tracing
X-Session-ID: sess_4Rt7Yp3KnM2qH9Lw                    # User session tracking

# ============================================
# CONTEXT & ROUTING
# ============================================
X-Organization-ID: org_abc123xyz789                    # Multi-tenancy context
X-Workspace-Slug: mombasa-imports                       # Human-readable identifier
X-Client-Version: web/2.4.1                             # Client app version
X-Platform: web                                         # web | ios | android | desktop
X-Device-ID: dev_Kp9mN2qL7rT4sY                        # Device fingerprint

# ============================================
# CACHING & CONDITIONAL REQUESTS
# ============================================
If-None-Match: "33a64df551425fcc55e4d42a148795d9"      # ETag for cache validation
If-Modified-Since: Mon, 05 Feb 2024 10:30:00 GMT       # Timestamp-based caching
Cache-Control: no-cache                                 # Force revalidation

# ============================================
# CONTENT NEGOTIATION & PREFERENCES
# ============================================
X-Fields: id,name,slug,status,settings.currency        # Sparse fieldsets (GraphQL-style)
X-Include: subscription,billing,usage_stats            # Eager load relationships
X-Exclude: metadata.audit_log                          # Exclude specific fields
X-Response-Format: compact                              # compact | detailed | minimal

# ============================================
# IDEMPOTENCY & RETRY SAFETY
# ============================================
Idempotency-Key: idem_8TrL3mN9pQ5kY2vX                 # Prevent duplicate operations
X-Retry-Count: 2                                        # Number of retries attempted
X-Original-Request-Time: 2024-02-09T10:30:00.000Z      # Original request timestamp

# ============================================
# FEATURE FLAGS & EXPERIMENTAL FEATURES
# ============================================
X-Feature-Flags: new-billing-engine,advanced-reports   # Opt-in to beta features
X-API-Version: 2024-02-01                               # API version pinning

# ============================================
# RATE LIMITING & QUOTAS
# ============================================
X-RateLimit-Account: acc_primary                       # Rate limit bucket identifier
X-Priority: high                                        # Request priority hint

# ============================================
# DEBUGGING & DEVELOPMENT
# ============================================
X-Debug-Mode: true                                      # Enable verbose responses (dev only)
X-Explain-Query: true                                   # Return SQL query plan (dev only)
X-Trace-Sampling: 1.0                                   # Force trace sampling

# ============================================
# WEBHOOK & CALLBACK CONFIGURATION
# ============================================
X-Webhook-URL: https://yourapp.com/webhooks/setup-complete
X-Webhook-Secret: whsec_9Km4Lp8rT3nY2qX                # For webhook signature verification

# ============================================
# BUSINESS CONTEXT
# ============================================
X-User-Role: super_admin                               # Current user role
X-Timezone: Africa/Nairobi                              # User timezone
X-Currency: KES                                         # Preferred currency
X-Locale: en-KE                                         # Locale for i18n

# ============================================
# SECURITY & COMPLIANCE
# ============================================
X-Forwarded-For: 197.248.123.45                        # Client IP (via proxy)
X-Real-IP: 197.248.123.45                              # Actual client IP
X-Origin-Country: KE                                    # Request origin (geo-compliance)
X-CSRF-Token: csrf_7Np2Km9rL4tY8sX                     # CSRF protection
```

---

##  Request Body Patterns

### Pattern 1: Simple Resource Creation

```json
POST /api/v1/organizations
Content-Type: application/json

{
  "name": "Coastal Logistics Ltd",
  "slug": "coastal-logistics",
  "industry": "LOGISTICS_TRANSPORT",
  "company_size": "MEDIUM",
  "contact": {
    "email": "admin@coastallogistics.co.ke",
    "phone": "+254712345678",
    "address": {
      "street": "Mombasa Road, Miritini",
      "city": "Mombasa",
      "county": "Mombasa",
      "postal_code": "80100",
      "country": "KE"
    }
  },
  "preferences": {
    "timezone": "Africa/Nairobi",
    "currency": "KES",
    "language": "en-KE",
    "fiscal_year_start": "2024-01-01",
    "date_format": "DD/MM/YYYY",
    "number_format": {
      "decimal_separator": ".",
      "thousand_separator": ",",
      "decimal_places": 2
    }
  }
}
```

### Pattern 2: Complex Nested Operations

```json
POST /api/v1/orgs/onboard
Content-Type: application/json
Idempotency-Key: idem_8TrL3mN9pQ5kY2vX

{
  "organization": {
    "name": "Nakuru Dairy Farmers Coop",
    "slug": "nakuru-dairy-coop",
    "industry": "AGRICULTURE_DAIRY",
    "company_size": "LARGE",
    "registration": {
      "business_number": "PVT-123456/2020",
      "tax_pin": "P051234567X",
      "vat_number": "VAT-9876543",
      "registration_date": "2020-03-15"
    },
    "contact": {
      "email": "info@nakurudairy.co.ke",
      "phone": "+254722123456"
    }
  },
  
  "primary_admin": {
    "full_name": "Peter Kamau",
    "email": "peter.kamau@nakurudairy.co.ke",
    "phone": "+254722123456",
    "role": "MANAGING_DIRECTOR",
    "employee_id": "EMP-001",
    "send_welcome_email": true,
    "require_password_change": true
  },
  
  "initial_setup": {
    "subscription_plan": "PROFESSIONAL",
    "billing_cycle": "ANNUAL",
    "payment_method": "MPESA",
    
    "modules": {
      "enabled": [
        "FINANCIAL_ACCOUNTING",
        "PROCUREMENT",
        "INVENTORY_MANAGEMENT",
        "SALES_DISTRIBUTION",
        "HR_PAYROLL"
      ],
      "configure": {
        "FINANCIAL_ACCOUNTING": {
          "chart_of_accounts_template": "KENYA_MANUFACTURING",
          "accounting_method": "ACCRUAL",
          "multi_currency": true,
          "supported_currencies": ["KES", "USD", "EUR"]
        },
        "INVENTORY_MANAGEMENT": {
          "valuation_method": "WEIGHTED_AVERAGE",
          "enable_batch_tracking": true,
          "enable_serial_numbers": false,
          "default_warehouse": "Main Warehouse - Nakuru"
        }
      }
    },
    
    "integrations": {
      "mpesa": {
        "enabled": true,
        "business_shortcode": "174379",
        "environment": "production"
      },
      "kra_itax": {
        "enabled": true,
        "pin_number": "P051234567X"
      }
    },
    
    "data_import": {
      "source": "QUICKBOOKS",
      "include": ["chart_of_accounts", "customers", "suppliers", "products"],
      "file_urls": [
        "s3://imports/nakuru-dairy/customers.csv",
        "s3://imports/nakuru-dairy/products.xlsx"
      ],
      "start_async": true
    }
  },
  
  "metadata": {
    "referral_source": "SALES_TEAM",
    "sales_rep": "John Omondi",
    "campaign_id": "Q1-2024-DAIRY",
    "trial_expiry": "2024-03-09",
    "custom_fields": {
      "member_count": 450,
      "daily_milk_collection_liters": 25000
    }
  }
}
```

### Pattern 3: Partial Updates with Field Masks

```json
PATCH /api/v1/orgs/org_abc123xyz789
Content-Type: application/json
X-Update-Mask: preferences.currency,preferences.fiscal_year_start,contact.phone

{
  "preferences": {
    "currency": "USD",
    "fiscal_year_start": "2024-07-01"
  },
  "contact": {
    "phone": "+254733999888"
  }
}
```

### Pattern 4: Batch Operations

```json
POST /api/v1/orgs/batch
Content-Type: application/json

{
  "operations": [
    {
      "operation": "update_status",
      "organization_id": "org_abc123",
      "data": {
        "status": "SUSPENDED",
        "reason": "Payment overdue",
        "effective_at": "2024-02-10T00:00:00Z"
      }
    },
    {
      "operation": "update_plan",
      "organization_id": "org_xyz789",
      "data": {
        "plan": "ENTERPRISE",
        "billing_cycle": "ANNUAL",
        "apply_proration": true
      }
    },
    {
      "operation": "add_module",
      "organization_id": "org_def456",
      "data": {
        "module": "HR_PAYROLL",
        "auto_activate": true
      }
    }
  ],
  "execution_mode": "ATOMIC",
  "on_error": "ROLLBACK",
  "notify_on_complete": true,
  "callback_url": "https://yourapp.com/webhooks/batch-complete"
}
```

---

##  Response Anatomy

### Response Headers (In-Depth)

```http
HTTP/1.1 201 Created
Content-Type: application/json; charset=utf-8
Content-Length: 2847
Date: Fri, 09 Feb 2024 10:30:15 GMT
Server: nginx/1.24.0

# ============================================
# REQUEST TRACKING & DEBUGGING
# ============================================
X-Request-ID: req_2KnD8xY4mP7jR3vN                     # Echo back client request ID
X-Correlation-ID: trace_9Zm5LpQ2wX8hK1sT               # Distributed trace ID
X-Response-Time: 247                                    # Processing time in milliseconds
X-Database-Query-Time: 89                               # DB query time in ms
X-Cache-Hit: false                                      # Cache hit/miss indicator
X-Cache-Key: org:abc123:include:subscription           # Cache key used

# ============================================
# RESOURCE LOCATION & VERSIONING
# ============================================
Location: /api/v1/orgs/org_Kp9Lm2Qx7Rt4       # Newly created resource URI
Content-Location: /api/v1/orgs/org_Kp9Lm2Qx7Rt4  # Current representation
ETag: "a7f8d9e2c1b4a6f3e8d2c9b1a5f7e4d8"              # Resource version identifier
Last-Modified: Fri, 09 Feb 2024 10:30:15 GMT           # Last modification timestamp

# ============================================
# CACHING DIRECTIVES
# ============================================
Cache-Control: private, max-age=300, must-revalidate   # Cache for 5 minutes
Vary: Accept, Accept-Encoding, X-Organization-ID       # Vary cache by these headers
Age: 0                                                  # Response age in seconds

# ============================================
# RATE LIMITING & QUOTAS
# ============================================
X-RateLimit-Limit: 5000                                # Total requests allowed per window
X-RateLimit-Remaining: 4847                            # Remaining requests
X-RateLimit-Reset: 1707480000                          # Unix timestamp when limit resets
X-RateLimit-Window: 3600                               # Rate limit window in seconds (1 hour)
X-RateLimit-Policy: 5000/hour                          # Human-readable rate limit policy
X-RateLimit-Scope: account                             # Rate limit scope: ip | account | organization

X-Quota-Limit: 100                                     # Resource quota limit (e.g., orgs per account)
X-Quota-Remaining: 73                                  # Remaining quota
X-Quota-Reset: 1709251200                              # Quota reset timestamp (monthly)

# ============================================
# PAGINATION (For List Responses)
# ============================================
X-Total-Count: 1847                                    # Total records available
X-Page-Count: 93                                       # Total pages
X-Page-Number: 1                                       # Current page number
X-Page-Size: 20                                        # Records per page
X-Has-More: true                                       # More pages available
Link: <https://api.awoerp.com/v1/organizations?page=2>; rel="next",
      <https://api.awoerp.com/v1/organizations?page=93>; rel="last",
      <https://api.awoerp.com/v1/organizations?page=1>; rel="first"

# ============================================
# API VERSIONING & DEPRECATION
# ============================================
X-API-Version: 2024-02-01                              # Current API version
X-Supported-Versions: 2024-02-01, 2023-11-15, 2023-08-01
X-Deprecated-Version: 2023-05-01                       # Deprecated version warning
X-Deprecation-Date: 2024-05-01                         # When deprecated version will be removed
X-Sunset: Sat, 01 Jun 2024 00:00:00 GMT                # RFC 8594 Sunset header
Deprecation: true                                       # Endpoint is deprecated
Warning: 299 - "This endpoint version will be removed on 2024-05-01. Migrate to v2024-02-01"

# ============================================
# SECURITY HEADERS
# ============================================
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Content-Security-Policy: default-src 'self'
Referrer-Policy: strict-origin-when-cross-origin

# ============================================
# IDEMPOTENCY & OPERATION STATUS
# ============================================
Idempotency-Key: idem_8TrL3mN9pQ5kY2vX                 # Echo back idempotency key
X-Idempotent-Replayed: false                           # Whether this was a replayed request
X-Operation-ID: op_7Nm3Kp2Lq9Yx4Rt                     # Async operation identifier
X-Operation-Status: COMPLETED                          # PROCESSING | COMPLETED | FAILED

# ============================================
# WEBHOOK & CALLBACK
# ============================================
X-Webhook-Delivered: true                              # Webhook delivery status
X-Webhook-Delivery-Attempts: 1                         # Number of delivery attempts

# ============================================
# BUSINESS CONTEXT
# ============================================
X-Organization-ID: org_Kp9Lm2Qx7Rt4                   # Created organization ID
X-Organization-Slug: nakuru-dairy-coop                 # Human-readable slug
X-Subscription-Tier: PROFESSIONAL                      # Current subscription tier
X-Feature-Flags: advanced-reports,multi-currency       # Active feature flags

# ============================================
# PERFORMANCE & MONITORING
# ============================================
X-Backend-Server: app-server-03                        # Backend server that handled request
X-Load-Balancer: lb-ke-01                              # Load balancer identifier
X-Region: africa-east-1                                # Geographic region
Server-Timing: db;dur=89, cache;dur=12, total;dur=247 # Performance timing breakdown

# ============================================
# CONTENT NEGOTIATION
# ============================================
Content-Language: en-KE                                # Response language
X-Response-Format: detailed                            # Response format used
```

---

## ✅ Success Response Structure

### Pattern 1: Simple Resource Response (201 Created)

```json
HTTP/1.1 201 Created
Location: /api/v1/orgs/org_Kp9Lm2Qx7Rt4
X-Request-ID: req_2KnD8xY4mP7jR3vN
X-Response-Time: 247
ETag: "a7f8d9e2c1b4a6f3"

{
  "success": true,
  "data": {
    "id": "org_Kp9Lm2Qx7Rt4",
    "object": "organization",
    "slug": "coastal-logistics",
    "name": "Coastal Logistics Ltd",
    "display_name": "Coastal Logistics",
    "status": "ACTIVE",
    
    "industry": {
      "code": "LOGISTICS_TRANSPORT",
      "name": "Logistics & Transportation",
      "category": "Services"
    },
    
    "company_size": {
      "code": "MEDIUM",
      "range": "50-200 employees",
      "tier": "SME"
    },
    
    "contact": {
      "email": "admin@coastallogistics.co.ke",
      "phone": "+254712345678",
      "verified_email": false,
      "verified_phone": false,
      "address": {
        "formatted": "Mombasa Road, Miritini, Mombasa 80100, Kenya",
        "street": "Mombasa Road, Miritini",
        "city": "Mombasa",
        "county": "Mombasa",
        "postal_code": "80100",
        "country": "KE",
        "coordinates": {
          "latitude": -4.0435,
          "longitude": 39.6682
        }
      }
    },
    
    "subscription": {
      "plan": "STARTER",
      "status": "TRIALING",
      "trial_ends_at": "2024-03-09T23:59:59Z",
      "current_period_start": "2024-02-09T10:30:15Z",
      "current_period_end": "2024-03-09T23:59:59Z",
      "billing_cycle": "MONTHLY",
      "currency": "KES",
      "amount": 0,
      "next_billing_date": "2024-03-09T00:00:00Z"
    },
    
    "preferences": {
      "timezone": "Africa/Nairobi",
      "currency": "KES",
      "language": "en-KE",
      "fiscal_year_start": "2024-01-01",
      "date_format": "DD/MM/YYYY",
      "time_format": "24h",
      "week_start": "monday",
      "number_format": {
        "decimal_separator": ".",
        "thousand_separator": ",",
        "decimal_places": 2,
        "currency_symbol": "KSh",
        "currency_position": "before"
      }
    },
    
    "access": {
      "subdomain": "coastal-logistics",
      "custom_domain": null,
      "web_url": "https://coastal-logistics.awoerp.com",
      "api_url": "https://api.awoerp.com/v1",
      "sso_enabled": false,
      "mfa_required": false
    },
    
    "limits": {
      "max_users": 10,
      "max_storage_gb": 10,
      "max_transactions_per_month": 1000,
      "max_api_calls_per_hour": 1000,
      "modules_allowed": 3
    },
    
    "timestamps": {
      "created_at": "2024-02-09T10:30:15.247Z",
      "updated_at": "2024-02-09T10:30:15.247Z",
      "activated_at": "2024-02-09T10:30:15.247Z",
      "suspended_at": null,
      "deleted_at": null
    },
    
    "metadata": {
      "created_by": {
        "id": "user_9Km2Lp7Qx3Rt",
        "name": "System Onboarding",
        "type": "SYSTEM"
      },
      "creation_source": "WEB_SIGNUP",
      "creation_ip": "197.248.123.45",
      "creation_user_agent": "Mozilla/5.0...",
      "referral_source": null,
      "tags": ["new", "trial", "kenya"],
      "custom_fields": {}
    }
  },
  
  "meta": {
    "request_id": "req_2KnD8xY4mP7jR3vN",
    "api_version": "2024-02-01",
    "response_time_ms": 247,
    "timestamp": "2024-02-09T10:30:15.247Z",
    "warnings": [],
    "next_steps": [
      {
        "action": "verify_email",
        "description": "Verify organization email address",
        "url": "/api/v1/orgs/org_Kp9Lm2Qx7Rt4/verify-email",
        "required": false
      },
      {
        "action": "create_admin_user",
        "description": "Create administrator account",
        "url": "/api/v1/orgs/org_Kp9Lm2Qx7Rt4/users",
        "required": true
      },
      {
        "action": "complete_setup",
        "description": "Complete organization setup wizard",
        "url": "/api/v1/orgs/org_Kp9Lm2Qx7Rt4/setup",
        "required": true
      }
    ]
  },
  
  "links": {
    "self": "/api/v1/orgs/org_Kp9Lm2Qx7Rt4",
    "dashboard": "/api/v1/orgs/org_Kp9Lm2Qx7Rt4/dashboard",
    "users": "/api/v1/orgs/org_Kp9Lm2Qx7Rt4/users",
    "settings": "/api/v1/orgs/org_Kp9Lm2Qx7Rt4/settings",
    "subscription": "/api/v1/orgs/org_Kp9Lm2Qx7Rt4/subscription",
    "billing": "/api/v1/orgs/org_Kp9Lm2Qx7Rt4/billing",
    "usage": "/api/v1/orgs/org_Kp9Lm2Qx7Rt4/usage",
    "webhooks": "/api/v1/orgs/org_Kp9Lm2Qx7Rt4/webhooks",
    "audit_log": "/api/v1/orgs/org_Kp9Lm2Qx7Rt4/audit-log"
  }
}
```

### Pattern 2: Complex Operation with Nested Results (202 Accepted)

```json
HTTP/1.1 202 Accepted
Location: /api/v1/jobs/job_8Nm3Kp9Lq2Yx7Rt
X-Request-ID: req_5Tp8Km2Nq4Rx9Ys
X-Operation-ID: job_8Nm3Kp9Lq2Yx7Rt
X-Webhook-URL: https://yourapp.com/webhooks/onboarding-complete

{
  "success": true,
  "data": {
    "job_id": "job_8Nm3Kp9Lq2Yx7Rt",
    "object": "async_job",
    "type": "ORGANIZATION_ONBOARDING",
    "status": "PROCESSING",
    
    "organization": {
      "id": "org_Np7Km3Qx2Rt9",
      "slug": "nakuru-dairy-coop",
      "name": "Nakuru Dairy Farmers Coop",
      "status": "PROVISIONING"
    },
    
    "progress": {
      "current_step": "CREATING_DATABASE_SCHEMA",
      "total_steps": 8,
      "completed_steps": 2,
      "percentage": 25,
      "current_activity": "Setting up multi-tenant database partitions",
      "estimated_time_remaining_seconds": 45
    },
    
    "steps": [
      {
        "step": 1,
        "name": "VALIDATE_INPUTS",
        "status": "COMPLETED",
        "started_at": "2024-02-09T10:35:00.100Z",
        "completed_at": "2024-02-09T10:35:00.287Z",
        "duration_ms": 187,
        "result": {
          "validation_passed": true,
          "checks_performed": 12
        }
      },
      {
        "step": 2,
        "name": "CREATE_ORGANIZATION_RECORD",
        "status": "COMPLETED",
        "started_at": "2024-02-09T10:35:00.288Z",
        "completed_at": "2024-02-09T10:35:00.445Z",
        "duration_ms": 157,
        "result": {
          "organization_id": "org_Np7Km3Qx2Rt9",
          "slug": "nakuru-dairy-coop"
        }
      },
      {
        "step": 3,
        "name": "CREATING_DATABASE_SCHEMA",
        "status": "PROCESSING",
        "started_at": "2024-02-09T10:35:00.446Z",
        "progress": "Creating tables and indexes"
      },
      {
        "step": 4,
        "name": "SETUP_MODULES",
        "status": "PENDING"
      },
      {
        "step": 5,
        "name": "CREATE_ADMIN_USER",
        "status": "PENDING"
      },
      {
        "step": 6,
        "name": "IMPORT_INITIAL_DATA",
        "status": "PENDING"
      },
      {
        "step": 7,
        "name": "CONFIGURE_INTEGRATIONS",
        "status": "PENDING"
      },
      {
        "step": 8,
        "name": "SEND_WELCOME_EMAILS",
        "status": "PENDING"
      }
    ],
    
    "timestamps": {
      "created_at": "2024-02-09T10:35:00.000Z",
      "started_at": "2024-02-09T10:35:00.100Z",
      "estimated_completion_at": "2024-02-09T10:35:45.000Z",
      "completed_at": null
    },
    
    "callback": {
      "webhook_url": "https://yourapp.com/webhooks/onboarding-complete",
      "webhook_events": ["job.completed", "job.failed"],
      "email_notifications": ["admin@nakurudairy.co.ke"],
      "sms_notifications": ["+254722123456"]
    }
  },
  
  "meta": {
    "request_id": "req_5Tp8Km2Nq4Rx9Ys",
    "api_version": "2024-02-01",
    "response_time_ms": 78,
    "timestamp": "2024-02-09T10:35:00.078Z",
    "polling": {
      "recommended_interval_seconds": 2,
      "max_wait_seconds": 120,
      "timeout_at": "2024-02-09T10:37:00.000Z"
    },
    "warnings": [
      {
        "code": "DATA_IMPORT_SIZE_LARGE",
        "message": "Data import file size is large (45MB). This may take several minutes.",
        "severity": "INFO"
      }
    ]
  },
  
  "links": {
    "self": "/api/v1/jobs/job_8Nm3Kp9Lq2Yx7Rt",
    "poll": "/api/v1/jobs/job_8Nm3Kp9Lq2Yx7Rt/status",
    "cancel": "/api/v1/jobs/job_8Nm3Kp9Lq2Yx7Rt/cancel",
    "organization": "/api/v1/orgs/org_Np7Km3Qx2Rt9",
    "logs": "/api/v1/jobs/job_8Nm3Kp9Lq2Yx7Rt/logs"
  }
}
```

### Pattern 3: List Response with Rich Pagination

```json
HTTP/1.1 200 OK
X-Request-ID: req_7Km9Lp3Qx2Rt5Yn
X-Total-Count: 1847
X-Page-Count: 93
X-Has-More: true
Link: <https://api.awoerp.com/v1/organizations?page=2&limit=20>; rel="next",
      <https://api.awoerp.com/v1/organizations?page=93&limit=20>; rel="last"

{
  "success": true,
  "data": [
    {
      "id": "org_Kp9Lm2Qx7Rt4",
      "object": "organization",
      "slug": "coastal-logistics",
      "name": "Coastal Logistics Ltd",
      "status": "ACTIVE",
      "industry": "LOGISTICS_TRANSPORT",
      "subscription_plan": "PROFESSIONAL",
      "created_at": "2024-01-15T08:00:00Z",
      
      "summary": {
        "active_users": 45,
        "total_revenue_this_month": {
          "amount": 125000,
          "currency": "KES",
          "formatted": "KSh 125,000.00"
        },
        "transactions_this_month": 1247,
        "storage_used_mb": 3456
      }
    },
    {
      "id": "org_Np7Km3Qx2Rt9",
      "object": "organization",
      "slug": "nakuru-dairy-coop",
      "name": "Nakuru Dairy Farmers Coop",
      "status": "ACTIVE",
      "industry": "AGRICULTURE_DAIRY",
      "subscription_plan": "ENTERPRISE",
      "created_at": "2024-01-10T12:30:00Z",
      
      "summary": {
        "active_users": 187,
        "total_revenue_this_month": {
          "amount": 4567000,
          "currency": "KES",
          "formatted": "KSh 4,567,000.00"
        },
        "transactions_this_month": 8934,
        "storage_used_mb": 12847
      }
    }
  ],
  
  "pagination": {
    "page": 1,
    "limit": 20,
    "total_records": 1847,
    "total_pages": 93,
    "has_previous": false,
    "has_next": true,
    "previous_page": null,
    "next_page": 2,
    "first_page": 1,
    "last_page": 93,
    
    "cursors": {
      "before": null,
      "after": "eyJpZCI6Im9yZ19OcDdLbTNReDJSdDkiLCJjcmVhdGVkX2F0IjoiMjAyNC0wMS0xMFQxMjozMDowMFoifQ==",
      "has_more": true
    },
    
    "meta": {
      "records_on_page": 20,
      "first_record_index": 1,
      "last_record_index": 20
    }
  },
  
  "filters": {
    "applied": {
      "status": ["ACTIVE"],
      "created_after": "2024-01-01T00:00:00Z",
      "subscription_plan": ["PROFESSIONAL", "ENTERPRISE"]
    },
    "available": {
      "status": ["ACTIVE", "SUSPENDED", "ARCHIVED", "PENDING"],
      "industry": ["LOGISTICS_TRANSPORT", "AGRICULTURE_DAIRY", "RETAIL", "..."],
      "subscription_plan": ["STARTER", "PROFESSIONAL", "ENTERPRISE"],
      "company_size": ["SMALL", "MEDIUM", "LARGE"]
    }
  },
  
  "sorting": {
    "applied": ["-created_at"],
    "available": ["name", "created_at", "updated_at", "status", "subscription_plan"]
  },
  
  "aggregations": {
    "total_active": 1456,
    "total_suspended": 234,
    "total_revenue_all_orgs": {
      "amount": 45678900,
      "currency": "KES",
      "formatted": "KSh 45,678,900.00"
    },
    "by_subscription_plan": {
      "STARTER": 567,
      "PROFESSIONAL": 689,
      "ENTERPRISE": 200
    },
    "by_industry": {
      "AGRICULTURE": 456,
      "LOGISTICS": 234,
      "RETAIL": 345,
      "MANUFACTURING": 178,
      "OTHER": 243
    }
  },
  
  "meta": {
    "request_id": "req_7Km9Lp3Qx2Rt5Yn",
    "api_version": "2024-02-01",
    "response_time_ms": 156,
    "timestamp": "2024-02-09T11:00:00.000Z",
    "cached": false,
    "query_cost": {
      "complexity": 15,
      "cost_units": 2
    }
  },
  
  "links": {
    "self": "/api/v1/organizations?page=1&limit=20",
    "first": "/api/v1/organizations?page=1&limit=20",
    "prev": null,
    "next": "/api/v1/organizations?page=2&limit=20",
    "last": "/api/v1/organizations?page=93&limit=20",
    "export": "/api/v1/orgs/export?format=csv&filters=...",
    "bulk_actions": "/api/v1/orgs/batch"
  }
}
```

### Pattern 4: Minimal Response (204 No Content)

```http
HTTP/1.1 204 No Content
X-Request-ID: req_3Np8Km5Lq9Rx2Yt
X-Response-Time: 34
X-Operation-Status: COMPLETED
ETag: "b8g9e3f4d2c5b7g4f9e3d2c0b6g8f5e9"
```

---

## ❌ Error Response Structure

### Pattern 1: Validation Errors (400 Bad Request)

```json
HTTP/1.1 400 Bad Request
Content-Type: application/json
X-Request-ID: req_9Lp2Km8Qx5Rt3Yn

{
  "success": false,
  "error": {
    "type": "VALIDATION_ERROR",
    "code": "ERR_INVALID_REQUEST_BODY",
    "message": "The request contains invalid or missing fields. Please review and correct the errors below.",
    
    "details": {
      "invalid_fields": 3,
      "errors": [
        {
          "field": "organization.slug",
          "code": "INVALID_FORMAT",
          "message": "Slug must contain only lowercase letters, numbers, and hyphens",
          "current_value": "Coastal_Logistics!",
          "expected_format": "^[a-z0-9-]+$",
          "suggestion": "coastal-logistics"
        },
        {
          "field": "organization.contact.email",
          "code": "INVALID_EMAIL",
          "message": "Email address is not in a valid format",
          "current_value": "admin@coastallogistics",
          "expected_format": "user@domain.com",
          "suggestion": "admin@coastallogistics.co.ke"
        },
        {
          "field": "initial_setup.modules.enabled",
          "code": "EXCEEDS_LIMIT",
          "message": "Number of enabled modules exceeds plan limit",
          "current_value": 8,
          "max_allowed": 5,
          "plan": "STARTER",
          "suggestion": "Upgrade to PROFESSIONAL plan to enable 8 modules"
        }
      ]
    },
    
    "request_id": "req_9Lp2Km8Qx5Rt3Yn",
    "timestamp": "2024-02-09T11:10:00.000Z",
    "documentation_url": "https://docs.awoerp.com/api/errors/validation-error",
    
    "help": {
      "action": "Please correct the highlighted fields and resubmit your request",
      "support_contact": "support@awoerp.com",
      "chat_available": true
    }
  },
  
  "meta": {
    "api_version": "2024-02-01",
    "response_time_ms": 23
  }
}
```

### Pattern 2: Business Logic Error (422 Unprocessable Entity)

```json
HTTP/1.1 422 Unprocessable Entity
Content-Type: application/json
X-Request-ID: req_4Mp7Kn3Lp8Qx2Rt

{
  "success": false,
  "error": {
    "type": "BUSINESS_RULE_VIOLATION",
    "code": "ERR_SUBDOMAIN_ALREADY_TAKEN",
    "message": "The subdomain 'nakuru-dairy' is already registered by another organization",
    
    "details": {
      "field": "organization.slug",
      "requested_value": "nakuru-dairy",
      "conflict_type": "DUPLICATE_SUBDOMAIN",
      "existing_organization": {
        "id": "org_Xy9Km2Lp7Qt4",
        "name": "Nakuru Dairy Products Ltd",
        "registered_on": "2023-08-15T09:00:00Z"
      },
      "suggestions": [
        "nakuru-dairy-coop",
        "nakuru-dairy-farmers",
        "nakuru-dairy-2024"
      ]
    },
    
    "resolution": {
      "action": "CHOOSE_DIFFERENT_SLUG",
      "description": "Please select a different subdomain from the suggestions or create your own unique subdomain",
      "retry_allowed": true,
      "alternative_endpoints": null
    },
    
    "request_id": "req_4Mp7Kn3Lp8Qx2Rt",
    "timestamp": "2024-02-09T11:15:00.000Z",
    "documentation_url": "https://docs.awoerp.com/api/errors/subdomain-conflict"
  }
}
```

### Pattern 3: Authorization Error (403 Forbidden)

```json
HTTP/1.1 403 Forbidden
Content-Type: application/json
X-Request-ID: req_8Kp3Nm9Lq2Rx7Yt

{
  "success": false,
  "error": {
    "type": "AUTHORIZATION_ERROR",
    "code": "ERR_ORGANIZATION_SUSPENDED",
    "message": "This organization has been suspended and access is restricted",
    
    "details": {
      "organization_id": "org_Kp9Lm2Qx7Rt4",
      "organization_name": "Coastal Logistics Ltd",
      "status": "SUSPENDED",
      "reason": "Payment overdue for 30+ days",
      "suspended_at": "2024-01-25T00:00:00Z",
      "suspended_by": {
        "id": "user_admin",
        "name": "Billing System",
        "type": "SYSTEM"
      },
      
      "restrictions": {
        "read_access": "LIMITED",
        "write_access": "BLOCKED",
        "user_login": "ADMIN_ONLY",
        "api_access": "READ_ONLY",
        "features_disabled": [
          "INVOICE_GENERATION",
          "PAYMENT_PROCESSING",
          "INVENTORY_UPDATES",
          "NEW_TRANSACTIONS"
        ]
      },
      
      "outstanding_balance": {
        "amount": 45000,
        "currency": "KES",
        "formatted": "KSh 45,000.00",
        "due_date": "2024-01-15T00:00:00Z",
        "days_overdue": 25
      }
    },
    
    "resolution": {
      "action": "SETTLE_OUTSTANDING_PAYMENT",
      "description": "To restore full access, please settle the outstanding balance of KSh 45,000.00",
      "steps": [
        {
          "step": 1,
          "action": "View outstanding invoices",
          "url": "/api/v1/orgs/org_Kp9Lm2Qx7Rt4/billing/invoices?status=UNPAID"
        },
        {
          "step": 2,
          "action": "Make payment via M-Pesa or Bank Transfer",
          "url": "/api/v1/orgs/org_Kp9Lm2Qx7Rt4/billing/pay"
        },
        {
          "step": 3,
          "action": "Contact billing support if you need payment arrangements",
          "contact": {
            "email": "billing@awoerp.com",
            "phone": "+254-700-123456",
            "hours": "Mon-Fri 8:00 AM - 6:00 PM EAT"
          }
        }
      ],
      "automatic_reactivation": "Account will be reactivated within 1 hour of payment confirmation",
      "grace_period_expires": "2024-02-15T00:00:00Z",
      "data_deletion_warning": "Organization data will be archived if suspension exceeds 60 days"
    },
    
    "request_id": "req_8Kp3Nm9Lq2Rx7Yt",
    "timestamp": "2024-02-09T11:20:00.000Z",
    "documentation_url": "https://docs.awoerp.com/api/errors/organization-suspended"
  }
}
```

### Pattern 4: Rate Limit Error (429 Too Many Requests)

```json
HTTP/1.1 429 Too Many Requests
Content-Type: application/json
Retry-After: 3600
X-RateLimit-Limit: 5000
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1707483600
X-Request-ID: req_2Lp9Km7Qx3Rt5Yn

{
  "success": false,
  "error": {
    "type": "RATE_LIMIT_EXCEEDED",
    "code": "ERR_TOO_MANY_REQUESTS",
    "message": "You have exceeded your API rate limit. Please retry after the specified time",
    
    "details": {
      "limit_type": "API_CALLS",
      "limit_scope": "ORGANIZATION",
      "limit_window": "1 hour",
      "limit_policy": "5000 requests per hour",
      
      "current_usage": {
        "requests_made": 5000,
        "requests_allowed": 5000,
        "requests_remaining": 0,
        "usage_percentage": 100
      },
      
      "window": {
        "started_at": "2024-02-09T10:00:00Z",
        "resets_at": "2024-02-09T11:00:00Z",
        "seconds_until_reset": 3600,
        "formatted_reset_time": "11:00 AM EAT"
      },
      
      "organization": {
        "id": "org_Kp9Lm2Qx7Rt4",
        "plan": "PROFESSIONAL",
        "rate_limits": {
          "per_hour": 5000,
          "per_day": 100000,
          "burst_limit": 100
        }
      }
    },
    
    "resolution": {
      "action": "WAIT_AND_RETRY",
      "description": "Wait until rate limit resets or upgrade your plan for higher limits",
      "retry_after_seconds": 3600,
      "retry_at": "2024-02-09T11:00:00Z",
      
      "alternatives": [
        {
          "option": "Upgrade to ENTERPRISE plan",
          "benefit": "10,000 requests/hour (2x increase)",
          "url": "/api/v1/orgs/org_Kp9Lm2Qx7Rt4/subscription/upgrade"
        },
        {
          "option": "Implement request batching",
          "benefit": "Reduce API calls by combining multiple operations",
          "documentation": "https://docs.awoerp.com/api/batching"
        },
        {
          "option": "Use webhooks for real-time updates",
          "benefit": "Eliminate polling and reduce API calls",
          "documentation": "https://docs.awoerp.com/webhooks"
        }
      ],
      
      "best_practices": [
        "Implement exponential backoff for retries",
        "Cache responses where appropriate",
        "Use ETags for conditional requests",
        "Batch multiple operations into single requests",
        "Subscribe to webhooks instead of polling"
      ]
    },
    
    "request_id": "req_2Lp9Km7Qx3Rt5Yn",
    "timestamp": "2024-02-09T10:30:00.000Z",
    "documentation_url": "https://docs.awoerp.com/api/rate-limits"
  }
}
```

### Pattern 5: System Error (500 Internal Server Error)

```json
HTTP/1.1 500 Internal Server Error
Content-Type: application/json
X-Request-ID: req_5Np8Km3Lq9Rx2Yt

{
  "success": false,
  "error": {
    "type": "INTERNAL_SERVER_ERROR",
    "code": "ERR_DATABASE_CONNECTION_FAILED",
    "message": "We encountered an unexpected error while processing your request. Our team has been notified and is investigating.",
    
    "details": {
      "error_id": "err_7Km9Lp2Qx5Rt3Yn",
      "occurred_at": "2024-02-09T11:25:00.000Z",
      "severity": "HIGH",
      "component": "DATABASE_LAYER",
      "automatically_reported": true,
      "incident_id": "INC-20240209-0047"
    },
    
    "resolution": {
      "action": "RETRY_LATER",
      "description": "This is a temporary issue on our end. Please retry your request in a few moments",
      "retry_allowed": true,
      "recommended_retry_delay_seconds": 30,
      "exponential_backoff": true,
      "max_retries": 3,
      
      "escalation": {
        "if_persists": "Contact support if this error continues after 3 retries",
        "support_contact": {
          "email": "support@awoerp.com",
          "phone": "+254-700-123456",
          "chat": "https://awoerp.com/support/chat",
          "hours": "24/7"
        },
        "reference_codes": {
          "request_id": "req_5Np8Km3Lq9Rx2Yt",
          "error_id": "err_7Km9Lp2Qx5Rt3Yn",
          "incident_id": "INC-20240209-0047"
        }
      },
      
      "status_page": {
        "url": "https://status.awoerp.com",
        "current_status": "DEGRADED_PERFORMANCE",
        "message": "We are experiencing intermittent database connectivity issues. Our engineers are actively working on a resolution."
      }
    },
    
    "request_id": "req_5Np8Km3Lq9Rx2Yt",
    "timestamp": "2024-02-09T11:25:00.000Z",
    "documentation_url": "https://docs.awoerp.com/api/errors/server-errors",
    
    "debugging": {
      "safe_to_retry": true,
      "idempotent_operation": true,
      "data_modified": false,
      "partial_completion": false
    }
  }
}
```

### Pattern 6: Field-Level Validation Errors (400)

```json
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
  "success": false,
  "error": {
    "type": "VALIDATION_ERROR",
    "code": "ERR_MULTIPLE_VALIDATION_FAILURES",
    "message": "Multiple validation errors were found in your request",
    
    "validation_errors": [
      {
        "field": "organization.name",
        "location": "body",
        "code": "REQUIRED_FIELD_MISSING",
        "message": "Organization name is required",
        "severity": "ERROR",
        "constraint": {
          "type": "required",
          "value": true
        }
      },
      {
        "field": "organization.slug",
        "location": "body",
        "code": "PATTERN_MISMATCH",
        "message": "Slug must be lowercase alphanumeric with hyphens only",
        "severity": "ERROR",
        "current_value": "Nakuru_Dairy!",
        "constraint": {
          "type": "pattern",
          "pattern": "^[a-z0-9-]+$",
          "description": "lowercase letters, numbers, and hyphens"
        },
        "suggestions": [
          "nakuru-dairy",
          "nakuru-dairy-coop"
        ]
      },
      {
        "field": "organization.contact.email",
        "location": "body",
        "code": "INVALID_EMAIL_FORMAT",
        "message": "Email address is not valid",
        "severity": "ERROR",
        "current_value": "admin@nakuru",
        "constraint": {
          "type": "email",
          "must_have_domain": true
        }
      },
      {
        "field": "organization.contact.phone",
        "location": "body",
        "code": "INVALID_PHONE_FORMAT",
        "message": "Phone number must be in international format",
        "severity": "WARNING",
        "current_value": "0722123456",
        "constraint": {
          "type": "phone",
          "format": "E.164",
          "country_code_required": true
        },
        "suggestions": [
          "+254722123456"
        ],
        "auto_corrected": false
      },
      {
        "field": "initial_setup.modules.enabled",
        "location": "body",
        "code": "ARRAY_LENGTH_EXCEEDED",
        "message": "Too many modules selected for STARTER plan",
        "severity": "ERROR",
        "current_value": [
          "FINANCIAL_ACCOUNTING",
          "INVENTORY_MANAGEMENT",
          "PROCUREMENT",
          "SALES_DISTRIBUTION",
          "HR_PAYROLL",
          "MANUFACTURING"
        ],
        "constraint": {
          "type": "maxItems",
          "maxItems": 3,
          "current_count": 6,
          "plan_limit": "STARTER"
        },
        "resolution": {
          "option_1": "Select maximum 3 modules for STARTER plan",
          "option_2": "Upgrade to PROFESSIONAL plan (allows 10 modules)",
          "upgrade_url": "/api/v1/subscriptions/plans/professional"
        }
      },
      {
        "field": "initial_setup.settings.fiscal_year_start",
        "location": "body",
        "code": "INVALID_DATE_FORMAT",
        "message": "Fiscal year start must be a valid ISO 8601 date",
        "severity": "ERROR",
        "current_value": "01/01/2024",
        "constraint": {
          "type": "dateFormat",
          "format": "ISO 8601",
          "example": "2024-01-01"
        }
      }
    ],
    
    "summary": {
      "total_errors": 5,
      "total_warnings": 1,
      "fields_with_errors": [
        "organization.name",
        "organization.slug",
        "organization.contact.email",
        "initial_setup.modules.enabled",
        "initial_setup.settings.fiscal_year_start"
      ],
      "fields_with_warnings": [
        "organization.contact.phone"
      ]
    },
    
    "request_id": "req_9Lp2Km8Qx5Rt3Yn",
    "timestamp": "2024-02-09T11:30:00.000Z"
  }
}
```

---

##  Advanced Response Patterns

### Batch Operation Results

```json
HTTP/1.1 207 Multi-Status
Content-Type: application/json

{
  "success": true,
  "data": {
    "batch_id": "batch_3Kp9Lm2Qx7Rt5Yn",
    "operation_type": "UPDATE_SUBSCRIPTION_PLAN",
    "execution_mode": "ATOMIC",
    "status": "PARTIAL_SUCCESS",
    
    "summary": {
      "total_operations": 5,
      "successful": 3,
      "failed": 2,
      "skipped": 0,
      "success_rate": 60
    },
    
    "results": [
      {
        "organization_id": "org_abc123",
        "operation": "update_plan",
        "status": "SUCCESS",
        "http_status": 200,
        "result": {
          "plan": "PROFESSIONAL",
          "effective_date": "2024-02-09T00:00:00Z",
          "prorated_amount": 15000,
          "next_billing_date": "2024-03-09T00:00:00Z"
        },
        "duration_ms": 234
      },
      {
        "organization_id": "org_xyz789",
        "operation": "update_plan",
        "status": "SUCCESS",
        "http_status": 200,
        "result": {
          "plan": "ENTERPRISE",
          "effective_date": "2024-02-09T00:00:00Z",
          "prorated_amount": 45000,
          "next_billing_date": "2024-03-09T00:00:00Z"
        },
        "duration_ms": 198
      },
      {
        "organization_id": "org_def456",
        "operation": "update_plan",
        "status": "FAILED",
        "http_status": 403,
        "error": {
          "code": "ERR_PAYMENT_METHOD_REQUIRED",
          "message": "Cannot upgrade plan without valid payment method on file",
          "resolution": "Add payment method before upgrading"
        },
        "duration_ms": 45
      },
      {
        "organization_id": "org_ghi789",
        "operation": "update_plan",
        "status": "SUCCESS",
        "http_status": 200,
        "result": {
          "plan": "PROFESSIONAL",
          "effective_date": "2024-02-09T00:00:00Z"
        },
        "duration_ms": 156
      },
      {
        "organization_id": "org_jkl012",
        "operation": "update_plan",
        "status": "FAILED",
        "http_status": 422,
        "error": {
          "code": "ERR_DOWNGRADE_NOT_ALLOWED",
          "message": "Cannot downgrade from ENTERPRISE to PROFESSIONAL while using enterprise-only features",
          "details": {
            "enterprise_features_in_use": [
              "ADVANCED_REPORTING",
              "SSO_INTEGRATION",
              "DEDICATED_SUPPORT"
            ]
          },
          "resolution": "Disable enterprise features before downgrading"
        },
        "duration_ms": 67
      }
    ],
    
    "execution": {
      "started_at": "2024-02-09T11:35:00.000Z",
      "completed_at": "2024-02-09T11:35:01.200Z",
      "total_duration_ms": 1200,
      "parallel_execution": true,
      "max_concurrency": 5
    },
    
    "rollback": {
      "required": false,
      "reason": "Execution mode was not ATOMIC or failures were non-critical"
    }
  },
  
  "meta": {
    "request_id": "req_8Kp3Nm9Lq2Rx7Yt",
    "timestamp": "2024-02-09T11:35:01.200Z"
  }
}
```

---

This comprehensive API design provides:

**Business-friendly naming** (organizations, workspaces, onboarding)  
**Rich request context** via headers (tracing, caching, idempotency)  
**Detailed success responses** with metadata, links, next steps  
**Actionable error messages** with resolution steps  
**Performance optimization** through field selection, caching, batching  
**Developer experience** with comprehensive documentation in responses



## Security & Compliance

###  Authentication Flows

#### 1. OAuth 2.0 Token Exchange

```typescript
// ============================================
// TOKEN ACQUISITION (Authorization Code Flow)
// ============================================

// Step 1: Redirect to authorization endpoint
GET https://auth.awoerp.com/oauth/authorize
Query Parameters:
  ?response_type=code
  &client_id=app_9Km2nP4qL8xR3vZ
  &redirect_uri=https://yourapp.com/callback
  &scope=organizations:read organizations:write analytics:read
  &state=randomly_generated_state_string
  &code_challenge=base64url_encoded_challenge        // PKCE for security
  &code_challenge_method=S256

// Step 2: Exchange code for tokens
POST https://auth.awoerp.com/oauth/token
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code
&code=auth_code_from_step_1
&redirect_uri=https://yourapp.com/callback
&client_id=app_9Km2nP4qL8xR3vZ
&client_secret=secret_abc123                         // Or PKCE verifier
&code_verifier=original_code_verifier                // PKCE

Response: 200 OK
{
  "access_token": "at_9Km2nP4qL8xR3vZ1234567890abcdef",
  "token_type": "Bearer",
  "expires_in": 3600,                                 // 1 hour
  "refresh_token": "rt_3Km8nP2qL9vR4xZ0987654321zyxwv",
  "scope": "organizations:read organizations:write analytics:read",
  "organization_id": "org_2kF8w9mNpL3qR7vX",        // Default organization
  
  // Token metadata
  "issued_at": "2024-02-09T11:30:00Z",
  "refresh_token_expires_in": 2592000,               // 30 days
  
  // User context
  "user": {
    "id": "usr_9Km2nP4qL8xR3vZ",
    "email": "james@nyerifarms.co.ke",
    "name": "James Mwangi Kamau"
  }
}

// ============================================
// TOKEN REFRESH
// ============================================
POST https://auth.awoerp.com/oauth/token
Content-Type: application/x-www-form-urlencoded

grant_type=refresh_token
&refresh_token=rt_3Km8nP2qL9vR4xZ0987654321zyxwv
&client_id=app_9Km2nP4qL8xR3vZ
&client_secret=secret_abc123                         // Server-to-server only
&scope=organizations:read organizations:write        // Optional: reduce scope

Response: 200 OK
{
  "access_token": "at_NEW_TOKEN_HERE",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "rt_NEW_REFRESH_TOKEN",           // Rotating refresh tokens
  "scope": "organizations:read organizations:write"
}

// ============================================
// TOKEN INTROSPECTION (Validate Token)
// ============================================
POST https://auth.awoerp.com/oauth/introspect
Authorization: Basic base64(client_id:client_secret)
Content-Type: application/x-www-form-urlencoded

token=at_9Km2nP4qL8xR3vZ1234567890abcdef
&token_type_hint=access_token

Response: 200 OK
{
  "active": true,
  "scope": "organizations:read organizations:write analytics:read",
  "client_id": "app_9Km2nP4qL8xR3vZ",
  "username": "james@nyerifarms.co.ke",
  "token_type": "Bearer",
  "exp": 1707483600,                                 // Unix timestamp
  "iat": 1707480000,
  "sub": "usr_9Km2nP4qL8xR3vZ",                     // Subject (user ID)
  "aud": "https://api.awoerp.com",                   // Audience
  "iss": "https://auth.awoerp.com",                  // Issuer
  
  // Custom claims
  "organization_id": "org_2kF8w9mNpL3qR7vX",
  "role": "ADMIN",
  "permissions": [
    "organizations:read",
    "organizations:write",
    "analytics:read"
  ]
}

// ============================================
// TOKEN REVOCATION
// ============================================
POST https://auth.awoerp.com/oauth/revoke
Authorization: Basic base64(client_id:client_secret)
Content-Type: application/x-www-form-urlencoded

token=rt_3Km8nP2qL9vR4xZ0987654321zyxwv
&token_type_hint=refresh_token

Response: 200 OK
{
  "revoked": true,
  "revoked_at": "2024-02-09T11:35:00Z"
}
```

#### 2. API Key Authentication (Server-to-Server)

```typescript
// ============================================
// CREATE API KEY
// ============================================
POST /api/v1/orgs/:id/api-keys
Authorization: Bearer <user_access_token>

Body: {
  "name": "Production Integration - Accounting System",
  "description": "Integration with QuickBooks for invoice sync",
  "scopes": [
    "invoices:read",
    "invoices:write",
    "customers:read"
  ],
  "expires_at": "2025-02-09T00:00:00Z",              // Optional: expiry date
  "ip_whitelist": [                                   // Optional: IP restrictions
    "41.90.64.0/24",
    "197.232.0.0/16"
  ],
  "rate_limit_tier": "STANDARD",                      // STANDARD | HIGH | UNLIMITED
  "metadata": {
    "environment": "production",
    "system": "quickbooks",
    "owner": "finance-team"
  }
}

Response: 201 Created
{
  "data": {
    "id": "key_5Km8nP2qL9vR4xZ",
    "name": "Production Integration - Accounting System",
    "key_prefix": "live_sk_",
    "api_key": "live_sk_9Km2nP4qL8xR3vZ1234567890abcdef",  // ONLY SHOWN ONCE
    "scopes": ["invoices:read", "invoices:write", "customers:read"],
    "created_at": "2024-02-09T11:30:00Z",
    "expires_at": "2025-02-09T00:00:00Z",
    "last_used_at": null,
    "usage_count": 0
  },
  
  "meta": {
    "warning": "This API key will only be displayed once. Store it securely.",
    "security_tips": [
      "Never commit API keys to version control",
      "Rotate keys regularly (recommended: every 90 days)",
      "Use environment variables to store keys",
      "Monitor key usage in the dashboard"
    ]
  }
}

// ============================================
// USE API KEY
// ============================================
GET /api/v1/orgs/org_2kF8w9mNpL3qR7vX/analytics/usage
Authorization: Bearer live_sk_9Km2nP4qL8xR3vZ1234567890abcdef
X-Organization-ID: org_2kF8w9mNpL3qR7vX              // Required for API keys

Response: 200 OK
Headers:
  X-API-Key-ID: key_5Km8nP2qL9vR4xZ
  X-API-Key-Scopes: invoices:read,invoices:write,customers:read
  X-RateLimit-Tier: STANDARD

// ============================================
// LIST API KEYS
// ============================================
GET /api/v1/orgs/:id/api-keys
Query: ?status=ACTIVE&sort=-created_at

Response: 200 OK
{
  "data": [
    {
      "id": "key_5Km8nP2qL9vR4xZ",
      "name": "Production Integration - Accounting System",
      "key_prefix": "live_sk_****",                   // Masked for security
      "scopes": ["invoices:read", "invoices:write"],
      "status": "ACTIVE",                             // ACTIVE | EXPIRED | REVOKED
      "created_at": "2024-02-09T11:30:00Z",
      "last_used_at": "2024-02-09T11:45:00Z",
      "usage_count": 1247,
      "expires_at": "2025-02-09T00:00:00Z"
    }
  ]
}

// ============================================
// REVOKE API KEY
// ============================================
DELETE /api/v1/orgs/:id/api-keys/:key_id
Response: 204 No Content

// ============================================
// ROTATE API KEY
// ============================================
POST /api/v1/orgs/:id/api-keys/:key_id/rotate
Body: {
  "grace_period_hours": 24                            // Old key valid for 24h
}

Response: 200 OK
{
  "data": {
    "new_key": {
      "id": "key_NEW_ID",
      "api_key": "live_sk_NEW_KEY_HERE",              // New key (shown once)
      "scopes": ["invoices:read", "invoices:write"],
      "created_at": "2024-02-09T12:00:00Z"
    },
    "old_key": {
      "id": "key_5Km8nP2qL9vR4xZ",
      "status": "GRACE_PERIOD",
      "expires_at": "2024-02-10T12:00:00Z"            // 24 hours from now
    }
  },
  
  "meta": {
    "grace_period_ends_at": "2024-02-10T12:00:00Z",
    "migration_instructions": "Update your integration to use the new key before the grace period ends."
  }
}
```

### ️ Permission System (RBAC/ABAC)

```typescript
// ============================================
// CHECK PERMISSIONS
// ============================================
POST /api/v1/permissions/check
Authorization: Bearer <access_token>

Body: {
  "user_id": "usr_9Km2nP4qL8xR3vZ",
  "organization_id": "org_2kF8w9mNpL3qR7vX",
  "permissions": [
    {
      "resource": "organizations",
      "action": "update",
      "context": {
        "organization_id": "org_2kF8w9mNpL3qR7vX"
      }
    },
    {
      "resource": "invoices",
      "action": "approve",
      "context": {
        "organization_id": "org_2kF8w9mNpL3qR7vX",
        "invoice_amount_ksh": 500000
      }
    }
  ]
}

Response: 200 OK
{
  "data": {
    "user": {
      "id": "usr_9Km2nP4qL8xR3vZ",
      "role": "FINANCE_MANAGER",
      "department": "FINANCE"
    },
    "results": [
      {
        "resource": "organizations",
        "action": "update",
        "allowed": true,
        "reason": "User has role FINANCE_MANAGER with permission organizations:update"
      },
      {
        "resource": "invoices",
        "action": "approve",
        "allowed": false,
        "reason": "Invoice amount (500,000 KES) exceeds user approval limit (100,000 KES)",
        "required_role": "FINANCE_DIRECTOR",
        "approval_workflow": {
          "required": true,
          "approvers": [
            {
              "id": "usr_DIRECTOR_ID",
              "name": "Sarah Wanjiku",
              "role": "FINANCE_DIRECTOR"
            }
          ]
        }
      }
    ]
  }
}

// ============================================
// GET USER PERMISSIONS
// ============================================
GET /api/v1/users/:user_id/permissions
Query: ?organization_id=org_2kF8w9mNpL3qR7vX&include_inherited=true

Response: 200 OK
{
  "data": {
    "user": {
      "id": "usr_9Km2nP4qL8xR3vZ",
      "role": "FINANCE_MANAGER"
    },
    "permissions": {
      "direct": [
        "invoices:read",
        "invoices:create",
        "invoices:update",
        "customers:read",
        "reports:financial:read"
      ],
      "inherited_from_role": [
        "organizations:read",
        "users:read",
        "analytics:basic:read"
      ],
      "inherited_from_groups": [
        "accounting:reconciliation:execute"
      ]
    },
    "limits": {
      "invoice_approval_limit_ksh": 100000,
      "payment_approval_limit_ksh": 50000,
      "discount_percentage_max": 10
    },
    "restrictions": {
      "can_delete_transactions": false,
      "can_modify_closed_periods": false,
      "requires_mfa_for": ["payments:approve", "users:delete"]
    }
  }
}
```

###  Data Encryption & Privacy

```typescript
// ============================================
// FIELD-LEVEL ENCRYPTION (Transparent to API)
// ============================================

// Request: Sensitive fields automatically encrypted
POST /api/v1/orgs/:id/bank-accounts
Body: {
  "bank_name": "Equity Bank",
  "account_number": "0123456789",                     // Encrypted at rest
  "account_name": "Nyeri Farms Ltd",
  "branch": "Nyeri Branch",
  "swift_code": "EQBLKENA",
  
  // PII (Personally Identifiable Information)
  "contact_person": {
    "name": "James Mwangi",                           // Encrypted
    "phone": "+254712345678",                         // Encrypted
    "email": "james@nyerifarms.co.ke",               // Encrypted
    "id_number": "12345678"                           // Encrypted + hashed
  }
}

Response: 200 OK
{
  "data": {
    "id": "bank_5Km8nP2qL9vR4xZ",
    "bank_name": "Equity Bank",
    "account_number": "****6789",                     // Masked in response
    "account_name": "Nyeri Farms Ltd",
    "branch": "Nyeri Branch",
    "swift_code": "EQBLKENA",
    "contact_person": {
      "name": "James M****",                          // Masked
      "phone": "+254****5678",                        // Masked
      "email": "j****@nyerifarms.co.ke",             // Masked
      "id_number": "****5678"                         // Masked
    },
    "encryption": {
      "algorithm": "AES-256-GCM",
      "key_version": "v2024.02",
      "encrypted_fields": [
        "account_number",
        "contact_person.name",
        "contact_person.phone",
        "contact_person.email",
        "contact_person.id_number"
      ]
    }
  }
}

// ============================================
// DATA EXPORT (GDPR Compliance)
// ============================================
POST /api/v1/orgs/:id/data-export/request
Body: {
  "export_type": "GDPR_SUBJECT_ACCESS",               // GDPR | FULL_BACKUP
  "data_categories": [
    "PROFILE_DATA",
    "TRANSACTION_HISTORY",
    "COMMUNICATION_LOGS",
    "AUDIT_TRAIL"
  ],
  "format": "JSON",                                   // JSON | CSV | PDF
  "include_attachments": true,
  "encryption": {
    "enabled": true,
    "password_protected": true                        // User sets password
  },
  "delivery_method": "EMAIL",                         // EMAIL | DOWNLOAD_LINK
  "legal_basis": "GDPR Article 15 - Right of Access"
}

Response: 202 Accepted
{
  "job_id": "export_9Km2nP4qL9vR4xZ",
  "estimated_completion_minutes": 15,
  "status_url": "/api/v1/jobs/export_9Km2nP4qL9vR4xZ"
}

// When complete:
GET /api/v1/jobs/export_9Km2nP4qL9vR4xZ
Response: 200 OK
{
  "data": {
    "job_id": "export_9Km2nP4qL9vR4xZ",
    "status": "COMPLETED",
    "download_url": "https://exports.awoerp.com/secure/export_9Km2nP4qL9vR4xZ.zip",
    "expires_at": "2024-02-16T11:30:00Z",             // 7 days
    "file_size_bytes": 15728640,
    "encryption": {
      "enabled": true,
      "password_hint": "Your organization registration year + initials"
    },
    "manifest": {
      "total_records": 12456,
      "files_included": [
        "profile.json",
        "transactions.csv",
        "communications.json",
        "audit_trail.csv",
        "attachments/..."
      ]
    }
  }
}

// ============================================
// DATA DELETION (Right to be Forgotten)
// ============================================
POST /api/v1/orgs/:id/data-deletion/request
Body: {
  "deletion_type": "GDPR_RIGHT_TO_ERASURE",
  "reason": "User requested account deletion",
  "retain_legal_data": true,                          // Keep tax/legal records
  "retention_period_days": 2555,                      // 7 years for tax
  "anonymize_instead": false,                         // true = anonymize, false = delete
  "confirm": "PERMANENTLY_DELETE_ALL_DATA"
}

Response: 202 Accepted
{
  "job_id": "deletion_3Km8nP2qL9vR4xZ",
  "grace_period_days": 30,
  "deletion_scheduled_for": "2024-03-10T11:30:00Z",
  "cancellation_url": "/api/v1/jobs/deletion_3Km8nP2qL9vR4xZ/cancel",
  
  "retention_policy": {
    "what_will_be_deleted": [
      "Personal information",
      "Transaction history (except legal retention)",
      "Communication logs",
      "Uploaded files and attachments"
    ],
    "what_will_be_retained": [
      "Invoice records (tax law requirement - 7 years)",
      "Audit logs (anonymized)",
      "Aggregated analytics (anonymized)"
    ]
  }
}
```

###  Data Residency & Compliance

```typescript
// ============================================
// SET DATA RESIDENCY PREFERENCE
// ============================================
PATCH /api/v1/orgs/:id/compliance/data-residency
Body: {
  "primary_region": "KENYA_NAIROBI",                  // Data center location
  "backup_regions": ["KENYA_MOMBASA"],
  "allow_cross_border_transfers": false,              // GDPR/DPA compliance
  "approved_countries": [],                           // If transfers allowed
  "compliance_frameworks": [
    "KENYA_DPA_2019",                                 // Kenya Data Protection Act
    "GDPR",                                           // EU GDPR
    "ISO_27001"
  ]
}

Response: 200 OK
{
  "data": {
    "data_residency": {
      "primary_region": "KENYA_NAIROBI",
      "data_centers": [
        {
          "location": "Nairobi, Kenya",
          "provider": "AWS Africa (Cape Town) - Kenya Edge",
          "certifications": ["ISO 27001", "SOC 2 Type II"]
        }
      ],
      "cross_border_transfers": {
        "allowed": false,
        "approved_countries": []
      },
      "migration_status": "COMPLETED",
      "migrated_at": "2024-02-09T12:00:00Z"
    }
  }
}

// ============================================
// COMPLIANCE REPORT
// ============================================
GET /api/v1/orgs/:id/compliance/report
Query: ?frameworks=KENYA_DPA_2019,GDPR&format=PDF

Response: 200 OK
{
  "data": {
    "organization_id": "org_2kF8w9mNpL3qR7vX",
    "report_date": "2024-02-09T11:30:00Z",
    "compliance_status": {
      "KENYA_DPA_2019": {
        "compliant": true,
        "score": 98,
        "last_audit": "2024-01-15T00:00:00Z",
        "next_audit": "2024-07-15T00:00:00Z",
        "issues": []
      },
      "GDPR": {
        "compliant": true,
        "score": 96,
        "last_audit": "2024-01-20T00:00:00Z",
        "next_audit": "2024-07-20T00:00:00Z",
        "issues": [
          {
            "severity": "LOW",
            "issue": "Cookie consent banner could be more prominent",
            "remediation": "Update UI to make consent more visible"
          }
        ]
      }
    },
    "data_processing_activities": {
      "total": 47,
      "documented": 47,
      "compliance_rate": 100
    },
    "security_measures": {
      "encryption_at_rest": true,
      "encryption_in_transit": true,
      "access_controls": true,
      "audit_logging": true,
      "data_backup": true,
      "disaster_recovery": true
    },
    "report_url": "https://reports.awoerp.com/compliance/org_2kF8w9mNpL3qR7vX_2024-02-09.pdf"
  }
}
```

---

## WebSocket & Real-time Updates

###  WebSocket Connection

```typescript
// ============================================
// ESTABLISH WEBSOCKET CONNECTION
// ============================================
const ws = new WebSocket('wss://ws.awoerp.com/v1/realtime');

// Connection with authentication
ws.addEventListener('open', () => {
  ws.send(JSON.stringify({
    type: 'authenticate',
    token: 'at_9Km2nP4qL8xR3vZ1234567890abcdef',
    organization_id: 'org_2kF8w9mNpL3qR7vX'
  }));
});

// Authentication response
Server → Client:
{
  "type": "authenticated",
  "user_id": "usr_9Km2nP4qL8xR3vZ",
  "organization_id": "org_2kF8w9mNpL3qR7vX",
  "session_id": "ws_session_abc123",
  "server_time": "2024-02-09T11:30:00Z",
  "capabilities": [
    "organization_updates",
    "user_activity",
    "analytics_realtime",
    "notifications"
  ]
}

// ============================================
// SUBSCRIBE TO CHANNELS
// ============================================
Client → Server:
{
  "type": "subscribe",
  "subscriptions": [
    {
      "channel": "organization:org_2kF8w9mNpL3qR7vX",
      "events": [
        "status_changed",
        "settings_updated",
        "subscription_changed"
      ]
    },
    {
      "channel": "analytics:org_2kF8w9mNpL3qR7vX",
      "events": [
        "usage_updated",
        "threshold_exceeded"
      ],
      "filters": {
        "metrics": ["active_users", "storage_used"]
      }
    },
    {
      "channel": "notifications:usr_9Km2nP4qL8xR3vZ",
      "events": ["*"]                                  // All notification events
    }
  ]
}

Server → Client:
{
  "type": "subscribed",
  "subscriptions": [
    {
      "channel": "organization:org_2kF8w9mNpL3qR7vX",
      "status": "active",
      "subscribed_at": "2024-02-09T11:30:01Z"
    },
    {
      "channel": "analytics:org_2kF8w9mNpL3qR7vX",
      "status": "active",
      "subscribed_at": "2024-02-09T11:30:01Z"
    },
    {
      "channel": "notifications:usr_9Km2nP4qL8xR3vZ",
      "status": "active",
      "subscribed_at": "2024-02-09T11:30:01Z"
    }
  ]
}

// ============================================
// RECEIVE REAL-TIME EVENTS
// ============================================

// Organization status changed
Server → Client:
{
  "type": "event",
  "channel": "organization:org_2kF8w9mNpL3qR7vX",
  "event": "status_changed",
  "data": {
    "organization_id": "org_2kF8w9mNpL3qR7vX",
    "old_status": "ACTIVE",
    "new_status": "PAUSED",
    "reason": "Payment overdue - 30 days",
    "changed_at": "2024-02-09T11:35:00Z",
    "changed_by": {
      "type": "SYSTEM",
      "automated_action": "PAYMENT_OVERDUE_SUSPENSION"
    },
    "effective_immediately": true,
    "restrictions": {
      "read_only": true,
      "api_access": "LIMITED",
      "user_logins": "ADMIN_ONLY"
    }
  },
  "timestamp": "2024-02-09T11:35:00.123Z",
  "event_id": "evt_5Km8nP2qL9vR4xZ"
}

// Analytics threshold exceeded
Server → Client:
{
  "type": "event",
  "channel": "analytics:org_2kF8w9mNpL3qR7vX",
  "event": "threshold_exceeded",
  "data": {
    "metric": "storage_used",
    "current_value": 9216,                            // MB
    "threshold": 9011,                                 // 90% of 10240 MB
    "quota": 10240,
    "usage_percentage": 90,
    "severity": "WARNING",
    "recommended_action": "UPGRADE_PLAN",
    "grace_period_days": 7,
    "consequences": "Storage uploads will be blocked after grace period"
  },
  "timestamp": "2024-02-09T11:40:00.456Z",
  "event_id": "evt_8Lp3mK9qN2rP4vZ"
}

// User activity event
Server → Client:
{
  "type": "event",
  "channel": "organization:org_2kF8w9mNpL3qR7vX",
  "event": "user_logged_in",
  "data": {
    "user": {
      "id": "usr_NEW_USER",
      "name": "Sarah Wanjiru",
      "email": "sarah@nyerifarms.co.ke"
    },
    "login_time": "2024-02-09T11:42:00Z",
    "ip_address": "41.90.64.123",
    "location": {
      "city": "Nairobi",
      "country": "Kenya"
    },
    "device": {
      "type": "DESKTOP",
      "os": "Windows 11",
      "browser": "Chrome 121"
    }
  },
  "timestamp": "2024-02-09T11:42:00.789Z",
  "event_id": "evt_2kF8w9mNpL3qR7vX"
}

// ============================================
// HEARTBEAT (Keep-Alive)
// ============================================
Server → Client (every 30 seconds):
{
  "type": "ping",
  "timestamp": "2024-02-09T11:30:00Z"
}

Client → Server:
{
  "type": "pong",
  "timestamp": "2024-02-09T11:30:00Z"
}

// ============================================
// UNSUBSCRIBE
// ============================================
Client → Server:
{
  "type": "unsubscribe",
  "channels": [
    "analytics:org_2kF8w9mNpL3qR7vX"
  ]
}

Server → Client:
{
  "type": "unsubscribed",
  "channels": [
    "analytics:org_2kF8w9mNpL3qR7vX"
  ],
  "unsubscribed_at": "2024-02-09T11:45:00Z"
}

// ============================================
// ERROR HANDLING
// ============================================
Server → Client:
{
  "type": "error",
  "error": {
    "code": "SUBSCRIPTION_LIMIT_EXCEEDED",
    "message": "Maximum 10 active subscriptions allowed per connection",
    "current_subscriptions": 10,
    "max_subscriptions": 10,
    "requested_channel": "invoices:org_2kF8w9mNpL3qR7vX"
  },
  "timestamp": "2024-02-09T11:46:00Z"
}

// Connection closed
Server → Client:
{
  "type": "close",
  "reason": "TOKEN_EXPIRED",
  "message": "Your authentication token has expired. Please reconnect with a valid token.",
  "reconnect_allowed": true,
  "retry_after_seconds": 5
}
```

###  Server-Sent Events (SSE) Alternative

```typescript
// ============================================
// SSE CONNECTION (Simpler than WebSocket)
// ============================================
const eventSource = new EventSource(
  'https://api.awoerp.com/v1/events/stream?' +
  'token=at_9Km2nP4qL8xR3vZ1234567890abcdef&' +
  'org=org_2kF8w9mNpL3qR7vX&' +
  'channels=organization,analytics,notifications'
);

// Listen to specific event types
eventSource.addEventListener('status_changed', (event) => {
  const data = JSON.parse(event.data);
  console.log('Organization status changed:', data);
});

eventSource.addEventListener('threshold_exceeded', (event) => {
  const data = JSON.parse(event.data);
  console.log('Threshold exceeded:', data);
});

// Generic message handler
eventSource.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Event received:', data);
};

// Error handling
eventSource.onerror = (error) => {
  console.error('SSE error:', error);
  // Auto-reconnects by default
};

// Server sends events
HTTP/1.1 200 OK
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive

event: status_changed
id: evt_5Km8nP2qL9vR4xZ
data: {"organization_id":"org_2kF8w9mNpL3qR7vX","old_status":"ACTIVE","new_status":"PAUSED"}

event: threshold_exceeded
id: evt_8Lp3mK9qN2rP4vZ
data: {"metric":"storage_used","current_value":9216,"threshold":9011}

event: heartbeat
data: {"timestamp":"2024-02-09T11:30:00Z"}
```

---

## Webhooks

###  Webhook Configuration

```typescript
// ============================================
// CREATE WEBHOOK
// ============================================
POST /api/v1/orgs/:id/webhooks
Body: {
  "url": "https://yourapp.com/webhooks/awo-erp",
  "description": "Receive organization status changes",
  "events": [
    "organization.status_changed",
    "organization.subscription_changed",
    "organization.settings_updated",
    "analytics.threshold_exceeded",
    "billing.invoice_created",
    "billing.payment_received"
  ],
  "filters": {
    "organization_ids": ["org_2kF8w9mNpL3qR7vX"],    // Optional: filter by org
    "severity": ["WARNING", "CRITICAL"]                // Optional: filter by severity
  },
  "secret": "whsec_a1b2c3d4e5f6g7h8i9j0",            // For HMAC verification
  "retry_policy": {
    "max_attempts": 3,
    "backoff_strategy": "EXPONENTIAL",                // LINEAR | EXPONENTIAL
    "initial_interval_seconds": 60
  },
  "timeout_seconds": 30,
  "headers": {                                        // Custom headers
    "X-Custom-Header": "value",
    "X-Environment": "production"
  },
  "active": true
}

Response: 201 Created
{
  "data": {
    "id": "webhook_5Km8nP2qL9vR4xZ",
    "url": "https://yourapp.com/webhooks/awo-erp",
    "events": ["organization.status_changed", "..."],
    "secret": "whsec_****j0",                         // Masked
    "status": "ACTIVE",
    "created_at": "2024-02-09T11:30:00Z",
    "last_triggered_at": null,
    "delivery_success_rate": null,
    
    // Test webhook
    "test_url": "/api/v1/webhooks/webhook_5Km8nP2qL9vR4xZ/test"
  }
}

// ============================================
// WEBHOOK PAYLOAD FORMAT
// ============================================
POST https://yourapp.com/webhooks/awo-erp
Headers:
  Content-Type: application/json
  X-AWO-Event: organization.status_changed
  X-AWO-Signature: sha256=abc123...                   // HMAC signature
  X-AWO-Delivery-ID: delivery_9Km2nP4qL8xR3vZ
  X-AWO-Webhook-ID: webhook_5Km8nP2qL9vR4xZ
  X-AWO-Timestamp: 1707480000
  User-Agent: AWO-ERP-Webhooks/1.0

Body:
{
  "id": "evt_5Km8nP2qL9vR4xZ",
  "type": "organization.status_changed",
  "api_version": "2024-02-01",
  "created_at": "2024-02-09T11:35:00Z",
  
  "data": {
    "organization_id": "org_2kF8w9mNpL3qR7vX",
    "old_status": "ACTIVE",
    "new_status": "PAUSED",
    "reason": "Payment overdue - 30 days",
    "changed_at": "2024-02-09T11:35:00Z",
    "changed_by": {
      "type": "SYSTEM",
      "automated_action": "PAYMENT_OVERDUE_SUSPENSION"
    }
  },
  
  "organization": {
    "id": "org_2kF8w9mNpL3qR7vX",
    "business": {
      "legal_name": "Nyeri Agribusiness Enterprises Limited",
      "trading_name": "Nyeri Farms"
    }
  },
  
  "metadata": {
    "webhook_id": "webhook_5Km8nP2qL9vR4xZ",
    "delivery_attempt": 1,
    "delivery_id": "delivery_9Km2nP4qL8xR3vZ"
  }
}

// Your endpoint should respond:
HTTP/1.1 200 OK
{
  "received": true
}

// ============================================
// VERIFY WEBHOOK SIGNATURE
// ============================================
// Go implementation
func VerifyWebhookSignature(payload []byte, signature, secret string) bool {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(payload)
    expectedMAC := mac.Sum(nil)
    expectedSignature := "sha256=" + hex.EncodeToString(expectedMAC)
    return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

// TypeScript implementation
import crypto from 'crypto';

function verifyWebhookSignature(
  payload: string,
  signature: string,
  secret: string
): boolean {
  const hmac = crypto.createHmac('sha256', secret);
  hmac.update(payload);
  const expectedSignature = 'sha256=' + hmac.digest('hex');
  return crypto.timingSafeEqual(
    Buffer.from(expectedSignature),
    Buffer.from(signature)
  );
}

// ============================================
// TEST WEBHOOK
// ============================================
POST /api/v1/webhooks/:webhook_id/test
Body: {
  "event_type": "organization.status_changed"         // Optional: specific event
}

Response: 200 OK
{
  "data": {
    "test_sent": true,
    "delivery_id": "delivery_TEST_123",
    "response": {
      "status_code": 200,
      "response_time_ms": 145,
      "body": "{\"received\":true}"
    }
  }
}

// ============================================
// GET WEBHOOK DELIVERIES
// ============================================
GET /api/v1/webhooks/:webhook_id/deliveries
Query: ?status=FAILED&from=2024-02-01&limit=50

Response: 200 OK
{
  "data": [
    {
      "id": "delivery_9Km2nP4qL8xR3vZ",
      "event_id": "evt_5Km8nP2qL9vR4xZ",
      "event_type": "organization.status_changed",
      "attempt": 1,
      "status": "SUCCESS",
      "delivered_at": "2024-02-09T11:35:01Z",
      "response": {
        "status_code": 200,
        "response_time_ms": 145
      }
    },
    {
      "id": "delivery_8Lp3mK9qN2rP4vZ",
      "event_id": "evt_2kF8w9mNpL3qR7vX",
      "event_type": "billing.payment_received",
      "attempt": 3,
      "status": "FAILED",
      "delivered_at": "2024-02-09T11:40:05Z",
      "response": {
        "status_code": 500,
        "response_time_ms": 30000,
        "error": "Connection timeout"
      },
      "next_retry_at": "2024-02-09T11:50:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 50,
    "total": 247
  }
}

// ============================================
// RETRY FAILED DELIVERY
// ============================================
POST /api/v1/webhooks/:webhook_id/deliveries/:delivery_id/retry

Response: 200 OK
{
  "data": {
    "delivery_id": "delivery_8Lp3mK9qN2rP4vZ",
    "retry_scheduled": true,
    "retry_at": "2024-02-09T11:45:00Z"
  }
}

// ============================================
// WEBHOOK EVENTS CATALOG
// ============================================
GET /api/v1/webhooks/events

Response: 200 OK
{
  "data": {
    "categories": [
      {
        "category": "organization",
        "events": [
          {
            "type": "organization.created",
            "description": "New organization was created",
            "example_payload": { ... }
          },
          {
            "type": "organization.status_changed",
            "description": "Organization status changed",
            "example_payload": { ... }
          },
          {
            "type": "organization.subscription_changed",
            "description": "Subscription plan changed",
            "example_payload": { ... }
          }
        ]
      },
      {
        "category": "billing",
        "events": [
          {
            "type": "billing.invoice_created",
            "description": "New invoice was generated",
            "example_payload": { ... }
          },
          {
            "type": "billing.payment_received",
            "description": "Payment was successfully processed",
            "example_payload": { ... }
          }
        ]
      }
    ]
  }
}
```

---

## Advanced Features

###  Multi-Currency Support

```typescript
// ============================================
// ENABLE MULTI-CURRENCY
// ============================================
POST /api/v1/orgs/:id/features/multi-currency/enable
Body: {
  "base_currency": "KES",                             // Primary currency
  "enabled_currencies": [
    "USD",
    "EUR",
    "GBP",
    "UGX",
    "TZS"
  ],
  "auto_update_rates": true,                          // Fetch daily rates
  "rate_provider": "CENTRAL_BANK_OF_KENYA",          // CBK | ECB | MANUAL
  "rounding_method": "HALF_UP",                       // Banking rounding
  "decimal_places": 2
}

Response: 200 OK
{
  "data": {
    "multi_currency_enabled": true,
    "base_currency": "KES",
    "enabled_currencies": ["USD", "EUR", "GBP", "UGX", "TZS"],
    "exchange_rates": {
      "USD": {
        "rate": 150.50,
        "last_updated": "2024-02-09T09:00:00Z",
        "source": "CENTRAL_BANK_OF_KENYA"
      },
      "EUR": {
        "rate": 165.75,
        "last_updated": "2024-02-09T09:00:00Z",
        "source": "CENTRAL_BANK_OF_KENYA"
      }
    }
  }
}

// ============================================
// CONVERT AMOUNT
// ============================================
GET /api/v1/currency/convert
Query: ?from=USD&to=KES&amount=100&date=2024-02-09

Response: 200 OK
{
  "data": {
    "from": {
      "currency": "USD",
      "amount": 100.00
    },
    "to": {
      "currency": "KES",
      "amount": 15050.00
    },
    "exchange_rate": 150.50,
    "rate_date": "2024-02-09T09:00:00Z",
    "source": "CENTRAL_BANK_OF_KENYA",
    "calculation": "100 USD × 150.50 = 15,050.00 KES"
  }
}

// ============================================
// TRANSACTION WITH MULTI-CURRENCY
// ============================================
POST /api/v1/orgs/:id/transactions
Body: {
  "type": "INVOICE",
  "customer_id": "cust_5Km8nP2qL9vR4xZ",
  "currency": "USD",                                  // Transaction currency
  "amount": 1000.00,
  "items": [
    {
      "description": "Coffee beans - 500kg",
      "quantity": 500,
      "unit_price_usd": 2.00,
      "total_usd": 1000.00
    }
  ],
  
  // Multi-currency details
  "multi_currency": {
    "base_currency_ksh": 150500.00,                   // Auto-converted to base
    "exchange_rate": 150.50,
    "rate_date": "2024-02-09T09:00:00Z",
    "rate_locked": true                               // Lock rate on invoice
  }
}

Response: 201 Created
{
  "data": {
    "id": "inv_9Km2nP4qL8xR3vZ",
    "currency": "USD",
    "amount_usd": 1000.00,
    "amount_ksh": 150500.00,                          // Base currency
    "exchange_rate": 150.50,
    "exchange_rate_locked": true,
    "display_amounts": {
      "usd": "$1,000.00",
      "ksh": "KSh 150,500.00"
    }
  }
}
```

###  Localization (i18n)

```typescript
// ============================================
// SET ORGANIZATION LOCALE
// ============================================
PATCH /api/v1/orgs/:id/settings/localization
Body: {
  "language": "sw-KE",                                // Swahili (Kenya)
  "timezone": "Africa/Nairobi",
  "date_format": "DD/MM/YYYY",
  "time_format": "24H",
  "number_format": {
    "decimal_separator": ".",
    "thousands_separator": ",",
    "decimal_places": 2
  },
  "currency_format": {
    "symbol_position": "PREFIX",                      // PREFIX | SUFFIX
    "symbol": "KSh",
    "format": "KSh #,##0.00"
  },
  "address_format": "KE"                              // Kenya format
}

// ============================================
// GET LOCALIZED RESPONSE
// ============================================
GET /api/v1/orgs/:id
Accept-Language: sw-KE
X-Timezone: Africa/Nairobi

Response: 200 OK
{
  "data": {
    "id": "org_2kF8w9mNpL3qR7vX",
    "status": "ACTIVE",
    "status_display": "Inafanya Kazi",                // Swahili translation
    "created_at": "2024-01-15T08:00:00Z",
    "created_at_display": "15 Januari 2024, 11:00",  // Localized (EAT = UTC+3)
    "subscription": {
      "plan": "PROFESSIONAL",
      "plan_display": "Mtaalamu",
      "amount_ksh": 15000,
      "amount_display": "KSh 15,000.00"
    }
  },
  
  "meta": {
    "locale": "sw-KE",
    "timezone": "Africa/Nairobi",
    "currency": "KES"
  }
}
```

Next: [Troubleshooting Guide](./20-troubleshooting-guide.md)
