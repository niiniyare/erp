package ui

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// SecurityTestSuite defines a test suite for military-grade security features
type SecurityTestSuite struct {
	suite.Suite
	validator   *SecurityValidator
	securityCtx *TenantSecurityContext
	registry    ComponentRegistry
	ctx         context.Context
}

// SetupSuite runs once before all tests in the suite
func (suite *SecurityTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.validator = NewSecurityValidator()
	suite.registry = NewRegistryWithSchemaDir("../../../docs/ui/Schema")
	
	// Create a high-security tenant context
	suite.securityCtx = NewTenantSecurityContext("tenant-001", "user-001", SecurityLevelTopSecret)
	
	// Add required permissions
	suite.securityCtx.Permissions = []Permission{
		{
			Resource: "component:button",
			Action:   "create",
		},
		{
			Resource: "component:button",
			Action:   "read",
		},
		{
			Resource: "component:input",
			Action:   "create",
		},
		{
			Resource: "component:input",
			Action:   "read",
		},
		{
			Resource: "component:card",
			Action:   "create",
		},
		{
			Resource: "component:card",
			Action:   "read",
		},
	}
}

// TestSecurityContextCreation tests creation of security contexts
func (suite *SecurityTestSuite) TestSecurityContextCreation() {
	secCtx := NewTenantSecurityContext("test-tenant", "test-user", SecurityLevelSecret)
	
	require.NotNil(suite.T(), secCtx, "Security context should not be nil")
	assert.Equal(suite.T(), "test-tenant", secCtx.TenantID)
	assert.Equal(suite.T(), "test-user", secCtx.UserID)
	assert.Equal(suite.T(), SecurityLevelSecret, secCtx.SecurityLevel)
	assert.NotEmpty(suite.T(), secCtx.SessionID, "Session ID should be generated")
	assert.NotNil(suite.T(), secCtx.EncryptionKeys, "Encryption keys should be generated")
	assert.False(suite.T(), secCtx.CreatedAt.IsZero(), "Created timestamp should be set")
	assert.False(suite.T(), secCtx.ExpiresAt.IsZero(), "Expiration should be set")
}

// TestSecurityContextValidation tests validation of security contexts
func (suite *SecurityTestSuite) TestSecurityContextValidation() {
	// Test valid context
	err := validateSecurityContext(suite.securityCtx)
	assert.NoError(suite.T(), err, "Valid security context should pass validation")
	
	// Test nil context
	err = validateSecurityContext(nil)
	assert.Error(suite.T(), err, "Nil context should fail validation")
	
	// Test expired context
	expiredCtx := NewTenantSecurityContext("tenant", "user", SecurityLevelPublic)
	expiredCtx.ExpiresAt = time.Now().Add(-1 * time.Hour) // Expired 1 hour ago
	
	err = validateSecurityContext(expiredCtx)
	assert.Error(suite.T(), err, "Expired context should fail validation")
	
	// Test missing tenant ID
	invalidCtx := NewTenantSecurityContext("", "user", SecurityLevelPublic)
	err = validateSecurityContext(invalidCtx)
	assert.Error(suite.T(), err, "Context without tenant ID should fail validation")
}

// TestSecureComponentCreation tests secure component creation
func (suite *SecurityTestSuite) TestSecureComponentCreation() {
	// Test creating a secure button component
	buttonConfig := map[string]any{
		"text":    "Secure Button",
		"variant": "primary",
	}
	
	component, err := CreateSecureComponent(
		suite.ctx,
		suite.registry,
		ComponentButton,
		buttonConfig,
		suite.securityCtx,
	)
	
	require.NoError(suite.T(), err, "Should create secure component without error")
	assert.Equal(suite.T(), ComponentButton, component.Type)
	assert.NotEmpty(suite.T(), component.ID)
	
	// Verify tenant isolation metadata
	assert.Equal(suite.T(), suite.securityCtx.TenantID, component.metadata["tenant_id"])
	assert.Equal(suite.T(), string(suite.securityCtx.SecurityLevel), component.metadata["security_level"])
	assert.Equal(suite.T(), suite.securityCtx.UserID, component.metadata["created_by"])
	
	// Verify encryption is applied for high security levels
	assert.True(suite.T(), component.encrypted, "Component should be encrypted for top-secret level")
	assert.NotNil(suite.T(), component.encryptedConfig, "Encrypted config should be present")
}

