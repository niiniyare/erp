package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/platform/config"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// CacheTestSuite is the test suite for cache operations
type CacheTestSuite struct {
	suite.Suite
	miniRedis *miniredis.Miniredis
	cache     Service
	ctx       context.Context
	tenantID  uuid.UUID
}

// SetupSuite runs once before all tests
func (s *CacheTestSuite) SetupSuite() {
	var err error
	s.miniRedis, err = miniredis.Run()
	require.NoError(s.T(), err)
}

// TearDownSuite runs once after all tests
func (s *CacheTestSuite) TearDownSuite() {
	if s.miniRedis != nil {
		s.miniRedis.Close()
	}
}

// SetupTest runs before each test
func (s *CacheTestSuite) SetupTest() {
	s.miniRedis.FlushAll()

	cfg := &RedisConfig{
		RedisConfig: &config.RedisConfig{
			Host:     s.miniRedis.Host(),
			Port:     s.miniRedis.Server().Addr().Port,
			Password: "",
			DB:       0,
		},
		PoolSize:                   10,
		MinIdleConns:               3,
		MaxConnAge:                 30 * time.Minute,
		PoolTimeout:                4 * time.Second,
		IdleTimeout:                5 * time.Minute,
		EnableCompression:          true,
		CompressionLevel:           6,
		KeyPrefix:                  "test",
		MaxRetries:                 3,
		RetryDelay:                 100 * time.Millisecond,
		BatchDeleteSize:            1000,
		EnableCircuitBreaker:       false, // Disable for tests
		RequireTenantContext:       true,
		AllowGlobalOperations:      false,
		EnableMemoryCache:          true,
		MemoryCacheMaxSize:         100,
		MemoryCacheDefaultTTL:      5 * time.Minute,
		MemoryCacheCleanupInterval: 1 * time.Minute,
		EnableTracing:              false,
		EnableMetrics:              false,
		EnableLogging:              false,
	}

	var err error
	s.cache, err = NewRedisClient(cfg)
	require.NoError(s.T(), err)

	s.tenantID = uuid.New()
	s.ctx = WithTenantID(context.Background(), s.tenantID)
}

// TearDownTest runs after each test
func (s *CacheTestSuite) TearDownTest() {
	if s.cache != nil {
		s.cache.Close()
	}
}

// TestGetSet tests basic Get and Set operations
func (s *CacheTestSuite) TestGetSet() {
	tests := []struct {
		name        string
		key         string
		value       any
		ttl         time.Duration
		expectError bool
		errorType   error
	}{
		{
			name:        "set and get string",
			key:         "test:string",
			value:       "hello world",
			ttl:         time.Hour,
			expectError: false,
		},
		{
			name:        "set and get int",
			key:         "test:int",
			value:       12345,
			ttl:         time.Hour,
			expectError: false,
		},
		{
			name: "set and get struct",
			key:  "test:struct",
			value: struct {
				Name  string
				Email string
				Age   int
			}{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   30,
			},
			ttl:         time.Hour,
			expectError: false,
		},
		{
			name: "set and get map",
			key:  "test:map",
			value: map[string]any{
				"name":  "Jane Doe",
				"email": "jane@example.com",
				"age":   25,
			},
			ttl:         time.Hour,
			expectError: false,
		},
		{
			name: "set and get slice",
			key:  "test:slice",
			value: []string{
				"item1",
				"item2",
				"item3",
			},
			ttl:         time.Hour,
			expectError: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Set value
			err := s.cache.Set(s.ctx, tt.key, tt.value, tt.ttl)
			if tt.expectError {
				s.Error(err)
				if tt.errorType != nil {
					s.True(errors.Is(err, tt.errorType))
				}
				return
			}
			s.NoError(err)

			// Get value
			var result any
			err = s.cache.Get(s.ctx, tt.key, &result)
			s.NoError(err)
			s.NotNil(result)
		})
	}
}

