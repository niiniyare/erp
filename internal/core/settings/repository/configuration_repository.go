package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/settings/domain"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// ConfigurationRepository provides data access for configuration management
type ConfigurationRepository interface {
	// Configuration Resolution
	ResolveConfiguration(ctx context.Context, req ResolutionRequest) (*domain.Configuration, error)
	GetEffectiveConfiguration(ctx context.Context, entityID *uuid.UUID, module domain.ModuleName, key domain.ConfigKey) (*domain.Configuration, error)
	ListEffectiveConfigurations(ctx context.Context, entityID *uuid.UUID, module domain.ModuleName) ([]*domain.Configuration, error)

	// Configuration CRUD
	GetTenantConfiguration(ctx context.Context, module domain.ModuleName, key domain.ConfigKey) (*domain.Configuration, error)
	UpdateTenantConfiguration(ctx context.Context, module domain.ModuleName, key domain.ConfigKey, value domain.ConfigValue, userID uuid.UUID) error
	DeleteTenantConfiguration(ctx context.Context, module domain.ModuleName, key domain.ConfigKey, userID uuid.UUID) error

	GetEntityConfiguration(ctx context.Context, entityID uuid.UUID, module domain.ModuleName, key domain.ConfigKey) (*domain.Configuration, error)
	UpdateEntityConfiguration(ctx context.Context, entityID uuid.UUID, module domain.ModuleName, key domain.ConfigKey, value domain.ConfigValue, userID uuid.UUID) error
	DeleteEntityConfiguration(ctx context.Context, entityID uuid.UUID, module domain.ModuleName, key domain.ConfigKey, userID uuid.UUID) error

	// Configuration Definitions
	GetConfigDefinition(ctx context.Context, module domain.ModuleName, key domain.ConfigKey) (*ConfigDefinition, error)
	ListConfigDefinitions(ctx context.Context, module domain.ModuleName) ([]*ConfigDefinition, error)
	CreateConfigDefinition(ctx context.Context, def *ConfigDefinition) error
	UpdateConfigDefinition(ctx context.Context, def *ConfigDefinition) error

	// Template Management
	CreateTemplate(ctx context.Context, template *domain.Template) error
	GetTemplate(ctx context.Context, templateID domain.TemplateID) (*domain.Template, error)
	ListTemplates(ctx context.Context, filters TemplateFilters) ([]*domain.Template, error)
	UpdateTemplate(ctx context.Context, template *domain.Template) error
	DeactivateTemplate(ctx context.Context, templateID domain.TemplateID) error

	// Template Application
	ApplyTemplate(ctx context.Context, templateID domain.TemplateID, target TemplateTarget, options ApplyOptions, userID uuid.UUID) (*domain.ApplicationResult, error)
	GetTemplateApplicationHistory(ctx context.Context, entityID *uuid.UUID, limit, offset int) ([]*TemplateApplication, error)

	// Audit Trail
	CreateAuditRecord(ctx context.Context, record *AuditRecord) error
	GetConfigurationHistory(ctx context.Context, entityID *uuid.UUID, configKey string, limit, offset int) ([]*AuditRecord, error)

	// Bulk Operations
	BulkUpdateConfigurations(ctx context.Context, updates []BulkConfigurationUpdate) (*BulkOperationResult, error)

	// Search and Analytics
	SearchConfigurations(ctx context.Context, criteria SearchCriteria) (*SearchResult, error)
}

// configurationRepository implements ConfigurationRepository
type configurationRepository struct {
	store   db.Store
	cache   cache.Service
	tracing tracing.Service
	metrics metrics.MetricsProvider
}

// NewConfigurationRepository creates a new configuration repository
func NewConfigurationRepository(
	store db.Store,
	cache cache.Service,
	tracing tracing.Service,
	metrics metrics.MetricsProvider,
) ConfigurationRepository {
	return &configurationRepository{
		store:   store,
		cache:   cache,
		tracing: tracing,
		metrics: metrics,
	}
}

// ResolutionRequest represents a configuration resolution request
type ResolutionRequest struct {
	EntityID        *uuid.UUID
	Module          domain.ModuleName
	ConfigKey       domain.ConfigKey
	IncludeMetadata bool
}

