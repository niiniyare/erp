package schema

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// MixinTestSuite tests the Mixin functionality
type MixinTestSuite struct {
	suite.Suite
	registry *MixinRegistry
	mixin    *Mixin
}

func (s *MixinTestSuite) SetupTest() {
	s.registry = NewMixinRegistry()
	s.mixin = &Mixin{
		ID:          "test_mixin",
		Name:        "Test Mixin",
		Description: "A test mixin for testing",
		Version:     "1.0.0",
		Fields: []Field{
			{Name: "field1", Type: FieldText, Label: "Field 1"},
			{Name: "field2", Type: FieldEmail, Label: "Field 2"},
		},
		Actions: []Action{
			{ID: "action1", Type: ActionButton, Text: "Test Action"},
		},
		Category: "test",
		Tags:     []string{"test", "example"},
		Metadata: map[string]any{"version": "1.0"},
	}
}

func (s *MixinTestSuite) TestNewMixinRegistry() {
	registry := NewMixinRegistry()
	require.NotNil(s.T(), registry)
	require.NotNil(s.T(), registry.mixins)

	// Should have built-in mixins registered
	require.Greater(s.T(), len(registry.mixins), 0, "Should have built-in mixins")

	// Check for expected built-in mixins
	_, err := registry.Get("audit_fields")
	require.NoError(s.T(), err, "Should have audit_fields mixin")

	_, err = registry.Get("address_fields")
	require.NoError(s.T(), err, "Should have address_fields mixin")

	_, err = registry.Get("contact_fields")
	require.NoError(s.T(), err, "Should have contact_fields mixin")

	_, err = registry.Get("status_fields")
	require.NoError(s.T(), err, "Should have status_fields mixin")
}

func (s *MixinTestSuite) TestMixinRegistration() {
	err := s.registry.Register(s.mixin)
	require.NoError(s.T(), err)

	// Test duplicate registration
	err = s.registry.Register(s.mixin)
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "already exists")
}

func (s *MixinTestSuite) TestMixinRetrieval() {
	// Test getting non-existent mixin
	_, err := s.registry.Get("nonexistent")
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "not found")

	// Register and retrieve mixin
	err = s.registry.Register(s.mixin)
	require.NoError(s.T(), err)

	retrieved, err := s.registry.Get("test_mixin")
	require.NoError(s.T(), err)
	require.Equal(s.T(), s.mixin.ID, retrieved.ID)
	require.Equal(s.T(), s.mixin.Name, retrieved.Name)
	require.Len(s.T(), retrieved.Fields, 2)
	require.Len(s.T(), retrieved.Actions, 1)
}

func (s *MixinTestSuite) TestMixinListing() {
	// Get initial count (built-in mixins)
	initialList := s.registry.List()
	initialCount := len(initialList)
	require.Greater(s.T(), initialCount, 0, "Should have built-in mixins")

	// Register a new mixin
	err := s.registry.Register(s.mixin)
	require.NoError(s.T(), err)

	// Check updated list
	updatedList := s.registry.List()
	require.Len(s.T(), updatedList, initialCount+1)
	require.Contains(s.T(), updatedList, "test_mixin")
}

func (s *MixinTestSuite) TestMixinApplication() {
	// Register the test mixin
	err := s.registry.Register(s.mixin)
	require.NoError(s.T(), err)

	// Create a schema to apply mixin to
	schema := NewSchema("test_schema", TypeForm, "Test Schema")
	require.Len(s.T(), schema.Fields, 0)
	require.Len(s.T(), schema.Actions, 0)

	// Apply mixin
	err = s.registry.ApplyMixin(schema, "test_mixin", "")
	require.NoError(s.T(), err)

	// Check that fields and actions were added
	require.Len(s.T(), schema.Fields, 2)
	require.Len(s.T(), schema.Actions, 1)
	require.Equal(s.T(), "field1", schema.Fields[0].Name)
	require.Equal(s.T(), "field2", schema.Fields[1].Name)
	require.Equal(s.T(), "action1", schema.Actions[0].ID)
}

