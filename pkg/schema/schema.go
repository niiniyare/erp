// Package schema provides a JSON-driven UI schema system for building enterprise forms and components.
// Backend developers can define complete UIs by writing JSON files - no frontend code required.
//
// Example usage:
//
//	schema := schema.NewSchema("user-form", schema.TypeForm, "Create User")
//	schema.AddField(schema.Field{Name: "email", Type: schema.FieldEmail, Label: "Email"})
//	schema.AddAction(schema.Action{ID: "submit", Type: schema.ActionSubmit, Text: "Save"})
package schema

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/pkg/condition"
)

// Schema is the main entry point - defines a complete form, page, or UI component.
// All UI elements are configured through this structure via JSON.
// This implementation is thread-safe and can be safely used concurrently.
type Schema struct {
	mu sync.RWMutex // Protects all mutable fields

	// Identity
	ID          string `json:"id" validate:"required,min=1,max=100" example:"user-create-form"`
	Type        Type   `json:"type" validate:"required" example:"form"`
	Version     string `json:"version,omitempty" validate:"semver" example:"1.2.0"`
	Title       string `json:"title" validate:"required,min=1,max=200" example:"Create User"`
	Description string `json:"description,omitempty" validate:"max=1000" example:"Add a new user to the system"`

	// Core UI structure
	Config   *Config  `json:"config,omitempty"`                  // HTTP/API configuration
	Layout   *Layout  `json:"layout,omitempty"`                  // Visual layout (grid, tabs, sections)
	Fields   []Field  `json:"fields,omitempty" validate:"dive"`  // Form fields and inputs
	Actions  []Action `json:"actions,omitempty" validate:"dive"` // Buttons and actions
	Children []Schema `json:"children,omitempty"`                // Nested components

	// Performance optimization: field name -> index mapping
	fieldMap map[string]int `json:"-"` // Not serialized

	// Enterprise features
	Security   *Security   `json:"security,omitempty"`   // CSRF, rate limiting, encryption
	Tenant     *Tenant     `json:"tenant,omitempty"`     // Multi-tenancy isolation
	Workflow   *Workflow   `json:"workflow,omitempty"`   // Approval workflows
	Validation *Validation `json:"validation,omitempty"` // Cross-field validation rules
	Events     *Events     `json:"events,omitempty"`     // Lifecycle event handlers
	I18n       *I18n       `json:"i18n,omitempty"`       // Internationalization
	Mixin      *Mixin

	// Frontend framework integration
	HTMX   *HTMX   `json:"htmx,omitempty"`   // HTMX configuration
	Alpine *Alpine `json:"alpine,omitempty"` // Alpine.js configuration

	// Organization and discovery
	Meta     *Meta    `json:"meta,omitempty"`                          // Creation/update metadata
	Tags     []string `json:"tags,omitempty" validate:"dive,alphanum"` // Search tags
	Category string   `json:"category,omitempty" validate:"alphanum"`  // UI category
	Module   string   `json:"module,omitempty" validate:"alphanum"`    // ERP module name

	// Runtime state (managed by system, not in JSON)
	State   *State   `json:"state,omitempty"`   // Form values, errors, dirty state
	Context *Context `json:"context,omitempty"` // Request context, user info, permissions

	// Internal fields (not serialized)
	evaluator *condition.Evaluator `json:"-"` // Condition evaluator for dynamic behavior
}

// Type defines what kind of UI element this schema represents
type Type string

const (
	TypeForm      Type = "form"      // Complete form with fields and submission
	TypeComponent Type = "component" // Reusable UI component
	TypeLayout    Type = "layout"    // Layout container only
	TypeWorkflow  Type = "workflow"  // Multi-step workflow
	TypeTheme     Type = "theme"     // Theme/styling configuration
	TypePage      Type = "page"      // Full page layout
)

// Config holds HTTP and API configuration for form submission
type Config struct {
	Action   string            `json:"action,omitempty" validate:"url" example:"/api/v1/users"`
	Method   string            `json:"method,omitempty" validate:"oneof=GET POST PUT PATCH DELETE" example:"POST"`
	Target   string            `json:"target,omitempty" validate:"html_id" example:"#main-content"` // HTMX target
	Encoding string            `json:"encoding,omitempty" validate:"oneof=application/json multipart/form-data application/x-www-form-urlencoded"`
	Timeout  int               `json:"timeout,omitempty" validate:"min=0,max=300000" example:"30000"` // milliseconds
	Headers  map[string]string `json:"headers,omitempty"`                                             // Custom request headers
	Params   map[string]string `json:"params,omitempty"`                                              // Query parameters
	Cache    bool              `json:"cache,omitempty"`                                               // Enable response caching
	CacheTTL int               `json:"cacheTTL,omitempty" validate:"min=0" example:"300"`             // seconds
}

