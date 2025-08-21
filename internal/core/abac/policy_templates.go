package abac

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/core/abac/repository"
	"github.com/niiniyare/erp/internal/shared/convert"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// Policy Templates and Import/Export types and methods

// Policy Templates

type PolicyTemplate struct {
	ID          uuid.UUID             `json:"id"`
	Name        string                `json:"name"`
	DisplayName *string               `json:"display_name,omitempty"`
	Description string                `json:"description"`
	Category    types.PolicyCategory  `json:"category"`
	Tags        []string              `json:"tags,omitempty"`
	Template    PolicyTemplateContent `json:"template"`
	Parameters  []TemplateParameter   `json:"parameters"`
	Examples    []TemplateExample     `json:"examples,omitempty"`
	Metadata    TemplateMetadata      `json:"metadata"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	CreatedBy   *uuid.UUID            `json:"created_by,omitempty"`
}

type PolicyTemplateContent struct {
	PolicyType  types.PolicyType   `json:"policy_type"`
	Effect      types.PolicyEffect `json:"effect"`
	Priority    int32              `json:"priority"`
	Target      map[string]any     `json:"target"`
	Rule        map[string]any     `json:"rule"`
	Obligations map[string]any     `json:"obligations,omitempty"`
	Advice      map[string]any     `json:"advice,omitempty"`
}

type TemplateParameter struct {
	Name            string         `json:"name"`
	DisplayName     string         `json:"display_name"`
	Description     string         `json:"description"`
	Type            string         `json:"type"` // "string", "number", "boolean", "array", "object"
	Required        bool           `json:"required"`
	DefaultValue    any            `json:"default_value,omitempty"`
	AllowedValues   []any          `json:"allowed_values,omitempty"`
	ValidationRules map[string]any `json:"validation_rules,omitempty"`
	Placeholder     string         `json:"placeholder,omitempty"`
}

type TemplateExample struct {
	Name           string                `json:"name"`
	Description    string                `json:"description"`
	Parameters     map[string]any        `json:"parameters"`
	ExpectedPolicy PolicyTemplateContent `json:"expected_policy"`
}

type TemplateMetadata struct {
	Version              string     `json:"version"`
	Author               string     `json:"author,omitempty"`
	UsageCount           int64      `json:"usage_count"`
	LastUsed             *time.Time `json:"last_used,omitempty"`
	Rating               float64    `json:"rating"`
	Complexity           string     `json:"complexity"` // "simple", "medium", "advanced"
	RequiredSkills       []string   `json:"required_skills,omitempty"`
	Industry             []string   `json:"industry,omitempty"`
	ComplianceFrameworks []string   `json:"compliance_frameworks,omitempty"`
}

type CreatePolicyTemplateRequest struct {
	Name        string                `json:"name" validate:"required"`
	DisplayName *string               `json:"display_name,omitempty"`
	Description string                `json:"description" validate:"required"`
	Category    types.PolicyCategory  `json:"category" validate:"required"`
	Tags        []string              `json:"tags,omitempty"`
	Template    PolicyTemplateContent `json:"template" validate:"required"`
	Parameters  []TemplateParameter   `json:"parameters,omitempty"`
	Examples    []TemplateExample     `json:"examples,omitempty"`
	CreatedBy   *uuid.UUID            `json:"created_by,omitempty"`
}

func (pm *policyManager) CreatePolicyTemplate(ctx context.Context, req *CreatePolicyTemplateRequest) (*PolicyTemplate, error) {
	ctx, span := pm.tracer.StartSpan(ctx, "abac.policy_manager.CreatePolicyTemplate",
		tracing.WithAttributes(
			attribute.String("template_name", req.Name),
			attribute.String("category", string(req.Category)),
		))
	defer span.End()

	pm.logger.InfoContext(ctx, "Creating policy template",
		logger.Fields{
			"name":       req.Name,
			"category":   req.Category,
			"parameters": len(req.Parameters),
			"examples":   len(req.Examples),
		})

	// Validate template structure
	if err := pm.validatePolicyTemplate(ctx, req); err != nil {
		pm.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, errors.NewBusinessErrorWithContext(ctx, "TEMPLATE_VALIDATION_FAILED", "Policy template validation failed")
	}

	// Create template
	template := &PolicyTemplate{
		ID:          uuid.New(),
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Category:    req.Category,
		Tags:        req.Tags,
		Template:    req.Template,
		Parameters:  req.Parameters,
		Examples:    req.Examples,
		Metadata: TemplateMetadata{
			Version:    "1.0.0",
			UsageCount: 0,
			Rating:     0.0,
			Complexity: pm.calculateTemplateComplexity(req),
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: req.CreatedBy,
	}

	// In a real implementation, you would save this to a template repository
	// For now, we'll just return the template

	pm.metrics.IncrementCounter("policy_manager_template_created", nil)

	pm.logger.InfoContext(ctx, "Policy template created successfully",
		logger.Fields{
			"template_id": template.ID,
			"name":        template.Name,
			"complexity":  template.Metadata.Complexity,
		})

	return template, nil
}

// Policy Import/Export

type ImportPoliciesRequest struct {
	Format    string        `json:"format"` // "json", "yaml", "xml"
	Data      []byte        `json:"data"`
	Options   ImportOptions `json:"options"`
	CreatedBy *uuid.UUID    `json:"created_by,omitempty"`
	DryRun    bool          `json:"dry_run"`
}

type ImportOptions struct {
	OverwriteExisting  bool                  `json:"overwrite_existing"`
	SkipInvalid        bool                  `json:"skip_invalid"`
	ValidateOnly       bool                  `json:"validate_only"`
	DefaultEntity      *uuid.UUID            `json:"default_entity,omitempty"`
	TagsToAdd          []string              `json:"tags_to_add,omitempty"`
	CategoryOverride   *types.PolicyCategory `json:"category_override,omitempty"`
	PriorityAdjustment int32                 `json:"priority_adjustment"`
	MakeDraft          bool                  `json:"make_draft"` // Import as inactive policies
}

type ImportResult struct {
	Summary           ImportSummary            `json:"summary"`
	ImportedPolicies  []ImportedPolicy         `json:"imported_policies"`
	FailedImports     []FailedImport           `json:"failed_imports"`
	Warnings          []ImportWarning          `json:"warnings"`
	ValidationResults []PolicyValidationResult `json:"validation_results"`
	ExecutionTime     time.Duration            `json:"execution_time"`
	Timestamp         time.Time                `json:"timestamp"`
}

type ImportSummary struct {
	TotalPolicies     int32 `json:"total_policies"`
	SuccessfulImports int32 `json:"successful_imports"`
	FailedImports     int32 `json:"failed_imports"`
	SkippedPolicies   int32 `json:"skipped_policies"`
	ValidatedPolicies int32 `json:"validated_policies"`
	WarningsCount     int32 `json:"warnings_count"`
}

type ImportedPolicy struct {
	OriginalName   string         `json:"original_name"`
	ImportedPolicy *models.Policy `json:"imported_policy"`
	Changes        []string       `json:"changes,omitempty"`
	Warnings       []string       `json:"warnings,omitempty"`
}

type FailedImport struct {
	OriginalName string         `json:"original_name"`
	Reason       string         `json:"reason"`
	Errors       []string       `json:"errors"`
	PolicyData   map[string]any `json:"policy_data,omitempty"`
}

type ImportWarning struct {
	PolicyName  string `json:"policy_name"`
	WarningType string `json:"warning_type"`
	Message     string `json:"message"`
	Severity    string `json:"severity"`
}

type ExportPoliciesRequest struct {
	PolicyIDs   []uuid.UUID         `json:"policy_ids,omitempty"`
	Filters     PolicyExportFilters `json:"filters,omitempty"`
	Format      string              `json:"format"` // "json", "yaml", "xml"
	Options     ExportOptions       `json:"options"`
	RequestedBy *uuid.UUID          `json:"requested_by,omitempty"`
}

type PolicyExportFilters struct {
	Categories    []types.PolicyCategory `json:"categories,omitempty"`
	EntityID      *uuid.UUID             `json:"entity_id,omitempty"`
	IsActive      *bool                  `json:"is_active,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
	CreatedAfter  *time.Time             `json:"created_after,omitempty"`
	CreatedBefore *time.Time             `json:"created_before,omitempty"`
	SearchTerms   []string               `json:"search_terms,omitempty"`
}

