package abac

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// Performance test configuration
const (
	// Test durations
	ShortTestDuration  = 10 * time.Second
	MediumTestDuration = 30 * time.Second
	LongTestDuration   = 60 * time.Second

	// Concurrency levels
	LowConcurrency    = 10
	MediumConcurrency = 50
	HighConcurrency   = 100

	// Performance targets
	MaxEvaluationLatency = 10 * time.Millisecond
	MaxAttributeLatency  = 5 * time.Millisecond
	MinThroughput        = 1000 // evaluations per second
	MaxErrorRate         = 0.01 // 1%
)

// PerformanceTestSuite provides comprehensive performance testing
type PerformanceTestSuite struct {
	ctx context.Context

	// ABAC components
	policyManager      PolicyManager
	evaluationEngine   PolicyEvaluationEngine
	attributeService   AttributeService
	attributeCollector AttributeCollector
	attributeResolver  AttributeResolver
	performanceOpt     PerformanceOptimizer

	// Test infrastructure
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.TracingService

	// Test data
	testPolicies   []*models.Policy
	testUsers      []uuid.UUID
	testResources  []uuid.UUID
	testAttributes map[string]interface{}
}

// SetupPerformanceTestSuite initializes the performance test environment
func SetupPerformanceTestSuite() *PerformanceTestSuite {
	suite := &PerformanceTestSuite{
		ctx: context.Background(),
	}

	// Initialize test infrastructure
	suite.logger = logger.NewMockLogger()
	suite.metrics = metrics.NewMockMetricsProvider()
	suite.tracer = tracing.NewMockTracingService()

	// Initialize ABAC components with performance-optimized mocks
	suite.setupABACComponents()

	// Create test data
	suite.createTestData()

	return suite
}

func (suite *PerformanceTestSuite) setupABACComponents() {
	// Performance-optimized mock repositories
	policyRepo := &MockPolicyRepository{}
	attributeRepo := &MockAttributeRepository{}

	// Setup policy repository with batch operations
	suite.setupPolicyRepoMocks(policyRepo)

	// Setup attribute repository with caching
	suite.setupAttributeRepoMocks(attributeRepo)

	// Initialize ABAC services
	suite.attributeService = NewAttributeService(
		attributeRepo,
		[]byte("performance-test-key-32-bytes!"),
		suite.logger,
		suite.metrics,
		suite.tracer,
	)

	suite.attributeCollector = NewAttributeCollector(
		attributeRepo,
		nil, // sourceRepo
		suite.logger,
		suite.metrics,
		suite.tracer,
	)

	suite.attributeResolver = NewAttributeResolver(
		attributeRepo,
		suite.attributeCollector,
		suite.logger,
		suite.metrics,
		suite.tracer,
	)

	suite.evaluationEngine = NewPolicyEvaluationEngine(
		policyRepo,
		suite.attributeResolver,
		NewExternalAttributeSourceManager(nil, suite.logger, suite.metrics, suite.tracer),
		suite.logger,
		suite.metrics,
		suite.tracer,
	)

	suite.policyManager = NewPolicyManager(
		policyRepo,
		suite.evaluationEngine,
		suite.logger,
		suite.metrics,
		suite.tracer,
	)

	suite.performanceOpt = NewPerformanceOptimizer(
		policyRepo,
		nil, // evaluationRepo
		attributeRepo,
		suite.logger,
		suite.metrics,
		suite.tracer,
	)
}

func (suite *PerformanceTestSuite) setupPolicyRepoMocks(mockRepo *MockPolicyRepository) {
	// Mock fast policy retrieval
	mockRepo.On("GetPolicyByID", mock.Anything, mock.Anything).Return(
		func(ctx context.Context, id uuid.UUID) *models.Policy {
			for _, policy := range suite.testPolicies {
				if policy.ID == id {
					return policy
				}
			}
			return &models.Policy{
				ID:                 id,
				Name:               "Performance Test Policy",
				Status:             types.PolicyStatusActive,
				Effect:             types.PolicyEffectPermit,
				CombiningAlgorithm: types.CombiningAlgorithmDenyOverrides,
				Rules: []*models.PolicyRule{
					{
						ID:     uuid.New(),
						Effect: types.PolicyEffectPermit,
						Condition: &models.PolicyCondition{
							Expression: "subject.department == 'engineering'",
						},
					},
				},
			}
		},
		nil,
	)

	// Mock batch policy operations
	mockRepo.On("ListPolicies", mock.Anything, mock.Anything).Return(
		func(ctx context.Context, req *ListPoliciesRequest) *PolicyListResult {
			return &PolicyListResult{
				Policies:   suite.testPolicies,
				TotalCount: len(suite.testPolicies),
				Page:       req.Page,
				PageSize:   req.PageSize,
			}
		},
		nil,
	)
}

