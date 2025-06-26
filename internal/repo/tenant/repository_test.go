package tenant

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/niiniyare/erp/internal/domain/tenant"
	"github.com/niiniyare/erp/test/testutil"
)

func TestTenantRepository_Create(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := NewTenantRepository(tdb.Pool)
	ctx := context.Background()

	req := &tenant.CreateTenantRequest{
		Name:     "Test Tenant",
		Status:   tenant.TenantStatusActive,
		Industry: stringPtr("Technology"),
	}

	createdTenant, err := repo.Create(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, createdTenant)
	assert.Equal(t, req.Name, createdTenant.Name)
	assert.Equal(t, req.Status, createdTenant.Status)
	assert.NotZero(t, createdTenant.ID)
	assert.NotZero(t, createdTenant.Uuid)
}

func TestTenantRepository_GetByID(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := NewTenantRepository(tdb.Pool)
	ctx := context.Background()

	// Create a tenant first
	req := &tenant.CreateTenantRequest{
		Name:   "Test Tenant",
		Status: tenant.TenantStatusActive,
	}
	createdTenant, err := repo.Create(ctx, req)
	require.NoError(t, err)

	// Test GetByID
	foundTenant, err := repo.GetByID(ctx, createdTenant.ID)
	require.NoError(t, err)
	assert.Equal(t, createdTenant.ID, foundTenant.ID)
	assert.Equal(t, createdTenant.Name, foundTenant.Name)
	assert.Equal(t, createdTenant.Status, foundTenant.Status)
}

func TestTenantRepository_GetByUUID(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := NewTenantRepository(tdb.Pool)
	ctx := context.Background()

	// Create a tenant first
	req := &tenant.CreateTenantRequest{
		Name:   "Test Tenant",
		Status: tenant.TenantStatusActive,
	}
	createdTenant, err := repo.Create(ctx, req)
	require.NoError(t, err)

	// Test GetByUUID
	foundTenant, err := repo.GetByUUID(ctx, createdTenant.Uuid)
	require.NoError(t, err)
	assert.Equal(t, createdTenant.Uuid, foundTenant.Uuid)
	assert.Equal(t, createdTenant.Name, foundTenant.Name)
}

func TestTenantRepository_GetBySubdomain(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := NewTenantRepository(tdb.Pool)
	ctx := context.Background()

	subdomain := "test-subdomain"
	req := &tenant.CreateTenantRequest{
		Name:      "Test Tenant",
		Subdomain: &subdomain,
		Status:    tenant.TenantStatusActive,
	}

	createdTenant, err := repo.Create(ctx, req)
	require.NoError(t, err)

	// Test GetBySubdomain
	foundTenant, err := repo.GetBySubdomain(ctx, subdomain)
	require.NoError(t, err)
	assert.Equal(t, createdTenant.ID, foundTenant.ID)
	assert.Equal(t, subdomain, foundTenant.Subdomain)
}

func TestTenantRepository_Update(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := NewTenantRepository(tdb.Pool)
	ctx := context.Background()

	// Create a tenant first
	req := &tenant.CreateTenantRequest{
		Name:   "Test Tenant",
		Status: tenant.TenantStatusActive,
	}
	createdTenant, err := repo.Create(ctx, req)
	require.NoError(t, err)

	// Update the tenant
	newName := "Updated Tenant"
	suspendedStatus := tenant.TenantStatusSuspended
	updateReq := &tenant.UpdateTenantRequest{
		Name:   &newName,
		Status: &suspendedStatus,
	}

	updatedTenant, err := repo.Update(ctx, createdTenant.ID, updateReq)
	require.NoError(t, err)
	assert.Equal(t, newName, updatedTenant.Name)
	assert.Equal(t, suspendedStatus, updatedTenant.Status)
}

