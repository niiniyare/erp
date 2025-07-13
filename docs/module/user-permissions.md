# User Management & Permissions

## 👥 Overview

The ERP system implements a comprehensive identity and access management framework based on a sophisticated **Person-Employee-User separation pattern** with advanced **Role-Based Access Control (RBAC)** and **Attribute-Based Access Control (ABAC)**. The system provides enterprise-grade security with tenant isolation, row-level security, and comprehensive audit logging while maintaining usability for administrators and end users.

## 🏗️ Identity Management Architecture

### Person-Employee-User Separation Pattern

The system employs a three-tier identity model that separates concerns and provides maximum flexibility:

#### 1. Persons Table
Stores generic person entities representing individuals across various contexts.

```sql
CREATE TABLE persons (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE RESTRICT,
    person_type VARCHAR(20) NOT NULL 
        CHECK (person_type IN ('INDIVIDUAL', 'EMPLOYEE', 'CONTACT', 'CUSTOMER', 'VENDOR', 'CONTRACTOR')),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    middle_name VARCHAR(100),
    email VARCHAR(255),
    phone VARCHAR(20),
    birth_date DATE,
    national_id VARCHAR(50),
    tax_id VARCHAR(50),
    address JSONB DEFAULT '{}'::jsonb,
    security_attributes JSONB DEFAULT '{}'::jsonb, -- ABAC attributes
    metadata JSONB DEFAULT '{}'::jsonb,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ -- Soft delete support
);
```

**Key Features:**
- **Multi-purpose identity**: Supports employees, customers, vendors, contractors
- **ABAC integration**: Security attributes for fine-grained access control
- **Tenant isolation**: Built-in multi-tenancy with Row Level Security (RLS)
- **Soft delete**: Maintains data integrity while allowing deactivation
- **Flexible metadata**: JSONB storage for extensible person data

#### 2. Employees Table
Extends persons with employment-specific data and organizational hierarchy.

```sql
CREATE TABLE employees (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    person_id UUID NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    employee_number VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE RESTRICT,
    position_title VARCHAR(100),
    department_id UUID REFERENCES entities(uuid),
    manager_id UUID REFERENCES employees(id), -- Self-referential hierarchy
    hire_date DATE NOT NULL,
    termination_date DATE,
    salary_info JSONB DEFAULT '{}'::jsonb, -- Encrypted salary data
    employment_status VARCHAR(20) DEFAULT 'ACTIVE',
    work_schedule JSONB DEFAULT '{}'::jsonb,
    security_level INTEGER DEFAULT 0, -- Numeric clearance level
    access_attributes JSONB DEFAULT '{}'::jsonb, -- ABAC attributes
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
```

**Key Features:**
- **Organizational hierarchy**: Self-referential manager relationships
- **Security levels**: Numeric clearance for access control (0=lowest)
- **Employment tracking**: Status, schedules, salary information
- **ABAC attributes**: Employment-specific access control data

#### 3. Users Table
System access accounts with authentication data and security controls.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE RESTRICT,
    person_id UUID REFERENCES persons(id) ON DELETE SET NULL,
    employee_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    username VARCHAR(100),
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255),
    user_type VARCHAR(20) NOT NULL DEFAULT 'INTERNAL',
    account_status VARCHAR(20) DEFAULT 'ACTIVE',
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_login_at TIMESTAMPTZ,
    password_changed_at TIMESTAMPTZ DEFAULT NOW(),
    failed_login_attempts INTEGER DEFAULT 0,
    lockout_until TIMESTAMPTZ,
    session_timeout_minutes INTEGER DEFAULT 480, -- 8 hours
    mfa_enabled BOOLEAN DEFAULT false,
    mfa_secret VARCHAR(255),
    user_attributes JSONB DEFAULT '{}'::jsonb, -- ABAC attributes
    settings JSONB DEFAULT '{}'::jsonb, -- User preferences
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
```

**Key Features:**
- **Flexible linking**: Can link to persons/employees or exist independently
- **Multiple user types**: INTERNAL, CUSTOMER, VENDOR, PARTNER, API, SERVICE, ADMIN
- **Security controls**: Account lockout, session timeouts, MFA support
- **Service accounts**: API and system users without person association

### Session Management

```sql
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    refresh_token VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    device_info JSONB DEFAULT '{}'::jsonb, -- Device fingerprinting
    location_info JSONB DEFAULT '{}'::jsonb, -- Geographic context
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_accessed_at TIMESTAMPTZ DEFAULT NOW(),
    is_active BOOLEAN DEFAULT true
);
```

**Enhanced Features:**
- **Device fingerprinting**: Security context for ABAC evaluation
- **Location tracking**: Geographic and network location for access control
- **Security monitoring**: IP tracking and session analytics

## 🔐 Enhanced RBAC System

### Module-Based Organization

The system organizes functionality into modules for scalable permission management:

```sql
CREATE TABLE modules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    display_name VARCHAR(100),
    description TEXT,
    category VARCHAR(50), -- 'CORE', 'HR', 'FINANCE', 'SALES', etc.
    version VARCHAR(20),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Standard Module Categories:**
- **CORE**: System administration and basic functionality
- **HR**: Human resources management
- **FINANCE**: Financial operations and accounting
- **SALES**: Sales and customer management
- **INVENTORY**: Warehouse and inventory management
- **PROJECT**: Project management and tracking

### Resource Definition

Resources represent system objects that can be protected:

```sql
CREATE TABLE resources (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    module_id UUID NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    entity_id UUID REFERENCES entities(uuid),
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(150),
    description TEXT,
    resource_type VARCHAR(50) NOT NULL 
        CHECK (resource_type IN ('API', 'UI', 'DATA', 'FILE', 'REPORT', 'WORKFLOW', 'FUNCTION')),
    parent_resource_id UUID REFERENCES resources(id),
    path VARCHAR(500), -- URL path, API endpoint, file path
    resource_attributes JSONB DEFAULT '{}'::jsonb, -- ABAC attributes
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
```

**Resource Types:**
- **API**: REST endpoints and GraphQL operations
- **UI**: User interface components and pages
- **DATA**: Database tables and data objects
- **FILE**: Documents and file system resources
- **REPORT**: Generated reports and analytics
- **WORKFLOW**: Business process workflows
- **FUNCTION**: System functions and procedures

