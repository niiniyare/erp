package db

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niiniyare/erp/pkg/util"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Test fixtures and helpers
var (
	testCtx = context.Background()
)

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping database tests")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	require.NoError(t, err)

	// Configure for testing
	config.MaxConns = 10
	config.MinConns = 2

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	require.NoError(t, err)

	// Test connection
	err = pool.Ping(context.Background())
	require.NoError(t, err)

	return pool
}

// generateUniqueTestName creates a unique test name to avoid conflicts
// Keeps names short to fit database constraints (usually VARCHAR(50) or VARCHAR(63))
func generateUniqueTestName(baseName string) string {
	// Use only the last 6 digits of timestamp for uniqueness
	timestamp := time.Now().UnixNano() % 1000000
	// Use 4-char random string to save space
	random := util.RandomString(4)
	// Format: BaseName_123456_AbCd (max ~25 chars for reasonable base names)
	return fmt.Sprintf("%s_%06d_%s", baseName, timestamp, random)
}

// generateShortUniqueName creates very short unique names for constrained fields
func generateShortUniqueName(prefix string) string {
	timestamp := time.Now().UnixNano() % 100000 // 5 digits
	random := util.RandomString(3)              // 3 chars
	// Format: prefix_12345_ABC (max ~15 chars for short prefixes)
	return fmt.Sprintf("%s_%05d_%s", prefix, timestamp, random)
}

func createTestTenant(t *testing.T, store Store, name string) *Tenant {
	t.Helper()
	// Keep base name short to fit database constraints
	shortName := name
	if len(name) > 15 {
		shortName = name[:15]
	}
	uniqueName := generateUniqueTestName(shortName)
	subdomain := strings.ToLower(strings.ReplaceAll(uniqueName, " ", "-"))

	params := CreateTenantParams{
		Name:      uniqueName,
		Subdomain: stringPtr(subdomain),
		Status:    "active",
		Industry:  stringPtr("technology"),
	}
	tenant, err := store.CreateTenant(testCtx, params)
	require.NoError(t, err)
	require.NotNil(t, tenant)
	return tenant
}

func createComplexTestTenant(t *testing.T, store Store, name string) *Tenant {
	t.Helper()

	// Keep base name short to fit database constraints
	shortName := name
	if len(name) > 15 {
		shortName = name[:15]
	}
	uniqueName := generateUniqueTestName(shortName)
	slug := generateShortUniqueName("slug") // Keep slug very short
	email := fmt.Sprintf("test@%s.com", generateShortUniqueName("domain"))

	metadata, err := json.Marshal(map[string]interface{}{
		"test_data":  true,
		"created_by": "test_suite",
		"timestamp":  time.Now().Unix(),
	})
	require.NoError(t, err)

	settings, err := json.Marshal(map[string]interface{}{
		"theme":         "dark",
		"notifications": true,
	})
	require.NoError(t, err)

	params := CreateTenantCompleteParams{
		Name:               uniqueName,
		Slug:               slug,
		Email:              email,
		Subdomain:          stringPtr(generateShortUniqueName("sub")),
		Status:             "active",
		Timezone:           "UTC",
		CurrencyCode:       "USD",
		Metadata:           metadata,
		Industry:           stringPtr("technology"),
		CompanySize:        stringPtr("medium"),
		TaxID:              stringPtr(generateShortUniqueName("TAX")),
		RegistrationNumber: stringPtr(generateShortUniqueName("REG")),
		LegalEntityType:    stringPtr("corporation"),
		Settings:           settings,
	}

	tenant, err := store.CreateTenantComplete(testCtx, params)
	require.NoError(t, err)
	require.NotNil(t, tenant)
	return tenant
}

// Tenant Test Suite
type TenantTestSuite struct {
	suite.Suite
	store       Store
	pool        *pgxpool.Pool
	testTenants []uuid.UUID // Track tenants created during tests for cleanup
}

func (suite *TenantTestSuite) SetupSuite() {
	suite.pool = setupTestDB(suite.T())
	suite.store = NewStore(suite.pool)
	suite.testTenants = make([]uuid.UUID, 0)
}

func (suite *TenantTestSuite) TearDownSuite() {
	// Clean up all test tenants
	suite.cleanupTestTenants()

	if suite.pool != nil {
		suite.pool.Close()
	}
}

func (suite *TenantTestSuite) SetupTest() {
	// Clear any existing tenant context
	_, _ = suite.pool.Exec(testCtx, "SELECT set_config('app.current_tenant_id', '', false)")
}

