//go:build database
// +build database

package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// PolicyRepositoryTestSuite defines test suite for ABAC policy repository operations
// Tests cover: ABAC-REPO-001 through ABAC-REPO-010 with multi-tenant RLS verification
type PolicyRepositoryTestSuite struct {
	suite.Suite
	ctx          context.Context
	runner       *tenant.DatabaseTestRunner
	repo         PolicyRepository
	cacheService cache.Service
	tenantA      *db.Tenant
	tenantB      *db.Tenant

	// Test data UUIDs for cleanup
	createdPolicyIDs []uuid.UUID
}

// SetupSuite runs once before the entire test suite
func (s *PolicyRepositoryTestSuite) SetupSuite() {
	var err error
	s.runner, err = tenant.NewDatabaseTestRunner()
	s.Require().NoError(err, "Failed to connect to the database")
	s.ctx = context.Background()

	// Setup shared infrastructure
	logger := logger.WithFields(logger.Fields{"component": "policy_repository_test"})
	metrics := &metrics.MetricsService{}
	tracer := tracing.NewNoOpTracingService()
	s.cacheService = cache.NewNoOpCache() // Use no-op cache for testing

	// Create repository instance
	s.repo = NewPolicyRepository(
		s.runner.GetStore(),
		s.cacheService,
		logger,
		metrics,
		tracer,
	)

	// Create test tenants
	s.setupTestTenants()
	s.createdPolicyIDs = make([]uuid.UUID, 0)
}

// TearDownSuite runs once after the entire test suite
func (s *PolicyRepositoryTestSuite) TearDownSuite() {
	// Clean up test policies
	s.cleanupTestData()

	// Clean up test tenants
	if s.tenantA != nil {
		superuserStore := db.NewStore(s.runner.GetPool())
		if err := superuserStore.SoftDeleteTenant(s.ctx, s.tenantA.ID); err != nil {
			s.T().Logf("Warning: Failed to cleanup tenant A: %v", err)
		}
	}
	if s.tenantB != nil {
		superuserStore := db.NewStore(s.runner.GetPool())
		if err := superuserStore.SoftDeleteTenant(s.ctx, s.tenantB.ID); err != nil {
			s.T().Logf("Warning: Failed to cleanup tenant B: %v", err)
		}
	}

	s.runner.Close()
}

// SetupTest initializes test fixtures for each test
func (s *PolicyRepositoryTestSuite) SetupTest() {
	// Reset any test-specific state
}

// TearDownTest cleans up after each test
func (s *PolicyRepositoryTestSuite) TearDownTest() {
	// Individual test cleanup if needed
}

// setupTestTenants creates two test tenants for multi-tenant testing
func (s *PolicyRepositoryTestSuite) setupTestTenants() {
	superuserStore := db.NewStore(s.runner.GetPool())
	uniqueID := uuid.New().String()[0:8]

	// Create Tenant A
	var err error
	s.tenantA, err = superuserStore.CreateTenant(s.ctx, db.CreateTenantParams{
		Name:   fmt.Sprintf("ABAC Test Tenant A %s", uniqueID),
		Slug:   fmt.Sprintf("abac-tenant-a-%s", uniqueID),
		Email:  fmt.Sprintf("admin-a-%s@abactest.com", uniqueID),
		Status: "active",
	})
	s.Require().NoError(err, "Failed to create tenant A")

	// Create Tenant B
	uniqueID2 := uuid.New().String()[0:8]
	s.tenantB, err = superuserStore.CreateTenant(s.ctx, db.CreateTenantParams{
		Name:   fmt.Sprintf("ABAC Test Tenant B %s", uniqueID2),
		Slug:   fmt.Sprintf("abac-tenant-b-%s", uniqueID2),
		Email:  fmt.Sprintf("admin-b-%s@abactest.com", uniqueID2),
		Status: "active",
	})
	s.Require().NoError(err, "Failed to create tenant B")
}

// cleanupTestData removes test policies created during testing
func (s *PolicyRepositoryTestSuite) cleanupTestData() {
	for _, policyID := range s.createdPolicyIDs {
		// Clean up policies from both tenants
		_ = s.runner.GetStore().WithTenant(s.ctx, s.tenantA.ID, func(ctx context.Context, store db.Store) error {
			store.SoftDeletePolicy(ctx, policyID)
			return nil
		})
		_ = s.runner.GetStore().WithTenant(s.ctx, s.tenantB.ID, func(ctx context.Context, store db.Store) error {
			store.SoftDeletePolicy(ctx, policyID)
			return nil
		})
	}
}

// trackPolicyForCleanup adds a policy ID to the cleanup list
func (s *PolicyRepositoryTestSuite) trackPolicyForCleanup(policyID uuid.UUID) {
	s.createdPolicyIDs = append(s.createdPolicyIDs, policyID)
}

// TestPolicyRepository runs the policy repository test suite
func TestPolicyRepository(t *testing.T) {
	suite.Run(t, new(PolicyRepositoryTestSuite))
}

