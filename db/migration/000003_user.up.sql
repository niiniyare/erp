-- ================================================================================================
-- PostgreSQL  RBAC/ABAC Schema with UUID Primary Keys and Row Level Security (RLS)
-- ================================================================================================
-- 
-- This schema implements a comprehensive Role-Based Access Control (RBAC) and 
-- Attribute-Based Access Control (ABAC) system with the following features:
-- 
-- * UUID primary keys throughout for better distributed system support
-- * Row Level Security (RLS) for tenant data isolation
-- * Person-Employee-User separation pattern for flexible identity management
-- * Module-based permission organization for scalable authorization
-- * ABAC policies with advanced rule engine and compliance tracking
-- * Comprehensive audit logging and access request workflows
-- * Performance-optimized indexes and views
-- 
-- Prerequisites: 
-- - `tenants` table with INTEGER primary key
-- - `entities` table with UUID primary key
-- - Both tables must exist before running this schema
-- ================================================================================================

-- ================================================================================================
-- PERSON-EMPLOYEE-USER IDENTITY PATTERN
-- ================================================================================================

-- ------------------------------------------------------------------------------------------------
-- PERSONS TABLE
-- ------------------------------------------------------------------------------------------------
-- Stores generic person entities that can represent individuals, employees, contacts, customers, etc.
-- Supports ABAC with security attributes and flexible person typing.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS persons (
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
    security_attributes JSONB DEFAULT '{}'::jsonb, -- ABAC attributes: clearance level, department, location, etc.
    metadata JSONB DEFAULT '{}'::jsonb,            -- Additional flexible data storage
    is_active BOOLEAN NOT NULL DEFAULT true,
    
    -- Standard validation columns
    version INTEGER NOT NULL DEFAULT 1,
    last_validation_run TIMESTAMPTZ,
    validation_status VARCHAR(20) DEFAULT 'PENDING' CHECK (
        validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR')
    ),
    validation_errors JSONB DEFAULT '[]'::jsonb,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,                        -- Soft delete timestamp
    
    -- Ensure email uniqueness per tenant (excluding soft-deleted records)
    CONSTRAINT persons_email_unique_active 
        EXCLUDE (tenant_id WITH =, email WITH =) 
        WHERE (email IS NOT NULL AND deleted_at IS NULL),
    
    -- Ensure national ID uniqueness per tenant (excluding soft-deleted records)
    CONSTRAINT persons_national_id_unique_active 
        EXCLUDE (tenant_id WITH =, national_id WITH =) 
        WHERE (national_id IS NOT NULL AND deleted_at IS NULL)
);

-- Add table and column comments
COMMENT ON TABLE persons IS 
'Stores person entities with ABAC security attributes. Supports multiple person types including employees, customers, vendors, and contractors. Implements soft delete and tenant isolation.';

COMMENT ON COLUMN persons.id IS 'UUID primary key for the person record';
COMMENT ON COLUMN persons.tenant_id IS 'Foreign key to tenants table for multi-tenant isolation';
COMMENT ON COLUMN persons.entity_id IS 'Foreign key to entities table for hierarchical organization';
COMMENT ON COLUMN persons.person_type IS 'Classification of person: INDIVIDUAL, EMPLOYEE, CONTACT, CUSTOMER, VENDOR, CONTRACTOR';
COMMENT ON COLUMN persons.security_attributes IS 'JSONB containing ABAC attributes like clearance level, department, location for access control';
COMMENT ON COLUMN persons.metadata IS 'Flexible JSONB storage for additional person-related data';
COMMENT ON COLUMN persons.deleted_at IS 'Soft delete timestamp - NULL means record is active';

-- Enable RLS and create policies
ALTER TABLE persons ENABLE ROW LEVEL SECURITY;

-- Enhanced RLS policies following recommendations
CREATE POLICY persons_tenant_isolation ON persons
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    );

CREATE POLICY persons_admin_access ON persons
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);

-- ------------------------------------------------------------------------------------------------
-- EMPLOYEES TABLE
-- ------------------------------------------------------------------------------------------------
-- Extends persons with employment-specific data and organizational hierarchy.
-- Contains role context, security levels, and employment status tracking.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS employees (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    person_id UUID NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    employee_number VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE RESTRICT,
    position_title VARCHAR(100),
    department_id UUID REFERENCES entities(uuid), -- Department entity reference
    manager_id UUID REFERENCES employees(id),     -- Self-referential for org hierarchy
    hire_date DATE NOT NULL,
    termination_date DATE,
    salary_info JSONB DEFAULT '{}'::jsonb,        -- Encrypted/sensitive salary data
    employment_status VARCHAR(20) DEFAULT 'ACTIVE'
        CHECK (employment_status IN ('ACTIVE', 'INACTIVE', 'TERMINATED', 'ON_LEAVE', 'SUSPENDED')),
    work_schedule JSONB DEFAULT '{}'::jsonb,      -- Flexible work schedule definition
    security_level INTEGER DEFAULT 0,             -- Numeric security clearance level (0=lowest)
    access_attributes JSONB DEFAULT '{}'::jsonb,  -- Employment-specific ABAC attributes
    
    -- Standard validation columns
    version INTEGER NOT NULL DEFAULT 1,
    last_validation_run TIMESTAMPTZ,
    validation_status VARCHAR(20) DEFAULT 'PENDING' CHECK (
        validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR')
    ),
    validation_errors JSONB DEFAULT '[]'::jsonb,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    -- Ensure employee number uniqueness per tenant
    CONSTRAINT employees_number_unique_active 
        EXCLUDE (tenant_id WITH =, employee_number WITH =) 
        WHERE (deleted_at IS NULL)
);

COMMENT ON TABLE employees IS 
'Employee records extending persons with employment-specific data, organizational hierarchy, and security levels for access control.';

COMMENT ON COLUMN employees.employee_number IS 'Unique employee identifier within tenant';
COMMENT ON COLUMN employees.department_id IS 'Foreign key to entities table representing department';
COMMENT ON COLUMN employees.manager_id IS 'Self-referential foreign key for organizational hierarchy';
COMMENT ON COLUMN employees.security_level IS 'Numeric security clearance level (0=lowest, higher numbers = higher clearance)';
COMMENT ON COLUMN employees.access_attributes IS 'JSONB containing employment-specific ABAC attributes for access control';
COMMENT ON COLUMN employees.salary_info IS 'JSONB containing encrypted/sensitive salary and compensation data';

-- Enable RLS and create policies
ALTER TABLE employees ENABLE ROW LEVEL SECURITY;

CREATE POLICY employees_tenant_isolation ON employees
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    );

CREATE POLICY employees_admin_access ON employees
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);

-- ------------------------------------------------------------------------------------------------
-- USERS TABLE
-- ------------------------------------------------------------------------------------------------
-- System access accounts with authentication data and RBAC integration.
-- Can be linked to persons/employees or exist independently for service accounts.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE RESTRICT,
    person_id UUID REFERENCES persons(id) ON DELETE SET NULL,
    employee_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    username VARCHAR(100),
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255),
    user_type VARCHAR(20) NOT NULL DEFAULT 'INTERNAL'
        CHECK (user_type IN ('INTERNAL', 'CUSTOMER', 'VENDOR', 'PARTNER', 'API', 'SERVICE', 'ADMIN')),
    account_status VARCHAR(20) DEFAULT 'ACTIVE'
        CHECK (account_status IN ('ACTIVE', 'INACTIVE', 'LOCKED', 'SUSPENDED', 'PENDING_VERIFICATION')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_login_at TIMESTAMPTZ,
    password_changed_at TIMESTAMPTZ DEFAULT NOW(),
    failed_login_attempts INTEGER DEFAULT 0,
    lockout_until TIMESTAMPTZ,
    session_timeout_minutes INTEGER DEFAULT 480,   -- 8 hours default
    mfa_enabled BOOLEAN DEFAULT false,
    mfa_secret VARCHAR(255),
    user_attributes JSONB DEFAULT '{}'::jsonb,     -- ABAC user attributes
    settings JSONB DEFAULT '{}'::jsonb,            -- User preferences and settings
    
    -- Standard validation columns
    version INTEGER NOT NULL DEFAULT 1,
    last_validation_run TIMESTAMPTZ,
    validation_status VARCHAR(20) DEFAULT 'PENDING' CHECK (
        validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR')
    ),
    validation_errors JSONB DEFAULT '[]'::jsonb,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    -- Ensure email uniqueness per tenant
    CONSTRAINT users_email_unique_active 
        EXCLUDE (tenant_id WITH =, email WITH =) 
        WHERE (deleted_at IS NULL),
    
    -- Ensure username uniqueness per tenant when provided
    CONSTRAINT users_username_unique_active 
        EXCLUDE (tenant_id WITH =, username WITH =) 
        WHERE (username IS NOT NULL AND deleted_at IS NULL)
);