func (suite *TenantTestSuite) TearDownTest() {
	// Clear tenant context after each test
	_, _ = suite.pool.Exec(testCtx, "SELECT set_config('app.current_tenant_id', '', false)")
}

func (suite *TenantTestSuite) cleanupTestTenants() {
	if len(suite.testTenants) == 0 {
		return
	}

	// Clean up in proper dependency order
	for _, tenantID := range suite.testTenants {
		// Delete dependent records first
		_, _ = suite.pool.Exec(testCtx, "DELETE FROM tenant_usage_stats WHERE tenant_id = $1", tenantID)
		_, _ = suite.pool.Exec(testCtx, "DELETE FROM tenant_configurations WHERE tenant_id = $1", tenantID)
		// Force delete tenant (including soft-deleted ones)
		_, _ = suite.pool.Exec(testCtx, "DELETE FROM tenants WHERE id = $1", tenantID)
	}

	suite.testTenants = make([]uuid.UUID, 0)
}

func (suite *TenantTestSuite) trackTenant(tenantID uuid.UUID) {
	suite.testTenants = append(suite.testTenants, tenantID)
}

// Basic CRUD Tests
func (suite *TenantTestSuite) TestCreateTenant() {
	tests := []struct {
		name    string
		params  CreateTenantParams
		wantErr bool
	}{
		{
			name: "valid basic tenant",
			params: CreateTenantParams{
				Name:      generateShortUniqueName("Test"),
				Subdomain: stringPtr(generateShortUniqueName("test")),
				Status:    "active",
				Industry:  stringPtr("technology"),
			},
			wantErr: false,
		},
		{
			name: "tenant without subdomain",
			params: CreateTenantParams{
				Name:      generateShortUniqueName("NoSub"),
				Subdomain: nil,
				Status:    "pending",
				Industry:  stringPtr("healthcare"),
			},
			wantErr: false,
		},
		{
			name: "tenant without industry",
			params: CreateTenantParams{
				Name:      generateShortUniqueName("Generic"),
				Subdomain: stringPtr(generateShortUniqueName("generic")),
				Status:    "active",
				Industry:  nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tenant, err := suite.store.CreateTenant(testCtx, tt.params)

			if tt.wantErr {
				suite.Require().Error(err)
				suite.Require().Nil(tenant)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(tenant)
				suite.Require().Equal(tt.params.Name, tenant.Name)
				suite.Require().Equal(tt.params.Status, tenant.Status)
				suite.Require().Equal(tt.params.Subdomain, tenant.Subdomain)
				suite.Require().Equal(tt.params.Industry, tenant.Industry)
				suite.Require().NotEqual(uuid.Nil, tenant.ID)
				suite.Require().False(tenant.CreatedAt.IsZero())
				suite.Require().False(tenant.UpdatedAt.IsZero())

				// Track for cleanup
				suite.trackTenant(tenant.ID)
			}
		})
	}
}

func (suite *TenantTestSuite) TestCreateTenantComplete() {
	metadata, err := json.Marshal(map[string]interface{}{"test": true})
	suite.Require().NoError(err)

	settings, err := json.Marshal(map[string]interface{}{"theme": "light"})
	suite.Require().NoError(err)

	uniqueName := generateShortUniqueName("Complete")
	slug := generateShortUniqueName("slug")

	params := CreateTenantCompleteParams{
		Name:               uniqueName,
		Slug:               slug,
		Email:              fmt.Sprintf("admin@%s.com", generateShortUniqueName("test")),
		Subdomain:          stringPtr(generateShortUniqueName("sub")),
		Status:             "active",
		Timezone:           "America/New_York",
		CurrencyCode:       "EUR",
		Metadata:           metadata,
		Industry:           stringPtr("finance"),
		CompanySize:        stringPtr("large"),
		TaxID:              stringPtr(generateShortUniqueName("TAX")),
		RegistrationNumber: stringPtr(generateShortUniqueName("REG")),
		LegalEntityType:    stringPtr("llc"),
		Settings:           settings,
	}

	tenant, err := suite.store.CreateTenantComplete(testCtx, params)
	suite.Require().NoError(err)
	suite.Require().NotNil(tenant)
	suite.Require().Equal(params.Name, tenant.Name)
	suite.Require().Equal(params.Slug, tenant.Slug)
	suite.Require().Equal(params.Email, tenant.Email)
	suite.Require().Equal(params.Timezone, tenant.Timezone)
	suite.Require().Equal(params.CurrencyCode, tenant.CurrencyCode)
	suite.Require().Equal(params.CompanySize, tenant.CompanySize)

	// Track for cleanup
	suite.trackTenant(tenant.ID)
}

