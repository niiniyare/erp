package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	db "awo.so/db/sqlc"
	"awo.so/internal/core/settings/domain"
	"awo.so/internal/shared"
	"awo.so/internal/shared/metrics"
)

// Implement the remaining repository methods

// DeleteTenantConfiguration deletes a tenant-level configuration
func (r *configurationRepository) DeleteTenantConfiguration(ctx context.Context, module domain.ModuleName, key domain.ConfigKey, userID uuid.UUID) error {
	// Get tenant ID from context
	tenantID, _ := shared.GetTenantID(ctx)
	ctx, span := r.tracing.StartSpan(ctx, "repository.delete_tenant_configuration")
	defer span.End()

	// Ensure tenant context for RLS
	ctx = shared.WithTenantID(ctx, tenantID)

	// Get current value for audit
	currentConfig, _ := r.GetTenantConfiguration(ctx, module, key)

	jsonPath := []string{fmt.Sprintf("%s.%s", module, key)}

	err := r.store.DeleteTenantConfigurationSettings(ctx, jsonPath)
	if err != nil {
		return fmt.Errorf("failed to delete tenant configuration: %w", err)
	}

	// Create audit record
	if currentConfig != nil {
		// For delete operations, NewValue should be empty
		emptyValue := domain.NewStringValue("")
		auditRecord := &AuditRecord{
			TenantID:  tenantID,
			ConfigKey: fmt.Sprintf("%s.%s", module, key),
			OldValue:  &currentConfig.Value,
			NewValue:  emptyValue,
			Source:    domain.ConfigSourceTenant,
			Operation: "delete",
			UserID:    userID,
			AppliedAt: time.Now(),
		}

		if err := r.CreateAuditRecord(ctx, auditRecord); err != nil {
			span.RecordError(err)
		}
	}

	// Invalidate cache
	r.invalidateConfigurationCache(ctx, tenantID, nil, module, key)

	return nil
}

