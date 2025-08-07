package featureflag

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/audit"
)

// In-memory template storage (in production, this would be database-backed)
var templateStore = make(map[uuid.UUID]*FlagTemplate)

// CreateFlagTemplate creates a new feature flag template
func (s *adminServiceImpl) CreateFlagTemplate(ctx context.Context, request *CreateFlagTemplateRequest) (*FlagTemplate, error) {
	ctx, span := s.tracing.StartSpan(ctx, "admin.create_flag_template")
	defer span.End()

	// Validate template request
	if err := s.validateTemplateRequest(request); err != nil {
		return nil, fmt.Errorf("invalid template request: %w", err)
	}

	template := &FlagTemplate{
		ID:              uuid.New(),
		Name:            request.Name,
		Description:     request.Description,
		Category:        request.Category,
		FlagType:        request.FlagType,
		DefaultValue:    request.DefaultValue,
		RolloutStrategy: request.RolloutStrategy,
		TargetAudience:  request.TargetAudience,
		Metadata:        request.Metadata,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Store template (in production, this would be in database)
	templateStore[template.ID] = template

	s.logger.Info("Flag template created", map[string]interface{}{
		"template_id":   template.ID.String(),
		"template_name": template.Name,
		"category":      template.Category,
		"flag_type":     string(template.FlagType),
	})

	// Audit template creation
	s.auditTemplateOperation(ctx, "create_template", template, nil)

	return template, nil
}

// GetFlagTemplate retrieves a specific flag template
func (s *adminServiceImpl) GetFlagTemplate(ctx context.Context, templateID uuid.UUID) (*FlagTemplate, error) {
	ctx, span := s.tracing.StartSpan(ctx, "admin.get_flag_template")
	defer span.End()

	template, exists := templateStore[templateID]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", templateID.String())
	}

	s.logger.Debug("Flag template retrieved", map[string]interface{}{
		"template_id":   template.ID.String(),
		"template_name": template.Name,
	})

	return template, nil
}

// ListFlagTemplates returns a paginated list of flag templates
func (s *adminServiceImpl) ListFlagTemplates(ctx context.Context, request *ListTemplatesRequest) (*ListTemplatesResponse, error) {
	ctx, span := s.tracing.StartSpan(ctx, "admin.list_flag_templates")
	defer span.End()

	// Set defaults
	if request.Page < 1 {
		request.Page = 1
	}
	if request.PageSize < 1 {
		request.PageSize = 20
	}
	if request.PageSize > 100 {
		request.PageSize = 100
	}

	// Filter templates
	filteredTemplates := []*FlagTemplate{}
	for _, template := range templateStore {
		// Apply filters
		if request.Category != "" && template.Category != request.Category {
			continue
		}
		if request.FlagType != "" && string(template.FlagType) != request.FlagType {
			continue
		}
		filteredTemplates = append(filteredTemplates, template)
	}

	// Paginate
	total := int64(len(filteredTemplates))
	start := (request.Page - 1) * request.PageSize
	end := start + request.PageSize

	if start >= len(filteredTemplates) {
		filteredTemplates = []*FlagTemplate{}
	} else {
		if end > len(filteredTemplates) {
			end = len(filteredTemplates)
		}
		filteredTemplates = filteredTemplates[start:end]
	}

	response := &ListTemplatesResponse{
		Templates: filteredTemplates,
		Total:     total,
		Page:      request.Page,
		PageSize:  request.PageSize,
	}

	s.logger.Debug("Templates listed", map[string]interface{}{
		"total":           total,
		"returned":        len(filteredTemplates),
		"page":            request.Page,
		"category_filter": request.Category,
		"type_filter":     request.FlagType,
	})

	return response, nil
}

