# User Management & Permissions

## 👥 Overview

The ERP system implements a comprehensive user management and permissions framework based on Role-Based Access Control (RBAC) with support for Attribute-Based Access Control (ABAC) for complex scenarios. The system provides fine-grained control over user access while maintaining simplicity for administrators.

## 🏗️ User Management Architecture

### User Entity Model

```sql
-- Core user table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    email_verified BOOLEAN DEFAULT false,
    
    -- Personal information
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    middle_name VARCHAR(100),
    display_name VARCHAR(200),
    avatar_url VARCHAR(500),
    
    -- Authentication
    password_hash VARCHAR(255),
    password_salt VARCHAR(255),
    password_last_changed TIMESTAMPTZ,
    password_reset_token VARCHAR(255),
    password_reset_expires TIMESTAMPTZ,
    
    -- Multi-factor authentication
    mfa_enabled BOOLEAN DEFAULT false,
    mfa_secret VARCHAR(255),
    mfa_backup_codes TEXT[],
    
    -- Account status
    status VARCHAR(20) DEFAULT 'active', -- active, inactive, suspended, locked
    last_login_at TIMESTAMPTZ,
    last_login_ip INET,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMPTZ,
    
    -- Contact information
    phone VARCHAR(20),
    mobile VARCHAR(20),
    address JSONB,
    
    -- Preferences
    language VARCHAR(10) DEFAULT 'en',
    timezone VARCHAR(50) DEFAULT 'UTC',
    date_format VARCHAR(20) DEFAULT 'YYYY-MM-DD',
    currency_code CHAR(3),
    
    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT valid_status CHECK (status IN ('active', 'inactive', 'suspended', 'locked')),
    CONSTRAINT valid_email CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

-- User sessions for tracking active sessions
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    refresh_token VARCHAR(255) UNIQUE,
    ip_address INET,
    user_agent TEXT,
    device_info JSONB,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_activity TIMESTAMPTZ DEFAULT NOW(),
    is_active BOOLEAN DEFAULT true
);

-- User preferences and settings
CREATE TABLE user_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category VARCHAR(50) NOT NULL, -- ui, notifications, security, etc.
    settings JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(user_id, category)
);
```

### User-Organization Relationships

```sql
-- User organization assignments
CREATE TABLE user_organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    -- Employment details
    employee_id VARCHAR(50), -- company employee ID
    job_title VARCHAR(100),
    department VARCHAR(100),
    manager_id UUID REFERENCES users(id),
    
    -- Employment status
    employment_type VARCHAR(50), -- full_time, part_time, contract, consultant
    employment_status VARCHAR(50), -- active, on_leave, terminated
    hire_date DATE,
    termination_date DATE,
    
    -- Access control
    is_primary_org BOOLEAN DEFAULT false,
    access_level VARCHAR(50) DEFAULT 'standard', -- admin, manager, standard, readonly
    
    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(user_id, organization_id),
    CONSTRAINT valid_employment_type CHECK (employment_type IN ('full_time', 'part_time', 'contract', 'consultant')),
    CONSTRAINT valid_employment_status CHECK (employment_status IN ('active', 'on_leave', 'terminated'))
);
```

## 🔐 Permission System

### Role-Based Access Control (RBAC)

#### Role Definition

```sql
-- System-wide roles
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    role_type VARCHAR(50) DEFAULT 'custom', -- system, tenant, custom
    is_active BOOLEAN DEFAULT true,
    
    -- Role metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    UNIQUE(tenant_id, name),
    CONSTRAINT valid_role_type CHECK (role_type IN ('system', 'tenant', 'custom'))
);

-- Permissions catalog
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    module VARCHAR(50) NOT NULL, -- financial, inventory, hr, etc.
    resource VARCHAR(50) NOT NULL, -- invoices, purchase_orders, employees
    action VARCHAR(50) NOT NULL, -- create, read, update, delete, approve
    description TEXT,
    
    -- Permission categorization
    permission_type VARCHAR(50) DEFAULT 'data', -- data, system, ui
    scope VARCHAR(50) DEFAULT 'tenant', -- system, tenant, organization, user
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_permission_type CHECK (permission_type IN ('data', 'system', 'ui')),
    CONSTRAINT valid_scope CHECK (scope IN ('system', 'tenant', 'organization', 'user'))
);

-- Role-permission assignments
CREATE TABLE role_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    
    -- Conditional permissions
    conditions JSONB DEFAULT '{}',
    granted BOOLEAN DEFAULT true,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(role_id, permission_id)
);

-- User-role assignments
CREATE TABLE user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    
    -- Role scope and conditions
    scope_type VARCHAR(50) DEFAULT 'organization', -- global, organization, department
    scope_id UUID, -- organization_id, department_id, etc.
    
    -- Temporal access
    granted_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    
    -- Assignment metadata
    granted_by UUID REFERENCES users(id),
    reason TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(user_id, role_id, organization_id),
    CONSTRAINT valid_scope_type CHECK (scope_type IN ('global', 'organization', 'department', 'project'))
);
```

