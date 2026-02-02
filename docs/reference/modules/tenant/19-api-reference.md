[<-- Back to Index](README.md)

## API Reference

### Tenant CRUD Operations

```markdown
CREATE TENANT:
  POST /api/v1/tenants
  Body: {
    "name": "Nyeri Farms Ltd",
    "email": "admin@nyerifarms.co.ke",
    "subdomain": "nyerifarms",
    "industry": "Agriculture",
    "company_size": "Small",
    "currency_code": "KES",
    "timezone": "Africa/Nairobi",
    "plan_type": "basic"
  }
  Response: 201 { "id": "uuid", "slug": "nyeri-farms-ltd", "status": "PENDING" }

GET TENANT:
  GET /api/v1/tenants/:id
  Response: 200 { full tenant record }

UPDATE TENANT:
  PUT /api/v1/tenants/:id
  Body: { fields to update }
  Response: 200 { updated tenant record }

DELETE TENANT (soft):
  DELETE /api/v1/tenants/:id
  Response: 204

LIST TENANTS:
  GET /api/v1/tenants?status=ACTIVE&page=1&limit=20
  Response: 200 { "data": [...], "total": 150, "page": 1 }
```

### Provisioning API

```markdown
PROVISION COMPLETE:
  POST /api/v1/tenants/provision
  Body: {
    "tenant": {
      "name": "Coastal Coffee Co.",
      "email": "admin@coastalcoffee.co.ke",
      "subdomain": "coastalcoffee",
      "industry": "Agriculture",
      "company_size": "Medium",
      "currency_code": "KES",
      "timezone": "Africa/Nairobi",
      "plan_type": "professional"
    },
    "admin_user": {
      "name": "James Mwangi",
      "email": "james@coastalcoffee.co.ke",
      "password": "SecurePass123!"
    },
    "settings": {
      "fiscal_year_start_month": 1,
      "accounting_method": "FIFO",
      "enabled_modules": ["financial", "selling", "buying", "inventory"]
    }
  }
  Response: 201 {
    "tenant_id": "uuid",
    "slug": "coastal-coffee-co",
    "status": "ACTIVE",
    "admin_user_id": "uuid"
  }
```

### Status Management API

```markdown
SUSPEND TENANT:
  POST /api/v1/tenants/:id/suspend
  Body: {
    "reason": "Payment overdue 30+ days",
    "notify": true
  }
  Response: 200 { "status": "SUSPENDED" }

REACTIVATE TENANT:
  POST /api/v1/tenants/:id/reactivate
  Body: {
    "reason": "Payment received"
  }
  Response: 200 { "status": "ACTIVE" }

ARCHIVE TENANT:
  POST /api/v1/tenants/:id/archive
  Body: {
    "reason": "Business closure",
    "retention_days": 2555
  }
  Response: 200 { "status": "ARCHIVED" }
```

### Bulk Operations API

```markdown
BULK OPERATION:
  POST /api/v1/tenants/bulk
  Body: {
    "operation_type": "SUSPEND",
    "tenant_ids": ["uuid-1", "uuid-2", "uuid-3"],
    "parameters": {
      "reason": "Payment overdue batch - Feb 2024"
    }
  }
  Response: 202 {
    "operation_id": "op-uuid",
    "status": "IN_PROGRESS",
    "total_tenants": 3
  }

GET BULK OPERATION STATUS:
  GET /api/v1/tenants/bulk/:operation_id
  Response: 200 {
    "operation_id": "op-uuid",
    "operation_type": "SUSPEND",
    "status": "COMPLETED",
    "total_tenants": 3,
    "successful_count": 2,
    "failed_count": 0,
    "in_progress_count": 0,
    "duration_seconds": 3
  }
```

### Configuration API

```markdown
GET CONFIGURATION:
  GET /api/v1/tenants/:id/config
  Response: 200 {
    "max_users": 50,
    "max_entities": 20,
    "max_transactions_per_month": 10000,
    "storage_quota_mb": 10240,
    "allowed_modules": ["financial", "selling", "buying", "inventory"],
    "accounting_method": "FIFO",
    "fiscal_year_start_month": 1,
    "password_policy": { ... },
    "api_rate_limits": { ... }
  }

UPDATE CONFIGURATION:
  PUT /api/v1/tenants/:id/config
  Body: { fields to update }
  Response: 200 { updated config }
```

### Usage Stats API

```markdown
GET USAGE STATS:
  GET /api/v1/tenants/:id/usage
  Response: 200 {
    "active_users": 38,
    "total_entities": 12,
    "total_transactions": 7234,
    "storage_used_mb": 4812,
    "api_calls_today": 12450,
    "avg_response_time_ms": 145,
    "error_rate": 0.003,
    "monthly_revenue": 25000,
    "last_calculated_at": "2024-07-15T10:00:00Z"
  }
```

### Context Management API

```markdown
RESOLVE TENANT (subdomain → ID):
  GET /api/v1/tenants/resolve?subdomain=coastalcoffee
  Response: 200 { "tenant_id": "uuid" }

VALIDATE TENANT ACCESS:
  GET /api/v1/tenants/:id/validate
  Response: 200 { "valid": true, "status": "ACTIVE" }
  Response: 403 { "valid": false, "reason": "Tenant is suspended" }
```

### Error Responses

```markdown
STANDARD ERROR FORMAT:

{
  "error": {
    "code": "TENANT_NOT_FOUND",
    "message": "Tenant not found: uuid",
    "details": {}
  }
}

ERROR CODES:

TENANT_NOT_FOUND          404    Tenant UUID does not exist
TENANT_DELETED            410    Tenant has been soft-deleted
TENANT_SUSPENDED          403    Tenant is suspended (write blocked)
TENANT_ARCHIVED           403    Tenant is archived (all access blocked)
TENANT_LIMIT_EXCEEDED     429    Resource limit reached
MODULE_NOT_AVAILABLE      403    Module not in allowed_modules
INVALID_OPERATION         400    Invalid status transition
SUBDOMAIN_NOT_FOUND       404    Subdomain not registered
DUPLICATE_SUBDOMAIN       409    Subdomain already taken
```

---

Next: [Troubleshooting Guide](./20-troubleshooting-guide.md)