type ExportOptions struct {
	IncludeMetadata       bool     `json:"include_metadata"`
	IncludeUsageStats     bool     `json:"include_usage_stats"`
	IncludeVersionHistory bool     `json:"include_version_history"`
	ExcludeFields         []string `json:"exclude_fields,omitempty"`
	MinifyOutput          bool     `json:"minify_output"`
	IncludeSchema         bool     `json:"include_schema"`
	CompressOutput        bool     `json:"compress_output"`
	SplitByCategory       bool     `json:"split_by_category"`
}

type ExportResult struct {
	Summary       ExportSummary     `json:"summary"`
	Data          []byte            `json:"data,omitempty"`
	Files         map[string][]byte `json:"files,omitempty"` // For split exports
	Metadata      ExportMetadata    `json:"metadata"`
	ExecutionTime time.Duration     `json:"execution_time"`
	Timestamp     time.Time         `json:"timestamp"`
}

type ExportSummary struct {
	TotalPolicies    int32   `json:"total_policies"`
	ExportedPolicies int32   `json:"exported_policies"`
	SkippedPolicies  int32   `json:"skipped_policies"`
	OutputSize       int64   `json:"output_size_bytes"`
	CompressionRatio float64 `json:"compression_ratio,omitempty"`
}

type ExportMetadata struct {
	Format        string     `json:"format"`
	SchemaVersion string     `json:"schema_version"`
	ExportedAt    time.Time  `json:"exported_at"`
	ExportedBy    *uuid.UUID `json:"exported_by,omitempty"`
	TenantID      uuid.UUID  `json:"tenant_id"`
	SourceSystem  string     `json:"source_system"`
	Checksum      string     `json:"checksum"`
}

