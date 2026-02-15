




# Dynamic Report Generation & Customer Workflow Engine

---

# Document 1: Dynamic Report Generation System

## Table of Contents
1. [System Overview](#system-overview)
2. [Architecture Deep Dive](#architecture-deep-dive)
3. [Database Schema](#database-schema)
4. [Backend Implementation](#backend-implementation)
5. [AMIS Frontend Integration](#amis-frontend-integration)
6. [Security & Performance](#security-performance)
7. [Testing Strategy](#testing-strategy)
8. [Deployment & Operations](#deployment-operations)
9. [Extension Points](#extension-points)

---

## System Overview

### What This System Provides

The Dynamic Report Generation System enables tenants to create, customize, schedule, and share reports without writing code. It provides:

- **Visual Report Builder**: AMIS-powered drag-and-drop interface
- **Query Engine**: Safe SQL generation from user selections
- **Multiple Output Formats**: Tables, charts, PDFs, Excel, CSV
- **Scheduling**: Temporal-based automated report execution
- **Distribution**: Email, webhook, or download delivery
- **Access Control**: Entity and role-based report permissions

### Key Design Principles

1. **Security First**: All user input is validated and sanitized
2. **Performance Aware**: Query limits, timeouts, and read replica usage
3. **Tenant Isolated**: RLS ensures data isolation
4. **Extensible**: Easy to add new data sources and output formats
5. **Observable**: Comprehensive logging and metrics

---

## Architecture Deep Dive

### System Components

```
┌─────────────────────────────────────────────────────────────┐
│                     AMIS Frontend                            │
│  ┌──────────────────┐  ┌──────────────────┐                │
│  │  Report Builder  │  │  Report Viewer   │                │
│  │  (JSON Config)   │  │  (CRUD-List)     │                │
│  └──────────────────┘  └──────────────────┘                │
└────────────────┬────────────────┬──────────────────────────┘
                 │                │
                 │  HTTP/REST     │
                 ▼                ▼
┌─────────────────────────────────────────────────────────────┐
│                   API Layer (Fiber)                          │
│  ┌──────────────────┐  ┌──────────────────┐                │
│  │ Report Handler   │  │ Schedule Handler │                │
│  └──────────────────┘  └──────────────────┘                │
└────────────────┬────────────────┬──────────────────────────┘
                 │                │
                 ▼                ▼
┌─────────────────────────────────────────────────────────────┐
│              Service Layer (Business Logic)                  │
│  ┌──────────────────────────────────────────────────────┐  │
│  │                  Report Service                       │  │
│  │  - Validation    - Execution    - Formatting         │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────┬────────────────┬──────────────────────────┘
                 │                │
                 ▼                ▼
┌────────────────────────┐  ┌───────────────────────────────┐
│   Report Engine        │  │   Temporal Workflows          │
│   - Query Builder      │  │   - Scheduled Execution       │
│   - Query Executor     │  │   - Report Generation         │
│   - Result Processor   │  │   - Distribution              │
└────────────────────────┘  └───────────────────────────────┘
                 │                │
                 ▼                ▼
┌─────────────────────────────────────────────────────────────┐
│          PostgreSQL (Read Replica for Reports)               │
│  ┌──────────────────┐  ┌──────────────────┐                │
│  │  Report Views    │  │  Report Storage  │                │
│  └──────────────────┘  └──────────────────┘                │
└─────────────────────────────────────────────────────────────┘
```

### Data Flow

#### Report Creation Flow
1. User designs report in AMIS UI
2. AMIS generates JSON configuration
3. Frontend sends JSON to API
4. Service validates configuration
5. Report definition stored in PostgreSQL
6. User can preview or schedule

#### Report Execution Flow
1. User clicks "Run Report" OR schedule triggers
2. API/Temporal initiates execution
3. Report Engine builds SQL query
4. Query executed on read replica
5. Results processed and formatted
6. Output stored/delivered
7. Execution history recorded

#### Scheduled Report Flow
```
User Creates Schedule → Temporal Cron Workflow Started
         ↓
    Schedule Time Reached
         ↓
    Temporal Triggers Execution
         ↓
    Report Generated → Output Formatted
         ↓
    Email/Webhook Delivery
         ↓
    Execution Logged
```

---

## Database Schema

### Core Tables

```sql
-- =====================================================
-- REPORT MANAGEMENT SCHEMA
-- =====================================================

-- Report Definitions
-- Stores the structure and configuration of user-created reports
CREATE TABLE report_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Basic Information
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(50),  -- financial, operational, analytical
    tags TEXT[],           -- Searchable tags
    
    -- Report Configuration
    data_source VARCHAR(100) NOT NULL,  -- View name to query
    columns JSONB NOT NULL,             -- Column definitions
    filters JSONB DEFAULT '[]'::jsonb,  -- Filter conditions
    grouping JSONB DEFAULT '[]'::jsonb, -- GROUP BY configuration
    sorting JSONB DEFAULT '[]'::jsonb,  -- ORDER BY configuration
    
    -- Aggregation & Calculations
    calculations JSONB DEFAULT '[]'::jsonb,  -- Custom calculated fields
    having_conditions JSONB DEFAULT '[]'::jsonb,  -- HAVING clause filters
    
    -- Display Configuration
    output_format VARCHAR(20) DEFAULT 'table',  -- table, chart, pivot
    chart_config JSONB,                          -- Chart type and settings
    page_config JSONB,                           -- Pagination settings
    
    -- AMIS UI Configuration
    amis_schema JSONB,  -- Complete AMIS JSON schema for rendering
    
    -- Access Control
    is_public BOOLEAN DEFAULT false,
    is_system BOOLEAN DEFAULT false,  -- System reports (pre-built)
    owner_id UUID NOT NULL,           -- Creator user ID
    allowed_roles TEXT[],             -- Roles that can view/run
    allowed_entity_ids UUID[],        -- Entities with access
    
    -- Usage Tracking
    view_count INTEGER DEFAULT 0,
    last_run_at TIMESTAMPTZ,
    last_run_by UUID,
    run_count INTEGER DEFAULT 0,
    avg_execution_time_ms INTEGER,
    
    -- Versioning
    version INTEGER DEFAULT 1,
    parent_version_id UUID REFERENCES report_definitions(id),
    
    -- Audit Fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    CONSTRAINT valid_output_format CHECK (
        output_format IN ('table', 'chart', 'pivot', 'card')
    )
);

-- Indexes for Performance
CREATE INDEX idx_report_definitions_tenant ON report_definitions(tenant_id);
CREATE INDEX idx_report_definitions_owner ON report_definitions(owner_id);
CREATE INDEX idx_report_definitions_category ON report_definitions(category);
CREATE INDEX idx_report_definitions_data_source ON report_definitions(data_source);
CREATE INDEX idx_report_definitions_tags ON report_definitions USING gin(tags);
CREATE INDEX idx_report_definitions_deleted ON report_definitions(deleted_at) 
    WHERE deleted_at IS NOT NULL;

-- Comments
COMMENT ON TABLE report_definitions IS 'User-created report configurations with AMIS schema integration';
COMMENT ON COLUMN report_definitions.amis_schema IS 'Complete AMIS JSON schema for frontend rendering';
COMMENT ON COLUMN report_definitions.calculations IS 'Custom calculated fields using expressions';

-- =====================================================
-- Report Execution History
-- =====================================================
CREATE TABLE report_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID NOT NULL REFERENCES report_definitions(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Execution Context
    executed_by UUID NOT NULL,  -- User who ran the report
    execution_source VARCHAR(20) NOT NULL,  -- manual, scheduled, api
    
    -- Execution Details
    status VARCHAR(20) NOT NULL,  -- running, completed, failed, timeout
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    execution_time_ms INTEGER,
    
    -- Query Information
    generated_query TEXT,  -- Actual SQL executed
    query_parameters JSONB,
    
    -- Results
    row_count INTEGER,
    result_size_bytes BIGINT,
    result_file_path TEXT,  -- S3/storage path for result file
    
    -- Error Handling
    error_message TEXT,
    error_stack TEXT,
    
    -- Resource Usage
    peak_memory_mb INTEGER,
    cpu_time_ms INTEGER,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT valid_execution_status CHECK (
        status IN ('running', 'completed', 'failed', 'timeout', 'cancelled')
    ),
    CONSTRAINT valid_execution_source CHECK (
        execution_source IN ('manual', 'scheduled', 'api', 'webhook')
    )
);

-- Indexes
CREATE INDEX idx_report_executions_report ON report_executions(report_id);
CREATE INDEX idx_report_executions_tenant ON report_executions(tenant_id);
CREATE INDEX idx_report_executions_status ON report_executions(status);
CREATE INDEX idx_report_executions_started ON report_executions(started_at DESC);

COMMENT ON TABLE report_executions IS 'Audit trail of all report executions with performance metrics';

-- =====================================================
-- Report Schedules
-- =====================================================
CREATE TABLE report_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID NOT NULL REFERENCES report_definitions(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Schedule Configuration
    schedule_type VARCHAR(20) NOT NULL,  -- daily, weekly, monthly, cron
    cron_expression VARCHAR(100),        -- For complex schedules
    timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',
    
    -- Time Windows
    active_from TIMESTAMPTZ,
    active_until TIMESTAMPTZ,
    
    -- Output Configuration
    output_formats TEXT[] NOT NULL DEFAULT '{pdf}',  -- pdf, excel, csv
    
    -- Delivery Configuration
    delivery_method VARCHAR(20) NOT NULL,  -- email, webhook, storage
    email_recipients TEXT[],
    webhook_url TEXT,
    webhook_headers JSONB,
    
    -- Email Configuration
    email_subject VARCHAR(500),
    email_body TEXT,
    email_attachment_name VARCHAR(255),
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    last_run_at TIMESTAMPTZ,
    last_run_status VARCHAR(20),
    next_run_at TIMESTAMPTZ,
    
    -- Temporal Integration
    temporal_workflow_id VARCHAR(255),  -- Temporal cron workflow ID
    temporal_schedule_id VARCHAR(255),
    
    -- Failure Handling
    consecutive_failures INTEGER DEFAULT 0,
    max_consecutive_failures INTEGER DEFAULT 3,
    failure_notification_emails TEXT[],
    
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT valid_schedule_type CHECK (
        schedule_type IN ('daily', 'weekly', 'monthly', 'cron', 'once')
    ),
    CONSTRAINT valid_delivery_method CHECK (
        delivery_method IN ('email', 'webhook', 'storage', 'sftp')
    )
);

-- Indexes
CREATE INDEX idx_report_schedules_report ON report_schedules(report_id);
CREATE INDEX idx_report_schedules_tenant ON report_schedules(tenant_id);
CREATE INDEX idx_report_schedules_active ON report_schedules(is_active);
CREATE INDEX idx_report_schedules_next_run ON report_schedules(next_run_at) 
    WHERE is_active = true;

COMMENT ON TABLE report_schedules IS 'Automated report execution schedules with Temporal integration';

-- =====================================================
-- Data Source Definitions
-- =====================================================
CREATE TABLE report_data_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id),  -- NULL for system-wide sources
    
    -- Source Information
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(50),  -- financial, sales, inventory, hr
    
    -- Database Configuration
    source_type VARCHAR(20) NOT NULL,  -- view, table, function
    source_name VARCHAR(255) NOT NULL,  -- Actual database object name
    schema_name VARCHAR(63) DEFAULT 'public',
    
    -- Column Metadata
    columns JSONB NOT NULL,  -- Available columns with metadata
    
    -- Query Limits
    max_rows INTEGER DEFAULT 10000,
    default_timeout_seconds INTEGER DEFAULT 30,
    
    -- Access Control
    requires_entity_filter BOOLEAN DEFAULT true,
    is_tenant_specific BOOLEAN DEFAULT true,
    minimum_role VARCHAR(50),  -- Minimum role required to access
    
    -- Performance
    is_materialized BOOLEAN DEFAULT false,
    last_refresh_at TIMESTAMPTZ,
    refresh_schedule VARCHAR(100),  -- Cron expression
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    is_deprecated BOOLEAN DEFAULT false,
    deprecation_message TEXT,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT valid_source_type CHECK (
        source_type IN ('view', 'table', 'function', 'materialized_view')
    ),
    CONSTRAINT unique_source_per_tenant UNIQUE (tenant_id, name)
);

-- Indexes
CREATE INDEX idx_report_data_sources_category ON report_data_sources(category);
CREATE INDEX idx_report_data_sources_active ON report_data_sources(is_active);

COMMENT ON TABLE report_data_sources IS 'Registry of available data sources for report building';

-- =====================================================
-- Column Metadata Schema
-- =====================================================
-- Example column metadata structure in report_data_sources.columns:
/*
[
  {
    "name": "entity_name",
    "display_name": "Entity Name",
    "type": "string",
    "description": "Name of the business entity",
    "is_filterable": true,
    "is_sortable": true,
    "is_groupable": true,
    "aggregations": [],
    "format": null
  },
  {
    "name": "invoice_amount",
    "display_name": "Invoice Amount",
    "type": "decimal",
    "description": "Total invoice amount in tenant currency",
    "is_filterable": true,
    "is_sortable": true,
    "is_groupable": false,
    "aggregations": ["sum", "avg", "min", "max", "count"],
    "format": "currency",
    "precision": 2
  },
  {
    "name": "invoice_date",
    "display_name": "Invoice Date",
    "type": "date",
    "description": "Date the invoice was created",
    "is_filterable": true,
    "is_sortable": true,
    "is_groupable": true,
    "aggregations": ["count"],
    "format": "date",
    "date_format": "YYYY-MM-DD"
  }
]
*/

-- =====================================================
-- Report Shares
-- =====================================================
CREATE TABLE report_shares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID NOT NULL REFERENCES report_definitions(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Share Configuration
    shared_by UUID NOT NULL,
    share_token VARCHAR(64) NOT NULL UNIQUE,  -- For public links
    
    -- Access Control
    share_type VARCHAR(20) NOT NULL,  -- user, role, entity, public
    shared_with_user_id UUID,
    shared_with_role VARCHAR(50),
    shared_with_entity_id UUID REFERENCES entities(uuid),
    
    -- Permissions
    can_view BOOLEAN DEFAULT true,
    can_edit BOOLEAN DEFAULT false,
    can_delete BOOLEAN DEFAULT false,
    can_schedule BOOLEAN DEFAULT false,
    
    -- Time Limits
    expires_at TIMESTAMPTZ,
    
    -- Usage Tracking
    access_count INTEGER DEFAULT 0,
    last_accessed_at TIMESTAMPTZ,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT valid_share_type CHECK (
        share_type IN ('user', 'role', 'entity', 'public')
    )
);

-- Indexes
CREATE INDEX idx_report_shares_report ON report_shares(report_id);
CREATE INDEX idx_report_shares_token ON report_shares(share_token);
CREATE INDEX idx_report_shares_user ON report_shares(shared_with_user_id);

COMMENT ON TABLE report_shares IS 'Report sharing configuration for collaboration';

-- =====================================================
-- Row Level Security
-- =====================================================
ALTER TABLE report_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE report_executions ENABLE ROW LEVEL SECURITY;
ALTER TABLE report_schedules ENABLE ROW LEVEL SECURITY;
ALTER TABLE report_data_sources ENABLE ROW LEVEL SECURITY;
ALTER TABLE report_shares ENABLE ROW LEVEL SECURITY;

-- Policies for report_definitions
CREATE POLICY tenant_isolation_policy ON report_definitions
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id());

CREATE POLICY admin_full_access_policy ON report_definitions
    FOR ALL TO admin_role
    USING (TRUE);

-- Policies for report_executions
CREATE POLICY tenant_isolation_policy ON report_executions
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id());

-- Policies for report_schedules
CREATE POLICY tenant_isolation_policy ON report_schedules
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id());

-- Policies for report_data_sources
-- Allow access to both system-wide and tenant-specific sources
CREATE POLICY data_source_access_policy ON report_data_sources
    FOR SELECT TO application_role
    USING (
        tenant_id IS NULL  -- System-wide sources
        OR tenant_id = current_tenant_id()  -- Tenant-specific sources
    );

-- Policies for report_shares
CREATE POLICY tenant_isolation_policy ON report_shares
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id());

-- =====================================================
-- Triggers
-- =====================================================

-- Update report execution statistics
CREATE OR REPLACE FUNCTION update_report_statistics()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'completed' THEN
        UPDATE report_definitions SET
            last_run_at = NEW.completed_at,
            last_run_by = NEW.executed_by,
            run_count = run_count + 1,
            avg_execution_time_ms = CASE
                WHEN avg_execution_time_ms IS NULL THEN NEW.execution_time_ms
                ELSE (avg_execution_time_ms * run_count + NEW.execution_time_ms) / (run_count + 1)
            END
        WHERE id = NEW.report_id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_report_statistics_trigger
    AFTER INSERT OR UPDATE OF status ON report_executions
    FOR EACH ROW
    EXECUTE FUNCTION update_report_statistics();

COMMENT ON FUNCTION update_report_statistics() IS 'Maintains report usage statistics based on executions';

-- Update schedule next_run_at
CREATE OR REPLACE FUNCTION calculate_next_run_time()
RETURNS TRIGGER AS $$
BEGIN
    -- Simple calculation - in production use more sophisticated logic
    NEW.next_run_at := CASE
        WHEN NEW.schedule_type = 'daily' THEN 
            NEW.last_run_at + INTERVAL '1 day'
        WHEN NEW.schedule_type = 'weekly' THEN 
            NEW.last_run_at + INTERVAL '7 days'
        WHEN NEW.schedule_type = 'monthly' THEN 
            NEW.last_run_at + INTERVAL '1 month'
        ELSE 
            NEW.next_run_at  -- For cron, calculated separately
    END;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER calculate_next_run_trigger
    BEFORE UPDATE OF last_run_at ON report_schedules
    FOR EACH ROW
    EXECUTE FUNCTION calculate_next_run_time();

-- =====================================================
-- Permissions
-- =====================================================
GRANT SELECT, INSERT, UPDATE, DELETE ON report_definitions TO application_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON report_executions TO application_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON report_schedules TO application_role;
GRANT SELECT ON report_data_sources TO application_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON report_shares TO application_role;

GRANT ALL ON report_definitions TO admin_role;
GRANT ALL ON report_executions TO admin_role;
GRANT ALL ON report_schedules TO admin_role;
GRANT ALL ON report_data_sources TO admin_role;
GRANT ALL ON report_shares TO admin_role;
```

### Predefined Report Views

```sql
-- =====================================================
-- PREDEFINED REPORTING VIEWS
-- =====================================================

-- Invoice Summary View
CREATE OR REPLACE VIEW v_report_invoice_summary AS
SELECT
    i.id AS invoice_id,
    i.invoice_number,
    i.invoice_date,
    i.due_date,
    i.status,
    i.subtotal,
    i.tax_amount,
    i.total_amount,
    i.currency_code,
    e.uuid AS entity_id,
    e.name AS entity_name,
    e.type AS entity_type,
    v.name AS vendor_name,
    v.code AS vendor_code,
    c.name AS customer_name,
    c.code AS customer_code,
    u.email AS created_by_email,
    i.created_at,
    i.updated_at,
    -- Calculated fields
    CASE
        WHEN i.due_date < CURRENT_DATE AND i.status != 'paid' 
        THEN 'overdue'
        WHEN i.due_date <= CURRENT_DATE + INTERVAL '7 days' AND i.status != 'paid'
        THEN 'due_soon'
        ELSE 'current'
    END AS payment_status,
    CURRENT_DATE - i.due_date AS days_overdue,
    -- Tenant context
    i.tenant_id
FROM invoices i
JOIN entities e ON i.entity_id = e.uuid
LEFT JOIN vendors v ON i.vendor_id = v.id
LEFT JOIN customers c ON i.customer_id = c.id
LEFT JOIN users u ON i.created_by = u.id
WHERE i.deleted_at IS NULL;

COMMENT ON VIEW v_report_invoice_summary IS 'Comprehensive invoice data for reporting';

-- Enable RLS on view (inherits from base table)
ALTER VIEW v_report_invoice_summary SET (security_barrier = true);

-- Account Balance View
CREATE OR REPLACE VIEW v_report_account_balances AS
SELECT
    a.id AS account_id,
    a.code AS account_code,
    a.name AS account_name,
    a.type AS account_type,
    a.category,
    pa.name AS parent_account_name,
    e.uuid AS entity_id,
    e.name AS entity_name,
    e.type AS entity_type,
    -- Balance calculations
    COALESCE(SUM(t.debit_amount), 0) AS total_debits,
    COALESCE(SUM(t.credit_amount), 0) AS total_credits,
    COALESCE(SUM(t.debit_amount - t.credit_amount), 0) AS current_balance,
    -- Period analysis
    COUNT(t.id) AS transaction_count,
    MIN(t.transaction_date) AS first_transaction_date,
    MAX(t.transaction_date) AS last_transaction_date,
    a.tenant_id
FROM accounts a
JOIN entities e ON a.entity_id = e.uuid
LEFT JOIN accounts pa ON a.parent_id = pa.id
LEFT JOIN transactions t ON a.id = t.account_id
WHERE a.deleted_at IS NULL
GROUP BY 
    a.id, a.code, a.name, a.type, a.category,
    pa.name, e.uuid, e.name, e.type, a.tenant_id;

COMMENT ON VIEW v_report_account_balances IS 'Account balances with transaction aggregations';

-- Entity Performance View
CREATE OR REPLACE VIEW v_report_entity_performance AS
SELECT
    e.uuid AS entity_id,
    e.name AS entity_name,
    e.type AS entity_type,
    e.code AS entity_code,
    pe.name AS parent_entity_name,
    -- Revenue metrics
    COALESCE(SUM(CASE WHEN i.status = 'paid' THEN i.total_amount END), 0) AS revenue_ytd,
    COALESCE(SUM(CASE 
        WHEN i.status = 'paid' 
        AND i.invoice_date >= DATE_TRUNC('month', CURRENT_DATE)
        THEN i.total_amount 
    END), 0) AS revenue_mtd,
    -- Invoice metrics
    COUNT(DISTINCT i.id) AS total_invoices,
    COUNT(DISTINCT CASE WHEN i.status = 'paid' THEN i.id END) AS paid_invoices,
    COUNT(DISTINCT CASE WHEN i.status = 'pending' THEN i.id END) AS pending_invoices,
    -- Employee metrics (placeholder - requires users table)
    0 AS employee_count,
    e.tenant_id
FROM entities e
LEFT JOIN entities pe ON e.parent_id = pe.uuid
LEFT JOIN invoices i ON e.uuid = i.entity_id
WHERE e.deleted_at IS NULL
GROUP BY e.uuid, e.name, e.type, e.code, pe.name, e.tenant_id;

COMMENT ON VIEW v_report_entity_performance IS 'Entity-level performance metrics';

-- Register these views in report_data_sources
INSERT INTO report_data_sources (name, display_name, description, source_type, source_name, category, columns, is_tenant_specific)
VALUES
(
    'invoice_summary',
    'Invoice Summary',
    'Comprehensive invoice data with customer, vendor, and entity information',
    'view',
    'v_report_invoice_summary',
    'financial',
    '[
        {"name": "invoice_number", "display_name": "Invoice #", "type": "string", "is_filterable": true, "is_sortable": true},
        {"name": "invoice_date", "display_name": "Date", "type": "date", "is_filterable": true, "is_sortable": true},
        {"name": "total_amount", "display_name": "Total", "type": "decimal", "is_filterable": true, "aggregations": ["sum", "avg", "min", "max"]},
        {"name": "status", "display_name": "Status", "type": "string", "is_filterable": true, "is_groupable": true},
        {"name": "entity_name", "display_name": "Entity", "type": "string", "is_filterable": true, "is_groupable": true},
        {"name": "vendor_name", "display_name": "Vendor", "type": "string", "is_filterable": true, "is_groupable": true},
        {"name": "payment_status", "display_name": "Payment Status", "type": "string", "is_filterable": true, "is_groupable": true}
    ]'::jsonb,
    true
),
(
    'account_balances',
    'Account Balances',
    'Account balances with transaction summaries',
    'view',
    'v_report_account_balances',
    'financial',
    '[
        {"name": "account_code", "display_name": "Account Code", "type": "string", "is_filterable": true, "is_sortable": true},
        {"name": "account_name", "display_name": "Account Name", "type": "string", "is_filterable": true, "is_sortable": true},
        {"name": "account_type", "display_name": "Type", "type": "string", "is_filterable": true, "is_groupable": true},
        {"name": "current_balance", "display_name": "Balance", "type": "decimal", "is_filterable": true, "aggregations": ["sum"]},
        {"name": "entity_name", "display_name": "Entity", "type": "string", "is_filterable": true, "is_groupable": true},
        {"name": "transaction_count", "display_name": "Transactions", "type": "integer", "is_filterable": true, "aggregations": ["sum"]}
    ]'::jsonb,
    true
);
```

---

## Backend Implementation

### Report Engine Core

```go
// internal/reporting/engine.go
package reporting

import (
    "context"
    "database/sql"
    "fmt"
    "strings"
    "time"
    
    "github.com/google/uuid"
    "github.com/lib/pq"
)

// Engine is the core report generation engine
type Engine struct {
    db              *sql.DB
    readReplica     *sql.DB  // For running reports
    dataSources     map[string]*DataSource
    formatters      map[string]Formatter
    cache           Cache
    metricsRecorder MetricsRecorder
}

// DataSource defines an available reporting data source
type DataSource struct {
    Name              string
    DisplayName       string
    SourceType        string  // view, table, function
    SourceName        string  // Actual DB object name
    Columns           []ColumnMetadata
    MaxRows           int
    DefaultTimeout    time.Duration
    RequiresEntity    bool
    MinimumRole       string
}

// ColumnMetadata describes a column's properties
type ColumnMetadata struct {
    Name         string   `json:"name"`
    DisplayName  string   `json:"display_name"`
    Type         string   `json:"type"`  // string, integer, decimal, date, boolean
    Description  string   `json:"description"`
    IsFilterable bool     `json:"is_filterable"`
    IsSortable   bool     `json:"is_sortable"`
    IsGroupable  bool     `json:"is_groupable"`
    Aggregations []string `json:"aggregations"`  // sum, avg, count, min, max
    Format       string   `json:"format,omitempty"`  // currency, percent, date
    Precision    int      `json:"precision,omitempty"`
}

// ReportDefinition represents a user-created report
type ReportDefinition struct {
    ID          uuid.UUID
    TenantID    uuid.UUID
    Name        string
    Description string
    DataSource  string
    Columns     []ColumnConfig
    Filters     []FilterConfig
    Grouping    []string
    Sorting     []SortConfig
    Calculations []CalculationConfig
    OutputFormat string
    ChartConfig  *ChartConfig
    AMISSchema   map[string]interface{}
}

// ColumnConfig defines a selected column
type ColumnConfig struct {
    Name         string  `json:"name"`
    Label        string  `json:"label"`
    Aggregation  string  `json:"aggregation,omitempty"`  // sum, avg, count, etc.
    Format       string  `json:"format,omitempty"`
    Width        int     `json:"width,omitempty"`
    Visible      bool    `json:"visible"`
}

// FilterConfig defines a filter condition
type FilterConfig struct {
    Column     string      `json:"column"`
    Operator   string      `json:"operator"`  // =, !=, >, <, >=, <=, in, like, between
    Value      interface{} `json:"value"`
    Type       string      `json:"type"`      // string, integer, decimal, date
    Conjunction string     `json:"conjunction,omitempty"`  // AND, OR
}

// SortConfig defines sorting
type SortConfig struct {
    Column    string `json:"column"`
    Direction string `json:"direction"`  // asc, desc
}

// CalculationConfig defines a calculated field
type CalculationConfig struct {
    Name       string `json:"name"`
    Expression string `json:"expression"`
    Type       string `json:"type"`
    Format     string `json:"format,omitempty"`
}

// ChartConfig for visualization
type ChartConfig struct {
    Type       string   `json:"type"`  // bar, line, pie, area
    XAxis      string   `json:"x_axis"`
    YAxis      []string `json:"y_axis"`
    Series     []string `json:"series,omitempty"`
    Title      string   `json:"title,omitempty"`
    SubTitle   string   `json:"subtitle,omitempty"`
}

// ReportResult contains execution results
type ReportResult struct {
    Columns   []string
    Rows      [][]interface{}
    RowCount  int
    Metadata  map[string]interface{}
    ExecutionTime time.Duration
}

// NewEngine creates a new report engine
func NewEngine(db, readReplica *sql.DB) *Engine {
    engine := &Engine{
        db:          db,
        readReplica: readReplica,
        dataSources: make(map[string]*DataSource),
        formatters:  make(map[string]Formatter),
    }
    
    // Register formatters
    engine.RegisterFormatter("pdf", &PDFFormatter{})
    engine.RegisterFormatter("excel", &ExcelFormatter{})
    engine.RegisterFormatter("csv", &CSVFormatter{})
    engine.RegisterFormatter("json", &JSONFormatter{})
    
    return engine
}

// LoadDataSources loads available data sources from database
func (e *Engine) LoadDataSources(ctx context.Context) error {
    query := `
        SELECT 
            name, display_name, source_type, source_name,
            columns, max_rows, default_timeout_seconds,
            requires_entity_filter, minimum_role
        FROM report_data_sources
        WHERE is_active = true
          AND (tenant_id IS NULL OR tenant_id = current_tenant_id())
    `
    
    rows, err := e.db.QueryContext(ctx, query)
    if err != nil {
        return fmt.Errorf("failed to load data sources: %w", err)
    }
    defer rows.Close()
    
    for rows.Next() {
        ds := &DataSource{}
        var columnsJSON []byte
        var timeoutSeconds int
        
        err := rows.Scan(
            &ds.Name, &ds.DisplayName, &ds.SourceType, &ds.SourceName,
            &columnsJSON, &ds.MaxRows, &timeoutSeconds,
            &ds.RequiresEntity, &ds.MinimumRole,
        )
        if err != nil {
            return err
        }
        
        // Parse columns metadata
        if err := json.Unmarshal(columnsJSON, &ds.Columns); err != nil {
            return fmt.Errorf("failed to parse columns for %s: %w", ds.Name, err)
        }
        
        ds.DefaultTimeout = time.Duration(timeoutSeconds) * time.Second
        e.dataSources[ds.Name] = ds
    }
    
    return rows.Err()
}

// ExecuteReport generates and runs a report
func (e *Engine) ExecuteReport(ctx context.Context, reportDef *ReportDefinition) (*ReportResult, error) {
    startTime := time.Now()
    
    // Validate report definition
    if err := e.validateReport(reportDef); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    // Build SQL query
    query, params, err := e.buildQuery(reportDef)
    if err != nil {
        return nil, fmt.Errorf("query building failed: %w", err)
    }
    
    // Log the query (for debugging/auditing)
    e.metricsRecorder.RecordQuery(ctx, reportDef.ID, query)
    
    // Execute query with timeout
    queryCtx, cancel := context.WithTimeout(ctx, 
        e.dataSources[reportDef.DataSource].DefaultTimeout)
    defer cancel()
    
    // Use read replica for queries
    rows, err := e.readReplica.QueryContext(queryCtx, query, params...)
    if err != nil {
        if err == context.DeadlineExceeded {
            return nil, fmt.Errorf("query timeout exceeded")
        }
        return nil, fmt.Errorf("query execution failed: %w", err)
    }
    defer rows.Close()
    
    // Process results
    result, err := e.processResults(rows, reportDef)
    if err != nil {
        return nil, fmt.Errorf("result processing failed: %w", err)
    }
    
    result.ExecutionTime = time.Since(startTime)
    
    // Record metrics
    e.metricsRecorder.RecordExecution(ctx, reportDef.ID, result.ExecutionTime, result.RowCount)
    
    return result, nil
}

// validateReport ensures the report definition is valid and safe
func (e *Engine) validateReport(reportDef *ReportDefinition) error {
    // Check data source exists
    ds, exists := e.dataSources[reportDef.DataSource]
    if !exists {
        return fmt.Errorf("unknown data source: %s", reportDef.DataSource)
    }
    
    // Build column map for quick lookup
    allowedCols := make(map[string]ColumnMetadata)
    for _, col := range ds.Columns {
        allowedCols[col.Name] = col
    }
    
    // Validate columns
    for _, col := range reportDef.Columns {
        allowed, ok := allowedCols[col.Name]
        if !ok {
            return fmt.Errorf("column not allowed: %s", col.Name)
        }
        
        // Validate aggregation
        if col.Aggregation != "" {
            if !contains(allowed.Aggregations, col.Aggregation) {
                return fmt.Errorf("aggregation %s not allowed for column %s", 
                    col.Aggregation, col.Name)
            }
        }
    }
    
    // Validate filters
    for _, filter := range reportDef.Filters {
        colMeta, ok := allowedCols[filter.Column]
        if !ok {
            return fmt.Errorf("filter column not allowed: %s", filter.Column)
        }
        
        if !colMeta.IsFilterable {
            return fmt.Errorf("column %s is not filterable", filter.Column)
        }
        
        // Validate operator
        if !isValidOperator(filter.Operator) {
            return fmt.Errorf("invalid operator: %s", filter.Operator)
        }
        
        // Validate value type matches column type
        if err := validateValueType(filter.Value, colMeta.Type); err != nil {
            return fmt.Errorf("invalid filter value for %s: %w", filter.Column, err)
        }
    }
    
    // Validate grouping columns
    for _, groupCol := range reportDef.Grouping {
        colMeta, ok := allowedCols[groupCol]
        if !ok {
            return fmt.Errorf("grouping column not allowed: %s", groupCol)
        }
        
        if !colMeta.IsGroupable {
            return fmt.Errorf("column %s is not groupable", groupCol)
        }
    }
    
    // Validate sorting
    for _, sort := range reportDef.Sorting {
        colMeta, ok := allowedCols[sort.Column]
        if !ok {
            return fmt.Errorf("sort column not allowed: %s", sort.Column)
        }
        
        if !colMeta.IsSortable {
            return fmt.Errorf("column %s is not sortable", sort.Column)
        }
        
        if sort.Direction != "asc" && sort.Direction != "desc" {
            return fmt.Errorf("invalid sort direction: %s", sort.Direction)
        }
    }
    
    return nil
}

// buildQuery constructs SQL from report definition
func (e *Engine) buildQuery(reportDef *ReportDefinition) (string, []interface{}, error) {
    ds := e.dataSources[reportDef.DataSource]
    
    var query strings.Builder
    var params []interface{}
    paramIndex := 1
    
    // SELECT clause
    query.WriteString("SELECT ")
    
    selectClauses := make([]string, 0, len(reportDef.Columns))
    for _, col := range reportDef.Columns {
        if col.Aggregation != "" {
            // Aggregated column
            selectClauses = append(selectClauses,
                fmt.Sprintf("%s(%s) AS %s", 
                    strings.ToUpper(col.Aggregation), 
                    e.quoteIdentifier(col.Name),
                    e.quoteIdentifier(col.Name)))
        } else {
            // Regular column
            selectClauses = append(selectClauses, e.quoteIdentifier(col.Name))
        }
    }
    
    query.WriteString(strings.Join(selectClauses, ", "))
    
    // FROM clause
    query.WriteString(fmt.Sprintf(" FROM %s", e.quoteIdentifier(ds.SourceName)))
    
    // WHERE clause
    whereClauses := []string{}
    
    // Add filters
    for i, filter := range reportDef.Filters {
        clause, filterParams := e.buildFilterClause(filter, &paramIndex)
        
        // Handle conjunction for multiple filters
        if i > 0 {
            conjunction := filter.Conjunction
            if conjunction == "" {
                conjunction = "AND"
            }
            whereClauses = append(whereClauses, conjunction+" "+clause)
        } else {
            whereClauses = append(whereClauses, clause)
        }
        
        params = append(params, filterParams...)
    }
    
    if len(whereClauses) > 0 {
        query.WriteString(" WHERE ")
        query.WriteString(strings.Join(whereClauses, " "))
    }
    
    // GROUP BY clause
    if len(reportDef.Grouping) > 0 {
        query.WriteString(" GROUP BY ")
        groupClauses := make([]string, len(reportDef.Grouping))
        for i, col := range reportDef.Grouping {
            groupClauses[i] = e.quoteIdentifier(col)
        }
        query.WriteString(strings.Join(groupClauses, ", "))
    }
    
    // ORDER BY clause
    if len(reportDef.Sorting) > 0 {
        query.WriteString(" ORDER BY ")
        sortClauses := make([]string, len(reportDef.Sorting))
        for i, sort := range reportDef.Sorting {
            sortClauses[i] = fmt.Sprintf("%s %s", 
                e.quoteIdentifier(sort.Column),
                strings.ToUpper(sort.Direction))
        }
        query.WriteString(strings.Join(sortClauses, ", "))
    }
    
    // LIMIT clause (safety measure)
    query.WriteString(fmt.Sprintf(" LIMIT %d", ds.MaxRows))
    
    return query.String(), params, nil
}

// buildFilterClause constructs a WHERE condition
func (e *Engine) buildFilterClause(filter FilterConfig, paramIndex *int) (string, []interface{}) {
    var params []interface{}
    var clause string
    
    column := e.quoteIdentifier(filter.Column)
    
    switch filter.Operator {
    case "=", "!=", ">", ">=", "<", "<=":
        clause = fmt.Sprintf("%s %s $%d", column, filter.Operator, *paramIndex)
        params = append(params, filter.Value)
        *paramIndex++
        
    case "in":
        values := filter.Value.([]interface{})
        placeholders := make([]string, len(values))
        for i := range values {
            placeholders[i] = fmt.Sprintf("$%d", *paramIndex)
            params = append(params, values[i])
            *paramIndex++
        }
        clause = fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ", "))
        
    case "not_in":
        values := filter.Value.([]interface{})
        placeholders := make([]string, len(values))
        for i := range values {
            placeholders[i] = fmt.Sprintf("$%d", *paramIndex)
            params = append(params, values[i])
            *paramIndex++
        }
        clause = fmt.Sprintf("%s NOT IN (%s)", column, strings.Join(placeholders, ", "))
        
    case "like":
        clause = fmt.Sprintf("%s LIKE $%d", column, *paramIndex)
        params = append(params, filter.Value)
        *paramIndex++
        
    case "ilike":
        clause = fmt.Sprintf("%s ILIKE $%d", column, *paramIndex)
        params = append(params, filter.Value)
        *paramIndex++
        
    case "between":
        values := filter.Value.([]interface{})
        clause = fmt.Sprintf("%s BETWEEN $%d AND $%d", column, *paramIndex, *paramIndex+1)
        params = append(params, values[0], values[1])
        *paramIndex += 2
        
    case "is_null":
        clause = fmt.Sprintf("%s IS NULL", column)
        
    case "is_not_null":
        clause = fmt.Sprintf("%s IS NOT NULL", column)
    }
    
    return clause, params
}

// processResults converts SQL rows to report result
func (e *Engine) processResults(rows *sql.Rows, reportDef *ReportDefinition) (*ReportResult, error) {
    columns, err := rows.Columns()
    if err != nil {
        return nil, err
    }
    
    result := &ReportResult{
        Columns: columns,
        Rows:    make([][]interface{}, 0),
        Metadata: make(map[string]interface{}),
    }
    
    // Get column types
    columnTypes, err := rows.ColumnTypes()
    if err != nil {
        return nil, err
    }
    
    for rows.Next() {
        // Create slice for scanning
        values := make([]interface{}, len(columns))
        valuePtrs := make([]interface{}, len(columns))
        
        for i := range values {
            valuePtrs[i] = &values[i]
        }
        
        if err := rows.Scan(valuePtrs...); err != nil {
            return nil, err
        }
        
        // Format values based on column configuration
        formattedValues := make([]interface{}, len(values))
        for i, val := range values {
            formattedValues[i] = e.formatValue(val, columnTypes[i], reportDef.Columns[i])
        }
        
        result.Rows = append(result.Rows, formattedValues)
    }
    
    if err := rows.Err(); err != nil {
        return nil, err
    }
    
    result.RowCount = len(result.Rows)
    
    // Add metadata
    result.Metadata["generated_at"] = time.Now()
    result.Metadata["row_count"] = result.RowCount
    result.Metadata["column_count"] = len(columns)
    
    return result, nil
}

// formatValue applies formatting to a value
func (e *Engine) formatValue(val interface{}, colType *sql.ColumnType, config ColumnConfig) interface{} {
    if val == nil {
        return nil
    }
    
    // Handle byte arrays (common for numeric types)
    if b, ok := val.([]byte); ok {
        val = string(b)
    }
    
    // Apply format based on configuration
    switch config.Format {
    case "currency":
        if num, ok := val.(float64); ok {
            return fmt.Sprintf("%.2f", num)
        }
    case "percent":
        if num, ok := val.(float64); ok {
            return fmt.Sprintf("%.2f%%", num*100)
        }
    case "date":
        if t, ok := val.(time.Time); ok {
            return t.Format("2006-01-02")
        }
    case "datetime":
        if t, ok := val.(time.Time); ok {
            return t.Format("2006-01-02 15:04:05")
        }
    }
    
    return val
}

// quoteIdentifier safely quotes a column or table name
func (e *Engine) quoteIdentifier(name string) string {
    return pq.QuoteIdentifier(name)
}

// Helper functions
func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}

func isValidOperator(op string) bool {
    validOps := []string{"=", "!=", ">", ">=", "<", "<=", "in", "not_in", "like", "ilike", "between", "is_null", "is_not_null"}
    return contains(validOps, op)
}

func validateValueType(value interface{}, expectedType string) error {
    // Type validation logic
    switch expectedType {
    case "string":
        if _, ok := value.(string); !ok {
            return fmt.Errorf("expected string, got %T", value)
        }
    case "integer":
        switch value.(type) {
        case int, int32, int64, float64:
            // OK
        default:
            return fmt.Errorf("expected integer, got %T", value)
        }
    case "decimal":
        switch value.(type) {
        case float64, float32:
            // OK
        default:
            return fmt.Errorf("expected decimal, got %T", value)
        }
    }
    return nil
}

// RegisterFormatter registers an output formatter
func (e *Engine) RegisterFormatter(format string, formatter Formatter) {
    e.formatters[format] = formatter
}
```

### Output Formatters

```go
// internal/reporting/formatters.go
package reporting

import (
    "bytes"
    "encoding/csv"
    "encoding/json"
    "fmt"
    "io"
    
    "github.com/xuri/excelize/v2"
    "github.com/jung-kurt/gofpdf"
)

// Formatter interface for output formats
type Formatter interface {
    Format(result *ReportResult, config map[string]interface{}) ([]byte, error)
    ContentType() string
    FileExtension() string
}

// ==================================================
// CSV Formatter
// ==================================================
type CSVFormatter struct{}

func (f *CSVFormatter) Format(result *ReportResult, config map[string]interface{}) ([]byte, error) {
    var buf bytes.Buffer
    writer := csv.NewWriter(&buf)
    
    // Write header
    if err := writer.Write(result.Columns); err != nil {
        return nil, err
    }
    
    // Write rows
    for _, row := range result.Rows {
        strRow := make([]string, len(row))
        for i, val := range row {
            strRow[i] = fmt.Sprintf("%v", val)
        }
        if err := writer.Write(strRow); err != nil {
            return nil, err
        }
    }
    
    writer.Flush()
    if err := writer.Error(); err != nil {
        return nil, err
    }
    
    return buf.Bytes(), nil
}

func (f *CSVFormatter) ContentType() string {
    return "text/csv"
}

func (f *CSVFormatter) FileExtension() string {
    return ".csv"
}

// ==================================================
// Excel Formatter
// ==================================================
type ExcelFormatter struct{}

func (f *ExcelFormatter) Format(result *ReportResult, config map[string]interface{}) ([]byte, error) {
    file := excelize.NewFile()
    sheetName := "Report"
    
    // Create sheet
    index, err := file.NewSheet(sheetName)
    if err != nil {
        return nil, err
    }
    
    // Write headers
    for i, col := range result.Columns {
        cell, _ := excelize.CoordinatesToCellName(i+1, 1)
        file.SetCellValue(sheetName, cell, col)
    }
    
    // Style header row
    headerStyle, _ := file.NewStyle(&excelize.Style{
        Font: &excelize.Font{Bold: true},
        Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
    })
    file.SetCellStyle(sheetName, "A1", 
        fmt.Sprintf("%s1", string(rune('A'+len(result.Columns)-1))), headerStyle)
    
    // Write data rows
    for rowIdx, row := range result.Rows {
        for colIdx, val := range row {
            cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
            file.SetCellValue(sheetName, cell, val)
        }
    }
    
    // Auto-fit columns
    for i := range result.Columns {
        col, _ := excelize.ColumnNumberToName(i + 1)
        file.SetColWidth(sheetName, col, col, 15)
    }
    
    file.SetActiveSheet(index)
    
    // Write to buffer
    var buf bytes.Buffer
    if err := file.Write(&buf); err != nil {
        return nil, err
    }
    
    return buf.Bytes(), nil
}

func (f *ExcelFormatter) ContentType() string {
    return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
}

func (f *ExcelFormatter) FileExtension() string {
    return ".xlsx"
}

// ==================================================
// PDF Formatter
// ==================================================
type PDFFormatter struct{}

func (f *PDFFormatter) Format(result *ReportResult, config map[string]interface{}) ([]byte, error) {
    pdf := gofpdf.New("L", "mm", "A4", "")
    pdf.AddPage()
    
    // Title
    pdf.SetFont("Arial", "B", 16)
    title := "Report"
    if t, ok := config["title"].(string); ok {
        title = t
    }
    pdf.Cell(0, 10, title)
    pdf.Ln(12)
    
    // Table header
    pdf.SetFont("Arial", "B", 10)
    pdf.SetFillColor(240, 240, 240)
    
    colWidth := 270.0 / float64(len(result.Columns))
    for _, col := range result.Columns {
        pdf.CellFormat(colWidth, 7, col, "1", 0, "", true, 0, "")
    }
    pdf.Ln(-1)
    
    // Table rows
    pdf.SetFont("Arial", "", 9)
    pdf.SetFillColor(255, 255, 255)
    
    for _, row := range result.Rows {
        for _, val := range row {
            pdf.CellFormat(colWidth, 6, fmt.Sprintf("%v", val), "1", 0, "", false, 0, "")
        }
        pdf.Ln(-1)
    }
    
    // Write to buffer
    var buf bytes.Buffer
    if err := pdf.Output(&buf); err != nil {
        return nil, err
    }
    
    return buf.Bytes(), nil
}

func (f *PDFFormatter) ContentType() string {
    return "application/pdf"
}

func (f *PDFFormatter) FileExtension() string {
    return ".pdf"
}

// ==================================================
// JSON Formatter
// ==================================================
type JSONFormatter struct{}

func (f *JSONFormatter) Format(result *ReportResult, config map[string]interface{}) ([]byte, error) {
    // Convert to JSON-friendly structure
    output := map[string]interface{}{
        "columns":   result.Columns,
        "rows":      result.Rows,
        "row_count": result.RowCount,
        "metadata":  result.Metadata,
    }
    
    return json.MarshalIndent(output, "", "  ")
}

func (f *JSONFormatter) ContentType() string {
    return "application/json"
}

func (f *JSONFormatter) FileExtension() string {
    return ".json"
}
```

### Service Layer

```go
// internal/reporting/service.go
package reporting

import (
    "context"
    "fmt"
    "time"
    
    "github.com/google/uuid"
)

// Service handles report business logic
type Service struct {
    engine     *Engine
    repo       Repository
    storage    Storage
    temporal   TemporalClient
    authz      AuthorizationService
}

// CreateReport creates a new report definition
func (s *Service) CreateReport(ctx context.Context, req CreateReportRequest) (*ReportDefinition, error) {
    // Get user from context
    userID := getUserIDFromContext(ctx)
    tenantID := getTenantIDFromContext(ctx)
    
    // Validate request
    if err := s.validateCreateRequest(req); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    // Generate AMIS schema from report definition
    amisSchema, err := s.generateAMISSchema(req)
    if err != nil {
        return nil, fmt.Errorf("failed to generate AMIS schema: %w", err)
    }
    
    // Create report definition
    reportDef := &ReportDefinition{
        ID:           uuid.New(),
        TenantID:     tenantID,
        Name:         req.Name,
        Description:  req.Description,
        DataSource:   req.DataSource,
        Columns:      req.Columns,
        Filters:      req.Filters,
        Grouping:     req.Grouping,
        Sorting:      req.Sorting,
        OutputFormat: req.OutputFormat,
        ChartConfig:  req.ChartConfig,
        AMISSchema:   amisSchema,
    }
    
    // Save to database
    if err := s.repo.CreateReport(ctx, reportDef); err != nil {
        return nil, err
    }
    
    return reportDef, nil
}

// ExecuteReport runs a report and returns results
func (s *Service) ExecuteReport(ctx context.Context, reportID uuid.UUID) (*ReportResult, error) {
    userID := getUserIDFromContext(ctx)
    
    // Get report definition
    reportDef, err := s.repo.GetReport(ctx, reportID)
    if err != nil {
        return nil, err
    }
    
    // Check authorization
    if !s.authz.CanViewReport(ctx, userID, reportDef) {
        return nil, ErrUnauthorized
    }
    
    // Create execution record
    execution := &ReportExecution{
        ID:          uuid.New(),
        ReportID:    reportID,
        TenantID:    reportDef.TenantID,
        ExecutedBy:  userID,
        Status:      "running",
        Source:      "manual",
        StartedAt:   time.Now(),
    }
    
    if err := s.repo.CreateExecution(ctx, execution); err != nil {
        return nil, err
    }
    
    // Execute report
    result, err := s.engine.ExecuteReport(ctx, reportDef)
    
    // Update execution record
    execution.CompletedAt = time.Now()
    execution.ExecutionTime = time.Since(execution.StartedAt).Milliseconds()
    
    if err != nil {
        execution.Status = "failed"
        execution.ErrorMessage = err.Error()
        s.repo.UpdateExecution(ctx, execution)
        return nil, err
    }
    
    execution.Status = "completed"
    execution.RowCount = result.RowCount
    s.repo.UpdateExecution(ctx, execution)
    
    return result, nil
}

// ScheduleReport creates a scheduled report
func (s *Service) ScheduleReport(ctx context.Context, req CreateScheduleRequest) (*ReportSchedule, error) {
    userID := getUserIDFromContext(ctx)
    tenantID := getTenantIDFromContext(ctx)
    
    // Validate report exists and user has access
    reportDef, err := s.repo.GetReport(ctx, req.ReportID)
    if err != nil {
        return nil, err
    }
    
    if !s.authz.CanScheduleReport(ctx, userID, reportDef) {
        return nil, ErrUnauthorized
    }
    
    // Create schedule
    schedule := &ReportSchedule{
        ID:             uuid.New(),
        ReportID:       req.ReportID,
        TenantID:       tenantID,
        ScheduleType:   req.ScheduleType,
        CronExpression: req.CronExpression,
        Timezone:       req.Timezone,
        DeliveryMethod: req.DeliveryMethod,
        Recipients:     req.Recipients,
        OutputFormats:  req.OutputFormats,
        IsActive:       true,
        CreatedBy:      userID,
    }
    
    // Calculate next run time
    schedule.NextRunAt = calculateNextRun(schedule)
    
    // Save schedule
    if err := s.repo.CreateSchedule(ctx, schedule); err != nil {
        return nil, err
    }
    
    // Start Temporal cron workflow
    workflowID, err := s.startScheduledWorkflow(ctx, schedule)
    if err != nil {
        // Rollback schedule creation
        s.repo.DeleteSchedule(ctx, schedule.ID)
        return nil, fmt.Errorf("failed to start workflow: %w", err)
    }
    
    schedule.TemporalWorkflowID = workflowID
    s.repo.UpdateSchedule(ctx, schedule)
    
    return schedule, nil
}

// generateAMISSchema creates AMIS JSON schema for report rendering
func (s *Service) generateAMISSchema(req CreateReportRequest) (map[string]interface{}, error) {
    // Get data source metadata
    ds := s.engine.dataSources[req.DataSource]
    
    // Build AMIS schema based on output format
    schema := map[string]interface{}{
        "type": "page",
        "title": req.Name,
        "body": []interface{}{},
    }
    
    if req.OutputFormat == "table" {
        tableSchema := s.generateTableSchema(req, ds)
        schema["body"] = []interface{}{tableSchema}
    } else if req.OutputFormat == "chart" {
        chartSchema := s.generateChartSchema(req, ds)
        schema["body"] = []interface{}{chartSchema}
    }
    
    return schema, nil
}

// generateTableSchema creates AMIS table configuration
func (s *Service) generateTableSchema(req CreateReportRequest, ds *DataSource) map[string]interface{} {
    columns := make([]map[string]interface{}, 0, len(req.Columns))
    
    for _, col := range req.Columns {
        colSchema := map[string]interface{}{
            "name":  col.Name,
            "label": col.Label,
            "type":  "text",
        }
        
        // Find column metadata
        for _, meta := range ds.Columns {
            if meta.Name == col.Name {
                // Apply type-specific formatting
                switch meta.Type {
                case "decimal":
                    colSchema["type"] = "number"
                    if col.Format == "currency" {
                        colSchema["prefix"] = "$"
                        colSchema["precision"] = 2
                    }
                case "date":
                    colSchema["type"] = "date"
                    colSchema["format"] = "YYYY-MM-DD"
                case "boolean":
                    colSchema["type"] = "status"
                }
                break
            }
        }
        
        columns = append(columns, colSchema)
    }
    
    return map[string]interface{}{
        "type": "crud",
        "api":  fmt.Sprintf("/api/reports/%s/data", req.Name),
        "columns": columns,
        "headerToolbar": []interface{}{
            "filter-toggler",
            "reload",
            map[string]interface{}{
                "type":  "export-excel",
                "label": "Export",
            },
        },
        "footerToolbar": []interface{}{"pagination"},
        "perPage": 50,
    }
}

// generateChartSchema creates AMIS chart configuration
func (s *Service) generateChartSchema(req CreateReportRequest, ds *DataSource) map[string]interface{} {
    if req.ChartConfig == nil {
        return map[string]interface{}{}
    }
    
    return map[string]interface{}{
        "type": "chart",
        "api":  fmt.Sprintf("/api/reports/%s/data", req.Name),
        "config": map[string]interface{}{
            "xAxis": map[string]interface{}{
                "type": "category",
                "data": fmt.Sprintf("${%s}", req.ChartConfig.XAxis),
            },
            "yAxis": map[string]interface{}{
                "type": "value",
            },
            "series": s.generateChartSeries(req.ChartConfig),
        },
    }
}

func (s *Service) generateChartSeries(config *ChartConfig) []map[string]interface{} {
    series := make([]map[string]interface{}, 0, len(config.YAxis))
    
    for _, yAxis := range config.YAxis {
        series = append(series, map[string]interface{}{
            "type": config.Type,  // bar, line, pie
            "name": yAxis,
            "data": fmt.Sprintf("${%s}", yAxis),
        })
    }
    
    return series
}
```

---

## AMIS Frontend Integration

AMIS is perfect for this use case because it's JSON-schema driven. Let me show you how to build the report builder and viewer.

### Report Builder UI (AMIS)

```json
{
  "type": "page",
  "title": "Create Report",
  "body": {
    "type": "wizard",
    "steps": [
      {
        "title": "Basic Information",
        "body": [
          {
            "type": "input-text",
            "name": "name",
            "label": "Report Name",
            "required": true,
            "placeholder": "Enter report name"
          },
          {
            "type": "textarea",
            "name": "description",
            "label": "Description",
            "placeholder": "Describe what this report shows"
          },
          {
            "type": "select",
            "name": "data_source",
            "label": "Data Source",
            "required": true,
            "source": "/api/reports/data-sources",
            "labelField": "display_name",
            "valueField": "name",
            "description": "Choose the data you want to report on"
          }
        ]
      },
      {
        "title": "Select Columns",
        "body": [
          {
            "type": "transfer",
            "name": "columns",
            "label": "Report Columns",
            "source": "/api/reports/data-sources/${data_source}/columns",
            "searchable": true,
            "selectMode": "list",
            "labelField": "display_name",
            "valueField": "name",
            "statistics": true
          },
          {
            "type": "combo",
            "name": "column_config",
            "label": "Column Configuration",
            "multiple": true,
            "items": [
              {
                "type": "select",
                "name": "column",
                "label": "Column",
                "source": "${columns}",
                "required": true
              },
              {
                "type": "input-text",
                "name": "label",
                "label": "Display Label",
                "placeholder": "How to display this column"
              },
              {
                "type": "select",
                "name": "aggregation",
                "label": "Aggregation",
                "options": [
                  {"label": "None", "value": ""},
                  {"label": "Sum", "value": "sum"},
                  {"label": "Average", "value": "avg"},
                  {"label": "Count", "value": "count"},
                  {"label": "Min", "value": "min"},
                  {"label": "Max", "value": "max"}
                ],
                "visibleOn": "${column.aggregations && column.aggregations.length > 0}"
              },
              {
                "type": "select",
                "name": "format",
                "label": "Format",
                "options": [
                  {"label": "Default", "value": ""},
                  {"label": "Currency", "value": "currency"},
                  {"label": "Percentage", "value": "percent"},
                  {"label": "Date", "value": "date"}
                ]
              }
            ]
          }
        ]
      },
      {
        "title": "Add Filters",
        "body": [
          {
            "type": "combo",
            "name": "filters",
            "label": "Filter Conditions",
            "multiple": true,
            "multiLine": true,
            "addable": true,
            "removable": true,
            "items": [
              {
                "type": "select",
                "name": "column",
                "label": "Column",
                "source": "/api/reports/data-sources/${data_source}/columns?filterable=true",
                "labelField": "display_name",
                "valueField": "name",
                "required": true
              },
              {
                "type": "select",
                "name": "operator",
                "label": "Operator",
                "options": [
                  {"label": "Equals", "value": "="},
                  {"label": "Not Equals", "value": "!="},
                  {"label": "Greater Than", "value": ">"},
                  {"label": "Less Than", "value": "<"},
                  {"label": "Contains", "value": "like"},
                  {"label": "In List", "value": "in"},
                  {"label": "Between", "value": "between"}
                ],
                "required": true
              },
              {
                "type": "input-text",
                "name": "value",
                "label": "Value",
                "required": true,
                "visibleOn": "${operator !== 'in' && operator !== 'between'}"
              },
              {
                "type": "input-tag",
                "name": "value",
                "label": "Values",
                "required": true,
                "visibleOn": "${operator === 'in'}"
              },
              {
                "type": "input-date-range",
                "name": "value",
                "label": "Range",
                "required": true,
                "visibleOn": "${operator === 'between'}"
              },
              {
                "type": "radios",
                "name": "conjunction",
                "label": "Combine with next filter",
                "options": [
                  {"label": "AND", "value": "AND"},
                  {"label": "OR", "value": "OR"}
                ],
                "value": "AND"
              }
            ]
          }
        ]
      },
      {
        "title": "Grouping & Sorting",
        "body": [
          {
            "type": "transfer",
            "name": "grouping",
            "label": "Group By",
            "source": "/api/reports/data-sources/${data_source}/columns?groupable=true",
            "labelField": "display_name",
            "valueField": "name",
            "searchable": true,
            "sortable": true
          },
          {
            "type": "combo",
            "name": "sorting",
            "label": "Sort Order",
            "multiple": true,
            "multiLine": true,
            "addable": true,
            "removable": true,
            "draggable": true,
            "items": [
              {
                "type": "select",
                "name": "column",
                "label": "Column",
                "source": "${columns}",
                "required": true
              },
              {
                "type": "radios",
                "name": "direction",
                "label": "Direction",
                "options": [
                  {"label": "Ascending", "value": "asc"},
                  {"label": "Descending", "value": "desc"}
                ],
                "value": "asc",
                "required": true
              }
            ]
          }
        ]
      },
      {
        "title": "Display Options",
        "body": [
          {
            "type": "radios",
            "name": "output_format",
            "label": "Display As",
            "options": [
              {"label": "Table", "value": "table"},
              {"label": "Chart", "value": "chart"},
              {"label": "Pivot Table", "value": "pivot"}
            ],
            "value": "table",
            "required": true
          },
          {
            "type": "container",
            "visibleOn": "${output_format === 'chart'}",
            "body": [
              {
                "type": "select",
                "name": "chart_config.type",
                "label": "Chart Type",
                "options": [
                  {"label": "Bar Chart", "value": "bar"},
                  {"label": "Line Chart", "value": "line"},
                  {"label": "Pie Chart", "value": "pie"},
                  {"label": "Area Chart", "value": "area"}
                ],
                "required": true
              },
              {
                "type": "select",
                "name": "chart_config.x_axis",
                "label": "X-Axis",
                "source": "${columns}",
                "required": true
              },
              {
                "type": "select",
                "name": "chart_config.y_axis",
                "label": "Y-Axis",
                "source": "${columns}",
                "multiple": true,
                "required": true
              }
            ]
          }
        ]
      }
    ],
    "api": {
      "method": "post",
      "url": "/api/reports",
      "adaptor": "return {\n  ...api,\n  data: {\n    ...api.data,\n    tenant_id: window.tenantId\n  }\n};"
    },
    "redirect": "/reports/${id}"
  }
}
```

### Report Viewer UI (AMIS)

```json
{
  "type": "page",
  "title": "${report.name}",
  "subTitle": "${report.description}",
  "toolbar": [
    {
      "type": "button",
      "label": "Refresh",
      "icon": "fa fa-refresh",
      "actionType": "reload",
      "target": "report-table"
    },
    {
      "type": "dropdown-button",
      "label": "Export",
      "icon": "fa fa-download",
      "buttons": [
        {
          "type": "button",
          "label": "Excel",
          "actionType": "download",
          "api": "/api/reports/${reportId}/export?format=excel"
        },
        {
          "type": "button",
          "label": "PDF",
          "actionType": "download",
          "api": "/api/reports/${reportId}/export?format=pdf"
        },
        {
          "type": "button",
          "label": "CSV",
          "actionType": "download",
          "api": "/api/reports/${reportId}/export?format=csv"
        }
      ]
    },
    {
      "type": "button",
      "label": "Schedule",
      "icon": "fa fa-clock-o",
      "actionType": "dialog",
      "dialog": {
        "title": "Schedule Report",
        "body": {
          "type": "form",
          "api": "/api/reports/${reportId}/schedules",
          "body": [
            {
              "type": "radios",
              "name": "schedule_type",
              "label": "Frequency",
              "options": [
                {"label": "Daily", "value": "daily"},
                {"label": "Weekly", "value": "weekly"},
                {"label": "Monthly", "value": "monthly"},
                {"label": "Custom", "value": "cron"}
              ],
              "required": true
            },
            {
              "type": "input-text",
              "name": "cron_expression",
              "label": "Cron Expression",
              "visibleOn": "${schedule_type === 'cron'}",
              "placeholder": "0 9 * * *",
              "required": true
            },
            {
              "type": "select",
              "name": "timezone",
              "label": "Timezone",
              "source": "/api/timezones",
              "value": "UTC",
              "searchable": true
            },
            {
              "type": "checkboxes",
              "name": "output_formats",
              "label": "Output Formats",
              "options": [
                {"label": "PDF", "value": "pdf"},
                {"label": "Excel", "value": "excel"},
                {"label": "CSV", "value": "csv"}
              ],
              "value": ["pdf"]
            },
            {
              "type": "radios",
              "name": "delivery_method",
              "label": "Delivery Method",
              "options": [
                {"label": "Email", "value": "email"},
                {"label": "Webhook", "value": "webhook"},
                {"label": "Storage Only", "value": "storage"}
              ],
              "value": "email"
            },
            {
              "type": "input-tag",
              "name": "email_recipients",
              "label": "Email Recipients",
              "visibleOn": "${delivery_method === 'email'}",
              "placeholder": "Enter email addresses",
              "required": true
            },
            {
              "type": "input-text",
              "name": "webhook_url",
              "label": "Webhook URL",
              "visibleOn": "${delivery_method === 'webhook'}",
              "required": true
            }
          ]
        }
      }
    },
    {
      "type": "button",
      "label": "Edit",
      "icon": "fa fa-edit",
      "actionType": "link",
      "link": "/reports/${reportId}/edit"
    }
  ],
  "body": [
    {
      "type": "service",
      "api": "/api/reports/${reportId}",
      "body": [
        {
          "type": "crud",
          "name": "report-table",
          "syncLocation": false,
          "api": "/api/reports/${reportId}/data",
          "columns": "${report.columns}",
          "filter": {
            "title": "Quick Filters",
            "body": "${report.filters}"
          },
          "headerToolbar": [
            "filter-toggler",
            "reload",
            {
              "type": "columns-toggler",
              "align": "right"
            }
          ],
          "footerToolbar": [
            "statistics",
            "pagination"
          ],
          "perPage": 50,
          "perPageAvailable": [10, 25, 50, 100],
          "keepItemSelectionOnPageChange": true,
          "maxKeepItemSelectionLength": 11,
          "labelTpl": "${report.name}"
        }
      ]
    }
  ],
  "aside": {
    "type": "panel",
    "title": "Report Info",
    "body": [
      {
        "type": "property",
        "title": "Details",
        "items": [
          {"label": "Created", "content": "${report.created_at|date:YYYY-MM-DD}"},
          {"label": "Last Run", "content": "${report.last_run_at|date:YYYY-MM-DD HH:mm}"},
          {"label": "Run Count", "content": "${report.run_count}"},
          {"label": "Avg Time", "content": "${report.avg_execution_time_ms}ms"}
        ]
      },
      {
        "type": "divider"
      },
      {
        "type": "panel",
        "title": "Recent Executions",
        "body": {
          "type": "list",
          "source": "/api/reports/${reportId}/executions?limit=5",
          "listItem": {
            "body": [
              {
                "type": "tpl",
                "tpl": "<div class=\"flex justify-between\">\n  <span>${executed_by}</span>\n  <span class=\"text-sm text-gray-500\">${started_at|fromNow}</span>\n</div>\n<div class=\"text-xs ${status === 'completed' ? 'text-green-500' : 'text-red-500'}\">\n  ${status} ${execution_time_ms ? `(${execution_time_ms}ms)` : ''}\n</div>"
              }
            ]
          }
        }
      }
    ]
  }
}
```

### Report List Page (AMIS)

```json
{
  "type": "page",
  "title": "Reports",
  "body": {
    "type": "crud",
    "api": "/api/reports",
    "filter": {
      "title": "Search Reports",
      "body": [
        {
          "type": "input-text",
          "name": "search",
          "placeholder": "Search by name...",
          "clearable": true
        },
        {
          "type": "select",
          "name": "category",
          "label": "Category",
          "placeholder": "All Categories",
          "clearable": true,
          "options": [
            {"label": "Financial", "value": "financial"},
            {"label": "Operational", "value": "operational"},
            {"label": "Analytical", "value": "analytical"}
          ]
        },
        {
          "type": "select",
          "name": "data_source",
          "label": "Data Source",
          "placeholder": "All Sources",
          "clearable": true,
          "source": "/api/reports/data-sources"
        }
      ]
    },
    "headerToolbar": [
      "filter-toggler",
      {
        "type": "button",
        "label": "New Report",
        "icon": "fa fa-plus",
        "actionType": "link",
        "link": "/reports/create",
        "level": "primary"
      },
      "reload"
    ],
    "footerToolbar": ["pagination"],
    "perPage": 20,
    "columns": [
      {
        "name": "name",
        "label": "Report Name",
        "type": "text",
        "searchable": true
      },
      {
        "name": "description",
        "label": "Description",
        "type": "text",
        "breakpoint": "*"
      },
      {
        "name": "category",
        "label": "Category",
        "type": "tag",
        "breakpoint": "*"
      },
      {
        "name": "data_source",
        "label": "Data Source",
        "type": "text",
        "breakpoint": "*"
      },
      {
        "name": "run_count",
        "label": "Runs",
        "type": "number"
      },
      {
        "name": "last_run_at",
        "label": "Last Run",
        "type": "date",
        "format": "YYYY-MM-DD HH:mm"
      },
      {
        "type": "operation",
        "label": "Actions",
        "buttons": [
          {
            "type": "button",
            "label": "View",
            "icon": "fa fa-eye",
            "actionType": "link",
            "link": "/reports/${id}"
          },
          {
            "type": "button",
            "label": "Run",
            "icon": "fa fa-play",
            "actionType": "ajax",
            "api": "post:/api/reports/${id}/execute",
            "confirmText": "Run this report now?"
          },
          {
            "type": "dropdown-button",
            "label": "More",
            "buttons": [
              {
                "type": "button",
                "label": "Edit",
                "actionType": "link",
                "link": "/reports/${id}/edit"
              },
              {
                "type": "button",
                "label": "Duplicate",
                "actionType": "ajax",
                "api": "post:/api/reports/${id}/duplicate"
              },
              {
                "type": "button",
                "label": "Share",
                "actionType": "dialog",
                "dialog": {
                  "title": "Share Report",
                  "body": "Share functionality"
                }
              },
              {
                "type": "button",
                "label": "Delete",
                "level": "danger",
                "actionType": "ajax",
                "api": "delete:/api/reports/${id}",
                "confirmText": "Are you sure you want to delete this report?"
              }
            ]
          }
        ]
      }
    ]
  }
}
```

