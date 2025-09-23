# ERP Application State, Configuration & Access Control Guide

## Overview

This guide covers the implementation of application state management, global configuration, feature flags, and Attribute-Based Access Control (ABAC) for ERP systems using **Templ + HTMX + Alpine.js + Flowbite**.

## Application State Management

### State Architecture

```
Application State
├── Client State (Alpine.js)
│   ├── UI State (modals, menus, notifications)
│   ├── Form State (validation, drafts)
│   ├── User Preferences (theme, layout)
│   └── Cached Data (frequently accessed)
├── Server State (Go)
│   ├── Session State (user context, permissions)
│   ├── Application State (feature flags, config)
│   ├── Business State (workflow, transactions)
│   └── Cache State (Redis/Memory)
└── Persistent State (Database)
    ├── User Data (profiles, preferences)
    ├── Business Data (entities, transactions)
    ├── Configuration Data (settings, rules)
    └── Audit Data (logs, changes)
```

### Client-Side State Management

**Global Alpine.js Stores:**
```javascript
// Global state stores
document.addEventListener('alpine:init', () => {
    // Application state
    Alpine.store('app', {
        // UI State
        sidebarOpen: false,
        theme: localStorage.getItem('theme') || 'light',
        notifications: [],
        
        // User context
        user: null,
        permissions: [],
        
        // Feature flags (synced from server)
        features: {},
        
        // Configuration
        config: {
            dateFormat: 'YYYY-MM-DD',
            currency: 'USD',
            timezone: 'UTC'
        },
        
        // Methods
        toggleSidebar() {
            this.sidebarOpen = !this.sidebarOpen;
        },
        
        addNotification(type, message, duration = 5000) {
            const id = Date.now();
            this.notifications.push({ id, type, message });
            setTimeout(() => this.removeNotification(id), duration);
        },
        
        removeNotification(id) {
            this.notifications = this.notifications.filter(n => n.id !== id);
        },
        
        setTheme(theme) {
            this.theme = theme;
            localStorage.setItem('theme', theme);
            document.documentElement.setAttribute('data-theme', theme);
        },
        
        hasPermission(permission) {
            return this.permissions.includes(permission) || this.permissions.includes('*');
        },
        
        hasFeature(feature) {
            return this.features[feature] === true;
        }
    });
    
    // Form state management
    Alpine.store('forms', {
        drafts: {},
        validation: {},
        
        saveDraft(formId, data) {
            this.drafts[formId] = {
                data,
                timestamp: Date.now()
            };
            // Persist to localStorage for recovery
            localStorage.setItem(`draft_${formId}`, JSON.stringify(this.drafts[formId]));
        },
        
        loadDraft(formId) {
            const stored = localStorage.getItem(`draft_${formId}`);
            if (stored) {
                this.drafts[formId] = JSON.parse(stored);
                return this.drafts[formId].data;
            }
            return null;
        },
        
        clearDraft(formId) {
            delete this.drafts[formId];
            localStorage.removeItem(`draft_${formId}`);
        },
        
        setFieldValidation(form, field, isValid, message = '') {
            if (!this.validation[form]) {
                this.validation[form] = {};
            }
            this.validation[form][field] = { isValid, message };
        }
    });
    
    // Data cache store
    Alpine.store('cache', {
        data: {},
        timestamps: {},
        ttl: 5 * 60 * 1000, // 5 minutes
        
        set(key, value) {
            this.data[key] = value;
            this.timestamps[key] = Date.now();
        },
        
        get(key) {
            const timestamp = this.timestamps[key];
            if (!timestamp || (Date.now() - timestamp) > this.ttl) {
                delete this.data[key];
                delete this.timestamps[key];
                return null;
            }
            return this.data[key];
        },
        
        clear(pattern = null) {
            if (pattern) {
                Object.keys(this.data).forEach(key => {
                    if (key.includes(pattern)) {
                        delete this.data[key];
                        delete this.timestamps[key];
                    }
                });
            } else {
                this.data = {};
                this.timestamps = {};
            }
        }
    });
});
```

**State Synchronization:**
```javascript
// Sync state with server
function syncAppState() {
    fetch('/api/app-state', {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        }
    })
    .then(response => response.json())
    .then(data => {
        Alpine.store('app').user = data.user;
        Alpine.store('app').permissions = data.permissions;
        Alpine.store('app').features = data.features;
        Alpine.store('app').config = { ...Alpine.store('app').config, ...data.config };
    })
    .catch(error => console.error('Failed to sync app state:', error));
}

// Sync on page load and periodically
document.addEventListener('DOMContentLoaded', syncAppState);
setInterval(syncAppState, 10 * 60 * 1000); // Every 10 minutes
```