### Action Definition

Actions define what operations can be performed on resources:

```sql
CREATE TABLE actions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(150),
    description TEXT,
    action_type VARCHAR(50) NOT NULL 
        CHECK (action_type IN ('CREATE', 'READ', 'UPDATE', 'DELETE', 'EXECUTE', 'APPROVE', 'REJECT', 'EXPORT', 'IMPORT')),
    action_category VARCHAR(50) DEFAULT 'STANDARD'
        CHECK (action_category IN ('STANDARD', 'ADMINISTRATIVE', 'SENSITIVE', 'BULK', 'SYSTEM')),
    risk_level VARCHAR(20) DEFAULT 'LOW'
        CHECK (risk_level IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    requires_approval BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Risk-Based Actions:**
- **LOW**: Standard CRUD operations
- **MEDIUM**: Bulk operations and data exports
- **HIGH**: Administrative functions and approvals
- **CRITICAL**: System configuration and security changes

### Enhanced Role Management

```sql
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE RESTRICT,
    name VARCHAR(50) NOT NULL,
    display_name VARCHAR(100),
    description TEXT,
    module_id UUID REFERENCES modules(id), -- Module association
    role_type VARCHAR(20) DEFAULT 'CUSTOM'
        CHECK (role_type IN ('SYSTEM', 'TENANT', 'ENTITY', 'CUSTOM', 'FUNCTIONAL')),
    parent_role_id UUID REFERENCES roles(id), -- Role hierarchy
    level INTEGER DEFAULT 0, -- Calculated hierarchy level
    permissions JSONB NOT NULL DEFAULT '{}'::jsonb, -- Cached permissions
    entity_scope JSONB DEFAULT '{}'::jsonb, -- Entity access rules
    conditions JSONB DEFAULT '{}'::jsonb, -- Conditional access
    is_system_role BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
```

**Role Hierarchy Features:**
- **Inheritance**: Child roles inherit parent permissions
- **Scoping**: Entity-specific role assignments
- **Conditions**: Time, location, and device restrictions
- **Performance optimization**: Cached permission calculations

### Granular Permissions

```sql
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    resource_id UUID NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
    action_id UUID NOT NULL REFERENCES actions(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    display_name VARCHAR(250),
    description TEXT,
    effect VARCHAR(20) DEFAULT 'ALLOW' CHECK (effect IN ('ALLOW', 'DENY')),
    conditions JSONB DEFAULT '{}'::jsonb, -- ABAC conditions
    data_filters JSONB DEFAULT '{}'::jsonb, -- Row-level security
    field_restrictions JSONB DEFAULT '{}'::jsonb, -- Column restrictions
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Advanced Permission Features:**
- **ALLOW/DENY effects**: Explicit permission grants and denials
- **Data filters**: Row-level security with dynamic conditions
- **Field restrictions**: Column-level access control
- **ABAC integration**: Attribute-based conditional access

### User Role Assignments

```sql
CREATE TABLE user_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE RESTRICT,
    assignment_type VARCHAR(20) DEFAULT 'DIRECT'
        CHECK (assignment_type IN ('DIRECT', 'INHERITED', 'DELEGATED', 'TEMPORARY')),
    delegated_by UUID REFERENCES users(id),
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    assigned_by UUID REFERENCES users(id),
    expires_at TIMESTAMPTZ, -- Temporary assignments
    conditions JSONB DEFAULT '{}'::jsonb, -- Conditional access
    is_active BOOLEAN DEFAULT true
);
```

**Assignment Types:**
- **DIRECT**: Explicitly assigned by administrator
- **INHERITED**: Inherited from organizational hierarchy
- **DELEGATED**: Temporarily delegated by another user
- **TEMPORARY**: Time-limited assignments with auto-expiration

## 🎯 Advanced ABAC System

### Attribute Definitions

Standardized attributes for consistent ABAC evaluation:

```sql
CREATE TABLE attribute_definitions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(150),
    description TEXT,
    data_type VARCHAR(50) NOT NULL 
        CHECK (data_type IN ('STRING', 'NUMBER', 'BOOLEAN', 'DATE', 'TIME', 'JSON', 'ARRAY', 'ENUM')),
    category VARCHAR(50) NOT NULL 
        CHECK (category IN ('USER', 'RESOURCE', 'ENVIRONMENT', 'ACTION', 'ENTITY', 'SESSION')),
    is_required BOOLEAN DEFAULT false,
    is_sensitive BOOLEAN DEFAULT false, -- PII protection
    default_value TEXT,
    allowed_values JSONB, -- For enum types
    validation_rules JSONB DEFAULT '{}'::jsonb,
    encryption_required BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### ABAC Policies

Advanced policy engine with rule-based evaluation:

```sql
CREATE TABLE policies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID REFERENCES entities(uuid),
    name VARCHAR(150) NOT NULL,
    display_name VARCHAR(200),
    description TEXT,
    policy_type VARCHAR(50) DEFAULT 'ABAC' 
        CHECK (policy_type IN ('ABAC', 'RBAC', 'HYBRID', 'TIME_BASED', 'LOCATION_BASED')),
    effect VARCHAR(20) DEFAULT 'ALLOW' CHECK (effect IN ('ALLOW', 'DENY')),
    priority INTEGER DEFAULT 100, -- Conflict resolution
    category VARCHAR(50) DEFAULT 'ACCESS'
        CHECK (category IN ('ACCESS', 'DATA_FILTER', 'FIELD_MASK', 'AUDIT', 'COMPLIANCE')),
    target JSONB NOT NULL, -- When policy applies
    rule JSONB NOT NULL, -- Policy logic
    obligations JSONB DEFAULT '{}'::jsonb, -- Required actions
    advice JSONB DEFAULT '{}'::jsonb, -- Optional actions
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    deleted_at TIMESTAMPTZ
);
```

### Policy Examples

#### Time-Based Access Control
```go
type TimeBasedPolicy struct {
    Name       string            `json:"name"`
    PolicyType string            `json:"policy_type"`
    Effect     string            `json:"effect"`
    Target     PolicyTarget      `json:"target"`
    Rule       PolicyRule        `json:"rule"`
}

type PolicyTarget struct {
    Users     map[string]interface{} `json:"users"`
    Resources map[string]interface{} `json:"resources"`
    Actions   []string               `json:"actions"`
}

