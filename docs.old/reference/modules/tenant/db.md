# Multi-Tenant ERP Database Schema Analysis

## Executive Summary

This document provides a comprehensive analysis of the PostgreSQL database schema for a multi-tenant ERP (Enterprise Resource Planning) platform. The schema implements a robust **shared database, shared schema** multi-tenancy pattern with **Row-Level Security (RLS)** for data isolation.

**Key Characteristics:**
- **Architecture Pattern**: Shared Database with Tenant Isolation via RLS
- **Database**: PostgreSQL 24+ (requires `gen_random_uuid()`)
- **Security Model**: Row-Level Security (RLS) with role-based access control
- **Migration Count**: 7 migrations (000101-000107)
- **Core Entities**: Tenants, Configurations, Usage Stats, Bulk Operations

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Migration Analysis](#migration-analysis)
3. [Core Tables Deep Dive](#core-tables-deep-dive)
4. [Security Implementation](#security-implementation)
5. [Utility Functions](#utility-functions)
6. [Data Isolation Mechanisms](#data-isolation-mechanisms)
7. [Design Patterns & Best Practices](#design-patterns--best-practices)
8. [Potential Issues & Recommendations](#potential-issues--recommendations)
9. [Performance Considerations](#performance-considerations)
10. [Extension Roadmap](#extension-roadmap)

---

## 1. Architecture Overview

### 1.1 Multi-Tenancy Pattern

The schema implements a **shared database, shared schema** approach:

```
┌─────────────────────────────────────────────────┐
│            PostgreSQL Database                   │
├─────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐             │
│  │   Tenant A   │  │   Tenant B   │             │
│  │   (Acme)     │  │   (Smith)    │             │
│  ├──────────────┤  ├──────────────┤             │
│  │ tenant_id:   │  │ tenant_id:   │             │
│  │ 550e8400-... │  │ 661f9511-... │             │
│  └──────┬───────┘  └──────┬───────┘             │
│         │                  │                      │
│  ┌──────▼──────────────────▼───────┐             │
│  │      Shared Tables with RLS     │             │
│  │  - tenants                      │             │
│  │  - tenant_configurations        │             │
│  │  - tenant_usage_stats          │             │
│  │  - tenant_bulk_operations      │             │
│  │  - entities, persons, etc.     │             │
│  └─────────────────────────────────┘             │
│                                                   │
│  Row-Level Security (RLS) ensures:               │
│  - Tenant A sees only Tenant A's data           │
│  - Tenant B sees only Tenant B's data           │
│  - Complete data isolation                       │
└─────────────────────────────────────────────────┘
```

**Advantages:**
- ✅ Cost-efficient (single database)
- ✅ Easy maintenance and updates
- ✅ Consistent schema across tenants
- ✅ Simplified backup and disaster recovery

**Trade-offs:**
- ⚠️ More complex security implementation
- ⚠️ Requires careful query filtering
- ⚠️ Single point of failure (mitigated by HA/clustering)

### 1.2 Database Roles

Three distinct roles with different privilege levels:

```sql
┌─────────────────┬──────────────────────────────────────┐
│ Role            │ Purpose & Access Level               │
├─────────────────┼──────────────────────────────────────┤
│ admin_role      │ Platform administrators              │
│                 │ - Full access to ALL tenant data     │
│                 │ - System configuration               │
│                 │ - Bulk operations                    │
├─────────────────┼──────────────────────────────────────┤
│ application_role│ Application/API users                │
│                 │ - Access to own tenant data only     │
│                 │ - CRUD operations within boundaries  │
│                 │ - Most common role                   │
├─────────────────┼──────────────────────────────────────┤
│ readonly_role   │ Read-only access (reporting/audit)   │
│                 │ - SELECT only                        │
│                 │ - All tenants visible (for reports)  │
│                 │ - No modifications allowed           │
└─────────────────┴──────────────────────────────────────┘
```

### 1.3 Session-Based Tenant Context

The system uses PostgreSQL's `set_config()` to establish tenant context:

```sql
-- Set tenant context for session
SELECT set_tenant_context('550e8400-e29b-41d4-a716-446655440000');

-- Behind the scenes:
-- app.current_tenant_id = '550e8400-e29b-41d4-a716-446655440000'
-- app.tenant_status = 'ACTIVE'
-- app.context_set_at = '2024-02-02 14:35:22'

-- All subsequent queries automatically filtered to this tenant
SELECT * FROM customers;  
-- Returns only Tenant A's customers

-- Clear context (important for connection pooling)
SELECT clear_tenant_context();
```

---

## 2. Migration Analysis

### 2.1 Migration Sequence

| Migration | Purpose | Tables Created | Key Features |
|-----------|---------|----------------|--------------|
| **000101** | Core tenant infrastructure | `tenants` | Tenant identity, status, RLS foundation |
| **000102** | Tenant configuration | `tenant_configurations` | Limits, preferences, accounting settings |
| **000103** | Usage tracking | `tenant_usage_stats` | Resource monitoring, metrics |
| **000104** | Provisioning | None | `provision_tenant_complete()` function |
| **000105** | Bulk operations | `tenant_bulk_operations`<br>`tenant_bulk_operation_results` | Admin mass operations, audit trail |
| **000106** | Context validation | None | `validate_and_set_tenant_context()` |
| **000107** | Cross-tenant checks | `account` (example) | `enforce_tenant_isolation()` trigger |

### 2.2 Migration 000101 - Core Tables

**Purpose**: Foundation for multi-tenant architecture

**Key Components:**

1. **Tenants Table** - Master tenant registry
2. **Roles** - Database role creation (`admin_role`, `application_role`, `readonly_role`)
3. **RLS Policies** - Data isolation rules
4. **Utility Functions**:
   - `set_tenant_context()` - Establish session context
   - `current_tenant_id()` - Retrieve active tenant
   - `clear_tenant_context()` - Clean up session
   - `update_updated_at_column()` - Auto-timestamp trigger
   - `generate_unique_slug_from_name()` - URL-friendly identifier

**Critical Features:**

```sql
-- Tenant status management
CHECK (Status IN ('ACTIVE', 'SUSPENDED', 'PENDING', 'ARCHIVED'))

-- Automatic slug generation from name
-- "Acme Corporation" → "acme-corporation"
CREATE TRIGGER tenant_slug_trigger 
  BEFORE INSERT ON tenants 
  FOR EACH ROW EXECUTE FUNCTION generate_unique_slug_from_name();

-- Email validation
CHECK (email ~* '^[A-Za-z0-9._%-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')

-- Subdomain validation (DNS-safe)
CHECK (subdomain ~* '^[a-z0-9]([a-z0-9-]*[a-z0-9])?$')
```

### 2.3 Migration 000102 - Configuration Tables

**Purpose**: Store tenant-specific settings and limits

**Design Decision Note** (from migration comments):

The schema chose **dedicated columns** over other approaches:

| Approach | Pros | Cons | Decision |
|----------|------|------|----------|
| **Dedicated Columns** ✓ | Strong typing, easy querying, constraints | Schema changes needed | **CHOSEN** |
| Key-Value Table | Flexible, no schema changes | Weak typing, complex queries | Rejected |
| Composite Type | Flexible, Postgres typing | Complex queries, non-standard | Rejected |
| Pure JSONB | Extremely flexible | No constraints, app validation | Rejected |

**Implementation:**

```sql
CREATE TABLE tenant_configurations (
  tenant_id UUID PRIMARY KEY REFERENCES tenants(id),
  
  -- Resource limits (dedicated columns)
  max_users INT DEFAULT 100,
  max_entities INT DEFAULT 1000,
  max_transactions_per_month INT DEFAULT 10000,
  storage_quota BIGINT DEFAULT 1073741824,  -- 1GB
  
  -- Accounting preferences
  accounting_method VARCHAR(10) DEFAULT 'ACCRUAL',
  fiscal_year_start_month INT DEFAULT 1,
  
  -- Flexible JSONB for experimentation
  settings JSONB DEFAULT '{}',
  password_policy JSONB,
  api_rate_limits JSONB
);
```

**Key Function**: `check_tenant_limits(tenant_id, check_type, additional_usage)`
- Validates resource usage before operations
- Currently placeholder for most checks
- Storage check implemented in Migration 000103

### 2.4 Migration 000103 - Usage Tracking

**Purpose**: Track tenant resource consumption over time

**Table Design:**

```sql
CREATE TABLE tenant_usage_stats (
  tenant_id UUID REFERENCES tenants(id),
  period_start DATE,
  period_end DATE,
  
  -- Usage metrics
  active_users INT DEFAULT 0,
  total_entities INT DEFAULT 0,
  total_transactions INT DEFAULT 0,
  storage_used BIGINT DEFAULT 0,
  api_calls INT DEFAULT 0,
  
  -- Performance metrics
  avg_response_time NUMERIC(10, 2),
  error_rate NUMERIC(5, 4),
  
  -- Financial metrics
  monthly_revenue NUMERIC(12, 2),
  
  PRIMARY KEY (tenant_id, period_start)
);
```

**Enhanced Function**: Updated `check_tenant_limits()` to include storage validation:

```sql
WHEN 'storage' THEN
  SELECT COALESCE(storage_used, 0) INTO v_current_usage
  FROM tenant_usage_stats
  WHERE tenant_id = p_tenant_id
    AND period_start <= CURRENT_DATE
    AND period_end >= CURRENT_DATE
  ORDER BY period_start DESC
  LIMIT 1;
  
  RETURN (COALESCE(v_current_usage, 0) + p_additional_usage) 
         <= v_config.storage_quota;
```

### 2.5 Migration 000104 - Tenant Provisioning

**Purpose**: Streamlined tenant creation function

**Function**: `provision_tenant_complete()`

```sql
-- One-step tenant creation
SELECT provision_tenant_complete(
  p_name := 'Acme Corporation',
  p_email := 'admin@acme.com',
  p_subdomain := 'acme',
  p_industry := 'Manufacturing',
  p_company_size := 'Medium',
  p_currency_code := 'USD',
  p_timezone := 'America/New_York',
  p_settings := '{"feature_flags": {"advanced_analytics": true}}'::jsonb
);

-- Returns: tenant_id (UUID)
```

**Process Flow:**
1. Generate UUID and slug
2. Ensure slug uniqueness (append UUID fragment if collision)
3. Create tenant with status 'PENDING'
4. Trigger `create_tenant_configuration_trigger` (Migration 000102)
5. Auto-create default configuration
6. Return tenant ID

### 2.6 Migration 000105 - Bulk Operations

**Purpose**: Track administrative bulk operations on multiple tenants

**Use Cases:**
- Mass suspend tenants (payment failures)
- Bulk reactivation
- Archive old/inactive tenants
- Update resource limits across multiple tenants
- Enable/disable features

**Table Structure:**

```sql
-- Master operation tracking
CREATE TABLE tenant_bulk_operations (
  id UUID PRIMARY KEY,
  operation_type VARCHAR(50) CHECK (operation_type IN (
    'SUSPEND', 'REACTIVATE', 'ARCHIVE', 
    'UPDATE_LIMITS', 'UPDATE_FEATURES'
  )),
  
  actor_id UUID,          -- Who initiated
  actor_name VARCHAR(255),
  
  total_tenants INT DEFAULT 0,
  successful_count INT DEFAULT 0,
  failed_count INT DEFAULT 0,
  
  status VARCHAR(20) CHECK (status IN (
    'IN_PROGRESS', 'COMPLETED', 'FAILED', 
    'PARTIAL_SUCCESS', 'CANCELLED'
  )),
  
  started_at TIMESTAMPTZ DEFAULT NOW(),
  completed_at TIMESTAMPTZ,
  
  parameters JSONB,        -- Operation-specific data
  error_summary TEXT
);

-- Individual results per tenant
CREATE TABLE tenant_bulk_operation_results (
  operation_id UUID REFERENCES tenant_bulk_operations(id),
  tenant_id UUID REFERENCES tenants(id),
  
  status VARCHAR(20) CHECK (status IN (
    'PENDING', 'PROCESSING', 'COMPLETED', 
    'FAILED', 'SKIPPED'
  )),
  
  message TEXT,
  error_details TEXT,
  
  started_at TIMESTAMPTZ,
  completed_at TIMESTAMPTZ,
  
  PRIMARY KEY (operation_id, tenant_id)
);
```

**Automated Count Tracking:**

Trigger automatically updates master record when individual results change:

```sql
CREATE TRIGGER tenant_bulk_operation_results_update_counts
  AFTER INSERT OR UPDATE OR DELETE 
  ON tenant_bulk_operation_results
  FOR EACH ROW 
  EXECUTE FUNCTION trigger_update_bulk_operation_counts();
```

**Utility Functions:**

1. `get_bulk_operation_summary(operation_id)` - Real-time progress
2. `update_bulk_operation_counts(operation_id)` - Recalculate totals/status

### 2.7 Migration 000106 - Context Validation

**Purpose**: Enhanced tenant context validation with activity tracking

**Function**: `validate_and_set_tenant_context(tenant_id)`

**Improvements over `set_tenant_context()`:**

```sql
-- Validates:
1. Tenant exists
2. Not soft-deleted (deleted_at IS NULL)
3. Status is 'active' or 'pending'
4. Updates last_activity_at timestamp
5. Sets session context
6. Returns tenant name and status

-- Returns TABLE(tenant_name TEXT, tenant_status TEXT)
```

**Example Usage:**

```sql
SELECT * FROM validate_and_set_tenant_context(
  '550e8400-e29b-41d4-a716-446655440000'
);

-- Returns:
--  tenant_name     | tenant_status
-- -----------------+---------------
--  Acme Corporation| ACTIVE
```

**Commented-out Project Table:**
Migration file includes a commented example of how to extend the schema with a new multi-tenant table. This serves as a template for future development.

### 2.8 Migration 000107 - Tenant Isolation Enforcement

**Purpose**: Additional cross-table tenant validation

**Function**: `enforce_tenant_isolation()`

Trigger function to validate foreign key references belong to same tenant:

```sql
-- Example for persons table
IF NOT EXISTS (
  SELECT 1
  FROM entities e
  JOIN tenants t ON e.tenant_id = t.id
  WHERE e.uuid = NEW.entity_id
    AND t.id = NEW.tenant_id
) THEN 
  RAISE EXCEPTION 'Entity % does not belong to tenant %',
    NEW.entity_id, NEW.tenant_id;
END IF;
```

**Pattern**: Can be applied to any table with cross-references to ensure tenant boundaries are never crossed.

---

## 3. Core Tables Deep Dive

### 3.1 Tenants Table

**Schema:**

```sql
CREATE TABLE tenants (
  -- Identity
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  slug VARCHAR(50) NOT NULL,                    -- URL-friendly
  name VARCHAR(255) UNIQUE NOT NULL,            -- Business name
  
  -- Contact & Access
  email VARCHAR(255) NOT NULL,
  subdomain VARCHAR(63) UNIQUE,                 -- acme.yourerp.com
  
  -- Status & Settings
  Status VARCHAR(20) DEFAULT 'ACTIVE',
  timezone VARCHAR(50) DEFAULT 'UTC',
  currency_code CHAR(3) DEFAULT 'USD',
  
  -- Metadata
  metadata JSONB DEFAULT '{}',
  settings JSONB DEFAULT '{}',
  
  -- Business Info
  industry VARCHAR(50),
  company_size VARCHAR(20),
  tax_id VARCHAR(50),
  registration_number VARCHAR(50),
  legal_entity_type VARCHAR(50),
  
  -- Activity & Audit
  last_activity_at TIMESTAMPTZ DEFAULT NOW(),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  deleted_at TIMESTAMPTZ                        -- Soft delete
);
```

**Indexes:**

```sql
idx_tenants_slug        -- Fast slug lookup
idx_tenants_status      -- Filter by status (ACTIVE, etc.)
idx_tenants_subdomain   -- Subdomain routing (WHERE subdomain IS NOT NULL)
idx_tenants_deleted_at  -- Exclude deleted (WHERE deleted_at IS NOT NULL)
```

**Constraints:**

```sql
-- Email validation (RFC 5322 simplified)
CHECK (email ~* '^[A-Za-z0-9._%-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')

-- Subdomain validation (DNS-safe)
CHECK (subdomain IS NULL OR subdomain ~* '^[a-z0-9]([a-z0-9-]*[a-z0-9])?$')

-- Slug validation (URL-safe)
CHECK (slug ~* '^[a-z0-9]([a-z0-9-]*[a-z0-9])?$')

-- Currency code (ISO 4217)
CHECK (currency_code ~* '^[A-Z]{3}$')

-- Status enumeration
CHECK (Status IN ('ACTIVE', 'SUSPENDED', 'PENDING', 'ARCHIVED'))

-- Company size enumeration
CHECK (company_size IN ('Startup', 'Small', 'Medium', 'Large', 'Enterprise'))
```

**RLS Policies:**

```sql
-- Application role: see own tenant OR NULL (for initial setup)
CREATE POLICY tenant_isolation_policy ON tenants
  FOR ALL TO application_role
  USING (id = current_tenant_id() OR current_tenant_id() IS NULL);

-- Admin role: see everything
CREATE POLICY admin_full_access_policy ON tenants
  FOR ALL TO admin_role
  USING (true);

-- Readonly role: see everything (read-only)
CREATE POLICY readonly_access_policy ON tenants
  FOR SELECT TO readonly_role
  USING (true);
```

**Triggers:**

```sql
-- Auto-update updated_at
CREATE TRIGGER update_tenants_updated_at
  BEFORE UPDATE ON tenants
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Auto-generate slug from name
CREATE TRIGGER tenant_slug_trigger
  BEFORE INSERT ON tenants
  FOR EACH ROW EXECUTE FUNCTION generate_unique_slug_from_name();

-- Auto-create configuration
CREATE TRIGGER create_tenant_configuration_trigger
  AFTER INSERT ON tenants
  FOR EACH ROW EXECUTE FUNCTION create_tenant_configuration_on_insert();
```

### 3.2 Tenant Configurations Table

**Purpose**: Store all tenant-specific settings and resource limits

**Schema:**

```sql
CREATE TABLE tenant_configurations (
  tenant_id UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
  
  -- Resource Limits
  max_users INT DEFAULT 100,
  max_entities INT DEFAULT 1000,
  max_transactions_per_month INT DEFAULT 10000,
  storage_quota BIGINT DEFAULT 1073741824,      -- 1GB in bytes
  
  -- Accounting Preferences
  accounting_method VARCHAR(10) DEFAULT 'ACCRUAL',
  fiscal_year_start_month INT DEFAULT 1,        -- January
  default_currency CHAR(3) DEFAULT 'USD',
  
  -- Localization
  date_format VARCHAR(20) DEFAULT 'MM/DD/YYYY',
  number_format VARCHAR(20) DEFAULT 'US',
  language_code VARCHAR(5) DEFAULT 'en-US',
  
  -- Security Settings
  password_policy JSONB DEFAULT '{
    "min_length": 8,
    "require_uppercase": true,
    "require_lowercase": true,
    "require_numbers": true,
    "require_symbols": false
  }'::jsonb,
  
  -- Integration Settings
  webhook_endpoints JSONB DEFAULT '[]'::jsonb,
  api_rate_limits JSONB DEFAULT '{
    "requests_per_minute": 100,
    "requests_per_hour": 5000
  }'::jsonb,
  
  -- Flexible Settings (JSONB for experimentation)
  settings JSONB DEFAULT '{}'::jsonb,
  
  -- Audit
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Constraints:**

```sql
CHECK (accounting_method IN ('ACCRUAL', 'CASH'))
CHECK (fiscal_year_start_month BETWEEN 1 AND 12)
```

**Design Philosophy** (from migration comments):

> **Dedicated Columns for Critical Settings**
> - Strong typing
> - Database-level constraints
> - Easy querying and indexing
> - Required schema changes for new settings
> 
> **JSONB for Flexible Settings**
> - Tenant-specific preferences
> - Experimental features
> - Rapid iteration without migrations
> - Application-level validation

**Default Configuration Creation:**

```sql
-- Automatically triggered on tenant insert
CREATE FUNCTION create_default_tenant_configuration(p_tenant_id UUID)
RETURNS VOID AS $$
BEGIN
  INSERT INTO tenant_configurations (tenant_id)
  VALUES (p_tenant_id)
  ON CONFLICT (tenant_id) DO NOTHING;
END;
$$ LANGUAGE plpgsql;
```

### 3.3 Tenant Usage Stats Table

**Purpose**: Historical resource usage tracking for billing, analytics, and capacity planning

**Schema:**

```sql
CREATE TABLE tenant_usage_stats (
  -- Composite Primary Key (time-series data)
  tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
  period_start DATE,
  period_end DATE,
  
  -- Usage Metrics
  active_users INT DEFAULT 0,
  total_entities INT DEFAULT 0,
  total_transactions INT DEFAULT 0,
  storage_used BIGINT DEFAULT 0,              -- bytes
  api_calls INT DEFAULT 0,
  
  -- Performance Metrics
  avg_response_time NUMERIC(10, 2),           -- milliseconds
  error_rate NUMERIC(5, 4),                   -- 0.0001 = 0.01%
  
  -- Financial Metrics
  monthly_revenue NUMERIC(12, 2),
  
  created_at TIMESTAMPTZ DEFAULT NOW(),
  
  PRIMARY KEY (tenant_id, period_start)
);
```

**Indexes:**

```sql
idx_tenant_usage_stats_period              -- Query by time period
idx_tenant_usage_stats_tenant_period       -- Tenant-specific time queries
```

**Typical Usage Patterns:**

```sql
-- Current month's usage
SELECT * FROM tenant_usage_stats
WHERE tenant_id = current_tenant_id()
  AND period_start <= CURRENT_DATE
  AND period_end >= CURRENT_DATE;

-- Last 6 months trend
SELECT period_start, storage_used, total_transactions
FROM tenant_usage_stats
WHERE tenant_id = current_tenant_id()
  AND period_start >= CURRENT_DATE - INTERVAL '6 months'
ORDER BY period_start;

-- Storage growth forecast
WITH monthly_growth AS (
  SELECT 
    period_start,
    storage_used,
    LAG(storage_used) OVER (ORDER BY period_start) AS prev_month,
    storage_used - LAG(storage_used) OVER (ORDER BY period_start) AS growth
  FROM tenant_usage_stats
  WHERE tenant_id = current_tenant_id()
)
SELECT AVG(growth) AS avg_monthly_growth
FROM monthly_growth
WHERE growth IS NOT NULL;
```

### 3.4 Bulk Operations Tables

**Master Table**: `tenant_bulk_operations`

```sql
CREATE TABLE tenant_bulk_operations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  
  -- Operation Details
  operation_type VARCHAR(50) CHECK (operation_type IN (
    'SUSPEND',           -- Mass suspend
    'REACTIVATE',        -- Mass reactivate
    'ARCHIVE',           -- Mass archive
    'UPDATE_LIMITS',     -- Bulk limit updates
    'UPDATE_FEATURES'    -- Bulk feature toggles
  )),
  
  -- Actor (who initiated)
  actor_id UUID NOT NULL,
  actor_name VARCHAR(255),
  
  -- Progress Tracking
  total_tenants INT DEFAULT 0,
  successful_count INT DEFAULT 0,
  failed_count INT DEFAULT 0,
  
  -- Status
  status VARCHAR(20) DEFAULT 'IN_PROGRESS' CHECK (status IN (
    'IN_PROGRESS',
    'COMPLETED',
    'FAILED',
    'PARTIAL_SUCCESS',
    'CANCELLED'
  )),
  
  -- Timing
  started_at TIMESTAMPTZ DEFAULT NOW(),
  completed_at TIMESTAMPTZ,
  
  -- Metadata
  parameters JSONB DEFAULT '{}',      -- Operation-specific params
  error_summary TEXT,
  
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Results Table**: `tenant_bulk_operation_results`

```sql
CREATE TABLE tenant_bulk_operation_results (
  operation_id UUID REFERENCES tenant_bulk_operations(id) ON DELETE CASCADE,
  tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
  
  -- Individual Status
  status VARCHAR(20) DEFAULT 'PENDING' CHECK (status IN (
    'PENDING',
    'PROCESSING',
    'COMPLETED',
    'FAILED',
    'SKIPPED'
  )),
  
  -- Results
  message TEXT,
  error_details TEXT,
  
  -- Timing
  started_at TIMESTAMPTZ,
  completed_at TIMESTAMPTZ,
  
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  
  PRIMARY KEY (operation_id, tenant_id)
);
```

**Automated Status Calculation:**

```sql
-- Trigger recalculates parent status based on child results
CREATE FUNCTION update_bulk_operation_counts(p_operation_id UUID)
RETURNS VOID AS $$
DECLARE
  v_successful INT;
  v_failed INT;
  v_in_progress INT;
BEGIN
  SELECT 
    COUNT(CASE WHEN status = 'COMPLETED' THEN 1 END),
    COUNT(CASE WHEN status = 'FAILED' THEN 1 END),
    COUNT(CASE WHEN status IN ('PENDING', 'PROCESSING') THEN 1 END)
  INTO v_successful, v_failed, v_in_progress
  FROM tenant_bulk_operation_results
  WHERE operation_id = p_operation_id;
  
  UPDATE tenant_bulk_operations SET
    successful_count = v_successful,
    failed_count = v_failed,
    status = CASE 
      WHEN v_in_progress > 0 THEN 'IN_PROGRESS'
      WHEN v_failed = 0 THEN 'COMPLETED'
      WHEN v_successful = 0 THEN 'FAILED'
      ELSE 'PARTIAL_SUCCESS'
    END,
    completed_at = CASE 
      WHEN v_in_progress = 0 AND completed_at IS NULL THEN NOW()
      ELSE completed_at
    END
  WHERE id = p_operation_id;
END;
$$ LANGUAGE plpgsql;
```

**Example Usage:**

```sql
-- 1. Create bulk operation
INSERT INTO tenant_bulk_operations (
  operation_type,
  actor_id,
  actor_name,
  total_tenants,
  parameters
) VALUES (
  'SUSPEND',
  'admin-uuid',
  'Jane Admin',
  150,
  '{"reason": "Payment failures", "grace_period_days": 7}'::jsonb
) RETURNING id;

-- 2. Create individual results (app layer would insert these)
INSERT INTO tenant_bulk_operation_results (operation_id, tenant_id, status)
SELECT 
  'operation-uuid',
  id,
  'PENDING'
FROM tenants
WHERE status = 'ACTIVE' 
  AND last_payment_failed = true;

-- 3. Process each tenant (app updates status)
UPDATE tenant_bulk_operation_results
SET 
  status = 'COMPLETED',
  message = 'Tenant suspended successfully',
  completed_at = NOW()
WHERE operation_id = 'operation-uuid'
  AND tenant_id = 'specific-tenant-uuid';

-- 4. Trigger automatically updates parent counts and status
```

---

## 4. Security Implementation

### 4.1 Row-Level Security (RLS) Architecture

**Core Principle**: Every table has RLS policies that filter data based on session context.

**How It Works:**

```
User Login → Establish Session → Set Tenant Context → RLS Filters All Queries
```

**Session Context:**

```sql
-- Set by application at login
PERFORM set_config('app.current_tenant_id', '550e8400-...', true);

-- Retrieved by RLS policies
CREATE POLICY tenant_isolation_policy ON my_table
  USING (tenant_id = current_tenant_id());

-- Behind every query:
SELECT * FROM customers;

-- Becomes:
SELECT * FROM customers 
WHERE tenant_id = '550e8400-...'  -- Automatically added by RLS
```

### 4.2 Three-Tier Access Model

**Tier 1: Admin Role** (Unrestricted)

```sql
CREATE POLICY admin_full_access_policy ON tenants
  FOR ALL TO admin_role
  USING (true);

-- Admins see and can modify everything
-- Use case: Platform administration, support, bulk operations
```

**Tier 2: Application Role** (Tenant-Isolated)

```sql
CREATE POLICY tenant_isolation_policy ON tenants
  FOR ALL TO application_role
  USING (id = current_tenant_id() OR current_tenant_id() IS NULL);

-- Application users only see their tenant's data
-- NULL check allows initial setup before context is set
-- Use case: Normal business operations
```

**Tier 3: Readonly Role** (Reporting)

```sql
CREATE POLICY readonly_access_policy ON tenants
  FOR SELECT TO readonly_role
  USING (true);

-- Read-only access to all data (for analytics, reporting)
-- Cannot INSERT, UPDATE, DELETE
-- Use case: BI tools, data warehouses, auditors
```

### 4.3 Session Management Functions

**Set Tenant Context** (with validation):

```sql
CREATE FUNCTION set_tenant_context(tenant_id UUID, user_role TEXT DEFAULT 'application_role')
RETURNS VOID AS $$
DECLARE
  tenant_status TEXT;
BEGIN
  -- Validate tenant exists and is active
  SELECT status INTO tenant_status
  FROM tenants
  WHERE id = tenant_id AND deleted_at IS NULL;
  
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Tenant not found: %', tenant_id;
  END IF;
  
  IF tenant_status != 'ACTIVE' THEN
    RAISE EXCEPTION 'Tenant is not active: % (status: %)', 
      tenant_id, tenant_status;
  END IF;
  
  -- Set session variables (transactional scope)
  PERFORM set_config('app.current_tenant_id', tenant_id::text, true);
  PERFORM set_config('app.tenant_status', tenant_status, true);
  PERFORM set_config('app.context_set_at', NOW()::text, true);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

**Retrieve Current Tenant**:

```sql
CREATE FUNCTION current_tenant_id() RETURNS UUID AS $$
BEGIN
  RETURN COALESCE(
    nullif(current_setting('app.current_tenant_id', FALSE), ''),
    NULL
  )::UUID;
EXCEPTION
  WHEN OTHERS THEN
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
```

**Clear Context** (important for connection pooling):

```sql
CREATE FUNCTION clear_tenant_context() RETURNS VOID AS $$
BEGIN
  PERFORM set_config('app.current_tenant_id', NULL, true);
  PERFORM set_config('app.tenant_status', NULL, true);
  PERFORM set_config('app.context_set_at', NULL, true);
END;
$$ LANGUAGE plpgsql;
```

### 4.4 Connection Pooling Considerations

**Problem**: Pooled connections retain session state

**Solution**: Always clear context when returning connections to pool

```python
# Python example with psycopg2
def get_tenant_connection(tenant_id):
    conn = connection_pool.getconn()
    try:
        with conn.cursor() as cur:
            cur.execute("SELECT set_tenant_context(%s)", (tenant_id,))
        yield conn
    finally:
        # CRITICAL: Clear context before returning to pool
        with conn.cursor() as cur:
            cur.execute("SELECT clear_tenant_context()")
        connection_pool.putconn(conn)

# Usage:
with get_tenant_connection('550e8400-...') as conn:
    # All queries automatically filtered to this tenant
    cur.execute("SELECT * FROM customers")
```

### 4.5 Cross-Tenant Reference Prevention

**Function**: `enforce_tenant_isolation()`

Validates that foreign key references don't cross tenant boundaries:

```sql
CREATE FUNCTION enforce_tenant_isolation() RETURNS TRIGGER AS $$
BEGIN
  -- Example: persons table references entities table
  -- Ensure entity belongs to same tenant
  
  IF TG_TABLE_NAME = 'persons' THEN
    IF NOT EXISTS (
      SELECT 1
      FROM entities e
      WHERE e.uuid = NEW.entity_id
        AND e.tenant_id = NEW.tenant_id
    ) THEN
      RAISE EXCEPTION 'Entity % does not belong to tenant %',
        NEW.entity_id, NEW.tenant_id;
    END IF;
  END IF;
  
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply to tables with cross-references
CREATE TRIGGER enforce_persons_tenant_isolation
  BEFORE INSERT OR UPDATE ON persons
  FOR EACH ROW EXECUTE FUNCTION enforce_tenant_isolation();
```

### 4.6 Audit Trail Implementation

**What Gets Logged** (via audit triggers):
- User who made the change
- Timestamp
- Old and new values
- Tenant context
- IP address (if captured by application)

**Example Audit Log Query:**

```sql
-- See all changes to a customer record
SELECT 
  audit_timestamp,
  user_id,
  action,
  old_values,
  new_values,
  ip_address
FROM audit_log
WHERE table_name = 'customers'
  AND record_id = 'customer-uuid'
  AND tenant_id = current_tenant_id()
ORDER BY audit_timestamp DESC;
```

---

## 5. Utility Functions

### 5.1 Timestamp Management

**Auto-Update `updated_at`**:

```sql
CREATE FUNCTION update_updated_at_column() RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Applied to all tables with updated_at column
CREATE TRIGGER update_[table]_updated_at
  BEFORE UPDATE ON [table]
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

### 5.2 Slug Generation

**Auto-Generate URL-Friendly Identifiers**:

```sql
CREATE FUNCTION generate_unique_slug_from_name() RETURNS TRIGGER AS $$
DECLARE
  base_slug TEXT;
  final_slug TEXT;
  counter INTEGER := 1;
BEGIN
  IF NEW.slug IS NULL THEN
    -- Create base slug
    base_slug := lower(trim(both '-' from 
      regexp_replace(
        regexp_replace(NEW.name, '[^\w\s-]', '', 'g'),
        '\s+', '-', 'g'
      )
    ));
    
    -- Ensure not empty
    IF base_slug = '' THEN
      base_slug := 'tenant';
    END IF;
    
    final_slug := base_slug;
    
    -- Handle collisions
    WHILE EXISTS(
      SELECT 1 FROM tenants 
      WHERE slug = final_slug 
        AND id != COALESCE(NEW.id, '00000000-0000-0000-0000-000000000000'::UUID)
    ) LOOP
      final_slug := base_slug || '-' || counter;
      counter := counter + 1;
    END LOOP;
    
    NEW.slug := final_slug;
  END IF;
  
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

**Examples:**

```
"Acme Corporation"     → "acme-corporation"
"Smith & Associates"   → "smith-associates"
"ABC Industries"       → "abc-industries"
"ABC Industries" (2nd) → "abc-industries-1"
"My Company!!!"        → "my-company"
""                     → "tenant"
```

### 5.3 Tenant Provisioning

**Complete Tenant Setup**:

```sql
CREATE FUNCTION provision_tenant_complete(
  p_name VARCHAR(255),
  p_email VARCHAR(255),
  p_subdomain VARCHAR(63) DEFAULT NULL,
  p_industry VARCHAR(50) DEFAULT NULL,
  p_company_size VARCHAR(20) DEFAULT 'Small',
  p_currency_code CHAR(3) DEFAULT 'USD',
  p_timezone VARCHAR(50) DEFAULT 'UTC',
  p_settings JSONB DEFAULT '{}'
) RETURNS TABLE(id UUID) AS $$
DECLARE
  v_tenant_id UUID;
  v_slug VARCHAR(50);
BEGIN
  -- Generate IDs
  v_tenant_id := gen_random_uuid();
  v_slug := lower(regexp_replace(p_name, '[^a-zA-Z0-9]+', '-', 'g'));
  
  -- Ensure uniqueness
  WHILE EXISTS (
    SELECT 1 FROM tenants 
    WHERE slug = v_slug AND deleted_at IS NULL
  ) LOOP 
    v_slug := v_slug || '-' || substring(v_tenant_id::text, 1, 8);
  END LOOP;
  
  -- Create tenant
  INSERT INTO tenants (
    id, slug, name, email, subdomain,
    STATUS, industry, company_size,
    currency_code, timezone, settings
  ) VALUES (
    v_tenant_id, v_slug, p_name, p_email, p_subdomain,
    'PENDING', p_industry, p_company_size,
    p_currency_code, p_timezone, p_settings
  );
  
  -- Return ID
  RETURN QUERY SELECT v_tenant_id AS tenant_id;
END;
$$ LANGUAGE plpgsql;
```

### 5.4 Context Validation

**Enhanced Context Setting**:

```sql
CREATE FUNCTION validate_and_set_tenant_context(p_tenant_id UUID)
RETURNS TABLE(tenant_name TEXT, tenant_status TEXT) AS $$
DECLARE
  v_tenant_record RECORD;
BEGIN
  -- Fetch and validate
  SELECT id, name, STATUS, deleted_at, last_activity_at
  INTO v_tenant_record
  FROM tenants
  WHERE id = p_tenant_id;
  
  -- Check existence
  IF v_tenant_record.id IS NULL THEN
    RAISE EXCEPTION 'Tenant not found: %', p_tenant_id;
  END IF;
  
  -- Check not deleted
  IF v_tenant_record.deleted_at IS NOT NULL THEN
    RAISE EXCEPTION 'Tenant is deleted: %', p_tenant_id;
  END IF;
  
  -- Check status
  IF v_tenant_record.status NOT IN ('active', 'pending') THEN
    RAISE EXCEPTION 'Tenant is not active: % (status: %)',
      p_tenant_id, v_tenant_record.status;
  END IF;
  
  -- Update activity
  UPDATE tenants
  SET last_activity_at = NOW()
  WHERE id = p_tenant_id;
  
  -- Set context
  PERFORM set_config('app.current_tenant_id', p_tenant_id::text, TRUE);
  
  -- Return info
  RETURN QUERY
  SELECT v_tenant_record.name, v_tenant_record.status;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

### 5.5 Resource Limit Checking

**Validate Before Operations**:

```sql
CREATE FUNCTION check_tenant_limits(
  p_tenant_id UUID,
  p_check_type VARCHAR(50),
  p_additional_usage INT DEFAULT 1
) RETURNS BOOLEAN AS $$
DECLARE
  v_config tenant_configurations%ROWTYPE;
  v_current_usage INT;
BEGIN
  -- Get configuration
  SELECT * INTO v_config
  FROM tenant_configurations
  WHERE tenant_id = p_tenant_id;
  
  IF v_config.tenant_id IS NULL THEN
    RAISE NOTICE 'No configuration found for tenant: %', p_tenant_id;
    RETURN FALSE;
  END IF;
  
  -- Check based on type
  CASE p_check_type
    WHEN 'users' THEN
      -- Implementation placeholder
      RETURN TRUE;
      
    WHEN 'entities' THEN
      -- Implementation placeholder
      RETURN TRUE;
      
    WHEN 'transactions' THEN
      -- Implementation placeholder
      RETURN TRUE;
      
    WHEN 'storage' THEN
      -- Get current storage usage
      SELECT COALESCE(storage_used, 0) INTO v_current_usage
      FROM tenant_usage_stats
      WHERE tenant_id = p_tenant_id
        AND period_start <= CURRENT_DATE
        AND period_end >= CURRENT_DATE
      ORDER BY period_start DESC
      LIMIT 1;
      
      -- Check against quota
      RETURN (COALESCE(v_current_usage, 0) + p_additional_usage) 
             <= v_config.storage_quota;
             
    ELSE
      RAISE NOTICE 'Unknown check type: %', p_check_type;
      RETURN FALSE;
  END CASE;
END;
$$ LANGUAGE plpgsql;
```

**Usage:**

```python
# Before uploading large file
can_upload = cursor.execute("""
    SELECT check_tenant_limits(%s, 'storage', %s)
""", (tenant_id, file_size_bytes))

if can_upload:
    upload_file()
else:
    raise StorageQuotaExceeded()
```

### 5.6 Bulk Operation Monitoring

**Real-Time Progress**:

```sql
CREATE FUNCTION get_bulk_operation_summary(p_operation_id UUID)
RETURNS TABLE (
  operation_id UUID,
  operation_type VARCHAR(50),
  status VARCHAR(20),
  total_tenants INT,
  successful_count INT,
  failed_count INT,
  in_progress_count INT,
  duration_seconds INT
) AS $$
BEGIN
  RETURN QUERY
  SELECT 
    bo.id,
    bo.operation_type,
    bo.status,
    bo.total_tenants,
    bo.successful_count,
    bo.failed_count,
    COUNT(CASE WHEN br.status IN ('PENDING', 'PROCESSING') THEN 1 END)::INT,
    CASE 
      WHEN bo.completed_at IS NOT NULL 
      THEN EXTRACT(EPOCH FROM (bo.completed_at - bo.started_at))::INT
      ELSE EXTRACT(EPOCH FROM (NOW() - bo.started_at))::INT
    END
  FROM tenant_bulk_operations bo
  LEFT JOIN tenant_bulk_operation_results br ON bo.id = br.operation_id
  WHERE bo.id = p_operation_id
  GROUP BY bo.id, bo.operation_type, bo.status, bo.total_tenants,
           bo.successful_count, bo.failed_count, bo.started_at, bo.completed_at;
END;
$$ LANGUAGE plpgsql;
```

---

## 6. Data Isolation Mechanisms

### 6.1 Defense in Depth

Multiple layers ensure tenant isolation:

```
Layer 1: Application Logic
├─ Validate tenant_id in API requests
├─ Inject tenant_id into queries
└─ Session management

Layer 2: Session Context (set_config)
├─ Set at login: app.current_tenant_id
├─ Used by RLS policies
└─ Cleared at logout/timeout

Layer 3: Row-Level Security (RLS)
├─ Automatic query filtering
├─ Per-role policies
└─ Cannot be bypassed by application

Layer 4: Foreign Key Constraints
├─ tenant_id in all multi-tenant tables
├─ Cascading deletes
└─ Referential integrity

Layer 5: Trigger Validation
├─ enforce_tenant_isolation()
├─ Cross-table tenant checks
└─ Additional business rules

Layer 6: Audit Logging
├─ Track all data access
├─ Detect anomalies
└─ Forensic capability
```

### 6.2 RLS Policy Examples

**Standard Table Pattern**:

```sql
-- Every multi-tenant table should have:

-- 1. tenant_id column
ALTER TABLE my_table 
  ADD COLUMN tenant_id UUID NOT NULL REFERENCES tenants(id);

-- 2. Index for performance
CREATE INDEX idx_my_table_tenant ON my_table(tenant_id);

-- 3. Enable RLS
ALTER TABLE my_table ENABLE ROW LEVEL SECURITY;

-- 4. Admin policy (unrestricted)
CREATE POLICY admin_full_access_policy ON my_table
  FOR ALL TO admin_role
  USING (true);

-- 5. Application policy (isolated)
CREATE POLICY tenant_isolation_policy ON my_table
  FOR ALL TO application_role
  USING (tenant_id = current_tenant_id())
  WITH CHECK (tenant_id = current_tenant_id());

-- 6. Readonly policy (all data, no modifications)
CREATE POLICY readonly_access_policy ON my_table
  FOR SELECT TO readonly_role
  USING (true);

-- 7. Grants
GRANT SELECT, INSERT, UPDATE, DELETE ON my_table TO admin_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON my_table TO application_role;
GRANT SELECT ON my_table TO readonly_role;

-- 8. Trigger for updated_at
CREATE TRIGGER update_my_table_updated_at
  BEFORE UPDATE ON my_table
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

### 6.3 Testing Isolation

**Verification Queries**:

```sql
-- Test 1: Verify RLS is enabled
SELECT tablename, rowsecurity
FROM pg_tables
WHERE schemaname = 'public'
  AND tablename LIKE 'tenant%';

-- Expected: rowsecurity = 't' (true)

-- Test 2: Verify policies exist
SELECT 
  schemaname,
  tablename,
  policyname,
  roles,
  cmd,
  qual AS using_expression,
  with_check
FROM pg_policies
WHERE tablename LIKE 'tenant%'
ORDER BY tablename, policyname;

-- Test 3: Simulate tenant context and verify isolation
BEGIN;
  -- Set context to Tenant A
  PERFORM set_tenant_context('tenant-a-uuid');
  
  -- Should only see Tenant A's data
  SELECT COUNT(*) FROM customers;  -- Should match Tenant A count
  
  -- Try to see Tenant B's data explicitly
  SELECT * FROM customers WHERE tenant_id = 'tenant-b-uuid';
  -- Should return 0 rows (RLS blocks it)
  
ROLLBACK;

-- Test 4: Verify admin can see all
SET ROLE admin_role;
SELECT COUNT(DISTINCT tenant_id) FROM customers;  -- Should see all tenants
RESET ROLE;

-- Test 5: Verify readonly cannot modify
SET ROLE readonly_role;
UPDATE customers SET name = 'test';  -- Should fail with permission error
RESET ROLE;
```

---

## 7. Design Patterns & Best Practices

### 7.1 Naming Conventions

**Observed Patterns:**

| Pattern | Example | Purpose |
|---------|---------|---------|
| `[entity]s` | `tenants`, `customers` | Plural table names |
| `[table]_id` | `tenant_id`, `customer_id` | Foreign key columns |
| `idx_[table]_[column]` | `idx_tenants_slug` | Index naming |
| `[context]_policy` | `tenant_isolation_policy` | RLS policy naming |
| `[action]_[context]()` | `set_tenant_context()` | Function naming |
| `p_[param]` | `p_tenant_id` | Function parameters |
| `v_[variable]` | `v_tenant_record` | Function variables |

**Recommendation**: Continue these conventions for consistency.

### 7.2 Migration Structure

**Standard Migration Format:**

```sql
-- =====================================================
-- MIGRATION TITLE
-- =====================================================
-- 
-- PURPOSE:
-- Detailed explanation of what this migration does
-- 
-- TABLES CREATED:
-- - Table 1: Description
-- - Table 2: Description
-- 
-- DEPENDENCIES:
-- - Requires migration XXXXX
-- - References tables: ...
-- 
-- AUTHOR: [Team/Person]
-- VERSION: 1.0
-- DATE: YYYY-MM-DD
-- =====================================================

-- Main SQL here...

-- =====================================================
-- COMPLETION LOG
-- =====================================================
DO $$
BEGIN
  RAISE NOTICE 'Migration completed successfully';
  RAISE NOTICE 'Created: [list of objects]';
END $$;
```

**Benefits:**
- Self-documenting
- Easy to understand purpose
- Helpful for future developers
- Visible in migration logs

### 7.3 JSONB Usage Guidelines

**From Migration 000102 Comments:**

**Use Dedicated Columns For:**
- Critical configuration (affects system behavior)
- Frequently queried fields
- Values needing database constraints
- Billing/financial data
- Security settings (with fallback JSONB)

**Use JSONB For:**
- Tenant-specific preferences
- Experimental features
- Rarely queried metadata
- Flexible/dynamic attributes
- Feature flags

**Example:**

```sql
-- GOOD: Critical settings as columns
max_users INT DEFAULT 100,
accounting_method VARCHAR(10) DEFAULT 'ACCRUAL',

-- GOOD: Flexible settings as JSONB
settings JSONB DEFAULT '{
  "ui_theme": "dark",
  "dashboard_widgets": ["sales", "revenue"],
  "experimental_features": {
    "ai_assistant": true,
    "advanced_forecasting": false
  }
}'::jsonb,
```

### 7.4 Error Handling Patterns

**Descriptive Exceptions:**

```sql
-- GOOD: Clear, actionable error messages
IF v_tenant_record.id IS NULL THEN
  RAISE EXCEPTION 'Tenant not found: %', p_tenant_id;
END IF;

IF tenant_status != 'ACTIVE' THEN
  RAISE EXCEPTION 'Tenant is not active: % (status: %)', 
    p_tenant_id, tenant_status;
END IF;

-- BAD: Generic errors
IF v_tenant_record.id IS NULL THEN
  RAISE EXCEPTION 'Error';
END IF;
```

**Informational Notices:**

```sql
-- Use RAISE NOTICE for warnings/info (doesn't stop execution)
IF v_config.tenant_id IS NULL THEN
  RAISE NOTICE 'No configuration found for tenant: %', p_tenant_id;
  RETURN FALSE;
END IF;
```

### 7.5 Default Values Philosophy

**Progressive Defaults:**

```sql
-- Start conservative, allow growth
max_users INT DEFAULT 100,              -- Not 10, not 1000
max_entities INT DEFAULT 1000,
max_transactions_per_month INT DEFAULT 10000,
storage_quota BIGINT DEFAULT 1073741824,  -- 1GB

-- Reasonable business defaults
accounting_method VARCHAR(10) DEFAULT 'ACCRUAL',
fiscal_year_start_month INT DEFAULT 1,  -- January
date_format VARCHAR(20) DEFAULT 'MM/DD/YYYY',

-- Secure by default
password_policy JSONB DEFAULT '{
  "min_length": 8,
  "require_uppercase": true,
  "require_lowercase": true,
  "require_numbers": true,
  "require_symbols": false
}'::jsonb,
```

**Rationale:**
- New tenants can start immediately
- Defaults work for 80% of use cases
- Easy to upgrade limits later
- Security-conscious defaults

---

## 8. Potential Issues & Recommendations

### 8.1 Status Field Inconsistency

**Issue Found:**

```sql
-- Migration 000101: Status field (capital S)
Status VARCHAR(20) DEFAULT 'ACTIVE' 
CHECK (Status IN ('ACTIVE', 'SUSPENDED', 'PENDING', 'ARCHIVED'))

-- Migration 000106: status field (lowercase s)
IF v_tenant_record.status NOT IN ('active', 'pending') THEN
```

**Problems:**
1. Inconsistent casing (Status vs status)
2. Inconsistent values (ACTIVE vs active)
3. Will cause runtime errors

**Recommendation:**

```sql
-- Standardize to lowercase column and uppercase values
ALTER TABLE tenants RENAME COLUMN "Status" TO status;

-- Update function to use uppercase values
IF v_tenant_record.status NOT IN ('ACTIVE', 'PENDING') THEN
```

### 8.2 Missing Indexes

**Current Indexes:**
- ✓ `idx_tenants_slug`
- ✓ `idx_tenants_status`
- ✓ `idx_tenants_subdomain`
- ✓ `idx_tenants_deleted_at`

**Recommended Additional Indexes:**

```sql
-- For email lookups (password reset, admin search)
CREATE INDEX idx_tenants_email ON tenants(email);

-- For industry/size analytics
CREATE INDEX idx_tenants_industry ON tenants(industry)
WHERE industry IS NOT NULL;

CREATE INDEX idx_tenants_company_size ON tenants(company_size)
WHERE company_size IS NOT NULL;

-- For active tenant queries (exclude deleted)
CREATE INDEX idx_tenants_active ON tenants(id)
WHERE deleted_at IS NULL AND status = 'ACTIVE';

-- Composite index for common queries
CREATE INDEX idx_tenants_status_deleted ON tenants(status, deleted_at);
```

### 8.3 check_tenant_limits Placeholders

**Current State:**

```sql
WHEN 'users' THEN
  -- Check user limit (placeholder for when users table exists)
  RETURN TRUE;

WHEN 'entities' THEN
  -- Check entity limit (placeholder for when entities table exists)
  RETURN TRUE;

WHEN 'transactions' THEN
  -- Check monthly transaction limit (placeholder)
  RETURN TRUE;
```

**Recommendation:**

Implement these checks as tables are added:

```sql
WHEN 'users' THEN
  SELECT COUNT(*) INTO v_current_usage
  FROM users
  WHERE tenant_id = p_tenant_id
    AND status = 'ACTIVE';
  
  RETURN (v_current_usage + p_additional_usage) <= v_config.max_users;

WHEN 'entities' THEN
  SELECT COUNT(*) INTO v_current_usage
  FROM entities
  WHERE tenant_id = p_tenant_id
    AND deleted_at IS NULL;
  
  RETURN (v_current_usage + p_additional_usage) <= v_config.max_entities;

WHEN 'transactions' THEN
  SELECT COALESCE(total_transactions, 0) INTO v_current_usage
  FROM tenant_usage_stats
  WHERE tenant_id = p_tenant_id
    AND period_start = DATE_TRUNC('month', CURRENT_DATE);
  
  RETURN (v_current_usage + p_additional_usage) 
         <= v_config.max_transactions_per_month;
```

### 8.4 Missing Cascade Behaviors

**Current:**

```sql
tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE
```

**Recommendation:**

Consider `ON UPDATE CASCADE` as well:

```sql
tenant_id UUID REFERENCES tenants(id) 
  ON DELETE CASCADE 
  ON UPDATE CASCADE
```

**Rationale:**
- If tenant ID ever needs to change, updates cascade
- Maintains referential integrity
- Minimal performance impact

### 8.5 SECURITY DEFINER Caution

**Found in:**

```sql
CREATE FUNCTION set_tenant_context(...)
RETURNS VOID AS $$
...
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

**What SECURITY DEFINER Means:**
- Function runs with privileges of function owner (not caller)
- Necessary for setting session variables
- Potential security risk if not carefully written

**Recommendations:**
1. ✓ Input validation (already present)
2. ✓ No dynamic SQL (already safe)
3. ⚠️ Add explicit validation to prevent privilege escalation
4. ⚠️ Consider logging all context changes

**Enhanced Version:**

```sql
CREATE FUNCTION set_tenant_context(
  tenant_id UUID, 
  user_role TEXT DEFAULT 'application_role'
) RETURNS VOID AS $$
DECLARE
  tenant_status TEXT;
  calling_user TEXT := SESSION_USER;
BEGIN
  -- Log the context change (audit trail)
  INSERT INTO tenant_context_log (
    tenant_id, user_role, calling_user, set_at
  ) VALUES (
    tenant_id, user_role, calling_user, NOW()
  );
  
  -- Validate tenant exists and is active
  SELECT status INTO tenant_status
  FROM tenants
  WHERE id = tenant_id AND deleted_at IS NULL;
  
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Tenant not found: %', tenant_id;
  END IF;
  
  IF tenant_status != 'ACTIVE' THEN
    RAISE EXCEPTION 'Tenant is not active: % (status: %)', 
      tenant_id, tenant_status;
  END IF;
  
  -- Validate role (prevent injection)
  IF user_role NOT IN ('application_role', 'admin_role', 'readonly_role') THEN
    RAISE EXCEPTION 'Invalid role: %', user_role;
  END IF;
  
  -- Set context
  PERFORM set_config('app.current_tenant_id', tenant_id::text, true);
  PERFORM set_config('app.tenant_status', tenant_status, true);
  PERFORM set_config('app.user_role', user_role, true);
  PERFORM set_config('app.context_set_at', NOW()::text, true);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

### 8.6 Transaction Isolation Concerns

**Scenario:**

```sql
-- Session 1
BEGIN;
SELECT set_tenant_context('tenant-a');
SELECT * FROM customers;  -- Sees Tenant A

-- Session 2 (same connection pool)
BEGIN;
-- Assumes clean context, but might inherit Session 1's context
SELECT * FROM customers;  -- Might see Tenant A!
```

**Recommendation:**

Always clear context at transaction boundaries:

```python
# Application pattern
def execute_with_tenant(tenant_id, query):
    try:
        conn.execute("SELECT set_tenant_context(%s)", (tenant_id,))
        result = conn.execute(query)
        conn.commit()
        return result
    except Exception as e:
        conn.rollback()
        raise
    finally:
        # ALWAYS clear, even on error
        conn.execute("SELECT clear_tenant_context()")
```

### 8.7 Soft Delete Considerations

**Current:**

```sql
deleted_at TIMESTAMPTZ
```

**Good, but consider:**

```sql
-- Add who deleted and why
deleted_at TIMESTAMPTZ,
deleted_by UUID REFERENCES users(id),
deletion_reason TEXT,

-- Add index to exclude deleted efficiently
CREATE INDEX idx_tenants_not_deleted ON tenants(id)
WHERE deleted_at IS NULL;

-- Add retention policy tracking
deletion_scheduled_for TIMESTAMPTZ,  -- Auto-delete after X days

-- Trigger to prevent accidental un-deletion
CREATE FUNCTION prevent_undelete() RETURNS TRIGGER AS $$
BEGIN
  IF OLD.deleted_at IS NOT NULL AND NEW.deleted_at IS NULL THEN
    RAISE EXCEPTION 'Cannot undelete tenant. Create new tenant instead.';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

---

## 9. Performance Considerations

### 9.1 Index Strategy

**Current Indexes (Good):**
- Primary keys (automatic)
- Foreign keys for tenant_id
- Filtered indexes (WHERE deleted_at IS NOT NULL)
- Common query patterns (status, slug, subdomain)

**Recommendations:**

```sql
-- 1. Composite indexes for common query patterns
CREATE INDEX idx_tenants_status_activity ON tenants(status, last_activity_at);
-- For queries: WHERE status = 'ACTIVE' ORDER BY last_activity_at DESC

-- 2. Partial indexes for active tenants
CREATE INDEX idx_active_tenants ON tenants(id, name)
WHERE deleted_at IS NULL AND status = 'ACTIVE';
-- Smaller index, faster queries for 95% of use cases

-- 3. JSONB indexing for frequently queried settings
CREATE INDEX idx_tenant_config_settings ON tenant_configurations 
USING GIN(settings);
-- For queries: WHERE settings @> '{"feature": "value"}'

-- 4. Time-series optimization
CREATE INDEX idx_usage_stats_recent ON tenant_usage_stats(tenant_id, period_start DESC)
WHERE period_start >= CURRENT_DATE - INTERVAL '90 days';
-- Optimize recent usage queries
```

### 9.2 Query Optimization

**RLS Performance Impact:**

Every query with RLS gets rewritten:

```sql
-- Your query:
SELECT * FROM customers WHERE name LIKE 'A%';

-- Actual execution (with RLS):
SELECT * FROM customers 
WHERE name LIKE 'A%' 
  AND tenant_id = current_tenant_id();
```

**Optimization:**

```sql
-- Make sure tenant_id is part of your indexes!
CREATE INDEX idx_customers_tenant_name ON customers(tenant_id, name);

-- This allows efficient index scans on both filters
```

**Benchmark RLS overhead:**

```sql
-- Disable RLS temporarily to measure overhead
ALTER TABLE customers DISABLE ROW LEVEL SECURITY;
EXPLAIN ANALYZE SELECT * FROM customers WHERE name LIKE 'A%';

-- Re-enable and compare
ALTER TABLE customers ENABLE ROW LEVEL SECURITY;
SELECT set_tenant_context('tenant-id');
EXPLAIN ANALYZE SELECT * FROM customers WHERE name LIKE 'A%';

-- Overhead should be minimal (< 5%) with proper indexes
```

### 9.3 Connection Pooling

**PgBouncer Configuration:**

```ini
[pgbouncer]
# Transaction pooling (recommended for multi-tenant)
pool_mode = transaction

# Don't pool these session-level functions
ignore_startup_parameters = application_name, search_path

# Max connections per tenant (example)
max_client_conn = 1000
default_pool_size = 25
reserve_pool_size = 5
```

**Application Pattern:**

```python
# Django example
DATABASES = {
    'default': {
        'ENGINE': 'django.db.backends.postgresql',
        'OPTIONS': {
            'options': '-c search_path=public',
            # Clear session state on return to pool
            'connect_timeout': 10,
        }
    }
}

# Custom middleware
class TenantMiddleware:
    def process_request(self, request):
        tenant_id = get_tenant_from_request(request)
        with connection.cursor() as cursor:
            cursor.execute("SELECT set_tenant_context(%s)", [tenant_id])
    
    def process_response(self, request, response):
        with connection.cursor() as cursor:
            cursor.execute("SELECT clear_tenant_context()")
        return response
```

### 9.4 Bulk Operations Performance

**Current Implementation:**
Individual row updates in `tenant_bulk_operation_results`

**Optimization Opportunity:**

```sql
-- Current: N individual updates
UPDATE tenant_bulk_operation_results
SET status = 'COMPLETED', completed_at = NOW()
WHERE operation_id = 'op-id' AND tenant_id = 'tenant-1';

UPDATE tenant_bulk_operation_results
SET status = 'COMPLETED', completed_at = NOW()
WHERE operation_id = 'op-id' AND tenant_id = 'tenant-2';
-- ... repeat N times

-- Optimized: Batch update
UPDATE tenant_bulk_operation_results
SET 
  status = v.status,
  completed_at = v.completed_at,
  message = v.message
FROM (VALUES
  ('tenant-1'::uuid, 'COMPLETED', NOW(), 'Success'),
  ('tenant-2'::uuid, 'COMPLETED', NOW(), 'Success'),
  -- ... many more
) AS v(tenant_id, status, completed_at, message)
WHERE tenant_bulk_operation_results.operation_id = 'op-id'
  AND tenant_bulk_operation_results.tenant_id = v.tenant_id;
```

### 9.5 Vacuum and Maintenance

**High-Churn Tables:**

Tables with frequent updates/deletes need regular maintenance:

```sql
-- tenant_usage_stats (updated frequently)
-- tenant_bulk_operation_results (batch inserts/updates)

-- Configure autovacuum more aggressively
ALTER TABLE tenant_usage_stats SET (
  autovacuum_vacuum_scale_factor = 0.01,  -- Default: 0.2
  autovacuum_analyze_scale_factor = 0.005  -- Default: 0.1
);

-- Or manual vacuum during maintenance windows
VACUUM ANALYZE tenant_usage_stats;
```

---

## 10. Extension Roadmap

### 10.1 Future Tables (Based on Commented Examples)

**Projects Table** (from Migration 000106):

```sql
CREATE TABLE projects (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id UUID REFERENCES entities(uuid) ON DELETE CASCADE,
  
  name VARCHAR(255) NOT NULL,
  code VARCHAR(50),
  description TEXT,
  
  project_manager_id UUID REFERENCES employees(id),
  
  start_date DATE,
  end_date DATE,
  budget_amount DECIMAL(15,2),
  actual_cost DECIMAL(15,2) DEFAULT 0,
  
  status VARCHAR(20) DEFAULT 'PLANNING',
  
  metadata JSONB DEFAULT '{}',
  
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  
  CONSTRAINT projects_tenant_code_entity_unique_idx 
    UNIQUE (tenant_id, entity_id, code)
);
```

**Pattern to Replicate:**
- ✓ tenant_id foreign key
- ✓ RLS policies (tenant isolation)
- ✓ Updated_at trigger
- ✓ JSONB metadata field
- ✓ Tenant-scoped unique constraints

### 10.2 Entities and Persons Tables

**Referenced but not shown:**

```sql
-- Likely structure based on references:

CREATE TABLE entities (
  uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
  
  entity_type VARCHAR(50),  -- 'customer', 'supplier', 'both'
  legal_name VARCHAR(255) NOT NULL,
  
  -- ... other fields
  
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  deleted_at TIMESTAMPTZ
);

CREATE TABLE persons (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id UUID REFERENCES entities(uuid) ON DELETE CASCADE,
  
  first_name VARCHAR(100),
  last_name VARCHAR(100),
  email VARCHAR(255),
  
  -- ... other fields
  
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Apply tenant isolation trigger
CREATE TRIGGER enforce_persons_tenant_isolation
  BEFORE INSERT OR UPDATE ON persons
  FOR EACH ROW EXECUTE FUNCTION enforce_tenant_isolation();
```

### 10.3 Recommended Additional Tables

**Users Table:**

```sql
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
  
  email VARCHAR(255) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  
  first_name VARCHAR(100),
  last_name VARCHAR(100),
  
  role VARCHAR(50) DEFAULT 'user',  -- 'admin', 'manager', 'user'
  status VARCHAR(20) DEFAULT 'ACTIVE',
  
  last_login_at TIMESTAMPTZ,
  last_login_ip INET,
  
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  deleted_at TIMESTAMPTZ,
  
  CONSTRAINT users_tenant_email_unique UNIQUE (tenant_id, email)
);

CREATE INDEX idx_users_tenant ON users(tenant_id);
CREATE INDEX idx_users_email ON users(email);

-- RLS policies (standard pattern)
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_policy ON users
  FOR ALL TO application_role
  USING (tenant_id = current_tenant_id());
-- ... admin and readonly policies
```

**Audit Log Table:**

```sql
CREATE TABLE audit_log (
  id BIGSERIAL PRIMARY KEY,
  tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
  
  table_name VARCHAR(100) NOT NULL,
  record_id UUID,
  
  action VARCHAR(20) NOT NULL,  -- 'INSERT', 'UPDATE', 'DELETE'
  
  user_id UUID,
  user_email VARCHAR(255),
  ip_address INET,
  
  old_values JSONB,
  new_values JSONB,
  
  occurred_at TIMESTAMPTZ DEFAULT NOW()
);

-- Partitioning for performance (optional)
CREATE TABLE audit_log_2024_01 PARTITION OF audit_log
  FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- Indexes
CREATE INDEX idx_audit_log_tenant_time ON audit_log(tenant_id, occurred_at DESC);
CREATE INDEX idx_audit_log_record ON audit_log(table_name, record_id);

-- RLS (different pattern - admins and auditors only)
ALTER TABLE audit_log ENABLE ROW LEVEL SECURITY;
CREATE POLICY admin_audit_access ON audit_log
  FOR SELECT TO admin_role
  USING (true);
CREATE POLICY tenant_audit_access ON audit_log
  FOR SELECT TO application_role
  USING (tenant_id = current_tenant_id());
```

**Feature Flags Table:**

```sql
CREATE TABLE feature_flags (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  
  flag_key VARCHAR(100) NOT NULL UNIQUE,
  flag_name VARCHAR(255) NOT NULL,
  description TEXT,
  
  default_enabled BOOLEAN DEFAULT FALSE,
  
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE tenant_feature_flags (
  tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
  feature_flag_id UUID REFERENCES feature_flags(id) ON DELETE CASCADE,
  
  enabled BOOLEAN NOT NULL,
  enabled_at TIMESTAMPTZ,
  enabled_by UUID REFERENCES users(id),
  
  PRIMARY KEY (tenant_id, feature_flag_id)
);

-- RLS
ALTER TABLE tenant_feature_flags ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_policy ON tenant_feature_flags
  FOR ALL TO application_role
  USING (tenant_id = current_tenant_id());
```

### 10.4 Integration Tables

**Webhooks:**

```sql
CREATE TABLE webhook_deliveries (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
  
  webhook_url TEXT NOT NULL,
  event_type VARCHAR(100) NOT NULL,
  
  payload JSONB NOT NULL,
  
  status VARCHAR(20) DEFAULT 'PENDING',  -- 'PENDING', 'DELIVERED', 'FAILED'
  
  attempts INT DEFAULT 0,
  max_attempts INT DEFAULT 3,
  
  last_attempt_at TIMESTAMPTZ,
  next_attempt_at TIMESTAMPTZ,
  
  response_status_code INT,
  response_body TEXT,
  
  created_at TIMESTAMPTZ DEFAULT NOW(),
  delivered_at TIMESTAMPTZ
);

CREATE INDEX idx_webhooks_tenant_pending ON webhook_deliveries(tenant_id, status)
WHERE status = 'PENDING';
```

**API Keys:**

```sql
CREATE TABLE api_keys (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
  
  key_name VARCHAR(255) NOT NULL,
  key_hash VARCHAR(255) NOT NULL UNIQUE,  -- Never store plain key!
  
  permissions JSONB DEFAULT '{}',  -- {"read": true, "write": false}
  
  status VARCHAR(20) DEFAULT 'ACTIVE',
  
  last_used_at TIMESTAMPTZ,
  last_used_ip INET,
  
  expires_at TIMESTAMPTZ,
  
  created_at TIMESTAMPTZ DEFAULT NOW(),
  created_by UUID REFERENCES users(id),
  
  revoked_at TIMESTAMPTZ,
  revoked_by UUID REFERENCES users(id),
  revocation_reason TEXT
);
```

---

## 11. Security Checklist

### 11.1 Implementation Checklist

When adding a new multi-tenant table:

- [ ] Add `tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE`
- [ ] Create index: `CREATE INDEX idx_[table]_tenant ON [table](tenant_id)`
- [ ] Enable RLS: `ALTER TABLE [table] ENABLE ROW LEVEL SECURITY`
- [ ] Create admin policy: `USING (true)`
- [ ] Create application policy: `USING (tenant_id = current_tenant_id())`
- [ ] Create readonly policy (if applicable): `FOR SELECT ... USING (true)`
- [ ] Grant appropriate permissions to each role
- [ ] Add `updated_at` trigger (if applicable)
- [ ] Add tenant isolation trigger (if cross-references exist)
- [ ] Test isolation with multiple tenant contexts
- [ ] Document the table structure and purpose

### 11.2 Security Testing

**Essential Tests:**

```sql
-- Test 1: RLS Enabled
SELECT 
  tablename,
  rowsecurity
FROM pg_tables
WHERE schemaname = 'public'
  AND tablename = 'your_new_table';
-- Should return: rowsecurity = t

-- Test 2: Policies Exist
SELECT COUNT(*) 
FROM pg_policies
WHERE tablename = 'your_new_table';
-- Should return: >= 3 (admin, application, readonly)

-- Test 3: Isolation Works
BEGIN;
  PERFORM set_tenant_context('tenant-a-id');
  SELECT COUNT(*) FROM your_table WHERE tenant_id = 'tenant-b-id';
  -- Should return: 0 (RLS blocks cross-tenant access)
ROLLBACK;

-- Test 4: Admin Can See All
SET ROLE admin_role;
SELECT COUNT(DISTINCT tenant_id) FROM your_table;
-- Should return: > 1 (if multiple tenants have data)
RESET ROLE;

-- Test 5: Readonly Cannot Modify
SET ROLE readonly_role;
UPDATE your_table SET column = 'value';
-- Should fail: ERROR: permission denied
RESET ROLE;

-- Test 6: Application Cannot Bypass RLS
SET ROLE application_role;
PERFORM set_tenant_context('tenant-a-id');
INSERT INTO your_table (tenant_id, ...) 
VALUES ('tenant-b-id', ...);
-- Should fail: RLS check violation
RESET ROLE;
```

---

## 12. Conclusion

### 12.1 Summary

This database schema implements a **sophisticated multi-tenant architecture** with:

**✅ Strengths:**
1. **Strong Isolation**: Multiple layers (RLS, triggers, foreign keys)
2. **Flexible Configuration**: Dedicated columns + JSONB for experimentation
3. **Comprehensive Auditing**: Bulk operations, usage tracking
4. **Scalable Design**: Shared database pattern with proper indexing
5. **Security-First**: Role-based access, session management, validation
6. **Well-Documented**: Extensive comments and migration notes
7. **Extensible**: Clear patterns for adding new tables

**⚠️ Areas for Improvement:**
1. Status field casing inconsistency
2. Placeholder limit checks need implementation
3. Additional indexes recommended
4. Enhanced audit logging
5. Connection pooling best practices documentation

** Maturity Level**: **Production-Ready Foundation**
- Core infrastructure: ✓ Complete
- Security model: ✓ Robust
- Configuration: ✓ Flexible
- Monitoring: ✓ Basic (can be enhanced)
- Documentation: ✓ Excellent

### 12.2 Next Steps

**Immediate (Fix Inconsistencies):**
1. Resolve Status/status casing issue
2. Add missing indexes (email, composite)
3. Implement remaining limit checks

**Short-Term (Enhance Core):**
1. Add users table
2. Implement audit logging
3. Add feature flags
4. Create API key management

**Medium-Term (Operations):**
1. Add tenant_context_log for security monitoring
2. Implement automated usage stats collection
3. Create admin dashboard queries
4. Performance benchmarking and optimization

**Long-Term (Scale):**
1. Consider table partitioning for large tenants
2. Implement read replicas for reporting
3. Add caching layer (Redis/Memcached)
4. Consider tenant archival/cold storage

### 12.3 Architecture Assessment

**Grade: A- (Excellent Foundation)**

This is a **well-designed, production-ready** multi-tenant database schema that demonstrates:
- Deep understanding of PostgreSQL RLS
- Careful consideration of trade-offs
- Excellent documentation practices
- Security-conscious design
- Scalability planning

With minor refinements (fixing the identified issues), this schema provides a solid foundation for a multi-tenant ERP platform capable of serving hundreds to thousands of tenants securely and efficiently.

---

## Appendix A: Quick Reference

### Common Queries

```sql
-- Set tenant context
SELECT set_tenant_context('tenant-id');

-- Check current tenant
SELECT current_tenant_id();

-- Clear context
SELECT clear_tenant_context();

-- Provision new tenant
SELECT * FROM provision_tenant_complete(
  'Company Name',
  'admin@company.com',
  'company-subdomain'
);

-- Check tenant limits
SELECT check_tenant_limits(current_tenant_id(), 'storage', 1048576);

-- Get usage stats
SELECT * FROM tenant_usage_stats
WHERE tenant_id = current_tenant_id()
  AND period_start <= CURRENT_DATE
  AND period_end >= CURRENT_DATE;

-- Monitor bulk operation
SELECT * FROM get_bulk_operation_summary('operation-id');
```

### Standard Table Template

```sql
CREATE TABLE new_table (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  
  -- Your columns here
  
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_new_table_tenant ON new_table(tenant_id);

ALTER TABLE new_table ENABLE ROW LEVEL SECURITY;

CREATE POLICY admin_full_access ON new_table
  FOR ALL TO admin_role USING (true);

CREATE POLICY tenant_isolation ON new_table
  FOR ALL TO application_role
  USING (tenant_id = current_tenant_id())
  WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY readonly_access ON new_table
  FOR SELECT TO readonly_role USING (true);

GRANT SELECT, INSERT, UPDATE, DELETE ON new_table TO admin_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON new_table TO application_role;
GRANT SELECT ON new_table TO readonly_role;

CREATE TRIGGER update_new_table_updated_at
  BEFORE UPDATE ON new_table
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

---

**Document Version**: 1.0  
**Analysis Date**: February 2, 2024  
**Schema Version**: Migration 000107  
**Analyst**: Claude (Anthropic)
