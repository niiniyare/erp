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

	db "awo/db/sqlc"
	"awo/internal/core/abac/models"
	"awo/internal/core/tenant"
	"awo/internal/platform/cache"
	"awo/internal/shared/logger"
	"awo/internal/shared/metrics"
	"awo/internal/shared/tracing"
	"awo/internal/shared/types"
)

// AttributeRepositoryTestSuite defines test suite for ABAC attribute repository operations
// Tests cover: ABAC-ATTR-REPO-001 through ABAC-ATTR-REPO-008 with multi-tenant RLS verification
type AttributeRepositoryTestSuite struct {
	suite.Suite
	ctx          context.Context
	runner       *tenant.DatabaseTestRunner
	repo         AttributeRepository
	cacheService cache.Service
	tenantA      *db.Tenant
	tenantB      *db.Tenant

	// Test data UUIDs for cleanup
	createdAttributeDefinitionIDs []uuid.UUID
	createdAttributeValueIDs      []uuid.UUID
}

// SetupSuite runs once before the entire test suite
func (s *AttributeRepositoryTestSuite) SetupSuite() {
	var err error
	s.runner, err = tenant.NewDatabaseTestRunner()
	s.Require().NoError(err, "Failed to connect to the database")
	s.ctx = context.Background()

	// Setup shared infrastructure
	logger := logger.WithFields(logger.Fields{"component": "attribute_repository_test"})
	metrics := &metrics.MetricsService{}
	tracer := tracing.NewNoOpTracingService()
	s.cacheService = cache.NewNoOpCache()

	// Create repository instance
	s.repo = NewAttributeRepository(
		s.runner.GetStore(),
		s.cacheService,
		logger,
		metrics,
		tracer,
	)

	// Create test tenants
	s.setupTestTenants()
	s.createdAttributeDefinitionIDs = make([]uuid.UUID, 0)
	s.createdAttributeValueIDs = make([]uuid.UUID, 0)
}

