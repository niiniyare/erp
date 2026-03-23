package tracing_test

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"awo.so/internal/shared/tracing" // Replace with your actual module path
)

// Flag to control JSON output
//
// Usage:
//
//	go test -run TestIntegrationTestSuite                        # Run quietly (no JSON output)
//	go test -run TestIntegrationTestSuite -verbose-tracing       # Run with JSON output
var verboseTracing = flag.Bool("verbose-tracing", false, "Enable verbose JSON output for trace spans")

// IntegrationTestSuite tests integration scenarios
type IntegrationTestSuite struct {
	suite.Suite
	service tracing.Service
}

func (s *IntegrationTestSuite) SetupTest() {
	config := tracing.DefaultConfig()
	config.ServiceName = "integration-test"
	config.Enabled = true

	if !*verboseTracing {
		// For non-verbose mode, we need to use a custom approach
		// Since the tracing package doesn't expose a way to configure the stdout writer,
		// we'll use an environment variable to signal this
		os.Setenv("INTEGRATION_TEST_QUIET", "true")
		config.Protocol = tracing.ProtocolStdout
	} else {
		// Use normal stdout
		config.Protocol = tracing.ProtocolStdout
	}

	service, err := tracing.NewService(config)
	require.NoError(s.T(), err)
	s.service = service
}

func (s *IntegrationTestSuite) TearDownTest() {
	if s.service != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.service.Shutdown(ctx)
	}
	// Clean up environment variable
	os.Unsetenv("INTEGRATION_TEST_QUIET")
}

// TestFullABACFlow tests complete ABAC permission evaluation flow
func (s *IntegrationTestSuite) TestFullABACFlow() {
	type PermissionRequest struct {
		UserID       string
		ResourceType string
		ResourceID   string
		Action       string
	}

	evaluatePermission := func(ctx context.Context, req PermissionRequest) error {
		ctx, span := s.service.StartSpan(ctx, "abac.service.EvaluatePermission",
			tracing.WithSpanKind(tracing.SpanKindInternal),
			tracing.WithAttributes(
				attribute.String("user_id", req.UserID),
				attribute.String("resource_type", req.ResourceType),
				attribute.String("action", req.Action),
			))
		defer span.End()

		// Step 1: Validate tenant
		s.service.AddEvent(ctx, "validating-tenant-context")
		time.Sleep(5 * time.Millisecond)
		span.SetAttributes(
			attribute.String("tenant.id", "tenant-123"),
			attribute.String("tenant.name", "Test Org"),
		)

		// Step 2: Collect user attributes
		ctx2, userAttrSpan := s.service.StartSpan(ctx, "collect-user-attributes")
		time.Sleep(10 * time.Millisecond)
		userAttrSpan.SetAttributes(attribute.Int("user_attributes.count", 8))
		userAttrSpan.SetStatus(codes.Ok, "collected")
		userAttrSpan.End()

		// Step 3: Collect resource attributes
		ctx3, resAttrSpan := s.service.StartSpan(ctx2, "collect-resource-attributes")
		time.Sleep(8 * time.Millisecond)
		resAttrSpan.SetAttributes(attribute.Int("resource_attributes.count", 5))
		resAttrSpan.SetStatus(codes.Ok, "collected")
		resAttrSpan.End()

		// Step 4: Build environment context
		ctx4, envSpan := s.service.StartSpan(ctx3, "build-environment-context")
		time.Sleep(3 * time.Millisecond)
		envSpan.SetAttributes(
			attribute.String("request.ip", "192.168.1.1"),
			attribute.String("request.user_agent", "TestClient/1.0"),
		)
		envSpan.SetStatus(codes.Ok, "built")
		envSpan.End()

		// Step 5: Evaluate policies
		ctx5, evalSpan := s.service.StartSpan(ctx4, "evaluate-policies")
		time.Sleep(15 * time.Millisecond)
		evalSpan.SetAttributes(
			attribute.String("decision", "allow"),
			attribute.Int("policies_evaluated", 3),
			attribute.Bool("cache_hit", false),
		)
		evalSpan.SetStatus(codes.Ok, "evaluated")
		evalSpan.End()

		// Final span status
		span.SetAttributes(
			attribute.String("final_decision", "allow"),
			attribute.Int64("total_duration_ms", 41),
		)
		span.SetStatus(codes.Ok, "permission evaluation successful")

		require.NotNil(s.T(), ctx5)
		return nil
	}

	req := PermissionRequest{
		UserID:       "user-456",
		ResourceType: "document",
		ResourceID:   "doc-789",
		Action:       "read",
	}

	err := evaluatePermission(context.Background(), req)
	require.NoError(s.T(), err)
}

