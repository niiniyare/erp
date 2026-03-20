package service

import (
	"context"
	"fmt"
	"time"

	"awo/internal/core/tenant/domain"
	"awo/internal/core/tenant/repository"
	"awo/internal/platform/cache"
	"awo/internal/shared/logger"
	"awo/internal/shared/tracing"
)

const analyticsCacheTTL = 5 * time.Minute

// AnalyticsService provides tenant analytics and statistics.
type AnalyticsService struct {
	repo   repository.Repository
	cache  cache.Service
	tracer tracing.Service
	log    logger.Logger
}

// NewAnalyticsService creates a new AnalyticsService.
func NewAnalyticsService(repo repository.Repository, c cache.Service, tracer tracing.Service, log logger.Logger) *AnalyticsService {
	return &AnalyticsService{repo: repo, cache: c, tracer: tracer, log: log}
}

// GetGrowthStats returns tenant growth statistics for the given number of days.
func (s *AnalyticsService) GetGrowthStats(ctx context.Context, days int) ([]domain.GrowthStat, error) {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.analytics.GrowthStats")
	defer span.End()

	stats, err := s.repo.GetGrowthStats(ctx, days)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get growth stats: %w", err)
	}
	return stats, nil
}

// GetStatusDistribution returns the distribution of tenant statuses.
func (s *AnalyticsService) GetStatusDistribution(ctx context.Context) (*domain.StatusCount, error) {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.analytics.StatusDistribution")
	defer span.End()

	dist, err := s.repo.GetStatusDistribution(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get status distribution: %w", err)
	}
	return dist, nil
}
