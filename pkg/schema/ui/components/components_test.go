package components

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test Component Factory Creation and Defaults
func TestComponentFactoryCreation(t *testing.T) {
	factory := NewComponentFactory()
	assert.NotNil(t, factory, "Factory should not be nil")

	// Test extended factory
	extendedFactory := &DefaultComponentFactory{}
	assert.NotNil(t, extendedFactory, "Extended factory should not be nil")
}

// Test CRUD Schema Creation and Validation
func TestCRUDSchemaCreation(t *testing.T) {
	factory := NewComponentFactory()
	
	config := CRUDSchema{
		Title: "User Management",
		Mode:  "table",
	}

	crud, err := factory.CreateCRUD(config)
	require.NoError(t, err, "CRUD creation should not error")
	assert.Equal(t, "crud", crud.Type, "Type should be set to crud")
	assert.Equal(t, "table", crud.Mode, "Mode should be preserved")
	assert.Equal(t, "asc", crud.OrderDir, "Default order direction should be asc")
	assert.Equal(t, []int{10, 20, 50, 100}, crud.PerPageAvailable, "Default per page options should be set")
	assert.Equal(t, "page", crud.PageField, "Default page field should be set")
	assert.Equal(t, "perPage", crud.PerPageField, "Default per page field should be set")

	// Test validation
	err = ValidateComponent(crud)
	assert.NoError(t, err, "CRUD validation should pass")
}

func TestCRUDSchemaValidation(t *testing.T) {
	testCases := []struct {
		name      string
		crud      CRUDSchema
		expectErr bool
		errMsg    string
	}{
		{
			name: "Valid CRUD",
			crud: CRUDSchema{Type: "crud", Mode: "table"},
			expectErr: false,
		},
		{
			name: "Invalid Type",
			crud: CRUDSchema{Type: "invalid", Mode: "table"},
			expectErr: true,
			errMsg: "invalid CRUD type",
		},
		{
			name: "Invalid Mode",
			crud: CRUDSchema{Type: "crud", Mode: "invalid"},
			expectErr: true,
			errMsg: "invalid CRUD mode",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCRUD(&tc.crud)
			if tc.expectErr {
				assert.Error(t, err, "Should return error")
				assert.Contains(t, err.Error(), tc.errMsg, "Error message should contain expected text")
			} else {
				assert.NoError(t, err, "Should not return error")
			}
		})
	}
}

// Test Table Schema Creation and Validation
func TestTableSchemaCreation(t *testing.T) {
	factory := NewComponentFactory()
	
	config := TableSchema{
		Title: "Data Table",
		Columns: []TableColumn{
			{Name: "id", Label: "ID", Type: "text"},
			{Name: "name", Label: "Name", Type: "text"},
		},
	}

	table, err := factory.CreateTable(config)
	require.NoError(t, err, "Table creation should not error")
	assert.Equal(t, "table", table.Type, "Type should be set to table")
	assert.True(t, table.ShowHeader, "ShowHeader should default to true")
	assert.Equal(t, "No data available", table.Placeholder, "Default placeholder should be set")
	assert.Len(t, table.Columns, 2, "Columns should be preserved")

	// Test validation
	err = ValidateComponent(table)
	assert.NoError(t, err, "Table validation should pass")
}

// Test Form Schema Creation and Validation
func TestFormSchemaCreation(t *testing.T) {
	factory := NewComponentFactory()
	
	config := FormSchema{
		Title: "User Form",
		Body: []any{
			map[string]any{
				"type": "input-text",
				"name": "username",
				"label": "Username",
			},
		},
	}

	form, err := factory.CreateForm(config)
	require.NoError(t, err, "Form creation should not error")
	assert.Equal(t, "form", form.Type, "Type should be set to form")
	assert.Equal(t, "normal", form.Mode, "Default mode should be normal")
	assert.Equal(t, "left", form.LabelAlign, "Default label align should be left")
	assert.Equal(t, "Submit", form.SubmitText, "Default submit text should be set")
	assert.Equal(t, 3000, form.CheckInterval, "Default check interval should be 3 seconds")

	// Test validation
	err = ValidateComponent(form)
	assert.NoError(t, err, "Form validation should pass")
}

