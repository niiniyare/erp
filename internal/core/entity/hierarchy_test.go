//go:build database
// +build database

package entity

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// OrganizationHierarchyTestSuite tests organization hierarchy functionality with tenant isolation
type OrganizationHierarchyTestSuite struct {
	suite.Suite
	ctrl     *gomock.Controller
	dbRunner *tenant.DatabaseTestRunner
	service  Service
	repo     Repository
	tracing  *tracing.MockTracingService
	metrics  *metrics.MockMetricsProvider
	ctx      context.Context

	// Test data cleanup
	createdEntities []uuid.UUID
	testTenantIDs   []uuid.UUID
}

// SetupSuite runs before the test suite
func (s *OrganizationHierarchyTestSuite) SetupSuite() {
	s.ctx = context.Background()

	// Setup database test runner
	dbRunner, err := tenant.NewDatabaseTestRunner()
	if err != nil {
		s.T().Skipf("Database not available for testing: %v", err)
		return
	}
	s.dbRunner = dbRunner

	// Initialize controller and mocks
	s.ctrl = gomock.NewController(s.T())
	s.tracing = tracing.NewMockTracingService(s.ctrl)
	s.metrics = metrics.NewMockMetricsProvider(s.ctrl)

	// Setup tracing expectations
	mockSpan := tracing.NewMockSpan(s.ctrl)
	mockSpan.EXPECT().End(gomock.Any()).AnyTimes()
	mockSpan.EXPECT().SetAttributes(gomock.Any()).AnyTimes()
	mockSpan.EXPECT().SetStatus(gomock.Any(), gomock.Any()).AnyTimes()
	mockSpan.EXPECT().RecordError(gomock.Any()).AnyTimes()
	s.tracing.EXPECT().StartSpan(gomock.Any(), gomock.Any(), gomock.Any()).Return(context.Background(), mockSpan).AnyTimes()

	// Setup metrics expectations
	mockTimer := metrics.NewMockTimer(s.ctrl)
	mockTimer.EXPECT().Stop().Return(time.Duration(100) * time.Millisecond).AnyTimes()
	s.metrics.EXPECT().Timer(gomock.Any(), gomock.Any()).Return(mockTimer).AnyTimes()
	s.metrics.EXPECT().IncrementCounter(gomock.Any(), gomock.Any()).AnyTimes()
	s.metrics.EXPECT().ObserveHistogram(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	// Create repository and service
	s.repo = NewRepository(s.dbRunner.GetStore(), s.tracing)
	s.service = NewService(s.repo, s.tracing, s.metrics)

	// Initialize cleanup tracking
	s.createdEntities = make([]uuid.UUID, 0)
	s.testTenantIDs = make([]uuid.UUID, 0)
}

// TearDownSuite runs after the test suite
func (s *OrganizationHierarchyTestSuite) TearDownSuite() {
	if s.dbRunner != nil {
		// Clean up created entities
		for _, entityID := range s.createdEntities {
			_ = s.service.DeleteEntity(s.ctx, entityID, true)
		}

		// Clean up test tenants
		for _, tenantID := range s.testTenantIDs {
			_ = s.dbRunner.GetStore().SoftDeleteTenant(s.ctx, tenantID)
		}

		s.dbRunner.Close()
	}
	if s.ctrl != nil {
		s.ctrl.Finish()
	}
}

// createTestTenant creates a test tenant and sets the context
func (s *OrganizationHierarchyTestSuite) createTestTenant(name string) uuid.UUID {
	// Make tenant name unique by adding UUID suffix
	uniqueName := name + "-" + uuid.New().String()[:8]
	tenant, err := s.dbRunner.CreateTestTenant(s.ctx, uniqueName)
	s.Require().NoError(err)
	s.testTenantIDs = append(s.testTenantIDs, tenant.ID)

	// Set tenant context
	err = s.dbRunner.GetStore().SetTenantContext(s.ctx, tenant.ID)
	s.Require().NoError(err)

	return tenant.ID
}

// createTestEntity creates a test entity and tracks it for cleanup
func (s *OrganizationHierarchyTestSuite) createTestEntity(name, code string, entityType EntityType, parentID *uuid.UUID) *Entity {
	req := CreateEntityRequest{
		Name:          name,
		Code:          code,
		Type:          entityType,
		ParentID:      parentID,
		IsActive:      true,
		IsHidden:      false,
		AccrualMethod: true, // Default to accrual accounting
		FYStartMonth:  1,    // Default to January fiscal year start
		Address:       map[string]any{"country": "US", "city": "Test City"},
		Picture:       "",
		Settings:      map[string]any{"currency": "USD", "timezone": "UTC"},
		Metadata:      map[string]any{"test": true},
	}

	entity, err := s.service.CreateEntity(s.ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(entity)

	s.createdEntities = append(s.createdEntities, entity.ID)
	return entity
}

// TestOrganizationCreation tests MT-ORG-001: Organization Creation within tenant
func (s *OrganizationHierarchyTestSuite) TestOrganizationCreation() {
	s.Run("MT-ORG-001: Organization creation within tenant context", func() {
		// Create test tenant
		tenantID := s.createTestTenant("org-creation-test")

		// Test creating different types of organizations
		testCases := []struct {
			name       string
			orgName    string
			orgCode    string
			orgType    EntityType
			shouldPass bool
		}{
			{
				name:       "Create Company",
				orgName:    "ACME Corporation",
				orgCode:    "ACME",
				orgType:    EntityTypeCompany,
				shouldPass: true,
			},
			{
				name:       "Create Department",
				orgName:    "Engineering Department",
				orgCode:    "ENG",
				orgType:    EntityTypeDepartment,
				shouldPass: true,
			},
			{
				name:       "Create Location",
				orgName:    "New York Office",
				orgCode:    "NYC",
				orgType:    EntityTypeLocation,
				shouldPass: true,
			},
			{
				name:       "Create Project",
				orgName:    "ERP Implementation",
				orgCode:    "ERP-IMPL",
				orgType:    EntityTypeProject,
				shouldPass: true,
			},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				// Create the organization
				entity := s.createTestEntity(tc.orgName, tc.orgCode, tc.orgType, nil)

				// Verify the organization was created correctly
				s.Equal(tc.orgName, entity.Name)
				s.Equal(tc.orgCode, entity.Code)
				s.Equal(tc.orgType, entity.Type)
				s.True(entity.IsActive)
				s.False(entity.IsHidden)
				s.NotEqual(uuid.Nil, entity.ID)
				s.WithinDuration(time.Now(), entity.CreatedAt, 5*time.Second)

				// Verify the organization can be retrieved
				retrievedEntity, err := s.service.GetEntityByID(s.ctx, entity.ID)
				s.Require().NoError(err)
				s.Equal(entity.ID, retrievedEntity.ID)
				s.Equal(entity.Name, retrievedEntity.Name)

				// Verify code uniqueness within tenant
				duplicateReq := CreateEntityRequest{
					Name:          tc.orgName + " Duplicate",
					Code:          tc.orgCode, // Same code
					Type:          tc.orgType,
					IsActive:      true,
					AccrualMethod: true,
					FYStartMonth:  1,
				}

				_, err = s.service.CreateEntity(s.ctx, duplicateReq)
				s.Error(err, "Should not allow duplicate codes within tenant")
			})
		}

		s.T().Logf("Successfully tested organization creation for tenant %s", tenantID)
	})
}

