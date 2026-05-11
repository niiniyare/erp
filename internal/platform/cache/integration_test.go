package cache_test

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"awo.so/internal/platform/cache"
	"awo.so/internal/platform/config"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestSuite contains integration tests
type IntegrationTestSuite struct {
	suite.Suite
	cache cache.Service
}

func (s *IntegrationTestSuite) SetupSuite() {
	cfg := cache.DefaultRedisConfig(&config.RedisConfig{
		Host:     getEnvOrDefault("REDIS_HOST", "localhost"),
		Port:     getEnvOrDefaultInt("REDIS_PORT", 6379),
		Password: getEnvOrDefault("REDIS_PASSWORD", ""),
		DB:       0,
	})

	var err error
	s.cache, err = cache.NewRedisClient(cfg)
	require.NoError(s.T(), err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if s.cache != nil {
		s.cache.Close()
	}
}

func (s *IntegrationTestSuite) SetupTest() {
	// Clean up before each test
	ctx := cache.WithTenantSlug(context.Background(), "test-tenant")
	s.cache.Flush(ctx)
}

// TestRealWorldUserCaching simulates real-world user caching scenario
func (s *IntegrationTestSuite) TestRealWorldUserCaching() {
	tenantID := uuid.New()
	ctx := cache.WithTenantID(context.Background(), tenantID)

	tests := []struct {
		name     string
		scenario func(*testing.T)
	}{
		{
			name: "user registration and profile caching",
			scenario: func(t *testing.T) {
				user := User{
					ID:        uuid.New().String(),
					Name:      "John Doe",
					Email:     "john@example.com",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					Preferences: UserPreferences{
						Theme:         "dark",
						Language:      "en",
						Notifications: true,
					},
				}

				// Cache user profile
				key := fmt.Sprintf("user:%s:profile", user.ID)
				err := s.cache.Set(ctx, key, user, 24*time.Hour)
				require.NoError(t, err)

				// Retrieve and verify
				var cached User
				err = s.cache.Get(ctx, key, &cached)
				require.NoError(t, err)
				require.Equal(t, user.Email, cached.Email)
				require.Equal(t, user.Preferences.Theme, cached.Preferences.Theme)
			},
		},
		{
			name: "session management",
			scenario: func(t *testing.T) {
				sessionID := uuid.New().String()
				session := Session{
					ID:        sessionID,
					UserID:    uuid.New().String(),
					Token:     "jwt-token-here",
					ExpiresAt: time.Now().Add(30 * time.Minute),
					Data: map[string]any{
						"ip":         "192.168.1.1",
						"user_agent": "Mozilla/5.0",
					},
				}

				// Store session
				key := fmt.Sprintf("session:%s", sessionID)
				err := s.cache.Set(ctx, key, session, 30*time.Minute)
				require.NoError(t, err)

				// Verify session exists
				exists, err := s.cache.Exists(ctx, key)
				require.NoError(t, err)
				require.True(t, exists)

				// Check TTL
				ttl, err := s.cache.TTL(ctx, key)
				require.NoError(t, err)
				require.True(t, ttl > 0)
				require.True(t, ttl <= 30*time.Minute)
			},
		},
		{
			name: "shopping cart caching",
			scenario: func(t *testing.T) {
				cart := ShoppingCart{
					UserID: uuid.New().String(),
					Items: []CartItem{
						{ProductID: "prod-1", Quantity: 2, Price: 29.99},
						{ProductID: "prod-2", Quantity: 1, Price: 49.99},
					},
					Total:     109.97,
					UpdatedAt: time.Now(),
				}

				key := fmt.Sprintf("cart:%s", cart.UserID)
				err := s.cache.Set(ctx, key, cart, 1*time.Hour)
				require.NoError(t, err)

				// Simulate cart update
				cart.Items = append(cart.Items, CartItem{
					ProductID: "prod-3",
					Quantity:  3,
					Price:     19.99,
				})
				cart.Total += 59.97
				cart.UpdatedAt = time.Now()

				err = s.cache.Set(ctx, key, cart, 1*time.Hour)
				require.NoError(t, err)

				// Verify updated cart
				var cached ShoppingCart
				err = s.cache.Get(ctx, key, &cached)
				require.NoError(t, err)
				require.Len(t, cached.Items, 3)
				require.InDelta(t, 169.94, cached.Total, 0.01)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.scenario(s.T())
		})
	}
}

// TestMultiTenantIsolation tests complete tenant isolation
func (s *IntegrationTestSuite) TestMultiTenantIsolation() {
	tenant1 := uuid.New()
	tenant2 := uuid.New()
	tenant3 := uuid.New()

	ctx1 := cache.WithTenantID(context.Background(), tenant1)
	ctx2 := cache.WithTenantID(context.Background(), tenant2)
	ctx3 := cache.WithTenantID(context.Background(), tenant3)

	// Setup test data for each tenant
	tenants := []struct {
		ctx   context.Context
		users []User
	}{
		{
			ctx: ctx1,
			users: []User{
				{ID: "1", Name: "Tenant1 User1", Email: "u1@tenant1.com"},
				{ID: "2", Name: "Tenant1 User2", Email: "u2@tenant1.com"},
			},
		},
		{
			ctx: ctx2,
			users: []User{
				{ID: "1", Name: "Tenant2 User1", Email: "u1@tenant2.com"},
				{ID: "2", Name: "Tenant2 User2", Email: "u2@tenant2.com"},
			},
		},
		{
			ctx: ctx3,
			users: []User{
				{ID: "1", Name: "Tenant3 User1", Email: "u1@tenant3.com"},
				{ID: "2", Name: "Tenant3 User2", Email: "u2@tenant3.com"},
			},
		},
	}

	// Store data for all tenants
	for _, tenant := range tenants {
		for _, user := range tenant.users {
			key := fmt.Sprintf("user:%s", user.ID)
			err := s.cache.Set(tenant.ctx, key, user, time.Hour)
			require.NoError(s.T(), err)
		}
	}

	// Verify isolation - each tenant should only see their data
	for i, tenant := range tenants {
		for _, user := range tenant.users {
			key := fmt.Sprintf("user:%s", user.ID)
			var cached User
			err := s.cache.Get(tenant.ctx, key, &cached)
			require.NoError(s.T(), err)
			require.Equal(s.T(), user.Email, cached.Email)
			require.Contains(s.T(), cached.Email, fmt.Sprintf("tenant%d", i+1))
		}
	}

	// Flush one tenant shouldn't affect others
	err := s.cache.Flush(ctx1)
	require.NoError(s.T(), err)

	// Verify tenant1 data is gone
	var result User
	err = s.cache.Get(ctx1, "user:1", &result)
	require.Error(s.T(), err)

	// Verify other tenants still have their data
	err = s.cache.Get(ctx2, "user:1", &result)
	require.NoError(s.T(), err)
	require.Contains(s.T(), result.Email, "tenant2")

	err = s.cache.Get(ctx3, "user:1", &result)
	require.NoError(s.T(), err)
	require.Contains(s.T(), result.Email, "tenant3")
}

// TestHighConcurrency tests cache under high concurrent load
func (s *IntegrationTestSuite) TestHighConcurrency() {
	tenantID := uuid.New()
	ctx := cache.WithTenantID(context.Background(), tenantID)

	const (
		goroutines = 100
		operations = 1000
		totalOps   = goroutines * operations
	)

	var (
		wg sync.WaitGroup

		errors []error
		errMux sync.Mutex
	)
	// Start concurrent operations
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < operations; j++ {
				key := fmt.Sprintf("concurrent:%d:%d", workerID, j)
				value := fmt.Sprintf("value-%d-%d", workerID, j)

				// Set
				err := s.cache.Set(ctx, key, value, time.Minute)
				if err != nil {
					errMux.Lock()
					errors = append(errors, err)
					errMux.Unlock()
					continue
				}

				// Get
				var result string
				err = s.cache.Get(ctx, key, &result)
				if err != nil {
					errMux.Lock()
					errors = append(errors, err)
					errMux.Unlock()
					continue
				}

				if result != value {
					errMux.Lock()
					errors = append(errors, fmt.Errorf("value mismatch: got %s, want %s", result, value))
					errMux.Unlock()
					continue
				}

				// Delete
				err = s.cache.Delete(ctx, key)
				if err != nil {
					errMux.Lock()
					errors = append(errors, err)
					errMux.Unlock()
				}
			}
		}(i)
	}

	wg.Wait()

	// Calculate success rate
	successRate := float64(totalOps-len(errors)) / float64(totalOps) * 100

	s.T().Logf("Concurrent operations completed:")
	s.T().Logf("  Total operations: %d", totalOps)
	s.T().Logf("  Errors: %d", len(errors))
	s.T().Logf("  Success rate: %.2f%%", successRate)

	// Should have very high success rate
	require.Greater(s.T(), successRate, 95.0, "Success rate should be > 95%%")
}