type PolicyRule struct {
    And []map[string]interface{} `json:"and,omitempty"`
    Or  []map[string]interface{} `json:"or,omitempty"`
}

// Example: Business hours access policy
businessHoursPolicy := TimeBasedPolicy{
    Name:       "business_hours_access",
    PolicyType: "TIME_BASED",
    Effect:     "ALLOW",
    Target: PolicyTarget{
        Users:     map[string]interface{}{"department": "finance"},
        Resources: map[string]interface{}{"type": "financial_reports"},
        Actions:   []string{"read", "export"},
    },
    Rule: PolicyRule{
        And: []map[string]interface{}{
            {"time_of_day": map[string][]string{"between": {"09:00", "17:00"}}},
            {"day_of_week": map[string][]int{"in": {1, 2, 3, 4, 5}}},
        },
    },
}
```

#### Location-Based Security
```go
type LocationBasedPolicy struct {
    Name       string       `json:"name"`
    PolicyType string       `json:"policy_type"`
    Effect     string       `json:"effect"`
    Target     PolicyTarget `json:"target"`
    Rule       PolicyRule   `json:"rule"`
}

// Example: Secure facility access policy
secureAccessPolicy := LocationBasedPolicy{
    Name:       "secure_facility_access",
    PolicyType: "LOCATION_BASED",
    Effect:     "ALLOW",
    Target: PolicyTarget{
        Resources: map[string]interface{}{"classification": "confidential"},
    },
    Rule: PolicyRule{
        Or: []map[string]interface{}{
            {"location": map[string]string{"within": "secure_facility"}},
            {"vpn_connection": true},
        },
    },
}
```

#### Hierarchical Access Control
```go
type HierarchicalPolicy struct {
    Name       string       `json:"name"`
    PolicyType string       `json:"policy_type"`
    Effect     string       `json:"effect"`
    Target     PolicyTarget `json:"target"`
    Rule       PolicyRule   `json:"rule"`
}

// Example: Manager employee data access policy
managerAccessPolicy := HierarchicalPolicy{
    Name:       "manager_employee_data",
    PolicyType: "ABAC",
    Effect:     "ALLOW",
    Target: PolicyTarget{
        Resources: map[string]interface{}{"type": "employee_record"},
        Actions:   []string{"read", "update"},
    },
    Rule: PolicyRule{
        Or: []map[string]interface{}{
            {"user.id": "resource.employee_id"},
            {"user.manager_id": "resource.manager_id"},
            {"user.security_level": map[string]int{"gte": 8}},
        },
    },
}
```

## 📊 Comprehensive Audit & Compliance

### Enhanced Audit Logging

```sql
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    event_category VARCHAR(50) DEFAULT 'ACCESS'
        CHECK (event_category IN ('ACCESS', 'ADMIN', 'DATA', 'AUTH', 'SYSTEM', 'COMPLIANCE')),
    severity VARCHAR(20) DEFAULT 'INFO'
        CHECK (severity IN ('LOW', 'INFO', 'WARN', 'HIGH', 'CRITICAL')),
    user_id UUID REFERENCES users(id),
    target_user_id UUID REFERENCES users(id),
    entity_id UUID REFERENCES entities(uuid),
    resource_id UUID REFERENCES resources(id),
    action_id UUID REFERENCES actions(id),
    role_id UUID REFERENCES roles(id),
    permission_id UUID REFERENCES permissions(id),
    decision VARCHAR(20), -- ALLOW/DENY
    reason TEXT,
    risk_score INTEGER DEFAULT 0, -- 0-100 risk calculation
    context JSONB DEFAULT '{}'::jsonb,
    ip_address INET,
    user_agent TEXT,
    session_id UUID REFERENCES user_sessions(id),
    compliance_flags JSONB DEFAULT '{}'::jsonb, -- GDPR, SOX, HIPAA, etc.
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Compliance Features:**
- **Risk scoring**: Automated risk assessment (0-100)
- **Compliance flags**: GDPR, SOX, HIPAA, PCI tracking
- **Complete context**: Session, location, device information
- **Regulatory support**: Immutable audit trail

### Access Request Workflow

```sql
CREATE TABLE access_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    requester_id UUID NOT NULL REFERENCES users(id),
    target_user_id UUID REFERENCES users(id),
    entity_id UUID NOT NULL REFERENCES entities(uuid),
    request_type VARCHAR(20) NOT NULL
        CHECK (request_type IN ('ROLE_ASSIGNMENT', 'PERMISSION_GRANT', 'RESOURCE_ACCESS', 'ELEVATION')),
    role_id UUID REFERENCES roles(id),
    permission_id UUID REFERENCES permissions(id),
    resource_id UUID REFERENCES resources(id),
    justification TEXT NOT NULL,
    business_reason VARCHAR(500),
    duration_hours INTEGER, -- Temporary access
    approval_status VARCHAR(20) DEFAULT 'PENDING'
        CHECK (approval_status IN ('PENDING', 'APPROVED', 'REJECTED', 'EXPIRED', 'REVOKED')),
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    approval_comments TEXT,
    expires_at TIMESTAMPTZ,
    auto_revoke BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

## 🔑 Authentication & Security

### Multi-Factor Authentication

Enhanced MFA implementation with multiple methods:

```go
type MFAConfiguration struct {
    Enabled            bool           `json:"enabled"`
    RequiredForRoles   []string       `json:"required_for_roles"`
    Methods            []MFAMethod    `json:"methods"`
    BackupCodesCount   int            `json:"backup_codes_count"`
    GracePeriodDays    int            `json:"grace_period_days"`
    RiskBasedTriggers  []RiskTrigger  `json:"risk_based_triggers"`
}

type MFAMethod struct {
    Type          string                 `json:"type"` // totp, sms, email, hardware_token
    Enabled       bool                   `json:"enabled"`
    Configuration map[string]interface{} `json:"configuration"`
}