// State represents runtime form state - managed by the frontend
type State struct {
	Values      map[string]any    `json:"values,omitempty"`      // Current field values
	Errors      map[string]string `json:"errors,omitempty"`      // Validation errors by field
	Touched     map[string]bool   `json:"touched,omitempty"`     // Fields user has interacted with
	Dirty       map[string]bool   `json:"dirty,omitempty"`       // Fields that changed from default
	Valid       bool              `json:"valid,omitempty"`       // Overall form validity
	Submitting  bool              `json:"submitting,omitempty"`  // Submission in progress
	SubmitCount int               `json:"submitCount,omitempty"` // Number of submission attempts
	LastUpdated time.Time         `json:"lastUpdated"`           // Last state change timestamp
	CurrentStep string            `json:"currentStep,omitempty"` // For multi-step forms
	CurrentTab  string            `json:"currentTab,omitempty"`  // For tabbed forms
}

// Context provides runtime execution context - injected by the system
type Context struct {
	// User and session
	UserID    string `json:"userId,omitempty" validate:"uuid"`
	TenantID  string `json:"tenantId,omitempty" validate:"uuid"`
	SessionID string `json:"sessionId,omitempty" validate:"uuid"`
	RequestID string `json:"requestId,omitempty" validate:"uuid"`

	// Request metadata
	IP        string `json:"ip,omitempty" validate:"ip"`
	UserAgent string `json:"userAgent,omitempty"`
	Locale    string `json:"locale,omitempty" validate:"locale"`
	Timezone  string `json:"timezone,omitempty" validate:"timezone"`

	// Authorization
	Permissions []string `json:"permissions,omitempty"` // User permissions
	Roles       []string `json:"roles,omitempty"`       // User roles

	// Environment
	Environment string         `json:"environment,omitempty" validate:"oneof=development staging production"`
	Debug       bool           `json:"debug,omitempty"` // Enable debug mode
	Data        map[string]any `json:"data,omitempty"`  // Custom context data
}

// Core interfaces - implement these to extend the schema system

// Validator validates schema structure and submitted data
type Validator interface {
	ValidateSchema(ctx context.Context, schema *Schema) error
	ValidateData(ctx context.Context, schema *Schema, data map[string]any) error
}

// Enricher provides configurable schema enrichment with functional options
type Enricher interface {
	// Enrich enriches schema with provided options
	Enrich(ctx context.Context, schema *Schema, opts ...EnrichOption) error
}

// EnrichOption configures enrichment behavior
type EnrichOption func(*EnrichConfig)

// EnrichConfig holds enrichment configuration
type EnrichConfig struct {
	User           uuid.UUID
	Environment    string
	Features       map[string]bool
	FieldFilter    func(*Field) bool
	DepthLimit     int
	SkipValidation bool
	Extensions     map[string]any
}

// Renderer converts schema to HTML/templates
type Renderer interface {
	Render(ctx context.Context, schema *Schema, data map[string]any) (string, error)
	RenderField(ctx context.Context, field *Field, value any) (string, error)
}

// Registry manages schema storage and retrieval
type Registry interface {
	Register(ctx context.Context, schema *Schema) error
	Get(ctx context.Context, id string) (*Schema, error)
	List(ctx context.Context, filter map[string]any) ([]*Schema, error)
	Update(ctx context.Context, schema *Schema) error
	Delete(ctx context.Context, id string) error
	GetVersion(ctx context.Context, id, version string) (*Schema, error) // Version support
}

// DataSourceResolver resolves dynamic options for select fields
type DataSourceResolver interface {
	Resolve(ctx context.Context, source *DataSource, params map[string]string) ([]Option, error)
	InvalidateCache(ctx context.Context, source *DataSource) error // Clear cached options
}

// TransformProcessor handles data transformation
type TransformProcessor interface {
	Transform(ctx context.Context, transform *Transform, value any) (any, error)
}

// PermissionEvaluator interface - evaluates ABAC permissions
// Backend services implement this to check permissions like "contact:create", "transaction:delete"
type PermissionEvaluator interface {
	// EvaluatePermission evaluates if user has permission based on resource:action format
	// e.g., "contact:create", "transaction:delete", "report:view"
	EvaluatePermission(ctx context.Context, permission string, renderCtx *Context) bool
}