// Test Chart Schema Creation and Validation
func TestChartSchemaCreation(t *testing.T) {
	factory := NewComponentFactory()
	
	config := ChartSchema{
		Config: map[string]any{
			"type": "line",
			"data": map[string]any{
				"datasets": []map[string]any{
					{"label": "Sales", "data": []int{10, 20, 30}},
				},
			},
		},
	}

	chart, err := factory.CreateChart(config)
	require.NoError(t, err, "Chart creation should not error")
	assert.Equal(t, "chart", chart.Type, "Type should be set to chart")
	assert.Equal(t, "100%", chart.Width, "Default width should be 100%")
	assert.Equal(t, 400, chart.Height, "Default height should be 400")

	// Test validation
	err = ValidateComponent(chart)
	assert.NoError(t, err, "Chart validation should pass")
}

// Test Action Schema Creation and Validation
func TestActionSchemaCreation(t *testing.T) {
	factory := NewComponentFactory()
	
	config := ActionSchema{
		Label: "Save",
		ActionType: "ajax",
		API: &APIConfig{
			URL: "/api/save",
			Method: "POST",
		},
	}

	action, err := factory.CreateAction(config)
	require.NoError(t, err, "Action creation should not error")
	assert.Equal(t, "action", action.Type, "Type should be set to action")
	assert.Equal(t, "primary", action.Level, "Default level should be primary")
	assert.Equal(t, "md", action.Size, "Default size should be md")

	// Test validation
	err = ValidateComponent(action)
	assert.NoError(t, err, "Action validation should pass")
}

// Test Dialog Schema Creation and Validation
func TestDialogSchemaCreation(t *testing.T) {
	factory := NewComponentFactory()
	
	config := DialogSchema{
		Title: "Confirmation",
		Body: []any{
			map[string]any{
				"type": "tpl",
				"tpl": "Are you sure?",
			},
		},
	}

	dialog, err := factory.CreateDialog(config)
	require.NoError(t, err, "Dialog creation should not error")
	assert.Equal(t, "dialog", dialog.Type, "Type should be set to dialog")
	assert.Equal(t, "md", dialog.Size, "Default size should be md")
	assert.True(t, dialog.CloseOnEsc, "CloseOnEsc should default to true")
	assert.True(t, dialog.ShowCloseButton, "ShowCloseButton should default to true")
	assert.True(t, dialog.Overlay, "Overlay should default to true")

	// Test validation
	err = ValidateComponent(dialog)
	assert.NoError(t, err, "Dialog validation should pass")
}

// Test Text Input Schema Creation and Validation
func TestTextInputSchemaCreation(t *testing.T) {
	factory := NewComponentFactory()
	
	config := TextControlSchema{
		Name: "email",
		Label: "Email Address",
		Placeholder: "Enter your email",
		Required: true,
	}

	textInput, err := factory.CreateTextInput(config)
	require.NoError(t, err, "Text input creation should not error")
	assert.Equal(t, "input-text", textInput.Type, "Type should be set to input-text")
	assert.Equal(t, "md", textInput.Size, "Default size should be md")
	assert.True(t, textInput.Trim, "Trim should default to true")
	assert.Equal(t, "email", textInput.Name, "Name should be preserved")
	assert.True(t, textInput.Required, "Required should be preserved")

	// Test validation
	err = ValidateComponent(textInput)
	assert.NoError(t, err, "Text input validation should pass")
}

func TestTextInputValidation(t *testing.T) {
	testCases := []struct {
		name      string
		input     TextControlSchema
		expectErr bool
		errMsg    string
	}{
		{
			name: "Valid Text Input",
			input: TextControlSchema{Type: "input-text", Name: "username"},
			expectErr: false,
		},
		{
			name: "Missing Name",
			input: TextControlSchema{Type: "input-text"},
			expectErr: true,
			errMsg: "text control name is required",
		},
		{
			name: "Invalid Type",
			input: TextControlSchema{Type: "invalid", Name: "username"},
			expectErr: true,
			errMsg: "invalid text control type",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateTextControl(&tc.input)
			if tc.expectErr {
				assert.Error(t, err, "Should return error")
				assert.Contains(t, err.Error(), tc.errMsg, "Error message should contain expected text")
			} else {
				assert.NoError(t, err, "Should not return error")
			}
		})
	}
}

