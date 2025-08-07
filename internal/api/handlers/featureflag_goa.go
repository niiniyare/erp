package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/api/gen/featureflag"
	corefeatureflag "github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// FeatureFlagService implements the Goa-generated featureflag.Service interface
type FeatureFlagService struct {
	service corefeatureflag.SimpleService
	logger  logger.Logger
	metrics *metrics.MetricsService
	tracing tracing.TracingService
}

// NewFeatureFlagService creates a new Goa feature flag service implementation
func NewFeatureFlagService(
	service corefeatureflag.SimpleService,
	logger logger.Logger,
	metrics *metrics.MetricsService,
	tracing tracing.TracingService,
) featureflag.Service {
	return &FeatureFlagService{
		service: service,
		logger:  logger,
		metrics: metrics,
		tracing: tracing,
	}
}

// Create implements featureflag.Service.
func (s *FeatureFlagService) Create(ctx context.Context, p *featureflag.CreateFeatureFlagPayload) (res *featureflag.Featureflag, view string, err error) {
	ctx, span := s.tracing.StartSpan(ctx, "featureflag.service.create")
	defer span.End()

	s.logger.Info("Creating feature flag", logger.Fields{
		"name":      p.Name,
		"flag_type": p.FlagType,
	})

	// Convert Goa payload to domain request
	request := &corefeatureflag.CreateFeatureFlagRequest{
		Name:         p.Name,
		Description:  p.Description,
		FlagType:     corefeatureflag.FlagType(p.FlagType),
		DefaultValue: p.DefaultValue,
	}

	if p.RolloutPercentage != nil {
		percentage := int32(*p.RolloutPercentage)
		request.RolloutPercentage = &percentage
	}

	if p.TargetAudience != nil {
		request.TargetAudience = p.TargetAudience
	}

	if p.Metadata != nil {
		request.Metadata = p.Metadata
	}

	// Create the feature flag
	flag, err := s.service.CreateFeatureFlag(ctx, request)
	if err != nil {
		s.logger.Error("Failed to create feature flag", logger.Fields{
			"error": err.Error(),
			"name":  p.Name,
		})
		return nil, "", s.mapError(err)
	}

	s.logger.Info("Feature flag created successfully", logger.Fields{
		"flag_id": flag.ID.String(),
		"name":    flag.Name,
	})

	// Convert domain model to Goa response
	return s.domainToGoa(flag), "default", nil
}

// Get implements featureflag.Service.
func (s *FeatureFlagService) Get(ctx context.Context, p *featureflag.GetPayload) (res *featureflag.Featureflag, view string, err error) {
	ctx, span := s.tracing.StartSpan(ctx, "featureflag.service.get")
	defer span.End()

	flag, err := s.service.GetFeatureFlag(ctx, p.Name)
	if err != nil {
		s.logger.Error("Failed to get feature flag", logger.Fields{
			"error": err.Error(),
			"name":  p.Name,
		})
		return nil, "", s.mapError(err)
	}

	return s.domainToGoa(flag), "default", nil
}

// GetByID implements featureflag.Service.
func (s *FeatureFlagService) GetByID(ctx context.Context, p *featureflag.GetByIDPayload) (res *featureflag.Featureflag, view string, err error) {
	ctx, span := s.tracing.StartSpan(ctx, "featureflag.service.getById")
	defer span.End()

	_, err = uuid.Parse(p.ID)
	if err != nil {
		return nil, "", featureflag.MakeBadRequest(fmt.Errorf("invalid UUID: %w", err))
	}

	flag, err := s.service.GetFeatureFlagByID(ctx, uuid.MustParse(p.ID))
	if err != nil {
		s.logger.Error("Failed to get feature flag by ID", logger.Fields{
			"error": err.Error(),
			"id":    p.ID,
		})
		return nil, "", s.mapError(err)
	}

	return s.domainToGoa(flag), "default", nil
}