// ConditionalUpdate  Conditional types
type ConditionalUpdate struct {
	FieldName string `json:"fieldName"`
	Visible   bool   `json:"visible"`
	Required  bool   `json:"required"`
	Editable  bool   `json:"editable"`
	Reason    string `json:"reason,omitempty"`
}

type ConditionalResults struct {
	Updates   []ConditionalUpdate `json:"updates"`
	Changed   bool                `json:"changed"`
	Timestamp time.Time           `json:"timestamp"`
}

// LayoutConditionalResult represents conditional evaluation for layout components
type LayoutConditionalResult struct {
	Changed         bool                `json:"changed"`         // Whether any conditions changed
	VisibleSections map[string]bool     `json:"visibleSections"` // Section visibility
	VisibleTabs     map[string]bool     `json:"visibleTabs"`     // Tab visibility
	VisibleSteps    map[string]bool     `json:"visibleSteps"`    // Step visibility
	VisibleGroups   map[string]bool     `json:"visibleGroups"`   // Group visibility
	FieldUpdates    []ConditionalUpdate `json:"fieldUpdates"`    // Field updates
}

// NewSchema constructor - creates a new schema with sensible defaults
func NewSchema(id string, schemaType Type, title string) *Schema {
	now := time.Now()
	return &Schema{
		ID:       id,
		Type:     schemaType,
		Version:  "1.0.0",
		Title:    title,
		fieldMap: make(map[string]int),
		Meta: &Meta{
			CreatedAt: now,
			UpdatedAt: now,
		},
		State: &State{
			Values:      make(map[string]any),
			Errors:      make(map[string]string),
			Touched:     make(map[string]bool),
			Dirty:       make(map[string]bool),
			Valid:       true,
			LastUpdated: now,
		},
		Fields:  []Field{},
		Actions: []Action{},
	}
}

// buildFieldMap builds the internal field lookup map for O(1) access
func (s *Schema) buildFieldMap() {
	s.fieldMap = make(map[string]int, len(s.Fields))
	for i, field := range s.Fields {
		s.fieldMap[field.Name] = i
	}
}

// SetEvaluator sets the condition evaluator for the schema and all its fields
func (s *Schema) SetEvaluator(evaluator *condition.Evaluator) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.evaluator = evaluator

	// Set evaluator on all fields
	for i := range s.Fields {
		s.Fields[i].SetEvaluator(evaluator)
	}
}

// GetEvaluator returns the schema's condition evaluator
func (s *Schema) GetEvaluator() *condition.Evaluator {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.evaluator
}

// AddField appends a field to the schema (thread-safe)
func (s *Schema) AddField(field Field) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Fields == nil {
		s.Fields = []Field{}
	}

	// Set evaluator on new field if schema has one
	if s.evaluator != nil {
		field.SetEvaluator(s.evaluator)
	}

	s.Fields = append(s.Fields, field)

	// Update field map
	s.fieldMap[field.Name] = len(s.Fields) - 1

	s.updateTimestampUnsafe()
}

// AddFields appends multiple fields to the schema (thread-safe)
func (s *Schema) AddFields(fields ...Field) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, field := range fields {
		if s.evaluator != nil {
			field.SetEvaluator(s.evaluator)
		}
		s.Fields = append(s.Fields, field)
		s.fieldMap[field.Name] = len(s.Fields) - 1
	}

	s.updateTimestampUnsafe()
}

// RemoveField removes a field by name (thread-safe)
func (s *Schema) RemoveField(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx, exists := s.fieldMap[name]
	if !exists {
		return false
	}

	// Remove from slice
	s.Fields = append(s.Fields[:idx], s.Fields[idx+1:]...)

	// Rebuild field map
	s.buildFieldMap()

	s.updateTimestampUnsafe()
	return true
}

// UpdateField updates an existing field by name (thread-safe)
func (s *Schema) UpdateField(name string, updatedField Field) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx, exists := s.fieldMap[name]
	if !exists {
		return false
	}

	// Preserve evaluator
	if s.evaluator != nil {
		updatedField.SetEvaluator(s.evaluator)
	}

	s.Fields[idx] = updatedField

	// Update field map if name changed
	if updatedField.Name != name {
		delete(s.fieldMap, name)
		s.fieldMap[updatedField.Name] = idx
	}

	s.updateTimestampUnsafe()
	return true
}

// AddAction appends an action button to the schema (thread-safe)
func (s *Schema) AddAction(action Action) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Actions == nil {
		s.Actions = []Action{}
	}
	s.Actions = append(s.Actions, action)
	s.updateTimestampUnsafe()
}

