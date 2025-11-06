package enrich

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/niiniyare/erp/pkg/schema"
	"github.com/stretchr/testify/suite"
)

type EnricherTestSuite struct {
	suite.Suite
	enricher *DefaultEnricher
	ctx      context.Context
}

func (suite *EnricherTestSuite) SetupTest() {
	suite.enricher = NewEnricher()
	suite.ctx = context.Background()
}

func TestEnricherTestSuite(t *testing.T) {
	suite.Run(t, new(EnricherTestSuite))
}

// Mock implementations for testing
type MockTenantProvider struct {
	customizations map[string]*TenantCustomization
}

func NewMockTenantProvider() *MockTenantProvider {
	return &MockTenantProvider{
		customizations: make(map[string]*TenantCustomization),
	}
}

func (m *MockTenantProvider) GetCustomization(ctx context.Context, schemaID, tenantID string) (*TenantCustomization, error) {
	key := schemaID + ":" + tenantID
	customization, exists := m.customizations[key]
	if !exists {
		return nil, nil // No customization found
	}
	return customization, nil
}

func (m *MockTenantProvider) SetCustomization(schemaID, tenantID string, customization *TenantCustomization) {
	key := schemaID + ":" + tenantID
	m.customizations[key] = customization
}

// Test basic enricher creation
func (suite *EnricherTestSuite) TestNewEnricher() {
	enricher := NewEnricher()
	suite.Require().NotNil(enricher)
	suite.Require().Nil(enricher.tenantProvider)
}

// Test setting tenant provider
func (suite *EnricherTestSuite) TestSetTenantProvider() {
	provider := NewMockTenantProvider()
	suite.enricher.SetTenantProvider(provider)
	suite.Require().Equal(provider, suite.enricher.tenantProvider)
}

// Test basic schema enrichment
func (suite *EnricherTestSuite) TestEnrichBasic() {
	user := &DefaultUser{
		ID:          "user123",
		TenantID:    "tenant456",
		Permissions: []string{"read", "write"},
		Roles:       []string{"user"},
	}

	testSchemaTemplate := &schema.Schema{
		ID:    "test-schema",
		Type:  schema.TypeForm,
		Title: "Test Schema",
		Fields: []schema.Field{
			{
				Name:     "username",
				Type:     schema.FieldText,
				Label:    "Username",
				Required: true,
			},
			{
				Name:  "email",
				Type:  schema.FieldEmail,
				Label: "Email",
			},
		},
	}

	// Clone the schema to test enrichment
	testSchema, err := testSchemaTemplate.Clone()
	suite.Require().NoError(err)

	err = suite.enricher.Enrich(suite.ctx, testSchema, WithUser(user))
	suite.Require().NoError(err)

	// Check field enrichment
	for _, field := range testSchema.Fields {
		suite.Require().NotNil(field.Runtime)
		suite.Require().True(field.Runtime.Visible)
		suite.Require().True(field.Runtime.Editable)
	}
}