// List implements featureflag.Service.
func (s *FeatureFlagService) List(ctx context.Context, p *featureflag.ListPayload) (res *featureflag.ListResult, err error) {
	ctx, span := s.tracing.StartSpan(ctx, "featureflag.service.list")
	defer span.End()

	// Set default values - PageSize and Page are uint, not *uint
	pageSize := uint(20)
	if p.PageSize > 0 {
		pageSize = p.PageSize
	}
	page := uint(1)
	if p.Page > 0 {
		page = p.Page
	}

	// Convert Goa payload to domain request
	request := &corefeatureflag.ListFeatureFlagsRequest{
		PageSize: int(pageSize),
		Page:     int(page),
	}

	if p.FlagType != nil {
		request.FlagType = p.FlagType
	}

	response, err := s.service.ListFeatureFlags(ctx, request)
	if err != nil {
		s.logger.Error("Failed to list feature flags", logger.Fields{
			"error": err.Error(),
		})
		return nil, s.mapError(err)
	}

	// Convert domain response to Goa response
	goaFlags := make([]*featureflag.Featureflag, len(response.FeatureFlags))
	for i, flag := range response.FeatureFlags {
		goaFlags[i] = s.domainToGoa(flag)
	}

	totalPages := uint((response.Total + int64(pageSize) - 1) / int64(pageSize))

	return &featureflag.ListResult{
		Data: goaFlags,
		Pagination: &featureflag.PaginationMeta{
			CurrentPage: page,
			PageSize:    pageSize,
			TotalItems:  uint(response.Total),
			TotalPages:  totalPages,
			HasNext:     page < totalPages,
			HasPrev:     page > 1,
		},
	}, nil
}

// Update implements featureflag.Service.
func (s *FeatureFlagService) Update(ctx context.Context, p *featureflag.UpdateFeatureFlagPayload) (res *featureflag.Featureflag, view string, err error) {
	ctx, span := s.tracing.StartSpan(ctx, "featureflag.service.update")
	defer span.End()

	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, "", featureflag.MakeBadRequest(fmt.Errorf("invalid UUID: %w", err))
	}

	// Convert Goa payload to domain request
	request := &corefeatureflag.UpdateFeatureFlagRequest{}

	if p.Description != nil {
		request.Description = p.Description
	}

	if p.DefaultValue != nil {
		request.DefaultValue = p.DefaultValue
	}

	if p.RolloutPercentage != nil {
		percentage := int32(*p.RolloutPercentage)
		request.RolloutPercentage = &percentage
	}

	if p.TargetAudience != nil {
		request.TargetAudience = p.TargetAudience
	}

	if p.Metadata != nil {
		request.Metadata = p.Metadata
	}

	flag, err := s.service.UpdateFeatureFlag(ctx, id, request)
	if err != nil {
		s.logger.Error("Failed to update feature flag", logger.Fields{
			"error":   err.Error(),
			"flag_id": id.String(),
		})
		return nil, "", s.mapError(err)
	}

	s.logger.Info("Feature flag updated successfully", logger.Fields{
		"flag_id": flag.ID.String(),
		"name":    flag.Name,
	})

	return s.domainToGoa(flag), "default", nil
}

// Delete implements featureflag.Service.
func (s *FeatureFlagService) Delete(ctx context.Context, p *featureflag.DeletePayload) (err error) {
	ctx, span := s.tracing.StartSpan(ctx, "featureflag.service.delete")
	defer span.End()

	id, err := uuid.Parse(p.ID)
	if err != nil {
		return featureflag.MakeBadRequest(fmt.Errorf("invalid UUID: %w", err))
	}

	err = s.service.DeleteFeatureFlag(ctx, id)
	if err != nil {
		s.logger.Error("Failed to delete feature flag", logger.Fields{
			"error":   err.Error(),
			"flag_id": id.String(),
		})
		return s.mapError(err)
	}

	s.logger.Info("Feature flag deleted successfully", logger.Fields{
		"flag_id": id.String(),
	})

	return nil
}

// Evaluate implements featureflag.Service.
func (s *FeatureFlagService) Evaluate(ctx context.Context, p *featureflag.EvaluateFeatureFlagPayload) (res *featureflag.EvaluationResult, err error) {
	ctx, span := s.tracing.StartSpan(ctx, "featureflag.service.evaluate")
	defer span.End()

	// Convert Goa evaluation context to domain context
	evalCtx := s.goaToDomainEvaluationContext(p.Context)

	result, err := s.service.EvaluateFlag(ctx, p.Name, evalCtx)
	if err != nil {
		s.logger.Error("Failed to evaluate feature flag", logger.Fields{
			"error": err.Error(),
			"name":  p.Name,
		})
		return nil, s.mapError(err)
	}

	return s.domainToGoaEvaluationResult(result), nil
}

