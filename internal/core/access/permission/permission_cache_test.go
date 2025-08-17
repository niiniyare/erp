package permission

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// // MockCache implements cache.Service for testing
//
//	type MockCache struct {
//		mock.Mock
//	}
//
//	func (m *MockCache) Get(ctx context.Context, key string, dest interface{}) error {
//		args := m.Called(ctx, key, dest)
//		if f, ok := args.Get(0).(func(interface{})); ok {
//			f(dest)
//			return nil
//		}
//		return args.Error(0)
//	}
//
//	func (m *MockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
//		args := m.Called(ctx, key, value, ttl)
//		return args.Error(0)
//	}
//
//	func (m *MockCache) Delete(ctx context.Context, key string) error {
//		args := m.Called(ctx, key)
//		return args.Error(0)
//	}
//
//	func (m *MockCache) Exists(ctx context.Context, key string) (bool, error) {
//		args := m.Called(ctx, key)
//		return args.Bool(0), args.Error(1)
//	}
//
//	func (m *MockCache) Clear(ctx context.Context) error {
//		args := m.Called(ctx)
//		return args.Error(0)
//	}
//
//	func (m *MockCache) Flush(ctx context.Context) error {
//		args := m.Called(ctx)
//		return args.Error(0)
//	}
//
//	func (m *MockCache) MGet(ctx context.Context, keys []string, dest interface{}) error {
//		args := m.Called(ctx, keys, dest)
//		return args.Error(0)
//	}
//
//	func (m *MockCache) MSet(ctx context.Context, pairs map[string]interface{}, expiration time.Duration) error {
//		args := m.Called(ctx, pairs, expiration)
//		return args.Error(0)
//	}
//
//	func (m *MockCache) MDelete(ctx context.Context, keys []string) error {
//		args := m.Called(ctx, keys)
//		return args.Error(0)
//	}
//
//	func (m *MockCache) DeletePattern(ctx context.Context, pattern string) error {
//		args := m.Called(ctx, pattern)
//		return args.Error(0)
//	}
//
//	func (m *MockCache) TTL(ctx context.Context, key string) (time.Duration, error) {
//		args := m.Called(ctx, key)
//		return args.Get(0).(time.Duration), args.Error(1)
//	}
//
//	func (m *MockCache) Ping(ctx context.Context) error {
//		args := m.Called(ctx)
//		return args.Error(0)
//	}
//
//	func (m *MockCache) Stats() *cache.CacheStats {
//		args := m.Called()
//		return args.Get(0).(*cache.CacheStats)
//	}
//
//	func (m *MockCache) Close() error {
//		args := m.Called()
//		return args.Error(0)
//	}
//
// setupPermissionCacheService creates a PermissionCacheService with mock dependencies for testing
func setupPermissionCacheService() (*PermissionCacheService, *MockCache) {
	mockCache := new(MockCache)

	// Create minimal working metrics service for testing
	mockMetrics, _ := metrics.NewMetricsService(metrics.MetricsConfig{
		Namespace: "test",
		Subsystem: "cache",
		Provider:  "noop",
		Enabled:   false,
	})

	// Create minimal working tracing service for testing
	mockTracing, _ := tracing.NewTracingService(tracing.TracingConfig{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		ExporterType:   tracing.NoopExporter,
		SamplingRatio:  0.0,
	})

	service := NewPermissionCacheService(mockCache, mockMetrics, mockTracing)
	return service, mockCache
}