func (suite *PerformanceTestSuite) setupAttributeRepoMocks(mockRepo *MockAttributeRepository) {
	// Mock fast attribute definition retrieval
	mockRepo.On("GetAttributeDefinition", mock.Anything, mock.Anything).Return(
		&AttributeDefinitionDetails{
			ID:           uuid.New(),
			Name:         "department",
			DataType:     types.AttributeDataTypeString,
			Category:     types.AttributeCategoryUser,
			IsRequired:   true,
			IsMultiValue: false,
		},
		nil,
	)

	// Mock attribute definition listing
	mockRepo.On("ListAttributeDefinitions", mock.Anything, mock.Anything).Return(
		&AttributeDefinitionListResult{
			Definitions: []*AttributeDefinitionResult{
				{
					ID:       uuid.New(),
					Name:     "department",
					DataType: types.AttributeDataTypeString,
					Category: types.AttributeCategoryUser,
				},
				{
					ID:       uuid.New(),
					Name:     "clearance_level",
					DataType: types.AttributeDataTypeString,
					Category: types.AttributeCategoryUser,
				},
			},
			TotalCount: 2,
		},
		nil,
	)
}

func (suite *PerformanceTestSuite) createTestData() {
	// Create test policies
	suite.testPolicies = make([]*models.Policy, 100)
	for i := 0; i < 100; i++ {
		suite.testPolicies[i] = &models.Policy{
			ID:                 uuid.New(),
			Name:               fmt.Sprintf("Policy %d", i),
			Status:             types.PolicyStatusActive,
			Effect:             types.PolicyEffectPermit,
			CombiningAlgorithm: types.CombiningAlgorithmDenyOverrides,
			Priority:           i,
			Rules: []*models.PolicyRule{
				{
					ID:     uuid.New(),
					Effect: types.PolicyEffectPermit,
					Condition: &models.PolicyCondition{
						Expression: fmt.Sprintf("subject.department == 'dept_%d'", i%10),
					},
				},
			},
		}
	}

	// Create test users
	suite.testUsers = make([]uuid.UUID, 1000)
	for i := 0; i < 1000; i++ {
		suite.testUsers[i] = uuid.New()
	}

	// Create test resources
	suite.testResources = make([]uuid.UUID, 1000)
	for i := 0; i < 1000; i++ {
		suite.testResources[i] = uuid.New()
	}

	// Create test attributes
	suite.testAttributes = map[string]interface{}{
		"department":      "engineering",
		"clearance_level": "standard",
		"role":            "developer",
		"location":        "office",
		"shift":           "day",
	}
}

// TestPolicyEvaluationPerformance tests policy evaluation performance
func TestPolicyEvaluationPerformance(t *testing.T) {
	suite := SetupPerformanceTestSuite()

	t.Run("SingleEvaluationLatency", func(t *testing.T) {
		testSingleEvaluationLatency(t, suite)
	})

	t.Run("ConcurrentEvaluationThroughput", func(t *testing.T) {
		testConcurrentEvaluationThroughput(t, suite)
	})

	t.Run("BatchEvaluationPerformance", func(t *testing.T) {
		testBatchEvaluationPerformance(t, suite)
	})

	t.Run("PolicyCachePerformance", func(t *testing.T) {
		testPolicyCachePerformance(t, suite)
	})
}

func testSingleEvaluationLatency(t *testing.T, suite *PerformanceTestSuite) {
	policyID := suite.testPolicies[0].ID
	userID := suite.testUsers[0]
	resourceID := suite.testResources[0]

	request := &PolicyEvaluationRequest{
		RequestID: uuid.New(),
		PolicyID:  policyID,
		Subject: SubjectContext{
			UserID:     userID,
			Roles:      []string{"developer"},
			Attributes: suite.testAttributes,
		},
		Resource: ResourceContext{
			ResourceID:   resourceID,
			ResourceType: "document",
			Attributes: map[string]interface{}{
				"classification": "internal",
			},
		},
		Action: ActionContext{
			Action: "read",
		},
		Environment: EnvironmentContext{
			Timestamp: time.Now(),
		},
		EvaluationMode: EvaluationModeStandard,
	}

	// Warm up
	for i := 0; i < 10; i++ {
		_, _ = suite.evaluationEngine.EvaluatePolicy(suite.ctx, request)
	}

	// Measure latency
	iterations := 1000
	start := time.Now()

	for i := 0; i < iterations; i++ {
		result, err := suite.evaluationEngine.EvaluatePolicy(suite.ctx, request)
		require.NoError(t, err)
		assert.NotNil(t, result)
	}

	elapsed := time.Since(start)
	avgLatency := elapsed / time.Duration(iterations)

	t.Logf("Single evaluation average latency: %v", avgLatency)
	t.Logf("Evaluations per second: %.0f", float64(iterations)/elapsed.Seconds())

	// Assert performance target
	assert.Less(t, avgLatency, MaxEvaluationLatency,
		"Single evaluation latency exceeds target of %v", MaxEvaluationLatency)
}

