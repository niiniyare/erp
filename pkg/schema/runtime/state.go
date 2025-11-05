package runtime

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/niiniyare/erp/pkg/schema"
)

// State holds runtime form state and tracks user interactions
type State struct {
	values  map[string]interface{} // Current field values
	touched map[string]bool        // Fields user has interacted with
	dirty   map[string]bool        // Fields that changed from initial
	errors  map[string][]string    // Validation errors per field
	initial map[string]interface{} // Initial values for dirty checking
	mu      sync.RWMutex          // Concurrent access protection
}

// NewState creates a new state manager
func NewState() *State {
	return &State{
		values:  make(map[string]interface{}),
		touched: make(map[string]bool),
		dirty:   make(map[string]bool),
		errors:  make(map[string][]string),
		initial: make(map[string]interface{}),
	}
}

// Initialize sets initial state from schema and provided data
func (s *State) Initialize(schema *schema.Schema, data map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear existing state
	s.values = make(map[string]interface{})
	s.touched = make(map[string]bool)
	s.dirty = make(map[string]bool)
	s.errors = make(map[string][]string)
	s.initial = make(map[string]interface{})

	// Initialize with schema field defaults first
	for _, field := range schema.Fields {
		// Use Default if available, otherwise use Value
		var defaultValue interface{}
		if field.Default != nil {
			defaultValue = field.Default
		} else if field.Value != nil {
			defaultValue = field.Value
		}

		if defaultValue != nil {
			s.values[field.Name] = defaultValue
			s.initial[field.Name] = defaultValue
		}
	}

	// Override with provided data
	for key, value := range data {
		s.values[key] = value
		s.initial[key] = value
	}

	return nil
}

// SetValue updates a field value and tracks dirty state
func (s *State) SetValue(path string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if path == "" {
		return fmt.Errorf("field path cannot be empty")
	}

	s.values[path] = value

	// Mark as dirty if changed from initial
	initialValue, hasInitial := s.initial[path]
	if !hasInitial || !s.valuesEqual(initialValue, value) {
		s.dirty[path] = true
	} else {
		s.dirty[path] = false
	}

	return nil
}

// GetValue retrieves a field value
func (s *State) GetValue(path string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.values[path]
	return value, exists
}

// Touch marks a field as touched (user interacted with it)
func (s *State) Touch(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.touched[path] = true
}

// IsTouched checks if field has been touched
func (s *State) IsTouched(path string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.touched[path]
}

// IsDirty checks if field changed from initial value
func (s *State) IsDirty(path string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.dirty[path]
}

// IsAnyDirty checks if any field has been modified
func (s *State) IsAnyDirty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, dirty := range s.dirty {
		if dirty {
			return true
		}
	}
	return false
}

// GetErrors returns validation errors for a field
func (s *State) GetErrors(path string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	errors, exists := s.errors[path]
	if !exists {
		return nil
	}

	// Return a copy to prevent external modification
	result := make([]string, len(errors))
	copy(result, errors)
	return result
}

// SetErrors sets validation errors for a field
func (s *State) SetErrors(path string, errors []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(errors) == 0 {
		delete(s.errors, path)
	} else {
		// Store a copy to prevent external modification
		errorsCopy := make([]string, len(errors))
		copy(errorsCopy, errors)
		s.errors[path] = errorsCopy
	}
}

// ClearErrors removes all errors for a field
func (s *State) ClearErrors(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.errors, path)
}

// GetAllErrors returns all current validation errors
func (s *State) GetAllErrors() map[string][]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a deep copy to prevent external modification
	result := make(map[string][]string)
	for field, errors := range s.errors {
		errorsCopy := make([]string, len(errors))
		copy(errorsCopy, errors)
		result[field] = errorsCopy
	}

	return result
}

// IsValid checks if entire form is valid (no errors)
func (s *State) IsValid() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.errors) == 0
}

// GetAll returns all current values
func (s *State) GetAll() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]interface{}, len(s.values))
	for k, v := range s.values {
		result[k] = v
	}

	return result
}

// GetInitialValues returns all initial values
func (s *State) GetInitialValues() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]interface{}, len(s.initial))
	for k, v := range s.initial {
		result[k] = v
	}

	return result
}

// GetTouchedFields returns list of touched field names
func (s *State) GetTouchedFields() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var touched []string
	for field, isTouched := range s.touched {
		if isTouched {
			touched = append(touched, field)
		}
	}

	return touched
}

// GetDirtyFields returns list of dirty field names
func (s *State) GetDirtyFields() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var dirty []string
	for field, isDirty := range s.dirty {
		if isDirty {
			dirty = append(dirty, field)
		}
	}

	return dirty
}