func TestPermissionCacheService(t *testing.T) {
	// Setup
	service, mockCache := setupPermissionCacheService()
	ctx := context.Background()

	// Test data
	userID := uuid.New()
	resourceName := "documents"
	actionName := "read"
	entityID := uuid.New()

	req := &PermissionEvaluationRequest{
		UserID:       userID,
		ResourceName: resourceName,
		ActionName:   actionName,
		EntityID:     &entityID,
		Context:      map[string]any{"department": "engineering"},
	}

	result := &PermissionEvaluationResult{
		Allowed:          true,
		EvaluationTimeMS: 25,
		CacheHit:         false,
		PolicyDecisions:  []string{"allow"},
	}

	t.Run("SetPermissionEvaluationResult", func(t *testing.T) {
		// Mock cache set call
		mockCache.On("Set", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("permission.CachedPermissionEvaluationResult"), MediumCacheTTL).Return(nil)

		err := service.SetPermissionEvaluationResult(ctx, req, result)
		assert.NoError(t, err)

		mockCache.AssertExpectations(t)
	})

	t.Run("GetPermissionEvaluationResult_Hit", func(t *testing.T) {
		// Setup cached result
		cachedResult := CachedPermissionEvaluationResult{
			Result:       result,
			CachedAt:     time.Now().Add(-5 * time.Minute),
			ExpiresAt:    time.Now().Add(10 * time.Minute),
			CacheVersion: "1.0",
			UserID:       userID,
			ResourceName: resourceName,
			ActionName:   actionName,
			EntityID:     &entityID,
		}

		// Mock cache get call
		mockCache.On("Get", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("*permission.CachedPermissionEvaluationResult")).Return(func(dest interface{}) {
			*(dest.(*CachedPermissionEvaluationResult)) = cachedResult
		})

		retrievedResult, found, err := service.GetPermissionEvaluationResult(ctx, req)
		assert.NoError(t, err)
		assert.True(t, found)
		assert.NotNil(t, retrievedResult)
		assert.True(t, retrievedResult.CacheHit)
		assert.Equal(t, result.Allowed, retrievedResult.Allowed)

		mockCache.AssertExpectations(t)
	})

	t.Run("GetPermissionEvaluationResult_Miss", func(t *testing.T) {
		// Mock cache miss
		mockCache.On("Get", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("*permission.CachedPermissionEvaluationResult")).Return(cache.ErrCacheMiss)

		retrievedResult, found, err := service.GetPermissionEvaluationResult(ctx, req)
		assert.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrievedResult)

		mockCache.AssertExpectations(t)
	})

	t.Run("GetPermissionEvaluationResult_Expired", func(t *testing.T) {
		// Setup expired cached result
		expiredResult := CachedPermissionEvaluationResult{
			Result:       result,
			CachedAt:     time.Now().Add(-30 * time.Minute),
			ExpiresAt:    time.Now().Add(-10 * time.Minute), // Expired
			CacheVersion: "1.0",
			UserID:       userID,
			ResourceName: resourceName,
			ActionName:   actionName,
			EntityID:     &entityID,
		}

		// Mock cache get call
		mockCache.On("Get", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("*permission.CachedPermissionEvaluationResult")).Return(func(dest interface{}) {
			*(dest.(*CachedPermissionEvaluationResult)) = expiredResult
		})

		// Mock cache delete call for expired entry
		mockCache.On("Delete", mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("string")).Return(nil)

		retrievedResult, found, err := service.GetPermissionEvaluationResult(ctx, req)
		assert.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrievedResult)

		mockCache.AssertExpectations(t)
	})
}

func TestPermissionCacheService_UserPermissions(t *testing.T) {
	// Setup
	service, mockCache := setupPermissionCacheService()
	ctx := context.Background()

	// Test data
	userID := uuid.New()
	entityID := uuid.New()

	permissions := []*EffectivePermission{
		{
			Permission: &Permission{
				ID:         uuid.New(),
				Name:       "read_documents",
				ResourceID: uuid.New(),
				ActionID:   uuid.New(),
				Effect:     "ALLOW",
			},
			GrantedByRole: &Role{
				ID:   uuid.New(),
				Name: "Editor",
			},
			EntityID:       entityID,
			AssignmentType: "direct",
		},
	}

	t.Run("SetUserPermissions", func(t *testing.T) {
		// Mock cache set call
		mockCache.On("Set", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("permission.CachedUserPermissions"), LongCacheTTL).Return(nil)

		err := service.SetUserPermissions(ctx, userID, &entityID, permissions)
		assert.NoError(t, err)

		mockCache.AssertExpectations(t)
	})

	t.Run("GetUserPermissions_Hit", func(t *testing.T) {
		// Setup cached permissions
		cachedPermissions := CachedUserPermissions{
			UserID:       userID,
			EntityID:     &entityID,
			Permissions:  permissions,
			CachedAt:     time.Now().Add(-5 * time.Minute),
			ExpiresAt:    time.Now().Add(55 * time.Minute),
			CacheVersion: "1.0",
		}

		// Mock cache get call
		mockCache.On("Get", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("*permission.CachedUserPermissions")).Return(func(dest interface{}) {
			*(dest.(*CachedUserPermissions)) = cachedPermissions
		})

		retrievedPermissions, found, err := service.GetUserPermissions(ctx, userID, &entityID)
		assert.NoError(t, err)
		assert.True(t, found)
		assert.NotNil(t, retrievedPermissions)
		assert.Equal(t, len(permissions), len(retrievedPermissions))
		assert.Equal(t, permissions[0].Permission.Name, retrievedPermissions[0].Permission.Name)

		mockCache.AssertExpectations(t)
	})
}