// DeleteEntityConfiguration deletes an entity-level configuration
func (r *configurationRepository) DeleteEntityConfiguration(ctx context.Context, entityID uuid.UUID, module domain.ModuleName, key domain.ConfigKey, userID uuid.UUID) error {
	// Get tenant ID from context
	tenantID, _ := shared.GetTenantID(ctx)
	ctx, span := r.tracing.StartSpan(ctx, "repository.delete_entity_configuration")
	defer span.End()

	// Ensure tenant context for RLS
	ctx = shared.WithTenantID(ctx, tenantID)

	// Get current value for audit
	currentConfig, _ := r.GetEntityConfiguration(ctx, entityID, module, key)

	jsonPath := []string{fmt.Sprintf("%s.%s", module, key)}

	err := r.store.DeleteEntityConfiguration(ctx, db.DeleteEntityConfigurationParams{
		ConfigPath: jsonPath,
		EntityID:   entityID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete entity configuration: %w", err)
	}

	// Create audit record
	if currentConfig != nil {
		// For delete operations, NewValue should be empty
		emptyValue := domain.NewStringValue("")
		auditRecord := &AuditRecord{
			TenantID:  tenantID,
			EntityID:  &entityID,
			ConfigKey: fmt.Sprintf("%s.%s", module, key),
			OldValue:  &currentConfig.Value,
			NewValue:  emptyValue,
			Source:    domain.ConfigSourceEntity,
			Operation: "delete",
			UserID:    userID,
			AppliedAt: time.Now(),
		}

		if err := r.CreateAuditRecord(ctx, auditRecord); err != nil {
			span.RecordError(err)
		}
	}

	// Invalidate cache
	r.invalidateConfigurationCache(ctx, tenantID, &entityID, module, key)

	return nil
}

// GetConfigDefinition gets a configuration definition
func (r *configurationRepository) GetConfigDefinition(ctx context.Context, module domain.ModuleName, key domain.ConfigKey) (*ConfigDefinition, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.get_config_definition")
	defer span.End()

	result, err := r.store.GetConfigDefinition(ctx, db.GetConfigDefinitionParams{
		ModuleName: string(module),
		ConfigKey:  string(key),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrConfigDefinitionNotFound
		}
		return nil, fmt.Errorf("failed to get config definition: %w", err)
	}

	return r.mapToConfigDefinition(result)
}

// ListConfigDefinitions lists configuration definitions for a module
func (r *configurationRepository) ListConfigDefinitions(ctx context.Context, module domain.ModuleName) ([]*ConfigDefinition, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.list_config_definitions")
	defer span.End()

	moduleFilter := ""
	if module != "" {
		moduleFilter = string(module)
	}

	results, err := r.store.ListConfigDefinitions(ctx, &moduleFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to list config definitions: %w", err)
	}

	definitions := make([]*ConfigDefinition, 0, len(results))
	for _, result := range results {
		def, err := r.mapToConfigDefinition(result)
		if err != nil {
			continue // Skip invalid definitions
		}
		definitions = append(definitions, def)
	}

	return definitions, nil
}

// CreateConfigDefinition creates a new configuration definition
func (r *configurationRepository) CreateConfigDefinition(ctx context.Context, def *ConfigDefinition) error {
	ctx, span := r.tracing.StartSpan(ctx, "repository.create_config_definition")
	defer span.End()

	var defaultValueJSON []byte
	if def.DefaultValue != nil {
		var err error
		defaultValueJSON, err = json.Marshal(def.DefaultValue.Raw)
		if err != nil {
			return fmt.Errorf("failed to marshal default value: %w", err)
		}
	}

	validationRulesJSON, err := json.Marshal(def.ValidationRules)
	if err != nil {
		return fmt.Errorf("failed to marshal validation rules: %w", err)
	}

	var requiredFeatureFlag *string
	if def.RequiredFeatureFlag != nil {
		flag := string(*def.RequiredFeatureFlag)
		requiredFeatureFlag = &flag
	}

	_, err = r.store.CreateConfigDefinition(ctx, db.CreateConfigDefinitionParams{
		ModuleName:          string(def.ModuleName),
		ConfigKey:           string(def.ConfigKey),
		ConfigType:          string(def.DataType),
		DefaultValue:        defaultValueJSON,
		ValidationRules:     validationRulesJSON,
		Description:         def.Description,
		RequiredFeatureFlag: requiredFeatureFlag,
	})
	if err != nil {
		return fmt.Errorf("failed to create config definition: %w", err)
	}

	return nil
}

// UpdateConfigDefinition updates a configuration definition
func (r *configurationRepository) UpdateConfigDefinition(ctx context.Context, def *ConfigDefinition) error {
	ctx, span := r.tracing.StartSpan(ctx, "repository.update_config_definition")
	defer span.End()

	var defaultValueJSON []byte
	if def.DefaultValue != nil {
		var err error
		defaultValueJSON, err = json.Marshal(def.DefaultValue.Raw)
		if err != nil {
			return fmt.Errorf("failed to marshal default value: %w", err)
		}
	}

	validationRulesJSON, err := json.Marshal(def.ValidationRules)
	if err != nil {
		return fmt.Errorf("failed to marshal validation rules: %w", err)
	}

	var requiredFeatureFlag *string
	if def.RequiredFeatureFlag != nil {
		flag := string(*def.RequiredFeatureFlag)
		requiredFeatureFlag = &flag
	}

	var dataType *string
	if def.DataType != "" {
		dt := string(def.DataType)
		dataType = &dt
	}

	_, err = r.store.UpdateConfigDefinition(ctx, db.UpdateConfigDefinitionParams{
		ConfigType:          dataType,
		DefaultValue:        defaultValueJSON,
		ValidationRules:     validationRulesJSON,
		Description:         &def.Description,
		RequiredFeatureFlag: requiredFeatureFlag,
		ModuleName:          string(def.ModuleName),
		ConfigKey:           string(def.ConfigKey),
	})
	if err != nil {
		return fmt.Errorf("failed to update config definition: %w", err)
	}

	return nil
}

// CreateAuditRecord creates a configuration audit record
func (r *configurationRepository) CreateAuditRecord(ctx context.Context, record *AuditRecord) error {
	ctx, span := r.tracing.StartSpan(ctx, "repository.create_audit_record")
	defer span.End()

	// Tenant ID will be set automatically by current_tenant_id() in SQL

	var oldValueJSON []byte
	if record.OldValue != nil {
		var err error
		oldValueJSON, err = json.Marshal(record.OldValue.Raw)
		if err != nil {
			return fmt.Errorf("failed to marshal old value: %w", err)
		}
	}

	newValueJSON, err := json.Marshal(record.NewValue.Raw)
	if err != nil {
		return fmt.Errorf("failed to marshal new value: %w", err)
	}

	// Split "module.key" → ModuleName + ConfigKeyName (matches DB schema split columns)
	parts := strings.SplitN(record.ConfigKey, ".", 2)
	moduleName := parts[0]
	configKeyName := ""
	if len(parts) == 2 {
		configKeyName = parts[1]
	}

	_, err = r.store.CreateConfigurationAudit(ctx, db.CreateConfigurationAuditParams{
		EntityID:      record.EntityID,
		ModuleName:    moduleName,
		ConfigKeyName: configKeyName,
		OldValue:      oldValueJSON,
		NewValue:      newValueJSON,
		Source:        string(record.Source),
		Operation:     record.Operation,
		UserID:        record.UserID,
		SessionID:     record.SessionID,
		CorrelationID: record.CorrelationID,
	})
	if err != nil {
		return fmt.Errorf("failed to create audit record: %w", err)
	}

	return nil
}

// GetConfigurationHistory gets configuration change history
func (r *configurationRepository) GetConfigurationHistory(ctx context.Context, entityID *uuid.UUID, configKey string, limit, offset int) ([]*AuditRecord, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.get_configuration_history")
	defer span.End()

	// Tenant ID will be resolved automatically by current_tenant_id() in SQL

	// Split "module.key" for the new per-column filter params
	parts := strings.SplitN(configKey, ".", 2)
	moduleFilter := parts[0]
	keyFilter := ""
	if len(parts) == 2 {
		keyFilter = parts[1]
	}

	results, err := r.store.GetConfigurationHistory(ctx, db.GetConfigurationHistoryParams{
		EntityIDFilter:  entityID,
		ModuleFilter:    &moduleFilter,
		ConfigKeyFilter: &keyFilter,
		OffsetCount:     int32(offset),
		LimitCount:      int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration history: %w", err)
	}

	records := make([]*AuditRecord, 0, len(results))
	for _, result := range results {
		record, err := r.mapToAuditRecord(result)
		if err != nil {
			continue // Skip invalid records
		}
		records = append(records, record)
	}

	return records, nil
}

// BulkUpdateConfigurations performs bulk configuration updates
func (r *configurationRepository) BulkUpdateConfigurations(ctx context.Context, updates []BulkConfigurationUpdate) (*BulkOperationResult, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.bulk_update_configurations")
	defer span.End()

	timer := r.metrics.Timer("settings.bulk_update", metrics.Fields{
		"count": fmt.Sprintf("%d", len(updates)),
	})
	defer timer.Stop()

	startTime := time.Now()
	operationID := uuid.New().String()

	result := &BulkOperationResult{
		OperationID:  operationID,
		TotalTargets: len(updates),
		StartedAt:    startTime,
		Errors:       make([]BulkOperationError, 0),
	}

	// Group updates by target type for efficient processing
	updatesByType := r.groupUpdatesByTenant(updates)

	for updateType, typeUpdates := range updatesByType {
		err := r.processTenantBulkUpdates(ctx, updateType, typeUpdates, result)
		if err != nil {
			return nil, err
		}
	}

	completedAt := time.Now()
	result.CompletedAt = &completedAt
	result.DurationMS = completedAt.Sub(startTime).Milliseconds()
	result.ProcessedTargets = result.SuccessfulUpdates + result.FailedUpdates

	return result, nil
}

// SearchConfigurations searches configurations based on criteria
func (r *configurationRepository) SearchConfigurations(ctx context.Context, criteria SearchCriteria) (*SearchResult, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.search_configurations")
	defer span.End()

	startTime := time.Now()

	// Tenant ID will be resolved automatically by current_tenant_id() in SQL

	// Build search parameters
	var modules []string
	for _, module := range criteria.Modules {
		modules = append(modules, string(module))
	}

	var sources []string
	for _, source := range criteria.Sources {
		sources = append(sources, string(source))
	}

	var modifiedAfter time.Time
	if criteria.ModifiedAfter != nil {
		modifiedAfter = *criteria.ModifiedAfter
	}

	sortBy := criteria.SortBy
	if sortBy == "" {
		sortBy = "config_key"
	}

	results, err := r.store.SearchConfigurations(ctx, db.SearchConfigurationsParams{
		ModulesFilter:  modules,
		SearchTerm:     &criteria.ValueContains,
		SourcesFilter:  sources,
		UpdatedAfter:   sql.NullTime{Time: modifiedAfter, Valid: !modifiedAfter.IsZero()},
		SortBy:         sortBy,
		OffsetCount:    int32(criteria.Offset),
		LimitCount:     int32(criteria.Limit),
		EntityIDFilter: criteria.EntityID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search configurations: %w", err)
	}

	configurations := make([]*ConfigurationSearchResult, 0, len(results))
	for _, result := range results {
		config, err := r.mapToSearchResult(result)
		if err != nil {
			continue // Skip invalid results
		}
		configurations = append(configurations, config)
	}

	queryTime := time.Since(startTime).Milliseconds()

	return &SearchResult{
		Configurations: configurations,
		TotalCount:     len(configurations), // This would be a separate count query in production
		QueryTimeMS:    queryTime,
		Filters:        criteria,
	}, nil
}

// Helper methods for bulk operations and mapping

// groupUpdatesByTenant groups bulk updates by tenant for efficient processing
func (r *configurationRepository) groupUpdatesByTenant(updates []BulkConfigurationUpdate) map[string][]BulkConfigurationUpdate {
	// Since tenant ID comes from context, we group by target type instead
	updatesByType := make(map[string][]BulkConfigurationUpdate)

	for _, update := range updates {
		typeUpdates := updatesByType[update.TargetType]
		updatesByType[update.TargetType] = append(typeUpdates, update)
	}

	return updatesByType
}

// processTenantBulkUpdates processes bulk updates for a single tenant
func (r *configurationRepository) processTenantBulkUpdates(ctx context.Context, updateType string, updates []BulkConfigurationUpdate, result *BulkOperationResult) error {
	// Tenant context is already set at the boundary

	for _, update := range updates {
		err := r.processSingleBulkUpdate(ctx, update)
		if err != nil {
			result.FailedUpdates++
			result.Errors = append(result.Errors, BulkOperationError{
				TargetID:  update.TargetID.String(),
				ConfigKey: fmt.Sprintf("%s.%s", update.Module, update.ConfigKey),
				Error:     err.Error(),
			})
		} else {
			result.SuccessfulUpdates++
		}
	}

	return nil
}

// processSingleBulkUpdate processes a single bulk update
func (r *configurationRepository) processSingleBulkUpdate(ctx context.Context, update BulkConfigurationUpdate) error {
	switch update.TargetType {
	case "tenant":
		return r.UpdateTenantConfiguration(ctx, update.Module, update.ConfigKey, update.Value, update.UserID)
	case "entity":
		if update.EntityID == nil {
			return fmt.Errorf("entity ID required for entity-level configuration")
		}
		return r.UpdateEntityConfiguration(ctx, *update.EntityID, update.Module, update.ConfigKey, update.Value, update.UserID)
	default:
		return fmt.Errorf("invalid target type: %s", update.TargetType)
	}
}

// mapToConfigDefinition maps database result to ConfigDefinition
// TODO: Update this when SQLC generates config_definitions types
func (r *configurationRepository) mapToConfigDefinition(result any) (*ConfigDefinition, error) {
	// For now, return a placeholder since the SQLC types aren't generated yet
	return nil, fmt.Errorf("config definition mapping not yet implemented - SQLC types not generated")
}

// mapToAuditRecord maps database result to AuditRecord
// TODO: Update this when SQLC generates configuration_audit types
func (r *configurationRepository) mapToAuditRecord(result any) (*AuditRecord, error) {
	// For now, return a placeholder since the SQLC types aren't generated yet
	return nil, fmt.Errorf("audit record mapping not yet implemented - SQLC types not generated")
}

// mapToSearchResult maps database result to ConfigurationSearchResult
// TODO: Update this when SQLC generates search types
func (r *configurationRepository) mapToSearchResult(result any) (*ConfigurationSearchResult, error) {
	// For now, return a placeholder since the SQLC types aren't generated yet
	return nil, fmt.Errorf("search result mapping not yet implemented - SQLC types not generated")
}

// Template-related stub implementations (can be expanded later)

// CreateTemplate creates a new template (stub implementation)
func (r *configurationRepository) CreateTemplate(ctx context.Context, template *domain.Template) error {
	// TODO: Implement template creation
	return fmt.Errorf("template creation not yet implemented")
}

// GetTemplate gets a template by ID (stub implementation)
func (r *configurationRepository) GetTemplate(ctx context.Context, templateID domain.TemplateID) (*domain.Template, error) {
	// TODO: Implement template retrieval
	return nil, fmt.Errorf("template retrieval not yet implemented")
}

// ListTemplates lists templates with filters (stub implementation)
func (r *configurationRepository) ListTemplates(ctx context.Context, filters TemplateFilters) ([]*domain.Template, error) {
	// TODO: Implement template listing
	return nil, fmt.Errorf("template listing not yet implemented")
}

// UpdateTemplate updates a template (stub implementation)
func (r *configurationRepository) UpdateTemplate(ctx context.Context, template *domain.Template) error {
	// TODO: Implement template updating
	return fmt.Errorf("template updating not yet implemented")
}

// DeactivateTemplate deactivates a template (stub implementation)
func (r *configurationRepository) DeactivateTemplate(ctx context.Context, templateID domain.TemplateID) error {
	// TODO: Implement template deactivation
	return fmt.Errorf("template deactivation not yet implemented")
}

// ApplyTemplate applies a template to a target (stub implementation)
func (r *configurationRepository) ApplyTemplate(ctx context.Context, templateID domain.TemplateID, target TemplateTarget, options ApplyOptions, userID uuid.UUID) (*domain.ApplicationResult, error) {
	// TODO: Implement template application
	return nil, fmt.Errorf("template application not yet implemented")
}

// GetTemplateApplicationHistory gets template application history (stub implementation)
func (r *configurationRepository) GetTemplateApplicationHistory(ctx context.Context, entityID *uuid.UUID, limit, offset int) ([]*TemplateApplication, error) {
	// TODO: Implement template application history
	return nil, fmt.Errorf("template application history not yet implemented")
}