### Server-Side State Management

**Session Management:**
```go
// session/manager.go
type SessionManager struct {
    store   SessionStore
    timeout time.Duration
}

type Session struct {
    ID          string                 `json:"id"`
    UserID      int                    `json:"user_id"`
    User        *User                  `json:"user"`
    Permissions []string               `json:"permissions"`
    Features    map[string]bool        `json:"features"`
    Data        map[string]interface{} `json:"data"`
    CreatedAt   time.Time             `json:"created_at"`
    LastAccess  time.Time             `json:"last_access"`
    ExpiresAt   time.Time             `json:"expires_at"`
}

func (sm *SessionManager) CreateSession(userID int) (*Session, error) {
    user, err := sm.getUserWithPermissions(userID)
    if err != nil {
        return nil, err
    }
    
    session := &Session{
        ID:          generateSessionID(),
        UserID:      userID,
        User:        user,
        Permissions: user.GetPermissions(),
        Features:    sm.getEnabledFeatures(user),
        Data:        make(map[string]interface{}),
        CreatedAt:   time.Now(),
        LastAccess:  time.Now(),
        ExpiresAt:   time.Now().Add(sm.timeout),
    }
    
    return session, sm.store.Save(session)
}

func (sm *SessionManager) GetSession(sessionID string) (*Session, error) {
    session, err := sm.store.Get(sessionID)
    if err != nil {
        return nil, err
    }
    
    if time.Now().After(session.ExpiresAt) {
        sm.store.Delete(sessionID)
        return nil, errors.New("session expired")
    }
    
    // Update last access
    session.LastAccess = time.Now()
    session.ExpiresAt = time.Now().Add(sm.timeout)
    sm.store.Save(session)
    
    return session, nil
}
```

**Context Management:**
```go
// context/app_context.go
type AppContext struct {
    Session     *Session
    User        *User
    Permissions []string
    Features    map[string]bool
    Config      *AppConfig
    RequestID   string
    StartTime   time.Time
}

func (ctx *AppContext) HasPermission(permission string) bool {
    if contains(ctx.Permissions, "*") {
        return true
    }
    return contains(ctx.Permissions, permission)
}

func (ctx *AppContext) HasFeature(feature string) bool {
    enabled, exists := ctx.Features[feature]
    return exists && enabled
}

func (ctx *AppContext) CanAccess(resource string, action string) bool {
    return ctx.ABAC().Evaluate(ABACRequest{
        Subject:  ctx.User,
        Resource: resource,
        Action:   action,
        Context:  ctx,
    })
}
```

## Global Configuration Management

### Configuration Structure