func TestPermissionCacheService_UserRoles(t *testing.T) {
	// Setup
	service, mockCache := setupPermissionCacheService()
	ctx := context.Background()

	// Test data
	userID := uuid.New()
	roleID := uuid.New()
	entityID := uuid.New()

	roles := []*UserRole{
		{
			ID:             uuid.New(),
			UserID:         userID,
			RoleID:         roleID,
			EntityID:       entityID,
			AssignmentType: "direct",
			AssignedAt:     time.Now().Add(-1 * time.Hour),
			IsActive:       true,
		},
	}

	t.Run("SetUserRoles", func(t *testing.T) {
		// Mock cache set call
		mockCache.On("Set", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("permission.CachedUserRoles"), LongCacheTTL).Return(nil)

		err := service.SetUserRoles(ctx, userID, roles)
		assert.NoError(t, err)

		mockCache.AssertExpectations(t)
	})

	t.Run("GetUserRoles_Hit", func(t *testing.T) {
		// Setup cached roles
		cachedRoles := CachedUserRoles{
			UserID:       userID,
			Roles:        roles,
			CachedAt:     time.Now().Add(-5 * time.Minute),
			ExpiresAt:    time.Now().Add(55 * time.Minute),
			CacheVersion: "1.0",
		}

		// Mock cache get call
		mockCache.On("Get", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("*permission.CachedUserRoles")).Return(func(dest interface{}) {
			*(dest.(*CachedUserRoles)) = cachedRoles
		})

		retrievedRoles, found, err := service.GetUserRoles(ctx, userID)
		assert.NoError(t, err)
		assert.True(t, found)
		assert.NotNil(t, retrievedRoles)
		assert.Equal(t, len(roles), len(retrievedRoles))
		assert.Equal(t, roles[0].RoleID, retrievedRoles[0].RoleID)

		mockCache.AssertExpectations(t)
	})
}

func TestPermissionCacheService_RoleHierarchy(t *testing.T) {
	// Setup
	service, mockCache := setupPermissionCacheService()
	ctx := context.Background()

	// Test data
	roleID := uuid.New()
	parentRoleID := uuid.New()

	hierarchy := []*Role{
		{
			ID:           roleID,
			Name:         "Editor",
			ParentRoleID: &parentRoleID,
		},
		{
			ID:           parentRoleID,
			Name:         "Viewer",
			ParentRoleID: nil,
		},
	}

	t.Run("SetRoleHierarchy", func(t *testing.T) {
		// Mock cache set call
		mockCache.On("Set", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("permission.CachedRoleHierarchy"), LongCacheTTL).Return(nil)

		err := service.SetRoleHierarchy(ctx, roleID, hierarchy)
		assert.NoError(t, err)

		mockCache.AssertExpectations(t)
	})

	t.Run("GetRoleHierarchy_Hit", func(t *testing.T) {
		// Setup cached hierarchy
		cachedHierarchy := CachedRoleHierarchy{
			RoleID:       roleID,
			Hierarchy:    hierarchy,
			CachedAt:     time.Now().Add(-5 * time.Minute),
			ExpiresAt:    time.Now().Add(55 * time.Minute),
			CacheVersion: "1.0",
		}

		// Mock cache get call
		mockCache.On("Get", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("*permission.CachedRoleHierarchy")).Return(func(dest interface{}) {
			*(dest.(*CachedRoleHierarchy)) = cachedHierarchy
		})

		retrievedHierarchy, found, err := service.GetRoleHierarchy(ctx, roleID)
		assert.NoError(t, err)
		assert.True(t, found)
		assert.NotNil(t, retrievedHierarchy)
		assert.Equal(t, len(hierarchy), len(retrievedHierarchy))
		assert.Equal(t, hierarchy[0].Name, retrievedHierarchy[0].Name)

		mockCache.AssertExpectations(t)
	})
}