// TestCacheMiss tests cache miss scenarios
func (s *CacheTestSuite) TestCacheMiss() {
	var result string
	err := s.cache.Get(s.ctx, "nonexistent:key", &result)
	s.Error(err)
	s.True(errors.Is(err, ErrCacheMiss))
}

// TestDelete tests delete operations
func (s *CacheTestSuite) TestDelete() {
	tests := []struct {
		name        string
		setupFunc   func() string
		expectError bool
	}{
		{
			name: "delete existing key",
			setupFunc: func() string {
				key := "test:delete:existing"
				s.cache.Set(s.ctx, key, "value", time.Hour)
				return key
			},
			expectError: false,
		},
		{
			name: "delete non-existent key",
			setupFunc: func() string {
				return "test:delete:nonexistent"
			},
			expectError: false, // Delete non-existent key shouldn't error
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			key := tt.setupFunc()
			err := s.cache.Delete(s.ctx, key)
			if tt.expectError {
				s.Error(err)
			} else {
				s.NoError(err)
			}

			// Verify deletion
			var result string
			err = s.cache.Get(s.ctx, key, &result)
			s.Error(err)
			s.True(errors.Is(err, ErrCacheMiss))
		})
	}
}

// TestTenantIsolation tests that tenants are properly isolated
func (s *CacheTestSuite) TestTenantIsolation() {
	tenant1 := uuid.New()
	tenant2 := uuid.New()

	ctx1 := WithTenantID(context.Background(), tenant1)
	ctx2 := WithTenantID(context.Background(), tenant2)

	key := "shared:key"
	value1 := "tenant1-value"
	value2 := "tenant2-value"

	// Set value for tenant1
	err := s.cache.Set(ctx1, key, value1, time.Hour)
	s.NoError(err)

	// Set value for tenant2
	err = s.cache.Set(ctx2, key, value2, time.Hour)
	s.NoError(err)

	// Get value for tenant1
	var result1 string
	err = s.cache.Get(ctx1, key, &result1)
	s.NoError(err)
	s.Equal(value1, result1)

	// Get value for tenant2
	var result2 string
	err = s.cache.Get(ctx2, key, &result2)
	s.NoError(err)
	s.Equal(value2, result2)

	// Verify isolation
	s.NotEqual(result1, result2)
}

// TestTenantContextMethods tests different tenant context methods
func (s *CacheTestSuite) TestTenantContextMethods() {
	tests := []struct {
		name      string
		ctxFunc   func(context.Context) context.Context
		shouldErr bool
	}{
		{
			name: "with tenant ID",
			ctxFunc: func(ctx context.Context) context.Context {
				return WithTenantID(ctx, uuid.New())
			},
			shouldErr: false,
		},
		{
			name: "with tenant slug",
			ctxFunc: func(ctx context.Context) context.Context {
				return WithTenantSlug(ctx, "test-tenant")
			},
			shouldErr: false,
		},
		{
			name: "with tenant subdomain",
			ctxFunc: func(ctx context.Context) context.Context {
				return WithTenantSubdomain(ctx, "test")
			},
			shouldErr: false,
		},
		{
			name: "without tenant context",
			ctxFunc: func(ctx context.Context) context.Context {
				return context.Background()
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := tt.ctxFunc(context.Background())
			err := s.cache.Set(ctx, "test:key", "value", time.Hour)
			if tt.shouldErr {
				s.Error(err)
				s.True(errors.Is(err, ErrNoTenantContext))
			} else {
				s.NoError(err)
			}
		})
	}
}

