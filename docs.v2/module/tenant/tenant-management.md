# Multi-Tenant ERP System Documentation

## 🏢 Overview

The ERP system implements a robust multi-tenant architecture using Row-Level Security (RLS) with PostgreSQL and Go/sqlc for type-safe database operations. Each tenant represents a distinct business entity with complete data isolation while sharing infrastructure for optimal resource utilization.

## 🎯 Architecture Design

### Technology Stack Integration

- **Backend**: Go with Gin framework
- **Database**: PostgreSQL with Row-Level Security (RLS)
- **Query Generation**: sqlc for type-safe SQL queries
- **Connection Pooling**: pgxpool for efficient connection management
- **Transaction Management**: pgx/v5 for robust transaction handling

### Multi-Tenancy Strategy

The system uses PostgreSQL's Row-Level Security with a Go-based tenant context management layer:

```sql
-- Global RLS enablement
SET row_security = on;

-- Tenant context management functions
CREATE OR REPLACE FUNCTION set_tenant_context(tenant_id UUID)
RETURNS VOID AS $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM tenants
        WHERE id = tenant_id AND status = 'active' AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION 'Invalid or inactive tenant: %', tenant_id;
    END IF;
    
    PERFORM set_config('app.current_tenant_id', tenant_id::text, true);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

CREATE OR REPLACE FUNCTION current_tenant_id() RETURNS UUID AS $$
BEGIN
    RETURN COALESCE(nullif(current_setting('app.current_tenant_id', true), ''), NULL)::UUID;
EXCEPTION
    WHEN OTHERS THEN
        RETURN NULL;
END;
$$ LANGUAGE plpgsql;
```

## 🏗️ Database Structure

### Core Tenants Table

```sql
CREATE TABLE tenants (
    -- Primary identifiers
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    slug VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL UNIQUE,

    -- Contact and access information
    email VARCHAR(255) NOT NULL,
    subdomain VARCHAR(63) UNIQUE,
    domain VARCHAR(255),

    -- Status and operational settings
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'suspended', 'pending', 'archived', 'terminated')),
    timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',
    currency_code CHAR(3) NOT NULL DEFAULT 'USD',

    -- Business classification
    industry VARCHAR(50),
    company_size VARCHAR(20)
        CHECK (company_size IN ('startup', 'small', 'medium', 'large', 'enterprise')),

    -- Compliance and legal information
    tax_id VARCHAR(50),
    registration_number VARCHAR(50),
    legal_entity_type VARCHAR(50),
    country_code CHAR(2),
    fiscal_year_start DATE,

    -- Billing information
    billing_email VARCHAR(255),
    billing_address JSONB,
    payment_method_id VARCHAR(255),
    subscription_plan VARCHAR(50),

    -- Flexible configuration storage
    metadata JSONB DEFAULT '{}' NOT NULL,
    settings JSONB DEFAULT '{}' NOT NULL,

    -- Audit and lifecycle management
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ, -- Soft delete support
    last_activity_at TIMESTAMPTZ DEFAULT NOW(),

    -- Constraint to ensure soft-deleted tenants can reuse slugs/subdomains
    CONSTRAINT unique_active_slug UNIQUE NULLS NOT DISTINCT (slug, (CASE WHEN deleted_at IS NULL THEN NULL ELSE 'deleted' END)),
    CONSTRAINT unique_active_subdomain UNIQUE NULLS NOT DISTINCT (subdomain, (CASE WHEN deleted_at IS NULL THEN NULL ELSE 'deleted' END)),
    CONSTRAINT valid_status CHECK (status IN ('active', 'suspended', 'terminated', 'pending', 'archived')),
    CONSTRAINT valid_country CHECK (country_code ~ '^[A-Z]{2}$'),
    CONSTRAINT valid_currency CHECK (currency_code ~ '^[A-Z]{3}$')
);

-- Performance and lookup indexes
CREATE INDEX idx_tenants_slug ON tenants(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_tenants_status ON tenants(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_tenants_subdomain ON tenants(subdomain) WHERE subdomain IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_tenants_last_activity ON tenants(last_activity_at);
CREATE INDEX idx_tenants_metadata_gin ON tenants USING GIN (metadata);
CREATE INDEX idx_tenants_settings_gin ON tenants USING GIN (settings);
```

### Tenant Configurations

```sql
CREATE TABLE tenant_configurations (
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE PRIMARY KEY,

    -- Resource limits with progressive tiers
    max_users INT NOT NULL DEFAULT 100,
    max_entities INT NOT NULL DEFAULT 1000,
    max_transactions_per_month INT NOT NULL DEFAULT 10000,
    storage_quota BIGINT NOT NULL DEFAULT 1073741824, -- 1GB
    api_rate_limit_per_minute INT NOT NULL DEFAULT 100,
    api_rate_limit_per_hour INT NOT NULL DEFAULT 5000,

    -- Enabled modules configuration
    modules_enabled JSONB NOT NULL DEFAULT '[
        "accounting", 
        "inventory", 
        "contacts", 
        "sales"
    ]'::jsonb,

    -- Accounting and business logic preferences
    accounting_method VARCHAR(10) NOT NULL DEFAULT 'accrual'
        CHECK (accounting_method IN ('accrual', 'cash')),
    fiscal_year_start_month INT NOT NULL DEFAULT 1
        CHECK (fiscal_year_start_month BETWEEN 1 AND 12),
    default_currency CHAR(3) NOT NULL DEFAULT 'USD',
    multi_currency_enabled BOOLEAN NOT NULL DEFAULT false,

    -- Localization and formatting
    date_format VARCHAR(20) NOT NULL DEFAULT 'MM/DD/YYYY',
    number_format VARCHAR(20) NOT NULL DEFAULT 'US',
    language_code VARCHAR(5) NOT NULL DEFAULT 'en-US',
    decimal_places INT NOT NULL DEFAULT 2 CHECK (decimal_places BETWEEN 0 AND 4),

    -- Security policies
    password_policy JSONB NOT NULL DEFAULT '{
        "min_length": 8,
        "max_length": 128,
        "require_uppercase": true,
        "require_lowercase": true,
        "require_numbers": true,
        "require_symbols": false,
        "prevent_common_passwords": true,
        "password_history_count": 5,
        "max_login_attempts": 5,
        "lockout_duration_minutes": 30
    }'::jsonb,

    -- Session and security settings
    session_timeout_minutes INT NOT NULL DEFAULT 480, -- 8 hours
    require_2fa BOOLEAN NOT NULL DEFAULT false,
    allowed_ip_ranges JSONB DEFAULT '[]'::jsonb,

    -- Integration and webhook configuration
    webhook_endpoints JSONB DEFAULT '[]'::jsonb,
    api_settings JSONB DEFAULT '{
        "webhook_retry_attempts": 3,
        "webhook_timeout_seconds": 30,
        "enable_webhook_signatures": true
    }'::jsonb,

    -- Backup and retention policies
    backup_retention_days INT NOT NULL DEFAULT 30,
    auto_backup_enabled BOOLEAN NOT NULL DEFAULT true,

    -- Audit timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for module lookups
CREATE INDEX idx_tenant_configs_modules ON tenant_configurations USING GIN (modules_enabled);
```

<!-- ### Organization Hierarchy -->
<!---->
<!-- ```sql -->
<!-- -- Organization hierarchy within tenants -->
<!-- CREATE TABLE organizations ( -->
<!--     id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -->
<!--     tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE, -->
<!--     parent_id UUID REFERENCES organizations(id), -->
<!--     name VARCHAR(255) NOT NULL, -->
<!--     org_type VARCHAR(50) NOT NULL, -- company, division, department, team, subsidiary, branch -->
<!--     code VARCHAR(20) UNIQUE, -->
<!--     description TEXT, -->
<!---->
<!--     -- Address and contact -->
<!--     address JSONB, -->
<!--     phone VARCHAR(20), -->
<!--     email VARCHAR(255), -->
<!--     website VARCHAR(255), -->
<!---->
<!--     -- Financial settings -->
<!--     cost_center_code VARCHAR(20), -->
<!--     profit_center_code VARCHAR(20), -->
<!--     budget_allocated DECIMAL(15,2), -->
<!---->
<!--     -- Operational details -->
<!--     manager_id UUID, -->
<!--     is_active BOOLEAN DEFAULT true, -->
<!--     created_at TIMESTAMPTZ DEFAULT NOW(), -->
<!--     updated_at TIMESTAMPTZ DEFAULT NOW(), -->
<!---->
<!--     CONSTRAINT valid_org_type CHECK (org_type IN ('company', 'division', 'department', 'team', 'subsidiary', 'branch')) -->
<!-- ); -->
<!---->
<!-- -- RLS policy for organizations -->
<!-- ALTER TABLE organizations ENABLE ROW LEVEL SECURITY; -->
<!---->
<!-- CREATE POLICY organizations_tenant_isolation ON organizations -->
<!--     FOR ALL TO application_role -->
<!--     USING (tenant_id = current_tenant_id()); -->
<!---->
<!-- -- Indexes for performance -->
<!-- CREATE INDEX idx_organizations_tenant_id ON organizations(tenant_id); -->
<!-- CREATE INDEX idx_organizations_parent_id ON organizations(parent_id); -->
<!-- CREATE INDEX idx_organizations_manager_id ON organizations(manager_id); -->
<!-- ``` -->

### Usage Statistics with Granular Tracking

```sql
CREATE TABLE tenant_usage_stats (
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,

    -- User activity metrics
    active_users INT NOT NULL DEFAULT 0,
    peak_concurrent_users INT NOT NULL DEFAULT 0,
    new_users_registered INT NOT NULL DEFAULT 0,

    -- Data volume metrics
    total_entities INT NOT NULL DEFAULT 0,
    entities_created INT NOT NULL DEFAULT 0,
    entities_updated INT NOT NULL DEFAULT 0,
    entities_deleted INT NOT NULL DEFAULT 0,

    -- Transaction metrics
    total_transactions INT NOT NULL DEFAULT 0,
    successful_transactions INT NOT NULL DEFAULT 0,
    failed_transactions INT NOT NULL DEFAULT 0,

    -- Storage and performance metrics
    storage_used BIGINT NOT NULL DEFAULT 0,
    documents_stored INT NOT NULL DEFAULT 0,
    avg_document_size BIGINT NOT NULL DEFAULT 0,

    -- API usage metrics
    api_calls INT NOT NULL DEFAULT 0,
    api_calls_successful INT NOT NULL DEFAULT 0,
    api_calls_failed INT NOT NULL DEFAULT 0,
    api_bandwidth_bytes BIGINT NOT NULL DEFAULT 0,

    -- Performance metrics
    avg_response_time NUMERIC(10,3),
    p95_response_time NUMERIC(10,3),
    error_rate NUMERIC(7,6), -- Supports up to 99.9999%
    uptime_percentage NUMERIC(5,4), -- 99.99% format

    -- Business metrics
    monthly_revenue NUMERIC(15,4),
    transactions_revenue NUMERIC(15,4),

    -- Security metrics
    login_attempts INT NOT NULL DEFAULT 0,
    failed_login_attempts INT NOT NULL DEFAULT 0,
    security_incidents INT NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (tenant_id, period_start),
    
    -- Ensure period_end is always after period_start
    CHECK (period_end >= period_start),
    -- Ensure error rate is between 0 and 1
    CHECK (error_rate >= 0 AND error_rate <= 1),
    -- Ensure uptime percentage is between 0 and 1
    CHECK (uptime_percentage >= 0 AND uptime_percentage <= 1)
);

-- Optimized indexes for common queries
CREATE INDEX idx_usage_stats_tenant_period ON tenant_usage_stats(tenant_id, period_start DESC);
CREATE INDEX idx_usage_stats_period_range ON tenant_usage_stats(period_start, period_end);

-- Partitioning for better performance on large datasets
-- CREATE TABLE tenant_usage_stats_y2024m01 PARTITION OF tenant_usage_stats
-- FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
```

## 🔐 Security Implementation

### Row-Level Security Policies

```sql
-- Enable RLS on all tenant-related tables
ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_configurations ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_usage_stats ENABLE ROW LEVEL SECURITY;

-- Tenant isolation policies with role-based access
CREATE POLICY tenant_isolation_policy ON tenants
    FOR ALL TO application_role
    USING (
        id = current_tenant_id() OR 
        -- Allow system admin to access all tenants
        current_user = 'system_admin' OR
        -- Allow tenant admin to access their tenant
        (current_user = 'tenant_admin' AND id = current_tenant_id())
    );

CREATE POLICY tenant_configurations_isolation_policy ON tenant_configurations
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_usage_stats_isolation_policy ON tenant_usage_stats
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id());

-- Separate policies for different operations
CREATE POLICY tenant_select_policy ON tenants
    FOR SELECT TO readonly_role
    USING (id = current_tenant_id());

CREATE POLICY tenant_insert_policy ON tenants
    FOR INSERT TO application_role
    WITH CHECK (true); -- New tenants can be created

CREATE POLICY tenant_update_policy ON tenants
    FOR UPDATE TO application_role
    USING (id = current_tenant_id())
    WITH CHECK (id = current_tenant_id());

-- Prevent deletion unless authorized
CREATE POLICY tenant_delete_policy ON tenants
    FOR DELETE TO application_role
    USING (
        id = current_tenant_id() AND 
        current_user IN ('system_admin', 'tenant_admin')
    );
```