// TearDownSuite runs once after the entire test suite
func (s *AttributeRepositoryTestSuite) TearDownSuite() {
	// Clean up test data
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

// setupTestTenants creates two test tenants for multi-tenant testing
func (s *AttributeRepositoryTestSuite) setupTestTenants() {
	superuserStore := db.NewStore(s.runner.GetPool())
	uniqueID := uuid.New().String()[0:8]

	// Create Tenant A
	var err error
	s.tenantA, err = superuserStore.CreateTenant(s.ctx, db.CreateTenantParams{
		Name:   fmt.Sprintf("ABAC Attr Test Tenant A %s", uniqueID),
		Slug:   fmt.Sprintf("abac-attr-tenant-a-%s", uniqueID),
		Email:  fmt.Sprintf("attr-admin-a-%s@abactest.com", uniqueID),
		Status: "active",
	})
	s.Require().NoError(err, "Failed to create tenant A")

	// Create Tenant B
	uniqueID2 := uuid.New().String()[0:8]
	s.tenantB, err = superuserStore.CreateTenant(s.ctx, db.CreateTenantParams{
		Name:   fmt.Sprintf("ABAC Attr Test Tenant B %s", uniqueID2),
		Slug:   fmt.Sprintf("abac-attr-tenant-b-%s", uniqueID2),
		Email:  fmt.Sprintf("attr-admin-b-%s@abactest.com", uniqueID2),
		Status: "active",
	})
	s.Require().NoError(err, "Failed to create tenant B")
}

// cleanupTestData removes test attribute definitions and values created during testing
func (s *AttributeRepositoryTestSuite) cleanupTestData() {
	// NOTE: This is a simplified cleanup. In reality, we'd need proper cleanup methods
	// For now, tenant deletion will cascade delete the test data
}

// trackAttributeDefinitionForCleanup adds an attribute definition ID to the cleanup list
func (s *AttributeRepositoryTestSuite) trackAttributeDefinitionForCleanup(id uuid.UUID) {
	s.createdAttributeDefinitionIDs = append(s.createdAttributeDefinitionIDs, id)
}

// trackAttributeValueForCleanup adds an attribute value ID to the cleanup list
func (s *AttributeRepositoryTestSuite) trackAttributeValueForCleanup(id uuid.UUID) {
	s.createdAttributeValueIDs = append(s.createdAttributeValueIDs, id)
}

// TestAttributeRepository runs the attribute repository test suite
func TestAttributeRepository(t *testing.T) {
	suite.Run(t, new(AttributeRepositoryTestSuite))
}

// TestCreateAttributeDefinition tests ABAC-ATTR-REPO-001: Attribute Definition Creation with Multi-tenant Isolation
func (s *AttributeRepositoryTestSuite) TestCreateAttributeDefinition() {
	testCases := []struct {
		name        string
		spec        string
		tenantID    uuid.UUID
		request     *CreateAttributeDefinitionRequest
		expectError string
		validate    func(*models.AttributeDefinition)
	}{
		{
			name:     "CreateValidAttributeDefinition_Success",
			spec:     "ABAC-ATTR-REPO-001",
			tenantID: s.tenantA.ID,
			request: &CreateAttributeDefinitionRequest{
				Name:               fmt.Sprintf("user_department_%s", uuid.New().String()[0:8]),
				DisplayName:        stringPtr("User Department"),
				Description:        stringPtr("The department the user belongs to"),
				DataType:           types.AttributeDataTypeString,
				Category:           types.AttributeCategoryUser,
				IsRequired:         true,
				IsSensitive:        false,
				DefaultValue:       stringPtr("unassigned"),
				AllowedValues:      []string{"engineering", "marketing", "sales", "hr", "finance"},
				ValidationRules:    map[string]any{"max_length": 50},
				EncryptionRequired: false,
				IsActive:           true,
			},
			validate: func(def *models.AttributeDefinition) {
				require.NotNil(s.T(), def)
				require.NotEqual(s.T(), uuid.Nil, def.ID)
				require.Equal(s.T(), s.tenantA.ID, def.TenantID)
				require.Equal(s.T(), types.AttributeDataTypeString, def.DataType)
				require.Equal(s.T(), types.AttributeCategoryUser, def.Category)
				require.True(s.T(), def.IsRequired)
				require.False(s.T(), def.IsSensitive)
				require.True(s.T(), def.IsActive)
				require.NotZero(s.T(), def.CreatedAt)
				require.NotNil(s.T(), def.DefaultValue)
				require.Equal(s.T(), "unassigned", *def.DefaultValue)
				require.Len(s.T(), def.AllowedValues, 5)
			},
		},
		{
			name:     "CreateAttributeDefinitionEmptyName_ValidationError",
			spec:     "ABAC-ATTR-REPO-001",
			tenantID: s.tenantA.ID,
			request: &CreateAttributeDefinitionRequest{
				Name:     "", // Invalid: empty name
				DataType: types.AttributeDataTypeString,
				Category: types.AttributeCategoryUser,
				IsActive: true,
			},
			expectError: "validation",
		},
		{
			name:     "CreateSensitiveAttributeDefinition_Success",
			spec:     "ABAC-ATTR-REPO-001",
			tenantID: s.tenantA.ID,
			request: &CreateAttributeDefinitionRequest{
				Name:               fmt.Sprintf("user_ssn_%s", uuid.New().String()[0:8]),
				DisplayName:        stringPtr("Social Security Number"),
				Description:        stringPtr("User's sensitive SSN information"),
				DataType:           types.AttributeDataTypeString,
				Category:           types.AttributeCategoryUser,
				IsRequired:         false,
				IsSensitive:        true,
				ValidationRules:    map[string]any{"pattern": "^\\d{3}-\\d{2}-\\d{4}$"},
				EncryptionRequired: true,
				IsActive:           true,
			},
			validate: func(def *models.AttributeDefinition) {
				require.NotNil(s.T(), def)
				require.True(s.T(), def.IsSensitive)
				require.True(s.T(), def.EncryptionRequired)
				require.Equal(s.T(), types.AttributeCategoryUser, def.Category)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("%s_%s", tc.spec, tc.name), func() {
			var definition *models.AttributeDefinition
			var err error

			execErr := s.runner.GetStore().WithTenant(s.ctx, tc.tenantID, func(ctx context.Context, store db.Store) error {
				definition, err = s.repo.CreateAttributeDefinition(ctx, tc.request)
				return nil
			})
			s.Require().NoError(execErr, "Tenant context execution should not fail")

			if tc.expectError != "" {
				require.Error(s.T(), err)
				require.Contains(s.T(), err.Error(), tc.expectError)
				require.Nil(s.T(), definition)
			} else {
				require.NoError(s.T(), err)
				require.NotNil(s.T(), definition)
				if tc.validate != nil {
					tc.validate(definition)
				}
				s.trackAttributeDefinitionForCleanup(definition.ID)
			}
		})
	}
}

// TestGetAttributeDefinitionByID tests ABAC-ATTR-REPO-002: Attribute Definition Retrieval with Tenant Isolation
func (s *AttributeRepositoryTestSuite) TestGetAttributeDefinitionByID() {
	// Create test attribute definition in tenant A
	var defA *models.AttributeDefinition
	err := s.runner.GetStore().WithTenant(s.ctx, s.tenantA.ID, func(ctx context.Context, store db.Store) error {
		var createErr error
		defA, createErr = s.repo.CreateAttributeDefinition(ctx, &CreateAttributeDefinitionRequest{
			Name:        fmt.Sprintf("get_test_attr_%s", uuid.New().String()[0:8]),
			DisplayName: stringPtr("Get Test Attribute"),
			DataType:    types.AttributeDataTypeString,
			Category:    types.AttributeCategoryUser,
			IsRequired:  true,
			IsActive:    true,
		})
		return createErr
	})
	s.Require().NoError(err)
	s.trackAttributeDefinitionForCleanup(defA.ID)

	testCases := []struct {
		name         string
		spec         string
		tenantID     uuid.UUID
		definitionID uuid.UUID
		expectError  string
		validate     func(*models.AttributeDefinition)
	}{
		{
			name:         "GetExistingAttributeDefinition_Success",
			spec:         "ABAC-ATTR-REPO-002",
			tenantID:     s.tenantA.ID,
			definitionID: defA.ID,
			validate: func(def *models.AttributeDefinition) {
				require.NotNil(s.T(), def)
				require.Equal(s.T(), defA.ID, def.ID)
				require.Equal(s.T(), defA.Name, def.Name)
				require.Equal(s.T(), s.tenantA.ID, def.TenantID)
			},
		},
		{
			name:         "GetAttributeDefinitionFromDifferentTenant_NotFound",
			spec:         "ABAC-ATTR-REPO-002",
			tenantID:     s.tenantB.ID, // Different tenant
			definitionID: defA.ID,      // Definition from tenant A
			expectError:  "not found",
		},
		{
			name:         "GetNonExistentAttributeDefinition_NotFound",
			spec:         "ABAC-ATTR-REPO-002",
			tenantID:     s.tenantA.ID,
			definitionID: uuid.New(), // Random UUID
			expectError:  "not found",
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("%s_%s", tc.spec, tc.name), func() {
			var definition *models.AttributeDefinition
			var err error

			execErr := s.runner.GetStore().WithTenant(s.ctx, tc.tenantID, func(ctx context.Context, store db.Store) error {
				definition, err = s.repo.GetAttributeDefinitionByID(ctx, tc.definitionID)
				return nil
			})
			s.Require().NoError(execErr)

			if tc.expectError != "" {
				require.Error(s.T(), err)
				require.Contains(s.T(), err.Error(), tc.expectError)
				require.Nil(s.T(), definition)
			} else {
				require.NoError(s.T(), err)
				require.NotNil(s.T(), definition)
				if tc.validate != nil {
					tc.validate(definition)
				}
			}
		})
	}
}

// TestListAttributeDefinitions tests ABAC-ATTR-REPO-003: Attribute Definition Listing with Tenant Isolation
func (s *AttributeRepositoryTestSuite) TestListAttributeDefinitions() {
	// Create multiple attribute definitions in both tenants
	var defsA, defsB []*models.AttributeDefinition

	// Create definitions in tenant A
	err := s.runner.GetStore().WithTenant(s.ctx, s.tenantA.ID, func(ctx context.Context, store db.Store) error {
		categories := []types.AttributeCategory{
			types.AttributeCategoryUser,
			types.AttributeCategoryResource,
			types.AttributeCategoryEnvironment,
		}

		for i, category := range categories {
			def, createErr := s.repo.CreateAttributeDefinition(ctx, &CreateAttributeDefinitionRequest{
				Name:        fmt.Sprintf("list_test_a_%d_%s", i, uuid.New().String()[0:8]),
				DisplayName: stringPtr(fmt.Sprintf("List Test A %d", i)),
				DataType:    types.AttributeDataTypeString,
				Category:    category,
				IsRequired:  i%2 == 0, // Alternating required/optional
				IsActive:    true,
			})
			if createErr != nil {
				return createErr
			}
			defsA = append(defsA, def)
			s.trackAttributeDefinitionForCleanup(def.ID)
		}
		return nil
	})
	s.Require().NoError(err)

	// Create definitions in tenant B
	err = s.runner.GetStore().WithTenant(s.ctx, s.tenantB.ID, func(ctx context.Context, store db.Store) error {
		for i := 0; i < 2; i++ {
			def, createErr := s.repo.CreateAttributeDefinition(ctx, &CreateAttributeDefinitionRequest{
				Name:        fmt.Sprintf("list_test_b_%d_%s", i, uuid.New().String()[0:8]),
				DisplayName: stringPtr(fmt.Sprintf("List Test B %d", i)),
				DataType:    types.AttributeDataTypeNumber,
				Category:    types.AttributeCategoryUser,
				IsRequired:  true,
				IsActive:    true,
			})
			if createErr != nil {
				return createErr
			}
			defsB = append(defsB, def)
			s.trackAttributeDefinitionForCleanup(def.ID)
		}
		return nil
	})
	s.Require().NoError(err)

	testCases := []struct {
		name        string
		spec        string
		tenantID    uuid.UUID
		request     *ListAttributeDefinitionsRequest
		expectError string
		validate    func([]*models.AttributeDefinition)
	}{
		{
			name:     "ListAllDefinitionsInTenantA_Success",
			spec:     "ABAC-ATTR-REPO-003",
			tenantID: s.tenantA.ID,
			request:  &ListAttributeDefinitionsRequest{Limit: 100},
			validate: func(definitions []*models.AttributeDefinition) {
				require.GreaterOrEqual(s.T(), len(definitions), len(defsA))

				// All definitions should belong to tenant A
				tenantADefIDs := make(map[uuid.UUID]bool)
				for _, def := range definitions {
					require.Equal(s.T(), s.tenantA.ID, def.TenantID)
					tenantADefIDs[def.ID] = true
				}

				// Should contain our created definitions
				for _, expectedDef := range defsA {
					require.True(s.T(), tenantADefIDs[expectedDef.ID],
						"Should find definition: %s", expectedDef.Name)
				}

				// Should not contain tenant B definitions
				for _, tenantBDef := range defsB {
					require.False(s.T(), tenantADefIDs[tenantBDef.ID],
						"Should not find tenant B definition: %s", tenantBDef.Name)
				}
			},
		},
		{
			name:     "ListDefinitionsByCategory_Success",
			spec:     "ABAC-ATTR-REPO-003",
			tenantID: s.tenantA.ID,
			request: &ListAttributeDefinitionsRequest{
				Category: &[]types.AttributeCategory{types.AttributeCategoryUser}[0],
				Limit:    100,
			},
			validate: func(definitions []*models.AttributeDefinition) {
				require.GreaterOrEqual(s.T(), len(definitions), 1)

				// All should be user category
				for _, def := range definitions {
					require.Equal(s.T(), types.AttributeCategoryUser, def.Category)
					require.Equal(s.T(), s.tenantA.ID, def.TenantID)
				}
			},
		},
		{
			name:     "ListRequiredDefinitions_Success",
			spec:     "ABAC-ATTR-REPO-003",
			tenantID: s.tenantA.ID,
			request: &ListAttributeDefinitionsRequest{
				IsRequired: &[]bool{true}[0],
				Limit:      100,
			},
			validate: func(definitions []*models.AttributeDefinition) {
				require.GreaterOrEqual(s.T(), len(definitions), 1)

				// All should be required
				for _, def := range definitions {
					require.True(s.T(), def.IsRequired)
					require.Equal(s.T(), s.tenantA.ID, def.TenantID)
				}
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("%s_%s", tc.spec, tc.name), func() {
			var definitions []*models.AttributeDefinition
			var err error

			execErr := s.runner.GetStore().WithTenant(s.ctx, tc.tenantID, func(ctx context.Context, store db.Store) error {
				definitions, err = s.repo.ListAttributeDefinitions(ctx, tc.request)
				return nil
			})
			s.Require().NoError(execErr)

			if tc.expectError != "" {
				require.Error(s.T(), err)
				require.Contains(s.T(), err.Error(), tc.expectError)
			} else {
				require.NoError(s.T(), err)
				require.NotNil(s.T(), definitions)
				if tc.validate != nil {
					tc.validate(definitions)
				}
			}
		})
	}
}

// TestCreateAttributeValue tests ABAC-ATTR-REPO-004: Attribute Value Creation with Tenant Isolation
func (s *AttributeRepositoryTestSuite) TestCreateAttributeValue() {
	// Create attribute definition first
	var definition *models.AttributeDefinition
	err := s.runner.GetStore().WithTenant(s.ctx, s.tenantA.ID, func(ctx context.Context, store db.Store) error {
		var createErr error
		definition, createErr = s.repo.CreateAttributeDefinition(ctx, &CreateAttributeDefinitionRequest{
			Name:        fmt.Sprintf("value_test_def_%s", uuid.New().String()[0:8]),
			DisplayName: stringPtr("Value Test Definition"),
			DataType:    types.AttributeDataTypeString,
			Category:    types.AttributeCategoryUser,
			IsRequired:  false,
			IsActive:    true,
		})
		return createErr
	})
	s.Require().NoError(err)
	s.trackAttributeDefinitionForCleanup(definition.ID)

	testCases := []struct {
		name        string
		spec        string
		tenantID    uuid.UUID
		request     *CreateAttributeValueRequest
		expectError string
		validate    func(*models.AttributeValue)
	}{
		{
			name:     "CreateValidAttributeValue_Success",
			spec:     "ABAC-ATTR-REPO-004",
			tenantID: s.tenantA.ID,
			request: &CreateAttributeValueRequest{
				DefinitionID:  definition.ID,
				EntityID:      uuid.New(),
				Value:         "engineering",
				IsEncrypted:   false,
				Version:       1,
				EffectiveFrom: time.Now(),
				CreatedBy:     uuid.New(),
			},
			validate: func(value *models.AttributeValue) {
				require.NotNil(s.T(), value)
				require.NotEqual(s.T(), uuid.Nil, value.ID)
				require.Equal(s.T(), definition.ID, value.DefinitionID)
				require.Equal(s.T(), "engineering", value.Value)
				require.False(s.T(), value.IsEncrypted)
				require.Equal(s.T(), int32(1), value.Version)
			},
		},
		{
			name:     "CreateAttributeValueInvalidDefinition_Error",
			spec:     "ABAC-ATTR-REPO-004",
			tenantID: s.tenantA.ID,
			request: &CreateAttributeValueRequest{
				DefinitionID:  uuid.New(), // Non-existent definition
				EntityID:      uuid.New(),
				Value:         "test",
				Version:       1,
				EffectiveFrom: time.Now(),
				CreatedBy:     uuid.New(),
			},
			expectError: "foreign key",
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("%s_%s", tc.spec, tc.name), func() {
			var value *models.AttributeValue
			var err error

			execErr := s.runner.GetStore().WithTenant(s.ctx, tc.tenantID, func(ctx context.Context, store db.Store) error {
				value, err = s.repo.CreateAttributeValue(ctx, tc.request)
				return nil
			})
			s.Require().NoError(execErr)

			if tc.expectError != "" {
				require.Error(s.T(), err)
				require.Contains(s.T(), err.Error(), tc.expectError)
				require.Nil(s.T(), value)
			} else {
				require.NoError(s.T(), err)
				require.NotNil(s.T(), value)
				if tc.validate != nil {
					tc.validate(value)
				}
				s.trackAttributeValueForCleanup(value.ID)
			}
		})
	}
}

// TestGetAttributeValuesByEntity tests ABAC-ATTR-REPO-005: Attribute Value Retrieval by Entity
func (s *AttributeRepositoryTestSuite) TestGetAttributeValuesByEntity() {
	// Create attribute definitions and values for testing
	var userDef, resourceDef *models.AttributeDefinition
	testEntityID := uuid.New()

	err := s.runner.GetStore().WithTenant(s.ctx, s.tenantA.ID, func(ctx context.Context, store db.Store) error {
		// Create user attribute definition
		var createErr error
		userDef, createErr = s.repo.CreateAttributeDefinition(ctx, &CreateAttributeDefinitionRequest{
			Name:     fmt.Sprintf("user_attr_%s", uuid.New().String()[0:8]),
			DataType: types.AttributeDataTypeString,
			Category: types.AttributeCategoryUser,
			IsActive: true,
		})
		if createErr != nil {
			return createErr
		}
		s.trackAttributeDefinitionForCleanup(userDef.ID)

		// Create resource attribute definition
		resourceDef, createErr = s.repo.CreateAttributeDefinition(ctx, &CreateAttributeDefinitionRequest{
			Name:     fmt.Sprintf("resource_attr_%s", uuid.New().String()[0:8]),
			DataType: types.AttributeDataTypeString,
			Category: types.AttributeCategoryResource,
			IsActive: true,
		})
		if createErr != nil {
			return createErr
		}
		s.trackAttributeDefinitionForCleanup(resourceDef.ID)

		// Create attribute values
		userValue, createErr := s.repo.CreateAttributeValue(ctx, &CreateAttributeValueRequest{
			DefinitionID:  userDef.ID,
			EntityID:      testEntityID,
			Value:         "user_value",
			Version:       1,
			EffectiveFrom: time.Now(),
			CreatedBy:     uuid.New(),
		})
		if createErr != nil {
			return createErr
		}
		s.trackAttributeValueForCleanup(userValue.ID)

		resourceValue, createErr := s.repo.CreateAttributeValue(ctx, &CreateAttributeValueRequest{
			DefinitionID:  resourceDef.ID,
			EntityID:      testEntityID,
			Value:         "resource_value",
			Version:       1,
			EffectiveFrom: time.Now(),
			CreatedBy:     uuid.New(),
		})
		if createErr != nil {
			return createErr
		}
		s.trackAttributeValueForCleanup(resourceValue.ID)

		return nil
	})
	s.Require().NoError(err)

	testCases := []struct {
		name        string
		spec        string
		tenantID    uuid.UUID
		entityID    uuid.UUID
		category    types.AttributeCategory
		expectError string
		validate    func([]*models.AttributeValue)
	}{
		{
			name:     "GetUserAttributesByEntity_Success",
			spec:     "ABAC-ATTR-REPO-005",
			tenantID: s.tenantA.ID,
			entityID: testEntityID,
			category: types.AttributeCategoryUser,
			validate: func(values []*models.AttributeValue) {
				require.GreaterOrEqual(s.T(), len(values), 1)

				// All values should be user category and match entity
				found := false
				for _, value := range values {
					require.Equal(s.T(), testEntityID, value.EntityID)
					if value.DefinitionID == userDef.ID {
						require.Equal(s.T(), "user_value", value.Value)
						found = true
					}
				}
				require.True(s.T(), found, "Should find the user attribute value")
			},
		},
		{
			name:     "GetResourceAttributesByEntity_Success",
			spec:     "ABAC-ATTR-REPO-005",
			tenantID: s.tenantA.ID,
			entityID: testEntityID,
			category: types.AttributeCategoryResource,
			validate: func(values []*models.AttributeValue) {
				require.GreaterOrEqual(s.T(), len(values), 1)

				// All values should be resource category and match entity
				found := false
				for _, value := range values {
					require.Equal(s.T(), testEntityID, value.EntityID)
					if value.DefinitionID == resourceDef.ID {
						require.Equal(s.T(), "resource_value", value.Value)
						found = true
					}
				}
				require.True(s.T(), found, "Should find the resource attribute value")
			},
		},
		{
			name:     "GetAttributesByEntityDifferentTenant_Empty",
			spec:     "ABAC-ATTR-REPO-005",
			tenantID: s.tenantB.ID, // Different tenant
			entityID: testEntityID,
			category: types.AttributeCategoryUser,
			validate: func(values []*models.AttributeValue) {
				// Should not find values from tenant A
				for _, value := range values {
					require.NotEqual(s.T(), userDef.ID, value.DefinitionID,
						"Should not find tenant A attribute in tenant B context")
					require.NotEqual(s.T(), resourceDef.ID, value.DefinitionID,
						"Should not find tenant A attribute in tenant B context")
				}
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("%s_%s", tc.spec, tc.name), func() {
			var values []*models.AttributeValue
			var err error

			execErr := s.runner.GetStore().WithTenant(s.ctx, tc.tenantID, func(ctx context.Context, store db.Store) error {
				values, err = s.repo.GetAttributeValuesByEntity(ctx, tc.entityID, tc.category)
				return nil
			})
			s.Require().NoError(execErr)

			if tc.expectError != "" {
				require.Error(s.T(), err)
				require.Contains(s.T(), err.Error(), tc.expectError)
			} else {
				require.NoError(s.T(), err)
				require.NotNil(s.T(), values)
				if tc.validate != nil {
					tc.validate(values)
				}
			}
		})
	}
}

// TestMultiTenantAttributeIsolation tests ABAC-ATTR-REPO-006:  Multi-tenant RLS for Attributes
func (s *AttributeRepositoryTestSuite) TestMultiTenantAttributeIsolation() {
	s.Run("ABAC-ATTR-REPO-006_AttributeRLS", func() {
		// Create attribute definitions in both tenants
		var defsA, defsB []*models.AttributeDefinition

		// Create definitions in tenant A
		err := s.runner.GetStore().WithTenant(s.ctx, s.tenantA.ID, func(ctx context.Context, store db.Store) error {
			for i := 0; i < 3; i++ {
				def, createErr := s.repo.CreateAttributeDefinition(ctx, &CreateAttributeDefinitionRequest{
					Name:        fmt.Sprintf("isolation_test_a_%d_%s", i, uuid.New().String()[0:8]),
					DisplayName: stringPtr(fmt.Sprintf("Isolation Test A %d", i)),
					DataType:    types.AttributeDataTypeString,
					Category:    types.AttributeCategoryUser,
					IsActive:    true,
				})
				if createErr != nil {
					return createErr
				}
				defsA = append(defsA, def)
				s.trackAttributeDefinitionForCleanup(def.ID)
			}
			return nil
		})
		s.Require().NoError(err)

		// Create definitions in tenant B
		err = s.runner.GetStore().WithTenant(s.ctx, s.tenantB.ID, func(ctx context.Context, store db.Store) error {
			for i := 0; i < 2; i++ {
				def, createErr := s.repo.CreateAttributeDefinition(ctx, &CreateAttributeDefinitionRequest{
					Name:        fmt.Sprintf("isolation_test_b_%d_%s", i, uuid.New().String()[0:8]),
					DisplayName: stringPtr(fmt.Sprintf("Isolation Test B %d", i)),
					DataType:    types.AttributeDataTypeNumber,
					Category:    types.AttributeCategoryResource,
					IsActive:    true,
				})
				if createErr != nil {
					return createErr
				}
				defsB = append(defsB, def)
				s.trackAttributeDefinitionForCleanup(def.ID)
			}
			return nil
		})
		s.Require().NoError(err)

		// Test 1: Verify tenant A can only see its own definitions
		err = s.runner.GetStore().WithTenant(s.ctx, s.tenantA.ID, func(ctx context.Context, store db.Store) error {
			definitions, listErr := s.repo.ListAttributeDefinitions(ctx, &ListAttributeDefinitionsRequest{Limit: 100})
			if listErr != nil {
				return listErr
			}

			tenantADefIDs := make(map[uuid.UUID]bool)
			for _, def := range definitions {
				require.Equal(s.T(), s.tenantA.ID, def.TenantID,
					"All definitions should belong to tenant A")
				tenantADefIDs[def.ID] = true
			}

			// Should find all tenant A definitions
			for _, expectedDef := range defsA {
				require.True(s.T(), tenantADefIDs[expectedDef.ID],
					"Should find tenant A definition: %s", expectedDef.Name)
			}

			// Should not find any tenant B definitions
			for _, tenantBDef := range defsB {
				require.False(s.T(), tenantADefIDs[tenantBDef.ID],
					"Should not find tenant B definition: %s", tenantBDef.Name)
			}

			return nil
		})
		s.Require().NoError(err)

		// Test 2: Verify tenant B can only see its own definitions
		err = s.runner.GetStore().WithTenant(s.ctx, s.tenantB.ID, func(ctx context.Context, store db.Store) error {
			definitions, listErr := s.repo.ListAttributeDefinitions(ctx, &ListAttributeDefinitionsRequest{Limit: 100})
			if listErr != nil {
				return listErr
			}

			tenantBDefIDs := make(map[uuid.UUID]bool)
			for _, def := range definitions {
				require.Equal(s.T(), s.tenantB.ID, def.TenantID,
					"All definitions should belong to tenant B")
				tenantBDefIDs[def.ID] = true
			}

			// Should find all tenant B definitions
			for _, expectedDef := range defsB {
				require.True(s.T(), tenantBDefIDs[expectedDef.ID],
					"Should find tenant B definition: %s", expectedDef.Name)
			}

			// Should not find any tenant A definitions
			for _, tenantADef := range defsA {
				require.False(s.T(), tenantBDefIDs[tenantADef.ID],
					"Should not find tenant A definition: %s", tenantADef.Name)
			}

			return nil
		})
		s.Require().NoError(err)

		// Test 3: Verify cross-tenant individual definition access is blocked
		for _, tenantADef := range defsA {
			err = s.runner.GetStore().WithTenant(s.ctx, s.tenantB.ID, func(ctx context.Context, store db.Store) error {
				_, getErr := s.repo.GetAttributeDefinitionByID(ctx, tenantADef.ID)
				require.Error(s.T(), getErr, "Tenant B should not access tenant A definition")
				require.Contains(s.T(), getErr.Error(), "not found")
				return nil
			})
			s.Require().NoError(err)
		}

		s.T().Logf("✅ Multi-tenant attribute RLS verification completed successfully")
		s.T().Logf("   - Verified isolation for %d tenant A definitions", len(defsA))
		s.T().Logf("   - Verified isolation for %d tenant B definitions", len(defsB))
	})
}

// TestAttributeDefinitionByName tests ABAC-ATTR-REPO-007: Attribute Definition Name-based Retrieval
func (s *AttributeRepositoryTestSuite) TestAttributeDefinitionByName() {
	// Create attribute definition with specific name
	definitionName := fmt.Sprintf("unique_name_test_%s", uuid.New().String()[0:8])
	var defA *models.AttributeDefinition

	err := s.runner.GetStore().WithTenant(s.ctx, s.tenantA.ID, func(ctx context.Context, store db.Store) error {
		var createErr error
		defA, createErr = s.repo.CreateAttributeDefinition(ctx, &CreateAttributeDefinitionRequest{
			Name:        definitionName,
			DisplayName: stringPtr("Unique Name Test"),
			DataType:    types.AttributeDataTypeString,
			Category:    types.AttributeCategoryUser,
			IsActive:    true,
		})
		return createErr
	})
	s.Require().NoError(err)
	s.trackAttributeDefinitionForCleanup(defA.ID)

	testCases := []struct {
		name        string
		spec        string
		tenantID    uuid.UUID
		defName     string
		expectError string
		validate    func(*models.AttributeDefinition)
	}{
		{
			name:     "GetByNameInSameTenant_Success",
			spec:     "ABAC-ATTR-REPO-007",
			tenantID: s.tenantA.ID,
			defName:  definitionName,
			validate: func(def *models.AttributeDefinition) {
				require.NotNil(s.T(), def)
				require.Equal(s.T(), defA.ID, def.ID)
				require.Equal(s.T(), definitionName, def.Name)
				require.Equal(s.T(), s.tenantA.ID, def.TenantID)
			},
		},
		{
			name:        "GetByNameInDifferentTenant_NotFound",
			spec:        "ABAC-ATTR-REPO-007",
			tenantID:    s.tenantB.ID,
			defName:     definitionName,
			expectError: "not found",
		},
		{
			name:        "GetByNonExistentName_NotFound",
			spec:        "ABAC-ATTR-REPO-007",
			tenantID:    s.tenantA.ID,
			defName:     "non_existent_name",
			expectError: "not found",
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("%s_%s", tc.spec, tc.name), func() {
			var definition *models.AttributeDefinition
			var err error

			execErr := s.runner.GetStore().WithTenant(s.ctx, tc.tenantID, func(ctx context.Context, store db.Store) error {
				definition, err = s.repo.GetAttributeDefinitionByName(ctx, tc.defName)
				return nil
			})
			s.Require().NoError(execErr)

			if tc.expectError != "" {
				require.Error(s.T(), err)
				require.Contains(s.T(), err.Error(), tc.expectError)
				require.Nil(s.T(), definition)
			} else {
				require.NoError(s.T(), err)
				require.NotNil(s.T(), definition)
				if tc.validate != nil {
					tc.validate(definition)
				}
			}
		})
	}
}

// TestAttributePerformance tests ABAC-ATTR-REPO-008: Performance and Scalability for Attributes
func (s *AttributeRepositoryTestSuite) TestAttributePerformance() {
	s.Run("ABAC-ATTR-REPO-008_AttributePerformanceBaseline", func() {
		const definitionCount = 20 // Reasonable number for unit tests
		var createdDefinitions []*models.AttributeDefinition

		// Create multiple attribute definitions and measure time
		startTime := time.Now()

		err := s.runner.GetStore().WithTenant(s.ctx, s.tenantA.ID, func(ctx context.Context, store db.Store) error {
			categories := []types.AttributeCategory{
				types.AttributeCategoryUser,
				types.AttributeCategoryResource,
				types.AttributeCategoryEnvironment,
				types.AttributeCategoryAction,
			}

			for i := 0; i < definitionCount; i++ {
				category := categories[i%len(categories)]
				def, createErr := s.repo.CreateAttributeDefinition(ctx, &CreateAttributeDefinitionRequest{
					Name:        fmt.Sprintf("perf_test_%d_%s", i, uuid.New().String()[0:8]),
					DisplayName: stringPtr(fmt.Sprintf("Performance Test %d", i)),
					DataType:    types.AttributeDataTypeString,
					Category:    category,
					IsRequired:  i%3 == 0,
					IsActive:    true,
				})
				if createErr != nil {
					return createErr
				}
				createdDefinitions = append(createdDefinitions, def)
				s.trackAttributeDefinitionForCleanup(def.ID)
			}
			return nil
		})
		s.Require().NoError(err)

		creationTime := time.Since(startTime)
		s.T().Logf("Created %d attribute definitions in %v (avg: %v per definition)",
			definitionCount, creationTime, creationTime/definitionCount)

		// Test retrieval performance
		startTime = time.Now()

		err = s.runner.GetStore().WithTenant(s.ctx, s.tenantA.ID, func(ctx context.Context, store db.Store) error {
			for _, def := range createdDefinitions[:5] { // Test first 5
				retrievedDef, getErr := s.repo.GetAttributeDefinitionByID(ctx, def.ID)
				if getErr != nil {
					return getErr
				}
				require.NotNil(s.T(), retrievedDef)
				require.Equal(s.T(), def.ID, retrievedDef.ID)
			}
			return nil
		})
		s.Require().NoError(err)

		retrievalTime := time.Since(startTime)
		s.T().Logf("Retrieved 5 definitions in %v (avg: %v per retrieval)",
			retrievalTime, retrievalTime/5)

		// Test list performance
		startTime = time.Now()

		err = s.runner.GetStore().WithTenant(s.ctx, s.tenantA.ID, func(ctx context.Context, store db.Store) error {
			definitions, listErr := s.repo.ListAttributeDefinitions(ctx, &ListAttributeDefinitionsRequest{Limit: 100})
			if listErr != nil {
				return listErr
			}
			require.GreaterOrEqual(s.T(), len(definitions), definitionCount)
			return nil
		})
		s.Require().NoError(err)

		listTime := time.Since(startTime)
		s.T().Logf("Listed all definitions in %v", listTime)

		// Performance assertions
		require.Less(s.T(), creationTime.Milliseconds(), int64(2000), // 2 seconds max for 20 definitions
			"Definition creation should be reasonably fast")
		require.Less(s.T(), retrievalTime.Milliseconds(), int64(200), // 200ms max for 5 retrievals
			"Definition retrieval should be fast")
		require.Less(s.T(), listTime.Milliseconds(), int64(500), // 500ms max for listing
			"Definition listing should be fast")

		s.T().Logf("✅ Attribute repository performance baseline established successfully")
	})
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}
