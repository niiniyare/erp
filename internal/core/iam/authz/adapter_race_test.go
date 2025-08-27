package authz

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/niiniyare/erp/internal/core/abac"
	"github.com/niiniyare/erp/internal/core/abac/activities"
	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/core/access"
	"github.com/niiniyare/erp/internal/core/iam/model"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// RaceConditionTestSuite tests concurrent operations across multiple tenants
type RaceConditionTestSuite struct {
	suite.Suite
	ctrl        *gomock.Controller
	mockABAC    *abac.MockService
	mockAccess  *access.MockService
	mockLogger  *logger.MockLogger
	mockMetrics *metrics.MockMetricsProvider
	mockTracer  *tracing.MockTracingService
	mockSpan    *tracing.MockSpan
	adapter     Service
	ctx         context.Context

	// Test data for 5 tenants
	tenantIDs []uuid.UUID
	userIDs   []uuid.UUID
}

func TestRaceConditionTestSuite(t *testing.T) {
	suite.Run(t, new(RaceConditionTestSuite))
}

func (s *RaceConditionTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockABAC = abac.NewMockService(s.ctrl)
	s.mockAccess = access.NewMockService(s.ctrl)
	s.mockLogger = logger.NewMockLogger(s.ctrl)
	s.mockMetrics = metrics.NewMockMetricsProvider(s.ctrl)
	s.mockTracer = tracing.NewMockTracingService(s.ctrl)
	s.mockSpan = tracing.NewMockSpan(s.ctrl)

	// Setup very permissive mock expectations to avoid race conditions with call counting
	s.mockSpan.EXPECT().End().AnyTimes()
	s.mockSpan.EXPECT().SetAttributes(gomock.Any()).AnyTimes()
	s.mockTracer.EXPECT().StartSpan(gomock.Any(), gomock.Any()).
		Return(context.Background(), s.mockSpan).AnyTimes()
	s.mockMetrics.EXPECT().ObserveHistogram(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	s.mockMetrics.EXPECT().IncrementCounter(gomock.Any(), gomock.Any()).AnyTimes()
	s.mockLogger.EXPECT().InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	s.mockTracer.EXPECT().RecordError(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	s.adapter = NewAdapter(s.mockABAC, s.mockAccess, s.mockLogger, s.mockMetrics, s.mockTracer)
	s.ctx = context.Background()

	// Initialize test data for 5 tenants
	s.tenantIDs = make([]uuid.UUID, 5)
	s.userIDs = make([]uuid.UUID, 5)
	for i := 0; i < 5; i++ {
		s.tenantIDs[i] = uuid.New()
		s.userIDs[i] = uuid.New()
	}
}

func (s *RaceConditionTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

// ─── CONCURRENT PERMISSION EVALUATION TESTS ───────────────────────────────

func (s *RaceConditionTestSuite) TestConcurrentPermissionEvaluation() {
	// Setup ABAC mock to return successful results
	s.mockABAC.EXPECT().EvaluatePermission(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, req *abac.PermissionEvaluationRequest) (*abac.PermissionEvaluationResult, error) {
			return &abac.PermissionEvaluationResult{
				Decision: types.PolicyDecisionAllow,
				PolicyDecisions: []*models.PolicyDecision{
					{
						PolicyID:      uuid.New(),
						Decision:      types.PolicyDecisionAllow,
						Reason:        "Concurrent test policy",
						MatchedRule:   "test_rule",
						EvaluationMS:  10,
						TargetMatched: true,
					},
				},
				EvaluationTimeMS: 25,
				CacheHit:         false,
				RequestID:        req.RequestID,
				Timestamp:        time.Now(),
			}, nil
		}).AnyTimes()

	const numGoroutines = 100
	const numTenants = 5

	// Use channels to collect results and errors
	resultChan := make(chan *PermissionEvaluationResult, numGoroutines)
	errorChan := make(chan error, numGoroutines)

	var wg sync.WaitGroup
	startTime := time.Now()

	// Launch 100 concurrent goroutines across 5 tenants
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			tenantIndex := i % numTenants
			userID := s.userIDs[tenantIndex]

			req := &PermissionEvaluationRequest{
				UserID:       userID,
				ResourceType: "document",
				ResourceID:   func() *uuid.UUID { id := uuid.New(); return &id }(),
				Action:       "read",
				EntityID:     &s.tenantIDs[tenantIndex],
				Context:      map[string]any{"goroutine_id": i, "tenant_index": tenantIndex},
				RequestID:    uuid.New().String(),
			}

			result, err := s.adapter.EvaluatePermission(s.ctx, req)
			if err != nil {
				errorChan <- err
			} else {
				resultChan <- result
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(resultChan)
	close(errorChan)

	duration := time.Since(startTime)
	s.T().Logf("Concurrent permission evaluation completed in %v", duration)

	// Collect and verify results
	var results []*PermissionEvaluationResult
	var errors []error

	for result := range resultChan {
		results = append(results, result)
	}

	for err := range errorChan {
		errors = append(errors, err)
	}

	// Assert no errors occurred
	s.Empty(errors, "No errors should occur during concurrent execution")

	// Assert all requests were processed
	s.Len(results, numGoroutines, "All permission evaluations should complete successfully")

	// Verify each result
	for _, result := range results {
		s.NotNil(result, "Result should not be nil")
		s.Equal(model.PolicyDecisionAllow, result.Decision, "All evaluations should be allowed")
		s.NotEmpty(result.RequestID, "Request ID should be present")
	}

	s.T().Logf("Successfully processed %d concurrent permission evaluations across %d tenants", len(results), numTenants)
}

func (s *RaceConditionTestSuite) TestConcurrentBulkEvaluation() {
	// Setup ABAC mock for bulk evaluation
	s.mockABAC.EXPECT().BulkEvaluatePermissions(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, req *abac.BulkPermissionEvaluationRequest) (*abac.BulkPermissionEvaluationResult, error) {
			results := make([]*abac.PermissionEvaluationResult, len(req.Requests))
			for i := range req.Requests {
				results[i] = &abac.PermissionEvaluationResult{
					Decision:         types.PolicyDecisionAllow,
					EvaluationTimeMS: 15,
					CacheHit:         i%2 == 0, // Alternate cache hits
					RequestID:        req.Requests[i].RequestID,
					Timestamp:        time.Now(),
				}
			}

			return &abac.BulkPermissionEvaluationResult{
				Results:         results,
				TotalRequests:   len(req.Requests),
				SuccessfulCount: len(req.Requests),
				FailedCount:     0,
				TotalTimeMS:     int64(len(req.Requests) * 15),
				AverageTimeMS:   15,
				RequestID:       req.RequestID,
				Timestamp:       time.Now(),
			}, nil
		}).AnyTimes()

	const numGoroutines = 50
	const requestsPerBulk = 5
	const numTenants = 5

	resultChan := make(chan *BulkPermissionEvaluationResult, numGoroutines)
	errorChan := make(chan error, numGoroutines)

	var wg sync.WaitGroup
	startTime := time.Now()

	// Launch concurrent bulk evaluations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			tenantIndex := i % numTenants
			userID := s.userIDs[tenantIndex]
			entityID := s.tenantIDs[tenantIndex]

			// Create bulk request with multiple sub-requests
			requests := make([]*PermissionEvaluationRequest, requestsPerBulk)
			for j := 0; j < requestsPerBulk; j++ {
				requests[j] = &PermissionEvaluationRequest{
					UserID:       userID,
					ResourceType: "file",
					ResourceID:   func() *uuid.UUID { id := uuid.New(); return &id }(),
					Action:       []string{"read", "write", "delete"}[j%3],
					EntityID:     &entityID,
					RequestID:    uuid.New().String(),
				}
			}

			bulkReq := &BulkPermissionEvaluationRequest{
				Requests:  requests,
				RequestID: uuid.New().String(),
			}

			result, err := s.adapter.BulkEvaluatePermissions(s.ctx, bulkReq)
			if err != nil {
				errorChan <- err
			} else {
				resultChan <- result
			}
		}(i)
	}

	wg.Wait()
	close(resultChan)
	close(errorChan)

	duration := time.Since(startTime)
	s.T().Logf("Concurrent bulk evaluation completed in %v", duration)

	// Collect results
	var results []*BulkPermissionEvaluationResult
	var errors []error

	for result := range resultChan {
		results = append(results, result)
	}

	for err := range errorChan {
		errors = append(errors, err)
	}

	s.Empty(errors, "No errors should occur during concurrent bulk evaluation")
	s.Len(results, numGoroutines, "All bulk evaluations should complete")

	totalRequestsProcessed := 0
	for _, result := range results {
		s.Equal(requestsPerBulk, result.TotalRequests, "Each bulk request should contain expected number of requests")
		s.Equal(requestsPerBulk, result.SuccessfulCount, "All requests should be successful")
		s.Equal(0, result.FailedCount, "No requests should fail")
		totalRequestsProcessed += result.TotalRequests
	}

	s.T().Logf("Successfully processed %d bulk evaluations (%d total requests) across %d tenants",
		len(results), totalRequestsProcessed, numTenants)
}