func (pm *policyManager) ImportPolicies(ctx context.Context, req *ImportPoliciesRequest) (*ImportResult, error) {
	ctx, span := pm.tracer.StartSpan(ctx, "abac.policy_manager.ImportPolicies",
		tracing.WithAttributes(
			attribute.String("format", req.Format),
			attribute.Bool("dry_run", req.DryRun),
		))
	defer span.End()

	startTime := time.Now()

	pm.logger.InfoContext(ctx, "Starting policy import",
		logger.Fields{
			"format":    req.Format,
			"dry_run":   req.DryRun,
			"data_size": len(req.Data),
		})

	// Parse import data
	parsedPolicies, err := pm.parseImportData(ctx, req.Format, req.Data)
	if err != nil {
		pm.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, errors.NewBusinessErrorWithContext(ctx, "IMPORT_PARSE_FAILED", "Failed to parse import data")
	}

	var importedPolicies []ImportedPolicy
	var failedImports []FailedImport
	var warnings []ImportWarning
	var validationResults []PolicyValidationResult

	// Process each policy
	for i, policyData := range parsedPolicies {
		pm.logger.DebugContext(ctx, "Processing import policy",
			logger.Fields{
				"index":       i,
				"policy_name": policyData.Name,
			})

		// Validate policy
		validation := pm.validateImportedPolicy(ctx, policyData, req.Options)
		validationResults = append(validationResults, validation)

		if !validation.IsValid && !req.Options.SkipInvalid {
			failedImports = append(failedImports, FailedImport{
				OriginalName: policyData.Name,
				Reason:       "Validation failed",
				Errors:       validation.Errors,
			})
			continue
		}

		// Apply import transformations
		transformedPolicy, policyWarnings := pm.transformImportedPolicy(ctx, policyData, req.Options)
		if len(policyWarnings) > 0 {
			for _, warning := range policyWarnings {
				warnings = append(warnings, ImportWarning{
					PolicyName:  policyData.Name,
					WarningType: "transformation",
					Message:     warning,
					Severity:    "medium",
				})
			}
		}

		// Skip validation-only mode
		if req.Options.ValidateOnly || req.DryRun {
			importedPolicies = append(importedPolicies, ImportedPolicy{
				OriginalName:   policyData.Name,
				ImportedPolicy: nil, // Not actually created in dry run
				Warnings:       policyWarnings,
			})
			continue
		}

		// Check for existing policy with same name
		existingPolicy, _ := pm.policyRepo.GetPolicyByName(ctx, transformedPolicy.Name)
		if existingPolicy != nil && !req.Options.OverwriteExisting {
			warnings = append(warnings, ImportWarning{
				PolicyName:  policyData.Name,
				WarningType: "duplicate",
				Message:     fmt.Sprintf("Policy with name '%s' already exists and overwrite is disabled", transformedPolicy.Name),
				Severity:    "high",
			})
			continue
		}

		// Create or update policy
		var resultPolicy *models.Policy
		if existingPolicy != nil && req.Options.OverwriteExisting {
			// Update existing policy
			updateReq := pm.convertToUpdateRequest(transformedPolicy, existingPolicy.ID)
			resultPolicy, err = pm.policyRepo.UpdatePolicy(ctx, existingPolicy.ID, updateReq)
		} else {
			// Create new policy
			createReq := pm.convertToCreateRequest(transformedPolicy, req.CreatedBy)
			resultPolicy, err = pm.policyRepo.CreatePolicy(ctx, createReq)
		}

		if err != nil {
			failedImports = append(failedImports, FailedImport{
				OriginalName: policyData.Name,
				Reason:       "Creation/Update failed",
				Errors:       []string{err.Error()},
			})
			continue
		}

		importedPolicies = append(importedPolicies, ImportedPolicy{
			OriginalName:   policyData.Name,
			ImportedPolicy: resultPolicy,
			Warnings:       policyWarnings,
		})
	}

	executionTime := time.Since(startTime)

	// Calculate summary
	totalPolicies, err := convert.IntToInt32(len(parsedPolicies))
	if err != nil {
		return nil, fmt.Errorf("invalid total policies count: %w", err)
	}
	successfulImports, err := convert.IntToInt32(len(importedPolicies))
	if err != nil {
		return nil, fmt.Errorf("invalid successful imports count: %w", err)
	}
	failedImportsCount, err := convert.IntToInt32(len(failedImports))
	if err != nil {
		return nil, fmt.Errorf("invalid failed imports count: %w", err)
	}
	validatedPolicies, err := convert.IntToInt32(len(validationResults))
	if err != nil {
		return nil, fmt.Errorf("invalid validated policies count: %w", err)
	}
	warningsCount, err := convert.IntToInt32(len(warnings))
	if err != nil {
		return nil, fmt.Errorf("invalid warnings count: %w", err)
	}

	summary := ImportSummary{
		TotalPolicies:     totalPolicies,
		SuccessfulImports: successfulImports,
		FailedImports:     failedImportsCount,
		ValidatedPolicies: validatedPolicies,
		WarningsCount:     warningsCount,
	}

	// Record metrics
	pm.metrics.IncrementCounter("policy_manager_import", nil)
	pm.metrics.SetGauge("policy_import_success_rate",
		float64(summary.SuccessfulImports)/float64(summary.TotalPolicies)*100,
		metrics.Fields{"format": req.Format})
	pm.metrics.ObserveHistogram("policy_import_duration_seconds", executionTime.Seconds(),
		metrics.Fields{"format": req.Format, "count": fmt.Sprintf("%d", summary.TotalPolicies)})

	result := &ImportResult{
		Summary:           summary,
		ImportedPolicies:  importedPolicies,
		FailedImports:     failedImports,
		Warnings:          warnings,
		ValidationResults: validationResults,
		ExecutionTime:     executionTime,
		Timestamp:         time.Now(),
	}

	pm.logger.InfoContext(ctx, "Policy import completed",
		logger.Fields{
			"total_policies":     summary.TotalPolicies,
			"successful_imports": summary.SuccessfulImports,
			"failed_imports":     summary.FailedImports,
			"warnings":           summary.WarningsCount,
			"execution_time":     executionTime.Milliseconds(),
		})

	return result, nil
}

