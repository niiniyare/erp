package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"awo/internal/core/audit"
	"awo/internal/core/settings/domain"
	"awo/internal/core/settings/repository"
	"awo/internal/shared/logger"
	"awo/internal/shared/metrics"
	"awo/internal/shared/tracing"
)

// ConfigurationService provides business logic for configuration management
type ConfigurationService interface {
	// Configuration Resolution
	GetEffectiveConfiguration(ctx context.Context, entityID *uuid.UUID, module domain.ModuleName, key domain.ConfigKey) (*domain.Configuration, error)
	ListEffectiveConfigurations(ctx context.Context, entityID *uuid.UUID, module domain.ModuleName) ([]*domain.Configuration, error)

	// Configuration Management
	GetTenantConfiguration(ctx context.Context, module domain.ModuleName, key domain.ConfigKey) (*domain.Configuration, error)
	UpdateTenantConfiguration(ctx context.Context, module domain.ModuleName, key domain.ConfigKey, value domain.ConfigValue, userID uuid.UUID) error
	DeleteTenantConfiguration(ctx context.Context, module domain.ModuleName, key domain.ConfigKey, userID uuid.UUID) error

	GetEntityConfiguration(ctx context.Context, entityID uuid.UUID, module domain.ModuleName, key domain.ConfigKey) (*domain.Configuration, error)
	UpdateEntityConfiguration(ctx context.Context, entityID uuid.UUID, module domain.ModuleName, key domain.ConfigKey, value domain.ConfigValue, userID uuid.UUID) error
	DeleteEntityConfiguration(ctx context.Context, entityID uuid.UUID, module domain.ModuleName, key domain.ConfigKey, userID uuid.UUID) error

	// Validation
	ValidateConfigurationValue(ctx context.Context, module domain.ModuleName, key domain.ConfigKey, value domain.ConfigValue) error

	// Bulk Operations
	BulkUpdateConfigurations(ctx context.Context, updates []repository.BulkConfigurationUpdate) (*repository.BulkOperationResult, error)

	// Search
	SearchConfigurations(ctx context.Context, criteria repository.SearchCriteria) (*repository.SearchResult, error)
}

// configurationService implements ConfigurationService
type configurationService struct {
	repo         repository.ConfigurationRepository
	auditService audit.Service
	logger       logger.Logger
	tracing      tracing.Service
	metrics      metrics.MetricsProvider
}

// NewConfigurationService creates a new configuration service
func NewConfigurationService(
	repo repository.ConfigurationRepository,
	auditService audit.Service,
	logger logger.Logger,
	tracing tracing.Service,
	metrics metrics.MetricsProvider,
) ConfigurationService {
	return &configurationService{
		repo:         repo,
		auditService: auditService,
		logger:       logger,
		tracing:      tracing,
		metrics:      metrics,
	}
}

// GetEffectiveConfiguration resolves the effective configuration value
func (s *configurationService) GetEffectiveConfiguration(ctx context.Context, entityID *uuid.UUID, module domain.ModuleName, key domain.ConfigKey) (*domain.Configuration, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.get_effective_configuration")
	defer span.End()

	timer := s.metrics.Timer("settings.get_effective_configuration", metrics.Fields{
		"module": string(module),
	})
	defer timer.Stop()

	config, err := s.repo.GetEffectiveConfiguration(ctx, entityID, module, key)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to get effective configuration", logger.Fields{
			"module":    string(module),
			"key":       string(key),
			"entity_id": entityID,
			"error":     err.Error(),
		})
		s.metrics.IncrementCounter("settings.get_effective_configuration_error", metrics.Fields{
			"module": string(module),
			"error":  err.Error(),
		})
		return nil, fmt.Errorf("failed to get effective configuration: %w", err)
	}

	s.logger.DebugContext(ctx, "Successfully retrieved effective configuration", logger.Fields{
		"module":    string(module),
		"key":       string(key),
		"entity_id": entityID,
		"source":    string(config.Source),
	})

	s.metrics.IncrementCounter("settings.get_effective_configuration_success", metrics.Fields{
		"module": string(module),
		"source": string(config.Source),
	})

	return config, nil
}

