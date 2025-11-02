## Section 11: Security & Permissions

### 🎯 Learning Objectives

By the end of this section, you will understand:
- Security struct configuration
- CSRF protection
- Rate limiting
- Permission system
- Tenant isolation
- Audit logging

### 11.1 Security Struct

```go
type Security struct {
    CSRF       *CSRF       `json:"csrf,omitempty"`
    RateLimit  *RateLimit  `json:"rateLimit,omitempty"`
    Encryption bool        `json:"encryption,omitempty"`
}
```

### 11.2 CSRF Protection

**Cross-Site Request Forgery Protection**

```go
type CSRF struct {
    Enabled    bool   `json:"enabled"`
    FieldName  string `json:"fieldName"`  // Hidden field name
    HeaderName string `json:"headerName"` // HTTP header name
}
```

**Example:**
```json
{
  "security": {
    "csrf": {
      "enabled": true,
      "fieldName": "_csrf",
      "headerName": "X-CSRF-Token"
    }
  }
}
```

**Implementation:**

**Backend generates token:**
```go
func GenerateCSRFToken(sessionID string) string {
    // Generate cryptographically secure token
    token := generateSecureToken()
    
    // Store in session
    session.Set("csrf_token", token)
    
    return token
}
```

**templ includes token:**
```go
templ FormRenderer(schema *Schema, data map[string]any) {
    <form hx-post={schema.HTMX.Post}>
        // CSRF token (if enabled)
        if schema.Security != nil && schema.Security.CSRF != nil && schema.Security.CSRF.Enabled {
            <input 
                type="hidden" 
                name={schema.Security.CSRF.FieldName} 
                value={getCSRFToken(ctx)} 
            />
        }
        
        // Fields...
        for _, field := range schema.Fields {
            @FieldRenderer(&field, data[field.Name])
        }
    </form>
}
```

**HTML output:**
```html
<form hx-post="/api/users">
  <input type="hidden" name="_csrf" value="abc123xyz789..." />
  <!-- Fields -->
</form>
```

**Backend validates:**
```go
func ValidateCSRF(r *http.Request) error {
    // Get token from form
    formToken := r.FormValue("_csrf")
    
    // Get token from session
    session, _ := store.Get(r, "session")
    sessionToken, _ := session.Values["csrf_token"].(string)
    
    // Compare
    if !secureCompare(formToken, sessionToken) {
        return errors.New("CSRF token mismatch")
    }
    
    return nil
}
```

### 11.3 Rate Limiting

```go
type RateLimit struct {
    Enabled     bool   `json:"enabled"`
    MaxRequests int    `json:"maxRequests"` // Max requests
    WindowSec   int64  `json:"windowSec"`   // Time window in seconds
    ByUser      bool   `json:"byUser"`      // Limit per user
    ByIP        bool   `json:"byIP"`        // Limit per IP
}
```

**Example:**
```json
{
  "security": {
    "rateLimit": {
      "enabled": true,
      "maxRequests": 100,
      "windowSec": 3600,
      "byUser": true,
      "byIP": true
    }
  }
}
```

**Meaning:** Max 100 requests per hour per user AND per IP.