COMMENT ON TABLE users IS 
'System user accounts with authentication, authorization, and session management. Can be linked to persons/employees or exist independently for service accounts.';

COMMENT ON COLUMN users.user_type IS 'Classification of user account: INTERNAL, CUSTOMER, VENDOR, PARTNER, API, SERVICE, ADMIN';
COMMENT ON COLUMN users.account_status IS 'Current account status affecting login ability';
COMMENT ON COLUMN users.session_timeout_minutes IS 'Session timeout in minutes (default 480 = 8 hours)';
COMMENT ON COLUMN users.failed_login_attempts IS 'Counter for failed login attempts for security monitoring';
COMMENT ON COLUMN users.user_attributes IS 'JSONB containing ABAC attributes for fine-grained access control';
COMMENT ON COLUMN users.settings IS 'JSONB containing user preferences and application settings';

-- Enable RLS and create policies
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

CREATE POLICY users_tenant_isolation ON users
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    );

CREATE POLICY users_admin_access ON users
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);

-- ------------------------------------------------------------------------------------------------
-- USER SESSIONS TABLE
-- ------------------------------------------------------------------------------------------------
-- Tracks active user sessions with security context for ABAC evaluation.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    refresh_token VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    device_info JSONB DEFAULT '{}'::jsonb,         -- Device fingerprinting data
    location_info JSONB DEFAULT '{}'::jsonb,       -- Geographic/network location for ABAC
    expires_at TIMESTAMPTZ NOT NULL,
    
    -- Standard validation columns
    version INTEGER NOT NULL DEFAULT 1,
    last_validation_run TIMESTAMPTZ,
    validation_status VARCHAR(20) DEFAULT 'PENDING' CHECK (
        validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR')
    ),
    validation_errors JSONB DEFAULT '[]'::jsonb,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_accessed_at TIMESTAMPTZ DEFAULT NOW(),
    is_active BOOLEAN DEFAULT true
);

COMMENT ON TABLE user_sessions IS 
'Active user sessions with security context including device, location, and access patterns for ABAC evaluation and security monitoring.';

COMMENT ON COLUMN user_sessions.device_info IS 'JSONB containing device fingerprinting data for security analysis';
COMMENT ON COLUMN user_sessions.location_info IS 'JSONB containing geographic and network location data for location-based access control';

-- Enable RLS and create policies
ALTER TABLE user_sessions ENABLE ROW LEVEL SECURITY;

CREATE POLICY user_sessions_tenant_isolation ON user_sessions
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    );

CREATE POLICY user_sessions_admin_access ON user_sessions
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);

-- ================================================================================================
-- ENHANCED RBAC SYSTEM
-- ================================================================================================

-- ------------------------------------------------------------------------------------------------
-- MODULES TABLE
-- ------------------------------------------------------------------------------------------------
-- Organizes system functionality into modules for permission management and feature control.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS modules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    display_name VARCHAR(100),
    description TEXT,
    category VARCHAR(50),                          -- 'CORE', 'HR', 'FINANCE', 'SALES', etc.
    version VARCHAR(20),
    is_active BOOLEAN DEFAULT true,
    
    -- Standard validation columns
    validation_version INTEGER NOT NULL DEFAULT 1,
    last_validation_run TIMESTAMPTZ,
    validation_status VARCHAR(20) DEFAULT 'PENDING' CHECK (
        validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR')
    ),
    validation_errors JSONB DEFAULT '[]'::jsonb,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT modules_name_unique_per_tenant UNIQUE (tenant_id, name)
);

COMMENT ON TABLE modules IS 
'System modules for organizing permissions and features into logical groups. Enables modular permission management and feature toggles.';

COMMENT ON COLUMN modules.category IS 'Module category for grouping: CORE, HR, FINANCE, SALES, INVENTORY, etc.';
COMMENT ON COLUMN modules.version IS 'Module version for tracking feature updates and compatibility';

-- Enable RLS and create policies
ALTER TABLE modules ENABLE ROW LEVEL SECURITY;

CREATE POLICY modules_tenant_isolation ON modules
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    );

CREATE POLICY modules_admin_access ON modules
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);

-- ------------------------------------------------------------------------------------------------
-- RESOURCES TABLE
-- ------------------------------------------------------------------------------------------------
-- Defines system resources that can be protected by permissions (APIs, UI components, data, etc.).
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS resources (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    module_id UUID NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    entity_id UUID REFERENCES entities(uuid),      -- Resource can belong to specific entity
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(150),
    description TEXT,
    resource_type VARCHAR(50) NOT NULL 
        CHECK (resource_type IN ('API', 'UI', 'DATA', 'FILE', 'REPORT', 'WORKFLOW', 'FUNCTION')),
    parent_resource_id UUID REFERENCES resources(id),
    path VARCHAR(500),                             -- URL path, API endpoint, file path, etc.
    resource_attributes JSONB DEFAULT '{}'::jsonb, -- ABAC resource attributes
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    CONSTRAINT resources_name_unique_per_module UNIQUE (tenant_id, module_id, name)
);

COMMENT ON TABLE resources IS 
'System resources that can be protected by permissions including APIs, UI components, data objects, files, reports, and workflows.';

COMMENT ON COLUMN resources.resource_type IS 'Type of resource: API, UI, DATA, FILE, REPORT, WORKFLOW, FUNCTION';
COMMENT ON COLUMN resources.parent_resource_id IS 'Self-referential for resource hierarchy (e.g., API endpoints under API group)';
COMMENT ON COLUMN resources.path IS 'Resource path: URL, API endpoint, file path, database object, etc.';
COMMENT ON COLUMN resources.resource_attributes IS 'JSONB containing ABAC attributes like classification level, sensitivity, department ownership';

-- Enable RLS and create policies
ALTER TABLE resources ENABLE ROW LEVEL SECURITY;

CREATE POLICY resources_tenant_isolation ON resources
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    );

CREATE POLICY resources_admin_access ON resources
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);

-- ------------------------------------------------------------------------------------------------
-- ACTIONS TABLE
-- ------------------------------------------------------------------------------------------------
-- Defines actions that can be performed on resources with risk and approval requirements.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS actions (
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
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT actions_name_unique_per_tenant UNIQUE (tenant_id, name)
);

COMMENT ON TABLE actions IS 
'Defines actions that can be performed on resources with risk assessment and approval workflow requirements.';

COMMENT ON COLUMN actions.action_type IS 'Standard action type: CREATE, READ, UPDATE, DELETE, EXECUTE, APPROVE, REJECT, EXPORT, IMPORT';
COMMENT ON COLUMN actions.action_category IS 'Action category for risk assessment: STANDARD, ADMINISTRATIVE, SENSITIVE, BULK, SYSTEM';
COMMENT ON COLUMN actions.risk_level IS 'Risk level for audit and approval workflows: LOW, MEDIUM, HIGH, CRITICAL';
COMMENT ON COLUMN actions.requires_approval IS 'Whether this action requires explicit approval before execution';

-- Enable RLS and create policies
ALTER TABLE actions ENABLE ROW LEVEL SECURITY;

CREATE POLICY actions_tenant_isolation ON actions
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    );

CREATE POLICY actions_admin_access ON actions
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);