func testConcurrentEvaluationThroughput(t *testing.T, suite *PerformanceTestSuite) {
	concurrencyLevels := []int{LowConcurrency, MediumConcurrency, HighConcurrency}

	for _, concurrency := range concurrencyLevels {
		t.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(t *testing.T) {
			testConcurrentThroughput(t, suite, concurrency, ShortTestDuration)
		})
	}
}

func testConcurrentThroughput(t *testing.T, suite *PerformanceTestSuite, concurrency int, duration time.Duration) {
	var (
		totalEvaluations int64
		totalErrors      int64
		totalLatency     int64
		wg               sync.WaitGroup
		ctx, cancel      = context.WithTimeout(suite.ctx, duration)
	)
	defer cancel()

	// Start workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			var (
				evaluations int64
				errors      int64
				latency     int64
			)

			for {
				select {
				case <-ctx.Done():
					atomic.AddInt64(&totalEvaluations, evaluations)
					atomic.AddInt64(&totalErrors, errors)
					atomic.AddInt64(&totalLatency, latency)
					return
				default:
					// Create evaluation request
					policyID := suite.testPolicies[workerID%len(suite.testPolicies)].ID
					userID := suite.testUsers[workerID%len(suite.testUsers)]
					resourceID := suite.testResources[workerID%len(suite.testResources)]

					request := &PolicyEvaluationRequest{
						RequestID: uuid.New(),
						PolicyID:  policyID,
						Subject: SubjectContext{
							UserID:     userID,
							Roles:      []string{"developer"},
							Attributes: suite.testAttributes,
						},
						Resource: ResourceContext{
							ResourceID:   resourceID,
							ResourceType: "document",
						},
						Action: ActionContext{
							Action: "read",
						},
						Environment: EnvironmentContext{
							Timestamp: time.Now(),
						},
					}

					// Measure evaluation
					start := time.Now()
					_, err := suite.evaluationEngine.EvaluatePolicy(ctx, request)
					elapsed := time.Since(start)

					evaluations++
					latency += elapsed.Nanoseconds()

					if err != nil {
						errors++
					}
				}
			}
		}(i)
	}

	wg.Wait()

	// Calculate metrics
	throughput := float64(totalEvaluations) / duration.Seconds()
	avgLatency := time.Duration(totalLatency / totalEvaluations)
	errorRate := float64(totalErrors) / float64(totalEvaluations)

	t.Logf("Concurrency: %d", concurrency)
	t.Logf("Total evaluations: %d", totalEvaluations)
	t.Logf("Throughput: %.2f evaluations/second", throughput)
	t.Logf("Average latency: %v", avgLatency)
	t.Logf("Error rate: %.2f%%", errorRate*100)

	// Assert performance targets
	assert.Greater(t, throughput, float64(MinThroughput),
		"Throughput below target of %d evaluations/second", MinThroughput)
	assert.Less(t, avgLatency, MaxEvaluationLatency,
		"Average latency exceeds target of %v", MaxEvaluationLatency)
	assert.Less(t, errorRate, MaxErrorRate,
		"Error rate exceeds target of %.2f%%", MaxErrorRate*100)
}