**Implementation:**
```go
type RateLimiter struct {
    store *redis.Client
}

func (rl *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
    // Use Redis for distributed rate limiting
    count, err := rl.store.Incr(ctx, key).Result()
    if err != nil {
        return false, err
    }
    
    if count == 1 {
        // First request, set expiration
        rl.store.Expire(ctx, key, window)
    }
    
    return count <= int64(limit), nil
}

func RateLimitMiddleware(schema *Schema) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if schema.Security == nil || schema.Security.RateLimit == nil || !schema.Security.RateLimit.Enabled {
                next.ServeHTTP(w, r)
                return
            }
            
            rl := schema.Security.RateLimit
            limiter := NewRateLimiter(redisClient)
            
            // Check by user
            if rl.ByUser {
                user := GetUserFromContext(r.Context())
                key := fmt.Sprintf("ratelimit:user:%s:%s", user.ID, schema.ID)
                allowed, _ := limiter.Allow(r.Context(), key, rl.MaxRequests, time.Duration(rl.WindowSec)*time.Second)
                if !allowed {
                    http.Error(w, "Rate limit exceeded", 429)
                    return
                }
            }
            
            // Check by IP
            if rl.ByIP {
                ip := getClientIP(r)
                key := fmt.Sprintf("ratelimit:ip:%s:%s", ip, schema.ID)
                allowed, _ := limiter.Allow(r.Context(), key, rl.MaxRequests, time.Duration(rl.WindowSec)*time.Second)
                if !allowed {
                    http.Error(w, "Rate limit exceeded", 429)
                    return
                }
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

### 11.4 Permission System

**Field-Level Permissions:**

```go
type FieldPermissions struct {
    View     []string `json:"view,omitempty"`     // Roles that can view
    Edit     []string `json:"edit,omitempty"`     // Roles that can edit
    Required []string `json:"required,omitempty"` // Permissions needed
}
```

**Example:**
```json
{
  "name": "salary",
  "type": "currency",
  "label": "Salary",
  "permissions": {
    "view": ["hr_manager", "ceo"],
    "edit": ["hr_manager"]
  }
}
```

**OR use simpler RequirePermission:**
```json
{
  "name": "salary",
  "type": "currency",
  "label": "Salary",
  "requirePermission": "hr.view_salary"
}
```

**User struct:**
```go
type User struct {
    ID          string
    TenantID    string
    Roles       []string
    Permissions []string
}

func (u *User) HasRole(role string) bool {
    for _, r := range u.Roles {
        if r == role || r == "*" {
            return true
        }
    }
    return false
}

func (u *User) HasPermission(permission string) bool {
    for _, p := range u.Permissions {
        if p == permission || p == "*" {
            return true
        }
        
        // Support wildcard permissions
        // "hr.*" matches "hr.view_salary", "hr.edit_salary", etc.
        if strings.HasSuffix(p, ".*") {
            prefix := strings.TrimSuffix(p, ".*")
            if strings.HasPrefix(permission, prefix+".") {
                return true
            }
        }
    }
    return false
}

func (u *User) HasAnyRole(roles []string) bool {
    for _, role := range roles {
        if u.HasRole(role) {
            return true
        }
    }
    return false
}
```

**Enricher applies permissions:**
```go
func (e *Enricher) Enrich(ctx context.Context, schema *Schema, user *User) (*Schema, error) {
    for i := range schema.Fields {
        field := &schema.Fields[i]
        
        visible := true
        editable := true
        
        // Check RequirePermission (simpler)
        if field.RequirePermission != "" {
            if !user.HasPermission(field.RequirePermission) {
                visible = false
                editable = false
            }
        }
        
        // Check Permissions (more granular)
        if field.Permissions != nil {
            if len(field.Permissions.View) > 0 {
                visible = user.HasAnyRole(field.Permissions.View)
            }
            if len(field.Permissions.Edit) > 0 {
                editable = user.HasAnyRole(field.Permissions.Edit)
            }
        }
        
        field.Runtime = &FieldRuntime{
            Visible:  visible,
            Editable: editable && !field.Readonly,
            Reason:   "permission_check",
        }
    }
    
    return schema, nil
}
```

### 11.5 Tenant Isolation

```go
type Tenant struct {
    Enabled   bool   `json:"enabled"`
    Field     string `json:"field"`     // Field name containing tenant ID
    Isolation string `json:"isolation"` // "strict", "shared"
}
```

**Example:**
```json
{
  "tenant": {
    "enabled": true,
    "field": "tenant_id",
    "isolation": "strict"
  }
}
```

**Isolation Strategies:**

**Strict Isolation:**
- Each tenant's data completely separate
- Cannot see/access other tenant's data
- Enforced at database level (RLS)

**Shared Isolation:**
- Data can be shared between tenants
- Explicit sharing required
- Useful for multi-company scenarios

**Implementation:**

**Middleware:**
```go
func TenantMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user := GetUserFromContext(r.Context())
        
        // Add tenant to context
        ctx := context.WithValue(r.Context(), tenantKey, user.TenantID)
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

**Database queries (PostgreSQL RLS):**
```sql
-- Enable Row Level Security
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

-- Create policy
CREATE POLICY tenant_isolation ON users
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

-- Set tenant for session
SET app.current_tenant = 'tenant-uuid';
```