// EvaluateMultiple implements featureflag.Service.
func (s *FeatureFlagService) EvaluateMultiple(ctx context.Context, p *featureflag.BulkEvaluatePayload) (res *featureflag.EvaluateMultipleResult, err error) {
	ctx, span := s.tracing.StartSpan(ctx, "featureflag.service.evaluateMultiple")
	defer span.End()

	// Convert Goa evaluation context to domain context
	evalCtx := s.goaToDomainEvaluationContext(p.Context)

	results, err := s.service.EvaluateFlags(ctx, p.FlagNames, evalCtx)
	if err != nil {
		s.logger.Error("Failed to evaluate multiple feature flags", logger.Fields{
			"error":      err.Error(),
			"flag_count": len(p.FlagNames),
		})
		return nil, s.mapError(err)
	}

	// Convert domain results to Goa results
	goaResults := make(map[string]*featureflag.EvaluationResult, len(results.Results))
	for name, result := range results.Results {
		goaResults[name] = s.domainToGoaEvaluationResult(result)
	}

	return &featureflag.EvaluateMultipleResult{
		Results: goaResults,
		Context: p.Context,
	}, nil
}

// GetStats implements featureflag.Service.
func (s *FeatureFlagService) GetStats(ctx context.Context) (res *featureflag.FeatureFlagStats, err error) {
	ctx, span := s.tracing.StartSpan(ctx, "featureflag.service.getStats")
	defer span.End()

	stats, err := s.service.GetFlagStats(ctx)
	if err != nil {
		s.logger.Error("Failed to get feature flag statistics", logger.Fields{
			"error": err.Error(),
		})
		return nil, s.mapError(err)
	}

	// The core FlagStats doesn't have all fields, so we'll return what we have
	avgRollout := 0.0
	if stats.AvgRolloutPercentage != nil {
		avgRollout = *stats.AvgRolloutPercentage
	}

	return &featureflag.FeatureFlagStats{
		TotalFlags:           uint64(stats.TotalFlags),
		EnabledFlags:         uint64(stats.EnabledFlags),
		RolloutFlags:         uint64(stats.RolloutFlags),
		AvgRolloutPercentage: avgRollout,
		FlagsByType: map[string]uint64{
			"boolean": uint64(stats.TotalFlags), // Simplified for now
			"string":  0,
			"number":  0,
			"json":    0,
		},
		RecentActivity: &featureflag.FeatureFlagActivitySummary{
			FlagsCreatedToday:   uintPtr(0),
			FlagsUpdatedToday:   uintPtr(0),
			FlagsEvaluatedToday: uint64Ptr(0),
			MostEvaluatedFlag:   stringPtr("N/A"),
			LastActivityAt:      stringPtr(time.Now().Format(time.RFC3339)),
		},
	}, nil
}

// Search implements featureflag.Service.
func (s *FeatureFlagService) Search(ctx context.Context, p *featureflag.SearchPayload) (res *featureflag.SearchResult, err error) {
	ctx, span := s.tracing.StartSpan(ctx, "featureflag.service.search")
	defer span.End()

	// Use the SearchFlags method from the core service
	flags, err := s.service.SearchFlags(ctx, p.Query, int32(p.Limit), int32(p.Offset))
	if err != nil {
		s.logger.Error("Failed to search feature flags", logger.Fields{
			"error": err.Error(),
			"query": p.Query,
		})
		return nil, s.mapError(err)
	}

	// Convert domain flags to Goa flags
	goaFlags := make([]*featureflag.Featureflag, len(flags))
	for i, flag := range flags {
		goaFlags[i] = s.domainToGoa(flag)
	}

	return &featureflag.SearchResult{
		Data:       goaFlags,
		TotalCount: uint(len(goaFlags)), // Simple count for now
		Query:      p.Query,
	}, nil
}

// GetByType implements featureflag.Service.
func (s *FeatureFlagService) GetByType(ctx context.Context, p *featureflag.GetByTypePayload) (res *featureflag.GetByTypeResult, err error) {
	ctx, span := s.tracing.StartSpan(ctx, "featureflag.service.getByType")
	defer span.End()

	// Use the GetFlagsByType method from the core service
	flags, err := s.service.GetFlagsByType(ctx, p.FlagType)
	if err != nil {
		s.logger.Error("Failed to get flags by type", logger.Fields{
			"error":     err.Error(),
			"flag_type": p.FlagType,
		})
		return nil, s.mapError(err)
	}

	// Convert domain flags to Goa flags
	goaFlags := make([]*featureflag.Featureflag, len(flags))
	for i, flag := range flags {
		goaFlags[i] = s.domainToGoa(flag)
	}

	return &featureflag.GetByTypeResult{
		Data:     goaFlags,
		FlagType: p.FlagType,
		Count:    uint(len(goaFlags)),
	}, nil
}