```go
// config/app_config.go
type AppConfig struct {
    // Application settings
    App struct {
        Name        string `yaml:"name" default:"ERP System"`
        Version     string `yaml:"version" default:"1.0.0"`
        Environment string `yaml:"environment" default:"development"`
        Debug       bool   `yaml:"debug" default:"false"`
    } `yaml:"app"`
    
    // Server configuration
    Server struct {
        Host            string        `yaml:"host" default:"localhost"`
        Port            int           `yaml:"port" default:"8080"`
        ReadTimeout     time.Duration `yaml:"read_timeout" default:"30s"`
        WriteTimeout    time.Duration `yaml:"write_timeout" default:"30s"`
        ShutdownTimeout time.Duration `yaml:"shutdown_timeout" default:"10s"`
    } `yaml:"server"`
    
    // Database configuration
    Database struct {
        Driver          string `yaml:"driver" default:"postgres"`
        Host            string `yaml:"host" default:"localhost"`
        Port            int    `yaml:"port" default:"5432"`
        Name            string `yaml:"name" default:"erp_db"`
        User            string `yaml:"user" default:"erp_user"`
        Password        string `yaml:"password" default:"password"`
        SSLMode         string `yaml:"ssl_mode" default:"disable"`
        MaxOpenConns    int    `yaml:"max_open_conns" default:"25"`
        MaxIdleConns    int    `yaml:"max_idle_conns" default:"25"`
        ConnMaxLifetime string `yaml:"conn_max_lifetime" default:"5m"`
    } `yaml:"database"`
    
    // Cache configuration  
    Cache struct {
        Driver    string        `yaml:"driver" default:"memory"`
        RedisURL  string        `yaml:"redis_url"`
        TTL       time.Duration `yaml:"ttl" default:"1h"`
        MaxMemory string        `yaml:"max_memory" default:"100MB"`
    } `yaml:"cache"`
    
    // Feature flags
    Features map[string]FeatureConfig `yaml:"features"`
    
    // Business configuration
    Business struct {
        Currency      string   `yaml:"currency" default:"USD"`
        Timezone      string   `yaml:"timezone" default:"UTC"`
        DateFormat    string   `yaml:"date_format" default:"2006-01-02"`
        Languages     []string `yaml:"languages" default:"[en]"`
        WorkingHours  struct {
            Start string `yaml:"start" default:"09:00"`
            End   string `yaml:"end" default:"17:00"`
        } `yaml:"working_hours"`
        Departments   []string `yaml:"departments"`
        UserRoles     []string `yaml:"user_roles"`
    } `yaml:"business"`
    
    // Integration settings
    Integrations struct {
        Email struct {
            Provider string `yaml:"provider" default:"smtp"`
            SMTPHost string `yaml:"smtp_host"`
            SMTPPort int    `yaml:"smtp_port"`
            Username string `yaml:"username"`
            Password string `yaml:"password"`
        } `yaml:"email"`
        
        Storage struct {
            Provider string `yaml:"provider" default:"local"`
            S3Bucket string `yaml:"s3_bucket"`
            S3Region string `yaml:"s3_region"`
            LocalPath string `yaml:"local_path" default:"./uploads"`
        } `yaml:"storage"`
    } `yaml:"integrations"`
    
    // Security settings
    Security struct {
        JWTSecret           string        `yaml:"jwt_secret"`
        SessionTimeout      time.Duration `yaml:"session_timeout" default:"24h"`
        PasswordMinLength   int           `yaml:"password_min_length" default:"8"`
        RequireStrongPassword bool        `yaml:"require_strong_password" default:"true"`
        MaxLoginAttempts    int           `yaml:"max_login_attempts" default:"5"`
        LockoutDuration     time.Duration `yaml:"lockout_duration" default:"30m"`
        EnableTwoFactor     bool          `yaml:"enable_two_factor" default:"false"`
    } `yaml:"security"`
}

type FeatureConfig struct {
    Enabled     bool                   `yaml:"enabled"`
    Rollout     float64               `yaml:"rollout"` // 0-100 percentage
    Rules       []FeatureRule         `yaml:"rules"`
    Metadata    map[string]interface{} `yaml:"metadata"`
    Description string                `yaml:"description"`
}

type FeatureRule struct {
    Type      string      `yaml:"type"`
    Field     string      `yaml:"field"`
    Operator  string      `yaml:"operator"`
    Value     interface{} `yaml:"value"`
    Enabled   bool        `yaml:"enabled"`
}
```

### Configuration Management

**Configuration Loader:**
```go
// config/loader.go
type ConfigManager struct {
    config   *AppConfig
    watchers []ConfigWatcher
    mutex    sync.RWMutex
}

func NewConfigManager() *ConfigManager {
    return &ConfigManager{
        config:   &AppConfig{},
        watchers: make([]ConfigWatcher, 0),
    }
}

func (cm *ConfigManager) Load(configPath string) error {
    cm.mutex.Lock()
    defer cm.mutex.Unlock()
    
    // Load from file
    data, err := os.ReadFile(configPath)
    if err != nil {
        return err
    }
    
    // Parse YAML
    if err := yaml.Unmarshal(data, cm.config); err != nil {
        return err
    }
    
    // Apply environment variable overrides
    cm.applyEnvironmentOverrides()
    
    // Validate configuration
    if err := cm.validate(); err != nil {
        return err
    }
    
    // Notify watchers
    cm.notifyWatchers()
    
    return nil
}

func (cm *ConfigManager) Get() *AppConfig {
    cm.mutex.RLock()
    defer cm.mutex.RUnlock()
    
    // Return a copy to prevent mutations
    configCopy := *cm.config
    return &configCopy
}

func (cm *ConfigManager) Watch(watcher ConfigWatcher) {
    cm.mutex.Lock()
    defer cm.mutex.Unlock()
    
    cm.watchers = append(cm.watchers, watcher)
}

func (cm *ConfigManager) applyEnvironmentOverrides() {
    // Override with environment variables
    if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
        cm.config.Database.Host = dbHost
    }
    if dbPort := os.Getenv("DB_PORT"); dbPort != "" {
        if port, err := strconv.Atoi(dbPort); err == nil {
            cm.config.Database.Port = port
        }
    }
    // Add more overrides as needed
}
```