func (pm *policyManager) ExportPolicies(ctx context.Context, req *ExportPoliciesRequest) (*ExportResult, error) {
	ctx, span := pm.tracer.StartSpan(ctx, "abac.policy_manager.ExportPolicies",
		tracing.WithAttributes(
			attribute.String("format", req.Format),
			attribute.Int("policy_ids", len(req.PolicyIDs)),
		))
	defer span.End()

	startTime := time.Now()

	pm.logger.InfoContext(ctx, "Starting policy export",
		logger.Fields{
			"format":     req.Format,
			"policy_ids": len(req.PolicyIDs),
			"filters":    req.Filters,
		})

	// Get policies to export
	var policiesToExport []*models.Policy
	var err error

	if len(req.PolicyIDs) > 0 {
		// Export specific policies
		policiesToExport, err = pm.policyRepo.GetPoliciesByIDs(ctx, req.PolicyIDs)
	} else {
		// Export based on filters
		policiesToExport, err = pm.getPoliciesByFilters(ctx, req.Filters)
	}

	if err != nil {
		pm.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, errors.NewBusinessErrorWithContext(ctx, "EXPORT_POLICIES_FAILED", "Failed to get policies for export")
	}

	// Transform policies for export
	exportData := pm.transformPoliciesForExport(ctx, policiesToExport, req.Options)

	// Generate output data
	var outputData []byte
	var outputFiles map[string][]byte

	if req.Options.SplitByCategory {
		outputFiles = pm.generateSplitExport(ctx, exportData, req.Format, req.Options)
	} else {
		outputData, err = pm.generateExportData(ctx, exportData, req.Format, req.Options)
		if err != nil {
			pm.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
			return nil, errors.NewBusinessErrorWithContext(ctx, "EXPORT_GENERATION_FAILED", "Failed to generate export data")
		}
	}

	executionTime := time.Since(startTime)

	// Calculate summary
	outputSize := int64(len(outputData))
	if outputFiles != nil {
		for _, fileData := range outputFiles {
			outputSize += int64(len(fileData))
		}
	}

	totalPolicies, err := convert.IntToInt32(len(policiesToExport))
	if err != nil {
		return nil, fmt.Errorf("invalid total policies count: %w", err)
	}
	exportedPoliciesCount, err := convert.IntToInt32(len(policiesToExport))
	if err != nil {
		return nil, fmt.Errorf("invalid exported policies count: %w", err)
	}

	summary := ExportSummary{
		TotalPolicies:    totalPolicies,
		ExportedPolicies: exportedPoliciesCount,
		OutputSize:       outputSize,
	}

	// Generate metadata
	metadata := ExportMetadata{
		Format:        req.Format,
		SchemaVersion: "1.0.0",
		ExportedAt:    time.Now(),
		ExportedBy:    req.RequestedBy,
		SourceSystem:  "ABAC Policy Manager",
		Checksum:      pm.calculateChecksum(outputData),
	}

	// Record metrics
	pm.metrics.IncrementCounter("policy_manager_export", nil)
	pm.metrics.SetGauge("policy_export_size_bytes", float64(outputSize),
		metrics.Fields{"format": req.Format})
	pm.metrics.ObserveHistogram("policy_export_duration_seconds", executionTime.Seconds(),
		metrics.Fields{"format": req.Format, "count": fmt.Sprintf("%d", summary.TotalPolicies)})

	result := &ExportResult{
		Summary:       summary,
		Data:          outputData,
		Files:         outputFiles,
		Metadata:      metadata,
		ExecutionTime: executionTime,
		Timestamp:     time.Now(),
	}

	pm.logger.InfoContext(ctx, "Policy export completed",
		logger.Fields{
			"total_policies":    summary.TotalPolicies,
			"exported_policies": summary.ExportedPolicies,
			"output_size":       outputSize,
			"execution_time":    executionTime.Milliseconds(),
		})

	return result, nil
}

