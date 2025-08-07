package featureflag

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/audit"
	"github.com/niiniyare/erp/internal/core/tenant"
)

// SimpleService defines the business logic interface for feature flags
type SimpleService interface {
	// Feature Flag Management
	CreateFeatureFlag(ctx context.Context, request *CreateFeatureFlagRequest) (*FeatureFlag, error)
	GetFeatureFlag(ctx context.Context, name string) (*FeatureFlag, error)
	GetFeatureFlagByID(ctx context.Context, id uuid.UUID) (*FeatureFlag, error)
	UpdateFeatureFlag(ctx context.Context, id uuid.UUID, request *UpdateFeatureFlagRequest) (*FeatureFlag, error)
	DeleteFeatureFlag(ctx context.Context, id uuid.UUID) error
	ListFeatureFlags(ctx context.Context, request *ListFeatureFlagsRequest) (*ListFeatureFlagsResponse, error)

	// Flag Evaluation
	EvaluateFlag(ctx context.Context, name string, evalCtx *EvaluationContext) (*EvaluationResult, error)
	EvaluateFlags(ctx context.Context, names []string, evalCtx *EvaluationContext) (*BulkEvaluationResponse, error)
	IsEnabled(ctx context.Context, name string, evalCtx *EvaluationContext) (bool, error)

	// Analytics and Monitoring
	GetFlagStats(ctx context.Context) (*FlagStats, error)
	SearchFlags(ctx context.Context, query string, limit, offset int32) ([]*FeatureFlag, error)
	GetFlagsByType(ctx context.Context, flagType string) ([]*FeatureFlag, error)
}

// simpleServiceImpl implements the SimpleService interface
type simpleServiceImpl struct {
	repository    SimpleRepository
	tenantService tenant.Service
	store         db.Store
	auditService  audit.Service
}

// NewSimpleService creates a new simple feature flag service
func NewSimpleService(repository SimpleRepository, tenantService tenant.Service, store db.Store, auditService audit.Service) SimpleService {
	return &simpleServiceImpl{
		repository:    repository,
		tenantService: tenantService,
		store:         store,
		auditService:  auditService,
	}
}

// Request/Response types
type ListFeatureFlagsRequest struct {
	Page     int     `json:"page" validate:"min=1"`
	PageSize int     `json:"page_size" validate:"min=1,max=100"`
	FlagType *string `json:"flag_type,omitempty"`
}

type ListFeatureFlagsResponse struct {
	FeatureFlags []*FeatureFlag `json:"feature_flags"`
	Total        int64          `json:"total"`
	Page         int            `json:"page"`
	PageSize     int            `json:"page_size"`
}

// CreateFeatureFlag creates a new feature flag with tenant context validation
func (s *simpleServiceImpl) CreateFeatureFlag(ctx context.Context, request *CreateFeatureFlagRequest) (*FeatureFlag, error) {
	// Validate tenant context exists
	if err := s.tenantService.ValidateCurrentTenant(ctx); err != nil {
		return nil, fmt.Errorf("invalid tenant context: %w", err)
	}

	// Check for duplicate name
	existing, err := s.repository.GetFeatureFlagByName(ctx, request.Name)
	if err != nil && err != ErrFeatureFlagNotFound {
		return nil, fmt.Errorf("failed to check existing flag: %w", err)
	}
	if existing != nil {
		return nil, ErrFeatureFlagAlreadyExists
	}

	// Validate flag type
	switch request.FlagType {
	case FlagTypeBoolean, FlagTypeString, FlagTypeNumber, FlagTypeJSON:
		// Valid types
	default:
		return nil, fmt.Errorf("invalid flag type: %s", request.FlagType)
	}

	// Create the flag
	flag, err := s.repository.CreateFeatureFlag(ctx, request)
	if err != nil {
		return nil, err
	}

	// Audit the creation
	s.auditFlagCreated(ctx, flag, request)

	return flag, nil
}