// TestNamespace tests namespace functionality
func (s *CacheTestSuite) TestNamespace() {
	baseCtx := WithTenantID(context.Background(), s.tenantID)

	ctx1 := WithNamespace(baseCtx, "users")
	ctx2 := WithNamespace(baseCtx, "sessions")

	key := "123"
	value1 := "user-data"
	value2 := "session-data"

	// Set in different namespaces
	s.cache.Set(ctx1, key, value1, time.Hour)
	s.cache.Set(ctx2, key, value2, time.Hour)

	// Get from namespace 1
	var result1 string
	err := s.cache.Get(ctx1, key, &result1)
	s.NoError(err)
	s.Equal(value1, result1)

	// Get from namespace 2
	var result2 string
	err = s.cache.Get(ctx2, key, &result2)
	s.NoError(err)
	s.Equal(value2, result2)

	// Verify namespace isolation
	s.NotEqual(result1, result2)
}

// TestMGet tests bulk get operations
func (s *CacheTestSuite) TestMGet() {
	// Setup test data
	testData := map[string]string{
		"user:1": "Alice",
		"user:2": "Bob",
		"user:3": "Charlie",
	}

	for key, value := range testData {
		err := s.cache.Set(s.ctx, key, value, time.Hour)
		s.NoError(err)
	}

	tests := []struct {
		name          string
		keys          []string
		expectResults int
		expectMisses  int
	}{
		{
			name:          "get all existing keys",
			keys:          []string{"user:1", "user:2", "user:3"},
			expectResults: 3,
			expectMisses:  0,
		},
		{
			name:          "get mix of existing and non-existing keys",
			keys:          []string{"user:1", "user:999", "user:3"},
			expectResults: 3,
			expectMisses:  1,
		},
		{
			name:          "get non-existing keys",
			keys:          []string{"user:999", "user:888"},
			expectResults: 2,
			expectMisses:  2,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			results, err := s.cache.MGet(s.ctx, tt.keys)
			s.NoError(err)
			s.Len(results, tt.expectResults)

			misses := 0
			for _, result := range results {
				if errors.Is(result.Err, ErrCacheMiss) {
					misses++
				}
			}
			s.Equal(tt.expectMisses, misses)
		})
	}
}

// TestMSet tests bulk set operations
func (s *CacheTestSuite) TestMSet() {
	pairs := map[string]any{
		"user:1": "Alice",
		"user:2": "Bob",
		"user:3": "Charlie",
	}

	err := s.cache.MSet(s.ctx, pairs, time.Hour)
	s.NoError(err)

	// Verify all values were set
	for key, expectedValue := range pairs {
		var result string
		err := s.cache.Get(s.ctx, key, &result)
		s.NoError(err)
		s.Equal(expectedValue, result)
	}
}

// TestMDelete tests bulk delete operations
func (s *CacheTestSuite) TestMDelete() {
	// Setup test data
	keys := []string{"user:1", "user:2", "user:3"}
	for _, key := range keys {
		err := s.cache.Set(s.ctx, key, "value", time.Hour)
		s.NoError(err)
	}

	// Delete all keys
	err := s.cache.MDelete(s.ctx, keys)
	s.NoError(err)

	// Verify all keys were deleted
	for _, key := range keys {
		var result string
		err := s.cache.Get(s.ctx, key, &result)
		s.Error(err)
		s.True(errors.Is(err, ErrCacheMiss))
	}
}

// TestExists tests key existence check
func (s *CacheTestSuite) TestExists() {
	tests := []struct {
		name      string
		setupFunc func() string
		expect    bool
	}{
		{
			name: "existing key",
			setupFunc: func() string {
				key := "test:exists:yes"
				s.cache.Set(s.ctx, key, "value", time.Hour)
				return key
			},
			expect: true,
		},
		{
			name: "non-existing key",
			setupFunc: func() string {
				return "test:exists:no"
			},
			expect: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			key := tt.setupFunc()
			exists, err := s.cache.Exists(s.ctx, key)
			s.NoError(err)
			s.Equal(tt.expect, exists)
		})
	}
}