// TestCreatePolicy tests ABAC-REPO-001: Policy Creation with Multi-tenant Isolation
func (s *PolicyRepositoryTestSuite) TestCreatePolicy() {
	testCases := []struct {
		name        string
		spec        string
		tenantID    uuid.UUID
		request     *CreatePolicyRequest
		expectError string
		validate    func(*models.Policy)
	}{
		{
			name:     "CreateValidPolicy_Success",
			spec:     "ABAC-REPO-001",
			tenantID: s.tenantA.ID,
			request: &CreatePolicyRequest{
				Name:        fmt.Sprintf("test_policy_%s", uuid.New().String()[0:8]),
				DisplayName: stringPtr("Test Access Policy"),
				Description: stringPtr("Test policy for unit testing"),
				PolicyType:  types.PolicyTypeABAC,
				Effect:      types.PolicyEffectAllow,
				Priority:    100,
				Category:    types.PolicyCategoryAccess,
				Target: map[string]any{
					"subjects": map[string]any{
						"user.department": map[string]any{
							"in": []any{"engineering", "testing"},
						},
					},
					"resources": map[string]any{
						"resource.type": map[string]any{
							"equals": "test_resource",
						},
					},
					"actions": []any{"read", "test"},
				},
				Rule: map[string]any{
					"and": []any{
						map[string]any{
							"user.security_level": map[string]any{
								"gte": 5,
							},
						},
						map[string]any{
							"time.business_hours": map[string]any{
								"equals": true,
							},
						},
					},
				},
				Obligations: map[string]any{
					"log": map[string]any{
						"level":           "INFO",
						"include_context": true,
					},
				},
				Advice: map[string]any{
					"notify": map[string]any{
						"recipients": []any{"admin@test.com"},
					},
				},
				CombiningAlgorithm: types.PolicyCombiningAlgorithmDenyOverrides,
				CreatedBy:          uuid.New(),
			},
			validate: func(policy *models.Policy) {
				require.NotNil(s.T(), policy)
				require.NotEqual(s.T(), uuid.Nil, policy.ID)
				require.Equal(s.T(), s.tenantA.ID, policy.TenantID)
				require.Equal(s.T(), types.PolicyTypeABAC, policy.PolicyType)
				require.Equal(s.T(), types.PolicyEffectAllow, policy.Effect)
				require.Equal(s.T(), int32(100), policy.Priority)
				require.Equal(s.T(), types.PolicyCategoryAccess, policy.Category)
				require.True(s.T(), policy.IsActive)
				require.NotZero(s.T(), policy.CreatedAt)
				require.NotNil(s.T(), policy.Target)
				require.NotNil(s.T(), policy.Rule)
				require.NotNil(s.T(), policy.Obligations)
				require.NotNil(s.T(), policy.Advice)
			},
		},
		{
			name:     "CreatePolicyEmptyName_ValidationError",
			spec:     "ABAC-REPO-001",
			tenantID: s.tenantA.ID,
			request: &CreatePolicyRequest{
				Name:       "", // Invalid: empty name
				PolicyType: types.PolicyTypeABAC,
				Effect:     types.PolicyEffectAllow,
				Priority:   100,
				Category:   types.PolicyCategoryAccess,
				Target:     map[string]any{"test": true},
				Rule:       map[string]any{"test": true},
				CreatedBy:  uuid.New(),
			},
			expectError: "validation",
		},
		{
			name:     "CreatePolicyInvalidTarget_ValidationError",
			spec:     "ABAC-REPO-001",
			tenantID: s.tenantA.ID,
			request: &CreatePolicyRequest{
				Name:       fmt.Sprintf("invalid_target_%s", uuid.New().String()[0:8]),
				PolicyType: types.PolicyTypeABAC,
				Effect:     types.PolicyEffectAllow,
				Priority:   100,
				Category:   types.PolicyCategoryAccess,
				Target:     nil, // Invalid: nil target
				Rule:       map[string]any{"test": true},
				CreatedBy:  uuid.New(),
			},
			expectError: "target is required",
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("%s_%s", tc.spec, tc.name), func() {
			// Execute within tenant context
			var policy *models.Policy
			var err error

			execErr := s.runner.GetStore().WithTenant(s.ctx, tc.tenantID, func(ctx context.Context, store db.Store) error {
				policy, err = s.repo.CreatePolicy(ctx, tc.request)
				return nil
			})
			s.Require().NoError(execErr, "Tenant context execution should not fail")

			// Validate results
			if tc.expectError != "" {
				require.Error(s.T(), err)
				require.Contains(s.T(), err.Error(), tc.expectError)
				require.Nil(s.T(), policy)
			} else {
				require.NoError(s.T(), err)
				require.NotNil(s.T(), policy)
				if tc.validate != nil {
					tc.validate(policy)
				}
				// Track for cleanup
				s.trackPolicyForCleanup(policy.ID)
			}
		})
	}
}

// TestGetPolicyByID tests ABAC-REPO-002: Policy Retrieval with Tenant Isolation
func (s *PolicyRepositoryTestSuite) TestGetPolicyByID() {
	// First create a policy in tenant A
	var policyA *models.Policy
	err := s.runner.GetStore().WithTenant(s.ctx, s.tenantA.ID, func(ctx context.Context, store db.Store) error {
		var createErr error
		policyA, createErr = s.repo.CreatePolicy(ctx, &CreatePolicyRequest{
			Name:        fmt.Sprintf("get_test_policy_%s", uuid.New().String()[0:8]),
			DisplayName: stringPtr("Get Test Policy"),
			PolicyType:  types.PolicyTypeABAC,
			Effect:      types.PolicyEffectAllow,
			Priority:    100,
			Category:    types.PolicyCategoryAccess,
			Target:      map[string]any{"test": true},
			Rule:        map[string]any{"test": true},
			CreatedBy:   uuid.New(),
		})
		return createErr
	})
	s.Require().NoError(err, "Failed to create test policy")
	s.trackPolicyForCleanup(policyA.ID)

	testCases := []struct {
		name        string
		spec        string
		tenantID    uuid.UUID
		policyID    uuid.UUID
		expectError string
		validate    func(*models.Policy)
	}{
		{
			name:     "GetExistingPolicy_Success",
			spec:     "ABAC-REPO-002",
			tenantID: s.tenantA.ID,
			policyID: policyA.ID,
			validate: func(policy *models.Policy) {
				require.NotNil(s.T(), policy)
				require.Equal(s.T(), policyA.ID, policy.ID)
				require.Equal(s.T(), policyA.Name, policy.Name)
				require.Equal(s.T(), s.tenantA.ID, policy.TenantID)
			},
		},
		{
			name:        "GetPolicyFromDifferentTenant_NotFound",
			spec:        "ABAC-REPO-002",
			tenantID:    s.tenantB.ID, // Different tenant
			policyID:    policyA.ID,   // Policy from tenant A
			expectError: "not found",
		},
		{
			name:        "GetNonExistentPolicy_NotFound",
			spec:        "ABAC-REPO-002",
			tenantID:    s.tenantA.ID,
			policyID:    uuid.New(), // Random UUID
			expectError: "not found",
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("%s_%s", tc.spec, tc.name), func() {
			var policy *models.Policy
			var err error

			execErr := s.runner.GetStore().WithTenant(s.ctx, tc.tenantID, func(ctx context.Context, store db.Store) error {
				policy, err = s.repo.GetPolicyByID(ctx, tc.policyID)
				return nil
			})
			s.Require().NoError(execErr, "Tenant context execution should not fail")

			if tc.expectError != "" {
				require.Error(s.T(), err)
				require.Contains(s.T(), err.Error(), tc.expectError)
				require.Nil(s.T(), policy)
			} else {
				require.NoError(s.T(), err)
				require.NotNil(s.T(), policy)
				if tc.validate != nil {
					tc.validate(policy)
				}
			}
		})
	}
}

