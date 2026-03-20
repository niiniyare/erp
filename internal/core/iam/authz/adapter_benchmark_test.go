package authz

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"awo/internal/core/abac"
	"awo/internal/core/abac/models"
	"awo/internal/core/access"
	"awo/internal/shared/logger"
	"awo/internal/shared/metrics"
	"awo/internal/shared/tracing"
	"awo/internal/shared/types"
)

// BenchmarkEvaluatePermission benchmarks single permission evaluation
func BenchmarkEvaluatePermission(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	// Setup mocks
	mockABAC := abac.NewMockService(ctrl)
	mockAccess := access.NewMockService(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockMetrics := metrics.NewMockMetricsProvider(ctrl)
	mockTracer := tracing.NewMockService(ctrl)
	mockSpan := tracing.NewMockSpan(ctrl)

	// Setup lenient mock expectations for benchmarking
	mockSpan.EXPECT().End().AnyTimes()
	mockTracer.EXPECT().StartSpan(gomock.Any(), gomock.Any()).
		Return(context.Background(), mockSpan).AnyTimes()
	mockMetrics.EXPECT().ObserveHistogram(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockMetrics.EXPECT().IncrementCounter(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	// Setup ABAC mock to return fast successful results
	mockABAC.EXPECT().EvaluatePermission(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, req *abac.PermissionEvaluationRequest) (*abac.PermissionEvaluationResult, error) {
			return &abac.PermissionEvaluationResult{
				Decision: types.PolicyDecisionAllow,
				PolicyDecisions: []*models.PolicyDecision{
					{
						PolicyID:      uuid.New(),
						Decision:      types.PolicyDecisionAllow,
						Reason:        "Benchmark policy",
						MatchedRule:   "benchmark_rule",
						EvaluationMS:  1,
						TargetMatched: true,
					},
				},
				EvaluationTimeMS: 5,
				CacheHit:         false,
				RequestID:        req.RequestID,
				Timestamp:        time.Now(),
			}, nil
		}).AnyTimes()

	adapter := NewAdapter(mockABAC, mockAccess, mockLogger, mockMetrics, mockTracer)
	ctx := context.Background()

	// Prepare test request
	req := &PermissionEvaluationRequest{
		UserID:       uuid.New(),
		ResourceType: "document",
		ResourceID:   func() *uuid.UUID { id := uuid.New(); return &id }(),
		Action:       "read",
		EntityID:     func() *uuid.UUID { id := uuid.New(); return &id }(),
		Context:      map[string]any{"department": "engineering"},
		RequestID:    "benchmark-request",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req.RequestID = "benchmark-request-" + string(rune(i))
		result, err := adapter.EvaluatePermission(ctx, req)
		if err != nil {
			b.Fatalf("EvaluatePermission failed: %v", err)
		}
		if result == nil {
			b.Fatal("Expected non-nil result")
		}
	}
}

// BenchmarkBulkEvaluatePermissions benchmarks bulk permission evaluation
func BenchmarkBulkEvaluatePermissions(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	// Setup mocks
	mockABAC := abac.NewMockService(ctrl)
	mockAccess := access.NewMockService(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockMetrics := metrics.NewMockMetricsProvider(ctrl)
	mockTracer := tracing.NewMockService(ctrl)
	mockSpan := tracing.NewMockSpan(ctrl)

	// Setup lenient mock expectations
	mockSpan.EXPECT().End().AnyTimes()
	mockTracer.EXPECT().StartSpan(gomock.Any(), gomock.Any()).
		Return(context.Background(), mockSpan).AnyTimes()
	mockMetrics.EXPECT().ObserveHistogram(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockMetrics.EXPECT().IncrementCounter(gomock.Any(), gomock.Any()).AnyTimes()

	// Setup ABAC mock for bulk operations
	mockABAC.EXPECT().BulkEvaluatePermissions(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, req *abac.BulkPermissionEvaluationRequest) (*abac.BulkPermissionEvaluationResult, error) {
			results := make([]*abac.PermissionEvaluationResult, len(req.Requests))
			for i, subReq := range req.Requests {
				results[i] = &abac.PermissionEvaluationResult{
					Decision:         types.PolicyDecisionAllow,
					EvaluationTimeMS: 3,
					CacheHit:         i%2 == 0, // Alternate cache hits for realism
					RequestID:        subReq.RequestID,
					Timestamp:        time.Now(),
				}
			}

			return &abac.BulkPermissionEvaluationResult{
				Results:         results,
				TotalRequests:   len(req.Requests),
				SuccessfulCount: len(req.Requests),
				FailedCount:     0,
				TotalTimeMS:     int64(len(req.Requests) * 3),
				AverageTimeMS:   3.0,
				RequestID:       req.RequestID,
				Timestamp:       time.Now(),
			}, nil
		}).AnyTimes()

	adapter := NewAdapter(mockABAC, mockAccess, mockLogger, mockMetrics, mockTracer)
	ctx := context.Background()

	// Test different bulk sizes
	bulkSizes := []int{5, 10, 20, 50}

	for _, size := range bulkSizes {
		b.Run("size_"+string(rune(size+'0')), func(b *testing.B) {
			// Prepare bulk request
			requests := make([]*PermissionEvaluationRequest, size)
			for i := 0; i < size; i++ {
				requests[i] = &PermissionEvaluationRequest{
					UserID:       uuid.New(),
					ResourceType: "document",
					ResourceID:   func() *uuid.UUID { id := uuid.New(); return &id }(),
					Action:       []string{"read", "write", "delete"}[i%3],
					EntityID:     func() *uuid.UUID { id := uuid.New(); return &id }(),
					RequestID:    "bulk-" + string(rune(i)),
				}
			}

			bulkReq := &BulkPermissionEvaluationRequest{
				Requests:  requests,
				RequestID: "benchmark-bulk",
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				bulkReq.RequestID = "benchmark-bulk-" + string(rune(i))
				result, err := adapter.BulkEvaluatePermissions(ctx, bulkReq)
				if err != nil {
					b.Fatalf("BulkEvaluatePermissions failed: %v", err)
				}
				if result == nil {
					b.Fatal("Expected non-nil result")
				}
				if len(result.Results) != size {
					b.Fatalf("Expected %d results, got %d", size, len(result.Results))
				}
			}
		})
	}
}

// BenchmarkConcurrentPermissionEvaluation benchmarks concurrent permission evaluations
func BenchmarkConcurrentPermissionEvaluation(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	// Setup mocks with very lenient expectations for high concurrency
	mockABAC := abac.NewMockService(ctrl)
	mockAccess := access.NewMockService(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockMetrics := metrics.NewMockMetricsProvider(ctrl)
	mockTracer := tracing.NewMockService(ctrl)
	mockSpan := tracing.NewMockSpan(ctrl)

	// Extremely permissive mocks for concurrent benchmarking
	mockSpan.EXPECT().End().AnyTimes()
	mockTracer.EXPECT().StartSpan(gomock.Any(), gomock.Any()).
		Return(context.Background(), mockSpan).AnyTimes()
	mockMetrics.EXPECT().ObserveHistogram(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockMetrics.EXPECT().IncrementCounter(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockABAC.EXPECT().EvaluatePermission(gomock.Any(), gomock.Any()).
		Return(&abac.PermissionEvaluationResult{
			Decision:         types.PolicyDecisionAllow,
			EvaluationTimeMS: 2,
			CacheHit:         true,
			RequestID:        "concurrent-bench",
			Timestamp:        time.Now(),
		}, nil).AnyTimes()

	adapter := NewAdapter(mockABAC, mockAccess, mockLogger, mockMetrics, mockTracer)
	ctx := context.Background()

	// Test different concurrency levels
	concurrencyLevels := []int{1, 5, 10, 20}

	for _, concurrency := range concurrencyLevels {
		b.Run("concurrency_"+string(rune(concurrency+'0')), func(b *testing.B) {
			b.ResetTimer()
			b.SetParallelism(concurrency)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					req := &PermissionEvaluationRequest{
						UserID:       uuid.New(),
						ResourceType: "document",
						Action:       "read",
						RequestID:    "concurrent-benchmark",
					}

					result, err := adapter.EvaluatePermission(ctx, req)
					if err != nil {
						b.Errorf("EvaluatePermission failed: %v", err)
					}
					if result == nil {
						b.Error("Expected non-nil result")
					}
				}
			})
		})
	}
}

// BenchmarkCacheOperations benchmarks cache-related operations
func BenchmarkCacheOperations(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	// Setup mocks
	mockABAC := abac.NewMockService(ctrl)
	mockAccess := access.NewMockService(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockMetrics := metrics.NewMockMetricsProvider(ctrl)
	mockTracer := tracing.NewMockService(ctrl)
	mockSpan := tracing.NewMockSpan(ctrl)

	mockSpan.EXPECT().End().AnyTimes()
	mockTracer.EXPECT().StartSpan(gomock.Any(), gomock.Any()).
		Return(context.Background(), mockSpan).AnyTimes()
	mockMetrics.EXPECT().IncrementCounter(gomock.Any(), gomock.Any()).AnyTimes()

	// Mock cache operations to return quickly
	mockABAC.EXPECT().InvalidateUserCache(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockABAC.EXPECT().InvalidatePolicyCache(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	adapter := NewAdapter(mockABAC, mockAccess, mockLogger, mockMetrics, mockTracer)
	ctx := context.Background()

	b.Run("InvalidateUserCache", func(b *testing.B) {
		userID := uuid.New()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			err := adapter.InvalidateUserCache(ctx, userID)
			if err != nil {
				b.Fatalf("InvalidateUserCache failed: %v", err)
			}
		}
	})

	b.Run("InvalidatePolicyCache", func(b *testing.B) {
		policyIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			err := adapter.InvalidatePolicyCache(ctx, policyIDs)
			if err != nil {
				b.Fatalf("InvalidatePolicyCache failed: %v", err)
			}
		}
	})
}

// BenchmarkAccessRequestOperations benchmarks access request operations
func BenchmarkAccessRequestOperations(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	// Setup mocks
	mockABAC := abac.NewMockService(ctrl)
	mockAccess := access.NewMockService(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockMetrics := metrics.NewMockMetricsProvider(ctrl)
	mockTracer := tracing.NewMockService(ctrl)
	mockSpan := tracing.NewMockSpan(ctrl)

	mockSpan.EXPECT().End().AnyTimes()
	mockTracer.EXPECT().StartSpan(gomock.Any(), gomock.Any()).
		Return(context.Background(), mockSpan).AnyTimes()
	mockMetrics.EXPECT().IncrementCounter(gomock.Any(), gomock.Any()).AnyTimes()

	// Mock access request operations
	mockAccess.EXPECT().CreateAccessRequest(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, req *access.CreateAccessRequestRequest) (*access.AccessRequest, error) {
			return &access.AccessRequest{
				ID:            uuid.New(),
				TenantID:      uuid.New(),
				RequesterID:   req.UserID,
				TargetUserID:  &req.UserID,
				EntityID:      uuid.New(),
				RequestType:   access.RequestType("RESOURCE_ACCESS"),
				ResourceID:    func() *uuid.UUID { id := uuid.New(); return &id }(),
				Justification: req.Justification,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}, nil
		}).AnyTimes()

	mockAccess.EXPECT().ProcessAccessRequest(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	adapter := NewAdapter(mockABAC, mockAccess, mockLogger, mockMetrics, mockTracer)
	ctx := context.Background()

	b.Run("CreateAccessRequest", func(b *testing.B) {
		userID := uuid.New()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			req := &CreateAccessRequestRequest{
				UserID:        userID,
				ResourceType:  "document",
				ResourceID:    func() *uuid.UUID { id := uuid.New(); return &id }(),
				Action:        "read",
				Justification: "Benchmark test access",
				Duration:      func() *time.Duration { d := time.Hour * 24; return &d }(),
				Priority:      "medium",
			}

			result, err := adapter.CreateAccessRequest(ctx, req)
			if err != nil {
				b.Fatalf("CreateAccessRequest failed: %v", err)
			}
			if result == nil {
				b.Fatal("Expected non-nil result")
			}
		}
	})

	b.Run("ProcessAccessRequest", func(b *testing.B) {
		requestID := uuid.New()
		approverID := uuid.New()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			req := &ProcessAccessRequestRequest{
				RequestID:  requestID,
				Action:     "approve",
				ApproverID: approverID,
				Comments:   "Benchmark approval",
				Conditions: []string{"time_limited"},
			}

			err := adapter.ProcessAccessRequest(ctx, req)
			if err != nil {
				b.Fatalf("ProcessAccessRequest failed: %v", err)
			}
		}
	})
}

// BenchmarkMemoryUsage benchmarks memory allocation patterns
func BenchmarkMemoryUsage(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	// Setup minimal mocks
	mockABAC := abac.NewMockService(ctrl)
	mockAccess := access.NewMockService(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockMetrics := metrics.NewMockMetricsProvider(ctrl)
	mockTracer := tracing.NewMockService(ctrl)
	mockSpan := tracing.NewMockSpan(ctrl)

	mockSpan.EXPECT().End().AnyTimes()
	mockTracer.EXPECT().StartSpan(gomock.Any(), gomock.Any()).
		Return(context.Background(), mockSpan).AnyTimes()
	mockMetrics.EXPECT().ObserveHistogram(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockMetrics.EXPECT().IncrementCounter(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockABAC.EXPECT().EvaluatePermission(gomock.Any(), gomock.Any()).
		Return(&abac.PermissionEvaluationResult{
			Decision:         types.PolicyDecisionAllow,
			PolicyDecisions:  []*models.PolicyDecision{},
			EvaluationTimeMS: 1,
			CacheHit:         true,
			RequestID:        "memory-bench",
			Timestamp:        time.Now(),
		}, nil).AnyTimes()

	adapter := NewAdapter(mockABAC, mockAccess, mockLogger, mockMetrics, mockTracer)
	ctx := context.Background()

	b.Run("RequestCreation", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = &PermissionEvaluationRequest{
				UserID:       uuid.New(),
				ResourceType: "document",
				ResourceID:   func() *uuid.UUID { id := uuid.New(); return &id }(),
				Action:       "read",
				EntityID:     func() *uuid.UUID { id := uuid.New(); return &id }(),
				Context:      map[string]any{"key": "value"},
				RequestID:    "memory-test",
			}
		}
	})

	b.Run("FullEvaluationCycle", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			req := &PermissionEvaluationRequest{
				UserID:       uuid.New(),
				ResourceType: "document",
				Action:       "read",
				RequestID:    "memory-cycle-test",
			}

			result, err := adapter.EvaluatePermission(ctx, req)
			if err != nil {
				b.Fatalf("EvaluatePermission failed: %v", err)
			}

			// Force the result to be used to prevent optimization
			_ = result.Decision
		}
	})
}
