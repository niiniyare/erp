package runtime

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/niiniyare/erp/pkg/schema"
)

// State holds runtime form state and tracks user interactions
// Implements schema.RuntimeStateManager interface for clean integration
type State struct {
	// Core state maps
	values  map[string]any      // Current field values
	initial map[string]any      // Initial values for dirty checking
	touched map[string]bool     // Fields user has interacted with
	dirty   map[string]bool     // Fields that changed from initial
	errors  map[string][]string // Validation errors per field

	// Metadata
	initialized time.Time // When state was initialized
	lastUpdated time.Time // Last state modification

	// Concurrency control
	mu sync.RWMutex // Concurrent access protection
}

// NewState creates a new state manager
func NewState() *State {
	now := time.Now()
	return &State{
		values:      make(map[string]any),
		initial:     make(map[string]any),
		touched:     make(map[string]bool),
		dirty:       make(map[string]bool),
		errors:      make(map[string][]string),
		initialized: now,
		lastUpdated: now,
	}
}

// ═══════════════════════════════════════════════════════════════════════════
// RuntimeStateManager Interface Implementation
// ═══════════════════════════════════════════════════════════════════════════

// Initialize sets initial state from schema and provided data
// Implements schema.RuntimeStateManager interface
func (s *State) Initialize(sch *schema.Schema, data map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sch == nil {
		return fmt.Errorf("schema cannot be nil")
	}

	// Clear existing state
	s.values = make(map[string]any)
	s.initial = make(map[string]any)
	s.touched = make(map[string]bool)
	s.dirty = make(map[string]bool)
	s.errors = make(map[string][]string)

	// Initialize with schema field defaults first
	for i := range sch.Fields {
		field := &sch.Fields[i]

		// Use Default if available, otherwise use Value
		var defaultValue any
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

	s.initialized = time.Now()
	s.lastUpdated = time.Now()

	return nil
}

// GetValue retrieves a field value
// Implements schema.RuntimeStateManager interface
func (s *State) GetValue(fieldName string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if fieldName == "" {
		return nil, false
	}

	value, exists := s.values[fieldName]
	return value, exists
}

// SetValue updates a field value and tracks dirty state
// Implements schema.RuntimeStateManager interface
func (s *State) SetValue(fieldName string, value any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if fieldName == "" {
		return fmt.Errorf("field name cannot be empty")
	}

	// Update value
	s.values[fieldName] = value

	// Mark as dirty if changed from initial
	initialValue, hasInitial := s.initial[fieldName]
	if !hasInitial || !s.valuesEqual(initialValue, value) {
		s.dirty[fieldName] = true
	} else {
		s.dirty[fieldName] = false
	}

	s.lastUpdated = time.Now()

	return nil
}

// GetAll returns all current values
// Implements schema.RuntimeStateManager interface
func (s *State) GetAll() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a deep copy to prevent external modification
	result := make(map[string]any, len(s.values))
	for k, v := range s.values {
		result[k] = s.deepCopyValue(v)
	}

	return result
}

// IsTouched checks if field has been touched
// Implements schema.RuntimeStateManager interface
func (s *State) IsTouched(fieldName string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.touched[fieldName]
}

// Touch marks a field as touched (user interacted with it)
// Implements schema.RuntimeStateManager interface
func (s *State) Touch(fieldName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if fieldName == "" {
		return
	}

	s.touched[fieldName] = true
	s.lastUpdated = time.Now()
}

// IsDirty checks if field changed from initial value
// Implements schema.RuntimeStateManager interface
func (s *State) IsDirty(fieldName string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.dirty[fieldName]
}

// IsAnyDirty checks if any field has been modified
// Implements schema.RuntimeStateManager interface
func (s *State) IsAnyDirty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, isDirty := range s.dirty {
		if isDirty {
			return true
		}
	}
	return false
}

// GetErrors returns validation errors for a field
// Implements schema.RuntimeStateManager interface
func (s *State) GetErrors(fieldName string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	errors, exists := s.errors[fieldName]
	if !exists {
		return nil
	}

	// Return a copy to prevent external modification
	result := make([]string, len(errors))
	copy(result, errors)
	return result
}