#### Standard System Roles

```yaml
system_roles:
  super_admin:
    description: "Complete system access across all tenants"
    permissions: ["*"]
    scope: system
    
  tenant_admin:
    description: "Full administrative access within tenant"
    permissions:
      - "tenant:manage"
      - "users:manage"
      - "roles:manage"
      - "organizations:manage"
      - "settings:manage"
      - "*:read"
      - "*:write"
    scope: tenant
    
  organization_admin:
    description: "Administrative access within organization"
    permissions:
      - "organization:manage"
      - "users:invite"
      - "users:read"
      - "departments:manage"
      - "projects:manage"
    scope: organization
    
  financial_manager:
    description: "Financial operations management"
    permissions:
      - "accounts:read"
      - "accounts:write"
      - "invoices:create"
      - "invoices:approve"
      - "payments:process"
      - "reports:financial"
    scope: organization
    
  inventory_manager:
    description: "Inventory and warehouse operations"
    permissions:
      - "inventory:read"
      - "inventory:write"
      - "warehouses:manage"
      - "purchase_orders:create"
      - "stock_transfers:approve"
    scope: organization
    
  hr_manager:
    description: "Human resources management"
    permissions:
      - "employees:read"
      - "employees:write"
      - "payroll:process"
      - "leave_requests:approve"
      - "performance:manage"
    scope: organization
    
  project_manager:
    description: "Project management and tracking"
    permissions:
      - "projects:read"
      - "projects:write"
      - "tasks:assign"
      - "timesheets:approve"
      - "expenses:approve"
    scope: project
    
  employee:
    description: "Standard employee access"
    permissions:
      - "profile:read"
      - "profile:update"
      - "timesheets:submit"
      - "expenses:submit"
      - "leave_requests:submit"
    scope: user
```

### Attribute-Based Access Control (ABAC)

#### Dynamic Permission Evaluation

```typescript
interface PermissionContext {
  user: User;
  resource: Resource;
  action: string;
  environment: Environment;
  
  attributes: {
    user_attributes: {
      department: string;
      job_level: number;
      hire_date: Date;
      manager_id: string;
    };
    
    resource_attributes: {
      owner_id: string;
      department: string;
      confidentiality_level: string;
      created_date: Date;
      amount?: number;
    };
    
    environment_attributes: {
      ip_address: string;
      time_of_day: number;
      day_of_week: number;
      location: string;
    };
  };
}

interface PolicyRule {
  id: string;
  name: string;
  effect: 'permit' | 'deny';
  conditions: PolicyCondition[];
  priority: number;
}

interface PolicyCondition {
  attribute: string;
  operator: 'eq' | 'ne' | 'gt' | 'lt' | 'in' | 'contains' | 'regex';
  value: any;
  type: 'user' | 'resource' | 'environment' | 'action';
}
```

#### Example Policy Rules

```yaml
policy_rules:
  financial_approval_limits:
    name: "Financial approval based on amount and user level"
    effect: permit
    conditions:
      - attribute: "resource.amount"
        operator: "lt"
        value: 1000
        type: "resource"
      - attribute: "user.job_level"
        operator: "gte"
        value: 3
        type: "user"
      - attribute: "action"
        operator: "eq"
        value: "approve"
        type: "action"
        
  confidential_access:
    name: "Restrict access to confidential documents"
    effect: deny
    conditions:
      - attribute: "resource.confidentiality_level"
        operator: "eq"
        value: "confidential"
        type: "resource"
      - attribute: "user.security_clearance"
        operator: "lt"
        value: "high"
        type: "user"
        
  after_hours_restriction:
    name: "Restrict sensitive operations after hours"
    effect: deny
    conditions:
      - attribute: "environment.time_of_day"
        operator: "gt"
        value: 18
        type: "environment"
      - attribute: "action"
        operator: "in"
        value: ["delete", "approve", "transfer"]
        type: "action"
      - attribute: "resource.sensitivity"
        operator: "eq"
        value: "high"
        type: "resource"
```