// Helper methods for templates and import/export

func (pm *policyManager) validatePolicyTemplate(ctx context.Context, req *CreatePolicyTemplateRequest) error {
	// Validate template structure
	if req.Template.PolicyType == "" {
		return fmt.Errorf("template policy_type is required")
	}

	if req.Template.Effect == "" {
		return fmt.Errorf("template effect is required")
	}

	if req.Template.Target == nil || len(req.Template.Target) == 0 {
		return fmt.Errorf("template target is required")
	}

	if req.Template.Rule == nil || len(req.Template.Rule) == 0 {
		return fmt.Errorf("template rule is required")
	}

	// Validate parameters
	for _, param := range req.Parameters {
		if param.Name == "" {
			return fmt.Errorf("parameter name is required")
		}
		if param.Type == "" {
			return fmt.Errorf("parameter type is required")
		}
	}

	return nil
}

func (pm *policyManager) calculateTemplateComplexity(req *CreatePolicyTemplateRequest) string {
	score := 0

	// Count parameters
	score += len(req.Parameters) * 2

	// Check rule complexity
	if ruleStr, err := json.Marshal(req.Template.Rule); err == nil {
		if strings.Contains(string(ruleStr), "AND") || strings.Contains(string(ruleStr), "OR") {
			score += 5
		}
		if strings.Contains(string(ruleStr), "NOT") {
			score += 3
		}
	}

	// Check obligations and advice
	if req.Template.Obligations != nil && len(req.Template.Obligations) > 0 {
		score += 3
	}
	if req.Template.Advice != nil && len(req.Template.Advice) > 0 {
		score += 2
	}

	if score < 5 {
		return "simple"
	} else if score < 15 {
		return "medium"
	} else {
		return "advanced"
	}
}