**Dynamic Configuration Updates:**
```go
// config/dynamic.go
type DynamicConfigManager struct {
    static  *ConfigManager
    dynamic map[string]interface{}
    mutex   sync.RWMutex
    store   ConfigStore
}

func (dcm *DynamicConfigManager) UpdateFeature(featureName string, config FeatureConfig) error {
    dcm.mutex.Lock()
    defer dcm.mutex.Unlock()
    
    // Update in-memory config
    if dcm.dynamic["features"] == nil {
        dcm.dynamic["features"] = make(map[string]FeatureConfig)
    }
    dcm.dynamic["features"].(map[string]FeatureConfig)[featureName] = config
    
    // Persist to store
    return dcm.store.UpdateFeature(featureName, config)
}

func (dcm *DynamicConfigManager) GetFeatureConfig(featureName string) (FeatureConfig, bool) {
    dcm.mutex.RLock()
    defer dcm.mutex.RUnlock()
    
    // Check dynamic config first
    if features, ok := dcm.dynamic["features"].(map[string]FeatureConfig); ok {
        if config, exists := features[featureName]; exists {
            return config, true
        }
    }
    
    // Fall back to static config
    static := dcm.static.Get()
    config, exists := static.Features[featureName]
    return config, exists
}
```

## Feature Flags System

### Feature Flag Implementation

**Feature Flag Engine:**
```go
// features/engine.go
type FeatureFlagEngine struct {
    config  *DynamicConfigManager
    cache   FeatureCache
    metrics FeatureMetrics
}

func (ffe *FeatureFlagEngine) IsEnabled(featureName string, context *EvaluationContext) bool {
    // Get feature configuration
    featureConfig, exists := ffe.config.GetFeatureConfig(featureName)
    if !exists {
        ffe.metrics.RecordMissingFeature(featureName)
        return false
    }
    
    // Check if feature is globally disabled
    if !featureConfig.Enabled {
        ffe.metrics.RecordDisabledFeature(featureName)
        return false
    }
    
    // Evaluate rollout percentage
    if !ffe.evaluateRollout(featureName, featureConfig.Rollout, context) {
        ffe.metrics.RecordRolloutExclusion(featureName)
        return false
    }
    
    // Evaluate feature rules
    enabled := ffe.evaluateRules(featureConfig.Rules, context)
    ffe.metrics.RecordEvaluation(featureName, enabled)
    
    return enabled
}

func (ffe *FeatureFlagEngine) evaluateRollout(featureName string, rollout float64, context *EvaluationContext) bool {
    if rollout >= 100.0 {
        return true
    }
    if rollout <= 0.0 {
        return false
    }
    
    // Use consistent hashing for user-based rollout
    hash := fnv.New32a()
    hash.Write([]byte(featureName + ":" + context.UserID))
    hashValue := float64(hash.Sum32()%1000) / 10.0
    
    return hashValue < rollout
}

func (ffe *FeatureFlagEngine) evaluateRules(rules []FeatureRule, context *EvaluationContext) bool {
    for _, rule := range rules {
        if !rule.Enabled {
            continue
        }
        
        if !ffe.evaluateRule(rule, context) {
            return false
        }
    }
    return true
}

func (ffe *FeatureFlagEngine) evaluateRule(rule FeatureRule, context *EvaluationContext) bool {
    contextValue := context.GetValue(rule.Field)
    
    switch rule.Operator {
    case "equals":
        return contextValue == rule.Value
    case "not_equals":
        return contextValue != rule.Value
    case "in":
        if values, ok := rule.Value.([]interface{}); ok {
            for _, v := range values {
                if contextValue == v {
                    return true
                }
            }
        }
        return false
    case "greater_than":
        if num1, ok := contextValue.(float64); ok {
            if num2, ok := rule.Value.(float64); ok {
                return num1 > num2
            }
        }
        return false
    case "less_than":
        if num1, ok := contextValue.(float64); ok {
            if num2, ok := rule.Value.(float64); ok {
                return num1 < num2
            }
        }
        return false
    case "contains":
        if str1, ok := contextValue.(string); ok {
            if str2, ok := rule.Value.(string); ok {
                return strings.Contains(str1, str2)
            }
        }
        return false
    }
    
    return false
}
```