// Test permission-based field enrichment
func (suite *EnricherTestSuite) TestEnrichWithPermissions() {
	user := &DefaultUser{
		ID:          "user123",
		TenantID:    "tenant456",
		Permissions: []string{"read"},
		Roles:       []string{"user"},
	}

	testSchemaTemplate := &schema.Schema{
		ID:    "test-schema",
		Type:  schema.TypeForm,
		Title: "Test Schema",
		Fields: []schema.Field{
			{
				Name:              "public_field",
				Type:              schema.FieldText,
				Label:             "Public Field",
				RequirePermission: "",
			},
			{
				Name:              "admin_field",
				Type:              schema.FieldText,
				Label:             "Admin Field",
				RequirePermission: "admin",
			},
			{
				Name:         "role_field",
				Type:         schema.FieldText,
				Label:        "Role Field",
				RequireRoles: []string{"admin", "manager"},
			},
		},
	}

	// Clone the schema to test enrichment using JSON marshaling to rebuild fieldMap
	schemaBytes, err := json.Marshal(testSchemaTemplate)
	suite.Require().NoError(err)
	var testSchema schema.Schema
	err = json.Unmarshal(schemaBytes, &testSchema)
	suite.Require().NoError(err)

	err = suite.enricher.Enrich(suite.ctx, &testSchema, WithUser(user))
	suite.Require().NoError(err)

	// Public field should be visible
	publicField, exists := testSchema.GetFieldPtr("public_field")
	suite.Require().True(exists)
	suite.Require().True(publicField.Runtime.Visible)
	suite.Require().True(publicField.Runtime.Editable)

	// Admin field should not be visible (user doesn't have admin permission)
	adminField, exists := testSchema.GetFieldPtr("admin_field")
	suite.Require().True(exists)
	suite.Require().False(adminField.Runtime.Visible)
	suite.Require().False(adminField.Runtime.Editable)
	suite.Require().Equal("permission_required", adminField.Runtime.Reason)

	// Role field should not be visible (user doesn't have required roles)
	roleField, exists := testSchema.GetFieldPtr("role_field")
	suite.Require().True(exists)
	suite.Require().False(roleField.Runtime.Visible)
	suite.Require().False(roleField.Runtime.Editable)
	suite.Require().Equal("role_required", roleField.Runtime.Reason)
}

// Test dynamic default values
func (suite *EnricherTestSuite) TestEnrichDynamicDefaults() {
	user := &DefaultUser{
		ID:       "user123",
		TenantID: "tenant456",
	}

	testSchemaTemplate := &schema.Schema{
		ID:    "test-schema",
		Type:  schema.TypeForm,
		Title: "Test Schema",
		Fields: []schema.Field{
			{
				Name:  "created_by",
				Type:  schema.FieldText,
				Label: "Created By",
			},
			{
				Name:  "tenant_id",
				Type:  schema.FieldText,
				Label: "Tenant ID",
			},
			{
				Name:  "created_at",
				Type:  schema.FieldDateTime,
				Label: "Created At",
			},
			{
				Name:    "static_default",
				Type:    schema.FieldText,
				Label:   "Static Default",
				Default: "static_value",
			},
			{
				Name:  "existing_value",
				Type:  schema.FieldText,
				Label: "Existing Value",
				Value: "existing",
			},
		},
	}

	// Clone the schema to test enrichment by using JSON marshaling (which properly rebuilds fieldMap)
	schemaBytes, err := json.Marshal(testSchemaTemplate)
	suite.Require().NoError(err)
	var testSchema schema.Schema
	err = json.Unmarshal(schemaBytes, &testSchema)
	suite.Require().NoError(err)

	err = suite.enricher.Enrich(suite.ctx, &testSchema, WithUser(user))
	suite.Require().NoError(err)

	// Check dynamic defaults
	createdByField, exists := testSchema.GetFieldPtr("created_by")
	suite.Require().True(exists)
	suite.Require().NotNil(createdByField)
	suite.Require().Equal("user123", createdByField.Value)

	tenantField, _ := testSchema.GetFieldPtr("tenant_id")
	suite.Require().Equal("tenant456", tenantField.Value)
	suite.Require().True(tenantField.Hidden)
	suite.Require().True(tenantField.Readonly)

	createdAtField, _ := testSchema.GetFieldPtr("created_at")
	suite.Require().NotNil(createdAtField.Value)

	// Static default should be applied
	staticField, _ := testSchema.GetFieldPtr("static_default")
	suite.Require().Equal("static_value", staticField.Value)

	// Existing value should not be overridden
	existingField, _ := testSchema.GetFieldPtr("existing_value")
	suite.Require().Equal("existing", existingField.Value)
}

