[<-- Back to Index](README.md)

## Business Rules & Validation

> **Purpose**: This document explains the business rules and validation requirements for managing tenants in a multi-tenant system. Written for both technical and non-technical audiences, it describes what rules exist, why they matter, and how they protect your data.

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

### What Are Status Transitions?

A tenant's status represents its current state in the system. Think of it like a traffic light - you can only move from one color to another in specific ways. These rules prevent invalid state changes that could cause data problems or security issues.

### Why Do We Need Status Rules?

**Business Protection**: Prevents accidentally suspending or archiving active, paying customers  
**Data Integrity**: Ensures proper cleanup and retention policies are followed  
**Audit Trail**: Documents why and when status changes occurred  
**Security**: Prevents unauthorized reactivation of closed accounts

### The Four States Explained

**PENDING**: New tenant registration, awaiting activation
- Like a job application - submitted but not yet approved
- Limited or no system access
- Can be activated or rejected

**ACTIVE**: Fully operational tenant
- Normal business operations
- All features available
- Can be suspended or archived

**SUSPENDED**: Temporarily disabled
- Like a paused subscription
- No access to system
- Can be reactivated or archived
- Common reasons: payment issues, policy violations

**ARCHIVED**: Permanently closed
- Like a closed bank account
- Cannot be reactivated
- Data retained for compliance
- Final state - no further changes

### State Flow Diagram

```
    ┌─────────┐
    │ PENDING │ ← All new tenants start here
    └────┬────┘
         │
         │ activate (approval required)
         ▼
    ┌────────┐      suspend       ┌───────────┐
    │ ACTIVE │ ←──────────────────→│ SUSPENDED │
    └───┬────┘      reactivate    └─────┬─────┘
        │                               │
        │ archive                       │ archive
        ▼                               ▼
    ┌──────────┐                   ┌──────────┐
    │ ARCHIVED │ ← Final state - no way out
    └──────────┘                   └──────────┘
```

### Valid Transitions

| Current Status | Can Change To | What's Required | Real-World Example |
|---------------|---------------|-----------------|-------------------|
| PENDING | ACTIVE | • Administrator approval<br>• Complete registration<br>• Payment method (if needed) | A new company finishes signup and admin activates their account |
| ACTIVE | SUSPENDED | • Clear reason for suspension<br>• Documentation of issue<br>• Notification to tenant | Company's payment fails for 3 months in a row |
| ACTIVE | ARCHIVED | • Archive reason documented<br>• Data backup completed<br>• Customer notification sent | Company closes down and requests account deletion |
| SUSPENDED | ACTIVE | • Original problem resolved<br>• Verification completed<br>• Reactivation approved | Company updates payment method and pays outstanding balance |
| SUSPENDED | ARCHIVED | • Archive reason<br>• Retention rules applied | Suspended company never comes back after 6 months |

### What Transitions Are Blocked?

| Trying To Go From | To | Why It's Not Allowed | What To Do Instead |
|------------------|----|--------------------|-------------------|
| PENDING | SUSPENDED | Can't suspend what isn't active yet | Activate first, then suspend if needed |
| PENDING | ARCHIVED | Need to establish baseline first | Reject the application or delete it |
| ARCHIVED | ACTIVE | Once closed, stays closed | Customer must register as new tenant |
| ARCHIVED | SUSPENDED | Already closed | Not applicable |
| SUSPENDED | PENDING | Can't go backwards | Reactivate or archive |

### Why These Restrictions Matter

**Example Scenario**: Imagine a company (Acme Corp) that was ARCHIVED last year comes back and wants service again.

**❌ Wrong Approach**: Reactivating the ARCHIVED tenant
- Old data mixed with new
- Compliance issues with "closed" accounts
- Audit trail confusion
- Security risks

**✅ Correct Approach**: Create new tenant registration
- Clean separation of data
- Fresh start with current pricing
- Clear audit history
- Proper compliance tracking

### Technical Implementation (Minimal)

```go
// ValidTransitions defines what status changes are allowed
var ValidTransitions = map[string][]string{
    "PENDING":   {"ACTIVE"},
    "ACTIVE":    {"SUSPENDED", "ARCHIVED"},
    "SUSPENDED": {"ACTIVE", "ARCHIVED"},
    "ARCHIVED":  {}, // Terminal state - no transitions
}

// TransitionStatus changes tenant status with validation
func TransitionStatus(tenant *Tenant, newStatus, reason string) error {
    // Check if transition is valid
    allowed := ValidTransitions[tenant.Status]
    if !contains(allowed, newStatus) {
        return fmt.Errorf("cannot transition from %s to %s", 
            tenant.Status, newStatus)
    }
    
    // Require reason for suspension/archival
    if (newStatus == "SUSPENDED" || newStatus == "ARCHIVED") && reason == "" {
        return errors.New("reason required")
    }
    
    // Update tenant
    tenant.Status = newStatus
    tenant.StatusReason = reason
    tenant.StatusChangedAt = time.Now()
    
    return nil
}
```

---

## Data Validation Rules

### Why Validate Data?

Data validation ensures that information entering your system is:
- **Correct**: Meets basic format requirements (valid email, proper length)
- **Consistent**: Follows the same rules everywhere
- **Secure**: Prevents malicious input
- **Unique**: Avoids duplicate accounts when necessary

Think of it like a bouncer at a club - checking IDs, enforcing dress code, and preventing troublemakers from entering.

### Tenant Creation Fields

When creating a new tenant account, certain information is required and must meet specific criteria.

---

#### 1. Company Name

**Purpose**: The main identifier for the business

| Rule | Explanation | Example |
|------|-------------|---------|
| **Required** | Must provide a name | ✅ "Acme Corporation" |
| **Maximum Length** | Up to 255 characters | ❌ A 300-character name won't work |
| **Automatic Cleanup** | Extra spaces removed | "  Acme Corp  " becomes "Acme Corp" |
| **Special Characters** | Allowed | ✅ "Smith & Sons, Ltd." is fine |
| **Duplicates** | Allowed | ✅ Multiple companies can have same name |

**Why these rules?**
- Length limit: Database storage constraints
- Spaces trimmed: Prevents accidental duplicates from formatting
- Duplicates allowed: Many companies share common names (e.g., "ABC Trading")

---

#### 2. Email Address

**Purpose**: Primary contact and account recovery

| Rule | Explanation | Example |
|------|-------------|---------|
| **Required** | Must provide email | ✅ admin@acme.com |
| **Valid Format** | Must look like an email | ❌ "notanemail" won't work |
| **Maximum Length** | Up to 255 characters total | Including the "@domain.com" part |
| **Case Insensitive** | ADMIN@acme.com = admin@acme.com | Stored as lowercase |
| **Must Be Unique** | One email per active tenant | ❌ Can't reuse active email |