**Evaluation Context:**
```go
// features/context.go
type EvaluationContext struct {
    UserID      string                 `json:"user_id"`
    User        *User                  `json:"user"`
    Request     *http.Request          `json:"-"`
    Session     *Session               `json:"-"`
    Attributes  map[string]interface{} `json:"attributes"`
    Timestamp   time.Time             `json:"timestamp"`
}

func NewEvaluationContext(user *User, request *http.Request) *EvaluationContext {
    return &EvaluationContext{
        UserID:     strconv.Itoa(user.ID),
        User:       user,
        Request:    request,
        Attributes: make(map[string]interface{}),
        Timestamp:  time.Now(),
    }
}

func (ec *EvaluationContext) GetValue(field string) interface{} {
    switch field {
    case "user_id":
        return ec.UserID
    case "user_role":
        return ec.User.Role
    case "user_department":
        return ec.User.Department
    case "user_level":
        return ec.User.Level
    case "request_ip":
        return ec.getClientIP()
    case "request_user_agent":
        return ec.Request.UserAgent()
    case "time_hour":
        return float64(ec.Timestamp.Hour())
    case "time_day_of_week":
        return float64(ec.Timestamp.Weekday())
    default:
        if value, exists := ec.Attributes[field]; exists {
            return value
        }
        return nil
    }
}

func (ec *EvaluationContext) SetAttribute(key string, value interface{}) {
    ec.Attributes[key] = value
}

func (ec *EvaluationContext) getClientIP() string {
    if ec.Request == nil {
        return ""
    }
    
    // Check X-Forwarded-For header
    if xff := ec.Request.Header.Get("X-Forwarded-For"); xff != "" {
        return strings.Split(xff, ",")[0]
    }
    
    // Check X-Real-IP header
    if xri := ec.Request.Header.Get("X-Real-IP"); xri != "" {
        return xri
    }
    
    // Fall back to RemoteAddr
    ip, _, _ := net.SplitHostPort(ec.Request.RemoteAddr)
    return ip
}
```

### Feature Flag Usage Patterns

**Template Integration:**
```go
// templates/components/feature_flag.templ
templ FeatureFlag(feature string, content templ.Component) {
    if ctx.HasFeature(feature) {
        {content}
    }
}

templ FeatureFlagWithFallback(feature string, content templ.Component, fallback templ.Component) {
    if ctx.HasFeature(feature) {
        {content}
    } else {
        {fallback}
    }
}

// Usage in templates
@FeatureFlag("advanced_reporting", 
    @AdvancedReportingButton()
)

@FeatureFlagWithFallback("new_dashboard",
    @NewDashboard(),
    @LegacyDashboard()
)
```

**HTMX Integration:**
```go
// handlers/feature_handler.go
func (h *Handler) CheckFeature(w http.ResponseWriter, r *http.Request) {
    featureName := r.URL.Query().Get("feature")
    
    context := features.NewEvaluationContext(h.getCurrentUser(r), r)
    enabled := h.featureEngine.IsEnabled(featureName, context)
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]bool{
        "enabled": enabled,
    })
}
```

**Alpine.js Integration:**
```javascript
// Feature flag checking in Alpine
function featureCheck() {
    return {
        async hasFeature(featureName) {
            // Check client-side cache first
            const cached = Alpine.store('app').features[featureName];
            if (cached !== undefined) {
                return cached;
            }
            
            // Fetch from server
            try {
                const response = await fetch(`/api/features/${featureName}`);
                const data = await response.json();
                
                // Cache the result
                Alpine.store('app').features[featureName] = data.enabled;
                return data.enabled;
            } catch (error) {
                console.error('Feature check failed:', error);
                return false;
            }
        }
    }
}
```

## Attribute-Based Access Control (ABAC)

### ABAC Implementation