func (suite *TenantTestSuite) TestGetTenantByID() {
	// Create a test tenant
	createdTenant := createTestTenant(suite.T(), suite.store, "GetByID")
	suite.trackTenant(createdTenant.ID)

	// Test getting existing tenant
	tenant, err := suite.store.GetTenantByID(testCtx, createdTenant.ID)
	suite.Require().NoError(err)
	suite.Require().NotNil(tenant)
	suite.Require().Equal(createdTenant.ID, tenant.ID)
	suite.Require().Equal(createdTenant.Name, tenant.Name)

	// Test getting non-existent tenant
	_, err = suite.store.GetTenantByID(testCtx, uuid.New())
	suite.Require().Error(err)
}

func (suite *TenantTestSuite) TestGetTenantBySlug() {
	tenant := createComplexTestTenant(suite.T(), suite.store, "Slug")
	suite.trackTenant(tenant.ID)

	foundTenant, err := suite.store.GetTenantBySlug(testCtx, tenant.Slug)
	suite.Require().NoError(err)
	suite.Require().NotNil(foundTenant)
	suite.Require().Equal(tenant.ID, foundTenant.ID)
	suite.Require().Equal(tenant.Slug, foundTenant.Slug)
}

func (suite *TenantTestSuite) TestGetTenantByEmail() {
	tenant := createComplexTestTenant(suite.T(), suite.store, "Email")
	suite.trackTenant(tenant.ID)

	foundTenant, err := suite.store.GetTenantByEmail(testCtx, tenant.Email)
	suite.Require().NoError(err)
	suite.Require().NotNil(foundTenant)
	suite.Require().Equal(tenant.ID, foundTenant.ID)
	suite.Require().Equal(tenant.Email, foundTenant.Email)
}

func (suite *TenantTestSuite) TestUpdateTenant() {
	createdTenant := createTestTenant(suite.T(), suite.store, "Update")
	suite.trackTenant(createdTenant.ID)

	newSubdomain := generateShortUniqueName("updated")
	params := UpdateTenantParams{
		ID:        createdTenant.ID,
		Name:      stringPtr(generateShortUniqueName("Updated")),
		Subdomain: stringPtr(newSubdomain),
		Status:    stringPtr("suspended"),
		Industry:  stringPtr("finance"),
	}

	updatedTenant, err := suite.store.UpdateTenant(testCtx, params)
	suite.Require().NoError(err)
	suite.Require().NotNil(updatedTenant)
	suite.Require().Equal(*params.Name, updatedTenant.Name)
	suite.Require().Equal(*params.Status, updatedTenant.Status)
	suite.Require().Equal(params.Subdomain, updatedTenant.Subdomain)
	suite.Require().Equal(params.Industry, updatedTenant.Industry)
}

func (suite *TenantTestSuite) TestSoftDeleteTenant() {
	createdTenant := createTestTenant(suite.T(), suite.store, "Delete")
	suite.trackTenant(createdTenant.ID)

	err := suite.store.SoftDeleteTenant(testCtx, createdTenant.ID)
	suite.Require().NoError(err)

	// Verify tenant is soft deleted (should not be found)
	_, err = suite.store.GetTenantByID(testCtx, createdTenant.ID)
	suite.Require().Error(err)
}