-- ------------------------------------------------------------------------------------------------
-- ROLES TABLE
-- ------------------------------------------------------------------------------------------------
-- Defines roles with module scope, entity context, and hierarchical structure.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE RESTRICT,
    name VARCHAR(50) NOT NULL,
    display_name VARCHAR(100),
    description TEXT,
    module_id UUID REFERENCES modules(id),         -- Module this role belongs to
    role_type VARCHAR(20) DEFAULT 'CUSTOM'
        CHECK (role_type IN ('SYSTEM', 'TENANT', 'ENTITY', 'CUSTOM', 'FUNCTIONAL')),
    parent_role_id UUID REFERENCES roles(id),      -- Role hierarchy
    level INTEGER DEFAULT 0,                       -- Hierarchy level (calculated)
    permissions JSONB NOT NULL DEFAULT '{}'::jsonb, -- Direct permissions cache for performance
    entity_scope JSONB DEFAULT '{}'::jsonb,        -- Entity-level access rules
    conditions JSONB DEFAULT '{}'::jsonb,          -- Time, location, device conditions
    is_system_role BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    
    -- Standard validation columns
    version INTEGER NOT NULL DEFAULT 1,
    last_validation_run TIMESTAMPTZ,
    validation_status VARCHAR(20) DEFAULT 'PENDING' CHECK (
        validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR')
    ),
    validation_errors JSONB DEFAULT '[]'::jsonb,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    CONSTRAINT roles_name_unique_per_tenant UNIQUE (tenant_id, name)
);

COMMENT ON TABLE roles IS 
'Roles with module association, entity scoping, and hierarchical structure. Supports both RBAC and ABAC with conditional access rules.';

COMMENT ON COLUMN roles.role_type IS 'Role classification: SYSTEM (built-in), TENANT (tenant-wide), ENTITY (entity-scoped), CUSTOM (user-defined), FUNCTIONAL (job-based)';
COMMENT ON COLUMN roles.parent_role_id IS 'Parent role for inheritance hierarchy';
COMMENT ON COLUMN roles.level IS 'Calculated hierarchy level (0=root, higher=deeper)';
COMMENT ON COLUMN roles.permissions IS 'Cached permissions JSONB for performance optimization';
COMMENT ON COLUMN roles.entity_scope IS 'JSONB defining which entities this role can access';
COMMENT ON COLUMN roles.conditions IS 'JSONB containing time, location, device, and other conditional access rules';

-- Enable RLS and create policies
ALTER TABLE roles ENABLE ROW LEVEL SECURITY;

CREATE POLICY roles_tenant_isolation ON roles
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    );

CREATE POLICY roles_admin_access ON roles
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);

-- ------------------------------------------------------------------------------------------------
-- PERMISSIONS TABLE
-- ------------------------------------------------------------------------------------------------
-- Granular permissions combining resources and actions with ABAC conditions.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    resource_id UUID NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
    action_id UUID NOT NULL REFERENCES actions(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    display_name VARCHAR(250),
    description TEXT,
    effect VARCHAR(20) DEFAULT 'ALLOW' CHECK (effect IN ('ALLOW', 'DENY')),
    conditions JSONB DEFAULT '{}'::jsonb,          -- ABAC evaluation conditions
    data_filters JSONB DEFAULT '{}'::jsonb,       -- Row-level security filters
    field_restrictions JSONB DEFAULT '{}'::jsonb, -- Column-level restrictions
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT permissions_unique_per_resource_action UNIQUE (tenant_id, resource_id, action_id, name)
);

COMMENT ON TABLE permissions IS 
'Granular permissions combining resources and actions with ABAC conditions, data filters, and field restrictions for fine-grained access control.';

COMMENT ON COLUMN permissions.effect IS 'Permission effect: ALLOW (grant access) or DENY (explicitly deny access)';
COMMENT ON COLUMN permissions.conditions IS 'JSONB containing ABAC evaluation conditions (time, location, attributes, etc.)';
COMMENT ON COLUMN permissions.data_filters IS 'JSONB containing row-level security filters to limit data access';
COMMENT ON COLUMN permissions.field_restrictions IS 'JSONB containing column-level restrictions to limit field access';

-- Enable RLS and create policies
ALTER TABLE permissions ENABLE ROW LEVEL SECURITY;

CREATE POLICY permissions_tenant_isolation ON permissions
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    );

CREATE POLICY permissions_admin_access ON permissions
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);

-- ------------------------------------------------------------------------------------------------
-- ROLE PERMISSIONS MAPPING
-- ------------------------------------------------------------------------------------------------
-- Maps permissions to roles with entity-specific scoping and conditions.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS role_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    entity_scope UUID REFERENCES entities(uuid),   -- Entity-specific permission scope
    granted_by UUID REFERENCES users(id),
    granted_at TIMESTAMPTZ DEFAULT NOW(),
    conditions JSONB DEFAULT '{}'::jsonb,          -- Additional conditions beyond permission
    is_active BOOLEAN DEFAULT true,
    
    CONSTRAINT role_permissions_unique_assignment UNIQUE (tenant_id, role_id, permission_id, entity_scope)
);

COMMENT ON TABLE role_permissions IS 
'Maps permissions to roles with optional entity-specific scoping and additional conditions for flexible authorization.';

COMMENT ON COLUMN role_permissions.entity_scope IS 'Optional entity restriction - if specified, permission only applies within this entity';
COMMENT ON COLUMN role_permissions.granted_by IS 'User who granted this permission assignment';
COMMENT ON COLUMN role_permissions.conditions IS 'Additional JSONB conditions beyond those in the permission itself';

-- Enable RLS and create policies
ALTER TABLE role_permissions ENABLE ROW LEVEL SECURITY;

CREATE POLICY role_permissions_tenant_isolation ON role_permissions
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    );

CREATE POLICY role_permissions_admin_access ON role_permissions
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);

-- ------------------------------------------------------------------------------------------------
-- USER ROLE ASSIGNMENTS
-- ------------------------------------------------------------------------------------------------
-- Assigns roles to users with entity context, delegation, and temporal controls.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE RESTRICT,
    assignment_type VARCHAR(20) DEFAULT 'DIRECT'
        CHECK (assignment_type IN ('DIRECT', 'INHERITED', 'DELEGATED', 'TEMPORARY')),
    delegated_by UUID REFERENCES users(id),        -- If delegated assignment
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    assigned_by UUID REFERENCES users(id),
    expires_at TIMESTAMPTZ,                        -- For temporary assignments
    conditions JSONB DEFAULT '{}'::jsonb,          -- Time/location/device conditions
    is_active BOOLEAN DEFAULT true,
    
    CONSTRAINT user_roles_unique_assignment UNIQUE (user_id, role_id, entity_id)
);

COMMENT ON TABLE user_roles IS 
'Assigns roles to users with entity context, delegation support, and temporal controls for dynamic authorization.';

COMMENT ON COLUMN user_roles.assignment_type IS 'Type of assignment: DIRECT (explicitly assigned), INHERITED (from hierarchy), DELEGATED (from another user), TEMPORARY (time-limited)';
COMMENT ON COLUMN user_roles.delegated_by IS 'User who delegated this role assignment (for DELEGATED type)';
COMMENT ON COLUMN user_roles.expires_at IS 'Expiration timestamp for temporary role assignments';
COMMENT ON COLUMN user_roles.conditions IS 'JSONB containing conditional access rules (time, location, device, etc.)';

-- Enable RLS and create policies
ALTER TABLE user_roles ENABLE ROW LEVEL SECURITY;

CREATE POLICY user_roles_tenant_isolation ON user_roles
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL 
        AND EXISTS (
            SELECT 1 FROM users u 
            WHERE u.id = user_roles.user_id 
            AND u.tenant_id = current_tenant_id()
        )
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL 
        AND EXISTS (
            SELECT 1 FROM users u 
            WHERE u.id = user_roles.user_id 
            AND u.tenant_id = current_tenant_id()
        )
    );