**Policy Engine:**
```go
// abac/engine.go
type ABACEngine struct {
    policyStore PolicyStore
    evaluator   PolicyEvaluator
    cache       PolicyCache
    logger      Logger
}

type ABACRequest struct {
    Subject  *User                  `json:"subject"`
    Resource string                `json:"resource"`
    Action   string                `json:"action"`
    Context  map[string]interface{} `json:"context"`
}

type ABACResponse struct {
    Decision    Decision               `json:"decision"`
    Reason      string                 `json:"reason"`
    Policies    []string               `json:"policies"`
    Attributes  map[string]interface{} `json:"attributes"`
    Duration    time.Duration          `json:"duration"`
}

type Decision int

const (
    Deny Decision = iota
    Permit
    Indeterminate
    NotApplicable
)

func (ae *ABACEngine) Evaluate(request ABACRequest) ABACResponse {
    start := time.Now()
    
    // Get applicable policies
    policies := ae.policyStore.GetPolicies(request.Resource, request.Action)
    
    // Evaluate each policy
    var results []PolicyResult
    for _, policy := range policies {
        result := ae.evaluator.Evaluate(policy, request)
        results = append(results, result)
    }
    
    // Combine results using policy combining algorithm
    decision := ae.combineResults(results)
    
    response := ABACResponse{
        Decision: decision.Decision,
        Reason:   decision.Reason,
        Policies: decision.ApplicablePolicies,
        Duration: time.Since(start),
    }
    
    // Log the decision
    ae.logger.LogDecision(request, response)
    
    return response
}

func (ae *ABACEngine) combineResults(results []PolicyResult) CombinedResult {
    // Implement policy combining algorithm (e.g., deny-unless-permit)
    hasPermit := false
    hasDeny := false
    reasons := []string{}
    applicablePolicies := []string{}
    
    for _, result := range results {
        switch result.Decision {
        case Permit:
            hasPermit = true
            reasons = append(reasons, result.Reason)
            applicablePolicies = append(applicablePolicies, result.PolicyID)
        case Deny:
            hasDeny = true
            reasons = append(reasons, result.Reason)
            applicablePolicies = append(applicablePolicies, result.PolicyID)
        }
    }
    
    // Deny-unless-permit: if any deny, then deny; otherwise permit if any permit
    if hasDeny {
        return CombinedResult{
            Decision:            Deny,
            Reason:             strings.Join(reasons, "; "),
            ApplicablePolicies: applicablePolicies,
        }
    } else if hasPermit {
        return CombinedResult{
            Decision:            Permit,
            Reason:             strings.Join(reasons, "; "),
            ApplicablePolicies: applicablePolicies,
        }
    }
    
    return CombinedResult{
        Decision: Deny,
        Reason:   "No applicable policies found",
    }
}
```

**Policy Definition:**
```go
// abac/policy.go
type Policy struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Description string            `json:"description"`
    Version     string            `json:"version"`
    Effect      Effect            `json:"effect"`
    Target      Target            `json:"target"`
    Condition   *Condition        `json:"condition,omitempty"`
    Obligations []Obligation      `json:"obligations,omitempty"`
    Metadata    map[string]string `json:"metadata"`
    CreatedAt   time.Time        `json:"created_at"`
    UpdatedAt   time.Time        `json:"updated_at"`
    CreatedBy   string           `json:"created_by"`
}

type Effect int

const (
    EffectDeny Effect = iota
    EffectPermit
)

type Target struct {
    Subjects  []AttributeMatch `json:"subjects"`
    Resources []AttributeMatch `json:"resources"`
    Actions   []AttributeMatch `json:"actions"`
}

type AttributeMatch struct {
    Attribute string      `json:"attribute"`
    Operator  string      `json:"operator"`
    Value     interface{} `json:"value"`
}

type Condition struct {
    Type     string       `json:"type"`
    Operator string       `json:"operator"`
    Left     *Condition   `json:"left,omitempty"`
    Right    *Condition   `json:"right,omitempty"`
    Match    *AttributeMatch `json:"match,omitempty"`
}

type Obligation struct {
    Type       string                 `json:"type"`
    Action     string                 `json:"action"`
    Parameters map[string]interface{} `json:"parameters"`
}
```

**Policy Examples:**
```yaml
# policies/user_management.yaml
- id: "allow_user_read_own_profile"
  name: "Allow users to read their own profile"
  effect: "permit"
  target:
    subjects:
      - attribute: "user.role"
        operator: "in"
        value: ["employee", "manager", "admin"]
    resources:
      - attribute: "resource.type"
        operator: "equals"
        value: "user_profile"
    actions:
      - attribute: "action.name"
        operator: "equals"
        value: "read"
  condition:
    type: "match"
    match:
      attribute: "resource.owner_id"
      operator: "equals"
      value: "subject.user_id"

- id: "allow_manager_read_team_profiles"
  name: "Allow managers to read their team member profiles"
  effect: "permit"
  target:
    subjects:
      - attribute: "user.role"
        operator: "equals"
        value: "manager"
    resources:
      - attribute: "resource.type"
        operator: "equals"
        value: "user_profile"
    actions:
      - attribute: "action.name"
        operator: "equals"
        value: "read"
  condition:
    type: "match"
    match:
      attribute: "resource.department"
      operator: "equals"
      value: "subject.department"

- id: "deny_access_outside_business_hours"
  name: "Deny access to sensitive data outside business hours"
  effect: "deny"
  target:
    subjects:
      - attribute: "user.role"
        operator: "not_equals"
        value: "admin"
    resources:
      - attribute: "resource.classification"
        operator: "equals"
        value: "sensitive"
    actions:
      - attribute: "action.name"
        operator: "in"
        value: ["read", "write", "delete"]
  condition:
    type: "or"
    left:
      type: "match"
      match:
        attribute: "context.time_hour"
        operator: "less_than"
        value: 9
    right:
      type: "match"
      match:
        attribute: "context.time_hour"
        operator: "greater_than"
        value: 17
```