## 🔑 Authentication & Security

### Multi-Factor Authentication

```typescript
interface MFAConfiguration {
  enabled: boolean;
  required_for_roles: string[];
  methods: MFAMethod[];
  backup_codes_count: number;
  grace_period_days: number;
}

interface MFAMethod {
  type: 'totp' | 'sms' | 'email' | 'hardware_token';
  enabled: boolean;
  configuration: any;
}

// TOTP (Time-based One-Time Password) implementation
class TOTPService {
  generateSecret(): string {
    return crypto.randomBytes(20).toString('hex');
  }
  
  generateQRCode(user: User, secret: string): string {
    const issuer = 'ERP System';
    const label = `${issuer}:${user.email}`;
    const uri = `otpauth://totp/${label}?secret=${secret}&issuer=${issuer}`;
    return qrcode.toDataURL(uri);
  }
  
  verifyToken(secret: string, token: string): boolean {
    const totp = new TOTP(secret);
    return totp.verify(token, { window: 1 }); // Allow 1 step tolerance
  }
  
  generateBackupCodes(count: number = 10): string[] {
    return Array.from({ length: count }, () => 
      crypto.randomBytes(8).toString('hex').toUpperCase()
    );
  }
}
```

### Session Management

```typescript
interface SessionConfiguration {
  max_concurrent_sessions: number;
  session_timeout_minutes: number;
  idle_timeout_minutes: number;
  remember_me_days: number;
  require_reauth_for_sensitive: boolean;
  ip_binding: boolean;
  device_tracking: boolean;
}

class SessionManager {
  async createSession(user: User, deviceInfo: DeviceInfo): Promise<Session> {
    // Check concurrent session limits
    const activeSessions = await this.getActiveSessions(user.id);
    if (activeSessions.length >= this.config.max_concurrent_sessions) {
      await this.terminateOldestSession(user.id);
    }
    
    // Create new session
    const session = {
      id: uuid(),
      user_id: user.id,
      tenant_id: user.tenant_id,
      session_token: this.generateSecureToken(),
      refresh_token: this.generateSecureToken(),
      ip_address: deviceInfo.ip_address,
      user_agent: deviceInfo.user_agent,
      device_info: deviceInfo,
      expires_at: this.calculateExpiry(),
      created_at: new Date(),
      last_activity: new Date(),
      is_active: true
    };
    
    await this.saveSession(session);
    return session;
  }
  
  async validateSession(token: string): Promise<Session | null> {
    const session = await this.getSessionByToken(token);
    
    if (!session || !session.is_active) {
      return null;
    }
    
    // Check expiry
    if (session.expires_at < new Date()) {
      await this.terminateSession(session.id);
      return null;
    }
    
    // Check idle timeout
    const idleTimeout = this.config.idle_timeout_minutes * 60 * 1000;
    if (new Date().getTime() - session.last_activity.getTime() > idleTimeout) {
      await this.terminateSession(session.id);
      return null;
    }
    
    // Update last activity
    await this.updateLastActivity(session.id);
    
    return session;
  }
}
```

### Password Security

```typescript
interface PasswordPolicy {
  min_length: number;
  max_length: number;
  require_uppercase: boolean;
  require_lowercase: boolean;
  require_numbers: boolean;
  require_symbols: boolean;
  min_symbols: number;
  max_age_days: number;
  history_count: number; // Prevent reusing recent passwords
  lockout_attempts: number;
  lockout_duration_minutes: number;
  complexity_score_min: number;
}

class PasswordService {
  async hashPassword(password: string): Promise<{ hash: string; salt: string }> {
    const salt = crypto.randomBytes(32).toString('hex');
    const hash = await bcrypt.hash(password + salt, 12);
    return { hash, salt };
  }
  
  async verifyPassword(password: string, hash: string, salt: string): Promise<boolean> {
    return await bcrypt.compare(password + salt, hash);
  }
  