// TestHierarchicalRelationships tests MT-ORG-002: Parent-child organization relationships
func (s *OrganizationHierarchyTestSuite) TestHierarchicalRelationships() {
	s.Run("MT-ORG-002: Hierarchical organization relationships", func() {
		// Create test tenant
		tenantID := s.createTestTenant("hierarchy-test")

		// Create organization hierarchy: Company → Division → Department → Team
		company := s.createTestEntity("ACME Corporation", "ACME", EntityTypeCompany, nil)
		division := s.createTestEntity("Technology Division", "TECH-DIV", EntityTypeDivision, &company.ID)
		department := s.createTestEntity("Engineering Department", "ENG-DEPT", EntityTypeDepartment, &division.ID)
		team := s.createTestEntity("Backend Team", "BE-TEAM", EntityTypeProject, &department.ID)

		s.T().Logf("Created hierarchy for tenant %s: Company(%s) → Division(%s) → Department(%s) → Team(%s)",
			tenantID, company.ID, division.ID, department.ID, team.ID)

		// Test parent-child relationships
		s.Run("Verify parent-child relationships", func() {
			// Test getting children
			companyChildren, err := s.service.GetEntityChildren(s.ctx, company.ID)
			s.Require().NoError(err)
			s.Len(companyChildren, 1)
			s.Equal(division.ID, companyChildren[0].ID)

			divisionChildren, err := s.service.GetEntityChildren(s.ctx, division.ID)
			s.Require().NoError(err)
			s.Len(divisionChildren, 1)
			s.Equal(department.ID, divisionChildren[0].ID)

			departmentChildren, err := s.service.GetEntityChildren(s.ctx, department.ID)
			s.Require().NoError(err)
			s.Len(departmentChildren, 1)
			s.Equal(team.ID, departmentChildren[0].ID)

			// Leaf node should have no children
			teamChildren, err := s.service.GetEntityChildren(s.ctx, team.ID)
			s.Require().NoError(err)
			s.Len(teamChildren, 0)
		})

		s.Run("Verify ancestor relationships", func() {
			// Test getting ancestors
			teamAncestors, err := s.service.GetEntityAncestors(s.ctx, team.ID)
			s.Require().NoError(err)
			s.Len(teamAncestors, 3) // Should have department, division, company

			// Verify ancestor order (should be from immediate parent to root)
			ancestorIDs := make([]uuid.UUID, len(teamAncestors))
			for i, ancestor := range teamAncestors {
				ancestorIDs[i] = ancestor.ID
			}

			s.Contains(ancestorIDs, department.ID)
			s.Contains(ancestorIDs, division.ID)
			s.Contains(ancestorIDs, company.ID)
		})

		s.Run("Prevent circular references", func() {
			// Try to make company a child of team (would create circular reference)
			updateReq := UpdateEntityRequest{
				ParentID: &team.ID,
			}

			_, err := s.service.UpdateEntity(s.ctx, company.ID, updateReq)
			s.Error(err, "Should prevent circular reference")
		})

		s.Run("Prevent deletion of entities with children", func() {
			// Try to delete company (has children)
			err := s.service.DeleteEntity(s.ctx, company.ID, true)
			s.Error(err, "Should prevent deletion of entity with children")

			// Should be able to delete leaf node
			err = s.service.DeleteEntity(s.ctx, team.ID, false) // Soft delete
			s.NoError(err, "Should allow deletion of entity without children")

			// Restore for cleanup
			err = s.service.RestoreEntity(s.ctx, team.ID)
			s.NoError(err)
		})

		s.Run("Test hierarchy depth limits", func() {
			// Create additional levels to test depth
			subTeam := s.createTestEntity("Core Backend Team", "CORE-BE", EntityTypeProject, &team.ID)
			workGroup := s.createTestEntity("API Work Group", "API-WG", EntityTypeProject, &subTeam.ID)

			// Verify we can retrieve the deep hierarchy
			workGroupAncestors, err := s.service.GetEntityAncestors(s.ctx, workGroup.ID)
			s.Require().NoError(err)
			s.GreaterOrEqual(len(workGroupAncestors), 5) // Should have at least 5 levels

			s.T().Logf("Successfully tested hierarchy depth with %d levels", len(workGroupAncestors)+1)
		})

		s.Run("Test entity tree retrieval", func() {
			// Get the complete entity tree
			entityTree, err := s.service.GetEntityTree(s.ctx)
			s.Require().NoError(err)
			s.NotEmpty(entityTree)

			// Verify tree structure contains our created entities
			foundCompany := false
			for _, treeEntity := range entityTree {
				if treeEntity.ID == company.ID {
					foundCompany = true
					s.GreaterOrEqual(treeEntity.Level, 0)
					s.NotEmpty(treeEntity.Path)
					break
				}
			}
			s.True(foundCompany, "Company should be found in entity tree")

			s.T().Logf("Successfully retrieved entity tree with %d entities", len(entityTree))
		})
	})
}