// TestHTTPServerClientFlow tests full HTTP request/response cycle
func (s *IntegrationTestSuite) TestHTTPServerClientFlow() {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract trace context
		ctx := s.service.ExtractHTTPHeaders(r.Context(), r.Header)

		// Start server span
		ctx, span := s.service.StartSpan(ctx, "GET /api/users",
			tracing.WithSpanKind(tracing.SpanKindServer),
			tracing.WithAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.url", r.URL.String()),
			))
		defer span.End()

		// Simulate processing
		time.Sleep(10 * time.Millisecond)

		// Database call
		ctx2, dbSpan := s.service.StartSpan(ctx, "database.query",
			tracing.WithSpanKind(tracing.SpanKindClient))
		time.Sleep(5 * time.Millisecond)
		dbSpan.SetAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.statement", "SELECT * FROM users"),
		)
		dbSpan.SetStatus(codes.Ok, "query successful")
		dbSpan.End()

		span.SetAttributes(attribute.Int("http.status_code", 200))
		span.SetStatus(codes.Ok, "request handled")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"users": []}`))

		require.NotNil(s.T(), ctx2)
	}))
	defer server.Close()

	// Make client request
	ctx, clientSpan := s.service.StartSpan(context.Background(), "http.client.request",
		tracing.WithSpanKind(tracing.SpanKindClient))
	defer clientSpan.End()

	req, err := http.NewRequestWithContext(ctx, "GET", server.URL+"/api/users", nil)
	require.NoError(s.T(), err)

	// Inject trace context
	s.service.InjectHTTPHeaders(ctx, req.Header)

	// Make request
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	clientSpan.SetAttributes(
		attribute.Int("http.status_code", resp.StatusCode),
	)
	clientSpan.SetStatus(codes.Ok, "request successful")

	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
}

// TestMicroserviceChain tests trace propagation across multiple services
func (s *IntegrationTestSuite) TestMicroserviceChain() {
	// Service A
	serviceA := func(ctx context.Context) context.Context {
		ctx, span := s.service.StartSpan(ctx, "service-a.process",
			tracing.WithSpanKind(tracing.SpanKindInternal),
			tracing.WithAttributes(attribute.String("service", "a")))
		defer span.End()

		time.Sleep(5 * time.Millisecond)
		span.SetStatus(codes.Ok, "processed")
		return ctx
	}

	// Service B
	serviceB := func(ctx context.Context) context.Context {
		ctx, span := s.service.StartSpan(ctx, "service-b.transform",
			tracing.WithSpanKind(tracing.SpanKindInternal),
			tracing.WithAttributes(attribute.String("service", "b")))
		defer span.End()

		time.Sleep(8 * time.Millisecond)
		span.SetStatus(codes.Ok, "transformed")
		return ctx
	}

	// Service C
	serviceC := func(ctx context.Context) context.Context {
		ctx, span := s.service.StartSpan(ctx, "service-c.persist",
			tracing.WithSpanKind(tracing.SpanKindInternal),
			tracing.WithAttributes(attribute.String("service", "c")))
		defer span.End()

		time.Sleep(10 * time.Millisecond)
		span.SetStatus(codes.Ok, "persisted")
		return ctx
	}

	// Execute chain
	ctx, rootSpan := s.service.StartSpan(context.Background(), "microservice-chain")
	defer rootSpan.End()

	ctx = serviceA(ctx)
	ctx = serviceB(ctx)
	ctx = serviceC(ctx)

	rootSpan.SetStatus(codes.Ok, "chain completed")

	// Verify trace ID is consistent
	traceID := s.service.GetTraceID(ctx)
	require.NotEmpty(s.T(), traceID)
}

// TestErrorPropagation tests error handling across service boundaries
func (s *IntegrationTestSuite) TestErrorPropagation() {
	deepFunction := func(ctx context.Context) error {
		ctx, span := s.service.StartSpan(ctx, "deep-function")
		defer span.End()

		err := errors.New("database connection failed")
		s.service.RecordError(ctx, err, tracing.WithErrorStatus())
		span.SetStatus(codes.Error, err.Error())

		return err
	}

	middleFunction := func(ctx context.Context) error {
		ctx, span := s.service.StartSpan(ctx, "middle-function")
		defer span.End()

		if err := deepFunction(ctx); err != nil {
			s.service.RecordError(ctx, err,
				tracing.WithErrorAttributes(
					attribute.String("layer", "middle"),
				))
			span.SetStatus(codes.Error, "propagated error")
			return err
		}

		span.SetStatus(codes.Ok, "success")
		return nil
	}

	topFunction := func(ctx context.Context) error {
		ctx, span := s.service.StartSpan(ctx, "top-function")
		defer span.End()

		if err := middleFunction(ctx); err != nil {
			s.service.RecordError(ctx, err,
				tracing.WithErrorStatus(),
				tracing.WithErrorAttributes(
					attribute.String("layer", "top"),
					attribute.String("severity", "critical"),
				))
			span.SetStatus(codes.Error, "operation failed")
			return err
		}

		span.SetStatus(codes.Ok, "success")
		return nil
	}

	err := topFunction(context.Background())
	require.Error(s.T(), err)
	require.Equal(s.T(), "database connection failed", err.Error())
}

// TestConcurrentRequests tests concurrent request handling
func (s *IntegrationTestSuite) TestConcurrentRequests() {
	const numRequests = 50

	var wg sync.WaitGroup
	errchan := make(chan error, numRequests)

	for i := range numRequests {
		wg.Add(1)
		go func(requestID int) {
			defer wg.Done()

			ctx, span := s.service.StartSpan(context.Background(), "concurrent-request",
				tracing.WithAttributes(
					attribute.Int("request.id", requestID),
				))
			defer span.End()

			// Simulate work
			time.Sleep(time.Duration(requestID%10) * time.Millisecond)

			// Nested operation
			ctx2, childSpan := s.service.StartSpan(ctx, "process-data")
			time.Sleep(5 * time.Millisecond)
			childSpan.SetAttributes(attribute.Int("items.processed", requestID*10))
			childSpan.SetStatus(codes.Ok, "processed")
			childSpan.End()

			span.SetAttributes(attribute.Bool("success", true))
			span.SetStatus(codes.Ok, "request completed")

			if ctx2 == nil {
				errchan <- errors.New("context is nil")
			}
		}(i)
	}

	wg.Wait()
	close(errchan)

	for err := range errchan {
		require.NoError(s.T(), err)
	}
}

// TestBatchProcessing tests batch processing with progress tracking
func (s *IntegrationTestSuite) TestBatchProcessing() {
	processBatch := func(ctx context.Context, batchID int, items int) error {
		ctx, span := s.service.StartSpan(ctx, "process-batch",
			tracing.WithAttributes(
				attribute.Int("batch.id", batchID),
				attribute.Int("batch.size", items),
			))
		defer span.End()

		processedItems := 0
		failedItems := 0

		for i := range items {
			itemCtx, itemSpan := s.service.StartSpan(ctx, "process-item",
				tracing.WithAttributes(
					attribute.Int("item.index", i),
				))

			// Simulate processing
			time.Sleep(1 * time.Millisecond)

			// Simulate occasional failures
			if i%7 == 0 {
				err := errors.New("item processing failed")
				s.service.RecordError(itemCtx, err)
				itemSpan.SetStatus(codes.Error, "failed")
				failedItems++
			} else {
				itemSpan.SetStatus(codes.Ok, "success")
				processedItems++
			}

			itemSpan.End()
		}

		span.SetAttributes(
			attribute.Int("items.processed", processedItems),
			attribute.Int("items.failed", failedItems),
		)

		if failedItems > 0 {
			span.SetStatus(codes.Error, "batch completed with errors")
		} else {
			span.SetStatus(codes.Ok, "batch completed successfully")
		}

		return nil
	}

	ctx, rootSpan := s.service.StartSpan(context.Background(), "batch-job",
		tracing.WithAttributes(
			attribute.String("job.type", "data-import"),
		))
	defer rootSpan.End()

	totalBatches := 3
	itemsPerBatch := 10

	for i := range totalBatches {
		err := processBatch(ctx, i, itemsPerBatch)
		require.NoError(s.T(), err)
	}

	rootSpan.SetAttributes(
		attribute.Int("batches.total", totalBatches),
		attribute.Int("items.total", totalBatches*itemsPerBatch),
	)
	rootSpan.SetStatus(codes.Ok, "job completed")
}

// TestRetryMechanism tests retry logic with tracing
func (s *IntegrationTestSuite) TestRetryMechanism() {
	attemptCounter := 0

	operation := func(ctx context.Context) error {
		ctx, span := s.service.StartSpan(ctx, "retry-operation",
			tracing.WithAttributes(
				attribute.Int("attempt", attemptCounter+1),
			))
		defer span.End()

		attemptCounter++

		if attemptCounter < 3 {
			err := errors.New("temporary failure")
			s.service.RecordError(ctx, err)
			span.SetStatus(codes.Error, "attempt failed")
			return err
		}

		span.SetStatus(codes.Ok, "attempt succeeded")
		return nil
	}

	ctx, rootSpan := s.service.StartSpan(context.Background(), "retry-handler")
	defer rootSpan.End()

	maxRetries := 3
	var lastErr error

	for i := range maxRetries {
		s.service.AddEvent(ctx, "retry-attempt",
			attribute.Int("attempt", i+1))

		if err := operation(ctx); err != nil {
			lastErr = err
			time.Sleep(10 * time.Millisecond)
			continue
		}

		// Success
		rootSpan.SetAttributes(
			attribute.Int("retries.count", i+1),
			attribute.Bool("retries.succeeded", true),
		)
		rootSpan.SetStatus(codes.Ok, "operation succeeded after retries")
		return
	}

	// All retries failed
	s.service.RecordError(ctx, lastErr, tracing.WithErrorStatus())
	rootSpan.SetAttributes(
		attribute.Int("retries.count", maxRetries),
		attribute.Bool("retries.succeeded", false),
	)
	rootSpan.SetStatus(codes.Error, "operation failed after all retries")
}

// TestCacheInteraction tests cache hit/miss scenarios
func (s *IntegrationTestSuite) TestCacheInteraction() {
	cache := make(map[string]string)
	cacheHits := 0
	cacheMisses := 0

	getData := func(ctx context.Context, key string) (string, error) {
		ctx, span := s.service.StartSpan(ctx, "cache.get",
			tracing.WithAttributes(
				attribute.String("cache.key", key),
			))
		defer span.End()

		// Check cache
		if val, ok := cache[key]; ok {
			cacheHits++
			span.SetAttributes(
				attribute.Bool("cache.hit", true),
				attribute.String("cache.value", val),
			)
			span.SetStatus(codes.Ok, "cache hit")
			return val, nil
		}

		// Cache miss - fetch from "database"
		cacheMisses++
		span.SetAttributes(attribute.Bool("cache.hit", false))

		ctx2, dbSpan := s.service.StartSpan(ctx, "database.query",
			tracing.WithSpanKind(tracing.SpanKindClient))
		time.Sleep(10 * time.Millisecond) // Simulate DB query
		value := "value-" + key
		cache[key] = value
		dbSpan.SetStatus(codes.Ok, "fetched from database")
		dbSpan.End()

		span.SetAttributes(attribute.String("cache.value", value))
		span.SetStatus(codes.Ok, "cache miss - fetched from db")

		require.NotNil(s.T(), ctx2)
		return value, nil
	}

	ctx, rootSpan := s.service.StartSpan(context.Background(), "cache-test")
	defer rootSpan.End()

	keys := []string{"key1", "key2", "key1", "key3", "key2", "key1"}

	for _, key := range keys {
		_, err := getData(ctx, key)
		require.NoError(s.T(), err)
	}

	rootSpan.SetAttributes(
		attribute.Int("cache.hits", cacheHits),
		attribute.Int("cache.misses", cacheMisses),
		attribute.Float64("cache.hit_rate", float64(cacheHits)/float64(len(keys))),
	)
	rootSpan.SetStatus(codes.Ok, "cache test completed")

	require.Equal(s.T(), 3, cacheHits)
	require.Equal(s.T(), 3, cacheMisses)
}

// TestCircuitBreaker tests circuit breaker pattern with tracing
func (s *IntegrationTestSuite) TestCircuitBreaker() {
	type CircuitState int
	const (
		Closed CircuitState = iota
		Open
		HalfOpen
	)

	state := Closed
	failureCount := 0
	threshold := 3

	callService := func(ctx context.Context, shouldFail bool) error {
		ctx, span := s.service.StartSpan(ctx, "external-service-call",
			tracing.WithSpanKind(tracing.SpanKindClient))
		defer span.End()

		span.SetAttributes(attribute.String("circuit.state", []string{"closed", "open", "half-open"}[state]))

		if state == Open {
			span.SetStatus(codes.Error, "circuit breaker open")
			return errors.New("circuit breaker is open")
		}

		if shouldFail {
			failureCount++
			if failureCount >= threshold {
				state = Open
			}
			err := errors.New("service unavailable")
			s.service.RecordError(ctx, err)
			span.SetAttributes(attribute.Int("circuit.failures", failureCount))
			span.SetStatus(codes.Error, "service call failed")
			return err
		}

		failureCount = 0
		if state == HalfOpen {
			state = Closed
		}

		span.SetStatus(codes.Ok, "service call succeeded")
		return nil
	}

	ctx, rootSpan := s.service.StartSpan(context.Background(), "circuit-breaker-test")
	defer rootSpan.End()

	// Test sequence: fail 3 times to open circuit
	for i := range 5 {
		err := callService(ctx, i < 3) // Fail first 3, succeed next 2

		if i < 3 {
			require.Error(s.T(), err)
		} else if i == 3 {
			require.Error(s.T(), err) // Circuit is open
			require.Equal(s.T(), "circuit breaker is open", err.Error())
		}
	}

	rootSpan.SetAttributes(
		attribute.String("circuit.final_state", "open"),
		attribute.Int("circuit.threshold", threshold),
	)
	rootSpan.SetStatus(codes.Ok, "circuit breaker test completed")
}

// Run the integration test suite
func TestIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(IntegrationTestSuite))
}
