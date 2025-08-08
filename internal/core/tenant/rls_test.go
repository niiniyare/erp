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
func (s *RLSTestSuite) TestTenantDataIsolation() {
	// 1. Setup: Create two tenants
	uniqueID := uuid.New().String()
	tenantA, err := s.runner.store.CreateTenant(s.ctx, db.CreateTenantParams{
		Name:  fmt.Sprintf("Wayne Enterprises %s", uniqueID),
		Slug:  fmt.Sprintf("wayne-enterprises-%s", uniqueID[0:8]),
		Email: fmt.Sprintf("bruce-%s@wayne.com", uniqueID),
		Status: "active",
	})
	s.Require().NoError(err)
	defer s.runner.store.SoftDeleteTenant(s.ctx, tenantA.ID)

	uniqueID2 := uuid.New().String()
	tenantB, err := s.runner.store.CreateTenant(s.ctx, db.CreateTenantParams{
		Name:  fmt.Sprintf("Stark Industries %s", uniqueID2),
		Slug:  fmt.Sprintf("stark-industries-%s", uniqueID2[0:8]),
		Email: fmt.Sprintf("tony-%s@stark.com", uniqueID2),
		Status: "active",
	})
	s.Require().NoError(err)
	defer s.runner.store.SoftDeleteTenant(s.ctx, tenantB.ID)

	// 2. Set context for Tenant A and create an entity
	ctxTenantA := context.WithValue(s.ctx, "tenant_id", tenantA.ID)
	err = s.runner.store.SetTenantContext(ctxTenantA, tenantA.ID)
	s.Require().NoError(err)

	entityA, err := s.runner.store.CreateEntity(ctxTenantA, db.CreateEntityParams{
		Uuid: uuid.New(),
		Name: "Gotham HQ",
		Code: stringPtr("gotham-hq"),
		Type: "COMPANY",
		IsActive: true,
		AccrualMethod: true,
		FyStartMonth: 1,
	})
	s.Require().NoError(err)

	// 3. Verify Tenant A can only see its own entity
	entitiesA, err := s.runner.store.ListEntities(ctxTenantA)
	s.Require().NoError(err)
	s.Require().Len(entitiesA, 1, "Tenant A should only see 1 entity")
	s.Require().Equal(entityA.Uuid, entitiesA[0].Uuid, "The entity UUID should match for Tenant A")

	// 4. Set context for Tenant B
	ctxTenantB := context.WithValue(s.ctx, "tenant_id", tenantB.ID)
	err = s.runner.store.SetTenantContext(ctxTenantB, tenantB.ID)
	s.Require().NoError(err)

	// 5. Verify Tenant B cannot see Tenant A's entity
	entitiesB, err := s.runner.store.ListEntities(ctxTenantB)
	s.Require().NoError(err)
	s.Require().Len(entitiesB, 0, "Tenant B should not see any of Tenant A's entities")

	// 6. Create an entity for Tenant B
	entityB, err := s.runner.store.CreateEntity(ctxTenantB, db.CreateEntityParams{
		Uuid: uuid.New(),
		Name: "Stark Tower",
		Code: stringPtr("stark-tower"),
		Type: "COMPANY",
		IsActive: true,
		AccrualMethod: true,
		FyStartMonth: 1,
	})
	s.Require().NoError(err)

	// 7. Verify Tenant B can now see its own entity
	entitiesB, err = s.runner.store.ListEntities(ctxTenantB)
	s.Require().NoError(err)
	s.Require().Len(entitiesB, 1, "Tenant B should now see 1 entity")
	s.Require().Equal(entityB.Uuid, entitiesB[0].Uuid)

	// 8. Verify Tenant B cannot update Tenant A's entity
	_, err = s.runner.store.UpdateEntity(ctxTenantB, db.UpdateEntityParams{Uuid: entityA.Uuid, Name: "Updated by Stark"})
	s.Require().Error(err, "Expected an error when Tenant B tries to update Tenant A's entity")

	// 9. Verify Tenant B cannot delete Tenant A's entity
	err = s.runner.store.WithTenant(s.ctx, tenantB.ID, func(ctx context.Context, q db.Store) error {
		// Explicitly set role to application_role to ensure RLS is active
		// Access the underlying pool from the store to execute raw SQL
		_, err := s.runner.pool.Exec(ctx, "SET ROLE application_role")
		if err != nil {
			return err
		}
		// Attempt to delete entityA. This should fail due to RLS.
		return q.SoftDeleteEntity(ctx, entityA.Uuid)
	})
	s.Require().Error(err, "Expected an error when Tenant B tries to delete Tenant A's entity")
}