### Context Management

```sql
-- Tenant context validation
CREATE OR REPLACE FUNCTION validate_and_set_tenant_context(p_tenant_id UUID)
RETURNS TABLE(tenant_name TEXT, tenant_status TEXT) AS $$
DECLARE
    v_tenant_record RECORD;
BEGIN
    -- Validate and fetch tenant information
    SELECT id, name, status, deleted_at, last_activity_at
    INTO v_tenant_record
    FROM tenants
    WHERE id = p_tenant_id;

    -- Check if tenant exists
    IF v_tenant_record.id IS NULL THEN
        RAISE EXCEPTION 'Tenant not found: %', p_tenant_id;
    END IF;

    -- Check if tenant is soft-deleted
    IF v_tenant_record.deleted_at IS NOT NULL THEN
        RAISE EXCEPTION 'Tenant is deleted: %', p_tenant_id;
    END IF;

    -- Check tenant status
    IF v_tenant_record.status NOT IN ('active', 'pending') THEN
        RAISE EXCEPTION 'Tenant is not active: % (status: %)', p_tenant_id, v_tenant_record.status;
    END IF;

    -- Update last activity
    UPDATE tenants 
    SET last_activity_at = NOW() 
    WHERE id = p_tenant_id;

    -- Set tenant context
    PERFORM set_config('app.current_tenant_id', p_tenant_id::text, true);

    -- Return tenant information
    RETURN QUERY SELECT v_tenant_record.name, v_tenant_record.status;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

## 🛠️ Go Integration Best Practices

### Context Management in Go

```go
// Context management
type TenantContextKey string
const TenantIDKey TenantContextKey = "tenant_id"

// Context helpers
func WithTenantID(ctx context.Context, tenantID uuid.UUID) context.Context {
    return context.WithValue(ctx, TenantIDKey, tenantID)
}

func GetTenantID(ctx context.Context) (uuid.UUID, bool) {
    tenantID, ok := ctx.Value(TenantIDKey).(uuid.UUID)
    return tenantID, ok
}
```

### Error Handling

```go
// Define specific error types for better error handling
type TenantError struct {
    TenantID uuid.UUID
    Code     string
    Message  string
}

func (e *TenantError) Error() string {
    return fmt.Sprintf("tenant %s: %s - %s", e.TenantID, e.Code, e.Message)
}

// Error constants
const (
    ErrTenantNotFound     = "TENANT_NOT_FOUND"
    ErrTenantSuspended    = "TENANT_SUSPENDED"
    ErrTenantInactive     = "TENANT_INACTIVE"
    ErrTenantLimitReached = "TENANT_LIMIT_REACHED"
)
```

### Connection Pool Optimization

```go
// Optimized pool configuration for multi-tenant systems
func NewDB(databaseURL string) (Store, error) {
    config, err := pgxpool.ParseConfig(databaseURL)
    if err != nil {
        return nil, fmt.Errorf("failed to parse database URL: %w", err)
    }

    // Optimize for multi-tenant workload
    config.MaxConns = 50              // Increase for multi-tenant load
    config.MinConns = 10              // Higher minimum for quick response
    config.MaxConnLifetime = 1 * time.Hour
    config.MaxConnIdleTime = 30 * time.Minute
    config.HealthCheckPeriod = 1 * time.Minute

    // Connection-level settings for RLS
    config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
        // Ensure RLS is enabled for all connections
        _, err := conn.Exec(ctx, "SET row_security = on")
        if err != nil {
            return fmt.Errorf("failed to enable row security: %w", err)
        }
        return nil
    }

    // ... rest of implementation
}
```

## 📊 Usage Tracking and Limits

### Smart Limit Checking

```sql
CREATE OR REPLACE FUNCTION check_tenant_limits(
    p_tenant_id UUID,
    p_check_type VARCHAR(50),
    p_additional_usage BIGINT DEFAULT 1
)
RETURNS TABLE(
    allowed BOOLEAN,
    current_usage BIGINT,
    limit_value BIGINT,
    usage_percentage NUMERIC(5,2)
) AS $$
DECLARE
    v_config tenant_configurations%ROWTYPE;
    v_current_usage BIGINT;
    v_limit BIGINT;
    v_period_start DATE;
BEGIN
    -- Get tenant configuration
    SELECT * INTO v_config
    FROM tenant_configurations
    WHERE tenant_id = p_tenant_id;

    IF v_config.tenant_id IS NULL THEN
        RAISE EXCEPTION 'No configuration found for tenant: %', p_tenant_id;
    END IF;

    -- Calculate current period start
    v_period_start := date_trunc('month', CURRENT_DATE);

    CASE p_check_type
        WHEN 'storage' THEN
            SELECT COALESCE(storage_used, 0) INTO v_current_usage
            FROM tenant_usage_stats
            WHERE tenant_id = p_tenant_id
              AND period_start = v_period_start;
            
            v_limit := v_config.storage_quota;

        WHEN 'users' THEN
            -- This would require a users table with tenant_id
            v_current_usage := 0; -- Placeholder
            v_limit := v_config.max_users;

        WHEN 'transactions' THEN
            SELECT COALESCE(total_transactions, 0) INTO v_current_usage
            FROM tenant_usage_stats
            WHERE tenant_id = p_tenant_id
              AND period_start = v_period_start;
            
            v_limit := v_config.max_transactions_per_month;

        WHEN 'api_calls_minute' THEN
            -- Check last minute's API calls
            SELECT COALESCE(COUNT(*), 0) INTO v_current_usage
            FROM api_requests -- This table would need to be created
            WHERE tenant_id = p_tenant_id
              AND created_at >= NOW() - INTERVAL '1 minute';
            
            v_limit := v_config.api_rate_limit_per_minute;

        WHEN 'api_calls_hour' THEN
            SELECT COALESCE(COUNT(*), 0) INTO v_current_usage
            FROM api_requests
            WHERE tenant_id = p_tenant_id
              AND created_at >= NOW() - INTERVAL '1 hour';
            
            v_limit := v_config.api_rate_limit_per_hour;

        ELSE
            RAISE EXCEPTION 'Unknown limit type: %', p_check_type;
    END CASE;

    -- Return results
    RETURN QUERY SELECT
        (COALESCE(v_current_usage, 0) + p_additional_usage) <= v_limit AS allowed,
        COALESCE(v_current_usage, 0) AS current_usage,
        v_limit AS limit_value,
        CASE 
            WHEN v_limit > 0 THEN 
                ROUND((COALESCE(v_current_usage, 0)::NUMERIC / v_limit::NUMERIC) * 100, 2)
            ELSE 0 
        END AS usage_percentage;
END;
$$ LANGUAGE plpgsql;
```

### Automated Usage Recording

```sql
-- Function to record tenant activity with automatic aggregation
CREATE OR REPLACE FUNCTION record_tenant_activity(
    p_tenant_id UUID,
    p_activity_type VARCHAR(50),
    p_amount BIGINT DEFAULT 1,
    p_metadata JSONB DEFAULT '{}'
)
RETURNS VOID AS $$
DECLARE
    v_period_start DATE;
    v_period_end DATE;
BEGIN
    -- Calculate current period
    v_period_start := date_trunc('month', CURRENT_DATE)::DATE;
    v_period_end := (date_trunc('month', CURRENT_DATE) + INTERVAL '1 month - 1 day')::DATE;

    -- Insert or update usage statistics
    INSERT INTO tenant_usage_stats (
        tenant_id, 
        period_start, 
        period_end,
        api_calls,
        total_transactions,
        storage_used,
        entities_created,
        entities_updated,
        entities_deleted
    ) VALUES (
        p_tenant_id, 
        v_period_start, 
        v_period_end,
        CASE WHEN p_activity_type = 'api_call' THEN p_amount ELSE 0 END,
        CASE WHEN p_activity_type = 'transaction' THEN p_amount ELSE 0 END,
        CASE WHEN p_activity_type = 'storage' THEN p_amount ELSE 0 END,
        CASE WHEN p_activity_type = 'entity_created' THEN p_amount ELSE 0 END,
        CASE WHEN p_activity_type = 'entity_updated' THEN p_amount ELSE 0 END,
        CASE WHEN p_activity_type = 'entity_deleted' THEN p_amount ELSE 0 END
    )
    ON CONFLICT (tenant_id, period_start) DO UPDATE SET
        api_calls = tenant_usage_stats.api_calls + 
            CASE WHEN p_activity_type = 'api_call' THEN p_amount ELSE 0 END,
        total_transactions = tenant_usage_stats.total_transactions + 
            CASE WHEN p_activity_type = 'transaction' THEN p_amount ELSE 0 END,
        storage_used = tenant_usage_stats.storage_used + 
            CASE WHEN p_activity_type = 'storage' THEN p_amount ELSE 0 END,
        entities_created = tenant_usage_stats.entities_created + 
            CASE WHEN p_activity_type = 'entity_created' THEN p_amount ELSE 0 END,
        entities_updated = tenant_usage_stats.entities_updated + 
            CASE WHEN p_activity_type = 'entity_updated' THEN p_amount ELSE 0 END,
        entities_deleted = tenant_usage_stats.entities_deleted + 
            CASE WHEN p_activity_type = 'entity_deleted' THEN p_amount ELSE 0 END;

    -- Update tenant last activity
    UPDATE tenants SET last_activity_at = NOW() WHERE id = p_tenant_id;
END;
$$ LANGUAGE plpgsql;
```

## 🚀 Tenant Provisioning

### Complete Provisioning Workflow

```sql
-- Tenant provisioning function
CREATE OR REPLACE FUNCTION provision_tenant_complete(
    p_name VARCHAR(255),
    p_email VARCHAR(255),
    p_subdomain VARCHAR(63) DEFAULT NULL,
    p_industry VARCHAR(50) DEFAULT NULL,
    p_company_size VARCHAR(20) DEFAULT 'small',
    p_currency_code CHAR(3) DEFAULT 'USD',
    p_timezone VARCHAR(50) DEFAULT 'UTC',
    p_settings JSONB DEFAULT '{}'
)
RETURNS TABLE(
    tenant_id UUID,
    tenant_slug VARCHAR(50),
    subdomain VARCHAR(63),
    status VARCHAR(20)
) AS $$
DECLARE
    v_tenant_id UUID;
    v_slug VARCHAR(50);
BEGIN
    -- Generate UUID and slug
    v_tenant_id := uuid_generate_v4();
    v_slug := lower(regexp_replace(p_name, '[^a-zA-Z0-9]+', '-', 'g'));
    
    -- Ensure slug uniqueness
    WHILE EXISTS (SELECT 1 FROM tenants WHERE slug = v_slug AND deleted_at IS NULL) LOOP
        v_slug := v_slug || '-' || substring(v_tenant_id::text, 1, 8);
    END LOOP;

    -- Create tenant record
    INSERT INTO tenants (
        id, slug, name, email, subdomain, status, industry, 
        company_size, currency_code, timezone, settings
    ) VALUES (
        v_tenant_id, v_slug, p_name, p_email, p_subdomain, 'pending',
        p_industry, p_company_size, p_currency_code, p_timezone, p_settings
    );

    -- Create default configuration
    INSERT INTO tenant_configurations (
        tenant_id, default_currency
    ) VALUES (
        v_tenant_id, p_currency_code
    );

    -- Initialize usage statistics for current month
    INSERT INTO tenant_usage_stats (tenant_id, period_start, period_end)
    VALUES (
        v_tenant_id,
        date_trunc('month', CURRENT_DATE)::DATE,
        (date_trunc('month', CURRENT_DATE) + INTERVAL '1 month - 1 day')::DATE
    );

    -- Return tenant information
    RETURN QUERY SELECT v_tenant_id, v_slug, p_subdomain, 'pending'::VARCHAR(20);
END;
$$ LANGUAGE plpgsql;
```

### Tenant Onboarding API

```typescript
interface TenantProvisioningRequest {
  company_name: string;
  admin_email: string;
  admin_first_name: string;
  admin_last_name: string;
  country_code: string;
  industry: string;
  company_size: number;
  subscription_plan: string;
  
  customization?: {
    custom_domain?: string;
    branding?: BrandingConfig;
  };
}

// POST /api/v1/tenants/provision
async function provisionTenant(request: TenantProvisioningRequest): Promise<TenantProvisioningResponse> {
  // Step 1: Validate request
  await validateProvisioningRequest(request);
  
  // Step 2: Create tenant in transaction
  const tenant = await database.transaction(async (tx) => {
    const tenant = await createTenant(tx, request);
    await setupTenantSchema(tx, tenant.id);
    await createAdminUser(tx, tenant.id, request);
    return tenant;
  });
  
  // Step 3: Initialize tenant data
  await initializeTenantData(tenant.id, request);
  
  // Step 4: Send welcome email
  await sendTenantWelcomeEmail(tenant, request.admin_email);
  
  return {
    tenant_id: tenant.id,
    tenant_slug: tenant.slug,
    admin_setup_url: `https://${tenant.slug}.awo.com/setup`,
    status: 'provisioned'
  };
}
```

## 🔄 Performance Optimization

### Indexing Strategy

```sql
-- Partial indexes for active tenants only
CREATE INDEX idx_tenants_active_status ON tenants(status) 
    WHERE deleted_at IS NULL AND status = 'active';