// AddActions appends multiple actions to the schema (thread-safe)
func (s *Schema) AddActions(actions ...Action) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, action := range actions {
		s.Actions = append(s.Actions, action)
	}
	s.updateTimestampUnsafe()
}

// RemoveAction removes an action by ID (thread-safe)
func (s *Schema) RemoveAction(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, action := range s.Actions {
		if action.ID == id {
			s.Actions = append(s.Actions[:i], s.Actions[i+1:]...)
			s.updateTimestampUnsafe()
			return true
		}
	}
	return false
}

// UpdateAction updates an existing action by ID (thread-safe)
func (s *Schema) UpdateAction(id string, updatedAction Action) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, action := range s.Actions {
		if action.ID == id {
			s.Actions[i] = updatedAction
			s.updateTimestampUnsafe()
			return true
		}
	}
	return false
}

// GetField retrieves a field by name (returns a copy for safety)
func (s *Schema) GetField(name string) (Field, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	idx, exists := s.fieldMap[name]
	if !exists {
		return Field{}, false
	}

	return s.Fields[idx], true
}

// GetFieldPtr retrieves a pointer to a field by name (use with caution - for internal use)
// This is only for performance-critical paths where you need direct access
func (s *Schema) GetFieldPtr(name string) (*Field, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	idx, exists := s.fieldMap[name]
	if !exists {
		return nil, false
	}

	return &s.Fields[idx], true
}

// GetFieldByIndex retrieves a field by index (returns a copy)
func (s *Schema) GetFieldByIndex(index int) (Field, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if index >= 0 && index < len(s.Fields) {
		return s.Fields[index], true
	}
	return Field{}, false
}

// GetAction retrieves an action by ID (returns a copy)
func (s *Schema) GetAction(id string) (Action, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, action := range s.Actions {
		if action.ID == id {
			return action, true
		}
	}
	return Action{}, false
}

// HasField checks if a field exists
func (s *Schema) HasField(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.fieldMap[name]
	return exists
}

// HasFieldOfType checks if a field exists with the specified type
func (s *Schema) HasFieldOfType(name string, fieldType FieldType) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	idx, exists := s.fieldMap[name]
	return exists && s.Fields[idx].Type == fieldType
}

// HasAction checks if an action exists
func (s *Schema) HasAction(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, action := range s.Actions {
		if action.ID == id {
			return true
		}
	}
	return false
}

// GetFieldNames returns all field names
func (s *Schema) GetFieldNames() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	names := make([]string, len(s.Fields))
	for i, field := range s.Fields {
		names[i] = field.Name
	}
	return names
}

// GetActionIDs returns all action IDs
func (s *Schema) GetActionIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]string, len(s.Actions))
	for i, action := range s.Actions {
		ids[i] = action.ID
	}
	return ids
}

// GetVisibleFields returns fields visible for given data context
func (s *Schema) GetVisibleFields(ctx context.Context, data map[string]any) []Field {
	s.mu.RLock()
	defer s.mu.RUnlock()

	visible := make([]Field, 0, len(s.Fields))
	for i := range s.Fields {
		if ok, _ := s.Fields[i].IsVisible(ctx, data); ok {
			visible = append(visible, s.Fields[i]) // Return copy
		}
	}
	return visible
}

// GetHiddenFields returns fields hidden for given data context
func (s *Schema) GetHiddenFields(ctx context.Context, data map[string]any) []Field {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hidden := make([]Field, 0, len(s.Fields))
	for i := range s.Fields {
		if ok, _ := s.Fields[i].IsVisible(ctx, data); !ok {
			hidden = append(hidden, s.Fields[i]) // Return copy
		}
	}
	return hidden
}

// GetRequiredFields returns required fields for given data context
func (s *Schema) GetRequiredFields(ctx context.Context, data map[string]any) []Field {
	s.mu.RLock()
	defer s.mu.RUnlock()

	required := make([]Field, 0, len(s.Fields))
	for i := range s.Fields {
		field := &s.Fields[i]

		// Check if field is required (ignore errors, treat as not required)
		isRequired, err := field.IsRequired(ctx, data)
		if err != nil {
			continue
		}

		if isRequired {
			if ok, _ := field.IsVisible(ctx, data); ok {
				required = append(required, *field) // Return copy
			}
		}
	}
	return required
}