func testBatchEvaluationPerformance(t *testing.T, suite *PerformanceTestSuite) {
	batchSizes := []int{10, 50, 100, 500}

	for _, batchSize := range batchSizes {
		t.Run(fmt.Sprintf("BatchSize_%d", batchSize), func(t *testing.T) {
			// Create batch request
			evaluations := make([]PolicyEvaluationRequest, batchSize)
			for i := 0; i < batchSize; i++ {
				evaluations[i] = PolicyEvaluationRequest{
					RequestID: uuid.New(),
					PolicyID:  suite.testPolicies[i%len(suite.testPolicies)].ID,
					Subject: SubjectContext{
						UserID:     suite.testUsers[i%len(suite.testUsers)],
						Roles:      []string{"developer"},
						Attributes: suite.testAttributes,
					},
					Resource: ResourceContext{
						ResourceID:   suite.testResources[i%len(suite.testResources)],
						ResourceType: "document",
					},
					Action: ActionContext{
						Action: "read",
					},
				}
			}

			request := &BatchEvaluationRequest{
				RequestID:   uuid.New(),
				Evaluations: evaluations,
				Options: BatchEvaluationOptions{
					Parallel:    true,
					MaxWorkers:  10,
					EnableCache: true,
				},
			}

			// Measure batch evaluation
			start := time.Now()
			result, err := suite.performanceOpt.BatchEvaluate(suite.ctx, request)
			elapsed := time.Since(start)

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Len(t, result.Results, batchSize)

			avgLatencyPerEval := elapsed / time.Duration(batchSize)
			throughput := float64(batchSize) / elapsed.Seconds()

			t.Logf("Batch size: %d", batchSize)
			t.Logf("Total time: %v", elapsed)
			t.Logf("Average latency per evaluation: %v", avgLatencyPerEval)
			t.Logf("Throughput: %.2f evaluations/second", throughput)

			// Batch should be more efficient than individual evaluations
			assert.Less(t, avgLatencyPerEval, MaxEvaluationLatency,
				"Batch evaluation average latency exceeds target")
		})
	}
}

func testPolicyCachePerformance(t *testing.T, suite *PerformanceTestSuite) {
	policyID := suite.testPolicies[0].ID

	// Test cache warm-up
	start := time.Now()
	for i := 0; i < 100; i++ {
		request := &PolicyCompilationRequest{
			PolicyID: policyID,
			Options: map[string]interface{}{
				"cache_compiled": true,
			},
		}
		_, err := suite.performanceOpt.CompilePolicy(suite.ctx, request)
		require.NoError(t, err)
	}
	warmupTime := time.Since(start)

	// Test cache hits
	start = time.Now()
	for i := 0; i < 1000; i++ {
		cacheRequest := &CacheRequest{
			Key: fmt.Sprintf("policy_cache_test_%d", i%10),
			Parameters: map[string]interface{}{
				"policy_id": policyID.String(),
			},
		}
		_, _ = suite.performanceOpt.GetCachedEvaluation(suite.ctx, cacheRequest)
	}
	cacheTime := time.Since(start)

	t.Logf("Cache warm-up time (100 operations): %v", warmupTime)
	t.Logf("Cache access time (1000 operations): %v", cacheTime)
	t.Logf("Average cache access time: %v", cacheTime/1000)

	// Cache should be significantly faster
	avgCacheTime := cacheTime / 1000
	assert.Less(t, avgCacheTime, time.Millisecond,
		"Cache access time should be under 1ms")
}

// TestAttributeResolutionPerformance tests attribute resolution performance
func TestAttributeResolutionPerformance(t *testing.T) {
	suite := SetupPerformanceTestSuite()

	t.Run("SingleAttributeResolution", func(t *testing.T) {
		testSingleAttributeResolution(t, suite)
	})

	t.Run("BulkAttributeResolution", func(t *testing.T) {
		testBulkAttributeResolution(t, suite)
	})

	t.Run("AttributeCachePerformance", func(t *testing.T) {
		testAttributeCachePerformance(t, suite)
	})

	t.Run("ConcurrentAttributeResolution", func(t *testing.T) {
		testConcurrentAttributeResolution(t, suite)
	})
}

func testSingleAttributeResolution(t *testing.T, suite *PerformanceTestSuite) {
	userID := suite.testUsers[0]

	request := &AttributeResolutionRequest{
		RequestID:    uuid.New(),
		TargetType:   AttributeTargetTypeUser,
		TargetID:     userID,
		AttributeIDs: []string{"department", "clearance_level", "role"},
		Options: AttributeResolutionOptions{
			IncludeDerived: true,
			MaxAge:         5 * time.Minute,
		},
	}

	// Warm up
	for i := 0; i < 10; i++ {
		_, _ = suite.attributeResolver.ResolveAttributes(suite.ctx, request)
	}

	// Measure latency
	iterations := 1000
	start := time.Now()

	for i := 0; i < iterations; i++ {
		result, err := suite.attributeResolver.ResolveAttributes(suite.ctx, request)
		require.NoError(t, err)
		assert.NotNil(t, result)
	}

	elapsed := time.Since(start)
	avgLatency := elapsed / time.Duration(iterations)

	t.Logf("Single attribute resolution average latency: %v", avgLatency)
	t.Logf("Resolutions per second: %.0f", float64(iterations)/elapsed.Seconds())

	// Assert performance target
	assert.Less(t, avgLatency, MaxAttributeLatency,
		"Attribute resolution latency exceeds target of %v", MaxAttributeLatency)
}