func (s *MixinTestSuite) TestMixinApplicationWithPrefix() {
	// Register the test mixin
	err := s.registry.Register(s.mixin)
	require.NoError(s.T(), err)

	// Create a schema to apply mixin to
	schema := NewSchema("test_schema", TypeForm, "Test Schema")

	// Apply mixin with prefix
	err = s.registry.ApplyMixin(schema, "test_mixin", "pre")
	require.NoError(s.T(), err)

	// Check that fields have prefixed names
	require.Len(s.T(), schema.Fields, 2)
	require.Equal(s.T(), "pre_field1", schema.Fields[0].Name)
	require.Equal(s.T(), "pre_field2", schema.Fields[1].Name)
}

func (s *MixinTestSuite) TestMixinApplicationNonExistent() {
	schema := NewSchema("test_schema", TypeForm, "Test Schema")

	// Try to apply non-existent mixin
	err := s.registry.ApplyMixin(schema, "nonexistent", "")
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "not found")
}

func (s *MixinTestSuite) TestMixinValidation() {
	// Test valid mixin
	validMixin := &Mixin{
		ID:   "valid_mixin",
		Name: "Valid Mixin",
		Fields: []Field{
			{Name: "test_field", Type: FieldText, Label: "Test Field"},
		},
	}

	err := s.registry.Register(validMixin)
	require.NoError(s.T(), err)

	// Test invalid mixin (empty ID)
	invalidMixin := &Mixin{
		Name: "Invalid Mixin",
	}

	err = s.registry.Register(invalidMixin)
	require.Error(s.T(), err)
}

func (s *MixinTestSuite) TestMixinMetadata() {
	now := time.Now()
	s.mixin.Meta = &Meta{
		CreatedAt:  now,
		UpdatedAt:  now,
		CreatedBy:  "user123",
		CustomData: map[string]any{"test": "value"},
	}

	err := s.registry.Register(s.mixin)
	require.NoError(s.T(), err)

	retrieved, err := s.registry.Get("test_mixin")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), retrieved.Meta)
	require.Equal(s.T(), "user123", retrieved.Meta.CreatedBy)
	require.Equal(s.T(), "value", retrieved.Meta.CustomData["test"])
}

func (s *MixinTestSuite) TestMixinSecurity() {
	s.mixin.Security = &Security{
		CSRF: &CSRF{
			Enabled:   true,
			FieldName: "_csrf",
		},
	}

	err := s.registry.Register(s.mixin)
	require.NoError(s.T(), err)

	retrieved, err := s.registry.Get("test_mixin")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), retrieved.Security)
	require.NotNil(s.T(), retrieved.Security.CSRF)
	require.True(s.T(), retrieved.Security.CSRF.Enabled)
}

func (s *MixinTestSuite) TestMixinEvents() {
	s.mixin.Events = &Events{
		OnLoad:   "handleLoad",
		OnSubmit: "handleSubmit",
	}

	err := s.registry.Register(s.mixin)
	require.NoError(s.T(), err)

	retrieved, err := s.registry.Get("test_mixin")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), retrieved.Events)
	require.Equal(s.T(), "handleLoad", retrieved.Events.OnLoad)
	require.Equal(s.T(), "handleSubmit", retrieved.Events.OnSubmit)
}

func (s *MixinTestSuite) TestBuiltInAuditFields() {
	auditMixin, err := s.registry.Get("audit_fields")
	require.NoError(s.T(), err)
	require.Equal(s.T(), "audit_fields", auditMixin.ID)
	require.Equal(s.T(), "Audit Fields", auditMixin.Name)
	require.Greater(s.T(), len(auditMixin.Fields), 0)

	// Check for expected audit fields
	fieldNames := make([]string, len(auditMixin.Fields))
	for i, field := range auditMixin.Fields {
		fieldNames[i] = field.Name
	}
	require.Contains(s.T(), fieldNames, "created_at")
	require.Contains(s.T(), fieldNames, "updated_at")
	require.Contains(s.T(), fieldNames, "created_by")
	require.Contains(s.T(), fieldNames, "updated_by")
}

