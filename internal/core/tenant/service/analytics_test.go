package service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/niiniyare/erp/internal/core/tenant/domain"
	"github.com/niiniyare/erp/internal/core/tenant/service"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ---------------------------------------------------------------------------
// AnalyticsServiceSuite
// ---------------------------------------------------------------------------

type AnalyticsServiceSuite struct {
	suite.Suite
	repo *mockRepo
	svc  *service.AnalyticsService
}

func TestAnalyticsServiceSuite(t *testing.T) { suite.Run(t, new(AnalyticsServiceSuite)) }

func (s *AnalyticsServiceSuite) SetupTest() {
	s.repo = newMockRepo()
	s.svc = service.NewAnalyticsService(s.repo, newMockCache(), noopTracer{}, noopLogger{})
}

// ---- GetGrowthStats -------------------------------------------------------

func (s *AnalyticsServiceSuite) TestGetGrowthStats_Success() {
	expected := []domain.GrowthStat{
		{Month: "2025-01", NewTenants: 10, ActiveNewTenants: 8},
		{Month: "2025-02", NewTenants: 15, ActiveNewTenants: 12},
	}
	s.repo.getGrowthStatsFn = func(_ context.Context, days int) ([]domain.GrowthStat, error) {
		require.Equal(s.T(), 180, days)
		return expected, nil
	}

	stats, err := s.svc.GetGrowthStats(context.Background(), 180)

	require.NoError(s.T(), err)
	require.Len(s.T(), stats, 2)
	require.Equal(s.T(), expected, stats)
}

func (s *AnalyticsServiceSuite) TestGetGrowthStats_RepoError() {
	s.repo.getGrowthStatsFn = func(_ context.Context, _ int) ([]domain.GrowthStat, error) {
		return nil, fmt.Errorf("query timeout")
	}

	stats, err := s.svc.GetGrowthStats(context.Background(), 30)

	require.Error(s.T(), err)
	require.Nil(s.T(), stats)
	require.Contains(s.T(), err.Error(), "failed to get growth stats")
}

func (s *AnalyticsServiceSuite) TestGetGrowthStats_Empty() {
	s.repo.getGrowthStatsFn = func(_ context.Context, _ int) ([]domain.GrowthStat, error) {
		return []domain.GrowthStat{}, nil
	}

	stats, err := s.svc.GetGrowthStats(context.Background(), 7)

	require.NoError(s.T(), err)
	require.Empty(s.T(), stats)
}

// ---- GetStatusDistribution ------------------------------------------------

func (s *AnalyticsServiceSuite) TestGetStatusDistribution_Success() {
	expected := &domain.StatusCount{
		TotalTenants:     100,
		ActiveTenants:    70,
		SuspendedTenants: 20,
		PendingTenants:   10,
		ActivePct:        70.0,
		SuspendedPct:     20.0,
		PendingPct:       10.0,
	}
	s.repo.getStatusDistFn = func(_ context.Context) (*domain.StatusCount, error) {
		return expected, nil
	}

	dist, err := s.svc.GetStatusDistribution(context.Background())

	require.NoError(s.T(), err)
	require.NotNil(s.T(), dist)
	require.Equal(s.T(), expected, dist)
}

func (s *AnalyticsServiceSuite) TestGetStatusDistribution_RepoError() {
	s.repo.getStatusDistFn = func(_ context.Context) (*domain.StatusCount, error) {
		return nil, fmt.Errorf("db offline")
	}

	dist, err := s.svc.GetStatusDistribution(context.Background())

	require.Error(s.T(), err)
	require.Nil(s.T(), dist)
	require.Contains(s.T(), err.Error(), "failed to get status distribution")
}
