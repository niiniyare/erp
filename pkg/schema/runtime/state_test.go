package runtime

import (
	"reflect"
	"testing"
)

func TestNewState(t *testing.T) {
	state := NewState()

	if state == nil {
		t.Fatal("NewState() returned nil")
	}

	if state.values == nil {
		t.Error("State values not initialized")
	}

	if state.touched == nil {
		t.Error("State touched not initialized")
	}

	if state.dirty == nil {
		t.Error("State dirty not initialized")
	}

	if state.errors == nil {
		t.Error("State errors not initialized")
	}

	if state.initial == nil {
		t.Error("State initial not initialized")
	}
}

func TestState_Initialize(t *testing.T) {
	state := NewState()
	schema := createTestSchema()

	// Set some default values in schema
	schema.Fields[0].Default = "Default Name"
	schema.Fields[1].Value = "default@example.com"

	initialData := map[string]any{
		"name": "Provided Name", // Should override default
		"age":  25,              // No default in schema
	}

	err := state.Initialize(schema, initialData)
	if err != nil {
		t.Errorf("Initialize() error = %v", err)
	}

	// Check that provided data overrides defaults
	name, exists := state.GetValue("name")
	if !exists || name != "Provided Name" {
		t.Errorf("Expected 'Provided Name', got %v", name)
	}

	// Check that schema Value is used when no provided data
	email, exists := state.GetValue("email")
	if !exists || email != "default@example.com" {
		t.Errorf("Expected 'default@example.com', got %v", email)
	}

	// Check that provided data is set
	age, exists := state.GetValue("age")
	if !exists || age != 25 {
		t.Errorf("Expected 25, got %v", age)
	}

	// Check initial values are stored
	initialValues := state.GetInitialValues()
	if initialValues["name"] != "Provided Name" {
		t.Error("Initial value for name not stored correctly")
	}
}

func TestState_SetValue(t *testing.T) {
	state := NewState()

	// Initialize with some data
	initialData := map[string]any{
		"name": "John Doe",
		"age":  30,
	}
	state.Initialize(createTestSchema(), initialData)

	tests := []struct {
		name      string
		path      string
		value     any
		wantError bool
		wantDirty bool
	}{
		{
			name:      "set new value",
			path:      "name",
			value:     "Jane Doe",
			wantError: false,
			wantDirty: true,
		},
		{
			name:      "set same value as initial",
			path:      "age",
			value:     30,
			wantError: false,
			wantDirty: false,
		},
		{
			name:      "empty path",
			path:      "",
			value:     "test",
			wantError: true,
			wantDirty: false,
		},
		{
			name:      "set new field",
			path:      "new_field",
			value:     "new value",
			wantError: false,
			wantDirty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := state.SetValue(tt.path, tt.value)

			if (err != nil) != tt.wantError {
				t.Errorf("SetValue() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				// Check value was set
				value, exists := state.GetValue(tt.path)
				if !exists {
					t.Errorf("Value not set for path %s", tt.path)
				} else if value != tt.value {
					t.Errorf("Value = %v, want %v", value, tt.value)
				}

				// Check dirty state
				isDirty := state.IsDirty(tt.path)
				if isDirty != tt.wantDirty {
					t.Errorf("IsDirty() = %v, want %v", isDirty, tt.wantDirty)
				}
			}
		})
	}
}

func TestState_Touch(t *testing.T) {
	state := NewState()

	// Initially not touched
	if state.IsTouched("name") {
		t.Error("Field should not be touched initially")
	}

	// Touch the field
	state.Touch("name")

	// Now should be touched
	if !state.IsTouched("name") {
		t.Error("Field should be touched after Touch()")
	}

	// Check touched count
	if state.GetTouchedCount() != 1 {
		t.Errorf("Expected 1 touched field, got %d", state.GetTouchedCount())
	}
}

