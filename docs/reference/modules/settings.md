# Multi-Tenant ERP Configuration Service
## Complete Technical Specification

### Document Information
- **Version**: 1.0
- **Date**: September 2025
- **Author**: System Architecture Team
- **Status**: Draft for Review

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Architecture Overview](#architecture-overview)
3. [Database Design](#database-design)
4. [Security & Access Control](#security--access-control)
5. [Configuration Management](#configuration-management)
6. [Default Values System](#default-values-system)
7. [Integration Patterns](#integration-patterns)
8. [Performance & Scalability](#performance--scalability)
9. [Implementation Roadmap](#implementation-roadmap)
10. [Risk Analysis & Mitigations](#risk-analysis--mitigations)
11. [Operational Procedures](#operational-procedures)
12. [Appendices](#appendices)

---

## Executive Summary

### Recommendation
**Use PostgreSQL Row-Level Security (RLS) for the ERP configuration service** with important conditions:
- Suitable for organizations with up to 10,000 tenants
- Acceptable 10-20% performance overhead
- Strong PostgreSQL RLS operational expertise required
- Critical security isolation needs

### Key Benefits
- **Strong Security**: Database-level enforcement prevents cross-tenant data leakage
- **Comprehensive Audit**: Complete configuration change tracking and compliance support
- **Flexible Schema**: Supports evolving ERP module requirements without schema changes
- **Template System**: Accelerates tenant onboarding with industry-specific configurations
- **Default Management**: Three-tier hierarchy (system → tenant → entity) with inheritance

### Architecture Approach
The solution extends your existing sophisticated ERP system rather than replacing it, leveraging:
- Existing `tenants`, `entities`, `users` tables
- Current `modules` and IAM infrastructure  
- Established `audit_log` and `finance_*` tables
- Your proven RLS implementation patterns

---

## Architecture Overview

### High-Level Components

#### Core Service Architecture
```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   API Gateway   │────│  Config Service  │────│   IAM Service   │
│                 │    │     (Golang)     │    │   (Existing)    │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                               │
                               │
                    ┌──────────────────┐
                    │ Tenant Context   │
                    │    Manager       │
                    └──────────────────┘
                               │
                    ┌──────────────────┐    ┌─────────────────┐
                    │   Policy Engine  │────│ Audit & Monitor │
                    │                  │    │                 │
                    └──────────────────┘    └─────────────────┘
                               │
                    ┌──────────────────┐
                    │  PostgreSQL DB   │
                    │   with RLS       │
                    │  (Extended)      │
                    └──────────────────┘
```

#### Component Responsibilities

**API Gateway**
- Routes requests and extracts tenant context from JWT tokens
- Implements rate limiting per tenant
- Validates initial authentication

**Config Service (Golang)**
- Core business logic for configuration management
- Validates tenant permissions and manages CRUD operations
- Handles bulk operations, migrations, and template applications
- Integrates with existing finance modules

**Tenant Context Manager**
- Establishes database connections with proper tenant context
- Manages connection pooling per tenant context
- Handles context propagation through request lifecycle

**Policy Engine**
- Maintains RLS policy definitions and deployment
- Provides policy testing framework
- Manages admin bypass scenarios with audit trails

**Audit & Monitor**
- Captures all configuration access attempts
- Monitors RLS policy effectiveness
- Provides debugging tools for access denied scenarios

### Data Flow Patterns

#### Configuration Read Flow
1. **Request** → API Gateway validates JWT and extracts tenant context
2. **Service Layer** → Validates user permissions for tenant/module
3. **Context Setup** → Sets `app.current_tenant` session variable
4. **Database Query** → RLS policies filter results to tenant data
5. **Default Resolution** → Merges with system/tenant defaults if needed
6. **Response** → Returns configured values with metadata

#### Configuration Write Flow
1. **Request Validation** → Validates schema and business rules
2. **Dependency Check** → Validates configuration dependencies
3. **Change Request** → Creates approval workflow if required
4. **Transaction** → Applies changes with proper audit logging
5. **Propagation** → Updates dependent configurations and caches
6. **Notification** → Triggers relevant stakeholder notifications

---

## Database Design

### Extension Strategy

Rather than replacing your existing sophisticated system, we extend it strategically:

**Existing Infrastructure (Leveraged)**
- ✅ Core tenant/entity structure (`tenants`, `entities`)
- ✅ Robust IAM system (`roles`, `permissions`, `user_roles`)
- ✅ Basic configuration (`tenant_configurations`, `feature_flags`)
- ✅ Audit infrastructure (`audit_log`, `user_activities`)
- ✅ Finance module (`finance_accounts`, `finance_transactions`)
- ✅ Flexible attribute system (`attribute_definitions`, `attribute_values`)
- ✅ Hierarchy management (`hierarchy_paths`)
- ✅ Module system (`modules`)

### New Configuration Tables

#### 1. Configuration Schemas
```sql
CREATE TABLE config_schemas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID, -- NULL = system-wide schema
    module_id UUID NOT NULL, -- References existing modules
    schema_name VARCHAR(100) NOT NULL,
    schema_version INTEGER NOT NULL DEFAULT 1,
    json_schema JSONB NOT NULL, -- JSON Schema validation
    default_values JSONB, -- Default configuration values
    is_system_schema BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID NOT NULL,
    
    UNIQUE(tenant_id, module_id, schema_name, schema_version),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    FOREIGN KEY (module_id) REFERENCES modules(id),
    FOREIGN KEY (created_by) REFERENCES users(id)
);
```

**Purpose**: Defines structure and validation rules for configuration values, enabling schema evolution without data migration.

#### 2. Extended Configuration Values
```sql
CREATE TABLE config_values_extended (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    entity_id UUID, -- NULL = tenant-wide config
    module_id UUID NOT NULL,
    config_key VARCHAR(200) NOT NULL,
    config_value JSONB NOT NULL,
    schema_id UUID, -- References config_schemas for validation
    environment VARCHAR(20) DEFAULT 'production',
    effective_from TIMESTAMPTZ DEFAULT NOW(),
    effective_until TIMESTAMPTZ,
    priority INTEGER DEFAULT 0, -- For inheritance resolution
    source_type VARCHAR(50) DEFAULT 'manual', -- 'manual', 'template', 'inherited', 'system'
    source_id UUID, -- Reference to template or parent config
    is_default_derived BOOLEAN DEFAULT FALSE,
    default_source VARCHAR(50), -- 'system', 'tenant', 'entity', 'template', 'manual'
    default_source_id UUID, -- Reference to source record
    version INTEGER DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,
    
    UNIQUE(tenant_id, entity_id, module_id, config_key, environment, effective_from),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    FOREIGN KEY (entity_id) REFERENCES entities(id),
    FOREIGN KEY (module_id) REFERENCES modules(id),
    FOREIGN KEY (schema_id) REFERENCES config_schemas(id),
    FOREIGN KEY (created_by) REFERENCES users(id),
    FOREIGN KEY (updated_by) REFERENCES users(id)
);
```

**Purpose**: Stores actual configuration values with support for time-based changes, inheritance, templates, and multi-environment configurations.

#### 3. Configuration Templates
```sql
CREATE TABLE config_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID, -- NULL = global/system template
    name VARCHAR(200) NOT NULL,
    description TEXT,
    module_id UUID NOT NULL,
    template_type VARCHAR(50) NOT NULL, -- 'starter', 'industry', 'custom'
    base_template_id UUID, -- For template inheritance
    is_system_template BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    tags TEXT[],
    metadata JSONB, -- Template-specific metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID NOT NULL,
    
    UNIQUE(tenant_id, module_id, name),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    FOREIGN KEY (module_id) REFERENCES modules(id),
    FOREIGN KEY (base_template_id) REFERENCES config_templates(id),
    FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE TABLE config_template_values (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL,
    config_key VARCHAR(200) NOT NULL,
    config_value JSONB NOT NULL,
    is_required BOOLEAN DEFAULT FALSE,
    validation_rules JSONB,
    display_order INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(template_id, config_key),
    FOREIGN KEY (template_id) REFERENCES config_templates(id) ON DELETE CASCADE
);
```

**Purpose**: Enables configuration reuse through templates, supporting industry-specific and custom configuration patterns.

#### 4. Configuration Dependencies
```sql
CREATE TABLE config_dependencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    source_config_key VARCHAR(200) NOT NULL,
    source_module_id UUID NOT NULL,
    target_config_key VARCHAR(200) NOT NULL,
    target_module_id UUID NOT NULL,
    dependency_type VARCHAR(50) NOT NULL, -- 'requires', 'conflicts', 'implies', 'excludes'
    validation_rule JSONB,
    error_message TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID NOT NULL,
    
    UNIQUE(tenant_id, source_module_id, source_config_key, target_module_id, target_config_key, dependency_type),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    FOREIGN KEY (source_module_id) REFERENCES modules(id),
    FOREIGN KEY (target_module_id) REFERENCES modules(id),
    FOREIGN KEY (created_by) REFERENCES users(id)
);
```

**Purpose**: Models complex ERP configuration relationships and prevents invalid configuration combinations.

#### 5. Configuration Change Requests
```sql
CREATE TABLE config_change_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    entity_id UUID,
    config_value_id UUID,
    change_type VARCHAR(50) NOT NULL, -- 'create', 'update', 'delete', 'bulk_update'
    current_value JSONB,
    proposed_value JSONB NOT NULL,
    change_reason TEXT NOT NULL,
    business_justification TEXT,
    risk_assessment VARCHAR(20), -- 'low', 'medium', 'high'
    status VARCHAR(20) DEFAULT 'draft', -- 'draft', 'submitted', 'approved', 'rejected', 'applied', 'cancelled'
    scheduled_for TIMESTAMPTZ,
    auto_approve BOOLEAN DEFAULT FALSE,
    
    -- Integration with existing IAM
    requested_by UUID NOT NULL,
    requested_at TIMESTAMPTZ DEFAULT NOW(),
    assigned_to UUID,
    reviewed_by UUID,
    reviewed_at TIMESTAMPTZ,
    review_notes TEXT,
    applied_by UUID,
    applied_at TIMESTAMPTZ,
    
    FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    FOREIGN KEY (entity_id) REFERENCES entities(id),
    FOREIGN KEY (config_value_id) REFERENCES config_values_extended(id),
    FOREIGN KEY (requested_by) REFERENCES users(id),
    FOREIGN KEY (assigned_to) REFERENCES users(id),
    FOREIGN KEY (reviewed_by) REFERENCES users(id),
    FOREIGN KEY (applied_by) REFERENCES users(id)
);
```

**Purpose**: Provides approval workflows for sensitive configuration changes with comprehensive audit trails.

#### 6. Enhanced Configuration Audit
```sql
CREATE TABLE config_audit_details (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    audit_log_id UUID, -- References existing audit_log if needed
    tenant_id UUID NOT NULL,
    entity_id UUID,
    config_id UUID,
    operation_type VARCHAR(20) NOT NULL, -- 'read', 'create', 'update', 'delete'
    config_key VARCHAR(200) NOT NULL,
    old_value JSONB,
    new_value JSONB,
    change_source VARCHAR(50), -- 'manual', 'template_apply', 'bulk_operation', 'api', 'migration'
    change_request_id UUID,
    session_context JSONB,
    performed_at TIMESTAMPTZ DEFAULT NOW(),
    performed_by UUID NOT NULL,
    
    FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    FOREIGN KEY (entity_id) REFERENCES entities(id),
    FOREIGN KEY (config_id) REFERENCES config_values_extended(id),
    FOREIGN KEY (change_request_id) REFERENCES config_change_requests(id),
    FOREIGN KEY (performed_by) REFERENCES users(id)
);
```

**Purpose**: Complements existing audit system with configuration-specific audit details and compliance support.

---

## Security & Access Control

### Row-Level Security Implementation

#### Core RLS Policies

**1. Tenant Isolation Policy**
```sql
CREATE POLICY tenant_isolation_policy ON config_values_extended
    FOR ALL 
    TO application_role
    USING (
        tenant_id = current_setting('app.current_tenant')::uuid
        OR current_setting('app.admin_mode', true)::boolean = true
    );
```

**2. Admin Bypass Policy**
```sql
CREATE POLICY admin_access_policy ON config_values_extended 
    FOR ALL 
    TO admin_role 
    USING (true);
```

**3. Module-Scoped Access Policy**
```sql
CREATE POLICY module_access_policy ON config_values_extended
    FOR ALL
    TO application_role
    USING (
        tenant_id = current_setting('app.current_tenant')::uuid
        AND has_module_permission(module_id, current_setting('app.current_user')::uuid)
    );
```

#### Context Propagation Strategy

**Database Session Context**
- Every database connection sets `app.current_tenant` using `set_config()`
- Context automatically cleared when connection returns to pool
- Validates tenant context exists before allowing operations

**Application Layer Validation**
- Validates tenant access before database calls
- Database RLS provides final enforcement layer
- Comprehensive audit logging captures all access attempts

### Access Control Integration

**Permission Model**
- Leverages existing `roles` and `permissions` tables
- Configuration permissions integrate with `role_permissions` system
- Fine-grained access control using pattern matching

**Module-Based Security**
- Configuration access tied to module permissions
- Cross-module dependencies require appropriate permissions
- Admin roles can manage cross-tenant configurations

---

## Configuration Management

### Schema Management

#### Configuration Schema Lifecycle
1. **Schema Definition** → JSON Schema-based validation rules
2. **Version Management** → Support for schema evolution
3. **Default Values** → System and tenant-specific defaults
4. **Validation** → Runtime validation against defined schemas

#### Template System

**Template Types**
- **System Templates**: Pre-built industry configurations
- **Tenant Templates**: Custom tenant-specific templates  
- **Entity Templates**: Entity-level configuration patterns

**Template Application Process**
1. **Selection** → Choose appropriate template for tenant/entity
2. **Customization** → Modify template values for specific needs
3. **Validation** → Ensure template values meet schema requirements
4. **Application** → Apply template configurations with audit trail

### Configuration Inheritance

#### Inheritance Hierarchy
```
System Defaults (Global)
    ↓ (inherits/overrides)
Tenant Defaults 
    ↓ (inherits/overrides)  
Entity Defaults
    ↓ (inherits/overrides)
Actual Configuration Values
```

#### Merge Strategies
- **Replace**: Complete replacement of parent configuration
- **Shallow Merge**: Top-level property merge
- **Deep Merge**: Recursive merge of nested objects
- **Selective Override**: Override specific properties while inheriting others

---

## Default Values System

### Three-Tier Default Architecture

#### 1. System Defaults
```sql
CREATE TABLE config_system_defaults (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    module_id UUID NOT NULL,
    schema_id UUID NOT NULL,
    config_key VARCHAR(200) NOT NULL,
    default_value JSONB NOT NULL,
    is_required BOOLEAN DEFAULT FALSE,
    is_overridable BOOLEAN DEFAULT TRUE,
    override_scope VARCHAR(20) DEFAULT 'full', -- 'none', 'partial', 'full'
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID NOT NULL,
    
    UNIQUE(module_id, config_key),
    FOREIGN KEY (module_id) REFERENCES modules(id),
    FOREIGN KEY (schema_id) REFERENCES config_schemas(id),
    FOREIGN KEY (created_by) REFERENCES users(id)
);
```

**Purpose**: Provides baseline configurations for all tenants, ensuring consistent ERP behavior.

#### 2. Tenant Defaults
```sql
CREATE TABLE config_tenant_defaults (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    module_id UUID NOT NULL,
    config_key VARCHAR(200) NOT NULL,
    default_value JSONB NOT NULL,
    inherit_from_system BOOLEAN DEFAULT TRUE,
    merge_strategy VARCHAR(20) DEFAULT 'deep_merge', -- 'replace', 'shallow_merge', 'deep_merge'
    applies_to_entities BOOLEAN DEFAULT TRUE,
    environment VARCHAR(20) DEFAULT 'production',
    effective_from TIMESTAMPTZ DEFAULT NOW(),
    effective_until TIMESTAMPTZ,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID NOT NULL,
    updated_by UUID,
    
    UNIQUE(tenant_id, module_id, config_key, environment, effective_from),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    FOREIGN KEY (module_id) REFERENCES modules(id),
    FOREIGN KEY (created_by) REFERENCES users(id),
    FOREIGN KEY (updated_by) REFERENCES users(id)
);
```

**Purpose**: Allows tenants to customize default behaviors while inheriting system baselines.

#### 3. Default Resolution Function
```sql
CREATE OR REPLACE FUNCTION resolve_config_defaults(
    p_tenant_id UUID,
    p_entity_id UUID DEFAULT NULL,
    p_module_name VARCHAR DEFAULT NULL,
    p_config_key VARCHAR DEFAULT NULL,
    p_environment VARCHAR DEFAULT 'production',
    p_as_of_date TIMESTAMPTZ DEFAULT NOW()
) RETURNS TABLE (
    config_key VARCHAR,
    resolved_value JSONB,
    source_hierarchy JSONB,
    final_source VARCHAR
) AS $$
-- [Implementation details in full specification]
$$ LANGUAGE plpgsql;
```

**Purpose**: Resolves configuration values through the default hierarchy with proper inheritance and merge strategies.

---

## Integration Patterns

### Finance Module Integration

#### Document Sequence Integration

**Enhanced Document Number Generation**
```sql
CREATE OR REPLACE FUNCTION generate_document_number_with_defaults(
    p_tenant_id UUID,
    p_entity_id UUID,
    p_document_type VARCHAR,
    p_date DATE DEFAULT CURRENT_DATE
) RETURNS VARCHAR AS $$
-- [Full implementation with default resolution]
$$ LANGUAGE plpgsql;
```

**Configuration Schema for Document Sequences**
```json
{
    "type": "object",
    "properties": {
        "document_type": {"type": "string", "enum": ["invoice", "payment", "receipt", "journal", "purchase_order"]},
        "prefix": {"type": "string", "maxLength": 20},
        "suffix": {"type": "string", "maxLength": 20},
        "current_number": {"type": "integer", "minimum": 1},
        "increment_by": {"type": "integer", "minimum": 1, "maximum": 1000},
        "pad_length": {"type": "integer", "minimum": 1, "maximum": 20},
        "reset_frequency": {"type": "string", "enum": ["never", "yearly", "monthly", "daily"]},
        "format_template": {"type": "string"}
    },
    "required": ["document_type", "current_number", "increment_by", "pad_length"]
}
```

#### Migration from Existing Tables
```sql
CREATE OR REPLACE FUNCTION migrate_document_sequences_to_config() 
RETURNS VOID AS $$
-- [Migration logic to sync existing finance_document_sequences]
$$ LANGUAGE plpgsql;
```

### Other Module Integration Patterns

#### HR Module Configuration
- Employee onboarding workflows
- Payroll calculation rules
- Leave policies and accrual rules
- Performance review configurations

#### Inventory Module Configuration
- Stock valuation methods
- Reorder point calculations
- Warehouse location settings
- Product categorization rules

#### IAM Module Configuration
- Password policies
- Session timeout settings
- MFA requirements
- Role assignment rules

---

## Performance & Scalability

### Performance Optimization Strategies

#### 1. Indexing Strategy
```sql
-- Primary lookup patterns
CREATE INDEX idx_config_values_tenant_module ON config_values_extended (tenant_id, module_id, config_key);
CREATE INDEX idx_config_values_entity_key ON config_values_extended (tenant_id, entity_id, config_key);
CREATE INDEX idx_config_values_effective ON config_values_extended (tenant_id, effective_from, effective_until) 
    WHERE effective_until IS NOT NULL;

-- JSON path queries
CREATE INDEX idx_config_values_json_paths ON config_values_extended USING GIN (config_value jsonb_path_ops);

-- Default resolution performance
CREATE INDEX idx_config_values_default_source ON config_values_extended (default_source, default_source_id);
```

#### 2. Materialized View for Default Resolution
```sql
CREATE MATERIALIZED VIEW mv_resolved_defaults AS
SELECT 
    tenant_id, entity_id, module_name, config_key,
    resolved_value, final_source, last_updated
FROM [complex default resolution query]
WITH DATA;

CREATE UNIQUE INDEX idx_mv_resolved_defaults_lookup 
ON mv_resolved_defaults (tenant_id, entity_id, module_name, config_key);
```

#### 3. Connection Pooling Strategy
- **Per-Tenant Context**: Dedicated connection pools per tenant context
- **Session Reuse**: Efficient session variable management
- **Context Validation**: Middleware ensures proper tenant context

### Scalability Targets

#### Performance Benchmarks
- **Read Latency**: 95th percentile < 100ms for typical configuration queries
- **Write Latency**: 95th percentile < 200ms for configuration updates
- **Throughput**: Support 1000+ concurrent tenant operations
- **RLS Overhead**: <20% performance impact from RLS policies

#### Scaling Strategies
- **Horizontal Read Scaling**: Read replicas for configuration queries
- **Caching Layer**: Redis caching for frequently accessed configurations
- **Query Optimization**: Tenant-aware query plans and statistics
- **Partitioning**: Consider tenant-based partitioning for large deployments

---

## Implementation Roadmap

### Phase 1: Foundation (Months 1-2)
**Objectives**: Establish core configuration infrastructure

**Deliverables**:
- [ ] Core configuration tables (schemas, values, templates)
- [ ] Basic RLS policies and tenant context management
- [ ] Integration with existing IAM system
- [ ] Initial audit logging and monitoring
- [ ] Simple configuration CRUD operations

**Success Criteria**:
- Configuration values can be stored and retrieved securely
- Tenant isolation is enforced at database level
- Basic audit trail is captured

### Phase 2: Default Management System (Months 2-3)
**Objectives**: Implement comprehensive default management

**Deliverables**:
- [ ] System and tenant default tables
- [ ] Default resolution function with inheritance
- [ ] Template system with industry patterns
- [ ] Migration tools for existing configurations
- [ ] Default coverage reporting

**Success Criteria**:
- New tenants can be onboarded with appropriate defaults
- Configuration inheritance works correctly
- Templates can be applied and customized

### Phase 3: Finance Module Integration (Months 3-4)
**Objectives**: Full integration with existing finance module

**Deliverables**:
- [ ] Document sequence configuration integration
- [ ] Enhanced document number generation
- [ ] Migration from existing `finance_document_sequences`
- [ ] Configuration dependencies for finance workflows
- [ ] Approval workflows for sensitive financial configurations

**Success Criteria**:
- Finance document sequences work with new configuration system
- No disruption to existing finance operations
- Enhanced configuration capabilities are available

### Phase 4: Advanced Features (Months 4-5)
**Objectives**: Add sophisticated configuration management features

**Deliverables**:
- [ ] Configuration change request workflows
- [ ] Cross-module dependency management
- [ ] Time-based configuration changes
- [ ] Bulk configuration operations
- [ ] Configuration validation and testing framework

**Success Criteria**:
- Complex configuration changes can be managed through approval workflows
- Configuration dependencies prevent invalid combinations
- Bulk operations maintain data integrity

### Phase 5: Performance & Operations (Months 5-6)
**Objectives**: Optimize performance and establish operational procedures

**Deliverables**:
- [ ] Performance optimization and caching
- [ ] Comprehensive monitoring and alerting
- [ ] Backup and recovery procedures
- [ ] Configuration import/export tools
- [ ] Documentation and training materials

**Success Criteria**:
- System meets performance benchmarks
- Operations team can manage the system effectively
- Disaster recovery procedures are tested and documented

### Phase 6: Additional Modules (Months 6-12)
**Objectives**: Extend to other ERP modules

**Deliverables**:
- [ ] HR module configuration integration
- [ ] Inventory module configuration integration
- [ ] IAM module configuration enhancement
- [ ] Custom module configuration framework
- [ ] Advanced reporting and analytics

**Success Criteria**:
- All major ERP modules use centralized configuration
- Configuration consistency across all modules
- Advanced configuration analytics available

---

## Risk Analysis & Mitigations

### High-Priority Risks

#### 1. RLS Policy Bugs
**Risk**: Incorrect RLS policies could cause cross-tenant data leakage
- **Impact**: Critical - Data security breach
- **Probability**: Medium
- **Mitigation**: 
  - Automated RLS policy testing in CI/CD
  - Comprehensive security testing with tenant context variations
  - Staged deployment with security validation
  - Regular security audits and penetration testing

#### 2. Performance Degradation
**Risk**: RLS overhead could significantly impact query performance
- **Impact**: Medium - User experience degradation
- **Probability**: Medium
- **Mitigation**:
  - Continuous performance monitoring with alerting
  - Query optimization and explain plan analysis
  - Materialized view caching for complex queries
  - Escape hatches for critical performance paths

#### 3. Context Injection Failures
**Risk**: Tenant context might not be properly set, causing access issues
- **Impact**: High - Service disruption
- **Probability**: Low
- **Mitigation**:
  - Connection middleware validation with circuit breakers
  - Comprehensive context propagation testing
  - Fallback mechanisms for context failures
  - Real-time context monitoring and alerting

### Medium-Priority Risks

#### 4. Configuration Migration Issues
**Risk**: Data corruption during migration from existing systems
- **Impact**: High - Data integrity issues
- **Probability**: Low
- **Mitigation**:
  - Extensive testing with production data copies
  - Staged migration with rollback procedures
  - Comprehensive validation at each migration step
  - Backup verification before and after migration

#### 5. Default Resolution Performance
**Risk**: Complex default resolution could impact response times
- **Impact**: Medium - Degraded user experience
- **Probability**: Medium
- **Mitigation**:
  - Materialized view caching for common default patterns
  - Query optimization and indexing strategies
  - Lazy default resolution where appropriate
  - Performance benchmarking and monitoring

#### 6. Operational Complexity
**Risk**: Increased operational burden for database administration
- **Impact**: Medium - Higher operational costs
- **Probability**: Medium
- **Mitigation**:
  - Comprehensive documentation and runbooks
  - Automated operational procedures
  - Training for operations team
  - Monitoring and alerting for common issues

---

## Operational Procedures

### Monitoring & Alerting

#### Key Metrics to Monitor
- **RLS Policy Effectiveness**: Cross-tenant access attempt alerts
- **Configuration Access Patterns**: Unusual access pattern detection
- **Performance Metrics**: Query latency and throughput tracking
- **Default Resolution Performance**: Default lookup timing
- **Audit Trail Completeness**: Missing audit entries detection

#### Alert Thresholds
- **Security**: Any cross-tenant data access attempts
- **Performance**: 95th percentile latency > 200ms
- **Errors**: Configuration validation failures > 5%
- **Availability**: Service response time > 500ms

### Backup & Recovery

#### Backup Strategy
- **Configuration Data**: Point-in-time recovery for all configuration tables
- **Tenant-Specific Recovery**: Ability to restore individual tenant configurations
- **Schema Versioning**: Backup and recovery of configuration schemas
- **Audit Trail Preservation**: Long-term audit log retention

#### Recovery Procedures
1. **Individual Configuration Recovery**: Restore specific configuration values
2. **Tenant Data Recovery**: Complete tenant configuration restoration
3. **Schema Rollback**: Revert schema changes with data migration
4. **Disaster Recovery**: Full system restoration from backups

### Maintenance Procedures

#### Regular Maintenance Tasks
- **Default Cache Refresh**: Automated materialized view maintenance
- **Audit Log Archival**: Automated old audit data archival
- **Performance Analysis**: Monthly query performance review
- **Security Review**: Quarterly RLS policy effectiveness review

#### Schema Migration Procedures
1. **Schema Change Preparation**: Validate against existing configurations
2. **Staged Deployment**: Deploy schema changes incrementally
3. **Data Migration**: Migrate existing data to new schema
4. **Validation**: Verify data integrity and functionality
5. **Rollback Plan**: Prepared rollback procedures if issues occur

### Testing & Validation

#### Automated Test Suite
```yaml
RLS Policy Tests:
  - Tenant isolation verification
  - Admin bypass functionality
  - Null context handling
  - Cross-tenant access prevention
  - Module-scoped access validation

Configuration Tests:
  - Schema validation
  - Default resolution accuracy
  - Template application
  - Dependency enforcement
  - Change request workflows

Integration Tests:
  - Finance module integration
  - Document sequence generation
  - Cross-module dependencies
  - Bulk operations
  - Performance benchmarks
```

#### Security Testing Checklist
- [ ] Cross-tenant data access attempts (should fail)
- [ ] JWT token manipulation testing
- [ ] SQL injection attempts with RLS bypass
- [ ] Admin privilege escalation testing
- [ ] Context injection failure scenarios

### Deployment Procedures

#### Deployment Checklist
- [ ] Schema migration scripts tested
- [ ] RLS policies validated in staging
- [ ] Performance benchmarks meet targets
- [ ] Security tests pass completely
- [ ] Audit logging is functional
- [ ] Default resolution works correctly
- [ ] Integration with existing modules validated
- [ ] Rollback procedures tested

---

## Appendices

### A. Configuration Schema Examples

#### Document Sequence Schema
```json
{
  "type": "object",
  "properties": {
    "document_type": {
      "type": "string",
      "enum": ["invoice", "payment", "receipt", "journal", "purchase_order"]
    },
    "prefix": {"type": "string", "maxLength": 20},
    "suffix": {"type": "string", "maxLength": 20},
    "current_number": {"type": "integer", "minimum": 1},
    "increment_by": {"type": "integer", "minimum": 1, "maximum": 1000},
    "pad_length": {"type": "integer", "minimum": 1, "maximum": 20},
    "reset_frequency": {
      "type": "string", 
      "enum": ["never", "yearly", "monthly", "daily"]
    },
    "format_template": {"type": "string"}
  },
  "required": ["document_type", "current_number", "increment_by", "pad_length"],
  "additionalProperties": false
}
```

#### HR Payroll Configuration Schema
```json
{
  "type": "object",
  "properties": {
    "pay_frequency": {
      "type": "string",
      "enum": ["weekly", "biweekly", "monthly", "semimonthly"]
    },
    "overtime_rules": {
      "type": "object",
      "properties": {
        "enabled": {"type": "boolean"},
        "threshold_hours": {"type": "number", "minimum": 0},
        "multiplier": {"type": "number", "minimum": 1}
      }
    },
    "tax_settings": {
      "type": "object",
      "properties": {
        "jurisdiction": {"type": "string"},
        "tax_tables": {"type": "array"}
      }
    }
  }
}
```

### B. Template Examples

#### Standard Manufacturing Company Template
```sql
INSERT INTO config_templates (name, description, module_id, template_type, is_system_template)
VALUES ('Standard Manufacturing', 'Configuration for manufacturing companies', 
        (SELECT id FROM modules WHERE name = 'finance'), 'industry', true);

INSERT INTO config_template_values VALUES
('invoice_sequence', '{
  "prefix": "INV-",
  "reset_frequency": "yearly",
  "pad_length": 6
}', true),
('purchase_order_sequence', '{
  "prefix": "PO-", 
  "reset_frequency": "never",
  "pad_length": 8
}', true);
```

### C. Performance Benchmarks

#### Expected Performance Targets
- **Simple Configuration Read**: < 50ms (95th percentile)
- **Complex Default Resolution**: < 100ms (95th percentile)  
- **Configuration Update**: < 200ms (95th percentile)
- **Bulk Template Application**: < 5s for 100 configurations
- **Tenant Onboarding**: < 30s for complete setup

### D. Troubleshooting Guide

#### Common Issues

**Issue**: Cross-tenant data visible
- **Cause**: RLS policy not applied or tenant context not set
- **Resolution**: Verify `app.current_tenant` is set, check RLS policy

**Issue**: Performance degradation
- **Cause**: Missing indexes or inefficient RLS policies
- **Resolution**: Analyze query plans, optimize indexes, review policies

**Issue**: Default resolution errors
- **Cause**: Circular dependencies or invalid merge strategies
- **Resolution**: Check dependency graph, validate merge logic

### E. Security Compliance

#### Compliance Requirements Met
- **SOC 2 Type II**: Audit logging and access controls
- **GDPR**: Data isolation and right to deletion
- **SOX**: Financial configuration change controls
- **ISO 27001**: Information security management

#### Audit Trail Requirements
- **Configuration Changes**: Complete change history with user attribution
- **Access Logging**: All configuration access attempts logged
- **Administrative Actions**: Admin bypass usage tracked
- **Schema Changes**: Schema evolution audit trail

---

## Document Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | Sep 2025 | Architecture Team | Initial complete specification |

---

**End of Document**
