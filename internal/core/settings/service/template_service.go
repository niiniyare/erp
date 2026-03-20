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

// TemplateService provides business logic for template management
type TemplateService interface {
	// Template Management
	CreateTemplate(ctx context.Context, template *domain.Template) error
	GetTemplate(ctx context.Context, templateID domain.TemplateID) (*domain.Template, error)
	ListTemplates(ctx context.Context, filters repository.TemplateFilters) ([]*domain.Template, error)
	UpdateTemplate(ctx context.Context, template *domain.Template) error
	DeactivateTemplate(ctx context.Context, templateID domain.TemplateID, userID uuid.UUID) error

	// Template Application
	ApplyTemplate(ctx context.Context, templateID domain.TemplateID, target repository.TemplateTarget, options repository.ApplyOptions, userID uuid.UUID) (*domain.ApplicationResult, error)
	ValidateTemplateApplication(ctx context.Context, templateID domain.TemplateID, target repository.TemplateTarget) (*TemplateValidationResult, error)
	GetTemplateApplicationHistory(ctx context.Context, entityID *uuid.UUID, limit, offset int) ([]*repository.TemplateApplication, error)

	// Template Analytics
	GetTemplateUsageStats(ctx context.Context, templateID domain.TemplateID, startTime, endTime time.Time) (*TemplateUsageStats, error)
	GetMostUsedTemplates(ctx context.Context, limit int, startTime, endTime time.Time) ([]*TemplateUsageStats, error)
}

// templateService implements TemplateService
type templateService struct {
	repo         repository.ConfigurationRepository
	auditService audit.Service
	logger       logger.Logger
	tracing      tracing.Service
	metrics      metrics.MetricsProvider
}

// NewTemplateService creates a new template service
func NewTemplateService(
	repo repository.ConfigurationRepository,
	auditService audit.Service,
	logger logger.Logger,
	tracing tracing.Service,
	metrics metrics.MetricsProvider,
) TemplateService {
	return &templateService{
		repo:         repo,
		auditService: auditService,
		logger:       logger,
		tracing:      tracing,
		metrics:      metrics,
	}
}

// CreateTemplate creates a new configuration template with validation
func (s *templateService) CreateTemplate(ctx context.Context, template *domain.Template) error {
	ctx, span := s.tracing.StartSpan(ctx, "service.create_template")
	defer span.End()

	timer := s.metrics.Timer("settings.create_template", metrics.Fields{
		"category": string(template.Category),
	})
	defer timer.Stop()

	s.logger.InfoContext(ctx, "Creating new template", logger.Fields{
		"template_name":       template.Name,
		"template_category":   string(template.Category),
		"configuration_count": len(template.Configurations),
	})

	// Validate template structure
	if err := s.validateTemplate(ctx, template); err != nil {
		s.logger.WarnContext(ctx, "Template validation failed", logger.Fields{
			"template_name": template.Name,
			"error":         err.Error(),
		})
		s.metrics.IncrementCounter("settings.template_validation_error", nil)
		return fmt.Errorf("template validation failed: %w", err)
	}

	// Create the template
	err := s.repo.CreateTemplate(ctx, template)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to create template", logger.Fields{
			"template_name": template.Name,
			"template_id":   template.ID,
			"error":         err.Error(),
		})
		s.metrics.IncrementCounter("settings.create_template_error", nil)
		return fmt.Errorf("failed to create template: %w", err)
	}

	s.logger.InfoContext(ctx, "Successfully created template", logger.Fields{
		"template_name":       template.Name,
		"template_id":         template.ID,
		"template_category":   string(template.Category),
		"configuration_count": len(template.Configurations),
	})

	// Create audit event
	if err := s.createTemplateAuditEvent(ctx, audit.CreateAuditEventRequest{
		UserID:        &template.CreatedBy,
		EventType:     "TEMPLATE_CREATE",
		EventCategory: "SETTINGS",
		Severity:      "INFO",
		Context:       s.buildTemplateAuditContext("create", template, nil),
	}); err != nil {
		span.RecordError(err)
	}

	s.metrics.IncrementCounter("settings.create_template_success", metrics.Fields{
		"category": string(template.Category),
	})

	return nil
}