**Why these rules?**
- Format validation: Ensures we can actually send emails
- Uniqueness: Prevents account confusion and security issues
- Case insensitive: Most email systems treat case the same way

**What about deleted tenants?**  
If a tenant is deleted, their email becomes available again. This allows companies that previously closed accounts to return.

**Common Errors:**
- `"Email is required"` - You forgot to provide an email
- `"Invalid email format"` - Doesn't look like a proper email address  
- `"Email already in use"` - Another active tenant uses this email

---

#### 3. Subdomain (Optional)

**Purpose**: Custom web address for the tenant (e.g., `acme.yoursystem.com`)

| Rule | Explanation | Example |
|------|-------------|---------|
| **Optional** | Can be added later | null or "acme-corp" |
| **Maximum Length** | Up to 63 characters | Internet standard for domain names |
| **Format Rules** | Letters, numbers, hyphens only | ✅ "acme-123" |
| **Cannot Start/End with Hyphen** | Must have letter/number at edges | ❌ "-acme" or "acme-" |
| **Reserved Words** | Some names are off-limits | ❌ "admin", "api", "www" |
| **Must Be Unique** | One subdomain per tenant | ❌ Two tenants can't share subdomain |

**Reserved Subdomains (Cannot Use):**
```
admin, api, www, app, cdn, static, mail, ftp, 
dev, staging, test, dashboard, portal, auth, 
login, signup, billing, support, help, status
```

**Why these rules?**
- Length limit: Internet DNS standards (RFC 1035)
- Format rules: Ensures subdomain works in web browsers
- Reserved words: Protects system functionality and common use cases
- Uniqueness: Each tenant needs their own web address

**Examples:**
- ✅ "acme-corporation" - Good
- ✅ "acme123" - Good
- ❌ "Acme Corporation" - Spaces not allowed
- ❌ "admin" - Reserved word
- ❌ "my_company" - Underscores not allowed (use hyphens)

---

#### 4. Slug (Auto-Generated)

**Purpose**: URL-friendly version of company name

| Rule | Explanation | Example |
|------|-------------|---------|
| **Auto-Generated** | System creates this automatically | From "Acme Corporation" |
| **Cannot Set Manually** | Users don't choose this | System controlled |
| **Maximum Length** | Up to 50 characters | May be truncated |
| **Format** | Lowercase, letters, numbers, hyphens | "acme-corporation" |
| **Always Unique** | If collision, adds random suffix | "acme-corp-a1b2c3d4" |

**How It Works:**
1. Takes your company name: "Smith & Sons Ltd."
2. Converts to lowercase: "smith & sons ltd."
3. Replaces special characters with hyphens: "smith-sons-ltd"
4. Checks if already used
5. If duplicate, adds random code: "smith-sons-ltd-x7y8z9"

**More Examples:**
- "Acme Corporation" → `acme-corporation`
- "123 Trading Co." → `123-trading-co`
- "ABC-XYZ Company!!!" → `abc-xyz-company`

**Why automatic?**
- Consistency: Everyone follows the same pattern
- No conflicts: System handles duplicates automatically
- URL-safe: Always works in web addresses

---

#### 5. Currency Code

**Purpose**: Default currency for financial transactions

| Rule | Explanation | Example |
|------|-------------|---------|
| **Required** | Must specify currency | "USD" |
| **Exactly 3 Letters** | ISO standard format | ❌ "US" or "USDT" won't work |
| **Uppercase** | Always capital letters | USD, EUR, KES |
| **Default** | USD if not specified | Most common choice |
| **Cannot Change Later** | Locked after first transaction | Prevents financial chaos |

**Common Currencies:**
- USD - US Dollar
- EUR - Euro
- GBP - British Pound
- KES - Kenyan Shilling
- JPY - Japanese Yen
- CAD - Canadian Dollar

**Why can't you change currency?**  
Once you start creating invoices and recording transactions, changing currency would make historical data meaningless. Imagine if last month's $1,000 invoice suddenly showed as 1,000 Euros!

---

#### 6. Timezone

**Purpose**: Determines when "today" starts/ends for this tenant

| Rule | Explanation | Example |
|------|-------------|---------|
| **Required** | Must specify timezone | "Africa/Nairobi" |
| **Standard Format** | IANA timezone names | ✅ "America/New_York" |
| **Case Sensitive** | Must match exactly | ❌ "america/new_york" won't work |
| **Default** | UTC if not specified | Neutral choice |

**Common Timezones:**
- UTC - Universal Time (neutral)
- Africa/Nairobi - East Africa
- America/New_York - US Eastern Time
- Europe/London - UK Time
- Asia/Tokyo - Japan Time

**Why not "GMT+3" or "EST"?**  
Abbreviated codes don't account for daylight saving time. "America/New_York" automatically switches between EST and EDT.

**Why does this matter?**  
Your timezone affects:
- When daily reports are generated
- When "today's" transactions start/end
- When scheduled jobs run
- What date/time users see in the interface

---

#### 7. Company Size (Optional)

**Purpose**: Business analytics and feature recommendations

| Value | Typical Definition |
|-------|-------------------|
| **Small** | 1-50 employees (default) |
| **Medium** | 51-250 employees |
| **Large** | 251-1000 employees |
| **Enterprise** | 1000+ employees |

**Note**: This is optional and used for analytics only. Not enforced - it's based on your self-reporting.

---

#### 8. Status

**Purpose**: Current state of the tenant account

| Rule | Explanation |
|------|-------------|
| **Required** | Must have a status |
| **Valid Values** | PENDING, ACTIVE, SUSPENDED, ARCHIVED |
| **Default** | PENDING for new accounts |
| **Must Follow Transition Rules** | Can't jump between states randomly |