func TestPermissionCacheService_CacheInvalidation(t *testing.T) {
	// Setup
	service, mockCache := setupPermissionCacheService()
	ctx := context.Background()

	// Test data
	userID := uuid.New()
	roleID := uuid.New()

	t.Run("InvalidateUserPermissions", func(t *testing.T) {
		// Mock cache delete calls
		mockCache.On("Delete", ctx, mock.AnythingOfType("string")).Return(nil).Times(2)

		err := service.InvalidateUserPermissions(ctx, userID)
		assert.NoError(t, err)

		mockCache.AssertExpectations(t)
	})

	t.Run("InvalidateRoleCache", func(t *testing.T) {
		// Mock cache delete calls
		mockCache.On("Delete", ctx, mock.AnythingOfType("string")).Return(nil).Times(2)

		err := service.InvalidateRoleCache(ctx, roleID)
		assert.NoError(t, err)

		mockCache.AssertExpectations(t)
	})

	t.Run("FlushPermissionCache", func(t *testing.T) {
		// Mock cache flush call
		mockCache.On("Flush", ctx).Return(nil)

		err := service.FlushPermissionCache(ctx)
		assert.NoError(t, err)

		mockCache.AssertExpectations(t)
	})
}

func TestPermissionCacheService_GeneratePermissionEvaluationKey(t *testing.T) {
	// Setup
	service, _ := setupPermissionCacheService()

	// Test data
	userID := uuid.New()
	entityID := uuid.New()

	req1 := &PermissionEvaluationRequest{
		UserID:       userID,
		ResourceName: "documents",
		ActionName:   "read",
		EntityID:     &entityID,
		Context:      map[string]any{"department": "engineering"},
	}

	req2 := &PermissionEvaluationRequest{
		UserID:       userID,
		ResourceName: "documents",
		ActionName:   "read",
		EntityID:     &entityID,
		Context:      map[string]any{"department": "engineering"},
	}

	req3 := &PermissionEvaluationRequest{
		UserID:       userID,
		ResourceName: "documents",
		ActionName:   "write", // Different action
		EntityID:     &entityID,
		Context:      map[string]any{"department": "engineering"},
	}

	t.Run("GenerateConsistentKeys", func(t *testing.T) {
		key1 := service.GeneratePermissionEvaluationKey(req1)
		key2 := service.GeneratePermissionEvaluationKey(req2)
		key3 := service.GeneratePermissionEvaluationKey(req3)

		// Same requests should generate same keys
		assert.Equal(t, key1, key2)

		// Different requests should generate different keys
		assert.NotEqual(t, key1, key3)

		// Keys should have the expected format
		assert.Contains(t, key1, "perm:eval:")
		assert.Contains(t, key3, "perm:eval:")
	})
}

func TestPermissionCacheService_GetCacheStats(t *testing.T) {
	// Setup
	service, _ := setupPermissionCacheService()
	ctx := context.Background()

	stats := service.GetCacheStats(ctx)

	assert.NotNil(t, stats)
	assert.Equal(t, "permission_cache", stats["cache_type"])
	assert.Equal(t, "active", stats["status"])
	assert.Contains(t, stats, "ttl_config")

	ttlConfig := stats["ttl_config"].(map[string]any)
	assert.Equal(t, ShortCacheTTL.String(), ttlConfig["short_ttl"])
	assert.Equal(t, MediumCacheTTL.String(), ttlConfig["medium_ttl"])
	assert.Equal(t, LongCacheTTL.String(), ttlConfig["long_ttl"])
}