// ListEffectiveConfigurations lists all effective configurations
func (s *configurationService) ListEffectiveConfigurations(ctx context.Context, entityID *uuid.UUID, module domain.ModuleName) ([]*domain.Configuration, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.list_effective_configurations")
	defer span.End()

	timer := s.metrics.Timer("settings.list_effective_configurations", metrics.Fields{
		"module": string(module),
	})
	defer timer.Stop()

	configs, err := s.repo.ListEffectiveConfigurations(ctx, entityID, module)
	if err != nil {
		s.metrics.IncrementCounter("settings.list_effective_configurations_error", metrics.Fields{
			"module": string(module),
		})
		return nil, fmt.Errorf("failed to list effective configurations: %w", err)
	}

	s.metrics.IncrementCounter("settings.list_effective_configurations_success", metrics.Fields{
		"module": string(module),
		"count":  fmt.Sprintf("%d", len(configs)),
	})

	return configs, nil
}

// GetTenantConfiguration gets a tenant-level configuration
func (s *configurationService) GetTenantConfiguration(ctx context.Context, module domain.ModuleName, key domain.ConfigKey) (*domain.Configuration, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.get_tenant_configuration")
	defer span.End()

	config, err := s.repo.GetTenantConfiguration(ctx, module, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant configuration: %w", err)
	}

	return config, nil
}

// UpdateTenantConfiguration updates a tenant-level configuration with audit logging
func (s *configurationService) UpdateTenantConfiguration(ctx context.Context, module domain.ModuleName, key domain.ConfigKey, value domain.ConfigValue, userID uuid.UUID) error {
	ctx, span := s.tracing.StartSpan(ctx, "service.update_tenant_configuration")
	defer span.End()

	timer := s.metrics.Timer("settings.update_tenant_configuration", metrics.Fields{
		"module": string(module),
	})
	defer timer.Stop()

	s.logger.InfoContext(ctx, "Starting tenant configuration update", logger.Fields{
		"module": string(module),
		"key":    string(key),
		"user":   userID.String(),
	})

	// Validate the configuration value
	if err := s.ValidateConfigurationValue(ctx, module, key, value); err != nil {
		s.logger.WarnContext(ctx, "Configuration validation failed", logger.Fields{
			"module": string(module),
			"key":    string(key),
			"level":  "tenant",
			"error":  err.Error(),
		})
		s.metrics.IncrementCounter("settings.validation_error", metrics.Fields{
			"module": string(module),
			"level":  "tenant",
		})
		return fmt.Errorf("validation failed: %w", err)
	}

	// Get current value for audit trail
	currentConfig, _ := s.repo.GetTenantConfiguration(ctx, module, key)

	// Update the configuration
	err := s.repo.UpdateTenantConfiguration(ctx, module, key, value, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to update tenant configuration", logger.Fields{
			"module": string(module),
			"key":    string(key),
			"user":   userID.String(),
			"error":  err.Error(),
		})
		s.metrics.IncrementCounter("settings.update_tenant_configuration_error", metrics.Fields{
			"module": string(module),
		})
		return fmt.Errorf("failed to update tenant configuration: %w", err)
	}

	// Create audit event
	if err := s.createConfigurationAuditEvent(ctx, audit.CreateAuditEventRequest{
		UserID:        &userID,
		EventType:     "CONFIGURATION_UPDATE",
		EventCategory: "SETTINGS",
		Severity:      "INFO",
		Context:       s.buildAuditContext("tenant", module, key, currentConfig, &value),
	}); err != nil {
		// Log audit error but don't fail the operation
		s.logger.WarnContext(ctx, "Failed to create audit event for configuration update", logger.Fields{
			"module": string(module),
			"key":    string(key),
			"user":   userID.String(),
			"error":  err.Error(),
		})
		span.RecordError(err)
	}

	s.logger.InfoContext(ctx, "Successfully updated tenant configuration", logger.Fields{
		"module": string(module),
		"key":    string(key),
		"user":   userID.String(),
	})

	s.metrics.IncrementCounter("settings.update_tenant_configuration_success", metrics.Fields{
		"module": string(module),
	})

	return nil
}