// ─── CONCURRENT CACHE OPERATIONS TESTS ────────────────────────────────────

func (s *RaceConditionTestSuite) TestConcurrentCacheOperations() {
	// Setup ABAC mocks for cache operations
	s.mockABAC.EXPECT().InvalidateUserCache(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	s.mockABAC.EXPECT().InvalidatePolicyCache(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	s.mockABAC.EXPECT().GetCacheStatistics(gomock.Any()).
		Return(&activities.GetCacheStatsOutput{
			PolicyEvaluationStats: nil, // Simplified for test
			AttributeStats:        nil, // Simplified for test
			GeneratedAt:           time.Now(),
		}, nil).AnyTimes()

	const numOperations = 100
	const numTenants = 5

	errorChan := make(chan error, numOperations*3) // 3 types of operations
	var wg sync.WaitGroup

	startTime := time.Now()

	// Concurrent user cache invalidations
	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			tenantIndex := i % numTenants
			userID := s.userIDs[tenantIndex]

			err := s.adapter.InvalidateUserCache(s.ctx, userID)
			if err != nil {
				errorChan <- err
			}
		}(i)
	}

	// Concurrent policy cache invalidations
	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			policyIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

			err := s.adapter.InvalidatePolicyCache(s.ctx, policyIDs)
			if err != nil {
				errorChan <- err
			}
		}(i)
	}

	// Concurrent cache statistics requests
	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			_, err := s.adapter.GetCacheStatistics(s.ctx)
			if err != nil {
				errorChan <- err
			}
		}(i)
	}

	wg.Wait()
	close(errorChan)

	duration := time.Since(startTime)
	s.T().Logf("Concurrent cache operations completed in %v", duration)

	// Check for errors
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	s.Empty(errors, "No errors should occur during concurrent cache operations")
	s.T().Logf("Successfully processed %d concurrent cache operations across %d tenants",
		numOperations*3, numTenants)
}