// GetOptionalFields returns optional fields for given data context
func (s *Schema) GetOptionalFields(ctx context.Context, data map[string]any) []Field {
	s.mu.RLock()
	defer s.mu.RUnlock()

	optional := make([]Field, 0, len(s.Fields))
	for i := range s.Fields {
		field := &s.Fields[i]

		isRequired, err := field.IsRequired(ctx, data)
		if err != nil || isRequired {
			continue
		}

		if ok, _ := field.IsVisible(ctx, data); ok {
			optional = append(optional, *field) // Return copy
		}
	}
	return optional
}

// GetFieldsByType returns all fields of a specific type
func (s *Schema) GetFieldsByType(fieldType FieldType) []Field {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fields := make([]Field, 0)
	for i := range s.Fields {
		if s.Fields[i].Type == fieldType {
			fields = append(fields, s.Fields[i]) // Return copy
		}
	}
	return fields
}

// GetEnabledActions returns all enabled (non-disabled) actions
func (s *Schema) GetEnabledActions(ctx context.Context, data map[string]any) []Action {
	s.mu.RLock()
	defer s.mu.RUnlock()

	enabled := make([]Action, 0, len(s.Actions))
	for _, action := range s.Actions {
		if !action.Disabled && !action.Hidden {
			// Check conditions if present
			if action.Condition != nil && s.evaluator != nil {
				evalCtx := condition.NewEvalContext(data, condition.DefaultEvalOptions())
				result, err := s.evaluator.Evaluate(ctx, action.Condition, evalCtx)
				if err != nil || !result {
					continue
				}
			}
			enabled = append(enabled, action)
		}
	}
	return enabled
}

// ValidateData validates submitted data against the schema
func (s *Schema) ValidateData(ctx context.Context, data map[string]any) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	collector := NewErrorCollector()

	// Get visible required fields
	requiredFields := s.getRequiredFieldsUnsafe(ctx, data)

	// Check required fields
	for _, field := range requiredFields {
		value, exists := data[field.Name]
		if !exists || isEmpty(value) {
			collector.AddValidationError(
				field.Name,
				"required",
				fmt.Sprintf("%s is required", field.Label),
			)
		}
	}

	// Validate all provided field values
	for i := range s.Fields {
		field := &s.Fields[i]
		value, exists := data[field.Name]
		if exists {
			if err := field.ValidateValue(ctx, value); err != nil {
				var schemaErr SchemaError
				if errors.As(err, &schemaErr) {
					collector.AddFieldError(field.Name, schemaErr)
				} else {
					collector.AddValidationError(field.Name, "validation_failed", err.Error())
				}
			}
		}
	}

	if collector.HasErrors() {
		return collector.Errors()
	}

	return nil
}

// getRequiredFieldsUnsafe is an internal helper that doesn't lock (caller must hold lock)
func (s *Schema) getRequiredFieldsUnsafe(ctx context.Context, data map[string]any) []Field {
	required := make([]Field, 0, len(s.Fields))
	for i := range s.Fields {
		field := &s.Fields[i]

		isRequired, err := field.IsRequired(ctx, data)
		if err != nil {
			continue
		}

		if isRequired {
			if ok, _ := field.IsVisible(ctx, data); ok {
				required = append(required, *field)
			}
		}
	}
	return required
}

// DetectCircularDependencies checks for circular field dependencies
func (s *Schema) DetectCircularDependencies() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	visited := make(map[string]bool)
	stack := make(map[string]bool)

	for _, field := range s.Fields {
		if err := s.checkCycleUnsafe(field.Name, visited, stack); err != nil {
			return err
		}
	}
	return nil
}

// checkCycleUnsafe performs DFS to detect circular dependencies (caller must hold lock)
func (s *Schema) checkCycleUnsafe(fieldName string, visited, stack map[string]bool) error {
	if stack[fieldName] {
		return NewValidationError(
			"circular_dependency",
			"circular dependency detected involving field: "+fieldName,
		).WithField(fieldName)
	}
	if visited[fieldName] {
		return nil
	}

	visited[fieldName] = true
	stack[fieldName] = true

	idx, exists := s.fieldMap[fieldName]
	if exists {
		for _, dep := range s.Fields[idx].Dependencies {
			if err := s.checkCycleUnsafe(dep, visited, stack); err != nil {
				return err
			}
		}
	}

	stack[fieldName] = false
	return nil
}