// Test Extended Components (from components_extended.go)
func TestSelectControlSchemaCreation(t *testing.T) {
	factory := &DefaultComponentFactory{}
	
	config := SelectControlSchema{
		Name: "status",
		Label: "Status",
		Multiple: true,
		Options: []SelectOption{
			{Label: "Active", Value: "active"},
			{Label: "Inactive", Value: "inactive"},
		},
	}

	selectControl, err := factory.CreateSelect(config)
	require.NoError(t, err, "Select creation should not error")
	assert.Equal(t, "select", selectControl.Type, "Type should be set to select")
	assert.Equal(t, "md", selectControl.Size, "Default size should be md")
	assert.Equal(t, "label", selectControl.LabelField, "Default label field should be label")
	assert.Equal(t, "value", selectControl.ValueField, "Default value field should be value")
	assert.Equal(t, ",", selectControl.Delimiter, "Default delimiter should be comma for multiple")
	assert.Len(t, selectControl.Options, 2, "Options should be preserved")

	// Test validation
	err = validateExtendedComponent(selectControl)
	assert.NoError(t, err, "Select validation should pass")
}

func TestDateControlSchemaCreation(t *testing.T) {
	factory := &DefaultComponentFactory{}
	
	config := DateControlSchema{
		Name: "birthdate",
		Label: "Birth Date",
		Required: true,
	}

	dateControl, err := factory.CreateDateControl(config)
	require.NoError(t, err, "Date control creation should not error")
	assert.Equal(t, "input-date", dateControl.Type, "Type should be set to input-date")
	assert.Equal(t, "md", dateControl.Size, "Default size should be md")
	assert.Equal(t, "YYYY-MM-DD", dateControl.Format, "Default format should be YYYY-MM-DD")
	assert.Equal(t, "Select date", dateControl.Placeholder, "Default placeholder should be set")

	// Test validation
	err = validateExtendedComponent(dateControl)
	assert.NoError(t, err, "Date control validation should pass")
}

func TestNavSchemaCreation(t *testing.T) {
	factory := &DefaultComponentFactory{}
	
	config := NavSchema{
		Links: []NavItem{
			{Label: "Dashboard", To: "/dashboard", Icon: "dashboard"},
			{Label: "Users", To: "/users", Icon: "users"},
		},
	}

	nav, err := factory.CreateNav(config)
	require.NoError(t, err, "Nav creation should not error")
	assert.Equal(t, "nav", nav.Type, "Type should be set to nav")
	assert.Equal(t, "inline", nav.Mode, "Default mode should be inline")
	assert.True(t, nav.Stacked, "Stacked should default to true")
	assert.Len(t, nav.Links, 2, "Links should be preserved")

	// Test validation
	err = validateExtendedComponent(nav)
	assert.NoError(t, err, "Nav validation should pass")
}

// Test Component Registry
func TestComponentRegistry(t *testing.T) {
	registry := NewComponentRegistry()
	
	// Test registration
	crud := &CRUDSchema{
		BaseComponentProps: BaseComponentProps{ID: "test-crud"},
		Type: "crud",
		Title: "Test CRUD",
	}
	
	err := registry.Register("test-crud", crud)
	assert.NoError(t, err, "Registration should succeed")
	
	// Test retrieval
	retrieved, exists := registry.Get("test-crud")
	assert.True(t, exists, "Component should exist")
	assert.Equal(t, crud, retrieved, "Retrieved component should match registered")
	
	// Test non-existent component
	_, exists = registry.Get("non-existent")
	assert.False(t, exists, "Non-existent component should not exist")
	
	// Test listing
	names := registry.List()
	assert.Contains(t, names, "test-crud", "List should contain registered component")
}

func TestExtendedComponentRegistry(t *testing.T) {
	registry := NewExtendedComponentRegistry()
	
	// Test extended component registration
	selectControl := &SelectControlSchema{
		BaseComponentProps: BaseComponentProps{ID: "test-select"},
		Type: "select",
		Name: "category",
		Options: []SelectOption{
			{Label: "Option 1", Value: "opt1"},
		},
	}
	
	err := registry.RegisterExtended("test-select", selectControl)
	assert.NoError(t, err, "Extended registration should succeed")
	
	// Test retrieval
	retrieved, exists := registry.Get("test-select")
	assert.True(t, exists, "Extended component should exist")
	assert.Equal(t, selectControl, retrieved, "Retrieved component should match registered")
}