// ─── CONCURRENT ACCESS REQUEST OPERATIONS ─────────────────────────────────

func (s *RaceConditionTestSuite) TestConcurrentAccessRequestOperations() {
	// Setup access service mocks
	s.mockAccess.EXPECT().CreateAccessRequest(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, req *access.CreateAccessRequestRequest) (*access.AccessRequest, error) {
			return &access.AccessRequest{
				ID:            uuid.New(),
				TenantID:      s.tenantIDs[0], // Use first tenant for simplicity
				RequesterID:   req.UserID,
				TargetUserID:  &req.UserID,
				EntityID:      s.tenantIDs[0],
				RequestType:   access.RequestType("RESOURCE_ACCESS"),
				ResourceID:    func() *uuid.UUID { id := uuid.New(); return &id }(),
				Justification: req.Justification,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}, nil
		}).AnyTimes()

	s.mockAccess.EXPECT().ProcessAccessRequest(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	s.mockAccess.EXPECT().ListAccessRequests(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, req *access.ListAccessRequestsRequest) (*access.ListAccessRequestsResult, error) {
			// Return empty list for simplicity
			return &access.ListAccessRequestsResult{
				Requests: []*access.AccessRequest{},
				Total:    0,
				Limit:    req.Limit,
				Offset:   req.Offset,
				HasMore:  false,
			}, nil
		}).AnyTimes()

	const numOperations = 50
	const numTenants = 5

	var wg sync.WaitGroup
	errorChan := make(chan error, numOperations*3)
	var createdRequestIDs sync.Map // Thread-safe map for storing created request IDs

	startTime := time.Now()

	// Concurrent access request creation
	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			tenantIndex := i % numTenants
			userID := s.userIDs[tenantIndex]

			req := &CreateAccessRequestRequest{
				UserID:        userID,
				ResourceType:  "document",
				ResourceID:    func() *uuid.UUID { id := uuid.New(); return &id }(),
				Action:        "read",
				Justification: "Concurrent test access request",
				Duration:      func() *time.Duration { d := time.Hour * 24; return &d }(),
				Priority:      "medium",
				Metadata:      map[string]any{"test_id": i},
			}

			result, err := s.adapter.CreateAccessRequest(s.ctx, req)
			if err != nil {
				errorChan <- err
			} else {
				createdRequestIDs.Store(i, result.ID)
			}
		}(i)
	}

	// Wait for creation to complete before processing
	wg.Wait()

	// Concurrent access request processing
	createdRequestIDs.Range(func(key, value any) bool {
		wg.Add(1)
		go func(requestID uuid.UUID) {
			defer wg.Done()

			processReq := &ProcessAccessRequestRequest{
				RequestID:  requestID,
				Action:     "approve",
				ApproverID: s.userIDs[0], // Use first user as approver
				Comments:   "Auto-approved for concurrent test",
				Conditions: []string{"time_limited"},
			}

			err := s.adapter.ProcessAccessRequest(s.ctx, processReq)
			if err != nil {
				errorChan <- err
			}
		}(value.(uuid.UUID))
		return true
	})

	// Concurrent list operations
	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			tenantIndex := i % numTenants
			userID := s.userIDs[tenantIndex]
			entityID := s.tenantIDs[tenantIndex]

			listReq := &ListAccessRequestsRequest{
				UserID:   &userID,
				Status:   func() *string { s := "pending"; return &s }(),
				EntityID: &entityID,
				Limit:    10,
				Offset:   0,
			}

			_, err := s.adapter.ListAccessRequests(s.ctx, listReq)
			if err != nil {
				errorChan <- err
			}
		}(i)
	}

	wg.Wait()
	close(errorChan)

	duration := time.Since(startTime)
	s.T().Logf("Concurrent access request operations completed in %v", duration)

	// Check for errors
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	s.Empty(errors, "No errors should occur during concurrent access request operations")
	s.T().Logf("Successfully processed concurrent access request operations across %d tenants", numTenants)
}