// Validate performs comprehensive schema validation with context support
func (s *Schema) Validate(ctx context.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check context cancellation (only if context is provided)
	if ctx != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	collector := NewErrorCollector()

	// Basic field validation
	if s.ID == "" {
		collector.AddError(ErrInvalidSchemaID)
	}
	if s.Type == "" {
		collector.AddError(ErrInvalidSchemaType)
	}
	if s.Title == "" {
		collector.AddError(ErrInvalidSchemaTitle)
	}

	// Validate all fields
	for i, field := range s.Fields {
		// Check cancellation periodically (only if context is provided)
		if ctx != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
		}

		if err := field.Validate(ctx); err != nil {
			var schemaErr SchemaError
			if errors.As(err, &schemaErr) {
				collector.AddFieldError(fmt.Sprintf("Fields[%d]", i), schemaErr)
			} else {
				collector.AddValidationError(
					fmt.Sprintf("Fields[%d]", i),
					"field_validation",
					err.Error(),
				)
			}
		}
	}

	// Validate all actions
	for i, action := range s.Actions {
		if err := action.Validate(ctx); err != nil {
			var schemaErr SchemaError
			if errors.As(err, &schemaErr) {
				collector.AddFieldError(fmt.Sprintf("Actions[%d]", i), schemaErr)
			} else {
				collector.AddValidationError(
					fmt.Sprintf("Actions[%d]", i),
					"action_validation",
					err.Error(),
				)
			}
		}
	}

	// Check for duplicate field names
	seen := make(map[string]bool)
	for _, field := range s.Fields {
		if seen[field.Name] {
			collector.AddFieldError(
				field.Name,
				ErrFieldDuplicate.WithDetail("name", field.Name),
			)
		}
		seen[field.Name] = true
	}

	// Check for duplicate action IDs
	seenActions := make(map[string]bool)
	for _, action := range s.Actions {
		if seenActions[action.ID] {
			collector.AddValidationError(
				action.ID,
				"duplicate_action",
				fmt.Sprintf("duplicate action ID: %s", action.ID),
			)
		}
		seenActions[action.ID] = true
	}

	// Validate field dependencies exist
	for _, field := range s.Fields {
		for _, dep := range field.Dependencies {
			if _, exists := s.fieldMap[dep]; !exists {
				collector.AddValidationError(
					field.Name,
					"invalid_dependency",
					fmt.Sprintf("depends on non-existent field: %s", dep),
				)
			}
		}
	}

	// Check for circular dependencies
	visited := make(map[string]bool)
	stack := make(map[string]bool)
	for _, field := range s.Fields {
		if err := s.checkCycleUnsafe(field.Name, visited, stack); err != nil {
			var schemaErr SchemaError
			if errors.As(err, &schemaErr) {
				collector.AddError(schemaErr)
			} else {
				collector.AddValidationError("", "circular_dependency", err.Error())
			}
			break
		}
	}

	if collector.HasErrors() {
		return collector.Errors()
	}

	return nil
}

// Clone creates a deep copy of the schema with proper error handling
func (s *Schema) Clone() (*Schema, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a type alias to avoid calling custom MarshalJSON
	type Alias Schema
	alias := (*Alias)(s)

	// Marshal to JSON for deep copy using the alias (bypassing custom MarshalJSON)
	data, err := json.Marshal(alias)
	if err != nil {
		return nil, fmt.Errorf("clone marshal: %w", err)
	}

	var clone Schema
	if err := json.Unmarshal(data, &clone); err != nil {
		return nil, fmt.Errorf("clone unmarshal: %w", err)
	}

	// Preserve evaluator (doesn't serialize)
	clone.evaluator = s.evaluator

	// Rebuild field map
	clone.buildFieldMap()

	// Set evaluator on all fields
	if clone.evaluator != nil {
		for i := range clone.Fields {
			clone.Fields[i].SetEvaluator(clone.evaluator)
		}
	}

	return &clone, nil
}

// Reset resets the schema state (thread-safe)
func (s *Schema) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.State == nil {
		s.State = &State{}
	}

	s.State.Values = make(map[string]any)
	s.State.Errors = make(map[string]string)
	s.State.Touched = make(map[string]bool)
	s.State.Dirty = make(map[string]bool)
	s.State.Valid = true
	s.State.Submitting = false
	s.State.SubmitCount = 0
	s.State.LastUpdated = time.Now()
}

// SetFieldValue sets a value in the schema state (thread-safe)
func (s *Schema) SetFieldValue(fieldName string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.State == nil {
		s.State = &State{
			Values: make(map[string]any),
		}
	}
	if s.State.Values == nil {
		s.State.Values = make(map[string]any)
	}
	s.State.Values[fieldName] = value
	s.State.LastUpdated = time.Now()
}

// SetFieldValues sets multiple field values at once (thread-safe)
func (s *Schema) SetFieldValues(values map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.State == nil {
		s.State = &State{
			Values: make(map[string]any),
		}
	}
	if s.State.Values == nil {
		s.State.Values = make(map[string]any)
	}

	for k, v := range values {
		s.State.Values[k] = v
	}
	s.State.LastUpdated = time.Now()
}