// UpdateEntityConfiguration updates an entity-level configuration with audit logging
func (s *configurationService) UpdateEntityConfiguration(ctx context.Context, entityID uuid.UUID, module domain.ModuleName, key domain.ConfigKey, value domain.ConfigValue, userID uuid.UUID) error {
	ctx, span := s.tracing.StartSpan(ctx, "service.update_entity_configuration")
	defer span.End()

	timer := s.metrics.Timer("settings.update_entity_configuration", metrics.Fields{
		"module": string(module),
	})
	defer timer.Stop()

	// Validate the configuration value
	if err := s.ValidateConfigurationValue(ctx, module, key, value); err != nil {
		s.metrics.IncrementCounter("settings.validation_error", metrics.Fields{
			"module": string(module),
			"level":  "entity",
		})
		return fmt.Errorf("validation failed: %w", err)
	}

	// Get current value for audit trail
	currentConfig, _ := s.repo.GetEntityConfiguration(ctx, entityID, module, key)

	// Update the configuration
	err := s.repo.UpdateEntityConfiguration(ctx, entityID, module, key, value, userID)
	if err != nil {
		s.metrics.IncrementCounter("settings.update_entity_configuration_error", metrics.Fields{
			"module": string(module),
		})
		return fmt.Errorf("failed to update entity configuration: %w", err)
	}

	// Create audit event
	if err := s.createConfigurationAuditEvent(ctx, audit.CreateAuditEventRequest{
		UserID:        &userID,
		EventType:     "CONFIGURATION_UPDATE",
		EventCategory: "SETTINGS",
		Severity:      "INFO",
		EntityID:      &entityID,
		Context:       s.buildAuditContext("entity", module, key, currentConfig, &value),
	}); err != nil {
		// Log audit error but don't fail the operation
		span.RecordError(err)
	}

	s.metrics.IncrementCounter("settings.update_entity_configuration_success", metrics.Fields{
		"module": string(module),
	})

	return nil
}

// DeleteTenantConfiguration deletes a tenant-level configuration with audit logging
func (s *configurationService) DeleteTenantConfiguration(ctx context.Context, module domain.ModuleName, key domain.ConfigKey, userID uuid.UUID) error {
	ctx, span := s.tracing.StartSpan(ctx, "service.delete_tenant_configuration")
	defer span.End()

	// Get current value for audit trail
	currentConfig, _ := s.repo.GetTenantConfiguration(ctx, module, key)

	// Delete the configuration
	err := s.repo.DeleteTenantConfiguration(ctx, module, key, userID)
	if err != nil {
		s.metrics.IncrementCounter("settings.delete_tenant_configuration_error", metrics.Fields{
			"module": string(module),
		})
		return fmt.Errorf("failed to delete tenant configuration: %w", err)
	}

	// Create audit event
	if err := s.createConfigurationAuditEvent(ctx, audit.CreateAuditEventRequest{
		UserID:        &userID,
		EventType:     "CONFIGURATION_DELETE",
		EventCategory: "SETTINGS",
		Severity:      "WARN",
		Context:       s.buildAuditContext("tenant", module, key, currentConfig, nil),
	}); err != nil {
		// Log audit error but don't fail the operation
		span.RecordError(err)
	}

	s.metrics.IncrementCounter("settings.delete_tenant_configuration_success", metrics.Fields{
		"module": string(module),
	})

	return nil
}

// GetEntityConfiguration gets an entity-level configuration
func (s *configurationService) GetEntityConfiguration(ctx context.Context, entityID uuid.UUID, module domain.ModuleName, key domain.ConfigKey) (*domain.Configuration, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.get_entity_configuration")
	defer span.End()

	config, err := s.repo.GetEntityConfiguration(ctx, entityID, module, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity configuration: %w", err)
	}

	return config, nil
}