type RiskTrigger struct {
    Condition              string `json:"condition"` // new_device, unusual_location, high_risk_action
    ForceMFA              bool   `json:"force_mfa"`
    AdditionalVerification bool   `json:"additional_verification"`
}
```

### Advanced Session Management

```go
type SessionConfiguration struct {
    MaxConcurrentSessions     int  `json:"max_concurrent_sessions"`
    SessionTimeoutMinutes     int  `json:"session_timeout_minutes"`
    IdleTimeoutMinutes        int  `json:"idle_timeout_minutes"`
    RememberMeDays           int  `json:"remember_me_days"`
    RequireReauthForSensitive bool `json:"require_reauth_for_sensitive"`
    IPBinding                bool `json:"ip_binding"`
    DeviceTracking           bool `json:"device_tracking"`
    LocationTracking         bool `json:"location_tracking"`
    RiskBasedTimeout         bool `json:"risk_based_timeout"`
}
```

### Enhanced Password Security

```go
type PasswordPolicy struct {
    MinLength               int            `json:"min_length"`
    MaxLength               int            `json:"max_length"`
    RequireUppercase        bool           `json:"require_uppercase"`
    RequireLowercase        bool           `json:"require_lowercase"`
    RequireNumbers          bool           `json:"require_numbers"`
    RequireSymbols          bool           `json:"require_symbols"`
    MinSymbols              int            `json:"min_symbols"`
    MaxAgeDays              int            `json:"max_age_days"`
    HistoryCount            int            `json:"history_count"`
    LockoutAttempts         int            `json:"lockout_attempts"`
    LockoutDurationMinutes  int            `json:"lockout_duration_minutes"`
    ComplexityScoreMin      int            `json:"complexity_score_min"`
    BreachDetection         bool           `json:"breach_detection"`
    CustomRules             []PasswordRule `json:"custom_rules"`
}

type PasswordRule struct {
    Name        string `json:"name"`
    Pattern     string `json:"pattern"`
    Description string `json:"description"`
    Required    bool   `json:"required"`
}
```

## 👥 User Lifecycle Management

### Enhanced Onboarding Workflow

```yaml
user_onboarding:
  invitation_process:
    1. admin_sends_invitation:
        - validate_email_format
        - check_existing_users_across_types
        - generate_secure_invitation_token
        - create_pending_person_record
        - send_invitation_email_with_expiry
        - log_invitation_audit_event
    
    2. user_accepts_invitation:
        - validate_invitation_token_and_expiry
        - collect_comprehensive_user_information
        - create_person_record_if_needed
        - create_employee_record_if_applicable
        - create_user_account_with_security_attributes
        - set_initial_password_with_policy_validation
        - verify_email_address_with_token
    
    3. initial_setup:
        - assign_default_roles_based_on_position
        - set_entity_and_department_associations
        - configure_security_attributes_for_abac
        - setup_mfa_based_on_risk_profile
        - initialize_user_preferences
        - send_welcome_email_with_resources
    
    4. first_login:
        - force_password_change_if_temporary
        - complete_profile_setup_wizard
        - accept_terms_of_service_with_version_tracking
        - configure_notification_preferences
        - setup_security_questions_if_required
        - start_guided_tour_based_on_role
        - trigger_manager_notification
```

### Advanced Offboarding Process

```go

// Event type constants (replacing TS union types)
const (
	EventTypeEmploymentTerminated = "employment_terminated"
	EventTypeResignationSubmitted = "resignation_submitted"
	EventTypeContractExpired      = "contract_expired"
	EventTypeSecurityViolation    = "security_violation"
)

// Step type constants
const (
	StepTypeAutomatic       = "automatic"
	StepTypeManual         = "manual"
	StepTypeApprovalRequired = "approval_required"
)

// Action type constants
const (
	ActionTypeRevokeAllAccess     = "revoke_all_access"
	ActionTypeTransferOwnership   = "transfer_ownership"
	ActionTypeArchivePersonalData = "archive_personal_data"
)

// Compliance flag constants
const (
	ComplianceSOX          = "SOX"
	ComplianceGDPR         = "GDPR"
	ComplianceCCPA         = "CCPA"
	ComplianceDataRetention = "DATA_RETENTION"
)

// Core structs
type OffboardingProcess struct {
	TriggerEvents          []TriggerEvent           `json:"trigger_events"`
	Steps                  []OffboardingStep        `json:"steps"`
	DataRetentionPolicy    DataRetentionPolicy      `json:"data_retention_policy"`
	AccessRevocation       AccessRevocationPolicy   `json:"access_revocation"`
	ComplianceRequirements []ComplianceRequirement  `json:"compliance_requirements"`
}

type TriggerEvent struct {
	EventType              string   `json:"event_type"`
	AutoTrigger            bool     `json:"auto_trigger"`
	GracePeriodHours       int      `json:"grace_period_hours"`
	NotificationRecipients []string `json:"notification_recipients"`
}

type OffboardingStep struct {
	Name            string              `json:"name"`
	Type            string              `json:"type"`
	Assignee        *string             `json:"assignee,omitempty"` // pointer for optional field
	SLAHours        int                 `json:"sla_hours"`
	Dependencies    []string            `json:"dependencies"`
	Actions         []OffboardingAction `json:"actions"`
	ComplianceFlags []string            `json:"compliance_flags"`
}

type OffboardingAction struct {
	ActionType       string                 `json:"action_type"`
	Parameters       map[string]interface{} `json:"parameters"`
	AuditTrail       bool                   `json:"audit_trail"`
	RollbackPossible bool                   `json:"rollback_possible"`
}

type DataRetentionPolicy struct {
	PersonalData    string `json:"personal_data"`
	WorkDocuments   string `json:"work_documents"`
	AuditLogs       string `json:"audit_logs"`
}

type AccessRevocationPolicy struct {
	ImmediateRevocation   []string `json:"immediate_revocation"`
	DelayedRevocation     []string `json:"delayed_revocation"`
	ManagerReviewRequired []string `json:"manager_review_required"`
}

type ComplianceRequirement struct {
	Standard string   `json:"standard"`
	Actions  []string `json:"actions"`
	Timeline string   `json:"timeline"`
}

// Helper function to create string pointer (idiomatic Go pattern for optional strings)
func stringPtr(s string) *string {
	return &s
}