// GetTemplate retrieves a template by ID
func (s *templateService) GetTemplate(ctx context.Context, templateID domain.TemplateID) (*domain.Template, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.get_template")
	defer span.End()

	template, err := s.repo.GetTemplate(ctx, templateID)
	if err != nil {
		s.metrics.IncrementCounter("settings.get_template_error", nil)
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	s.metrics.IncrementCounter("settings.get_template_success", nil)
	return template, nil
}

// ListTemplates lists templates with filtering
func (s *templateService) ListTemplates(ctx context.Context, filters repository.TemplateFilters) ([]*domain.Template, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.list_templates")
	defer span.End()

	templates, err := s.repo.ListTemplates(ctx, filters)
	if err != nil {
		s.metrics.IncrementCounter("settings.list_templates_error", nil)
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}

	s.metrics.IncrementCounter("settings.list_templates_success", metrics.Fields{
		"count": fmt.Sprintf("%d", len(templates)),
	})

	return templates, nil
}

// UpdateTemplate updates an existing template with audit logging
func (s *templateService) UpdateTemplate(ctx context.Context, template *domain.Template) error {
	ctx, span := s.tracing.StartSpan(ctx, "service.update_template")
	defer span.End()

	timer := s.metrics.Timer("settings.update_template", nil)
	defer timer.Stop()

	// Get current template for audit trail
	currentTemplate, _ := s.repo.GetTemplate(ctx, template.ID)

	// Validate updated template
	if err := s.validateTemplate(ctx, template); err != nil {
		s.metrics.IncrementCounter("settings.template_validation_error", nil)
		return fmt.Errorf("template validation failed: %w", err)
	}

	// Update the template
	err := s.repo.UpdateTemplate(ctx, template)
	if err != nil {
		s.metrics.IncrementCounter("settings.update_template_error", nil)
		return fmt.Errorf("failed to update template: %w", err)
	}

	// Create audit event
	if err := s.createTemplateAuditEvent(ctx, audit.CreateAuditEventRequest{
		UserID:        &template.CreatedBy, // Templates don't have UpdatedBy, use CreatedBy
		EventType:     "TEMPLATE_UPDATE",
		EventCategory: "SETTINGS",
		Severity:      "INFO",
		Context:       s.buildTemplateAuditContext("update", template, currentTemplate),
	}); err != nil {
		span.RecordError(err)
	}

	s.metrics.IncrementCounter("settings.update_template_success", nil)
	return nil
}

// DeactivateTemplate deactivates a template
func (s *templateService) DeactivateTemplate(ctx context.Context, templateID domain.TemplateID, userID uuid.UUID) error {
	ctx, span := s.tracing.StartSpan(ctx, "service.deactivate_template")
	defer span.End()

	// Get current template for audit trail
	currentTemplate, err := s.repo.GetTemplate(ctx, templateID)
	if err != nil {
		return fmt.Errorf("failed to get template for deactivation: %w", err)
	}

	// Deactivate the template
	err = s.repo.DeactivateTemplate(ctx, templateID)
	if err != nil {
		s.metrics.IncrementCounter("settings.deactivate_template_error", nil)
		return fmt.Errorf("failed to deactivate template: %w", err)
	}

	// Create audit event
	if err := s.createTemplateAuditEvent(ctx, audit.CreateAuditEventRequest{
		UserID:        &userID,
		EventType:     "TEMPLATE_DEACTIVATE",
		EventCategory: "SETTINGS",
		Severity:      "WARN",
		Context:       s.buildTemplateAuditContext("deactivate", currentTemplate, nil),
	}); err != nil {
		span.RecordError(err)
	}

	s.metrics.IncrementCounter("settings.deactivate_template_success", nil)
	return nil
}

// ApplyTemplate applies a template to target entities with validation
func (s *templateService) ApplyTemplate(ctx context.Context, templateID domain.TemplateID, target repository.TemplateTarget, options repository.ApplyOptions, userID uuid.UUID) (*domain.ApplicationResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.apply_template")
	defer span.End()

	timer := s.metrics.Timer("settings.apply_template", metrics.Fields{
		"target_type": target.Type,
	})
	defer timer.Stop()

	s.logger.InfoContext(ctx, "Starting template application", logger.Fields{
		"template_id":   templateID,
		"target_type":   target.Type,
		"target_entity": target.EntityID,
	})

	// Get and validate template
	template, err := s.repo.GetTemplate(ctx, templateID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to retrieve template for application", logger.Fields{
			"template_id": templateID,
			"error":       err.Error(),
		})
		s.metrics.IncrementCounter("settings.apply_template_error", nil)
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	if !template.IsActive {
		s.logger.WarnContext(ctx, "Attempted to apply inactive template", logger.Fields{
			"template_id":   templateID,
			"template_name": template.Name,
		})
		s.metrics.IncrementCounter("settings.apply_inactive_template_error", nil)
		return nil, fmt.Errorf("cannot apply inactive template")
	}

	// Validate template application
	validation, err := s.ValidateTemplateApplication(ctx, templateID, target)
	if err != nil {
		return nil, fmt.Errorf("template application validation failed: %w", err)
	}

	if !validation.IsValid {
		s.metrics.IncrementCounter("settings.template_application_validation_error", nil)
		return nil, fmt.Errorf("template application validation failed: %s", validation.ErrorMessage)
	}

	// Apply the template
	result, err := s.repo.ApplyTemplate(ctx, templateID, target, options, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to apply template", logger.Fields{
			"template_id":   templateID,
			"template_name": template.Name,
			"target_type":   target.Type,
			"target_entity": target.EntityID,
			"error":         err.Error(),
		})
		s.metrics.IncrementCounter("settings.apply_template_error", nil)
		return nil, fmt.Errorf("failed to apply template: %w", err)
	}

	s.logger.InfoContext(ctx, "Successfully applied template", logger.Fields{
		"template_id":   templateID,
		"template_name": template.Name,
		"target_type":   target.Type,
		"target_entity": target.EntityID,
		"applied":       result.Summary.Applied,
		"conflicts":     result.Summary.Conflicts,
		"errors":        result.Summary.Errors,
		"duration_ms":   result.Summary.DurationMS,
	})

	// Create audit event
	if err := s.createTemplateApplicationAuditEvent(ctx, audit.CreateAuditEventRequest{
		UserID:        &userID,
		EventType:     "TEMPLATE_APPLICATION",
		EventCategory: "SETTINGS",
		Severity:      "INFO",
		EntityID:      target.EntityID,
		Context:       s.buildTemplateApplicationAuditContext(template, target, result),
	}); err != nil {
		span.RecordError(err)
	}

	s.metrics.IncrementCounter("settings.apply_template_success", metrics.Fields{
		"target_type": target.Type,
		"applied":     fmt.Sprintf("%d", result.Summary.Applied),
		"conflicts":   fmt.Sprintf("%d", result.Summary.Conflicts),
		"errors":      fmt.Sprintf("%d", result.Summary.Errors),
	})

	return result, nil
}