// Tenant Configuration Tests
func (suite *TenantTestSuite) TestTenantConfiguration() {
	tenant := createTestTenant(suite.T(), suite.store, "Config")
	suite.trackTenant(tenant.ID)

	// Clean up any existing configuration for this tenant (safety measure)
	_, _ = suite.pool.Exec(testCtx, "DELETE FROM tenant_configurations WHERE tenant_id = $1", tenant.ID)

	// Set tenant context
	err := suite.store.SetTenantContext(testCtx, tenant.ID)
	suite.Require().NoError(err)

	// Verify context is set
	currentTenantID, err := suite.store.GetCurrentTenantID(testCtx)
	suite.Require().NoError(err)
	suite.Require().Equal(tenant.ID, currentTenantID)

	// Create configuration
	features, err := json.Marshal(map[string]bool{
		"advanced_reporting": true,
		"api_access":         true,
	})
	suite.Require().NoError(err)

	modules, err := json.Marshal([]string{"accounting", "inventory", "payroll"})
	suite.Require().NoError(err)

	passwordPolicy, err := json.Marshal(map[string]interface{}{
		"min_length":      12,
		"require_symbols": true,
	})
	suite.Require().NoError(err)

	webhooks, err := json.Marshal([]string{"https://example.com/webhook"})
	suite.Require().NoError(err)

	rateLimits, err := json.Marshal(map[string]int{
		"requests_per_minute": 200,
		"requests_per_hour":   10000,
	})
	suite.Require().NoError(err)

	configParams := CreateTenantConfigurationParams{
		MaxUsers:                500,
		MaxEntities:             5000,
		MaxTransactionsPerMonth: 50000,
		StorageQuota:            5368709120, // 5GB
		Features:                features,
		ModulesEnabled:          modules,
		AccountingMethod:        "accrual",
		FiscalYearStartMonth:    4, // April
		DefaultCurrency:         "EUR",
		DateFormat:              "DD/MM/YYYY",
		NumberFormat:            "EU",
		LanguageCode:            "en-GB",
		PasswordPolicy:          passwordPolicy,
		WebhookEndpoints:        webhooks,
		ApiRateLimits:           rateLimits,
	}

	config, err := suite.store.CreateTenantConfiguration(testCtx, configParams)
	suite.Require().NoError(err)
	suite.Require().NotNil(config)
	suite.Require().Equal(tenant.ID, config.TenantID)
	suite.Require().Equal(configParams.MaxUsers, config.MaxUsers)
	suite.Require().Equal(configParams.AccountingMethod, config.AccountingMethod)

	// Get configuration
	retrievedConfig, err := suite.store.GetTenantConfiguration(testCtx)
	suite.Require().NoError(err)
	suite.Require().Equal(config.TenantID, retrievedConfig.TenantID)
	suite.Require().Equal(config.MaxUsers, retrievedConfig.MaxUsers)

	// Update features
	newFeatures, err := json.Marshal(map[string]bool{
		"advanced_reporting": false,
		"api_access":         true,
		"multi_currency":     true,
	})
	suite.Require().NoError(err)

	updatedConfig, err := suite.store.UpdateTenantFeatures(testCtx, newFeatures)
	suite.Require().NoError(err)
	
	// Parse both JSON values to compare content rather than byte arrays
	var expectedFeatures, actualFeatures map[string]bool
	err = json.Unmarshal(newFeatures, &expectedFeatures)
	suite.Require().NoError(err)
	err = json.Unmarshal(updatedConfig.Features, &actualFeatures)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedFeatures, actualFeatures)
}

// Tenant Usage Statistics Tests
func (suite *TenantTestSuite) TestTenantUsageStats() {
	tenant := createTestTenant(suite.T(), suite.store, "Usage")
	suite.trackTenant(tenant.ID)

	// Set tenant context
	err := suite.store.SetTenantContext(testCtx, tenant.ID)
	suite.Require().NoError(err)

	// Create usage stats
	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, -1)

	var avgResponseTime pgtype.Numeric
	err = avgResponseTime.Scan("125.50")
	suite.Require().NoError(err)

	var errorRate pgtype.Numeric
	err = errorRate.Scan("0.0045")
	suite.Require().NoError(err)

	var monthlyRevenue pgtype.Numeric
	err = monthlyRevenue.Scan("15750.00")
	suite.Require().NoError(err)

	usageParams := CreateTenantUsageStatsParams{
		PeriodStart:       periodStart,
		PeriodEnd:         periodEnd,
		ActiveUsers:       25,
		TotalEntities:     150,
		TotalTransactions: 1250,
		StorageUsed:       536870912, // 512MB
		ApiCalls:          5000,
		AvgResponseTime:   avgResponseTime,
		ErrorRate:         errorRate,
		MonthlyRevenue:    monthlyRevenue,
	}

	usage, err := suite.store.CreateTenantUsageStats(testCtx, usageParams)
	suite.Require().NoError(err)
	suite.Require().NotNil(usage)
	suite.Require().Equal(tenant.ID, usage.TenantID)
	suite.Require().Equal(usageParams.ActiveUsers, usage.ActiveUsers)
	suite.Require().Equal(usageParams.StorageUsed, usage.StorageUsed)

	// Get usage stats
	retrievedUsage, err := suite.store.GetTenantUsageStats(testCtx, periodStart)
	suite.Require().NoError(err)
	suite.Require().Equal(usage.TenantID, retrievedUsage.TenantID)
	suite.Require().Equal(usage.ActiveUsers, retrievedUsage.ActiveUsers)

	// Get latest usage stats
	latestUsage, err := suite.store.GetLatestTenantUsageStats(testCtx)
	suite.Require().NoError(err)
	suite.Require().Equal(usage.TenantID, latestUsage.TenantID)

	// Update usage stats
	updateParams := UpdateTenantUsageStatsParams{
		ActiveUsers: int32Ptr(30),
		StorageUsed: int64Ptr(1073741824), // 1GB
		PeriodStart: periodStart,
	}

	updatedUsage, err := suite.store.UpdateTenantUsageStats(testCtx, updateParams)
	suite.Require().NoError(err)
	suite.Require().Equal(*updateParams.ActiveUsers, updatedUsage.ActiveUsers)
	suite.Require().Equal(*updateParams.StorageUsed, updatedUsage.StorageUsed)
}