// DeleteEntityConfiguration deletes an entity-level configuration with audit logging
func (s *configurationService) DeleteEntityConfiguration(ctx context.Context, entityID uuid.UUID, module domain.ModuleName, key domain.ConfigKey, userID uuid.UUID) error {
	ctx, span := s.tracing.StartSpan(ctx, "service.delete_entity_configuration")
	defer span.End()

	// Get current value for audit trail
	currentConfig, _ := s.repo.GetEntityConfiguration(ctx, entityID, module, key)

	// Delete the configuration
	err := s.repo.DeleteEntityConfiguration(ctx, entityID, module, key, userID)
	if err != nil {
		s.metrics.IncrementCounter("settings.delete_entity_configuration_error", metrics.Fields{
			"module": string(module),
		})
		return fmt.Errorf("failed to delete entity configuration: %w", err)
	}

	// Create audit event
	if err := s.createConfigurationAuditEvent(ctx, audit.CreateAuditEventRequest{
		UserID:        &userID,
		EventType:     "CONFIGURATION_DELETE",
		EventCategory: "SETTINGS",
		Severity:      "WARN",
		EntityID:      &entityID,
		Context:       s.buildAuditContext("entity", module, key, currentConfig, nil),
	}); err != nil {
		// Log audit error but don't fail the operation
		span.RecordError(err)
	}

	s.metrics.IncrementCounter("settings.delete_entity_configuration_success", metrics.Fields{
		"module": string(module),
	})

	return nil
}

// ValidateConfigurationValue validates a configuration value
func (s *configurationService) ValidateConfigurationValue(ctx context.Context, module domain.ModuleName, key domain.ConfigKey, value domain.ConfigValue) error {
	ctx, span := s.tracing.StartSpan(ctx, "service.validate_configuration_value")
	defer span.End()

	// Get configuration definition for validation rules
	definition, err := s.repo.GetConfigDefinition(ctx, module, key)
	if err != nil {
		if err == domain.ErrConfigDefinitionNotFound {
			// If no definition exists, allow the value (permissive mode)
			return nil
		}
		return fmt.Errorf("failed to get configuration definition: %w", err)
	}

	// Validate data type
	if definition.DataType != value.DataType {
		return fmt.Errorf("invalid data type: expected %s, got %s", definition.DataType, value.DataType)
	}

	// Apply validation rules
	if !definition.ValidationRules.IsEmpty() {
		if err := definition.ValidationRules.Validate(value); err != nil {
			return fmt.Errorf("validation rule failed: %w", err)
		}
	}

	return nil
}