// TestTTL tests TTL operations
func (s *CacheTestSuite) TestTTL() {
	tests := []struct {
		name      string
		setupFunc func() string
		checkFunc func(time.Duration)
	}{
		{
			name: "key with TTL",
			setupFunc: func() string {
				key := "test:ttl:with"
				s.cache.Set(s.ctx, key, "value", time.Hour)
				return key
			},
			checkFunc: func(ttl time.Duration) {
				s.True(ttl > 0)
				s.True(ttl <= time.Hour)
			},
		},
		{
			name: "non-existing key",
			setupFunc: func() string {
				return "test:ttl:nonexistent"
			},
			checkFunc: func(ttl time.Duration) {
				s.Equal(time.Duration(-2), ttl) // Redis returns -2 for non-existent keys
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			key := tt.setupFunc()
			ttl, err := s.cache.TTL(s.ctx, key)
			s.NoError(err)
			tt.checkFunc(ttl)
		})
	}
}

// TestExpire tests setting expiration on existing keys
func (s *CacheTestSuite) TestExpire() {
	key := "test:expire"
	value := "test-value"

	// Set key without expiration
	err := s.cache.Set(s.ctx, key, value, 0)
	s.NoError(err)

	// Set expiration
	err = s.cache.Expire(s.ctx, key, time.Hour)
	s.NoError(err)

	// Verify expiration was set
	ttl, err := s.cache.TTL(s.ctx, key)
	s.NoError(err)
	s.True(ttl > 0)
	s.True(ttl <= time.Hour)
}

// TestFlush tests flushing tenant cache
func (s *CacheTestSuite) TestFlush() {
	// Set multiple keys for current tenant
	keys := []string{"user:1", "user:2", "session:1"}
	for _, key := range keys {
		err := s.cache.Set(s.ctx, key, "value", time.Hour)
		s.NoError(err)
	}

	// Set keys for different tenant
	otherTenantCtx := WithTenantID(context.Background(), uuid.New())
	err := s.cache.Set(otherTenantCtx, "user:1", "other-value", time.Hour)
	s.NoError(err)

	// Flush current tenant's cache
	err = s.cache.Flush(s.ctx)
	s.NoError(err)

	// Verify current tenant's keys are deleted
	for _, key := range keys {
		var result string
		err := s.cache.Get(s.ctx, key, &result)
		s.Error(err)
		s.True(errors.Is(err, ErrCacheMiss))
	}

	// Verify other tenant's keys still exist
	var otherResult string
	err = s.cache.Get(otherTenantCtx, "user:1", &otherResult)
	s.NoError(err)
	s.Equal("other-value", otherResult)
}

// TestCompression tests data compression
func (s *CacheTestSuite) TestCompression() {
	// Create large data that should trigger compression
	largeData := make([]byte, 2048) // 2KB, exceeds 1KB threshold
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	tests := []struct {
		name           string
		key            string
		data           any
		shouldCompress bool
	}{
		{
			name:           "large data triggers compression",
			key:            "test:large",
			data:           largeData,
			shouldCompress: true,
		},
		{
			name:           "small data no compression",
			key:            "test:small",
			data:           "small",
			shouldCompress: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := s.cache.Set(s.ctx, tt.key, tt.data, time.Hour)
			s.NoError(err)

			var result any
			err = s.cache.Get(s.ctx, tt.key, &result)
			s.NoError(err)
		})
	}
}

// TestMemoryCache tests in-memory cache operations
func (s *CacheTestSuite) TestMemoryCache() {
	tests := []struct {
		name  string
		key   string
		value interface{}
		ttl   time.Duration
	}{
		{
			name:  "set and get from memory",
			key:   "mem:test1",
			value: "memory-value",
			ttl:   time.Hour,
		},
		{
			name: "set and get struct from memory",
			key:  "mem:test2",
			value: struct {
				Name string
				Age  int
			}{
				Name: "John",
				Age:  30,
			},
			ttl: time.Hour,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Set in memory cache
			err := s.cache.SetMemory(s.ctx, tt.key, tt.value, tt.ttl)
			s.NoError(err)

			// Get from memory cache
			var result any
			err = s.cache.GetMemory(s.ctx, tt.key, &result)
			s.NoError(err)
			s.NotNil(result)
		})
	}
}