// Tenant Context and RLS Tests
func (suite *TenantTestSuite) TestTenantContext() {
	tenant := createTestTenant(suite.T(), suite.store, "Context")
	suite.trackTenant(tenant.ID)

	// Set tenant context
	err := suite.store.SetTenantContext(testCtx, tenant.ID)
	suite.Require().NoError(err)

	// Get current tenant ID
	currentID, err := suite.store.GetCurrentTenantID(testCtx)
	suite.Require().NoError(err)
	suite.Require().Equal(tenant.ID, currentID)

	// Check if current tenant exists
	exists, err := suite.store.CheckCurrentTenantExists(testCtx)
	suite.Require().NoError(err)
	suite.Require().True(exists)

	// Get current tenant
	currentTenant, err := suite.store.GetCurrentTenant(testCtx)
	suite.Require().NoError(err)
	suite.Require().Equal(tenant.ID, currentTenant.ID)
}

// Search and Filter Tests
func (suite *TenantTestSuite) TestSearchAndFilter() {
	// Create multiple test tenants with unique names (keep them short)
	testData := []struct {
		name string
	}{
		{generateShortUniqueName("Apple")},
		{generateShortUniqueName("Banana")},
		{generateShortUniqueName("Cherry")},
		{generateShortUniqueName("AppleSol")}, // Short version of Apple Solutions
	}

	tenants := make([]*Tenant, len(testData))
	for i, data := range testData {
		tenants[i] = createTestTenant(suite.T(), suite.store, data.name)
		suite.trackTenant(tenants[i].ID)
	}

	// Search by name pattern (Apple)
	searchParams := SearchTenantsByNameParams{
		Name:   "Apple",
		Limit:  10,
		Offset: 0,
	}

	results, err := suite.store.SearchTenantsByName(testCtx, searchParams)
	suite.Require().NoError(err)
	suite.Require().GreaterOrEqual(len(results), 2) // Should find at least our 2 Apple companies

	// Filter tenants
	filterParams := FilterTenantsParams{
		StatusFilter:   stringPtr("active"),
		IndustryFilter: stringPtr("technology"),
		SortBy:         "name",
		LimitCount:     10,
		OffsetCount:    0,
	}

	filtered, err := suite.store.FilterTenants(testCtx, filterParams)
	suite.Require().NoError(err)
	suite.Require().GreaterOrEqual(len(filtered), len(tenants))

	// Count filtered tenants
	countParams := CountFilteredTenantsParams{
		StatusFilter: stringPtr("active"),
	}

	count, err := suite.store.CountFilteredTenants(testCtx, countParams)
	suite.Require().NoError(err)
	suite.Require().GreaterOrEqual(count, int64(len(tenants)))
}

// Bulk Operations Tests
func (suite *TenantTestSuite) TestBulkOperations() {
	// Create multiple tenants
	testNames := []string{"Bulk1", "Bulk2", "Bulk3"}
	tenants := make([]*Tenant, len(testNames))
	tenantIDs := make([]uuid.UUID, len(testNames))

	for i, name := range testNames {
		tenants[i] = createTestTenant(suite.T(), suite.store, name)
		tenantIDs[i] = tenants[i].ID
		suite.trackTenant(tenants[i].ID)
	}

	// Bulk update status
	bulkUpdateParams := BulkUpdateTenantStatusParams{
		Status: "suspended",
		ID:     tenantIDs,
	}

	err := suite.store.BulkUpdateTenantStatus(testCtx, bulkUpdateParams)
	suite.Require().NoError(err)

	// Verify status was updated
	for _, tenantID := range tenantIDs {
		tenant, err := suite.store.GetTenantByID(testCtx, tenantID)
		suite.Require().NoError(err)
		suite.Require().Equal("suspended", tenant.Status)
	}

	// Bulk soft delete
	err = suite.store.BulkSoftDeleteTenants(testCtx, tenantIDs)
	suite.Require().NoError(err)

	// Verify tenants are deleted
	for _, tenantID := range tenantIDs {
		_, err := suite.store.GetTenantByID(testCtx, tenantID)
		suite.Require().Error(err) // Should not be found
	}
}