// ─── MIXED WORKLOAD RACE TEST ─────────────────────────────────────────────

func (s *RaceConditionTestSuite) TestMixedWorkloadRaceConditions() {
	// Setup all mocks for mixed workload
	s.setupAllMocks()

	const totalOperations = 200
	const numTenants = 5

	var wg sync.WaitGroup
	errorChan := make(chan error, totalOperations)
	operationTypes := make(chan string, totalOperations)

	startTime := time.Now()

	// Launch mixed operations across all adapter methods
	for i := 0; i < totalOperations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			tenantIndex := i % numTenants
			userID := s.userIDs[tenantIndex]
			entityID := s.tenantIDs[tenantIndex]

			// Randomly choose operation type based on index
			switch i % 8 {
			case 0, 1: // Permission evaluation (most common)
				operationTypes <- "permission_eval"
				req := &PermissionEvaluationRequest{
					UserID:       userID,
					ResourceType: "document",
					Action:       "read",
					EntityID:     &entityID,
					RequestID:    uuid.New().String(),
				}
				_, err := s.adapter.EvaluatePermission(s.ctx, req)
				if err != nil {
					errorChan <- err
				}

			case 2: // User effective permissions
				operationTypes <- "user_permissions"
				_, err := s.adapter.GetUserEffectivePermissions(s.ctx, userID, &entityID)
				if err != nil {
					errorChan <- err
				}

			case 3: // Cache operations
				operationTypes <- "cache_invalidate"
				err := s.adapter.InvalidateUserCache(s.ctx, userID)
				if err != nil {
					errorChan <- err
				}

			case 4: // Grant permission
				operationTypes <- "grant_permission"
				req := &GrantPermissionRequest{
					UserID:       userID,
					ResourceType: "file",
					Action:       "write",
					EntityID:     &entityID,
				}
				err := s.adapter.GrantPermission(s.ctx, req)
				if err != nil {
					errorChan <- err
				}

			case 5: // Access request creation
				operationTypes <- "create_access_request"
				req := &CreateAccessRequestRequest{
					UserID:        userID,
					ResourceType:  "document",
					Action:        "read",
					Justification: "Mixed workload test",
				}
				_, err := s.adapter.CreateAccessRequest(s.ctx, req)
				if err != nil {
					errorChan <- err
				}

			case 6: // Decision history
				operationTypes <- "decision_history"
				_, err := s.adapter.GetDecisionHistory(s.ctx, userID, 5)
				if err != nil {
					errorChan <- err
				}

			case 7: // Cache statistics
				operationTypes <- "cache_stats"
				_, err := s.adapter.GetCacheStatistics(s.ctx)
				if err != nil {
					errorChan <- err
				}
			}
		}(i)
	}

	wg.Wait()
	close(errorChan)
	close(operationTypes)

	duration := time.Since(startTime)
	s.T().Logf("Mixed workload race test completed in %v", duration)

	// Collect and analyze results
	var errors []error
	operationCounts := make(map[string]int)

	for err := range errorChan {
		errors = append(errors, err)
	}

	for opType := range operationTypes {
		operationCounts[opType]++
	}

	s.Empty(errors, "No errors should occur during mixed workload execution")

	// Count total operations processed
	totalProcessed := 0
	for _, count := range operationCounts {
		totalProcessed += count
	}
	s.Equal(totalOperations, totalProcessed, "All operations should be tracked")

	s.T().Logf("Mixed workload completed successfully:")
	for opType, count := range operationCounts {
		s.T().Logf("  %s: %d operations", opType, count)
	}
	s.T().Logf("Total operations across %d tenants: %d", numTenants, totalOperations)
}

