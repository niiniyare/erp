# ERP Configuration & Settings Module - Architecture Guide

**Version**: 1.0  
**Date**: September 2025  
**Status**: Technical Specification  

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Domain-Driven Design](#domain-driven-design)
3. [Data Layer Architecture](#data-layer-architecture)
4. [Service Layer Design](#service-layer-design)
5. [API Layer Implementation](#api-layer-implementation)
6. [Security Architecture](#security-architecture)
7. [Performance Architecture](#performance-architecture)
8. [Integration Patterns](#integration-patterns)
9. [Caching Strategy](#caching-strategy)
10. [Monitoring & Observability](#monitoring-observability)

---

## Architecture Overview

### **Clean Architecture Implementation**

The ERP Settings Module follows **Clean Architecture** principles with **Configuration-as-Code** patterns, ensuring flexible configuration management while maintaining system integrity and operational simplicity.

```
┌─────────────────────────────────────────────────────────────────┐
│                    Configuration Sources                       │
│  Templates, External Config, Environment Variables           │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      API Layer (Adapters)                      │
│  @internal/api/handlers/settings_handler.go                   │
│  • Goa-generated REST/gRPC APIs                               │
│  • Configuration validation and transformation                │
│  • Template management endpoints                              │
│  • Bulk operations support                                    │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Application Layer (Ports)                    │
│  @internal/core/settings/service/                             │
│  • Configuration resolution orchestration                     │
│  • Template application logic                                 │
│  • IAM permission enforcement                                 │
│  • Bulk operation coordination                                │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Domain Layer (Core)                        │
│  @internal/core/settings/domain/                              │
│  • Configuration entities and aggregates                      │
│  • Inheritance resolution logic                               │
│  • Template processing rules                                  │
│  • Validation and business rules                              │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                  Infrastructure Layer (Adapters)               │
│  @internal/core/settings/repository/                          │
│  • Configuration persistence (SQLC)                           │
│  • Cache integration (Redis)                                  │
│  • Template storage                                           │
│  • Event publishing                                           │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Database Layer                            │
│  PostgreSQL with Row-Level Security                           │
│  • Multi-tenant data isolation                                │
│  • Configuration inheritance views                            │
│  • Template versioning                                        │
│  • Audit trail preservation                                   │
└─────────────────────────────────────────────────────────────────┘
```

### **Key Architectural Principles**

#### **1. Configuration Inheritance**
```go
// Three-level inheritance: System → Tenant → Entity
// Lower levels override higher levels with explicit inheritance tracking

// ✅ Correct: Service resolves inheritance chain
type ConfigurationResolver struct {
    systemConfig ConfigRepository
    tenantConfig TenantConfigRepository  
    entityConfig EntityRepository
    cache        CacheService
}

func (r *ConfigurationResolver) Resolve(ctx context.Context, req ResolutionRequest) (*ResolvedConfiguration, error) {
    // Build inheritance chain from most specific to most general
    chain := r.buildInheritanceChain(ctx, req.TenantID, req.EntityID, req.ConfigKey)
    
    // Apply resolution logic with source tracking
    return r.applyInheritance(chain), nil
}
```

#### **2. Template-Driven Configuration**
```go
// Templates provide structured configuration sets for rapid deployment

// ✅ Template application preserves existing customizations
type TemplateApplicator struct {
    repo       ConfigRepository
    validator  TemplateValidator
    merger     ConfigMerger
}

func (t *TemplateApplicator) Apply(ctx context.Context, templateID string, targetID string) (*ApplicationResult, error) {
    // Load template and target configurations
    template := t.repo.GetTemplate(ctx, templateID)
    existing := t.repo.GetConfigurations(ctx, targetID)
    
    // Merge with conflict resolution
    merged := t.merger.MergeWithPreservation(template, existing)
    
    return t.repo.ApplyConfigurations(ctx, targetID, merged)
}
```

#### **3. Type-Safe Configuration**
```go
// Critical settings use dedicated columns, flexible settings use JSONB

// ✅ Strongly typed for critical configurations
type TenantConfiguration struct {
    // Critical settings - dedicated columns
    DefaultCurrency     Currency    `db:"default_currency"`
    FiscalYearStart     int         `db:"fiscal_year_start_month"`
    AccountingMethod    string      `db:"accounting_method"`
    
    // Flexible settings - JSONB
    ModuleSettings      ModuleConfig `db:"settings" json:"settings"`
    FeatureFlags        FeatureFlags `db:"feature_flags" json:"feature_flags"`
}

type ModuleConfig struct {
    Finance   *FinanceConfig   `json:"finance,omitempty"`
    HR        *HRConfig        `json:"hr,omitempty"`
    Inventory *InventoryConfig `json:"inventory,omitempty"`
}
```

---

## Domain-Driven Design

### **Aggregate Design**

#### **Configuration Aggregate**
```go
// @internal/core/settings/domain/configuration.go

// Configuration is an aggregate root managing setting inheritance
type Configuration struct {
    // Identity
    ID       ConfigurationID
    TenantID tenant.ID
    EntityID *entity.ID // Nil for tenant-level configs
    
    // Configuration Scope
    Module    ModuleName
    ConfigKey ConfigKey
    
    // Value and Metadata
    Value         ConfigValue
    Source        ConfigSource     // system, tenant, entity, template
    IsInherited   bool             // true if inherited from parent level
    OverridesFrom *ConfigSource    // tracks what this overrides
    
    // Validation and Rules
    DataType      DataType
    ValidationRules ValidationRules
    
    // Permissions
    RequiredPermissions []Permission
    RequiredFeatureFlag *FeatureFlag
    
    // Metadata
    CreatedAt time.Time
    UpdatedAt time.Time
    Version   int64 // Optimistic locking
}

// Domain invariants enforced at aggregate boundary
func (c *Configuration) Validate() error {
    if c.Module == "" {
        return ErrModuleRequired
    }
    if c.ConfigKey == "" {
        return ErrConfigKeyRequired
    }
    if !c.Value.IsValidForType(c.DataType) {
        return ErrInvalidValueForType
    }
    if c.IsInherited && c.Source != ConfigSourceSystem {
        // Inherited values must have parent source
        if c.OverridesFrom == nil {
            return ErrInheritedValueMustHaveParent
        }
    }
    return nil
}

// Business operations
func (c *Configuration) Override(newValue ConfigValue, source ConfigSource) error {
    if !c.CanBeOverridden() {
        return ErrConfigurationNotOverridable
    }
    
    oldSource := c.Source
    c.Value = newValue
    c.Source = source
    c.IsInherited = false
    c.OverridesFrom = &oldSource
    c.UpdatedAt = time.Now()
    c.Version++
    
    return c.Validate()
}

func (c *Configuration) ResetToInherited(parentValue ConfigValue, parentSource ConfigSource) error {
    if c.Source == ConfigSourceSystem {
        return ErrCannotResetSystemConfiguration
    }
    
    c.Value = parentValue
    c.Source = parentSource
    c.IsInherited = true
    c.OverridesFrom = nil
    c.UpdatedAt = time.Now()
    c.Version++
    
    return nil
}
```

#### **Template Aggregate**
```go
// @internal/core/settings/domain/template.go

// Template is an aggregate root for configuration templates
type Template struct {
    // Identity
    ID        TemplateID
    Name      string
    Category  TemplateCategory // industry, functional, regional
    
    // Content
    Configurations []TemplateConfiguration
    Dependencies   []TemplateDependency
    
    // Metadata
    Description string
    Version     string
    IsActive    bool
    
    // Application Rules
    ApplicableToTenantTypes []TenantType
    RequiredFeatureFlags    []FeatureFlag
    ConflictResolution      ConflictStrategy
    
    // Audit
    CreatedAt time.Time
    UpdatedAt time.Time
    CreatedBy identity.UserID
}

type TemplateConfiguration struct {
    Module         ModuleName
    ConfigKey      ConfigKey
    Value          ConfigValue
    OverridePolicy OverridePolicy // preserve, replace, merge
    Priority       int
}

// Template application logic
func (t *Template) Apply(ctx context.Context, target ConfigurationTarget) (*ApplicationResult, error) {
    if !t.IsActive {
        return nil, ErrTemplateNotActive
    }
    
    // Validate target compatibility
    if err := t.validateTarget(target); err != nil {
        return nil, err
    }
    
    result := &ApplicationResult{
        TemplateID: t.ID,
        TargetID:   target.ID,
        Applied:    make([]ConfigurationChange, 0),
        Conflicts:  make([]ConfigurationConflict, 0),
    }
    
    // Apply configurations in priority order
    for _, config := range t.getSortedConfigurations() {
        change, conflict := t.applyConfiguration(config, target)
        
        if conflict != nil {
            result.Conflicts = append(result.Conflicts, *conflict)
        } else {
            result.Applied = append(result.Applied, change)
        }
    }
    
    return result, nil
}
```

### **Value Objects**

#### **Configuration Value Object**
```go
// @internal/core/settings/domain/config_value.go

// ConfigValue is immutable value object handling different data types
type ConfigValue struct {
    Raw      interface{}
    DataType DataType
}

// Value object constructor with validation
func NewConfigValue(value interface{}, dataType DataType) (ConfigValue, error) {
    if !isValidForType(value, dataType) {
        return ConfigValue{}, ErrInvalidValueForType
    }
    
    // Normalize value based on type
    normalized, err := normalizeValue(value, dataType)
    if err != nil {
        return ConfigValue{}, err
    }
    
    return ConfigValue{
        Raw:      normalized,
        DataType: dataType,
    }, nil
}

// Type-safe value extraction
func (cv ConfigValue) AsString() (string, error) {
    if cv.DataType != DataTypeString {
        return "", ErrWrongDataType
    }
    return cv.Raw.(string), nil
}

func (cv ConfigValue) AsInt() (int, error) {
    if cv.DataType != DataTypeInteger {
        return 0, ErrWrongDataType
    }
    return cv.Raw.(int), nil
}

func (cv ConfigValue) AsBool() (bool, error) {
    if cv.DataType != DataTypeBoolean {
        return false, ErrWrongDataType
    }
    return cv.Raw.(bool), nil
}

func (cv ConfigValue) AsJSON() (map[string]interface{}, error) {
    if cv.DataType != DataTypeJSON {
        return nil, ErrWrongDataType
    }
    return cv.Raw.(map[string]interface{}), nil
}

// Value object equality
func (cv ConfigValue) Equals(other ConfigValue) bool {
    if cv.DataType != other.DataType {
        return false
    }
    
    return reflect.DeepEqual(cv.Raw, other.Raw)
}
```

#### **Configuration Key Value Object**
```go
// @internal/core/settings/domain/config_key.go

// ConfigKey enforces business rules for configuration keys
type ConfigKey string

func NewConfigKey(module ModuleName, key string) (ConfigKey, error) {
    if module == "" {
        return "", ErrModuleRequired
    }
    if key == "" {
        return "", ErrKeyRequired
    }
    if !isValidKeyFormat(key) {
        return "", ErrInvalidKeyFormat
    }
    
    fullKey := fmt.Sprintf("%s.%s", module, key)
    return ConfigKey(fullKey), nil
}

func (ck ConfigKey) Module() ModuleName {
    parts := strings.SplitN(string(ck), ".", 2)
    return ModuleName(parts[0])
}

func (ck ConfigKey) Key() string {
    parts := strings.SplitN(string(ck), ".", 2)
    if len(parts) < 2 {
        return string(ck)
    }
    return parts[1]
}

func (ck ConfigKey) IsValid() bool {
    _, err := NewConfigKey(ck.Module(), ck.Key())
    return err == nil
}

// Business logic for key validation
func isValidKeyFormat(key string) bool {
    // Keys must be lowercase with underscores
    matched, _ := regexp.MatchString(`^[a-z0-9_]+$`, key)
    return matched && len(key) <= 100
}
```

### **Domain Events**

#### **Event Definitions**
```go
// @internal/core/settings/domain/events.go

// Configuration events capture important business occurrences
type DomainEvent interface {
    EventID() string
    EventType() string
    EventTime() time.Time
    AggregateID() string
    TenantID() tenant.ID
}

// Configuration change events
type ConfigurationUpdatedEvent struct {
    ID            string
    Type          string
    Time          time.Time
    ConfigID      ConfigurationID
    TenantIDVal   tenant.ID
    EntityID      *entity.ID
    Module        ModuleName
    ConfigKey     ConfigKey
    OldValue      ConfigValue
    NewValue      ConfigValue
    Source        ConfigSource
    UpdatedBy     identity.UserID
}

func (e ConfigurationUpdatedEvent) EventID() string     { return e.ID }
func (e ConfigurationUpdatedEvent) EventType() string   { return e.Type }
func (e ConfigurationUpdatedEvent) EventTime() time.Time { return e.Time }
func (e ConfigurationUpdatedEvent) AggregateID() string  { return string(e.ConfigID) }
func (e ConfigurationUpdatedEvent) TenantID() tenant.ID  { return e.TenantIDVal }

// Template application events
type TemplateAppliedEvent struct {
    ID              string
    Type            string
    Time            time.Time
    TemplateID      TemplateID
    TenantIDVal     tenant.ID
    TargetID        string
    TargetType      string // tenant, entity
    AppliedConfigs  int
    ConflictCount   int
    AppliedBy       identity.UserID
}

func (e TemplateAppliedEvent) EventID() string     { return e.ID }
func (e TemplateAppliedEvent) EventType() string   { return e.Type }
func (e TemplateAppliedEvent) EventTime() time.Time { return e.Time }
func (e TemplateAppliedEvent) AggregateID() string  { return string(e.TemplateID) }
func (e TemplateAppliedEvent) TenantID() tenant.ID  { return e.TenantIDVal }

// Bulk operation events
type BulkConfigurationUpdateEvent struct {
    ID           string
    Type         string
    Time         time.Time
    OperationID  string
    TenantIDVal  tenant.ID
    TargetCount  int
    SuccessCount int
    FailureCount int
    UpdatedBy    identity.UserID
}

func (e BulkConfigurationUpdateEvent) EventID() string     { return e.ID }
func (e BulkConfigurationUpdateEvent) EventType() string   { return e.Type }
func (e BulkConfigurationUpdateEvent) EventTime() time.Time { return e.Time }
func (e BulkConfigurationUpdateEvent) AggregateID() string  { return e.OperationID }
func (e BulkConfigurationUpdateEvent) TenantID() tenant.ID  { return e.TenantIDVal }
```

---

## Data Layer Architecture

### **Database Schema Design**

#### **Multi-Tenant Configuration Tables**
```sql
-- @db/migration/068_settings_core_tables.up.sql

-- Configuration definitions - system-wide metadata
CREATE TABLE config_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    module_name VARCHAR(50) NOT NULL,
    config_key VARCHAR(100) NOT NULL,
    data_type VARCHAR(20) NOT NULL CHECK (data_type IN ('string', 'integer', 'boolean', 'decimal', 'json')),
    default_value JSONB,
    validation_rules JSONB DEFAULT '{}',
    description TEXT,
    required_permission VARCHAR(100),
    required_feature_flag VARCHAR(100),
    is_overridable BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(module_name, config_key)
);

-- Tenant-level configurations (enhances existing tenant_configurations)
ALTER TABLE tenant_configurations ADD COLUMN IF NOT EXISTS settings_version INTEGER DEFAULT 1;
ALTER TABLE tenant_configurations ADD COLUMN IF NOT EXISTS last_template_applied UUID;
ALTER TABLE tenant_configurations ADD COLUMN IF NOT EXISTS template_applied_at TIMESTAMPTZ;

-- Entity-level configurations (enhances existing entities.settings)
-- This leverages the existing entities.settings JSONB column

-- Configuration templates
CREATE TABLE configuration_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    category VARCHAR(50) NOT NULL,
    description TEXT,
    version VARCHAR(20) NOT NULL,
    configurations JSONB NOT NULL,
    applicable_tenant_types TEXT[],
    required_feature_flags TEXT[],
    conflict_resolution VARCHAR(20) DEFAULT 'merge',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    UNIQUE(name, version)
);

-- -- Configuration audit trail
-- CREATE TABLE configuration_audit (
--     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
--     tenant_id UUID NOT NULL REFERENCES tenants(id),
--     entity_id UUID REFERENCES entities(uuid),
--     config_key VARCHAR(150) NOT NULL,
--     old_value JSONB,
--     new_value JSONB,
--     source VARCHAR(20) NOT NULL,
--     operation VARCHAR(20) NOT NULL,
--     user_id UUID NOT NULL,
--     applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
--     session_id VARCHAR(100),
--     correlation_id VARCHAR(100)
-- );
-- Note We will use audit_log table 

-- Enable RLS on new tables
ALTER TABLE config_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE configuration_templates ENABLE ROW LEVEL SECURITY;
ALTER TABLE configuration_audit ENABLE ROW LEVEL SECURITY;

-- Create tenant isolation policies for audit
CREATE POLICY configuration_audit_tenant_isolation ON configuration_audit
    FOR ALL TO authenticated
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

-- Templates are globally readable but only creatable by system admins
CREATE POLICY configuration_templates_read ON configuration_templates
    FOR SELECT TO authenticated
    USING (true);

CREATE POLICY configuration_templates_modify ON configuration_templates
    FOR ALL TO authenticated
    USING (current_setting('app.user_role', true) = 'system_admin');
```

#### **Configuration Resolution Views**
```sql
-- Efficient configuration resolution with inheritance
CREATE OR REPLACE VIEW configuration_resolution AS
WITH RECURSIVE config_inheritance AS (
    -- System defaults from definitions
    SELECT 
        cd.module_name,
        cd.config_key,
        cd.default_value as value,
        'system' as source,
        NULL::UUID as tenant_id,
        NULL::UUID as entity_id,
        0 as inheritance_level,
        cd.data_type,
        cd.is_overridable
    FROM config_definitions cd
    
    UNION ALL
    
    -- Tenant-level configurations
    SELECT 
        split_part(key, '.', 1) as module_name,
        split_part(key, '.', 2) as config_key,
        value,
        'tenant' as source,
        tc.tenant_id,
        NULL::UUID as entity_id,
        1 as inheritance_level,
        cd.data_type,
        cd.is_overridable
    FROM tenant_configurations tc,
         jsonb_each(tc.settings) as setting(key, value)
    JOIN config_definitions cd ON cd.module_name = split_part(key, '.', 1) 
                              AND cd.config_key = split_part(key, '.', 2)
    
    UNION ALL
    
    -- Entity-level configurations
    SELECT 
        split_part(key, '.', 1) as module_name,
        split_part(key, '.', 2) as config_key,
        value,
        'entity' as source,
        e.tenant_id,
        e.uuid as entity_id,
        2 as inheritance_level,
        cd.data_type,
        cd.is_overridable
    FROM entities e,
         jsonb_each(e.settings) as setting(key, value)
    JOIN config_definitions cd ON cd.module_name = split_part(key, '.', 1) 
                              AND cd.config_key = split_part(key, '.', 2)
    WHERE e.deleted_at IS NULL
)
SELECT DISTINCT ON (tenant_id, entity_id, module_name, config_key)
    module_name,
    config_key,
    value,
    source,
    tenant_id,
    entity_id,
    data_type,
    is_overridable
FROM config_inheritance
ORDER BY tenant_id, entity_id, module_name, config_key, inheritance_level DESC;

-- Performance optimization indexes
CREATE INDEX CONCURRENTLY idx_config_definitions_module_key 
ON config_definitions(module_name, config_key);

CREATE INDEX CONCURRENTLY idx_tenant_configurations_tenant_settings 
ON tenant_configurations(tenant_id) INCLUDE (settings);

CREATE INDEX CONCURRENTLY idx_entities_tenant_settings 
ON entities(tenant_id) INCLUDE (settings) 
WHERE deleted_at IS NULL;

CREATE INDEX CONCURRENTLY idx_configuration_audit_tenant_config 
ON configuration_audit(tenant_id, config_key, applied_at);

-- JSONB indexes for efficient configuration lookup
CREATE INDEX CONCURRENTLY idx_tenant_configurations_settings_gin 
ON tenant_configurations USING GIN (settings);

CREATE INDEX CONCURRENTLY idx_entities_settings_gin 
ON entities USING GIN (settings);
```

### **SQLC Integration**

#### **Type-Safe Configuration Queries**
```sql
-- @db/queries/settings.sql

-- name: GetConfigDefinition :one
SELECT * FROM config_definitions 
WHERE module_name = $1 AND config_key = $2;

-- name: ListConfigDefinitions :many
SELECT * FROM config_definitions 
WHERE ($1::TEXT = '' OR module_name = $1)
ORDER BY module_name, config_key;

-- name: ResolveConfiguration :one
SELECT 
    module_name,
    config_key,
    value,
    source,
    data_type
FROM configuration_resolution
WHERE module_name = $1 
AND config_key = $2
AND (tenant_id = $3 OR tenant_id IS NULL)
AND (entity_id = $4 OR entity_id IS NULL)
ORDER BY 
    CASE WHEN tenant_id = $3 AND entity_id = $4 THEN 3
         WHEN tenant_id = $3 AND entity_id IS NULL THEN 2
         WHEN tenant_id IS NULL AND entity_id IS NULL THEN 1
         ELSE 0 END DESC
LIMIT 1;

-- name: GetTenantConfigurations :one
SELECT settings FROM tenant_configurations 
WHERE tenant_id = $1;

-- name: UpdateTenantConfiguration :exec
UPDATE tenant_configurations 
SET 
    settings = jsonb_set(
        COALESCE(settings, '{}'),
        $2::text[],
        $3::jsonb
    ),
    updated_at = NOW(),
    settings_version = settings_version + 1
WHERE tenant_id = $1;

-- name: GetEntityConfigurations :one
SELECT settings FROM entities 
WHERE tenant_id = $1 AND uuid = $2 AND deleted_at IS NULL;

-- name: UpdateEntityConfiguration :exec
UPDATE entities 
SET 
    settings = jsonb_set(
        COALESCE(settings, '{}'),
        $3::text[],
        $4::jsonb
    ),
    updated_at = NOW()
WHERE tenant_id = $1 AND uuid = $2;

-- name: CreateConfigurationAudit :one
INSERT INTO configuration_audit (
    tenant_id, entity_id, config_key, old_value, new_value, 
    source, operation, user_id, session_id, correlation_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetConfigurationHistory :many
SELECT * FROM configuration_audit
WHERE tenant_id = $1 
AND ($2::UUID IS NULL OR entity_id = $2)
AND ($3::TEXT = '' OR config_key = $3)
ORDER BY applied_at DESC
LIMIT $4 OFFSET $5;

-- name: CreateConfigurationTemplate :one
INSERT INTO configuration_templates (
    name, category, description, version, configurations,
    applicable_tenant_types, required_feature_flags,
    conflict_resolution, created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetConfigurationTemplate :one
SELECT * FROM configuration_templates 
WHERE id = $1 AND is_active = true;

-- name: ListConfigurationTemplates :many
SELECT * FROM configuration_templates
WHERE is_active = true
AND ($1::TEXT = '' OR category = $1)
ORDER BY name;

-- name: BulkUpdateConfigurations :exec
UPDATE entities 
SET settings = settings || $2::jsonb,
    updated_at = NOW()
WHERE tenant_id = $1 
AND uuid = ANY($3::UUID[]);
```

#### **Repository Implementation**
```go
// @internal/core/settings/repository/configuration_repository.go

type configurationRepository struct {
    store   db.Store
    cache   cache.Service
    tracing tracing.TracingService
    metrics metrics.MetricsProvider
}

func (r *configurationRepository) ResolveConfiguration(ctx context.Context, req ResolutionRequest) (*ResolvedConfiguration, error) {
    ctx, span := r.tracing.StartSpan(ctx, "repository.resolve_configuration")
    defer span.End()
    
    timer := r.metrics.Timer("settings.resolve_configuration", nil)
    defer timer.Stop()
    
    // Check cache first
    cacheKey := fmt.Sprintf("config:resolved:%s:%s:%s:%s", 
        req.Module, req.ConfigKey, req.TenantID, req.EntityID)
    
    if cached, found := r.cache.Get(ctx, cacheKey); found {
        r.metrics.IncrementCounter("settings.cache_hit", nil)
        return cached.(*ResolvedConfiguration), nil
    }
    
    // Ensure tenant context for RLS
    ctx = tenant.WithContext(ctx, req.TenantID)
    
    var entityID *uuid.UUID
    if req.EntityID != nil {
        id := uuid.UUID(*req.EntityID)
        entityID = &id
    }
    
    result, err := r.store.ResolveConfiguration(ctx, db.ResolveConfigurationParams{
        ModuleName: string(req.Module),
        ConfigKey:  string(req.ConfigKey),
        TenantID:   uuid.UUID(req.TenantID),
        EntityID:   entityID,
    })
    if err != nil {
        if err == db.ErrNoRows {
            return nil, domain.ErrConfigurationNotFound
        }
        return nil, fmt.Errorf("failed to resolve configuration: %w", err)
    }
    
    resolved := r.mapToResolvedConfiguration(result)
    
    // Cache the result
    r.cache.Set(ctx, cacheKey, resolved, 5*time.Minute)
    r.metrics.IncrementCounter("settings.cache_miss", nil)
    
    return resolved, nil
}

func (r *configurationRepository) UpdateTenantConfiguration(ctx context.Context, tenantID tenant.ID, module ModuleName, key string, value ConfigValue) error {
    ctx, span := r.tracing.StartSpan(ctx, "repository.update_tenant_configuration")
    defer span.End()
    
    // Ensure tenant context for RLS
    ctx = tenant.WithContext(ctx, tenantID)
    
    jsonPath := []string{fmt.Sprintf("%s.%s", module, key)}
    valueJSON, err := json.Marshal(value.Raw)
    if err != nil {
        return fmt.Errorf("failed to marshal configuration value: %w", err)
    }
    
    err = r.store.UpdateTenantConfiguration(ctx, db.UpdateTenantConfigurationParams{
        TenantID: uuid.UUID(tenantID),
        Path:     jsonPath,
        Value:    json.RawMessage(valueJSON),
    })
    if err != nil {
        return fmt.Errorf("failed to update tenant configuration: %w", err)
    }
    
    // Invalidate related cache entries
    r.invalidateConfigurationCache(ctx, tenantID, nil, module, key)
    
    return nil
}

func (r *configurationRepository) BulkUpdateEntityConfigurations(ctx context.Context, updates []BulkConfigurationUpdate) error {
    ctx, span := r.tracing.StartSpan(ctx, "repository.bulk_update_entity_configurations")
    defer span.End()
    
    timer := r.metrics.Timer("settings.bulk_update", map[string]string{
        "count": fmt.Sprintf("%d", len(updates)),
    })
    defer timer.Stop()
    
    // Group updates by tenant for efficient processing
    updatesByTenant := r.groupUpdatesByTenant(updates)
    
    for tenantID, tenantUpdates := range updatesByTenant {
        err := r.processTenantBulkUpdates(ctx, tenantID, tenantUpdates)
        if err != nil {
            return err
        }
    }
    
    return nil
}

func (r *configurationRepository) invalidateConfigurationCache(ctx context.Context, tenantID tenant.ID, entityID *entity.ID, module ModuleName, key string) {
    // Invalidate specific configuration
    cacheKey := fmt.Sprintf("config:resolved:%s:%s:%s:%s", module, key, tenantID, entityID)
    r.cache.Delete(ctx, cacheKey)
    
    // Invalidate related patterns
    pattern := fmt.Sprintf("config:resolved:%s:%s:%s:*", module, key, tenantID)
    r.cache.DeletePattern(ctx, pattern)
}
```

---

This architectural foundation demonstrates enterprise-grade configuration management with proper separation of concerns, robust domain modeling,  caching strategies, and high-performance data access patterns. The remaining sections would continue with similar detail for API implementation, security architecture, performance optimization, and integration patterns.

**Document Status**: Core Architecture Complete
**Next Sections**: API Layer Implementation, Security Architecture, Performance Optimization