func testBulkAttributeResolution(t *testing.T, suite *PerformanceTestSuite) {
	batchSizes := []int{10, 50, 100}

	for _, batchSize := range batchSizes {
		t.Run(fmt.Sprintf("BatchSize_%d", batchSize), func(t *testing.T) {
			userIDs := suite.testUsers[:batchSize]

			start := time.Now()
			for _, userID := range userIDs {
				request := &AttributeResolutionRequest{
					RequestID:    uuid.New(),
					TargetType:   AttributeTargetTypeUser,
					TargetID:     userID,
					AttributeIDs: []string{"department", "clearance_level"},
				}

				_, err := suite.attributeResolver.ResolveAttributes(suite.ctx, request)
				require.NoError(t, err)
			}
			elapsed := time.Since(start)

			avgLatencyPerUser := elapsed / time.Duration(batchSize)
			throughput := float64(batchSize) / elapsed.Seconds()

			t.Logf("Batch size: %d", batchSize)
			t.Logf("Total time: %v", elapsed)
			t.Logf("Average latency per user: %v", avgLatencyPerUser)
			t.Logf("Throughput: %.2f users/second", throughput)

			assert.Less(t, avgLatencyPerUser, MaxAttributeLatency*2,
				"Bulk attribute resolution average latency exceeds target")
		})
	}
}

func testAttributeCachePerformance(t *testing.T, suite *PerformanceTestSuite) {
	userID := suite.testUsers[0]

	// Test cache miss (first access)
	request := &AttributeResolutionRequest{
		RequestID:    uuid.New(),
		TargetType:   AttributeTargetTypeUser,
		TargetID:     userID,
		AttributeIDs: []string{"department", "clearance_level"},
		Options: AttributeResolutionOptions{
			EnableCaching: true,
		},
	}

	start := time.Now()
	_, err := suite.attributeResolver.ResolveAttributes(suite.ctx, request)
	require.NoError(t, err)
	cacheMissTime := time.Since(start)

	// Test cache hit (subsequent accesses)
	start = time.Now()
	for i := 0; i < 100; i++ {
		_, err := suite.attributeResolver.ResolveAttributes(suite.ctx, request)
		require.NoError(t, err)
	}
	cacheHitTime := time.Since(start) / 100

	t.Logf("Cache miss time: %v", cacheMissTime)
	t.Logf("Cache hit average time: %v", cacheHitTime)
	t.Logf("Cache speedup: %.2fx", float64(cacheMissTime)/float64(cacheHitTime))

	// Cache hits should be significantly faster
	assert.Less(t, cacheHitTime, cacheMissTime/5,
		"Cache hits should be at least 5x faster than cache misses")
}

func testConcurrentAttributeResolution(t *testing.T, suite *PerformanceTestSuite) {
	concurrency := MediumConcurrency
	duration := ShortTestDuration

	var (
		totalResolutions int64
		totalErrors      int64
		wg               sync.WaitGroup
		ctx, cancel      = context.WithTimeout(suite.ctx, duration)
	)
	defer cancel()

	// Start workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			var resolutions, errors int64

			for {
				select {
				case <-ctx.Done():
					atomic.AddInt64(&totalResolutions, resolutions)
					atomic.AddInt64(&totalErrors, errors)
					return
				default:
					userID := suite.testUsers[workerID%len(suite.testUsers)]
					request := &AttributeResolutionRequest{
						RequestID:    uuid.New(),
						TargetType:   AttributeTargetTypeUser,
						TargetID:     userID,
						AttributeIDs: []string{"department", "clearance_level"},
					}

					_, err := suite.attributeResolver.ResolveAttributes(ctx, request)
					resolutions++
					if err != nil {
						errors++
					}
				}
			}
		}(i)
	}

	wg.Wait()

	throughput := float64(totalResolutions) / duration.Seconds()
	errorRate := float64(totalErrors) / float64(totalResolutions)

	t.Logf("Concurrent attribute resolution results:")
	t.Logf("Concurrency: %d", concurrency)
	t.Logf("Total resolutions: %d", totalResolutions)
	t.Logf("Throughput: %.2f resolutions/second", throughput)
	t.Logf("Error rate: %.2f%%", errorRate*100)

	assert.Greater(t, throughput, float64(MinThroughput)/2,
		"Attribute resolution throughput below target")
	assert.Less(t, errorRate, MaxErrorRate,
		"Attribute resolution error rate exceeds target")
}

