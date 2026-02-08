[<-- Back to Index](README.md)

## Business Rules & Validation

> **Purpose**: This document defines the business rules, validation constraints, and operational guidelines for multi-tenant system management. These rules ensure data integrity, tenant isolation, and consistent system behavior.

---

## Table of Contents

1. [Status Transition Rules](#status-transition-rules)
2. [Data Validation Rules](#data-validation-rules)
3. [Resource Limit Rules](#resource-limit-rules)
4. [Soft Delete Rules](#soft-delete-rules)
5. [Configuration Rules](#configuration-rules)
6. [Tenant Isolation Rules](#tenant-isolation-rules)
7. [Quick Reference](#quick-reference)
8. [Troubleshooting](#troubleshooting)

---

## Status Transition Rules

### Overview

Tenant status follows a strict state machine model to ensure proper lifecycle management and prevent invalid transitions.

### State Diagram

```
    ┌─────────┐
    │ PENDING │ (Initial state)
    └────┬────┘
         │
         │ activate
         ▼
    ┌────────┐      suspend       ┌───────────┐
    │ ACTIVE │◄──────────────────►│ SUSPENDED │
    └───┬────┘      reactivate    └─────┬─────┘
        │                               │
        │ archive                       │ archive
        ▼                               ▼
    ┌──────────┐                   ┌──────────┐
    │ ARCHIVED │◄──────────────────┤ ARCHIVED │
    └──────────┘   (Terminal)      └──────────┘
```

### Valid Transitions

| From State | To State | Required Conditions | Example Use Case |
|-----------|----------|-------------------|------------------|
| `PENDING` | `ACTIVE` | • Admin approval OR auto-activation trigger<br>• All required fields complete<br>• Payment method on file (if applicable) | New tenant completes onboarding |
| `ACTIVE` | `SUSPENDED` | • Suspension reason (required)<br>• One of: payment_failure, policy_violation, security_concern, admin_request | Payment failed for 3rd consecutive month |
| `ACTIVE` | `ARCHIVED` | • Archive reason (required)<br>• Data retention policy specified<br>• Final export completed (if requested) | Tenant cancels subscription permanently |
| `SUSPENDED` | `ACTIVE` | • Reactivation reason (required)<br>• Original suspension issue resolved<br>• Verification of resolution | Payment issue resolved |
| `SUSPENDED` | `ARCHIVED` | • Archive reason (required)<br>• Retention policy specified | Suspended tenant never returned |

### Invalid Transitions

| From State | To State | Why Blocked | Alternative Action |
|-----------|----------|-------------|-------------------|
| `PENDING` | `SUSPENDED` | Cannot suspend before activation | Complete activation or reject application |
| `PENDING` | `ARCHIVED` | Must activate first to establish baseline | Delete application if rejected |
| `ARCHIVED` | `ACTIVE` | Terminal state - no reactivation | Create new tenant if customer returns |
| `ARCHIVED` | `SUSPENDED` | Already permanently deactivated | N/A - consider new tenant |
| `SUSPENDED` | `PENDING` | Cannot revert to initial state | Reactivate or archive |

### Implementation Example

```python
def transition_status(tenant, new_status, reason=None, **kwargs):
    """
    Transition tenant to new status with validation.
    
    Raises:
        InvalidTransitionError: If transition not allowed
        ValidationError: If required fields missing
    """
    valid_transitions = {
        'PENDING': ['ACTIVE'],
        'ACTIVE': ['SUSPENDED', 'ARCHIVED'],
        'SUSPENDED': ['ACTIVE', 'ARCHIVED'],
        'ARCHIVED': []  # Terminal state
    }
    
    if new_status not in valid_transitions[tenant.status]:
        raise InvalidTransitionError(
            f"Cannot transition from {tenant.status} to {new_status}"
        )
    
    # Validate required fields based on transition
    if new_status in ['SUSPENDED', 'ARCHIVED'] and not reason:
        raise ValidationError("Reason required for suspension/archival")
    
    # Perform transition
    tenant.status = new_status
    tenant.status_reason = reason
    tenant.status_changed_at = now()
    tenant.status_changed_by = current_user()
    
    # Create audit log
    create_audit_log('status_change', tenant, old=tenant.status, new=new_status)
```

---

## Data Validation Rules

### Tenant Creation Fields

#### Name

| Constraint | Rule | Example | Notes |
|-----------|------|---------|-------|
| Required | Yes | ✅ "Acme Corporation" | Cannot be empty or null |
| Max Length | 255 characters | ❌ "A" × 300 | Database column limit |
| Whitespace | Trimmed automatically | "  Acme Corp  " → "Acme Corp" | Leading/trailing only |
| Special Characters | Allowed | ✅ "Smith & Sons Ltd." | No restrictions |
| Uniqueness | Not enforced | ✅ Multiple "Acme Corp" tenants allowed | Use slug for uniqueness |

**Validation Logic**:
```javascript
name:
  - required: true
  - maxLength: 255
  - transform: trim
  - message: "Company name is required and must be under 255 characters"
```

#### Email

| Constraint | Rule | Example | Notes |
|-----------|------|---------|-------|
| Required | Yes | ✅ admin@acme.com | Primary contact |
| Format | RFC 5322 compliant | ❌ "not-an-email" | Standard email validation |
| Max Length | 255 characters | ❌ "a" × 240 + "@domain.com" | Including domain |
| Case Sensitivity | Case-insensitive | "Admin@Acme.com" = "admin@acme.com" | Normalized to lowercase |
| Uniqueness | Across non-deleted tenants | ❌ Duplicate active emails | Can reuse after soft delete |

**Validation Logic**:
```javascript
email:
  - required: true
  - format: email
  - maxLength: 255
  - transform: toLowerCase
  - unique: { where: { deleted_at: null } }
  - message: "Valid email required (must be unique)"
```

**Example Errors**:
- `"email is required"` - Field missing
- `"email must be a valid email address"` - Format invalid
- `"email already exists for another active tenant"` - Duplicate

#### Subdomain

| Constraint | Rule | Example | Notes |
|-----------|------|---------|-------|
| Required | No (Optional) | ✅ null or "acme-corp" | Can be set later |
| Max Length | 63 characters | ❌ "a" × 70 | DNS label limit (RFC 1035) |
| Format | `^[a-z0-9]([a-z0-9-]*[a-z0-9])?$` | ✅ "acme-123"<br>❌ "-acme" or "acme-" | DNS-safe characters only |
| Reserved Words | Blocked | ❌ "admin", "api", "www", "app" | System reserved |
| Uniqueness | Across non-deleted tenants | ❌ Duplicate subdomains | One per tenant |

**Reserved Subdomain List**:
```
admin, api, www, app, cdn, static, assets, mail, ftp, smtp,
dev, staging, prod, production, test, localhost, 
dashboard, portal, auth, login, signup, register,
billing, payment, invoice, support, help, docs, status,
blog, news, about, contact, legal, privacy, terms
```

**Validation Logic**:
```javascript
subdomain:
  - required: false
  - maxLength: 63
  - pattern: /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/
  - transform: toLowerCase
  - notIn: RESERVED_SUBDOMAINS
  - unique: { where: { deleted_at: null } }
```

#### Slug

| Constraint | Rule | Example | Notes |
|-----------|------|---------|-------|
| Required | Auto-generated | "acme-corporation" | From name field |
| Max Length | 50 characters | Truncated if needed | Allows UUID suffix |
| Format | Lowercase, alphanumeric, hyphens | "acme-corp-123" | Generated from name |
| Uniqueness | Enforced | If collision, append UUID | "acme-corp-a1b2c3d4" |
| Generation | Automatic | Cannot be manually set | System-controlled |

**Generation Algorithm**:
```python
def generate_slug(name):
    # Convert to lowercase and replace spaces/special chars with hyphens
    base = re.sub(r'[^a-z0-9]+', '-', name.lower()).strip('-')
    
    # Truncate to 40 chars (leave room for UUID suffix)
    base = base[:40].rstrip('-')
    
    # Check uniqueness
    slug = base
    if Tenant.exists(slug=slug, deleted_at=None):
        # Append short UUID if collision
        suffix = uuid4().hex[:8]
        slug = f"{base}-{suffix}"
    
    return slug
```

**Examples**:
- "Acme Corporation" → `acme-corporation`
- "Smith & Sons Ltd." → `smith-sons-ltd`
- "ABC-123 Company!!!" → `abc-123-company`
- "Acme Corporation" (duplicate) → `acme-corporation-a1b2c3d4`

#### Currency Code

| Constraint | Rule | Example | Notes |
|-----------|------|---------|-------|
| Required | Yes | "USD" | Financial transactions |
| Length | Exactly 3 characters | ❌ "US" or "USDT" | ISO 4217 standard |
| Format | Uppercase letters | "USD", "EUR", "KES" | Validated against ISO list |
| Default | USD | Applied if not specified | Most common currency |
| Immutability | Cannot change after transactions | Locked after first invoice | Prevents data inconsistency |

**Supported Currencies** (Common):
```
USD - US Dollar          EUR - Euro               GBP - British Pound
KES - Kenyan Shilling    NGN - Nigerian Naira    ZAR - South African Rand
JPY - Japanese Yen       CNY - Chinese Yuan       INR - Indian Rupee
AUD - Australian Dollar  CAD - Canadian Dollar    CHF - Swiss Franc
```

#### Timezone

| Constraint | Rule | Example | Notes |
|-----------|------|---------|-------|
| Required | Yes | "Africa/Nairobi" | IANA timezone database |
| Format | Region/City or UTC | ✅ "America/New_York"<br>✅ "UTC" | Case-sensitive |
| Validation | Must exist in IANA TZ DB | ❌ "EST" or "GMT+3" | Use full names |
| Default | UTC | Neutral default | User can change |

**Common Timezones**:
```
UTC                     # Coordinated Universal Time
Africa/Nairobi         # East Africa Time (EAT)
America/New_York       # Eastern Time (US)
Europe/London          # British Time
Asia/Tokyo             # Japan Standard Time
Australia/Sydney       # Australian Eastern Time
```

**Validation Logic**:
```python
from pytz import all_timezones

def validate_timezone(tz):
    if tz not in all_timezones:
        raise ValidationError(f"Invalid timezone: {tz}")
```

#### Company Size

| Constraint | Rule | Example | Notes |
|-----------|------|---------|-------|
| Required | No | Default: "Small" | Used for analytics |
| Valid Values | `Small`, `Medium`, `Large`, `Enterprise` | Case-sensitive | Enum type |
| Default | Small | Applied if not specified | Most common |

**Size Definitions** (Guidelines):
- **Small**: 1-50 employees
- **Medium**: 51-250 employees
- **Large**: 251-1000 employees
- **Enterprise**: 1000+ employees

#### Status

| Constraint | Rule | Example | Notes |
|-----------|------|---------|-------|
| Required | Yes | "PENDING" | Lifecycle state |
| Valid Values | `PENDING`, `ACTIVE`, `SUSPENDED`, `ARCHIVED` | Uppercase | Enum type |
| Default | PENDING | All new tenants | Must activate to use |
| Immutability | Use transition methods | Don't set directly | Enforced workflows |

---

## Resource Limit Rules

### Overview

Resource limits prevent abuse, ensure fair usage, and maintain system performance. Limits are enforced at the application layer before resource creation.

### Limit Types

#### 1. User Limit

**Rule**: `active_users_count < max_users`

**When Checked**: Before creating a new user account

**Enforcement**:
```python
def create_user(tenant, user_data):
    active_count = tenant.users.filter(status='ACTIVE').count()
    
    if active_count >= tenant.max_users:
        raise ResourceLimitError(
            f"Tenant has reached maximum user limit ({tenant.max_users}). "
            f"Currently: {active_count} active users. "
            f"Contact support to upgrade your plan."
        )
    
    return User.create(**user_data)
```

**Override**: Platform administrators can temporarily increase limits

**User Experience**:
- Display usage: "5 of 10 users" in dashboard
- Warning at 80%: "You've used 8 of 10 available user slots"
- Upgrade prompt: "Need more users? Upgrade your plan"

#### 2. Entity Limit

**Rule**: `total_entities < max_entities`

**Entities**: Companies, branches, departments, cost centers

**When Checked**: Before creating any new entity

**Enforcement**:
```python
def create_entity(tenant, entity_type, data):
    total = tenant.get_entity_count()  # All entity types combined
    
    if total >= tenant.max_entities:
        raise ResourceLimitError(
            f"Entity limit reached ({tenant.max_entities}). "
            f"Current: {total} entities. "
            f"Delete unused entities or upgrade."
        )
    
    return Entity.create(type=entity_type, **data)
```

**Typical Limits by Plan**:
- Starter: 5 entities
- Professional: 25 entities
- Enterprise: 100+ entities

#### 3. Transaction Limit

**Rule**: `transactions_this_month < max_transactions_per_month`

**When Checked**: Before posting any financial transaction (invoice, payment, journal entry)

**Reset**: First day of each month (tenant's timezone)

**Enforcement**:
```python
def post_transaction(tenant, transaction):
    current_month = get_month_start(tenant.timezone)
    count = tenant.transactions.filter(
        created_at__gte=current_month
    ).count()
    
    if count >= tenant.max_transactions_per_month:
        raise ResourceLimitError(
            f"Monthly transaction limit reached ({tenant.max_transactions_per_month}). "
            f"Limit resets on {next_month_start(tenant.timezone)}. "
            f"Upgrade for higher limits."
        )
    
    return Transaction.create(**transaction)
```

**Grace Period**: 10% overage allowed for end-of-month processing

**Typical Limits**:
- Starter: 500 transactions/month
- Professional: 5,000 transactions/month
- Enterprise: Unlimited

#### 4. Storage Limit

**Rule**: `storage_used_mb < storage_quota_mb`

**Includes**: Documents, attachments, exported files, backups

**When Checked**: Before any file upload

**Enforcement**:
```python
def upload_file(tenant, file):
    file_size_mb = file.size / (1024 * 1024)
    total_used = tenant.calculate_storage_used()
    
    if (total_used + file_size_mb) > tenant.storage_quota_mb:
        raise ResourceLimitError(
            f"Storage quota exceeded. "
            f"Used: {total_used:.1f} MB / {tenant.storage_quota_mb} MB. "
            f"This file: {file_size_mb:.1f} MB."
        )
    
    # Warning threshold
    if total_used > (tenant.storage_quota_mb * 0.8):
        notify_tenant_storage_warning(tenant, total_used)
    
    return File.upload(file)
```

**Monitoring**:
- Warning at 80% utilization
- Alert at 90% utilization
- Soft block at 100% (admin override available)

**Cleanup**:
- Auto-delete exports older than 30 days
- Compress attachments over 1 MB
- Suggest archival of old documents

#### 5. API Rate Limit

**Rule**: `requests_in_window < rate_limit_per_minute`

**When Checked**: Every API request (middleware layer)

**Enforcement**:
```python
@middleware
def rate_limit_check(request, tenant):
    key = f"rate_limit:{tenant.id}:{current_minute()}"
    current = redis.incr(key)
    
    if current == 1:
        redis.expire(key, 60)  # 1-minute window
    
    if current > tenant.api_rate_limit_per_minute:
        raise RateLimitExceededError(
            status_code=429,
            message=f"Rate limit exceeded: {tenant.api_rate_limit_per_minute} req/min",
            retry_after=redis.ttl(key)
        )
    
    # Add headers
    response.headers['X-RateLimit-Limit'] = tenant.api_rate_limit_per_minute
    response.headers['X-RateLimit-Remaining'] = tenant.api_rate_limit_per_minute - current
    response.headers['X-RateLimit-Reset'] = current_minute() + 60
```

**Rate Limit Tiers**:
- Free: 60 requests/minute
- Professional: 600 requests/minute
- Enterprise: 6,000 requests/minute

**Burst Allowance**: 2× rate for up to 10 seconds

**Response Headers**:
```
HTTP/1.1 429 Too Many Requests
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1643472120
Retry-After: 45
Content-Type: application/json

{
  "error": "rate_limit_exceeded",
  "message": "Rate limit of 60 requests per minute exceeded",
  "retry_after": 45
}
```

### Limit Monitoring Dashboard

**Admin View**:
```
Resource Usage - Acme Corporation (Plan: Professional)
─────────────────────────────────────────────────────
Users:         8 / 10       [████████--] 80%  ⚠️ Warning
Entities:     15 / 25       [██████----] 60%
Transactions: 3,245 / 5,000 [██████----] 65%  (Resets: Mar 1)
Storage:      850 MB / 2 GB [████------] 42%
API Rate:     45 / 600 rpm  [█---------] 8%   (Current minute)
```

---

## Soft Delete Rules

### Overview

Soft deletion marks records as deleted without physically removing them from the database. This enables audit trails, data recovery, and regulatory compliance.

### Implementation

**Database Schema**:
```sql
ALTER TABLE tenants ADD COLUMN deleted_at TIMESTAMP NULL DEFAULT NULL;
CREATE INDEX idx_tenants_deleted_at ON tenants(deleted_at);

-- Unique constraints scoped to non-deleted
CREATE UNIQUE INDEX idx_tenants_slug_unique 
  ON tenants(slug) WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX idx_tenants_subdomain_unique 
  ON tenants(subdomain) WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX idx_tenants_email_unique 
  ON tenants(email) WHERE deleted_at IS NULL;
```

### Delete Operation

**Process Flow**:
```python
def soft_delete_tenant(tenant_id, reason):
    """
    Soft delete a tenant with audit trail.
    
    Steps:
    1. Validate tenant can be deleted
    2. Set deleted_at timestamp
    3. Create audit log entry
    4. Notify relevant parties
    5. Schedule data retention cleanup
    """
    tenant = Tenant.find(tenant_id)
    
    # Validation
    if tenant.status not in ['ARCHIVED']:
        raise ValidationError("Tenant must be archived before deletion")
    
    # Perform soft delete
    tenant.deleted_at = now()
    tenant.deleted_by = current_user()
    tenant.deletion_reason = reason
    tenant.save()
    
    # Audit trail
    create_audit_log('tenant_deleted', {
        'tenant_id': tenant.id,
        'tenant_name': tenant.name,
        'reason': reason,
        'retention_until': now() + timedelta(days=90)
    })
    
    # Notifications
    notify_admins(f"Tenant {tenant.name} soft deleted")
    
    # Schedule permanent deletion
    schedule_hard_delete(tenant, after_days=90)
```

### Query Filtering

**Automatic Filtering**:
```python
# All queries automatically filter soft-deleted records
class TenantQuerySet:
    def get_queryset(self):
        return super().get_queryset().filter(deleted_at__isnull=True)

# Example: This only returns active tenants
tenants = Tenant.objects.all()  # WHERE deleted_at IS NULL

# To include deleted (admin only)
tenants = Tenant.objects.with_deleted()  # No filter

# Only deleted (for recovery)
tenants = Tenant.objects.only_deleted()  # WHERE deleted_at IS NOT NULL
```

### Unique Constraint Behavior

**Slug Reusability**:
```python
# Create tenant
tenant1 = Tenant.create(name="Acme Corp", slug="acme-corp")

# Soft delete
tenant1.soft_delete()

# Slug becomes available again
tenant2 = Tenant.create(name="Acme Corporation", slug="acme-corp")  # ✅ Allowed
```

### Cascade Behavior

**Non-Cascading Delete**:
```sql
-- Tenant is soft deleted
UPDATE tenants SET deleted_at = NOW() WHERE id = 123;

-- Child records remain UNAFFECTED
-- Users, invoices, transactions still reference tenant_id = 123
-- Access is prevented by application-layer tenant context
```

**Isolation Strategy**:
- Tenant context filter prevents access to child records
- Child records remain in database for audit/compliance
- Hard delete (after retention) removes all data

### Restoration Process

**Restore Deleted Tenant**:
```python
def restore_tenant(tenant_id, restored_by):
    """
    Restore a soft-deleted tenant.
    
    Checks:
    - Verify slug/subdomain still available
    - Ensure within retention period
    - Require platform admin approval
    """
    tenant = Tenant.only_deleted().find(tenant_id)
    
    # Check retention period
    if (now() - tenant.deleted_at).days > 90:
        raise ValidationError("Beyond restoration period (90 days)")
    
    # Check uniqueness
    if Tenant.exists(slug=tenant.slug):
        raise ValidationError(f"Slug '{tenant.slug}' already in use")
    
    if tenant.subdomain and Tenant.exists(subdomain=tenant.subdomain):
        raise ValidationError(f"Subdomain '{tenant.subdomain}' already in use")
    
    # Restore
    tenant.deleted_at = None
    tenant.deleted_by = None
    tenant.deletion_reason = None
    tenant.restored_at = now()
    tenant.restored_by = restored_by
    tenant.save()
    
    create_audit_log('tenant_restored', tenant)
    
    return tenant
```

### Data Retention Schedule

| Phase | Timeframe | Data State | Access | Action |
|-------|-----------|----------|--------|---------|
| Active | 0-30 days | Soft deleted | Admin read-only | Quick restore available |
| Grace | 30-60 days | Soft deleted | Admin read-only | Restore with approval |
| Final | 60-90 days | Soft deleted | No access | Restore requires executive approval |
| Purged | 90+ days | Hard deleted | Permanently removed | Cannot restore |

**Hard Delete** (Permanent Removal):
```python
def hard_delete_tenant(tenant_id):
    """
    Permanently delete tenant and all related data.
    
    WARNING: This is irreversible.
    """
    tenant = Tenant.only_deleted().find(tenant_id)
    
    # Verify deletion period elapsed
    if (now() - tenant.deleted_at).days < 90:
        raise ValidationError("Must wait 90 days before permanent deletion")
    
    # Delete all related data
    tenant.users.hard_delete()
    tenant.transactions.hard_delete()
    tenant.documents.hard_delete()
    # ... all related entities
    
    # Finally delete tenant
    tenant.hard_delete()  # Physical DELETE from database
    
    create_audit_log('tenant_purged', {'tenant_id': tenant_id})
```

---

## Configuration Rules

### Fiscal Year Configuration

**Field**: `fiscal_year_start_month`

**Constraints**:
- Must be between 1 (January) and 12 (December)
- Cannot be changed mid-fiscal-year without migration
- Affects financial reporting periods

**Validation**:
```python
def update_fiscal_year_start(tenant, new_month):
    if not 1 <= new_month <= 12:
        raise ValidationError("Month must be between 1-12")
    
    current_date = now().date()
    fiscal_year_end = get_fiscal_year_end(tenant, current_date)
    
    # Warning if mid-fiscal-year
    if current_date < fiscal_year_end:
        warn(
            "Changing fiscal year mid-period may affect reporting. "
            "Recommend waiting until current fiscal year ends."
        )
        
        if not confirm("Continue anyway?"):
            return
    
    tenant.fiscal_year_start_month = new_month
    tenant.save()
    
    create_audit_log('fiscal_year_changed', {
        'old_month': tenant.fiscal_year_start_month,
        'new_month': new_month
    })
```

**Common Configurations**:
- Calendar Year: `fiscal_year_start_month = 1` (January)
- UK Tax Year: `fiscal_year_start_month = 4` (April)
- US Federal: `fiscal_year_start_month = 10` (October)

### Accounting Method

**Field**: `accounting_method`

**Valid Values**: `FIFO`, `LIFO`, `WEIGHTED_AVERAGE`

**Constraints**:
- Method selection affects inventory valuation
- Changing method requires audit trail
- Cannot change retroactively without recalculation

**Method Comparison**:

| Method | Full Name | Best For | Tax Impact |
|--------|-----------|----------|------------|
| `FIFO` | First-In, First-Out | Perishables, fashion | Higher taxes in inflation |
| `LIFO` | Last-In, First-Out | Non-perishables | Lower taxes in inflation |
| `WEIGHTED_AVERAGE` | Weighted Average Cost | Bulk commodities | Stable, predictable |

**Change Process**:
```python
def change_accounting_method(tenant, new_method):
    if new_method not in ['FIFO', 'LIFO', 'WEIGHTED_AVERAGE']:
        raise ValidationError("Invalid accounting method")
    
    old_method = tenant.accounting_method
    
    # Warning
    warn(
        f"Changing from {old_method} to {new_method} will affect:\n"
        "- Inventory valuations\n"
        "- Cost of goods sold\n"
        "- Financial statements\n"
        "Recommend consulting accountant before proceeding."
    )
    
    # Require reason
    reason = prompt("Reason for change:")
    if not reason:
        raise ValidationError("Change reason required")
    
    # Update
    tenant.accounting_method = new_method
    tenant.save()
    
    # Audit trail
    create_audit_log('accounting_method_changed', {
        'old_method': old_method,
        'new_method': new_method,
        'reason': reason,
        'requires_restatement': True
    })
    
    # Trigger recalculation
    schedule_inventory_recalculation(tenant)
```

### Password Policy

**Configurable Fields**:
```python
password_policy = {
    'min_length': 8,              # Minimum: 8 characters
    'require_uppercase': True,    # At least one A-Z
    'require_lowercase': True,    # At least one a-z
    'require_numbers': True,      # At least one 0-9
    'require_special': True,      # At least one !@#$%^&*
    'max_age_days': 90,          # Password expires after 90 days (null = never)
    'prevent_reuse': 5,          # Cannot reuse last 5 passwords
    'lockout_attempts': 5,       # Lock after 5 failed attempts
    'lockout_duration': 30       # Lockout for 30 minutes
}
```

**Validation**:
```python
def validate_password_policy(policy):
    # Min length
    if policy['min_length'] < 8:
        raise ValidationError("Minimum password length must be at least 8")
    
    # At least one complexity requirement
    complexity_rules = [
        policy.get('require_uppercase'),
        policy.get('require_lowercase'),
        policy.get('require_numbers'),
        policy.get('require_special')
    ]
    
    if not any(complexity_rules):
        warn("Recommended: Enable at least one complexity requirement")
    
    # Max age
    if policy['max_age_days'] and policy['max_age_days'] < 1:
        raise ValidationError("Max age must be positive or null (no expiry)")
```

**Enforcement Example**:
```python
def validate_password(password, tenant):
    policy = tenant.password_policy
    errors = []
    
    if len(password) < policy['min_length']:
        errors.append(f"Must be at least {policy['min_length']} characters")
    
    if policy['require_uppercase'] and not re.search(r'[A-Z]', password):
        errors.append("Must contain uppercase letter")
    
    if policy['require_lowercase'] and not re.search(r'[a-z]', password):
        errors.append("Must contain lowercase letter")
    
    if policy['require_numbers'] and not re.search(r'\d', password):
        errors.append("Must contain number")
    
    if policy['require_special'] and not re.search(r'[!@#$%^&*(),.?":{}|<>]', password):
        errors.append("Must contain special character")
    
    if errors:
        raise ValidationError("Password does not meet requirements: " + "; ".join(errors))
```

### API Rate Limits

**Configuration**:
```python
api_rate_limits = {
    'requests_per_minute': 60,    # Must be > 0
    'burst_multiplier': 2,        # Burst = 2x per-minute rate
    'burst_duration': 10,         # Burst allowed for 10 seconds
    'apply_to_webhooks': False    # Webhooks exempt from rate limit
}
```

**Validation**:
```python
def validate_rate_limits(tenant, limits):
    # Must be positive
    if limits['requests_per_minute'] <= 0:
        raise ValidationError("Rate limit must be greater than 0")
    
    # Cannot exceed plan maximum
    plan_max = tenant.plan.max_api_rate_limit
    if limits['requests_per_minute'] > plan_max:
        raise ValidationError(
            f"Rate limit cannot exceed plan maximum ({plan_max} req/min). "
            "Upgrade plan for higher limits."
        )
    
    # Burst validation
    if limits['burst_multiplier'] > 3:
        warn("Burst multiplier > 3 may impact system performance")
    
    # Reasonable burst duration
    if limits['burst_duration'] > 60:
        warn("Burst duration over 60s defeats rate limiting purpose")
```

### Allowed Modules

**Module System**:
```python
AVAILABLE_MODULES = [
    'financial',      # Always included (core)
    'selling',        # Always included (core)
    'buying',         # Purchase orders, suppliers
    'inventory',      # Stock management
    'manufacturing',  # Production, BOM
    'projects',       # Project management
    'hr',            # Human resources
    'payroll',       # Payroll processing
    'crm',           # Customer relationship
    'support',       # Help desk
]

PLAN_MODULES = {
    'starter': ['financial', 'selling'],
    'professional': ['financial', 'selling', 'buying', 'inventory', 'crm'],
    'enterprise': AVAILABLE_MODULES  # All modules
}
```

**Validation**:
```python
def enable_module(tenant, module_name):
    # Valid module?
    if module_name not in AVAILABLE_MODULES:
        raise ValidationError(f"Invalid module: {module_name}")
    
    # Already enabled?
    if module_name in tenant.enabled_modules:
        return  # No-op
    
    # Core modules always enabled
    if module_name in ['financial', 'selling']:
        return  # Cannot disable
    
    # Check plan allows module
    allowed = PLAN_MODULES[tenant.plan.tier]
    if module_name not in allowed:
        raise ValidationError(
            f"Module '{module_name}' not available in {tenant.plan.tier} plan. "
            f"Upgrade to access this module."
        )
    
    # Enable module
    tenant.enabled_modules.append(module_name)
    tenant.save()
    
    create_audit_log('module_enabled', {'module': module_name})
```

---

## Tenant Isolation Rules

### Security Invariants

These rules MUST be enforced at all times to maintain multi-tenant security:

#### 1. Row-Level Security (RLS)

**Rule**: Every tenant-scoped table has RLS policies enabled

**Implementation**:
```sql
-- Enable RLS on table
ALTER TABLE invoices ENABLE ROW LEVEL SECURITY;

-- Policy: Users can only see their tenant's data
CREATE POLICY tenant_isolation_policy ON invoices
    FOR ALL
    TO application_role
    USING (tenant_id = current_setting('app.tenant_id')::INTEGER);

-- Admin bypass (platform admin only)
CREATE POLICY admin_all_access ON invoices
    FOR ALL
    TO admin_role
    USING (true);
```

**Tables Requiring RLS**:
```
✅ All business data tables:
   - invoices, payments, customers, products, users, etc.

❌ Exempt tables:
   - tenants (tenant list itself)
   - system_settings (global configuration)
   - audit_logs (access controlled separately)
```

#### 2. Application Role Filtering

**Rule**: All queries through `application_role` are automatically filtered by tenant

**Connection Setup**:
```python
def set_tenant_context(connection, tenant_id):
    """
    Set tenant context for database connection.
    Must be called before any data access.
    """
    connection.execute(f"SET app.tenant_id = '{tenant_id}'")
    connection.execute(f"SET ROLE application_role")

def clear_tenant_context(connection):
    """
    Clear tenant context after transaction.
    Called automatically on connection close.
    """
    connection.execute("RESET app.tenant_id")
    connection.execute("RESET ROLE")
```

**Middleware**:
```python
@middleware
def tenant_context_middleware(request):
    # Extract tenant from request (subdomain, header, JWT claim, etc.)
    tenant = extract_tenant_from_request(request)
    
    if not tenant:
        raise AuthenticationError("No tenant context")
    
    # Set context
    with db.connection() as conn:
        set_tenant_context(conn, tenant.id)
        
        try:
            response = next_handler(request)
        finally:
            clear_tenant_context(conn)
    
    return response
```

#### 3. Foreign Key Constraints

**Rule**: Foreign keys cannot reference records from different tenants

**Enforcement**:
```sql
-- Example: Invoice -> Customer
CREATE TABLE invoices (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES tenants(id),
    customer_id INTEGER NOT NULL,
    amount DECIMAL(10,2),
    
    -- Compound foreign key ensures same tenant
    FOREIGN KEY (tenant_id, customer_id) 
        REFERENCES customers(tenant_id, customer_id)
);

-- Customer table must have compound unique constraint
CREATE UNIQUE INDEX idx_customers_tenant_customer 
    ON customers(tenant_id, id);
```

#### 4. Mandatory Tenant Context

**Rule**: Tenant context must be set before any data access

**Enforcement**:
```python
class TenantRequiredMiddleware:
    def process_request(self, request):
        # Check if tenant context set
        if not hasattr(request, 'tenant'):
            raise TenantContextError("Tenant context not set")
        
        # Verify tenant is active
        if request.tenant.status != 'ACTIVE':
            raise TenantInactiveError(
                f"Tenant {request.tenant.name} is {request.tenant.status}"
            )

# Database query wrapper
def execute_query(query, params):
    # Verify context
    tenant_id = get_current_tenant_id()
    if not tenant_id:
        raise TenantContextError("No tenant context for query")
    
    return db.execute(query, params)
```

#### 5. Transaction-Scoped Context

**Rule**: Tenant context is automatically cleared after each transaction

**Implementation**:
```python
class TenantScopedConnection:
    def __enter__(self):
        self.conn = db.get_connection()
        self.tenant_id = current_tenant.id
        set_tenant_context(self.conn, self.tenant_id)
        return self.conn
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        # Always clear context
        clear_tenant_context(self.conn)
        self.conn.close()

# Usage
with TenantScopedConnection() as conn:
    conn.execute("SELECT * FROM invoices")  # Tenant-filtered
# Context automatically cleared here
```

#### 6. Admin Role Bypass

**Rule**: Only platform administrators can bypass RLS

**Role Definitions**:
```sql
-- Application role (normal users)
CREATE ROLE application_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO application_role;

-- Admin role (platform admin)
CREATE ROLE admin_role;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO admin_role;
ALTER ROLE admin_role SET row_security = OFF;  -- Bypass RLS

-- Readonly role (monitoring)
CREATE ROLE readonly_role;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO readonly_role;
ALTER ROLE readonly_role SET row_security = OFF;  -- See all tenants
```

**Access Control**:
```python
def switch_to_admin_role(user):
    if not user.is_platform_admin:
        raise PermissionError("Admin role restricted to platform admins")
    
    # Switch role
    db.execute("SET ROLE admin_role")
    
    # Audit
    create_audit_log('admin_access', {
        'user': user.email,
        'reason': 'Platform administration'
    })
```

#### 7. Readonly Role

**Rule**: Monitoring systems can view all tenants without modification rights

**Usage**:
```python
def generate_system_report():
    with db.connection() as conn:
        # Switch to readonly role
        conn.execute("SET ROLE readonly_role")
        
        # Can query across all tenants
        tenants = conn.execute("SELECT * FROM tenants WHERE status = 'ACTIVE'")
        
        for tenant in tenants:
            stats = conn.execute(
                "SELECT COUNT(*) FROM invoices WHERE tenant_id = %s",
                [tenant.id]
            )
            # Generate report...
```

#### 8. Superuser Prohibition

**Rule**: Superuser access is prohibited in production

**Enforcement**:
```sql
-- Production database users should NEVER be superuser
CREATE USER app_user WITH PASSWORD 'secure_password';
GRANT application_role TO app_user;

-- Verify
SELECT usename, usesuper FROM pg_user WHERE usename = 'app_user';
-- usesuper should be 'f' (false)
```

**Monitoring**:
```python
def check_superuser_connections():
    """Alert if any superuser connections detected in production"""
    if os.getenv('ENVIRONMENT') != 'production':
        return
    
    superusers = db.execute("""
        SELECT usename, application_name, client_addr
        FROM pg_stat_activity
        WHERE usesuper = true
    """)
    
    if superusers:
        alert_security_team(
            "CRITICAL: Superuser connection detected in production",
            details=superusers
        )
```

### Isolation Testing

**Test Suite**:
```python
def test_tenant_isolation():
    """Verify tenant isolation is working correctly"""
    
    # Create two test tenants
    tenant_a = create_test_tenant("Tenant A")
    tenant_b = create_test_tenant("Tenant B")
    
    # Create data for each
    with tenant_context(tenant_a):
        invoice_a = Invoice.create(amount=100)
    
    with tenant_context(tenant_b):
        invoice_b = Invoice.create(amount=200)
    
    # Verify isolation
    with tenant_context(tenant_a):
        assert Invoice.count() == 1
        assert Invoice.first().amount == 100
        # Should NOT see tenant B's invoice
    
    with tenant_context(tenant_b):
        assert Invoice.count() == 1
        assert Invoice.first().amount == 200
        # Should NOT see tenant A's invoice
    
    # Verify direct access blocked
    with tenant_context(tenant_a):
        with pytest.raises(PermissionError):
            Invoice.find(invoice_b.id)  # Should fail
```

---

## Quick Reference

### Status Transitions (Valid)

```
PENDING → ACTIVE → SUSPENDED → ARCHIVED
          ↓         ↑
          ARCHIVED  ACTIVE
```

### Validation Checklist

**Tenant Creation**:
- [ ] Name: Required, max 255 chars
- [ ] Email: Required, valid format, unique
- [ ] Subdomain: Optional, DNS-safe, unique, not reserved
- [ ] Slug: Auto-generated, unique
- [ ] Currency: 3-char ISO code, default USD
- [ ] Timezone: Valid IANA timezone
- [ ] Status: Defaults to PENDING

**Resource Limits**:
- [ ] Users: `active_users < max_users`
- [ ] Entities: `total_entities < max_entities`
- [ ] Transactions: `monthly_transactions < limit`
- [ ] Storage: `used_mb < quota_mb`
- [ ] API Rate: `requests_per_minute < limit`

**Isolation**:
- [ ] RLS enabled on all tenant tables
- [ ] Tenant context set before queries
- [ ] Foreign keys enforce same-tenant
- [ ] Admin access audited
- [ ] No superuser in production

### Common Errors

| Error Code | Message | Solution |
|-----------|---------|----------|
| `INVALID_TRANSITION` | Cannot transition from X to Y | Check status transition rules |
| `RESOURCE_LIMIT_EXCEEDED` | User/entity/transaction limit reached | Upgrade plan or delete unused resources |
| `VALIDATION_ERROR` | Field validation failed | Check field constraints |
| `DUPLICATE_KEY` | Email/subdomain/slug already exists | Choose unique value |
| `TENANT_CONTEXT_MISSING` | No tenant context set | Ensure middleware is configured |
| `RATE_LIMIT_EXCEEDED` | Too many requests | Wait and retry, or upgrade plan |

### Reserved Subdomains

```
admin, api, www, app, cdn, static, assets, mail, ftp, smtp,
dev, staging, prod, production, test, localhost, 
dashboard, portal, auth, login, signup, register,
billing, payment, invoice, support, help, docs, status,
blog, news, about, contact, legal, privacy, terms
```

---

## Troubleshooting

### Problem: "Email already exists" on tenant creation

**Cause**: Another active tenant uses this email

**Solution**:
1. Search for existing tenant: `Tenant.find_by_email(email)`
2. If tenant is soft-deleted, it will be available again
3. Use different email, or restore the deleted tenant if same organization

### Problem: "Tenant has reached maximum user limit"

**Cause**: Plan limit exceeded

**Solutions**:
1. **Deactivate unused users**: Archive or remove users no longer needed
2. **Upgrade plan**: Increase user limit
3. **Admin override**: Platform admin can temporarily raise limit

### Problem: "Cannot transition from ARCHIVED to ACTIVE"

**Cause**: ARCHIVED is a terminal state

**Solution**:
- Create new tenant if customer returns
- Archived tenants cannot be reactivated (by design)
- Use soft-delete restoration if within retention period

### Problem: Tenant isolation not working

**Symptoms**: Users seeing other tenants' data

**Debugging**:
```python
# Check current tenant context
print(db.execute("SHOW app.tenant_id"))

# Check current role
print(db.execute("SELECT current_user"))

# Verify RLS enabled
print(db.execute("""
    SELECT tablename, rowsecurity 
    FROM pg_tables 
    WHERE schemaname = 'public'
"""))
```

**Common Causes**:
1. Tenant context not set in middleware
2. Using superuser role (bypasses RLS)
3. RLS not enabled on table
4. Policy not created correctly

### Problem: Rate limit errors for legitimate traffic

**Symptoms**: Frequent 429 errors during normal usage

**Solutions**:
1. **Check burst config**: Ensure burst multiplier is 2×
2. **Optimize requests**: Batch API calls, use webhooks
3. **Upgrade plan**: Higher rate limits available
4. **Request increase**: Contact support for temporary increase

### Problem: Storage quota exceeded unexpectedly

**Investigation**:
```python
# Check storage breakdown
tenant.get_storage_breakdown()
# Returns: {
#   'documents': 450 MB,
#   'attachments': 200 MB,
#   'exports': 150 MB,
#   'backups': 50 MB
# }
```

**Solutions**:
1. Delete old exports (auto-deleted after 30 days)
2. Compress large attachments
3. Archive old documents
4. Upgrade storage quota

---

## Related Documentation

- [Architecture Overview](./01-architecture-overview.md)
- [Tenant Lifecycle](./05-tenant-lifecycle.md)
- [API Reference](./15-api-reference.md)
- [Security & Compliance](./18-security-compliance.md)

---

**Last Updated**: February 2026  
**Version**: 2.0  
**Maintained By**: Platform Engineering Team
<!-- [<-- Back to Index](README.md) -->
<!---->
<!-- ## Business Rules & Validation -->
<!---->
<!-- ### Status Transition Rules -->
<!---->
<!-- ```markdown -->
<!-- VALID STATUS TRANSITIONS: -->
<!---->
<!-- From          To              Condition -->
<!-- ───────────────────────────────────────────────── -->
<!-- PENDING    → ACTIVE          Admin activation or auto-activate -->
<!-- ACTIVE     → SUSPENDED       Payment/policy/security reason required -->
<!-- ACTIVE     → ARCHIVED        Archive reason + retention required -->
<!-- SUSPENDED  → ACTIVE          Reactivation reason required -->
<!-- SUSPENDED  → ARCHIVED        Archive reason + retention required -->
<!-- ARCHIVED   → (none)          Terminal state - no transitions allowed -->
<!---->
<!-- INVALID TRANSITIONS (blocked): -->
<!-- PENDING    → SUSPENDED       Must activate first -->
<!-- PENDING    → ARCHIVED        Must activate first -->
<!-- ARCHIVED   → ACTIVE          Cannot reactivate archived tenant -->
<!-- ARCHIVED   → SUSPENDED       Already permanently deactivated -->
<!-- SUSPENDED  → PENDING         Cannot revert to pending -->
<!-- ``` -->
<!---->
<!-- ### Data Validation Rules -->
<!---->
<!-- ```markdown -->
<!-- TENANT CREATION RULES: -->
<!---->
<!-- name: -->
<!-- ├── Required -->
<!-- ├── Max 255 characters -->
<!-- ├── Must not be empty string -->
<!-- └── Trimmed of leading/trailing whitespace -->
<!---->
<!-- email: -->
<!-- ├── Required -->
<!-- ├── Must be valid email format -->
<!-- ├── Max 255 characters -->
<!-- └── Unique across non-deleted tenants -->
<!---->
<!-- subdomain: -->
<!-- ├── Optional -->
<!-- ├── Max 63 characters (DNS limit) -->
<!-- ├── Lowercase alphanumeric + hyphens only -->
<!-- ├── Cannot start or end with hyphen -->
<!-- ├── Unique across non-deleted tenants -->
<!-- └── Cannot be reserved word (admin, api, www, app, etc.) -->
<!---->
<!-- slug: -->
<!-- ├── Auto-generated from name -->
<!-- ├── Max 50 characters -->
<!-- ├── Lowercase alphanumeric + hyphens -->
<!-- ├── Unique across non-deleted tenants -->
<!-- └── UUID suffix appended if collision -->
<!---->
<!-- currency_code: -->
<!-- ├── 3 characters exactly -->
<!-- ├── ISO 4217 format -->
<!-- └── Default: USD -->
<!---->
<!-- timezone: -->
<!-- ├── Valid IANA timezone string -->
<!-- ├── e.g., Africa/Nairobi, UTC, America/New_York -->
<!-- └── Default: UTC -->
<!---->
<!-- company_size: -->
<!-- ├── Must be: Small, Medium, Large, Enterprise -->
<!-- └── Default: Small -->
<!---->
<!-- status: -->
<!-- ├── Must be: PENDING, ACTIVE, SUSPENDED, ARCHIVED -->
<!-- └── Default: PENDING (on creation) -->
<!-- ``` -->
<!---->
<!-- ### Resource Limit Rules -->
<!---->
<!-- ```markdown -->
<!-- LIMIT ENFORCEMENT RULES: -->
<!---->
<!-- 1. User Limit: -->
<!--    Rule: active_users < max_users -->
<!--    Checked: Before creating new user -->
<!--    Error: "Tenant has reached maximum user limit" -->
<!--    Override: Platform admin can temporarily increase -->
<!---->
<!-- 2. Entity Limit: -->
<!--    Rule: total_entities < max_entities -->
<!--    Checked: Before creating new company/branch -->
<!--    Error: "Tenant has reached maximum entity limit" -->
<!---->
<!-- 3. Transaction Limit: -->
<!--    Rule: total_transactions < max_transactions_per_month -->
<!--    Checked: Before posting any transaction -->
<!--    Error: "Monthly transaction limit reached" -->
<!--    Reset: First day of each month -->
<!---->
<!-- 4. Storage Limit: -->
<!--    Rule: storage_used_mb < storage_quota_mb -->
<!--    Checked: Before file upload -->
<!--    Error: "Storage quota exceeded" -->
<!--    Threshold: Warning at 80% utilization -->
<!---->
<!-- 5. API Rate Limit: -->
<!--    Rule: requests_in_window < rate_limit -->
<!--    Checked: Per request via middleware -->
<!--    Error: 429 Too Many Requests -->
<!--    Window: Per minute (configurable) -->
<!-- ``` -->
<!---->
<!-- ### Soft Delete Rules -->
<!---->
<!-- ```markdown -->
<!-- SOFT DELETE BEHAVIOR: -->
<!---->
<!-- On Delete: -->
<!-- ├── deleted_at = NOW() (not physical delete) -->
<!-- ├── All queries filter: WHERE deleted_at IS NULL -->
<!-- ├── Slug and subdomain become reusable (unique constraint scoped) -->
<!-- └── Audit log entry created -->
<!---->
<!-- Constraints: -->
<!-- ├── Unique slug:      WHERE deleted_at IS NULL -->
<!-- ├── Unique subdomain: WHERE deleted_at IS NULL -->
<!-- ├── Unique email:     WHERE deleted_at IS NULL -->
<!-- └── Cascade: Soft delete does NOT cascade to child tables -->
<!---->
<!-- Restoration: -->
<!-- ├── Set deleted_at = NULL -->
<!-- ├── Verify slug/subdomain still available -->
<!-- └── Requires platform admin action -->
<!-- ``` -->
<!---->
<!-- ### Configuration Rules -->
<!---->
<!-- ```markdown -->
<!-- CONFIGURATION VALIDATION: -->
<!---->
<!-- fiscal_year_start_month: -->
<!-- ├── Must be 1-12 -->
<!-- └── Cannot be changed mid-fiscal-year (warning) -->
<!---->
<!-- accounting_method: -->
<!-- ├── Must be: FIFO, LIFO, or WEIGHTED_AVERAGE -->
<!-- └── Changing method requires audit trail entry -->
<!---->
<!-- password_policy: -->
<!-- ├── min_length: Must be >= 8 -->
<!-- ├── max_age_days: Must be > 0 or null (no expiry) -->
<!-- └── At least one complexity rule recommended -->
<!---->
<!-- api_rate_limits: -->
<!-- ├── requests_per_minute: Must be > 0 -->
<!-- ├── Must not exceed plan maximum -->
<!-- └── Burst limit <= 2x per-minute rate -->
<!---->
<!-- allowed_modules: -->
<!-- ├── Must be valid module names -->
<!-- ├── "financial" and "selling" always included -->
<!-- └── Cannot enable modules above plan tier -->
<!-- ``` -->
<!---->
<!-- ### Isolation Rules -->
<!---->
<!-- ```markdown -->
<!-- TENANT ISOLATION INVARIANTS: -->
<!---->
<!-- 1. Every tenant-scoped table has RLS enabled -->
<!-- 2. Every query through application_role is filtered -->
<!-- 3. Foreign keys cannot reference cross-tenant records -->
<!-- 4. Tenant context must be set before any data access -->
<!-- 5. Tenant context is transaction-scoped (auto-cleared) -->
<!-- 6. Admin role bypasses RLS (platform admin only) -->
<!-- 7. Readonly role can see all tenants (monitoring only) -->
<!-- 8. Superuser access is prohibited in production -->
<!-- ``` -->
<!---->
<!-- --- -->
<!---->
<!-- Next: [Summary](./22-summary.md) -->