func TestState_Errors(t *testing.T) {
	state := NewState()

	// Initially no errors
	if !state.IsValid() {
		t.Error("State should be valid initially")
	}

	if len(state.GetErrors("name")) != 0 {
		t.Error("Should have no errors initially")
	}

	// Set some errors
	errors := []string{"Field is required", "Invalid format"}
	state.SetErrors("name", errors)

	// Check errors are set
	if state.IsValid() {
		t.Error("State should not be valid with errors")
	}

	retrievedErrors := state.GetErrors("name")
	if !reflect.DeepEqual(retrievedErrors, errors) {
		t.Errorf("GetErrors() = %v, want %v", retrievedErrors, errors)
	}

	// Check error count
	if state.GetErrorCount() != 2 {
		t.Errorf("Expected 2 errors, got %d", state.GetErrorCount())
	}

	// Clear errors
	state.ClearErrors("name")

	// Check errors are cleared
	if !state.IsValid() {
		t.Error("State should be valid after clearing errors")
	}

	if len(state.GetErrors("name")) != 0 {
		t.Error("Errors should be cleared")
	}
}

func TestState_GetAll(t *testing.T) {
	state := NewState()

	testData := map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}

	// Set values
	for key, value := range testData {
		state.SetValue(key, value)
	}

	// Get all values
	allValues := state.GetAll()

	// Check all values are returned
	for key, expectedValue := range testData {
		if value, exists := allValues[key]; !exists {
			t.Errorf("Missing value for key %s", key)
		} else if value != expectedValue {
			t.Errorf("Value for %s = %v, want %v", key, value, expectedValue)
		}
	}

	// Modify returned map (should not affect internal state)
	allValues["new_key"] = "new_value"

	// Check internal state is not modified
	if _, exists := state.GetValue("new_key"); exists {
		t.Error("Internal state should not be modified by external changes to GetAll() result")
	}
}

func TestState_Reset(t *testing.T) {
	state := NewState()

	// Initialize with data
	initialData := map[string]any{
		"name": "John Doe",
		"age":  30,
	}
	state.Initialize(createTestSchema(), initialData)

	// Make changes
	state.SetValue("name", "Jane Doe")
	state.Touch("name")
	state.SetErrors("name", []string{"Some error"})

	// Verify changes
	if !state.IsDirty("name") {
		t.Error("Field should be dirty before reset")
	}
	if !state.IsTouched("name") {
		t.Error("Field should be touched before reset")
	}
	if state.IsValid() {
		t.Error("State should be invalid before reset")
	}

	// Reset
	state.Reset()

	// Check state is reset
	if state.IsDirty("name") {
		t.Error("Field should not be dirty after reset")
	}
	if state.IsTouched("name") {
		t.Error("Field should not be touched after reset")
	}
	if !state.IsValid() {
		t.Error("State should be valid after reset")
	}

	// Check values are back to initial
	name, _ := state.GetValue("name")
	if name != "John Doe" {
		t.Errorf("Name should be reset to initial value, got %v", name)
	}
}

func TestState_ResetField(t *testing.T) {
	state := NewState()

	// Initialize with data
	initialData := map[string]any{
		"name": "John Doe",
		"age":  30,
	}
	state.Initialize(createTestSchema(), initialData)

	// Make changes to both fields
	state.SetValue("name", "Jane Doe")
	state.SetValue("age", 25)
	state.Touch("name")
	state.Touch("age")
	state.SetErrors("name", []string{"Name error"})
	state.SetErrors("age", []string{"Age error"})

	// Reset only name field
	state.ResetField("name")

	// Check name field is reset
	name, _ := state.GetValue("name")
	if name != "John Doe" {
		t.Errorf("Name should be reset to initial value, got %v", name)
	}
	if state.IsDirty("name") {
		t.Error("Name field should not be dirty after reset")
	}
	if state.IsTouched("name") {
		t.Error("Name field should not be touched after reset")
	}
	if len(state.GetErrors("name")) != 0 {
		t.Error("Name field should have no errors after reset")
	}

	// Check age field is unchanged
	age, _ := state.GetValue("age")
	if age != 25 {
		t.Errorf("Age should remain changed, got %v", age)
	}
	if !state.IsDirty("age") {
		t.Error("Age field should still be dirty")
	}
}