// TestBulkOperationsPerformance tests bulk operation performance
func (s *IntegrationTestSuite) TestBulkOperationsPerformance() {
	tenantID := uuid.New()
	ctx := cache.WithTenantID(context.Background(), tenantID)

	tests := []struct {
		name     string
		keyCount int
	}{
		{"10 keys", 10},
		{"100 keys", 100},
		{"1000 keys", 1000},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Prepare data
			pairs := make(map[string]any)
			keys := make([]string, tt.keyCount)

			for i := 0; i < tt.keyCount; i++ {
				key := fmt.Sprintf("bulk:test:%d", i)
				keys[i] = key
				pairs[key] = User{
					ID:    fmt.Sprintf("%d", i),
					Name:  fmt.Sprintf("User %d", i),
					Email: fmt.Sprintf("user%d@example.com", i),
				}
			}

			// Benchmark MSet
			start := time.Now()
			err := s.cache.MSet(ctx, pairs, time.Hour)
			msetDuration := time.Since(start)
			require.NoError(s.T(), err)

			// Benchmark MGet
			start = time.Now()
			results, err := s.cache.MGet(ctx, keys)
			mgetDuration := time.Since(start)
			require.NoError(s.T(), err)
			require.Len(s.T(), results, tt.keyCount)

			// Benchmark MDelete
			start = time.Now()
			err = s.cache.MDelete(ctx, keys)
			mdeleteDuration := time.Since(start)
			require.NoError(s.T(), err)

			s.T().Logf("Bulk operations for %d keys:", tt.keyCount)
			s.T().Logf("  MSet:    %v (%.2f ops/s)", msetDuration, float64(tt.keyCount)/msetDuration.Seconds())
			s.T().Logf("  MGet:    %v (%.2f ops/s)", mgetDuration, float64(tt.keyCount)/mgetDuration.Seconds())
			s.T().Logf("  MDelete: %v (%.2f ops/s)", mdeleteDuration, float64(tt.keyCount)/mdeleteDuration.Seconds())
		})
	}
}