CREATE INDEX idx_tenants_activity_recent ON tenants(last_activity_at DESC) 
    WHERE deleted_at IS NULL AND last_activity_at > NOW() - INTERVAL '30 days';

-- Composite indexes for common query patterns
CREATE INDEX idx_usage_stats_tenant_activity ON tenant_usage_stats(
    tenant_id, period_start DESC, total_transactions DESC
);

-- Expression indexes for JSON queries
CREATE INDEX idx_tenants_settings_branding ON tenants 
    USING GIN ((settings->'branding'));
```

### Query Optimization Patterns

```sql
-- Materialized view for tenant health dashboard
CREATE MATERIALIZED VIEW tenant_health_summary AS
SELECT 
    t.id as tenant_id,
    t.name,
    t.status,
    t.last_activity_at,
    tc.storage_quota,
    COALESCE(tus.storage_used, 0) as storage_used,
    CASE 
        WHEN COALESCE(tus.storage_used, 0)::FLOAT / tc.storage_quota > 0.9 THEN 'critical'
        WHEN COALESCE(tus.storage_used, 0)::FLOAT / tc.storage_quota > 0.75 THEN 'warning'
        ELSE 'healthy'
    END as storage_status,
    COALESCE(tus.error_rate, 0) as error_rate,
    COALESCE(tus.avg_response_time, 0) as avg_response_time
FROM tenants t
JOIN tenant_configurations tc ON t.id = tc.tenant_id
LEFT JOIN tenant_usage_stats tus ON t.id = tus.tenant_id 
    AND tus.period_start = date_trunc('month', CURRENT_DATE)::DATE
WHERE t.deleted_at IS NULL
ORDER BY t.name;

-- Index for the materialized view
CREATE INDEX idx_tenant_health_summary_status ON tenant_health_summary(storage_status);
CREATE INDEX idx_tenant_health_summary_activity ON tenant_health_summary(last_activity_at DESC);

-- Refresh schedule (would be handled by a job scheduler)
-- REFRESH MATERIALIZED VIEW CONCURRENTLY tenant_health_summary;
```

## 🛡️ Rate Limiting Implementation

### API Request Tracking

```sql
-- API request tracking table
CREATE TABLE api_requests (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    endpoint VARCHAR(255) NOT NULL,
    method VARCHAR(10) NOT NULL,
    status_code INT NOT NULL,
    response_time_ms INT,
    request_size_bytes INT,
    response_size_bytes INT,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Partitioning by month for better performance
-- ALTER TABLE api_requests PARTITION BY RANGE (created_at);

-- RLS policy
ALTER TABLE api_requests ENABLE ROW LEVEL SECURITY;
CREATE POLICY api_requests_policy ON api_requests
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id());

-- Indexes for rate limiting queries
CREATE INDEX idx_api_requests_tenant_time ON api_requests(tenant_id, created_at DESC);
CREATE INDEX idx_api_requests_minute_window ON api_requests(tenant_id, created_at) 
    WHERE created_at > NOW() - INTERVAL '1 minute';
CREATE INDEX idx_api_requests_hour_window ON api_requests(tenant_id, created_at) 
    WHERE created_at > NOW() - INTERVAL '1 hour';

-- Function to check rate limits
CREATE OR REPLACE FUNCTION check_api_rate_limit(
    p_tenant_id UUID,
    p_window VARCHAR(10) DEFAULT 'minute'
)
RETURNS TABLE(
    allowed BOOLEAN,
    current_count BIGINT,
    limit_value BIGINT,
    reset_time TIMESTAMPTZ
) AS $
DECLARE
    v_config tenant_configurations%ROWTYPE;
    v_current_count BIGINT;
    v_limit BIGINT;
    v_window_start TIMESTAMPTZ;
    v_reset_time TIMESTAMPTZ;
BEGIN
    -- Get tenant configuration
    SELECT * INTO v_config FROM tenant_configurations WHERE tenant_id = p_tenant_id;
    
    IF v_config.tenant_id IS NULL THEN
        RAISE EXCEPTION 'No configuration found for tenant: %', p_tenant_id;
    END IF;

    -- Set window parameters
    CASE p_window
        WHEN 'minute' THEN
            v_window_start := date_trunc('minute', NOW());
            v_reset_time := v_window_start + INTERVAL '1 minute';
            v_limit := v_config.api_rate_limit_per_minute;
        WHEN 'hour' THEN
            v_window_start := date_trunc('hour', NOW());
            v_reset_time := v_window_start + INTERVAL '1 hour';
            v_limit := v_config.api_rate_limit_per_hour;
        ELSE
            RAISE EXCEPTION 'Invalid window type: %', p_window;
    END CASE;

    -- Count requests in current window
    SELECT COUNT(*) INTO v_current_count
    FROM api_requests
    WHERE tenant_id = p_tenant_id
      AND created_at >= v_window_start;

    -- Return rate limit status
    RETURN QUERY SELECT
        v_current_count < v_limit AS allowed,
        v_current_count AS current_count,
        v_limit AS limit_value,
        v_reset_time AS reset_time;
END;
$ LANGUAGE plpgsql;
```

## 🏥 Health Monitoring & Alerting

### Health Check System

```sql
-- Tenant health monitoring function
CREATE OR REPLACE FUNCTION get_tenant_health_status(p_tenant_id UUID)
RETURNS TABLE(
    tenant_id UUID,
    overall_status VARCHAR(20),
    storage_status VARCHAR(20),
    performance_status VARCHAR(20),
    error_status VARCHAR(20),
    activity_status VARCHAR(20),
    alerts JSONB,
    warnings JSONB,
    metrics JSONB,
    checked_at TIMESTAMPTZ
) AS $
DECLARE
    v_config tenant_configurations%ROWTYPE;
    v_usage tenant_usage_stats%ROWTYPE;
    v_tenant tenants%ROWTYPE;
    v_storage_pct NUMERIC;
    v_alerts JSONB := '[]';
    v_warnings JSONB := '[]';
    v_overall_status VARCHAR(20) := 'healthy';
    v_storage_status VARCHAR(20) := 'healthy';
    v_performance_status VARCHAR(20) := 'healthy';
    v_error_status VARCHAR(20) := 'healthy';
    v_activity_status VARCHAR(20) := 'healthy';
BEGIN
    -- Get tenant information
    SELECT * INTO v_tenant FROM tenants WHERE id = p_tenant_id;
    SELECT * INTO v_config FROM tenant_configurations WHERE tenant_id = p_tenant_id;
    
    -- Get current month's usage statistics
    SELECT * INTO v_usage 
    FROM tenant_usage_stats 
    WHERE tenant_id = p_tenant_id 
      AND period_start = date_trunc('month', CURRENT_DATE)::DATE;

    -- Check storage usage
    IF v_usage.storage_used IS NOT NULL AND v_config.storage_quota > 0 THEN
        v_storage_pct := (v_usage.storage_used::NUMERIC / v_config.storage_quota) * 100;
        
        IF v_storage_pct > 95 THEN
            v_storage_status := 'critical';
            v_overall_status := 'critical';
            v_alerts := v_alerts || jsonb_build_object(
                'type', 'storage_critical',
                'message', format('Storage usage at %.1f%% (%.2f GB used of %.2f GB)', 
                    v_storage_pct, 
                    v_usage.storage_used::NUMERIC / 1073741824,
                    v_config.storage_quota::NUMERIC / 1073741824)
            );
        ELSIF v_storage_pct > 80 THEN
            v_storage_status := 'warning';
            IF v_overall_status = 'healthy' THEN v_overall_status := 'warning'; END IF;
            v_warnings := v_warnings || jsonb_build_object(
                'type', 'storage_warning',
                'message', format('Storage usage at %.1f%%', v_storage_pct)
            );
        END IF;
    END IF;

    -- Check error rate
    IF v_usage.error_rate IS NOT NULL THEN
        IF v_usage.error_rate > 0.1 THEN -- 10% error rate
            v_error_status := 'critical';
            v_overall_status := 'critical';
            v_alerts := v_alerts || jsonb_build_object(
                'type', 'high_error_rate',
                'message', format('Error rate at %.2f%%', v_usage.error_rate * 100)
            );
        ELSIF v_usage.error_rate > 0.05 THEN -- 5% error rate
            v_error_status := 'warning';
            IF v_overall_status = 'healthy' THEN v_overall_status := 'warning'; END IF;
            v_warnings := v_warnings || jsonb_build_object(
                'type', 'elevated_error_rate',
                'message', format('Error rate at %.2f%%', v_usage.error_rate * 100)
            );
        END IF;
    END IF;

    -- Check performance
    IF v_usage.avg_response_time IS NOT NULL THEN
        IF v_usage.avg_response_time > 5000 THEN -- 5 seconds
            v_performance_status := 'critical';
            v_overall_status := 'critical';
            v_alerts := v_alerts || jsonb_build_object(
                'type', 'slow_response_time',
                'message', format('Average response time: %.0fms', v_usage.avg_response_time)
            );
        ELSIF v_usage.avg_response_time > 2000 THEN -- 2 seconds
            v_performance_status := 'warning';
            IF v_overall_status = 'healthy' THEN v_overall_status := 'warning'; END IF;
            v_warnings := v_warnings || jsonb_build_object(
                'type', 'degraded_performance',
                'message', format('Average response time: %.0fms', v_usage.avg_response_time)
            );
        END IF;
    END IF;

    -- Check tenant activity
    IF v_tenant.last_activity_at < NOW() - INTERVAL '7 days' THEN
        v_activity_status := 'warning';
        IF v_overall_status = 'healthy' THEN v_overall_status := 'warning'; END IF;
        v_warnings := v_warnings || jsonb_build_object(
            'type', 'inactive_tenant',
            'message', format('No activity since %s', v_tenant.last_activity_at)
        );
    END IF;

    -- Check if tenant is suspended
    IF v_tenant.status = 'suspended' THEN
        v_overall_status := 'critical';
        v_alerts := v_alerts || jsonb_build_object(
            'type', 'tenant_suspended',
            'message', 'Tenant account is suspended'
        );
    END IF;

    RETURN QUERY SELECT
        p_tenant_id,
        v_overall_status,
        v_storage_status,
        v_performance_status,
        v_error_status,
        v_activity_status,
        v_alerts,
        v_warnings,
        jsonb_build_object(
            'storage_used_bytes', COALESCE(v_usage.storage_used, 0),
            'storage_quota_bytes', v_config.storage_quota,
            'storage_usage_percent', COALESCE(v_storage_pct, 0),
            'error_rate', COALESCE(v_usage.error_rate, 0),
            'avg_response_time_ms', COALESCE(v_usage.avg_response_time, 0),
            'active_users', COALESCE(v_usage.active_users, 0),
            'total_transactions', COALESCE(v_usage.total_transactions, 0),
            'api_calls', COALESCE(v_usage.api_calls, 0),
            'uptime_percentage', COALESCE(v_usage.uptime_percentage, 1.0),
            'last_activity', v_tenant.last_activity_at
        ) as metrics,
        NOW() as checked_at;
END;
$ LANGUAGE plpgsql;
```

### Automated Cleanup and Maintenance

```sql
-- Automated cleanup procedures
CREATE OR REPLACE FUNCTION cleanup_tenant_data()
RETURNS TABLE(
    tenants_cleaned INT,
    old_stats_removed INT,
    old_requests_removed INT
) AS $
DECLARE
    v_tenants_cleaned INT := 0;
    v_old_stats_removed INT := 0;
    v_old_requests_removed INT := 0;
    v_retention_days INT;
BEGIN
    -- Clean up old usage statistics (keep last 24 months)
    DELETE FROM tenant_usage_stats 
    WHERE period_start < CURRENT_DATE - INTERVAL '24 months';
    
    GET DIAGNOSTICS v_old_stats_removed = ROW_COUNT;

    -- Clean up old API request logs (based on tenant configuration)
    FOR v_retention_days IN 
        SELECT COALESCE((settings->>'api_log_retention_days')::INT, 90) 
        FROM tenant_configurations 
    LOOP
        DELETE FROM api_requests 
        WHERE created_at < NOW() - (v_retention_days || ' days')::INTERVAL;
    END LOOP;
    
    GET DIAGNOSTICS v_old_requests_removed = ROW_COUNT;

    -- Archive inactive tenants (no activity for 2+ years)
    UPDATE tenants 
    SET status = 'archived'
    WHERE status = 'active' 
      AND last_activity_at < NOW() - INTERVAL '2 years'
      AND deleted_at IS NULL;
      
    GET DIAGNOSTICS v_tenants_cleaned = ROW_COUNT;

    RETURN QUERY SELECT v_tenants_cleaned, v_old_stats_removed, v_old_requests_removed;