// SetErrors sets validation errors for a field
// Implements schema.RuntimeStateManager interface
func (s *State) SetErrors(fieldName string, errors []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if fieldName == "" {
		return
	}

	if len(errors) == 0 {
		delete(s.errors, fieldName)
	} else {
		// Store a copy to prevent external modification
		errorsCopy := make([]string, len(errors))
		copy(errorsCopy, errors)
		s.errors[fieldName] = errorsCopy
	}

	s.lastUpdated = time.Now()
}

// GetAllErrors returns all current validation errors
// Implements schema.RuntimeStateManager interface
func (s *State) GetAllErrors() map[string][]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a deep copy to prevent external modification
	result := make(map[string][]string, len(s.errors))
	for field, errors := range s.errors {
		errorsCopy := make([]string, len(errors))
		copy(errorsCopy, errors)
		result[field] = errorsCopy
	}

	return result
}

// IsValid checks if entire form is valid (no errors)
// Implements schema.RuntimeStateManager interface
func (s *State) IsValid() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.errors) == 0
}

// Reset clears all state back to initial
// Implements schema.RuntimeStateManager interface
func (s *State) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear mutable state
	s.values = make(map[string]any)
	s.touched = make(map[string]bool)
	s.dirty = make(map[string]bool)
	s.errors = make(map[string][]string)

	// Restore initial values
	for k, v := range s.initial {
		s.values[k] = s.deepCopyValue(v)
	}

	s.lastUpdated = time.Now()
}

// CreateSnapshot creates a snapshot of current state
// Implements schema.RuntimeStateManager interface
// Returns schema.StateSnapshot type to avoid import cycles
func (s *State) CreateSnapshot() any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshot := &schema.StateSnapshot{
		Values:      make(map[string]any, len(s.values)),
		Initial:     make(map[string]any, len(s.initial)),
		Touched:     make(map[string]bool, len(s.touched)),
		Dirty:       make(map[string]bool, len(s.dirty)),
		Errors:      make(map[string][]string, len(s.errors)),
		Timestamp:   time.Now(),
		Initialized: s.initialized,
	}

	// Deep copy all state
	for k, v := range s.values {
		snapshot.Values[k] = s.deepCopyValue(v)
	}
	for k, v := range s.initial {
		snapshot.Initial[k] = s.deepCopyValue(v)
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

	return snapshot
}

// RestoreSnapshot restores state from a snapshot
// Implements schema.RuntimeStateManager interface
// Accepts schema.StateSnapshot type to avoid import cycles
func (s *State) RestoreSnapshot(snapshot any) {
	stateSnapshot, ok := snapshot.(*schema.StateSnapshot)
	if !ok {
		return // Invalid snapshot type
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Create new maps
	s.values = make(map[string]any, len(stateSnapshot.Values))
	s.initial = make(map[string]any, len(stateSnapshot.Initial))
	s.touched = make(map[string]bool, len(stateSnapshot.Touched))
	s.dirty = make(map[string]bool, len(stateSnapshot.Dirty))
	s.errors = make(map[string][]string, len(stateSnapshot.Errors))

	// Restore all state with deep copies
	for k, v := range stateSnapshot.Values {
		s.values[k] = s.deepCopyValue(v)
	}
	for k, v := range stateSnapshot.Initial {
		s.initial[k] = s.deepCopyValue(v)
	}
	for k, v := range stateSnapshot.Touched {
		s.touched[k] = v
	}
	for k, v := range stateSnapshot.Dirty {
		s.dirty[k] = v
	}
	for k, v := range stateSnapshot.Errors {
		errorsCopy := make([]string, len(v))
		copy(errorsCopy, v)
		s.errors[k] = errorsCopy
	}

	s.initialized = stateSnapshot.Initialized
	s.lastUpdated = time.Now()
}

// ═══════════════════════════════════════════════════════════════════════════
// Additional Methods Beyond RuntimeStateManager Interface
// ═══════════════════════════════════════════════════════════════════════════

// GetInitialValues returns all initial values
func (s *State) GetInitialValues() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]any, len(s.initial))
	for k, v := range s.initial {
		result[k] = s.deepCopyValue(v)
	}

	return result
}

// GetChangedValues returns only the values that have changed from initial
func (s *State) GetChangedValues() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]any)
	for field, isDirty := range s.dirty {
		if isDirty {
			if value, exists := s.values[field]; exists {
				result[field] = s.deepCopyValue(value)
			}
		}
	}

	return result
}