See [Status Transition Rules](#status-transition-rules) section for complete details.

---

### Technical Implementation (Minimal)

```go
// TenantValidation contains validation rules
type TenantValidation struct{}

// ValidateName checks company name
func (v *TenantValidation) ValidateName(name string) error {
    name = strings.TrimSpace(name)
    
    if name == "" {
        return errors.New("company name is required")
    }
    if len(name) > 255 {
        return errors.New("company name must be under 255 characters")
    }
    return nil
}

// ValidateEmail checks email format and uniqueness
func (v *TenantValidation) ValidateEmail(email string) error {
    email = strings.ToLower(strings.TrimSpace(email))
    
    if email == "" {
        return errors.New("email is required")
    }
    if !isValidEmail(email) {
        return errors.New("invalid email format")
    }
    if emailExists(email) {
        return errors.New("email already in use")
    }
    return nil
}

// ValidateSubdomain checks subdomain format
func (v *TenantValidation) ValidateSubdomain(subdomain string) error {
    if subdomain == "" {
        return nil // Optional field
    }
    
    subdomain = strings.ToLower(subdomain)
    
    if len(subdomain) > 63 {
        return errors.New("subdomain too long (max 63 characters)")
    }
    
    // Check format: lowercase letters, numbers, hyphens only
    matched, _ := regexp.MatchString(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`, subdomain)
    if !matched {
        return errors.New("invalid subdomain format")
    }
    
    if isReservedSubdomain(subdomain) {
        return errors.New("subdomain is reserved")
    }
    
    return nil
}
```

---

## Resource Limit Rules

### What Are Resource Limits?

Resource limits are like the capacity of a restaurant - there's a maximum number of tables, dishes per hour, and storage space. These limits:
- **Prevent System Overload**: Keep the platform fast for everyone
- **Ensure Fair Usage**: No single tenant monopolizes resources
- **Support Business Tiers**: Different plans offer different capacities
- **Maintain Performance**: System runs smoothly for all users

### The Five Types of Limits

---

#### 1. User Limit

**What It Means**: Maximum number of people who can have accounts

**How It Works**:
- Plan determines maximum (e.g., 10 users on Starter, 50 on Professional)
- Checked before creating each new user account
- Only counts ACTIVE users (not disabled/archived ones)
- Admins can temporarily override in special cases

**Real-World Example**:
```
Your Plan: Professional (50 users)
Current Usage: 47 active users
Status: 3 slots remaining

Trying to add 51st user → ❌ Blocked
"Tenant has reached maximum user limit (50)"
```

**What You See in Dashboard**:
```
Users: 47 / 50  [█████████-] 94% ⚠️ Warning
```

**Solutions When You Hit the Limit**:
1. **Deactivate unused accounts**: Remove people who left company
2. **Upgrade plan**: Get more user slots
3. **Contact support**: Temporary increase for special situations

**Typical Limits by Plan**:
- Starter: 5-10 users
- Professional: 25-50 users
- Enterprise: 100+ users

---

#### 2. Entity Limit

**What It Means**: Maximum number of organizational units (companies, branches, departments)

**Why This Matters**:
Each entity adds complexity to your setup. Too many can make the system hard to manage and slower to use.

**What Counts as an Entity**:
- Companies or business units
- Branch locations
- Departments
- Cost centers
- Divisions

**Real-World Example**:
```
Your Plan: Professional (25 entities)
Current: 5 companies, 12 branches, 6 departments = 23 total
Status: 2 slots remaining

Adding new branch → ✅ Allowed (24/25)
Adding another department → ✅ Allowed (25/25)
Adding one more anything → ❌ Blocked
```

**Solutions When You Hit the Limit**:
1. **Remove unused entities**: Clean up old branches/departments
2. **Consolidate**: Merge similar entities
3. **Upgrade plan**: Get higher limit

**Typical Limits**:
- Starter: 5 entities
- Professional: 25 entities
- Enterprise: 100+ entities

---

#### 3. Transaction Limit

**What It Means**: Maximum financial transactions per month

**What Counts as a Transaction**:
- Creating an invoice
- Recording a payment
- Journal entries
- Purchase orders
- Credit notes
- Any financial record

**How It Works**:
- Limit resets on 1st day of each month (in your timezone)
- Checked before posting each transaction
- 10% grace period at month-end (helps with closing books)
- Overages discussed with support team

**Real-World Example**:
```
Month: February 2026
Your Plan: Professional (5,000 transactions/month)
Current: 4,890 transactions
Status: 110 remaining

Posting new invoice → ✅ Allowed (4,891/5,000)
At 5,001 transactions → ❌ Blocked until March 1
"Monthly transaction limit reached"
```

**What You See**:
```
Transactions: 4,245 / 5,000  [████████--] 85%
Resets: March 1, 2026
```

**Why This Limit Exists**:
- High transaction volume impacts database performance
- Different plans = different infrastructure costs
- Prevents accidental bulk imports from overwhelming system

**Typical Limits**:
- Starter: 500 transactions/month
- Professional: 5,000 transactions/month
- Enterprise: Unlimited

---

#### 4. Storage Limit

**What It Means**: Total file storage space available

**What Uses Storage**:
- Document attachments (PDFs, images, etc.)
- Exported reports
- Invoice copies
- Email attachments saved to system
- Backup files

**How It Works**:
- Measured in MB or GB
- Checked before every file upload
- Warning at 80% full
- Alert at 90% full
- Blocked at 100% (admin can override)

**Real-World Example**:
```
Your Plan: Professional (2 GB storage)
Current Usage: 1.7 GB
Status: 300 MB remaining

Uploading 50 MB file → ✅ Allowed (1.75 GB used)
Uploading 400 MB file → ❌ Blocked
"Storage quota exceeded"
```

**Dashboard View**:
```
Storage: 1.7 GB / 2.0 GB  [████████--] 85% ⚠️ Warning
```

**Automatic Cleanup**:
- Exported reports deleted after 30 days
- Attachments over 1 MB automatically compressed
- System suggests archiving old documents

**Solutions When Storage is Full**:
1. **Delete old exports**: Remove downloaded reports
2. **Remove unused attachments**: Clean up unnecessary files
3. **Archive old documents**: Move to external storage
4. **Upgrade plan**: Get more storage space

**Typical Limits**:
- Starter: 500 MB - 1 GB
- Professional: 2-5 GB
- Enterprise: 10+ GB

---

#### 5. API Rate Limit

**What It Means**: Maximum API requests per minute

**Who This Affects**:
- Developers using the API
- Third-party integrations
- Mobile apps
- Automated processes

**How It Works**:
- Counted per minute (sliding window)
- 2× burst allowed for 10 seconds (handles brief spikes)
- Returns error code 429 when exceeded
- Tells you how long to wait before retrying

**Real-World Example**:
```
Your Plan: Professional (600 requests/minute)
Current: 550 requests in last 60 seconds
Status: 50 requests remaining

Next API call → ✅ Allowed
Making 100 calls rapidly → First 50 succeed, next 50 blocked

Error Response:
"Rate limit exceeded: 600 req/min"
"Retry after: 45 seconds"
```

**Response You Get**:
```
HTTP 429 Too Many Requests
X-RateLimit-Limit: 600
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1643472120
Retry-After: 45
```

**Why This Limit Exists**:
- Prevents accidental infinite loops from taking down system
- Protects against abuse
- Ensures fair resource sharing
- Maintains system responsiveness

**Solutions When You Hit Rate Limit**:
1. **Batch requests**: Combine multiple operations
2. **Add delays**: Space out API calls
3. **Use webhooks**: Get notifications instead of polling
4. **Upgrade plan**: Higher rate limits available

**Typical Limits**:
- Free/Starter: 60 requests/minute
- Professional: 600 requests/minute
- Enterprise: 6,000+ requests/minute

---

### Dashboard Summary

When you log in as admin, you see all limits at a glance:

```
Resource Usage - Acme Corporation
Plan: Professional