// GetComprehensiveOffboarding returns the comprehensive offboarding process
// This replaces the const object literal in TypeScript
func GetComprehensiveOffboarding() OffboardingProcess {
	return OffboardingProcess{
		TriggerEvents: []TriggerEvent{
			{
				EventType:              EventTypeEmploymentTerminated,
				AutoTrigger:            true,
				GracePeriodHours:       0,
				NotificationRecipients: []string{"hr_team", "security_team", "manager"},
			},
		},
		
		Steps: []OffboardingStep{
			{
				Name:         "immediate_access_revocation",
				Type:         StepTypeAutomatic,
				Assignee:     nil, // explicitly nil for optional field
				SLAHours:     0,
				Dependencies: []string{}, // empty slice instead of null
				Actions: []OffboardingAction{
					{
						ActionType: ActionTypeRevokeAllAccess,
						Parameters: map[string]interface{}{
							"disable_login":               true,
							"revoke_all_tokens":          true,
							"disable_api_access":         true,
							"revoke_delegated_permissions": true,
						},
						AuditTrail:       true,
						RollbackPossible: true,
					},
				},
				ComplianceFlags: []string{ComplianceSOX, ComplianceGDPR},
			},
			
			{
				Name:         "data_ownership_transfer",
				Type:         StepTypeManual,
				Assignee:     stringPtr("manager"), // using helper function
				SLAHours:     24,
				Dependencies: []string{"immediate_access_revocation"},
				Actions: []OffboardingAction{
					{
						ActionType: ActionTypeTransferOwnership,
						Parameters: map[string]interface{}{
							"transfer_to":         "manager",
							"include_documents":   true,
							"include_projects":    true,
							"notify_stakeholders": true,
						},
						AuditTrail:       true,
						RollbackPossible: true,
					},
				},
				ComplianceFlags: []string{ComplianceDataRetention},
			},
			
			{
				Name:         "compliance_data_handling",
				Type:         StepTypeAutomatic,
				Assignee:     nil,
				SLAHours:     72,
				Dependencies: []string{"data_ownership_transfer"},
				Actions: []OffboardingAction{
					{
						ActionType: ActionTypeArchivePersonalData,
						Parameters: map[string]interface{}{
							"retention_policy":     "employee_data_retention",
							"anonymize_analytics":  true,
							"preserve_audit_trail": true,
						},
						AuditTrail:       true,
						RollbackPossible: false,
					},
				},
				ComplianceFlags: []string{ComplianceGDPR, ComplianceCCPA},
			},
		},
		
		// You would typically set these fields as well in a complete implementation
		DataRetentionPolicy: DataRetentionPolicy{
			PersonalData:  "30_days",
			WorkDocuments: "7_years",
			AuditLogs:     "permanent",
		},
		
		AccessRevocation: AccessRevocationPolicy{
			ImmediateRevocation:   []string{"system_access", "email_access", "vpn_access"},
			DelayedRevocation:     []string{"building_access", "company_devices"},
			ManagerReviewRequired: []string{"project_ownership", "client_relationships"},
		},
		
		ComplianceRequirements: []ComplianceRequirement{
			{
				Standard: ComplianceGDPR,
				Actions:  []string{"data_inventory", "consent_withdrawal", "data_deletion"},
				Timeline: "30_days",
			},
			{
				Standard: ComplianceSOX,
				Actions:  []string{"access_certification", "financial_data_protection"},
				Timeline: "immediate",
			},
		},
	}
}

// Alternative constructor pattern (common in Go)
func NewOffboardingProcess() *OffboardingProcess {
	process := GetComprehensiveOffboarding()
	return &process
}

// Validation methods (idiomatic Go pattern)
func (p *OffboardingProcess) Validate() error {
	if len(p.TriggerEvents) == 0 {
		return fmt.Errorf("at least one trigger event is required")
	}
	
	if len(p.Steps) == 0 {
		return fmt.Errorf("at least one step is required")
	}
	
	// Validate each step
	for i, step := range p.Steps {
		if err := step.Validate(); err != nil {
			return fmt.Errorf("step %d validation failed: %w", i, err)
		}
	}
	
	return nil
}

func (s *OffboardingStep) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("step name is required")
	}
	
	if s.Type != StepTypeAutomatic && s.Type != StepTypeManual && s.Type != StepTypeApprovalRequired {
		return fmt.Errorf("invalid step type: %s", s.Type)
	}
	
	if s.Type == StepTypeManual && s.Assignee == nil {
		return fmt.Errorf("manual steps require an assignee")
	}
	
	return nil
}

// Method to check if a step depends on another step
func (s *OffboardingStep) DependsOn(stepName string) bool {
	for _, dep := range s.Dependencies {
		if dep == stepName {
			return true
		}
	}
	return false
}

// Method to add compliance flag
func (s *OffboardingStep) AddComplianceFlag(flag string) {
	// Check if flag already exists to avoid duplicates
	for _, existing := range s.ComplianceFlags {
		if existing == flag {
			return
		}
	}
	s.ComplianceFlags = append(s.ComplianceFlags, flag)
}
```

## 📈 User Analytics & Monitoring

### Advanced User Activity Tracking

```sql
-- Partitioned audit table for performance
CREATE TABLE user_activities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    session_id UUID REFERENCES user_sessions(id),
    
    -- Activity classification
    activity_type VARCHAR(50) NOT NULL,
    module VARCHAR(50),
    resource_type VARCHAR(50),
    resource_id UUID,
    action_performed VARCHAR(50),
    
    -- Security context
    ip_address INET,
    user_agent TEXT,
    device_fingerprint VARCHAR(255),
    location_data JSONB,
    
    -- Performance metrics
    request_method VARCHAR(10),
    request_path TEXT,
    request_params JSONB,
    response_status INTEGER,
    response_time_ms INTEGER,
    
    -- Risk assessment
    risk_indicators JSONB DEFAULT '{}'::jsonb,
    anomaly_score DECIMAL(5,2) DEFAULT 0.00,
    
    -- Metadata
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    additional_data JSONB DEFAULT '{}'::jsonb
) PARTITION BY RANGE (timestamp);
```

### Behavioral Analytics

```go
type DateRange struct {
    Start time.Time `json:"start"`
    End   time.Time `json:"end"`
}

type ResourceUsage struct {
    ResourceName string `json:"resource_name"`
    AccessCount  int    `json:"access_count"`
    LastAccessed time.Time `json:"last_accessed"`
}

type DeviceType struct {
    Type  string `json:"type"`
    Count int    `json:"count"`
}

