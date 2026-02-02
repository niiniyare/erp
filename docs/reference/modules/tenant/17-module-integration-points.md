[<-- Back to Index](README.md)

## Module Integration Points

### Overview

The Tenant Module is the foundation that every other module depends on. All data access flows through tenant context, and all resource limits are enforced by tenant configuration.

### Integration Map

```markdown
TENANT MODULE INTEGRATION:

                    ┌─────────────────┐
                    │  Tenant Module   │
                    │  (Foundation)    │
                    └────────┬────────┘
                             │
        ┌────────────┬───────┼───────┬────────────┐
        ▼            ▼       ▼       ▼            ▼
  ┌──────────┐ ┌─────────┐ ┌────┐ ┌──────────┐ ┌────┐
  │Financial │ │ Selling │ │Buy │ │Inventory │ │ HR │
  └──────────┘ └─────────┘ └────┘ └──────────┘ └────┘

Every module:
├── Receives tenant context from middleware
├── Has RLS policies filtering by current_tenant_id()
├── Respects resource limits from tenant_configurations
└── Logs operations in tenant-scoped audit trail
```

### Financial Module Integration

```markdown
FINANCIAL ↔ TENANT:

Tenant Provides:
├── currency_code        → Default currency for all transactions
├── fiscal_year_start    → When the fiscal year begins (month 1-12)
├── accounting_method    → FIFO, LIFO, or WEIGHTED_AVERAGE
├── tenant_id            → Scopes chart of accounts, journals, entries
└── max_transactions     → Monthly transaction limit

Financial Uses:
├── Tenant context for GL entry isolation
├── Currency setting for report formatting
├── Fiscal year for period calculations
└── Transaction limits before posting entries

Example:
  Tenant: Coastal Coffee (KES, fiscal year starts July)
  Action: Post journal entry
  Check:  total_transactions < max_transactions_per_month
  Scope:  Journal entries filtered by tenant_id via RLS
  Format: Amounts in KES per tenant currency_code
```

### Selling Module Integration

```markdown
SELLING ↔ TENANT:

Tenant Provides:
├── tenant_id            → Scopes customers, orders, invoices
├── allowed_modules      → Must include "selling"
├── max_entities         → Limits customer records
└── currency_code        → Default pricing currency

Selling Uses:
├── Tenant context for customer isolation
├── Module access check before any sales operation
├── Entity limits when creating new customers
└── Currency for pricing and invoicing

Example:
  Tenant: Highland Textiles (Basic plan)
  Action: Create new customer
  Check:  "selling" IN allowed_modules → YES (Basic includes selling)
  Check:  total_entities < max_entities (5) → 4 < 5 → ALLOW
  Scope:  Customer visible only to Highland Textiles users
```

### Buying Module Integration

```markdown
BUYING ↔ TENANT:

Tenant Provides:
├── tenant_id            → Scopes vendors, purchase orders
├── allowed_modules      → Must include "buying"
├── max_entities         → Limits vendor records
└── currency_code        → Default purchase currency

Buying Uses:
├── Tenant context for vendor isolation
├── Module access check (not available on Basic plan)
├── Entity limits for vendor creation
└── Multi-currency support per tenant settings

Example:
  Tenant: Highland Textiles (Basic plan)
  Action: Access Buying module
  Check:  "buying" IN allowed_modules → NO (Basic excludes buying)
  Result: 403 - Module not available on current plan
  Suggestion: Upgrade to Professional
```

### Entity Module Integration

```markdown
ENTITY ↔ TENANT:

Tenant Provides:
├── tenant_id            → Scopes companies, branches, departments
├── max_entities         → Limits total entity count
└── company_size         → Determines default structure

Entity Uses:
├── Tenant context for entity isolation
├── Entity limits from configuration
└── enforce_tenant_isolation() trigger for FK validation

Cross-Tenant Protection:
  Trigger: enforce_tenant_isolation()
  Table:   persons (and future tables)
  Check:   FK references (entity_id) belong to same tenant
  Block:   Cross-tenant entity references
```

### HR Module Integration

```markdown
HR ↔ TENANT:

Tenant Provides:
├── tenant_id            → Scopes employees, payroll, leave
├── allowed_modules      → Must include "hr" (Enterprise only)
├── max_users            → Relates to employee count
└── timezone             → Attendance and leave calculations

HR Uses:
├── Tenant context for employee data isolation
├── Module access (Enterprise plan required)
├── User limits for employee onboarding
└── Timezone for shift and attendance management
```

### Common Integration Pattern

```markdown
EVERY MODULE FOLLOWS THIS PATTERN:

1. Middleware:
   ├── Extract tenant ID (header or subdomain)
   ├── Call ValidateTenantAccess()
   └── Call SetTenant() to set DB context

2. Service Layer:
   ├── Check module access (allowed_modules)
   ├── Check resource limits (check_tenant_limits)
   └── Execute business logic

3. Repository Layer:
   ├── All queries auto-filtered by RLS
   ├── No manual tenant_id WHERE clauses needed
   └── Cross-tenant references blocked by triggers

4. Cleanup:
   └── Call ResetTenant() to clear context
```

---

Next: [Common Business Scenarios](./18-common-business-scenarios.md)