END;
$ LANGUAGE plpgsql;
```

## 📊 Analytics & Reporting

### Multi-Dimensional Analytics

```sql
-- Tenant metrics aggregation
CREATE OR REPLACE FUNCTION get_tenant_analytics(
    p_tenant_id UUID,
    p_period VARCHAR(20) DEFAULT 'month',
    p_start_date DATE DEFAULT NULL,
    p_end_date DATE DEFAULT NULL
)
RETURNS TABLE(
    period_label TEXT,
    active_users INT,
    new_users INT,
    total_transactions INT,
    successful_transactions INT,
    failed_transactions INT,
    success_rate NUMERIC(5,2),
    storage_used BIGINT,
    api_calls INT,
    api_success_rate NUMERIC(5,2),
    avg_response_time NUMERIC(10,3),
    revenue NUMERIC(15,4),
    growth_rate NUMERIC(8,4)
) AS $
DECLARE
    v_date_trunc TEXT;
    v_start_date DATE;
    v_end_date DATE;
BEGIN
    -- Set default date range if not provided
    v_start_date := COALESCE(p_start_date, CURRENT_DATE - INTERVAL '12 months');
    v_end_date := COALESCE(p_end_date, CURRENT_DATE);

    -- Determine date truncation based on period
    CASE p_period
        WHEN 'day' THEN v_date_trunc := 'day';
        WHEN 'week' THEN v_date_trunc := 'week';
        WHEN 'month' THEN v_date_trunc := 'month';
        WHEN 'quarter' THEN v_date_trunc := 'quarter';
        WHEN 'year' THEN v_date_trunc := 'year';
        ELSE v_date_trunc := 'month';
    END CASE;

    RETURN QUERY
    WITH period_stats AS (
        SELECT
            date_trunc(v_date_trunc, period_start::TIMESTAMP)::DATE as period_date,
            SUM(active_users) as period_active_users,
            SUM(new_users_registered) as period_new_users,
            SUM(total_transactions) as period_total_transactions,
            SUM(successful_transactions) as period_successful_transactions,
            SUM(failed_transactions) as period_failed_transactions,
            AVG(storage_used) as period_storage_used,
            SUM(api_calls) as period_api_calls,
            SUM(api_calls_successful) as period_api_successful,
            AVG(avg_response_time) as period_avg_response_time,
            SUM(monthly_revenue) as period_revenue
        FROM tenant_usage_stats
        WHERE tenant_id = p_tenant_id
          AND period_start BETWEEN v_start_date AND v_end_date
        GROUP BY date_trunc(v_date_trunc, period_start::TIMESTAMP)::DATE
    ),
    period_growth AS (
        SELECT 
            *,
            LAG(period_revenue) OVER (ORDER BY period_date) as prev_revenue
        FROM period_stats
    )
    SELECT
        to_char(period_date, 'YYYY-MM-DD') as period_label,
        period_active_users::INT,
        period_new_users::INT,
        period_total_transactions::INT,
        period_successful_transactions::INT,
        period_failed_transactions::INT,
        CASE 
            WHEN period_total_transactions > 0 THEN 
                ROUND((period_successful_transactions::NUMERIC / period_total_transactions) * 100, 2)
            ELSE 0 
        END as success_rate,
        period_storage_used::BIGINT,
        period_api_calls::INT,
        CASE 
            WHEN period_api_calls > 0 THEN 
                ROUND((period_api_successful::NUMERIC / period_api_calls) * 100, 2)
            ELSE 0 
        END as api_success_rate,
        period_avg_response_time,
        period_revenue,
        CASE 
            WHEN prev_revenue > 0 THEN 
                ROUND(((period_revenue - prev_revenue) / prev_revenue) * 100, 4)
            ELSE 0 
        END as growth_rate
    FROM period_growth
    ORDER BY period_date;
END;
$ LANGUAGE plpgsql;
```

### Tenant Benchmarking

```sql
-- Tenant benchmarking function
CREATE OR REPLACE FUNCTION get_tenant_benchmarks(
    p_tenant_id UUID,
    p_comparison_group VARCHAR(50) DEFAULT 'industry' -- 'industry', 'size', 'all'
)
RETURNS TABLE(
    metric_name VARCHAR(50),
    tenant_value NUMERIC,
    benchmark_median NUMERIC,
    benchmark_p75 NUMERIC,
    benchmark_p90 NUMERIC,
    percentile_rank NUMERIC(5,2)
) AS $
BEGIN
    RETURN QUERY
    WITH tenant_metrics AS (
        SELECT 
            t.id as tenant_id,
            t.industry,
            t.company_size,
            COALESCE(tus.avg_response_time, 0) as response_time,
            COALESCE(tus.error_rate, 0) as error_rate,
            COALESCE(tus.storage_used::NUMERIC / tc.storage_quota, 0) as storage_utilization,
            COALESCE(tus.api_calls, 0) as api_usage,
            COALESCE(tus.active_users, 0) as active_users,
            COALESCE(tus.uptime_percentage, 1.0) as uptime
        FROM tenants t
        JOIN tenant_configurations tc ON t.id = tc.tenant_id
        LEFT JOIN tenant_usage_stats tus ON t.id = tus.tenant_id 
            AND tus.period_start = date_trunc('month', CURRENT_DATE)::DATE
        WHERE t.deleted_at IS NULL AND t.status = 'active'
    ),
    filtered_metrics AS (
        SELECT *
        FROM tenant_metrics tm
        WHERE (p_comparison_group = 'all') OR
              (p_comparison_group = 'industry' AND tm.industry = (
                  SELECT industry FROM tenant_metrics WHERE tenant_id = p_tenant_id
              )) OR
              (p_comparison_group = 'size' AND tm.company_size = (
                  SELECT company_size FROM tenant_metrics WHERE tenant_id = p_tenant_id
              ))
    ),
    tenant_data AS (
        SELECT * FROM filtered_metrics WHERE tenant_id = p_tenant_id
    )
    -- Response Time Benchmark
    SELECT 
        'response_time'::VARCHAR(50),
        td.response_time,
        percentile_cont(0.5) WITHIN GROUP (ORDER BY fm.response_time),
        percentile_cont(0.75) WITHIN GROUP (ORDER BY fm.response_time),
        percentile_cont(0.90) WITHIN GROUP (ORDER BY fm.response_time),
        percent_rank() WITHIN GROUP (ORDER BY fm.response_time) * 100
    FROM tenant_data td, filtered_metrics fm
    
    UNION ALL
    
    -- Error Rate Benchmark
    SELECT 
        'error_rate'::VARCHAR(50),
        td.error_rate * 100, -- Convert to percentage
        percentile_cont(0.5) WITHIN GROUP (ORDER BY fm.error_rate) * 100,
        percentile_cont(0.75) WITHIN GROUP (ORDER BY fm.error_rate) * 100,
        percentile_cont(0.90) WITHIN GROUP (ORDER BY fm.error_rate) * 100,
        (1 - percent_rank() WITHIN GROUP (ORDER BY fm.error_rate)) * 100 -- Lower is better
    FROM tenant_data td, filtered_metrics fm
    
    UNION ALL
    
    -- Storage Utilization Benchmark
    SELECT 
        'storage_utilization'::VARCHAR(50),
        td.storage_utilization * 100, -- Convert to percentage
        percentile_cont(0.5) WITHIN GROUP (ORDER BY fm.storage_utilization) * 100,
        percentile_cont(0.75) WITHIN GROUP (ORDER BY fm.storage_utilization) * 100,
        percentile_cont(0.90) WITHIN GROUP (ORDER BY fm.storage_utilization) * 100,
        percent_rank() WITHIN GROUP (ORDER BY fm.storage_utilization) * 100
    FROM tenant_data td, filtered_metrics fm
    
    UNION ALL
    
    -- Uptime Benchmark
    SELECT 
        'uptime'::VARCHAR(50),
        td.uptime * 100, -- Convert to percentage
        percentile_cont(0.5) WITHIN GROUP (ORDER BY fm.uptime) * 100,
        percentile_cont(0.75) WITHIN GROUP (ORDER BY fm.uptime) * 100,
        percentile_cont(0.90) WITHIN GROUP (ORDER BY fm.uptime) * 100,
        percent_rank() WITHIN GROUP (ORDER BY fm.uptime) * 100
    FROM tenant_data td, filtered_metrics fm;
END;
$ LANGUAGE plpgsql;
```

## 💰 Billing & Subscription Management

### Subscription Plans

```yaml
subscription_plans:
  starter:
    name: "Starter"
    price: 29
    billing_cycle: monthly
    limits:
      users: 5
      storage_gb: 10
      api_calls_monthly: 10000
    modules:
      - accounting
      - inventory
      - contacts
  
  professional:
    name: "Professional"
    price: 99
    billing_cycle: monthly
    limits:
      users: 25
      storage_gb: 100
      api_calls_monthly: 100000
    modules:
      - accounting
      - inventory
      - contacts
      - sales
      - projects
  
  enterprise:
    name: "Enterprise"
    price: 299
    billing_cycle: monthly
    limits:
      users: unlimited
      storage_gb: 1000
      api_calls_monthly: 1000000
    modules:
      - all_modules
```

### Usage-Based Billing

```typescript
interface UsageBillingRules {
  base_price: number;
  
  usage_tiers: {
    users: {
      included: number;
      overage_price_per_user: number;
    };
    
    storage: {
      included_gb: number;
      overage_price_per_gb: number;
    };
    
    api_calls: {
      included_monthly: number;
      overage_price_per_1000: number;
    };
    
    transactions: {
      included_monthly: number;
      overage_price_per_transaction: number;
    };
  };
}
```

## 🔧 Go Integration Patterns

### Middleware Stack

```go
// Tenant context middleware
func TenantContextMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract tenant from subdomain, header, or JWT
        tenantID := extractTenantID(c)
        
        if tenantID == "" {
            c.JSON(401, gin.H{"error": "Tenant not found"})
            c.Abort()
            return
        }
        
        // Validate tenant status
        tenant, err := tenantService.GetTenant(tenantID)
        if err != nil || tenant.Status != "active" {
            c.JSON(403, gin.H{"error": "Tenant access denied"})
            c.Abort()
            return
        }
        
        // Set tenant context
        c.Set("tenant_id", tenantID)
        c.Set("tenant", tenant)
        
        // Set database context for RLS
        db := database.GetConnection()
        db.Exec("SET app.current_tenant_id = ?", tenantID)
        
        c.Next()
    }
}

func handleTenantError(c *gin.Context, err error) {
    var tenantErr *TenantError
    if errors.As(err, &tenantErr) {
        switch tenantErr.Code {
        case ErrTenantNotFound:
            c.JSON(http.StatusNotFound, gin.H{
                "error": tenantErr.Message,
                "code":  tenantErr.Code,
            })
        case ErrTenantSuspended, ErrTenantInactive:
            c.JSON(http.StatusForbidden, gin.H{
                "error": tenantErr.Message,
                "code":  tenantErr.Code,
            })
        case ErrTenantLimitReached:
            c.JSON(http.StatusPaymentRequired, gin.H{
                "error":   tenantErr.Message,
                "code":    tenantErr.Code,
            })
        default:
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "Internal tenant error",
                "code":  "INTERNAL_ERROR",
            })
        }
    } else {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Internal server error",
            "code":  "INTERNAL_ERROR",
        })
    }
}
```

### Repository Pattern

```go
// Repository pattern with automatic tenant filtering
type BaseRepository struct {
    db       *gorm.DB
    tenantID string
}

func (r *BaseRepository) Find(dest interface{}, conditions ...interface{}) error {
    // Automatically add tenant filter to all queries
    return r.db.Where("tenant_id = ?", r.tenantID).Find(dest, conditions...).Error
}

func (r *BaseRepository) Create(value interface{}) error {
    // Automatically set tenant_id before creating
    if model, ok := value.(TenantModel); ok {
        model.SetTenantID(r.tenantID)
    }
    return r.db.Create(value).Error
}
```

## 🛠️ Tenant Management API

### Core Operations

```typescript
// Tenant management endpoints
interface TenantManagementAPI {
  // Tenant CRUD operations
  GET    /api/v1/tenants/{tenant_id}
  PUT    /api/v1/tenants/{tenant_id}
  DELETE /api/v1/tenants/{tenant_id}
  
  // // Organization management
  // GET    /api/v1/tenants/{tenant_id}/organizations
  // POST   /api/v1/tenants/{tenant_id}/organizations
  // PUT    /api/v1/tenants/{tenant_id}/organizations/{org_id}
  // DELETE /api/v1/tenants/{tenant_id}/organizations/{org_id}
  //
  // Configuration management
  GET    /api/v1/tenants/{tenant_id}/settings
  PUT    /api/v1/tenants/{tenant_id}/settings
  GET    /api/v1/tenants/{tenant_id}/modules
  PUT    /api/v1/tenants/{tenant_id}/modules
  
  // User management
  GET    /api/v1/tenants/{tenant_id}/users
  POST   /api/v1/tenants/{tenant_id}/users/invite
  PUT    /api/v1/tenants/{tenant_id}/users/{user_id}
  DELETE /api/v1/tenants/{tenant_id}/users/{user_id}
  