// GetFieldValue gets a value from the schema state (thread-safe)
func (s *Schema) GetFieldValue(fieldName string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.State == nil || s.State.Values == nil {
		return nil, false
	}
	value, exists := s.State.Values[fieldName]
	return value, exists
}

// SetFieldError sets an error for a field (thread-safe)
func (s *Schema) SetFieldError(fieldName string, errorMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.State == nil {
		s.State = &State{
			Errors: make(map[string]string),
		}
	}
	if s.State.Errors == nil {
		s.State.Errors = make(map[string]string)
	}
	s.State.Errors[fieldName] = errorMsg
	s.State.Valid = false
}

// ClearFieldError clears an error for a field (thread-safe)
func (s *Schema) ClearFieldError(fieldName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.State != nil && s.State.Errors != nil {
		delete(s.State.Errors, fieldName)
		// Check if there are any remaining errors
		s.State.Valid = len(s.State.Errors) == 0
	}
}

// ClearAllErrors clears all field errors (thread-safe)
func (s *Schema) ClearAllErrors() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.State != nil {
		s.State.Errors = make(map[string]string)
		s.State.Valid = true
	}
}

// HasErrors checks if the schema has any validation errors (thread-safe)
func (s *Schema) HasErrors() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.State != nil && s.State.Errors != nil && len(s.State.Errors) > 0
}

// GetErrors returns all validation errors (thread-safe, returns a copy)
func (s *Schema) GetErrors() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.State == nil || s.State.Errors == nil {
		return make(map[string]string)
	}

	// Return a copy to prevent external modification
	errors := make(map[string]string, len(s.State.Errors))
	for k, v := range s.State.Errors {
		errors[k] = v
	}
	return errors
}

// MarkFieldTouched marks a field as touched (thread-safe)
func (s *Schema) MarkFieldTouched(fieldName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.State == nil {
		s.State = &State{
			Touched: make(map[string]bool),
		}
	}
	if s.State.Touched == nil {
		s.State.Touched = make(map[string]bool)
	}
	s.State.Touched[fieldName] = true
}

// IsFieldTouched checks if a field has been touched (thread-safe)
func (s *Schema) IsFieldTouched(fieldName string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.State != nil && s.State.Touched != nil && s.State.Touched[fieldName]
}

// MarkFieldDirty marks a field as dirty (changed from default) (thread-safe)
func (s *Schema) MarkFieldDirty(fieldName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.State == nil {
		s.State = &State{
			Dirty: make(map[string]bool),
		}
	}
	if s.State.Dirty == nil {
		s.State.Dirty = make(map[string]bool)
	}
	s.State.Dirty[fieldName] = true
}

// IsFieldDirty checks if a field is dirty (thread-safe)
func (s *Schema) IsFieldDirty(fieldName string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.State != nil && s.State.Dirty != nil && s.State.Dirty[fieldName]
}

// IsDirty checks if any field is dirty (thread-safe)
func (s *Schema) IsDirty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.State == nil || s.State.Dirty == nil {
		return false
	}
	for _, dirty := range s.State.Dirty {
		if dirty {
			return true
		}
	}
	return false
}

// GetFieldCount returns the number of fields in the schema (thread-safe)
func (s *Schema) GetFieldCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Fields)
}

// GetActionCount returns the number of actions in the schema (thread-safe)
func (s *Schema) GetActionCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Actions)
}

// IsEmpty checks if the schema has no fields or actions (thread-safe)
func (s *Schema) IsEmpty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Fields) == 0 && len(s.Actions) == 0
}

// updateTimestamp updates the schema's last modified timestamp (NOT thread-safe - caller must hold lock)
func (s *Schema) updateTimestampUnsafe() {
	if s.Meta == nil {
		s.Meta = &Meta{}
	}
	s.Meta.UpdatedAt = time.Now()
}

// updateTimestamp updates the schema's last modified timestamp (thread-safe)
func (s *Schema) updateTimestamp() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updateTimestampUnsafe()
}

// MarshalJSON implements custom JSON marshaling (thread-safe)
func (s *Schema) MarshalJSON() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	type Alias Schema
	s.updateTimestampUnsafe()
	return json.Marshal((*Alias)(s))
}

// UnmarshalJSON implements custom JSON unmarshaling (no validation in unmarshal)
func (s *Schema) UnmarshalJSON(data []byte) error {
	type Alias Schema
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Build field map after unmarshaling
	s.buildFieldMap()

	// Don't validate here - let caller decide when to validate
	return nil
}