// GetFeatureFlag retrieves a feature flag by name with tenant context
func (s *simpleServiceImpl) GetFeatureFlag(ctx context.Context, name string) (*FeatureFlag, error) {
	if err := s.tenantService.ValidateCurrentTenant(ctx); err != nil {
		return nil, fmt.Errorf("invalid tenant context: %w", err)
	}

	return s.repository.GetFeatureFlagByName(ctx, name)
}

// GetFeatureFlagByID retrieves a feature flag by ID with tenant context
func (s *simpleServiceImpl) GetFeatureFlagByID(ctx context.Context, id uuid.UUID) (*FeatureFlag, error) {
	if err := s.tenantService.ValidateCurrentTenant(ctx); err != nil {
		return nil, fmt.Errorf("invalid tenant context: %w", err)
	}

	return s.repository.GetFeatureFlagByID(ctx, id)
}

// UpdateFeatureFlag updates a feature flag with tenant context validation
func (s *simpleServiceImpl) UpdateFeatureFlag(ctx context.Context, id uuid.UUID, request *UpdateFeatureFlagRequest) (*FeatureFlag, error) {
	if err := s.tenantService.ValidateCurrentTenant(ctx); err != nil {
		return nil, fmt.Errorf("invalid tenant context: %w", err)
	}

	// Check if flag exists
	_, err := s.repository.GetFeatureFlagByID(ctx, id)
	if err != nil {
		return nil, err
	}

	flag, err := s.repository.UpdateFeatureFlag(ctx, id, request)
	if err != nil {
		return nil, err
	}

	// Audit the update
	s.auditFlagUpdated(ctx, flag, request)

	return flag, nil
}

// DeleteFeatureFlag deletes a feature flag with tenant context validation
func (s *simpleServiceImpl) DeleteFeatureFlag(ctx context.Context, id uuid.UUID) error {
	if err := s.tenantService.ValidateCurrentTenant(ctx); err != nil {
		return fmt.Errorf("invalid tenant context: %w", err)
	}

	// Check if flag exists and get it for audit logging
	flag, err := s.repository.GetFeatureFlagByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete the flag
	err = s.repository.DeleteFeatureFlag(ctx, id)
	if err != nil {
		return err
	}

	// Audit the deletion
	s.auditFlagDeleted(ctx, flag)

	return nil
}

// ListFeatureFlags lists feature flags with tenant context
func (s *simpleServiceImpl) ListFeatureFlags(ctx context.Context, request *ListFeatureFlagsRequest) (*ListFeatureFlagsResponse, error) {
	if err := s.tenantService.ValidateCurrentTenant(ctx); err != nil {
		return nil, fmt.Errorf("invalid tenant context: %w", err)
	}

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

	offset := (request.Page - 1) * request.PageSize

	flags, err := s.repository.ListFeatureFlags(ctx, ListFeatureFlagsParams{
		FlagType: request.FlagType,
		Limit:    int32(request.PageSize),
		Offset:   int32(offset),
	})
	if err != nil {
		return nil, err
	}

	// For total count, we would need a separate count query
	// This is simplified for now
	total := int64(len(flags))

	response := &ListFeatureFlagsResponse{
		FeatureFlags: flags,
		Total:        total,
		Page:         request.Page,
		PageSize:     request.PageSize,
	}

	// Audit the listing
	s.auditFlagListed(ctx, request, response)

	return response, nil
}

// EvaluateFlag evaluates a single feature flag with tenant context
func (s *simpleServiceImpl) EvaluateFlag(ctx context.Context, name string, evalCtx *EvaluationContext) (*EvaluationResult, error) {
	if err := s.tenantService.ValidateCurrentTenant(ctx); err != nil {
		return nil, fmt.Errorf("invalid tenant context: %w", err)
	}

	// Get the flag
	flag, err := s.repository.GetFeatureFlagByName(ctx, name)
	if err != nil {
		if err == ErrFeatureFlagNotFound {
			// Return default disabled result for non-existent flags
			return &EvaluationResult{
				FlagName: name,
				Value:    false,
				Enabled:  false,
				Reason:   string(ReasonFlagNotFound),
				Metadata: ResultMetadata{
					EvaluatedAt:   time.Now(),
					CacheHit:      false,
					EvaluationMs:  0.0,
					ConfigVersion: "1.0",
				},
			}, nil
		}
		return nil, err
	}

	// Simple evaluation logic
	result := s.evaluateSimpleFlag(flag, evalCtx)

	// Audit the evaluation
	s.auditFlagEvaluated(ctx, flag, evalCtx, result)

	return result, nil
}