// Test JSON Serialization
func TestJSONSerialization(t *testing.T) {
	// Test CRUD serialization
	crud := &CRUDSchema{
		BaseComponentProps: BaseComponentProps{
			ID: "user-crud",
			ClassName: "custom-crud",
		},
		Type: "crud",
		Title: "User Management",
		Mode: "table",
		Columns: []TableColumn{
			{Name: "id", Label: "ID", Type: "text", Width: 100},
			{Name: "name", Label: "Name", Type: "text", Sortable: true},
		},
	}
	
	// Test ToJSON
	jsonStr, err := ToJSON(crud)
	require.NoError(t, err, "JSON serialization should not error")
	assert.NotEmpty(t, jsonStr, "JSON string should not be empty")
	assert.Contains(t, jsonStr, `"type": "crud"`, "JSON should contain type")
	assert.Contains(t, jsonStr, `"title": "User Management"`, "JSON should contain title")
	
	// Test FromJSON
	var deserializedCRUD CRUDSchema
	err = FromJSON(jsonStr, &deserializedCRUD)
	require.NoError(t, err, "JSON deserialization should not error")
	assert.Equal(t, crud.Type, deserializedCRUD.Type, "Type should match")
	assert.Equal(t, crud.Title, deserializedCRUD.Title, "Title should match")
	assert.Equal(t, crud.ID, deserializedCRUD.ID, "ID should match")
	assert.Len(t, deserializedCRUD.Columns, 2, "Columns should be preserved")
}

func TestFormSchemaComplex(t *testing.T) {
	// Test complex form with validation rules
	form := &FormSchema{
		BaseComponentProps: BaseComponentProps{ID: "complex-form"},
		Type: "form",
		Title: "User Registration",
		Mode: "horizontal",
		Horizontal: &FormHorizontal{
			Left: 3,
			Right: 9,
		},
		Body: []any{
			map[string]any{
				"type": "input-text",
				"name": "username",
				"label": "Username",
				"required": true,
				"minLength": 3,
			},
			map[string]any{
				"type": "input-email",
				"name": "email",
				"label": "Email",
				"required": true,
			},
		},
		Rules: []ComponentValidationRule{
			{
				Name: "username_required",
				Field: "username",
				Required: true,
				Message: "Username is required",
				Severity: SeverityError,
			},
		},
		API: &APIConfig{
			URL: "/api/users",
			Method: "POST",
		},
	}
	
	// Test JSON serialization of complex form
	jsonStr, err := ToJSON(form)
	require.NoError(t, err, "Complex form JSON serialization should not error")
	assert.Contains(t, jsonStr, `"horizontal"`, "JSON should contain horizontal config")
	assert.Contains(t, jsonStr, `"rules"`, "JSON should contain validation rules")
	
	// Test validation
	err = ValidateComponent(form)
	assert.NoError(t, err, "Complex form validation should pass")
}

// Test API Config
func TestAPIConfig(t *testing.T) {
	api := &APIConfig{
		URL: "/api/test",
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/json",
			"Authorization": "Bearer token",
		},
		Data: map[string]any{
			"key": "value",
		},
		Cache: 300,
	}
	
	jsonStr, err := ToJSON(api)
	require.NoError(t, err, "API config serialization should not error")
	assert.Contains(t, jsonStr, `"url": "/api/test"`, "JSON should contain URL")
	assert.Contains(t, jsonStr, `"method": "POST"`, "JSON should contain method")
	assert.Contains(t, jsonStr, `"headers"`, "JSON should contain headers")
}

// Test Table Column with Quick Edit
func TestTableColumnQuickEdit(t *testing.T) {
	column := TableColumn{
		Name: "status",
		Label: "Status",
		Type: "select",
		QuickEdit: &QuickEditConfig{
			Type: "select",
			Mode: "inline",
			SaveAPI: &APIConfig{
				URL: "/api/update-status",
				Method: "PUT",
			},
			SaveImmediate: true,
		},
		Sortable: true,
		Width: 120,
	}
	
	jsonStr, err := ToJSON(column)
	require.NoError(t, err, "Table column serialization should not error")
	assert.Contains(t, jsonStr, `"quickEdit"`, "JSON should contain quick edit config")
	assert.Contains(t, jsonStr, `"saveImmediate": true`, "JSON should contain save immediate flag")
}