Users:         47 / 50       [█████████-] 94%  ⚠️ Near limit
Entities:      23 / 25       [█████████-] 92%  ⚠️ Near limit
Transactions:  3,245 / 5,000 [██████----] 65%  Resets: Mar 1
Storage:       850 MB / 2 GB [████------] 42%  
API Rate:      45 / 600/min  [█---------] 8%   Current minute
```

**Color Coding**:
-  Green (0-79%): Healthy usage
-  Yellow (80-95%): Warning - consider action
-  Red (96-100%): Critical - action required

---

### Technical Implementation (Minimal)

```go
// CheckUserLimit verifies before creating user
func CheckUserLimit(tenantID int) error {
    activeCount := countActiveUsers(tenantID)
    maxUsers := getMaxUsers(tenantID)
    
    if activeCount >= maxUsers {
        return fmt.Errorf(
            "tenant has reached maximum user limit (%d). "+
            "currently: %d active users. "+
            "contact support to upgrade", 
            maxUsers, activeCount,
        )
    }
    return nil
}

// CheckStorageLimit verifies before file upload
func CheckStorageLimit(tenantID int, fileSizeMB float64) error {
    usedMB := calculateStorageUsed(tenantID)
    quotaMB := getStorageQuota(tenantID)
    
    if (usedMB + fileSizeMB) > quotaMB {
        return fmt.Errorf(
            "storage quota exceeded. "+
            "used: %.1f MB / %.0f MB. "+
            "this file: %.1f MB", 
            usedMB, quotaMB, fileSizeMB,
        )
    }
    
    // Warning at 80%
    if usedMB > (quotaMB * 0.8) {
        notifyStorageWarning(tenantID, usedMB, quotaMB)
    }
    
    return nil
}

// CheckRateLimit enforces API rate limiting
func CheckRateLimit(tenantID int) error {
    key := fmt.Sprintf("rate_limit:%d:%s", tenantID, getCurrentMinute())
    count := redis.Incr(key)
    
    if count == 1 {
        redis.Expire(key, 60) // 1-minute window
    }
    
    maxRate := getRateLimit(tenantID)
    if count > maxRate {
        ttl := redis.TTL(key)
        return &RateLimitError{
            Limit:      maxRate,
            RetryAfter: ttl,
        }
    }
    
    return nil
}
```

---

## Soft Delete Rules

### What Is Soft Delete?

**Soft delete** means marking data as deleted without actually removing it from the database. Think of it like moving a file to the Recycle Bin instead of permanently deleting it.

**Real-World Analogy**:
- **Soft Delete**: Employee leaves company → mark as "inactive", keep records
- **Hard Delete**: Permanently erase from existence → no recovery possible

### Why Not Just Delete Everything?

**Legal & Compliance**: Many regulations require keeping records for years  
**Audit Trail**: Need history of who had access and what they did  
**Data Recovery**: Mistakes happen - soft delete allows undo  
**Business Intelligence**: Historical data valuable for analytics  
**Dispute Resolution**: May need old records for legal/financial disputes

### How It Works

**The Deleted Timestamp**:
```
Normal record:  deleted_at = NULL (visible to everyone)
Soft deleted:   deleted_at = "2026-01-15 14:30:00" (hidden from normal queries)
```

When you "delete" a tenant:
1. System adds deletion timestamp
2. Record stays in database
3. Normal queries can't see it anymore
4. Admins can still view/restore it
5. After retention period, permanently removed

### The Deletion Process

**Step-by-Step**:

1. **Request Deletion**: Admin initiates tenant deletion
2. **Validation**: System checks tenant is ARCHIVED (can't delete active tenants)
3. **Mark as Deleted**: Set `deleted_at` timestamp to current time
4. **Record Reason**: Document why deletion occurred
5. **Audit Log**: Create permanent record of deletion
6. **Schedule Cleanup**: Set timer for permanent removal (90 days)

**What Happens to Data**:
- Tenant record: Marked deleted, still in database
- Users: Remain in database, blocked by tenant context
- Invoices/Transactions: Remain in database, blocked by tenant context
- Files/Documents: Remain in storage, inaccessible
- External systems: Notified of deletion (webhooks)

### Reusing Information

When a tenant is soft deleted, certain unique fields become available again:

**Reusable After Soft Delete**:
- ✅ Email address
- ✅ Subdomain
- ✅ Slug

**Example**:
```
1. Company "Acme Corp" creates account with email admin@acme.com
2. Later, they delete their account (soft delete)
3. New company can now use admin@acme.com
4. Old Acme Corp data still exists but is hidden
```

**Why This Works**:
- Unique constraints only check non-deleted records
- Prevents blocking future customers
- Old data preserved for compliance
- No conflicts between old and new

### Child Records (Non-Cascading)

**Important**: Deleting a tenant does NOT automatically delete related records.

**What This Means**:
```
Tenant deleted → Users, invoices, transactions still in database
Access prevented by → Application-layer security (tenant context)
```

**Why Not Auto-Delete Everything?**
- Regulatory compliance requires keeping transaction history
- Audit trail must remain intact
- Investigations may need old data
- Hard delete handles permanent removal later

### The Restoration Process

Soft-deleted tenants can be restored within the retention period.

**Requirements for Restoration**:
1. Within 90-day retention window
2. Original slug/subdomain still available
3. Platform admin approval
4. Valid business reason

**Restoration Steps**:
1. Admin requests restoration
2. System checks availability of slug/subdomain
3. System verifies within retention period
4. Admin provides reason for restoration
5. Deletion timestamp removed
6. Tenant accessible again
7. Audit log records restoration

**Restoration Timeline**:

| Days Since Deletion | Can Restore? | Who Can Authorize |
|-------------------|--------------|-------------------|
| 0-30 days | ✅ Yes | Support team can quickly restore |
| 30-60 days | ✅ Yes | Requires manager approval |
| 60-90 days | ⚠️ Maybe | Requires executive approval |
| 90+ days | ❌ No | Data permanently deleted |

### Permanent Deletion (Hard Delete)

After the retention period, data is permanently removed.

**Hard Delete Process**:
1. 90 days have passed since soft delete
2. Automated job runs monthly
3. Deletes all tenant data:
   - Users and permissions
   - Financial transactions
   - Documents and files
   - All related records
4. Tenant record physically deleted from database
5. Final audit log entry created
6. **No recovery possible**

**Why 90 Days?**
- Industry standard for data retention
- Balances compliance with storage costs
- Gives reasonable time for recovery requests
- Allows investigation of disputed deletions

### Technical Implementation (Minimal)

```go
// SoftDeleteTenant marks tenant as deleted
func SoftDeleteTenant(tenantID int, reason string) error {
    tenant := findTenant(tenantID)
    
    // Must be archived first
    if tenant.Status != "ARCHIVED" {
        return errors.New("tenant must be archived before deletion")
    }
    
    // Mark as deleted
    tenant.DeletedAt = time.Now()
    tenant.DeletedBy = getCurrentUser()
    tenant.DeletionReason = reason
    
    // Save and audit
    saveTenant(tenant)
    createAuditLog("tenant_deleted", tenant)
    
    // Schedule permanent deletion after 90 days
    scheduleHardDelete(tenantID, 90)
    
    return nil
}

