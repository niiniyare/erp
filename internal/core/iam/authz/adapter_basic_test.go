package authz

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/niiniyare/erp/internal/core/abac"
	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/core/access"
	"github.com/niiniyare/erp/internal/core/iam/model"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

func TestNewAdapter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockABAC := abac.NewMockService(ctrl)
	mockAccess := access.NewMockService(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockMetrics := metrics.NewMockMetricsProvider(ctrl)
	mockTracer := tracing.NewMockTracingService(ctrl)

	// Test adapter creation
	adapter := NewAdapter(mockABAC, mockAccess, mockLogger, mockMetrics, mockTracer)
	require.NotNil(t, adapter)
}

func TestEvaluatePermission_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockABAC := abac.NewMockService(ctrl)
	mockAccess := access.NewMockService(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockMetrics := metrics.NewMockMetricsProvider(ctrl)
	mockTracer := tracing.NewMockTracingService(ctrl)

	// Setup common mock expectations
	mockSpan := tracing.NewMockSpan(ctrl)
	mockSpan.EXPECT().End().AnyTimes()
	mockSpan.EXPECT().SetAttributes(gomock.Any()).AnyTimes()
	mockTracer.EXPECT().StartSpan(gomock.Any(), gomock.Any()).
		Return(context.Background(), mockSpan).AnyTimes()
	mockMetrics.EXPECT().ObserveHistogram(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockMetrics.EXPECT().IncrementCounter(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	// Create adapter
	adapter := NewAdapter(mockABAC, mockAccess, mockLogger, mockMetrics, mockTracer)

	// Test data
	ctx := context.Background()
	userID := uuid.New()
	req := &PermissionEvaluationRequest{
		UserID:       userID,
		ResourceType: "document",
		Action:       "read",
		RequestID:    "test-request",
	}

	abacResult := &abac.PermissionEvaluationResult{
		Decision: types.PolicyDecisionAllow,
		PolicyDecisions: []*models.PolicyDecision{
			{
				PolicyID:      uuid.New(),
				Decision:      types.PolicyDecisionAllow,
				Reason:        "Policy allows access",
				MatchedRule:   "default_rule",
				EvaluationMS:  5,
				TargetMatched: true,
			},
		},
		EvaluationTimeMS: 25,
		CacheHit:         false,
		RequestID:        "test-request",
		Timestamp:        time.Now(),
	}

	// Setup ABAC mock expectation
	mockABAC.EXPECT().EvaluatePermission(gomock.Any(), gomock.Any()).
		Return(abacResult, nil)

	// Execute test
	result, err := adapter.EvaluatePermission(ctx, req)

	// Verify results
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, model.PolicyDecisionAllow, result.Decision)
	require.Equal(t, int64(25), result.EvaluationTimeMS)
	require.Equal(t, "test-request", result.RequestID)
	require.Len(t, result.PolicyDecisions, 1)
}

func TestEvaluatePermission_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockABAC := abac.NewMockService(ctrl)
	mockAccess := access.NewMockService(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockMetrics := metrics.NewMockMetricsProvider(ctrl)
	mockTracer := tracing.NewMockTracingService(ctrl)

	// Setup common mock expectations
	mockSpan := tracing.NewMockSpan(ctrl)
	mockSpan.EXPECT().End().AnyTimes()
	mockSpan.EXPECT().SetAttributes(gomock.Any()).AnyTimes()
	mockTracer.EXPECT().StartSpan(gomock.Any(), gomock.Any()).
		Return(context.Background(), mockSpan).AnyTimes()
	mockTracer.EXPECT().RecordError(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockMetrics.EXPECT().ObserveHistogram(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockMetrics.EXPECT().IncrementCounter(gomock.Any(), gomock.Any()).AnyTimes()

	// Create adapter
	adapter := NewAdapter(mockABAC, mockAccess, mockLogger, mockMetrics, mockTracer)

	// Test data
	ctx := context.Background()
	userID := uuid.New()
	req := &PermissionEvaluationRequest{
		UserID:       userID,
		ResourceType: "document",
		Action:       "read",
		RequestID:    "test-request",
	}

	// Setup ABAC mock to return error
	expectedErr := errors.NewBusinessError("ABAC_ERROR", "ABAC service failed")
	mockABAC.EXPECT().EvaluatePermission(gomock.Any(), gomock.Any()).
		Return(nil, expectedErr)

	// Execute test
	result, err := adapter.EvaluatePermission(ctx, req)

	// Verify results
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "failed to evaluate permission via ABAC")
}

func TestGetUserEffectivePermissions_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockABAC := abac.NewMockService(ctrl)
	mockAccess := access.NewMockService(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockMetrics := metrics.NewMockMetricsProvider(ctrl)
	mockTracer := tracing.NewMockTracingService(ctrl)

	// Setup common mock expectations
	mockSpan := tracing.NewMockSpan(ctrl)
	mockSpan.EXPECT().End().AnyTimes()
	mockSpan.EXPECT().SetAttributes(gomock.Any()).AnyTimes()
	mockTracer.EXPECT().StartSpan(gomock.Any(), gomock.Any()).
		Return(context.Background(), mockSpan).AnyTimes()
	mockMetrics.EXPECT().IncrementCounter(gomock.Any(), gomock.Any()).AnyTimes()

	// Create adapter
	adapter := NewAdapter(mockABAC, mockAccess, mockLogger, mockMetrics, mockTracer)

	// Test data
	ctx := context.Background()
	userID := uuid.New()
	entityID := uuid.New()

	abacResult := &abac.UserEffectivePermissions{
		UserID:   userID,
		EntityID: &entityID,
		Permissions: map[string][]string{
			"document": {"read", "write"},
			"file":     {"read"},
		},
		Roles:     []string{"user", "editor"},
		Timestamp: time.Now(),
	}

	// Setup ABAC mock expectation
	mockABAC.EXPECT().GetUserEffectivePermissions(gomock.Any(), gomock.Eq(userID), gomock.Eq(&entityID)).
		Return(abacResult, nil)

	// Execute test
	result, err := adapter.GetUserEffectivePermissions(ctx, userID, &entityID)

	// Verify results
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, userID, result.UserID)
	require.Equal(t, &entityID, result.EntityID)
	require.Len(t, result.Permissions, 2) // 2 resource types
	require.Len(t, result.Roles, 2)       // 2 roles
}

func TestInvalidateUserCache_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockABAC := abac.NewMockService(ctrl)
	mockAccess := access.NewMockService(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockMetrics := metrics.NewMockMetricsProvider(ctrl)
	mockTracer := tracing.NewMockTracingService(ctrl)

	// Setup common mock expectations
	mockSpan := tracing.NewMockSpan(ctrl)
	mockSpan.EXPECT().End().AnyTimes()
	mockSpan.EXPECT().SetAttributes(gomock.Any()).AnyTimes()
	mockTracer.EXPECT().StartSpan(gomock.Any(), gomock.Any()).
		Return(context.Background(), mockSpan).AnyTimes()
	mockMetrics.EXPECT().IncrementCounter(gomock.Any(), gomock.Any()).AnyTimes()

	// Create adapter
	adapter := NewAdapter(mockABAC, mockAccess, mockLogger, mockMetrics, mockTracer)

	// Test data
	ctx := context.Background()
	userID := uuid.New()

	// Setup ABAC mock expectation
	mockABAC.EXPECT().InvalidateUserCache(gomock.Any(), gomock.Eq(userID)).
		Return(nil)

	// Execute test
	err := adapter.InvalidateUserCache(ctx, userID)

	// Verify results
	require.NoError(t, err)
}