// EvaluateFlags evaluates multiple feature flags with tenant context
func (s *simpleServiceImpl) EvaluateFlags(ctx context.Context, names []string, evalCtx *EvaluationContext) (*BulkEvaluationResponse, error) {
	if err := s.tenantService.ValidateCurrentTenant(ctx); err != nil {
		return nil, fmt.Errorf("invalid tenant context: %w", err)
	}

	results := make(map[string]*EvaluationResult)
	errors := make(map[string]string)

	for _, name := range names {
		result, err := s.EvaluateFlag(ctx, name, evalCtx)
		if err != nil {
			errors[name] = err.Error()
		} else {
			results[name] = result
		}
	}

	response := &BulkEvaluationResponse{
		Results: results,
	}

	if len(errors) > 0 {
		response.Errors = errors
	}

	return response, nil
}

// IsEnabled checks if a feature flag is enabled with tenant context
func (s *simpleServiceImpl) IsEnabled(ctx context.Context, name string, evalCtx *EvaluationContext) (bool, error) {
	result, err := s.EvaluateFlag(ctx, name, evalCtx)
	if err != nil {
		return false, err
	}
	return result.Enabled, nil
}

// GetFlagStats retrieves feature flag statistics with tenant context
func (s *simpleServiceImpl) GetFlagStats(ctx context.Context) (*FlagStats, error) {
	if err := s.tenantService.ValidateCurrentTenant(ctx); err != nil {
		return nil, fmt.Errorf("invalid tenant context: %w", err)
	}

	return s.repository.GetFlagStats(ctx)
}

// SearchFlags searches feature flags with tenant context
func (s *simpleServiceImpl) SearchFlags(ctx context.Context, query string, limit, offset int32) ([]*FeatureFlag, error) {
	if err := s.tenantService.ValidateCurrentTenant(ctx); err != nil {
		return nil, fmt.Errorf("invalid tenant context: %w", err)
	}

	results, err := s.repository.SearchFlags(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}

	// Audit the search
	s.auditFlagSearched(ctx, query, limit, offset, results)

	return results, nil
}

// GetFlagsByType retrieves flags by type with tenant context
func (s *simpleServiceImpl) GetFlagsByType(ctx context.Context, flagType string) ([]*FeatureFlag, error) {
	if err := s.tenantService.ValidateCurrentTenant(ctx); err != nil {
		return nil, fmt.Errorf("invalid tenant context: %w", err)
	}

	return s.repository.GetFlagsByType(ctx, flagType)
}

// evaluateSimpleFlag performs basic flag evaluation
func (s *simpleServiceImpl) evaluateSimpleFlag(flag *FeatureFlag, evalCtx *EvaluationContext) *EvaluationResult {
	now := time.Now()

	// Basic evaluation logic
	var enabled bool
	var reason EvaluationReason
	var value interface{}

	// Check rollout percentage if present
	if flag.RolloutPercentage != nil && *flag.RolloutPercentage > 0 {
		// Simple hash-based rollout
		hash := s.calculateSimpleHash(flag.Name, evalCtx.UserID)
		if hash < int(*flag.RolloutPercentage) {
			enabled = flag.DefaultValue
			reason = ReasonPercentRollout
			value = flag.DefaultValue
		} else {
			enabled = false
			reason = ReasonPercentRollout
			value = false
		}
	} else {
		// Use default value
		enabled = flag.DefaultValue
		reason = ReasonDefaultValue
		value = flag.DefaultValue
	}

	return &EvaluationResult{
		FlagName: flag.Name,
		Value:    value,
		Enabled:  enabled,
		Reason:   string(reason),
		Metadata: ResultMetadata{
			EvaluatedAt:   now,
			CacheHit:      false,
			EvaluationMs:  1.0, // Simple timing
			ConfigVersion: "1.0",
		},
	}
}