// RestoreTenant brings back soft-deleted tenant
func RestoreTenant(tenantID int) error {
    tenant := findDeletedTenant(tenantID)
    
    // Check retention period
    daysSince := time.Since(tenant.DeletedAt).Hours() / 24
    if daysSince > 90 {
        return errors.New("beyond 90-day restoration period")
    }
    
    // Check if slug/subdomain available
    if slugInUse(tenant.Slug) {
        return errors.New("slug no longer available")
    }
    
    // Restore
    tenant.DeletedAt = nil
    tenant.DeletedBy = nil
    tenant.RestoredAt = time.Now()
    saveTenant(tenant)
    
    return nil
}
```

### Business Impact

**Storage Costs**: Soft-deleted data uses storage space  
**Compliance**: Helps meet regulatory requirements  
**Customer Service**: Allows fixing accidental deletions  
**Data Safety**: Extra layer of protection against mistakes

---

## Configuration Rules

### What Are Configuration Rules?

Configuration settings control how your tenant's accounting and security systems work. Unlike data that changes frequently (like invoices), these are structural settings that should remain stable once set.

---

### 1. Fiscal Year Start

**What It Is**: The month when your financial year begins

**Why It Matters**:
- Determines reporting periods
- Affects when "Year-to-Date" calculations start
- Impacts financial statement generation
- Must align with tax reporting requirements

**Common Configurations**:

| Configuration | Start Month | Used By |
|--------------|-------------|---------|
| Calendar Year | January (1) | Most businesses worldwide |
| UK Tax Year | April (4) | UK companies |
| US Federal Government | October (10) | US government agencies |
| Australia Tax Year | July (7) | Australian companies |

**Can You Change It?**

Yes, but with caution:
- ✅ Easy to change before fiscal year starts
- ⚠️ Risky to change mid-year (affects reports)
- ❌ Very problematic to change after year-end closing

**Example Impact**:
```
Current Setting: Fiscal year starts April 1
Current Date: August 15, 2025

Trying to change to January 1:
→ System warns: "This will affect 5 months of existing reports"
→ Recommendation: Wait until March 31, 2026
```

---

### 2. Accounting Method

**What It Is**: How you calculate the cost of inventory sold

**The Three Methods**:

**FIFO (First-In, First-Out)**
- Sells oldest inventory first
- Like a grocery store - old milk sells before new milk
- **Best for**: Perishable goods, fashion, anything that expires
- **Tax impact**: Higher taxes during inflation (selling cheaper old stock)

**LIFO (Last-In, First-Out)**
- Sells newest inventory first
- Like a coal pile - use what's on top
- **Best for**: Non-perishable commodities
- **Tax impact**: Lower taxes during inflation (selling expensive new stock)
- **Note**: Not allowed in many countries (including under IFRS)

**WEIGHTED AVERAGE**
- Averages cost of all inventory
- Like mixing paint - everything blends together
- **Best for**: Bulk commodities, interchangeable goods
- **Tax impact**: Moderate, predictable

**Why You Can't Easily Change**:
```
Example: Your company has 1,000 items in stock

Using FIFO:
- Oldest 500 items cost $5 each = $2,500
- Newest 500 items cost $8 each = $4,000
- Total inventory value = $6,500

Switching to WEIGHTED AVERAGE:
- All 1,000 items now cost $6.50 each
- Total inventory value = $6,500 (same)
- BUT: Cost of next sale changes!

This affects:
✗ All historical reports
✗ Tax filings
✗ Financial statements
✗ Profit calculations
```

**Changing Methods Requires**:
1. Valid business reason (must document)
2. Recalculation of all inventory values
3. Restatement of financial reports
4. Accountant approval recommended
5. Tax authority notification (in some regions)

---

### 3. Password Policy

**What It Controls**: Security requirements for user passwords

**Configurable Settings**:

| Setting | Default | Purpose |
|---------|---------|---------|
| Minimum Length | 8 characters | Harder to guess |
| Require Uppercase | Yes | Increases complexity |
| Require Lowercase | Yes | Increases complexity |
| Require Numbers | Yes | Harder to crack |
| Require Special Characters | Yes | Maximum security |
| Password Expiry | 90 days | Regular rotation |
| Prevent Reuse | Last 5 passwords | Can't cycle old passwords |
| Lockout Attempts | 5 failed tries | Prevents brute force |
| Lockout Duration | 30 minutes | Balance security/usability |

**Minimum Requirements**:
- At least 8 characters long
- At least ONE complexity requirement (uppercase/lowercase/number/special)

**Example Strong Password**:
```
✅ "MyP@ssw0rd2026" - Has uppercase, lowercase, numbers, special chars
✅ "Coffee&Tea#42" - Easy to remember, meets all requirements
❌ "password" - Too simple
❌ "12345678" - No letters
❌ "Password" - No numbers or special chars
```

**Real-World Example**:
```
Your Policy:
- 10 characters minimum
- Must have uppercase, lowercase, number
- Expires every 90 days
- Cannot reuse last 5 passwords

User tries "welcome123":
❌ Only 10 characters but no uppercase
❌ Rejected: "Password must contain uppercase letter"

User tries "Welcome123":
✅ Accepted! Meets all requirements
```

---

### 4. API Rate Limits

**What They Control**: How many API requests tenants can make

**Why Configurable**:
- Different plans = different needs
- Prevents system abuse
- Allows burst traffic
- Fair resource allocation

**Configuration Options**:

| Setting | Description | Typical Value |
|---------|-------------|---------------|
| Requests Per Minute | Base rate limit | 60-6,000 depending on plan |
| Burst Multiplier | Short spike allowance | 2× (allows double for 10 seconds) |
| Burst Duration | How long burst lasts | 10 seconds |
| Apply to Webhooks | Include webhook calls | Usually No (webhooks exempt) |

**Why Burst Matters**:
```
Normal: 600 requests/minute allowed

Without Burst:
User makes 700 requests in 10 seconds → 100 rejected immediately