// TestMemoryUsageAndLeaks tests memory usage patterns and potential leaks
func TestMemoryUsageAndLeaks(t *testing.T) {
	suite := SetupPerformanceTestSuite()

	t.Run("MemoryUsageUnderLoad", func(t *testing.T) {
		testMemoryUsageUnderLoad(t, suite)
	})

	t.Run("GarbageCollectionImpact", func(t *testing.T) {
		testGarbageCollectionImpact(t, suite)
	})
}

func testMemoryUsageUnderLoad(t *testing.T, suite *PerformanceTestSuite) {
	// Force garbage collection before test
	runtime.GC()
	runtime.GC() // Call twice to ensure cleanup

	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Run sustained load
	iterations := 10000
	for i := 0; i < iterations; i++ {
		policyID := suite.testPolicies[i%len(suite.testPolicies)].ID
		userID := suite.testUsers[i%len(suite.testUsers)]
		resourceID := suite.testResources[i%len(suite.testResources)]

		request := &PolicyEvaluationRequest{
			RequestID: uuid.New(),
			PolicyID:  policyID,
			Subject: SubjectContext{
				UserID:     userID,
				Attributes: suite.testAttributes,
			},
			Resource: ResourceContext{
				ResourceID:   resourceID,
				ResourceType: "document",
			},
			Action: ActionContext{
				Action: "read",
			},
		}

		_, err := suite.evaluationEngine.EvaluatePolicy(suite.ctx, request)
		require.NoError(t, err)

		// Trigger GC every 1000 iterations
		if i%1000 == 0 {
			runtime.GC()
		}
	}

	// Force garbage collection after test
	runtime.GC()
	runtime.GC()

	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	memoryGrowth := m2.Alloc - m1.Alloc
	heapGrowth := m2.HeapAlloc - m1.HeapAlloc

	t.Logf("Memory usage after %d evaluations:", iterations)
	t.Logf("Memory growth: %d bytes (%.2f MB)", memoryGrowth, float64(memoryGrowth)/(1024*1024))
	t.Logf("Heap growth: %d bytes (%.2f MB)", heapGrowth, float64(heapGrowth)/(1024*1024))
	t.Logf("GC cycles: %d", m2.NumGC-m1.NumGC)

	// Memory growth should be reasonable (less than 100MB for 10k evaluations)
	maxMemoryGrowth := int64(100 * 1024 * 1024) // 100MB
	assert.Less(t, int64(memoryGrowth), maxMemoryGrowth,
		"Memory growth exceeds threshold, possible memory leak")
}

func testGarbageCollectionImpact(t *testing.T, suite *PerformanceTestSuite) {
	policyID := suite.testPolicies[0].ID
	userID := suite.testUsers[0]
	resourceID := suite.testResources[0]

	request := &PolicyEvaluationRequest{
		RequestID: uuid.New(),
		PolicyID:  policyID,
		Subject: SubjectContext{
			UserID:     userID,
			Attributes: suite.testAttributes,
		},
		Resource: ResourceContext{
			ResourceID:   resourceID,
			ResourceType: "document",
		},
		Action: ActionContext{
			Action: "read",
		},
	}

	// Measure latency without GC pressure
	var latenciesNoGC []time.Duration
	for i := 0; i < 100; i++ {
		start := time.Now()
		_, err := suite.evaluationEngine.EvaluatePolicy(suite.ctx, request)
		require.NoError(t, err)
		latenciesNoGC = append(latenciesNoGC, time.Since(start))
	}

	// Create GC pressure
	garbage := make([][]byte, 1000)
	for i := range garbage {
		garbage[i] = make([]byte, 1024*1024) // 1MB each
	}

	// Measure latency with GC pressure
	var latenciesWithGC []time.Duration
	for i := 0; i < 100; i++ {
		start := time.Now()
		_, err := suite.evaluationEngine.EvaluatePolicy(suite.ctx, request)
		require.NoError(t, err)
		latenciesWithGC = append(latenciesWithGC, time.Since(start))

		// Modify garbage to maintain GC pressure
		if i%10 == 0 {
			garbage[i%len(garbage)] = make([]byte, 1024*1024)
		}
	}

	// Calculate averages
	var avgNoGC, avgWithGC time.Duration
	for _, lat := range latenciesNoGC {
		avgNoGC += lat
	}
	avgNoGC /= time.Duration(len(latenciesNoGC))

	for _, lat := range latenciesWithGC {
		avgWithGC += lat
	}
	avgWithGC /= time.Duration(len(latenciesWithGC))

	gcImpact := float64(avgWithGC) / float64(avgNoGC)

	t.Logf("Average latency without GC pressure: %v", avgNoGC)
	t.Logf("Average latency with GC pressure: %v", avgWithGC)
	t.Logf("GC impact factor: %.2fx", gcImpact)

	// GC impact should be minimal (less than 2x slowdown)
	assert.Less(t, gcImpact, 2.0,
		"GC pressure causes excessive performance degradation")

	// Keep garbage reference to prevent optimization
	_ = garbage[0][0]
}