func (s *MixinTestSuite) TestBuiltInAddressFields() {
	addressMixin, err := s.registry.Get("address_fields")
	require.NoError(s.T(), err)
	require.Equal(s.T(), "address_fields", addressMixin.ID)
	require.Equal(s.T(), "Address Fields", addressMixin.Name)
	require.Greater(s.T(), len(addressMixin.Fields), 0)

	// Check for expected address fields
	fieldNames := make([]string, len(addressMixin.Fields))
	for i, field := range addressMixin.Fields {
		fieldNames[i] = field.Name
	}
	require.Contains(s.T(), fieldNames, "street_address")
	require.Contains(s.T(), fieldNames, "city")
	require.Contains(s.T(), fieldNames, "state_province")
	require.Contains(s.T(), fieldNames, "postal_code")
	require.Contains(s.T(), fieldNames, "country")
}

func (s *MixinTestSuite) TestBuiltInContactFields() {
	contactMixin, err := s.registry.Get("contact_fields")
	require.NoError(s.T(), err)
	require.Equal(s.T(), "contact_fields", contactMixin.ID)
	require.Equal(s.T(), "Contact Fields", contactMixin.Name)
	require.Greater(s.T(), len(contactMixin.Fields), 0)

	// Check for expected contact fields
	fieldNames := make([]string, len(contactMixin.Fields))
	for i, field := range contactMixin.Fields {
		fieldNames[i] = field.Name
	}
	require.Contains(s.T(), fieldNames, "phone")
	require.Contains(s.T(), fieldNames, "email")
	require.Contains(s.T(), fieldNames, "fax")
	require.Contains(s.T(), fieldNames, "website")
}

func (s *MixinTestSuite) TestBuiltInStatusFields() {
	statusMixin, err := s.registry.Get("status_fields")
	require.NoError(s.T(), err)
	require.Equal(s.T(), "status_fields", statusMixin.ID)
	require.Equal(s.T(), "Status Fields", statusMixin.Name)
	require.Greater(s.T(), len(statusMixin.Fields), 0)

	// Check for expected status fields
	fieldNames := make([]string, len(statusMixin.Fields))
	for i, field := range statusMixin.Fields {
		fieldNames[i] = field.Name
	}
	require.Contains(s.T(), fieldNames, "status")
	require.Contains(s.T(), fieldNames, "is_active")
	require.Contains(s.T(), fieldNames, "is_deleted")
}

func (s *MixinTestSuite) TestMixinCategoriesAndTags() {
	s.mixin.Category = "forms"
	s.mixin.Tags = []string{"user", "contact", "form"}

	err := s.registry.Register(s.mixin)
	require.NoError(s.T(), err)

	retrieved, err := s.registry.Get("test_mixin")
	require.NoError(s.T(), err)
	require.Equal(s.T(), "forms", retrieved.Category)
	require.Contains(s.T(), retrieved.Tags, "user")
	require.Contains(s.T(), retrieved.Tags, "contact")
	require.Contains(s.T(), retrieved.Tags, "form")
}

func (s *MixinTestSuite) TestMixinFieldIntegration() {
	// Create a complex field in mixin
	complexField := Field{
		Name:     "complex_field",
		Type:     FieldSelect,
		Label:    "Complex Field",
		Required: true,
		Options: []Option{
			{Value: "option1", Label: "Option 1"},
			{Value: "option2", Label: "Option 2"},
		},
		Validation: &FieldValidation{
			Messages: &Messages{
				Required: "This field is required",
			},
		},
	}

	s.mixin.Fields = append(s.mixin.Fields, complexField)

	err := s.registry.Register(s.mixin)
	require.NoError(s.T(), err)

	// Apply to schema and verify complex field structure
	schema := NewSchema("test_schema", TypeForm, "Test Schema")
	err = s.registry.ApplyMixin(schema, "test_mixin", "")
	require.NoError(s.T(), err)

	// Find the complex field
	var found *Field
	for i := range schema.Fields {
		if schema.Fields[i].Name == "complex_field" {
			found = &schema.Fields[i]
			break
		}
	}

	require.NotNil(s.T(), found)
	require.Equal(s.T(), FieldSelect, found.Type)
	require.True(s.T(), found.Required)
	require.Len(s.T(), found.Options, 2)
	require.NotNil(s.T(), found.Validation)
}

// Run the test suite
func TestMixinTestSuite(t *testing.T) {
	suite.Run(t, new(MixinTestSuite))
}