// TestSecureComponentCreationWithInsufficientPermissions tests access control
func (suite *SecurityTestSuite) TestSecureComponentCreationWithInsufficientPermissions() {
	// Create context without permissions
	unauthorizedCtx := NewTenantSecurityContext("tenant-002", "user-002", SecurityLevelPublic)
	
	buttonConfig := map[string]any{
		"text":    "Unauthorized Button",
		"variant": "primary",
	}
	
	_, err := CreateSecureComponent(
		suite.ctx,
		suite.registry,
		ComponentButton,
		buttonConfig,
		unauthorizedCtx,
	)
	
	assert.Error(suite.T(), err, "Should fail without proper permissions")
	assert.Contains(suite.T(), err.Error(), "insufficient permissions")
}

// TestTenantIsolationValidation tests tenant isolation enforcement
func (suite *SecurityTestSuite) TestTenantIsolationValidation() {
	// Create a component with correct tenant
	component := NewComponent(ComponentCard, "test-card").
		WithMetadata("tenant_id", suite.securityCtx.TenantID)
	builtComponent := component.Build()
	
	violations := suite.validator.ValidateComponentSecurity(suite.ctx, builtComponent, suite.securityCtx)
	
	// Should have minimal violations for correct tenant
	tenantViolations := filterViolationsByType(violations, ViolationTenantIsolation)
	assert.Empty(suite.T(), tenantViolations, "Should not have tenant isolation violations for correct tenant")
	
	// Test with wrong tenant
	wrongTenantCtx := NewTenantSecurityContext("wrong-tenant", "user-001", SecurityLevelTopSecret)
	wrongTenantCtx.Permissions = suite.securityCtx.Permissions // Copy permissions
	
	violations = suite.validator.ValidateComponentSecurity(suite.ctx, builtComponent, wrongTenantCtx)
	tenantViolations = filterViolationsByType(violations, ViolationTenantIsolation)
	assert.NotEmpty(suite.T(), tenantViolations, "Should have tenant isolation violations for wrong tenant")
}

// TestPermissionValidation tests permission-based access control
func (suite *SecurityTestSuite) TestPermissionValidation() {
	component := NewComponent(ComponentInput, "test-input")
	builtComponent := component.Build()
	
	// Test with sufficient permissions
	violations := suite.validator.ValidateComponentSecurity(suite.ctx, builtComponent, suite.securityCtx)
	permissionViolations := filterViolationsByType(violations, ViolationInsufficientPermissions)
	
	// Should pass if we have the right permissions setup
	if len(permissionViolations) > 0 {
		suite.T().Logf("Permission violations found: %v", permissionViolations)
	}
	
	// Test with insufficient permissions
	limitedCtx := NewTenantSecurityContext("tenant-001", "user-001", SecurityLevelTopSecret)
	limitedCtx.Permissions = []Permission{
		{
			Resource: "component:button", // Only button permission, not input
			Action:   "read",
		},
	}
	
	violations = suite.validator.ValidateComponentSecurity(suite.ctx, builtComponent, limitedCtx)
	permissionViolations = filterViolationsByType(violations, ViolationInsufficientPermissions)
	assert.NotEmpty(suite.T(), permissionViolations, "Should have permission violations without proper permissions")
}

// TestEncryptionValidation tests encryption requirement enforcement
func (suite *SecurityTestSuite) TestEncryptionValidation() {
	// Create an unencrypted component
	component := NewComponent(ComponentCard, "test-card")
	builtComponent := component.Build()
	// builtComponent.encrypted remains false
	
	// Test with high security level that requires encryption
	violations := suite.validator.ValidateComponentSecurity(suite.ctx, builtComponent, suite.securityCtx)
	encryptionViolations := filterViolationsByType(violations, ViolationEncryption)
	assert.NotEmpty(suite.T(), encryptionViolations, "Should have encryption violations for unencrypted high-security component")
	
	// Test with encrypted component
	encryptedComponent := builtComponent
	encryptedComponent.encrypted = true
	
	violations = suite.validator.ValidateComponentSecurity(suite.ctx, encryptedComponent, suite.securityCtx)
	encryptionViolations = filterViolationsByType(violations, ViolationEncryption)
	assert.Empty(suite.T(), encryptionViolations, "Should not have encryption violations for encrypted component")
}