// ResolveConfiguration resolves a configuration value through inheritance hierarchy
func (r *configurationRepository) ResolveConfiguration(ctx context.Context, req ResolutionRequest) (*domain.Configuration, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.resolve_configuration")
	defer span.End()

	timer := r.metrics.Timer("settings.resolve_configuration", metrics.Fields{
		"module": string(req.Module),
	})
	defer timer.Stop()
	tenantId, _ := shared.GetTenantID(ctx)

	// Check cache first
	cacheKey := r.buildCacheKey("config:resolved", string(req.Module), string(req.ConfigKey), tenantId.String(), getEntityIDString(req.EntityID))

	var cachedConfig domain.Configuration
	if err := r.cache.Get(ctx, cacheKey, &cachedConfig); err == nil {
		r.metrics.IncrementCounter("settings.cache_hit", nil)
		return &cachedConfig, nil
	}
	// Ensure tenant context for RLS
	ctx = shared.WithTenantID(ctx, tenantId)

	var entityID uuid.UUID
	if req.EntityID != nil {
		entityID = *req.EntityID
	}

	result, err := r.store.GetEffectiveConfiguration(ctx, db.GetEffectiveConfigurationParams{
		ModuleName: string(req.Module),
		ConfigKey:  string(req.ConfigKey),
		EntityID:   entityID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrConfigurationNotFound
		}
		return nil, fmt.Errorf("failed to resolve configuration: %w", err)
	}

	config, err := r.mapToConfiguration(result.ModuleName, result.ConfigKey, result.Value, result.Source, result.ConfigType, tenantId, req.EntityID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	r.cache.Set(ctx, cacheKey, config, 5*time.Minute)
	r.metrics.IncrementCounter("settings.cache_miss", nil)

	return config, nil
}

// GetEffectiveConfiguration gets the effective configuration value for a specific context
func (r *configurationRepository) GetEffectiveConfiguration(ctx context.Context, entityID *uuid.UUID, module domain.ModuleName, key domain.ConfigKey) (*domain.Configuration, error) {
	// Get tenant ID from context
	req := ResolutionRequest{
		EntityID:  entityID,
		Module:    module,
		ConfigKey: key,
	}
	return r.ResolveConfiguration(ctx, req)
}

// ListEffectiveConfigurations lists all effective configurations for a tenant/entity
func (r *configurationRepository) ListEffectiveConfigurations(ctx context.Context, entityID *uuid.UUID, module domain.ModuleName) ([]*domain.Configuration, error) {
	// Get tenant ID from context
	tenantID, _ := shared.GetTenantID(ctx)
	ctx, span := r.tracing.StartSpan(ctx, "repository.list_effective_configurations")
	defer span.End()

	// Ensure tenant context for RLS
	ctx = shared.WithTenantID(ctx, tenantID)

	moduleFilter := ""
	if module != "" {
		moduleFilter = string(module)
	}

	var uuid uuid.UUID
	if entityID != nil {
		uuid = *entityID
	}

	results, err := r.store.ListTenantEffectiveConfigurations(ctx, db.ListTenantEffectiveConfigurationsParams{
		EntityID:     uuid,
		ModuleFilter: &moduleFilter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list effective configurations: %w", err)
	}

	configs := make([]*domain.Configuration, 0, len(results))
	for _, result := range results {
		valueBytes, ok := result.Value.([]byte)
		if !ok {
			continue // Skip values that can't be converted to bytes
		}

		config, err := r.mapToConfiguration(result.ModuleName, result.ConfigKey, json.RawMessage(valueBytes), result.Source, "", tenantID, entityID)
		if err != nil {
			continue // Skip invalid configurations
		}

		configs = append(configs, config)
	}

	return configs, nil
}

// UpdateTenantConfiguration updates a tenant-level configuration
func (r *configurationRepository) UpdateTenantConfiguration(ctx context.Context, module domain.ModuleName, key domain.ConfigKey, value domain.ConfigValue, userID uuid.UUID) error {
	// Get tenant ID from context
	tenantID, _ := shared.GetTenantID(ctx)
	ctx, span := r.tracing.StartSpan(ctx, "repository.update_tenant_configuration")
	defer span.End()

	// Ensure tenant context for RLS
	ctx = shared.WithTenantID(ctx, tenantID)

	// Get current value for audit
	currentConfig, _ := r.GetTenantConfiguration(ctx, module, key)

	jsonPath := []string{fmt.Sprintf("%s.%s", module, key)}
	valueJSON, err := json.Marshal(value.Raw)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration value: %w", err)
	}

	err = r.store.UpdateTenantConfigurationSettings(ctx, db.UpdateTenantConfigurationSettingsParams{
		ConfigPath:  jsonPath,
		ConfigValue: valueJSON,
	})
	if err != nil {
		return fmt.Errorf("failed to update tenant configuration: %w", err)
	}

	// Create audit record
	auditRecord := &AuditRecord{
		TenantID:  tenantID,
		ConfigKey: fmt.Sprintf("%s.%s", module, key),
		NewValue:  value,
		Source:    domain.ConfigSourceTenant,
		Operation: "update",
		UserID:    userID,
		AppliedAt: time.Now(),
	}

	if currentConfig != nil {
		auditRecord.OldValue = &currentConfig.Value
	}

	if err := r.CreateAuditRecord(ctx, auditRecord); err != nil {
		// Log error but don't fail the operation
		span.RecordError(err)
	}

	// Invalidate related cache entries
	r.invalidateConfigurationCache(ctx, tenantID, nil, module, key)

	return nil
}

// UpdateEntityConfiguration updates an entity-level configuration
func (r *configurationRepository) UpdateEntityConfiguration(ctx context.Context, entityID uuid.UUID, module domain.ModuleName, key domain.ConfigKey, value domain.ConfigValue, userID uuid.UUID) error {
	// Get tenant ID from context
	tenantID, _ := shared.GetTenantID(ctx)
	ctx, span := r.tracing.StartSpan(ctx, "repository.update_entity_configuration")
	defer span.End()

	// Ensure tenant context for RLS
	ctx = shared.WithTenantID(ctx, tenantID)

	// Get current value for audit
	currentConfig, _ := r.GetEntityConfiguration(ctx, entityID, module, key)

	jsonPath := []string{fmt.Sprintf("%s.%s", module, key)}
	valueJSON, err := json.Marshal(value.Raw)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration value: %w", err)
	}

	err = r.store.UpdateEntityConfiguration(ctx, db.UpdateEntityConfigurationParams{
		ConfigPath:  jsonPath,
		ConfigValue: valueJSON,
		EntityID:    entityID,
	})
	if err != nil {
		return fmt.Errorf("failed to update entity configuration: %w", err)
	}

	// Create audit record
	auditRecord := &AuditRecord{
		TenantID:  tenantID,
		EntityID:  &entityID,
		ConfigKey: fmt.Sprintf("%s.%s", module, key),
		NewValue:  value,
		Source:    domain.ConfigSourceEntity,
		Operation: "update",
		UserID:    userID,
		AppliedAt: time.Now(),
	}

	if currentConfig != nil {
		auditRecord.OldValue = &currentConfig.Value
	}

	if err := r.CreateAuditRecord(ctx, auditRecord); err != nil {
		// Log error but don't fail the operation
		span.RecordError(err)
	}

	// Invalidate related cache entries
	r.invalidateConfigurationCache(ctx, tenantID, &entityID, module, key)

	return nil
}