// ValidateTemplateApplication validates a template can be applied to target
func (s *templateService) ValidateTemplateApplication(ctx context.Context, templateID domain.TemplateID, target repository.TemplateTarget) (*TemplateValidationResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.validate_template_application")
	defer span.End()

	result := &TemplateValidationResult{
		TemplateID:  templateID,
		Target:      target,
		IsValid:     true,
		Warnings:    make([]string, 0),
		Validations: make(map[string]bool),
		ValidatedAt: time.Now(),
	}

	// Get template
	template, err := s.repo.GetTemplate(ctx, templateID)
	if err != nil {
		result.IsValid = false
		result.ErrorMessage = fmt.Sprintf("failed to get template: %v", err)
		return result, nil
	}

	// Check if template is active
	if !template.IsActive {
		result.IsValid = false
		result.ErrorMessage = "template is not active"
		return result, nil
	}

	// Validate target compatibility
	if err := s.validateTargetCompatibility(ctx, template, target); err != nil {
		result.IsValid = false
		result.ErrorMessage = err.Error()
		return result, nil
	}

	// Validate each configuration in the template
	validationErrors := make([]string, 0)
	for _, config := range template.Configurations {
		if err := s.validateTemplateConfiguration(ctx, config, target); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("config %s.%s: %v", config.Module, config.ConfigKey, err))
		}
	}

	if len(validationErrors) > 0 {
		result.IsValid = false
		result.ErrorMessage = fmt.Sprintf("configuration validation failed: %v", validationErrors)
	}

	result.Validations["target_compatibility"] = true
	result.Validations["configuration_validity"] = len(validationErrors) == 0

	return result, nil
}