// ToJSON converts the schema to JSON string (thread-safe)
func (s *Schema) ToJSON() (string, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return "", WrapError(err, "json_marshal_failed", "failed to marshal schema to JSON")
	}
	return string(data), nil
}

// ToJSONPretty converts the schema to pretty-printed JSON string (thread-safe)
func (s *Schema) ToJSONPretty() (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	type Alias Schema
	data, err := json.MarshalIndent((*Alias)(s), "", "  ")
	if err != nil {
		return "", WrapError(err, "json_marshal_failed", "failed to marshal schema to JSON")
	}
	return string(data), nil
}

// FromJSON creates a schema from JSON string
func FromJSON(jsonStr string) (*Schema, error) {
	var schema Schema
	if err := json.Unmarshal([]byte(jsonStr), &schema); err != nil {
		return nil, WrapError(err, "json_unmarshal_failed", "failed to unmarshal schema from JSON")
	}
	return &schema, nil
}

// Interface implementation checks

func (s *Schema) GetID() string      { return s.ID }
func (s *Schema) GetType() string    { return string(s.Type) }
func (s *Schema) GetVersion() string { return s.Version }

func (s *Schema) GetMetadata() *Meta {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Meta
}

// Utility functions

// MergeSchemas merges multiple schemas into one (experimental)
func MergeSchemas(base *Schema, others ...*Schema) (*Schema, error) {
	merged, err := base.Clone()
	if err != nil {
		return nil, fmt.Errorf("failed to clone base schema: %w", err)
	}

	for _, other := range others {
		// Merge fields (skip duplicates)
		for _, field := range other.Fields {
			if !merged.HasField(field.Name) {
				merged.AddField(field)
			}
		}

		// Merge actions (skip duplicates)
		for _, action := range other.Actions {
			if !merged.HasAction(action.ID) {
				merged.AddAction(action)
			}
		}

		// Merge tags
		if len(other.Tags) > 0 {
			merged.mu.Lock()
			tagSet := make(map[string]bool)
			for _, tag := range merged.Tags {
				tagSet[tag] = true
			}
			for _, tag := range other.Tags {
				if !tagSet[tag] {
					merged.Tags = append(merged.Tags, tag)
				}
			}
			merged.mu.Unlock()
		}
	}

	merged.updateTimestamp()
	return merged, nil
}

// FilterFields returns a new schema with only fields matching the predicate
func (s *Schema) FilterFields(predicate func(field Field) bool) (*Schema, error) {
	filtered, err := s.Clone()
	if err != nil {
		return nil, fmt.Errorf("failed to clone schema: %w", err)
	}

	filtered.mu.Lock()
	filtered.Fields = []Field{}
	filtered.buildFieldMap()
	filtered.mu.Unlock()

	s.mu.RLock()
	fields := make([]Field, len(s.Fields))
	copy(fields, s.Fields)
	s.mu.RUnlock()

	for _, field := range fields {
		if predicate(field) {
			filtered.AddField(field)
		}
	}

	return filtered, nil
}

// MapFields applies a transformation to all fields
func (s *Schema) MapFields(transform func(field Field) Field) (*Schema, error) {
	mapped, err := s.Clone()
	if err != nil {
		return nil, fmt.Errorf("failed to clone schema: %w", err)
	}

	mapped.mu.Lock()
	for i, field := range mapped.Fields {
		mapped.Fields[i] = transform(field)
	}
	mapped.buildFieldMap()
	mapped.updateTimestampUnsafe()
	mapped.mu.Unlock()

	return mapped, nil
}

// Equals performs a deep equality check between two schemas
func (s *Schema) Equals(other *Schema) bool {
	if s == nil && other == nil {
		return true
	}
	if s == nil || other == nil {
		return false
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	other.mu.RLock()
	defer other.mu.RUnlock()

	// Compare basic fields
	if s.ID != other.ID || s.Type != other.Type || s.Version != other.Version || s.Title != other.Title {
		return false
	}

	// Compare field counts
	if len(s.Fields) != len(other.Fields) || len(s.Actions) != len(other.Actions) {
		return false
	}

	// Compare fields
	for i := range s.Fields {
		if s.Fields[i].Name != other.Fields[i].Name || s.Fields[i].Type != other.Fields[i].Type {
			return false
		}
	}

	// Compare actions
	for i := range s.Actions {
		if s.Actions[i].ID != other.Actions[i].ID {
			return false
		}
	}

	return true
}