CREATE POLICY user_roles_admin_access ON user_roles
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);

-- ------------------------------------------------------------------------------------------------
-- DIRECT USER PERMISSIONS
-- ------------------------------------------------------------------------------------------------
-- Direct permission grants to users bypassing roles for exceptional access.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS user_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    entity_id UUID REFERENCES entities(uuid),
    effect VARCHAR(20) DEFAULT 'ALLOW' CHECK (effect IN ('ALLOW', 'DENY')),
    reason TEXT,                                   -- Justification for direct permission
    granted_by UUID REFERENCES users(id),
    granted_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ,                       -- For temporary permissions
    conditions JSONB DEFAULT '{}'::jsonb,         -- Additional conditions
    is_active BOOLEAN DEFAULT true,
    
    CONSTRAINT user_permissions_unique_assignment UNIQUE (tenant_id, user_id, permission_id, entity_id)
);

COMMENT ON TABLE user_permissions IS 
'Direct permission grants to users bypassing roles. Used for exceptional access, denials, and temporary permissions.';

COMMENT ON COLUMN user_permissions.effect IS 'Permission effect: ALLOW (grant access) or DENY (explicitly deny - overrides role permissions)';
COMMENT ON COLUMN user_permissions.reason IS 'Business justification for this direct permission assignment';
COMMENT ON COLUMN user_permissions.granted_by IS 'User who granted this direct permission';

-- Enable RLS and create policies
ALTER TABLE user_permissions ENABLE ROW LEVEL SECURITY;

CREATE POLICY user_permissions_tenant_isolation ON user_permissions
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    );

CREATE POLICY user_permissions_admin_access ON user_permissions
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);

-- ================================================================================================
-- ABAC SYSTEM
-- ================================================================================================

-- ------------------------------------------------------------------------------------------------
-- ATTRIBUTE DEFINITIONS
-- ------------------------------------------------------------------------------------------------
-- Defines attributes used in ABAC policies with validation and encryption controls.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS attribute_definitions (
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
    is_sensitive BOOLEAN DEFAULT false,            -- For PII/sensitive attributes
    default_value TEXT,
    allowed_values JSONB,                          -- For enum types
    validation_rules JSONB DEFAULT '{}'::jsonb,    -- Custom validation rules
    encryption_required BOOLEAN DEFAULT false,     -- Whether values must be encrypted
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT attribute_definitions_name_unique_per_tenant UNIQUE (tenant_id, name)
);

COMMENT ON TABLE attribute_definitions IS 
'Defines attributes used in ABAC policies with data types, validation rules, and security controls for consistent attribute management.';

COMMENT ON COLUMN attribute_definitions.data_type IS 'Attribute data type: STRING, NUMBER, BOOLEAN, DATE, TIME, JSON, ARRAY, ENUM';
COMMENT ON COLUMN attribute_definitions.category IS 'Attribute category: USER (user attributes), RESOURCE (resource attributes), ENVIRONMENT (context), ACTION (action attributes), ENTITY (entity attributes), SESSION (session context)';
COMMENT ON COLUMN attribute_definitions.is_sensitive IS 'Whether attribute contains PII or sensitive data requiring special handling';
COMMENT ON COLUMN attribute_definitions.allowed_values IS 'JSONB array of allowed values for ENUM data type';
COMMENT ON COLUMN attribute_definitions.validation_rules IS 'JSONB containing custom validation rules (regex, ranges, etc.)';
COMMENT ON COLUMN attribute_definitions.encryption_required IS 'Whether attribute values must be encrypted at rest';

-- Enable RLS and create policies
ALTER TABLE attribute_definitions ENABLE ROW LEVEL SECURITY;

CREATE POLICY attribute_definitions_tenant_isolation ON attribute_definitions
    FOR ALL TO public
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

-- ------------------------------------------------------------------------------------------------
-- ABAC POLICIES
-- ------------------------------------------------------------------------------------------------
-- ABAC policies with advanced rule engine, priorities, and compliance tracking.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS policies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID REFERENCES entities(uuid),      -- Entity scope for policy
    name VARCHAR(150) NOT NULL,
    display_name VARCHAR(200),
    description TEXT,
    policy_type VARCHAR(50) DEFAULT 'ABAC' 
        CHECK (policy_type IN ('ABAC', 'RBAC', 'HYBRID', 'TIME_BASED', 'LOCATION_BASED')),
    effect VARCHAR(20) DEFAULT 'ALLOW' CHECK (effect IN ('ALLOW', 'DENY')),
    priority INTEGER DEFAULT 100,                  -- Higher numbers = higher priority
    category VARCHAR(50) DEFAULT 'ACCESS'
        CHECK (category IN ('ACCESS', 'DATA_FILTER', 'FIELD_MASK', 'AUDIT', 'COMPLIANCE')),
    target JSONB NOT NULL,                         -- When policy applies (conditions)
    rule JSONB NOT NULL,                           -- Policy logic/evaluation rules
    obligations JSONB DEFAULT '{}'::jsonb,        -- Required actions when policy fires
    advice JSONB DEFAULT '{}'::jsonb,              -- Optional actions/logging recommendations
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    deleted_at TIMESTAMPTZ,
    
    CONSTRAINT policies_name_unique_per_tenant UNIQUE (tenant_id, name)
);

COMMENT ON TABLE policies IS 
'ABAC policies with advanced rule engine supporting multiple policy types, priorities, obligations, and compliance tracking.';

COMMENT ON COLUMN policies.policy_type IS 'Policy type: ABAC (attribute-based), RBAC (role-based), HYBRID (combined), TIME_BASED (temporal), LOCATION_BASED (geographic)';
COMMENT ON COLUMN policies.priority IS 'Policy priority for conflict resolution (higher numbers processed first)';
COMMENT ON COLUMN policies.category IS 'Policy category: ACCESS (authorization), DATA_FILTER (row-level), FIELD_MASK (column-level), AUDIT (logging), COMPLIANCE (regulatory)';
COMMENT ON COLUMN policies.target IS 'JSONB defining when policy applies (subjects, resources, actions, conditions)';
COMMENT ON COLUMN policies.rule IS 'JSONB containing policy evaluation logic and conditions';
COMMENT ON COLUMN policies.obligations IS 'JSONB defining required actions when policy fires (logging, notifications, etc.)';
COMMENT ON COLUMN policies.advice IS 'JSONB defining optional actions and recommendations';

-- Enable RLS and create policies
ALTER TABLE policies ENABLE ROW LEVEL SECURITY;

CREATE POLICY policies_tenant_isolation ON policies
    FOR ALL TO public
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

-- ------------------------------------------------------------------------------------------------
-- POLICY EVALUATIONS CACHE
-- ------------------------------------------------------------------------------------------------
-- Caches policy evaluation results for performance optimization.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS policy_evaluations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    resource_id UUID NOT NULL REFERENCES resources(id),
    action_id UUID NOT NULL REFERENCES actions(id),
    context_hash VARCHAR(64) NOT NULL,             -- Hash of evaluation context
    decision VARCHAR(20) NOT NULL CHECK (decision IN ('ALLOW', 'DENY', 'NOT_APPLICABLE')),
    applicable_policies UUID[] DEFAULT '{}',       -- Array of policy IDs that fired
    evaluation_time_ms INTEGER,                    -- Performance metric
    evaluated_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ DEFAULT (NOW() + INTERVAL '1 hour')
);

COMMENT ON TABLE policy_evaluations IS 
'Caches ABAC policy evaluation results for performance optimization with configurable TTL and context tracking.';

COMMENT ON COLUMN policy_evaluations.context_hash IS 'SHA-256 hash of evaluation context for cache key uniqueness';
COMMENT ON COLUMN policy_evaluations.applicable_policies IS 'Array of policy UUIDs that were evaluated and fired';
COMMENT ON COLUMN policy_evaluations.evaluation_time_ms IS 'Policy evaluation time in milliseconds for performance monitoring';

-- Enable RLS and create policies
ALTER TABLE policy_evaluations ENABLE ROW LEVEL SECURITY;