// GetFieldsWithErrors returns list of field names that have errors
func (s *State) GetFieldsWithErrors() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var fields []string
	for field := range s.errors {
		fields = append(fields, field)
	}

	return fields
}

// Reset clears all state back to initial
func (s *State) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.values = make(map[string]interface{})
	s.touched = make(map[string]bool)
	s.dirty = make(map[string]bool)
	s.errors = make(map[string][]string)

	// Restore initial values
	for k, v := range s.initial {
		s.values[k] = v
	}
}

// ResetField resets a single field to its initial value
func (s *State) ResetField(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if initialValue, exists := s.initial[path]; exists {
		s.values[path] = initialValue
	} else {
		delete(s.values, path)
	}

	s.dirty[path] = false
	s.touched[path] = false
	delete(s.errors, path)
}

// GetTouchedCount returns number of touched fields
func (s *State) GetTouchedCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, touched := range s.touched {
		if touched {
			count++
		}
	}
	return count
}

// GetDirtyCount returns number of dirty fields
func (s *State) GetDirtyCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, dirty := range s.dirty {
		if dirty {
			count++
		}
	}
	return count
}

// GetErrorCount returns total number of validation errors
func (s *State) GetErrorCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, errors := range s.errors {
		count += len(errors)
	}
	return count
}

// HasField checks if a field exists in the state
func (s *State) HasField(path string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.values[path]
	return exists
}

// RemoveField removes a field completely from state
func (s *State) RemoveField(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.values, path)
	delete(s.initial, path)
	delete(s.touched, path)
	delete(s.dirty, path)
	delete(s.errors, path)
}

// SetInitialValue updates the initial value for a field (useful for dynamic forms)
func (s *State) SetInitialValue(path string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.initial[path] = value

	// Recalculate dirty state based on new initial value
	currentValue, exists := s.values[path]
	if exists {
		s.dirty[path] = !s.valuesEqual(value, currentValue)
	}
}

// GetChangedValues returns only the values that have changed from initial
func (s *State) GetChangedValues() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]interface{})
	for field, isDirty := range s.dirty {
		if isDirty {
			if value, exists := s.values[field]; exists {
				result[field] = value
			}
		}
	}

	return result
}

// UpdateValues updates multiple field values at once
func (s *State) UpdateValues(updates map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for field, value := range updates {
		if field == "" {
			return fmt.Errorf("field path cannot be empty")
		}

		s.values[field] = value

		// Update dirty state
		initialValue, hasInitial := s.initial[field]
		if !hasInitial || !s.valuesEqual(initialValue, value) {
			s.dirty[field] = true
		} else {
			s.dirty[field] = false
		}
	}

	return nil
}

// valuesEqual compares two values for equality, handling different types
func (s *State) valuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	return reflect.DeepEqual(a, b)
}

// StateSnapshot represents a point-in-time snapshot of the state
type StateSnapshot struct {
	Values  map[string]interface{} `json:"values"`
	Touched map[string]bool        `json:"touched"`
	Dirty   map[string]bool        `json:"dirty"`
	Errors  map[string][]string    `json:"errors"`
	Initial map[string]interface{} `json:"initial"`
}

// CreateSnapshot creates a snapshot of current state
func (s *State) CreateSnapshot() *StateSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshot := &StateSnapshot{
		Values:  make(map[string]interface{}),
		Touched: make(map[string]bool),
		Dirty:   make(map[string]bool),
		Errors:  make(map[string][]string),
		Initial: make(map[string]interface{}),
	}

	// Deep copy all state
	for k, v := range s.values {
		snapshot.Values[k] = v
	}
	for k, v := range s.touched {
		snapshot.Touched[k] = v
	}
	for k, v := range s.dirty {
		snapshot.Dirty[k] = v
	}
	for k, v := range s.errors {
		errorsCopy := make([]string, len(v))
		copy(errorsCopy, v)
		snapshot.Errors[k] = errorsCopy
	}
	for k, v := range s.initial {
		snapshot.Initial[k] = v
	}

	return snapshot
}

// RestoreSnapshot restores state from a snapshot
func (s *State) RestoreSnapshot(snapshot *StateSnapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.values = make(map[string]interface{})
	s.touched = make(map[string]bool)
	s.dirty = make(map[string]bool)
	s.errors = make(map[string][]string)
	s.initial = make(map[string]interface{})

	// Restore all state
	for k, v := range snapshot.Values {
		s.values[k] = v
	}
	for k, v := range snapshot.Touched {
		s.touched[k] = v
	}
	for k, v := range snapshot.Dirty {
		s.dirty[k] = v
	}
	for k, v := range snapshot.Errors {
		errorsCopy := make([]string, len(v))
		copy(errorsCopy, v)
		s.errors[k] = errorsCopy
	}
	for k, v := range snapshot.Initial {
		s.initial[k] = v
	}
}