With 2× Burst (10 seconds):
User makes 700 requests in 10 seconds → All accepted
System smooths out the spike
```

**Constraints**:
- ✅ Can increase up to plan maximum
- ❌ Cannot exceed plan limits (upgrade required)
- ⚠️ Burst > 3× may impact system performance
- ⚠️ Burst duration > 60 seconds defeats the purpose

---

### 5. Allowed Modules

**What They Are**: Features/functionality available to tenant

**Module Categories**:

**Core Modules** (Always Enabled):
- Financial Accounting
- Sales/Selling

**Optional Modules** (Plan-Dependent):
- Buying/Purchasing
- Inventory Management
- Manufacturing
- Project Management
- Human Resources
- Payroll
- CRM (Customer Relationship Management)
- Help Desk/Support

**Plan Comparison**:

| Plan | Available Modules |
|------|------------------|
| **Starter** | Financial, Selling only |
| **Professional** | + Buying, Inventory, CRM |
| **Enterprise** | All modules available |

**Why Restrict Modules?**
- Simpler interface for smaller businesses
- Lower infrastructure costs for basic plans
- Encourages upgrades as businesses grow
- Each module adds complexity and support burden

**Enabling New Module**:
```
Current Plan: Professional
Current Modules: Financial, Selling, Buying, Inventory

Request: Enable "Payroll" module
↓
Check: Is Payroll in Professional plan?
Result: ❌ No - Payroll only in Enterprise
Message: "Upgrade to Enterprise to access Payroll"
```

**Cannot Disable Core Modules**:
```
Request: Disable "Financial" module
Result: ❌ Blocked
Reason: "Financial module is core functionality"
```

---

### Technical Reference (Minimal)

```go
// ConfigValidation validates configuration changes
type ConfigValidation struct{}

// ValidateFiscalYear checks fiscal year month
func (v *ConfigValidation) ValidateFiscalYear(month int) error {
    if month < 1 || month > 12 {
        return errors.New("month must be between 1-12")
    }
    
    // Warn if mid-fiscal-year
    if isMiddleOfFiscalYear() {
        log.Warn("Changing fiscal year mid-period may affect reporting")
    }
    
    return nil
}

// ValidateAccountingMethod checks method change
func (v *ConfigValidation) ValidateAccountingMethod(method string) error {
    validMethods := []string{"FIFO", "LIFO", "WEIGHTED_AVERAGE"}
    
    if !contains(validMethods, method) {
        return errors.New("invalid accounting method")
    }
    
    // Require audit trail
    if requiresReasonForChange() {
        return errors.New("reason required for method change")
    }
    
    return nil
}
```

---

## Tenant Isolation Rules

### What Is Tenant Isolation?

**The Core Principle**: Your data must NEVER be visible to other tenants.

**Real-World Analogy**:  
Think of a multi-tenant system like an apartment building:
- Each tenant (company) has their own apartment (data space)
- You have your own key (authentication)
- You can't see your neighbor's mail or enter their apartment
- The building manager (platform admin) can access all units if needed
- Each apartment's contents are completely isolated

### Why Isolation Matters

**Security**: Prevents data breaches between customers  
**Privacy**: Keeps business information confidential  
**Compliance**: Required by regulations like GDPR, HIPAA  
**Trust**: Customers must know their data is safe  
**Legal**: Breaches can result in lawsuits and fines

**What Could Go Wrong Without It**:
```
❌ Company A sees Company B's invoices
❌ Company A can modify Company B's data
❌ Company A's employee list includes Company B's staff
❌ Financial reports mixed between companies
```

### The Eight Security Rules

These rules work together to create perfect isolation. They're listed from highest-level (business logic) to lowest-level (database).

---

#### 1. Row-Level Security (Database Level)

**What It Is**: The database itself blocks cross-tenant queries

**How It Works**:
- Every data row has a `tenant_id` field
- Database automatically filters all queries
- Only sees rows matching current tenant
- Happens at database level (can't be bypassed)

**Example**:
```
Invoices Table:
ID | Tenant ID | Amount | Customer
1  | 100       | $500   | Acme Corp
2  | 101       | $300   | Beta LLC
3  | 100       | $200   | Acme Corp

User from Tenant 100 queries invoices:
→ Database automatically adds: WHERE tenant_id = 100
→ User only sees invoices #1 and #3
→ Invoice #2 is invisible (different tenant)
```

**Which Tables Need This**:
- ✅ All business data (invoices, customers, products, users)
- ❌ System tables (tenant list itself, global settings)

---

#### 2. Application Role Filtering

**What It Is**: Database connection automatically knows which tenant

**How It Works**:
1. User logs in to system
2. System determines their tenant (from subdomain, JWT, etc.)
3. Database connection tagged with tenant ID
4. All queries automatically filtered
5. Connection cleared after each request

**Why This Matters**:
- Developers don't have to remember to filter
- Automatic protection against mistakes
- Consistent across entire application
- Single point of security enforcement

**Real-World Example**:
```
User: john@acme.com logs in
System: Determines tenant_id = 100
Database: Sets connection context to tenant 100
John queries: SELECT * FROM invoices
Database returns: Only tenant 100's invoices
Connection closes: Context cleared automatically
```

---

#### 3. Foreign Key Constraints

**What It Is**: Database prevents linking to other tenants' data

**The Problem Without It**:
```
Invoice for Tenant A → Customer from Tenant B
This should NEVER be possible!
```

**The Solution**:
- Foreign keys include tenant_id
- Database enforces same tenant
- Impossible to create cross-tenant relationships

**Example**:
```
Creating Invoice:
- Tenant ID: 100
- Customer ID: 5

Database checks:
→ Does Customer #5 belong to Tenant 100?
→ Yes: ✅ Invoice created
→ No: ❌ Error: "Customer not found"
```

---

#### 4. Mandatory Tenant Context

**What It Is**: System must know which tenant before any data access

**How It Works**:
- Every request must have tenant identified
- No queries allowed without tenant context
- System rejects requests lacking tenant info
- Protects against programming errors

**What You See**:
```
Trying to access data without logging in:
❌ Error: "No tenant context"
→ Must authenticate first

Tenant is SUSPENDED:
❌ Error: "Tenant Acme Corp is SUSPENDED"
→ Cannot access system
```

**Why This Rule Exists**:
- Prevents accidental queries across all tenants
- Forces proper authentication
- Catches programming mistakes early
- Ensures every operation has accountability

---

#### 5. Transaction-Scoped Context

**What It Is**: Tenant context automatically cleared after each operation

**Why This Matters**:
- Prevents context from "leaking" to next request
- Connection pooling doesn't mix tenants
- Each request starts fresh
- Eliminates a whole class of security bugs

**Technical Example**:
```
Request 1: User from Tenant A
→ Context set to Tenant A
→ Query executes for Tenant A
→ Request ends
→ Context AUTOMATICALLY cleared