func TestTenantRepository_Delete(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := NewTenantRepository(tdb.Pool)
	ctx := context.Background()

	// Create a tenant first
	req := &tenant.CreateTenantRequest{
		Name:   "Test Tenant",
		Status: tenant.TenantStatusActive,
	}
	createdTenant, err := repo.Create(ctx, req)
	require.NoError(t, err)

	// Delete the tenant
	err = repo.Delete(ctx, createdTenant.ID)
	require.NoError(t, err)

	// Verify it's deleted (should return error)
	_, err = repo.GetByID(ctx, createdTenant.ID)
	assert.Error(t, err)
}

func TestTenantRepository_List(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := NewTenantRepository(tdb.Pool)
	ctx := context.Background()

	// Create multiple tenants
	for i := 0; i < 3; i++ {
		req := &tenant.CreateTenantRequest{
			Name:   fmt.Sprintf("Test Tenant %d", i+1),
			Status: tenant.TenantStatusActive,
		}
		_, err := repo.Create(ctx, req)
		require.NoError(t, err)
	}

	// Test List
	tenants, err := repo.List(ctx, 10, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(tenants), 3)
}

func TestTenantRepository_UpdateStatus(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := NewTenantRepository(tdb.Pool)
	ctx := context.Background()

	// Create a tenant first
	req := &tenant.CreateTenantRequest{
		Name:   "Test Tenant",
		Status: tenant.TenantStatusActive,
	}
	createdTenant, err := repo.Create(ctx, req)
	require.NoError(t, err)

	// Update status to suspended
	updatedTenant, err := repo.UpdateStatus(ctx, createdTenant.ID, tenant.TenantStatusSuspended)
	require.NoError(t, err)
	assert.Equal(t, tenant.TenantStatusSuspended, updatedTenant.Status)
}

func TestTenantRepository_CheckExists(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := NewTenantRepository(tdb.Pool)
	ctx := context.Background()

	// Create a tenant first
	req := &tenant.CreateTenantRequest{
		Name:   "Test Tenant",
		Status: tenant.TenantStatusActive,
	}
	createdTenant, err := repo.Create(ctx, req)
	require.NoError(t, err)

	// Check if it exists
	exists, err := repo.CheckExists(ctx, createdTenant.ID)
	require.NoError(t, err)
	assert.True(t, exists)

	// Check non-existing tenant
	exists, err = repo.CheckExists(ctx, 99999)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestTenantRepository_CheckSubdomainExists(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := NewTenantRepository(tdb.Pool)
	ctx := context.Background()

	subdomain := "unique-subdomain"
	req := &tenant.CreateTenantRequest{
		Name:      "Test Tenant",
		Subdomain: &subdomain,
		Status:    tenant.TenantStatusActive,
	}
	_, err := repo.Create(ctx, req)
	require.NoError(t, err)

	// Check if subdomain exists
	exists, err := repo.CheckSubdomainExists(ctx, subdomain)
	require.NoError(t, err)
	assert.True(t, exists)

	// Check non-existing subdomain
	exists, err = repo.CheckSubdomainExists(ctx, "non-existing")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestTenantRepository_GetActiveTenants(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := NewTenantRepository(tdb.Pool)
	ctx := context.Background()

	// Create active tenant
	activeReq := &tenant.CreateTenantRequest{
		Name:   "Active Tenant",
		Status: tenant.TenantStatusActive,
	}
	_, err := repo.Create(ctx, activeReq)
	require.NoError(t, err)

	// Create suspended tenant
	suspendedReq := &tenant.CreateTenantRequest{
		Name:   "Suspended Tenant",
		Status: tenant.TenantStatusSuspended,
	}
	_, err = repo.Create(ctx, suspendedReq)
	require.NoError(t, err)

	// Get only active tenants
	activeTenants, err := repo.GetActiveTenants(ctx)
	require.NoError(t, err)
	
	// Verify all returned tenants are active
	for _, tenant := range activeTenants {
		assert.Equal(t, tenant.TenantStatusActive, tenant.Status)
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}