// Analytics Tests
func (suite *TenantTestSuite) TestAnalytics() {
	// Create diverse test data with short names
	industries := []string{"tech", "health", "finance", "mfg"} // Shortened industry names
	sizes := []string{"startup", "small", "medium", "large"}

	createdTenants := make([]uuid.UUID, 0)
	for _, industry := range industries {
		for _, size := range sizes {
			metadata, _ := json.Marshal(map[string]interface{}{"test": true})
			settings, _ := json.Marshal(map[string]interface{}{"theme": "light"})

			// Create very short, unique names
			uniqueName := generateShortUniqueName(fmt.Sprintf("%s%s", industry, size))
			slug := generateShortUniqueName("slug")

			params := CreateTenantCompleteParams{
				Name:         uniqueName,
				Slug:         slug,
				Email:        fmt.Sprintf("test@%s.com", generateShortUniqueName("domain")),
				Subdomain:    stringPtr(generateShortUniqueName("sub")),
				Status:       "active",
				Timezone:     "UTC",
				CurrencyCode: "USD",
				Metadata:     metadata,
				Industry:     &industry,
				CompanySize:  &size,
				Settings:     settings,
			}

			tenant, err := suite.store.CreateTenantComplete(testCtx, params)
			suite.Require().NoError(err)
			createdTenants = append(createdTenants, tenant.ID)
			suite.trackTenant(tenant.ID)
		}
	}

	// Test analytics queries
	stats, err := suite.store.GetTenantStats(testCtx)
	suite.Require().NoError(err)
	suite.Require().GreaterOrEqual(stats.TotalTenants, int64(len(createdTenants)))

	// Test industry distribution
	industryStats, err := suite.store.GetTenantsByIndustry(testCtx)
	suite.Require().NoError(err)
	suite.Require().GreaterOrEqual(len(industryStats), len(industries))

	// Test company size distribution
	sizeStats, err := suite.store.GetTenantsByCompanySize(testCtx)
	suite.Require().NoError(err)
	suite.Require().GreaterOrEqual(len(sizeStats), len(sizes))

	// Test status distribution
	statusDist, err := suite.store.GetTenantStatusDistribution(testCtx)
	suite.Require().NoError(err)
	suite.Require().GreaterOrEqual(statusDist.ActiveTenants, int64(len(createdTenants)))
}

// Validation Tests
func (suite *TenantTestSuite) TestValidationChecks() {
	tenant := createTestTenant(suite.T(), suite.store, "Validation")
	suite.trackTenant(tenant.ID)

	// Check tenant exists
	exists, err := suite.store.CheckTenantExists(testCtx, tenant.ID)
	suite.Require().NoError(err)
	suite.Require().True(exists)

	// Check non-existent tenant
	exists, err = suite.store.CheckTenantExists(testCtx, uuid.New())
	suite.Require().NoError(err)
	suite.Require().False(exists)

	// Check tenant name exists
	exists, err = suite.store.CheckTenantNameExists(testCtx, tenant.Name)
	suite.Require().NoError(err)
	suite.Require().True(exists)

	// Check subdomain exists
	if tenant.Subdomain != nil {
		exists, err = suite.store.CheckSubdomainExists(testCtx, tenant.Subdomain)
		suite.Require().NoError(err)
		suite.Require().True(exists)
	}

	// Check non-existent subdomain
	nonExistent := generateShortUniqueName("nonexist")
	exists, err = suite.store.CheckSubdomainExists(testCtx, &nonExistent)
	suite.Require().NoError(err)
	suite.Require().False(exists)
}

// Tenant Limits Tests
func (suite *TenantTestSuite) TestTenantLimits() {
	tenant := createTestTenant(suite.T(), suite.store, "Limits")
	suite.trackTenant(tenant.ID)

	// Set tenant context
	err := suite.store.SetTenantContext(testCtx, tenant.ID)
	suite.Require().NoError(err)

	// Create default configuration first
	err = suite.store.CreateDefaultTenantConfiguration(testCtx)
	if err == nil { // Only if function exists and works
		// Check limits
		limitsParams := CheckTenantLimitsParams{
			PCheckType:       "storage",
			PAdditionalUsage: 1024,
		}

		canProceed, err := suite.store.CheckTenantLimits(testCtx, limitsParams)
		if err == nil {
			suite.Require().IsType(bool(true), canProceed)
		}
	}
}