// BulkUpdateConfigurations performs bulk configuration updates
func (s *configurationService) BulkUpdateConfigurations(ctx context.Context, updates []repository.BulkConfigurationUpdate) (*repository.BulkOperationResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.bulk_update_configurations")
	defer span.End()

	s.logger.InfoContext(ctx, "Starting bulk configuration update", logger.Fields{
		"update_count": len(updates),
	})

	timer := s.metrics.Timer("settings.bulk_update_configurations", metrics.Fields{
		"count": fmt.Sprintf("%d", len(updates)),
	})
	defer timer.Stop()

	// Validate all updates before processing
	for i, update := range updates {
		if err := s.ValidateConfigurationValue(ctx, update.Module, update.ConfigKey, update.Value); err != nil {
			s.metrics.IncrementCounter("settings.bulk_validation_error", nil)
			return nil, fmt.Errorf("validation failed for update %d: %w", i, err)
		}
	}

	// Process bulk updates
	result, err := s.repo.BulkUpdateConfigurations(ctx, updates)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to perform bulk configuration update", logger.Fields{
			"update_count": len(updates),
			"error":        err.Error(),
		})
		s.metrics.IncrementCounter("settings.bulk_update_configurations_error", nil)
		return nil, fmt.Errorf("failed to perform bulk update: %w", err)
	}

	s.logger.InfoContext(ctx, "Successfully completed bulk configuration update", logger.Fields{
		"total_targets":      result.TotalTargets,
		"processed_targets":  result.ProcessedTargets,
		"successful_updates": result.SuccessfulUpdates,
		"failed_updates":     result.FailedUpdates,
		"duration_ms":        result.DurationMS,
	})

	// Create audit event for bulk operation
	if len(updates) > 0 {
		if err := s.createConfigurationAuditEvent(ctx, audit.CreateAuditEventRequest{
			UserID:        &updates[0].UserID,
			EventType:     "CONFIGURATION_BULK_UPDATE",
			EventCategory: "SETTINGS",
			Severity:      "INFO",
			Context:       s.buildBulkAuditContext(result),
		}); err != nil {
			// Log audit error but don't fail the operation
			span.RecordError(err)
		}
	}

	s.metrics.IncrementCounter("settings.bulk_update_configurations_success", metrics.Fields{
		"processed":  fmt.Sprintf("%d", result.ProcessedTargets),
		"successful": fmt.Sprintf("%d", result.SuccessfulUpdates),
		"failed":     fmt.Sprintf("%d", result.FailedUpdates),
	})

	return result, nil
}

// SearchConfigurations searches configurations
func (s *configurationService) SearchConfigurations(ctx context.Context, criteria repository.SearchCriteria) (*repository.SearchResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.search_configurations")
	defer span.End()

	timer := s.metrics.Timer("settings.search_configurations", nil)
	defer timer.Stop()

	result, err := s.repo.SearchConfigurations(ctx, criteria)
	if err != nil {
		s.metrics.IncrementCounter("settings.search_configurations_error", nil)
		return nil, fmt.Errorf("failed to search configurations: %w", err)
	}

	s.metrics.IncrementCounter("settings.search_configurations_success", metrics.Fields{
		"results": fmt.Sprintf("%d", len(result.Configurations)),
	})

	return result, nil
}

// Helper methods

// createConfigurationAuditEvent creates an audit event for configuration changes
func (s *configurationService) createConfigurationAuditEvent(ctx context.Context, req audit.CreateAuditEventRequest) error {
	// tenantID, _ := shared.GetTenantID(ctx)
	_, err := s.auditService.CreateAuditEvent(ctx, req)
	return err
}

// buildAuditContext creates audit context for configuration changes
func (s *configurationService) buildAuditContext(level string, module domain.ModuleName, key domain.ConfigKey, oldConfig *domain.Configuration, newValue *domain.ConfigValue) json.RawMessage {
	context := map[string]any{
		"level":      level,
		"module":     string(module),
		"config_key": string(key),
		"timestamp":  time.Now().UTC(),
	}

	if oldConfig != nil {
		context["old_value"] = oldConfig.Value.Raw
		context["old_source"] = string(oldConfig.Source)
	}

	if newValue != nil {
		context["new_value"] = newValue.Raw
		context["new_type"] = string(newValue.DataType)
	}

	contextBytes, _ := json.Marshal(context)
	return json.RawMessage(contextBytes)
}

// buildBulkAuditContext creates audit context for bulk operations
func (s *configurationService) buildBulkAuditContext(result *repository.BulkOperationResult) json.RawMessage {
	context := map[string]any{
		"operation_id":       result.OperationID,
		"total_targets":      result.TotalTargets,
		"processed_targets":  result.ProcessedTargets,
		"successful_updates": result.SuccessfulUpdates,
		"failed_updates":     result.FailedUpdates,
		"duration_ms":        result.DurationMS,
		"timestamp":          time.Now().UTC(),
	}

	if len(result.Errors) > 0 {
		context["error_count"] = len(result.Errors)
		context["errors"] = result.Errors
	}

	contextBytes, _ := json.Marshal(context)
	return json.RawMessage(contextBytes)
}