// TestScalabilityLimits tests system behavior at scale limits
func TestScalabilityLimits(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping scalability tests in short mode")
	}

	suite := SetupPerformanceTestSuite()

	t.Run("MaxConcurrencyTest", func(t *testing.T) {
		testMaxConcurrency(t, suite)
	})

	t.Run("LongDurationStabilityTest", func(t *testing.T) {
		testLongDurationStability(t, suite)
	})
}

func testMaxConcurrency(t *testing.T, suite *PerformanceTestSuite) {
	maxConcurrency := 500 // Stress test level
	duration := MediumTestDuration

	var (
		totalEvaluations int64
		totalErrors      int64
		wg               sync.WaitGroup
		ctx, cancel      = context.WithTimeout(suite.ctx, duration)
	)
	defer cancel()

	t.Logf("Testing maximum concurrency: %d workers for %v", maxConcurrency, duration)

	// Start workers
	for i := 0; i < maxConcurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			var evaluations, errors int64

			for {
				select {
				case <-ctx.Done():
					atomic.AddInt64(&totalEvaluations, evaluations)
					atomic.AddInt64(&totalErrors, errors)
					return
				default:
					policyID := suite.testPolicies[workerID%len(suite.testPolicies)].ID
					userID := suite.testUsers[workerID%len(suite.testUsers)]
					resourceID := suite.testResources[workerID%len(suite.testResources)]

					request := &PolicyEvaluationRequest{
						RequestID: uuid.New(),
						PolicyID:  policyID,
						Subject: SubjectContext{
							UserID:     userID,
							Attributes: suite.testAttributes,
						},
						Resource: ResourceContext{
							ResourceID:   resourceID,
							ResourceType: "document",
						},
						Action: ActionContext{
							Action: "read",
						},
					}

					_, err := suite.evaluationEngine.EvaluatePolicy(ctx, request)
					evaluations++
					if err != nil {
						errors++
					}

					// Brief pause to prevent overwhelming
					time.Sleep(time.Microsecond * 100)
				}
			}
		}(i)
	}

	wg.Wait()

	throughput := float64(totalEvaluations) / duration.Seconds()
	errorRate := float64(totalErrors) / float64(totalEvaluations)

	t.Logf("Maximum concurrency test results:")
	t.Logf("Workers: %d", maxConcurrency)
	t.Logf("Duration: %v", duration)
	t.Logf("Total evaluations: %d", totalEvaluations)
	t.Logf("Throughput: %.2f evaluations/second", throughput)
	t.Logf("Error rate: %.2f%%", errorRate*100)

	// System should handle high concurrency gracefully
	assert.Greater(t, totalEvaluations, int64(1000),
		"System should handle some evaluations under max concurrency")
	assert.Less(t, errorRate, 0.05, // 5% error rate acceptable under stress
		"Error rate too high under maximum concurrency")
}