// GetTemplateApplicationHistory gets template application history
func (s *templateService) GetTemplateApplicationHistory(ctx context.Context, entityID *uuid.UUID, limit, offset int) ([]*repository.TemplateApplication, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.get_template_application_history")
	defer span.End()

	history, err := s.repo.GetTemplateApplicationHistory(ctx, entityID, limit, offset)
	if err != nil {
		s.metrics.IncrementCounter("settings.get_template_history_error", nil)
		return nil, fmt.Errorf("failed to get template application history: %w", err)
	}

	s.metrics.IncrementCounter("settings.get_template_history_success", metrics.Fields{
		"count": fmt.Sprintf("%d", len(history)),
	})

	return history, nil
}

// GetTemplateUsageStats gets usage statistics for a template
func (s *templateService) GetTemplateUsageStats(ctx context.Context, templateID domain.TemplateID, startTime, endTime time.Time) (*TemplateUsageStats, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.get_template_usage_stats")
	defer span.End()

	// For now, return placeholder stats since repository method isn't implemented
	stats := &TemplateUsageStats{
		TemplateID:        templateID,
		TotalApplications: 0,
		UniqueTargets:     0,
		SuccessfulRuns:    0,
		FailedRuns:        0,
		LastApplied:       nil,
		PeriodStart:       startTime,
		PeriodEnd:         endTime,
	}

	return stats, nil
}

// GetMostUsedTemplates gets the most frequently used templates
func (s *templateService) GetMostUsedTemplates(ctx context.Context, limit int, startTime, endTime time.Time) ([]*TemplateUsageStats, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.get_most_used_templates")
	defer span.End()

	// For now, return empty results since repository method isn't implemented
	return make([]*TemplateUsageStats, 0), nil
}

// Helper methods

// validateTemplate validates template structure and content
func (s *templateService) validateTemplate(ctx context.Context, template *domain.Template) error {
	if template.Name == "" {
		return fmt.Errorf("template name is required")
	}

	if len(template.Configurations) == 0 {
		return fmt.Errorf("template must contain at least one configuration")
	}

	// Validate each configuration
	for i, config := range template.Configurations {
		if config.Module == "" {
			return fmt.Errorf("configuration %d: module is required", i)
		}
		if config.ConfigKey == "" {
			return fmt.Errorf("configuration %d: key is required", i)
		}
		if config.Value.Raw == nil {
			return fmt.Errorf("configuration %d: value is required", i)
		}
	}

	return nil
}

// validateTargetCompatibility validates template can be applied to target
func (s *templateService) validateTargetCompatibility(ctx context.Context, template *domain.Template, target repository.TemplateTarget) error {
	// Check if template category compatibility
	if template.Category != "" && string(template.Category) != target.Type {
		return fmt.Errorf("template category %s is not compatible with target type %s", template.Category, target.Type)
	}

	return nil
}