  // Analytics and monitoring
  GET    /api/v1/tenants/{tenant_id}/metrics
  GET    /api/v1/tenants/{tenant_id}/usage
  GET    /api/v1/tenants/{tenant_id}/health
}
```

## 📈 Scaling Patterns

### Horizontal Scaling Preparation

```sql
-- Table for future sharding decisions
CREATE TABLE tenant_shards (
    tenant_id UUID NOT NULL REFERENCES tenants(id) PRIMARY KEY,
    shard_key VARCHAR(50) NOT NULL,
    shard_database VARCHAR(100),
    migrated_at TIMESTAMPTZ,
    migration_status VARCHAR(20) DEFAULT 'pending' 
        CHECK (migration_status IN ('pending', 'in_progress', 'completed', 'failed'))
);

-- Function to determine shard assignment
CREATE OR REPLACE FUNCTION assign_tenant_shard(p_tenant_id UUID)
RETURNS VARCHAR(50) AS $
DECLARE
    v_shard_key VARCHAR(50);
BEGIN
    -- Simple hash-based sharding (can be enhanced)
    v_shard_key := 'shard_' || (hashtext(p_tenant_id::TEXT) % 16 + 1);
    
    -- Could implement more sophisticated logic based on:
    -- - Tenant size
    -- - Geographic location
    -- - Usage patterns
    -- - Resource requirements
    
    INSERT INTO tenant_shards (tenant_id, shard_key)
    VALUES (p_tenant_id, v_shard_key)
    ON CONFLICT (tenant_id) DO NOTHING;
    
    RETURN v_shard_key;
