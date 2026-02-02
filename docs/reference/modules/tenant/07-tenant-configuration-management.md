[<-- Back to Index](README.md)

## Tenant Configuration Management

### Overview

Each tenant has a dedicated configuration record that controls resource limits, enabled modules, financial settings, security policies, and API rate limits.

### Configuration Record Structure

```markdown
TENANT CONFIGURATION (tenant_configurations):

Resource Limits:
├── max_users                  : Maximum user accounts
├── max_entities               : Maximum companies/branches
├── max_transactions_per_month : Monthly transaction cap
└── storage_quota_mb           : Storage in megabytes

Module Access:
└── allowed_modules (JSONB)    : List of enabled module names
    Example: ["financial", "selling", "buying", "inventory"]

Financial Settings:
├── accounting_method          : FIFO, LIFO, or WEIGHTED_AVERAGE
└── fiscal_year_start_month    : 1-12 (January = 1)

Security Policies:
└── password_policy (JSONB):
    ├── min_length              : Minimum password length
    ├── require_uppercase       : Boolean
    ├── require_lowercase       : Boolean
    ├── require_numbers         : Boolean
    ├── require_special_chars   : Boolean
    └── max_age_days            : Password expiry period

API Controls:
└── api_rate_limits (JSONB):
    ├── requests_per_minute     : Rate limit threshold
    ├── requests_per_hour       : Hourly cap
    └── burst_limit             : Max burst requests
```

### Configuration by Plan

```markdown
PLAN COMPARISON:

Feature                    Basic      Professional   Enterprise
─────────────────────────────────────────────────────────────
max_users                  10         50             Unlimited
max_entities               5          20             Unlimited
max_transactions/month     1,000      10,000         Unlimited
storage_quota_mb           1,024      10,240         102,400
api_rate_limit (req/min)   100        500            2,000

Modules Included:
  Financial                ✓          ✓              ✓
  Selling                  ✓          ✓              ✓
  Buying                   ✗          ✓              ✓
  Inventory                ✗          ✓              ✓
  HR                       ✗          ✗              ✓
  Manufacturing            ✗          ✗              ✓
  Projects                 ✗          ✗              ✓

Password Policy:
  min_length               8          10             12
  require_uppercase        ✗          ✓              ✓
  require_special_chars    ✗          ✗              ✓
  max_age_days             None       90             60
```

### Auto-Creation via Trigger

When a new tenant is inserted, the configuration is created automatically:

```markdown
TRIGGER: create_default_tenant_config

Event:  AFTER INSERT on tenants
Action: INSERT into tenant_configurations with defaults

Default Values:
├── max_users:                  10
├── max_entities:               5
├── max_transactions_per_month: 1,000
├── storage_quota_mb:           1,024
├── fiscal_year_start_month:    1 (January)
├── accounting_method:          'FIFO'
├── allowed_modules:            '["financial", "selling"]'
├── password_policy:            '{"min_length": 8}'
└── api_rate_limits:            '{"requests_per_minute": 100}'

Note: These defaults can be overridden immediately after
provisioning based on the selected plan.
```

### Updating Configuration

```markdown
EXAMPLE: Upgrading Coastal Coffee Co. from Professional to Enterprise

BEFORE (Professional):
  max_users:            50
  max_entities:         20
  max_transactions:     10,000
  storage_quota:        10,240 MB
  allowed_modules:      ["financial", "selling", "buying", "inventory"]

AFTER (Enterprise):
  max_users:            999999 (effectively unlimited)
  max_entities:         999999
  max_transactions:     999999
  storage_quota:        102,400 MB
  allowed_modules:      ["financial", "selling", "buying",
                         "inventory", "hr", "manufacturing", "projects"]

AUDIT LOG:
  Action:    UPDATE_CONFIGURATION
  Actor:     Platform Admin (admin@awo-erp.com)
  Tenant:    Coastal Coffee Co.
  Changes:   Plan upgrade Professional → Enterprise
  Timestamp: 2024-03-15 10:30:00 EAT
```

### Limit Enforcement

The `check_tenant_limits()` function validates resource usage before allowing operations:

```markdown
LIMIT CHECK FLOW:

  API Request: "Create new user for Coastal Coffee"
       │
       ▼
  check_tenant_limits(tenant_id, 'users', current_count)
       │
       ├── Fetch: max_users from tenant_configurations
       ├── Fetch: active_users from tenant_usage_stats
       │
       ├── IF active_users < max_users → ALLOW
       │
       └── IF active_users >= max_users → RAISE EXCEPTION
           "Tenant has reached maximum user limit (50)"
```

---

Next: [Row-Level Security & Isolation](./08-row-level-security-and-isolation.md)