// TestCrossTenantOrganizationIsolation tests MT-ORG-003: Cross-tenant organization isolation
func (s *OrganizationHierarchyTestSuite) TestCrossTenantOrganizationIsolation() {
	s.Run("MT-ORG-003: Cross-tenant organization isolation", func() {
		// Create two separate tenants
		tenant1ID := s.createTestTenant("isolation-tenant-1")

		// Create organizations in tenant 1
		company1 := s.createTestEntity("ACME Corp Tenant 1", "ACME-T1", EntityTypeCompany, nil)
		dept1 := s.createTestEntity("Engineering Dept T1", "ENG-T1", EntityTypeDepartment, &company1.ID)

		// Switch to tenant 2
		tenant2ID := s.createTestTenant("isolation-tenant-2")

		// Create organizations in tenant 2 (can reuse codes due to tenant isolation)
		company2 := s.createTestEntity("ACME Corp Tenant 2", "ACME-T1", EntityTypeCompany, nil)          // Same code as tenant 1
		dept2 := s.createTestEntity("Engineering Dept T2", "ENG-T1", EntityTypeDepartment, &company2.ID) // Same code as tenant 1

		s.T().Logf("Created isolated organizations for tenant1=%s and tenant2=%s", tenant1ID, tenant2ID)

		s.Run("Verify tenant data isolation", func() {
			// Switch back to tenant 1
			err := s.dbRunner.GetStore().SetTenantContext(s.ctx, tenant1ID)
			s.Require().NoError(err)

			// Should only see tenant 1's organizations
			entities, err := s.service.ListEntities(s.ctx, ListEntitiesRequest{Limit: 100})
			s.Require().NoError(err)

			tenant1EntityIDs := make(map[uuid.UUID]bool)
			for _, entity := range entities {
				tenant1EntityIDs[entity.ID] = true
			}

			// Should contain tenant 1's entities
			s.True(tenant1EntityIDs[company1.ID], "Should contain tenant 1's company")
			s.True(tenant1EntityIDs[dept1.ID], "Should contain tenant 1's department")

			// Should NOT contain tenant 2's entities
			s.False(tenant1EntityIDs[company2.ID], "Should NOT contain tenant 2's company")
			s.False(tenant1EntityIDs[dept2.ID], "Should NOT contain tenant 2's department")

			// Switch to tenant 2
			err = s.dbRunner.GetStore().SetTenantContext(s.ctx, tenant2ID)
			s.Require().NoError(err)

			// Should only see tenant 2's organizations
			entities, err = s.service.ListEntities(s.ctx, ListEntitiesRequest{Limit: 100})
			s.Require().NoError(err)

			tenant2EntityIDs := make(map[uuid.UUID]bool)
			for _, entity := range entities {
				tenant2EntityIDs[entity.ID] = true
			}

			// Should contain tenant 2's entities
			s.True(tenant2EntityIDs[company2.ID], "Should contain tenant 2's company")
			s.True(tenant2EntityIDs[dept2.ID], "Should contain tenant 2's department")

			// Should NOT contain tenant 1's entities
			s.False(tenant2EntityIDs[company1.ID], "Should NOT contain tenant 1's company")
			s.False(tenant2EntityIDs[dept1.ID], "Should NOT contain tenant 1's department")
		})

		s.Run("Prevent cross-tenant parent references", func() {
			// Switch to tenant 2
			err := s.dbRunner.GetStore().SetTenantContext(s.ctx, tenant2ID)
			s.Require().NoError(err)

			// Try to create an entity in tenant 2 with parent from tenant 1
			req := CreateEntityRequest{
				Name:          "Cross Tenant Child",
				Code:          "CROSS-TENANT",
				Type:          EntityTypeProject,
				ParentID:      &company1.ID, // Parent from tenant 1
				IsActive:      true,
				AccrualMethod: true,
				FYStartMonth:  1,
			}

			_, err = s.service.CreateEntity(s.ctx, req)
			s.Error(err, "Should prevent cross-tenant parent references")
		})

		s.Run("Prevent cross-tenant entity access", func() {
			// Switch to tenant 1
			err := s.dbRunner.GetStore().SetTenantContext(s.ctx, tenant1ID)
			s.Require().NoError(err)

			// Try to access entity from tenant 2
			_, err = s.service.GetEntityByID(s.ctx, company2.ID)
			s.Error(err, "Should prevent access to entities from other tenants")

			// Try to update entity from tenant 2
			updateReq := UpdateEntityRequest{
				Name: stringPtr("Hacked Name"),
			}
			_, err = s.service.UpdateEntity(s.ctx, company2.ID, updateReq)
			s.Error(err, "Should prevent updating entities from other tenants")

			// Try to delete entity from tenant 2
			err = s.service.DeleteEntity(s.ctx, company2.ID, true)
			s.Error(err, "Should prevent deleting entities from other tenants")
		})

		s.Run("Verify hierarchy operations respect tenant boundaries", func() {
			// Switch to tenant 1
			err := s.dbRunner.GetStore().SetTenantContext(s.ctx, tenant1ID)
			s.Require().NoError(err)

			// Get children of tenant 1's company
			children, err := s.service.GetEntityChildren(s.ctx, company1.ID)
			s.Require().NoError(err)
			s.Len(children, 1)
			s.Equal(dept1.ID, children[0].ID)

			// Try to get children of tenant 2's company (should fail or return empty)
			_, err = s.service.GetEntityChildren(s.ctx, company2.ID)
			s.Error(err, "Should not be able to get children of entities from other tenants")

			// Get entity tree should only return tenant 1's entities
			entityTree, err := s.service.GetEntityTree(s.ctx)
			s.Require().NoError(err)

			treeEntityIDs := make(map[uuid.UUID]bool)
			for _, treeEntity := range entityTree {
				treeEntityIDs[treeEntity.ID] = true
			}

			s.True(treeEntityIDs[company1.ID], "Tree should contain tenant 1's entities")
			s.False(treeEntityIDs[company2.ID], "Tree should NOT contain tenant 2's entities")
		})

		s.Run("Test code uniqueness within tenant boundaries", func() {
			// Verify that the same codes can exist in different tenants
			// This was already tested in the setup, but let's be explicit

			// Switch to tenant 1
			err := s.dbRunner.GetStore().SetTenantContext(s.ctx, tenant1ID)
			s.Require().NoError(err)

			entity1, err := s.service.GetEntity(s.ctx, "ACME-T1")
			s.Require().NoError(err)
			s.Equal(company1.ID, entity1.ID)

			// Switch to tenant 2
			err = s.dbRunner.GetStore().SetTenantContext(s.ctx, tenant2ID)
			s.Require().NoError(err)

			entity2, err := s.service.GetEntity(s.ctx, "ACME-T1")
			s.Require().NoError(err)
			s.Equal(company2.ID, entity2.ID)

			// Should be different entities despite same code
			s.NotEqual(entity1.ID, entity2.ID)
		})
	})
}

// TestOrganizationHierarchy runs the organization hierarchy test suite
func TestOrganizationHierarchy(t *testing.T) {
	suite.Run(t, new(OrganizationHierarchyTestSuite))
}

// Helper function for string pointers
func stringPtr(s string) *string {
	return &s
}