CREATE POLICY policy_evaluations_tenant_isolation ON policy_evaluations
    FOR ALL TO public
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

-- ================================================================================================
-- AUDIT AND COMPLIANCE
-- ================================================================================================

-- ------------------------------------------------------------------------------------------------
-- AUDIT LOG
-- ------------------------------------------------------------------------------------------------
-- Comprehensive audit logging with compliance flags and risk scoring.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    event_category VARCHAR(50) DEFAULT 'ACCESS'
        CHECK (event_category IN ('ACCESS', 'ADMIN', 'DATA', 'AUTH', 'SYSTEM', 'COMPLIANCE')),
    severity VARCHAR(20) DEFAULT 'INFO'
        CHECK (severity IN ('LOW', 'INFO', 'WARN', 'HIGH', 'CRITICAL')),
    user_id UUID REFERENCES users(id),
    target_user_id UUID REFERENCES users(id),      -- For admin actions on other users
    entity_id UUID REFERENCES entities(uuid),
    resource_id UUID REFERENCES resources(id),
    action_id UUID REFERENCES actions(id),
    role_id UUID REFERENCES roles(id),
    permission_id UUID REFERENCES permissions(id),
    decision VARCHAR(20),                          -- ALLOW/DENY for access attempts
    reason TEXT,                                   -- Human-readable reason
    risk_score INTEGER DEFAULT 0,                 -- Calculated risk score (0-100)
    context JSONB DEFAULT '{}'::jsonb,             -- Additional event context
    ip_address INET,
    user_agent TEXT,
    session_id UUID REFERENCES user_sessions(id),
    compliance_flags JSONB DEFAULT '{}'::jsonb,   -- GDPR, SOX, HIPAA, etc.
    created_at TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE audit_log IS 
'Comprehensive audit log with compliance tracking, risk scoring, and detailed context for security monitoring and regulatory compliance.';

COMMENT ON COLUMN audit_log.event_category IS 'Event category: ACCESS (authorization), ADMIN (administrative), DATA (data access), AUTH (authentication), SYSTEM (system events), COMPLIANCE (regulatory)';
COMMENT ON COLUMN audit_log.severity IS 'Event severity level: LOW, INFO, WARN, HIGH, CRITICAL';
COMMENT ON COLUMN audit_log.target_user_id IS 'Target user for administrative actions (e.g., admin modifying another user)';
COMMENT ON COLUMN audit_log.risk_score IS 'Calculated risk score from 0-100 based on action, context, and user behavior';
COMMENT ON COLUMN audit_log.compliance_flags IS 'JSONB containing compliance-related flags (GDPR, SOX, HIPAA, PCI, etc.)';

-- Enable RLS and create policies
ALTER TABLE audit_log ENABLE ROW LEVEL SECURITY;

CREATE POLICY audit_log_tenant_isolation ON audit_log
    FOR ALL TO public
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

-- ------------------------------------------------------------------------------------------------
-- ACCESS REQUESTS
-- ------------------------------------------------------------------------------------------------
-- Approval workflow for access requests with business justification and lifecycle management.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS access_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    requester_id UUID NOT NULL REFERENCES users(id),
    target_user_id UUID REFERENCES users(id),      -- If requesting for someone else
    entity_id UUID NOT NULL REFERENCES entities(uuid),
    request_type VARCHAR(20) NOT NULL
        CHECK (request_type IN ('ROLE_ASSIGNMENT', 'PERMISSION_GRANT', 'RESOURCE_ACCESS', 'ELEVATION')),
    role_id UUID REFERENCES roles(id),
    permission_id UUID REFERENCES permissions(id),
    resource_id UUID REFERENCES resources(id),
    justification TEXT NOT NULL,
    business_reason VARCHAR(500),
    duration_hours INTEGER,                        -- For temporary access
    approval_status VARCHAR(20) DEFAULT 'PENDING'
        CHECK (approval_status IN ('PENDING', 'APPROVED', 'REJECTED', 'EXPIRED', 'REVOKED')),
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    approval_comments TEXT,
    expires_at TIMESTAMPTZ,
    auto_revoke BOOLEAN DEFAULT true,              -- Auto-revoke when expires
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE access_requests IS 
'Access request approval workflow with business justification, lifecycle management, and automatic revocation for governance and compliance.';

COMMENT ON COLUMN access_requests.request_type IS 'Type of access request: ROLE_ASSIGNMENT, PERMISSION_GRANT, RESOURCE_ACCESS, ELEVATION';
COMMENT ON COLUMN access_requests.target_user_id IS 'User receiving the access (if different from requester)';
COMMENT ON COLUMN access_requests.duration_hours IS 'Requested access duration in hours for temporary access';
COMMENT ON COLUMN access_requests.auto_revoke IS 'Whether to automatically revoke access when it expires';

-- Enable RLS and create policies
ALTER TABLE access_requests ENABLE ROW LEVEL SECURITY;

CREATE POLICY access_requests_tenant_isolation ON access_requests
    FOR ALL TO public
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

-- ================================================================================================
-- PERFORMANCE INDEXES
-- ================================================================================================

-- Person table indexes
CREATE INDEX idx_persons_tenant_id ON persons(tenant_id);
CREATE INDEX idx_persons_entity_id ON persons(entity_id);
CREATE INDEX idx_persons_person_type ON persons(person_type);
CREATE INDEX idx_persons_active ON persons(is_active) WHERE deleted_at IS NULL;
CREATE INDEX idx_persons_email ON persons(email) WHERE email IS NOT NULL AND deleted_at IS NULL;

-- Employee table indexes
CREATE INDEX idx_employees_tenant_id ON employees(tenant_id);
CREATE INDEX idx_employees_person_id ON employees(person_id);
CREATE INDEX idx_employees_entity_id ON employees(entity_id);
CREATE INDEX idx_employees_department_id ON employees(department_id);
CREATE INDEX idx_employees_manager_id ON employees(manager_id);
CREATE INDEX idx_employees_employment_status ON employees(employment_status);
CREATE INDEX idx_employees_security_level ON employees(security_level);

-- User table indexes
CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_entity_id ON users(entity_id);
CREATE INDEX idx_users_person_id ON users(person_id);
CREATE INDEX idx_users_employee_id ON users(employee_id);
CREATE INDEX idx_users_user_type ON users(user_type);
CREATE INDEX idx_users_account_status ON users(account_status);
CREATE INDEX idx_users_active ON users(is_active) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;

-- Session table indexes
CREATE INDEX idx_user_sessions_tenant_id ON user_sessions(tenant_id);
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_token ON user_sessions(session_token);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX idx_user_sessions_active ON user_sessions(is_active);

-- RBAC table indexes
CREATE INDEX idx_modules_tenant_id ON modules(tenant_id);
CREATE INDEX idx_modules_category ON modules(category);

CREATE INDEX idx_resources_tenant_module ON resources(tenant_id, module_id);
CREATE INDEX idx_resources_entity_id ON resources(entity_id);
CREATE INDEX idx_resources_parent_resource_id ON resources(parent_resource_id);
CREATE INDEX idx_resources_resource_type ON resources(resource_type);

CREATE INDEX idx_actions_tenant_id ON actions(tenant_id);
CREATE INDEX idx_actions_action_type ON actions(action_type);
CREATE INDEX idx_actions_action_category ON actions(action_category);
CREATE INDEX idx_actions_risk_level ON actions(risk_level);

CREATE INDEX idx_roles_tenant_id ON roles(tenant_id);
CREATE INDEX idx_roles_entity_id ON roles(entity_id);
CREATE INDEX idx_roles_module_id ON roles(module_id);
CREATE INDEX idx_roles_parent_role_id ON roles(parent_role_id);
CREATE INDEX idx_roles_role_type ON roles(role_type);

CREATE INDEX idx_permissions_tenant_id ON permissions(tenant_id);
CREATE INDEX idx_permissions_resource_action ON permissions(resource_id, action_id);

CREATE INDEX idx_role_permissions_tenant_id ON role_permissions(tenant_id);
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);
CREATE INDEX idx_role_permissions_entity_scope ON role_permissions(entity_scope);

CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX idx_user_roles_entity_id ON user_roles(entity_id);
CREATE INDEX idx_user_roles_assignment_type ON user_roles(assignment_type);
CREATE INDEX idx_user_roles_expires_at ON user_roles(expires_at);

CREATE INDEX idx_user_permissions_tenant_id ON user_permissions(tenant_id);
CREATE INDEX idx_user_permissions_user_id ON user_permissions(user_id);
CREATE INDEX idx_user_permissions_permission_id ON user_permissions(permission_id);
CREATE INDEX idx_user_permissions_entity_id ON user_permissions(entity_id);

-- ABAC table indexes
CREATE INDEX idx_attribute_definitions_tenant_id ON attribute_definitions(tenant_id);
CREATE INDEX idx_attribute_definitions_category ON attribute_definitions(category);
CREATE INDEX idx_attribute_definitions_data_type ON attribute_definitions(data_type);

CREATE INDEX idx_policies_tenant_id ON policies(tenant_id);
CREATE INDEX idx_policies_entity_id ON policies(entity_id);
CREATE INDEX idx_policies_policy_type ON policies(policy_type);
CREATE INDEX idx_policies_priority ON policies(priority DESC);
CREATE INDEX idx_policies_category ON policies(category);

CREATE INDEX idx_policy_evaluations_tenant_id ON policy_evaluations(tenant_id);
CREATE INDEX idx_policy_evaluations_user_resource_action ON policy_evaluations(user_id, resource_id, action_id);
CREATE INDEX idx_policy_evaluations_context_hash ON policy_evaluations(context_hash);
CREATE INDEX idx_policy_evaluations_expires_at ON policy_evaluations(expires_at);

-- Audit table indexes
CREATE INDEX idx_audit_log_tenant_id ON audit_log(tenant_id);
CREATE INDEX idx_audit_log_user_id ON audit_log(user_id);
CREATE INDEX idx_audit_log_entity_id ON audit_log(entity_id);
CREATE INDEX idx_audit_log_event_type ON audit_log(event_type);
CREATE INDEX idx_audit_log_event_category ON audit_log(event_category);
CREATE INDEX idx_audit_log_severity ON audit_log(severity);
CREATE INDEX idx_audit_log_created_at ON audit_log(created_at);
CREATE INDEX idx_audit_log_risk_score ON audit_log(risk_score);

CREATE INDEX idx_access_requests_tenant_id ON access_requests(tenant_id);
CREATE INDEX idx_access_requests_requester_id ON access_requests(requester_id);
CREATE INDEX idx_access_requests_approval_status ON access_requests(approval_status);
CREATE INDEX idx_access_requests_expires_at ON access_requests(expires_at);

-- JSONB indexes for ABAC attributes
CREATE INDEX idx_persons_security_attributes ON persons USING GIN(security_attributes);
CREATE INDEX idx_employees_access_attributes ON employees USING GIN(access_attributes);
CREATE INDEX idx_users_user_attributes ON users USING GIN(user_attributes);
CREATE INDEX idx_users_settings ON users USING GIN(settings);
CREATE INDEX idx_resources_resource_attributes ON resources USING GIN(resource_attributes);
CREATE INDEX idx_roles_permissions ON roles USING GIN(permissions);
CREATE INDEX idx_roles_entity_scope ON roles USING GIN(entity_scope);
CREATE INDEX idx_roles_conditions ON roles USING GIN(conditions);
CREATE INDEX idx_policies_target ON policies USING GIN(target);
CREATE INDEX idx_policies_rule ON policies USING GIN(rule);
CREATE INDEX idx_audit_log_context ON audit_log USING GIN(context);
CREATE INDEX idx_audit_log_compliance_flags ON audit_log USING GIN(compliance_flags);

-- ================================================================================================
-- TRIGGERS AND FUNCTIONS
-- ================================================================================================

-- ------------------------------------------------------------------------------------------------
-- Updated timestamp trigger function
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION update_updated_at_column() IS 
'Trigger function to automatically update the updated_at timestamp when a record is modified.';

-- Apply updated_at triggers to relevant tables
CREATE TRIGGER update_persons_updated_at BEFORE UPDATE ON persons 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_employees_updated_at BEFORE UPDATE ON employees 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_policies_updated_at BEFORE UPDATE ON policies 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_access_requests_updated_at BEFORE UPDATE ON access_requests 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ------------------------------------------------------------------------------------------------
-- Tenant isolation validation trigger
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION enforce_tenant_isolation()
RETURNS TRIGGER AS $$
BEGIN
    -- Ensure all foreign key references belong to the same tenant
    IF TG_TABLE_NAME = 'persons' THEN
        -- Validate entity belongs to same tenant
        IF NOT EXISTS (
            SELECT 1 FROM entities e 
            JOIN tenants t ON e.tenant_id = t.id
            WHERE e.uuid = NEW.entity_id AND t.id = NEW.tenant_id
        ) THEN
            RAISE EXCEPTION 'Entity % does not belong to tenant %', NEW.entity_id, NEW.tenant_id;
        END IF;
    END IF;
    
    -- Add similar validations for other tables as needed
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION enforce_tenant_isolation() IS 
'Trigger function to enforce tenant isolation by validating that all foreign key references belong to the same tenant.';

CREATE TRIGGER enforce_persons_tenant_isolation BEFORE INSERT OR UPDATE ON persons 
    FOR EACH ROW EXECUTE FUNCTION enforce_tenant_isolation();

-- ------------------------------------------------------------------------------------------------
-- Role hierarchy validation and level calculation
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION validate_role_hierarchy()
RETURNS TRIGGER AS $$
DECLARE
    max_depth INTEGER := 10;
    current_depth INTEGER := 0;
    current_role_id UUID;