Request 2: User from Tenant B
→ Context set to Tenant B (fresh start)
→ No pollution from Request 1
→ Query executes for Tenant B only
```

---

#### 6. Admin Role Bypass

**What It Is**: Platform administrators can see all tenants

**Who Has This**:
- Platform operators (your company's staff)
- Support engineers (for customer assistance)
- Security team (for investigations)

**NOT available to**:
- Regular users
- Tenant administrators
- Any tenant employees

**When It's Used**:
- Customer support requests
- System monitoring
- Security investigations
- Data migration tasks
- Platform maintenance

**Audit Trail**:
```
Every admin access is logged:
- Who accessed (admin name)
- When (timestamp)
- Which tenant (tenant ID)
- Why (reason/ticket number)
- What they did (actions taken)
```

---

#### 7. Readonly Role

**What It Is**: Monitoring systems can view (but not change) all tenant data

**Purpose**:
- System health monitoring
- Generating platform-wide reports
- Performance analytics
- Capacity planning
- No modification allowed

**Example Usage**:
```
Monitoring Dashboard:
→ Total active tenants: 1,250
→ Total invoices today: 5,430
→ System storage used: 2.3 TB
→ Average response time: 120ms

This requires seeing across all tenants
But cannot modify any data
```

---

#### 8. Superuser Prohibition

**What It Is**: Database superuser access is blocked in production

**Why This Rule**:
- Superuser bypasses ALL security
- Too dangerous for production
- Violates isolation principles
- Creates audit trail gaps

**Allowed**:
- ✅ Development environment
- ✅ Testing environment
- ✅ Local development

**Blocked**:
- ❌ Production environment
- ❌ Staging with real data
- ❌ Any customer-facing system

**Monitoring**:
```
System checks every minute:
"Are there any superuser connections?"
↓
If YES → Alert security team immediately
"CRITICAL: Superuser detected in production!"
```

---

### How Isolation Is Tested

**Regular Security Tests**:
1. Create two test tenants (A and B)
2. Create data for each tenant
3. Log in as Tenant A user
4. Try to access Tenant B data
5. Verify: Access denied

**Example Test**:
```
Setup:
- Tenant A: Create invoice #123 for $500
- Tenant B: Create invoice #456 for $300

Test 1: Query all invoices as Tenant A
Result: Only see invoice #123 ✅

Test 2: Query specific invoice #456 as Tenant A
Result: "Invoice not found" ✅

Test 3: Direct database ID access to invoice #456 as Tenant A
Result: Blocked by row-level security ✅

All tests must pass before deployment
```

---

### What Happens If Isolation Fails?

**Immediate Actions**:
1. System automatically logs the incident
2. Security team alerted
3. Affected tenants notified
4. Incident investigation started
5. Audit of all access during time window
6. Remediation plan created

**Customer Communication**:
- Transparent notification
- Explanation of what happened
- What data was potentially exposed
- What steps are being taken
- Follow-up timeline

---

### Technical Implementation (Minimal)

```go
// SetTenantContext configures database connection for tenant
func SetTenantContext(ctx context.Context, tenantID int) error {
    // Set tenant ID in database session
    _, err := db.Exec(ctx, "SET app.tenant_id = $1", tenantID)
    if err != nil {
        return fmt.Errorf("failed to set tenant context: %w", err)
    }
    
    // Switch to application role (has RLS enabled)
    _, err = db.Exec(ctx, "SET ROLE application_role")
    if err != nil {
        return fmt.Errorf("failed to set role: %w", err)
    }
    
    return nil
}

// ClearTenantContext removes tenant from connection
func ClearTenantContext(ctx context.Context) {
    db.Exec(ctx, "RESET app.tenant_id")
    db.Exec(ctx, "RESET ROLE")
}

// TenantMiddleware ensures tenant context for all requests
func TenantMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract tenant from request
        tenant, err := extractTenantFromRequest(r)
        if err != nil {
            http.Error(w, "No tenant context", http.StatusUnauthorized)
            return
        }
        
        // Set context for this request
        ctx := r.Context()
        SetTenantContext(ctx, tenant.ID)
        defer ClearTenantContext(ctx)
        
        // Continue to next handler
        next.ServeHTTP(w, r)
    })
}
```

---

### Business Impact

**Trust**: Customers confident their data is safe  
**Compliance**: Meets regulatory requirements  
**Liability**: Reduces legal risk  
**Reputation**: Protects company brand  
**Sales**: Security is a competitive advantage

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

### Common Problems and Solutions

---

### Problem 1: "Email already exists"

**What happened**: Trying to create a new tenant with an email that's already in use

**Why it happens**:
- Another active tenant is using this email
- Email must be unique across all active tenants
- Prevents duplicate accounts and confusion

**Solutions**:

**Option A - Use Different Email**
```
Current: admin@acme.com (taken)
Try: admin+new@acme.com
Or: contact@acme.com
```

**Option B - Check If It's Your Old Account**
1. Search for existing tenant with that email
2. If found and soft-deleted → Can be restored
3. If found and active → Contact support to merge/transfer
4. If same organization → Restore old account instead of creating new

**Option C - Add Subdomain to Distinguish**
```
tenant1@acme.com → acme-east.yoursystem.com
tenant2@acme.com → acme-west.yoursystem.com
Different subdomains = different tenants = OK
```

---

### Problem 2: "Maximum user limit reached"

**What happened**: Trying to add a new user but plan limit is full

**Current situation example**:
```
Your Plan: Professional (50 users)
Active Users: 50
Trying to add: User #51
Result: ❌ Blocked
```

**Solutions (in order of preference)**:

**1. Remove Inactive Users** (Free)
- Review user list
- Deactivate employees who left
- Archive unused accounts
- This frees up slots immediately

**2. Upgrade Your Plan** (Recommended)
- Professional → Enterprise
- Get more user slots
- Usually more features too

**3. Temporary Override** (Special cases)
- Contact platform support
- Explain the situation
- May get temporary increase
- Usually for short-term needs

**Checking Current Usage**:
```
Dashboard → Settings → Users
Shows: "47 of 50 users (94% used)"
Click to see full list of active users
```

---

### Problem 3: "Cannot transition from ARCHIVED to ACTIVE"

**What happened**: Trying to reactivate an archived tenant

**Why it's blocked**:
- ARCHIVED is a **terminal state**
- Like a closed bank account
- Cannot be reopened (by design)
- Prevents data confusion

**What to do instead**:

**If Customer Wants to Return**:
1. Create NEW tenant registration
2. Fresh start with current pricing
3. Clean separation of data
4. Better audit trail

**If Need Old Data**:
1. Data retained for 90 days after archival
2. Can export old data before new setup
3. Support can assist with data transfer
4. Keep old and new accounts separate

**If Archived by Mistake** (Within 90 days):
1. Check if soft-deleted (not archived)
2. Soft-deleted can be restored
3. Archived cannot be reversed
4. Contact support immediately if urgent

**The Difference**:
```
ARCHIVED → Closed permanently (by customer)
→ Cannot reactivate
→ Create new tenant

