# Tracing Package Test Suite

Comprehensive test coverage for the distributed tracing package using `testify/suite` and `testify/require`.

## Test Structure

```
tracing/
├── tracing.go                    # Main implementation
├── tracing_test.go               # Unit tests with test suite
├── tracing_integration_test.go   # Integration tests
└── README_TESTS.md              # This file
```

## Running Tests

### Run All Tests

```bash
go test -v ./...
```

### Run Unit Tests Only

```bash
go test -v -run TestTracingServiceTestSuite
```

### Run Integration Tests Only

```bash
go test -v -run TestIntegrationTestSuite
```

### Run Specific Test

```bash
go test -v -run TestTracingServiceTestSuite/TestStartSpan_Basic
```

### Skip Integration Tests (Short Mode)

```bash
go test -short -v ./...
```

### Run with Coverage

```bash
go test -v -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Run with Race Detector

```bash
go test -race -v ./...
```

### Run Benchmarks

```bash
go test -bench=. -benchmem ./...
```

### Run Specific Benchmark

```bash
go test -bench=BenchmarkStartSpan -benchmem ./...
```

## Test Suites

### 1. TracingServiceTestSuite

**Location**: `tracing_test.go`

**Purpose**: Unit tests for core tracing functionality

**Coverage**:
- ✅ Service creation and configuration
- ✅ Span lifecycle (start, end, attributes)
- ✅ Error recording and handling
- ✅ Event tracking
- ✅ HTTP header propagation
- ✅ Context management
- ✅ Shutdown behavior
- ✅ Concurrent operations
- ✅ No-op service behavior
- ✅ Configuration validation

**Test Count**: 50+ test cases

**Example**:
```bash
go test -v -run TestTracingServiceTestSuite/TestStartSpan_WithOptions
```

### 2. IntegrationTestSuite

**Location**: `tracing_integration_test.go`

**Purpose**: Integration tests for real-world scenarios

**Coverage**:
- ✅ Full ABAC permission evaluation flow
- ✅ HTTP server/client communication
- ✅ Microservice trace propagation
- ✅ Error propagation across services
- ✅ Concurrent request handling
- ✅ Batch processing with progress
- ✅ Retry mechanisms
- ✅ Cache interaction patterns
- ✅ Circuit breaker patterns

**Test Count**: 10+ complex scenarios

**Example**:
```bash
go test -v -run TestIntegrationTestSuite/TestFullABACFlow
```

## Test Categories

### Configuration Tests
```bash
go test -v -run "TestNewService|TestDefaultConfig|TestConfigValidation"
```

### Span Tests
```bash
go test -v -run "TestStartSpan|TestSpanEnd|TestSetAttributes"
```

### Error Handling Tests
```bash
go test -v -run "TestRecordError|TestErrorPropagation"
```

### HTTP Propagation Tests
```bash
go test -v -run "TestHTTPHeaderPropagation|TestHTTPServerClientFlow"
```

### Concurrency Tests
```bash
go test -v -run "TestConcurrent"
```

### Shutdown Tests
```bash
go test -v -run "TestShutdown"
```

## Benchmark Results (Expected)

```
BenchmarkStartSpan-8                     500000    2500 ns/op     800 B/op    12 allocs/op
BenchmarkStartSpanWithAttributes-8       400000    3000 ns/op    1200 B/op    18 allocs/op
BenchmarkNestedSpans-8                   200000    8000 ns/op    2400 B/op    36 allocs/op
BenchmarkSetAttributes-8                2000000     800 ns/op     400 B/op     6 allocs/op
BenchmarkRecordError-8                  1000000    1500 ns/op     600 B/op     9 allocs/op
BenchmarkNoOpService-8                 10000000     150 ns/op       0 B/op     0 allocs/op
```

## Coverage Goals

- **Unit Tests**: > 90% coverage
- **Integration Tests**: Cover all critical paths
- **Edge Cases**: Nil checks, empty values, concurrent access
- **Error Scenarios**: All error paths tested

## Writing New Tests

### Adding a Unit Test

```go
func (s *TracingServiceTestSuite) TestMyNewFeature() {
	ctx, span := s.service.StartSpan(s.ctx, "test-operation")
	defer span.End()

	// Your test logic
	span.SetAttributes(attribute.String("key", "value"))

	require.True(s.T(), span.IsRecording())
}
```

### Adding an Integration Test

```go
func (s *IntegrationTestSuite) TestMyIntegration() {
	ctx, rootSpan := s.service.StartSpan(context.Background(), "integration-test")
	defer rootSpan.End()

	// Simulate real-world scenario
	// ...

	require.NotNil(s.T(), ctx)
}
```

### Adding a Benchmark

```go
func BenchmarkMyOperation(b *testing.B) {
	service, _ := tracing.NewService(tracing.DefaultConfig())
	defer service.Shutdown(context.Background())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, span := service.StartSpan(context.Background(), "operation")
		span.End()
	}
}
```

## Test Patterns

### Table-Driven Tests

```go
func (s *TracingServiceTestSuite) TestMultipleScenarios() {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"case1", "input1", "output1"},
		{"case2", "input2", "output2"},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Test logic
			require.Equal(s.T(), tc.expected, result)
		})
	}
}
```

### Error Testing

```go
func (s *TracingServiceTestSuite) TestErrorScenario() {
	ctx, span := s.service.StartSpan(s.ctx, "operation")
	defer span.End()

	err := errors.New("test error")
	s.service.RecordError(ctx, err, tracing.WithErrorStatus())

	require.Error(s.T(), err)
	require.Equal(s.T(), "test error", err.Error())
}
```

### Concurrent Testing

```go
func (s *TracingServiceTestSuite) TestConcurrency() {
	const numGoroutines = 100
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, span := s.service.StartSpan(s.ctx, "concurrent-op")
			defer span.End()
		}()
	}

	wg.Wait()
}
```

## Continuous Integration

### GitHub Actions Example

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

## Troubleshooting

### Tests Hanging

```bash
# Run with timeout
go test -timeout 30s -v ./...
```

### Memory Issues

```bash
# Run tests sequentially
go test -p 1 -v ./...
```

### Verbose Output

```bash
# Maximum verbosity
go test -v -trace trace.out ./...
go tool trace trace.out
```

## Test Maintenance

### Regular Tasks

1. **Run full test suite**: `make test`
2. **Check coverage**: `make coverage`
3. **Run benchmarks**: `make bench`
4. **Update mocks**: `go generate ./...`

### Before Committing

```bash
# Full validation
go test -v -race -cover ./...
go vet ./...
golangci-lint run
```

## Test Data

### Mock Service

```go
service := tracing.NewNoOpService()
// Use for testing without real tracing
```

### Test Configuration

```go
config := tracing.Config{
	ServiceName:        "test-service",
	Protocol:           tracing.ProtocolStdout,
	Enabled:            true,
	SamplingRate:       1.0,
	BatchTimeout:       1 * time.Second,
	MaxExportBatchSize: 100,
	MaxQueueSize:       1000,
}
```

## Performance Testing

### Load Test

```go
func TestLoadScenario(t *testing.T) {
	service, _ := tracing.NewService(config)
	defer service.Shutdown(context.Background())

	for i := 0; i < 10000; i++ {
		ctx, span := service.StartSpan(context.Background(), "load-test")
		span.End()
	}
}
```

### Stress Test

```bash
go test -run TestConcurrentOperations -count 100
```

## Test Reports

### Generate HTML Report

```bash
go test -v -json ./... | go-test-report
```

### Generate Coverage Badge

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep total | awk '{print $3}'
```

## Best Practices

1. ✅ Use `testify/suite` for test organization
2. ✅ Use `testify/require` for assertions (fails fast)
3. ✅ Clean up resources in `TearDownTest`
4. ✅ Use table-driven tests for multiple scenarios
5. ✅ Test both success and error paths
6. ✅ Use subtests for organization: `s.Run("subtest", func() {...})`
7. ✅ Test concurrent scenarios with race detector
8. ✅ Benchmark performance-critical paths
9. ✅ Mock external dependencies
10. ✅ Keep tests independent and idempotent

## Quick Reference

| Command | Purpose |
|---------|---------|
| `go test -v ./...` | Run all tests with verbose output |
| `go test -short ./...` | Skip integration tests |
| `go test -race ./...` | Run with race detector |
| `go test -cover ./...` | Run with coverage |
| `go test -bench=. ./...` | Run benchmarks |
| `go test -run TestName` | Run specific test |
| `go test -timeout 30s` | Set test timeout |
| `go test -count=10` | Run tests 10 times |

## Support

For questions or issues:
1. Check test output for detailed error messages
2. Run with `-v` flag for verbose output
3. Use `-race` flag to detect race conditions
4. Check coverage reports for untested code paths