BEGIN
    -- If no parent role, set level to 0
    IF NEW.parent_role_id IS NULL THEN
        NEW.level = 0;
        RETURN NEW;
    END IF;
    
    -- Ensure parent role belongs to same tenant
    IF NOT EXISTS (
        SELECT 1 FROM roles 
        WHERE id = NEW.parent_role_id AND tenant_id = NEW.tenant_id AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION 'Parent role % does not belong to tenant % or is deleted', NEW.parent_role_id, NEW.tenant_id;
    END IF;
    
    -- Check for cycles and calculate depth
    current_role_id := NEW.parent_role_id;
    current_depth := 1;
    
    WHILE current_role_id IS NOT NULL AND current_depth <= max_depth LOOP
        -- Check if we've hit the new role (cycle detection)
        IF current_role_id = NEW.id THEN
            RAISE EXCEPTION 'Role hierarchy cycle detected for role %', NEW.id;
        END IF;
        
        -- Get the next parent
        SELECT parent_role_id INTO current_role_id
        FROM roles 
        WHERE id = current_role_id AND tenant_id = NEW.tenant_id AND deleted_at IS NULL;
        
        current_depth := current_depth + 1;
    END LOOP;
    
    IF current_depth > max_depth THEN
        RAISE EXCEPTION 'Role hierarchy depth exceeds maximum of %', max_depth;
    END IF;
    
    -- Set the level
    NEW.level = current_depth - 1;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION validate_role_hierarchy() IS 
'Validates role hierarchy integrity, prevents cycles, enforces depth limits, and calculates hierarchy levels.';

CREATE TRIGGER validate_role_hierarchy_trigger 
    BEFORE INSERT OR UPDATE ON roles 
    FOR EACH ROW EXECUTE FUNCTION validate_role_hierarchy();

-- ------------------------------------------------------------------------------------------------
-- User permission evaluation function with ABAC support
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION user_has_permission(
    p_user_id UUID,
    p_resource_name VARCHAR(100),
    p_action_name VARCHAR(100),
    p_tenant_id INTEGER,
    p_entity_id UUID DEFAULT NULL,
    p_context JSONB DEFAULT '{}'::jsonb
)
RETURNS BOOLEAN AS $$
DECLARE
    v_has_permission BOOLEAN := FALSE;
    v_user_context JSONB;
    v_cache_key VARCHAR(64);
    v_cached_result BOOLEAN;
BEGIN
    -- Generate cache key
    v_cache_key := encode(digest(
        p_user_id::text || p_resource_name || p_action_name || 
        COALESCE(p_entity_id::text, '') || p_context::text, 'sha256'
    ), 'hex');
    
    -- Check cache first
    SELECT decision = 'ALLOW' INTO v_cached_result
    FROM policy_evaluations
    WHERE user_id = p_user_id
        AND context_hash = v_cache_key
        AND expires_at > NOW()
        AND tenant_id = p_tenant_id;
    
    IF FOUND THEN
        RETURN v_cached_result;
    END IF;
    
    -- Build comprehensive user context for ABAC evaluation
    SELECT jsonb_build_object(
        'user_id', u.id,
        'user_type', u.user_type,
        'account_status', u.account_status,
        'user_attributes', COALESCE(u.user_attributes, '{}'::jsonb),
        'person_attributes', COALESCE(p.security_attributes, '{}'::jsonb),
        'employee_attributes', COALESCE(e.access_attributes, '{}'::jsonb),
        'employment_status', e.employment_status,
        'security_level', COALESCE(e.security_level, 0),
        'entity_id', u.entity_id,
        'department_id', e.department_id,
        'context', p_context
    ) INTO v_user_context
    FROM users u
    LEFT JOIN persons p ON u.person_id = p.id
    LEFT JOIN employees e ON u.employee_id = e.id
    WHERE u.id = p_user_id AND u.tenant_id = p_tenant_id;
    
    -- Check for explicit DENY in direct user permissions first
    SELECT true INTO v_has_permission
    FROM user_permissions up
    JOIN permissions perm ON up.permission_id = perm.id
    JOIN resources r ON perm.resource_id = r.id
    JOIN actions a ON perm.action_id = a.id
    WHERE up.user_id = p_user_id
        AND up.tenant_id = p_tenant_id
        AND r.name = p_resource_name
        AND a.name = p_action_name
        AND up.is_active = true
        AND up.effect = 'DENY'
        AND (up.expires_at IS NULL OR up.expires_at > NOW())
        AND (p_entity_id IS NULL OR up.entity_id = p_entity_id);
    
    -- If explicit DENY found, return false immediately
    IF FOUND THEN
        RETURN FALSE;
    END IF;
    
    -- Check direct user permissions for ALLOW
    SELECT true INTO v_has_permission
    FROM user_permissions up
    JOIN permissions perm ON up.permission_id = perm.id
    JOIN resources r ON perm.resource_id = r.id
    JOIN actions a ON perm.action_id = a.id
    WHERE up.user_id = p_user_id
        AND up.tenant_id = p_tenant_id
        AND r.name = p_resource_name
        AND a.name = p_action_name
        AND up.is_active = true
        AND up.effect = 'ALLOW'
        AND (up.expires_at IS NULL OR up.expires_at > NOW())
        AND (p_entity_id IS NULL OR up.entity_id = p_entity_id);
    
    -- If no direct permission, check role-based permissions
    IF NOT FOUND THEN
        SELECT true INTO v_has_permission
        FROM user_roles ur
        JOIN role_permissions rp ON ur.role_id = rp.role_id
        JOIN permissions perm ON rp.permission_id = perm.id
        JOIN resources r ON perm.resource_id = r.id
        JOIN actions a ON perm.action_id = a.id
        WHERE ur.user_id = p_user_id
            AND r.name = p_resource_name
            AND a.name = p_action_name
            AND ur.is_active = true
            AND rp.is_active = true
            AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
            AND (p_entity_id IS NULL OR ur.entity_id = p_entity_id OR rp.entity_scope = p_entity_id);
    END IF;
    
    RETURN COALESCE(v_has_permission, FALSE);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMENT ON FUNCTION user_has_permission(UUID, VARCHAR, VARCHAR, INTEGER, UUID, JSONB) IS 
'Evaluates user permissions with ABAC context, caching, and comprehensive policy evaluation including direct permissions and role-based permissions.';

-- ------------------------------------------------------------------------------------------------
-- Cleanup function for expired data and maintenance
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION cleanup_expired_data(p_tenant_id UUID DEFAULT NULL)
RETURNS INTEGER AS $$
DECLARE
    v_cleanup_count INTEGER := 0;
    v_tenant_filter TEXT := '';
BEGIN
    -- Build tenant filter if specified
    IF p_tenant_id IS NOT NULL THEN
        v_tenant_filter := ' AND tenant_id = ' || p_tenant_id;
    END IF;
    
    -- Clean expired sessions
    EXECUTE 'DELETE FROM user_sessions WHERE expires_at < NOW()' || v_tenant_filter;
    GET DIAGNOSTICS v_cleanup_count = ROW_COUNT;
    
    -- Clean expired policy evaluations
    EXECUTE 'DELETE FROM policy_evaluations WHERE expires_at < NOW()' || v_tenant_filter;
    
    -- Deactivate expired user roles
    EXECUTE 'UPDATE user_roles SET is_active = false 
             WHERE expires_at < NOW() AND is_active = true' || 
             CASE WHEN p_tenant_id IS NOT NULL THEN 
                ' AND EXISTS (SELECT 1 FROM users WHERE id = user_roles.user_id AND tenant_id = ' || p_tenant_id || ')'
             ELSE '' END;
    
    -- Deactivate expired user permissions
    EXECUTE 'UPDATE user_permissions SET is_active = false 
             WHERE expires_at < NOW() AND is_active = true' || v_tenant_filter;
    
    -- Expire approved access requests
    EXECUTE 'UPDATE access_requests SET approval_status = ''EXPIRED''
             WHERE expires_at < NOW() AND approval_status = ''APPROVED''.' || v_tenant_filter;
    
    RETURN v_cleanup_count;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMENT ON FUNCTION cleanup_expired_data(UUID) IS 
'Cleans up expired sessions, policy evaluations, user roles, permissions, and access requests. Can be run for all tenants or a specific tenant.';

-- ================================================================================================
-- INITIAL SYSTEM DATA
-- ================================================================================================

-- Create default system roles function
CREATE OR REPLACE FUNCTION create_default_system_data(p_tenant_id UUID)
RETURNS VOID AS $$
DECLARE
    v_core_module_id UUID;
    v_user_mgmt_module_id UUID;
    v_admin_role_id UUID;
    v_user_role_id UUID;
    v_create_action_id UUID;
    v_read_action_id UUID;
    v_update_action_id UUID;
    v_delete_action_id UUID;
BEGIN
    -- Insert default modules
    INSERT INTO modules (tenant_id, name, display_name, category, description) 
    VALUES 
        (p_tenant_id, 'CORE', 'Core System', 'CORE', 'Core system functionality and resources'),
        (p_tenant_id, 'USER_MANAGEMENT', 'User Management', 'ADMIN', 'User and role management functionality')
    RETURNING id INTO v_core_module_id, v_user_mgmt_module_id;
    
    -- Insert default actions
    INSERT INTO actions (tenant_id, name, display_name, action_type, action_category, description)
    VALUES
        (p_tenant_id, 'CREATE', 'Create', 'CREATE', 'STANDARD', 'Create new records'),
        (p_tenant_id, 'READ', 'Read', 'READ', 'STANDARD', 'View and read records'),
        (p_tenant_id, 'UPDATE', 'Update', 'UPDATE', 'STANDARD', 'Modify existing records'),
        (p_tenant_id, 'DELETE', 'Delete', 'DELETE', 'STANDARD', 'Delete records'),
        (p_tenant_id, 'EXECUTE', 'Execute', 'EXECUTE', 'STANDARD', 'Execute functions and processes'),
        (p_tenant_id, 'APPROVE', 'Approve', 'APPROVE', 'ADMINISTRATIVE', 'Approve requests and workflows'),
        (p_tenant_id, 'ADMIN', 'Administer', 'EXECUTE', 'ADMINISTRATIVE', 'Administrative access and control');
    
    -- Insert default attribute definitions for ABAC
    INSERT INTO attribute_definitions (tenant_id, name, display_name, data_type, category, description)
    VALUES
        (p_tenant_id, 'department', 'Department', 'STRING', 'USER', 'User department affiliation'),
        (p_tenant_id, 'security_level', 'Security Level', 'NUMBER', 'USER', 'User security clearance level (0-10)'),
        (p_tenant_id, 'location', 'Location', 'STRING', 'USER', 'User office or work location'),
        (p_tenant_id, 'employment_type', 'Employment Type', 'ENUM', 'USER', 'Type of employment (FULL_TIME, PART_TIME, CONTRACT, INTERN)'),
        (p_tenant_id, 'classification', 'Classification', 'STRING', 'RESOURCE', 'Resource security classification level'),
        (p_tenant_id, 'sensitivity', 'Sensitivity', 'STRING', 'RESOURCE', 'Resource data sensitivity level'),
        (p_tenant_id, 'time_of_day', 'Time of Day', 'TIME', 'ENVIRONMENT', 'Current time for time-based access control'),
        (p_tenant_id, 'ip_range', 'IP Range', 'STRING', 'ENVIRONMENT', 'Allowed IP address ranges'),
        (p_tenant_id, 'device_type', 'Device Type', 'STRING', 'ENVIRONMENT', 'Type of device accessing the system');
    
    RAISE NOTICE 'Default system data created for tenant %', p_tenant_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION create_default_system_data(UUID) IS 
'Creates default modules, actions, and attribute definitions for a new tenant. Should be called after tenant creation.';

-- ================================================================================================
-- UTILITY VIEWS
-- ================================================================================================

-- ------------------------------------------------------------------------------------------------
-- Comprehensive user view with all related data
-- ------------------------------------------------------------------------------------------------
CREATE VIEW user_complete_view AS
SELECT 
    u.id as user_id,
    u.tenant_id,
    u.entity_id,
    u.username,
    u.email,
    u.user_type,
    u.account_status,
    u.is_active as user_active,
    u.last_login_at,
    u.mfa_enabled,
    p.id as person_id,
    p.first_name,
    p.last_name,
    p.middle_name,
    p.person_type,
    e.id as employee_id,
    e.employee_number,
    e.position_title,
    e.department_id,
    e.employment_status,
    e.security_level,
    -- Combine all attributes for ABAC evaluation
    COALESCE(p.security_attributes, '{}'::jsonb) || 
    COALESCE(e.access_attributes, '{}'::jsonb) || 
    COALESCE(u.user_attributes, '{}'::jsonb) as combined_attributes,
    -- Aggregate role information - FIXED
    array_agg(DISTINCT r.name ORDER BY r.name) as role_names,
    array_agg(DISTINCT r.id ORDER BY r.id) as role_ids,  -- Changed to ORDER BY r.id
    count(DISTINCT ur.id) FILTER (WHERE ur.is_active = true) as active_role_count
FROM users u
LEFT JOIN persons p ON u.person_id = p.id
LEFT JOIN employees e ON u.employee_id = e.id
LEFT JOIN user_roles ur ON u.id = ur.user_id AND ur.is_active = true AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
LEFT JOIN roles r ON ur.role_id = r.id AND r.is_active = true AND r.deleted_at IS NULL
WHERE u.deleted_at IS NULL
GROUP BY u.id, u.tenant_id, u.entity_id, u.username, u.email, u.user_type, 
         u.account_status, u.is_active, u.last_login_at, u.mfa_enabled,
         p.id, p.first_name, p.last_name, p.middle_name, p.person_type,
         e.id, e.employee_number, e.position_title, e.department_id, 
         e.employment_status, e.security_level,
         p.security_attributes, e.access_attributes, u.user_attributes;

COMMENT ON VIEW user_complete_view IS 
'Comprehensive view combining user, person, and employee data with role aggregations and combined ABAC attributes for authorization decisions.';

-- ------------------------------------------------------------------------------------------------
-- Role permissions summary view
-- ------------------------------------------------------------------------------------------------
CREATE VIEW role_permissions_summary AS
SELECT 
    r.tenant_id,
    r.id as role_id,
    r.name as role_name,
    r.display_name as role_display_name,
    r.role_type,
    r.level as hierarchy_level,
    r.module_id,
    m.name as module_name,
    array_agg(DISTINCT res.name ORDER BY res.name) FILTER (WHERE res.name IS NOT NULL) as resource_names,
    array_agg(DISTINCT a.name ORDER BY a.name) FILTER (WHERE a.name IS NOT NULL) as action_names,
    count(DISTINCT rp.permission_id) as permission_count,
    count(DISTINCT ur.user_id) FILTER (WHERE ur.is_active = true) as assigned_user_count
FROM roles r
LEFT JOIN modules m ON r.module_id = m.id
LEFT JOIN role_permissions rp ON r.id = rp.role_id AND rp.is_active = true
LEFT JOIN permissions p ON rp.permission_id = p.id AND p.is_active = true
LEFT JOIN resources res ON p.resource_id = res.id AND res.is_active = true
LEFT JOIN actions a ON p.action_id = a.id AND a.is_active = true
LEFT JOIN user_roles ur ON r.id = ur.role_id AND ur.is_active = true AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
WHERE r.is_active = true AND r.deleted_at IS NULL
GROUP BY r.tenant_id, r.id, r.name, r.display_name, r.role_type, r.level, r.module_id, m.name;

COMMENT ON VIEW role_permissions_summary IS 
'Summary view of roles with their permissions, resources, actions, and user assignment counts for role management and analysis.';

-- ------------------------------------------------------------------------------------------------
-- Audit summary view for security monitoring
-- ------------------------------------------------------------------------------------------------
CREATE VIEW audit_summary_view AS
SELECT 
    tenant_id,
    event_category,
    severity,
    DATE_TRUNC('hour', created_at) as hour_bucket,
    count(*) as event_count,
    count(DISTINCT user_id) as unique_users,
    avg(risk_score) as avg_risk_score,
    max(risk_score) as max_risk_score,
    count(*) FILTER (WHERE decision = 'DENY') as denied_attempts,
    count(*) FILTER (WHERE decision = 'ALLOW') as allowed_attempts
FROM audit_log
WHERE created_at >= NOW() - INTERVAL '7 days'
GROUP BY tenant_id, event_category, severity, DATE_TRUNC('hour', created_at)
ORDER BY hour_bucket DESC, event_count DESC;

COMMENT ON VIEW audit_summary_view IS 
'Hourly audit event summary for the last 7 days with risk metrics and access decision counts for security monitoring dashboards.';
-- Additional RLS policies for remaining tables that need validation
ALTER TABLE audit_log ENABLE ROW LEVEL SECURITY;
CREATE POLICY audit_log_tenant_isolation ON audit_log
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    );

CREATE POLICY audit_log_admin_access ON audit_log
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);

ALTER TABLE access_requests ENABLE ROW LEVEL SECURITY;
CREATE POLICY access_requests_tenant_isolation ON access_requests
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    );

CREATE POLICY access_requests_admin_access ON access_requests
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);


-- ================================================================================================
-- SCHEMA COMPLETION
-- ================================================================================================

-- Add final schema-level comment
COMMENT ON SCHEMA public IS 
' RBAC/ABAC authorization schema with UUID primary keys, Row Level Security, comprehensive audit logging, and ABAC policy engine for multi-tenant applications.';

-- Success message
DO $$
BEGIN
    RAISE NOTICE ' RBAC/ABAC schema with UUID keys and RLS successfully created!';
    RAISE NOTICE 'Remember to:';
    RAISE NOTICE '1. Set app.current_tenant_id session variable for RLS';
    RAISE NOTICE '2. Call create_default_system_data(tenant_id) for new tenants';
    RAISE NOTICE '3. Run cleanup_expired_data() periodically for maintenance';
    RAISE NOTICE '4. Configure appropriate database roles and permissions';
END;
$$;