type ImportPolicyData struct {
	Name        string               `json:"name"`
	DisplayName *string              `json:"display_name,omitempty"`
	Description *string              `json:"description,omitempty"`
	PolicyType  types.PolicyType     `json:"policy_type"`
	Effect      types.PolicyEffect   `json:"effect"`
	Priority    int32                `json:"priority"`
	Category    types.PolicyCategory `json:"category"`
	Target      map[string]any       `json:"target"`
	Rule        map[string]any       `json:"rule"`
	Obligations map[string]any       `json:"obligations,omitempty"`
	Advice      map[string]any       `json:"advice,omitempty"`
	IsActive    bool                 `json:"is_active"`
	Tags        []string             `json:"tags,omitempty"`
}

func (pm *policyManager) parseImportData(ctx context.Context, format string, data []byte) ([]ImportPolicyData, error) {
	var policies []ImportPolicyData

	switch strings.ToLower(format) {
	case "json":
		// Try to parse as array first
		var policyArray []ImportPolicyData
		if err := json.Unmarshal(data, &policyArray); err == nil {
			policies = policyArray
		} else {
			// Try to parse as single policy
			var singlePolicy ImportPolicyData
			if err := json.Unmarshal(data, &singlePolicy); err != nil {
				return nil, fmt.Errorf("failed to parse JSON data: %w", err)
			}
			policies = []ImportPolicyData{singlePolicy}
		}
	case "yaml":
		// YAML parsing would go here
		return nil, fmt.Errorf("YAML format not yet implemented")
	case "xml":
		// XML parsing would go here
		return nil, fmt.Errorf("XML format not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	return policies, nil
}

func (pm *policyManager) validateImportedPolicy(ctx context.Context, policyData ImportPolicyData, options ImportOptions) PolicyValidationResult {
	result := PolicyValidationResult{IsValid: true}

	// Basic validation
	if policyData.Name == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "Policy name is required")
	}

	if policyData.PolicyType == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "Policy type is required")
	}

	if policyData.Effect == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "Policy effect is required")
	}

	if policyData.Target == nil || len(policyData.Target) == 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Policy target is required")
	}

	if policyData.Rule == nil || len(policyData.Rule) == 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Policy rule is required")
	}

	// Warnings for best practices
	if policyData.Priority < 1 || policyData.Priority > 1000 {
		result.Warnings = append(result.Warnings, "Priority should be between 1 and 1000")
	}

	return result
}

func (pm *policyManager) transformImportedPolicy(ctx context.Context, policyData ImportPolicyData, options ImportOptions) (ImportPolicyData, []string) {
	var warnings []string
	transformed := policyData

	// Apply category override
	if options.CategoryOverride != nil {
		if transformed.Category != *options.CategoryOverride {
			warnings = append(warnings, fmt.Sprintf("Category changed from %s to %s", transformed.Category, *options.CategoryOverride))
			transformed.Category = *options.CategoryOverride
		}
	}

	// Apply priority adjustment
	if options.PriorityAdjustment != 0 {
		newPriority := transformed.Priority + options.PriorityAdjustment
		if newPriority < 1 {
			newPriority = 1
		} else if newPriority > 1000 {
			newPriority = 1000
		}
		if newPriority != transformed.Priority {
			warnings = append(warnings, fmt.Sprintf("Priority adjusted from %d to %d", transformed.Priority, newPriority))
			transformed.Priority = newPriority
		}
	}

	// Add tags
	if len(options.TagsToAdd) > 0 {
		transformed.Tags = append(transformed.Tags, options.TagsToAdd...)
		warnings = append(warnings, fmt.Sprintf("Added tags: %v", options.TagsToAdd))
	}

	// Make draft if requested
	if options.MakeDraft {
		if transformed.IsActive {
			warnings = append(warnings, "Policy imported as inactive (draft mode)")
			transformed.IsActive = false
		}
	}

	return transformed, warnings
}