// validateTemplateConfiguration validates a single template configuration
func (s *templateService) validateTemplateConfiguration(ctx context.Context, config domain.TemplateConfiguration, target repository.TemplateTarget) error {
	// Basic validation - could be extended with more complex rules
	if config.Value.Raw == nil {
		return fmt.Errorf("configuration value cannot be nil")
	}

	return nil
}

// createTemplateAuditEvent creates audit event for template operations
func (s *templateService) createTemplateAuditEvent(ctx context.Context, req audit.CreateAuditEventRequest) error {
	_, err := s.auditService.CreateAuditEvent(ctx, req)
	return err
}

// createTemplateApplicationAuditEvent creates audit event for template applications
func (s *templateService) createTemplateApplicationAuditEvent(ctx context.Context, req audit.CreateAuditEventRequest) error {
	_, err := s.auditService.CreateAuditEvent(ctx, req)
	return err
}

// buildTemplateAuditContext creates audit context for template operations
func (s *templateService) buildTemplateAuditContext(operation string, template *domain.Template, oldTemplate *domain.Template) json.RawMessage {
	context := map[string]any{
		"operation":           operation,
		"template_id":         template.ID,
		"template_name":       template.Name,
		"template_category":   string(template.Category),
		"configuration_count": len(template.Configurations),
		"is_active":           template.IsActive,
		"timestamp":           time.Now().UTC(),
	}

	if oldTemplate != nil {
		context["old_template_name"] = oldTemplate.Name
		context["old_is_active"] = oldTemplate.IsActive
		context["old_configuration_count"] = len(oldTemplate.Configurations)
	}

	contextBytes, _ := json.Marshal(context)
	return json.RawMessage(contextBytes)
}

// buildTemplateApplicationAuditContext creates audit context for template applications
func (s *templateService) buildTemplateApplicationAuditContext(template *domain.Template, target repository.TemplateTarget, result *domain.ApplicationResult) json.RawMessage {
	context := map[string]any{
		"template_id":      template.ID,
		"template_name":    template.Name,
		"target_type":      target.Type,
		"target_entity_id": target.EntityID,
		"total_configs":    result.Summary.TotalConfigs,
		"applied":          result.Summary.Applied,
		"conflicts":        result.Summary.Conflicts,
		"errors":           result.Summary.Errors,
		"duration_ms":      result.Summary.DurationMS,
		"timestamp":        time.Now().UTC(),
	}

	if len(result.Conflicts) > 0 {
		context["conflict_count"] = len(result.Conflicts)
		context["conflict_details"] = result.Conflicts
	}

	contextBytes, _ := json.Marshal(context)
	return json.RawMessage(contextBytes)
}

// Supporting types for template service

// TemplateValidationResult represents the result of template validation
type TemplateValidationResult struct {
	TemplateID   domain.TemplateID         `json:"template_id"`
	Target       repository.TemplateTarget `json:"target"`
	IsValid      bool                      `json:"is_valid"`
	ErrorMessage string                    `json:"error_message,omitempty"`
	Warnings     []string                  `json:"warnings"`
	Validations  map[string]bool           `json:"validations"`
	ValidatedAt  time.Time                 `json:"validated_at"`
}

// TemplateUsageStats represents usage statistics for a template
type TemplateUsageStats struct {
	TemplateID        domain.TemplateID `json:"template_id"`
	TemplateName      string            `json:"template_name"`
	TotalApplications int               `json:"total_applications"`
	UniqueTargets     int               `json:"unique_targets"`
	SuccessfulRuns    int               `json:"successful_runs"`
	FailedRuns        int               `json:"failed_runs"`
	LastApplied       *time.Time        `json:"last_applied"`
	PeriodStart       time.Time         `json:"period_start"`
	PeriodEnd         time.Time         `json:"period_end"`
	AvgExecutionTime  int64             `json:"avg_execution_time_ms"`
}