// GetTenantConfiguration gets a tenant-level configuration
func (r *configurationRepository) GetTenantConfiguration(ctx context.Context, module domain.ModuleName, key domain.ConfigKey) (*domain.Configuration, error) {
	// Get tenant ID from context
	tenantID, _ := shared.GetTenantID(ctx)
	ctx = shared.WithTenantID(ctx, tenantID)

	result, err := r.store.GetTenantConfigurations(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrConfigurationNotFound
		}
		return nil, fmt.Errorf("failed to get tenant configurations: %w", err)
	}

	configKey := fmt.Sprintf("%s.%s", module, key)
	if result.Settings == nil {
		return nil, domain.ErrConfigurationNotFound
	}

	var settings map[string]any
	if err := json.Unmarshal(result.Settings, &settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	rawValue, exists := settings[configKey]
	if !exists {
		return nil, domain.ErrConfigurationNotFound
	}

	// This is simplified - in reality we'd need to determine the data type
	value := domain.NewStringValue(fmt.Sprintf("%v", rawValue))

	config, err := domain.NewConfiguration(
		nil,
		module,
		key,
		value,
		domain.ConfigSourceTenant,
	)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// GetEntityConfiguration gets an entity-level configuration
func (r *configurationRepository) GetEntityConfiguration(ctx context.Context, entityID uuid.UUID, module domain.ModuleName, key domain.ConfigKey) (*domain.Configuration, error) {
	// Get tenant ID from context
	tenantID, _ := shared.GetTenantID(ctx)
	ctx = shared.WithTenantID(ctx, tenantID)

	result, err := r.store.GetEntityConfigurations(ctx, entityID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrConfigurationNotFound
		}
		return nil, fmt.Errorf("failed to get entity configurations: %w", err)
	}

	configKey := fmt.Sprintf("%s.%s", module, key)
	if result == nil {
		return nil, domain.ErrConfigurationNotFound
	}

	var settings map[string]any
	if err := json.Unmarshal(result, &settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	rawValue, exists := settings[configKey]
	if !exists {
		return nil, domain.ErrConfigurationNotFound
	}

	// This is simplified - in reality we'd need to determine the data type
	value := domain.NewStringValue(fmt.Sprintf("%v", rawValue))

	config, err := domain.NewConfiguration(
		&entityID,
		module,
		key,
		value,
		domain.ConfigSourceEntity,
	)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// Helper methods

// buildCacheKey constructs a cache key from components
func (r *configurationRepository) buildCacheKey(prefix string, components ...string) string {
	parts := append([]string{prefix}, components...)
	return strings.Join(parts, ":")
}

// getEntityIDString returns entity ID as string or "nil" if nil
func getEntityIDString(entityID *uuid.UUID) string {
	if entityID == nil {
		return "nil"
	}
	return entityID.String()
}

// mapToConfiguration maps database result to domain Configuration
func (r *configurationRepository) mapToConfiguration(moduleName, configKey string, value json.RawMessage, source, dataType string, tenantID uuid.UUID, entityID *uuid.UUID) (*domain.Configuration, error) {
	module := domain.ModuleName(moduleName)
	key, err := domain.NewConfigKey(configKey)
	if err != nil {
		return nil, err
	}

	// Parse the value based on data type
	var configValue domain.ConfigValue
	if dataType != "" {
		dt := domain.DataType(dataType)
		configValue, err = r.parseConfigValue(value, dt)
		if err != nil {
			return nil, err
		}
	} else {
		// Default to string if data type is unknown
		var rawValue string
		if err := json.Unmarshal(value, &rawValue); err != nil {
			return nil, err
		}
		configValue = domain.NewStringValue(rawValue)
	}

	configSource := domain.ConfigSource(source)

	config, err := domain.NewConfiguration(
		entityID,
		module,
		key,
		configValue,
		configSource,
	)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// parseConfigValue parses a JSON value based on data type
func (r *configurationRepository) parseConfigValue(value json.RawMessage, dataType domain.DataType) (domain.ConfigValue, error) {
	switch dataType {
	case domain.DataTypeString:
		var str string
		if err := json.Unmarshal(value, &str); err != nil {
			return domain.ConfigValue{}, err
		}
		return domain.NewStringValue(str), nil

	case domain.DataTypeInteger:
		var intVal int
		if err := json.Unmarshal(value, &intVal); err != nil {
			return domain.ConfigValue{}, err
		}
		return domain.NewIntegerValue(intVal), nil

	case domain.DataTypeBoolean:
		var boolVal bool
		if err := json.Unmarshal(value, &boolVal); err != nil {
			return domain.ConfigValue{}, err
		}
		return domain.NewBooleanValue(boolVal), nil

	case domain.DataTypeDecimal:
		var floatVal float64
		if err := json.Unmarshal(value, &floatVal); err != nil {
			return domain.ConfigValue{}, err
		}
		return domain.NewDecimalValue(floatVal), nil

	case domain.DataTypeJSON:
		var jsonVal map[string]any
		if err := json.Unmarshal(value, &jsonVal); err != nil {
			return domain.ConfigValue{}, err
		}
		return domain.NewJSONValue(jsonVal), nil

	default:
		return domain.ConfigValue{}, domain.ErrInvalidDataType
	}
}

// invalidateConfigurationCache invalidates related cache entries
func (r *configurationRepository) invalidateConfigurationCache(ctx context.Context, tenantID uuid.UUID, entityID *uuid.UUID, module domain.ModuleName, key domain.ConfigKey) {
	// Invalidate specific configuration cache
	cacheKey := r.buildCacheKey("config:resolved", string(module), string(key), tenantID.String(), getEntityIDString(entityID))
	r.cache.Delete(ctx, cacheKey)

	// Invalidate related patterns
	pattern := r.buildCacheKey("config:resolved", string(module), string(key), tenantID.String(), "*")
	r.cache.DeletePattern(ctx, pattern)
}