// calculateSimpleHash calculates a simple hash for rollout distribution
func (s *simpleServiceImpl) calculateSimpleHash(flagName string, userID *uuid.UUID) int {
	var input string
	if userID != nil {
		input = flagName + ":" + userID.String()
	} else {
		input = flagName + ":anonymous"
	}

	// Simple hash calculation
	hash := 0
	for _, char := range input {
		hash = hash*31 + int(char)
	}

	// Return value between 0-99
	return (hash%100 + 100) % 100
}

// Audit Event Types
const (
	AuditEventFeatureFlagCreated   = "feature_flag_created"
	AuditEventFeatureFlagUpdated   = "feature_flag_updated"
	AuditEventFeatureFlagDeleted   = "feature_flag_deleted"
	AuditEventFeatureFlagEvaluated = "feature_flag_evaluated"
	AuditEventFeatureFlagListed    = "feature_flag_listed"
	AuditEventFeatureFlagSearched  = "feature_flag_searched"
)

// Audit Event Categories
const (
	AuditCategoryFeatureManagement = "FEATURE_MANAGEMENT"
	AuditCategoryAccess            = "ACCESS"
	AuditCategoryAdmin             = "ADMIN"
)

// Audit Severity Levels
const (
	AuditSeverityInfo     = "INFO"
	AuditSeverityWarn     = "WARN"
	AuditSeverityHigh     = "HIGH"
	AuditSeverityCritical = "CRITICAL"
)

// auditFlagCreated logs flag creation event
func (s *simpleServiceImpl) auditFlagCreated(ctx context.Context, flag *FeatureFlag, request *CreateFeatureFlagRequest) {
	contextData, _ := json.Marshal(map[string]any{
		"flag_id":            flag.ID.String(),
		"flag_name":          flag.Name,
		"flag_type":          string(flag.FlagType),
		"default_value":      flag.DefaultValue,
		"rollout_percentage": flag.RolloutPercentage,
		"target_audience":    flag.TargetAudience,
		"metadata":           flag.Metadata,
		"tenant_id":          flag.TenantID.String(),
	})

	s.auditService.Record(ctx, audit.AuditEvent{
		EventType:     AuditEventFeatureFlagCreated,
		EventCategory: AuditCategoryFeatureManagement,
		Severity:      AuditSeverityInfo,
		EntityID:      uuid.NullUUID{UUID: flag.ID, Valid: true},
		Decision:      "CREATED",
		Reason:        fmt.Sprintf("Feature flag '%s' created successfully", flag.Name),
		Context:       contextData,
	})
}

// auditFlagUpdated logs flag update event
func (s *simpleServiceImpl) auditFlagUpdated(ctx context.Context, flag *FeatureFlag, request *UpdateFeatureFlagRequest) {
	contextData, _ := json.Marshal(map[string]any{
		"flag_id":        flag.ID.String(),
		"flag_name":      flag.Name,
		"updates":        request,
		"tenant_id":      flag.TenantID.String(),
		"updated_fields": s.getUpdatedFields(request),
	})

	severity := AuditSeverityInfo
	if request.DefaultValue != nil || request.RolloutPercentage != nil {
		severity = AuditSeverityWarn // Value changes are more significant
	}

	s.auditService.Record(ctx, audit.AuditEvent{
		EventType:     AuditEventFeatureFlagUpdated,
		EventCategory: AuditCategoryFeatureManagement,
		Severity:      severity,
		EntityID:      uuid.NullUUID{UUID: flag.ID, Valid: true},
		Decision:      "UPDATED",
		Reason:        fmt.Sprintf("Feature flag '%s' updated successfully", flag.Name),
		Context:       contextData,
	})
}