### ABAC Integration Patterns

**Middleware Integration:**
```go
// middleware/abac.go
func ABACMiddleware(abacEngine *ABACEngine) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract resource and action from request
            resource := extractResource(r)
            action := extractAction(r)
            
            // Get current user
            user := getCurrentUser(r)
            if user == nil {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            
            // Build ABAC request
            abacRequest := ABACRequest{
                Subject:  user,
                Resource: resource,
                Action:   action,
                Context:  buildContext(r),
            }
            
            // Evaluate access
            response := abacEngine.Evaluate(abacRequest)
            
            if response.Decision != Permit {
                http.Error(w, "Access Denied: "+response.Reason, http.StatusForbidden)
                return
            }
            
            // Process obligations
            processObligations(response.Attributes, w, r)
            
            next.ServeHTTP(w, r)
        })
    }
}

func extractResource(r *http.Request) string {
    // Extract resource from URL path
    path := strings.TrimPrefix(r.URL.Path, "/api/")
    parts := strings.Split(path, "/")
    
    if len(parts) > 0 {
        return parts[0]
    }
    return "unknown"
}

func extractAction(r *http.Request) string {
    switch r.Method {
    case "GET":
        return "read"
    case "POST":
        return "create"
    case "PUT", "PATCH":
        return "update"
    case "DELETE":
        return "delete"
    default:
        return "unknown"
    }
}

func buildContext(r *http.Request) map[string]interface{} {
    return map[string]interface{}{
        "request_ip":         getClientIP(r),
        "request_user_agent": r.UserAgent(),
        "time_hour":          float64(time.Now().Hour()),
        "time_day_of_week":   float64(time.Now().Weekday()),
        "request_path":       r.URL.Path,
        "request_method":     r.Method,
    }
}
```

**Template Integration:**
```go
// templates/components/abac_guard.templ
templ ABACGuard(resource, action string, content templ.Component) {
    if ctx.CanAccess(resource, action) {
        {content}
    }
}

templ ABACGuardWithFallback(resource, action string, content, fallback templ.Component) {
    if ctx.CanAccess(resource, action) {
        {content}
    } else {
        {fallback}
    }
}

// Usage examples
@ABACGuard("users", "create",
    @CreateUserButton()
)

@ABACGuard("reports", "read",
    @ReportsSection()
)

@ABACGuardWithFallback("admin_panel", "access",
    @AdminPanelLink(),
    @div { class="text-gray-500" } { "Access Denied" }
)
```

**Alpine.js ABAC Integration:**
```javascript
// ABAC checking in Alpine
function abacCheck() {
    return {
        accessCache: new Map(),
        
        async canAccess(resource, action) {
            const cacheKey = `${resource}:${action}`;
            
            // Check cache first
            if (this.accessCache.has(cacheKey)) {
                const cached = this.accessCache.get(cacheKey);
                if (Date.now() - cached.timestamp < 60000) { // 1 minute cache
                    return cached.allowed;
                }
            }
            
            try {
                const response = await fetch('/api/abac/check', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        resource: resource,
                        action: action,
                        context: this.buildContext()
                    })
                });
                
                const data = await response.json();
                const allowed = data.decision === 'permit';
                
                // Cache the result
                this.accessCache.set(cacheKey, {
                    allowed: allowed,
                    timestamp: Date.now()
                });
                
                return allowed;
            } catch (error) {
                console.error('ABAC check failed:', error);
                return false; // Fail closed
            }
        },
        
        buildContext() {
            return {
                time_hour: new Date().getHours(),
                time_day_of_week: new Date().getDay(),
                user_agent: navigator.userAgent,
                // Add more context as needed
            };
        }
    }
}
```

## Performance Optimization

### Caching Strategies