// UpdateValues updates multiple field values at once (batch operation)
func (s *State) UpdateValues(updates map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for field, value := range updates {
		if field == "" {
			return fmt.Errorf("field name cannot be empty")
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

	s.lastUpdated = time.Now()
	return nil
}

// ClearErrors removes all errors for a field
func (s *State) ClearErrors(fieldName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.errors, fieldName)
	s.lastUpdated = time.Now()
}

// ClearAllErrors removes all validation errors
func (s *State) ClearAllErrors() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.errors = make(map[string][]string)
	s.lastUpdated = time.Now()
}

// ResetField resets a single field to its initial value
func (s *State) ResetField(fieldName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if fieldName == "" {
		return
	}

	// Restore initial value
	if initialValue, exists := s.initial[fieldName]; exists {
		s.values[fieldName] = s.deepCopyValue(initialValue)
	} else {
		delete(s.values, fieldName)
	}

	// Clear field state
	s.dirty[fieldName] = false
	s.touched[fieldName] = false
	delete(s.errors, fieldName)

	s.lastUpdated = time.Now()
}

// SetInitialValue updates the initial value for a field (useful for dynamic forms)
func (s *State) SetInitialValue(fieldName string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if fieldName == "" {
		return
	}

	s.initial[fieldName] = value

	// Recalculate dirty state based on new initial value
	if currentValue, exists := s.values[fieldName]; exists {
		s.dirty[fieldName] = !s.valuesEqual(value, currentValue)
	}

	s.lastUpdated = time.Now()
}

// HasField checks if a field exists in the state
func (s *State) HasField(fieldName string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.values[fieldName]
	return exists
}

// RemoveField removes a field completely from state
func (s *State) RemoveField(fieldName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if fieldName == "" {
		return
	}

	delete(s.values, fieldName)
	delete(s.initial, fieldName)
	delete(s.touched, fieldName)
	delete(s.dirty, fieldName)
	delete(s.errors, fieldName)

	s.lastUpdated = time.Now()
}

// ═══════════════════════════════════════════════════════════════════════════
// Statistics and Query Methods
// ═══════════════════════════════════════════════════════════════════════════

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

// GetStats returns comprehensive state statistics
// Returns schema.StateStats to avoid import cycles
func (s *State) GetStats() *schema.StateStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return &schema.StateStats{
		FieldCount:        len(s.values),
		TouchedCount:      s.getTouchedCountUnsafe(),
		DirtyCount:        s.getDirtyCountUnsafe(),
		ErrorCount:        s.getErrorCountUnsafe(),
		IsValid:           len(s.errors) == 0,
		InitializedAt:     s.initialized,
		LastUpdated:       s.lastUpdated,
		HasUnsavedChanges: s.hasUnsavedChangesUnsafe(),
	}
}

// ═══════════════════════════════════════════════════════════════════════════
// Private Helper Methods
// ═══════════════════════════════════════════════════════════════════════════

// valuesEqual compares two values for equality, handling different types
func (s *State) valuesEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	return reflect.DeepEqual(a, b)
}

// deepCopyValue creates a deep copy of a value
func (s *State) deepCopyValue(value any) any {
	if value == nil {
		return nil
	}

	// For maps and slices, we need proper deep copy
	switch v := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(v))
		for k, val := range v {
			result[k] = s.deepCopyValue(val)
		}
		return result
	case []any:
		result := make([]any, len(v))
		for i, val := range v {
			result[i] = s.deepCopyValue(val)
		}
		return result
	default:
		// For primitive types, direct assignment is fine
		return value
	}
}

// Unsafe methods (must be called with lock held)

func (s *State) getTouchedCountUnsafe() int {
	count := 0
	for _, touched := range s.touched {
		if touched {
			count++
		}
	}
	return count
}

func (s *State) getDirtyCountUnsafe() int {
	count := 0
	for _, dirty := range s.dirty {
		if dirty {
			count++
		}
	}
	return count
}

func (s *State) getErrorCountUnsafe() int {
	count := 0
	for _, errors := range s.errors {
		count += len(errors)
	}
	return count
}

func (s *State) hasUnsavedChangesUnsafe() bool {
	for _, dirty := range s.dirty {
		if dirty {
			return true
		}
	}
	return false
}