SOFT-DELETED → Marked for cleanup
→ Can be restored within 90 days
→ Same tenant reactivated
```

---

### Problem 4: Tenant isolation not working

**Symptoms**:
- Users seeing other companies' data
- Invoices from wrong tenant appearing
- User list includes other tenants' staff

** This is a CRITICAL security issue**

**Immediate Steps**:

**For Users**:
1. Log out immediately
2. Report to support/security team
3. Document what was seen
4. Do not share or use other tenant's data

**For Administrators**:
1. Check tenant context is set correctly
2. Verify user is in correct tenant
3. Check subdomain/URL is correct
4. Review recent authentication logs

**Common Causes**:

**Cause 1: Wrong Subdomain**
```
Should be: acme.yoursystem.com
Using: beta.yoursystem.com
Result: Seeing Beta's data instead of Acme's

Solution: Use correct subdomain
```

**Cause 2: Technical - Tenant Context Not Set**
```
Database connection missing tenant filter
All data visible instead of just one tenant

Solution: Contact development team immediately
```

**Cause 3: Technical - Using Superuser Role**
```
Connection using database superuser
Bypasses all security filters

Solution: Never use superuser in production
```

**Cause 4: Technical - RLS Not Enabled**
```
Row-level security not configured on table
No filtering at database level

Solution: Enable RLS on affected tables
```

**Verification Test**:
```
Test 1: Can you see your own invoices? ✅
Test 2: Search for known other tenant invoice number
Result should be: "Not found" ✅
If you can see it: ❌ Report immediately
```

---

### Problem 5: Rate limit errors (429 responses)

**What happened**: Getting "Too Many Requests" errors from API

**Common Scenarios**:

**Scenario A: Bulk Import**
```
Problem: Uploading 1,000 invoices rapidly
Rate Limit: 600 requests/minute
Result: First 600 succeed, rest fail

Solution: Add delay between requests
Example: Upload 10 per second (600/minute)
```

**Scenario B: Polling Loop**
```
Problem: Checking for updates every second
Rate Limit: Hit within 10 minutes

Solution: Use webhooks instead
System notifies you when something changes
No need to keep asking
```

**Scenario C: Legitimate High Volume**
```
Problem: Business needs more API calls
Current Plan: 600 requests/minute
Need: 1,500 requests/minute

Solution: Upgrade to Enterprise plan
Offers: 6,000 requests/minute
```

**How to Handle 429 Errors**:
```
1. Check response headers:
   X-RateLimit-Limit: 600
   X-RateLimit-Remaining: 0
   Retry-After: 45

2. Wait the specified time: 45 seconds

3. Retry your request

4. Implement exponential backoff:
   - First retry: Wait 1 second
   - Second retry: Wait 2 seconds
   - Third retry: Wait 4 seconds
   - Etc.
```

**Prevention**:
1. Batch multiple operations into one request
2. Cache results instead of re-requesting
3. Use webhooks for real-time updates
4. Spread requests over time
5. Monitor your rate limit usage

---

### Problem 6: Storage quota exceeded

**What happened**: Cannot upload files - storage is full

**Check Current Usage**:
```
Dashboard → Settings → Storage
Shows: "1.8 GB of 2.0 GB used (90%)"
```

**Where Space Goes**:

| Category | Typical Usage | Can Clean? |
|----------|--------------|------------|
| Documents | 40% | ⚠️ Carefully |
| Attachments | 30% | ✅ Yes |
| Exports | 20% | ✅ Yes - auto-deleted after 30 days |
| Backups | 10% | ⚠️ System-managed |

**Solutions**:

**Quick Win - Delete Old Exports**
```
1. Go to Reports → Export History
2. Delete exports older than 7 days
3. These were for temporary download
4. Can regenerate if needed later
5. Usually frees 100-300 MB
```

**Medium Effort - Remove Attachments**
```
1. Review invoices/documents with attachments
2. Remove duplicates
3. Delete attachments for old records
4. Keep important documents only
5. Can free 200-500 MB
```

**Long Term - Archive Old Data**
```
1. Export old records (> 2 years)
2. Save export to your own storage
3. Delete old records from system
4. Reduces both storage and complexity
5. Can free up significant space
```

**Upgrade Plan**
```
Professional: 2 GB → 5 GB ($X more/month)
Enterprise: 10 GB+ (contact sales)
```

**Prevention**:
1. Set up automatic export deletion
2. Regular cleanup schedule (quarterly)
3. Compress large files before upload
4. Use external storage for large archives

---

### Problem 7: Forgot which subdomain we use

**Symptoms**: Cannot access tenant account - forgot URL

**Finding Your Subdomain**:

**Option 1: Check Email**
- Look for welcome/confirmation email
- Contains your full login URL
- Example: "Access your account at acme.yoursystem.com"

**Option 2: Try Common Patterns**
```
Company name: Acme Corporation
Try:
- acme.yoursystem.com
- acme-corp.yoursystem.com
- acmecorp.yoursystem.com
```

**Option 3: Contact Support**
- Provide: Company name and email
- Support can look up your subdomain
- They'll verify your identity first

**Option 4: Check Browser History**
- Look for yoursystem.com in history
- Filter by domain
- Find your specific subdomain

**Future Prevention**:
- Bookmark your login page
- Save in password manager
- Add to company wiki/documentation

---

## Getting More Help

### When to Contact Support

Contact support if:
- ✅ Security/isolation issues (URGENT)
- ✅ Cannot access your account
- ✅ Data appears corrupted
- ✅ Billing/plan questions
- ✅ Need temporary limit increases
- ✅ Technical errors not listed here

### Information to Provide

Include this in your support request:
1. **Tenant Name**: Company/organization name
2. **Subdomain**: If you know it
3. **Email**: Primary contact email
4. **Problem Description**: What you were trying to do
5. **Error Message**: Exact text of any errors
6. **When It Started**: Date/time
7. **Frequency**: One-time or ongoing
8. **Impact**: How many users affected

### Response Times

| Priority | Response Time | Examples |
|----------|--------------|----------|
|  Critical | 1 hour | Security breach, system down |
|  High | 4 hours | Cannot access, data loss |
|  Normal | 24 hours | Feature questions, minor bugs |
| ⚪ Low | 48 hours | Feature requests, optimization |

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