// ApplyTemplate creates a new feature flag from a template
func (s *adminServiceImpl) ApplyTemplate(ctx context.Context, request *ApplyTemplateRequest) (*FeatureFlag, error) {
	ctx, span := s.tracing.StartSpan(ctx, "admin.apply_template")
	defer span.End()

	// Get the template
	template, err := s.GetFlagTemplate(ctx, request.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	// Create flag request from template
	// Handle type assertion for DefaultValue (CreateFeatureFlagRequest expects bool)
	var defaultValue bool
	if template.DefaultValue != nil {
		if boolVal, ok := template.DefaultValue.(bool); ok {
			defaultValue = boolVal
		}
		// For now, we only support boolean flags in CreateFeatureFlagRequest
	}

	flagRequest := &CreateFeatureFlagRequest{
		Name:         request.FlagName,
		Description:  template.Description,
		FlagType:     template.FlagType,
		DefaultValue: defaultValue,
	}

	// Apply template values
	if template.RolloutStrategy != nil {
		if percentage, ok := template.RolloutStrategy.Configuration["percentage"].(float64); ok {
			rolloutPercentage := int32(percentage)
			flagRequest.RolloutPercentage = &rolloutPercentage
		}
	}

	if template.TargetAudience != nil {
		flagRequest.TargetAudience = template.TargetAudience
	}

	// Start with template metadata
	metadata := make(map[string]interface{})
	if template.Metadata != nil {
		for k, v := range template.Metadata {
			metadata[k] = v
		}
	}

	// Add template source info
	metadata["template_id"] = template.ID.String()
	metadata["template_name"] = template.Name
	metadata["applied_at"] = time.Now().Format(time.RFC3339)

	// Apply overrides if provided
	if request.Overrides != nil {
		for key, value := range request.Overrides {
			switch key {
			case "description":
				if desc, ok := value.(string); ok {
					flagRequest.Description = desc
				}
			case "default_value":
				if boolVal, ok := value.(bool); ok {
					flagRequest.DefaultValue = boolVal
				}
			case "rollout_percentage":
				if pct, ok := value.(float64); ok {
					rolloutPercentage := int32(pct)
					flagRequest.RolloutPercentage = &rolloutPercentage
				}
			default:
				// Add to metadata
				metadata[key] = value
			}
		}
	}

	flagRequest.Metadata = metadata

	// Create the flag using the base service
	flag, err := s.baseService.CreateFeatureFlag(ctx, flagRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create flag from template: %w", err)
	}

	s.logger.Info("Template applied successfully", map[string]interface{}{
		"template_id":   template.ID.String(),
		"template_name": template.Name,
		"flag_id":       flag.ID.String(),
		"flag_name":     flag.Name,
		"overrides":     request.Overrides,
	})

	// Audit template application
	s.auditTemplateOperation(ctx, "apply_template", template, map[string]interface{}{
		"flag_id":   flag.ID.String(),
		"flag_name": flag.Name,
		"overrides": request.Overrides,
	})

	return flag, nil
}

// Validation and helper methods

func (s *adminServiceImpl) validateTemplateRequest(request *CreateFlagTemplateRequest) error {
	if request.Name == "" {
		return fmt.Errorf("template name is required")
	}

	if request.Description == "" {
		return fmt.Errorf("template description is required")
	}

	if request.Category == "" {
		return fmt.Errorf("template category is required")
	}

	// Validate flag type
	switch request.FlagType {
	case FlagTypeBoolean, FlagTypeString, FlagTypeNumber, FlagTypeJSON:
		// Valid types
	default:
		return fmt.Errorf("invalid flag type: %s", request.FlagType)
	}

	// Validate default value matches flag type
	if err := s.validateDefaultValueForType(request.DefaultValue, request.FlagType); err != nil {
		return fmt.Errorf("default value validation failed: %w", err)
	}

	// Validate rollout strategy if provided
	if request.RolloutStrategy != nil {
		if err := s.validateRolloutStrategyForTemplate(request.RolloutStrategy); err != nil {
			return fmt.Errorf("rollout strategy validation failed: %w", err)
		}
	}

	return nil
}

func (s *adminServiceImpl) validateDefaultValueForType(value interface{}, flagType FlagType) error {
	switch flagType {
	case FlagTypeBoolean:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("boolean flag type requires boolean default value")
		}
	case FlagTypeString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("string flag type requires string default value")
		}
	case FlagTypeNumber:
		switch value.(type) {
		case int, int32, int64, float32, float64:
			// Valid number types
		default:
			return fmt.Errorf("number flag type requires numeric default value")
		}
	case FlagTypeJSON:
		// For JSON, we accept any structure that can be marshaled
		if _, err := json.Marshal(value); err != nil {
			return fmt.Errorf("JSON flag type requires valid JSON structure: %w", err)
		}
	}
	return nil
}

func (s *adminServiceImpl) validateRolloutStrategyForTemplate(strategy *RolloutStrategy) error {
	if strategy.Name == "" {
		return fmt.Errorf("rollout strategy name is required")
	}

	if strategy.Configuration == nil {
		return fmt.Errorf("rollout strategy configuration is required")
	}

	switch strategy.Type {
	case RolloutStrategyPercentage:
		if percentage, ok := strategy.Configuration["percentage"]; ok {
			if pct, ok := percentage.(float64); !ok || pct < 0 || pct > 100 {
				return fmt.Errorf("percentage must be between 0 and 100")
			}
		} else {
			return fmt.Errorf("percentage rollout strategy requires 'percentage' configuration")
		}
	case RolloutStrategyUserAttribute:
		if _, ok := strategy.Configuration["attribute"]; !ok {
			return fmt.Errorf("user attribute strategy requires 'attribute' configuration")
		}
		if _, ok := strategy.Configuration["value"]; !ok {
			return fmt.Errorf("user attribute strategy requires 'value' configuration")
		}
	}

	return nil
}

// Predefined template creation helpers