func TestState_UpdateValues(t *testing.T) {
	state := NewState()

	// Initialize with data
	initialData := map[string]any{
		"name": "John Doe",
		"age":  30,
	}
	state.Initialize(createTestSchema(), initialData)

	// Update multiple values
	updates := map[string]any{
		"name":  "Jane Doe",
		"email": "jane@example.com",
		"age":   30, // Same as initial
	}

	err := state.UpdateValues(updates)
	if err != nil {
		t.Errorf("UpdateValues() error = %v", err)
	}

	// Check all values were updated
	for key, expectedValue := range updates {
		value, exists := state.GetValue(key)
		if !exists {
			t.Errorf("Value not set for key %s", key)
		} else if value != expectedValue {
			t.Errorf("Value for %s = %v, want %v", key, value, expectedValue)
		}
	}

	// Check dirty states
	if !state.IsDirty("name") {
		t.Error("Name should be dirty (changed from initial)")
	}
	if !state.IsDirty("email") {
		t.Error("Email should be dirty (new field)")
	}
	if state.IsDirty("age") {
		t.Error("Age should not be dirty (same as initial)")
	}
}

func TestState_GetChangedValues(t *testing.T) {
	state := NewState()

	// Initialize with data
	initialData := map[string]any{
		"name": "John Doe",
		"age":  30,
	}
	state.Initialize(createTestSchema(), initialData)

	// Make some changes
	state.SetValue("name", "Jane Doe")      // Changed
	state.SetValue("age", 30)               // Same as initial
	state.SetValue("email", "new@test.com") // New field

	changedValues := state.GetChangedValues()

	// Should only include changed values
	expectedChanged := map[string]any{
		"name":  "Jane Doe",
		"email": "new@test.com",
	}

	if !reflect.DeepEqual(changedValues, expectedChanged) {
		t.Errorf("GetChangedValues() = %v, want %v", changedValues, expectedChanged)
	}
}

func TestState_Snapshot(t *testing.T) {
	state := NewState()

	// Initialize and modify state
	initialData := map[string]any{
		"name": "John Doe",
	}
	state.Initialize(createTestSchema(), initialData)
	state.SetValue("name", "Jane Doe")
	state.Touch("name")
	state.SetErrors("name", []string{"Some error"})

	// Create snapshot
	snapshot := state.CreateSnapshot()

	// Modify state further
	state.SetValue("name", "Bob Smith")
	state.SetValue("age", 25)

	// Restore snapshot
	state.RestoreSnapshot(snapshot)

	// Check state is restored
	name, _ := state.GetValue("name")
	if name != "Jane Doe" {
		t.Errorf("Name should be restored to snapshot value, got %v", name)
	}

	if !state.IsTouched("name") {
		t.Error("Touched state should be restored")
	}

	if len(state.GetErrors("name")) == 0 {
		t.Error("Errors should be restored")
	}

	// Check new field is gone
	if _, exists := state.GetValue("age"); exists {
		t.Error("New field should not exist after restore")
	}
}

func TestState_ConcurrentAccess(t *testing.T) {
	state := NewState()

	// Initialize
	state.Initialize(createTestSchema(), map[string]any{
		"counter": 0,
	})

	// Run concurrent operations
	done := make(chan bool, 10)

	// Start 10 goroutines that increment counter
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			for j := 0; j < 10; j++ {
				// Get current value
				current, _ := state.GetValue("counter")
				currentInt := current.(int)

				// Increment
				state.SetValue("counter", currentInt+1)

				// Touch and set errors
				state.Touch("counter")
				state.SetErrors("counter", []string{"error"})
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check final value (should be 100)
	final, _ := state.GetValue("counter")
	finalInt := final.(int)

	// Due to race conditions, the final value might not be exactly 100
	// but it should be positive and the test should not panic
	if finalInt <= 0 {
		t.Errorf("Expected positive counter value, got %d", finalInt)
	}

	// Check that concurrent access doesn't cause panics or data corruption
	if !state.IsTouched("counter") {
		t.Error("Counter should be touched")
	}
}

func TestState_EdgeCases(t *testing.T) {
	state := NewState()

	// Test nil values
	err := state.SetValue("nil_field", nil)
	if err != nil {
		t.Errorf("Should be able to set nil value: %v", err)
	}

	value, exists := state.GetValue("nil_field")
	if !exists || value != nil {
		t.Error("Nil value should be stored and retrieved correctly")
	}

	// Test empty string path in UpdateValues
	err = state.UpdateValues(map[string]any{
		"": "value",
	})
	if err == nil {
		t.Error("Should not allow empty path in UpdateValues")
	}

	// Test removing non-existent field
	state.RemoveField("non_existent")
	// Should not panic

	// Test getting errors for non-existent field
	errors := state.GetErrors("non_existent")
	if errors != nil {
		t.Error("Should return nil for non-existent field errors")
	}
}