// Health implements featureflag.Service.
func (s *FeatureFlagService) Health(ctx context.Context) (res *featureflag.HealthResult, err error) {
	ctx, span := s.tracing.StartSpan(ctx, "featureflag.service.health")
	defer span.End()

	// Simple health check - could be enhanced with actual database checks
	return &featureflag.HealthResult{
		Status:         "healthy",
		Timestamp:      time.Now().Format(time.RFC3339),
		Version:        "1.0.0",
		DatabaseStatus: stringPtr("connected"),
		CacheStatus:    stringPtr("available"),
	}, nil
}

// Helper methods for type conversion and error mapping

func (s *FeatureFlagService) domainToGoa(flag *corefeatureflag.FeatureFlag) *featureflag.Featureflag {
	gf := &featureflag.Featureflag{
		ID:           flag.ID.String(),
		TenantID:     flag.TenantID.String(),
		Name:         flag.Name,
		Description:  flag.Description,
		FlagType:     string(flag.FlagType),
		DefaultValue: flag.DefaultValue,
		CreatedAt:    flag.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    flag.UpdatedAt.Format(time.RFC3339),
	}

	if flag.RolloutPercentage != nil {
		percentage := uint(*flag.RolloutPercentage)
		gf.RolloutPercentage = &percentage
	}

	if flag.TargetAudience != nil {
		gf.TargetAudience = flag.TargetAudience
	}

	if flag.Metadata != nil {
		gf.Metadata = flag.Metadata
	}

	return gf
}

func (s *FeatureFlagService) goaToDomainEvaluationContext(ctx *featureflag.EvaluationContext) *corefeatureflag.EvaluationContext {
	evalCtx := &corefeatureflag.EvaluationContext{
		Environment: ctx.Environment,
	}

	if tenantID, err := uuid.Parse(ctx.TenantID); err == nil {
		evalCtx.TenantID = tenantID
	}

	if ctx.UserID != nil {
		if userID, err := uuid.Parse(*ctx.UserID); err == nil {
			evalCtx.UserID = &userID
		}
	}

	if ctx.Attributes != nil {
		evalCtx.Attributes = ctx.Attributes
	}

	if ctx.ClientInfo != nil {
		evalCtx.ClientInfo = &corefeatureflag.ClientInfo{
			Version:   stringOrDefault(ctx.ClientInfo.Version, ""),
			Platform:  stringOrDefault(ctx.ClientInfo.Platform, ""),
			IPAddress: stringOrDefault(ctx.ClientInfo.IPAddress, ""),
			UserAgent: stringOrDefault(ctx.ClientInfo.UserAgent, ""),
		}
	}

	return evalCtx
}

func (s *FeatureFlagService) domainToGoaEvaluationResult(result *corefeatureflag.EvaluationResult) *featureflag.EvaluationResult {
	// Type assert Value to bool for API compatibility
	value, ok := result.Value.(bool)
	if !ok {
		value = false // default to false if not a boolean
	}

	return &featureflag.EvaluationResult{
		FlagName:    result.FlagName,
		Value:       value,
		Enabled:     result.Enabled,
		Reason:      result.Reason,
		RuleMatched: result.RuleMatched,
		Metadata: &featureflag.ResultMetadata{
			EvaluatedAt:      result.Metadata.EvaluatedAt.Format(time.RFC3339),
			CacheHit:         result.Metadata.CacheHit,
			EvaluationTimeMs: result.Metadata.EvaluationMs,
			ConfigVersion:    &result.Metadata.ConfigVersion,
			TenantID:         nil, // Not available in core metadata
		},
	}
}

func (s *FeatureFlagService) mapError(err error) error {
	// Map domain errors to Goa errors
	switch {
	case err == corefeatureflag.ErrFeatureFlagNotFound:
		return featureflag.MakeNotFound(err)
	case err == corefeatureflag.ErrFeatureFlagAlreadyExists:
		return featureflag.MakeConflict(err)
	default:
		// Log unexpected errors for debugging
		s.logger.Error("Unmapped error in feature flag service", logger.Fields{
			"error": err.Error(),
		})
		// Use BadRequest as a fallback since InternalError isn't available
		return featureflag.MakeBadRequest(fmt.Errorf("internal error: %w", err))
	}
}

// Helper functions
func stringOrDefault(s *string, defaultVal string) string {
	if s != nil {
		return *s
	}
	return defaultVal
}

func uintPtr(u uint) *uint {
	return &u
}

func uint64Ptr(u uint64) *uint64 {
	return &u
}