type Anomaly struct {
    Type        string    `json:"type"`
    Severity    string    `json:"severity"`
    Description string    `json:"description"`
    DetectedAt  time.Time `json:"detected_at"`
    Score       float64   `json:"score"`
}

type SecurityAction struct {
    Action      string `json:"action"`
    Priority    string `json:"priority"`
    Description string `json:"description"`
    Automated   bool   `json:"automated"`
}

type UserBehaviorAnalytics struct {
    UserID         string    `json:"user_id"`
    TenantID       string    `json:"tenant_id"`
    AnalysisPeriod DateRange `json:"analysis_period"`
    
    ActivityPatterns struct {
        PeakUsageHours        []int           `json:"peak_usage_hours"`
        CommonLocations       []string        `json:"common_locations"`
        TypicalSessionDuration int            `json:"typical_session_duration"`
        FrequentResources     []ResourceUsage `json:"frequent_resources"`
        DevicePreferences     []DeviceType    `json:"device_preferences"`
    } `json:"activity_patterns"`
    
    ProductivityMetrics struct {
        DocumentsCreated       int     `json:"documents_created"`
        TransactionsProcessed  int     `json:"transactions_processed"`
        ApprovalsCompleted     int     `json:"approvals_completed"`
        CollaborationScore     float64 `json:"collaboration_score"`
        FeatureAdoptionRate    float64 `json:"feature_adoption_rate"`
    } `json:"productivity_metrics"`
    
    SecurityProfile struct {
        LoginPatternConsistency float64 `json:"login_pattern_consistency"`
        LocationVariance        float64 `json:"location_variance"`
        DeviceConsistency       float64 `json:"device_consistency"`
        PermissionUsageRate     float64 `json:"permission_usage_rate"`
        PolicyViolationCount    int     `json:"policy_violation_count"`
    } `json:"security_profile"`
    
    RiskAssessment struct {
        OverallRiskScore      int              `json:"overall_risk_score"`
        BehavioralAnomalies   []Anomaly        `json:"behavioral_anomalies"`
        RecommendedActions    []SecurityAction `json:"recommended_actions"`
    } `json:"risk_assessment"`
}
```

## 🔧 Advanced Permission Management API

### Comprehensive User Management Endpoints

```go
package handlers

import (
    "encoding/json"
    "net/http"
    "time"
    
    "github.com/gorilla/mux"
    "github.com/google/uuid"
)