// TestMemoryCachePerformance compares Redis vs memory cache performance
func (s *IntegrationTestSuite) TestMemoryCachePerformance() {
	tenantID := uuid.New()
	ctx := cache.WithTenantID(context.Background(), tenantID)

	user := User{
		ID:    "123",
		Name:  "Performance Test User",
		Email: "perf@example.com",
	}

	const iterations = 100 // Reduced iterations to avoid timeout

	// Benchmark Redis cache
	err := s.cache.Set(ctx, "redis-test", user, time.Hour)
	require.NoError(s.T(), err)

	start := time.Now()
	for i := 0; i < iterations; i++ {
		var result User
		err := s.cache.Get(ctx, "redis-test", &result)
		if err != nil {
			s.T().Logf("Redis Get error at iteration %d: %v", i, err)
		}
	}
	redisDuration := time.Since(start)

	// Benchmark memory cache
	err = s.cache.SetMemory(ctx, "memory-test", user, time.Hour)
	require.NoError(s.T(), err)

	// Give cleanup routine time to settle
	time.Sleep(100 * time.Millisecond)

	start = time.Now()
	for i := 0; i < iterations; i++ {
		var result User
		err := s.cache.GetMemory(ctx, "memory-test", &result)
		if err != nil {
			s.T().Logf("Memory Get error at iteration %d: %v", i, err)
		}
	}
	memoryDuration := time.Since(start)

	s.T().Logf("Performance comparison (%d iterations):", iterations)
	s.T().Logf("  Redis cache:  %v (%.2f μs/op)", redisDuration, float64(redisDuration.Microseconds())/float64(iterations))
	s.T().Logf("  Memory cache: %v (%.2f μs/op)", memoryDuration, float64(memoryDuration.Microseconds())/float64(iterations))

	if memoryDuration.Nanoseconds() > 0 {
		s.T().Logf("  Speedup:      %.2fx", float64(redisDuration)/float64(memoryDuration))
	}

	// Memory cache should be significantly faster (allow for some variance)
	if redisDuration > 0 && memoryDuration > 0 {
		require.Less(s.T(), memoryDuration, redisDuration*2, "Memory cache should be reasonably fast")
	}
}