// Test Event Configuration
func TestEventConfig(t *testing.T) {
	eventConfig := EventConfig{
		Weight: 10,
		Actions: []ActionSchema{
			{
				Type: "action",
				ActionType: "ajax",
				Label: "Save",
				API: &APIConfig{
					URL: "/api/save",
					Method: "POST",
				},
			},
		},
		Debounce: &DebounceConfig{
			Wait: 300,
			MaxWait: 1000,
			Leading: false,
			Trailing: true,
		},
	}
	
	jsonStr, err := ToJSON(eventConfig)
	require.NoError(t, err, "Event config serialization should not error")
	assert.Contains(t, jsonStr, `"debounce"`, "JSON should contain debounce config")
	assert.Contains(t, jsonStr, `"actions"`, "JSON should contain actions")
}

// Test Component Factory Interface Compliance
func TestFactoryInterfaceCompliance(t *testing.T) {
	factory := NewComponentFactory()
	
	// Test that factory implements ComponentFactory interface
	var _ ComponentFactory = factory
	
	extendedFactory := &DefaultComponentFactory{}
	
	// Test that extended factory implements ExtendedComponentFactory interface
	var _ ExtendedComponentFactory = extendedFactory
}

// Test Edge Cases and Error Conditions
func TestEdgeCases(t *testing.T) {
	t.Run("Empty component validation", func(t *testing.T) {
		err := ValidateComponent(&CRUDSchema{})
		assert.Error(t, err, "Empty CRUD should fail validation")
	})
	
	t.Run("Nil component validation", func(t *testing.T) {
		err := ValidateComponent(nil)
		assert.Error(t, err, "Nil component should fail validation")
	})
	
	t.Run("Invalid JSON deserialization", func(t *testing.T) {
		var crud CRUDSchema
		err := FromJSON("invalid json", &crud)
		assert.Error(t, err, "Invalid JSON should fail deserialization")
	})
	
	t.Run("Unsupported component type", func(t *testing.T) {
		type UnsupportedComponent struct{}
		err := ValidateComponent(&UnsupportedComponent{})
		assert.Error(t, err, "Unsupported component should fail validation")
		assert.Contains(t, err.Error(), "unsupported component type", "Error should mention unsupported type")
	})
}