func (pm *policyManager) convertToCreateRequest(policyData ImportPolicyData, createdBy *uuid.UUID) *repository.CreatePolicyRequest {
	return &repository.CreatePolicyRequest{
		Name:        policyData.Name,
		DisplayName: policyData.DisplayName,
		Description: policyData.Description,
		PolicyType:  policyData.PolicyType,
		Effect:      policyData.Effect,
		Priority:    policyData.Priority,
		Category:    policyData.Category,
		Target:      policyData.Target,
		Rule:        policyData.Rule,
		Obligations: policyData.Obligations,
		Advice:      policyData.Advice,
		CreatedBy:   *createdBy,
	}
}

func (pm *policyManager) convertToUpdateRequest(policyData ImportPolicyData, id uuid.UUID) *repository.UpdatePolicyRequest {
	return &repository.UpdatePolicyRequest{
		DisplayName: &policyData.Name,
		Description: policyData.Description,
		Priority:    &policyData.Priority,
		Target:      policyData.Target,
		Rule:        policyData.Rule,
		Obligations: policyData.Obligations,
		Advice:      policyData.Advice,
		IsActive:    &policyData.IsActive,
	}
}

func (pm *policyManager) getPoliciesByFilters(ctx context.Context, filters PolicyExportFilters) ([]*models.Policy, error) {
	// This is a simplified implementation
	// In a real scenario, you would build complex filter queries
	var category *types.PolicyCategory
	if len(filters.Categories) > 0 {
		category = &filters.Categories[0]
	}

	listReq := &repository.ListPoliciesRequest{
		Category: category,
		IsActive: filters.IsActive,
		Limit:    1000,
		Offset:   0,
	}

	return pm.policyRepo.ListPolicies(ctx, listReq)
}

func (pm *policyManager) transformPoliciesForExport(ctx context.Context, policies []*models.Policy, options ExportOptions) []map[string]any {
	var exportData []map[string]any

	for _, policy := range policies {
		policyMap := map[string]any{
			"id":           policy.ID,
			"name":         policy.Name,
			"display_name": policy.DisplayName,
			"description":  policy.Description,
			"policy_type":  string(policy.PolicyType),
			"effect":       string(policy.Effect),
			"priority":     policy.Priority,
			"category":     string(policy.Category),
			"target":       policy.Target,
			"rule":         policy.Rule,
			"obligations":  policy.Obligations,
			"advice":       policy.Advice,
			"is_active":    policy.IsActive,
		}

		// Include metadata if requested
		if options.IncludeMetadata {
			policyMap["created_at"] = policy.CreatedAt
			policyMap["updated_at"] = policy.UpdatedAt
			policyMap["created_by"] = policy.CreatedBy
		}

		// Exclude specified fields
		for _, field := range options.ExcludeFields {
			delete(policyMap, field)
		}

		exportData = append(exportData, policyMap)
	}

	return exportData
}

func (pm *policyManager) generateExportData(ctx context.Context, data []map[string]any, format string, options ExportOptions) ([]byte, error) {
	switch strings.ToLower(format) {
	case "json":
		if options.MinifyOutput {
			return json.Marshal(data)
		} else {
			return json.MarshalIndent(data, "", "  ")
		}
	case "yaml":
		// YAML generation would go here
		return nil, fmt.Errorf("YAML format not yet implemented")
	case "xml":
		// XML generation would go here
		return nil, fmt.Errorf("XML format not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func (pm *policyManager) generateSplitExport(ctx context.Context, data []map[string]any, format string, options ExportOptions) map[string][]byte {
	files := make(map[string][]byte)

	// Group by category
	categoryGroups := make(map[string][]map[string]any)
	for _, policyData := range data {
		if category, ok := policyData["category"].(string); ok {
			categoryGroups[category] = append(categoryGroups[category], policyData)
		}
	}

	// Generate file for each category
	for category, policies := range categoryGroups {
		fileName := fmt.Sprintf("policies_%s.%s", strings.ToLower(category), format)
		if fileData, err := pm.generateExportData(ctx, policies, format, options); err == nil {
			files[fileName] = fileData
		}
	}

	return files
}

func (pm *policyManager) calculateChecksum(data []byte) string {
	// Simple checksum calculation
	// In a real implementation, you would use a proper hash function
	return fmt.Sprintf("%x", len(data))
}