**Go code:**
```go
func (r *Repository) List(ctx context.Context) ([]*User, error) {
    tenantID := GetTenantFromContext(ctx)
    
    // Set tenant for session
    _, err := r.db.ExecContext(ctx, "SET app.current_tenant = $1", tenantID)
    if err != nil {
        return nil, err
    }
    
    // Query - RLS automatically filters by tenant
    rows, err := r.db.QueryContext(ctx, "SELECT * FROM users")
    // ...
}
```

**Enricher enforces tenant:**
```go
func (e *Enricher) Enrich(ctx context.Context, schema *Schema, user *User) (*Schema, error) {
    if schema.Tenant != nil && schema.Tenant.Enabled {
        // Ensure all data operations include tenant_id
        for i := range schema.Fields {
            field := &schema.Fields[i]
            
            // If this is the tenant field, make it readonly
            if field.Name == schema.Tenant.Field {
                field.Readonly = true
                field.Value = user.TenantID
            }
        }
    }
    
    return schema, nil
}
```

### 11.6 Audit Logging

Track who did what and when:

```go
type AuditLog struct {
    ID        string    `json:"id"`
    SchemaID  string    `json:"schemaId"`
    UserID    string    `json:"userId"`
    TenantID  string    `json:"tenantId"`
    Action    string    `json:"action"`    // "create", "update", "delete", "view"
    Before    any       `json:"before"`    // State before change
    After     any       `json:"after"`     // State after change
    IP        string    `json:"ip"`
    UserAgent string    `json:"userAgent"`
    Timestamp time.Time `json:"timestamp"`
}
```

**Implementation:**
```go
func LogAudit(ctx context.Context, action string, schemaID string, before, after any) {
    user := GetUserFromContext(ctx)
    
    log := &AuditLog{
        ID:        uuid.New().String(),
        SchemaID:  schemaID,
        UserID:    user.ID,
        TenantID:  user.TenantID,
        Action:    action,
        Before:    before,
        After:     after,
        IP:        getClientIP(ctx),
        UserAgent: getUserAgent(ctx),
        Timestamp: time.Now(),
    }
    
    // Save to audit log table
    auditRepo.Save(ctx, log)
}

func HandleUpdate(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Get existing data
    before, _ := repo.Get(ctx, id)
    
    // Update
    after := updateData(before, newData)
    repo.Save(ctx, after)
    
    // Log audit
    LogAudit(ctx, "update", schemaID, before, after)
}
```

### 11.7 Complete Security Example

```json
{
  "id": "employee-form",
  "type": "form",
  "title": "Employee Information",
  
  "security": {
    "csrf": {
      "enabled": true,
      "fieldName": "_csrf",
      "headerName": "X-CSRF-Token"
    },
    "rateLimit": {
      "enabled": true,
      "maxRequests": 50,
      "windowSec": 3600,
      "byUser": true,
      "byIP": true
    },
    "encryption": true
  },
  
  "tenant": {
    "enabled": true,
    "field": "tenant_id",
    "isolation": "strict"
  },
  
  "fields": [
    {
      "name": "tenant_id",
      "type": "hidden"
    },
    {
      "name": "first_name",
      "type": "text",
      "label": "First Name",
      "required": true
    },
    {
      "name": "last_name",
      "type": "text",
      "label": "Last Name",
      "required": true
    },
    {
      "name": "email",
      "type": "email",
      "label": "Email",
      "required": true,
      "validation": {
        "server": {
          "unique": true,
          "uniqueWith": ["tenant_id"]
        }
      }
    },
    {
      "name": "ssn",
      "type": "text",
      "label": "SSN",
      "requirePermission": "hr.view_ssn",
      "mask": {
        "pattern": "999-99-9999"
      }
    },
    {
      "name": "salary",
      "type": "currency",
      "label": "Salary",
      "permissions": {
        "view": ["hr_manager", "ceo"],
        "edit": ["hr_manager"]
      }
    },
    {
      "name": "department",
      "type": "select",
      "label": "Department",
      "required": true
    },
    {
      "name": "hire_date",
      "type": "date",
      "label": "Hire Date",
      "required": true
    }
  ]
}
```

**This provides:**
- ✅ CSRF protection on form submission
- ✅ Rate limiting (50 requests/hour per user and IP)
- ✅ Tenant isolation (strict)
- ✅ SSN visible only to HR with permission
- ✅ Salary visible to HR manager and CEO, editable by HR manager only
- ✅ Email unique per tenant
- ✅ Encrypted data transmission

---