// Benchmark tests
func BenchmarkCRUDCreation(b *testing.B) {
	factory := NewComponentFactory()
	config := CRUDSchema{
		Title: "Benchmark CRUD",
		Mode: "table",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := factory.CreateCRUD(config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONSerialization(b *testing.B) {
	crud := &CRUDSchema{
		BaseComponentProps: BaseComponentProps{ID: "benchmark"},
		Type: "crud",
		Title: "Benchmark CRUD",
		Columns: make([]TableColumn, 10),
	}
	
	// Initialize columns
	for i := 0; i < 10; i++ {
		crud.Columns[i] = TableColumn{
			Name: "col" + string(rune(i)),
			Label: "Column " + string(rune(i)),
			Type: "text",
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ToJSON(crud)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidation(b *testing.B) {
	crud := &CRUDSchema{
		Type: "crud",
		Mode: "table",
		Title: "Benchmark CRUD",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := ValidateComponent(crud)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Test helper functions
func TestHelperFunctions(t *testing.T) {
	t.Run("ToJSON with nil component", func(t *testing.T) {
		jsonStr, err := ToJSON(nil)
		assert.NoError(t, err, "ToJSON with nil should not error")
		assert.Equal(t, "null", strings.TrimSpace(jsonStr), "Should serialize to null")
	})
	
	t.Run("FromJSON with empty string", func(t *testing.T) {
		var crud CRUDSchema
		err := FromJSON("", &crud)
		assert.Error(t, err, "Empty JSON should fail")
	})
}

// Integration test with all components
func TestFullComponentIntegration(t *testing.T) {
	registry := NewExtendedComponentRegistry()
	factory := registry.ExtendedFactory()
	
	// Create and register all top 20 priority components
	components := make(map[string]any)
	
	// 1. CRUD
	crud, err := factory.CreateCRUD(CRUDSchema{Title: "Users"})
	require.NoError(t, err)
	components["crud"] = crud
	
	// 2. Table
	table, err := factory.CreateTable(TableSchema{Title: "Data Table"})
	require.NoError(t, err)
	components["table"] = table
	
	// 3. Form
	form, err := factory.CreateForm(FormSchema{Title: "User Form"})
	require.NoError(t, err)
	components["form"] = form
	
	// 4. Chart
	chart, err := factory.CreateChart(ChartSchema{})
	require.NoError(t, err)
	components["chart"] = chart
	
	// 5. Action
	action, err := factory.CreateAction(ActionSchema{Label: "Save"})
	require.NoError(t, err)
	components["action"] = action
	
	// 6. Dialog
	dialog, err := factory.CreateDialog(DialogSchema{Title: "Confirm"})
	require.NoError(t, err)
	components["dialog"] = dialog
	
	// 7. Text Input
	textInput, err := factory.CreateTextInput(TextControlSchema{Name: "username"})
	require.NoError(t, err)
	components["textInput"] = textInput
	
	// 8. Select
	selectControl, err := factory.CreateSelect(SelectControlSchema{Name: "status"})
	require.NoError(t, err)
	components["select"] = selectControl
	
	// 9. Date Control
	dateControl, err := factory.CreateDateControl(DateControlSchema{Name: "date"})
	require.NoError(t, err)
	components["dateControl"] = dateControl
	
	// 10. Date Range
	dateRange, err := factory.CreateDateRange(DateRangeControlSchema{Name: "dateRange"})
	require.NoError(t, err)
	components["dateRange"] = dateRange
	
	// Register all components
	for name, component := range components {
		err := registry.RegisterExtended(name, component)
		assert.NoError(t, err, "Should register component: %s", name)
	}
	
	// Verify all components are registered
	names := registry.List()
	assert.Len(t, names, len(components), "All components should be registered")
	
	// Test JSON serialization of all components
	for name, component := range components {
		t.Run("Serialize_"+name, func(t *testing.T) {
			jsonStr, err := ToJSON(component)
			assert.NoError(t, err, "Should serialize component: %s", name)
			assert.NotEmpty(t, jsonStr, "JSON should not be empty for: %s", name)
			
			// Verify JSON is valid by parsing it back
			var parsed map[string]any
			err = json.Unmarshal([]byte(jsonStr), &parsed)
			assert.NoError(t, err, "Generated JSON should be valid for: %s", name)
		})
	}
}

// Test component inheritance and composition
func TestComponentComposition(t *testing.T) {
	// Test complex form with multiple controls
	form := &FormSchema{
		BaseComponentProps: BaseComponentProps{ID: "composition-test"},
		Type: "form",
		Title: "Complex Form",
		Body: []any{
			// Text input
			map[string]any{
				"type": "input-text",
				"name": "firstName",
				"label": "First Name",
				"required": true,
			},
			// Select control
			map[string]any{
				"type": "select",
				"name": "department",
				"label": "Department",
				"options": []map[string]any{
					{"label": "Engineering", "value": "eng"},
					{"label": "Marketing", "value": "mkt"},
				},
			},
			// Date control
			map[string]any{
				"type": "input-date",
				"name": "startDate",
				"label": "Start Date",
				"required": true,
			},
			// Nested panel with more controls
			map[string]any{
				"type": "panel",
				"title": "Additional Information",
				"body": []map[string]any{
					{
						"type": "textarea",
						"name": "notes",
						"label": "Notes",
					},
				},
			},
		},
		Actions: []ActionSchema{
			{
				Type: "action",
				ActionType: "submit",
				Label: "Save",
				Level: "primary",
			},
			{
				Type: "action", 
				ActionType: "reset",
				Label: "Reset",
				Level: "secondary",
			},
		},
	}
	
	// Test serialization of complex form
	jsonStr, err := ToJSON(form)
	require.NoError(t, err, "Complex form serialization should succeed")
	
	// Verify structure
	assert.Contains(t, jsonStr, `"type": "form"`, "Should contain form type")
	assert.Contains(t, jsonStr, `"body"`, "Should contain body")
	assert.Contains(t, jsonStr, `"actions"`, "Should contain actions")
	assert.Contains(t, jsonStr, `"input-text"`, "Should contain text input")
	assert.Contains(t, jsonStr, `"select"`, "Should contain select")
	assert.Contains(t, jsonStr, `"input-date"`, "Should contain date input")
	assert.Contains(t, jsonStr, `"panel"`, "Should contain panel")
	
	// Test validation of complex form
	err = ValidateComponent(form)
	assert.NoError(t, err, "Complex form validation should pass")
}