// Request/Response structures
type PersonRequest struct {
    FirstName          string                 `json:"first_name" validate:"required"`
    LastName           string                 `json:"last_name" validate:"required"`
    MiddleName         *string                `json:"middle_name,omitempty"`
    Email              *string                `json:"email,omitempty"`
    Phone              *string                `json:"phone,omitempty"`
    PersonType         string                 `json:"person_type" validate:"required"`
    SecurityAttributes map[string]interface{} `json:"security_attributes,omitempty"`
    Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

type EmployeeRequest struct {
    PersonID         uuid.UUID              `json:"person_id" validate:"required"`
    EmployeeNumber   string                 `json:"employee_number" validate:"required"`
    PositionTitle    *string                `json:"position_title,omitempty"`
    DepartmentID     *uuid.UUID             `json:"department_id,omitempty"`
    ManagerID        *uuid.UUID             `json:"manager_id,omitempty"`
    HireDate         time.Time              `json:"hire_date" validate:"required"`
    SecurityLevel    int                    `json:"security_level"`
    AccessAttributes map[string]interface{} `json:"access_attributes,omitempty"`
}

type UserRequest struct {
    Username       *string                `json:"username,omitempty"`
    Email          string                 `json:"email" validate:"required,email"`
    UserType       string                 `json:"user_type" validate:"required"`
    PersonID       *uuid.UUID             `json:"person_id,omitempty"`
    EmployeeID     *uuid.UUID             `json:"employee_id,omitempty"`
    UserAttributes map[string]interface{} `json:"user_attributes,omitempty"`
    Settings       map[string]interface{} `json:"settings,omitempty"`
}

type RoleRequest struct {
    Name         string                 `json:"name" validate:"required"`
    DisplayName  *string                `json:"display_name,omitempty"`
    Description  *string                `json:"description,omitempty"`
    RoleType     string                 `json:"role_type"`
    ModuleID     *uuid.UUID             `json:"module_id,omitempty"`
    ParentRoleID *uuid.UUID             `json:"parent_role_id,omitempty"`
    EntityScope  map[string]interface{} `json:"entity_scope,omitempty"`
    Conditions   map[string]interface{} `json:"conditions,omitempty"`
}

type PermissionEvaluationRequest struct {
    UserID       uuid.UUID              `json:"user_id" validate:"required"`
    ResourceName string                 `json:"resource_name" validate:"required"`
    ActionName   string                 `json:"action_name" validate:"required"`
    EntityID     *uuid.UUID             `json:"entity_id,omitempty"`
    Context      map[string]interface{} `json:"context,omitempty"`
}

type PermissionEvaluationResponse struct {
    Allowed           bool     `json:"allowed"`
    PolicyDecisions   []string `json:"policy_decisions"`
    EffectiveRoles    []string `json:"effective_roles"`
    EvaluationTimeMS  int      `json:"evaluation_time_ms"`
    CacheHit          bool     `json:"cache_hit"`
}

type AccessRequestRequest struct {
    TargetUserID     *uuid.UUID `json:"target_user_id,omitempty"`
    RequestType      string     `json:"request_type" validate:"required"`
    RoleID           *uuid.UUID `json:"role_id,omitempty"`
    PermissionID     *uuid.UUID `json:"permission_id,omitempty"`
    ResourceID       *uuid.UUID `json:"resource_id,omitempty"`
    Justification    string     `json:"justification" validate:"required"`
    BusinessReason   *string    `json:"business_reason,omitempty"`
    DurationHours    *int       `json:"duration_hours,omitempty"`
}

// API Handler Interface
type UserManagementAPI interface {
    // Person Management
    ListPersons(w http.ResponseWriter, r *http.Request)      // GET /api/v1/persons
    CreatePerson(w http.ResponseWriter, r *http.Request)     // POST /api/v1/persons
    GetPerson(w http.ResponseWriter, r *http.Request)        // GET /api/v1/persons/{person_id}
    UpdatePerson(w http.ResponseWriter, r *http.Request)     // PUT /api/v1/persons/{person_id}
    DeletePerson(w http.ResponseWriter, r *http.Request)     // DELETE /api/v1/persons/{person_id}
    
    // Employee Management
    ListEmployees(w http.ResponseWriter, r *http.Request)              // GET /api/v1/employees
    CreateEmployee(w http.ResponseWriter, r *http.Request)             // POST /api/v1/employees
    GetEmployee(w http.ResponseWriter, r *http.Request)                // GET /api/v1/employees/{employee_id}
    UpdateEmployee(w http.ResponseWriter, r *http.Request)             // PUT /api/v1/employees/{employee_id}
    GetEmployeeSubordinates(w http.ResponseWriter, r *http.Request)    // GET /api/v1/employees/{employee_id}/subordinates
    
    // User CRUD Operations
    ListUsers(w http.ResponseWriter, r *http.Request)        // GET /api/v1/users
    CreateUser(w http.ResponseWriter, r *http.Request)       // POST /api/v1/users
    GetUser(w http.ResponseWriter, r *http.Request)          // GET /api/v1/users/{user_id}
    UpdateUser(w http.ResponseWriter, r *http.Request)       // PUT /api/v1/users/{user_id}
    DeactivateUser(w http.ResponseWriter, r *http.Request)   // DELETE /api/v1/users/{user_id}
    
    // Advanced Role Management
    ListRoles(w http.ResponseWriter, r *http.Request)                // GET /api/v1/roles
    CreateRole(w http.ResponseWriter, r *http.Request)               // POST /api/v1/roles
    GetRole(w http.ResponseWriter, r *http.Request)                  // GET /api/v1/roles/{role_id}
    UpdateRole(w http.ResponseWriter, r *http.Request)               // PUT /api/v1/roles/{role_id}
    GetRoleInheritance(w http.ResponseWriter, r *http.Request)       // GET /api/v1/roles/{role_id}/inheritance
    
    // Permission Evaluation
    EvaluatePermission(w http.ResponseWriter, r *http.Request)        // POST /api/v1/permissions/evaluate
    GetUserEffectivePermissions(w http.ResponseWriter, r *http.Request) // GET /api/v1/users/{user_id}/effective-permissions
    BulkEvaluatePermissions(w http.ResponseWriter, r *http.Request)   // POST /api/v1/permissions/bulk-evaluate
    
    // ABAC Policy Management
    ListPolicies(w http.ResponseWriter, r *http.Request)     // GET /api/v1/policies
    CreatePolicy(w http.ResponseWriter, r *http.Request)     // POST /api/v1/policies
    TestPolicy(w http.ResponseWriter, r *http.Request)       // POST /api/v1/policies/test
    GetPolicyImpact(w http.ResponseWriter, r *http.Request)  // GET /api/v1/policies/{policy_id}/impact
    
    // Access Request Workflow
    SubmitAccessRequest(w http.ResponseWriter, r *http.Request)      // POST /api/v1/access-requests
    ListAccessRequests(w http.ResponseWriter, r *http.Request)       // GET /api/v1/access-requests
    ApproveAccessRequest(w http.ResponseWriter, r *http.Request)     // POST /api/v1/access-requests/{request_id}/approve
    RejectAccessRequest(w http.ResponseWriter, r *http.Request)      // POST /api/v1/access-requests/{request_id}/reject
    
    // Advanced Authentication
    Login(w http.ResponseWriter, r *http.Request)                    // POST /api/v1/auth/login
    StepUpAuthentication(w http.ResponseWriter, r *http.Request)     // POST /api/v1/auth/step-up
    AssessLoginRisk(w http.ResponseWriter, r *http.Request)          // POST /api/v1/auth/risk-assessment
    
    // Comprehensive Audit & Analytics
    QueryActivityLogs(w http.ResponseWriter, r *http.Request)        // GET /api/v1/audit/activities
    GetSecurityEvents(w http.ResponseWriter, r *http.Request)        // GET /api/v1/audit/security-events
    GetUserBehaviorAnalytics(w http.ResponseWriter, r *http.Request) // GET /api/v1/analytics/user-behavior
    AnalyzeAccessPatterns(w http.ResponseWriter, r *http.Request)    // GET /api/v1/analytics/access-patterns
    GenerateRiskAssessment(w http.ResponseWriter, r *http.Request)   // POST /api/v1/analytics/risk-assessment
    
    // Compliance & Reporting
    GetGDPRData(w http.ResponseWriter, r *http.Request)              // GET /api/v1/compliance/gdpr-data
    ExportUserData(w http.ResponseWriter, r *http.Request)           // POST /api/v1/compliance/data-export
    ProcessRightToBeForgotten(w http.ResponseWriter, r *http.Request) // POST /api/v1/compliance/right-to-be-forgotten
    GenerateAccessReview(w http.ResponseWriter, r *http.Request)     // GET /api/v1/reports/access-review
}

// Example handler implementation
type UserManagementHandler struct {
    userService       UserService
    permissionService PermissionService
    auditService      AuditService
}

func (h *UserManagementHandler) EvaluatePermission(w http.ResponseWriter, r *http.Request) {
    var req PermissionEvaluationRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    startTime := time.Now()
    
    // Evaluate permission using ABAC engine
    result, err := h.permissionService.EvaluatePermission(r.Context(), req)
    if err != nil {
        http.Error(w, "Permission evaluation failed", http.StatusInternalServerError)
        return
    }
    
    evaluationTime := int(time.Since(startTime).Milliseconds())
    
    response := PermissionEvaluationResponse{
        Allowed:          result.Allowed,
        PolicyDecisions:  result.PolicyDecisions,
        EffectiveRoles:   result.EffectiveRoles,
        EvaluationTimeMS: evaluationTime,
        CacheHit:         result.CacheHit,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
    
    // Audit the permission check
    h.auditService.LogPermissionEvaluation(r.Context(), req, response)
}

// Router setup example
func SetupUserManagementRoutes(r *mux.Router, handler UserManagementAPI) {
    // Person Management
    r.HandleFunc("/api/v1/persons", handler.ListPersons).Methods("GET")
    r.HandleFunc("/api/v1/persons", handler.CreatePerson).Methods("POST")
    r.HandleFunc("/api/v1/persons/{person_id}", handler.GetPerson).Methods("GET")
    r.HandleFunc("/api/v1/persons/{person_id}", handler.UpdatePerson).Methods("PUT")
    r.HandleFunc("/api/v1/persons/{person_id}", handler.DeletePerson).Methods("DELETE")
    
    // Employee Management
    r.HandleFunc("/api/v1/employees", handler.ListEmployees).Methods("GET")
    r.HandleFunc("/api/v1/employees", handler.CreateEmployee).Methods("POST")
    r.HandleFunc("/api/v1/employees/{employee_id}", handler.GetEmployee).Methods("GET")
    r.HandleFunc("/api/v1/employees/{employee_id}", handler.UpdateEmployee).Methods("PUT")
    r.HandleFunc("/api/v1/employees/{employee_id}/subordinates", handler.GetEmployeeSubordinates).Methods("GET")
    
    // User CRUD Operations
    r.HandleFunc("/api/v1/users", handler.ListUsers).Methods("GET")
    r.HandleFunc("/api/v1/users", handler.CreateUser).Methods("POST")
    r.HandleFunc("/api/v1/users/{user_id}", handler.GetUser).Methods("GET")
    r.HandleFunc("/api/v1/users/{user_id}", handler.UpdateUser).Methods("PUT")
    r.HandleFunc("/api/v1/users/{user_id}", handler.DeactivateUser).Methods("DELETE")
    
    // Permission Evaluation
    r.HandleFunc("/api/v1/permissions/evaluate", handler.EvaluatePermission).Methods("POST")
    r.HandleFunc("/api/v1/users/{user_id}/effective-permissions", handler.GetUserEffectivePermissions).Methods("GET")
    r.HandleFunc("/api/v1/permissions/bulk-evaluate", handler.BulkEvaluatePermissions).Methods("POST")
    
    // Access Request Workflow
    r.HandleFunc("/api/v1/access-requests", handler.SubmitAccessRequest).Methods("POST")
    r.HandleFunc("/api/v1/access-requests", handler.ListAccessRequests).Methods("GET")
    r.HandleFunc("/api/v1/access-requests/{request_id}/approve", handler.ApproveAccessRequest).Methods("POST")
    r.HandleFunc("/api/v1/access-requests/{request_id}/reject", handler.RejectAccessRequest).Methods("POST")
    
    // Analytics and Reporting
    r.HandleFunc("/api/v1/analytics/user-behavior", handler.GetUserBehaviorAnalytics).Methods("GET")
    r.HandleFunc("/api/v1/analytics/access-patterns", handler.AnalyzeAccessPatterns).Methods("GET")
    r.HandleFunc("/api/v1/analytics/risk-assessment", handler.GenerateRiskAssessment).Methods("POST")
    
    // Compliance
    r.HandleFunc("/api/v1/compliance/gdpr-data", handler.GetGDPRData).Methods("GET")
    r.HandleFunc("/api/v1/compliance/data-export", handler.ExportUserData).Methods("POST")
    r.HandleFunc("/api/v1/reports/access-review", handler.GenerateAccessReview).Methods("GET")
}
```

## 🚀 Performance Optimization Features

### Intelligent Caching

```sql
-- Policy evaluation cache for performance
CREATE TABLE policy_evaluations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    resource_id UUID NOT NULL REFERENCES resources(id),
    action_id UUID NOT NULL REFERENCES actions(id),
    context_hash VARCHAR(64) NOT NULL, -- SHA-256 of evaluation context
    decision VARCHAR(20) NOT NULL CHECK (decision IN ('ALLOW', 'DENY', 'NOT_APPLICABLE')),
    applicable_policies UUID[] DEFAULT '{}', -- Policy IDs that fired
    evaluation_time_ms INTEGER, -- Performance metric
    evaluated_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ DEFAULT (NOW() + INTERVAL '1 hour')
);
```

### Utility Views for Performance

```sql
-- Comprehensive user view with aggregated data
CREATE VIEW user_complete_view AS
SELECT 
    u.id as user_id,
    u.tenant_id,
    u.username,
    u.email,
    u.user_type,
    u.account_status,
    p.first_name,
    p.last_name,
    e.employee_number,
    e.security_level,
    -- Combined ABAC attributes
    COALESCE(p.security_attributes, '{}'::jsonb) || 
    COALESCE(e.access_attributes, '{}'::jsonb) || 
    COALESCE(u.user_attributes, '{}'::jsonb) as combined_attributes,
    -- Role aggregations
    array_agg(DISTINCT r.name ORDER BY r.name) as role_names,
    count(DISTINCT ur.id) FILTER (WHERE ur.is_active = true) as active_role_count
FROM users u
LEFT JOIN persons p ON u.person_id = p.id
LEFT JOIN employees e ON u.employee_id = e.id
LEFT JOIN user_roles ur ON u.id = ur.user_id AND ur.is_active = true
LEFT JOIN roles r ON ur.role_id = r.id AND r.is_active = true
WHERE u.deleted_at IS NULL
GROUP BY u.id, u.tenant_id, u.username, u.email, u.user_type, 
         u.account_status, p.first_name, p.last_name, e.employee_number, 
         e.security_level, p.security_attributes, e.access_attributes, u.user_attributes;
```

## 💡 Key System Features

### Row Level Security (RLS)
- **Automatic tenant isolation**: All tables enforce tenant-based access
- **Performance optimized**: Efficient policy evaluation with indexes
- **Flexible scoping**: Entity and department-level isolation

### Advanced ABAC Engine
- **Real-time evaluation**: Sub-millisecond policy decisions with caching
- **Rich context**: User, resource, environment, and session attributes
- **Policy conflicts**: Priority-based resolution with explicit DENY support

### Compliance Ready
- **Audit completeness**: Immutable audit trail with risk scoring
- **Regulatory support**: GDPR, SOX, HIPAA, PCI compliance features
- **Data governance**: Retention policies and automated cleanup

### Enterprise Scale
- **UUID primary keys**: Distributed system friendly
- **Partitioned tables**: Time-based partitioning for large datasets
- **Optimized indexes**: JSONB GIN indexes for ABAC attributes
- **Maintenance functions**: Automated cleanup and cache invalidation

This comprehensive user management and permissions system provides enterprise-grade security, scalability, and compliance while maintaining flexibility for complex organizational structures and access patterns.