func testLongDurationStability(t *testing.T, suite *PerformanceTestSuite) {
	duration := LongTestDuration
	concurrency := MediumConcurrency

	var (
		totalEvaluations int64
		totalErrors      int64
		latencySum       int64
		wg               sync.WaitGroup
		ctx, cancel      = context.WithTimeout(suite.ctx, duration)
	)
	defer cancel()

	t.Logf("Testing long-duration stability: %d workers for %v", concurrency, duration)

	// Start workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			var evaluations, errors, workerLatencySum int64

			for {
				select {
				case <-ctx.Done():
					atomic.AddInt64(&totalEvaluations, evaluations)
					atomic.AddInt64(&totalErrors, errors)
					atomic.AddInt64(&latencySum, workerLatencySum)
					return
				default:
					policyID := suite.testPolicies[workerID%len(suite.testPolicies)].ID
					userID := suite.testUsers[workerID%len(suite.testUsers)]
					resourceID := suite.testResources[workerID%len(suite.testResources)]

					request := &PolicyEvaluationRequest{
						RequestID: uuid.New(),
						PolicyID:  policyID,
						Subject: SubjectContext{
							UserID:     userID,
							Attributes: suite.testAttributes,
						},
						Resource: ResourceContext{
							ResourceID:   resourceID,
							ResourceType: "document",
						},
						Action: ActionContext{
							Action: "read",
						},
					}

					start := time.Now()
					_, err := suite.evaluationEngine.EvaluatePolicy(ctx, request)
					elapsed := time.Since(start)

					evaluations++
					workerLatencySum += elapsed.Nanoseconds()
					if err != nil {
						errors++
					}

					// Small delay to simulate realistic load
					time.Sleep(time.Millisecond)
				}
			}
		}(i)
	}

	wg.Wait()

	throughput := float64(totalEvaluations) / duration.Seconds()
	avgLatency := time.Duration(latencySum / totalEvaluations)
	errorRate := float64(totalErrors) / float64(totalEvaluations)

	t.Logf("Long-duration stability test results:")
	t.Logf("Duration: %v", duration)
	t.Logf("Workers: %d", concurrency)
	t.Logf("Total evaluations: %d", totalEvaluations)
	t.Logf("Throughput: %.2f evaluations/second", throughput)
	t.Logf("Average latency: %v", avgLatency)
	t.Logf("Error rate: %.2f%%", errorRate*100)

	// System should remain stable over long periods
	assert.Greater(t, throughput, float64(MinThroughput)/2,
		"Throughput degraded significantly over time")
	assert.Less(t, avgLatency, MaxEvaluationLatency*2,
		"Average latency increased significantly over time")
	assert.Less(t, errorRate, MaxErrorRate*2,
		"Error rate increased significantly over time")
}

// Benchmark functions for continuous performance monitoring
func BenchmarkPolicyEvaluation(b *testing.B) {
	suite := SetupPerformanceTestSuite()

	policyID := suite.testPolicies[0].ID
	userID := suite.testUsers[0]
	resourceID := suite.testResources[0]

	request := &PolicyEvaluationRequest{
		RequestID: uuid.New(),
		PolicyID:  policyID,
		Subject: SubjectContext{
			UserID:     userID,
			Attributes: suite.testAttributes,
		},
		Resource: ResourceContext{
			ResourceID:   resourceID,
			ResourceType: "document",
		},
		Action: ActionContext{
			Action: "read",
		},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := suite.evaluationEngine.EvaluatePolicy(suite.ctx, request)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkAttributeResolution(b *testing.B) {
	suite := SetupPerformanceTestSuite()

	userID := suite.testUsers[0]
	request := &AttributeResolutionRequest{
		RequestID:    uuid.New(),
		TargetType:   AttributeTargetTypeUser,
		TargetID:     userID,
		AttributeIDs: []string{"department", "clearance_level", "role"},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := suite.attributeResolver.ResolveAttributes(suite.ctx, request)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkBatchEvaluation(b *testing.B) {
	suite := SetupPerformanceTestSuite()

	// Create batch request with 10 evaluations
	evaluations := make([]PolicyEvaluationRequest, 10)
	for i := 0; i < 10; i++ {
		evaluations[i] = PolicyEvaluationRequest{
			RequestID: uuid.New(),
			PolicyID:  suite.testPolicies[i%len(suite.testPolicies)].ID,
			Subject: SubjectContext{
				UserID:     suite.testUsers[i%len(suite.testUsers)],
				Attributes: suite.testAttributes,
			},
			Resource: ResourceContext{
				ResourceID:   suite.testResources[i%len(suite.testResources)],
				ResourceType: "document",
			},
			Action: ActionContext{
				Action: "read",
			},
		}
	}

	request := &BatchEvaluationRequest{
		RequestID:   uuid.New(),
		Evaluations: evaluations,
		Options: BatchEvaluationOptions{
			Parallel:   true,
			MaxWorkers: 4,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := suite.performanceOpt.BatchEvaluate(suite.ctx, request)
		if err != nil {
			b.Fatal(err)
		}
	}
}
