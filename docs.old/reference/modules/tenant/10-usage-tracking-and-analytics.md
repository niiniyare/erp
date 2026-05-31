[<-- Back to Index](README.md)

## Usage Tracking & Analytics

### Overview

The `tenant_usage_stats` table tracks real-time resource consumption for every tenant. This data powers billing, limit enforcement, health monitoring, and analytics dashboards.

### Tracked Metrics

```markdown
USAGE METRICS (tenant_usage_stats):

User Activity:
├── active_users          : Currently active user accounts
└── Updated when:         Users created, deactivated, deleted

Data Volume:
├── total_entities        : Companies, branches, departments
├── total_transactions    : Invoices, payments, journal entries
└── Updated when:         Records created or deleted

Storage:
├── storage_used_mb       : Disk space consumed by tenant
└── Updated when:         Files uploaded, documents stored

API Performance:
├── api_calls_today       : API requests for current day
├── avg_response_time_ms  : Average response latency
├── error_rate            : Percentage of failed requests
└── Updated:              Periodically by monitoring system

Revenue:
├── monthly_revenue       : Current month billing amount
└── Updated:              By billing system

Timing:
└── last_calculated_at    : When stats were last refreshed
```

### Usage vs Limits Comparison

```markdown
EXAMPLE: Coastal Coffee Co. Usage Dashboard

Metric                  Used        Limit       Utilization
──────────────────────────────────────────────────────────
Active Users            38          50          76%  ████████░░
Entities                12          20          60%  ██████░░░░
Transactions/Month      7,234       10,000      72%  ███████░░░
Storage                 4,812 MB    10,240 MB   47%  █████░░░░░
API Calls Today         12,450      N/A         -
Avg Response Time       145 ms      N/A         -
Error Rate              0.3%        N/A         -
Monthly Revenue         KES 25,000  N/A         -
```

### Limit Enforcement

The `check_tenant_limits()` function validates resource usage:

```markdown
LIMIT CHECK FLOW:

  Operation: Create 39th user for Coastal Coffee

  check_tenant_limits('bbb-222', 'users', 38)
  │
  ├── Fetch config: max_users = 50
  ├── Fetch usage:  active_users = 38
  ├── Check: 38 < 50 → ALLOW
  └── User created successfully

  Later: Create 51st user

  check_tenant_limits('bbb-222', 'users', 50)
  │
  ├── Fetch config: max_users = 50
  ├── Fetch usage:  active_users = 50
  ├── Check: 50 >= 50 → DENY
  └── EXCEPTION: "Tenant has reached maximum user limit (50)"

STORAGE CHECK:

  check_tenant_limits('bbb-222', 'storage', current_mb)
  │
  ├── Fetch config: storage_quota_mb = 10,240
  ├── Fetch usage:  storage_used_mb = 4,812
  ├── Check: 4,812 < 10,240 → ALLOW
  └── Upload proceeds
```

### Analytics Queries

```markdown
COMMON ANALYTICS OPERATIONS:

1. Tenant Health Score:
   Combine: error_rate, avg_response_time, active_users
   Score: 0-100 based on weighted metrics

2. Usage Trends:
   Track: transactions over time
   Alert: When approaching 80% of limit

3. Churn Risk Indicators:
   Monitor: Declining active_users
   Monitor: Decreasing transaction volume
   Monitor: Increasing error_rate

4. Revenue Analytics:
   Aggregate: monthly_revenue across tenants
   Compare:   Revenue vs plan tier
   Identify:  Upsell candidates (high usage on lower plan)

5. Platform Capacity:
   Sum: storage_used_mb across all tenants
   Sum: api_calls_today across all tenants
   Plan: Infrastructure scaling needs
```

### Auto-Update Mechanism

```markdown
USAGE STAT UPDATES:

Real-Time Updates (via application):
├── active_users:      On user create/deactivate
├── total_entities:    On entity create/delete
├── total_transactions: On transaction create
└── storage_used_mb:   On file upload/delete

Periodic Updates (via background job):
├── api_calls_today:      Reset daily, increment per request
├── avg_response_time_ms: Calculated from request logs
├── error_rate:           Calculated from error logs
├── monthly_revenue:      Synced from billing system
└── last_calculated_at:   Updated on each calculation run
```

---

Next: [Subscription Plans & Limits](./11-subscription-plans-and-limits.md)