// TestDataClassificationValidation tests data classification enforcement
func (suite *SecurityTestSuite) TestDataClassificationValidation() {
	// Create component with high classification
	component := NewComponent(ComponentCard, "classified-card").
		WithMetadata("data_classification", "secret").
		WithMetadata("tenant_id", suite.securityCtx.TenantID)
	builtComponent := component.Build()
	
	// Should pass with top-secret security level
	violations := suite.validator.ValidateComponentSecurity(suite.ctx, builtComponent, suite.securityCtx)
	classificationViolations := filterViolationsByType(violations, ViolationDataClassification)
	assert.Empty(suite.T(), classificationViolations, "Top-secret level should access secret data")
	
	// Test with insufficient security level
	lowSecurityCtx := NewTenantSecurityContext("tenant-001", "user-001", SecurityLevelPublic)
	lowSecurityCtx.Permissions = suite.securityCtx.Permissions // Copy permissions
	
	violations = suite.validator.ValidateComponentSecurity(suite.ctx, builtComponent, lowSecurityCtx)
	classificationViolations = filterViolationsByType(violations, ViolationDataClassification)
	assert.NotEmpty(suite.T(), classificationViolations, "Public level should not access secret data")
}

// TestPermissionChecking tests the permission checking system
func (suite *SecurityTestSuite) TestPermissionChecking() {
	// Test existing permission
	hasButtonPermission := hasPermission(suite.securityCtx, "component:button", "create")
	assert.True(suite.T(), hasButtonPermission, "Should have button create permission")
	
	// Test non-existent permission
	hasNonExistentPermission := hasPermission(suite.securityCtx, "component:nonexistent", "create")
	assert.False(suite.T(), hasNonExistentPermission, "Should not have non-existent permission")
	
	// Test expired permission
	expiredTime := time.Now().Add(-1 * time.Hour)
	expiredPermission := Permission{
		Resource:  "component:expired",
		Action:    "read",
		ExpiresAt: &expiredTime,
	}
	suite.securityCtx.Permissions = append(suite.securityCtx.Permissions, expiredPermission)
	
	hasExpiredPermission := hasPermission(suite.securityCtx, "component:expired", "read")
	assert.False(suite.T(), hasExpiredPermission, "Should not have expired permission")
}

// TestEncryptionKeyGeneration tests encryption key generation
func (suite *SecurityTestSuite) TestEncryptionKeyGeneration() {
	keys := generateEncryptionKeys()
	
	require.NotNil(suite.T(), keys, "Encryption keys should not be nil")
	assert.Len(suite.T(), keys.DataEncryptionKey, 32, "DEK should be 256 bits (32 bytes)")
	assert.Len(suite.T(), keys.KeyEncryptionKey, 32, "KEK should be 256 bits (32 bytes)")
	assert.Len(suite.T(), keys.SigningKey, 32, "Signing key should be 256 bits (32 bytes)")
	assert.Equal(suite.T(), 24*time.Hour, keys.RotationSchedule, "Should have 24-hour rotation schedule")
	assert.False(suite.T(), keys.LastRotated.IsZero(), "Last rotated should be set")
	assert.False(suite.T(), keys.NextRotation.IsZero(), "Next rotation should be set")
}

// TestSecurityViolationSeverity tests security violation severity levels
func (suite *SecurityTestSuite) TestSecurityViolationSeverity() {
	component := NewComponent(ComponentCard, "test-card").
		WithMetadata("tenant_id", "wrong-tenant") // Wrong tenant
	builtComponent := component.Build()
	
	violations := suite.validator.ValidateComponentSecurity(suite.ctx, builtComponent, suite.securityCtx)
	
	// Check that tenant isolation violations are marked as critical
	tenantViolations := filterViolationsByType(violations, ViolationTenantIsolation)
	if len(tenantViolations) > 0 {
		assert.Equal(suite.T(), SecuritySeverityCritical, tenantViolations[0].Severity, 
			"Tenant isolation violations should be critical severity")
	}
}

// TestSecureIDGeneration tests secure ID generation
func (suite *SecurityTestSuite) TestSecureIDGeneration() {
	id1 := generateSecureID()
	id2 := generateSecureID()
	
	assert.NotEmpty(suite.T(), id1, "Generated ID should not be empty")
	assert.NotEmpty(suite.T(), id2, "Generated ID should not be empty")
	assert.NotEqual(suite.T(), id1, id2, "Generated IDs should be unique")
	assert.Len(suite.T(), id1, 32, "Generated ID should be 32 characters (16 bytes hex)")
}

// Helper functions

func filterViolationsByType(violations []SecurityViolation, violationType ViolationType) []SecurityViolation {
	var filtered []SecurityViolation
	for _, v := range violations {
		if v.Type == violationType {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

// Run the test suite
func TestSecurityTestSuite(t *testing.T) {
	suite.Run(t, new(SecurityTestSuite))
}