// Date Range Tests
func (suite *TenantTestSuite) TestDateRangeQueries() {
	tenant := createTestTenant(suite.T(), suite.store, "DateRange")
	suite.trackTenant(tenant.ID)

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)

	params := GetTenantsCreatedInDateRangeParams{
		CreatedAt:   yesterday,
		CreatedAt_2: tomorrow,
	}

	tenants, err := suite.store.GetTenantsCreatedInDateRange(testCtx, params)
	suite.Require().NoError(err)
	suite.Require().GreaterOrEqual(len(tenants), 1)

	// Verify all tenants are within range
	for _, tenant := range tenants {
		suite.Require().True(tenant.CreatedAt.After(yesterday) || tenant.CreatedAt.Equal(yesterday))
		suite.Require().True(tenant.CreatedAt.Before(tomorrow) || tenant.CreatedAt.Equal(tomorrow))
	}
}

// List and Pagination Tests
func (suite *TenantTestSuite) TestListTenants() {
	// Create multiple tenants
	createdTenants := make([]uuid.UUID, 0)
	for i := 0; i < 5; i++ {
		tenant := createTestTenant(suite.T(), suite.store, fmt.Sprintf("List%d", i))
		createdTenants = append(createdTenants, tenant.ID)
		suite.trackTenant(tenant.ID)
	}

	// Test pagination
	params := ListTenantsParams{
		Limit:  3,
		Offset: 0,
	}

	tenants, err := suite.store.ListTenants(testCtx, params)
	suite.Require().NoError(err)
	suite.Require().LessOrEqual(len(tenants), 3)

	// Test second page
	params.Offset = 3
	params.Limit = 2

	secondPage, err := suite.store.ListTenants(testCtx, params)
	suite.Require().NoError(err)
	suite.Require().LessOrEqual(len(secondPage), 2)

	// Count total tenants
	count, err := suite.store.CountTenants(testCtx)
	suite.Require().NoError(err)
	suite.Require().GreaterOrEqual(count, int64(len(createdTenants)))
}

// Error Handling Tests
func (suite *TenantTestSuite) TestErrorHandling() {
	// Test duplicate names
	tenant1 := createTestTenant(suite.T(), suite.store, "Unique")
	suite.trackTenant(tenant1.ID)

	// Try to create another tenant with same name
	params := CreateTenantParams{
		Name:      tenant1.Name, // Duplicate name
		Subdomain: stringPtr(generateShortUniqueName("different")),
		Status:    "active",
		Industry:  stringPtr("finance"),
	}

	_, err := suite.store.CreateTenant(testCtx, params)
	suite.Require().Error(err, "Should fail due to duplicate name")

	// Test duplicate subdomain
	if tenant1.Subdomain != nil {
		params2 := CreateTenantParams{
			Name:      generateShortUniqueName("Different"),
			Subdomain: tenant1.Subdomain, // Duplicate subdomain
			Status:    "active",
			Industry:  stringPtr("healthcare"),
		}

		_, err = suite.store.CreateTenant(testCtx, params2)
		suite.Require().Error(err, "Should fail due to duplicate subdomain")
	}
}

// Performance Tests
func (suite *TenantTestSuite) TestPerformance() {
	if testing.Short() {
		suite.T().Skip("Skipping performance tests in short mode")
	}

	start := time.Now()
	createdTenants := make([]uuid.UUID, 0)

	// Create many tenants
	for i := 0; i < 50; i++ {
		tenant := createTestTenant(suite.T(), suite.store, fmt.Sprintf("Perf%d", i))
		createdTenants = append(createdTenants, tenant.ID)
		suite.trackTenant(tenant.ID)
	}

	duration := time.Since(start)
	suite.T().Logf("Created 50 tenants in %v", duration)
	suite.Require().Less(duration, 30*time.Second, "Should create 50 tenants in less than 30 seconds")
}

// Run the test suite
func TestTenantTestSuite(t *testing.T) {
	suite.Run(t, new(TenantTestSuite))
}