// auditFlagDeleted logs flag deletion event
func (s *simpleServiceImpl) auditFlagDeleted(ctx context.Context, flag *FeatureFlag) {
	contextData, _ := json.Marshal(map[string]any{
		"flag_id":    flag.ID.String(),
		"flag_name":  flag.Name,
		"flag_type":  string(flag.FlagType),
		"tenant_id":  flag.TenantID.String(),
		"deleted_at": time.Now(),
	})

	s.auditService.Record(ctx, audit.AuditEvent{
		EventType:     AuditEventFeatureFlagDeleted,
		EventCategory: AuditCategoryFeatureManagement,
		Severity:      AuditSeverityHigh, // Deletions are high severity
		EntityID:      uuid.NullUUID{UUID: flag.ID, Valid: true},
		Decision:      "DELETED",
		Reason:        fmt.Sprintf("Feature flag '%s' deleted successfully", flag.Name),
		Context:       contextData,
	})
}

// auditFlagEvaluated logs flag evaluation event
func (s *simpleServiceImpl) auditFlagEvaluated(ctx context.Context, flag *FeatureFlag, evalCtx *EvaluationContext, result *EvaluationResult) {
	contextData, _ := json.Marshal(map[string]any{
		"flag_name":          flag.Name,
		"flag_id":            flag.ID.String(),
		"evaluation_result":  result.Value,
		"enabled":            result.Enabled,
		"reason":             result.Reason,
		"rule_matched":       result.RuleMatched,
		"evaluation_context": evalCtx,
		"tenant_id":          flag.TenantID.String(),
		"evaluation_time_ms": result.Metadata.EvaluationMs,
		"cache_hit":          result.Metadata.CacheHit,
	})

	s.auditService.Record(ctx, audit.AuditEvent{
		EventType:     AuditEventFeatureFlagEvaluated,
		EventCategory: AuditCategoryAccess,
		Severity:      AuditSeverityInfo,
		EntityID:      uuid.NullUUID{UUID: flag.ID, Valid: true},
		Decision:      fmt.Sprintf("EVALUATED_%t", result.Enabled),
		Reason:        fmt.Sprintf("Feature flag '%s' evaluated: %v", flag.Name, result.Value),
		Context:       contextData,
	})
}

// auditFlagListed logs flag listing event
func (s *simpleServiceImpl) auditFlagListed(ctx context.Context, request *ListFeatureFlagsRequest, response *ListFeatureFlagsResponse) {
	contextData, _ := json.Marshal(map[string]any{
		"page":      request.Page,
		"page_size": request.PageSize,
		"flag_type": request.FlagType,
		"total":     response.Total,
		"returned":  len(response.FeatureFlags),
	})

	s.auditService.Record(ctx, audit.AuditEvent{
		EventType:     AuditEventFeatureFlagListed,
		EventCategory: AuditCategoryAccess,
		Severity:      AuditSeverityInfo,
		Decision:      "LISTED",
		Reason:        fmt.Sprintf("Listed %d feature flags (page %d)", len(response.FeatureFlags), request.Page),
		Context:       contextData,
	})
}

// auditFlagSearched logs flag search event
func (s *simpleServiceImpl) auditFlagSearched(ctx context.Context, query string, limit, offset int32, results []*FeatureFlag) {
	contextData, _ := json.Marshal(map[string]any{
		"query":         query,
		"limit":         limit,
		"offset":        offset,
		"results_found": len(results),
		"searched_at":   time.Now(),
	})

	s.auditService.Record(ctx, audit.AuditEvent{
		EventType:     AuditEventFeatureFlagSearched,
		EventCategory: AuditCategoryAccess,
		Severity:      AuditSeverityInfo,
		Decision:      "SEARCHED",
		Reason:        fmt.Sprintf("Searched feature flags with query '%s', found %d results", query, len(results)),
		Context:       contextData,
	})
}

// getUpdatedFields returns a list of fields that were updated
func (s *simpleServiceImpl) getUpdatedFields(request *UpdateFeatureFlagRequest) []string {
	var fields []string
	if request.Description != nil {
		fields = append(fields, "description")
	}
	if request.DefaultValue != nil {
		fields = append(fields, "default_value")
	}
	if request.RolloutPercentage != nil {
		fields = append(fields, "rollout_percentage")
	}
	if request.TargetAudience != nil {
		fields = append(fields, "target_audience")
	}
	if request.Metadata != nil {
		fields = append(fields, "metadata")
	}
	return fields
}