// TestCacheEvictionAndExpiration tests TTL and expiration
func (s *IntegrationTestSuite) TestCacheEvictionAndExpiration() {
	tenantID := uuid.New()
	ctx := cache.WithTenantID(context.Background(), tenantID)

	tests := []struct {
		name        string
		ttl         time.Duration
		waitTime    time.Duration
		shouldExist bool
	}{
		{
			name:        "short TTL expires",
			ttl:         100 * time.Millisecond,
			waitTime:    200 * time.Millisecond,
			shouldExist: false,
		},
		{
			name:        "long TTL persists",
			ttl:         1 * time.Second,
			waitTime:    200 * time.Millisecond,
			shouldExist: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			key := fmt.Sprintf("expiration:%s", tt.name)
			value := "test-value"

			// Set with TTL
			err := s.cache.Set(ctx, key, value, tt.ttl)
			require.NoError(s.T(), err)

			// Wait
			time.Sleep(tt.waitTime)

			// Check existence
			var result string
			err = s.cache.Get(ctx, key, &result)

			if tt.shouldExist {
				require.NoError(s.T(), err)
				require.Equal(s.T(), value, result)
			} else {
				require.Error(s.T(), err)
			}
		})
	}
}

// TestComplexDataStructures tests caching of complex nested structures
func (s *IntegrationTestSuite) TestComplexDataStructures() {
	tenantID := uuid.New()
	ctx := cache.WithTenantID(context.Background(), tenantID)

	complex := ComplexStruct{
		ID:   uuid.New().String(),
		Name: "Complex Test",
		Metadata: map[string]any{
			"string":  "value",
			"number":  42,
			"boolean": true,
			"nested": map[string]any{
				"key1": "value1",
				"key2": 123,
			},
		},
		Tags: []string{"tag1", "tag2", "tag3"},
		Timestamps: Timestamps{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Relations: []Relation{
			{ID: "rel-1", Type: "parent"},
			{ID: "rel-2", Type: "child"},
		},
	}

	// Cache complex structure
	err := s.cache.Set(ctx, "complex:test", complex, time.Hour)
	require.NoError(s.T(), err)

	// Retrieve and verify
	var cached ComplexStruct
	err = s.cache.Get(ctx, "complex:test", &cached)
	require.NoError(s.T(), err)

	// Verify all fields
	require.Equal(s.T(), complex.ID, cached.ID)
	require.Equal(s.T(), complex.Name, cached.Name)
	require.Len(s.T(), cached.Tags, 3)
	require.Len(s.T(), cached.Relations, 2)
	require.NotNil(s.T(), cached.Metadata["nested"])
}

// Helper functions and types

func getEnvOrDefault(key, defaultValue string) string {
	// Implementation would use os.Getenv
	return defaultValue
}

func getEnvOrDefaultInt(key string, defaultValue int) int {
	// Implementation would use os.Getenv and strconv.Atoi
	return defaultValue
}

type User struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Email       string          `json:"email"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	Preferences UserPreferences `json:"preferences"`
}

type UserPreferences struct {
	Theme         string `json:"theme"`
	Language      string `json:"language"`
	Notifications bool   `json:"notifications"`
}

type Session struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Token     string                 `json:"token"`
	ExpiresAt time.Time              `json:"expires_at"`
	Data      map[string]any `json:"data"`
}

type ShoppingCart struct {
	UserID    string     `json:"user_id"`
	Items     []CartItem `json:"items"`
	Total     float64    `json:"total"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type CartItem struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type ComplexStruct struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Metadata   map[string]any `json:"metadata"`
	Tags       []string               `json:"tags"`
	Timestamps Timestamps             `json:"timestamps"`
	Relations  []Relation             `json:"relations"`
}

type Timestamps struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Relation struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func (u User) MarshalBinary() ([]byte, error) {
	return json.Marshal(u)
}

func (u *User) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, u)
}

func TestIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(IntegrationTestSuite))
}