// Test tenant customization
func (suite *EnricherTestSuite) TestEnrichWithTenantCustomization() {
	provider := NewMockTenantProvider()
	suite.enricher.SetTenantProvider(provider)

	user := &DefaultUser{
		ID:       "user123",
		TenantID: "tenant456",
	}

	// Set up tenant customization
	customization := &TenantCustomization{
		SchemaID: "test-schema",
		TenantID: "tenant456",
		Overrides: map[string]TenantOverride{
			"username": {
				FieldName:   "username",
				Label:       "Custom Username Label",
				Required:    boolPtr(false),
				Placeholder: "Enter custom username",
			},
			"email": {
				FieldName:    "email",
				Hidden:       boolPtr(true),
				DefaultValue: "default@tenant.com",
			},
		},
	}
	provider.SetCustomization("test-schema", "tenant456", customization)

	testSchemaTemplate := &schema.Schema{
		ID:    "test-schema",
		Type:  schema.TypeForm,
		Title: "Test Schema",
		Fields: []schema.Field{
			{
				Name:        "username",
				Type:        schema.FieldText,
				Label:       "Original Username",
				Required:    true,
				Placeholder: "Original placeholder",
			},
			{
				Name:  "email",
				Type:  schema.FieldEmail,
				Label: "Email",
			},
		},
	}

	// Clone the schema to test enrichment using JSON marshaling to rebuild fieldMap
	schemaBytes, err := json.Marshal(testSchemaTemplate)
	suite.Require().NoError(err)
	var testSchema schema.Schema
	err = json.Unmarshal(schemaBytes, &testSchema)
	suite.Require().NoError(err)

	err = suite.enricher.Enrich(suite.ctx, &testSchema, WithUser(user))
	suite.Require().NoError(err)

	// Check username customization
	usernameField, _ := testSchema.GetFieldPtr("username")
	suite.Require().Equal("Custom Username Label", usernameField.Label)
	suite.Require().False(usernameField.Required)
	suite.Require().Equal("Enter custom username", usernameField.Placeholder)

	// Check email customization
	emailField, _ := testSchema.GetFieldPtr("email")
	suite.Require().True(emailField.Hidden)
	suite.Require().Equal("default@tenant.com", emailField.Default)
}

// Test action enrichment
func (suite *EnricherTestSuite) TestEnrichActions() {
	user := &DefaultUser{
		ID:          "user123",
		TenantID:    "tenant456",
		Permissions: []string{"read", "write"},
		Roles:       []string{"user"},
	}

	testSchemaTemplate := &schema.Schema{
		ID:    "test-schema",
		Type:  schema.TypeForm,
		Title: "Test Schema",
		Actions: []schema.Action{
			{
				ID:   "public-action",
				Type: schema.ActionButton,
				Text: "Public Action",
			},
			{
				ID:   "admin-action",
				Type: schema.ActionButton,
				Text: "Admin Action",
				Permissions: &schema.ActionPermissions{
					View:    []string{"admin"},
					Execute: []string{"admin"},
				},
			},
			{
				ID:   "user-action",
				Type: schema.ActionButton,
				Text: "User Action",
				Permissions: &schema.ActionPermissions{
					View:    []string{"read"},
					Execute: []string{"write"},
				},
			},
		},
	}

	// Clone the schema to test enrichment
	testSchema, err := testSchemaTemplate.Clone()
	suite.Require().NoError(err)

	err = suite.enricher.Enrich(suite.ctx, testSchema, WithUser(user))
	suite.Require().NoError(err)

	// Public action should remain unchanged
	publicAction := testSchema.Actions[0]
	suite.Require().False(publicAction.Hidden)
	suite.Require().False(publicAction.Disabled)

	// Admin action should be hidden (user doesn't have admin permission)
	adminAction := testSchema.Actions[1]
	suite.Require().True(adminAction.Hidden)
	suite.Require().True(adminAction.Disabled)

	// User action should be visible and enabled (user has read and write permissions)
	userAction := testSchema.Actions[2]
	suite.Require().False(userAction.Hidden)
	suite.Require().False(userAction.Disabled)
}

// Test enrichment with nil inputs
func (suite *EnricherTestSuite) TestEnrichNilInputs() {
	user := &DefaultUser{ID: "user123", TenantID: "tenant456"}

	// Test nil schema
	err := suite.enricher.Enrich(suite.ctx, nil, WithUser(user))
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "schema cannot be nil")

	// Test no user in options
	testSchemaTemplate := &schema.Schema{ID: "test", Type: schema.TypeForm, Title: "Test"}
	err = suite.enricher.Enrich(suite.ctx, testSchemaTemplate)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "no user information provided")
}