END;
$ LANGUAGE plpgsql;
```

### Read Replica Configuration

```sql
-- Read preference configuration per tenant
CREATE TABLE tenant_read_preferences (
    tenant_id UUID NOT NULL REFERENCES tenants(id) PRIMARY KEY,
    read_preference VARCHAR(20) NOT NULL DEFAULT 'primary'
        CHECK (read_preference IN ('primary', 'secondary', 'nearest')),
    max_staleness_seconds INT DEFAULT 30,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Function to get optimal read connection for tenant
CREATE OR REPLACE FUNCTION get_tenant_read_preference(p_tenant_id UUID)
RETURNS TABLE(
    read_preference VARCHAR(20),
    max_staleness_seconds INT
) AS $
BEGIN
    RETURN QUERY
    SELECT 
        trp.read_preference,
        trp.max_staleness_seconds
    FROM tenant_read_preferences trp
    WHERE trp.tenant_id = p_tenant_id
    
    UNION ALL
    
    -- Default values if no preference set
    SELECT 'primary'::VARCHAR(20), 30::INT
    WHERE NOT EXISTS (
        SELECT 1 FROM tenant_read_preferences WHERE tenant_id = p_tenant_id
    )
    
    LIMIT 1;
END;
$ LANGUAGE plpgsql;
```

## 📋 Recommended sqlc Queries

```sql
-- name: GetTenantBySlug :one
SELECT * FROM tenants 
WHERE slug = $1 AND deleted_at IS NULL AND status = 'active';

-- name: GetTenantConfiguration :one
SELECT * FROM tenant_configurations WHERE tenant_id = $1;

-- name: CheckTenantLimits :one
SELECT * FROM check_tenant_limits($1, $2, $3);

-- name: RecordTenantActivity :exec
SELECT record_tenant_activity($1, $2, $3, $4);

-- name: GetTenantHealthStatus :one
SELECT * FROM get_tenant_health_status($1);

-- name: GetTenantAnalytics :many
SELECT * FROM get_tenant_analytics($1, $2, $3, $4);

-- name: UpdateTenantLastActivity :exec
UPDATE tenants SET last_activity_at = NOW() WHERE id = $1;

-- name: GetTenantsRequiringCleanup :many
SELECT id, name, last_activity_at 
FROM tenants 
WHERE status = 'active' 
  AND last_activity_at < NOW() - INTERVAL '90 days'
  AND deleted_at IS NULL;

-- name: GetTenantUsageStats :one
SELECT * FROM tenant_usage_stats 
WHERE tenant_id = $1 
  AND period_start = date_trunc('month', CURRENT_DATE)::DATE;

-- name: ListTenantsWithHealth :many
SELECT 
    t.*,
    th.overall_status,
    th.storage_status,
    th.alerts,
    th.warnings
FROM tenants t
LEFT JOIN get_tenant_health_status(t.id) th ON true
WHERE t.deleted_at IS NULL
ORDER BY t.name;
```

## 🔄 Data Migration & Backup

### Data Migration

```typescript
interface TenantMigrationPlan {
  source_tenant_id: string;
  target_tenant_id: string;
  migration_type: 'full' | 'selective' | 'merge';
  
  data_selection: {
    include_users: boolean;
    include_transactions: boolean;
    include_historical_data: boolean;
    date_range?: DateRange;
    entity_filters?: EntityFilter[];
  };
  
  mapping_rules: {
    user_mapping: UserMapping[];
    organization_mapping: OrganizationMapping[];
    account_mapping: AccountMapping[];
  };
  
  validation_rules: ValidationRule[];
  rollback_plan: RollbackPlan;
}
```

### Backup Strategy

```yaml
backup_configuration:
  database_backups:
    frequency: daily
    retention: 90_days
    compression: true
    encryption: true
    
  file_backups:
    frequency: daily
    retention: 30_days
    incremental: true
    
  point_in_time_recovery:
    enabled: true
    retention: 7_days
    granularity: 15_minutes
    
  cross_region_replication:
    enabled: true
    regions: [us-east-1, eu-west-1]
    sync_frequency: hourly
```

## 🔄 Database Schema Evolution

### Tenant-Safe Migration Strategy

```sql
-- Migration tracking table
CREATE TABLE schema_migrations (
    version BIGINT NOT NULL PRIMARY KEY,
    filename VARCHAR(255) NOT NULL,
    checksum VARCHAR(64),
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    execution_time_ms INT,
    success BOOLEAN NOT NULL DEFAULT true,
    error_message TEXT
);

-- Tenant-specific migration tracking
CREATE TABLE tenant_schema_versions (
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    schema_version BIGINT NOT NULL,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    migration_data JSONB DEFAULT '{}',
    PRIMARY KEY (tenant_id, schema_version)
);

-- Function to safely execute tenant-aware migrations
CREATE OR REPLACE FUNCTION execute_tenant_migration(
    p_tenant_id UUID,
    p_migration_version BIGINT,
    p_migration_sql TEXT
)
RETURNS BOOLEAN AS $
DECLARE
    v_start_time TIMESTAMPTZ;
    v_execution_time INT;
    v_error_message TEXT;
BEGIN
    -- Check if migration already applied for this tenant
    IF EXISTS (
        SELECT 1 FROM tenant_schema_versions 
        WHERE tenant_id = p_tenant_id AND schema_version = p_migration_version
    ) THEN
        RAISE NOTICE 'Migration % already applied for tenant %', p_migration_version, p_tenant_id;
        RETURN true;
    END IF;

    v_start_time := NOW();
    
    BEGIN
        -- Set tenant context for migration
        PERFORM set_tenant_context(p_tenant_id);
        
        -- Execute migration SQL
        EXECUTE p_migration_sql;
        
        -- Record successful migration
        v_execution_time := EXTRACT(epoch FROM (NOW() - v_start_time)) * 1000;
        
        INSERT INTO tenant_schema_versions (tenant_id, schema_version, applied_at)
        VALUES (p_tenant_id, p_migration_version, NOW());
        
        RETURN true;
        
    EXCEPTION WHEN OTHERS THEN
        v_error_message := SQLERRM;
        v_execution_time := EXTRACT(epoch FROM (NOW() - v_start_time)) * 1000;
        
        -- Log migration failure (but don't insert into tenant_schema_versions)
        RAISE WARNING 'Migration % failed for tenant %: %', p_migration_version, p_tenant_id, v_error_message;
        
        RETURN false;
    END;
END;
$ LANGUAGE plpgsql;
```

### Rolling Migrations for All Tenants

```sql
-- Function to execute migration across all active tenants
CREATE OR REPLACE FUNCTION execute_migration_all_tenants(
    p_migration_version BIGINT,
    p_migration_sql TEXT,
    p_batch_size INT DEFAULT 10
)
RETURNS TABLE(
    tenant_id UUID,
    tenant_name TEXT,
    success BOOLEAN,
    execution_time_ms INT,
    error_message TEXT
) AS $
DECLARE
    v_tenant RECORD;
    v_batch_count INT := 0;
    v_start_time TIMESTAMPTZ;
    v_execution_time INT;
    v_success BOOLEAN;
    v_error_message TEXT;
BEGIN
    -- Process tenants in batches
    FOR v_tenant IN 
        SELECT id, name FROM tenants 
        WHERE status = 'active' AND deleted_at IS NULL
        ORDER BY created_at
    LOOP
        v_start_time := NOW();
        v_error_message := NULL;
        
        BEGIN
            -- Execute migration for this tenant
            v_success := execute_tenant_migration(
                v_tenant.id, 
                p_migration_version, 
                p_migration_sql
            );
            
            v_execution_time := EXTRACT(epoch FROM (NOW() - v_start_time)) * 1000;
            
        EXCEPTION WHEN OTHERS THEN
            v_success := false;
            v_error_message := SQLERRM;
            v_execution_time := EXTRACT(epoch FROM (NOW() - v_start_time)) * 1000;
        END;
        
        -- Return result for this tenant
        RETURN QUERY SELECT 
            v_tenant.id,
            v_tenant.name,
            v_success,
            v_execution_time,
            v_error_message;
        
        v_batch_count := v_batch_count + 1;
        
        -- Pause between batches to avoid overwhelming the system
        IF v_batch_count >= p_batch_size THEN
            PERFORM pg_sleep(1); -- 1 second pause
            v_batch_count := 0;
        END IF;
    END LOOP;
END;
$ LANGUAGE plpgsql;
```

## 🏗️ Data Archiving System

### Automated Data Lifecycle Management

```sql
-- Data retention policies per tenant
CREATE TABLE tenant_data_retention (
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    table_name VARCHAR(100) NOT NULL,
    retention_days INT NOT NULL,
    archive_enabled BOOLEAN NOT NULL DEFAULT true,
    deletion_enabled BOOLEAN NOT NULL DEFAULT false,
    last_cleanup_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, table_name)
);

-- Generic archival function
CREATE OR REPLACE FUNCTION archive_tenant_data(
    p_tenant_id UUID,
    p_table_name VARCHAR(100),
    p_retention_days INT,
    p_batch_size INT DEFAULT 1000
)
RETURNS TABLE(
    archived_count INT,
    deleted_count INT
) AS $
DECLARE
    v_archive_table VARCHAR(100);
    v_cutoff_date DATE;
    v_archived INT := 0;
    v_deleted INT := 0;
    v_sql TEXT;
BEGIN
    v_archive_table := p_table_name || '_archive';
    v_cutoff_date := CURRENT_DATE - p_retention_days;
    
    -- Set tenant context
    PERFORM set_tenant_context(p_tenant_id);
    
    -- Create archive table if it doesn't exist
    v_sql := format('
        CREATE TABLE IF NOT EXISTS %I (
            LIKE %I INCLUDING ALL,
            archived_at TIMESTAMPTZ DEFAULT NOW()
        )', v_archive_table, p_table_name);
    
    EXECUTE v_sql;
    
    -- Move old records to archive table
    v_sql := format('
        WITH moved_data AS (
            DELETE FROM %I 
            WHERE created_at < $1
            RETURNING *
        )
        INSERT INTO %I 
        SELECT *, NOW() as archived_at FROM moved_data',
        p_table_name, v_archive_table);
    
    EXECUTE v_sql USING v_cutoff_date;
    GET DIAGNOSTICS v_archived = ROW_COUNT;
    
    -- Update retention tracking
    INSERT INTO tenant_data_retention (
        tenant_id, table_name, retention_days, last_cleanup_at
    ) VALUES (
        p_tenant_id, p_table_name, p_retention_days, NOW()
    ) ON CONFLICT (tenant_id, table_name) DO UPDATE SET
        last_cleanup_at = NOW();
    
    RETURN QUERY SELECT v_archived, v_deleted;
END;
$ LANGUAGE plpgsql;
```

## 🔍 Predictive Analytics

### Tenant Health Predictions

```sql
-- Tenant health predictions based on historical data
CREATE OR REPLACE FUNCTION predict_tenant_health(
    p_tenant_id UUID,
    p_days_ahead INT DEFAULT 30
)
RETURNS TABLE(
    prediction_date DATE,
    predicted_storage_usage BIGINT,
    predicted_transaction_volume INT,
    risk_score NUMERIC(3,2),
    risk_factors JSONB
) AS $
DECLARE
    v_historical_days INT := 90;
    v_current_usage BIGINT;
    v_storage_quota BIGINT;
BEGIN
    -- Get current metrics and quotas
    SELECT 
        COALESCE(tus.storage_used, 0),
        tc.storage_quota
    INTO v_current_usage, v_storage_quota
    FROM tenant_configurations tc
    LEFT JOIN tenant_usage_stats tus ON tc.tenant_id = tus.tenant_id 
        AND tus.period_start = date_trunc('month', CURRENT_DATE)::DATE
    WHERE tc.tenant_id = p_tenant_id;
    
    RETURN QUERY
    WITH historical_data AS (
        SELECT 
            period_start,
            storage_used,
            total_transactions,
            LAG(storage_used) OVER (ORDER BY period_start) as prev_storage,
            LAG(total_transactions) OVER (ORDER BY period_start) as prev_transactions
        FROM tenant_usage_stats
        WHERE tenant_id = p_tenant_id
          AND period_start >= CURRENT_DATE - (v_historical_days || ' days')::INTERVAL
        ORDER BY period_start
    ),
    growth_analysis AS (
        SELECT 
            AVG(CASE 
                WHEN prev_storage > 0 THEN 
                    (storage_used - prev_storage)::NUMERIC / prev_storage 
                ELSE 0 
            END) as avg_storage_growth_rate,
            AVG(CASE 
                WHEN prev_transactions > 0 THEN 
                    (total_transactions - prev_transactions)::NUMERIC / prev_transactions 
                ELSE 0 
            END) as avg_transaction_growth_rate
        FROM historical_data
        WHERE prev_storage IS NOT NULL
    )
    SELECT 
        (CURRENT_DATE + (generate_series * '1 day'::INTERVAL))::DATE,
        GREATEST(0, v_current_usage + 
            (v_current_usage * ga.avg_storage_growth_rate * generate_series / 30)::BIGINT),
        GREATEST(0, COALESCE(
            (SELECT total_transactions FROM tenant_usage_stats 
             WHERE tenant_id = p_tenant_id 
             ORDER BY period_start DESC LIMIT 1), 0) +
            (COALESCE(
                (SELECT total_transactions FROM tenant_usage_stats 
                 WHERE tenant_id = p_tenant_id 
                 ORDER BY period_start DESC LIMIT 1), 0) * 
             ga.avg_transaction_growth_rate * generate_series / 30)::INT),
        LEAST(1.0, GREATEST(0.0, 
            (v_current_usage + 
             (v_current_usage * ga.avg_storage_growth_rate * generate_series / 30)::BIGINT)::NUMERIC 
            / v_storage_quota)),
        jsonb_build_object(
            'storage_trend', CASE 
                WHEN ga.avg_storage_growth_rate > 0.1 THEN 'high_growth'
                WHEN ga.avg_storage_growth_rate > 0.05 THEN 'moderate_growth'
                ELSE 'stable'
            END,
            'transaction_trend', CASE 
                WHEN ga.avg_transaction_growth_rate > 0.2 THEN 'high_growth'
                WHEN ga.avg_transaction_growth_rate > 0.1 THEN 'moderate_growth'
                ELSE 'stable'
            END
        )
    FROM generate_series(1, p_days_ahead) generate_series, growth_analysis ga;
END;
$ LANGUAGE plpgsql;
```

### Tenant Churn Prediction

```sql
-- Churn risk analysis
CREATE OR REPLACE FUNCTION analyze_tenant_churn_risk(p_tenant_id UUID)
RETURNS TABLE(
    churn_risk_score NUMERIC(3,2),
    risk_level VARCHAR(10),
    risk_factors JSONB,
    recommendation TEXT
) AS $
DECLARE
    v_tenant RECORD;
    v_usage RECORD;
    v_score NUMERIC := 0;
    v_factors JSONB := '{}';
    v_risk_level VARCHAR(10);
    v_recommendation TEXT;
BEGIN
    -- Get tenant information
    SELECT 
        t.*,
        EXTRACT(days FROM NOW() - t.last_activity_at) as days_inactive,
        EXTRACT(days FROM NOW() - t.created_at) as tenant_age_days
    INTO v_tenant
    FROM tenants t
    WHERE t.id = p_tenant_id;
    
    -- Get recent usage statistics
    SELECT * INTO v_usage
    FROM tenant_usage_stats
    WHERE tenant_id = p_tenant_id
      AND period_start = date_trunc('month', CURRENT_DATE)::DATE;
    
    -- Calculate churn risk factors
    
    -- Factor 1: Inactivity (0-40 points)
    IF v_tenant.days_inactive > 30 THEN
        v_score := v_score + LEAST(40, v_tenant.days_inactive);
        v_factors := v_factors || jsonb_build_object('inactivity_days', v_tenant.days_inactive);
    END IF;
    
    -- Factor 2: Low engagement (0-20 points)
    IF COALESCE(v_usage.active_users, 0) = 0 THEN
        v_score := v_score + 20;
        v_factors := v_factors || jsonb_build_object('no_active_users', true);
    ELSIF COALESCE(v_usage.active_users, 0) = 1 THEN
        v_score := v_score + 10;
        v_factors := v_factors || jsonb_build_object('single_user', true);
    END IF;
    
    -- Factor 3: Low transaction volume (0-15 points)
    IF COALESCE(v_usage.total_transactions, 0) < 10 THEN
        v_score := v_score + 15;
        v_factors := v_factors || jsonb_build_object('low_transaction_volume', true);
    END IF;
    
    -- Factor 4: New tenant with no adoption (0-25 points)
    IF v_tenant.tenant_age_days < 30 AND COALESCE(v_usage.total_transactions, 0) = 0 THEN
        v_score := v_score + 25;
        v_factors := v_factors || jsonb_build_object('new_tenant_no_adoption', true);
    END IF;
    
    -- Normalize score to 0-1 scale
    v_score := LEAST(1.0, v_score / 100.0);
    
    -- Determine risk level and recommendation
    CASE 
        WHEN v_score >= 0.7 THEN
            v_risk_level := 'high';
            v_recommendation := 'Immediate outreach required. Consider offering onboarding assistance or promotional pricing.';
        WHEN v_score >= 0.4 THEN
            v_risk_level := 'medium';
            v_recommendation := 'Monitor closely. Send engagement emails and check for support needs.';
        WHEN v_score >= 0.2 THEN
            v_risk_level := 'low';
            v_recommendation := 'Tenant appears healthy. Continue regular check-ins.';
        ELSE
            v_risk_level := 'minimal';
            v_recommendation := 'Tenant is actively engaged. Focus on expansion opportunities.';
    END CASE;
    
    RETURN QUERY SELECT v_score, v_risk_level, v_factors, v_recommendation;
END;
$ LANGUAGE plpgsql;
```

## 🛠️ Go Implementation Extensions

### Store Interface Extensions

```go
// Extended Store interface recommendations
type Store interface {
    Querier
    
    // Existing methods...
    SetTenantContext(ctx context.Context, tenantID uuid.UUID) error
    WithTenant(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, Store) error) error
    
    // New methods
    ExecuteTenantMigration(ctx context.Context, tenantID uuid.UUID, version int64, sql string) error
    GetTenantHealthPredictions(ctx context.Context, tenantID uuid.UUID, daysAhead int) (*TenantHealthPredictions, error)
    AnalyzeChurnRisk(ctx context.Context, tenantID uuid.UUID) (*ChurnRiskAnalysis, error)
    ArchiveTenantData(ctx context.Context, tenantID uuid.UUID, tableName string, retentionDays int) (*ArchivalResult, error)
    
    // Batch operations
    ExecuteMigrationAllTenants(ctx context.Context, version int64, sql string, batchSize int) ([]TenantMigrationResult, error)
    BulkHealthCheck(ctx context.Context, tenantIDs []uuid.UUID) ([]TenantHealthStatus, error)
    
    // Performance monitoring
    RecordAPIRequest(ctx context.Context, req *APIRequestLog) error
    CheckRateLimit(ctx context.Context, tenantID uuid.UUID, window string) (*RateLimitStatus, error)
}

// Supporting types
type TenantHealthPredictions struct {
    TenantID            uuid.UUID                `json:"tenant_id"`
    Predictions         []HealthPrediction       `json:"predictions"`
    GeneratedAt         time.Time                `json:"generated_at"`
}

type HealthPrediction struct {
    Date                     time.Time `json:"date"`
    PredictedStorageUsage    int64     `json:"predicted_storage_usage"`
    PredictedTransactionVol  int32     `json:"predicted_transaction_volume"`
    RiskScore               float64   `json:"risk_score"`
    RiskFactors             map[string]interface{} `json:"risk_factors"`
}

type ChurnRiskAnalysis struct {
    TenantID        uuid.UUID              `json:"tenant_id"`
    ChurnRiskScore  float64                `json:"churn_risk_score"`
    RiskLevel       string                 `json:"risk_level"`
    RiskFactors     map[string]interface{} `json:"risk_factors"`
    Recommendation  string                 `json:"recommendation"`
    AnalyzedAt      time.Time              `json:"analyzed_at"`
}

type ArchivalResult struct {
    TenantID      uuid.UUID `json:"tenant_id"`
    TableName     string    `json:"table_name"`
    ArchivedCount int       `json:"archived_count"`
    DeletedCount  int       `json:"deleted_count"`
    ArchivedAt    time.Time `json:"archived_at"`
}

type TenantMigrationResult struct {
    TenantID        uuid.UUID `json:"tenant_id"`
    TenantName      string    `json:"tenant_name"`
    Success         bool      `json:"success"`
    ExecutionTimeMs int       `json:"execution_time_ms"`
    ErrorMessage    string    `json:"error_message,omitempty"`
}

type APIRequestLog struct {
    TenantID         uuid.UUID `json:"tenant_id"`
    Endpoint         string    `json:"endpoint"`
    Method           string    `json:"method"`
    StatusCode       int       `json:"status_code"`
    ResponseTimeMs   int       `json:"response_time_ms"`
    RequestSizeBytes int       `json:"request_size_bytes"`
    ResponseSizeBytes int      `json:"response_size_bytes"`
    IPAddress        string    `json:"ip_address"`
    UserAgent        string    `json:"user_agent"`
}

type RateLimitStatus struct {
    Allowed      bool      `json:"allowed"`
    CurrentCount int64     `json:"current_count"`
    LimitValue   int64     `json:"limit_value"`
    ResetTime    time.Time `json:"reset_time"`
}
```

### Production-Ready Middleware Stack

```go
// Complete middleware stack for production
func SetupTenantMiddleware(store Store) []gin.HandlerFunc {
    return []gin.HandlerFunc{
        // 1. Request logging with tenant context
        RequestLoggingMiddleware(),
        
        // 2. Rate limiting per tenant
        TenantRateLimitMiddleware(store),
        
        // 3. Tenant context extraction and validation
        TenantContextMiddleware(store),
        
        // 4. Tenant health monitoring
        TenantHealthMiddleware(store),
        
        // 5. Usage tracking
        UsageTrackingMiddleware(store),
    }
}

func TenantRateLimitMiddleware(store Store) gin.HandlerFunc {
    return func(c *gin.Context) {
        tenantID := getTenantIDFromRequest(c)
        if tenantID == uuid.Nil {
            c.Next()
            return
        }
        
        // Check rate limit
        status, err := store.CheckRateLimit(c.Request.Context(), tenantID, "minute")
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "Rate limit check failed",
                "code":  "RATE_LIMIT_ERROR",
            })
            c.Abort()
            return
        }
        
        // Set rate limit headers
        c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", status.LimitValue))
        c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", status.LimitValue-status.CurrentCount))
        c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", status.ResetTime.Unix()))
        
        if !status.Allowed {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "Rate limit exceeded",
                "code":  "RATE_LIMIT_EXCEEDED",
                "retry_after": status.ResetTime.Sub(time.Now()).Seconds(),
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}

func UsageTrackingMiddleware(store Store) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        // Record API usage after request completes
        tenantID := getTenantIDFromRequest(c)
        if tenantID != uuid.Nil {
            go func() {
                requestLog := &APIRequestLog{
                    TenantID:          tenantID,
                    Endpoint:          c.Request.URL.Path,
                    Method:            c.Request.Method,
                    StatusCode:        c.Writer.Status(),
                    ResponseTimeMs:    int(time.Since(start).Milliseconds()),
                    RequestSizeBytes:  int(c.Request.ContentLength),
                    ResponseSizeBytes: c.Writer.Size(),
                    IPAddress:         c.ClientIP(),
                    UserAgent:         c.Request.UserAgent(),
                }
                
                if err := store.RecordAPIRequest(context.Background(), requestLog); err != nil {
                    log.Printf("Failed to record API request: %v", err)
                }
            }()
        }
    }
}
```

### Background Job Processing

```go
// Background job system for tenant maintenance
type TenantMaintenanceJob struct {
    store Store
    logger *log.Logger
}

func NewTenantMaintenanceJob(store Store) *TenantMaintenanceJob {
    return &TenantMaintenanceJob{
        store:  store,
        logger: log.New(os.Stdout, "[TENANT-MAINTENANCE] ", log.LstdFlags),
    }
}