// CreatePredefinedTemplates creates a set of commonly used flag templates
func (s *adminServiceImpl) CreatePredefinedTemplates(ctx context.Context) error {
	predefinedTemplates := []CreateFlagTemplateRequest{
		{
			Name:         "Feature Toggle",
			Description:  "Simple boolean feature toggle for enabling/disabling features",
			Category:     "feature_toggle",
			FlagType:     FlagTypeBoolean,
			DefaultValue: false,
			Metadata: map[string]interface{}{
				"use_case":   "feature_enablement",
				"complexity": "low",
				"risk_level": "low",
			},
		},
		{
			Name:         "A/B Test Flag",
			Description:  "Boolean flag for A/B testing with 50% rollout",
			Category:     "experimentation",
			FlagType:     FlagTypeBoolean,
			DefaultValue: false,
			RolloutStrategy: &RolloutStrategy{
				ID:   uuid.New(),
				Name: "50% A/B Test",
				Type: RolloutStrategyPercentage,
				Configuration: map[string]interface{}{
					"percentage": 50.0,
				},
			},
			Metadata: map[string]interface{}{
				"use_case":   "ab_testing",
				"complexity": "medium",
				"risk_level": "medium",
			},
		},
		{
			Name:         "Gradual Rollout",
			Description:  "Feature flag with gradual rollout strategy",
			Category:     "rollout",
			FlagType:     FlagTypeBoolean,
			DefaultValue: false,
			RolloutStrategy: &RolloutStrategy{
				ID:   uuid.New(),
				Name: "Gradual 5% to 100%",
				Type: RolloutStrategyGradual,
				Configuration: map[string]interface{}{
					"initial_percentage": 5.0,
					"final_percentage":   100.0,
					"duration_hours":     168.0, // 1 week
				},
			},
			Metadata: map[string]interface{}{
				"use_case":   "safe_rollout",
				"complexity": "high",
				"risk_level": "medium",
			},
		},
		{
			Name:         "Configuration Value",
			Description:  "String-based configuration flag for dynamic values",
			Category:     "configuration",
			FlagType:     FlagTypeString,
			DefaultValue: "default_config",
			Metadata: map[string]interface{}{
				"use_case":   "configuration",
				"complexity": "low",
				"risk_level": "low",
			},
		},
		{
			Name:         "Numeric Parameter",
			Description:  "Numeric flag for adjustable parameters",
			Category:     "parameter",
			FlagType:     FlagTypeNumber,
			DefaultValue: 10.0,
			Metadata: map[string]interface{}{
				"use_case":   "parameter_tuning",
				"complexity": "low",
				"risk_level": "low",
			},
		},
	}

	successCount := 0
	for _, templateReq := range predefinedTemplates {
		_, err := s.CreateFlagTemplate(ctx, &templateReq)
		if err != nil {
			s.logger.Warn("Failed to create predefined template", map[string]interface{}{
				"template_name": templateReq.Name,
				"error":         err.Error(),
			})
		} else {
			successCount++
		}
	}

	s.logger.Info("Predefined templates created", map[string]interface{}{
		"total_templates":      len(predefinedTemplates),
		"successfully_created": successCount,
	})

	return nil
}

// GetTemplatesByCategory returns templates filtered by category
func (s *adminServiceImpl) GetTemplatesByCategory(ctx context.Context, category string) ([]*FlagTemplate, error) {
	request := &ListTemplatesRequest{
		Page:     1,
		PageSize: 100,
		Category: category,
	}

	response, err := s.ListFlagTemplates(ctx, request)
	if err != nil {
		return nil, err
	}

	return response.Templates, nil
}

// GetRecommendedTemplate suggests a template based on use case
func (s *adminServiceImpl) GetRecommendedTemplate(ctx context.Context, useCase string) (*FlagTemplate, error) {
	var recommendedCategory string

	switch useCase {
	case "feature_release", "feature_enablement":
		recommendedCategory = "feature_toggle"
	case "ab_test", "experiment":
		recommendedCategory = "experimentation"
	case "safe_deploy", "gradual_release":
		recommendedCategory = "rollout"
	case "config", "parameter":
		recommendedCategory = "configuration"
	default:
		recommendedCategory = "feature_toggle" // Default recommendation
	}

	templates, err := s.GetTemplatesByCategory(ctx, recommendedCategory)
	if err != nil {
		return nil, err
	}

	if len(templates) == 0 {
		return nil, fmt.Errorf("no templates found for category: %s", recommendedCategory)
	}

	// Return the first template in the category as recommendation
	return templates[0], nil
}

// Audit helper for template operations
func (s *adminServiceImpl) auditTemplateOperation(ctx context.Context, operation string, template *FlagTemplate, additionalData map[string]interface{}) {
	contextData := map[string]interface{}{
		"operation":     operation,
		"template_id":   template.ID.String(),
		"template_name": template.Name,
		"category":      template.Category,
		"flag_type":     string(template.FlagType),
	}

	// Add additional data if provided
	if additionalData != nil {
		for k, v := range additionalData {
			contextData[k] = v
		}
	}

	contextJSON, _ := json.Marshal(contextData)

	severity := AuditSeverityInfo
	if operation == "apply_template" {
		severity = AuditSeverityWarn // Template applications create new flags
	}

	s.auditService.Record(ctx, audit.AuditEvent{
		EventType:     "admin_template_operation",
		EventCategory: AuditCategoryAdmin,
		Severity:      severity,
		EntityID:      uuid.NullUUID{UUID: template.ID, Valid: true},
		Decision:      fmt.Sprintf("EXECUTED_%s", operation),
		Reason:        fmt.Sprintf("Template operation: %s on template %s", operation, template.Name),
		Context:       contextJSON,
	})
}