// TestMemoryCacheExpiration tests memory cache expiration
func (s *CacheTestSuite) TestMemoryCacheExpiration() {
	key := "mem:expire"
	value := "test"

	// Set with short TTL
	err := s.cache.SetMemory(s.ctx, key, value, 100*time.Millisecond)
	s.NoError(err)

	// Should exist immediately
	var result string
	err = s.cache.GetMemory(s.ctx, key, &result)
	s.NoError(err)

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Should be expired
	err = s.cache.GetMemory(s.ctx, key, &result)
	s.Error(err)
	s.True(errors.Is(err, ErrCacheMiss))
}

// TestGlobalMemoryCache tests global memory cache (not tenant-aware)
func (s *CacheTestSuite) TestGlobalMemoryCache() {
	key := "formula:sum"
	value := "A1+B1"

	// Set in global cache
	err := s.cache.SetGlobalMemory(key, value, time.Hour)
	s.NoError(err)

	// Get from global cache
	var result string
	err = s.cache.GetGlobalMemory(key, &result)
	s.NoError(err)
	s.Equal(value, result)

	// Should be accessible regardless of tenant
	anotherTenantCtx := WithTenantID(context.Background(), uuid.New())
	_ = anotherTenantCtx // Global cache ignores tenant context

	err = s.cache.GetGlobalMemory(key, &result)
	s.NoError(err)
	s.Equal(value, result)
}

// TestStats tests cache statistics
func (s *CacheTestSuite) TestStats() {
	// Reset stats
	s.cache.Reset()

	// Perform operations
	s.cache.Set(s.ctx, "test:1", "value", time.Hour)
	s.cache.Get(s.ctx, "test:1", new(string))      // Hit
	s.cache.Get(s.ctx, "nonexistent", new(string)) // Miss

	stats := s.cache.Stats()
	s.Equal(int64(1), stats.Hits)
	s.Equal(int64(1), stats.Misses)
	s.Equal(int64(1), stats.Sets)
	s.True(stats.HitRatio >= 0 && stats.HitRatio <= 1)
}

// TestPing tests connection health check
func (s *CacheTestSuite) TestPing() {
	err := s.cache.Ping(s.ctx)
	s.NoError(err)
}

// TestConfigValidation tests configuration validation
func (s *CacheTestSuite) TestConfigValidation() {
	tests := []struct {
		name      string
		config    *RedisConfig
		expectErr bool
	}{
		{
			name: "valid config",
			config: &RedisConfig{
				RedisConfig: &config.RedisConfig{
					Host: "localhost",
					Port: 6379,
				},
				PoolSize:           10,
				CompressionLevel:   6,
				KeyPrefix:          "test",
				BatchDeleteSize:    1000,
				MemoryCacheMaxSize: 100,
			},
			expectErr: false,
		},
		{
			name: "nil base config",
			config: &RedisConfig{
				RedisConfig: nil,
			},
			expectErr: true,
		},
		{
			name: "invalid pool size",
			config: &RedisConfig{
				RedisConfig: &config.RedisConfig{
					Host: "localhost",
					Port: 6379,
				},
				PoolSize: 0,
			},
			expectErr: true,
		},
		{
			name: "invalid compression level",
			config: &RedisConfig{
				RedisConfig: &config.RedisConfig{
					Host: "localhost",
					Port: 6379,
				},
				PoolSize:         10,
				CompressionLevel: 15, // Invalid: > 9
			},
			expectErr: true,
		},
		{
			name: "empty key prefix",
			config: &RedisConfig{
				RedisConfig: &config.RedisConfig{
					Host: "localhost",
					Port: 6379,
				},
				PoolSize:  10,
				KeyPrefix: "",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := tt.config.Validate()
			if tt.expectErr {
				s.Error(err)
			} else {
				s.NoError(err)
			}
		})
	}
}