// Run daily maintenance tasks
func (j *TenantMaintenanceJob) RunDailyMaintenance(ctx context.Context) error {
    j.logger.Println("Starting daily tenant maintenance...")
    
    // 1. Health check all tenants
    if err := j.performHealthChecks(ctx); err != nil {
        j.logger.Printf("Health check failed: %v", err)
    }
    
    // 2. Update usage statistics
    if err := j.updateUsageStatistics(ctx); err != nil {
        j.logger.Printf("Usage statistics update failed: %v", err)
    }
    
    // 3. Perform data archival
    if err := j.performDataArchival(ctx); err != nil {
        j.logger.Printf("Data archival failed: %v", err)
    }
    
    // 4. Analyze churn risk
    if err := j.analyzeChurnRisk(ctx); err != nil {
        j.logger.Printf("Churn risk analysis failed: %v", err)
    }
    
    j.logger.Println("Daily tenant maintenance completed")
    return nil
}

func (j *TenantMaintenanceJob) performHealthChecks(ctx context.Context) error {
    // Get all active tenants
    tenants, err := j.store.ListActiveTenants(ctx)
    if err != nil {
        return fmt.Errorf("failed to list tenants: %w", err)
    }
    
    var tenantIDs []uuid.UUID
    for _, tenant := range tenants {
        tenantIDs = append(tenantIDs, tenant.ID)
    }
    
    // Perform bulk health check
    healthStatuses, err := j.store.BulkHealthCheck(ctx, tenantIDs)
    if err != nil {
        return fmt.Errorf("bulk health check failed: %w", err)
    }
    
    // Process results and send alerts for critical issues
    for _, status := range healthStatuses {
        if status.Status == "critical" {
            j.sendHealthAlert(status)
        }
    }
    
    return nil
}