// Test field enrichment with nil field
func (suite *EnricherTestSuite) TestEnrichFieldNil() {
	user := &DefaultUser{ID: "user123", TenantID: "tenant456"}
	data := map[string]any{}

	err := suite.enricher.EnrichField(suite.ctx, nil, user, data)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "field cannot be nil")
}

// Test default user implementation
func (suite *EnricherTestSuite) TestDefaultUser() {
	user := &DefaultUser{
		ID:          "user123",
		TenantID:    "tenant456",
		Permissions: []string{"read", "write", "admin"},
		Roles:       []string{"user", "manager"},
	}

	suite.Require().Equal("user123", user.GetID())
	suite.Require().Equal("tenant456", user.GetTenantID())
	suite.Require().Equal([]string{"read", "write", "admin"}, user.GetPermissions())
	suite.Require().Equal([]string{"user", "manager"}, user.GetRoles())

	// Test permission checks
	suite.Require().True(user.HasPermission("read"))
	suite.Require().True(user.HasPermission("write"))
	suite.Require().True(user.HasPermission("admin"))
	suite.Require().False(user.HasPermission("super_admin"))

	// Test role checks
	suite.Require().True(user.HasRole("user"))
	suite.Require().True(user.HasRole("manager"))
	suite.Require().False(user.HasRole("admin"))
}

// Test tenant provider error handling
func (suite *EnricherTestSuite) TestTenantProviderError() {
	// Mock provider that returns an error
	errorProvider := &ErrorMockTenantProvider{}
	suite.enricher.SetTenantProvider(errorProvider)

	user := &DefaultUser{
		ID:       "user123",
		TenantID: "tenant456",
	}

	testSchemaTemplate := &schema.Schema{
		ID:    "test-schema",
		Type:  schema.TypeForm,
		Title: "Test Schema",
		Fields: []schema.Field{
			{
				Name:  "username",
				Type:  schema.FieldText,
				Label: "Username",
			},
		},
	}

	// Should not fail even if tenant provider returns error
	// Clone the schema to test enrichment
	testSchema, err := testSchemaTemplate.Clone()
	suite.Require().NoError(err)

	err = suite.enricher.Enrich(suite.ctx, testSchema, WithUser(user))
	suite.Require().NoError(err)
	suite.Require().NotNil(&testSchema)
}

// Test readonly field enrichment
func (suite *EnricherTestSuite) TestEnrichReadonlyField() {
	user := &DefaultUser{
		ID:       "user123",
		TenantID: "tenant456",
	}

	testSchemaTemplate := &schema.Schema{
		ID:    "test-schema",
		Type:  schema.TypeForm,
		Title: "Test Schema",
		Fields: []schema.Field{
			{
				Name:     "readonly_field",
				Type:     schema.FieldText,
				Label:    "Readonly Field",
				Readonly: true,
			},
		},
	}

	// Clone the schema to test enrichment using JSON marshaling to rebuild fieldMap
	schemaBytes, err := json.Marshal(testSchemaTemplate)
	suite.Require().NoError(err)
	var testSchema schema.Schema
	err = json.Unmarshal(schemaBytes, &testSchema)
	suite.Require().NoError(err)

	err = suite.enricher.Enrich(suite.ctx, &testSchema, WithUser(user))
	suite.Require().NoError(err)

	field, _ := testSchema.GetFieldPtr("readonly_field")
	suite.Require().True(field.Runtime.Visible)
	suite.Require().False(field.Runtime.Editable) // Should be false due to readonly
}

// Mock provider that returns errors
type ErrorMockTenantProvider struct{}

func (e *ErrorMockTenantProvider) GetCustomization(ctx context.Context, schemaID, tenantID string) (*TenantCustomization, error) {
	return nil, fmt.Errorf("mock error")
}

// Helper function to create bool pointer
func boolPtr(b bool) *bool {
	return &b
}