// setupAllMocks configures all mock services for mixed workload testing
func (s *RaceConditionTestSuite) setupAllMocks() {
	// ABAC mocks
	s.mockABAC.EXPECT().EvaluatePermission(gomock.Any(), gomock.Any()).
		Return(&abac.PermissionEvaluationResult{
			Decision:         types.PolicyDecisionAllow,
			EvaluationTimeMS: 20,
			CacheHit:         false,
			RequestID:        uuid.New().String(),
			Timestamp:        time.Now(),
		}, nil).AnyTimes()

	s.mockABAC.EXPECT().GetUserEffectivePermissions(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&abac.UserEffectivePermissions{
			UserID:      uuid.New(),
			Permissions: map[string][]string{"document": {"read"}},
			Roles:       []string{"user"},
			Timestamp:   time.Now(),
		}, nil).AnyTimes()

	s.mockABAC.EXPECT().InvalidateUserCache(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	s.mockABAC.EXPECT().GetDecisionHistory(gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]*abac.DecisionHistoryEntry{}, nil).AnyTimes()

	s.mockABAC.EXPECT().GetCacheStatistics(gomock.Any()).
		Return(&activities.GetCacheStatsOutput{
			PolicyEvaluationStats: nil, // Simplified for race test
			AttributeStats:        nil, // Simplified for race test
			GeneratedAt:           time.Now(),
		}, nil).AnyTimes()

	// Access service mocks
	s.mockAccess.EXPECT().GrantPermission(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	s.mockAccess.EXPECT().CreateAccessRequest(gomock.Any(), gomock.Any()).
		Return(&access.AccessRequest{
			ID:          uuid.New(),
			TenantID:    s.tenantIDs[0],
			RequesterID: uuid.New(),
			EntityID:    s.tenantIDs[0],
			RequestType: access.RequestType("RESOURCE_ACCESS"),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}, nil).AnyTimes()
}

// TestRaceDetection runs the full race condition test with Go's race detector
func TestRaceDetection(t *testing.T) {
	// This test is designed to be run with: go test -race
	// It performs a simple concurrent test that should pass race detection
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockABAC := abac.NewMockService(ctrl)
	mockAccess := access.NewMockService(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockMetrics := metrics.NewMockMetricsProvider(ctrl)
	mockTracer := tracing.NewMockTracingService(ctrl)
	mockSpan := tracing.NewMockSpan(ctrl)

	// Permissive mock setup
	mockSpan.EXPECT().End().AnyTimes()
	mockSpan.EXPECT().SetAttributes(gomock.Any()).AnyTimes()
	mockTracer.EXPECT().StartSpan(gomock.Any(), gomock.Any()).
		Return(context.Background(), mockSpan).AnyTimes()
	mockMetrics.EXPECT().ObserveHistogram(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockMetrics.EXPECT().IncrementCounter(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockABAC.EXPECT().EvaluatePermission(gomock.Any(), gomock.Any()).
		Return(&abac.PermissionEvaluationResult{
			Decision:         types.PolicyDecisionAllow,
			EvaluationTimeMS: 10,
			CacheHit:         false,
			RequestID:        "race-test",
			Timestamp:        time.Now(),
		}, nil).AnyTimes()

	adapter := NewAdapter(mockABAC, mockAccess, mockLogger, mockMetrics, mockTracer)
	ctx := context.Background()

	// Simple concurrent test for race detection
	const numGoroutines = 50
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			req := &PermissionEvaluationRequest{
				UserID:       uuid.New(),
				ResourceType: "document",
				Action:       "read",
				RequestID:    uuid.New().String(),
			}

			result, err := adapter.EvaluatePermission(ctx, req)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, model.PolicyDecisionAllow, result.Decision)
		}(i)
	}

	wg.Wait()
	t.Logf("Race detection test completed successfully with %d concurrent operations", numGoroutines)
}