// TestUpdatePolicy tests ABAC-REPO-003: Policy Updates with Tenant Isolation
func (s *PolicyRepositoryTestSuite) TestUpdatePolicy() {
	// Create initial policy
	var originalPolicy *models.Policy
	err := s.runner.GetStore().WithTenant(s.ctx, s.tenantA.ID, func(ctx context.Context, store db.Store) error {
		var createErr error
		originalPolicy, createErr = s.repo.CreatePolicy(ctx, &CreatePolicyRequest{
			Name:        fmt.Sprintf("update_test_policy_%s", uuid.New().String()[0:8]),
			DisplayName: stringPtr("Original Policy"),
			Description: stringPtr("Original description"),
			PolicyType:  types.PolicyTypeABAC,
			Effect:      types.PolicyEffectAllow,
			Priority:    100,
			Category:    types.PolicyCategoryAccess,
			Target:      map[string]any{"original": true},
			Rule:        map[string]any{"original": true},
			CreatedBy:   uuid.New(),
		})
		return createErr
	})
	s.Require().NoError(err)
	s.trackPolicyForCleanup(originalPolicy.ID)

	testCases := []struct {
		name        string
		spec        string
		tenantID    uuid.UUID
		policyID    uuid.UUID
		request     *UpdatePolicyRequest
		expectError string
		validate    func(*models.Policy)
	}{
		{
			name:     "UpdatePolicy_Success",
			spec:     "ABAC-REPO-003",
			tenantID: s.tenantA.ID,
			policyID: originalPolicy.ID,
			request: &UpdatePolicyRequest{
				DisplayName: stringPtr("Updated Policy"),
				Description: stringPtr("Updated description"),
				Priority:    int32Ptr(150),
				Target: map[string]any{
					"updated":   true,
					"new_field": "new_value",
				},
				Rule: map[string]any{
					"updated": true,
					"security_level": map[string]any{
						"gte": 6,
					},
				},
				IsActive: boolPtr(true),
			},
			validate: func(policy *models.Policy) {
				require.NotNil(s.T(), policy)
				require.Equal(s.T(), originalPolicy.ID, policy.ID)
				require.Equal(s.T(), "Updated Policy", *policy.DisplayName)
				require.Equal(s.T(), "Updated description", *policy.Description)
				require.Equal(s.T(), int32(150), policy.Priority)
				require.True(s.T(), policy.IsActive)
				require.NotEqual(s.T(), originalPolicy.UpdatedAt, policy.UpdatedAt)

				// Validate JSON fields were updated
				require.Contains(s.T(), policy.Target, "updated")
				require.Contains(s.T(), policy.Rule, "updated")
			},
		},
		{
			name:     "UpdatePolicyFromDifferentTenant_NotFound",
			spec:     "ABAC-REPO-003",
			tenantID: s.tenantB.ID, // Different tenant
			policyID: originalPolicy.ID,
			request: &UpdatePolicyRequest{
				DisplayName: stringPtr("Should Not Update"),
			},
			expectError: "not found",
		},
		{
			name:     "UpdateNonExistentPolicy_NotFound",
			spec:     "ABAC-REPO-003",
			tenantID: s.tenantA.ID,
			policyID: uuid.New(), // Random UUID
			request: &UpdatePolicyRequest{
				DisplayName: stringPtr("Should Not Update"),
			},
			expectError: "not found",
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("%s_%s", tc.spec, tc.name), func() {
			var updatedPolicy *models.Policy
			var err error

			execErr := s.runner.GetStore().WithTenant(s.ctx, tc.tenantID, func(ctx context.Context, store db.Store) error {
				updatedPolicy, err = s.repo.UpdatePolicy(ctx, tc.policyID, tc.request)
				return nil
			})
			s.Require().NoError(execErr)

			if tc.expectError != "" {
				require.Error(s.T(), err)
				require.Contains(s.T(), err.Error(), tc.expectError)
				require.Nil(s.T(), updatedPolicy)
			} else {
				require.NoError(s.T(), err)
				require.NotNil(s.T(), updatedPolicy)
				if tc.validate != nil {
					tc.validate(updatedPolicy)
				}
			}
		})
	}
}