// TestErrorHandling tests various error scenarios
func (s *CacheTestSuite) TestErrorHandling() {
	tests := []struct {
		name      string
		operation func() error
		expectErr error
	}{
		{
			name: "nil value on set",
			operation: func() error {
				return s.cache.Set(s.ctx, "test", nil, time.Hour)
			},
			expectErr: ErrNilValue,
		},
		{
			name: "nil destination on get",
			operation: func() error {
				return s.cache.Get(s.ctx, "test", nil)
			},
			expectErr: ErrNilValue,
		},
		{
			name: "empty key",
			operation: func() error {
				return s.cache.Set(s.ctx, "", "value", time.Hour)
			},
			expectErr: nil, // Should return validation error
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := tt.operation()
			s.Error(err)
			if tt.expectErr != nil {
				s.True(errors.Is(err, tt.expectErr))
			}
		})
	}
}

// TestConcurrentAccess tests concurrent cache operations
func (s *CacheTestSuite) TestConcurrentAccess() {
	const goroutines = 10
	const operations = 100

	done := make(chan bool, goroutines)

	for i := range goroutines {
		go func(id int) {
			for j := range operations {
				key := "concurrent:test"
				value := id*operations + j

				// Set
				err := s.cache.Set(s.ctx, key, value, time.Hour)
				s.NoError(err)

				// Get
				var result int
				err = s.cache.Get(s.ctx, key, &result)
				if err != nil && !errors.Is(err, ErrCacheMiss) {
					s.NoError(err)
				}
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for range goroutines {
		<-done
	}
}

// TestKeyPatterns tests pattern-based operations
func (s *CacheTestSuite) TestKeyPatterns() {
	// Setup test data
	keys := []string{
		"user:1:profile",
		"user:1:settings",
		"user:2:profile",
		"session:abc",
	}

	for _, key := range keys {
		err := s.cache.Set(s.ctx, key, "value", time.Hour)
		s.NoError(err)
	}

	tests := []struct {
		name          string
		pattern       string
		expectMinKeys int
	}{
		{
			name:          "match user 1 keys",
			pattern:       "user:1:*",
			expectMinKeys: 2,
		},
		{
			name:          "match all user keys",
			pattern:       "user:*",
			expectMinKeys: 3,
		},
		{
			name:          "match session keys",
			pattern:       "session:*",
			expectMinKeys: 1,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			foundKeys, err := s.cache.Keys(s.ctx, tt.pattern)
			s.NoError(err)
			s.GreaterOrEqual(len(foundKeys), tt.expectMinKeys)
		})
	}
}

// TestDeletePattern tests pattern-based deletion
func (s *CacheTestSuite) TestDeletePattern() {
	// Setup test data
	userKeys := []string{"user:1", "user:2", "user:3"}
	sessionKeys := []string{"session:a", "session:b"}

	for _, key := range userKeys {
		s.cache.Set(s.ctx, key, "value", time.Hour)
	}
	for _, key := range sessionKeys {
		s.cache.Set(s.ctx, key, "value", time.Hour)
	}

	// Delete user keys
	err := s.cache.DeletePattern(s.ctx, "user:*")
	s.NoError(err)

	// Verify user keys are deleted
	for _, key := range userKeys {
		var result string
		err := s.cache.Get(s.ctx, key, &result)
		s.Error(err)
		s.True(errors.Is(err, ErrCacheMiss))
	}

	// Verify session keys still exist
	for _, key := range sessionKeys {
		var result string
		err := s.cache.Get(s.ctx, key, &result)
		s.NoError(err)
	}
}

// Run the test suite
func TestCacheTestSuite(t *testing.T) {
	suite.Run(t, new(CacheTestSuite))
}
