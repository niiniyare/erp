//go:build database
// +build database

package tenant

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/stretchr/testify/suite"
)

// RLSTestSuite is the test suite for Row-Level Security tests
type RLSTestSuite struct {
	suite.Suite
	runner *DatabaseTestRunner
	ctx    context.Context
}

// SetupSuite runs once before the entire test suite
func (s *RLSTestSuite) SetupSuite() {
	var err error
	s.runner, err = NewDatabaseTestRunner()
	s.Require().NoError(err, "Failed to connect to the database")
	s.ctx = context.Background()
}

// TearDownSuite runs once after the entire test suite
func (s *RLSTestSuite) TearDownSuite() {
	s.runner.Close()
}

// TestRLS runs the RLS test suite
func TestRLS(t *testing.T) {
	suite.Run(t, new(RLSTestSuite))
}

// TestTenantDataIsolation covers test cases MT-RLS-001 and MT-RLS-002
// It has been refactored to use the WithTenant transactional context method.
func (s *RLSTestSuite) TestTenantDataIsolation() {
	// 1. Setup: Create two tenants with unique identifiers using a superuser context
	superuserStore := db.NewStore(s.runner.pool)
	uniqueID := uuid.New().String()
	tenantASlug := fmt.Sprintf("wayne-enterprises-%s", uniqueID[0:13])
	tenantA, err := superuserStore.CreateTenant(s.ctx, db.CreateTenantParams{
		Name:   fmt.Sprintf("Wayne Enterprises %s", uniqueID[0:8]),
		Slug:   tenantASlug,
		Email:  fmt.Sprintf("bruce-%s@wayne.com", uniqueID[0:8]),
		Status: "active",
	})
	s.Require().NoError(err, "Failed to create tenant A")
	defer func() {
		if err := superuserStore.SoftDeleteTenant(s.ctx, tenantA.ID); err != nil {
			s.T().Logf("Warning: Failed to cleanup tenant A: %v", err)
		}
	}()

	uniqueID2 := uuid.New().String()
	tenantBSlug := fmt.Sprintf("stark-industries-%s", uniqueID2[0:13])
	tenantB, err := superuserStore.CreateTenant(s.ctx, db.CreateTenantParams{
		Name:   fmt.Sprintf("Stark Industries %s", uniqueID2[0:8]),
		Slug:   tenantBSlug,
		Email:  fmt.Sprintf("tony-%s@stark.com", uniqueID2[0:8]),
		Status: "active",
	})
	s.Require().NoError(err, "Failed to create tenant B")
	defer func() {
		if err := superuserStore.SoftDeleteTenant(s.ctx, tenantB.ID); err != nil {
			s.T().Logf("Warning: Failed to cleanup tenant B: %v", err)
		}
	}()

	var entityA *db.Entity
	// 2. Use WithTenant for Tenant A to create and list an entity
	err = s.runner.store.WithTenant(s.ctx, tenantA.ID, func(ctx context.Context, txStore db.Store) error {
		var innerErr error
		entityA, innerErr = txStore.CreateEntity(ctx, db.CreateEntityParams{
			Uuid:          uuid.New(),
			Name:          fmt.Sprintf("Gotham HQ %s", uniqueID[0:8]),
			Code:          stringPtr(fmt.Sprintf("gotham-hq-%s", uniqueID[0:8])),
			Type:          "COMPANY",
			IsActive:      true,
			AccrualMethod: true,
			FyStartMonth:  1,
		})
		if innerErr != nil {
			return fmt.Errorf("failed to create entity A: %w", innerErr)
		}

		// 3. Verify Tenant A can only see its own entity
		entitiesA, innerErr := txStore.ListEntities(ctx)
		if innerErr != nil {
			return fmt.Errorf("failed to list entities for tenant A: %w", innerErr)
		}
		s.Require().Len(entitiesA, 1, "Tenant A should only see 1 entity")
		s.Require().Equal(entityA.Uuid, entitiesA[0].Uuid, "The entity UUID should match for Tenant A")
		return nil
	})
	s.Require().NoError(err, "Transaction for Tenant A failed")

	// 4. Use WithTenant for Tenant B
	err = s.runner.store.WithTenant(s.ctx, tenantB.ID, func(ctx context.Context, txStore db.Store) error {
		// 5. Verify Tenant B cannot see Tenant A's entity
		entitiesB, innerErr := txStore.ListEntities(ctx)
		if innerErr != nil {
			return fmt.Errorf("failed to list entities for tenant B: %w", innerErr)
		}
		s.Require().Len(entitiesB, 0, "Tenant B should not see any of Tenant A's entities")

		// 6. Create an entity for Tenant B
		entityB, innerErr := txStore.CreateEntity(ctx, db.CreateEntityParams{
			Uuid:          uuid.New(),
			Name:          fmt.Sprintf("Stark Tower %s", uniqueID2[0:8]),
			Code:          stringPtr(fmt.Sprintf("stark-tower-%s", uniqueID2[0:8])),
			Type:          "COMPANY",
			IsActive:      true,
			AccrualMethod: true,
			FyStartMonth:  1,
		})
		if innerErr != nil {
			return fmt.Errorf("failed to create entity B: %w", innerErr)
		}

		// 7. Verify Tenant B can now see its own entity
		entitiesB, innerErr = txStore.ListEntities(ctx)
		if innerErr != nil {
			return fmt.Errorf("failed to list entities for tenant B after creation: %w", innerErr)
		}
		s.Require().Len(entitiesB, 1, "Tenant B should now see 1 entity")
		s.Require().Equal(entityB.Uuid, entitiesB[0].Uuid)

		// 8. Verify Tenant B cannot update Tenant A's entity
		_, innerErr = txStore.UpdateEntity(ctx, db.UpdateEntityParams{Uuid: entityA.Uuid, Name: "Updated by Stark"})
		s.Require().Error(innerErr, "Expected an error when Tenant B tries to update Tenant A's entity")

		// 9. Verify Tenant B cannot delete Tenant A's entity
		innerErr = txStore.SoftDeleteEntity(ctx, entityA.Uuid)
		s.Require().NoError(innerErr, "Cross-tenant soft delete should not produce an error, it should just affect 0 rows.")

		return nil
	})
	s.Require().NoError(err, "Transaction for Tenant B failed")

	// 10. Verify that entity A was NOT deleted by trying to fetch it again within its own context
	err = s.runner.store.WithTenant(s.ctx, tenantA.ID, func(ctx context.Context, txStore db.Store) error {
		_, getErr := txStore.GetEntity(ctx, entityA.Uuid)
		s.Require().NoError(getErr, "Entity A should still exist after a failed cross-tenant delete attempt")
		return nil
	})
	s.Require().NoError(err, "Verification transaction for Tenant A failed")
}