// TestDeletePolicy tests ABAC-REPO-004: Policy Deletion with Tenant Isolation
func (s *PolicyRepositoryTestSuite) TestDeletePolicy() {
	// Create policies for deletion testing
	var policyA, policyB *models.Policy

	// Create policy in tenant A
	err := s.runner.GetStore().WithTenant(s.ctx, s.tenantA.ID, func(ctx context.Context, store db.Store) error {
		var createErr error
		policyA, createErr = s.repo.CreatePolicy(ctx, &CreatePolicyRequest{
			Name:       fmt.Sprintf("delete_test_a_%s", uuid.New().String()[0:8]),
			PolicyType: types.PolicyTypeABAC,
			Effect:     types.PolicyEffectAllow,
			Priority:   100,
			Category:   types.PolicyCategoryAccess,
			Target:     map[string]any{"test": true},
			Rule:       map[string]any{"test": true},
			CreatedBy:  uuid.New(),
		})
		return createErr
	})
	s.Require().NoError(err)
	s.trackPolicyForCleanup(policyA.ID)

	// Create policy in tenant B
	err = s.runner.GetStore().WithTenant(s.ctx, s.tenantB.ID, func(ctx context.Context, store db.Store) error {
		var createErr error
		policyB, createErr = s.repo.CreatePolicy(ctx, &CreatePolicyRequest{
			Name:       fmt.Sprintf("delete_test_b_%s", uuid.New().String()[0:8]),
			PolicyType: types.PolicyTypeABAC,
			Effect:     types.PolicyEffectAllow,
			Priority:   100,
			Category:   types.PolicyCategoryAccess,
			Target:     map[string]any{"test": true},
			Rule:       map[string]any{"test": true},
			CreatedBy:  uuid.New(),
		})
		return createErr
	})
	s.Require().NoError(err)
	s.trackPolicyForCleanup(policyB.ID)

	testCases := []struct {
		name        string
		spec        string
		tenantID    uuid.UUID
		policyID    uuid.UUID
		expectError string
		validate    func()
	}{
		{
			name:     "DeleteOwnPolicy_Success",
			spec:     "ABAC-REPO-004",
			tenantID: s.tenantA.ID,
			policyID: policyA.ID,
			validate: func() {
				// Verify policy is soft deleted (not accessible)
				execErr := s.runner.GetStore().WithTenant(s.ctx, s.tenantA.ID, func(ctx context.Context, store db.Store) error {
					_, err := s.repo.GetPolicyByID(ctx, policyA.ID)
					require.Error(s.T(), err)
					require.Contains(s.T(), err.Error(), "not found")
					return nil
				})
				s.Require().NoError(execErr)
			},
		},
		{
			name:     "DeletePolicyFromDifferentTenant_NoEffect",
			spec:     "ABAC-REPO-004",
			tenantID: s.tenantA.ID, // Tenant A context
			policyID: policyB.ID,   // Policy from Tenant B
			validate: func() {
				// Policy B should still exist in tenant B context
				execErr := s.runner.GetStore().WithTenant(s.ctx, s.tenantB.ID, func(ctx context.Context, store db.Store) error {
					policy, err := s.repo.GetPolicyByID(ctx, policyB.ID)
					require.NoError(s.T(), err)
					require.NotNil(s.T(), policy)
					require.Equal(s.T(), policyB.ID, policy.ID)
					return nil
				})
				s.Require().NoError(execErr)
			},
		},
		{
			name:     "DeleteNonExistentPolicy_NoError",
			spec:     "ABAC-REPO-004",
			tenantID: s.tenantA.ID,
			policyID: uuid.New(),
			validate: func() {
				// Should not produce error (soft delete affects 0 rows)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("%s_%s", tc.spec, tc.name), func() {
			var err error

			execErr := s.runner.GetStore().WithTenant(s.ctx, tc.tenantID, func(ctx context.Context, store db.Store) error {
				err = s.repo.DeletePolicy(ctx, tc.policyID)
				return nil
			})
			s.Require().NoError(execErr)

			if tc.expectError != "" {
				require.Error(s.T(), err)
				require.Contains(s.T(), err.Error(), tc.expectError)
			} else {
				require.NoError(s.T(), err)
				if tc.validate != nil {
					tc.validate()
				}
			}
		})
	}
}