// Test comprehensive enrichment scenario
func (suite *EnricherTestSuite) TestComprehensiveEnrichment() {
	// Set up provider with customizations
	provider := NewMockTenantProvider()
	suite.enricher.SetTenantProvider(provider)

	// User with specific permissions and roles
	user := &DefaultUser{
		ID:          "user123",
		TenantID:    "tenant456",
		Permissions: []string{"read", "write", "user_management"},
		Roles:       []string{"manager"},
	}

	// Tenant customization
	customization := &TenantCustomization{
		SchemaID: "comprehensive-schema",
		TenantID: "tenant456",
		Overrides: map[string]TenantOverride{
			"company_name": {
				FieldName:    "company_name",
				Label:        "Organization Name",
				Required:     boolPtr(true),
				DefaultValue: "Default Org",
			},
		},
	}
	provider.SetCustomization("comprehensive-schema", "tenant456", customization)

	// Complex schema with various field types and permissions
	testSchemaTemplate := &schema.Schema{
		ID:    "comprehensive-schema",
		Type:  schema.TypeForm,
		Title: "Comprehensive Test Schema",
		Fields: []schema.Field{
			{
				Name:  "created_by",
				Type:  schema.FieldText,
				Label: "Created By",
			},
			{
				Name:  "tenant_id",
				Type:  schema.FieldText,
				Label: "Tenant ID",
			},
			{
				Name:              "admin_notes",
				Type:              schema.FieldTextarea,
				Label:             "Admin Notes",
				RequirePermission: "admin",
			},
			{
				Name:         "manager_field",
				Type:         schema.FieldText,
				Label:        "Manager Field",
				RequireRoles: []string{"manager", "admin"},
			},
			{
				Name:     "company_name",
				Type:     schema.FieldText,
				Label:    "Company Name",
				Required: false,
				Default:  "Original Default",
			},
		},
		Actions: []schema.Action{
			{
				ID:   "submit",
				Type: schema.ActionSubmit,
				Text: "Submit",
			},
			{
				ID:   "delete",
				Type: schema.ActionButton,
				Text: "Delete",
				Permissions: &schema.ActionPermissions{
					Execute: []string{"admin"},
				},
			},
		},
	}

	// Clone the schema to test enrichment using JSON marshaling to rebuild fieldMap
	schemaBytes, err := json.Marshal(testSchemaTemplate)
	suite.Require().NoError(err)
	var testSchema schema.Schema
	err = json.Unmarshal(schemaBytes, &testSchema)
	suite.Require().NoError(err)

	err = suite.enricher.Enrich(suite.ctx, &testSchema, WithUser(user))
	suite.Require().NoError(err)
	suite.Require().NotNil(&testSchema)

	// Check created_by field got user ID
	createdByField, _ := testSchema.GetFieldPtr("created_by")
	suite.Require().Equal("user123", createdByField.Value)

	// Check tenant_id field is hidden and readonly
	tenantField, _ := testSchema.GetFieldPtr("tenant_id")
	suite.Require().Equal("tenant456", tenantField.Value)
	suite.Require().True(tenantField.Hidden)
	suite.Require().True(tenantField.Readonly)

	// Check admin field is not visible (user doesn't have admin permission)
	adminField, _ := testSchema.GetFieldPtr("admin_notes")
	suite.Require().False(adminField.Runtime.Visible)

	// Check manager field is visible (user has manager role)
	managerField, _ := testSchema.GetFieldPtr("manager_field")
	suite.Require().True(managerField.Runtime.Visible)

	// Check tenant customization was applied
	companyField, _ := testSchema.GetFieldPtr("company_name")
	suite.Require().Equal("Organization Name", companyField.Label)
	suite.Require().True(companyField.Required)
	suite.Require().Equal("Default Org", companyField.Default)

	// Check actions
	suite.Require().Len(testSchema.Actions, 2)

	// Submit action should be enabled
	submitAction := testSchema.Actions[0]
	suite.Require().False(submitAction.Hidden)
	suite.Require().False(submitAction.Disabled)

	// Delete action should be disabled (user doesn't have admin permission)
	deleteAction := testSchema.Actions[1]
	suite.Require().False(deleteAction.Hidden)  // View permission not set, so not hidden
	suite.Require().True(deleteAction.Disabled) // Execute permission requires admin
}