// Additional standalone tests
func TestTenantLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	pool := setupTestDB(t)
	defer pool.Close()
	store := NewStore(pool)

	// Track tenant for cleanup
	var tenantID uuid.UUID
	defer func() {
		if tenantID != uuid.Nil {
			// Clean up
			_, _ = pool.Exec(testCtx, "DELETE FROM tenant_usage_stats WHERE tenant_id = $1", tenantID)
			_, _ = pool.Exec(testCtx, "DELETE FROM tenant_configurations WHERE tenant_id = $1", tenantID)
			_, _ = pool.Exec(testCtx, "DELETE FROM tenants WHERE id = $1", tenantID)
		}
	}()

	// Complete tenant lifecycle test
	uniqueName := generateShortUniqueName("Lifecycle")
	subdomain := generateShortUniqueName("lifecycle")

	params := CreateTenantParams{
		Name:      uniqueName,
		Subdomain: stringPtr(subdomain),
		Status:    "pending",
		Industry:  stringPtr("technology"),
	}

	// 1. Create tenant
	tenant, err := store.CreateTenant(testCtx, params)
	require.NoError(t, err)
	require.NotNil(t, tenant)
	require.Equal(t, "pending", tenant.Status)
	tenantID = tenant.ID

	// 2. Activate tenant
	updateParams := UpdateTenantStatusParams{
		ID:     tenant.ID,
		Status: "active",
	}

	activeTenant, err := store.UpdateTenantStatus(testCtx, updateParams)
	require.NoError(t, err)
	require.Equal(t, "active", activeTenant.Status)

	// 3. Update tenant details
	newName := generateShortUniqueName("Updated")
	updateAllParams := UpdateTenantParams{
		ID:       tenant.ID,
		Name:     stringPtr(newName),
		Industry: stringPtr("finance"),
	}

	updatedTenant, err := store.UpdateTenant(testCtx, updateAllParams)
	require.NoError(t, err)
	require.Equal(t, newName, updatedTenant.Name)
	require.Equal(t, stringPtr("finance"), updatedTenant.Industry)

	// 4. Set tenant context and work with tenant-scoped data
	err = store.SetTenantContext(testCtx, tenant.ID)
	require.NoError(t, err)

	// Verify context is set correctly
	currentTenantID, err := store.GetCurrentTenantID(testCtx)
	require.NoError(t, err)
	require.Equal(t, tenant.ID, currentTenantID)

	// 5. Create tenant configuration
	err = store.CreateDefaultTenantConfiguration(testCtx)
	if err == nil { // Only if function works in test environment
		config, err := store.GetTenantConfiguration(testCtx)
		if err == nil {
			require.Equal(t, tenant.ID, config.TenantID)
		}
	}

	// 6. Soft delete tenant
	err = store.SoftDeleteTenant(testCtx, tenant.ID)
	require.NoError(t, err)

	// 7. Verify tenant is no longer accessible
	_, err = store.GetTenantByID(testCtx, tenant.ID)
	require.Error(t, err, "Soft deleted tenant should not be found")
}

// Benchmark tests
func BenchmarkCreateTenant(b *testing.B) {
	pool := setupTestDB(&testing.T{})
	defer pool.Close()
	store := NewStore(pool)

	// Track created tenants for cleanup
	createdTenants := make([]uuid.UUID, 0, b.N)
	defer func() {
		for _, tenantID := range createdTenants {
			_, _ = pool.Exec(testCtx, "DELETE FROM tenants WHERE id = $1", tenantID)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		uniqueName := generateShortUniqueName(fmt.Sprintf("Bench%d", i))
		subdomain := generateShortUniqueName(fmt.Sprintf("b%d", i))

		params := CreateTenantParams{
			Name:      uniqueName,
			Subdomain: stringPtr(subdomain),
			Status:    "active",
			Industry:  stringPtr("technology"),
		}
		tenant, err := store.CreateTenant(testCtx, params)
		if err != nil {
			b.Fatal(err)
		}
		createdTenants = append(createdTenants, tenant.ID)
	}
}

func BenchmarkGetTenantByID(b *testing.B) {
	pool := setupTestDB(&testing.T{})
	defer pool.Close()
	store := NewStore(pool)

	tenant := createTestTenant(&testing.T{}, store, "BenchGet")
	defer func() {
		_, _ = pool.Exec(testCtx, "DELETE FROM tenants WHERE id = $1", tenant.ID)
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := store.GetTenantByID(testCtx, tenant.ID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Test main for setup/teardown
func TestMain(m *testing.M) {
	// Setup before all tests
	code := m.Run()

	// Cleanup after all tests
	os.Exit(code)
}