func (j *TenantMaintenanceJob) sendHealthAlert(status TenantHealthStatus) {
    // Implementation would integrate with your alerting system
    // (email, Slack, PagerDuty, etc.)
    j.logger.Printf("CRITICAL: Tenant %s has critical health issues: %v", 
        status.TenantID, status.Alerts)
}
```

This documentation provides a complete foundation for implementing and maintaining a production-ready multi-tenant ERP system using PostgreSQL Row-Level Security with Go/sqlc integration. The system ensures secure data isolation, scalable resource management, and operational excellence for enterprise multi-tenant applications.
<!-- # Tenant Management -->
<!---->
<!-- ## 🏢 Overview -->
<!---->
<!-- The ERP system is built on a sophisticated multi-tenant architecture that allows multiple organizations to share the same application infrastructure while maintaining complete data isolation and customization capabilities. Each tenant represents a distinct business entity with its own users, data, and configuration. -->
<!---->
<!-- ## 🎯 Multi-Tenancy Strategy -->
<!---->
<!-- ### Tenant Isolation Models -->
<!---->
<!-- #### 1. Database-Level Isolation (Schema per Tenant) -->
<!-- ```sql -->
<!-- -- Each tenant gets its own database schema -->
<!-- CREATE SCHEMA tenant_acme_corp; -->
<!-- CREATE SCHEMA tenant_global_ltd; -->
<!---->
<!-- -- Tables are created within tenant schemas -->
<!-- CREATE TABLE tenant_acme_corp.sales_orders ( -->
<!--     id UUID PRIMARY KEY, -->
<!--     order_number VARCHAR(50) NOT NULL, -->
<!--     customer_id UUID, -->
<!--     total_amount DECIMAL(15,2), -->
<!--     created_at TIMESTAMPTZ DEFAULT NOW() -->
<!-- ); -->
<!-- ``` -->
<!---->
<!-- #### 2. Row-Level Security (Shared Tables) -->
<!-- ```sql -->
<!-- -- Enable RLS on shared tables -->
<!-- ALTER TABLE sales_orders ENABLE ROW LEVEL SECURITY; -->
<!---->
<!-- -- Create tenant-specific policies -->
<!-- CREATE POLICY tenant_isolation ON sales_orders -->
<!--     FOR ALL TO application_user -->
<!--     USING (tenant_id = current_setting('app.current_tenant_id')::UUID); -->
<!---->
<!-- -- Set tenant context in application -->
<!-- SET app.current_tenant_id = 'acme-corp-uuid-here'; -->
<!-- ``` -->
<!---->
<!-- #### 3. Hybrid Approach (Recommended) -->
<!-- - **Core tables**: Row-level security for better resource utilization -->
<!-- - **High-volume tables**: Schema separation for performance -->
<!-- - **Sensitive data**: Complete schema isolation -->
<!---->
<!-- ## 🏗️ Tenant Hierarchy & Structure -->
<!---->
<!-- ### Organization Hierarchy -->
<!---->
<!-- ```mermaid -->
<!-- graph TD -->
<!--     A[Tenant Root] --> B[Organization] -->
<!--     B --> C[Division/Region] -->
<!--     C --> D[Department] -->
<!--     D --> E[Team/Unit] -->
<!--     E --> F[Employees] -->
<!---->
<!--     B --> G[Subsidiary] -->
<!--     G --> H[Branch Office] -->
<!--     H --> I[Department] -->
<!---->
<!--     B --> J[Cost Centers] -->
<!--     B --> K[Profit Centers] -->
<!--     B --> L[Projects] -->
<!-- ``` -->
<!---->
<!-- ### Data Model -->
<!---->
<!-- ```sql -->
<!-- -- Tenant root entity -->
<!-- CREATE TABLE tenants ( -->
<!--     id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -->
<!--     name VARCHAR(255) NOT NULL, -->
<!--     slug VARCHAR(100) UNIQUE NOT NULL, -->
<!--     domain VARCHAR(255), -->
<!--     status VARCHAR(20) DEFAULT 'active', -- active, suspended, terminated -->
<!--     created_at TIMESTAMPTZ DEFAULT NOW(), -->
<!--     subscription_plan VARCHAR(50), -->
<!--     settings JSONB DEFAULT '{}', -->
<!---->
<!--     -- Compliance and legal -->
<!--     country_code CHAR(2), -->
<!--     currency_code CHAR(3), -->
<!--     timezone VARCHAR(50), -->
<!--     fiscal_year_start DATE, -->
<!---->
<!--     -- Billing information -->
<!--     billing_email VARCHAR(255), -->
<!--     billing_address JSONB, -->
<!--     payment_method_id VARCHAR(255), -->
<!---->
<!--     CONSTRAINT valid_status CHECK (status IN ('active', 'suspended', 'terminated')), -->
<!--     CONSTRAINT valid_country CHECK (country_code ~ '^[A-Z]{2}$'), -->
<!--     CONSTRAINT valid_currency CHECK (currency_code ~ '^[A-Z]{3}$') -->
<!-- ); -->
<!---->
<!-- -- Organization hierarchy -->
<!-- CREATE TABLE organizations ( -->
<!--     id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -->
<!--     tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE, -->
<!--     parent_id UUID REFERENCES organizations(id), -->
<!--     name VARCHAR(255) NOT NULL, -->
<!--     org_type VARCHAR(50) NOT NULL, -- company, division, department, team -->
<!--     code VARCHAR(20) UNIQUE, -->
<!--     description TEXT, -->
<!---->
<!--     -- Address and contact -->
<!--     address JSONB, -->
<!--     phone VARCHAR(20), -->
<!--     email VARCHAR(255), -->
<!--     website VARCHAR(255), -->
<!---->
<!--     -- Financial settings -->
<!--     cost_center_code VARCHAR(20), -->
<!--     profit_center_code VARCHAR(20), -->
<!--     budget_allocated DECIMAL(15,2), -->
<!---->
<!--     -- Operational details -->
<!--     manager_id UUID, -->
<!--     is_active BOOLEAN DEFAULT true, -->
<!--     created_at TIMESTAMPTZ DEFAULT NOW(), -->
<!--     updated_at TIMESTAMPTZ DEFAULT NOW(), -->
<!---->
<!--     CONSTRAINT valid_org_type CHECK (org_type IN ('company', 'division', 'department', 'team', 'subsidiary', 'branch')) -->
<!-- ); -->
<!---->
<!-- -- Create indexes for performance -->
<!-- CREATE INDEX idx_organizations_tenant_id ON organizations(tenant_id); -->
<!-- CREATE INDEX idx_organizations_parent_id ON organizations(parent_id); -->
<!-- CREATE INDEX idx_organizations_manager_id ON organizations(manager_id); -->
<!-- ``` -->
<!---->
<!-- ## ⚙️ Tenant Configuration -->
<!---->
<!-- ### Feature Configuration -->
<!---->
<!-- ```yaml -->
<!-- # Tenant-specific feature configuration -->
<!-- tenant_features: -->
<!--   modules: -->
<!--     financial_management: -->
<!--       enabled: true -->
<!--       features: -->
<!--         multi_currency: true -->
<!--         budget_management: true -->
<!--         advanced_reporting: false -->
<!---->
<!--     inventory_management: -->
<!--       enabled: true -->
<!--       features: -->
<!--         multi_warehouse: true -->
<!--         serial_tracking: true -->
<!--         batch_tracking: false -->
<!---->
<!--     hr_management: -->
<!--       enabled: true -->
<!--       features: -->
<!--         payroll_processing: true -->
<!--         performance_management: false -->
<!--         recruitment: true -->
<!---->
<!--     industry_modules: -->
<!--       airline_reservation: false -->
<!--       restaurant_management: false -->
<!--       retail_management: true -->
<!--       forecourt_management: false -->
<!---->
<!--   integrations: -->
<!--     payment_gateways: -->
<!--       stripe: true -->
<!--       paypal: false -->
<!--       square: true -->
<!---->
<!--     accounting_software: -->
<!--       quickbooks: true -->
<!--       xero: false -->
<!--       sage: false -->
<!---->
<!--     communication: -->
<!--       slack: true -->
<!--       microsoft_teams: false -->
<!--       email_notifications: true -->
<!-- ``` -->
<!---->
<!-- ### Customization Framework -->
<!---->
<!-- ```typescript -->
<!-- interface TenantSettings { -->
<!--   branding: { -->
<!--     logo_url: string; -->
<!--     primary_color: string; -->
<!--     secondary_color: string; -->
<!--     font_family: string; -->
<!--     custom_css?: string; -->
<!--   }; -->
<!---->
<!--   business_rules: { -->
<!--     approval_workflows: ApprovalWorkflow[]; -->
<!--     number_sequences: NumberSequence[]; -->
<!--     validation_rules: ValidationRule[]; -->
<!--     custom_fields: CustomField[]; -->
<!--   }; -->
<!---->
<!--   localization: { -->
<!--     language: string; -->
<!--     country: string; -->
<!--     currency: string; -->
<!--     date_format: string; -->
<!--     number_format: string; -->
<!--     timezone: string; -->
<!--   }; -->
<!---->
<!--   security: { -->
<!--     password_policy: PasswordPolicy; -->
<!--     session_timeout: number; -->
<!--     ip_restrictions: string[]; -->
<!--     two_factor_required: boolean; -->
<!--   }; -->
<!---->
<!--   notifications: { -->
<!--     email_settings: EmailSettings; -->
<!--     sms_settings: SmsSettings; -->
<!--     push_notification_settings: PushSettings; -->
<!--   }; -->
<!-- } -->
<!-- ``` -->
<!---->
<!-- ## 🔐 Tenant Security -->
<!---->
<!-- ### Data Isolation -->
<!---->
<!-- #### Context-Based Security -->
<!-- ```go -->
<!-- // Tenant context middleware -->
<!-- func TenantContextMiddleware() gin.HandlerFunc { -->
<!--     return func(c *gin.Context) { -->
<!--         // Extract tenant from subdomain, header, or JWT -->
<!--         tenantID := extractTenantID(c) -->
<!---->
<!--         if tenantID == "" { -->
<!--             c.JSON(401, gin.H{"error": "Tenant not found"}) -->
<!--             c.Abort() -->
<!--             return -->
<!--         } -->
<!---->
<!--         // Validate tenant status -->
<!--         tenant, err := tenantService.GetTenant(tenantID) -->
<!--         if err != nil || tenant.Status != "active" { -->
<!--             c.JSON(403, gin.H{"error": "Tenant access denied"}) -->
<!--             c.Abort() -->
<!--             return -->
<!--         } -->
<!---->
<!--         // Set tenant context -->
<!--         c.Set("tenant_id", tenantID) -->
<!--         c.Set("tenant", tenant) -->
<!---->
<!--         // Set database context for RLS -->
<!--         db := database.GetConnection() -->
<!--         db.Exec("SET app.current_tenant_id = ?", tenantID) -->
<!---->
<!--         c.Next() -->
<!--     } -->
<!-- } -->
<!-- ``` -->
<!---->
<!-- #### Database Query Enforcement -->
<!-- ```go -->
<!-- // Repository pattern with automatic tenant filtering -->
<!-- type BaseRepository struct { -->
<!--     db       *gorm.DB -->
<!--     tenantID string -->
<!-- } -->
<!---->
<!-- func (r *BaseRepository) Find(dest interface{}, conditions ...interface{}) error { -->
<!--     // Automatically add tenant filter to all queries -->
<!--     return r.db.Where("tenant_id = ?", r.tenantID).Find(dest, conditions...).Error -->
<!-- } -->
<!---->
<!-- func (r *BaseRepository) Create(value interface{}) error { -->
<!--     // Automatically set tenant_id before creating -->
<!--     if model, ok := value.(TenantModel); ok { -->
<!--         model.SetTenantID(r.tenantID) -->
<!--     } -->
<!--     return r.db.Create(value).Error -->
<!-- } -->
<!-- ``` -->
<!---->
<!-- ### Access Control -->
<!---->
<!-- #### Tenant-Level Permissions -->
<!-- ```json -->
<!-- { -->
<!--   "tenant_permissions": { -->
<!--     "tenant_admin": [ -->
<!--       "tenant:settings:read", -->
<!--       "tenant:settings:write", -->
<!--       "tenant:users:manage", -->
<!--       "tenant:modules:configure", -->
<!--       "tenant:integrations:manage" -->
<!--     ], -->
<!--     "organization_admin": [ -->
<!--       "organization:settings:read", -->
<!--       "organization:settings:write", -->
<!--       "organization:users:read", -->
<!--       "organization:users:invite" -->
<!--     ], -->
<!--     "department_manager": [ -->
<!--       "department:view", -->
<!--       "department:users:read", -->
<!--       "department:reports:generate" -->
<!--     ] -->
<!--   } -->
<!-- } -->
<!-- ``` -->
<!---->
<!-- ## 🚀 Tenant Provisioning -->
<!---->
<!-- ### Automated Tenant Setup -->
<!---->
<!-- ```yaml -->
<!-- # Tenant provisioning workflow -->
<!-- tenant_provisioning: -->
<!--   steps: -->
<!--     1. tenant_creation: -->
<!--         - validate_tenant_data -->
<!--         - create_tenant_record -->
<!--         - generate_tenant_slug -->
<!--         - setup_default_organization -->
<!---->
<!--     2. database_setup: -->
<!--         - create_tenant_schema -->
<!--         - run_schema_migrations -->
<!--         - setup_default_data -->
<!--         - configure_row_level_security -->
<!---->
<!--     3. user_setup: -->
<!--         - create_admin_user -->
<!--         - setup_default_roles -->
<!--         - send_welcome_email -->
<!--         - generate_setup_wizard_token -->
<!---->
<!--     4. configuration: -->
<!--         - apply_plan_features -->
<!--         - setup_default_workflows -->
<!--         - configure_number_sequences -->
<!--         - setup_default_chart_of_accounts -->
<!---->
<!--     5. integration_setup: -->
<!--         - configure_email_service -->
<!--         - setup_notification_channels -->
<!--         - initialize_audit_logging -->
<!--         - setup_backup_schedule -->
<!---->
<!--   rollback_procedures: -->
<!--     - cleanup_database_schema -->
<!--     - remove_user_accounts -->
<!--     - delete_tenant_record -->
<!--     - cleanup_file_storage -->
<!-- ``` -->
<!---->
<!-- ### Tenant Onboarding API -->
<!---->
<!-- ```typescript -->
<!-- interface TenantProvisioningRequest { -->
<!--   company_name: string; -->
<!--   admin_email: string; -->
<!--   admin_first_name: string; -->
<!--   admin_last_name: string; -->
<!--   country_code: string; -->
<!--   industry: string; -->
<!--   company_size: number; -->
<!--   subscription_plan: string; -->
<!---->
<!--   customization?: { -->
<!--     custom_domain?: string; -->
<!--     branding?: BrandingConfig; -->
<!--     features?: FeatureConfig; -->
<!--   }; -->
<!-- } -->
<!---->
<!-- // POST /api/v1/tenants/provision -->
<!-- async function provisionTenant(request: TenantProvisioningRequest): Promise<TenantProvisioningResponse> { -->
<!--   // Step 1: Validate request -->
<!--   await validateProvisioningRequest(request); -->
<!---->
<!--   // Step 2: Create tenant in transaction -->
<!--   const tenant = await database.transaction(async (tx) => { -->
<!--     const tenant = await createTenant(tx, request); -->
<!--     await setupTenantSchema(tx, tenant.id); -->
<!--     await createAdminUser(tx, tenant.id, request); -->
<!--     return tenant; -->
<!--   }); -->
<!---->
<!--   // Step 3: Initialize tenant data -->
<!--   await initializeTenantData(tenant.id, request); -->
<!---->
<!--   // Step 4: Send welcome email -->
<!--   await sendTenantWelcomeEmail(tenant, request.admin_email); -->
<!---->
<!--   return { -->
<!--     tenant_id: tenant.id, -->
<!--     tenant_slug: tenant.slug, -->
<!--     admin_setup_url: `https://${tenant.slug}.awo.com/setup`, -->
<!--     status: 'provisioned' -->
<!--   }; -->
<!-- } -->
<!-- ``` -->
<!---->
<!-- ## 📊 Tenant Analytics & Monitoring -->
<!---->
<!-- ### Usage Metrics -->
<!---->
<!-- ```yaml -->
<!-- tenant_metrics: -->
<!--   usage_tracking: -->
<!--     - active_users_daily -->
<!--     - active_users_monthly -->
<!--     - api_requests_count -->
<!--     - storage_usage_gb -->
<!--     - data_transfer_gb -->
<!--     - feature_usage_stats -->
<!---->
<!--   business_metrics: -->
<!--     - transactions_processed -->
<!--     - invoices_generated -->
<!--     - orders_processed -->
<!--     - projects_completed -->
<!--     - employees_managed -->
<!---->
<!--   performance_metrics: -->
<!--     - average_response_time -->
<!--     - error_rate_percentage -->
<!--     - uptime_percentage -->
<!--     - database_query_performance -->
<!-- ``` -->
<!---->
<!-- ### Tenant Health Dashboard -->
<!---->
<!-- ```typescript -->
<!-- interface TenantHealthMetrics { -->
<!--   tenant_id: string; -->
<!--   period: 'daily' | 'weekly' | 'monthly'; -->
<!---->
<!--   usage: { -->
<!--     active_users: number; -->
<!--     api_calls: number; -->
<!--     storage_used_gb: number; -->
<!--     bandwidth_used_gb: number; -->
<!--   }; -->
<!---->
<!--   performance: { -->
<!--     avg_response_time_ms: number; -->
<!--     error_rate: number; -->
<!--     uptime_percentage: number; -->
<!--   }; -->
<!---->
<!--   business: { -->
<!--     transactions_count: number; -->
<!--     revenue_processed: number; -->
<!--     documents_created: number; -->
<!--   }; -->
<!---->
<!--   alerts: TenantAlert[]; -->
<!--   recommendations: TenantRecommendation[]; -->
<!-- } -->
<!-- ``` -->
<!---->
<!-- ## 💰 Billing & Subscription Management -->
<!---->
<!-- ### Subscription Plans -->
<!---->
<!-- ```yaml -->
<!-- subscription_plans: -->
<!--   starter: -->
<!--     name: "Starter" -->
<!--     price: 29 -->
<!--     billing_cycle: monthly -->
<!--     limits: -->
<!--       users: 5 -->
<!--       storage_gb: 10 -->
<!--       api_calls_monthly: 10000 -->
<!--     features: -->
<!--       - core_modules -->
<!--       - basic_reporting -->
<!--       - email_support -->
<!---->
<!--   professional: -->
<!--     name: "Professional" -->
<!--     price: 99 -->
<!--     billing_cycle: monthly -->
<!--     limits: -->
<!--       users: 25 -->
<!--       storage_gb: 100 -->
<!--       api_calls_monthly: 100000 -->
<!--     features: -->
<!--       - all_core_modules -->
<!--       - advanced_reporting -->
<!--       - workflow_automation -->
<!--       - priority_support -->
<!---->
<!--   enterprise: -->
<!--     name: "Enterprise" -->
<!--     price: 299 -->
<!--     billing_cycle: monthly -->
<!--     limits: -->
<!--       users: unlimited -->
<!--       storage_gb: 1000 -->
<!--       api_calls_monthly: 1000000 -->
<!--     features: -->
<!--       - all_modules -->
<!--       - custom_integrations -->
<!--       - dedicated_support -->
<!--       - sla_guarantee -->
<!-- ``` -->
<!---->
<!-- ### Usage-Based Billing -->
<!---->
<!-- ```typescript -->
<!-- interface UsageBillingRules { -->
<!--   base_price: number; -->
<!---->
<!--   usage_tiers: { -->
<!--     users: { -->
<!--       included: number; -->
<!--       overage_price_per_user: number; -->
<!--     }; -->
<!---->
<!--     storage: { -->
<!--       included_gb: number; -->
<!--       overage_price_per_gb: number; -->
<!--     }; -->
<!---->
<!--     api_calls: { -->
<!--       included_monthly: number; -->
<!--       overage_price_per_1000: number; -->
<!--     }; -->
<!---->
<!--     transactions: { -->
<!--       included_monthly: number; -->
<!--       overage_price_per_transaction: number; -->
<!--     }; -->
<!--   }; -->
<!---->
<!--   feature_addons: { -->
<!--     advanced_analytics: number; -->
<!--     custom_branding: number; -->
<!--     priority_support: number; -->
<!--     additional_integrations: number; -->
<!--   }; -->
<!-- } -->
<!-- ``` -->
<!---->
<!-- ## 🔄 Tenant Migration & Backup -->
<!---->
<!-- ### Data Migration -->
<!---->
<!-- ```typescript -->
<!-- interface TenantMigrationPlan { -->
<!--   source_tenant_id: string; -->
<!--   target_tenant_id: string; -->
<!--   migration_type: 'full' | 'selective' | 'merge'; -->
<!---->
<!--   data_selection: { -->
<!--     include_users: boolean; -->
<!--     include_transactions: boolean; -->
<!--     include_historical_data: boolean; -->
<!--     date_range?: DateRange; -->
<!--     entity_filters?: EntityFilter[]; -->
<!--   }; -->
<!---->
<!--   mapping_rules: { -->
<!--     user_mapping: UserMapping[]; -->
<!--     organization_mapping: OrganizationMapping[]; -->
<!--     account_mapping: AccountMapping[]; -->
<!--   }; -->
<!---->
<!--   validation_rules: ValidationRule[]; -->
<!--   rollback_plan: RollbackPlan; -->
<!-- } -->
<!-- ``` -->
<!---->
<!-- ### Backup Strategy -->
<!---->
<!-- ```yaml -->
<!-- backup_configuration: -->
<!--   database_backups: -->
<!--     frequency: daily -->
<!--     retention: 90_days -->
<!--     compression: true -->
<!--     encryption: true -->
<!---->
<!--   file_backups: -->
<!--     frequency: daily -->
<!--     retention: 30_days -->
<!--     incremental: true -->
<!---->
<!--   point_in_time_recovery: -->
<!--     enabled: true -->
<!--     retention: 7_days -->
<!--     granularity: 15_minutes -->
<!---->
<!--   cross_region_replication: -->
<!--     enabled: true -->
<!--     regions: [us-east-1, eu-west-1] -->
<!--     sync_frequency: hourly -->
<!-- ``` -->
<!---->
<!-- ## 🛠️ Tenant Management API -->
<!---->
<!-- ### Core Operations -->
<!---->
<!-- ```typescript -->
<!-- // Tenant management endpoints -->
<!-- interface TenantManagementAPI { -->
<!--   // Tenant CRUD operations -->
<!--   GET    /api/v1/tenants/{tenant_id} -->
<!--   PUT    /api/v1/tenants/{tenant_id} -->
<!--   DELETE /api/v1/tenants/{tenant_id} -->
<!---->
<!--   // Organization management -->
<!--   GET    /api/v1/tenants/{tenant_id}/organizations -->
<!--   POST   /api/v1/tenants/{tenant_id}/organizations -->
<!--   PUT    /api/v1/tenants/{tenant_id}/organizations/{org_id} -->
<!--   DELETE /api/v1/tenants/{tenant_id}/organizations/{org_id} -->
<!---->
<!--   // Configuration management -->
<!--   GET    /api/v1/tenants/{tenant_id}/settings -->
<!--   PUT    /api/v1/tenants/{tenant_id}/settings -->
<!--   GET    /api/v1/tenants/{tenant_id}/features -->
<!--   PUT    /api/v1/tenants/{tenant_id}/features -->
<!---->
<!--   // User management -->
<!--   GET    /api/v1/tenants/{tenant_id}/users -->
<!--   POST   /api/v1/tenants/{tenant_id}/users/invite -->
<!--   PUT    /api/v1/tenants/{tenant_id}/users/{user_id} -->
<!--   DELETE /api/v1/tenants/{tenant_id}/users/{user_id} -->
<!---->
<!--   // Analytics and monitoring -->
<!--   GET    /api/v1/tenants/{tenant_id}/metrics -->
<!--   GET    /api/v1/tenants/{tenant_id}/usage -->
<!--   GET    /api/v1/tenants/{tenant_id}/health -->
<!-- } -->
<!-- ``` -->
<!---->
<!-- This  tenant management system ensures secure, scalable, and customizable multi-tenant operations while maintaining data isolation and providing flexibility for diverse business needs. -->