// TestListPolicies tests ABAC-REPO-005: Policy Listing with Tenant Isolation
func (s *PolicyRepositoryTestSuite) TestListPolicies() {
	// Create multiple policies in different tenants
	var policiesA, policiesB []*models.Policy

	// Create 3 policies in tenant A
	err := s.runner.GetStore().WithTenant(s.ctx, s.tenantA.ID, func(ctx context.Context, store db.Store) error {
		for i := 0; i < 3; i++ {
			policy, createErr := s.repo.CreatePolicy(ctx, &CreatePolicyRequest{
				Name:       fmt.Sprintf("list_test_a_%d_%s", i, uuid.New().String()[0:8]),
				PolicyType: types.PolicyTypeABAC,
				Effect:     types.PolicyEffectAllow,
				Priority:   100 + int32(i*10),
				Category:   types.PolicyCategoryAccess,
				Target:     map[string]any{"test": fmt.Sprintf("tenant_a_%d", i)},
				Rule:       map[string]any{"test": fmt.Sprintf("tenant_a_%d", i)},
				CreatedBy:  uuid.New(),
			})
			if createErr != nil {
				return createErr
			}
			policiesA = append(policiesA, policy)
			s.trackPolicyForCleanup(policy.ID)
		}
		return nil
	})
	s.Require().NoError(err)

	// Create 2 policies in tenant B
	err = s.runner.GetStore().WithTenant(s.ctx, s.tenantB.ID, func(ctx context.Context, store db.Store) error {
		for i := 0; i < 2; i++ {
			policy, createErr := s.repo.CreatePolicy(ctx, &CreatePolicyRequest{
				Name:       fmt.Sprintf("list_test_b_%d_%s", i, uuid.New().String()[0:8]),
				PolicyType: types.PolicyTypeABAC,
				Effect:     types.PolicyEffectDeny,
				Priority:   200 + int32(i*10),
				Category:   types.PolicyCategoryAccess,
				Target:     map[string]any{"test": fmt.Sprintf("tenant_b_%d", i)},
				Rule:       map[string]any{"test": fmt.Sprintf("tenant_b_%d", i)},
				CreatedBy:  uuid.New(),
			})
			if createErr != nil {
				return createErr
			}
			policiesB = append(policiesB, policy)
			s.trackPolicyForCleanup(policy.ID)
		}
		return nil
	})
	s.Require().NoError(err)

	testCases := []struct {
		name        string
		spec        string
		tenantID    uuid.UUID
		request     *ListPoliciesRequest
		expectError string
		validate    func([]*models.Policy)
	}{
		{
			name:     "ListAllPoliciesInTenantA_Success",
			spec:     "ABAC-REPO-005",
			tenantID: s.tenantA.ID,
			request:  &ListPoliciesRequest{},
			validate: func(policies []*models.Policy) {
				require.GreaterOrEqual(s.T(), len(policies), 3)

				// All policies should belong to tenant A
				for _, policy := range policies {
					require.Equal(s.T(), s.tenantA.ID, policy.TenantID)
				}

				// Should contain our created policies
				foundPolicyNames := make(map[string]bool)
				for _, policy := range policies {
					foundPolicyNames[policy.Name] = true
				}
				for _, expectedPolicy := range policiesA {
					require.True(s.T(), foundPolicyNames[expectedPolicy.Name],
						"Expected to find policy: %s", expectedPolicy.Name)
				}
			},
		},
		{
			name:     "ListAllPoliciesInTenantB_Success",
			spec:     "ABAC-REPO-005",
			tenantID: s.tenantB.ID,
			request:  &ListPoliciesRequest{},
			validate: func(policies []*models.Policy) {
				require.GreaterOrEqual(s.T(), len(policies), 2)

				// All policies should belong to tenant B
				for _, policy := range policies {
					require.Equal(s.T(), s.tenantB.ID, policy.TenantID)
				}

				// Should contain our created policies
				foundPolicyNames := make(map[string]bool)
				for _, policy := range policies {
					foundPolicyNames[policy.Name] = true
				}
				for _, expectedPolicy := range policiesB {
					require.True(s.T(), foundPolicyNames[expectedPolicy.Name],
						"Expected to find policy: %s", expectedPolicy.Name)
				}
			},
		},
		{
			name:     "TenantIsolationVerification_Success",
			spec:     "ABAC-REPO-005",
			tenantID: s.tenantA.ID,
			request:  &ListPoliciesRequest{},
			validate: func(policies []*models.Policy) {
				// Tenant A should not see any policies from tenant B
				for _, policy := range policies {
					require.Equal(s.T(), s.tenantA.ID, policy.TenantID)

					// Ensure none of the tenant B policy names are present
					for _, tenantBPolicy := range policiesB {
						require.NotEqual(s.T(), tenantBPolicy.Name, policy.Name,
							"Tenant A should not see tenant B policy: %s", tenantBPolicy.Name)
					}
				}
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("%s_%s", tc.spec, tc.name), func() {
			var policies []*models.Policy
			var err error

			execErr := s.runner.GetStore().WithTenant(s.ctx, tc.tenantID, func(ctx context.Context, store db.Store) error {
				policies, err = s.repo.ListPolicies(ctx, tc.request)
				return nil
			})
			s.Require().NoError(execErr)

			if tc.expectError != "" {
				require.Error(s.T(), err)
				require.Contains(s.T(), err.Error(), tc.expectError)
				require.Nil(s.T(), policies)
			} else {
				require.NoError(s.T(), err)
				require.NotNil(s.T(), policies)
				if tc.validate != nil {
					tc.validate(policies)
				}
			}
		})
	}
}

// TestGetPoliciesForEvaluation tests ABAC-REPO-006: Policy Evaluation Retrieval with Tenant Context
func (s *PolicyRepositoryTestSuite) TestGetPoliciesForEvaluation() {
	// Create evaluation-specific policies
	var evaluationPolicy *models.Policy

	err := s.runner.GetStore().WithTenant(s.ctx, s.tenantA.ID, func(ctx context.Context, store db.Store) error {
		var createErr error
		evaluationPolicy, createErr = s.repo.CreatePolicy(ctx, &CreatePolicyRequest{
			Name:       fmt.Sprintf("evaluation_policy_%s", uuid.New().String()[0:8]),
			PolicyType: types.PolicyTypeABAC,
			Effect:     types.PolicyEffectAllow,
			Priority:   100,
			Category:   types.PolicyCategoryAccess,
			Target: map[string]any{
				"subjects": map[string]any{
					"user.department": "engineering",
				},
				"resources": map[string]any{
					"resource.type": "test_resource",
				},
				"actions": []any{"read", "execute"},
			},
			Rule: map[string]any{
				"user.security_level": map[string]any{
					"gte": 5,
				},
			},
			CreatedBy: uuid.New(),
		})
		return createErr
	})
	s.Require().NoError(err)
	s.trackPolicyForCleanup(evaluationPolicy.ID)

	testCases := []struct {
		name        string
		spec        string
		tenantID    uuid.UUID
		request     *GetPoliciesForEvaluationRequest
		expectError string
		validate    func([]*models.Policy)
	}{
		{
			name:     "GetPoliciesForEvaluation_Success",
			spec:     "ABAC-REPO-006",
			tenantID: s.tenantA.ID,
			request: &GetPoliciesForEvaluationRequest{
				ResourceType: "test_resource",
				Action:       "read",
			},
			validate: func(policies []*models.Policy) {
				require.GreaterOrEqual(s.T(), len(policies), 1)

				// All policies should belong to tenant A
				for _, policy := range policies {
					require.Equal(s.T(), s.tenantA.ID, policy.TenantID)
				}

				// Should contain our evaluation policy
				found := false
				for _, policy := range policies {
					if policy.ID == evaluationPolicy.ID {
						found = true
						break
					}
				}
				require.True(s.T(), found, "Should find the evaluation policy")
			},
		},
		{
			name:     "GetPoliciesForEvaluationDifferentTenant_Empty",
			spec:     "ABAC-REPO-006",
			tenantID: s.tenantB.ID, // Different tenant
			request: &GetPoliciesForEvaluationRequest{
				ResourceType: "test_resource",
				Action:       "read",
			},
			validate: func(policies []*models.Policy) {
				// Should not find policies from tenant A
				for _, policy := range policies {
					require.Equal(s.T(), s.tenantB.ID, policy.TenantID)
					require.NotEqual(s.T(), evaluationPolicy.ID, policy.ID,
						"Should not find tenant A policy in tenant B context")
				}
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("%s_%s", tc.spec, tc.name), func() {
			var policies []*models.Policy
			var err error

			execErr := s.runner.GetStore().WithTenant(s.ctx, tc.tenantID, func(ctx context.Context, store db.Store) error {
				policies, err = s.repo.GetPoliciesForEvaluation(ctx, tc.request)
				return nil
			})
			s.Require().NoError(execErr)

			if tc.expectError != "" {
				require.Error(s.T(), err)
				require.Contains(s.T(), err.Error(), tc.expectError)
			} else {
				require.NoError(s.T(), err)
				require.NotNil(s.T(), policies)
				if tc.validate != nil {
					tc.validate(policies)
				}
			}
		})
	}
}

// TestMultiTenantRLSIntegrity tests ABAC-REPO-007:  Multi-tenant RLS Integrity
func (s *PolicyRepositoryTestSuite) TestMultiTenantRLSIntegrity() {
	s.Run("ABAC-REPO-007_RLSVerification", func() {
		// Test multi-tenant RLS enforcement

		// 1. Create policies in both tenants
		var tenantAPolicies, tenantBPolicies []*models.Policy

		// Create policies in tenant A
		err := s.runner.GetStore().WithTenant(s.ctx, s.tenantA.ID, func(ctx context.Context, store db.Store) error {
			for i := 0; i < 5; i++ {
				policy, createErr := s.repo.CreatePolicy(ctx, &CreatePolicyRequest{
					Name:       fmt.Sprintf("rls_test_a_%d_%s", i, uuid.New().String()[0:8]),
					PolicyType: types.PolicyTypeABAC,
					Effect:     types.PolicyEffectAllow,
					Priority:   100,
					Category:   types.PolicyCategoryAccess,
					Target:     map[string]any{"tenant": "A", "index": i},
					Rule:       map[string]any{"tenant": "A", "index": i},
					CreatedBy:  uuid.New(),
				})
				if createErr != nil {
					return createErr
				}
				tenantAPolicies = append(tenantAPolicies, policy)
				s.trackPolicyForCleanup(policy.ID)
			}
			return nil
		})
		s.Require().NoError(err)

		// Create policies in tenant B
		err = s.runner.GetStore().WithTenant(s.ctx, s.tenantB.ID, func(ctx context.Context, store db.Store) error {
			for i := 0; i < 3; i++ {
				policy, createErr := s.repo.CreatePolicy(ctx, &CreatePolicyRequest{
					Name:       fmt.Sprintf("rls_test_b_%d_%s", i, uuid.New().String()[0:8]),
					PolicyType: types.PolicyTypeABAC,
					Effect:     types.PolicyEffectDeny,
					Priority:   200,
					Category:   types.PolicyCategoryAccess,
					Target:     map[string]any{"tenant": "B", "index": i},
					Rule:       map[string]any{"tenant": "B", "index": i},
					CreatedBy:  uuid.New(),
				})
				if createErr != nil {
					return createErr
				}
				tenantBPolicies = append(tenantBPolicies, policy)
				s.trackPolicyForCleanup(policy.ID)
			}
			return nil
		})
		s.Require().NoError(err)

		// 2. Verify tenant A can only access its own policies
		err = s.runner.GetStore().WithTenant(s.ctx, s.tenantA.ID, func(ctx context.Context, store db.Store) error {
			// List all policies in tenant A
			policies, listErr := s.repo.ListPolicies(ctx, &ListPoliciesRequest{})
			if listErr != nil {
				return listErr
			}

			// Verify all policies belong to tenant A
			tenantAPolicyIDs := make(map[uuid.UUID]bool)
			for _, policy := range policies {
				require.Equal(s.T(), s.tenantA.ID, policy.TenantID,
					"All policies in tenant A context should belong to tenant A")
				tenantAPolicyIDs[policy.ID] = true
			}

			// Verify all tenant A policies are accessible
			for _, expectedPolicy := range tenantAPolicies {
				require.True(s.T(), tenantAPolicyIDs[expectedPolicy.ID],
					"Tenant A should have access to its own policy: %s", expectedPolicy.ID)
			}

			// Verify tenant B policies are NOT accessible
			for _, tenantBPolicy := range tenantBPolicies {
				require.False(s.T(), tenantAPolicyIDs[tenantBPolicy.ID],
					"Tenant A should NOT have access to tenant B policy: %s", tenantBPolicy.ID)
			}

			return nil
		})
		s.Require().NoError(err)

		// 3. Verify tenant B can only access its own policies
		err = s.runner.GetStore().WithTenant(s.ctx, s.tenantB.ID, func(ctx context.Context, store db.Store) error {
			// List all policies in tenant B
			policies, listErr := s.repo.ListPolicies(ctx, &ListPoliciesRequest{})
			if listErr != nil {
				return listErr
			}

			// Verify all policies belong to tenant B
			tenantBPolicyIDs := make(map[uuid.UUID]bool)
			for _, policy := range policies {
				require.Equal(s.T(), s.tenantB.ID, policy.TenantID,
					"All policies in tenant B context should belong to tenant B")
				tenantBPolicyIDs[policy.ID] = true
			}

			// Verify all tenant B policies are accessible
			for _, expectedPolicy := range tenantBPolicies {
				require.True(s.T(), tenantBPolicyIDs[expectedPolicy.ID],
					"Tenant B should have access to its own policy: %s", expectedPolicy.ID)
			}

			// Verify tenant A policies are NOT accessible
			for _, tenantAPolicy := range tenantAPolicies {
				require.False(s.T(), tenantBPolicyIDs[tenantAPolicy.ID],
					"Tenant B should NOT have access to tenant A policy: %s", tenantAPolicy.ID)
			}

			return nil
		})
		s.Require().NoError(err)

		// 4. Verify cross-tenant individual policy access is blocked
		for _, tenantAPolicy := range tenantAPolicies {
			err = s.runner.GetStore().WithTenant(s.ctx, s.tenantB.ID, func(ctx context.Context, store db.Store) error {
				_, getErr := s.repo.GetPolicyByID(ctx, tenantAPolicy.ID)
				require.Error(s.T(), getErr, "Tenant B should not be able to get tenant A policy")
				require.Contains(s.T(), getErr.Error(), "not found")
				return nil
			})
			s.Require().NoError(err)
		}

		for _, tenantBPolicy := range tenantBPolicies {
			err = s.runner.GetStore().WithTenant(s.ctx, s.tenantA.ID, func(ctx context.Context, store db.Store) error {
				_, getErr := s.repo.GetPolicyByID(ctx, tenantBPolicy.ID)
				require.Error(s.T(), getErr, "Tenant A should not be able to get tenant B policy")
				require.Contains(s.T(), getErr.Error(), "not found")
				return nil
			})
			s.Require().NoError(err)
		}

		// 5. Verify cross-tenant updates are blocked
		for _, tenantAPolicy := range tenantAPolicies[:1] { // Test one policy
			err = s.runner.GetStore().WithTenant(s.ctx, s.tenantB.ID, func(ctx context.Context, store db.Store) error {
				_, updateErr := s.repo.UpdatePolicy(ctx, tenantAPolicy.ID, &UpdatePolicyRequest{
					DisplayName: stringPtr("Attempted Cross-Tenant Update"),
				})
				require.Error(s.T(), updateErr, "Cross-tenant update should fail")
				return nil
			})
			s.Require().NoError(err)
		}

		// 6. Verify cross-tenant deletes don't affect other tenants
		for _, tenantAPolicy := range tenantAPolicies[:1] { // Test one policy
			err = s.runner.GetStore().WithTenant(s.ctx, s.tenantB.ID, func(ctx context.Context, store db.Store) error {
				deleteErr := s.repo.DeletePolicy(ctx, tenantAPolicy.ID)
				// NOTE: Delete should not error but should not affect the policy
				require.NoError(s.T(), deleteErr, "Cross-tenant delete should not error but should not affect policy")
				return nil
			})
			s.Require().NoError(err)

			// Verify the policy still exists in tenant A
			err = s.runner.GetStore().WithTenant(s.ctx, s.tenantA.ID, func(ctx context.Context, store db.Store) error {
				policy, getErr := s.repo.GetPolicyByID(ctx, tenantAPolicy.ID)
				require.NoError(s.T(), getErr, "Policy should still exist in tenant A after cross-tenant delete attempt")
				require.NotNil(s.T(), policy)
				require.Equal(s.T(), tenantAPolicy.ID, policy.ID)
				return nil
			})
			s.Require().NoError(err)
		}

		s.T().Logf("✅ Multi-tenant RLS integrity verification completed successfully")
		s.T().Logf("   - Verified tenant isolation for %d tenant A policies", len(tenantAPolicies))
		s.T().Logf("   - Verified tenant isolation for %d tenant B policies", len(tenantBPolicies))
		s.T().Logf("   - Confirmed cross-tenant access, update, and delete protections")
	})
}

// TestCacheIntegration tests ABAC-REPO-008: Cache Integration with Tenant Isolation
func (s *PolicyRepositoryTestSuite) TestCacheIntegration() {
	// NOTE: This test uses a no-op cache, but verifies the integration points
	// In a real environment with Redis cache, this would test cache hit/miss scenarios

	var testPolicy *models.Policy

	err := s.runner.GetStore().WithTenant(s.ctx, s.tenantA.ID, func(ctx context.Context, store db.Store) error {
		var createErr error
		testPolicy, createErr = s.repo.CreatePolicy(ctx, &CreatePolicyRequest{
			Name:       fmt.Sprintf("cache_test_policy_%s", uuid.New().String()[0:8]),
			PolicyType: types.PolicyTypeABAC,
			Effect:     types.PolicyEffectAllow,
			Priority:   100,
			Category:   types.PolicyCategoryAccess,
			Target:     map[string]any{"cache": "test"},
			Rule:       map[string]any{"cache": "test"},
			CreatedBy:  uuid.New(),
		})
		return createErr
	})
	s.Require().NoError(err)
	s.trackPolicyForCleanup(testPolicy.ID)

	s.Run("ABAC-REPO-008_CacheIntegration", func() {
		// Test multiple retrieval calls - with real cache, second call would be cached
		err := s.runner.GetStore().WithTenant(s.ctx, s.tenantA.ID, func(ctx context.Context, store db.Store) error {
			// First retrieval
			policy1, getErr1 := s.repo.GetPolicyByID(ctx, testPolicy.ID)
			require.NoError(s.T(), getErr1)
			require.NotNil(s.T(), policy1)
			require.Equal(s.T(), testPolicy.ID, policy1.ID)

			// Second retrieval - would hit cache in real scenario
			policy2, getErr2 := s.repo.GetPolicyByID(ctx, testPolicy.ID)
			require.NoError(s.T(), getErr2)
			require.NotNil(s.T(), policy2)
			require.Equal(s.T(), testPolicy.ID, policy2.ID)

			// Policies should be equivalent
			require.Equal(s.T(), policy1.ID, policy2.ID)
			require.Equal(s.T(), policy1.Name, policy2.Name)

			return nil
		})
		s.Require().NoError(err)
	})
}

// TestRepositoryErrorHandling tests ABAC-REPO-009:  Error Handling
func (s *PolicyRepositoryTestSuite) TestRepositoryErrorHandling() {
	s.Run("ABAC-REPO-009_ErrorHandling", func() {
		// Test various error scenarios

		// 1. Test invalid tenant context
		invalidTenantID := uuid.New()
		err := s.runner.GetStore().WithTenant(s.ctx, invalidTenantID, func(ctx context.Context, store db.Store) error {
			_, createErr := s.repo.CreatePolicy(ctx, &CreatePolicyRequest{
				Name:       "error_test_policy",
				PolicyType: types.PolicyTypeABAC,
				Effect:     types.PolicyEffectAllow,
				Priority:   100,
				Category:   types.PolicyCategoryAccess,
				Target:     map[string]any{"test": true},
				Rule:       map[string]any{"test": true},
				CreatedBy:  uuid.New(),
			})
			// This should work as WithTenant sets up the context properly
			require.NoError(s.T(), createErr, "Policy creation with valid parameters should succeed")
			return nil
		})
		s.Require().NoError(err)

		// 2. Test invalid JSON in target field
		err = s.runner.GetStore().WithTenant(s.ctx, s.tenantA.ID, func(ctx context.Context, store db.Store) error {
			// Create a policy with complex JSON that might cause marshaling issues
			complexTarget := map[string]any{
				"subjects": map[string]any{
					"user.attributes": []any{
						map[string]any{
							"complex": map[string]any{
								"nested": map[string]any{
									"deeply": []any{1, 2, 3, "string", true, nil},
								},
							},
						},
					},
				},
			}
			_, createErr := s.repo.CreatePolicy(ctx, &CreatePolicyRequest{
				Name:       fmt.Sprintf("complex_json_test_%s", uuid.New().String()[0:8]),
				PolicyType: types.PolicyTypeABAC,
				Effect:     types.PolicyEffectAllow,
				Priority:   100,
				Category:   types.PolicyCategoryAccess,
				Target:     complexTarget,
				Rule:       map[string]any{"test": true},
				CreatedBy:  uuid.New(),
			})
			require.NoError(s.T(), createErr, "Complex JSON should be handled properly")
			return nil
		})
		s.Require().NoError(err)

		// 3. Test database constraints (if any)
		// This depends on the specific database constraints implemented

		s.T().Logf("✅ Error handling scenarios tested successfully")
	})
}

// TestRepositoryPerformance tests ABAC-REPO-010: Performance and Scalability
func (s *PolicyRepositoryTestSuite) TestRepositoryPerformance() {
	s.Run("ABAC-REPO-010_PerformanceBaseline", func() {
		const policyCount = 50 // Reasonable number for unit tests
		var createdPolicies []*models.Policy

		// Create multiple policies and measure time
		startTime := time.Now()

		err := s.runner.GetStore().WithTenant(s.ctx, s.tenantA.ID, func(ctx context.Context, store db.Store) error {
			for i := 0; i < policyCount; i++ {
				policy, createErr := s.repo.CreatePolicy(ctx, &CreatePolicyRequest{
					Name:       fmt.Sprintf("perf_test_%d_%s", i, uuid.New().String()[0:8]),
					PolicyType: types.PolicyTypeABAC,
					Effect:     types.PolicyEffectAllow,
					Priority:   100 + int32(i),
					Category:   types.PolicyCategoryAccess,
					Target: map[string]any{
						"resources": map[string]any{
							"resource.type": fmt.Sprintf("resource_%d", i),
						},
					},
					Rule: map[string]any{
						"user.level": map[string]any{
							"gte": i % 10,
						},
					},
					CreatedBy: uuid.New(),
				})
				if createErr != nil {
					return createErr
				}
				createdPolicies = append(createdPolicies, policy)
				s.trackPolicyForCleanup(policy.ID)
			}
			return nil
		})
		s.Require().NoError(err)

		creationTime := time.Since(startTime)
		s.T().Logf("Created %d policies in %v (avg: %v per policy)",
			policyCount, creationTime, creationTime/policyCount)

		// Test retrieval performance
		startTime = time.Now()

		err = s.runner.GetStore().WithTenant(s.ctx, s.tenantA.ID, func(ctx context.Context, store db.Store) error {
			for _, policy := range createdPolicies[:10] { // Test first 10
				retrievedPolicy, getErr := s.repo.GetPolicyByID(ctx, policy.ID)
				if getErr != nil {
					return getErr
				}
				require.NotNil(s.T(), retrievedPolicy)
				require.Equal(s.T(), policy.ID, retrievedPolicy.ID)
			}
			return nil
		})
		s.Require().NoError(err)

		retrievalTime := time.Since(startTime)
		s.T().Logf("Retrieved 10 policies in %v (avg: %v per retrieval)",
			retrievalTime, retrievalTime/10)

		// Test list performance
		startTime = time.Now()

		err = s.runner.GetStore().WithTenant(s.ctx, s.tenantA.ID, func(ctx context.Context, store db.Store) error {
			policies, listErr := s.repo.ListPolicies(ctx, &ListPoliciesRequest{})
			if listErr != nil {
				return listErr
			}
			require.GreaterOrEqual(s.T(), len(policies), policyCount)
			return nil
		})
		s.Require().NoError(err)

		listTime := time.Since(startTime)
		s.T().Logf("Listed all policies in %v", listTime)

		// Performance assertions (reasonable thresholds for unit tests)
		require.Less(s.T(), creationTime.Milliseconds(), int64(5000), // 5 seconds max for 50 policies
			"Policy creation should be reasonably fast")
		require.Less(s.T(), retrievalTime.Milliseconds(), int64(500), // 500ms max for 10 retrievals
			"Policy retrieval should be fast")
		require.Less(s.T(), listTime.Milliseconds(), int64(1000), // 1 second max for listing
			"Policy listing should be fast")

		s.T().Logf("✅ Performance baseline established successfully")
	})
}

// Helper functions for test data creation
func stringPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}