**Multi-level Caching:**
```go
// cache/multilevel.go
type MultiLevelCache struct {
    l1 *sync.Map          // In-memory cache
    l2 *redis.Client      // Redis cache
    l3 Database           // Database fallback
}

func (mlc *MultiLevelCache) Get(key string) (interface{}, error) {
    // L1 Cache (memory)
    if value, ok := mlc.l1.Load(key); ok {
        return value, nil
    }
    
    // L2 Cache (Redis)
    if mlc.l2 != nil {
        value, err := mlc.l2.Get(key).Result()
        if err == nil {
            // Store in L1 for next time
            mlc.l1.Store(key, value)
            return value, nil
        }
    }
    
    // L3 Cache (Database)
    value, err := mlc.l3.Get(key)
    if err != nil {
        return nil, err
    }
    
    // Store in both L1 and L2
    mlc.l1.Store(key, value)
    if mlc.l2 != nil {
        mlc.l2.Set(key, value, time.Hour)
    }
    
    return value, nil
}
```

### State Synchronization

**Real-time State Updates:**
```go
// websocket/state_sync.go
type StateSyncManager struct {
    connections map[string]*websocket.Conn
    broadcast   chan StateUpdate
    mutex       sync.RWMutex
}

type StateUpdate struct {
    Type    string                 `json:"type"`
    Data    map[string]interface{} `json:"data"`
    UserID  string                 `json:"user_id,omitempty"`
    Scope   string                 `json:"scope"`
}

func (ssm *StateSyncManager) BroadcastStateUpdate(update StateUpdate) {
    ssm.broadcast <- update
}

func (ssm *StateSyncManager) handleStateUpdates() {
    for update := range ssm.broadcast {
        ssm.mutex.RLock()
        for userID, conn := range ssm.connections {
            // Send to specific user or broadcast to all
            if update.UserID == "" || update.UserID == userID {
                go func(c *websocket.Conn, u StateUpdate) {
                    if err := c.WriteJSON(u); err != nil {
                        ssm.removeConnection(userID)
                    }
                }(conn, update)
            }
        }
        ssm.mutex.RUnlock()
    }
}
```

## Monitoring and Analytics

### State Monitoring

**Metrics Collection:**
```go
// metrics/state_metrics.go
type StateMetrics struct {
    SessionCount       prometheus.Gauge
    FeatureUsage       *prometheus.CounterVec
    ABACEvaluations    *prometheus.HistogramVec
    ConfigReloads      prometheus.Counter
    StateUpdates       *prometheus.CounterVec
}

func (sm *StateMetrics) RecordFeatureUsage(feature string, enabled bool) {
    status := "disabled"
    if enabled {
        status = "enabled"
    }
    sm.FeatureUsage.WithLabelValues(feature, status).Inc()
}

func (sm *StateMetrics) RecordABACEvaluation(resource, action string, decision string, duration time.Duration) {
    sm.ABACEvaluations.WithLabelValues(resource, action, decision).Observe(duration.Seconds())
}
```

### Health Checks

**System Health Monitoring:**
```go
// health/checker.go
type HealthChecker struct {
    checks map[string]HealthCheck
}

type HealthCheck interface {
    Name() string
    Check() HealthStatus
}

type HealthStatus struct {
    Healthy   bool                   `json:"healthy"`
    Message   string                 `json:"message"`
    Details   map[string]interface{} `json:"details"`
    CheckedAt time.Time             `json:"checked_at"`
}

func (hc *HealthChecker) CheckAll() map[string]HealthStatus {
    results := make(map[string]HealthStatus)
    
    for name, check := range hc.checks {
        results[name] = check.Check()
    }
    
    return results
}

// Example health checks
type ConfigHealthCheck struct {
    configManager *ConfigManager
}

func (c *ConfigHealthCheck) Check() HealthStatus {
    config := c.configManager.Get()
    return HealthStatus{
        Healthy:   config != nil,
        Message:   "Configuration loaded successfully",
        CheckedAt: time.Now(),
    }
}

type FeatureFlagHealthCheck struct {
    engine *FeatureFlagEngine
}

func (f *FeatureFlagHealthCheck) Check() HealthStatus {
    // Test a simple feature flag evaluation
    context := &EvaluationContext{UserID: "health-check"}
    enabled := f.engine.IsEnabled("health_check_feature", context)
    
    return HealthStatus{
        Healthy:   true,
        Message:   "Feature flag engine operational",
        Details:   map[string]interface{}{"test_result": enabled},
        CheckedAt: time.Now(),
    }
}
```

This comprehensive guide provides a complete foundation for managing application state, configuration, feature flags, and access control in your ERP system. The combination of server-side management with client-side optimization ensures both security and performance while maintaining excellent user experience.
