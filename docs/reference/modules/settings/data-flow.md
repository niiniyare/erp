# Settings System Data Flow & Architecture

## Overview

This document details the data flow patterns, architectural decisions, and integration points for the Settings system within the ERP platform. It serves as a technical reference for understanding how configuration data moves through the system and how services interact with the Settings module.

## Table of Contents

1. [System Architecture](#system-architecture)
2. [Data Flow Patterns](#data-flow-patterns)
3. [Service Layer Architecture](#service-layer-architecture)
4. [Repository Layer Data Access](#repository-layer-data-access)
5. [Configuration Resolution Engine](#configuration-resolution-engine)
6. [Template System Architecture](#template-system-architecture)
7. [Audit and Compliance Flow](#audit-and-compliance-flow)
8. [Cache and Performance Strategy](#cache-and-performance-strategy)
9. [Security and Tenant Isolation](#security-and-tenant-isolation)

---

## System Architecture

### High-Level Architecture

```mermaid
graph TB
    subgraph "External Services"
        FS[Finance Service]
        HS[HR Service]
        IS[Inventory Service]
        OS[Other Services]
    end
    
    subgraph "Settings Module"
        subgraph "Service Layer"
            CS[Configuration Service]
            TS[Template Service]
        end
        
        subgraph "Repository Layer"  
            CR[Configuration Repository]
            TR[Template Repository]
        end
        
        subgraph "Domain Layer"
            CD[Configuration Domain]
            TD[Template Domain]
            VD[Validation Domain]
        end
    end
    
    subgraph "Infrastructure Layer"
        DB[(PostgreSQL Database)]
        CACHE[(Redis Cache)]
        AS[Audit Service]
        MS[Metrics Service]
        LS[Logging Service]
    end
    
    subgraph "Platform Services"
        IAM[IAM Service]
        FF[Feature Flag Service]
        TR_SVC[Tracing Service]
    end
    
    FS --> CS
    HS --> CS
    IS --> CS
    OS --> CS
    
    FS --> TS
    HS --> TS
    IS --> TS
    
    CS --> CR
    TS --> TR
    CR --> TR
    
    CR --> DB
    TR --> DB
    
    CS --> AS
    TS --> AS
    
    CS --> CACHE
    TS --> CACHE
    
    CS --> MS
    TS --> MS
    
    CS --> LS
    TS --> LS
    
    CR --> IAM
    TR --> IAM
    
    CR --> FF
    TR --> FF
    
    CS --> TR_SVC
    TS --> TR_SVC
```

### Component Responsibilities

#### Service Layer
- **Configuration Service**: Business logic for configuration CRUD operations, validation, and resolution
- **Template Service**: Template lifecycle management, application workflows, and conflict resolution

#### Repository Layer  
- **Configuration Repository**: Data access abstraction with SQLC integration and caching
- **Template Repository**: Template storage and retrieval with audit trail management

#### Domain Layer
- **Configuration Domain**: Value objects, entities, and business rules for configurations
- **Template Domain**: Template aggregates, application logic, and validation rules
- **Validation Domain**: Rich validation framework with extensible rule engine

---

## Data Flow Patterns

### Configuration Read Flow

```mermaid
sequenceDiagram
    participant Service as External Service
    participant CS as Configuration Service
    participant Cache as Cache Layer
    participant Repo as Configuration Repository
    participant DB as Database
    participant Logger as Logger
    participant Metrics as Metrics
    
    Service->>CS: GetEffectiveConfiguration(ctx, entityID, module, key)
    CS->>Metrics: Start timer
    CS->>Logger: Log configuration request
    CS->>Repo: GetEffectiveConfiguration(ctx, entityID, module, key)
    
    Repo->>Cache: Check cache
    alt Cache Hit
        Cache-->>Repo: Return cached config
        Repo-->>CS: Return configuration
    else Cache Miss
        Repo->>DB: Query with RLS (current_tenant_id())
        DB-->>Repo: Return raw data
        Repo->>Repo: Map to domain object
        Repo->>Cache: Store in cache
        Repo-->>CS: Return configuration
    end
    
    CS->>Logger: Log success/failure
    CS->>Metrics: Record latency
    CS-->>Service: Return configuration
```

### Configuration Write Flow

```mermaid
sequenceDiagram
    participant Service as External Service
    participant CS as Configuration Service
    participant Repo as Configuration Repository
    participant DB as Database
    participant Audit as Audit Service
    participant Cache as Cache Layer
    participant Logger as Logger
    
    Service->>CS: UpdateTenantConfiguration(ctx, module, key, value, userID)
    CS->>Logger: Log update start
    
    CS->>CS: ValidateConfigurationValue(ctx, module, key, value)
    alt Validation Fails
        CS->>Logger: Log validation error
        CS-->>Service: Return error
    end
    
    CS->>Repo: GetTenantConfiguration(ctx, module, key)
    Repo-->>CS: Current config (for audit)
    
    CS->>Repo: UpdateTenantConfiguration(ctx, module, key, value, userID)
    Repo->>DB: UPDATE with RLS protection
    DB-->>Repo: Confirm update
    Repo-->>CS: Success
    
    CS->>Audit: CreateAuditEvent(old_value, new_value, context)
    Audit-->>CS: Audit logged
    
    CS->>Cache: Invalidate related entries
    CS->>Logger: Log successful update
    CS-->>Service: Success response
```

### Template Application Flow

```mermaid
sequenceDiagram
    participant Service as External Service
    participant TS as Template Service
    participant Repo as Repository
    participant CS as Configuration Service
    participant Audit as Audit Service
    participant Logger as Logger
    
    Service->>TS: ApplyTemplate(ctx, templateID, target, options, userID)
    TS->>Logger: Log application start
    
    TS->>Repo: GetTemplate(ctx, templateID)
    Repo-->>TS: Template with configurations
    
    TS->>TS: ValidateTemplateApplication(ctx, templateID, target)
    alt Validation Fails
        TS->>Logger: Log validation failure
        TS-->>Service: Return validation error
    end
    
    loop For each configuration in template
        TS->>CS: GetEffectiveConfiguration(ctx, target.entityID, config.module, config.key)
        CS-->>TS: Current value (if exists)
        
        alt Conflict detected
            TS->>TS: Apply conflict resolution strategy
        end
        
        TS->>CS: UpdateEntityConfiguration(ctx, target.entityID, config.module, config.key, config.value, userID)
        CS-->>TS: Update result
    end
    
    TS->>Audit: CreateAuditEvent(template_application_context)
    TS->>Logger: Log application completion
    TS-->>Service: Application result with metrics
```

---

## Service Layer Architecture

### Configuration Service Architecture

```mermaid
graph TD
    subgraph "Configuration Service"
        subgraph "Public Interface"
            CRUD[CRUD Operations]
            RES[Resolution Methods]
            BULK[Bulk Operations]
            SEARCH[Search Methods]
        end
        
        subgraph "Business Logic"
            VAL[Validation Engine]
            INH[Inheritance Resolver]
            CACHE_MGR[Cache Manager]
            AUDIT_MGR[Audit Manager]
        end
        
        subgraph "Infrastructure"
            REPO[Repository Interface]
            LOGGER[Logger]
            METRICS[Metrics Provider]
            TRACING[Tracing Service]
        end
    end
    
    CRUD --> VAL
    RES --> INH
    BULK --> VAL
    BULK --> INH
    SEARCH --> CACHE_MGR
    
    VAL --> REPO
    INH --> REPO
    CACHE_MGR --> REPO
    
    VAL --> AUDIT_MGR
    INH --> AUDIT_MGR
    
    AUDIT_MGR --> LOGGER
    CACHE_MGR --> LOGGER
    
    VAL --> METRICS
    INH --> METRICS
    BULK --> METRICS
```

### Template Service Architecture

```mermaid
graph TD
    subgraph "Template Service"
        subgraph "Template Management"
            TEMP_CRUD[Template CRUD]
            TEMP_VAL[Template Validation]
            TEMP_VER[Version Management]
        end
        
        subgraph "Application Engine"
            APP_VAL[Application Validation]
            CONFLICT[Conflict Resolution]
            ROLLBACK[Rollback Manager]
        end
        
        subgraph "Analytics & Reporting"
            USAGE[Usage Analytics]
            HISTORY[Application History]
            STATS[Template Statistics]
        end
        
        subgraph "Infrastructure"
            REPO[Repository Interface]
            CONFIG_SVC[Configuration Service]
            LOGGER[Logger]
            METRICS[Metrics Provider]
        end
    end
    
    TEMP_CRUD --> TEMP_VAL
    TEMP_VAL --> REPO
    
    APP_VAL --> CONFLICT
    CONFLICT --> CONFIG_SVC
    CONFLICT --> ROLLBACK
    
    ROLLBACK --> CONFIG_SVC
    
    USAGE --> REPO
    HISTORY --> REPO
    STATS --> REPO
    
    TEMP_CRUD --> LOGGER
    APP_VAL --> LOGGER
    CONFLICT --> LOGGER
    
    TEMP_CRUD --> METRICS
    APP_VAL --> METRICS
    USAGE --> METRICS
```

---

## Repository Layer Data Access

### SQLC Integration Architecture

```mermaid
graph TD
    subgraph "Repository Layer"
        subgraph "Configuration Repository"
            IFACE[Repository Interface]
            IMPL[Repository Implementation]
            CACHE_LAYER[Cache Integration]
        end
        
        subgraph "SQLC Generated"
            QUERIES[Generated Queries]
            TYPES[Generated Types] 
            STORE[SQLC Store Interface]
        end
        
        subgraph "Database Layer"
            CONN[Connection Pool]
            RLS[Row Level Security]
            TENANT_CTX[Tenant Context]
        end
    end
    
    IMPL --> QUERIES
    IMPL --> TYPES
    IMPL --> STORE
    IMPL --> CACHE_LAYER
    
    QUERIES --> CONN
    STORE --> CONN
    
    CONN --> RLS
    RLS --> TENANT_CTX
    
    CACHE_LAYER --> REDIS[(Redis)]
```

### Data Access Patterns

#### Context-Based Tenant Resolution
```go
// All repository methods use context for tenant ID resolution
func (r *configurationRepository) GetEffectiveConfiguration(
    ctx context.Context, 
    entityID *uuid.UUID, 
    module domain.ModuleName, 
    key domain.ConfigKey,
) (*domain.Configuration, error) {
    // Tenant ID automatically resolved from context
    tenantID, _ := shared.GetTenantID(ctx)
    ctx = shared.WithTenantID(ctx, tenantID)
    
    // SQLC query uses current_tenant_id() function
    result, err := r.store.GetEffectiveConfiguration(ctx, db.GetEffectiveConfigurationParams{
        ModuleName: string(module),
        ConfigKey:  string(key),
        EntityID:   *entityID,
    })
    
    // Map to domain object
    return r.mapToConfiguration(result)
}
```

#### Type-Safe Query Parameters
```sql
-- SQLC query with named parameters
-- name: GetEffectiveConfiguration :one
SELECT 
    module_name,
    config_key,
    value,
    source,
    data_type
FROM get_effective_configuration(
    sqlc.arg(module_name)::text,
    sqlc.arg(config_key)::text, 
    sqlc.arg(entity_id)::uuid
)
WHERE tenant_id = current_tenant_id();
```

#### Cache Integration Strategy
```go
func (r *configurationRepository) getWithCache(
    ctx context.Context, 
    cacheKey string, 
    queryFunc func() (*domain.Configuration, error),
) (*domain.Configuration, error) {
    // Check cache first
    var cached domain.Configuration
    if err := r.cache.Get(ctx, cacheKey, &cached); err == nil {
        r.metrics.IncrementCounter("settings.cache_hit", nil)
        return &cached, nil
    }
    
    // Cache miss - query database
    result, err := queryFunc()
    if err != nil {
        return nil, err
    }
    
    // Store in cache with TTL
    r.cache.Set(ctx, cacheKey, result, 5*time.Minute)
    r.metrics.IncrementCounter("settings.cache_miss", nil)
    
    return result, nil
}
```

---

## Configuration Resolution Engine

### Inheritance Resolution Algorithm

```mermaid
graph TD
    A[Resolution Request] --> B[Extract Context]
    B --> C[Parse Module & Key]
    C --> D[Build Cache Key]
    D --> E{Cache Hit?}
    
    E -->|Yes| F[Return Cached Value]
    E -->|No| G[Query Entity Level]
    
    G --> H{Found at Entity?}
    H -->|Yes| I[Map to Domain Object]
    H -->|No| J[Query Tenant Level]
    
    J --> K{Found at Tenant?}
    K -->|Yes| I
    K -->|No| L[Query System Default]
    
    L --> M{Found at System?}
    M -->|Yes| I
    M -->|No| N[Return Not Found Error]
    
    I --> O[Set Cache with TTL]
    O --> P[Record Metrics]
    P --> Q[Log Access]
    Q --> R[Return Configuration]
    
    F --> P
    N --> S[Log Not Found]
    S --> T[Return Error]
```

### Resolution Priority Matrix

| Level | Storage Location | Query Method | Cache TTL | Override Behavior |
|-------|-----------------|--------------|-----------|------------------|
| Entity | `entities.settings` | Direct JSONB query | 5 minutes | Overrides tenant/system |
| Tenant | `tenant_configurations.settings` | JSONB + dedicated columns | 15 minutes | Overrides system only |
| System | `config_definitions.default_value` | Static definition | 1 hour | Fallback only |

### Data Type Resolution

```go
type ConfigurationResolver struct {
    repo    repository.ConfigurationRepository
    cache   cache.Service
    logger  logger.Logger
    metrics metrics.MetricsProvider
}

func (r *ConfigurationResolver) ResolveWithType(
    ctx context.Context,
    entityID *uuid.UUID,
    module domain.ModuleName,
    key domain.ConfigKey,
) (*ResolvedConfiguration, error) {
    // Get configuration definition for type information
    definition, err := r.repo.GetConfigDefinition(ctx, module, key)
    if err != nil && err != domain.ErrConfigDefinitionNotFound {
        return nil, fmt.Errorf("failed to get configuration definition: %w", err)
    }
    
    // Resolve value through inheritance
    config, err := r.repo.GetEffectiveConfiguration(ctx, entityID, module, key)
    if err != nil {
        if err == domain.ErrConfigurationNotFound && definition != nil {
            // Use system default
            return &ResolvedConfiguration{
                Value:      definition.DefaultValue,
                Source:     domain.ConfigSourceSystem,
                DataType:   definition.DataType,
                Definition: definition,
            }, nil
        }
        return nil, err
    }
    
    // Validate data type consistency
    if definition != nil && config.Value.DataType != definition.DataType {
        r.logger.WarnContext(ctx, "Configuration data type mismatch", logger.Fields{
            "module":           string(module),
            "key":              string(key),
            "expected_type":    string(definition.DataType),
            "actual_type":      string(config.Value.DataType),
            "source":           string(config.Source),
        })
    }
    
    return &ResolvedConfiguration{
        Value:      config.Value,
        Source:     config.Source,
        DataType:   config.Value.DataType,
        Definition: definition,
        Metadata: ResolutionMetadata{
            ResolvedAt: time.Now(),
            CacheHit:   false, // Set by cache layer
            TenantID:   shared.GetTenantID(ctx),
            EntityID:   entityID,
        },
    }, nil
}
```

---

## Template System Architecture

### Template Domain Model

```mermaid
classDiagram
    class Template {
        +ID: TemplateID
        +Name: string
        +Category: TemplateCategory
        +Configurations: []TemplateConfiguration
        +Dependencies: []TemplateDependency
        +IsActive: bool
        +CreatedBy: UUID
        +ApplyTo(target) ApplicationResult
        +Validate() error
    }
    
    class TemplateConfiguration {
        +Module: ModuleName
        +ConfigKey: ConfigKey
        +Value: ConfigValue
        +OverridePolicy: OverridePolicy
        +Priority: int
        +Validate() error
    }
    
    class ApplicationResult {
        +TemplateID: TemplateID
        +TargetID: string
        +Applied: []ConfigurationChange
        +Conflicts: []ConfigurationConflict
        +Summary: ApplicationSummary
    }
    
    class ConfigurationChange {
        +Module: ModuleName
        +ConfigKey: ConfigKey
        +OldValue: *ConfigValue
        +NewValue: ConfigValue
        +Action: ChangeAction
    }
    
    Template --> TemplateConfiguration
    Template --> ApplicationResult
    ApplicationResult --> ConfigurationChange
```

### Template Application Engine

```mermaid
graph TD
    subgraph "Application Engine"
        A[Application Request] --> B[Load Template]
        B --> C[Validate Compatibility]
        C --> D[Sort by Priority]
        D --> E[Start Application]
        
        E --> F{For Each Config}
        F --> G[Get Current Value]
        G --> H{Conflict?}
        
        H -->|Yes| I[Apply Resolution Strategy]
        H -->|No| J[Apply Configuration]
        
        I --> K{Strategy Result}
        K -->|Preserve| L[Skip Configuration]
        K -->|Replace| J
        K -->|Merge| M[Merge Values]
        
        J --> N[Record Change]
        L --> O[Record Skip]
        M --> N
        
        N --> P{More Configs?}
        O --> P
        P -->|Yes| F
        P -->|No| Q[Create Audit Events]
        
        Q --> R[Return Results]
    end
```

### Conflict Resolution Strategies

```go
type ConflictResolver struct {
    logger logger.Logger
}

func (cr *ConflictResolver) ResolveConflict(
    templateConfig TemplateConfiguration,
    existingConfig *domain.Configuration,
    strategy ConflictStrategy,
) (*ConfigurationChange, *ConfigurationConflict, error) {
    
    switch strategy {
    case ConflictStrategyPreserve:
        if existingConfig != nil {
            return nil, &ConfigurationConflict{
                Module:        templateConfig.Module,
                ConfigKey:     templateConfig.ConfigKey,
                TemplateValue: templateConfig.Value,
                ExistingValue: existingConfig.Value,
                Resolution:    ConflictResolutionKeptExisting,
                Reason:        "preserve existing value strategy",
            }, nil
        }
        
    case ConflictStrategyReplace:
        return &ConfigurationChange{
            Module:    templateConfig.Module,
            ConfigKey: templateConfig.ConfigKey,
            OldValue:  existingConfig?.Value,
            NewValue:  templateConfig.Value,
            Action:    ChangeActionUpdate,
        }, nil, nil
        
    case ConflictStrategyMerge:
        if existingConfig != nil && 
           templateConfig.Value.DataType == domain.DataTypeJSON &&
           existingConfig.Value.DataType == domain.DataTypeJSON {
            
            mergedValue, err := cr.mergeJSONValues(existingConfig.Value, templateConfig.Value)
            if err != nil {
                return nil, nil, fmt.Errorf("failed to merge JSON values: %w", err)
            }
            
            return &ConfigurationChange{
                Module:    templateConfig.Module,
                ConfigKey: templateConfig.ConfigKey,
                OldValue:  &existingConfig.Value,
                NewValue:  mergedValue,
                Action:    ChangeActionMerge,
            }, nil, nil
        }
    }
    
    // Default to replace if no specific handling
    return &ConfigurationChange{
        Module:    templateConfig.Module,
        ConfigKey: templateConfig.ConfigKey,
        OldValue:  existingConfig?.Value,
        NewValue:  templateConfig.Value,
        Action:    ChangeActionCreate,
    }, nil, nil
}
```

---

## Audit and Compliance Flow

### Audit Event Creation

```mermaid
sequenceDiagram
    participant Service as Settings Service
    participant Audit as Audit Service
    participant Context as Audit Context Builder
    participant DB as Audit Database
    participant Logger as Logger
    
    Service->>Context: buildAuditContext(operation, old_value, new_value)
    Context-->>Service: JSON context with metadata
    
    Service->>Audit: CreateAuditEvent(request)
    Note over Audit: Request includes: UserID, EventType, Category, Severity, Context
    
    Audit->>DB: Store audit event with tenant isolation
    DB-->>Audit: Confirm storage
    
    alt Audit Success
        Audit-->>Service: Success response
        Service->>Logger: Log audit success
    else Audit Failure
        Audit-->>Service: Error response
        Service->>Logger: Log audit failure (non-blocking)
    end
```

### Audit Context Structure

```go
type ConfigurationAuditContext struct {
    Level       string                 `json:"level"`        // tenant, entity
    Module      string                 `json:"module"`       // finance, hr, inventory
    ConfigKey   string                 `json:"config_key"`   // approval_limit, etc.
    OldValue    interface{}            `json:"old_value,omitempty"`
    NewValue    interface{}            `json:"new_value,omitempty"`
    OldSource   string                 `json:"old_source,omitempty"`
    NewType     string                 `json:"new_type,omitempty"`
    Timestamp   time.Time              `json:"timestamp"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type TemplateAuditContext struct {
    Operation          string    `json:"operation"`           // create, update, apply
    TemplateID         string    `json:"template_id"`
    TemplateName       string    `json:"template_name"`
    TemplateCategory   string    `json:"template_category"`
    ConfigurationCount int       `json:"configuration_count"`
    IsActive           bool      `json:"is_active"`
    TargetType         string    `json:"target_type,omitempty"`
    TargetEntityID     *string   `json:"target_entity_id,omitempty"`
    AppliedConfigs     int       `json:"applied_configs,omitempty"`
    Conflicts          int       `json:"conflicts,omitempty"`
    Errors             int       `json:"errors,omitempty"`
    DurationMS         int64     `json:"duration_ms,omitempty"`
    Timestamp          time.Time `json:"timestamp"`
}
```

### Compliance Features

#### Change Tracking
- Every configuration change creates an audit event
- Old and new values are preserved in audit context  
- User attribution and timestamp tracking
- Source information (system/tenant/entity) preservation

#### Data Retention
- Audit events stored with configurable retention policies
- Compliance flags support for regulatory requirements
- Export capabilities for audit reporting
- Integration with external compliance systems

---

## Cache and Performance Strategy

### Caching Architecture

```mermaid
graph TD
    subgraph "Cache Strategy"
        subgraph "Cache Layers"
            L1[Service-Level Cache]
            L2[Repository Cache] 
            L3[Redis Distributed Cache]
        end
        
        subgraph "Cache Keys"
            CK1[config:resolved:module:key:tenant:entity]
            CK2[config:definition:module:key]
            CK3[template:id]
        end
        
        subgraph "TTL Strategy"
            TTL1[Entity: 5min]
            TTL2[Tenant: 15min]
            TTL3[System: 1hour]
            TTL4[Templates: 30min]
        end
    end
    
    L1 --> L2
    L2 --> L3
    
    CK1 --> TTL1
    CK2 --> TTL3
    CK3 --> TTL4
```

### Performance Optimization Patterns

#### Batch Operations
```go
func (s *configurationService) GetMultipleConfigurations(
    ctx context.Context,
    entityID *uuid.UUID,
    requests []ConfigurationRequest,
) (map[string]*domain.Configuration, error) {
    
    results := make(map[string]*domain.Configuration)
    cacheKeys := make([]string, len(requests))
    missedRequests := make([]ConfigurationRequest, 0)
    
    // Build cache keys and check cache
    for i, req := range requests {
        cacheKey := s.buildCacheKey(req.Module, req.Key, entityID)
        cacheKeys[i] = cacheKey
        
        var cached domain.Configuration
        if err := s.cache.Get(ctx, cacheKey, &cached); err == nil {
            results[req.String()] = &cached
        } else {
            missedRequests = append(missedRequests, req)
        }
    }
    
    // Batch query for cache misses
    if len(missedRequests) > 0 {
        dbResults, err := s.repo.GetMultipleEffectiveConfigurations(ctx, entityID, missedRequests)
        if err != nil {
            return nil, err
        }
        
        // Update cache and results
        for key, config := range dbResults {
            results[key] = config
            cacheKey := s.buildCacheKeyFromString(key, entityID)
            s.cache.Set(ctx, cacheKey, config, s.getCacheTTL(config.Source))
        }
    }
    
    return results, nil
}
```

#### Smart Cache Invalidation
```go
func (s *configurationService) invalidateRelatedCaches(
    ctx context.Context,
    tenantID uuid.UUID,
    entityID *uuid.UUID,
    module domain.ModuleName,
    key domain.ConfigKey,
) {
    patterns := []string{
        // Invalidate specific configuration
        s.buildCacheKey(module, key, tenantID, entityID),
        // Invalidate entity inheritance chain
        fmt.Sprintf("config:resolved:%s:%s:%s:*", module, key, tenantID),
        // Invalidate module-level caches
        fmt.Sprintf("config:list:%s:%s:*", module, tenantID),
    }
    
    for _, pattern := range patterns {
        if err := s.cache.DeletePattern(ctx, pattern); err != nil {
            s.logger.WarnContext(ctx, "Failed to invalidate cache pattern", logger.Fields{
                "pattern": pattern,
                "error":   err.Error(),
            })
        }
    }
}
```

### Performance Metrics

#### Key Metrics Tracked
- Configuration resolution latency (P50, P95, P99)
- Cache hit ratio by level (entity, tenant, system)
- Template application duration and success rate
- Database query performance and connection pool usage
- Bulk operation throughput and error rates

#### Monitoring Integration
```go
func (s *configurationService) recordPerformanceMetrics(
    ctx context.Context,
    operation string,
    startTime time.Time,
    err error,
    additionalFields metrics.Fields,
) {
    duration := time.Since(startTime)
    
    fields := metrics.Fields{
        "operation": operation,
        "duration_ms": duration.Milliseconds(),
    }
    
    // Merge additional fields
    for k, v := range additionalFields {
        fields[k] = v
    }
    
    if err != nil {
        fields["error"] = err.Error()
        s.metrics.IncrementCounter(fmt.Sprintf("settings.%s.error", operation), fields)
    } else {
        s.metrics.IncrementCounter(fmt.Sprintf("settings.%s.success", operation), fields)
    }
    
    s.metrics.RecordHistogram(fmt.Sprintf("settings.%s.duration", operation), float64(duration.Milliseconds()), fields)
}
```

---

## Security and Tenant Isolation

### Multi-Tenant Security Model

```mermaid
graph TD
    subgraph "Request Flow"
        A[Client Request] --> B[Authentication]
        B --> C[Authorization]
        C --> D[Tenant Context Extraction]
        D --> E[Service Call]
    end
    
    subgraph "Tenant Isolation"
        E --> F[Repository Layer]
        F --> G[Context Injection]
        G --> H[RLS Protection]
        H --> I[Database Query]
    end
    
    subgraph "Database Security"
        I --> J[current_tenant_id()]
        J --> K[Row Filter]
        K --> L[Result Set]
    end
    
    L --> M[Domain Mapping]
    M --> N[Response]
```

### Row-Level Security Implementation

#### Database-Level Protection
```sql
-- Configuration tables have RLS policies
CREATE POLICY tenant_configurations_isolation ON tenant_configurations
FOR ALL TO application_role
USING (tenant_id = current_tenant_id());

CREATE POLICY entity_configurations_isolation ON entities
FOR ALL TO application_role  
USING (tenant_id = current_tenant_id());

-- Function to get tenant ID from context
CREATE OR REPLACE FUNCTION current_tenant_id() RETURNS UUID AS $$
BEGIN
    RETURN current_setting('app.current_tenant_id')::UUID;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

#### Application-Level Security
```go
func (r *configurationRepository) ensureTenantContext(ctx context.Context) (context.Context, error) {
    tenantID, err := shared.GetTenantID(ctx)
    if err != nil {
        return nil, fmt.Errorf("tenant ID not found in context: %w", err)
    }
    
    // Set database session variable for RLS
    if err := r.store.SetTenantContext(ctx, tenantID); err != nil {
        return nil, fmt.Errorf("failed to set tenant context: %w", err)
    }
    
    return shared.WithTenantID(ctx, tenantID), nil
}
```

### Permission-Based Access Control

#### Configuration-Level Permissions
```go
type ConfigurationPermissions struct {
    CanRead   bool
    CanWrite  bool
    CanDelete bool
    Scope     PermissionScope // tenant, entity, specific_entity
}

func (s *configurationService) checkPermissions(
    ctx context.Context,
    operation string,
    module domain.ModuleName,
    key domain.ConfigKey,
    entityID *uuid.UUID,
) error {
    // Get current user from context
    userID, err := shared.GetUserID(ctx)
    if err != nil {
        return fmt.Errorf("user not authenticated: %w", err)
    }
    
    // Check IAM permissions
    permission := fmt.Sprintf("settings:%s:%s", module, operation)
    
    hasPermission, err := s.iamService.CheckPermission(ctx, userID, permission, map[string]interface{}{
        "module":    string(module),
        "key":       string(key),
        "entity_id": entityID,
    })
    
    if err != nil {
        return fmt.Errorf("permission check failed: %w", err)
    }
    
    if !hasPermission {
        return fmt.Errorf("insufficient permissions for %s on %s.%s", operation, module, key)
    }
    
    return nil
}
```

#### Feature Flag Integration
```go
func (s *configurationService) checkFeatureAvailability(
    ctx context.Context,
    module domain.ModuleName,
    key domain.ConfigKey,
) error {
    tenantID, _ := shared.GetTenantID(ctx)
    
    // Check if module/feature is enabled for tenant
    featureFlag := fmt.Sprintf("%s.%s", module, key)
    
    enabled, err := s.featureFlagService.IsEnabled(ctx, tenantID, featureFlag)
    if err != nil {
        s.logger.WarnContext(ctx, "Feature flag check failed, allowing access", logger.Fields{
            "feature_flag": featureFlag,
            "error":        err.Error(),
        })
        return nil // Fail open for availability
    }
    
    if !enabled {
        return fmt.Errorf("feature %s is not enabled for this tenant", featureFlag)
    }
    
    return nil
}
```

---

This  data flow and architecture document provides the technical foundation for understanding how the Settings system integrates with the broader ERP platform, ensuring security, performance, and maintainability.