  validatePassword(password: string, policy: PasswordPolicy, user?: User): ValidationResult {
    const errors: string[] = [];
    
    if (password.length < policy.min_length) {
      errors.push(`Password must be at least ${policy.min_length} characters`);
    }
    
    if (password.length > policy.max_length) {
      errors.push(`Password must not exceed ${policy.max_length} characters`);
    }
    
    if (policy.require_uppercase && !/[A-Z]/.test(password)) {
      errors.push('Password must contain at least one uppercase letter');
    }
    
    if (policy.require_lowercase && !/[a-z]/.test(password)) {
      errors.push('Password must contain at least one lowercase letter');
    }
    
    if (policy.require_numbers && !/\d/.test(password)) {
      errors.push('Password must contain at least one number');
    }
    
    if (policy.require_symbols && !/[!@#$%^&*(),.?":{}|<>]/.test(password)) {
      errors.push('Password must contain at least one symbol');
    }
    
    // Check complexity score
    const complexity = this.calculateComplexity(password);
    if (complexity < policy.complexity_score_min) {
      errors.push('Password does not meet complexity requirements');
    }
    
    return {
      valid: errors.length === 0,
      errors,
      complexity_score: complexity
    };
  }
  
  private calculateComplexity(password: string): number {
    let score = 0;
    
    // Length bonus
    score += Math.min(password.length * 2, 20);
    
    // Character variety
    if (/[a-z]/.test(password)) score += 5;
    if (/[A-Z]/.test(password)) score += 5;
    if (/\d/.test(password)) score += 5;
    if (/[!@#$%^&*(),.?":{}|<>]/.test(password)) score += 10;
    
    // Pattern penalties
    if (/(.)\1{2,}/.test(password)) score -= 10; // Repeated characters
    if (/123|abc|qwe/i.test(password)) score -= 15; // Common sequences
    
    return Math.max(0, Math.min(100, score));
  }
}
```

## 👥 User Lifecycle Management

### User Onboarding Workflow

```yaml
user_onboarding:
  invitation_process:
    1. admin_sends_invitation:
        - validate_email_format
        - check_existing_users
        - generate_invitation_token
        - send_invitation_email
        - set_expiration_time
    
    2. user_accepts_invitation:
        - validate_invitation_token
        - collect_user_information
        - set_initial_password
        - verify_email_address
        - create_user_account
    
    3. initial_setup:
        - assign_default_roles
        - set_organization_association
        - configure_user_preferences
        - setup_mfa_if_required
        - send_welcome_email
    
    4. first_login:
        - force_password_change
        - complete_profile_setup
        - accept_terms_of_service
        - configure_notification_preferences
        - start_guided_tour

  automatic_provisioning:
    triggers:
      - employee_hired
      - contractor_onboarded
      - external_partner_added
    
    data_sources:
      - hr_system_integration
      - active_directory_sync
      - manual_admin_entry
```

### User Deactivation & Offboarding

```typescript
interface OffboardingProcess {
  trigger_events: string[];
  steps: OffboardingStep[];
  data_retention_policy: DataRetentionPolicy;
  access_revocation: AccessRevocationPolicy;
}

interface OffboardingStep {
  name: string;
  type: 'automatic' | 'manual';
  assignee?: string;
  sla_hours: number;
  dependencies: string[];
  actions: OffboardingAction[];
}

interface OffboardingAction {
  action_type: 'revoke_access' | 'transfer_ownership' | 'backup_data' | 'notify_stakeholders';
  parameters: any;
  rollback_possible: boolean;
}

// Example offboarding workflow
const employeeOffboarding: OffboardingProcess = {
  trigger_events: ['employment_terminated', 'resignation_submitted', 'contract_expired'],
  
  steps: [
    {
      name: 'immediate_access_revocation',
      type: 'automatic',
      sla_hours: 0,
      dependencies: [],
      actions: [
        {
          action_type: 'revoke_access',
          parameters: { disable_login: true, revoke_tokens: true },
          rollback_possible: true
        }
      ]
    },
    
    {
      name: 'data_ownership_transfer',
      type: 'manual',
      assignee: 'manager',
      sla_hours: 24,
      dependencies: ['immediate_access_revocation'],
      actions: [
        {
          action_type: 'transfer_ownership',
          parameters: { transfer_to: 'manager', include_documents: true },
          rollback_possible: true
        }
      ]
    },
    
    {
      name: 'system_cleanup',
      type: 'automatic',
      sla_hours: 72,
      dependencies: ['data_ownership_transfer'],
      actions: [
        {
          action_type: 'backup_data',
          parameters: { retention_years: 7 },
          rollback_possible: false
        }
      ]
    }
  ],
  
  data_retention_policy: {
    personal_data: '30_days',
    work_documents: '7_years',
    audit_logs: 'permanent'
  },
  
  access_revocation: {
    immediate_revocation: ['system_access', 'email_access', 'vpn_access'],
    delayed_revocation: ['building_access', 'company_devices'],
    manager_review_required: ['project_ownership', 'client_relationships']
  }
};
```

## 📊 User Analytics & Monitoring

### User Activity Tracking

```sql
-- User activity logging
CREATE TABLE user_activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    session_id UUID REFERENCES user_sessions(id),
    
    -- Activity details
    activity_type VARCHAR(50) NOT NULL, -- login, logout, create, update, delete, view
    module VARCHAR(50), -- financial, inventory, hr, etc.
    resource_type VARCHAR(50), -- invoice, purchase_order, employee
    resource_id UUID,
    
    -- Request details
    ip_address INET,
    user_agent TEXT,
    request_method VARCHAR(10),
    request_path TEXT,
    request_params JSONB,
    response_status INTEGER,
    response_time_ms INTEGER,
    
    -- Metadata
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    additional_data JSONB DEFAULT '{}'
);

-- Create partitioned table for performance
CREATE TABLE user_activities_y2025m01 PARTITION OF user_activities
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
```

### User Behavior Analytics

```typescript
interface UserAnalytics {
  user_id: string;
  tenant_id: string;
  period: 'daily' | 'weekly' | 'monthly';
  
  activity_metrics: {
    total_sessions: number;
    total_duration_minutes: number;
    page_views: number;
    actions_performed: number;
    features_used: string[];
  };
  
  productivity_metrics: {
    documents_created: number;
    transactions_processed: number;
    approvals_completed: number;
    reports_generated: number;
  };
  
  engagement_metrics: {
    days_active: number;
    average_session_duration: number;
    bounce_rate: number;
    feature_adoption_rate: number;
  };
  
  security_metrics: {
    failed_login_attempts: number;
    suspicious_activities: number;
    policy_violations: number;
    mfa_usage_rate: number;
  };
}
```

## 🔧 Permission Management API

### User Management Endpoints

```typescript
interface UserManagementAPI {
  // User CRUD operations
  GET    /api/v1/users                           // List users with filters
  POST   /api/v1/users                           // Create new user
  GET    /api/v1/users/{user_id}                 // Get user details
  PUT    /api/v1/users/{user_id}                 // Update user
  DELETE /api/v1/users/{user_id}                 // Deactivate user
  
  // User invitation and onboarding
  POST   /api/v1/users/invite                    // Send invitation
  POST   /api/v1/users/accept-invitation         // Accept invitation
  POST   /api/v1/users/{user_id}/resend-invite   // Resend invitation
  
  // Role and permission management
  GET    /api/v1/users/{user_id}/roles           // Get user roles
  POST   /api/v1/users/{user_id}/roles           // Assign role
  DELETE /api/v1/users/{user_id}/roles/{role_id} // Remove role
  GET    /api/v1/users/{user_id}/permissions     // Get effective permissions
  
  // Authentication and security
  POST   /api/v1/auth/login                      // User login
  POST   /api/v1/auth/logout                     // User logout
  POST   /api/v1/auth/refresh                    // Refresh token
  POST   /api/v1/auth/forgot-password           // Password reset request
  POST   /api/v1/auth/reset-password            // Reset password
  POST   /api/v1/auth/change-password           // Change password
  
  // Multi-factor authentication
  POST   /api/v1/auth/mfa/setup                 // Setup MFA
  POST   /api/v1/auth/mfa/verify                // Verify MFA token
  POST   /api/v1/auth/mfa/disable               // Disable MFA
  GET    /api/v1/auth/mfa/backup-codes          // Get backup codes
  
  // Session management
  GET    /api/v1/users/{user_id}/sessions       // List active sessions
  DELETE /api/v1/users/{user_id}/sessions/{session_id} // Terminate session
  DELETE /api/v1/users/{user_id}/sessions       // Terminate all sessions
  
  // User preferences and settings
  GET    /api/v1/users/{user_id}/preferences    // Get preferences
  PUT    /api/v1/users/{user_id}/preferences    // Update preferences
  
  // User activity and analytics
  GET    /api/v1/users/{user_id}/activities     // Get activity log
  GET    /api/v1/users/{user_id}/analytics      // Get user analytics
}
```

This comprehensive user management and permissions system provides secure, scalable, and flexible access control suitable for enterprise-level ERP systems while maintaining usability for administrators and end users.
